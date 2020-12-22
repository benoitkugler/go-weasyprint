package pango

import (
	"fmt"
	"strings"
)

// Style specifies the various slant styles possible for a font.
type Style uint8

const (
	PANGO_STYLE_NORMAL  Style = iota //  the font is upright.
	PANGO_STYLE_OBLIQUE              //  the font is slanted, but in a roman style.
	PANGO_STYLE_ITALIC               //  the font is slanted in an italic style.
)

var style_map = [...]string{
	// { PANGO_STYLE_NORMAL, "" },
	PANGO_STYLE_NORMAL:  "Roman",
	PANGO_STYLE_OBLIQUE: "Oblique",
	PANGO_STYLE_ITALIC:  "Italic",
}

// Variant specifies capitalization variant of the font.
type Variant uint8

const (
	PANGO_VARIANT_NORMAL     Variant = iota // A normal font.
	PANGO_VARIANT_SMALL_CAPS                // A font with the lower case characters replaced by smaller variants of the capital characters.
)

var variant_map = [...]string{
	PANGO_VARIANT_NORMAL:     "",
	PANGO_VARIANT_SMALL_CAPS: "Small-Caps",
}

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

var weight_map = [...]string{
	PANGO_WEIGHT_THIN:       "Thin",
	PANGO_WEIGHT_ULTRALIGHT: "Ultra-Light",
	// PANGO_WEIGHT_ULTRALIGHT: "Extra-Light" ,
	PANGO_WEIGHT_LIGHT:     "Light",
	PANGO_WEIGHT_SEMILIGHT: "Semi-Light",
	// PANGO_WEIGHT_SEMILIGHT: "Demi-Light" ,
	PANGO_WEIGHT_BOOK: "Book",
	// PANGO_WEIGHT_NORMAL: "" ,
	PANGO_WEIGHT_NORMAL:   "Regular",
	PANGO_WEIGHT_MEDIUM:   "Medium",
	PANGO_WEIGHT_SEMIBOLD: "Semi-Bold",
	// PANGO_WEIGHT_SEMIBOLD: "Demi-Bold" ,
	PANGO_WEIGHT_BOLD:      "Bold",
	PANGO_WEIGHT_ULTRABOLD: "Ultra-Bold",
	// PANGO_WEIGHT_ULTRABOLD: "Extra-Bold" ,
	PANGO_WEIGHT_HEAVY: "Heavy",
	// PANGO_WEIGHT_HEAVY: "Black" ,
	PANGO_WEIGHT_ULTRAHEAVY: "Ultra-Heavy",
	// PANGO_WEIGHT_ULTRAHEAVY: "Extra-Heavy" ,
	// PANGO_WEIGHT_ULTRAHEAVY: "Ultra-Black" ,
	// PANGO_WEIGHT_ULTRAHEAVY: "Extra-Black",
}

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

var stretch_map = [...]string{
	PANGO_STRETCH_ULTRA_CONDENSED: "Ultra-Condensed",
	PANGO_STRETCH_EXTRA_CONDENSED: "Extra-Condensed",
	PANGO_STRETCH_CONDENSED:       "Condensed",
	PANGO_STRETCH_SEMI_CONDENSED:  "Semi-Condensed",
	PANGO_STRETCH_NORMAL:          "",
	PANGO_STRETCH_SEMI_EXPANDED:   "Semi-Expanded",
	PANGO_STRETCH_EXPANDED:        "Expanded",
	PANGO_STRETCH_EXTRA_EXPANDED:  "Extra-Expanded",
	PANGO_STRETCH_ULTRA_EXPANDED:  "Ultra-Expanded",
}

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

var pfd_defaults = FontDescription{
	family_name: "",

	style:      PANGO_STYLE_NORMAL,
	variant:    PANGO_VARIANT_NORMAL,
	weight:     PANGO_WEIGHT_NORMAL,
	stretch:    PANGO_STRETCH_NORMAL,
	gravity:    PANGO_GRAVITY_SOUTH,
	variations: "",

	mask:             0,
	size_is_absolute: false,

	size: 0,
}

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

	mask             FontMask
	size_is_absolute bool // = : 1;

	size int
}

// Creates a string representation of a font description.
// The family list in the string description will only have a terminating comma if the
// last word of the list is a valid style option.
func (desc FontDescription) String() string {
	var result string
	if desc.family_name != "" && (desc.mask&PANGO_FONT_MASK_FAMILY != 0) {
		result += desc.family_name

		/* We need to add a trailing comma if the family name ends
		* in a keyword like "Bold", or if the family name ends in
		* a number and no keywords will be added.
		 */
		// TODO:
		// strings.Split(desc.family_name, ",")
		//    p = getword (desc.family_name, desc.family_name + strlen(desc.family_name), &wordlen, ",");
		if desc.weight == PANGO_WEIGHT_NORMAL &&
			desc.style == PANGO_STYLE_NORMAL &&
			desc.stretch == PANGO_STRETCH_NORMAL &&
			desc.variant == PANGO_VARIANT_NORMAL &&
			(desc.mask&(PANGO_FONT_MASK_GRAVITY|PANGO_FONT_MASK_SIZE) == 0) {
			result += ","
		}
	}

	result += weight_map[desc.weight]
	result += style_map[desc.style]
	result += stretch_map[desc.stretch]
	result += variant_map[desc.variant]
	if desc.mask&PANGO_FONT_MASK_GRAVITY != 0 {
		result += gravity_map[desc.gravity]
	}

	if len(result) == 0 {
		result += "Normal"
	}
	if desc.mask&PANGO_FONT_MASK_SIZE != 0 {
		if result[len(result)-1] != ' ' {
			result = " "
		}

		result += fmt.Sprintf("%g", float64(desc.size)/pangoScale)

		if desc.size_is_absolute {
			result += "px"
		}
	}

	if desc.variations != "" && desc.mask&PANGO_FONT_MASK_VARIATIONS != 0 {
		result += " @"
		result += desc.variations
	}

	return result
}

// pango_font_description_equal compares two font descriptions for equality.
// Two font descriptions are considered equal if the fonts they describe are provably identical.
// This means that their masks do not have to match, as long as other fields
// are all the same.
// Note that two font descriptions may result in identical fonts
// being loaded, but still compare `false`.
func (desc1 FontDescription) pango_font_description_equal(desc2 FontDescription) bool {
	return desc1.style == desc2.style &&
		desc1.variant == desc2.variant &&
		desc1.weight == desc2.weight &&
		desc1.stretch == desc2.stretch &&
		desc1.size == desc2.size &&
		desc1.size_is_absolute == desc2.size_is_absolute &&
		desc1.gravity == desc2.gravity &&
		(desc1.family_name == desc2.family_name || strings.EqualFold(desc1.family_name, desc2.family_name)) &&
		desc1.variations == desc2.variations
}

// Sets the style field of a FontDescription. The
// Style enumeration describes whether the font is slanted and
// the manner in which it is slanted; it can be either
// `STYLE_NORMAL`, `STYLE_ITALIC`, or `STYLE_OBLIQUE`.
// Most fonts will either have a italic style or an oblique
// style, but not both, and font matching in Pango will
// match italic specifications with oblique fonts and vice-versa
// if an exact match is not found.
func (desc *FontDescription) pango_font_description_set_style(style Style) {
	if desc == nil {
		return
	}

	desc.style = style
	desc.mask |= PANGO_FONT_MASK_STYLE
}

// Sets the size field of a font description in fractional points.
// `size` is the size of the font in points, scaled by PANGO_SCALE.
// That is, a `size` value of 10 * PANGO_SCALE is a 10 point font. The conversion
// factor between points and device units depends on system configuration
// and the output device. For screen display, a logical DPI of 96 is
// common, in which case a 10 point font corresponds to a 10 * (96 / 72) = 13.3
// pixel font.
//
// This is mutually exclusive with pango_font_description_set_absolute_size(),
// to use if you need a particular size in device units
func (desc *FontDescription) pango_font_description_set_size(size int) {
	if desc == nil || size < 0 {
		return
	}

	desc.size = size
	desc.size_is_absolute = false
	desc.mask |= PANGO_FONT_MASK_SIZE
}

// Sets the size field of a font description, in device units.
// `size` is the new size, in Pango units. There are `PANGO_SCALE` Pango units in one
// device unit. For an output backend where a device unit is a pixel, a `size`
// value of 10 * PANGO_SCALE gives a 10 pixel font.
//
// This is mutually exclusive with pango_font_description_set_size() which sets the font size
// in points.
func (desc *FontDescription) pango_font_description_set_absolute_size(size int) {
	if desc == nil || size < 0 {
		return
	}

	desc.size = size
	desc.size_is_absolute = true
	desc.mask |= PANGO_FONT_MASK_SIZE
}

// Sets the stretch field of a font description. The stretch field
// specifies how narrow or wide the font should be.
func (desc *FontDescription) pango_font_description_set_stretch(stretch Stretch) {
	if desc == nil {
		return
	}
	desc.stretch = stretch
	desc.mask |= PANGO_FONT_MASK_STRETCH
}

// Sets the weight field of a font description. The weight field
// specifies how bold or light the font should be. In addition
// to the values of the Weight enumeration, other intermediate
// numeric values are possible.
func (desc *FontDescription) pango_font_description_set_weight(weight Weight) {
	if desc == nil {
		return
	}

	desc.weight = weight
	desc.mask |= PANGO_FONT_MASK_WEIGHT
}

// Sets the variant field of a font description. The variant
// can either be `VARIANT_NORMAL` or `VARIANT_SMALL_CAPS`.
func (desc *FontDescription) pango_font_description_set_variant(variant Variant) {
	if desc == nil {
		return
	}

	desc.variant = variant
	desc.mask |= PANGO_FONT_MASK_VARIANT
}

// pango_font_description_set_family sets the family name field of a font description. The family
// name represents a family of related font styles, and will
// resolve to a particular `FontFamily`. In some uses of
// `FontDescription`, it is also possible to use a comma
// separated list of family names for this field.
func (desc *FontDescription) pango_font_description_set_family(family string) {
	if desc == nil || desc.family_name == family {
		return
	}

	if family != "" {
		desc.family_name = family
		desc.mask |= PANGO_FONT_MASK_FAMILY
	} else {
		desc.family_name = pfd_defaults.family_name
		desc.mask &= ^PANGO_FONT_MASK_FAMILY
	}
}

// Merges the fields that are set in `descToMerge` into the fields in
// `desc`.  If `replaceExisting `is `false`, only fields in `desc` that
// are not already set are affected. If `true`, then fields that are
// already set will be replaced as well.
//
// If `descToMerge` is nil, this function performs nothing.
func (desc *FontDescription) pango_font_description_merge(descToMerge *FontDescription, replaceExisting bool) {
	if desc == nil || descToMerge == nil {
		return
	}
	var newMask FontMask
	if replaceExisting {
		newMask = descToMerge.mask
	} else {
		newMask = descToMerge.mask & ^desc.mask
	}
	if newMask&PANGO_FONT_MASK_FAMILY != 0 {
		desc.pango_font_description_set_family(descToMerge.family_name)
	}
	if newMask&PANGO_FONT_MASK_STYLE != 0 {
		desc.style = descToMerge.style
	}
	if newMask&PANGO_FONT_MASK_VARIANT != 0 {
		desc.variant = descToMerge.variant
	}
	if newMask&PANGO_FONT_MASK_WEIGHT != 0 {
		desc.weight = descToMerge.weight
	}
	if newMask&PANGO_FONT_MASK_STRETCH != 0 {
		desc.stretch = descToMerge.stretch
	}
	if newMask&PANGO_FONT_MASK_SIZE != 0 {
		desc.size = descToMerge.size
		desc.size_is_absolute = descToMerge.size_is_absolute
	}
	if newMask&PANGO_FONT_MASK_GRAVITY != 0 {
		desc.gravity = descToMerge.gravity
	}
	if newMask&PANGO_FONT_MASK_VARIATIONS != 0 {
		if descToMerge.variations != "" {
			desc.variations = descToMerge.variations
			desc.mask |= PANGO_FONT_MASK_VARIATIONS
		} else {
			desc.variations = pfd_defaults.variations
			desc.mask &= ^PANGO_FONT_MASK_VARIATIONS
		}
	}
	desc.mask |= newMask
}

// Unsets some of the fields in a `FontDescription`.  The unset
// fields will get back to their default values.
func (desc *FontDescription) pango_font_description_unset_fields(toUnset FontMask) {
	if desc == nil {
		return
	}

	unsetDesc := pfd_defaults
	unsetDesc.mask = toUnset

	desc.pango_font_description_merge(&unsetDesc, true)

	desc.mask &= ^toUnset
}
