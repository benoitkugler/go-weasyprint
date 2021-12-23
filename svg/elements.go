package svg

var elementBuilders = map[string]elementBuilder{
	// "a":        newText,
	"circle": newEllipse, // handle circles
	// "clipPath": newClipPath,
	"ellipse": newEllipse,
	// "image":    newImage,
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

// function parsing a generic node
// to build a specialized element
type elementBuilder = func(*cascadedNode) (drawable, error)

// <line> tag
type line struct {
	x1, y1, x2, y2 value
}

func newLine(node *cascadedNode) (drawable, error) {
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

// <rect> tag
type rect struct {
	// x, y, width, height are common attributes
	rx, ry value
}

func newRect(node *cascadedNode) (drawable, error) {
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

// polyline or polygon
type polyline struct {
	points [][2]Fl // x, y
	close  bool    // true for polygon
}

func newPolyline(node *cascadedNode) (drawable, error) {
	return parsePoly(node)
}

func newPolygon(node *cascadedNode) (drawable, error) {
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

// ellipse or circle
type ellipse struct {
	rx, ry, cx, cy value
}

func newEllipse(node *cascadedNode) (drawable, error) {
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

type path []pathItem

func newPath(node *cascadedNode) (drawable, error) {
	var (
		out path
		err error
	)
	out, err = parsePath(node.attrs["d"])
	if err != nil {
		return nil, err
	}

	return out, err
}
