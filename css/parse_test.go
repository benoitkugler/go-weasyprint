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
