package layout

import (
	"log"

	bo "github.com/benoitkugler/go-weasyprint/boxes"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
)

// Layout for floating boxes.

var floatWidth = handleMinMaxWidth(floatWidth_)

// @handleMinMaxWidth
// containingBlock must be block
func floatWidth_(box Box, context *layoutContext, containingBlock containingBlock) (bool, pr.Float) {
	// Check that box.width is auto even if the caller does it too, because
	// the handleMinMaxWidth decorator can change the value
	if w := box.Box().Width; w == pr.Auto {
		box.Box().Width = shrinkToFit(context, box, containingBlock.(block).Width)
	}
	return false, 0
}

// Set the width and position of floating ``box``.
func floatLayout(context *layoutContext, box_ Box, containingBlock *bo.BoxFields, absoluteBoxes,
	fixedBoxes *[]*AbsolutePlaceholder) Box {
	cbWidth, cbHeight := containingBlock.Width, containingBlock.Height
	resolvePercentages(box_, bo.MaybePoint{cbWidth, cbHeight}, "")

	// TODO: This is only handled later in blocks.blockContainerLayout
	// http://www.w3.org/TR/CSS21/visudet.html#normal-block
	if cbHeight == pr.Auto {
		cbHeight = containingBlock.PositionY - containingBlock.ContentBoxY()
	}

	box := box_.Box()
	resolvePositionPercentages(box, bo.Point{cbWidth.V(), cbHeight.V()})

	if box.MarginLeft == pr.Auto {
		box.MarginLeft = pr.Float(0)
	}
	if box.MarginRight == pr.Auto {
		box.MarginRight = pr.Float(0)
	}
	if box.MarginTop == pr.Auto {
		box.MarginTop = pr.Float(0)
	}
	if box.MarginBottom == pr.Auto {
		box.MarginBottom = pr.Float(0)
	}

	clearance := getClearance(context, box, 0)
	if clearance != nil {
		box.PositionY += clearance.V()
	}

	if bo.BlockReplacedBoxT.IsInstance(box_) {
		inlineReplacedBoxWidthHeight(box_, containingBlock)
	} else if box.Width == pr.Auto {
		floatWidth(box_, context, block{Width: containingBlock.Width.V()})
	}

	if box.IsTableWrapper {
		tableWrapperWidth(context, box, bo.MaybePoint{cbWidth, cbHeight})
	}

	if bo.BlockContainerBoxT.IsInstance(box_) {
		context.createBlockFormattingContext()
		box_, _ = blockContainerLayout(context, box_, pr.Inf,
			nil, false, absoluteBoxes, fixedBoxes, nil, false)
		context.finishBlockFormattingContext(box_)
	} else if bo.FlexContainerBoxT.IsInstance(box_) {
		box_, _ = flexLayout(context, box_, pr.Inf, nil, containingBlock,
			false, absoluteBoxes, fixedBoxes)
	} else if !bo.BlockReplacedBoxT.IsInstance(box_) {
		log.Fatalf("expected BlockReplaced , got %v", box)
	}

	box_ = findFloatPosition(context, box_, containingBlock)

	*context.excludedShapes = append(*context.excludedShapes, box_.Box())

	return box_
}

// Get the right position of the float ``box``.
func findFloatPosition(context *layoutContext, box_ Box, containingBlock *bo.BoxFields) Box {
	box := box_.Box()
	// See http://www.w3.org/TR/CSS2/visuren.html#float-position

	// Point 4 is already handled as box.positionY is set according to the
	// containing box top position, with collapsing margins handled

	// Points 5 and 6, box.positionY is set to the highest positionY possible
	if L := len(*context.excludedShapes); L != 0 {
		highestY := (*context.excludedShapes)[L-1].PositionY
		if box.PositionY < highestY {
			box_.Translate(box_, 0, highestY-box.PositionY, false)
		}
	}

	// Points 1 and 2
	positionX, positionY, availableWidth := avoidCollisions(context, box_, containingBlock, true)

	// Point 9
	// positionY is set now, let's define positionX
	// for float: left elements, it's already done!
	if box.Style.GetFloat() == "right" {
		positionX += availableWidth - box.MarginWidth()
	}

	box_.Translate(box_, positionX-box.PositionX, positionY-box.PositionY, false)

	return box_
}

// Return nil if there is no clearance, otherwise the clearance value (as Float)
// collapseMargin = 0
func getClearance(context *layoutContext, box *bo.BoxFields, collapsedMargin pr.Float) (clearance pr.MaybeFloat) {
	hypotheticalPosition := box.PositionY + collapsedMargin
	// Hypothetical position is the position of the top border edge
	for _, excludedShape := range *context.excludedShapes {
		if clear := box.Style.GetClear(); clear == excludedShape.Style.GetFloat() || clear == "both" {
			y, h := excludedShape.PositionY, excludedShape.MarginHeight()
			if hypotheticalPosition < y+h {
				var safeClearance pr.Float
				if clearance != nil {
					safeClearance = clearance.V()
				}
				clearance = pr.Max(safeClearance, y+h-hypotheticalPosition)
			}
		}
	}
	return clearance
}

// outer=true
func avoidCollisions(context *layoutContext, box_ Box, containingBlock *bo.BoxFields, outer bool) (pr.Float, pr.Float, pr.Float) {
	excludedShapes := context.excludedShapes
	box := box_.Box()
	positionY := box.BorderBoxY()
	boxWidth := box.BorderWidth()
	boxHeight := box.BorderHeight()
	if outer {
		positionY = box.PositionY
		boxWidth = box.MarginWidth()
		boxHeight = box.MarginHeight()
	}

	if box.BorderHeight() == 0 && box.IsFloated() {
		return 0, 0, containingBlock.Width.V()
	}
	var maxLeftBound, maxRightBound pr.Float
	for {
		var collidingShapes []*bo.BoxFields
		for _, shape := range *excludedShapes {
			// Assign locals to avoid slow attribute lookups.
			shapePositionY := shape.PositionY
			shapeMarginHeight := shape.MarginHeight()
			if (shapePositionY < positionY && positionY < shapePositionY+shapeMarginHeight) ||
				(shapePositionY < positionY+boxHeight && positionY+boxHeight < shapePositionY+shapeMarginHeight) ||
				(shapePositionY >= positionY && shapePositionY+shapeMarginHeight <= positionY+boxHeight) {
				collidingShapes = append(collidingShapes, shape)
			}
		}
		var leftBounds, rightBounds []pr.Float
		for _, shape := range collidingShapes {
			if shape.Style.GetFloat() == "left" {
				leftBounds = append(leftBounds, shape.PositionX+shape.MarginWidth())
			}
			if shape.Style.GetFloat() == "right" {
				rightBounds = append(rightBounds, shape.PositionX)
			}

		}

		// Set the default maximum bounds
		maxLeftBound = containingBlock.ContentBoxX()
		maxRightBound = containingBlock.ContentBoxX() + containingBlock.Width.V()

		if !outer {
			maxLeftBound += box.MarginLeft.V()
			maxRightBound -= box.MarginRight.V()
		}

		// Set the real maximum bounds according to sibling float elements
		if len(leftBounds) != 0 || len(rightBounds) != 0 {
			if len(leftBounds) != 0 {
				maxLeftBound = pr.Max(pr.Maxs(leftBounds...), maxLeftBound)
			}
			if len(rightBounds) != 0 {
				maxRightBound = pr.Min(pr.Mins(rightBounds...), maxRightBound)
			}

			// Points 3, 7 && 8
			if boxWidth > maxRightBound-maxLeftBound {
				// The box does not fit here
				var min pr.Float
				for _, shape := range collidingShapes {
					if v := shape.PositionY + shape.MarginHeight(); v < min {
						min = v
					}
				}
				newPositonY := min

				if newPositonY > positionY {
					// We can find a solution with a higher positionY
					positionY = newPositonY
					continue
				} // No solution, we must put the box here
			}
		}
		break
	}
	positionX := maxLeftBound
	availableWidth := maxRightBound - maxLeftBound

	if !outer {
		positionX -= box.MarginLeft.V()
		positionY -= box.MarginTop.V()
	}

	return positionX, positionY, availableWidth
}
