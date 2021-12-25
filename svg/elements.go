package svg

import (
	"fmt"
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

func (l line) draw(dst backend.GraphicTarget, _ *attributes, ctx drawingContext) {
	x1, y1 := ctx.point(l.x1, l.y1)
	x2, y2 := ctx.point(l.x2, l.y2)
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

func (r rect) draw(dst backend.GraphicTarget, attrs *attributes, ctx drawingContext) {
	width, height := ctx.point(attrs.width, attrs.height)
	if width <= 0 || height <= 0 { // nothing to draw
		return
	}
	x, y := ctx.point(attrs.x, attrs.y)
	rx, ry := ctx.point(r.rx, r.ry)

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

func (r polyline) draw(dst backend.GraphicTarget, attrs *attributes, ctx drawingContext) {
	if len(r.points) == 0 {
		return
	}
	p1, points := r.points[0], r.points[1:]
	dst.MoveTo(p1[0], p1[1])
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

// <path> tag
type path []pathItem

func newPath(node *cascadedNode, context *svgContext) (drawable, error) {
	out, err := context.pathParser.parsePath(node.attrs["d"])
	if err != nil {
		return nil, err
	}

	return out, err
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
