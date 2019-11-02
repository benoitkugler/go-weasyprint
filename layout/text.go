package layout

import (
	"github.com/benoitkugler/go-weasyprint/backend"
	bo "github.com/benoitkugler/go-weasyprint/boxes"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
)

type Splitted struct {
	// pango Layout with the first line
	Layout *bo.PangoLayout

	// length in UTF-8 bytes of the first line
	Length int

	// the number of UTF-8 bytes to skip for the next line.
	// May be ``None`` if the whole text fits in one line.
	// This may be greater than ``length`` in case of preserved
	// newline characters.
	ResumeAt *int

	// width in pixels of the first line
	Width pr.Float

	// height in pixels of the first line
	Height pr.Float

	// baseline in pixels of the first line
	Baseline pr.Float
}

// Return an opaque Pango layout with default Pango line-breaks.
// :param style: a style dict of computed values
// :param maxWidth: The maximum available width in the same unit as ``style['font_size']``,
// or ``None`` for unlimited width.
func createLayout(text string, style pr.Properties, context *LayoutContext, maxWidth pr.MaybeFloat, justificationSpacing pr.Float) *bo.PangoLayout {
	// FIXME: à implémenter
	return &bo.PangoLayout{}
}

// Fit as much as possible in the available width for one line of text.
// minimum=False
func SplitFirstLine(text []rune, style pr.Properties, context LayoutContext, maxWidth pr.MaybeFloat, justificationSpacing pr.Float,
	minimum bool) Splitted {
	// FIXME: a implémenter
	return Splitted{}
}

func CanBreakText(text, lang string) MaybeBool {
	if len(text) < 2 {
		return Bool(false)
	}
	// FIXME: à implémenter
	return Bool(true)
}

// Return a tuple of the used value of ``line-height`` and the baseline.
// The baseline is given from the top edge of line height.
func StrutLayout(style pr.Properties, context *LayoutContext) (pr.Float, pr.Float) {
	// FIXME: à implémenter
	return 0.5, 0.5
}

// Return the ratio 1ex/font_size, according to given style.
func ExRatio(style pr.Properties) pr.Float {
	// FIXME: à implémenter
	return .5
}

// Draw the given ``textbox`` line to the cairo ``context``.
func ShowFirstLine(context backend.Drawer, textbox bo.TextBox, textOverflow string) {
	// FIXME: à implémenter
	// pango.pangoLayoutSetSingleParagraphMode(textbox.PangoLayout.Layout, true)

	// if textOverflow == "ellipsis" {
	// 	maxWidth := context.ClipExtents()[2] - float64(textbox.PositionX)
	// 	pango.pangoLayoutSetWidth(textbox.PangoLayout.Layout, unitsFromDouble(maxWidth))
	// 	pango.pangoLayoutSetEllipsize(textbox.PangoLayout.Layout, pango.PANGOELLIPSIZEEND)
	// }

	// firstLine, _ = textbox.PangoLayout.GetFirstLine()
	// context = ffi.cast("cairoT *", context.Pointer)
	// pangocairo.pangoCairoShowLayoutLine(context, firstLine)
}
