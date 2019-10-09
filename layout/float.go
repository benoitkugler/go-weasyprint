package layout

import (
	bo "github.com/benoitkugler/go-weasyprint/boxes"
)



// Layout for floating boxes.

var floatWidth = handleMinMaxWidth(floatWidth_)

// @handleMinMaxWidth
func floatWidth_(box Box, context LayoutContext, containingBlock block) (bool, float32) {
    // Check that box.width is auto even if the caller does it too, because
    // the handleMinMaxWidth decorator can change the value
	if w := box.Box().Width; w.Auto() {
        box.Box().Width = bo.MF(shrinkToFit(context, box, containingBlock.Width))
	}
	return false,0 
} 

// Set the width and position of floating ``box``.
func floatLayout(context LayoutContext, box, containingBlock_ Box, absoluteBoxes []AbsolutePlaceholder, 
	fixedBoxes []Box) {
		containingBlock := containingBlock_.Box()
    cbWidth, cbHeight := containingBlock.Width, containingBlock.Height
    resolvePercentages(box, bo.Point{cbWidth, cbHeight}, "")

    // TODO: This is only handled later in blocks.blockContainerLayout
    // http://www.w3.org/TR/CSS21/visudet.html#normal-block
    if cbHeight == "auto" {
        cbHeight = (
            containingBlock.positionY - containingBlock.contentBoxY())
    }

    resolvePositionPercentages(box, (cbWidth, cbHeight))

    if box.marginLeft == "auto" {
        box.marginLeft = 0
    } if box.marginRight == "auto" {
        box.marginRight = 0
    } if box.marginTop == "auto" {
        box.marginTop = 0
    } if box.marginBottom == "auto" {
        box.marginBottom = 0
    }

    clearance = getClearance(context, box)
    if clearance is not None {
        box.positionY += clearance
    }

    if isinstance(box, boxes.BlockReplacedBox) {
        inlineReplacedBoxWidthHeight(box, containingBlock)
    } else if box.width == "auto" {
        floatWidth(box, context, containingBlock)
    }

    if box.isTableWrapper {
        tableWrapperWidth(context, box, (cbWidth, cbHeight))
    }

    if isinstance(box, boxes.BlockContainerBox) {
        context.createBlockFormattingContext()
        box, _, _, _, _ = blockContainerLayout(
            context, box, maxPositionY=float("inf"),
            skipStack=None, pageIsEmpty=false,
            absoluteBoxes=absoluteBoxes, fixedBoxes=fixedBoxes,
            adjoiningMargins=None)
        context.finishBlockFormattingContext(box)
    } else if isinstance(box, boxes.FlexContainerBox) {
        box, _, _, _, _ = flexLayout(
            context, box, maxPositionY=float("inf"),
            skipStack=None, containingBlock=containingBlock,
            pageIsEmpty=false, absoluteBoxes=absoluteBoxes,
            fixedBoxes=fixedBoxes)
    } else {
        assert isinstance(box, boxes.BlockReplacedBox)
    }

    box = findFloatPosition(context, box, containingBlock)

    context.excludedShapes.append(box)

    return box


// Get the right position of the float ``box``.
func findFloatPosition(context, box, containingBlock) {
    // See http://www.w3.org/TR/CSS2/visuren.html#float-position
} 
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


// Return None if there is no clearance, otherwise the clearance value.
func getClearance(context, box, collapsedMargin=0) {
    clearance = None
    hypotheticalPosition = box.positionY + collapsedMargin
    // Hypothetical position is the position of the top border edge
    for excludedShape := range context.excludedShapes {
        if box.style["clear"] := range (excludedShape.style["float"], "both") {
            y, h = excludedShape.positionY, excludedShape.marginHeight()
            if hypotheticalPosition < y + h {
                clearance = max(
                    (clearance || 0), y + h - hypotheticalPosition)
            }
        }
    } return clearance
} 

func avoidCollisions(context, box, containingBlock, outer=true) {
    excludedShapes = context.excludedShapes
    positionY = box.positionY if outer else box.borderBoxY()
} 
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