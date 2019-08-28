package css

import (
	"fmt"
	"testing"

	"github.com/aymerick/douceur/parser"
)

func TestCss(t *testing.T) {
	s, err := parser.Parse("eee")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(s.String())
}

func TestInterface(t *testing.T) {
	var i Token
	i = QualifiedRule{}
	i = AtRule{}
	i = Declaration{}

	var t ComponentValue
	t = ParseErrorType{}
	t = ParseError{}
	t = stringToken{}
	t = Comment{}
	t = WhitespaceToken{}
	t = LiteralToken{}
	t = IdentToken{}
	t = AtKeywordToken{}
	t = HashToken{}
	t = StringToken{}
	t = URLToken{}
	t = UnicodeRangeToken{}
	t = NumberToken{}
	t = PercentageToken{}
	t = DimensionToken{}
	t = ParenthesesBlock{}
	t = SquareBracketsBlock{}
	t = CurlyBracketsBlock{}
	t = FunctionBlock{}
}
