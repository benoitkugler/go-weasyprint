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

func isOtherTerm(sbType sentenceBreakType) bool {
	/* not in (OLetter | Upper | Lower | ParaSep | SATerm) */
	return !(sbType == sb_OLetter ||
		sbType == sb_Upper || sbType == sb_Lower ||
		sbType == sb_ParaSep ||
		sbType == sb_ATerm || sbType == sb_STerm ||
		sbType == sb_ATerm_Close_Sp ||
		sbType == sb_STerm_Close_Sp)
}

/* Types of Japanese characters */
func _JAPANESE(wc rune) bool { return wc >= 0x2F00 && wc <= 0x30FF }
func _KANJI(wc rune) bool    { return wc >= 0x2F00 && wc <= 0x2FDF }
func _HIRAGANA(wc rune) bool { return wc >= 0x3040 && wc <= 0x309F }
func _KATAKANA(wc rune) bool { return wc >= 0x30A0 && wc <= 0x30FF }

// This is the default break algorithm. It applies Unicode
// rules without language-specific tailoring.
//
// See pango_tailor_break() for language-specific breaks.
func pangoDefaultBreak(text_ string) []PangoLogAttr {
	// The rationale for all this is in section 5.15 of the Unicode 3.0 book,
	// the line breaking stuff is also in TR14 on unicode.org
	// This is a default break implementation that should work for nearly all
	// languages. Language engines can override it optionally.
	attrs := make([]PangoLogAttr, len(text_)+1)
	text := charPointer{data: []byte(text_)}
	next := text // share same data but with different adresses
	var (
		prevWc, nextWc rune

		prevJamo = linebreak.NO_JAMO

		nextBreakType     = linebreak.G_UNICODE_BREAK_XX
		prevBreakType     linebreak.GUnicodeBreakType
		prevPrevBreakType = linebreak.G_UNICODE_BREAK_XX

		prevGbType              = gb_Other
		metExtendedPictographic = false

		prevPrevWbType = wb_Other
		prevWbType     = wb_Other
		prevWbI        = -1

		prevPrevSbType = sb_Other
		prevSbType     = sb_Other
		prevSbI        = -1

		prevLbType = lb_Other

		currentWordType               = wordNone
		lastWordLetter, baseCharacter rune

		lastSentenceStart, lastNonSpace = -1, -1

		almostDone, done bool
		i                int
	)

	if len(text.data) == 0 {
		nextWc = PARAGRAPH_SEPARATOR
		almostDone = true
	} else {
		nextWc = text.getUTF8Char()
	}

	nextBreakType = linebreak.ResolveClass(nextWc)
	for i = 0; !done; i++ {
		var (
			makesHangulSyllable                                    bool
			isWordBoundary, isGraphemeBoundary, isSentenceBoundary bool
			breakOp                                                breakOpportunity
			rowBreakType                                           linebreak.GUnicodeBreakType
		)

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

		type_ := unicodeCategorie(wc)
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
		{
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
		}
		/* ---- UAX#29 Word Boundaries ---- */
		{
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
					case linebreak.G_UNICODE_BREAK_NU:
						if wc != 0x066C {
							wbType = wb_Numeric /* Numeric */
						}
					case linebreak.G_UNICODE_BREAK_IS:
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
				}

			Alphabetic:
				if breakType != linebreak.G_UNICODE_BREAK_SA && script != unicode.Hiragana {
					wbType = wb_ALetter /* ALetter */
				}

				if wbType == wb_Other {
					if type_ == unicode.Zs && breakType != linebreak.G_UNICODE_BREAK_GL {
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

		/* ---- UAX#29 Sentence Boundaries ---- */
		{
			isSentenceBoundary := false
			if isWordBoundary || wc == '\r' || wc == '\n' { /* Rules SB3 and SB5 */
				/* Find the SentenceBreakType of wc */
				sbType := sb_Other

				if breakType == linebreak.G_UNICODE_BREAK_NU {
					sbType = sb_Numeric /* Numeric */
				}

				if sbType == sb_Other {
					switch type_ {
					case unicode.Cc:
						if wc == '\r' || wc == '\n' {
							sbType = sb_ParaSep
						} else if wc == 0x0009 || wc == 0x000B || wc == 0x000C {
							sbType = sb_Sp
						} else if wc == 0x0085 {
							sbType = sb_ParaSep
						}
					case unicode.Zs:
						if wc == 0x0020 || wc == 0x00A0 || wc == 0x1680 ||
							(wc >= 0x2000 && wc <= 0x200A) ||
							wc == 0x202F || wc == 0x205F || wc == 0x3000 {
							sbType = sb_Sp
						}

					case unicode.Zl, unicode.Zp:
						sbType = sb_ParaSep
					case unicode.Cf, unicode.Mc, unicode.Me, unicode.Mn:
						sbType = sb_ExtendFormat /* Extend, Format */
					case unicode.Lm:
						if wc >= 0xFF9E && wc <= 0xFF9F {
							sbType = sb_ExtendFormat /* Other_Grapheme_Extend */
						}
					case unicode.Lt:
						sbType = sb_Upper
					case unicode.Pd:
						if wc == 0x002D ||
							(wc >= 0x2013 && wc <= 0x2014) ||
							(wc >= 0xFE31 && wc <= 0xFE32) ||
							wc == 0xFE58 ||
							wc == 0xFE63 ||
							wc == 0xFF0D {
							sbType = sb_SContinue
						}
					case unicode.Po:
						if wc == 0x05F3 {
							sbType = sb_OLetter
						} else if wc == 0x002E || wc == 0x2024 ||
							wc == 0xFE52 || wc == 0xFF0E {
							sbType = sb_ATerm
						}
						if wc == 0x002C ||
							wc == 0x003A ||
							wc == 0x055D ||
							(wc >= 0x060C && wc <= 0x060D) ||
							wc == 0x07F8 ||
							wc == 0x1802 ||
							wc == 0x1808 ||
							wc == 0x3001 ||
							(wc >= 0xFE10 && wc <= 0xFE11) ||
							wc == 0xFE13 ||
							(wc >= 0xFE50 && wc <= 0xFE51) ||
							wc == 0xFE55 ||
							wc == 0xFF0C ||
							wc == 0xFF1A ||
							wc == 0xFF64 {
							sbType = sb_SContinue
						}
						if unicode.Is(unicode.STerm, wc) {
							sbType = sb_STerm
						}
					}
				}

				if sbType == sb_Other {
					switch type_ {
					case unicode.Ll:
						sbType = sb_Lower
					case unicode.Lu:
						sbType = sb_Upper
					case unicode.Lt, unicode.Lm, unicode.Lo:
						sbType = sb_OLetter
					}

					if type_ == unicode.Pe || type_ == unicode.Ps || breakType == linebreak.G_UNICODE_BREAK_QU {
						sbType = sb_Close
					}
				}

				/* Sentence Boundary Rules */

				/* We apply Rules SB1 and SB2 at the end of the function */
				switch {
				case wc == '\n' && prevWc == '\r':
					isSentenceBoundary = false /* Rule SB3 */
				case prevSbType == sb_ParaSep && prevSbI+1 == i:
					/* The extra check for prevSbI is to correctly handle sequences like
					 * ParaSep ÷ Extend × Extend
					 * since we have not skipped ExtendFormat yet.
					 */

					isSentenceBoundary = true /* Rule SB4 */

				case sbType == sb_ExtendFormat:
					isSentenceBoundary = false /* Rule SB5? */
				case prevSbType == sb_ATerm && sbType == sb_Numeric:
					isSentenceBoundary = false /* Rule SB6 */
				case (prevPrevSbType == sb_Upper ||
					prevPrevSbType == sb_Lower) &&
					prevSbType == sb_ATerm &&
					sbType == sb_Upper:
					isSentenceBoundary = false /* Rule SB7 */
				case prevSbType == sb_ATerm && sbType == sb_Close:
					sbType = sb_ATerm
				case prevSbType == sb_STerm && sbType == sb_Close:
					sbType = sb_STerm
				case prevSbType == sb_ATerm && sbType == sb_Sp:
					sbType = sb_ATerm_Close_Sp
				case prevSbType == sb_STerm && sbType == sb_Sp:
					sbType = sb_STerm_Close_Sp
				/* Rule SB8 */
				case (prevSbType == sb_ATerm ||
					prevSbType == sb_ATerm_Close_Sp) &&
					sbType == sb_Lower:
					isSentenceBoundary = false
				case (prevPrevSbType == sb_ATerm ||
					prevPrevSbType == sb_ATerm_Close_Sp) &&
					isOtherTerm(prevSbType) &&
					sbType == sb_Lower:
					attrs[prevSbI].IsSentenceBoundary = false
				case (prevSbType == sb_ATerm ||
					prevSbType == sb_ATerm_Close_Sp ||
					prevSbType == sb_STerm ||
					prevSbType == sb_STerm_Close_Sp) &&
					(sbType == sb_SContinue ||
						sbType == sb_ATerm || sbType == sb_STerm):
					isSentenceBoundary = false /* Rule SB8a */
				case (prevSbType == sb_ATerm ||
					prevSbType == sb_STerm) &&
					(sbType == sb_Close || sbType == sb_Sp ||
						sbType == sb_ParaSep):
					isSentenceBoundary = false /* Rule SB9 */
				case (prevSbType == sb_ATerm ||
					prevSbType == sb_ATerm_Close_Sp ||
					prevSbType == sb_STerm ||
					prevSbType == sb_STerm_Close_Sp) &&
					(sbType == sb_Sp || sbType == sb_ParaSep):
					isSentenceBoundary = false /* Rule SB10 */
				case (prevSbType == sb_ATerm ||
					prevSbType == sb_ATerm_Close_Sp ||
					prevSbType == sb_STerm ||
					prevSbType == sb_STerm_Close_Sp) &&
					sbType != sb_ParaSep:
					isSentenceBoundary = true /* Rule SB11 */
				default:
					isSentenceBoundary = false /* Rule SB998 */
				}

				if sbType != sb_ExtendFormat &&
					!((prevPrevSbType == sb_ATerm ||
						prevPrevSbType == sb_ATerm_Close_Sp) &&
						isOtherTerm(prevSbType) &&
						isOtherTerm(sbType)) {
					prevPrevSbType = prevSbType
					prevSbType = sbType
					prevSbI = i
				}
			}

			if i == 0 || done {
				isSentenceBoundary = true /* Rules SB1 and SB2 */
			}
			attrs[i].IsSentenceBoundary = isSentenceBoundary
		}
		/* ---- Line breaking ---- */
		breakOp = break_ALREADY_HANDLED

		rowBreakType = prevBreakType
		if prevBreakType == linebreak.G_UNICODE_BREAK_SP {
			rowBreakType = prevPrevBreakType
		}

		attrs[i].IsCharBreak = false
		attrs[i].IsLineBreak = false
		attrs[i].IsMandatoryBreak = false

		/* Rule LB1:
		assign a line breaking class to each code point of the input. */
		switch breakType {
		case linebreak.G_UNICODE_BREAK_AI, linebreak.G_UNICODE_BREAK_SG, linebreak.G_UNICODE_BREAK_XX:
			breakType = linebreak.G_UNICODE_BREAK_AL
		case linebreak.G_UNICODE_BREAK_SA:
			if type_ == unicode.Mn || type_ == unicode.Mc {
				breakType = linebreak.G_UNICODE_BREAK_CM
			} else {
				breakType = linebreak.G_UNICODE_BREAK_AL
			}
		case linebreak.G_UNICODE_BREAK_CJ:
			breakType = linebreak.G_UNICODE_BREAK_NS
		}

		/* If it's not a grapheme boundary, it's not a line break either */
		if attrs[i].IsCursorPosition ||
			// breakType == linebreak.G_UNICODE_BREAK_EM ||
			// breakType == linebreak.G_UNICODE_BREAK_ZWJ ||
			breakType == linebreak.G_UNICODE_BREAK_CM ||
			breakType == linebreak.G_UNICODE_BREAK_JL ||
			breakType == linebreak.G_UNICODE_BREAK_JV ||
			breakType == linebreak.G_UNICODE_BREAK_JT ||
			breakType == linebreak.G_UNICODE_BREAK_H2 ||
			breakType == linebreak.G_UNICODE_BREAK_H3 ||
			breakType == linebreak.G_UNICODE_BREAK_RI {

			/* Find the LineBreakType of wc */
			lbType := lb_Other

			if breakType == linebreak.G_UNICODE_BREAK_NU {
				lbType = lb_Numeric
			}
			if breakType == linebreak.G_UNICODE_BREAK_SY ||
				breakType == linebreak.G_UNICODE_BREAK_IS {
				if !(prevLbType == lb_Numeric) {
					lbType = lb_Other
				}
			}

			if breakType == linebreak.G_UNICODE_BREAK_CL ||
				breakType == linebreak.G_UNICODE_BREAK_CP {
				if prevLbType == lb_Numeric {
					lbType = lb_Numeric_Close
				} else {
					lbType = lb_Other
				}
			}

			if breakType == linebreak.G_UNICODE_BREAK_RI {
				if prevLbType == lb_RI_Odd {
					lbType = lb_RI_Even
				} else if prevLbType == lb_RI_Even {
					lbType = lb_RI_Odd
				} else {
					lbType = lb_RI_Odd
				}
			}

			attrs[i].IsLineBreak = true /* Rule LB31 */
			/* Unicode doesn't specify char wrap;
			   we wrap around all chars currently. */
			if attrs[i].IsCursorPosition {
				attrs[i].IsCharBreak = true
			}
			/* Make any necessary replacements first */
			if rowBreakType == linebreak.G_UNICODE_BREAK_XX {
				rowBreakType = linebreak.G_UNICODE_BREAK_AL
			}
			/* add the line break rules in reverse order to override
			   the lower priority rules. */

			/* Rule LB30 */
			if (prevBreakType == linebreak.G_UNICODE_BREAK_AL ||
				prevBreakType == linebreak.G_UNICODE_BREAK_HL ||
				prevBreakType == linebreak.G_UNICODE_BREAK_NU) &&
				breakType == linebreak.G_UNICODE_BREAK_OP {
				breakOp = break_PROHIBITED
			}
			if prevBreakType == linebreak.G_UNICODE_BREAK_CP &&
				(breakType == linebreak.G_UNICODE_BREAK_AL ||
					breakType == linebreak.G_UNICODE_BREAK_HL ||
					breakType == linebreak.G_UNICODE_BREAK_NU) {
				breakOp = break_PROHIBITED
			}
			/* Rule LB30a */
			if prevLbType == lb_RI_Odd && lbType == lb_RI_Even {
				breakOp = break_PROHIBITED
			}
			/* Rule LB30b */
			if prevBreakType == linebreak.G_UNICODE_BREAK_EB &&
				breakType == linebreak.G_UNICODE_BREAK_EM {
				breakOp = break_PROHIBITED
			}
			/* Rule LB29 */
			if prevBreakType == linebreak.G_UNICODE_BREAK_IS &&
				(breakType == linebreak.G_UNICODE_BREAK_AL ||
					breakType == linebreak.G_UNICODE_BREAK_HL) {
				breakOp = break_PROHIBITED
			}
			/* Rule LB28 */
			if (prevBreakType == linebreak.G_UNICODE_BREAK_AL ||
				prevBreakType == linebreak.G_UNICODE_BREAK_HL) &&
				(breakType == linebreak.G_UNICODE_BREAK_AL ||
					breakType == linebreak.G_UNICODE_BREAK_HL) {
				breakOp = break_PROHIBITED
			}
			/* Rule LB27 */
			if (prevBreakType == linebreak.G_UNICODE_BREAK_JL ||
				prevBreakType == linebreak.G_UNICODE_BREAK_JV ||
				prevBreakType == linebreak.G_UNICODE_BREAK_JT ||
				prevBreakType == linebreak.G_UNICODE_BREAK_H2 ||
				prevBreakType == linebreak.G_UNICODE_BREAK_H3) &&
				(breakType == linebreak.G_UNICODE_BREAK_IN ||
					breakType == linebreak.G_UNICODE_BREAK_PO) {
				breakOp = break_PROHIBITED
			}
			if prevBreakType == linebreak.G_UNICODE_BREAK_PR &&
				(breakType == linebreak.G_UNICODE_BREAK_JL ||
					breakType == linebreak.G_UNICODE_BREAK_JV ||
					breakType == linebreak.G_UNICODE_BREAK_JT ||
					breakType == linebreak.G_UNICODE_BREAK_H2 ||
					breakType == linebreak.G_UNICODE_BREAK_H3) {
				breakOp = break_PROHIBITED
			}
			/* Rule LB26 */
			if prevBreakType == linebreak.G_UNICODE_BREAK_JL &&
				(breakType == linebreak.G_UNICODE_BREAK_JL ||
					breakType == linebreak.G_UNICODE_BREAK_JV ||
					breakType == linebreak.G_UNICODE_BREAK_H2 ||
					breakType == linebreak.G_UNICODE_BREAK_H3) {
				breakOp = break_PROHIBITED
			}
			if (prevBreakType == linebreak.G_UNICODE_BREAK_JV ||
				prevBreakType == linebreak.G_UNICODE_BREAK_H2) &&
				(breakType == linebreak.G_UNICODE_BREAK_JV ||
					breakType == linebreak.G_UNICODE_BREAK_JT) {
				breakOp = break_PROHIBITED
			}
			if (prevBreakType == linebreak.G_UNICODE_BREAK_JT ||
				prevBreakType == linebreak.G_UNICODE_BREAK_H3) &&
				breakType == linebreak.G_UNICODE_BREAK_JT {
				breakOp = break_PROHIBITED
			}
			/* Rule LB25 with Example 7 of Customization */
			if (prevBreakType == linebreak.G_UNICODE_BREAK_PR ||
				prevBreakType == linebreak.G_UNICODE_BREAK_PO) &&
				breakType == linebreak.G_UNICODE_BREAK_NU {
				breakOp = break_PROHIBITED
			}
			if (prevBreakType == linebreak.G_UNICODE_BREAK_PR ||
				prevBreakType == linebreak.G_UNICODE_BREAK_PO) &&
				(breakType == linebreak.G_UNICODE_BREAK_OP ||
					breakType == linebreak.G_UNICODE_BREAK_HY) &&
				nextBreakType == linebreak.G_UNICODE_BREAK_NU {
				breakOp = break_PROHIBITED
			}
			if (prevBreakType == linebreak.G_UNICODE_BREAK_OP ||
				prevBreakType == linebreak.G_UNICODE_BREAK_HY) &&
				breakType == linebreak.G_UNICODE_BREAK_NU {
				breakOp = break_PROHIBITED
			}
			if prevBreakType == linebreak.G_UNICODE_BREAK_NU &&
				(breakType == linebreak.G_UNICODE_BREAK_NU ||
					breakType == linebreak.G_UNICODE_BREAK_SY ||
					breakType == linebreak.G_UNICODE_BREAK_IS) {
				breakOp = break_PROHIBITED
			}
			if prevLbType == lb_Numeric &&
				(breakType == linebreak.G_UNICODE_BREAK_NU ||
					breakType == linebreak.G_UNICODE_BREAK_SY ||
					breakType == linebreak.G_UNICODE_BREAK_IS ||
					breakType == linebreak.G_UNICODE_BREAK_CL ||
					breakType == linebreak.G_UNICODE_BREAK_CP) {
				breakOp = break_PROHIBITED
			}
			if (prevLbType == lb_Numeric ||
				prevLbType == lb_Numeric_Close) &&
				(breakType == linebreak.G_UNICODE_BREAK_PO ||
					breakType == linebreak.G_UNICODE_BREAK_PR) {
				breakOp = break_PROHIBITED
			}
			/* Rule LB24 */
			if (prevBreakType == linebreak.G_UNICODE_BREAK_PR ||
				prevBreakType == linebreak.G_UNICODE_BREAK_PO) &&
				(breakType == linebreak.G_UNICODE_BREAK_AL ||
					breakType == linebreak.G_UNICODE_BREAK_HL) {
				breakOp = break_PROHIBITED
			}
			if (prevBreakType == linebreak.G_UNICODE_BREAK_AL ||
				prevBreakType == linebreak.G_UNICODE_BREAK_HL) &&
				(breakType == linebreak.G_UNICODE_BREAK_PR ||
					breakType == linebreak.G_UNICODE_BREAK_PO) {
				breakOp = break_PROHIBITED
			}
			/* Rule LB23 */
			if (prevBreakType == linebreak.G_UNICODE_BREAK_AL ||
				prevBreakType == linebreak.G_UNICODE_BREAK_HL) &&
				breakType == linebreak.G_UNICODE_BREAK_NU {
				breakOp = break_PROHIBITED
			}
			if prevBreakType == linebreak.G_UNICODE_BREAK_NU &&
				(breakType == linebreak.G_UNICODE_BREAK_AL ||
					breakType == linebreak.G_UNICODE_BREAK_HL) {
				breakOp = break_PROHIBITED
			}
			/* Rule LB23a */
			if prevBreakType == linebreak.G_UNICODE_BREAK_PR &&
				(breakType == linebreak.G_UNICODE_BREAK_ID ||
					breakType == linebreak.G_UNICODE_BREAK_EB ||
					breakType == linebreak.G_UNICODE_BREAK_EM) {
				breakOp = break_PROHIBITED
			}
			if (prevBreakType == linebreak.G_UNICODE_BREAK_ID ||
				prevBreakType == linebreak.G_UNICODE_BREAK_EB ||
				prevBreakType == linebreak.G_UNICODE_BREAK_EM) &&
				breakType == linebreak.G_UNICODE_BREAK_PO {
				breakOp = break_PROHIBITED
			}

			/* Rule LB22 */
			if breakType == linebreak.G_UNICODE_BREAK_IN {
				if prevBreakType == linebreak.G_UNICODE_BREAK_AL ||
					prevBreakType == linebreak.G_UNICODE_BREAK_HL {
					breakOp = break_PROHIBITED
				}
				if prevBreakType == linebreak.G_UNICODE_BREAK_EX {
					breakOp = break_PROHIBITED
				}
				if prevBreakType == linebreak.G_UNICODE_BREAK_ID ||
					prevBreakType == linebreak.G_UNICODE_BREAK_EB ||
					prevBreakType == linebreak.G_UNICODE_BREAK_EM {
					breakOp = break_PROHIBITED
				}
				if prevBreakType == linebreak.G_UNICODE_BREAK_IN {
					breakOp = break_PROHIBITED
				}
				if prevBreakType == linebreak.G_UNICODE_BREAK_NU {
					breakOp = break_PROHIBITED
				}
			}

			if breakType == linebreak.G_UNICODE_BREAK_BA ||
				breakType == linebreak.G_UNICODE_BREAK_HY ||
				breakType == linebreak.G_UNICODE_BREAK_NS ||
				prevBreakType == linebreak.G_UNICODE_BREAK_BB {
				breakOp = break_PROHIBITED /* Rule LB21 */
			}
			if prevPrevBreakType == linebreak.G_UNICODE_BREAK_HL &&
				(prevBreakType == linebreak.G_UNICODE_BREAK_HY ||
					prevBreakType == linebreak.G_UNICODE_BREAK_BA) {
				breakOp = break_PROHIBITED /* Rule LB21a */
			}
			if prevBreakType == linebreak.G_UNICODE_BREAK_SY &&
				breakType == linebreak.G_UNICODE_BREAK_HL {
				breakOp = break_PROHIBITED /* Rule LB21b */
			}
			if prevBreakType == linebreak.G_UNICODE_BREAK_CB ||
				breakType == linebreak.G_UNICODE_BREAK_CB {
				breakOp = break_ALLOWED /* Rule LB20 */
			}
			if prevBreakType == linebreak.G_UNICODE_BREAK_QU ||
				breakType == linebreak.G_UNICODE_BREAK_QU {
				breakOp = break_PROHIBITED /* Rule LB19 */
			}

			/* handle related rules for Space as state machine here,
			   and override the pair table result. */
			if prevBreakType == linebreak.G_UNICODE_BREAK_SP { /* Rule LB18 */
				breakOp = break_ALLOWED
			}
			if rowBreakType == linebreak.G_UNICODE_BREAK_B2 &&
				breakType == linebreak.G_UNICODE_BREAK_B2 {
				breakOp = break_PROHIBITED /* Rule LB17 */
			}
			if (rowBreakType == linebreak.G_UNICODE_BREAK_CL ||
				rowBreakType == linebreak.G_UNICODE_BREAK_CP) &&
				breakType == linebreak.G_UNICODE_BREAK_NS {
				breakOp = break_PROHIBITED /* Rule LB16 */
			}
			if rowBreakType == linebreak.G_UNICODE_BREAK_QU &&
				breakType == linebreak.G_UNICODE_BREAK_OP {
				breakOp = break_PROHIBITED /* Rule LB15 */
			}
			if rowBreakType == linebreak.G_UNICODE_BREAK_OP {
				breakOp = break_PROHIBITED /* Rule LB14 */
			}
			/* Rule LB13 with Example 7 of Customization */
			if breakType == linebreak.G_UNICODE_BREAK_EX {
				breakOp = break_PROHIBITED
			}
			if prevBreakType != linebreak.G_UNICODE_BREAK_NU &&
				(breakType == linebreak.G_UNICODE_BREAK_CL ||
					breakType == linebreak.G_UNICODE_BREAK_CP ||
					breakType == linebreak.G_UNICODE_BREAK_IS ||
					breakType == linebreak.G_UNICODE_BREAK_SY) {
				breakOp = break_PROHIBITED
			}
			if prevBreakType == linebreak.G_UNICODE_BREAK_GL {
				breakOp = break_PROHIBITED /* Rule LB12 */
			}
			if breakType == linebreak.G_UNICODE_BREAK_GL &&
				(prevBreakType != linebreak.G_UNICODE_BREAK_SP &&
					prevBreakType != linebreak.G_UNICODE_BREAK_BA &&
					prevBreakType != linebreak.G_UNICODE_BREAK_HY) {
				breakOp = break_PROHIBITED /* Rule LB12a */
			}
			if prevBreakType == linebreak.G_UNICODE_BREAK_WJ ||
				breakType == linebreak.G_UNICODE_BREAK_WJ {
				breakOp = break_PROHIBITED /* Rule LB11 */
			}

			/* Rule LB9 */
			if breakType == linebreak.G_UNICODE_BREAK_CM ||
				breakType == linebreak.G_UNICODE_BREAK_ZWJ {
				if !(prevBreakType == linebreak.G_UNICODE_BREAK_BK ||
					prevBreakType == linebreak.G_UNICODE_BREAK_CR ||
					prevBreakType == linebreak.G_UNICODE_BREAK_LF ||
					prevBreakType == linebreak.G_UNICODE_BREAK_NL ||
					prevBreakType == linebreak.G_UNICODE_BREAK_SP ||
					prevBreakType == linebreak.G_UNICODE_BREAK_ZW) {
					breakOp = break_PROHIBITED
				}
			}

			if rowBreakType == linebreak.G_UNICODE_BREAK_ZW {
				breakOp = break_ALLOWED /* Rule LB8 */
			}
			if prevWc == 0x200D {
				breakOp = break_PROHIBITED /* Rule LB8a */
			}
			if breakType == linebreak.G_UNICODE_BREAK_SP ||
				breakType == linebreak.G_UNICODE_BREAK_ZW {
				breakOp = break_PROHIBITED /* Rule LB7 */
			}
			/* Rule LB6 */
			if breakType == linebreak.G_UNICODE_BREAK_BK ||
				breakType == linebreak.G_UNICODE_BREAK_CR ||
				breakType == linebreak.G_UNICODE_BREAK_LF ||
				breakType == linebreak.G_UNICODE_BREAK_NL {
				breakOp = break_PROHIBITED
			}
			/* Rules LB4 and LB5 */
			if prevBreakType == linebreak.G_UNICODE_BREAK_BK ||
				(prevBreakType == linebreak.G_UNICODE_BREAK_CR &&
					wc != '\n') ||
				prevBreakType == linebreak.G_UNICODE_BREAK_LF ||
				prevBreakType == linebreak.G_UNICODE_BREAK_NL {
				attrs[i].IsMandatoryBreak = true
				breakOp = break_ALLOWED
			}

			switch breakOp {
			case break_PROHIBITED:
				/* can't break here */
				attrs[i].IsLineBreak = false
			case break_IF_SPACES:
				/* break if prev char was space */
				if prevBreakType != linebreak.G_UNICODE_BREAK_SP {
					attrs[i].IsLineBreak = false
				}
			case break_ALLOWED:
				attrs[i].IsLineBreak = true
			case break_ALREADY_HANDLED:
			}

			/* Rule LB9 */
			if !(breakType == linebreak.G_UNICODE_BREAK_CM ||
				breakType == linebreak.G_UNICODE_BREAK_ZWJ) {
				/* Rule LB25 with Example 7 of Customization */
				if breakType == linebreak.G_UNICODE_BREAK_NU ||
					breakType == linebreak.G_UNICODE_BREAK_SY ||
					breakType == linebreak.G_UNICODE_BREAK_IS {
					if prevLbType != lb_Numeric {
						prevLbType = lbType
					} /* else don't change the prevLbType */
				} else {
					prevLbType = lbType
				}
			}
			/* else don't change the prevLbType for Rule LB9 */
		}

		if breakType != linebreak.G_UNICODE_BREAK_SP {
			/* Rule LB9 */
			if breakType == linebreak.G_UNICODE_BREAK_CM ||
				breakType == linebreak.G_UNICODE_BREAK_ZWJ {
				if i == 0 /* start of text */ ||
					prevBreakType == linebreak.G_UNICODE_BREAK_BK ||
					prevBreakType == linebreak.G_UNICODE_BREAK_CR ||
					prevBreakType == linebreak.G_UNICODE_BREAK_LF ||
					prevBreakType == linebreak.G_UNICODE_BREAK_NL ||
					prevBreakType == linebreak.G_UNICODE_BREAK_SP ||
					prevBreakType == linebreak.G_UNICODE_BREAK_ZW {
					prevBreakType = linebreak.G_UNICODE_BREAK_AL /* Rule LB10 */
				} /* else don't change the prevBreakType for Rule LB9 */
			} else {
				prevPrevBreakType = prevBreakType
				prevBreakType = breakType
			}
			prevJamo = jamo
		} else {
			if prevBreakType != linebreak.G_UNICODE_BREAK_SP {
				prevPrevBreakType = prevBreakType
				prevBreakType = breakType
			}
			/* else don't change the prevBreakType */
		}

		/* ---- Word breaks ---- */

		/* default to not a word start/end */
		attrs[i].IsWordStart = false
		attrs[i].IsWordEnd = false

		if currentWordType != wordNone {
			/* Check for a word end */
			switch type_ {
			case unicode.Mc, unicode.Me, unicode.Mn, unicode.Cf:
			/* nothing, we just eat these up as part of the word */
			case unicode.Ll, unicode.Lm, unicode.Lo, unicode.Lt, unicode.Lu:
				if currentWordType == wordLetters {
					/* Japanese special cases for ending the word */
					if _JAPANESE(lastWordLetter) || _JAPANESE(wc) {
						if (_HIRAGANA(lastWordLetter) &&
							!_HIRAGANA(wc)) ||
							(_KATAKANA(lastWordLetter) &&
								!(_KATAKANA(wc) || _HIRAGANA(wc))) ||
							(_KANJI(lastWordLetter) &&
								!(_HIRAGANA(wc) || _KANJI(wc))) ||
							(_JAPANESE(lastWordLetter) &&
								!_JAPANESE(wc)) ||
							(!_JAPANESE(lastWordLetter) &&
								_JAPANESE(wc)) {
							attrs[i].IsWordEnd = true
						}
					}
				}
				lastWordLetter = wc
			case unicode.Nd, unicode.Nl, unicode.No:
				lastWordLetter = wc
			default:
				/* Punctuation, control/format chars, etc. all end a word. */
				attrs[i].IsWordEnd = true
				currentWordType = wordNone
			}
		} else {
			/* Check for a word start */
			switch type_ {
			case unicode.Ll, unicode.Lm, unicode.Lo, unicode.Lt, unicode.Lu:
				currentWordType = wordLetters
				lastWordLetter = wc
				attrs[i].IsWordStart = true
			case unicode.Nd, unicode.Nl, unicode.No:
				currentWordType = wordNumbers
				lastWordLetter = wc
				attrs[i].IsWordStart = true
			default:
				/* No word here */
			}
		}

		/* ---- Sentence breaks ---- */
		{

			/* default to not a sentence start/end */
			attrs[i].IsSentenceStart = false
			attrs[i].IsSentenceEnd = false

			/* maybe start sentence */
			if lastSentenceStart == -1 && !isSentenceBoundary {
				lastSentenceStart = i - 1
			}
			/* remember last non space character position */
			if i > 0 && !attrs[i-1].IsWhite {
				lastNonSpace = i
			}
			/* meets sentence end, mark both sentence start and end */
			if lastSentenceStart != -1 && isSentenceBoundary {
				if lastNonSpace != -1 {
					attrs[lastSentenceStart].IsSentenceStart = true
					attrs[lastNonSpace].IsSentenceEnd = true
				}

				lastSentenceStart = -1
				lastNonSpace = -1
			}

			/* meets space character, move sentence start */
			if lastSentenceStart != -1 &&
				lastSentenceStart == i-1 &&
				attrs[i-1].IsWhite {
				lastSentenceStart++
			}
		}
		prevWc = wc

		/* wc might not be a valid Unicode base character, but really all we
		 * need to know is the last non-combining character */
		if type_ != unicode.Mc &&
			type_ != unicode.Me &&
			type_ != unicode.Mn {
			baseCharacter = wc
		}
	}

	i--

	attrs[i].IsCursorPosition = true /* Rule GB2 */
	attrs[0].IsCursorPosition = true /* Rule GB1 */

	attrs[i].IsWordBoundary = true /* Rule WB2 */
	attrs[0].IsWordBoundary = true /* Rule WB1 */

	attrs[i].IsLineBreak = true  /* Rule LB3 */
	attrs[0].IsLineBreak = false /* Rule LB2 */
	return attrs
}
