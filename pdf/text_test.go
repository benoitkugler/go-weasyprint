package pdf

import (
	"reflect"
	"testing"

	"github.com/benoitkugler/pdf/model"
	"github.com/benoitkugler/webrender/backend"
)

func TestCIDWidths(t *testing.T) {
	input := map[backend.GID]backend.GlyphExtents{
		2: {Width: 10},
		3: {Width: 11},
		4: {Width: 5},
		8: {Width: 13},
		6: {Width: 0},
		9: {Width: 11},
	}
	expected := []model.CIDWidth{
		model.CIDWidthArray{Start: 2, W: []model.Fl{10, 11, 5}},
		model.CIDWidthArray{Start: 6, W: []model.Fl{0}},
		model.CIDWidthArray{Start: 8, W: []model.Fl{13, 11}},
	}
	if got := cidWidths(input); !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}
