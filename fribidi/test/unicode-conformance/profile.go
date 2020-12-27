package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"

	"github.com/benoitkugler/go-weasyprint/fribidi"
)

type testData struct {
	codePoints       []rune
	parDir           int
	resolvedParLevel int
	levels           []fribidi.Level
	visualOrdering   []int
}

func parseOrdering(line string) ([]int, error) {
	fields := strings.Fields(line)
	out := make([]int, len(fields))
	for i, posLit := range fields {
		pos, err := strconv.Atoi(posLit)
		if err != nil {
			return nil, fmt.Errorf("invalid position %s: %s", posLit, err)
		}
		out[i] = pos
	}
	return out, nil
}

func parseLevels(line string) ([]fribidi.Level, error) {
	fields := strings.Fields(line)
	out := make([]fribidi.Level, len(fields))
	for i, f := range fields {
		if f == "x" {
			out[i] = -1
		} else {
			lev, err := strconv.Atoi(f)
			if err != nil {
				return nil, fmt.Errorf("invalid level %s: %s", f, err)
			}
			out[i] = fribidi.Level(lev)
		}
	}
	return out, nil
}

func parseTestLine(line []byte) (out testData, err error) {
	fields := strings.Split(string(line), ";")
	if len(fields) < 5 {
		return out, fmt.Errorf("invalid line %s", line)
	}

	//  Field 0. Code points
	for _, runeLit := range strings.Fields(fields[0]) {
		var c rune
		if _, err := fmt.Sscanf(runeLit, "%04x", &c); err != nil {
			return out, fmt.Errorf("invalid rune %s: %s", runeLit, err)
		}
		out.codePoints = append(out.codePoints, c)
	}

	// Field 1. Paragraph direction
	out.parDir, err = strconv.Atoi(fields[1])
	if err != nil {
		return out, fmt.Errorf("invalid paragraph direction %s: %s", fields[1], err)
	}

	// Field 2. resolved paragraph_dir
	out.resolvedParLevel, err = strconv.Atoi(fields[2])
	if err != nil {
		return out, fmt.Errorf("invalid resolved paragraph embedding level %s: %s", fields[2], err)
	}

	// Field 3. resolved levels (or -1)
	out.levels, err = parseLevels(fields[3])
	if err != nil {
		return out, err
	}

	if len(out.levels) != len(out.codePoints) {
		return out, errors.New("different lengths for levels and codepoints")
	}

	//  Field 4 - resulting visual ordering
	out.visualOrdering, err = parseOrdering(fields[4])

	return out, err
}

func parse() ([]testData, error) {
	const filename = "BidiCharacterTest.txt"

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var out []testData
	for lineNumber, line := range bytes.Split(b, []byte{'\n'}) {
		if len(line) == 0 || line[0] == '#' || line[0] == '\n' {
			continue
		}

		lineData, err := parseTestLine(line)
		if err != nil {
			return nil, fmt.Errorf("invalid line %d: %s", lineNumber+1, err)
		}
		out = append(out, lineData)
	}
	return out, nil
}

func main() {
	datas, err := parse()
	if err != nil {
		log.Fatal(err)
	}
	runtime.GC() // get up-to-date statistics

	f, err := os.Create("cpuprofile.prof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	defer f.Close() // error handling omitted for example
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	// ... rest of the program ...
	for range [4]int{} {
		for _, lineData := range datas {

			bracketTypes := make([]fribidi.BracketType, len(lineData.codePoints))
			types := make([]fribidi.CharType, len(lineData.codePoints))

			for i, c := range lineData.codePoints {
				types[i] = fribidi.GetBidiType(c)

				// Note the optimization that a bracket is always of type neutral
				if types[i] == fribidi.ON {
					bracketTypes[i] = fribidi.GetBracket(c)
				} else {
					bracketTypes[i] = fribidi.NoBracket
				}
			}

			var baseDir fribidi.ParType
			switch lineData.parDir {
			case 0:
				baseDir = fribidi.LTR
			case 1:
				baseDir = fribidi.RTL
			case 2:
				baseDir = fribidi.ON
			}

			levels, _ := fribidi.GetParEmbeddingLevels(types, bracketTypes, &baseDir)

			ltor := make([]int, len(lineData.codePoints))
			for i := range ltor {
				ltor[i] = i
			}

			fribidi.ReorderLine(0 /*FRIBIDI_FLAG_REORDER_NSM*/, types, len(types), 0, baseDir, levels, nil, ltor)
		}
	}

	f, err = os.Create("memprofile.prof")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer f.Close() // error handling omitted for example
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
}
