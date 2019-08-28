package css

import (
	"github.com/benoitkugler/go-weasyprint/utils"
)

// Link the css parser with Weasyprint representation of CSS properties

type Token interface {
	isToken()
	String() string
}

type ComponentValue interface {
	Token
	isComponentValue()
}

// LowerableString is a string which can be
// normalized to ASCII lower case
type LowerableString string

func (s LowerableString) Lower() string {
	return utils.AsciiLower(string(s))
}

// guards type
type token struct{}
type componentValue struct{}

func (n token) isToken()                   {}
func (n componentValue) isComponentValue() {}

type QualifiedRule struct {
	token
	Prelude, Content []ComponentValue
}
type AtRule struct {
	QualifiedRule
	AtKeyword LowerableString
}
type Declaration struct {
	token
	Name      LowerableString
	Value     []ComponentValue
	Important bool
}

type ParseErrorType uint8

type ParseError struct {
	componentValue
	Kind    ParseErrorType
	Message string
}

type stringToken struct {
	componentValue
	Value string
}

type Comment stringToken
type WhitespaceToken stringToken
type LiteralToken stringToken
type IdentToken struct {
	componentValue
	Value LowerableString
}
type AtKeywordToken struct {
	componentValue
	Value LowerableString
}
type HashToken struct {
	componentValue
	Value        string
	IsIdentifier bool
}
type StringToken stringToken
type URLToken stringToken
type UnicodeRangeToken struct {
	componentValue
	Start, End uint32
}

type numericToken struct {
	componentValue
	Value          float32
	IsInteger      bool
	Representation string
}

type NumberToken numericToken
type PercentageToken numericToken
type DimensionToken struct {
	numericToken
	Unit LowerableString
}
type bracketsBlock struct {
	componentValue
	Content []ComponentValue
}

type ParenthesesBlock bracketsBlock
type SquareBracketsBlock bracketsBlock
type CurlyBracketsBlock bracketsBlock
type FunctionBlock struct {
	componentValue
	Name      LowerableString
	Arguments []ComponentValue
}

type MaybeInt struct {
	Valid bool
	Int   int
}

// type Token struct {
// 	Type tokenType
// 	Dimension
// 	String                string
// 	LowerValue, LowerName string
// 	IntValue              MaybeInt
// 	Arguments             []Token
// }
