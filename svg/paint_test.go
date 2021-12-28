package svg

import (
	"reflect"
	"testing"

	"github.com/benoitkugler/go-weasyprint/style/parser"
)

func Test_newPainter(t *testing.T) {
	tests := []struct {
		args    string
		want    painter
		wantErr bool
	}{
		{
			"red",
			painter{"", parser.ColorKeywords["red"].RGBA, true},
			false,
		},
		{
			"",
			painter{"", parser.RGBA{}, false},
			false,
		},
		{
			"none",
			painter{"", parser.RGBA{}, false},
			false,
		},
		{
			"black",
			painter{"", parser.RGBA{A: 1}, true},
			false,
		},
		{
			"url(ddd",
			painter{},
			true,
		},
		{
			"url(#myPaint)",
			painter{"myPaint", parser.RGBA{}, true},
			false,
		},
		{
			"url(#myPaint)  green",
			painter{"myPaint", parser.ColorKeywords["green"].RGBA, true},
			false,
		},
	}
	for _, tt := range tests {
		got, err := newPainter(tt.args)
		if (err != nil) != tt.wantErr {
			t.Errorf("newPainter() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("newPainter() = %v, want %v", got, tt.want)
		}
	}
}
