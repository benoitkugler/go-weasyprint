package pdf

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
func splitFirstLine(text, style, context, maxWidth, justificationSpacing,
	minimum bool) Splitted {

}
