package pango

// GlyphItem is a pair of a Item and the glyphs
// resulting from shaping the text corresponding to an item.
// As an example of the usage of GlyphItem, the results
// of shaping text with Layout is a list of LayoutLine,
// each of which contains a list of GlyphItem.
type GlyphItem struct {
	item   *Item
	glyphs *GlyphString
}
