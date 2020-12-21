package pango

// Gravity represents the orientation of glyphs in a segment
// of text.  This is useful when rendering vertical text layouts.  In
// those situations, the layout is rotated using a non-identity PangoMatrix,
// and then glyph orientation is controlled using Gravity.
//
// Not every value in this enumeration makes sense for every usage of
// Gravity; for example, `PANGO_GRAVITY_AUTO` only can be passed to
// pango_context_set_base_gravity() and can only be returned by
// pango_context_get_base_gravity().
//
// See also: GravityHint
type Gravity uint8

const (
	PANGO_GRAVITY_SOUTH Gravity = iota // Glyphs stand upright (default)
	PANGO_GRAVITY_EAST                 // Glyphs are rotated 90 degrees clockwise
	PANGO_GRAVITY_NORTH                // Glyphs are upside-down
	PANGO_GRAVITY_WEST                 // Glyphs are rotated 90 degrees counter-clockwise
	PANGO_GRAVITY_AUTO                 // Gravity is resolved from the context matrix
)
