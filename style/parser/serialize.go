package parser

import (
	"fmt"
	"log"
	"strings"
	"unicode/utf8"
)

var badPairs = map[[2]string]bool{}

func init() {
	for _, a := range []string{"ident", "at-keyword", "hash", "dimension", "#", "-", "number"} {
		for _, b := range []string{"ident", "function", "url", "number", "percentage", "dimension", "unicode-range"} {
			badPairs[[2]string{a, b}] = true
		}
	}
	for _, a := range []string{"ident", "at-keyword", "hash", "dimension"} {
		for _, b := range []string{"-", "-->"} {
			badPairs[[2]string{a, b}] = true
		}
	}
	for _, a := range []string{"#", "-", "number", "@"} {
		for _, b := range []string{"ident", "function", "url"} {
			badPairs[[2]string{a, b}] = true
		}
	}
	for _, a := range []string{"unicode-range", ".", "+"} {
		for _, b := range []string{"number", "percentage", "dimension"} {
			badPairs[[2]string{a, b}] = true
		}
	}
	for _, b := range []string{"ident", "function", "url", "unicode-range", "-"} {
		badPairs[[2]string{"@", b}] = true
	}
	for _, b := range []string{"ident", "function", "?"} {
		badPairs[[2]string{"unicode-range", b}] = true
	}
	for _, a := range []string{"$", "*", "^", "~", "|"} {
		badPairs[[2]string{a, "="}] = true
	}
	badPairs[[2]string{"ident", "() block"}] = true
	badPairs[[2]string{"|", "|"}] = true
	badPairs[[2]string{"/", "*"}] = true
}

// Serialize nodes to CSS syntax.
// This should be used for `ComponentValue` as it takes care of corner cases such as ``;`` between declarations,
// and consecutive identifiers that would otherwise parse back as the same token.
func Serialize(nodes []Token) string {
	var chunks []string
	write := func(s string) { chunks = append(chunks, s) }
	serializeTo(nodes, write)
	return strings.Join(chunks, "")
}

// Serialize this node to CSS syntax
func SerializeOne(node Token) string {
	var chunks []string
	write := func(s string) { chunks = append(chunks, s) }
	node.serializeTo(write)
	return strings.Join(chunks, "")
}

// Serialize any string as a CSS identifier
// Returns an Unicode string
// that would parse as an `IdentToken`
// whose value attribute equals the passed `value` argument.
func serializeIdentifier(value string) string {
	if value == "-" {
		return `\-`
	}

	if len(value) >= 2 && value[:2] == "--" {
		return "--" + serializeName(value[2:])
	}
	var result string
	if value[0] == '-' {
		result = "-"
		value = value[1:]
	} else {
		result = ""
	}
	c, w := utf8.DecodeRuneInString(value)
	var suffix string
	switch c {
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '_', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		suffix = string(c)
	case '\n':
		suffix = `\A `
	case '\r':
		suffix = `\D `
	case '\f':
		suffix = `\C `
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		suffix = fmt.Sprintf("\\%X", c)
	default:
		if c > 0x7F {
			suffix = string(c)
		} else {
			suffix = "\\" + string(c)
		}

	}
	result += suffix + serializeName(value[w:])
	return result
}

func serializeName(value string) string {
	chuncks := make([]string, 0, len(value))
	for _, c := range value {
		var mapped string
		switch c {
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '-', '_', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
			mapped = string(c)
		case '\n':
			mapped = `\A `
		case '\r':
			mapped = `\D `
		case '\f':
			mapped = `\C `
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			mapped = string(c)
		default:
			if c > 0x7F {
				mapped = string(c)
			} else {
				mapped = "\\" + string(c)
			}
		}
		chuncks = append(chuncks, mapped)
	}
	return strings.Join(chuncks, "")
}

func serializeStringValue(value string) string {
	chuncks := make([]string, 0, len(value))
	for _, c := range value {
		var mapped string
		switch c {
		case '"':
			mapped = `\"`
		case '\\':
			mapped = `\\`
		case '\n':
			mapped = `\A `
		case '\r':
			mapped = `\D `
		case '\f':
			mapped = `\C `
		default:
			mapped = string(c)
		}
		chuncks = append(chuncks, mapped)
	}
	return strings.Join(chuncks, "")
}

// http://dev.w3.org/csswg/css-syntax/#serialization-tables
// Serialize an iterable of nodes to CSS syntax,
// writing chunks as Unicode string
// by calling the provided `write` callback.
func serializeTo(nodes []Token, write func(s string)) {
	var previousType string
	for _, node := range nodes {
		serializationType := string(node.Type())
		if literal, ok := node.(LiteralToken); ok {
			serializationType = literal.Value
		}
		if badPairs[[2]string{previousType, serializationType}] {
			write("/**/")
		} else if previousType == "\\" {
			whitespace, ok := node.(WhitespaceToken)
			ok = ok && strings.HasPrefix(whitespace.Value, "\n")
			if !ok {
				write("\n")
			}
		}
		node.serializeTo(write)
		if serializationType == string(TypeDeclaration) {
			write(";")
		}
		previousType = serializationType
	}
}

func (t QualifiedRule) serializeTo(write func(s string)) {
	serializeTo(*t.Prelude, write)
	write("{")
	serializeTo(*t.Content, write)
	write("}")
}

func (t AtRule) serializeTo(write func(s string)) {
	write("@")
	write(serializeIdentifier(string(t.AtKeyword)))
	serializeTo(*t.Prelude, write)
	if t.Content == nil {
		write(";")
	} else {
		write("{")
		serializeTo(*t.Content, write)
		write("}")
	}
}

func (t Declaration) serializeTo(write func(s string)) {
	write(serializeIdentifier(string(t.Name)))
	write(":")
	serializeTo(t.Value, write)
	if t.Important {
		write("!important")
	}
}

func (t ParseError) serializeTo(write func(s string)) {
	switch t.Kind {
	case "bad-string":
		write("\"[bad string]\n")
	case "bad-url":
		write("url([bad url])")
	case ")", "]", "}":
		write(t.Kind)
	default: // pragma: no cover
		log.Fatal("Can not serialize token", t)
	}
}

func (t Comment) serializeTo(write func(s string)) {
	write("/*")
	write(t.Value)
	write("*/")
}

func (t WhitespaceToken) serializeTo(write func(s string)) {
	write(t.Value)
}

func (t LiteralToken) serializeTo(write func(s string)) {
	write(t.Value)
}

func (t IdentToken) serializeTo(write func(s string)) {
	write(serializeIdentifier(string(t.Value)))
}

func (t AtKeywordToken) serializeTo(write func(s string)) {
	write("@")
	write(serializeIdentifier(string(t.Value)))
}

func (t HashToken) serializeTo(write func(s string)) {
	write("#")
	if t.IsIdentifier {
		write(serializeIdentifier(t.Value))
	} else {
		write(serializeName(t.Value))
	}
}

func (t StringToken) serializeTo(write func(s string)) {
	write(`"`)
	write(serializeStringValue(t.Value))
	write(`"`)
}

func (t URLToken) serializeTo(write func(s string)) {
	write(`url("`)
	write(serializeStringValue(t.Value))
	write(`")`)
}

func (t UnicodeRangeToken) serializeTo(write func(s string)) {
	if t.End == t.Start {
		write(fmt.Sprintf("U+%X", t.Start))
	} else {
		write(fmt.Sprintf("U+%X-%X", t.Start, t.End))
	}
}

func (t NumberToken) serializeTo(write func(s string)) {
	write(t.Representation)
}

func (t PercentageToken) serializeTo(write func(s string)) {
	write(t.Representation)
	write("%")
}

func (t DimensionToken) serializeTo(write func(s string)) {
	write(t.Representation)
	// Disambiguate with scientific notation
	unit := string(t.Unit)
	if unit == "e" || unit == "E" || strings.HasPrefix(unit, "e-") || strings.HasPrefix(unit, "E-") {
		write("\\65 ")
		write(serializeName(unit[1:]))
	} else {
		write(serializeIdentifier(unit))
	}
}

func (t ParenthesesBlock) serializeTo(write func(s string)) {
	write("(")
	serializeTo(*t.Content, write)
	write(")")
}

func (t SquareBracketsBlock) serializeTo(write func(s string)) {
	write("[")
	serializeTo(*t.Content, write)
	write("]")
}

func (t CurlyBracketsBlock) serializeTo(write func(s string)) {
	write("{")
	serializeTo(*t.Content, write)
	write("}")
}

func (t FunctionBlock) serializeTo(write func(s string)) {
	write(serializeIdentifier(string(t.Name)))
	write("(")
	serializeTo(*t.Arguments, write)
	write(")")
}
