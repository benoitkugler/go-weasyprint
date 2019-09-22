package validation

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strings"

	. "github.com/benoitkugler/go-weasyprint/css"
	"github.com/benoitkugler/go-weasyprint/css/parser"
	"github.com/benoitkugler/go-weasyprint/structure/counters"
	"github.com/benoitkugler/go-weasyprint/utils"
)

// Expand shorthands and validate property values.
// See http://www.w3.org/TR/CSS21/propidx.html and various CSS3 modules.

// :copyright: Copyright 2011-2014 Simon Sapin and contributors, see AUTHORS.
// :license: BSD, see LICENSE for details.

const prefix = "-weasy-"

var (
	InvalidValue = errors.New("invalid or unsupported values for a known CSS property")

	LENGTHUNITS = map[string]Unit{"ex": Ex, "em": Em, "ch": Ch, "rem": Rem, "px": Px, "pt": Pt, "pc": Pc, "in": In, "cm": Cm, "mm": Mm, "q": Q}

	// keyword -> (open, insert)
	ContentQuoteKeywords = map[string]Quote{
		"open-quote":     {Open: true, Insert: true},
		"close-quote":    {Open: false, Insert: true},
		"no-open-quote":  {Open: true, Insert: false},
		"no-close-quote": {Open: false, Insert: false},
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
		"text-decoration-color":      otherColors,
		"outline-color":              outlineColor,
		"border-collapse":            borderCollapse,
		"empty-cells":                emptyCells,
		"color":                      color,
		"transform-origin":           transformOrigin,
		"object-position":            backgroundPosition,
		"background-position":        backgroundPosition,
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
		"break-inside":               breakInside,
		"box-decoration-break":       boxDecorationBreak,
		"margin-break":               marginBreak,
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
		"font-family":                fontFamily,
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
		"letter-spacing":             spacing,
		"word-spacing":               spacing,
		"line-height":                lineHeight,
		"list-style-position":        listStylePosition,
		"list-style-type":            listStyleType,
		"padding-top":                lengthOrPercentage,
		"padding-right":              lengthOrPercentage,
		"padding-bottom":             lengthOrPercentage,
		"padding-left":               lengthOrPercentage,
		"min-width":                  minWidthHeight,
		"min-height":                 minWidthHeight,
		"max-width":                  maxWidthHeight,
		"max-height":                 maxWidthHeight,
		"opacity":                    opacity,
		"z-index":                    zIndex,
		"orphansWidows":              orphansWidows,
		"column-count":               columnCount,
		"overflow":                   overflow,
		"position":                   position,
		"quotes":                     quotes,
		"table-layout":               tableLayout,
		"text-align":                 textAlign,
		"text-decoration-line":       textDecorationLine,
		"text-decoration-style":      textDecorationStyle,
		"text-indent":                textIndent,
		"text-transform":             textTransform,
		"vertical-align":             verticalAlign,
		"visibility":                 visibility,
		"white-space":                whiteSpace,
		"overflow-wrap":              overflowWrap,
		"image-rendering":            imageRendering,
		"size":                       size,
		"anchor":                     anchor,
		"tab-size":                   tabSize,
		"hyphens":                    hyphens,
		"hyphenate-character":        hyphenateCharacter,
		"hyphenate-limit-zone":       hyphenateLimitZone,
		"hyphenate-limit-chars":      hyphenateLimitChars,
		"lang":                       lang,
		"bookmark-level":             bookmarkLevel,
		"bookmark-state":             bookmarkState,
		"object-fit":                 objectFit,
		"text-overflow":              textOverflow,
		"flex-basis":                 flexBasis,
		"flex-direction":             flexDirection,
		"flex-grow":                  flexGrowShrink,
		"flex-shrink":                flexGrowShrink,
		"order":                      order,
		"flex-wrap":                  flexWrap,
		"justify-content":            justifyContent,
		"align-items":                alignItems,
		"align-self":                 alignSelf,
		"align-content":              alignContent,
	}
	validatorsError = map[string]validatorError{
		"background-image":  backgroundImage,
		"list-style-image":  listStyleImage,
		"content":           content,
		"counter-increment": counterIncrement,
		"counter-reset":     counterReset,
		"font-size":         fontSize,
		"link":              link,
		"bookmark-label":    bookmarkLabel,
		"transform":         transform,
		"string-set":        stringSet,
	}
	// regroup the two cases (with error or without error)
	allValidators = Set{}

	proprietary = NewSet(
		"anchor",
		"link",
		"lang",
	)
	unstable = NewSet(
		"transform-origin",
		"size",
		"hyphens",
		"hyphenate-character",
		"hyphenate-limit-zone",
		"hyphenate-limit-chars",
		"bookmark-label",
		"bookmark-level",
		"bookmark-state",
		"string-set",
		"column-rule-color",
		"column-rule-style",
		"column-rule-width",
		"column-width",
		"column-span",
		"column-gap",
		"column-fill",
		"column-count",
		"bleed-left",
		"bleed-right",
		"bleed-top",
		"bleed-bottom",
		"marks",
	)
)

func init() {
	for _, couple := range couplesLigatures {
		for _, cc := range couple {
			allLigaturesValues[cc] = Has
		}
	}
	for _, couple := range couplesNumeric {
		for _, cc := range couple {
			allNumericValues[cc] = Has
		}
	}
	for _, couple := range couplesEastAsian {
		for _, cc := range couple {
			allEastAsianValues[cc] = Has
		}
	}
	for name := range validators {
		allValidators[name] = Has
	}
	for name := range validatorsError {
		allValidators[name] = Has
	}
}

type Token = parser.Token

type validator func(tokens []Token, baseUrl string) CssProperty
type validatorError func(tokens []Token, baseUrl string) (CssProperty, error)
type expander func(baseUrl, name string, tokens []Token) (NamedProperties, error)

type ValidatedProperty struct {
	Name      string
	Value     CssProperty
	Important bool
}

// If `token` is a keyword, return its name.
// Otherwise return empty string.
func getKeyword(token Token) string {
	if ident, ok := token.(parser.IdentToken); ok {
		return ident.Value.Lower()
	}
	return ""
}

// If `tokens` is a 1-element list of keywords, return its name.
// Otherwise return empty string.
func getSingleKeyword(tokens []Token) string {
	if len(tokens) == 1 {
		return getKeyword(tokens[0])
	}
	return ""
}

// negative  = true, percentage = false
func getLength(_token Token, negative, percentage bool) Dimension {
	switch token := _token.(type) {
	case parser.PercentageToken:
		if percentage && (negative || token.Value >= 0) {
			return Dimension{Value: token.Value, Unit: Percentage}
		}
	case parser.DimensionToken:
		unit, isKnown := LENGTHUNITS[string(token.Unit)]
		if isKnown && (negative || token.Value >= 0) {
			return Dimension{Value: token.Value, Unit: unit}
		}
	case parser.NumberToken:
		if token.Value == 0 {
			return Dimension{Unit: Scalar}
		}
	}
	return Dimension{}
}

// Return the value in radians of an <angle> token, or None.
func getAngle(token Token) (float32, bool) {
	if dim, ok := token.(parser.DimensionToken); ok {
		factor, in := ANGLETORADIANS[string(dim.Unit)]
		if in {
			return dim.Value * factor, true
		}
	}
	return 0, false
}

// Return the value := range dppx of a <resolution> token, || None.
func getResolution(token Token) (float32, bool) {
	if dim, ok := token.(parser.DimensionToken); ok {
		factor, in := RESOLUTIONTODPPX[string(dim.Unit)]
		if in {
			return dim.Value * factor, true
		}
	}
	return 0, false
}

func safeUrljoin(baseUrl, urls string) (string, error) {
	return utils.SafeUrljoin(baseUrl, urls, false)
}

//@validator()
//@commaSeparatedList
//@singleKeyword
// ``background-attachment`` property validation.
func _backgroundAttachment(tokens []Token) string {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "scroll", "fixed", "local":
		return keyword
	default:
		return ""
	}
}

func backgroundAttachment(tokens []Token, _ string) CssProperty {
	var out Strings
	for _, part := range SplitOnComma(tokens) {
		part = RemoveWhitespace(part)
		result := _backgroundAttachment(part)
		if result == "" {
			return nil
		}
		out = append(out, result)
	}
	return out
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
		c := parser.ParseColor(tokens[0])
		if !c.IsNone() {
			return Color(c)
		}
	}
	return nil
}

//@validator()
//@singleToken
func outlineColor(tokens []Token, _ string) CssProperty {
	if len(tokens) == 1 {
		token := tokens[0]
		if getKeyword(token) == "invert" {
			return Color{Type: parser.ColorCurrentColor}
		} else {
			return Color(parser.ParseColor(token))
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
		result := parser.ParseColor(token)
		if result.Type == parser.ColorCurrentColor {
			return Color{Type: parser.ColorInherit}
		} else {
			return Color(result)
		}
	}
	return nil
}

// @validator("background-image", wantsBaseUrl=true)
// @commaSeparatedList
// @singleToken
func _backgroundImage(tokens []Token, baseUrl string) (Image, error) {
	if len(tokens) != 1 {
		return nil, nil
	}
	_token := tokens[0]

	token, ok := _token.(parser.FunctionBlock)
	if !ok {
		if getKeyword(token) == "none" {
			return NoneImage{}, nil
		}
	}
	return getImage(_token, baseUrl)
}

func backgroundImage(tokens []Token, baseUrl string) (CssProperty, error) {
	var out Images
	for _, part := range SplitOnComma(tokens) {
		part = RemoveWhitespace(part)
		result, err := _backgroundImage(part, baseUrl)
		if err != nil {
			return nil, err
		}
		if result == nil {
			return nil, nil
		}
		out = append(out, result)

	}
	return out, nil
}

var directionKeywords = map[[3]string]DirectionType{
	// ("angle", radians)  0 upwards, then clockwise
	{"to", "top", ""}:    {Angle: 0},
	{"to", "right", ""}:  {Angle: math.Pi / 2},
	{"to", "bottom", ""}: {Angle: math.Pi},
	{"to", "left", ""}:   {Angle: math.Pi * 3 / 2},
	// ("corner", keyword)
	{"to", "top", "left"}:     {Corner: "top_left"},
	{"to", "left", "top"}:     {Corner: "top_left"},
	{"to", "top", "right"}:    {Corner: "top_right"},
	{"to", "right", "top"}:    {Corner: "top_right"},
	{"to", "bottom", "left"}:  {Corner: "bottom_left"},
	{"to", "left", "bottom"}:  {Corner: "bottom_left"},
	{"to", "bottom", "right"}: {Corner: "bottom_right"},
	{"to", "right", "bottom"}: {Corner: "bottom_right"},
}

//@validator("list-style-image", wantsBaseUrl=true)
//@singleToken
// ``list-style-image`` property validation.
func listStyleImage(tokens []Token, baseUrl string) (CssProperty, error) {
	if len(tokens) != 1 {
		return nil, nil
	}
	token := tokens[0]

	if token.Type() != "function" {
		if getKeyword(token) == "none" {
			return NoneImage{}, nil
		}
		parsedUrl, err := getUrl(token, baseUrl)
		if err != nil {
			return nil, err
		}
		if contentUrl, ok := parsedUrl.Content.(Url); parsedUrl.Type == "url" && ok && !contentUrl.Internal {
			return UrlImage(contentUrl.Url), nil
		}
	}
	return nil, nil
}

var centerKeywordFakeToken = parser.IdentToken{Value: "center"}

//@validator(unstable=true)
func transformOrigin(tokens []Token, _ string) CssProperty {
	if len(tokens) == 3 {
		// Ignore third parameter as 3D transforms are ignored.
		tokens = tokens[:2]
	}
	return parse2dPosition(tokens)
}

//@validator()
//@commaSeparatedList
// ``background-position`` and ``object-position`` property validation.
// See http://dev.w3.org/csswg/css3-background/#the-background-position
func backgroundPosition(tokens []Token, _ string) CssProperty {
	var out Centers
	for _, part := range SplitOnComma(tokens) {
		result := parsePosition(RemoveWhitespace(part))
		if result.IsNone() {
			return nil
		}
		out = append(out, result)
	}
	return out
}

// Common syntax of background-position and transform-origin.
func parse2dPosition(tokens []Token) Point {
	if len(tokens) == 1 {
		tokens = []Token{tokens[0], centerKeywordFakeToken}
	} else if len(tokens) != 2 {
		return Point{}
	}

	token1, token2 := tokens[0], tokens[1]
	length1 := getLength(token1, true, true)
	length2 := getLength(token2, true, true)
	if !length1.IsNone() && !length2.IsNone() {
		return Point{length1, length2}
	}
	keyword1, keyword2 := getKeyword(token1), getKeyword(token2)
	if !length1.IsNone() && (keyword2 == "top" || keyword2 == "center" || keyword2 == "bottom") {
		return Point{length1, backgroundPositionsPercentages[keyword2]}
	} else if !length2.IsNone() && (keyword1 == "left" || keyword1 == "center" || keyword1 == "right") {
		return Point{backgroundPositionsPercentages[keyword1], length2}
	} else if (keyword1 == "left" || keyword1 == "center" || keyword1 == "right") &&
		(keyword2 == "top" || keyword2 == "center" || keyword2 == "bottom") {
		return Point{backgroundPositionsPercentages[keyword1], backgroundPositionsPercentages[keyword2]}
	} else if (keyword1 == "top" || keyword1 == "center" || keyword1 == "bottom") &&
		(keyword2 == "left" || keyword2 == "center" || keyword2 == "right") {
		// Swap tokens. They need to be in (horizontal, vertical) order.
		return Point{backgroundPositionsPercentages[keyword2], backgroundPositionsPercentages[keyword1]}
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
	var out Repeats
	for _, part := range SplitOnComma(tokens) {
		result := _backgroundRepeat(RemoveWhitespace(part))
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
			return Size{Width: SToV("auto"), Height: SToV("auto")}
		}
		length := getLength(token, false, true)
		if !length.IsNone() {
			return Size{Width: Value{Dimension: length}, Height: SToV("auto")}
		}
	case 2:
		var out Size
		lengthW := getLength(tokens[0], false, true)
		lengthH := getLength(tokens[1], false, true)
		if !lengthW.IsNone() {
			out.Width = Value{Dimension: lengthW}
		} else if getKeyword(tokens[0]) == "auto" {
			out.Width = SToV("auto")
		} else {
			return Size{}
		}
		if !lengthH.IsNone() {
			out.Height = Value{Dimension: lengthH}
		} else if getKeyword(tokens[1]) == "auto" {
			out.Height = SToV("auto")
		} else {
			return Size{}
		}
		return out
	}
	return Size{}
}

func backgroundSize(tokens []Token, _ string) CssProperty {
	var out Sizes
	for _, part := range SplitOnComma(tokens) {
		result := _backgroundSize(RemoveWhitespace(part))
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
func _box(tokens []Token) string {
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
	for _, part := range SplitOnComma(tokens) {
		result := _box(RemoveWhitespace(part))
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
			return Point{lengths[0], lengths[0]}
		} else if len(lengths) == 2 {
			return Point{lengths[0], lengths[1]}
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
		return String(keyword)
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

//@validator()
//@singleKeyword
// ``box-decoration-break`` property validation.
func boxDecorationBreak(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "slice", "clone":
		return String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleKeyword
// ``margin-break`` property validation.
func marginBreak(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "keep", "discard":
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
	if ident, ok := token.(parser.IdentToken); ok {
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
		return Value{String: "auto"}
	} else {
		return Value{Dimension: getLength(token, true, false)}
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
		return Value{Dimension: length}
	}
	keyword := getKeyword(token)
	if keyword == "thin" || keyword == "medium" || keyword == "thick" {
		return Value{String: keyword}
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
		return Value{Dimension: length}
	}
	keyword := getKeyword(token)
	if keyword == "auto" {
		return Value{String: keyword}
	}
	return nil
}

// //@validator()
// //@singleKeyword
// ``column-span`` property validation.
func columnSpan(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "all", "none":
		return String(keyword)
	default:
		return nil
	}
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
			var values Values
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
		return Values{}
	}
	return nil
}

// //@validator(wantsBaseUrl=true)
// ``content`` property validation.
func content(tokens []Token, baseUrl string) (CssProperty, error) {
	var (
		parsedTokens []ContentProperty
		token        Token
	)
	for len(tokens) > 0 {
		if len(tokens) >= 2 {
			if lit, ok := tokens[1].(parser.LiteralToken); ok && lit.Value == "," {
				token, tokens = tokens[0], tokens[2:]
				im, err := getImage(token, baseUrl)
				if err != nil {
					return nil, err
				}
				cp := ContentProperty{Type: "image", Content: im}
				if im == nil {
					cp, err = getUrl(token, baseUrl)
					if err != nil {
						return nil, err
					}
				}
				if !cp.IsNone() {
					parsedTokens = append(parsedTokens, cp)
				} else {
					return nil, nil
				}
			} else {
				break
			}
		} else {
			break
		}
	}

	if len(tokens) == 0 {
		return nil, nil
	}
	if len(tokens) >= 3 {
		lit, ok := tokens[len(tokens)-2].(parser.LiteralToken)
		if tokens[len(tokens)-1].Type() == parser.TypeStringToken && ok && lit.Value == "/" {
			// Ignore text for speech
			tokens = tokens[:len(tokens)-2]
		}
	}

	keyword := getSingleKeyword(tokens)
	if keyword == "normal" || keyword == "none" {
		return SContent{String: keyword}, nil
	}
	l, err := getContentList(tokens, baseUrl)
	if l == nil || err != nil {
		return nil, err
	}
	return SContent{Contents: l}, nil
}

// // helpers for validateContentToken type switches
// func _isIdent(args []Token) (bool, parser.IdentToken) {
// 	if len(args) == 1 {
// 		out, ok := args[0].(parser.IdentToken)
// 		return ok, out
// 	}
// 	return false, parser.IdentToken{}
// }
// func _isIdent2(args []Token) (bool, parser.IdentToken, parser.IdentToken) {
// 	if len(args) == 2 {
// 		out1, ok1 := args[0].(parser.IdentToken)
// 		out2, ok2 := args[1].(parser.IdentToken)
// 		return ok1 && ok2, out1, out2
// 	}
// 	return false, parser.IdentToken{}, parser.IdentToken{}
// }
// func _isIdentString(args []Token) (bool, parser.IdentToken, parser.StringToken) {
// 	if len(args) == 2 {
// 		out1, ok1 := args[0].(parser.IdentToken)
// 		out2, ok2 := args[1].(parser.StringToken)
// 		return ok1 && ok2, out1, out2
// 	}
// 	return false, parser.IdentToken{}, parser.StringToken{}
// }
// func _isIdentStringIdent(args []Token) (bool, parser.IdentToken, parser.StringToken, parser.IdentToken) {
// 	if len(args) == 3 {
// 		out1, ok1 := args[0].(parser.IdentToken)
// 		out2, ok2 := args[1].(parser.StringToken)
// 		out3, ok3 := args[2].(parser.IdentToken)
// 		return ok1 && ok2 && ok3, out1, out2, out3
// 	}
// 	return false, parser.IdentToken{}, parser.StringToken{}, parser.IdentToken{}
// }

// // share attr, counter, counters cases
// func _parseContentArgs(name string, args []Token) ContentProperty {
// 	switch name {
// 	case "attr":
// 		ok, ident := _isIdent(args)
// 		if ok {
// 			return ContentProperty{Type: ContentAttr, SStrings: SStrings{String: string(ident.Value)}}
// 		}
// 	case "counter":
// 		if ok, ident := _isIdent(args); ok {
// 			return ContentProperty{Type: ContentCounter, SStrings: SStrings{Strings: []string{string(ident.Value), "decimal"}}}
// 		}
// 		if ok, ident, ident2 := _isIdent2(args); ok {
// 			style := string(ident2.Value)
// 			_, isIn := counters.STYLES[style]
// 			if style == "none" || style == "decimal" || isIn {
// 				return ContentProperty{Type: ContentCounter, SStrings: SStrings{Strings: []string{string(ident.Value), style}}}
// 			}
// 		}
// 	case "counters":
// 		if ok, ident, stri := _isIdentString(args); ok {
// 			return ContentProperty{Type: ContentCounters, SStrings: SStrings{Strings: []string{string(ident.Value), stri.Value, "decimal"}}}
// 		}
// 		if ok, ident, stri, ident2 := _isIdentStringIdent(args); ok {
// 			style := string(ident2.Value)
// 			_, isIn := counters.STYLES[style]
// 			if style == "none" || style == "decimal" || isIn {
// 				return ContentProperty{Type: ContentCounters, SStrings: SStrings{Strings: []string{string(ident.Value), stri.Value, style}}}
// 			}
// 		}
// 	}
// 	return ContentProperty{}
// }

// // Validation for a single token for the ``content`` property.
// // Return (type, content) || zero for invalid tokens.
// func validateContentToken(baseUrl string, token Token) (ContentProperty, error) {
// 	quoteType, isNotNone := ContentQuoteKeywords[getKeyword(token)]
// 	if isNotNone {
// 		return ContentProperty{Type: ContentQUOTE, Quote: quoteType}, nil
// 	}

// 	switch tt := token.(type) {
// 	case parser.StringToken:
// 		return ContentProperty{Type: ContentSTRING, SStrings: SStrings{String: tt.Value}}, nil
// 	case parser.URLToken:
// 		url, err := safeUrljoin(baseUrl, tt.Value)
// 		if err != nil {
// 			return ContentProperty{}, err
// 		}
// 		return ContentProperty{Type: ContentURI, SStrings: SStrings{String: url}}, nil
// 	}

// 	name, args := parseFunction(token)
// 	if name != "" {
// 		switch name {
// 		case "string":
// 			var stringArgs []string
// 			if ok, ident := _isIdent(args); ok {
// 				stringArgs = []string{string(ident.Value)}
// 			}
// 			if ok, ident, ident2 := _isIdent2(args); ok {
// 				args2 := ident2.Value.Lower()
// 				if args2 != "first" && args2 != "start" && args2 != "last" && args2 != "first-except" {
// 					return ContentProperty{}, fmt.Errorf("Invalid or unsupported CSS value : %s", args2)
// 				}
// 				stringArgs = []string{string(ident.Value), args2}
// 			}
// 			if stringArgs != nil { // thus one of the checks passed
// 				return ContentProperty{Type: ContentString, SStrings: SStrings{Strings: stringArgs}}, nil
// 			}
// 		default:
// 			return _parseContentArgs(name, args), nil
// 		}
// 	}
// 	return ContentProperty{}, nil
// }

// //@validator()
// ``counter-increment`` property validation.
func counterIncrement(tokens []Token, _ string) (CssProperty, error) {
	ci, err := counter(tokens, 1)
	if err != nil || ci == nil {
		return nil, err
	}
	return SIntStrings{Values: ci}, nil
}

// //@validator()
// ``counter-reset`` property validation.
func counterReset(tokens []Token, _ string) (CssProperty, error) {
	iss, err := counter(tokens, 0)
	if err != nil || iss == nil {
		return nil, err
	}
	return IntStrings(iss), err
}

// ``counter-increment`` && ``counter-reset`` properties validation.
func counter(tokens []Token, defaultInteger int) ([]IntString, error) {
	if getSingleKeyword(tokens) == "none" {
		return []IntString{}, nil
	}
	if len(tokens) == 0 {
		return nil, errors.New("got an empty token list")
	}
	var (
		results []IntString
		integer int
	)
	iter := parser.NewTokenIterator(tokens)
	token := iter.Next()
	for token != nil {
		ident, ok := token.(parser.IdentToken)
		if !ok {
			return nil, nil // expected a keyword here
		}
		counterName := ident.Value
		if counterName == "none" || counterName == "initial" || counterName == "inherit" {
			return nil, fmt.Errorf("Invalid counter name: %s", counterName)
		}
		token = iter.Next()
		if number, ok := token.(parser.NumberToken); ok && number.IsInteger { // implies token != nil
			// Found an integer. Use it and get the next token
			integer = number.IntValue()
			token = iter.Next()
		} else {
			// Not an integer. Might be the next counter name.
			// Keep `token` for the next loop iteration.
			integer = defaultInteger
		}
		results = append(results, IntString{String: string(counterName), Int: integer})
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
		return Value{Dimension: length}
	}
	if getKeyword(token) == "auto" {
		return Value{String: "auto"}
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
		return Value{Dimension: length}
	}
	if getKeyword(token) == "auto" {
		return Value{String: "auto"}
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
		return Value{Dimension: length}
	}
	if getKeyword(token) == "normal" {
		return Value{String: "normal"}
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
		"table-row", "table-column-group", "table-column", "table-cell",
		"flex", "inline-flex":
		return String(keyword)
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
		return String(keyword)
	default:
		return nil
	}
}

func _fontFamily(tokens []Token) string {
	if len(tokens) == 0 {
		return ""
	}
	if tt, ok := tokens[0].(parser.StringToken); len(tokens) == 1 && ok {
		return tt.Value
	} else if len(tokens) > 0 {
		var values []string
		for _, token := range tokens {
			if tt, ok := token.(parser.IdentToken); ok {
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
func fontFamily(tokens []Token, _ string) CssProperty {
	var out Strings
	for _, part := range SplitOnComma(tokens) {
		result := _fontFamily(RemoveWhitespace(part))
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
	if tt, ok := token.(parser.StringToken); ok {
		return String(tt.Value)
	}
	return nil
}

func parseFontVariant(tokens []Token, all Set, couples [][]string) SStrings {
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
		ident, isIdent := token.(parser.IdentToken)
		if !isIdent {
			return SStrings{}
		}
		identValue := string(ident.Value)
		if all.Has(identValue) {
			var concurrentValues []string
			for _, couple := range couples {
				if isInValues(identValue, couple) {
					concurrentValues = couple
					break
				}
			}
			for _, value := range concurrentValues {
				if isInValues(value, values) {
					return SStrings{}
				}
			}
			values = append(values, identValue)
		} else {
			return SStrings{}
		}
	}
	if len(values) > 0 {
		return SStrings{Strings: values}
	}
	return SStrings{}
}

// //@validator()
func fontVariantLigatures(tokens []Token, _ string) CssProperty {
	if len(tokens) == 1 {
		keyword := getKeyword(tokens[0])
		if keyword == "normal" || keyword == "none" {
			return SStrings{String: keyword}
		}
	}
	ss := parseFontVariant(tokens, allLigaturesValues, couplesLigatures)
	if ss.IsNone() {
		return nil
	}
	return ss
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
			return SStrings{String: keyword}
		}
	}
	ss := parseFontVariant(tokens, allNumericValues, couplesNumeric)
	if ss.IsNone() {
		return nil
	}
	return ss
}

// //@validator()
// ``font-feature-settings`` property validation.
func fontFeatureSettings(tokens []Token, _ string) CssProperty {
	s := _fontFeatureSettings(tokens)
	if s.IsNone() {
		return nil
	}
	return s
}

func _fontFeatureSettings(tokens []Token) SIntStrings {
	if len(tokens) == 1 && getKeyword(tokens[0]) == "normal" {
		return SIntStrings{String: "normal"}
	}

	fontFeatureSettingsList := func(tokens []Token) IntString {
		var token Token
		feature, value := "", 0

		if len(tokens) == 2 {
			tokens, token = tokens[0:1], tokens[1]
			switch tt := token.(type) {
			case parser.IdentToken:
				if tt.Value == "on" {
					value = 1
				} else {
					value = 0
				}
			case parser.NumberToken:
				if tt.IsInteger && tt.IntValue() >= 0 {
					value = tt.IntValue()
				}
			}
		} else if len(tokens) == 1 {
			value = 1
		}

		if len(tokens) == 1 {
			token = tokens[0]
			tt, ok := token.(parser.StringToken)
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
			return IntString{String: feature, Int: value}
		}
		return IntString{}
	}

	var out SIntStrings
	for _, part := range SplitOnComma(tokens) {
		result := fontFeatureSettingsList(RemoveWhitespace(part))
		if (result == IntString{}) {
			return SIntStrings{}
		}
		out.Values = append(out.Values, result)
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
			return SStrings{String: keyword}
		}
	}
	ss := parseFontVariant(tokens, allEastAsianValues, couplesEastAsian)
	if ss.IsNone() {
		return nil
	}
	return ss
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
		return Value{Dimension: length}, nil
	}
	fontSizeKeyword := getKeyword(token)
	if _, isIn := FontSizeKeywords[fontSizeKeyword]; isIn || fontSizeKeyword == "smaller" || fontSizeKeyword == "larger" {
		return Value{String: fontSizeKeyword}, nil
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
		return IntString{String: keyword}
	}
	if number, ok := token.(parser.NumberToken); ok {
		intValue := number.IntValue()
		if number.IsInteger && (intValue == 100 || intValue == 200 || intValue == 300 || intValue == 400 || intValue == 500 || intValue == 600 || intValue == 700 || intValue == 800 || intValue == 900) {
			return IntString{Int: intValue}
		}
	}
	return nil
}

// @validator()
// @single_keyword
func objectFit(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	//TODO: Figure out what the spec means by "'scale-down' flag".
	//  As of this writing, neither Firefox nor chrome support
	//  anything other than a single keyword as is done here.
	switch keyword {
	case "fill", "contain", "cover", "none", "scale-down":
		return String(keyword)
	default:
		return nil
	}
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
	return FToV(value)
}

//@validator("letter-spacing")
//@validator("word-spacing")
//@singleToken
// Validation for ``letter-spacing`` && ``word-spacing``.
func spacing(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if getKeyword(token) == "normal" {
		return Value{String: "normal"}
	}
	length := getLength(token, true, false)
	if !length.IsNone() {
		return Value{Dimension: length}
	}
	return nil
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
		return Value{String: "normal"}
	}

	switch tt := token.(type) {
	case parser.NumberToken:
		if tt.Value >= 0 {
			return Value{Dimension: Dimension{Value: tt.Value, Unit: Scalar}}
		}
	case parser.PercentageToken:
		if tt.Value >= 0 {
			return Value{Dimension: Dimension{Value: tt.Value, Unit: Percentage}}
		}
	case parser.DimensionToken:
		if tt.Value >= 0 {
			l := getLength(token, true, false)
			if l.IsNone() {
				return nil
			}
			return l.ToValue()
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
	if keyword == "none" || inStyles {
		return String(keyword)
	} else {
		return nil
	}
}

// @validator("min-width")
// @validator("min-height")
// @singleToken
// ``min-width`` && ``min-height`` properties validation.
func minWidthHeight(tokens []Token, _ string) CssProperty {
	// See https://www.w3.org/TR/css-flexbox-1/#min-size-auto
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	keyword := getKeyword(token)
	if keyword == "auto" {
		return SToV(keyword)
	} else {
		return lengthOrPercentage([]Token{token}, "")
	}
}

//@validator("padding-top")
//@validator("padding-right")
//@validator("padding-bottom")
//@validator("padding-left")
//@singleToken
// ``padding-*`` properties validation.
func lengthOrPercentage(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	l := getLength(token, false, true)
	if l.IsNone() {
		return nil
	}
	return Value{Dimension: l}
}

//@validator("max-width")
//@validator("max-height")
//@singleToken
// Validation for max-width && max-height
func maxWidthHeight(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	length := getLength(token, false, true)
	if !length.IsNone() {
		return Value{Dimension: length}
	}
	if getKeyword(token) == "none" {
		return Value{Dimension: Dimension{Value: float32(math.Inf(1.)), Unit: Px}}
	}
	return nil
}

//@validator()
//@singleToken
// Validation for the ``opacity`` property.
func opacity(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if number, ok := token.(parser.NumberToken); ok {
		return Float(utils.Min(1, utils.Max(0, number.Value)))
	}
	return nil
}

//@validator()
//@singleToken
// Validation for the ``z-index`` property.
func zIndex(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if getKeyword(token) == "auto" {
		return IntString{String: "auto"}
	}
	if number, ok := token.(parser.NumberToken); ok {
		if number.IsInteger {
			return IntString{Int: number.IntValue()}
		}
	}
	return nil
}

//@validator("orphans")
//@validator("widows")
//@singleToken
// Validation for the ``orphans`` && ``widows`` properties.
func orphansWidows(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if number, ok := token.(parser.NumberToken); ok {
		value := number.IntValue()
		if number.IsInteger && value >= 1 {
			return Int(value)
		}
	}
	return nil
}

//@validator()
//@singleToken
// Validation for the ``column-count`` property.
func columnCount(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if number, ok := token.(parser.NumberToken); ok {
		value := number.IntValue()
		if number.IsInteger && value >= 1 {
			return IntString{Int: value}
		}
	}
	if getKeyword(token) == "auto" {
		return IntString{String: "auto"}
	}
	return nil
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

// @validator()
// @single_keyword
// Validation for the ``text-overflow`` property.
func textOverflow(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "clip", "ellipsis":
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
func quotes(tokens []Token, _ string) CssProperty {
	var opens, closes []string
	if len(tokens) > 0 && len(tokens)%2 == 0 {
		// Separate open && close quotes.
		// eg.  ("«", "»", "“", "”")  -> (("«", "“"), ("»", "”"))
		for i := 0; i < len(tokens); i += 2 {
			open, ok1 := tokens[i].(parser.StringToken)
			close_, ok2 := tokens[i+1].(parser.StringToken)
			if ok1 && ok2 {
				opens = append(opens, open.Value)
				closes = append(closes, close_.Value)
			} else {
				return nil
			}
		}
		return Quotes{Open: opens, Close: closes}
	}
	return nil
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

// @validator()
// ``text-decoration-line`` property validation.
func textDecorationLine(tokens []Token, _ string) CssProperty {
	uniqKeywords := Set{}
	valid := true
	for _, token := range tokens {
		keyword := getKeyword(token)
		if !(keyword == "underline" || keyword == "overline" || keyword == "line-through" || keyword == "blink") {
			valid = false
		}
		uniqKeywords.Add(keyword)
	}
	if _, in := uniqKeywords["none"]; len(uniqKeywords) == 1 && in { // then uniqKeywords == {"none"}
		return NDecorations{None: true}
	}
	if valid && len(uniqKeywords) == len(tokens) {
		return NDecorations{Decorations: uniqKeywords}
	}
	return nil
}

// @validator()
// @singleKeyword
// ``text-decoration-style`` property validation.
func textDecorationStyle(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "solid", "double", "dotted", "dashed", "wavy":
		return String(keyword)
	default:
		return nil
	}
}

//@validator()
//@singleToken
// ``text-indent`` property validation.
func textIndent(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	l := getLength(token, true, true)
	if l.IsNone() {
		return nil
	}
	return Value{Dimension: l}
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
func verticalAlign(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	length := getLength(token, true, true)
	if !length.IsNone() {
		return Value{Dimension: length}
	}
	keyword := getKeyword(token)
	if keyword == "baseline" || keyword == "middle" || keyword == "sub" || keyword == "super" || keyword == "text-top" || keyword == "text-bottom" || keyword == "top" || keyword == "bottom" {
		return Value{String: keyword}
	}
	return nil
}

//@validator()
//@singleKeyword
// ``visibility`` property validation.
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

// @validator()
// @singleToken
// ``flex-basis`` property validation.
func flexBasis(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	basis := widthHeight(tokens, "")
	if basis != nil {
		return basis
	}
	if getKeyword(token) == "content" {
		return SToV("content")
	}
	return nil
}

// @validator()
// @singleKeyword
// ``flex-direction`` property validation.
func flexDirection(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "row", "row-reverse", "column", "column-reverse":
		return String(keyword)
	default:
		return nil
	}
}

// @validator("flex-grow")
// @validator("flex-shrink")
// @singleToken
func flexGrowShrink(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if number, ok := token.(parser.NumberToken); ok {
		return Float(number.Value)
	}
	return nil
}

// @validator()
// @singleToken
func order(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if number, ok := token.(parser.NumberToken); ok && number.IsInteger {
		return Int(number.IntValue())
	}
	return nil
}

// @validator()
// @singleKeyword
// ``flex-wrap`` property validation.
func flexWrap(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "nowrap", "wrap", "wrap-reverse":
		return String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
// ``justify-content`` property validation.
func justifyContent(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "flex-start", "flex-end", "center", "space-between", "space-around", "space-evenly", "stretch":
		return String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
// ``align-items`` property validation.
func alignItems(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "flex-start", "flex-end", "center", "baseline", "stretch":
		return String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
// ``align-self`` property validation.
func alignSelf(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "auto", "flex-start", "flex-end", "center", "baseline", "stretch":
		return String(keyword)
	default:
		return nil
	}
}

// @validator()
// @singleKeyword
// ``align-content`` property validation.
func alignContent(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "flex-start", "flex-end", "center", "space-between", "space-around",
		"space-evenly", "stretch":
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
func size(tokens []Token, _ string) CssProperty {
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
			return Point{lengths[0], lengths[0]}
		} else if len(lengths) == 2 {
			return Point{lengths[0], lengths[1]}
		}
	}

	if len(keywords) == 1 {
		keyword := keywords[0]
		if psize, in := PageSizes[keyword]; in {
			return psize
		} else if keyword == "auto" || keyword == "portrait" {
			return InitialWidthHeight
		} else if keyword == "landscape" {
			return Point{InitialWidthHeight[1], InitialWidthHeight[0]}
		}
	}

	if len(keywords) == 2 {
		var orientation, pageSize string
		if keywords[0] == "portrait" || keywords[0] == "landscape" {
			orientation, pageSize = keywords[0], keywords[1]
		} else if keywords[1] == "portrait" || keywords[1] == "landscape" {
			pageSize, orientation = keywords[0], keywords[1]
		}
		if widthHeight, in := PageSizes[pageSize]; in {
			if orientation == "portrait" {
				return widthHeight
			} else {
				return Point{widthHeight[1], widthHeight[0]}
			}
		}
	}
	return nil
}

//@validator(proprietary=true)
//@singleToken
// Validation for ``anchor``.
func anchor(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if getKeyword(token) == "none" {
		return NamedString{Name: "none"}
	}
	name, args := parseFunction(token)
	if name != "" {
		if len(args) == 1 {
			if ident, ok := args[0].(parser.IdentToken); ok && name == "attr" {
				return NamedString{Name: "attr()", String: string(ident.Value)}
			}
		}
	}
	return nil
}

//@validator(proprietary=true, wantsBaseUrl=true)
//@singleToken
// Validation for ``link``.
func link(tokens []Token, baseUrl string) (CssProperty, error) {
	if len(tokens) != 1 {
		return nil, nil
	}
	token := tokens[0]
	if getKeyword(token) == "none" {
		return NamedString{Name: "none"}, nil
	}

	parsedUrl, err := getUrl(token, baseUrl)
	if err != nil {
		return nil, err
	}
	if !parsedUrl.IsNone() {
		return parsedUrl, nil
	}
	name, args := parseFunction(token)
	if name != "" {
		if len(args) == 1 {
			if ident, ok := args[0].(parser.IdentToken); ok && name == "attr" {
				return ContentProperty{Type: "attr()", Content: String(ident.Value)}, nil
			}
		}
	}
	return nil, nil
}

//@validator()
//@singleToken
// Validation for ``tab-size``.
// See https://www.w3.org/TR/css-text-3/#tab-size
func tabSize(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if number, ok := token.(parser.NumberToken); ok {
		if number.IsInteger && number.Value >= 0 {
			return FToV(number.Value)
		}
	}
	return Value{Dimension: getLength(token, false, false)}
}

//@validator(unstable=true)
//@singleToken
// Validation for ``hyphens``.
func hyphens(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
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
func hyphenateCharacter(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	keyword := getKeyword(token)
	if keyword == "auto" {
		return String("‐")
	} else if str, ok := token.(parser.StringToken); ok {
		return String(str.Value)
	}
	return nil
}

//@validator(unstable=true)
//@singleToken
// Validation for ``hyphenate-limit-zone``.
func hyphenateLimitZone(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	d := getLength(token, false, true)
	if d.IsNone() {
		return nil
	}
	return Value{Dimension: d}
}

//@validator(unstable=true)
// Validation for ``hyphenate-limit-chars``.
func hyphenateLimitChars(tokens []Token, _ string) CssProperty {
	switch len(tokens) {
	case 1:
		token := tokens[0]
		keyword := getKeyword(token)
		if keyword == "auto" {
			return Ints3{5, 2, 2}
		} else if number, ok := token.(parser.NumberToken); ok && number.IsInteger {
			return Ints3{number.IntValue(), 2, 2}

		}
	case 2:
		total, left := tokens[0], tokens[1]
		totalKeyword := getKeyword(total)
		leftKeyword := getKeyword(left)
		if totalNumber, ok := total.(parser.NumberToken); ok && totalNumber.IsInteger {
			if leftNumber, ok := left.(parser.NumberToken); ok && leftNumber.IsInteger {
				return Ints3{totalNumber.IntValue(), leftNumber.IntValue(), leftNumber.IntValue()}
			} else if leftKeyword == "auto" {
				return Ints3{totalNumber.IntValue(), 2, 2}
			}
		} else if totalKeyword == "auto" {
			if leftNumber, ok := left.(parser.NumberToken); ok && leftNumber.IsInteger {
				return Ints3{5, leftNumber.IntValue(), leftNumber.IntValue()}
			} else if leftKeyword == "auto" {
				return Ints3{5, 2, 2}
			}
		}
	case 3:
		total, left, right := tokens[0], tokens[1], tokens[2]
		totalNumber, okT := total.(parser.NumberToken)
		leftNumber, okL := left.(parser.NumberToken)
		rightNumber, okR := right.(parser.NumberToken)
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
			return Ints3{totalInt, leftInt, rightInt}
		}
	}
	return nil
}

//@validator(proprietary=true)
//@singleToken
// Validation for ``lang``.
func lang(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if getKeyword(token) == "none" {
		return NamedString{Name: "none"}
	}
	name, args := parseFunction(token)
	if name != "" {
		if len(args) == 1 {
			if ident, ok := args[0].(parser.IdentToken); ok && name == "attr" {
				return NamedString{Name: "attr()", String: string(ident.Value)}
			}
		}

	} else if str, ok := token.(parser.StringToken); ok {
		return NamedString{Name: "string", String: str.Value}
	}
	return nil
}

//@validator(unstable=true)
// Validation for ``bookmark-label``.
func bookmarkLabel(tokens []Token, baseUrl string) (out CssProperty, err error) {
	parsedTokens := make(BookmarkLabel, len(tokens))
	for index, v := range tokens {
		parsedTokens[index], err = getContentListToken(v, baseUrl)
		if err != nil {
			return nil, err
		}
		if parsedTokens[index].IsNone() {
			return nil, nil
		}
	}
	return parsedTokens, nil
}

//@validator(unstable=true)
//@singleToken
// Validation for ``bookmark-level``.
func bookmarkLevel(tokens []Token, _ string) CssProperty {
	if len(tokens) != 1 {
		return nil
	}
	token := tokens[0]
	if number, ok := token.(parser.NumberToken); ok && number.IsInteger && number.IntValue() >= 1 {
		return IntString{Int: number.IntValue()}
	} else if getKeyword(token) == "none" {
		return IntString{String: "none"}
	}
	return nil
}

// @validator(unstable=True)
// @single_keyword
// Validation for ``bookmark-state``.
func bookmarkState(tokens []Token, _ string) CssProperty {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "open", "closed":
		return String(keyword)
	default:
		return nil
	}
}

//@validator(unstable=true)
//@commaSeparatedList
// Validation for ``string-set``.
func _stringSet(tokens []Token, baseUrl string) (out SContent, err error) {
	// Spec asks for strings after custom keywords, but we allow content-lists
	if len(tokens) >= 2 {
		varName := getCustomIdent(tokens[0])
		if varName == "" {
			return
		}
		parsedTokens := make([]ContentProperty, len(tokens)-1)
		for i, token := range tokens[1:] {
			parsedTokens[i], err = getContentListToken(token, baseUrl)
			if err != nil {
				return
			}
			if parsedTokens[i].IsNone() {
				return
			}
		}
		return SContent{String: varName, Contents: parsedTokens}, nil
	} else if len(tokens) > 0 && getKeyword(tokens[0]) == "none" {
		return SContent{String: "none"}, nil
	}
	return
}

func stringSet(tokens []Token, baseUrl string) (CssProperty, error) {
	var out StringSet
	for _, part := range SplitOnComma(tokens) {
		result, err := _stringSet(RemoveWhitespace(part), baseUrl)
		if err != nil {
			return nil, err
		}
		if result.IsNone() {
			return nil, nil
		}
		out.Contents = append(out.Contents, result)
	}
	return out, nil
}

//@validator()
func transform(tokens []Token, _ string) (CssProperty, error) {
	if getSingleKeyword(tokens) == "none" {
		return Transforms{}, nil
	}
	out := make(Transforms, len(tokens))
	var err error
	for index, v := range tokens {
		out[index], err = transformFunction(v)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func transformFunction(token Token) (SDimensions, error) {
	name, args := parseFunction(token)
	if name == "" {
		return SDimensions{}, InvalidValue
	}

	lengths, values := make([]Dimension, len(args)), make([]Dimension, len(args))
	isAllNumber, isAllLengths := true, true
	for index, a := range args {
		lengths[index] = getLength(a, true, true)
		isAllLengths = isAllLengths && !lengths[index].IsNone()
		if aNumber, ok := a.(parser.NumberToken); ok {
			values[index] = FToD(aNumber.Value)
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
				return SDimensions{String: name, Dimensions: []Dimension{FToD(angle)}}, nil
			}
		case "translatex", "translate":
			if !length.IsNone() {
				return SDimensions{String: "translate", Dimensions: []Dimension{length, ZeroPixels}}, nil
			}
		case "translatey":
			if !length.IsNone() {
				return SDimensions{String: "translate", Dimensions: []Dimension{ZeroPixels, length}}, nil
			}
		case "scalex":
			if number, ok := args[0].(parser.NumberToken); ok {
				return SDimensions{String: "scale", Dimensions: []Dimension{FToD(number.Value), FToD(1.)}}, nil
			}
		case "scaley":
			if number, ok := args[0].(parser.NumberToken); ok {
				return SDimensions{String: "scale", Dimensions: []Dimension{FToD(1.), FToD(number.Value)}}, nil
			}
		case "scale":
			if number, ok := args[0].(parser.NumberToken); ok {
				return SDimensions{String: "scale", Dimensions: []Dimension{FToD(number.Value), FToD(number.Value)}}, nil
			}
		}
	case 2:
		if name == "scale" && isAllNumber {
			return SDimensions{String: name, Dimensions: values}, nil
		}
		if name == "translate" && isAllLengths {
			return SDimensions{String: name, Dimensions: lengths}, nil
		}
	case 6:
		if name == "matrix" && isAllNumber {
			return SDimensions{String: name, Dimensions: values}, nil
		}
	}
	return SDimensions{}, InvalidValue
}

// Expand shorthand properties and filter unsupported properties and values.
// Log a warning for every ignored declaration.
// Return a iterable of ``(name, value, important)`` tuples.
//
func PreprocessDeclarations(baseUrl string, declarations []Token) []ValidatedProperty {
	var out []ValidatedProperty
	for _, _declaration := range declarations {
		if errToken, ok := _declaration.(parser.ParseError); ok {
			log.Printf("Error: %s \n", errToken.Message)
		}

		declaration, ok := _declaration.(parser.Declaration)
		if !ok {
			continue
		}

		name := declaration.Name.Lower()

		validationError := func(reason string) {
			log.Printf("Ignored `%s:%s` , %s. \n", declaration.Name, parser.Serialize(declaration.Value), reason)
		}

		if _, in := NotPrintMedia[name]; in {
			validationError("the property does not apply for the print media")
			continue
		}

		if strings.HasPrefix(name, prefix) {
			unprefixedName := strings.TrimPrefix(name, prefix)
			if _, in := proprietary[unprefixedName]; in {
				name = unprefixedName
			} else if _, in := unstable[unprefixedName]; in {
				log.Printf("Deprecated `%s:%s`, prefixes on unstable attributes are deprecated, use `%s` instead. \n",
					declaration.Name, parser.Serialize(declaration.Value), unprefixedName)
				name = unprefixedName
			} else {
				log.Printf("Ignored `%s:%s`,prefix on this attribute is not supported, use `%s` instead. \n",
					declaration.Name, parser.Serialize(declaration.Value), unprefixedName)
				continue
			}
		}

		expander_, in := expanders[name]
		if !in {
			expander_ = defaultValidateShorthand
		}

		tokens := RemoveWhitespace(declaration.Value)
		result, err := expander_(baseUrl, name, tokens)
		if err != nil {
			validationError(err.Error())
			continue
		}

		important := declaration.Important

		for _, np := range result {
			out = append(out, ValidatedProperty{
				Name:      strings.ReplaceAll(np.Name, "-", "_"),
				Value:     np.Property,
				Important: important,
			})
		}
	}
	return out
}

// Remove any top-level whitespace in a token list.
func RemoveWhitespace(tokens []Token) []Token {
	var out []Token
	for _, token := range tokens {
		if token.Type() != parser.TypeWhitespaceToken && token.Type() != parser.TypeComment {
			out = append(out, token)
		}
	}
	return out
}
