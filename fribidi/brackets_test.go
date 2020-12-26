package fribidi

import (
	"fmt"
	"testing"

	"golang.org/x/text/unicode/bidi"
)

func TestBracketsTable(t *testing.T) {
	for r, p := range bracketsTable {

		if prop, _ := bidi.LookupRune(r); !prop.IsBracket() {
			t.Errorf("rune %d is not a bracket", r)
		}

		if prop, _ := bidi.LookupRune(p); !prop.IsBracket() {
			t.Errorf("rune %d is not a bracket", p)
		}
	}
}

func Test(t *testing.T) {
	r := 9001
	fmt.Println(r&0x3f, r>>6)
}
