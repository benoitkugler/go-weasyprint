package pango

// Glyph represents a single glyph in the output form of a string.
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

// GlyphString structure is used to store strings
// of glyphs with geometry and visual attribute information - ready for drawing
type GlyphString struct {
	// array of glyph information for the glyph string
	// with size num_glyphs
	glyphs []GlyphInfo

	// logical cluster info, indexed by the byte index
	// within the text corresponding to the glyph string
	log_clusters []int

	space int
}

//  /**
//   * PANGO_TYPE_GLYPH_STRING:
//   *
//   * The #GObject type for #GlyphString.
//   */
//  #define PANGO_TYPE_GLYPH_STRING (pango_glyph_string_get_type ())

//  PANGO_AVAILABLE_IN_ALL
//  GlyphString *pango_glyph_string_new      (void);
//  PANGO_AVAILABLE_IN_ALL
//  void              pango_glyph_string_set_size (GlyphString *string,
// 							gint              new_len);
//  PANGO_AVAILABLE_IN_ALL
//  GType             pango_glyph_string_get_type (void) G_GNUC_CONST;
//  PANGO_AVAILABLE_IN_ALL
//  GlyphString *pango_glyph_string_copy     (GlyphString *string);
//  PANGO_AVAILABLE_IN_ALL
//  void              pango_glyph_string_free     (GlyphString *string);
//  PANGO_AVAILABLE_IN_ALL
//  void              pango_glyph_string_extents  (GlyphString *glyphs,
// 							PangoFont        *font,
// 							PangoRectangle   *ink_rect,
// 							PangoRectangle   *logical_rect);
//  PANGO_AVAILABLE_IN_1_14
//  int               pango_glyph_string_get_width(GlyphString *glyphs);

//  PANGO_AVAILABLE_IN_ALL
//  void              pango_glyph_string_extents_range  (GlyphString *glyphs,
// 							  int               start,
// 							  int               end,
// 							  PangoFont        *font,
// 							  PangoRectangle   *ink_rect,
// 							  PangoRectangle   *logical_rect);

//  PANGO_AVAILABLE_IN_ALL
//  void pango_glyph_string_get_logical_widths (GlyphString *glyphs,
// 						 const char       *text,
// 						 int               length,
// 						 int               embedding_level,
// 						 int              *logical_widths);

//  PANGO_AVAILABLE_IN_ALL
//  void pango_glyph_string_index_to_x (GlyphString *glyphs,
// 					 char             *text,
// 					 int               length,
// 					 PangoAnalysis    *analysis,
// 					 int               index_,
// 					 gboolean          trailing,
// 					 int              *x_pos);
//  PANGO_AVAILABLE_IN_ALL
//  void pango_glyph_string_x_to_index (GlyphString *glyphs,
// 					 char             *text,
// 					 int               length,
// 					 PangoAnalysis    *analysis,
// 					 int               x_pos,
// 					 int              *index_,
// 					 int              *trailing);

//  /* Turn a string of characters into a string of glyphs
//   */
//  PANGO_AVAILABLE_IN_ALL
//  void pango_shape (const char          *text,
// 				   int                  length,
// 				   const PangoAnalysis *analysis,
// 				   GlyphString    *glyphs);

//  PANGO_AVAILABLE_IN_1_32
//  void pango_shape_full (const char          *item_text,
// 						int                  item_length,
// 						const char          *paragraph_text,
// 						int                  paragraph_length,
// 						const PangoAnalysis *analysis,
// 						GlyphString    *glyphs);

//  /**
//   * PangoShapeFlags:
//   * @PANGO_SHAPE_NONE: Default value.
//   * @PANGO_SHAPE_ROUND_POSITIONS: Round glyph positions
//   *     and widths to whole device units. This option should
//   *     be set if the target renderer can't do subpixel
//   *     positioning of glyphs.
//   *
//   * Flags influencing the shaping process.
//   * These can be passed to pango_shape_with_flags().
//   */
//  typedef enum {
//    PANGO_SHAPE_NONE            = 0,
//    PANGO_SHAPE_ROUND_POSITIONS = 1 << 0,
//  } PangoShapeFlags;

//  PANGO_AVAILABLE_IN_1_44
//  void pango_shape_with_flags (const char          *item_text,
// 							  int                  item_length,
// 							  const char          *paragraph_text,
// 							  int                  paragraph_length,
// 							  const PangoAnalysis *analysis,
// 							  GlyphString    *glyphs,
// 							  PangoShapeFlags      flags);

//  PANGO_AVAILABLE_IN_ALL
//  GList *pango_reorder_items (GList *logical_items);

//  G_END_DECLS

//  #endif /* __GLYPH_H__ */
