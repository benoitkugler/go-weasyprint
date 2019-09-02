package css

import (
	"math"
	"strings"
)

type Set = map[string]bool

var (
	Inherited          = Set{}
	InitialNotComputed = Set{}

	InitialValues = map[string]CssProperty{
		"bottom":       sToV("auto"),
		"caption_side": String("top"),
		// "clear": "none",
		// "clip": TBD,  // computed value for "auto"
		// "color": parse_color("black"),  // chosen by the user agent

		"content": Content{String: "normal"},

		// Means "none", but allow `display: list-item` to increment the
		// list-item counter. If we ever have a way for authors to query
		// computed values (JavaScript?), this value should serialize to "none".
		"counter_increment": CounterIncrements{String: "auto"},
		"counter_reset":     CounterResets{}, // parsed value for "none"
		"direction":         String("ltr"),
		"display":           Display("inline"),
		// "empty_cells": "show",
		"float":            Floating("none"),
		"height":           sToV("auto"),
		"left":             sToV("auto"),
		"line_height":      sToV("normal"),
		"list_style_image": ListStyleImage{Type: "none"},
		// "list_style_position": "outside",
		"list_style_type": String("disc"),
		"margin_top":      ZeroPixels,
		"margin_right":    ZeroPixels,
		"margin_bottom":   ZeroPixels,
		"margin_left":     ZeroPixels,
		"max_height":      Value{Dimension: Dimension{Value: math.Inf(+1), Unit: Pixels}}, // parsed value for "none}"
		"max_width":       Value{Dimension: Dimension{Value: math.Inf(+1), Unit: Pixels}},
		"min_height":      ZeroPixels,
		"min_width":       ZeroPixels,
		"overflow":        String("visible"),
		"padding_top":     ZeroPixels,
		"padding_right":   ZeroPixels,
		"padding_bottom":  ZeroPixels,
		"padding_left":    ZeroPixels,
		"quotes":          Quotes{Open: []string{"“", "‘"}, Close: []string{"”", "’"}}, // chosen by the user agent
		// "position": "static",
		"right": sToV("auto"),
		// "table_layout": "auto",
		// "text_decoration": "none",
		// "top": "auto",
		// "unicode_bidi": "normal",
		// "vertical_align": "baseline",
		// "visibility": "visible",
		// "z_index": "auto",

		// Backgrounds and Borders 3 (CR): https://www.w3.org/TR/css3-background/
		// "background_attachment": ("scroll",),
		// "background_clip": ("border-box",),
		// "background_color": parse_color("transparent"),
		// "background_origin": ("padding-box",),
		"background_position": BackgroundPosition{
			Center{OriginX: "left", Pos: Point{X: Dimension{Unit: Percentage}}},
			Center{OriginX: "top", Pos: Point{X: Dimension{Unit: Percentage}}},
		},
		// "background_repeat": (("repeat", "repeat"),),
		"background_size": BackgroundSize{Size{Width: sToV("auto"), Height: sToV("auto")}},
		// "border_bottom_color": "currentColor",
		// "border_left_color": "currentColor",
		// "border_right_color": "currentColor",
		// "border_top_color": "currentColor",
		// "border_bottom_style": "none",
		// "border_left_style": "none",
		// "border_right_style": "none",
		// "border_top_style": "none",
		"border_collapse":     String("separate"),
		"border_bottom_style": String("none"),
		"border_left_style":   String("none"),
		"border_right_style":  String("none"),
		"border_top_style":    String("none"),

		"border_bottom_left_radius":  Lengths{ZeroPixels, ZeroPixels},
		"border_bottom_right_radius": Lengths{ZeroPixels, ZeroPixels},
		"border_top_left_radius":     Lengths{ZeroPixels, ZeroPixels},
		"border_top_right_radius":    Lengths{ZeroPixels, ZeroPixels},

		// // Color 3 (REC): https://www.w3.org/TR/css3-color/
		// "opacity": 1,

		// Multi-column Layout (CR): https://www.w3.org/TR/css3-multicol/
		"column_width": sToV("auto"),
		"column_count": sToV("auto"),
		"column_gap":   Value{Dimension: Dimension{Value: 1, Unit: "em"}},
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
		"break_after":  Break("auto"),
		"break_before": Break("auto"),
		"break_inside": String("auto"),
		// "orphans": 2,
		// "widows": 2,

		// // Generated Content for Paged Media (WD): https://www.w3.org/TR/css-gcpm-3/
		// "bookmark_label": (("content", "text"),),
		// "bookmark_level": "none",
		// "string_set": "none",

		// // Images 3/4 (CR/WD): https://www.w3.org/TR/css4-images/
		// "image_resolution": 1,  // dppx
		// "image_rendering": "auto",

		// Paged Media 3 (WD): https://www.w3.org/TR/css3-page/
		"size": PageSize{
			{Value: initialPageSize[0].Value * LengthsToPixels[initialPageSize[0].Unit]},
			{Value: initialPageSize[1].Value * LengthsToPixels[initialPageSize[1].Unit]},
		},
		"page":         Page{String: "auto", Valid: true},
		"bleed_left":   sToV("auto"),
		"bleed_right":  sToV("auto"),
		"bleed_top":    sToV("auto"),
		"bleed_bottom": sToV("auto"),
		// "marks": "none",

		BackgroundImage: BackgroundImage{Gradient{Type: "none"}},

		// Transforms 1 (WD): https://www.w3.org/TR/css-transforms-1/
		Transforms: Transforms{}, // computed value for "none"

		// Internal, to implement the "static position" for absolute boxes.
		weasySpecifiedDisplay: Display("inline"),

		"width":               sToV("auto"),
		"border_bottom_width": Value{Dimension: Dimension{Value: 3}},
		"border_left_width":   Value{Dimension: Dimension{Value: 3}},
		"border_top_width":    Value{Dimension: Dimension{Value: 3}}, // computed value for "medium}"
		"border_right_width":  Value{Dimension: Dimension{Value: 3}},

		"font_size":   Value{Dimension: Dimension{Value: 16}}, // actually medium, but we define medium from thi}s
		"font_weight": Value{Dimension: Dimension{Value: 400}},

		// Text 3/4 (WD/WD): https://www.w3.org/TR/css-text-4/
		"hyphenate_limit_zone": ZeroPixels,
		"tab_size":             Value{Dimension: Dimension{Value: 8}},
		"text_indent":          ZeroPixels,
		"letter_spacing":       sToV("normal"),
		"word_spacing":         Value{Dimension: Dimension{Value: 0}}, // computed value for "normal"

		// User Interface 3 (CR): https://www.w3.org/TR/css-ui-3/
		"outline_width": Value{Dimension: Dimension{Value: 3}}, // computed value for "medium"

		"position": "static",

		// // Text 3/4 (WD/WD): https://www.w3.org/TR/css-text-4/
		// "hyphenate_character": "‐",  // computed value chosen by the user agent
		// "hyphenate_limit_chars": (5, 2, 2),
		"hyphens": "manual",
		// "text_align": "-weasy-start",
		"text_transform": "none",
		"white_space":    "normal",

		// Paged Media 3 (WD): https://www.w3.org/TR/css3-page/
		"size": nil, // set to A4 in computed_values

		// Transforms 1 (WD): https://www.w3.org/TR/css-transforms-1/
		"transform_origin": Lengths{Value{Dimension: Dimension{Value: 50, Unit: Percentage}}, Value{Dimension: Dimension{Value: 50, Unit: Percentage}}},

		// Proprietary
		"anchor": Link{}, // computed value of "none"
		"link":   Link{}, // computed value of "none"
		"lang":   Link{}, // computed value of "none"

		// // Backgrounds and Borders 3 (CR): https://www.w3.org/TR/css3-background/

		// "border_spacing": (0, 0),

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

	knownProperties = Set{}

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
	for name := range InitialValues {
		knownProperties[strings.ReplaceAll(name, "_", "-")] = true
	}
}

// string-set
type StringSet struct {
	String   string
	Contents []StringContent
}

func (s StringSet) IsNone() bool {
	return s.String == "" && s.Contents == nil
}

func (s StringSet) Copy() StringSet {
	out := s
	out.Contents = make([]StringContent, len(s.Contents))
	for index, l := range s.Contents {
		out.Contents[index] = l.Copy()
	}
	return out
}

// background-image
type BackgroundImage []Image

// background-position
type BackgroundPosition []Center

// background-size
type BackgroundSize []Size

// background-repeat
type BackgroundRepeat [][2]string

// content
type Content struct {
	String string // 'none' ou 'normal'
	List   []ContentProperty
}

func (c Content) IsNone() bool {
	return c.String == "" && c.List == nil
}

// deep copy
func (c Content) Copy() Content {
	out := c
	out.List = append([]ContentProperty{}, c.List...)
	return out
}

type TextDecoration struct {
	None        bool
	Decorations Set
}

// width, height
type PageSize [2]Dimension

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

// marks
type Marks struct {
	Crop, Cross bool
}

func (m Marks) IsNone() bool {
	return m == Marks{}
}

// break_after
// break_before
type Break string

// display
type Display string

// float
type Floating string

// Standard string property
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
	Type string
	Attr string
}

// anchor
type Anchor Link

// lang
type Lang Link
