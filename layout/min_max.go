package layout

import (
	bo "github.com/benoitkugler/go-weasyprint/structure"
)

type block struct {
	X,Y, Width, Height float32
}

func (b block) unpack() (float32, float32, float32, float32) {
	return b.X, b.Y, b.Width, b.Height
}

type funcMinMax = func(box bo.Box, context *Context, containingBlock block) (float32, float32)

// Decorate a function that sets the used width of a box to handle
// {min,max}-width.    
func handleMinMaxWidth(function funcMinMax) funcMinMax {
 wrapper := funcMinMax(box bo.Box, context *Context, containingBlock block) (float32, float32) {
        computedMarginL,computedMarginR = box.Box().marginLeft, box.Box().marginRight
        res1, res2 := function(box, context, containingBlock)
        if box.Box().width > box.Box().maxWidth {
            box.Box().width = box.Box().maxWidth
            box.Box().marginLeft, box.Box().marginRight = computedMarginL,computedMarginR
            res1, res2 = function(box, context, containingBlock)
		} 
		if box.Box().width < box.Box().minWidth {
            box.Box().width = box.Box().minWidth
            box.Box().marginLeft, box.Box().marginRight = computedMarginL,computedMarginR
            res1, res2 = function(box, context, containingBlock)
		} 
		return res1, res2
	} 
	// wrapper.withoutMinMax = function
    return wrapper
} 

// Decorate a function that sets the used height of a box to handle
// {min,max}-height. 
func handleMinMaxHeight(function funcMinMax) funcMinMax {
	wrapper := funcMinMax(box bo.Box, context *Context, containingBlock block) (float32, float32) {
        computedMarginT, computedMarginB := box.Box().marginTop, box.Box().marginBottom
         res1, res2 := function(box, context, containingBlock)
        if box.Box().height > box.Box().maxHeight {
            box.Box().height = box.Box().maxHeight
            box.Box().marginTop, box.Box().marginBottom = computedMarginT, computedMarginB
             res1, res2 = function(box, context, containingBlock)
        } if box.Box().height < box.Box().minHeight {
            box.Box().height = box.Box().minHeight
            box.Box().marginTop, box.Box().marginBottom = computedMarginT, computedMarginB
             res1, res2 = function(box, context, containingBlock)
		}
		 return  res1, res2
	}
	//  wrapper.withoutMinMax = function
    return wrapper
}