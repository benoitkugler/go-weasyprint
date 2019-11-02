package utils

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// http://whatwg.org/C#space-character
const htmlWhitespace = " \t\n\f\r"

var (
	htmlSpacesRe               = regexp.MustCompile(fmt.Sprintf("[%s]+", htmlWhitespace))
	htmlSpaceSeparatedTokensRe = regexp.MustCompile(fmt.Sprintf("[^%s]+", htmlWhitespace))
)

type PageElement struct {
	Side         string
	Blank, First bool
	Name         string
	Index        int
}

type ElementKey struct {
	Element    *HTMLNode
	PageType   PageElement
	PseudoType string
}

func (e ElementKey) IsPageType() bool {
	return e.Element == nil
}

func (p PageElement) ToKey(pseudoType string) ElementKey {
	return ElementKey{PseudoType: pseudoType, PageType: p}
}

type HTMLNode html.Node

func (h *HTMLNode) AsHtmlNode() *html.Node {
	return (*html.Node)(h)
}

func (h *HTMLNode) ToKey(pseudoType string) ElementKey {
	return ElementKey{PseudoType: pseudoType, Element: h}
}

// Get returns the attribute `name` or ""
func (h HTMLNode) Get(name string) string {
	for _, attr := range h.Attr {
		if attr.Key == name {
			return attr.Val
		}
	}
	return ""
}

func (h *HTMLNode) Iter(tags ...atom.Atom) HtmlIterator {
	return NewHtmlIterator((*html.Node)(h), tags...)
}

// ------------------------------------ html walk utilities ------------------------------------

// HtmlIterator simplify the (depth first) walk on an HTML tree.
type HtmlIterator struct {
	toVisit []*html.Node
	tagsMap map[atom.Atom]bool // if nil, means all
}

// NewHtmlIterator use `root` as start point.
// If `tags` is given, only node matching one of them are returned.
func NewHtmlIterator(root *html.Node, tags ...atom.Atom) HtmlIterator {
	tagsMap := make(map[atom.Atom]bool)
	for _, tag := range tags {
		tagsMap[tag] = true
	}
	return HtmlIterator{toVisit: []*html.Node{root}, tagsMap: tagsMap}
}

func (h HtmlIterator) HasNext() bool {
	return len(h.toVisit) > 0
}

func (h *HtmlIterator) Next() *HTMLNode {
	if len(h.toVisit) == 0 {
		return nil
	}
	next := h.toVisit[0]
	h.toVisit = h.toVisit[1:]
	if next.FirstChild != nil {
		h.toVisit = append(h.toVisit, next.FirstChild)
	}
	if next.NextSibling != nil {
		h.toVisit = append(h.toVisit, next.NextSibling)
	}
	if len(h.tagsMap) == 0 || h.tagsMap[next.DataAtom] {
		return (*HTMLNode)(next)
	}
	return h.Next()
}

// Iter recursively `element` (and its children and so on ...) and returns the elements matching one of the given tags
//func Iter(element html.Node, tags ...atom.Atom) []html.Node {
//	tagsMap := make(map[atom.Atom]bool)
//	for _, tag := range tags {
//		tagsMap[tag] = true
//	}
//	var aux func(html.Node) []html.Node
//	aux = func(el html.Node) (out []html.Node) {
//		if tagsMap[el.DataAtom] {
//			out = append(out, el)
//		}
//		child := el.FirstChild
//		for child != nil {
//			out = append(out, aux(*child)...)
//			child = child.NextSibling
//		}
//		return
//	}
//	return aux(element)
//}

// NodeChildren returns the direct children of `element`.
// Skip empty text nodes
func (element HTMLNode) NodeChildren(skipBlank bool) (children []*HTMLNode) {
	child := element.FirstChild
	for child != nil {
		if !(skipBlank && child.Type == html.TextNode && strings.TrimSpace(child.Data) == "") {
			children = append(children, (*HTMLNode)(child))
		}
		child = child.NextSibling
	}
	return
}

// IsText returns true if the node is a non empty text node.
func (element HTMLNode) IsText() (bool, string) {
	if text := element.Data; element.Type == html.TextNode && text != "" {
		return true, text
	}
	return false, ""
}

// GetChildText returns the text directly in the element, not descendants.
func (element HTMLNode) GetChildText() string {
	var content []string
	if element.Type == html.TextNode {
		content = []string{element.Data}
	}

	for _, child := range element.NodeChildren(false) {
		if child.Type == html.TextNode {
			content = append(content, child.Data)
		}
	}
	return strings.Join(content, "")
}

// Transform (only) ASCII letters to lower case: A-Z is mapped to a-z.
//     This is used for `ASCII case-insensitive
//     <http://whatwg.org/C#ascii-case-insensitive>`_ matching.
//     This is different from the strings.ToLower function
//     which also affect non-ASCII characters,
//     sometimes mapping them into the ASCII range:
//     		keyword = u"Bac\u212Aground"
//     		assert strings.ToLower(keyword) == u"background"
//     		assert asciiLower(keyword) != strings.ToLower(keyword)
//     		assert asciiLower(keyword) == u"bac\u212Aground"
//
func AsciiLower(s string) string {
	rs := []rune(s)
	out := make([]rune, len(rs))
	for index, c := range rs {
		if c < unicode.MaxASCII {
			c = unicode.ToLower(c)
		}
		out[index] = c
	}
	return string(out)
}

// Return whether the given element has a `rel` attribute with the
// given link type.
// `linkType` must be a lower-case string.
func (element HTMLNode) HasLinkType(linkType string) bool {
	attr := element.Get("rel")
	matchs := htmlSpaceSeparatedTokensRe.FindAllString(attr, -1)
	for _, token := range matchs {
		if AsciiLower(token) == linkType {
			return true
		}
	}
	return false
}

// Return the base URL for the document.
// See http://www.w3.org/TR/html5/urls.html#document-base-url
func FindBaseUrl(htmlDocument *html.Node, fallbackBaseUrl string) string {
	iter := NewHtmlIterator(htmlDocument, atom.Base)
	firstBaseElement := iter.Next()
	if firstBaseElement != nil {
		href := strings.TrimSpace(firstBaseElement.Get("href"))
		if href != "" {
			out, err := SafeUrljoin(fallbackBaseUrl, href, true)
			if err != nil {
				return ""
			}
			return out
		}
	}
	return fallbackBaseUrl
}

// Return whether the given element has a `rel` attribute with the
// given link type (must be a lower-case string).
func ElementHasLinkType(element *HTMLNode, linkType string) bool {
	for _, token := range htmlSpaceSeparatedTokensRe.FindAllString(element.Get("rel"), -1) {
		if AsciiLower(token) == linkType {
			return true
		}
	}
	return false
}

// Meta-information belonging to a whole `Document`.
type DocumentMetadata struct {
	// The title of the document, as a string.
	// Extracted from the `<title>` element in HTML
	// and written to the `/Title` info field in PDF.
	Title string

	// The description of the document, as a string.
	// Extracted from the `<meta name=description>` element in HTML
	// and written to the `/Subject` info field in PDF.
	Description string

	// The name of one of the software packages
	// used to generate the document, as a string.
	// Extracted from the `<meta name=generator>` element in HTML
	// and written to the `/Creator` info field in PDF.
	Generator string

	// Keywords associated with the document, as a list of strings.
	// (Defaults to the empty list.)
	// Extracted from `<meta name=keywords>` elements in HTML
	// and written to the `/Keywords` info field in PDF.
	Keywords []string

	// The authors of the document, as a list of strings.
	// (Defaults to the empty list.)
	// Extracted from the `<meta name=author>` elements in HTML
	// and written to the `/Author` info field in PDF.
	Authors []string

	// The creation date of the document, as a string.
	// Dates are in one of the six formats specified in
	// `W3C’s profile of ISO 8601 <http://www.w3.org/TR/NOTE-datetime>`.
	// Extracted from the `<meta name=dcterms.created>` element in HTML
	// and written to the `/CreationDate` info field in PDF.
	Created string

	// The modification date of the document, as a string.
	// Dates are in one of the six formats specified in
	// `W3C’s profile of ISO 8601 <http://www.w3.org/TR/NOTE-datetime>`.
	// Extracted from the `<meta name=dcterms.modified>` element in HTML
	// and written to the `/ModDate` info field in PDF.
	Modified string

	// File attachments, as a list of tuples of URL and a description.
	// (Defaults to the empty list.)
	// Extracted from the `<link rel=attachment>` elements in HTML
	// and written to the `/EmbeddedFiles` dictionary in PDF.
	Attachments []Attachment
}
type Attachment struct {
	Url, Title string
}

// Relevant specs:
//     http://www.whatwg.org/html#the-title-element
//     http://www.whatwg.org/html#standard-metadata-names
//     http://wiki.whatwg.org/wiki/MetaExtensions
//     http://microformats.org/wiki/existing-rel-values#HTML5LinkTypeExtensions
//
func GetHtmlMetadata(wrapperElement *HTMLNode, baseUrl string) DocumentMetadata {
	title := ""
	description := ""
	generator := ""
	keywordsSet := map[string]bool{}
	var authors []string
	created := ""
	modified := ""
	var attachments []Attachment
	iter := wrapperElement.Iter(atom.Title, atom.Meta, atom.Link)
	for iter.HasNext() {
		element := iter.Next()
		switch element.DataAtom {
		case atom.Title:
			if title == "" {
				title = element.GetChildText()
			}
		case atom.Meta:
			name := AsciiLower(element.Get("name"))
			content := element.Get("content")
			switch name {
			case "keywords":
				for _, _keyword := range strings.Split(content, ",") {
					keyword := stripWhitespace(_keyword)
					keywordsSet[keyword] = true
				}
			case "author":
				authors = append(authors, content)
			case "description":
				if description == "" {
					description = content
				}
			case "generator":
				if generator == "" {
					generator = content
				}
			case "dcterms.created":
				if created == "" {
					created = parseW3cDate(name, content)
				}
			case "dcterms.modified":
				if modified == "" {
					modified = parseW3cDate(name, content)
				}
			}
		case atom.Link:
			if ElementHasLinkType(element, "attachment") {
				url := element.GetUrlAttribute("href", baseUrl, false)
				title := element.Get("title")
				if url == "" {
					log.Println("Missing href in <link rel='attachment'>")
				} else {
					attachments = append(attachments, Attachment{Url: url, Title: title})
				}
			}
		}
	}
	keywords := make([]string, 0, len(keywordsSet))
	for kw := range keywordsSet {
		keywords = append(keywords, kw)
	}
	return DocumentMetadata{
		Title:       title,
		Description: description,
		Generator:   generator,
		Keywords:    keywords,
		Authors:     authors,
		Created:     created,
		Modified:    modified,
		Attachments: attachments,
	}
}

// Use the HTML definition of "space character",
//     not all Unicode Whitespace.
//     http://www.whatwg.org/html#strip-leading-and-trailing-whitespace
//     http://www.whatwg.org/html#space-character
//
func stripWhitespace(s string) string {
	return strings.Trim(s, htmlWhitespace)
}

// YYYY (eg 1997)
// YYYY-MM (eg 1997-07)
// YYYY-MM-DD (eg 1997-07-16)
// YYYY-MM-DDThh:mmTZD (eg 1997-07-16T19:20+01:00)
// YYYY-MM-DDThh:mm:ssTZD (eg 1997-07-16T19:20:30+01:00)
// YYYY-MM-DDThh:mm:ss.sTZD (eg 1997-07-16T19:20:30.45+01:00)
var W3CDateRe = regexp.MustCompile(
	`^` +
		"[ \t\n\f\r]*" +
		`(?P<year>\d\d\d\d)` +
		`(?:` +
		`-(?P<month>0\d|1[012])` +
		`(?:` +
		`-(?P<day>[012]\d|3[01])` +
		`(?:` +
		`T(?P<hour>[01]\d|2[0-3])` +
		`:(?P<minute>[0-5]\d)` +
		`(?:` +
		`:(?P<second>[0-5]\d)` +
		`(?:\.\d+)?` + // Second fraction, ignored
		`)?` +
		`(?:` +
		`Z |` + //# UTC
		`(?P<tzHour>[+-](?:[01]\d|2[0-3]))` +
		`:(?P<tzMinute>[0-5]\d)` +
		`)` +
		`)?` +
		`)?` +
		`)?` +
		"[ \t\n\f\r]*" +
		`$`)

// http://www.w3.org/TR/NOTE-datetime
func parseW3cDate(metaName, s string) string {
	if W3CDateRe.MatchString(s) {
		return s
	} else {
		log.Printf("Invalid date in <meta name='%s'> %s \n", metaName, s)
		return ""
	}
}
