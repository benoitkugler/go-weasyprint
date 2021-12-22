// Package svg implements parsing of SVG images.
// It transforms SVG text files into an in-memory structure
// that is easy to draw.
// CSS is supported via the style and cascadia packages.
package svg

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/benoitkugler/go-weasyprint/style/parser"
	"github.com/benoitkugler/go-weasyprint/utils"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// SVGImage is a parsed SVG file.
type SVGImage struct {
	defs map[string]*SVGNode

	root *SVGNode // with tag svg
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
	attrs    nodeAttributes
	children []*SVGNode
}

// raw attributes value of a node
// attibutes will be updated in the post processing
// step due to the cascade
type nodeAttributes map[string]string

func newNodeAttributes(attrs []html.Attribute) nodeAttributes {
	out := make(nodeAttributes, len(attrs))
	for _, attr := range attrs {
		out[attr.Key] = attr.Val
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

// Parse parsed the given SVG data. Warnings are
// logged for unsupported elements.
// An error is returned for invalid documents.
// `baseURL` is used as base path for url resources.
func Parse(svg io.Reader, baseURL string) (*SVGImage, error) {
	out, err := buildSVGTree(svg, baseURL)
	if err != nil {
		return nil, err
	}

	// out.postProcess(baseURL)

	return out, nil
}

// Convert from the html representation to an internal,
// simplified form, suitable for post-processing.
// The stylesheets are processed and applied, the values
// of the CSS properties begin stored as attributes
// Inheritable attributes are cascaded and 'inherit' special values are resolved.
func buildSVGTree(svg io.Reader, baseURL string) (*SVGImage, error) {
	root, err := html.Parse(svg)
	if err != nil {
		return nil, err
	}

	// extract the root svg node, which is not
	// always the first one
	iter := utils.NewHtmlIterator(root, atom.Svg)
	if !iter.HasNext() {
		return nil, errors.New("missing <svg> element")
	}
	svgRoot := iter.Next()

	stylesheets := fetchStylesheets(svgRoot)
	normalMatcher, importantMatcher := parseStylesheets(stylesheets, baseURL)

	// build the SVG tree and apply style attribute
	out := SVGImage{defs: map[string]*SVGNode{}}

	// may return nil to discard the node
	var buildTree func(node *html.Node, parentAttrs nodeAttributes) *SVGNode

	buildTree = func(node *html.Node, parentAttrs nodeAttributes) *SVGNode {
		// text is handled by the parent
		// style elements are no longer useful
		if node.Type != html.ElementNode || node.DataAtom == atom.Style {
			return nil
		}

		attrs := newNodeAttributes(node.Attr)
		// Cascade attributes
		for key, value := range parentAttrs {
			if _, isNotInherited := notInheritedAttributes[key]; !isNotInherited {
				if _, isSet := attrs[key]; !isSet {
					attrs[key] = value
				}
			}
		}

		// Apply style
		var normalAttr, importantAttr []declaration
		if styleAttr := attrs["style"]; styleAttr != "" {
			normalAttr, importantAttr = parseDeclarations(parser.Tokenize(styleAttr, false))
		}
		delete(attrs, "style") // not useful anymore

		var allProps []declaration
		allProps = append(allProps, normalMatcher.match((*html.Node)(node))...)
		allProps = append(allProps, normalAttr...)
		allProps = append(allProps, importantMatcher.match((*html.Node)(node))...)
		allProps = append(allProps, importantAttr...)
		for _, d := range allProps {
			attrs[d.property] = strings.TrimSpace(d.value)
		}

		// Replace 'currentColor' value
		for key := range colorAttributes {
			if attrs[key] == "currentColor" {
				if c, has := attrs["color"]; has {
					attrs[key] = c
				} else {
					attrs[key] = "black"
				}
			}
		}

		// Handle 'inherit' values
		for key, value := range attrs {
			if value == "inherit" {
				attrs[key] = parentAttrs[key]
			}
		}

		// recurse
		var children []*SVGNode
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if childSVG := buildTree(child, attrs); childSVG != nil {
				children = append(children, childSVG)
			}
		}

		nodeSVG := &SVGNode{
			tag:      node.Data,
			text:     []byte((*utils.HTMLNode)(node).GetChildrenText()),
			children: children,
			attrs:    attrs,
		}

		if node.Data == "defs" {
			// defs children have been registered
			// and defs elements are not used anymore
			return nil
		}

		// register the node used as "defs"
		if id := attrs["id"]; id != "" {
			out.defs[id] = nodeSVG
		}

		return nodeSVG
	}

	out.root = buildTree((*html.Node)(svgRoot), nil)
	return &out, nil
}

var (
	replacerPreserve   = strings.NewReplacer("\n", " ", "\r", " ", "\t", " ")
	replacerNoPreserve = strings.NewReplacer("\n", "", "\r", "", "\t", " ")
)

// replace newlines by spaces, and merge spaces if not preserved.
func processWhitespace(text []byte, preserveSpace bool) []byte {
	if preserveSpace {
		return []byte(replacerPreserve.Replace(string(text)))
	}
	return []byte(replacerNoPreserve.Replace(string(text)))
}

// handle text node by fixing whitespaces and flattening tails,
// updating node 'children' and 'text'
func (svg *SVGImage) handleText(node *SVGNode, trailingSpace, textRoot bool) bool {
	preserve := node.attrs.spacePreserve()
	node.text = processWhitespace(node.text, preserve)
	if trailingSpace && !preserve {
		node.text = bytes.TrimLeft(node.text, " ")
	}

	if len(node.text) != 0 {
		trailingSpace = bytes.HasSuffix(node.text, []byte{' '})
	}

	var newChildren []*SVGNode
	for _, child := range node.children {
		if child.tag == "tref" {
			// Retrieve the referenced node and get its flattened text
			// and remove the node children.
			id := parseURLFragment(child.attrs["xlink:href"])
			if ref := svg.defs[id]; ref != nil {
				node.text = append(node.text, ref.text...)
			}
			continue
		}

		trailingSpace = svg.handleText(child, trailingSpace, false)

		newChildren = append(newChildren, child)
	}

	if textRoot && len(newChildren) == 0 && !preserve {
		node.text = bytes.TrimRight(node.text, " ")
	}

	node.children = newChildren

	return trailingSpace
}

// finalize the parsing by applying the following steps,
// which require to have seen the whole document :
//	- register defs element
// 	- resolve text reference
//
func (svg *SVGImage) postProcess(baseURL string) {
	svg.registerDefs()

	// svg.postProcessNode(&svg.root)
}

// walk through the tree to register ids
func (svg *SVGImage) registerDefs() {
	// iter := svg.root.Iter()
	// for iter.HasNext() {
	// 	node := iter.Next()
	// 	if id := node.Get("id"); id != "" {
	// 		svg.defs[id] = node
	// 	}
	// }
}

// these attributes are not cascaded
var notInheritedAttributes = utils.NewSet(
	"clip",
	"clip-path",
	"filter",
	"height",
	"id",
	"mask",
	"opacity",
	"overflow",
	"rotate",
	"stop-color",
	"stop-opacity",
	"style",
	"transform",
	"transform-origin",
	"viewBox",
	"width",
	"x",
	"y",
	"dx",
	"dy",
	"href",
)

var colorAttributes = utils.NewSet(
	"fill",
	"flood-color",
	"lighting-color",
	"stop-color",
	"stroke",
)

func (svg *SVGImage) postProcessNode(node *utils.HTMLNode) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		// // Cascade attributes
		// for _, attr := range node.attrs {
		// 	if _, isNotInherited := notInheritedAttributes[key]; !isNotInherited {
		// 		if _, isSet := child.attrs[key]; !isSet {
		// 			child.attrs[key] = value
		// 		}
		// 	}
		// }

		// Apply style attribute
		// var normal_attr, important_attr []declaration
		// style_attr := child.attrs["style"]
		// if style_attr != "" {
		// 	normal_attr, important_attr = parseDeclarations(parser.Tokenize(style_attr, false))
		// }
		// normal_matcher, important_matcher := svg.normalStyle, svg.importantStyle
		// for _, rule :=  range svg.normalStyle {
		// 	rule.
		// }
		// normal = [rule[-1] for rule in normal_matcher.match(wrapper)]
		// important = [rule[-1] for rule in important_matcher.match(wrapper)]
		// declarations_lists = (
		// 	normal, [normal_attr], important, [important_attr])
		// for declarations_list in declarations_lists:
		// 	for declarations in declarations_list:
		// 		for name, value in declarations:
		// 			child.attrib[name] = value.strip()

		// svg.postProcessNode(child)
	}
	// if node.tag == "text" || node.tag == "textPath" || node.tag == "a" {
	// 	svg.handleText(node, true, true)
	// 	return
	// }
}
