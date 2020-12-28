package pango

// Fontset represents a set of Font to use when rendering text.
// It is the result of resolving a FontDescription against a particular Context.
type Fontset interface {
	// Returns the font in the fontset that contains the best glyph for the Unicode character `wc`.
	get_font(wc rune) Font
	// Get overall metric information for the fonts in the fontset.
	get_metrics() FontMetrics
	// Returns the language of the fontset
	get_language() Language
	// Iterates through all the fonts in a fontset, calling `fn` for each one.
	// If `fn` returns `true`, that stops the iteration.
	foreach(fn FontsetForeachFunc)

	// get_font_cache returns the cache object associated
	// to this fontset. Use `NewFontCache` to create a cache when needed
	get_font_cache() *FontCache
}

// Returns `true` stops the iteration
type FontsetForeachFunc = func(set Fontset, font Font) bool
