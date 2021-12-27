package svg

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/benoitkugler/go-weasyprint/backend"
	"github.com/benoitkugler/go-weasyprint/utils"
)

var elementBuilders = map[string]elementBuilder{
	// "a":        newText,
	"circle": newEllipse, // handle circles
	// "clipPath": newClipPath,
	"ellipse":  newEllipse,
	"image":    newImage,
	"line":     newLine,
	"path":     newPath,
	"polyline": newPolyline,
	"polygon":  newPolygon,
	"rect":     newRect,
	// "svg":      newSvg,
	// "text":     newText,
	// "textPath": newText,
	// "tspan":    newText,
	// "use":      newUse,
}

// function parsing a generic node to build a specialized element
// context holds global data sometimes needed, as well as a cache to
// reduce allocations
type elementBuilder = func(node *cascadedNode, context *svgContext) (drawable, error)

// <line> tag
type line struct {
	x1, y1, x2, y2 value
}

func newLine(node *cascadedNode, _ *svgContext) (drawable, error) {
	var (
		out line
		err error
	)
	out.x1, err = parseValue(node.attrs["x1"])
	if err != nil {
		return nil, err
	}
	out.y1, err = parseValue(node.attrs["y1"])
	if err != nil {
		return nil, err
	}
	out.x2, err = parseValue(node.attrs["x2"])
	if err != nil {
		return nil, err
	}
	out.y2, err = parseValue(node.attrs["y2"])
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (l line) draw(dst backend.GraphicTarget, _ *attributes, dims drawingDims) {
	x1, y1 := dims.point(l.x1, l.y1)
	x2, y2 := dims.point(l.x2, l.y2)
	dst.MoveTo(x1, y1)
	dst.LineTo(x2, y2)
	// TODO:
	// angle = atan2(y2 - y1, x2 - x1)
	// node.vertices = [(x1, y1), (pi - angle, angle), (x2, y2)]
}

// <rect> tag
type rect struct {
	// x, y, width, height are common attributes

	rx, ry value
}

func newRect(node *cascadedNode, _ *svgContext) (drawable, error) {
	rx_, ry_ := node.attrs["rx"], node.attrs["ry"]
	if rx_ == "" {
		rx_ = ry_
	} else if ry_ == "" {
		ry_ = rx_
	}

	var (
		out rect
		err error
	)
	out.rx, err = parseValue(rx_)
	if err != nil {
		return nil, err
	}
	out.ry, err = parseValue(rx_)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (r rect) draw(dst backend.GraphicTarget, attrs *attributes, dims drawingDims) {
	width, height := dims.point(attrs.width, attrs.height)
	if width <= 0 || height <= 0 { // nothing to draw
		return
	}
	x, y := dims.point(attrs.x, attrs.y)
	rx, ry := dims.point(r.rx, r.ry)

	if rx == 0 || ry == 0 { // no border radius
		dst.Rectangle(x, y, width, height)
		return
	}

	if rx > width/2 {
		rx = width / 2
	}
	if ry > height/2 {
		ry = height / 2
	}

	// Inspired by Cairo Cookbook
	// http://cairographics.org/cookbook/roundedrectangles/
	const ARC_TO_BEZIER = 4 * (math.Sqrt2 - 1) / 3
	c1, c2 := ARC_TO_BEZIER*rx, ARC_TO_BEZIER*ry

	dst.MoveTo(x+rx, y)
	dst.LineTo(x+width-rx, y)
	dst.CubicTo(x+width-rx+c1, y, x+width, y+c2, x+width, y+ry)
	dst.LineTo(x+width, y+height-ry)
	dst.CubicTo(
		x+width, y+height-ry+c2, x+width+c1-rx, y+height,
		x+width-rx, y+height)
	dst.LineTo(x+rx, y+height)
	dst.CubicTo(x+rx-c1, y+height, x, y+height-c2, x, y+height-ry)
	dst.LineTo(x, y+ry)
	dst.CubicTo(x, y+ry-c2, x+rx-c1, y, x+rx, y)
	dst.LineTo(x+rx, y)
}

// polyline or polygon
type polyline struct {
	points [][2]Fl // x, y
	close  bool    // true for polygon
}

func newPolyline(node *cascadedNode, _ *svgContext) (drawable, error) {
	return parsePoly(node)
}

func newPolygon(node *cascadedNode, _ *svgContext) (drawable, error) {
	out, err := parsePoly(node)
	out.close = true
	return out, err
}

func parsePoly(node *cascadedNode) (polyline, error) {
	var out polyline

	pts, err := parsePoints(node.attrs["points"], nil)
	if err != nil {
		return out, err
	}

	// "If the attribute contains an odd number of coordinates, the last one will be ignored."
	out.points = make([][2]Fl, len(pts)/2)
	for i := range out.points {
		out.points[i][0] = pts[2*i]
		out.points[i][1] = pts[2*i+1]
	}

	return out, nil
}

func (r polyline) draw(dst backend.GraphicTarget, _ *attributes, _ drawingDims) {
	if len(r.points) == 0 {
		return
	}
	p1, points := r.points[0], r.points[1:]
	dst.MoveTo(p1[0], p1[1])
	// node.vertices = [(x, y)]
	for _, point := range points {
		// angle = atan2(x - x_old, y - y_old)
		// node.vertices.append((pi - angle, angle))
		dst.LineTo(point[0], point[1])
		// node.vertices.append((x, y))
	}
	if r.close {
		dst.LineTo(p1[0], p1[1])
	}
}

// ellipse or circle
type ellipse struct {
	rx, ry, cx, cy value
}

func newEllipse(node *cascadedNode, _ *svgContext) (drawable, error) {
	r_, rx_, ry_ := node.attrs["r"], node.attrs["rx"], node.attrs["ry"]
	if rx_ == "" {
		rx_ = r_
	}
	if ry_ == "" {
		ry_ = r_
	}

	var (
		out ellipse
		err error
	)
	out.rx, err = parseValue(rx_)
	if err != nil {
		return nil, err
	}
	out.ry, err = parseValue(ry_)
	if err != nil {
		return nil, err
	}
	out.cx, err = parseValue(node.attrs["cx"])
	if err != nil {
		return nil, err
	}
	out.cy, err = parseValue(node.attrs["cy"])
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (e ellipse) draw(dst backend.GraphicTarget, _ *attributes, dims drawingDims) {
	rx, ry := dims.point(e.rx, e.ry)
	if rx == 0 || ry == 0 {
		return
	}
	ratioX := rx / math.SqrtPi
	ratioY := ry / math.SqrtPi
	cx, cy := dims.point(e.cx, e.cy)

	dst.MoveTo(cx+rx, cy)
	dst.CubicTo(cx+rx, cy+ratioY, cx+ratioX, cy+ry, cx, cy+ry)
	dst.CubicTo(cx-ratioX, cy+ry, cx-rx, cy+ratioY, cx-rx, cy)
	dst.CubicTo(cx-rx, cy-ratioY, cx-ratioX, cy-ry, cx, cy-ry)
	dst.CubicTo(cx+ratioX, cy-ry, cx+rx, cy-ratioY, cx+rx, cy)
	dst.LineTo(cx+rx, cy)
}

// <path> tag
type path []pathItem

func newPath(node *cascadedNode, context *svgContext) (drawable, error) {
	out, err := context.pathParser.parsePath(node.attrs["d"])
	if err != nil {
		return nil, err
	}

	return path(out), err
}

func (p path) draw(dst backend.GraphicTarget, _ *attributes, _ drawingDims) {
	for _, item := range p {
		item.draw(dst)
	}
}

// <image> tag
type image struct {
	// width, height are common attributes

	img                 backend.Image
	preserveAspectRatio [2]string
}

func newImage(node *cascadedNode, context *svgContext) (drawable, error) {
	baseURL := node.attrs["base"]
	if baseURL == "" {
		baseURL = context.baseURL
	}

	href := node.attrs["href"]
	url, err := utils.SafeUrljoin(baseURL, href, false)
	if err != nil {
		return nil, fmt.Errorf("invalid image source: %s", err)
	}
	img, err := context.imageLoader(url)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %s", err)
	}

	aspectRatio, has := node.attrs["preserveAspectRatio"]
	if !has {
		aspectRatio = "xMidYMid"
	}
	l := strings.Fields(aspectRatio)
	if len(l) > 2 {
		return nil, fmt.Errorf("invalid preserveAspectRatio property: %s", aspectRatio)
	}
	var out image
	copy(out.preserveAspectRatio[:], l)
	out.img = img

	return out, nil
}

func (img image) draw(dst backend.GraphicTarget, _ *attributes, _ drawingDims) {
	// FIXME: support nested images
	log.Println("nested image are not supported")
}

// definitions

type filter interface {
	isFilter()
}

func (filterOffset) isFilter() {}
func (filterBlend) isFilter()  {}

type filterOffset struct {
	dx, dy      value
	isUnitsBBox bool
}

type filterBlend string

// parse a <filter> node
func newFilter(node *cascadedNode) (out []filter, err error) {
	for _, child := range node.children {
		switch child.tag {
		case "feOffset":
			fi := filterOffset{
				isUnitsBBox: node.attrs["primitiveUnits"] == "objectBoundingBox",
			}
			fi.dx, err = parseValue(child.attrs["dx"])
			if err != nil {
				return nil, err
			}
			fi.dy, err = parseValue(child.attrs["dy"])
			if err != nil {
				return nil, err
			}
			out = append(out, fi)
		case "feBlend":
			fi := filterBlend("normal")
			if mode, has := child.attrs["mode"]; has {
				fi = filterBlend(mode)
			}
			out = append(out, fi)
		default:
			log.Printf("unsupported filter element: %s", child.tag)
		}
	}

	return out, nil
}

// clipPath is a container for
// graphic nodes, which will use as clipping path,
// that is drawn but not stroked nor filled.
type clipPath struct {
	children    []*svgNode
	isUnitsBBox bool
}

func newClipPath(node *cascadedNode, children []*svgNode) clipPath {
	return clipPath{
		children:    children,
		isUnitsBBox: node.attrs["clipPathUnits"] == "objectBoundingBox",
	}
}

// mask is a container for shape that will
// be used as an alpha mask
type mask struct {
	svgNode
	isUnitsBBox bool
}

func newMask(node *cascadedNode, children []*svgNode) (mask, error) {
	out := mask{
		svgNode: svgNode{
			children: children,
		},
		isUnitsBBox: node.attrs["maskUnits"] == "objectBoundingBox",
	}
	err := node.attrs.parseCommonAttributes(&out.svgNode.attributes)
	if err != nil {
		return mask{}, err
	}
	// default values
	if out.x.u == 0 {
		out.x = value{-10, Perc} // -10%
	}
	if out.y.u == 0 {
		out.y = value{-10, Perc} // -10%
	}
	if out.width.u == 0 {
		out.width = value{120, Perc} // 120%
	}
	if out.height.u == 0 {
		out.height = value{120, Perc} // 120%
	}
	return out, err
}
