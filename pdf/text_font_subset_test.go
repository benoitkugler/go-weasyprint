package pdf

import (
	"bytes"
	_ "embed"
	"reflect"
	"testing"

	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/font/opentype/tables"
)

func gs(gids ...ot.GID) glyphSet {
	out := make(glyphSet)
	for _, gid := range gids {
		out.Add(gid)
	}
	return out
}

func comp(deps ...ot.GID) tables.Glyph {
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

// extracted from /usr/share/fonts/truetype/dejavu/DejaVuSans.ttf
//
//go:embed test/table_post_20.bin
var post20 []byte

func TestSubsetPost20(t *testing.T) {
	table, _, err := tables.ParsePost(post20)
	if err != nil {
		t.Fatal(err)
	}
	names := table.Names.(tables.PostNames20)
	numG := len(names.GlyphNameIndexes)
	if numG != 6253 {
		t.Fatal()
	}
	exp1 := names.Strings[names.GlyphNameIndexes[456]-258]
	exp2 := names.Strings[names.GlyphNameIndexes[2000]-258]
	exp3 := names.Strings[names.GlyphNameIndexes[4566]-258]
	set := gs(0, 1, 10, 155, 456, 4566, 2000)
	namesSubsetted := subsetPost20(names, set)
	if L := len(namesSubsetted.GlyphNameIndexes); L != len(set) {
		t.Fatalf("expected %d, got %d", len(set), L)
	}
	if !reflect.DeepEqual(namesSubsetted.Strings, []string{exp1, exp2, exp3}) {
		t.Fatal()
	}

	encoded := writePost(post20, names)
	if !bytes.Equal(encoded, post20) {
		t.Fatal()
	}
	// check the subsetted table is valid
	final := writePost(post20, namesSubsetted)
	table, _, err = tables.ParsePost(final)
	if err != nil {
		t.Fatal(err)
	}
	if len(table.Names.(tables.PostNames20).GlyphNameIndexes) != len(set) {
		t.Fatal()
	}
}
