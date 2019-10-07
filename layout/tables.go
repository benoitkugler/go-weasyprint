package layout

import bo "github.com/benoitkugler/go-weasyprint/boxes"

// Layout for tables and internal table boxes.

// Find the width of each column && derive the wrapper width.
func tableWrapperWidth(context LayoutContext, wrapper bo.Box, containingBlock bo.Point) {
	table := wrapper.Box().GetWrappedTable()
	resolvePercentages(table, containingBlock, "")

	if table.Box().Style.GetTableLayout() == "fixed" && table.Box().Width != Auto {
		fixedTableLayout(wrapper)
	} else {
		autoTableLayout(context, wrapper, containingBlock)
	}

	wrapper.Box().Width = table.Box().BorderWidth()
}
