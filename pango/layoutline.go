package pango

// LayoutLine represents one of the lines resulting
// from laying out a paragraph via `Layout`. `LayoutLine`
// structures are obtained by calling pango_layout_get_line() and
// are only valid until the text, attributes, or settings of the
// parent `Layout` are modified.
type LayoutLine struct {
	layout             *Layout     // the layout this line belongs to, might be nil
	start_index        int         /* start of line as byte index into layout->text */
	length             int         /* length of line in bytes */
	runs               []GlyphItem // list of runs in the line, from left to right
	is_paragraph_start bool        // = 1;  /* true if this is the first line of the paragraph */
	resolved_dir       Direction   // = 3;  /* Resolved PangoDirection of line */
}

// The resolved direction for the line is always one
// of LTR/RTL; not a week or neutral directions
func (line *LayoutLine) line_set_resolved_dir(direction Direction) {
	switch direction {
	case PANGO_DIRECTION_RTL, PANGO_DIRECTION_WEAK_RTL:
		line.resolved_dir = PANGO_DIRECTION_RTL
	default:
		line.resolved_dir = PANGO_DIRECTION_LTR
	}

	// The direction vs. gravity dance:
	//	- If gravity is SOUTH, leave direction untouched.
	//	- If gravity is NORTH, switch direction.
	//	- If gravity is EAST, set to LTR, as
	//	  it's a clockwise-rotated layout, so the rotated
	//	  top is unrotated left.
	//	- If gravity is WEST, set to RTL, as
	//	  it's a counter-clockwise-rotated layout, so the rotated
	//	  top is unrotated right.
	//
	// A similar dance is performed in pango-context.c:
	// itemize_state_add_character().  Keep in synch.
	switch line.layout.context.resolved_gravity {
	case PANGO_GRAVITY_NORTH:
		line.resolved_dir = PANGO_DIRECTION_LTR + PANGO_DIRECTION_RTL - line.resolved_dir
	case PANGO_GRAVITY_EAST:
		// This is in fact why deprecated TTB_RTL is LTR
		line.resolved_dir = PANGO_DIRECTION_LTR
	case PANGO_GRAVITY_WEST:
		// This is in fact why deprecated TTB_LTR is RTL
		line.resolved_dir = PANGO_DIRECTION_RTL
	}
}

func (line *LayoutLine) shape_run(state *ParaBreakState, item *Item) *GlyphString {
	layout := line.layout
	glyphs := &GlyphString{}

	if layout.text[item.offset] == '\t' {
		shape_tab(line, item, glyphs)
	} else {
		shape_flags := PANGO_SHAPE_NONE

		if pango_context_get_round_glyph_positions(layout.context) {
			shape_flags |= PANGO_SHAPE_ROUND_POSITIONS
		}
		if state.properties.shape_set {
			_pango_shape_shape(layout.text+item.offset, item.num_chars,
				state.properties.shape_ink_rect, state.properties.shape_logical_rect,
				glyphs)
		} else {
			glyphs.pango_shape_with_flags(layout.text+item.offset, item.length,
				layout.text, layout.length,
				&item.analysis, shape_flags)
		}

		if state.properties.letter_spacing {
			//    PangoGlyphItem glyph_item;
			//    int space_left, space_right;

			glyph_item.item = item
			glyph_item.glyphs = glyphs

			pango_glyph_item_letter_space(&glyph_item,
				layout.text,
				layout.log_attrs+state.start_offset,
				state.properties.letter_spacing)

			distribute_letter_spacing(state.properties.letter_spacing, &space_left, &space_right)

			glyphs.glyphs[0].geometry.width += space_left
			glyphs.glyphs[0].geometry.x_offset += space_left
			glyphs.glyphs[glyphs.num_glyphs-1].geometry.width += space_right
		}
	}

	return glyphs
}

func (line *LayoutLine) shape_tab(item *Item, glyphs *GlyphString) {
	current_width := line.line_width()

	glyphs.pango_glyph_string_set_size(1)

	if item.analysis.showing_space() {
		glyphs.glyphs[0].glyph = PANGO_GET_UNKNOWN_GLYPH('\t')
	} else {
		glyphs.glyphs[0].glyph = PANGO_GLYPH_EMPTY
	}
	glyphs.glyphs[0].geometry.x_offset = 0
	glyphs.glyphs[0].geometry.y_offset = 0
	glyphs.glyphs[0].attr.is_cluster_start = true

	glyphs.log_clusters[0] = 0

	line.layout.ensure_tab_width()
	space_width := line.layout.tab_width / 8

	for i := 0; ; i++ {
		//    bool is_default;
		tab_pos := get_tab_pos(line.layout, i, &is_default)
		/* Make sure there is at least a space-width of space between
		* tab-aligned text and the text before it.  However, only do
		* this if no tab array is set on the layout, ie. using default
		* tab positions.  If use has set tab positions, respect it to
		* the pixel.
		 */
		sw := 1
		if is_default {
			sw = space_width
		}
		if tab_pos >= current_width+sw {
			glyphs.glyphs[0].geometry.width = tab_pos - current_width
			break
		}
	}
}

func (line *LayoutLine) line_width() GlyphUnit {
	var width GlyphUnit

	// Compute the width of the line currently - inefficient, but easier
	// than keeping the current width of the line up to date everywhere
	for _, run := range line.runs {
		for _, info := range run.glyphs.glyphs {
			width += info.geometry.width
		}
	}

	return width
}

/* Extents cache status:
*
* LEAKED means that the user has access to this line structure or a
* run included in this line, and so can change the glyphs/glyph-widths.
* If this is true, extents caching will be disabled.
 */
const (
	NOT_CACHED uint8 = iota
	CACHED
	LEAKED
)

type LayoutLinePrivate struct {
	LayoutLine

	cache_status uint8
	ink_rect     Rectangle
	logical_rect Rectangle
	height       int
}

func (layout *Layout) pango_layout_line_new() *LayoutLinePrivate {
	var private LayoutLinePrivate
	private.LayoutLine.layout = layout
	return &private
}
