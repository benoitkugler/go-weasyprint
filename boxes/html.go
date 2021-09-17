package boxes

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/benoitkugler/go-weasyprint/images"

	"github.com/benoitkugler/go-weasyprint/utils"

	"golang.org/x/net/html/atom"
)

type handlerFunction = func(element *utils.HTMLNode, box Box, getImageFromUri Gifu, baseUrl string) []Box

var htmlHandlers = map[string]handlerFunction{
	"img":      handleImg,
	"embded":   handleEmbed,
	"object":   handleObject,
	"colgroup": handleColgroup,
	"col":      handleCol,
	"th":       handleTd,
	"td":       handleTd,
	"a":        handleA,
}

// HandleElement handle HTML elements that need special care.
func handleElement(element *utils.HTMLNode, box Box, getImageFromUri Gifu, baseUrl string) []Box {
	handler, in := htmlHandlers[box.Box().ElementTag]
	if in {
		ls := handler(element, box, getImageFromUri, baseUrl)
		fmt.Println(box.Box().ElementTag, ls)
		return ls
	} else {
		return []Box{box}
	}
}

// Wrap an image in a replaced box.
//
// That box is either block-level || inline-level, depending on what the
// element should be.
func makeReplacedBox(element *utils.HTMLNode, box Box, image images.Image) Box {
	var newBox Box
	switch box.Box().Style.GetDisplay() {
	case "block", "list-item", "table":
		b := NewBlockReplacedBox(element.Data, box.Box().Style, image)
		newBox = &b
	default:
		// TODO: support images with "display: table-cell"?
		b := NewInlineReplacedBox(element.Data, box.Box().Style, image)
		newBox = &b
	}
	// TODO: check other attributes that need to be copied
	// TODO: find another solution
	newBox.Box().StringSet = box.Box().StringSet
	newBox.Box().BookmarkLabel = box.Box().BookmarkLabel
	return newBox
}

// Handle ``<img>`` elements, return either an image or the alt-text.
// See: http://www.w3.org/TR/html5/embedded-content-1.html#the-img-element
func handleImg(element *utils.HTMLNode, box Box, getImageFromUri Gifu, baseUrl string) []Box {
	src := element.GetUrlAttribute("src", baseUrl, false)
	alt := element.Get("alt")
	if src != "" {
		image := getImageFromUri(src, "")
		if image != nil {
			return []Box{makeReplacedBox(element, box, image)}
		}
		// Invalid image, use the alt-text.
		if alt != "" {
			box.Box().Children = []Box{TextBoxAnonymousFrom(box, alt)}
			return []Box{box}
		}
	} else {
		if alt != "" {
			box.Box().Children = []Box{TextBoxAnonymousFrom(box, alt)}
			return []Box{box}
		}
	}
	// The element represents nothing
	return nil
}

// Handle ``<embed>`` elements, return either an image || nothing.
// See: https://www.w3.org/TR/html5/embedded-content-0.html#the-embed-element
func handleEmbed(element *utils.HTMLNode, box Box, getImageFromUri Gifu, baseUrl string) []Box {
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
func handleObject(element *utils.HTMLNode, box Box, getImageFromUri Gifu, baseUrl string) []Box {
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
func handleColgroup(element *utils.HTMLNode, box Box, _ Gifu, _ string) []Box {
	if TypeTableColumnGroupBox.IsInstance(box) { // leaf
		f := box.Box()

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
			box.Box().Children = children
		}
	}
	return []Box{box}
}

// Handle the ``span`` attribute.
func handleCol(element *utils.HTMLNode, box Box, _ Gifu, _ string) []Box {
	if TypeTableColumnBox.IsInstance(box) { // leaf
		f := box.Box()

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
func handleTd(element *utils.HTMLNode, box Box, _ Gifu, _ string) []Box {
	if TypeTableCellBox.IsInstance(box) {
		// HTML 4.01 gives special meaning to colspan=0
		// http://www.w3.org/TR/html401/struct/tables.html#adef-rowspan
		// but HTML 5 removed it
		// http://www.w3.org/TR/html5/tabular-data.html#attr-tdth-colspan
		// rowspan=0 is still there though.

		f := box.Box()
		valid, span := integerAttribute(*element, "colspan", 1)
		if valid {
			f.Colspan = span
		}
		valid, span = integerAttribute(*element, "rowspan", 0)
		if valid {
			f.Rowspan = span
		}
	}
	return []Box{box}
}

// Handle the ``rel`` attribute.
func handleA(element *utils.HTMLNode, box Box, _ Gifu, _ string) []Box {
	box.Box().IsAttachment = utils.ElementHasLinkType(element, "attachment")
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
				log.Printf("invalid href : %s\n", err)
				return fallbackBaseUrl
			}
			return out
		}
	}
	return fallbackBaseUrl
}
