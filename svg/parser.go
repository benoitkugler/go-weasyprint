// Package svg implements parsing of SVG images.
// It transforms SVG text files into an in-memory structure
// that is easy to draw.
// CSS is supported via the style and cascadia packages.
package svg

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"

	"golang.org/x/net/html/charset"
)

// SVGImage is a parsed SVG file.
type SVGImage struct {
	root SVGNode

	normalStyle, importantStyle matcher
}

// all fields are optional
type nodeAttributes struct {
	viewBox  *[4]Fl
	filter   string
	fontSize value
	clipPath string
	mask     string

	marker                    string
	markerPosition            [3]string // [start, mid, end] should default to marker
	markerWidth, markerHeight value
	markerUnitsUserSpace      bool

	transform, preserveAspectRatio, orient, overflow string

	strokeDasharray  []value
	strokeDashoffset value
	fillEvenOdd      bool

	opacity              Fl // default to 1
	width, height        value
	noDisplay, noVisible bool
}

// SVGNode is a node in an SVG document.
type SVGNode struct {
	nodeAttributes
	tag      string
	text     []byte
	children []SVGNode
}

func parseNodeAttributes(attrs []xml.Attr) (node nodeAttributes, err error) {
	node.opacity = 1
	var noDisplay, noVisible bool
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "viewBox":
			var vb [4]Fl
			vb, err = parseViewbox(attr.Value)
			node.viewBox = &vb
		case "filter":
			node.filter = parseURLFragment(attr.Value)
		case "clip-path":
			node.clipPath = parseURLFragment(attr.Value)
		case "mask":
			node.mask = parseURLFragment(attr.Value)
		case "marker":
			node.marker = parseURLFragment(attr.Value)
		case "marker-start":
			node.markerPosition[0] = parseURLFragment(attr.Value)
		case "marker-mid":
			node.markerPosition[1] = parseURLFragment(attr.Value)
		case "marker-end":
			node.markerPosition[2] = parseURLFragment(attr.Value)
		case "transform":
			node.transform = attr.Value
		case "orient":
			node.orient = attr.Value
		case "overflow":
			node.overflow = attr.Value
		case "font-size":
			node.fontSize, err = parseValue(attr.Value)
		case "width":
			node.width, err = parseValue(attr.Value)
		case "height":
			node.height, err = parseValue(attr.Value)
		case "markerWidth":
			node.markerWidth, err = parseValue(attr.Value)
		case "markerHeight":
			node.markerHeight, err = parseValue(attr.Value)
		case "markerUnits":
			node.markerUnitsUserSpace = attr.Value == "userSpaceOnUse"
		case "opacity":
			node.opacity, err = strconv.ParseFloat(attr.Value, 64)
		case "display":
			noDisplay = attr.Value == "none"
		case "visibility":
			noVisible = attr.Value == "hidden"
		case "stroke-dasharray":
			node.strokeDasharray, err = parseFloatList(attr.Value)
		case "stroke-dashoffset":
			node.strokeDashoffset, err = parseValue(attr.Value)
		case "fill-rull":
			node.fillEvenOdd = attr.Value == "evenodd"
		}
		if err != nil {
			return nodeAttributes{}, err
		}
	}
	node.noDisplay = noDisplay
	node.noVisible = noDisplay || noVisible

	return node, nil
}

func parseViewbox(attr string) ([4]Fl, error) {
	points, err := parsePoints(attr)
	if err != nil {
		return [4]Fl{}, err
	}
	if len(points) != 4 {
		return [4]Fl{}, fmt.Errorf("expected 4 numbers for viewbox, got %s", attr)
	}
	return [4]Fl{points[0], points[1], points[2], points[3]}, nil
}

// Parse parsed the given SVG data. Warnings are
// logged for unsupported elements.
// An error is returned for invalid documents.
// `baseURL` is used as base path for url resources.
func Parse(svg io.Reader, baseURL string) (*SVGImage, error) {
	var pr svgParser
	err := pr.parse(svg)
	if err != nil {
		return nil, err
	}
	var out SVGImage
	out.root = *pr.root
	out.normalStyle, out.importantStyle = parseStylesheets(pr.stylesheets, baseURL)

	return &out, nil
}

type svgParser struct {
	root        *SVGNode
	stylesheets [][]byte
}

func (pr *svgParser) parse(svg io.Reader) error {
	decoder := xml.NewDecoder(svg)
	decoder.CharsetReader = charset.NewReaderLabel
	err := decoder.DecodeElement(&pr, nil)
	if err != nil {
		return err
	}
	if pr.root.tag != "svg" {
		return fmt.Errorf("invalid root tag: %s", pr.root.tag)
	}
	return nil
}

func (pr *svgParser) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	pr.root, pr.stylesheets, err = unmarshalXML(d, start)
	return err
}

func unmarshalXML(d *xml.Decoder, start xml.StartElement) (node *SVGNode, stylesheets [][]byte, err error) {
	// special case for <style>
	css, err := handleStyleElement(d, start)
	if err != nil {
		return nil, nil, err
	}
	if css != nil {
		return nil, [][]byte{css}, nil
	}

	// start by handling the new element
	node = new(SVGNode)
	node.tag = start.Name.Local
	node.nodeAttributes, err = parseNodeAttributes(start.Attr)
	if err != nil {
		return nil, nil, err
	}

	// then process the inner content: text or kid element
	for {
		next, err := d.Token()
		if err != nil {
			return nil, nil, err
		}
		// Token is one of StartElement, EndElement, CharData, Comment, ProcInst, or Directive
		switch next := next.(type) {
		case xml.CharData:
			// handle text and keep going
			node.text = append(node.text, next...)
		case xml.EndElement:
			// closing current element: return after processing
			return node, stylesheets, nil
		case xml.StartElement:
			// new kid: recurse and keep going for other kids or text
			kid, css, err := unmarshalXML(d, next)
			if err != nil {
				return nil, nil, err
			}
			stylesheets = append(stylesheets, css...)
			if kid != nil {
				node.children = append(node.children, *kid)
			}
		default:
			// ignored, just keep going
		}
	}
}

// // Parse parsed the given SVG data. Warnings are
// // logged for unsupported elements.
// // An error is returned for invalid documents.
// func Parse(svg io.Reader) (*SVGImage, error) {
// 	decoder := xml.NewDecoder(svg)
// 	decoder.CharsetReader = charset.NewReaderLabel
// 	seenTag := false
// 	for {
// 		t, err := decoder.Token()
// 		if err != nil {
// 			if err == io.EOF {
// 				if !seenTag {
// 					return nil, errors.New("invalid svg xml icon")
// 				}
// 				break
// 			}
// 			return nil, err
// 		}
// 		// Inspect the type of the XML token
// 		switch se := t.(type) {
// 		case xml.StartElement:
// 			seenTag = true
// 			// Reads all recognized style attributes from the start element
// 			// and places it on top of the styleStack
// 			err = cursor.pushStyle(se.Attr)
// 			if err != nil {
// 				return icon, err
// 			}
// 			err = cursor.readStartElement(se)
// 			if err != nil {
// 				return icon, err
// 			}
// 		case xml.EndElement:
// 			// pop style
// 			cursor.styleStack = cursor.styleStack[:len(cursor.styleStack)-1]
// 			switch se.Name.Local {
// 			case "g":
// 				if cursor.inDefs {
// 					cursor.currentDef = append(cursor.currentDef, definition{
// 						Tag: "endg",
// 					})
// 				}
// 			case "title":
// 				cursor.inTitleText = false
// 			case "desc":
// 				cursor.inDescText = false
// 			case "defs":
// 				if len(cursor.currentDef) > 0 {
// 					cursor.icon.defs[cursor.currentDef[0].ID] = cursor.currentDef
// 					cursor.currentDef = make([]definition, 0)
// 				}
// 				cursor.inDefs = false
// 			case "radialGradient", "linearGradient":
// 				cursor.inGrad = false
// 			}
// 		case xml.CharData:
// 			if cursor.inTitleText {
// 				icon.Titles[len(icon.Titles)-1] += string(se)
// 			}
// 			if cursor.inDescText {
// 				icon.Descriptions[len(icon.Descriptions)-1] += string(se)
// 			}
// 		}
// 	}
// 	return icon, nil
// }
