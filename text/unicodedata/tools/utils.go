package tools

import (
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

func FetchData(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return string(data)
}

func convertHexa(s string) uint32 {
	i, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		log.Fatal(err)
	}
	return uint32(i)
}

type Rg struct {
	Start, End uint32
}

func SortRanges(l []Rg) {
	sort.Slice(l, func(i, j int) bool {
		return l[i].End < l[j].End
	})
	sort.SliceStable(l, func(i, j int) bool {
		return l[i].Start < l[j].Start
	})
}

func Parse(data string, outRanges map[string][]Rg) {
	lines := strings.Split(data, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' { // reading header or comment
			continue
		}
		parts := strings.Split(strings.Split(line, "#")[0], ";")[:2]
		if len(parts) != 2 {
			log.Fatalf("expected 2 parts, got %s", line)
		}
		rang, typ := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		rangS := strings.Split(rang, "..")
		start := convertHexa(rangS[0])
		end := start
		if len(rangS) > 1 {
			end = convertHexa(rangS[1])
		}
		if L := len(outRanges[typ]); L != 0 && outRanges[typ][L-1].End == start-1 {
			outRanges[typ][L-1].End = end
		} else {
			outRanges[typ] = append(outRanges[typ], Rg{Start: start, End: end})
		}
	}
}
