package structure

import (
	"fmt"
	"log"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/benoitkugler/go-weasyprint/css"
	"github.com/benoitkugler/go-weasyprint/utils"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type gifu = func(url string, mimeType string) css.ImageType
type HandlerFunction = func(element html.Node, box AllBox, getImageFromUri gifu, baseUrl string) []AllBox

var (
	HtmlHandlers = map[string]HandlerFunction{
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

// Transform (only) ASCII letters to lower case: A-Z is mapped to a-z.
//     :param string: An Unicode string.
//     :returns: A new Unicode string.
//     This is used for `ASCII case-insensitive
//     <http://whatwg.org/C#ascii-case-insensitive>`_ matching.
//     This is different from the :meth:`~py:str.lower` method of Unicode strings
//     which also affect non-ASCII characters,
//     sometimes mapping them into the ASCII range:
//     >>> keyword = u"Bac\N{KELVIN SIGN}ground"
//     >>> assert keyword.lower() == u"background"
//     >>> assert asciiLower(keyword) != keyword.lower()
//     >>> assert asciiLower(keyword) == u"bac\N{KELVIN SIGN}ground"
//
func asciiLower(s string) string {
	// is this implementation correct ?
	return strings.ToLower(s)
}

// Return whether the given element has a ``rel`` attribute with the
// given link type.
//     :param linkType: Must be a lower-case string.
//
func elementHasLinkType(element html.Node, linkType string) bool {
	for _, token := range HtmlSpaceSeparatedTokensRe.FindAllString(utils.GetAttribute(element, "rel"), -1) {
		if asciiLower(token) == linkType {
			return true
		}
	}
	return false
}

// HandleElement handle HTML elements that need special care.
func HandleElement(element html.Node, box AllBox, getImageFromUri gifu, baseUrl string) []AllBox {
	handler, in := HtmlHandlers[box.BaseBox().elementTag]
	if in {
		return handler(element, box, getImageFromUri, baseUrl)
	} else {
		return []AllBox{box}
	}
}

// Wrap an image in a replaced box.
//
// That box is either block-level || inline-level, depending on what the
// element should be.
func makeReplacedBox(element html.Node, box AllBox, image css.ImageType) AllBox {
	switch box.BaseBox().style.Strings["display"] {
	case "block", "list-item", "table":
		return NewBlockReplacedBox(element.Data, box.BaseBox().style, image)
	default:
		// TODO: support images with "display: table-cell"?
		return NewInlineReplacedBox(element.Data, box.BaseBox().style, image)
	}
}

// Handle ``<img>`` elements, return either an image || the alt-text.
// See: http://www.w3.org/TR/html5/embedded-content-1.html#the-img-element
func handleImg(element html.Node, box AllBox, getImageFromUri gifu, baseUrl string) []AllBox {
	src := utils.GetUrlAttribute(element, "src", baseUrl, false)
	alt := utils.GetAttribute(element, "alt")
	if src != "" {
		image := getImageFromUri(src, "")
		if image != nil {
			return []AllBox{makeReplacedBox(element, box, image)}
		}
	}
	// No src or invalid image, use the alt-text.
	if alt != "" {
		box.BaseBox().children = []AllBox{TextBoxAnonymousFrom(box, alt)}
		return []AllBox{box}
	} else {
		// The element represents nothing
		return nil

		// TODO: find some indicator that an image is missing.
		// For now, just remove the image.
	}
}

// Handle ``<embed>`` elements, return either an image || nothing.
// See: https://www.w3.org/TR/html5/embedded-content-0.html#the-embed-element
func handleEmbed(element html.Node, box AllBox, getImageFromUri gifu, baseUrl string) []AllBox {
	src := utils.GetUrlAttribute(element, "src", baseUrl, false)
	type_ := strings.TrimSpace(utils.GetAttribute(element, "type"))
	if src != "" {
		image := getImageFromUri(src, type_)
		if image != nil {
			return []AllBox{makeReplacedBox(element, box, image)}
		}
	}
	// No fallback.
	return nil
}

// Handle ``<object>`` elements, return either an image || the fallback
// content.
// See: https://www.w3.org/TR/html5/embedded-content-0.html#the-object-element
func handleObject(element html.Node, box AllBox, getImageFromUri gifu, baseUrl string) []AllBox {
	data := utils.GetUrlAttribute(element, "data", baseUrl, false)
	type_ := strings.TrimSpace(utils.GetAttribute(element, "type"))
	if data != "" {
		image := getImageFromUri(data, type_)
		if image != nil {
			return []AllBox{makeReplacedBox(element, box, image)}
		}
	}
	// The element’s children are the fallback.
	return []AllBox{box}
}

// Read an integer attribute from the HTML element. if true, the return value should be set on the box
// minimum = 1
func integerAttribute(element html.Node, name string, minimum int) (bool, int) {
	value := strings.TrimSpace(utils.GetAttribute(element, name))
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
func handleColgroup(element html.Node, box AllBox, _ gifu, _ string) []AllBox {
	if TypeTableColumnGroupBox.IsInstance(box) {
		f := box.TableFields()

		hasCol := false
		for _, child := range utils.NodeChildren(element) {
			if child.DataAtom == atom.Col {
				hasCol = true
				f.span = 0 // sum of the children’s spans
			}
		}
		if !hasCol {
			valid, span := integerAttribute(element, "span", 1)
			if valid {
				f.span = span
			}
			children := make([]AllBox, f.span)
			for i := range children {
				children[i] = TypeTableColumnBox.AnonymousFrom(box, nil)
			}
			box.BaseBox().children = children
		}
	}
	return []AllBox{box}
}

// Handle the ``span`` attribute.
func handleCol(element html.Node, box AllBox, _ gifu, _ string) []AllBox {
	if TypeTableColumnBox.IsInstance(box) {
		f := box.TableFields()

		valid, span := integerAttribute(element, "span", 1)
		if valid {
			f.span = span
		}
		if f.span > 1 {
			// Generate multiple boxes
			// http://lists.w3.org/Archives/Public/www-style/2011Nov/0293.html
			out := make([]AllBox, f.span)
			for i := range out {
				out[i] = box.Copy()
			}
			return out
		}
	}
	return []AllBox{box}
}

// Handle the ``colspan``, ``rowspan`` attributes.
func handleTd(element html.Node, box AllBox, _ gifu, _ string) []AllBox {
	if TypeTableCellBox.IsInstance(box) {
		// HTML 4.01 gives special meaning to colspan=0
		// http://www.w3.org/TR/html401/struct/tables.html#adef-rowspan
		// but HTML 5 removed it
		// http://www.w3.org/TR/html5/tabular-data.html#attr-tdth-colspan
		// rowspan=0 is still there though.

		f := box.TableFields()
		valid, span := integerAttribute(element, "colspan", 1)
		if valid {
			f.colspan = span
		}
		valid, span = integerAttribute(element, "rowspan", 0)
		if valid {
			f.rowspan = span
		}

	}
	return []AllBox{box}
}

// Handle the ``rel`` attribute.
func handleA(element html.Node, box AllBox, _ gifu, _ string) []AllBox {
	box.BaseBox().isAttachment = elementHasLinkType(element, "attachment")
	return []AllBox{box}
}

// Return the base URL for the document.
// See http://www.w3.org/TR/html5/urls.html#document-base-url
//
func findBaseUrl(htmlDocument html.Node, fallbackBaseUrl string) string {
	bases := utils.Iter(htmlDocument, atom.Base)
	if len(bases) > 0 {
		href := strings.TrimSpace(utils.GetAttribute(bases[0], "href"))
		if href != "" {
			return path.Join(fallbackBaseUrl, href)
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
func getHtmlMetadata(wrapperElement html.Node, baseUrl string) HtmlMetadata {
	title := ""
	description := ""
	generator := ""
	keywordsSet := map[string]bool{}
	var authors []string
	created := ""
	modified := ""
	var attachments []Attachment
	for _, element := range utils.Iter(wrapperElement, atom.Title, atom.Meta, atom.Link) {
		switch element.DataAtom {
		case atom.Title:
			if title == "" {
				title = utils.GetChildText(element)
			}
		case atom.Meta:
			name := asciiLower(utils.GetAttribute(element, "name"))
			content := utils.GetAttribute(element, "content")
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
				url := utils.GetUrlAttribute(element, "href", baseUrl, false)
				title := utils.GetAttribute(element, "title")
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
