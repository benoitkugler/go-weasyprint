package layout

import (
	"log"
	"strings"
	"unicode"

	"github.com/benoitkugler/go-weasyprint/boxes"
	bo "github.com/benoitkugler/go-weasyprint/boxes"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/style/tree"
	"github.com/benoitkugler/go-weasyprint/utils"
)

//     Line breaking and layout for inline-level boxes.

// IsLineBox || IsInlineBox
func isLine(box Box) bool {
	return bo.TypeLineBox.IsInstance(box) || bo.TypeInlineBox.IsInstance(box)
}

type lineBoxe struct {
	// laid-out LineBox with as much content as possible that
	// fits in the available width.
	line     *AbsolutePlaceholder
	resumeAt *tree.SkipStack
}

type lineBoxeIterator struct {
	boxes []lineBoxe
	index int
}

func (l lineBoxeIterator) Has() bool {
	return l.index < len(l.boxes)
}

func (l *lineBoxeIterator) Next() lineBoxe {
	b := l.boxes[l.index]
	l.index += 1
	return b
}

// `box` is a non-laid-out `LineBox`
// positionY is the vertical top position of the line box on the page
// skipStack is ``None`` to start at the beginning of ``linebox``,
// or a ``resumeAt`` value to continue just after an
// already laid-out line.
func iterLineBoxes(context *LayoutContext, box *bo.LineBox, positionY pr.Float, skipStack *tree.SkipStack, containingBlock Box,
	absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder, firstLetterStyle pr.Properties) lineBoxeIterator {
	resolvePercentages(box, bo.MaybePoint{containingBlock.Width, containingBlock.Height}, "")
	if skipStack == nil {
		// TODO: wrong, see https://github.com/Kozea/WeasyPrint/issues/679
		box.TextIndent = resolveOnePercentage(pr.MaybeFloatToValue(box.TextIndent), "textIndent", containingBlock.Width, "")
	} else {
		box.TextIndent = pr.Float(0)
	}
	var out []lineBoxe
	for {
		tmp := getNextLinebox(context, box, positionY, skipStack, containingBlock,
			absoluteBoxes, fixedBoxes, firstLetterStyle)
		line, resumeAt := tmp.line, tmp.resumeAt
		if line != nil {
			positionY = line.Box().PositionY + line.Box().Height.V()
		}
		if line == nil {
			return out
		}
		out = append(out, lineBoxe{line: line, resumeAt: resumeAt})
		if resumeAt == nil {
			return out
		}
		skipStack = resumeAt
		box.TextIndent = pr.Float(0)
		firstLetterStyle = nil
	}
}

func getNextLinebox(context *LayoutContext, linebox *bo.LineBox, positionY pr.Float, skipStack *tree.SkipStack,
	containingBlock Box, absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder,
	firstLetterStyle pr.Properties) lineBoxe {

	skipStack, cont := skipFirstWhitespace(linebox, skipStack)
	if cont {
		return lineBoxe{}
	}

	skipStack = firstLetterToBox(linebox, skipStack, firstLetterStyle)

	linebox.PositionY = positionY

	if len(context.excludedShapes) != 0 {
		// Width && height must be calculated to avoid floats
		linebox.Width = pr.Float(inlineMinContentWidth(*context, linebox, true, skipStack, true, false))
		linebox.Height, _ = StrutLayout(linebox.Style, context)
	} else {
		// No float, width && height will be set by the lines
		linebox.Width = pr.Float(0)
		linebox.Height = pr.Float(0)
	}
	positionX, positionY, availableWidth := avoidCollisions(context, linebox, containingBlock, false)

	candidateHeight := linebox.Height

	excludedShapes := make([]shape, len(context.excludedShapes))
	for i, v := range context.excludedShapes {
		excludedShapes[i] = v
	}

	var (
		line_                                      Box
		linePlaceholders, lineAbsolutes, lineFixed []*AbsolutePlaceholder
		waitingFloats                              []Box
	)
	for {
		linebox.PositionX = positionX
		linebox.PositionY = positionY
		maxX := positionX + availableWidth
		positionX += linebox.TextIndent

		var (
			resumeAt           *tree.SkipStack
			preservedLineBreak bool
			floatWidths        widths
		)

		spi := splitInlineBox(context, linebox, positionX, maxX, skipStack, containingBlock,
			lineAbsolutes, lineFixed, linePlaceholders, waitingFloats, nil)

		line_, resumeAt, preservedLineBreak, floatWidths = spi.newBox, spi.resumeAt, spi.preservedLineBreak, spi.floatWidths
		line := line_.Box()
		linebox.Width, linebox.Height = line.Width, line.Height

		if isPhantomLinebox(*line) && !preservedLineBreak {
			line.Height = pr.Float(0)
			break
		}

		removeLastWhitespace(*context, line_)

		newPositionX, _, newAvailableWidth := avoidCollisions(context, linebox, containingBlock, false)
		// TODO: handle rtl
		newAvailableWidth -= floatWidths.right
		alignmentAvailableWidth := newAvailableWidth + newPositionX - linebox.PositionX
		offsetX := textAlign(*context, line_, alignmentAvailableWidth, resumeAt == nil || preservedLineBreak)

		bottom_, top_ := lineBoxVerticality(line_)
		bottom, top := bottom_.(pr.Float), top_.(pr.Float)
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
		context.excludedShapes = excludedShapes
		positionX, positionY, availableWidth = avoidCollisions(context, line, containingBlock, false)
		if positionX == linebox.PositionX && positionY == linebox.PositionY {
			context.excludedShapes = newExcludedShapes
			break
		}
	}
	*absoluteBoxes = append(*absoluteBoxes, lineAbsolutes...)
	*fixedBoxes = append(*fixedBoxes, lineFixed...)

	line := line_.Box()
	for _, placeholder_ := range linePlaceholders {
		placeholder := placeholder_.Box
		if strings.HasPrefix(string(placeholder.Box().Style.GetWeasySpecifiedDisplay()), "inline") {
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

	return lineBoxe{line: line, resumeAt: resumeAt}
}

// Return the ``skipStack`` to start just after the remove spaces
//     at the beginning of the line.
//     See http://www.w3.org/TR/CSS21/text.html#white-space-model
func skipFirstWhitespace(box Box, skipStack *tree.SkipStack) (ss *tree.SkipStack, continue_ bool) {
	var (
		index         int
		nextSkipStack *tree.SkipStack
	)
	if skipStack != nil {
		index, nextSkipStack = skipStack.Skip, skipStack.Stack
	}

	if textBox, ok := box.(*bo.TextBox); ok {
		if nextSkipStack != nil {
			log.Fatalf("expected nil nextSkipStack, got %v", nextSkipStack)
		}
		whiteSpace := textBox.Style.GetWhitespace()
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
			return &tree.SkipStack{Skip: index}, false
		}
		return nil, false
	}

	if isLine(box) {
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
			result = skipFirstWhitespace(children[index], nil)
		}
		if index != 0 || result != nil {
			return &tree.SkipStack{Skip: index, Stack: result}, false
		}
		return nil, false
	}
	if skipStack != nil {
		log.Fatalf("unexpected skip inside %s", box)
	}
	return nil, false
}

// Remove in place space characters at the end of a line.
// This also reduces the width of the inline parents of the modified text.
func removeLastWhitespace(context LayoutContext, box Box) {
	var ancestors []Box
	for isLine(box) {
		ancestors = append(ancestors, box)
		ch := box.Box().Children
		if len(ch) == 0 {
			return
		}
		box = ch[len(ch)-1]
	}
	textBox, ok := box.(*bo.TextBox)
	if ws := textBox.Style.GetWhiteSpace(); !(ok && (ws == "normal" || ws == "nowrap" || ws == "pre-line")) {
		return
	}
	newText := strings.TrimRight(textBox.Text, " ")
	var spaceWidth pr.Float
	if newText != "" {
		if len(newText) == len(textBox.Text) {
			return
		}
		textBox.Text = newText
		newBox, resume, _ := splitTextBox(context, textBox, None, 0)
		if newBox == nil || resume != nil {
			log.Fatalf("expected newBox and no resume, got %v, %v", newBox, resume)
		}
		spaceWidth = textBox.Width.V() - newBox.Box().Width.V()
		textBox.Width = newBox.Box().Width
	} else {
		spaceWidth = textBox.Width
		textBox.Width = 0
		textBox.Text = ""
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
func firstLetterToBox(box Box, skipStack *tree.SkipStack, firstLetterStyle pr.Properties) *tree.SkipStack {
	if len(firstLetterStyle) != 0 && len(box.Box().Children) != 0 {
		// Some properties must be ignored :in first-letter boxes.
		// https://drafts.csswg.org/selectors-3/#application-in-css
		// At least, position is ignored to avoid layout troubles.
		firstLetterStyle.SetPosition(pr.String("static"))
	}

	firstLetter := ""
	child := box.Box().children[0]
	if textBox, ok := child.(*bo.TextBox); ok {
		letterStyle := tree.ComputedFromCascaded(nil, firstLetterStyle, nil)
		if strings.HasSuffix(textBox.ElementTag, "::first-letter") {
			letterBox := bo.NewInlineBox(textBox.ElementTag+"::first-letter", letterStyle, []Box{child})
			box.Box().Children[0] = letterBox
		} else if textBox.Text != "" {
			text := []rune(textBox.Text)
			characterFound = false
			if skipStack != nil {
				childSkipStack := skipStack.Stack
				if childSkipStack != nil {
					index = childSkipStack.Skip
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
					textBox.Children = append([]Box{&letterBox}, textBox.Children...)
				} else {
					letterBox := bo.NewBlockBox(textBox.ElementTag+"::first-letter", firstLetterStyle, nil)
					letterBox.FirstLetterStyle = nil
					lineBox := bo.NewLineBox(textBox.ElementTag+"::first-letter", firstLetterStyle, nil)
					letterBox.Children = []Box{&lineBox}
					textBox_ := bo.NewTextBox(textBox.ElementTag+"::first-letter", letterStyle, firstLetter)
					lineBox.Children = []Box{&textBox_}
					textBox.Children = append([]Box{&letterBox}, textBox.Children...)
				}
				if skipStack != nil && childSkipStack != nil {
					skipStack = tree.SkipStack{Skip: skipStack.Skip, Stack: &tree.SkipStack{
						Skip:  childSkipStack.Skip + 1,
						Stack: childSkipStack,
					}}
				}
			}
		}
	} else if bo.IsParentBox(child) {
		if skipStack != nil {
			childSkipStack = skipStack.Stack
		} else {
			childSkipStack = nil
		}
		childSkipStack = firstLetterToBox(child, childSkipStack, firstLetterStyle)
		if skipStack != nil {
			skipStack = &tree.SkipStack{Skip: skipStack.Skip, Stack: childSkipStack}
		}
	}
	return skipStack
}

var replacedBoxWidth = handleMinMaxWidth(replacedBoxWidth_)

// @handleMinMaxWidth
// Compute and set the used width for replaced boxes (inline- or block-level)
// containingBlock must be block
func replacedBoxWidth_(box_ Box, _ *LayoutContext, containingBlock containingBlock) (bool, pr.Float) {
	box, ok := box_.(bo.InstanceReplacedBox)
	if !ok {
		log.Fatalf("expected ReplacedBox instance, got %s", box_)
	}
	intrinsicWidth, intrinsicHeight := box.Replacement.GetIntrinsicSize(box.Style.GetImageResolution(), box.Style.GetFontSize())

	// This algorithm simply follows the different points of the specification
	// http://www.w3.org/TR/CSS21/visudet.html#inline-replaced-width
	if box.Height == Auto && box.Width == Auto {
		if intrinsicWidth != None {
			// Point #1
			box.Width = intrinsicWidth
		} else if box.Replacement.IntrinsicRatio() != None {
			if intrinsicHeight != None {
				// Point #2 first part
				box.Width = intrinsicHeight * box.Replacement.IntrinsicRatio()
			} else {
				// Point #3
				blockLevelWidth(box, containingBlock)
			}
		}
	}

	if box.Width == Auto {
		if box.Replacement.IntrinsicRatio != None {
			// Point #2 second part
			box.Width = box.Height * box.Replacement.IntrinsicRatio()
		} else if intrinsicWidth != None {
			// Point #4
			box.Width = intrinsicWidth
		} else {
			// Point #5
			// It's pretty useless to rely on device size to set width.
			box.Width = 300
		}
	}
	return false, 0
}

var replacedBoxHeight = handleMinMaxHeight(replacedBoxHeight_)

// @handleMinMaxHeight
//
//     Compute and set the used height for replaced boxes (inline- or block-level)
func replacedBoxHeight_(box_ Box, _ *LayoutContext, _ containingBlock) (bool, pr.Float) {
	box, ok := box_.(bo.InstanceReplacedBox)
	if !ok {
		log.Fatalf("expected ReplacedBox instance, got %s", box_)
	}

	// http://www.w3.org/TR/CSS21/visudet.html#inline-replaced-height
	intrinsicWidth, intrinsicHeight := box.Replacement.GetIntrinsicSize(
		box.Style.GetImageResolution(), box.Style.GetFontSize())
	intrinsicRatio := box.Replacement.IntrinsicRatio()

	// Test Auto on the computed width, not the used width
	if box.Height == Auto && box.Width == Auto {
		box.Height = intrinsicHeight
	} else if box.Height == Auto && intrinsicRatio != 0 {
		box.Height = box.Width / intrinsicRatio
	}

	if box.Height == Auto && box.Width == Auto && intrinsicHeight != None {
		box.Height = intrinsicHeight
	} else if intrinsicRatio != None && box.Height == Auto {
		box.Height = box.Width / intrinsicRatio
	} else if box.Height == Auto && intrinsicHeight != None {
		box.Height = intrinsicHeight
	} else if box.Height == Auto {
		// It"s pretty useless to rely on device size to set width.
		box.Height = 150
	}
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

// Lay out an inline :class:`boxes.ReplacedBox` ``box``.
func inlineReplacedBoxLayout(box_ Box, containingBlock block) {
	resolveMarginAuto(box_.Box())
	inlineReplacedBoxWidthHeight(box, containingBlock)
}

func inlineReplacedBoxWidthHeight(box Box, containingBlock block) {
	if style := box.Box().Style; style.GetWidth().String == "auto" && style.GetHeight().String == "auto" {
		replacedBoxWidth.withoutMinMax(box, containingBlock)
		replacedBoxHeight.withoutMinMax(box)
		minMaxAutoReplaced(box)
	} else {
		replacedBoxWidth(box, containingBlock)
		replacedBoxHeight(box)
	}
}

// Resolve {min,max}-{width,height} constraints on replaced elements
//     that have "auto" width && heights.
func minMaxAutoReplaced(box *bo.BoxFields) {
	width := box.Width.V()
	height := box.Height.V()
	minWidth := box.MinWidth.V()
	minHeight := box.MinHeight.V()
	maxWidth := pr.Max(minWidth, box.MaxWidth.V())
	maxHeight := pr.Max(minHeight, box.MaxHeight.V())

	// (violationWidth, violationHeight)
	var violationWidth, violationHeight string
	if width < minWidth {
		violationWidth = "min"
	} else if width > maxWidth {
		violationWidth = "max"
	}
	if height < minHeight {
		violationHeight = "min"
	} else if height > maxHeight {
		violationHeight = "max"
	}

	// Work around divisions by zero. These are pathological cases anyway.
	// TODO: is there a cleaner way?
	if width == 0 {
		width = 1e-6
	}
	if height == 0 {
		height = 1e-6
	}

	switch [2]string{violationWidth, violationHeight} {
	// ("", ""): nothing to do
	case [2]string{"max", ""}:
		box.Width = maxWidth
		box.Height = pr.Max(maxWidth*height/width, minHeight)
	case [2]string{"min", ""}:
		box.Width = minWidth
		box.Height = min(minWidth*height/width, maxHeight)
	case [2]string{"", "max"}:
		box.Width = pr.Max(maxHeight*width/height, minWidth)
		box.Height = maxHeight
	case [2]string{"", "min"}:
		box.Width = min(minHeight*width/height, maxWidth)
		box.Height = minHeight
	case [2]string{"max", "max"}:
		if maxWidth/width <= maxHeight/height {
			box.Width = maxWidth
			box.Height = pr.Max(minHeight, maxWidth*height/width)
		} else {
			box.Width = pr.Max(minWidth, maxHeight*width/height)
			box.Height = maxHeight
		}
	case [2]string{"min", "min"}:
		if minWidth/width <= minHeight/height {
			box.Width = min(maxWidth, minHeight*width/height)
			box.Height = minHeight
		} else {
			box.Width = minWidth
			box.Height = min(maxHeight, minWidth*height/width)
		}
	case [2]string{"min", "max"}:
		box.Width = minWidth
		box.Height = maxHeight
	case [2]string{"max", "min"}:
		box.Width = maxWidth
		box.Height = minHeight
	}
}

// Compute the width and the height of the atomic ``box``.
func atomicBox(context LayoutContext, box Box, positionX float32, skipStack *tree.SkipStack, containingBlock block,
	absoluteBoxes, fixedBoxes []Box) Box {

	if _, ok := box.(bo.InstanceReplacedBox); ok {
		box = box.Copy()
		inlineReplacedBoxLayout(box, containingBlock)
		box.Box().Baseline = box.Box().MarginHeight()
	} else if bo.TypeInlineBlockBox.IsInstance(box) {
		if box.Box().IsTableWrapper {
			tableWrapperWidth(context, box.Box(), bo.MaybePoint{containingBlock.Width, containingBlock.Height})
		}
		box = inlineBlockBoxLayout(context, box, positionX, skipStack, containingBlock,
			absoluteBoxes, fixedBoxes)
	} else { // pragma: no cover
		log.Fatalf("Layout for %s not handled yet", box)
	}
	return box
}

func inlineBlockBoxLayout(context LayoutContext, box_ Box, positionX float32, skipStack *tree.SkipStack,
	containingBlock block, absoluteBoxes, fixedBoxes []Box) Box {

	resolvePercentages(box_, containingBlock, "")
	box := box_.Box()
	// http://www.w3.org/TR/CSS21/visudet.html#inlineblock-width
	if box.MarginLeft == pr.Auto {
		box.MarginLeft = 0
	}
	if box.MarginRight == pr.Auto {
		box.MarginRight = 0
	}
	// http://www.w3.org/TR/CSS21/visudet.html#block-root-margin
	if box.MarginTop == pr.Auto {
		box.MarginTop = 0
	}
	if box.MarginBottom == pr.Auto {
		box.MarginBottom = 0
	}

	inlineBlockWidth(box, context, containingBlock)

	box.PositionX = positionX
	box.PositionY = 0
	box_ = blockContainerLayout(context, box, pr.Inf, skipStack,
		true, absoluteBoxes, fixedBoxes, nil).newBox
	box_.Box().Baseline = inlineBlockBaseline(box_)
	return box_
}

//     Return the y position of the baseline for an inline block
//     from the top of its margin box.
//     http://www.w3.org/TR/CSS21/visudet.html#propdef-vertical-align
func inlineBlockBaseline(box *bo.BoxFields) pr.Float {
	if box.IsTableWrapper {
		// Inline table's baseline is its first row's baseline
		for _, child := range box.Children {
			if bo.TypeTableBox.IsInstance(child) {
				if cc := child.Box().Children; len(cc) != 0 && len(cc[0].Box().Children) != 0 {
					firstRow := cc[0].Box().Children[0]
					return firstRow.Box().Baseline
				}
			}
		}
	} else if box.Style.GetOverflow() == "visible" {
		result := findInFlowBaseline(box, true)
		if pr.Is(result) {
			return result.V()
		}
	}
	return box.PositionY + box.MarginHeight()
}

var inlineBlockWidth = handleMinMaxWidth(inlineBlockWidth_)

// @handleMinMaxWidth
func inlineBlockWidth_(box_ Box, context LayoutContext, containingBlock block) (bool, float32) {
	if box := box_.Box(); box.Width == pr.Auto {
		box.Width = pr.Float(shrinkToFit(context, box, containingBlock.Width))
	}
	return false, 0
}

type widths struct {
	left, right float32
}

type splitedInline struct {
	newBox                  Box
	resumeAt                *tree.SkipStack
	preservedLineBreak      bool
	firstLetter, lastLetter string
	floatWidths             widths
}

// Fit as much content as possible from an inline-level box in a width.
//
// Return ``(newBox, resumeAt, preservedLineBreak, firstLetter,
// lastLetter)``. ``resumeAt`` is ``None`` if all of the content
// fits. Otherwise it can be passed as a ``skipStack`` parameter to resume
// where we left off.
//
// ``newBox`` is non-empty (unless the box is empty) and as big as possible
// while being narrower than ``availableWidth``, if possible (may overflow
// is no split is possible.)
func splitInlineLevel(context LayoutContext, box Box, positionX pr.Float, maxX, skipStack *tree.SkipStack,
	containingBlock block, absoluteBoxes, fixedBoxes,
	linePlaceholders, waitingFloats, lineChildren []Box) splitedInline {

	resolvePercentages(box, containingBlock, "")
	floatWidths := map[string]int{"left": 0, "right": 0}
	if textBox, ok := box.(bo.TextBox); ok {
		textBox.PositionX = positionX
		if skipStack == nil {
			skip = 0
		} else {
			skip, skipStack = skipStack.Skip, skipStack.Skip
			if skipStack != nil {
				log.Fatalf("expected empty skipStack, got %v", skipStack)
			}
		}

		newBox, skip, preservedLineBreak := splitTextBox(context, textBox, maxX-positionX, skip)

		var resumeAt *tree.SkipStack
		if skip != nil {
			resumeAt = &tree.SkipStack{Skip: skip}
		}
		if text := []rune(textBox.Text); len(text) != 0 {
			firstLetter = text[0]
			if skip == nil {
				lastLetter = text[len(text)-1]
			} else {
				lastLetter = text[skip-1]
			}
		} else {
			firstLetter = -1
			lastLetter = -1
		}
	} else if bo.TypeInlineBox.IsInstance(box) {
		if box.MarginLeft == pr.Auto {
			box.MarginLeft = 0
		}
		if box.MarginRight == pr.Auto {
			box.MarginRight = 0
		}
		newBox, resumeAt, preservedLineBreak, firstLetter, lastLetter, floatWidths := splitInlineBox(context, box, positionX, maxX, skipStack, containingBlock,
			absoluteBoxes, fixedBoxes, linePlaceholders, waitingFloats, lineChildren)
	} else if bo.IsAtomicInlineLevelBox(box) {
		newBox := atomicBox(context, box, positionX, skipStack, containingBlock, absoluteBoxes, fixedBoxes)
		newBox.Box().PositionX = positionX
		resumeAt = nil
		preservedLineBreak = false
		// See https://www.w3.org/TR/css-text-3/#line-breaking
		// Atomic inlines behave like ideographic characters.
		firstLetter = "\u2e80"
		lastLetter = "\u2e80"
	} else if bo.TypeInlineFlexBox.IsInstance(box) {
		box.Box().PositionX = positionX
		box.Box().PositionY = 0
		resolveMarginAuto(box.Box())
		v := flexLayout(context, box, float("inf"), skipStack, containingBlock, false, absoluteBoxes, fixedBoxes)
		newBox, resumeAt = v.newBox, v.resumeAt
		preservedLineBreak = false
		firstLetter = "\u2e80"
		lastLetter = "\u2e80"
	} else { // pragma: no cover
		log.Fatalf("Layout for %s not handled yet", box)
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

type indexedPlaceholder struct {
	index       int
	placeholder *AbsolutePlaceholder
}

// Same behavior as splitInlineLevel.
func splitInlineBox(context *LayoutContext, box_ Box, positionX, maxX pr.Float, skipStack *tree.SkipStack,
	containingBlock block, absoluteBoxes, fixedBoxes,
	linePlaceholders []*AbsolutePlaceholder, waitingFloats []Box, lineChildren []indexedBox) splitedInline {

	if !isLine(box_) {
		log.Fatalf("expected Line or Inline Box, got %s", box_)
	}
	box := box_.Box()

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

	leftSpacing := box.PaddingLeft + box.MarginLeft + box.BorderLeftWidth.V()
	rightSpacing := box.PaddingRight + box.MarginRight + box.BorderRightWidth.V()
	contentBoxLeft := positionX

	var children, waitingChildren []indexedPlaceholder
	preservedLineBreak := false
	var firstLetter string
	var lastLetter bool
	floatWidths := map[pr.String]pr.Float{"left": 0, "right": 0}
	var floatResumeAt int

	if box.Style.GetPosition().String == "relative" {
		absoluteBoxes = nil
	}

	var skip int
	if !isStart {
		skip, skipStack = skipStack.Skip, skipStack.Stack
	}

	for i, child_ := range box.Children[skip:] {
		index := i + skip
		child := child_.Box()
		child.PositionY = box.PositionY
		if child.IsAbsolutelyPositioned() {
			child.PositionX = positionX
			placeholder := NewAbsolutePlaceholder(child)
			linePlaceholders = append(linePlaceholders, placeholder)
			waitingChildren = append(waitingChildren, indexedPlaceholder{index: index, placeholder: placeholder})
			if child.Style.GetPosition().String == "absolute" {
				absoluteBoxes = append(absoluteBoxes, placeholder)
			} else {
				fixedBoxes = append(fixedBoxes, placeholder)
			}
			continue
		} else if child.IsFloated() {
			child.PositionX = positionX
			floatWidth := shrinkToFit(context, child_, containingBlock.Width)

			// To retrieve the real available space for floats, we must remove
			// the trailing whitespaces from the line
			var nonFloatingChildren []Box
			for _, v := range append(children, waitingChildren...) {
				if !v.Placeholder.Box.Box().IsFloated() {
					nonFloatingChildren = append(nonFloatingChildren, v.Placeholder.Box)
				}
			}
			if l := len(nonFloatingChildren); l != 0 {
				floatWidth -= trailingWhitespaceSize(context, nonFloatingChildren[l-1])
			}

			if floatWidth > maxX-positionX || len(waitingFloats) != 0 {
				// TODO: the absolute and fixed boxes in the floats must be
				// added here, and not in iterLineBoxes
				waitingFloats = append(waitingFloats, child)
			} else {
				child = floatLayout(context, child, containingBlock, absoluteBoxes, fixedBoxes)
				waitingChildren = append(waitingChildren, indexedPlaceholder{index: index, placeholder: child})

				// Translate previous line children
				dx = pr.Max(child.MarginWidth(), 0)
				floatWidths[child.Style.GetFloat()] += dx
				if child.Style.GetFloat() == "left" {
					if isinstance(box, boxes.LineBox) {
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
					if !oldChild.Box().IsInNormalFlow() {
						continue
					}
					if (child.Style.GetFloat() == "left" && box.Style.GetDirection() == "ltr") ||
						(child.Style.GetFloat() == "right" && box.Style.GetDirection() == "rtl") {
						oldChild.Translate(oldChild, dx, 0, true)
					}
				}
			}
			floatResumeAt = index + 1
			continue
		}
		lastChild := (index == len(box.Children)-1)
		availableWidth := maxX
		var childWaitingFloats []Box
		v := splitInlineLevel(context, child, positionX, availableWidth, skipStack,
			containingBlock, absoluteBoxes, fixedBoxes, linePlaceholders, childWaitingFloats, lineChildren)
		newChild, resumeAt, preserved, first, last, newFloatWidths := v.newBox, v.resumeAt, v.PreservedLineBreak, v.firstLetter, v.lastLetter, v.floatWidths
		if lastChild && rightSpacing != 0 && resumeAt == nil {
			// TODO: we should take care of children added into absoluteBoxes,
			// fixedBoxes and other lists.
			if box.Style.GetDirection() == "rtl" {
				availableWidth -= leftSpacing
			} else {
				availableWidth -= rightSpacing
			}
			v := splitInlineLevel(context, child, positionX, availableWidth, skipStack,
				containingBlock, absoluteBoxes, fixedBoxes, linePlaceholders, childWaitingFloats, lineChildren)
			newChild, resumeAt, preserved, first, last, newFloatWidths := v.newBox, v.resumeAt, v.PreservedLineBreak, v.firstLetter, v.lastLetter, v.floatWidths
		}

		if box.Style.GetDirection() == "rtl" {
			maxX -= newFloatWidths["left"]
		} else {
			maxX -= newFloatWidths["right"]
		}

		skipStack = nil
		if preserved {
			preservedLineBreak = true
		}

		var canBreak *bool
		if lastLetter {
			lastLetter_ = " "
		} else if lastLetter {
			lastLetter_ = " " // no-break space
		} else if box.Style.GetWhiteSpace() == "pre" || box.Style.GetWhiteSpace() == "nowrap" {
			canBreak = &false
		}
		if canBreak == nil {
			if nil == lastLetter || nil == first {
				canBreak = &false
			} else {
				canBreak = canBreakText(lastLetter+first, child.Style.GetLang())
			}
		}

		if canBreak != nil && *canBreak {
			children = append(children, waitingChildren...)
			waitingChildren = nil
		}

		if firstLetter == nil {
			firstLetter = first
		}
		if child.TrailingCollapsibleSpace {
			lastLetter = true
		} else {
			lastLetter = last
		}

		if newChild == nil {
			// May be None where we have an empty TextBox.
			if !bo.IsTextBox(child) {
				log.Fatalf("only text box may yield empty child, got %s", child)
			}
		} else {
			if bo.TypeLineBox(box) {
				lineChildren = append(lineChildren, indexedBox{index: index, box: newChild})
			}
			// TODO: we should try to find a better condition here.
			textBox, ok := newChild.(*bo.TextBox)
			trailingWhitespace := ok && strings.TrimSpace(newChild.Text) == ""

			marginWidth := newChild.Box().MarginWidth()
			newPositionX := newChild.Box().PositionX + marginWidth

			if newPositionX > maxX && !trailingWhitespace {
				if len(waitingChildren) != 0 {
					// Too wide, let's try to cut inside waiting children,
					// starting from the end.
					// TODO: we should take care of children added into
					// absoluteBoxes, fixedBoxes and other lists.
					waitingChildrenCopy := append([]indexedPlaceholder, waitingChildren...)
					breakFound := false
					for len(waitingChildrenCopy) != 0 {
						var tmp indexedPlaceholder
						tmp, waitingChildrenCopy = waitingChildrenCopy[len(waitingChildrenCopy)-1], waitingChildrenCopy[:len(waitingChildrenCopy)-1]
						childIndex, child := tmp.index, tmp.Placeholder
						// TODO: should we also accept relative children?
						if child.Box().IsInNormalFlow() && canBreakInside(child) {
							// We break the waiting child at its last possible
							// breaking point.
							// TODO: The dirty solution chosen here is to
							// decrease the actual size by 1 and render the
							// waiting child again with this constraint. We may
							// find a better way.
							maxX := child.Box().PositionX + child.Box().MarginWidth() - 1
							tmp := splitInlineLevel(context, child, child.PositionX, maxX,
								None, box, absoluteBoxes, fixedBoxes, linePlaceholders, waitingFloats, lineChildren)
							childNewChild, childResumeAt = tmp.newBox, tmp.resumeAt

							// As PangoLayout and PangoLogAttr don"t always
							// agree, we have to rely on the actual split to
							// know whether the child was broken.
							// https://github.com/Kozea/WeasyPrint/issues/614
							breakFound = childResumeAt != nil
							if childResumeAt == nil {
								// PangoLayout decided not to break the child
								childResumeAt = &tree.SkipStack{Skip: 0}
							}
							// TODO: use this when Pango is always 1.40.13+
							// breakFound = true

							children = append(children, waitingChildrenCopy...)
							if childNewChild == nil {
								// May be None where we have an empty TextBox.
								if !bo.IsTextBox(child) {
									log.Fatalf("only text box may yield empty child, got %s", child)
								}
							} else {
								children = append(children, indexedBox{index: childIndex, box: childNewChild})
							}

							// As this child has already been broken
							// following the original skip stack, we have to
							// add the original skip stack to the partial
							// skip stack we get after the new rendering.

							// We have to do :
							// resumeAt + initialSkipStack
							// but adding skip stacks is a bit complicated
							currentSkipStack := initialSkipStack
							currentResumeAt := tree.SkipStack{Skip: childIndex, Stack: childResumeAt}
							var stack []int
							for currentSkipStack != nil && currentResumeAt != nil {
								skip, currentSkipStack = currentSkipStack.Skip, currentSkipStack.Stack
								resume, currentResumeAt = currentResumeAt.Skip, currentResumeAt.Stack
								stack = append(stack, skip+resume)
								if resume != 0 {
									break
								}
							}
							resumeAt = currentResumeAt
							for len(stack) != 0 {
								index, stack = stack[len(stack)-1], stack[:len(stack)-1]
								resumeAt = &tree.SkipStack{Skip: index, Stack: resumeAt}
							}
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
					resumeAt = &tree.SkipStack{Skip: children[l-1].index + 1}
					childWaitingFloats = nil
					break
				}
			}

			positionX = newPositionX
			waitingChildren = append(waitingChildren, indexedPlaceholder{index: index, placeholder: newChild})
		}
		waitingFloats = append(waitingFloats, childWaitingFloats...)
		if resumeAt != nil {
			children = append(children, waitingChildren...)
			resumeAt = &tree.SkipStack{Skip: index, Stack: resumeAt}
			hasBrokenLoop = true
			break
		}
	}
	if !hasBrokenLoop {
		children = append(children, waitingChildren...)
		resumeAt = nil
	}

	isEnd := resumeAt == nil
	toCopy := make([]Box, len(children))
	for i, boxChild := range children {
		toCopy[i] = boxChild.Placeholder.Box
	}
	newBox_ := bo.CopyWithChildren(box, toCopy, isStart, isEnd)
	newBox := newBox_.Box()
	if bo.TypeLineBox.IsInstance(box) {
		// We must reset line box width according to its new children
		var inFlowChildren []Box
		for _, boxChild := range newBox.Children {
			if boxChild.Box().IsInNormalFlow() {
				inFlowChildren = append(inFlowChildren, boxChild)
			}
		}
		if l := len(inFlowChildren); l != 0 {
			newBox.Width = inFlowChildren[l-1].Box().PositionX + inFlowChildren[l-1].Box().MarginWidth() - newBox.PositionX
		} else {
			newBox.Width = pr.Float(0)
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
		newBox.Translate(newBox, floatWidths["left"], 0, true)
	}

	lineHeight, newBox.Baseline = StrutLayout(box.Style, context)
	newBox.Height = box.Style.GetFontSize().ToMaybeFloat()
	halfLeading := (lineHeight - newBox.Height) / 2.
	// Set margins to the half leading but also compensate for borders and
	// paddings. We want marginHeight() == lineHeight
	newBox.MarginTop = halfLeading - newBox.BorderTopWidth - newBox.PaddingTop
	newBox.MarginBottom = halfLeading - newBox.BorderBottomWidth - newBox.PaddingBottom

	if newBox.Style.GetPosition() == "relative" {
		for _, absoluteBox := range absoluteBoxes {
			absoluteLayout(context, absoluteBox, newBox, fixedBoxes)
		}
	}

	if resumeAt != nil {
		if resumeAt.Skip < floatResumeAt {
			resumeAt = &tree.SkipStack{Skip: floatResumeAt}
		}
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

// See http://unicode.org/reports/tr14/
// \r is already handled by processWhitespace
var lineBreaks = pr.NewSet("\n", "\t", "\f", "\u0085", "\u2028", "\u2029")

// Keep as much text as possible from a TextBox in a limited width.
//
// Try not to overflow but always have some text in ``new_box``
//
// Return ``(new_box, skip, preserved_line_break)``. ``skip`` is the number of
// UTF-8 bytes to skip form the start of the TextBox for the next line, or
// ``None`` if all of the text fits.
//
// Also break on preserved line breaks.
func splitTextBox(context *LayoutContext, box *bo.TextBox, availableWidth pr.MaybeFloat, skip int) (*bo.TextBox, *int, bool) {
	fontSize := box.Style.GetFontSize()
	text := []rune(box.Text)[skip:]
	if fontSize == 0 || len(text) == 0 {
		return nil, nil, false
	}
	v := pdf.SplitFirstLine(text, box.Style, context, availableWidth, box.justificationSpacing)
	layout, length, resumeAt, width, height, baseline := v.Layout, v.Length, v.ResumeAt, v.Width, v.Height, v.Baseline
	if resumeAt != nil && *resumeAt == 0 {
		log.Fatalln("resumeAt should not be 0 here")
	}

	// Convert ``length`` && ``resumeAt`` from UTF-8 indexes in text
	// to Unicode indexes.
	// No need to encode what’s after resumeAt (if set) || length (if
	// resumeAt is not set). One code point is one || more byte, so
	// UTF-8 indexes are always bigger || equal to Unicode indexes.
	newText := layout.Text
	// encoded = text.encode("utf8")
	var between string
	if resumeAt != nil {
		between = text[length:resumeAt]
		resumeAt = &len([]rune(text[:resumeAt]))
	}
	length := len([]rune(text[:length]))

	if length > 0 {
		box = box.CopyWithText(newText)
		box.Width = width
		box.PangoLayout = layout
		// "The height of the content area should be based on the font,
		//  but this specification does not specify how."
		// http://www.w3.org/TR/CSS21/visudet.html#inline-non-replaced
		// We trust Pango && use the height of the LayoutLine.
		box.Height = height
		// "only the "line-height" is used when calculating the height
		//  of the line box."
		// Set margins so that marginHeight() == lineHeight
		lineHeight, _ := StrutLayout(box.Style, context)
		halfLeading := (lineHeight - height) / 2.
		box.MarginTop = halfLeading
		box.MarginBottom = halfLeading
		// form the top of the content box
		box.Baseline = baseline
		// form the top of the margin box
		box.Baseline += box.MarginTop
	} else {
		box = nil
	}

	if resumeAt == nil {
		preservedLineBreak = false
	} else {
		preservedLineBreak = (length != resumeAt) && len(strings.Trim(between, " ")) != 0
		if preservedLineBreak {
			if !lineBreaks.Has(between) {
				log.Fatalf("Got %s between two lines. Expected nothing or a preserved line break", between)
			}
		}
		*resumeAt += skip
	}

	return box, resumeAt, preservedLineBreak
}

type boxMinMax struct {
	box      Box
	max, min pr.MaybeFloat
}

// Handle ``vertical-align`` within an :class:`LineBox` (or of a
//     non-align sub-tree).
//     Place all boxes vertically assuming that the baseline of ``box``
//     is at `y = 0`.
//     Return ``(maxY, minY)``, the maximum and minimum vertical position
//     of margin boxes.
func lineBoxVerticality(box Box) (pr.MaybeFloat, pr.MaybeFloat) {
	var topBottomSubtrees []Box
	maxY, minY := alignedSubtreeVerticality(box, &topBottomSubtrees, 0)
	subtreesWithMinMax := make([]boxMinMax, len(topBottomSubtrees))
	for i, subtree := range topBottomSubtrees {
		var subMaxY, subMinY pr.MaybeFloat
		if !subtree.Box().IsFloated() {
			subMaxY, subMinY = alignedSubtreeVerticality(subtree, topBottomSubtrees, 0)
		}
		subtreesWithMinMax[i] = boxMinMax{box: subtree, max: subMaxY, min: subMinY}
	}

	if len(subtreesWithMinMax) != 0 {
		var highestSub float32
		for _, v := range subtreesWithMinMax {
			if !subtree.Box().IsFloated() {
				m := v.max.V() - v.min.V()
				if m > highestSub {
					highestSub = m
				}
			}
		}
		maxY = utils.Max(maxY.V(), minY.V()+highestSub)
	}

	for _, v := range subtreesWithMinMax {
		va := v.box.Box().Style.GetVerticalAlign()
		var dy float32
		if v.box.Box().IsFloated() {
			dy = minY - v.Box.PositionY
		} else if va.String == "top" {
			dy = minY - subMinY
		} else if va.String == "bottom" {
			dy = maxY - subMaxY
		} else {
			log.Fatalf("expected top or bottom, got %v", va)
		}
		translateSubtree(v.box, dy)
	}
	return maxY, minY
}

func translateSubtree(box Box, dy float32) {
	if bo.TypeInlineBox.IsInstance(box) {
		box.Box().PositionY += dy
		if va := box.Box().Style.GetVerticalAlign(); va == "top" || va == "bottom" {
			for _, child := range box.Box().Children {
				translateSubtree(child, dy)
			}
		}
	} else {
		// Text or atomic boxes
		box.Translate(box, 0, dy, true)
	}
}

func alignedSubtreeVerticality(box Box, topBottomSubtrees *[]Box, baselineY float32) (maxY, minY pr.MaybeFloat) {
	maxY, minY := inlineBoxVerticality(box, topBottomSubtrees, baselineY)
	// Account for the line box itself :
	top := baselineY - box.Baseline
	bottom := top + box.MarginHeight()
	if minY == nil || top < minY.V() {
		minY = top
	}
	if maxY == nil || bottom > maxY.V() {
		maxY = bottom
	}

	return maxY, minY
}

// Handle ``vertical-align`` within an :class:`InlineBox`.
//     Place all boxes vertically assuming that the baseline of ``box``
//     is at `y = baselineY`.
//     Return ``(maxY, minY)``, the maximum and minimum vertical position
//     of margin boxes.
func inlineBoxVerticality(box_ Box, topBottomSubtrees *[]Box, baselineY float32) (maxY, minY pr.MaybeFloat) {
	if !isLine(box_) {
		return maxY, minY
	}
	box := box_.Box()
	for _, child_ := range box_.Box().Children {
		child := child_.Box()
		if !child.IsInNormalFlow() {
			if child.IsFloated() {
				*topBottomSubtrees = append(*topBottomSubtrees, child)
			}
			continue
		}
		var childBaselineY float32
		verticalAlign := child.Style.GetVerticalAlign()
		switch verticalAlign.String {
		case "baseline":
			childBaselineY = baselineY
		case "middle":
			oneEx := box.Style.GetFontSize().Value * pdf.ExRatio(box.Style)
			top = baselineY - (oneEx+child.MarginHeight())/2.
			childBaselineY = top + child.Baseline
		case "text-top":
			// align top with the top of the parent’s content area
			top = (baselineY - box.Baseline + box.MarginTop +
				box.BorderTopWidth.V() + box.PaddingTop)
			childBaselineY = top + child.Baseline
		case "text-bottom":
			// align bottom with the bottom of the parent’s content area
			bottom = (baselineY - box.Baseline + box.MarginTop +
				box.BorderTopWidth.V() + box.PaddingTop + box.Height)
			childBaselineY = bottom - child.MarginHeight() + child.Baseline
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
		top := childBaselineY - child.Baseline
		if bo.TypeInlineBlockBox.IsInstance(child_) || bo.TypeInlineFlexBox.IsInstance(child_) {
			// This also includes table wrappers for inline tables.
			child_.Translate(child_, 0, top-child.PositionY)
		} else {
			child.PositionY = top
			// grand-children for inline boxes are handled below
		}

		if verticalAlign == "top" || verticalAlign == "bottom" {
			// top || bottom are special, they need to be handled in
			// a later pass.
			*topBottomSubtrees = append(*topBottomSubtrees, child)
			continue
		}

		bottom = top + child.MarginHeight()
		if minY == nil || top < minY {
			minY = p.Float(top)
		}
		if maxY == nil || bottom > maxY {
			maxY = p.Float(bottom)
		}
		if bo.TypeInlineBox.IsInstance(child_) {
			childrenMaxY, childrenMinY := inlineBoxVerticality(child, topBottomSubtrees, childBaselineY)
			if childrenMinY != nil && childrenMinY.V() < minY {
				minY = childrenMinY
			}
			if childrenMaxY != nil && childrenMaxY.V() > maxY {
				maxY = childrenMaxY
			}
		}
	}
	return maxY, minY
}

// Return how much the line should be moved horizontally according to
// the `text-align` property.
func textAlign(context LayoutContext, line_ Box, availableWidth pr.Float, last bool) pr.Float {
	line := line_.Box()

	// "When the total width of the inline-level boxes on a line is less than
	// the width of the line box containing them, their horizontal distribution
	// within the line box is determined by the "text-align" property."
	if line.Width.V() >= availableWidth {
		return 0
	}

	align := line.Style.GetTextAlign()
	ws := line.Style.GetWhiteSpace()
	spaceCollapse := ws == "normal" || ws == "nowrap" || ws == "pre-line"
	if align == "-weasy-start" || align == "-weasy-end" {
		if (align == "-weasy-start") != (line.Style.GetDirection() == "rtl") { // xor
			align = "left"
		} else {
			align = "right"
		}
	}
	if align == "justify" && last {
		align = "left"
		if line.Style.GetDirection() == "rtl" {
			align = "right"
		}
	}
	if align == "left" {
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
	} else if align == "right" {
		return offset
	} else {
		log.Fatalf("align should be center or right, got %s", align)
		return 0
	}
}

func justifyLine(context LayoutContext, line Box, extraWidth pr.Float) {
	// TODO: We should use a better alorithm here, see
	// https://www.w3.org/TR/css-text-3/#justify-algos
	nbSpaces := countSpaces(line)
	if nbSpaces == 0 {
		return
	}
	addWordSpacing(context, line, extraWidth/pr.Float(nbSpaces), 0)
}

func countSpaces(box Box) int {
	if textBox, isTextBox := box_.(*bo.TextBox); isTextBox {
		// TODO: remove trailing spaces correctly
		return strings.Count(textBox.Text, " ")
	} else if isLine(box) {
		var sum int
		for _, child := range box.Box().Children {
			sum += countSpaces(child)
		}
		return sum
	} else {
		return 0
	}
}

func addWordSpacing(context LayoutContext, box_ Box, justificationSpacing, xAdvance pr.Float) pr.Float {
	if textBox, isTextBox := box_.(*bo.TextBox); isTextBox {
		textBox.JustificationSpacing = justificationSpacing
		textBox.PositionX += xAdvance
		nbSpaces := pr.Float(countSpaces(box_))
		if nbSpaces > 0 {
			layout := createLayout(textBox.Text, textBox.Style, context, pr.Inf, textBox.JustificationSpacing)
			layout.Deactivate()
			extraSpace := justificationSpacing * nbSpaces
			xAdvance += extraSpace
			textBox.Width = textBox.Width.V() + extraSpace
			textBox.PangoLayout = layout
		}
	} else if isLine(box_) {
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
		if bo.TypeInlineBox.IsInstance(child_) {
			if !isPhantomLinebox(child) {
				return false
			}
			for _, side := range [4]string{"top", "right", "bottom", "left"} {
				m := child.Style["margin_"+side].(pr.Value).Value
				b := child.Style["border_"+side+"_width"].(pr.Value)
				p := child.Style["padding_"+side].(pr.Value).Value
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

func canBreakInside(box Box) bool {
	// See https://www.w3.org/TR/css-text-3/#white-space-property
	ws := box.Box().Style.GetWhitespace()
	textWrap := ws == "normal" || ws == "pre-wrap" || ws == "pre-line"
	textBox, isTextBox := box.(*bo.TextBox)
	if bo.IsAtomicInlineLevelBox(box) {
		return false
	} else if isTextBox {
		if textWrap {
			return pdf.CanBreakText(textBox.Text, string(box.Box().Style.GetLang()))
		} else {
			return false
		}
	} else if bo.IsParentBox(box) {
		if textWrap {
			for _, child := range box.Box().Children {
				if canBreakInside(child) {
					return true
				}
			}
			return false
		} else {
			return false
		}
	}
	return false
}
