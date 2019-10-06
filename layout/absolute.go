package layout

import (
	bo "github.com/benoitkugler/go-weasyprint/structure"
)

// avoid to use interface or struct
const Auto float32 = -1784619812,12384158963

// AbsolutePlaceholder is left where an absolutely-positioned box was taken out of the flow.
type AbsolutePlaceholder struct {
	bo.Box 
	layoutDone bool 
}

func NewAbsolutePlaceholder(box bo.Box) *AbsolutePlaceholder {
	out := AbsolutePlaceholder{Box:box, layoutDone: false}
    return &out
} 

 func (abs *AbsolutePlaceholder) setLaidOutBox( newBox Box) {
    abs.Box = newBox
    abs.layoutDone = true
 }

 func (abs *AbsolutePlaceholder) translate( dx, dy float32, ignoreFloats bool) {
    if dx == 0 && dy == 0 {
        return
	}
	 if abs.layoutDone {
        abs.Box.translate(abs.Box, dx, dy, ignoreFloats)
    } else {
        // Descendants do not have a position yet.
        abs.Box.Box().positionX += dx
        abs.Box.Box().positionY += dy
    }
 }

 func (abs AbsolutePlaceholder) copy( {
    return AbsolutePlaceholder{Box: abs.Box.Copy(), layoutDone:abs.layoutDone}
 }

 func (abs AbsolutePlaceholder) String() string {
    return fmt.Sprintf("<Placeholder %s>", abs.Box)
 }


var absoluteWidth = handleMinMaxWidth(_absoluteWidth)

// @handleMinMaxWidth
func _absoluteWidth(box_ Box, context *Context, containingBlock block) (bool, float32){
    // http://www.w3.org/TR/CSS2/visudet.html#abs-replaced-width
	box := box_.Box()
    // These names are waaay too long
    marginL := box.marginLeft
    marginR := box.marginRight
    paddingL := box.paddingLeft
    paddingR := box.paddingRight
    borderL := box.borderLeftWidth
    borderR := box.borderRightWidth
    width := box.width
    left := box.left
    right := box.right

    cbX, cbY, cbWidth, cbHeight := containingBlock.unpack()

    // TODO: handle bidi
    paddingPlusBordersX := paddingL + paddingR + borderL + borderR
    translateX := 0
    translateBoxWidth := false
    defaultTranslateX := cbX - box.positionX
    if left == Auto && right == Auto && width == Auto {
        if marginL == Auto  {
            box.marginLeft = 0
		}
		 if marginR == Auto  {
            box.marginRight = 0
		}
		 availableWidth := cbWidth - (paddingPlusBordersX + box.marginLeft + box.marginRight)
        box.width = shrinkToFit(context, box, availableWidth)
    } else if left != Auto && right != Auto && width != Auto {
        widthForMargins := cbWidth - (right + left + paddingPlusBordersX)
        if marginL== Auto && marginR == Auto {
            if width + paddingPlusBordersX + right + left <= cbWidth {
				box.marginLeft =  widthForMargins / 2
				box.marginRight =box.marginLeft 
            } else {
                box.marginLeft = 0
                box.marginRight = widthForMargins
            }
        } else if marginL == Auto {
            box.marginLeft = widthForMargins
        } else if marginR == Auto {
            box.marginRight = widthForMargins
        } else {
            box.marginRight = widthForMargins
		} 
		translateX = left + defaultTranslateX
    } else {
        if marginL == Auto {
            box.marginLeft = 0
        } if marginR == Auto {
            box.marginRight = 0
		}
		spacing := paddingPlusBordersX + box.marginLeft + box.marginRight
        if left ==  Auto && width == Auto {
            box.width = shrinkToFit(context, box, cbWidth - spacing - right)
            translateX = cbWidth - right - spacing + defaultTranslateX
            translateBoxWidth = true
        } else if left == Auto && right == Auto {
            // Keep the static position
        } else if width == Auto && right == Auto {
            box.width = shrinkToFit(context, box, cbWidth - spacing - left)
            translateX = left + defaultTranslateX
        } else if left == Auto {
            translateX = (cbWidth + defaultTranslateX - right - spacing - width)
        } else if width == Auto {
            box.width = cbWidth - right - left - spacing
            translateX = left + defaultTranslateX
        } else if right == Auto {
            translateX = left + defaultTranslateX
        }
    }
    return translateBoxWidth, translateX
}

func absoluteHeight(box_ Box, context *Context, containingBlock block) (bool, float32){
	box := box_.Box()
	// These names are waaay too long
    marginT := box.marginTop
    marginB := box.marginBottom
    paddingT := box.paddingTop
    paddingB := box.paddingBottom
    borderT := box.borderTopWidth
    borderB := box.borderBottomWidth
    height := box.height
    top := box.top
    bottom := box.bottom
 
    cbX, cbY, cbWidth, cbHeight := containingBlock.unpack()

    // http://www.w3.org/TR/CSS2/visudet.html#abs-non-replaced-height

    paddingsPlusBordersY := paddingT + paddingB + borderT + borderB
    translateY := 0
    translateBoxHeight := false
    defaultTranslateY := cbY - box.positionY
    if top == Auto && bottom == Auto && height == Auto {
        // Keep the static position
        if marginT == Auto {
            box.marginTop = 0
		} 
		if marginB == Auto {
            box.marginBottom = 0
        }
    } else if top != Auto && bottom != Auto && height != Auto {
        heightForMargins := cbHeight - (top + bottom + paddingsPlusBordersY)
        if marginT ==Auto && marginB == Auto {
			box.marginTop =  heightForMargins / 2
			box.marginBottom = box.marginTop
        } else if marginT == Auto {
            box.marginTop = heightForMargins
        } else if marginB == Auto {
            box.marginBottom = heightForMargins
        } else {
            box.marginBottom = heightForMargins
		} 
		translateY = top + defaultTranslateY
    } else {
        if marginT == Auto {
            box.marginTop = 0
        } if marginB == Auto {
            box.marginBottom = 0
		} 
		spacing = paddingsPlusBordersY + box.marginTop + box.marginBottom
        if top == Auto && height == Auto {
            translateY = cbHeight - bottom - spacing + defaultTranslateY
            translateBoxHeight = true
        } else if top == Auto && bottom == Auto {
            // Keep the static position
        } else if height == Auto && bottom == Auto {
            translateY = top + defaultTranslateY
        } else if top == Auto {
            translateY = (cbHeight + defaultTranslateY - bottom - spacing - height)
        } else if height == Auto {
            box.height = cbHeight - bottom - top - spacing
            translateY = top + defaultTranslateY
        } else if bottom == Auto {
            translateY = top + defaultTranslateY
        }
    }
    return translateBoxHeight, translateY
}

func absoluteBlock(context *Context, box_ bo.Box, containingBlock block, fixedBoxes []bo.Box) bo.Box {
	box := box_.Box()
	cbX, cbY, cbWidth, cbHeight := containingBlock.unpack()

    translateBoxWidth, translateX := absoluteWidth( box, context, containingBlock)
    translateBoxHeight, translateY := absoluteHeight( box, context, containingBlock)

    // This box is the containing block for absolute descendants.
    var absoluteBoxes []bo.Box

    if box.isTableWrapper {
        tableWrapperWidth(context, box, bo.Point{cbWidth, cbHeight})
    }

    newBox, _, _, _, _ := blockContainerLayout(context, box, Inf, nil,
        false, absoluteBoxes,fixedBoxes, nil)

    for _, childPlaceholder := range absoluteBoxes {
        absoluteLayout(context, childPlaceholder, newBox, fixedBoxes)
    }

    if translateBoxWidth {
        translateX -= newBox.width
    } if translateBoxHeight {
        translateY -= newBox.height
    }

    newBox.translate(translateX, translateY)

    return newBox
}

// FIXME: waiting for weasyprint update
// func absoluteFlex(context, box, containingBlockSizes, fixedBoxes,
//                   containingBlock) {
//                   }
//     // Avoid a circular import
//     from .flex import flexLayout

//     // TODO: this function is really close to absoluteBlock, we should have
//     // only one function.
//     // TODO: having containingBlockSizes && containingBlock is stupid.
//     cbX, cbY, cbWidth, cbHeight = containingBlockSizes

//     translateBoxWidth, translateX = absoluteWidth(
//         box, context, containingBlockSizes)
//     translateBoxHeight, translateY = absoluteHeight(
//         box, context, containingBlockSizes)

//     // This box is the containing block for absolute descendants.
//     absoluteBoxes = []

//     if box.isTableWrapper {
//         tableWrapperWidth(context, box, (cbWidth, cbHeight))
//     }

//     newBox, _, _, _, _ = flexLayout(
//         context, box, maxPositionY=float("inf"), skipStack=None,
//         containingBlock=containingBlock, pageIsEmpty=false,
//         absoluteBoxes=absoluteBoxes, fixedBoxes=fixedBoxes)

//     for childPlaceholder := range absoluteBoxes {
//         absoluteLayout(context, childPlaceholder, newBox, fixedBoxes)
//     }

//     if translateBoxWidth {
//         translateX -= newBox.width
//     } if translateBoxHeight {
//         translateY -= newBox.height
//     }

//     newBox.translate(translateX, translateY)

//     return newBox


// Set the width of absolute positioned ``box``.
func absoluteLayout(context *Context,placeholder *AbsolutePlaceholder, containingBlock block, fixedBoxes []Box) {
    if placeholder.layoutDone {
		log.Fatalf("placeholder can't have its layout done.")
	}
    box := placeholder.Box
    placeholder.setLaidOutBox(absoluteBoxLayout(context, box, containingBlock, fixedBoxes))
} 

func absoluteBoxLayout(context *Context, box Box, cb Box, fixedBoxes []Box) Box {
    // TODO: handle inline boxes (point 10.1.4.1)
    // http://www.w3.org/TR/CSS2/visudet.html#containing-block-details
	var containingBlock block
	if _, isPageBox := containingBlock.(bo.InstancePageBox); isPageBox {
    	containingBlock.X = cb.contentBoxX()
    	containingBlock.Y = cb.contentBoxY()
    	containingBlock.Width = cb.width
    	containingBlock.Height = cb.height
    } else {
    	containingBlock.X = cb.paddingBoxX()
    	containingBlock.Y = cb.paddingBoxY()
    	containingBlock.Width = cb.paddingWidth()
    	containingBlock.Height = cb.paddingHeight()
	} 

    resolvePercentages(box, bo.Point{cbWidth, cbHeight})
    resolvePositionPercentages(box, bo.Point{cbWidth, cbHeight})

    context.createBlockFormattingContext()
	// Absolute tables are wrapped into block boxes
	var newBox Box
	_, isFlexCont := box.(bo.FlexContainerBox)
    if bo.TypeBlockBox.Isinstance(box) {
        newBox = absoluteBlock(context, box, containingBlock, fixedBoxes)
    } else if isFlexCont {
        newBox = absoluteFlex(context, box, containingBlock, fixedBoxes, cb)
    } else {
		if !bo.TypeBlockReplacedBox.Isinstance(box) {
			log.Fatalf("box should be a BlockReplaced, got %s", box)
		}
        newBox = absoluteReplaced(context, box, containingBlock)
	} 
	context.finishBlockFormattingContext(newBox)
    return newBox
	}

func intDiv(a float32, b int) float32 {
	return float32(int(math.Floor(float64(remaining))) / b)
}

func absoluteReplaced(context *Context, box_ Box, containingBlock block) {
    inlineReplacedBoxWidthHeight(box, containingBlock)
	box := box_.Box()
    cbX, cbY, cbWidth, cbHeight := containingBlock.unpack()
    ltr := box.style.GetDirection() == "ltr"

    // http://www.w3.org/TR/CSS21/visudet.html#abs-replaced-width
    if box.left == Auto && box.right {
        // static position:
        if ltr {
            box.left = box.positionX - cbX
        } else {
            box.right = cbX + cbWidth - box.positionX
        }
	}
	 if box.left == Auto ||  box.right == Auto {
        if box.marginLeft == Auto {
            box.marginLeft = 0
		}
		 if box.marginRight == Auto {
            box.marginRight = 0
		}
		 remaining := cbWidth - box.marginWidth()
        if box.left == Auto {
            box.left = remaining - box.right
		}
		 if box.right == Auto {
            box.right = remaining - box.left
        }
    } else if Auto == box.marginLeft || Auto == box.marginRight {
        remaining := cbWidth - (box.borderWidth() + box.left + box.right)
        if box.marginLeft == Auto && box.marginRight == Auto {
            if remaining >= 0 {
                box.marginLeft = intDiv(remaining, 2)
				box.marginRight = box.marginLeft
            } else if ltr {
                box.marginLeft = 0
                box.marginRight = remaining
            } else {
                box.marginLeft = remaining
                box.marginRight = 0
            }
        } else if box.marginLeft == Auto {
            box.marginLeft = remaining
        } else {
            box.marginRight = remaining
        }
    } else {
        // Over-constrained
        if ltr {
            box.right = cbWidth - (box.marginWidth() + box.left)
        } else {
            box.left = cbWidth - (box.marginWidth() + box.right)
        }
    }

    // http://www.w3.org/TR/CSS21/visudet.html#abs-replaced-height
    if box.top == Auto && box.bottom == Auto {
        box.top = box.positionY - cbY
	} 
	if box.top == Auto ||  box.bottom== Auto{
        if box.marginTop == Auto {
            box.marginTop = 0
		}
		 if box.marginBottom == Auto {
            box.marginBottom = 0
		}
		 remaining := cbHeight - box.marginHeight()
        if box.top == Auto {
            box.top = remaining
		} 
		if box.bottom == Auto {
            box.bottom = remaining
        }
    } else if box.marginTop == Auto || box.marginBottom == Auto {
        remaining := cbHeight - (box.borderHeight() + box.top + box.bottom)
        if box.marginTop == Auto && box.marginBottom == Auto {
			box.marginTop =intDiv(remaining, 2)
			box.marginBottom = box.marginTop
        } else if box.marginTop == Auto {
            box.marginTop = remaining
        } else {
            box.marginBottom = remaining
        }
    } else {
        // Over-constrained
        box.bottom = cbHeight - (box.marginHeight() + box.top)
    }

    // No children for replaced boxes, no need to .translate()
    box.positionX = cbX + box.left
    box.positionY = cbY + box.top
    return box_
}