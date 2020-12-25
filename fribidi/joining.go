package fribidi

import "unicode"

type JoiningType uint8

const (
	U JoiningType = iota /* nUn-joining, e.g. Full Stop */
	R                    /* Right-joining, e.g. Arabic Letter Dal */
	D                    /* Dual-joining, e.g. Arabic Letter Ain */
	C                    /* join-Causing, e.g. Tatweel, ZWJ */
	T                    /* Transparent, e.g. Arabic Fatha */
	L                    /* Left-joining, i.e. fictional */
	G                    /* iGnored, e.g. LRE, RLE, ZWNBSP */
)

// Define bit masks that joining types are based on
const (
	FRIBIDI_MASK_JOINS_RIGHT = 1 << iota /* May join to right */
	FRIBIDI_MASK_JOINS_LEFT              /* May join to left */
	FRIBIDI_MASK_ARAB_SHAPES             /* May Arabic shape */
	FRIBIDI_MASK_TRANSPARENT             /* Is transparent */
	FRIBIDI_MASK_IGNORED                 /* Is ignored */
	FRIBIDI_MASK_LIGATURED               /* Is ligatured */
)

/* iGnored */
func (p JoiningType) isG() bool {
	return FRIBIDI_MASK_IGNORED == (p)&(FRIBIDI_MASK_TRANSPARENT|FRIBIDI_MASK_IGNORED)
}

/* Is skipped in joining: T, G? */
func (p JoiningType) isJoinSkipped() bool {
	return p&(FRIBIDI_MASK_TRANSPARENT|FRIBIDI_MASK_IGNORED) != 0
}

/* May shape: R, D, L, T? */
func (p JoiningType) isArabShapes() bool {
	return p&FRIBIDI_MASK_ARAB_SHAPES != 0
}

// the return verifies 0 <= s < 4
func (p JoiningType) joinShape() uint8 {
	return uint8(p & (FRIBIDI_MASK_JOINS_RIGHT | FRIBIDI_MASK_JOINS_LEFT))
}

func getJoiningType(ch rune, bidi CharType) JoiningType {
	if jt, ok := joinings[ch]; ok {
		return jt
	}
	// - Those that are not explicitly listed and that are of General Category Mn, Me, or Cf
	//   have joining type T.
	// - All others not explicitly listed have joining type U.
	// general transparents
	if unicode.In(ch, unicode.Mn, unicode.Me, unicode.Cf) {
		return T
	}
	// ignored bidi types
	switch bidi {
	case BN, LRE, RLE, LRO, RLO, PDF, LRI, RLI, FSI, PDI:
		return G
	default:
		return U
	}
}

func getJoiningTypes(str []rune, bidiTypes []CharType) []JoiningType {
	out := make([]JoiningType, len(str))
	for i, r := range str {
		out[i] = getJoiningType(r, bidiTypes[i])
	}
	return out
}

// fribidi_join_arabic does the Arabic joining algorithm.  Means, given Arabic
// joining types of the characters in ar_props, this
// function modifies (in place) this properties to grasp the effect of neighboring
// characters. You probably need this information later to do Arabic shaping.
//
// This function implements rules R1 to R7 inclusive (all rules) of the Arabic
// Cursive Joining algorithm of the Unicode standard as available at
// http://www.unicode.org/versions/Unicode4.0.0/ch08.pdf#G7462.  It also
// interacts correctly with the bidirection algorithm as defined in Section
// 3.5 Shaping of the Unicode Bidirectional Algorithm available at
// http://www.unicode.org/reports/tr9/#Shaping.
func fribidi_join_arabic(bidi_types []CharType, embedding_levels []Level, ar_props []JoiningType) {
	/* The joining algorithm turned out very very dirty :(.  That's what happens
	 * when you follow the standard which has never been implemented closely
	 * before.
	 */

	/* 8.2 Arabic - Cursive Joining */
	var (
		saved                            = 0
		saved_level                Level = FRIBIDI_SENTINEL
		saved_shapes                     = false
		saved_joins_following_mask JoiningType
		joins                      = false
	)
	for i := range ar_props {
		if !ar_props[i].isG() {
			disjoin := false
			shapes := ar_props[i].isArabShapes()

			//  FRIBIDI_CONSISTENT_LEVEL
			var level Level = FRIBIDI_SENTINEL
			if !bidi_types[i].IsExplicitOrBn() {
				level = embedding_levels[i]
			}

			if levelMatch := saved_level == level || saved_level == FRIBIDI_SENTINEL || level == FRIBIDI_SENTINEL; joins && !levelMatch {
				disjoin = true
				joins = false
			}

			if !ar_props[i].isJoinSkipped() {
				var joins_preceding_mask JoiningType = FRIBIDI_MASK_JOINS_LEFT
				if level.isRtl() != 0 {
					joins_preceding_mask = FRIBIDI_MASK_JOINS_RIGHT
				}

				if !joins {
					if shapes {
						ar_props[i] &= ^joins_preceding_mask // unset bits
					}
				} else if ar_props[i]&joins_preceding_mask == 0 { // ! test bits
					disjoin = true
				} else {
					/* This is a FriBidi extension:  we set joining properties
					 * for skipped characters in between, so we can put NSMs on tatweel
					 * later if we want.  Useful on console for example.
					 */
					for j := saved + 1; j < i; j++ {
						ar_props[j] |= joins_preceding_mask | saved_joins_following_mask
					}
				}
			}

			if disjoin && saved_shapes {
				ar_props[saved] &= ^saved_joins_following_mask // unset bits
			}

			if !ar_props[i].isJoinSkipped() {
				saved = i
				saved_level = level
				saved_shapes = shapes
				// FRIBIDI_JOINS_FOLLOWING_MASK(level)
				if level.isRtl() != 0 {
					saved_joins_following_mask = FRIBIDI_MASK_JOINS_LEFT
				} else {
					saved_joins_following_mask = FRIBIDI_MASK_JOINS_RIGHT
				}
				joins = ar_props[i]&saved_joins_following_mask != 0
			}
		}
	}
	if joins && saved_shapes {
		ar_props[saved] &= ^saved_joins_following_mask
	}
}
