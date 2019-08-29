package css

import (
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
type token struct{}

func (n token) isToken() {}

// shared tokens
type stringToken struct {
	token
	Value string
}
type bracketsBlock struct {
	token
	Content []Token
}
type numericToken struct {
	token
	Value          float32
	IsInteger      bool
	Representation string
}

type QualifiedRule struct {
	token
	Prelude, Content []Token
}
type AtRule struct {
	QualifiedRule
	AtKeyword LowerableString
}
type Declaration struct {
	token
	Name      LowerableString
	Value     []Token
	Important bool
}
type ParseError struct {
	token
	Kind    string
	Message string
}
type Comment stringToken
type WhitespaceToken stringToken
type LiteralToken stringToken
type IdentToken struct {
	token
	Value LowerableString
}
type AtKeywordToken struct {
	token
	Value LowerableString
}
type HashToken struct {
	token
	Value        string
	IsIdentifier bool
}
type StringToken stringToken
type URLToken stringToken
type UnicodeRangeToken struct {
	token
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
	token
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

// type MaybeInt struct {
// 	Valid bool
// 	Int   int
// }

// type Token struct {
// 	Type tokenType
// 	Dimension
// 	String                string
// 	LowerValue, LowerName string
// 	IntValue              MaybeInt
// 	Arguments             []Token
// }
