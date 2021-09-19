package tree

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
	BorderWidthKeywords = map[string]pr.Float{
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
		"border_top_left_radius":     borderRadius,
		"border_top_right_radius":    borderRadius,
		"border_bottom_left_radius":  borderRadius,
		"border_bottom_right_radius": borderRadius,

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
		"bookmark_label":      bookmarkLabel,
		"string_set":          stringSet,
		"link":                link,
	}

	keywordsValues []pr.Float
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

	// Some computed value are required by others, so order matters.
	ComputingOrder = []string{
		"font_stretch", "font_weight", "font_family", "font_variant",
		"font_style", "font_size", "line_height", "marks",
	}
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

	keywordsValues = make([]pr.Float, len(pr.FontSizeKeywords))
	for i, k := range pr.FontSizeKeywordsOrder {
		keywordsValues[i] = pr.FontSizeKeywords[k]
	}
}

type computerFunc = func(*computer, string, pr.CssProperty) pr.CssProperty

func resolveVar(specified map[string]pr.CascadedProperty, var_ pr.VarData) pr.CustomProperty {
	knownVariableNames := utils.NewSet(var_.Name)
	default_ := var_.Declaration
	computedValue_, isIn := specified[var_.Name]
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
			computedValue_, in = specified[varFunction.Name]
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

// Return a dict of computed value for the given element from
// `specified`, which must contain values for all the properties.
// `computed` will be updated with the missing properties
// `parentStyle` is a dict of computed value of the parent
//  element (should contain value for all properties), or nil for the root element.
// `targetCollector` is used to get computed targets.
func compute(element Element, pseudoType string, specified map[string]pr.CascadedProperty, computed, parentStyle,
	rootStyle pr.Properties, baseUrl string, targetCollector *TargetCollector) {
	if parentStyle == nil {
		parentStyle = pr.InitialValues
	}

	computer := &computer{
		isRootElement:   parentStyle == nil,
		element:         element,
		pseudoType:      pseudoType,
		specified:       specified,
		computed:        computed,
		parentStyle:     parentStyle,
		rootStyle:       rootStyle,
		baseUrl:         baseUrl,
		TargetCollector: targetCollector,
	}
	// To have better typing, we start by resolving all var() and custom propperties.
	resolveds := make(pr.Properties, len(ComputingOrder))
	for _, name := range ComputingOrder {
		value := specified[name]
		if var_, ok := value.SpecialProperty.(pr.VarData); ok {
			computedValue := resolveVar(specified, var_)
			var newValue pr.CssProperty
			if computedValue != nil {
				val, err := validation.Validate(strings.ReplaceAll(name, "_", "-"), computedValue, baseUrl)
				if err != nil {
					log.Printf("Unsupported computed value set in variable `%s` for property `%s`.",
						strings.ReplaceAll(var_.Name, "_", "-"), strings.ReplaceAll(name, "_", "-"))
					continue
				}
				if val.Default != 0 || val.AsCascaded().SpecialProperty != nil {
					log.Printf("Unsupported 'inherited' or 'initial' set in variable `%s` for property `%s`",
						strings.ReplaceAll(var_.Name, "_", "-"), strings.ReplaceAll(name, "_", "-"))
				}
				newValue = val.AsCascaded().AsCss()
			}

			// See https://drafts.csswg.org/css-variables/#invalid-variables
			if newValue == nil {
				chunks := make([]string, len(computedValue))
				for i, token := range computedValue {
					chunks[i] = parser.SerializeOne(token)
				}
				cp := strings.Join(chunks, "")
				log.Printf("Unsupported computed value `%s` set in variable `%s` for property `%s`.", cp,
					strings.ReplaceAll(var_.Name, "_", "-"), strings.ReplaceAll(name, "_", "-"))
				if pr.Inherited.Has(name) && parentStyle != nil {
					newValue = parentStyle[name]
				} else {
					newValue = pr.InitialValues[name]
				}
			}
			resolveds[name] = newValue
		} else if _, ok := value.SpecialProperty.(pr.CustomProperty); ok {
			log.Printf("unexpected custom property for %s\n", name)
		} else {
			resolveds[name] = value.AsCss()
		}
	}

	for _, name := range ComputingOrder {

		if _, in := computed[name]; in {
			// Already computed
			continue
		}
		value := resolveds[name]

		fn := computerFunctions[name]
		if fn != nil {
			value = fn(computer, name, value)
		}
		// else: same as specified
		computed[name] = value
	}
	computed.SetWeasySpecifiedDisplay(resolveds.GetDisplay())
}

type computer struct {
	element         Element
	computed        pr.Properties
	specified       map[string]pr.CascadedProperty
	rootStyle       pr.Properties
	parentStyle     pr.Properties
	TargetCollector *TargetCollector
	pseudoType      string
	baseUrl         string
	isRootElement   bool
}

// backgroundImage computes lenghts in gradient background-image.
func backgroundImage(computer *computer, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Images)
	for i, image := range value {
		switch gradient := image.(type) {
		case pr.LinearGradient:
			for j, pos := range gradient.ColorStops {
				gradient.ColorStops[j].Position = length2(computer, name, pr.Value{Dimension: pos.Position}, -1, false).Dimension
			}
			image = gradient
		case pr.RadialGradient:
			for j, pos := range gradient.ColorStops {
				gradient.ColorStops[j].Position = length2(computer, name, pr.Value{Dimension: pos.Position}, -1, false).Dimension
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
				length2(computer, name, pr.Value{Dimension: v.Pos[0]}, -1, false).Dimension,
				length2(computer, name, pr.Value{Dimension: v.Pos[1]}, -1, false).Dimension,
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
		out[index] = length2(computer, name, v, -1, true)
	}
	return out
}

func borderRadius(computer *computer, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Values)
	out := make(pr.Values, len(value))
	for index, v := range value {
		out[index] = length2(computer, name, v, -1, false)
	}
	return out
}

// Compute the lists of lengths that can be percentages.
func _lengthOrPercentageTuple2(computer *computer, name string, value []pr.Dimension) []pr.Dimension {
	out := make([]pr.Dimension, len(value))
	for index, v := range value {
		out[index] = length2(computer, name, pr.Value{Dimension: v}, -1, false).Dimension
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
	return length2(computer, name, value, -1, false)
}

func asPixels(v pr.Value, pixelsOnly bool) pr.Value {
	if pixelsOnly {
		v.Unit = pr.Scalar
	}
	return v
}

// Computes a length ``value``.
// passing a negative fontSize means null
// Always returns a Value which is interpreted as float64 if Unit is zero.
// pixelsOnly=false
func length2(computer *computer, _ string, value pr.Value, fontSize pr.Float, pixelsOnly bool) pr.Value {
	if value.String == "auto" || value.String == "content" {
		return value
	}
	if value.Value == 0 {
		return asPixels(pr.ZeroPixels.ToValue(), pixelsOnly)
	}

	unit := value.Unit
	var result pr.Float
	switch unit {
	case pr.Px:
		return asPixels(value, pixelsOnly)
	case pr.Pt, pr.Pc, pr.In, pr.Cm, pr.Mm, pr.Q:
		// Convert absolute lengths to pixels
		result = value.Value * pr.LengthsToPixels[unit]
	case pr.Em, pr.Ex, pr.Ch, pr.Rem:
		if fontSize < 0 {
			fontSize = computer.computed.GetFontSize().Value
		}
		switch unit {
		// TODO: we dont support 'ex' and 'ch' units for now.
		case pr.Ex, pr.Ch, pr.Em:
			result = value.Value * fontSize
		case pr.Rem:
			result = value.Value * computer.rootStyle.GetFontSize().Value
		}

	default:
		// A percentage or "auto": no conversion needed.
		return value
	}
	return asPixels(pr.Dimension{Value: result, Unit: pr.Px}.ToValue(), pixelsOnly)
}

func bleed(computer *computer, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	if value.String == "auto" {
		if computer.computed.GetMarks().Crop {
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
	style := computer.computed[strings.ReplaceAll(name, "width", "style")].(pr.String)

	if style == "none" || style == "hidden" {
		return pr.FToV(0)
	}
	if bw, in := BorderWidthKeywords[value.String]; in {
		return bw.ToValue()
	}
	d := length2(computer, name, value, -1, true)
	return d
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

func computeAttrFunction(computer *computer, values pr.AttrData) (out pr.ContentProperty, err error) {
	// TODO: use real token parsing instead of casting with Python types

	attrName, typeOrUnit, fallback := values.Name, values.OrUnitT, values.Fallback
	// computer["Element"] sometimes is None
	// computer["Element"] sometimes is a "PageType" object without .get()
	node, ok := computer.element.(*utils.HTMLNode)
	if !ok {
		return
	}

	var prop pr.InnerContent
	attrValue := node.Get(attrName)
	if attrValue == "" {
		atrValue_, ok := fallback.(pr.String)
		if !ok {
			return out, fmt.Errorf("fallback type not supported : %t", fallback)
		}
		prop = atrValue_
	}
	switch typeOrUnit {
	case "string":
		prop = pr.String(attrValue) // Keep the string
	case "url":
		if strings.HasPrefix(attrValue, "#") {
			prop = pr.NamedString{Name: "internal", String: utils.Unquote(attrValue[1:])}
		} else {
			u, err := utils.SafeUrljoin(computer.baseUrl, attrValue, false)
			if err != nil {
				return out, err
			}
			prop = pr.NamedString{Name: "external", String: u}
		}
	case "color":
		prop = pr.Color(parser.ParseColor2(strings.TrimSpace(attrValue)))
	case "integer":
		i, err := strconv.Atoi(strings.TrimSpace(attrValue))
		if err != nil {
			return out, err
		}
		prop = pr.Int(i)
	case "number":
		f, err := strconv.ParseFloat(strings.TrimSpace(attrValue), 64)
		if err != nil {
			return out, err
		}
		prop = pr.Float(f)
	case "%":
		f, err := strconv.ParseFloat(strings.TrimSpace(attrValue), 64)
		if err != nil {
			return out, err
		}
		prop = pr.Dimension{Value: pr.Float(f), Unit: pr.Percentage}.ToValue()
		typeOrUnit = "length"
	default:
		unit, isUnit := validation.LENGTHUNITS[typeOrUnit]
		angle, isAngle := validation.AngleUnits[typeOrUnit]
		if isUnit {
			f, err := strconv.ParseFloat(strings.TrimSpace(attrValue), 64)
			if err != nil {
				return out, err
			}
			prop = pr.Dimension{Value: pr.Float(f), Unit: unit}.ToValue()
			typeOrUnit = "length"
		} else if isAngle {
			f, err := strconv.ParseFloat(strings.TrimSpace(attrValue), 64)
			if err != nil {
				return out, err
			}
			prop = pr.Dimension{Value: pr.Float(f), Unit: angle}.ToValue()
			typeOrUnit = "angle"
		}
	}
	return pr.ContentProperty{Type: typeOrUnit, Content: prop}, nil
}

func contentList(computer *computer, values pr.ContentProperties) (pr.ContentProperties, error) {
	var computedValues pr.ContentProperties
	for _, value := range values {
		var computedValue pr.ContentProperty
		switch value.Type {
		case "string", "content", "url", "quote", "leader()":
			computedValue = value
		case "attr()":
			attr, ok := value.Content.(pr.AttrData)
			if !ok || attr.OrUnitT != "string" {
				log.Fatalf("invalid attr() property : %v", value.Content)
			}
			var err error
			computedValue, err = computeAttrFunction(computer, attr)
			if err != nil {
				return nil, err
			}
		case "counter()", "counters()", "content()", "Element()", "string()":
			// Other values need layout context, their computed value cannot be
			// better than their specified value yet.
			// See build.computeContentList.
			computedValue = value
		case "target-counter()", "target-counters()", "target-text()":
			prop, ok := value.Content.(pr.SContentProps)
			if !ok || len(prop) == 0 {
				return nil, fmt.Errorf("expected a non empty list of String or ContentProperty, got %v", value.Content)
			}
			anchorToken := prop[0].ContentProperty
			if anchorToken.Type == "attr()" {
				proper, err := computeAttrFunction(computer, anchorToken.Content.(pr.AttrData))
				if err != nil {
					return nil, err
				}
				if !proper.IsNone() {
					computedValue = pr.ContentProperty{Type: value.Type, Content: append(pr.SContentProps{{ContentProperty: proper}}, prop[1:]...)}
				}
			} else {
				computedValue = value
			}
			if computer.TargetCollector != nil && computedValue.Content != nil {
				props := computedValue.Content.(pr.SContentProps) // here, this assertion is always fullfilled
				computer.TargetCollector.collectComputedTarget(props[0].ContentProperty)
			}
		}
		if computedValue.IsNone() {
			log.Printf("Unable to compute %v's value for content: %v\n", computer.element, value)
		} else {
			computedValues = append(computedValues, computedValue)
		}
	}
	return computedValues, nil
}

// Compute the ``bookmark-label`` property.
func bookmarkLabel(computer *computer, _ string, _value pr.CssProperty) pr.CssProperty {
	if value, ok := _value.(pr.ContentProperties); ok {
		out, err := contentList(computer, value)
		if err != nil {
			log.Printf("error computing bookmark-label : %s\n", err)
			return nil
		}
		return out
	}
	return nil
}

// Compute the ``string-set`` property.
func stringSet(computer *computer, _ string, _value pr.CssProperty) pr.CssProperty {
	// Spec asks for strings after custom keywords, but we allow content-lists
	if stringset, ok := _value.(pr.StringSet); ok {
		out := make(pr.SContents, len(stringset.Contents))
		for i, sset := range stringset.Contents {
			v, err := contentList(computer, sset.Contents)
			if err != nil {
				log.Printf("error computing string-set : %s \n", err)
				return nil
			}
			out[i] = pr.SContent{String: sset.String, Contents: v}
		}
		return pr.StringSet{String: stringset.String, Contents: out}
	}
	return nil
}

// Compute the ``content`` property.
func content(computer *computer, _ string, _value pr.CssProperty) pr.CssProperty {
	if value, ok := _value.(pr.SContent); ok {
		if value.String == "normal" {
			if computer.pseudoType != "" {
				return pr.SContent{String: "inhibit"}
			} else {
				return pr.SContent{String: "contents"}
			}
		} else if value.String == "none" {
			return pr.SContent{String: "inhibit"}
		}
		props, err := contentList(computer, value.Contents)
		if err != nil {
			log.Printf("error computing content : %s\n", err)
			return nil
		}
		return pr.SContent{Contents: props}
	}
	return nil
}

// Compute the ``display`` property.
// See http://www.w3.org/TR/CSS21/visuren.html#dis-pos-flo
func display(computer *computer, _ string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.String)
	float_ := computer.specified["float"].AsCss().(pr.String)
	position := computer.specified["position"].AsCss().(pr.BoolString)
	if (!position.Bool && (position.String == "absolute" || position.String == "fixed")) || float_ != "none" || computer.isRootElement {
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

// Compute the ``float`` property.
// See http://www.w3.org/TR/CSS21/visuren.html#dis-pos-flo
func floating(computer *computer, _ string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.String)
	position := computer.specified["position"].AsCss().(pr.BoolString)
	if position.String == "absolute" || position.String == "fixed" {
		return pr.String("none")
	}
	return value
}

// Compute the ``font-size`` property.
func fontSize(computer *computer, name string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.Value)
	if fs, in := pr.FontSizeKeywords[value.String]; in {
		return fs.ToValue()
	}

	parentFontSize := computer.parentStyle.GetFontSize().Value
	if value.String == "larger" {
		for _, keywordValue := range keywordsValues {
			if keywordValue > parentFontSize {
				return keywordValue.ToValue()
			}
		}
		return (parentFontSize * 1.2).ToValue()
	} else if value.String == "smaller" {
		for i := len(keywordsValues) - 1; i >= 0; i -= 1 {
			if keywordsValues[i] < parentFontSize {
				return (keywordsValues[i]).ToValue()
			}
		}
		return (parentFontSize * 0.8).ToValue()
	} else if value.Unit == pr.Percentage {
		return (value.Value * parentFontSize / 100.).ToValue()
	} else {
		return length2(computer, name, value, parentFontSize, true)
	}
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
	var pixels pr.Float
	switch {
	case value.String == "normal":
		return value
	case value.Unit == pr.Scalar:
		return value
	case value.Unit == pr.Percentage:
		factor := value.Value / 100.
		fontSizeValue := computer.computed.GetFontSize().Value
		pixels = factor * fontSizeValue
	default:
		pixels = length2(computer, name, value, -1, true).Value
	}
	return pr.Dimension{Value: pixels, Unit: pr.Px}.ToValue()
}

// Compute the ``anchor`` property.
func anchor(computer *computer, _ string, _value pr.CssProperty) pr.CssProperty {
	value := string(_value.(pr.String))
	if node, ok := computer.element.(*utils.HTMLNode); ok && value != "none" {
		anchorName := node.Get(value)
		if anchorName == "" {
			return nil
		}
		computer.TargetCollector.collectAnchor(anchorName)
		return pr.NamedString{String: anchorName}
	}
	return nil
}

// Compute the ``link`` property.
func link(computer *computer, _ string, _value pr.CssProperty) pr.CssProperty {
	switch value := _value.(type) {
	case pr.NamedString:
		if value.String == "none" {
			return nil
		} else {
			return value
		}

	case pr.AttrData:
		if node, ok := computer.element.(*utils.HTMLNode); ok {
			type_attr := utils.GetLinkAttribute(*node, value.Name, computer.baseUrl)
			if len(type_attr) < 2 {
				return nil
			}
			return pr.NamedString{Name: type_attr[0], String: type_attr[1]}
		}
	}
	return nil
}

// Compute the ``lang`` property.
func lang(computer *computer, _ string, _value pr.CssProperty) pr.CssProperty {
	value := _value.(pr.NamedString)
	if value.String == "none" {
		return nil
	}
	if node, ok := computer.element.(*utils.HTMLNode); ok && value.Name == "attr" {
		s := node.Get(value.String)
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
	if value.Unit == pr.Scalar {
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
		out.Value = length2(computer, name, value, -1, true).Value
	}
	if value.Unit == pr.Percentage {
		// TODO: support
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
