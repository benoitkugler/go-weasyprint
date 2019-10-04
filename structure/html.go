package structure

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/utils"

	"golang.org/x/net/html/atom"
)

type handlerFunction = func(element *utils.HTMLNode, box Box, getImageFromUri gifu, baseUrl string) []Box

var (
	HtmlHandlers = map[string]handlerFunction{
		"img":      handleImg,
		"embded":   handleEmbed,
		"object":   handleObject,
		"colgroup": handleColgroup,
		"col":      handleCol,
		"th":       handleTd,
		"td":       handleTd,
		"a":        handleA,
	}

	// http://whatwg.org/C#space-character
	HtmlWhitespace             = " \t\n\f\r"
	HtmlSpaceSeparatedTokensRe = regexp.MustCompile(fmt.Sprintf("[^%s]+", HtmlWhitespace))
)

// Return whether the given element has a ``rel`` attribute with the
// given link type (must be a lower-case string).
func elementHasLinkType(element *utils.HTMLNode, linkType string) bool {
	for _, token := range HtmlSpaceSeparatedTokensRe.FindAllString(element.Get("rel"), -1) {
		if utils.AsciiLower(token) == linkType {
			return true
		}
	}
	return false
}

// HandleElement handle HTML elements that need special care.
func handleElement(element *utils.HTMLNode, box Box, getImageFromUri gifu, baseUrl string) []Box {
	handler, in := HtmlHandlers[box.Box().elementTag]
	if in {
		return handler(element, box, getImageFromUri, baseUrl)
	} else {
		return []Box{box}
	}
}

// Wrap an image in a replaced box.
//
// That box is either block-level || inline-level, depending on what the
// element should be.
func makeReplacedBox(element *utils.HTMLNode, box Box, image pr.Image) Box {
	var newBox Box
	switch box.Box().style.GetDisplay() {
	case "block", "list-item", "table":
		b := NewBlockReplacedBox(element.Data, box.Box().style, image)
		newBox = &b
	default:
		// TODO: support images with "display: table-cell"?
		b := NewInlineReplacedBox(element.Data, box.Box().style, image)
		newBox = &b
	}
	// TODO: check other attributes that need to be copied
	// TODO: find another solution
	newBox.Box().stringSet = box.Box().stringSet
	newBox.Box().bookmarkLabel = box.Box().bookmarkLabel
	return newBox
}

// Handle ``<img>`` elements, return either an image || the alt-text.
// See: http://www.w3.org/TR/html5/embedded-content-1.html#the-img-element
func handleImg(element *utils.HTMLNode, box Box, getImageFromUri gifu, baseUrl string) []Box {
	src := element.GetUrlAttribute("src", baseUrl, false)
	alt := element.Get("alt")
	if src != "" {
		image := getImageFromUri(src, "")
		if image != nil {
			return []Box{makeReplacedBox(element, box, image)}
		}
		// Invalid image, use the alt-text.
		if alt != "" {
			box.Box().children = []Box{TextBoxAnonymousFrom(box, alt)}
			return []Box{box}
		}
	} else {
		if alt != "" {
			box.Box().children = []Box{TextBoxAnonymousFrom(box, alt)}
			return []Box{box}
		}
	}
	// The element represents nothing
	return nil
}

// Handle ``<embed>`` elements, return either an image || nothing.
// See: https://www.w3.org/TR/html5/embedded-content-0.html#the-embed-element
func handleEmbed(element *utils.HTMLNode, box Box, getImageFromUri gifu, baseUrl string) []Box {
	src := element.GetUrlAttribute("src", baseUrl, false)
	type_ := strings.TrimSpace(element.Get("type"))
	if src != "" {
		image := getImageFromUri(src, type_)
		if image != nil {
			return []Box{makeReplacedBox(element, box, image)}
		}
	}
	// No fallback.
	return nil
}

// Handle ``<object>`` elements, return either an image || the fallback
// content.
// See: https://www.w3.org/TR/html5/embedded-content-0.html#the-object-element
func handleObject(element *utils.HTMLNode, box Box, getImageFromUri gifu, baseUrl string) []Box {
	data := element.GetUrlAttribute("data", baseUrl, false)
	type_ := strings.TrimSpace(element.Get("type"))
	if data != "" {
		image := getImageFromUri(data, type_)
		if image != nil {
			return []Box{makeReplacedBox(element, box, image)}
		}
	}
	// The element’s children are the fallback.
	return []Box{box}
}

// Read an integer attribute from the HTML element. if true, the return value should be set on the box
// minimum = 1
func integerAttribute(element utils.HTMLNode, name string, minimum int) (bool, int) {
	value := strings.TrimSpace(element.Get(name))
	if value != "" {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return false, 0
		}
		if intValue >= minimum {
			return true, intValue
		}
	}
	return false, 0
}

// Handle the ``span`` attribute.
func handleColgroup(element *utils.HTMLNode, box Box, _ gifu, _ string) []Box {
	if tbox, ok := box.(*TableColumnGroupBox); ok { // leaf
		f := &tbox.TableFields

		hasCol := false
		for _, child := range element.NodeChildren(true) {
			if child.DataAtom == atom.Col {
				hasCol = true
				f.span = 0 // sum of the children’s spans
			}
		}
		if !hasCol {
			valid, span := integerAttribute(*element, "span", 1)
			if valid {
				f.span = span
			}
			children := make([]Box, f.span)
			for i := range children {
				children[i] = TableColumnBoxAnonymousFrom(box, nil)
			}
			box.Box().children = children
		}
	}
	return []Box{box}
}

// Handle the ``span`` attribute.
func handleCol(element *utils.HTMLNode, box Box, _ gifu, _ string) []Box {
	if tbox, ok := box.(*TableColumnBox); ok { // leaf
		f := &tbox.TableFields

		valid, span := integerAttribute(*element, "span", 1)
		if valid {
			f.span = span
		}
		if f.span > 1 {
			// Generate multiple boxes
			// http://lists.w3.org/Archives/Public/www-style/2011Nov/0293.html
			out := make([]Box, f.span)
			for i := range out {
				out[i] = box.Copy()
			}
			return out
		}
	}
	return []Box{box}
}

// Handle the ``colspan``, ``rowspan`` attributes.
func handleTd(element *utils.HTMLNode, box Box, _ gifu, _ string) []Box {
	if tbox, ok := box.(*TableCellBox); ok { // leaf
		// HTML 4.01 gives special meaning to colspan=0
		// http://www.w3.org/TR/html401/struct/tables.html#adef-rowspan
		// but HTML 5 removed it
		// http://www.w3.org/TR/html5/tabular-data.html#attr-tdth-colspan
		// rowspan=0 is still there though.

		f := &tbox.TableFields
		valid, span := integerAttribute(*element, "colspan", 1)
		if valid {
			f.colspan = span
		}
		valid, span = integerAttribute(*element, "rowspan", 0)
		if valid {
			f.rowspan = span
		}
	}
	return []Box{box}
}

// Handle the ``rel`` attribute.
func handleA(element *utils.HTMLNode, box Box, _ gifu, _ string) []Box {
	box.Box().isAttachment = elementHasLinkType(element, "attachment")
	return []Box{box}
}

// Return the base URL for the document.
// See http://www.w3.org/TR/html5/urls.html#document-base-url
//
func findBaseUrl(htmlDocument utils.HTMLNode, fallbackBaseUrl string) string {
	iter := htmlDocument.Iter(atom.Base)
	firstBaseElement := iter.Next()
	if firstBaseElement != nil {
		href := strings.TrimSpace(firstBaseElement.Get("href"))
		if href != "" {
			out, err := utils.BasicUrlJoin(fallbackBaseUrl, href)
			if err != nil {
				log.Println("invalid href : %s", err)
				return fallbackBaseUrl
			}
			return out
		}
	}
	return fallbackBaseUrl
}

type HtmlMetadata struct {
	title       string
	description string
	generator   string
	keywords    []string
	authors     []string
	created     string
	modified    string
	attachments []Attachment
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
func getHtmlMetadata(wrapperElement *utils.HTMLNode, baseUrl string) HtmlMetadata {
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
			name := utils.AsciiLower(element.Get("name"))
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
			if elementHasLinkType(element, "attachment") {
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
	return HtmlMetadata{
		title:       title,
		description: description,
		generator:   generator,
		keywords:    keywords,
		authors:     authors,
		created:     created,
		modified:    modified,
		attachments: attachments,
	}
}

// Use the HTML definition of "space character",
//     not all Unicode Whitespace.
//     http://www.whatwg.org/html#strip-leading-and-trailing-whitespace
//     http://www.whatwg.org/html#space-character
//
func stripWhitespace(s string) string {
	return strings.Trim(s, HtmlWhitespace)
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
