package pango

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

/**
 * Simple markup language for text with attributes
 *
 * Frequently, you want to display some text to the user with attributes
 * applied to part of the text (for example, you might want bold or
 * italicized words). With the base Pango interfaces, you could create a
 * `AttrList` and apply it to the text; the problem is that you'd
 * need to apply attributes to some numeric range of characters, for
 * example "characters 12-17." This is broken from an internationalization
 * standpoint; once the text is translated, the word you wanted to
 * italicize could be in a different position.
 *
 * The solution is to include the text attributes in the string to be
 * translated. Pango provides this feature with a small markup language.
 * You can parse a marked-up string into the string text plus a
 * `AttrList` using either of pango_parse_markup() or
 * pango_markup_parser_new().
 *
 * A simple example of a marked-up string might be:
 *
 * <span foreground="blue" size="x-large">Blue text</span> is <i>cool</i>!
 *
 *
 * Pango uses #GMarkup to parse this language, which means that XML
 * features such as numeric character entities such as `&#169;` for
 * © can be used too.
 *
 * The root tag of a marked-up document is `<markup>`, but
 * pango_parse_markup() allows you to omit this tag, so you will most
 * likely never need to use it. The most general markup tag is `<span>`,
 * then there are some convenience tags.
 *
 * ## Span attributes
 *
 * `<span>` has the following attributes:
 *
 * * `font_desc`:
 *   A font description string, such as "Sans Italic 12".
 *   See pango_font_description_from_string() for a description of the
 *   format of the string representation . Note that any other span
 *   attributes will override this description. So if you have "Sans Italic"
 *   and also a `style="normal"` attribute, you will get Sans normal,
 *   not italic.
 *
 * * `font_family`:
 *   A font family name
 *
 * * `font_size`, `size`:
 *   Font size in 1024ths of a point, or one of the absolute
 *   sizes `xx-small`, `x-small`, `small`, `medium`, `large`,
 *   `x-large`, `xx-large`, or one of the relative sizes `smaller`
 *   or `larger`. If you want to specify a absolute size, it's usually
 *   easier to take advantage of the ability to specify a partial
 *   font description using `font`; you can use `font='12.5'`
 *   rather than `size='12800'`.
 *
 * * `font_style`:
 *   One of `normal`, `oblique`, `italic`
 *
 * * `font_weight`:
 *   One of `ultralight`, `light`, `normal`, `bold`,
 *   `ultrabold`, `heavy`, or a numeric weight
 *
 * * `font_variant`:
 *   One of `normal` or `smallcaps`
 *
 * * `font_stretch`, `stretch`:
 *   One of `ultracondensed`, `extracondensed`, `condensed`,
 *   `semicondensed`, `normal`, `semiexpanded`, `expanded`,
 *   `extraexpanded`, `ultraexpanded`
 *
 * * `font_features`:
 *   A comma-separated list of OpenType font feature
 *   settings, in the same syntax as accepted by CSS. E.g:
 *   `font_features='dlig=1, -kern, afrc on'`
 *
 * * `foreground`, `fgcolor`:
 *   An RGB color specification such as `#00FF00` or a color
 *   name such as `red`. Since 1.38, an RGBA color specification such
 *   as `#00FF007F` will be interpreted as specifying both a foreground
 *   color and foreground alpha.
 *
 * * `background`, `bgcolor`:
 *   An RGB color specification such as `#00FF00` or a color
 *   name such as `red`.
 *   Since 1.38, an RGBA color specification such as `#00FF007F` will
 *   be interpreted as specifying both a background color and
 *   background alpha.
 *
 * * `alpha`, `fgalpha`:
 *   An alpha value for the foreground color, either a plain
 *   integer between 1 and 65536 or a percentage value like `50%`.
 *
 * * `background_alpha`, `bgalpha`:
 *   An alpha value for the background color, either a plain
 *   integer between 1 and 65536 or a percentage value like `50%`.
 *
 * * `underline`:
 *   One of `none`, `single`, `float64`, `low`, `error`,
 *   `single-line`, `float64-line` or `error-line`.
 *
 * * `underline_color`:
 *   The color of underlines; an RGB color
 *   specification such as `#00FF00` or a color name such as `red`
 *
 * * `overline`:
 *   One of `none` or `single`
 *
 * * `overline_color`:
 *   The color of overlines; an RGB color
 *   specification such as `#00FF00` or a color name such as `red`
 *
 * * `rise`:
 *   Vertical displacement, in Pango units. Can be negative for
 *   subscript, positive for superscript.
 *
 * * `strikethrough`
 *   `true` or `false` whether to strike through the text
 *
 * * `strikethrough_color`:
 *   The color of strikethrough lines; an RGB
 *   color specification such as `#00FF00` or a color name such as `red`
 *
 * * `fallback`:
 *   `true` or `false` whether to enable fallback. If
 *   disabled, then characters will only be used from the closest
 *   matching font on the system. No fallback will be done to other
 *   fonts on the system that might contain the characters in the text.
 *   Fallback is enabled by default. Most applications should not
 *   disable fallback.
 *
 * * `allow_breaks`:
 *   `true` or `false` whether to allow line breaks or not. If
 *   not allowed, the range will be kept in a single run as far
 *   as possible. Breaks are allowed by default.
 *
 * * `insert_hyphens`:`
 *   `true` or `false` whether to insert hyphens when breaking
 *   lines in the middle of a word. Hyphens are inserted by default.
 *
 * * `show`:
 *   A value determining how invisible characters are treated.
 *   Possible values are `spaces`, `line-breaks`, `ignorables`
 *   or combinations, such as `spaces|line-breaks`.
 *
 * * `lang`:
 *   A language code, indicating the text language
 *
 * * `letter_spacing`:
 *   Inter-letter spacing in 1024ths of a point.
 *
 * * `gravity`:
 *   One of `south`, `east`, `north`, `west`, `auto`.
 *
 * * `gravity_hint`:
 *   One of `natural`, `strong`, `line`.
 *
 * ## Convenience tags
 *
 * The following convenience tags are provided:
 *
 * * `<b>`:
 *   Bold
 *
 * * `<big>`:
 *   Makes font relatively larger, equivalent to `<span size="larger">`
 *
 * * `<i>`:
 *   Italic
 *
 * * `<s>`:
 *   Strikethrough
 *
 * * `<sub>`:
 *   Subscript
 *
 * * `<sup>`:
 *   Superscript
 *
 * * `<small>`:
 *   Makes font relatively smaller, equivalent to `<span size="smaller">`
 *
 * * `<tt>`:
 *   Monospace
 *
 * * `<u>`:
 *   Underline
 */

//  /* FIXME */
//  #define _(x) x

/* CSS size levels */
type SizeLevel int8

const (
	XXSmall SizeLevel = -3
	XSmall  SizeLevel = -2
	Small   SizeLevel = -1
	Medium  SizeLevel = 0
	Large   SizeLevel = 1
	XLarge  SizeLevel = 2
	XXLarge SizeLevel = 3
)

func (scaleLevel SizeLevel) scale_factor(base float64) float64 {
	factor := base

	// 1.2 is the CSS scale factor between sizes

	if scaleLevel > 0 {
		for i := SizeLevel(0); i < scaleLevel; i++ {
			factor *= 1.2
		}
	} else if scaleLevel < 0 {
		for i := scaleLevel; i < 0; i++ {
			factor /= 1.2
		}
	}

	return factor
}

type MarkupData struct {
	attr_list AttrList
	//    GString *text;
	//    GSList *tag_stack;
	//    gsize index;
	//    GSList *to_apply;
	accel_marker rune
	accel_char   rune
}

type OpenTag struct {
	attrs       AttrList
	start_index int
	/* Current total scale level; reset whenever
	* an absolute size is set.
	* Each "larger" ups it 1, each "smaller" decrements it 1
	 */
	scale_level int
	/* Our impact on scale_level, so we know whether we
	* need to create an attribute ourselves on close
	 */
	scale_level_delta int
	/* Base scale factor currently in effect
	* or size that this tag
	* forces, or parent's scale factor or size.
	 */
	base_scale_factor  float64
	base_font_size     int
	has_base_font_size bool // = 1;
}

func (ot *OpenTag) add_attribute(attr *Attribute) {
	if ot == nil {
		return
	}
	ot.attrs.insert(0, attr)
}

func (ot *OpenTag) open_tag_set_absolute_font_scale(scale float64) {
	ot.base_scale_factor = scale
	ot.has_base_font_size = false
	ot.scale_level = 0
	ot.scale_level_delta = 0
}

//  typedef gboolean (*TagParseFunc) (MarkupData            *md,
// 				   OpenTag               *tag,
// 				   const gchar          **names,
// 				   const gchar          **values,
// 				   GMarkupParseContext   *context,
// 				   GError               **error);

//  static gboolean b_parse_func        (MarkupData           *md,
// 					  OpenTag              *tag,
// 					  const gchar         **names,
// 					  const gchar         **values,
// 					  GMarkupParseContext  *context,
// 					  GError              **error);
//  static gboolean big_parse_func      (MarkupData           *md,
// 					  OpenTag              *tag,
// 					  const gchar         **names,
// 					  const gchar         **values,
// 					  GMarkupParseContext  *context,
// 					  GError              **error);
//  static gboolean span_parse_func     (MarkupData           *md,
// 					  OpenTag              *tag,
// 					  const gchar         **names,
// 					  const gchar         **values,
// 					  GMarkupParseContext  *context,
// 					  GError              **error);
//  static gboolean i_parse_func        (MarkupData           *md,
// 					  OpenTag              *tag,
// 					  const gchar         **names,
// 					  const gchar         **values,
// 					  GMarkupParseContext  *context,
// 					  GError              **error);
//  static gboolean markup_parse_func   (MarkupData           *md,
// 					  OpenTag              *tag,
// 					  const gchar         **names,
// 					  const gchar         **values,
// 					  GMarkupParseContext  *context,
// 					  GError              **error);
//  static gboolean s_parse_func        (MarkupData           *md,
// 					  OpenTag              *tag,
// 					  const gchar         **names,
// 					  const gchar         **values,
// 					  GMarkupParseContext  *context,
// 					  GError              **error);
//  static gboolean sub_parse_func      (MarkupData           *md,
// 					  OpenTag              *tag,
// 					  const gchar         **names,
// 					  const gchar         **values,
// 					  GMarkupParseContext  *context,
// 					  GError              **error);
//  static gboolean sup_parse_func      (MarkupData           *md,
// 					  OpenTag              *tag,
// 					  const gchar         **names,
// 					  const gchar         **values,
// 					  GMarkupParseContext  *context,
// 					  GError              **error);
//  static gboolean small_parse_func    (MarkupData           *md,
// 					  OpenTag              *tag,
// 					  const gchar         **names,
// 					  const gchar         **values,
// 					  GMarkupParseContext  *context,
// 					  GError              **error);
//  static gboolean tt_parse_func       (MarkupData           *md,
// 					  OpenTag              *tag,
// 					  const gchar         **names,
// 					  const gchar         **values,
// 					  GMarkupParseContext  *context,
// 					  GError              **error);
//  static gboolean u_parse_func        (MarkupData           *md,
// 					  OpenTag              *tag,
// 					  const gchar         **names,
// 					  const gchar         **values,
// 					  GMarkupParseContext  *context,
// 					  GError              **error);

//  static void
//  open_tag_free (OpenTag *ot)
//  {
//    g_slist_foreach (ot.attrs, (GFunc) pango_attribute_destroy, NULL);
//    g_slist_free (ot.attrs);
//    g_slice_free (OpenTag, ot);
//  }

//  static void
//  open_tag_set_absolute_font_size (OpenTag *ot,
// 				  int      font_size)
//  {
//    ot.base_font_size = font_size;
//    ot.has_base_font_size = true;
//    ot.scale_level = 0;
//    ot.scale_level_delta = 0;
//  }

//  static OpenTag*
//  markup_data_open_tag (MarkupData   *md)
//  {
//    OpenTag *ot;
//    OpenTag *parent = NULL;

//    if (md.attr_list == NULL)
// 	 return NULL;

//    if (md.tag_stack)
// 	 parent = md.tag_stack.data;

//    ot = g_slice_new (OpenTag);
//    ot.attrs = NULL;
//    ot.start_index = md.index;
//    ot.scale_level_delta = 0;

//    if (parent == NULL)
// 	 {
// 	   ot.base_scale_factor = 1.0;
// 	   ot.base_font_size = 0;
// 	   ot.has_base_font_size = false;
// 	   ot.scale_level = 0;
// 	 }
//    else
// 	 {
// 	   ot.base_scale_factor = parent.base_scale_factor;
// 	   ot.base_font_size = parent.base_font_size;
// 	   ot.has_base_font_size = parent.has_base_font_size;
// 	   ot.scale_level = parent.scale_level;
// 	 }

//    md.tag_stack = g_slist_prepend (md.tag_stack, ot);

//    return ot;
//  }

//  static void
//  markup_data_close_tag (MarkupData *md)
//  {
//    OpenTag *ot;
//    GSList *tmp_list;

//    if (md.attr_list == NULL)
// 	 return;

//    /* pop the stack */
//    ot = md.tag_stack.data;
//    md.tag_stack = g_slist_delete_link (md.tag_stack,
// 						md.tag_stack);

//    /* Adjust end indexes, and push each attr onto the front of the
// 	* to_apply list. This means that outermost tags are on the front of
// 	* that list; if we apply the list in order, then the innermost
// 	* tags will "win" which is correct.
// 	*/
//    tmp_list = ot.attrs;
//    for (tmp_list != NULL)
// 	 {
// 	   PangoAttribute *a = tmp_list.data;

// 	   a.start_index = ot.start_index;
// 	   a.end_index = md.index;

// 	   md.to_apply = g_slist_prepend (md.to_apply, a);

// 	   tmp_list = g_slist_next (tmp_list);
// 	 }

//    if (ot.scale_level_delta != 0)
// 	 {
// 	   /* We affected relative font size; create an appropriate
// 		* attribute and reverse our effects on the current level
// 		*/
// 	   PangoAttribute *a;

// 	   if (ot.has_base_font_size)
// 	 {
// 	   /* Create a font using the absolute point size
// 		* as the base size to be scaled from
// 		*/
// 	   a = pango_attr_size_new (scale_factor (ot.scale_level,
// 						  1.0) *
// 					ot.base_font_size);
// 	 }
// 	   else
// 	 {
// 	   /* Create a font using the current scale factor
// 		* as the base size to be scaled from
// 		*/
// 	   a = pango_attr_scale_new (scale_factor (ot.scale_level,
// 						   ot.base_scale_factor));
// 	 }

// 	   a.start_index = ot.start_index;
// 	   a.end_index = md.index;

// 	   md.to_apply = g_slist_prepend (md.to_apply, a);
// 	 }

//    g_slist_free (ot.attrs);
//    g_slice_free (OpenTag, ot);
//  }

//  static void
//  start_element_handler  (GMarkupParseContext *context,
// 			 const gchar         *element_name,
// 			 const gchar        **attribute_names,
// 			 const gchar        **attribute_values,
// 			 gpointer             user_data,
// 			 GError             **error)
//  {
//    TagParseFunc parse_func = NULL;
//    OpenTag *ot;

//    switch (*element_name)
// 	 {
// 	 case 'b':
// 	   if (strcmp ("b", element_name) == 0)
// 	 parse_func = b_parse_func;
// 	   else if (strcmp ("big", element_name) == 0)
// 	 parse_func = big_parse_func;
// 	   break;

// 	 case 'i':
// 	   if (strcmp ("i", element_name) == 0)
// 	 parse_func = i_parse_func;
// 	   break;

// 	 case 'm':
// 	   if (strcmp ("markup", element_name) == 0)
// 	 parse_func = markup_parse_func;
// 	   break;

// 	 case 's':
// 	   if (strcmp ("span", element_name) == 0)
// 	 parse_func = span_parse_func;
// 	   else if (strcmp ("s", element_name) == 0)
// 	 parse_func = s_parse_func;
// 	   else if (strcmp ("sub", element_name) == 0)
// 	 parse_func = sub_parse_func;
// 	   else if (strcmp ("sup", element_name) == 0)
// 	 parse_func = sup_parse_func;
// 	   else if (strcmp ("small", element_name) == 0)
// 	 parse_func = small_parse_func;
// 	   break;

// 	 case 't':
// 	   if (strcmp ("tt", element_name) == 0)
// 	 parse_func = tt_parse_func;
// 	   break;

// 	 case 'u':
// 	   if (strcmp ("u", element_name) == 0)
// 	 parse_func = u_parse_func;
// 	   break;
// 	 }

//    if (parse_func == NULL)
// 	 {
// 	   gint line_number, char_number;

// 	   g_markup_parse_context_get_position (context,
// 						&line_number, &char_number);

// 	   g_set_error (error,
// 			G_MARKUP_ERROR,
// 			G_MARKUP_ERROR_UNKNOWN_ELEMENT,
// 			_("Unknown tag '%s' on line %d char %d"),
// 			element_name,
// 			line_number, char_number);

// 	   return;
// 	 }

//    ot = markup_data_open_tag (user_data);

//    /* note ot may be NULL if the user didn't want the attribute list */

//    if (!(*parse_func) (user_data, ot,
// 			   attribute_names, attribute_values,
// 			   context, error))
// 	 {
// 	   /* there's nothing to do; we return an error, and end up
// 		* freeing ot off the tag stack later.
// 		*/
// 	 }
//  }

//  static void
//  end_element_handler    (GMarkupParseContext *context G_GNUC_UNUSED,
// 			 const gchar         *element_name G_GNUC_UNUSED,
// 			 gpointer             user_data,
// 			 GError             **error G_GNUC_UNUSED)
//  {
//    markup_data_close_tag (user_data);
//  }

//  static void
//  text_handler           (GMarkupParseContext *context G_GNUC_UNUSED,
// 			 const gchar         *text,
// 			 gsize                text_len,
// 			 gpointer             user_data,
// 			 GError             **error G_GNUC_UNUSED)
//  {
//    MarkupData *md = user_data;

//    if (md.accel_marker == 0)
// 	 {
// 	   /* Just append all the text */

// 	   md.index += text_len;

// 	   g_string_append_len (md.text, text, text_len);
// 	 }
//    else
// 	 {
// 	   /* Parse the accelerator */
// 	   const gchar *p;
// 	   const gchar *end;
// 	   const gchar *range_start;
// 	   const gchar *range_end;
// 	   gssize uline_index = -1;
// 	   gsize uline_len = 0;	/* Quiet GCC */

// 	   range_end = NULL;
// 	   range_start = text;
// 	   p = text;
// 	   end = text + text_len;

// 	   for (p != end)
// 	 {
// 	   gunichar c;

// 	   c = g_utf8_get_char (p);

// 	   if (range_end)
// 		 {
// 		   if (c == md.accel_marker)
// 		 {
// 		   /* escaped accel marker; move range_end
// 			* past the accel marker that came before,
// 			* append the whole thing
// 			*/
// 		   range_end = g_utf8_next_char (range_end);
// 		   g_string_append_len (md.text,
// 						range_start,
// 						range_end - range_start);
// 		   md.index += range_end - range_start;

// 		   /* set next range_start, skipping accel marker */
// 		   range_start = g_utf8_next_char (p);
// 		 }
// 		   else
// 		 {
// 		   /* Don't append the accel marker (leave range_end
// 			* alone); set the accel char to c; record location for
// 			* underline attribute
// 			*/
// 		   if (md.accel_char == 0)
// 			 md.accel_char = c;

// 		   g_string_append_len (md.text,
// 						range_start,
// 						range_end - range_start);
// 		   md.index += range_end - range_start;

// 		   /* The underline should go underneath the char
// 			* we're setting as the next range_start
// 			*/
// 		   uline_index = md.index;
// 		   uline_len = g_utf8_next_char (p) - p;

// 		   /* set next range_start to include this char */
// 		   range_start = p;
// 		 }

// 		   /* reset range_end */
// 		   range_end = NULL;
// 		 }
// 	   else if (c == md.accel_marker)
// 		 {
// 		   range_end = p;
// 		 }

// 	   p = g_utf8_next_char (p);
// 	 }

// 	   if (range_end)
// 	 {
// 	   g_string_append_len (md.text,
// 					range_start,
// 					range_end - range_start);
// 	   md.index += range_end - range_start;
// 	 }
// 	   else
// 	 {
// 	   g_string_append_len (md.text,
// 					range_start,
// 					end - range_start);
// 	   md.index += end - range_start;
// 	 }

// 	   if (md.attr_list != NULL && uline_index >= 0)
// 	 {
// 	   /* Add the underline indicating the accelerator */
// 	   PangoAttribute *attr;

// 	   attr = pango_attr_underline_new (PANGO_UNDERLINE_LOW);

// 	   attr.start_index = uline_index;
// 	   attr.end_index = uline_index + uline_len;

// 	   pango_attr_list_change (md.attr_list, attr);
// 	 }
// 	 }
//  }

// func xml_isspace (char c)  {
//    return c == ' ' || c == '\t' || c == '\n' || c == '\r';
//  }

//  static const GMarkupParser pango_markup_parser = {
//    start_element_handler,
//    end_element_handler,
//    text_handler,
//    NULL,
//    NULL
//  };

//  static void
//  destroy_markup_data (MarkupData *md)
//  {
//    g_slist_free_full (md.tag_stack, (GDestroyNotify) open_tag_free);
//    g_slist_free_full (md.to_apply, (GDestroyNotify) pango_attribute_destroy);
//    if (md.text)
// 	   g_string_free (md.text, true);

//    if (md.attr_list)
// 	 pango_attr_list_unref (md.attr_list);

//    g_slice_free (MarkupData, md);
//  }

// func pango_markup_parser_new_internal(accel_marker rune, want_attr_list bool) {
// 	//    GMarkupParseContext *context;

// 	md := &MarkupData{}

// 	md.accel_marker = accel_marker

// 	context = g_markup_parse_context_new(&pango_markup_parser, 0, md, destroy_markup_data)

// 	if !g_markup_parse_context_parse(context,
// 		"<markup>",
// 		-1,
// 		error) {
// 		goto error
// 	}

// 	return context

// error:
// 	g_markup_parse_context_free(context)
// 	return NULL
// }

//  /**
//   * pango_parse_markup:
//   * @markup_text: markup to parse (see <link linkend="PangoMarkupFormat">markup format</link>)
//   * @length: length of @markup_text, or -1 if nul-terminated
//   * @accel_marker: character that precedes an accelerator, or 0 for none
//   * @attr_list: (out) (allow-none): address of return location for a `AttrList`, or %NULL
//   * @text: (out) (allow-none): address of return location for text with tags stripped, or %NULL
//   * @accel_char: (out) (allow-none): address of return location for accelerator char, or %NULL
//   * @error: address of return location for errors, or %NULL
//   *
//   * Parses marked-up text (see
//   * <link linkend="PangoMarkupFormat">markup format</link>) to create
//   * a plain-text string and an attribute list.
//   *
//   * If @accel_marker is nonzero, the given character will mark the
//   * character following it as an accelerator. For example, @accel_marker
//   * might be an ampersand or underscore. All characters marked
//   * as an accelerator will receive a %PANGO_UNDERLINE_LOW attribute,
//   * and the first character so marked will be returned in @accel_char.
//   * Two @accel_marker characters following each other produce a single
//   * literal @accel_marker character.
//   *
//   * To parse a stream of pango markup incrementally, use pango_markup_parser_new().
//   *
//   * If any error happens, none of the output arguments are touched except
//   * for @error.
//   *
//   * Return value: %false if @error is set, otherwise %true
//   **/
//  gboolean
//  pango_parse_markup (const char                 *markup_text,
// 			 int                         length,
// 			 gunichar                    accel_marker,
// 			 PangoAttrList             **attr_list,
// 			 char                      **text,
// 			 gunichar                   *accel_char,
// 			 GError                    **error)
//  {
//    GMarkupParseContext *context = NULL;
//    gboolean ret = false;
//    const char *p;
//    const char *end;

//    g_return_val_if_fail (markup_text != NULL, false);

//    if (length < 0)
// 	 length = strlen (markup_text);

//    p = markup_text;
//    end = markup_text + length;
//    for (p != end && xml_isspace (*p))
// 	 ++p;

//    context = pango_markup_parser_new_internal (accel_marker,
// 											   error,
// 											   (attr_list != NULL));
//    if (context == NULL)
// 	 goto out;

//    if (!g_markup_parse_context_parse (context,
// 									  markup_text,
// 									  length,
// 									  error))
// 	 goto out;

//    if (!pango_markup_parser_finish (context,
// 									attr_list,
// 									text,
// 									accel_char,
// 									error))
// 	 goto out;

//    ret = true;

//   out:
//    if (context != NULL)
// 	 g_markup_parse_context_free (context);
//    return ret;
//  }

//  /**
//   * pango_markup_parser_new:
//   * @accel_marker: character that precedes an accelerator, or 0 for none
//   *
//   * Parses marked-up text (see
//   * <link linkend="PangoMarkupFormat">markup format</link>) to create
//   * a plain-text string and an attribute list.
//   *
//   * If @accel_marker is nonzero, the given character will mark the
//   * character following it as an accelerator. For example, @accel_marker
//   * might be an ampersand or underscore. All characters marked
//   * as an accelerator will receive a %PANGO_UNDERLINE_LOW attribute,
//   * and the first character so marked will be returned in @accel_char,
//   * when calling finish(). Two @accel_marker characters following each
//   * other produce a single literal @accel_marker character.
//   *
//   * To feed markup to the parser, use g_markup_parse_context_parse()
//   * on the returned #GMarkupParseContext. When done with feeding markup
//   * to the parser, use pango_markup_parser_finish() to get the data out
//   * of it, and then use g_markup_parse_context_free() to free it.
//   *
//   * This function is designed for applications that read pango markup
//   * from streams. To simply parse a string containing pango markup,
//   * the simpler pango_parse_markup() API is recommended instead.
//   *
//   * Return value: (transfer none): a #GMarkupParseContext that should be
//   * destroyed with g_markup_parse_context_free().
//   *
//   * Since: 1.31.0
//   **/
//  GMarkupParseContext *
//  pango_markup_parser_new (gunichar               accel_marker)
//  {
//    GError *error = NULL;
//    GMarkupParseContext *context;
//    context = pango_markup_parser_new_internal (accel_marker, &error, true);

//    if (context == NULL)
// 	 g_critical ("Had error when making markup parser: %s\n", error.message);

//    return context;
//  }

//  /**
//   * pango_markup_parser_finish:
//   * @context: A valid parse context that was returned from pango_markup_parser_new()
//   * @attr_list: (out) (allow-none): address of return location for a `AttrList`, or %NULL
//   * @text: (out) (allow-none): address of return location for text with tags stripped, or %NULL
//   * @accel_char: (out) (allow-none): address of return location for accelerator char, or %NULL
//   * @error: address of return location for errors, or %NULL
//   *
//   * After feeding a pango markup parser some data with g_markup_parse_context_parse(),
//   * use this function to get the list of pango attributes and text out of the
//   * markup. This function will not free @context, use g_markup_parse_context_free()
//   * to do so.
//   *
//   * Return value: %false if @error is set, otherwise %true
//   *
//   * Since: 1.31.0
//   */
//  gboolean
//  pango_markup_parser_finish (GMarkupParseContext   *context,
// 							 PangoAttrList        **attr_list,
// 							 char                 **text,
// 							 gunichar              *accel_char,
// 							 GError               **error)
//  {
//    gboolean ret = false;
//    MarkupData *md = g_markup_parse_context_get_user_data (context);
//    GSList *tmp_list;

//    if (!g_markup_parse_context_parse (context,
// 									  "</markup>",
// 									  -1,
// 									  error))
// 	 goto out;

//    if (!g_markup_parse_context_end_parse (context, error))
// 	 goto out;

//    if (md.attr_list)
// 	 {
// 	   /* The apply list has the most-recently-closed tags first;
// 		* we want to apply the least-recently-closed tag last.
// 		*/
// 	   tmp_list = md.to_apply;
// 	   for (tmp_list != NULL)
// 	 {
// 	   PangoAttribute *attr = tmp_list.data;

// 	   /* Innermost tags before outermost */
// 	   pango_attr_list_insert (md.attr_list, attr);

// 	   tmp_list = g_slist_next (tmp_list);
// 	 }
// 	   g_slist_free (md.to_apply);
// 	   md.to_apply = NULL;
// 	 }

//    if (attr_list)
// 	 {
// 	   *attr_list = md.attr_list;
// 	   md.attr_list = NULL;
// 	 }

//    if (text)
// 	 {
// 	   *text = g_string_free (md.text, true);
// 	   md.text = NULL;
// 	 }

//    if (accel_char)
// 	 *accel_char = md.accel_char;

//    g_assert (md.tag_stack == NULL);
//    ret = true;

//   out:
//    return ret;
//  }

func set_bad_attribute(element_name, attribute_name string) error {
	return fmt.Errorf("tag '%s' does not support attribute '%s'", element_name, attribute_name)
}

func CHECK_NO_ATTRS(elem string, names []xml.Attr) error {
	if names[0].Name.Local != "" {
		return set_bad_attribute(elem, names[0].Name.Local)
	}
	return nil
}

func b_parse_func(tag *OpenTag, names []xml.Attr) error {
	if err := CHECK_NO_ATTRS("b", names); err != nil {
		return err
	}
	tag.add_attribute(pango_attr_weight_new(PANGO_WEIGHT_BOLD))
	return nil
}

func big_parse_func(tag *OpenTag, names []xml.Attr) error {
	if err := CHECK_NO_ATTRS("big", names); err != nil {
		return err
	}
	/* Grow text one level */
	if tag != nil {
		tag.scale_level_delta += 1
		tag.scale_level += 1
	}

	return nil
}

func parse_absolute_size(tag *OpenTag, size string) {
	level := Medium
	switch size {
	case "xx-small":
		level = XXSmall
	case "x-small":
		level = XSmall
	case "small":
		level = Small
	case "medium":
		level = Medium
	case "large":
		level = Large
	case "x-large":
		level = XLarge
	case "xx-large":
		level = XXLarge
	default:
		return
	}

	// This is "absolute" in that it's relative to the base font,
	// but not to sizes created by any other tags
	factor := level.scale_factor(1.0)
	tag.add_attribute(pango_attr_scale_new(factor))
	if tag != nil {
		tag.open_tag_set_absolute_font_scale(factor)
	}
}

//  /* a string compare func that ignores '-' vs '_' differences */
//  static gint
//  attr_strcmp (gconstpointer pa,
// 		  gconstpointer pb)
//  {
//    const char *a = pa;
//    const char *b = pb;

//    int ca;
//    int cb;

//    for (*a && *b)
// 	 {
// 	   ca = *a++;
// 	   cb = *b++;

// 	   if (ca == cb)
// 	 continue;

// 	   ca = ca == '_' ? '-' : ca;
// 	   cb = cb == '_' ? '-' : cb;

// 	   if (ca != cb)
// 	 return cb - ca;
// 	 }

//    ca = *a;
//    cb = *b;

//    return cb - ca;
//  }

func span_parse_int(attrName, attrVal string) (int, error) {
	out, err := strconv.Atoi(attrVal)
	if err != nil {
		return 0, fmt.Errorf("value of '%s' attribute on <span> tag should be an integer, not '%s': %s", attrName, attrVal, err)
	}
	return out, nil
}

func span_parse_boolean(attrName, attrVal string) (bool, error) {
	switch attrVal {
	case "true", "yes", "t", "y":
		return true, nil
	case "false", "no", "f", "n":
		return false, nil
	default:
		return false, fmt.Errorf("value of '%s' attribute on <span> tag should have one of "+
			"'true/yes/t/y' or 'false/no/f/n': '%s' is not valid", attrName, attrVal)
	}
}

func span_parse_color(attrName, attrVal string, withAlpha bool) (AttrColor, uint16, error) {
	out, alpha, ok := pango_color_parse_with_alpha(attrVal, withAlpha)
	if !ok {
		return out, alpha, fmt.Errorf("value of '%s' attribute on <span> tag should be a color specification, not '%s'",
			attrName, attrVal)
	}

	return out, alpha, nil
}

func span_parse_alpha(attrName, attrVal string) (uint16, error) {
	hasPercent := false
	if strings.HasSuffix(attrVal, "%") {
		attrVal = attrVal[:len(attrVal)-1]
		hasPercent = true
	}
	intVal, err := strconv.Atoi(attrVal)
	if err != nil {
		return 0, fmt.Errorf("value of '%s' attribute on <span> tag should be an integer, not '%s': %s",
			attrName, attrVal, err)
	}

	if !hasPercent {
		return uint16(intVal), nil
	}

	if intVal > 0 && intVal <= 100 {
		return uint16(intVal * 0xffff / 100), nil
	}
	return 0, fmt.Errorf("value of '%s' attribute on <span> tag should be between 0 and 65536 or a percentage, not '%s'",
		attrName, attrVal)
}

// func span_parse_enum (const char *attrName, 		  const char *attrVal,
// 		  GType type,
// 		  int *val,
// 		  int line_number,
// 		  GError **error)
//  {
//    char *possible_values = NULL;

//    if (!_pango_parse_enum (type, attrVal, val, false, &possible_values))
// 	 {
// 	   g_set_error (error,
// 			G_MARKUP_ERROR,
// 			G_MARKUP_ERROR_INVALID_CONTENT,
// 			_("'%s' is not a valid value for the '%s' "
// 			  "attribute on <span> tag, line %d; valid "
// 			  "values are %s"),
// 			attrVal, attrName, line_number, possible_values);
// 	   g_free (possible_values);
// 	   return true;
// 	 }

//    return true;
//  }

// func span_parse_flags (const char  *attrName, 				   const char  *attrVal,
// 				   GType        type,
// 				   int         *val,
// 				   int          line_number,
// 				   GError     **error)
//  {
//    char *possible_values = NULL;

//    if (!pango_parse_flags (type, attrVal, val, &possible_values))
// 	 {
// 	   g_set_error (error,
// 					G_MARKUP_ERROR,
// 					G_MARKUP_ERROR_INVALID_CONTENT,
// 					_("'%s' is not a valid value for the '%s' "
// 					  "attribute on <span> tag, line %d; valid "
// 					  "values are %s or combinations with |"),
// 					attrVal, attrName, line_number, possible_values);
// 	   g_free (possible_values);
// 	   return true;
// 	 }

//    return true;
//  }

func checkAttribute(value *string, newAttrName, newAttrValue string) error {
	if *value != "" {
		return fmt.Errorf("attribute '%s' occurs twice on <span> tag, may only occur once", newAttrName)
	}
	*value = newAttrValue
	return nil
}

func span_parse_func(tag *OpenTag, attrs []xml.Attr) error {
	var (
		family              string
		size                string
		style               string
		weight              string
		variant             string
		stretch             string
		desc                string
		foreground          string
		background          string
		underline           string
		underline_color     string
		overline            string
		overline_color      string
		strikethrough       string
		strikethrough_color string
		rise                string
		letter_spacing      string
		lang                string
		fallback            string
		gravity             string
		gravity_hint        string
		font_features       string
		alpha               string
		background_alpha    string
		allow_breaks        string
		insert_hyphens      string
		show                string
	)

	for _, attr := range attrs {
		newAttrName := attr.Name.Local
		switch newAttrName {
		case "allow_breaks":
			err := checkAttribute(&allow_breaks, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "alpha":
			err := checkAttribute(&alpha, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "background":
			err := checkAttribute(&background, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "bgcolor":
			err := checkAttribute(&background, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "background_alpha":
			err := checkAttribute(&background_alpha, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "bgalpha":
			err := checkAttribute(&background_alpha, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "color":
			err := checkAttribute(&foreground, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "fallback":
			err := checkAttribute(&fallback, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "font":
			err := checkAttribute(&desc, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "font_desc":
			err := checkAttribute(&desc, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "face":
			err := checkAttribute(&family, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "font_family":
			err := checkAttribute(&family, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "font_size":
			err := checkAttribute(&size, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "font_stretch":
			err := checkAttribute(&stretch, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "font_style":
			err := checkAttribute(&style, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "font_variant":
			err := checkAttribute(&variant, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "font_weight":
			err := checkAttribute(&weight, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "foreground":
			err := checkAttribute(&foreground, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "fgcolor":
			err := checkAttribute(&foreground, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "fgalpha":
			err := checkAttribute(&alpha, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "font_features":
			err := checkAttribute(&font_features, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "show":
			err := checkAttribute(&show, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "size":
			err := checkAttribute(&size, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "stretch":
			err := checkAttribute(&stretch, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "strikethrough":
			err := checkAttribute(&strikethrough, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "strikethrough_color":
			err := checkAttribute(&strikethrough_color, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "style":
			err := checkAttribute(&style, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "gravity":
			err := checkAttribute(&gravity, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "gravity_hint":
			err := checkAttribute(&gravity_hint, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "insert_hyphens":
			err := checkAttribute(&insert_hyphens, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "lang":
			err := checkAttribute(&lang, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "letter_spacing":
			err := checkAttribute(&letter_spacing, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "overline":
			err := checkAttribute(&overline, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "overline_color":
			err := checkAttribute(&overline_color, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "underline":
			err := checkAttribute(&underline, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "underline_color":
			err := checkAttribute(&underline_color, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "rise":
			err := checkAttribute(&rise, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "variant":
			err := checkAttribute(&variant, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "weight":
			err := checkAttribute(&weight, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("attribute '%s' is not allowed on the <span> tag ", newAttrName)
		}
	}

	//    /* Parse desc first, then modify it with other font-related attributes. */
	//    if (desc)	 {
	// 	   PangoFontDescription *parsed;

	// 	   parsed = pango_font_description_from_string (desc);
	// 	   if (parsed)  {
	// 	   tag.add_attribute ( pango_attr_font_desc_new (parsed));
	// 	   if (tag){
	// 		 open_tag_set_absolute_font_size (tag, pango_font_description_get_size (parsed));}
	// 	   pango_font_description_free (parsed);
	// 	 }
	// 	 }

	//    if (family)	 {
	// 	   tag.add_attribute ( pango_attr_family_new (family));
	// 	 }

	//    if (size)	 {
	// 	   if (g_ascii_isdigit (*size)) 	 {
	// 	//    const char *end;
	// 	//    gint n;

	// 	   if ((end = size, !_pango_scan_int (&end, &n)) || *end != '\0' || n < 0)  {
	// 		   g_set_error (error,
	// 				G_MARKUP_ERROR,
	// 				G_MARKUP_ERROR_INVALID_CONTENT,
	// 				_("Value of 'size' attribute on <span> tag on line %d "
	// 				  "could not be parsed; should be an integer no more than %d,"
	// 				  " or a string such as 'small', not '%s'"),
	// 				line_number, INT_MAX, size);
	// 		   goto error;
	// 		 }

	// 	   tag.add_attribute ( pango_attr_size_new (n));
	// 	   if (tag){
	// 		 open_tag_set_absolute_font_size (tag, n);}
	// 	 }  else if (strcmp (size, "smaller") == 0)  {
	// 	   if (tag)  {
	// 		   tag.scale_level_delta -= 1;
	// 		   tag.scale_level -= 1;
	// 		 }
	// 	 } else if (strcmp (size, "larger") == 0) 	 {
	// 	   if (tag)  {
	// 		   tag.scale_level_delta += 1;
	// 		   tag.scale_level += 1;
	// 		 }
	// 	 }  else if (parse_absolute_size (tag, size)) {
	// 		 /* nothing */
	// 	  } else  {
	// 	   g_set_error (error,
	// 				G_MARKUP_ERROR,
	// 				G_MARKUP_ERROR_INVALID_CONTENT,
	// 				_("Value of 'size' attribute on <span> tag on line %d "+
	// 			  "could not be parsed; should be an integer, or a "+
	// 			  "string such as 'small', not '%s'"),
	// 				line_number, size);
	// 	   goto error;
	// 	 }
	// 	 }

	//    if (style)	 {
	// 	   PangoStyle pango_style;

	// 	   if (pango_parse_style (style, &pango_style, false)){
	// 	 tag.add_attribute ( pango_attr_style_new (pango_style));
	// 	  } else  {
	// 	   g_set_error (error,
	// 				G_MARKUP_ERROR,
	// 				G_MARKUP_ERROR_INVALID_CONTENT,
	// 				_("'%s' is not a valid value for the 'style' attribute "
	// 			  "on <span> tag, line %d; valid values are "
	// 			  "'normal', 'oblique', 'italic'"),
	// 				style, line_number);
	// 	   goto error;
	// 	 }
	// 	 }

	//    if (weight)	 {
	// 	   PangoWeight pango_weight;

	// 	   if (pango_parse_weight (weight, &pango_weight, false))
	// 	 tag.add_attribute (
	// 				pango_attr_weight_new (pango_weight));
	// 	   else
	// 	 {
	// 	   g_set_error (error,
	// 				G_MARKUP_ERROR,
	// 				G_MARKUP_ERROR_INVALID_CONTENT,
	// 				_("'%s' is not a valid value for the 'weight' "
	// 			  "attribute on <span> tag, line %d; valid "
	// 			  "values are for example 'light', 'ultrabold' or a number"),
	// 				weight, line_number);
	// 	   goto error;
	// 	 }
	// 	 }

	//    if (variant)	 {
	// 	   PangoVariant pango_variant;

	// 	   if (pango_parse_variant (variant, &pango_variant, false))
	// 	 tag.add_attribute ( pango_attr_variant_new (pango_variant));
	// 	   else
	// 	 {
	// 	   g_set_error (error,
	// 				G_MARKUP_ERROR,
	// 				G_MARKUP_ERROR_INVALID_CONTENT,
	// 				_("'%s' is not a valid value for the 'variant' "
	// 			  "attribute on <span> tag, line %d; valid values are "
	// 			  "'normal', 'smallcaps'"),
	// 				variant, line_number);
	// 	   goto error;
	// 	 }
	// 	 }

	//    if (stretch)	 {
	// 	   PangoStretch pango_stretch;

	// 	   if (pango_parse_stretch (stretch, &pango_stretch, false))
	// 	 tag.add_attribute ( pango_attr_stretch_new (pango_stretch));
	// 	   else
	// 	 {
	// 	   g_set_error (error,
	// 				G_MARKUP_ERROR,
	// 				G_MARKUP_ERROR_INVALID_CONTENT,
	// 				_("'%s' is not a valid value for the 'stretch' "
	// 			  "attribute on <span> tag, line %d; valid "
	// 			  "values are for example 'condensed', "
	// 			  "'ultraexpanded', 'normal'"),
	// 				stretch, line_number);
	// 	   goto error;
	// 	 }
	// 	 }

	//    if (foreground)	 {
	// 	   PangoColor color;
	// 	   guint16 alpha;

	// 	   if (!span_parse_color ("foreground", foreground, &color, &alpha, line_number, error))
	// 	 goto error;

	// 	   tag.add_attribute ( pango_attr_foreground_new (color.red, color.green, color.blue));
	// 	   if (alpha != 0xffff)
	// 		 tag.add_attribute ( pango_attr_foreground_alpha_new (alpha));
	// 	 }

	//    if (background)	 {
	// 	   PangoColor color;
	// 	   guint16 alpha;

	// 	   if (!span_parse_color ("background", background, &color, &alpha, line_number, error))
	// 	 goto error;

	// 	   tag.add_attribute ( pango_attr_background_new (color.red, color.green, color.blue));
	// 	   if (alpha != 0xffff)
	// 		 tag.add_attribute ( pango_attr_background_alpha_new (alpha));
	// 	 }

	//    if (alpha)	 {
	// 	   guint16 val;

	// 	   if (!span_parse_alpha ("alpha", alpha, &val, line_number, error))
	// 		 goto error;

	// 	   tag.add_attribute ( pango_attr_foreground_alpha_new (val));
	// 	 }

	//    if (background_alpha)	 {
	// 	   guint16 val;

	// 	   if (!span_parse_alpha ("background_alpha", background_alpha, &val, line_number, error))
	// 		 goto error;

	// 	   tag.add_attribute ( pango_attr_background_alpha_new (val));
	// 	 }

	//    if (underline)	 {
	// 	   PangoUnderline ul = PANGO_UNDERLINE_NONE;

	// 	   if (!span_parse_enum ("underline", underline, PANGO_TYPE_UNDERLINE, (int*)(void*)&ul, line_number, error))
	// 	 goto error;

	// 	   tag.add_attribute ( pango_attr_underline_new (ul));
	// 	 }

	//    if (underline_color)	 {
	// 	   PangoColor color;

	// 	   if (!span_parse_color ("underline_color", underline_color, &color, NULL, line_number, error))
	// 	 goto error;

	// 	   tag.add_attribute ( pango_attr_underline_color_new (color.red, color.green, color.blue));
	// 	 }

	//    if (overline)	 {
	// 	   PangoOverline ol = PANGO_OVERLINE_NONE;

	// 	   if (!span_parse_enum ("overline", overline, PANGO_TYPE_OVERLINE, (int*)(void*)&ol, line_number, error))
	// 	 goto error;

	// 	   tag.add_attribute ( pango_attr_overline_new (ol));
	// 	 }

	//    if (overline_color)	 {
	// 	   PangoColor color;

	// 	   if (!span_parse_color ("overline_color", overline_color, &color, NULL, line_number, error))
	// 	 goto error;

	// 	   tag.add_attribute ( pango_attr_overline_color_new (color.red, color.green, color.blue));
	// 	 }

	//    if (gravity)	 {
	// 	   PangoGravity gr = PANGO_GRAVITY_SOUTH;

	// 	   if (!span_parse_enum ("gravity", gravity, PANGO_TYPE_GRAVITY, (int*)(void*)&gr, line_number, error))
	// 	 goto error;

	// 	   if (gr == PANGO_GRAVITY_AUTO)
	// 		 {
	// 	   g_set_error (error,
	// 				G_MARKUP_ERROR,
	// 				G_MARKUP_ERROR_INVALID_CONTENT,
	// 				_("'%s' is not a valid value for the 'stretch' "
	// 			  "attribute on <span> tag, line %d; valid "
	// 			  "values are for example 'south', 'east', "
	// 			  "'north', 'west'"),
	// 				gravity, line_number);
	// 	   goto error;
	// 		 }

	// 	   tag.add_attribute ( pango_attr_gravity_new (gr));
	// 	 }

	//    if (gravity_hint)	 {
	// 	   PangoGravityHint hint = PANGO_GRAVITY_HINT_NATURAL;

	// 	   if (!span_parse_enum ("gravity_hint", gravity_hint, PANGO_TYPE_GRAVITY_HINT, (int*)(void*)&hint, line_number, error))
	// 	 goto error;

	// 	   tag.add_attribute ( pango_attr_gravity_hint_new (hint));
	// 	 }

	//    if (strikethrough)	 {
	// 	   gboolean b = false;

	// 	   if (!span_parse_boolean ("strikethrough", strikethrough, &b, line_number, error))
	// 	 goto error;

	// 	   tag.add_attribute ( pango_attr_strikethrough_new (b));
	// 	 }

	//    if (strikethrough_color)	 {
	// 	   PangoColor color;

	// 	   if (!span_parse_color ("strikethrough_color", strikethrough_color, &color, NULL, line_number, error))
	// 	 goto error;

	// 	   tag.add_attribute ( pango_attr_strikethrough_color_new (color.red, color.green, color.blue));
	// 	 }

	//    if (fallback)	 {
	// 	   gboolean b = false;

	// 	   if (!span_parse_boolean ("fallback", fallback, &b, line_number, error))
	// 	 goto error;

	// 	   tag.add_attribute ( pango_attr_fallback_new (b));
	// 	 }

	//    if (show)	 {
	// 	   PangoShowFlags flags;

	// 	   if (!span_parse_flags ("show", show, PANGO_TYPE_SHOW_FLAGS, (int*)(void*)&flags, line_number, error))
	// 	 goto error;

	// 	   tag.add_attribute ( pango_attr_show_new (flags));
	// 	 }

	//    if (rise)	 {
	// 	   gint n = 0;

	// 	   if (!span_parse_int ("rise", rise, &n, line_number, error))
	// 	 goto error;

	// 	   tag.add_attribute ( pango_attr_rise_new (n));
	// 	 }

	//    if (letter_spacing)	 {
	// 	   gint n = 0;

	// 	   if (!span_parse_int ("letter_spacing", letter_spacing, &n, line_number, error))
	// 	 goto error;

	// 	   tag.add_attribute ( pango_attr_letter_spacing_new (n));
	// 	 }

	//    if (lang)	 {
	// 	   tag.add_attribute (
	// 			  pango_attr_language_new (pango_language_from_string (lang)));
	// 	 }

	//    if (font_features)	 {
	// 	   tag.add_attribute ( pango_attr_font_features_new (font_features));
	// 	 }

	//    if (allow_breaks)	 {
	// 	   gboolean b = false;

	// 	   if (!span_parse_boolean ("allow_breaks", allow_breaks, &b, line_number, error))
	// 	 goto error;

	// 	   tag.add_attribute ( pango_attr_allow_breaks_new (b));
	// 	 }

	//    if (insert_hyphens)	 {
	// 	   gboolean b = false;

	// 	   if (!span_parse_boolean ("insert_hyphens", insert_hyphens, &b, line_number, error))
	// 	 goto error;

	// 	   tag.add_attribute ( pango_attr_insert_hyphens_new (b));
	// 	 }

	//    return true;

	//   error:

	return false
}

func i_parse_func(tag *OpenTag, names []xml.Attr) error {
	if err := CHECK_NO_ATTRS("i", names); err != nil {
		return err
	}
	tag.add_attribute(pango_attr_style_new(PANGO_STYLE_ITALIC))

	return nil
}

func markup_parse_func(tag *OpenTag, names []xml.Attr) error {
	/* We don't do anything with this tag at the moment. */
	return nil
}

func s_parse_func(tag *OpenTag, names []xml.Attr) error {
	if err := CHECK_NO_ATTRS("s", names); err != nil {
		return err
	}

	tag.add_attribute(pango_attr_strikethrough_new(true))

	return nil
}

const SUPERSUB_RISE = 5000

func sub_parse_func(tag *OpenTag, names []xml.Attr) error {
	if err := CHECK_NO_ATTRS("sub", names); err != nil {
		return err
	}

	/* Shrink font, and set a negative rise */
	if tag != nil {
		tag.scale_level_delta -= 1
		tag.scale_level -= 1
	}

	tag.add_attribute(pango_attr_rise_new(-SUPERSUB_RISE))

	return nil
}

func sup_parse_func(tag *OpenTag, names []xml.Attr) error {
	if err := CHECK_NO_ATTRS("sup", names); err != nil {
		return err
	}

	/* Shrink font, and set a positive rise */
	if tag != nil {
		tag.scale_level_delta -= 1
		tag.scale_level -= 1
	}

	tag.add_attribute(pango_attr_rise_new(SUPERSUB_RISE))

	return nil
}

func small_parse_func(tag *OpenTag, names []xml.Attr) error {
	if err := CHECK_NO_ATTRS("small", names); err != nil {
		return err
	}

	// Shrink text one level
	if tag != nil {
		tag.scale_level_delta -= 1
		tag.scale_level -= 1
	}

	return nil
}

func tt_parse_func(tag *OpenTag, names []xml.Attr) error {
	if err := CHECK_NO_ATTRS("tt", names); err != nil {
		return err
	}

	tag.add_attribute(pango_attr_family_new("Monospace"))

	return nil
}

func u_parse_func(tag *OpenTag, names []xml.Attr) error {
	if err := CHECK_NO_ATTRS("u", names); err != nil {
		return err
	}

	tag.add_attribute(pango_attr_underline_new(PANGO_UNDERLINE_SINGLE))

	return nil
}
