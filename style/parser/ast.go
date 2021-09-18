package parser

import (
	"encoding/json"

	"github.com/benoitkugler/go-weasyprint/utils"
)

const (
	QualifiedRuleT       tokenType = "qualified-rule"
	AtRuleT              tokenType = "at-rule"
	DeclarationT         tokenType = "declaration"
	ParseErrorT          tokenType = "error"
	CommentT             tokenType = "comment"
	WhitespaceTokenT     tokenType = "whitespace"
	LiteralTokenT        tokenType = "literal"
	IdentTokenT          tokenType = "ident"
	AtKeywordTokenT      tokenType = "at-keyword"
	HashTokenT           tokenType = "hash"
	StringTokenT         tokenType = "string"
	URLTokenT            tokenType = "url"
	UnicodeRangeTokenT   tokenType = "unicode-range"
	NumberTokenT         tokenType = "number"
	PercentageTokenT     tokenType = "percentage"
	DimensionTokenT      tokenType = "dimension"
	ParenthesesBlockT    tokenType = "() block"
	SquareBracketsBlockT tokenType = "[] block"
	CurlyBracketsBlockT  tokenType = "{} block"
	FunctionBlockT       tokenType = "function"
)

type tokenType string

type Token interface {
	jsonisable
	isToken()
	Position() Origine
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

type Origine struct {
	Line, Column int
}

func newOr(line, column int) Origine {
	return Origine{Line: line, Column: column}
}

// guards type
func (n Origine) isToken() {}

func (n Origine) Position() Origine {
	return n
}

// shared tokens
type stringToken struct {
	Value string
	Origine
}

type bracketsBlock struct {
	Content *[]Token
	Origine
}

type numericToken struct {
	Representation string
	Origine
	Value     float64
	IsInteger bool
}

type QualifiedRule struct {
	Prelude, Content *[]Token
	Origine
}

type AtRule struct {
	AtKeyword LowerableString
	QualifiedRule
}

type Declaration struct {
	Name  LowerableString
	Value []Token
	Origine
	Important bool
}

type ParseError struct {
	Kind    string
	Message string
	Origine
}

type (
	Comment         stringToken
	WhitespaceToken stringToken
	LiteralToken    stringToken
	IdentToken      struct {
		Value LowerableString
		Origine
	}
)

type AtKeywordToken struct {
	Value LowerableString
	Origine
}

type HashToken struct {
	Value string
	Origine
	IsIdentifier bool
}

type (
	StringToken       stringToken
	URLToken          stringToken
	UnicodeRangeToken struct {
		Origine
		Start, End uint32
	}
)

type (
	NumberToken     numericToken
	PercentageToken numericToken
	DimensionToken  struct {
		Unit LowerableString
		numericToken
	}
)

type (
	ParenthesesBlock    bracketsBlock
	SquareBracketsBlock bracketsBlock
	CurlyBracketsBlock  bracketsBlock
	FunctionBlock       struct {
		Arguments *[]Token
		Name      LowerableString
		Origine
	}
)

// ----------- boilerplate code for token type -------------------------------------

func (t QualifiedRule) Type() tokenType       { return QualifiedRuleT }
func (t AtRule) Type() tokenType              { return AtRuleT }
func (t Declaration) Type() tokenType         { return DeclarationT }
func (t ParseError) Type() tokenType          { return ParseErrorT }
func (t Comment) Type() tokenType             { return CommentT }
func (t WhitespaceToken) Type() tokenType     { return WhitespaceTokenT }
func (t LiteralToken) Type() tokenType        { return LiteralTokenT }
func (t IdentToken) Type() tokenType          { return IdentTokenT }
func (t AtKeywordToken) Type() tokenType      { return AtKeywordTokenT }
func (t HashToken) Type() tokenType           { return HashTokenT }
func (t StringToken) Type() tokenType         { return StringTokenT }
func (t URLToken) Type() tokenType            { return URLTokenT }
func (t UnicodeRangeToken) Type() tokenType   { return UnicodeRangeTokenT }
func (t NumberToken) Type() tokenType         { return NumberTokenT }
func (t PercentageToken) Type() tokenType     { return PercentageTokenT }
func (t DimensionToken) Type() tokenType      { return DimensionTokenT }
func (t ParenthesesBlock) Type() tokenType    { return ParenthesesBlockT }
func (t SquareBracketsBlock) Type() tokenType { return SquareBracketsBlockT }
func (t CurlyBracketsBlock) Type() tokenType  { return CurlyBracketsBlockT }
func (t FunctionBlock) Type() tokenType       { return FunctionBlockT }

// ---------------------------------- Methods ----------------------------------

// IntValue returns the rounded value
// Should be used only if  `IsInteger` is true
func (t numericToken) IntValue() int {
	return int(t.Value)
}

func (t NumberToken) IntValue() int {
	return numericToken(t).IntValue()
}

func (t PercentageToken) IntValue() int {
	return numericToken(t).IntValue()
}

// ---------------- JSON -------------------------------------------
type (
	myString string
	myFloat  float64
	myBool   bool
	myInt    int
)

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
	var content jsonisable
	if t.Content != nil {
		content = toJson(*t.Content)
	}
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
