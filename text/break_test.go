package text

import (
	"reflect"
	"testing"
)

func reduce(logs []PangoLogAttr) []int {
	var out []int
	for _, l := range logs {
		if l.IsLineBreak {
			out = append(out, 1)
		} else {
			out = append(out, 0)
		}
	}
	return out
}

func assertEqual(test lineBreakTest, got []PangoLogAttr, t *testing.T) {
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
