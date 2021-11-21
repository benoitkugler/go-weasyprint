package layout

import (
	"fmt"
	"log"
	"strings"
	"unicode"

	"github.com/benoitkugler/go-weasyprint/layout/text"
	"github.com/benoitkugler/go-weasyprint/utils"

	bo "github.com/benoitkugler/go-weasyprint/boxes"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/style/tree"
)

//     Line breaking and layout for inline-level boxes.

// IsLine returns wheter `box` is either
// an instance of LineBox or InlineBox.
func IsLine(box Box) bool {
	return bo.LineBoxT.IsInstance(box) || bo.InlineBoxT.IsInstance(box)
}

type lineBoxe struct {
	// laid-out LineBox with as much content as possible that
	// fits in the available width.
	line     *bo.LineBox
	resumeAt *tree.IntList
}

type lineBoxeIterator struct {
	firstLetterStyle pr.ElementStyle

	currentBox lineBoxe

	box             *bo.LineBox
	containingBlock *bo.BoxFields
	fixedBoxes      *[]*AbsolutePlaceholder
	skipStack       *tree.IntList
	context         *layoutContext
	absoluteBoxes   *[]*AbsolutePlaceholder
	positionY       pr.Float

	resetTextIndent bool
	done            bool
}

func (l *lineBoxeIterator) Has() bool {
	if l.resetTextIndent {
		l.box.TextIndent = pr.Float(0)
	}
	if l.done {
		return false
	}
	line, resumeAt := getNextLinebox(l.context, l.box, l.positionY, l.skipStack, l.containingBlock,
		l.absoluteBoxes, l.fixedBoxes, l.firstLetterStyle)

	if debugMode {
		debugLogger.Line("getNextLinebox resumeAt: %s", resumeAt)
	}

	if line != nil {
		handleLeader(l.context, line, l.containingBlock)
		l.positionY = line.Box().PositionY + line.Box().Height.V()
	}
	if line == nil {
		l.done = true
		return false
	}
	l.currentBox = lineBoxe{line: line, resumeAt: resumeAt}
	if resumeAt == nil {
		l.done = true
	}
	l.skipStack = resumeAt
	l.resetTextIndent = true
	l.firstLetterStyle = nil
	return true
}

func (l *lineBoxeIterator) Next() lineBoxe { return l.currentBox }

// `box` is a non-laid-out `LineBox`
// positionY is the vertical top position of the line box on the page
// skipStack is ``nil`` to start at the beginning of ``linebox``,
// or a ``resumeAt`` value to continue just after an
// already laid-out line.
func iterLineBoxes(context *layoutContext, box *bo.LineBox, positionY pr.Float, skipStack *tree.IntList, containingBlock *bo.BoxFields,
	absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder, firstLetterStyle pr.ElementStyle) *lineBoxeIterator {
	resolvePercentages(box, bo.MaybePoint{containingBlock.Width, containingBlock.Height}, "")
	if skipStack == nil {
		// TODO: wrong, see https://github.com/Kozea/WeasyPrint/issues/679
		box.TextIndent = resolveOnePercentage(box.Style.GetTextIndent(), "text_indent", containingBlock.Width.V(), "")
	} else {
		box.TextIndent = pr.Float(0)
	}
	return &lineBoxeIterator{
		box: box, context: context, positionY: positionY, containingBlock: containingBlock,
		absoluteBoxes: absoluteBoxes, fixedBoxes: fixedBoxes, skipStack: skipStack, firstLetterStyle: firstLetterStyle,
	}
}

func getNextLinebox(context *layoutContext, linebox *bo.LineBox, positionY pr.Float, skipStack *tree.IntList,
	containingBlock *bo.BoxFields, absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder,
	firstLetterStyle pr.ElementStyle) (line_ *bo.LineBox, resumeAt *tree.IntList) {

	skipStack, cont := skipFirstWhitespace(linebox, skipStack)
	if cont {
		return
	}

	skipStack = firstLetterToBox(context, linebox, skipStack, firstLetterStyle)

	linebox.PositionY = positionY

	if len(*context.excludedShapes) != 0 {
		// Width and height must be calculated to avoid floats
		linebox.Width = inlineMinContentWidth(context, linebox, true, skipStack, true, false)
		linebox.Height = text.StrutLayout(linebox.Style, context)[0]
	} else {
		// No float, width and height will be set by the lines
		linebox.Width = pr.Float(0)
		linebox.Height = pr.Float(0)
	}
	positionX, positionY, availableWidth := avoidCollisions(context, linebox, containingBlock, false)

	candidateHeight := linebox.Height

	excludedShapes := append([]*bo.BoxFields{}, *context.excludedShapes...)

	var (
		linePlaceholders, lineAbsolutes, lineFixed []*AbsolutePlaceholder
		waitingFloats                              []Box
	)
	for {
		linebox.PositionX, linebox.PositionY = positionX, positionY
		originalPositionX, originalPositionY := positionX, positionY
		originalWidth := linebox.Width.V()

		maxX := positionX + availableWidth
		positionX += linebox.TextIndent.V()

		var (
			preservedLineBreak bool
			floatWidths        widths
		)

		spi := splitInlineBox(context, linebox, positionX, maxX, skipStack, containingBlock,
			&lineAbsolutes, &lineFixed, &linePlaceholders, waitingFloats, nil)
		resumeAt, preservedLineBreak, floatWidths = spi.resumeAt, spi.preservedLineBreak, spi.floatWidths
		line_ = spi.newBox.(*bo.LineBox) // splitInlineBox preserve the concrete type

		line := line_.Box()
		linebox.Width, linebox.Height = line.Width, line.Height

		if isPhantomLinebox(*line) && !preservedLineBreak {
			line.Height = pr.Float(0)
			break
		}

		removeLastWhitespace(context, line_)

		newPositionX, _, newAvailableWidth := avoidCollisions(context, linebox, containingBlock, false)
		newAvailableWidth -= floatWidths.right
		alignmentAvailableWidth := newAvailableWidth + newPositionX - linebox.PositionX
		offsetX := textAlign(context, line_, alignmentAvailableWidth, resumeAt == nil || preservedLineBreak)

		if containingBlock.Style.GetDirection() == "rtl" {
			offsetX = -offsetX
			offsetX -= line.Width.V()
		}

		bottom, top := lineBoxVerticality(context, line_)
		line.Baseline = -top
		line.PositionY = top
		line.Height = bottom - top
		offsetY := positionY - top
		line.MarginTop = pr.Float(0)
		line.MarginBottom = pr.Float(0)

		line_.Translate(line_, offsetX, offsetY, false)
		// Avoid floating point errors, as positionY - top + top != positionY
		// Removing this line breaks the position == linebox.Position test below
		// See https://github.com/Kozea/WeasyPrint/issues/583
		line.PositionY = positionY

		if line.Height.V() <= candidateHeight.V() {
			break
		}
		candidateHeight = line.Height

		newExcludedShapes := context.excludedShapes
		context.excludedShapes = &excludedShapes
		positionX, positionY, availableWidth = avoidCollisions(context, line_, containingBlock, false)

		var condition bool
		if containingBlock.Style.GetDirection() == "ltr" {
			condition = positionX == originalPositionX && positionY == originalPositionY
		} else {
			condition = positionX+line.Width.V() == originalPositionX+originalWidth && positionY == originalPositionY
		}

		if condition {
			context.excludedShapes = newExcludedShapes
			break
		}
	}
	*absoluteBoxes = append(*absoluteBoxes, lineAbsolutes...)
	*fixedBoxes = append(*fixedBoxes, lineFixed...)

	line := line_.Box()
	for _, placeholder := range linePlaceholders {
		if placeholder.Box().Style.Specified().Display.Has("inline") {
			// Inline-level static position :
			placeholder.Translate(placeholder, 0, positionY-placeholder.Box().PositionY.V(), false)
		} else {
			// Block-level static position: at the start of the next line
			placeholder.Translate(placeholder, line.PositionX-placeholder.Box().PositionX.V(),
				positionY+line.Height.V()-placeholder.Box().PositionY.V(), false)
		}
	}

	var floatChildren []Box
	waitingFloatsY := line.PositionY + line.Height.V()
	for _, waitingFloat_ := range waitingFloats {
		waitingFloat := waitingFloat_.Box()
		waitingFloat.PositionY = waitingFloatsY
		waitingFloat_ = floatLayout(context, waitingFloat_, containingBlock, absoluteBoxes, fixedBoxes)
		floatChildren = append(floatChildren, waitingFloat_)
	}
	line.Children = append(line.Children, floatChildren...)

	return line_, resumeAt
}

// Return the ``skipStack`` to start just after the remove spaces
// at the beginning of the line.
// See http://www.w3.org/TR/CSS21/text.html#white-space-model
func skipFirstWhitespace(box Box, skipStack *tree.IntList) (*tree.IntList, bool) {
	var (
		index         int
		nextSkipStack *tree.IntList
	)
	if skipStack != nil {
		index, nextSkipStack = skipStack.Value, skipStack.Next
	}
	if textBox, ok := box.(*bo.TextBox); ok {
		if nextSkipStack != nil {
			panic(fmt.Sprintf("expected nil nextSkipStack, got %v", nextSkipStack))
		}
		whiteSpace := textBox.Style.GetWhiteSpace()
		text := []rune(textBox.Text)
		length := len(text)
		if index == length {
			// Starting a the end of the TextBox, no text to see: Continue
			return nil, true
		}
		if whiteSpace == "normal" || whiteSpace == "nowrap" || whiteSpace == "pre-line" {
			for index < length && text[index] == ' ' {
				index += 1
			}
		}
		if index != 0 {
			return &tree.IntList{Value: index}, false
		}
		return nil, false
	}

	if IsLine(box) {
		children := box.Box().Children
		if index == 0 && len(children) == 0 {
			return nil, false
		}
		result, cont := skipFirstWhitespace(children[index], nextSkipStack)
		if cont {
			index += 1
			if index >= len(children) {
				return nil, true
			}
			result, _ = skipFirstWhitespace(children[index], nil)
		}
		if index != 0 || result != nil {
			return &tree.IntList{Value: index, Next: result}, false
		}
		return nil, false
	}

	if skipStack != nil {
		panic(fmt.Sprintf("unexpected skip inside %s", box.Type()))
	}

	return nil, false
}

// Remove in place space characters at the end of a line.
// This also reduces the width of the inline parents of the modified text.
func removeLastWhitespace(context *layoutContext, box Box) {
	var ancestors []Box
	for IsLine(box) {
		ancestors = append(ancestors, box)
		ch := box.Box().Children
		if len(ch) == 0 {
			return
		}
		box = ch[len(ch)-1]
	}
	textBox, ok := box.(*bo.TextBox)
	if ws := box.Box().Style.GetWhiteSpace(); !(ok && (ws == "normal" || ws == "nowrap" || ws == "pre-line")) {
		return
	}
	newText := strings.TrimRight(textBox.Text, " ")
	var spaceWidth pr.Float
	if newText != "" {
		if len(newText) == len(textBox.Text) {
			return
		}
		textBox.Text = newText
		newBox, resume, _ := splitTextBox(context, textBox, nil, 0)
		if newBox == nil || resume != -1 {
			panic(fmt.Sprintf("expected newBox and no resume, got %v, %v", newBox, resume))
		}
		spaceWidth = textBox.Width.V() - newBox.Box().Width.V()
		textBox.Width = newBox.Box().Width

		// RTL line, the trailing space is at the left of the box. We have to
		// translate the box to align the stripped text with the right edge of
		// the box.
		if newBox.PangoLayout.FirstLineDirection%2 != 0 {
			textBox.PositionX -= spaceWidth
			for _, ancestor := range ancestors {
				ancestor.Box().PositionX -= spaceWidth
			}
		}
	} else {
		spaceWidth = textBox.Width.V()
		textBox.Width = pr.Float(0)
		textBox.Text = ""

		// RTL line, the textbox with a trailing space is now empty at the left
		// of the line. We have to translate the line to align it with the right
		// edge of the box.
		line := ancestors[0]
		if line.Box().Style.GetDirection() == "rtl" {
			line.Translate(line, -spaceWidth, 0, true)
		}
	}

	for _, ancestor := range ancestors {
		ancestor.Box().Width = ancestor.Box().Width.V() - spaceWidth
	}

	// TODO: All tabs (U+0009) are rendered as a horizontal shift that
	// lines up the start edge of the next glyph with the next tab stop.
	// Tab stops occur at points that are multiples of 8 times the width
	// of a space (U+0020) rendered in the block"s font from the block"s
	// starting content edge.

	// TODO: If spaces (U+0020) or tabs (U+0009) at the end of a line have
	// "white-space" set to "pre-wrap", UAs may visually collapse them.
}

// Create a box for the ::first-letter selector.
func firstLetterToBox(context *layoutContext, box Box, skipStack *tree.IntList, firstLetterStyle pr.ElementStyle) *tree.IntList {
	if firstLetterStyle == nil || len(box.Box().Children) == 0 {
		return skipStack
	}

	// Some properties must be ignored :in first-letter boxes.
	// https://drafts.csswg.org/selectors-3/#application-in-css
	// At least, position is ignored to avoid layout troubles.
	firstLetterStyle.SetPosition(pr.BoolString{String: "static"})

	firstLetter := ""
	child := box.Box().Children[0]
	var childSkipStack *tree.IntList
	if textBox, ok := child.(*bo.TextBox); ok {
		letterStyle := tree.ComputedFromCascaded(nil, nil, firstLetterStyle, nil, "", "", nil, context)
		if strings.HasSuffix(textBox.ElementTag, "::first-letter") {
			letterBox := bo.NewInlineBox(textBox.ElementTag+"::first-letter", letterStyle, []Box{child})
			box.Box().Children[0] = letterBox
		} else if textBox.Text != "" {
			text := []rune(textBox.Text)
			characterFound := false
			if skipStack != nil {
				childSkipStack = skipStack.Next
				if childSkipStack != nil {
					index := childSkipStack.Value
					text = text[index:]
					skipStack = nil
				}
			}
			for len(text) != 0 {
				nextLetter := text[0]
				isPunc := unicode.In(nextLetter, bo.TableFirstLetter...)
				if !isPunc {
					if characterFound {
						break
					}
					characterFound = true
				}
				firstLetter += string(nextLetter)
				text = text[1:]
			}
			textBox.Text = string(text)
			if strings.TrimLeft(firstLetter, "\n") != "" {
				// "This type of initial letter is similar to an
				// inline-level element if its "float" property is "none",
				// otherwise it is similar to a floated element."
				if firstLetterStyle.GetFloat() == "none" {
					letterBox := bo.NewInlineBox(textBox.ElementTag+"::first-letter", firstLetterStyle, nil)
					textBox_ := bo.NewTextBox(textBox.ElementTag+"::first-letter", letterStyle, firstLetter)
					letterBox.Children = []Box{&textBox_}
					textBox.Children = append([]Box{letterBox}, textBox.Children...)
				} else {
					letterBox := bo.NewBlockBox(textBox.ElementTag+"::first-letter", firstLetterStyle, nil)
					letterBox.FirstLetterStyle = nil
					lineBox := bo.NewLineBox(textBox.ElementTag+"::first-letter", firstLetterStyle, nil)
					letterBox.Children = []Box{&lineBox}
					textBox_ := bo.NewTextBox(textBox.ElementTag+"::first-letter", letterStyle, firstLetter)
					lineBox.Children = []Box{&textBox_}
					textBox.Children = append([]Box{letterBox}, textBox.Children...)
				}
				if skipStack != nil && childSkipStack != nil {
					skipStack = &tree.IntList{Value: skipStack.Value, Next: &tree.IntList{
						Value: childSkipStack.Value + 1,
						Next:  childSkipStack,
					}}
				}
			}
		}
	} else if bo.ParentBoxT.IsInstance(child) {
		if skipStack != nil {
			childSkipStack = skipStack.Next
		} else {
			childSkipStack = nil
		}
		childSkipStack = firstLetterToBox(context, child, childSkipStack, firstLetterStyle)
		if skipStack != nil {
			skipStack = &tree.IntList{Value: skipStack.Value, Next: childSkipStack}
		}
	}
	return skipStack
}

func resolveMarginAuto(box *bo.BoxFields) {
	if box.MarginTop == pr.Auto {
		box.MarginTop = pr.Float(0)
	}
	if box.MarginRight == pr.Auto {
		box.MarginRight = pr.Float(0)
	}
	if box.MarginBottom == pr.Auto {
		box.MarginBottom = pr.Float(0)
	}
	if box.MarginLeft == pr.Auto {
		box.MarginLeft = pr.Float(0)
	}
}

// Compute the width and the height of the atomic ``box``.
func atomicBox(context *layoutContext, box Box, positionX pr.Float, skipStack *tree.IntList, containingBlock *bo.BoxFields,
	absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder) Box {

	if _, ok := box.(bo.ReplacedBoxITF); ok {
		box = box.Copy()
		inlineReplacedBoxLayout(box, containingBlock)
		box.Box().Baseline = box.Box().MarginHeight()
	} else if bo.InlineBlockBoxT.IsInstance(box) {
		if box.Box().IsTableWrapper {
			tableWrapperWidth(context, box, bo.MaybePoint{containingBlock.Width, containingBlock.Height})
		}
		box = inlineBlockBoxLayout(context, box, positionX, skipStack, containingBlock,
			absoluteBoxes, fixedBoxes)
	} else {
		panic(fmt.Sprintf("Layout for %s not handled yet", box))
	}
	return box
}

func inlineBlockBoxLayout(context *layoutContext, box_ Box, positionX pr.Float, skipStack *tree.IntList,
	containingBlock *bo.BoxFields, absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder) Box {

	resolvePercentagesBox(box_, containingBlock, "")
	box := box_.Box()
	// http://www.w3.org/TR/CSS21/visudet.html#inlineblock-width
	if box.MarginLeft == pr.Auto {
		box.MarginLeft = pr.Float(0)
	}
	if box.MarginRight == pr.Auto {
		box.MarginRight = pr.Float(0)
	}
	// http://www.w3.org/TR/CSS21/visudet.html#block-root-margin
	if box.MarginTop == pr.Auto {
		box.MarginTop = pr.Float(0)
	}
	if box.MarginBottom == pr.Auto {
		box.MarginBottom = pr.Float(0)
	}

	inlineBlockWidth(box_, context, containingBlock)
	box.PositionX = positionX
	box.PositionY = 0
	box_, _ = blockContainerLayout(context, box_, pr.Inf, skipStack,
		true, absoluteBoxes, fixedBoxes, new([]pr.Float), false)
	box_.Box().Baseline = inlineBlockBaseline(box_)
	return box_
}

//     Return the y position of the baseline for an inline block
//     from the top of its margin box.
//     http://www.w3.org/TR/CSS21/visudet.html#propdef-vertical-align
func inlineBlockBaseline(box_ Box) pr.Float {
	box := box_.Box()
	if box.IsTableWrapper {
		// Inline table's baseline is its first row's baseline
		for _, child := range box.Children {
			if bo.TableBoxT.IsInstance(child) {
				if cc := child.Box().Children; len(cc) != 0 && len(cc[0].Box().Children) != 0 {
					firstRow := cc[0].Box().Children[0]
					return firstRow.Box().Baseline.V()
				}
			}
		}
	} else if box.Style.GetOverflow() == "visible" {
		result := findInFlowBaseline(box_, true)
		if pr.Is(result) {
			return result.V()
		}
	}
	return box.PositionY + box.MarginHeight()
}

var inlineBlockWidth = handleMinMaxWidth(inlineBlockWidth_)

// @handleMinMaxWidth
func inlineBlockWidth_(box_ Box, context *layoutContext, containingBlock containingBlock) (bool, pr.Float) {
	if box := box_.Box(); box.Width == pr.Auto {
		cbWidth, _ := containingBlock.ContainingBlock()
		availableContentWidth := cbWidth.V() - (box.MarginLeft.V() + box.MarginRight.V() +
			box.BorderLeftWidth.V() + box.BorderRightWidth.V() +
			box.PaddingLeft.V() + box.PaddingRight.V())
		box.Width = shrinkToFit(context, box_, availableContentWidth)
	}
	return false, 0
}

type widths struct {
	left, right pr.Float
}

func (w *widths) add(key pr.String, value pr.Float) {
	switch key {
	case "left":
		w.left += value
	case "right":
		w.right += value
	default:
		panic("unexpected key " + key)
	}
}

type splitedInline struct {
	newBox                  Box
	resumeAt                *tree.IntList
	preservedLineBreak      bool
	firstLetter, lastLetter rune // 0 for none
	floatWidths             widths
}

// Fit as much content as possible from an inline-level box in a width.
//
// Return ``(newBox, resumeAt, preservedLineBreak, firstLetter,
// lastLetter)``. ``resumeAt`` is ``nil`` if all of the content
// fits. Otherwise it can be passed as a ``skipStack`` parameter to resume
// where we left off.
//
// ``newBox`` is non-empty (unless the box is empty) and as big as possible
// while being narrower than ``availableWidth``, if possible (may overflow
// is no split is possible.)
func splitInlineLevel(context *layoutContext, box_ Box, positionX, maxX pr.Float, skipStack *tree.IntList,
	containingBlock *bo.BoxFields, absoluteBoxes, fixedBoxes,
	linePlaceholders *[]*AbsolutePlaceholder, waitingFloats []Box, lineChildren []indexedBox) splitedInline {
	box := box_.Box()
	resolvePercentagesBox(box_, containingBlock, "")
	floatWidths := widths{}
	var (
		newBox                  Box
		preservedLineBreak      bool
		resumeAt                *tree.IntList
		firstLetter, lastLetter rune
	)

	if debugMode {
		debugLogger.LineWithIndent("Split inline level: %T (maxX: %f, positionX: %f)...", box_, maxX, positionX)
	}

	if textBox, ok := box_.(*bo.TextBox); ok {
		textBox.PositionX = positionX
		skip := 0
		if skipStack != nil {
			skip, skipStack = skipStack.Value, skipStack.Next
			if skipStack != nil {
				panic(fmt.Sprintf("expected empty skipStack, got %v", skipStack))
			}
		}
		var newTextBox *bo.TextBox
		newTextBox, skip, preservedLineBreak = splitTextBox(context, textBox, maxX-positionX, skip)
		if skip != -1 {
			resumeAt = &tree.IntList{Value: skip}
		}
		if newTextBox != nil { // we dont want a non nil interface value with a nil pointer
			newBox = newTextBox
		}
		if text := []rune(textBox.Text); len(text) != 0 {
			firstLetter = text[0]
			if skip == -1 {
				lastLetter = text[len(text)-1]
			} else {
				lastLetter = text[skip-1]
			}
		}
	} else if bo.InlineBoxT.IsInstance(box_) {
		if box.MarginLeft == pr.Auto {
			box.MarginLeft = pr.Float(0)
		}
		if box.MarginRight == pr.Auto {
			box.MarginRight = pr.Float(0)
		}
		tmp := splitInlineBox(context, box_, positionX, maxX, skipStack, containingBlock,
			absoluteBoxes, fixedBoxes, linePlaceholders, waitingFloats, lineChildren)
		newBox, resumeAt, preservedLineBreak, firstLetter, lastLetter, floatWidths = tmp.newBox, tmp.resumeAt, tmp.preservedLineBreak, tmp.firstLetter, tmp.lastLetter, tmp.floatWidths
	} else if bo.AtomicInlineLevelBoxT.IsInstance(box_) {
		newBox = atomicBox(context, box_, positionX, skipStack, containingBlock, absoluteBoxes, fixedBoxes)
		newBox.Box().PositionX = positionX
		resumeAt = nil
		preservedLineBreak = false
		// See https://www.w3.org/TR/css-text-3/#line-breaking
		// Atomic inlines behave like ideographic characters.
		firstLetter = '\u2e80'
		lastLetter = '\u2e80'
	} else if bo.InlineFlexBoxT.IsInstance(box_) {
		box.PositionX = positionX
		box.PositionY = 0
		resolveMarginAuto(box)
		var v blockLayout
		newBox, v = flexLayout(context, box_, pr.Inf, skipStack, containingBlock, false, absoluteBoxes, fixedBoxes)
		resumeAt = v.resumeAt
		preservedLineBreak = false
		firstLetter = '\u2e80'
		lastLetter = '\u2e80'
	} else { // pragma: no cover
		log.Fatalf("Layout for %v not handled yet", box)
	}

	if debugMode {
		debugLogger.LineWithDedent("--> split inline level %T done (resumeAt :%s)", box_, resumeAt)
	}

	return splitedInline{
		newBox:             newBox,
		resumeAt:           resumeAt,
		preservedLineBreak: preservedLineBreak,
		firstLetter:        firstLetter,
		lastLetter:         lastLetter,
		floatWidths:        floatWidths,
	}
}

const (
	letterTrue  rune = -1
	letterFalse rune = -2
)

// Same behavior as splitInlineLevel.
// the returned newBox has same concrete type has box_
func splitInlineBox(context *layoutContext, box_ Box, positionX, maxX pr.Float, skipStack *tree.IntList,
	containingBlock *bo.BoxFields, absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder,
	linePlaceholders *[]*AbsolutePlaceholder, waitingFloats []Box, lineChildren []indexedBox) splitedInline {

	if !IsLine(box_) {
		panic(fmt.Sprintf("expected Line or Inline Box, got %T", box_))
	}
	box := box_.Box()

	if debugMode {
		debugLogger.LineWithIndent("Split inline box %T (with width %v)", box_, box.Width)
	}

	// In some cases (shrink-to-fit result being the preferred width)
	// maxX is coming from Pango itself,
	// but floating point errors have accumulated:
	//   width2 = (width + X) - X   // in some cases, width2 < width
	// Increase the value a bit to compensate and not introduce
	// an unexpected line break. The 1e-9 value comes from PEP 485.
	maxX *= 1 + 1e-9

	isStart := skipStack == nil
	initialPositionX := positionX
	initialSkipStack := skipStack

	leftSpacing := box.PaddingLeft.V() + box.MarginLeft.V() + box.BorderLeftWidth.V()
	rightSpacing := box.PaddingRight.V() + box.MarginRight.V() + box.BorderRightWidth.V()
	contentBoxLeft := positionX

	if box.Style.GetPosition().String == "relative" {
		absoluteBoxes = &[]*AbsolutePlaceholder{}
	}

	var skip int
	if !isStart {
		skip, skipStack = skipStack.Value, skipStack.Next
	}
	var (
		i, floatResumeAt          int
		L                         = len(box.Children[skip:])
		resumeAt                  *tree.IntList
		children, waitingChildren []indexedBox
		firstLetter, lastLetter   rune
		preservedLineBreak        = false
		floatWidths               widths
	)
	for ; i < L; i++ {
		index := i + skip
		child_ := box.Children[index]
		child := child_.Box()
		child.PositionY = box.PositionY
		if !child.IsInNormalFlow() {
			inlineOutOfFlowLayout(context, box_, containingBlock, index, child_, children, lineChildren,
				&waitingChildren, &waitingFloats, absoluteBoxes, fixedBoxes, linePlaceholders, &floatWidths, maxX, positionX)
			if child.IsFloated() {
				floatResumeAt = index + 1
			}
			continue
		}

		lastChild := index == len(box.Children)-1
		availableWidth := maxX
		var childWaitingFloats []Box

		if debugMode {
			debugLogger.LineWithIndent("Recursing for inline-level child %T...", child_)
		}
		v := splitInlineLevel(context, child_, positionX, availableWidth, skipStack,
			containingBlock, absoluteBoxes, fixedBoxes, linePlaceholders, childWaitingFloats, lineChildren)
		resumeAt = v.resumeAt
		newChild, preserved, first, last, newFloatWidths := v.newBox, v.preservedLineBreak, v.firstLetter, v.lastLetter, v.floatWidths

		var endSpacing pr.Float
		if box.Style.GetDirection() == "rtl" {
			endSpacing = leftSpacing
			maxX -= newFloatWidths.left
		} else {
			endSpacing = rightSpacing
			maxX -= newFloatWidths.right
		}

		if lastChild && rightSpacing != 0 && resumeAt == nil {
			// TODO: we should take care of children added into absoluteBoxes,
			// fixedBoxes and other lists.
			availableWidth -= endSpacing

			v := splitInlineLevel(context, child_, positionX, availableWidth, skipStack,
				containingBlock, absoluteBoxes, fixedBoxes, linePlaceholders, childWaitingFloats, lineChildren)
			newChild, resumeAt, preserved, first, last, newFloatWidths = v.newBox, v.resumeAt, v.preservedLineBreak, v.firstLetter, v.lastLetter, v.floatWidths
		}

		if debugMode {
			debugLogger.LineWithDedent("--> inline level child %T done (resumeAt : %s).", child_, resumeAt)
		}

		skipStack = nil
		if preserved {
			preservedLineBreak = true
		}

		var canBreak pr.MaybeBool
		if lastLetter == letterTrue {
			lastLetter = ' '
		} else if lastLetter == letterFalse {
			lastLetter = ' ' // no-break space
		} else if box.Style.GetWhiteSpace() == "pre" || box.Style.GetWhiteSpace() == "nowrap" {
			canBreak = pr.False
		}
		if canBreak == nil {
			if lastLetter == 0 || first == 0 {
				canBreak = pr.False
			} else if first == letterTrue {
				canBreak = pr.True
			} else if first == letterFalse {
				canBreak = pr.False
			} else {
				canBreak = text.CanBreakText([]rune{lastLetter, first})
			}
		}

		if canBreak == pr.True {
			children = append(children, waitingChildren...)
			waitingChildren = nil
		}

		if firstLetter == 0 {
			firstLetter = first
		}
		if child.TrailingCollapsibleSpace {
			lastLetter = letterTrue
		} else {
			lastLetter = last
		}

		if asTextBox, ok := newChild.(*bo.TextBox); newChild == nil || (ok && asTextBox == nil) {
			// May be nil where we have an empty TextBox.
		} else {
			if bo.LineBoxT.IsInstance(box_) {
				lineChildren = append(lineChildren, indexedBox{index: index, box: newChild})
			}
			// TODO: we should try to find a better condition here.
			textBox, ok := newChild.(*bo.TextBox)
			trailingWhitespace := ok && strings.TrimSpace(textBox.Text) == ""

			marginWidth := newChild.Box().MarginWidth()
			newPositionX := newChild.Box().PositionX + marginWidth
			if newPositionX > maxX && !trailingWhitespace {
				if len(waitingChildren) != 0 {
					// Too wide, let's try to cut inside waiting children,
					// starting from the end.
					// TODO: we should take care of children added into
					// absoluteBoxes, fixedBoxes and other lists.
					waitingChildrenCopy := append([]indexedBox{}, waitingChildren...)
					breakFound := false
					for len(waitingChildrenCopy) != 0 {
						var tmp indexedBox
						tmp, waitingChildrenCopy = waitingChildrenCopy[len(waitingChildrenCopy)-1], waitingChildrenCopy[:len(waitingChildrenCopy)-1]
						childIndex, child := tmp.index, tmp.box
						// TODO: should we also accept relative children?
						if child.Box().IsInNormalFlow() && canBreakInside(child) == pr.True {
							// We break the waiting child at its last possible
							// breaking point.
							// TODO: The dirty solution chosen here is to
							// decrease the actual size by 1 and render the
							// waiting child again with this constraint. We may
							// find a better way.
							maxX := child.Box().PositionX + child.Box().MarginWidth() - 1
							tmp := splitInlineLevel(context, child, child.Box().PositionX, maxX,
								nil, box, absoluteBoxes, fixedBoxes, linePlaceholders, waitingFloats, lineChildren)
							childNewChild, childResumeAt := tmp.newBox, tmp.resumeAt

							// As PangoLayout and PangoLogAttr don't always
							// agree, we have to rely on the actual split to
							// know whether the child was broken.
							// https://github.com/Kozea/WeasyPrint/issues/614
							breakFound = childResumeAt != nil
							if childResumeAt == nil {
								// PangoLayout decided not to break the child
								childResumeAt = &tree.IntList{Value: 0}
							}
							// TODO: use this when Pango is always 1.40.13+
							// breakFound = true

							children = append(children, waitingChildrenCopy...)
							if childNewChild == nil {
								// May be nil where we have an empty TextBox.
								if !bo.TextBoxT.IsInstance(child) {
									log.Fatalf("only text box may yield empty child, got %s", child)
								}
							} else {
								children = append(children, indexedBox{index: childIndex, box: childNewChild})
							}

							// As this child has already been broken
							// following the original skip stack, we have to
							// add the original skip stack to the partial
							// skip stack we get after the new rendering.
							//
							// Combining skip stacks is a bit complicated
							// We have to:
							// - set `childIndex` as the first number
							// - append the new stack if it's an absolute one
							// - otherwise append the combined stacks
							//   (resumeAt + initialSkipStack)
							//
							// extract the initial index
							var (
								initialIndex     int
								currentSkipStack *tree.IntList
							)
							if initialSkipStack != nil {
								initialIndex, currentSkipStack = initialSkipStack.Value, initialSkipStack.Next
							}
							// childResumeAt is an absolute skip stack
							if childIndex > initialIndex {
								resumeAt = &tree.IntList{Value: childIndex, Next: childResumeAt}
								break
							}

							// combine the stacks
							currentResumeAt := childResumeAt
							var stack []int
							for currentSkipStack != nil && currentResumeAt != nil {
								var resume int
								skip, currentSkipStack = currentSkipStack.Value, currentSkipStack.Next
								resume, currentResumeAt = currentResumeAt.Value, currentResumeAt.Next
								stack = append(stack, skip+resume)
								if resume != 0 {
									break
								}
							}
							resumeAt = currentResumeAt
							for len(stack) != 0 {
								index, stack = stack[len(stack)-1], stack[:len(stack)-1]
								resumeAt = &tree.IntList{Value: index, Next: resumeAt}
							}
							// insert the child index
							resumeAt = &tree.IntList{Value: childIndex, Next: resumeAt}
							break
						}
					}
					if breakFound {
						break
					}
				}
				if l := len(children); l != 0 {
					// Too wide, can't break waiting children and the inline is
					// non-empty: put child entirely on the next line.
					resumeAt = &tree.IntList{Value: children[l-1].index + 1}
					childWaitingFloats = nil
					break
				}
			}

			positionX = newPositionX
			waitingChildren = append(waitingChildren, indexedBox{index: index, box: newChild})
		}
		waitingFloats = append(waitingFloats, childWaitingFloats...)
		if resumeAt != nil {
			children = append(children, waitingChildren...)
			resumeAt = &tree.IntList{Value: index, Next: resumeAt}
			break
		}
	}
	if i == L {
		children = append(children, waitingChildren...)
		resumeAt = nil
	}

	// Reorder inline blocks when direction is rtl
	if box.Style.GetDirection() == "rtl" && len(children) > 1 {
		var inFlowChildren []Box
		for _, child := range children {
			if child.box.Box().IsInNormalFlow() {
				inFlowChildren = append(inFlowChildren, child.box)
			}
		}
		posX := inFlowChildren[0].Box().PositionX
		for _, child := range reverseBoxes(inFlowChildren) {
			child.Translate(child, (posX - child.Box().PositionX), 0, true)
			posX += child.Box().MarginWidth()
		}
	}

	isEnd := resumeAt == nil
	toCopy := make([]Box, len(children))
	for i, boxChild := range children {
		toCopy[i] = boxChild.box
	}

	newBox_ := bo.CopyWithChildren(box_, toCopy)
	newBox := newBox_.Box()
	newBox_.RemoveDecoration(newBox, !isStart, !isEnd)

	if bo.LineBoxT.IsInstance(box_) {
		// We must reset line box width according to its new children
		newBox.Width = pr.Float(0)
		children := newBox.Children
		if newBox.Style.GetDirection() == "ltr" {
			children = reverseBoxes(children)
		}
		for _, boxChild := range children {
			if boxChild.Box().IsInNormalFlow() {
				newBox.Width = boxChild.Box().PositionX + boxChild.Box().MarginWidth() - newBox.PositionX
				break
			}
		}
	} else {
		newBox.PositionX = initialPositionX
		var translationNeeded bool
		if box.Style.GetBoxDecorationBreak() == "clone" {
			translationNeeded = true
		} else if box.Style.GetDirection() == "ltr" {
			translationNeeded = isStart
		} else {
			translationNeeded = isEnd
		}
		if translationNeeded {
			for _, child := range newBox.Children {
				child.Translate(child, leftSpacing, 0, false)
			}
		}
		newBox.Width = positionX - contentBoxLeft
		newBox_.Translate(newBox_, floatWidths.left, 0, true)
	}
	stl := text.StrutLayout(box.Style, context)
	lineHeight := stl[0]
	newBox.Baseline = stl[1]
	newBox.Height = box.Style.GetFontSize().ToMaybeFloat()
	halfLeading := (lineHeight - newBox.Height.V()) / 2.
	// Set margins to the half leading but also compensate for borders and
	// paddings. We want marginHeight() == lineHeight
	newBox.MarginTop = halfLeading - newBox.BorderTopWidth.V() - newBox.PaddingTop.V()
	newBox.MarginBottom = halfLeading - newBox.BorderBottomWidth.V() - newBox.PaddingBottom.V()

	if newBox.Style.GetPosition().String == "relative" {
		for _, absoluteBox := range *absoluteBoxes {
			absoluteLayout(context, absoluteBox, newBox_, fixedBoxes)
		}
	}

	if resumeAt != nil {
		if resumeAt.Value < floatResumeAt {
			resumeAt = &tree.IntList{Value: floatResumeAt}
		}
	}

	if box.IsLeader {
		firstLetter = letterTrue
		lastLetter = letterFalse
	}

	return splitedInline{
		newBox:             newBox_,
		resumeAt:           resumeAt,
		preservedLineBreak: preservedLineBreak,
		firstLetter:        firstLetter,
		lastLetter:         lastLetter,
		floatWidths:        floatWidths,
	}
}

func (context *layoutContext) addRunning(child_ bo.Box) {
	runningName := child_.Box().Style.GetPosition().String
	currentRE, has := context.runningElements[runningName]
	if !has {
		currentRE = map[int][]Box{}
		context.runningElements[runningName] = currentRE
	}
	currentRE[context.currentPage] = append(currentRE[context.currentPage], child_)
}

func inlineOutOfFlowLayout(context *layoutContext, box Box, containingBlock *bo.BoxFields, index int, child_ Box, children,
	lineChildren []indexedBox, waitingChildren *[]indexedBox, waitingFloats *[]Box,
	absoluteBoxes, fixedBoxes, linePlaceholders *[]*AbsolutePlaceholder, floatWidths *widths,
	maxX, positionX pr.Float) {
	child := child_.Box()
	if child.IsAbsolutelyPositioned() {
		child.PositionX = positionX
		placeholder := NewAbsolutePlaceholder(child_)
		*linePlaceholders = append(*linePlaceholders, placeholder)
		*waitingChildren = append(*waitingChildren, indexedBox{index: index, box: placeholder})
		if child.Style.GetPosition().String == "absolute" {
			*absoluteBoxes = append(*absoluteBoxes, placeholder)
		} else {
			*fixedBoxes = append(*fixedBoxes, placeholder)
		}
	} else if child.IsFloated() {
		child.PositionX = positionX
		floatWidth := shrinkToFit(context, child_, containingBlock.Width.V())

		// To retrieve the real available space for floats, we must remove
		// the trailing whitespaces from the line
		var nonFloatingChildren []Box
		for _, v := range append(children, *waitingChildren...) {
			if !v.box.Box().IsFloated() {
				nonFloatingChildren = append(nonFloatingChildren, v.box)
			}
		}
		if Lf := len(nonFloatingChildren); Lf != 0 {
			floatWidth -= trailingWhitespaceSize(context, nonFloatingChildren[Lf-1])
		}

		if floatWidth > maxX-positionX || len(*waitingFloats) != 0 {
			// TODO: the absolute and fixed boxes in the floats must be
			// added here, and not in iterLineBoxes
			*waitingFloats = append(*waitingFloats, child_)
		} else {
			if debugMode {
				debugLogger.LineWithIndent("Recursing for float child %T...", child_)
			}

			child_ = floatLayout(context, child_, containingBlock, absoluteBoxes, fixedBoxes)

			if debugMode {
				debugLogger.LineWithDedent("-->Float child %T done.", child_)
			}

			*waitingChildren = append(*waitingChildren, indexedBox{index: index, box: child_})

			// Translate previous line children
			dx := pr.Max(child.MarginWidth(), 0)
			floatWidths.add(child.Style.GetFloat(), dx)
			if child.Style.GetFloat() == "left" {
				if bo.LineBoxT.IsInstance(box) {
					// The parent is the line, update the current position
					// for the next child. When the parent is not the line
					// (it is an inline block), the current position of the
					// line is updated by the box itself (see next
					// splitInlineLevel call).
					positionX += dx
				}
			} else if child.Style.GetFloat() == "right" {
				// Update the maximum x position for the next children
				maxX -= dx
			}
			for _, oldChild := range lineChildren {
				if !oldChild.box.Box().IsInNormalFlow() {
					continue
				}
				if (child.Style.GetFloat() == "left" && box.Box().Style.GetDirection() == "ltr") ||
					(child.Style.GetFloat() == "right" && box.Box().Style.GetDirection() == "rtl") {
					oldChild.box.Translate(oldChild.box, dx, 0, true)
				}
			}
		}
	} else if child.IsRunning() {
		context.addRunning(child_)
	}
}

// See http://unicode.org/reports/tr14/
// \r is already handled by processWhitespace
var lineBreaks = utils.NewSet("\n", "\t", "\f", "\u0085", "\u2028", "\u2029")

// Keep as much text as possible from a TextBox in a limited width.
//
// Try not to overflow but always have some text in ``new_box``
//
// Return ``(new_box, skip, preserved_line_break)``. ``skip`` is the number of
// UTF-8 bytes to skip form the start of the TextBox for the next line, or
// -1 if all of the text fits.
//
// Also break on preserved line breaks.
func splitTextBox(context *layoutContext, box *bo.TextBox, availableWidth pr.MaybeFloat, skip int) (*bo.TextBox, int, bool) {
	fontSize := box.Style.GetFontSize()
	text_ := []rune(box.Text)[skip:]
	if fontSize == pr.FToV(0) || len(text_) == 0 {
		return nil, -1, false
	}
	v := text.SplitFirstLine(string(text_), box.Style, context, availableWidth, box.JustificationSpacing, false)
	layout, length, resumeIndex, width, height, baseline := v.Layout, v.Length, v.ResumeAt, v.Width, v.Height, v.Baseline
	if resumeIndex == 0 {
		panic("resumeAt should not be 0 here")
	}
	newText := layout.Layout.Text
	if length > 0 {
		box = box.CopyWithText(string(newText))
		box.Width = width
		box.PangoLayout = layout
		// "The height of the content area should be based on the font,
		//  but this specification does not specify how."
		// http://www.w3.org/TR/CSS21/visudet.html#inline-non-replaced
		// We trust Pango and use the height of the LayoutLine.
		box.Height = height
		// "only the "line-height" is used when calculating the height
		//  of the line box."
		// Set margins so that marginHeight() == lineHeight
		lineHeight := text.StrutLayout(box.Style, context)[0]
		halfLeading := (lineHeight - height) / 2.
		box.MarginTop = halfLeading
		box.MarginBottom = halfLeading
		// form the top of the content box
		box.Baseline = baseline + box.MarginTop.V() // form the top of the margin box

	} else {
		box = nil
	}

	preservedLineBreak := false
	if resumeIndex != -1 {
		if resumeIndex > len(text_) {
			resumeIndex = len(text_)
		}
		between := string(text_[length:resumeIndex])
		preservedLineBreak = (length != resumeIndex) && len(strings.Trim(between, " ")) != 0
		if preservedLineBreak {
			if !lineBreaks.Has(between) {
				panic(fmt.Sprintf("Got %s between two lines. Expected nothing or a preserved line break", between))
			}
		}
		resumeIndex += skip
	}

	return box, resumeIndex, preservedLineBreak
}

type boxMinMax struct {
	box      Box
	max, min pr.MaybeFloat
}

// Handle ``vertical-align`` within an :class:`LineBox` (or of a
// non-align sub-tree).
// Place all boxes vertically assuming that the baseline of ``box``
// is at `y = 0`.
// Return ``(maxY, minY)``, the maximum and minimum vertical position
// of margin boxes.
func lineBoxVerticality(context *layoutContext, box Box) (pr.Float, pr.Float) {
	var topBottomSubtrees []Box
	maxY, minY := alignedSubtreeVerticality(context, box, &topBottomSubtrees, 0)
	subtreesWithMinMax := make([]boxMinMax, 0, len(topBottomSubtrees))
	for i := 0; i < len(topBottomSubtrees); i++ { // note that topBottomSubtrees may grow over this loop
		subtree := topBottomSubtrees[i]
		var subMaxY, subMinY pr.MaybeFloat
		if !subtree.Box().IsFloated() {
			subMaxY, subMinY = alignedSubtreeVerticality(context, subtree, &topBottomSubtrees, 0)
		}
		subtreesWithMinMax = append(subtreesWithMinMax, boxMinMax{box: subtree, max: subMaxY, min: subMinY})
	}

	if len(subtreesWithMinMax) != 0 {
		var highestSub pr.Float
		for _, subtree := range subtreesWithMinMax {
			if !subtree.box.Box().IsFloated() {
				m := subtree.max.V() - subtree.min.V()
				if m > highestSub {
					highestSub = m
				}
			}
		}
		maxY = pr.Max(maxY, minY+highestSub)
	}

	for _, v := range subtreesWithMinMax {
		va := v.box.Box().Style.GetVerticalAlign()
		var dy pr.Float
		if v.box.Box().IsFloated() {
			dy = minY - v.box.Box().PositionY
		} else if va.String == "top" {
			dy = minY - v.min.V()
		} else if va.String == "bottom" {
			dy = maxY - v.max.V()
		} else {
			panic(fmt.Sprintf("expected top or bottom, got %v", va))
		}
		translateSubtree(v.box, dy)
	}
	return maxY, minY
}

func translateSubtree(box Box, dy pr.Float) {
	if bo.InlineBoxT.IsInstance(box) {
		box.Box().PositionY += dy
		if va := box.Box().Style.GetVerticalAlign().String; va == "top" || va == "bottom" {
			for _, child := range box.Box().Children {
				translateSubtree(child, dy)
			}
		}
	} else {
		// Text or atomic boxes
		box.Translate(box, 0, dy, true)
	}
}

func alignedSubtreeVerticality(context *layoutContext, box Box, topBottomSubtrees *[]Box, baselineY pr.Float) (pr.Float, pr.Float) {
	maxY, minY := inlineBoxVerticality(context, box, topBottomSubtrees, baselineY)
	// Account for the line box itself :
	top := baselineY - box.Box().Baseline.V()
	bottom := top + box.Box().MarginHeight()
	if minY == nil || top < minY.V() {
		minY = top
	}
	if maxY == nil || bottom > maxY.V() {
		maxY = bottom
	}

	return maxY.V(), minY.V()
}

// Handle ``vertical-align`` within an :class:`InlineBox`.
//     Place all boxes vertically assuming that the baseline of ``box``
//     is at `y = baselineY`.
//     Return ``(maxY, minY)``, the maximum and minimum vertical position
//     of margin boxes.
func inlineBoxVerticality(context *layoutContext, box_ Box, topBottomSubtrees *[]Box, baselineY pr.Float) (maxY, minY pr.MaybeFloat) {
	if !IsLine(box_) {
		return maxY, minY
	}
	box := box_.Box()
	for _, child_ := range box_.Box().Children {
		child := child_.Box()
		if !child.IsInNormalFlow() {
			if child.IsFloated() {
				*topBottomSubtrees = append(*topBottomSubtrees, child_)
			}
			continue
		}
		var childBaselineY pr.Float
		verticalAlign := child.Style.GetVerticalAlign()
		switch verticalAlign.String {
		case "baseline":
			childBaselineY = baselineY
		case "middle":
			oneEx := box.Style.GetFontSize().Value * text.ExRatio(box.Style, context)
			top := baselineY - (oneEx+child.MarginHeight())/2.
			childBaselineY = top + child.Baseline.V()
		case "text-top":
			// align top with the top of the parent’s content area
			top := baselineY - box.Baseline.V() + box.MarginTop.V() +
				box.BorderTopWidth.V() + box.PaddingTop.V()
			childBaselineY = top + child.Baseline.V()
		case "text-bottom":
			// align bottom with the bottom of the parent’s content area
			bottom := baselineY - box.Baseline.V() + box.MarginTop.V() +
				box.BorderTopWidth.V() + box.PaddingTop.V() + box.Height.V()
			childBaselineY = bottom - child.MarginHeight() + child.Baseline.V()
		case "top", "bottom":
			// TODO: actually implement vertical-align: top and bottom
			// Later, we will assume for this subtree that its baseline
			// is at y=0.
			childBaselineY = 0
		default:
			// Numeric value: The child’s baseline is `verticalAlign` above
			// (lower y) the parent’s baseline.
			childBaselineY = baselineY - verticalAlign.Value
		}

		// the child’s `top` is `child.Baseline` above (lower y) its baseline.
		top := childBaselineY - child.Baseline.V()
		if bo.InlineBlockBoxT.IsInstance(child_) || bo.InlineFlexBoxT.IsInstance(child_) {
			// This also includes table wrappers for inline tables.
			child_.Translate(child_, 0, top-child.PositionY, false)
		} else {
			child.PositionY = top
			// grand-children for inline boxes are handled below
		}

		if verticalAlign.String == "top" || verticalAlign.String == "bottom" {
			// top or bottom are special, they need to be handled in
			// a later pass.
			*topBottomSubtrees = append(*topBottomSubtrees, child_)
			continue
		}

		bottom := top + child.MarginHeight()
		if minY == nil || top < minY.V() {
			minY = top
		}
		if maxY == nil || bottom > maxY.V() {
			maxY = bottom
		}
		if bo.InlineBoxT.IsInstance(child_) {
			childrenMaxY, childrenMinY := inlineBoxVerticality(context, child_, topBottomSubtrees, childBaselineY)
			if childrenMinY != nil && childrenMinY.V() < minY.V() {
				minY = childrenMinY
			}
			if childrenMaxY != nil && childrenMaxY.V() > maxY.V() {
				maxY = childrenMaxY
			}
		}
	}
	return maxY, minY
}

// Return how much the line should be moved horizontally according to
// the `text-align` property.
func textAlign(context *layoutContext, line_ Box, availableWidth pr.Float, last bool) pr.Float {
	line := line_.Box()

	// "When the total width of the inline-level boxes on a line is less than
	// the width of the line box containing them, their horizontal distribution
	// within the line box is determined by the "text-align" property."
	if line.Width.V() >= availableWidth {
		return 0
	}

	align := line.Style.GetTextAlignAll()
	if last {
		alignLast := line.Style.GetTextAlignLast()
		if alignLast != "auto" {
			align = alignLast
		}
	}
	ws := line.Style.GetWhiteSpace()
	spaceCollapse := ws == "normal" || ws == "nowrap" || ws == "pre-line"

	if align == "left" || align == "right" {
		if (align == "left") != (line.Style.GetDirection() == "rtl") { // xor
			align = "start"
		} else {
			align = "end"
		}
	}

	if align == "start" {
		return 0
	}

	offset := availableWidth - line.Width.V()
	if align == "justify" {
		if spaceCollapse {
			// Justification of texts where white space is not collapsing is
			// - forbidden by CSS 2, and
			// - not required by CSS 3 Text.
			justifyLine(context, line_, offset)
		}
		return 0
	}
	if align == "center" {
		return offset / 2
	} else if align == "end" {
		return offset
	} else {
		log.Printf("align should be center or right, got %s", align)
		return 0
	}
}

func justifyLine(context *layoutContext, line Box, extraWidth pr.Float) {
	// TODO: We should use a better alorithm here, see
	// https://www.w3.org/TR/css-text-3/#justify-algos
	nbSpaces := countSpaces(line)
	if nbSpaces == 0 {
		return
	}
	addWordSpacing(context, line, extraWidth/pr.Float(nbSpaces), 0)
}

func countSpaces(box Box) int {
	if textBox, isTextBox := box.(*bo.TextBox); isTextBox {
		// TODO: remove trailing spaces correctly
		return strings.Count(textBox.Text, " ")
	} else if IsLine(box) {
		var sum int
		for _, child := range box.Box().Children {
			sum += countSpaces(child)
		}
		return sum
	} else {
		return 0
	}
}

func addWordSpacing(context *layoutContext, box_ Box, justificationSpacing, xAdvance pr.Float) pr.Float {
	if textBox, isTextBox := box_.(*bo.TextBox); isTextBox {
		textBox.JustificationSpacing = justificationSpacing
		textBox.PositionX += xAdvance
		nbSpaces := pr.Float(countSpaces(box_))
		if nbSpaces > 0 {
			layout := text.CreateLayout(textBox.Text, textBox.Style, context, pr.Inf, textBox.JustificationSpacing)
			// layout.Deactivate()
			extraSpace := justificationSpacing * nbSpaces
			xAdvance += extraSpace
			textBox.Width = textBox.Width.V() + extraSpace
			textBox.PangoLayout = layout
		}
	} else if IsLine(box_) {
		box := box_.Box()
		box.PositionX += xAdvance
		previousXAdvance := xAdvance
		for _, child := range box.Children {
			if child.Box().IsInNormalFlow() {
				xAdvance = addWordSpacing(context, child, justificationSpacing, xAdvance)
			}
		}
		box.Width = box.Width.V() + xAdvance - previousXAdvance
	} else {
		// Atomic inline-level box
		box_.Translate(box_, xAdvance, 0, false)
	}
	return xAdvance
}

// http://www.w3.org/TR/CSS21/visuren.html#phantom-line-box
func isPhantomLinebox(linebox bo.BoxFields) bool {
	for _, child_ := range linebox.Children {
		child := *child_.Box()
		if bo.InlineBoxT.IsInstance(child_) {
			if !isPhantomLinebox(child) {
				return false
			}
			for _, side := range [4]string{"top", "right", "bottom", "left"} {
				m := child.Style.Get("margin_" + side).(pr.Value).Value
				b := child.Style.Get("border_" + side + "_width").(pr.Value)
				p := child.Style.Get("padding_" + side).(pr.Value).Value
				if m != 0 || !b.IsNone() || p != 0 {
					return false
				}
			}
		} else if child.IsInNormalFlow() {
			return false
		}
	}
	return true
}

func canBreakInside(box Box) pr.MaybeBool {
	// See https://www.w3.org/TR/css-text-3/#white-space-property
	ws := box.Box().Style.GetWhiteSpace()
	textWrap := ws == "normal" || ws == "pre-wrap" || ws == "pre-line"
	textBox, isTextBox := box.(*bo.TextBox)
	if bo.AtomicInlineLevelBoxT.IsInstance(box) {
		return pr.False
	} else if textWrap && isTextBox {
		return text.CanBreakText([]rune(textBox.Text))
	} else if textWrap && bo.ParentBoxT.IsInstance(box) {
		for _, child := range box.Box().Children {
			if canBreakInside(child) == pr.True {
				return pr.True
			}
		}
		return pr.False
	}
	return pr.False
}
