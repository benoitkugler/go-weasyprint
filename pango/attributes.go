package pango

import (
	"fmt"
	"math"
)

const (
	ATTR_INVALID             AttrType = iota // does not happen
	ATTR_LANGUAGE                            // language (AttrLanguage)
	ATTR_FAMILY                              // font family name list (AttrString)
	ATTR_STYLE                               // font slant style (AttrInt)
	ATTR_WEIGHT                              // font weight (AttrInt)
	ATTR_VARIANT                             // font variant (normal or small caps) (AttrInt)
	ATTR_STRETCH                             // font stretch (AttrInt)
	ATTR_SIZE                                // font size in points scaled by %PANGO_SCALE (AttrInt)
	ATTR_FONT_DESC                           // font description (AttrFontDesc)
	ATTR_FOREGROUND                          // foreground color (AttrColor)
	ATTR_BACKGROUND                          // background color (AttrColor)
	ATTR_UNDERLINE                           // whether the text has an underline (AttrInt)
	ATTR_STRIKETHROUGH                       // whether the text is struck-through (AttrInt)
	ATTR_RISE                                // baseline displacement (AttrInt)
	ATTR_SHAPE                               // shape (AttrShape)
	ATTR_SCALE                               // font size scale factor (AttrFloat)
	ATTR_FALLBACK                            // whether fallback is enabled (AttrInt)
	ATTR_LETTER_SPACING                      // letter spacing (AttrInt)
	ATTR_UNDERLINE_COLOR                     // underline color (AttrColor)
	ATTR_STRIKETHROUGH_COLOR                 // strikethrough color (AttrColor)
	ATTR_ABSOLUTE_SIZE                       // font size in pixels scaled by %PANGO_SCALE (AttrInt)
	ATTR_GRAVITY                             // base text gravity (AttrInt)
	ATTR_GRAVITY_HINT                        // gravity hint (AttrInt)
	ATTR_FONT_FEATURES                       // OpenType font features (AttrString). Since 1.38
	ATTR_FOREGROUND_ALPHA                    // foreground alpha (AttrInt). Since 1.38
	ATTR_BACKGROUND_ALPHA                    // background alpha (AttrInt). Since 1.38
	ATTR_ALLOW_BREAKS                        // whether breaks are allowed (AttrInt). Since 1.44
	ATTR_SHOW                                // how to render invisible characters (AttrInt). Since 1.44
	ATTR_INSERT_HYPHENS                      // whether to insert hyphens at intra-word line breaks (AttrInt). Since 1.44
	ATTR_OVERLINE                            // whether the text has an overline (AttrInt). Since 1.46
	ATTR_OVERLINE_COLOR                      // overline color (AttrColor). Since 1.46
)

// ShowFlags affects how Pango treats characters that are normally
// not visible in the output.
type ShowFlags uint8

const (
	PANGO_SHOW_NONE        ShowFlags = 0         //  No special treatment for invisible characters
	PANGO_SHOW_SPACES      ShowFlags = 1 << iota //  Render spaces, tabs and newlines visibly
	PANGO_SHOW_LINE_BREAKS                       //  Render line breaks visibly
	PANGO_SHOW_IGNORABLES                        //  Render default-ignorable Unicode characters visibly
)

type AttrType uint8

var typeNames = [...]string{
	ATTR_INVALID:             "",
	ATTR_LANGUAGE:            "language",
	ATTR_FAMILY:              "family",
	ATTR_STYLE:               "style",
	ATTR_WEIGHT:              "weight",
	ATTR_VARIANT:             "variant",
	ATTR_STRETCH:             "stretch",
	ATTR_SIZE:                "size",
	ATTR_FONT_DESC:           "font-desc",
	ATTR_FOREGROUND:          "foreground",
	ATTR_BACKGROUND:          "background",
	ATTR_UNDERLINE:           "underline",
	ATTR_STRIKETHROUGH:       "strikethrough",
	ATTR_RISE:                "rise",
	ATTR_SHAPE:               "shape",
	ATTR_SCALE:               "scale",
	ATTR_FALLBACK:            "fallback",
	ATTR_LETTER_SPACING:      "letter-spacing",
	ATTR_UNDERLINE_COLOR:     "underline-color",
	ATTR_STRIKETHROUGH_COLOR: "strikethrough-color",
	ATTR_ABSOLUTE_SIZE:       "absolute-size",
	ATTR_GRAVITY:             "gravity",
	ATTR_GRAVITY_HINT:        "gravity-hint",
	ATTR_FONT_FEATURES:       "font-features",
	ATTR_FOREGROUND_ALPHA:    "foreground-alpha",
	ATTR_BACKGROUND_ALPHA:    "background-alpha",
	ATTR_ALLOW_BREAKS:        "allow-breaks",
	ATTR_SHOW:                "show",
	ATTR_INSERT_HYPHENS:      "insert-hyphens",
	ATTR_OVERLINE:            "overline",
	ATTR_OVERLINE_COLOR:      "overline-color",
}

func (t AttrType) String() string {
	if int(t) >= len(typeNames) {
		return "<invalid>"
	}
	return typeNames[t]
}

type AttrData interface {
	fmt.Stringer
	copy() AttrData
	equals(other AttrData) bool
}

type AttrInt int

func (a AttrInt) copy() AttrData             { return a }
func (a AttrInt) String() string             { return fmt.Sprintf("%d", a) }
func (a AttrInt) equals(other AttrData) bool { return a == other }

type AttrFloat float64

func (a AttrFloat) copy() AttrData             { return a }
func (a AttrFloat) String() string             { return fmt.Sprintf("%g", a) }
func (a AttrFloat) equals(other AttrData) bool { return a == other }

type AttrString string

func (a AttrString) copy() AttrData             { return a }
func (a AttrString) String() string             { return string(a) }
func (a AttrString) equals(other AttrData) bool { return a == other }

func (a Language) copy() AttrData             { return a }
func (a Language) String() string             { return fmt.Sprintf("%s", a[:]) }
func (a Language) equals(other AttrData) bool { return a == other }

func (a FontDescription) copy() AttrData { return a }
func (a FontDescription) equals(other AttrData) bool {
	otherDesc, ok := other.(FontDescription)
	if !ok {
		return false
	}
	return a.mask == otherDesc.mask && a.pango_font_description_equal(otherDesc)
}

// AttrColor is used to represent a color in an uncalibrated RGB color-space.
type AttrColor struct {
	red, green, blue uint16
}

func (a AttrColor) copy() AttrData             { return a }
func (a AttrColor) equals(other AttrData) bool { return a == other }

// String returns a textual specification of the color in the hexadecimal form
// '#rrrrggggbbbb', where r,g and b are hex digits representing
// the red, green, and blue components respectively.
func (a AttrColor) String() string { return fmt.Sprintf("#%04x%04x%04x", a.red, a.green, a.blue) }

type Attr struct {
	Type                 AttrType
	Data                 AttrData
	StartIndex, EndIndex uint32
}

// Initializes StartIndex to 0 and EndIndex to MaxUint
// such that the attribute applies to the entire text by default.
func (attr *Attr) pango_attribute_init() {
	attr.StartIndex = 0
	attr.EndIndex = math.MaxUint32
}

// Compare two attributes for equality. This compares only the
// actual value of the two attributes and not the ranges that the
// attributes apply to.
func (attr1 Attr) pango_attribute_equal(attr2 Attr) bool {
	return attr1.Type == attr2.Type && attr1.Data.equals(attr2.Data)
}

// Make a deep copy of an attribute.
func (a *Attr) pango_attribute_copy() *Attr {
	if a == nil {
		return a
	}
	out := *a
	out.Data = a.Data.copy()
	return &out
}

func (attr Attr) String() string {
	// to obtain the same result as the C implementation
	// we convert to int32
	return fmt.Sprintf("[%d,%d]%s=%s", int32(attr.StartIndex), int32(attr.EndIndex), attr.Type, attr.Data)
}

// Create a new font description attribute. This attribute
// allows setting family, style, weight, variant, stretch,
// and size simultaneously.
func pango_attr_font_desc_new(desc FontDescription) *Attr {
	//    PangoAttrFontDesc *result = g_slice_new (PangoAttrFontDesc);
	//    pango_attribute_init (&result.attr, &klass);
	//    result.desc = pango_font_description_copy (desc);

	out := Attr{Type: ATTR_FONT_DESC, Data: desc}
	out.pango_attribute_init()
	return &out
}

// Create a new attribute that influences how invisible
// characters are rendered.
func pango_attr_show_new(flags ShowFlags) *Attr {
	out := Attr{Type: ATTR_SHOW, Data: AttrInt(flags)}
	out.pango_attribute_init()
	return &out
}

// Create a new font-size attribute in fractional points.
func pango_attr_size_new(size int) *Attr {
	out := Attr{Type: ATTR_SIZE, Data: AttrInt(size)}
	out.pango_attribute_init()
	return &out
}

//  Create a new font weight attribute.
func pango_attr_weight_new(weight Weight) *Attr {
	out := Attr{Type: ATTR_WEIGHT, Data: AttrInt(weight)}
	out.pango_attribute_init()
	return &out
}

// Create a new font variant attribute (normal or small caps)
func pango_attr_variant_new(variant Variant) *Attr {
	out := Attr{Type: ATTR_VARIANT, Data: AttrInt(variant)}
	out.pango_attribute_init()
	return &out
}

// Create a new font-size attribute in device units.
func pango_attr_size_new_absolute(size int) *Attr {
	out := Attr{Type: ATTR_ABSOLUTE_SIZE, Data: AttrInt(size)}
	out.pango_attribute_init()
	return &out
}

// Create a new font stretch attribute
func pango_attr_stretch_new(stretch Stretch) *Attr {
	out := Attr{Type: ATTR_STRETCH, Data: AttrInt(stretch)}
	out.pango_attribute_init()
	return &out
}

// Create a new language tag attribute
func pango_attr_language_new(language Language) *Attr {
	out := Attr{Type: ATTR_LANGUAGE, Data: language}
	out.pango_attribute_init()
	return &out
}

type AttrList []*Attr

// Copy `list` and return an identical new list.
func (list AttrList) pango_attr_list_copy() AttrList {
	out := make(AttrList, len(list))
	for i, v := range list {
		out[i] = v.pango_attribute_copy()
	}
	return out
}

func (list *AttrList) pango_attr_list_insert_internal(attr *Attr, before bool) {
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

func (l *AttrList) insert(i int, attr *Attr) {
	*l = append(*l, nil)
	copy((*l)[i+1:], (*l)[i:])
	(*l)[i] = attr
}

func (l *AttrList) remove(i int) {
	copy((*l)[i:], (*l)[i+1:])
	(*l)[len((*l))-1] = nil
	(*l) = (*l)[:len((*l))-1]
}

// Insert the given attribute into the list. It will
// be inserted after all other attributes with a matching
// `StartIndex`.
func (list *AttrList) pango_attr_list_insert(attr *Attr) {
	if list == nil {
		return
	}
	list.pango_attr_list_insert_internal(attr, false)
}

// Insert the given attribute into the `AttrList`. It will
// be inserted before all other attributes with a matching
// `StartIndex`.
func (list *AttrList) pango_attr_list_insert_before(attr *Attr) {
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
	endIndex := attr.EndIndex

	if startIndex == endIndex {
		/* empty, nothing to do */
		return
	}

	if len(*list) == 0 {
		list.pango_attr_list_insert(&attr)
		return
	}

	var i, p int
	inserted := false
	for i, p = 0, len(*list); i < p; i++ {
		tmp_attr := (*list)[i]

		if tmp_attr.StartIndex > startIndex {
			list.insert(i, &attr)
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

		if tmp_attr.pango_attribute_equal(attr) { // We can merge the new attribute with this attribute
			if tmp_attr.EndIndex >= endIndex {
				// We are totally overlapping the previous attribute.
				// No action is needed.
				return
			}
			tmp_attr.EndIndex = endIndex
			attr = *tmp_attr
			inserted = true
			break
		} else { // Split, truncate, or remove the old attribute
			if tmp_attr.EndIndex > endIndex {
				end_attr := tmp_attr.pango_attribute_copy()
				end_attr.StartIndex = endIndex
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
		list.pango_attr_list_insert(&attr)
		return
	}

	/* We now have the range inserted into the list one way or the
	* other. Fix up the remainder */
	/* Attention: No i = 0 here. */
	for i, p = i+1, len(*list); i < p; i++ {
		tmp_attr := (*list)[i]

		if tmp_attr.StartIndex > endIndex {
			break
		}

		if tmp_attr.Type != attr.Type {
			continue
		}

		if tmp_attr.EndIndex <= attr.EndIndex || tmp_attr.pango_attribute_equal(attr) {
			/* We can merge the new attribute with this attribute. */
			attr.EndIndex = max(endIndex, tmp_attr.EndIndex)
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
				attr.EndIndex == otherAttr.EndIndex && attr.pango_attribute_equal(*otherAttr) {
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

// Return value: `true` if the attribute should be selected for
// filtering, `false` otherwise.
type pangoAttrFilterFunc = func(attr *Attr) bool

// Given a AttrList and callback function, removes any elements
// of `list` for which `fn` returns `true` and inserts them into
// a new list (possibly empty if no attributes of the given types were found)
func (list *AttrList) pango_attr_list_filter(fn pangoAttrFilterFunc) AttrList {
	if list == nil {
		return nil
	}
	var out AttrList
	for i, p := 0, len(*list); i < p; i++ {
		tmp_attr := (*list)[i]
		if fn(tmp_attr) {
			list.remove(i)
			i-- /* Need to look at this index again */
			p--
			out = append(out, tmp_attr)
		}
	}

	return out
}

// AttrIterator is used to represent an
// iterator through an `AttrList`. A new iterator is created
// with pango_attr_list_get_iterator(). Once the iterator
// is created, it can be advanced through the style changes
// in the text using pango_attr_iterator_next(). At each
// style change, the range of the current style segment and the
// attributes currently in effect can be queried.
type AttrIterator struct {
	attrs *AttrList

	attribute_stack AttrList

	attr_index int
	StartIndex uint32
	EndIndex   uint32
}

// pango_attr_list_get_iterator creates a iterator initialized to the beginning of the list.
// `list` must not be modified until this iterator is freed.
func (list *AttrList) pango_attr_list_get_iterator() *AttrIterator {
	if list == nil {
		return nil
	}
	iterator := AttrIterator{attrs: list}

	if !iterator.pango_attr_iterator_next() {
		iterator.EndIndex = math.MaxUint32
	}

	return &iterator
}

// pango_attr_iterator_next advances the iterator until the next change of style, and
// returns `false` if the iterator is at the end of the list, otherwise `true`
func (iterator *AttrIterator) pango_attr_iterator_next() bool {
	if iterator == nil {
		return false
	}

	if iterator.attr_index >= len(*iterator.attrs) && len(iterator.attribute_stack) == 0 {
		return false
	}
	iterator.StartIndex = iterator.EndIndex
	iterator.EndIndex = math.MaxUint32

	for i := len(iterator.attribute_stack) - 1; i >= 0; i-- {
		attr := iterator.attribute_stack[i]
		if attr.EndIndex == iterator.StartIndex {
			iterator.attribute_stack.remove(i)
		} else {
			iterator.EndIndex = min(iterator.EndIndex, attr.EndIndex)
		}
	}

	for {
		if iterator.attr_index >= len(*iterator.attrs) {
			break
		}
		attr := (*iterator.attrs)[iterator.attr_index]

		if attr.StartIndex != iterator.StartIndex {
			break
		}

		if attr.EndIndex > iterator.StartIndex {
			iterator.attribute_stack = append(iterator.attribute_stack, attr)
			iterator.EndIndex = min(iterator.EndIndex, attr.EndIndex)
		}

		iterator.attr_index++ /* NEXT! */
	}

	if iterator.attr_index < len(*iterator.attrs) {
		attr := (*iterator.attrs)[iterator.attr_index]
		iterator.EndIndex = min(iterator.EndIndex, attr.StartIndex)
	}

	return true
}

// pango_attr_iterator_get_attrs gets a list of all attributes at the current position of the
// iterator.
func (iterator AttrIterator) pango_attr_iterator_get_attrs() AttrList {
	var attrs AttrList

	for i := len(iterator.attribute_stack) - 1; i >= 0; i-- {
		attr := iterator.attribute_stack[i]
		found := false
		for _, old_attr := range attrs {
			if attr.Type == old_attr.Type {
				found = true
				break
			}
		}
		if !found {
			attrs = append(AttrList{attr}, attrs...)
		}
	}

	return attrs
}

/**
 * pango_attr_iterator_get:
 * @iterator: a #PangoAttrIterator
 * @type: the type of attribute to find.
 *
 * Find the current attribute of a particular type at the iterator
 * location. When multiple attributes of the same type overlap,
 * the attribute whose range starts closest to the current location
 * is used.
 *
 * Return value: (nullable) (transfer none): the current attribute of the given type,
 *               or %nil if no attribute of that type applies to the
 *               current location.
 **/
func (iterator AttrIterator) pango_attr_iterator_get(type_ AttrType) *Attr {
	if len(iterator.attribute_stack) == 0 {
		return nil

	}

	for i := len(iterator.attribute_stack) - 1; i >= 0; i-- {
		attr := iterator.attribute_stack[i]

		if attr.Type == type_ {
			return attr
		}
	}

	return nil
}

// @extra_attrs: (allow-none) (element-type Pango.Attribute) (transfer full): if non-nil,
//           location in which to store a list of non-font
//           attributes at the the current position; only the highest priority
//           value of each attribute will be added to this list. In order
//           to free this value, you must call pango_attribute_destroy() on
//           each member.
//
// pango_attr_iterator_get_font gets the font and other attributes at the current iterator position.
// `desc` is a FontDescription to fill in with the current values.
// If non-nil, `language` is a location to store language tag for item, or zero if none is found.
// TODO: support extra_attrs GSList               **extra_attrs
func (iterator AttrIterator) pango_attr_iterator_get_font(desc *FontDescription, language *Language) {
	if desc == nil {
		return
	}
	//    int i;

	if language != nil {
		*language = [2]byte{}
	}

	// if extra_attrs {
	// 	*extra_attrs = nil
	// }

	if len(iterator.attribute_stack) == 0 {
		return
	}

	var (
		mask                    FontMask
		haveScale, haveLanguage bool
		scale                   AttrFloat
	)
	for i := len(iterator.attribute_stack) - 1; i >= 0; i-- {
		attr := iterator.attribute_stack[i]

		switch attr.Type {
		case ATTR_FONT_DESC:
			attrDesc := attr.Data.(FontDescription)
			new_mask := attrDesc.mask & ^mask
			mask |= new_mask
			desc.pango_font_description_unset_fields(new_mask)
			desc.pango_font_description_merge(&attrDesc, false)
		case ATTR_FAMILY:
			if mask&PANGO_FONT_MASK_FAMILY == 0 {
				mask |= PANGO_FONT_MASK_FAMILY
				desc.pango_font_description_set_family(string(attr.Data.(AttrString)))
			}
		case ATTR_STYLE:
			if mask&PANGO_FONT_MASK_STYLE == 0 {
				mask |= PANGO_FONT_MASK_STYLE
				desc.pango_font_description_set_style(Style(attr.Data.(AttrInt)))
			}
		case ATTR_VARIANT:
			if mask&PANGO_FONT_MASK_VARIANT == 0 {
				mask |= PANGO_FONT_MASK_VARIANT
				desc.pango_font_description_set_variant(Variant(attr.Data.(AttrInt)))
			}
		case ATTR_WEIGHT:
			if mask&PANGO_FONT_MASK_WEIGHT == 0 {
				mask |= PANGO_FONT_MASK_WEIGHT
				desc.pango_font_description_set_weight(Weight(attr.Data.(AttrInt)))
			}
		case ATTR_STRETCH:
			if mask&PANGO_FONT_MASK_STRETCH == 0 {
				mask |= PANGO_FONT_MASK_STRETCH
				desc.pango_font_description_set_stretch(Stretch(attr.Data.(AttrInt)))
			}
		case ATTR_SIZE:
			if mask&PANGO_FONT_MASK_SIZE == 0 {
				mask |= PANGO_FONT_MASK_SIZE
				desc.pango_font_description_set_size(int(attr.Data.(AttrInt)))
			}
		case ATTR_ABSOLUTE_SIZE:
			if mask&PANGO_FONT_MASK_SIZE == 0 {
				mask |= PANGO_FONT_MASK_SIZE
				desc.pango_font_description_set_absolute_size(int(attr.Data.(AttrInt)))
			}
		case ATTR_SCALE:
			if !haveScale {
				haveScale = true
				scale = attr.Data.(AttrFloat)
			}
		case ATTR_LANGUAGE:
			if language != nil {
				if !haveLanguage {
					haveLanguage = true
					*language = attr.Data.(Language)
				}
			}
		default:
			//    if (extra_attrs) {
			// 	   gboolean found = false;

			// 	   /* Hack: special-case FONT_FEATURES.  We don't want them to
			// 		* override each other, so we never merge them.  This should
			// 		* be fixed when we implement attr-merging. */
			// 	   if (attr.klass.type != ATTR_FONT_FEATURES)  {
			// 			   GSList *tmp_list = *extra_attrs;
			// 			   for (tmp_list)  {
			// 				   PangoAttribute *old_attr = tmp_list.data;
			// 				   if (attr.klass.type == old_attr.klass.type)  {
			// 					   found = true;
			// 					   break;
			// 					 }

			// 				   tmp_list = tmp_list.next;
			// 				 }
			// 			 }

			// 	   if (!found){
			// 	 *extra_attrs = g_slist_prepend (*extra_attrs, pango_attribute_copy (attr));}
			// 	 }
		}
	}

	if haveScale {
		if desc.size_is_absolute {
			desc.pango_font_description_set_absolute_size(int(scale * AttrFloat(desc.size)))
		} else {
			desc.pango_font_description_set_size(int(scale * AttrFloat(desc.size)))
		}
	}
}
