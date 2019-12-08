package main

import (
	"fmt"
	"unicode"

	"github.com/benoitkugler/go-weasyprint/text/unicodedata/tools"
)

const (
	url    = "https://www.unicode.org/Public/emoji/12.0/emoji-data.txt"
	format = "\t\t{Lo:0x%04x, Hi:0x%04x, Stride:1},\n"
)

func genereRanges(raws []tools.Rg) (r16 []unicode.Range16, r32 []unicode.Range32) {
	tools.SortRanges(raws)
	for _, r := range raws {
		if r.End >= 1<<16 {
			r32 = append(r32, unicode.Range32{Lo: r.Start, Hi: r.End, Stride: 1})
		} else {
			r16 = append(r16, unicode.Range16{Lo: uint16(r.Start), Hi: uint16(r.End), Stride: 1})
		}
	}
	return
}

func main() {
	s := tools.FetchData(url)
	ranges := map[string][]tools.Rg{}
	tools.Parse(s, ranges)

	out := `
	// Generated by makeemojis.go
	// DO NOT EDIT

	package unicodedata
	
	// data from : ` + url + "\n\n"

	for typ, s := range ranges {
		if typ != "Emoji" && typ != "Emoji_Presentation" && typ != "Emoji_Modifier" && typ != "Emoji_Modifier_Base" && typ != "Extended_Pictographic" {
			continue
		}
		r16, r32 := genereRanges(s)
		out += fmt.Sprintf("var _Pango%sTable = &unicode.RangeTable {\n", typ)
		if len(r16) > 0 {
			out += "R16: []unicode.Range16{\n"
			for _, r1 := range r16 {
				out += fmt.Sprintf(format, r1.Lo, r1.Hi)
			}
			out += "},\n"
		}
		if len(r32) > 0 {
			out += "R32: []unicode.Range32{\n"
			for _, r3 := range r32 {
				out += fmt.Sprintf(format, r3.Lo, r3.Hi)
			}
			out += "},\n"
		}
		out += "}\n"
	}

	fmt.Println(out)
}
