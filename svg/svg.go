// Package svg implements parsing of SVG images.
// It transforms SVG text files into an in-memory structure
// that is easy to draw.
// CSS is supported via the style and cascadia packages.
package svg

import (
	"fmt"
	"io"
	"log"
	"math"

	"github.com/benoitkugler/go-weasyprint/backend"
	"github.com/benoitkugler/go-weasyprint/matrix"
)

// convert from an svg tree to the final form

// nodes that are not directly draw but may be referenced
// from other nodes
type definitions struct {
	filters map[string][]filter
}

func newDefinitions() definitions {
	return definitions{
		filters: make(map[string][]filter),
	}
}

type SVGImage struct {
	root *svgNode

	definitions definitions

	// ViewBox is the optional value of the "viewBox" attribute
	ViewBox *Rectangle
}

// Draw draws the parsed SVG image into the given `dst` output,
// with the given `width` and `height`.
func (svg *SVGImage) Draw(dst backend.OutputGraphic, width, height Fl) {
	var ctx drawingDims
	ctx.concreteWidth, ctx.concreteHeight = width, height
	if vb := svg.ViewBox; vb != nil {
		ctx.innerWidth, ctx.innerHeight = vb.Width, vb.Height
	} else {
		ctx.innerWidth, ctx.innerHeight = width, height
	}
	ctx.fontSize = defaultFontSize
	ctx.setupDiagonal()
}

func (svg *SVGImage) drawNode(dst backend.OutputGraphic, node *svgNode, dims drawingDims, fillStroke bool) {
	dims.fontSize = node.attributes.fontSize.resolve(dims.fontSize, dims.fontSize)

	// if fill_stroke:
	// self.stream.push_state()

	// apply filters
	if filters := svg.definitions.filters[node.filterID]; filters != nil {
		applyFilters(dst, filters, node, dims)
	}

	// create sub group for opacity
	opacity := node.attributes.opacity
	var originalDst backend.OutputGraphic
	if fillStroke && 0 <= opacity && opacity < 1 {
		originalDst = dst
		var x, y, width, height Fl = 0, 0, dims.concreteWidth, dims.concreteHeight
		if box, ok := node.resolveBoundingBox(dims, true); ok {
			x, y, width, height = box.X, box.Y, box.Width, box.Height
		}
		dst = dst.AddOpacityGroup(x, y, width, height)
	}

	// apply transform attribute
	// self.transform(node.get('transform'), font_size)

	// apply opacity group and restore original target
	if fillStroke && 0 <= opacity && opacity < 1 {
		originalDst.DrawOpacityGroup(opacity, dst)
		dst = originalDst
	}

	// if fill_stroke:
	// self.stream.pop_state()
}

func applyFilters(dst backend.GraphicTarget, filters []filter, node *svgNode, dims drawingDims) {
	for _, filter := range filters {
		switch filter := filter.(type) {
		case filterOffset:
			var dx, dy Fl
			if filter.isUnitsBBox {
				bbox, _ := node.resolveBoundingBox(dims, true)
				dx = filter.dx.resolve(dims.fontSize, 1) * bbox.Width
				dy = filter.dy.resolve(dims.fontSize, 1) * bbox.Height
			} else {
				dx, dy = dims.point(filter.dx, filter.dy)
			}
			dst.Transform(matrix.New(1, 0, 0, 1, dx, dy))
		case filterBlend:
			// TODO:
			log.Println("blend filter not implemented")
		}
	}
}

// ImageLoader is used to resolve and process image url found in SVG files.
type ImageLoader = func(url string) (backend.Image, error)

// Parse parsed the given SVG data. Warnings are
// logged for unsupported elements.
// An error is returned for invalid documents.
// `baseURL` is used as base path for url resources.
// `imageLoader` is required to handle inner images.
func Parse(svg io.Reader, baseURL string, imageLoader ImageLoader) (*SVGImage, error) {
	out, err := buildSVGTree(svg, baseURL)
	if err != nil {
		return nil, err
	}

	out.imageLoader = imageLoader

	return out.postProcess()
}

type svgNode struct {
	content  drawable
	children []*svgNode
	attributes
}

type drawable interface {
	// draws the node onto `dst` with the given font size
	draw(dst backend.GraphicTarget, attrs *attributes, dims drawingDims)

	// computes the bounding box of the node, or returns false
	// if the node has no valid bounding box, like empty paths.
	boundingBox(attrs *attributes, dims drawingDims) (Rectangle, bool)
}

// drawingDims stores the configuration to use
// when drawing
type drawingDims struct {
	// width and height as requested by the user
	// when calling Draw.
	concreteWidth, concreteHeight Fl

	fontSize Fl

	// either the root viewbox width and height,
	// or the concreteWidth, concreteHeight if
	// no viewBox is provided
	innerWidth, innerHeight Fl

	// cached value of norm(innerWidth, innerHeight) / sqrt(2)
	innerDiagonal Fl
}

// update `innerDiagonal` from `innerWidth` and `innerHeight`.
func (dims *drawingDims) setupDiagonal() {
	dims.innerDiagonal = Fl(math.Hypot(float64(dims.innerWidth), float64(dims.innerHeight)) / math.Sqrt2)
}

// resolve the size of an x/y or width/height couple.
func (dims drawingDims) point(xv, yv value) (x, y Fl) {
	x = xv.resolve(dims.fontSize, dims.innerWidth)
	y = yv.resolve(dims.fontSize, dims.innerHeight)
	return
}

// resolve a length
func (dims drawingDims) length(length value) Fl {
	return length.resolve(dims.fontSize, dims.innerDiagonal)
}

// attributes stores the SVG attributes
// shared by all node types.
type attributes struct {
	transform, clipPath, mask, filterID       string
	marker, markerStart, markerMid, markerEnd string
	stroke                                    string

	fontSize    value
	strokeWidth value

	x, y, width, height value

	opacity Fl // default to 1

	display, visible bool
}

// Build the drawable items by parsing attributes
func (tree *svgContext) postProcess() (*SVGImage, error) {
	vb, err := tree.root.attrs.viewBox()
	if err != nil {
		return nil, err
	}
	var out SVGImage
	out.definitions = newDefinitions()
	out.ViewBox = vb
	out.root, err = tree.processNode(tree.root, out.definitions)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func (tree *svgContext) processNode(node *cascadedNode, defs definitions) (*svgNode, error) {
	var out svgNode

	out.children = make([]*svgNode, len(node.children))
	for i, c := range node.children {
		var err error
		out.children[i], err = tree.processNode(c, defs)
		if err != nil {
			return nil, err
		}
	}

	// actual processing of the node, with the following cases
	//	- graphic element to display -> elementBuilders
	//	- node used as definition
	switch node.tag {
	case "filter":
		filters, err := newFilter(node)
		if err != nil {
			return nil, err
		}
		defs.filters[node.attrs["id"]] = filters
	}

	builder := elementBuilders[node.tag]
	if builder == nil {
		// this node is not drawn, return early
		return &out, nil
	}

	err := node.attrs.parseCommonAttributes(&out.attributes)
	if err != nil {
		return nil, err
	}

	out.content, err = builder(node, tree)
	if err != nil {
		return nil, fmt.Errorf("invalid element %s: %s", node.tag, err)
	}

	return &out, nil
}

func (na nodeAttributes) parseCommonAttributes(out *attributes) error {
	var err error
	out.fontSize, err = na.fontSize()
	if err != nil {
		return err
	}
	out.strokeWidth, err = na.strokeWidth()
	if err != nil {
		return err
	}
	out.opacity, err = na.opacity()
	if err != nil {
		return err
	}
	out.transform = na["transform"]
	out.stroke = na["stroke"] // TODO: preprocess
	out.filterID = parseURLFragment(na["filter"])
	out.clipPath = parseURLFragment(na["clip-path"])
	out.mask = parseURLFragment(na["mask"])

	out.marker = parseURLFragment(na["marker"])
	out.markerStart = parseURLFragment(na["marker-start"])
	out.markerMid = parseURLFragment(na["marker-mid"])
	out.markerEnd = parseURLFragment(na["marker-end"])

	out.x, err = parseValue(na["x"])
	if err != nil {
		return err
	}
	out.y, err = parseValue(na["y"])
	if err != nil {
		return err
	}
	out.width, err = parseValue(na["width"])
	if err != nil {
		return err
	}
	out.height, err = parseValue(na["height"])
	if err != nil {
		return err
	}

	out.display = na.display()
	out.visible = na.visible()
	return nil
}
