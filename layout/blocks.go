package layout

import (
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
	adjoiningMargins  []float32
	collapsingThrough bool
}

// Set the ``box`` height.
func blockContainerLayout(context *LayoutContext, box bo.Box, maxPositionY float32, skipStack *bo.SkipStack,
	pageIsEmpty bool, absoluteBoxes *[]*AbsolutePlaceholder, fixedBoxes []bo.Box, adjoiningMargins []float32) blockLayout {
	// FIXME: a implémenter
	return blockLayout{}
}
