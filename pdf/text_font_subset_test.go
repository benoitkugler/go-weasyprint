package pdf

import (
	"reflect"
	"testing"

	"github.com/go-text/typesetting/opentype/api"
	"github.com/go-text/typesetting/opentype/tables"
)

func gs(gids ...api.GID) glyphSet {
	out := make(glyphSet)
	for _, gid := range gids {
		out.Add(gid)
	}
	return out
}

func comp(deps ...api.GID) tables.Glyph {
	var parts []tables.CompositeGlyphPart
	for _, dep := range deps {
		parts = append(parts, tables.CompositeGlyphPart{GlyphIndex: uint16(dep)})
	}
	return tables.Glyph{
		Data: tables.CompositeGlyph{
			Glyphs: parts,
		},
	}
}

func Test_handleComposite(t *testing.T) {
	sg := tables.Glyph{Data: tables.SimpleGlyph{}}
	tests := []struct {
		glyphs glyphSet
		glyf   tables.Glyf
		want   glyphSet
	}{
		{glyphSet{}, nil, glyphSet{}},
		{ // only simple glyphs
			gs(0, 1, 2, 3, 4),
			tables.Glyf{tables.Glyph{}, sg, sg, sg},
			gs(0, 1, 2, 3, 4),
		},
		{
			gs(0, 1),
			tables.Glyf{comp(2, 3), sg, sg, sg},
			gs(0, 1, 2, 3),
		},
		{
			gs(0),
			tables.Glyf{comp(1), comp(2, 3), sg, sg},
			gs(0, 1, 2, 3),
		},
		{ // with cycle
			gs(0),
			tables.Glyf{comp(0), sg, sg, sg},
			gs(0),
		},
	}
	for _, tt := range tests {
		handleComposite(tt.glyphs, tt.glyf)
		if !reflect.DeepEqual(tt.glyphs, tt.want) {
			t.Fatalf("expected %v, got %v", tt.want, tt.glyphs)
		}
	}
}
