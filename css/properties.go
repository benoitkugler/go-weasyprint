package css

import (
	"math"
	"strings"
)

type Set = map[string]bool

var (
	Inherited          = Set{}
	InitialNotComputed = Set{}

	zeroPixelsValue = Value{Dimension: ZeroPixels}

	InitialValues = Properties{
		"bottom":       Length(sToV("auto")),
		"caption_side": String("top"),
		// "clear": "none",
		// "clip": TBD,  // computed value for "auto"
		"color": parseColorString("black"), // chosen by the user agent

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
		"height":           Length(sToV("auto")),
		"left":             Length(sToV("auto")),
		"right":            Length(sToV("auto")),
		"line_height":      LineHeight(sToV("normal")),
		"list_style_image": ListStyleImage{Type: "none"},
		// "list_style_position": "outside",
		"list_style_type": String("disc"),
		"margin_top":      Length(zeroPixelsValue),
		"margin_right":    Length(zeroPixelsValue),
		"margin_bottom":   Length(zeroPixelsValue),
		"margin_left":     Length(zeroPixelsValue),
		"max_height":      Length(Value{Dimension: Dimension{Value: float32(math.Inf(+1)), Unit: Px}}), // parsed value for "none})"
		"max_width":       Length(Value{Dimension: Dimension{Value: float32(math.Inf(+1)), Unit: Px}}),
		"min_height":      Length(zeroPixelsValue),
		"min_width":       Length(zeroPixelsValue),
		"overflow":        String("visible"),
		"padding_top":     Length(zeroPixelsValue),
		"padding_right":   Length(zeroPixelsValue),
		"padding_bottom":  Length(zeroPixelsValue),
		"padding_left":    Length(zeroPixelsValue),
		"quotes":          Quotes{Open: []string{"“", "‘"}, Close: []string{"”", "’"}}, // chosen by the user agent
		"position":        String("static"),
		// "table_layout": "auto",
		// "text_decoration": "none",
		// "top": "auto",
		// "unicode_bidi": "normal",
		// "vertical_align": "baseline",
		// "visibility": "visible",
		// "z_index": "auto",
		"width": Length(sToV("auto")),

		// Backgrounds and Borders 3 (CR): https://www.w3.org/TR/css3-background/
		// "background_attachment": ("scroll",),
		// "background_clip": ("border-box",),
		// "background_color": parse_color("transparent"),
		// "background_origin": ("padding-box",),
		"background_position": BackgroundPosition{
			Center{OriginX: "left", Pos: Point{X: Dimension{Unit: Percentage}}},
			Center{OriginX: "top", Pos: Point{X: Dimension{Unit: Percentage}}},
		},
		"background_image": BackgroundImage{NoneImage{}},

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
		// "border_spacing": (0, 0),
		"border_bottom_width": BorderWidth(Value{Dimension: Dimension{Value: 3}}),
		"border_left_width":   BorderWidth(Value{Dimension: Dimension{Value: 3}}),
		"border_top_width":    BorderWidth(Value{Dimension: Dimension{Value: 3}}), // computed value for "medium}"
		"border_right_width":  BorderWidth(Value{Dimension: Dimension{Value: 3}}),

		"border_bottom_left_radius":  Lengths{zeroPixelsValue, zeroPixelsValue},
		"border_bottom_right_radius": Lengths{zeroPixelsValue, zeroPixelsValue},
		"border_top_left_radius":     Lengths{zeroPixelsValue, zeroPixelsValue},
		"border_top_right_radius":    Lengths{zeroPixelsValue, zeroPixelsValue},

		// // Color 3 (REC): https://www.w3.org/TR/css3-color/
		// "opacity": 1,

		// Multi-column Layout (CR): https://www.w3.org/TR/css3-multicol/
		"column_width": ColumnWidth(sToV("auto")),
		"column_count": sToV("auto"),
		"column_gap":   ColumnGap(Value{Dimension: Dimension{Value: 1, Unit: Em}}),
		// "column_rule_color": "currentColor",
		// "column_rule_style": "none",
		"column_rule_width": BorderWidth(sToV("medium")),
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
		"font_size":   FontSize(Value{Dimension: Dimension{Value: 16}}), // actually medium, but we define medium from thi}s
		"font_weight": FontWeight(IntString{Value: 400}),

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
		"size": WidthHeight{
			{Value: initialWidthHeight[0].Value * LengthsToPixels[initialWidthHeight[0].Unit]},
			{Value: initialWidthHeight[1].Value * LengthsToPixels[initialWidthHeight[1].Unit]},
		},
		"page":         Page{String: "auto", Valid: true},
		"bleed_left":   Bleed(sToV("auto")),
		"bleed_right":  Bleed(sToV("auto")),
		"bleed_top":    Bleed(sToV("auto")),
		"bleed_bottom": Bleed(sToV("auto")),
		// "marks": "none",

		// Text 3/4 (WD/WD): https://www.w3.org/TR/css-text-4/
		// "hyphenate_character": "‐",  // computed value chosen by the user agent
		// "hyphenate_limit_chars": (5, 2, 2),
		"hyphens":              String("manual"),
		"letter_spacing":       PixelLength(sToV("normal")),
		"hyphenate_limit_zone": Length(zeroPixelsValue),
		"tab_size":             TabSize(Value{Dimension: Dimension{Value: 8}}),
		// "text_align": "-weasy-start",
		"text_indent":    Length(zeroPixelsValue),
		"text_transform": String("none"),
		"white_space":    String("normal"),
		"word_spacing":   WordSpacing{}, // computed value for "normal"

		// Transforms 1 (WD): https://www.w3.org/TR/css-transforms-1/
		"transform_origin": Lengths{Value{Dimension: Dimension{Value: 50, Unit: Percentage}}, Value{Dimension: Dimension{Value: 50, Unit: Percentage}}},
		"transform":        Transforms{}, // computed value for "none"

		// User Interface 3 (CR): https://www.w3.org/TR/css-ui-3/
		// "box_sizing": "content-box",
		// "outline_color": "currentColor",  // invert is not supported
		// "outline_style": "none",
		// "overflow_wrap": "normal",
		"outline_width": BorderWidth(Value{Dimension: Dimension{Value: 3}}), // computed value for "medium"

		// Proprietary
		"anchor": Link{}, // computed value of "none"
		"link":   Link{}, // computed value of "none"
		"lang":   Link{}, // computed value of "none"

		// Internal, to implement the "static position" for absolute boxes.
		"_weasy_specified_display": Display("inline"),
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

// prop:string-set
type StringSet struct {
	String   string
	Contents []SStrings
}

// prop:background-image
type Images []Image

// prop:background-position
type Centers []Center

// prop:background-size
type Sizes []Size

// prop:background-repeat
type Repeats [][2]string

// prop:background-clip
// prop:background-origin
// prop:font-familly
type Strings []string

// prop:content
type SContent struct {
	String   string // prop:'none' ou 'normal'
	Contents []ContentProperty
}

type NDecorations struct {
	None        bool
	Decorations Set
}

// prop:transform
type Transforms []SDimensions

// prop:transform-origin
// prop:border-spacing
// prop:size
// prop:clip
// prop:border-top-left-radius
// prop:border-top-right-radius
// prop:border-bottom-left-radius
// prop:border-bottom-right-radius
type Values []Value

// prop:font-feature-settings
type SIntStrings struct {
	String string
	Values []IntString
}

type SStrings struct {
	String  string
	Strings []string
}

type SDimensions struct {
	String     string
	Dimensions []Dimension
}

type IntStrings []IntString

type Quotes struct {
	Open, Close []string
}

type BookmarkLabel []ContentProperty

// -------------- value type ---------------------

type Float float32

type Int int

type Ints3 [3]int

// prop:page
type Page struct {
	Valid  bool
	String string
	Page   int
}

// prop:color
type Color struct {
	Type ColorType
	RGBA RGBA
}

// prop:link
// prop:anchor
// prop:lang
type NamedString struct {
	Name   string
	String string
}

type Point [2]Dimension

// prop:marks
type Marks struct {
	Crop, Cross bool
}

// prop:font-weight
type IntString struct {
	Int    int
	String string
}

// prop:break-after
// prop:break-before
// prop:display
// prop:float
type String string

// prop:top
// prop:right
// prop:left
// prop:bottom
// prop:margin-top
// prop:margin-right
// prop:margin-bottom
// prop:margin-left
// prop:height
// prop:width
// prop:min-width
// prop:min-height
// prop:max-width
// prop:max-height
// prop:padding-top
// prop:padding-right
// prop:padding-bottom
// prop:padding-left
// prop:text-indent
// prop:hyphenate-limit-zone
// prop:bleed-left
// prop:bleed-right
// prop:bleed-top
// prop:bleed-bottom
// prop:border-top-width
// prop:border-right-width
// prop:border-left-width
// prop:border-bottom-width
// prop:column-rule-width
// prop:outline-width
// prop:letter-spacing
// prop:column-width
// prop:column-gap
// prop:font-size
// prop:line-height
// prop:tab-size
// prop:vertical-align
// prop:word-spacing
type Value struct {
	Dimension
	String string
}
