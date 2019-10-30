package layout

import (
	"log"

	"github.com/benoitkugler/go-weasyprint/style/tree"

	bo "github.com/benoitkugler/go-weasyprint/boxes"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
)

// Page breaking and layout for block-level and block-container boxes.

type blockLayout struct {
	// newBox            bo.Box
	resumeAt          *tree.SkipStack
	nextPage          tree.PageBreak
	adjoiningMargins  []pr.Float
	collapsingThrough bool
}

// Lay out the block-level ``box``.
//
// `maxPositionY` is the absolute vertical position (as in
// ``someBox.PositionY``) of the bottom of the
// content box of the current page area.
func blockLevelLayout(context *LayoutContext, box_ bo.InstanceBlockLevelBox, maxPositionY pr.Float, skipStack *tree.SkipStack,
	containingBlock bo.BoxFields, pageIsEmpty bool, absoluteBoxes,
	fixedBoxes *[]*AbsolutePlaceholder, adjoiningMargins []pr.Float) (bo.InstanceBlockLevelBox, blockLayout) {

	box := box_.Box()
	if !bo.TypeTableBox.IsInstance(box_) {
		resolvePercentages2(box_, containingBlock, "")

		if box.MarginTop == pr.Auto {
			box.MarginTop = pr.Float(0)
		}
		if box.MarginBottom == pr.Auto {
			box.MarginBottom = pr.Float(0)
		}

		if context.currentPage > 1 && pageIsEmpty {
			// TODO: we should take care of cases when this box doesn't have
			// collapsing margins with the first child of the page, see
			// testMarginBreakClearance.
			if box.Style.GetMarginBreak() == "discard" {
				box.MarginTop = pr.Float(0)
			} else if box.Style.GetMarginBreak() == "auto" {
				if !context.forcedBreak {
					box.MarginTop = pr.Float(0)
				}
			}
		}

		collapsedMargin := collapseMargin(append(adjoiningMargins, box.MarginTop.V()))
		bl := box_.BlockLevel()
		bl.Clearance = getClearance(*context, *box, collapsedMargin)
		if bl.Clearance != nil {
			topBorderEdge := box.PositionY + collapsedMargin + bl.Clearance.V()
			box.PositionY = topBorderEdge - box.MarginTop.V()
			adjoiningMargins = nil
		}
	}
	return blockLevelLayoutSwitch(context, box_, maxPositionY, skipStack, containingBlock,
		pageIsEmpty, absoluteBoxes, fixedBoxes, adjoiningMargins)
}

// Call the layout function corresponding to the ``box`` type.
func blockLevelLayoutSwitch(context *LayoutContext, box_ bo.InstanceBlockLevelBox, maxPositionY pr.Float, skipStack *tree.SkipStack,
	containingBlock bo.BoxFields, pageIsEmpty bool, absoluteBoxes,
	fixedBoxes *[]*AbsolutePlaceholder, adjoiningMargins []pr.Float) (bo.InstanceBlockLevelBox, blockLayout) {

	blockBox, isBlockBox := box_.(bo.InstanceBlockBox)
	replacedBox, isReplacedBox := box_.(bo.InstanceReplacedBox)
	if table, ok := box_.(bo.InstanceTableBox); ok {
		return tableLayout(context, table, maxPositionY, skipStack, pageIsEmpty, absoluteBoxes, fixedBoxes)
	} else if isBlockBox {
		return blockBoxLayout(context, blockBox, maxPositionY, skipStack, containingBlock,
			pageIsEmpty, absoluteBoxes, fixedBoxes, adjoiningMargins)
	} else if isReplacedBox && bo.IsBlockReplacedBox(box_) {
		box_ = blockReplacedBoxLayout(replacedBox, containingBlock).(bo.InstanceBlockLevelBox) // blockReplacedBoxLayout is type stable
		box := replacedBox.Box()
		// Don't collide with floats
		// http://www.w3.org/TR/CSS21/visuren.html#floats
		box.PositionX, box.PositionY, _ = avoidCollisions(*context, replacedBox, containingBlock, false)
		nextPage := tree.PageBreak{Break: "any"}
		return box_, blockLayout{resumeAt: nil, nextPage: nextPage, adjoiningMargins: nil, collapsingThrough: false}
	} else if bo.TypeFlexBox.IsInstance(box_) {
		return flexLayout(context, box_, maxPositionY, skipStack, containingBlock,
			pageIsEmpty, absoluteBoxes, fixedBoxes)
	} else { // pragma: no cover
		log.Fatalf("Layout for %s not handled yet", box_)
		return nil, blockLayout{}
	}
}

// Lay out the block ``box``.
func blockBoxLayout(context *LayoutContext, box_ bo.InstanceBlockBox, maxPositionY pr.Float, skipStack *tree.SkipStack,
	containingBlock bo.BoxFields, pageIsEmpty bool, absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder, adjoiningMargins []pr.Float) (bo.InstanceBlockLevelBox, blockLayout) {
	box := box_.Box()
	if box.Style.GetColumnWidth().String != "auto" || box.Style.GetColumnCount().String != "auto" {
		newBox_, result := columnsLayout(context, box_, maxPositionY, skipStack, containingBlock,
			pageIsEmpty, absoluteBoxes, fixedBoxes, adjoiningMargins)
		newBox := newBox_.Box()
		resumeAt := result.resumeAt
		// TODO: this condition and the whole relayout are probably wrong
		if resumeAt == nil {
			bottomSpacing := newBox.MarginBottom.V() + newBox.PaddingBottom.V() + newBox.BorderBottomWidth.V()
			if bottomSpacing != 0 {
				maxPositionY -= bottomSpacing
				newBox_, result = columnsLayout(context, box_, maxPositionY, skipStack,
					containingBlock, pageIsEmpty, absoluteBoxes, fixedBoxes, adjoiningMargins)
			}
		}
		return newBox_, result
	} else if box.IsTableWrapper {
		tableWrapperWidth(context, box, bo.MaybePoint{containingBlock.Width, containingBlock.Height})
	}
	blockLevelWidth(box_, nil, containingBlock)

	newBox__, result := blockContainerLayout(context, box_, maxPositionY, skipStack, pageIsEmpty,
		absoluteBoxes, fixedBoxes, adjoiningMargins)
	newBox := newBox__.(bo.InstanceBlockBox) // blockContainerLayout is type stable
	if newBox != nil && newBox.Box().IsTableWrapper {
		// Don't collide with floats
		// http://www.w3.org/TR/CSS21/visuren.html#floats
		positionX, positionY, _ := avoidCollisions(*context, newBox, containingBlock, false)
		newBox.Translate(newBox, positionX-newBox.Box().PositionX, positionY-newBox.Box().PositionY, false)
	}
	return newBox, result
}

var blockReplacedWidth = handleMinMaxWidth(blockReplacedWidth_)

// @handleMinMaxWidth
func blockReplacedWidth_(box Box, _ *LayoutContext, containingBlock containingBlock) (bool, pr.Float) {
	// http://www.w3.org/TR/CSS21/visudet.html#block-replaced-width
	replacedBoxWidth_(box, nil, containingBlock)
	blockLevelWidth_(box, nil, containingBlock)
	return false, 0
}

// Lay out the block :class:`boxes.ReplacedBox` ``box``.
func blockReplacedBoxLayout(box_ bo.InstanceReplacedBox, containingBlock bo.BoxFields) bo.InstanceReplacedBox {
	box_ = box_.Copy().(bo.InstanceReplacedBox) // Copy is type stable
	box := box_.Box()
	if box.Style.GetWidth().String == "auto" && box.Style.GetHeight().String == "auto" {
		computedMarginsL, computedMarginsR := box.MarginLeft, box.MarginRight
		blockReplacedWidth_(box_, nil, containingBlock)
		replacedBoxHeight_(box_, nil, nil)
		minMaxAutoReplaced(box)
		box.MarginLeft, box.MarginRight = computedMarginsL, computedMarginsR
		blockLevelWidth_(box_, nil, containingBlock)
	} else {
		blockReplacedWidth(box_, nil, containingBlock)
		replacedBoxHeight(box_, nil, nil)
	}

	return box_
}

var blockLevelWidth = handleMinMaxWidth(blockLevelWidth_)

// @handleMinMaxWidth
// Set the ``box`` width.
// containingBlock must be bo.BoxFields
func blockLevelWidth_(box_ Box, _ *LayoutContext, containingBlock_ containingBlock) (bool, pr.Float) {
	box := box_.Box()
	// "cb" stands for "containing block"
	containingBlock := containingBlock_.(bo.BoxFields)
	cbWidth := containingBlock.Width.V()

	// http://www.w3.org/TR/CSS21/visudet.html#blockwidth

	// These names are waaay too long
	marginL := box.MarginLeft
	marginR := box.MarginRight
	width := box.Width
	paddingL := box.PaddingLeft.V()
	paddingR := box.PaddingRight.V()
	borderL := box.BorderLeftWidth.V()
	borderR := box.BorderRightWidth.V()

	// Only margin-left, margin-right and width can be "auto".
	// We want:  width of containing block ==
	//               margin-left + border-left-width + padding-left + width
	//               + padding-right + border-right-width + margin-right

	paddingsPlusBorders := paddingL + paddingR + borderL + borderR
	if width != pr.Auto {
		total := paddingsPlusBorders + width.V()
		if marginL != pr.Auto {
			total += marginL.V()
		}
		if marginR != pr.Auto {
			total += marginR.V()
		}
		if total > cbWidth {
			if marginL == pr.Auto {
				marginL = pr.Float(0)
				box.MarginLeft = pr.Float(0)
			}
			if marginR == pr.Auto {
				marginR = pr.Float(0)
				box.MarginRight = pr.Float(0)
			}
		}
	}
	if width != pr.Auto && marginL != pr.Auto && marginR != pr.Auto {
		// The equation is over-constrained.
		if containingBlock.Style.GetDirection() == "rtl" && !box.IsColumn {
			box.PositionX += (cbWidth - paddingsPlusBorders - width.V() - marginR.V() - marginL.V())
		} // Do nothing := range ltr.
	}
	if width == pr.Auto {
		if marginL == pr.Auto {
			marginL = pr.Float(0)
			box.MarginLeft = pr.Float(0)
		}
		if marginR == pr.Auto {
			marginR = pr.Float(0)
			box.MarginRight = pr.Float(0)
		}
		width = cbWidth - (paddingsPlusBorders + marginL.V() + marginR.V())
		box.Width = width
	}
	marginSum := cbWidth - paddingsPlusBorders - width.V()
	if marginL == pr.Auto && marginR == pr.Auto {
		box.MarginLeft = marginSum / 2.
		box.MarginRight = marginSum / 2.
	} else if marginL == pr.Auto && marginR != pr.Auto {
		box.MarginLeft = marginSum - marginR.V()
	} else if marginL != pr.Auto && marginR == pr.Auto {
		box.MarginRight = marginSum - marginL.V()
	}
	return false, 0
}

// Translate the ``box`` if it is relatively positioned.
func relativePositioning(box_ Box, containingBlock bo.Point) {
	box := box_.Box()
	if box.Style.GetPosition().String == "relative" {
		resolvePositionPercentages(box, containingBlock)
		var translateX, translateY pr.Float
		if box.Left != pr.Auto && box.Right != pr.Auto {
			if box.Style.GetDirection() == "ltr" {
				translateX = box.Left.V()
			} else {
				translateX = -box.Right.V()
			}
		} else if box.Left != pr.Auto {
			translateX = box.Left.V()
		} else if box.Right != pr.Auto {
			translateX = -box.Right.V()
		} else {
			translateX = 0
		}

		if box.Top != pr.Auto {
			translateY = box.Top.V()
		} else if box.Style.GetBottom().String != "auto" {
			translateY = -box.Bottom.V()
		} else {
			translateY = 0
		}

		box_.Translate(box_, translateX, translateY, false)
	}
	if isLine(box_) {
		for _, child := range box.Children {
			relativePositioning(child, containingBlock)
		}
	}
}

func reversed(f []*AbsolutePlaceholder) []*AbsolutePlaceholder {
	L := len(f)
	out := make([]*AbsolutePlaceholder, L)
	for i, v := range f {
		out[L-1-i] = v
	}
	return out
}

func reverseBoxes(in []Box) []Box {
	N := len(in)
	out := make([]Box, N)
	for i, v := range in {
		out[N-1-i] = v
	}
	return out
}

type ChildrenBlockLevel interface {
	Box
	Children() []bo.InstanceBlockLevelBox
}

// Set the ``box`` height.
func blockContainerLayout(context *LayoutContext, box_ Box, maxPositionY pr.Float, skipStack *tree.SkipStack,
	pageIsEmpty bool, absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder, adjoiningMargins []pr.Float) (Box, blockLayout) {
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
	if !bo.TypeBlockBox.IsInstance(box_) {
		context.createBlockFormattingContext()
	}

	isStart := skipStack == nil
	if box.Style.GetBoxDecorationBreak() == "slice" && !isStart {
		// Remove top margin, border && padding :
		box_.RemoveDecoration(box, true, false)
	}

	if box.Style.GetBoxDecorationBreak() == "clone" {
		maxPositionY -= box.PaddingBottom.V() + box.BorderBottomWidth.V() + box.MarginBottom.V()
	}

	adjoiningMargins = append(adjoiningMargins, box.MarginTop.V())
	thisBoxAdjoiningMargins := adjoiningMargins

	collapsingWithChildren := !(pr.Is(box.BorderTopWidth) || pr.Is(box.PaddingTop) || box.IsFlexItem ||
		establishesFormattingContext(box_) || box.IsForRootElement)
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

	var newChildren []Box
	nextPage := tree.PageBreak{Break: "any"}
	var resumeAt *tree.SkipStack
	var lastInFlowChild Box

	skip := 0
	firstLetterStyle := box.FirstLetterStyle
	if !isStart {
		skip, skipStack = skipStack.Skip, skipStack.Stack
		firstLetterStyle = nil
	}
	outerHasBroken := false
	for i, child_ := range box.Children[skip:] {
		index := i + skip
		child := child_.Box()
		child.PositionX = positionX
		// XXX does not count margins in adjoiningMargins :
		child.PositionY = positionY

		if !child.IsInNormalFlow() {
			child.PositionY += collapseMargin(adjoiningMargins)
			if child.IsAbsolutelyPositioned() {
				placeholder := NewAbsolutePlaceholder(child_)
				placeholder.index = index
				newChildren = append(newChildren, placeholder)
				if child.Style.GetPosition().String == "absolute" {
					*absoluteBoxes = append(*absoluteBoxes, placeholder)
				} else {
					*fixedBoxes = append(*fixedBoxes, placeholder)
				}
			} else if child.IsFloated() {
				newChild_ := floatLayout(context, child_, box_, absoluteBoxes, fixedBoxes)
				newChild := newChild_.Box()
				// New page if overflow
				if (pageIsEmpty && len(newChildren) == 0) || !(newChild.PositionY+newChild.Height.V() > allowedMaxPositionY) {
					asPlaceholder := AbsolutePlaceholder{_Box: newChild_}
					asPlaceholder.index = index
					newChildren = append(newChildren, &asPlaceholder)
				} else {
					for _, previousChild := range reverseBoxes(newChildren) {
						if previousChild.Box().IsInNormalFlow() {
							lastInFlowChild = previousChild
							break
						}
					}
					pageBreak := blockLevelPageBreak(lastInFlowChild, child_)
					if len(newChildren) != 0 && (pageBreak == "avoid" || pageBreak == "avoid-page") {
						r1, r2 := findEarlierPageBreak(newChildren, absoluteBoxes, fixedBoxes)
						if r1 != nil || r2 != nil {
							newChildren, resumeAt = r1, r2
							outerHasBroken = true
							break
						}
					}
					resumeAt = &tree.SkipStack{Skip: index}
					outerHasBroken = true
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
				currentRE[context.currentPage-1] = child_
			}
			continue
		}

		if childLineBox, ok := child_.(*bo.LineBox); ok { //LineBox is a final type
			if len(box.Children) != 1 {
				log.Fatalf("line box with siblings before layout")
			}
			if len(adjoiningMargins) != 0 {
				positionY += collapseMargin(adjoiningMargins)
				adjoiningMargins = nil
			}
			newContainingBlock := box_
			linesIterator := iterLineBoxes(context, childLineBox, positionY, skipStack,
				newContainingBlock, absoluteBoxes, fixedBoxes, firstLetterStyle)
			isPageBreak := false
			for linesIterator.Has() {
				tmp := linesIterator.Next()
				line_, resumeAt := tmp.line, tmp.resumeAt
				line := line_.Box()
				line.ResumeAt = resumeAt
				newPositionY := line.PositionY + line.Height.V()

				// Add bottom padding and border to the bottom position of the box if needed
				var offsetY pr.Float
				if resumeAt == nil || box.Style.GetBoxDecorationBreak() == "clone" {
					offsetY = box.BorderBottomWidth.V() + box.PaddingBottom.V()
				}

				// Allow overflow if the first line of the page is higher
				// than the page itself so that we put *something* on this
				// page and can advance in the context.
				if newPositionY+offsetY > allowedMaxPositionY && (len(newChildren) != 0 || !pageIsEmpty) {
					overOrphans := len(newChildren) - int(box.Style.GetOrphans())
					if overOrphans < 0 && !pageIsEmpty {
						// Reached the bottom of the page before we had
						// enough lines for orphans, cancel the whole box.
						page_, _ := child_.PageValues()
						return nil, blockLayout{nextPage: tree.PageBreak{Break: "any", Page: page_}}
					}
					// How many lines we need on the next page to satisfy widows
					// -1 for the current line.
					needed := int(box.Style.GetWidows() - 1)
					if needed != 0 {
						for linesIterator.Has() {
							needed -= 1
							if needed == 0 {
								break
							}
						}
					}
					if needed > overOrphans && !pageIsEmpty {
						// Total number of lines < orphans + widows
						page_, _ := child.PageValues()
						return nil, blockLayout{nextPage: tree.PageBreak{Break: "any", Page: page_}}
					}
					if needed != 0 && needed <= overOrphans {
						// Remove lines to keep them for the next page
						newChildren = newChildren[:needed-1]
					}
					// Page break here, resume before this line
					resumeAt = &tree.SkipStack{Skip: index, Stack: skipStack}
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
					line_.Translate(line_, 0, -box.MarginTop.V(), false)
					box.MarginTop = pr.Float(0)
				}
				newChildren = append(newChildren, line_)
				positionY = newPositionY
				skipStack = resumeAt
			}

			if len(newChildren) != 0 {
				resumeAt = &tree.SkipStack{Skip: index, Stack: newChildren[len(newChildren)-1].Box().ResumeAt}
			}
			if isPageBreak {
				outerHasBroken = true
				break
			}
		} else {
			hasBroken := false
			for _, previousChild := range reverseBoxes(newChildren) {
				if previousChild.Box().IsInNormalFlow() {
					lastInFlowChild = previousChild
					hasBroken = true
					break
				}
			}
			if !hasBroken {
				lastInFlowChild = nil
			}

			pageBreak := "auto"
			if lastInFlowChild != nil {
				// Between in-flow siblings
				pageBreak = blockLevelPageBreak(lastInFlowChild, child_)
				pageName_ := blockLevelPageName(lastInFlowChild, child_)
				if pageName_ != nil ||
					pageBreak == "page" || pageBreak == "left" || pageBreak == "right" ||
					pageBreak == "recto" || pageBreak == "verso" {
					pageName, _ := child.PageValues()
					nextPage = tree.PageBreak{Break: pageBreak, Page: pageName}
					resumeAt = &tree.SkipStack{Skip: index}
					outerHasBroken = true
					break
				}
			}

			newContainingBlock := box_

			if !newContainingBlock.Box().IsTableWrapper {
				resolvePercentages2(child_, *newContainingBlock.Box(), "")
				if child.IsInNormalFlow() && lastInFlowChild == nil && collapsingWithChildren {
					// TODO: add the adjoining descendants' margin top to
					// [child.MarginTop]
					oldCollapsedMargin := collapseMargin(adjoiningMargins)
					var childMarginTop pr.Float
					if child.MarginTop != pr.Auto {
						childMarginTop = child.MarginTop.V()
					}
					newCollapsedMargin := collapseMargin(append(adjoiningMargins, childMarginTop))
					collapsedMarginDifference := newCollapsedMargin - oldCollapsedMargin
					for _, previousNewChild := range newChildren {
						previousNewChild.Translate(previousNewChild, 0, collapsedMarginDifference, false)
					}

					if clearance := getClearance(*context, *child, newCollapsedMargin); clearance != nil {
						for _, previousNewChild := range newChildren {
							previousNewChild.Translate(previousNewChild, 0, -collapsedMarginDifference, false)
						}

						collapsedMargin := collapseMargin(adjoiningMargins)
						box.PositionY += collapsedMargin - box.MarginTop.V()
						// Count box.MarginTop as we emptied adjoiningMargins
						adjoiningMargins = nil
						positionY = box.ContentBoxY()
					}
				}
			}
			if len(adjoiningMargins) != 0 && box.IsTableWrapper {
				collapsedMargin := collapseMargin(adjoiningMargins)
				child.PositionY += collapsedMargin
				positionY += collapsedMargin
				adjoiningMargins = nil
			}

			notOnlyPlaceholder := false
			for _, child := range newChildren {
				if _, isAbsPlac := child.(*AbsolutePlaceholder); !isAbsPlac {
					notOnlyPlaceholder = true
					break
				}
			}
			pageIsEmptyWithNoChildren := pageIsEmpty && !notOnlyPlaceholder

			if len(child.FirstLetterStyle) == 0 {
				child.FirstLetterStyle = firstLetterStyle
			}
			newChild_, tmp := blockLevelLayout(context, child_.(bo.InstanceBlockLevelBox), maxPositionY, skipStack,
				*newContainingBlock.Box(), pageIsEmptyWithNoChildren, absoluteBoxes, fixedBoxes, adjoiningMargins)
			resumeAt, nextPage = tmp.resumeAt, tmp.nextPage
			nextAdjoiningMargins, collapsingThrough := tmp.adjoiningMargins, tmp.collapsingThrough
			skipStack = nil

			newChildPlace := AbsolutePlaceholder{_Box: newChild_}
			if newChild_ != nil {
				newChild := newChild_.Box()
				// index in its non-laid-out parent, not in future new parent
				// May be used in findEarlierPageBreak()
				newChildPlace.index = index

				// We need to do this after the child layout to have the
				// used value for marginTop (eg. it might be a percentage.)
				if !(bo.TypeBlockBox.IsInstance(newChild_) || bo.TypeTableBox.IsInstance(newChild_)) {
					adjoiningMargins = append(adjoiningMargins, newChild.MarginTop.V())
					offsetY := collapseMargin(adjoiningMargins) - newChild.MarginTop.V()
					newChild_.Translate(newChild_, 0, offsetY, false)
					adjoiningMargins = nil
				}
				// else: blocks handle that themselves.

				adjoiningMargins = nextAdjoiningMargins
				adjoiningMargins = append(adjoiningMargins, newChild.MarginBottom.V())

				if !collapsingThrough {
					newPositionY := newChild.BorderBoxY() + newChild.BorderHeight()

					if newPositionY > allowedMaxPositionY && !pageIsEmptyWithNoChildren {
						// The child overflows the page area, put it on the
						// next page. (But don’t delay whole blocks if eg.
						// only the bottom border overflows.)
						newChild_ = nil
					} else {
						positionY = newPositionY
					}
				}
				if newChild_ != nil && newChild_.BlockLevel().Clearance != nil {
					positionY = newChild.BorderBoxY() + newChild.BorderHeight()
				}
			}
			if newChild_ == nil {
				// Nothing fits in the remaining space of this page: break
				if pageBreak == "avoid" || pageBreak == "avoid-page" {
					// TODO: fill the blank space at the bottom of the page
					r1, r2 := findEarlierPageBreak(newChildren, absoluteBoxes, fixedBoxes)
					if r1 != nil || r2 != nil {
						newChildren, resumeAt = r1, r2
						outerHasBroken = true
						break
					} else {
						// We did not find any page break opportunity
						if !pageIsEmpty {
							// The page has content *before* this block:
							// cancel the block and try to find a break
							// in the parent.
							page_, _ := child.PageValues()
							return nil, blockLayout{nextPage: tree.PageBreak{Break: "any", Page: page_}}
						}
						// else : ignore this "avoid" and break anyway.
					}
				}
				allAbsPos := true
				for _, child := range newChildren {
					if !child.Box().IsAbsolutelyPositioned() {
						allAbsPos = false
						break
					}
				}
				if allAbsPos {
					// This box has only rendered absolute children, keep them
					// for the next page. This is for example useful for list
					// markers.
					removePlaceholders(newChildren, absoluteBoxes, fixedBoxes)
					newChildren = nil
				}
				if len(newChildren) != 0 {
					resumeAt = &tree.SkipStack{Skip: index}
					outerHasBroken = true
					break
				} else {
					// This was the first child of this box, cancel the box
					// completly
					page_, _ := child.PageValues()
					return nil, blockLayout{nextPage: tree.PageBreak{Break: "any", Page: page_}}
				}
			}
			// Bottom borders may overflow here
			// TODO: back-track somehow when all lines fit but not borders
			newChildren = append(newChildren, &newChildPlace)
			if resumeAt != nil {
				resumeAt = &tree.SkipStack{Skip: index, Stack: resumeAt}
				outerHasBroken = true
				break
			}
		}
	}
	if !outerHasBroken {
		resumeAt = nil
	}

	if bi := box.Style.GetBreakInside(); resumeAt != nil && (bi == "avoid" || bi == "avoid-page") && !pageIsEmpty {
		return nil, blockLayout{nextPage: tree.PageBreak{Break: "any"}}
	}

	if collapsingWithChildren {
		box.PositionY += collapseMargin(thisBoxAdjoiningMargins) - box.MarginTop.V()
	}

	lastInFlowChild = nil
	for _, previousChild := range reverseBoxes(newChildren) {
		if previousChild.Box().IsInNormalFlow() {
			lastInFlowChild = previousChild
			break
		}
	}
	collapsingThrough := false
	if lastInFlowChild == nil {
		collapsedMargin := collapseMargin(adjoiningMargins)
		// top && bottom margin of this box
		if (box.Height == pr.Auto || box.Height == pr.Float(0)) &&
			getClearance(*context, *box, collapsedMargin) == nil &&
			box.MinHeight == pr.Float(0) && box.BorderTopWidth == pr.Float(0) && box.PaddingTop == pr.Float(0) &&
			box.BorderBottomWidth == pr.Float(0) && box.PaddingBottom == pr.Float(0) {
			collapsingThrough = true
		} else {
			positionY += collapsedMargin
			adjoiningMargins = nil
		}
	} else {
		// bottom margin of the last child && bottom margin of this box ...
		if box.Height != pr.Auto {
			// not adjoining. (positionY is not used afterwards.)
			adjoiningMargins = nil
		}
	}

	if pr.Is(box.BorderBottomWidth) || pr.Is(box.PaddingBottom) ||
		establishesFormattingContext(box_) || box.IsForRootElement || box.IsTableWrapper {
		positionY += collapseMargin(adjoiningMargins)
		adjoiningMargins = nil
	}

	newBox_ := bo.CopyWithChildren(box_, newChildren, isStart, resumeAt == nil)
	newBox := newBox_.Box()
	// TODO: See corner cases in
	// http://www.w3.org/TR/CSS21/visudet.html#normal-block
	// TODO: See float.floatLayout
	if newBox.Height == pr.Auto {
		if len(context.excludedShapes) != 0 && newBox.Style.GetOverflow() != "visible" {
			maxFloatPositionY := -pr.Inf
			for _, floatBox := range context.excludedShapes {
				v := floatBox.PositionY + floatBox.MarginHeight()
				if v > maxFloatPositionY {
					maxFloatPositionY = v
				}
			}
			positionY = pr.Max(maxFloatPositionY, positionY)
		}
		newBox.Height = positionY - newBox.ContentBoxY()
	}

	if newBox.Style.GetPosition().String == "relative" {
		// New containing block, resolve the layout of the absolute descendants
		for _, absoluteBox := range *absoluteBoxes {
			absoluteLayout(context, absoluteBox, newBox_, fixedBoxes)
		}
	}

	for _, child := range newBox.Children {
		relativePositioning(child, bo.Point{newBox.Width.V(), newBox.Height.V()})
	}

	if !bo.TypeBlockBox.IsInstance(newBox_) {
		context.finishBlockFormattingContext(newBox_)
	}

	if resumeAt == nil {
		// After finishBlockFormattingContext which may increment
		// newBox.Height
		newBox.Height = pr.Max(pr.Min(newBox.Height.V(), newBox.MaxHeight.V()), newBox.MinHeight.V())
	} else {
		// Make the box fill the blank space at the bottom of the page
		// https://www.w3.org/TR/css-break-3/#box-splitting
		newBox.Height = maxPositionY - newBox.PositionY - (newBox.MarginHeight() - newBox.Height.V())
		if box.Style.GetBoxDecorationBreak() == "clone" {
			newBox.Height = newBox.Height.V() + box.PaddingBottom.V() + box.BorderBottomWidth.V() + box.MarginBottom.V()
		}
	}

	if nextPage.Page.IsNone() {
		_, nextPage.Page = newBox.PageValues()
	}

	return newBox_, blockLayout{resumeAt: resumeAt, nextPage: nextPage,
		adjoiningMargins: adjoiningMargins, collapsingThrough: collapsingThrough}
}

// Return the amount of collapsed margin for a list of adjoining margins.
func collapseMargin(adjoiningMargins []pr.Float) pr.Float {
	var maxPos, minNeg pr.Float
	for _, m := range adjoiningMargins {
		if m > maxPos {
			maxPos = m
		} else if m < minNeg {
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
func blockLevelPageBreak(siblingBefore, siblingAfter Box) string {
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

	return string(result)
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
func findEarlierPageBreak(children []Box, absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder) (newChildren []Box, resumeAt *tree.SkipStack) {
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
		resumeAt := &tree.SkipStack{Skip: 0, Stack: newChildren[len(newChildren)-1].Box().ResumeAt}
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
				resumeAt = &tree.SkipStack{Skip: children[index].Box().Index}
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
					resumeAt = &tree.SkipStack{Skip: newChild.Box().Index, Stack: resumeAt}
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

func removeBox(list *[]*AbsolutePlaceholder, box Box) {
	out := make([]*AbsolutePlaceholder, 0, len(*list))
	for _, v := range *list {
		if v != box {
			out = append(out, v)
		}
	}
	*list = out
}

// For boxes that have been removed in findEarlierPageBreak(),
// also remove the matching placeholders in absoluteBoxes and fixedBoxes.
func removePlaceholders(boxList []Box, absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder) {
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
