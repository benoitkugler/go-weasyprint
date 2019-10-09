package pdf

import (
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
)

type Splitted struct {
	// ``layout``: a pango Layout with the first line

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
