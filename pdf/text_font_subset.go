package pdf

import (
	"encoding/binary"
	"fmt"

	"github.com/go-text/typesetting/font"
	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/font/opentype/tables"
)

// basic font subset

var (
	maxpTag = ot.MustNewTag("maxp")
	locaTag = ot.MustNewTag("loca")
	glyfTag = ot.MustNewTag("glyf")

	fvarTag = ot.MustNewTag("fvar")
	avarTag = ot.MustNewTag("avar")
	mVARTag = ot.MustNewTag("MVAR")
	gvarTag = ot.MustNewTag("gvar")
	hVARTag = ot.MustNewTag("HVAR")
	vVARTag = ot.MustNewTag("VVAR")

	gSUBTag = ot.MustNewTag("GSUB")
	gPOSTag = ot.MustNewTag("GPOS")
)

// several tables are not useful in PDF
func ignoreTable(table ot.Tag) bool {
	switch table {
	case fvarTag, avarTag, mVARTag, gvarTag, hVARTag, vVARTag, gSUBTag, gPOSTag:
		return true
	default:
		return false
	}
}

type glyphSet map[ot.GID]struct{}

func (gs glyphSet) Add(g ot.GID) { gs[g] = struct{}{} }

// recursively fetch composite glyphs deps, adding it to the set
func handleComposite(glyphs glyphSet, glyf tables.Glyf) {
	queue := make([]ot.GID, 0, len(glyphs))
	// start with the given glyphs
	for g := range glyphs {
		queue = append(queue, g)
	}
	for len(queue) != 0 {
		// pop the last glyph
		g := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if int(g) >= len(glyf) { // invalid glyph
			continue
		}
		switch data := glyf[g].Data.(type) {
		case tables.SimpleGlyph:
			// simply add it to the final set
			glyphs.Add(g)
		case tables.CompositeGlyph:
			// fetch deps
			for _, part := range data.Glyphs {
				dep := ot.GID(part.GlyphIndex)
				// if not already seen, add it to the queue
				if _, ok := glyphs[dep]; !ok {
					glyphs.Add(dep)
					queue = append(queue, dep)
				}
				// else : already processed, continue
			}
		default:
			// nil : nothing to do
		}
	}
}

// Variables font are not supported : the variable tables will be dropped
// TODO: For now, [subset] only supports the 'glyf' table
func subset(input ot.Resource, glyphs glyphSet) ([]byte, error) {
	ld, err := ot.NewLoader(input)
	if err != nil {
		return nil, fmt.Errorf("subsetting failed: %s", err)
	}

	headT, _, err := font.LoadHeadTable(ld, nil)
	if err != nil {
		return nil, fmt.Errorf("subsetting failed: %s", err)
	}
	maxp, err := ld.RawTable(maxpTag)
	if err != nil {
		return nil, fmt.Errorf("subsetting failed: %s", err)
	}
	maxpT, _, err := tables.ParseMaxp(maxp)
	if err != nil {
		return nil, fmt.Errorf("subsetting failed: %s", err)
	}

	var glyfNew, locaNew []byte
	if ld.HasTable(locaTag) && ld.HasTable(glyfTag) {
		// load 'locaT' and 'glyf' tables
		locaT, err := ld.RawTable(locaTag)
		if err != nil {
			return nil, fmt.Errorf("subsetting failed: %s", err)
		}

		isLong := headT.IndexToLocFormat == 1

		loca, err := tables.ParseLoca(locaT, int(maxpT.NumGlyphs), isLong)
		if err != nil {
			return nil, fmt.Errorf("subsetting failed: %s", err)
		}
		glyfRaw, err := ld.RawTable(glyfTag)
		if err != nil {
			return nil, fmt.Errorf("subsetting failed: %s", err)
		}

		glyf, err := tables.ParseGlyf(glyfRaw, loca)
		if err != nil {
			return nil, fmt.Errorf("subsetting failed: %s", err)
		}
		// handle composite glyph
		handleComposite(glyphs, glyf)

		loca, glyfNew = subsetGlyf(loca, glyfRaw, glyphs)
		locaNew = writeLoca(loca, isLong)
	}

	origTablesTags := ld.Tables()
	tables := make([]ot.Table, 0, len(origTablesTags))
	for _, tag := range origTablesTags {
		if ignoreTable(tag) {
			continue
		}

		table := ot.Table{Tag: tag}

		if tag == glyfTag && glyfNew != nil {
			table.Content = glyfNew
		} else if tag == locaTag && locaNew != nil {
			table.Content = locaNew
		} else {
			table.Content, err = ld.RawTable(tag)
			if err != nil {
				return nil, fmt.Errorf("subsetting failed: %s", err)
			}
		}

		tables = append(tables, table)
	}

	return ot.WriteTTF(tables), nil
}

// mutate and returns [loca] and [glyph]
func subsetGlyf(loca []uint32, glyf []byte, glyphs glyphSet) ([]uint32, []byte) {
	// trim the unused glyph data, and adjust the offset to have start == end
	currentStart := loca[0]
	nbGlyphs := len(loca) - 1
	maxGlyphUsed := ot.GID(0)
	for gid := ot.GID(0); gid < ot.GID(nbGlyphs); gid++ {
		origStart, origEnd := loca[gid], loca[gid+1]
		// start and end offsets will change, but not the length
		glyphLength := origEnd - origStart

		newStart := currentStart
		var newEnd uint32
		if _, has := glyphs[gid]; has {
			// keep the data
			newEnd = newStart + glyphLength

			if gid > maxGlyphUsed {
				maxGlyphUsed = gid
			}
		} else {
			// set the length to 0
			newEnd = newStart
		}

		// avoid useless copy that will happen until we skip a glyph
		if newStart != origStart {
			copy(glyf[newStart:], glyf[origStart:origEnd])
		}

		loca[gid] = newStart

		// update the offset
		currentStart = newEnd
	}
	// update the final offset
	loca[nbGlyphs] = currentStart
	glyf = glyf[:currentStart]

	// trim the local table
	loca = loca[:maxGlyphUsed+2]

	return loca, glyf
}

// writeLoca performs the reverse operation implemented by [ParseLoca]
func writeLoca(offsets []uint32, isLong bool) []byte {
	if isLong {
		out := make([]byte, 4*len(offsets))
		for i, off := range offsets {
			binary.BigEndian.PutUint32(out[4*i:], off)
		}
		return out
	} else {
		out := make([]byte, 2*len(offsets))
		for i, off := range offsets {
			binary.BigEndian.PutUint16(out[2*i:], uint16(off>>1))
		}
		return out
	}
}
