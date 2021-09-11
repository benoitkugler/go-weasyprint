package layout

import (
	"strings"

	"github.com/benoitkugler/go-weasyprint/backend"
	bo "github.com/benoitkugler/go-weasyprint/boxes"
	"github.com/benoitkugler/go-weasyprint/hyphen"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/utils"
	"github.com/benoitkugler/textlayout/pango"
)

type Splitted struct {
	// pango Layout with the first line
	Layout *bo.PangoLayout

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

// Return an opaque Pango layout with default Pango line-breaks.
// :param style: a style dict of computed values
// :param maxWidth: The maximum available width in the same unit as ``style['font_size']``,
// or ``nil`` for unlimited width.
func createLayout(text string, style pr.Properties, context *LayoutContext, maxWidth pr.MaybeFloat, justificationSpacing float32) *bo.PangoLayout {
	layout := bo.NewPangoLayout(context, float32(style.GetFontSize().Value), style, justificationSpacing, maxWidth)
	// Make sure that maxWidth * Pango.SCALE == maxWidth * 1024 fits in a
	// signed integer. Treat bigger values same as None: unconstrained width.
	ws := style.GetWhiteSpace()
	textWrap := "normal" == ws || "pre-wrap" == ws || "pre-line" == ws
	if maxWidth, ok := maxWidth.(pr.Float); ok && textWrap && maxWidth < 2<<21 {
		layout.Layout.SetWidth(pango.GlyphUnit(utils.PangoUnitsFromFloat(float32(utils.Maxs(0, float32(maxWidth))))))
	}

	layout.SetText(text, false)
	return layout
}

// Fit as much as possible in the available width for one line of text.
// minimum=False
func SplitFirstLine(text_ string, style pr.Properties, context *LayoutContext,
	maxWidth pr.MaybeFloat, justificationSpacing float32, minimum bool) Splitted {

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
		layout    *bo.PangoLayout
		fontSize  = style.GetFontSize().Value
		firstLine *pango.LayoutLine
		index     int
	)
	if maxWidth, ok := maxWidth.(pr.Float); ok && maxWidth != pr.Inf && fontSize != 0 {
		var expectedLength int
		if maxWidth == 0 {
			// Trying to find minimum size, let"s naively split on spaces and
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
			layout = createLayout(text_[:expectedLength], style, context, maxWidth, justificationSpacing)
			firstLine, index = layout.GetFirstLine()
			if index == -1 {
				// The small amount of text fits in one line, give up and use  the whole text
				layout = nil
			}
		}
	}

	if layout == nil {
		layout = createLayout(text_, style, context, originalMaxWidth, justificationSpacing)
		firstLine, index = layout.GetFirstLine()
	}

	resumeIndex := index
	text := []rune(text_)

	// Step #2: Don"t split lines when it"s not needed
	if maxWidth == nil {
		// The first line can take all the place needed
		return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, false, "")
	}
	maxWidthV := float32(maxWidth.V())

	firstLineWidth, _ := bo.LineSize(firstLine, style)
	if index == -1 && firstLineWidth <= maxWidthV {
		// The first line fits in the available width
		return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, false, "")
	}

	// Step #3: Try to put the first word of the second line on the first line
	// https://mail.gnome.org/archives/gtk-i18n-list/2013-September/msg00006
	// is a good thread related to this problem.
	firstLineText := string(text[:index])
	// We can’t rely on firstLineWidth, see
	// https://github.com/Kozea/WeasyPrint/issues/1051
	firstLineFits := (firstLineWidth <= maxWidthV ||
		strings.ContainsRune(strings.TrimSpace(firstLineText), ' ') ||
		canBreakText(strings.TrimSpace(firstLineText), style.GetLang()))
	var secondLineText []rune
	if firstLineFits {
		// The first line fits but may have been cut too early by Pango
		secondLineText = text[index:]
	} else {
		// The line can't be split earlier, try to hyphenate the first word.
		firstLineText = ""
		secondLineText = text
	}

	nextWord := strings.SplitN(string(secondLineText), " ", 1)[0]
	if nextWord != "" {
		if spaceCollapse {
			// nextWord might fit without a space afterwards
			// only try when space collapsing is allowed
			newFirstLineText := firstLineText + nextWord
			layout.SetText(newFirstLineText, false)
			firstLine, index = layout.GetFirstLine()
			firstLineWidth, _ = bo.LineSize(firstLine, style)
			if index == -1 && firstLineText != "" {
				// The next word fits in the first line, keep the layout
				resumeIndex = len(newFirstLineText.encode("utf-8")) + 1
				return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style, false, "")
			} else if index != -1 && index != 0 {
				// Text may have been split elsewhere by Pango earlier
				resumeIndex = index
			} else {
				// Second line is none
				resumeIndex = firstLine.Length + 1
				if resumeIndex >= len(text.encode("utf-8")) {
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
	lang := style.GetLang().String
	if lang != "" {
		lang = hyphen.LanguageFallback(lang)
	}
	limit := style.GetHyphenateLimitChars()
	total, left, right := limit[0], limit[1], limit[2]
	hyphenated := false
	softHyphen := '\u00ad'

	tryHyphenate := false
	if hyphens != "none" {
		nextWordBoundaries := getNextWordBoundaries(secondLineText)
		if len(nextWordBoundaries) == 2 {
			// We have a word to hyphenate
			startWord, stopWord := nextWordBoundaries[0], nextWordBoundaries[1]
			nextWord = string(secondLineText[startWord:stopWord])
			if stopWord-startWord >= total {
				// This word is long enough
				firstLineWidth, _ = bo.LineSize(firstLine, style)
				space := maxWidthV - firstLineWidth
				zone := style.GetHyphenateLimitZone()
				limitZone := float32(zone.Value)
				if zone.Unit == pr.Percentage {
					limitZone = (maxWidthV * float32(zone.Value) / 100.)
				}
				if space > limitZone || space < 0 {
					// Available space is worth the try, or the line is even too
					// long to fit: try to hyphenate
					tryHyphenate = true
				}
			}
		}

		var manualHyphenation bool
		if tryHyphenate {
			// Automatic hyphenation possible and next word is long enough
			autoHyphenation := hyphens == "auto" && lang != ""
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

		if manualHyphenation {
			// Manual hyphenation: check that the line ends with a soft
			// hyphen && add the missing hyphen
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
			// softHyphenIndexes = []{match.start() for match := range re.finditer(softHyphen, nextWord)}
			softHyphenIndexes.reverse()
			// dictionaryIterations = [nextWord[:i + 1] for i := range softHyphenIndexes]
		} else if autoHyphenation {
			// dictionaryKey = (lang, left, right, total)
			dictionary = context.dictionaries.get(dictionaryKey)
			if dictionary == nil {
				dictionary = hyphen.Pyphen(lang, left, right)
				context.dictionaries[dictionaryKey] = dictionary
			}
			//  dictionaryIterations = [    start for start, end := range dictionary.iterate(nextWord)]
		} else {
			// dictionaryIterations = []
		}

		if dictionaryIterations {
			for firstWordPart := range dictionaryIterations {
				newFirstLineText = (firstLineText +
					secondLineText[:startWord] +
					firstWordPart)
				hyphenatedFirstLineText = (newFirstLineText + style["hyphenateCharacter"])
				newLayout = createLayout(
					hyphenatedFirstLineText, style, context, maxWidth,
					justificationSpacing)
				newFirstLine, newIndex = newLayout.getFirstLine()
				newFirstLineWidth, _ = bo.LineSize(newFirstLine, style)
				newSpace = maxWidth - newFirstLineWidth
				if newIndex == nil && (newSpace >= 0 ||
					firstWordPart == dictionaryIterations[-1]) {

					hyphenated = true
					layout = newLayout
					firstLine = newFirstLine
					index = newIndex
					resumeIndex = len(newFirstLineText.encode("utf8"))
					if text[len(newFirstLineText)] == softHyphen {
						// Recreate the layout with no maxWidth to be sure that
						// we don't break before the soft hyphen
						pango.pangoLayoutSetWidth(
							layout.layout, unitsFromDouble(-1))
						resumeIndex += len(softHyphen.encode("utf8"))
					}
					break
				}
			}

			if !hyphenated && !firstLineText {
				// Recreate the layout with no maxWidth to be sure that
				// we don't break before || inside the hyphenate character
				hyphenated = true
				layout.setText(hyphenatedFirstLineText)
				pango.pangoLayoutSetWidth(
					layout.layout, unitsFromDouble(-1))
				firstLine, index = layout.getFirstLine()
				resumeIndex = len(newFirstLineText.encode("utf8"))
				if text[len(firstLineText)] == softHyphen {
					resumeIndex += len(softHyphen.encode("utf8"))
				}
			}
		}
	}

	if !hyphenated && firstLineText.endswith(softHyphen) {
		// Recreate the layout with no maxWidth to be sure that
		// we don't break inside the hyphenate-character string
		hyphenated = true
		hyphenatedFirstLineText = (firstLineText + style["hyphenateCharacter"])
		layout.setText(hyphenatedFirstLineText)
		pango.pangoLayoutSetWidth(
			layout.layout, unitsFromDouble(-1))
		firstLine, index = layout.getFirstLine()
		resumeIndex = len(firstLineText.encode("utf8"))
	}

	// Step 5: Try to break word if it"s too long for the line
	overflowWrap = style["overflowWrap"]
	firstLineWidth, _ = bo.LineSize(firstLine, style)
	space = maxWidth - firstLineWidth
	// If we can break words && the first line is too long
	if !minimum && overflowWrap == "break-word" && space < 0 {
		// Is it really OK to remove hyphenation for word-break ?
		hyphenated = false
		// TODO: Modify code to preserve W3C condition {
		// "Shaping characters are still shaped as if the word were ! broken"
		// The way new lines are processed := range this function (one by one with no
		// memory of the last) prevents shaping characters (arabic, for
		// instance) from keeping their shape when wrapped on the next line with
		// pango layout. Maybe insert Unicode shaping characters := range text?
		layout.setText(text)
		pango.pangoLayoutSetWidth(
			layout.layout, unitsFromDouble(maxWidth))
		pango.pangoLayoutSetWrap(
			layout.layout, PANGOWRAPMODE["WRAPCHAR"])
		firstLine, index = layout.getFirstLine()
		resumeIndex = index || firstLine.length
		if resumeIndex >= len(text.encode("utf-8")) {
			resumeIndex = nil
		}
	}

	return firstLineMetrics(firstLine, text, layout, resumeIndex, spaceCollapse, style,
		hyphenated, style.GetHyphenateCharacter())

	// return Splitted{}
}

func firstLineMetrics(firstLine *pango.LayoutLine, text []rune, layout *bo.PangoLayout, resumeAt int, spaceCollapse bool,
	style pr.Properties, hyphenated bool, hyphenationCharacter string) Splitted {
	length := firstLine.Length
	if hyphenated {
		length -= len([]rune(hyphenationCharacter))
	} else if resumeAt != 0 {
		// Set an infinite width as we don't want to break lines when drawing,
		// the lines have already been split and the size may differ. Rendering
		// is also much faster when no width is set.
		layout.Layout.SetWidth(-1)
	}

	// Create layout with final text
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

	width, height := bo.LineSize(firstLine, style)
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
	// TODO: this should be removed when bidi is supported
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

func CanBreakText(t []rune, lang string) MaybeBool {
	if len(t) < 2 {
		return nil
	}
	logs := text.GetLogAttrs(t)
	for _, l := range logs[1 : len(logs)-1] {
		if l.IsLineBreak() {
			return Bool(true)
		}
	}
	return Bool(false)
}

// Return a tuple of the used value of ``line-height`` and the baseline.
// The baseline is given from the top edge of line height.
func StrutLayout(style pr.Properties, context *LayoutContext) (pr.Float, pr.Float) {
	// FIXME: à implémenter
	return 0.5, 0.5
}

// Return the ratio 1ex/font_size, according to given style.
func ExRatio(style pr.Properties) pr.Float {
	// FIXME: à implémenter
	return .5
}

// Draw the given ``textbox`` line to the cairo ``context``.
func ShowFirstLine(context backend.Drawer, textbox bo.TextBox, textOverflow string) {
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
}
