package text

import (
	"unicode/utf8"

	"github.com/benoitkugler/go-weasyprint/text/linebreak"
)

const PARAGRAPH_SEPARATOR rune = 0x2029

// This is the default break algorithm. It applies Unicode
// rules without language-specific tailoring.
//
// See pango_tailor_break() for language-specific breaks.
func pangoDefaultBreak(text string) []PangoLogAttr {
	// The rationale for all this is in section 5.15 of the Unicode 3.0 book,
	// the line breaking stuff is also in TR14 on unicode.org
	// This is a default break implementation that should work for nearly all
	// languages. Language engines can override it optionally.

	var (
		next *gchar
		i    int

		prev_wc = 0
		next_wc rune

		prev_jamo = NO_JAMO

		next_break_type      = linebreak.G_UNICODE_BREAK_XX
		prev_break_type      linebreak.GUnicodeBreakType
		prev_prev_break_type = linebreak.G_UNICODE_BREAK_XX

		prev_GB_type              = GB_Other
		met_Extended_Pictographic = false

		prev_prev_WB_type = WB_Other
		prev_WB_type      = WB_Other
		prev_WB_i         = -1

		prev_prev_SB_type = SB_Other
		prev_SB_type      = SB_Other
		prev_SB_i         = -1

		prev_LB_type = LB_Other

		current_word_type                = WordNone
		last_word_letter, base_character = 0, 0

		last_sentence_start, last_non_space = -1, -1

		almost_done, done bool
	)
	if len(text) == 0 {
		next_wc = PARAGRAPH_SEPARATOR
		almost_done = true
	} else {
		next_wc, _ = utf8.DecodeRuneInString(text)
	}

	next_break_type = linebreak.ResolveClass(next_wc)

}
