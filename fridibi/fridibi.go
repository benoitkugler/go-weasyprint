// Golang port of fridibi, using golang.org/x/text/unicode/bidi
// as data source.
package fridibi

import "golang.org/x/text/unicode/bidi"

type fridibiClass uint8

const (
	ltr fridibiClass = iota
	rtl
	al
	en
	an
	es
	et
	nsm
	cs
	bn
	bs
	ss
	ws
	on
	lre
	rle
	lro
	rlo
	pdf
	lri
	rli
	fsi
	pdi

	numClass // safety guard for classToType

	l = ltr
	r = rtl
	b = bs
	s = ss
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

type CharType uint32

var classToType = [numClass]CharType{
	(maskSTRONG | maskLETTER),
	(maskSTRONG | maskLETTER | maskRTL),
	(maskSTRONG | maskLETTER | maskRTL | maskARABIC),
	(maskWEAK | maskNUMBER),
	(maskWEAK | maskNUMBER | maskARABIC),
	(maskWEAK | maskNUMSEPTER | maskES),
	(maskWEAK | maskNUMSEPTER | maskET),
	(maskWEAK | maskNUMSEPTER | maskCS),
	(maskWEAK | maskNSM),
	(maskWEAK | maskSPACE | maskBN),
	(maskNEUTRAL | maskSPACE | maskSEPARATOR | maskBS),
	(maskNEUTRAL | maskSPACE | maskSEPARATOR | maskSS),
	(maskNEUTRAL | maskSPACE | maskWS),
	(maskNEUTRAL),
	(maskSTRONG | maskEXPLICIT),
	(maskSTRONG | maskEXPLICIT | maskRTL),
	(maskSTRONG | maskEXPLICIT | maskOVERRIDE),
	(maskSTRONG | maskEXPLICIT | maskRTL | maskOVERRIDE),
	(maskWEAK | maskEXPLICIT),
	(maskNEUTRAL | maskISOLATE),
	(maskNEUTRAL | maskISOLATE | maskRTL),
	(maskNEUTRAL | maskISOLATE | maskFIRST),
	(maskNEUTRAL | maskWEAK | maskISOLATE),
}

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

//  /* start or end of text (run list) SENTINEL.  Only used internally */
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

//  /* Change numbers to RTL: EN,AN -> RTL. */
//  #define FRIBIDI_CHANGE_NUMBER_TO_RTL(p) \
// 	 (FRIBIDI_IS_NUMBER(p) ? FRIBIDI_TYPE_RTL : (p))

//  /* Override status of an explicit mark:
//   * LRO,LRE->LTR, RLO,RLE->RTL, otherwise->ON. */
//  #define FRIBIDI_EXPLICIT_TO_OVERRIDE_DIR(p) \
// 	 (FRIBIDI_IS_OVERRIDE(p) ? FRIBIDI_LEVEL_TO_DIR(FRIBIDI_DIR_TO_LEVEL(p)) \
// 				 : FRIBIDI_TYPE_ON)

//  /* Weaken type for paragraph fallback purposes:
//   * LTR->WLTR, RTL->WRTL. */
//  #define FRIBIDI_WEAK_PARAGRAPH(p) (FRIBIDI_PAR_WLTR | p & maskRTL))

// convert from golang enums to frididi class
func newFribidiClass(class bidi.Class) fridibiClass {
	switch class {
	case bidi.L: // LeftToRight
		return l
	case bidi.R: // RightToLeft
		return r
	case bidi.EN: // EuropeanNumber
		return en
	case bidi.ES: // EuropeanSeparator
		return es
	case bidi.ET: // EuropeanTerminator
		return et
	case bidi.AN: // ArabicNumber
		return an
	case bidi.CS: // CommonSeparator
		return cs
	case bidi.B: // ParagraphSeparator
		return b
	case bidi.S: // SegmentSeparator
		return s
	case bidi.WS: // WhiteSpace
		return ws
	case bidi.ON: // OtherNeutral
		return on
	case bidi.BN: // BoundaryNeutral
		return bn
	case bidi.NSM: // NonspacingMark
		return nsm
	case bidi.AL: // ArabicLetter
		return al
	case bidi.LRO: // LeftToRightOverride
		return lro
	case bidi.RLO: // RightToLeftOverride
		return rlo
	case bidi.LRE: // LeftToRightEmbedding
		return lre
	case bidi.RLE: // RightToLeftEmbedding
		return rle
	case bidi.PDF: // PopDirectionalFormat
		return pdf
	case bidi.LRI: // LeftToRightIsolate
		return lri
	case bidi.RLI: // RightToLeftIsolate
		return rli
	case bidi.FSI: // FirstStrongIsolate
		return fsi
	case bidi.PDI: // PopDirectionalIsolate
		return pdi
	default:
		return ltr
	}
}

// GetBidiType returns the bidi type of a character as defined in Table 3.7
// Bidirectional Character Types of the Unicode Bidirectional Algorithm
// available at
// http://www.unicode.org/reports/tr9/#Bidirectional_Character_Types, using
// data provided by golang.org/x/text/unicode/bidi
func GetBidiType(ch rune) CharType {
	props, _ := bidi.LookupRune(ch)
	return classToType[newFribidiClass(props.Class())]
}
