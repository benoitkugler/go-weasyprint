package fribidi

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

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

func parseLevels(line string) ([]Level, error) {
	fields := strings.Fields(line)
	out := make([]Level, len(fields))
	for i, f := range fields {
		if f == "x" {
			out[i] = -1
		} else {
			lev, err := strconv.Atoi(f)
			if err != nil {
				return nil, fmt.Errorf("invalid level %s: %s", f, err)
			}
			out[i] = Level(lev)
		}
	}
	return out, nil
}

type testData struct {
	codePoints       []rune
	parDir           int
	resolvedParLevel int
	levels           []Level
	visualOrdering   []int
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

func TestBidiCharacters(t *testing.T) {
	const filename = "test/unicode-conformance/BidiCharacterTest.txt"

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	for lineNumber, line := range bytes.Split(b, []byte{'\n'}) {

		if len(line) == 0 || line[0] == '#' || line[0] == '\n' {
			continue
		}

		lineData, err := parseTestLine(line)
		if err != nil {
			t.Fatalf("invalid line %d: %s", lineNumber+1, err)
		}

		bracketTypes := make([]BracketType, len(lineData.codePoints))
		types := make([]CharType, len(lineData.codePoints))

		for i, c := range lineData.codePoints {
			types[i] = GetBidiType(c)

			// Note the optimization that a bracket is always of type neutral
			if types[i] == ON {
				bracketTypes[i] = GetBracket(c)
			} else {
				bracketTypes[i] = NoBracket
			}
		}

		var baseDir ParType
		switch lineData.parDir {
		case 0:
			baseDir = LTR
		case 1:
			baseDir = RTL
		case 2:
			baseDir = ON
		}

		levels, _ := fribidi_get_par_embedding_levels_ex(types, bracketTypes, &baseDir)

		ltor := make([]int, len(lineData.codePoints))
		for i := range ltor {
			ltor[i] = i
		}

		fribidi_reorder_line(0 /*FRIBIDI_FLAG_REORDER_NSM*/, types, len(types), 0, baseDir, levels, nil, ltor)

		j := 0
		for _, lr := range ltor {
			if !types[lr].IsExplicitOrBn() {
				ltor[j] = lr
				j++
			}
		}
		ltor = ltor[0:j] // slice to length

		/* Compare */
		for i, level := range levels {
			if exp := lineData.levels[i]; level != exp && exp != -1 {
				t.Fatalf("failure on line %d: levels[%d]: expected %d, got %d", lineNumber+1, i, exp, level)
				break
			}
		}

		if !reflect.DeepEqual(ltor, lineData.visualOrdering) {
			t.Fatalf("failure on line %d: visual ordering: got %v, expected %v", lineNumber+1, ltor, lineData.visualOrdering)
		}
	}
}

func parseCharType(s string) (CharType, error) {
	switch s {
	case "L":
		return LTR, nil
	case "R":
		return RTL, nil
	case "AL":
		return AL, nil
	case "EN":
		return EN, nil
	case "AN":
		return AN, nil
	case "ES":
		return ES, nil
	case "ET":
		return ET, nil
	case "CS":
		return CS, nil
	case "NSM":
		return NSM, nil
	case "BN":
		return BN, nil
	case "B":
		return BS, nil
	case "S":
		return SS, nil
	case "WS":
		return WS, nil
	case "ON":
		return ON, nil
	case "LRE":
		return LRE, nil
	case "RLE":
		return RLE, nil
	case "LRO":
		return LRO, nil
	case "RLO":
		return RLO, nil
	case "PDF":
		return PDF, nil
	case "LRI":
		return LRI, nil
	case "RLI":
		return RLI, nil
	case "FSI":
		return FSI, nil
	case "PDI":
		return PDI, nil
	default:
		return 0, fmt.Errorf("invalid char type %s", s)
	}
}

func parse_levels_line(line string) ([]Level, error) {
	line = strings.TrimPrefix(line, "@Levels:")
	return parseLevels(line)
}

func parse_reorder_line(line string) ([]int, error) {
	line = strings.TrimPrefix(line, "@Reorder:")
	return parseOrdering(line)
}

func parse_test_line(line string) ([]CharType, int, error) {
	fields := strings.Split(line, ";")
	if len(fields) != 2 {
		return nil, 0, fmt.Errorf("invalid line: %s", line)
	}
	var err error
	chars := strings.Fields(fields[0])
	out := make([]CharType, len(chars))
	for i, cs := range chars {
		out[i], err = parseCharType(cs)
		if err != nil {
			return nil, 0, err
		}
	}
	baseDirFlags, err := strconv.Atoi(strings.TrimSpace(fields[1]))
	return out, baseDirFlags, err
}

func TestBidi(t *testing.T) {
	const filename = "test/unicode-conformance/BidiTest.txt"

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	var (
		expected_ltor   []int
		expected_levels []Level
	)
	for lineNumber, lineB := range bytes.Split(b, []byte{'\n'}) {
		line := string(lineB)
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		if strings.HasPrefix(line, "@Reorder:") {
			expected_ltor, err = parse_reorder_line(line)
			if err != nil {
				t.Fatalf("invalid  line %d: %s", lineNumber+1, err)
			}
			continue
		} else if strings.HasPrefix(line, "@Levels:") {
			expected_levels, err = parse_levels_line(line)
			if err != nil {
				t.Fatalf("invalid line %d: %s", lineNumber+1, err)
			}
			continue
		}

		/* Test line */
		types, base_dir_flags, err := parse_test_line(line)
		if err != nil {
			t.Fatalf("invalid line %d: %s", lineNumber+1, err)
		}

		/* Test it */
		for base_dir_mode := 0; base_dir_mode < 3; base_dir_mode++ {

			if (base_dir_flags & (1 << base_dir_mode)) == 0 {
				continue
			}

			var base_dir ParType
			switch base_dir_mode {
			case 0:
				base_dir = ON
			case 1:
				base_dir = LTR
			case 2:
				base_dir = RTL
			}

			// Brackets are not used in the BidiTest.txt file
			levels, _ := fribidi_get_par_embedding_levels_ex(types, nil, &base_dir)

			ltor := make([]int, len(levels))
			for i := range ltor {
				ltor[i] = i
			}

			fribidi_reorder_line(0 /*FRIBIDI_FLAG_REORDER_NSM*/, types, len(types),
				0, base_dir, levels,
				nil, ltor)

			j := 0
			for _, lr := range ltor {
				if !types[lr].IsExplicitOrBn() {
					ltor[j] = lr
					j++
				}
			}
			ltor = ltor[0:j] // slice to length

			/* Compare */
			for i, level := range levels {
				if exp := expected_levels[i]; level != exp && exp != -1 {
					t.Fatalf("failure on line %d: levels[%d]: expected %d, got %d", lineNumber+1, i, exp, level)
					break
				}
			}

			if !reflect.DeepEqual(ltor, expected_ltor) {
				t.Fatalf("failure on line %d: visual ordering: got %v, expected %v", lineNumber+1, ltor, expected_ltor)
			}
		}
	}
}
