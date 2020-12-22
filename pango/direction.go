package pango

import "github.com/benoitkugler/go-weasyprint/fridibi"

// The Direction type represents a direction in the
// Unicode bidirectional algorithm.
//
// Not every value in this
// enumeration makes sense for every usage of Direction;
// for example, the return value of pango_unichar_direction()
// and pango_find_base_dir() cannot be `PANGO_DIRECTION_WEAK_LTR`
// or `PANGO_DIRECTION_WEAK_RTL`, since every character is either
// neutral or has a strong direction; on the other hand
// `PANGO_DIRECTION_NEUTRAL` doesn't make sense to pass
// to pango_itemize_with_base_dir().
//
// See `Gravity` for how vertical text is handled in Pango.
//
// If you are interested in text direction, you should
// really use fribidi directly. Direction is only
// retained because it is used in some public apis.
type Direction uint8

const (
	PANGO_DIRECTION_LTR      Direction = iota // A strong left-to-right direction
	PANGO_DIRECTION_RTL                       // A strong right-to-left direction
	_                                         // Deprecated value; treated the same as `PANGO_DIRECTION_RTL`.
	_                                         // Deprecated value; treated the same as `PANGO_DIRECTION_LTR`
	PANGO_DIRECTION_WEAK_LTR                  // A weak left-to-right direction
	PANGO_DIRECTION_WEAK_RTL                  // A weak right-to-left direction
	PANGO_DIRECTION_NEUTRAL                   // No direction specified
)

/**
 * pango_find_base_dir:
 * @text:   the text to process. Must be valid UTF-8
 * @length: length of @text in bytes (may be -1 if @text is nul-terminated)
 *
 * Searches a string the first character that has a strong
 * direction, according to the Unicode bidirectional algorithm.
 *
 * Return value: The direction corresponding to the first strong character.
 * If no such character is found, then `PANGO_DIRECTION_NEUTRAL` is returned.
 *
 * Since: 1.4
 */
func pango_find_base_dir(text []rune) Direction {
	dir := PANGO_DIRECTION_NEUTRAL
	for _, wc := range text {
		dir = pango_unichar_direction(wc)
		if dir != PANGO_DIRECTION_NEUTRAL {
			break
		}
	}

	return dir
}

// pango_unichar_direction determines the inherent direction of a character; either
// `PANGO_DIRECTION_LTR`, `PANGO_DIRECTION_RTL`, or
// `PANGO_DIRECTION_NEUTRAL`.
//
// This function is useful to categorize characters into left-to-right
// letters, right-to-left letters, and everything else.  If full
// Unicode bidirectional type of a character is needed,
// pango_bidi_type_for_unichar() can be used instead.
func pango_unichar_direction(ch rune) Direction {
	fType := fridibi.GetBidiType(ch)
	if !fType.IsStrong() {
		return PANGO_DIRECTION_NEUTRAL
	} else if fType.IsRtl() {
		return PANGO_DIRECTION_RTL
	} else {
		return PANGO_DIRECTION_LTR
	}
}
