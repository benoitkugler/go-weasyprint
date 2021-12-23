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

// convert from html nodes to an intermediate svg tree

// svgTree is a parsed SVG file,
// where CSS has been applied, and text has been processed
type svgTree struct {
	defs map[string]*cascadedNode

	root *cascadedNode // with tag svg
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

// cascadedNode is a node in an SVG document.
type cascadedNode struct {
	tag      string
	text     []byte
	attrs    nodeAttributes
	children []*cascadedNode
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

func (na nodeAttributes) viewBox() ([4]Fl, error) {
	attrValue := na["viewBox"]
	return parseViewbox(attrValue)
}

func (na nodeAttributes) fontSize() (value, error) {
	attrValue, has := na["font-size"]
	if !has {
		attrValue = "1em"
	}
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

func (na nodeAttributes) opacity() (Fl, error) {
	if attrValue, has := na["opacity"]; has {
		out, err := strconv.ParseFloat(attrValue, 32)
		return Fl(out), err
	}
	return 1, nil
}

func (na nodeAttributes) display() bool {
	attrValue := na["display"]
	return attrValue != "none"
}

func (na nodeAttributes) visible() bool {
	attrValue := na["visibility"]
	visible := attrValue != "hidden"
	return na.display() && visible
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

// walk the tree to extract content needed to build the SVG tree
func fetchStyleAndTextRefs(root *utils.HTMLNode) ([][]byte, map[string][]byte) {
	var (
		stylesheets [][]byte
		trefs       = make(map[string][]byte)
	)
	iter := root.Iter()
	for iter.HasNext() {
		node := iter.Next()
		if css := handleStyleElement(node); len(css) != 0 {
			stylesheets = append(stylesheets, css)
			continue
		}

		// register text refs
		if id := node.Get("id"); id != "" {
			trefs[id] = node.GetChildrenText()
		}
	}
	return stylesheets, trefs
}

// Convert from the html representation to an internal,
// simplified form, suitable for post-processing.
// The stylesheets are processed and applied, the values
// of the CSS properties begin stored as attributes
// Inheritable attributes are cascaded and 'inherit' special values are resolved.
func buildSVGTree(svg io.Reader, baseURL string) (*svgTree, error) {
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

	stylesheets, trefs := fetchStyleAndTextRefs(svgRoot)
	normalMatcher, importantMatcher := parseStylesheets(stylesheets, baseURL)

	// build the SVG tree and apply style attribute
	out := svgTree{defs: map[string]*cascadedNode{}}

	// may return nil to discard the node
	var buildTree func(node *html.Node, parentAttrs nodeAttributes) *cascadedNode

	buildTree = func(node *html.Node, parentAttrs nodeAttributes) *cascadedNode {
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

		nodeSVG := &cascadedNode{
			tag:   node.Data,
			text:  (*utils.HTMLNode)(node).GetChildrenText(),
			attrs: attrs,
		}

		// recurse
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if childSVG := buildTree(child, attrs); childSVG != nil {
				nodeSVG.children = append(nodeSVG.children, childSVG)
			}
		}

		// Fix text in text tags
		if node.Data == "text" || node.Data == "textPath" || node.Data == "a" {
			handleText(nodeSVG, true, true, trefs)
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
func handleText(node *cascadedNode, trailingSpace, textRoot bool, trefs map[string][]byte) bool {
	preserve := node.attrs.spacePreserve()
	node.text = processWhitespace(node.text, preserve)
	if trailingSpace && !preserve {
		node.text = bytes.TrimLeft(node.text, " ")
	}

	if len(node.text) != 0 {
		trailingSpace = bytes.HasSuffix(node.text, []byte{' '})
	}

	var newChildren []*cascadedNode
	for _, child := range node.children {
		if child.tag == "tref" {
			// Retrieve the referenced node and get its flattened text
			// and remove the node children.
			id := parseURLFragment(child.attrs["href"])
			node.text = append(node.text, trefs[id]...)
			continue
		}

		trailingSpace = handleText(child, trailingSpace, false, trefs)

		newChildren = append(newChildren, child)
	}

	if textRoot && len(newChildren) == 0 && !preserve {
		node.text = bytes.TrimRight(node.text, " ")
	}

	node.children = newChildren

	return trailingSpace
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
