package pango

import "github.com/benoitkugler/go-weasyprint/fribidi"

/**
 * SECTION:bidi
 * @short_description:Types and functions for bidirectional text
 * @title:Bidirectional Text
 * @see_also:
 * pango_context_get_base_dir(),
 * pango_context_set_base_dir(),
 * pango_itemize_with_base_dir()
 *
 * Pango supports bidirectional text (like Arabic and Hebrew) automatically.
 * Some applications however, need some help to correctly handle bidirectional text.
 *
 * The Direction type can be used with pango_context_set_base_dir() to
 * instruct Pango about direction of text, though in most cases Pango detects
 * that correctly and automatically.  The rest of the facilities in this section
 * are used internally by Pango already, and are provided to help applications
 * that need more direct control over bidirectional setting of text.
 */

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
// letters, right-to-left letters, and everything else.
func pango_unichar_direction(ch rune) Direction {
	fType := fribidi.GetBidiType(ch)
	if !fType.IsStrong() {
		return PANGO_DIRECTION_NEUTRAL
	} else if fType.IsRtl() {
		return PANGO_DIRECTION_RTL
	} else {
		return PANGO_DIRECTION_LTR
	}
}

// pango_log2vis_get_embedding_levels returns the bidirectional embedding levels of the input paragraph
// as defined by the Unicode Bidirectional Algorithm available at:
//
//   http://www.unicode.org/reports/tr9/
//
// If the input base direction is a weak direction, the direction of the
// characters in the text will determine the final resolved direction.
// The embedding levels slice as one item per Unicode character.
func pango_log2vis_get_embedding_levels(text []rune, pbase_dir Direction) (Direction, []uint8) {
	// glong n_chars, i;
	// guint8 *embedding_levels_list;
	// const gchar *p;
	// FriBidiParType fribidi_base_dir;
	// FriBidiCharType *bidi_types;
	//   #ifdef USE_FRIBIDI_EX_API
	// FriBidiBracketType *bracket_types;
	//   #endif
	// FriBidiLevel max_level;
	// FriBidiCharType ored_types = 0;
	anded_strongs := FRIBIDI_TYPE_RLE

	// G_STATIC_ASSERT (sizeof (FriBidiLevel) == sizeof (guint8));
	// G_STATIC_ASSERT (sizeof (FriBidiChar) == sizeof (rune));

	switch pbase_dir {
	case PANGO_DIRECTION_LTR, PANGO_DIRECTION_TTB_RTL:
		fribidi_base_dir = FRIBIDI_PAR_LTR
	case PANGO_DIRECTION_RTL, PANGO_DIRECTION_TTB_LTR:
		fribidi_base_dir = FRIBIDI_PAR_RTL
	case PANGO_DIRECTION_WEAK_RTL:
		fribidi_base_dir = FRIBIDI_PAR_WRTL
	case PANGO_DIRECTION_WEAK_LTR, PANGO_DIRECTION_NEUTRAL:
	default:
		fribidi_base_dir = FRIBIDI_PAR_WLTR
	}

	n_chars := len(text)

	bidi_types = g_new(FriBidiCharType, n_chars)
	bracket_types = g_new(FriBidiBracketType, n_chars)
	embedding_levels_list = g_new(guint8, n_chars)

	for i, p := 0, text; p < text+length; i, p = i+1, g_utf8_next_char(p) {
		ch := g_utf8_get_char(p)
		char_type := fribidi.GetBidiType(ch)

		if i == n_chars {
			break
		}

		bidi_types[i] = char_type
		ored_types |= char_type
		if FRIBIDI_IS_STRONG(char_type) {
			anded_strongs &= char_type
		}
		if G_UNLIKELY(bidi_types[i] == FRIBIDI_TYPE_ON) {
			bracket_types[i] = fribidi_get_bracket(ch)
		} else {
			bracket_types[i] = FRIBIDI_NO_BRACKET
		}
	}

	/* Short-circuit (malloc-expensive) FriBidi call for unidirectional
	 * text.
	 *
	 * For details see:
	 * https://bugzilla.gnome.org/show_bug.cgi?id=590183
	 */

	//   #ifndef FRIBIDI_IS_ISOLATE
	//   #define FRIBIDI_IS_ISOLATE(x) 0
	//   #endif
	/* The case that all resolved levels will be ltr.
	 * No isolates, all strongs be LTR, there should be no Arabic numbers
	 * (or letters for that matter), and one of the following:
	 *
	 * o base_dir doesn't have an RTL taste.
	 * o there are letters, and base_dir is weak.
	 */
	if !FRIBIDI_IS_ISOLATE(ored_types) &&
		!FRIBIDI_IS_RTL(ored_types) &&
		!FRIBIDI_IS_ARABIC(ored_types) &&
		(!FRIBIDI_IS_RTL(fribidi_base_dir) ||
			(FRIBIDI_IS_WEAK(fribidi_base_dir) &&
				FRIBIDI_IS_LETTER(ored_types))) {
		/* all LTR */
		fribidi_base_dir = FRIBIDI_PAR_LTR
		memset(embedding_levels_list, 0, n_chars)
		goto resolved
	} else if !FRIBIDI_IS_ISOLATE(ored_types) &&
		!FRIBIDI_IS_NUMBER(ored_types) &&
		FRIBIDI_IS_RTL(anded_strongs) &&
		(FRIBIDI_IS_RTL(fribidi_base_dir) ||
			(FRIBIDI_IS_WEAK(fribidi_base_dir) &&
				FRIBIDI_IS_LETTER(ored_types))) {
		/* The case that all resolved levels will be RTL is much more complex.
		 * No isolates, no numbers, all strongs are RTL, and one of
		 * the following:
		 *
		 * o base_dir has an RTL taste (may be weak).
		 * o there are letters, and base_dir is weak.
		 */

		/* all RTL */
		fribidi_base_dir = FRIBIDI_PAR_RTL
		memset(embedding_levels_list, 1, n_chars)
		goto resolved
	}

	max_level = fribidi_get_par_embedding_levels_ex(bidi_types, bracket_types, n_chars,
		&fribidi_base_dir, embedding_levels_list)

	if max_level == 0 {
		/* fribidi_get_par_embedding_levels() failed. */
		memset(embedding_levels_list, 0, length)
	}

resolved:
	g_free(bidi_types)

	*pbase_dir = PANGO_DIRECTION_RTL
	if fribidi_base_dir == FRIBIDI_PAR_LTR {
		pbase_dir = PANGO_DIRECTION_LTR
	}

	return pbase_dir, embedding_levels_list
}

//   /**
//    * pango_unichar_direction:
//    * @ch: a Unicode character
//    *
//    * Determines the inherent direction of a character; either
//    * %PANGO_DIRECTION_LTR, %PANGO_DIRECTION_RTL, or
//    * %PANGO_DIRECTION_NEUTRAL.
//    *
//    * This function is useful to categorize characters into left-to-right
//    * letters, right-to-left letters, and everything else.  If full
//    * Unicode bidirectional type of a character is needed,
//    * pango_bidi_type_for_unichar() can be used instead.
//    *
//    * Return value: the direction of the character.
//    */
//   PangoDirection
//   pango_unichar_direction (rune ch)
//   {
// 	FriBidiCharType fribidi_ch_type;

// 	G_STATIC_ASSERT (sizeof (FriBidiChar) == sizeof (rune));

// 	fribidi_ch_type = fribidi.GetBidiType (ch);

// 	if (!FRIBIDI_IS_STRONG (fribidi_ch_type))
// 	  return PANGO_DIRECTION_NEUTRAL;
// 	else if (FRIBIDI_IS_RTL (fribidi_ch_type))
// 	  return PANGO_DIRECTION_RTL;
// 	else
// 	  return PANGO_DIRECTION_LTR;
//   }

//   /**
//    * pango_get_mirror_char:
//    * @ch: a Unicode character
//    * @mirrored_ch: location to store the mirrored character
//    *
//    * If @ch has the Unicode mirrored property and there is another Unicode
//    * character that typically has a glyph that is the mirror image of @ch's
//    * glyph, puts that character in the address pointed to by @mirrored_ch.
//    *
//    * Use g_unichar_get_mirror_char() instead; the docs for that function
//    * provide full details.
//    *
//    * Return value: %TRUE if @ch has a mirrored character and @mirrored_ch is
//    * filled in, %FALSE otherwise
//    **/
//   gboolean
//   pango_get_mirror_char (rune        ch,
// 				 rune       *mirrored_ch)
//   {
// 	return g_unichar_get_mirror_char (ch, mirrored_ch);
//   }
