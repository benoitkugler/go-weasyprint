package css

import (
	"math"
	"path"

	"github.com/benoitkugler/go-weasyprint/utils"
)

// Expand shorthands and validate property values.
// See http://www.w3.org/TR/CSS21/propidx.html and various CSS3 modules.

// :copyright: Copyright 2011-2014 Simon Sapin and contributors, see AUTHORS.
// :license: BSD, see LICENSE for details.

// get the sets of keys
var (
	LENGTHUNITS = Set{"ex": true, "em": true, "ch": true, "rem": true}

	// keyword -> (open, insert)
	CONTENTQUOTEKEYWORDS = map[string]quote{
		"open-quote":     {open: true, insert: true},
		"close-quote":    {open: false, insert: true},
		"no-open-quote":  {open: true, insert: false},
		"no-close-quote": {open: false, insert: false},
	}

	ZEROPERCENT    = Dimension{Value: 0, Unit: "%"}
	FIFTYPERCENT   = Dimension{Value: 50, Unit: "%"}
	HUNDREDPERCENT = Dimension{Value: 100, Unit: "%"}

	BACKGROUNDPOSITIONPERCENTAGES = map[string]Dimension{
		"top":    ZEROPERCENT,
		"left":   ZEROPERCENT,
		"center": FIFTYPERCENT,
		"bottom": HUNDREDPERCENT,
		"right":  HUNDREDPERCENT,
	}

	// yes/no validators for non-shorthand properties
	// Maps property names to functions taking a property name && a value list,
	// returning a value || None for invalid.
	// Also transform values: keyword && URLs are returned as strings.
	// For properties that take a single value, that value is returned by itself
	// instead of a list.
	VALIDATORS = map[string]validator{}

	// http://dev.w3.org/csswg/css3-values/#angles
	// 1<unit> is this many radians.
	ANGLETORADIANS = map[string]float64{
		"rad":  1,
		"turn": 2 * math.Pi,
		"deg":  math.Pi / 180,
		"grad": math.Pi / 200,
	}

	// http://dev.w3.org/csswg/css-values/#resolution
	RESOLUTIONTODPPX = map[string]float64{
		"dppx": 1,
		"dpi":  1 / LengthsToPixels["in"],
		"dpcm": 1 / LengthsToPixels["cm"],
	}
)

type validator func(string)

type quote struct {
	open, insert bool
}

type Token struct {
	Dimension
	Type, LowerValue string
}

func init() {
	for k := range LengthsToPixels {
		LENGTHUNITS[k] = true
	}
}

// If ``token`` is a keyword, return its name.
//     Otherwise return ``None``.
func getKeyword(token Token) string {
	if token.Type == "ident" {
		return token.LowerValue
	}
	return ""
}

// If ``tokens`` is a 1-element list of keywords, return its name.
//     Otherwise return ``None``.
func getSingleKeyword(tokens []Token) string {
	if len(tokens) == 1 {
		token := tokens[0]
		return getKeyword(token)
	}
	return ""
}

// negative  = true, percentage = false
func getLength(token Token, negative, percentage bool) Dimension {
	if percentage && token.Type == "percentage" {
		if negative || token.Value >= 0 {
			return Dimension{Value: token.Value, Unit: "%"}
		}
	}
	if token.Type == "dimension" && LENGTHUNITS[token.Unit] {
		if negative || token.Value >= 0 {
			return token.Dimension
		}
	}
	if token.Type == "number" && token.Value == 0 {
		return Dimension{Unit: "None"}
	}
	return Dimension{}
}

// Return the value in radians of an <angle> token, or None.
func getAngle(token Token) (float64, bool) {
	if token.Type == "dimension" {
		factor, in := ANGLETORADIANS[token.Unit]
		if in {
			return token.Value * factor, true
		}
	}
	return 0, false
}

// Return the value := range dppx of a <resolution> token, || None.
func getResolution(token Token) (float64, bool) {
	if token.Type == "dimension" {
		factor, in := RESOLUTIONTODPPX[token.Unit]
		if in {
			return token.Value * factor, true
		}
	}
	return 0, false
}

func safeUrljoin(baseUrl, url string) string {
	if path.IsAbs(url) {
		return utils.IriToUri(url)
	} else if baseUrl != "" {
		return utils.IriToUri(path.Join(baseUrl, url))
	} else {
		panic("Relative URI reference without a base URI: " + url)
	}
}
