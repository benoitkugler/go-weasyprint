package pango

// Font is used to represent a font in a rendering-system-independent matter.
type Font interface {
	// TODO:
}

// FontDescription represents the description
// of an ideal font. These structures are used both to list
// what fonts are available on the system and also for specifying
// the characteristics of a font to load.
type FontDescription struct {
	family_name string

	style   Style
	variant Variant
	weight  Weight
	stretch Stretch
	gravity Gravity

	variations string

	mask uint16
	//   guint static_family : 1;
	//   guint static_variations : 1;
	size_is_absolute bool // = : 1;

	size int
}

// Style specifies the various slant styles possible for a font.
type Style uint8

const (
	PANGO_STYLE_NORMAL  Style = iota //  the font is upright.
	PANGO_STYLE_OBLIQUE              //  the font is slanted, but in a roman style.
	PANGO_STYLE_ITALIC               //  the font is slanted in an italic style.
)

// Variant specifies capitalization variant of the font.
type Variant uint8

const (
	PANGO_VARIANT_NORMAL     Variant = iota // A normal font.
	PANGO_VARIANT_SMALL_CAPS                // A font with the lower case characters replaced by smaller variants of the capital characters.
)

//  Weight specifies the weight (boldness) of a font. This is a numerical
//  value ranging from 100 to 1000, but there are some predefined values:
type Weight int

const (
	PANGO_WEIGHT_THIN       Weight = 100  // the thin weight (= 100; Since: 1.24)
	PANGO_WEIGHT_ULTRALIGHT Weight = 200  // the ultralight weight (= 200)
	PANGO_WEIGHT_LIGHT      Weight = 300  // the light weight (= 300)
	PANGO_WEIGHT_SEMILIGHT  Weight = 350  // the semilight weight (= 350; Since: 1.36.7)
	PANGO_WEIGHT_BOOK       Weight = 380  // the book weight (= 380; Since: 1.24)
	PANGO_WEIGHT_NORMAL     Weight = 400  // the default weight (= 400)
	PANGO_WEIGHT_MEDIUM     Weight = 500  // the normal weight (= 500; Since: 1.24)
	PANGO_WEIGHT_SEMIBOLD   Weight = 600  // the semibold weight (= 600)
	PANGO_WEIGHT_BOLD       Weight = 700  // the bold weight (= 700)
	PANGO_WEIGHT_ULTRABOLD  Weight = 800  // the ultrabold weight (= 800)
	PANGO_WEIGHT_HEAVY      Weight = 900  // the heavy weight (= 900)
	PANGO_WEIGHT_ULTRAHEAVY Weight = 1000 // the ultraheavy weight (= 1000; Since: 1.24)
)

//  Stretch specifies the width of the font relative to other designs within a family.
type Stretch uint8

const (
	PANGO_STRETCH_ULTRA_CONDENSED Stretch = iota // ultra condensed width
	PANGO_STRETCH_EXTRA_CONDENSED                // extra condensed width
	PANGO_STRETCH_CONDENSED                      // condensed width
	PANGO_STRETCH_SEMI_CONDENSED                 // semi condensed width
	PANGO_STRETCH_NORMAL                         // the normal width
	PANGO_STRETCH_SEMI_EXPANDED                  // semi expanded width
	PANGO_STRETCH_EXPANDED                       // expanded width
	PANGO_STRETCH_EXTRA_EXPANDED                 // extra expanded width
	PANGO_STRETCH_ULTRA_EXPANDED                 // ultra expanded width
)

// FontMask bits correspond to fields in a `FontDescription` that have been set.
type FontMask int16

const (
	PANGO_FONT_MASK_FAMILY     FontMask = 1 << iota // the font family is specified.
	PANGO_FONT_MASK_STYLE                           // the font style is specified.
	PANGO_FONT_MASK_VARIANT                         // the font variant is specified.
	PANGO_FONT_MASK_WEIGHT                          // the font weight is specified.
	PANGO_FONT_MASK_STRETCH                         // the font stretch is specified.
	PANGO_FONT_MASK_SIZE                            // the font size is specified.
	PANGO_FONT_MASK_GRAVITY                         // the font gravity is specified (Since: 1.16.)
	PANGO_FONT_MASK_VARIATIONS                      // OpenType font variations are specified (Since: 1.42)
)

/* CSS scale factors (1.2 factor between each size) */
const (
	PANGO_SCALE_XX_SMALL = 0.5787037037037 //  The scale factor for three shrinking steps (1 / (1.2 * 1.2 * 1.2)).
	PANGO_SCALE_X_SMALL  = 0.6944444444444 //  The scale factor for two shrinking steps (1 / (1.2 * 1.2)).
	PANGO_SCALE_SMALL    = 0.8333333333333 //  The scale factor for one shrinking step (1 / 1.2).
	PANGO_SCALE_MEDIUM   = 1.0             //  The scale factor for normal size (1.0).
	PANGO_SCALE_LARGE    = 1.2             //  The scale factor for one magnification step (1.2).
	PANGO_SCALE_X_LARGE  = 1.44            //  The scale factor for two magnification steps (1.2 * 1.2).
	PANGO_SCALE_XX_LARGE = 1.728           //  The scale factor for three magnification steps (1.2 * 1.2 * 1.2).
)
