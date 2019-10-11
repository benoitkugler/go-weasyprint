package layout

import (
	"github.com/benoitkugler/go-weasyprint/pdf"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	bo "github.com/benoitkugler/go-weasyprint/boxes"
)

//     Line breaking and layout for inline-level boxes.

func isLine(box Box) bool {
    return bo.TypeLineBox.IsInstance(box) || bo.TypeInlineBox.IsInstance(box)
}


type lineBoxe struct {
	line *bo.LineBox
	resumeAt *float32
}

// Return an iterator of ``(line, resumeAt)``.
// 
// ``line`` is a laid-out LineBox with as much content as possible that
// fits in the available width.
// 
// :param box: a non-laid-out :class:`LineBox`
// :param positionY: vertical top position of the line box on the page
// :param skipStack: ``None`` to start at the beginning of ``linebox``,
// 				   or a ``resumeAt`` value to continue just after an
// 				   already laid-out line.
func iterLineBoxes(context LayoutContext, box *bo.LineBox, positionY, skipStack, containingBlock Box,
                    absoluteBoxes []AbsolutePlaceholder, fixedBoxes []Box, firstLetterStyle pr.Properties) []lineBoxe {
    resolvePercentages(box, containingBlock)
    if skipStack == nil {
        // TODO: wrong, see https://github.com/Kozea/WeasyPrint/issues/679
        box.TextIndent = resolveOnePercentage(box, "textIndent", containingBlock.width)
    } else {
        box.TextIndent = 0
	} 
	var out []lineBoxe
	for {
        line, resumeAt := getNextLinebox( context, box, positionY, skipStack, containingBlock,
            absoluteBoxes, fixedBoxes, firstLetterStyle)
        if line {
            positionY = line.positionY + line.height
		}
		 if line == nil {
            return out
		} 
		out = append(out, lineBoxe{line: line, resumeAt: resumeAt})
        if resumeAt == nil {
            return out
		} 
		skipStack = resumeAt
        box.TextIndent = 0
        firstLetterStyle = None
	}
}


func getNextLinebox(context LayoutContext, linebox *bo.LineBox, positionY, skipStack,
                     containingBlock, absoluteBoxes, fixedBoxes,
                     firstLetterStyle pr.Properties)  lineBoxe {
                     
    skipStack , cont := skipFirstWhitespace(linebox, skipStack)
    if cont {
        return lineBoxe{}
    }

    skipStack = firstLetterToBox(linebox, skipStack, firstLetterStyle)

    linebox.PositionY = positionY

    if len(context.excludedShapes) != 0 {
        // Width && height must be calculated to avoid floats
        linebox.Width = MF(inlineMinContentWidth( context, linebox,true, skipStack, true, false))
        linebox.Height, _ = strutLayout(linebox.Style, context)
    } else {
        // No float, width && height will be set by the lines
		linebox.Width = MF(0)
		linebox.Height = MF(0)
	} 
	positionX, positionY, availableWidth := avoidCollisions(context, linebox, containingBlock, false)

    candidateHeight := linebox.Height

	excludedShapes := make([]shape, len(context.excludedShapes))
	for i,v := range context.excludedShapes {
		excludedShapes[i] = v
	}

	var line *BoxFields
    for {
        linebox.PositionX = positionX
        linebox.PositionY = positionY
        maxX := positionX + availableWidth
        positionX += linebox.TextIndent
    
        var (
			linePlaceholders  []AbsolutePlaceholder
        lineAbsolutes  []AbsolutePlaceholder
        lineFixed  []Box
		waitingFloats  []Box
		)

		spi  := splitInlineBox(context, linebox, positionX, maxX, skipStack, containingBlock,
			lineAbsolutes, lineFixed, linePlaceholders, waitingFloats,nil)

		line, resumeAt, preservedLineBreak, firstLetter, lastLetter, floatWidth := spi.NewBox, spi.resumeAt, spi.preservedLineBreak,
			spi.firstLetter, spi.lastLetter, spi.floatWidths
 			 
        linebox.Width, linebox.Height = line.Width, line.Height

        if isPhantomLinebox(line) && !preservedLineBreak {
            line.Height = bo.MF(0)
            break
        }

        removeLastWhitespace(context, line)

        newPositionX, _, newAvailableWidth = avoidCollisions(context, linebox, containingBlock, outer=false)
        // TODO: handle rtl
        newAvailableWidth -= floatWidth["right"]
        alignmentAvailableWidth = (
            newAvailableWidth + newPositionX - linebox.PositionX)
        offsetX = textAlign(
            context, line, alignmentAvailableWidth,
            last=(resumeAt == nil || preservedLineBreak))

        bottom, top = lineBoxVerticality(line)
        assert top is not None
        assert bottom is not None
        line.baseline = -top
        line.positionY = top
        line.height = bottom - top
        offsetY = positionY - top
        line.marginTop = 0
        line.marginBottom = 0

        line.translate(offsetX, offsetY)
        // Avoid floating point errors, as positionY - top + top != positionY
        // Removing this line breaks the position == linebox.Position test below
        // See https://github.com/Kozea/WeasyPrint/issues/583
        line.positionY = positionY

        if line.height <= candidateHeight {
            break
        } candidateHeight = line.height

        newExcludedShapes = context.excludedShapes
        context.excludedShapes = excludedShapes
        positionX, positionY, availableWidth = avoidCollisions(
            context, line, containingBlock, outer=false)
        if (positionX, positionY) == (
                linebox.PositionX, linebox.PositionY) {
                }
            context.excludedShapes = newExcludedShapes
            break

    absoluteBoxes.extend(lineAbsolutes)
    fixedBoxes.extend(lineFixed)

    for placeholder := range linePlaceholders {
        if placeholder.style["WeasySpecifiedDisplay"].startswith("inline") {
            // Inline-level static position {
            } placeholder.translate(0, positionY - placeholder.positionY)
        } else {
            // Block-level static position: at the start of the next line
            placeholder.translate(
                line.positionX - placeholder.positionX,
                positionY + line.height - placeholder.positionY)
        }
    }

    floatChildren = []
    waitingFloatsY = line.positionY + line.height
    for waitingFloat := range waitingFloats {
        waitingFloat.positionY = waitingFloatsY
        waitingFloat = floatLayout(
            context, waitingFloat, containingBlock, absoluteBoxes,
            fixedBoxes)
        floatChildren.append(waitingFloat)
    } if floatChildren {
        line.children += tuple(floatChildren)
    }

    return line, resumeAt


// Return the ``skipStack`` to start just after the remove spaces
//     at the beginning of the line.
//     See http://www.w3.org/TR/CSS21/text.html#white-space-model
func skipFirstWhitespace(box Box, skipStack *bo.SkipStack) (ss *bo.SkipStack, continue_ bool){
	var (
		index int 
		nextSkipStack *bo.SkipStack
	)
	if skipStack != nil {
        index, nextSkipStack = skipStack.Skip, skipStack.Stack
    }

    if textBox, ok :=box.(*bo.TextBox) ; ok {
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
		if index != 0{
			return &bo.SkipStack{Skip:index}, false
		}
		return nil, false
    }

    if isLine(box) {
		children := box.Box().Children
		if index == 0 && len(children) == 0{
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
			return &bo.SkipStack{Skip:index, Stack: result}, false
		}
		return nil, false
    }
	if skipStack != nil {
		log.Fatalf("unexpected skip inside %s", box)
	} 
    return nil, false
}

// Remove := range place space characters at the end of a line.
//     This also reduces the width of the inline parents of the modified text.
//     
func removeLastWhitespace(context, box) {
    ancestors = []
    while isinstance(box, (boxes.LineBox, boxes.InlineBox)) {
        ancestors.append(box)
        if not box.children {
            return
        } box = box.children[-1]
    } if not (isinstance(box, boxes.TextBox) and
            box.Style["whiteSpace"] := range ("normal", "nowrap", "pre-line")) {
            }
        return
    newText = box.text.rstrip(" ")
    if newText {
        if len(newText) == len(box.text) {
            return
        } box.text = newText
        newBox, resume, _ = splitTextBox(context, box, None, 0)
        assert newBox is not None
        assert resume == nil
        spaceWidth = box.Width - newBox.Width
        box.Width = newBox.Width
    } else {
        spaceWidth = box.Width
        box.Width = 0
        box.text = ""
    }
} 
    for ancestor := range ancestors {
        ancestor.width -= spaceWidth
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
func firstLetterToBox(box Box, skipStack *bo.SkipStack, firstLetterStyle pr.Properties) *bo.SkipStack {
    if len(firstLetterStyle) != 0 && len(box.Box().Children) != 0 {
        // Some properties must be ignored :in first-letter boxes.
        // https://drafts.csswg.org/selectors-3/#application-in-css
        // At least, position is ignored to avoid layout troubles.
        firstLetterStyle.SetPosition(pr.String("static"))
    }

        firstLetter := ""
        child := box.Box().children[0]
        if textBox, ok := child.(*bo.TextBox); ok {
            letterStyle := tree.ComputedFromCascaded(cascaded={}, parentStyle=firstLetterStyle, element=None)
            if strings.HasSuffix(textBox.ElementTag, "::first-letter") {
                letterBox := bo.NewInlineBox(textBox.ElementTag + "::first-letter", letterStyle, []Box{child})
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
                        letterBox := bo.NewInlineBox(textBox.ElementTag + "::first-letter", firstLetterStyle, nil)
                        textBox_ := bo.NewTextBox(textBox.ElementTag + "::first-letter", letterStyle, firstLetter)
                        letterBox.Children = []Box{&textBox_}
                        textBox.Children = append([]Box{&letterBox}, textBox.Children...)
                    } else {
                        letterBox := bo.NewBlockBox(textBox.ElementTag + "::first-letter", firstLetterStyle, nil)
                        letterBox.FirstLetterStyle = nil
                        lineBox := bo.NewLineBox(textBox.ElementTag + "::first-letter", firstLetterStyle, nil)
                        letterBox.Children = []Box{&lineBox}
                        textBox_ := bo.NewTextBox(textBox.ElementTag + "::first-letter", letterStyle, firstLetter)
                        lineBox.Children = []Box{&textBox_}
                        textBox.Children = append([]Box{&letterBox}, textBox.Children...)
						} 
					if skipStack  != nil && childSkipStack != nil {
                        skipStack = bo.SkipStack{Skip: skipStack.Skip, Stack: &bo.SkipStack{
							Skip: childSkipStack.Skip + 1,
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
                skipStack = &bo.SkipStack{Skip: skipStack.Skip,Stack: childSkipStack}
            }
        }
    return skipStack
	}

var replacedBoxWidth = handleMinMaxWidth(replacedBoxWidth_)

// @handleMinMaxWidth
// 
//     Compute and set the used width for replaced boxes (inline- or block-level)
//     
func replacedBoxWidth_(box_ Box , _ LayoutContext, containingBlock block) (bool, float32) {
	box, ok := bo.AsReplaced(box)
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
func replacedBoxHeight_(box_ Box, _ LayoutContext, _ block) (bool, float32) {
	box, ok := bo.AsReplaced(box)
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


// Lay out an inline :class:`boxes.ReplacedBox` ``box``.
func inlineReplacedBoxLayout(box, containingBlock) {
    for side := range ["top", "right", "bottom", "left"] {
        if getattr(box, "margin" + side) == Auto {
            setattr(box, "margin" + side, 0)
        }
	} 
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
//     
func minMaxAutoReplaced(box) {
    width = box.Width
    height = box.Height
    minWidth = box.minWidth
    minHeight = box.minHeight
    maxWidth = max(minWidth, box.maxWidth)
    maxHeight = max(minHeight, box.maxHeight)
} 
    // (violationWidth, violationHeight)
    violations = (
        "min" if width < minWidth else "max" if width > maxWidth else "",
        "min" if height < minHeight else "max" if height > maxHeight else "")

    // Work around divisions by zero. These are pathological cases anyway.
    // TODO: is there a cleaner way?
    if width == 0 {
        width = 1e-6
    } if height == 0 {
        height = 1e-6
    }

    // ("", ""): nothing to do
    if violations == ("max", "") {
        box.Width = maxWidth
        box.Height = max(maxWidth * height / width, minHeight)
    } else if violations == ("min", "") {
        box.Width = minWidth
        box.Height = min(minWidth * height / width, maxHeight)
    } else if violations == ("", "max") {
        box.Width = max(maxHeight * width / height, minWidth)
        box.Height = maxHeight
    } else if violations == ("", "min") {
        box.Width = min(minHeight * width / height, maxWidth)
        box.Height = minHeight
    } else if violations == ("max", "max") {
        if maxWidth / width <= maxHeight / height {
            box.Width = maxWidth
            box.Height = max(minHeight, maxWidth * height / width)
        } else {
            box.Width = max(minWidth, maxHeight * width / height)
            box.Height = maxHeight
        }
    } else if violations == ("min", "min") {
        if minWidth / width <= minHeight / height {
            box.Width = min(maxWidth, minHeight * width / height)
            box.Height = minHeight
        } else {
            box.Width = minWidth
            box.Height = min(maxHeight, minWidth * height / width)
        }
    } else if violations == ("min", "max") {
        box.Width = minWidth
        box.Height = maxHeight
    } else if violations == ("max", "min") {
        box.Width = maxWidth
        box.Height = minHeight
    }


func atomicBox(context, box, positionX, skipStack, containingBlock,
               absoluteBoxes, fixedBoxes) {
               }
    """Compute the width && the height of the atomic ``box``."""
    if isinstance(box, boxes.ReplacedBox) {
        box = box.copy()
        inlineReplacedBoxLayout(box, containingBlock)
        box.Baseline = box.marginHeight()
    } else if isinstance(box, boxes.InlineBlockBox) {
        if box.isTableWrapper {
            tableWrapperWidth(
                context, box,
                (containingBlock.width, containingBlock.height))
        } box = inlineBlockBoxLayout(
            context, box, positionX, skipStack, containingBlock,
            absoluteBoxes, fixedBoxes)
    } else:  // pragma: no cover
        raise TypeError("Layout for %s not handled yet" % type(box)._Name_)
    return box


func inlineBlockBoxLayout(context, box, positionX, skipStack,
                            containingBlock, absoluteBoxes, fixedBoxes) {
                            }
    // Avoid a circular import
    from .blocks import blockContainerLayout

    resolvePercentages(box, containingBlock)

    // http://www.w3.org/TR/CSS21/visudet.html#inlineblock-width
    if box.marginLeft == "auto" {
        box.marginLeft = 0
    } if box.marginRight == "auto" {
        box.marginRight = 0
    } // http://www.w3.org/TR/CSS21/visudet.html#block-root-margin
    if box.marginTop == "auto" {
        box.marginTop = 0
    } if box.marginBottom == "auto" {
        box.marginBottom = 0
    }

    inlineBlockWidth(box, context, containingBlock)

    box.PositionX = positionX
    box.PositionY = 0
    box, _, _, _, _ = blockContainerLayout(
        context, box, maxPositionY=float("inf"), skipStack=skipStack,
        pageIsEmpty=true, absoluteBoxes=absoluteBoxes,
        fixedBoxes=fixedBoxes)
    box.Baseline = inlineBlockBaseline(box)
    return box


// 
//     Return the y position of the baseline for an inline block
//     from the top of its margin box.
//     http://www.w3.org/TR/CSS21/visudet.html#propdef-vertical-align
//     
func inlineBlockBaseline(box) {
    if box.isTableWrapper {
        // Inline table"s baseline is its first row"s baseline
        for child := range box.children {
            if isinstance(child, boxes.TableBox) {
                if child.children && child.children[0].children {
                    firstRow = child.children[0].children[0]
                    return firstRow.baseline
                }
            }
        }
    } else if box.Style["overflow"] == "visible" {
        result = findInFlowBaseline(box, last=true)
        if result {
            return result
        }
    } return box.PositionY + box.marginHeight()
} 

@handleMinMaxWidth
func inlineBlockWidth(box, context, containingBlock) {
    if box.Width == "auto" {
        box.Width = shrinkToFit(context, box, containingBlock.width)
    }
} 

type splitedInline struct {
	newBox Box
	resumeAt *int 
	preservedLineBreak bool
	firstLetter, lastLetter string
	floatWidths [2]float32 // left, right
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
func splitInlineLevel(context LayoutContext, box Box, positionX, maxX, skipStack,
                       containingBlock, absoluteBoxes, fixedBoxes,
                       linePlaceholders, waitingFloats, lineChildren []Box) splitedInline {

    resolvePercentages(box, containingBlock)
    floatWidths = {"left": 0, "right": 0}
    if isinstance(box, boxes.TextBox) {
        box.PositionX = positionX
        if skipStack == nil {
            skip = 0
        } else {
            skip, skipStack = skipStack
            skip = skip || 0
            assert skipStack == nil
        }
    }

        newBox, skip, preservedLineBreak = splitTextBox(
            context, box, maxX - positionX, skip)

        if skip == nil {
            resumeAt = None
        } else {
            resumeAt = (skip, None)
        } if box.text {
            firstLetter = box.text[0]
            if skip == nil {
                lastLetter = box.text[-1]
            } else {
                lastLetter = box.text[skip - 1]
            }
        } else {
            firstLetter = lastLetter = None
        }
    else if isinstance(box, boxes.InlineBox) {
        if box.marginLeft == "auto" {
            box.marginLeft = 0
        } if box.marginRight == "auto" {
            box.marginRight = 0
        } (newBox, resumeAt, preservedLineBreak, firstLetter,
         lastLetter, floatWidths) = splitInlineBox(
            context, box, positionX, maxX, skipStack, containingBlock,
            absoluteBoxes, fixedBoxes, linePlaceholders, waitingFloats,
             lineChildren)
    } else if isinstance(box, boxes.AtomicInlineLevelBox) {
        newBox = atomicBox(
            context, box, positionX, skipStack, containingBlock,
            absoluteBoxes, fixedBoxes)
        newBox.PositionX = positionX
        resumeAt = None
        preservedLineBreak = false
        // See https://www.w3.org/TR/css-text-3/#line-breaking
        // Atomic inlines behave like ideographic characters.
        firstLetter = "\u2e80"
        lastLetter = "\u2e80"
    } else if isinstance(box, boxes.InlineFlexBox) {
        box.PositionX = positionX
        box.PositionY = 0
        for side := range ["top", "right", "bottom", "left"] {
            if getattr(box, "margin" + side) == "auto" {
                setattr(box, "margin" + side, 0)
            }
        } newBox, resumeAt, _, _, _ = flexLayout(
            context, box, float("inf"), skipStack, containingBlock,
            false, absoluteBoxes, fixedBoxes)
        preservedLineBreak = false
        firstLetter = "\u2e80"
        lastLetter = "\u2e80"
    } else:  // pragma: no cover
        raise TypeError("Layout for %s not handled yet" % type(box)._Name_)
    return (
        newBox, resumeAt, preservedLineBreak, firstLetter, lastLetter,
        floatWidths)
	}


    type indexedPlaceholder struct {
        index int 
        placeholder *AbsolutePlaceholder
    }

// Same behavior as splitInlineLevel.
func splitInlineBox(context LayoutContext, box_ Box, positionX, maxX, skipStack *bo.SkipStack,
                     containingBlock block, absoluteBoxes, fixedBoxes,
                     linePlaceholders []*AbsolutePlaceholder, waitingFloats, lineChildren) splitedInline {
                     
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

    var children , waitingChildren []indexedPlaceholder 
    preservedLineBreak /= false
    var firstLetter, lastLetter string
    floatWidths := map[string]float32{"left": 0, "right": 0}
    var floatResumeAt float32

    if box.Style.GetPosition().String == "relative" {
        absoluteBoxes = []
    }

     var   skip int 
    if ! isStart {
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
            waitingChildren = append(waitingChildren, indexedPlaceholder{index:index,placeholder: placeholder})
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
            nonFloatingChildren = [
                child_ for _, child_ := range (children + waitingChildren)
                if not child.isFloated()]
            if nonFloatingChildren {
                floatWidth -= trailingWhitespaceSize(
                    context, nonFloatingChildren[-1])
            }

            if floatWidth > maxX - positionX || waitingFloats {
                // TODO: the absolute && fixed boxes := range the floats must be
                // added here, && not := range iterLineBoxes
                waitingFloats.append(child)
            } else {
                child = floatLayout(
                    context, child, containingBlock, absoluteBoxes,
                    fixedBoxes)
                waitingChildren.append((index, child))
            }

                // Translate previous line children
                dx = max(child.MarginWidth(), 0)
                floatWidths[child.Style["float"]] += dx
                if child.Style["float"] == "left" {
                    if isinstance(box, boxes.LineBox) {
                        // The parent is the line, update the current position
                        // for the next child. When the parent is not the line
                        // (it is an inline block), the current position of the
                        // line is updated by the box itself (see next
                        // splitInlineLevel call).
                        positionX += dx
                    }
                } else if child.Style["float"] == "right" {
                    // Update the maximum x position for the next children
                    maxX -= dx
                } for _, oldChild := range lineChildren {
                    if not oldChild.isInNormalFlow() {
                        continue
                    } if ((child.Style["float"] == "left" and
                            box.Style["direction"] == "ltr") or
                        (child.Style["float"] == "right" and
                            box.Style["direction"] == "rtl")) {
                            }
                        oldChild.translate(dx=dx)
                }
            floatResumeAt = index + 1
            continue

        lastChild = (index == len(box.children) - 1)
        availableWidth = maxX
        childWaitingFloats = []
        newChild, resumeAt, preserved, first, last, newFloatWidths = (
            splitInlineLevel(
                context, child, positionX, availableWidth, skipStack,
                containingBlock, absoluteBoxes, fixedBoxes,
                linePlaceholders, childWaitingFloats, lineChildren))
        if lastChild && rightSpacing && resumeAt == nil {
            // TODO: we should take care of children added into absoluteBoxes,
            // fixedBoxes && other lists.
            if box.Style["direction"] == "rtl" {
                availableWidth -= leftSpacing
            } else {
                availableWidth -= rightSpacing
            } newChild, resumeAt, preserved, first, last, newFloatWidths = (
                splitInlineLevel(
                    context, child, positionX, availableWidth, skipStack,
                    containingBlock, absoluteBoxes, fixedBoxes,
                    linePlaceholders, childWaitingFloats, lineChildren))
        }

        if box.Style["direction"] == "rtl" {
            maxX -= newFloatWidths["left"]
        } else {
            maxX -= newFloatWidths["right"]
        }

        skipStack = None
        if preserved {
            preservedLineBreak = true
        }

        canBreak = None
        if lastLetter is true {
            lastLetter = " "
        } else if lastLetter is false {
            lastLetter = " "  // no-break space
        } else if box.Style["whiteSpace"] := range ("pre", "nowrap") {
            canBreak = false
        } if canBreak == nil {
            if None := range (lastLetter, first) {
                canBreak = false
            } else {
                canBreak = canBreakText(
                    lastLetter + first, child.Style["lang"])
            }
        }

        if canBreak {
            children.extend(waitingChildren)
            waitingChildren = []
        }

        if firstLetter == nil {
            firstLetter = first
        } if child.trailingCollapsibleSpace {
            lastLetter = true
        } else {
            lastLetter = last
        }

        if newChild == nil {
            // May be None where we have an empty TextBox.
            assert isinstance(child, boxes.TextBox)
        } else {
            if isinstance(box, boxes.LineBox) {
                lineChildren.append((index, newChild))
            } // TODO: we should try to find a better condition here.
            trailingWhitespace = (
                isinstance(newChild, boxes.TextBox) and
                not newChild.text.strip())
        }

            marginWidth = newChild.MarginWidth()
            newPositionX = newChild.PositionX + marginWidth

            if newPositionX > maxX && not trailingWhitespace {
                if waitingChildren {
                    // Too wide, let"s try to cut inside waiting children,
                    // starting from the end.
                    // TODO: we should take care of children added into
                    // absoluteBoxes, fixedBoxes && other lists.
                    waitingChildrenCopy = waitingChildren[:]
                    breakFound = false
                    while waitingChildrenCopy {
                        childIndex, child = waitingChildrenCopy.pop()
                        // TODO: should we also accept relative children?
                        if (child.isInNormalFlow() and
                                canBreakInside(child)) {
                                }
                            // We break the waiting child at its last possible
                            // breaking point.
                            // TODO: The dirty solution chosen here is to
                            // decrease the actual size by 1 && render the
                            // waiting child again with this constraint. We may
                            // find a better way.
                            maxX = child.PositionX + child.MarginWidth() - 1
                            childNewChild, childResumeAt, _, _, _, _ = (
                                splitInlineLevel(
                                    context, child, child.PositionX, maxX,
                                    None, box, absoluteBoxes, fixedBoxes,
                                    linePlaceholders, waitingFloats,
                                    lineChildren))
                    }
                }
            }

                            // As PangoLayout && PangoLogAttr don"t always
                            // agree, we have to rely on the actual split to
                            // know whether the child was broken.
                            // https://github.com/Kozea/WeasyPrint/issues/614
                            breakFound = childResumeAt != None
                            if childResumeAt == nil {
                                // PangoLayout decided not to break the child
                                childResumeAt = (0, None)
                            } // TODO: use this when Pango is always 1.40.13+ {
                            } // breakFound = true

                            children = children + waitingChildrenCopy
                            if childNewChild == nil {
                                // May be None where we have an empty TextBox.
                                assert isinstance(child, boxes.TextBox)
                            } else {
                                children += [(childIndex, childNewChild)]
                            }

                            // As this child has already been broken
                            // following the original skip stack, we have to
                            // add the original skip stack to the partial
                            // skip stack we get after the new rendering.

                            // We have to do {
                            } // resumeAt + initialSkipStack
                            // but adding skip stacks is a bit complicated
                            currentSkipStack = initialSkipStack
                            currentResumeAt = (childIndex, childResumeAt)
                            stack = []
                            while currentSkipStack && currentResumeAt {
                                skip, currentSkipStack = (
                                    currentSkipStack)
                                resume, currentResumeAt = (
                                    currentResumeAt)
                                stack.append(skip + resume)
                                if resume != 0 {
                                    break
                                }
                            } resumeAt = currentResumeAt
                            while stack {
                                resumeAt = (stack.pop(), resumeAt)
                            } break
                    if breakFound {
                        break
                    }
                if children {
                    // Too wide, can"t break waiting children && the inline is
                    // non-empty: put child entirely on the next line.
                    resumeAt = (children[-1][0] + 1, None)
                    childWaitingFloats = []
                    break
                }

            positionX = newPositionX
            waitingChildren.append((index, newChild))

        waitingFloats.extend(childWaitingFloats)
        if resumeAt is not None {
            children.extend(waitingChildren)
            resumeAt = (index, resumeAt)
            break
        }
    else {
        children.extend(waitingChildren)
        resumeAt = None
    }

    isEnd = resumeAt == nil
    newBox = box.copyWithChildren(
        [boxChild for index, boxChild := range children],
        isStart=isStart, isEnd=isEnd)
    if isinstance(box, boxes.LineBox) {
        // We must reset line box width according to its new children
        inFlowChildren = [
            boxChild for boxChild := range newBox.children
            if boxChild.isInNormalFlow()]
        if inFlowChildren {
            newBox.Width = (
                inFlowChildren[-1].positionX +
                inFlowChildren[-1].MarginWidth() -
                newBox.PositionX)
        } else {
            newBox.Width = 0
        }
    } else {
        newBox.PositionX = initialPositionX
        if box.Style["boxDecorationBreak"] == "clone" {
            translationNeeded = true
        } else {
            translationNeeded = (
                isStart if box.Style["direction"] == "ltr" else isEnd)
        } if translationNeeded {
            for child := range newBox.children {
                child.translate(dx=leftSpacing)
            }
        } newBox.Width = positionX - contentBoxLeft
        newBox.translate(dx=floatWidths["left"], ignoreFloats=true)
    }

    lineHeight, newBox.Baseline = strutLayout(box.Style, context)
    newBox.Height = box.Style["fontSize"]
    halfLeading = (lineHeight - newBox.Height) / 2.
    // Set margins to the half leading but also compensate for borders and
    // paddings. We want marginHeight() == lineHeight
    newBox.MarginTop = (halfLeading - newBox.BorderTopWidth -
                          newBox.PaddingTop)
    newBox.MarginBottom = (halfLeading - newBox.BorderBottomWidth -
                             newBox.PaddingBottom)

    if newBox.style["position"] == "relative" {
        for absoluteBox := range absoluteBoxes {
            absoluteLayout(context, absoluteBox, newBox, fixedBoxes)
        }
    }

    if resumeAt is not None {
        if resumeAt[0] < floatResumeAt {
            resumeAt = (floatResumeAt, None)
        }
    }

    return (
        newBox, resumeAt, preservedLineBreak, firstLetter, lastLetter,
        floatWidths)
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
        return nil,nil, false
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
        preservedLineBreak = (length != resumeAt) && len(strings.Trim(between," "))  != 0 
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
    box Box
   max ,min  pr.MaybeFloat
}

// Handle ``vertical-align`` within an :class:`LineBox` (or of a
//     non-align sub-tree).
//     Place all boxes vertically assuming that the baseline of ``box``
//     is at `y = 0`.
//     Return ``(maxY, minY)``, the maximum and minimum vertical position
//     of margin boxes.
func lineBoxVerticality(box Box) {
    var topBottomSubtrees []Box
    maxY, minY := alignedSubtreeVerticality(box, &topBottomSubtrees, 0)
    subtreesWithMinMax :=make([]boxMinMax, len(topBottomSubtrees))
    for i, subtree := range topBottomSubtrees {
        var subMaxY, subMinY pr.MaybeFloat
        if !subtree.Box().IsFloated() {
             subMaxY, subMinY =  alignedSubtreeVerticality(subtree, topBottomSubtrees, baselineY=0)
        }
        subtreesWithMinMax[i] = boxMinMax{box:subtree, max:subMaxY, min; subMinY}
    }
 
    if len(subtreesWithMinMax) != 0{
        var highestSub float32
        for _, v := range subtreesWithMinMax {
            if ! subtree.Box().IsFloated() {
            m := v.max.V() - v.min.V()
            if m > highestSub {
                highestSub = m
            }
        }
    }            
            maxY = utils.Max(maxY.V(), minY.V() + highestSub)
    }

    for _, v := range subtreesWithMinMax {
        va := v.box.Box().Style.GetVerticalAlign()
        var dy float32
        if v.box.Box().IsFloated() {
            dy = minY - v.Box.positionY
        } else if va.String == "top" {
            dy = minY - subMinY
        } else if va.String == "bottom"{
            dy = maxY - subMaxY
             }else {
                 log.Fatalf("expected top or bottom, got %v", va)
        } 
        translateSubtree(v.box, dy)
    } 
    return maxY, minY
        }

func translateSubtree(box Box, dy float32) {
    if bo.TypeInlineBox.IsInstance(box)  {
        box.Box().PositionY += dy
        if va :=  box.Box().Style.GetVerticalAlign(); va ==  "top" || va == "bottom" {
            for  _, child := range box.Box().Children {
                translateSubtree(child, dy)
            }
        }
    } else {
        // Text or atomic boxes
        box.Translate(box, 0, dy)
    }
} 

func alignedSubtreeVerticality(box Box, topBottomSubtrees *[]Box, baselineY float32) (maxY , minY pr.MaybeFloat) {
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
func inlineBoxVerticality(box_ Box, topBottomSubtrees *[]Box, baselineY float32) (maxY , minY pr.MaybeFloat) {
    if !isLine(box_) {
        return maxY, minY
    }
    box := box_.Box()
    for _, child_ := range box_.Box().Children {
        child := child_.Box()
        if ! child.IsInNormalFlow() {
            if child.IsFloated() {
                *topBottomSubtrees = append(*topBottomSubtrees,child)
            } 
            continue
        } 
        var childBaselineY float32
        verticalAlign := child.Style.GetVerticalAlign()
        switch verticalAlign.String {
        case "baseline" :
            childBaselineY = baselineY
        case "middle" :
            oneEx := box.Style.GetFontSize().Value * pdf.ExRatio(box.Style)
            top = baselineY - (oneEx + child.MarginHeight()) / 2.
            childBaselineY = top + child.Baseline
        case "text-top" :
            // align top with the top of the parent’s content area
            top = (baselineY - box.Baseline + box.MarginTop +
                   box.BorderTopWidth.V() + box.PaddingTop)
            childBaselineY = top + child.Baseline
        case "text-bottom" :
            // align bottom with the bottom of the parent’s content area
            bottom = (baselineY - box.Baseline + box.MarginTop +
                      box.BorderTopWidth.V() + box.PaddingTop + box.Height)
            childBaselineY = bottom - child.MarginHeight() + child.Baseline
            case "top", "bottom" :
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
        if bo.TypeInlineBlockBox.IsInstance(child_) || bo.TypeInlineFlexBox.IsInstance(child_)   {
            // This also includes table wrappers for inline tables.
            child_.Translate(child_, 0, top - child.PositionY)
        } else {
            child.PositionY = top
            // grand-children for inline boxes are handled below
        }

        if verticalAlign == "top" || verticalAlign ==  "bottom" {
            // top || bottom are special, they need to be handled in
            // a later pass.
            *topBottomSubtrees = append(*topBottomSubtrees,child)
            continue
        }

        bottom = top + child.MarginHeight()
        if minY == nil || top < minY {
            minY = p.Float(top)
        } 
        if maxY == nil || bottom > maxY {
            maxY = p.Float(bottom)
        } 
        if bo.TypeInlineBox.IsInstance(child_)  {
            childrenMaxY, childrenMinY := inlineBoxVerticality( child, topBottomSubtrees, childBaselineY)
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
//     the `text-align` property.
func textAlign(context LayoutContext, line_ Box, availableWidth float32, last bool) float32 {
    line := line_.Box()

    // "When the total width of the inline-level boxes on a line is less than
    // the width of the line box containing them, their horizontal distribution
    // within the line box is determined by the "text-align" property."
    if line.Width.V() >= availableWidth {
        return 0
    }

    align := line.Style.GetTextAlign()
    ws := line.Style.GetWhitespace()
    spaceCollapse = ws == "normal" || ws == "nowrap" || ws == "pre-line"
    if align == "-weasy-start" || align ==  "-weasy-end" {
        if (align == "-weasy-start") != (line.Style.GetDirection() == "rtl") { // xor
            align = "left"
        } else {
            align = "right"
        }
    }
     if align == "justify" && last {
        align = "left"
         if line.Style.GetDirection() == "rtl"  {
             align =  "right"
         }
    }
     if align == "left" {
        return 0
    } 
    offset := availableWidth - line.width
    if align == "justify" {
        if spaceCollapse {
            // Justification of texts where white space is not collapsing is
            // - forbidden by CSS 2, and
            // - not required by CSS 3 Text.
            justifyLine(context, line, offset)
        } 
        return 0
    } 
    if align == "center" {
        return offset / 2
    } else if  align == "right" {
        return offset
    } else {
        log.Fatalf("align should be center or right, got %s", align)
        return 0
    }
}

func justifyLine(context LayoutContext, line Box, extraWidth float32) {
    // TODO: We should use a better alorithm here, see
    // https://www.w3.org/TR/css-text-3/#justify-algos
    nbSpaces := countSpaces(line)
    if nbSpaces == 0 {
        return
    } 
    addWordSpacing(context, line, extraWidth / float32(nbSpaces), 0)
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

func addWordSpacing(context LayoutContext, box_ Box, justificationSpacing float32, xAdvance float32) float32 {
    if textBox, isTextBox := box_.(*bo.TextBox); isTextBox  {
        textBox.JustificationSpacing = justificationSpacing
        textBox.PositionX += xAdvance
        nbSpaces := countSpaces(box_)
        if nbSpaces > 0 {
            layout = createLayout(textBox.Text, textBox.Style, context, pr.Inf, textBox.JustificationSpacing)
            layout.Deactivate()
            extraSpace := justificationSpacing * nbSpaces
            xAdvance += extraSpace
            textBox.Width = textBox.Width.V() + extraSpace
            textBox.PangoLayout = layout
        }
    } else if isLine(box_)  {
        box := box_.Box()
        box.PositionX += xAdvance
        previousXAdvance := xAdvance
        for _, child := range box.Children {
            if child.IsInNormalFlow() {
                xAdvance = addWordSpacing(context, child, justificationSpacing, xAdvance)
            }
        } 
        box.Width = box.Width.V() + xAdvance - previousXAdvance
    } else {
        // Atomic inline-level box
        box.Translate(box, xAdvance, 0)
    } 
    return xAdvance
} 

// http://www.w3.org/TR/CSS21/visuren.html#phantom-line-box
func isPhantomLinebox(linebox Box) bool {
    for _,  child := range linebox.Box().Children {
        if bo.TypeInlineBox.IsInstance(child) {
            if ! isPhantomLinebox(child) {
                return false
			}
			 for _, side := range ("top", "right", "bottom", "left") {
				 m := child.Box().Style["margin_" + side].(pr.Value).Value
				 b := child.Box().Style["border_" + side + "_width"].(pr.Value)
				 p := child.Box().Style["padding_" + side].(pr.Value).Value
                if m != 0 || !b.IsNone() || p != 0 {
                        
					return false
				}
            }
        } else if child.Box().IsInNormalFlow() {
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