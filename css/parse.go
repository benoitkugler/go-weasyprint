package css

import (
	"fmt"

	"github.com/benoitkugler/go-weasyprint/utils"
)

// Link the css parser with Weasyprint representation of CSS properties

const (
	TypeQualifiedRule       tokenType = "qualified-rule"
	TypeAtRule              tokenType = "at-rule"
	TypeDeclaration         tokenType = "declaration"
	TypeParseError          tokenType = "error"
	TypeComment             tokenType = "comment"
	TypeWhitespaceToken     tokenType = "whitespace"
	TypeLiteralToken        tokenType = "literal"
	TypeIdentToken          tokenType = "ident"
	TypeAtKeywordToken      tokenType = "at-keyword"
	TypeHashToken           tokenType = "hash"
	TypeStringToken         tokenType = "string"
	TypeURLToken            tokenType = "url"
	TypeUnicodeRangeToken   tokenType = "unicode-range"
	TypeNumberToken         tokenType = "number"
	TypePercentageToken     tokenType = "percentage"
	TypeDimensionToken      tokenType = "dimension"
	TypeParenthesesBlock    tokenType = "() block"
	TypeSquareBracketsBlock tokenType = "[] block"
	TypeCurlyBracketsBlock  tokenType = "{} block"
	TypeFunctionBlock       tokenType = "function"
)

type tokenType string

type Token interface {
	isToken()
	Type() tokenType
	serializeTo(write func(s string))
}

// LowerableString is a string which can be
// normalized to ASCII lower case
type LowerableString string

func (s LowerableString) Lower() string {
	return utils.AsciiLower(string(s))
}

// guards type
type _token struct{}

func (n _token) isToken() {}

// shared tokens
type stringToken struct {
	_token
	Value string
}
type bracketsBlock struct {
	_token
	Content []Token
}
type numericToken struct {
	_token
	Value          float32
	IsInteger      bool
	Representation string
}

type QualifiedRule struct {
	_token
	Prelude, Content []Token
}
type AtRule struct {
	QualifiedRule
	AtKeyword LowerableString
}
type Declaration struct {
	_token
	Name      LowerableString
	Value     []Token
	Important bool
}
type ParseError struct {
	_token
	Kind    string
	Message string
}
type Comment stringToken
type WhitespaceToken stringToken
type LiteralToken stringToken
type IdentToken struct {
	_token
	Value LowerableString
}
type AtKeywordToken struct {
	_token
	Value LowerableString
}
type HashToken struct {
	_token
	Value        string
	IsIdentifier bool
}
type StringToken stringToken
type URLToken stringToken
type UnicodeRangeToken struct {
	_token
	Start, End uint32
}
type NumberToken numericToken
type PercentageToken numericToken
type DimensionToken struct {
	numericToken
	Unit LowerableString
}
type ParenthesesBlock bracketsBlock
type SquareBracketsBlock bracketsBlock
type CurlyBracketsBlock bracketsBlock
type FunctionBlock struct {
	_token
	Name      LowerableString
	Arguments []Token
}

// ----------- boilerplate code for token type -------------------------------------

func (t QualifiedRule) Type() tokenType       { return TypeQualifiedRule }
func (t AtRule) Type() tokenType              { return TypeAtRule }
func (t Declaration) Type() tokenType         { return TypeDeclaration }
func (t ParseError) Type() tokenType          { return TypeParseError }
func (t Comment) Type() tokenType             { return TypeComment }
func (t WhitespaceToken) Type() tokenType     { return TypeWhitespaceToken }
func (t LiteralToken) Type() tokenType        { return TypeLiteralToken }
func (t IdentToken) Type() tokenType          { return TypeIdentToken }
func (t AtKeywordToken) Type() tokenType      { return TypeAtKeywordToken }
func (t HashToken) Type() tokenType           { return TypeHashToken }
func (t StringToken) Type() tokenType         { return TypeStringToken }
func (t URLToken) Type() tokenType            { return TypeURLToken }
func (t UnicodeRangeToken) Type() tokenType   { return TypeUnicodeRangeToken }
func (t NumberToken) Type() tokenType         { return TypeNumberToken }
func (t PercentageToken) Type() tokenType     { return TypePercentageToken }
func (t DimensionToken) Type() tokenType      { return TypeDimensionToken }
func (t ParenthesesBlock) Type() tokenType    { return TypeParenthesesBlock }
func (t SquareBracketsBlock) Type() tokenType { return TypeSquareBracketsBlock }
func (t CurlyBracketsBlock) Type() tokenType  { return TypeCurlyBracketsBlock }
func (t FunctionBlock) Type() tokenType       { return TypeFunctionBlock }

// ---------------------------------- Methods ----------------------------------

// IntValue returns the rounded value
// Should be used only if  `IsInteger` is true
func (t NumberToken) IntValue() int {
	return int(t.Value)
}

// ---------------- Parsing ----------------------------------------------------

type tokenIterator struct {
	tokens []Token
	index  int
}

func NewTokenIterator(tokens []Token) *tokenIterator {
	return &tokenIterator{
		tokens: tokens,
	}
}

func (it tokenIterator) HasNext() bool {
	return it.index < len(it.tokens)
}

func (it *tokenIterator) Next() Token {
	t := it.tokens[it.index]
	it.index += 1
	return t
}

// Return the next significant (neither whitespace || comment) token.
//     :param tokens: An *iterator* yielding :term:`component values`.
//     :returns: A :term:`component value`, || :obj:`None`.
//
func nextSignificant(tokens *tokenIterator) Token {
	for tokens.HasNext() {
		token := tokens.Next()
		switch token.(type) {
		case WhitespaceToken, Comment:
			continue
		default:
			return token
		}
	}
	return nil
}

// Parse a declaration.
//     Consume :obj:`tokens` until the end of the declaration || the first error.
//     :param firstToken: The first :term:`component value` of the rule.
//     :param tokens: An *iterator* yielding :term:`component values`.
func parseDeclaration(firstToken Token, tokens *tokenIterator) Token {
	name, ok := firstToken.(IdentToken)
	if !ok {
		return ParseError{
			Kind:    "invalid",
			Message: fmt.Sprintf("Expected <ident> for declaration name, got %s.", firstToken.Type()),
		}
	}
	colon := nextSignificant(tokens)
	if colon == nil {
		return ParseError{Kind: "invalid",
			Message: "Expected ':' after declaration name, got EOF",
		}
	}
	lit, ok := colon.(LiteralToken)
	if !ok || lit.Value != ":" {
		return ParseError{Kind: "invalid",
			Message: fmt.Sprintf("Expected ':' after declaration name, got %s.", colon.Type()),
		}
	}

	var value []Token
	state := "value"
	bangPosition, i := 0, 0
	for tokens.HasNext() {
		i += 1
		_token := tokens.Next()
		switch token := _token.(type) {
		case LiteralToken:
			if state == "value" && lit.Value == "!" {
				state = "bang"
				bangPosition = i
			}
		case IdentToken:
			if state == "bang" && token.Value.Lower() == "important" {
				state = "important"
			}
		default:
			if _token.Type() != "whitespace" && _token.Type() != "comment" {
				state = "value"
			}
		}
		value = append(value, _token)
	}

	if state == "important" {
		value = value[:bangPosition]
	}

	return Declaration{
		Name:      name.Value,
		Value:     value,
		Important: state == "important",
	}
}

// Like :func:`parseDeclaration`, but stop at the first ``;``.
func consumeDeclarationInList(firstToken Token, tokens *tokenIterator) Token {
	var otherDeclarationTokens []Token
	for tokens.HasNext() {
		token := tokens.Next()
		if lit, ok := token.(LiteralToken); ok && lit.Value == ";" {
			break
		}
		otherDeclarationTokens = append(otherDeclarationTokens, token)
	}
	return parseDeclaration(firstToken, NewTokenIterator(otherDeclarationTokens))
}

// Parse a :diagram:`declaration list` (which may also contain at-rules).
// This is used e.g. for the `QualifiedRule.content`
// of a style rule or ``@page`` rule, or for the ``style`` attribute of an HTML element.
// In contexts that don’t expect any at-rule, all :class:`AtRule` objects should simply be rejected as invalid.
// If `skipComments`, ignore CSS comments at the top-level of the list. If the input is a string, ignore all comments.
// If `skipWhitespace`, ignore whitespace at the top-level of the list. Whitespace is still preserved in
// the `Declaration.value` of declarations and the `AtRule.prelude` and `AtRule.content` of at-rules.
// skipComments = false, skipWhitespace = false
func ParseDeclarationList(input []Token, skipComments, skipWhitespace bool) []Token {
	tokens := NewTokenIterator(input)
	var result []Token

	for tokens.HasNext() {
		switch token := tokens.Next().(type) {
		case WhitespaceToken:
			if !skipWhitespace {
				result = append(result, token)
			}
		case Comment:
			if !skipComments {
				result = append(result, token)
			}
		case AtKeywordToken:
			val := consumeAtRule(token, tokens)
			result = append(result, val)
		case LiteralToken:
			if token.Value != ";" {
				val := consumeDeclarationInList(token, tokens)
				result = append(result, val)
			}
		default:
			val := consumeDeclarationInList(token, tokens)
			result = append(result, val)
		}
	}
	return result
}

// Parse a non-top-level :diagram:`rule list`.
// This is used for parsing the `AtRule.content`
// of nested rules like ``@media``.
// This differs from :func:`parseStylesheet` in that
// top-level ``<!--`` and ``-->`` tokens are not ignored.
// :param skipComments:
//     Ignore CSS comments at the top-level of the list.
//     If the input is a string, ignore all comments.
// :param skipWhitespace:
//     Ignore whitespace at the top-level of the list.
//     Whitespace is still preserved
//     in the `QualifiedRule.prelude`
//     and the `QualifiedRule.content` of rules.
// skipComments=false, skipWhitespace=false
func parseRuleList(input []Token, skipComments, skipWhitespace bool) []Token {
	tokens := NewTokenIterator(input)
	var result []Token
	for tokens.HasNext() {
		token := tokens.Next()
		switch token.Type() {
		case TypeWhitespaceToken:
			if !skipWhitespace {
				result = append(result, token)
			}
		case TypeComment:
			if !skipComments {
				result = append(result, token)
			}
		default:
			val := consumeRule(token, tokens)
			result = append(result, val)
		}
	}
	return result
}

// Parse a qualified rule or at-rule.
// Consume just enough of :obj:`tokens` for this rule.
// :param firstToken: The first :term:`component value` of the rule.
// :param tokens: An *iterator* yielding :term:`component values`.
func consumeRule(_firstToken Token, tokens *tokenIterator) Token {
	var (
		prelude []Token
		block   CurlyBracketsBlock
	)
	switch firstToken := _firstToken.(type) {
	case AtKeywordToken:
		return consumeAtRule(firstToken, tokens)
	case CurlyBracketsBlock:
		block = firstToken
	default:
		prelude = []Token{firstToken}
		hasBroken := false
		for tokens.HasNext() {
			token := tokens.Next()
			if curly, ok := token.(CurlyBracketsBlock); ok {
				block = curly
				hasBroken = true
				break
			}
			prelude = append(prelude, token)
		}
		if !hasBroken {
			return ParseError{
				Kind:    "invalid",
				Message: "EOF reached before {} block for a qualified rule.",
			}
		}
	}
	return QualifiedRule{
		Content: block.Content,
		Prelude: prelude,
	}
}

// Parse an at-rule.
// Consume just enough of :obj:`tokens` for this rule.
// :param atKeyword: The :class:`AtKeywordToken` object starting this rule.
// :param tokens: An *iterator* yielding :term:`component values`.
func consumeAtRule(atKeyword AtKeywordToken, tokens *tokenIterator) AtRule {
	var prelude, content []Token
	for tokens.HasNext() {
		token := tokens.Next()
		if curly, ok := token.(CurlyBracketsBlock); ok {
			content = curly.Content
			break
		}
		lit, ok := token.(LiteralToken)
		if ok && lit.Value == ";" {
			break
		}
		prelude = append(prelude, token)
	}
	return AtRule{
		AtKeyword: atKeyword.Value,
		QualifiedRule: QualifiedRule{
			Prelude: prelude,
			Content: content,
		},
	}
}
