package layout

import (
	bo "github.com/benoitkugler/go-weasyprint/boxes"
)

//     Line breaking and layout for inline-level boxes.

// Keep as much text as possible from a TextBox in a limited width.
//
// Try not to overflow but always have some text in ``new_box``
//
// Return ``(new_box, skip, preserved_line_break)``. ``skip`` is the number of
// UTF-8 bytes to skip form the start of the TextBox for the next line, or
// ``None`` if all of the text fits.
//
// Also break on preserved line breaks.
func splitTextBox(context LayoutContext, box bo.TextBox, availableWidth float32, skip *int) (*bo.TextBox, *int, bool) {

}

func inlineReplacedBoxWidthHeight(box bo.Box, containingBlock block) {
	if style := box.Box().Style; style.GetWidth().String == "auto" && style.GetHeight().String == "auto" {
		replacedBoxWidth.withoutMinMax(box, containingBlock)
		replacedBoxHeight.withoutMinMax(box)
		minMaxAutoReplaced(box)
	} else {
		replacedBoxWidth(box, containingBlock)
		replacedBoxHeight(box)
	}
}
