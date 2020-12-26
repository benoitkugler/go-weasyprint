package pango

import "log"

// /**
//  * SECTION:main
//  * @title:Rendering
//  * @short_description:Functions to run the rendering pipeline
//  *
//  * The Pango rendering pipeline takes a string of
//  * Unicode characters and converts it into glyphs.
//  * The functions described in this section accomplish
//  * various steps of this process.
//  *
//  * ![](pipeline.png)
//  */

//  /**
//   * SECTION:context
//   * @title:Contexts
//   * @short_description: Global context object
//   *
//   * The #Context structure stores global information
//   * influencing Pango's operation, such as the fontmap used
//   * to look up fonts, and default values such as the default
//   * language, default gravity, or default font.
//   */

// Context stores global information
// used to control the itemization process.
type Context struct {
	//    GObject parent_instance;
	serial uint
	//     fontmap_serial uint

	set_language Language // the global language tag for the context.

	language Language
	base_dir Direction
	//    PangoGravity base_gravity;
	resolved_gravity Gravity
	gravity_hint     GravityHint

	font_desc FontDescription

	//    PangoMatrix *matrix;

	font_map FontMap

	//    bool round_glyph_positions;
}

// pango_context_load_font loads the font in one of the fontmaps in the context
// that is the closest match for `desc`, or nil if no font matched.
func (context *Context) pango_context_load_font(desc FontDescription) Font {
	if context == nil || context.font_map == nil {
		return nil
	}
	return context.font_map.load_font(context, desc)
}

// pango_itemize_with_base_dir is like pango_itemize(), but the base direction to use when
// computing bidirectional levels (see pango_context_set_base_dir ()),
// is specified explicitly rather than gotten from the Context.
// func (context *Context) pango_itemize_with_base_dir(base_dir Direction, text []rune,
// 	//   start_index,   length int,
// 	attrs AttrList, cached_iter *AttrIterator) []Item {
// 	//    ItemizeState state;

// 	//    g_return_val_if_fail (context != nil, nil);
// 	//    g_return_val_if_fail (start_index >= 0, nil);
// 	//    g_return_val_if_fail (length >= 0, nil);
// 	//    g_return_val_if_fail (length == 0 || text != nil, nil);

// 	if context == nil || len(text) == 0 {
// 		return nil
// 	}

// 	itemize_state_init(&state, context, text, base_dir, start_index, length,
// 		attrs, cached_iter, nil)

// 	do := true // do ... for
// 	for do {
// 		itemize_state_process_run(&state)
// 		do = itemize_state_next(&state)
// 	}

// 	itemize_state_finish(&state)

// 	return g_list_reverse(state.result)
// }

//  struct _ContextClass
//  {
//    GObjectClass parent_class;

//  };

//  static void pango_context_finalize    (GObject       *object);
//  static void context_changed           (Context  *context);

//  G_DEFINE_TYPE (Context, pango_context, G_TYPE_OBJECT)

//  static void
//  pango_context_init (Context *context)
//  {
//    context.base_dir = PANGO_DIRECTION_WEAK_LTR;
//    context.resolved_gravity = context.base_gravity = PANGO_GRAVITY_SOUTH;
//    context.gravity_hint = PANGO_GRAVITY_HINT_NATURAL;

//    context.serial = 1;
//    context.set_language = nil;
//    context.language = pango_language_get_default ();
//    context.font_map = nil;
//    context.round_glyph_positions = true;

//    context.font_desc = pango_font_description_new ();
//    pango_font_description_set_family_static (context.font_desc, "serif");
//    pango_font_description_set_style (context.font_desc, PANGO_STYLE_NORMAL);
//    pango_font_description_set_variant (context.font_desc, PANGO_VARIANT_NORMAL);
//    pango_font_description_set_weight (context.font_desc, PANGO_WEIGHT_NORMAL);
//    pango_font_description_set_stretch (context.font_desc, PANGO_STRETCH_NORMAL);
//    pango_font_description_set_size (context.font_desc, 12 * PANGO_SCALE);
//  }

//  static void
//  pango_context_class_init (ContextClass *klass)
//  {
//    GObjectClass *object_class = G_OBJECT_CLASS (klass);

//    object_class.finalize = pango_context_finalize;
//  }

//  static void
//  pango_context_finalize (GObject *object)
//  {
//    Context *context;

//    context = PANGO_CONTEXT (object);

//    if (context.font_map)
// 	 g_object_unref (context.font_map);

//    pango_font_description_free (context.font_desc);
//    if (context.matrix)
// 	 pango_matrix_free (context.matrix);

//    G_OBJECT_CLASS (pango_context_parent_class).finalize (object);
//  }

//  /**
//   * pango_context_new:
//   *
//   * Creates a new #Context initialized to default values.
//   *
//   * This function is not particularly useful as it should always
//   * be followed by a pango_context_set_font_map() call, and the
//   * function pango_font_map_create_context() does these two steps
//   * together and hence users are recommended to use that.
//   *
//   * If you are using Pango as part of a higher-level system,
//   * that system may have it's own way of create a #Context.
//   * For instance, the GTK+ toolkit has, among others,
//   * gdk_pango_context_get_for_screen(), and
//   * gtk_widget_get_pango_context().  Use those instead.
//   *
//   * Return value: the newly allocated #Context, which should
//   *               be freed with g_object_unref().
//   **/
//  Context *
//  pango_context_new (void)
//  {
//    Context *context;

//    context = g_object_new (PANGO_TYPE_CONTEXT, nil);

//    return context;
//  }

//  static void
//  update_resolved_gravity (Context *context)
//  {
//    if (context.base_gravity == PANGO_GRAVITY_AUTO)
// 	 context.resolved_gravity = pango_gravity_get_for_matrix (context.matrix);
//    else
// 	 context.resolved_gravity = context.base_gravity;
//  }

//  /**
//   * pango_context_set_matrix:
//   * `context`: a #Context
//   * @matrix: (allow-none): a #PangoMatrix, or %nil to unset any existing
//   * matrix. (No matrix set is the same as setting the identity matrix.)
//   *
//   * Sets the transformation matrix that will be applied when rendering
//   * with this context. Note that reported metrics are in the user space
//   * coordinates before the application of the matrix, not device-space
//   * coordinates after the application of the matrix. So, they don't scale
//   * with the matrix, though they may change slightly for different
//   * matrices, depending on how the text is fit to the pixel grid.
//   *
//   * Since: 1.6
//   **/
//  void
//  pango_context_set_matrix (Context       *context,
// 			   const PangoMatrix  *matrix)
//  {
//    g_return_if_fail (PANGO_IS_CONTEXT (context));

//    if (context.matrix || matrix)
// 	 context_changed (context);

//    if (context.matrix)
// 	 pango_matrix_free (context.matrix);
//    if (matrix)
// 	 context.matrix = pango_matrix_copy (matrix);
//    else
// 	 context.matrix = nil;

//    update_resolved_gravity (context);
//  }

//  /**
//   * pango_context_get_matrix:
//   * `context`: a #Context
//   *
//   * Gets the transformation matrix that will be applied when
//   * rendering with this context. See pango_context_set_matrix().
//   *
//   * Return value: (nullable): the matrix, or %nil if no matrix has
//   *  been set (which is the same as the identity matrix). The returned
//   *  matrix is owned by Pango and must not be modified or freed.
//   *
//   * Since: 1.6
//   **/
//  const PangoMatrix *
//  pango_context_get_matrix (Context *context)
//  {
//    g_return_val_if_fail (PANGO_IS_CONTEXT (context), nil);

//    return context.matrix;
//  }

//  /**
//   * pango_context_set_font_map:
//   * `context`: a #Context
//   * @font_map: the #PangoFontMap to set.
//   *
//   * Sets the font map to be searched when fonts are looked-up in this context.
//   * This is only for internal use by Pango backends, a #Context obtained
//   * via one of the recommended methods should already have a suitable font map.
//   **/
//  void
//  pango_context_set_font_map (Context *context,
// 				 PangoFontMap *font_map)
//  {
//    g_return_if_fail (PANGO_IS_CONTEXT (context));
//    g_return_if_fail (!font_map || PANGO_IS_FONT_MAP (font_map));

//    if (font_map == context.font_map)
// 	 return;

//    context_changed (context);

//    if (font_map)
// 	 g_object_ref (font_map);

//    if (context.font_map)
// 	 g_object_unref (context.font_map);

//    context.font_map = font_map;
//    context.fontmap_serial = pango_font_map_get_serial (font_map);
//  }

//  /**
//   * pango_context_get_font_map:
//   * `context`: a #Context
//   *
//   * Gets the #PangoFontMap used to look up fonts for this context.
//   *
//   * Return value: (transfer none): the font map for the #Context.
//   *               This value is owned by Pango and should not be unreferenced.
//   *
//   * Since: 1.6
//   **/
//  PangoFontMap *
//  pango_context_get_font_map (Context *context)
//  {
//    g_return_val_if_fail (PANGO_IS_CONTEXT (context), nil);

//    return context.font_map;
//  }

//  /**
//   * pango_context_list_families:
//   * `context`: a #Context
//   * @families: (out) (array length=n_families) (transfer container): location to store a pointer to
//   *            an array of #PangoFontFamily *. This array should be freed
//   *            with g_free().
//   * @n_families: (out): location to store the number of elements in @descs
//   *
//   * List all families for a context.
//   **/
//  void
//  pango_context_list_families (Context          *context,
// 				  PangoFontFamily     ***families,
// 				  int                   *n_families)
//  {
//    g_return_if_fail (context != nil);
//    g_return_if_fail (families == nil || n_families != nil);

//    if (n_families == nil)
// 	 return;

//    if (context.font_map == nil)
// 	 {
// 	   *n_families = 0;
// 	   if (families)
// 	 *families = nil;

// 	   return;
// 	 }
//    else
// 	 pango_font_map_list_families (context.font_map, families, n_families);
//  }

//  /**
//   * pango_context_load_fontset:
//   * `context`: a #Context
//   * @desc: a #PangoFontDescription describing the fonts to load
//   * @language: a #PangoLanguage the fonts will be used for
//   *
//   * Load a set of fonts in the context that can be used to render
//   * a font matching @desc.
//   *
//   * Returns: (transfer full) (nullable): the newly allocated
//   *          #PangoFontset loaded, or %nil if no font matched.
//   **/
//  PangoFontset *
//  pango_context_load_fontset (Context               *context,
// 				 const PangoFontDescription *desc,
// 				 PangoLanguage             *language)
//  {
//    g_return_val_if_fail (context != nil, nil);

//    return pango_font_map_load_fontset (context.font_map, context, desc, language);
//  }

//  /**
//   * pango_context_set_font_description:
//   * `context`: a #Context
//   * @desc: the new pango font description
//   *
//   * Set the default font description for the context
//   **/
//  void
//  pango_context_set_font_description (Context               *context,
// 					 const PangoFontDescription *desc)
//  {
//    g_return_if_fail (context != nil);
//    g_return_if_fail (desc != nil);

//    if (desc != context.font_desc &&
// 	   (!desc || !context.font_desc || !pango_font_description_equal(desc, context.font_desc)))
// 	 {
// 	   context_changed (context);

// 	   pango_font_description_free (context.font_desc);
// 	   context.font_desc = pango_font_description_copy (desc);
// 	 }
//  }

//  /**
//   * pango_context_get_font_description:
//   * `context`: a #Context
//   *
//   * Retrieve the default font description for the context.
//   *
//   * Return value: (transfer none): a pointer to the context's default font
//   *               description. This value must not be modified or freed.
//   **/
//  PangoFontDescription *
//  pango_context_get_font_description (Context *context)
//  {
//    g_return_val_if_fail (context != nil, nil);

//    return context.font_desc;
//  }

//  /**
//   * pango_context_set_language:
//   * `context`: a #Context
//   * @language: the new language tag.
//   *
//   * Sets the global language tag for the context.  The default language
//   * for the locale of the running process can be found using
//   * pango_language_get_default().
//   **/
//  void
//  pango_context_set_language (Context *context,
// 				 PangoLanguage    *language)
//  {
//    g_return_if_fail (context != nil);

//    if (language != context.language)
// 	 context_changed (context);

//    context.set_language = language;
//    if (language)
// 	 context.language = language;
//    else
// 	 context.language = pango_language_get_default ();
//  }

//  /**
//   * pango_context_get_language:
//   * `context`: a #Context
//   *
//   * Retrieves the global language tag for the context.
//   *
//   * Return value: the global language tag.
//   **/
//  PangoLanguage *
//  pango_context_get_language (Context *context)
//  {
//    g_return_val_if_fail (context != nil, nil);

//    return context.set_language;
//  }

//  /**
//   * pango_context_set_base_dir:
//   * `context`: a #Context
//   * @direction: the new base direction
//   *
//   * Sets the base direction for the context.
//   *
//   * The base direction is used in applying the Unicode bidirectional
//   * algorithm; if the @direction is %PANGO_DIRECTION_LTR or
//   * %PANGO_DIRECTION_RTL, then the value will be used as the paragraph
//   * direction in the Unicode bidirectional algorithm.  A value of
//   * %PANGO_DIRECTION_WEAK_LTR or %PANGO_DIRECTION_WEAK_RTL is used only
//   * for paragraphs that do not contain any strong characters themselves.
//   **/
//  void
//  pango_context_set_base_dir (Context  *context,
// 				 PangoDirection direction)
//  {
//    g_return_if_fail (context != nil);

//    if (direction != context.base_dir)
// 	 context_changed (context);

//    context.base_dir = direction;
//  }

//  /**
//   * pango_context_get_base_dir:
//   * `context`: a #Context
//   *
//   * Retrieves the base direction for the context. See
//   * pango_context_set_base_dir().
//   *
//   * Return value: the base direction for the context.
//   **/
//  PangoDirection
//  pango_context_get_base_dir (Context *context)
//  {
//    g_return_val_if_fail (context != nil, PANGO_DIRECTION_LTR);

//    return context.base_dir;
//  }

//  /**
//   * pango_context_set_base_gravity:
//   * `context`: a #Context
//   * @gravity: the new base gravity
//   *
//   * Sets the base gravity for the context.
//   *
//   * The base gravity is used in laying vertical text out.
//   *
//   * Since: 1.16
//   **/
//  void
//  pango_context_set_base_gravity (Context  *context,
// 				 PangoGravity gravity)
//  {
//    g_return_if_fail (context != nil);

//    if (gravity != context.base_gravity)
// 	 context_changed (context);

//    context.base_gravity = gravity;

//    update_resolved_gravity (context);
//  }

//  /**
//   * pango_context_get_base_gravity:
//   * `context`: a #Context
//   *
//   * Retrieves the base gravity for the context. See
//   * pango_context_set_base_gravity().
//   *
//   * Return value: the base gravity for the context.
//   *
//   * Since: 1.16
//   **/
//  PangoGravity
//  pango_context_get_base_gravity (Context *context)
//  {
//    g_return_val_if_fail (context != nil, PANGO_GRAVITY_SOUTH);

//    return context.base_gravity;
//  }

//  /**
//   * pango_context_get_gravity:
//   * `context`: a #Context
//   *
//   * Retrieves the gravity for the context. This is similar to
//   * pango_context_get_base_gravity(), except for when the base gravity
//   * is %PANGO_GRAVITY_AUTO for which pango_gravity_get_for_matrix() is used
//   * to return the gravity from the current context matrix.
//   *
//   * Return value: the resolved gravity for the context.
//   *
//   * Since: 1.16
//   **/
//  PangoGravity
//  pango_context_get_gravity (Context *context)
//  {
//    g_return_val_if_fail (context != nil, PANGO_GRAVITY_SOUTH);

//    return context.resolved_gravity;
//  }

//  /**
//   * pango_context_set_gravity_hint:
//   * `context`: a #Context
//   * @hint: the new gravity hint
//   *
//   * Sets the gravity hint for the context.
//   *
//   * The gravity hint is used in laying vertical text out, and is only relevant
//   * if gravity of the context as returned by pango_context_get_gravity()
//   * is set %PANGO_GRAVITY_EAST or %PANGO_GRAVITY_WEST.
//   *
//   * Since: 1.16
//   **/
//  void
//  pango_context_set_gravity_hint (Context    *context,
// 				 PangoGravityHint hint)
//  {
//    g_return_if_fail (context != nil);

//    if (hint != context.gravity_hint)
// 	 context_changed (context);

//    context.gravity_hint = hint;
//  }

//  /**
//   * pango_context_get_gravity_hint:
//   * `context`: a #Context
//   *
//   * Retrieves the gravity hint for the context. See
//   * pango_context_set_gravity_hint() for details.
//   *
//   * Return value: the gravity hint for the context.
//   *
//   * Since: 1.16
//   **/
//  PangoGravityHint
//  pango_context_get_gravity_hint (Context *context)
//  {
//    g_return_val_if_fail (context != nil, PANGO_GRAVITY_HINT_NATURAL);

//    return context.gravity_hint;
//  }

//  /**********************************************************************/

func (iterator *AttrIterator) advance_attr_iterator_to(start_index int) bool {
	start_range, end_range := iterator.StartIndex, iterator.EndIndex

	for start_index >= end_range {
		if !iterator.pango_attr_iterator_next() {
			return false
		}
		start_range, end_range = iterator.StartIndex, iterator.EndIndex
	}

	if start_range > start_index {
		log.Println("In pango_itemize(), the cached iterator passed in " +
			"had already moved beyond the start_index")
	}

	return true
}

//  /***************************************************************************
//   * We cache the results of character,fontset => font in a hash table
//   ***************************************************************************/

//  typedef struct {
//    GHashTable *hash;
//  } FontCache;

//  typedef struct {
//    PangoFont *font;
//  } FontElement;

//  static void
//  font_cache_destroy (FontCache *cache)
//  {
//    g_hash_table_destroy (cache.hash);
//    g_slice_free (FontCache, cache);
//  }

//  static void
//  font_element_destroy (FontElement *element)
//  {
//    if (element.font)
// 	 g_object_unref (element.font);
//    g_slice_free (FontElement, element);
//  }

//  static FontCache *
//  get_font_cache (PangoFontset *fontset)
//  {
//    FontCache *cache;

//    static GQuark cache_quark = 0; /* MT-safe */
//    if (G_UNLIKELY (!cache_quark))
// 	 cache_quark = g_quark_from_static_string ("pango-font-cache");

//  retry:
//    cache = g_object_get_qdata (G_OBJECT (fontset), cache_quark);
//    if (G_UNLIKELY (!cache))
// 	 {
// 	   cache = g_slice_new (FontCache);
// 	   cache.hash = g_hash_table_new_full (g_direct_hash, nil,
// 						nil, (GDestroyNotify)font_element_destroy);
// 	   if (!g_object_replace_qdata (G_OBJECT (fontset), cache_quark, nil,
// 									cache, (GDestroyNotify)font_cache_destroy,
// 									nil))
// 		 {
// 		   font_cache_destroy (cache);
// 		   goto retry;
// 		 }
// 	 }

//    return cache;
//  }

//  static bool
//  font_cache_get (FontCache   *cache,
// 		 gunichar     wc,
// 		 PangoFont  **font)
//  {
//    FontElement *element;

//    element = g_hash_table_lookup (cache.hash, GUINT_TO_POINTER (wc));
//    if (element)
// 	 {
// 	   *font = element.font;

// 	   return true;
// 	 }
//    else
// 	 return false;
//  }

//  static void
//  font_cache_insert (FontCache   *cache,
// 			gunichar           wc,
// 			PangoFont         *font)
//  {
//    FontElement *element = g_slice_new (FontElement);
//    element.font = font ? g_object_ref (font) : nil;

//    g_hash_table_insert (cache.hash, GUINT_TO_POINTER (wc), element);
//  }

//  /**********************************************************************/

type ChangedFlags uint8

const (
	EMBEDDING_CHANGED ChangedFlags = 1 << iota
	SCRIPT_CHANGED
	LANG_CHANGED
	FONT_CHANGED
	DERIVED_LANG_CHANGED
	WIDTH_CHANGED
	EMOJI_CHANGED
)

type WidthIter struct {
	text       []rune
	start, end int
	upright    bool
}

func (iter *WidthIter) width_iter_init(text []rune) {
	iter.text = text
	iter.width_iter_next()
}

/* https://www.unicode.org/Public/11.0.0/ucd/VerticalOrientation.txt
* VO=U or Tu table generated by tools/gen-vertical-orientation-U-table.py.
*
* FIXME: In the future, If GLib supports VerticalOrientation, please use it.
 */
var upright = [...][2]rune{
	{0x00A7, 0x00A7}, {0x00A9, 0x00A9}, {0x00AE, 0x00AE}, {0x00B1, 0x00B1},
	{0x00BC, 0x00BE}, {0x00D7, 0x00D7}, {0x00F7, 0x00F7}, {0x02EA, 0x02EB},
	{0x1100, 0x11FF}, {0x1401, 0x167F}, {0x18B0, 0x18FF}, {0x2016, 0x2016},
	{0x2020, 0x2021}, {0x2030, 0x2031}, {0x203B, 0x203C}, {0x2042, 0x2042},
	{0x2047, 0x2049}, {0x2051, 0x2051}, {0x2065, 0x2065}, {0x20DD, 0x20E0},
	{0x20E2, 0x20E4}, {0x2100, 0x2101}, {0x2103, 0x2109}, {0x210F, 0x210F},
	{0x2113, 0x2114}, {0x2116, 0x2117}, {0x211E, 0x2123}, {0x2125, 0x2125},
	{0x2127, 0x2127}, {0x2129, 0x2129}, {0x212E, 0x212E}, {0x2135, 0x213F},
	{0x2145, 0x214A}, {0x214C, 0x214D}, {0x214F, 0x2189}, {0x218C, 0x218F},
	{0x221E, 0x221E}, {0x2234, 0x2235}, {0x2300, 0x2307}, {0x230C, 0x231F},
	{0x2324, 0x2328}, {0x232B, 0x232B}, {0x237D, 0x239A}, {0x23BE, 0x23CD},
	{0x23CF, 0x23CF}, {0x23D1, 0x23DB}, {0x23E2, 0x2422}, {0x2424, 0x24FF},
	{0x25A0, 0x2619}, {0x2620, 0x2767}, {0x2776, 0x2793}, {0x2B12, 0x2B2F},
	{0x2B50, 0x2B59}, {0x2BB8, 0x2BD1}, {0x2BD3, 0x2BEB}, {0x2BF0, 0x2BFF},
	{0x2E80, 0x3007}, {0x3012, 0x3013}, {0x3020, 0x302F}, {0x3031, 0x309F},
	{0x30A1, 0x30FB}, {0x30FD, 0xA4CF}, {0xA960, 0xA97F}, {0xAC00, 0xD7FF},
	{0xE000, 0xFAFF}, {0xFE10, 0xFE1F}, {0xFE30, 0xFE48}, {0xFE50, 0xFE57},
	{0xFE5F, 0xFE62}, {0xFE67, 0xFE6F}, {0xFF01, 0xFF07}, {0xFF0A, 0xFF0C},
	{0xFF0E, 0xFF19}, {0xFF1F, 0xFF3A}, {0xFF3C, 0xFF3C}, {0xFF3E, 0xFF3E},
	{0xFF40, 0xFF5A}, {0xFFE0, 0xFFE2}, {0xFFE4, 0xFFE7}, {0xFFF0, 0xFFF8},
	{0xFFFC, 0xFFFD}, {0x10980, 0x1099F}, {0x11580, 0x115FF}, {0x11A00, 0x11AAF},
	{0x13000, 0x1342F}, {0x14400, 0x1467F}, {0x16FE0, 0x18AFF}, {0x1B000, 0x1B12F},
	{0x1B170, 0x1B2FF}, {0x1D000, 0x1D1FF}, {0x1D2E0, 0x1D37F}, {0x1D800, 0x1DAAF},
	{0x1F000, 0x1F7FF}, {0x1F900, 0x1FA6F}, {0x20000, 0x2FFFD}, {0x30000, 0x3FFFD},
	{0xF0000, 0xFFFFD}, {0x100000, 0x10FFFD},
}

func width_iter_is_upright(ch rune) bool {
	const max = len(upright)
	st := 0
	ed := max

	for st <= ed {
		mid := (st + ed) / 2
		if upright[mid][0] <= ch && ch <= upright[mid][1] {
			return true
		} else {
			if upright[mid][0] <= ch {
				st = mid + 1
			} else {
				ed = mid - 1
			}
		}
	}

	return false
}

func (iter *WidthIter) width_iter_next() {
	met_joiner := false
	iter.start = iter.end

	if iter.end < len(iter.text)-1 {
		ch := iter.text[iter.end]
		iter.end++
		iter.upright = width_iter_is_upright(ch)
	}

	for iter.end < len(iter.text)-1 {
		ch := iter.text[iter.end]
		iter.end++

		/* for zero width joiner */
		if ch == 0x200D {
			iter.end++
			met_joiner = true
			continue
		}

		/* ignore the upright check if met joiner */
		if met_joiner {
			iter.end++
			met_joiner = false
			continue
		}

		/* for variation selector, tag and emoji modifier. */
		if ch == 0xFE0E || ch == 0xFE0F || (ch >= 0xE0020 && ch <= 0xE007F) || (ch >= 0x1F3FB && ch <= 0x1F3FF) {
			iter.end++
			continue
		}

		if width_iter_is_upright(ch) != iter.upright {
			break
		}
		iter.end++
	}
}

type ItemizeState struct {
	context *Context
	text    []rune
	end     int // index into text

	run_start, run_end int // index in text

	//    GList *result;
	item *Item

	embedding_levels     []uint8
	embedding_end_offset int
	embedding_end        int
	embedding            uint8

	gravity           Gravity
	gravity_hint      GravityHint
	resolved_gravity  Gravity
	font_desc_gravity Gravity
	centered_baseline bool

	attr_iter *AttrIterator
	// free_attr_iter bool
	attr_end         int
	font_desc        *FontDescription
	emoji_font_desc  *FontDescription
	lang             Language
	extra_attrs      AttrList
	copy_extra_attrs bool

	changed ChangedFlags

	script_iter ScriptIter
	script_end  int    // copied from `script_iter`
	script      Script // copied from `script_iter`

	width_iter WidthIter
	emoji_iter EmojiIter

	derived_lang *Language

	current_fonts *Fontset
	// cache           *FontCache
	base_font       *Font
	enable_fallback bool
}

func (state *ItemizeState) update_embedding_end() {
	state.embedding = state.embedding_levels[state.embedding_end_offset]
	for state.embedding_end < state.end &&
		state.embedding_levels[state.embedding_end_offset] == state.embedding {
		state.embedding_end_offset++
		state.embedding_end++
	}

	state.changed |= EMBEDDING_CHANGED
}

func (attr_list AttrList) find_attribute(type_ AttrType) *Attribute {
	for _, attr := range attr_list {
		if attr.Type == type_ {
			return attr
		}
	}
	return nil
}

func (state *ItemizeState) update_attr_iterator() {
	//    PangoLanguage *old_lang;
	//    PangoAttribute *attr;
	//    int end_index;
	end_index := state.attr_iter.EndIndex // pango_attr_iterator_range (state.attr_iter, nil, &end_index);
	if end_index < state.end {
		state.attr_end = end_index
	} else {
		state.attr_end = state.end
	}

	if state.emoji_font_desc != nil {
		state.emoji_font_desc = nil
	}

	old_lang := state.lang

	cp := state.context.font_desc // copy
	state.font_desc = &cp
	state.attr_iter.pango_attr_iterator_get_font(state.font_desc, &state.lang, &state.extra_attrs)
	if state.font_desc.mask&PANGO_FONT_MASK_GRAVITY != 0 {
		state.font_desc_gravity = state.font_desc.gravity
	} else {
		state.font_desc_gravity = PANGO_GRAVITY_AUTO
	}

	state.copy_extra_attrs = false

	if state.lang == "" {
		state.lang = state.context.language
	}

	attr := state.extra_attrs.find_attribute(ATTR_FALLBACK)
	state.enable_fallback = (attr == nil || attr.Data.(AttrInt) != 0)

	attr = state.extra_attrs.find_attribute(ATTR_GRAVITY)
	state.gravity = PANGO_GRAVITY_AUTO
	if attr != nil {
		state.gravity = Gravity(attr.Data.(AttrInt))
	}

	attr = state.extra_attrs.find_attribute(ATTR_GRAVITY_HINT)
	state.gravity_hint = state.context.gravity_hint
	if attr != nil {
		state.gravity_hint = GravityHint(attr.Data.(AttrInt))
	}

	state.changed |= FONT_CHANGED
	if state.lang != old_lang {
		state.changed |= LANG_CHANGED
	}
}

func (state *ItemizeState) update_end() {
	state.run_end = state.embedding_end
	if state.attr_end < state.run_end {
		state.run_end = state.attr_end
	}
	if state.script_end < state.run_end {
		state.run_end = state.script_end
	}
	if state.width_iter.end < state.run_end {
		state.run_end = state.width_iter.end
	}
	if state.emoji_iter.end < state.run_end {
		state.run_end = state.emoji_iter.end
	}
}

//  }

//  static void
//  width_iter_fini (PangoWidthIter* iter)
//  {
//  }

func (context *Context) itemize_state_init(text []rune,
	base_dir Direction,
	start_index, length int,
	attrs AttrList,
	cached_iter *AttrIterator,
	desc *FontDescription) *ItemizeState {

	var state ItemizeState
	state.context = context
	state.text = text
	state.end = start_index + length

	// state.result = nil
	// state.item = nil

	state.run_start = start_index
	state.changed = EMBEDDING_CHANGED | SCRIPT_CHANGED | LANG_CHANGED |
		FONT_CHANGED | WIDTH_CHANGED | EMOJI_CHANGED

	// First, apply the bidirectional algorithm to break the text into directional runs.
	// TODO:
	// state.embedding_levels = pango_log2vis_get_embedding_levels(text+start_index, length, &base_dir)

	state.embedding_end_offset = 0
	state.embedding_end = start_index
	state.update_embedding_end()

	/* Initialize the attribute iterator
	 */
	if cached_iter != nil {
		state.attr_iter = cached_iter
	} else if len(attrs) != 0 {
		state.attr_iter = attrs.pango_attr_list_get_iterator()
	}

	if state.attr_iter != nil {
		state.attr_iter.advance_attr_iterator_to(start_index)
		state.update_attr_iterator()
	} else {
		if desc == nil {
			cp := state.context.font_desc
			state.font_desc = &cp
		} else {
			state.font_desc = desc
		}
		state.lang = state.context.language
		state.extra_attrs = nil
		state.copy_extra_attrs = false

		state.attr_end = state.end
		state.enable_fallback = true
	}

	/* Initialize the script iterator
	 */
	state.script_iter._pango_script_iter_init(text[start_index:])
	state.script_end, state.script = state.script_iter.script_end, state.script_iter.script_code

	state.width_iter.width_iter_init(text[start_index:])
	state.emoji_iter._pango_emoji_iter_init(text[start_index:])

	if state.emoji_iter.is_emoji {
		state.width_iter.end = max(state.width_iter.end, state.emoji_iter.end)
	}

	state.update_end()

	if state.font_desc.mask&PANGO_FONT_MASK_GRAVITY != 0 {
		state.font_desc_gravity = state.font_desc.gravity
	} else {
		state.font_desc_gravity = PANGO_GRAVITY_AUTO
	}

	state.gravity = PANGO_GRAVITY_AUTO
	state.centered_baseline = state.context.resolved_gravity.isVertical()
	state.gravity_hint = state.context.gravity_hint
	state.resolved_gravity = PANGO_GRAVITY_AUTO
	// state.derived_lang = nil
	// state.current_fonts = nil
	// state.cache = nil
	// state.base_font = nil

	return &state
}

//  static bool
//  itemize_state_next (ItemizeState *state)
//  {
//    if (state.run_end == state.end)
// 	 return false;

//    state.changed = 0;

//    state.run_start = state.run_end;

//    if (state.run_end == state.embedding_end)
// 	 {
// 	   update_embedding_end (state);
// 	 }

//    if (state.run_end == state.attr_end)
// 	 {
// 	   pango_attr_iterator_next (state.attr_iter);
// 	   update_attr_iterator (state);
// 	 }

//    if (state.run_end == state.script_end)
// 	 {
// 	   pango_script_iter_next (&state.script_iter);
// 	   pango_script_iter_get_range (&state.script_iter, nil,
// 					&state.script_end, &state.script);
// 	   state.changed |= SCRIPT_CHANGED;
// 	 }
//    if (state.run_end == state.emoji_iter.end)
// 	 {
// 	   _pango_emoji_iter_next (&state.emoji_iter);
// 	   state.changed |= EMOJI_CHANGED;

// 	   if (state.emoji_iter.is_emoji)
// 		 state.width_iter.end = MAX (state.width_iter.end, state.emoji_iter.end);
// 	 }
//    if (state.run_end == state.width_iter.end)
// 	 {
// 	   width_iter_next (&state.width_iter);
// 	   state.changed |= WIDTH_CHANGED;
// 	 }

//    update_end (state);

//    return true;
//  }

//  static GSList *
//  copy_attr_slist (GSList *attr_slist)
//  {
//    GSList *new_list = nil;
//    GSList *l;

//    for (l = attr_slist; l; l = l.next)
// 	 new_list = g_slist_prepend (new_list, pango_attribute_copy (l.data));

//    return g_slist_reverse (new_list);
//  }

//  static void
//  itemize_state_fill_font (ItemizeState *state,
// 			  PangoFont    *font)
//  {
//    GList *l;

//    for (l = state.result; l; l = l.next)
// 	 {
// 	   PangoItem *item = l.data;
// 	   if (item.analysis.font)
// 		 break;
// 	   if (font)
// 	 item.analysis.font = g_object_ref (font);
// 	 }
//  }

//  static void
//  itemize_state_add_character (ItemizeState *state,
// 				  PangoFont    *font,
// 				  bool      force_break,
// 				  const char   *pos)
//  {
//    if (state.item)
// 	 {
// 	   if (!state.item.analysis.font && font)
// 	 {
// 	   itemize_state_fill_font (state, font);
// 	 }
// 	   else if (state.item.analysis.font && !font)
// 	 {
// 	   font = state.item.analysis.font;
// 	 }

// 	   if (!force_break &&
// 	   state.item.analysis.font == font)
// 	 {
// 	   state.item.num_chars++;
// 	   return;
// 	 }

// 	   state.item.length = (pos - state.text) - state.item.offset;
// 	 }

//    state.item = pango_item_new ();
//    state.item.offset = pos - state.text;
//    state.item.length = 0;
//    state.item.num_chars = 1;

//    if (font)
// 	 g_object_ref (font);
//    state.item.analysis.font = font;

//    state.item.analysis.level = state.embedding;
//    state.item.analysis.gravity = state.resolved_gravity;

//    /* The level vs. gravity dance:
// 	*	- If gravity is SOUTH, leave level untouched.
// 	*	- If gravity is NORTH, step level one up, to
// 	*	  not get mirrored upside-down text.
// 	*	- If gravity is EAST, step up to an even level, as
// 	*	  it's a clockwise-rotated layout, so the rotated
// 	*	  top is unrotated left.
// 	*	- If gravity is WEST, step up to an odd level, as
// 	*	  it's a counter-clockwise-rotated layout, so the rotated
// 	*	  top is unrotated right.
// 	*
// 	* A similar dance is performed in pango-layout.c:
// 	* line_set_resolved_dir().  Keep in synch.
// 	*/
//    switch (state.item.analysis.gravity)
// 	 {
// 	   case PANGO_GRAVITY_SOUTH:
// 	   default:
// 	 break;
// 	   case PANGO_GRAVITY_NORTH:
// 	 state.item.analysis.level++;
// 	 break;
// 	   case PANGO_GRAVITY_EAST:
// 	 state.item.analysis.level += 1;
// 	 state.item.analysis.level &= ~1;
// 	 break;
// 	   case PANGO_GRAVITY_WEST:
// 	 state.item.analysis.level |= 1;
// 	 break;
// 	 }

//    state.item.analysis.flags = state.centered_baseline ? PANGO_ANALYSIS_FLAG_CENTERED_BASELINE : 0;

//    state.item.analysis.script = state.script;
//    state.item.analysis.language = state.derived_lang;

//    if (state.copy_extra_attrs)
// 	 {
// 	   state.item.analysis.extra_attrs = copy_attr_slist (state.extra_attrs);
// 	 }
//    else
// 	 {
// 	   state.item.analysis.extra_attrs = state.extra_attrs;
// 	   state.copy_extra_attrs = true;
// 	 }

//    state.result = g_list_prepend (state.result, state.item);
//  }

//  typedef struct {
//    PangoLanguage *lang;
//    gunichar wc;
//    PangoFont *font;
//  } GetFontInfo;

//  static bool
//  get_font_foreach (PangoFontset *fontset,
// 		   PangoFont    *font,
// 		   gpointer      data)
//  {
//    GetFontInfo *info = data;

//    if (G_UNLIKELY (!font))
// 	 return false;

//    if (pango_font_has_char (font, info.wc))
// 	 {
// 	   info.font = font;
// 	   return true;
// 	 }

//    if (!fontset)
// 	 {
// 	   info.font = font;
// 	   return true;
// 	 }

//    return false;
//  }

//  static PangoFont *
//  get_base_font (ItemizeState *state)
//  {
//    if (!state.base_font)
// 	 state.base_font = pango_font_map_load_font (state.context.font_map,
// 						  state.context,
// 						  state.font_desc);
//    return state.base_font;
//  }

//  static bool
//  get_font (ItemizeState  *state,
// 		   gunichar       wc,
// 		   PangoFont    **font)
//  {
//    GetFontInfo info;

//    /* We'd need a separate cache when fallback is disabled, but since lookup
// 	* with fallback disabled is faster anyways, we just skip caching */
//    if (state.enable_fallback && font_cache_get (state.cache, wc, font))
// 	 return true;

//    info.lang = state.derived_lang;
//    info.wc = wc;
//    info.font = nil;

//    if (state.enable_fallback)
// 	 pango_fontset_foreach (state.current_fonts, get_font_foreach, &info);
//    else
// 	 get_font_foreach (nil, get_base_font (state), &info);

//    *font = info.font;

//    /* skip caching if fallback disabled (see above) */
//    if (state.enable_fallback)
// 	 font_cache_insert (state.cache, wc, *font);

//    return true;
//  }

//  static PangoLanguage *
//  compute_derived_language (PangoLanguage *lang,
// 			   PangoScript    script)
//  {
//    PangoLanguage *derived_lang;

//    /* Make sure the language tag is consistent with the derived
// 	* script. There is no point in marking up a section of
// 	* Arabic text with the "en" language tag.
// 	*/
//    if (lang && pango_language_includes_script (lang, script))
// 	 derived_lang = lang;
//    else
// 	 {
// 	   derived_lang = pango_script_get_sample_language (script);
// 	   /* If we don't find a sample language for the script, we
// 		* use a language tag that shouldn't actually be used
// 		* anywhere. This keeps fontconfig (for the PangoFc*
// 		* backend) from using the language tag to affect the
// 		* sort order. I don't have a reference for 'xx' being
// 		* safe here, though Keith Packard claims it is.
// 		*/
// 	   if (!derived_lang)
// 	 derived_lang = pango_language_from_string ("xx");
// 	 }

//    return derived_lang;
//  }

//  static void
//  itemize_state_update_for_new_run (ItemizeState *state)
//  {
//    /* This block should be moved to update_attr_iterator, but I'm too lazy to
// 	* do it right now */
//    if (state.changed & (FONT_CHANGED | SCRIPT_CHANGED | WIDTH_CHANGED))
// 	 {
// 	   /* Font-desc gravity overrides everything */
// 	   if (state.font_desc_gravity != PANGO_GRAVITY_AUTO)
// 	 {
// 	   state.resolved_gravity = state.font_desc_gravity;
// 	 }
// 	   else
// 	 {
// 	   PangoGravity gravity = state.gravity;
// 	   PangoGravityHint gravity_hint = state.gravity_hint;

// 	   if (G_LIKELY (gravity == PANGO_GRAVITY_AUTO))
// 		 gravity = state.context.resolved_gravity;

// 	   state.resolved_gravity = pango_gravity_get_for_script_and_width (state.script,
// 										 state.width_iter.upright,
// 										 gravity,
// 										 gravity_hint);
// 	 }

// 	   if (state.font_desc_gravity != state.resolved_gravity)
// 	 {
// 	   pango_font_description_set_gravity (state.font_desc, state.resolved_gravity);
// 	   state.changed |= FONT_CHANGED;
// 	 }
// 	 }

//    if (state.changed & (SCRIPT_CHANGED | LANG_CHANGED))
// 	 {
// 	   PangoLanguage *old_derived_lang = state.derived_lang;
// 	   state.derived_lang = compute_derived_language (state.lang, state.script);
// 	   if (old_derived_lang != state.derived_lang)
// 	 state.changed |= DERIVED_LANG_CHANGED;
// 	 }

//    if (state.changed & (EMOJI_CHANGED))
// 	 {
// 	   state.changed |= FONT_CHANGED;
// 	 }

//    if (state.changed & (FONT_CHANGED | DERIVED_LANG_CHANGED) &&
// 	   state.current_fonts)
// 	 {
// 	   g_object_unref (state.current_fonts);
// 	   state.current_fonts = nil;
// 	   state.cache = nil;
// 	 }

//    if (!state.current_fonts)
// 	 {
// 	   bool is_emoji = state.emoji_iter.is_emoji;
// 	   if (is_emoji && !state.emoji_font_desc)
// 	   {
// 		 state.emoji_font_desc = pango_font_description_copy_static (state.font_desc);
// 		 pango_font_description_set_family_static (state.emoji_font_desc, "emoji");
// 	   }
// 	   state.current_fonts = pango_font_map_load_fontset (state.context.font_map,
// 							   state.context,
// 							   is_emoji ? state.emoji_font_desc : state.font_desc,
// 							   state.derived_lang);
// 	   state.cache = get_font_cache (state.current_fonts);
// 	 }

//    if ((state.changed & FONT_CHANGED) && state.base_font)
// 	 {
// 	   g_object_unref (state.base_font);
// 	   state.base_font = nil;
// 	 }
//  }

//  static void
//  itemize_state_process_run (ItemizeState *state)
//  {
//    const char *p;
//    bool last_was_forced_break = false;

//    /* Only one character has type G_UNICODE_LINE_SEPARATOR in Unicode 4.0;
// 	* update this if that changes. */
//  #define LINE_SEPARATOR 0x2028

//    itemize_state_update_for_new_run (state);

//    /* We should never get an empty run */
//    g_assert (state.run_end != state.run_start);

//    for (p = state.run_start;
// 		p < state.run_end;
// 		p = g_utf8_next_char (p))
// 	 {
// 	   gunichar wc = g_utf8_get_char (p);
// 	   bool is_forced_break = (wc == '\t' || wc == LINE_SEPARATOR);
// 	   PangoFont *font;
// 	   GUnicodeType type;

// 	   /* We don't want space characters to affect font selection; in general,
// 		* it's always wrong to select a font just to render a space.
// 		* We assume that all fonts have the ASCII space, and for other space
// 		* characters if they don't, HarfBuzz will compatibility-decompose them
// 		* to ASCII space...
// 		* See bugs #355987 and #701652.
// 		*
// 		* We don't want to change fonts just for variation selectors.
// 		* See bug #781123.
// 		*
// 		* Finally, don't change fonts for line or paragraph separators.
// 		*/
// 	   type = g_unichar_type (wc);
// 	   if (G_UNLIKELY (type == G_UNICODE_CONTROL ||
// 					   type == G_UNICODE_FORMAT ||
// 					   type == G_UNICODE_SURROGATE ||
// 					   type == G_UNICODE_LINE_SEPARATOR ||
// 					   type == G_UNICODE_PARAGRAPH_SEPARATOR ||
// 					   (type == G_UNICODE_SPACE_SEPARATOR && wc != 0x1680u /* OGHAM SPACE MARK */) ||
// 					   (wc >= 0xfe00u && wc <= 0xfe0fu) ||
// 					   (wc >= 0xe0100u && wc <= 0xe01efu)))
// 		 {
// 	   font = nil;
// 		 }
// 	   else
// 		 {
// 	   get_font (state, wc, &font);
// 	 }

// 	   itemize_state_add_character (state, font,
// 					is_forced_break || last_was_forced_break,
// 					p);

// 	   last_was_forced_break = is_forced_break;
// 	 }

//    /* Finish the final item from the current segment */
//    state.item.length = (p - state.text) - state.item.offset;
//    if (!state.item.analysis.font)
// 	 {
// 	   PangoFont *font;

// 	   if (G_UNLIKELY (!get_font (state, ' ', &font)))
// 		 {
// 		   /* If no font was found, warn once per fontmap/script pair */
// 		   PangoFontMap *fontmap = state.context.font_map;
// 		   char *script_tag = g_strdup_printf ("g-unicode-script-%d", state.script);

// 		   if (!g_object_get_data (G_OBJECT (fontmap), script_tag))
// 			 {
// 			   g_warning ("failed to choose a font, expect ugly output. script='%d'",
// 						  state.script);

// 			   g_object_set_data_full (G_OBJECT (fontmap), script_tag,
// 									   GINT_TO_POINTER (1), nil);
// 			 }

// 		   g_free (script_tag);

// 		   font = nil;
// 		 }
// 	   itemize_state_fill_font (state, font);
// 	 }
//    state.item = nil;
//  }

//  static void
//  itemize_state_finish (ItemizeState *state)
//  {
//    g_free (state.embedding_levels);
//    if (state.free_attr_iter)
// 	 pango_attr_iterator_destroy (state.attr_iter);
//    _pango_script_iter_fini (&state.script_iter);
//    pango_font_description_free (state.font_desc);
//    pango_font_description_free (state.emoji_font_desc);
//    width_iter_fini (&state.width_iter);
//    _pango_emoji_iter_fini (&state.emoji_iter);

//    if (state.current_fonts)
// 	 g_object_unref (state.current_fonts);
//    if (state.base_font)
// 	 g_object_unref (state.base_font);
//  }

//  static GList *
//  itemize_with_font (Context               *context,
// 			const char                 *text,
// 			int                         start_index,
// 			int                         length,
// 			const PangoFontDescription *desc)
//  {
//    ItemizeState state;

//    if (length == 0)
// 	 return nil;

//    itemize_state_init (&state, context, text, context.base_dir, start_index, length,
// 			   nil, nil, desc);

//    do
// 	 itemize_state_process_run (&state);
//    for (itemize_state_next (&state));

//    itemize_state_finish (&state);

//    return g_list_reverse (state.result);
//  }

//  /**
//   * pango_itemize:
//   * `context`:   a structure holding information that affects
// 			the itemization process.
//   * @text:      the text to itemize. Must be valid UTF-8
//   * @start_index: first byte in @text to process
//   * @length:    the number of bytes (not characters) to process
//   *             after @start_index.
//   *             This must be >= 0.
//   * @attrs:     the set of attributes that apply to @text.
//   * @cached_iter: (allow-none): Cached attribute iterator, or %nil
//   *
//   * Breaks a piece of text into segments with consistent
//   * directional level and shaping engine. Each byte of @text will
//   * be contained in exactly one of the items in the returned list;
//   * the generated list of items will be in logical order (the start
//   * offsets of the items are ascending).
//   *
//   * @cached_iter should be an iterator over @attrs currently positioned at a
//   * range before or containing @start_index; @cached_iter will be advanced to
//   * the range covering the position just after @start_index + @length.
//   * (i.e. if itemizing in a loop, just keep passing in the same @cached_iter).
//   *
//   * Return value: (transfer full) (element-type Pango.Item): a #GList of #PangoItem
//   *               structures. The items should be freed using pango_item_free()
//   *               probably in combination with g_list_foreach(), and the list itself
//   *               using g_list_free().
//   */
//  GList *
//  pango_itemize (Context      *context,
// 			const char        *text,
// 			int                start_index,
// 			int                length,
// 			PangoAttrList     *attrs,
// 			PangoAttrIterator *cached_iter)
//  {
//    g_return_val_if_fail (context != nil, nil);
//    g_return_val_if_fail (start_index >= 0, nil);
//    g_return_val_if_fail (length >= 0, nil);
//    g_return_val_if_fail (length == 0 || text != nil, nil);

//    return pango_itemize_with_base_dir (context, context.base_dir,
// 					   text, start_index, length, attrs, cached_iter);
//  }

//  static bool
//  get_first_metrics_foreach (PangoFontset  *fontset,
// 				PangoFont     *font,
// 				gpointer       data)
//  {
//    PangoFontMetrics *fontset_metrics = data;
//    PangoLanguage *language = PANGO_FONTSET_GET_CLASS (fontset).get_language (fontset);
//    PangoFontMetrics *font_metrics = pango_font_get_metrics (font, language);
//    guint save_ref_count;

//    /* Initialize the fontset metrics to metrics of the first font in the
// 	* fontset; saving the refcount and restoring it is a bit of hack but avoids
// 	* having to update this code for each metrics addition.
// 	*/
//    save_ref_count = fontset_metrics.ref_count;
//    *fontset_metrics = *font_metrics;
//    fontset_metrics.ref_count = save_ref_count;

//    pango_font_metrics_unref (font_metrics);

//    return true;			/* Stops iteration */
//  }

//  static PangoFontMetrics *
//  get_base_metrics (PangoFontset *fontset)
//  {
//    PangoFontMetrics *metrics = pango_font_metrics_new ();

//    /* Initialize the metrics from the first font in the fontset */
//    pango_fontset_foreach (fontset, get_first_metrics_foreach, metrics);

//    return metrics;
//  }

//  static void
//  update_metrics_from_items (PangoFontMetrics *metrics,
// 				PangoLanguage    *language,
// 				const char       *text,
// 				unsigned int      text_len,
// 				GList            *items)

//  {
//    GHashTable *fonts_seen = g_hash_table_new (nil, nil);
//    PangoGlyphString *glyphs = pango_glyph_string_new ();
//    GList *l;
//    glong text_width;

//    /* This should typically be called with a sample text string. */
//    g_return_if_fail (text_len > 0);

//    metrics.approximate_char_width = 0;

//    for (l = items; l; l = l.next)
// 	 {
// 	   PangoItem *item = l.data;
// 	   PangoFont *font = item.analysis.font;

// 	   if (font != nil && g_hash_table_lookup (fonts_seen, font) == nil)
// 	 {
// 	   PangoFontMetrics *raw_metrics = pango_font_get_metrics (font, language);
// 	   g_hash_table_insert (fonts_seen, font, font);

// 	   /* metrics will already be initialized from the first font in the fontset */
// 	   metrics.ascent = MAX (metrics.ascent, raw_metrics.ascent);
// 	   metrics.descent = MAX (metrics.descent, raw_metrics.descent);
// 	   metrics.height = MAX (metrics.height, raw_metrics.height);
// 	   pango_font_metrics_unref (raw_metrics);
// 	 }

// 	   pango_shape_full (text + item.offset, item.length,
// 			 text, text_len,
// 			 &item.analysis, glyphs);
// 	   metrics.approximate_char_width += pango_glyph_string_get_width (glyphs);
// 	 }

//    pango_glyph_string_free (glyphs);
//    g_hash_table_destroy (fonts_seen);

//    text_width = pango_utf8_strwidth (text);
//    g_assert (text_width > 0);
//    metrics.approximate_char_width /= text_width;
//  }

//  /**
//   * pango_context_get_metrics:
//   * `context`: a #Context
//   * @desc: (allow-none): a #PangoFontDescription structure.  %nil means that the
//   *            font description from the context will be used.
//   * @language: (allow-none): language tag used to determine which script to get
//   *            the metrics for. %nil means that the language tag from the context
//   *            will be used. If no language tag is set on the context, metrics
//   *            for the default language (as determined by pango_language_get_default())
//   *            will be returned.
//   *
//   * Get overall metric information for a particular font
//   * description.  Since the metrics may be substantially different for
//   * different scripts, a language tag can be provided to indicate that
//   * the metrics should be retrieved that correspond to the script(s)
//   * used by that language.
//   *
//   * The #PangoFontDescription is interpreted in the same way as
//   * by pango_itemize(), and the family name may be a comma separated
//   * list of figures. If characters from multiple of these families
//   * would be used to render the string, then the returned fonts would
//   * be a composite of the metrics for the fonts loaded for the
//   * individual families.
//   *
//   * Return value: a #PangoFontMetrics object. The caller must call pango_font_metrics_unref()
//   *   when finished using the object.
//   **/
//  PangoFontMetrics *
//  pango_context_get_metrics (Context                 *context,
// 				const PangoFontDescription   *desc,
// 				PangoLanguage                *language)
//  {
//    PangoFontset *current_fonts = nil;
//    PangoFontMetrics *metrics;
//    const char *sample_str;
//    unsigned int text_len;
//    GList *items;

//    g_return_val_if_fail (PANGO_IS_CONTEXT (context), nil);

//    if (!desc)
// 	 desc = context.font_desc;

//    if (!language)
// 	 language = context.language;

//    current_fonts = pango_font_map_load_fontset (context.font_map, context, desc, language);
//    metrics = get_base_metrics (current_fonts);

//    sample_str = pango_language_get_sample_string (language);
//    text_len = strlen (sample_str);
//    items = itemize_with_font (context, sample_str, 0, text_len, desc);

//    update_metrics_from_items (metrics, language, sample_str, text_len, items);

//    g_list_foreach (items, (GFunc)pango_item_free, nil);
//    g_list_free (items);

//    g_object_unref (current_fonts);

//    return metrics;
//  }

//  static void
//  context_changed  (Context *context)
//  {
//    context.serial++;
//    if (context.serial == 0)
// 	 context.serial++;
//  }

//  /**
//   * pango_context_changed:
//   * `context`: a #Context
//   *
//   * Forces a change in the context, which will cause any Layout
//   * using this context to re-layout.
//   *
//   * This function is only useful when implementing a new backend
//   * for Pango, something applications won't do. Backends should
//   * call this function if they have attached extra data to the context
//   * and such data is changed.
//   *
//   * Since: 1.32.4
//   **/
//  void
//  pango_context_changed  (Context *context)
//  {
//    context_changed (context);
//  }

//  static void
//  check_fontmap_changed (Context *context)
//  {
//    guint old_serial = context.fontmap_serial;

//    if (!context.font_map)
// 	 return;

//    context.fontmap_serial = pango_font_map_get_serial (context.font_map);

//    if (old_serial != context.fontmap_serial)
// 	 context_changed (context);
//  }

// Returns the current serial number of `context`.  The serial number is
// initialized to an small number larger than zero when a new context
// is created and is increased whenever the context is changed using any
// of the setter functions, or the #PangoFontMap it uses to find fonts has
// changed. The serial may wrap, but will never have the value 0. Since it
// can wrap, never compare it with "less than", always use "not equals".
//
// This can be used to automatically detect changes to a #Context, and
// is only useful when implementing objects that need update when their
// #Context changes, like Layout.
func (context *Context) pango_context_get_serial() uint {
	context.check_fontmap_changed()
	return context.serial
}

func (context *Context) check_fontmap_changed() {} // TODO:

//  /**
//  // pango_context_set_round_glyph_positions:
//   * `context`: a #Context
//   * @round_positions: whether to round glyph positions
//   *
//   * Sets whether font rendering with this context should
//   * round glyph positions and widths to integral positions,
//   * in device units.
//   *
//   * This is useful when the renderer can't handle subpixel
//   * positioning of glyphs.
//   *
//   * The default value is to round glyph positions, to remain
//   * compatible with previous Pango behavior.
//   *
//   * Since: 1.44
//   */
//  void
//  pango_context_set_round_glyph_positions (Context *context,
// 										  bool      round_positions)
//  {
//    if (context.round_glyph_positions != round_positions)
// 	 {
// 	   context.round_glyph_positions = round_positions;
// 	   context_changed (context);
// 	 }
//  }

//  /**
//   * pango_context_get_round_glyph_positions:
//   * `context`: a #Context
//   *
//   * Returns whether font rendering with this context should
//   * round glyph positions and widths.
//   *
//   * Since: 1.44
//   */
//  bool
//  pango_context_get_round_glyph_positions (Context *context)
//  {
//    return context.round_glyph_positions;
//  }