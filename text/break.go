package text

import (
	"unicode"
	"unicode/utf8"

	"github.com/benoitkugler/go-weasyprint/text/unicodedata"

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
	r, _ := utf8.DecodeRune(c.data[c.index:])
	return r
}

// increase the position
func (c *charPointer) nextUTF8() {
	_, l := utf8.DecodeRune(c.data[c.index:])
	c.index += l
}

// return `true` if we are at the end of the string
func (c charPointer) end() bool {
	return c.index >= len(c.data)
}

func unicodeCategorie(r rune) *unicode.RangeTable {
	for cat, table := range unicode.Categories {
		if len(cat) == 2 && unicode.Is(table, r) {
			return table
		}
	}
	return nil
}

func unicodeScript(r rune) *unicode.RangeTable {
	for _, table := range unicode.Scripts {
		if unicode.Is(table, r) {
			return table
		}
	}
	return nil
}

func backspaceDeleteCharacter(wc rune) bool {
	return !((wc >= 0x0020 && wc <= 0x02AF) || (wc >= 0x1E00 && wc <= 0x1EFF)) &&
		!(wc >= 0x0400 && wc <= 0x052F) &&
		!((wc >= 0x0370 && wc <= 0x3FF) || (wc >= 0x1F00 && wc <= 0x1FFF)) &&
		!(wc >= 0x3040 && wc <= 0x30FF) &&
		!(wc >= 0xAC00 && wc <= 0xD7A3) &&
		!unicodedata.IsEmojiBaseCharacter(wc)
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
	attrs := make([]PangoLogAttr, len(text_))
	text := charPointer{data: []byte(text_)}
	next := text // share same data but with different adresses
	var (
		prevWc = 0
		nextWc rune

		prevJamo = linebreak.NO_JAMO

		nextBreakType     = linebreak.G_UNICODE_BREAK_XX
		prevBreakType     linebreak.GUnicodeBreakType
		prevPrevBreakType = linebreak.G_UNICODE_BREAK_XX

		prevGbType              = gb_Other
		metExtendedPictographic = false

		prevPrevWbType = wb_Other
		prevWbType     = wb_Other
		prevWbI        = -1

		prevPrevSbType = SB_Other
		prevSbType     = SB_Other
		prevSbI        = -1

		prevLbType = LB_Other

		currentWordType = wordNone
		lastWordLetter  = 0
		baseCharacter   rune

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
			nextBreakType = linebreak.G_UNICODE_BREAK_XX
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

		type_ := unicodeCategorie(wc)
		switch type_ {
		case unicode.Zs, unicode.Zl, unicode.Zp:
			attrs[i].IsWhite = true
		default:
			attrs[i].IsWhite = wc == '\t' || wc == '\n' || wc == '\r' || wc == '\f'
		}

		// Just few spaces have variable width. So explicitly mark them.
		attrs[i].IsExpandableSpace = (0x0020 == wc || 0x00A0 == wc)

		isExtendedPictographic := unicodedata.IsEmojiExtendedPictographic(wc)

		// ---- UAX#29 Grapheme Boundaries ----

		/* Find the GraphemeBreakType of wc */
		gbType := gb_Other
		switch type_ {
		case unicode.Cf:
			if wc == 0x200C {
				gbType = gb_Extend
			} else if wc == 0x200D {
				gbType = gb_ZWJ
			} else if (wc >= 0x600 && wc <= 0x605) ||
				wc == 0x6DD ||
				wc == 0x70F ||
				wc == 0x8E2 ||
				wc == 0xD4E ||
				wc == 0x110BD ||
				(wc >= 0x111C2 && wc <= 0x111C3) {
				gbType = gb_Prepend
			}
			fallthrough
		case unicode.Cc, unicode.Zl, unicode.Zp, unicode.Cs:
			gbType = gb_ControlCRLF
		case nil:
			/* Unassigned default ignorables */
			if (wc >= 0xFFF0 && wc <= 0xFFF8) || (wc >= 0xE0000 && wc <= 0xE0FFF) {
				gbType = gb_ControlCRLF
				break
			}
			fallthrough
		case unicode.Lo:
			if makesHangulSyllable {
				gbType = gb_InHangulSyllable
			}
		case unicode.Lm:
			if wc >= 0xFF9E && wc <= 0xFF9F {
				gbType = gb_Extend /* Other_Grapheme_Extend */
			}
		case unicode.Mc:
			gbType = gb_SpacingMark /* SpacingMark */
			if wc >= 0x0900 {
				if wc == 0x09BE || wc == 0x09D7 ||
					wc == 0x0B3E || wc == 0x0B57 || wc == 0x0BBE || wc == 0x0BD7 ||
					wc == 0x0CC2 || wc == 0x0CD5 || wc == 0x0CD6 ||
					wc == 0x0D3E || wc == 0x0D57 || wc == 0x0DCF || wc == 0x0DDF ||
					wc == 0x1D165 || (wc >= 0x1D16E && wc <= 0x1D172) {
					gbType = gb_Extend /* Other_Grapheme_Extend */
				}
			}
		case unicode.Me, unicode.Mn:
			gbType = gb_Extend /* Grapheme_Extend */
		case unicode.So:
			if wc >= 0x1F1E6 && wc <= 0x1F1FF {
				if prevGbType == gb_RI_Odd {
					gbType = gb_RI_Even
				} else if prevGbType == gb_RI_Even {
					gbType = gb_RI_Odd
				} else {
					gbType = gb_RI_Odd
				}
			}
		case unicode.Sk:
			if wc >= 0x1F3FB && wc <= 0x1F3FF {
				gbType = gb_Extend
			}
		}

		/* Rule GB11 */
		if metExtendedPictographic {
			if gbType == gb_Extend {
				metExtendedPictographic = true
			} else if unicodedata.IsEmojiExtendedPictographic(prevWc) && gbType == gb_ZWJ {
				metExtendedPictographic = true
			} else if prevGbType == gb_Extend && gbType == gb_ZWJ {
				metExtendedPictographic = true
			} else if prevGbType == gb_ZWJ && isExtendedPictographic {
				metExtendedPictographic = true
			} else {
				metExtendedPictographic = false
			}
		}

		/* Grapheme Cluster Boundary Rules */
		isGraphemeBoundary := true /* Rule GB999 */
		/* We apply Rules GB1 && GB2 at the end of the function */
		if wc == '\n' && prevWc == '\r' {
			isGraphemeBoundary = false /* Rule GB3 */
		} else if prevGbType == gb_ControlCRLF || gbType == gb_ControlCRLF {
			isGraphemeBoundary = true /* Rules GB4 && GB5 */
		} else if gbType == gb_InHangulSyllable {
			isGraphemeBoundary = false /* Rules GB6, GB7, GB8 */
		} else if gbType == gb_Extend {
			isGraphemeBoundary = false /* Rule GB9 */
		} else if gbType == gb_ZWJ {
			isGraphemeBoundary = false /* Rule GB9 */
		} else if gbType == gb_SpacingMark {
			isGraphemeBoundary = false /* Rule GB9a */
		} else if prevGbType == gb_Prepend {
			isGraphemeBoundary = false /* Rule GB9b */
		} else if isExtendedPictographic { /* Rule GB11 */
			if prevGbType == gb_ZWJ && metExtendedPictographic {
				isGraphemeBoundary = false
			}
		} else if prevGbType == gb_RI_Odd && gbType == gb_RI_Even {
			isGraphemeBoundary = false /* Rule GB12 && GB13 */
		}

		if isExtendedPictographic {
			metExtendedPictographic = true
		}

		attrs[i].IsCursorPosition = isGraphemeBoundary
		/* If this is a grapheme boundary, we have to decide if backspace
		 * deletes a character or the whole grapheme cluster */
		if isGraphemeBoundary {
			attrs[i].BackspaceDeletesCharacter = backspaceDeleteCharacter(baseCharacter)

			/* Dependent Vowels for Indic language */
			if unicodedata.IsVirama(prevWc) || unicodedata.IsVowel_Dependent(prevWc) {
				attrs[i].BackspaceDeletesCharacter = true
			}
		} else {
			attrs[i].BackspaceDeletesCharacter = false
		}

		prevGbType = gbType

		/* ---- UAX#29 Word Boundaries ---- */

		isWordBoundary := false
		if isGraphemeBoundary || (wc >= 0x1F1E6 && wc <= 0x1F1FF) { /* Rules WB3 and WB4 */
			script := unicodeScript(wc)

			/* Find the WordBreakType of wc */
			wbType := wb_Other

			if script == unicode.Katakana {
				wbType = wb_Katakana
			}

			if script == unicode.Hebrew && type_ == unicode.Lo {
				wbType = wb_Hebrew_Letter
			}

			if wbType == wb_Other {
				switch wc >> 8 {
				case 0x30:
					if wc == 0x3031 || wc == 0x3032 || wc == 0x3033 || wc == 0x3034 || wc == 0x3035 ||
						wc == 0x309b || wc == 0x309c || wc == 0x30a0 || wc == 0x30fc {
						wbType = wb_Katakana /* Katakana exceptions */
					}
				case 0xFF:
					if wc == 0xFF70 {
						wbType = wb_Katakana /* Katakana exceptions */
					} else if wc >= 0xFF9E && wc <= 0xFF9F {
						wbType = wb_ExtendFormat /* Other_Grapheme_Extend */
					}
				case 0x05:
					if wc == 0x05F3 {
						wbType = wb_ALetter /* ALetter exceptions */
					}
				}
			}

			if wbType == wb_Other {
				switch breakType {
				case linebreak.G_UNICODE_BREAK_NUMERIC:
					if wc != 0x066C {
						wbType = wb_Numeric /* Numeric */
					}
				case linebreak.G_UNICODE_BREAK_INFIX_SEPARATOR:
					if wc != 0x003A && wc != 0xFE13 && wc != 0x002E {
						wbType = wb_MidNum /* MidNum */
					}
				}
			}

			if wbType == wb_Other {
				switch type_ {
				case unicode.Cc:
					if wc != 0x000D && wc != 0x000A && wc != 0x000B && wc != 0x000C && wc != 0x0085 {
						break
					}
					fallthrough
				case unicode.Zl, unicode.Zp:
					wbType = wb_NewlineCRLF /* CR, LF, Newline */
				case unicode.Cf, unicode.Mc, unicode.Me, unicode.Mn:
					wbType = wb_ExtendFormat /* Extend, Format */
				case unicode.Pc:
					wbType = wb_ExtendNumLet /* ExtendNumLet */
				case unicode.Pf, unicode.Pi:
					if wc == 0x2018 || wc == 0x2019 {
						wbType = wb_MidNumLet /* MidNumLet */
					}
				case unicode.Po:
					if wc == 0x0027 || wc == 0x002e || wc == 0x2024 ||
						wc == 0xfe52 || wc == 0xff07 || wc == 0xff0e {
						wbType = wb_MidNumLet /* MidNumLet */
					} else if wc == 0x00b7 || wc == 0x05f4 || wc == 0x2027 || wc == 0x003a || wc == 0x0387 ||
						wc == 0xfe13 || wc == 0xfe55 || wc == 0xff1a {
						wbType = wb_MidLetter /* wb_MidLetter */
					} else if wc == 0x066c ||
						wc == 0xfe50 || wc == 0xfe54 || wc == 0xff0c || wc == 0xff1b {
						wbType = wb_MidNum /* MidNum */
					}
				case unicode.So:
					if wc >= 0x24B6 && wc <= 0x24E9 { /* Other_Alphabetic */
						goto Alphabetic
					}
					if wc >= 0x1F1E6 && wc <= 0x1F1FF {
						if prevWbType == wb_RI_Odd {
							wbType = wb_RI_Even
						} else if prevWbType == wb_RI_Even {
							wbType = wb_RI_Odd
						} else {
							wbType = wb_RI_Odd
						}
					}

				case unicode.Lo, unicode.Nl:
					if wc == 0x3006 || wc == 0x3007 ||
						(wc >= 0x3021 && wc <= 0x3029) ||
						(wc >= 0x3038 && wc <= 0x303A) ||
						(wc >= 0x3400 && wc <= 0x4DB5) ||
						(wc >= 0x4E00 && wc <= 0x9FC3) ||
						(wc >= 0xF900 && wc <= 0xFA2D) ||
						(wc >= 0xFA30 && wc <= 0xFA6A) ||
						(wc >= 0xFA70 && wc <= 0xFAD9) ||
						(wc >= 0x20000 && wc <= 0x2A6D6) ||
						(wc >= 0x2F800 && wc <= 0x2FA1D) {
						break /* ALetter exceptions: Ideographic */
					}
					goto Alphabetic
				case unicode.Ll, unicode.Lm, unicode.Lu, unicode.Lt:
					goto Alphabetic
				}

			Alphabetic:
				if breakType != linebreak.G_UNICODE_BREAK_SA && script != unicode.Hiragana {
					wbType = wb_ALetter /* ALetter */
				}

				if wbType == wb_Other {
					if type_ == unicode.Zs && breakType != linebreak.G_UNICODE_BREAK_NON_BREAKING_GLUE {
						wbType = wb_WSegSpace
					}
				}

				/* Word Cluster Boundary Rules */

				/* We apply Rules WB1 and WB2 at the end of the function */

				if prevWbType == wb_NewlineCRLF && prevWbI+1 == i {
					/* The extra check for prevWbI is to correctly handle sequences like
					 * Newline ÷ Extend × Extend
					 * since we have not skipped ExtendFormat yet.
					 */
					isWordBoundary = true /* Rule WB3a */
				} else if wbType == wb_NewlineCRLF {
					isWordBoundary = true /* Rule WB3b */
				} else if prevWc == 0x200D && isExtendedPictographic {
					isWordBoundary = false /* Rule WB3c */
				} else if prevWbType == wb_WSegSpace &&
					wbType == wb_WSegSpace && prevWbI+1 == i {
					isWordBoundary = false /* Rule WB3d */
				} else if wbType == wb_ExtendFormat {
					isWordBoundary = false /* Rules WB4? */
				} else if (prevWbType == wb_ALetter ||
					prevWbType == wb_Hebrew_Letter ||
					prevWbType == wb_Numeric) &&
					(wbType == wb_ALetter ||
						wbType == wb_Hebrew_Letter ||
						wbType == wb_Numeric) {
					isWordBoundary = false /* Rules WB5, WB8, WB9, WB10 */
				} else if prevWbType == wb_Katakana && wbType == wb_Katakana {
					isWordBoundary = false /* Rule WB13 */
				} else if (prevWbType == wb_ALetter ||
					prevWbType == wb_Hebrew_Letter ||
					prevWbType == wb_Numeric ||
					prevWbType == wb_Katakana ||
					prevWbType == wb_ExtendNumLet) &&
					wbType == wb_ExtendNumLet {
					isWordBoundary = false /* Rule WB13a */
				} else if prevWbType == wb_ExtendNumLet &&
					(wbType == wb_ALetter ||
						wbType == wb_Hebrew_Letter ||
						wbType == wb_Numeric ||
						wbType == wb_Katakana) {
					isWordBoundary = false /* Rule WB13b */
				} else if ((prevPrevWbType == wb_ALetter ||
					prevPrevWbType == wb_Hebrew_Letter) &&
					(wbType == wb_ALetter ||
						wbType == wb_Hebrew_Letter)) &&
					(prevWbType == wb_MidLetter ||
						prevWbType == wb_MidNumLet ||
						prevWc == 0x0027) {
					attrs[prevWbI].IsWordBoundary = false /* Rule WB6 */
					isWordBoundary = false                /* Rule WB7 */
				} else if prevWbType == wb_Hebrew_Letter && wc == 0x0027 {
					isWordBoundary = false /* Rule WB7a */
				} else if prevPrevWbType == wb_Hebrew_Letter && prevWc == 0x0022 &&
					wbType == wb_Hebrew_Letter {
					attrs[prevWbI].IsWordBoundary = false /* Rule WB7b */
					isWordBoundary = false                /* Rule WB7c */
				} else if (prevPrevWbType == wb_Numeric && wbType == wb_Numeric) &&
					(prevWbType == wb_MidNum || prevWbType == wb_MidNumLet ||
						prevWc == 0x0027) {
					isWordBoundary = false                /* Rule WB11 */
					attrs[prevWbI].IsWordBoundary = false /* Rule WB12 */
				} else if prevWbType == wb_RI_Odd && wbType == wb_RI_Even {
					isWordBoundary = false /* Rule WB15 and WB16 */
				} else {
					isWordBoundary = true /* Rule WB999 */
				}

				if wbType != wb_ExtendFormat {
					prevPrevWbType = prevWbType
					prevWbType = wbType
					prevWbI = i
				}
			}

			attrs[i].IsWordBoundary = isWordBoundary
		}

	}
}
