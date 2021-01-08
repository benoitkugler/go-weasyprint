package fontconfig

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// ported from fontconfig/src/fcname.c Copyright © 2000 Keith Packard

// used to identify a type
type typeMeta interface {
	parse(str string, object FcObject) (FcValue, error)
}

type FcObjectType struct {
	object FcObject // exposed name of the object
	parser typeMeta
}

var objects = map[string]FcObjectType{
	ObjectNames[FC_FAMILY]:          {object: FC_FAMILY, parser: typeString{}},          // String
	ObjectNames[FC_FAMILYLANG]:      {object: FC_FAMILYLANG, parser: typeString{}},      // String
	ObjectNames[FC_STYLE]:           {object: FC_STYLE, parser: typeString{}},           // String
	ObjectNames[FC_STYLELANG]:       {object: FC_STYLELANG, parser: typeString{}},       // String
	ObjectNames[FC_FULLNAME]:        {object: FC_FULLNAME, parser: typeString{}},        // String
	ObjectNames[FC_FULLNAMELANG]:    {object: FC_FULLNAMELANG, parser: typeString{}},    // String
	ObjectNames[FC_SLANT]:           {object: FC_SLANT, parser: typeInteger{}},          // Integer
	ObjectNames[FC_WEIGHT]:          {object: FC_WEIGHT, parser: typeRange{}},           // Range
	ObjectNames[FC_WIDTH]:           {object: FC_WIDTH, parser: typeRange{}},            // Range
	ObjectNames[FC_SIZE]:            {object: FC_SIZE, parser: typeRange{}},             // Range
	ObjectNames[FC_ASPECT]:          {object: FC_ASPECT, parser: typeFloat{}},           // Double
	ObjectNames[FC_PIXEL_SIZE]:      {object: FC_PIXEL_SIZE, parser: typeFloat{}},       // Double
	ObjectNames[FC_SPACING]:         {object: FC_SPACING, parser: typeInteger{}},        // Integer
	ObjectNames[FC_FOUNDRY]:         {object: FC_FOUNDRY, parser: typeString{}},         // String
	ObjectNames[FC_ANTIALIAS]:       {object: FC_ANTIALIAS, parser: typeBool{}},         // Bool
	ObjectNames[FC_HINT_STYLE]:      {object: FC_HINT_STYLE, parser: typeInteger{}},     // Integer
	ObjectNames[FC_HINTING]:         {object: FC_HINTING, parser: typeBool{}},           // Bool
	ObjectNames[FC_VERTICAL_LAYOUT]: {object: FC_VERTICAL_LAYOUT, parser: typeBool{}},   // Bool
	ObjectNames[FC_AUTOHINT]:        {object: FC_AUTOHINT, parser: typeBool{}},          // Bool
	ObjectNames[FC_GLOBAL_ADVANCE]:  {object: FC_GLOBAL_ADVANCE, parser: typeBool{}},    // Bool
	ObjectNames[FC_FILE]:            {object: FC_FILE, parser: typeString{}},            // String
	ObjectNames[FC_INDEX]:           {object: FC_INDEX, parser: typeInteger{}},          // Integer
	ObjectNames[FC_RASTERIZER]:      {object: FC_RASTERIZER, parser: typeString{}},      // String
	ObjectNames[FC_OUTLINE]:         {object: FC_OUTLINE, parser: typeBool{}},           // Bool
	ObjectNames[FC_SCALABLE]:        {object: FC_SCALABLE, parser: typeBool{}},          // Bool
	ObjectNames[FC_DPI]:             {object: FC_DPI, parser: typeFloat{}},              // Double
	ObjectNames[FC_RGBA]:            {object: FC_RGBA, parser: typeInteger{}},           // Integer
	ObjectNames[FC_SCALE]:           {object: FC_SCALE, parser: typeFloat{}},            // Double
	ObjectNames[FC_MINSPACE]:        {object: FC_MINSPACE, parser: typeBool{}},          // Bool
	ObjectNames[FC_CHARWIDTH]:       {object: FC_CHARWIDTH, parser: typeInteger{}},      // Integer
	ObjectNames[FC_CHAR_HEIGHT]:     {object: FC_CHAR_HEIGHT, parser: typeInteger{}},    // Integer
	ObjectNames[FC_MATRIX]:          {object: FC_MATRIX, parser: typeMatrix{}},          // Matrix
	ObjectNames[FC_CHARSET]:         {object: FC_CHARSET, parser: typeCharSet{}},        // CharSet
	ObjectNames[FC_LANG]:            {object: FC_LANG, parser: typeLangSet{}},           // LangSet
	ObjectNames[FC_FONTVERSION]:     {object: FC_FONTVERSION, parser: typeInteger{}},    // Integer
	ObjectNames[FC_CAPABILITY]:      {object: FC_CAPABILITY, parser: typeString{}},      // String
	ObjectNames[FC_FONTFORMAT]:      {object: FC_FONTFORMAT, parser: typeString{}},      // String
	ObjectNames[FC_EMBOLDEN]:        {object: FC_EMBOLDEN, parser: typeBool{}},          // Bool
	ObjectNames[FC_EMBEDDED_BITMAP]: {object: FC_EMBEDDED_BITMAP, parser: typeBool{}},   // Bool
	ObjectNames[FC_DECORATIVE]:      {object: FC_DECORATIVE, parser: typeBool{}},        // Bool
	ObjectNames[FC_LCD_FILTER]:      {object: FC_LCD_FILTER, parser: typeInteger{}},     // Integer
	ObjectNames[FC_NAMELANG]:        {object: FC_NAMELANG, parser: typeString{}},        // String
	ObjectNames[FC_FONT_FEATURES]:   {object: FC_FONT_FEATURES, parser: typeString{}},   // String
	ObjectNames[FC_PRGNAME]:         {object: FC_PRGNAME, parser: typeString{}},         // String
	ObjectNames[FC_HASH]:            {object: FC_HASH, parser: typeString{}},            // String
	ObjectNames[FC_POSTSCRIPT_NAME]: {object: FC_POSTSCRIPT_NAME, parser: typeString{}}, // String
	ObjectNames[FC_COLOR]:           {object: FC_COLOR, parser: typeBool{}},             // Bool
	ObjectNames[FC_SYMBOL]:          {object: FC_SYMBOL, parser: typeBool{}},            // Bool
	ObjectNames[FC_FONT_VARIATIONS]: {object: FC_FONT_VARIATIONS, parser: typeString{}}, // String
	ObjectNames[FC_VARIABLE]:        {object: FC_VARIABLE, parser: typeBool{}},          // Bool
	ObjectNames[FC_FONT_HAS_HINT]:   {object: FC_FONT_HAS_HINT, parser: typeBool{}},     // Bool
	ObjectNames[FC_ORDER]:           {object: FC_ORDER, parser: typeInteger{}},          // Integer
}

//  static const FcObjectType FcObjects[] = {
//  #define FC_OBJECT(NAME, Type, Cmp) { FC_##NAME, Type },
//  #include "fcobjs.h"
//  #undef FC_OBJECT
//  };

//  #define NUM_OBJECT_TYPES ((int) (sizeof FcObjects / sizeof FcObjects[0]))

//  static const FcObjectType *
//  FcObjectFindById (FcObject object)
//  {
// 	 if (1 <= object && object <= NUM_OBJECT_TYPES)
// 	 return &FcObjects[object - 1];
// 	 return FcObjectLookupOtherTypeById (object);
//  }

//  FcBool
//  FcNameRegisterObjectTypes (const FcObjectType *types, int ntypes)
//  {
// 	 /* Deprecated. */
// 	 return false;
//  }

//  FcBool
//  FcNameUnregisterObjectTypes (const FcObjectType *types, int ntypes)
//  {
// 	 /* Deprecated. */
// 	 return false;
//  }

// Return the object type for the pattern element named object
func getObjectType(object string) (FcObjectType, error) {
	if builtin, ok := objects[object]; ok {
		return builtin, nil
	}
	// TODO: support custom objects
	return FcObjectType{}, fmt.Errorf("fontconfig: invalid object name %s", object)
}

//  FcBool
//  hasValidType (FcObject object, FcType type)
//  {
// 	 const FcObjectType    *t = FcObjectFindById (object);

// 	 if (t) {
// 	 switch ((int) t.type) {
// 	 case FcTypeUnknown:
// 		 return true;
// 	 case FcTypeDouble:
// 	 case FcTypeInteger:
// 		 if (type == FcTypeDouble || type == FcTypeInteger)
// 		 return true;
// 		 break;
// 	 case FcTypeLangSet:
// 		 if (type == FcTypeLangSet || type == FcTypeString)
// 		 return true;
// 		 break;
// 	 case FcTypeRange:
// 		 if (type == FcTypeRange ||
// 		 type == FcTypeDouble ||
// 		 type == FcTypeInteger)
// 		 return true;
// 		 break;
// 	 default:
// 		 if (type == t.type)
// 		 return true;
// 		 break;
// 	 }
// 	 return false;
// 	 }
// 	 return true;
//  }

//  FcObject
//  FcObjectFromName (const char * name)
//  {
// 	 return FcObjectLookupIdByName (name);
//  }

//  FcObjectSet *
//  FcObjectGetSet (void)
//  {
// 	 int		i;
// 	 FcObjectSet	*os = NULL;

// 	 os = FcObjectSetCreate ();
// 	 for (i = 0; i < NUM_OBJECT_TYPES; i++)
// 	 FcObjectSetAdd (os, FcObjects[i].object);

// 	 return os;
//  }

//  const char *
//  FcObjectName (FcObject object)
//  {
// 	 const FcObjectType   *o = FcObjectFindById (object);

// 	 if (o)
// 	 return o.object;

// 	 return FcObjectLookupOtherNameById (object);
//  }

type FcConstant struct {
	name   string
	object FcObject
	value  int
}

var baseConstants = [...]FcConstant{
	{"thin", FC_WEIGHT, FC_WEIGHT_THIN},
	{"extralight", FC_WEIGHT, FC_WEIGHT_EXTRALIGHT},
	{"ultralight", FC_WEIGHT, FC_WEIGHT_EXTRALIGHT},
	{"demilight", FC_WEIGHT, FC_WEIGHT_DEMILIGHT},
	{"semilight", FC_WEIGHT, FC_WEIGHT_DEMILIGHT},
	{"light", FC_WEIGHT, FC_WEIGHT_LIGHT},
	{"book", FC_WEIGHT, FC_WEIGHT_BOOK},
	{"regular", FC_WEIGHT, FC_WEIGHT_REGULAR},
	{"medium", FC_WEIGHT, FC_WEIGHT_MEDIUM},
	{"demibold", FC_WEIGHT, FC_WEIGHT_DEMIBOLD},
	{"semibold", FC_WEIGHT, FC_WEIGHT_DEMIBOLD},
	{"bold", FC_WEIGHT, FC_WEIGHT_BOLD},
	{"extrabold", FC_WEIGHT, FC_WEIGHT_EXTRABOLD},
	{"ultrabold", FC_WEIGHT, FC_WEIGHT_EXTRABOLD},
	{"black", FC_WEIGHT, FC_WEIGHT_BLACK},
	{"heavy", FC_WEIGHT, FC_WEIGHT_HEAVY},

	{"roman", FC_SLANT, FC_SLANT_ROMAN},
	{"italic", FC_SLANT, FC_SLANT_ITALIC},
	{"oblique", FC_SLANT, FC_SLANT_OBLIQUE},

	{"ultracondensed", FC_WIDTH, FC_WIDTH_ULTRACONDENSED},
	{"extracondensed", FC_WIDTH, FC_WIDTH_EXTRACONDENSED},
	{"condensed", FC_WIDTH, FC_WIDTH_CONDENSED},
	{"semicondensed", FC_WIDTH, FC_WIDTH_SEMICONDENSED},
	{"normal", FC_WIDTH, FC_WIDTH_NORMAL},
	{"semiexpanded", FC_WIDTH, FC_WIDTH_SEMIEXPANDED},
	{"expanded", FC_WIDTH, FC_WIDTH_EXPANDED},
	{"extraexpanded", FC_WIDTH, FC_WIDTH_EXTRAEXPANDED},
	{"ultraexpanded", FC_WIDTH, FC_WIDTH_ULTRAEXPANDED},

	{"proportional", FC_SPACING, FC_PROPORTIONAL},
	{"dual", FC_SPACING, FC_DUAL},
	{"mono", FC_SPACING, FC_MONO},
	{"charcell", FC_SPACING, FC_CHARCELL},

	{"unknown", FC_RGBA, FC_RGBA_UNKNOWN},
	{"rgb", FC_RGBA, FC_RGBA_RGB},
	{"bgr", FC_RGBA, FC_RGBA_BGR},
	{"vrgb", FC_RGBA, FC_RGBA_VRGB},
	{"vbgr", FC_RGBA, FC_RGBA_VBGR},
	{"none", FC_RGBA, FC_RGBA_NONE},

	{"hintnone", FC_HINT_STYLE, FC_HINT_NONE},
	{"hintslight", FC_HINT_STYLE, FC_HINT_SLIGHT},
	{"hintmedium", FC_HINT_STYLE, FC_HINT_MEDIUM},
	{"hintfull", FC_HINT_STYLE, FC_HINT_FULL},

	{"antialias", FC_ANTIALIAS, 1},
	{"hinting", FC_HINTING, 1},
	{"verticallayout", FC_VERTICAL_LAYOUT, 1},
	{"autohint", FC_AUTOHINT, 1},
	{"globaladvance", FC_GLOBAL_ADVANCE, 1}, /* deprecated */
	{"outline", FC_OUTLINE, 1},
	{"scalable", FC_SCALABLE, 1},
	{"minspace", FC_MINSPACE, 1},
	{"embolden", FC_EMBOLDEN, 1},
	{"embeddedbitmap", FC_EMBEDDED_BITMAP, 1},
	{"decorative", FC_DECORATIVE, 1},

	{"lcdnone", FC_LCD_FILTER, FC_LCD_NONE},
	{"lcddefault", FC_LCD_FILTER, FC_LCD_DEFAULT},
	{"lcdlight", FC_LCD_FILTER, FC_LCD_LIGHT},
	{"lcdlegacy", FC_LCD_FILTER, FC_LCD_LEGACY},
}

//  #define NUM_FC_CONSTANTS   (sizeof baseConstants/sizeof baseConstants[0])

//  FcBool
//  FcNameRegisterConstants (const FcConstant *consts, int nconsts)
//  {
// 	 /* Deprecated. */
// 	 return false;
//  }

//  FcBool
//  FcNameUnregisterConstants (const FcConstant *consts, int nconsts)
//  {
// 	 /* Deprecated. */
// 	 return false;
//  }

func FcNameGetConstant(str string) *FcConstant {
	for i := range baseConstants {
		if FcStrCmpIgnoreCase(str, baseConstants[i].name) == 0 {
			return &baseConstants[i]
		}
	}
	return nil
}

func FcNameConstant(str string) (int, bool) {
	if c := FcNameGetConstant(str); c != nil {
		return c.value, true
	}
	return 0, false
}

// FcNameParse converts `name` from the standard text format described above into a pattern.
func FcNameParse(name []byte) (*FcPattern, error) {
	var (
		delim byte
		save  string
		pat   = FcPattern{elts: make(map[FcObject]FcValueList)}
	)

	for {
		delim, name, save = nameFindNext(name, "-,:")
		if len(save) != 0 {
			pat.Add(FC_FAMILY, save, true)
		}
		if delim != ',' {
			break
		}
	}
	if delim == '-' {
		for {
			delim, name, save = nameFindNext(name, "-,:")
			d, err := strconv.ParseFloat(save, 64)
			if err == nil {
				pat.Add(FC_SIZE, d, true)
			}
			if delim != ',' {
				break
			}
		}
	}
	for delim == ':' {
		delim, name, save = nameFindNext(name, "=_:")
		if len(save) != 0 {
			if delim == '=' || delim == '_' {
				t, err := getObjectType(save)
				if err != nil {
					return nil, err
				}
				for {
					delim, name, save = nameFindNext(name, ":,")
					v, err := t.parser.parse(save, t.object)
					if err != nil {
						return nil, err
					}
					pat.Add(t.object, v, true)
					if delim != ',' {
						break
					}
				}
			} else {
				if c := FcNameGetConstant(save); c != nil {
					t, err := getObjectType(ObjectNames[c.object])
					if err != nil {
						return nil, err
					}

					switch t.parser.(type) {
					case typeInteger, typeFloat, typeRange:
						pat.Add(c.object, c.value, true)
					case typeBool:
						pat.Add(c.object, FcBool(c.value), true)
					}
				}
			}
		}
	}

	return &pat, nil
}

func nameFindNext(cur []byte, delim string) (byte, []byte, string) {
	cur = bytes.TrimLeftFunc(cur, unicode.IsSpace)
	i := 0
	var save []byte
	for i < len(cur) {
		if cur[i] == '\\' {
			i++
			if i == len(cur) {
				break
			}
		} else if strings.IndexByte(delim, cur[i]) != -1 {
			break
		}
		save = append(save, cur[i])
		i++
	}
	var last byte
	if i < len(cur) {
		last = cur[i]
		i++
	}
	return last, cur[i:], string(save)
}

func constantWithObjectCheck(str string, object FcObject) (int, bool, error) {
	c := FcNameGetConstant(str)
	if c != nil {
		if c.object != object {
			return 0, false, fmt.Errorf("fontconfig : unexpected constant name %s used for object %s: should be %s\n", str, object, c.object)
		}
		return c.value, true, nil
	}
	return 0, false, nil
}

func nameBool(v string) (FcBool, error) {
	c0 := FcToLower(v)
	if c0 == 't' || c0 == 'y' || c0 == '1' {
		return FcTrue, nil
	}
	if c0 == 'f' || c0 == 'n' || c0 == '0' {
		return FcFalse, nil
	}
	if c0 == 'd' || c0 == 'x' || c0 == '2' {
		return FcDontCare, nil
	}
	if c0 == 'o' {
		c1 := FcToLower(v[1:])
		if c1 == 'n' {
			return FcTrue, nil
		}
		if c1 == 'f' {
			return FcFalse, nil
		}
		if c1 == 'r' {
			return FcDontCare, nil
		}
	}
	return 0, fmt.Errorf("fontconfig: unknown boolean %s", v)
}

type typeInteger struct{}

func (typeInteger) parse(str string, object FcObject) (FcValue, error) {
	v, builtin, err := constantWithObjectCheck(str, object)
	if err != nil {
		return nil, err
	}
	if !builtin {
		v, err = strconv.Atoi(str)
	}
	return v, err
}

type typeString struct{}

func (typeString) parse(str string, object FcObject) (FcValue, error) { return str, nil }

type typeBool struct{}

func (typeBool) parse(str string, object FcObject) (FcValue, error) { return nameBool(str) }

type typeFloat struct{}

func (typeFloat) parse(str string, object FcObject) (FcValue, error) {
	return strconv.ParseFloat(str, 64)
}

type typeMatrix struct{}

func (typeMatrix) parse(str string, object FcObject) (FcValue, error) {
	var m FcMatrix
	_, err := fmt.Sscanf(str, "%g %g %g %g", &m.xx, &m.xy, &m.yx, &m.yy)
	return m, err
}

type typeCharSet struct{}

func (typeCharSet) parse(str string, object FcObject) (FcValue, error) {
	return FcNameParseCharSet(str)
}

type typeLangSet struct{}

func (typeLangSet) parse(str string, object FcObject) (FcValue, error) {
	return FcNameParseLangSet(str), nil
}

type typeRange struct{}

func (typeRange) parse(str string, object FcObject) (FcValue, error) {
	var b, e float64
	n, _ := fmt.Sscanf(str, "[%g %g]", &b, &e)
	if n == 2 {
		return FcRange{Begin: b, End: e}, nil
	}

	var sc, ec string
	n, _ = fmt.Sscanf(strings.TrimSuffix(str, "]"), "[%s %s", &sc, &ec)
	if n == 2 {
		si, oks, err := constantWithObjectCheck(sc, object)
		if err != nil {
			return nil, err
		}
		ei, oke, err := constantWithObjectCheck(ec, object)
		if err != nil {
			return nil, err
		}
		if oks && oke {
			return FcRange{Begin: float64(si), End: float64(ei)}, nil
		}
	}

	si, ok, err := constantWithObjectCheck(str, object)
	if err != nil {
		return nil, err
	}
	if ok {
		return float64(si), nil
	}
	v, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return nil, err
	}
	return v, nil
}

//  static FcValue
//  FcNameConvert (FcType type, const char *object, FcChar8 *string)
//  {
// 	 FcValue	v;
// 	 FcMatrix	m;
// 	 double	b, e;
// 	 char	*p;

// 	 v.type = type;
// 	 switch ((int) v.type) {
// 	 case FcTypeInteger:
// 	 if (!constantWithObjectCheck (string, object, &v.u.i))
// 		 v.u.i = atoi ((char *) string);
// 	 break;
// 	 case FcTypeString:
// 	 v.u.s = FcStrdup (string);
// 	 if (!v.u.s)
// 		 v.type = FcTypeVoid;
// 	 break;
// 	 case FcTypeBool:
// 	 if (!nameBool (string, &v.u.b))
// 		 v.u.b = false;
// 	 break;
// 	 case FcTypeDouble:
// 	 v.u.d = strtod ((char *) string, 0);
// 	 break;
// 	 case FcTypeMatrix:
// 	 FcMatrixInit (&m);
// 	 sscanf ((char *) string, "%lg %lg %lg %lg", &m.xx, &m.xy, &m.yx, &m.yy);
// 	 v.u.m = FcMatrixCopy (&m);
// 	 break;
// 	 case FcTypeCharSet:
// 	 v.u.c = FcNameParseCharSet (string);
// 	 if (!v.u.c)
// 		 v.type = FcTypeVoid;
// 	 break;
// 	 case FcTypeLangSet:
// 	 v.u.l = FcNameParseLangSet (string);
// 	 if (!v.u.l)
// 		 v.type = FcTypeVoid;
// 	 break;
// 	 case FcTypeRange:
// 	 if (sscanf ((char *) string, "[%lg %lg]", &b, &e) != 2)
// 	 {
// 		 char *sc, *ec;
// 		 size_t len = strlen ((const char *) string);
// 		 int si, ei;

// 		 sc = malloc (len + 1);
// 		 ec = malloc (len + 1);
// 		 if (sc && ec && sscanf ((char *) string, "[%s %[^]]]", sc, ec) == 2)
// 		 {
// 		 if (constantWithObjectCheck ((const FcChar8 *) sc, object, &si) &&
// 			 constantWithObjectCheck ((const FcChar8 *) ec, object, &ei))
// 			 v.u.r =  FcRangeCreateDouble (si, ei);
// 		 else
// 			 goto bail1;
// 		 }
// 		 else
// 		 {
// 		 bail1:
// 		 v.type = FcTypeDouble;
// 		 if (constantWithObjectCheck (string, object, &si))
// 		 {
// 			 v.u.d = (double) si;
// 		 } else {
// 			 v.u.d = strtod ((char *) string, &p);
// 			 if (p != NULL && p[0] != 0)
// 			 v.type = FcTypeVoid;
// 		 }
// 		 }
// 		 if (sc)
// 		 free (sc);
// 		 if (ec)
// 		 free (ec);
// 	 }
// 	 else
// 		 v.u.r = FcRangeCreateDouble (b, e);
// 	 break;
// 	 default:
// 	 break;
// 	 }
// 	 return v;
//  }

//  static FcBool
//  FcNameUnparseString (FcStrBuf	    *buf,
// 			  const FcChar8  *string,
// 			  const FcChar8  *escape)
//  {
// 	 FcChar8 c;
// 	 for ((c = *string++))
// 	 {
// 	 if (escape && strchr ((char *) escape, (char) c))
// 	 {
// 		 if (!FcStrBufChar (buf, escape[0]))
// 		 return false;
// 	 }
// 	 if (!FcStrBufChar (buf, c))
// 		 return false;
// 	 }
// 	 return true;
//  }

//  FcBool
//  FcNameUnparseValue (FcStrBuf	*buf,
// 			 FcValue	*v0,
// 			 FcChar8	*escape)
//  {
// 	 FcChar8	temp[1024];
// 	 FcValue v = FcValueCanonicalize(v0);

// 	 switch (v.type) {
// 	 case FcTypeUnknown:
// 	 case FcTypeVoid:
// 	 return true;
// 	 case FcTypeInteger:
// 	 sprintf ((char *) temp, "%d", v.u.i);
// 	 return FcNameUnparseString (buf, temp, 0);
// 	 case FcTypeDouble:
// 	 sprintf ((char *) temp, "%g", v.u.d);
// 	 return FcNameUnparseString (buf, temp, 0);
// 	 case FcTypeString:
// 	 return FcNameUnparseString (buf, v.u.s, escape);
// 	 case FcTypeBool:
// 	 return FcNameUnparseString (buf,
// 					 v.u.b == true  ? (FcChar8 *) "True" :
// 					 v.u.b == false ? (FcChar8 *) "False" :
// 										(FcChar8 *) "DontCare", 0);
// 	 case FcTypeMatrix:
// 	 sprintf ((char *) temp, "%g %g %g %g",
// 		  v.u.m.xx, v.u.m.xy, v.u.m.yx, v.u.m.yy);
// 	 return FcNameUnparseString (buf, temp, 0);
// 	 case FcTypeCharSet:
// 	 return FcNameUnparseCharSet (buf, v.u.c);
// 	 case FcTypeLangSet:
// 	 return FcNameUnparseLangSet (buf, v.u.l);
// 	 case FcTypeFTFace:
// 	 return true;
// 	 case FcTypeRange:
// 	 sprintf ((char *) temp, "[%g %g]", v.u.r.begin, v.u.r.end);
// 	 return FcNameUnparseString (buf, temp, 0);
// 	 }
// 	 return false;
//  }

//  FcBool
//  FcNameUnparseValueList (FcStrBuf	*buf,
// 			 FcValueListPtr	v,
// 			 FcChar8		*escape)
//  {
// 	 for (v)
// 	 {
// 	 if (!FcNameUnparseValue (buf, &v.value, escape))
// 		 return false;
// 	 if ((v = FcValueListNext(v)) != NULL)
// 		 if (!FcNameUnparseString (buf, (FcChar8 *) ",", 0))
// 		 return false;
// 	 }
// 	 return true;
//  }

//  #define FC_ESCAPE_FIXED    "\\-:,"
//  #define FC_ESCAPE_VARIABLE "\\=_:,"

//  FcChar8 *
//  FcNameUnparse (FcPattern *pat)
//  {
// 	 return FcNameUnparseEscaped (pat, true);
//  }

//  FcChar8 *
//  FcNameUnparseEscaped (FcPattern *pat, FcBool escape)
//  {
// 	 FcStrBuf		    buf, buf2;
// 	 FcChar8		    buf_static[8192], buf2_static[256];
// 	 int			    i;
// 	 FcPatternElt	    *e;

// 	 FcStrBufInit (&buf, buf_static, sizeof (buf_static));
// 	 FcStrBufInit (&buf2, buf2_static, sizeof (buf2_static));
// 	 e = FcPatternObjectFindElt (pat, FC_FAMILY_OBJECT);
// 	 if (e)
// 	 {
// 		 if (!FcNameUnparseValueList (&buf, FcPatternEltValues(e), escape ? (FcChar8 *) FC_ESCAPE_FIXED : 0))
// 		 goto bail0;
// 	 }
// 	 e = FcPatternObjectFindElt (pat, FC_SIZE_OBJECT);
// 	 if (e)
// 	 {
// 	 FcChar8 *p;

// 	 if (!FcNameUnparseString (&buf2, (FcChar8 *) "-", 0))
// 		 goto bail0;
// 	 if (!FcNameUnparseValueList (&buf2, FcPatternEltValues(e), escape ? (FcChar8 *) FC_ESCAPE_FIXED : 0))
// 		 goto bail0;
// 	 p = FcStrBufDoneStatic (&buf2);
// 	 FcStrBufDestroy (&buf2);
// 	 if (strlen ((const char *)p) > 1)
// 		 if (!FcStrBufString (&buf, p))
// 		 goto bail0;
// 	 }
// 	 for (i = 0; i < NUM_OBJECT_TYPES; i++)
// 	 {
// 	 FcObject id = i + 1;
// 	 const FcObjectType	    *o;
// 	 o = &FcObjects[i];
// 	 if (!strcmp (o.object, FC_FAMILY) ||
// 		 !strcmp (o.object, FC_SIZE))
// 		 continue;

// 	 e = FcPatternObjectFindElt (pat, id);
// 	 if (e)
// 	 {
// 		 if (!FcNameUnparseString (&buf, (FcChar8 *) ":", 0))
// 		 goto bail0;
// 		 if (!FcNameUnparseString (&buf, (FcChar8 *) o.object, escape ? (FcChar8 *) FC_ESCAPE_VARIABLE : 0))
// 		 goto bail0;
// 		 if (!FcNameUnparseString (&buf, (FcChar8 *) "=", 0))
// 		 goto bail0;
// 		 if (!FcNameUnparseValueList (&buf, FcPatternEltValues(e), escape ?
// 					  (FcChar8 *) FC_ESCAPE_VARIABLE : 0))
// 		 goto bail0;
// 	 }
// 	 }
// 	 return FcStrBufDone (&buf);
//  bail0:
// 	 FcStrBufDestroy (&buf);
// 	 return 0;
//  }
