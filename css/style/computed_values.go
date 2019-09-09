package style

import (
	"log"
	"sort"
	"strings"

	. "github.com/benoitkugler/go-weasyprint/css"
	"github.com/benoitkugler/go-weasyprint/utils"
	"golang.org/x/net/html"
)

// Convert *specified* property values (the result of the cascade and
//     inheritance) into *computed* values (that are inherited).

// :copyright: Copyright 2011-2014 Simon Sapin and contributors, see AUTHORS.
// :license: BSD, see LICENSE for details.

var (
	// These are unspecified, other than 'thin' <='medium' <= 'thick'.
	// Values are in pixels.
	BorderWidthKeywords = map[string]float32{
		"thin":   1,
		"medium": 3,
		"thick":  5,
	}

	// http://www.w3.org/TR/CSS21/fonts.html#propdef-font-weight
	FontWeightRelative = struct {
		bolder, lighter map[int]int
	}{
		bolder: map[int]int{
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
		lighter: map[int]int{
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

	computerFunctions = map[string]computerFunc{
		"background_image":    backgroundImage,
		"background_position": backgroundPosition,
		"transform-origin":    transformOrigin,

		"border-spacing":             lengths,
		"size":                       lengths,
		"clip":                       lengths,
		"border_top_left_radius":     lengths,
		"border_top_right_radius":    lengths,
		"border_bottom_left_radius":  lengths,
		"border_bottom_right_radius": lengths,

		"break_before": break_,
		"break_after":  break_,

		"top":                  length,
		"right":                length,
		"left":                 length,
		"bottom":               length,
		"margin_top":           length,
		"margin_right":         length,
		"margin_bottom":        length,
		"margin_left":          length,
		"height":               length,
		"width":                length,
		"min_width":            length,
		"min_height":           length,
		"max_width":            length,
		"max_height":           length,
		"padding_top":          length,
		"padding_right":        length,
		"padding_bottom":       length,
		"padding_left":         length,
		"text_indent":          length,
		"hyphenate_limit_zone": length,

		"bleed-left":          bleed,
		"bleed-right":         bleed,
		"bleed-top":           bleed,
		"bleed-bottom":        bleed,
		"letter-spacing":      pixelLength,
		"background_size":     backgroundSize,
		"border_top_width":    borderWidth,
		"border_right_width":  borderWidth,
		"border_left_width":   borderWidth,
		"border_bottom_width": borderWidth,
		"column_rule_width":   borderWidth,
		"outline_width":       borderWidth,
		"column_width":        columnWidth,
		"column_gap":          columnGap,
		"content":             content,
		"display":             display,
		"float":               floating,
		"font_size":           fontSize,
		"font_weight":         fontWeight,
		"line_height":         lineHeight,
		"anchor":              anchor,
		"link":                link,
		"lang":                lang,
		"tab_size":            tabSize,
		"transform":           transforms,
		"vertical_align":      verticalAlign,
		"word_spacing":        wordSpacing,
	}
)

func init() {
	if InitialValues.GetBorderTopWidth().Value != BorderWidthKeywords["medium"] {
		log.Fatal("border-top-width and medium should be the same !")
	}

	// In "portrait" orientation.
	for _, size := range PageSizes {
		if size[0].Value > size[1].Value {
			log.Fatal("page size should be in portrait orientation")
		}
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

type computerFunc = func(*computer, string, CssProperty) CssProperty

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
		parentStyle = StyleDict{Properties: InitialValues}
	}

	computer.element = element
	computer.specified = specified
	computer.computed = computed
	computer.parentStyle = parentStyle
	computer.rootStyle = rootStyle
	computer.baseUrl = baseUrl

	for _, name := range ComputingOrder {
		if _, in := computed.Properties[name]; in {
			// Already computed
			continue
		}

		value := specified.Properties[name]
		fn := computerFunctions[name]
		if fn != nil {
			value = fn(computer, name, value)
		}
		// else: same as specified

		computed.Properties[name] = value
	}

	computed.SetWeasySpecifiedDisplay(specified.GetDisplay())
	return computed
}

type computer struct {
	isRootElement                               bool
	computed, rootStyle, parentStyle, specified StyleDict
	element                                     html.Node
	baseUrl                                     string
}

// backgroundImage computes lenghts in gradient background-image.
func backgroundImage(computer *computer, name string, _value CssProperty) CssProperty {
	value := _value.(Images)
	for i, image := range value {
		switch gradient := image.(type) {
		case LinearGradient:
			for j, pos := range gradient.ColorStops {
				gradient.ColorStops[j].Position = length2(computer, name, Value{Dimension: pos.Position}, -1).Dimension
			}
			image = gradient
		case RadialGradient:
			for j, pos := range gradient.ColorStops {
				gradient.ColorStops[j].Position = length2(computer, name, Value{Dimension: pos.Position}, -1).Dimension
			}
			gradient.Center = _backgroundPosition(computer, name, []Center{gradient.Center})[0]
			if gradient.Size.IsExplicit() {
				l := _lengthOrPercentageTuple2(computer, name, gradient.Size.Explicit.ToSlice())
				gradient.Size.Explicit = Point{l[0], l[1]}
			}
			image = gradient
		}
		value[i] = image
	}
	return value
}

func _backgroundPosition(computer *computer, name string, value Centers) Centers {
	out := make(Centers, len(value))
	for index, v := range value {
		out[index] = Center{
			OriginX: v.OriginX,
			OriginY: v.OriginY,
			Pos: Point{
				length2(computer, name, Value{Dimension: v.Pos[0]}, -1).Dimension,
				length2(computer, name, Value{Dimension: v.Pos[1]}, -1).Dimension,
			},
		}
	}
	return out
}

// backgroundPosition compute lengths in background-position.
func backgroundPosition(computer *computer, name string, _value CssProperty) CssProperty {
	value := _value.(Centers)
	return _backgroundPosition(computer, name, value)
}

func transformOrigin(computer *computer, name string, _value CssProperty) CssProperty {
	value := _value.(Point)
	l := _lengthOrPercentageTuple2(computer, name, value.ToSlice())
	return Point{l[0], l[1]}
}

// Compute the lists of lengths that can be percentages.
func lengths(computer *computer, name string, _value CssProperty) CssProperty {
	value := _value.(Values)
	out := make(Values, len(value))
	for index, v := range value {
		out[index] = length2(computer, name, v, -1)
	}
	return out
}

// Compute the lists of lengths that can be percentages.
func _lengthOrPercentageTuple2(computer *computer, name string, value []Dimension) []Dimension {
	out := make([]Dimension, len(value))
	for index, v := range value {
		out[index] = length2(computer, name, Value{Dimension: v}, -1).Dimension
	}
	return out
}

// Compute the ``break-before`` and ``break-after`` properties.
func break_(_ *computer, _ string, _value CssProperty) CssProperty {
	value := _value.(String)
	// "always" is defined as an alias to "page" in multi-column
	// https://www.w3.org/TR/css3-multicol/#column-breaks
	if value == "always" {
		return String("page")
	}
	return value
}

func length(computer *computer, name string, _value CssProperty) CssProperty {
	value := _value.(Value)
	return length2(computer, name, value, -1)
}

// Computes a length ``value``.
// passing a negative fontSize means null
// Always returns a Value which is interpreted as float32 if Unit is zero.
func length2(computer *computer, _ string, value Value, fontSize float32) Value {
	if value.String == "auto" || value.String == "content" {
		return value
	}
	if value.Value == 0 {
		return ZeroPixels.ToValue()
	}

	unit := value.Unit
	var result float32
	switch unit {
	case Px:
		return value
	case Pt, Pc, In, Cm, Mm, Q:
		// Convert absolute lengths to pixels
		result = value.Value * LengthsToPixels[unit]
	case Em, Ex, Ch, Rem:
		if fontSize < 0 {
			fontSize = computer.computed.GetFontSize().Value
		}
		switch unit {
		// TODO: we dont support 'ex' and 'ch' units for now.
		case Ex, Ch, Em:
			result = value.Value * fontSize
		case Rem:
			result = value.Value * computer.rootStyle.GetFontSize().Value
		default:
			// A percentage or "auto": no conversion needed.
			return value
		}
	}
	return Dimension{Value: result, Unit: Px}.ToValue()
}

func bleed(computer *computer, name string, _value CssProperty) CssProperty {
	value := _value.(Value)
	if value.String == "auto" {
		if computer.computed.GetMarks().Crop {
			return Dimension{Value: 8, Unit: Px}.ToValue() // 6pt
		}
		return ZeroPixels.ToValue()
	}
	return length(computer, name, value)
}

func pixelLength(computer *computer, name string, _value CssProperty) CssProperty {
	value := _value.(Value)
	if value.String == "normal" {
		return value
	}
	return length(computer, name, value)
}

// Compute the ``background-size`` properties.
func backgroundSize(computer *computer, name string, _value CssProperty) CssProperty {
	value := _value.(Sizes)
	out := make(Sizes, len(value))
	for index, v := range value {
		if v.String == "contain" || v.String == "cover" {
			out[index] = Size{String: v.String}
		} else {
			l := _lengthOrPercentageTuple2(computer, name, []Dimension{
				v.Width.Dimension,
				v.Height.Dimension,
			})
			out[index] = Size{
				Width:  Value{Dimension: l[0]},
				Height: Value{Dimension: l[1]},
			}
		}
	}
	return out
}

// Compute the ``border-*-width`` properties.
// value.String may be the string representation of an int
func borderWidth(computer *computer, name string, _value CssProperty) CssProperty {
	value := _value.(Value)
	style := computer.computed.Properties[strings.ReplaceAll(name, "width", "style")].(String)
	if style == "none" || style == "hidden" {
		return Value{}
	}

	if bw, in := BorderWidthKeywords[value.String]; in {
		return FToV(bw)
	}
	return length(computer, name, value)
}

// Compute the ``column-width`` property.
func columnWidth(computer *computer, name string, _value CssProperty) CssProperty {
	value := _value.(Value)
	return length(computer, name, value)
}

// Compute the ``column-gap`` property.
func columnGap(computer *computer, name string, _value CssProperty) CssProperty {
	value := _value.(Value)
	if value.String == "normal" {
		value = Value{Dimension: Dimension{Value: 1, Unit: Em}}
	}
	return length(computer, name, value)
}

// Compute the ``content`` property.
func content(computer *computer, _ string, _value CssProperty) CssProperty {
	value := _value.(SContent)
	if value.String == "normal" || value.String == "none" {
		return value
	}
	lis := make([]ContentProperty, len(value.Contents))
	for index, v := range value.Contents {
		// type_, value := v[0], v[1]
		if v.Type == ContentAttr {
			lis[index].Type = ContentSTRING
			lis[index].String = utils.GetAttribute(computer.element, value.String)
		} else {
			lis[index] = v
		}
	}
	return SContent{Contents: lis}
}

//Compute the ``display`` property.
// See http://www.w3.org/TR/CSS21/visuren.html#dis-pos-flo
func display(computer *computer, _ string, _value CssProperty) CssProperty {
	value := _value.(String)
	float_ := computer.specified.GetFloat()
	position := computer.specified.GetPosition()
	if (position == "absolute" || position == "fixed") || float_ != "none" || computer.isRootElement {
		switch value {
		case "inline-table":
			return String("table")
		case "inline", "table-row-group", "table-column",
			"table-column-group", "table-header-group",
			"table-footer-group", "table-row", "table-cell",
			"table-caption", "inline-block":
			return String("block")
		}
	}
	return value
}

//Compute the ``float`` property.
// See http://www.w3.org/TR/CSS21/visuren.html#dis-pos-flo
func floating(computer *computer, _ string, _value CssProperty) CssProperty {
	value := _value.(String)
	position := computer.specified.GetPosition()
	if position == "absolute" || position == "fixed" {
		return String("none")
	}
	return value
}

// Compute the ``font-size`` property.
func fontSize(computer *computer, name string, _value CssProperty) CssProperty {
	value := _value.(Value)
	if fs, in := FontSizeKeywords[value.String]; in {
		return FToV(fs)
	}
	// TODO: support "larger" and "smaller"

	parentFontSize := computer.parentStyle.GetFontSize().Value
	if value.Unit == Percentage {
		return FToV(value.Value * parentFontSize / 100.)
	}
	return length2(computer, name, value, parentFontSize)
}

// Compute the ``font-weight`` property.
func fontWeight(computer *computer, _ string, _value CssProperty) CssProperty {
	value := _value.(IntString)
	var out int
	switch value.String {
	case "normal":
		out = 400
	case "bold":
		out = 700
	case "bolder":
		parentValue := computer.parentStyle.GetFontWeight().Int
		out = FontWeightRelative.bolder[parentValue]
	case "lighter":
		parentValue := computer.parentStyle.GetFontWeight().Int
		out = FontWeightRelative.lighter[parentValue]
	default:
		out = value.Int
	}
	return IntString{Int: out}
}

// Compute the ``line-height`` property.
func lineHeight(computer *computer, name string, _value CssProperty) CssProperty {
	value := _value.(Value)
	var pixels float32
	switch {
	case value.String == "normal":
		return value
	case value.Unit == NoUnit:
		return value
	case value.Unit == Percentage:
		factor := value.Value / 100.
		fontSizeValue := computer.computed.GetFontSize().Value
		pixels = factor * fontSizeValue
	default:
		pixels = length2(computer, name, value, -1).Value
	}
	return Dimension{Value: pixels, Unit: Px}.ToValue()
}

// Compute the ``anchor`` property.
func anchor(computer *computer, _ string, _value CssProperty) CssProperty {
	value := _value.(NamedString)
	if value.String != "none" {
		s := utils.GetAttribute(computer.element, value.String)
		if s == "" {
			return nil
		}
		return NamedString{String: s}
	}
	return nil
}

// Compute the ``link`` property.
func link(computer *computer, _ string, _value CssProperty) CssProperty {
	value := _value.(NamedString)
	if value.String == "none" {
		return nil
	}
	if value.Name == "attr" {
		type_attr := utils.GetLinkAttribute(computer.element, value.String, computer.baseUrl)
		if len(type_attr) < 2 {
			return nil
		}
		return NamedString{Name: type_attr[0], String: type_attr[1]}
	}
	return nil
}

// Compute the ``lang`` property.
func lang(computer *computer, _ string, _value CssProperty) CssProperty {
	value := _value.(NamedString)
	if value.String == "none" {
		return nil
	}
	if value.Name == "attr" {
		s := utils.GetAttribute(computer.element, value.String)
		if s == "" {
			return nil
		}
		return NamedString{String: s}
	} else if value.Name == "string" {
		return NamedString{String: value.String}
	}
	return nil
}

// Compute the ``tab-size`` property.
func tabSize(computer *computer, name string, _value CssProperty) CssProperty {
	value := _value.(Value)
	if value.Unit == NoUnit {
		return value
	}
	return length(computer, name, value)
}

// Compute the ``transform`` property.
func transforms(computer *computer, name string, _value CssProperty) CssProperty {
	value := _value.(Transforms)
	result := make(Transforms, len(value))
	for index, tr := range value {
		if tr.String == "translate" {
			tr.Dimensions = _lengthOrPercentageTuple2(computer, name, tr.Dimensions)
		}
		result[index] = tr
	}
	return result
}

// Compute the ``vertical-align`` property.
func verticalAlign(computer *computer, name string, _value CssProperty) CssProperty {
	value := _value.(Value)
	// Use +/- half an em for super and sub, same as Pango.
	// (See the SUPERSUBRISE constant in pango-markup.c)
	var out Value
	switch value.String {
	case "baseline", "middle", "text-top", "text-bottom", "top", "bottom":
		out.String = value.String
	case "super":
		out.Value = computer.computed.GetFontSize().Value * 0.5
	case "sub":
		out.Value = computer.computed.GetFontSize().Value * -0.5
	default:
		out.Value = length2(computer, name, value, -1).Value
	}
	if value.Unit == Percentage {
		//TODO: support
		// height, _ = strutLayout(computer.computed)
		// return height * value.value / 100.
		log.Println("% not supported for vertical-align")
	}
	return out
}

// Compute the ``word-spacing`` property.
func wordSpacing(computer *computer, name string, _value CssProperty) CssProperty {
	value := _value.(Value)
	if value.String == "normal" {
		return Value{}
	}
	return length(computer, name, value)
}
