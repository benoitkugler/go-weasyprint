package parser

import (
	"encoding/json"

	"github.com/benoitkugler/go-weasyprint/utils"
)

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
	MarshalJSON() ([]byte, error) // json representation, for tests
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
	Content *[]Token
}
type numericToken struct {
	_token
	Value          float32
	IsInteger      bool
	Representation string
}

type QualifiedRule struct {
	_token
	Prelude, Content *[]Token
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
	Arguments *[]Token
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

// ---------------- JSON -------------------------------------------
func (n numericToken) toList() []interface{} {
	l := []interface{}{n.Representation, n.Value}
	if n.IsInteger {
		l = append(l, "integer")
	} else {
		l = append(l, "number")
	}
	return l
}

func toJson(l []Token) (out [][]byte, err error) {
	out = make([][]byte, len(l))
	for i, t := range l {
		out[i], err = t.MarshalJSON()
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (t QualifiedRule) MarshalJSON() ([]byte, error) {
	prelude, err := toJson(*t.Prelude)
	if err != nil {
		return nil, err
	}
	content, err := toJson(*t.Content)
	if err != nil {
		return nil, err
	}
	l := []interface{}{"qualified rule", prelude, content}
	return json.Marshal(l)
}
func (t AtRule) MarshalJSON() ([]byte, error) {
	prelude, err := toJson(*t.Prelude)
	if err != nil {
		return nil, err
	}
	content, err := toJson(*t.Content)
	if err != nil {
		return nil, err
	}
	l := []interface{}{"at-rule", t.AtKeyword, prelude, content}
	return json.Marshal(l)
}
func (t Declaration) MarshalJSON() ([]byte, error) {
	content, err := toJson(t.Value)
	if err != nil {
		return nil, err
	}
	l := []interface{}{"declaration", t.Name, content, t.Important}
	return json.Marshal(l)
}
func (t ParseError) MarshalJSON() ([]byte, error) {
	l := []string{"error", t.Kind}
	return json.Marshal(l)
}
func (t Comment) MarshalJSON() ([]byte, error) {
	return []byte("/* … */"), nil
}
func (t WhitespaceToken) MarshalJSON() ([]byte, error) {
	return []byte(" "), nil
}
func (t LiteralToken) MarshalJSON() ([]byte, error) {
	return []byte(t.Value), nil
}
func (t IdentToken) MarshalJSON() ([]byte, error) {
	l := []string{"ident", string(t.Value)}
	return json.Marshal(l)
}
func (t AtKeywordToken) MarshalJSON() ([]byte, error) {
	l := []string{"at-keyword", string(t.Value)}
	return json.Marshal(l)
}
func (t HashToken) MarshalJSON() ([]byte, error) {
	l := []string{"hash", t.Value}
	if t.IsIdentifier {
		l = append(l, "id")
	} else {
		l = append(l, "unrestricted")
	}
	return json.Marshal(l)
}
func (t StringToken) MarshalJSON() ([]byte, error) {
	l := []string{"string", t.Value}
	return json.Marshal(l)
}
func (t URLToken) MarshalJSON() ([]byte, error) {
	l := []string{"url", t.Value}
	return json.Marshal(l)
}
func (t UnicodeRangeToken) MarshalJSON() ([]byte, error) {
	l := []interface{}{"unicode-range", t.Start, t.End}
	return json.Marshal(l)
}
func (t NumberToken) MarshalJSON() ([]byte, error) {
	l := append([]interface{}{"number"}, numericToken(t).toList()...)
	return json.Marshal(l)
}
func (t PercentageToken) MarshalJSON() ([]byte, error) {
	l := append([]interface{}{"percentage"}, numericToken(t).toList()...)
	return json.Marshal(l)
}
func (t DimensionToken) MarshalJSON() ([]byte, error) {
	l := append(append([]interface{}{"dimension"}, t.toList()...), t.Unit)
	return json.Marshal(l)
}
func (t ParenthesesBlock) MarshalJSON() ([]byte, error) {
	content, err := toJson(*t.Content)
	if err != nil {
		return nil, err
	}
	l := append([][]byte{[]byte("()")}, content...)
	return json.Marshal(l)
}
func (t SquareBracketsBlock) MarshalJSON() ([]byte, error) {
	content, err := toJson(*t.Content)
	if err != nil {
		return nil, err
	}
	l := append([][]byte{[]byte("[]")}, content...)
	return json.Marshal(l)
}
func (t CurlyBracketsBlock) MarshalJSON() ([]byte, error) {
	content, err := toJson(*t.Content)
	if err != nil {
		return nil, err
	}
	l := append([][]byte{[]byte("{}")}, content...)
	return json.Marshal(l)
}
func (t FunctionBlock) MarshalJSON() ([]byte, error) {
	content, err := toJson(*t.Arguments)
	if err != nil {
		return nil, err
	}
	l := append([][]byte{[]byte("function"), []byte(t.Name)}, content...)
	return json.Marshal(l)
}
