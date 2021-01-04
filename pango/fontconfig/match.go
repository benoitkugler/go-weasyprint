package fontconfig

import (
	"fmt"
	"math"
	"strings"
)

// ported from fontconfig/src/fcmatch.c Copyright © 2000 Keith Packard

func FcCompareNumber(value1, value2 interface{}) (interface{}, float64) {
	var v1, v2 float64
	switch value := value1.(type) {
	case int:
		v1 = float64(value)
	case float64:
		v1 = value
	default:
		return nil, -1.0
	}
	switch value := value2.(type) {
	case int:
		v2 = float64(value)
	case float64:
		v2 = value
	default:
		return nil, -1.0
	}

	v := v2 - v1
	if v < 0 {
		v = -v
	}
	return value2, v
}

func FcCompareString(v1, v2 interface{}) (interface{}, float64) {
	bestValue := v2
	if strings.EqualFold(v1.(string), v2.(string)) {
		return bestValue, 0
	}
	return bestValue, 1
}

func FcToLower(s string) byte {
	if 0101 <= s[0] && s[0] <= 0132 {
		return s[0] - 0101 + 0141
	}
	return s[0]
}

func FcCompareFamily(v1, v2 interface{}) (interface{}, float64) {
	// rely on the guarantee in FcPatternObjectAddWithBinding that
	// families are always FcTypeString.
	v1_string := v1.(string)
	v2_string := v2.(string)

	bestValue := v2

	if FcToLower(v1_string) != FcToLower(v2_string) &&
		v1_string[0] != ' ' && v2_string[0] != ' ' {
		return bestValue, 1.0
	}

	if ignoreBlanksAndCase(v1_string) == ignoreBlanksAndCase(v2_string) {
		return bestValue, 0
	}
	return bestValue, 1
}

var delimReplacer = strings.NewReplacer(" ", "", "-", "")

func matchIgnoreCaseAndDelims(s1, s2 string) int {
	s1, s2 = delimReplacer.Replace(s1), delimReplacer.Replace(s2)
	s1, s2 = strings.ToLower(s1), strings.ToLower(s2)
	l := len(s1)
	if len(s2) < l {
		l = len(s2)
	}
	i := 0
	for ; i < l; i++ {
		if s1[i] != s2[i] {
			break
		}
	}
	return i
}

func FcComparePostScript(v1, v2 interface{}) (interface{}, float64) {
	v1_string := v1.(string)
	v2_string := v2.(string)

	bestValue := v2

	if FcToLower(v1_string) != FcToLower(v2_string) &&
		v1_string[0] != ' ' && v2_string[0] != ' ' {
		return bestValue, 1.0
	}

	n := matchIgnoreCaseAndDelims(v1_string, v2_string)
	length := len(v1_string)

	return bestValue, float64(length-n) / float64(length)
}

// func FcCompareLang (v1,v2 interface{}) (interface{}, float64)
//  {
// 	 FcLangResult    result;

// 	 switch ((int) v1.type) {
// 	 case FcTypeLangSet:
// 	 switch ((int) v2.type) {
// 	 case FcTypeLangSet:
// 		 result = FcLangSetCompare (FcValueLangSet (v1), FcValueLangSet (v2));
// 		 break;
// 	 case FcTypeString:
// 		 result = FcLangSetHasLang (FcValueLangSet (v1), FcValueString (v2));
// 		 break;
// 	 default:
// 		 return -1.0;
// 	 }
// 	 break;
// 	 case FcTypeString:
// 	 switch ((int) v2.type) {
// 	 case FcTypeLangSet:
// 		 result = FcLangSetHasLang (FcValueLangSet (v2), FcValueString (v1));
// 		 break;
// 	 case FcTypeString:
// 		 result = FcLangCompare (FcValueString (v1), FcValueString (v2));
// 		 break;
// 	 default:
// 		 return -1.0;
// 	 }
// 	 break;
// 	 default:
// 	 return -1.0;
// 	 }
// 	 *bestValue = FcValueCanonicalize (v2);
// 	 switch (result) {
// 	 case FcLangEqual:
// 	 return 0;
// 	 case FcLangDifferentCountry:
// 	 return 1;
// 	 case FcLangDifferentLang:
// 	 default:
// 	 return 2;
// 	 }
//  }

func FcCompareBool(val1, val2 interface{}) (interface{}, float64) {
	v1, ok1 := val1.(FcBool)
	v2, ok2 := val2.(FcBool)
	if !ok1 || !ok2 {
		return nil, -1.0
	}

	var bestValue FcBool
	if v2 != FcDontCare {
		bestValue = v2
	} else {
		bestValue = v1
	}

	if v1 == v2 {
		return bestValue, 0
	}
	return bestValue, 1
}

// func FcCompareCharSet (v1,v2 interface{}) (interface{}, float64)
//  {
// 	 *bestValue = FcValueCanonicalize (v2); /* TODO Improve. */
// 	 return float64 FcCharSetSubtractCount (FcValueCharSet(v1), FcValueCharSet(v2));
//  }

func FcCompareRange(v1, v2 interface{}) (interface{}, float64) {
	var b1, e1, b2, e2, d float64

	switch value1 := v1.(type) {
	case int:
		e1 = float64(value1)
		b1 = e1
	case float64:
		e1 = value1
		b1 = e1
	case FcRange:
		b1 = value1.Begin
		e1 = value1.End
	default:
		return nil, -1
	}
	switch value2 := v2.(type) {
	case int:
		e2 = float64(value2)
		b2 = e2
	case float64:
		e2 = value2
		b2 = e2
	case FcRange:
		b2 = value2.Begin
		e2 = value2.End
	default:
		return nil, -1
	}

	if e1 < b2 {
		d = b2
	} else if e2 < b1 {
		d = e2
	} else {
		d = (math.Max(b1, b2) + math.Min(e1, e2)) * .5
	}

	bestValue := d

	/// if the ranges overlap, it's a match, otherwise return closest distance.
	if e1 < b2 || e2 < b1 {
		return bestValue, math.Min(math.Abs(b2-e1), math.Abs(b1-e2))
	}
	return bestValue, 0.0
}

func FcCompareSize(v1, v2 interface{}) (interface{}, float64) {
	var b1, e1, b2, e2 float64

	switch value1 := v1.(type) {
	case int:
		e1 = float64(value1)
		b1 = e1
	case float64:
		e1 = value1
		b1 = e1
	case FcRange:
		b1 = value1.Begin
		e1 = value1.End
	default:
		return nil, -1
	}
	switch value2 := v2.(type) {
	case int:
		e2 = float64(value2)
		b2 = e2
	case float64:
		e2 = value2
		b2 = e2
	case FcRange:
		b2 = value2.Begin
		e2 = value2.End
	default:
		return nil, -1
	}

	bestValue := (b1 + e1) * .5

	// if the ranges overlap, it's a match, otherwise return closest distance.
	if e1 < b2 || e2 < b1 {
		return bestValue, math.Min(math.Abs(b2-e1), math.Abs(b1-e2))
	}
	if b2 != e2 && b1 == e2 { /* Semi-closed interval. */
		return bestValue, 1e-15
	}
	return bestValue, 0.0
}

func strGlobMatch(glob, st string) bool {
	var str int // index in st
	for i, c := range []byte(glob) {
		switch c {
		case '*':
			// short circuit common case
			if i == len(glob)-1 {
				return true
			}
			// short circuit another common case
			if i < len(glob)-1 && glob[i+1] == '*' {

				l1 := len(st) - str
				l2 := len(glob)
				if l1 < l2 {
					return false
				}
				str += (l1 - l2)
			}
			for str < len(st) {
				if strGlobMatch(glob, st[str:]) {
					return true
				}
				str++
			}
			return false
		case '?':
			if str == len(st) {
				return false
			}
			str++
		default:
			if st[str] != c {
				return false
			}
			str++
		}
	}
	return str == len(st)
}

func FcCompareFilename(v1, v2 interface{}) (interface{}, float64) {
	s1, s2 := v1.(string), v2.(string)
	bestValue := s2
	if s1 == s2 {
		return bestValue, 0.0
	}
	if strings.EqualFold(s1, s2) {
		return bestValue, 1.0
	}
	if strGlobMatch(s1, s2) {
		return bestValue, 2.0
	}
	return bestValue, 3.0
}

// Canonical match priority order
type FcMatcherPriority int8

const (
	PRI_FILE FcMatcherPriority = iota
	PRI_FONTFORMAT
	PRI_VARIABLE
	PRI_SCALABLE
	PRI_COLOR
	PRI_FOUNDRY
	PRI_CHARSET
	PRI_FAMILY_STRONG
	PRI_POSTSCRIPT_NAME_STRONG
	PRI_LANG
	PRI_FAMILY_WEAK
	PRI_POSTSCRIPT_NAME_WEAK
	PRI_SYMBOL
	PRI_SPACING
	PRI_SIZE
	PRI_PIXEL_SIZE
	PRI_STYLE
	PRI_SLANT
	PRI_WEIGHT
	PRI_WIDTH
	PRI_FONT_HAS_HINT
	PRI_DECORATIVE
	PRI_ANTIALIAS
	PRI_RASTERIZER
	PRI_OUTLINE
	PRI_ORDER
	PRI_FONTVERSION
	PRI_END

	PRI_FILE_WEAK            = PRI_FILE
	PRI_FILE_STRONG          = PRI_FILE
	PRI_FONTFORMAT_WEAK      = PRI_FONTFORMAT
	PRI_FONTFORMAT_STRONG    = PRI_FONTFORMAT
	PRI_VARIABLE_WEAK        = PRI_VARIABLE
	PRI_VARIABLE_STRONG      = PRI_VARIABLE
	PRI_SCALABLE_WEAK        = PRI_SCALABLE
	PRI_SCALABLE_STRONG      = PRI_SCALABLE
	PRI_COLOR_WEAK           = PRI_COLOR
	PRI_COLOR_STRONG         = PRI_COLOR
	PRI_FOUNDRY_WEAK         = PRI_FOUNDRY
	PRI_FOUNDRY_STRONG       = PRI_FOUNDRY
	PRI_CHARSET_WEAK         = PRI_CHARSET
	PRI_CHARSET_STRONG       = PRI_CHARSET
	PRI_LANG_WEAK            = PRI_LANG
	PRI_LANG_STRONG          = PRI_LANG
	PRI_SYMBOL_WEAK          = PRI_SYMBOL
	PRI_SYMBOL_STRONG        = PRI_SYMBOL
	PRI_SPACING_WEAK         = PRI_SPACING
	PRI_SPACING_STRONG       = PRI_SPACING
	PRI_SIZE_WEAK            = PRI_SIZE
	PRI_SIZE_STRONG          = PRI_SIZE
	PRI_PIXEL_SIZE_WEAK      = PRI_PIXEL_SIZE
	PRI_PIXEL_SIZE_STRONG    = PRI_PIXEL_SIZE
	PRI_STYLE_WEAK           = PRI_STYLE
	PRI_STYLE_STRONG         = PRI_STYLE
	PRI_SLANT_WEAK           = PRI_SLANT
	PRI_SLANT_STRONG         = PRI_SLANT
	PRI_WEIGHT_WEAK          = PRI_WEIGHT
	PRI_WEIGHT_STRONG        = PRI_WEIGHT
	PRI_WIDTH_WEAK           = PRI_WIDTH
	PRI_WIDTH_STRONG         = PRI_WIDTH
	PRI_FONT_HAS_HINT_WEAK   = PRI_FONT_HAS_HINT
	PRI_FONT_HAS_HINT_STRONG = PRI_FONT_HAS_HINT
	PRI_DECORATIVE_WEAK      = PRI_DECORATIVE
	PRI_DECORATIVE_STRONG    = PRI_DECORATIVE
	PRI_ANTIALIAS_WEAK       = PRI_ANTIALIAS
	PRI_ANTIALIAS_STRONG     = PRI_ANTIALIAS
	PRI_RASTERIZER_WEAK      = PRI_RASTERIZER
	PRI_RASTERIZER_STRONG    = PRI_RASTERIZER
	PRI_OUTLINE_WEAK         = PRI_OUTLINE
	PRI_OUTLINE_STRONG       = PRI_OUTLINE
	PRI_ORDER_WEAK           = PRI_ORDER
	PRI_ORDER_STRONG         = PRI_ORDER
	PRI_FONTVERSION_WEAK     = PRI_FONTVERSION
	PRI_FONTVERSION_STRONG   = PRI_FONTVERSION
)

type FcMatcher struct {
	object       FcObject
	compare      func(v1, v2 interface{}) (interface{}, float64)
	strong, weak FcMatcherPriority
}

// Order is significant, it defines the precedence of
// each value, earlier values are more significant than
// later values
var _FcMatchers = [...]FcMatcher{
	{FC_INVALID, nil, -1, -1},
	{FC_FAMILY, FcCompareFamily, PRI_FAMILY_STRONG, PRI_FAMILY_WEAK},
	{FC_FAMILYLANG, nil, -1, -1},
	{FC_STYLE, FcCompareString, PRI_STYLE_STRONG, PRI_STYLE_WEAK},
	{FC_STYLELANG, nil, -1, -1},
	{FC_FULLNAME, nil, -1, -1},
	{FC_FULLNAMELANG, nil, -1, -1},
	{FC_SLANT, FcCompareNumber, PRI_SLANT_STRONG, PRI_SLANT_WEAK},
	{FC_WEIGHT, FcCompareRange, PRI_WEIGHT_STRONG, PRI_WEIGHT_WEAK},
	{FC_WIDTH, FcCompareRange, PRI_WIDTH_STRONG, PRI_WIDTH_WEAK},
	{FC_SIZE, FcCompareSize, PRI_SIZE_STRONG, PRI_SIZE_WEAK},
	{FC_ASPECT, nil, -1, -1},
	{FC_PIXEL_SIZE, FcCompareNumber, PRI_PIXEL_SIZE_STRONG, PRI_PIXEL_SIZE_WEAK},
	{FC_SPACING, FcCompareNumber, PRI_SPACING_STRONG, PRI_SPACING_WEAK},
	{FC_FOUNDRY, FcCompareString, PRI_FOUNDRY_STRONG, PRI_FOUNDRY_WEAK},
	{FC_ANTIALIAS, FcCompareBool, PRI_ANTIALIAS_STRONG, PRI_ANTIALIAS_WEAK},
	{FC_HINT_STYLE, nil, -1, -1},
	{FC_HINTING, nil, -1, -1},
	{FC_VERTICAL_LAYOUT, nil, -1, -1},
	{FC_AUTOHINT, nil, -1, -1},
	{FC_GLOBAL_ADVANCE, nil, -1, -1},
	{FC_FILE, FcCompareFilename, PRI_FILE_STRONG, PRI_FILE_WEAK},
	{FC_INDEX, nil, -1, -1},
	{FC_RASTERIZER, FcCompareString, PRI_RASTERIZER_STRONG, PRI_RASTERIZER_WEAK},
	{FC_OUTLINE, FcCompareBool, PRI_OUTLINE_STRONG, PRI_OUTLINE_WEAK},
	{FC_SCALABLE, FcCompareBool, PRI_SCALABLE_STRONG, PRI_SCALABLE_WEAK},
	{FC_DPI, nil, -1, -1},
	{FC_RGBA, nil, -1, -1},
	{FC_SCALE, nil, -1, -1},
	{FC_MINSPACE, nil, -1, -1},
	{FC_CHARWIDTH, nil, -1, -1},
	{FC_CHAR_HEIGHT, nil, -1, -1},
	{FC_MATRIX, nil, -1, -1},
	{FC_CHARSET, FcCompareCharSet, PRI_CHARSET_STRONG, PRI_CHARSET_WEAK},
	{FC_LANG, FcCompareLang, PRI_LANG_STRONG, PRI_LANG_WEAK},
	{FC_FONTVERSION, FcCompareNumber, PRI_FONTVERSION_STRONG, PRI_FONTVERSION_WEAK},
	{FC_CAPABILITY, nil, -1, -1},
	{FC_FONTFORMAT, FcCompareString, PRI_FONTFORMAT_STRONG, PRI_FONTFORMAT_WEAK},
	{FC_EMBOLDEN, nil, -1, -1},
	{FC_EMBEDDED_BITMAP, nil, -1, -1},
	{FC_DECORATIVE, FcCompareBool, PRI_DECORATIVE_STRONG, PRI_DECORATIVE_WEAK},
	{FC_LCD_FILTER, nil, -1, -1},
	{FC_NAMELANG, nil, -1, -1},
	{FC_FONT_FEATURES, nil, -1, -1},
	{FC_PRGNAME, nil, -1, -1},
	{FC_HASH, nil, -1, -1},
	{FC_POSTSCRIPT_NAME, FcComparePostScript, PRI_POSTSCRIPT_NAME_STRONG, PRI_POSTSCRIPT_NAME_WEAK},
	{FC_COLOR, FcCompareBool, PRI_COLOR_STRONG, PRI_COLOR_WEAK},
	{FC_SYMBOL, FcCompareBool, PRI_SYMBOL_STRONG, PRI_SYMBOL_WEAK},
	{FC_FONT_VARIATIONS, nil, -1, -1},
	{FC_VARIABLE, FcCompareBool, PRI_VARIABLE_STRONG, PRI_VARIABLE_WEAK},
	{FC_FONT_HAS_HINT, FcCompareBool, PRI_FONT_HAS_HINT_STRONG, PRI_FONT_HAS_HINT_WEAK},
	{FC_ORDER, FcCompareNumber, PRI_ORDER_STRONG, PRI_ORDER_WEAK},
}

//  static const FcMatcher*
//  FcObjectToMatcher (FcObject object,
// 			FcBool   include_lang)
//  {
// 	 if (include_lang)
// 	 {
// 	 switch (object) {
// 	 case FC_FAMILYLANG:
// 	 case FC_STYLELANG:
// 	 case FC_FULLNAMELANG:
// 		 object = FC_LANG;
// 		 break;
// 	 }
// 	 }
// 	 if (object > FC_MAX_BASE ||
// 	 !_FcMatchers[object].compare ||
// 	 _FcMatchers[object].strong == -1 ||
// 	 _FcMatchers[object].weak == -1)
// 	 return nil;

// 	 return _FcMatchers + object;
//  }

//  static FcBool
//  FcCompareValueList (FcObject	     object,
// 			 const FcMatcher *match,
// 			 FcValueListPtr   v1orig,	/* pattern */
// 			 FcValueListPtr   v2orig,	/* target */
// 			 FcValue         *bestValue,
// 			 double          *value,
// 			 int             *n,
// 			 FcResult        *result)
//  {
// 	 FcValueListPtr  v1, v2;
// 	 double    	    v, best, bestStrong, bestWeak;
// 	 int		    j, k, pos = 0;
// 	 int weak, strong;

// 	 if (!match)
// 	 {
// 	 if (bestValue)
// 		 *bestValue = FcValueCanonicalize(&v2orig.value);
// 	 if (n)
// 		 *n = 0;
// 	 return true;
// 	 }

// 	 weak    = match.weak;
// 	 strong  = match.strong;

// 	 best = 1e99;
// 	 bestStrong = 1e99;
// 	 bestWeak = 1e99;
// 	 for (v1 = v1orig, j = 0; v1; v1 = FcValueListNext(v1), j++)
// 	 {
// 	 for (v2 = v2orig, k = 0; v2; v2 = FcValueListNext(v2), k++)
// 	 {
// 		 FcValue matchValue;
// 		 v = (match.compare) (&v1.value, &v2.value, &matchValue);
// 		 if (v < 0)
// 		 {
// 		 *result = FcResultTypeMismatch;
// 		 return false;
// 		 }
// 		 v = v * 1000 + j;
// 		 if (v < best)
// 		 {
// 		 if (bestValue)
// 			 *bestValue = matchValue;
// 		 best = v;
// 		 pos = k;
// 		 }
// 			 if (weak == strong)
// 			 {
// 				 /* found the best possible match */
// 				 if (best < 1000)
// 					 goto done;
// 			 }
// 			 else if (v1.binding == FcValueBindingStrong)
// 		 {
// 		 if (v < bestStrong)
// 			 bestStrong = v;
// 		 }
// 		 else
// 		 {
// 		 if (v < bestWeak)
// 			 bestWeak = v;
// 		 }
// 	 }
// 	 }
//  done:
// 	 if (FcDebug () & FC_DBG_MATCHV)
// 	 {
// 	 printf (" %s: %g ", FcObjectName (object), best);
// 	 FcValueListPrint (v1orig);
// 	 printf (", ");
// 	 FcValueListPrint (v2orig);
// 	 printf ("\n");
// 	 }
// 	 if (value)
// 	 {
// 	 if (weak == strong)
// 		 value[strong] += best;
// 	 else
// 	 {
// 		 value[weak] += bestWeak;
// 		 value[strong] += bestStrong;
// 	 }
// 	 }
// 	 if (n)
// 	 *n = pos;

// 	 return true;
//  }

type FcCompareData = FcHashTable

func (pat *FcPattern) FcCompareDataInit() FcCompareData {
	table := make(FcHashTable)

	elt := pat.elts[FC_FAMILY]
	for i, l := range elt {
		key := string(l.hash()) // l must have type string, but we are cautious
		e, ok := table.lookup(key)
		if !ok {
			e = new(FamilyEntry)
			e.strong_value = 1e99
			e.weak_value = 1e99
			table.add(key, e)
		}
		if l.binding == FcValueBindingWeak {
			if i := float64(i); i < e.weak_value {
				e.weak_value = i
			}
		} else {
			if i := float64(i); i < e.strong_value {
				e.strong_value = i
			}
		}
	}

	return table
}

func (table FcHashTable) FcCompareFamilies(v2orig FcValueList, value []float64) {
	strong_value := 1e99
	weak_value := 1e99

	for _, v2 := range v2orig {
		key := string(v2.hash()) // should be string, but we are cautious
		e, ok := table.lookup(key)
		if ok {
			if e.strong_value < strong_value {
				strong_value = e.strong_value
			}
			if e.weak_value < weak_value {
				weak_value = e.weak_value
			}
		}
	}

	value[PRI_FAMILY_STRONG] = strong_value
	value[PRI_FAMILY_WEAK] = weak_value
}

// FcCompare returns a value indicating the distance between the two lists of values
func (data FcCompareData) FcCompare(pat, fnt *FcPattern, value []float64) (bool, FcResult) {
	for i := range value {
		value[i] = 0.0
	}

	for i1, elt_i1 := range pat.elts {
		elt_i2, ok := fnt.elts[i1]
		if !ok {
			continue
		}

		if i1 == FC_FAMILY && data != nil {
			data.FcCompareFamilies(elt_i2, value)
		} else {
			match := FcObjectToMatcher(elt_i1.object, false)
			if !FcCompareValueList(elt_i1.object, match,
				FcPatternEltValues(elt_i1),
				FcPatternEltValues(elt_i2),
				nil, value, nil, result) {
				return false
			}
		}
	}
	return true
}

//  FcPattern *
//  FcFontRenderPrepare (FcConfig	    *config,
// 			  FcPattern	    *pat,
// 			  FcPattern	    *font)
//  {
// 	 FcPattern	    *new;
// 	 int		    i;
// 	 FcPatternElt    *fe, *pe;
// 	 FcValue	    v;
// 	 FcResult	    result;
// 	 FcBool	    variable = false;
// 	 FcStrBuf        variations;

// 	 assert (pat != nil);
// 	 assert (font != nil);

// 	 FcPatternObjectGetBool (font, FC_VARIABLE, 0, &variable);
// 	 assert (variable != FcDontCare);
// 	 if (variable)
// 	 FcStrBufInit (&variations, nil, 0);

// 	 new = FcPatternCreate ();
// 	 if (!new)
// 	 return nil;
// 	 for (i = 0; i < font.num; i++)
// 	 {
// 	 fe = &FcPatternElts(font)[i];
// 	 if (fe.object == FC_FAMILYLANG ||
// 		 fe.object == FC_STYLELANG ||
// 		 fe.object == FC_FULLNAMELANG)
// 	 {
// 		 /* ignore those objects. we need to deal with them
// 		  * another way */
// 		 continue;
// 	 }
// 	 if (fe.object == FC_FAMILY ||
// 		 fe.object == FC_STYLE ||
// 		 fe.object == FC_FULLNAME)
// 	 {
// 		 FcPatternElt    *fel, *pel;

// 		 FC_ASSERT_STATIC ((FC_FAMILY + 1) == FC_FAMILYLANG);
// 		 FC_ASSERT_STATIC ((FC_STYLE + 1) == FC_STYLELANG);
// 		 FC_ASSERT_STATIC ((FC_FULLNAME + 1) == FC_FULLNAMELANG);

// 		 fel = FcPatternObjectFindElt (font, fe.object + 1);
// 		 pel = FcPatternObjectFindElt (pat, fe.object + 1);

// 		 if (fel && pel)
// 		 {
// 		 /* The font has name languages, and pattern asks for specific language(s).
// 		  * Match on language and and prefer that result.
// 		  * Note:  Currently the code only give priority to first matching language.
// 		  */
// 		 int n = 1, j;
// 		 FcValueListPtr l1, l2, ln = nil, ll = nil;
// 		 const FcMatcher *match = FcObjectToMatcher (pel.object, true);

// 		 if (!FcCompareValueList (pel.object, match,
// 					  FcPatternEltValues (pel),
// 					  FcPatternEltValues (fel), nil, nil, &n, &result))
// 		 {
// 			 FcPatternDestroy (new);
// 			 return nil;
// 		 }

// 		 for (j = 0, l1 = FcPatternEltValues (fe), l2 = FcPatternEltValues (fel);
// 			  l1 != nil || l2 != nil;
// 			  j++, l1 = l1 ? FcValueListNext (l1) : nil, l2 = l2 ? FcValueListNext (l2) : nil)
// 		 {
// 			 if (j == n)
// 			 {
// 			 if (l1)
// 				 ln = FcValueListPrepend (ln,
// 							  FcValueCanonicalize (&l1.value),
// 							  FcValueBindingStrong);
// 			 if (l2)
// 				 ll = FcValueListPrepend (ll,
// 							  FcValueCanonicalize (&l2.value),
// 							  FcValueBindingStrong);
// 			 }
// 			 else
// 			 {
// 			 if (l1)
// 				 ln = FcValueListAppend (ln,
// 							 FcValueCanonicalize (&l1.value),
// 							 FcValueBindingStrong);
// 			 if (l2)
// 				 ll = FcValueListAppend (ll,
// 							 FcValueCanonicalize (&l2.value),
// 							 FcValueBindingStrong);
// 			 }
// 		 }
// 		 FcPatternObjectListAdd (new, fe.object, ln, false);
// 		 FcPatternObjectListAdd (new, fel.object, ll, false);

// 		 continue;
// 		 }
// 		 else if (fel)
// 		 {
// 		 /* Pattern doesn't ask for specific language.  Copy all for name and
// 		  * lang. */
// 		 FcValueListPtr l1, l2;

// 		 l1 = FcValueListDuplicate (FcPatternEltValues (fe));
// 		 l2 = FcValueListDuplicate (FcPatternEltValues (fel));
// 		 FcPatternObjectListAdd (new, fe.object, l1, false);
// 		 FcPatternObjectListAdd (new, fel.object, l2, false);

// 		 continue;
// 		 }
// 	 }

// 	 pe = FcPatternObjectFindElt (pat, fe.object);
// 	 if (pe)
// 	 {
// 		 const FcMatcher *match = FcObjectToMatcher (pe.object, false);
// 		 if (!FcCompareValueList (pe.object, match,
// 					  FcPatternEltValues(pe),
// 					  FcPatternEltValues(fe), &v, nil, nil, &result))
// 		 {
// 		 FcPatternDestroy (new);
// 		 return nil;
// 		 }
// 		 FcPatternObjectAdd (new, fe.object, v, false);

// 		 /* Set font-variations settings for standard axes in variable fonts. */
// 		 if (variable &&
// 		 FcPatternEltValues(fe).value.type == FcTypeRange &&
// 		 (fe.object == FC_WEIGHT ||
// 		  fe.object == FC_WIDTH ||
// 		  fe.object == FC_SIZE))
// 		 {
// 		 double num;
// 		 FcChar8 temp[128];
// 		 const char *tag = "    ";
// 		 assert (v.type == FcTypeDouble);
// 		 num = v.u.d;
// 		 if (variations.len)
// 			 FcStrBufChar (&variations, ',');
// 		 switch (fe.object)
// 		 {
// 			 case FC_WEIGHT:
// 			 tag = "wght";
// 			 num = FcWeightToOpenType (num);
// 			 break;

// 			 case FC_WIDTH:
// 			 tag = "wdth";
// 			 break;

// 			 case FC_SIZE:
// 			 tag = "opsz";
// 			 break;
// 		 }
// 		 sprintf ((char *) temp, "%4s=%g", tag, num);
// 		 FcStrBufString (&variations, temp);
// 		 }
// 	 }
// 	 else
// 	 {
// 		 FcPatternObjectListAdd (new, fe.object,
// 					 FcValueListDuplicate (FcPatternEltValues (fe)),
// 					 true);
// 	 }
// 	 }
// 	 for (i = 0; i < pat.num; i++)
// 	 {
// 	 pe = &FcPatternElts(pat)[i];
// 	 fe = FcPatternObjectFindElt (font, pe.object);
// 	 if (!fe &&
// 		 pe.object != FC_FAMILYLANG &&
// 		 pe.object != FC_STYLELANG &&
// 		 pe.object != FC_FULLNAMELANG)
// 	 {
// 		 FcPatternObjectListAdd (new, pe.object,
// 					 FcValueListDuplicate (FcPatternEltValues(pe)),
// 					 false);
// 	 }
// 	 }

// 	 if (variable && variations.len)
// 	 {
// 	 FcChar8 *vars = nil;
// 	 if (FcPatternObjectGetString (new, FC_FONT_VARIATIONS, 0, &vars) == FcResultMatch)
// 	 {
// 		 FcStrBufChar (&variations, ',');
// 		 FcStrBufString (&variations, vars);
// 		 FcPatternObjectDel (new, FC_FONT_VARIATIONS);
// 	 }

// 	 FcPatternObjectAddString (new, FC_FONT_VARIATIONS, FcStrBufDoneStatic (&variations));
// 	 FcStrBufDestroy (&variations);
// 	 }

// 	 FcConfigSubstituteWithPat (config, new, pat, FcMatchFont);
// 	 return new;
//  }

func (p *FcPattern) FcFontSetMatchInternal(sets []FcFontSet) (*FcPattern, FcResult) {
	var (
		score, bestscore [PRI_END]float64
		best             *FcPattern
	)

	if debugMode {
		fmt.Println("Match ")
		fmt.Println(p.String())
	}

	data := p.FcCompareDataInit()

	for _, s := range sets {
		if s == nil {
			continue
		}
		for f, pat := range s {
			if debugMode {
				fmt.Printf("Font %d %s", f, pat)
			}
			if !FcCompare(p, pat, score, result, &data) {
				return nil, result
			}
			if debugMode {
				fmt.Printf("Score %v\n", score)
			}
			for i, bs := range bestscore {
				if best != nil && bs < score[i] {
					break
				}
				if best == nil || score[i] < bs {
					for j, s := range score {
						bestscore[j] = s
					}
					best = pat
					break
				}
			}
		}
	}

	if debugMode {
		fmt.Printf("Best score %v\n", bestscore)
	}

	result := FcResultNoMatch
	if best != nil {
		result = FcResultMatch
	}

	return best, result
}

//  FcPattern *
//  FcFontSetMatch (FcConfig    *config,
// 		 FcFontSet   **sets,
// 		 int	    nsets,
// 		 FcPattern   *p,
// 		 FcResult    *result)
//  {
// 	 FcPattern	    *best, *ret = nil;

// 	 assert (sets != nil);
// 	 assert (p != nil);
// 	 assert (result != nil);

// 	 *result = FcResultNoMatch;

// 	 config = FcConfigReference (config);
// 	 if (!config)
// 		 return nil;
// 	 best = FcFontSetMatchInternal (sets, nsets, p, result);
// 	 if (best)
// 	 ret = FcFontRenderPrepare (config, p, best);

// 	 FcConfigDestroy (config);

// 	 return ret;
//  }

func (config *FcConfig) FcFontMatch(p *FcPattern) (*FcPattern, FcResult) {
	//  int		nsets;
	//  FcPattern   *best, *ret = nil;

	result := FcResultNoMatch

	var (
		sets  [2]FcFontSet
		nsets = 0
	)
	if config.fonts[FcSetSystem] != nil {
		sets[nsets] = config.fonts[FcSetSystem]
		nsets++
	}
	if config.fonts[FcSetApplication] != nil {
		sets[nsets] = config.fonts[FcSetApplication]
		nsets++
	}

	var ret *FcPattern
	best := FcFontSetMatchInternal(sets, nsets, p, result)
	if best {
		ret = FcFontRenderPrepare(config, p, best)
	}

	return ret, result
}

//  typedef struct _FcSortNode {
// 	 pattern *FcPattern;
// 	 double	score[PRI_END];
//  } FcSortNode;

//  static int
//  FcSortCompare (const void *aa, const void *ab)
//  {
// 	 FcSortNode  *a = *(FcSortNode **) aa;
// 	 FcSortNode  *b = *(FcSortNode **) ab;
// 	 double	*as = &a.score[0];
// 	 double	*bs = &b.score[0];
// 	 double	ad = 0, bd = 0;
// 	 int         i;

// 	 i = PRI_END;
// 	 for (i-- && (ad = *as++) == (bd = *bs++))
// 	 ;
// 	 return ad < bd ? -1 : ad > bd ? 1 : 0;
//  }

//  static FcBool
//  FcSortWalk (FcSortNode **n, int nnode, FcFontSet *fs, FcCharSet **csp, FcBool trim)
//  {
// 	 FcBool ret = false;
// 	 FcCharSet *cs;
// 	 int i;

// 	 cs = 0;
// 	 if (trim || csp)
// 	 {
// 	 cs = FcCharSetCreate ();
// 	 if (cs == nil)
// 		 goto bail;
// 	 }

// 	 for (i = 0; i < nnode; i++)
// 	 {
// 	 FcSortNode	*node = *n++;
// 	 FcBool		adds_chars = false;

// 	 /*
// 	  * Only fetch node charset if we'd need it
// 	  */
// 	 if (cs)
// 	 {
// 		 FcCharSet	*ncs;

// 		 if (FcPatternGetCharSet (node.pattern, FC_CHARSET, 0, &ncs) !=
// 		 FcResultMatch)
// 			 continue;

// 		 if (!FcCharSetMerge (cs, ncs, &adds_chars))
// 		 goto bail;
// 	 }

// 	 /*
// 	  * If this font isn't a subset of the previous fonts,
// 	  * add it to the list
// 	  */
// 	 if (!i || !trim || adds_chars)
// 	 {
// 		 FcPatternReference (node.pattern);
// 		 if (FcDebug () & FC_DBG_MATCHV)
// 		 {
// 		 printf ("Add ");
// 		 FcPatternPrint (node.pattern);
// 		 }
// 		 if (!FcFontSetAdd (fs, node.pattern))
// 		 {
// 		 FcPatternDestroy (node.pattern);
// 		 goto bail;
// 		 }
// 	 }
// 	 }
// 	 if (csp)
// 	 {
// 	 *csp = cs;
// 	 cs = 0;
// 	 }

// 	 ret = true;

//  bail:
// 	 if (cs)
// 	 FcCharSetDestroy (cs);

// 	 return ret;
//  }

//  void
//  FcFontSetSortDestroy (FcFontSet *fs)
//  {
// 	 FcFontSetDestroy (fs);
//  }

//  FcFontSet *
//  FcFontSetSort (FcConfig	    *config FC_UNUSED,
// 			FcFontSet    **sets,
// 			int	    nsets,
// 			FcPattern    *p,
// 			FcBool	    trim,
// 			FcCharSet    **csp,
// 			FcResult	    *result)
//  {
// 	 FcFontSet	    *ret;
// 	 FcFontSet	    *s;
// 	 FcSortNode	    *nodes;
// 	 FcSortNode	    **nodeps, **nodep;
// 	 int		    nnodes;
// 	 FcSortNode	    *new;
// 	 int		    set;
// 	 int		    f;
// 	 int		    i;
// 	 int		    nPatternLang;
// 	 FcBool    	    *patternLangSat;
// 	 FcValue	    patternLang;
// 	 FcCompareData   data;

// 	 assert (sets != nil);
// 	 assert (p != nil);
// 	 assert (result != nil);

// 	 /* There are some implementation that relying on the result of
// 	  * "result" to check if the return value of FcFontSetSort
// 	  * is valid or not.
// 	  * So we should initialize it to the conservative way since
// 	  * this function doesn't return nil anymore.
// 	  */
// 	 if (result)
// 	 *result = FcResultNoMatch;

// 	 if (FcDebug () & FC_DBG_MATCH)
// 	 {
// 	 printf ("Sort ");
// 	 FcPatternPrint (p);
// 	 }
// 	 nnodes = 0;
// 	 for (set = 0; set < nsets; set++)
// 	 {
// 	 s = sets[set];
// 	 if (!s)
// 		 continue;
// 	 nnodes += s.nfont;
// 	 }
// 	 if (!nnodes)
// 	 return FcFontSetCreate ();

// 	 for (nPatternLang = 0;
// 	  FcPatternGet (p, FC_LANG, nPatternLang, &patternLang) == FcResultMatch;
// 	  nPatternLang++)
// 	 ;

// 	 /* freed below */
// 	 nodes = malloc (nnodes * sizeof (FcSortNode) +
// 			 nnodes * sizeof (FcSortNode *) +
// 			 nPatternLang * sizeof (FcBool));
// 	 if (!nodes)
// 	 goto bail0;
// 	 nodeps = (FcSortNode **) (nodes + nnodes);
// 	 patternLangSat = (FcBool *) (nodeps + nnodes);

// 	 FcCompareDataInit (p, &data);

// 	 new = nodes;
// 	 nodep = nodeps;
// 	 for (set = 0; set < nsets; set++)
// 	 {
// 	 s = sets[set];
// 	 if (!s)
// 		 continue;
// 	 for (f = 0; f < s.nfont; f++)
// 	 {
// 		 if (FcDebug () & FC_DBG_MATCHV)
// 		 {
// 		 printf ("Font %d ", f);
// 		 FcPatternPrint (s.fonts[f]);
// 		 }
// 		 new.pattern = s.fonts[f];
// 		 if (!FcCompare (p, new.pattern, new.score, result, &data))
// 		 goto bail1;
// 		 if (FcDebug () & FC_DBG_MATCHV)
// 		 {
// 		 printf ("Score");
// 		 for (i = 0; i < PRI_END; i++)
// 		 {
// 			 printf (" %g", new.score[i]);
// 		 }
// 		 printf ("\n");
// 		 }
// 		 *nodep = new;
// 		 new++;
// 		 nodep++;
// 	 }
// 	 }

// 	 FcCompareDataClear (&data);

// 	 nnodes = new - nodes;

// 	 qsort (nodeps, nnodes, sizeof (FcSortNode *),
// 		FcSortCompare);

// 	 for (i = 0; i < nPatternLang; i++)
// 	 patternLangSat[i] = false;

// 	 for (f = 0; f < nnodes; f++)
// 	 {
// 	 FcBool	satisfies = false;
// 	 /*
// 	  * If this node matches any language, go check
// 	  * which ones and satisfy those entries
// 	  */
// 	 if (nodeps[f].score[PRI_LANG] < 2000)
// 	 {
// 		 for (i = 0; i < nPatternLang; i++)
// 		 {
// 		 FcValue	    nodeLang;

// 		 if (!patternLangSat[i] &&
// 			 FcPatternGet (p, FC_LANG, i, &patternLang) == FcResultMatch &&
// 			 FcPatternGet (nodeps[f].pattern, FC_LANG, 0, &nodeLang) == FcResultMatch)
// 		 {
// 			 FcValue matchValue;
// 			 double  compare = FcCompareLang (&patternLang, &nodeLang, &matchValue);
// 			 if (compare >= 0 && compare < 2)
// 			 {
// 			 if (FcDebug () & FC_DBG_MATCHV)
// 			 {
// 				 FcChar8 *family;
// 				 FcChar8 *style;

// 				 if (FcPatternGetString (nodeps[f].pattern, FC_FAMILY, 0, &family) == FcResultMatch &&
// 				 FcPatternGetString (nodeps[f].pattern, FC_STYLE, 0, &style) == FcResultMatch)
// 				 printf ("Font %s:%s matches language %d\n", family, style, i);
// 			 }
// 			 patternLangSat[i] = true;
// 			 satisfies = true;
// 			 break;
// 			 }
// 		 }
// 		 }
// 	 }
// 	 if (!satisfies)
// 	 {
// 		 nodeps[f].score[PRI_LANG] = 10000.0;
// 	 }
// 	 }

// 	 /*
// 	  * Re-sort once the language issues have been settled
// 	  */
// 	 qsort (nodeps, nnodes, sizeof (FcSortNode *),
// 		FcSortCompare);

// 	 ret = FcFontSetCreate ();
// 	 if (!ret)
// 	 goto bail1;

// 	 if (!FcSortWalk (nodeps, nnodes, ret, csp, trim))
// 	 goto bail2;

// 	 free (nodes);

// 	 if (FcDebug() & FC_DBG_MATCH)
// 	 {
// 	 printf ("First font ");
// 	 FcPatternPrint (ret.fonts[0]);
// 	 }
// 	 if (ret.nfont > 0)
// 	 *result = FcResultMatch;

// 	 return ret;

//  bail2:
// 	 FcFontSetDestroy (ret);
//  bail1:
// 	 free (nodes);
//  bail0:
// 	 return 0;
//  }

//  FcFontSet *
//  FcFontSort (config *FcConfig,
// 		 p *FcPattern,
// 		 FcBool	trim,
// 		 FcCharSet	**csp,
// 		 result *FcResult)
//  {
// 	 FcFontSet	*sets[2], *ret;
// 	 int		nsets;

// 	 assert (p != nil);
// 	 assert (result != nil);

// 	 *result = FcResultNoMatch;

// 	 config = FcConfigReference (config);
// 	 if (!config)
// 	 return nil;
// 	 nsets = 0;
// 	 if (config.fonts[FcSetSystem])
// 	 sets[nsets++] = config.fonts[FcSetSystem];
// 	 if (config.fonts[FcSetApplication])
// 	 sets[nsets++] = config.fonts[FcSetApplication];
// 	 ret = FcFontSetSort (config, sets, nsets, p, trim, csp, result);
// 	 FcConfigDestroy (config);

// 	 return ret;
//  }
