package layout

import (
	"log"

	bo "github.com/benoitkugler/go-weasyprint/boxes"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
)

// Layout for columns.

type boxWithType interface {
	bo.InstanceBlockBox
	Type() bo.BoxType
}

// if box is nil, then represents a list
type boxOrList struct {
	box  bo.InstanceBlockLevelBox
	list []Box
}

// Lay out a multi-column ``box``.
func columnsLayout(context *LayoutContext, box_ boxWithType, maxPositionY pr.Float, skipStack *bo.SkipStack, containingBlock Box,
	pageIsEmpty bool, absoluteBoxes []*AbsolutePlaceholder, fixedBoxes []Box, adjoiningMargins []pr.Float) blockLayout {
	// Implementation of the multi-column pseudo-algorithm :
	// https://www.w3.org/TR/css3-multicol/#pseudo-algorithm
	style := box_.Box().Style
	originalMaxPositionY := maxPositionY

	if style.GetPosition().String == "relative" {
		// New containing block, use a new absolute list
		absoluteBoxes = nil
	}

	box_ = bo.CopyWithChildren(box_, box_.Box().Children, true, true).(boxWithType) // CopyWithChildren preserves the concrete type of box_
	box := box_.Box()
	box.PositionY += collapseMargin(adjoiningMargins) - box.MarginTop

	height_ := box.Style.GetHeight()
	knownHeight := false
	if height_.String != "auto" && height_.Unit != pr.Percentage {
		if height_.Unit != pr.Px {
			log.Fatalf("expected Px got %v", height_)
		}
		knownHeight = true
		maxPositionY = pr.Min(maxPositionY, box.ContentBoxY()+height_.Value)
	}
	// TODO: the available width can be unknown if the containing block needs
	// the size of this block to know its own size.
	blockLevelWidth(box, containingBlock)
	availableWidth := box.Width.V()
	var (
		count_, width pr.Float
		count         int
	)
	cwidth, cccount, cgap := style.GetColumnWidth(), style.GetColumnCount(), style.GetColumnGap().Value
	if cwidth.String == "auto" && cccount.String != "auto" {
		count = int(cccount.Value)
		count_ = pr.Float(count)
		width = pr.Max(0, availableWidth-(count_-1)*cgap) / count_
	} else if cwidth.String != "auto" && cccount.String == "auto" {
		count = int(pr.Max(1, pr.Floor((availableWidth+cgap)/(cwidth.Value+cgap))))
		count_ = pr.Float(count)
		width = (availableWidth+cgap)/count_ - cgap
	} else {
		count = int(pr.Min(cccount.Value, pr.Floor((availableWidth+cgap)/(cwidth.Value+cgap))))
		count_ = pr.Float(count)
		width = (availableWidth+cgap)/count_ - cgap
	}

	createColumnBox := func(children []Box) bo.InstanceBlockBox {
		columnBox := box_.Type().AnonymousFrom(box_, children).(bo.InstanceBlockBox) // AnonymousFrom preserves concrete types
		resolvePercentages2(columnBox, containingBlock, "")
		columnBox.Box().IsColumn = true
		columnBox.Box().Width = width
		columnBox.Box().PositionX = box.ContentBoxX()
		columnBox.Box().PositionY = box.ContentBoxY()
		return columnBox
	}

	// Handle column-span property.
	// We want to get the following structure :
	// columnsAndBlocks = [
	//     [columnChild1, columnChild2],
	//     spanningBlock,
	//     …
	// ]
	var (
		columnsAndBlocks []boxOrList
		columnChildren   []Box
	)
	for _, child := range box.Children {
		if child.Box().Style.GetColumnSpan() == "all" {
			if len(columnChildren) != 0 {
				columnsAndBlocks = append(columnsAndBlocks, boxOrList{list: columnChildren})
			}
			columnsAndBlocks = append(columnsAndBlocks, boxOrList{box: child.Copy().(bo.InstanceBlockLevelBox)})
			columnChildren = nil
			continue
		}
		columnChildren = append(columnChildren, child.Copy())
	}
	if len(columnChildren) != 0 {
		columnsAndBlocks = append(columnsAndBlocks, boxOrList{list: columnChildren})
	}

	var nextPage page
	if len(box.Children) == 0 {
		nextPage = page{break_: "any", page: nil}
		skipStack = nil
	}

	// Balance.
	//
	// The current algorithm starts from the ideal height (the total height
	// divided by the number of columns). We then iterate until the last column
	// is not the highest one. At the end of each loop, we add the minimal
	// height needed to make one direct child at the top of one column go to the
	// end of the previous column.
	//
	// We rely on a real rendering for each loop, and with a stupid algorithm
	// like this it can last minutes…

	adjoiningMargins = nil
	currentPositionY := box.ContentBoxY()
	var newChildren []Box
	for _, columnChildrenOrBlock := range columnsAndBlocks {
		if block := columnChildrenOrBlock.box; block != nil {
			// We get a spanning block, we display it like other blocks.
			resolvePercentages2(block, containingBlock, "")
			block.Box().PositionX = box.ContentBoxX()
			block.Box().PositionY = currentPositionY
			tmp := blockLevelLayout(*context, block, originalMaxPositionY, skipStack,
				containingBlock, pageIsEmpty, absoluteBoxes, fixedBoxes, adjoiningMargins)
			newChild, adjoiningMargins := tmp.newBox, tmp.adjoiningMargins
			newChildren = append(newChildren, newChild)
			currentPositionY = newChild.Box().BorderHeight() + newChild.Box().BorderBoxY()
			adjoiningMargins = append(adjoiningMargins, newChild.Box().MarginBottom.V())
			continue
		}

		excludedShapes := append([]shape{}, context.excludedShapes...)

		// We have a list of children that we have to balance between columns.
		columnChildren := columnChildrenOrBlock.list

		// Find the total height of the content
		currentPositionY += collapseMargin(adjoiningMargins)
		adjoiningMargins = nil
		columnBox := createColumnBox(columnChildren)
		newChild := blockBoxLayout(*context, columnBox, pr.Inf, skipStack, containingBlock,
			pageIsEmpty, nil, nil, nil).newBox
		height := newChild.Box().MarginHeight()
		if style.GetColumnFill() == "balance" {
			height /= count_
		}

		// Try to render columns until the content fits, increase the column
		// height step by step.
		columnSkipStack := skipStack
		lostSpace := pr.Inf
		for {
			// Remove extra excluded shapes introduced during previous loop
			context.excludedShapes = context.excludedShapes[:len(excludedShapes)]

			for i := 0; i < count; i += 1 {
				// Render the column
				tmp := blockBoxLayout(*context, columnBox, box.ContentBoxY()+height,
					columnSkipStack, containingBlock, pageIsEmpty, nil, nil, nil)
				newBox, resumeAt := tmp.newBox, tmp.resumeAt
				nextPage = tmp.nextPage
				if newBox == nil {
					// We didn"t render anything. Give up and use the max
					// content height.
					height *= count_
					continue
				}
				columnSkipStack = resumeAt

				var lastInFlowChildren *bo.BoxFields
				for _, child := range newBox.Box().Children {
					if ch := child.Box(); ch.IsInNormalFlow() {
						lastInFlowChildren = ch
					}
				}

				var emptySpace, nextBoxSize pr.Float
				if lastInFlowChildren != nil {
					// Get the empty space at the bottom of the column box
					emptySpace = height - (lastInFlowChildren.PositionY - box.ContentBoxY() + lastInFlowChildren.MarginHeight())

					// Get the minimum size needed to render the next box
					nextBox := blockBoxLayout(*context, columnBox, box.ContentBoxY(),
						columnSkipStack, containingBlock, true, nil, nil, nil).newBox
					for _, child := range nextBox.Box().Children {
						if child.Box().IsInNormalFlow() {
							nextBoxSize = child.Box().MarginHeight()
							break
						}
					}
				} else {
					emptySpace = 0
					nextBoxSize = 0
				}

				// Append the size needed to render the next box in this
				// column.
				//
				// The next box size may be smaller than the empty space, for
				// example when the next box can't be separated from its own
				// next box. In this case we don't try to find the real value
				// and let the workaround below fix this for us.
				//
				// We also want to avoid very small values that may have been
				// introduced by rounding errors. As the workaround below at
				// least adds 1 pixel for each loop, we can ignore lost spaces
				// lower than 1px.
				if nextBoxSize-emptySpace > 1 {
					lostSpace = pr.Min(lostSpace, nextBoxSize-emptySpace)
				}

				// Stop if we already rendered the whole content
				if resumeAt == nil {
					break
				}
			}

			if columnSkipStack == nil {
				// We rendered the whole content, stop
				break
			} else {
				if lostSpace == pr.Inf {
					// We didn't find the extra size needed to render a child in
					// the previous column, increase height by the minimal
					// value.
					height += 1
				} else {
					// Increase the columns heights and render them once again
					height += lostSpace
				}
				columnSkipStack = skipStack
			}
		}

		// TODO: check style["max"]-height
		maxPositionY = pr.Min(maxPositionY, box.ContentBoxY()+height)

		// Replace the current box children with columns
		i := 0
		var maxColumnHeight pr.Float
		var columns []Box
		for {
			i_ := pr.Float(i)
			if i == count-1 {
				maxPositionY = originalMaxPositionY
			}
			columnBox = createColumnBox(columnChildren)
			columnBox.Box().PositionY = currentPositionY
			if style.GetDirection() == "rtl" {
				columnBox.Box().PositionX += (box.Width.V() - (i_+1)*width - i_*style.GetColumnGap().Value)
			} else {
				columnBox.Box().PositionX += i_ * (width + style.GetColumnGap().Value)
			}
			tmp := blockBoxLayout(*context, columnBox, maxPositionY, skipStack,
				containingBlock, pageIsEmpty, absoluteBoxes,
				fixedBoxes, nil)
			newChild, columnSkipStack = tmp.newBox, tmp.resumeAt
			columnNextPage := tmp.nextPage
			if newChild == nil {
				break
			}
			nextPage = columnNextPage
			skipStack = columnSkipStack
			columns = append(columns, newChild)
			maxColumnHeight = pr.Max(maxColumnHeight, newChild.Box().MarginHeight())
			if skipStack == nil {
				break
			}
			i += 1
			if i == count && !knownHeight {
				// [If] a declaration that constrains the column height
				// (e.g., using height || max-height). In this case,
				// additional column boxes are created in the inline
				// direction.
				break
			}
		}

		currentPositionY += maxColumnHeight
		for _, column := range columns {
			column.Box().Height = maxColumnHeight
			newChildren = append(newChildren, column)
		}
	}

	if len(box.Children) != 0 && len(newChildren) == 0 {
		// The box has children but none can be drawn, let's skip the whole box
		return blockLayout{resumeAt: &bo.SkipStack{Skip: 0}, nextPage: page{break_: "any", page: nil}}
	}

	// Set the height of box and the columns
	box.Children = newChildren
	currentPositionY += collapseMargin(adjoiningMargins)
	var heightDifference pr.Float
	if box.Height.Auto() {
		box.Height = currentPositionY - box.PositionY
		heightDifference = 0
	} else {
		heightDifference = box.Height.V() - (currentPositionY - box.PositionY)
	}
	if !box.MinHeight.Auto() && box.MinHeight.V() > box.Height.V() {
		heightDifference += box.MinHeight.V() - box.Height.V()
		box.Height = box.MinHeight
	}
	for _, child := range reversed(newChildren) {
		if child.Box().IsColumn {
			child.Box().Height = child.Box().Height.V() + heightDifference
		} else {
			break
		}
	}

	if box.Style.GetPosition().String == "relative" {
		// New containing block, resolve the layout of the absolute descendants
		for _, absoluteBox := range absoluteBoxes {
			absoluteLayout(context, absoluteBox, box_, fixedBoxes)
		}
	}

	return blockLayout{newBox: box_, resumeAt: skipStack, nextPage: nextPage}
}
