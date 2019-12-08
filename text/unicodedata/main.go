package unicodedata

import (
	"unicode"
)

// An enum that works as the states of the Hangul syllables system.
type JamoType int8

const (
	JAMO_LV  JamoType = iota /* break HANGUL_LV_SYLLABLE */
	JAMO_LVT                 /* break HANGUL_LVT_SYLLABLE */
	JAMO_L                   /* break HANGUL_L_JAMO */
	JAMO_V                   /* break HANGUL_V_JAMO */
	JAMO_T                   /* break HANGUL_T_JAMO */
	NO_JAMO                  /* Other */
)

func IsEmojiExtendedPictographic(r rune) bool {
	return unicode.Is(_Extended_Pictographic, r)
}

func IsEmojiBaseCharacter(r rune) bool {
	return unicode.Is(_Emoji, r)
}

func IsVirama(r rune) bool {
	return unicode.Is(_Virama, r)
}

func IsVowelDependent(r rune) bool {
	return unicode.Is(_Vowel_Dependent, r)
}

func BreakClass(r rune) *unicode.RangeTable {
	for _, class := range Breaks {
		if unicode.Is(class, r) {
			return class
		}
	}
	return BreakXX
}

// Jamo returns the Jamo Type of `btype` or NO_JAMO
func Jamo(btype *unicode.RangeTable) JamoType {
	switch btype {
	case BreakH2:
		return JAMO_LV
	case BreakH3:
		return JAMO_LVT
	case BreakJL:
		return JAMO_L
	case BreakJV:
		return JAMO_V
	case BreakJT:
		return JAMO_T
	default:
		return NO_JAMO
	}
}
