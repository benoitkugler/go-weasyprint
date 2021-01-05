package pango

import (
	"container/list"
	"strings"

	"github.com/benoitkugler/go-weasyprint/pango/fontconfig"
)

// pangofc-fontmap.c: Base fontmap type for fontconfig-based backends

/*
 * PangoFcFontMap is a base class for font map implementations using the
 * Fontconfig and FreeType libraries. It is used in the
 * <link linkend="pango-Xft-Fonts-and-Rendering">Xft</link> and
 * <link linkend="pango-FreeType-Fonts-and-Rendering">FreeType</link>
 * backends shipped with Pango, but can also be used when creating
 * new backends. Any backend deriving from this base class will
 * take advantage of the wide range of shapers implemented using
 * FreeType that come with Pango.
 */

const FONTSET_CACHE_SIZE = 256

/* Overview:
 *
 * All programming is a practice in caching data. PangoFcFontMap is the
 * major caching container of a Pango system on a Linux desktop. Here is
 * a short overview of how it all works.
 *
 * In short, Fontconfig search patterns are constructed and a fontset loaded
 * using them. Here is how we achieve that:
 *
 * - All FcPattern's referenced by any object in the fontmap are uniquified
 *   and cached in the fontmap. This both speeds lookups based on patterns
 *   faster, and saves memory. This is handled by fontmap.priv.pattern_hash.
 *   The patterns are cached indefinitely.
 *
 * - The results of a FcFontSort() are used to populate fontsets.  However,
 *   FcFontSort() relies on the search pattern only, which includes the font
 *   size but not the full font matrix.  The fontset however depends on the
 *   matrix.  As a result, multiple fontsets may need results of the
 *   FcFontSort() on the same input pattern (think rotating text).  As such,
 *   we cache FcFontSort() results in fontmap.priv.patterns_hash which
 *   is a refcounted structure.  This level of abstraction also allows for
 *   optimizations like calling FcFontMatch() instead of FcFontSort(), and
 *   only calling FcFontSort() if any patterns other than the first match
 *   are needed.  Another possible optimization would be to call FcFontSort()
 *   without trimming, and do the trimming lazily as we go.  Only pattern sets
 *   already referenced by a fontset are cached.
 *
 * - A number of most-recently-used fontsets are cached and reused when
 *   needed.  This is achieved using fontmap.priv.fontset_hash and
 *   fontmap.priv.fontset_cache.
 *
 * - All fonts created by any of our fontsets are also cached and reused.
 *   This is what fontmap.priv.font_hash does.
 *
 * - Data that only depends on the font file and face index is cached and
 *   reused by multiple fonts.  This includes coverage and cmap cache info.
 *   This is done using fontmap.priv.font_face_data_hash.
 *
 * Upon a cache_clear() request, all caches are emptied.  All objects (fonts,
 * fontsets, faces, families) having a reference from outside will still live
 * and may reference the fontmap still, but will not be reused by the fontmap.
 *
 *
 * Todo:
 *
 * - Make PangoCoverage a GObject and subclass it as PangoFcCoverage which
 *   will directly use FcCharset. (#569622)
 *
 * - Lazy trimming of FcFontSort() results.  Requires fontconfig with
 *   FcCharSetMerge().
 */

/**
 * String representing a fontconfig property name that Pango sets on any
 * fontconfig pattern it passes to fontconfig if a `Gravity` other
 * than PANGO_GRAVITY_SOUTH is desired.
 *
 * The property will have a `Gravity` value as a string, like "east".
 * This can be used to write fontconfig configuration rules to choose
 * different fonts for horizontal and vertical writing directions.
 */
const PANGO_FC_GRAVITY fontconfig.FcObject = "pangogravity"

type PangoCairoFcFontMap struct {
	parent_instance PangoFcFontMap

	serial uint
	dpi    float64
}

// PangoFcFontMap is a base class for font map implementations
// using the Fontconfig and FreeType libraries. To create a new
// backend using Fontconfig and FreeType, you derive from this class
// and implement a new_font() virtual function that creates an
// instance deriving from PangoFcFont.
type PangoFcFontMap struct {
	parent_instance FontMap

	priv *PangoFcFontMapPrivate

	// Function to call on prepared patterns to do final config tweaking.
	// substitute_func    PangoFcSubstituteFunc
	// substitute_data    gpointer
	// substitute_destroy GDestroyNotify

	// TODO: check the design of C "class"
	context_key_get        func(*Context) int
	fontset_key_substitute func(*PangoFcFontsetKey, *fontconfig.FcPattern)
	default_substitute     func(*fontconfig.FcPattern)
}

/**
 * PangoFcFont:
 *
 * #PangoFcFont is a base class for font implementations
 * using the Fontconfig and FreeType libraries and is used in
 * conjunction with #PangoFcFontMap. When deriving from this
 * class, you need to implement all of its virtual functions
 * other than shutdown() along with the get_glyph_extents()
 * virtual function from #PangoFont.
 **/
type PangoFcFont struct {
	parent_instance Font

	font_pattern *fontconfig.FcPattern /* fully resolved pattern */
	fontmap      FontMap               /* associated map */
	// priv         gpointer      /* used internally */
	matrix      Matrix /* used internally */
	description FontDescription

	metrics_by_lang []interface{}

	is_hinted      bool //  = 1;
	is_transformed bool //  = 1;
}

//  #define PANGO_FC_TYPE_FAMILY            (pango_fc_family_get_type ())
//  #define PANGO_FC_FAMILY(object)         (G_TYPE_CHECK_INSTANCE_CAST ((object), PANGO_FC_TYPE_FAMILY, PangoFcFamily))
//  #define PANGO_FC_IS_FAMILY(object)      (G_TYPE_CHECK_INSTANCE_TYPE ((object), PANGO_FC_TYPE_FAMILY))

//  #define PANGO_FC_TYPE_FACE              (pango_fc_face_get_type ())
//  #define PANGO_FC_FACE(object)           (G_TYPE_CHECK_INSTANCE_CAST ((object), PANGO_FC_TYPE_FACE, PangoFcFace))
//  #define PANGO_FC_IS_FACE(object)        (G_TYPE_CHECK_INSTANCE_TYPE ((object), PANGO_FC_TYPE_FACE))

//  #define PANGO_FC_TYPE_FONTSET           (pango_fc_fontset_get_type ())
//  #define PANGO_FC_FONTSET(object)        (G_TYPE_CHECK_INSTANCE_CAST ((object), PANGO_FC_TYPE_FONTSET, PangoFcFontset))
//  #define PANGO_FC_IS_FONTSET(object)     (G_TYPE_CHECK_INSTANCE_TYPE ((object), PANGO_FC_TYPE_FONTSET))

type PangoFcFontMapPrivate struct {
	fontset_hash  fontsetHash
	fontset_cache *list.List // *PangoFcFontset /* Recently used fontsets */

	font_hash fontHash

	patterns_hash map[*fontconfig.FcPattern]*PangoFcPatterns

	// pattern_hash is used to make sure we only store one copy of each identical pattern. (Speeds up lookup).
	pattern_hash fcPatternHash

	font_face_data_hash map[PangoFcFontFaceData]bool /* Maps font file name/id . data */ // (GHashFunc)pango_fc_font_face_data_hash,  (GEqualFunc)pango_fc_font_face_data_equal,

	/* List of all families availible */
	families [][]PangoFcFamily
	//    int n_families;		/* -1 == uninitialized */

	dpi float64

	/* Decoders */
	// GSList *findfuncs

	closed bool // = 1;

	config *fontconfig.FcConfig
}

type PangoFcFontFaceData struct {
	/* Key */
	filename string
	id       int /* needed to handle TTC files with multiple faces */

	//    /* Data */
	//    FcPattern *pattern;  /* Referenced pattern that owns filename */
	//    PangoCoverage *coverage;
	//    PangoLanguage **languages;

	//    hb_face_t *hb_face;
}

type PangoFcFace struct {
	parent_instance FontFace

	family  *PangoFcFamily
	style   string
	pattern fontconfig.FcPattern

	fake    bool // = 1;
	regular bool // = 1;
}

type PangoFcFamily struct {
	parent_instance FontFamily

	fontmap     *PangoFcFontMap
	family_name string

	patterns *fontconfig.FcFontSet
	faces    []*PangoFcFace
	// int      n_faces /* -1 == uninitialized */

	spacing  int /* FC_SPACING */
	variable bool
}

//  type  PangoFcFindFuncInfo struct
//  {
//    PangoFcDecoderFindFunc findfunc;
//    gpointer               user_data;
//    GDestroyNotify         dnotify;
//    gpointer               ddata;
//  };

//  static FcPattern *uniquifyPattern ( fcfontmap *PangoFcFontMap,
// 					 pattern *FcPattern      );

//  gpointer get_gravity_class (void);

//  gpointer
//  get_gravity_class (void)
//  {
//    static GEnumClass *class = nil; /* MT-safe */

//    if (g_once_init_enter (&class))
// 	 g_once_init_leave (&class, (gpointer)g_type_class_ref (PANGO_TYPE_GRAVITY));

//    return class;
//  }

//  static guint
//  pango_fc_font_face_data_hash (PangoFcFontFaceData *key)
//  {
//    return g_str_hash (key.filename) ^ key.id;
//  }

//  static bool
//  pango_fc_font_face_data_equal (PangoFcFontFaceData *key1,
// 					PangoFcFontFaceData *key2)
//  {
//    return key1.id == key2.id &&
// 	  (key1 == key2 || 0 == strcmp (key1.filename, key2.filename));
//  }

//  static void
//  pango_fc_font_face_data_free (PangoFcFontFaceData *data)
//  {
//    FcPatternDestroy (data.pattern);

//    if (data.coverage)
// 	 pango_coverage_unref (data.coverage);

//    g_free (data.languages);

//    hb_face_destroy (data.hb_face);

//    g_slice_free (PangoFcFontFaceData, data);
//  }

//  /* Fowler / Noll / Vo (FNV) Hash (http://www.isthe.com/chongo/tech/comp/fnv/)
//   *
//   * Not necessarily better than a lot of other hashes, but should be OK, and
//   * well tested with binary data.
//   */

//  #define FNV_32_PRIME ((guint32)0x01000193)
//  #define FNV1_32_INIT ((guint32)0x811c9dc5)

//  static guint32
//  hash_bytes_fnv (unsigned char *buffer,
// 		 int            len,
// 		 guint32        hval)
//  {
//    while (len--)
// 	 {
// 	   hval *= FNV_32_PRIME;
// 	   hval ^= *buffer++;
// 	 }

//    return hval;
//  }

func (fcfontmap *PangoFcFontMap) getScaledSize(context *Context, desc *FontDescription) int {
	size := float64(desc.size)

	if !desc.size_is_absolute {
		dpi := fcfontmap.pango_fc_font_map_get_resolution(context)

		size = size * dpi / 72.
	}

	return int(.5 + context.matrix.pango_matrix_get_font_scale_factor()*size)
}

type PangoFcFontsetKey struct {
	// fontmap     *PangoFcFontMap
	language    Language
	desc        FontDescription
	matrix      Matrix
	pixelsize   int
	resolution  float64
	context_key int
	variations  string
}

func (key *PangoFcFontsetKey) pango_fc_fontset_key_make_pattern() *fontconfig.FcPattern {
	// pango_fc_make_pattern(key.desc, key.language, key.pixelsize, key.resolution, key.variations)

	//  pango_fc_make_pattern (const  FontDescription *description,
	// 				PangoLanguage               *language,
	// 				int                          pixel_size,
	// 				float64                       dpi,
	// 						const char                  *variations)

	//    FcPattern *pattern;
	//    const char *prgname;
	//    int slant;
	//    float64 weight;
	//    PangoGravity gravity;
	//    FcBool vertical;
	//    char **families;
	//    int i;
	//    int width;

	slant := pango_fc_convert_slant_to_fc(key.desc.style)
	weight := fontconfig.FcWeightFromOpenTypeDouble(float64(key.desc.weight))
	width := pango_fc_convert_width_to_fc(key.desc.stretch)

	gravity := key.desc.gravity
	vertical := fontconfig.FcFalse
	if gravity.isVertical() {
		vertical = fontconfig.FcTrue
	}

	/* The reason for passing in FC_SIZE as well as FC_PIXEL_SIZE is
	* to work around a bug in libgnomeprint where it doesn't look
	* for FC_PIXEL_SIZE. See http://bugzilla.gnome.org/show_bug.cgi?id=169020
	*
	* Putting FC_SIZE in here slightly reduces the efficiency
	* of caching of patterns and fonts when working with multiple different
	* dpi values.
	 */
	pattern := fontconfig.FcPatternBuild([]fontconfig.PatternElement{
		// {Object: PANGO_FC_VERSION, Value: pango_version()},       // FcTypeInteger
		{Object: fontconfig.FC_WEIGHT, Value: weight},            // FcTypeDouble
		{Object: fontconfig.FC_SLANT, Value: slant},              // FcTypeInteger
		{Object: fontconfig.FC_WIDTH, Value: width},              // FcTypeInteger
		{Object: fontconfig.FC_VERTICAL_LAYOUT, Value: vertical}, // FcTypeBool
		// {Object:fontconfig.FC_VARIABLE,Value:  FcDontCare}, //  FcTypeBool
		{Object: fontconfig.FC_DPI, Value: key.resolution},                                           // FcTypeDouble
		{Object: fontconfig.FC_SIZE, Value: float64(key.pixelsize) * (72. / 1024. / key.resolution)}, // FcTypeDouble
		{Object: fontconfig.FC_PIXEL_SIZE, Value: float64(key.pixelsize) / 1024.},                    // FcTypeDouble
	}...)

	if key.variations != "" {
		pattern.Add(fontconfig.FC_FONT_VARIATIONS, key.variations)
	}

	if key.desc.family_name != "" {
		families := strings.Split(key.desc.family_name, ",")
		for _, fam := range families {
			pattern.Add(fontconfig.FC_FAMILY, fam)
		}
	}

	if key.language != "" {
		pattern.Add(fontconfig.FC_LANG, string(key.language))
	}

	if gravity != PANGO_GRAVITY_SOUTH {
		pattern.Add(PANGO_FC_GRAVITY, gravity_map.toString("gravity", int(gravity)))
	}

	return pattern
}

type PangoFcFontKey struct {
	fontmap *PangoFcFontMap
	pattern *fontconfig.FcPattern
	matrix  Matrix
	// context_key gpointer
	variations string
}

func (fcfontmap *PangoFcFontMap) newFontsetKey(context *Context, desc *FontDescription, language Language) PangoFcFontsetKey {
	if language == "" && context != nil {
		language = context.set_language
	}

	var key PangoFcFontsetKey
	// key.fontmap = fcfontmap

	if context != nil && context.matrix != nil {
		key.matrix = *context.matrix
	} else {
		key.matrix = PANGO_MATRIX_INIT
	}
	key.matrix.x0, key.matrix.y0 = 0, 0

	key.pixelsize = fcfontmap.getScaledSize(context, desc)
	key.resolution = fcfontmap.pango_fc_font_map_get_resolution(context)
	key.language = language
	key.variations = desc.variations
	key.desc = *desc
	key.desc.pango_font_description_unset_fields(PANGO_FONT_MASK_SIZE | PANGO_FONT_MASK_VARIATIONS)

	if context != nil {
		key.context_key = fcfontmap.context_key_get(context)
	}
	return key
}

//  static bool
//  pango_fc_fontset_key_equal (const key *PangoFcFontsetKey_a,
// 				 const key *PangoFcFontsetKey_b)
//  {
//    if (key_a.language == key_b.language &&
// 	   key_a.pixelsize == key_b.pixelsize &&
// 	   key_a.resolution == key_b.resolution &&
// 	   ((key_a.variations == nil && key_b.variations == nil) ||
// 		(key_a.variations && key_b.variations && (strcmp (key_a.variations, key_b.variations) == 0))) &&
// 	   pango_font_description_equal (key_a.desc, key_b.desc) &&
// 	   0 == memcmp (&key_a.matrix, &key_b.matrix, 4 * sizeof (float64)))
// 	 {
// 	   if (key_a.context_key)
// 	 return PANGO_FC_FONT_MAP_GET_CLASS (key_a.fontmap).context_key_equal (key_a.fontmap,
// 										 key_a.context_key,
// 										 key_b.context_key);
// 	   else
// 		 return key_a.context_key == key_b.context_key;
// 	 }
//    else
// 	 return false;
//  }

//  static void
//  pango_fc_fontset_key_free (key *PangoFcFontsetKey)
//  {
//    pango_font_description_free (key.desc);
//    g_free (key.variations);

//    if (key.context_key)
// 	 PANGO_FC_FONT_MAP_GET_CLASS (key.fontmap).context_key_free (key.fontmap,
// 								   key.context_key);

//    g_slice_free (PangoFcFontsetKey, key);
//  }

//  /**
//   * pango_fc_fontset_key_get_language:
//   * @key: the fontset key
//   *
//   * Gets the language member of @key.
//   *
//   * Returns: the language
//   *
//   * Since: 1.24
//   **/
//  PangoLanguage *
//  pango_fc_fontset_key_get_language (const key *PangoFcFontsetKey)
//  {
//    return key.language;
//  }

//  /**
//   * pango_fc_fontset_key_get_description:
//   * @key: the fontset key
//   *
//   * Gets the font description of @key.
//   *
//   * Returns: the font description, which is owned by @key and should not be modified.
//   *
//   * Since: 1.24
//   **/
//  const FontDescription *
//  pango_fc_fontset_key_get_description (const key *PangoFcFontsetKey)
//  {
//    return key.desc;
//  }

//  /**
//   * pango_fc_fontset_key_get_matrix:
//   * @key: the fontset key
//   *
//   * Gets the matrix member of @key.
//   *
//   * Returns: the matrix, which is owned by @key and should not be modified.
//   *
//   * Since: 1.24
//   **/
//  const Matrix *
//  pango_fc_fontset_key_get_matrix      (const key *PangoFcFontsetKey)
//  {
//    return &key.matrix;
//  }

//  /**
//   * pango_fc_fontset_key_get_absolute_size:
//   * @key: the fontset key
//   *
//   * Gets the absolute font size of @key in Pango units.  This is adjusted
//   * for both resolution and transformation matrix.
//   *
//   * Returns: the pixel size of @key.
//   *
//   * Since: 1.24
//   **/
//  float64
//  pango_fc_fontset_key_get_absolute_size   (const key *PangoFcFontsetKey)
//  {
//    return key.pixelsize;
//  }

//  /**
//   * pango_fc_fontset_key_get_resolution:
//   * @key: the fontset key
//   *
//   * Gets the resolution of @key
//   *
//   * Returns: the resolution of @key
//   *
//   * Since: 1.24
//   **/
//  float64
//  pango_fc_fontset_key_get_resolution  (const key *PangoFcFontsetKey)
//  {
//    return key.resolution;
//  }

//  /**
//   * pango_fc_fontset_key_get_context_key:
//   * @key: the font key
//   *
//   * Gets the context key member of @key.
//   *
//   * Returns: the context key, which is owned by @key and should not be modified.
//   *
//   * Since: 1.24
//   **/
//  gpointer
//  pango_fc_fontset_key_get_context_key (const key *PangoFcFontsetKey)
//  {
//    return key.context_key;
//  }

//  /*
//   * PangoFcFontKey
//   */

//  static guint
//  pango_fc_font_key_hash (const PangoFcFontKey *key)
//  {
// 	 guint32 hash = FNV1_32_INIT;

// 	 /* We do a bytewise hash on the doubles */
// 	 hash = hash_bytes_fnv ((unsigned char *)(&key.matrix), sizeof (float64) * 4, hash);

// 	 if (key.variations)
// 	   hash ^= g_str_hash (key.variations);

// 	 if (key.context_key)
// 	   hash ^= PANGO_FC_FONT_MAP_GET_CLASS (key.fontmap).context_key_hash (key.fontmap,
// 										 key.context_key);

// 	 return (hash ^ GPOINTER_TO_UINT (key.pattern));
//  }

//  static void
//  pango_fc_font_key_free (PangoFcFontKey *key)
//  {
//    if (key.pattern)
// 	 FcPatternDestroy (key.pattern);

//    if (key.context_key)
// 	 PANGO_FC_FONT_MAP_GET_CLASS (key.fontmap).context_key_free (key.fontmap,
// 								   key.context_key);

//    g_free (key.variations);

//    g_slice_free (PangoFcFontKey, key);
//  }

//  static PangoFcFontKey *
//  pango_fc_font_key_copy (const PangoFcFontKey *old)
//  {
//    PangoFcFontKey *key = g_slice_new (PangoFcFontKey);

//    key.fontmap = old.fontmap;
//    FcPatternReference (old.pattern);
//    key.pattern = old.pattern;
//    key.matrix = old.matrix;
//    key.variations = g_strdup (old.variations);
//    if (old.context_key)
// 	 key.context_key = PANGO_FC_FONT_MAP_GET_CLASS (key.fontmap).context_key_copy (key.fontmap,
// 											  old.context_key);
//    else
// 	 key.context_key = nil;

//    return key;
//  }

//  static void
//  pango_fc_font_key_init (PangoFcFontKey    *key,
// 			 PangoFcFontMap    *fcfontmap,
// 			 PangoFcFontsetKey *fontset_key,
// 			 pattern fontconfig.FcPattern         )
//  {
//    key.fontmap = fcfontmap;
//    key.pattern = pattern;
//    key.matrix = *pango_fc_fontset_key_get_matrix (fontset_key);
//    key.variations = fontset_key.variations;
//    key.context_key = pango_fc_fontset_key_get_context_key (fontset_key);
//  }

//  /* Public API */

//  /**
//   * pango_fc_font_key_get_pattern:
//   * @key: the font key
//   *
//   * Gets the fontconfig pattern member of @key.
//   *
//   * Returns: the pattern, which is owned by @key and should not be modified.
//   *
//   * Since: 1.24
//   **/
//  const FcPattern *
//  pango_fc_font_key_get_pattern (const PangoFcFontKey *key)
//  {
//    return key.pattern;
//  }

//  /**
//   * pango_fc_font_key_get_matrix:
//   * @key: the font key
//   *
//   * Gets the matrix member of @key.
//   *
//   * Returns: the matrix, which is owned by @key and should not be modified.
//   *
//   * Since: 1.24
//   **/
//  const Matrix *
//  pango_fc_font_key_get_matrix (const PangoFcFontKey *key)
//  {
//    return &key.matrix;
//  }

//  /**
//   * pango_fc_font_key_get_context_key:
//   * @key: the font key
//   *
//   * Gets the context key member of @key.
//   *
//   * Returns: the context key, which is owned by @key and should not be modified.
//   *
//   * Since: 1.24
//   **/
//  gpointer
//  pango_fc_font_key_get_context_key (const PangoFcFontKey *key)
//  {
//    return key.context_key;
//  }

//  const char *
//  pango_fc_font_key_get_variations (const PangoFcFontKey *key)
//  {
//    return key.variations;
//  }

// ------------------------------- PangoFcPatterns -------------------------------

type PangoFcPatterns struct {
	fontmap *PangoFcFontMap

	pattern *fontconfig.FcPattern
	match   *fontconfig.FcPattern
	fontset fontconfig.FcFontSet
}

func (fontmap *PangoFcFontMap) pango_fc_patterns_new(pat *fontconfig.FcPattern) *PangoFcPatterns {
	pat = fontmap.uniquifyPattern(pat)

	if pats := fontmap.priv.patterns_hash[pat]; pats != nil {
		return pats
	}

	var pats PangoFcPatterns

	pats.fontmap = fontmap
	pats.pattern = pat
	fontmap.priv.patterns_hash[pats.pattern] = &pats

	return &pats
}

//  static PangoFcPatterns *
//  pango_fc_patterns_ref (pats *PangoFcPatterns)
//  {
//    g_return_val_if_fail (pats.ref_count > 0, nil);

//    pats.ref_count++;

//    return pats;
//  }

//  static void
//  pango_fc_patterns_unref (pats *PangoFcPatterns)
//  {
//    g_return_if_fail (pats.ref_count > 0);

//    pats.ref_count--;

//    if (pats.ref_count)
// 	 return;

//    /* Only remove from fontmap hash if we are in it.  This is not necessarily
// 	* the case after a cache_clear() call. */
//    if (pats.fontmap.priv.patterns_hash &&
// 	   pats == g_hash_table_lookup (pats.fontmap.priv.patterns_hash, pats.pattern))
// 	 g_hash_table_remove (pats.fontmap.priv.patterns_hash,
// 			  pats.pattern);

//    if (pats.pattern)
// 	 FcPatternDestroy (pats.pattern);

//    if (pats.match)
// 	 FcPatternDestroy (pats.match);

//    if (pats.fontset)
// 	 FcFontSetDestroy (pats.fontset);

//    g_slice_free (PangoFcPatterns, pats);
//  }

func pango_fc_is_supported_font_format(pattern *fontconfig.FcPattern) bool {
	fontformat, res := FcPatternGetString(pattern, FC_FONTFORMAT, 0)
	if res != fontconfig.FcResultMatch {
		return false
	}

	/* harfbuzz supports only SFNT fonts. */
	/* FIXME: "CFF" is used for both CFF in OpenType and bare CFF files, but
	* HarfBuzz does not support the later and FontConfig does not seem
	* to have a way to tell them apart.
	 */
	return fontformat == "TrueType" || fontformat == "CFF"
}

func filter_fontset_by_format(fontset fontconfig.FcFontSet) fontconfig.FcFontSet {
	var result fontconfig.FcFontSet

	for _, fontPattern := range fontset {
		if pango_fc_is_supported_font_format(fontPattern) {
			result = append(result, fontPattern)
		}
	}

	return result
}

func (pats *PangoFcPatterns) pango_fc_patterns_get_font_pattern(i int) (*fontconfig.FcPattern, bool) {
	if i == 0 {
		if pats.match == nil && pats.fontset == nil {
			pats.match, _ = FcFontMatch(pats.fontmap.priv.config, pats.pattern)
		}

		if pats.match != nil && pango_fc_is_supported_font_format(pats.match) {
			return pats.match, false
		}
	}

	if pats.fontset == nil {
		var (
			filtered [2]fontconfig.FcFontSet
			n        int
		)

		for i := range filtered {
			fonts := FcConfigGetFonts(pats.fontmap.priv.config, i)
			if fonts != nil {
				filtered[n] = filter_fontset_by_format(fonts)
				n++
			}
		}

		pats.fontset, _ = FcFontSetSort(pats.fontmap.priv.config, filtered, n, pats.pattern, FcTrue, nil)

		if pats.match != nil {
			pats.match = nil
		}
	}

	if i < len(pats.fontset) {
		return pats.fontset[i], true
	}
	return nil, true
}

/*
 * PangoFcFontset
 */

type PangoFcFontset struct {
	parent_instance Fontset

	key *PangoFcFontsetKey

	patterns   *PangoFcPatterns
	patterns_i int

	fonts     []Font
	coverages []Coverage

	cache_link *list.Element
}

//  typedef PangoFontsetClass PangoFcFontsetClass;

//  G_DEFINE_TYPE (PangoFcFontset, pango_fc_fontset, PANGO_TYPE_FONTSET)

func pango_fc_fontset_new(key PangoFcFontsetKey, patterns *PangoFcPatterns) *PangoFcFontset {
	var fontset PangoFcFontset

	fontset.key = &key
	fontset.patterns = patterns

	return &fontset
}

func (fontset *PangoFcFontset) pango_fc_fontset_load_next_font() Font {

	pattern := fontset.patterns.pattern
	font_pattern, prepare := fontset.patterns.pango_fc_patterns_get_font_pattern(fontset.patterns_i)
	fontset.patterns_i++
	if font_pattern == nil {
		return nil
	}

	if prepare {
		font_pattern = FcFontRenderPrepare(nil, pattern, font_pattern)
		if font_pattern == nil {
			return nil
		}
	}

	font := pango_fc_font_map_new_font(fontset.key.fontmap,
		fontset.key, font_pattern)

	if prepare {
		FcPatternDestroy(font_pattern)
	}

	return font
}

func (fontset *PangoFcFontset) pango_fc_fontset_get_font_at(i int) Font {
	for i >= len(fontset.fonts) {
		font := pango_fc_fontset_load_next_font(fontset)
		g_ptr_array_add(fontset.fonts, font)
		g_ptr_array_add(fontset.coverages, nil)
		if !font {
			return nil
		}
	}

	return g_ptr_array_index(fontset.fonts, i)
}

//  static void
//  pango_fc_fontset_class_init (PangoFcFontsetClass *class)
//  {
//    GObjectClass *object_class = G_OBJECT_CLASS (class);
//    PangoFontsetClass *fontset_class = PANGO_FONTSET_CLASS (class);

//    object_class.finalize = pango_fc_fontset_finalize;

//    fontset_class.get_font = pango_fc_fontset_get_font;
//    fontset_class.get_language = pango_fc_fontset_get_language;
//    fontset_class.foreach = pango_fc_fontset_foreach;
//  }

//  static void
//  pango_fc_fontset_init (fontset *PangoFcFontset)
//  {
//    fontset.fonts = g_ptr_array_new ();
//    fontset.coverages = g_ptr_array_new ();
//  }

//  static void
//  pango_fc_fontset_finalize (GObject *object)
//  {
//    fontset *PangoFcFontset = PANGO_FC_FONTSET (object);
//    unsigned int i;

//    for (i = 0; i < fontset.fonts.len; i++)
//    {
// 	 PangoFont *font = g_ptr_array_index(fontset.fonts, i);
// 	 if (font)
// 	   g_object_unref (font);
//    }
//    g_ptr_array_free (fontset.fonts, true);

//    for (i = 0; i < fontset.coverages.len; i++)
// 	 {
// 	   PangoCoverage *coverage = g_ptr_array_index (fontset.coverages, i);
// 	   if (coverage)
// 	 pango_coverage_unref (coverage);
// 	 }
//    g_ptr_array_free (fontset.coverages, true);

//    if (fontset.key)
// 	 pango_fc_fontset_key_free (fontset.key);

//    if (fontset.patterns)
// 	 pango_fc_patterns_unref (fontset.patterns);

//    G_OBJECT_CLASS (pango_fc_fontset_parent_class).finalize (object);
//  }

//  static PangoLanguage *
//  pango_fc_fontset_get_language (PangoFontset  *fontset)
//  {
//    PangoFcFontset *fcfontset = PANGO_FC_FONTSET (fontset);

//    return pango_fc_fontset_key_get_language (pango_fc_fontset_get_key (fcfontset));
//  }

//  static PangoFont *
//  pango_fc_fontset_get_font (PangoFontset  *fontset,
// 				guint          wc)
//  {
//    PangoFcFontset *fcfontset = PANGO_FC_FONTSET (fontset);
//    PangoCoverageLevel best_level = PANGO_COVERAGE_NONE;
//    PangoCoverageLevel level;
//    PangoFont *font;
//    PangoCoverage *coverage;
//    int result = -1;
//    unsigned int i;

//    for (i = 0;
// 		pango_fc_fontset_get_font_at (fcfontset, i);
// 		i++)
// 	 {
// 	   coverage = g_ptr_array_index (fcfontset.coverages, i);

// 	   if (coverage == nil)
// 	 {
// 	   font = g_ptr_array_index (fcfontset.fonts, i);

// 	   coverage = pango_font_get_coverage (font, fcfontset.key.language);
// 	   g_ptr_array_index (fcfontset.coverages, i) = coverage;
// 	 }

// 	   level = pango_coverage_get (coverage, wc);

// 	   if (result == -1 || level > best_level)
// 	 {
// 	   result = i;
// 	   best_level = level;
// 	   if (level == PANGO_COVERAGE_EXACT)
// 		 break;
// 	 }
// 	 }

//    if (G_UNLIKELY (result == -1))
// 	 return nil;

//    font = g_ptr_array_index (fcfontset.fonts, result);
//    return g_object_ref (font);
//  }

func (fcfontset *PangoFcFontset) foreach(fn FontsetForeachFunc) {
	//    PangoFont *font;
	//    unsigned int i;

	// TODO:
	for i := 0; ; i++ {
		font := pango_fc_fontset_get_font_at(fcfontset, i)
		if fn(font) {
			return
		}
	}
}

//  /*
//   * PangoFcFontMap
//   */

//  static GType
//  pango_fc_font_map_get_item_type (GListModel *list)
//  {
//    return PANGO_TYPE_FONT_FAMILY;
//  }

//  static void ensure_families ( fcfontmap *PangoFcFontMap);

//  static guint
//  pango_fc_font_map_get_n_items (GListModel *list)
//  {
//     fcfontmap *PangoFcFontMap = PANGO_FC_FONT_MAP (list);

//    ensure_families (fcfontmap);

//    return fcfontmap.priv.n_families;
//  }

//  static gpointer
//  pango_fc_font_map_get_item (GListModel *list,
// 							 guint       position)
//  {
//     fcfontmap *PangoFcFontMap = PANGO_FC_FONT_MAP (list);

//    ensure_families (fcfontmap);

//    if (position < fcfontmap.priv.n_families)
// 	 return g_object_ref (fcfontmap.priv.families[position]);

//    return nil;
//  }

//  static void
//  pango_fc_font_map_list_model_init (GListModelInterface *iface)
//  {
//    iface.get_item_type = pango_fc_font_map_get_item_type;
//    iface.get_n_items = pango_fc_font_map_get_n_items;
//    iface.get_item = pango_fc_font_map_get_item;
//  }

//  G_DEFINE_ABSTRACT_TYPE_WITH_CODE (PangoFcFontMap, pango_fc_font_map, PANGO_TYPE_FONT_MAP,
// 								   G_ADD_PRIVATE (PangoFcFontMap)
// 								   G_IMPLEMENT_INTERFACE (G_TYPE_LIST_MODEL, pango_fc_font_map_list_model_init))

func pango_fc_font_map_init() *PangoFcFontMap {
	var priv PangoFcFontMapPrivate

	priv.font_hash = make(fontHash)

	priv.fontset_hash = make(fontsetHash)

	priv.patterns_hash = make(map[*fontconfig.FcPattern]*PangoFcPatterns)

	priv.pattern_hash = make(fcPatternHash)

	priv.font_face_data_hash = make(map[PangoFcFontFaceData]bool)
	priv.dpi = -1

	return &PangoFcFontMap{priv: &priv}
}

//  static void
//  pango_fc_font_map_class_init (PangoFcFontMapClass *class)
//  {
//    GObjectClass *object_class = G_OBJECT_CLASS (class);
//    PangoFontMapClass *fontmap_class = PANGO_FONT_MAP_CLASS (class);

//    object_class.finalize = pango_fc_font_map_finalize;
//    fontmap_class.load_font = pango_fc_font_map_load_font;
//    fontmap_class.load_fontset = pango_fc_font_map_load_fontset;
//    fontmap_class.list_families = pango_fc_font_map_list_families;
//    fontmap_class.get_family = pango_fc_font_map_get_family;
//    fontmap_class.get_face = pango_fc_font_map_get_face;
//    fontmap_class.shape_engine_type = PANGO_RENDER_TYPE_FC;
//    fontmap_class.changed = pango_fc_font_map_changed;
//  }

//  /**
//   * pango_fc_font_map_add_decoder_find_func:
//   * @fcfontmap: The #PangoFcFontMap to add this method to.
//   * @findfunc: The #PangoFcDecoderFindFunc callback function
//   * @user_data: User data.
//   * @dnotify: A #GDestroyNotify callback that will be called when the
//   *  fontmap is finalized and the decoder is released.
//   *
//   * This function saves a callback method in the #PangoFcFontMap that
//   * will be called whenever new fonts are created.  If the
//   * function returns a #PangoFcDecoder, that decoder will be used to
//   * determine both coverage via a #FcCharSet and a one-to-one mapping of
//   * characters to glyphs.  This will allow applications to have
//   * application-specific encodings for various fonts.
//   *
//   * Since: 1.6
//   **/
//  void
//  pango_fc_font_map_add_decoder_find_func (PangoFcFontMap        *fcfontmap,
// 					  PangoFcDecoderFindFunc findfunc,
// 					  gpointer               user_data,
// 					  GDestroyNotify         dnotify)
//  {
//    PangoFcFontMapPrivate *priv;
//    PangoFcFindFuncInfo *info;

//    g_return_if_fail (PANGO_IS_FC_FONT_MAP (fcfontmap));

//    priv = fcfontmap.priv;

//    info = g_slice_new (PangoFcFindFuncInfo);

//    info.findfunc = findfunc;
//    info.user_data = user_data;
//    info.dnotify = dnotify;

//    priv.findfuncs = g_slist_append (priv.findfuncs, info);
//  }

//  /**
//   * pango_fc_font_map_find_decoder:
//   * @fcfontmap: The #PangoFcFontMap to use.
//   * @pattern: The #FcPattern to find the decoder for.
//   *
//   * Finds the decoder to use for @pattern.  Decoders can be added to
//   * a font map using pango_fc_font_map_add_decoder_find_func().
//   *
//   * Returns: (transfer full) (nullable): a newly created #PangoFcDecoder
//   *   object or %nil if no decoder is set for @pattern.
//   *
//   * Since: 1.26
//   **/
//  PangoFcDecoder *
//  pango_fc_font_map_find_decoder  ( fcfontmap *PangoFcFontMap,
// 				  pattern *FcPattern      )
//  {
//    GSList *l;

//    g_return_val_if_fail (PANGO_IS_FC_FONT_MAP (fcfontmap), nil);
//    g_return_val_if_fail (pattern != nil, nil);

//    for (l = fcfontmap.priv.findfuncs; l && l.data; l = l.next)
// 	 {
// 	   PangoFcFindFuncInfo *info = l.data;
// 	   PangoFcDecoder *decoder;

// 	   decoder = info.findfunc (pattern, info.user_data);
// 	   if (decoder)
// 	 return decoder;
// 	 }

//    return nil;
//  }

//  static void
//  pango_fc_font_map_finalize (GObject *object)
//  {
//     fcfontmap *PangoFcFontMap = PANGO_FC_FONT_MAP (object);

//    pango_fc_font_map_shutdown (fcfontmap);

//    if (fcfontmap.substitute_destroy)
// 	 fcfontmap.substitute_destroy (fcfontmap.substitute_data);

//    G_OBJECT_CLASS (pango_fc_font_map_parent_class).finalize (object);
//  }

//  /* Add a mapping from key to fcfont */
//  static void
//  pango_fc_font_map_add ( fcfontmap *PangoFcFontMap,
// 				PangoFcFontKey *key,
// 				PangoFcFont    *fcfont)
//  {
//    PangoFcFontMapPrivate *priv = fcfontmap.priv;
//    PangoFcFontKey *key_copy;

//    key_copy = pango_fc_font_key_copy (key);
//    _pango_fc_font_set_font_key (fcfont, key_copy);
//    g_hash_table_insert (priv.font_hash, key_copy, fcfont);
//  }

//  /* Remove mapping from fcfont.key to fcfont */
//  /* Closely related to shutdown_font() */
//  void
//  _pango_fc_font_map_remove ( fcfontmap *PangoFcFontMap,
// 				PangoFcFont    *fcfont)
//  {
//    PangoFcFontMapPrivate *priv = fcfontmap.priv;
//    PangoFcFontKey *key;

//    key = _pango_fc_font_get_font_key (fcfont);
//    if (key)
// 	 {
// 	   /* Only remove from fontmap hash if we are in it.  This is not necessarily
// 		* the case after a cache_clear() call. */
// 	   if (priv.font_hash &&
// 	   fcfont == g_hash_table_lookup (priv.font_hash, key))
// 		 {
// 	   g_hash_table_remove (priv.font_hash, key);
// 	 }
// 	   _pango_fc_font_set_font_key (fcfont, nil);
// 	   pango_fc_font_key_free (key);
// 	 }
//  }

//  static PangoFcFamily *
//  create_family ( fcfontmap *PangoFcFontMap,
// 			const char     *family_name,
// 			int             spacing)
//  {
//    PangoFcFamily *family = g_object_new (PANGO_FC_TYPE_FAMILY, nil);
//    family.fontmap = fcfontmap;
//    family.family_name = g_strdup (family_name);
//    family.spacing = spacing;
//    family.variable = false;
//    family.patterns = FcFontSetCreate ();

//    return family;
//  }

//  static bool
//  is_alias_family (const char *family_name)
//  {
//    switch (family_name[0])
// 	 {
// 	 case 'c':
// 	 case 'C':
// 	   return (g_ascii_strcasecmp (family_name, "cursive") == 0);
// 	 case 'f':
// 	 case 'F':
// 	   return (g_ascii_strcasecmp (family_name, "fantasy") == 0);
// 	 case 'm':
// 	 case 'M':
// 	   return (g_ascii_strcasecmp (family_name, "monospace") == 0);
// 	 case 's':
// 	 case 'S':
// 	   return (g_ascii_strcasecmp (family_name, "sans") == 0 ||
// 		   g_ascii_strcasecmp (family_name, "serif") == 0 ||
// 		   g_ascii_strcasecmp (family_name, "system-ui") == 0);
// 	 }

//    return false;
//  }

//  static void
//  ensure_families ( fcfontmap *PangoFcFontMap)
//  {
//    PangoFcFontMapPrivate *priv = fcfontmap.priv;
//    FcFontSet *fontset;
//    int i;
//    int count;

//    if (priv.n_families < 0)
// 	 {
// 	   FcObjectSet *os = FcObjectSetBuild (FC_FAMILY, FC_SPACING, FC_STYLE, FC_WEIGHT, FC_WIDTH, FC_SLANT,
//  #ifdef FC_VARIABLE
// 										   FC_VARIABLE,
//  #endif
// 										   FC_FONTFORMAT,
// 										   nil);
// 	   FcPattern *pat = FcPatternCreate ();
// 	   GHashTable *temp_family_hash;

// 	   fontset = FcFontList (priv.config, pat, os);

// 	   FcPatternDestroy (pat);
// 	   FcObjectSetDestroy (os);

// 	   priv.families = g_new (PangoFcFamily *, fontset.nfont + 4); /* 4 standard aliases */
// 	   temp_family_hash = g_hash_table_new_full (g_str_hash, g_str_equal, g_free, nil);

// 	   count = 0;
// 	   for (i = 0; i < fontset.nfont; i++)
// 	 {
// 	   char *s;
// 	   FcResult res;
// 	   int spacing;
// 		   int variable;
// 	   PangoFcFamily *temp_family;

// 		   if (!pango_fc_is_supported_font_format (fontset.fonts[i]))
// 			 continue;

// 	   res = FcPatternGetString (fontset.fonts[i], FC_FAMILY, 0, (FcChar8 **)(void*)&s);
// 	   g_assert (res == FcResultMatch);

// 	   temp_family = g_hash_table_lookup (temp_family_hash, s);
// 	   if (!is_alias_family (s) && !temp_family)
// 		 {
// 		   res = FcPatternGetInteger (fontset.fonts[i], FC_SPACING, 0, &spacing);
// 		   g_assert (res == FcResultMatch || res == FcResultNoMatch);
// 		   if (res == FcResultNoMatch)
// 		 spacing = FC_PROPORTIONAL;

// 		   temp_family = create_family (fcfontmap, s, spacing);
// 		   g_hash_table_insert (temp_family_hash, g_strdup (s), temp_family);
// 		   priv.families[count++] = temp_family;
// 		 }

// 	   if (temp_family)
// 		 {
// 			   variable = false;
//  #ifdef FC_VARIABLE
// 			   variable = FcPatternGetBool (fontset.fonts[i], FC_VARIABLE, 0, &variable);
//  #endif
// 			   if (variable)
// 				 temp_family.variable = true;

// 		   FcPatternReference (fontset.fonts[i]);
// 		   FcFontSetAdd (temp_family.patterns, fontset.fonts[i]);
// 		 }
// 	 }

// 	   FcFontSetDestroy (fontset);
// 	   g_hash_table_destroy (temp_family_hash);

// 	   priv.families[count++] = create_family (fcfontmap, "Sans", FC_PROPORTIONAL);
// 	   priv.families[count++] = create_family (fcfontmap, "Serif", FC_PROPORTIONAL);
// 	   priv.families[count++] = create_family (fcfontmap, "Monospace", FC_MONO);
// 	   priv.families[count++] = create_family (fcfontmap, "System-ui", FC_PROPORTIONAL);

// 	   priv.n_families = count;
// 	 }
//  }

//  static void
//  pango_fc_font_map_list_families (fontmap *PangoFontMap,
// 				  PangoFontFamily ***families,
// 				  int               *n_families)
//  {
//     fcfontmap *PangoFcFontMap = PANGO_FC_FONT_MAP (fontmap);
//    PangoFcFontMapPrivate *priv = fcfontmap.priv;

//    if (priv.closed)
// 	 {
// 	   if (families)
// 	 *families = nil;
// 	   if (n_families)
// 	 *n_families = 0;

// 	   return;
// 	 }

//    ensure_families (fcfontmap);

//    if (n_families)
// 	 *n_families = priv.n_families;

//    if (families)
// 	 *families = g_memdup (priv.families, priv.n_families * sizeof (PangoFontFamily *));
//  }

//  static PangoFontFamily *
//  pango_fc_font_map_get_family (PangoFontMap *fontmap,
// 							   const char   *name)
//  {
//     fcfontmap *PangoFcFontMap = PANGO_FC_FONT_MAP (fontmap);
//    PangoFcFontMapPrivate *priv = fcfontmap.priv;
//    int i;

//    if (priv.closed)
// 	 return nil;

//    ensure_families (fcfontmap);

//    for (i = 0; i < priv.n_families; i++)
// 	 {
// 	   PangoFontFamily *family = PANGO_FONT_FAMILY (priv.families[i]);
// 	   if (strcmp (name, pango_font_family_get_name (family)) == 0)
// 		 return family;
// 	 }

//    return nil;
//  }

func pango_fc_convert_slant_to_fc(pangoStyle Style) int {
	switch pangoStyle {
	case PANGO_STYLE_NORMAL:
		return fontconfig.FC_SLANT_ROMAN
	case PANGO_STYLE_ITALIC:
		return fontconfig.FC_SLANT_ITALIC
	case PANGO_STYLE_OBLIQUE:
		return fontconfig.FC_SLANT_OBLIQUE
	default:
		return fontconfig.FC_SLANT_ROMAN
	}
}

func pango_fc_convert_width_to_fc(pangoStretch Stretch) int {
	switch pangoStretch {
	case PANGO_STRETCH_NORMAL:
		return fontconfig.FC_WIDTH_NORMAL
	case PANGO_STRETCH_ULTRA_CONDENSED:
		return fontconfig.FC_WIDTH_ULTRACONDENSED
	case PANGO_STRETCH_EXTRA_CONDENSED:
		return fontconfig.FC_WIDTH_EXTRACONDENSED
	case PANGO_STRETCH_CONDENSED:
		return fontconfig.FC_WIDTH_CONDENSED
	case PANGO_STRETCH_SEMI_CONDENSED:
		return fontconfig.FC_WIDTH_SEMICONDENSED
	case PANGO_STRETCH_SEMI_EXPANDED:
		return fontconfig.FC_WIDTH_SEMIEXPANDED
	case PANGO_STRETCH_EXPANDED:
		return fontconfig.FC_WIDTH_EXPANDED
	case PANGO_STRETCH_EXTRA_EXPANDED:
		return fontconfig.FC_WIDTH_EXTRAEXPANDED
	case PANGO_STRETCH_ULTRA_EXPANDED:
		return fontconfig.FC_WIDTH_ULTRAEXPANDED
	default:
		return fontconfig.FC_WIDTH_NORMAL
	}
}

func (fcfontmap *PangoFcFontMap) uniquifyPattern(pattern *fontconfig.FcPattern) *fontconfig.FcPattern {
	priv := fcfontmap.priv
	if old_pattern := priv.pattern_hash.lookup(*pattern); old_pattern != nil {
		return old_pattern
	}
	priv.pattern_hash.insert(pattern)
	return pattern
}

//  static PangoFont *
//  pango_fc_font_map_new_font (PangoFcFontMap    *fcfontmap,
// 				 PangoFcFontsetKey *fontset_key,
// 				 FcPattern         *match)
//  {
//    PangoFcFontMapClass *class;
//    PangoFcFontMapPrivate *priv = fcfontmap.priv;
//    FcPattern *pattern;
//    PangoFcFont *fcfont;
//    PangoFcFontKey key;

//    if (priv.closed)
// 	 return nil;

//    match = uniquifyPattern (fcfontmap, match);

//    pango_fc_font_key_init (&key, fcfontmap, fontset_key, match);

//    fcfont = g_hash_table_lookup (priv.font_hash, &key);
//    if (fcfont)
// 	 return g_object_ref (PANGO_FONT (fcfont));

//    class = PANGO_FC_FONT_MAP_GET_CLASS (fcfontmap);

//    if (class.create_font)
// 	 {
// 	   fcfont = class.create_font (fcfontmap, &key);
// 	 }
//    else
// 	 {
// 	   const Matrix *pango_matrix = pango_fc_fontset_key_get_matrix (fontset_key);
// 	   FcMatrix fc_matrix, *fc_matrix_val;
// 	   int i;

// 	   /* Fontconfig has the Y axis pointing up, Pango, down.
// 		*/
// 	   fc_matrix.xx = pango_matrix.xx;
// 	   fc_matrix.xy = - pango_matrix.xy;
// 	   fc_matrix.yx = - pango_matrix.yx;
// 	   fc_matrix.yy = pango_matrix.yy;

// 	   pattern = FcPatternDuplicate (match);

// 	   for (i = 0; FcPatternGetMatrix (pattern, FC_MATRIX, i, &fc_matrix_val) == FcResultMatch; i++)
// 	 FcMatrixMultiply (&fc_matrix, &fc_matrix, fc_matrix_val);

// 	   FcPatternDel (pattern, FC_MATRIX);
// 	   FcPatternAddMatrix (pattern, FC_MATRIX, &fc_matrix);

// 	   fcfont = class.new_font (fcfontmap, uniquifyPattern (fcfontmap, pattern));

// 	   FcPatternDestroy (pattern);
// 	 }

//    if (!fcfont)
// 	 return nil;

//    fcfont.matrix = key.matrix;
//    /* In case the backend didn't set the fontmap */
//    if (!fcfont.fontmap)
// 	 g_object_set (fcfont,
// 		   "fontmap", fcfontmap,
// 		   nil);

//    /* cache it on fontmap */
//    pango_fc_font_map_add (fcfontmap, &key, fcfont);

//    return (PangoFont *)fcfont;
//  }

//  static PangoFontFace *
//  pango_fc_font_map_get_face (PangoFontMap *fontmap,
// 							 PangoFont    *font)
//  {
//    PangoFcFont *fcfont = PANGO_FC_FONT (font);
//    FcResult res;
//    const char *s;
//    PangoFontFamily *family;

//    res = FcPatternGetString (fcfont.font_pattern, FC_FAMILY, 0, (FcChar8 **) &s);
//    g_assert (res == FcResultMatch);

//    family = pango_font_map_get_family (fontmap, s);

//    res = FcPatternGetString (fcfont.font_pattern, FC_STYLE, 0, (FcChar8 **)(void*)&s);
//    g_assert (res == FcResultMatch);

//    return pango_font_family_get_face (family, s);
//  }

func (fontsetkey *PangoFcFontsetKey) pango_fc_default_substitute(fontmap *PangoFcFontMap, pattern *fontconfig.FcPattern) {
	if fontmap.fontset_key_substitute != nil {
		fontmap.fontset_key_substitute(fontsetkey, pattern)
	} else if fontmap.default_substitute != nil {
		fontmap.default_substitute(pattern)
	}
}

//  void
//  pango_fc_font_map_set_default_substitute (PangoFcFontMap        *fontmap,
// 					   PangoFcSubstituteFunc func,
// 					   gpointer              data,
// 					   GDestroyNotify        notify)
//  {
//    if (fontmap.substitute_destroy)
// 	 fontmap.substitute_destroy (fontmap.substitute_data);

//    fontmap.substitute_func = func;
//    fontmap.substitute_data = data;
//    fontmap.substitute_destroy = notify;

//    pango_fc_font_map_substitute_changed (fontmap);
//  }

//  void
//  pango_fc_font_map_substitute_changed (fontmap *PangoFcFontMap) {
//    pango_fc_font_map_cache_clear(fontmap);
//    pango_font_map_changed(PANGO_FONT_MAP (fontmap));
//  }

func (fcfontmap *PangoFcFontMap) pango_fc_font_map_get_resolution(context *Context) float64 {
	// if PANGO_FC_FONT_MAP_GET_CLASS(fcfontmap).get_resolution {
	// 	return PANGO_FC_FONT_MAP_GET_CLASS(fcfontmap).get_resolution(fcfontmap, context)
	// }

	// the default subtitution from fontconfig is 75DPI
	// TODO: check if the user provided subsitution are needed
	if fcfontmap.priv.dpi < 0 {
		// result := fontconfig.FcResultNoMatch
		// tmp := FcPatternBuild(nil, FC_FAMILY, FcTypeString, "Sans",
		// 	FC_SIZE, FcTypeDouble, 10., nil)
		// if tmp {
		// 	pango_fc_default_substitute(fcfontmap, nil, tmp)
		// 	result = FcPatternGetDouble(tmp, FC_DPI, 0, &fcfontmap.priv.dpi)
		// }

		// if result != FcResultMatch {
		// 	g_warning("Error getting DPI from fontconfig, using 72.0")
		// }
		fcfontmap.priv.dpi = 75
	}

	return fcfontmap.priv.dpi
}

func (fcfontmap *PangoFcFontMap) pango_fc_font_map_get_patterns(key *PangoFcFontsetKey) *PangoFcPatterns {
	pattern := key.pango_fc_fontset_key_make_pattern()
	key.pango_fc_default_substitute(fcfontmap, pattern)

	return fcfontmap.pango_fc_patterns_new(pattern)
}

//  static bool
//  get_first_font (PangoFontset  *fontset G_GNUC_UNUSED,
// 		 PangoFont     *font,
// 		 gpointer       data)
//  {
//    *(PangoFont **)data = font;

//    return true;
//  }

//  static PangoFont *
//  pango_fc_font_map_load_font (PangoFontMap               *fontmap,
// 				                 context *Context,
// 				    description *FontDescription)
//  {
//    PangoLanguage *language;
//    PangoFontset *fontset;
//    PangoFont *font = nil;

//    if (context)
// 	 language = pango_context_get_language (context);
//    else
// 	 language = nil;

//    fontset = pango_font_map_load_fontset (fontmap, context, description, language);

//    if (fontset)
// 	 {
// 	   pango_fontset_foreach (fontset, get_first_font, &font);

// 	   if (font)
// 	 g_object_ref (font);

// 	   g_object_unref (fontset);
// 	 }

//    return font;
//  }

func (fcfontmap *PangoFcFontMap) pango_fc_fontset_cache(fontset *PangoFcFontset) {
	priv := fcfontmap.priv
	cache := priv.fontset_cache

	if fontset.cache_link != nil {
		if fontset.cache_link == cache.Front() {
			return
		}
		// Already in cache, move to head
		// if fontset.cache_link == cache.Back() {
		// 	cache.tail = fontset.cache_link.prev
		// }
		cache.Remove(fontset.cache_link)
	} else {
		// Add to cache initially
		if cache.Len() == FONTSET_CACHE_SIZE {
			tmp_fontset := cache.Remove(cache.Front()).(*PangoFcFontset)
			tmp_fontset.cache_link = nil
			priv.fontset_hash.remove(*tmp_fontset.key)
		}

		fontset.cache_link = &list.Element{Value: fontset}
	}

	cache.PushFront(fontset.cache_link.Value)
}

func (fcfontmap *PangoFcFontMap) load_fontset(context *Context, desc *FontDescription, language Language) Fontset {
	priv := fcfontmap.priv

	key := fcfontmap.newFontsetKey(context, desc, language)

	fontset := priv.fontset_hash.lookup(key)
	if fontset == nil {
		patterns := fcfontmap.pango_fc_font_map_get_patterns(&key)

		if patterns == nil {
			return nil
		}

		fontset = pango_fc_fontset_new(key, patterns)
		priv.fontset_hash.insert(*fontset.key, fontset)
	}

	fcfontmap.pango_fc_fontset_cache(fontset)

	return fontset
}

//  /**
//   * pango_fc_font_map_cache_clear:
//   * @fcfontmap: a #PangoFcFontMap
//   *
//   * Clear all cached information and fontsets for this font map;
//   * this should be called whenever there is a change in the
//   * output of the default_substitute() virtual function of the
//   * font map, or if fontconfig has been reinitialized to new
//   * configuration.
//   *
//   * Since: 1.4
//   **/
//  void
//  pango_fc_font_map_cache_clear ( fcfontmap *PangoFcFontMap)
//  {
//    guint removed, added;

//    if (G_UNLIKELY (fcfontmap.priv.closed))
// 	 return;

//    removed = fcfontmap.priv.n_families;

//    pango_fc_font_map_fini (fcfontmap);
//    pango_fc_font_map_init (fcfontmap);

//    ensure_families (fcfontmap);

//    added = fcfontmap.priv.n_families;

//    g_list_model_items_changed (G_LIST_MODEL (fcfontmap), 0, removed, added);

//    pango_font_map_changed (PANGO_FONT_MAP (fcfontmap));
//  }

//  static void
//  pango_fc_font_map_changed (PangoFontMap *fontmap)
//  {
//    /* we emit GListModel::changed in pango_fc_font_map_cache_clear() */
//  }

//  /**
//   * pango_fc_font_map_config_changed:
//   * @fcfontmap: a #PangoFcFontMap
//   *
//   * Informs font map that the fontconfig configuration (ie, FcConfig object)
//   * used by this font map has changed.  This currently calls
//   * pango_fc_font_map_cache_clear() which ensures that list of fonts, etc
//   * will be regenerated using the updated configuration.
//   *
//   * Since: 1.38
//   **/
//  void
//  pango_fc_font_map_config_changed ( fcfontmap *PangoFcFontMap)
//  {
//    pango_fc_font_map_cache_clear (fcfontmap);
//  }

//  /**
//   * pango_fc_font_map_set_config: (skip)
//   * @fcfontmap: a #PangoFcFontMap
//   * @fcconfig: (nullable): a `FcConfig`, or %nil
//   *
//   * Set the FcConfig for this font map to use.  The default value
//   * is %nil, which causes Fontconfig to use its global "current config".
//   * You can create a new FcConfig object and use this API to attach it
//   * to a font map.
//   *
//   * This is particularly useful for example, if you want to use application
//   * fonts with Pango.  For that, you would create a fresh FcConfig, add your
//   * app fonts to it, and attach it to a new Pango font map.
//   *
//   * If @fcconfig is different from the previous config attached to the font map,
//   * pango_fc_font_map_config_changed() is called.
//   *
//   * This function acquires a reference to the FcConfig object; the caller
//   * does NOT need to retain a reference.
//   *
//   * Since: 1.38
//   **/
//  void
//  pango_fc_font_map_set_config ( fcfontmap *PangoFcFontMap,
// 				   FcConfig       *fcconfig)
//  {
//    FcConfig *oldconfig;

//    g_return_if_fail (PANGO_IS_FC_FONT_MAP (fcfontmap));

//    oldconfig = fcfontmap.priv.config;

//    if (fcconfig)
// 	 FcConfigReference (fcconfig);

//    fcfontmap.priv.config = fcconfig;

//    if (oldconfig != fcconfig)
// 	 pango_fc_font_map_config_changed (fcfontmap);

//    if (oldconfig)
// 	 FcConfigDestroy (oldconfig);
//  }

//  /**
//   * pango_fc_font_map_get_config: (skip)
//   * @fcfontmap: a #PangoFcFontMap
//   *
//   * Fetches the `FcConfig` attached to a font map.
//   *
//   * See also: pango_fc_font_map_set_config()
//   *
//   * Returns: (nullable): the `FcConfig` object attached to @fcfontmap, which
//   *          might be %nil.
//   *
//   * Since: 1.38
//   **/
//  FcConfig *
//  pango_fc_font_map_get_config ( fcfontmap *PangoFcFontMap)
//  {
//    g_return_val_if_fail (PANGO_IS_FC_FONT_MAP (fcfontmap), nil);

//    return fcfontmap.priv.config;
//  }

//  static PangoFcFontFaceData *
//  pango_fc_font_map_get_font_face_data ( fcfontmap *PangoFcFontMap,
// 					   FcPattern      *font_pattern)
//  {
//    PangoFcFontMapPrivate *priv = fcfontmap.priv;
//    PangoFcFontFaceData key;
//    PangoFcFontFaceData *data;

//    if (FcPatternGetString (font_pattern, FC_FILE, 0, (FcChar8 **)(void*)&key.filename) != FcResultMatch)
// 	 return nil;

//    if (FcPatternGetInteger (font_pattern, FC_INDEX, 0, &key.id) != FcResultMatch)
// 	 return nil;

//    data = g_hash_table_lookup (priv.font_face_data_hash, &key);
//    if (G_LIKELY (data))
// 	 return data;

//    data = g_slice_new0 (PangoFcFontFaceData);
//    data.filename = key.filename;
//    data.id = key.id;

//    data.pattern = font_pattern;
//    FcPatternReference (data.pattern);

//    g_hash_table_insert (priv.font_face_data_hash, data, data);

//    return data;
//  }

//  typedef struct {
//    PangoCoverage parent_instance;

//    FcCharSet *charset;
//  } PangoFcCoverage;

//  typedef struct {
//    PangoCoverageClass parent_class;
//  } PangoFcCoverageClass;

//  GType pango_fc_coverage_get_type (void) G_GNUC_CONST;

//  G_DEFINE_TYPE (PangoFcCoverage, pango_fc_coverage, PANGO_TYPE_COVERAGE)

//  static void
//  pango_fc_coverage_init (PangoFcCoverage *coverage)
//  {
//  }

//  static PangoCoverageLevel
//  pango_fc_coverage_real_get (PangoCoverage *coverage,
// 							 int            index)
//  {
//    PangoFcCoverage *fc_coverage = (PangoFcCoverage*)coverage;

//    return FcCharSetHasChar (fc_coverage.charset, index)
// 		  ? PANGO_COVERAGE_EXACT
// 		  : PANGO_COVERAGE_NONE;
//  }

//  static void
//  pango_fc_coverage_real_set (PangoCoverage *coverage,
// 							 int            index,
// 							 PangoCoverageLevel level)
//  {
//    PangoFcCoverage *fc_coverage = (PangoFcCoverage*)coverage;

//    if (level == PANGO_COVERAGE_NONE)
// 	 FcCharSetDelChar (fc_coverage.charset, index);
//    else
// 	 FcCharSetAddChar (fc_coverage.charset, index);
//  }

//  static PangoCoverage *
//  pango_fc_coverage_real_copy (PangoCoverage *coverage)
//  {
//    PangoFcCoverage *fc_coverage = (PangoFcCoverage*)coverage;
//    PangoFcCoverage *copy;

//    copy = g_object_new (pango_fc_coverage_get_type (), nil);
//    copy.charset = FcCharSetCopy (fc_coverage.charset);

//    return (PangoCoverage *)copy;
//  }

//  static void
//  pango_fc_coverage_finalize (GObject *object)
//  {
//    PangoFcCoverage *fc_coverage = (PangoFcCoverage*)object;

//    FcCharSetDestroy (fc_coverage.charset);

//    G_OBJECT_CLASS (pango_fc_coverage_parent_class).finalize (object);
//  }

//  static void
//  pango_fc_coverage_class_init (PangoFcCoverageClass *class)
//  {
//    GObjectClass *object_class = G_OBJECT_CLASS (class);
//    PangoCoverageClass *coverage_class = PANGO_COVERAGE_CLASS (class);

//    object_class.finalize = pango_fc_coverage_finalize;

//    coverage_class.get = pango_fc_coverage_real_get;
//    coverage_class.set = pango_fc_coverage_real_set;
//    coverage_class.copy = pango_fc_coverage_real_copy;
//  }

//  PangoCoverage *
//  _pango_fc_font_map_get_coverage ( fcfontmap *PangoFcFontMap,
// 				  PangoFcFont    *fcfont)
//  {
//    PangoFcFontFaceData *data;
//    FcCharSet *charset;

//    data = pango_fc_font_map_get_font_face_data (fcfontmap, fcfont.font_pattern);
//    if (G_UNLIKELY (!data))
// 	 return nil;

//    if (G_UNLIKELY (data.coverage == nil))
// 	 {
// 	   /*
// 		* Pull the coverage out of the pattern, this
// 		* doesn't require loading the font
// 		*/
// 	   if (FcPatternGetCharSet (fcfont.font_pattern, FC_CHARSET, 0, &charset) != FcResultMatch)
// 		 return nil;

// 	   data.coverage = _pango_fc_font_map_fc_to_coverage (charset);
// 	 }

//    return pango_coverage_ref (data.coverage);
//  }

//  /**
//   * _pango_fc_font_map_fc_to_coverage:
//   * @charset: #FcCharSet to convert to a #PangoCoverage object.
//   *
//   * Convert the given #FcCharSet into a new #PangoCoverage object.  The
//   * caller is responsible for freeing the newly created object.
//   *
//   * Since: 1.6
//   **/
//  PangoCoverage  *
//  _pango_fc_font_map_fc_to_coverage (FcCharSet *charset)
//  {
//    PangoFcCoverage *coverage;

//    coverage = g_object_new (pango_fc_coverage_get_type (), nil);
//    coverage.charset = FcCharSetCopy (charset);

//    return (PangoCoverage *)coverage;
//  }

//  static PangoLanguage **
//  _pango_fc_font_map_fc_to_languages (FcLangSet *langset)
//  {
//    FcStrSet *strset;
//    FcStrList *list;
//    FcChar8 *s;
//    GArray *langs;

//    langs = g_array_new (true, false, sizeof (PangoLanguage *));

//    strset = FcLangSetGetLangs (langset);
//    list = FcStrListCreate (strset);

//    FcStrListFirst (list);
//    while ((s = FcStrListNext (list)))
// 	 {
// 	   PangoLanguage *l = pango_language_from_string ((const char *)s);
// 	   g_array_append_val (langs, l);
// 	 }

//    FcStrListDone (list);
//    FcStrSetDestroy (strset);

//    return (PangoLanguage **) g_array_free (langs, false);
//  }

//  PangoLanguage **
//  _pango_fc_font_map_get_languages ( fcfontmap *PangoFcFontMap,
// 								   PangoFcFont    *fcfont)
//  {
//    PangoFcFontFaceData *data;
//    FcLangSet *langset;

//    data = pango_fc_font_map_get_font_face_data (fcfontmap, fcfont.font_pattern);
//    if (G_UNLIKELY (!data))
// 	 return nil;

//    if (G_UNLIKELY (data.languages == nil))
// 	 {
// 	   /*
// 		* Pull the languages out of the pattern, this
// 		* doesn't require loading the font
// 		*/
// 	   if (FcPatternGetLangSet (fcfont.font_pattern, FC_LANG, 0, &langset) != FcResultMatch)
// 		 return nil;

// 	   data.languages = _pango_fc_font_map_fc_to_languages (langset);
// 	 }

//    return data.languages;
//  }
//  /**
//   * pango_fc_font_map_create_context:
//   * @fcfontmap: a #PangoFcFontMap
//   *
//   * Creates a new context for this fontmap. This function is intended
//   * only for backend implementations deriving from #PangoFcFontMap;
//   * it is possible that a backend will store additional information
//   * needed for correct operation on the #Context after calling
//   * this function.
//   *
//   * Return value: (transfer full): a new #Context
//   *
//   * Since: 1.4
//   *
//   * Deprecated: 1.22: Use pango_font_map_create_context() instead.
//   **/
//  Context *
//  pango_fc_font_map_create_context ( fcfontmap *PangoFcFontMap)
//  {
//    g_return_val_if_fail (PANGO_IS_FC_FONT_MAP (fcfontmap), nil);

//    return pango_font_map_create_context (PANGO_FONT_MAP (fcfontmap));
//  }

//  static void
//  shutdown_font (gpointer        key,
// 			PangoFcFont    *fcfont,
// 			 fcfontmap *PangoFcFontMap)
//  {
//    _pango_fc_font_shutdown (fcfont);

//    _pango_fc_font_set_font_key (fcfont, nil);
//    pango_fc_font_key_free (key);
//  }

//  /**
//   * pango_fc_font_map_shutdown:
//   * @fcfontmap: a #PangoFcFontMap
//   *
//   * Clears all cached information for the fontmap and marks
//   * all fonts open for the fontmap as dead. (See the shutdown()
//   * virtual function of #PangoFcFont.) This function might be used
//   * by a backend when the underlying windowing system for the font
//   * map exits. This function is only intended to be called
//   * only for backend implementations deriving from #PangoFcFontMap.
//   *
//   * Since: 1.4
//   **/
//  void
//  pango_fc_font_map_shutdown ( fcfontmap *PangoFcFontMap)
//  {
//    PangoFcFontMapPrivate *priv = fcfontmap.priv;
//    int i;

//    if (priv.closed)
// 	 return;

//    g_hash_table_foreach (priv.font_hash, (GHFunc) shutdown_font, fcfontmap);
//    for (i = 0; i < priv.n_families; i++)
// 	 priv.families[i].fontmap = nil;

//    pango_fc_font_map_fini (fcfontmap);

//    while (priv.findfuncs)
// 	 {
// 	   PangoFcFindFuncInfo *info;
// 	   info = priv.findfuncs.data;
// 	   if (info.dnotify)
// 	 info.dnotify (info.user_data);

// 	   g_slice_free (PangoFcFindFuncInfo, info);
// 	   priv.findfuncs = g_slist_delete_link (priv.findfuncs, priv.findfuncs);
// 	 }

//    priv.closed = true;
//  }

//  static PangoWeight
//  pango_fc_convert_weight_to_pango (float64 fc_weight)
//  {
//  #ifdef HAVE_FCWEIGHTFROMOPENTYPEDOUBLE
//    return FcWeightToOpenTypeDouble (fc_weight);
//  #else
//    return FcWeightToOpenType (fc_weight);
//  #endif
//  }

//  static PangoStyle
//  pango_fc_convert_slant_to_pango (int fc_style)
//  {
//    switch (fc_style)
// 	 {
// 	 case FC_SLANT_ROMAN:
// 	   return PANGO_STYLE_NORMAL;
// 	 case FC_SLANT_ITALIC:
// 	   return PANGO_STYLE_ITALIC;
// 	 case FC_SLANT_OBLIQUE:
// 	   return PANGO_STYLE_OBLIQUE;
// 	 default:
// 	   return PANGO_STYLE_NORMAL;
// 	 }
//  }

//  static PangoStretch
//  pango_fc_convert_width_to_pango (int fc_stretch)
//  {
//    switch (fc_stretch)
// 	 {
// 	 case FC_WIDTH_NORMAL:
// 	   return PANGO_STRETCH_NORMAL;
// 	 case FC_WIDTH_ULTRACONDENSED:
// 	   return PANGO_STRETCH_ULTRA_CONDENSED;
// 	 case FC_WIDTH_EXTRACONDENSED:
// 	   return PANGO_STRETCH_EXTRA_CONDENSED;
// 	 case FC_WIDTH_CONDENSED:
// 	   return PANGO_STRETCH_CONDENSED;
// 	 case FC_WIDTH_SEMICONDENSED:
// 	   return PANGO_STRETCH_SEMI_CONDENSED;
// 	 case FC_WIDTH_SEMIEXPANDED:
// 	   return PANGO_STRETCH_SEMI_EXPANDED;
// 	 case FC_WIDTH_EXPANDED:
// 	   return PANGO_STRETCH_EXPANDED;
// 	 case FC_WIDTH_EXTRAEXPANDED:
// 	   return PANGO_STRETCH_EXTRA_EXPANDED;
// 	 case FC_WIDTH_ULTRAEXPANDED:
// 	   return PANGO_STRETCH_ULTRA_EXPANDED;
// 	 default:
// 	   return PANGO_STRETCH_NORMAL;
// 	 }
//  }

//  /**
//   * pango_fc_font_description_from_pattern:
//   * @pattern: a #FcPattern
//   * @include_size: if %true, the pattern will include the size from
//   *   the @pattern; otherwise the resulting pattern will be unsized.
//   *   (only %FC_SIZE is examined, not %FC_PIXEL_SIZE)
//   *
//   * Creates a #FontDescription that matches the specified
//   * Fontconfig pattern as closely as possible. Many possible Fontconfig
//   * pattern values, such as %FC_RASTERIZER or %FC_DPI, don't make sense in
//   * the context of #FontDescription, so will be ignored.
//   *
//   * Return value: a new #FontDescription. Free with
//   *  pango_font_description_free().
//   *
//   * Since: 1.4
//   **/
//  FontDescription *
//  pango_fc_font_description_from_pattern (FcPattern *pattern, bool include_size)
//  {
//    FontDescription *desc;
//    PangoStyle style;
//    PangoWeight weight;
//    PangoStretch stretch;
//    float64 size;
//    PangoGravity gravity;

//    FcChar8 *s;
//    int i;
//    float64 d;
//    FcResult res;

//    desc = pango_font_description_new ();

//    res = FcPatternGetString (pattern, FC_FAMILY, 0, (FcChar8 **) &s);
//    g_assert (res == FcResultMatch);

//    pango_font_description_set_family (desc, (gchar *)s);

//    if (FcPatternGetInteger (pattern, FC_SLANT, 0, &i) == FcResultMatch)
// 	 style = pango_fc_convert_slant_to_pango (i);
//    else
// 	 style = PANGO_STYLE_NORMAL;

//    pango_font_description_set_style (desc, style);

//    if (FcPatternGetDouble (pattern, FC_WEIGHT, 0, &d) == FcResultMatch)
// 	 weight = pango_fc_convert_weight_to_pango (d);
//    else
// 	 weight = PANGO_WEIGHT_NORMAL;

//    pango_font_description_set_weight (desc, weight);

//    if (FcPatternGetInteger (pattern, FC_WIDTH, 0, &i) == FcResultMatch)
// 	 stretch = pango_fc_convert_width_to_pango (i);
//    else
// 	 stretch = PANGO_STRETCH_NORMAL;

//    pango_font_description_set_stretch (desc, stretch);

//    pango_font_description_set_variant (desc, PANGO_VARIANT_NORMAL);

//    if (include_size && FcPatternGetDouble (pattern, FC_SIZE, 0, &size) == FcResultMatch)
// 	 pango_font_description_set_size (desc, size * PANGO_SCALE);

//    /* gravity is a bit different.  we don't want to set it if it was not set on
// 	* the pattern */
//    if (FcPatternGetString (pattern, PANGO_FC_GRAVITY, 0, (FcChar8 **)&s) == FcResultMatch)
// 	 {
// 	   GEnumValue *value = g_enum_get_value_by_nick (get_gravity_class (), (char *)s);
// 	   gravity = value.value;

// 	   pango_font_description_set_gravity (desc, gravity);
// 	 }

//    if (include_size && FcPatternGetString (pattern, PANGO_FC_FONT_VARIATIONS, 0, (FcChar8 **)&s) == FcResultMatch)
// 	 {
// 	   if (s && *s)
// 		 pango_font_description_set_variations (desc, (char *)s);
// 	 }

//    return desc;
//  }

//  /*
//   * PangoFcFace
//   */

//  typedef PangoFontFaceClass PangoFcFaceClass;

//  G_DEFINE_TYPE (PangoFcFace, pango_fc_face, PANGO_TYPE_FONT_FACE)

//  static FontDescription *
//  make_alias_description (PangoFcFamily *fcfamily,
// 			 bool        bold,
// 			 bool        italic)
//  {
//    FontDescription *desc = pango_font_description_new ();

//    pango_font_description_set_family (desc, fcfamily.family_name);
//    pango_font_description_set_style (desc, italic ? PANGO_STYLE_ITALIC : PANGO_STYLE_NORMAL);
//    pango_font_description_set_weight (desc, bold ? PANGO_WEIGHT_BOLD : PANGO_WEIGHT_NORMAL);

//    return desc;
//  }

//  static FontDescription *
//  pango_fc_face_describe (PangoFontFace *face)
//  {
//    PangoFcFace *fcface = PANGO_FC_FACE (face);
//    PangoFcFamily *fcfamily = fcface.family;
//    FontDescription *desc = nil;

//    if (G_UNLIKELY (!fcfamily))
// 	 return pango_font_description_new ();

//    if (fcface.fake)
// 	 {
// 	   if (strcmp (fcface.style, "Regular") == 0)
// 	 return make_alias_description (fcfamily, false, false);
// 	   else if (strcmp (fcface.style, "Bold") == 0)
// 	 return make_alias_description (fcfamily, true, false);
// 	   else if (strcmp (fcface.style, "Italic") == 0)
// 	 return make_alias_description (fcfamily, false, true);
// 	   else			/* Bold Italic */
// 	 return make_alias_description (fcfamily, true, true);
// 	 }

//    g_assert (fcface.pattern);
//    desc = pango_fc_font_description_from_pattern (fcface.pattern, false);

//    return desc;
//  }

//  static const char *
//  pango_fc_face_get_face_name (PangoFontFace *face)
//  {
//    PangoFcFace *fcface = PANGO_FC_FACE (face);

//    return fcface.style;
//  }

//  static int
//  compare_ints (gconstpointer ap,
// 		   gconstpointer bp)
//  {
//    int a = *(int *)ap;
//    int b = *(int *)bp;

//    if (a == b)
// 	 return 0;
//    else if (a > b)
// 	 return 1;
//    else
// 	 return -1;
//  }

//  static void
//  pango_fc_face_list_sizes (PangoFontFace  *face,
// 			   int           **sizes,
// 			   int            *n_sizes)
//  {
//    PangoFcFace *fcface = PANGO_FC_FACE (face);
//    FcPattern *pattern;
//    FcFontSet *fontset;
//    FcObjectSet *objectset;

//    *sizes = nil;
//    *n_sizes = 0;
//    if (G_UNLIKELY (!fcface.family || !fcface.family.fontmap))
// 	 return;

//    pattern = FcPatternCreate ();
//    FcPatternAddString (pattern, FC_FAMILY, (FcChar8*)(void*)fcface.family.family_name);
//    FcPatternAddString (pattern, FC_STYLE, (FcChar8*)(void*)fcface.style);

//    objectset = FcObjectSetCreate ();
//    FcObjectSetAdd (objectset, FC_PIXEL_SIZE);

//    fontset = FcFontList (nil, pattern, objectset);

//    if (fontset)
// 	 {
// 	   GArray *size_array;
// 	   float64 size, dpi = -1.0;
// 	   int i, size_i, j;

// 	   size_array = g_array_new (false, false, sizeof (int));

// 	   for (i = 0; i < fontset.nfont; i++)
// 	 {
// 	   for (j = 0;
// 			FcPatternGetDouble (fontset.fonts[i], FC_PIXEL_SIZE, j, &size) == FcResultMatch;
// 			j++)
// 		 {
// 		   if (dpi < 0)
// 		 dpi = pango_fc_font_map_get_resolution (fcface.family.fontmap, nil);

// 		   size_i = (int) (PANGO_SCALE * size * 72.0 / dpi);
// 		   g_array_append_val (size_array, size_i);
// 		 }
// 	 }

// 	   g_array_sort (size_array, compare_ints);

// 	   if (size_array.len == 0)
// 	 {
// 	   *n_sizes = 0;
// 	   if (sizes)
// 		 *sizes = nil;
// 	   g_array_free (size_array, true);
// 	 }
// 	   else
// 	 {
// 	   *n_sizes = size_array.len;
// 	   if (sizes)
// 		 {
// 		   *sizes = (int *) size_array.data;
// 		   g_array_free (size_array, false);
// 		 }
// 	   else
// 		 g_array_free (size_array, true);
// 	 }

// 	   FcFontSetDestroy (fontset);
// 	 }
//    else
// 	 {
// 	   *n_sizes = 0;
// 	   if (sizes)
// 	 *sizes = nil;
// 	 }

//    FcPatternDestroy (pattern);
//    FcObjectSetDestroy (objectset);
//  }

//  static bool
//  pango_fc_face_is_synthesized (PangoFontFace *face)
//  {
//    PangoFcFace *fcface = PANGO_FC_FACE (face);

//    return fcface.fake;
//  }

//  static PangoFontFamily *
//  pango_fc_face_get_family (PangoFontFace *face)
//  {
//    PangoFcFace *fcface = PANGO_FC_FACE (face);

//    return PANGO_FONT_FAMILY (fcface.family);
//  }

//  static void
//  pango_fc_face_finalize (GObject *object)
//  {
//    PangoFcFace *fcface = PANGO_FC_FACE (object);

//    g_free (fcface.style);
//    FcPatternDestroy (fcface.pattern);

//    G_OBJECT_CLASS (pango_fc_face_parent_class).finalize (object);
//  }

//  static void
//  pango_fc_face_init (PangoFcFace *self)
//  {
//  }

//  static void
//  pango_fc_face_class_init (PangoFcFaceClass *class)
//  {
//    GObjectClass *object_class = G_OBJECT_CLASS (class);

//    object_class.finalize = pango_fc_face_finalize;

//    class.describe = pango_fc_face_describe;
//    class.get_face_name = pango_fc_face_get_face_name;
//    class.list_sizes = pango_fc_face_list_sizes;
//    class.is_synthesized = pango_fc_face_is_synthesized;
//    class.get_family = pango_fc_face_get_family;
//  }

//  /*
//   * PangoFcFamily
//   */

//  typedef PangoFontFamilyClass PangoFcFamilyClass;

//  static GType
//  pango_fc_family_get_item_type (GListModel *list)
//  {
//    return PANGO_TYPE_FONT_FACE;
//  }

//  static void ensure_faces (PangoFcFamily *family);

//  static guint
//  pango_fc_family_get_n_items (GListModel *list)
//  {
//    PangoFcFamily *fcfamily = PANGO_FC_FAMILY (list);

//    ensure_faces (fcfamily);

//    return (guint)fcfamily.n_faces;
//  }

//  static gpointer
//  pango_fc_family_get_item (GListModel *list,
// 						   guint       position)
//  {
//    PangoFcFamily *fcfamily = PANGO_FC_FAMILY (list);

//    ensure_faces (fcfamily);

//    if (position < fcfamily.n_faces)
// 	 return g_object_ref (fcfamily.faces[position]);

//    return nil;
//  }

//  static void
//  pango_fc_family_list_model_init (GListModelInterface *iface)
//  {
//    iface.get_item_type = pango_fc_family_get_item_type;
//    iface.get_n_items = pango_fc_family_get_n_items;
//    iface.get_item = pango_fc_family_get_item;
//  }

//  G_DEFINE_TYPE_WITH_CODE (PangoFcFamily, pango_fc_family, PANGO_TYPE_FONT_FAMILY,
// 						  G_IMPLEMENT_INTERFACE (G_TYPE_LIST_MODEL, pango_fc_family_list_model_init))

//  static PangoFcFace *
//  create_face (PangoFcFamily *fcfamily,
// 		  const char    *style,
// 		  FcPattern     *pattern,
// 		  bool       fake)
//  {
//    PangoFcFace *face = g_object_new (PANGO_FC_TYPE_FACE, nil);
//    face.style = g_strdup (style);
//    if (pattern)
// 	 FcPatternReference (pattern);
//    face.pattern = pattern;
//    face.family = fcfamily;
//    face.fake = fake;

//    return face;
//  }

//  static int
//  compare_face (const void *p1, const void *p2)
//  {
//    const PangoFcFace *f1 = *(const void **)p1;
//    const PangoFcFace *f2 = *(const void **)p2;
//    int w1, w2;
//    int s1, s2;

//    if (FcPatternGetInteger (f1.pattern, FC_WEIGHT, 0, &w1) != FcResultMatch)
// 	 w1 = FC_WEIGHT_MEDIUM;

//    if (FcPatternGetInteger (f1.pattern, FC_SLANT, 0, &s1) != FcResultMatch)
// 	 s1 = FC_SLANT_ROMAN;

//    if (FcPatternGetInteger (f2.pattern, FC_WEIGHT, 0, &w2) != FcResultMatch)
// 	 w2 = FC_WEIGHT_MEDIUM;

//    if (FcPatternGetInteger (f2.pattern, FC_SLANT, 0, &s2) != FcResultMatch)
// 	 s2 = FC_SLANT_ROMAN;

//    if (s1 != s2)
// 	 return s1 - s2; /* roman < italic < oblique */

//    return w1 - w2; /* from light to heavy */
//  }

//  static void
//  ensure_faces (PangoFcFamily *fcfamily)
//  {
//     fcfontmap *PangoFcFontMap = fcfamily.fontmap;
//    PangoFcFontMapPrivate *priv = fcfontmap.priv;

//    if (fcfamily.n_faces < 0)
// 	 {
// 	   FcFontSet *fontset;
// 	   int i;

// 	   if (is_alias_family (fcfamily.family_name) || priv.closed)
// 	 {
// 	   fcfamily.n_faces = 4;
// 	   fcfamily.faces = g_new (PangoFcFace *, fcfamily.n_faces);

// 	   i = 0;
// 	   fcfamily.faces[i++] = create_face (fcfamily, "Regular", nil, true);
// 	   fcfamily.faces[i++] = create_face (fcfamily, "Bold", nil, true);
// 	   fcfamily.faces[i++] = create_face (fcfamily, "Italic", nil, true);
// 	   fcfamily.faces[i++] = create_face (fcfamily, "Bold Italic", nil, true);
// 		   fcfamily.faces[0].regular = 1;
// 	 }
// 	   else
// 	 {
// 	   enum {
// 		 REGULAR,
// 		 ITALIC,
// 		 BOLD,
// 		 BOLD_ITALIC
// 	   };
// 	   /* Regular, Italic, Bold, Bold Italic */
// 	   bool has_face [4] = { false, false, false, false };
// 	   PangoFcFace **faces;
// 	   gint num = 0;
// 		   int regular_weight;
// 		   int regular_idx;

// 	   fontset = fcfamily.patterns;

// 	   /* at most we have 3 additional artifical faces */
// 	   faces = g_new (PangoFcFace *, fontset.nfont + 3);

// 		   regular_weight = 0;
// 		   regular_idx = -1;

// 	   for (i = 0; i < fontset.nfont; i++)
// 		 {
// 		   const char *style, *font_style = nil;
// 		   int weight, slant;

// 		   if (FcPatternGetInteger(fontset.fonts[i], FC_WEIGHT, 0, &weight) != FcResultMatch)
// 		 weight = FC_WEIGHT_MEDIUM;

// 		   if (FcPatternGetInteger(fontset.fonts[i], FC_SLANT, 0, &slant) != FcResultMatch)
// 		 slant = FC_SLANT_ROMAN;

//  #ifdef FC_VARIABLE
// 			   {
// 				 bool variable;
// 				 if (FcPatternGetBool(fontset.fonts[i], FC_VARIABLE, 0, &variable) != FcResultMatch)
// 				   variable = false;
// 				 if (variable) /* skip the variable face */
// 				   continue;
// 			   }
//  #endif

// 		   if (FcPatternGetString (fontset.fonts[i], FC_STYLE, 0, (FcChar8 **)(void*)&font_style) != FcResultMatch)
// 		 font_style = nil;

// 			   if (font_style && strcmp (font_style, "Regular") == 0)
// 				 {
// 				   regular_weight = FC_WEIGHT_MEDIUM;
// 				   regular_idx = num;
// 				 }

// 		   if (weight <= FC_WEIGHT_MEDIUM)
// 		 {
// 		   if (slant == FC_SLANT_ROMAN)
// 			 {
// 			   has_face[REGULAR] = true;
// 			   style = "Regular";
// 					   if (weight > regular_weight)
// 						 {
// 						   regular_weight = weight;
// 						   regular_idx = num;
// 						 }
// 			 }
// 		   else
// 			 {
// 			   has_face[ITALIC] = true;
// 			   style = "Italic";
// 			 }
// 		 }
// 		   else
// 		 {
// 		   if (slant == FC_SLANT_ROMAN)
// 			 {
// 			   has_face[BOLD] = true;
// 			   style = "Bold";
// 			 }
// 		   else
// 			 {
// 			   has_face[BOLD_ITALIC] = true;
// 			   style = "Bold Italic";
// 			 }
// 		 }

// 		   if (!font_style)
// 		 font_style = style;
// 		   faces[num++] = create_face (fcfamily, font_style, fontset.fonts[i], false);
// 		 }

// 	   if (has_face[REGULAR])
// 		 {
// 		   if (!has_face[ITALIC])
// 		 faces[num++] = create_face (fcfamily, "Italic", nil, true);
// 		   if (!has_face[BOLD])
// 		 faces[num++] = create_face (fcfamily, "Bold", nil, true);

// 		 }
// 	   if ((has_face[REGULAR] || has_face[ITALIC] || has_face[BOLD]) && !has_face[BOLD_ITALIC])
// 		 faces[num++] = create_face (fcfamily, "Bold Italic", nil, true);

// 		   if (regular_idx != -1)
// 			 faces[regular_idx].regular = 1;

// 	   faces = g_renew (PangoFcFace *, faces, num);

// 		   qsort (faces, num, sizeof (PangoFcFace *), compare_face);

// 	   fcfamily.n_faces = num;
// 	   fcfamily.faces = faces;
// 	 }
// 	 }
//  }

//  static void
//  pango_fc_family_list_faces (PangoFontFamily  *family,
// 				 PangoFontFace  ***faces,
// 				 int              *n_faces)
//  {
//    PangoFcFamily *fcfamily = PANGO_FC_FAMILY (family);

//    if (faces)
// 	 *faces = nil;

//    if (n_faces)
// 	 *n_faces = 0;

//    if (G_UNLIKELY (!fcfamily.fontmap))
// 	 return;

//    ensure_faces (fcfamily);

//    if (n_faces)
// 	 *n_faces = fcfamily.n_faces;

//    if (faces)
// 	 *faces = g_memdup (fcfamily.faces, fcfamily.n_faces * sizeof (PangoFontFace *));
//  }

//  static PangoFontFace *
//  pango_fc_family_get_face (PangoFontFamily *family,
// 						   const char      *name)
//  {
//    PangoFcFamily *fcfamily = PANGO_FC_FAMILY (family);
//    int i;

//    ensure_faces (fcfamily);

//    for (i = 0; i < fcfamily.n_faces; i++)
// 	 {
// 	   PangoFontFace *face = PANGO_FONT_FACE (fcfamily.faces[i]);

// 	   if ((name != nil && strcmp (name, pango_font_face_get_face_name (face)) == 0) ||
// 		   (name == nil && PANGO_FC_FACE (face).regular))
// 		 return face;
// 	 }

//    return nil;
//  }

//  static const char *
//  pango_fc_family_get_name (PangoFontFamily  *family)
//  {
//    PangoFcFamily *fcfamily = PANGO_FC_FAMILY (family);

//    return fcfamily.family_name;
//  }

//  static bool
//  pango_fc_family_is_monospace (PangoFontFamily *family)
//  {
//    PangoFcFamily *fcfamily = PANGO_FC_FAMILY (family);

//    return fcfamily.spacing == FC_MONO ||
// 	  fcfamily.spacing == FC_DUAL ||
// 	  fcfamily.spacing == FC_CHARCELL;
//  }

//  static bool
//  pango_fc_family_is_variable (PangoFontFamily *family)
//  {
//    PangoFcFamily *fcfamily = PANGO_FC_FAMILY (family);

//    return fcfamily.variable;
//  }

//  static void
//  pango_fc_family_finalize (GObject *object)
//  {
//    int i;
//    PangoFcFamily *fcfamily = PANGO_FC_FAMILY (object);

//    g_free (fcfamily.family_name);

//    for (i = 0; i < fcfamily.n_faces; i++)
// 	 {
// 	   fcfamily.faces[i].family = nil;
// 	   g_object_unref (fcfamily.faces[i]);
// 	 }
//    FcFontSetDestroy (fcfamily.patterns);
//    g_free (fcfamily.faces);

//    G_OBJECT_CLASS (pango_fc_family_parent_class).finalize (object);
//  }

//  static void
//  pango_fc_family_class_init (PangoFcFamilyClass *class)
//  {
//    GObjectClass *object_class = G_OBJECT_CLASS (class);

//    object_class.finalize = pango_fc_family_finalize;

//    class.list_faces = pango_fc_family_list_faces;
//    class.get_face = pango_fc_family_get_face;
//    class.get_name = pango_fc_family_get_name;
//    class.is_monospace = pango_fc_family_is_monospace;
//    class.is_variable = pango_fc_family_is_variable;
//  }

//  static void
//  pango_fc_family_init (PangoFcFamily *fcfamily)
//  {
//    fcfamily.n_faces = -1;
//  }

//  /**
//   * pango_fc_font_map_get_hb_face: (skip)
//   * @fcfontmap: a #PangoFcFontMap
//   * @fcfont: a #PangoFcFont
//   *
//   * Retrieves the `hb_face_t` for the given #PangoFcFont.
//   *
//   * Returns: (transfer none) (nullable): the `hb_face_t` for the given Pango font
//   *
//   * Since: 1.44
//   */
//  hb_face_t *
//  pango_fc_font_map_get_hb_face ( fcfontmap *PangoFcFontMap,
// 								PangoFcFont    *fcfont)
//  {
//    PangoFcFontFaceData *data;

//    data = pango_fc_font_map_get_font_face_data (fcfontmap, fcfont.font_pattern);

//    if (!data.hb_face)
// 	 {
// 	   hb_blob_t *blob;

// 	   if (!hb_version_atleast (2, 0, 0))
// 		 g_error ("Harfbuzz version too old (%s)\n", hb_version_string ());

// 	   blob = hb_blob_create_from_file (data.filename);
// 	   data.hb_face = hb_face_create (blob, data.id);
// 	   hb_blob_destroy (blob);
// 	 }

//    return data.hb_face;
//  }
