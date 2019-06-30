package css

import (
	"log"
	"strconv"
	"strings"

	"github.com/benoitkugler/go-weasyprint/utils"
	"golang.org/x/net/html"
)

var (
	ZeroPixels = Dimension{Unit: "px", Value: 0}

	// How many CSS pixels is one <unit>?
	// http://www.w3.org/TR/CSS21/syndata.html#length-units
	LengthsToPixels = map[string]float64{
		"px": 1,
		"pt": 1. / 0.75,
		"pc": 16.,             // LengthsToPixels["pt"] * 12
		"in": 96.,             // LengthsToPixels["pt"] * 72
		"cm": 96. / 2.54,      // LengthsToPixels["in"] / 2.54
		"mm": 96. / 25.4,      // LengthsToPixels["in"] / 25.4
		"q":  96. / 25.4 / 4., // LengthsToPixels["mm"] / 4
	}

	// These are unspecified, other than 'thin' <='medium' <= 'thick'.
	// Values are in pixels.
	BorderWidthKeywords = map[string]int{
		"thin":   1,
		"medium": 3,
		"thick":  5,
	}

	// Value in pixels of font-size for <absolute-size> keywords: 12pt (16px) for
	// medium, and scaling factors given in CSS3 for others:
	// http://www.w3.org/TR/css3-fonts/#font-size-prop
	// TODO: this will need to be ordered to implement 'smaller' and 'larger'
	FontSizeKeywords = map[string]int{ // medium is 16px, others are a ratio of medium
		"xx-small": InitialValues.Ints["font_size"] * 3 / 5,
		"x-small":  InitialValues.Ints["font_size"] * 3 / 4,
		"small":    InitialValues.Ints["font_size"] * 8 / 9,
		"medium":   InitialValues.Ints["font_size"] * 1 / 1,
		"large":    InitialValues.Ints["font_size"] * 6 / 5,
		"x-large":  InitialValues.Ints["font_size"] * 3 / 2,
		"xx-large": InitialValues.Ints["font_size"] * 2 / 1,
	}

	// http://www.w3.org/TR/CSS21/fonts.html#propdef-font-weight
	FontWeightRelative = struct {
		bolder, lighter map[string]int
	}{
		bolder: map[string]int{
			"100": 400,
			"200": 400,
			"300": 400,
			"400": 700,
			"500": 700,
			"600": 900,
			"700": 900,
			"800": 900,
			"900": 900,
		},
		lighter: map[string]int{
			"100": 100,
			"200": 100,
			"300": 100,
			"400": 100,
			"500": 100,
			"600": 400,
			"700": 400,
			"800": 700,
			"900": 700,
		},
	}
)

func init() {
	if InitialValues.Ints["border_top_width"] != BorderWidthKeywords["medium"] {
		log.Fatal("border-top-width and medium should be the same !")
	}

}

// Return a dict of computed values.

// :param element: The HTML element these style apply to
// :param pseudo_type: The type of pseudo-element, eg 'before', None
// :param specified: a dict of specified values. Should contain
// 			  values for all properties.
// :param computed: a dict of already known computed values.
// 			 Only contains some properties (or none).
// :param parent_style: a dict of computed values of the parent
// 				 element (should contain values for all properties),
// 				 or ``None`` if ``element`` is the root element.
// :param base_url: The base URL used to resolve relative URLs.
func compute(element html.Node, pseudoType string,
	specified, computed, parent_style,
	root_style StyleDict, base_url string) StyleDict {
	// func computer() {
	// """Dummy object that holds attributes."""
	// return 0

	// computer.is_root_element = parent_style is None
	// if parent_style is None:
	// parent_style = INITIAL_VALUES

	// computer.element = element
	// computer.pseudo_type = pseudo_type
	// computer.specified = specified
	// computer.computed = computed
	// computer.parent_style = parent_style
	// computer.root_style = root_style
	// computer.base_url = base_url

	// getter = COMPUTER_FUNCTIONS.get

	// for name in COMPUTING_ORDER:
	// if name in computed:
	// 	// Already computed
	// 	continue

	// value = specified[name]
	// function = getter(name)
	// if function is not None:
	// 	value = function(computer computer, name string, value)
	// // else: same as specified

	// computed[name] = value

	// computed['_weasy_specified_display'] = specified['display']
	return computed
}

type computer struct {
	isRootElement                               bool
	computed, rootStyle, parentStyle, specified StyleDict
	element                                     html.Node
}

type IntString struct {
	Int    int
	String string
}

// Dimension or string
type Value struct {
	Dimension
	String string
}

// Dimension or "auto" or "cover" or "contain"
type Size struct {
	Width, Height Dimension
	String        string
}

type Content struct {
	List   [][2]string
	String string
}

type Link struct {
	String string
	Type   string
	Attr   string
}
type GradientValue struct {
	StopPositions []Dimension
	Center        Center
	SizeType      string
	Size          []Dimension
}

type Gradient struct {
	Type  string
	Value GradientValue
}

type Center struct {
	OriginX, OriginY int
	PosX, PosY       Dimension
}

type Transform struct {
	Function string
	Args     []Dimension
}

// backgroundImage computes lenghts in gradient background-image.
func backgroundImage(computer computer, name string, values []Gradient) []Gradient {
	for _, gradient := range values {
		value := gradient.Value
		if gradient.Type == "linear-gradient" || gradient.Type == "radial-gradient" {
			for index, pos := range value.StopPositions {
				value.StopPositions[index] = length(computer, name, Value{Dimension: pos}, -1)
			}
		}
		if gradient.Type == "radial-gradient" {
			value.Center = backgroundPosition(computer, name, []Center{value.Center})[0]
			if value.SizeType == "explicit" {
				value.Size = lengthOrPercentageTuple2(computer, name, value.Size)
			}
		}
	}
	return values
}

// Compute lengths in background-position.
func backgroundPosition(computer computer, name string, values []Center) []Center {
	out := make([]Center, len(values))
	for index, v := range values {
		out[index] = Center{
			OriginX: v.OriginX,
			OriginY: v.OriginY,
			PosX:    length(computer, name, Value{Dimension: v.PosX}, -1),
			PosY:    length(computer, name, Value{Dimension: v.PosY}, -1),
		}
	}
	return out
}

// Compute the lists of lengths that can be percentages.
func lengthOrPercentageTuple(computer computer, name string, values []Value) []Dimension {
	out := make([]Dimension, len(values))
	for index, v := range values {
		out[index] = length(computer, name, v, -1)
	}
	return out
}

// Compute the lists of lengths that can be percentages.
func lengthOrPercentageTuple2(computer computer, name string, values []Dimension) []Dimension {
	out := make([]Dimension, len(values))
	for index, v := range values {
		out[index] = length(computer, name, Value{Dimension: v}, -1)
	}
	return out
}

// Compute the properties with a list of lengths.
func lengthTuple(computer computer, name string, values []Value) []int {
	out := make([]int, len(values))
	for index, v := range values {
		out[index] = length(computer, name, v, -1).Value
	}
	return out
}

// Compute the ``break-before`` and ``break-after`` properties.
func breakBeforeAfter(computer, name, value string) string {
	// "always" is defined as an alias to "page" in multi-column
	// https://www.w3.org/TR/css3-multicol/#column-breaks
	if value == "always" {
		return "page"
	}
	return value
}

// Compute a length ``value``.
// passing a negative fontSize means null
func length(computer computer, name string, value Value, fontSize int) Dimension {
	if value.String == "auto" || value.String == "content" {
		return value.Dimension
	}
	if value.Value == 0 {
		return ZeroPixels
	}

	unit := value.Unit
	var result float64
	switch unit {
	case "px":
		return value.Dimension
	case "pt", "pc", "in", "cm", "mm", "q":
		// Convert absolute lengths to pixels
		result = float64(value.Value) * LengthsToPixels[unit]
	case "em", "ex", "ch", "rem":
		if fontSize < 0 {
			fontSize = computer.computed.Ints["font_size"]
		}
		switch unit {
		// we dont support 'ex' and 'ch' units for now.
		case "ex", "ch", "em":
			result = float64(value.Value * fontSize)
		case "rem":
			result = float64(value.Value * computer.rootStyle.Ints["font_size"])
		default:
			// A percentage or "auto": no conversion needed.
			return value.Dimension
		}
		return Dimension{Value: int(result), Unit: "px"}
	}
	return Dimension{}
}

func bleed(computer computer, name string, value Value) Dimension {
	if value.String == "auto" {
		if strings.Contains(computer.computed.Strings["marks"], "crop") {
			return Dimension{Value: 8, Unit: "px"} // 6pt
		}
		return ZeroPixels
	}
	return length(computer, name, value, -1)
}

func pixelLength(computer computer, name string, value Value) IntString {
	if value.String == "normal" {
		return IntString{String: value.String}
	}
	return IntString{Int: length(computer, name, value, -1).Value}
}

// Compute the ``background-size`` properties.
func backgroundSize(computer computer, name string, values []Size) []Size {
	out := make([]Size, len(values))
	for index, v := range values {
		if v.String == "contain" || v.String == "cover" {
			out[index] = Size{String: v.String}
		} else {
			l := lengthOrPercentageTuple(computer, name, []Value{
				Value{Dimension: v.Height},
				Value{Dimension: v.Width},
			})
			out[index] = Size{
				Height: l[0],
				Width:  l[1],
			}
		}
	}
	return out
}

// Compute the ``border-*-width`` properties.
// value.String may be the string representation of an int
func borderWidth(computer computer, name string, value Value) int {
	style := computer.computed.Strings[strings.ReplaceAll(name, "width", "style")]
	if style == "none" || style == "hidden" {
		return 0
	}

	if bw, in := BorderWidthKeywords[value.String]; in {
		return bw
	}

	if i, err := strconv.Atoi(value.String); err == nil {
		// The initial value can get here, but length() would fail as
		// it does not have a "unit" attribute.
		return i
	}
	return length(computer, name, value, -1).Value
}

// Compute the ``column-width`` property.
func columnWidth(computer computer, name string, value Value) int {
	return length(computer, name, value, -1).Value
}

// Compute the ``border-*-radius`` properties.
func borderRadius(computer computer, name string, values []Value) []Dimension {
	out := make([]Dimension, len(values))
	for index, v := range values {
		out[index] = length(computer, name, v, -1)
	}
	return out
}

// Compute the ``column-gap`` property.
func columnGap(computer computer, name string, value Value) int {
	if value.String == "normal" {
		value = Value{Dimension: Dimension{Value: 1, Unit: "em"}}
	}
	return length(computer, name, value, -1).Value
}

// Compute the ``content`` property.
func content(computer computer, name string, values Content) Content {
	if values.String == "normal" || values.String == "none" {
		return values
	}
	lis := make([][2]string, len(values.List))
	for index, v := range values.List {
		type_, value := v[0], v[1]
		if type_ == "attr" {
			lis[index][0] = "STRING"
			lis[index][1] = utils.GetAttribute(computer.element, value, "")
		} else {
			lis[index] = v
		}
	}
	return Content{List: lis}
}

//Compute the ``display`` property.
// See http://www.w3.org/TR/CSS21/visuren.html#dis-pos-flo
func display(computer computer, name string, value string) string {
	float_ := computer.specified.Strings["float"]
	position := computer.specified.Strings["position"]
	if (position == "absolute" || position == "fixed") || float_ != "none" || computer.isRootElement {
		switch value {
		case "inline-table":
			return "table"
		case "inline", "table-row-group", "table-column",
			"table-column-group", "table-header-group",
			"table-footer-group", "table-row", "table-cell",
			"table-caption", "inline-block":
			return "block"
		}
	}
	return value
}

//Compute the ``float`` property.
// See http://www.w3.org/TR/CSS21/visuren.html#dis-pos-flo
func computeFloat(computer computer, name string, value string) string {
	position := computer.specified.Strings["position"]
	if position == "absolute" || position == "fixed" {
		return "none"
	}
	return value
}

// Compute the ``font-size`` property.
func fontSize(computer computer, name string, value Value) int {
	if fs, in := FontSizeKeywords[value.String]; in {
		return fs
	}
	// TODO: support "larger" and "smaller"

	parentFontSize := computer.parentStyle.Ints["font_size"]
	if value.Unit == "%" {
		return value.Value * parentFontSize / 100.
	}
	return length(computer, name, value, parentFontSize).Value
}

// Compute the ``font-weight`` property.
// value may be a string representation of an int
func fontWeight(computer computer, name string, value string) int {
	switch value {
	case "normal":
		return 400
	case "bold":
		return 700
	case "bolder":
		parentValue := computer.parentStyle.Strings["font_weight"]
		return FontWeightRelative.bolder[parentValue]
	case "lighter":
		parentValue := computer.parentStyle.Strings["font_weight"]
		return FontWeightRelative.lighter[parentValue]
	default:
		i, err := strconv.Atoi(value)
		if err != nil {
			log.Fatal("font_weight should be an int !")
		}
		return i
	}
}

// Compute the ``line-height`` property.
func lineHeight(computer computer, name string, value Value) Value {
	var pixels int
	switch {
	case value.String == "normal":
		return value
	case value.Unit == "":
		return Value{Dimension: Dimension{Value: value.Value, Unit: "NUMBER"}}
	case value.Unit == "%":
		factor := value.Value / 100.
		fontSizeValue := computer.computed.Ints["font_size"]
		pixels = factor * fontSizeValue
	default:
		pixels = length(computer, name, value, -1).Value
	}
	return Value{Dimension: Dimension{Value: pixels, Unit: "PIXELS"}}
}

// Compute the ``anchor`` property.
func anchor(computer computer, name string, values Link) string {
	if values.String != "none" {
		return utils.GetAttribute(computer.element, values.Attr, "")
	}
	return ""
}

// Compute the ``link`` property.
func link(computer computer, name string, values Link) []string {
	if values.String == "none" {
		return nil
	}
	if values.Type == "attr" {
		return utils.GetLinkAttribute(computer.element, values.Attr, computer.baseUrl)
	}
	return []string{values.Type, values.Attr}
}

// Compute the ``lang`` property.
func lang(computer computer, name string, values Link) string {
	if values.String == "none" {
		return ""
	}
	if values.Type == "attr()" {
		return utils.GetAttribute(computer.element, values.Attr, "")
	} else if values.Type == "string" {
		return values.Attr
	}
	return ""
}

// Compute the ``tab-size`` property.
func tabSize(computer computer, name string, value Value) Dimension {
	if value.Unit == "" {
		return value.Dimension
	}
	return length(computer, name, value, -1)
}

// Compute the ``transform`` property.
func transform(computer computer, name string, value []Transform) []Transform {
	result := make([]Transform, len(value))
	for index, tr := range value {
		if tr.Function == "translate" {
			tr.Args = lengthOrPercentageTuple2(computer, name, tr.Args)
		}
		result[index] = tr
	}
	return result
}

// Compute the ``vertical-align`` property.
func verticalAlign(computer computer, name string, value Value) IntString {
	// Use +/- half an em for super and sub, same as Pango.
	// (See the SUPERSUBRISE constant in pango-markup.c)
	var out IntString
	switch value.String {
	case "baseline", "middle", "text-top", "text-bottom", "top", "bottom":
		out.String = value.String
	case "super":
		out.Int = computer.computed.Ints["font_size"] * 0.5
	case "sub":
		out.Int = computer.computed.Ints["font_size"] * -0.5
	default:
		out.Int = length(computer, name, value, -1).Value
	}
	if value.Unit == "%" {
		//TODO: support
		// height, _ = strutLayout(computer.computed)
		// return height * value.value / 100.
		log.Println("% not supported for vertical-align")
	}
	return out
}

// Compute the ``word-spacing`` property.
func wordSpacing(computer computer, name string, value Value) int {
	if value.String == "normal" {
		return 0
	}
	return length(computer, name, value, -1).Value
}
