package layout

import (
	bo "github.com/benoitkugler/go-weasyprint/boxes"
	"github.com/benoitkugler/go-weasyprint/utils"
)

// Preferred and minimum preferred width, aka. max-content and min-content
// width, aka. the shrink-to-fit algorithm.

// Terms used (max-content width, min-content width) are defined in David
// Baron's unofficial draft (http://dbaron.org/css/intrinsic/).

// Return the shrink-to-fit width of ``box``.
// *Warning:* both availableOuterWidth and the return value are
// for width of the *content area*, not margin area.
// http://www.w3.org/TR/CSS21/visudet.html#float-width
func shrinkToFit(context LayoutContext, box bo.Box, availableWidth float32) float32 {
	return utils.Min(utils.Max(minContentWidth(context, box, false), availableWidth), maxContentWidth(context, box, false))
}
