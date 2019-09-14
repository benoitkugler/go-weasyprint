package parser

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

	i = ParseError{}
	i = Comment{}
	i = WhitespaceToken{}
	i = LiteralToken{}
	i = IdentToken{}
	i = AtKeywordToken{}
	i = HashToken{}
	i = StringToken{}
	i = URLToken{}
	i = UnicodeRangeToken{}
	i = NumberToken{}
	i = PercentageToken{}
	i = DimensionToken{}
	i = ParenthesesBlock{}
	i = SquareBracketsBlock{}
	i = CurlyBracketsBlock{}
	i = FunctionBlock{}
	fmt.Println(i)
}
