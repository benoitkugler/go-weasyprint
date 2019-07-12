package css

import "math"

type Set = map[string]bool

var (
	ConvertersValue = map[string]func(Value) CssProperty{
		"top":                  valueToLength,
		"right":                valueToLength,
		"left":                 valueToLength,
		"bottom":               valueToLength,
		"margin_top":           valueToLength,
		"margin_right":         valueToLength,
		"margin_bottom":        valueToLength,
		"margin_left":          valueToLength,
		"height":               valueToLength,
		"width":                valueToLength,
		"min_width":            valueToLength,
		"min_height":           valueToLength,
		"max_width":            valueToLength,
		"max_height":           valueToLength,
		"padding_top":          valueToLength,
		"padding_right":        valueToLength,
		"padding_bottom":       valueToLength,
		"padding_left":         valueToLength,
		"text_indent":          valueToLength,
		"hyphenate_limit_zone": valueToLength,

		"bleed_left":   valueToBleed,
		"bleed_right":  valueToBleed,
		"bleed_top":    valueToBleed,
		"bleed_bottom": valueToBleed,

		"border_top_width":    valueToBorderWidth,
		"border_right_width":  valueToBorderWidth,
		"border_left_width":   valueToBorderWidth,
		"border_bottom_width": valueToBorderWidth,
		"column_rule_width":   valueToBorderWidth,
		"outline_width":       valueToBorderWidth,

		"column_width":   valueToColumnWidth,
		"column_gap":     valueToColumnGap,
		"font_size":      valueToFontSize,
		"font_weight":    valueToFontWeight,
		"line_height":    valueToLineHeight,
		"tab_size":       valueToTabSize,
		"vertical_align": valueToVerticalAlign,
		"word_spacing":   valueToWordSpacing,
		"letter_spacing": valueToPixelLength,
	}
	ConvertersString = map[string]func(string) CssProperty{
		"break_after":  func(s string) CssProperty { return Break(s) },
		"break_before": func(s string) CssProperty { return Break(s) },
		"display":      func(s string) CssProperty { return Display(s) },
		"float":        func(s string) CssProperty { return Float(s) },
	}
	ConvertersLink = map[string]func(Link) CssProperty{
		"link":   func(l Link) CssProperty { return l },
		"anchor": func(l Link) CssProperty { return Anchor(l) },
		"lang":   func(l Link) CssProperty { return Lang(l) },
	}
	Inherited          Set
	InitialNotComputed Set

	InitialValues = StyleDict{
		MiscProperties: MiscProperties{
			Content: Content{String: "normal"},

			// Means "none", but allow `display: list-item` to increment the
			// list-item counter. If we ever have a way for authors to query
			// computed values (JavaScript?), this value should serialize to "none".
			CounterIncrements: CounterIncrements{String: "auto"},
			CounterResets:     CounterResets{}, // parsed value for "none"

			BackgroundPosition: BackgroundPosition{
				Center{OriginX: "left", PosX: Dimension{Unit: "%"}},
				Center{OriginX: "top", PosX: Dimension{Unit: "%"}},
			},

			BackgroundImage: BackgroundImage{Gradient{Type: "none"}},
			BackgroundSize:  BackgroundSize{Size{Width: vs("auto"), Height: vs("auto")}},

			// Paged Media 3 (WD): https://www.w3.org/TR/css3-page/
			Page: Page{String: "auto", Valid: true},

			// Transforms 1 (WD): https://www.w3.org/TR/css-transforms-1/
			Transforms: Transforms{}, // computed value for "none"

			ListStyleImage: ListStyleImage{Type: "none"},

			// Internal, to implement the "static position" for absolute boxes.
			weasySpecifiedDisplay: Display("inline"),
		},
		Values: map[string]Value{
			"bottom":              vs("auto"),
			"height":              vs("auto"),
			"left":                vs("auto"),
			"line_height":         vs("normal"),
			"margin_top":          ZeroPixels,
			"margin_right":        ZeroPixels,
			"margin_bottom":       ZeroPixels,
			"margin_left":         ZeroPixels,
			"max_height":          Value{Dimension: Dimension{Value: int(math.Inf(+1)), Unit: "px"}}, // parsed value for "none}"
			"max_width":           Value{Dimension: Dimension{Value: int(math.Inf(+1)), Unit: "px"}},
			"min_height":          ZeroPixels,
			"min_width":           ZeroPixels,
			"padding_top":         ZeroPixels,
			"padding_right":       ZeroPixels,
			"padding_bottom":      ZeroPixels,
			"padding_left":        ZeroPixels,
			"right":               vs("auto"),
			"width":               vs("auto"),
			"border_bottom_width": Value{Dimension: Dimension{Value: 3}},
			"border_left_width":   Value{Dimension: Dimension{Value: 3}},
			"border_top_width":    Value{Dimension: Dimension{Value: 3}}, // computed value for "medium}"
			"border_right_width":  Value{Dimension: Dimension{Value: 3}},

			// Multi-column Layout (CR): https://www.w3.org/TR/css3-multicol/
			"column_width": vs("auto"),
			"column_count": vs("auto"),
			"column_gap":   Value{Dimension: Dimension{Value: 1, Unit: "em"}},

			"font_size":   Value{Dimension: Dimension{Value: 16}}, // actually medium, but we define medium from thi}s
			"font_weight": Value{Dimension: Dimension{Value: 400}},

			// Paged Media 3 (WD): https://www.w3.org/TR/css3-page/
			"bleed_left":   vs("auto"),
			"bleed_right":  vs("auto"),
			"bleed_top":    vs("auto"),
			"bleed_bottom": vs("auto"),

			// Text 3/4 (WD/WD): https://www.w3.org/TR/css-text-4/
			"hyphenate_limit_zone": ZeroPixels,
			"tab_size":             Value{Dimension: Dimension{Value: 8}},
			"text_indent":          ZeroPixels,
			"letter_spacing":       vs("normal"),
			"word_spacing":         Value{Dimension: Dimension{Value: 0}}, // computed value for "normal"

			// User Interface 3 (CR): https://www.w3.org/TR/css-ui-3/
			"outline_width": Value{Dimension: Dimension{Value: 3}}, // computed value for "medium"
		},
		Strings: map[string]string{
			"display":  "inline",
			"float":    "none",
			"position": "static",

			// Fragmentation 3 (CR): https://www.w3.org/TR/css-break-3/
			"break_after":  "auto",
			"break_before": "auto",
			"break_inside": "auto",

			"direction": "ltr",

			// // Text 3/4 (WD/WD): https://www.w3.org/TR/css-text-4/
			// "hyphenate_character": "‐",  // computed value chosen by the user agent
			// "hyphenate_limit_chars": (5, 2, 2),
			"hyphens": "manual",
			// "text_align": "-weasy-start",
			"text_transform": "none",
			"white_space":    "normal",

			"caption_side": "top",

			"list_style_type": "disc",

			// Backgrounds and Borders 3 (CR): https://www.w3.org/TR/css3-background/
			"border_collapse": "separate",
		},
		Lengthss: map[string]Lengths{
			"border_bottom_left_radius":  Lengths{ZeroPixels, ZeroPixels},
			"border_bottom_right_radius": Lengths{ZeroPixels, ZeroPixels},
			"border_top_left_radius":     Lengths{ZeroPixels, ZeroPixels},
			"border_top_right_radius":    Lengths{ZeroPixels, ZeroPixels},

			// Paged Media 3 (WD): https://www.w3.org/TR/css3-page/
			"size": nil, // set to A4 in computed_values

			// Transforms 1 (WD): https://www.w3.org/TR/css-transforms-1/
			"transform_origin": Lengths{Value{Dimension: Dimension{Value: 50, Unit: "%"}}, Value{Dimension: Dimension{Value: 50, Unit: "%"}}},
		},
		Links: map[string]Link{

			// Proprietary
			"anchor": Link{}, // computed value of "none"
			"link":   Link{}, // computed value of "none"
			"lang":   Link{}, // computed value of "none"
		},

		// "column_rule_color": "currentColor",
		// "column_rule_style": "none",
		// "column_rule_width": "medium",
		// "column_fill": "balance",
		// "column_span": "none",

		// "clear": "none",
		// "clip": TBD,  // computed value for "auto"

		// "color": parse_color("black"),  // chosen by the user agent

		// "empty_cells": "show",

		// "list_style_position": "outside",

		// "overflow": "visible",
		// "quotes": list("“”‘’"),  // chosen by the user agent
		// "position": "static",
		// "table_layout": "auto",
		// "text_decoration": "none",
		// "top": "auto",
		// "unicode_bidi": "normal",
		// "vertical_align": "baseline",
		// "visibility": "visible",
		// "z_index": "auto",

		// // Backgrounds and Borders 3 (CR): https://www.w3.org/TR/css3-background/
		// "background_attachment": ("scroll",),
		// "background_clip": ("border-box",),
		// "background_color": parse_color("transparent"),
		// "background_origin": ("padding-box",),
		// "background_repeat": (("repeat", "repeat"),),
		// "border_bottom_color": "currentColor",
		// "border_bottom_style": "none",

		// "border_left_color": "currentColor",
		// "border_left_style": "none",
		// "border_right_color": "currentColor",
		// "border_right_style": "none",
		// "border_spacing": (0, 0),
		// "border_top_color": "currentColor",
		// "border_top_style": "none",

		// // Color 3 (REC): https://www.w3.org/TR/css3-color/
		// "opacity": 1,

		// // Fonts 3 (CR): https://www.w3.org/TR/css-fonts-3/
		// "font_family": ("serif",),  // depends on user agent
		// "font_feature_settings": "normal",
		// "font_kerning": "auto",
		// "font_language_override": "normal",
		// "font_stretch": "normal",
		// "font_style": "normal",
		// "font_variant": "normal",
		// "font_variant_alternates": "normal",
		// "font_variant_caps": "normal",
		// "font_variant_east_asian": "normal",
		// "font_variant_ligatures": "normal",
		// "font_variant_numeric": "normal",
		// "font_variant_position": "normal",

		// // Fragmentation 3 (CR): https://www.w3.org/TR/css-break-3/
		// "orphans": 2,
		// "widows": 2,

		// // Generated Content for Paged Media (WD): https://www.w3.org/TR/css-gcpm-3/
		// "bookmark_label": (("content", "text"),),
		// "bookmark_level": "none",
		// "string_set": "none",

		// // Images 3/4 (CR/WD): https://www.w3.org/TR/css4-images/
		// "image_resolution": 1,  // dppx
		// "image_rendering": "auto",
		// "marks": "none",

		// // Text 3/4 (WD/WD): https://www.w3.org/TR/css-text-4/
		// "hyphenate_character": "‐",  // computed value chosen by the user agent
		// "hyphenate_limit_chars": (5, 2, 2),
		// "hyphens": "manual",
		// "text_align": "-weasy-start",
		// "text_transform": "none",
		// "white_space": "normal",

		// // User Interface 3 (CR): https://www.w3.org/TR/css-ui-3/
		// "box_sizing": "content-box",
		// "outline_color": "currentColor",  // invert is not supported
		// "outline_style": "none",
		// "overflow_wrap": "normal",
	}

	InitialValuesItems map[string]CssProperty

	// http://www.w3.org/TR/CSS21/tables.html#model
	// See also http://lists.w3.org/Archives/Public/www-style/2012Jun/0066.html
	// Only non-inherited properties need to be included here.
	TableWrapperBoxProperties = Set{
		"bottom":            true,
		"break_after":       true,
		"break_before":      true,
		"break_inside":      true,
		"clear":             true,
		"counter_increment": true,
		"counter_reset":     true,
		"float":             true,
		"left":              true,
		"margin_top":        true,
		"margin_bottom":     true,
		"margin_left":       true,
		"margin_right":      true,
		"opacity":           true,
		"overflow":          true,
		"position":          true,
		"right":             true,
		"top":               true,
		"transform":         true,
		"transform_origin":  true,
		"vertical_align":    true,
		"z_index":           true,
	}
)

func init() {
	InitialValuesItems = InitialValues.Items()
}

func vs(s string) Value { return Value{String: s} }

// Dimension or string
type Value struct {
	Dimension
	String string
}

// background-image
type BackgroundImage []Gradient

// background-position
type BackgroundPosition []Center

// background-size
type BackgroundSize []Size

// content
type Content struct {
	List   [][2]string
	String string
}

func (c Content) IsNil() bool {
	return c.String == "" && c.List == nil
}

// deep copy
func (c Content) Copy() Content {
	out := c
	out.List = append([][2]string{}, c.List...)
	return out
}

// transform
type Transforms []Transform

// transform-origin
// border-spacing
// size
// clip
// border-top-left-radius
// border-top-right-radius
// border-bottom-left-radius
// border-bottom-right-radius
type Lengths []Value

// break_after
// break_before
type Break string

// display
type Display string

// float
type Float string

type String string

// top
// right
// left
// bottom
// margin_top
// margin_right
// margin_bottom
// margin_left
// height
// width
// min_width
// min_height
// max_width
// max_height
// padding_top
// padding_right
// padding_bottom
// padding_left
// text_indent
// hyphenate_limit_zone
type Length Value

// bleed_left
// bleed_right
// bleed_top
// bleed_bottom
type Bleed Value

// border_top_width
// border_right_width
// border_left_width
// border_bottom_width
// column_rule_width
// outline_width
type BorderWidth Value

// letter_spacing
type PixelLength Value

// column_width
type ColumnWidth Value

// column_gap
type ColumnGap Value

// font_size
type FontSize Value

// font_weight
type FontWeight Value

// line_height
type LineHeight Value

// tab_size
type TabSize Value

// vertical_align
type VerticalAlign Value

// word_spacing
type WordSpacing Value

// link
type Link struct {
	String string
	Type   string
	Attr   string
}

// anchor
type Anchor Link

// lang
type Lang Link

func valueToLength(v Value) CssProperty        { return Length(v) }
func valueToBleed(v Value) CssProperty         { return Bleed(v) }
func valueToPixelLength(v Value) CssProperty   { return PixelLength(v) }
func valueToBorderWidth(v Value) CssProperty   { return BorderWidth(v) }
func valueToColumnWidth(v Value) CssProperty   { return ColumnWidth(v) }
func valueToColumnGap(v Value) CssProperty     { return ColumnGap(v) }
func valueToFontSize(v Value) CssProperty      { return FontSize(v) }
func valueToFontWeight(v Value) CssProperty    { return FontWeight(v) }
func valueToLineHeight(v Value) CssProperty    { return LineHeight(v) }
func valueToTabSize(v Value) CssProperty       { return TabSize(v) }
func valueToVerticalAlign(v Value) CssProperty { return VerticalAlign(v) }
func valueToWordSpacing(v Value) CssProperty   { return WordSpacing(v) }

func (v Value) SetOn(name string, s *StyleDict) {
	s.Values[name] = v
}

func (v Length) SetOn(name string, s *StyleDict) {
	Value(v).SetOn(name, s)
}
func (v Bleed) SetOn(name string, s *StyleDict) {
	Value(v).SetOn(name, s)
}
func (v PixelLength) SetOn(name string, s *StyleDict) {
	Value(v).SetOn(name, s)
}
func (v BorderWidth) SetOn(name string, s *StyleDict) {
	Value(v).SetOn(name, s)
}
func (v ColumnWidth) SetOn(name string, s *StyleDict) {
	Value(v).SetOn(name, s)
}
func (v ColumnGap) SetOn(name string, s *StyleDict) {
	Value(v).SetOn(name, s)
}
func (v FontSize) SetOn(name string, s *StyleDict) {
	Value(v).SetOn(name, s)
}
func (v FontWeight) SetOn(name string, s *StyleDict) {
	Value(v).SetOn(name, s)
}
func (v LineHeight) SetOn(name string, s *StyleDict) {
	Value(v).SetOn(name, s)
}
func (v TabSize) SetOn(name string, s *StyleDict) {
	Value(v).SetOn(name, s)
}
func (v VerticalAlign) SetOn(name string, s *StyleDict) {
	Value(v).SetOn(name, s)
}
func (v WordSpacing) SetOn(name string, s *StyleDict) {
	Value(v).SetOn(name, s)
}

func (v Break) SetOn(name string, s *StyleDict) {
	s.Strings[name] = string(v)
}
func (v Display) SetOn(name string, s *StyleDict) {
	s.Strings[name] = string(v)
}
func (v Float) SetOn(name string, s *StyleDict) {
	s.Strings[name] = string(v)
}
func (v String) SetOn(name string, s *StyleDict) {
	s.Strings[name] = string(v)
}

func (v Link) SetOn(name string, s *StyleDict) {
	s.Links[name] = v
}
func (v Anchor) SetOn(name string, s *StyleDict) {
	s.Links[name] = Link(v)
}
func (v Lang) SetOn(name string, s *StyleDict) {
	s.Links[name] = Link(v)
}

func (v Lengths) SetOn(name string, s *StyleDict) {
	s.Lengthss[name] = v
}

func (v CounterResets) SetOn(name string, s *StyleDict) {
	s.CounterResets = v
}
func (v CounterIncrements) SetOn(name string, s *StyleDict) {
	s.CounterIncrements = v
}
func (v Page) SetOn(name string, s *StyleDict) {
	s.Page = v
}

func (v BackgroundImage) SetOn(name string, s *StyleDict) {
	s.BackgroundImage = v
}
func (v BackgroundPosition) SetOn(name string, s *StyleDict) {
	s.BackgroundPosition = v
}
func (v BackgroundSize) SetOn(name string, s *StyleDict) {
	s.BackgroundSize = v
}
func (v Content) SetOn(name string, s *StyleDict) {
	s.Content = v
}
func (v Transforms) SetOn(name string, s *StyleDict) {
	s.Transforms = v
}
