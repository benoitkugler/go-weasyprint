package css

import (
	"math"
	"strings"

	"github.com/benoitkugler/go-weasyprint/css/parser"
)

var (
	zeroPixelsValue = Value{Dimension: ZeroPixels}
	CurrentColor    = Color{Type: parser.ColorCurrentColor}
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
		"clear":        String("none"),
		"clip":         Values{},                           // computed value for "auto"
		"color":        Color(parser.ParseColor2("black")), // chosen by the user agent

		// Means "none", but allow `display: list-item` to increment the
		// list-item counter. If we ever have a way for authors to query
		// computed values (JavaScript?), this value should serialize to "none".
		"counter_increment":   SIntStrings{String: "auto"},
		"counter_reset":       IntStrings{}, // parsed value for "none"
		"direction":           String("ltr"),
		"display":             String("inline"),
		"empty_cells":         String("show"),
		"float":               String("none"),
		"height":              SToV("auto"),
		"left":                SToV("auto"),
		"right":               SToV("auto"),
		"line_height":         SToV("normal"),
		"list_style_image":    NoneImage{},
		"list_style_position": String("outside"),
		"list_style_type":     String("disc"),
		"margin_top":          zeroPixelsValue,
		"margin_right":        zeroPixelsValue,
		"margin_bottom":       zeroPixelsValue,
		"margin_left":         zeroPixelsValue,
		"max_height":          Value{Dimension: Dimension{Value: float32(math.Inf(+1)), Unit: Px}}, // parsed value for "none}"
		"max_width":           Value{Dimension: Dimension{Value: float32(math.Inf(+1)), Unit: Px}},
		"padding_top":         zeroPixelsValue,
		"padding_right":       zeroPixelsValue,
		"padding_bottom":      zeroPixelsValue,
		"padding_left":        zeroPixelsValue,
		"position":            String("static"),
		"table_layout":        String("auto"),
		"top":                 SToV("auto"),
		// "unicode_bidi":        String("normal"),
		"vertical_align": SToV("baseline"),
		"visibility":     String("visible"),
		"width":          SToV("auto"),
		"z_index":        IntString{String: "auto"},

		// Backgrounds and Borders 3 (CR): https://www.w3.org/TR/css3-background/
		"background_attachment": Strings{"scroll"},
		"background_clip":       Strings{"border-box"},
		"background_color":      Color(parser.ParseColor2("transparent")),
		"background_image":      Images{NoneImage{}},
		"background_origin":     Strings{"padding-box"},
		"background_position": Centers{
			Center{OriginX: "left", OriginY: "top", Pos: Point{Dimension{Unit: Percentage}, Dimension{Unit: Percentage}}},
		},
		"background_repeat":   Repeats{{"repeat", "repeat"}},
		"background_size":     Sizes{Size{Width: SToV("auto"), Height: SToV("auto")}},
		"border_bottom_color": CurrentColor,
		"border_left_color":   CurrentColor,
		"border_right_color":  CurrentColor,
		"border_top_color":    CurrentColor,
		"border_bottom_style": String("none"),
		"border_left_style":   String("none"),
		"border_right_style":  String("none"),
		"border_top_style":    String("none"),
		"border_collapse":     String("separate"),
		"border_spacing":      Point{Dimension{Unit: Scalar}, Dimension{Unit: Scalar}},
		"border_bottom_width": Value{Dimension: Dimension{Value: 3}},
		"border_left_width":   Value{Dimension: Dimension{Value: 3}},
		"border_top_width":    Value{Dimension: Dimension{Value: 3}}, // computed value for "medium"
		"border_right_width":  Value{Dimension: Dimension{Value: 3}},

		"border_bottom_left_radius":  Values{zeroPixelsValue, zeroPixelsValue},
		"border_bottom_right_radius": Values{zeroPixelsValue, zeroPixelsValue},
		"border_top_left_radius":     Values{zeroPixelsValue, zeroPixelsValue},
		"border_top_right_radius":    Values{zeroPixelsValue, zeroPixelsValue},

		// // Color 3 (REC): https://www.w3.org/TR/css3-color/
		"opacity": Float(1),

		// Multi-column Layout (WD): https://www.w3.org/TR/css-multicol-1/
		"column_width":      SToV("auto"),
		"column_count":      SToV("auto"),
		"column_gap":        Value{Dimension: Dimension{Value: 1, Unit: Em}},
		"column_rule_color": CurrentColor,
		"column_rule_style": String("none"),
		"column_rule_width": SToV("medium"),
		"column_fill":       String("balance"),
		"column_span":       String("none"),

		// Fonts 3 (REC): https://www.w3.org/TR/css-fonts-3/
		"font_family":             Strings{"serif"}, // depends on user agent
		"font_feature_settings":   SIntStrings{String: "normal"},
		"font_kerning":            String("auto"),
		"font_language_override":  String("normal"),
		"font_size":               Value{Dimension: Dimension{Value: 16}}, // actually medium, but we define medium from this
		"font_stretch":            String("normal"),
		"font_style":              String("normal"),
		"font_variant":            String("normal"),
		"font_variant_alternates": String("normal"),
		"font_variant_caps":       String("normal"),
		"font_variant_east_asian": SStrings{String: "normal"},
		"font_variant_ligatures":  SStrings{String: "normal"},
		"font_variant_numeric":    SStrings{String: "normal"},
		"font_variant_position":   String("normal"),
		"font_weight":             IntString{Int: 400},

		// Fragmentation 3/4 (CR/WD): https://www.w3.org/TR/css-break-4/
		"box_decoration_break": String("slice"),
		"break_after":          String("auto"),
		"break_before":         String("auto"),
		"break_inside":         String("auto"),
		"margin_break":         String("auto"),
		"orphans":              Int(2),
		"widows":               Int(2),

		// Generated Content 3 (WD): https://www.w3.org/TR/css-content-3/
		"bookmark_label": BookmarkLabel{{Type: "content", Content: String("text")}},
		"bookmark_level": IntString{String: "none"},
		"bookmark_state": String("open"),
		"content":        SContent{String: "normal"},
		"quotes":         Quotes{Open: []string{"“", "‘"}, Close: []string{"”", "’"}}, // chosen by the user agent
		"string_set":     StringSet{String: "none"},

		// // Images 3/4 (CR/WD): https://www.w3.org/TR/css4-images/
		"image_resolution": FToV(1), // dppx
		"image_rendering":  String("auto"),
		// https://drafts.csswg.org/css-images-3/
		"object_fit": String("fill"),
		"object_position": Center{OriginX: "left", OriginY: "top", Pos: Point{
			Dimension{Value: 50, Unit: Percentage}, Dimension{Value: 50, Unit: Percentage},
		}},

		// Paged Media 3 (WD): https://www.w3.org/TR/css-page-3/
		"size":         Values{},
		"page":         Page{String: "auto", Valid: true},
		"bleed_left":   SToV("auto"),
		"bleed_right":  SToV("auto"),
		"bleed_top":    SToV("auto"),
		"bleed_bottom": SToV("auto"),
		"marks":        Marks{}, //computed value for 'none'

		// Text 3/4 (WD/WD): https://www.w3.org/TR/css-text-4/
		"hyphenate_character":   String("‐"), // computed value chosen by the user agent
		"hyphenate_limit_chars": Ints3{5, 2, 2},
		"hyphenate_limit_zone":  zeroPixelsValue,
		"hyphens":               String("manual"),
		"letter_spacing":        SToV("normal"),
		"tab_size":              Value{Dimension: Dimension{Value: 8}},
		"text_align":            String("-weasy-start"),
		"text_indent":           zeroPixelsValue,
		"text_transform":        String("none"),
		"white_space":           String("normal"),
		"word_spacing":          Value{}, // computed value for "normal"

		// Transforms 1 (CR): https://www.w3.org/TR/css-transforms-1/
		"transform_origin": Point{{Value: 50, Unit: Percentage}, {Value: 50, Unit: Percentage}},
		"transform":        Transforms{}, // computed value for "none"

		// User Interface 3 (REC): https://www.w3.org/TR/css-ui-3/
		"box_sizing":    String("content-box"),
		"outline_color": CurrentColor, // invert is not supported
		"outline_style": String("none"),
		"outline_width": Value{Dimension: Dimension{Value: 3}}, // computed value for "medium"
		"overflow_wrap": String("normal"),

		// Flexible Box Layout Module 1 (CR): https://www.w3.org/TR/css-flexbox-1/
		"align_content":   String("stretch"),
		"align_items":     String("stretch"),
		"align_self":      String("auto"),
		"flex_basis":      SToV("auto"),
		"flex_direction":  String("row"),
		"flex_grow":       Float(0),
		"flex_shrink":     Float(1),
		"flex_wrap":       String("nowrap"),
		"justify_content": String("flex-start"),
		"min_height":      zeroPixelsValue,
		"min_width":       zeroPixelsValue,
		"order":           Int(0),

		// Text Decoration Module 3 (CR): https://www.w3.org/TR/css-text-decor-3/
		"text_decoration_line":  NDecorations{None: true},
		"text_decoration_color": CurrentColor,
		"text_decoration_style": String("solid"),

		// Overflow Module 3 (WD): https://www.w3.org/TR/css-overflow-3/
		"overflow":      String("visible"),
		"text_overflow": String("clip"),

		// Proprietary
		"anchor": NamedString{}, // computed value of "none"
		"link":   NamedString{}, // computed value of "none"
		"lang":   NamedString{}, // computed value of "none"

		// Internal, to implement the "static position" for absolute boxes.
		"_weasy_specified_display": String("inline"),
	}

	KnownProperties = Set{}

	// Do not list shorthand properties here as we handle them before inheritance.
	//
	// text_decoration is not a really inherited, see
	// http://www.w3.org/TR/CSS2/text.html#propdef-text-decoration
	//
	// link: click events normally bubble up to link ancestors
	// See http://lists.w3.org/Archives/Public/www-style/2012Jun/0315.html
	Inherited = NewSet(
		"border_collapse",
		"border_spacing",
		"caption_side",
		"color",
		"direction",
		"empty_cells",
		"font_family",
		"font_feature_settings",
		"font_kerning",
		"font_language_override",
		"font_size",
		"font_style",
		"font_stretch",
		"font_variant",
		"font_variant_alternates",
		"font_variant_caps",
		"font_variant_east_asian",
		"font_variant_ligatures",
		"font_variant_numeric",
		"font_variant_position",
		"font_weight",
		"hyphens",
		"hyphenate_character",
		"hyphenate_limit_chars",
		"hyphenate_limit_zone",
		"image_rendering",
		"image_resolution",
		"lang",
		"letter_spacing",
		"line_height",
		"link",
		"list_style_image",
		"list_style_position",
		"list_style_type",
		"orphans",
		"overflow_wrap",
		"quotes",
		"tab_size",
		"text_align",
		"text_decoration_line",
		"text_decoration_color",
		"text_decoration_style",
		"text_indent",
		"text_transform",
		"visibility",
		"white_space",
		"widows",
		"word_spacing",
	)

	// Not applicable to the print media
	NotPrintMedia = NewSet(
		// Aural media
		"azimuth",
		"cue",
		"cue-after",
		"cue-before",
		"elevation",
		"pause",
		"pause-after",
		"pause-before",
		"pitch-range",
		"pitch",
		"play-during",
		"richness",
		"speak-header",
		"speak-numeral",
		"speak-punctuation",
		"speak",
		"speech-rate",
		"stress",
		"voice-family",
		"volume",
		// Interactive
		"cursor",
		// Animations and transitions
		"animation",
		"animation-delay",
		"animation-direction",
		"animation-duration",
		"animation-fill-mode",
		"animation-iteration-count",
		"animation-name",
		"animation-play-state",
		"animation-timing-function",
		"transition",
		"transition-delay",
		"transition-duration",
		"transition-property",
		"transition-timing-function",
	)

	// http://www.w3.org/TR/CSS21/tables.html#model
	// See also http://lists.w3.org/Archives/Public/www-style/2012Jun/0066.html
	// Only non-inherited properties need to be included here.
	TableWrapperBoxProperties = NewSet(
		"bottom",
		"break_after",
		"break_before",
		"break_inside",
		"clear",
		"counter_increment",
		"counter_reset",
		"float",
		"left",
		"margin_top",
		"margin_bottom",
		"margin_left",
		"margin_right",
		"opacity",
		"overflow",
		"position",
		"right",
		"top",
		"transform",
		"transform_origin",
		"vertical_align",
		"z_index",
	)

	InitialNotComputed = NewSet(
		"display",
		"column_gap",
		"bleed_top",
		"bleed_left",
		"bleed_bottom",
		"bleed_right",
		"outline_width",
		"outline_color",
		"column_rule_width",
		"column_rule_color",
		"border_top_width",
		"border_left_width",
		"border_bottom_width",
		"border_right_width",
		"border_top_color",
		"border_left_color",
		"border_bottom_color",
		"border_right_color",
	)
)

func init() {
	for name := range InitialValues {
		KnownProperties[strings.ReplaceAll(name, "_", "-")] = Has
	}
}

// Properties is the general container for validated and typed properties.
// Each property has a type (not exclusive), so that a CssProperty will be of one
// of the three options :
// 	- nil : means no property
//	- String("inherit") : special case for inherited properties
//	- the concrete type linked to the property, see comment below
// TODO: add a special type for inherited, and update accessors to handle it.
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

// prop:object-position
type Center struct {
	OriginX, OriginY string
	Pos              Point
}

// prop:color
// prop:background-color
type Color parser.Color

// prop:link
type ContentProperty struct {
	Type string

	// SStrings for type STRING, attr or string, counter, counters
	// Quote for type QUOTE
	// Url for URI
	Content InnerContents
}

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

// prop:border-collapse
// prop:break-after
// prop:break-before
// prop:break-inside
// prop:display
// prop:float
// prop:-weasy-specified-display
// prop:position
// prop:list-style-position
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
