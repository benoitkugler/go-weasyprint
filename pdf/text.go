package pdf

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/benoitkugler/pdf/contentstream"
	pdfFonts "github.com/benoitkugler/pdf/fonts"
	"github.com/benoitkugler/pdf/fonts/cmaps"
	"github.com/benoitkugler/pdf/model"
	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/matrix"
	"github.com/benoitkugler/webrender/text"
	drawText "github.com/benoitkugler/webrender/text/draw"
	"github.com/go-text/typesetting/opentype/api"
	"github.com/go-text/typesetting/opentype/api/font"
	"github.com/go-text/typesetting/opentype/loader"
	"github.com/go-text/typesetting/opentype/tables"
)

type pdfFont struct {
	*backend.FontChars
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

			// do not use bitmap fonts
			useFont := g.cache.fontFiles[run.Font.Origin()].isSupported

			if useFont {
				g.app.SetFontAndSize(pdfFonts.BuiltFont{Meta: pf.FontDict}, 1)
			}

			var out []contentstream.SpacedGlyph
			for _, posGlyph := range run.Glyphs {
				out = append(out, contentstream.SpacedGlyph{
					SpaceSubtractedBefore: -int(posGlyph.Offset),
					GID:                   posGlyph.Glyph,
					SpaceSubtractedAfter:  posGlyph.Kerning,
				})

				// PDF readers don't support colored bitmap glyphs
				// so we have to add them as an image
				drawText.DrawEmoji(run.Font, posGlyph.Glyph, pf.Extents[posGlyph.Glyph],
					text.FontSize, text.X, text.Y, posGlyph.XAdvance, g)
			}

			if useFont {
				g.app.Ops(contentstream.OpShowSpaceGlyph{Glyphs: out})
			}
		}
	}
}

func (f pdfFont) newFontDescriptor(font backend.Font, content *model.FontFile) model.FontDescriptor {
	desc := font.Description()

	hash_ := md5.Sum([]byte(fmt.Sprint(desc.Family,
		desc.Style, desc.Weight, desc.Size)))
	hash := string(hex.EncodeToString(hash_[:]))

	flags := model.Symbolic // since we use a custom char set
	if desc.Style != text.FSyNormal {
		flags |= model.Italic
	}
	if strings.Contains(desc.Family, "Serif") {
		flags |= model.Serif
	}
	if f.FontChars.IsFixedPitch() {
		flags |= model.FixedPitch
	}

	bbox := f.FontChars.Bbox
	return model.FontDescriptor{
		FontName:    model.ObjName(hash + "+" + strings.ReplaceAll(desc.Family, " ", "")),
		FontFamily:  desc.Family,
		Flags:       flags,
		FontBBox:    model.Rectangle{Llx: fl(bbox[0]), Lly: fl(bbox[1]), Urx: fl(bbox[2]), Ury: fl(bbox[3])},
		ItalicAngle: 0,
		Ascent:      desc.Ascent,
		Descent:     desc.Descent,
		CapHeight:   fl(bbox[3]),
		StemV:       80,
		StemH:       80,
		FontFile:    content,
	}
}

// AddFont register a new font to be used in the output and return
// an object used to store associated metadata.
// This method will be called several times with the same `face` argument,
// so caching is advised.
func (g *group) AddFont(font backend.Font, content []byte) *backend.FontChars {
	// check the cache
	if ft, has := g.fonts[font]; has {
		return ft.FontChars
	}
	out := &backend.FontChars{
		Cmap:    make(map[backend.GID][]rune),
		Extents: make(map[backend.GID]backend.GlyphExtents),
	}
	// we only initialize the FontDict pointer,
	// which will be filled later in `writeFonts`
	g.fonts[font] = pdfFont{
		FontChars: out,
		FontDict:  &model.FontDict{},
	}

	origin := font.Origin()
	// until then, we store the content
	if _, ok := g.fontFiles[origin]; !ok {
		g.fontFiles[origin] = fontContent{
			content:     content,
			isSupported: isSupportedFont(content),
		}
	}

	return out
}

func cidWidths(dict map[backend.GID]backend.GlyphExtents) []model.CIDWidth {
	var (
		widths       []model.CIDWidth
		keys         = make([]backend.GID, 0, len(dict))
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
		currentBlock.W = append(currentBlock.W, model.Fl(dict[gid].Width))
	}

	if len(currentBlock.W) != 0 {
		widths = append(widths, currentBlock)
	}

	return widths
}

// returns true if a valid 'glyf' or 'cff ' tables is present
func isSupportedFont(content []byte) bool {
	ld, err := loader.NewLoader(bytes.NewReader(content))
	if err != nil {
		return false
	}

	headT, _, err := font.LoadHeadTable(ld, nil)
	if err != nil {
		return false
	}
	maxp, err := ld.RawTable(maxpTag)
	if err != nil {
		return false
	}
	maxpT, _, err := tables.ParseMaxp(maxp)
	if err != nil {
		return false
	}

	var hasValidGlyf bool
	if ld.HasTable(locaTag) && ld.HasTable(glyfTag) {
		// load 'locaT' and 'glyf' tables
		locaT, _ := ld.RawTable(locaTag)

		isLong := headT.IndexToLocFormat == 1
		loca, _ := tables.ParseLoca(locaT, int(maxpT.NumGlyphs), isLong)
		glyfRaw, _ := ld.RawTable(glyfTag)
		glyf, _ := tables.ParseGlyf(glyfRaw, loca)
		hasValidGlyf = len(glyf) > 0
	}
	hasCff := ld.HasTable(loader.MustNewTag("CFF "))
	return hasValidGlyf || hasCff
}

func newFontFile(fontDesc backend.FontDescription, font pdfFont, content []byte) *model.FontFile {
	fs := &model.FontFile{}
	if fontDesc.IsOpentype {
		// subset the font
		set := glyphSet{}
		for gid := range font.Cmap {
			set.Add(api.GID(gid))
		}
		contentS, err := subset(bytes.NewReader(content), set)
		if err != nil {
			log.Printf("font subsetting failed: %s", err)
		} else {
			content = contentS
		}

		if fontDesc.IsOpentypeOpentype {
			fs.Subtype = "OpenType"
		} else {
			fs.Length1 = len(content)
		}
	}
	fs.Stream = model.Stream{Content: content}
	return fs
}

// post-process the font used
func (c *Output) writeFonts() {
	for bFont, font := range c.cache.fonts {
		if len(font.Cmap) == 0 {
			continue
		}

		content := c.cache.fontFiles[bFont.Origin()]
		// PDF readers do not support bitmap fonts
		if !content.isSupported {
			continue
		}

		fs := newFontFile(bFont.Description(), font, content.content)
		desc := font.newFontDescriptor(bFont, fs)
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
