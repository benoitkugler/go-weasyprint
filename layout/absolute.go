package layout

import (
	"fmt"
	"log"
	"math"

	bo "github.com/benoitkugler/go-weasyprint/boxes"
)

// AbsolutePlaceholder is left where an absolutely-positioned box was taken out of the flow.
type AbsolutePlaceholder struct {
	bo.Box
	layoutDone bool
}

func NewAbsolutePlaceholder(box bo.Box) *AbsolutePlaceholder {
	out := AbsolutePlaceholder{Box: box, layoutDone: false}
	return &out
}

func (abs *AbsolutePlaceholder) setLaidOutBox(newBox bo.Box) {
	abs.Box = newBox
	abs.layoutDone = true
}

func (abs *AbsolutePlaceholder) translate(dx, dy float32, ignoreFloats bool) {
	if dx == 0 && dy == 0 {
		return
	}
	if abs.layoutDone {
		abs.Box.Translate(abs.Box, dx, dy, ignoreFloats)
	} else {
		// Descendants do not have a position yet.
		abs.Box.Box().PositionX += dx
		abs.Box.Box().PositionY += dy
	}
}

func (abs AbsolutePlaceholder) copy() AbsolutePlaceholder {
	return AbsolutePlaceholder{Box: abs.Box.Copy(), layoutDone: abs.layoutDone}
}

func (abs AbsolutePlaceholder) String() string {
	return fmt.Sprintf("<Placeholder %s>", abs.Box)
}

var absoluteWidth = handleMinMaxWidth(_absoluteWidth)

// @handleMinMaxWidth
func _absoluteWidth(box_ bo.Box, context LayoutContext, containingBlock block) (bool, float32) {
	// http://www.w3.org/TR/CSS2/visudet.html#abs-replaced-width
	box := box_.Box()
	// These names are waaay too long
	marginL := box.MarginLeft
	marginR := box.MarginRight
	paddingL := box.PaddingLeft
	paddingR := box.PaddingRight
	borderL := box.BorderLeftWidth
	borderR := box.BorderRightWidth
	width := box.Width
	left := box.Left
	right := box.Right

	cbX, _, cbWidth, _ := containingBlock.unpack()

	// TODO: handle bidi
	paddingPlusBordersX := paddingL + paddingR + borderL + borderR
	var translateX float32 = 0
	translateBoxWidth := false
	defaultTranslateX := cbX - box.PositionX
	if left == Auto && right == Auto && width == Auto {
		if marginL == Auto {
			box.MarginLeft = 0
		}
		if marginR == Auto {
			box.MarginRight = 0
		}
		availableWidth := cbWidth - (paddingPlusBordersX + box.MarginLeft + box.MarginRight)
		box.Width = shrinkToFit(context, box_, availableWidth)
	} else if left != Auto && right != Auto && width != Auto {
		widthForMargins := cbWidth - (right + left + paddingPlusBordersX)
		if marginL == Auto && marginR == Auto {
			if width+paddingPlusBordersX+right+left <= cbWidth {
				box.MarginLeft = widthForMargins / 2
				box.MarginRight = box.MarginLeft
			} else {
				box.MarginLeft = 0
				box.MarginRight = widthForMargins
			}
		} else if marginL == Auto {
			box.MarginLeft = widthForMargins
		} else if marginR == Auto {
			box.MarginRight = widthForMargins
		} else {
			box.MarginRight = widthForMargins
		}
		translateX = left + defaultTranslateX
	} else {
		if marginL == Auto {
			box.MarginLeft = 0
		}
		if marginR == Auto {
			box.MarginRight = 0
		}
		spacing := paddingPlusBordersX + box.MarginLeft + box.MarginRight
		if left == Auto && width == Auto {
			box.Width = shrinkToFit(context, box_, cbWidth-spacing-right)
			translateX = cbWidth - right - spacing + defaultTranslateX
			translateBoxWidth = true
		} else if left == Auto && right == Auto {
			// Keep the static position
		} else if width == Auto && right == Auto {
			box.Width = shrinkToFit(context, box_, cbWidth-spacing-left)
			translateX = left + defaultTranslateX
		} else if left == Auto {
			translateX = (cbWidth + defaultTranslateX - right - spacing - width)
		} else if width == Auto {
			box.Width = cbWidth - right - left - spacing
			translateX = left + defaultTranslateX
		} else if right == Auto {
			translateX = left + defaultTranslateX
		}
	}
	return translateBoxWidth, translateX
}

func absoluteHeight(box_ bo.Box, context LayoutContext, containingBlock block) (bool, float32) {
	box := box_.Box()
	// These names are waaay too long
	marginT := box.MarginTop
	marginB := box.MarginBottom
	paddingT := box.PaddingTop
	paddingB := box.PaddingBottom
	borderT := box.BorderTopWidth
	borderB := box.BorderBottomWidth
	height := box.Height
	top := box.Top
	bottom := box.Bottom

	_, cbY, _, cbHeight := containingBlock.unpack()

	// http://www.w3.org/TR/CSS2/visudet.html#abs-non-replaced-height

	paddingsPlusBordersY := paddingT + paddingB + borderT + borderB
	var translateY float32 = 0
	translateBoxHeight := false
	defaultTranslateY := cbY - box.PositionY
	if top == Auto && bottom == Auto && height == Auto {
		// Keep the static position
		if marginT == Auto {
			box.MarginTop = 0
		}
		if marginB == Auto {
			box.MarginBottom = 0
		}
	} else if top != Auto && bottom != Auto && height != Auto {
		heightForMargins := cbHeight - (top + bottom + paddingsPlusBordersY)
		if marginT == Auto && marginB == Auto {
			box.MarginTop = heightForMargins / 2
			box.MarginBottom = box.MarginTop
		} else if marginT == Auto {
			box.MarginTop = heightForMargins
		} else if marginB == Auto {
			box.MarginBottom = heightForMargins
		} else {
			box.MarginBottom = heightForMargins
		}
		translateY = top + defaultTranslateY
	} else {
		if marginT == Auto {
			box.MarginTop = 0
		}
		if marginB == Auto {
			box.MarginBottom = 0
		}
		spacing := paddingsPlusBordersY + box.MarginTop + box.MarginBottom
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
			box.Height = cbHeight - bottom - top - spacing
			translateY = top + defaultTranslateY
		} else if bottom == Auto {
			translateY = top + defaultTranslateY
		}
	}
	return translateBoxHeight, translateY
}

func absoluteBlock(context LayoutContext, box_ bo.Box, containingBlock block, fixedBoxes []bo.Box) bo.Box {
	box := box_.Box()
	_, _, cbWidth, cbHeight := containingBlock.unpack()

	translateBoxWidth, translateX := absoluteWidth(box_, context, containingBlock)
	translateBoxHeight, translateY := absoluteHeight(box_, context, containingBlock)

	// This box is the containing block for absolute descendants.
	var absoluteBoxes []bo.Box

	if box.IsTableWrapper {
		tableWrapperWidth(context, box, bo.Point{cbWidth, cbHeight})
	}

	newBox, _, _, _, _ := blockContainerLayout(context, box, pr.Inf, nil,
		false, absoluteBoxes, fixedBoxes, nil)

	for _, childPlaceholder := range absoluteBoxes {
		absoluteLayout(context, childPlaceholder, newBox, fixedBoxes)
	}

	if translateBoxWidth {
		translateX -= newBox.Width
	}
	if translateBoxHeight {
		translateY -= newBox.Height
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
//         translateX -= newBox.Width
//     } if translateBoxHeight {
//         translateY -= newBox.Height
//     }

//     newBox.translate(translateX, translateY)

//     return newBox

// Set the width of absolute positioned ``box``.
func absoluteLayout(context *LayoutContext, placeholder *AbsolutePlaceholder, containingBlock bo.Box, fixedBoxes []bo.Box) {
	if placeholder.layoutDone {
		log.Fatalf("placeholder can't have its layout done.")
	}
	box := placeholder.Box
	placeholder.setLaidOutBox(absoluteBoxLayout(context, box, containingBlock, fixedBoxes))
}

func absoluteBoxLayout(context *LayoutContext, box bo.Box, cb_ bo.Box, fixedBoxes []bo.Box) bo.Box {
	// TODO: handle inline boxes (point 10.1.4.1)
	// http://www.w3.org/TR/CSS2/visudet.html#containing-block-details
	var containingBlock block
	cb := cb_.Box()
	if _, isPageBox := cb_.(bo.InstancePageBox); isPageBox {
		containingBlock.X = cb.ContentBoxX()
		containingBlock.Y = cb.ContentBoxY()
		containingBlock.Width = cb.Width
		containingBlock.Height = cb.Height
	} else {
		containingBlock.X = cb.PaddingBoxX()
		containingBlock.Y = cb.PaddingBoxY()
		containingBlock.Width = cb.PaddingWidth()
		containingBlock.Height = cb.PaddingHeight()
	}

	resolvePercentages(box, bo.Point{containingBlock.Width, containingBlock.Height})
	resolvePositionPercentages(box, bo.Point{containingBlock.Width, containingBlock.Height})

	context.createBlockFormattingContext()
	// Absolute tables are wrapped into block boxes
	var newBox bo.Box
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
	return float32(int(math.Floor(float64(a))) / b)
}

func absoluteReplaced(context *LayoutContext, box_ bo.Box, containingBlock block) {
	inlineReplacedBoxWidthHeight(box, containingBlock)
	box := box_.Box()
	cbX, cbY, cbWidth, cbHeight := containingBlock.unpack()
	ltr := box.style.GetDirection() == "ltr"

	// http://www.w3.org/TR/CSS21/visudet.html#abs-replaced-width
	if box.left == Auto && box.right {
		// static position:
		if ltr {
			box.left = box.PositionX - cbX
		} else {
			box.right = cbX + cbWidth - box.PositionX
		}
	}
	if box.left == Auto || box.right == Auto {
		if box.MarginLeft == Auto {
			box.MarginLeft = 0
		}
		if box.MarginRight == Auto {
			box.MarginRight = 0
		}
		remaining := cbWidth - box.MarginWidth()
		if box.left == Auto {
			box.left = remaining - box.right
		}
		if box.right == Auto {
			box.right = remaining - box.left
		}
	} else if Auto == box.MarginLeft || Auto == box.MarginRight {
		remaining := cbWidth - (box.BorderWidth() + box.left + box.right)
		if box.MarginLeft == Auto && box.MarginRight == Auto {
			if remaining >= 0 {
				box.MarginLeft = intDiv(remaining, 2)
				box.MarginRight = box.MarginLeft
			} else if ltr {
				box.MarginLeft = 0
				box.MarginRight = remaining
			} else {
				box.MarginLeft = remaining
				box.MarginRight = 0
			}
		} else if box.MarginLeft == Auto {
			box.MarginLeft = remaining
		} else {
			box.MarginRight = remaining
		}
	} else {
		// Over-constrained
		if ltr {
			box.right = cbWidth - (box.MarginWidth() + box.left)
		} else {
			box.left = cbWidth - (box.MarginWidth() + box.right)
		}
	}

	// http://www.w3.org/TR/CSS21/visudet.html#abs-replaced-height
	if box.Top == Auto && box.Bottom == Auto {
		box.Top = box.PositionY - cbY
	}
	if box.Top == Auto || box.Bottom == Auto {
		if box.MarginTop == Auto {
			box.MarginTop = 0
		}
		if box.MarginBottom == Auto {
			box.MarginBottom = 0
		}
		remaining := cbHeight - box.MarginHeight()
		if box.Top == Auto {
			box.Top = remaining
		}
		if box.Bottom == Auto {
			box.Bottom = remaining
		}
	} else if box.MarginTop == Auto || box.MarginBottom == Auto {
		remaining := cbHeight - (box.BorderHeight() + box.Top + box.Bottom)
		if box.MarginTop == Auto && box.MarginBottom == Auto {
			box.MarginTop = intDiv(remaining, 2)
			box.MarginBottom = box.MarginTop
		} else if box.MarginTop == Auto {
			box.MarginTop = remaining
		} else {
			box.MarginBottom = remaining
		}
	} else {
		// Over-constrained
		box.Bottom = cbHeight - (box.MarginHeight() + box.Top)
	}

	// No children for replaced boxes, no need to .translate()
	box.PositionX = cbX + box.left
	box.PositionY = cbY + box.Top
	return box_
}
