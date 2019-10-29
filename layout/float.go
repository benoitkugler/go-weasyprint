package layout

import (
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	bo "github.com/benoitkugler/go-weasyprint/boxes"
)



// Layout for floating boxes.

var floatWidth = handleMinMaxWidth(floatWidth_)

// @handleMinMaxWidth
func floatWidth_(box Box, context LayoutContext, containingBlock block) (bool, float32) {
    // Check that box.width is auto even if the caller does it too, because
    // the handleMinMaxWidth decorator can change the value
	if w := box.Box().Width; w== pr.Auto {
        box.Box().Width = bo.MF(shrinkToFit(context, box, containingBlock.Width))
	}
	return false,0 
} 

// Set the width and position of floating ``box``.
func floatLayout(context *LayoutContext, box_, containingBlock_ Box, absoluteBoxes , 
	fixedBoxes *[]*AbsolutePlaceholder) Box {
		containingBlock := containingBlock_.Box()
    cbWidth, cbHeight := containingBlock.Width, containingBlock.Height
    resolvePercentages(box, bo.MaybePoint{cbWidth, cbHeight}, "")

    // TODO: This is only handled later in blocks.blockContainerLayout
    // http://www.w3.org/TR/CSS21/visudet.html#normal-block
    if cbHeight== pr.Auto {
        cbHeight = pr.Float(containingBlock.PositionY - containingBlock.ContentBoxY())
	}

	box := box_.Box()
    resolvePositionPercentages(box, bo.MaybePoint{cbWidth, cbHeight})

    if box.MarginLeft== pr.Auto {
        box.MarginLeft = pr.Float(0)
	} 
	if box.MarginRight== pr.Auto {
        box.MarginRight = pr.Float(0)
	} 
	if box.MarginTop== pr.Auto {
        box.MarginTop = pr.Float(0)
	} 
	if box.MarginBottom== pr.Auto {
        box.MarginBottom = pr.Float(0)
    }

    clearance := getClearance(context, box, 0)
    if clearance != nil {
        box.PositionY += clearance.V()
    }

    if bo.IsBlockReplacedBox(box_) {
        inlineReplacedBoxWidthHeight(box, containingBlock)
    } else if box.Width== pr.Auto {
        floatWidth(box_, context, block{Width: containingBlock.Width.V()})
    }

    if box.IsTableWrapper {
        tableWrapperWidth(context, box, bo.MaybePoint{cbWidth, cbHeight})
    }

    if bo.IsBlockContainerBox(box_) {
        context.createBlockFormattingContext()
        box = blockContainerLayout(context, box, pr.Inf,
            nil, false, absoluteBoxes, fixedBoxes, nil).newBox
        context.finishBlockFormattingContext(box_)
    } else if bo.IsFlexContainerBox(box_) {
        box = flexLayout(context, box_, pr.Inf, nil, containingBlock_,
            false, absoluteBoxes, fixedBoxes).newBox
    } else {
        assert isinstance(box, boxes.BlockReplacedBox)
    }

    box = findFloatPosition(context, box, containingBlock)

    context.excludedShapes.append(box)

    return box


// Get the right position of the float ``box``.
func findFloatPosition(context, box, containingBlock) {
    // See http://www.w3.org/TR/CSS2/visuren.html#float-position

    // Point 4 is already handled as box.positionY is set according to the
    // containing box top position, with collapsing margins handled

    // Points 5 && 6, box.positionY is set to the highest positionY possible
    if context.excludedShapes {
        highestY = context.excludedShapes[-1].positionY
        if box.positionY < highestY {
            box.translate(0, highestY - box.positionY)
        }
    }

    // Points 1 && 2
    positionX, positionY, availableWidth = avoidCollisions(
        context, box, containingBlock)

    // Point 9
    // positionY is set now, let"s define positionX
    // for float: left elements, it"s already done!
    if box.style["float"] == "right" {
        positionX += availableWidth - box.marginWidth()
    }

    box.translate(positionX - box.positionX, positionY - box.positionY)

    return box
	}

// Return nil if there is no clearance, otherwise the clearance value (as Float)
// collapseMargin = 0
func getClearance(context LayoutContext, box bo.BoxFields, collapsedMargin float32) (clearance pr.MaybeFloat) {
    hypotheticalPosition := box.PositionY + collapsedMargin
    // Hypothetical position is the position of the top border edge
    for _, excludedShape := range context.excludedShapes {
        if clear := box.Style.GetClear(); clear == excludedShape.Style.GetFloat() || clear == "both" {
            y, h := excludedShape.positionY, excludedShape.marginHeight()
            if hypotheticalPosition < y + h {
				var safeClearance float32
				if clearance != nil {
					safeClearance = clearance.V()
				}
                clearance = pr.Float(utils.Max(safeClearance, y + h - hypotheticalPosition))
            }
        }
	} 
	return clearance
} 

func avoidCollisions(context, box, containingBlock, outer=true) {
    excludedShapes = context.excludedShapes
    positionY = box.positionY if outer else box.borderBoxY()

    boxWidth = box.marginWidth() if outer else box.borderWidth()
    boxHeight = box.marginHeight() if outer else box.borderHeight()

    if box.borderHeight() == 0 && box.isFloated() {
        return 0, 0, containingBlock.width
    }

    while true {
        collidingShapes = []
        for shape := range excludedShapes {
            // Assign locals to avoid slow attribute lookups.
            shapePositionY = shape.positionY
            shapeMarginHeight = shape.marginHeight()
            if ((shapePositionY < positionY <
                 shapePositionY + shapeMarginHeight) or
                (shapePositionY < positionY + boxHeight <
                 shapePositionY + shapeMarginHeight) or
                (shapePositionY >= positionY and
                 shapePositionY + shapeMarginHeight <=
                 positionY + boxHeight)) {
                 }
                collidingShapes.append(shape)
        } leftBounds = [
            shape.positionX + shape.marginWidth()
            for shape := range collidingShapes
            if shape.style["float"] == "left"]
        rightBounds = [
            shape.positionX
            for shape := range collidingShapes
            if shape.style["float"] == "right"]
    }

        // Set the default maximum bounds
        maxLeftBound = containingBlock.contentBoxX()
        maxRightBound = \
            containingBlock.contentBoxX() + containingBlock.width

        if not outer {
            maxLeftBound += box.marginLeft
            maxRightBound -= box.marginRight
        }

        // Set the real maximum bounds according to sibling float elements
        if leftBounds || rightBounds {
            if leftBounds {
                maxLeftBound = max(max(leftBounds), maxLeftBound)
            } if rightBounds {
                maxRightBound = min(min(rightBounds), maxRightBound)
            }
        }

            // Points 3, 7 && 8
            if boxWidth > maxRightBound - maxLeftBound {
                // The box does not fit here
                newPositonY = min(
                    shape.positionY + shape.marginHeight()
                    for shape := range collidingShapes)
                if newPositonY > positionY {
                    // We can find a solution with a higher positionY
                    positionY = newPositonY
                    continue
                } // No solution, we must put the box here
            }
        break

    positionX = maxLeftBound
    availableWidth = maxRightBound - maxLeftBound

    if not outer {
        positionX -= box.marginLeft
        positionY -= box.marginTop
    }

    return positionX, positionY, availableWidth
}