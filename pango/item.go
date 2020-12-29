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

func (analysis *Analysis) showing_space() bool {
	for _, attr := range analysis.extra_attrs {
		if attr.Type == ATTR_SHOW && (ShowFlags(attr.Data.(AttrInt))&PANGO_SHOW_SPACES != 0) {
			return true
		}
	}
	return false
}

// Item stores information about a segment of text.
type Item struct {
	offset    int      // byte offset of the start of this item in text.
	num_chars int      // number of Unicode characters in the item.
	analysis  Analysis // analysis results for the item.
}

//  #define PANGO_TYPE_ITEM (pango_item_get_type ())

// pango_item_apply_attrs add attributes to an `Item`. The idea is that you have
// attributes that don't affect itemization, such as font features,
// so you filter them out using pango_attr_list_filter(), itemize
// your text, then reapply the attributes to the resulting items
// using this function.
//
// `iter` should be positioned before the range of the item,
// and will be advanced past it. This function is meant to be called
// in a loop over the items resulting from itemization, while passing
// `iter` to each call.
func (item *Item) pango_item_apply_attrs(iter *AttrIterator) {

	compare_attr := func(a1, a2 *Attribute) bool {
		return a1.pango_attribute_equal(*a2) &&
			a1.StartIndex == a2.StartIndex &&
			a1.EndIndex == a2.EndIndex
	}

	var attrs AttrList

	isInList := func(data *Attribute) bool {
		for _, a := range attrs {
			if compare_attr(a, data) {
				return true
			}
		}
		return false
	}

	for do := true; do; do = iter.pango_attr_iterator_next() {
		start, end := iter.StartIndex, iter.EndIndex

		if start >= item.offset+item.num_chars {
			break
		}

		if end >= item.offset {
			list := iter.pango_attr_iterator_get_attrs()
			for _, data := range list {
				if !isInList(data) {
					attrs.insert(0, data.pango_attribute_copy())
				}
			}
		}

		if end >= item.offset+item.num_chars {
			break
		}
	}

	item.analysis.extra_attrs = append(item.analysis.extra_attrs, attrs...)
}

// Note that rise, letter_spacing, shape are constant across items,
// since we pass them into itemization.
//
// uline and strikethrough can vary across an item, so we collect
// all the values that we find.
type ItemProperties struct {
	uline_single       bool // = 1;
	uline_double       bool // = 1;
	uline_low          bool // = 1;
	uline_error        bool // = 1;
	strikethrough      bool // = 1;
	oline_single       bool // = 1;
	rise               int
	letter_spacing     int
	shape_set          bool
	shape_ink_rect     *Rectangle
	shape_logical_rect *Rectangle
}

func (item *Item) pango_layout_get_item_properties() ItemProperties {
	var properties ItemProperties
	for _, attr := range item.analysis.extra_attrs {
		switch attr.Type {
		case ATTR_UNDERLINE:
			switch Underline(attr.Data.(AttrInt)) {
			case PANGO_UNDERLINE_SINGLE, PANGO_UNDERLINE_SINGLE_LINE:
				properties.uline_single = true
			case PANGO_UNDERLINE_DOUBLE, PANGO_UNDERLINE_DOUBLE_LINE:
				properties.uline_double = true
			case PANGO_UNDERLINE_LOW:
				properties.uline_low = true
			case PANGO_UNDERLINE_ERROR, PANGO_UNDERLINE_ERROR_LINE:
				properties.uline_error = true
			}
		case ATTR_OVERLINE:
			switch Overline(attr.Data.(AttrInt)) {
			case PANGO_OVERLINE_SINGLE:
				properties.oline_single = true
			}
		case ATTR_STRIKETHROUGH:
			properties.strikethrough = attr.Data.(AttrInt) == 1
		case ATTR_RISE:
			properties.rise = int(attr.Data.(AttrInt))
		case ATTR_LETTER_SPACING:
			properties.letter_spacing = int(attr.Data.(AttrInt))
		case ATTR_SHAPE:
			s := attr.Data.(AttrShape)
			properties.shape_set = true
			properties.shape_logical_rect = &s.logical
			properties.shape_ink_rect = &s.ink
		}
	}
	return properties
}
