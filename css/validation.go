package css

import (
	"errors"
	"fmt"
	"math"
	"path"
	"strings"

	"github.com/benoitkugler/go-weasyprint/structure/counters"
	"github.com/benoitkugler/go-weasyprint/utils"
)

// Expand shorthands and validate property values.
// See http://www.w3.org/TR/CSS21/propidx.html and various CSS3 modules.

// :copyright: Copyright 2011-2014 Simon Sapin and contributors, see AUTHORS.
// :license: BSD, see LICENSE for details.

var (
	LENGTHUNITS = Set{"ex": true, "em": true, "ch": true, "rem": true}

	// keyword -> (open, insert)
	ContentQuoteKeywords = map[string]quote{
		"open-quote":     {open: true, insert: true},
		"close-quote":    {open: false, insert: true},
		"no-open-quote":  {open: true, insert: false},
		"no-close-quote": {open: false, insert: false},
	}

	ZEROPERCENT    = Dimension{Value: 0, Unit: "%"}
	fiftyPercent   = Dimension{Value: 50, Unit: "%"}
	HUNDREDPERCENT = Dimension{Value: 100, Unit: "%"}

	backgroundPositionsPercentages = map[string]Dimension{
		"top":    ZEROPERCENT,
		"left":   ZEROPERCENT,
		"center": fiftyPercent,
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

	couplesLigatures = [][]string{
		{"common-ligatures", "no-common-ligatures"},
		{"historical-ligatures", "no-historical-ligatures"},
		{"discretionary-ligatures", "no-discretionary-ligatures"},
		{"contextual", "no-contextual"},
	}
	couplesNumeric = [][]string{
		{"lining-nums", "oldstyle-nums"},
		{"proportional-nums", "tabular-nums"},
		{"diagonal-fractions", "stacked-fractions"},
		{"ordinal"},
		{"slashed-zero"},
	}
	allLigaturesValues = Set{}
	allNumericValues   = Set{}
)

func init() {
	for k := range LengthsToPixels {
		LENGTHUNITS[k] = true
	}

	for _, couple := range couplesLigatures {
		for _, cc := range couple {
			allLigaturesValues[cc] = true
		}
	}
	for _, couple := range couplesNumeric {
		for _, cc := range couple {
			allNumericValues[cc] = true
		}
	}
}

func parseColor(color Token) Color {
	// return Color{}
}

type validator func(string)

type quote struct {
	open, insert bool
}

type MaybeInt struct {
	Valid bool
	Int   int
}

type Token struct {
	Dimension
	String                      string
	Type, LowerValue, LowerName string
	IntValue                    MaybeInt
	Arguments                   []Token
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

func safeUrljoin(baseUrl, url string) (string, error) {
	if path.IsAbs(url) {
		return utils.IriToUri(url), nil
	} else if baseUrl != "" {
		return utils.IriToUri(path.Join(baseUrl, url)), nil
	} else {
		return "", errors.New("Relative URI reference without a base URI: " + url)
	}
}

//@validator()
//@commaSeparatedList
//@singleKeyword
// ``background-attachment`` property validation.
func backgroundAttachment(keyword string) bool {
	switch keyword {
	case "scroll", "fixed", "local":
		return true
	default:
		return false
	}
}

//@validator("background-color")
//@validator("border-top-color")
//@validator("border-right-color")
//@validator("border-bottom-color")
//@validator("border-left-color")
//@validator("column-rule-color")
//@singleToken
func otherColors(token Token) Color {
	return parseColor(token)
}

//@validator()
//@singleToken
func outlineColor(token Token) Color {
	if getKeyword(token) == "invert" {
		return Color{String: "currentColor"}
	} else {
		return parseColor(token)
	}
}

//@validator()
//@singleKeyword
func borderCollapse(keyword string) bool {
	switch keyword {
	case "separate", "collapse":
		return true
	default:
		return false
	}
}

//@validator()
//@singleKeyword
// ``empty-cells`` property validation.
func emptyCells(keyword string) bool {
	switch keyword {
	case "show", "hide":
		return true
	default:
		return false
	}
}

//@validator("color")
//@singleToken
// ``*-color`` && ``color`` properties validation.
func color(token Token) Color {
	result := parseColor(token)
	if result.String == "currentColor" {
		return Color{String: "inherit"}
	} else {
		return result
	}
}

//@validator("background-image", wantsBaseUrl=true)
//@commaSeparatedList
//@singleToken
// func backgroundImage(token Token, baseUrl string) {
//     if token.Type != "function" {
//         return imageUrl([]Token{token}, baseUrl)
// 	}
// 	arguments := splitOnComma(removeWhitespace(token.Arguments))
// 	name := token.LowerName
// 	switch name {
// 	case "linear-gradient", "repeating-linear-gradient":
// 		 direction, colorStops := parseLinearGradientParameters(arguments)
//         if colorStops {
//             return "linear-gradient", LinearGradient(
//                 [parseColorStop(stop) for stop := range colorStops],
//                 direction, "repeating" := range name)
// 		}
// 	case "radial-gradient", "repeating-radial-gradient":
// 	        result = parseRadialGradientParameters(arguments)
//         if result is not None {
//             shape, size, position, colorStops = result
//         } else {
//             shape = "ellipse"
//             size = "keyword", "farthest-corner"
//             position = "left", fiftyPercent, "top", fiftyPercent
//             colorStops = arguments
//         } if colorStops {
//             return "radial-gradient", RadialGradient(
//                 [parseColorStop(stop) for stop := range colorStops],
//                 shape, size, position, "repeating" := range name)
// 		}
// 	}

// }

var directionKeywords = map[[3]string]directionType{
	// ("angle", radians)  0 upwards, then clockwise
	{"to", "top", ""}:    {angle: 0},
	{"to", "right", ""}:  {angle: math.Pi / 2},
	{"to", "bottom", ""}: {angle: math.Pi},
	{"to", "left", ""}:   {angle: math.Pi * 3 / 2},
	// ("corner", keyword)
	{"to", "top", "left"}:     {corner: "topLeft"},
	{"to", "left", "top"}:     {corner: "topLeft"},
	{"to", "top", "right"}:    {corner: "topRight"},
	{"to", "right", "top"}:    {corner: "topRight"},
	{"to", "bottom", "left"}:  {corner: "bottomLeft"},
	{"to", "left", "bottom"}:  {corner: "bottomLeft"},
	{"to", "bottom", "right"}: {corner: "bottomRight"},
	{"to", "right", "bottom"}: {corner: "bottomRight"},
}

type directionType struct {
	corner string
	angle  float64
}

func parseLinearGradientParameters(arguments [][]Token) (directionType, [][]Token) {
	firstArg := arguments[0]
	if len(firstArg) == 1 {
		angle, isNotNone := getAngle(firstArg[0])
		if isNotNone {
			return directionType{angle: angle}, arguments[1:]
		}
	} else {
		var mapped [3]string
		for index, token := range firstArg {
			if index < 3 {
				mapped[index] = getKeyword(token)
			}
		}
		result, isNotNone := directionKeywords[mapped]
		if isNotNone {
			return result, arguments[1:]
		}
	}
	return directionType{angle: math.Pi}, arguments // Default direction is "to bottom"
}

func reverse(a []Token) []Token {
	n := len(a)
	out := make([]Token, n)
	for i := range a {
		out[n-1-i] = a[i]
	}
	return out
}

type gradientSize struct {
	keyword  string
	explicit [2]Dimension
}

func (s gradientSize) isNone() bool {
	return s == gradientSize{}
}

type gradientPosition struct {
	keyword1 string
	length1  Dimension
	keyword2 string
	length2  Dimension
}

func (g gradientPosition) IsNone() bool {
	return g == gradientPosition{}
}

type radialGradientParameters struct {
	shape      string
	size       gradientSize
	position   gradientPosition
	colorStops [][]Token
}

func (r radialGradientParameters) IsNone() bool {
	return r.shape == "" && r.size.isNone() && r.position.IsNone() && r.colorStops == nil
}

func parseRadialGradientParameters(arguments [][]Token) radialGradientParameters {
	var shape, sizeShape string
	var position gradientPosition
	var size gradientSize
	stack := reverse(arguments[0])
	for len(stack) > 0 {
		token := stack[len(stack)-1]
		keyword := getKeyword(token)
		switch keyword {
		case "at":
			position = valideBackgroundPosition(reverse(stack))
			if position.IsNone() {
				return radialGradientParameters{}
			}
			break
		case "circle", "ellipse":
			if shape == "" {
				shape = keyword
			}
		case "closest-corner", "farthest-corner", "closest-side", "farthest-side":
			if size.isNone() {
				size = gradientSize{keyword: keyword}
			}
		default:
			if len(stack) > 0 && size.isNone() {
				length1 := getLength(token, true, true)
				length2 := getLength(stack[len(stack)-1], true, true)
				if !length1.IsNone() && !length2.IsNone() {
					size = gradientSize{explicit: [2]Dimension{length1, length2}}
					sizeShape = "ellipse"
					stack = stack[:len(stack)-2]
				}
			}
			if size.isNone() {
				length1 := getLength(token, true, false)
				if !length1.IsNone() {
					size = gradientSize{explicit: [2]Dimension{length1, length1}}
					sizeShape = "circle"
				}
			}
			if size.isNone() {
				return radialGradientParameters{}
			}
		}
	}
	if shape == "circle" && sizeShape == "ellipse" {
		return radialGradientParameters{}
	}
	out := radialGradientParameters{
		shape:      shape,
		size:       size,
		position:   position,
		colorStops: arguments[1:],
	}
	if shape == "" {
		if sizeShape != "" {
			out.shape = sizeShape
		} else {
			out.shape = "ellipse"
		}
	}
	if size.isNone() {
		out.size = gradientSize{keyword: "farthest-corner"}
	}
	if position.IsNone() {
		out.position = gradientPosition{
			keyword1: "left",
			length1:  fiftyPercent,
			keyword2: "top",
			length2:  fiftyPercent,
		}
	}
	return out
}

func parseColorStop(tokens []Token) (Color, Dimension, bool) {
	switch len(tokens) {
	case 1:
		color := parseColor(tokens[0])
		if !color.IsNone() {
			return color, Dimension{}, true
		}
	case 2:
		color := parseColor(tokens[0])
		position := getLength(tokens[1], true, true)
		if !color.IsNone() && !position.IsNone() {
			return color, position, true
		}
	default:
		panic("Invalid or unsupported values for a known CSS property.")
	}
	return Color{}, Dimension{}, false
}

//@validator("list-style-image", wantsBaseUrl=true)
//@singleToken
// ``*-image`` properties validation.
func imageUrl(token Token, baseUrl string) (string, string, error) {
	if getKeyword(token) == "none" {
		return "none", "", nil
	}
	if token.Type == "url" {
		s, err := safeUrljoin(baseUrl, token.String)
		return "url", s, err
	}
	return "", "", nil
}

var centerKeywordFakeToken = Token{
	Type: "ident", LowerValue: "center",
}

//@validator(unstable=true)
func transformOrigin(tokens []Token) (posX, posY Dimension, isNotNone bool) {
	// TODO: parse (and ignore) a third value for Z.
	return simple2dPosition(tokens)
}

//@validator()
//@commaSeparatedList
// ``background-position`` property validation.
//     See http://dev.w3.org/csswg/css3-background/#the-background-position
//
func valideBackgroundPosition(tokens []Token) gradientPosition {
	posX, posY, isNotNone := simple2dPosition(tokens)
	if isNotNone {
		return gradientPosition{
			keyword1: "left",
			length1:  posX,
			keyword2: "top",
			length2:  posY,
		}
	}

	if len(tokens) == 4 {
		keyword1 := getKeyword(tokens[0])
		keyword2 := getKeyword(tokens[2])
		length1 := getLength(tokens[1], true, true)
		length2 := getLength(tokens[3], true, true)
		if !length1.IsNone() && !length2.IsNone() {
			if (keyword1 == "left" || keyword1 == "right") && (keyword2 == "top" || keyword2 == "bottom") {
				return gradientPosition{keyword1: keyword1,
					length1:  length1,
					keyword2: keyword2,
					length2:  length2,
				}
			}
			if (keyword2 == "left" || keyword2 == "right") && (keyword1 == "top" || keyword1 == "bottom") {
				return gradientPosition{keyword1: keyword2,
					length1:  length2,
					keyword2: keyword1,
					length2:  length1,
				}
			}
		}
	}

	if len(tokens) == 3 {
		length := getLength(tokens[2], true, true)
		var keyword, otherKeyword string
		if !length.IsNone() {
			keyword = getKeyword(tokens[1])
			otherKeyword = getKeyword(tokens[0])
		} else {
			length = getLength(tokens[1], true, true)
			otherKeyword = getKeyword(tokens[2])
			keyword = getKeyword(tokens[0])
		}

		if !length.IsNone() {
			switch otherKeyword {
			case "center":
				switch keyword {
				case "top", "bottom":
					return gradientPosition{keyword1: "left", length1: fiftyPercent, keyword2: keyword, length2: length}
				case "left", "right":
					return gradientPosition{keyword1: keyword, length1: length, keyword2: "top", length2: fiftyPercent}
				}
			case "top", "bottom":
				if keyword == "left" || keyword == "right" {
					return gradientPosition{keyword1: keyword, length1: length, keyword2: otherKeyword, length2: ZEROPERCENT}
				}
			case "left", "right":
				if keyword == "top" || keyword == "bottom" {
					return gradientPosition{keyword1: otherKeyword, length1: ZEROPERCENT, keyword2: keyword, length2: length}
				}
			}
		}
	}
	return gradientPosition{}
}

// Common syntax of background-position && transform-origin.
func simple2dPosition(tokens []Token) (posX, posY Dimension, isNotNone bool) {
	if len(tokens) == 1 {
		tokens = []Token{tokens[0], centerKeywordFakeToken}
	} else if len(tokens) != 2 {
		return
	}

	token1, token2 := tokens[0], tokens[1]
	length1 := getLength(token1, true, true)
	length2 := getLength(token2, true, true)
	if !length1.IsNone() && !length2.IsNone() {
		return length1, length2, true
	}
	keyword1, keyword2 := getKeyword(token1), getKeyword(token2)
	if !length1.IsNone() && (keyword2 == "top" || keyword2 == "center" || keyword2 == "bottom") {
		return length1, backgroundPositionsPercentages[keyword2], true
	} else if !length2.IsNone() && (keyword1 == "left" || keyword1 == "center" || keyword1 == "right") {
		return backgroundPositionsPercentages[keyword1], length2, true
	} else if (keyword1 == "left" || keyword1 == "center" || keyword1 == "right") &&
		(keyword2 == "top" || keyword2 == "center" || keyword2 == "bottom") {
		return backgroundPositionsPercentages[keyword1], backgroundPositionsPercentages[keyword2], true
	} else if (keyword1 == "top" || keyword1 == "center" || keyword1 == "bottom") &&
		(keyword2 == "left" || keyword2 == "center" || keyword2 == "right") {
		// Swap tokens. They need to be in (horizontal, vertical) order.
		return backgroundPositionsPercentages[keyword2], backgroundPositionsPercentages[keyword1], true
	}
	return
}

//@validator()
//@commaSeparatedList
// ``background-repeat`` property validation.
func backgroundRepeat(tokens []Token) [2]string {
	keywords := make([]string, len(tokens))
	for index, token := range tokens {
		keywords[index] = getKeyword(token)
	}

	switch len(keywords) {
	case 1:
		switch keywords[0] {
		case "repeat-x":
			return [2]string{"repeat", "no-repeat"}
		case "repeat-y":
			return [2]string{"no-repeat", "repeat"}
		case "no-repeat", "repeat", "space", "round":
			return [2]string{keywords[0], keywords[0]}
		}
	case 2:
		for _, k := range keywords {
			if !(k == "no-repeat" || k == "repeat" || k == "space" || k == "round") {
				return [2]string{}
			}
		}
		// OK
		return [2]string{keywords[0], keywords[1]}
	}
	return [2]string{}
}

//@validator()
//@commaSeparatedList
// Validation for ``background-size``.
func backgroundSize(tokens []Token) []Value {
	switch len(tokens) {
	case 1:
		token := tokens[0]
		keyword := getKeyword(token)
		switch keyword {
		case "contain", "cover":
			return []Value{Value{String: keyword}}
		case "auto":
			return []Value{Value{String: "auto"}, Value{String: "auto"}}
		}
		length := getLength(token, false, true)
		if !length.IsNone() {
			return []Value{Value{Dimension: length}, Value{String: "auto"}}
		}
	case 2:
		var values []Value
		for _, token := range tokens {
			length := getLength(token, false, true)
			if !length.IsNone() {
				values = append(values, Value{Dimension: length})
			} else if getKeyword(token) == "auto" {
				values = append(values, Value{String: "auto"})
			}
		}
		if len(values) == 2 {
			return values
		}
	}
	return nil
}

//@validator("background-clip")
//@validator("background-origin")
//@commaSeparatedList
//@singleKeyword
// Validation for the ``<box>`` type used in ``background-clip``
//     and ``background-origin``.
func box(keyword string) bool {
	return keyword == "border-box" || keyword == "padding-box" || keyword == "content-box"
}

func borderDims(tokens []Token, negative bool) [2]Dimension {
	lengths := make([]Dimension, len(tokens))
	allLengths := true
	for index, token := range tokens {
		lengths[index] = getLength(token, negative, false)
		allLengths = allLengths && !lengths[index].IsNone()
	}
	if allLengths {
		if len(lengths) == 1 {
			return [2]Dimension{lengths[0], lengths[0]}
		} else if len(lengths) == 2 {
			return [2]Dimension{lengths[0], lengths[1]}
		}
	}
	return [2]Dimension{}
}

//@validator()
// Validator for the `border-spacing` property.
func borderSpacing(tokens []Token) [2]Dimension {
	return borderDims(tokens, true)
}

//@validator("border-top-right-radius")
//@validator("border-bottom-right-radius")
//@validator("border-bottom-left-radius")
//@validator("border-top-left-radius")
// Validator for the `border-*-radius` properties.
func borderCornerRadius(tokens []Token) [2]Dimension {
	return borderDims(tokens, false)
}

//@validator("border-top-style")
//@validator("border-right-style")
//@validator("border-left-style")
//@validator("border-bottom-style")
//@validator("column-rule-style")
//@singleKeyword
// ``border-*-style`` properties validation.
func borderStyle(keyword string) bool {
	switch keyword {
	case "none", "hidden", "dotted", "dashed", "double",
		"inset", "outset", "groove", "ridge", "solid":
		return true
	default:
		return false
	}
}

//@validator("break-before")
//@validator("break-after")
//@singleKeyword
// ``break-before`` && ``break-after`` properties validation.
func breakBeforeAfter(keyword string) bool {
	// "always" is defined as an alias to "page" := range multi-column
	// https://www.w3.org/TR/css3-multicol/#column-breaks
	switch keyword {
	case "auto", "avoid", "avoid-page", "page", "left", "right",
		"recto", "verso", "avoid-column", "column", "always":
		return true
	default:
		return false
	}
}

//@validator()
//@singleKeyword
// ``break-inside`` property validation.
func breakInside(keyword string) bool {
	switch keyword {
	case "auto", "avoid", "avoid-page", "avoid-column":
		return true
	default:
		return false
	}
}

//@validator(unstable=true)
//@singleToken
// ``page`` property validation.
func page(token Token) string {
	if token.Type == "ident" {
		if token.LowerValue == "auto" {
			return "auto"
		}
		return token.String
	}
	return ""
}

//@validator("bleed-left")
//@validator("bleed-right")
//@validator("bleed-top")
//@validator("bleed-bottom")
//@singleToken
// ``bleed`` property validation.
func bleed(token Token) Bleed {
	keyword := getKeyword(token)
	if keyword == "auto" {
		return Bleed{String: "auto"}
	} else {
		return Bleed{Dimension: getLength(token, true, false)}
	}
}

//@validator()
// ``marks`` property validation.
func marks(tokens []Token) Marks {
	if len(tokens) == 2 {
		keywords := [2]string{getKeyword(tokens[0]), getKeyword(tokens[1])}
		if keywords == [2]string{"crop", "cross"} || keywords == [2]string{"cross", "crop"} {
			return Marks{Crop: true, Cross: true}
		}
	} else if len(tokens) == 1 {
		keyword := getKeyword(tokens[0])
		switch keyword {
		case "crop":
			return Marks{Crop: true}
		case "cross":
			return Marks{Cross: true}
		case "none":
			return Marks{}
		}
	}
	return Marks{}
}

//@validator("outline-style")
//@singleKeyword
// ``outline-style`` properties validation.
func outlineStyle(keyword string) bool {
	switch keyword {
	case "none", "dotted", "dashed", "double", "inset",
		"outset", "groove", "ridge", "solid":
		return true
	default:
		return false
	}
}

// //@validator("border-top-width")
// //@validator("border-right-width")
// //@validator("border-left-width")
// //@validator("border-bottom-width")
// //@validator("column-rule-width")
// //@validator("outline-width")
// //@singleToken
// Border, column rule && outline widths properties validation.
func borderWidth(token Token) Value {
	length := getLength(token, false, false)
	if !length.IsNone() {
		return Value{Dimension: length}
	}
	keyword := getKeyword(token)
	if keyword == "thin" || keyword == "medium" || keyword == "thick" {
		return Value{String: keyword}
	}
	return Value{}
}

// //@validator()
// //@singleToken
// ``column-width`` property validation.
func columnWidth(token Token) Value {
	length := getLength(token, false, false)
	if !length.IsNone() {
		return Value{Dimension: length}
	}
	keyword := getKeyword(token)
	if keyword == "auto" {
		return Value{String: keyword}
	}
	return Value{}
}

// //@validator()
// //@singleKeyword
// ``column-span`` property validation.
func columnSpan(keyword string) bool {
	// TODO: uncomment this when it is supported
	// return keyword := range ("all", "none")
	return false
}

// //@validator()
// //@singleKeyword
// Validation for the ``box-sizing`` property from css3-ui
func boxSizing(keyword string) bool {
	return keyword == "padding-box" || keyword == "border-box" || keyword == "content-box"
}

// //@validator()
// //@singleKeyword
// ``caption-side`` properties validation.
func captionSide(keyword string) bool {
	return keyword == "top" || keyword == "bottom"
}

// //@validator()
// //@singleKeyword
// ``clear`` property validation.
func clear(keyword string) bool {
	return keyword == "left" || keyword == "right" || keyword == "both" || keyword == "none"
}

// //@validator()
// //@singleToken
// Validation for the ``clip`` property.
func clip(token Token) []Value {
	name, args := parseFunction(token)
	if name != "" {
		if name == "rect" && len(args) == 4 {
			var values []Value
			for _, arg := range args {
				if getKeyword(arg) == "auto" {
					values = append(values, Value{String: "auto"})
				} else {
					length := getLength(arg, true, false)
					if !length.IsNone() {
						values = append(values, Value{Dimension: length})
					}
				}
			}
			if len(values) == 4 {
				return values
			}
		}
	}
	if getKeyword(token) == "auto" {
		return []Value{}
	}
	return nil
}

// //@validator(wantsBaseUrl=true)
// ``content`` property validation.
func content(tokens []Token, baseUrl string) (Content, error) {
	keyword := getSingleKeyword(tokens)
	if keyword == "normal" || keyword == "none" {
		return Content{String: keyword}, nil
	}
	out := make([]ContentProperty, len(tokens))
	for index, v := range tokens {
		contentProperty, err := validateContentToken(baseUrl, v)
		if err != nil {
			return Content{}, err
		}
		if contentProperty.IsNil() {
			return Content{}, nil
		}
		out[index] = contentProperty
	}
	return Content{List: out}, nil
}

func _equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Validation for a single token for the ``content`` property.
// Return (type, content) || zero for invalid tokens.
func validateContentToken(baseUrl string, token Token) (ContentProperty, error) {
	quoteType, isNotNone := ContentQuoteKeywords[getKeyword(token)]
	if isNotNone {
		return ContentProperty{Type: ContentQUOTE, Quote: quoteType}, nil
	}

	switch token.Type {
	case "string":
		return ContentProperty{Type: ContentSTRING, String: token.String}, nil
	case "url":
		url, err := safeUrljoin(baseUrl, token.String)
		if err != nil {
			return ContentProperty{}, err
		}
		return ContentProperty{Type: ContentURI, String: url}, nil
	}

	name, args := parseFunction(token)
	if name != "" {
		prototypeArgs := make([]string, len(args))
		for index, arg := range args {
			prototypeArgs[index] = arg.Type
		}

		// args = [getattr(a, "value", a) for a := range args]

		ty := ContentCounter
		if name == "counters" {
			ty = ContentCounters
		}
		var argStrings []string
		for _, arg := range args {
			argStrings = append(argStrings, arg.String)
		}
		if name == "attr" && _equal(prototypeArgs, []string{"ident"}) {
			return ContentProperty{Type: ContentAttr, String: args[0].String}, nil
		} else if (name == "counter" && _equal(prototypeArgs, []string{"ident"})) ||
			(name == "counters" && _equal(prototypeArgs, []string{"ident", "string"})) {

			argStrings = append(argStrings, "decimal")
			return ContentProperty{Type: ty, Strings: argStrings}, nil
		} else if (name == "counter" && _equal(prototypeArgs, []string{"ident", "ident"})) ||
			(name == "counters" && _equal(prototypeArgs, []string{"ident", "string", "ident"})) {
			style := args[len(args)-1].String
			_, isIn := counters.STYLES[style]
			if style == "none" || style == "decimal" || isIn {
				return ContentProperty{Type: ty, Strings: argStrings}, nil
			}
		} else if (name == "string" && _equal(prototypeArgs, []string{"ident"})) ||
			(name == "string" && _equal(prototypeArgs, []string{"ident", "ident"})) {
			if len(args) > 1 {
				argStrings[1] = strings.ToLower(argStrings[1])
				if argStrings[1] != "first" && argStrings[1] != "start" && argStrings[1] != "last" && argStrings[1] != "first-except" {
					return ContentProperty{}, fmt.Errorf("Invalid or unsupported CSS value : %s", argStrings[1])
				}
			}
			return ContentProperty{Type: ContentString, Strings: argStrings}, nil

		}
	}
	return ContentProperty{}, nil
}

// Return ``(name, args)`` if the given token is a function
//     with comma-separated arguments
func parseFunction(functionToken Token) (string, []Token) {
	if functionToken.Type == "function" {
		content := removeWhitespace(functionToken.Arguments)
		if len(content) == 0 || len(content)%2 == 1 {
			for i := 1; i < len(content); i += 2 { // token in content[1::2]
				token := content[i]
				if token.Type != "literal" || token.String != "," {
					return "", nil
				}
			}
			var args []Token
			for i := 0; i < len(content); i += 2 {
				args = append(args, content[i])
			}
			return functionToken.LowerName, args
		}
	}
	return "", nil
}

// //@validator()
// ``counter-increment`` property validation.
func counterIncrement(tokens []Token) (CounterResets, error) {
	return counter(tokens, 1)
}

// //@validator()
// ``counter-reset`` property validation.
func counterReset(tokens []Token) (CounterResets, error) {
	return counter(tokens, 0)
}

// ``counter-increment`` && ``counter-reset`` properties validation.
func counter(tokens []Token, defaultInteger int) (CounterResets, error) {
	if getSingleKeyword(tokens) == "none" {
		return nil, nil
	}
	if len(tokens) == 0 {
		return nil, errors.New("got an empty token list")
	}
	var (
		results    CounterResets
		i, integer int
		token      Token
	)
	for i < len(tokens) {
		token = tokens[i]

		if token.Type != "ident" {
			return nil, nil // expected a keyword here
		}
		counterName := token.String
		if counterName == "none" || counterName == "initial" || counterName == "inherit" {
			return nil, fmt.Errorf("Invalid counter name: %s", counterName)
		}
		i += 1
		token = tokens[i]
		if !token.IsNone() && token.Type == "number" && token.IntValue.Valid {
			// Found an integer. Use it and get the next token
			integer = token.IntValue.Int
			i += 1
			token = tokens[i]
		} else {
			// Not an integer. Might be the next counter name.
			// Keep `token` for the next loop iteration.
			integer = defaultInteger
		}
		results = append(results, NameInt{Name: counterName, Value: integer})
	}
	return results, nil
}

// //@validator("top")
// //@validator("right")
// //@validator("left")
// //@validator("bottom")
// //@validator("margin-top")
// //@validator("margin-right")
// //@validator("margin-bottom")
// //@validator("margin-left")
// //@singleToken
// ``margin-*`` properties validation.
func lenghtPrecentageOrAuto(token Token) Value {
	length := getLength(token, true, true)
	if !length.IsNone() {
		return Value{Dimension: length}
	}
	if getKeyword(token) == "auto" {
		return Value{String: "auto"}
	}
	return Value{}
}

// //@validator("height")
// //@validator("width")
// //@singleToken
// Validation for the ``width`` && ``height`` properties.
func widthHeight(token Token) Value {
	length := getLength(token, false, true)
	if !length.IsNone() {
		return Value{Dimension: length}
	}
	if getKeyword(token) == "auto" {
		return Value{String: "auto"}
	}
	return Value{}
}

// //@validator()
// //@singleToken
// Validation for the ``column-gap`` property.
func columnGap(token Token) Value {
	length := getLength(token, false, false)
	if !length.IsNone() {
		return Value{Dimension: length}
	}
	if getKeyword(token) == "normal" {
		return Value{String: "normal"}
	}
	return Value{}
}

//@validator()
//@singleKeyword
// ``column-fill`` property validation.
func columnFill(keyword string) bool {
	return keyword == "auto" || keyword == "balance"
}

//@validator()
//@singleKeyword
// ``direction`` property validation.
func direction(keyword string) bool {
	return keyword == "ltr" || keyword == "rtl"
}

//@validator()
//@singleKeyword
// ``display`` property validation.
func display(keyword string) bool {
	return keyword == "inline" || keyword == "block" || keyword == "inline-block" || keyword == "list-item" || keyword == "none" || keyword ==
		"table" || keyword == "inline-table" || keyword == "table-caption" || keyword ==
		"table-row-group" || keyword == "table-header-group" || keyword == "table-footer-group" || keyword ==
		"table-row" || keyword == "table-column-group" || keyword == "table-column" || keyword == "table-cell"
}

//@validator("float")
//@singleKeyword
// ``float`` property validation.
func float(keyword string) bool {
	return keyword == "left" || keyword == "right" || keyword == "none"
}

// //@validator()
// //@commaSeparatedList
// ``font-family`` property validation.
func fontFamily(tokens []Token) string {
	if len(tokens) == 1 && tokens[0].Type == "string" {
		return tokens[0].String
	} else if len(tokens) > 0 {
		var values []string
		for _, token := range tokens {
			if token.Type != "ident" {
				return ""
			}
			values = append(values, token.String)
		}
		return strings.Join(values, " ")
	}
	return ""
}

// //@validator()
// //@singleKeyword
func fontKerning(keyword string) bool {
	return keyword == "auto" || keyword == "normal" || keyword == "none"
}

// //@validator()
// //@singleToken
func fontLanguageOverride(token Token) string {
	keyword := getKeyword(token)
	if keyword == "normal" {
		return keyword
	} else if token.Type == "string" {
		return token.String
	}
	return ""
}

func parseFontVariant(tokens []Token, all Set, couples [][]string) FontVariant {
	var values []string
	isInValues := func(s string, vs []string) bool {
		for _, v := range vs {
			if s == v {
				return true
			}
		}
		return false
	}
	for _, token := range tokens {
		if token.Type != "ident" {
			return FontVariant{}
		}
		if all[token.String] {
			var concurrentValues []string
			for _, couple := range couples {
				if isInValues(token.String, couple) {
					concurrentValues = couple
					break
				}
			}
			for _, value := range concurrentValues {
				if isInValues(value, values) {
					return FontVariant{}
				}
			}
			values = append(values, token.String)
		} else {
			return FontVariant{}
		}
	}
	if len(values) > 0 {
		return FontVariant{Values: values}
	}
	return FontVariant{}
}

// //@validator()
func fontVariantLigatures(tokens []Token) FontVariant {
	if len(tokens) == 1 {
		keyword := getKeyword(tokens[0])
		if keyword == "normal" || keyword == "none" {
			return FontVariant{String: keyword}
		}
	}
	return parseFontVariant(tokens, allLigaturesValues, couplesLigatures)
}

// //@validator()
// //@singleKeyword
func fontVariantPosition(keyword string) bool {
	return keyword == "normal" || keyword == "sub" || keyword == "super"
}

// //@validator()
// //@singleKeyword
func fontVariantCaps(keyword string) bool {
	return keyword == "normal" || keyword == "small-caps" || keyword == "all-small-caps" || keyword == "petite-caps" || keyword ==
		"all-petite-caps" || keyword == "unicase" || keyword == "titling-caps"
}

//@validator()
func fontVariantNumeric(tokens []Token) FontVariant {
	if len(tokens) == 1 {
		keyword := getKeyword(tokens[0])
		if keyword == "normal" {
			return FontVariant{String: keyword}
		}
	}
	return parseFontVariant(tokens, allNumericValues, couplesNumeric)
}

// //@validator()
// ``font-feature-settings`` property validation.
func fontFeatureSettings(tokens []Token) NameInt {
	if len(tokens) == 1 && getKeyword(tokens[0]) == "normal" {
		return NameInt{Name: "normal"}
	}

	//@commaSeparatedList
	fontFeatureSettingsList := func(tokens []Token) NameInt {
		var token Token
		feature, value := "", 0

		if len(tokens) == 2 {
			tokens, token = tokens[0:1], tokens[1]
			if token.Type == "ident" {
				if token.String == "on" {
					value = 1
				} else {
					value = 0
				}
			} else if token.Type == "number" && token.IntValue.Valid && token.IntValue.Int >= 0 {
				value = token.IntValue.Int
			}
		} else if len(tokens) == 1 {
			value = 1
		}

		if len(tokens) == 1 {
			token = tokens[0]
			if token.Type == "string" && len(token.String) == 4 {
				ok := true
				for _, letter := range token.String {
					if !(0x20 <= letter && letter <= 0x7f) {
						ok = false
						break
					}
				}
				if ok {
					feature = token.String
				}
			}
		}

		if feature != "" && value != 0 {
			return NameInt{Name: feature, Value: value}
		}
	}

	return fontFeatureSettingsList(tokens)
}

//@validator()
//@singleKeyword
func fontVariantAlternates(keyword string) bool {
	// TODO: support other values
	// See https://www.w3.org/TR/css-fonts-3/#font-variant-caps-prop
	return keyword == "normal" || keyword == "historical-forms"
}

// //@validator()
// func fontVariantEastAsian(tokens []Token) {
//     if len(tokens) == 1 {
//         keyword = getKeyword(tokens[0])
//         if keyword == "normal" {
//             return keyword
//         }
//     } values = []
//     couples = (
//         ("jis78", "jis83", "jis90", "jis04", "simplified", "traditional"),
//         ("full-width", "proportional-width"),
//         ("ruby",))
//     allValues = []
//     for couple := range couples {
//         allValues.extend(couple)
//     } for token := range tokens {
//         if token.type != "ident" {
//             return None
//         } if token.value := range allValues {
//             concurrentValues = [
//                 couple for couple := range couples if token.value := range couple][0]
//             if any(value := range values for value := range concurrentValues) {
//                 return None
//             } else {
//                 values.append(token.value)
//             }
//         } else {
//             return None
//         }
//     } if values {
//         return tuple(values)
//     }
// }

// //@validator()
// //@singleToken
// // ``font-size`` property validation.
// func fontSize(token Token) {
//     length = getLength(token, negative=false, percentage=true)
//     if length {
//         return length
//     } fontSizeKeyword = getKeyword(token)
//     if fontSizeKeyword := range ("smaller", "larger") {
//         raise InvalidValues("value not supported yet")
//     } if fontSizeKeyword := range computedValues.FONTSIZEKEYWORDS {
//         // || keyword := range ("smaller", "larger")
//         return fontSizeKeyword
//     }
// }

// //@validator()
// //@singleKeyword
// // ``font-style`` property validation.
// func fontStyle(keyword) {
//     return keyword := range ("normal", "italic", "oblique")
// }

// //@validator()
// //@singleKeyword
// // Validation for the ``font-stretch`` property.
// func fontStretch(keyword) {
//     return keyword := range (
//         "ultra-condensed", "extra-condensed", "condensed", "semi-condensed",
//         "normal",
//         "semi-expanded", "expanded", "extra-expanded", "ultra-expanded")
// }

// //@validator()
// //@singleToken
// // ``font-weight`` property validation.
// func fontWeight(token Token) {
//     keyword = getKeyword(token)
//     if keyword := range ("normal", "bold", "bolder", "lighter") {
//         return keyword
//     } if token.type == "number" && token.intValue is not None {
//         if token.intValue := range (100, 200, 300, 400, 500, 600, 700, 800, 900) {
//             return token.intValue
//         }
//     }
// }

// //@validator(unstable=true)
// //@singleToken
// func imageResolution(token Token) {
//     // TODO: support "snap" && "from-image"
//     return getResolution(token)
// }

// //@validator("letter-spacing")
// //@validator("word-spacing")
// //@singleToken
// // Validation for ``letter-spacing`` && ``word-spacing``.
// func spacing(token Token) {
//     if getKeyword(token) == "normal" {
//         return "normal"
//     } length = getLength(token)
//     if length {
//         return length
//     }
// }

// //@validator()
// //@singleToken
// // ``line-height`` property validation.
// func lineHeight(token Token) {
//     if getKeyword(token) == "normal" {
//         return "normal"
//     } if token.type == "number" && token.value >= 0 {
//         return Dimension(token.value, None)
//     } if token.type == "percentage" && token.value >= 0 {
//         return Dimension(token.value, "%")
//     } else if token.type == "dimension" && token.value >= 0 {
//         return getLength(token)
//     }
// }

// //@validator()
// //@singleKeyword
// // ``list-style-position`` property validation.
// func listStylePosition(keyword) {
//     return keyword := range ("inside", "outside")
// }

// //@validator()
// //@singleKeyword
// // ``list-style-type`` property validation.
// func listStyleType(keyword) {
//     return keyword := range ("none", "decimal") || keyword := range counters.STYLES
// }

// //@validator("padding-top")
// //@validator("padding-right")
// //@validator("padding-bottom")
// //@validator("padding-left")
// //@validator("min-width")
// //@validator("min-height")
// //@singleToken
// // ``padding-*`` properties validation.
// func lengthOrPrecentage(token Token) {
//     length = getLength(token, negative=false, percentage=true)
//     if length {
//         return length
//     }
// }

// //@validator("max-width")
// //@validator("max-height")
// //@singleToken
// // Validation for max-width && max-height
// func maxWidthHeight(token Token) {
//     length = getLength(token, negative=false, percentage=true)
//     if length {
//         return length
//     } if getKeyword(token) == "none" {
//         return Dimension(float("inf"), "px")
//     }
// }

// //@validator()
// //@singleToken
// // Validation for the ``opacity`` property.
// func opacity(token Token) {
//     if token.type == "number" {
//         return min(1, max(0, token.value))
//     }
// }

// //@validator()
// //@singleToken
// // Validation for the ``z-index`` property.
// func zIndex(token Token) {
//     if getKeyword(token) == "auto" {
//         return "auto"
//     } if token.type == "number" && token.intValue is not None {
//         return token.intValue
//     }
// }

// //@validator("orphans")
// //@validator("widows")
// //@singleToken
// // Validation for the ``orphans`` && ``widows`` properties.
// func orphansWidows(token Token) {
//     if token.type == "number" && token.intValue is not None {
//         value = token.intValue
//         if value >= 1 {
//             return value
//         }
//     }
// }

// //@validator()
// //@singleToken
// // Validation for the ``column-count`` property.
// func columnCount(token Token) {
//     if token.type == "number" && token.intValue is not None {
//         value = token.intValue
//         if value >= 1 {
//             return value
//         }
//     } if getKeyword(token) == "auto" {
//         return "auto"
//     }
// }

// //@validator()
// //@singleKeyword
// // Validation for the ``overflow`` property.
// func overflow(keyword) {
//     return keyword := range ("auto", "visible", "hidden", "scroll")
// }

// //@validator()
// //@singleKeyword
// // ``position`` property validation.
// func position(keyword) {
//     return keyword := range ("static", "relative", "absolute", "fixed")
// }

// //@validator()
// // ``quotes`` property validation.
// func quotes(tokens []Token) {
//     if (tokens && len(tokens) % 2 == 0 and
//             all(v.type == "string" for v := range tokens)) {
//             }
//         strings = tuple(token.value for token := range tokens)
//         // Separate open && close quotes.
//         // eg.  ("«", "»", "“", "”")  -> (("«", "“"), ("»", "”"))
//         return strings[::2], strings[1::2]
// }

// //@validator()
// //@singleKeyword
// // Validation for the ``table-layout`` property
// func tableLayout(keyword) {
//     if keyword := range ("fixed", "auto") {
//         return keyword
//     }
// }

// //@validator()
// //@singleKeyword
// // ``text-align`` property validation.
// func textAlign(keyword) {
//     return keyword := range ("left", "right", "center", "justify")
// }

// //@validator()
// // ``text-decoration`` property validation.
// func textDecoration(tokens []Token) {
//     keywords = [getKeyword(v) for v := range tokens]
//     if keywords == ["none"] {
//         return "none"
//     } if all(keyword := range ("underline", "overline", "line-through", "blink")
//             for keyword := range keywords) {
//             }
//         unique = set(keywords)
//         if len(unique) == len(keywords) {
//             // No duplicate
//             // blink is accepted but ignored
//             // "Conforming user agents may simply not blink the text."
//             return frozenset(unique - set(["blink"]))
//         }
// }

// //@validator()
// //@singleToken
// // ``text-indent`` property validation.
// func textIndent(token Token) {
//     length = getLength(token, percentage=true)
//     if length {
//         return length
//     }
// }

// //@validator()
// //@singleKeyword
// // ``text-align`` property validation.
// func textTransform(keyword) {
//     return keyword := range (
//         "none", "uppercase", "lowercase", "capitalize", "full-width")
// }

// //@validator()
// //@singleToken
// // Validation for the ``vertical-align`` property
// func verticalAlign(token Token) {
//     length = getLength(token, percentage=true)
//     if length {
//         return length
//     } keyword = getKeyword(token)
//     if keyword := range ("baseline", "middle", "sub", "super",
//                    "text-top", "text-bottom", "top", "bottom") {
//                    }
//         return keyword
// }

// //@validator()
// //@singleKeyword
// // ``white-space`` property validation.
// func visibility(keyword) {
//     return keyword := range ("visible", "hidden", "collapse")
// }

// //@validator()
// //@singleKeyword
// // ``white-space`` property validation.
// func whiteSpace(keyword) {
//     return keyword := range ("normal", "pre", "nowrap", "pre-wrap", "pre-line")
// }

// //@validator()
// //@singleKeyword
// // ``overflow-wrap`` property validation.
// func overflowWrap(keyword) {
//     return keyword := range ("normal", "break-word")
// }

// //@validator(unstable=true)
// //@singleKeyword
// // Validation for ``image-rendering``.
// func imageRendering(keyword) {
//     return keyword := range ("auto", "crisp-edges", "pixelated")
// }

// //@validator(unstable=true)
// // ``size`` property validation.
// //     See http://www.w3.org/TR/css3-page/#page-size-prop
// //
// func size(tokens []Token) {
//     lengths = [getLength(token, negative=false) for token := range tokens]
//     if all(lengths) {
//         if len(lengths) == 1 {
//             return (lengths[0], lengths[0])
//         } else if len(lengths) == 2 {
//             return tuple(lengths)
//         }
//     }
// }
//     keywords = [getKeyword(v) for v := range tokens]
//     if len(keywords) == 1 {
//         keyword = keywords[0]
//         if keyword := range computedValues.PAGESIZES {
//             return computedValues.PAGESIZES[keyword]
//         } else if keyword := range ("auto", "portrait") {
//             return computedValues.INITIALPAGESIZE
//         } else if keyword == "landscape" {
//             return computedValues.INITIALPAGESIZE[::-1]
//         }
//     }

//     if len(keywords) == 2 {
//         if keywords[0] := range ("portrait", "landscape") {
//             orientation, pageSize = keywords
//         } else if keywords[1] := range ("portrait", "landscape") {
//             pageSize, orientation = keywords
//         } else {
//             pageSize = None
//         } if pageSize := range computedValues.PAGESIZES {
//             widthHeight = computedValues.PAGESIZES[pageSize]
//             if orientation == "portrait" {
//                 return widthHeight
//             } else {
//                 height, width = widthHeight
//                 return width, height
//             }
//         }
//     }

// //@validator(proprietary=true)
// //@singleToken
// // Validation for ``anchor``.
// func anchor(token Token) {
//     if getKeyword(token) == "none" {
//         return "none"
//     } function = parseFunction(token)
//     if function {
//         name, args = function
//         prototype = (name, [a.type for a := range args])
//         args = [getattr(a, "value", a) for a := range args]
//         if prototype == ("attr", ["ident"]) {
//             return (name, args[0])
//         }
//     }
// }

// //@validator(proprietary=true, wantsBaseUrl=true)
// //@singleToken
// // Validation for ``link``.
// func link(token, baseUrl) {
//     if getKeyword(token) == "none" {
//         return "none"
//     } else if token.type == "url" {
//         if token.value.startswith("#") {
//             return "internal", unquote(token.value[1:])
//         } else {
//             return "external", safeUrljoin(baseUrl, token.value)
//         }
//     } function = parseFunction(token)
//     if function {
//         name, args = function
//         prototype = (name, [a.type for a := range args])
//         args = [getattr(a, "value", a) for a := range args]
//         if prototype == ("attr", ["ident"]) {
//             return (name, args[0])
//         }
//     }
// }

// //@validator()
// //@singleToken
// // Validation for ``tab-size``.
// //     See https://www.w3.org/TR/css-text-3/#tab-size
// //
// func tabSize(token Token) {
//     if token.type == "number" && token.intValue is not None {
//         value = token.intValue
//         if value >= 0 {
//             return value
//         }
//     } return getLength(token, negative=false)
// }

// //@validator(unstable=true)
// //@singleToken
// // Validation for ``hyphens``.
// func hyphens(token Token) {
//     keyword = getKeyword(token)
//     if keyword := range ("none", "manual", "auto") {
//         return keyword
//     }
// }

// //@validator(unstable=true)
// //@singleToken
// // Validation for ``hyphenate-character``.
// func hyphenateCharacter(token Token) {
//     keyword = getKeyword(token)
//     if keyword == "auto" {
//         return "‐"
//     } else if token.type == "string" {
//         return token.value
//     }
// }

// //@validator(unstable=true)
// //@singleToken
// // Validation for ``hyphenate-limit-zone``.
// func hyphenateLimitZone(token Token) {
//     return getLength(token, negative=false, percentage=true)
// }

// //@validator(unstable=true)
// // Validation for ``hyphenate-limit-chars``.
// func hyphenateLimitChars(tokens []Token) {
//     if len(tokens) == 1 {
//         token, = tokens
//         keyword = getKeyword(token)
//         if keyword == "auto" {
//             return (5, 2, 2)
//         } else if token.type == "number" && token.intValue is not None {
//             return (token.intValue, 2, 2)
//         }
//     } else if len(tokens) == 2 {
//         total, left = tokens
//         totalKeyword = getKeyword(total)
//         leftKeyword = getKeyword(left)
//         if total.type == "number" && total.intValue is not None {
//             if left.type == "number" && left.intValue is not None {
//                 return (total.intValue, left.intValue, left.intValue)
//             } else if leftKeyword == "auto" {
//                 return (total.value, 2, 2)
//             }
//         } else if totalKeyword == "auto" {
//             if left.type == "number" && left.intValue is not None {
//                 return (5, left.intValue, left.intValue)
//             } else if leftKeyword == "auto" {
//                 return (5, 2, 2)
//             }
//         }
//     } else if len(tokens) == 3 {
//         total, left, right = tokens
//         if (
//             (getKeyword(total) == "auto" or
//                 (total.type == "number" && total.intValue is not None)) and
//             (getKeyword(left) == "auto" or
//                 (left.type == "number" && left.intValue is not None)) and
//             (getKeyword(right) == "auto" or
//                 (right.type == "number" && right.intValue is not None))
//         ) {
//             total = total.intValue if total.type == "number" else 5
//             left = left.intValue if left.type == "number" else 2
//             right = right.intValue if right.type == "number" else 2
//             return (total, left, right)
//         }
//     }
// }

// //@validator(proprietary=true)
// //@singleToken
// // Validation for ``lang``.
// func lang(token Token) {
//     if getKeyword(token) == "none" {
//         return "none"
//     } function = parseFunction(token)
//     if function {
//         name, args = function
//         prototype = (name, [a.type for a := range args])
//         args = [getattr(a, "value", a) for a := range args]
//         if prototype == ("attr", ["ident"]) {
//             return (name, args[0])
//         }
//     } else if token.type == "string" {
//         return ("string", token.value)
//     }
// }

// //@validator(unstable=true)
// // Validation for ``bookmark-label``.
// func bookmarkLabel(tokens []Token) {
//     parsedTokens = tuple(validateContentListToken(v) for v := range tokens)
//     if None not := range parsedTokens {
//         return parsedTokens
//     }
// }

// //@validator(unstable=true)
// //@singleToken
// // Validation for ``bookmark-level``.
// func bookmarkLevel(token Token) {
//     if token.type == "number" && token.intValue is not None {
//         value = token.intValue
//         if value >= 1 {
//             return value
//         }
//     } else if getKeyword(token) == "none" {
//         return "none"
//     }
// }

// //@validator(unstable=true)
// //@commaSeparatedList
// // Validation for ``string-set``.
// func stringSet(tokens []Token) {
//     if len(tokens) >= 2 {
//         varName = getKeyword(tokens[0])
//         parsedTokens = tuple(
//             validateContentListToken(v) for v := range tokens[1:])
//         if None not := range parsedTokens {
//             return (varName, parsedTokens)
//         }
//     } else if tokens && tokens[0].value == "none" {
//         return "none"
//     }
// }

// // Validation for a single token of <content-list> used := range GCPM.
// //     Return (type, content) || false for invalid tokens.
// //
// func validateContentListToken(token Token) {
//     type_ = token.type
//     if type_ == "string" {
//         return ("STRING", token.value)
//     } function = parseFunction(token)
//     if function {
//         name, args = function
//         prototype = (name, tuple(a.type for a := range args))
//         args = tuple(getattr(a, "value", a) for a := range args)
//         if prototype == ("attr", ("ident",)) {
//             return (name, args[0])
//         } else if prototype := range (("content", ()), ("content", ("ident",))) {
//             if not args {
//                 return (name, "text")
//             } else if args[0] := range ("text", "after", "before", "first-letter") {
//                 return (name, args[0])
//             }
//         } else if prototype := range (("counter", ("ident",)),
//                            ("counters", ("ident", "string"))) {
//                            }
//             args += ("decimal",)
//             return (name, args)
//         else if prototype := range (("counter", ("ident", "ident")),
//                            ("counters", ("ident", "string", "ident"))) {
//                            }
//             style = args[-1]
//             if style := range ("none", "decimal") || style := range counters.STYLES {
//                 return (name, args)
//             }
//     }
// }

// //@validator(unstable=true)
// func transform(tokens []Token) {
//     if getSingleKeyword(tokens) == "none" {
//         return ()
//     } else {
//         return tuple(transformFunction(v) for v := range tokens)
//     }
// }

// func transformFunction(token Token) {
//     function = parseFunction(token)
//     if not function {
//         raise InvalidValues
//     } name, args = function
// }
//     if len(args) == 1 {
//         angle = getAngle(args[0])
//         length = getLength(args[0], percentage=true)
//         if name := range ("rotate", "skewx", "skewy") && angle {
//             return name, angle
//         } else if name := range ("translatex", "translate") && length {
//             return "translate", (length, computedValues.ZEROPIXELS)
//         } else if name == "translatey" && length {
//             return "translate", (computedValues.ZEROPIXELS, length)
//         } else if name == "scalex" && args[0].type == "number" {
//             return "scale", (args[0].value, 1)
//         } else if name == "scaley" && args[0].type == "number" {
//             return "scale", (1, args[0].value)
//         } else if name == "scale" && args[0].type == "number" {
//             return "scale", (args[0].value,) * 2
//         }
//     } else if len(args) == 2 {
//         if name == "scale" && all(a.type == "number" for a := range args) {
//             return name, tuple(arg.value for arg := range args)
//         } lengths = tuple(getLength(token, percentage=true) for token := range args)
//         if name == "translate" && all(lengths) {
//             return name, lengths
//         }
//     } else if len(args) == 6 && name == "matrix" && all(
//             a.type == "number" for a := range args) {
//             }
//         return name, tuple(arg.value for arg := range args)
//     raise InvalidValues

// // Expanders

// // Let"s be consistent, always use ``name`` as an argument even
// // when it is useless.
// // pylint: disable=W0613

// // Decorator adding a function to the ``EXPANDERS``.
// func expander(propertyName) {
//     def expanderDecorator(function) {
//         """Add ``function`` to the ``EXPANDERS``."""
//         assert propertyName not := range EXPANDERS, propertyName
//         EXPANDERS[propertyName] = function
//         return function
//     } return expanderDecorator
// }

// //@expander("border-color")
// //@expander("border-style")
// //@expander("border-width")
// //@expander("margin")
// //@expander("padding")
// //@expander("bleed")
// // Expand properties setting a token for the four sides of a box.
// func expandFourSides(baseUrl, name, tokens []Token) {
//     // Make sure we have 4 tokens
//     if len(tokens) == 1 {
//         tokens *= 4
//     } else if len(tokens) == 2 {
//         tokens *= 2  // (bottom, left) defaults to (top, right)
//     } else if len(tokens) == 3 {
//         tokens += (tokens[1],)  // left defaults to right
//     } else if len(tokens) != 4 {
//         raise InvalidValues(
//             "Expected 1 to 4 token components got %i" % len(tokens))
//     } for suffix, token := range zip(("-top", "-right", "-bottom", "-left"), tokens []Token) {
//         i = name.rfind("-")
//         if i == -1 {
//             newName = name + suffix
//         } else {
//             // eg. border-color becomes border-*-color, not border-color-*
//             newName = name[:i] + suffix + name[i:]
//         }
//     }
// }
//         // validateNonShorthand returns ((name, value),), we want
//         // to yield (name, value)
//         result, = validateNonShorthand(
//             baseUrl, newName, [token], required=true)
//         yield result

// //@expander("border-radius")
// // Validator for the `border-radius` property.
// func borderRadius(baseUrl, name, tokens []Token) {
//     current = horizontal = []
//     vertical = []
//     for token := range tokens {
//         if token.type == "literal" && token.value == "/" {
//             if current is horizontal {
//                 if token == tokens[-1] {
//                     raise InvalidValues("Expected value after "/" separator")
//                 } else {
//                     current = vertical
//                 }
//             } else {
//                 raise InvalidValues("Expected only one "/" separator")
//             }
//         } else {
//             current.append(token)
//         }
//     }
// }
//     if not vertical {
//         vertical = horizontal[:]
//     }

//     for values := range horizontal, vertical {
//         // Make sure we have 4 tokens
//         if len(values) == 1 {
//             values *= 4
//         } else if len(values) == 2 {
//             values *= 2  // (br, bl) defaults to (tl, tr)
//         } else if len(values) == 3 {
//             values.append(values[1])  // bl defaults to tr
//         } else if len(values) != 4 {
//             raise InvalidValues(
//                 "Expected 1 to 4 token components got %i" % len(values))
//         }
//     } corners = ("top-left", "top-right", "bottom-right", "bottom-left")
//     for corner, tokens := range zip(corners, zip(horizontal, vertical)) {
//         newName = "border-%s-radius" % corner
//         // validateNonShorthand returns [(name, value)], we want
//         // to yield (name, value)
//         result, = validateNonShorthand(
//             baseUrl, newName, tokens, required=true)
//         yield result
//     }

// // Decorator helping expanders to handle ``inherit`` && ``initial``.
// //     Wrap an expander so that it does not have to handle the "inherit" and
// //     "initial" cases, && can just yield name suffixes. Missing suffixes
// //     get the initial value.
// //
// func genericExpander(*expandedNames, **kwargs) {
//     wantsBaseUrl = kwargs.pop("wantsBaseUrl", false)
//     assert not kwargs
// }
//     def genericExpanderDecorator(wrapped) {
//         """Decorate the ``wrapped`` expander."""
//         //@functools.wraps(wrapped)
//         def genericExpanderWrapper(baseUrl, name, tokens []Token) {
//             """Wrap the expander."""
//             keyword = getSingleKeyword(tokens)
//             if keyword := range ("inherit", "initial") {
//                 results = dict.fromkeys(expandedNames, keyword)
//                 skipValidation = true
//             } else {
//                 skipValidation = false
//                 results = {}
//                 if wantsBaseUrl {
//                     result = wrapped(name, tokens, baseUrl)
//                 } else {
//                     result = wrapped(name, tokens)
//                 } for newName, newToken := range result {
//                     assert newName := range expandedNames, newName
//                     if newName := range results {
//                         raise InvalidValues(
//                             "got multiple %s values := range a %s shorthand"
//                             % (newName.strip("-"), name))
//                     } results[newName] = newToken
//                 }
//             }
//         }
//     }

//             for newName := range expandedNames {
//                 if newName.startswith("-") {
//                     // newName is a suffix
//                     actualNewName = name + newName
//                 } else {
//                     actualNewName = newName
//                 }
//             }

//                 if newName := range results {
//                     value = results[newName]
//                     if not skipValidation {
//                         // validateNonShorthand returns ((name, value),)
//                         (actualNewName, value), = validateNonShorthand(
//                             baseUrl, actualNewName, value, required=true)
//                     }
//                 } else {
//                     value = "initial"
//                 }

//                 yield actualNewName, value
//         return genericExpanderWrapper
//     return genericExpanderDecorator

// //@expander("list-style")
// //@genericExpander("-type", "-position", "-image", wantsBaseUrl=true)
// // Expand the ``list-style`` shorthand property.
// //     See http://www.w3.org/TR/CSS21/generate.html#propdef-list-style
// //
// func expandListStyle(name, tokens, baseUrl) {
//     typeSpecified = imageSpecified = false
//     noneCount = 0
//     for token := range tokens {
//         if getKeyword(token) == "none" {
//             // Can be either -style || -image, see at the end which is not
//             // otherwise specified.
//             noneCount += 1
//             noneToken = token
//             continue
//         }
//     }
// }
//         if listStyleType([token]) is not None {
//             suffix = "-type"
//             typeSpecified = true
//         } else if listStylePosition([token]) is not None {
//             suffix = "-position"
//         } else if imageUrl([token], baseUrl) is not None {
//             suffix = "-image"
//             imageSpecified = true
//         } else {
//             raise InvalidValues
//         } yield suffix, [token]

//     if not typeSpecified && noneCount {
//         yield "-type", [noneToken]
//         noneCount -= 1
//     }

//     if not imageSpecified && noneCount {
//         yield "-image", [noneToken]
//         noneCount -= 1
//     }

//     if noneCount {
//         // Too many none tokens.
//         raise InvalidValues
//     }

// //@expander("border")
// // Expand the ``border`` shorthand property.
// //     See http://www.w3.org/TR/CSS21/box.html#propdef-border
// //
// func expandBorder(baseUrl, name, tokens []Token) {
//     for suffix := range ("-top", "-right", "-bottom", "-left") {
//         for newProp := range expandBorderSide(baseUrl, name + suffix, tokens []Token) {
//             yield newProp
//         }
//     }
// }

// //@expander("border-top")
// //@expander("border-right")
// //@expander("border-bottom")
// //@expander("border-left")
// //@expander("column-rule")
// //@expander("outline")
// //@genericExpander("-width", "-color", "-style")
// // Expand the ``border-*`` shorthand properties.
// //     See http://www.w3.org/TR/CSS21/box.html#propdef-border-top
// //
// func expandBorderSide(name, tokens []Token) {
//     for token := range tokens {
//         if parseColor(token) is not None {
//             suffix = "-color"
//         } else if borderWidth([token]) is not None {
//             suffix = "-width"
//         } else if borderStyle([token]) is not None {
//             suffix = "-style"
//         } else {
//             raise InvalidValues
//         } yield suffix, [token]
//     }
// }

// //@expander("background")
// // Expand the ``background`` shorthand property.
// //     See http://dev.w3.org/csswg/css3-background/#the-background
// //
// func expandBackground(baseUrl, name, tokens []Token) {
//     properties = [
//         "backgroundColor", "backgroundImage", "backgroundRepeat",
//         "backgroundAttachment", "backgroundPosition", "backgroundSize",
//         "backgroundClip", "backgroundOrigin"]
//     keyword = getSingleKeyword(tokens)
//     if keyword := range ("initial", "inherit") {
//         for name := range properties {
//             yield name, keyword
//         } return
//     }
// }
//     def parseLayer(tokens, finalLayer=false) {
//         results = {}
//     }

//         def add(name, value) {
//             if value is None {
//                 return false
//             } name = "background" + name
//             if name := range results {
//                 raise InvalidValues
//             } results[name] = value
//             return true
//         }

//         // Make `tokens` a stack
//         tokens = tokens[::-1]
//         while tokens {
//             if add("repeat",
//                    backgroundRepeat.singleValue(tokens[-2:][::-1])) {
//                    }
//                 del tokens[-2:]
//                 continue
//             token = tokens[-1:]
//             if finalLayer && add("color", otherColors(token)) {
//                 tokens.pop()
//                 continue
//             } if add("image", backgroundImage.singleValue(token, baseUrl)) {
//                 tokens.pop()
//                 continue
//             } if add("repeat", backgroundRepeat.singleValue(token)) {
//                 tokens.pop()
//                 continue
//             } if add("attachment", backgroundAttachment.singleValue(token)) {
//                 tokens.pop()
//                 continue
//             } for n := range (4, 3, 2, 1)[-len(tokens):] {
//                 nTokens = tokens[-n:][::-1]
//                 position = backgroundPosition.singleValue(nTokens)
//                 if position is not None {
//                     assert add("position", position)
//                     del tokens[-n:]
//                     if (tokens && tokens[-1].type == "literal" and
//                             tokens[-1].value == "/") {
//                             }
//                         for n := range (3, 2)[-len(tokens):] {
//                             // n includes the "/" delimiter.
//                             nTokens = tokens[-n:-1][::-1]
//                             size = backgroundSize.singleValue(nTokens)
//                             if size is not None {
//                                 assert add("size", size)
//                                 del tokens[-n:]
//                             }
//                         }
//                     break
//                 }
//             } if position is not None {
//                 continue
//             } if add("origin", box.singleValue(token)) {
//                 tokens.pop()
//                 nextToken = tokens[-1:]
//                 if add("clip", box.singleValue(nextToken)) {
//                     tokens.pop()
//                 } else {
//                     // The same keyword sets both {
//                     } assert add("clip", box.singleValue(token))
//                 } continue
//             } raise InvalidValues
//         }

//         color = results.pop(
//             "backgroundColor", INITIALVALUES["backgroundColor"])
//         for name := range properties {
//             if name not := range results {
//                 results[name] = INITIALVALUES[name][0]
//             }
//         } return color, results

//     layers = reversed(splitOnComma(tokens))
//     color, lastLayer = parseLayer(next(layers), finalLayer=true)
//     results = dict((k, [v]) for k, v := range lastLayer.items())
//     for tokens := range layers {
//         _, layer = parseLayer(tokens)
//         for name, value := range layer.items() {
//             results[name].append(value)
//         }
//     } for name, values := range results.items() {
//         yield name, values[::-1]  // "Un-reverse"
//     } yield "background-color", color

// //@expander("page-break-after")
// //@expander("page-break-before")
// // Expand legacy ``page-break-before`` && ``page-break-after`` properties.
// //     See https://www.w3.org/TR/css-break-3/#page-break-properties
// //
// func expandPageBreakBeforeAfter(baseUrl, name, tokens []Token) {
//     keyword = getSingleKeyword(tokens)
//     newName = name.split("-", 1)[1]
//     if keyword := range ("auto", "left", "right", "avoid") {
//         yield newName, keyword
//     } else if keyword == "always" {
//         yield newName, "page"
//     }
// }

// //@expander("page-break-inside")
// // Expand the legacy ``page-break-inside`` property.
// //     See https://www.w3.org/TR/css-break-3/#page-break-properties
// //
// func expandPageBreakInside(baseUrl, name, tokens []Token) {
//     keyword = getSingleKeyword(tokens)
//     if keyword := range ("auto", "avoid") {
//         yield "break-inside", keyword
//     }
// }

// //@expander("columns")
// //@genericExpander("column-width", "column-count")
// // Expand the ``columns`` shorthand property.
// func expandColumns(name, tokens []Token) {
//     name = None
//     if len(tokens) == 2 && getKeyword(tokens[0]) == "auto" {
//         tokens = tokens[::-1]
//     } for token := range tokens {
//         if columnWidth([token]) is not None && name != "column-width" {
//             name = "column-width"
//         } else if columnCount([token]) is not None {
//             name = "column-count"
//         } else {
//             raise InvalidValues
//         } yield name, [token]
//     }
// }

// class NoneFakeToken(object) {
//     type = "ident"
//     lowerValue = "none"
// }

// class NormalFakeToken(object) {
//     type = "ident"
//     lowerValue = "normal"
// }

// //@expander("font-variant")
// //@genericExpander("-alternates", "-caps", "-east-asian", "-ligatures",
//                   "-numeric", "-position")
// // Expand the ``font-variant`` shorthand property.
// //     https://www.w3.org/TR/css-fonts-3/#font-variant-prop
// //
// func fontVariant(name, tokens []Token) {
//     return expandFontVariant(tokens)
// }

// func expandFontVariant(tokens []Token) {
//     keyword = getSingleKeyword(tokens)
//     if keyword := range ("normal", "none") {
//         for suffix := range (
//                 "-alternates", "-caps", "-east-asian", "-numeric",
//                 "-position") {
//                 }
//             yield suffix, [NormalFakeToken]
//         token = NormalFakeToken if keyword == "normal" else NoneFakeToken
//         yield "-ligatures", [token]
//     } else {
//         features = {
//             "alternates": [],
//             "caps": [],
//             "east-asian": [],
//             "ligatures": [],
//             "numeric": [],
//             "position": []}
//         for token := range tokens {
//             keyword = getKeyword(token)
//             if keyword == "normal" {
//                 // We don"t allow "normal", only the specific values
//                 raise InvalidValues
//             } for feature := range features {
//                 functionName = "fontVariant%s" % feature.replace("-", "")
//                 if globals()[functionName]([token]) {
//                     features[feature].append(token)
//                     break
//                 }
//             } else {
//                 raise InvalidValues
//             }
//         } for feature, tokens := range features.items() {
//             if tokens {
//                 yield "-%s" % feature, tokens
//             }
//         }
//     }
// }

// //@expander("font")
// //@genericExpander("-style", "-variant-caps", "-weight", "-stretch", "-size",
//                   "line-height", "-family")  // line-height is not a suffix
// // Expand the ``font`` shorthand property.
// //     https://www.w3.org/TR/css-fonts-3/#font-prop
// //
// func expandFont(name, tokens []Token) {
//     expandFontKeyword = getSingleKeyword(tokens)
//     if expandFontKeyword := range ("caption", "icon", "menu", "message-box",
//                                "small-caption", "status-bar") {
//                                }
//         raise InvalidValues("System fonts are not supported")
// }
//     // Make `tokens` a stack
//     tokens = list(reversed(tokens))
//     // Values for font-style font-variant && font-weight can come := range any
//     // order && are all optional.
//     while tokens {
//         token = tokens.pop()
//         if getKeyword(token) == "normal" {
//             // Just ignore "normal" keywords. Unspecified properties will get
//             // their initial token, which is "normal" for all three here.
//             // TODO: fail if there is too many "normal" values?
//             continue
//         }
//     }

//         if fontStyle([token]) is not None {
//             suffix = "-style"
//         } else if getKeyword(token) := range ("normal", "small-caps") {
//             suffix = "-variant-caps"
//         } else if fontWeight([token]) is not None {
//             suffix = "-weight"
//         } else if fontStretch([token]) is not None {
//             suffix = "-stretch"
//         } else {
//             // We’re done with these three, continue with font-size
//             break
//         } yield suffix, [token]

//     // Then font-size is mandatory
//     // Latest `token` from the loop.
//     if fontSize([token]) is None {
//         raise InvalidValues
//     } yield "-size", [token]

//     // Then line-height is optional, but font-family is not so the list
//     // must not be empty yet
//     if not tokens {
//         raise InvalidValues
//     }

//     token = tokens.pop()
//     if token.type == "literal" && token.value == "/" {
//         token = tokens.pop()
//         if lineHeight([token]) is None {
//             raise InvalidValues
//         } yield "line-height", [token]
//     } else {
//         // We pop()ed a font-family, add it back
//         tokens.append(token)
//     }

//     // Reverse the stack to get normal list
//     tokens.reverse()
//     if fontFamily(tokens) is None {
//         raise InvalidValues
//     } yield "-family", tokens

// //@expander("word-wrap")
// // Expand the ``word-wrap`` legacy property.
// //     See http://http://www.w3.org/TR/css3-text/#overflow-wrap
// //
// func expandWordWrap(baseUrl, name, tokens []Token) {
//     keyword = overflowWrap(tokens)
//     if keyword is None {
//         raise InvalidValues
//     }
// }
//     yield "overflow-wrap", keyword

// // Default validator for non-shorthand properties.
// func validateNonShorthand(baseUrl, name, tokens, required=false) {
//     if not required && name not := range KNOWNPROPERTIES {
//         hyphensName = name.replace("", "-")
//         if hyphensName := range KNOWNPROPERTIES {
//             raise InvalidValues("did you mean %s?" % hyphensName)
//         } else {
//             raise InvalidValues("unknown property")
//         }
//     }
// }
//     if not required && name not := range VALIDATORS {
//         raise InvalidValues("property not supported yet")
//     }

//     keyword = getSingleKeyword(tokens)
//     if keyword := range ("initial", "inherit") {
//         value = keyword
//     } else {
//         function = VALIDATORS[name]
//         if function.wantsBaseUrl {
//             value = function(tokens, baseUrl)
//         } else {
//             value = function(tokens)
//         } if value is None {
//             raise InvalidValues
//         }
//     } return ((name, value),)

// //
// //     Expand shorthand properties && filter unsupported properties && values.
// //     Log a warning for every ignored declaration.
// //     Return a iterable of ``(name, value, important)`` tuples.
// //
// func preprocessDeclarations(baseUrl, declarations) {
//     for declaration := range declarations {
//         if declaration.type == "error" {
//             LOGGER.warning(
//                 "Error: %s at %i:%i.",
//                 declaration.message,
//                 declaration.sourceLine, declaration.sourceColumn)
//         }
//     }
// }
//         if declaration.type != "declaration" {
//             continue
//         }

//         name = declaration.lowerName

//         def validationError(level, reason) {
//             getattr(LOGGER, level)(
//                 "Ignored `%s:%s` at %i:%i, %s.",
//                 declaration.name, tinycss2.serialize(declaration.value),
//                 declaration.sourceLine, declaration.sourceColumn, reason)
//         }

//         if name := range NOTPRINTMEDIA {
//             validationError(
//                 "warning", "the property does not apply for the print media")
//             continue
//         }

//         if name.startswith(PREFIX) {
//             unprefixedName = name[len(PREFIX):]
//             if unprefixedName := range PROPRIETARY {
//                 name = unprefixedName
//             } else if unprefixedName := range UNSTABLE {
//                 LOGGER.warning(
//                     "Deprecated `%s:%s` at %i:%i, "
//                     "prefixes on unstable attributes are deprecated, "
//                     "use `%s` instead.",
//                     declaration.name, tinycss2.serialize(declaration.value),
//                     declaration.sourceLine, declaration.sourceColumn,
//                     unprefixedName)
//                 name = unprefixedName
//             } else {
//                 LOGGER.warning(
//                     "Ignored `%s:%s` at %i:%i, "
//                     "prefix on this attribute is not supported, "
//                     "use `%s` instead.",
//                     declaration.name, tinycss2.serialize(declaration.value),
//                     declaration.sourceLine, declaration.sourceColumn,
//                     unprefixedName)
//                 continue
//             }
//         }

//         expander_ = EXPANDERS.get(name, validateNonShorthand)
//         tokens = removeWhitespace(declaration.value)
//         try {
//             // Use list() to consume generators now && catch any error.
//             result = list(expander(baseUrl, name, tokens))
//         } except InvalidValues as exc {
//             validationError(
//                 "warning",
//                 exc.args[0] if exc.args && exc.args[0] else "invalid value")
//             continue
//         }

//         important = declaration.important
//         for longName, value := range result {
//             yield longName.replace("-", ""), value, important
//         }

// Remove any top-level whitespace in a token list.
func removeWhitespace(tokens []Token) []Token {
	var out []Token
	for _, token := range tokens {
		if token.Type != "whitespace" && token.Type != "comment" {
			out = append(out, token)
		}
	}
	return out
}

// Split a list of tokens on commas, ie ``LiteralToken(",")``.
//     Only "top-level" comma tokens are splitting points, not commas inside a
//     function or blocks.
//
func splitOnComma(tokens []Token) [][]Token {
	var parts [][]Token
	var thisPart []Token
	for _, token := range tokens {
		if token.Type == "literal" && token.String == "," {
			parts = append(parts, thisPart)
			thisPart = nil
		} else {
			thisPart = append(thisPart, token)
		}
	}
	parts = append(parts, thisPart)
	return parts
}
