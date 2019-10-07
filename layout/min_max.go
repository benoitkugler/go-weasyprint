package layout

import (
	bo "github.com/benoitkugler/go-weasyprint/boxes"
)

type block struct {
	X, Y, Width, Height float32
}

func (b block) unpack() (float32, float32, float32, float32) {
	return b.X, b.Y, b.Width, b.Height
}

type funcMinMax = func(box bo.Box, context LayoutContext, containingBlock block) (bool, float32)

// Decorate a function that sets the used width of a box to handle
// {min,max}-width.
func handleMinMaxWidth(function funcMinMax) funcMinMax {
	wrapper := func(box bo.Box, context LayoutContext, containingBlock block) (bool, float32) {
		computedMarginL, computedMarginR := box.Box().MarginLeft, box.Box().MarginRight
		res1, res2 := function(box, context, containingBlock)
		if box.Box().Width > box.Box().MaxWidth {
			box.Box().Width = box.Box().MaxWidth
			box.Box().MarginLeft, box.Box().MarginRight = computedMarginL, computedMarginR
			res1, res2 = function(box, context, containingBlock)
		}
		if box.Box().Width < box.Box().MinWidth {
			box.Box().Width = box.Box().MinWidth
			box.Box().MarginLeft, box.Box().MarginRight = computedMarginL, computedMarginR
			res1, res2 = function(box, context, containingBlock)
		}
		return res1, res2
	}
	// wrapper.WithoutMinMax = function
	return wrapper
}

// Decorate a function that sets the used height of a box to handle
// {min,max}-height.
func handleMinMaxHeight(function funcMinMax) funcMinMax {
	wrapper := func(box bo.Box, context LayoutContext, containingBlock block) (bool, float32) {
		computedMarginT, computedMarginB := box.Box().MarginTop, box.Box().MarginBottom
		res1, res2 := function(box, context, containingBlock)
		if box.Box().Height > box.Box().MaxHeight {
			box.Box().Height = box.Box().MaxHeight
			box.Box().MarginTop, box.Box().MarginBottom = computedMarginT, computedMarginB
			res1, res2 = function(box, context, containingBlock)
		}
		if box.Box().Height < box.Box().MinHeight {
			box.Box().Height = box.Box().MinHeight
			box.Box().MarginTop, box.Box().MarginBottom = computedMarginT, computedMarginB
			res1, res2 = function(box, context, containingBlock)
		}
		return res1, res2
	}
	//  wrapper.WithoutMinMax = function
	return wrapper
}
