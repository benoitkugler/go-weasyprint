package svg

import (
	"fmt"
	"strings"

	"github.com/benoitkugler/go-weasyprint/style/parser"
)

// handle painter for fill and stroke values

// painter is either a simple RGBA color,
// or a reference to a more complex `paintServer`
type painter struct {
	// value of the url attribute, refering
	// to a paintServer element
	refID string

	color parser.RGBA

	// if 'false', no painting occurs (not the same as black)
	valid bool
}

// parse a fill or stroke attribute
func newPainter(attr string) (painter, error) {
	attr = strings.TrimSpace(attr)
	if attr == "" || attr == "none" {
		return painter{}, nil
	}

	var out painter
	if strings.HasPrefix(attr, "url(") {
		if i := strings.IndexByte(attr, ')'); i != -1 {
			out.refID = parseURLFragment(attr[:i])
			attr = attr[i+1:] // skip the )
		} else {
			return out, fmt.Errorf("invalid url in color '%s'", attr)
		}
	}

	color := parser.ParseColorString(attr)
	// currentColor has been resolved during tree building
	out.color = color.RGBA
	out.valid = true

	return out, nil
}

// gradient or pattern
type paintServer interface { // TODO:
}

// either linear or radial
type gradientKind interface {
	isGradient()
}

func (gradientLinear) isGradient() {}
func (gradientRadial) isGradient() {}

type gradientLinear struct {
	x1, y1, x2, y2 value
}

func newGradientLinear(node *cascadedNode) (out gradientLinear, err error) {
	out.x1, err = parseValue(node.attrs["x1"])
	if err != nil {
		return out, err
	}
	out.y1, err = parseValue(node.attrs["y1"])
	if err != nil {
		return out, err
	}
	out.x2, err = parseValue(node.attrs["x2"])
	if err != nil {
		return out, err
	}
	out.y2, err = parseValue(node.attrs["y2"])
	if err != nil {
		return out, err
	}

	// default values
	if out.x2.u == 0 {
		out.x2 = value{100, Perc} // 100%
	}

	return out, nil
}

type gradientRadial struct {
	cx, cy, r, fx, fy, fr value
}

func newGradientRadial(node *cascadedNode) (out gradientRadial, err error) {
	cx, cy, r := node.attrs["cx"], node.attrs["cy"], node.attrs["r"]
	if cx == "" {
		cx = "50%"
	}
	if cy == "" {
		cy = "50%"
	}
	if r == "" {
		r = "50%"
	}
	fx, fy, fr := node.attrs["fx"], node.attrs["fy"], node.attrs["fr"]
	if fx == "" {
		fx = cx
	}
	if fy == "" {
		fy = cy
	}

	out.cx, err = parseValue(cx)
	if err != nil {
		return out, err
	}
	out.cy, err = parseValue(cy)
	if err != nil {
		return out, err
	}
	out.r, err = parseValue(r)
	if err != nil {
		return out, err
	}
	out.fx, err = parseValue(fx)
	if err != nil {
		return out, err
	}
	out.fy, err = parseValue(fy)
	if err != nil {
		return out, err
	}
	out.fr, err = parseValue(fr)
	if err != nil {
		return out, err
	}

	return out, nil
}

// gradient specification, prior to resolving units
type gradient struct {
	kind gradientKind

	spreadMethod string // default to "pad"

	positions []value
	colors    []parser.RGBA

	transforms []transform

	isUnitsUserSpace bool
}

// parse a linearGradient or radialGradient node
func newGradient(node *cascadedNode) (out gradient, err error) {
	out.positions = make([]value, len(node.children))
	out.colors = make([]parser.RGBA, len(node.children))
	for i, child := range node.children {
		out.positions[i], err = parseValue(child.attrs["offset"])
		if err != nil {
			return out, err
		}

		sc, has := child.attrs["stop-color"]
		if !has {
			sc = "black"
		}
		stopColor := parser.ParseColorString(sc).RGBA

		stopColor.A, err = parseOpacity(child.attrs["stop-opacity"])
		if err != nil {
			return out, err
		}

		out.colors[i] = stopColor
	}

	out.isUnitsUserSpace = node.attrs["gradientUnits"] == "userSpaceOnUse"
	out.spreadMethod = "pad"
	if sm, has := node.attrs["spreadMethod"]; has {
		out.spreadMethod = sm
	}

	out.transforms, err = parseTransform(node.attrs["gradientTransform"])
	if err != nil {
		return out, err
	}

	switch node.tag {
	case "linearGradient":
		out.kind, err = newGradientLinear(node)
		if err != nil {
			return out, fmt.Errorf("invalid linear gradient: %s", err)
		}
	case "radialGradient":
		out.kind, err = newGradientRadial(node)
		if err != nil {
			return out, fmt.Errorf("invalid radial gradient: %s", err)
		}
	default:
		panic("unexpected node tag " + node.tag)
	}

	return out, nil
}

type pattern struct { // TODO:
}
