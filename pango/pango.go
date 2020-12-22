package pango

// Rectangle represents a rectangle. It is frequently
// used to represent the logical or ink extents of a single glyph or section
// of text. (See, for instance, pango_font_get_glyph_extents())
type Rectangle struct {
	x      int // X coordinate of the left side of the rectangle.
	y      int // Y coordinate of the the top side of the rectangle.
	width  int // width of the rectangle.
	height int // height of the rectangle.
}
