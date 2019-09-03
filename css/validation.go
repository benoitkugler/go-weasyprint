package css

import (
	"errors"
	"fmt"
	"math"
	"net/url"
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
	LENGTHUNITS = map[string]Unit{"ex": Ex, "em": Em, "ch": Ch, "rem": Rem, "px": Px, "pt": Pt, "pc": Pc, "in": In, "cm": Cm, "mm": Mm, "q": Q}

	// keyword -> (open, insert)
	ContentQuoteKeywords = map[string]quote{
		"open-quote":     {open: true, insert: true},
		"close-quote":    {open: false, insert: true},
		"no-open-quote":  {open: true, insert: false},
		"no-close-quote": {open: false, insert: false},
	}

	ZEROPERCENT    = Dimension{Value: 0, Unit: Percentage}
	fiftyPercent   = Dimension{Value: 50, Unit: Percentage}
	HUNDREDPERCENT = Dimension{Value: 100, Unit: Percentage}

	backgroundPositionsPercentages = map[string]Dimension{
		"top":    ZEROPERCENT,
		"left":   ZEROPERCENT,
		"center": fiftyPercent,
		"bottom": HUNDREDPERCENT,
		"right":  HUNDREDPERCENT,
	}

	// yes/no validators for non-shorthand properties
	// Maps property names to functions taking a property name and a value list,
	// returning a value or None for invalid.
	// Also transform values: keyword && URLs are returned as strings.
	// For properties that take a single value, that value is returned by itself
	// instead of a list.
	validators = map[string]validator{
		"background-attachment":      backgroundAttachment,
		"background-color":           otherColors,
		"border-top-color":           otherColors,
		"border-right-color":         otherColors,
		"border-bottom-color":        otherColors,
		"border-left-color":          otherColors,
		"column-rule-color":          otherColors,
		"outline-color":              outlineColor,
		"border-collapse":            borderCollapse,
		"empty-cells":                emptyCells,
		"color":                      color,
		"transform-origin":           transformOrigin,
		"background-position":        valideBackgroundPosition,
		"background-repeat":          backgroundRepeat,
		"background-size":            backgroundSize,
		"background-clip":            box,
		"background-origin":          box,
		"border-spacing":             borderSpacing,
		"border-top-right-radius":    borderCornerRadius,
		"border-bottom-right-radius": borderCornerRadius,
		"border-bottom-left-radius":  borderCornerRadius,
		"border-top-left-radius":     borderCornerRadius,
		"border-top-style":           borderStyle,
		"border-right-style":         borderStyle,
		"border-left-style":          borderStyle,
		"border-bottom-style":        borderStyle,
		"column-rule-style":          borderStyle,
		"break-before":               breakBeforeAfter,
		"break-after":                breakBeforeAfter,
		"page":                       page,
		"bleed-left":                 bleed,
		"bleed-right":                bleed,
		"bleed-top":                  bleed,
		"bleed-bottom":               bleed,
		"marks":                      marks,
		"outline-style":              outlineStyle,
		"border-top-width":           borderWidth,
		"border-right-width":         borderWidth,
		"border-left-width":          borderWidth,
		"border-bottom-width":        borderWidth,
		"colunm-rule-width":          borderWidth,
		"outline-width":              borderWidth,
		"column-width":               columnWidth,
		"column-span":                columnSpan,
		"box-sizing":                 boxSizing,
		"caption-side":               captionSide,
		"clear":                      clear,
		"clip":                       clip,
		"top":                        lengthPercOrAuto,
		"right":                      lengthPercOrAuto,
		"left":                       lengthPercOrAuto,
		"bottom":                     lengthPercOrAuto,
		"margin-top":                 lengthPercOrAuto,
		"margin-right":               lengthPercOrAuto,
		"margin-bottom":              lengthPercOrAuto,
		"margin-left":                lengthPercOrAuto,
		"height":                     widthHeight,
		"width":                      widthHeight,
		"column-gap":                 columnGap,
		"column-fill":                columnFill,
		"direction":                  direction,
		"display":                    display,
		"float":                      float,
		"font-familly":               fontFamilly,
		"font-kerning":               fontKerning,
		"font-language-override":     fontLanguageOverride,
		"font-variant-ligatures":     fontVariantLigatures,
		"font-variant-position":      fontVariantPosition,
		"font-variant-caps":          fontVariantCaps,
		"font-variant-numeric":       fontVariantNumeric,
		"font-feature-settings":      fontFeatureSettings,
		"font-variant-alternates":    fontVariantAlternates,
		"font-variant-east-asian":    fontVariantEastAsian,
		"font-style":                 fontStyle,
		"font-stretch":               fontStretch,
		"font-weight":                fontWeight,
		"image-resolution":           imageResolution,
		"letter-spacing":             letterSpacing,
		"word-spacing":               wordSpacing,
		"line-height":                lineHeight,
		"list-style-position":        listStylePosition,
		"list-style-type":            listStyleType,
	}
	validatorsError = map[string]validatorError{
		"background-image":  backgroundImage,
		"list-style-image":  imageUrl,
		"content":           content,
		"counter-increment": counterIncrement,
		"counter-reset":     counterReset,
		"font-size":         fontSize,
	}

	proprietary = Set{}
	unstable    = Set{
		"transform-origin": true,
	}

	// http://dev.w3.org/csswg/css3-values/#angles
	// 1<unit> is this many radians.
	ANGLETORADIANS = map[string]float32{
		"rad":  1,
		"turn": 2 * math.Pi,
		"deg":  math.Pi / 180,
		"grad": math.Pi / 200,
	}

	// http://dev.w3.org/csswg/css-values/#resolution
	RESOLUTIONTODPPX = map[string]float32{
		"dppx": 1,
		"dpi":  1 / LengthsToPixels[In],
		"dpcm": 1 / LengthsToPixels[Cm],
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

	couplesEastAsian = [][]string{
		{"jis78", "jis83", "jis90", "jis04", "simplified", "traditional"},
		{"full-width", "proportional-width"},
		{"ruby"},
	}

	allLigaturesValues = Set{}
	allNumericValues   = Set{}
	allEastAsianValues = Set{}
)

func init() {
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
	for _, couple := range couplesEastAsian {
		for _, cc := range couple {
			allEastAsianValues[cc] = true
		}
	}
}

type validator func(tokens []Token, baseUrl string) CssProperty
type validatorError func(tokens []Token, baseUrl string) (CssProperty, error)

type quote struct {
	open, insert bool
}

// If `token` is a keyword, return its name.
// Otherwise return empty string.
func getKeyword(token Token) string {
	if ident, ok := token.(IdentToken); ok {
		return ident.Value.Lower()
	}
	return ""
}

// If `tokens` is a 1-element list of keywords, return its name.
// Otherwise return empty string.
func getSingleKeyword(tokens []Token, _ string) CssProperty {
	if len(tokens) == 1 {
		return getKeyword(tokens[0])
	}
	return ""
}

// negative  = true, percentage = false
func getLength(_token Token, negative, percentage bool) Dimension {
	switch token := _token.(type) {
	case PercentageToken:
		if percentage && (negative || token.Value >= 0) {
			return Dimension{Value: token.Value, Unit: Percentage}
		}
	case DimensionToken:
		unit, isKnown := LENGTHUNITS[string(token.Unit)]
		if isKnown && (negative || token.Value >= 0) {
			return Dimension{Value: token.Value, Unit: unit}
		}
	case NumberToken:
		if token.Value == 0 {
			return Dimension{Unit: NoUnit}
		}
	}
	return Dimension{}
}

// Return the value in radians of an <angle> token, or None.
func getAngle(token Token) (float32, bool) {
	if dim, ok := token.(DimensionToken); ok {
		factor, in := ANGLETORADIANS[string(dim.Unit)]
		if in {
			return dim.Value * factor, true
		}
	}
	return 0, false
}

// Return the value := range dppx of a <resolution> token, || None.
func getResolution(token Token) (float32, bool) {
	if dim, ok := token.(DimensionToken); ok {
		factor, in := RESOLUTIONTODPPX[string(dim.Unit)]
		if in {
			return dim.Value * factor, true
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
func backgroundAttachment(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "scroll", "fixed", "local":
		return String(keyword)
	default:
		return nil
	}
}

//@validator("background-color")
//@validator("border-top-color")
//@validator("border-right-color")
//@validator("border-bottom-color")
//@validator("border-left-color")
//@validator("column-rule-color")
//@singleToken
func otherColors(tokens []Token, _ string) CssProperty {
	if len(tokens) == 1 {
		return parseColor(tokens[0])
	}
	return nil
}

//@validator()
//@singleToken
func outlineColor(tokens []Token, _ string) CssProperty {
	if len(tokens) == 1 {
		token := tokens[0]
		if getKeyword(token) == "invert" {
			return Color{Type: ColorCurrentColor}
		} else {
			return parseColor(token)
		}
	}
	return nil
}

//@validator()
//@singleKeyword
func borderCollapse(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "separate", "collapse":
		return String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleKeyword
// ``empty-cells`` property validation.
func emptyCells(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "show", "hide":
		return String(keyword)
	default:
		return nil
	}
}

//@validator("color")
//@singleToken
// ``*-color`` && ``color`` properties validation.
func color(tokens []Token, _ string) CssProperty {
	if len(tokens) == 1 {
		token := tokens[0]
		result := parseColor(token)
		if result.Type == ColorCurrentColor {
			return Color{Type: ColorInherit}
		} else {
			return result
		}
	}
	return nil
}

type LinearGradient struct {
	ColorStops []ColorStop
	Direction  directionType
	Repeating  bool
}

type RadialGradient struct {
	ColorStops []ColorStop
	Shape      string
	Size       Size
	Center     Center
	Repeating  bool
}

// Might be an existing image or a gradient
type Image interface {
	isImage()
	Copy() Image
}

func (l LinearGradient) isImage() {}
func (l RadialGradient) isImage() {}

// @validator("background-image", wantsBaseUrl=true)
// @commaSeparatedList
// @singleToken
func _backgroundImage(_token Token, baseUrl string) (Image, error) {
	token, ok := _token.(FunctionBlock)
	if !ok {
		return imageUrl([]Token{token}, baseUrl)
	}
	arguments := splitOnComma(removeWhitespace(token.Arguments))
	name := token.Name.Lower()
	var err error
	switch name {
	case "linear-gradient", "repeating-linear-gradient":
		direction, colorStops := parseLinearGradientParameters(arguments)
		if len(colorStops) > 0 {
			parsedColorsStop := make([]ColorStop, len(colorStops))
			for index, stop := range colorStops {
				parsedColorsStop[index], err = parseColorStop(stop)
				if err != nil {
					return nil, err
				}
			}
			return LinearGradient{
				Direction:  direction,
				Repeating:  name == "repeating-linear-gradient",
				ColorStops: parsedColorsStop,
			}, nil
		}
	case "radial-gradient", "repeating-radial-gradient":
		result := parseRadialGradientParameters(arguments)
		if (result == radialGradientParameters{}) {
			result.shape = "ellipse"
			result.size = "keyword", "farthest-corner"
			result.position = "left", fiftyPercent, "top", fiftyPercent
			result.colorStops = arguments
		}
		if len(result.colorStops) > 0 {
			parsedColorsStop := make([]ColorStop, len(result.colorStops))
			for index, stop := range result.colorStops {
				parsedColorsStop[index], err = parseColorStop(stop)
				if err != nil {
					return nil, err
				}
			}
			return RadialGradient{
				ColorStops: parsedColorsStop,
				Shape:      result.shape,
				Size:       result.size,
				Center:     result.position,
				Repeating:  name == "repeating-radial-gradient",
			}, nil
		}
	}
	return nil, nil
}

func backgroundImage(tokens []Token, baseUrl string) (CssProperty, error) {
	var out BackgroundImage
	for _, part := range splitOnComma(tokens) {
		part = removeWhitespace(part)
		if len(part) > 0 {
			result, err := _backgroundImage(part[0], baseUrl)
			if err != nil {
				return nil, err
			}
			if result == nil {
				return nil, nil
			}
			out = append(out, result)
		} else {
			return nil, nil
		}
	}
	return out, nil
}

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
	angle  float32
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

type ColorStop struct {
	Color    Color
	Position Dimension
}

func parseColorStop(tokens []Token) (ColorStop, error) {
	switch len(tokens) {
	case 1:
		color := parseColor(tokens[0])
		if !color.IsNone() {
			return ColorStop{Color: color}, nil
		}
	case 2:
		color := parseColor(tokens[0])
		position := getLength(tokens[1], true, true)
		if !color.IsNone() && !position.IsNone() {
			return ColorStop{Color: color, Position: position}, nil
		}
	default:
		return ColorStop{}, errors.New("Invalid or unsupported values for a known CSS property.")
	}
	return ColorStop{}, nil
}

func _imageUrl(token Token, baseUrl string) (Image, error) {
	if getKeyword(token) == "none" {
		return NoneImage{}, nil
	}
	if urlT, ok := token.(URLToken); ok {
		s, err := safeUrljoin(baseUrl, urlT.Value)
		return UrlImage(s), err
	}
	return nil, nil
}

//@validator("list-style-image", wantsBaseUrl=true)
//@singleToken
// ``*-image`` properties validation.
func imageUrl(tokens []Token, baseUrl string) (CssProperty, error) {
	if len(tokens) != 1 {
		return nil, nil
	}
	token := tokens[0]
	return _imageUrl(token, baseUrl)
}

var centerKeywordFakeToken = IdentToken{Value: "center"}

//@validator(unstable=true)
func transformOrigin(tokens []Token, _ string) CssProperty {
	// TODO: parse (and ignore) a third value for Z.
	return simple2dPosition(tokens)
}

//@validator()
//@commaSeparatedList
// ``background-position`` property validation.
//     See http://dev.w3.org/csswg/css3-background/#the-background-position
//
func _valideBackgroundPosition(tokens []Token) Center {
	center := simple2dPosition(tokens)
	if !center.IsNone() {
		return Center{
			OriginX: "left",
			OriginY: "top",
			Pos:     center,
		}
	}

	if len(tokens) == 4 {
		keyword1 := getKeyword(tokens[0])
		keyword2 := getKeyword(tokens[2])
		length1 := getLength(tokens[1], true, true)
		length2 := getLength(tokens[3], true, true)
		if !length1.IsNone() && !length2.IsNone() {
			if (keyword1 == "left" || keyword1 == "right") && (keyword2 == "top" || keyword2 == "bottom") {
				return Center{OriginX: keyword1,
					OriginY: keyword2,
					Pos:     Point{X: length1, Y: length2},
				}
			}
			if (keyword2 == "left" || keyword2 == "right") && (keyword1 == "top" || keyword1 == "bottom") {
				return Center{OriginX: keyword2,
					OriginY: keyword1,
					Pos:     Point{X: length2, Y: length1},
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
					return Center{OriginX: "left", OriginY: keyword, Pos: Point{X: fiftyPercent, Y: length}}
				case "left", "right":
					return Center{OriginX: keyword, OriginY: "top", Pos: Point{X: length, Y: fiftyPercent}}
				}
			case "top", "bottom":
				if keyword == "left" || keyword == "right" {
					return Center{OriginX: keyword, OriginY: otherKeyword, Pos: Point{X: length, Y: ZEROPERCENT}}
				}
			case "left", "right":
				if keyword == "top" || keyword == "bottom" {
					return Center{OriginX: otherKeyword, OriginY: keyword, Pos: Point{X: ZEROPERCENT, Y: length}}
				}
			}
		}
	}
	return Center{}
}

func valideBackgroundPosition(tokens []Token, _ string) CssProperty {
	var out BackgroundPosition
	for _, part := range splitOnComma(tokens) {
		result := _valideBackgroundPosition(removeWhitespace(part))
		if result.IsNone() {
			return nil
		}
		out = append(out, result)
	}
	return out
}

// Common syntax of background-position and transform-origin.
func simple2dPosition(tokens []Token) Point {
	if len(tokens) == 1 {
		tokens = []Token{tokens[0], centerKeywordFakeToken}
	} else if len(tokens) != 2 {
		return Point{}
	}

	token1, token2 := tokens[0], tokens[1]
	length1 := getLength(token1, true, true)
	length2 := getLength(token2, true, true)
	if !length1.IsNone() && !length2.IsNone() {
		return Point{X: length1, Y: length2}
	}
	keyword1, keyword2 := getKeyword(token1), getKeyword(token2)
	if !length1.IsNone() && (keyword2 == "top" || keyword2 == "center" || keyword2 == "bottom") {
		return Point{X: length1, Y: backgroundPositionsPercentages[keyword2]}
	} else if !length2.IsNone() && (keyword1 == "left" || keyword1 == "center" || keyword1 == "right") {
		return Point{X: backgroundPositionsPercentages[keyword1], Y: length2}
	} else if (keyword1 == "left" || keyword1 == "center" || keyword1 == "right") &&
		(keyword2 == "top" || keyword2 == "center" || keyword2 == "bottom") {
		return Point{X: backgroundPositionsPercentages[keyword1], Y: backgroundPositionsPercentages[keyword2]}
	} else if (keyword1 == "top" || keyword1 == "center" || keyword1 == "bottom") &&
		(keyword2 == "left" || keyword2 == "center" || keyword2 == "right") {
		// Swap tokens. They need to be in (horizontal, vertical) order.
		return Point{X: backgroundPositionsPercentages[keyword2], Y: backgroundPositionsPercentages[keyword1]}
	}
	return Point{}
}

//@validator()
//@commaSeparatedList
// ``background-repeat`` property validation.
func _backgroundRepeat(tokens []Token) [2]string {
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

func backgroundRepeat(tokens []Token, _ string) CssProperty {
	var out BackgroundRepeat
	for _, part := range splitOnComma(tokens) {
		result := _backgroundRepeat(removeWhitespace(part))
		if result == [2]string{} {
			return nil
		}
		out = append(out, result)
	}
	return out
}

//@validator()
//@commaSeparatedList
// Validation for ``background-size``.
func _backgroundSize(tokens []Token) Size {
	switch len(tokens) {
	case 1:
		token := tokens[0]
		keyword := getKeyword(token)
		switch keyword {
		case "contain", "cover":
			return Size{String: keyword}
		case "auto":
			return Size{Width: sToV("auto"), Height: sToV("auto")}
		}
		length := getLength(token, false, true)
		if !length.IsNone() {
			return Size{Width: Value{Dimension: length}, Height: sToV("auto")}
		}
	case 2:
		var out Size
		lengthW := getLength(tokens[0], false, true)
		lengthH := getLength(tokens[1], false, true)
		if !lengthW.IsNone() {
			out.Width = Value{Dimension: lengthW}
		} else if getKeyword(tokens[0]) == "auto" {
			out.Width = sToV("auto")
		} else {
			return Size{}
		}
		if !lengthH.IsNone() {
			out.Height = Value{Dimension: lengthH}
		} else if getKeyword(tokens[1]) == "auto" {
			out.Height = sToV("auto")
		} else {
			return Size{}
		}
		return out
	}
	return Size{}
}

func backgroundSize(tokens []Token, _ string) CssProperty {
	var out BackgroundSize
	for _, part := range splitOnComma(tokens) {
		result := _backgroundSize(removeWhitespace(part))
		if (result == Size{}) {
			return nil
		}
		out = append(out, result)
	}
	return out
}

//@validator("background-clip")
//@validator("background-origin")
//@commaSeparatedList
//@singleKeyword
// Validation for the ``<box>`` type used in ``background-clip``
//     and ``background-origin``.
func _box(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "border-box", "padding-box", "content-box":
		return keyword
	default:
		return ""
	}
}

func box(tokens []Token, _ string) CssProperty {
	var out Strings
	for _, part := range splitOnComma(tokens) {
		result := _box(removeWhitespace(part))
		if result == "" {
			return nil
		}
		out = append(out, result)
	}
	return out
}

func borderDims(tokens []Token, negative bool) CssProperty {
	lengths := make([]Dimension, len(tokens))
	allLengths := true
	for index, token := range tokens {
		lengths[index] = getLength(token, negative, false)
		allLengths = allLengths && !lengths[index].IsNone()
	}
	if allLengths {
		if len(lengths) == 1 {
			return WidthHeight{lengths[0], lengths[0]}
		} else if len(lengths) == 2 {
			return WidthHeight{lengths[0], lengths[1]}
		}
	}
	return nil
}

//@validator()
// Validator for the `border-spacing` property.
func borderSpacing(tokens []Token, _ string) CssProperty {
	return borderDims(tokens, true)
}

//@validator("border-top-right-radius")
//@validator("border-bottom-right-radius")
//@validator("border-bottom-left-radius")
//@validator("border-top-left-radius")
// Validator for the `border-*-radius` properties.
func borderCornerRadius(tokens []Token, _ string) CssProperty {
	return borderDims(tokens, false)
}

//@validator("border-top-style")
//@validator("border-right-style")
//@validator("border-left-style")
//@validator("border-bottom-style")
//@validator("column-rule-style")
//@singleKeyword
// ``border-*-style`` properties validation.
func borderStyle(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "none", "hidden", "dotted", "dashed", "double",
		"inset", "outset", "groove", "ridge", "solid":
		return String(keyword)
	default:
		return nil
	}
}

//@validator("break-before")
//@validator("break-after")
//@singleKeyword
// ``break-before`` && ``break-after`` properties validation.
func breakBeforeAfter(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	// "always" is defined as an alias to "page" := range multi-column
	// https://www.w3.org/TR/css3-multicol/#column-breaks
	switch keyword {
	case "auto", "avoid", "avoid-page", "page", "left", "right",
		"recto", "verso", "avoid-column", "column", "always":
		return Break(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleKeyword
// ``break-inside`` property validation.
func breakInside(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "avoid", "avoid-page", "avoid-column":
		return String(keyword)
	default:
		return nil
	}
}

//@validator(unstable=true)
//@singleToken
// ``page`` property validation.
func page(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if ident, ok := token.(IdentToken); ok {
		if ident.Value.Lower() == "auto" {
			return Page{String: "auto"}
		}
		return Page{String: string(ident.Value)}
	}
	return nil
}

//@validator("bleed-left")
//@validator("bleed-right")
//@validator("bleed-top")
//@validator("bleed-bottom")
//@singleToken
// ``bleed`` property validation.
func bleed(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	keyword := getKeyword(token)
	if keyword == "auto" {
		return Bleed{String: "auto"}
	} else {
		return Bleed{Dimension: getLength(token, true, false)}
	}
}

//@validator()
// ``marks`` property validation.
func marks(tokens []Token, _ string) CssProperty {
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
	return nil
}

//@validator("outline-style")
//@singleKeyword
// ``outline-style`` properties validation.
func outlineStyle(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "none", "dotted", "dashed", "double", "inset",
		"outset", "groove", "ridge", "solid":
		return String(keyword)
	default:
		return nil
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
func borderWidth(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	length := getLength(token, false, false)
	if !length.IsNone() {
		return BorderWidth{Dimension: length}
	}
	keyword := getKeyword(token)
	if keyword == "thin" || keyword == "medium" || keyword == "thick" {
		return BorderWidth{String: keyword}
	}
	return nil
}

// //@validator()
// //@singleToken
// ``column-width`` property validation.
func columnWidth(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	length := getLength(token, false, false)
	if !length.IsNone() {
		return ColumnWidth{Dimension: length}
	}
	keyword := getKeyword(token)
	if keyword == "auto" {
		return ColumnWidth{String: keyword}
	}
	return nil
}

// //@validator()
// //@singleKeyword
// ``column-span`` property validation.
func columnSpan(tokens []Token, _ string) CssProperty {
	// keyword := getSingleKeyword(tokens)
	// TODO: uncomment this when it is supported
	// return keyword := range ("all", "none")
	return nil
}

// //@validator()
// //@singleKeyword
// Validation for the ``box-sizing`` property from css3-ui
func boxSizing(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "padding-box", "border-box", "content-box":
		return String(keyword)
	default:
		return nil
	}
}

// //@validator()
// //@singleKeyword
// ``caption-side`` properties validation.
func captionSide(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "top", "bottom":
		return String(keyword)
	default:
		return nil
	}
}

// //@validator()
// //@singleKeyword
// ``clear`` property validation.
func clear(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "left", "right", "both", "none":
		return String(keyword)
	default:
		return nil
	}
}

// //@validator()
// //@singleToken
// Validation for the ``clip`` property.
func clip(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	name, args := parseFunction(token)
	if name != "" {
		if name == "rect" && len(args) == 4 {
			var values Lengths
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
		return Lengths{}
	}
	return nil
}

// //@validator(wantsBaseUrl=true)
// ``content`` property validation.
func content(tokens []Token, baseUrl string) (CssProperty, error) {
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
		if contentProperty.IsNone() {
			return nil, nil
		}
		out[index] = contentProperty
	}
	return Content{List: out}, nil
}

// helpers for validateContentToken type switches
func _isIdent(args []Token) (bool, IdentToken) {
	if len(args) == 1 {
		out, ok := args[0].(IdentToken)
		return ok, out
	}
	return false, IdentToken{}
}
func _isIdent2(args []Token) (bool, IdentToken, IdentToken) {
	if len(args) == 2 {
		out1, ok1 := args[0].(IdentToken)
		out2, ok2 := args[1].(IdentToken)
		return ok1 && ok2, out1, out2
	}
	return false, IdentToken{}, IdentToken{}
}
func _isIdentString(args []Token) (bool, IdentToken, StringToken) {
	if len(args) == 2 {
		out1, ok1 := args[0].(IdentToken)
		out2, ok2 := args[1].(StringToken)
		return ok1 && ok2, out1, out2
	}
	return false, IdentToken{}, StringToken{}
}
func _isIdentStringIdent(args []Token) (bool, IdentToken, StringToken, IdentToken) {
	if len(args) == 3 {
		out1, ok1 := args[0].(IdentToken)
		out2, ok2 := args[1].(StringToken)
		out3, ok3 := args[1].(IdentToken)
		return ok1 && ok2 && ok3, out1, out2, out3
	}
	return false, IdentToken{}, StringToken{}, IdentToken{}
}

// share attr, counter, counters cases
func _parseContentArgs(name string, args []Token) ContentProperty {
	switch name {
	case "attr":
		ok, ident := _isIdent(args)
		if ok {
			return ContentProperty{Type: ContentAttr, String: string(ident.Value)}
		}
	case "counter":
		if ok, ident := _isIdent(args); ok {
			return ContentProperty{Type: ContentCounter, Strings: []string{string(ident.Value), "decimal"}}
		}
		if ok, ident, ident2 := _isIdent2(args); ok {
			style := string(ident2.Value)
			_, isIn := counters.STYLES[style]
			if style == "none" || style == "decimal" || isIn {
				return ContentProperty{Type: ContentCounter, Strings: []string{string(ident.Value), style}}
			}
		}
	case "counters":
		if ok, ident, stri := _isIdentString(args); ok {
			return ContentProperty{Type: ContentCounter, Strings: []string{string(ident.Value), stri.Value, "decimal"}}
		}
		if ok, ident, stri, ident2 := _isIdentStringIdent(args); ok {
			style := string(ident2.Value)
			_, isIn := counters.STYLES[style]
			if style == "none" || style == "decimal" || isIn {
				return ContentProperty{Type: ContentCounter, Strings: []string{string(ident.Value), stri.Value, style}}
			}
		}
	}
	return ContentProperty{}
}

// Validation for a single token for the ``content`` property.
// Return (type, content) || zero for invalid tokens.
func validateContentToken(baseUrl string, token Token) (ContentProperty, error) {
	quoteType, isNotNone := ContentQuoteKeywords[getKeyword(token)]
	if isNotNone {
		return ContentProperty{Type: ContentQUOTE, Quote: quoteType}, nil
	}

	switch tt := token.(type) {
	case StringToken:
		return ContentProperty{Type: ContentSTRING, String: tt.Value}, nil
	case URLToken:
		url, err := safeUrljoin(baseUrl, tt.Value)
		if err != nil {
			return ContentProperty{}, err
		}
		return ContentProperty{Type: ContentURI, String: url}, nil
	}

	name, args := parseFunction(token)
	if name != "" {
		switch name {
		case "string":
			var stringArgs []string
			if ok, ident := _isIdent(args); ok {
				stringArgs = []string{string(ident.Value)}
			}
			if ok, ident, ident2 := _isIdent2(args); ok {
				args2 := ident2.Value.Lower()
				if args2 != "first" && args2 != "start" && args2 != "last" && args2 != "first-except" {
					return ContentProperty{}, fmt.Errorf("Invalid or unsupported CSS value : %s", args2)
				}
				stringArgs = []string{string(ident.Value), args2}
			}
			if stringArgs != nil { // thus one of the checks passed
				return ContentProperty{Type: ContentString, Strings: stringArgs}, nil
			}
		default:
			return _parseContentArgs(name, args), nil
		}
	}
	return ContentProperty{}, nil
}

// Return ``(name, args)`` if the given token is a function
//     with comma-separated arguments
func parseFunction(functionToken Token) (string, []Token) {
	if fun, ok := functionToken.(FunctionBlock); ok {
		content := removeWhitespace(fun.Arguments)
		if len(content) == 0 || len(content)%2 == 1 {
			for i := 1; i < len(content); i += 2 { // token in content[1::2]
				token := content[i]
				lit, isLit := token.(LiteralToken)
				if !isLit || lit.Value != "," {
					return nil, nil
				}
			}
			var args []Token
			for i := 0; i < len(content); i += 2 {
				args = append(args, content[i])
			}
			return fun.Name.Lower(), args
		}
	}
	return "", nil
}

// //@validator()
// ``counter-increment`` property validation.
func counterIncrement(tokens []Token, _ string) (CssProperty, error) {
	ci, err := counter(tokens, 1)
	if err != nil || ci == nil {
		return nil, err
	}
	return CounterIncrements{CI: ci}, nil
}

// //@validator()
// ``counter-reset`` property validation.
func counterReset(tokens []Token, _ string) (CssProperty, error) {
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
		ident, ok := token.(IdentToken)
		if !ok {
			return nil, nil // expected a keyword here
		}
		counterName := ident.Value
		if counterName == "none" || counterName == "initial" || counterName == "inherit" {
			return nil, fmt.Errorf("Invalid counter name: %s", counterName)
		}
		i += 1
		if i < len(tokens) {
			token = tokens[i]
			if number, ok := token.(NumberToken); ok {
				if number.IsInteger {
					// Found an integer. Use it and get the next token
					integer = number.IntValue()
					i += 1
				}
			} else {
				// Not an integer. Might be the next counter name.
				// Keep `token` for the next loop iteration.
				integer = defaultInteger
			}
		}
		results = append(results, IntString{Name: string(counterName), Value: integer})
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
func lengthPercOrAuto(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	length := getLength(token, true, true)
	if !length.IsNone() {
		return Length{Dimension: length}
	}
	if getKeyword(token) == "auto" {
		return Length{String: "auto"}
	}
	return nil
}

// //@validator("height")
// //@validator("width")
// //@singleToken
// Validation for the ``width`` && ``height`` properties.
func widthHeight(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	length := getLength(token, false, true)
	if !length.IsNone() {
		return Length{Dimension: length}
	}
	if getKeyword(token) == "auto" {
		return Length{String: "auto"}
	}
	return nil
}

// //@validator()
// //@singleToken
// Validation for the ``column-gap`` property.
func columnGap(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	length := getLength(token, false, false)
	if !length.IsNone() {
		return ColumnGap{Dimension: length}
	}
	if getKeyword(token) == "normal" {
		return ColumnGap{String: "normal"}
	}
	return nil
}

//@validator()
//@singleKeyword
// ``column-fill`` property validation.
func columnFill(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "balance":
		return String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleKeyword
// ``direction`` property validation.
func direction(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "ltr", "rtl":
		return String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleKeyword
// ``display`` property validation.
func display(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "inline", "block", "inline-block", "list-item", "none",
		"table", "inline-table", "table-caption",
		"table-row-group", "table-header-group", "table-footer-group",
		"table-row", "table-column-group", "table-column", "table-cell":
		return Display(keyword)
	default:
		return nil
	}
}

//@validator("float")
//@singleKeyword
// ``float`` property validation.
func float(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "left", "right", "none":
		return Floating(keyword)
	default:
		return nil
	}
}

func _fontFamily(tokens []Token) string {
	if len(tokens) == 1 {
		if tt, ok := tokens[0].(StringToken); ok {
			return tt.Value
		}
	} else if len(tokens) > 0 {
		var values []string
		for _, token := range tokens {
			if tt, ok := token.(IdentToken); ok {
				values = append(values, string(tt.Value))
			} else {
				return ""
			}
		}
		return strings.Join(values, " ")
	}
	return ""
}

// //@validator()
// //@commaSeparatedList
// ``font-family`` property validation.
func fontFamilly(tokens []Token, _ string) CssProperty {
	var out Strings
	for _, part := range splitOnComma(tokens) {
		result := _fontFamily(removeWhitespace(part))
		if result == "" {
			return nil
		}
		out = append(out, result)
	}
	return out
}

// //@validator()
// //@singleKeyword
func fontKerning(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "normal", "none":
		return String(keyword)
	default:
		return nil
	}
}

// //@validator()
// //@singleToken
func fontLanguageOverride(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	keyword := getKeyword(token)
	if keyword == "normal" {
		return String(keyword)
	}
	if tt, ok := token.(StringToken); ok {
		return String(tt.Value)
	}
	return nil
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
		ident, isIdent := token.(IdentToken)
		if !isIdent {
			return FontVariant{}
		}
		identValue := string(ident.Value)
		if all[identValue] {
			var concurrentValues []string
			for _, couple := range couples {
				if isInValues(identValue, couple) {
					concurrentValues = couple
					break
				}
			}
			for _, value := range concurrentValues {
				if isInValues(value, values) {
					return FontVariant{}
				}
			}
			values = append(values, identValue)
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
func fontVariantLigatures(tokens []Token, _ string) CssProperty {
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
func fontVariantPosition(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "normal", "sub", "super":
		return String(keyword)
	default:
		return nil
	}
}

// //@validator()
// //@singleKeyword
func fontVariantCaps(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "normal", "small-caps", "all-small-caps", "petite-caps",
		"all-petite-caps", "unicase", "titling-caps":
		return String(keyword)
	default:
		return nil
	}
}

//@validator()
func fontVariantNumeric(tokens []Token, _ string) CssProperty {
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
func fontFeatureSettings(tokens []Token, _ string) CssProperty {
	if len(tokens) == 1 && getKeyword(tokens[0]) == "normal" {
		return FontFeatureSettings{Normal: true}
	}

	fontFeatureSettingsList := func(tokens []Token) IntString {
		var token Token
		feature, value := "", 0

		if len(tokens) == 2 {
			tokens, token = tokens[0:1], tokens[1]
			switch tt := token.(type) {
			case IdentToken:
				if tt.Value == "on" {
					value = 1
				} else {
					value = 0
				}
			case NumberToken:
				if tt.IsInteger && tt.IntValue() >= 0 {
					value = tt.IntValue()
				}
			}
		} else if len(tokens) == 1 {
			value = 1
		}

		if len(tokens) == 1 {
			token = tokens[0]
			tt, ok := token.(StringToken)
			if ok && len(tt.Value) == 4 {
				ok := true
				for _, letter := range tt.Value {
					if !(0x20 <= letter && letter <= 0x7f) {
						ok = false
						break
					}
				}
				if ok {
					feature = tt.Value
				}
			}
		}

		if feature != "" && value != 0 {
			return IntString{Name: feature, Value: value}
		}
		return IntString{}
	}

	var out FontFeatureSettings
	for _, part := range splitOnComma(tokens) {
		result := fontFeatureSettingsList(removeWhitespace(part))
		if (result == IntString{}) {
			return nil
		}
		out.TagValues = append(out.TagValues, result)
	}
	return out
}

//@validator()
//@singleKeyword
func fontVariantAlternates(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	// TODO: support other values
	// See https://www.w3.org/TR/css-fonts-3/#font-variant-caps-prop
	switch keyword {
	case "normal", "historical-forms":
		return String(keyword)
	default:
		return nil
	}
}

// //@validator()
func fontVariantEastAsian(tokens []Token, _ string) CssProperty {
	if len(tokens) == 1 {
		keyword := getKeyword(tokens[0])
		if keyword == "normal" {
			return FontVariant{String: keyword}
		}
	}
	return parseFontVariant(tokens, allEastAsianValues, couplesEastAsian)
}

//@validator()
//@singleToken
// ``font-size`` property validation.
func fontSize(tokens []Token, _ string) (CssProperty, error) {
	if len(tokens) != 1 {
		return nil, nil
	}
	token := tokens[0]
	length := getLength(token, false, true)
	if !length.IsNone() {
		return FontSize{Dimension: length}, nil
	}
	fontSizeKeyword := getKeyword(token)
	if fontSizeKeyword == "smaller" || fontSizeKeyword == "larger" {
		return nil, fmt.Errorf("value %s not supported yet", fontSizeKeyword)
	}
	if _, isIn := FontSizeKeywords[fontSizeKeyword]; isIn {
		// || keyword := range ("smaller", "larger")
		return FontSize{String: fontSizeKeyword}, nil
	}
	return nil, nil
}

//@validator()
//@singleKeyword
// ``font-style`` property validation.
func fontStyle(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "normal", "italic", "oblique":
		return String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleKeyword
// Validation for the ``font-stretch`` property.
func fontStretch(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "ultra-condensed", "extra-condensed", "condensed", "semi-condensed",
		"normal", "semi-expanded", "expanded", "extra-expanded", "ultra-expanded":
		return String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleToken
// ``font-weight`` property validation.
func fontWeight(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	keyword := getKeyword(token)
	if keyword == "normal" || keyword == "bold" || keyword == "bolder" || keyword == "lighter" {
		return FontWeight{Name: keyword}
	}
	if number, ok := token.(NumberToken); ok {
		intValue := number.IntValue()
		if number.IsInteger && (intValue == 100 || intValue == 200 || intValue == 300 || intValue == 400 || intValue == 500 || intValue == 600 || intValue == 700 || intValue == 800 || intValue == 900) {
			return FontWeight{Value: intValue}
		}
	}
	return nil
}

//@validator(unstable=true)
//@singleToken
func imageResolution(tokens []Token, _ string) CssProperty {
	// TODO: support "snap" && "from-image"
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	value, ok := getResolution(token)
	if !ok {
		return nil
	}
	return fToV(value)
}

//@validator("letter-spacing")
//@validator("word-spacing")
//@singleToken
// Validation for ``letter-spacing`` && ``word-spacing``.
func _spacing(tokens []Token) Value {
	if len(tokens) != 1 {
		return Value{}
	}
	token := tokens[0]
	if getKeyword(token) == "normal" {
		return Value{String: "normal"}
	}
	length := getLength(token, true, false)
	if !length.IsNone() {
		return Value{Dimension: length}
	}
	return Value{}
}

func letterSpacing(tokens []Token, _ string) CssProperty {
	v := _spacing(tokens)
	if v.IsNone() {
		return nil
	}
	return PixelLength(v)
}
func wordSpacing(tokens []Token, _ string) CssProperty {
	v := _spacing(tokens)
	if v.IsNone() {
		return nil
	}
	return WordSpacing(v)
}

//@validator()
//@singleToken
// ``line-height`` property validation.
func lineHeight(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if getKeyword(token) == "normal" {
		return LineHeight{String: "normal"}
	}

	switch tt := token.(type) {
	case NumberToken:
		if tt.Value >= 0 {
			return LineHeight{Dimension: Dimension{Value: tt.Value, Unit: NoUnit}}
		}
	case PercentageToken:
		if tt.Value >= 0 {
			return LineHeight{Dimension: Dimension{Value: tt.Value, Unit: Percentage}}
		}
	case DimensionToken:
		if tt.Value >= 0 {
			return LineHeight{Dimension: getLength(token, true, false)}
		}
	}
	return nil
}

//@validator()
//@singleKeyword
// ``list-style-position`` property validation.
func listStylePosition(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "inside", "outside":
		return String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleKeyword
// ``list-style-type`` property validation.
func listStyleType(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	_, inStyles := counters.STYLES[keyword]
	if keyword == "none" || keyword == "decimal" || inStyles {
		return String(keyword)
	} else {
		return nil
	}
}

//@validator("padding-top")
//@validator("padding-right")
//@validator("padding-bottom")
//@validator("padding-left")
//@validator("min-width")
//@validator("min-height")
//@singleToken
// ``padding-*`` properties validation.
func lengthOrPrecentage(token Token) Dimension {
	return getLength(token, false, true)
}

//@validator("max-width")
//@validator("max-height")
//@singleToken
// Validation for max-width && max-height
func maxWidthHeight(token Token) Dimension {
	length := getLength(token, false, true)
	if !length.IsNone() {
		return length
	}
	if getKeyword(token) == "none" {
		return Dimension{Value: float32(math.Inf(1.)), Unit: Px}
	}
	return Dimension{}
}

//@validator()
//@singleToken
// Validation for the ``opacity`` property.
func opacity(token Token) (float32, bool) {
	if number, ok := token.(NumberToken); ok {
		return min(1, max(0, number.Value)), true
	}
	return 0, false
}

//@validator()
//@singleToken
// Validation for the ``z-index`` property.
func zIndex(token Token) IntString {
	if getKeyword(token) == "auto" {
		return IntString{Name: "auto"}
	}
	if number, ok := token.(NumberToken); ok {
		if number.IsInteger {
			return IntString{Value: number.IntValue()}
		}
	}
	return IntString{}
}

//@validator("orphans")
//@validator("widows")
//@singleToken
// Validation for the ``orphans`` && ``widows`` properties.
func orphansWidows(token Token) (int, bool) {
	if number, ok := token.(NumberToken); ok {
		value := number.IntValue()
		if number.IsInteger && value >= 1 {
			return value, true
		}
	}
	return 0, false
}

//@validator()
//@singleToken
// Validation for the ``column-count`` property.
func columnCount(token Token) IntString {
	if number, ok := token.(NumberToken); ok {
		value := number.IntValue()
		if number.IsInteger && value >= 1 {
			return IntString{Value: value}
		}
	}
	if getKeyword(token) == "auto" {
		return IntString{Name: "auto"}
	}
	return IntString{}
}

//@validator()
//@singleKeyword
// Validation for the ``overflow`` property.
func overflow(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "visible", "hidden", "scroll":
		return String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleKeyword
// ``position`` property validation.
func position(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "static", "relative", "absolute", "fixed":
		return String(keyword)
	default:
		return nil
	}
}

//@validator()
// ``quotes`` property validation.
func quotes(tokens []Token) Quotes {
	var opens, closes []string
	if len(tokens) > 0 && len(tokens)%2 == 0 {
		// Separate open && close quotes.
		// eg.  ("«", "»", "“", "”")  -> (("«", "“"), ("»", "”"))
		for i := 0; i < len(tokens); i += 2 {
			open, ok1 := tokens[i].(StringToken)
			close_, ok2 := tokens[i+1].(StringToken)
			if ok1 && ok2 {
				opens = append(opens, open.Value)
				closes = append(closes, close_.Value)
			} else {
				return Quotes{}
			}
		}
	}
	return Quotes{Open: opens, Close: closes}
}

//@validator()
//@singleKeyword
// Validation for the ``table-layout`` property
func tableLayout(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "fixed", "auto":
		return String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleKeyword
// ``text-align`` property validation.
func textAlign(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "left", "right", "center", "justify":
		return String(keyword)
	default:
		return nil
	}
}

//@validator()
// ``text-decoration`` property validation.
func textDecoration(tokens []Token) TextDecoration {
	uniqKeywords := Set{}
	valid := true
	for _, token := range tokens {
		keyword := getKeyword(token)
		uniqKeywords[keyword] = true
		if !(keyword == "underline" || keyword == "overline" || keyword == "line-through" || keyword == "blink") {
			valid = false
		}
	}
	if len(tokens) == 1 && uniqKeywords["none"] {
		return TextDecoration{None: true}
	}
	if valid && len(uniqKeywords) == len(tokens) {
		// No duplicate
		// blink is accepted but ignored
		// "Conforming user agents may simply not blink the text."
		delete(uniqKeywords, "blink")
		return TextDecoration{Decorations: uniqKeywords}
	}
	return TextDecoration{}
}

//@validator()
//@singleToken
// ``text-indent`` property validation.
func textIndent(token Token) Dimension {
	return getLength(token, true, true)
}

//@validator()
//@singleKeyword
// ``text-align`` property validation.
func textTransform(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "none", "uppercase", "lowercase", "capitalize", "full-width":
		return String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleToken
// Validation for the ``vertical-align`` property
func verticalAlign(token Token) Value {
	length := getLength(token, true, true)
	if !length.IsNone() {
		return Value{Dimension: length}
	}
	keyword := getKeyword(token)
	if keyword == "baseline" || keyword == "middle" || keyword == "sub" || keyword == "super" || keyword == "text-top" || keyword == "text-bottom" || keyword == "top" || keyword == "bottom" {
		return Value{String: keyword}
	}
	return Value{}
}

//@validator()
//@singleKeyword
// ``white-space`` property validation.
func visibility(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "visible", "hidden", "collapse":
		return String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleKeyword
// ``white-space`` property validation.
func whiteSpace(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "normal", "pre", "nowrap", "pre-wrap", "pre-line":
		return String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleKeyword
// ``overflow-wrap`` property validation.
func overflowWrap(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "normal", "break-word":
		return String(keyword)
	default:
		return nil
	}
}

//@validator(unstable=true)
//@singleKeyword
// Validation for ``image-rendering``.
func imageRendering(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "crisp-edges", "pixelated":
		return String(keyword)
	default:
		return nil
	}
}

//@validator(unstable=true)
// ``size`` property validation.
// See http://www.w3.org/TR/css3-page/#page-size-prop
func size(tokens []Token) WidthHeight {
	var (
		lengths        []Dimension
		keywords       []string
		lengthsNotNone bool = true
	)
	for _, token := range tokens {
		length, keyword := getLength(token, false, false), getKeyword(token)
		lengthsNotNone = lengthsNotNone && !length.IsNone()
		lengths = append(lengths, length)
		keywords = append(keywords, keyword)
	}

	if lengthsNotNone {
		if len(lengths) == 1 {
			return WidthHeight{lengths[0], lengths[0]}
		} else if len(lengths) == 2 {
			return WidthHeight{lengths[0], lengths[1]}
		}
	}

	if len(keywords) == 1 {
		keyword := keywords[0]
		if psize, in := pageSizes[keyword]; in {
			return psize
		} else if keyword == "auto" || keyword == "portrait" {
			return initialWidthHeight
		} else if keyword == "landscape" {
			return WidthHeight{initialWidthHeight[1], initialWidthHeight[0]}
		}
	}

	if len(keywords) == 2 {
		var orientation, pageSize string
		if keywords[0] == "portrait" || keywords[0] == "landscape" {
			orientation, pageSize = keywords[0], keywords[1]
		} else if keywords[1] == "portrait" || keywords[1] == "landscape" {
			pageSize, orientation = keywords[0], keywords[1]
		}
		if widthHeight, in := pageSizes[pageSize]; in {
			if orientation == "portrait" {
				return widthHeight
			} else {
				return WidthHeight{widthHeight[1], widthHeight[0]}
			}
		}
	}
	return WidthHeight{}
}

//@validator(proprietary=true)
//@singleToken
// Validation for ``anchor``.
func anchor(token Token) Anchor {
	if getKeyword(token) == "none" {
		return Anchor{Type: "none"}
	}
	name, args := parseFunction(token)
	if name != "" {
		if len(args) == 1 {
			if ident, ok := args[0].(IdentToken); ok {
				return Anchor{Type: name, Attr: string(ident.Value)}
			}
		}
	}
	return Anchor{}
}

//@validator(proprietary=true, wantsBaseUrl=true)
//@singleToken
// Validation for ``link``.
func link(token Token, baseUrl string) (Link, error) {
	if getKeyword(token) == "none" {
		return Link{Type: "none"}, nil
	} else if urlToken, isUrl := token.(URLToken); isUrl {
		if strings.HasPrefix(urlToken.Value, "#") {
			unescaped, err := url.PathUnescape(urlToken.Value[1:])
			if err != nil {
				return Link{}, fmt.Errorf("Invalid internal url : %s", err)
			}
			return Link{Type: "internal", Attr: unescaped}, nil
		} else {
			safeurl, err := safeUrljoin(baseUrl, urlToken.Value)
			if err != nil {
				return Link{}, fmt.Errorf("Invalid external url : %s", err)
			}
			return Link{Type: "external", Attr: safeurl}, nil
		}
	}
	name, args := parseFunction(token)
	if name != "" {
		if len(args) == 1 {
			if ident, ok := args[0].(IdentToken); ok {
				return Link{Type: name, Attr: string(ident.Value)}, nil
			}
		}
	}
	return Link{}, nil
}

//@validator()
//@singleToken
// Validation for ``tab-size``.
// See https://www.w3.org/TR/css-text-3/#tab-size
func tabSize(token Token) Dimension {
	if number, ok := token.(NumberToken); ok {
		if number.IsInteger && number.Value >= 0 {
			return Dimension{Value: number.Value}
		}
	}
	return getLength(token, false, false)
}

//@validator(unstable=true)
//@singleToken
// Validation for ``hyphens``.
func hyphens(token Token) CssProperty {
	keyword := getKeyword(token)
	switch keyword {
	case "none", "manual", "auto":
		return String(keyword)
	default:
		return nil
	}
}

//@validator(unstable=true)
//@singleToken
// Validation for ``hyphenate-character``.
func hyphenateCharacter(token Token) string {
	keyword := getKeyword(token)
	if keyword == "auto" {
		return "‐"
	} else if str, ok := token.(StringToken); ok {
		return str.Value
	}
	return ""
}

//@validator(unstable=true)
//@singleToken
// Validation for ``hyphenate-limit-zone``.
func hyphenateLimitZone(token Token) Dimension {
	return getLength(token, false, true)
}

//@validator(unstable=true)
// Validation for ``hyphenate-limit-chars``.
func hyphenateLimitChars(tokens []Token) [3]int {
	switch len(tokens) {
	case 1:
		token := tokens[0]
		keyword := getKeyword(token)
		if keyword == "auto" {
			return [3]int{5, 2, 2}
		} else if number, ok := token.(NumberToken); ok && number.IsInteger {
			return [3]int{number.IntValue(), 2, 2}

		}
	case 2:
		total, left := tokens[0], tokens[1]
		totalKeyword := getKeyword(total)
		leftKeyword := getKeyword(left)
		if totalNumber, ok := total.(NumberToken); ok && totalNumber.IsInteger {
			if leftNumber, ok := left.(NumberToken); ok && leftNumber.IsInteger {
				return [3]int{totalNumber.IntValue(), leftNumber.IntValue(), leftNumber.IntValue()}
			} else if leftKeyword == "auto" {
				return [3]int{totalNumber.IntValue(), 2, 2}
			}
		} else if totalKeyword == "auto" {
			if leftNumber, ok := left.(NumberToken); ok && leftNumber.IsInteger {
				return [3]int{5, leftNumber.IntValue(), leftNumber.IntValue()}
			} else if leftKeyword == "auto" {
				return [3]int{5, 2, 2}
			}
		}
	case 3:
		total, left, right := tokens[0], tokens[1], tokens[2]
		totalNumber, okT := total.(NumberToken)
		leftNumber, okL := left.(NumberToken)
		rightNumber, okR := right.(NumberToken)
		if ((okT && totalNumber.IsInteger) || getKeyword(total) == "auto") &&
			((okL && leftNumber.IsInteger) || getKeyword(left) == "auto") &&
			((okR && rightNumber.IsInteger) || getKeyword(right) == "auto") {
			totalInt := 5
			if okT {
				totalInt = totalNumber.IntValue()
			}
			leftInt := 2
			if okL {
				leftInt = leftNumber.IntValue()
			}
			rightInt := 2
			if okR {
				rightInt = rightNumber.IntValue()
			}
			return [3]int{totalInt, leftInt, rightInt}
		}
	}
	return [3]int{}
}

//@validator(proprietary=true)
//@singleToken
// Validation for ``lang``.
func lang(token Token) Lang {
	if getKeyword(token) == "none" {
		return Lang{Type: "none"}
	}
	name, args := parseFunction(token)
	if name != "" {
		if len(args) == 1 {
			if ident, ok := args[0].(IdentToken); ok {
				return Lang{Type: name, Attr: string(ident.Value)}
			}
		}

	} else if str, ok := token.(StringToken); ok {
		return Lang{Type: "string", Attr: str.Value}
	}
	return Lang{}
}

//@validator(unstable=true)
// Validation for ``bookmark-label``.
func bookmarkLabel(tokens []Token) []ContentProperty {
	parsedTokens := make([]ContentProperty, len(tokens))
	for index, v := range tokens {
		parsedTokens[index] = validateContentListToken(v)
		if parsedTokens[index].IsNil() {
			parsedTokens = nil
			break
		}
	}
	return parsedTokens
}

//@validator(unstable=true)
//@singleToken
// Validation for ``bookmark-level``.
func bookmarkLevel(token Token) IntString {
	if number, ok := token.(NumberToken); ok && number.IsInteger && number.IntValue() >= 1 {
		return IntString{Value: number.IntValue()}
	} else if getKeyword(token) == "none" {
		return IntString{Name: "none"}
	}
	return IntString{}
}

//@validator(unstable=true)
//@commaSeparatedList
// Validation for ``string-set``.
func stringSet(tokens []Token) Content {
	if len(tokens) >= 2 {
		varName := getKeyword(tokens[0])
		parsedTokens := make([]ContentProperty, len(tokens[1:]))
		isNotNone := true
		for index, v := range tokens {
			parsedTokens[index] = validateContentListToken(v)
			if parsedTokens[index].IsNil() {
				isNotNone = false
				break
			}
		}
		if isNotNone {
			return Content{String: varName, List: parsedTokens}
		}
	} else if len(tokens) > 0 {
		switch tt := tokens[0].(type) {
		case StringToken:
			if tt.Value == "none" {
				return Content{String: "none"}
			}
		case IdentToken:
			if tt.Value == "none" {
				return Content{String: "none"}
			}
		}
	}
	return Content{}
}

// Validation for a single token of <content-list> used in GCPM.
//     Return (type, content) || false for invalid tokens.
//
func validateContentListToken(token Token) ContentProperty {
	if tt, ok := token.(StringToken); ok {
		return ContentProperty{Type: ContentSTRING, String: tt.Value}
	}

	name, args := parseFunction(token)
	if name != "" {
		switch name {
		case "content":
			// if prototype := range (("content", ()), ("content", ("ident",))) {
			if len(args) == 0 {
				return ContentProperty{Type: ContentContent, String: "text"}
			} else if len(args) == 1 {
				if ident, ok := args[0].(IdentToken); ok && (ident.Value == "text" || ident.Value == "after" || ident.Value == "before" || ident.Value == "first-letter") {
					return ContentProperty{Type: ContentContent, String: string(ident.Value)}
				}
			}
		default:
			return _parseContentArgs(name, args)
		}
	}
	return ContentProperty{}
}

//@validator(unstable=true)
func transform(tokens []Token) ([]Transform, error) {
	if getSingleKeyword(tokens) == "none" {
		return nil, nil
	}
	out := make([]Transform, len(tokens))
	var err error
	for index, v := range tokens {
		out[index], err = transformFunction(v)
		if err != nil {
			return nil, err
		}
	}
	return out, nil

}

func transformFunction(token Token) (Transform, error) {
	name, args := parseFunction(token)
	if name == "" {
		return Transform{}, errors.New("Invalid or unsupported values for a known CSS property.")
	}

	lengths, values := make([]Dimension, len(args)), make([]Dimension, len(args))
	isAllNumber, isAllLengths := true, true
	for index, a := range args {
		lengths[index] = getLength(token, true, true)
		isAllLengths = isAllLengths && !lengths[index].IsNone()
		if aNumber, ok := a.(NumberToken); ok {
			values[index] = toDim(aNumber.Value)
		} else {
			isAllNumber = false
		}
	}

	switch len(args) {
	case 1:
		angle, notNone := getAngle(args[0])
		length := getLength(args[0], true, true)
		switch name {
		case "rotate", "skewx", "skewy":
			if notNone && angle != 0 {
				return Transform{Function: name, Args: []Dimension{toDim(angle)}}, nil
			}
		case "translatex", "translate":
			if !length.IsNone() {
				return Transform{Function: "translate", Args: []Dimension{length, ZeroPixels}}, nil
			}
		case "translatey":
			if !length.IsNone() {
				return Transform{Function: "translate", Args: []Dimension{ZeroPixels, length}}, nil
			}
		case "scalex":
			if number, ok := args[0].(NumberToken); ok {
				return Transform{Function: "scale", Args: []Dimension{toDim(number.Value), toDim(1.)}}, nil
			}
		case "scaley":
			if number, ok := args[0].(NumberToken); ok {
				return Transform{Function: "scale", Args: []Dimension{toDim(1.), toDim(number.Value)}}, nil
			}
		case "scale":
			if number, ok := args[0].(NumberToken); ok {
				return Transform{Function: "scale", Args: []Dimension{toDim(number.Value), toDim(number.Value)}}, nil
			}
		}
	case 2:
		if name == "scale" && isAllNumber {
			return Transform{Function: name, Args: values}, nil
		}

		if name == "translate" && isAllLengths {
			return Transform{Function: name, Args: lengths}, nil
		}
	case 6:
		if name == "matrix" && isAllNumber {
			return Transform{Function: name, Args: values}, nil
		}
	}
	return Transform{}, errors.New("Invalid or unsupported values for a known CSS property.")
}

// Expanders

// Let"s be consistent, always use ``name`` as an argument even
// when it is useless.

// // Decorator adding a function to the ``EXPANDERS``.
// func expander(propertyName) {
//     def expanderDecorator(function) {
//         """Add ``function`` to the ``EXPANDERS``."""
//         assert propertyName not := range EXPANDERS, propertyName
//         EXPANDERS[propertyName] = function
//         return function
//     } return expanderDecorator
// }

//@expander("border-color")
//@expander("border-style")
//@expander("border-width")
//@expander("margin")
//@expander("padding")
//@expander("bleed")
// Expand properties setting a token for the four sides of a box.
func expandFourSides(baseUrl, name string, tokens []Token) (out [4][2]string, err error) {
	// Make sure we have 4 tokens
	if len(tokens) == 1 {
		tokens = []Token{tokens[0], tokens[0], tokens[0], tokens[0]}
	} else if len(tokens) == 2 {
		tokens = []Token{tokens[0], tokens[1], tokens[0], tokens[1]} // (bottom, left) defaults to (top, right)
	} else if len(tokens) == 3 {
		tokens = append(tokens, tokens[1]) // left defaults to right
	} else if len(tokens) != 4 {
		return out, fmt.Errorf("Expected 1 to 4 token components got %d", len(tokens))
	}
	var newName string
	for index, suffix := range [4]string{"-top", "-right", "-bottom", "-left"} {
		token := tokens[index]
		i := strings.LastIndex(name, "-")
		if i == -1 {
			newName = name + suffix
		} else {
			// eg. border-color becomes border-*-color, not border-color-*
			newName = name[:i] + suffix + name[i:]
		}

		out[index][0], out[index][1], err = validateNonShorthand(baseUrl, newName, []Token{token}, true)
		if err != nil {
			return out, err
		}
	}
	return out, nil
}

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

var (
	noneFakeToken   = IdentToken{Value: "none"}
	normalFakeToken = IdentToken{Value: "normal"}
)

// //@expander("font-variant")
// //@genericExpander("-alternates", "-caps", "-east-asian", "-ligatures",
//                   "-numeric", "-position")
// // Expand the ``font-variant`` shorthand property.
// //     https://www.w3.org/TR/css-fonts-3/#font-variant-prop
// //
// func fontVariant(name, tokens []Token) {
//     return expandFontVariant(tokens)
// }

type fontVariantToken struct {
	feature string
	tokens  []Token
}

func expandFontVariant(tokens []Token) (out []fontVariantToken, err error) {
	keyword := getSingleKeyword(tokens)
	if keyword == "normal" || keyword == "none" {
		out = make([]fontVariantToken, 6)
		for index, suffix := range [5]string{"-alternates", "-caps", "-east-asian", "-numeric",
			"-position"} {
			out[index] = fontVariantToken{feature: suffix, tokens: []Token{normalFakeToken}}
		}
		token := noneFakeToken
		if keyword == "normal" {
			token = normalFakeToken
		}
		out[6] = fontVariantToken{feature: "-ligatures", tokens: []Token{token}}
	} else {
		features := map[string][]Token{
			"alternates": nil,
			"caps":       nil,
			"east-asian": nil,
			"ligatures":  nil,
			"numeric":    nil,
			"position":   nil,
		}
		for _, token := range tokens {
			keyword := getKeyword(token)
			if keyword == "normal" {
				// We don"t allow "normal", only the specific values
				return nil, errors.New("normal not allowed")
			}
			found := false
			for feature := range features {
				if fontVariantMapper[feature]([]Token{token}) != nil {
					features[feature] = append(features[feature], token)
					found = true
					break
				}
			}
			if !found {
				return nil, errors.New("font variant not supported")
			}
		}
		var out []fontVariantToken
		for feature, tokens := range features {
			if len(tokens) > 0 {
				out = append(out, fontVariantToken{feature: fmt.Sprintf("-%s", feature), tokens: tokens})
			}
		}
	}
	return out, nil
}

var fontVariantMapper = map[string]func(tokens []Token, _ string) CssProperty{
	"alternates": fontVariantAlternates,
	"caps":       fontVariantCaps,
	"east-asian": fontVariantEastAsian,
	"ligatures":  fontVariantLigatures,
	"numeric":    fontVariantNumeric,
	"position":   fontVariantPosition,
}

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

// Default validator for non-shorthand properties.
// required = false
func validateNonShorthand(baseUrl, name string, tokens []Token, required bool) (string, string, error) {
	if !required && !knownProperties[name] {
		hyphensName := strings.ReplaceAll(name, "", "-")
		if knownProperties[hyphensName] {
			return nil, "", fmt.Errorf("did you mean %s?", hyphensName)
		} else {
			return nil, "", errors.New("unknown property")
		}
	}

	if _, isIn := validators[name]; !required && !isIn {
		fmt.Errorf("property %s not supported yet", name)
	}
	var value string
	keyword := getSingleKeyword(tokens)
	if keyword == "initial" || keyword == "inherit" {
		value = keyword
	} else {
		function := validators[name]
		value = function(tokens, baseUrl)

		if value == "" {
			return nil, "", errors.New("invalid property (empty function return)")
		}
	}
	return name, value, nil
}

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
		if token.Type() != TypeWhitespaceToken && token.Type() != TypeComment {
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
		litteral, ok := token.(LiteralToken)
		if ok && litteral.Value == "," {
			parts = append(parts, thisPart)
			thisPart = nil
		} else {
			thisPart = append(thisPart, token)
		}
	}
	parts = append(parts, thisPart)
	return parts
}
