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
	normalStyle, importantStyle matcher

	// keys are #<id>
	defs map[string]SVGNode

	root SVGNode
}

// type nodeAttributes struct {
// 	viewBox *[4]Fl

// 	markerPosition [3]string // [start, mid, end] should default to marker
// 	marker         string

// 	filter   string
// 	clipPath string
// 	mask     string

// 	transform, preserveAspectRatio, orient, overflow string

// 	strokeDasharray []value

// 	fontSize                  value
// 	strokeDashoffset          value
// 	markerWidth, markerHeight value
// 	width, height             value

// 	opacity Fl // default to 1

// 	fillEvenOdd          bool
// 	markerUnitsUserSpace bool
// 	noDisplay, noVisible bool

// 	rawArgs map[string]string // parsing is deferred
// }

// SVGNode is a node in an SVG document.
type SVGNode struct {
	tag      string
	text     []byte
	children []SVGNode
	nodeAttributes
}

// raw attributes value of a node
// attibutes will be updated in the post processing
// step due to the cascade
type nodeAttributes map[string]string

func newNodeAttributes(attrs []xml.Attr) nodeAttributes {
	out := make(nodeAttributes, len(attrs))
	for _, attr := range attrs {
		out[attr.Name.Local] = attr.Value
	}
	return out
}

func (na nodeAttributes) viewBox() ([4]float64, error) {
	attrValue := na["viewBox"]
	return parseViewbox(attrValue)
}

func (na nodeAttributes) filter() string {
	attrValue := na["filter"]
	return parseURLFragment(attrValue)
}

func (na nodeAttributes) clipPath() string {
	attrValue := na["clip-path"]
	return parseURLFragment(attrValue)
}

func (na nodeAttributes) mask() string {
	attrValue := na["mask"]
	return parseURLFragment(attrValue)
}

func (na nodeAttributes) marker() string {
	attrValue := na["marker"]
	return parseURLFragment(attrValue)
}

func (na nodeAttributes) markerStart() string {
	attrValue := na["marker-start"]
	return parseURLFragment(attrValue)
}

func (na nodeAttributes) markerMid() string {
	attrValue := na["marker-mid"]
	return parseURLFragment(attrValue)
}

func (na nodeAttributes) markerEnd() string {
	attrValue := na["marker-end"]
	return parseURLFragment(attrValue)
}

func (na nodeAttributes) fontSize() (value, error) {
	attrValue := na["font-size"]
	return parseValue(attrValue)
}

func (na nodeAttributes) width() (value, error) {
	attrValue := na["width"]
	return parseValue(attrValue)
}

func (na nodeAttributes) height() (value, error) {
	attrValue := na["height"]
	return parseValue(attrValue)
}

func (na nodeAttributes) markerWidth() (value, error) {
	attrValue := na["markerWidth"]
	return parseValue(attrValue)
}

func (na nodeAttributes) markerHeight() (value, error) {
	attrValue := na["markerHeight"]
	return parseValue(attrValue)
}

func (na nodeAttributes) markerUnitsUserSpace() bool {
	attrValue := na["markerUnits"]
	return attrValue == "userSpaceOnUse"
}

func (na nodeAttributes) opacity() (float64, error) {
	if attrValue, has := na["opacity"]; has {
		return strconv.ParseFloat(attrValue, 64)
	}
	return 1, nil
}

func (na nodeAttributes) noDisplay() bool {
	attrValue := na["display"]
	return attrValue == "none"
}

func (na nodeAttributes) noVisible() bool {
	attrValue := na["visibility"]
	noVisible := attrValue == "hidden"
	return na.noDisplay() || noVisible
}

func (na nodeAttributes) strokeDasharray() ([]value, error) {
	attrValue := na["stroke-dasharray"]
	return parseFloatList(attrValue)
}

func (na nodeAttributes) strokeDashoffset() (value, error) {
	attrValue := na["stroke-dashoffset"]
	return parseValue(attrValue)
}

func (na nodeAttributes) fillEvenOdd() bool {
	attrValue := na["fill-rull"]
	return attrValue == "evenodd"
}

func (na nodeAttributes) spacePreserve() bool {
	return na["space"] == "preserve"
}

// func parseNodeAttributes(attrs []xml.Attr) (node nodeAttributes, err error) {
// 	node.opacity = 1
// 	var noDisplay, noVisible bool
// 	node.rawArgs = make(map[string]string)
// 	for _, attr := range attrs {
// 		switch attr.Name.Local {
// 		case "viewBox":
// 			var vb [4]Fl
// 			vb, err = parseViewbox(attr.Value)
// 			node.viewBox = &vb
// 		case "filter":
// 			node.filter = parseURLFragment(attr.Value)
// 		case "clip-path":
// 			node.clipPath = parseURLFragment(attr.Value)
// 		case "mask":
// 			node.mask = parseURLFragment(attr.Value)
// 		case "marker":
// 			node.marker = parseURLFragment(attr.Value)
// 		case "marker-start":
// 			node.markerPosition[0] = parseURLFragment(attr.Value)
// 		case "marker-mid":
// 			node.markerPosition[1] = parseURLFragment(attr.Value)
// 		case "marker-end":
// 			node.markerPosition[2] = parseURLFragment(attr.Value)
// 		case "transform":
// 			node.transform = attr.Value
// 		case "orient":
// 			node.orient = attr.Value
// 		case "overflow":
// 			node.overflow = attr.Value
// 		case "font-size":
// 			node.fontSize, err = parseValue(attr.Value)
// 		case "width":
// 			node.width, err = parseValue(attr.Value)
// 		case "height":
// 			node.height, err = parseValue(attr.Value)
// 		case "markerWidth":
// 			node.markerWidth, err = parseValue(attr.Value)
// 		case "markerHeight":
// 			node.markerHeight, err = parseValue(attr.Value)
// 		case "markerUnits":
// 			node.markerUnitsUserSpace = attr.Value == "userSpaceOnUse"
// 		case "opacity":
// 			node.opacity, err = strconv.ParseFloat(attr.Value, 64)
// 		case "display":
// 			noDisplay = attr.Value == "none"
// 		case "visibility":
// 			noVisible = attr.Value == "hidden"
// 		case "stroke-dasharray":
// 			node.strokeDasharray, err = parseFloatList(attr.Value)
// 		case "stroke-dashoffset":
// 			node.strokeDashoffset, err = parseValue(attr.Value)
// 		case "fill-rull":
// 			node.fillEvenOdd = attr.Value == "evenodd"
// 		default:
// 			node.rawArgs[attr.Name.Local] = attr.Value
// 		}
// 		if err != nil {
// 			return nodeAttributes{}, err
// 		}
// 	}
// 	node.noDisplay = noDisplay
// 	node.noVisible = noDisplay || noVisible

// 	return node, nil
// }

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
	pr := xmlParser{defs: make(map[string]SVGNode)}
	err := pr.parse(svg)
	if err != nil {
		return nil, err
	}

	if pr.root.tag != "svg" {
		return nil, fmt.Errorf("invalid root tag: %s", pr.root.tag)
	}

	var out SVGImage
	out.root = *pr.root
	out.normalStyle, out.importantStyle = parseStylesheets(pr.stylesheets, baseURL)
	out.defs = pr.defs

	out.postProcessNode(&out.root)

	return &out, nil
}

type xmlParser struct {
	root        *SVGNode
	defs        map[string]SVGNode
	stylesheets [][]byte // raw style sheets
}

func (pr *xmlParser) parse(svg io.Reader) error {
	decoder := xml.NewDecoder(svg)
	decoder.CharsetReader = charset.NewReaderLabel
	err := decoder.DecodeElement(&pr, nil)
	if err != nil {
		return err
	}
	return nil
}

func (pr *xmlParser) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	pr.root, err = pr.unmarshalXML(d, start)
	return err
}

func (pr *xmlParser) handleDefs(element *SVGNode) *SVGNode {
	if element.tag == "defs" {
		// save the defined elements and return nil
		for _, child := range element.children {
			pr.defs[child.nodeAttributes["id"]] = child
		}
		return nil
	}
	return element
}

func (pr *xmlParser) unmarshalXML(d *xml.Decoder, start xml.StartElement) (node *SVGNode, err error) {
	// special case for <style>
	isCSS, err := pr.handleStyleElement(d, start)
	if err != nil {
		return nil, err
	}
	if isCSS {
		return nil, nil
	}

	// start by handling the new element
	node = new(SVGNode)
	node.tag = start.Name.Local
	node.nodeAttributes = newNodeAttributes(start.Attr)

	// then process the inner content: text or kid element
	for {
		next, err := d.Token()
		if err != nil {
			return nil, err
		}
		// Token is one of StartElement, EndElement, CharData, Comment, ProcInst, or Directive
		switch next := next.(type) {
		case xml.CharData:
			// handle text and keep going
			node.text = append(node.text, next...)
		case xml.EndElement:
			// closing current element: return after processing
			node = pr.handleDefs(node)
			return node, nil
		case xml.StartElement:
			// new kid: recurse and keep going for other kids or text
			kid, err := pr.unmarshalXML(d, next)
			if err != nil {
				return nil, err
			}
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
