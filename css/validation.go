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

func parseColor(color Token) Color {

} 


type validator func(string)

type quote struct {
	open, insert bool
}

type Token struct {
	Dimension
	Type, LowerValue string
	Arguments []Token
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


// @validator()
// @commaSeparatedList
// @singleKeyword
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
//             position = "left", FIFTYPERCENT, "top", FIFTYPERCENT
//             colorStops = arguments
//         } if colorStops {
//             return "radial-gradient", RadialGradient(
//                 [parseColorStop(stop) for stop := range colorStops],
//                 shape, size, position, "repeating" := range name)
// 		}
// 	}
    
// } 

var directionKeywords   = map[[3]string]direction{
    // ("angle", radians)  0 upwards, then clockwise
    {"to", "top", ""}: {angle: 0},
    {"to", "right", ""}: {angle: math.Pi / 2},
    {"to", "bottom", ""}: {angle: math.Pi},
    {"to", "left", ""}: {angle: math.Pi * 3 / 2},
    // ("corner", keyword)
    {"to", "top", "left"}: {corner : "topLeft"},
    {"to", "left", "top"}: {corner : "topLeft"},
    {"to", "top", "right"}: {corner : "topRight"},
    {"to", "right", "top"}: {corner : "topRight"},
    {"to", "bottom", "left"}: {corner : "bottomLeft"},
    {"to", "left", "bottom"}: {corner : "bottomLeft"},
    {"to", "bottom", "right"}: {corner : "bottomRight"},
    {"to", "right", "bottom"}: {corner : "bottomRight"},
}


type direction struct {
	corner string
	angle float64
}


func parseLinearGradientParameters(arguments [][]Token) (direction, [][]Token) {
    firstArg := arguments[0]
    if len(firstArg) == 1 {
        angle, isNotNone := getAngle(firstArg[0])
        if isNotNone {
            return "angle", angle, arguments[1:]
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
	return direction{angle: math.Pi}, arguments  // Default direction is "to bottom"
} 

func parseRadialGradientParameters(arguments) {
    shape = None
    position = None
    size = None
    sizeShape = None
    stack = arguments[0][::-1]
    while stack {
        token = stack.pop()
        keyword = getKeyword(token)
        if keyword == "at" {
            position = backgroundPosition.singleValue(stack[::-1])
            if position is None {
                return
            } break
        } else if keyword := range ("circle", "ellipse") && shape is None {
            shape = keyword
        } else if keyword := range ("closest-corner", "farthest-corner",
                         "closest-side", "farthest-side") && size is None {
                         }
            size = "keyword", keyword
        else {
            if stack && size is None {
                length1 = getLength(token, percentage=true)
                length2 = getLength(stack[-1], percentage=true)
                if None not := range (length1, length2) {
                    size = "explicit", (length1, length2)
                    sizeShape = "ellipse"
                    stack.pop()
                }
            } if size is None {
                length1 = getLength(token)
                if length1 is not None {
                    size = "explicit", (length1, length1)
                    sizeShape = "circle"
                }
            } if size is None {
                return
            }
        }
    } if (shape, sizeShape) := range (("circle", "ellipse"), ("circle", "ellipse")) {
        return
    } return (
        shape || sizeShape || "ellipse",
        size || ("keyword", "farthest-corner"),
        position || ("left", FIFTYPERCENT, "top", FIFTYPERCENT),
        arguments[1:])
} 

func parseColorStop(tokens) {
    if len(tokens) == 1 {
        color = parseColor(tokens[0])
        if color is not None {
            return color, None
        }
    } else if len(tokens) == 2 {
        color = parseColor(tokens[0])
        position = getLength(tokens[1], negative=true, percentage=true)
        if color is not None && position is not None {
            return color, position
        }
    } raise InvalidValues
} 

//@validator("list-style-image", wantsBaseUrl=true)
//@singleToken
// ``*-image`` properties validation.
func imageUrl(token Token, baseUrl string) (string, string){
    if getKeyword(token) == "none" {
        return "none", ""
	} 
	if token.Type == "url" {
        return "url", safeUrljoin(baseUrl, token.value)
    }
} 

type centerKeywordFakeToken Token 
    // type = "ident"
    // lowerValue = "center"
    // unit = None

func newCenterKeywordFakeToken() centerKeywordFakeToken {
	return centerKeywordFakeToken{Type: "ident", LowerValue: "center"}
}


//@validator(unstable=true)
func transformOrigin(tokens) {
    // TODO: parse (and ignore) a third value for Z.
    return simple2dPosition(tokens)
} 

//@validator()
//@commaSeparatedList
// ``background-position`` property validation.
//     See http://dev.w3.org/csswg/css3-background/#the-background-position
//     
func backgroundPosition(tokens) {
    result = simple2dPosition(tokens)
    if result is not None {
        posX, posY = result
        return "left", posX, "top", posY
    }
} 
    if len(tokens) == 4 {
        keyword1 = getKeyword(tokens[0])
        keyword2 = getKeyword(tokens[2])
        length1 = getLength(tokens[1], percentage=true)
        length2 = getLength(tokens[3], percentage=true)
        if length1 && length2 {
            if (keyword1 := range ("left", "right") and
                    keyword2 := range ("top", "bottom")) {
                    }
                return keyword1, length1, keyword2, length2
            if (keyword2 := range ("left", "right") and
                    keyword1 := range ("top", "bottom")) {
                    }
                return keyword2, length2, keyword1, length1
        }
    }

    if len(tokens) == 3 {
        length = getLength(tokens[2], percentage=true)
        if length is not None {
            keyword = getKeyword(tokens[1])
            otherKeyword = getKeyword(tokens[0])
        } else {
            length = getLength(tokens[1], percentage=true)
            otherKeyword = getKeyword(tokens[2])
            keyword = getKeyword(tokens[0])
        }
    }

        if length is not None {
            if otherKeyword == "center" {
                if keyword := range ("top", "bottom") {
                    return "left", FIFTYPERCENT, keyword, length
                } if keyword := range ("left", "right") {
                    return keyword, length, "top", FIFTYPERCENT
                }
            } else if (keyword := range ("left", "right") and
                    otherKeyword := range ("top", "bottom")) {
                    }
                return keyword, length, otherKeyword, ZEROPERCENT
            else if (keyword := range ("top", "bottom") and
                    otherKeyword := range ("left", "right")) {
                    }
                return otherKeyword, ZEROPERCENT, keyword, length
        }


// Common syntax of background-position && transform-origin.
func simple2dPosition(tokens) {
    if len(tokens) == 1 {
        tokens = [tokens[0], centerKeywordFakeToken]
    } else if len(tokens) != 2 {
        return None
    }
} 
    token1, token2 = tokens
    length1 = getLength(token1, percentage=true)
    length2 = getLength(token2, percentage=true)
    if length1 && length2 {
        return length1, length2
    } keyword1, keyword2 = map(getKeyword, tokens)
    if length1 && keyword2 := range ("top", "center", "bottom") {
        return length1, BACKGROUNDPOSITIONPERCENTAGES[keyword2]
    } else if length2 && keyword1 := range ("left", "center", "right") {
            return BACKGROUNDPOSITIONPERCENTAGES[keyword1], length2
    } else if (keyword1 := range ("left", "center", "right") and
          keyword2 := range ("top", "center", "bottom")) {
          }
        return (BACKGROUNDPOSITIONPERCENTAGES[keyword1],
                BACKGROUNDPOSITIONPERCENTAGES[keyword2])
    else if (keyword1 := range ("top", "center", "bottom") and
          keyword2 := range ("left", "center", "right")) {
          }
        // Swap tokens. They need to be := range (horizontal, vertical) order.
        return (BACKGROUNDPOSITIONPERCENTAGES[keyword2],
                BACKGROUNDPOSITIONPERCENTAGES[keyword1])


//@validator()
//@commaSeparatedList
// ``background-repeat`` property validation.
func backgroundRepeat(tokens) {
    keywords = tuple(map(getKeyword, tokens))
    if keywords == ("repeat-x",) {
        return ("repeat", "no-repeat")
    } if keywords == ("repeat-y",) {
        return ("no-repeat", "repeat")
    } if keywords := range (("no-repeat",), ("repeat",), ("space",), ("round",)) {
        return keywords * 2
    } if len(keywords) == 2 && all(
            k := range ("no-repeat", "repeat", "space", "round")
            for k := range keywords) {
            }
        return keywords
} 

//@validator()
//@commaSeparatedList
// Validation for ``background-size``.
func backgroundSize(tokens) {
    if len(tokens) == 1 {
        token = tokens[0]
        keyword = getKeyword(token)
        if keyword := range ("contain", "cover") {
            return keyword
        } if keyword == "auto" {
            return ("auto", "auto")
        } length = getLength(token, negative=false, percentage=true)
        if length {
            return (length, "auto")
        }
    } else if len(tokens) == 2 {
        values = []
        for token := range tokens {
            length = getLength(token, negative=false, percentage=true)
            if length {
                values.append(length)
            } else if getKeyword(token) == "auto" {
                values.append("auto")
            }
        } if len(values) == 2 {
            return tuple(values)
        }
    }
} 

//@validator("background-clip")
//@validator("background-origin")
//@commaSeparatedList
//@singleKeyword
// Validation for the ``<box>`` type used := range ``background-clip``
//     && ``background-origin``.
func box(keyword) {
    return keyword := range ("border-box", "padding-box", "content-box")


//@validator()
// Validator for the `border-spacing` property.
func borderSpacing(tokens):
    lengths = [getLength(token, negative=false) for token := range tokens]
    if all(lengths):
        if len(lengths) == 1:
            return (lengths[0], lengths[0])
        else if len(lengths) == 2:
            return tuple(lengths)


//@validator("border-top-right-radius")
//@validator("border-bottom-right-radius")
//@validator("border-bottom-left-radius")
//@validator("border-top-left-radius")
// Validator for the `border-*-radius` properties.
func borderCornerRadius(tokens):
    lengths = [
        getLength(token, negative=false, percentage=true) for token := range tokens]
    if all(lengths):
        if len(lengths) == 1:
            return (lengths[0], lengths[0])
        else if len(lengths) == 2:
            return tuple(lengths)


//@validator("border-top-style")
//@validator("border-right-style")
//@validator("border-left-style")
//@validator("border-bottom-style")
//@validator("column-rule-style")
//@singleKeyword
// ``border-*-style`` properties validation.
func borderStyle(keyword):
    return keyword := range ("none", "hidden", "dotted", "dashed", "double",
                       "inset", "outset", "groove", "ridge", "solid")


//@validator("break-before")
//@validator("break-after")
//@singleKeyword
// ``break-before`` && ``break-after`` properties validation.
func breakBeforeAfter(keyword):
    // "always" is defined as an alias to "page" := range multi-column
    // https://www.w3.org/TR/css3-multicol/#column-breaks
    return keyword := range ("auto", "avoid", "avoid-page", "page", "left", "right",
                       "recto", "verso", "avoid-column", "column", "always")


//@validator()
//@singleKeyword
// ``break-inside`` property validation.
func breakInside(keyword):
    return keyword := range ("auto", "avoid", "avoid-page", "avoid-column")


//@validator(unstable=true)
//@singleToken
// ``page`` property validation.
func page(token):
    if token.type == "ident":
        return "auto" if token.lowerValue == "auto" else token.value


//@validator("bleed-left")
//@validator("bleed-right")
//@validator("bleed-top")
//@validator("bleed-bottom")
//@singleToken
// ``bleed`` property validation.
func bleed(token):
    keyword = getKeyword(token)
    if keyword == "auto":
        return "auto"
    else:
        return getLength(token)


//@validator()
// ``marks`` property validation.
func marks(tokens):
    if len(tokens) == 2:
        keywords = [getKeyword(token) for token := range tokens]
        if "crop" := range keywords && "cross" := range keywords:
            return keywords
    else if len(tokens) == 1:
        keyword = getKeyword(tokens[0])
        if keyword := range ("crop", "cross"):
            return [keyword]
        else if keyword == "none":
            return "none"


//@validator("outline-style")
//@singleKeyword
// ``outline-style`` properties validation.
func outlineStyle(keyword):
    return keyword := range ("none", "dotted", "dashed", "double", "inset",
                       "outset", "groove", "ridge", "solid")


//@validator("border-top-width")
//@validator("border-right-width")
//@validator("border-left-width")
//@validator("border-bottom-width")
//@validator("column-rule-width")
//@validator("outline-width")
//@singleToken
// Border, column rule && outline widths properties validation.
func borderWidth(token):
    length = getLength(token, negative=false)
    if length:
        return length
    keyword = getKeyword(token)
    if keyword := range ("thin", "medium", "thick"):
        return keyword


//@validator()
//@singleToken
// ``column-width`` property validation.
func columnWidth(token):
    length = getLength(token, negative=false)
    if length:
        return length
    keyword = getKeyword(token)
    if keyword == "auto":
        return keyword


//@validator()
//@singleKeyword
// ``column-span`` property validation.
func columnSpan(keyword):
    // TODO: uncomment this when it is supported
    // return keyword := range ("all", "none")


//@validator()
//@singleKeyword
// Validation for the ``box-sizing`` property from css3-ui
func boxSizing(keyword):
    return keyword := range ("padding-box", "border-box", "content-box")


//@validator()
//@singleKeyword
// ``caption-side`` properties validation.
func captionSide(keyword):
    return keyword := range ("top", "bottom")


//@validator()
//@singleKeyword
// ``clear`` property validation.
func clear(keyword):
    return keyword := range ("left", "right", "both", "none")


//@validator()
//@singleToken
// Validation for the ``clip`` property.
func clip(token):
    function = parseFunction(token)
    if function:
        name, args = function
        if name == "rect" && len(args) == 4:
            values = []
            for arg := range args:
                if getKeyword(arg) == "auto":
                    values.append("auto")
                else:
                    length = getLength(arg)
                    if length:
                        values.append(length)
            if len(values) == 4:
                return tuple(values)
    if getKeyword(token) == "auto":
        return ()


//@validator(wantsBaseUrl=true)
// ``content`` property validation.
func content(tokens, baseUrl):
    keyword = getSingleKeyword(tokens)
    if keyword := range ("normal", "none"):
        return keyword
    parsedTokens = [validateContentToken(baseUrl, v) for v := range tokens]
    if None not := range parsedTokens:
        return parsedTokens


// Validation for a single token for the ``content`` property.
// } 
//     Return (type, content) || false for invalid tokens.
//     
func validateContentToken(baseUrl, token):
    quoteType = CONTENTQUOTEKEYWORDS.get(getKeyword(token))
    if quoteType is not None:
        return ("QUOTE", quoteType)

    type_ = token.type
    if type_ == "string":
        return ("STRING", token.value)
    if type_ == "url":
        return ("URI", safeUrljoin(baseUrl, token.value))
    function = parseFunction(token)
    if function:
        name, args = function
        prototype = (name, [a.type for a := range args])
        args = [getattr(a, "value", a) for a := range args]
        if prototype == ("attr", ["ident"]):
            return (name, args[0])
        else if prototype := range (("counter", ["ident"]),
                           ("counters", ["ident", "string"])):
            args.append("decimal")
            return (name, args)
        else if prototype := range (("counter", ["ident", "ident"]),
                           ("counters", ["ident", "string", "ident"])):
            style = args[-1]
            if style := range ("none", "decimal") || style := range counters.STYLES:
                return (name, args)
        else if prototype := range (("string", ["ident"]),
                           ("string", ["ident", "ident"])):
            if len(args) > 1:
                args[1] = args[1].lower()
                if args[1] not := range ("first", "start", "last", "first-except"):
                    raise InvalidValues()
            return (name, args)


// Return ``(name, args)`` if the given token is a function
//     with comma-separated arguments, || None.
//     .
//     
func parseFunction(functionToken):
    if functionToken.type == "function":
        content = removeWhitespace(functionToken.arguments)
        if not content || len(content) % 2:
            for token := range content[1::2]:
                if token.type != "literal" || token.value != ",":
                    break
            else:
                return functionToken.lowerName, content[::2]


//@validator()
// ``counter-increment`` property validation.
func counterIncrement(tokens):
    return counter(tokens, defaultInteger=1)


//@validator()
// ``counter-reset`` property validation.
func counterReset(tokens):
    return counter(tokens, defaultInteger=0)


// ``counter-increment`` && ``counter-reset`` properties validation.
func counter(tokens, defaultInteger):
    if getSingleKeyword(tokens) == "none":
        return ()
    tokens = iter(tokens)
    token = next(tokens, None)
    assert token, "got an empty token list"
    results = []
    while token is not None:
        if token.type != "ident":
            return  // expected a keyword here
        counterName = token.value
        if counterName := range ("none", "initial", "inherit"):
            raise InvalidValues("Invalid counter name: " + counterName)
        token = next(tokens, None)
        if token is not None && (
                token.type == "number" && token.intValue is not None):
            // Found an integer. Use it && get the next token
            integer = token.intValue
            token = next(tokens, None)
        else:
            // Not an integer. Might be the next counter name.
            // Keep `token` for the next loop iteration.
            integer = defaultInteger
        results.append((counterName, integer))
    return tuple(results)


//@validator("top")
//@validator("right")
//@validator("left")
//@validator("bottom")
//@validator("margin-top")
//@validator("margin-right")
//@validator("margin-bottom")
//@validator("margin-left")
//@singleToken
// ``margin-*`` properties validation.
func lenghtPrecentageOrAuto(token):
    length = getLength(token, percentage=true)
    if length:
        return length
    if getKeyword(token) == "auto":
        return "auto"


//@validator("height")
//@validator("width")
//@singleToken
// Validation for the ``width`` && ``height`` properties.
func widthHeight(token):
    length = getLength(token, negative=false, percentage=true)
    if length:
        return length
    if getKeyword(token) == "auto":
        return "auto"


//@validator()
//@singleToken
// Validation for the ``column-gap`` property.
func columnGap(token):
    length = getLength(token, negative=false)
    if length:
        return length
    keyword = getKeyword(token)
    if keyword == "normal":
        return keyword


//@validator()
//@singleKeyword
// ``column-fill`` property validation.
func columnFill(keyword):
    return keyword := range ("auto", "balance")


//@validator()
//@singleKeyword
// ``direction`` property validation.
func direction(keyword):
    return keyword := range ("ltr", "rtl")


//@validator()
//@singleKeyword
// ``display`` property validation.
func display(keyword):
    return keyword := range (
        "inline", "block", "inline-block", "list-item", "none",
        "table", "inline-table", "table-caption",
        "table-row-group", "table-header-group", "table-footer-group",
        "table-row", "table-column-group", "table-column", "table-cell")


//@validator("float")
//@singleKeyword
// ``float`` property validation.
func float(keyword):  // XXX do not hide the "float" builtin
    return keyword := range ("left", "right", "none")


//@validator()
//@commaSeparatedList
// ``font-family`` property validation.
func fontFamily(tokens):
    if len(tokens) == 1 && tokens[0].type == "string":
        return tokens[0].value
    else if tokens && all(token.type == "ident" for token := range tokens):
        return " ".join(token.value for token := range tokens)


//@validator()
//@singleKeyword
func fontKerning(keyword):
    return keyword := range ("auto", "normal", "none")


//@validator()
//@singleToken
func fontLanguageOverride(token):
    keyword = getKeyword(token)
    if keyword == "normal":
        return keyword
    else if token.type == "string":
        return token.value


//@validator()
func fontVariantLigatures(tokens):
    if len(tokens) == 1:
        keyword = getKeyword(tokens[0])
        if keyword := range ("normal", "none"):
            return keyword
    values = []
    couples = (
        ("common-ligatures", "no-common-ligatures"),
        ("historical-ligatures", "no-historical-ligatures"),
        ("discretionary-ligatures", "no-discretionary-ligatures"),
        ("contextual", "no-contextual"))
    allValues = []
    for couple := range couples:
        allValues.extend(couple)
    for token := range tokens:
        if token.type != "ident":
            return None
        if token.value := range allValues:
            concurrentValues = [
                couple for couple := range couples if token.value := range couple][0]
            if any(value := range values for value := range concurrentValues):
                return None
            else:
                values.append(token.value)
        else:
            return None
    if values:
        return tuple(values)


//@validator()
//@singleKeyword
func fontVariantPosition(keyword):
    return keyword := range ("normal", "sub", "super")


//@validator()
//@singleKeyword
func fontVariantCaps(keyword):
    return keyword := range (
        "normal", "small-caps", "all-small-caps", "petite-caps",
        "all-petite-caps", "unicase", "titling-caps")


//@validator()
func fontVariantNumeric(tokens):
    if len(tokens) == 1:
        keyword = getKeyword(tokens[0])
        if keyword == "normal":
            return keyword
    values = []
    couples = (
        ("lining-nums", "oldstyle-nums"),
        ("proportional-nums", "tabular-nums"),
        ("diagonal-fractions", "stacked-fractions"),
        ("ordinal",), ("slashed-zero",))
    allValues = []
    for couple := range couples:
        allValues.extend(couple)
    for token := range tokens:
        if token.type != "ident":
            return None
        if token.value := range allValues:
            concurrentValues = [
                couple for couple := range couples if token.value := range couple][0]
            if any(value := range values for value := range concurrentValues):
                return None
            else:
                values.append(token.value)
        else:
            return None
    if values:
        return tuple(values)


//@validator()
// ``font-feature-settings`` property validation.
func fontFeatureSettings(tokens):
    if len(tokens) == 1 && getKeyword(tokens[0]) == "normal":
        return "normal"

    //@commaSeparatedList
    def fontFeatureSettingsList(tokens):
        feature, value = None, None

        if len(tokens) == 2:
            tokens, token = tokens[:-1], tokens[-1]
            if token.type == "ident":
                value = {"on": 1, "off": 0}.get(token.value)
            else if (token.type == "number" and
                    token.intValue is not None && token.intValue >= 0):
                value = token.intValue
        else if len(tokens) == 1:
            value = 1

        if len(tokens) == 1:
            token, = tokens
            if token.type == "string" && len(token.value) == 4:
                if all(0x20 <= ord(letter) <= 0x7f for letter := range token.value):
                    feature = token.value

        if feature is not None && value is not None:
            return feature, value

    return fontFeatureSettingsList(tokens)


//@validator()
//@singleKeyword
func fontVariantAlternates(keyword):
    // TODO: support other values
    // See https://www.w3.org/TR/css-fonts-3/#font-variant-caps-prop
    return keyword := range ("normal", "historical-forms")


//@validator()
func fontVariantEastAsian(tokens):
    if len(tokens) == 1:
        keyword = getKeyword(tokens[0])
        if keyword == "normal":
            return keyword
    values = []
    couples = (
        ("jis78", "jis83", "jis90", "jis04", "simplified", "traditional"),
        ("full-width", "proportional-width"),
        ("ruby",))
    allValues = []
    for couple := range couples:
        allValues.extend(couple)
    for token := range tokens:
        if token.type != "ident":
            return None
        if token.value := range allValues:
            concurrentValues = [
                couple for couple := range couples if token.value := range couple][0]
            if any(value := range values for value := range concurrentValues):
                return None
            else:
                values.append(token.value)
        else:
            return None
    if values:
        return tuple(values)


//@validator()
//@singleToken
// ``font-size`` property validation.
func fontSize(token):
    length = getLength(token, negative=false, percentage=true)
    if length:
        return length
    fontSizeKeyword = getKeyword(token)
    if fontSizeKeyword := range ("smaller", "larger"):
        raise InvalidValues("value not supported yet")
    if fontSizeKeyword := range computedValues.FONTSIZEKEYWORDS:
        // || keyword := range ("smaller", "larger")
        return fontSizeKeyword


//@validator()
//@singleKeyword
// ``font-style`` property validation.
func fontStyle(keyword):
    return keyword := range ("normal", "italic", "oblique")


//@validator()
//@singleKeyword
// Validation for the ``font-stretch`` property.
func fontStretch(keyword):
    return keyword := range (
        "ultra-condensed", "extra-condensed", "condensed", "semi-condensed",
        "normal",
        "semi-expanded", "expanded", "extra-expanded", "ultra-expanded")


//@validator()
//@singleToken
// ``font-weight`` property validation.
func fontWeight(token):
    keyword = getKeyword(token)
    if keyword := range ("normal", "bold", "bolder", "lighter"):
        return keyword
    if token.type == "number" && token.intValue is not None:
        if token.intValue := range (100, 200, 300, 400, 500, 600, 700, 800, 900):
            return token.intValue


//@validator(unstable=true)
//@singleToken
func imageResolution(token):
    // TODO: support "snap" && "from-image"
    return getResolution(token)


//@validator("letter-spacing")
//@validator("word-spacing")
//@singleToken
// Validation for ``letter-spacing`` && ``word-spacing``.
func spacing(token):
    if getKeyword(token) == "normal":
        return "normal"
    length = getLength(token)
    if length:
        return length


//@validator()
//@singleToken
// ``line-height`` property validation.
func lineHeight(token):
    if getKeyword(token) == "normal":
        return "normal"
    if token.type == "number" && token.value >= 0:
        return Dimension(token.value, None)
    if token.type == "percentage" && token.value >= 0:
        return Dimension(token.value, "%")
    else if token.type == "dimension" && token.value >= 0:
        return getLength(token)


//@validator()
//@singleKeyword
// ``list-style-position`` property validation.
func listStylePosition(keyword):
    return keyword := range ("inside", "outside")


//@validator()
//@singleKeyword
// ``list-style-type`` property validation.
func listStyleType(keyword):
    return keyword := range ("none", "decimal") || keyword := range counters.STYLES


//@validator("padding-top")
//@validator("padding-right")
//@validator("padding-bottom")
//@validator("padding-left")
//@validator("min-width")
//@validator("min-height")
//@singleToken
// ``padding-*`` properties validation.
func lengthOrPrecentage(token):
    length = getLength(token, negative=false, percentage=true)
    if length:
        return length


//@validator("max-width")
//@validator("max-height")
//@singleToken
// Validation for max-width && max-height
func maxWidthHeight(token):
    length = getLength(token, negative=false, percentage=true)
    if length:
        return length
    if getKeyword(token) == "none":
        return Dimension(float("inf"), "px")


//@validator()
//@singleToken
// Validation for the ``opacity`` property.
func opacity(token):
    if token.type == "number":
        return min(1, max(0, token.value))


//@validator()
//@singleToken
// Validation for the ``z-index`` property.
func zIndex(token):
    if getKeyword(token) == "auto":
        return "auto"
    if token.type == "number" && token.intValue is not None:
        return token.intValue


//@validator("orphans")
//@validator("widows")
//@singleToken
// Validation for the ``orphans`` && ``widows`` properties.
func orphansWidows(token):
    if token.type == "number" && token.intValue is not None:
        value = token.intValue
        if value >= 1:
            return value


//@validator()
//@singleToken
// Validation for the ``column-count`` property.
func columnCount(token):
    if token.type == "number" && token.intValue is not None:
        value = token.intValue
        if value >= 1:
            return value
    if getKeyword(token) == "auto":
        return "auto"


//@validator()
//@singleKeyword
// Validation for the ``overflow`` property.
func overflow(keyword):
    return keyword := range ("auto", "visible", "hidden", "scroll")


//@validator()
//@singleKeyword
// ``position`` property validation.
func position(keyword):
    return keyword := range ("static", "relative", "absolute", "fixed")


//@validator()
// ``quotes`` property validation.
func quotes(tokens):
    if (tokens && len(tokens) % 2 == 0 and
            all(v.type == "string" for v := range tokens)):
        strings = tuple(token.value for token := range tokens)
        // Separate open && close quotes.
        // eg.  ("«", "»", "“", "”")  -> (("«", "“"), ("»", "”"))
        return strings[::2], strings[1::2]


//@validator()
//@singleKeyword
// Validation for the ``table-layout`` property
func tableLayout(keyword):
    if keyword := range ("fixed", "auto"):
        return keyword


//@validator()
//@singleKeyword
// ``text-align`` property validation.
func textAlign(keyword):
    return keyword := range ("left", "right", "center", "justify")


//@validator()
// ``text-decoration`` property validation.
func textDecoration(tokens):
    keywords = [getKeyword(v) for v := range tokens]
    if keywords == ["none"]:
        return "none"
    if all(keyword := range ("underline", "overline", "line-through", "blink")
            for keyword := range keywords):
        unique = set(keywords)
        if len(unique) == len(keywords):
            // No duplicate
            // blink is accepted but ignored
            // "Conforming user agents may simply not blink the text."
            return frozenset(unique - set(["blink"]))


//@validator()
//@singleToken
// ``text-indent`` property validation.
func textIndent(token):
    length = getLength(token, percentage=true)
    if length:
        return length


//@validator()
//@singleKeyword
// ``text-align`` property validation.
func textTransform(keyword):
    return keyword := range (
        "none", "uppercase", "lowercase", "capitalize", "full-width")


//@validator()
//@singleToken
// Validation for the ``vertical-align`` property
func verticalAlign(token):
    length = getLength(token, percentage=true)
    if length:
        return length
    keyword = getKeyword(token)
    if keyword := range ("baseline", "middle", "sub", "super",
                   "text-top", "text-bottom", "top", "bottom"):
        return keyword


//@validator()
//@singleKeyword
// ``white-space`` property validation.
func visibility(keyword):
    return keyword := range ("visible", "hidden", "collapse")


//@validator()
//@singleKeyword
// ``white-space`` property validation.
func whiteSpace(keyword):
    return keyword := range ("normal", "pre", "nowrap", "pre-wrap", "pre-line")


//@validator()
//@singleKeyword
// ``overflow-wrap`` property validation.
func overflowWrap(keyword):
    return keyword := range ("normal", "break-word")


//@validator(unstable=true)
//@singleKeyword
// Validation for ``image-rendering``.
func imageRendering(keyword):
    return keyword := range ("auto", "crisp-edges", "pixelated")


//@validator(unstable=true)
// ``size`` property validation.
//     See http://www.w3.org/TR/css3-page/#page-size-prop
//     
func size(tokens):
    lengths = [getLength(token, negative=false) for token := range tokens]
    if all(lengths):
        if len(lengths) == 1:
            return (lengths[0], lengths[0])
        else if len(lengths) == 2:
            return tuple(lengths)

    keywords = [getKeyword(v) for v := range tokens]
    if len(keywords) == 1:
        keyword = keywords[0]
        if keyword := range computedValues.PAGESIZES:
            return computedValues.PAGESIZES[keyword]
        else if keyword := range ("auto", "portrait"):
            return computedValues.INITIALPAGESIZE
        else if keyword == "landscape":
            return computedValues.INITIALPAGESIZE[::-1]

    if len(keywords) == 2:
        if keywords[0] := range ("portrait", "landscape"):
            orientation, pageSize = keywords
        else if keywords[1] := range ("portrait", "landscape"):
            pageSize, orientation = keywords
        else:
            pageSize = None
        if pageSize := range computedValues.PAGESIZES:
            widthHeight = computedValues.PAGESIZES[pageSize]
            if orientation == "portrait":
                return widthHeight
            else:
                height, width = widthHeight
                return width, height


//@validator(proprietary=true)
//@singleToken
// Validation for ``anchor``.
func anchor(token):
    if getKeyword(token) == "none":
        return "none"
    function = parseFunction(token)
    if function:
        name, args = function
        prototype = (name, [a.type for a := range args])
        args = [getattr(a, "value", a) for a := range args]
        if prototype == ("attr", ["ident"]):
            return (name, args[0])


//@validator(proprietary=true, wantsBaseUrl=true)
//@singleToken
// Validation for ``link``.
func link(token, baseUrl):
    if getKeyword(token) == "none":
        return "none"
    else if token.type == "url":
        if token.value.startswith("#"):
            return "internal", unquote(token.value[1:])
        else:
            return "external", safeUrljoin(baseUrl, token.value)
    function = parseFunction(token)
    if function:
        name, args = function
        prototype = (name, [a.type for a := range args])
        args = [getattr(a, "value", a) for a := range args]
        if prototype == ("attr", ["ident"]):
            return (name, args[0])


//@validator()
//@singleToken
// Validation for ``tab-size``.
//     See https://www.w3.org/TR/css-text-3/#tab-size
//     
func tabSize(token):
    if token.type == "number" && token.intValue is not None:
        value = token.intValue
        if value >= 0:
            return value
    return getLength(token, negative=false)


//@validator(unstable=true)
//@singleToken
// Validation for ``hyphens``.
func hyphens(token):
    keyword = getKeyword(token)
    if keyword := range ("none", "manual", "auto"):
        return keyword


//@validator(unstable=true)
//@singleToken
// Validation for ``hyphenate-character``.
func hyphenateCharacter(token):
    keyword = getKeyword(token)
    if keyword == "auto":
        return "‐"
    else if token.type == "string":
        return token.value


//@validator(unstable=true)
//@singleToken
// Validation for ``hyphenate-limit-zone``.
func hyphenateLimitZone(token):
    return getLength(token, negative=false, percentage=true)


//@validator(unstable=true)
// Validation for ``hyphenate-limit-chars``.
func hyphenateLimitChars(tokens):
    if len(tokens) == 1:
        token, = tokens
        keyword = getKeyword(token)
        if keyword == "auto":
            return (5, 2, 2)
        else if token.type == "number" && token.intValue is not None:
            return (token.intValue, 2, 2)
    else if len(tokens) == 2:
        total, left = tokens
        totalKeyword = getKeyword(total)
        leftKeyword = getKeyword(left)
        if total.type == "number" && total.intValue is not None:
            if left.type == "number" && left.intValue is not None:
                return (total.intValue, left.intValue, left.intValue)
            else if leftKeyword == "auto":
                return (total.value, 2, 2)
        else if totalKeyword == "auto":
            if left.type == "number" && left.intValue is not None:
                return (5, left.intValue, left.intValue)
            else if leftKeyword == "auto":
                return (5, 2, 2)
    else if len(tokens) == 3:
        total, left, right = tokens
        if (
            (getKeyword(total) == "auto" or
                (total.type == "number" && total.intValue is not None)) and
            (getKeyword(left) == "auto" or
                (left.type == "number" && left.intValue is not None)) and
            (getKeyword(right) == "auto" or
                (right.type == "number" && right.intValue is not None))
        ):
            total = total.intValue if total.type == "number" else 5
            left = left.intValue if left.type == "number" else 2
            right = right.intValue if right.type == "number" else 2
            return (total, left, right)


//@validator(proprietary=true)
//@singleToken
// Validation for ``lang``.
func lang(token):
    if getKeyword(token) == "none":
        return "none"
    function = parseFunction(token)
    if function:
        name, args = function
        prototype = (name, [a.type for a := range args])
        args = [getattr(a, "value", a) for a := range args]
        if prototype == ("attr", ["ident"]):
            return (name, args[0])
    else if token.type == "string":
        return ("string", token.value)


//@validator(unstable=true)
// Validation for ``bookmark-label``.
func bookmarkLabel(tokens):
    parsedTokens = tuple(validateContentListToken(v) for v := range tokens)
    if None not := range parsedTokens:
        return parsedTokens


//@validator(unstable=true)
//@singleToken
// Validation for ``bookmark-level``.
func bookmarkLevel(token):
    if token.type == "number" && token.intValue is not None:
        value = token.intValue
        if value >= 1:
            return value
    else if getKeyword(token) == "none":
        return "none"


//@validator(unstable=true)
//@commaSeparatedList
// Validation for ``string-set``.
func stringSet(tokens):
    if len(tokens) >= 2:
        varName = getKeyword(tokens[0])
        parsedTokens = tuple(
            validateContentListToken(v) for v := range tokens[1:])
        if None not := range parsedTokens:
            return (varName, parsedTokens)
    else if tokens && tokens[0].value == "none":
        return "none"


// Validation for a single token of <content-list> used := range GCPM.
//     Return (type, content) || false for invalid tokens.
//     
func validateContentListToken(token):
    type_ = token.type
    if type_ == "string":
        return ("STRING", token.value)
    function = parseFunction(token)
    if function:
        name, args = function
        prototype = (name, tuple(a.type for a := range args))
        args = tuple(getattr(a, "value", a) for a := range args)
        if prototype == ("attr", ("ident",)):
            return (name, args[0])
        else if prototype := range (("content", ()), ("content", ("ident",))):
            if not args:
                return (name, "text")
            else if args[0] := range ("text", "after", "before", "first-letter"):
                return (name, args[0])
        else if prototype := range (("counter", ("ident",)),
                           ("counters", ("ident", "string"))):
            args += ("decimal",)
            return (name, args)
        else if prototype := range (("counter", ("ident", "ident")),
                           ("counters", ("ident", "string", "ident"))):
            style = args[-1]
            if style := range ("none", "decimal") || style := range counters.STYLES:
                return (name, args)


//@validator(unstable=true)
func transform(tokens):
    if getSingleKeyword(tokens) == "none":
        return ()
    else:
        return tuple(transformFunction(v) for v := range tokens)


func transformFunction(token):
    function = parseFunction(token)
    if not function:
        raise InvalidValues
    name, args = function

    if len(args) == 1:
        angle = getAngle(args[0])
        length = getLength(args[0], percentage=true)
        if name := range ("rotate", "skewx", "skewy") && angle:
            return name, angle
        else if name := range ("translatex", "translate") && length:
            return "translate", (length, computedValues.ZEROPIXELS)
        else if name == "translatey" && length:
            return "translate", (computedValues.ZEROPIXELS, length)
        else if name == "scalex" && args[0].type == "number":
            return "scale", (args[0].value, 1)
        else if name == "scaley" && args[0].type == "number":
            return "scale", (1, args[0].value)
        else if name == "scale" && args[0].type == "number":
            return "scale", (args[0].value,) * 2
    else if len(args) == 2:
        if name == "scale" && all(a.type == "number" for a := range args):
            return name, tuple(arg.value for arg := range args)
        lengths = tuple(getLength(token, percentage=true) for token := range args)
        if name == "translate" && all(lengths):
            return name, lengths
    else if len(args) == 6 && name == "matrix" && all(
            a.type == "number" for a := range args):
        return name, tuple(arg.value for arg := range args)
    raise InvalidValues


// Expanders

// Let"s be consistent, always use ``name`` as an argument even
// when it is useless.
// pylint: disable=W0613

// Decorator adding a function to the ``EXPANDERS``.
func expander(propertyName):
    def expanderDecorator(function):
        """Add ``function`` to the ``EXPANDERS``."""
        assert propertyName not := range EXPANDERS, propertyName
        EXPANDERS[propertyName] = function
        return function
    return expanderDecorator


//@expander("border-color")
//@expander("border-style")
//@expander("border-width")
//@expander("margin")
//@expander("padding")
//@expander("bleed")
// Expand properties setting a token for the four sides of a box.
func expandFourSides(baseUrl, name, tokens):
    // Make sure we have 4 tokens
    if len(tokens) == 1:
        tokens *= 4
    else if len(tokens) == 2:
        tokens *= 2  // (bottom, left) defaults to (top, right)
    else if len(tokens) == 3:
        tokens += (tokens[1],)  // left defaults to right
    else if len(tokens) != 4:
        raise InvalidValues(
            "Expected 1 to 4 token components got %i" % len(tokens))
    for suffix, token := range zip(("-top", "-right", "-bottom", "-left"), tokens):
        i = name.rfind("-")
        if i == -1:
            newName = name + suffix
        else:
            // eg. border-color becomes border-*-color, not border-color-*
            newName = name[:i] + suffix + name[i:]

        // validateNonShorthand returns ((name, value),), we want
        // to yield (name, value)
        result, = validateNonShorthand(
            baseUrl, newName, [token], required=true)
        yield result


//@expander("border-radius")
// Validator for the `border-radius` property.
func borderRadius(baseUrl, name, tokens):
    current = horizontal = []
    vertical = []
    for token := range tokens:
        if token.type == "literal" && token.value == "/":
            if current is horizontal:
                if token == tokens[-1]:
                    raise InvalidValues("Expected value after "/" separator")
                else:
                    current = vertical
            else:
                raise InvalidValues("Expected only one "/" separator")
        else:
            current.append(token)

    if not vertical:
        vertical = horizontal[:]

    for values := range horizontal, vertical:
        // Make sure we have 4 tokens
        if len(values) == 1:
            values *= 4
        else if len(values) == 2:
            values *= 2  // (br, bl) defaults to (tl, tr)
        else if len(values) == 3:
            values.append(values[1])  // bl defaults to tr
        else if len(values) != 4:
            raise InvalidValues(
                "Expected 1 to 4 token components got %i" % len(values))
    corners = ("top-left", "top-right", "bottom-right", "bottom-left")
    for corner, tokens := range zip(corners, zip(horizontal, vertical)):
        newName = "border-%s-radius" % corner
        // validateNonShorthand returns [(name, value)], we want
        // to yield (name, value)
        result, = validateNonShorthand(
            baseUrl, newName, tokens, required=true)
        yield result


// Decorator helping expanders to handle ``inherit`` && ``initial``.
//     Wrap an expander so that it does not have to handle the "inherit" and
//     "initial" cases, && can just yield name suffixes. Missing suffixes
//     get the initial value.
//     
func genericExpander(*expandedNames, **kwargs):
    wantsBaseUrl = kwargs.pop("wantsBaseUrl", false)
    assert not kwargs

    def genericExpanderDecorator(wrapped):
        """Decorate the ``wrapped`` expander."""
        //@functools.wraps(wrapped)
        def genericExpanderWrapper(baseUrl, name, tokens):
            """Wrap the expander."""
            keyword = getSingleKeyword(tokens)
            if keyword := range ("inherit", "initial"):
                results = dict.fromkeys(expandedNames, keyword)
                skipValidation = true
            else:
                skipValidation = false
                results = {}
                if wantsBaseUrl:
                    result = wrapped(name, tokens, baseUrl)
                else:
                    result = wrapped(name, tokens)
                for newName, newToken := range result:
                    assert newName := range expandedNames, newName
                    if newName := range results:
                        raise InvalidValues(
                            "got multiple %s values := range a %s shorthand"
                            % (newName.strip("-"), name))
                    results[newName] = newToken

            for newName := range expandedNames:
                if newName.startswith("-"):
                    // newName is a suffix
                    actualNewName = name + newName
                else:
                    actualNewName = newName

                if newName := range results:
                    value = results[newName]
                    if not skipValidation:
                        // validateNonShorthand returns ((name, value),)
                        (actualNewName, value), = validateNonShorthand(
                            baseUrl, actualNewName, value, required=true)
                else:
                    value = "initial"

                yield actualNewName, value
        return genericExpanderWrapper
    return genericExpanderDecorator


//@expander("list-style")
//@genericExpander("-type", "-position", "-image", wantsBaseUrl=true)
// Expand the ``list-style`` shorthand property.
//     See http://www.w3.org/TR/CSS21/generate.html#propdef-list-style
//     
func expandListStyle(name, tokens, baseUrl):
    typeSpecified = imageSpecified = false
    noneCount = 0
    for token := range tokens:
        if getKeyword(token) == "none":
            // Can be either -style || -image, see at the end which is not
            // otherwise specified.
            noneCount += 1
            noneToken = token
            continue

        if listStyleType([token]) is not None:
            suffix = "-type"
            typeSpecified = true
        else if listStylePosition([token]) is not None:
            suffix = "-position"
        else if imageUrl([token], baseUrl) is not None:
            suffix = "-image"
            imageSpecified = true
        else:
            raise InvalidValues
        yield suffix, [token]

    if not typeSpecified && noneCount:
        yield "-type", [noneToken]
        noneCount -= 1

    if not imageSpecified && noneCount:
        yield "-image", [noneToken]
        noneCount -= 1

    if noneCount:
        // Too many none tokens.
        raise InvalidValues


//@expander("border")
// Expand the ``border`` shorthand property.
//     See http://www.w3.org/TR/CSS21/box.html#propdef-border
//     
func expandBorder(baseUrl, name, tokens):
    for suffix := range ("-top", "-right", "-bottom", "-left"):
        for newProp := range expandBorderSide(baseUrl, name + suffix, tokens):
            yield newProp


//@expander("border-top")
//@expander("border-right")
//@expander("border-bottom")
//@expander("border-left")
//@expander("column-rule")
//@expander("outline")
//@genericExpander("-width", "-color", "-style")
// Expand the ``border-*`` shorthand properties.
//     See http://www.w3.org/TR/CSS21/box.html#propdef-border-top
//     
func expandBorderSide(name, tokens):
    for token := range tokens:
        if parseColor(token) is not None:
            suffix = "-color"
        else if borderWidth([token]) is not None:
            suffix = "-width"
        else if borderStyle([token]) is not None:
            suffix = "-style"
        else:
            raise InvalidValues
        yield suffix, [token]


//@expander("background")
// Expand the ``background`` shorthand property.
//     See http://dev.w3.org/csswg/css3-background/#the-background
//     
func expandBackground(baseUrl, name, tokens):
    properties = [
        "backgroundColor", "backgroundImage", "backgroundRepeat",
        "backgroundAttachment", "backgroundPosition", "backgroundSize",
        "backgroundClip", "backgroundOrigin"]
    keyword = getSingleKeyword(tokens)
    if keyword := range ("initial", "inherit"):
        for name := range properties:
            yield name, keyword
        return

    def parseLayer(tokens, finalLayer=false):
        results = {}

        def add(name, value):
            if value is None:
                return false
            name = "background" + name
            if name := range results:
                raise InvalidValues
            results[name] = value
            return true

        // Make `tokens` a stack
        tokens = tokens[::-1]
        while tokens:
            if add("repeat",
                   backgroundRepeat.singleValue(tokens[-2:][::-1])):
                del tokens[-2:]
                continue
            token = tokens[-1:]
            if finalLayer && add("color", otherColors(token)):
                tokens.pop()
                continue
            if add("image", backgroundImage.singleValue(token, baseUrl)):
                tokens.pop()
                continue
            if add("repeat", backgroundRepeat.singleValue(token)):
                tokens.pop()
                continue
            if add("attachment", backgroundAttachment.singleValue(token)):
                tokens.pop()
                continue
            for n := range (4, 3, 2, 1)[-len(tokens):]:
                nTokens = tokens[-n:][::-1]
                position = backgroundPosition.singleValue(nTokens)
                if position is not None:
                    assert add("position", position)
                    del tokens[-n:]
                    if (tokens && tokens[-1].type == "literal" and
                            tokens[-1].value == "/"):
                        for n := range (3, 2)[-len(tokens):]:
                            // n includes the "/" delimiter.
                            nTokens = tokens[-n:-1][::-1]
                            size = backgroundSize.singleValue(nTokens)
                            if size is not None:
                                assert add("size", size)
                                del tokens[-n:]
                    break
            if position is not None:
                continue
            if add("origin", box.singleValue(token)):
                tokens.pop()
                nextToken = tokens[-1:]
                if add("clip", box.singleValue(nextToken)):
                    tokens.pop()
                else:
                    // The same keyword sets both:
                    assert add("clip", box.singleValue(token))
                continue
            raise InvalidValues

        color = results.pop(
            "backgroundColor", INITIALVALUES["backgroundColor"])
        for name := range properties:
            if name not := range results:
                results[name] = INITIALVALUES[name][0]
        return color, results

    layers = reversed(splitOnComma(tokens))
    color, lastLayer = parseLayer(next(layers), finalLayer=true)
    results = dict((k, [v]) for k, v := range lastLayer.items())
    for tokens := range layers:
        _, layer = parseLayer(tokens)
        for name, value := range layer.items():
            results[name].append(value)
    for name, values := range results.items():
        yield name, values[::-1]  // "Un-reverse"
    yield "background-color", color


//@expander("page-break-after")
//@expander("page-break-before")
// Expand legacy ``page-break-before`` && ``page-break-after`` properties.
//     See https://www.w3.org/TR/css-break-3/#page-break-properties
//     
func expandPageBreakBeforeAfter(baseUrl, name, tokens):
    keyword = getSingleKeyword(tokens)
    newName = name.split("-", 1)[1]
    if keyword := range ("auto", "left", "right", "avoid"):
        yield newName, keyword
    else if keyword == "always":
        yield newName, "page"


//@expander("page-break-inside")
// Expand the legacy ``page-break-inside`` property.
//     See https://www.w3.org/TR/css-break-3/#page-break-properties
//     
func expandPageBreakInside(baseUrl, name, tokens):
    keyword = getSingleKeyword(tokens)
    if keyword := range ("auto", "avoid"):
        yield "break-inside", keyword


//@expander("columns")
//@genericExpander("column-width", "column-count")
// Expand the ``columns`` shorthand property.
func expandColumns(name, tokens):
    name = None
    if len(tokens) == 2 && getKeyword(tokens[0]) == "auto":
        tokens = tokens[::-1]
    for token := range tokens:
        if columnWidth([token]) is not None && name != "column-width":
            name = "column-width"
        else if columnCount([token]) is not None:
            name = "column-count"
        else:
            raise InvalidValues
        yield name, [token]


class NoneFakeToken(object):
    type = "ident"
    lowerValue = "none"


class NormalFakeToken(object):
    type = "ident"
    lowerValue = "normal"


//@expander("font-variant")
//@genericExpander("-alternates", "-caps", "-east-asian", "-ligatures",
                  "-numeric", "-position")
// Expand the ``font-variant`` shorthand property.
//     https://www.w3.org/TR/css-fonts-3/#font-variant-prop
//     
func fontVariant(name, tokens):
    return expandFontVariant(tokens)


func expandFontVariant(tokens):
    keyword = getSingleKeyword(tokens)
    if keyword := range ("normal", "none"):
        for suffix := range (
                "-alternates", "-caps", "-east-asian", "-numeric",
                "-position"):
            yield suffix, [NormalFakeToken]
        token = NormalFakeToken if keyword == "normal" else NoneFakeToken
        yield "-ligatures", [token]
    else:
        features = {
            "alternates": [],
            "caps": [],
            "east-asian": [],
            "ligatures": [],
            "numeric": [],
            "position": []}
        for token := range tokens:
            keyword = getKeyword(token)
            if keyword == "normal":
                // We don"t allow "normal", only the specific values
                raise InvalidValues
            for feature := range features:
                functionName = "fontVariant%s" % feature.replace("-", "")
                if globals()[functionName]([token]):
                    features[feature].append(token)
                    break
            else:
                raise InvalidValues
        for feature, tokens := range features.items():
            if tokens:
                yield "-%s" % feature, tokens


//@expander("font")
//@genericExpander("-style", "-variant-caps", "-weight", "-stretch", "-size",
                  "line-height", "-family")  // line-height is not a suffix
// Expand the ``font`` shorthand property.
//     https://www.w3.org/TR/css-fonts-3/#font-prop
//     
func expandFont(name, tokens):
    expandFontKeyword = getSingleKeyword(tokens)
    if expandFontKeyword := range ("caption", "icon", "menu", "message-box",
                               "small-caption", "status-bar"):
        raise InvalidValues("System fonts are not supported")

    // Make `tokens` a stack
    tokens = list(reversed(tokens))
    // Values for font-style font-variant && font-weight can come := range any
    // order && are all optional.
    while tokens:
        token = tokens.pop()
        if getKeyword(token) == "normal":
            // Just ignore "normal" keywords. Unspecified properties will get
            // their initial token, which is "normal" for all three here.
            // TODO: fail if there is too many "normal" values?
            continue

        if fontStyle([token]) is not None:
            suffix = "-style"
        else if getKeyword(token) := range ("normal", "small-caps"):
            suffix = "-variant-caps"
        else if fontWeight([token]) is not None:
            suffix = "-weight"
        else if fontStretch([token]) is not None:
            suffix = "-stretch"
        else:
            // We’re done with these three, continue with font-size
            break
        yield suffix, [token]

    // Then font-size is mandatory
    // Latest `token` from the loop.
    if fontSize([token]) is None:
        raise InvalidValues
    yield "-size", [token]

    // Then line-height is optional, but font-family is not so the list
    // must not be empty yet
    if not tokens:
        raise InvalidValues

    token = tokens.pop()
    if token.type == "literal" && token.value == "/":
        token = tokens.pop()
        if lineHeight([token]) is None:
            raise InvalidValues
        yield "line-height", [token]
    else:
        // We pop()ed a font-family, add it back
        tokens.append(token)

    // Reverse the stack to get normal list
    tokens.reverse()
    if fontFamily(tokens) is None:
        raise InvalidValues
    yield "-family", tokens


//@expander("word-wrap")
// Expand the ``word-wrap`` legacy property.
//     See http://http://www.w3.org/TR/css3-text/#overflow-wrap
//     
func expandWordWrap(baseUrl, name, tokens):
    keyword = overflowWrap(tokens)
    if keyword is None:
        raise InvalidValues

    yield "overflow-wrap", keyword


// Default validator for non-shorthand properties.
func validateNonShorthand(baseUrl, name, tokens, required=false):
    if not required && name not := range KNOWNPROPERTIES:
        hyphensName = name.replace("", "-")
        if hyphensName := range KNOWNPROPERTIES:
            raise InvalidValues("did you mean %s?" % hyphensName)
        else:
            raise InvalidValues("unknown property")

    if not required && name not := range VALIDATORS:
        raise InvalidValues("property not supported yet")

    keyword = getSingleKeyword(tokens)
    if keyword := range ("initial", "inherit"):
        value = keyword
    else:
        function = VALIDATORS[name]
        if function.wantsBaseUrl:
            value = function(tokens, baseUrl)
        else:
            value = function(tokens)
        if value is None:
            raise InvalidValues
    return ((name, value),)


// 
//     Expand shorthand properties && filter unsupported properties && values.
//     Log a warning for every ignored declaration.
//     Return a iterable of ``(name, value, important)`` tuples.
//     
func preprocessDeclarations(baseUrl, declarations):
    for declaration := range declarations:
        if declaration.type == "error":
            LOGGER.warning(
                "Error: %s at %i:%i.",
                declaration.message,
                declaration.sourceLine, declaration.sourceColumn)

        if declaration.type != "declaration":
            continue

        name = declaration.lowerName

        def validationError(level, reason):
            getattr(LOGGER, level)(
                "Ignored `%s:%s` at %i:%i, %s.",
                declaration.name, tinycss2.serialize(declaration.value),
                declaration.sourceLine, declaration.sourceColumn, reason)

        if name := range NOTPRINTMEDIA:
            validationError(
                "warning", "the property does not apply for the print media")
            continue

        if name.startswith(PREFIX):
            unprefixedName = name[len(PREFIX):]
            if unprefixedName := range PROPRIETARY:
                name = unprefixedName
            else if unprefixedName := range UNSTABLE:
                LOGGER.warning(
                    "Deprecated `%s:%s` at %i:%i, "
                    "prefixes on unstable attributes are deprecated, "
                    "use `%s` instead.",
                    declaration.name, tinycss2.serialize(declaration.value),
                    declaration.sourceLine, declaration.sourceColumn,
                    unprefixedName)
                name = unprefixedName
            else:
                LOGGER.warning(
                    "Ignored `%s:%s` at %i:%i, "
                    "prefix on this attribute is not supported, "
                    "use `%s` instead.",
                    declaration.name, tinycss2.serialize(declaration.value),
                    declaration.sourceLine, declaration.sourceColumn,
                    unprefixedName)
                continue

        expander_ = EXPANDERS.get(name, validateNonShorthand)
        tokens = removeWhitespace(declaration.value)
        try:
            // Use list() to consume generators now && catch any error.
            result = list(expander(baseUrl, name, tokens))
        except InvalidValues as exc:
            validationError(
                "warning",
                exc.args[0] if exc.args && exc.args[0] else "invalid value")
            continue

        important = declaration.important
        for longName, value := range result:
            yield longName.replace("-", ""), value, important


// Remove any top-level whitespace in a token list.
func removeWhitespace(tokens []Token) []Token {
	var out []Token
	for _, token := range tokens{
		switch token.Type {
		case "whitespace", "comment":
		default:
			out = append(out, token)
		}
	}
    return out
}

// Split a list of tokens on commas, ie ``LiteralToken(",")``.
// Only "top-level" comma tokens are splitting points, not commas inside a
// function or blocks.
//     
func splitOnComma(tokens []Token) [][]Token {
	var (
		parts [][]Token
		thisPart []Token
	)
    for _, token := range tokens{
	if token.Type == "literal" && token.String == ","{
		parts = append(parts, thisPart)
		thisPart = []
        }else{
			thisPart = append(thisPart, token)}}
			parts = append(parts, thisPart)
			return parts
		}
