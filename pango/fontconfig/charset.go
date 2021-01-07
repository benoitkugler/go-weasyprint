package fontconfig

import (
	"fmt"
	"math/bits"
	"strconv"
	"strings"
)

// ported from fontconfig/src/fccharset.c   Copyright © 2001 Keith Packard

type FcCharSet struct {
	leaves  []*FcCharLeaf // same size 'num'
	numbers []uint16      // same size 'num'
}

// func  parseCharRange (str []byte) ([]byte, uint32, uint32, bool) {
// 	//  char *s = (char *) *string;
// 	//  char *t;
// 	//  long first, last;

// 	str = bytes.TrimLeftFunc(str, unicode.IsSpace)

// 	 t = s;
// 	 errno = 0;
// 	 first = last = strtol (s, &s, 16);
// 	strconv.ParseInt (string())

// 	 if (errno)
// 		 return false;
// 	 for (isspace(*s))
// 		 s++;
// 	 if (*s == '-')
// 	 {
// 		 s++;
// 		 errno = 0;
// 		 last = strtol (s, &s, 16);
// 		 if (errno)
// 		 return false;
// 	 }

// 	 if (s == t || first < 0 || last < 0 || last < first || last > 0x10ffff)
// 		  return false;

// 	 *string = (FcChar8 *) s;
// 	 *pfirst = first;
// 	 *plast = last;
// 	 return true;
//  }

func FcNameParseCharSet(str string) (FcCharSet, error) {
	fields := strings.Fields(str)

	var out FcCharSet
	for len(fields) != 0 {
		// either the separator is surrounded by space or not
		chunks := strings.Split(fields[0], "-")
		firstS, lastS := fields[0], ""
		if len(chunks) == 2 {
			firstS, lastS = chunks[0], chunks[1]
			fields = fields[1:]
		} else if strings.HasSuffix(fields[0], "-") && len(fields) >= 2 {
			firstS = strings.TrimSuffix(firstS, "-")
			lastS = fields[1]
			fields = fields[2:]
		} else if len(fields) >= 3 && fields[1] == "-" {
			lastS = fields[2]
			fields = fields[3:]
		}

		first, err := strconv.ParseInt(firstS, 16, 32)
		if err != nil {
			return out, fmt.Errorf("fontconfig: invalid charset: invalid first element %s: %s", firstS, err)
		}
		last := first
		if lastS != "" {
			last, err = strconv.ParseInt(lastS, 16, 32)
			if err != nil {
				return out, fmt.Errorf("fontconfig: invalid charset: invalid last element %s: %s", lastS, err)
			}
		}

		if first < 0 || last < 0 || last < first || last > 0x10ffff {
			return out, fmt.Errorf("invalid charset range %s %s", firstS, lastS)
		}

		for u := uint32(first); u < uint32(last+1); u++ {
			out.addChar(u)
		}
	}
	return out, nil
}

// Search for the leaf containing with the specified num.
// Return its index if it exists, otherwise return negative of
// the (position + 1) where it should be inserted
func (fcs FcCharSet) findLeafForward(start int, num uint16) int {
	numbers := fcs.numbers
	low := start
	high := len(numbers) - 1

	for low <= high {
		mid := (low + high) >> 1
		page := numbers[mid]
		if page == num {
			return mid
		}
		if page < num {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	if high < 0 || (high < len(numbers) && numbers[high] < num) {
		high++
	}
	return -(high + 1)
}

// Locate the leaf containing the specified char, return
// its index if it exists, otherwise return negative of
// the (position + 1) where it should be inserted
func (fcs FcCharSet) findLeafPos(ucs4 uint32) int {
	return fcs.findLeafForward(0, uint16(ucs4>>8))
}

func popCount(c1 uint32) uint32 { return uint32(bits.OnesCount32(c1)) }

// Returns the number of chars that are in `a` but not in `b`.
func FcCharSetSubtractCount(a, b FcCharSet) uint32 {
	var (
		count  uint32
		ai, bi FcCharSetIter
	)
	ai.start(a)
	bi.start(b)
	for ai.leaf != nil {
		if ai.ucs4 <= bi.ucs4 {
			am := *ai.leaf
			if ai.ucs4 == bi.ucs4 {
				bm := *bi.leaf
				for i := range am { // ; i != 0; i--
					count += popCount(am[i] & ^bm[i]) // *am++ & ~*bm++
				}
			} else {
				for i := range am { //; i != 0; i--
					count += popCount(am[i])
				}
			}
			ai.next(a)
		} else if bi.leaf != nil {
			bi.ucs4 = ai.ucs4
			bi.set(b)
		}
	}
	return count
}

// Returns whether `a` and `b` contain the same set of Unicode chars.
func FcCharSetEqual(a, b FcCharSet) bool {
	var ai, bi FcCharSetIter
	ai.start(a)
	bi.start(b)
	for ai.leaf != nil && bi.leaf != nil {
		if ai.ucs4 != bi.ucs4 {
			return false
		}
		if *ai.leaf != *bi.leaf {
			return false
		}
		ai.next(a)
		bi.next(b)
	}
	return ai.leaf == bi.leaf
}

// return true iff a is a subset of b
func (a FcCharSet) isSubset(b FcCharSet) bool {
	ai, bi := 0, 0
	for ai < len(a.numbers) && bi < len(b.numbers) {
		an := a.numbers[ai]
		bn := b.numbers[ai]
		// Check matching pages
		if an == bn {
			am := *a.leaves[ai]
			bm := *b.leaves[ai]

			if am != bm {
				//  Does am have any bits not in bm?
				for j, av := range am {
					if av & ^bm[j] != 0 {
						return false
					}
				}
			}
			ai++
			bi++
		} else if an < bn { // Does a have any pages not in b?
			return false
		} else {
			bi = b.findLeafForward(bi+1, an)
			if bi < 0 {
				bi = -bi - 1
			}
		}
	}
	//  did we look at every page?
	return ai >= len(a.numbers)
}

// Locate the leaf containing the specified char, creating it if desired
func (fcs *FcCharSet) findLeafCreate(ucs4 uint32) *FcCharLeaf {
	pos := fcs.findLeafPos(ucs4)
	if pos >= 0 {
		return fcs.leaves[pos]
	}

	leaf := new(FcCharLeaf)

	pos = -pos - 1
	if !fcs.putLeaf(ucs4, leaf, pos) {
		return nil
	}
	return leaf
}

// insert at pos
func (fcs *FcCharSet) putLeaf(ucs4 uint32, leaf *FcCharLeaf, pos int) bool {
	leaves := fcs.leaves
	numbers := fcs.numbers

	ucs4 >>= 8
	if ucs4 >= 0x10000 {
		return false
	}

	numbers = append(numbers, 0)
	leaves = append(leaves, nil)
	copy(numbers[pos+1:], numbers[pos:])
	copy(leaves[pos+1:], leaves[pos:])
	numbers[pos] = uint16(ucs4)
	leaves[pos] = leaf
	return true
}

func (fcs *FcCharSet) addLeaf(ucs4 uint32, leaf *FcCharLeaf) bool {
	new := fcs.findLeafCreate(ucs4)
	if new == nil {
		return false
	}
	*new = *leaf
	return true
}

func (fcs *FcCharSet) addChar(ucs4 uint32) bool {
	leaf := fcs.findLeafCreate(ucs4)
	if leaf == nil {
		return false
	}
	b := &leaf[(ucs4&0xff)>>5]
	*b |= (1 << (ucs4 & 0x1f))
	return true
}

func FcCharSetUnion(a, b FcCharSet) *FcCharSet {
	return operate(a, b, unionLeaf, true, true)
}

func operate(a, b FcCharSet, overlap func(result, al, bl *FcCharLeaf) bool,
	aonly, bonly bool) *FcCharSet {
	var ai, bi FcCharSetIter
	var fcs FcCharSet
	ai.start(a)
	bi.start(b)
	for (ai.leaf != nil || (bonly && bi.leaf != nil)) && (bi.leaf != nil || (aonly && ai.leaf != nil)) {
		if ai.ucs4 < bi.ucs4 {
			if aonly {
				if !fcs.addLeaf(ai.ucs4, ai.leaf) {
					return nil
				}
				ai.next(a)
			} else {
				ai.ucs4 = bi.ucs4
				ai.set(a)
			}
		} else if bi.ucs4 < ai.ucs4 {
			if bonly {
				if !fcs.addLeaf(bi.ucs4, bi.leaf) {
					return nil
				}
				bi.next(b)
			} else {
				bi.ucs4 = ai.ucs4
				bi.set(b)
			}
		} else {
			var leaf FcCharLeaf
			if overlap(&leaf, ai.leaf, bi.leaf) {
				if !fcs.addLeaf(ai.ucs4, &leaf) {
					return nil
				}
			}
			ai.next(a)
			bi.next(b)
		}
	}
	return &fcs
}

func unionLeaf(result, al, bl *FcCharLeaf) bool {
	for i := range result {
		result[i] = al[i] | bl[i]
	}
	return true
}

func subtractLeaf(result, al, bl *FcCharLeaf) bool {
	nonempty := false
	for i := range result {
		v := al[i] & ^bl[i]
		result[i] = v
		if v != 0 {
			nonempty = true
		}
	}
	return nonempty
}

func FcCharSetSubtract(a, b FcCharSet) *FcCharSet {
	return operate(a, b, subtractLeaf, true, false)
}

// FcCharSetIter is an iterator for the leaves of a charset
type FcCharSetIter struct {
	leaf *FcCharLeaf
	ucs4 uint32
	pos  int
}

// Set iter.leaf to the leaf containing iter.ucs4 or higher
func (iter *FcCharSetIter) set(fcs FcCharSet) {
	pos := fcs.findLeafPos(iter.ucs4)

	if pos < 0 {
		pos = -pos - 1
		if pos == len(fcs.numbers) {
			iter.ucs4 = ^uint32(0)
			iter.leaf = nil
			return
		}
		iter.ucs4 = uint32(fcs.numbers[pos]) << 8
	}
	iter.leaf = fcs.leaves[pos]
	iter.pos = pos
}

func (iter *FcCharSetIter) start(fcs FcCharSet) {
	iter.ucs4 = 0
	iter.pos = 0
	iter.set(fcs)
}

func (iter *FcCharSetIter) next(fcs FcCharSet) {
	pos := iter.pos + 1
	if pos >= len(fcs.numbers) {
		iter.ucs4 = ^uint32(0)
		iter.leaf = nil
	} else {
		iter.ucs4 = uint32(fcs.numbers[pos]) << 8
		iter.leaf = fcs.leaves[pos]
		iter.pos = pos
	}
}

//  FcCharSet *
//  FcCharSetCreate (void)
//  {
// 	 FcCharSet	*fcs;

// 	 fcs = (FcCharSet *) malloc (sizeof (FcCharSet));
// 	 if (!fcs)
// 	 return 0;
// 	 FcRefInit (&fcs.ref, 1);
// 	 len(fcs.numbers) = 0;
// 	 fcs.leaves_offset = 0;
// 	 fcs.numbers_offset = 0;
// 	 return fcs;
//  }

//  FcCharSet *
//  FcCharSetNew (void)
//  {
// 	 return FcCharSetCreate ();
//  }

//  void
//  FcCharSetDestroy (fcs *FcCharSet)
//  {
// 	 int i;

// 	 if (fcs)
// 	 {
// 	 if (FcRefIsConst (&fcs.ref))
// 	 {
// 		 FcCacheObjectDereference (fcs);
// 		 return;
// 	 }
// 	 if (FcRefDec (&fcs.ref) != 1)
// 		 return;
// 	 for (i = 0; i < len(fcs.numbers); i++)
// 		 free (FcCharSetLeaf (fcs, i));
// 	 if (len(fcs.numbers))
// 	 {
// 		 free (FcCharSetLeaves (fcs));
// 		 free (FcCharSetNumbers (fcs));
// 	 }
// 	 free (fcs);
// 	 }
//  }

//  static FcCharLeaf *
//  FcCharSetFindLeaf (fcs *FcCharSet, ucs4 uint32 )
//  {
// 	 int	pos = findLeafPos (fcs, ucs4);
// 	 if (pos >= 0)
// 	 return fcs.leaves[pos];
// 	 return 0;
//  }

//  static FcBool
//  FcCharSetInsertLeaf (fcs *FcCharSet, ucs4 uint32 , FcCharLeaf *leaf)
//  {
// 	 int		    pos;

// 	 pos = findLeafPos (fcs, ucs4);
// 	 if (pos >= 0)
// 	 {
// 	 free (FcCharSetLeaf (fcs, pos));
// 	 FcCharSetLeaves(fcs)[pos] = FcPtrToOffset (FcCharSetLeaves(fcs),
// 							leaf);
// 	 return true;
// 	 }
// 	 pos = -pos - 1;
// 	 return putLeaf (fcs, ucs4, leaf, pos);
//  }

//  FcBool
//  FcCharSetDelChar (fcs *FcCharSet, ucs4 uint32 )
//  {
// 	 leaf *FcCharLeaf;
// 	 uint32	*b;

// 	 if (fcs == NULL || FcRefIsConst (&fcs.ref))
// 	 return false;
// 	 leaf = FcCharSetFindLeaf (fcs, ucs4);
// 	 if (!leaf)
// 	 return true;
// 	 b = &leaf.map_[(ucs4 & 0xff) >> 5];
// 	 *b &= ~(1U << (ucs4 & 0x1f));
// 	 /* We don't bother removing the leaf if it's empty */
// 	 return true;
//  }

//  FcCharSet *
//  FcCharSetCopy (FcCharSet *src)
//  {
// 	 if (src)
// 	 {
// 	 if (!FcRefIsConst (&src.ref))
// 		 FcRefInc (&src.ref);
// 	 else
// 		 FcCacheObjectReference (src);
// 	 }
// 	 return src;
//  }

//  static FcBool
//  FcCharSetIntersectLeaf (FcCharLeaf *result,
// 			 const FcCharLeaf *al,
// 			 const FcCharLeaf *bl)
//  {
// 	 int	    i;
// 	 FcBool  nonempty = false;

// 	 for (i = 0; i < 256/32; i++)
// 	 if ((result.map_[i] = al.map_[i] & bl.map_[i]))
// 		 nonempty = true;
// 	 return nonempty;
//  }

//  FcCharSet *
//  FcCharSetIntersect (const FcCharSet *a, const FcCharSet *b)
//  {
// 	 return operate (a, b, FcCharSetIntersectLeaf, false, false);
//  }

// Adds all chars in `b` to `a`.
// In other words, this is an in-place version of FcCharSetUnion.
// It returns whether any new chars from `b` were added to `a`.
func (a *FcCharSet) merge(b FcCharSet) bool {
	//  int		ai = 0, bi = 0;
	//  FcChar16	an, bn;

	if a == nil {
		return false
	}

	changed := !b.isSubset(*a)
	if !changed {
		return changed
	}

	for ai, bi := 0, 0; bi < len(b.numbers); {
		an := ^uint16(0)
		if ai < len(a.numbers) {
			an = a.numbers[ai]
		}
		bn := b.numbers[bi]

		if an < bn {
			ai = a.findLeafForward(ai+1, bn)
			if ai < 0 {
				ai = -ai - 1
			}
		} else {
			bl := b.leaves[bi]
			if bn < an {
				if !a.addLeaf(uint32(bn)<<8, bl) {
					return false
				}
			} else {
				al := a.leaves[ai]
				unionLeaf(al, al, bl)
			}

			ai++
			bi++
		}
	}

	return changed
}

//  FcBool
//  FcCharSetHasChar (fcs *FcCharSet, ucs4 uint32 )
//  {
// 	 leaf *FcCharLeaf;

// 	 if (!fcs)
// 	 return false;
// 	 leaf = FcCharSetFindLeaf (fcs, ucs4);
// 	 if (!leaf)
// 	 return false;
// 	 return (leaf.map_[(ucs4 & 0xff) >> 5] & (1U << (ucs4 & 0x1f))) != 0;
//  }

//  uint32
//  FcCharSetIntersectCount (const FcCharSet *a, const FcCharSet *b)
//  {
// 	 FcCharSetIter   ai, bi;
// 	 uint32	    count = 0;

// 	 if (a && b)
// 	 {
// 	 start (a, &ai);
// 	 start (b, &bi);
// 	 for (ai.leaf && bi.leaf)
// 	 {
// 		 if (ai.ucs4 == bi.ucs4)
// 		 {
// 		 uint32	*am = ai.leaf.map_;
// 		 uint32	*bm = bi.leaf.map_;
// 		 int		i = 256/32;
// 		 for (i--)
// 			 count += popCount (*am++ & *bm++);
// 		 next (a, &ai);
// 		 }
// 		 else if (ai.ucs4 < bi.ucs4)
// 		 {
// 		 ai.ucs4 = bi.ucs4;
// 		 set (a, &ai);
// 		 }
// 		 if (bi.ucs4 < ai.ucs4)
// 		 {
// 		 bi.ucs4 = ai.ucs4;
// 		 set (b, &bi);
// 		 }
// 	 }
// 	 }
// 	 return count;
//  }

//  uint32
//  FcCharSetCount (const FcCharSet *a)
//  {
// 	 FcCharSetIter   ai;
// 	 uint32	    count = 0;

// 	 if (a)
// 	 {
// 	 for (start (a, &ai); ai.leaf; next (a, &ai))
// 	 {
// 		 int		    i = 256/32;
// 		 uint32	    *am = ai.leaf.map_;

// 		 for (i--)
// 		 count += popCount (*am++);
// 	 }
// 	 }
// 	 return count;
//  }

//  /*
//   * These two functions efficiently walk the entire charmap for
//   * other software (like pango) that want their own copy
//   */

//  uint32
//  FcCharSetNextPage (const FcCharSet  *a,
// 			uint32	    map_[FC_CHARSET_MAP_SIZE],
// 			uint32	    *next)
//  {
// 	 FcCharSetIter   ai;
// 	 uint32	    page;

// 	 if (!a)
// 	 return FC_CHARSET_DONE;
// 	 ai.ucs4 = *next;
// 	 set (a, &ai);
// 	 if (!ai.leaf)
// 	 return FC_CHARSET_DONE;

// 	 /*
// 	  * Save current information
// 	  */
// 	 page = ai.ucs4;
// 	 memcpy (map_, ai.leaf.map_, sizeof (ai.leaf.map_));
// 	 /*
// 	  * Step to next page
// 	  */
// 	 next (a, &ai);
// 	 *next = ai.ucs4;

// 	 return page;
//  }

//  uint32
//  FcCharSetFirstPage (const FcCharSet *a,
// 			 uint32	    map_[FC_CHARSET_MAP_SIZE],
// 			 uint32	    *next)
//  {
// 	 *next = 0;
// 	 return FcCharSetNextPage (a, map_, next);
//  }

//  /*
//   * old coverage API, rather hard to use correctly
//   */

//  uint32
//  FcCharSetCoverage (const FcCharSet *a, uint32 page, uint32 *result)
//  {
// 	 FcCharSetIter   ai;

// 	 ai.ucs4 = page;
// 	 set (a, &ai);
// 	 if (!ai.leaf)
// 	 {
// 	 memset (result, '\0', 256 / 8);
// 	 page = 0;
// 	 }
// 	 else
// 	 {
// 	 memcpy (result, ai.leaf.map_, sizeof (ai.leaf.map_));
// 	 next (a, &ai);
// 	 page = ai.ucs4;
// 	 }
// 	 return page;
//  }

//  static void
//  FcNameUnparseUnicode (FcStrBuf *buf, uint32 u)
//  {
// 	 FcChar8	    buf_static[64];
// 	 snprintf ((char *) buf_static, sizeof (buf_static), "%x", u);
// 	 FcStrBufString (buf, buf_static);
//  }

//  FcBool
//  FcNameUnparseCharSet (FcStrBuf *buf, const FcCharSet *c)
//  {
// 	 FcCharSetIter   ci;
// 	 uint32	    first, last;
// 	 int		    i;
//  #ifdef CHECK
// 	 int		    len = buf.len;
//  #endif

// 	 first = last = 0x7FFFFFFF;

// 	 for (start (c, &ci);
// 	  ci.leaf;
// 	  next (c, &ci))
// 	 {
// 	 for (i = 0; i < 256/32; i++)
// 	 {
// 		 uint32 bits = ci.leaf.map_[i];
// 		 uint32 u = ci.ucs4 + i * 32;

// 		 for (bits)
// 		 {
// 		 if (bits & 1)
// 		 {
// 			 if (u != last + 1)
// 			 {
// 				 if (last != first)
// 				 {
// 				 FcStrBufChar (buf, '-');
// 				 FcNameUnparseUnicode (buf, last);
// 				 }
// 				 if (last != 0x7FFFFFFF)
// 				 FcStrBufChar (buf, ' ');
// 				 /* Start new range. */
// 				 first = u;
// 				 FcNameUnparseUnicode (buf, u);
// 			 }
// 			 last = u;
// 		 }
// 		 bits >>= 1;
// 		 u++;
// 		 }
// 	 }
// 	 }
// 	 if (last != first)
// 	 {
// 	 FcStrBufChar (buf, '-');
// 	 FcNameUnparseUnicode (buf, last);
// 	 }
//  #ifdef CHECK
// 	 {
// 	 FcCharSet	*check;
// 	 uint32	missing;
// 	 FcCharSetIter	ci, checki;

// 	 /* null terminate for parser */
// 	 FcStrBufChar (buf, '\0');
// 	 /* step back over null for life after test */
// 	 buf.len--;
// 	 check = FcNameParseCharSet (buf.buf + len);
// 	 start (c, &ci);
// 	 start (check, &checki);
// 	 for (ci.leaf || checki.leaf)
// 	 {
// 		 if (ci.ucs4 < checki.ucs4)
// 		 {
// 		 printf ("Missing leaf node at 0x%x\n", ci.ucs4);
// 		 next (c, &ci);
// 		 }
// 		 else if (checki.ucs4 < ci.ucs4)
// 		 {
// 		 printf ("Extra leaf node at 0x%x\n", checki.ucs4);
// 		 next (check, &checki);
// 		 }
// 		 else
// 		 {
// 		 int	    i = 256/32;
// 		 uint32    *cm = ci.leaf.map_;
// 		 uint32    *checkm = checki.leaf.map_;

// 		 for (i = 0; i < 256; i += 32)
// 		 {
// 			 if (*cm != *checkm)
// 			 printf ("Mismatching sets at 0x%08x: 0x%08x != 0x%08x\n",
// 				 ci.ucs4 + i, *cm, *checkm);
// 			 cm++;
// 			 checkm++;
// 		 }
// 		 next (c, &ci);
// 		 next (check, &checki);
// 		 }
// 	 }
// 	 if ((missing = FcCharSetSubtractCount (c, check)))
// 		 printf ("%d missing in reparsed result\n", missing);
// 	 if ((missing = FcCharSetSubtractCount (check, c)))
// 		 printf ("%d extra in reparsed result\n", missing);
// 	 FcCharSetDestroy (check);
// 	 }
//  #endif

// 	 return true;
//  }

//  typedef struct _FcCharLeafEnt FcCharLeafEnt;

//  struct _FcCharLeafEnt {
// 	 FcCharLeafEnt   *next;
// 	 uint32	    hash;
// 	 FcCharLeaf	    leaf;
//  };

//  #define FC_CHAR_LEAF_BLOCK	(4096 / sizeof (FcCharLeafEnt))
//  #define FC_CHAR_LEAF_HASH_SIZE	257

//  typedef struct _FcCharSetEnt FcCharSetEnt;

//  struct _FcCharSetEnt {
// 	 FcCharSetEnt	*next;
// 	 uint32		hash;
// 	 FcCharSet		set;
//  };

//  typedef struct _FcCharSetOrigEnt FcCharSetOrigEnt;

//  struct _FcCharSetOrigEnt {
// 	 FcCharSetOrigEnt	*next;
// 	 const FcCharSet    	*orig;
// 	 const FcCharSet    	*frozen;
//  };

//  #define FC_CHAR_SET_HASH_SIZE    67

//  struct _FcCharSetFreezer {
// 	 FcCharLeafEnt   *leaf_hash_table[FC_CHAR_LEAF_HASH_SIZE];
// 	 FcCharLeafEnt   **leaf_blocks;
// 	 int		    leaf_block_count;
// 	 FcCharSetEnt    *set_hash_table[FC_CHAR_SET_HASH_SIZE];
// 	 FcCharSetOrigEnt	*orig_hash_table[FC_CHAR_SET_HASH_SIZE];
// 	 FcCharLeafEnt   *current_block;
// 	 int		    leaf_remain;
// 	 int		    leaves_seen;
// 	 int		    charsets_seen;
// 	 int		    leaves_allocated;
// 	 int		    charsets_allocated;
//  };

//  static FcCharLeafEnt *
//  FcCharLeafEntCreate (FcCharSetFreezer *freezer)
//  {
// 	 if (!freezer.leaf_remain)
// 	 {
// 	 FcCharLeafEnt **newBlocks;

// 	 freezer.leaf_block_count++;
// 	 newBlocks = realloc (freezer.leaf_blocks, freezer.leaf_block_count * sizeof (FcCharLeafEnt *));
// 	 if (!newBlocks)
// 		 return 0;
// 	 freezer.leaf_blocks = newBlocks;
// 	 freezer.current_block = freezer.leaf_blocks[freezer.leaf_block_count-1] = malloc (FC_CHAR_LEAF_BLOCK * sizeof (FcCharLeafEnt));
// 	 if (!freezer.current_block)
// 		 return 0;
// 	 freezer.leaf_remain = FC_CHAR_LEAF_BLOCK;
// 	 }
// 	 freezer.leaf_remain--;
// 	 freezer.leaves_allocated++;
// 	 return freezer.current_block++;
//  }

//  static uint32
//  FcCharLeafHash (FcCharLeaf *leaf)
//  {
// 	 uint32	hash = 0;
// 	 int		i;

// 	 for (i = 0; i < 256/32; i++)
// 	 hash = ((hash << 1) | (hash >> 31)) ^ leaf.map_[i];
// 	 return hash;
//  }

//  static FcCharLeaf *
//  FcCharSetFreezeLeaf (FcCharSetFreezer *freezer, FcCharLeaf *leaf)
//  {
// 	 uint32			hash = FcCharLeafHash (leaf);
// 	 FcCharLeafEnt		**bucket = &freezer.leaf_hash_table[hash % FC_CHAR_LEAF_HASH_SIZE];
// 	 FcCharLeafEnt		*ent;

// 	 for (ent = *bucket; ent; ent = ent.next)
// 	 {
// 	 if (ent.hash == hash && !memcmp (&ent.leaf, leaf, sizeof (FcCharLeaf)))
// 		 return &ent.leaf;
// 	 }

// 	 ent = FcCharLeafEntCreate(freezer);
// 	 if (!ent)
// 	 return 0;
// 	 ent.leaf = *leaf;
// 	 ent.hash = hash;
// 	 ent.next = *bucket;
// 	 *bucket = ent;
// 	 return &ent.leaf;
//  }

//  static uint32
//  FcCharSetHash (fcs *FcCharSet)
//  {
// 	 uint32	hash = 0;
// 	 int		i;

// 	 /* hash in leaves */
// 	 for (i = 0; i < len(fcs.numbers); i++)
// 	 hash = ((hash << 1) | (hash >> 31)) ^ FcCharLeafHash (FcCharSetLeaf(fcs,i));
// 	 /* hash in numbers */
// 	 for (i = 0; i < len(fcs.numbers); i++)
// 	 hash = ((hash << 1) | (hash >> 31)) ^ fcs.numbers[i];
// 	 return hash;
//  }

//  static FcBool
//  FcCharSetFreezeOrig (FcCharSetFreezer *freezer, const FcCharSet *orig, const FcCharSet *frozen)
//  {
// 	 FcCharSetOrigEnt	**bucket = &freezer.orig_hash_table[((uintptr_t) orig) % FC_CHAR_SET_HASH_SIZE];
// 	 FcCharSetOrigEnt	*ent;

// 	 ent = malloc (sizeof (FcCharSetOrigEnt));
// 	 if (!ent)
// 	 return false;
// 	 ent.orig = orig;
// 	 ent.frozen = frozen;
// 	 ent.next = *bucket;
// 	 *bucket = ent;
// 	 return true;
//  }

//  static FcCharSet *
//  FcCharSetFreezeBase (FcCharSetFreezer *freezer, fcs *FcCharSet)
//  {
// 	 uint32		hash = FcCharSetHash (fcs);
// 	 FcCharSetEnt	**bucket = &freezer.set_hash_table[hash % FC_CHAR_SET_HASH_SIZE];
// 	 FcCharSetEnt	*ent;
// 	 int			size;
// 	 int			i;

// 	 for (ent = *bucket; ent; ent = ent.next)
// 	 {
// 	 if (ent.hash == hash &&
// 		 ent.set.num == len(fcs.numbers) &&
// 		 !memcmp (FcCharSetNumbers(&ent.set),
// 			  fcs.numbers,
// 			  len(fcs.numbers) * sizeof (FcChar16)))
// 	 {
// 		 FcBool ok = true;
// 		 int i;

// 		 for (i = 0; i < len(fcs.numbers); i++)
// 		 if (FcCharSetLeaf(&ent.set, i) != FcCharSetLeaf(fcs, i))
// 			 ok = false;
// 		 if (ok)
// 		 return &ent.set;
// 	 }
// 	 }

// 	 size = (sizeof (FcCharSetEnt) +
// 		 len(fcs.numbers) * sizeof (FcCharLeaf *) +
// 		 len(fcs.numbers) * sizeof (FcChar16));
// 	 ent = malloc (size);
// 	 if (!ent)
// 	 return 0;

// 	 freezer.charsets_allocated++;

// 	 FcRefSetConst (&ent.set.ref);
// 	 ent.set.num = len(fcs.numbers);
// 	 if (len(fcs.numbers))
// 	 {
// 	 intptr_t    *ent_leaves;

// 	 ent.set.leaves_offset = sizeof (ent.set);
// 	 ent.set.numbers_offset = (ent.set.leaves_offset +
// 					len(fcs.numbers) * sizeof (intptr_t));

// 	 ent_leaves = FcCharSetLeaves (&ent.set);
// 	 for (i = 0; i < len(fcs.numbers); i++)
// 		 ent_leaves[i] = FcPtrToOffset (ent_leaves,
// 						FcCharSetLeaf (fcs, i));
// 	 memcpy (FcCharSetNumbers (&ent.set),
// 		 FcCharSetNumbers (fcs),
// 		 len(fcs.numbers) * sizeof (FcChar16));
// 	 }
// 	 else
// 	 {
// 	 ent.set.leaves_offset = 0;
// 	 ent.set.numbers_offset = 0;
// 	 }

// 	 ent.hash = hash;
// 	 ent.next = *bucket;
// 	 *bucket = ent;

// 	 return &ent.set;
//  }

//  static const FcCharSet *
//  FcCharSetFindFrozen (FcCharSetFreezer *freezer, const FcCharSet *orig)
//  {
// 	 FcCharSetOrigEnt    **bucket = &freezer.orig_hash_table[((uintptr_t) orig) % FC_CHAR_SET_HASH_SIZE];
// 	 FcCharSetOrigEnt	*ent;

// 	 for (ent = *bucket; ent; ent = ent.next)
// 	 if (ent.orig == orig)
// 		 return ent.frozen;
// 	 return NULL;
//  }

//  const FcCharSet *
//  FcCharSetFreeze (FcCharSetFreezer *freezer, fcs *FcCharSet)
//  {
// 	 FcCharSet	    *b;
// 	 const FcCharSet *n = 0;
// 	 FcCharLeaf	    *l;
// 	 int		    i;

// 	 b = FcCharSetCreate ();
// 	 if (!b)
// 	 goto bail0;
// 	 for (i = 0; i < len(fcs.numbers); i++)
// 	 {
// 	 l = FcCharSetFreezeLeaf (freezer, FcCharSetLeaf(fcs, i));
// 	 if (!l)
// 		 goto bail1;
// 	 if (!FcCharSetInsertLeaf (b, fcs.numbers[i] << 8, l))
// 		 goto bail1;
// 	 }
// 	 n = FcCharSetFreezeBase (freezer, b);
// 	 if (!FcCharSetFreezeOrig (freezer, fcs, n))
// 	 {
// 	 n = NULL;
// 	 goto bail1;
// 	 }
// 	 freezer.charsets_seen++;
// 	 freezer.leaves_seen += len(fcs.numbers);
//  bail1:
// 	 if (b.num)
// 	 free (FcCharSetLeaves (b));
// 	 if (b.num)
// 	 free (FcCharSetNumbers (b));
// 	 free (b);
//  bail0:
// 	 return n;
//  }

//  FcCharSetFreezer *
//  FcCharSetFreezerCreate (void)
//  {
// 	 FcCharSetFreezer	*freezer;

// 	 freezer = calloc (1, sizeof (FcCharSetFreezer));
// 	 return freezer;
//  }

//  void
//  FcCharSetFreezerDestroy (FcCharSetFreezer *freezer)
//  {
// 	 int i;

// 	 if (FcDebug() & FC_DBG_CACHE)
// 	 {
// 	 printf ("\ncharsets %d . %d leaves %d . %d\n",
// 		 freezer.charsets_seen, freezer.charsets_allocated,
// 		 freezer.leaves_seen, freezer.leaves_allocated);
// 	 }
// 	 for (i = 0; i < FC_CHAR_SET_HASH_SIZE; i++)
// 	 {
// 	 FcCharSetEnt	*ent, *next;
// 	 for (ent = freezer.set_hash_table[i]; ent; ent = next)
// 	 {
// 		 next = ent.next;
// 		 free (ent);
// 	 }
// 	 }

// 	 for (i = 0; i < FC_CHAR_SET_HASH_SIZE; i++)
// 	 {
// 	 FcCharSetOrigEnt	*ent, *next;
// 	 for (ent = freezer.orig_hash_table[i]; ent; ent = next)
// 	 {
// 		 next = ent.next;
// 		 free (ent);
// 	 }
// 	 }

// 	 for (i = 0; i < freezer.leaf_block_count; i++)
// 	 free (freezer.leaf_blocks[i]);

// 	 free (freezer.leaf_blocks);
// 	 free (freezer);
//  }

//  FcBool
//  FcCharSetSerializeAlloc (FcSerialize *serialize, const FcCharSet *cs)
//  {
// 	 intptr_t	    *leaves;
// 	 FcChar16	    *numbers;
// 	 int		    i;

// 	 if (!FcRefIsConst (&cs.ref))
// 	 {
// 	 if (!serialize.cs_freezer)
// 	 {
// 		 serialize.cs_freezer = FcCharSetFreezerCreate ();
// 		 if (!serialize.cs_freezer)
// 		 return false;
// 	 }
// 	 if (FcCharSetFindFrozen (serialize.cs_freezer, cs))
// 		 return true;

// 		 cs = FcCharSetFreeze (serialize.cs_freezer, cs);
// 	 }

// 	 leaves = FcCharSetLeaves (cs);
// 	 numbers = FcCharSetNumbers (cs);

// 	 if (!FcSerializeAlloc (serialize, cs, sizeof (FcCharSet)))
// 	 return false;
// 	 if (!FcSerializeAlloc (serialize, leaves, cs.num * sizeof (intptr_t)))
// 	 return false;
// 	 if (!FcSerializeAlloc (serialize, numbers, cs.num * sizeof (FcChar16)))
// 	 return false;
// 	 for (i = 0; i < cs.num; i++)
// 	 if (!FcSerializeAlloc (serialize, FcCharSetLeaf(cs, i),
// 					sizeof (FcCharLeaf)))
// 		 return false;
// 	 return true;
//  }

//  FcCharSet *
//  FcCharSetSerialize(FcSerialize *serialize, const FcCharSet *cs)
//  {
// 	 FcCharSet	*cs_serialized;
// 	 intptr_t	*leaves, *leaves_serialized;
// 	 FcChar16	*numbers, *numbers_serialized;
// 	 leaf *FcCharLeaf, *leaf_serialized;
// 	 int		i;

// 	 if (!FcRefIsConst (&cs.ref) && serialize.cs_freezer)
// 	 {
// 	 cs = FcCharSetFindFrozen (serialize.cs_freezer, cs);
// 	 if (!cs)
// 		 return NULL;
// 	 }

// 	 cs_serialized = FcSerializePtr (serialize, cs);
// 	 if (!cs_serialized)
// 	 return NULL;

// 	 FcRefSetConst (&cs_serialized.ref);
// 	 cs_serialized.num = cs.num;

// 	 if (cs.num)
// 	 {
// 	 leaves = FcCharSetLeaves (cs);
// 	 leaves_serialized = FcSerializePtr (serialize, leaves);
// 	 if (!leaves_serialized)
// 		 return NULL;

// 	 cs_serialized.leaves_offset = FcPtrToOffset (cs_serialized,
// 							   leaves_serialized);

// 	 numbers = FcCharSetNumbers (cs);
// 	 numbers_serialized = FcSerializePtr (serialize, numbers);
// 	 if (!numbers)
// 		 return NULL;

// 	 cs_serialized.numbers_offset = FcPtrToOffset (cs_serialized,
// 								numbers_serialized);

// 	 for (i = 0; i < cs.num; i++)
// 	 {
// 		 leaf = FcCharSetLeaf (cs, i);
// 		 leaf_serialized = FcSerializePtr (serialize, leaf);
// 		 if (!leaf_serialized)
// 		 return NULL;
// 		 *leaf_serialized = *leaf;
// 		 leaves_serialized[i] = FcPtrToOffset (leaves_serialized,
// 						   leaf_serialized);
// 		 numbers_serialized[i] = numbers[i];
// 	 }
// 	 }
// 	 else
// 	 {
// 	 cs_serialized.leaves_offset = 0;
// 	 cs_serialized.numbers_offset = 0;
// 	 }

// 	 return cs_serialized;
//  }
