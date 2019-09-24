package style

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/benoitkugler/go-weasyprint/style/validation"

	"github.com/benoitkugler/go-weasyprint/style/parser"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/utils"
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
		"object_position":     backgroundPosition,
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
		"flex_basis":           length,

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
		"lang":                lang,
		"tab_size":            tabSize,
		"transform":           transforms,
		"vertical_align":      verticalAlign,
		"word_spacing":        wordSpacing,
	}
	computerFromSpe = map[string]computerFuncSpe{
		"link": link,
	}
)

func init() {
	if pr.InitialValues.GetBorderTopWidth().Value != BorderWidthKeywords["medium"] {
		log.Fatal("border-top-width and medium should be the same !")
	}

	// In "portrait" orientation.
	for _, size := range pr.PageSizes {
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
	keys := pr.InitialValues.Keys()
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	for _, k := range keys {
		if !crible[k] {
			ComputingOrder = append(ComputingOrder, k)
		}
	}
}

type computerFunc = func(*computer, string, pr.CssProperty) pr.CssProperty
type computerFuncSpe = func(*computer, string, pr.CascadedProperty) pr.CssProperty

func resolveVar(computed map[string]pr.CascadedProperty, var_ pr.VarData) pr.CustomProperty {
	knownVariableNames := pr.NewSet(var_.Name)
	default_ := var_.Declaration
	computedValue_, isIn := computed[var_.Name]
	computedValue, ok := computedValue_.SpecialProperty.(pr.CustomProperty)
	if ok && len(computedValue) == 1 {
		value := computedValue[0]
		if ident, ok := value.(parser.IdentToken); ok && ident.Value == "initial" {
			return default_
		}
	}
	if !isIn {
		computedValue = default_
	}
	var in bool
	for len(computedValue) == 1 {
		varFunction := validation.CheckVarFunction(computedValue[0])
		if !varFunction.IsNone() {
			if knownVariableNames.Has(varFunction.Name) {
				computedValue = default_
				break
			}
			knownVariableNames.Add(varFunction.Name)
			computedValue_, in = computed[varFunction.Name]
			if !in {
				computedValue = varFunction.Declaration
			} else {
				computedValue, _ = computedValue_.SpecialProperty.(pr.CustomProperty)
			}
			default_ = varFunction.Declaration
		} else {
			break
		}
	}
	return computedValue
}

// Return a dict of computed value.

// :param element: The HTML element these style apply to
// :param specified: a dict of specified value. Should contain
// 			  value for all pr.
// :param computed: a dict of already known computed value.
// 			 Only contains some properties (or none).
// :param parentStyle: a dict of computed value of the parent
// 				 element (should contain value for all properties),
// 				 or `zero if ``element`` is the root element.
// :param baseUrl: The base URL used to resolve relative URLs.
// 		targetCollector: A target collector used to get computed targets.
func compute(element *utils.HTMLNode, specified, computed map[string]pr.CascadedProperty, parentStyle,
	rootStyle StyleDict, baseUrl string, targetCollector *targetCollector) (StyleDict, error) {

	if parentStyle.IsZero() {
		parentStyle = StyleDict{Properties: pr.InitialValues}
	}

	computer := &computer{
		isRootElement:   parentStyle.IsZero(),
		element:         element,
		specified:       specified,
		computed:        computed,
		parentStyle:     parentStyle,
		rootStyle:       rootStyle,
		baseUrl:         baseUrl,
		targetCollector: targetCollector,
	}
	out := make(pr.Properties, len(ComputingOrder))
	for _, name := range ComputingOrder {
		if _, in := computed[name]; in {
			// Already computed
			continue
		}

		value := specified[name]

		if var_, ok := value.SpecialProperty.(pr.VarData); ok {
			computedValue := resolveVar(computed, var_)
			var newValue pr.CascadedProperty
			if computedValue != nil {
				var err error
				newValue, err = validation.Validate(strings.ReplaceAll(name, "_", "-"), computedValue, baseUrl)
				if err != nil {
					return StyleDict{}, err
				}
			}

			// See https://drafts.csswg.org/css-variables/#invalid-variables
			if newValue.IsNone() {
				chunks := make([]string, len(computedValue))
				for i, token := range computedValue {
					chunks[i] = parser.SerializeOne(token)
				}
				cp := strings.Join(chunks, "")
				log.Printf("Unsupported computed value `%s` set in variable `%s` for property `%s`.", cp,
					strings.ReplaceAll(var_.Name, "_", "-"), strings.ReplaceAll(name, "_", "-"))
				if pr.Inherited.Has(name) && !parentStyle.IsZero() {
					value = pr.ToC(parentStyle.Properties[name])
				} else {
					value = pr.ToC(pr.InitialValues[name])
				}
			} else {
				value = newValue
			}

		}

		fnSpe := computerFromSpe[name]
		fn := computerFunctions[name]
		var finalValue pr.CssProperty
		if fnSpe != nil {
			finalValue = fnSpe(computer, name, value)
		} else {
			if value.SpecialProperty != nil {
				return StyleDict{}, fmt.Errorf("specified value not resolved : %v", value)
			}
			finalValue = value.AsCss()
			if fn != nil {
				finalValue = fn(computer, name, finalValue)
			}
			// else: same as specified
		}

		computed[name] = pr.ToC(finalValue)
		out[name] = finalValue
	}
	disp := specified["display"]
	if disp.SpecialProperty != nil {
		return StyleDict{}, fmt.Errorf("specified display value not resolved : %v", disp)
	}
	out["_weasy_specified_display"] = disp.AsCss()
	return StyleDict{Properties: out}, nil
}

type computer struct {
	isRootElement          bool
	computed, specified    map[string]pr.CascadedProperty
	rootStyle, parentStyle StyleDict
	element                *utils.HTMLNode
	baseUrl                string
	targetCollector        *targetCollector
}

// backgroundImage computes lenghts in gradient background-image.
func backgroundImage(computer *computer, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Images)
	for i, image := range value {
		switch gradient := image.(type) {
		case pr.LinearGradient:
			for j, pos := range gradient.ColorStops {
				gradient.ColorStops[j].Position = length2(computer, name, pr.Value{Dimension: pos.Position}, -1).Dimension
			}
			image = gradient
		case pr.RadialGradient:
			for j, pos := range gradient.ColorStops {
				gradient.ColorStops[j].Position = length2(computer, name, pr.Value{Dimension: pos.Position}, -1).Dimension
			}
			gradient.Center = _backgroundPosition(computer, name, []pr.Center{gradient.Center})[0]
			if gradient.Size.IsExplicit() {
				l := _lengthOrPercentageTuple2(computer, name, gradient.Size.Explicit.ToSlice())
				gradient.Size.Explicit = pr.Point{l[0], l[1]}
			}
			image = gradient
		}
		value[i] = image
	}
	return value
}

func _backgroundPosition(computer *computer, name string, value pr.Centers) pr.Centers {
	out := make(pr.Centers, len(value))
	for index, v := range value {
		out[index] = pr.Center{
			OriginX: v.OriginX,
			OriginY: v.OriginY,
			Pos: pr.Point{
				length2(computer, name, pr.Value{Dimension: v.Pos[0]}, -1).Dimension,
				length2(computer, name, pr.Value{Dimension: v.Pos[1]}, -1).Dimension,
			},
		}
	}
	return out
}

// backgroundPosition compute lengths in background-position.
func backgroundPosition(computer *computer, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Centers)
	return _backgroundPosition(computer, name, value)
}

func transformOrigin(computer *computer, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Point)
	l := _lengthOrPercentageTuple2(computer, name, value.ToSlice())
	return pr.Point{l[0], l[1]}
}

// Compute the lists of lengths that can be percentages.
func lengths(computer *computer, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Values)
	out := make(pr.Values, len(value))
	for index, v := range value {
		out[index] = length2(computer, name, v, -1)
	}
	return out
}

// Compute the lists of lengths that can be percentages.
func _lengthOrPercentageTuple2(computer *computer, name string, value []pr.Dimension) []pr.Dimension {
	out := make([]pr.Dimension, len(value))
	for index, v := range value {
		out[index] = length2(computer, name, pr.Value{Dimension: v}, -1).Dimension
	}
	return out
}

// Compute the ``break-before`` and ``break-after`` pr.
func break_(_ *computer, _ string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.String)
	// "always" is defined as an alias to "page" in multi-column
	// https://www.w3.org/TR/css3-multicol/#column-breaks
	if value == "always" {
		return pr.String("page")
	}
	return value
}

func length(computer *computer, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	return length2(computer, name, value, -1)
}

// Computes a length ``value``.
// passing a negative fontSize means null
// Always returns a Value which is interpreted as float32 if Unit is zero.
func length2(computer *computer, _ string, value pr.Value, fontSize float32) pr.Value {
	if value.String == "auto" || value.String == "content" {
		return value
	}
	if value.Value == 0 {
		return pr.ZeroPixels.ToValue()
	}

	unit := value.Unit
	var result float32
	switch unit {
	case pr.Px:
		return value
	case pr.Pt, pr.Pc, pr.In, pr.Cm, pr.Mm, pr.Q:
		// Convert absolute lengths to pixels
		result = value.Value * pr.LengthsToPixels[unit]
	case pr.Em, pr.Ex, pr.Ch, pr.Rem:
		if fontSize < 0 {
			fs := computer.computed["font_size"]
			fs_, _ := fs.AsCss().(pr.Float)
			fontSize = float32(fs_)
		}
		switch unit {
		// TODO: we dont support 'ex' and 'ch' units for now.
		case pr.Ex, pr.Ch, pr.Em:
			result = value.Value * fontSize
		case pr.Rem:
			result = value.Value * computer.rootStyle.GetFontSize().Value
		default:
			// A percentage or "auto": no conversion needed.
			return value
		}
	}
	return pr.Dimension{Value: result, Unit: pr.Px}.ToValue()
}

func bleed(computer *computer, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	if value.String == "auto" {
		if computer.computed["marks"].AsCss().(pr.Marks).Crop {
			return pr.Dimension{Value: 8, Unit: pr.Px}.ToValue() // 6pt
		}
		return pr.ZeroPixels.ToValue()
	}
	return length(computer, name, value)
}

func pixelLength(computer *computer, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	if value.String == "normal" {
		return value
	}
	return length(computer, name, value)
}

// Compute the ``background-size`` pr.
func backgroundSize(computer *computer, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Sizes)
	out := make(pr.Sizes, len(value))
	for index, v := range value {
		if v.String == "contain" || v.String == "cover" {
			out[index] = pr.Size{String: v.String}
		} else {
			l := _lengthOrPercentageTuple2(computer, name, []pr.Dimension{
				v.Width.Dimension,
				v.Height.Dimension,
			})
			out[index] = pr.Size{
				Width:  pr.Value{Dimension: l[0]},
				Height: pr.Value{Dimension: l[1]},
			}
		}
	}
	return out
}

// Compute the ``border-*-width`` pr.
// value.String may be the string representation of an int
func borderWidth(computer *computer, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	style := computer.computed[strings.ReplaceAll(name, "width", "style")].AsCss().(pr.String)
	if style == "none" || style == "hidden" {
		return pr.Value{}
	}

	if bw, in := BorderWidthKeywords[value.String]; in {
		return pr.FToV(bw)
	}
	return length(computer, name, value)
}

// Compute the ``column-width`` property.
func columnWidth(computer *computer, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	return length(computer, name, value)
}

// Compute the ``column-gap`` property.
func columnGap(computer *computer, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	if value.String == "normal" {
		value = pr.Value{Dimension: pr.Dimension{Value: 1, Unit: pr.Em}}
	}
	return length(computer, name, value)
}

func computeAttrFunction(computer *computer, values pr.ContentProperty) (string, pr.CssProperty, error) {
	// TODO: use real token parsing instead of casting with Python types
	value, ok := values.Content.(Attr)
	if values.Type != "attr()" || !ok {
		log.Fatalf("values should be attr() here")
	}
	attrName, typeOrUnit, fallback := value.Name, value.TypeOrUnit, value.Fallback
	// computer["element"] sometimes is None
	// computer["element"] sometimes is a "PageType" object without .get()
	if computer.element == nil {
		return "", nil, nil
	}

	attrValue := computer.element.Get(attrName)
	if attrValue == "" {
		attrValue = string(fallback.Value.(pr.String))
	}
	var out pr.CssProperty
	switch typeOrUnit {
	case "string":
		out = pr.String(attrValue) // Keep the string
	case "url":
		if strings.HasPrefix(attrValue, "#") {
			out = pr.NamedString{Name: "internal", String: utils.Unquote(attrValue[1:])}
		} else {
			u, err := utils.SafeUrljoin(computer.baseUrl, attrValue, false)
			if err != nil {
				return "", nil, err
			}
			out = pr.NamedString{Name: "external", String: u}
		}
	case "color":
		out = pr.Color(parser.ParseColor2(strings.TrimSpace(attrValue)))
	case "integer":
		i, err := strconv.Atoi(strings.TrimSpace(attrValue))
		if err != nil {
			return "", nil, err
		}
		out = pr.Int(i)
	case "number":
		f, err := strconv.ParseFloat(strings.TrimSpace(attrValue), 32)
		if err != nil {
			return "", nil, err
		}
		out = pr.Float(f)
	case "%":
		f, err := strconv.ParseFloat(strings.TrimSpace(attrValue), 32)
		if err != nil {
			return "", nil, err
		}
		out = pr.Dimension{Value: float32(f), Unit: pr.Percentage}.ToValue()
		typeOrUnit = "length"
	default:
		unit, isUnit := validation.LENGTHUNITS[typeOrUnit]
		angle, isAngle := validation.AngleUnits[typeOrUnit]
		if isUnit {
			f, err := strconv.ParseFloat(strings.TrimSpace(attrValue), 32)
			if err != nil {
				return "", nil, err
			}
			out = pr.Dimension{Value: float32(f), Unit: unit}.ToValue()
			typeOrUnit = "length"
		} else if isAngle {
			f, err := strconv.ParseFloat(strings.TrimSpace(attrValue), 32)
			if err != nil {
				return "", nil, err
			}
			out = pr.Dimension{Value: float32(f), Unit: angle}.ToValue()
			typeOrUnit = "angle"
		}
	}
	return typeOrUnit, out, nil
}

// Compute the ``content`` property.
func content(computer *computer, _ string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.SContent)
	if value.String == "normal" || value.String == "none" {
		return value
	}
	lis := make([]pr.ContentProperty, len(value.Contents))
	for index, v := range value.Contents {
		// type_, value := v[0], v[1]
		if v.Type == ContentAttr {
			lis[index].Type = ContentSTRING
			lis[index].String = computer.element.Get(value.String)
		} else {
			lis[index] = v
		}
	}
	return pr.SContent{Contents: lis}
}

//Compute the ``display`` property.
// See http://www.w3.org/TR/CSS21/visuren.html#dis-pos-flo
func display(computer *computer, _ string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.String)
	float_ := computer.specified.GetFloat()
	position := computer.specified.GetPosition()
	if (position == "absolute" || position == "fixed") || float_ != "none" || computer.isRootElement {
		switch value {
		case "inline-table":
			return pr.String("table")
		case "inline", "table-row-group", "table-column",
			"table-column-group", "table-header-group",
			"table-footer-group", "table-row", "table-cell",
			"table-caption", "inline-block":
			return pr.String("block")
		}
	}
	return value
}

//Compute the ``float`` property.
// See http://www.w3.org/TR/CSS21/visuren.html#dis-pos-flo
func floating(computer *computer, _ string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.String)
	position := computer.specified.GetPosition()
	if position == "absolute" || position == "fixed" {
		return pr.String("none")
	}
	return value
}

// Compute the ``font-size`` property.
func fontSize(computer *computer, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	if fs, in := pr.FontSizeKeywords[value.String]; in {
		return pr.FToV(fs)
	}
	// TODO: support "larger" and "smaller"

	parentFontSize := computer.parentStyle.GetFontSize().Value
	if value.Unit == pr.Percentage {
		return pr.FToV(value.Value * parentFontSize / 100.)
	}
	return length2(computer, name, value, parentFontSize)
}

// Compute the ``font-weight`` property.
func fontWeight(computer *computer, _ string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.IntString)
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
	return pr.IntString{Int: out}
}

// Compute the ``line-height`` property.
func lineHeight(computer *computer, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	var pixels float32
	switch {
	case value.String == "normal":
		return value
	case value.Unit == NoUnit:
		return value
	case value.Unit == pr.Percentage:
		factor := value.Value / 100.
		fontSizeValue := computer.computed.GetFontSize().Value
		pixels = factor * fontSizeValue
	default:
		pixels = length2(computer, name, value, -1).Value
	}
	return pr.Dimension{Value: pixels, Unit: pr.Px}.ToValue()
}

// Compute the ``anchor`` property.
func anchor(computer *computer, _ string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.NamedString)
	if value.String != "none" {
		s := computer.element.Get(value.String)
		if s == "" {
			return nil
		}
		return pr.NamedString{String: s}
	}
	return nil
}

// Compute the ``link`` property.
func link(computer *computer, _ string, _value pr.CascadedProperty) pr.CssProperty {
	value := _value.(pr.NamedString)
	if value.String == "none" {
		return nil
	}
	if value.Name == "attr" {
		type_attr := utils.GetLinkAttribute(*computer.element, value.String, computer.baseUrl)
		if len(type_attr) < 2 {
			return nil
		}
		return pr.NamedString{Name: type_attr[0], String: type_attr[1]}
	}
	return nil
}

// Compute the ``lang`` property.
func lang(computer *computer, _ string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.NamedString)
	if value.String == "none" {
		return nil
	}
	if value.Name == "attr" {
		s := computer.element.Get(value.String)
		if s == "" {
			return nil
		}
		return pr.NamedString{String: s}
	} else if value.Name == "string" {
		return pr.NamedString{String: value.String}
	}
	return nil
}

// Compute the ``tab-size`` property.
func tabSize(computer *computer, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	if value.Unit == NoUnit {
		return value
	}
	return length(computer, name, value)
}

// Compute the ``transform`` property.
func transforms(computer *computer, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Transforms)
	result := make(pr.Transforms, len(value))
	for index, tr := range value {
		if tr.String == "translate" {
			tr.Dimensions = _lengthOrPercentageTuple2(computer, name, tr.Dimensions)
		}
		result[index] = tr
	}
	return result
}

// Compute the ``vertical-align`` property.
func verticalAlign(computer *computer, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	// Use +/- half an em for super and sub, same as Pango.
	// (See the SUPERSUBRISE constant in pango-markup.c)
	var out pr.Value
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
	if value.Unit == pr.Percentage {
		//TODO: support
		// height, _ = strutLayout(computer.computed)
		// return height * value.value / 100.
		log.Println("% not supported for vertical-align")
	}
	return out
}

// Compute the ``word-spacing`` property.
func wordSpacing(computer *computer, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	if value.String == "normal" {
		return pr.Value{}
	}
	return length(computer, name, value)
}
