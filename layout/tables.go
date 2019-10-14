package layout

import (
	bo "github.com/benoitkugler/go-weasyprint/boxes"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
)

// Layout for tables and internal table boxes.

// Find the width of each column && derive the wrapper width.
func tableWrapperWidth(context LayoutContext, wrapper *bo.BoxFields, containingBlock bo.MaybePoint) {
	table := wrapper.GetWrappedTable()
	resolvePercentages(table, containingBlock, "")

	if table.Box().Style.GetTableLayout() == "fixed" && table.Box().Width != Auto {
		fixedTableLayout(wrapper)
	} else {
		autoTableLayout(context, wrapper, containingBlock)
	}

	wrapper.Width = table.Box().BorderWidth()
}

func reverseBoxes(in []Box) []Box {
	N := len(in)
	out := make([]Box, N)
	for i, v := range in {
		out[N-1-i] = v
	}
	return out
}

// Return the absolute Y position for the first (or last) in-flow baseline
// if any or nil
// last=false, baselineTypes=(boxes.LineBox,)
func findInFlowBaseline(box Box, last bool, baselineTypes ...bo.BoxType) pr.MaybeFloat {
	if baselineTypes == nil {
		baselineTypes = []bo.BoxType{bo.TypeLineBox}
	}
	// TODO: synthetize baseline when needed
	// See https://www.w3.org/TR/css-align-3/#synthesize-baseline
	for _, type_ := range baselineTypes { // if isinstance(box, baselineTypes)
		if type_.IsInstance(box) {
			return pr.Float(box.Box().PositionY + box.Box().Baseline)
		}
	}
	if bo.IsParentBox(box) && !bo.TypeTableCaptionBox.IsInstance(box) {
		children := box.Box().Children
		if last {
			children = reverseBoxes(children)
		}
		for _, child := range children {
			if child.Box().IsInNormalFlow() {
				result := findInFlowBaseline(child, last, baselineTypes)
				if result != nil {
					return result
				}
			}
		}
	}
	return nil
}

// Distribute available width to columns.
//
// Return excess width left when it's impossible without breaking rules.
//
// See http://dbaron.org/css/intrinsic/#distributetocols
func distributeExcessWidth(context LayoutContext, grid [][]bo.Box, excessWidth float32, columnWidths []float32,
	constrainedness []bool, columnIntrinsicPercentages, columnMaxContentWidths []float32, columnSlice [2]int) float32 {

}
