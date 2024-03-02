package pdf

import (
	"encoding/binary"
	"fmt"

	"github.com/go-text/typesetting/opentype/api"
	"github.com/go-text/typesetting/opentype/api/font"
	"github.com/go-text/typesetting/opentype/loader"
	"github.com/go-text/typesetting/opentype/tables"
)

// basic font subset

var (
	maxpTag = loader.MustNewTag("maxp")
	locaTag = loader.MustNewTag("loca")
	glyfTag = loader.MustNewTag("glyf")

	fvarTag = loader.MustNewTag("fvar")
	avarTag = loader.MustNewTag("avar")
	MVARTag = loader.MustNewTag("MVAR")
	gvarTag = loader.MustNewTag("gvar")
	HVARTag = loader.MustNewTag("HVAR")
	VVARTag = loader.MustNewTag("VVAR")
)

func isVar(table loader.Tag) bool {
	switch table {
	case fvarTag, avarTag, MVARTag, gvarTag, HVARTag, VVARTag:
		return true
	default:
		return false
	}
}

type glyphSet map[api.GID]struct{}

func (gs glyphSet) Add(g api.GID) { gs[g] = struct{}{} }

// Variables font are not supported : the variable tables will be dropped
// TODO: For now, [subset] only supports the 'glyf' table
func subset(input loader.Resource, glyphs glyphSet) ([]byte, error) {
	ld, err := loader.NewLoader(input)
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
		glyfNew, err = ld.RawTable(glyfTag)
		if err != nil {
			return nil, fmt.Errorf("subsetting failed: %s", err)
		}

		glyfNew = subsetGlyf(loca, glyfNew, glyphs)
		locaNew = writeLoca(loca, isLong)
	}

	origTablesTags := ld.Tables()
	tables := make([]loader.Table, 0, len(origTablesTags))
	for _, tag := range origTablesTags {
		if isVar(tag) {
			continue
		}

		table := loader.Table{Tag: tag}

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

	return loader.WriteTTF(tables), nil
}

// mutate [loca] and [glyph] and return [glyf]
func subsetGlyf(loca []uint32, glyf []byte, glyphs glyphSet) []byte {
	// trim the unused glyp data, and adjust the offset to have start == end
	currentStart := loca[0]
	nbGlyphs := len(loca) - 1
	for gid := api.GID(0); gid < api.GID(nbGlyphs); gid++ {
		origStart, origEnd := loca[gid], loca[gid+1]
		// start and end offsets will change, but not the length
		glyphLength := origEnd - origStart

		newStart := currentStart
		var newEnd uint32
		if _, has := glyphs[gid]; has {
			// keep the data
			newEnd = newStart + glyphLength
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
	return glyf
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
