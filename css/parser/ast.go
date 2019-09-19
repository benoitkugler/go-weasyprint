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
	jsonisable
	isToken()
	Type() tokenType
	serializeTo(write func(s string))
}

type jsonisable interface {
	toJson() jsonisable // json representation, for tests
}

// LowerableString is a string which can be
// normalized to ASCII lower case
type LowerableString string

func (s LowerableString) Lower() string {
	return utils.AsciiLower(string(s))
}

// guards type
type tk struct {
	line, column int
}

func newTk(line, column int) tk {
	return tk{line: line, column: column}
}

func (n tk) isToken() {}

// shared tokens
type stringToken struct {
	tk
	Value string
}
type bracketsBlock struct {
	tk
	Content *[]Token
}
type numericToken struct {
	tk
	Value          float32
	IsInteger      bool
	Representation string
}

type QualifiedRule struct {
	tk
	Prelude, Content *[]Token
}
type AtRule struct {
	QualifiedRule
	AtKeyword LowerableString
}
type Declaration struct {
	tk
	Name      LowerableString
	Value     []Token
	Important bool
}
type ParseError struct {
	tk
	Kind    string
	Message string
}
type Comment stringToken
type WhitespaceToken stringToken
type LiteralToken stringToken
type IdentToken struct {
	tk
	Value LowerableString
}
type AtKeywordToken struct {
	tk
	Value LowerableString
}
type HashToken struct {
	tk
	Value        string
	IsIdentifier bool
}
type StringToken stringToken
type URLToken stringToken
type UnicodeRangeToken struct {
	tk
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
	tk
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
type myString string
type myFloat float32
type myBool bool
type myInt int

func (s myString) toJson() jsonisable { return s }
func (s myFloat) toJson() jsonisable  { return s }
func (s myBool) toJson() jsonisable   { return s }
func (s myInt) toJson() jsonisable    { return s }

type jsonList []jsonisable

func (s jsonList) toJson() jsonisable {
	for i, v := range s {
		s[i] = v.toJson()
	}
	return s
}

func (n numericToken) toJson() jsonList {
	l := jsonList{myString(n.Representation), myFloat(n.Value)}
	if n.IsInteger {
		l = append(l, myString("integer"))
	} else {
		l = append(l, myString("number"))
	}
	return l
}

func toJson(l []Token) jsonList {
	out := make(jsonList, len(l))
	for i, t := range l {
		out[i] = t.toJson()
	}
	return out
}

func marshalJSON(l []Token) (string, error) {
	normalize := toJson(l)
	b, err := json.Marshal(normalize)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (t QualifiedRule) toJson() jsonisable {
	prelude := toJson(*t.Prelude)
	content := toJson(*t.Content)
	return jsonList{myString("qualified rule"), prelude, content}
}
func (t AtRule) toJson() jsonisable {
	prelude := toJson(*t.Prelude)
	content := toJson(*t.Content)
	return jsonList{myString("at-rule"), myString(t.AtKeyword), prelude, content}
}
func (t Declaration) toJson() jsonisable {
	content := toJson(t.Value)
	return jsonList{myString("declaration"), myString(t.Name), content, myBool(t.Important)}
}
func (t ParseError) toJson() jsonisable {
	return jsonList{myString("error"), myString(t.Kind)}
}
func (t Comment) toJson() jsonisable {
	return myString("/* … */")
}
func (t WhitespaceToken) toJson() jsonisable {
	return myString(" ")
}
func (t LiteralToken) toJson() jsonisable {
	return myString(t.Value)
}
func (t IdentToken) toJson() jsonisable {
	return jsonList{myString("ident"), myString(t.Value)}
}
func (t AtKeywordToken) toJson() jsonisable {
	return jsonList{myString("at-keyword"), myString(t.Value)}
}
func (t HashToken) toJson() jsonisable {
	l := jsonList{myString("hash"), myString(t.Value)}
	if t.IsIdentifier {
		l = append(l, myString("id"))
	} else {
		l = append(l, myString("unrestricted"))
	}
	return l
}
func (t StringToken) toJson() jsonisable {
	return jsonList{myString("string"), myString(t.Value)}
}
func (t URLToken) toJson() jsonisable {
	return jsonList{myString("url"), myString(t.Value)}
}
func (t UnicodeRangeToken) toJson() jsonisable {
	return jsonList{myString("unicode-range"), myInt(t.Start), myInt(t.End)}
}
func (t NumberToken) toJson() jsonisable {
	return append(jsonList{myString("number")}, numericToken(t).toJson()...)
}
func (t PercentageToken) toJson() jsonisable {
	return append(jsonList{myString("percentage")}, numericToken(t).toJson()...)
}
func (t DimensionToken) toJson() jsonisable {
	return append(append(jsonList{myString("dimension")}, t.numericToken.toJson()...), myString(t.Unit))
}
func (t ParenthesesBlock) toJson() jsonisable {
	content := toJson(*t.Content)
	return append(jsonList{myString("()")}, content...)
}
func (t SquareBracketsBlock) toJson() jsonisable {
	content := toJson(*t.Content)
	return append(jsonList{myString("[]")}, content...)
}
func (t CurlyBracketsBlock) toJson() jsonisable {
	content := toJson(*t.Content)
	return append(jsonList{myString("{}")}, content...)
}
func (t FunctionBlock) toJson() jsonisable {
	content := toJson(*t.Arguments)
	return append(jsonList{myString("function"), myString(t.Name)}, content...)
}
