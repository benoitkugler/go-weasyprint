package parser


// Parse `<An+B> <http://dev.w3.org/csswg/css-syntax-3/#anb>`_,
//     as found in `:nth-child()
//     <http://dev.w3.org/csswg/selectors/#nth-child-pseudo>`
//     and related Selector pseudo-classes.
//     Although tinycss2 does not include a full Selector parser,
//     this bit of syntax is included as it is particularly tricky to define
//     on top of a CSS tokenizer.
//     Returns  ``(a, b)`` slice of integers or nil
func parseNth(input []Token) []int {
    tokens := NewTokenIterator(input)
    token_ = nextSignificant(tokens)
    if token_ == nil {
        return nil
	} 
	switch token := token_.(type) {
	case NumberToken: 
    if  token.Integer {
		return parseEnd(tokens, 0, token.intValue)
	}
case DimensionToken:
	if token.isInteger {
        unit = token.lowerUnit
        if unit == "n" {
            return parseB(tokens, token.intValue)
        } else if unit == "n-" {
            return parseSignlessB(tokens, token.intValue, -1)
        } else {
            match = NDASHDIGITSRE.match(unit)
            if match {
                return parseEnd(tokens, token.intValue, int(match.group(1)))
            }
		}
	}
    case IdentToken:
        ident = token.lowerValue
        if ident == "even" {
            return parseEnd(tokens, 2, 0)
        } else if ident == "odd" {
            return parseEnd(tokens, 2, 1)
        } else if ident == "n" {
            return parseB(tokens, 1)
        } else if ident == "-n" {
            return parseB(tokens, -1)
        } else if ident == "n-" {
            return parseSignlessB(tokens, 1, -1)
        } else if ident == "-n-" {
            return parseSignlessB(tokens, -1, -1)
        } else if ident[0] == "-" {
            match = NDASHDIGITSRE.match(ident[1:])
            if match {
                return parseEnd(tokens, -1, int(match.group(1)))
            }
        } else {
            match = NDASHDIGITSRE.match(ident)
            if match {
                return parseEnd(tokens, 1, int(match.group(1)))
            }
        }
	case LiteralToken:
		if token == "+" {
        token = next(tokens)  // Whitespace after an initial "+" is invalid.
        if ident, ok := token.(IdentToken); ok {
            ident = token.lowerValue
            if ident == "n" {
                return parseB(tokens, 1)
            } else if ident == "n-" {
                return parseSignlessB(tokens, 1, -1)
            } else {
                match = NDASHDIGITSRE.match(ident)
                if match {
                    return parseEnd(tokens, 1, int(match.group(1)))
                }
            }
        }
	}
}
} 

// func parseB(tokens, a) {
//     token = NextSignificant(tokens)
//     if token is None {
//         return (a, 0)
//     } else if token == "+" {
//         return parseSignlessB(tokens, a, 1)
//     } else if token == "-" {
//         return parseSignlessB(tokens, a, -1)
//     } else if (token.type == "number" && token.isInteger and
//           token.representation[0] := range "-+") {
//           }
//         return parseEnd(tokens, a, token.intValue)
// } 

// func parseSignlessB(tokens, a, bSign) {
//     token = NextSignificant(tokens)
//     if (token.type == "number" && token.isInteger and
//             not token.representation[0] := range "-+") {
//             }
//         return parseEnd(tokens, a, bSign * token.intValue)
// } 

// func parseEnd(tokens, a, b) {
//     if NextSignificant(tokens) is None {
//         return (a, b)
//     }
// } 

NDASHDIGITSRE = re.compile("^n(-[0-9]+)$")
