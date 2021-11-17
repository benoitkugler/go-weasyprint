package text

import (
	"math"
	"strings"

	"github.com/benoitkugler/go-weasyprint/layout/text/hyphen"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/utils"
	"github.com/benoitkugler/textlayout/language"
	"github.com/benoitkugler/textlayout/pango"
)

// Splitted exposes the result of laying out
// one line of text
type Splitted struct {
	// pango Layout with the first line
	Layout *TextLayout

	// length in runes of the first line
	Length int

	// the number of runes to skip for the next line.
	// May be -1 if the whole text fits in one line.
	// This may be greater than `Length` in case of preserved
	// newline characters.
	ResumeAt int

	// width in pixels of the first line
	Width pr.Float

	// height in pixels of the first line
	Height pr.Float

	// baseline in pixels of the first line
	Baseline pr.Float
}

// CreateLayout returns a pango.Layout with default Pango line-breaks.
// `style` is a style dict of computed values.
// `maxWidth` is the maximum available width in the same unit as style.GetFontSize(),
// or `nil` for unlimited width.
func CreateLayout(text string, style pr.StyleAccessor, context TextLayoutContext, maxWidth pr.MaybeFloat, justificationSpacing pr.Float) *TextLayout {
	layout := NewTextLayout(context, pr.Fl(style.GetFontSize().Value), style, pr.Fl(justificationSpacing), maxWidth)
	// Make sure that maxWidth * Pango.SCALE == maxWidth * 1024 fits in a
	// signed integer. Treat bigger values same as None: unconstrained width.
	ws := style.GetWhiteSpace()
	textWrap := "normal" == ws || "pre-wrap" == ws || "pre-line" == ws
	if maxWidth, ok := maxWidth.(pr.Float); ok && textWrap && maxWidth < 2<<21 {
		layout.Layout.SetWidth(pango.GlyphUnit(utils.PangoUnitsFromFloat(utils.Maxs(0, pr.Fl(maxWidth)))))
	}

	layout.SetText(text, false)
	return layout
}

func hyphenDictionaryIterations(word string, hyphen rune) (out []string) {
	wordRunes := []rune(word)
	for i := len(wordRunes) - 1; i >= 0; i-- {
		if wordRunes[i] == hyphen {
			out = append(out, string(wordRunes[:i+1]))
		}
	}
	return out
}

type HyphenDictKey struct {
	lang               language.Language
	left, right, total int
}

// Fit as much as possible in the available width for one line of text.
// minimum=False
func SplitFirstLine(text_ string, style pr.StyleAccessor, context TextLayoutContext,
	maxWidth pr.MaybeFloat, justificationSpacing pr.Float, minimum bool) Splitted {
	// See https://www.w3.org/TR/css-text-3/#white-space-property
	ws := style.GetWhiteSpace()
	textWrap := "normal" == ws || "pre-wrap" == ws || "pre-line" == ws
	spaceCollapse := "normal" == ws || "nowrap" == ws || "pre-line" == ws

	originalMaxWidth := maxWidth
	if !textWrap {
		maxWidth = nil
	}
	// Step #1: Get a draft layout with the first line
	var (
		layout    *TextLayout
		fontSize  = style.GetFontSize().Value
		firstLine *pango.LayoutLine
		index     int
	)
	if maxWidth, ok := maxWidth.(pr.Float); ok && maxWidth != pr.Inf && fontSize != 0 {
		var expectedLength int
		if maxWidth <= 0 {
			// Trying to find minimum size, let's naively split on spaces and
			// keep one word + one letter

			if spaceIndex := strings.IndexByte(text_, ' '); spaceIndex != -1 {
				expectedLength = spaceIndex + 2 // index + space + one letter
			} else {
				expectedLength = len(text_)
			}
		} else {
			expectedLength = int(maxWidth / fontSize * 2.5)
		}

		if expectedLength < len(text_) {
			// Try to use a small amount of text instead of the whole text
			layout = CreateLayout(text_[:expectedLength], style, context, maxWidth, justificationSpacing)
			firstLine, index = layout.GetFirstLine()
			if index == -1 {
				// The small amount of text fits in one line, give up and use the whole text
				layout = nil
			}
		}
	}

	if layout == nil {
		layout = CreateLayout(text_, style, context, originalMaxWidth, justificationSpacing)
		firstLine, index = layout.GetFirstLine()
	}
	resumeIndex := index
	text := []rune(text_)

	// Step #2: Don't split lines when it's not needed
	if maxWidth == nil {
		// The first line can take all the place needed
		return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, false, "")
	}
	maxWidthV := pr.Fl(maxWidth.V())

	firstLineWidth, _ := LineSize(firstLine, style)
	if index == -1 && firstLineWidth <= maxWidthV {
		// The first line fits in the available width
		return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, false, "")
	}

	// Step #3: Try to put the first word of the second line on the first line
	// https://mail.gnome.org/archives/gtk-i18n-list/2013-September/msg00006
	// is a good thread related to this problem.
	if index == -1 || index >= len(text) {
		index = len(text)
	}
	firstLineText := string(text[:index])
	// We can’t rely on firstLineWidth, see
	// https://github.com/Kozea/WeasyPrint/issues/1051
	firstLineFits := (firstLineWidth <= maxWidthV ||
		strings.ContainsRune(strings.TrimSpace(firstLineText), ' ') ||
		CanBreakText([]rune(strings.TrimSpace(firstLineText))) == pr.True)
	var secondLineText []rune
	if firstLineFits {
		// The first line fits but may have been cut too early by Pango
		secondLineText = text[index:]
	} else {
		// The line can't be split earlier, try to hyphenate the first word.
		firstLineText = ""
		secondLineText = text
	}

	nextWord := strings.SplitN(string(secondLineText), " ", 2)[0]
	if nextWord != "" {
		if spaceCollapse {
			// nextWord might fit without a space afterwards
			// only try when space collapsing is allowed
			newFirstLineText := firstLineText + nextWord
			layout.SetText(newFirstLineText, false)
			firstLine, index = layout.GetFirstLine()
			firstLineWidth, _ = LineSize(firstLine, style)
			if index == -1 && firstLineText != "" {
				// The next word fits in the first line, keep the layout
				resumeIndex = len([]rune(newFirstLineText)) + 1
				return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, false, "")
			} else if index != -1 && index != 0 {
				// Text may have been split elsewhere by Pango earlier
				resumeIndex = index
			} else {
				// Second line is none
				resumeIndex = firstLine.Length + 1
				if resumeIndex >= len(text) {
					resumeIndex = -1
				}
			}
		}
	} else if firstLineText != "" {
		// We found something on the first line but we did ! find a word on
		// the next line, no need to hyphenate, we can keep the current layout
		return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, false, "")
	}

	// Step #4: Try to hyphenate
	hyphens := style.GetHyphens()
	lang := language.NewLanguage(style.GetLang().String)
	if lang != "" {
		lang = hyphen.LanguageFallback(lang)
	}
	limit := style.GetHyphenateLimitChars()
	total, left, right := limit[0], limit[1], limit[2]
	hyphenated := false
	softHyphen := '\u00ad'

	hyphenateCharacter := string(style.GetHyphenateCharacter())

	tryHyphenate := false
	var startWord, stopWord int
	if hyphens != "none" {
		nextWordBoundaries := getNextWordBoundaries(secondLineText)
		if len(nextWordBoundaries) == 2 {
			// We have a word to hyphenate
			startWord, stopWord = nextWordBoundaries[0], nextWordBoundaries[1]
			nextWord = string(secondLineText[startWord:stopWord])
			if stopWord-startWord >= total {
				// This word is long enough
				firstLineWidth, _ = LineSize(firstLine, style)
				space := maxWidthV - firstLineWidth
				zone := style.GetHyphenateLimitZone()
				limitZone := pr.Fl(zone.Value)
				if zone.Unit == pr.Percentage {
					limitZone = (maxWidthV * pr.Fl(zone.Value) / 100.)
				}
				if space > limitZone || space < 0 {
					// Available space is worth the try, or the line is even too
					// long to fit: try to hyphenate
					tryHyphenate = true
				}
			}
		}

		var manualHyphenation, autoHyphenation bool
		if tryHyphenate {
			// Automatic hyphenation possible and next word is long enough
			autoHyphenation = hyphens == "auto" && lang != ""
			manualHyphenation = false
			if autoHyphenation {
				if strings.ContainsRune(firstLineText, softHyphen) || strings.ContainsRune(nextWord, softHyphen) {
					// Automatic hyphenation opportunities within a word must be
					// ignored if the word contains a conditional hyphen, in favor
					// of the conditional hyphen(s).
					// See https://drafts.csswg.org/css-text-3/#valdef-hyphens-auto
					manualHyphenation = true
				}
			} else {
				manualHyphenation = hyphens == "manual"
			}
		}

		var dictionaryIterations []string
		if manualHyphenation {
			// Manual hyphenation: check that the line ends with a soft
			// hyphen and add the missing hyphen
			if strings.HasSuffix(firstLineText, string(softHyphen)) {
				// The first line has been split on a soft hyphen
				if id := strings.LastIndexByte(firstLineText, ' '); id != -1 {
					firstLineText, nextWord = firstLineText[:id], firstLineText[id+1:]
					nextWord = " " + nextWord
					layout.SetText(firstLineText, false)
					firstLine, index = layout.GetFirstLine()
					resumeIndex = len([]rune(firstLineText + " "))
				} else {
					firstLineText, nextWord = "", firstLineText
				}
			}
			dictionaryIterations = hyphenDictionaryIterations(nextWord, softHyphen)
		} else if autoHyphenation {
			dictionaryKey := HyphenDictKey{lang, left, right, total}
			dictionary, ok := context.HyphenCache()[dictionaryKey]
			if !ok {
				dictionary = hyphen.NewHyphener(lang, left, right)
				context.HyphenCache()[dictionaryKey] = dictionary
			}
			dictionaryIterations = dictionary.Iterate(nextWord)
		}

		if len(dictionaryIterations) != 0 {
			var newFirstLineText, hyphenatedFirstLineText string
			for _, firstWordPart := range dictionaryIterations {
				newFirstLineText = (firstLineText + string(secondLineText[:startWord]) + firstWordPart)
				hyphenatedFirstLineText = (newFirstLineText + hyphenateCharacter)
				newLayout := CreateLayout(hyphenatedFirstLineText, style, context, maxWidth, justificationSpacing)
				newFirstLine, newIndex := newLayout.GetFirstLine()
				newFirstLineWidth, _ := LineSize(newFirstLine, style)
				newSpace := maxWidthV - newFirstLineWidth
				if newIndex == -1 && (newSpace >= 0 || firstWordPart == dictionaryIterations[len(dictionaryIterations)-1]) {
					hyphenated = true
					layout = newLayout
					firstLine = newFirstLine
					index = newIndex
					resumeIndex = len([]rune(newFirstLineText))
					if text[resumeIndex] == softHyphen {
						// Recreate the layout with no maxWidth to be sure that
						// we don't break before the soft hyphen
						layout.Layout.SetWidth(-1)
						resumeIndex += 1
					}
					break
				}
			}

			if !hyphenated && firstLineText == "" {
				// Recreate the layout with no maxWidth to be sure that
				// we don't break before or inside the hyphenate character
				hyphenated = true
				layout.SetText(hyphenatedFirstLineText, false)
				layout.Layout.SetWidth(-1)
				firstLine, index = layout.GetFirstLine()
				resumeIndex = len([]rune(newFirstLineText))
				if text[resumeIndex] == softHyphen {
					resumeIndex += 1
				}
			}
		}
	}

	if !hyphenated && strings.HasSuffix(firstLineText, string(softHyphen)) {
		// Recreate the layout with no maxWidth to be sure that
		// we don't break inside the hyphenate-character string
		hyphenated = true
		hyphenatedFirstLineText := firstLineText + hyphenateCharacter
		layout.SetText(hyphenatedFirstLineText, false)
		layout.Layout.SetWidth(-1)
		firstLine, index = layout.GetFirstLine()
		resumeIndex = len([]rune(firstLineText))
	}

	// Step 5: Try to break word if it's too long for the line
	overflowWrap := style.GetOverflowWrap()
	firstLineWidth, _ = LineSize(firstLine, style)
	space := maxWidthV - firstLineWidth
	// If we can break words and the first line is too long
	if !minimum && overflowWrap == "break-word" && space < 0 {
		// Is it really OK to remove hyphenation for word-break ?
		hyphenated = false
		layout.SetText(string(text), false)
		layout.Layout.SetWidth(pango.GlyphUnit(utils.PangoUnitsFromFloat(maxWidthV)))
		layout.Layout.SetWrap(pango.WRAP_CHAR)
		firstLine, index = layout.GetFirstLine()
		resumeIndex = index
		if resumeIndex == 0 {
			resumeIndex = firstLine.Length
		}
		if resumeIndex >= len(text) {
			resumeIndex = -1
		}
	}

	return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, hyphenated, hyphenateCharacter)
}

func firstLineMetrics(firstLine *pango.LayoutLine, text []rune, layout *TextLayout, resumeAt int, spaceCollapse bool,
	style pr.StyleAccessor, hyphenated bool, hyphenationCharacter string) Splitted {
	length := firstLine.Length
	if hyphenated {
		length -= len([]rune(hyphenationCharacter))
	} else if resumeAt != -1 && resumeAt != 0 {
		// Set an infinite width as we don't want to break lines when drawing,
		// the lines have already been split and the size may differ. Rendering
		// is also much faster when no width is set.
		layout.Layout.SetWidth(-1)

		// Create layout with final text
		if length > len(text) {
			length = len(text)
		}
		firstLineText := string(text[:length])

		// Remove trailing spaces if spaces collapse
		if spaceCollapse {
			firstLineText = strings.TrimRight(firstLineText, " ")
		}

		// Remove soft hyphens
		textNoHyphens := strings.ReplaceAll(firstLineText, "\u00ad", "")
		layout.SetText(textNoHyphens, false)

		firstLine, _ = layout.GetFirstLine()
		length = 0

		if firstLine != nil {
			length = firstLine.Length
		}

		// add the number of hypen chars (\u00ad is 2 bytes in utf8)
		length += (len(firstLineText) - len(textNoHyphens)) / 2
	}

	width, height := LineSize(firstLine, style)
	baseline := utils.PangoUnitsToFloat(layout.Layout.GetBaseline())
	// layout.deactivate()
	return Splitted{Layout: layout, Length: length, ResumeAt: resumeAt, Width: pr.Float(width), Height: pr.Float(height), Baseline: pr.Float(baseline)}
}

var rp = strings.NewReplacer(
	"\u202a", "\u200b",
	"\u202b", "\u200b",
	"\u202c", "\u200b",
	"\u202d", "\u200b",
	"\u202e", "\u200b",
)

func getLogAttrs(text []rune) []pango.CharAttr {
	text = []rune(rp.Replace(string(text)))
	logAttrs := pango.ComputeCharacterAttributes(text, -1)
	return logAttrs
}

// returns nil or [wordStart, wordEnd]
func getNextWordBoundaries(t []rune) []int {
	if len(t) < 2 {
		return nil
	}
	out := make([]int, 2)
	hasBroken := false
	for i, attr := range getLogAttrs(t) {
		if attr.IsWordEnd() {
			out[1] = i // word end
			hasBroken = true
			break
		}
		if attr.IsWordBoundary() {
			out[0] = i // word start
		}
	}
	if !hasBroken {
		return nil
	}
	return out
}

// GetLastWordEnd returns the index in `t` if the last word,
// or -1
func GetLastWordEnd(t []rune) int {
	if len(t) < 2 {
		return -1
	}
	attrs := getLogAttrs(t)
	for i := 0; i < len(attrs); i++ {
		item := attrs[len(attrs)-1]
		if i != 0 && item.IsWordEnd() {
			return len(t) - i
		}
	}
	return -1
}

func CanBreakText(t []rune) pr.MaybeBool {
	if len(t) < 2 {
		return nil
	}
	logs := getLogAttrs(t)
	for _, l := range logs[1 : len(logs)-1] {
		if l.IsLineBreak() {
			return pr.True
		}
	}
	return pr.False
}

type StrutLayoutKey struct {
	fontLanguageOverride pr.String
	lang                 string
	fontFamily           string // joined
	fontStretch          pr.String
	fontStyle            pr.String
	fontWeight           pr.IntString
	lineHeight           pr.Value
	fontSize             pr.Float
}

// StrutLayout returns a tuple of the used value of `line-height` and the baseline.
// The baseline is given from the top edge of line height.
// `context` is mandatory for the text layout.
func StrutLayout(style pr.StyleAccessor, context TextLayoutContext) [2]pr.Float {
	fontSize := style.GetFontSize().Value
	lineHeight := style.GetLineHeight()
	if fontSize == 0 {
		return [2]pr.Float{}
	}

	key := StrutLayoutKey{
		fontSize:             fontSize,
		fontLanguageOverride: style.GetFontLanguageOverride(),
		lang:                 style.GetLang().String,
		fontFamily:           strings.Join(style.GetFontFamily(), ""),
		fontStyle:            style.GetFontStyle(),
		fontStretch:          style.GetFontStretch(),
		fontWeight:           style.GetFontWeight(),
		lineHeight:           lineHeight,
	}

	layouts := context.StrutLayoutsCache()
	if v, ok := layouts[key]; ok {
		return v
	}

	layout := NewTextLayout(context, pr.Fl(fontSize), style, 0, nil)
	layout.SetText(" ", false)
	line, _ := layout.GetFirstLine()
	sp := firstLineMetrics(line, nil, layout, -1, false, style, false, "")
	if lineHeight.String == "normal" {
		result := [2]pr.Float{sp.Height, sp.Baseline}
		if context != nil {
			context.StrutLayoutsCache()[key] = result
		}
		return result
	}
	lineHeightV := lineHeight.Value
	if lineHeight.Unit == pr.Scalar {
		lineHeightV *= fontSize
	}
	result := [2]pr.Float{lineHeightV, sp.Baseline + (lineHeightV-sp.Height)/2}
	if context != nil {
		context.StrutLayoutsCache()[key] = result
	}
	return result
}

// ExRatio returns the ratio 1ex/font_size, according to given style.
// It should be used with a valid text context to get accurate result.
// Otherwise, if context is `nil`, it returns 1 as a default value.
func ExRatio(style pr.ElementStyle, context TextLayoutContext) pr.Float {
	if context == nil {
		return 1
	}

	// Avoid recursion for letter-spacing && word-spacing properties
	style = style.Copy()
	style.SetLetterSpacing(pr.SToV("normal"))
	style.SetWordSpacing(pr.FToV(0))

	// Random big value
	var fontSize pr.Fl = 1000

	layout := NewTextLayout(context, fontSize, style, 0, nil)
	layout.SetText("x", false)
	line, _ := layout.GetFirstLine()

	var inkExtents pango.Rectangle
	line.GetExtents(&inkExtents, nil)
	heightAboveBaseline := utils.PangoUnitsToFloat(inkExtents.Y)

	// Zero means some kind of failure, fallback is 0.5.
	// We round to try keeping exact values that were altered by Pango.
	v := math.Round(float64(-heightAboveBaseline/fontSize)*100000) / 100000
	if v == 0 {
		return 0.5
	}
	return pr.Float(v)
}

// Draw the given ``textbox`` line to the cairo ``context``.
// func ShowFirstLine(context backend.Drawer, textbox TextBox, textOverflow string) {
// FIXME: à implémenter
// pango.pangoLayoutSetSingleParagraphMode(textbox.PangoLayout.Layout, true)

// if textOverflow == "ellipsis" {
// 	maxWidth := context.ClipExtents()[2] - float64(textbox.PositionX)
// 	pango.pangoLayoutSetWidth(textbox.PangoLayout.Layout, unitsFromDouble(maxWidth))
// 	pango.pangoLayoutSetEllipsize(textbox.PangoLayout.Layout, pango.PANGOELLIPSIZEEND)
// }

// firstLine, _ = textbox.PangoLayout.GetFirstLine()
// context = ffi.cast("cairoT *", context.Pointer)
// pangocairo.pangoCairoShowLayoutLine(context, firstLine)
// }
