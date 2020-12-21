package pango

const (
	ATINVALID             AttrType = iota // does not happen
	ATLANGUAGE                            // language (AttrLanguage)
	ATFAMILY                              // font family name list (AttrString)
	ATSTYLE                               // font slant style (AttrInt)
	ATWEIGHT                              // font weight (AttrInt)
	ATVARIANT                             // font variant (normal or small caps) (AttrInt)
	ATSTRETCH                             // font stretch (AttrInt)
	ATSIZE                                // font size in points scaled by %PANGO_SCALE (AttrInt)
	ATFONT_DESC                           // font description (AttrFontDesc)
	ATFOREGROUND                          // foreground color (AttrColor)
	ATBACKGROUND                          // background color (AttrColor)
	ATUNDERLINE                           // whether the text has an underline (AttrInt)
	ATSTRIKETHROUGH                       // whether the text is struck-through (AttrInt)
	ATRISE                                // baseline displacement (AttrInt)
	ATSHAPE                               // shape (AttrShape)
	ATSCALE                               // font size scale factor (AttrFloat)
	ATFALLBACK                            // whether fallback is enabled (AttrInt)
	ATLETTER_SPACING                      // letter spacing (AttrInt)
	ATUNDERLINE_COLOR                     // underline color (AttrColor)
	ATSTRIKETHROUGH_COLOR                 // strikethrough color (AttrColor)
	ATABSOLUTE_SIZE                       // font size in pixels scaled by %PANGO_SCALE (AttrInt)
	ATGRAVITY                             // base text gravity (AttrInt)
	ATGRAVITY_HINT                        // gravity hint (AttrInt)
	ATFONT_FEATURES                       // OpenType font features (AttrString). Since 1.38
	ATFOREGROUND_ALPHA                    // foreground alpha (AttrInt). Since 1.38
	ATBACKGROUND_ALPHA                    // background alpha (AttrInt). Since 1.38
	ATALLOW_BREAKS                        // whether breaks are allowed (AttrInt). Since 1.44
	ATSHOW                                // how to render invisible characters (AttrInt). Since 1.44
	ATINSERT_HYPHENS                      // whether to insert hyphens at intra-word line breaks (AttrInt). Since 1.44
	ATOVERLINE                            // whether the text has an overline (AttrInt). Since 1.46
	ATOVERLINE_COLOR                      // overline color (AttrColor). Since 1.46
)

type AttrType uint8

type AttrData interface {
	copy() AttrData // TODO a vérifier
}

type AttrInt int

func (a AttrInt) copy() AttrData { return a }

func (a FontDescription) copy() AttrData { return a }

type Attr struct {
	Type                 AttrType
	Data                 AttrData
	StartIndex, EndIndex uint32
}

// Initializes StartIndex to 0 and EndIndex to MaxUint
// such that the attribute applies to the entire text by default.
func (attr *Attr) pango_attribute_init() {
	attr.StartIndex = 0
	attr.EndIndex = ^uint32(0)
}

// Compare two attributes for equality. This compares only the
// actual value of the two attributes and not the ranges that the
// attributes apply to.
func (attr1 Attr) equals(attr2 Attr) bool {
	return attr1.Type == attr2.Type && attr1.Data == attr2.Data // TODO: à relacher, need equals method
}

// Make a copy of an attribute.
func (a Attr) copy() Attr {
	out := a
	out.Data = a.Data.copy()
	return out
}

// Create a new font description attribute. This attribute
// allows setting family, style, weight, variant, stretch,
// and size simultaneously.
func pango_attr_font_desc_new(desc FontDescription) Attr {
	//    PangoAttrFontDesc *result = g_slice_new (PangoAttrFontDesc);
	//    pango_attribute_init (&result->attr, &klass);
	//    result->desc = pango_font_description_copy (desc);

	out := Attr{Type: ATFONT_DESC, Data: desc}
	out.pango_attribute_init()
	return out
}

type AttrList []Attr

// Copy @list and return an identical new list.
func (list AttrList) pango_attr_list_copy() AttrList {
	out := make(AttrList, len(list))
	for i, v := range list {
		out[i] = v.copy()
	}
	return out
}

func (list *AttrList) pango_attr_list_insert_internal(attr Attr, before bool) {
	startIndex := attr.StartIndex

	if len(*list) == 0 {
		*list = append(*list, attr)
		return
	}

	lastAttr := (*list)[len(*list)-1]

	if lastAttr.StartIndex < startIndex || (!before && lastAttr.StartIndex == startIndex) {
		*list = append(*list, attr)
	} else {
		for i, cur := range *list {
			if cur.StartIndex > startIndex || (before && cur.StartIndex == startIndex) {
				list.insert(i, attr)
				break
			}
		}
	}
}

func (l *AttrList) insert(i int, attr Attr) {
	*l = append(*l, Attr{})
	copy((*l)[i+1:], (*l)[i:])
	(*l)[i] = attr
}

func (l *AttrList) remove(i int) {
	copy((*l)[i:], (*l)[i+1:])
	(*l)[len((*l))-1] = Attr{}
	(*l) = (*l)[:len((*l))-1]
}

// Insert the given attribute into the list. It will
// be inserted after all other attributes with a matching
// `StartIndex`.
func (list *AttrList) pango_attr_list_insert(attr Attr) {
	if list == nil {
		return
	}
	list.pango_attr_list_insert_internal(attr, false)
}

// Insert the given attribute into the `AttrList`. It will
// be inserted before all other attributes with a matching
// `StartIndex`.
func (list *AttrList) pango_attr_list_insert_before(attr Attr) {
	if list == nil {
		return
	}
	list.pango_attr_list_insert_internal(attr, true)
}

// Insert the given attribute into the `AttrList`. It will
// replace any attributes of the same type on that segment
// and be merged with any adjoining attributes that are identical.
//
// This function is slower than pango_attr_list_insert() for
// creating an attribute list in order (potentially much slower
// for large lists). However, pango_attr_list_insert() is not
// suitable for continually changing a set of attributes
// since it never removes or combines existing attributes.
func (list *AttrList) pango_attr_list_change(attr Attr) {
	if list == nil {
		return
	}

	startIndex := attr.StartIndex
	end_index := attr.EndIndex

	if startIndex == end_index {
		/* empty, nothing to do */
		return
	}

	if len(*list) == 0 {
		list.pango_attr_list_insert(attr)
		return
	}

	var i, p int
	inserted := false
	for i, p = 0, len(*list); i < p; i++ {
		tmp_attr := (*list)[i]

		if tmp_attr.StartIndex > startIndex {
			list.insert(i, attr)
			inserted = true
			break
		}

		if tmp_attr.Type != attr.Type {
			continue
		}

		if tmp_attr.EndIndex < startIndex {
			continue /* This attr does not overlap with the new one */
		}

		// tmp_attr.StartIndex <= startIndex
		// tmp_attr.EndIndex >= startIndex

		if tmp_attr.equals(attr) { // We can merge the new attribute with this attribute
			if tmp_attr.EndIndex >= end_index {
				// We are totally overlapping the previous attribute.
				// No action is needed.
				return
			}
			tmp_attr.EndIndex = end_index
			attr = tmp_attr
			inserted = true
			break
		} else { // Split, truncate, or remove the old attribute
			if tmp_attr.EndIndex > end_index {
				end_attr := tmp_attr.copy()
				end_attr.StartIndex = end_index
				list.pango_attr_list_insert(end_attr)
			}

			if tmp_attr.StartIndex == startIndex {
				list.remove(i)
				break
			} else {
				tmp_attr.EndIndex = startIndex
			}
		}
	}

	if !inserted { // we didn't insert attr yet
		list.pango_attr_list_insert(attr)
		return
	}

	/* We now have the range inserted into the list one way or the
	* other. Fix up the remainder */
	/* Attention: No i = 0 here. */
	for i, p = i+1, len(*list); i < p; i++ {
		tmp_attr := (*list)[i]

		if tmp_attr.StartIndex > end_index {
			break
		}

		if tmp_attr.Type != attr.Type {
			continue
		}

		if tmp_attr.EndIndex <= attr.EndIndex || tmp_attr.equals(attr) {
			/* We can merge the new attribute with this attribute. */
			attr.EndIndex = max(end_index, tmp_attr.EndIndex)
			list.remove(i)
			i--
			p--
			continue
		} else {
			/* Trim the start of this attribute that it begins at the end
			* of the new attribute. This may involve moving
			* it in the list to maintain the required non-decreasing
			* order of start indices
			 */
			tmp_attr.StartIndex = attr.EndIndex
			// TODO: Is the following loop missing something ?
			// for k, m := i+1, len(*list); k < m; k++ {
			// 	tmp_attr2 := (*list)[k]
			// 	if tmp_attr2.StartIndex >= tmp_attr.StartIndex {
			// 		break
			// 	}
			// }
		}
	}
}

// Checks whether `list` and `otherList` contain the same attributes and
// whether those attributes apply to the same ranges.
// Beware that this will return wrong values if any list contains duplicates.
func (list AttrList) pango_attr_list_equal(otherList AttrList) bool {
	if len(list) != len(otherList) {
		return false
	}

	var skipBitmask uint64
	for _, attr := range list {
		attrEqual := false
		for otherAttrIndex, otherAttr := range otherList {
			var otherAttrBitmask uint64
			if otherAttrIndex < 64 {
				otherAttrBitmask = 1 << otherAttrIndex
			}

			if (skipBitmask & otherAttrBitmask) != 0 {
				continue
			}

			if attr.StartIndex == otherAttr.StartIndex &&
				attr.EndIndex == otherAttr.EndIndex && attr.equals(otherAttr) {
				skipBitmask |= otherAttrBitmask
				attrEqual = true
				break
			}
		}

		if !attrEqual {
			return false
		}
	}

	return true
}
