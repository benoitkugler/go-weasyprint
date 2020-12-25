// Golang port of fribidi, using golang.org/x/text/unicode/bidi
// as data source.
package fribidi

import (
	"golang.org/x/text/unicode/bidi"
)

type Level int8

func (lev Level) isRtl() Level { return lev & 1 }

func maxL(l1, l2 Level) Level {
	if l1 < l2 {
		return l2
	}
	return l1
}

const (
	/* The maximum embedding level value assigned by explicit marks */
	FRIBIDI_BIDI_MAX_EXPLICIT_LEVEL = 125

	/* The maximum *number* of different resolved embedding levels: 0-126 */
	FRIBIDI_BIDI_MAX_RESOLVED_LEVELS = 127

	LOCAL_BRACKET_SIZE = 16

	/* The maximum *number* of nested brackets: 0-63 */
	FRIBIDI_BIDI_MAX_NESTED_BRACKET_PAIRS = 63
)

const (
	FRIBIDI_CHAR_LRM = 0x200E
	FRIBIDI_CHAR_RLM = 0x200F
	FRIBIDI_CHAR_LRE = 0x202A
	FRIBIDI_CHAR_RLE = 0x202B
	FRIBIDI_CHAR_PDF = 0x202C
	FRIBIDI_CHAR_LRO = 0x202D
	FRIBIDI_CHAR_RLO = 0x202E
	FRIBIDI_CHAR_LRI = 0x2066
	FRIBIDI_CHAR_RLI = 0x2067
	FRIBIDI_CHAR_FSI = 0x2068
	FRIBIDI_CHAR_PDI = 0x2069
)

/*
 * Define bit masks that bidi types are based on, each mask has
 * only one bit set.
 */
const (
	/* RTL mask better be the least significant bit. */
	maskRTL    = 0x00000001 /* Is right to left */
	maskARABIC = 0x00000002 /* Is arabic */

	/* Each char can be only one of the three following. */
	maskSTRONG   = 0x00000010 /* Is strong */
	maskWEAK     = 0x00000020 /* Is weak */
	maskNEUTRAL  = 0x00000040 /* Is neutral */
	maskSENTINEL = 0x00000080 /* Is sentinel */
	/* Sentinels are not valid chars, just identify the start/end of strings. */

	/* Each char can be only one of the six following. */
	maskLETTER    = 0x00000100 /* Is letter: L, R, AL */
	maskNUMBER    = 0x00000200 /* Is number: EN, AN */
	maskNUMSEPTER = 0x00000400 /* Is separator or terminator: ES, ET, CS */
	maskSPACE     = 0x00000800 /* Is space: BN, BS, SS, WS */
	maskEXPLICIT  = 0x00001000 /* Is explicit mark: LRE, RLE, LRO, RLO, PDF */
	maskISOLATE   = 0x00008000 /* Is isolate mark: LRI, RLI, FSI, PDI */

	/* Can be set only if maskSPACE is also set. */
	maskSEPARATOR = 0x00002000 /* Is text separator: BS, SS */
	/* Can be set only if maskEXPLICIT is also set. */
	maskOVERRIDE = 0x00004000 /* Is explicit override: LRO, RLO */
	maskFIRST    = 0x02000000 /* Whether direction is determined by first strong */

	/* The following exist to make types pairwise different, some of them can
	 * be removed but are here because of efficiency (make queries faster). */

	maskES = 0x00010000
	maskET = 0x00020000
	maskCS = 0x00040000

	maskNSM = 0x00080000
	maskBN  = 0x00100000

	maskBS = 0x00200000
	maskSS = 0x00400000
	maskWS = 0x00800000

	/* We reserve a single bit for user's private use: we will never use it. */
	maskPRIVATE = 0x01000000
)

type ParType = CharType

const (
	LTR  = maskSTRONG | maskLETTER
	RTL  = maskSTRONG | maskLETTER | maskRTL
	EN   = maskWEAK | maskNUMBER
	ON   = maskNEUTRAL
	WLTR = maskWEAK
	WRTL = maskWEAK | maskRTL
	PDF  = maskWEAK | maskEXPLICIT
	LRI  = maskNEUTRAL | maskISOLATE
	RLI  = maskNEUTRAL | maskISOLATE | maskRTL
	FSI  = maskNEUTRAL | maskISOLATE | maskFIRST
	BS   = maskNEUTRAL | maskSPACE | maskSEPARATOR | maskBS
	NSM  = maskWEAK | maskNSM
	AL   = maskSTRONG | maskLETTER | maskRTL | maskARABIC
	AN   = maskWEAK | maskNUMBER | maskARABIC
	CS   = maskWEAK | maskNUMSEPTER | maskCS
	ET   = maskWEAK | maskNUMSEPTER | maskET
	PDI  = maskNEUTRAL | maskWEAK | maskISOLATE // Pop Directional Isolate
	LRO  = maskSTRONG | maskEXPLICIT | maskOVERRIDE
	RLO  = maskSTRONG | maskEXPLICIT | maskRTL | maskOVERRIDE
	RLE  = maskSTRONG | maskEXPLICIT | maskRTL
	LRE  = maskSTRONG | maskEXPLICIT
	WS   = maskNEUTRAL | maskSPACE | maskWS
	ES   = maskWEAK | maskNUMSEPTER | maskES
	BN   = maskWEAK | maskSPACE | maskBN
	SS   = maskNEUTRAL | maskSPACE | maskSEPARATOR | maskSS
)

/* Return the minimum level of the direction, 0 for FRIBIDI_TYPE_LTR and
   1 for FRIBIDI_TYPE_RTL and FRIBIDI_TYPE_AL. */
func FRIBIDI_DIR_TO_LEVEL(dir ParType) Level {
	if dir.IsRtl() {
		return 1
	}
	return 0
}

/* Return the bidi type corresponding to the direction of the level number,
   FRIBIDI_TYPE_LTR for evens and FRIBIDI_TYPE_RTL for odds. */
func FRIBIDI_LEVEL_TO_DIR(lev Level) CharType {
	if lev.isRtl() != 0 {
		return RTL
	}
	return LTR
}

/* Override status of an explicit mark:
 * LRO,LRE->LTR, RLO,RLE->RTL, otherwise->ON. */
func FRIBIDI_EXPLICIT_TO_OVERRIDE_DIR(p CharType) CharType {
	if p.IsOverride() {
		FRIBIDI_LEVEL_TO_DIR(FRIBIDI_DIR_TO_LEVEL(p))
	}
	return ON
}

type CharType uint32

//  /*
//   * Define values for CharType
//   */

//  /* Strong types */

//  /* Left-To-Right letter */
//  #define FRIBIDI_TYPE_LTR_VAL	( maskSTRONG | maskLETTER )
//  /* Right-To-Left letter */
//  #define FRIBIDI_TYPE_RTL_VAL	( maskSTRONG | maskLETTER \
// 				 | maskRTL)
//  /* Arabic Letter */
//  #define FRIBIDI_TYPE_AL_VAL	( maskSTRONG | maskLETTER \
// 				 | maskRTL | maskARABIC )
//  /* Left-to-Right Embedding */
//  #define FRIBIDI_TYPE_LRE_VAL	( maskSTRONG | maskEXPLICIT)
//  /* Right-to-Left Embedding */
//  #define FRIBIDI_TYPE_RLE_VAL	( maskSTRONG | maskEXPLICIT \
// 				 | maskRTL )
//  /* Left-to-Right Override */
//  #define FRIBIDI_TYPE_LRO_VAL	( maskSTRONG | maskEXPLICIT \
// 				 | maskOVERRIDE )
//  /* Right-to-Left Override */
//  #define FRIBIDI_TYPE_RLO_VAL	( maskSTRONG | maskEXPLICIT \
// 				 | maskRTL | maskOVERRIDE )

//  /* Weak types */

//  /* Pop Directional Flag*/
//  #define FRIBIDI_TYPE_PDF_VAL	( maskWEAK | maskEXPLICIT )
//  /* European Numeral */
//  #define FRIBIDI_TYPE_EN_VAL	( maskWEAK | maskNUMBER )
//  /* Arabic Numeral */
//  #define FRIBIDI_TYPE_AN_VAL	( maskWEAK | maskNUMBER \
// 				 | maskARABIC )
//  /* European number Separator */
//  #define FRIBIDI_TYPE_ES_VAL	( maskWEAK | maskNUMSEPTER \
// 				 | maskES )
//  /* European number Terminator */
//  #define FRIBIDI_TYPE_ET_VAL	( maskWEAK | maskNUMSEPTER \
// 				 | maskET )
//  /* Common Separator */
//  #define FRIBIDI_TYPE_CS_VAL	( maskWEAK | maskNUMSEPTER \
// 				 | maskCS )
//  /* Non Spacing Mark */
//  #define FRIBIDI_TYPE_NSM_VAL	( maskWEAK | maskNSM )
//  /* Boundary Neutral */
//  #define FRIBIDI_TYPE_BN_VAL	( maskWEAK | maskSPACE \
// 				 | maskBN )

//  /* Neutral types */

//  /* Block Separator */
//  #define FRIBIDI_TYPE_BS_VAL	( maskNEUTRAL | maskSPACE \
// 				 | maskSEPARATOR | maskBS )
//  /* Segment Separator */
//  #define FRIBIDI_TYPE_SS_VAL	( maskNEUTRAL | maskSPACE \
// 				 | maskSEPARATOR | maskSS )
//  /* WhiteSpace */
//  #define FRIBIDI_TYPE_WS_VAL	( maskNEUTRAL | maskSPACE \
// 				 | maskWS )
//  /* Other Neutral */
//  #define FRIBIDI_TYPE_ON_VAL	( maskNEUTRAL )

//  /* The following are used in specifying paragraph direction only. */

//  /* Weak Left-To-Right */
//  #define FRIBIDI_TYPE_WLTR_VAL	( maskWEAK )
//  /* Weak Right-To-Left */
//  #define FRIBIDI_TYPE_WRTL_VAL	( maskWEAK | maskRTL )

/* start or end of text (run list) SENTINEL.  Only used internally */
//  #define FRIBIDI_TYPE_SENTINEL	( maskSENTINEL )

//  /* Private types for applications.  More private types can be obtained by
//   * summing up from this one. */
//  #define FRIBIDI_TYPE_PRIVATE	( maskPRIVATE )

//  /* New types in Unicode 6.3 */

//  /* Left-to-Right Isolate */
//  #define FRIBIDI_TYPE_LRI_VAL    ( maskNEUTRAL | maskISOLATE )
//  /* Right-to-Left Isolate */
//  #define FRIBIDI_TYPE_RLI_VAL    ( maskNEUTRAL | maskISOLATE | maskRTL )
//  /* First strong isolate */
//  #define FRIBIDI_TYPE_FSI_VAL    ( maskNEUTRAL | maskISOLATE | maskFIRST )

//  /* Pop Directional Isolate*/
//  #define FRIBIDI_TYPE_PDI_VAL	( maskNEUTRAL | maskWEAK | maskISOLATE )

//  /* Please don't use these two type names, use FRIBIDI_PAR_* form instead. */
//  #define FRIBIDI_TYPE_WLTR	FRIBIDI_PAR_WLTR
//  #define FRIBIDI_TYPE_WRTL	FRIBIDI_PAR_WRTL

//  /*
//   * Defining macros for needed queries, It is fully dependent on the
//   * implementation of CharType.
//   */

// IsRight checks is `p` is right -to-left level? */
//  #define FRIBIDI_LEVEL_IS_RTL(lev) ((lev) & 1)

//  /* Return the bidi type corresponding to the direction of the level number,
// 	FRIBIDI_TYPE_LTR for evens and FRIBIDI_TYPE_RTL for odds. */
//  #define FRIBIDI_LEVEL_TO_DIR(lev)	\
// 	 (FRIBIDI_LEVEL_IS_RTL (lev) ? FRIBIDI_TYPE_RTL : FRIBIDI_TYPE_LTR)

//  /* Return the minimum level of the direction, 0 for FRIBIDI_TYPE_LTR and
// 	1 for FRIBIDI_TYPE_RTL and FRIBIDI_TYPE_AL. */
//  #define FRIBIDI_DIR_TO_LEVEL(dir)	\
// 	 ((Level) (FRIBIDI_IS_RTL (dir) ? 1 : 0))

// IsStrong checks if `p` is string.
func (p CharType) IsStrong() bool { return p&maskSTRONG != 0 }

// IsRight checks is `p` is right to left: RTL, AL, RLE, RLO ?
func (p CharType) IsRtl() bool { return p&maskRTL != 0 }

// IsArabic checks is `p` is arabic : AL, AN ?
func (p CharType) IsArabic() bool { return p&maskARABIC != 0 }

// IsWeak checks is `p` is weak ?
func (p CharType) IsWeak() bool { return p&maskWEAK != 0 }

// IsNeutral checks is `p` is neutral ?
func (p CharType) IsNeutral() bool { return p&maskNEUTRAL != 0 }

// IsSentinel checks is `p` is sentinel ?
func (p CharType) IsSentinel() bool { return p&maskSENTINEL != 0 }

// IsLetter checks is `p` is letter : L, R, AL ?
func (p CharType) IsLetter() bool { return p&maskLETTER != 0 }

// IsNumber checks is `p` is number : EN, AN ?
func (p CharType) IsNumber() bool { return p&maskNUMBER != 0 }

// IsNumber checks is `p` is number  separator or terminator: ES, ET, CS ?
func (p CharType) IsNumberSeparatorOrTerminator() bool { return p&maskNUMSEPTER != 0 }

// IsSpace checks is `p` is space : BN, BS, SS, WS?
func (p CharType) IsSpace() bool { return p&maskSPACE != 0 }

// IsExplicit checks is `p` is explicit  mark: LRE, RLE, LRO, RLO, PDF ?
func (p CharType) IsExplicit() bool { return p&maskEXPLICIT != 0 }

// IsIsolator checks is `p` is isolator
func (p CharType) IsIsolate() bool { return p&maskISOLATE != 0 }

// IsText checks is `p` is text  separator: BS, SS ?
func (p CharType) IsSeparator() bool { return p&maskSEPARATOR != 0 }

// IsExplicit checks is `p` is explicit  override: LRO, RLO ?
func (p CharType) IsOverride() bool { return p&maskOVERRIDE != 0 }

// Some more:

// IsLeft checks is `p` is left  to right letter: LTR ?
func (p CharType) IsLtrLetter() bool { return p&(maskLETTER|maskRTL) == maskLETTER }

// IsRight checks is `p` is right  to left letter: RTL, AL ?
func (p CharType) IsRtlLetter() bool { return p&(maskLETTER|maskRTL) == (maskLETTER | maskRTL) }

// IsES checks is `p` is eS  or CS: ES, CS ?
func (p CharType) IsEsOrCs() bool { return p&(maskES|maskCS) != 0 }

// IsExplicit checks is `p` is explicit  or BN: LRE, RLE, LRO, RLO, PDF, BN ?
func (p CharType) IsExplicitOrBn() bool { return p&(maskEXPLICIT|maskBN) != 0 }

// IsExplicit checks is `p` is explicit  or BN or NSM: LRE, RLE, LRO, RLO, PDF, BN, NSM ?
func (p CharType) IsExplicitOrBnOrNsm() bool { return p&(maskEXPLICIT|maskBN|maskNSM) != 0 }

// IsExplicit checks is `p` is explicit  or BN or NSM: LRE, RLE, LRO, RLO, PDF, BN, NSM ?
func (p CharType) IsExplicitOrIsolateOrBnOrNsm() bool {
	return p&(maskEXPLICIT|maskISOLATE|maskBN|maskNSM) != 0
}

// IsExplicit checks is `p` is explicit  or BN or WS: LRE, RLE, LRO, RLO, PDF, BN, WS ?
func (p CharType) IsExplicitOrBnOrWs() bool { return p&(maskEXPLICIT|maskBN|maskWS) != 0 }

// IsExplicit checks is `p` is explicit  or separator or BN or WS: LRE, RLE, LRO, RLO, PDF, BS, SS, BN, WS ?
func (p CharType) IsExplicitOrSeparatorOrBnOrWs() bool {
	return p&(maskEXPLICIT|maskSEPARATOR|maskBN|maskWS) != 0
}

// IsPrivate checks is `p` is a private-use type for application
func (p CharType) IsPrivate() bool { return p&maskPRIVATE != 0 }

//  /* Define some conversions. */

/* Change numbers to RTL: EN,AN -> RTL. */
func (p CharType) FRIBIDI_CHANGE_NUMBER_TO_RTL() CharType {
	if p.IsNumber() {
		return RTL
	}
	return p
}

//  /* Override status of an explicit mark:
//   * LRO,LRE->LTR, RLO,RLE->RTL, otherwise->ON. */
//  #define FRIBIDI_EXPLICIT_TO_OVERRIDE_DIR(p) \
// 	 (FRIBIDI_IS_OVERRIDE(p) ? FRIBIDI_LEVEL_TO_DIR(FRIBIDI_DIR_TO_LEVEL(p)) \
// 				 : FRIBIDI_TYPE_ON)

//  /* Weaken type for paragraph fallback purposes:
//   * LTR->WLTR, RTL->WRTL. */
//  #define FRIBIDI_WEAK_PARAGRAPH(p) (FRIBIDI_PAR_WLTR | p & maskRTL))

// convert from golang enums to frididi types
func newCharType(class bidi.Class) CharType {
	switch class {
	case bidi.L: // LeftToRight
		return LTR
	case bidi.R: // RightToLeft
		return RTL
	case bidi.EN: // EuropeanNumber
		return EN
	case bidi.ES: // EuropeanSeparator
		return ES
	case bidi.ET: // EuropeanTerminator
		return ET
	case bidi.AN: // ArabicNumber
		return AN
	case bidi.CS: // CommonSeparator
		return CS
	case bidi.B: // ParagraphSeparator
		return BS
	case bidi.S: // SegmentSeparator
		return SS
	case bidi.WS: // WhiteSpace
		return WS
	case bidi.ON: // OtherNeutral
		return ON
	case bidi.BN: // BoundaryNeutral
		return BN
	case bidi.NSM: // NonspacingMark
		return NSM
	case bidi.AL: // ArabicLetter
		return AL
	case bidi.LRO: // LeftToRightOverride
		return LRO
	case bidi.RLO: // RightToLeftOverride
		return RLO
	case bidi.LRE: // LeftToRightEmbedding
		return LRE
	case bidi.RLE: // RightToLeftEmbedding
		return RLE
	case bidi.PDF: // PopDirectionalFormat
		return PDF
	case bidi.LRI: // LeftToRightIsolate
		return LRI
	case bidi.RLI: // RightToLeftIsolate
		return RLI
	case bidi.FSI: // FirstStrongIsolate
		return FSI
	case bidi.PDI: // PopDirectionalIsolate
		return PDI
	default:
		return LTR
	}
}

// GetBidiType returns the bidi type of a character as defined in Table 3.7
// Bidirectional Character Types of the Unicode Bidirectional Algorithm
// available at http://www.unicode.org/reports/tr9/#Bidirectional_Character_Types, using
// data provided by golang.org/x/text/unicode/bidi
func GetBidiType(ch rune) CharType {
	props, _ := bidi.LookupRune(ch)
	return newCharType(props.Class())
}

func getBidiTypes(str []rune) []CharType {
	out := make([]CharType, len(str))
	for i, r := range str {
		out[i] = GetBidiType(r)
	}
	return out
}

// Define option flags that various functions use.
const (
	FRIBIDI_FLAG_SHAPE_MIRRORING = 1
	FRIBIDI_FLAG_REORDER_NSM     = 1 << 1

	FRIBIDI_FLAG_SHAPE_ARAB_PRES    = 1 << 8
	FRIBIDI_FLAG_SHAPE_ARAB_LIGA    = 1 << 9
	FRIBIDI_FLAG_SHAPE_ARAB_CONSOLE = 1 << 10

	FRIBIDI_FLAG_REMOVE_BIDI     = 1 << 16
	FRIBIDI_FLAG_REMOVE_JOINING  = 1 << 17
	FRIBIDI_FLAG_REMOVE_SPECIALS = 1 << 18

	/*
	 * And their combinations.
	 */
	FRIBIDI_FLAGS_DEFAULT = FRIBIDI_FLAG_SHAPE_MIRRORING | FRIBIDI_FLAG_REORDER_NSM | FRIBIDI_FLAG_REMOVE_SPECIALS
	FRIBIDI_FLAGS_ARABIC  = FRIBIDI_FLAG_SHAPE_ARAB_PRES | FRIBIDI_FLAG_SHAPE_ARAB_LIGA
)

var flags = FRIBIDI_FLAGS_DEFAULT | FRIBIDI_FLAGS_ARABIC

// Visual is the visual output as specified by the Unicode Bidirectional Algorithm
type Visual struct {
	Str             []rune  // visual string
	VisualToLogical []int   // mapping from visual string back to the logical string indexes
	EmbeddingLevels []Level // list of embedding levels
}

// LogicalToVisual reverts `VisualToLogical`,
// return the mapping from logical to visual string indexes
func (v Visual) LogicalToVisual() []int {
	out := make([]int, len(v.VisualToLogical))
	for i, vToL := range v.VisualToLogical {
		out[vToL] = i
	}
	return out
}

// fribidi_log2vis converts the logical input string to the visual output
// strings as specified by the Unicode Bidirectional Algorithm.  As a side
// effect it also generates mapping lists between the two strings, and the
// list of embedding levels as defined by the algorithm.
//
// Note that this function handles one-line paragraphs. For multi-
// paragraph texts it is necessary to first split the text into
// separate paragraphs and then carry over the resolved paragraphBaseDir
// between the subsequent invocations.
//
// The maximum level found plus one is also returned.
func fribidi_log2vis(str []rune, paragraphBaseDir *ParType /* requested and resolved paragraph base direction */) (Visual, Level) {
	bidiTypes := getBidiTypes(str)

	bracketTypes := getBracketTypes(str, bidiTypes)

	embeddingLevels, maxLevel := fribidi_get_par_embedding_levels_ex(bidiTypes, bracketTypes, paragraphBaseDir)

	/* Set up the ordering array to identity order */
	positionsVToL := make([]int, len(str))
	for i := range positionsVToL {
		positionsVToL[i] = i
	}

	visualStr := append([]rune{}, str...) // copy

	/* Arabic joining */
	arProps := getJoiningTypes(str, bidiTypes)
	fribidi_join_arabic(bidiTypes, embeddingLevels, arProps)

	fribidi_shape(flags, embeddingLevels, arProps, visualStr)

	/* line breaking goes here, but we assume one line in this function */

	/* and this should be called once per line, but again, we assume one
	 * line in this deprecated function */
	fribidi_reorder_line(flags, bidiTypes, len(str), 0, *paragraphBaseDir,
		embeddingLevels, visualStr, positionsVToL)

	return Visual{Str: visualStr, VisualToLogical: positionsVToL, EmbeddingLevels: embeddingLevels}, maxLevel
}

// fribidi_remove_bidi_marks removes the bidi and boundary-neutral marks out of an string
// and the accompanying lists.  It implements rule X9 of the Unicode
// Bidirectional Algorithm available at
// http://www.unicode.org/reports/tr9/#X9, with the exception that it removes
// U+200E LEFT-TO-RIGHT MARK and U+200F RIGHT-TO-LEFT MARK too.
//
// If any of the input lists are NULL, the list is skipped.  If str is the
// visual string, then positions_to_this is positions_L_to_V and
// position_from_this_list is positions_V_to_L;  if str is the logical
// string, the other way. Moreover, the position maps should be filled with
// valid entries.
//
// A position map pointing to a removed character is filled with \(mi1. By the
// way, you should not use embedding_levels if str is visual string.
//
// For best results this function should be run on a whole paragraph, not
// lines; but feel free to do otherwise if you know what you are doing.
//
// The input slice is mutated and resliced to its new length, then returned
func fribidi_remove_bidi_marks(str []rune, positions_to_this, position_from_this_list []int, embedding_levels []Level) []rune {
	var j int

	/* If to_this is not NULL, we must have from_this as well. If it is
	   not given by the caller, we have to make a private instance of it. */
	if len(positions_to_this) != 0 && len(position_from_this_list) == 0 {
		position_from_this_list = make([]int, len(str))
		for i, to := range positions_to_this {
			position_from_this_list[to] = i
		}
	}

	for i, r := range str {
		if bType := GetBidiType(r); !bType.IsExplicitOrBn() && !bType.IsIsolate() &&
			r != FRIBIDI_CHAR_LRM && r != FRIBIDI_CHAR_RLM {
			str[j] = r
			embedding_levels[j] = embedding_levels[i]
			if len(position_from_this_list) != 0 {
				position_from_this_list[j] = position_from_this_list[i]
			}
			j++
		}
	}

	/* Convert the from_this list to to_this */
	if len(positions_to_this) != 0 {
		for i, from := range position_from_this_list {
			positions_to_this[from] = i
		}
	}

	return str[0:j]
}
