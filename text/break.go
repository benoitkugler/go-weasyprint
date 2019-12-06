package text

import (
	"unicode"
	"unicode/utf8"

	"github.com/benoitkugler/go-weasyprint/text/linebreak"
)

const PARAGRAPH_SEPARATOR rune = 0x2029

// replace a *char from c
type charPointer struct {
	data  []byte
	index int
}

// return rune at current position
// does not change it
func (c charPointer) getUTF8Char() rune {
	r, _ := utf8.DecodeRune(c.data[index:])
	return r
}

// increase the position
func (c *charPointer) nextUTF8() {
	_, l := utf8.DecodeRune(c.data[index:])
	c.index += l
}

// return `true` if we are at the end of the string
func (c charPointer) end() bool {
	return c.index >= len(c.data)
}

// This is the default break algorithm. It applies Unicode
// rules without language-specific tailoring.
//
// See pango_tailor_break() for language-specific breaks.
func pangoDefaultBreak(text_ string) []PangoLogAttr {
	// The rationale for all this is in section 5.15 of the Unicode 3.0 book,
	// the line breaking stuff is also in TR14 on unicode.org
	// This is a default break implementation that should work for nearly all
	// languages. Language engines can override it optionally.
	text := charPointer{data: []byte(text_)}
	next := text // share same data but with different adresses
	var (
		prevWc = 0
		nextWc rune

		prevJamo = NO_JAMO

		nextBreakType     = linebreak.G_UNICODE_BREAK_XX
		prevBreakType     linebreak.GUnicodeBreakType
		prevPrevBreakType = linebreak.G_UNICODE_BREAK_XX

		prevGbType              = GB_Other
		metExtendedPictographic = false

		prevPrevWbType = WB_Other
		prevWbType     = WB_Other
		prevWbI        = -1

		prevPrevSbType = SB_Other
		prevSbType     = SB_Other
		prevSbI        = -1

		prevLbType = LB_Other

		currentWordType               = WordNone
		lastWordLetter, baseCharacter = 0, 0

		lastSentenceStart, lastNonSpace = -1, -1

		almostDone, done bool
	)

	if len(text.data) == 0 {
		nextWc = PARAGRAPH_SEPARATOR
		almostDone = true
	} else {
		nextWc = text.getUTF8Char()
	}

	nextBreakType = linebreak.ResolveClass(nextWc)
	for i := 0; !done; i++ {
		var makesHangulSyllable bool
		wc := nextWc
		breakType := nextBreakType

		if almostDone {
			// If we have already reached the end of `text`, gUtf8NextChar()
			// may not increment next

			nextWc = 0
			nextBreakType = G_UNICODE_BREAK_XX
			done = true
		} else {
			next.nextUTF8()
			if next.end() {
				// This is how we fill in the last element (end position) of the
				// attr array - assume there"s a paragraph separators off the end
				// of @text.
				nextWc = PARAGRAPH_SEPARATOR
				almostDone = true
			} else {
				nextWc = next.getUTF8Char()
			}

			nextBreakType = linebreak.ResolveClass(nextWc)
		}

		// type = gUnicharType(wc);
		jamo := linebreak.Jamo(breakType)

		/* Determine wheter this forms a Hangul syllable with prev. */
		if jamo == linebreak.NO_JAMO {
			makesHangulSyllable = false
		} else {
			prevEnd := HangulJamoProps[prevJamo].end
			thisStart := HangulJamoProps[jamo].start

			/* See comments before ISJAMO */
			makesHangulSyllable = (prevEnd == thisStart) || (prevEnd+1 == thisStart)
		}

		switch {
		case unicode.In(wc, unicode.Zs, unicode.Zl, unicode.Zp):
			attrs[i].isWhite = true
		}

	}
}
