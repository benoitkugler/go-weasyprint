package css

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
		// "bottom": "auto",
		// "caption_side": "top",
		// "clear": "none",
		// "clip": TBD,  // computed value for "auto"

		// "color": parse_color("black"),  // chosen by the user agent

		// "content": "normal",
		// // Means "none", but allow `display: list-item` to increment the
		// // list-item counter. If we ever have a way for authors to query
		// // computed values (JavaScript?), this value should serialize to "none".
		// "counter_increment": "auto",
		// "counter_reset": (),  // parsed value for "none"
		// // "counter_set": (),  // parsed value for "none"
		// "direction": "ltr",
		// "display": "inline",
		// "empty_cells": "show",
		// "float": "none",
		// "height": "auto",
		// "left": "auto",
		// "line_height": "normal",
		// "list_style_image": ("none", None),
		// "list_style_position": "outside",
		// "list_style_type": "disc",
		// "margin_top": Dimension(0, "px"),
		// "margin_right": Dimension(0, "px"),
		// "margin_bottom": Dimension(0, "px"),
		// "margin_left": Dimension(0, "px"),
		// "max_height": Dimension(float("inf"), "px"),  // parsed value for "none"
		// "max_width": Dimension(float("inf"), "px"),
		// "min_height": Dimension(0, "px"),
		// "min_width": Dimension(0, "px"),
		// "overflow": "visible",
		// "padding_top": Dimension(0, "px"),
		// "padding_right": Dimension(0, "px"),
		// "padding_bottom": Dimension(0, "px"),
		// "padding_left": Dimension(0, "px"),
		// "quotes": list("“”‘’"),  // chosen by the user agent
		// "position": "static",
		// "right": "auto",
		// "table_layout": "auto",
		// "text_decoration": "none",
		// "top": "auto",
		// "unicode_bidi": "normal",
		// "vertical_align": "baseline",
		// "visibility": "visible",
		// "width": "auto",
		// "z_index": "auto",

		// // Backgrounds and Borders 3 (CR): https://www.w3.org/TR/css3-background/
		// "background_attachment": ("scroll",),
		// "background_clip": ("border-box",),
		// "background_color": parse_color("transparent"),
		// "background_image": (("none", None),),
		// "background_origin": ("padding-box",),
		// "background_position": (("left", Dimension(0, "%"),
		//                          "top", Dimension(0, "%")),),
		// "background_repeat": (("repeat", "repeat"),),
		// "background_size": (("auto", "auto"),),
		// "border_bottom_color": "currentColor",
		// "border_bottom_left_radius": (Dimension(0, "px"), Dimension(0, "px")),
		// "border_bottom_right_radius": (Dimension(0, "px"), Dimension(0, "px")),
		// "border_bottom_style": "none",
		// "border_bottom_width": 3,
		// "border_collapse": "separate",
		// "border_left_color": "currentColor",
		// "border_left_style": "none",
		// "border_left_width": 3,
		// "border_right_color": "currentColor",
		// "border_right_style": "none",
		// "border_right_width": 3,
		// "border_spacing": (0, 0),
		// "border_top_color": "currentColor",
		// "border_top_left_radius": (Dimension(0, "px"), Dimension(0, "px")),
		// "border_top_right_radius": (Dimension(0, "px"), Dimension(0, "px")),
		// "border_top_style": "none",
		// "border_top_width": 3,  // computed value for "medium"

		// // Color 3 (REC): https://www.w3.org/TR/css3-color/
		// "opacity": 1,

		// // Multi-column Layout (CR): https://www.w3.org/TR/css3-multicol/
		// "column_width": "auto",
		// "column_count": "auto",
		// "column_gap": Dimension(1, "em"),
		// "column_rule_color": "currentColor",
		// "column_rule_style": "none",
		// "column_rule_width": "medium",
		// "column_fill": "balance",
		// "column_span": "none",

		// // Fonts 3 (CR): https://www.w3.org/TR/css-fonts-3/
		// "font_family": ("serif",),  // depends on user agent
		// "font_feature_settings": "normal",
		// "font_kerning": "auto",
		// "font_language_override": "normal",
		// "font_size": 16,  // actually medium, but we define medium from this
		// "font_stretch": "normal",
		// "font_style": "normal",
		// "font_variant": "normal",
		// "font_variant_alternates": "normal",
		// "font_variant_caps": "normal",
		// "font_variant_east_asian": "normal",
		// "font_variant_ligatures": "normal",
		// "font_variant_numeric": "normal",
		// "font_variant_position": "normal",
		// "font_weight": 400,

		// // Fragmentation 3 (CR): https://www.w3.org/TR/css-break-3/
		// "break_after": "auto",
		// "break_before": "auto",
		// "break_inside": "auto",
		// "orphans": 2,
		// "widows": 2,

		// // Generated Content for Paged Media (WD): https://www.w3.org/TR/css-gcpm-3/
		// "bookmark_label": (("content", "text"),),
		// "bookmark_level": "none",
		// "string_set": "none",

		// // Images 3/4 (CR/WD): https://www.w3.org/TR/css4-images/
		// "image_resolution": 1,  // dppx
		// "image_rendering": "auto",

		// // Paged Media 3 (WD): https://www.w3.org/TR/css3-page/
		// "size": None,  // set to A4 in computed_values
		// "page": "auto",
		// "bleed_left": "auto",
		// "bleed_right": "auto",
		// "bleed_top": "auto",
		// "bleed_bottom": "auto",
		// "marks": "none",

		// // Text 3/4 (WD/WD): https://www.w3.org/TR/css-text-4/
		// "hyphenate_character": "‐",  // computed value chosen by the user agent
		// "hyphenate_limit_chars": (5, 2, 2),
		// "hyphenate_limit_zone": Dimension(0, "px"),
		// "hyphens": "manual",
		// "letter_spacing": "normal",
		// "tab_size": 8,
		// "text_align": "-weasy-start",
		// "text_indent": Dimension(0, "px"),
		// "text_transform": "none",
		// "white_space": "normal",
		// "word_spacing": 0,  // computed value for "normal"

		// // Transforms 1 (WD): https://www.w3.org/TR/css-transforms-1/
		// "transform_origin": (Dimension(50, "%"), Dimension(50, "%")),
		// "transform": (),  // computed value for "none"

		// // User Interface 3 (CR): https://www.w3.org/TR/css-ui-3/
		// "box_sizing": "content-box",
		// "outline_color": "currentColor",  // invert is not supported
		// "outline_style": "none",
		// "outline_width": 3,  // computed value for "medium"
		// "overflow_wrap": "normal",

		// // Proprietary
		// "anchor": None,  // computed value of "none"
		// "link": None,  // computed value of "none"
		// "lang": None,  // computed value of "none"

		// // Internal, to implement the "static position" for absolute boxes.
		// "_weasy_specified_display": "inline",
	}
)

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
