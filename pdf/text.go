package pdf

import (
	"github.com/benoitkugler/go-weasyprint/backend"
	"github.com/benoitkugler/pdf/contentstream"
	pdfFonts "github.com/benoitkugler/pdf/fonts"
	"github.com/benoitkugler/pdf/model"
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/pango"
)

type font struct {
	*backend.Font
	*model.FontDict
}

// DrawText draws the given text using the current fill color.
func (g *group) DrawText(text backend.TextDrawing) {
	g.app.BeginText()
	defer g.app.EndText()

	g.app.SetTextMatrix(text.FontSize, 0, 0, -text.FontSize, text.X, text.Y)

	for _, run := range text.Runs {
		pdfFont := g.fonts[run.Font]
		g.app.SetFontAndSize(pdfFonts.BuiltFont{Meta: pdfFont.FontDict}, 1)

		var out []contentstream.SpacedGlyph
		for _, g := range run.Glyphs {
			out = append(out, contentstream.SpacedGlyph{
				SpaceSubtractedBefore: -int(g.Offset),
				GID:                   g.Glyph.GID(),
				SpaceSubtractedAfter:  g.Kerning,
			})
		}

		g.app.Ops(contentstream.OpShowSpaceGlyph{Glyphs: out})
	}
}

// TODO:
func newPDFFontFromFace(face fonts.Face) *model.FontDict {
	return &model.FontDict{}
}

// AddFont register a new font to be used in the output and return
// an object used to store associated metadata.
// This method will be called several times with the same `face` argument,
// so caching is advised.
func (g *group) AddFont(face fonts.Face, content []byte) *backend.Font {
	// check the cache
	if ft, has := g.fonts[face]; has {
		return ft.Font
	}
	out := &backend.Font{
		Cmap:   make(map[pango.Glyph][]rune),
		Widths: make(map[pango.Glyph]int),
	}
	g.fonts[face] = font{
		Font:     out,
		FontDict: newPDFFontFromFace(face),
	}
	return out
}

// post-process the font used
func (c *Output) writeFonts() {
	// for _, font := range c.fonts {
	// 	switch fontType := font.FontDict.Subtype.(type) {
	// 	case model.FontType0:
	// 	case model.FontType1:
	// 		fontType.Widths
	// 	}
	// }
}
