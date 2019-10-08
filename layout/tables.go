package layout

import bo "github.com/benoitkugler/go-weasyprint/boxes"

// Layout for tables and internal table boxes.

// Find the width of each column && derive the wrapper width.
func tableWrapperWidth(context LayoutContext, wrapper *bo.BoxFields, containingBlock bo.Point) {
	table := wrapper.GetWrappedTable()
	resolvePercentages(table, containingBlock, "")

	if table.Box().Style.GetTableLayout() == "fixed" && table.Box().Width != Auto {
		fixedTableLayout(wrapper)
	} else {
		autoTableLayout(context, wrapper, containingBlock)
	}

	wrapper.Width = table.Box().BorderWidth()
}

// Distribute available width to columns.
//
// Return excess width left when it's impossible without breaking rules.
//
// See http://dbaron.org/css/intrinsic/#distributetocols
func distributeExcessWidth(context LayoutContext, grid [][]bo.Box, excessWidth float32, columnWidths []float32,
	constrainedness []bool, columnIntrinsicPercentages, columnMaxContentWidths []float32, columnSlice [2]int) float32 {

}
