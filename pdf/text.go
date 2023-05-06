package pdf

import (
	"crypto/md5"
	"encoding/hex"
	"sort"
	"strings"

	"github.com/benoitkugler/pdf/contentstream"
	pdfFonts "github.com/benoitkugler/pdf/fonts"
	"github.com/benoitkugler/pdf/fonts/cmaps"
	"github.com/benoitkugler/pdf/model"
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textprocessing/pango"
	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/matrix"
	drawText "github.com/benoitkugler/webrender/text/draw"
)

type pdfFont struct {
	*backend.Font
	*model.FontDict
}

func (g *group) SetTextPaint(op backend.PaintOp) {
	doFill := op&(backend.FillEvenOdd|backend.FillNonZero) != 0
	doStroke := op&(backend.Stroke) != 0

	var tr uint8
	if doFill && doStroke {
		tr = 2
	} else if doFill {
		tr = 0
	} else if doStroke {
		tr = 1
	} else {
		tr = 3
	}
	g.app.Ops(contentstream.OpSetTextRender{Render: tr})
}

// DrawText draws the given text using the current fill color.
func (g *group) DrawText(texts []backend.TextDrawing) {
	g.app.BeginText()
	defer g.app.EndText()

	for _, text := range texts {
		mat := matrix.New(text.FontSize, 0, 0, -text.FontSize, text.X, text.Y)
		if text.Angle != 0 { // avoid useless multiplication if angle == 0
			mat.RightMultBy(matrix.Rotation(text.Angle))
		}
		g.app.SetTextMatrix(mat.A, mat.B, mat.C, mat.D, mat.E, mat.F)

		for _, run := range text.Runs {
			pf := g.fonts[run.Font]
			g.app.SetFontAndSize(pdfFonts.BuiltFont{Meta: pf.FontDict}, 1)

			font := run.Font.GetHarfbuzzFont()

			var out []contentstream.SpacedGlyph
			for _, posGlyph := range run.Glyphs {
				out = append(out, contentstream.SpacedGlyph{
					SpaceSubtractedBefore: -int(posGlyph.Offset),
					GID:                   posGlyph.Glyph,
					SpaceSubtractedAfter:  posGlyph.Kerning,
				})

				// PDF readers don't support colored bitmap glyphs
				// so we have to add them as an image
				drawText.DrawEmoji(font, posGlyph.Glyph, pf.Extents[posGlyph.Glyph],
					text.FontSize, text.X, text.Y, posGlyph.XAdvance, g)

			}

			g.app.Ops(contentstream.OpShowSpaceGlyph{Glyphs: out})
		}
	}
}

func (f pdfFont) newFontDescriptor(font pango.Font, content *model.FontFile) model.FontDescriptor {
	desc := font.Describe(false)
	fontSize := desc.Size
	metrics := font.GetMetrics("")

	hash_ := md5.Sum([]byte(desc.String()))
	hash := string(hex.EncodeToString(hash_[:]))

	flags := model.Symbolic // since we use a custom char set
	if desc.Style != pango.STYLE_NORMAL {
		flags |= model.Italic
	}
	if strings.Contains(desc.FamilyName, "Serif") {
		flags |= model.Serif
	}
	if f.Font.IsFixedPitch() {
		flags |= model.FixedPitch
	}

	var ascent, descent fl
	if fontSize != 0 {
		ascent = fl(metrics.Ascent * 1000 / pango.Unit(fontSize))
		descent = fl(metrics.Descent * 1000 / pango.Unit(fontSize))
	}
	return model.FontDescriptor{
		FontName:    model.ObjName(hash + "+" + strings.ReplaceAll(desc.FamilyName, " ", "")),
		FontFamily:  desc.FamilyName,
		Flags:       flags,
		FontBBox:    model.Rectangle{Llx: fl(f.Font.Bbox[0]), Lly: fl(f.Font.Bbox[1]), Urx: fl(f.Font.Bbox[2]), Ury: fl(f.Font.Bbox[3])},
		ItalicAngle: 0,
		Ascent:      ascent,
		Descent:     descent,
		CapHeight:   fl(f.Font.Bbox[3]),
		StemV:       80,
		StemH:       80,
		FontFile:    content,
	}
}

// AddFont register a new font to be used in the output and return
// an object used to store associated metadata.
// This method will be called several times with the same `face` argument,
// so caching is advised.
func (g *group) AddFont(font pango.Font, content []byte) *backend.Font {
	// check the cache
	if ft, has := g.fonts[font]; has {
		return ft.Font
	}
	out := &backend.Font{
		Cmap:    make(map[fonts.GID][]rune),
		Extents: make(map[fonts.GID]backend.GlyphExtents),
	}
	// we only initialize the FontDict pointer,
	// which will be filled later in `writeFonts`
	g.fonts[font] = pdfFont{
		Font:     out,
		FontDict: &model.FontDict{},
	}

	// until then, we store the content
	face := font.GetHarfbuzzFont().Face()
	if g.fontFiles[face] == nil {
		fs := &model.FontFile{
			Stream: model.Stream{Content: content},
		}
		if face, ok := face.(*truetype.Font); ok {
			if isOpenType := face.Type == truetype.TypeOpenType; isOpenType {
				fs.Subtype = "OpenType"
			} else {
				fs.Length1 = len(content)
			}
		}
		g.fontFiles[face] = fs
	}

	return out
}

func cidWidths(dict map[fonts.GID]backend.GlyphExtents) []model.CIDWidth {
	var (
		widths       []model.CIDWidth
		keys         = make([]fonts.GID, 0, len(dict))
		currentBlock model.CIDWidthArray
	)

	for gid := range dict {
		keys = append(keys, gid)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	for _, gid := range keys {
		if _, has := dict[gid-1]; !has {
			// end the current block
			if len(currentBlock.W) != 0 {
				widths = append(widths, currentBlock)
			}
			currentBlock = model.CIDWidthArray{Start: model.CID(gid)}
		}
		currentBlock.W = append(currentBlock.W, dict[gid].Width)
	}

	if len(currentBlock.W) != 0 {
		widths = append(widths, currentBlock)
	}

	return widths
}

// post-process the font used
func (c *Output) writeFonts() {
	for pangoFont, font := range c.cache.fonts {
		if len(font.Cmap) == 0 {
			continue
		}
		content := c.cache.fontFiles[pangoFont.GetHarfbuzzFont().Face()]
		desc := font.newFontDescriptor(pangoFont, content)
		widths := cidWidths(font.Extents)

		font.FontDict.Subtype = model.FontType0{
			BaseFont: desc.FontName,
			Encoding: model.CMapEncodingPredefined("Identity-H"),
			DescendantFonts: model.CIDFontDictionary{
				Subtype:  "CIDFontType2",
				BaseFont: desc.FontName,
				CIDSystemInfo: model.CIDSystemInfo{
					Registry:   "Adobe",
					Ordering:   "Identity",
					Supplement: 0,
				},
				W:              widths,
				FontDescriptor: desc,
			},
		}
		cmap := cmaps.WriteAdobeIdentityUnicodeCMap(font.Cmap)
		font.FontDict.ToUnicode = &model.UnicodeCMap{Stream: model.Stream{Content: cmap}}
	}
}
