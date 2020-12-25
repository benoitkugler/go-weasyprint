package fribidi

import "golang.org/x/text/unicode/bidi"

// BracketType is a rune value with its MSB is used to indicate an opening bracket
type BracketType rune

const NoBracket BracketType = 0

func (bt BracketType) isOpen() bool {
	return bt&FRIBIDI_BRACKET_OPEN_MASK > 0
}

func (bt BracketType) id() BracketType {
	return bt & FRIBIDI_BRACKET_ID_MASK
}

var brackets = map[rune]rune{
	'\u0029': '\u0028',
	'\u0028': '\u0029',
	'\u005B': '\u005D',
	'\u005D': '\u005B',
	'\u007B': '\u007D',
	'\u007D': '\u007B',
	'\u0F3A': '\u0F3B',
	'\u0F3B': '\u0F3A',
	'\u0F3C': '\u0F3D',
	'\u0F3D': '\u0F3C',
	'\u169B': '\u169C',
	'\u169C': '\u169B',
	'\u2045': '\u2046',
	'\u2046': '\u2045',
	'\u207D': '\u207E',
	'\u207E': '\u207D',
	'\u208D': '\u208E',
	'\u208E': '\u208D',
	'\u2308': '\u2309',
	'\u2309': '\u2308',
	'\u230A': '\u230B',
	'\u230B': '\u230A',
	'\u2329': '\u232A',
	'\u232A': '\u2329',
	'\u2768': '\u2769',
	'\u2769': '\u2768',
	'\u276A': '\u276B',
	'\u276B': '\u276A',
	'\u276C': '\u276D',
	'\u276D': '\u276C',
	'\u276E': '\u276F',
	'\u276F': '\u276E',
	'\u2770': '\u2771',
	'\u2771': '\u2770',
	'\u2772': '\u2773',
	'\u2773': '\u2772',
	'\u2774': '\u2775',
	'\u2775': '\u2774',
	'\u27C5': '\u27C6',
	'\u27C6': '\u27C5',
	'\u27E6': '\u27E7',
	'\u27E7': '\u27E6',
	'\u27E8': '\u27E9',
	'\u27E9': '\u27E8',
	'\u27EA': '\u27EB',
	'\u27EB': '\u27EA',
	'\u27EC': '\u27ED',
	'\u27ED': '\u27EC',
	'\u27EE': '\u27EF',
	'\u27EF': '\u27EE',
	'\u2983': '\u2984',
	'\u2984': '\u2983',
	'\u2985': '\u2986',
	'\u2986': '\u2985',
	'\u2987': '\u2988',
	'\u2988': '\u2987',
	'\u2989': '\u298A',
	'\u298A': '\u2989',
	'\u298B': '\u298C',
	'\u298C': '\u298B',
	'\u298D': '\u2990',
	'\u298E': '\u298F',
	'\u298F': '\u298E',
	'\u2990': '\u298D',
	'\u2991': '\u2992',
	'\u2992': '\u2991',
	'\u2993': '\u2994',
	'\u2994': '\u2993',
	'\u2995': '\u2996',
	'\u2996': '\u2995',
	'\u2997': '\u2998',
	'\u2998': '\u2997',
	'\u29D8': '\u29D9',
	'\u29D9': '\u29D8',
	'\u29DA': '\u29DB',
	'\u29DB': '\u29DA',
	'\u29FC': '\u29FD',
	'\u29FD': '\u29FC',
	'\u2E22': '\u2E23',
	'\u2E23': '\u2E22',
	'\u2E24': '\u2E25',
	'\u2E25': '\u2E24',
	'\u2E26': '\u2E27',
	'\u2E27': '\u2E26',
	'\u2E28': '\u2E29',
	'\u2E29': '\u2E28',
	'\u3008': '\u3009',
	'\u3009': '\u3008',
	'\u300A': '\u300B',
	'\u300B': '\u300A',
	'\u300C': '\u300D',
	'\u300D': '\u300C',
	'\u300E': '\u300F',
	'\u300F': '\u300E',
	'\u3010': '\u3011',
	'\u3011': '\u3010',
	'\u3014': '\u3015',
	'\u3015': '\u3014',
	'\u3016': '\u3017',
	'\u3017': '\u3016',
	'\u3018': '\u3019',
	'\u3019': '\u3018',
	'\u301A': '\u301B',
	'\u301B': '\u301A',
	'\uFE59': '\uFE5A',
	'\uFE5A': '\uFE59',
	'\uFE5B': '\uFE5C',
	'\uFE5C': '\uFE5B',
	'\uFE5D': '\uFE5E',
	'\uFE5E': '\uFE5D',
	'\uFF08': '\uFF09',
	'\uFF09': '\uFF08',
	'\uFF3B': '\uFF3D',
	'\uFF3D': '\uFF3B',
	'\uFF5B': '\uFF5D',
	'\uFF5D': '\uFF5B',
	'\uFF5F': '\uFF60',
	'\uFF60': '\uFF5F',
	'\uFF62': '\uFF63',
	'\uFF63': '\uFF62',
}

const (
	FRIBIDI_BRACKET_OPEN_MASK BracketType = 1<<31 - 1
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
	pair := BracketType(brackets[ch])
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
