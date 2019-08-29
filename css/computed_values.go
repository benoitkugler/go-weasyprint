package css

import (
	"log"
	"sort"
	"strings"

	"github.com/benoitkugler/go-weasyprint/utils"
	"golang.org/x/net/html"
)

var (
	ZeroPixels = Value{Dimension: Dimension{Unit: "px"}}

	// How many CSS pixels is one <unit>?
	// http://www.w3.org/TR/CSS21/syndata.html#length-units
	LengthsToPixels = map[string]float32{
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
	BorderWidthKeywords = map[string]float32{
		"thin":   1,
		"medium": 3,
		"thick":  5,
	}

	// Value in pixels of font-size for <absolute-size> keywords: 12pt (16px) for
	// medium, and scaling factors given in CSS3 for others:
	// http://www.w3.org/TR/css3-fonts/#font-size-prop
	// TODO: this will need to be ordered to implement 'smaller' and 'larger'
	FontSizeKeywords = map[string]float32{ // medium is 16px, others are a ratio of medium
		"xx-small": InitialValues.Values["font_size"].Value * 3 / 5,
		"x-small":  InitialValues.Values["font_size"].Value * 3 / 4,
		"small":    InitialValues.Values["font_size"].Value * 8 / 9,
		"medium":   InitialValues.Values["font_size"].Value * 1 / 1,
		"large":    InitialValues.Values["font_size"].Value * 6 / 5,
		"x-large":  InitialValues.Values["font_size"].Value * 3 / 2,
		"xx-large": InitialValues.Values["font_size"].Value * 2 / 1,
	}

	// http://www.w3.org/TR/CSS21/fonts.html#propdef-font-weight
	FontWeightRelative = struct {
		bolder, lighter map[float32]float32
	}{
		bolder: map[float32]float32{
			100: 400,
			200: 400,
			300: 400,
			400: 700,
			500: 700,
			600: 900,
			700: 900,
			800: 900,
			900: 900,
		},
		lighter: map[float32]float32{
			100: 100,
			200: 100,
			300: 100,
			400: 100,
			500: 100,
			600: 400,
			700: 400,
			800: 700,
			900: 700,
		},
	}

	ComputingOrder []string
)

func init() {
	if InitialValues.Values["border_top_width"].Value != BorderWidthKeywords["medium"] {
		log.Fatal("border-top-width and medium should be the same !")
	}

	//Some computed value are required by others, so order matters.
	ComputingOrder = []string{"font_stretch", "font_weight", "font_family", "font_variant",
		"font_style", "font_size", "line_height", "marks"}
	crible := map[string]bool{}
	for _, k := range ComputingOrder {
		crible[k] = true
	}
	keys := InitialValues.Keys()
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	for _, k := range keys {
		if !crible[k] {
			ComputingOrder = append(ComputingOrder, k)
		}
	}
}

// Return a dict of computed value.

// :param element: The HTML element these style apply to
// :param specified: a dict of specified value. Should contain
// 			  value for all properties.
// :param computed: a dict of already known computed value.
// 			 Only contains some properties (or none).
// :param parentStyle: a dict of computed value of the parent
// 				 element (should contain value for all properties),
// 				 or `zero if ``element`` is the root element.
// :param baseUrl: The base URL used to resolve relative URLs.
func compute(element html.Node, specified, computed, parentStyle,
	rootStyle StyleDict, baseUrl string) StyleDict {

	computer := new(computer)

	computer.isRootElement = parentStyle.IsZero()
	if parentStyle.IsZero() {
		parentStyle = InitialValues
	}

	computer.element = element
	computer.specified = specified
	computer.computed = computed
	computer.parentStyle = parentStyle
	computer.rootStyle = rootStyle
	computer.baseUrl = baseUrl

	computedKeys := Set{}
	for _, k := range computed.Keys() {
		computedKeys[k] = true
	}

	specifiedItems := specified.Items()
	for _, name := range ComputingOrder {
		if computedKeys[name] {
			// Already computed
			continue
		}
		computedKeys[name] = true
		value := specifiedItems[name]
		value = value.ComputeValue(computer, name)
		value.SetOn(name, &computed)
	}

	computed.weasySpecifiedDisplay = Display(specified.Strings["display"])
	return computed
}

type computer struct {
	isRootElement                               bool
	computed, rootStyle, parentStyle, specified StyleDict
	element                                     html.Node
	baseUrl                                     string
}

// Dimension or "auto" or "cover" or "contain"
type Size struct {
	Width, Height Value
	String        string
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
	OriginX, OriginY string
	PosX, PosY       Dimension
}

type Transform struct {
	Function string
	Args     []Dimension
}

// backgroundImage computes lenghts in gradient background-image.
func (value BackgroundImage) ComputeValue(computer *computer, name string) CssProperty {
	for _, gradient := range value {
		value := gradient.Value
		if gradient.Type == "linear-gradient" || gradient.Type == "radial-gradient" {
			for index, pos := range value.StopPositions {
				value.StopPositions[index] = length(computer, name, Value{Dimension: pos}).Dimension
			}
		}
		if gradient.Type == "radial-gradient" {
			value.Center = backgroundPosition(computer, name, []Center{value.Center})[0]
			if value.SizeType == "explicit" {
				value.Size = lengthOrPercentageTuple2(computer, name, value.Size)
			}
		}
	}
	return value
}

// Compute lengths in background-position.
func (value BackgroundPosition) ComputeValue(computer *computer, name string) CssProperty {
	return BackgroundPosition(backgroundPosition(computer, name, []Center(value)))
}

func backgroundPosition(computer *computer, name string, value []Center) []Center {
	out := make([]Center, len(value))
	for index, v := range value {
		out[index] = Center{
			OriginX: v.OriginX,
			OriginY: v.OriginY,
			PosX:    length(computer, name, Value{Dimension: v.PosX}).Dimension,
			PosY:    length(computer, name, Value{Dimension: v.PosY}).Dimension,
		}
	}
	return out
}

// Compute the lists of lengths that can be percentages.
func (value Lengths) ComputeValue(computer *computer, name string) CssProperty {
	out := make(Lengths, len(value))
	for index, v := range value {
		out[index] = length(computer, name, v)
	}
	return out
}

// Compute the lists of lengths that can be percentages.
func lengthOrPercentageTuple2(computer *computer, name string, value []Dimension) []Dimension {
	out := make([]Dimension, len(value))
	for index, v := range value {
		out[index] = length(computer, name, Value{Dimension: v}).Dimension
	}
	return out
}

// Compute the ``break-before`` and ``break-after`` properties.
func (value Break) ComputeValue(computer *computer, name string) CssProperty {
	// "always" is defined as an alias to "page" in multi-column
	// https://www.w3.org/TR/css3-multicol/#column-breaks
	if value == "always" {
		return Break("page")
	}
	return value
}

func (value Length) ComputeValue(computer *computer, name string) CssProperty {
	return Length(length(computer, name, Value(value)))
}

func length(computer *computer, name string, value Value) Value {
	return length2(computer, name, value, -1)
}

// Compute a length ``value``.
// passing a negative fontSize means null
func length2(computer *computer, _ string, value Value, fontSize float32) Value {
	if value.String == "auto" || value.String == "content" {
		return value
	}
	if value.Value == 0 {
		return ZeroPixels
	}

	unit := value.Unit
	var result float32
	switch unit {
	case "px":
		return value
	case "pt", "pc", "in", "cm", "mm", "q":
		// Convert absolute lengths to pixels
		result = float32(value.Value) * LengthsToPixels[unit]
	case "em", "ex", "ch", "rem":
		if fontSize < 0 {
			fontSize = computer.computed.Values["font_size"].Value
		}
		switch unit {
		// we dont support 'ex' and 'ch' units for now.
		case "ex", "ch", "em":
			result = float32(value.Value * fontSize)
		case "rem":
			result = float32(value.Value * computer.rootStyle.Values["font_size"].Value)
		default:
			// A percentage or "auto": no conversion needed.
			return value
		}
	}
	return Value{Dimension: Dimension{Value: result, Unit: "px"}}
}

func (value Bleed) ComputeValue(computer *computer, name string) CssProperty {
	if value.String == "auto" {
		if strings.Contains(computer.computed.Strings["marks"], "crop") {
			return Bleed{Dimension: Dimension{Value: 8, Unit: "px"}} // 6pt
		}
		return Bleed(ZeroPixels)
	}
	return Bleed(length(computer, name, Value(value)))
}

func (value PixelLength) ComputeValue(computer *computer, name string) CssProperty {
	if value.String == "normal" {
		return PixelLength{String: value.String}
	}
	return PixelLength(length(computer, name, Value(value)))
}

// Compute the ``background-size`` properties.
func (value BackgroundSize) ComputeValue(computer *computer, name string) CssProperty {
	out := make(BackgroundSize, len(value))
	for index, v := range value {
		if v.String == "contain" || v.String == "cover" {
			out[index] = Size{String: v.String}
		} else {
			l := lengthOrPercentageTuple2(computer, name, []Dimension{
				v.Height.Dimension,
				v.Width.Dimension,
			})
			out[index] = Size{
				Height: Value{Dimension: l[0]},
				Width:  Value{Dimension: l[1]},
			}
		}
	}
	return out
}

// Compute the ``border-*-width`` properties.
// value.String may be the string representation of an int
func (value BorderWidth) ComputeValue(computer *computer, name string) CssProperty {
	style := computer.computed.Strings[strings.ReplaceAll(name, "width", "style")]
	if style == "none" || style == "hidden" {
		return BorderWidth{}
	}

	if bw, in := BorderWidthKeywords[value.String]; in {
		return BorderWidth{Dimension: Dimension{Value: bw}}
	}
	return BorderWidth(length(computer, name, Value(value)))
}

// Compute the ``column-width`` property.
func (value ColumnWidth) ComputeValue(computer *computer, name string) CssProperty {
	return ColumnWidth(length(computer, name, Value(value)))
}

// Compute the ``column-gap`` property.
func (value ColumnGap) ComputeValue(computer *computer, name string) CssProperty {
	if value.String == "normal" {
		value = ColumnGap{Dimension: Dimension{Value: 1, Unit: "em"}}
	}
	return ColumnGap(length(computer, name, Value(value)))
}

// Compute the ``content`` property.
func (value Content) ComputeValue(computer *computer, name string) CssProperty {
	if value.String == "normal" || value.String == "none" {
		return value
	}
	lis := make([]ContentProperty, len(value.List))
	for index, v := range value.List {
		// type_, value := v[0], v[1]
		if v.Type == ContentAttr {
			lis[index].Type = ContentSTRING
			lis[index].String = utils.GetAttribute(computer.element, value.String)
		} else {
			lis[index] = v
		}
	}
	return Content{List: lis}
}

//Compute the ``display`` property.
// See http://www.w3.org/TR/CSS21/visuren.html#dis-pos-flo
func (value Display) ComputeValue(computer *computer, name string) CssProperty {
	float_ := computer.specified.Strings["float"]
	position := computer.specified.Strings["position"]
	if (position == "absolute" || position == "fixed") || float_ != "none" || computer.isRootElement {
		switch value {
		case "inline-table":
			return Display("table")
		case "inline", "table-row-group", "table-column",
			"table-column-group", "table-header-group",
			"table-footer-group", "table-row", "table-cell",
			"table-caption", "inline-block":
			return Display("block")
		}
	}
	return value
}

//Compute the ``float`` property.
// See http://www.w3.org/TR/CSS21/visuren.html#dis-pos-flo
func (value Float) ComputeValue(computer *computer, name string) CssProperty {
	position := computer.specified.Strings["position"]
	if position == "absolute" || position == "fixed" {
		return Float("none")
	}
	return value
}

// Compute the ``font-size`` property.
func (value FontSize) ComputeValue(computer *computer, name string) CssProperty {
	if fs, in := FontSizeKeywords[value.String]; in {
		return FontSize{Dimension: Dimension{Value: fs}}
	}
	// TODO: support "larger" and "smaller"

	parentFontSize := computer.parentStyle.Values["font_size"].Value
	if value.Unit == "%" {
		return FontSize{Dimension: Dimension{Value: value.Value * parentFontSize / 100.}}
	}
	return FontSize(length2(computer, name, Value(value), parentFontSize))
}

// Compute the ``font-weight`` property.
func (value FontWeight) ComputeValue(computer *computer, name string) CssProperty {
	var out float32
	switch value.String {
	case "normal":
		out = 400
	case "bold":
		out = 700
	case "bolder":
		parentValue := computer.parentStyle.Values["font_weight"].Value
		out = FontWeightRelative.bolder[parentValue]
	case "lighter":
		parentValue := computer.parentStyle.Values["font_weight"].Value
		out = FontWeightRelative.lighter[parentValue]
	default:
		out = value.Value
	}
	return FontWeight{Dimension: Dimension{Value: out}}
}

// Compute the ``line-height`` property.
func (value LineHeight) ComputeValue(computer *computer, name string) CssProperty {
	var pixels float32
	switch {
	case value.String == "normal":
		return value
	case value.Unit == "":
		return LineHeight{Dimension: Dimension{Value: value.Value, Unit: "NUMBER"}}
	case value.Unit == "%":
		factor := value.Value / 100.
		fontSizeValue := computer.computed.Values["font_size"].Value
		pixels = factor * fontSizeValue
	default:
		pixels = length(computer, name, Value(value)).Value
	}
	return LineHeight{Dimension: Dimension{Value: pixels, Unit: "PIXELS"}}
}

// Compute the ``anchor`` property.
func (value Anchor) ComputeValue(computer *computer, name string) CssProperty {
	if value.String != "none" {
		return Anchor{String: utils.GetAttribute(computer.element, value.Attr)}
	}
	return Anchor{}
}

// Compute the ``link`` property.
func (value Link) ComputeValue(computer *computer, name string) CssProperty {
	if value.String == "none" {
		return Link{}
	}
	if value.Type == "attr" {
		type_attr := utils.GetLinkAttribute(computer.element, value.Attr, computer.baseUrl)
		if len(type_attr) < 2 {
			return Link{}
		}
		return Link{Type: type_attr[0], Attr: type_attr[1]}
	}
	return value
}

// Compute the ``lang`` property.
func (value Lang) ComputeValue(computer *computer, name string) CssProperty {
	if value.String == "none" {
		return Lang{}
	}
	if value.Type == "attr()" {
		return Lang{String: utils.GetAttribute(computer.element, value.Attr)}
	} else if value.Type == "string" {
		return Lang{String: value.Attr}
	}
	return Lang{}
}

// Compute the ``tab-size`` property.
func (value TabSize) ComputeValue(computer *computer, name string) CssProperty {
	if value.Unit == "" {
		return value
	}
	return TabSize(length(computer, name, Value(value)))
}

// Compute the ``transform`` property.
func (value Transforms) ComputeValue(computer *computer, name string) CssProperty {
	result := make(Transforms, len(value))
	for index, tr := range value {
		if tr.Function == "translate" {
			tr.Args = lengthOrPercentageTuple2(computer, name, tr.Args)
		}
		result[index] = tr
	}
	return result
}

// Compute the ``vertical-align`` property.
func (value VerticalAlign) ComputeValue(computer *computer, name string) CssProperty {
	// Use +/- half an em for super and sub, same as Pango.
	// (See the SUPERSUBRISE constant in pango-markup.c)
	var out Value
	switch value.String {
	case "baseline", "middle", "text-top", "text-bottom", "top", "bottom":
		out.String = value.String
	case "super":
		out.Value = float32(computer.computed.Values["font_size"].Value) * 0.5
	case "sub":
		out.Value = float32(computer.computed.Values["font_size"].Value) * -0.5
	default:
		out.Value = length(computer, name, Value(value)).Value
	}
	if value.Unit == "%" {
		//TODO: support
		// height, _ = strutLayout(computer.computed)
		// return height * value.value / 100.
		log.Println("% not supported for vertical-align")
	}
	return VerticalAlign(out)
}

// Compute the ``word-spacing`` property.
func (value WordSpacing) ComputeValue(computer *computer, name string) CssProperty {
	if value.String == "normal" {
		return WordSpacing{}
	}
	return WordSpacing(length(computer, name, Value(value)))
}

func (value CounterResets) ComputeValue(computer *computer, name string) CssProperty     { return value }
func (value CounterIncrements) ComputeValue(computer *computer, name string) CssProperty { return value }
func (value Page) ComputeValue(computer *computer, name string) CssProperty              { return value }
func (value Value) ComputeValue(computer *computer, name string) CssProperty             { return value }
func (value String) ComputeValue(computer *computer, name string) CssProperty            { return value }
func (value Color) ComputeValue(computer *computer, name string) CssProperty             { return value }
func (value Quotes) ComputeValue(computer *computer, name string) CssProperty            { return value }
func (value ListStyleImage) ComputeValue(computer *computer, name string) CssProperty    { return value }
func (value StringSet) ComputeValue(computer *computer, name string) CssProperty         { return value }
func (value StringContent) ComputeValue(computer *computer, name string) CssProperty     { return value }
