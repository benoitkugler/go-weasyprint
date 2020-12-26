package fribidi

import (
	"golang.org/x/text/unicode/bidi"
)

// BracketType is a rune value with its MSB is used to indicate an opening bracket
type BracketType uint32

const NoBracket BracketType = 0

func (bt BracketType) isOpen() bool {
	return bt&FRIBIDI_BRACKET_OPEN_MASK > 0
}

func (bt BracketType) id() BracketType {
	return bt & FRIBIDI_BRACKET_ID_MASK
}

const (
	FRIBIDI_BRACKET_OPEN_MASK BracketType = 1 << 31
	FRIBIDI_BRACKET_ID_MASK               = ^FRIBIDI_BRACKET_OPEN_MASK
)

// GetBracket finds the bracketed equivalent of a character as defined in
// the file BidiBrackets.txt of the Unicode Character Database available at
// http://www.unicode.org/Public/UNIDATA/BidiBrackets.txt.
//
// If the input character is declared as a brackets character in the
// Unicode standard and has a bracketed equivalent, the matching bracketed
// character is returned, with its high bit set.
// Otherwise `NoBracket` (zero) is returned.
func GetBracket(ch rune) BracketType {
	props, _ := bidi.LookupRune(ch)
	if !props.IsBracket() {
		return NoBracket
	}
	pair := BracketType(bracketsTable[ch])
	pair &= FRIBIDI_BRACKET_ID_MASK
	if props.IsOpeningBracket() {
		pair |= FRIBIDI_BRACKET_OPEN_MASK
	}
	return pair
}

// getBracketTypes finds the bracketed characters of an string of characters.
// `bidiTypes` is not needed strictly speaking, but is used as an optimization.
// see GetBracket for details.
func getBracketTypes(str []rune, bidiTypes []CharType) []BracketType {
	out := make([]BracketType, len(str))
	for i, r := range str {
		/* Optimization that bracket must be of types ON */
		if bidiTypes[i] == ON {
			out[i] = GetBracket(r)
		}
		// else -> zero
	}
	return out
}
