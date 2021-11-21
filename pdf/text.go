package pdf

import (
	"crypto/md5"
	"encoding/hex"
	"sort"
	"strings"

	"github.com/benoitkugler/go-weasyprint/backend"
	"github.com/benoitkugler/pdf/contentstream"
	pdfFonts "github.com/benoitkugler/pdf/fonts"
	"github.com/benoitkugler/pdf/fonts/cmaps"
	"github.com/benoitkugler/pdf/model"
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/pango"
)

type pdfFont struct {
	*backend.Font
	*model.FontDict
}

// DrawText draws the given text using the current fill color.
func (g *group) DrawText(text backend.TextDrawing) {
	g.app.BeginText()
	defer g.app.EndText()

	g.app.SetTextMatrix(text.FontSize, 0, 0, -text.FontSize, text.X, text.Y)

	for _, run := range text.Runs {
		pf := g.fonts[run.Font]
		g.app.SetFontAndSize(pdfFonts.BuiltFont{Meta: pf.FontDict}, 1)

		var out []contentstream.SpacedGlyph
		for _, g := range run.Glyphs {
			out = append(out, contentstream.SpacedGlyph{
				SpaceSubtractedBefore: -int(g.Offset),
				GID:                   g.Glyph,
				SpaceSubtractedAfter:  g.Kerning,
			})
		}

		g.app.Ops(contentstream.OpShowSpaceGlyph{Glyphs: out})
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

	return model.FontDescriptor{
		FontName:    model.ObjName(hash + "+" + strings.ReplaceAll(desc.FamilyName, " ", "")),
		FontFamily:  desc.FamilyName,
		Flags:       flags,
		FontBBox:    model.Rectangle{Llx: fl(f.Font.Bbox[0]), Lly: fl(f.Font.Bbox[1]), Urx: fl(f.Font.Bbox[2]), Ury: fl(f.Font.Bbox[3])},
		ItalicAngle: 0,
		Ascent:      fl(metrics.Ascent * 1000 / pango.GlyphUnit(fontSize)),
		Descent:     fl(metrics.Descent * 1000 / pango.GlyphUnit(fontSize)),
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
		Cmap:   make(map[fonts.GID][]rune),
		Widths: make(map[fonts.GID]int),
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

func cidWidths(dict map[fonts.GID]int) []model.CIDWidth {
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
		currentBlock.W = append(currentBlock.W, dict[gid])
	}

	if len(currentBlock.W) != 0 {
		widths = append(widths, currentBlock)
	}

	return widths
}

// post-process the font used
func (c *Output) writeFonts() {
	for pangoFont, font := range c.cache.fonts {
		content := c.cache.fontFiles[pangoFont.GetHarfbuzzFont().Face()]
		desc := font.newFontDescriptor(pangoFont, content)
		widths := cidWidths(font.Widths)

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