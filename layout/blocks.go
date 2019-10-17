package layout

import (
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	bo "github.com/benoitkugler/go-weasyprint/boxes"
)

// Page breaking and layout for block-level and block-container boxes.

type page struct {
	break_ string
	page   interface{}
}

type blockLayout struct {
	newBox            bo.Box
	resumeAt          *bo.SkipStack
	nextPage          page
	adjoiningMargins  []pr.Float
	collapsingThrough bool
}


// Lay out the block-level ``box``.
// 
// `maxPositionY` is the absolute vertical position (as in
// ``someBox.PositionY``) of the bottom of the
// content box of the current page area.
func blockLevelLayout(context LayoutContext, box_ bo.InstanceBlockLevelBox, maxPositionY pr.Float, skipStack *bo.SkipStack,
                       containingBlock Box, pageIsEmpty bool, absoluteBoxes []*AbsolutePlaceholder,
                       fixedBoxes []Box, adjoiningMargins []pr.Float ) blockLayout{

						box := box_.Box()
    if ! bo.TypeTableBox.IsInstance(box_) {
        resolvePercentages2(box_, containingBlock, "")
    
        if box.MarginTop.Auto() {
            box.MarginTop = pr.Float(0)
		} 
		if box.MarginBottom.Auto() {
            box.MarginBottom = pr.Float(0)
        }

        if (context.currentPage > 1 && pageIsEmpty) {
            // TODO: we should take care of cases when this box doesn't have
            // collapsing margins with the first child of the page, see
            // testMarginBreakClearance.
            if box.Style.GetMarginBreak() == "discard" {
                box.MarginTop = pr.Float(0)
            } else if box.Style.GetMarginBreak() == "auto" {
                if ! context.forcedBreak {
                    box.MarginTop = pr.Float(0)
                }
            }
        }

		collapsedMargin := collapseMargin(append(adjoiningMargins ,box.MarginTop))
		bl := box_.BlockLevel()
        bl.Clearance = getClearance(context, box_, collapsedMargin)
        if bl.Clearance  != nil  {
            topBorderEdge = box.PositionY + collapsedMargin + bl.Clearance.V()
            box.PositionY = topBorderEdge - box.MarginTop
            adjoiningMargins = nil
        }
	}
    return blockLevelLayoutSwitch( context, box, maxPositionY, skipStack, containingBlock,
        pageIsEmpty, absoluteBoxes, fixedBoxes, adjoiningMargins)
	}

// Call the layout function corresponding to the ``box`` type.
func blockLevelLayoutSwitch(context LayoutContext, box bo.InstanceBlockLevelBox, maxPositionY pr.Float, skipStack *bo.SkipStack,
                              containingBlock Box, pageIsEmpty bool, absoluteBoxes,
                              fixedBoxes, adjoiningMargins []Box) blockLayout {
                         
    if bo.TypeTableBox.IsInstance(box) {
        return tableLayout(context, box, maxPositionY, skipStack, containingBlock,
            pageIsEmpty, absoluteBoxes, fixedBoxes)
    } else if bo.TypeBlockBox.IsInstance(box) {
        return blockBoxLayout(context, box, maxPositionY, skipStack, containingBlock,
            pageIsEmpty, absoluteBoxes, fixedBoxes, adjoiningMargins)
    } else if isinstance(box, boxes.BlockReplacedBox) {
        box = blockReplacedBoxLayout(box, containingBlock)
        // Don't collide with floats
        // http://www.w3.org/TR/CSS21/visuren.html#floats
        box.PositionX, box.PositionY, _ = avoidCollisions(context, box, containingBlock, false)
        nextPage := page{break_: "any", page: nil}
        return blockLayout{
			box:box, resumeAt:nil, nextPage:nextPage, adjoiningMargins:nil, collapsingThrough:false,
		} 
    } else if bo.TypeFlexBox.IsInstance(box) {
        return flexLayout(context, box, maxPositionY, skipStack, containingBlock,
            pageIsEmpty, absoluteBoxes, fixedBoxes)
    } else {  // pragma: no cover
		log.Fatalf("Layout for %s not handled yet", box)
		return blockLayout{}
		}
	}

// Lay out the block ``box``.
func blockBoxLayout(context LayoutContext, box_ bo.InstanceBlockBox, maxPositionY pr.Float, skipStack *bo.SkipStack,
                     containingBlock Box, pageIsEmpty bool, absoluteBoxes []*AbsolutePlaceholder,fixedBoxes, adjoiningMargins []Box) blockLayout {
                     box := box_.Box()
    if box.Style.GetColumnWidth().String != "auto" || box.Style.GetColumnCount().String != "auto" {
        result := columnsLayout(context, box, maxPositionY, skipStack, containingBlock,
            pageIsEmpty, absoluteBoxes, fixedBoxes, adjoiningMargins)

        resumeAt = result[1]
        // TODO: this condition && the whole relayout are probably wrong
        if resumeAt  == nil  {
            newBox = result[0]
            bottomSpacing = (
                newBox.MarginBottom + newBox.PaddingBottom +
                newBox.borderBottomWidth)
            if bottomSpacing {
                maxPositionY -= bottomSpacing
                result = columnsLayout(context, box, maxPositionY, skipStack,
                    containingBlock, pageIsEmpty, absoluteBoxes,
                    fixedBoxes, adjoiningMargins)
            }
        }

        return result
    else if box.isTableWrapper {
        tableWrapperWidth(
            context, box, (containingBlock.width, containingBlock.height))
    } blockLevelWidth(box, containingBlock)

    newBox, resumeAt, nextPage, adjoiningMargins, collapsingThrough = \
        blockContainerLayout(
            context, box, maxPositionY, skipStack, pageIsEmpty,
            absoluteBoxes, fixedBoxes, adjoiningMargins)
    if newBox && newBox.isTableWrapper {
        // Don"t collide with floats
        // http://www.w3.org/TR/CSS21/visuren.html#floats
        positionX, positionY, _ = avoidCollisions(
            context, newBox, containingBlock, outer=false)
        newBox.translate(
            positionX - newBox.PositionX, positionY - newBox.PositionY)
    } return newBox, resumeAt, nextPage, adjoiningMargins, collapsingThrough


@handleMinMaxWidth
func blockReplacedWidth(box, containingBlock) {
    // http://www.w3.org/TR/CSS21/visudet.html#block-replaced-width
    replacedBoxWidth.withoutMinMax(box, containingBlock)
    blockLevelWidth.withoutMinMax(box, containingBlock)
} 

// Lay out the block :class:`boxes.ReplacedBox` ``box``.
func blockReplacedBoxLayout(box bo.InstanceBlockReplaced, containingBlock) {
    box = box.copy()
    if box.Style["width"] == "auto" && box.Style["height"] == "auto" {
        computedMargins = box.MarginLeft, box.MarginRight
        blockReplacedWidth.withoutMinMax(
            box, containingBlock)
        replacedBoxHeight.withoutMinMax(box)
        minMaxAutoReplaced(box)
        box.MarginLeft, box.MarginRight = computedMargins
        blockLevelWidth.withoutMinMax(box, containingBlock)
    } else {
        blockReplacedWidth(box, containingBlock)
        replacedBoxHeight(box)
    }
} 
    return box


@handleMinMaxWidth
// Set the ``box`` width.
func blockLevelWidth(box, containingBlock) {
    // "cb" stands for "containing block"
    cbWidth = containingBlock.width
} 
    // http://www.w3.org/TR/CSS21/visudet.html#blockwidth

    // These names are waaay too long
    marginL = box.MarginLeft
    marginR = box.MarginRight
    paddingL = box.PaddingLeft
    paddingR = box.PaddingRight
    borderL = box.borderLeftWidth
    borderR = box.borderRightWidth
    width = box.width

    // Only margin-left, margin-right && width can be "auto".
    // We want:  width of containing block ==
    //               margin-left + border-left-width + padding-left + width
    //               + padding-right + border-right-width + margin-right

    paddingsPlusBorders = paddingL + paddingR + borderL + borderR
    if box.width != "auto" {
        total = paddingsPlusBorders + width
        if marginL != "auto" {
            total += marginL
        } if marginR != "auto" {
            total += marginR
        } if total > cbWidth {
            if marginL == "auto" {
                marginL = box.MarginLeft = 0
            } if marginR == "auto" {
                marginR = box.MarginRight = 0
            }
        }
    } if width != "auto" && marginL != "auto" && marginR != "auto" {
        // The equation is over-constrained.
        if containingBlock.style["direction"] == "rtl" && ! box.isColumn {
            box.PositionX += (
                cbWidth - paddingsPlusBorders - width - marginR - marginL)
        } // Do nothing := range ltr.
    } if width == "auto" {
        if marginL == "auto" {
            marginL = box.MarginLeft = 0
        } if marginR == "auto" {
            marginR = box.MarginRight = 0
        } width = box.width = cbWidth - (
            paddingsPlusBorders + marginL + marginR)
    } marginSum = cbWidth - paddingsPlusBorders - width
    if marginL == "auto" && marginR == "auto" {
        box.MarginLeft = marginSum / 2.
        box.MarginRight = marginSum / 2.
    } else if marginL == "auto" && marginR != "auto" {
        box.MarginLeft = marginSum - marginR
    } else if marginL != "auto" && marginR == "auto" {
        box.MarginRight = marginSum - marginL
    }


// Translate the ``box`` if it is relatively positioned.
func relativePositioning(box, containingBlock) {
    if box.Style["position"] == "relative" {
        resolvePositionPercentages(box, containingBlock)
    }
} 
        if box.left != "auto" && box.right != "auto" {
            if box.Style["direction"] == "ltr" {
                translateX = box.left
            } else {
                translateX = -box.right
            }
        } else if box.left != "auto" {
            translateX = box.left
        } else if box.right != "auto" {
            translateX = -box.right
        } else {
            translateX = 0
        }

        if box.top != "auto" {
            translateY = box.top
        } else if box.Style["bottom"] != "auto" {
            translateY = -box.bottom
        } else {
            translateY = 0
        }

        box.translate(translateX, translateY)

    if isinstance(box, (boxes.InlineBox, boxes.LineBox)) {
        for child := range box.children {
            relativePositioning(child, containingBlock)
        }
    }


// Set the ``box`` height.
func blockContainerLayout(context *LayoutContext, box bo.Box, maxPositionY float32, skipStack *bo.SkipStack,
	pageIsEmpty bool, absoluteBoxes *[]*AbsolutePlaceholder, fixedBoxes []bo.Box, adjoiningMargins []float32) blockLayout {
	// FIXME: a implémenter
	return blockLayout{}

    // TODO: boxes.FlexBox is allowed here because flexLayout calls
    // blockContainerLayout, there"s probably a better solution.
    assert isinstance(box, (boxes.BlockContainerBox, boxes.FlexBox))

    // We have to work around floating point rounding errors here.
    // The 1e-9 value comes from PEP 485.
    allowedMaxPositionY = maxPositionY * (1 + 1e-9)

    // See http://www.w3.org/TR/CSS21/visuren.html#block-formatting
    if ! isinstance(box, boxes.BlockBox) {
        context.createBlockFormattingContext()
    }

    isStart = skipStack  == nil 
    if box.Style["boxDecorationBreak"] == "slice" && ! isStart {
        // Remove top margin, border && padding {
        } box.RemoveDecoration(start=true, end=false)
    }

    if adjoiningMargins  == nil  {
        adjoiningMargins = []
    }

    if box.Style["boxDecorationBreak"] == "clone" {
        maxPositionY -= (
            box.PaddingBottom + box.borderBottomWidth +
            box.MarginBottom)
    }

    adjoiningMargins.append(box.MarginTop)
    thisBoxAdjoiningMargins = adjoiningMargins

    collapsingWithChildren = ! (
        box.borderTopWidth || box.PaddingTop || box.isFlexItem or
        establishesFormattingContext(box) || box.isForRootElement)
    if collapsingWithChildren {
        // XXX ! counting margins := range adjoiningMargins, if any
        // (There are ! padding || borders, see above.)
        positionY = box.PositionY
    } else {
        box.PositionY += collapseMargin(adjoiningMargins) - box.MarginTop
        adjoiningMargins = []
        positionY = box.contentBoxY()
    }

    positionX = box.contentBoxX()

    if box.Style["position"] == "relative" {
        // New containing block, use a new absolute list
        absoluteBoxes = []
    }

    newChildren = []
    nextPage = {"break": "any", "page": None}

    lastInFlowChild = None

    if isStart {
        skip = 0
        firstLetterStyle = getattr(box, "firstLetterStyle", None)
    } else {
        skip, skipStack = skipStack
        firstLetterStyle = None
    } for i, child := range enumerate(box.children[skip:]) {
        index = i + skip
        child.positionX = positionX
        // XXX does ! count margins := range adjoiningMargins {
        } child.positionY = positionY
    }

        if ! child.isInNormalFlow() {
            child.positionY += collapseMargin(adjoiningMargins)
            if child.isAbsolutelyPositioned() {
                placeholder = AbsolutePlaceholder(child)
                placeholder.index = index
                newChildren.append(placeholder)
                if child.style["position"] == "absolute" {
                    absoluteBoxes.append(placeholder)
                } else {
                    fixedBoxes.append(placeholder)
                }
            } else if child.isFloated() {
                newChild = floatLayout(
                    context, child, box, absoluteBoxes, fixedBoxes)
                // New page if overflow
                if (pageIsEmpty && ! newChildren) || ! (
                        newChild.positionY + newChild.height >
                        allowedMaxPositionY) {
                        }
                    newChild.index = index
                    newChildren.append(newChild)
                else {
                    for previousChild := range reversed(newChildren) {
                        if previousChild.isInNormalFlow() {
                            lastInFlowChild = previousChild
                            break
                        }
                    } pageBreak = blockLevelPageBreak(
                        lastInFlowChild, child)
                    if newChildren && pageBreak := range ("avoid", "avoid-page") {
                        result = findEarlierPageBreak(
                            newChildren, absoluteBoxes, fixedBoxes)
                        if result {
                            newChildren, resumeAt = result
                            break
                        }
                    } resumeAt = (index, None)
                    break
                }
            } else if child.isRunning() {
                context.runningElements.setdefault(
                    child.style["position"][1], {}
                )[context.currentPage - 1] = child
            } continue
        }

        if isinstance(child, boxes.LineBox) {
            assert len(box.children) == 1, (
                "line box with siblings before layout")
            if adjoiningMargins {
                positionY += collapseMargin(adjoiningMargins)
                adjoiningMargins = []
            } newContainingBlock = box
            linesIterator = iterLineBoxes(
                context, child, positionY, skipStack,
                newContainingBlock, absoluteBoxes, fixedBoxes,
                firstLetterStyle)
            isPageBreak = false
            for line, resumeAt := range linesIterator {
                line.resumeAt = resumeAt
                newPositionY = line.positionY + line.height
            }
        }

                // Add bottom padding && border to the bottom position of the
                // box if needed
                if resumeAt  == nil  || (
                        box.Style["boxDecorationBreak"] == "clone") {
                        }
                    offsetY = box.borderBottomWidth + box.PaddingBottom
                else {
                    offsetY = 0
                }

                // Allow overflow if the first line of the page is higher
                // than the page itself so that we put *something* on this
                // page && can advance := range the context.
                if newPositionY + offsetY > allowedMaxPositionY && (
                        newChildren || ! pageIsEmpty) {
                        }
                    overOrphans = len(newChildren) - box.Style["orphans"]
                    if overOrphans < 0 && ! pageIsEmpty {
                        // Reached the bottom of the page before we had
                        // enough lines for orphans, cancel the whole box.
                        page = child.pageValues()[0]
                        return (
                            None, None, {"break": "any", "page": page}, [],
                            false)
                    } // How many lines we need on the next page to satisfy widows
                    // -1 for the current line.
                    needed = box.Style["widows"] - 1
                    if needed {
                        for _ := range linesIterator {
                            needed -= 1
                            if needed == 0 {
                                break
                            }
                        }
                    } if needed > overOrphans && ! pageIsEmpty {
                        // Total number of lines < orphans + widows
                        page = child.pageValues()[0]
                        return (
                            None, None, {"break": "any", "page": page}, [],
                            false)
                    } if needed && needed <= overOrphans {
                        // Remove lines to keep them for the next page
                        del newChildren[-needed:]
                    } // Page break here, resume before this line
                    resumeAt = (index, skipStack)
                    isPageBreak = true
                    break
                // TODO: this is incomplete.
                // See http://dev.w3.org/csswg/css3-page/#allowed-pg-brk
                // "When an unforced page break occurs here, both the adjoining
                //  ‘margin-top’ && ‘margin-bottom’ are set to zero."
                // See https://github.com/Kozea/WeasyPrint/issues/115
                else if pageIsEmpty && newPositionY > allowedMaxPositionY {
                    // Remove the top border when a page is empty && the box is
                    // too high to be drawn := range one page
                    newPositionY -= box.MarginTop
                    line.translate(0, -box.MarginTop)
                    box.MarginTop = 0
                } newChildren.append(line)
                positionY = newPositionY
                skipStack = resumeAt
            if newChildren {
                resumeAt = (index, newChildren[-1].resumeAt)
            } if isPageBreak {
                break
            }
        else {
            for previousChild := range reversed(newChildren) {
                if previousChild.isInNormalFlow() {
                    lastInFlowChild = previousChild
                    break
                }
            } else {
                lastInFlowChild = None
            } if lastInFlowChild  != nil  {
                // Between in-flow siblings
                pageBreak = blockLevelPageBreak(lastInFlowChild, child)
                pageName = blockLevelPageName(lastInFlowChild, child)
                if pageName || pageBreak := range (
                        "page", "left", "right", "recto", "verso") {
                        }
                    pageName = child.pageValues()[0]
                    nextPage = {"break": pageBreak, "page": pageName}
                    resumeAt = (index, None)
                    break
            } else {
                pageBreak = "auto"
            }
        }

            newContainingBlock = box

            if ! newContainingBlock.isTableWrapper {
                resolvePercentages(child, newContainingBlock)
                if (child.isInNormalFlow() and
                        lastInFlowChild  == nil  and
                        collapsingWithChildren) {
                        }
                    // TODO: add the adjoining descendants" margin top to
                    // [child.marginTop]
                    oldCollapsedMargin = collapseMargin(adjoiningMargins)
                    if child.marginTop == "auto" {
                        childMarginTop = 0
                    } else {
                        childMarginTop = child.marginTop
                    } newCollapsedMargin = collapseMargin(
                        adjoiningMargins + [childMarginTop])
                    collapsedMarginDifference = (
                        newCollapsedMargin - oldCollapsedMargin)
                    for previousNewChild := range newChildren {
                        previousNewChild.translate(
                            dy=collapsedMarginDifference)
                    } clearance = getClearance(
                        context, child, newCollapsedMargin)
                    if clearance  != nil  {
                        for previousNewChild := range newChildren {
                            previousNewChild.translate(
                                dy=-collapsedMarginDifference)
                        }
                    }
            }

                        collapsedMargin = collapseMargin(adjoiningMargins)
                        box.PositionY += collapsedMargin - box.MarginTop
                        // Count box.MarginTop as we emptied adjoiningMargins
                        adjoiningMargins = []
                        positionY = box.contentBoxY()

            if adjoiningMargins && box.isTableWrapper {
                collapsedMargin = collapseMargin(adjoiningMargins)
                child.positionY += collapsedMargin
                positionY += collapsedMargin
                adjoiningMargins = []
            }

            pageIsEmptyWithNoChildren = pageIsEmpty && ! any(
                child for child := range newChildren
                if ! isinstance(child, AbsolutePlaceholder))

            if ! getattr(child, "firstLetterStyle", None) {
                child.firstLetterStyle = firstLetterStyle
            } (newChild, resumeAt, nextPage, nextAdjoiningMargins,
                collapsingThrough) = blockLevelLayout(
                    context, child, maxPositionY, skipStack,
                    newContainingBlock, pageIsEmptyWithNoChildren,
                    absoluteBoxes, fixedBoxes, adjoiningMargins)
            skipStack = None

            if newChild  != nil  {
                // index := range its non-laid-out parent, ! := range future new parent
                // May be used := range findEarlierPageBreak()
                newChild.index = index
            }

                // We need to do this after the child layout to have the
                // used value for marginTop (eg. it might be a percentage.)
                if ! isinstance(
                        newChild, (boxes.BlockBox, boxes.TableBox)) {
                        }
                    adjoiningMargins.append(newChild.marginTop)
                    offsetY = (
                        collapseMargin(adjoiningMargins) -
                        newChild.marginTop)
                    newChild.translate(0, offsetY)
                    adjoiningMargins = []
                // else: blocks handle that themselves.

                adjoiningMargins = nextAdjoiningMargins
                adjoiningMargins.append(newChild.marginBottom)

                if ! collapsingThrough {
                    newPositionY = (
                        newChild.borderBoxY() + newChild.borderHeight())
                }

                    if (newPositionY > allowedMaxPositionY and
                            ! pageIsEmptyWithNoChildren) {
                            }
                        // The child overflows the page area, put it on the
                        // next page. (But don’t delay whole blocks if eg.
                        // only the bottom border overflows.)
                        newChild = None
                    else {
                        positionY = newPositionY
                    }

                if newChild  != nil  && newChild.clearance  != nil  {
                    positionY = (
                        newChild.borderBoxY() + newChild.borderHeight())
                }

            if newChild  == nil  {
                // Nothing fits := range the remaining space of this page: break
                if pageBreak := range ("avoid", "avoid-page") {
                    // TODO: fill the blank space at the bottom of the page
                    result = findEarlierPageBreak(
                        newChildren, absoluteBoxes, fixedBoxes)
                    if result {
                        newChildren, resumeAt = result
                        break
                    } else {
                        // We did ! find any page break opportunity
                        if ! pageIsEmpty {
                            // The page has content *before* this block {
                            } // cancel the block && try to find a break
                            // := range the parent.
                            page = child.pageValues()[0]
                            return (
                                None, None, {"break": "any", "page": page}, [],
                                false)
                        } // else {
                        } // ignore this "avoid" && break anyway.
                    }
                }
            }

                if all(child.isAbsolutelyPositioned()
                       for child := range newChildren) {
                       }
                    // This box has only rendered absolute children, keep them
                    // for the next page. This is for example useful for list
                    // markers.
                    removePlaceholders(
                        newChildren, absoluteBoxes, fixedBoxes)
                    newChildren = []

                if newChildren {
                    resumeAt = (index, None)
                    break
                } else {
                    // This was the first child of this box, cancel the box
                    // completly
                    page = child.pageValues()[0]
                    return (
                        None, None, {"break": "any", "page": page}, [], false)
                }

            // Bottom borders may overflow here
            // TODO: back-track somehow when all lines fit but ! borders
            newChildren.append(newChild)
            if resumeAt  != nil  {
                resumeAt = (index, resumeAt)
                break
            }
    else {
        resumeAt = None
    }

    if (resumeAt  != nil  and
            box.Style["breakInside"] := range ("avoid", "avoid-page") and
            ! pageIsEmpty) {
            }
        return (
            None, None, {"break": "any", "page": None}, [], false)

    if collapsingWithChildren {
        box.PositionY += (
            collapseMargin(thisBoxAdjoiningMargins) - box.MarginTop)
    }

    for previousChild := range reversed(newChildren) {
        if previousChild.isInNormalFlow() {
            lastInFlowChild = previousChild
            break
        }
    } else {
        lastInFlowChild = None
    } collapsingThrough = false
    if lastInFlowChild  == nil  {
        collapsedMargin = collapseMargin(adjoiningMargins)
        // top && bottom margin of this box
        if (box.height := range ("auto", 0) and
            getClearance(context, box, collapsedMargin)  == nil  and
            all(v == 0 for v := range [
                box.MinHeight, box.borderTopWidth, box.PaddingTop,
                box.borderBottomWidth, box.PaddingBottom])) {
                }
            collapsingThrough = true
        else {
            positionY += collapsedMargin
            adjoiningMargins = []
        }
    } else {
        // bottom margin of the last child && bottom margin of this box ...
        if box.height != "auto" {
            // ! adjoining. (positionY is ! used afterwards.)
            adjoiningMargins = []
        }
    }

    if (box.borderBottomWidth or
            box.PaddingBottom or
            establishesFormattingContext(box) or
            box.isForRootElement or
            box.isTableWrapper) {
            }
        positionY += collapseMargin(adjoiningMargins)
        adjoiningMargins = []

    newBox = box.copyWithChildren(
        newChildren, isStart=isStart, isEnd=resumeAt  == nil )

    // TODO: See corner cases in
    // http://www.w3.org/TR/CSS21/visudet.html#normal-block
    // TODO: See float.floatLayout
    if newBox.height == "auto" {
        if context.excludedShapes && newBox.Style["overflow"] != "visible" {
            maxFloatPositionY = max(
                floatBox.PositionY + floatBox.MarginHeight()
                for floatBox := range context.excludedShapes)
            positionY = max(maxFloatPositionY, positionY)
        } newBox.height = positionY - newBox.contentBoxY()
    }

    if newBox.Style["position"] == "relative" {
        // New containing block, resolve the layout of the absolute descendants
        for absoluteBox := range absoluteBoxes {
            absoluteLayout(context, absoluteBox, newBox, fixedBoxes)
        }
    }

    for child := range newBox.children {
        relativePositioning(child, (newBox.width, newBox.height))
    }

    if ! isinstance(newBox, boxes.BlockBox) {
        context.finishBlockFormattingContext(newBox)
    }

    if resumeAt  == nil  {
        // After finishBlockFormattingContext which may increment
        // newBox.height
        newBox.height = max(
            min(newBox.height, newBox.MaxHeight),
            newBox.MinHeight)
    } else {
        // Make the box fill the blank space at the bottom of the page
        // https://www.w3.org/TR/css-break-3/#box-splitting
        newBox.height = (
            maxPositionY - newBox.PositionY -
            (newBox.MarginHeight() - newBox.height))
        if box.Style["boxDecorationBreak"] == "clone" {
            newBox.height += (
                box.PaddingBottom + box.borderBottomWidth +
                box.MarginBottom)
        }
    }

    if nextPage["page"]  == nil  {
        nextPage["page"] = newBox.PageValues()[1]
    }

    return newBox, resumeAt, nextPage, adjoiningMargins, collapsingThrough


// Return the amount of collapsed margin for a list of adjoining margins.
func collapseMargin(adjoiningMargins []pr.Float) pr.Float {
	var maxPos, minNeg pr.Float
	for _, m := range adjoiningMargins {
		if m > maxPos {
			maxPos = m
		} else if  m < minNeg {
			minNeg = m
		}
	}
	return maxPos + minNeg
} 

// Return wether a box establishes a block formatting context.
//     See http://www.w3.org/TR/CSS2/visuren.html#block-formatting
//     
func establishesFormattingContext(box) {
    return (
        box.isFloated()
    ) || (
        box.isAbsolutelyPositioned()
    ) || (
        // TODO: columns shouldn"t be block boxes, this condition would then be
        // useless when this is fixed
        box.isColumn
    ) || (
        isinstance(box, boxes.BlockContainerBox) and
        ! isinstance(box, boxes.BlockBox)
    ) || (
        isinstance(box, boxes.BlockBox) && box.Style["overflow"] != "visible"
    )
} 

// Return the value of ``page-break-before`` || ``page-break-after``
//     that "wins" for boxes that meet at the margin between two sibling boxes.
//     For boxes before the margin, the "page-break-after" value is considered;
//     for boxes after the margin the "page-break-before" value is considered.
//     * "avoid" takes priority over "auto"
//     * "page" takes priority over "avoid" || "auto"
//     * "left" || "right" take priority over "always", "avoid" || "auto"
//     * Among "left" && "right", later values := range the tree take priority.
//     See http://dev.w3.org/csswg/css3-page/#allowed-pg-brk
//     
func blockLevelPageBreak(siblingBefore, siblingAfter) {
    values = []
    // https://drafts.csswg.org/css-break-3/#possible-breaks
    blockParallelBoxTypes = (
        boxes.BlockLevelBox, boxes.TableRowGroupBox, boxes.TableRowBox)
} 
    box = siblingBefore
    while isinstance(box, blockParallelBoxTypes) {
        values.append(box.Style["breakAfter"])
        if ! (isinstance(box, boxes.ParentBox) && box.children) {
            break
        } box = box.children[-1]
    } values.reverse()  // Have them := range tree order

    box = siblingAfter
    while isinstance(box, blockParallelBoxTypes) {
        values.append(box.Style["breakBefore"])
        if ! (isinstance(box, boxes.ParentBox) && box.children) {
            break
        } box = box.children[0]
    }

    result = "auto"
    for value := range values {
        if value := range ("left", "right", "recto", "verso") || (value, result) := range (
                ("page", "auto"),
                ("page", "avoid"),
                ("avoid", "auto"),
                ("page", "avoid-page"),
                ("avoid-page", "auto")) {
                }
            result = value
    }

    return result


// Return the next page name when siblings don"t have the same names.
func blockLevelPageName(siblingBefore, siblingAfter) {
    beforePage = siblingBefore.pageValues()[1]
    afterPage = siblingAfter.pageValues()[0]
    if beforePage != afterPage {
        return afterPage
    }
} 

// Because of a `page-break-before: avoid` || a `page-break-after: avoid`
//     we need to find an earlier page break opportunity inside `children`.
//     Absolute || fixed placeholders removed from children should also be
//     removed from `absoluteBoxes` || `fixedBoxes`.
//     Return (newChildren, resumeAt)
//     
func findEarlierPageBreak(children, absoluteBoxes, fixedBoxes) {
    if children && isinstance(children[0], boxes.LineBox) {
        // Normally `orphans` && `widows` apply to the block container, but
        // line boxes inherit them.
        orphans = children[0].style["orphans"]
        widows = children[0].style["widows"]
        index = len(children) - widows  // how many lines we keep
        if index < orphans {
            return None
        } newChildren = children[:index]
        resumeAt = (0, newChildren[-1].resumeAt)
        removePlaceholders(children[index:], absoluteBoxes, fixedBoxes)
        return newChildren, resumeAt
    }
} 
    previousInFlow = None
    for index, child := range reversedEnumerate(children) {
        if child.isInNormalFlow() {
            if previousInFlow  != nil  && (
                    blockLevelPageBreak(child, previousInFlow) ! in
                    ("avoid", "avoid-page")) {
                    }
                index += 1  // break after child
                newChildren = children[:index]
                // Get the index := range the original parent
                resumeAt = (children[index].index, None)
                break
            previousInFlow = child
        } if child.isInNormalFlow() && (
                child.style["breakInside"] ! := range ("avoid", "avoid-page")) {
                }
            breakableBoxTypes = (
                boxes.BlockBox, boxes.TableBox, boxes.TableRowGroupBox)
            if isinstance(child, breakableBoxTypes) {
                result = findEarlierPageBreak(
                    child.children, absoluteBoxes, fixedBoxes)
                if result {
                    newGrandChildren, resumeAt = result
                    newChild = child.copyWithChildren(newGrandChildren)
                    newChildren = list(children[:index]) + [newChild]
                    // Index := range the original parent
                    resumeAt = (newChild.index, resumeAt)
                    index += 1  // Remove placeholders after child
                    break
                }
            }
    } else {
        return None
    }

    removePlaceholders(children[index:], absoluteBoxes, fixedBoxes)
    return newChildren, resumeAt


// Like reversed(list(enumerate(seq))) without copying the whole seq.
func reversedEnumerate(seq) {
    return zip(reversed(range(len(seq))), reversed(seq))
} 

// For boxes that have been removed := range findEarlierPageBreak(),
//     also remove the matching placeholders := range absoluteBoxes && fixedBoxes.
//     
func removePlaceholders(boxList, absoluteBoxes, fixedBoxes) {
    for box := range boxList {
        if isinstance(box, boxes.ParentBox) {
            removePlaceholders(box.children, absoluteBoxes, fixedBoxes)
        } if box.Style["position"] == "absolute" && box := range absoluteBoxes {
            // box is ! := range absoluteBoxes if its parent has position: relative
            absoluteBoxes.remove(box)
        } else if box.Style["position"] == "fixed" {
            fixedBoxes.remove(box)
