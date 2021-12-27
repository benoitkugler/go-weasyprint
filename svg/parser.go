package svg

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/benoitkugler/go-weasyprint/utils"
)

// provide low-level functions to read basic SVG data

type Fl = utils.Fl

var root2 = math.Sqrt(2)

type unit uint8

// Units supported.
const (
	_ unit = iota
	Px
	Cm
	Mm
	Pt
	In
	Q
	Pc

	Perc // Special case : percentage (%) relative to the viewbox
	Em   // Special case : relative to the font size
	Ex   // idem
)

var units = [...]string{Px: "px", Cm: "cm", Mm: "mm", Pt: "pt", In: "in", Q: "Q", Pc: "pc", Perc: "%", Em: "em", Ex: "ex"}

var toPx = [...]Fl{
	Px: 1, Cm: 96. / 2.54, Mm: 9.6 / 2.54, Pt: 96. / 72., In: 96., Q: 96. / 40. / 2.54, Pc: 96. / 6.,
	// other units depend on context
}

// 12pt
const defaultFontSize Fl = 96 * 12 / 72

// value is a value expressed in a unit.
// it may be relative, meaning that context is needed
// to obtain the actual value (see the `resolve` method)
type value struct {
	v Fl
	u unit
}

// look for an absolute unit, or nothing (considered as pixels)
// % is also supported
// it returns an empty value when 's' is empty
func parseValue(s string) (value, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return value{}, nil
	}

	resolvedUnit := Px
	for u, suffix := range units {
		if u == 0 {
			continue
		}
		if strings.HasSuffix(s, suffix) {
			s = strings.TrimSpace(strings.TrimSuffix(s, suffix))
			resolvedUnit = unit(u)
			break
		}
	}
	v, err := strconv.ParseFloat(s, 32)
	return value{u: resolvedUnit, v: Fl(v)}, err
}

// resolve convert `v` to pixels, resolving percentage and
// relative units
func (v value) resolve(fontSize, percentageReference Fl) Fl {
	switch v.u {
	case Px: // fast path for a common case
		return v.v
	case Perc:
		return v.v * percentageReference / 100
	case Em:
		return v.v * fontSize
	case Ex: // assume that 1em == 2ex
		return v.v * fontSize / 2
	default: // use the convertion table
		return v.v * toPx[v.u]
	}
}

// // convert the unite to pixels. Return true if it is a %
// func parseUnit(s string) (Fl, bool, error) {
// 	value, err := parseValue(s)
// 	return value.v * toPx[value.u], value.u == Perc, err
// }

// type percentageReference uint8

// const (
// 	widthPercentage percentageReference = iota
// 	heightPercentage
// 	diagPercentage
// )

// // resolveUnit converts a length with a unit into its value in 'px'
// // percentage are supported, and refer to the viewBox
// // `asPerc` is only applied when `s` contains a percentage.
// func (viewBox Bounds) resolveUnit(s string, asPerc percentageReference) (Fl, error) {
// 	value, isPercentage, err := parseUnit(s)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if isPercentage {
// 		w, h := viewBox.W, viewBox.H
// 		switch asPerc {
// 		case widthPercentage:
// 			return value / 100 * w, nil
// 		case heightPercentage:
// 			return value / 100 * h, nil
// 		case diagPercentage:
// 			normalizedDiag := math.Sqrt(w*w+h*h) / root2
// 			return value / 100 * normalizedDiag, nil
// 		}
// 	}
// 	return value, nil
// }

// // parseUnit converts a length with a unit into its value in 'px'
// // percentage are supported, and refer to the current ViewBox
// func (c *iconCursor) parseUnit(s string, asPerc percentageReference) (Fl, error) {
// 	return c.icon.ViewBox.resolveUnit(s, asPerc)
// }

// func readFraction(v string) (f Fl, err error) {
// 	v = strings.TrimSpace(v)
// 	d := 1.0
// 	if strings.HasSuffix(v, "%") {
// 		d = 100
// 		v = strings.TrimSuffix(v, "%")
// 	}
// 	f, err = parseBasicFloat(v)
// 	f /= d
// 	return
// }

// func readAppendFloat(numStr string, points []Fl) ([]Fl, error) {
// 	fmt.Println(numStr)
// 	last := 0
// 	isFirst := true
// 	for i, n := range numStr {
// 		if n == '.' {
// 			if isFirst {
// 				isFirst = false
// 				continue
// 			}
// 			f, err := parseBasicFloat(numStr[last:i])
// 			if err != nil {
// 				return nil, err
// 			}
// 			points = append(points, f)
// 			last = i
// 		}
// 	}
// 	f, err := parseBasicFloat(numStr[last:])
// 	if err != nil {
// 		return nil, err
// 	}
// 	points = append(points, f)
// 	return points, nil
// }

// parsePoints reads a set of floating point values from the SVG format number string.
// units are not supported.
// values are appended to points, which is returned
func parsePoints(dataPoints string, points []Fl) ([]Fl, error) {
	lastIndex := -1
	lr := ' '
	for i, r := range dataPoints {
		if !unicode.IsNumber(r) && r != '.' && !(r == '-' && lr == 'e') && r != 'e' {
			if lastIndex != -1 {
				value, err := strconv.ParseFloat(dataPoints[lastIndex:i], 32)
				if err != nil {
					return nil, err
				}
				points = append(points, Fl(value))
			}
			if r == '-' {
				lastIndex = i
			} else {
				lastIndex = -1
			}
		} else if lastIndex == -1 {
			lastIndex = i
		}
		lr = r
	}
	if lastIndex != -1 && lastIndex != len(dataPoints) {
		value, err := strconv.ParseFloat(dataPoints[lastIndex:], 32)
		if err != nil {
			return nil, err
		}
		points = append(points, Fl(value))
	}
	return points, nil
}

// parseFloatList reads a list of whitespace or comma-separated list of value,
// with units.
func parseFloatList(dataPoints string) (points []value, err error) {
	fields := strings.FieldsFunc(dataPoints, func(r rune) bool { return r == ' ' || r == ',' })
	points = make([]value, len(fields))
	for i, v := range fields {
		val, err := parseValue(v)
		if err != nil {
			return nil, err
		}
		points[i] = val
	}
	return points, nil
}

// if the URL is invalid, the empty string is returned
func parseURLFragment(url_ string) string {
	u, err := parseURL(url_)
	if err != nil {
		return ""
	}
	return u.Fragment
}

// parse a URL, possibly in a "url(…)" string.
func parseURL(url_ string) (*url.URL, error) {
	if strings.HasPrefix(url_, "url(") && strings.HasSuffix(url_, ")") {
		url_ = url_[4 : len(url_)-1]
		if len(url_) >= 2 {
			if (url_[0] == '"' && url_[len(url_)-1] == '"') || (url_[0] == '\'' && url_[len(url_)-1] == '\'') {
				url_ = url_[1 : len(url_)-1]
			}
		}
	}
	return url.Parse(url_)
}

func parseViewbox(attr string) (Rectangle, error) {
	points, err := parsePoints(attr, nil)
	if err != nil {
		return Rectangle{}, err
	}
	if len(points) != 4 {
		return Rectangle{}, fmt.Errorf("expected 4 numbers for viewbox, got %s", attr)
	}
	return Rectangle{points[0], points[1], points[2], points[3]}, nil
}
