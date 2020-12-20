package text

const (
	LineBreak CharAttr = 1 << iota
	MandatoryBreak
	CharBreak
	White
	CursorPosition
	WordStart
	WordEnd
	SentenceBoundary
	SentenceStart
	SentenceEnd
	BackspaceDeletesCharacter
	ExpandableSpace
	WordBoundary
)

// CharAttr stores information
// about the attributes of a single character.
type CharAttr uint16

// if set, can break line in front of character
func (c CharAttr) IsLineBreak() bool {
	return c&LineBreak != 0
}

// if set, must break line in front of character
func (c CharAttr) IsMandatoryBreak() bool {
	return c&MandatoryBreak != 0
}

// if set, can break here when doing character wrapping
func (c CharAttr) IsCharBreak() bool {
	return c&CharBreak != 0
}

// is whitespace character
func (c CharAttr) IsWhite() bool {
	return c&White != 0
}

// if set, cursor can appear in front of character.
// i.e. this is a grapheme boundary, or the first character
// in the text.
// This flag implements Unicode's
// <ulink url="http://www.unicode.org/reports/tr29/">Grapheme
// Cluster Boundaries</ulink> semantics.
func (c CharAttr) IsCursorPosition() bool {
	return c&CursorPosition != 0
}

// is first character in a word
func (c CharAttr) IsWordStart() bool {
	return c&WordStart != 0
}

// is first non-word char after a word
// 	Note that in degenerate cases, you could have both is_word_start
//  and is_word_end set for some character.
func (c CharAttr) IsWordEnd() bool {
	return c&WordEnd != 0
}

// is a sentence boundary.
// There are two ways to divide sentences. The first assigns all
// inter-sentence whitespace/control/format chars to some sentence,
// so all chars are in some sentence; @is_sentence_boundary denotes
// the boundaries there. The second way doesn't assign
// between-sentence spaces, etc. to any sentence, so
// @is_sentence_start/@is_sentence_end mark the boundaries of those sentences.
func (c CharAttr) IsSentenceBoundary() bool {
	return c&SentenceBoundary != 0
}

// is first character in a sentence
func (c CharAttr) IsSentenceStart() bool {
	return c&SentenceStart != 0
}

// is first char after a sentence.
// Note that in degenerate cases, you could have both @is_sentence_start
// and @is_sentence_end set for some character. (e.g. no space after a
// period, so the next sentence starts right away)
func (c CharAttr) IsSentenceEnd() bool {
	return c&SentenceEnd != 0
}

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
func (c CharAttr) IsBackspaceDeletesCharacter() bool {
	return c&BackspaceDeletesCharacter != 0
}

// is a whitespace character that can possibly be
// expanded for justification purposes. (Since: 1.18)
func (c CharAttr) IsExpandableSpace() bool {
	return c&ExpandableSpace != 0
}

// is a word boundary, as defined by UAX#29.
// More specifically, means that this is not a position in the middle
// of a word.  For example, both sides of a punctuation mark are
// considered word boundaries.  This flag is particularly useful when
// selecting text word-by-word.
// This flag implements Unicode's
// <ulink url="http://www.unicode.org/reports/tr29/">Word
// Boundaries</ulink> semantics. (Since: 1.22)
func (c CharAttr) IsWordBoundary() bool {
	return c&WordBoundary != 0
}

func (c *CharAttr) SetLineBreak(is bool) {
	if is {
		*c = *c | LineBreak
	} else {
		*c = *c &^ LineBreak
	}
}

func (c *CharAttr) SetMandatoryBreak(is bool) {
	if is {
		*c = *c | MandatoryBreak
	} else {
		*c = *c &^ MandatoryBreak
	}
}

func (c *CharAttr) SetCharBreak(is bool) {
	if is {
		*c = *c | CharBreak
	} else {
		*c = *c &^ CharBreak
	}
}

func (c *CharAttr) SetWhite(is bool) {
	if is {
		*c = *c | White
	} else {
		*c = *c &^ White
	}
}

func (c *CharAttr) SetCursorPosition(is bool) {
	if is {
		*c = *c | CursorPosition
	} else {
		*c = *c &^ CursorPosition
	}
}

func (c *CharAttr) SetWordStart(is bool) {
	if is {
		*c = *c | WordStart
	} else {
		*c = *c &^ WordStart
	}
}

func (c *CharAttr) SetWordEnd(is bool) {
	if is {
		*c = *c | WordEnd
	} else {
		*c = *c &^ WordEnd
	}
}

func (c *CharAttr) SetSentenceBoundary(is bool) {
	if is {
		*c = *c | SentenceBoundary
	} else {
		*c = *c &^ SentenceBoundary
	}
}

func (c *CharAttr) SetSentenceStart(is bool) {
	if is {
		*c = *c | SentenceStart
	} else {
		*c = *c &^ SentenceStart
	}
}

func (c *CharAttr) SetSentenceEnd(is bool) {
	if is {
		*c = *c | SentenceEnd
	} else {
		*c = *c &^ SentenceEnd
	}
}

func (c *CharAttr) SetBackspaceDeletesCharacter(is bool) {
	if is {
		*c = *c | BackspaceDeletesCharacter
	} else {
		*c = *c &^ BackspaceDeletesCharacter
	}
}

func (c *CharAttr) SetExpandableSpace(is bool) {
	if is {
		*c = *c | ExpandableSpace
	} else {
		*c = *c &^ ExpandableSpace
	}
}

func (c *CharAttr) SetWordBoundary(is bool) {
	if is {
		*c = *c | WordBoundary
	} else {
		*c = *c &^ WordBoundary
	}
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
