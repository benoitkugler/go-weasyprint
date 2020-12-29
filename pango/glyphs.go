package pango

import (
	"log"
	"unicode"
)

// Glyph represents a single glyph in the output form of a str.
type Glyph uint32

// pangoScale represents the scale between dimensions used
// for Pango distances and device units. (The definition of device
// units is dependent on the output device; it will typically be pixels
// for a screen, and points for a printer.) pangoScale is currently
// 1024, but this may be changed in the future.
//
// When setting font sizes, device units are always considered to be
// points (as in "12 point font"), rather than pixels.
const pangoScale = 1024

const PANGO_UNKNOWN_GLYPH_WIDTH = 10
const PANGO_UNKNOWN_GLYPH_HEIGHT = 14

// GlyphUnit is used to store dimensions within
// Pango. Dimensions are stored in 1/pangoScale of a device unit.
// (A device unit might be a pixel for screen display, or
// a point on a printer.) pangoScale is currently 1024, and
// may change in the future (unlikely though), but you should not
// depend on its exact value. .
type GlyphUnit int32

// Pixels converts from glyph units into device units with correct rounding.
func (g GlyphUnit) Pixels() int {
	return (int(g) + 512) >> 10
}

// PANGO_UNITS_ROUND rounds a dimension to whole device units, but does not
// convert it to device units.
func (d GlyphUnit) PANGO_UNITS_ROUND() GlyphUnit {
	return (d + pangoScale>>1) & ^(pangoScale - 1)
}

// GlyphGeometry contains width and positioning
// information for a single glyph.
type GlyphGeometry struct {
	width    GlyphUnit // the logical width to use for the the character.
	x_offset GlyphUnit // horizontal offset from nominal character position.
	y_offset GlyphUnit // vertical offset from nominal character position.
}

// GlyphVisAttr is used to communicate information between
// the shaping phase and the rendering phase.
// More attributes may be added in the future.
type GlyphVisAttr struct {
	// set for the first logical glyph in each cluster. (Clusters
	// are stored in visual order, within the cluster, glyphs
	// are always ordered in logical order, since visual
	// order is meaningless; that is, in Arabic text, accent glyphs
	// follow the glyphs for the base character.)
	is_cluster_start bool // =  1;
}

// GlyphInfo represents a single glyph together with
// positioning information and visual attributes.
type GlyphInfo struct {
	glyph    Glyph         // the glyph itself.
	geometry GlyphGeometry // the positional information about the glyph.
	attr     GlyphVisAttr  // the visual attributes of the glyph.
}

// ShapeFlags influences the shaping process.
// These can be passed to pango_shape_with_flags().
type ShapeFlags uint8

const (
	PANGO_SHAPE_NONE ShapeFlags = 0 // Default value.
	// Round glyph positions and widths to whole device units. This option should
	// be set if the target renderer can't do subpixel positioning of glyphs.
	PANGO_SHAPE_ROUND_POSITIONS ShapeFlags = 1
)

// GlyphString structure is used to store strings
// of glyphs with geometry and visual attribute information - ready for drawing
type GlyphString struct {
	// array of glyph information for the glyph string
	// with size num_glyphs
	glyphs []GlyphInfo

	// logical cluster info, indexed by the rune index
	// within the text corresponding to the glyph string
	log_clusters []int

	// space int
}

//  pango_glyph_string_set_size resize a glyph string to the given length.
func (str *GlyphString) pango_glyph_string_set_size(newLen int) {
	if newLen < 0 {
		return
	}
	// the C implementation does a much more careful re-allocation...
	str.glyphs = make([]GlyphInfo, newLen)
	str.log_clusters = make([]int, newLen)
}

func (glyphs GlyphString) reverse() {
	gs, lc := glyphs.glyphs, glyphs.log_clusters
	for i := len(gs)/2 - 1; i >= 0; i-- { // gs and lc have the same size
		opp := len(gs) - 1 - i
		gs[i], gs[opp] = gs[opp], gs[i]
		lc[i], lc[opp] = lc[opp], lc[i]
	}
}

// pango_glyph_string_get_width computes the logical width of the glyph string as can also be computed
// using pango_glyph_string_extents().  However, since this only computes the
// width, it's much faster.
// This is in fact only a convenience function that
// computes the sum of geometry.width for each glyph in `glyphs`.
func (glyphs *GlyphString) pango_glyph_string_get_width() GlyphUnit {
	var width GlyphUnit

	for _, g := range glyphs.glyphs {
		width += g.geometry.width
	}

	return width
}

func (glyphs *GlyphString) fallback_shape(text []rune, analysis *Analysis) {

	glyphs.pango_glyph_string_set_size(len(text))

	cluster := 0
	for i, wc := range text {
		if !unicode.Is(unicode.Mn, wc) {
			cluster = i
		}

		var glyph Glyph
		if pango_is_zero_width(wc) {
			glyph = PANGO_GLYPH_EMPTY
		} else {
			glyph = PANGO_GET_UNKNOWN_GLYPH(wc)
		}

		var logical_rect Rectangle
		analysis.font.get_glyph_extents(glyph, nil, &logical_rect)

		glyphs.glyphs[i].glyph = glyph

		glyphs.glyphs[i].geometry.x_offset = 0
		glyphs.glyphs[i].geometry.y_offset = 0
		glyphs.glyphs[i].geometry.width = GlyphUnit(logical_rect.width)

		glyphs.log_clusters[i] = cluster
	}

	if analysis.level&1 != 0 {
		glyphs.reverse()
	}
}

/**
 * pango_shape_with_flags:
 * @item_text: valid UTF-8 text to shape
 * @item_length: the length (in bytes) of @item_text.
 *     -1 means nul-terminated text.
 * @paragraph_text: (allow-none): text of the paragraph (see details).
 *     May be %NULL.
 * @paragraph_length: the length (in bytes) of @paragraph_text.
 *     -1 means nul-terminated text.
 * @analysis:  #PangoAnalysis structure from pango_itemize()
 * @glyphs: glyph string in which to store results
 * @flags: flags influencing the shaping process
 *
 * Given a segment of text and the corresponding
 * #PangoAnalysis structure returned from pango_itemize(),
 * convert the characters into glyphs. You may also pass
 * in only a substring of the item from pango_itemize().
 *
 * This is similar to pango_shape_full(), except it also takes
 * flags that can influence the shaping process.
 *
 * Note that the extra attributes in the @analyis that is returned from
 * pango_itemize() have indices that are relative to the entire paragraph,
 * so you do not pass the full paragraph text as @paragraph_text, you need
 * to subtract the item offset from their indices before calling
 * pango_shape_with_flags().
 */
func (glyphs *GlyphString) pango_shape_with_flags(item_text, paragraph_text []rune, analysis *Analysis,
	flags ShapeFlags) {

	if len(paragraph_text) == 0 {
		paragraph_text = item_text
	}

	// g_return_if_fail(paragraph_text <= item_text)
	// g_return_if_fail(paragraph_text+paragraph_length >= item_text+item_length)

	if analysis.font != nil {
		pango_hb_shape(analysis.font, item_text, item_length,
			analysis, glyphs, paragraph_text, paragraph_length)

		if len(glyphs.glyphs) == 0 {
			/* If a font has been correctly chosen, but no glyphs are output,
			* there's probably something wrong with the font.
			*
			* Trying to be informative, we print out the font description,
			* and the text, but to not flood the terminal with
			* zillions of the message, we set a flag to only err once per
			* font.
			 */
			// TODO:
			//    GQuark warned_quark = g_quark_from_static_string ("pango-shape-fail-warned");

			// if !g_object_get_qdata(G_OBJECT(analysis.font), warned_quark) {
			// 	PangoFontDescription * desc
			// 	char * font_name

			// 	desc = pango_font_describe(analysis.font)
			// 	font_name = pango_font_description_to_string(desc)
			// 	pango_font_description_free(desc)

			// 	g_warning("shaping failure, expect ugly output. font='%s', text='%.*s'",
			// 		font_name, item_length, item_text)

			// 	g_free(font_name)

			// 	g_object_set_qdata(G_OBJECT(analysis.font), warned_quark,
			// 		GINT_TO_POINTER(1))
			// }
		}
	}

	if len(glyphs.glyphs) == 0 {
		glyphs.fallback_shape(item_text, analysis)
		if len(glyphs.glyphs) == 0 {
			return
		}
	}

	/* make sure last_cluster is invalid */
	last_cluster := glyphs.log_clusters[0] - 1
	for i, lo := range glyphs.log_clusters {
		/* Set glyphs[i].attr.is_cluster_start based on log_clusters[] */
		if lo != last_cluster {
			glyphs.glyphs[i].attr.is_cluster_start = true
			last_cluster = lo
		} else {
			glyphs.glyphs[i].attr.is_cluster_start = false
		}

		/* Shift glyph if width is negative, and negate width.
		* This is useful for rotated font matrices and shouldn't
		* harm in normal cases.
		 */
		if glyphs.glyphs[i].geometry.width < 0 {
			glyphs.glyphs[i].geometry.width = -glyphs.glyphs[i].geometry.width
			glyphs.glyphs[i].geometry.x_offset += glyphs.glyphs[i].geometry.width
		}
	}

	// Make sure glyphstring direction conforms to analysis.level
	if lc := glyphs.log_clusters; (analysis.level&1) != 0 && lc[0] < lc[len(lc)-1] {
		log.Println("pango: expected RTL run but got LTR. Fixing.")

		// *Fix* it so we don't crash later
		glyphs.reverse()
	}

	if flags&PANGO_SHAPE_ROUND_POSITIONS != 0 {
		for i := range glyphs.glyphs {
			glyphs.glyphs[i].geometry.width = glyphs.glyphs[i].geometry.width.PANGO_UNITS_ROUND()
			glyphs.glyphs[i].geometry.x_offset = glyphs.glyphs[i].geometry.x_offset.PANGO_UNITS_ROUND()
			glyphs.glyphs[i].geometry.y_offset = glyphs.glyphs[i].geometry.y_offset.PANGO_UNITS_ROUND()
		}
	}
}

func (glyphs *GlyphString) _pango_shape_shape(text []rune, shapeLogical *Rectangle) {

	glyphs.pango_glyph_string_set_size(len(text))

	for i := range text {
		glyphs.glyphs[i].glyph = PANGO_GLYPH_EMPTY
		glyphs.glyphs[i].geometry.x_offset = 0
		glyphs.glyphs[i].geometry.y_offset = 0
		glyphs.glyphs[i].geometry.width = GlyphUnit(shapeLogical.width)
		glyphs.glyphs[i].attr.is_cluster_start = true
		glyphs.log_clusters[i] = i
	}
}
