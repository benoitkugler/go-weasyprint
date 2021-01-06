package fontconfig

import (
	"fmt"
	"math"
	"strings"
)

// ported from fontconfig/src/fcmatch.c Copyright © 2000 Keith Packard

type FcMatchKind uint8

const (
	FcMatchPattern FcMatchKind = iota
	FcMatchFont
	FcMatchScan
	FcMatchKindEnd
	FcMatchKindBegin = FcMatchPattern
)

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

// returns 0 for empty strings
func FcToLower(s string) byte {
	if s == "" {
		return 0
	}
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

func FcCompareLang(val1, val2 interface{}) (interface{}, float64) {
	var result FcLangResult
	switch v1 := val1.(type) {
	case *FcLangSet:
		switch v2 := val2.(type) {
		case *FcLangSet:
			result = FcLangSetCompare(v1, v2)
		case string:
			result = v1.hasLang(v2)
		default:
			return nil, -1.0
		}
	case string:
		switch v2 := val2.(type) {
		case *FcLangSet:
			result = v2.hasLang(v1)
		case string:
			result = FcLangCompare(v1, v2)
		default:
			return nil, -1.0
		}
		break
	default:
		return nil, -1.0
	}
	bestValue := val2
	switch result {
	case FcLangEqual:
		return bestValue, 0
	case FcLangDifferentCountry:
		return bestValue, 1
	default:
		return bestValue, 2
	}
}

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

func FcCompareCharSet(v1, v2 interface{}) (interface{}, float64) {
	bestValue := v2
	return bestValue, float64(FcCharSetSubtractCount(v1.(FcCharSet), v2.(FcCharSet)))
}

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
var fcMatchers = [...]FcMatcher{
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

func (object FcObject) toMatcher(includeLang bool) *FcMatcher {
	if includeLang {
		switch object {
		case FC_FAMILYLANG, FC_STYLELANG, FC_FULLNAMELANG:
			object = FC_LANG
		}
	}
	if int(object) >= len(fcMatchers) ||
		fcMatchers[object].compare == nil ||
		fcMatchers[object].strong == -1 ||
		fcMatchers[object].weak == -1 {
		return nil
	}

	return &fcMatchers[object]
}

func compareValueList(object FcObject, match *FcMatcher,
	v1orig FcValueList, /* pattern */
	v2orig FcValueList, /* target */
	value []float64) (interface{}, FcResult, int, bool) {

	if match == nil {
		return v2orig[0].value, 0, 0, true
	}
	var (
		result    FcResult
		bestValue interface{}
		pos       int
	)
	weak := match.weak
	strong := match.strong

	best := 1e99
	bestStrong := 1e99
	bestWeak := 1e99
	for j, v1 := range v1orig {
		for k, v2 := range v2orig {
			matchValue, v := match.compare(v1.value, v2.value)
			if v < 0 {
				result = FcResultTypeMismatch
				return nil, result, 0, false
			}
			v = v*1000 + float64(j)
			if v < best {
				bestValue = matchValue
				best = v
				pos = k
			}
			if weak == strong {
				// found the best possible match
				if best < 1000 {
					goto done
				}
			} else if v1.binding == FcValueBindingStrong {
				if v < bestStrong {
					bestStrong = v
				}
			} else {
				if v < bestWeak {
					bestWeak = v
				}
			}
		}
	}
done:

	if debugMode {
		fmt.Printf(" %s: %g ", object, best)
		fmt.Println(v1orig)
		fmt.Print(", ")
		fmt.Println(v2orig)
		fmt.Println()
	}

	if value != nil {
		if weak == strong {
			value[strong] += best
		} else {
			value[weak] += bestWeak
			value[strong] += bestStrong
		}
	}
	return bestValue, result, pos, true
}

type FcCompareData = FcHashTable

func (pat *FcPattern) newCompareData() FcCompareData {
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

// compare returns a value indicating the distance between the two lists of values
func (data FcCompareData) compare(pat, fnt *FcPattern, value []float64) (bool, FcResult) {
	for i := range value {
		value[i] = 0.0
	}

	var result FcResult
	for i1, elt_i1 := range pat.elts {
		elt_i2, ok := fnt.elts[i1]
		if !ok {
			continue
		}

		if i1 == FC_FAMILY && data != nil {
			data.FcCompareFamilies(elt_i2, value)
		} else {
			match := i1.toMatcher(false)
			_, result, _, ok = compareValueList(i1, match, elt_i1, elt_i2, value)
			if !ok {
				return false, result
			}
		}
	}
	return true, result
}

// FcFontRenderPrepare creates a new pattern consisting of elements of `font` not appearing
// in `pat`, elements of `pat` not appearing in `font` and the best matching
// value from `pat` for elements appearing in both.  The result is passed to
// FcConfigSubstituteWithPat with `kind` FcMatchFont and then returned.
func (config *FcConfig) FcFontRenderPrepare(pat, font *FcPattern) *FcPattern {
	//  FcPattern	    *new;
	//  int		    i;
	//  FcPatternElt    *fe, *pe;
	//  FcBool	    variable = false;
	var (
		variations strings.Builder
		v          interface{}
	)

	variable, _ := font.FcPatternObjectGetBool(FC_VARIABLE, 0)

	new := FcPattern{elts: make(map[FcObject]FcValueList)}

	for _, obj := range font.sortedKeys() {
		fe := font.elts[obj]
		if obj == FC_FAMILYLANG || obj == FC_STYLELANG || obj == FC_FULLNAMELANG {
			// ignore those objects. we need to deal with them another way
			continue
		}
		if obj == FC_FAMILY || obj == FC_STYLE || obj == FC_FULLNAME {
			// using the fact that FC_FAMILY + 1 == FC_FAMILYLANG,
			// FC_STYLE + 1 == FC_STYLELANG,  FC_FULLNAME + 1 == FC_FULLNAMELANG
			lObject := obj + 1
			fel, pel := font.elts[lObject], pat.elts[lObject]

			if fel != nil && pel != nil {
				// The font has name languages, and pattern asks for specific language(s).
				// Match on language and and prefer that result.
				// Note:  Currently the code only give priority to first matching language.
				var (
					n  int
					ok bool
				)
				match := lObject.toMatcher(true)
				_, _, n, ok = compareValueList(lObject, match, pel, fel, nil)
				if !ok {
					return nil
				}

				var ln, ll FcValueList
				//  j = 0, l1 = FcPatternEltValues (fe), l2 = FcPatternEltValues (fel);
				// 	  l1 != nil || l2 != nil;
				// 	  j++, l1 = l1 ? FcValueListNext (l1) : nil, l2 = l2 ? FcValueListNext (l2) : nil)
				for j := 0; j < len(fe) || j < len(fel); j++ {
					if j == n {
						if j < len(fe) {
							ln = ln.prepend(valueElt{value: fe[j].value, binding: FcValueBindingStrong})
						}
						if j < len(fel) {
							ll = ll.prepend(valueElt{value: fel[j].value, binding: FcValueBindingStrong})
						}
					} else {
						if j < len(fe) {
							ln = append(ln, valueElt{value: fe[j].value, binding: FcValueBindingStrong})
						}
						if j < len(fel) {
							ll = append(ll, valueElt{value: fel[j].value, binding: FcValueBindingStrong})
						}
					}
				}
				new.AddList(obj, ln, false)
				new.AddList(lObject, ll, false)

				continue
			} else if fel != nil {
				//  Pattern doesn't ask for specific language.  Copy all for name and lang
				new.AddList(obj, fe.duplicate(), false)
				new.AddList(lObject, fel.duplicate(), false)

				continue
			}
		}

		pe := pat.elts[obj]
		if pe != nil {
			match := obj.toMatcher(false)
			var ok bool
			v, _, _, ok = compareValueList(obj, match, pe, fe, nil)
			if !ok {
				return nil
			}
			new.Add(obj, v, false)

			// Set font-variations settings for standard axes in variable fonts.
			if _, isRange := fe[0].value.(FcRange); variable != 0 && isRange &&
				(obj == FC_WEIGHT || obj == FC_WIDTH || obj == FC_SIZE) {
				//  double num;
				//  FcChar8 temp[128];
				tag := "    "
				num := v.(float64) //  v.type == FcTypeDouble
				if variations.Len() != 0 {
					variations.WriteByte(',')
				}
				switch obj {
				case FC_WEIGHT:
					tag = "wght"
					num = FcWeightToOpenTypeDouble(num)
				case FC_WIDTH:
					tag = "wdth"
				case FC_SIZE:
					tag = "opsz"
				}
				fmt.Fprintf(&variations, "%4s=%g", tag, num)
			}
		} else {
			new.AddList(obj, fe.duplicate(), true)
		}
	}
	for _, obj := range pat.sortedKeys() {
		pe := pat.elts[obj]
		fe := font.elts[obj]
		if fe == nil &&
			obj != FC_FAMILYLANG && obj != FC_STYLELANG && obj != FC_FULLNAMELANG {
			new.AddList(obj, pe.duplicate(), false)
		}
	}

	if variable != 0 && variations.Len() != 0 {
		if vars, res := new.FcPatternObjectGetString(FC_FONT_VARIATIONS, 0); res == FcResultMatch {
			variations.WriteByte(',')
			variations.WriteString(vars)
			new.del(FC_FONT_VARIATIONS)
		}

		new.Add(FC_FONT_VARIATIONS, variations.String(), true)
	}

	config.FcConfigSubstituteWithPat(&new, pat, FcMatchFont)
	return &new
}

func (p *FcPattern) fontSetMatchInternal(sets []FcFontSet) (*FcPattern, FcResult) {
	var (
		score, bestscore [PRI_END]float64
		best             *FcPattern
		result           FcResult
	)

	if debugMode {
		fmt.Println("Match ")
		fmt.Println(p.String())
	}

	data := p.newCompareData()

	for _, s := range sets {
		if s == nil {
			continue
		}
		for f, pat := range s {
			if debugMode {
				fmt.Printf("Font %d %s", f, pat)
			}
			var ok bool
			ok, result = data.compare(p, pat, score[:])
			if !ok {
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
// 	 best = fontSetMatchInternal (sets, nsets, p, result);
// 	 if (best)
// 	 ret = FcFontRenderPrepare (config, p, best);

// 	 FcConfigDestroy (config);

// 	 return ret;
//  }

func (config *FcConfig) FcFontMatch(p *FcPattern) (*FcPattern, FcResult) {
	var sets []FcFontSet
	if config.fonts[FcSetSystem] != nil {
		sets = append(sets, config.fonts[FcSetSystem])
	}
	if config.fonts[FcSetApplication] != nil {
		sets = append(sets, config.fonts[FcSetApplication])
	}

	var ret *FcPattern
	best, result := p.fontSetMatchInternal(sets)
	if best != nil {
		ret = config.FcFontRenderPrepare(p, best)
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

// 	 newCompareData (p, &data);

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
// 		 if (!compare (p, new.pattern, new.score, result, &data))
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
