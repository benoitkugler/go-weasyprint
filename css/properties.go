package css

import (
	"math"
	"strings"
)

var (
	Inherited          = Set{}
	InitialNotComputed = Set{}

	zeroPixelsValue = Value{Dimension: ZeroPixels}

	// How many CSS pixels is one <unit>?
	// http://www.w3.org/TR/CSS21/syndata.html#length-units
	LengthsToPixels = map[Unit]float32{
		Px: 1,
		Pt: 1. / 0.75,
		Pc: 16.,             // LengthsToPixels["pt"] * 12
		In: 96.,             // LengthsToPixels["pt"] * 72
		Cm: 96. / 2.54,      // LengthsToPixels["in"] / 2.54
		Mm: 96. / 25.4,      // LengthsToPixels["in"] / 25.4
		Q:  96. / 25.4 / 4., // LengthsToPixels["mm"] / 4
	}

	// Value in pixels of font-size for <absolute-size> keywords: 12pt (16px) for
	// medium, and scaling factors given in CSS3 for others:
	// http://www.w3.org/TR/css3-fonts/#font-size-prop
	// TODO: this will need to be ordered to implement 'smaller' and 'larger'
	FontSizeKeywords = map[string]float32{ // medium is 16px, others are a ratio of medium
		"xx-small": InitialValues.GetFontSize().Value * 3 / 5,
		"x-small":  InitialValues.GetFontSize().Value * 3 / 4,
		"small":    InitialValues.GetFontSize().Value * 8 / 9,
		"medium":   InitialValues.GetFontSize().Value * 1 / 1,
		"large":    InitialValues.GetFontSize().Value * 6 / 5,
		"x-large":  InitialValues.GetFontSize().Value * 3 / 2,
		"xx-large": InitialValues.GetFontSize().Value * 2 / 1,
	}

	// http://www.w3.org/TR/css3-page/#size
	// name=(width in pixels, height in pixels)
	PageSizes = map[string]Point{
		"a5":     {Dimension{Value: 148, Unit: Mm}, Dimension{Value: 210, Unit: Mm}},
		"a4":     InitialWidthHeight,
		"a3":     {Dimension{Value: 297, Unit: Mm}, Dimension{Value: 420, Unit: Mm}},
		"b5":     {Dimension{Value: 176, Unit: Mm}, Dimension{Value: 250, Unit: Mm}},
		"b4":     {Dimension{Value: 250, Unit: Mm}, Dimension{Value: 353, Unit: Mm}},
		"letter": {Dimension{Value: 8.5, Unit: In}, Dimension{Value: 11, Unit: In}},
		"legal":  {Dimension{Value: 8.5, Unit: In}, Dimension{Value: 14, Unit: In}},
		"ledger": {Dimension{Value: 11, Unit: In}, Dimension{Value: 17, Unit: In}},
	}

	InitialWidthHeight = Point{Dimension{Value: 210, Unit: Mm}, Dimension{Value: 297, Unit: Mm}}

	InitialValues = Properties{
		"bottom":       SToV("auto"),
		"caption_side": String("top"),
		// "clear": "none",
		// "clip": TBD,  // computed value for "auto"
		"color": parseColorString("black"), // chosen by the user agent

		"content": SContent{String: "normal"},

		// Means "none", but allow `display: list-item` to increment the
		// list-item counter. If we ever have a way for authors to query
		// computed values (JavaScript?), this value should serialize to "none".
		"counter_increment": SIntStrings{String: "auto"},
		"counter_reset":     IntStrings{}, // parsed value for "none"
		"direction":         String("ltr"),
		"display":           String("inline"),
		// "empty_cells": "show",
		"float":            String("none"),
		"height":           SToV("auto"),
		"left":             SToV("auto"),
		"right":            SToV("auto"),
		"line_height":      SToV("normal"),
		"list_style_image": NoneImage{},
		// "list_style_position": "outside",
		"list_style_type": String("disc"),
		"margin_top":      zeroPixelsValue,
		"margin_right":    zeroPixelsValue,
		"margin_bottom":   zeroPixelsValue,
		"margin_left":     zeroPixelsValue,
		"max_height":      Value{Dimension: Dimension{Value: float32(math.Inf(+1)), Unit: Px}}, // parsed value for "none}"
		"max_width":       Value{Dimension: Dimension{Value: float32(math.Inf(+1)), Unit: Px}},
		"min_height":      zeroPixelsValue,
		"min_width":       zeroPixelsValue,
		"overflow":        String("visible"),
		"padding_top":     zeroPixelsValue,
		"padding_right":   zeroPixelsValue,
		"padding_bottom":  zeroPixelsValue,
		"padding_left":    zeroPixelsValue,
		"quotes":          Quotes{Open: []string{"“", "‘"}, Close: []string{"”", "’"}}, // chosen by the user agent
		"position":        String("static"),
		// "table_layout": "auto",
		// "text_decoration": "none",
		// "top": "auto",
		// "unicode_bidi": "normal",
		// "vertical_align": "baseline",
		// "visibility": "visible",
		// "z_index": "auto",
		"width": SToV("auto"),

		// Backgrounds and Borders 3 (CR): https://www.w3.org/TR/css3-background/
		// "background_attachment": ("scroll",),
		// "background_clip": ("border-box",),
		"background_color": parseColorString("transparent"),
		// "background_origin": ("padding-box",),
		"background_position": Centers{
			Center{OriginX: "left", OriginY: "top", Pos: Point{Dimension{Unit: Percentage}, Dimension{Unit: Percentage}}},
		},
		"background_image": Images{NoneImage{}},

		// "background_repeat": (("repeat", "repeat"),),
		"background_size": Sizes{Size{Width: SToV("auto"), Height: SToV("auto")}},
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
		"border_bottom_width": Value{Dimension: Dimension{Value: 3}},
		"border_left_width":   Value{Dimension: Dimension{Value: 3}},
		"border_top_width":    Value{Dimension: Dimension{Value: 3}}, // computed value for "medium"
		"border_right_width":  Value{Dimension: Dimension{Value: 3}},

		"border_bottom_left_radius":  Values{zeroPixelsValue, zeroPixelsValue},
		"border_bottom_right_radius": Values{zeroPixelsValue, zeroPixelsValue},
		"border_top_left_radius":     Values{zeroPixelsValue, zeroPixelsValue},
		"border_top_right_radius":    Values{zeroPixelsValue, zeroPixelsValue},

		// // Color 3 (REC): https://www.w3.org/TR/css3-color/
		// "opacity": 1,

		// Multi-column Layout (CR): https://www.w3.org/TR/css3-multicol/
		"column_width": SToV("auto"),
		"column_count": SToV("auto"),
		"column_gap":   Value{Dimension: Dimension{Value: 1, Unit: Em}},
		// "column_rule_color": "currentColor",
		// "column_rule_style": "none",
		"column_rule_width": SToV("medium"),
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
		"font_size":   Value{Dimension: Dimension{Value: 16}}, // actually medium, but we define medium from this
		"font_weight": IntString{Int: 400},

		// // Fragmentation 3 (CR): https://www.w3.org/TR/css-break-3/
		"break_after":  String("auto"),
		"break_before": String("auto"),
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
		"size": Point{
			{Value: InitialWidthHeight[0].Value * LengthsToPixels[InitialWidthHeight[0].Unit]},
			{Value: InitialWidthHeight[1].Value * LengthsToPixels[InitialWidthHeight[1].Unit]},
		},
		"page":         Page{String: "auto", Valid: true},
		"bleed_left":   SToV("auto"),
		"bleed_right":  SToV("auto"),
		"bleed_top":    SToV("auto"),
		"bleed_bottom": SToV("auto"),
		// "marks": "none",

		// Text 3/4 (WD/WD): https://www.w3.org/TR/css-text-4/
		// "hyphenate_character": "‐",  // computed value chosen by the user agent
		// "hyphenate_limit_chars": (5, 2, 2),
		"hyphens":              String("manual"),
		"letter_spacing":       SToV("normal"),
		"hyphenate_limit_zone": zeroPixelsValue,
		"tab_size":             Value{Dimension: Dimension{Value: 8}},
		// "text_align": "-weasy-start",
		"text_indent":    zeroPixelsValue,
		"text_transform": String("none"),
		"white_space":    String("normal"),
		"word_spacing":   Value{}, // computed value for "normal"

		// Transforms 1 (WD): https://www.w3.org/TR/css-transforms-1/
		"transform_origin": Point{{Value: 50, Unit: Percentage}, {Value: 50, Unit: Percentage}},
		"transform":        Transforms{}, // computed value for "none"

		// User Interface 3 (CR): https://www.w3.org/TR/css-ui-3/
		// "box_sizing": "content-box",
		// "outline_color": "currentColor",  // invert is not supported
		// "outline_style": "none",
		// "overflow_wrap": "normal",
		"outline_width": Value{Dimension: Dimension{Value: 3}}, // computed value for "medium"

		// Proprietary
		"anchor": NamedString{}, // computed value of "none"
		"link":   NamedString{}, // computed value of "none"
		"lang":   NamedString{}, // computed value of "none"

		// Internal, to implement the "static position" for absolute boxes.
		"_weasy_specified_display": String("inline"),
	}

	KnownProperties = Set{}

	// Not applicable to the print media
	NotPrintMedia = Set{
		// Aural media:
		"azimuth":           Has,
		"cue":               Has,
		"cue-after":         Has,
		"cue-before":        Has,
		"cursor":            Has,
		"elevation":         Has,
		"pause":             Has,
		"pause-after":       Has,
		"pause-before":      Has,
		"pitch-range":       Has,
		"pitch":             Has,
		"play-during":       Has,
		"richness":          Has,
		"speak-header":      Has,
		"speak-numeral":     Has,
		"speak-punctuation": Has,
		"speak":             Has,
		"speech-rate":       Has,
		"stress":            Has,
		"voice-family":      Has,
		"volume":            Has,

		// outlines are not just for interactive but any visual media in css3-ui
	}

	// http://www.w3.org/TR/CSS21/tables.html#model
	// See also http://lists.w3.org/Archives/Public/www-style/2012Jun/0066.html
	// Only non-inherited properties need to be included here.
	TableWrapperBoxProperties = Set{
		"bottom":            Has,
		"break_after":       Has,
		"break_before":      Has,
		"break_inside":      Has,
		"clear":             Has,
		"counter_increment": Has,
		"counter_reset":     Has,
		"float":             Has,
		"left":              Has,
		"margin_top":        Has,
		"margin_bottom":     Has,
		"margin_left":       Has,
		"margin_right":      Has,
		"opacity":           Has,
		"overflow":          Has,
		"position":          Has,
		"right":             Has,
		"top":               Has,
		"transform":         Has,
		"transform_origin":  Has,
		"vertical_align":    Has,
		"z_index":           Has,
	}
)

func init() {
	for name := range InitialValues {
		KnownProperties[strings.ReplaceAll(name, "_", "-")] = Has
	}
}

type Properties map[string]CssProperty

// prop:string-set
type StringSet struct {
	String   string
	Contents []SContent
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
// prop:background-attachment
// prop:font-familly
type Strings []string

// prop:content
type SContent struct {
	String   string // prop:'none' ou 'normal'
	Contents []ContentProperty
}

// prop:text-decoration
type NDecorations struct {
	None        bool
	Decorations Set
}

// prop:transform
type Transforms []SDimensions

// prop:border-spacing
// prop:size
// prop:clip
// prop:border-top-left-radius
// prop:border-top-right-radius
// prop:border-bottom-left-radius
// prop:border-bottom-right-radius
type Values []Value

// prop:font-feature-settings
// prop:counter-increment
type SIntStrings struct {
	String string
	Values []IntString
}

// prop:font-variant-numeric
// prop:font-variant-ligatures
// prop:font-variant-east-asian
type SStrings struct {
	String  string
	Strings []string
}

type SDimensions struct {
	String     string
	Dimensions []Dimension
}

// prop:counter-reset
type IntStrings []IntString

type Quotes struct {
	Open, Close []string
}

// prop:bookmark-label
type BookmarkLabel []ContentProperty

// -------------- value type ---------------------

// prop:opacity
type Float float32

type Int int

// prop:hyphenate-limit-chars
type Ints3 [3]int

// prop:page
type Page struct {
	Valid  bool
	String string
	Page   int
}

// Dimension or "auto" or "cover" or "contain"
type Size struct {
	Width, Height Value
	String        string
}
type Center struct {
	OriginX, OriginY string
	Pos              Point
}

// prop:color
// prop:background-color
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

// prop:transform-origin
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
// prop:break-inside
// prop:display
// prop:float
// prop:-weasy-specified-display
// prop:position
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
// prop:column-width
// prop:column-gap
// prop:font-size
// prop:line-height
// prop:tab-size
// prop:vertical-align
// prop:letter-spacing
// prop:word-spacing
type Value struct {
	Dimension
	String string
}
