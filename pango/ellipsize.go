package pango

// EllipsizeMode describes what sort of (if any)
// ellipsization should be applied to a line of text. In
// the ellipsization process characters are removed from the
// text in order to make it fit to a given width and replaced
// with an ellipsis.
type EllipsizeMode uint8

const (
	PANGO_ELLIPSIZE_NONE   EllipsizeMode = iota // No ellipsization
	PANGO_ELLIPSIZE_START                       // Omit characters at the start of the text
	PANGO_ELLIPSIZE_MIDDLE                      // Omit characters in the middle of the text
	PANGO_ELLIPSIZE_END                         // Omit characters at the end of the text
)
