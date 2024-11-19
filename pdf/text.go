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
	"github.com/benoitkugler/webrender/text"
	drawText "github.com/benoitkugler/webrender/text/draw"
	"github.com/go-text/typesetting/font"
	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/font/opentype/tables"
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
	g.stream.Ops(contentstream.OpSetTextRender{Render: tr})
}

// DrawText draws the given text using the current fill color.
func (g *group) DrawText(texts []backend.TextDrawing) {
	g.stream.BeginText()
	defer g.stream.EndText()

	for _, text := range texts {
		mat := text.Matrix()
		g.stream.SetTextMatrix(mat.A, mat.B, mat.C, mat.D, mat.E, mat.F)

		for _, run := range text.Runs {
			pf := g.fonts[run.Font]

			// do not use bitmap fonts
			useFont := g.cache.fontFiles[run.Font.Origin()].isSupported

			if useFont {
				g.stream.SetFontAndSize(pdfFonts.BuiltFont{Meta: pf.FontDict}, text.FontSize)
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
				g.stream.Ops(contentstream.OpShowSpaceGlyph{Glyphs: out})
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

// returns true if a valid 'glyf' , 'cff ' 'or'  tables is present
func isSupportedFont(content []byte) bool {
	ld, err := ot.NewLoader(bytes.NewReader(content))
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
	hasCff := ld.HasTable(ot.MustNewTag("CFF "))

	hasBitmap := ld.HasTable(ot.MustNewTag("EBDT")) && ld.HasTable(ot.MustNewTag("EBLC"))

	return hasValidGlyf || hasCff || hasBitmap
}

func newFontFile(fontDesc backend.FontDescription, font pdfFont, content []byte) *model.FontFile {
	fs := &model.FontFile{}
	if fontDesc.IsOpentype {
		// subset the font
		set := glyphSet{}
		for gid := range font.Cmap {
			set.Add(ot.GID(gid))
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

// TODO: see https://github.com/benoitkugler/webrender/issues/3
// // https://docs.microsoft.com/typography/opentype/spec/ebdt
// func buildBitmapFontDictionary( pdf, font, widths,
//                                   compress_pdf, subset) (font_dictionary model.FontType3) {
//     font_dictionary.FontBBox = model.Rectangle{0, 0, 1, 1}
//     font_dictionary.FontMatrix = model.Matrix{1, 0, 0, 1, 0, 0}
//     if subset{
//         chars = tuple(sorted(font.cmap))
//    } else{
//         chars = tuple([]int{256})}
//     first, last = chars[0], chars[-1]
//     font_dictionary.FirstChar = first
//     // font_dictionary.LastChar = last
//     var differences []int
//     // for index, index_widths in zip(widths[::2], widths[1::2]):
//     //     differences.append(index)
//     //     for i in range(len(index_widths)):
//     //         if i + index in chars:
//     //             differences.append(f'/{i + index}')
//     // font_dictionary.Encoding = pydyf.Dictionary({        'Type': '/Encoding',        'Differences': pydyf.Array(differences),    })
//     // char_procs = pydyf.Dictionary({})
//     font_glyphs = font.ttfont["EBDT"].strikeData[0]
//     widths := make([]int,   (last - first + 1))
//     glyphs_info = map[int]int{}
//     for key, glyph := range font_glyphs {
//         glyph_format = glyph.getFormat()
//         glyph_id = font.ttfont.getGlyphID(key)

//         // Get and store glyph metrics
//         if glyph_format == 5{
//             data = glyph.data
//             subtables = font.ttfont["EBLC"].strikes[0].indexSubTables
//             for subtable in subtables:
//                 first_index = subtable.firstGlyphIndex
//                 last_index = subtable.lastGlyphIndex
//                 if first_index <= glyph_id <= last_index:
//                     height = subtable.metrics.height
//                     advance = width = subtable.metrics.width
//                     bearing_x = subtable.metrics.horiBearingX
//                     bearing_y = subtable.metrics.horiBearingY
//                     break
//             else:
//                 LOGGER.warning(f"Unknown bitmap metrics for glyph: {glyph_id}")
//                 continue
//         }else{
//             data_start = 5 if glyph_format in (1, 2, 8) else 8
//             data = glyph.data[data_start:]
//             height, width = glyph.data[0:2]
//             bearing_x = int.from_bytes(glyph.data[2:3], "big", signed=True)
//             bearing_y = int.from_bytes(glyph.data[3:4], "big", signed=True)
//             advance = glyph.data[4]}
//         position_y = bearing_y - height
//         if glyph_id in chars{
//             widths[glyph_id - first] = advance}
//         stride = ceil(width / 8)
//         glyph_info = glyphs_info[glyph_id] = {
//             "width": width,
//             "height": height,
//             "x": bearing_x,
//             "y": position_y,
//             "stride": stride,
//             "bitmap": None,
//             "subglyphs": None,
//         }

//         // Decode bitmaps
//         if 0 in (width, height) or not data{
//             glyph_info["bitmap"] = b""}
//         elif glyph_format in (1, 6):
//             glyph_info["bitmap"] = data
//         elif glyph_format in (2, 5, 7):
//             padding = (8 - (width % 8)) % 8
//             bits = bin(int(data.hex(), 16))[2:]
//             bits = bits.zfill(8 * len(data))
//             bitmap_bits = ''.join(
//                 bits[i * width:(i + 1) * width] + padding * '0'
//                 for i in range(height))
//             glyph_info['bitmap'] = int(bitmap_bits, 2).to_bytes(
//                 height * stride, 'big')
//         elif glyph_format in (8, 9):
//             subglyphs = glyph_info['subglyphs'] = []
//             i = 0 if glyph_format == 9 else 1
//             number_of_components = int.from_bytes(data[i:i+2], 'big')
//             for j in range(number_of_components):
//                 index = (i + 2) + (j * 4)
//                 subglyph_id = int.from_bytes(data[index:index+2], 'big')
//                 x = int.from_bytes(data[index+2:index+3], 'big', signed=True)
//                 y = int.from_bytes(data[index+3:index+4], 'big', signed=True)
//                 subglyphs.append({'id': subglyph_id, 'x': x, 'y': y})
//         else:  // pragma: no cover
//             LOGGER.warning(f'Unsupported bitmap glyph format: {glyph_format}')
//             glyph_info['bitmap'] = bytes(height * stride)
// 			}
//     for glyph_id, glyph_info in glyphs_info.items():
//         // Don’t store glyph not in cmap
//         if glyph_id not in chars:
//             continue

//         // Draw glyph
//         stride = glyph_info['stride']
//         width = glyph_info['width']
//         height = glyph_info['height']
//         x = glyph_info['x']
//         y = glyph_info['y']
//         if glyph_info['bitmap'] is None:
//             length = height * stride
//             bitmap_int = int.from_bytes(bytes(length), 'big')
//             for subglyph in glyph_info['subglyphs']:
//                 sub_x = subglyph['x']
//                 sub_y = subglyph['y']
//                 sub_id = subglyph['id']
//                 if sub_id not in glyphs_info:
//                     LOGGER.warning(f'Unknown subglyph: {sub_id}')
//                     continue
//                 subglyph = glyphs_info[sub_id]
//                 if subglyph['bitmap'] is None:
//                     // TODO: support subglyph in subglyph
//                     LOGGER.warning(
//                         f'Unsupported subglyph in subglyph: {sub_id}')
//                     continue
//                 for row_y in range(subglyph['height']):
//                     row_slice = slice(
//                         row_y * subglyph['stride'],
//                         (row_y + 1) * subglyph['stride'])
//                     row = subglyph['bitmap'][row_slice]
//                     row_int = int.from_bytes(row, 'big')
//                     shift = stride * 8 * (height - sub_y - row_y - 1)
//                     stride_difference = stride - subglyph['stride']
//                     if stride_difference > 0:
//                         row_int <<= stride_difference * 8
//                     elif stride_difference < 0:
//                         row_int >>= -stride_difference * 8
//                     if sub_x > 0:
//                         row_int >>= sub_x
//                     elif sub_x < 0:
//                         row_int <<= -sub_x
//                     row_int %= 1 << stride * 8
//                     row_int <<= shift
//                     bitmap_int |= row_int
//             bitmap = bitmap_int.to_bytes(length, 'big')
//         else:
//             bitmap = glyph_info['bitmap']
//         bitmap_stream = pydyf.Stream([
//             b'0 0 d0',
//             f'{width} 0 0 {height} {x} {y} cm'.encode(),
//             b'BI',
//             b'/IM true',
//             b'/W', width,
//             b'/H', height,
//             b'/BPC 1',
//             b'/D [1 0]',
//             b'ID', bitmap, b'EI'
//         ], compress=compress_pdf)
//         pdf.add_object(bitmap_stream)
//         char_procs[glyph_id] = bitmap_stream.reference

//     pdf.add_object(char_procs)
//     font_dictionary['Widths'] = pydyf.Array(widths)
//     font_dictionary['CharProcs'] = char_procs.reference
// 		}

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
