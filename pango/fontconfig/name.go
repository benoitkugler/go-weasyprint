package fontconfig

// ported from fontconfig/src/fcname.c Copyright © 2000 Keith Packard

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

//  const FcObjectType *
//  FcNameGetObjectType (const char *object)
//  {
// 	 int id = FcObjectLookupBuiltinIdByName (object);

// 	 if (!id)
// 	 return FcObjectLookupOtherTypeByName (object);

// 	 return &FcObjects[id - 1];
//  }

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
	object string
	value  int
}

var baseConstants = [...]FcConstant{
	{"thin", "weight", FC_WEIGHT_THIN},
	{"extralight", "weight", FC_WEIGHT_EXTRALIGHT},
	{"ultralight", "weight", FC_WEIGHT_EXTRALIGHT},
	{"demilight", "weight", FC_WEIGHT_DEMILIGHT},
	{"semilight", "weight", FC_WEIGHT_DEMILIGHT},
	{"light", "weight", FC_WEIGHT_LIGHT},
	{"book", "weight", FC_WEIGHT_BOOK},
	{"regular", "weight", FC_WEIGHT_REGULAR},
	{"medium", "weight", FC_WEIGHT_MEDIUM},
	{"demibold", "weight", FC_WEIGHT_DEMIBOLD},
	{"semibold", "weight", FC_WEIGHT_DEMIBOLD},
	{"bold", "weight", FC_WEIGHT_BOLD},
	{"extrabold", "weight", FC_WEIGHT_EXTRABOLD},
	{"ultrabold", "weight", FC_WEIGHT_EXTRABOLD},
	{"black", "weight", FC_WEIGHT_BLACK},
	{"heavy", "weight", FC_WEIGHT_HEAVY},

	{"roman", "slant", FC_SLANT_ROMAN},
	{"italic", "slant", FC_SLANT_ITALIC},
	{"oblique", "slant", FC_SLANT_OBLIQUE},

	{"ultracondensed", "width", FC_WIDTH_ULTRACONDENSED},
	{"extracondensed", "width", FC_WIDTH_EXTRACONDENSED},
	{"condensed", "width", FC_WIDTH_CONDENSED},
	{"semicondensed", "width", FC_WIDTH_SEMICONDENSED},
	{"normal", "width", FC_WIDTH_NORMAL},
	{"semiexpanded", "width", FC_WIDTH_SEMIEXPANDED},
	{"expanded", "width", FC_WIDTH_EXPANDED},
	{"extraexpanded", "width", FC_WIDTH_EXTRAEXPANDED},
	{"ultraexpanded", "width", FC_WIDTH_ULTRAEXPANDED},

	{"proportional", "spacing", FC_PROPORTIONAL},
	{"dual", "spacing", FC_DUAL},
	{"mono", "spacing", FC_MONO},
	{"charcell", "spacing", FC_CHARCELL},

	{"unknown", "rgba", FC_RGBA_UNKNOWN},
	{"rgb", "rgba", FC_RGBA_RGB},
	{"bgr", "rgba", FC_RGBA_BGR},
	{"vrgb", "rgba", FC_RGBA_VRGB},
	{"vbgr", "rgba", FC_RGBA_VBGR},
	{"none", "rgba", FC_RGBA_NONE},

	{"hintnone", "hintstyle", FC_HINT_NONE},
	{"hintslight", "hintstyle", FC_HINT_SLIGHT},
	{"hintmedium", "hintstyle", FC_HINT_MEDIUM},
	{"hintfull", "hintstyle", FC_HINT_FULL},

	{"antialias", "antialias", 1},
	{"hinting", "hinting", 1},
	{"verticallayout", "verticallayout", 1},
	{"autohint", "autohint", 1},
	{"globaladvance", "globaladvance", 1}, /* deprecated */
	{"outline", "outline", 1},
	{"scalable", "scalable", 1},
	{"minspace", "minspace", 1},
	{"embolden", "embolden", 1},
	{"embeddedbitmap", "embeddedbitmap", 1},
	{"decorative", "decorative", 1},
	{"lcdnone", "lcdfilter", FC_LCD_NONE},
	{"lcddefault", "lcdfilter", FC_LCD_DEFAULT},
	{"lcdlight", "lcdfilter", FC_LCD_LIGHT},
	{"lcdlegacy", "lcdfilter", FC_LCD_LEGACY},
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

//  FcBool
//  FcNameConstantWithObjectCheck (const FcChar8 *string, const char *object, int *result)
//  {
// 	 const FcConstant	*c;

// 	 if ((c = FcNameGetConstant(string)))
// 	 {
// 	 if (strcmp (c.object, object) != 0)
// 	 {
// 		 fprintf (stderr, "Fontconfig error: Unexpected constant name `%s' used for object `%s': should be `%s'\n", string, object, c.object);
// 		 return false;
// 	 }
// 	 *result = c.value;
// 	 return true;
// 	 }
// 	 return false;
//  }

//  FcBool
//  FcNameBool (const FcChar8 *v, FcBool *result)
//  {
// 	 char    c0, c1;

// 	 c0 = *v;
// 	 c0 = FcToLower (c0);
// 	 if (c0 == 't' || c0 == 'y' || c0 == '1')
// 	 {
// 	 *result = true;
// 	 return true;
// 	 }
// 	 if (c0 == 'f' || c0 == 'n' || c0 == '0')
// 	 {
// 	 *result = false;
// 	 return true;
// 	 }
// 	 if (c0 == 'd' || c0 == 'x' || c0 == '2')
// 	 {
// 	 *result = FcDontCare;
// 	 return true;
// 	 }
// 	 if (c0 == 'o')
// 	 {
// 	 c1 = v[1];
// 	 c1 = FcToLower (c1);
// 	 if (c1 == 'n')
// 	 {
// 		 *result = true;
// 		 return true;
// 	 }
// 	 if (c1 == 'f')
// 	 {
// 		 *result = false;
// 		 return true;
// 	 }
// 	 if (c1 == 'r')
// 	 {
// 		 *result = FcDontCare;
// 		 return true;
// 	 }
// 	 }
// 	 return false;
//  }

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
// 	 if (!FcNameConstantWithObjectCheck (string, object, &v.u.i))
// 		 v.u.i = atoi ((char *) string);
// 	 break;
// 	 case FcTypeString:
// 	 v.u.s = FcStrdup (string);
// 	 if (!v.u.s)
// 		 v.type = FcTypeVoid;
// 	 break;
// 	 case FcTypeBool:
// 	 if (!FcNameBool (string, &v.u.b))
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
// 		 if (FcNameConstantWithObjectCheck ((const FcChar8 *) sc, object, &si) &&
// 			 FcNameConstantWithObjectCheck ((const FcChar8 *) ec, object, &ei))
// 			 v.u.r =  FcRangeCreateDouble (si, ei);
// 		 else
// 			 goto bail1;
// 		 }
// 		 else
// 		 {
// 		 bail1:
// 		 v.type = FcTypeDouble;
// 		 if (FcNameConstantWithObjectCheck (string, object, &si))
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

//  static const FcChar8 *
//  FcNameFindNext (const FcChar8 *cur, const char *delim, FcChar8 *save, FcChar8 *last)
//  {
// 	 FcChar8    c;

// 	 while ((c = *cur))
// 	 {
// 	 if (!isspace (c))
// 		 break;
// 	 ++cur;
// 	 }
// 	 while ((c = *cur))
// 	 {
// 	 if (c == '\\')
// 	 {
// 		 ++cur;
// 		 if (!(c = *cur))
// 		 break;
// 	 }
// 	 else if (strchr (delim, c))
// 		 break;
// 	 ++cur;
// 	 *save++ = c;
// 	 }
// 	 *save = 0;
// 	 *last = *cur;
// 	 if (*cur)
// 	 cur++;
// 	 return cur;
//  }

//  FcPattern *
//  FcNameParse (const FcChar8 *name)
//  {
// 	 FcChar8		*save;
// 	 FcPattern		*pat;
// 	 double		d;
// 	 FcChar8		*e;
// 	 FcChar8		delim;
// 	 FcValue		v;
// 	 const FcObjectType	*t;
// 	 const FcConstant	*c;

// 	 /* freed below */
// 	 save = malloc (strlen ((char *) name) + 1);
// 	 if (!save)
// 	 goto bail0;
// 	 pat = FcPatternCreate ();
// 	 if (!pat)
// 	 goto bail1;

// 	 for (;;)
// 	 {
// 	 name = FcNameFindNext (name, "-,:", save, &delim);
// 	 if (save[0])
// 	 {
// 		 if (!FcPatternObjectAddString (pat, FC_FAMILY_OBJECT, save))
// 		 goto bail2;
// 	 }
// 	 if (delim != ',')
// 		 break;
// 	 }
// 	 if (delim == '-')
// 	 {
// 	 for (;;)
// 	 {
// 		 name = FcNameFindNext (name, "-,:", save, &delim);
// 		 d = strtod ((char *) save, (char **) &e);
// 		 if (e != save)
// 		 {
// 		 if (!FcPatternObjectAddDouble (pat, FC_SIZE_OBJECT, d))
// 			 goto bail2;
// 		 }
// 		 if (delim != ',')
// 		 break;
// 	 }
// 	 }
// 	 while (delim == ':')
// 	 {
// 	 name = FcNameFindNext (name, "=_:", save, &delim);
// 	 if (save[0])
// 	 {
// 		 if (delim == '=' || delim == '_')
// 		 {
// 		 t = FcNameGetObjectType ((char *) save);
// 		 for (;;)
// 		 {
// 			 name = FcNameFindNext (name, ":,", save, &delim);
// 			 if (t)
// 			 {
// 			 v = FcNameConvert (t.type, t.object, save);
// 			 if (!FcPatternAdd (pat, t.object, v, true))
// 			 {
// 				 FcValueDestroy (v);
// 				 goto bail2;
// 			 }
// 			 FcValueDestroy (v);
// 			 }
// 			 if (delim != ',')
// 			 break;
// 		 }
// 		 }
// 		 else
// 		 {
// 		 if ((c = FcNameGetConstant (save)))
// 		 {
// 			 t = FcNameGetObjectType ((char *) c.object);
// 			 if (t == NULL)
// 			 goto bail2;
// 			 switch ((int) t.type) {
// 			 case FcTypeInteger:
// 			 case FcTypeDouble:
// 			 if (!FcPatternAddInteger (pat, c.object, c.value))
// 				 goto bail2;
// 			 break;
// 			 case FcTypeBool:
// 			 if (!FcPatternAddBool (pat, c.object, c.value))
// 				 goto bail2;
// 			 break;
// 			 case FcTypeRange:
// 			 if (!FcPatternAddInteger (pat, c.object, c.value))
// 				 goto bail2;
// 			 break;
// 			 default:
// 			 break;
// 			 }
// 		 }
// 		 }
// 	 }
// 	 }

// 	 free (save);
// 	 return pat;

//  bail2:
// 	 FcPatternDestroy (pat);
//  bail1:
// 	 free (save);
//  bail0:
// 	 return 0;
//  }
//  static FcBool
//  FcNameUnparseString (FcStrBuf	    *buf,
// 			  const FcChar8  *string,
// 			  const FcChar8  *escape)
//  {
// 	 FcChar8 c;
// 	 while ((c = *string++))
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
// 	 while (v)
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
