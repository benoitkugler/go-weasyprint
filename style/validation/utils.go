package validation

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/benoitkugler/go-weasyprint/structure/counters"
	. "github.com/benoitkugler/go-weasyprint/style/css"
	"github.com/benoitkugler/go-weasyprint/style/parser"
	"github.com/benoitkugler/go-weasyprint/utils"
)

var (
	// Default fallback values used in attr() functions
	attrFallbacks = map[string]AttrFallback{
		"string":  {Name: "string", Value: String("")},
		"color":   {Name: "ident", Value: String("currentcolor")},
		"url":     {Name: "external", Value: String("about:invalid")},
		"integer": {Name: "number", Value: Dimension{Unit: Scalar}.ToValue()},
		"number":  {Name: "number", Value: Dimension{Unit: Scalar}.ToValue()},
		"%":       {Name: "number", Value: Dimension{Unit: Scalar}.ToValue()},
	}
)

func init() {
	for unitString, unit := range LENGTHUNITS {
		attrFallbacks[unitString] = AttrFallback{Name: "length", Value: Dimension{Unit: unit}.ToValue()}
	}
	for unit := range ANGLETORADIANS {
		attrFallbacks[unit] = AttrFallback{Name: "angle", Value: Dimension{Unit: LENGTHUNITS[unit]}.ToValue()}
	}
}

// Split a list of tokens on commas, ie ``parser.LiteralToken(",")``.
//     Only "top-level" comma tokens are splitting points, not commas inside a
//     function or blocks.
//
func SplitOnComma(tokens []Token) [][]Token {
	var parts [][]Token
	var thisPart []Token
	for _, token := range tokens {
		litteral, ok := token.(parser.LiteralToken)
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

// Split a list of tokens on optional commas, ie ``LiteralToken(",")``.
func splitOnOptionalComma(tokens []Token) (parts []Token) {
	for _, splitPart := range SplitOnComma(tokens) {
		if len(splitPart) == 0 {
			// Happens when there"s a comma at the beginning, at the end, or
			// when two commas are next to each other.
			return
		}
		parts = append(parts, splitPart...)
	}
	return parts
}

// If ``token`` is a keyword, return its name. Otherwise return empty string.
func getCustomIdent(token Token) string {
	if ident, ok := token.(parser.IdentToken); ok {
		return string(ident.Value)
	}
	return ""
}

// Parse an <image> token.
func getImage(_token Token, baseUrl string) (Image, error) {
	token, ok := _token.(parser.FunctionBlock)
	if !ok {
		parsed, err := getUrl(_token, baseUrl)
		if err != nil {
			return nil, err
		}
		parsedUrl, ok := parsed.Content.(NamedString)
		if !ok {
			log.Fatalln("content should be an url here")
		}
		if parsedUrl.Name == "external" {
			return UrlImage(parsedUrl.String), nil
		}
		return nil, nil
	}
	arguments := SplitOnComma(RemoveWhitespace(*token.Arguments))
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
		if result.IsNone() {
			result.shape = "ellipse"
			result.size = GradientSize{Keyword: "farthest-corner"}
			result.position = Center{OriginX: "left", OriginY: "top", Pos: Point{fiftyPercent, fiftyPercent}}
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

func parseLinearGradientParameters(arguments [][]Token) (DirectionType, [][]Token) {
	firstArg := arguments[0]
	if len(firstArg) == 1 {
		angle, isNotNone := getAngle(firstArg[0])
		if isNotNone {
			return DirectionType{Angle: angle}, arguments[1:]
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
	return DirectionType{Angle: math.Pi}, arguments // Default direction is "to bottom"
}

func reverse(a []Token) []Token {
	n := len(a)
	out := make([]Token, n)
	for i := range a {
		out[n-1-i] = a[i]
	}
	return out
}

type radialGradientParameters struct {
	shape      string
	size       GradientSize
	position   Center
	colorStops [][]Token
}

func (r radialGradientParameters) IsNone() bool {
	return r.shape == "" && r.size.IsNone() && r.position.IsNone() && r.colorStops == nil
}

func parseRadialGradientParameters(arguments [][]Token) radialGradientParameters {
	var shape, sizeShape string
	var position Center
	var size GradientSize
	stack := reverse(arguments[0])
	for len(stack) > 0 {
		token := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		keyword := getKeyword(token)
		if keyword == "at" {
			position = parsePosition(reverse(stack))
			if position.IsNone() {
				return radialGradientParameters{}
			}
			break
		} else if (keyword == "circle" || keyword == "ellipse") && shape == "" {
			shape = keyword
		} else if (keyword == "closest-corner" || keyword == "farthest-corner" || keyword == "closest-side" || keyword == "farthest-side") && size.IsNone() {
			size = GradientSize{Keyword: keyword}
		} else {
			if len(stack) > 0 && size.IsNone() {
				length1 := getLength(token, true, true)
				length2 := getLength(stack[len(stack)-1], true, true)
				if !length1.IsNone() && !length2.IsNone() {
					size = GradientSize{Explicit: [2]Dimension{length1, length2}}
					sizeShape = "ellipse"
					i := utils.MaxInt(len(stack)-1, 0)
					stack = stack[:i]
				}
			}
			if size.IsNone() {
				length1 := getLength(token, true, false)
				if !length1.IsNone() {
					size = GradientSize{Explicit: [2]Dimension{length1, length1}}
					sizeShape = "circle"
				}
			}
			if size.IsNone() {
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
	if size.IsNone() {
		out.size = GradientSize{Keyword: "farthest-corner"}
	}
	if position.IsNone() {
		out.position = Center{
			OriginX: "left",
			OriginY: "top",
			Pos:     Point{fiftyPercent, fiftyPercent},
		}
	}
	return out
}

func parseColorStop(tokens []Token) (ColorStop, error) {
	switch len(tokens) {
	case 1:
		color := parser.ParseColor(tokens[0])
		if !color.IsNone() {
			return ColorStop{Color: Color(color)}, nil
		}
	case 2:
		color := parser.ParseColor(tokens[0])
		position := getLength(tokens[1], true, true)
		if !color.IsNone() && !position.IsNone() {
			return ColorStop{Color: Color(color), Position: position}, nil
		}
	}
	return ColorStop{}, InvalidValue
}

func getUrl(_token Token, baseUrl string) (out ContentProperty, err error) {
	switch token := _token.(type) {
	case parser.URLToken:
		if strings.HasPrefix(token.Value, "#") {
			return ContentProperty{Type: "url", Content: NamedString{Name: "internal", String: utils.Unquote(token.Value[1:])}}, nil
		} else {
			var joined string
			joined, err = utils.SafeUrljoin(baseUrl, token.Value, false)
			if err != nil {
				return
			}
			return ContentProperty{Type: "url", Content: NamedString{Name: "external", String: joined}}, nil
		}
	case parser.FunctionBlock:
		if token.Name == "attr" {
			return checkAttrFunction(token, "url"), nil
		}
	}
	return
}

func checkStringFunction(token Token) (out ContentProperty) {
	name, args := parseFunction(token)
	if name == "" {
		return
	}
	if name == "string" && (len(args) == 1 || len(args) == 2) {
		customIdent_, ok := args[0].(parser.IdentToken)
		args = args[1:]
		if !ok {
			return
		}
		customIdent := customIdent_.Value

		var ident string
		if len(args) > 0 {
			ident_ := args[0]
			args = args[1:]
			identToken, ok := ident_.(parser.IdentToken)
			val := identToken.Value.Lower()
			if !ok || (val != "first" && val != "start" && val != "last" && val != "first-except") {
				return
			}
			ident = val
		} else {
			ident = "first"
		}
		return ContentProperty{Type: "string()", Content: Strings{string(customIdent), ident}}
	}
	return
}

func CheckVarFunction(token Token) (outName string, content NamedTokens) {
	name, args := parseFunction(token)
	if name == "" {
		return
	}
	if name == "var" && len(args) > 0 {
		ident, ok := args[0].(parser.LiteralToken)
		args = args[1:]
		if !ok || !strings.HasPrefix(ident.Value, "--") {
			return
		}
		// TODO: we should check authorized tokens
		// https://drafts.csswg.org/css-syntax-3/#typedef-declaration-value
		v := strings.ReplaceAll(ident.Value, "-", "_")
		return "var()", NamedTokens{Name: v, Tokens: args}
	}
	return
}

// Parse functional notation.
//	Return ``(name, args)`` if the given token is a function with comma- or
//	space-separated arguments. Return zero values otherwise.
func parseFunction(functionToken_ Token) (string, []Token) {
	functionToken, ok := functionToken_.(parser.FunctionBlock)
	if !ok {
		return "", nil
	}
	content := RemoveWhitespace(*functionToken.Arguments)
	var (
		arguments []Token
		token     Token
	)
	lastIsComma := false
	for len(content) > 0 {
		token, content = content[0], content[1:]
		lit, ok := token.(parser.LiteralToken)
		isComma := ok && lit.Value == ","
		if lastIsComma && isComma {
			return "", nil
		}
		if isComma {
			lastIsComma = true
		} else {
			lastIsComma = false
			if fn, isFunc := token.(parser.FunctionBlock); isFunc {
				innerName, _ := parseFunction(fn)
				if innerName == "" {
					return "", nil
				}
			}
			arguments = append(arguments, token)
		}
	}
	if lastIsComma {
		return "", nil
	}
	return functionToken.Name.Lower(), arguments
}

func checkAttrFunction(token parser.FunctionBlock, allowedType string) (out ContentProperty) {
	name, args := parseFunction(token)
	if name == "" {
		return
	}
	la := len(args)
	if name == "attr" && (la == 1 || la == 2 || la == 3) {
		ident, ok := args[0].(parser.IdentToken)
		if !ok {
			return
		}
		attrName := ident.Value
		var (
			typeOrUnit string
			fallback   AttrFallback
		)
		if la == 1 {
			typeOrUnit = "string"
		} else {
			ident2, ok := args[0].(parser.IdentToken)
			if !ok {
				return
			}
			typeOrUnit = string(ident2.Value)
			fb, isIN := attrFallbacks[typeOrUnit]
			if !isIN {
				return
			}
			if la == 2 {
				fallback = fb
			} else {
				switch fbValue := args[2].(type) {
				case parser.StringToken:
					fallback = AttrFallback{Name: "string", Value: String(fbValue.Value)}
				default:
					// TODO: handle other fallback types
					return
				}
			}
		}
		if allowedType == "" || allowedType == typeOrUnit {
			return ContentProperty{Type: "attr()", Content: Attr{Name: string(attrName), TypeOrUnit: typeOrUnit, Fallback: fallback}}
		}
	}
	return
}

// Parse background-position and object-position.
//
// See http://dev.w3.org/csswg/css3-background/#the-background-position
// https://drafts.csswg.org/css-images-3/#propdef-object-position
func parsePosition(tokens []Token) Center {
	center := parse2dPosition(tokens)
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
					Pos:     Point{length1, length2},
				}
			}
			if (keyword2 == "left" || keyword2 == "right") && (keyword1 == "top" || keyword1 == "bottom") {
				return Center{OriginX: keyword2,
					OriginY: keyword1,
					Pos:     Point{length2, length1},
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
					return Center{OriginX: "left", OriginY: keyword, Pos: Point{fiftyPercent, length}}
				case "left", "right":
					return Center{OriginX: keyword, OriginY: "top", Pos: Point{length, fiftyPercent}}
				}
			case "top", "bottom":
				if keyword == "left" || keyword == "right" {
					return Center{OriginX: keyword, OriginY: otherKeyword, Pos: Point{length, ZEROPERCENT}}
				}
			case "left", "right":
				if keyword == "top" || keyword == "bottom" {
					return Center{OriginX: otherKeyword, OriginY: keyword, Pos: Point{ZEROPERCENT, length}}
				}
			}
		}
	}
	return Center{}
}

// Parse a <string> token.
func getString(_token Token) ContentProperty {
	switch token := _token.(type) {
	case parser.StringToken:
		return ContentProperty{Type: "string", Content: String(token.Value)}
	case parser.FunctionBlock:
		switch token.Name {

		case "attr":
			return checkAttrFunction(token, "string")
		case "counter", "counters":
			return checkCounterFunction(token)
		case "content":
			return checkContentFunction(token)
		case "string":
			return checkStringFunction(token)
		}
	}
	return ContentProperty{}
}

func checkCounterFunction(token Token) (out ContentProperty) {
	name, args := parseFunction(token)
	if name == "" {
		return
	}
	var arguments Strings
	la := len(args)
	if (name == "counter" && (la == 1 || la == 2)) || (name == "counters" && (la == 2 || la == 3)) {
		ident, ok := args[0].(parser.IdentToken)
		args = args[1:]
		if !ok {
			return
		}
		arguments = append(arguments, string(ident.Value))

		if name == "counters" {
			str, ok := args[0].(parser.StringToken)
			args = args[1:]
			if !ok {
				return
			}
			arguments = append(arguments, str.Value)
		}

		if len(args) > 0 {
			counterStyle := getKeyword(args[0])
			args = args[1:]
			if _, in := counters.STYLES[counterStyle]; counterStyle != "none" && !in {
				return
			}
			arguments = append(arguments, counterStyle)
		} else {
			arguments = append(arguments, "decimal")
		}

		return ContentProperty{Type: fmt.Sprintf("%s()", name), Content: arguments}
	}
	return
}

func checkContentFunction(token Token) (out ContentProperty) {
	name, args := parseFunction(token)
	if name == "" {
		return
	}
	if name == "content" {
		if len(args) == 0 {
			return ContentProperty{Type: "content()", Content: String("text")}
		} else if len(args) == 1 {
			ident, ok := args[0].(parser.IdentToken)
			v := ident.Value.Lower()
			if ok && (v == "text" || v == "before" || v == "after" || v == "first-letter" || v == "marker") {
				return ContentProperty{Type: "content()", Content: String(v)}
			}
		}
	}
	return
}

// Parse a <quote> token.
func getQuote(token Token) (bool, Quote) {
	keyword := getKeyword(token)
	return false, ContentQuoteKeywords[keyword]
}

// Parse a <target> token.
func getTarget(token Token, baseUrl string) (out ContentProperty, err error) {
	name, args := parseFunction(token)
	if name == "" {
		return
	}
	args = splitOnOptionalComma(args)
	la := len(args)
	if la > 0 {
		return
	}

	switch name {
	case "target-counter":
		if la != 2 && la != 3 {
			return
		}
	case "target-counters":
		if la != 3 && la != 4 {
			return
		}
	case "target-text":
		if la != 1 && la != 2 {
			return
		}
	default:
		return
	}

	var (
		values SContentProps
		value  SContentProp
	)

	link := args[0]
	args = args[1:]
	stringLink := getString(link)
	if stringLink.IsNone() {
		value.ContentProperty, err = getUrl(link, baseUrl)
		if err != nil {
			return
		}
		if value.ContentProperty.IsNone() {
			return
		}
		values = append(values, value)
	} else {
		values = append(values, SContentProp{ContentProperty: stringLink})
	}

	if strings.HasPrefix(name, "target-counter") {
		if len(args) == 0 {
			return
		}

		ident_ := args[0]
		args = args[1:]
		ident, ok := ident_.(parser.IdentToken)
		if !ok {
			return
		}
		values = append(values, SContentProp{String: string(ident.Value)})

		if name == "target-counters" {
			string_ := getString(args[0])
			args = args[1:]
			if string_.IsNone() {
				return
			}
			values = append(values, SContentProp{ContentProperty: string_})
		}

		var counterStyle string
		if len(args) > 0 {
			counterStyle = getKeyword(args[0])
			args = args[1:]
			if _, in := counters.STYLES[counterStyle]; !in {
				return
			}
		} else {
			counterStyle = "decimal"
		}
		values = append(values, SContentProp{String: counterStyle})
	} else {
		var content string
		if len(args) > 0 {
			content = getKeyword(args[0])
			args = args[1:]
			if content != "content" && content != "before" && content != "after" && content != "first-letter" {
				return
			}
		} else {
			content = "content"
		}
		values = append(values, SContentProp{String: content})
	}
	return ContentProperty{Type: fmt.Sprintf("%s()", name), Content: values}, nil
}

// Parse <content-list> tokens.
func getContentList(tokens []Token, baseUrl string) (out []ContentProperty, err error) {
	// See https://www.w3.org/TR/css-content-3/#typedef-content-list
	parsedTokens := make([]ContentProperty, len(tokens))
	for i, token := range tokens {
		parsedTokens[i], err = getContentListToken(token, baseUrl)
		if err != nil {
			return nil, err
		}
		if parsedTokens[i].IsNone() {
			return nil, nil
		}
	}
	return parsedTokens, nil
}

// Parse one of the <content-list> tokens.
func getContentListToken(token Token, baseUrl string) (ContentProperty, error) {
	// See https://www.w3.org/TR/css-content-3/#typedef-content-list

	// <string>
	string_ := getString(token)
	if !string_.IsNone() {
		return string_, nil
	}

	// contents
	if getKeyword(token) == "contents" {
		return ContentProperty{Type: "content", Content: String("text")}, nil
	}

	// <uri>
	url, err := getUrl(token, baseUrl)
	if err != nil {
		return ContentProperty{}, err
	}
	if !url.IsNone() {
		return url, nil
	}

	// <quote>
	notNone, quote := getQuote(token)
	if notNone {
		return ContentProperty{Type: "quote", Content: quote}, nil
	}

	// <target>
	target, err := getTarget(token, baseUrl)
	if err != nil || !target.IsNone() {
		return target, err
	}

	// <leader>
	name, args := parseFunction(token)
	if name == "" {
		return ContentProperty{}, nil
	}
	if name == "leader" {
		if len(args) != 1 {
			return ContentProperty{}, nil
		}
		arg_ := args[0]
		var str string
		switch arg := arg_.(type) {
		case parser.IdentToken:
			switch arg.Value {
			case "dotted":
				str = "."
			case "solid":
				str = ""
			case "space":
				str = " "
			default:
				return ContentProperty{}, nil
			}
		case parser.StringToken:
			str = arg.Value
		}
		return ContentProperty{Type: "leader()", Content: Strings{"string", str}}, nil
	}
	return ContentProperty{}, nil
}
