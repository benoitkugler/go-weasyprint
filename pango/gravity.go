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

var gravity_map = enumMap{
	{value: int(PANGO_GRAVITY_SOUTH), str: "Not-Rotated"},
	{value: int(PANGO_GRAVITY_SOUTH), str: "South"},
	{value: int(PANGO_GRAVITY_NORTH), str: "Upside-Down"},
	{value: int(PANGO_GRAVITY_NORTH), str: "North"},
	{value: int(PANGO_GRAVITY_EAST), str: "Rotated-Left"},
	{value: int(PANGO_GRAVITY_EAST), str: "East"},
	{value: int(PANGO_GRAVITY_WEST), str: "Rotated-Right"},
	{value: int(PANGO_GRAVITY_WEST), str: "West"},
}

// whether `g` represents vertical writing directions.
func (g Gravity) isVertical() bool {
	return g == PANGO_GRAVITY_EAST || g == PANGO_GRAVITY_WEST
}

// GravityHint defines how horizontal scripts should behave in a
// vertical context.  That is, English excerpt in a vertical paragraph for
// example.
type GravityHint uint8

const (
	PANGO_GRAVITY_HINT_NATURAL GravityHint = iota // scripts will take their natural gravity based on the base gravity and the script
	PANGO_GRAVITY_HINT_STRONG                     // always use the base gravity set, regardless of the script
	// For scripts not in their natural direction (eg.
	// Latin in East gravity), choose per-script gravity such that every script
	// respects the line progression.  This means, Latin and Arabic will take
	// opposite gravities and both flow top-to-bottom for example.
	PANGO_GRAVITY_HINT_LINE
)
