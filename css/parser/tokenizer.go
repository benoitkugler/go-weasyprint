package parser

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/benoitkugler/go-weasyprint/utils"
	// "github.com/gorilla/css/scanner"

	"github.com/riking/cssparse/scanner"
)

// This file builds the ast from the Gorilla tokenizer.

var (
	numberRe    = regexp.MustCompile(`[-+]?([0-9]*\.)?[0-9]+([eE][+-]?[0-9]+)?`)
	hexEscapeRe = regexp.MustCompile(`([0-9A-Fa-f]{1,6})[ \n\t]?`)

	litteralTokens = map[int]string{
		int(scanner.TokenCDC):            "-->",
		int(scanner.TokenCDO):            "<!--",
		int(scanner.TokenIncludes):       "~=",
		int(scanner.TokenDashMatch):      "|=",
		int(scanner.TokenPrefixMatch):    "^=",
		int(scanner.TokenSuffixMatch):    "$=",
		int(scanner.TokenSubstringMatch): "*=",
	}
)

type nestedBlock struct {
	tokens  *[]Token
	endChar string
}

// Parse a list of component values.
// If `skipComments` is true, ignore CSS comments :
// the return values (and recursively its blocks and functions)
// will not contain any `Comment` object.
func ParseComponentValueList(css string, skipComments bool) []Token {
	// This turns out to be faster than a regexp:
	css = strings.ReplaceAll(css, "\r\n", "\n")
	css = strings.ReplaceAll(css, "\r", "\n")
	css = strings.ReplaceAll(css, "\f", "\n")

	scan := scanner.New(css)
	var out []Token    // possibly nested tokens
	ts := &out         // current stack of tokens
	var endChar string // Pop the stack when encountering this character.
	var stack []nestedBlock
	for {
		token := scan.Next()
		// fmt.Println(token)
		switch token.Type {
		case scanner.TokenEOF, scanner.TokenError:
			return out
		case scanner.TokenS:
			*ts = append(*ts, WhitespaceToken{Value: token.Value})
		case scanner.TokenUnicodeRange:
			start, end, err := consumeUnicodeRange(token.Value)
			if err != nil {
				*ts = append(*ts, ParseError{Kind: "invalid number", Message: err.Error()})
			} else {
				*ts = append(*ts, UnicodeRangeToken{Start: uint32(start), End: uint32(end)})
			}
		case scanner.TokenIdent:
			*ts = append(*ts, IdentToken{Value: LowerableString(consumeIdent(token.Value))})
		case scanner.TokenURI:
			url, err := consumeUrl(token.Value)
			if err != nil {
				*ts = append(*ts, ParseError{Kind: "bad-url", Message: err.Error()})
			} else {
				*ts = append(*ts, URLToken{Value: url})
			}
		case scanner.TokenFunction:
			funcBlock := FunctionBlock{
				Name:      LowerableString(strings.TrimSuffix(token.Value, "(")),
				Arguments: new([]Token),
			}
			*ts = append(*ts, funcBlock)
			stack = append(stack, nestedBlock{tokens: ts, endChar: endChar})
			endChar = ")"
			ts = funcBlock.Arguments
		case scanner.TokenDimension:
			num, unit := consumeNumber(token.Value)
			*ts = append(*ts, DimensionToken{Unit: LowerableString(unit), numericToken: num})
		case scanner.TokenPercentage:
			num, _ := consumeNumber(token.Value)
			*ts = append(*ts, PercentageToken(num))
		case scanner.TokenNumber:
			num, _ := consumeNumber(token.Value)
			*ts = append(*ts, NumberToken(num))
		case scanner.TokenAtKeyword:
			*ts = append(*ts, AtKeywordToken{
				Value: LowerableString(strings.TrimPrefix(token.Value, "@")),
			})
		case scanner.TokenCDC, scanner.TokenCDO, scanner.TokenIncludes, scanner.TokenDashMatch, scanner.TokenPrefixMatch, scanner.TokenSuffixMatch, scanner.TokenSubstringMatch:
			*ts = append(*ts, LiteralToken{Value: litteralTokens[int(token.Type)]})
		case scanner.TokenHash:
			isIdentifier := isIdentStart([]rune(token.Value), 1)
			*ts = append(*ts, HashToken{Value: token.Value[1:], IsIdentifier: isIdentifier})
		case scanner.TokenString:
			str, err := consumeQuotedString(token.Value)
			if err != nil {
				*ts = append(*ts, ParseError{Kind: "bad-string", Message: "bad string token"})
			} else {
				*ts = append(*ts, StringToken{Value: str})
			}
		case scanner.TokenComment:
			cmt := strings.TrimSuffix(strings.TrimPrefix(token.Value, "/*"), "*/")
			if !skipComments {
				*ts = append(*ts, Comment{Value: cmt})
			}
		case scanner.TokenChar:
			switch token.Value {
			case "{":
				brack := CurlyBracketsBlock{Content: new([]Token)}
				*ts = append(*ts, brack)
				stack = append(stack, nestedBlock{tokens: ts, endChar: endChar})
				endChar = "}"
				ts = brack.Content
			case "[":
				brack := SquareBracketsBlock{Content: new([]Token)}
				*ts = append(*ts, brack)
				stack = append(stack, nestedBlock{tokens: ts, endChar: endChar})
				endChar = "]"
				ts = brack.Content
			case "(":
				brack := ParenthesesBlock{Content: new([]Token)}
				*ts = append(*ts, brack)
				stack = append(stack, nestedBlock{tokens: ts, endChar: endChar})
				endChar = ")"
				ts = brack.Content
			case endChar: // Matching }, ] or )
				// The top-level endChar is "" (never equal to a character),
				// so we never get here if the stack is empty.
				var block nestedBlock
				block, stack = stack[len(stack)-1], stack[:len(stack)-1]
				ts, endChar = block.tokens, block.endChar
			case "}", "]", ")":
				*ts = append(*ts, ParseError{Kind: token.Value, Message: "Unmatched " + token.Value})
			default:
				*ts = append(*ts, LiteralToken{Value: token.Value})
			}
		default:
			log.Fatalf("unssupported token : %s", token)
		}
	}
}

const (
	charUnicodeRange = "0123456789abcdefABCDEF"
	nonPrintable     = "\"'(\x00\x01\x02\x03\x04\x05\x06\x07\x08\x0b\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f\x7f"
)

// Return true if the given character is a name-start code point.
func isNameStart(css []rune, pos int) bool {
	// https://www.w3.org/TR/css-syntax-3/#name-start-code-point
	c := css[pos]
	return c > 0x7F || strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", c)
}

// Return true if the given position is the start of a CSS identifier.
func isIdentStart(css []rune, pos int) bool {
	// https://www.w3.org/TR/css-syntax-3/#would-start-an-identifier
	if isNameStart(css, pos) {
		return true
	} else if css[pos] == '-' {
		pos += 1
		// Name-start code point
		nameStart := pos < len(css) && isNameStart(css, pos)
		// Valid escape
		validEscape := css[pos] == '\\' && !strings.HasPrefix(string(css[pos:]), "\\\n")
		return nameStart || validEscape
	} else if css[pos] == '\\' {
		return !strings.HasPrefix(string(css[pos:]), "\\\n")
	}
	return false
}

func consumeIdent(_value string) string {
	// http://dev.w3.org/csswg/css-syntax/#consume-a-name
	var chunks []string
	value := []rune(_value)
	length := len(value)
	pos := 0
	startPos := pos
	for pos < length {
		c := value[pos]
		if strings.ContainsRune("abcdefghijklmnopqrstuvwxyz-0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ", c) || c > 0x7F {
			pos += 1
		} else if c == '\\' && !strings.HasPrefix(string(value[pos:]), "\\\n") {
			// Valid escape
			chunks = append(chunks, string(value[startPos:pos]))
			var car string
			car, pos = consumeEscape(value, pos+1)
			chunks = append(chunks, car)
			startPos = pos
		} else {
			break
		}
	}
	chunks = append(chunks, string(value[startPos:pos]))
	return strings.Join(chunks, "")
}

func consumeNumber(value string) (numericToken, string) {
	match := numberRe.FindStringIndex(value)
	if match != nil {
		repr := value[match[0]:match[1]]
		suffix := value[match[1]:]
		value, err := strconv.ParseFloat(repr, 32)
		if err != nil {
			log.Fatalf("should be a number %s \n", err)
		}
		_, err = strconv.ParseInt(repr, 10, 0)
		isInt := err != nil
		return numericToken{
			Representation: repr,
			IsInteger:      isInt,
			Value:          float32(value),
		}, suffix
	}
	log.Fatalln("input should be a numeric token")
	return numericToken{}, ""
}

// Return the range
// http://dev.w3.org/csswg/css-syntax/#consume-a-unicode-range-token
func consumeUnicodeRange(value string) (start, end int64, err error) {
	css := []rune(value)
	length := len(css)
	pos := 2
	startPos := pos
	maxPos := utils.MinInt(6, length)
	var _start, _end string
	for pos < maxPos && strings.ContainsRune(charUnicodeRange, css[pos]) {
		pos += 1
	}
	_start = string(css[startPos:pos])

	startPos = pos
	// Same maxPos as before: total of hex digits && question marks <= 6
	for pos < maxPos && css[pos] == '?' {
		pos += 1
	}
	questionMarks := pos - startPos

	if questionMarks != 0 {
		_end = _start + strings.Repeat("F", questionMarks)
		_start = _start + strings.Repeat("0", questionMarks)
	} else if pos+1 < length && css[pos] == '-' && strings.ContainsRune(charUnicodeRange, css[pos+1]) {
		pos += 1
		startPos = pos
		maxPos = utils.MinInt(pos+6, length)
		for pos < maxPos && strings.ContainsRune(charUnicodeRange, css[pos]) {
			pos += 1
		}
		_end = string(css[startPos:pos])
	} else {
		_end = _start
	}

	start, err = strconv.ParseInt(_start, 16, 0)
	if err != nil {
		return
	}
	end, err = strconv.ParseInt(_end, 16, 0)
	return
}

// http://dev.w3.org/csswg/css-syntax/#consume-a-url-token
func consumeUrl(input string) (value string, err error) {
	css := []rune(input)
	pos := 4
	length := len(css)
	// Skip whitespace
	for pos < length && strings.ContainsRune(" \n\t", css[pos]) {
		pos += 1
	}
	if pos >= length { // EOF
		return "", nil
	}
	c := css[pos]
	if c == '"' || c == '\'' {
		value, err = consumeQuotedString(string(css[pos:]))
		if err != nil {
			return "", err
		}
	} else if c == ')' {
		return "", nil
	} else {
		var chunks []string
		startPos := pos
		for {
			if pos >= length { // EOF
				chunks = append(chunks, string(css[startPos:pos]))
				return strings.Join(chunks, ""), nil
			}
			c = css[pos]
			switch c {
			case ')':
				chunks = append(chunks, string(css[startPos:pos]))
				return strings.Join(chunks, ""), nil
			case ' ', '\n', '\t':
				chunks = append(chunks, string(css[startPos:pos]))
				value = strings.Join(chunks, "")
				pos += 1
				break
			case '\\':
				if !strings.HasPrefix("\\\n", string(css[pos:])) {
					// Valid escape
					chunks = append(chunks, string(css[startPos:pos]))
					var cs string
					cs, pos = consumeEscape(css, pos+1)
					chunks = append(chunks, cs)
					startPos = pos
				}
			default:
				// http://dev.w3.org/csswg/css-syntax/#non-printable-character
				if strings.ContainsRune(nonPrintable, c) {
					return "", errors.New("non printable char")
				} else {
					pos += 1
				}
			}
		}
	}

	if err == nil {
		for pos < length && strings.ContainsRune(" \n\t", css[pos]) {
			pos += 1
		}
		if pos < length {
			if css[pos] == ')' {
				return value, nil
			}
		} else {
			return value, nil
		}
	}

	// http://dev.w3.org/csswg/css-syntax/#consume-the-remnants-of-a-bad-url0
	// handled by gorilla
	return "", errors.New("bad URL token") // bad-url
}

// Returnq unescapedValue
// http://dev.w3.org/csswg/css-syntax/#consume-a-string-token
func consumeQuotedString(value string) (string, error) {
	css := []rune(value)
	quote := css[0]
	if quote != '"' && quote != '\'' {
		log.Fatal("first char should be a quote")
	}
	pos := 1
	var chunks []string
	length := len(css)
	startPos := pos
	for pos < length {
		switch css[pos] {
		case quote:
			chunks = append(chunks, string(css[startPos:pos]))
			pos += 1
			break
		case '\\':
			chunks = append(chunks, string(css[startPos:pos]))
			pos += 1
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
			return "", errors.New("bad-string") // bad-string
		default:
			pos += 1
		}
	}
	chunks = append(chunks, string(css[startPos:pos]))
	return strings.Join(chunks, ""), nil
}

// Return (unescapedChar, newPos).
// Assumes a valid escape: pos is just after '\' and not followed by '\n'.
func consumeEscape(css []rune, pos int) (string, int) {
	// http://dev.w3.org/csswg/css-syntax/#consume-an-escaped-character
	hexMatch := hexEscapeRe.FindStringSubmatch(string(css[pos:]))
	if len(hexMatch) >= 2 {
		codepoint, err := strconv.ParseInt(hexMatch[1], 16, 0)
		if err != nil {
			fmt.Println(string(css[pos:]), hexMatch)
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
