package pango

import "github.com/benoitkugler/go-weasyprint/fribidi"

const (
	// Whether the segment should be shifted to center around the baseline.
	// Used in vertical writing directions mostly.
	PANGO_ANALYSIS_FLAG_CENTERED_BASELINE = 1 << 0
)

//  /**
//   * PANGO_ANALYSIS_FLAG_IS_ELLIPSIS:
//   *
//   * This flag is used to mark runs that hold ellipsized text,
//   * in an ellipsized layout.
//   *
//   * Since: 1.36.7
//   */
//  #define PANGO_ANALYSIS_FLAG_IS_ELLIPSIS (1 << 1)

//  /**
//   * PANGO_ANALYSIS_FLAG_NEED_HYPHEN:
//   *
//   * This flag tells Pango to add a hyphen at the end of the
//   * run during shaping.
//   *
//   * Since: 1.44
//   */
//  #define PANGO_ANALYSIS_FLAG_NEED_HYPHEN (1 << 2)

// Analysis stores information about the properties of a segment of text.
type Analysis struct {
	// shape_engine *PangoEngineShape
	// lang_engine  *PangoEngineLang
	font Font // the font for this segment.

	level   fribidi.Level //  the bidirectional level for this segment.
	gravity Gravity       //  the glyph orientation for this segment (A #PangoGravity).
	flags   uint8         //  boolean flags for this segment (Since: 1.16).

	script   Script   // the detected script for this segment (A #PangoScript) (Since: 1.18).
	language Language // the detected language for this segment.

	extra_attrs AttrList // extra attributes for this segment.
}

// Item stores information about a segment of text.
type Item struct {
	offset    int      // byte offset of the start of this item in text.
	length    int      // length of this item in bytes.
	num_chars int      // number of Unicode characters in the item.
	analysis  Analysis // analysis results for the item.
}

//  #define PANGO_TYPE_ITEM (pango_item_get_type ())
