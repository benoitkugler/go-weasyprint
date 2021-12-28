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
	filters      map[string][]filter
	clipPaths    map[string]clipPath
	masks        map[string]mask
	paintServers map[string]paintServer
}

func newDefinitions() definitions {
	return definitions{
		filters:      make(map[string][]filter),
		clipPaths:    make(map[string]clipPath),
		masks:        make(map[string]mask),
		paintServers: make(map[string]paintServer),
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

	applyTransform(dst, node.attributes.transforms, dims)

	// clip
	if cp, has := svg.definitions.clipPaths[node.clipPathID]; has {
		applyClipPath(dst, cp, node, dims)
	}

	// manage display and visibility
	display := node.attributes.display
	visible := node.attributes.visible

	// draw the node itself.
	if visible && node.content != nil {
		node.content.draw(dst, &node.attributes, dims)
	}

	// then recurse
	if display {
		for _, child := range node.children {
			svg.drawNode(dst, child, dims, fillStroke)
		}
	}

	// apply mask
	if ma, has := svg.definitions.masks[node.maskID]; has {
		applyMask(dst, ma, node, dims)
	}

	// apply opacity group and restore original target
	if fillStroke && 0 <= opacity && opacity < 1 {
		originalDst.DrawOpacityGroup(opacity, dst)
		dst = originalDst
	}

	// if fill_stroke:
	// self.stream.pop_state()
}

func applyTransform(dst backend.GraphicTarget, transforms []transform, dims drawingDims) {
	if len(transforms) == 0 { // do not apply a useless identity transform
		return
	}

	// aggregate the transformations
	mat := matrix.Identity()
	for _, transform := range transforms {
		transform.applyTo(&mat, dims)
	}
	if mat.Determinant() != 0 {
		dst.Transform(mat)
	}
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

func applyClipPath(dst backend.GraphicTarget, clipPath clipPath, node *svgNode, dims drawingDims) {
	// old_ctm = self.stream.ctm
	if clipPath.isUnitsBBox {
		x, y := dims.point(node.attributes.x, node.attributes.y)
		width, height := dims.point(node.attributes.width, node.attributes.height)
		dst.Transform(matrix.New(width, 0, 0, height, x, y))
	}

	// FIXME:
	log.Println("applying clip path is not supported")
	// clip_path._etree_node.tag = 'g'
	// self.draw_node(clip_path, font_size, fill_stroke=False)

	// At least set the clipping area to an empty path, so that it’s
	// totally clipped when the clipping path is empty.
	dst.Rectangle(0, 0, 0, 0)
	dst.Clip(false)
	// new_ctm = self.stream.ctm
	// if new_ctm.determinant:
	//     self.stream.transform(*(old_ctm @ new_ctm.invert).values)
}

func applyMask(dst backend.GraphicTarget, mask mask, node *svgNode, dims drawingDims) {
	// mask._etree_node.tag = 'g'

	widthRef, heightRef := dims.innerWidth, dims.innerHeight
	if mask.isUnitsBBox {
		widthRef, heightRef = dims.point(node.width, node.height)
	}

	x := mask.x.resolve(dims.fontSize, widthRef)
	y := mask.y.resolve(dims.fontSize, heightRef)
	width := mask.width.resolve(dims.fontSize, widthRef)
	height := mask.height.resolve(dims.fontSize, heightRef)

	mask.x = value{x, Px}
	mask.y = value{y, Px}
	mask.width = value{width, Px}
	mask.height = value{height, Px}

	if mask.isUnitsBBox {
		x, y, width, height = 0, 0, widthRef, heightRef
	} else {
		// TODO: update viewbox if needed
		//     mask.attrib['viewBox'] = f'{x} {y} {width} {height}'
	}

	// FIXME:
	log.Println("mask not implemented")
	// alpha_stream = svg.stream.add_group([x, y, width, height])
	// state = pydyf.Dictionary({
	//     'Type': '/ExtGState',
	//     'SMask': pydyf.Dictionary({
	//         'Type': '/Mask',
	//         'S': '/Luminance',
	//         'G': alpha_stream,
	//     }),
	//     'ca': 1,
	//     'AIS': 'false',
	// })
	// svg.stream.set_state(state)

	// svg_stream = svg.stream
	// svg.stream = alpha_stream
	// svg.draw_node(mask, font_size)
	// svg.stream = svg_stream
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

type box struct {
	x, y, width, height value
}

// attributes stores the SVG attributes
// shared by all node types in the final rendering tree
type attributes struct {
	transforms []transform

	clipPathID, maskID, filterID              string
	marker, markerStart, markerMid, markerEnd string

	dashArray []value

	stroke, fill painter

	box

	fontSize    value
	strokeWidth value

	dashOffset value

	opacity, strokeOpacity, fillOpacity Fl // default to 1

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
	children := make([]*svgNode, len(node.children))
	for i, c := range node.children {
		var err error
		children[i], err = tree.processNode(c, defs)
		if err != nil {
			return nil, err
		}
	}

	// actual processing of the node, with the following cases
	//	- node used as definition, extracted from the svg tree
	//	- graphic element to display -> elementBuilders

	id := node.attrs["id"]
	switch node.tag {
	case "filter":
		filters, err := newFilter(node)
		if err != nil {
			return nil, err
		}
		defs.filters[id] = filters
		return nil, nil
	case "clipPath":
		defs.clipPaths[id] = newClipPath(node, children)
		return nil, nil
	case "mask":
		ma, err := newMask(node, children)
		if err != nil {
			return nil, err
		}
		defs.masks[id] = ma
		return nil, nil
	case "linearGradient", "radialGradient":
		grad, err := newGradient(node)
		if err != nil {
			return nil, err
		}
		defs.paintServers[id] = grad
		return nil, nil
	case "defs":
		// children has been processed and registred,
		// so we discard the node, which is not needed anymore
		return nil, nil
	}

	out := svgNode{children: children}
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

func (na nodeAttributes) parseBox(out *box) (err error) {
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

	return nil
}

func (na nodeAttributes) parseCommonAttributes(out *attributes) error {
	err := na.parseBox(&out.box)
	if err != nil {
		return err
	}
	out.fontSize, err = na.fontSize()
	if err != nil {
		return err
	}
	out.strokeWidth, err = na.strokeWidth()
	if err != nil {
		return err
	}

	out.opacity, err = parseOpacity(na["opacity"])
	if err != nil {
		return err
	}
	out.strokeOpacity, err = parseOpacity(na["stroke-opacity"])
	if err != nil {
		return err
	}
	out.fillOpacity, err = parseOpacity(na["fill-opacity"])
	if err != nil {
		return err
	}

	out.transforms, err = parseTransform(na["transform"])
	if err != nil {
		return err
	}

	out.stroke, err = newPainter(na["stroke"])
	if err != nil {
		return err
	}
	out.fill, err = newPainter(na["fill"])
	if err != nil {
		return err
	}

	out.dashOffset, err = parseValue(na["stroke-dashoffset"])
	if err != nil {
		return err
	}
	out.dashArray, err = parseValues(na["stroke-dasharray"])
	if err != nil {
		return err
	}

	out.filterID = parseURLFragment(na["filter"])
	out.clipPathID = parseURLFragment(na["clip-path"])
	out.maskID = parseURLFragment(na["mask"])

	out.marker = parseURLFragment(na["marker"])
	out.markerStart = parseURLFragment(na["marker-start"])
	out.markerMid = parseURLFragment(na["marker-mid"])
	out.markerEnd = parseURLFragment(na["marker-end"])

	out.display = na.display()
	out.visible = na.visible()
	return nil
}
