package parser

import (
	"errors"
	"log"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/benoitkugler/go-weasyprint/utils"
	// "github.com/gorilla/css/scanner"
)

// This file builds the ast from the Gorilla tokenizer.

var (
	numberRe    = regexp.MustCompile(`[-+]?([0-9]*\.)?[0-9]+([eE][+-]?[0-9]+)?`)
	hexEscapeRe = regexp.MustCompile(`([0-9A-Fa-f]{1,6})[ \n\t]?`)

	// litteralTokens = map[int]string{
	// 	int(scanner.TokenCDC):            "-->",
	// 	int(scanner.TokenCDO):            "<!--",
	// 	int(scanner.TokenIncludes):       "~=",
	// 	int(scanner.TokenDashMatch):      "|=",
	// 	int(scanner.TokenPrefixMatch):    "^=",
	// 	int(scanner.TokenSuffixMatch):    "$=",
	// 	int(scanner.TokenSubstringMatch): "*=",
	// }
)

type nestedBlock struct {
	tokens  *[]Token
	endChar string
}

// func ParseComponentValueList(css string, skipComments bool) []Token {
// 	// This turns out to be faster than a regexp:
// 	css = strings.ReplaceAll(css, "\r\n", "\n")
// 	css = strings.ReplaceAll(css, "\r", "\n")
// 	css = strings.ReplaceAll(css, "\f", "\n")

// 	scan := scanner.New(css)
// 	var out []Token    // possibly nested tokens
// 	ts := &out         // current stack of tokens
// 	var endChar string // Pop the stack when encountering this character.
// 	var stack []nestedBlock
// 	for {
// 		token := scan.Next()
// 		// fmt.Println(token)
// 		switch token.Type {
// 		case scanner.TokenEOF, scanner.TokenError:
// 			return out
// 		case scanner.TokenS:
// 			*ts = append(*ts, WhitespaceToken{Value: token.Value})
// 		case scanner.TokenUnicodeRange:
// 			start, end, err := consumeUnicodeRange(token.Value, pos+2)
// 			if err != nil {
// 				*ts = append(*ts, ParseError{Kind: "invalid number", Message: err.Error()})
// 			} else {
// 				*ts = append(*ts, UnicodeRangeToken{Start: uint32(start), End: uint32(end)})
// 			}
// 		case scanner.TokenIdent:
// 			*ts = append(*ts, IdentToken{Value: LowerableString(consumeIdent(token.Value))})
// 		case scanner.TokenURI:
// 			url, err := consumeUrl(token.Value)
// 			if err != nil {
// 				*ts = append(*ts, ParseError{Kind: "bad-url", Message: err.Error()})
// 			} else {
// 				*ts = append(*ts, URLToken{Value: url})
// 			}
// 		case scanner.TokenFunction:
// 			funcBlock := FunctionBlock{
// 				Name:      LowerableString(strings.TrimSuffix(token.Value, "(")),
// 				Arguments: new([]Token),
// 			}
// 			*ts = append(*ts, funcBlock)
// 			stack = append(stack, nestedBlock{tokens: ts, endChar: endChar})
// 			endChar = ")"
// 			ts = funcBlock.Arguments
// 		case scanner.TokenDimension:
// 			num, unit := consumeNumber(token.Value)
// 			*ts = append(*ts, DimensionToken{Unit: LowerableString(unit), numericToken: num})
// 		case scanner.TokenPercentage:
// 			num, _ := consumeNumber(token.Value)
// 			*ts = append(*ts, PercentageToken(num))
// 		case scanner.TokenNumber:
// 			num, _ := consumeNumber(token.Value)
// 			*ts = append(*ts, NumberToken(num))
// 		case scanner.TokenAtKeyword:
// 			*ts = append(*ts, AtKeywordToken{
// 				Value: LowerableString(strings.TrimPrefix(token.Value, "@")),
// 			})
// 		case scanner.TokenCDC, scanner.TokenCDO, scanner.TokenIncludes, scanner.TokenDashMatch, scanner.TokenPrefixMatch, scanner.TokenSuffixMatch, scanner.TokenSubstringMatch:
// 			*ts = append(*ts, LiteralToken{Value: litteralTokens[int(token.Type)]})
// 		case scanner.TokenHash:
// 			isIdentifier := isIdentStart([]rune(token.Value), 1)
// 			*ts = append(*ts, HashToken{Value: token.Value[1:], IsIdentifier: isIdentifier})
// 		case scanner.TokenString:
// 			str, err := consumeQuotedString(token.Value)
// 			if err != nil {
// 				*ts = append(*ts, ParseError{Kind: "bad-string", Message: "bad string token"})
// 			} else {
// 				*ts = append(*ts, StringToken{Value: str})
// 			}
// 		case scanner.TokenComment:
// 			cmt := strings.TrimSuffix(strings.TrimPrefix(token.Value, "/*"), "*/")
// 			if !skipComments {
// 				*ts = append(*ts, Comment{Value: cmt})
// 			}
// 		case scanner.TokenChar:
// 			switch token.Value {
// 			case "{":
// 				brack := CurlyBracketsBlock{Content: new([]Token)}
// 				*ts = append(*ts, brack)
// 				stack = append(stack, nestedBlock{tokens: ts, endChar: endChar})
// 				endChar = "}"
// 				ts = brack.Content
// 			case "[":
// 				brack := SquareBracketsBlock{Content: new([]Token)}
// 				*ts = append(*ts, brack)
// 				stack = append(stack, nestedBlock{tokens: ts, endChar: endChar})
// 				endChar = "]"
// 				ts = brack.Content
// 			case "(":
// 				brack := ParenthesesBlock{Content: new([]Token)}
// 				*ts = append(*ts, brack)
// 				stack = append(stack, nestedBlock{tokens: ts, endChar: endChar})
// 				endChar = ")"
// 				ts = brack.Content
// 			case endChar: // Matching }, ] or )
// 				// The top-level endChar is "" (never equal to a character),
// 				// so we never get here if the stack is empty.
// 				var block nestedBlock
// 				block, stack = stack[len(stack)-1], stack[:len(stack)-1]
// 				ts, endChar = block.tokens, block.endChar
// 			case "}", "]", ")":
// 				*ts = append(*ts, ParseError{Kind: token.Value, Message: "Unmatched " + token.Value})
// 			default:
// 				*ts = append(*ts, LiteralToken{Value: token.Value})
// 			}
// 		default:
// 			log.Fatalf("unssupported token : %s", token)
// 		}
// 	}
// }

// Parse a list of component values.
// If `skipComments` is true, ignore CSS comments :
// the return values (and recursively its blocks and functions)
// will not contain any `Comment` object.
func ParseComponentValueList(css string, skipComments bool) []Token {
	// This turns out to be faster than a regexp:
	css = strings.ReplaceAll(css, "\r\n", "\n")
	css = strings.ReplaceAll(css, "\r", "\n")
	css = strings.ReplaceAll(css, "\f", "\n")

	length := len(css)
	tokenStartPos, pos := 0, 0
	line, lastNewline := 1, -1
	var out []Token    // possibly nested tokens
	ts := &out         // current stack of tokens
	var endChar string // Pop the stack when encountering this character.
	var stack []nestedBlock
	var err error
	for pos < length {

		newline := strings.LastIndex(css[tokenStartPos:pos], "\n")
		if newline != -1 {
			newline += tokenStartPos
			line += 1 + strings.Count(css[tokenStartPos:newline], "\n")
			lastNewline = newline
		}
		// First character in a line is in column 1.
		column := pos - lastNewline
		tokenPos := newTk(line, column)

		tokenStartPos = pos
		c := string(css[pos])

		switch c {
		case " ", "\n", "\t":
			pos += 1
			for p := pos; p < length; p += 1 {
				u := css[pos]
				if !(u == ' ' || u == '\n' || u == '\t') {
					break
				}
			}
			value := css[tokenStartPos:pos]
			*ts = append(*ts, WhitespaceToken{tk: tokenPos, Value: value})
			continue
		case "U", "u":
			if pos+2 < length && css[pos+1] == '+' && strings.ContainsRune("0123456789abcdefABCDEF?", rune(css[pos+2])) {
				var start, end int64
				start, end, pos, err = consumeUnicodeRange(css, pos)
				if err != nil {
					*ts = append(*ts, ParseError{tk: tokenPos, Kind: "invalid number", Message: err.Error()})
				} else {
					*ts = append(*ts, UnicodeRangeToken{tk: tokenPos, Start: uint32(start), End: uint32(end)})
				}
				continue
			}
		}

		if strings.HasPrefix(css[pos:], "-->") { // Check before identifiers
			*ts = append(*ts, LiteralToken{tk: tokenPos, Value: "-->"})
			pos += 3
			continue
		} else if isIdentStart(css, pos) {
			var value string
			value, pos = consumeIdent(css, pos)
			if css[pos] != '(' { // Not a function
				*ts = append(*ts, IdentToken{tk: tokenPos, Value: LowerableString(value)})
				continue
			}
			pos += 1 // Skip the "("
			if utils.AsciiLower(value) == "url" {
				value, pos, err = consumeUrl(css, pos)
				if err != nil {
					*ts = append(*ts, ParseError{tk: tokenPos, Kind: "bad-url", Message: err.Error()})
				} else {
					*ts = append(*ts, URLToken{tk: tokenPos, Value: value})
				}
				continue
			}
			funcBlock := FunctionBlock{
				tk:        tokenPos,
				Name:      LowerableString(strings.TrimSuffix(value, "(")),
				Arguments: new([]Token),
			}
			*ts = append(*ts, funcBlock)
			stack = append(stack, nestedBlock{tokens: ts, endChar: endChar})
			endChar = ")"
			ts = funcBlock.Arguments
			continue
		}

		value := css[pos:]
		match := numberRe.FindStringIndex(value)
		if match != nil {
			repr := css[pos+match[0] : pos+match[1]]
			pos += match[1]
			value, err := strconv.ParseFloat(repr, 32)
			if err != nil {
				log.Fatalf("should be a number %s \n", err)
			}
			_, err = strconv.ParseInt(repr, 10, 0)
			isInt := err != nil
			n := numericToken{
				tk:             tokenPos,
				Representation: repr,
				IsInteger:      isInt,
				Value:          float32(value),
			}
			if pos < length && isIdentStart(css, pos) {
				var unit string
				unit, pos = consumeIdent(css, pos)
				*ts = append(*ts, DimensionToken{numericToken: n, Unit: LowerableString(unit)})
			} else if css[pos] == '%' {
				pos += 1
				*ts = append(*ts, PercentageToken(n))
			} else {
				*ts = append(*ts, NumberToken(n))
			}
		}
		switch c {
		case "@":
			pos += 1
			if pos < length && isIdentStart(css, pos) {
				value, pos = consumeIdent(css, pos)
				*ts = append(*ts, AtKeywordToken{tk: tokenPos, Value: LowerableString(value)})
			} else {
				*ts = append(*ts, LiteralToken{tk: tokenPos, Value: "@"})
			}
		case "#":
			pos += 1
			if pos < length {
				r, _ := utf8.DecodeRuneInString(css[pos:])
				if strings.ContainsRune("0123456789abcdefghijklmnopqrstuvwxyz-_ABCDEFGHIJKLMNOPQRSTUVWXYZ", r) ||
					r > 0x7F || // Non-ASCII
					r == '\\' && !strings.HasPrefix(css[pos:], "\\\n") { // Valid escape
					isIdentifier := isIdentStart(css, pos)
					value, pos = consumeIdent(css, pos)
					*ts = append(*ts, HashToken{tk: tokenPos, Value: value, IsIdentifier: isIdentifier})
				}
			} else {
				*ts = append(*ts, LiteralToken{tk: tokenPos, Value: "#"})
			}
		case "{":
			brack := CurlyBracketsBlock{tk: tokenPos, Content: new([]Token)}
			*ts = append(*ts, brack)
			stack = append(stack, nestedBlock{tokens: ts, endChar: endChar})
			endChar = "}"
			ts = brack.Content
			pos += 1
		case "[":
			brack := SquareBracketsBlock{tk: tokenPos, Content: new([]Token)}
			*ts = append(*ts, brack)
			stack = append(stack, nestedBlock{tokens: ts, endChar: endChar})
			endChar = "]"
			ts = brack.Content
			pos += 1
		case "(":
			brack := ParenthesesBlock{tk: tokenPos, Content: new([]Token)}
			*ts = append(*ts, brack)
			stack = append(stack, nestedBlock{tokens: ts, endChar: endChar})
			endChar = ")"
			ts = brack.Content
			pos += 1
		case endChar: // Matching }, ] or )
			// The top-level endChar is "" (never equal to a character),
			// so we never get here if the stack is empty.
			var block nestedBlock
			block, stack = stack[len(stack)-1], stack[:len(stack)-1]
			ts, endChar = block.tokens, block.endChar
			pos += 1
		case "}", "]", ")":
			*ts = append(*ts, ParseError{tk: tokenPos, Kind: c, Message: "Unmatched " + c})
			pos += 1
		case `'`, `"`:
			value, pos, err = consumeQuotedString(css, pos)
			if err != nil {
				*ts = append(*ts, ParseError{tk: tokenPos, Kind: "bad-string", Message: "bad string token"})
			} else {
				*ts = append(*ts, StringToken{tk: tokenPos, Value: value})
			}
		}
		switch {
		case strings.HasPrefix(css[pos:], "/*"): // Comment
			index := strings.Index(css[pos+2:], "*/")
			pos += 2 + index
			if index == -1 {
				if !skipComments {
					*ts = append(*ts, Comment{tk: tokenPos, Value: css[tokenStartPos+2:]})
				}
				break
			}
			if !skipComments {
				*ts = append(*ts, Comment{tk: tokenPos, Value: css[tokenStartPos+2 : pos]})
			}
			pos += 2
		case strings.HasPrefix(css[pos:], "<!--"):
			*ts = append(*ts, LiteralToken{tk: tokenPos, Value: "<!--"})
			pos += 4
		case strings.HasPrefix(css[pos:], "||"):
			*ts = append(*ts, LiteralToken{tk: tokenPos, Value: "||"})
			pos += 2
		case strings.Contains("~|^$*", c):
			pos += 1
			if strings.HasPrefix(css[pos:], "=") {
				pos += 1
				*ts = append(*ts, LiteralToken{tk: tokenPos, Value: c + "="})
			} else {
				*ts = append(*ts, LiteralToken{tk: tokenPos, Value: c})
			}
		default:
			pos += 1
			*ts = append(*ts, LiteralToken{tk: tokenPos, Value: c})
		}
	}
	return out
}

const (
	charUnicodeRange = "0123456789abcdefABCDEF"
	nonPrintable     = "\"'(\x00\x01\x02\x03\x04\x05\x06\x07\x08\x0b\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f\x7f"
)

// Return true if the given character is a name-start code point.
func isNameStart(css string, pos int) bool {
	// https://www.w3.org/TR/css-syntax-3/#name-start-code-point
	c, _ := utf8.DecodeRuneInString(css[pos:])
	return c > 0x7F || strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_", c)
}

// Return true if the given position is the start of a CSS identifier.
func isIdentStart(css string, pos int) bool {
	// https://www.w3.org/TR/css-syntax-3/#would-start-an-identifier
	if isNameStart(css, pos) {
		return true
	} else if css[pos] == '-' {
		pos += 1
		// Name-start code point
		nameStart := pos < len(css) && (isNameStart(css, pos) || css[pos] == '-')
		// Valid escape
		validEscape := css[pos] == '\\' && !strings.HasPrefix(css[pos:], "\\\n")
		return nameStart || validEscape
	} else if css[pos] == '\\' {
		return !strings.HasPrefix(css[pos:], "\\\n")
	}
	return false
}

func consumeIdent(value string, pos int) (string, int) {
	// http://dev.w3.org/csswg/css-syntax/#consume-a-name
	var chunks []string
	length := len(value)
	startPos := pos
	for pos < length {
		c, w := utf8.DecodeRuneInString(value[pos:])
		if strings.ContainsRune("abcdefghijklmnopqrstuvwxyz-_0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ", c) || c > 0x7F {
			pos += w
		} else if c == '\\' && !strings.HasPrefix(value[pos:], "\\\n") {
			// Valid escape
			chunks = append(chunks, value[startPos:pos])
			var car string
			car, pos = consumeEscape(value, pos+w)
			chunks = append(chunks, car)
			startPos = pos
		} else {
			break
		}
	}
	chunks = append(chunks, value[startPos:pos])
	return strings.Join(chunks, ""), pos
}

// Return the range
// http://dev.w3.org/csswg/css-syntax/#consume-a-unicode-range-token
func consumeUnicodeRange(css string, pos int) (start, end int64, newPos int, err error) {
	length := len(css)
	startPos := pos
	maxPos := utils.MinInt(6, length)
	var _start, _end string
	for pos < maxPos && strings.ContainsRune(charUnicodeRange, rune(css[pos])) {
		pos += utf8.RuneLen(rune(css[pos]))
	}
	_start = css[startPos:pos]

	startPos = pos
	// Same maxPos as before: total of hex digits && question marks <= 6
	for pos < maxPos && css[pos] == '?' {
		pos += 1
	}
	questionMarks := pos - startPos

	if questionMarks != 0 {
		_end = _start + strings.Repeat("F", questionMarks)
		_start = _start + strings.Repeat("0", questionMarks)
	} else if pos+1 < length && css[pos] == '-' && strings.ContainsRune(charUnicodeRange, rune(css[pos+1])) {
		pos += utf8.RuneLen(rune(css[pos+1]))
		startPos = pos
		maxPos = utils.MinInt(pos+6, length)
		for pos < maxPos && strings.ContainsRune(charUnicodeRange, rune(css[pos])) {
			pos += utf8.RuneLen(rune(css[pos]))
		}
		_end = css[startPos:pos]
	} else {
		_end = _start
	}

	start, err = strconv.ParseInt(_start, 16, 0)
	if err != nil {
		return
	}
	end, err = strconv.ParseInt(_end, 16, 0)
	return start, end, pos, err
}

// http://dev.w3.org/csswg/css-syntax/#consume-a-url-token
func consumeUrl(css string, pos int) (value string, newPos int, err error) {
	length := len(css)
	// Skip whitespace
	for pos < length && strings.ContainsRune(" \n\t", rune(css[pos])) {
		pos += 1
	}
	if pos >= length { // EOF
		return "", pos, nil
	}
	c := rune(css[pos])
	if c == '"' || c == '\'' {
		value, pos, err = consumeQuotedString(css, pos)
		if err != nil {
			return "", pos, err
		}
	} else if c == ')' {
		return "", pos + 1, nil
	} else {
		var chunks []string
		startPos := pos
		for {
			if pos >= length { // EOF
				chunks = append(chunks, css[startPos:pos])
				return strings.Join(chunks, ""), pos, nil
			}
			c, w := utf8.DecodeRuneInString(css[pos:])
			switch c {
			case ')':
				chunks = append(chunks, css[startPos:pos])
				pos += w
				return strings.Join(chunks, ""), pos, nil
			case ' ', '\n', '\t':
				chunks = append(chunks, css[startPos:pos])
				value = strings.Join(chunks, "")
				pos += w
				break
			case '\\':
				if !strings.HasPrefix("\\\n", css[pos:]) {
					// Valid escape
					chunks = append(chunks, css[startPos:pos])
					var cs string
					cs, pos = consumeEscape(css, pos+1)
					chunks = append(chunks, cs)
					startPos = pos
				}
			default:
				pos += w
				// http://dev.w3.org/csswg/css-syntax/#non-printable-character
				if strings.ContainsRune(nonPrintable, c) {
					return "", pos, errors.New("non printable char")
				}
			}
		}
	}

	if err == nil {
		for pos < length && strings.ContainsRune(" \n\t", rune(css[pos])) {
			pos += 1
		}
		if pos < length {
			if css[pos] == ')' {
				return value, pos + 1, nil
			}
		} else {
			return value, pos, nil
		}
	}

	// http://dev.w3.org/csswg/css-syntax/#consume-the-remnants-of-a-bad-url0
	for pos < length {
		if strings.HasPrefix(css[pos:], "\\)") {
			pos += 2
		} else if css[pos] == ')' {
			pos += 1
			break
		} else {
			pos += 1
		}
	}
	return "", pos, errors.New("bad URL token") // bad-url
}

// Returnq unescapedValue
// http://dev.w3.org/csswg/css-syntax/#consume-a-string-token
func consumeQuotedString(css string, pos int) (string, int, error) {
	quote := rune(css[0])
	if quote != '"' && quote != '\'' {
		log.Fatal("first char should be a quote")
	}
	pos += 1
	var chunks []string
	length := len(css)
	startPos := pos
	for pos < length {
		c, w := utf8.DecodeRuneInString(css[pos:])
		switch c {
		case quote:
			chunks = append(chunks, css[startPos:pos])
			pos += w
			break
		case '\\':
			chunks = append(chunks, css[startPos:pos])
			pos += w
			if pos < length {
				if css[pos] == '\n' { // Ignore escaped newlines
					pos += 1
				} else {
					var c string
					c, pos = consumeEscape(css, pos)
					chunks = append(chunks, c)
				}
			} // else: Escaped EOF, do nothing
			startPos = pos
		case '\n': // Unescaped newline
			return "", pos, errors.New("bad-string") // bad-string
		default:
			pos += 1
		}
	}
	chunks = append(chunks, css[startPos:pos])
	return strings.Join(chunks, ""), pos, nil
}

// Return (unescapedChar, newPos).
// Assumes a valid escape: pos is just after '\' and not followed by '\n'.
func consumeEscape(css string, pos int) (string, int) {
	// http://dev.w3.org/csswg/css-syntax/#consume-an-escaped-character
	hexMatch := hexEscapeRe.FindStringSubmatch(css[pos:])
	if len(hexMatch) >= 2 {
		codepoint, err := strconv.ParseInt(hexMatch[1], 16, 0)
		if err != nil {
			log.Fatalf("codepoint should be valid hexadecimal, got %s", hexMatch[0])
		}
		char := "\uFFFD"
		if 0 < codepoint && codepoint <= unicode.MaxRune {
			char = string(rune(codepoint))
		}
		return char, pos + len(hexMatch[0])
	} else if pos < len(css) {
		return string(css[pos]), pos + 1
	} else {
		return "\uFFFD", pos
	}
}
