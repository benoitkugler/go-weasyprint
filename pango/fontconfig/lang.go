package fontconfig

import (
	"fmt"
	"log"
	"strings"
)

// fontconfig/src/fclang.c Copyright © 2002 Keith Packard

type FcLangResult uint8

const (
	FcLangEqual              FcLangResult = 0
	FcLangDifferentCountry   FcLangResult = 1
	FcLangDifferentTerritory FcLangResult = 1
	FcLangDifferentLang      FcLangResult = 2
)

/* Objects MT-safe for readonly access. */

type FcCharLeaf [8]uint32

type FcLangCharSet struct {
	lang    string
	charset FcCharSet
}

type FcLangCharSetRange struct {
	begin, end int
}

// FcLangSet holds the set of languages supported
// by a font.
// These are computed for a font based on orthographic information built into the
// fontconfig library. Fontconfig has orthographies for all of the ISO 639-1
// languages except for MS, NA, PA, PS, QU, RN, RW, SD, SG, SN, SU and ZA.
type FcLangSet struct {
	extra FcStrSet
	map_  [NUM_LANG_SET_MAP]uint32
}

func (ls *FcLangSet) FcLangSetBitSet(id int) {
	id = int(fcLangCharSetIndices[id])
	bucket := id >> 5
	if bucket >= NUM_LANG_SET_MAP { // shouldn't happen really
		return
	}

	ls.map_[bucket] |= uint32('\u0001') << (id & 0x1f)
}

func (ls FcLangSet) bitGet(id int) bool {
	id = int(fcLangCharSetIndices[id])
	bucket := id >> 5
	if bucket >= NUM_LANG_SET_MAP {
		return false
	}

	return (ls.map_[bucket]>>(id&0x1f))&1 != 0
}

func FcStrSetAddLangs(strs FcStrSet, languages string) bool {
	var ret bool
	ls := strings.Split(languages, ":")
	for _, lang := range ls {
		if lang == "" { // ignore an empty item
			continue
		}
		normalizedLang := FcLangNormalize(lang)
		if normalizedLang != "" {
			strs[normalizedLang] = true
			ret = true
		}
	}

	return ret
}

//  static void
//  FcLangSetBitReset (ls *FcLangSet    *ls,
// 			unsigned int  id)
//  {
//    unsigned int bucket;

//    id = fcLangCharSetIndices[id];
//    bucket = id >> 5;
//    if (bucket >= NUM_LANG_SET_MAP)
// 	 return; /* shouldn't happen really */

//    ls.map_[bucket] &= ~((FcChar32) 1U << (id & 0x1f));
//  }

//  FcLangSet *
//  FcFreeTypeLangSet (const FcCharSet  *charset,
// 			const FcChar8    *exclusiveLang)
//  {
// 	 int		    i, j;
// 	 FcChar32	    missing;
// 	 const FcCharSet *exclusiveCharset = 0;
// 	 FcLangSet	    *ls;

// 	 if (exclusiveLang)
// 	 exclusiveCharset = FcLangGetCharSet (exclusiveLang);
// 	 ls = FcLangSetCreate ();
// 	 if (!ls)
// 	 return 0;
// 	 if (FcDebug() & FC_DBG_LANGSET)
// 	 {
// 	 printf ("font charset");
// 	 FcCharSetPrint (charset);
// 	 printf ("\n");
// 	 }
// 	 for (i = 0; i < NUM_LANG_CHAR_SET; i++)
// 	 {
// 	 if (FcDebug() & FC_DBG_LANGSET)
// 	 {
// 		 printf ("%s charset", fcLangCharSets[i].lang);
// 		 FcCharSetPrint (&fcLangCharSets[i].charset);
// 		 printf ("\n");
// 	 }

// 	 /*
// 	  * Check for Han charsets to make fonts
// 	  * which advertise support for a single language
// 	  * not support other Han languages
// 	  */
// 	 if (exclusiveCharset &&
// 		 FcFreeTypeIsExclusiveLang (fcLangCharSets[i].lang))
// 	 {
// 		 if (fcLangCharSets[i].charset.num != exclusiveCharset.num)
// 		 continue;

// 		 for (j = 0; j < fcLangCharSets[i].charset.num; j++)
// 		 if (FcCharSetLeaf(&fcLangCharSets[i].charset, j) !=
// 			 FcCharSetLeaf(exclusiveCharset, j))
// 			 continue;
// 	 }
// 	 missing = FcCharSetSubtractCount (&fcLangCharSets[i].charset, charset);
// 		 if (FcDebug() & FC_DBG_SCANV)
// 	 {
// 		 if (missing && missing < 10)
// 		 {
// 		 FcCharSet   *missed = FcCharSetSubtract (&fcLangCharSets[i].charset,
// 							  charset);
// 		 FcChar32    ucs4;
// 		 FcChar32    map_[FC_CHARSET_MAP_SIZE];
// 		 FcChar32    next;

// 		 printf ("\n%s(%u) ", fcLangCharSets[i].lang, missing);
// 		 printf ("{");
// 		 for (ucs4 = FcCharSetFirstPage (missed, map_, &next);
// 			  ucs4 != FC_CHARSET_DONE;
// 			  ucs4 = FcCharSetNextPage (missed, map_, &next))
// 		 {
// 			 int	    i, j;
// 			 for (i = 0; i < FC_CHARSET_MAP_SIZE; i++)
// 			 if (map_[i])
// 			 {
// 				 for (j = 0; j < 32; j++)
// 				 if (map_[i] & (1U << j))
// 					 printf (" %04x", ucs4 + i * 32 + j);
// 			 }
// 		 }
// 		 printf (" }\n\t");
// 		 FcCharSetDestroy (missed);
// 		 }
// 		 else
// 		 printf ("%s(%u) ", fcLangCharSets[i].lang, missing);
// 	 }
// 	 if (!missing)
// 		 FcLangSetBitSet (ls, i);
// 	 }

// 	 if (FcDebug() & FC_DBG_SCANV)
// 	 printf ("\n");

// 	 return ls;
//  }

func FcLangNormalize(lang string) string {
	var (
		result string
		orig   = lang
	)

	lang = strings.ToLower(lang)
	switch lang {
	case "c", "c.utf-8, c.utf8", "posix":
		return "en"
	}

	/* from the comments in glibc:
	 *
	 * LOCALE can consist of up to four recognized parts for the XPG syntax:
	 *
	 *            language[_territory[.codeset]][@modifier]
	 *
	 * Beside the first all of them are allowed to be missing.  If the
	 * full specified locale is not found, the less specific one are
	 * looked for.  The various part will be stripped off according to
	 * the following order:
	 *            (1) codeset
	 *            (2) normalized codeset
	 *            (3) territory
	 *            (4) modifier
	 *
	 * So since we don't take care of the codeset part here, what patterns
	 * we need to deal with is:
	 *
	 *   1. language_territory@modifier
	 *   2. language@modifier
	 *   3. language
	 *
	 * then. and maybe no need to try language_territory here.
	 */
	var modifier, territory string

	if modifierI := strings.IndexByte(lang, '@'); modifierI != -1 {
		modifier = lang[modifierI+1:]
		lang = lang[0:modifierI]
	}
	encoding := strings.IndexByte(lang, '.')
	if encoding != -1 {
		lang = lang[0:encoding]
	}
	territoryI := strings.IndexByte(lang, '_')
	if territoryI == -1 {
		territoryI = strings.IndexByte(lang, '-')
	}
	if territoryI != -1 {
		territory = lang[territoryI+1:]
		lang = lang[0:territoryI]
	}
	llen := len(lang)
	tlen := len(territory)
	tm := territory
	if llen < 2 || llen > 3 {
		log.Printf("fontconfig: ignoring %s: not a valid language tag", lang)
		return result
	}
	if tlen != 0 && (tlen < 2 || tlen > 3) && !(territory[0] == 'z' && tlen < 5) {
		log.Printf("fontconfig: ignoring %s: not a valid region tag", lang)
		return result
	}
	if modifier != "" {
		tm += "@" + modifier
	}
	if territory != "" {
		if debugMode {
			fmt.Printf("Checking the existence of %s.orth\n", lang+"_"+tm)
		}
		if FcLangSetIndex(lang+"_"+tm) < 0 {
		} else {
			return lang + "_" + tm
		}
	}
	if modifier != "" {
		if debugMode {
			fmt.Printf("Checking the existence of %s.orth\n", lang+"@"+modifier)
		}
		if FcLangSetIndex(lang+"@"+modifier) < 0 {
		} else {
			return lang + "@" + modifier
		}
	}
	if debugMode {
		fmt.Printf("Checking the existence of %s.orth\n", lang)
	}
	if FcLangSetIndex(lang) < 0 {
		// there seems no languages matched in orth. add the language as is for fallback.
		result = orig
	} else {
		result = lang
	}

	return result
}

func langEnd(c string) bool {
	return c == "" || c[0] == '-'
}

func FcLangCompare(s1, s2 string) FcLangResult {
	result := FcLangDifferentLang

	isUnd := FcToLower(s1) == 'u' && FcToLower(s1[1:]) == 'n' &&
		FcToLower(s1[2:]) == 'd' && langEnd(s1[3:])

	for i := 0; ; i++ {
		c1 := FcToLower(s1[i:])
		c2 := FcToLower(s2[i:])
		if c1 != c2 {
			if !isUnd && langEnd(s1[i:]) && langEnd(s2[i:]) {
				result = FcLangDifferentTerritory
			}
			return result
		} else if c1 == 0 {
			if isUnd {
				return result
			}
			return FcLangEqual
		} else if c1 == '-' {
			if !isUnd {
				result = FcLangDifferentTerritory
			}
		}

		// If we parsed past "und-", then do not consider it undefined anymore,
		// as there's *something* specified.
		if isUnd && i == 3 {
			isUnd = false
		}
	}
}

// Return true when super contains sub.
//
// super contains sub if super and sub have the same
// language and either the same country or one
// is missing the country
func langContains(super, sub string) bool {
	for {
		c1 := FcToLower(super)
		c2 := FcToLower(sub)
		if c1 != c2 {
			// see if super has a country for sub is missing one
			if c1 == '-' && c2 == 0 {
				return true
			}
			// see if sub has a country for super is missing one
			if c1 == 0 && c2 == '-' {
				return true
			}
			return false
		} else if c1 == 0 {
			return true
		}
		super, sub = super[1:], sub[1:]
	}
}

//  const FcCharSet *
//  FcLangGetCharSet (const FcChar8 *lang)
//  {
// 	 int		i;
// 	 int		country = -1;

// 	 for (i = 0; i < NUM_LANG_CHAR_SET; i++)
// 	 {
// 	 switch (FcLangCompare (lang, fcLangCharSets[i].lang)) {
// 	 case FcLangEqual:
// 		 return &fcLangCharSets[i].charset;
// 	 case FcLangDifferentTerritory:
// 		 if (country == -1)
// 		 country = i;
// 	 case FcLangDifferentLang:
// 	 default:
// 		 break;
// 	 }
// 	 }
// 	 if (country == -1)
// 	 return 0;
// 	 return &fcLangCharSets[country].charset;
//  }

//  FcStrSet *
//  FcGetLangs (void)
//  {
// 	 FcStrSet *langs;
// 	 int	i;

// 	 langs = FcStrSetCreate();
// 	 if (!langs)
// 	 return 0;

// 	 for (i = 0; i < NUM_LANG_CHAR_SET; i++)
// 	 FcStrSetAdd (langs, fcLangCharSets[i].lang);

// 	 return langs;
//  }

//  FcLangSet *
//  FcLangSetCreate (void)
//  {
// 	 FcLangSet	*ls;

// 	 ls = malloc (sizeof (FcLangSet));
// 	 if (!ls)
// 	 return 0;
// 	 memset (ls.map_, '\0', sizeof (ls.map_));
// 	 NUM_LANG_SET_MAP = NUM_LANG_SET_MAP;
// 	 ls.extra = 0;
// 	 return ls;
//  }

//  void
//  FcLangSetDestroy (FcLangSet *ls)
//  {
// 	 if (!ls)
// 	 return;

// 	 if (ls.extra)
// 	 FcStrSetDestroy (ls.extra);
// 	 free (ls);
//  }

//  FcLangSet *
//  FcLangSetCopy (const FcLangSet *ls)
//  {
// 	 FcLangSet	*new;

// 	 if (!ls)
// 	 return NULL;

// 	 new = FcLangSetCreate ();
// 	 if (!new)
// 	 goto bail0;
// 	 memset (new.map_, '\0', sizeof (new.map_));
// 	 memcpy (new.map_, ls.map_, FC_MIN (sizeof (new.map_), NUM_LANG_SET_MAP * sizeof (ls.map_[0])));
// 	 if (ls.extra)
// 	 {
// 	 FcStrList	*list;
// 	 FcChar8		*extra;

// 	 new.extra = FcStrSetCreate ();
// 	 if (!new.extra)
// 		 goto bail1;

// 	 list = FcStrListCreate (ls.extra);
// 	 if (!list)
// 		 goto bail1;

// 	 for ((extra = FcStrListNext (list)))
// 		 if (!FcStrSetAdd (new.extra, extra))
// 		 {
// 		 FcStrListDone (list);
// 		 goto bail1;
// 		 }
// 	 FcStrListDone (list);
// 	 }
// 	 return new;
//  bail1:
// 	 FcLangSetDestroy (new);
//  bail0:
// 	 return 0;
//  }

func FcStrCmpIgnoreCase(s1, s2 string) int {
	return strings.Compare(strings.ToLower(s1), strings.ToLower(s2))
}

/* When the language isn't found, the return value r is such that:
 *  1) r < 0
 *  2) -r -1 is the index of the first language in fcLangCharSets that comes
 *     after the 'lang' argument in lexicographic order.
 *
 *  The -1 is necessary to avoid problems with language id 0 (otherwise, we
 *  wouldn't be able to distinguish between “language found, id is 0” and
 *  “language not found, sorts right before the language with id 0”).
 */
func FcLangSetIndex(lang string) int {
	firstChar := FcToLower(lang)
	var secondChar byte
	if firstChar != 0 {
		secondChar = FcToLower(lang[1:])
	}

	var low, high, mid, cmp int
	if firstChar < 'a' {
		low = 0
		high = fcLangCharSetRanges[0].begin
	} else if firstChar > 'z' {
		low = fcLangCharSetRanges[25].begin
		high = NUM_LANG_CHAR_SET - 1
	} else {
		low = fcLangCharSetRanges[firstChar-'a'].begin
		high = fcLangCharSetRanges[firstChar-'a'].end
		/* no matches */
		if low > high {
			/* one past next entry after where it would be */
			return -(low + 1)
		}
	}
	for low <= high {
		mid = (high + low) >> 1
		if fcLangCharSets[mid].lang[0] != firstChar {
			cmp = FcStrCmpIgnoreCase(fcLangCharSets[mid].lang, lang)
		} else {
			/* fast path for resolving 2-letter languages (by far the most common) after
			 * finding the first char (probably already true because of the hash table) */
			cmp = int(fcLangCharSets[mid].lang[1] - secondChar)
			if cmp == 0 && (len(fcLangCharSets[mid].lang) >= 2 || len(lang) >= 2) {
				cmp = FcStrCmpIgnoreCase(fcLangCharSets[mid].lang[2:], lang[2:])
			}
		}
		if cmp == 0 {
			return mid
		}
		if cmp < 0 {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	if cmp < 0 {
		mid++
	}
	return -(mid + 1)
}

func (ls *FcLangSet) FcLangSetAdd(lang string) {
	id := FcLangSetIndex(lang)
	if id >= 0 {
		ls.FcLangSetBitSet(id)
		return
	}
	if ls.extra == nil {
		ls.extra = make(FcStrSet)
	}
	ls.extra[lang] = true
}

//  FcBool
//  FcLangSetDel (ls *FcLangSet, const FcChar8 *lang)
//  {
// 	 int	id;

// 	 id = FcLangSetIndex (lang);
// 	 if (id >= 0)
// 	 {
// 	 FcLangSetBitReset (ls, id);
// 	 }
// 	 else if (ls.extra)
// 	 {
// 	 FcStrSetDel (ls.extra, lang);
// 	 }
// 	 return true;
//  }

func (ls *FcLangSet) hasLang(lang string) FcLangResult {
	id := FcLangSetIndex(lang)
	if id < 0 {
		id = -id - 1
	} else if ls.bitGet(id) {
		return FcLangEqual
	}
	best := FcLangDifferentLang
	for i := id - 1; i >= 0; i-- {
		r := FcLangCompare(lang, fcLangCharSets[i].lang)
		if r == FcLangDifferentLang {
			break
		}
		if ls.bitGet(i) && r < best {
			best = r
		}
	}
	for i := id; i < NUM_LANG_CHAR_SET; i++ {
		r := FcLangCompare(lang, fcLangCharSets[i].lang)
		if r == FcLangDifferentLang {
			break
		}
		if ls.bitGet(i) && r < best {
			best = r
		}
	}
	for extra := range ls.extra {
		if best <= FcLangEqual {
			break
		}
		if r := FcLangCompare(lang, extra); r < best {
			best = r
		}
	}
	return best
}

func (ls *FcLangSet) compareStrSet(set FcStrSet) FcLangResult {
	best := FcLangDifferentLang
	for extra := range set {
		if best <= FcLangEqual {
			break
		}
		if r := ls.hasLang(extra); r < best {
			best = r
		}
	}
	return best
}

func FcLangSetCompare(lsa, lsb *FcLangSet) FcLangResult {
	var aInCountrySet, bInCountrySet uint32

	for i := range lsa.map_ {
		if lsa.map_[i]&lsb.map_[i] != 0 {
			return FcLangEqual
		}
	}
	best := FcLangDifferentLang
	for j := 0; j < NUM_COUNTRY_SET; j++ {
		aInCountrySet = 0
		bInCountrySet = 0

		for i := range lsa.map_ {
			aInCountrySet |= lsa.map_[i] & fcLangCountrySets[j][i]
			bInCountrySet |= lsb.map_[i] & fcLangCountrySets[j][i]

			if aInCountrySet != 0 && bInCountrySet != 0 {
				best = FcLangDifferentTerritory
				break
			}
		}
	}
	if lsa.extra != nil {
		if r := lsb.compareStrSet(lsa.extra); r < best {
			best = r
		}
	}
	if best > FcLangEqual && lsb.extra != nil {
		if r := lsa.compareStrSet(lsb.extra); r < best {
			best = r
		}
	}
	return best
}

//  /*
//   * Used in computing values -- mustn't allocate any storage
//   */
//  FcLangSet *
//  FcLangSetPromote (const FcChar8 *lang, FcValuePromotionBuffer *vbuf)
//  {
// 	 int		id;
// 	 typedef struct {
// 	 FcLangSet  ls;
// 	 FcStrSet   strs;
// 	 FcChar8   *str;
// 	 } FcLangSetPromotionBuffer;
// 	 FcLangSetPromotionBuffer *buf = (FcLangSetPromotionBuffer *) vbuf;

// 	 FC_ASSERT_STATIC (sizeof (FcLangSetPromotionBuffer) <= sizeof (FcValuePromotionBuffer));

// 	 memset (buf.ls.map_, '\0', sizeof (buf.ls.map_));
// 	 buf.NUM_LANG_SET_MAP = NUM_LANG_SET_MAP;
// 	 buf.ls.extra = 0;
// 	 if (lang)
// 	 {
// 	 id = FcLangSetIndex (lang);
// 	 if (id >= 0)
// 	 {
// 		 FcLangSetBitSet (&buf.ls, id);
// 	 }
// 	 else
// 	 {
// 		 buf.ls.extra = &buf.strs;
// 		 buf.strs.num = 1;
// 		 buf.strs.size = 1;
// 		 buf.strs.strs = &buf.str;
// 		 FcRefInit (&buf.strs.ref, 1);
// 		 buf.str = (FcChar8 *) lang;
// 	 }
// 	 }
// 	 return &buf.ls;
//  }

//  FcChar32
//  FcLangSetHash (const FcLangSet *ls)
//  {
// 	 FcChar32	h = 0;
// 	 int		i, count;

// 	 count = FC_MIN (NUM_LANG_SET_MAP, NUM_LANG_SET_MAP);
// 	 for (i = 0; i < count; i++)
// 	 h ^= ls.map_[i];
// 	 if (ls.extra)
// 	 h ^= ls.extra.num;
// 	 return h;
//  }

//  FcLangSet *
//  FcNameParseLangSet (const FcChar8 *string)
//  {
// 	 FcChar8	    lang[32], c = 0;
// 	 int i;
// 	 FcLangSet	    *ls;

// 	 ls = FcLangSetCreate ();
// 	 if (!ls)
// 	 goto bail0;

// 	 for(;;)
// 	 {
// 	 for(i = 0; i < 31;i++)
// 	 {
// 		 c = *string++;
// 		 if(c == '\0' || c == '|')
// 		 break; /* end of this code */
// 		 lang[i] = c;
// 	 }
// 	 lang[i] = '\0';
// 	 if (!FcLangSetAdd (ls, lang))
// 		 goto bail1;
// 	 if(c == '\0')
// 		 break;
// 	 }
// 	 return ls;
//  bail1:
// 	 FcLangSetDestroy (ls);
//  bail0:
// 	 return 0;
//  }

//  FcBool
//  FcNameUnparseLangSet (FcStrBuf *buf, const FcLangSet *ls)
//  {
// 	 int		i, bit, count;
// 	 FcChar32	bits;
// 	 FcBool	first = true;

// 	 count = FC_MIN (NUM_LANG_SET_MAP, NUM_LANG_SET_MAP);
// 	 for (i = 0; i < count; i++)
// 	 {
// 	 if ((bits = ls.map_[i]))
// 	 {
// 		 for (bit = 0; bit <= 31; bit++)
// 		 if (bits & (1U << bit))
// 		 {
// 			 int id = (i << 5) | bit;
// 			 if (!first)
// 			 if (!FcStrBufChar (buf, '|'))
// 				 return false;
// 			 if (!FcStrBufString (buf, fcLangCharSets[fcLangCharSetIndicesInv[id]].lang))
// 			 return false;
// 			 first = false;
// 		 }
// 	 }
// 	 }
// 	 if (ls.extra)
// 	 {
// 	 FcStrList   *list = FcStrListCreate (ls.extra);
// 	 FcChar8	    *extra;

// 	 if (!list)
// 		 return false;
// 	 for ((extra = FcStrListNext (list)))
// 	 {
// 		 if (!first)
// 		 if (!FcStrBufChar (buf, '|'))
// 				 {
// 					 FcStrListDone (list);
// 			 return false;
// 				 }
// 		 if (!FcStrBufString (buf, extra))
// 				 {
// 					 FcStrListDone (list);
// 					 return false;
// 				 }
// 		 first = false;
// 	 }
// 		 FcStrListDone (list);
// 	 }
// 	 return true;
//  }

//  FcBool
//  FcLangSetEqual (const FcLangSet *lsa, const FcLangSet *lsb)
//  {
// 	 int	    i, count;

// 	 count = FC_MIN (lNUM_LANG_SET_MAP, lNUM_LANG_SET_MAP);
// 	 count = FC_MIN (NUM_LANG_SET_MAP, count);
// 	 for (i = 0; i < count; i++)
// 	 {
// 	 if (lsa.map_[i] != lsb.map_[i])
// 		 return false;
// 	 }
// 	 if (!lsa.extra && !lsb.extra)
// 	 return true;
// 	 if (lsa.extra && lsb.extra)
// 	 return FcStrSetEqual (lsa.extra, lsb.extra);
// 	 return false;
//  }

func (ls FcLangSet) containsLang(lang string) bool {
	id := FcLangSetIndex(lang)
	if id < 0 {
		id = -id - 1
	} else if ls.bitGet(id) {
		return true
	}
	// search up and down among equal languages for a match
	for i := id - 1; i >= 0; i-- {
		if FcLangCompare(fcLangCharSets[i].lang, lang) == FcLangDifferentLang {
			break
		}
		if ls.bitGet(i) && langContains(fcLangCharSets[i].lang, lang) {
			return true
		}
	}
	for i := id; i < NUM_LANG_CHAR_SET; i++ {
		if FcLangCompare(fcLangCharSets[i].lang, lang) == FcLangDifferentLang {
			break
		}
		if ls.bitGet(i) && langContains(fcLangCharSets[i].lang, lang) {
			return true
		}
	}

	var extra string
	for extra = range ls.extra {
		if langContains(extra, lang) {
			break
		}
	}
	if extra != "" {
		return true
	}
	return false
}

// return true if lsa contains every language in lsb
func (lsa FcLangSet) FcLangSetContains(lsb FcLangSet) bool {
	//  int		    i, j, count;
	//  FcChar32	    missing;

	if debugMode {
		fmt.Println("FcLangSet ", lsa)
		fmt.Println(" contains ", lsb)
		fmt.Println("")
	}
	// check bitmaps for missing language support
	for i := range lsb.map_ {
		missing := lsb.map_[i] & ^lsa.map_[i]
		if missing != 0 {
			for j := 0; j < 32; j++ {
				if missing&('\u0001'<<j) != 0 {
					tmpL := fcLangCharSets[fcLangCharSetIndicesInv[i*32+j]].lang
					if !lsa.containsLang(tmpL) {
						if debugMode {
							fmt.Printf("\tMissing bitmap %s\n", tmpL)
						}
						return false
					}
				}
			}
		}
	}
	var extra string
	for extra := range lsb.extra {
		if !lsa.containsLang(extra) {
			if debugMode {
				fmt.Printf("\tMissing string %s\n", extra)
			}
			break
		}
	}
	if extra != "" {
		return false
	}
	return true
}

//  FcBool
//  FcLangSetSerializeAlloc (FcSerialize *serialize, const FcLangSet *l)
//  {
// 	 if (!FcSerializeAlloc (serialize, l, sizeof (FcLangSet)))
// 	 return false;
// 	 return true;
//  }

//  FcLangSet *
//  FcLangSetSerialize(FcSerialize *serialize, const FcLangSet *l)
//  {
// 	 FcLangSet	*l_serialize = FcSerializePtr (serialize, l);

// 	 if (!l_serialize)
// 	 return NULL;
// 	 memset (l_serialize.map_, '\0', sizeof (l_serialize.map_));
// 	 memcpy (l_serialize.map_, l.map_, FC_MIN (sizeof (l_serialize.map_),NUM_LANG_SET_MAP * sizeof (l.map_[0])));
// 	 l_serialiNUM_LANG_SET_MAP = NUM_LANG_SET_MAP;
// 	 l_serialize.extra = NULL; /* We don't serialize ls.extra */
// 	 return l_serialize;
//  }

//  FcStrSet *
//  FcLangSetGetLangs (const FcLangSet *ls)
//  {
// 	 FcStrSet *langs;
// 	 int	      i;

// 	 langs = FcStrSetCreate();
// 	 if (!langs)
// 	 return 0;

// 	 for (i = 0; i < NUM_LANG_CHAR_SET; i++)
// 	 if (bitGet (ls, i))
// 		 FcStrSetAdd (langs, fcLangCharSets[i].lang);

// 	 if (ls.extra)
// 	 {
// 	 FcStrList	*list = FcStrListCreate (ls.extra);
// 	 FcChar8		*extra;

// 	 if (list)
// 	 {
// 		 for ((extra = FcStrListNext (list)))
// 		 FcStrSetAdd (langs, extra);

// 		 FcStrListDone (list);
// 	 }
// 	 }

// 	 return langs;
//  }

//  static FcLangSet *
//  FcLangSetOperate(const FcLangSet	*a,
// 		  const FcLangSet	*b,
// 		  FcBool			(*func) (FcLangSet 	*ls,
// 						  const FcChar8	*s))
//  {
// 	 FcLangSet	*langset = FcLangSetCopy (a);
// 	 FcStrSet	*set = FcLangSetGetLangs (b);
// 	 FcStrList	*sl = FcStrListCreate (set);
// 	 FcChar8	*str;

// 	 FcStrSetDestroy (set);
// 	 for ((str = FcStrListNext (sl)))
// 	 {
// 	 func (langset, str);
// 	 }
// 	 FcStrListDone (sl);

// 	 return langset;
//  }

//  FcLangSet *
//  FcLangSetUnion (const FcLangSet *a, const FcLangSet *b)
//  {
// 	 return FcLangSetOperate(a, b, FcLangSetAdd);
//  }

//  FcLangSet *
//  FcLangSetSubtract (const FcLangSet *a, const FcLangSet *b)
//  {
// 	 return FcLangSetOperate(a, b, FcLangSetDel);
//  }

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
