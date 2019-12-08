package text

import "github.com/benoitkugler/go-weasyprint/text/unicodedata"

// The PangoLogAttr structure stores information
// about the attributes of a single character.
type PangoLogAttr struct {
	// if set, can break line in front of character
	IsLineBreak bool

	// if set, must break line in front of character
	IsMandatoryBreak bool

	// if set, can break here when doing character wrapping
	IsCharBreak bool

	// is whitespace character
	IsWhite bool

	// if set, cursor can appear in front of character.
	// i.e. this is a grapheme boundary, or the first character
	// in the text.
	// This flag implements Unicode's
	// <ulink url="http://www.unicode.org/reports/tr29/">Grapheme
	// Cluster Boundaries</ulink> semantics.
	IsCursorPosition bool

	// is first character in a word
	IsWordStart bool

	// is first non-word char after a word
	// 	Note that in degenerate cases, you could have both is_word_start
	//  and is_word_end set for some character.
	IsWordEnd bool

	// is a sentence boundary.
	// There are two ways to divide sentences. The first assigns all
	// inter-sentence whitespace/control/format chars to some sentence,
	// so all chars are in some sentence; @is_sentence_boundary denotes
	// the boundaries there. The second way doesn't assign
	// between-sentence spaces, etc. to any sentence, so
	// @is_sentence_start/@is_sentence_end mark the boundaries of those sentences.
	IsSentenceBoundary bool

	// is first character in a sentence
	IsSentenceStart bool

	// is first char after a sentence.
	// Note that in degenerate cases, you could have both @is_sentence_start
	// and @is_sentence_end set for some character. (e.g. no space after a
	// period, so the next sentence starts right away)
	IsSentenceEnd bool

	//  if set, backspace deletes one character
	// rather than the entire grapheme cluster. This
	// field is only meaningful on grapheme
	// boundaries (where @is_cursor_position is
	// set).  In some languages, the full grapheme
	// (e.g.  letter + diacritics) is considered a
	// unit, while in others, each decomposed
	// character in the grapheme is a unit. In the
	// default implementation of pango_break(), this
	// bit is set on all grapheme boundaries except
	// those following Latin, Cyrillic or Greek base characters.
	BackspaceDeletesCharacter bool

	// is a whitespace character that can possibly be
	// expanded for justification purposes. (Since: 1.18)
	IsExpandableSpace bool

	// is a word boundary, as defined by UAX#29.
	// More specifically, means that this is not a position in the middle
	// of a word.  For example, both sides of a punctuation mark are
	// considered word boundaries.  This flag is particularly useful when
	// selecting text word-by-word.
	// This flag implements Unicode's
	// <ulink url="http://www.unicode.org/reports/tr29/">Word
	// Boundaries</ulink> semantics. (Since: 1.22)
	IsWordBoundary bool
}

// See Grapheme_Cluster_Break Property Values table of UAX#29
type graphemeBreakType uint8

const (
	gb_Other graphemeBreakType = iota
	gb_ControlCRLF
	gb_Extend
	gb_ZWJ
	gb_Prepend
	gb_SpacingMark
	gb_InHangulSyllable /* Handles all of L, V, T, LV, LVT rules */
	/* Use state machine to handle emoji sequence */
	/* Rule GB12 and GB13 */
	gb_RI_Odd  /* Meets odd number of RI */
	gb_RI_Even /* Meets even number of RI */
)

/* See Word_Break Property Values table of UAX#29 */
type wordBreakType uint8

const (
	wb_Other wordBreakType = iota
	wb_NewlineCRLF
	wb_ExtendFormat
	wb_Katakana
	wb_Hebrew_Letter
	wb_ALetter
	wb_MidNumLet
	wb_MidLetter
	wb_MidNum
	wb_Numeric
	wb_ExtendNumLet
	wb_RI_Odd
	wb_RI_Even
	wb_WSegSpace
)

/* See Sentence_Break Property Values table of UAX#29 */
type sentenceBreakType uint8

const (
	sb_Other sentenceBreakType = iota
	sb_ExtendFormat
	sb_ParaSep
	sb_Sp
	sb_Lower
	sb_Upper
	sb_OLetter
	sb_Numeric
	sb_ATerm
	sb_SContinue
	sb_STerm
	sb_Close
	/* Rules SB8 and SB8a */
	sb_ATerm_Close_Sp
	sb_STerm_Close_Sp
)

/* Rule LB25 with Example 7 of Customization */
type lineBreakType uint8

const (
	lb_Other lineBreakType = iota
	lb_Numeric
	lb_Numeric_Close
	lb_RI_Odd
	lb_RI_Even
)

// Previously "123foo" was two words. But in UAX 29 of Unicode,
// we now don't break words between consecutive letters and numbers
type wordType uint8

const (
	wordNone wordType = iota
	wordLetters
	wordNumbers
)

type breakOpportunity uint8

const (
	break_ALREADY_HANDLED breakOpportunity = iota /* didn't use the table */
	break_PROHIBITED                              /* no break, even if spaces intervene */
	break_IF_SPACES                               /* "indirect break" (only if there are spaces) */
	break_ALLOWED                                 /* "direct break" (can always break here) */
	// TR 14 has two more break-opportunity classes,
	// "indirect break opportunity for combining marks following a space"
	// and "prohibited break for combining marks"
	// but we handle that inline in the code.
)

// There are Hangul syllables encoded as characters, that act like a
// sequence of Jamos. For each character we define a JamoType
// that the character starts with, and one that it ends with.  This
// decomposes JAMO_LV and JAMO_LVT to simple other JAMOs.  So for
// example, a character with LineBreak type
// G_UNICODE_BREAK_HANGUL_LV_SYLLABLE has start=JAMO_L and end=JAMO_V.
type charJamoProps struct {
	start, end unicodedata.JamoType
}

/* Map from JamoType to CharJamoProps that hold only simple
 * JamoTypes (no LV or LVT) or none.
 */
var HangulJamoProps = [6]charJamoProps{
	{unicodedata.JAMO_L, unicodedata.JAMO_L},   /* JAMO_L */
	{unicodedata.JAMO_V, unicodedata.JAMO_V},   /* JAMO_V */
	{unicodedata.JAMO_T, unicodedata.JAMO_T},   /* JAMO_T */
	{unicodedata.JAMO_L, unicodedata.JAMO_V},   /* JAMO_LV */
	{unicodedata.JAMO_L, unicodedata.JAMO_T},   /* JAMO_LVT */
	{unicodedata.NO_JAMO, unicodedata.NO_JAMO}, /* NO_JAMO */
}
