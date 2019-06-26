package css

import "golang.org/x/net/html"

const (
	Top    Side = "top"
	Bottom Side = "bottom"
	Left   Side = "left"
	Right  Side = "right"
)

var (
	stringsKeys = Set{
		"bottom":                   true,
		"caption_side":             true,
		"clear":                    true,
		"content":                  true,
		"counter_increment":        true,
		"direction":                true,
		"display":                  true,
		"empty_cells":              true,
		"float":                    true,
		"height":                   true,
		"left":                     true,
		"line_height":              true,
		"list_style_position":      true,
		"list_style_type":          true,
		"overflow":                 true,
		"position":                 true,
		"right":                    true,
		"table_layout":             true,
		"text_decoration":          true,
		"top":                      true,
		"unicode_bidi":             true,
		"vertical_align":           true,
		"visibility":               true,
		"width":                    true,
		"z_index":                  true,
		"border_bottom_color":      true,
		"border_bottom_style":      true,
		"border_collapse":          true,
		"border_left_color":        true,
		"border_left_style":        true,
		"border_right_color":       true,
		"border_right_style":       true,
		"border_top_color":         true,
		"border_top_style":         true,
		"column_width":             true,
		"column_count":             true,
		"column_rule_color":        true,
		"column_rule_style":        true,
		"column_rule_width":        true,
		"column_fill":              true,
		"column_span":              true,
		"font_feature_settings":    true,
		"font_kerning":             true,
		"font_language_override":   true,
		"font_stretch":             true,
		"font_style":               true,
		"font_variant":             true,
		"font_variant_alternates":  true,
		"font_variant_caps":        true,
		"font_variant_east_asian":  true,
		"font_variant_ligatures":   true,
		"font_variant_numeric":     true,
		"font_variant_position":    true,
		"break_after":              true,
		"break_before":             true,
		"break_inside":             true,
		"bookmark_level":           true,
		"string_set":               true,
		"image_rendering":          true,
		"page":                     true,
		"bleed_left":               true,
		"bleed_right":              true,
		"bleed_top":                true,
		"bleed_bottom":             true,
		"marks":                    true,
		"hyphenate_character":      true,
		"hyphens":                  true,
		"letter_spacing":           true,
		"text_align":               true,
		"text_transform":           true,
		"white_space":              true,
		"box_sizing":               true,
		"outline_color":            true,
		"outline_style":            true,
		"overflow_wrap":            true,
		"_weasy_specified_display": true,
	}
	dimensionsKeys = Set{
		"margin_top":           true,
		"margin_right":         true,
		"margin_bottom":        true,
		"margin_left":          true,
		"max_height":           true,
		"max_width":            true,
		"min_height":           true,
		"min_width":            true,
		"padding_top":          true,
		"padding_right":        true,
		"padding_bottom":       true,
		"padding_left":         true,
		"column_gap":           true,
		"hyphenate_limit_zone": true,
		"text_indent":          true,
	}
	intsKeys = Set{
		"border_bottom_width": true,
		"border_left_width":   true,
		"border_right_width":  true,
		"opacity":             true,
		"font_size":           true,
		"font_weight":         true,
		"orphans":             true,
		"widows":              true,
		"image_resolution":    true,
		"tab_size":            true,
		"word_spacing":        true,
	}
)

type Dimension struct {
	Unit  string
	Value int
}

type Side string

type CounterIncrement struct {
	Name  string
	Value int
}

type CounterIncrements struct {
	Auto bool
	CI   []CounterIncrement
}

type Page struct {
	Auto bool
	Zero bool
	Page int
}

func (x CounterIncrements) Copy() CounterIncrements {
	out := x
	copy(out.CI, x.CI)
	return out
}

type CounterReset struct {
	Name  string
	Value int
}

type CounterResets []CounterReset

func (x CounterResets) Copy() CounterResets {
	var out CounterResets
	copy(out, x)
	return out
}

// type StyleDict2 struct {
// 	Anonymous bool

// 	Float    string
// 	Position string
// 	Page     int

// 	Margin      map[Side]Dimension
// 	Padding     map[Side]Dimension
// 	BorderWidth map[Side]float64

// 	Direction string

// 	TextTransform, Hyphens string
// 	Display                string

// 	CounterReset []struct {
// 		Name  string
// 		Value int
// 	}
// 	CounterIncrement CounterIncrements
// }

type cascadedValue struct {
	value      string
	precedence int
}

// type StyleDict map[string]interface{}

type MiscProperties struct {
	CounterReset     CounterResets
	CounterIncrement CounterIncrements
	Page             Page
}

// Deep copy
func (s MiscProperties) Copy() MiscProperties {
	out := s
	out.CounterIncrement = s.CounterIncrement.Copy()
	out.CounterReset = s.CounterReset.Copy()
	return out
}

func (s MiscProperties) Items() map[string]interface{} {
	out := make(map[string]interface{})
	out["counter_increment"] = s.CounterIncrement
	out["counter_reset"] = s.CounterReset
	out["page"] = s.Page
	return out
}

// Set convert the value and set it.
// Should be avoided when performance matters.
func (s *MiscProperties) Set(key string, value interface{}) {
	switch key {
	case "counter_increment":
		s.CounterIncrement = value.(CounterIncrements)
	case "counter_reset":
		s.CounterReset = value.(CounterResets)
	case "page":
		s.Page = value.(Page)
	}
}

// SetFrom copy the given keys from src into s
func (s *MiscProperties) SetFrom(src MiscProperties, keys Set) {
	if keys["counter_reset"] {
		s.CounterReset = src.CounterReset
	}
	if keys["counter_increment"] {
		s.CounterIncrement = src.CounterIncrement
	}
	if keys["page"] {
		s.Page = src.Page
	}
}

type StyleDict struct {
	MiscProperties

	Anonymous  bool
	Strings    map[string]string
	Dimensions map[string]Dimension
	Ints       map[string]int
}

func NewStyleDict() StyleDict {
	var out StyleDict
	out.Strings = make(map[string]string)
	out.Dimensions = make(map[string]Dimension)
	out.Ints = make(map[string]int)
	return out
}

// IsZero returns `true` if the StyleDict is not initialized.
// Thus, we can use a zero StyleDict as null value.
func (s StyleDict) IsZero() bool {
	return s.Strings == nil
}

// Deep copy
func (s StyleDict) Copy() StyleDict {
	out := s
	out.MiscProperties = s.MiscProperties.Copy()
	out.Strings = make(map[string]string, len(s.Strings))
	out.Dimensions = make(map[string]Dimension, len(s.Dimensions))
	out.Ints = make(map[string]int, len(s.Ints))
	for k, v := range s.Strings {
		out.Strings[k] = v
	}
	for k, v := range s.Dimensions {
		out.Dimensions[k] = v
	}
	for k, v := range s.Ints {
		out.Ints[k] = v
	}
	return out
}

func (s StyleDict) Items() map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range s.Strings {
		out[k] = v
	}
	for k, v := range s.Dimensions {
		out[k] = v
	}
	for k, v := range s.Ints {
		out[k] = v
	}
	for k, v := range s.MiscProperties.Items() {
		out[k] = v
	}
	return out
}

// Set convert the value and set it.
// Should be avoided when performance matters.
func (s *StyleDict) Set(key string, value interface{}) {
	switch typed := value.(type) {
	case string:
		s.Strings[key] = typed
	case int:
		s.Ints[key] = typed
	case Dimension:
		s.Dimensions[key] = typed
	default:
		s.MiscProperties.Set(key, value)
	}
}

// SetFrom copy the given keys from src into s
func (s *StyleDict) SetFrom(src StyleDict, keys Set) {
	for key := range stringsKeys {
		if keys[key] {
			s.Strings[key] = src.Strings[key]
		}
	}
	for key := range dimensionsKeys {
		if keys[key] {
			s.Dimensions[key] = src.Dimensions[key]
		}
	}
	for key := range intsKeys {
		if keys[key] {
			s.Ints[key] = src.Ints[key]
		}
	}
	s.MiscProperties.SetFrom(src.MiscProperties, keys)
}

// Get a dict of computed style mixed from parent and cascaded styles.
func computedFromCascaded(element html.Node, cascaded map[string]cascadedValue, parentStyle StyleDict, pseudoType string,
	rootStyle StyleDict, baseUrl string) StyleDict {
	if cascaded == nil && !parentStyle.IsZero() {
		// Fast path for anonymous boxes:
		// no cascaded style, only implicitly initial or inherited values.
		computed := InitialValues.Copy()
		computed.SetFrom(parentStyle, Inherited)

		// page is not inherited but taken from the ancestor if "auto"
		computed.Page = parentStyle.Page
		// border-*-style is none, so border-width computes to zero.
		// Other than that, properties that would need computing are
		// border-*-color, but they do not apply.
		computed.Ints["border_top_width"] = 0
		computed.Ints["border_bottom_width"] = 0
		computed.Ints["border_left_width"] = 0
		computed.Ints["border_right_width"] = 0
		computed.Ints["outlineWidth"] = 0
		return computed
	}

	// Handle inheritance and initial values
	specified, computed := NewStyleDict(), NewStyleDict()
	parentItems := parentStyle.Items()
	for name, initial := range InitialValues.Items() {
		var (
			keyword string
			value   interface{}
		)
		if _, in := cascaded[name]; in {
			vp := cascaded[name]
			keyword = vp.value
			value = vp.value
		} else {
			if Inherited[name] {
				keyword = "inherit"
			} else {
				keyword = "initial"
			}
		}

		if keyword == "inherit" && parentStyle.IsZero() {
			// On the root element, "inherit" from initial values
			keyword = "initial"
		}

		if keyword == "initial" {
			value = initial
			if !InitialNotComputed[name] {
				// The value is the same as when computed
				computed.Set(name, value)
			}
		} else if keyword == "inherit" {
			value = parentItems[name]
			// Values in parentStyle are already computed.
			computed.Set(name, value)
		}
		specified.Set(name, value)
	}
	if specified.Page.Auto {
		// The page property does not inherit. However, if the page value on
		// an element is auto, then its used value is the value specified on
		// its nearest ancestor with a non-auto value. When specified on the
		// root element, the used value for auto is the empty string.
		val := Page{Zero: true}
		if !parentStyle.IsZero() {
			val = parentStyle.Page
		}
		computed.Page = val
		specified.Page = val
	}

	return StyleDict(compute(
		element, pseudoType, specified, computed, parentStyle, rootStyle,
		baseUrl))
}
