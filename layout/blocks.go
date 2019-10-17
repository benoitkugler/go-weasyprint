package layout

import (
	bo "github.com/benoitkugler/go-weasyprint/boxes"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
)

// Page breaking and layout for block-level and block-container boxes.

type page struct {
	break_ string
	page   pr.Page
}

type blockLayout struct {
	newBox            bo.Box
	resumeAt          *bo.SkipStack
	nextPage          page
	adjoiningMargins  []pr.Float
	collapsingThrough bool
}

// // Lay out the block-level ``box``.
// //
// // `maxPositionY` is the absolute vertical position (as in
// // ``someBox.PositionY``) of the bottom of the
// // content box of the current page area.
// func blockLevelLayout(context LayoutContext, box_ bo.InstanceBlockLevelBox, maxPositionY pr.Float, skipStack *bo.SkipStack,
//                        containingBlock Box, pageIsEmpty bool, absoluteBoxes []*AbsolutePlaceholder,
//                        fixedBoxes []Box, adjoiningMargins []pr.Float ) blockLayout{

// 						box := box_.Box()
//     if ! bo.TypeTableBox.IsInstance(box_) {
//         resolvePercentages2(box_, containingBlock, "")

//         if box.MarginTop.Auto() {
//             box.MarginTop = pr.Float(0)
// 		}
// 		if box.MarginBottom.Auto() {
//             box.MarginBottom = pr.Float(0)
//         }

//         if (context.currentPage > 1 && pageIsEmpty) {
//             // TODO: we should take care of cases when this box doesn't have
//             // collapsing margins with the first child of the page, see
//             // testMarginBreakClearance.
//             if box.Style.GetMarginBreak() == "discard" {
//                 box.MarginTop = pr.Float(0)
//             } else if box.Style.GetMarginBreak() == "auto" {
//                 if ! context.forcedBreak {
//                     box.MarginTop = pr.Float(0)
//                 }
//             }
//         }

// 		collapsedMargin := collapseMargin(append(adjoiningMargins ,box.MarginTop))
// 		bl := box_.BlockLevel()
//         bl.Clearance = getClearance(context, box_, collapsedMargin)
//         if bl.Clearance  != nil  {
//             topBorderEdge = box.PositionY + collapsedMargin + bl.Clearance.V()
//             box.PositionY = topBorderEdge - box.MarginTop
//             adjoiningMargins = nil
//         }
// 	}
//     return blockLevelLayoutSwitch( context, box, maxPositionY, skipStack, containingBlock,
//         pageIsEmpty, absoluteBoxes, fixedBoxes, adjoiningMargins)
// 	}

// // Call the layout function corresponding to the ``box`` type.
// func blockLevelLayoutSwitch(context LayoutContext, box bo.InstanceBlockLevelBox, maxPositionY pr.Float, skipStack *bo.SkipStack,
//                               containingBlock Box, pageIsEmpty bool, absoluteBoxes,
//                               fixedBoxes, adjoiningMargins []Box) blockLayout {

//     if bo.TypeTableBox.IsInstance(box) {
//         return tableLayout(context, box, maxPositionY, skipStack, containingBlock,
//             pageIsEmpty, absoluteBoxes, fixedBoxes)
//     } else if bo.TypeBlockBox.IsInstance(box) {
//         return blockBoxLayout(context, box, maxPositionY, skipStack, containingBlock,
//             pageIsEmpty, absoluteBoxes, fixedBoxes, adjoiningMargins)
//     } else if isinstance(box, boxes.BlockReplacedBox) {
//         box = blockReplacedBoxLayout(box, containingBlock)
//         // Don't collide with floats
//         // http://www.w3.org/TR/CSS21/visuren.html#floats
//         box.PositionX, box.PositionY, _ = avoidCollisions(context, box, containingBlock, false)
//         nextPage := page{break_: "any", page: nil}
//         return blockLayout{
// 			box:box, resumeAt:nil, nextPage:nextPage, adjoiningMargins:nil, collapsingThrough:false,
// 		}
//     } else if bo.TypeFlexBox.IsInstance(box) {
//         return flexLayout(context, box, maxPositionY, skipStack, containingBlock,
//             pageIsEmpty, absoluteBoxes, fixedBoxes)
//     } else {  // pragma: no cover
// 		log.Fatalf("Layout for %s not handled yet", box)
// 		return blockLayout{}
// 		}
// 	}

// // Lay out the block ``box``.
// func blockBoxLayout(context LayoutContext, box_ bo.InstanceBlockBox, maxPositionY pr.Float, skipStack *bo.SkipStack,
//                      containingBlock Box, pageIsEmpty bool, absoluteBoxes []*AbsolutePlaceholder,fixedBoxes, adjoiningMargins []Box) blockLayout {
//                      box := box_.Box()
//     if box.Style.GetColumnWidth().String != "auto" || box.Style.GetColumnCount().String != "auto" {
//         result := columnsLayout(context, box, maxPositionY, skipStack, containingBlock,
//             pageIsEmpty, absoluteBoxes, fixedBoxes, adjoiningMargins)

//         resumeAt = result[1]
//         // TODO: this condition && the whole relayout are probably wrong
//         if resumeAt  == nil  {
//             newBox = result[0]
//             bottomSpacing = (
//                 newBox.MarginBottom + newBox.PaddingBottom +
//                 newBox.BorderBottomWidth)
//             if bottomSpacing {
//                 maxPositionY -= bottomSpacing
//                 result = columnsLayout(context, box, maxPositionY, skipStack,
//                     containingBlock, pageIsEmpty, absoluteBoxes,
//                     fixedBoxes, adjoiningMargins)
//             }
//         }

//         return result
//     else if box.IsTableWrapper {
//         tableWrapperWidth(
//             context, box, (containingBlock.width, containingBlock.height))
//     } blockLevelWidth(box, containingBlock)

//     newBox, resumeAt, nextPage, adjoiningMargins, collapsingThrough = \
//         blockContainerLayout(
//             context, box, maxPositionY, skipStack, pageIsEmpty,
//             absoluteBoxes, fixedBoxes, adjoiningMargins)
//     if newBox && newBox.IsTableWrapper {
//         // Don"t collide with floats
//         // http://www.w3.org/TR/CSS21/visuren.html#floats
//         positionX, positionY, _ = avoidCollisions(
//             context, newBox, containingBlock, outer=false)
//         newBox.translate(
//             positionX - newBox.PositionX, positionY - newBox.PositionY)
//     } return newBox, resumeAt, nextPage, adjoiningMargins, collapsingThrough

// @handleMinMaxWidth
// func blockReplacedWidth(box, containingBlock) {
//     // http://www.w3.org/TR/CSS21/visudet.html#block-replaced-width
//     replacedBoxWidth.withoutMinMax(box, containingBlock)
//     blockLevelWidth.withoutMinMax(box, containingBlock)
// }

// // Lay out the block :class:`boxes.ReplacedBox` ``box``.
// func blockReplacedBoxLayout(box bo.InstanceBlockReplaced, containingBlock) {
//     box = box.Copy()
//     if box.Style["width"] == "auto" && box.Style["height"] == "auto" {
//         computedMargins = box.MarginLeft, box.MarginRight
//         blockReplacedWidth.withoutMinMax(
//             box, containingBlock)
//         replacedBoxHeight.withoutMinMax(box)
//         minMaxAutoReplaced(box)
//         box.MarginLeft, box.MarginRight = computedMargins
//         blockLevelWidth.withoutMinMax(box, containingBlock)
//     } else {
//         blockReplacedWidth(box, containingBlock)
//         replacedBoxHeight(box)
//     }
// }
//     return box

// @handleMinMaxWidth
// // Set the ``box`` width.
// func blockLevelWidth(box, containingBlock) {
//     // "cb" stands for "containing block"
//     cbWidth = containingBlock.width
// }
//     // http://www.w3.org/TR/CSS21/visudet.html#blockwidth

//     // These names are waaay too long
//     marginL = box.MarginLeft
//     marginR = box.MarginRight
//     paddingL = box.PaddingLeft
//     paddingR = box.PaddingRight
//     borderL = box.BorderLeftWidth
//     borderR = box.BorderRightWidth
//     width = box.width

//     // Only margin-left, margin-right && width can be "auto".
//     // We want:  width of containing block ==
//     //               margin-left + border-left-width + padding-left + width
//     //               + padding-right + border-right-width + margin-right

//     paddingsPlusBorders = paddingL + paddingR + borderL + borderR
//     if box.width != "auto" {
//         total = paddingsPlusBorders + width
//         if marginL != "auto" {
//             total += marginL
//         } if marginR != "auto" {
//             total += marginR
//         } if total > cbWidth {
//             if marginL == "auto" {
//                 marginL = box.MarginLeft = 0
//             } if marginR == "auto" {
//                 marginR = box.MarginRight = 0
//             }
//         }
//     } if width != "auto" && marginL != "auto" && marginR != "auto" {
//         // The equation is over-constrained.
//         if containingBlock.style["direction"] == "rtl" && ! box.IsColumn {
//             box.PositionX += (
//                 cbWidth - paddingsPlusBorders - width - marginR - marginL)
//         } // Do nothing := range ltr.
//     } if width == "auto" {
//         if marginL == "auto" {
//             marginL = box.MarginLeft = 0
//         } if marginR == "auto" {
//             marginR = box.MarginRight = 0
//         } width = box.width = cbWidth - (
//             paddingsPlusBorders + marginL + marginR)
//     } marginSum = cbWidth - paddingsPlusBorders - width
//     if marginL == "auto" && marginR == "auto" {
//         box.MarginLeft = marginSum / 2.
//         box.MarginRight = marginSum / 2.
//     } else if marginL == "auto" && marginR != "auto" {
//         box.MarginLeft = marginSum - marginR
//     } else if marginL != "auto" && marginR == "auto" {
//         box.MarginRight = marginSum - marginL
//     }

// // Translate the ``box`` if it is relatively positioned.
// func relativePositioning(box, containingBlock) {
//     if box.Style.GetPosition().String == "relative" {
//         resolvePositionPercentages(box, containingBlock)
//    nil
// }
//         if box.left != "auto" && box.right != "auto" {
//             if box.Style["direction"] == "ltr" {
//                 translateX = box.left
//             } else {
//                 translateX = -box.right
//             }
//         } else if box.left != "auto" {
//             translateX = box.left
//         } else if box.right != "auto" {
//             translateX = -box.right
//         } else {
//             translateX = 0
//         }

//         if box.top != "auto" {
//             translateY = box.top
//         } else if box.Style["bottom"] != "auto" {
//             translateY = -box.Bottom
//         } else {
//             translateY = 0
//         }

//         box.translate(translateX, translateY)

//     if isinstance(box, (boxes.InlineBox, boxes.LineBox)) {
//         for child := range box.Children {
//             relativePositioning(child, containingBlock)
//         }
//     }

func reversed(f []*AbsolutePlaceholder) []*AbsolutePlaceholder{
    L:=  len(f)
    out := make([]*AbsolutePlaceholder, L)
	for i,v := range f {
        out[L-1-i] = v
    }
    return out
}

// Set the ``box`` height.
func blockContainerLayout(context *LayoutContext, box_ bo.Box, maxPositionY pr.Float, skipStack *bo.SkipStack,
	pageIsEmpty bool, absoluteBoxes , fixedBoxes *[]*AbsolutePlaceholder, adjoiningMargins []pr.Float) blockLayout {
        box := box_.Box()
    // TODO: boxes.FlexBox is allowed here because flexLayout calls
    // blockContainerLayout, there's probably a better solution.
    if !(bo.IsBlockContainerBox(box_) || bo.TypeFlexBox.IsInstance(box_)) {
        log.Fatalf("expected BlockContainer or Flex, got %s", box_)
    }

    // We have to work around floating point rounding errors here.
    // The 1e-9 value comes from PEP 485.
    allowedMaxPositionY := maxPositionY * (1 + 1e-9)

    // See http://www.w3.org/TR/CSS21/visuren.html#block-formatting
    if ! bo.TypeBlockBox.IsInstance(box_) {
        context.createBlockFormattingContext()
    }

    isStart := skipStack  == nil
    if box.Style.GetBoxDecorationBreak() == "slice" && ! isStart {
        // Remove top margin, border && padding : 
        box_.RemoveDecoration(box, true, false)
    }

    if box.Style.GetBoxDecorationBreak() == "clone" {
        maxPositionY -= box.PaddingBottom.V() + box.BorderBottomWidth +box.MarginBottom.V()
    }

    adjoiningMargins = append(adjoiningMargins,box.MarginTop.V())
    thisBoxAdjoiningMargins := adjoiningMargins

    collapsingWithChildren := !(pr.Is(box.BorderTopWidth) || pr.Is(box.PaddingTop) || box.IsFlexItem ||
        establishesFormattingContext(box) || box.IsForRootElement)
    var positionY pr.Float
    if collapsingWithChildren {
        // XXX not counting margins in adjoiningMargins, if any
        // (There are not padding or borders, see above.)
        positionY = box.PositionY
    } else {
        box.PositionY += collapseMargin(adjoiningMargins) - box.MarginTop.V()
        adjoiningMargins = nil
        positionY = box.ContentBoxY()
    }

    positionX := box.ContentBoxX()

    if box.Style.GetPosition().String == "relative" {
        // New containing block, use a new absolute list
        absoluteBoxes = nil
    }

    var newChildren []*AbsolutePlaceholder
    nextPage := page{break_: "any", page: nil}

    var lastInFlowChild *AbsolutePlaceholder

    skip := 0
    firstLetterStyle = box.FirstLetterStyle
            if !isStart {
        skip, skipStack = skipStack.Skip, skipStack.Stack
        firstLetterStyle = nil
    } 
    for i, child_ := range box.Children[skip:] {
        index := i + skip
        child := child_.Box()
        child.PositionX = positionX
        // XXX does not count margins in adjoiningMargins :
        child.PositionY = positionY
    
        if ! child.IsInNormalFlow() {
            child.PositionY += collapseMargin(adjoiningMargins)
            if child.IsAbsolutelyPositioned() {
                placeholder := NewAbsolutePlaceholder(child_)
                placeholder.index = index
                newChildren = append(newChildren,placeholder)
                if child.Style.GetPosition().String == "absolute" {
                    *absoluteBoxes = append(*absoluteBoxes,placeholder)
                } else {
                    *fixedBoxes = append(*fixedBoxes,placeholder)
                }
            } else if child.IsFloated() {
                newChild_ := floatLayout(context, child_, box_, absoluteBoxes, fixedBoxes)
                newChild := newChild_.Box.Box()
                // New page if overflow
                if (pageIsEmpty && len(newChildren) == 0 ) || ! (newChild.PositionY + newChild.Height.V() > allowedMaxPositionY) {  
                    newChild_.index = index
                    newChildren = append(newChildren,&newChild)
                } else {
                    for _, previousChild := range reversed(newChildren) {
                        if previousChild.IsInNormalFlow() {
                            lastInFlowChild = previousChild
                            break
                        }
                    } 
                    pageBreak := blockLevelPageBreak(lastInFlowChild, child)
                    if len(newChildren) != 0 && (pageBreak == "avoid" || pageBreak == "avoid-page") {
                        r1, r2 := findEarlierPageBreak(newChildren, absoluteBoxes, fixedBoxes)
                        if r1 != nil || r2 != nil  {
                            newChildren, resumeAt = r1,r2
                            break
                        }
                    } 
                    resumeAt = &bo.SkipStack{inde: index}
                    break
                }
            } else if child.IsRunning() {
                run := child.Style.GetPosition().String
                default_ := map[int]Box{}
                currentRE, has := context.runningElements[run]
                if !has {
                    currentRE = default_
                    context.runningElements[run] = default_
                }
                currentRE[context.currentPage - 1] = child_
            } 
            continue
        }

        if bo.TypeLineBox.IsInstance(child_) {
            if len(box.Children) != 1 {
                log.Fatalf("line box with siblings before layout")
            } 
            if len(adjoiningMargins) != 0 {
                positionY += collapseMargin(adjoiningMargins)
                adjoiningMargins = nil
            } 
            newContainingBlock := box
            linesIterator := iterLineBoxes(context, child, positionY, skipStack,
                newContainingBlock, absoluteBoxes, fixedBoxes, firstLetterStyle)
            isPageBreak := false
            for linesIterator.Has() {
                tmp := linesIterator.Next()
                line_, resumeAt := tmp.line, tmp.resumeAt
                line_.resumeAt = resumeAt
                line := line_.Box.Box()
                newPositionY := line.PositionY + line.Height.V()
       
                // Add bottom padding and border to the bottom position of the box if needed
                var offsetY pr.Float
                if resumeAt  == nil  ||  box.Style.GetBoxDecorationBreak() == "clone" {
                    offsetY = box.BorderBottomWidth.V() + box.PaddingBottom.V()
                }
                

                // Allow overflow if the first line of the page is higher
                // than the page itself so that we put *something* on this
                // page and can advance in the context.
                if newPositionY + offsetY > allowedMaxPositionY && (len(newChildren) != 0 || ! pageIsEmpty) {
                    overOrphans = len(newChildren) - int(box.Style.GetOrphans())
                    if overOrphans < 0 && ! pageIsEmpty {
                        // Reached the bottom of the page before we had
                        // enough lines for orphans, cancel the whole box.
                        page , _ := child_.PageValues()
                        return blockLayout{nextPage: page{break_: "any", page: page}}
                    } 
                    // How many lines we need on the next page to satisfy widows
                    // -1 for the current line.
                    needed := box.Style.GetWidows() - 1
                    if needed != 0 {
                        for linesIterator.Has() {
                            needed -= 1
                            if needed == 0 {
                                break
                            }
                        }
                    } 
                    if needed > overOrphans && ! pageIsEmpty {
                        // Total number of lines < orphans + widows
                        page, _ := child.PageValues()
                        return blockLayout{nextPage: page{break_: "any", page: page}}
                    } 
                    if needed != 0 && needed <= overOrphans {
                        // Remove lines to keep them for the next page
                        newChildren = newChildren[:needed-1]
                    } 
                    // Page break here, resume before this line
                    resumeAt = &bo.SkipStack{Skip:index,Stack:  skipStack}
                    isPageBreak = true
                    break
                

                // TODO: this is incomplete.
                // See http://dev.w3.org/csswg/css3-page/#allowed-pg-brk
                // "When an unforced page break occurs here, both the adjoining
                //  ‘margin-top’ and ‘margin-bottom’ are set to zero."
                // See https://github.com/Kozea/WeasyPrint/issues/115
                } else if pageIsEmpty && newPositionY > allowedMaxPositionY {
                    // Remove the top border when a page is empty && the box is
                    // too high to be drawn := range one page
                    newPositionY -= box.MarginTop.V()
                    line_.Box.Translate(line_.Box,  0, -box.MarginTop.V(), false)
                    box.MarginTop = pr.Float(0)
                } 
                newChildren = append(newChildren, line)
                positionY = newPositionY
                skipStack = resumeAt
            }

            if len(newChildren) != 0 {
                resumeAt = bo.SkipStack{Skip:index,Stack:newChildren[len(newChildren)-1].resumeAt}
            } 
            if isPageBreak {
                break
            }
        } else {
            for previousChild := range reversed(newChildren) {
                if previousChild.IsInNormalFlow() {
                    lastInFlowChild = previousChild
                    break
                }
            } else {
                lastInFlowChild = None
            } 
            
            
            if lastInFlowChild  != nil  {
                // Between in-flow siblings
                pageBreak = blockLevelPageBreak(lastInFlowChild, child)
                pageName = blockLevelPageName(lastInFlowChild, child)
                if pageName || pageBreak := range (
                        "page", "left", "right", "recto", "verso") {
                        }
                    pageName = child.PageValues()[0]
                    nextPage = {"break": pageBreak, "page": pageName}
                    resumeAt = (index, None)
                    break
            } else {
                pageBreak = "auto"
            }
        
            newContainingBlock = box

            if ! newContainingBlock.isTableWrapper {
                resolvePercentages(child, newContainingBlock)
                if (child.IsInNormalFlow() and
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
                        positionY = box.ContentBoxY()

            if adjoiningMargins && box.IsTableWrapper {
                collapsedMargin = collapseMargin(adjoiningMargins)
                child.PositionY += collapsedMargin
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
                newChild.Index = index
            }

                // We need to do this after the child layout to have the
                // used value for marginTop (eg. it might be a percentage.)
                if ! isinstance(
                        newChild, (boxes.BlockBox, boxes.TableBox)) {
                        }
                    adjoiningMargins = append(adjoiningMargins, newChild.marginTop)
                    offsetY = (
                        collapseMargin(adjoiningMargins) -
                        newChild.marginTop)
                    newChild.translate(0, offsetY)
                    adjoiningMargins = []
                // else: blocks handle that themselves.

                adjoiningMargins = nextAdjoiningMargins
                adjoiningMargins = append(adjoiningMargins, newChild.marginBottom)

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
                            page = child.PageValues()[0]
                            return (
                                None, None, {"break": "any", "page": page}, [],
                                false)
                        } // else {
                        } // ignore this "avoid" && break anyway.
                    }
                }
            }

                if all(child.IsAbsolutelyPositioned()
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
                    page = child.PageValues()[0]
                    return (
                        None, None, {"break": "any", "page": page}, [], false)
                }

            // Bottom borders may overflow here
            // TODO: back-track somehow when all lines fit but ! borders
            newChildren = append(newChildren, newChild)
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
        if previousChild.IsInNormalFlow() {
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
                box.MinHeight, box.BorderTopWidth, box.PaddingTop,
                box.BorderBottomWidth, box.PaddingBottom])) {
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

    if (box.BorderBottomWidth or
            box.PaddingBottom or
            establishesFormattingContext(box) or
            box.IsForRootElement or
            box.IsTableWrapper) {
            }
        positionY += collapseMargin(adjoiningMargins)
        adjoiningMargins = []

    newBox = box.CopyWithChildren(
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

    if newBox.Style.GetPosition().String == "relative" {
        // New containing block, resolve the layout of the absolute descendants
        for absoluteBox := range absolnileBoxes {
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
        if box.Style.GetBoxDecorationBreak() == "clone" {
            newBox.height += (
                box.PaddingBottom + box.BorderBottomWidth +
                box.MarginBottom)
        }
    }

    if nextPage["page"]  == nil  {
        nextPage["page"] = newBox.PageValues()[1]
    }

    return newBox, resumeAt, nextPage, adjoiningMargins, collapsingThrough
    }
}

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
// See http://www.w3.org/TR/CSS2/visuren.html#block-formatting
func establishesFormattingContext(box_ Box) bool {
	box := box_.Box()
	return box.IsFloated() ||
		box.IsAbsolutelyPositioned() ||
		// TODO: columns shouldn't be block boxes, this condition would then be
		// useless when this is fixed
		box.IsColumn ||
		(bo.IsBlockContainerBox(box_) && !bo.TypeBlockBox.IsInstance(box_)) ||
		(bo.TypeBlockBox.IsInstance(box_) && box.Style.GetOverflow() != "visible")
}

// https://drafts.csswg.org/css-break-3/#possible-breaks
func isParallel(box Box) bool {
	return bo.IsBlockLevelBox(box) || bo.TypeTableRowGroupBox.IsInstance(box) || bo.TypeTableRowBox.IsInstance(box)
}

func reverseStrings(f []pr.String) {
	for left, right := 0, len(f)-1; left < right; left, right = left+1, right-1 {
		f[left], f[right] = f[right], f[left]
	}
}

// Return the value of ``page-break-before`` or ``page-break-after``
// that "wins" for boxes that meet at the margin between two sibling boxes.
// For boxes before the margin, the "page-break-after" value is considered;
// for boxes after the margin the "page-break-before" value is considered.
// * "avoid" takes priority over "auto"
// * "page" takes priority over "avoid" or "auto"
// * "left" or "right" take priority over "always", "avoid" or "auto"
// * Among "left" && "right", later values in the tree take priority.
// See http://dev.w3.org/csswg/css3-page/#allowed-pg-brk
func blockLevelPageBreak(siblingBefore, siblingAfter Box) pr.String {
	var values []pr.String

	box_ := siblingBefore
	for isParallel(box_) {
		box := box_.Box()
		values = append(values, box.Style.GetBreakAfter())
		if !(bo.IsParentBox(box_) && len(box.Children) != 0) {
			break
		}
		box_ = box.Children[len(box.Children)-1]
	}
	reverseStrings(values) // Have them in tree order

	box_ = siblingAfter
	for isParallel(box_) {
		box := box_.Box()
		values = append(values, box.Style.GetBreakBefore())
		if !(bo.IsParentBox(box_) && len(box.Children) != 0) {
			break
		}
		box_ = box.Children[0]
	}
	choices := map[[2]pr.String]bool{
		{"page", "auto"}:       true,
		{"page", "avoid"}:      true,
		{"avoid", "auto"}:      true,
		{"page", "avoid-page"}: true,
		{"avoid-page", "auto"}: true,
	}
	var result pr.String = "auto"
	for _, value := range values {
		tmp := [2]pr.String{value, result}
		if value == "left" || value == "right" || value == "recto" || value == "verso" || choices[tmp] {
			result = value
		}

	}

	return result
}

// Return the next page name when siblings don't have the same names.
func blockLevelPageName(siblingBefore, siblingAfter Box) *pr.Page {
	_, beforePage := siblingBefore.PageValues()
	afterPage, _ := siblingAfter.PageValues()
	if beforePage != afterPage {
		return &afterPage
	}
	return nil
}

// Because of a `page-break-before: avoid` or a `page-break-after: avoid`
// we need to find an earlier page break opportunity inside `children`.
// Absolute or fixed placeholders removed from children should also be
// removed from `absoluteBoxes` or `fixedBoxes`.
func findEarlierPageBreak(children []Box, absoluteBoxes, fixedBoxes *[]Box) (newChildren []Box, resumeAt *bo.SkipStack) {
	if len(children) != 0 && bo.TypeLineBox.IsInstance(children[0]) {
		// Normally `orphans` && `widows` apply to the block container, but
		// line boxes inherit them.
		orphans := int(children[0].Box().Style.GetOrphans())
		widows := int(children[0].Box().Style.GetWidows())
		index := len(children) - widows // how many lines we keep
		if index < orphans {
			return nil, nil
		}
		newChildren := children[:index]
		resumeAt := &bo.SkipStack{Skip: 0, Stack: newChildren[len(newChildren)-1].Box().ResumeAt}
		removePlaceholders(children[index:], absoluteBoxes, fixedBoxes)
		return newChildren, resumeAt
	}

	var (
		previousInFlow Box
		index          int
		hasBroken      bool
	)
	L := len(children)
	for i_ := range children { // reversed(list(enumerate(children)))
		index = L - i_ - 1
		child_ := children[index]
		child := child_.Box()
		if child.IsInNormalFlow() {
			if pb := blockLevelPageBreak(child_, previousInFlow); previousInFlow != nil && pb != "avoid" && pb != "avoid-page" {
				index += 1 // break after child
				newChildren = children[:index]
				// Get the index in the original parent
				resumeAt = &bo.SkipStack{Skip: children[index].Box().Index}
				hasBroken = true
				break
			}
			previousInFlow = child_
		}
		if bi := child.Style.GetBreakInside(); child.IsInNormalFlow() && bi != "avoid" && bi != "avoid-page" {
			if bo.TypeBlockBox.IsInstance(child_) || bo.TypeTableBox.IsInstance(child_) || bo.TypeTableRowGroupBox.IsInstance(child_) {
				var newGrandChildren []Box
				newGrandChildren, resumeAt = findEarlierPageBreak(child.Children, absoluteBoxes, fixedBoxes)
				if newGrandChildren != nil || resumeAt != nil {
					newChild := bo.CopyWithChildren(child_, newGrandChildren, true, true)
					newChildren = append(children[:index], newChild)
					// Index in the original parent
					resumeAt = &bo.SkipStack{Skip: newChild.Box().Index, Stack: resumeAt}
					index += 1 // Remove placeholders after child
					hasBroken = true
					break
				}
			}
		}
	}
	if !hasBroken {
		return nil, nil
	}

	removePlaceholders(children[index:], absoluteBoxes, fixedBoxes)
	return newChildren, resumeAt
}

func removeBox(list *[]Box, box Box) {
	out := make([]Box, 0, len(*list))
	for _, v := range *list {
		if v != box {
			out = append(out, v)
		}
	}
	*list = out
}

// For boxes that have been removed in findEarlierPageBreak(),
// also remove the matching placeholders in absoluteBoxes and fixedBoxes.
func removePlaceholders(boxList []Box, absoluteBoxes, fixedBoxes *[]Box) {
	for _, box_ := range boxList {
		box := box_.Box()
		if bo.IsParentBox(box_) {
			removePlaceholders(box.Children, absoluteBoxes, fixedBoxes)
		}
		if box.Style.GetPosition().String == "absolute" {
			// box is not in absoluteBoxes if its parent has position: relative
			removeBox(absoluteBoxes, box_)
		} else if box.Style.GetPosition().String == "fixed" {
			removeBox(fixedBoxes, box_)
		}
	}
}
