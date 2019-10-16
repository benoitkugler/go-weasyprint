package layout

import (
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
)

type PangoLayout struct {
	Text                 string
	JustificationSpacing float32
	Context              *LayoutContext
	Style                pr.Properties
}

type Splitted struct {
	// pango Layout with the first line
	Layout PangoLayout

	// length in UTF-8 bytes of the first line
	Length int

	// the number of UTF-8 bytes to skip for the next line.
	// May be ``None`` if the whole text fits in one line.
	// This may be greater than ``length`` in case of preserved
	// newline characters.
	ResumeAt *int

	// width in pixels of the first line
	Width float32

	// height in pixels of the first line
	Height float32

	// baseline in pixels of the first line
	Baseline float32
}

// Fit as much as possible in the available width for one line of text.
// minimum=False
func SplitFirstLine(text string, style pr.Properties, context interface{}, maxWidth *float32, justificationSpacing float32,
	minimum bool) Splitted {
	// FIXME: a implémenter
	return Splitted{}
}

func CanBreakText(text, lang string) bool {
	if len(text) < 2 {
		return false
	}
	// FIXME: à implémenter
	return true
}

// Return a tuple of the used value of ``line-height`` and the baseline.
// The baseline is given from the top edge of line height.
func StrutLayout(style pr.Properties, context *LayoutContext) (pr.Float, pr.Float) {
	// FIXME: à implémenter
	return 0.5, 0.5
}

// Return the ratio 1ex/font_size, according to given style.
func ExRatio(style pr.Properties) float32 {
	// FIXME: à implémenter
	return .5
}
