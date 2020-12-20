package text

import (
	"fmt"
	"reflect"
	"testing"
)

func reduce(logs []CharAttr) []int {
	var out []int
	for _, l := range logs {
		if l.IsLineBreak() {
			out = append(out, 1)
		} else {
			out = append(out, 0)
		}
	}
	return out
}

func assertEqual(test lineBreakTest, got []CharAttr, t *testing.T) {
	if red := reduce(got); !reflect.DeepEqual(red, test.breaks) {
		t.Fatalf("%d : %s : expected %v, got %v (%v)", test.id, test.text, test.breaks, red, got)
	}
}

func TestLineBreak(t *testing.T) {
	for _, test := range lineBreakTests {
		got := pangoDefaultBreak([]rune(test.text))
		assertEqual(test, got, t)
	}
}

func TestA(t *testing.T) {
	fmt.Println([]int{0, 1, 2, 3}[1 : 4-1])
}
