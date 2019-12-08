// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2013 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// Unicode table generator. Based on unicode/maketables.go.
// Data read from the web.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const (
	ClassPrefix   = "G_UNICODE_BREAK_"
	ClassTypeName = "GUnicodeBreakType"
)

func main() {
	flag.Parse()
	if *testData {
		printHeader()
		printTests()
	} else {
		printHeader()
		printClasses()
		printTables()
		printSizes()
	}
}

var (
	dataURL = flag.String("data", "",
		"full URL for LineBreak.txt; defaults to --url/LineBreak.txt")
	testData = flag.Bool("test", false,
		"Print the test tables.")
	testDataURL = flag.String("testdata", "",
		"full URL for LineBreakTest.txt; defaults to --url/auxiliary/LineBreakTest.txt")
	url = flag.String("url",
		"http://www.unicode.org/Public/6.2.0/ucd/",
		"URL of Unicode database directory")
	excludeclasses = flag.String("excludeclasses",
		"XX",
		"comma-separated list of (uppercase, two-letter) line breaking classes to ignore; default to XX")
	localFiles = flag.Bool("local", false,
		"data files have been copied to current directory; for debugging only")
)

type class struct {
	name, doc string
}

// Supported line breaking classes for Unicode 6.2.0.
//
// Table loading depends on this: classes not listed here aren't loaded.
var classes = []class{
	{"OP", "Open Punctuation"},
	{"CL", "Close Punctuation"},
	{"CP", "Close Parenthesis"},
	{"QU", "Quotation"},
	{"GL", "Non-breaking (\"Glue\")"},
	{"NS", "Nonstarter"},
	{"EX", "Exclamation/Interrogation"},
	{"SY", "Symbols Allowing Break After"},
	{"IS", "Infix Numeric Separator"},
	{"PR", "Prefix Numeric"},
	{"PO", "Postfix Numeric"},
	{"NU", "Numeric"},
	{"AL", "Alphabetic"},
	{"HL", "Hebrew Letter"},
	{"ID", "Ideographic"},
	{"IN", "Inseparable"},
	{"HY", "Hyphen"},
	{"BA", "Break After"},
	{"BB", "Break Before"},
	{"B2", "Break Opportunity Before and After"},
	{"ZW", "Zero Width Space"},
	{"CM", "Combining Mark"},
	{"WJ", "Word Joiner"},
	{"H2", "Hangul LV Syllable"},
	{"H3", "Hangul LVT Syllable"},
	{"JL", "Hangul L Jamo"},
	{"JV", "Hangul V Jamo"},
	{"JT", "Hangul T Jamo"},
	{"RI", "Regional Indicator"},
	// Resolved outside of the pair table (> 28).
	{"BK", "Mandatory Break"},
	{"CR", "Carriage Return"},
	{"LF", "Line Feed"},
	{"NL", "Next Line"},
	{"SG", "Surrogate"},
	{"SP", "Space"},
	{"CB", "Contingent Break Opportunity"},
	{"AI", "Ambiguous (Alphabetic or Ideographic)"},
	{"CJ", "Conditional Japanese Starter"},
	{"SA", "Complex Context Dependent (South East Asian)"},
	{"XX", "Unknown"},
}

var pairTableSize = 29

// ----------------------------------------------------------------------------

var logger = log.New(os.Stderr, "", log.Lshortfile)

func allClassNames() []string {
	a := make([]string, 0, len(classes))
	for _, c := range classes {
		a = append(a, c.name)
	}
	sort.Strings(a)
	return a
}

func open(url string) *reader {
	file := filepath.Base(url)
	if *localFiles {
		fd, err := os.Open(file)
		if err != nil {
			logger.Fatal(err)
		}
		return &reader{bufio.NewReader(fd), fd, nil}
	}
	resp, err := http.Get(url)
	if err != nil {
		logger.Fatal(err)
	}
	if resp.StatusCode != 200 {
		logger.Fatalf("bad GET status for %s: %d", file, resp.Status)
	}
	return &reader{bufio.NewReader(resp.Body), nil, resp}
}

type reader struct {
	*bufio.Reader
	fd   *os.File
	resp *http.Response
}

func (r *reader) close() {
	if r.fd != nil {
		r.fd.Close()
	} else {
		r.resp.Body.Close()
	}
}

// ----------------------------------------------------------------------------
// Line break tables
// go run maketables.go > tables.go
// ----------------------------------------------------------------------------

// codePoint represents a code point (or range of code points) for a line
// breaking class.
type codePoint struct {
	lo, hi uint32 // range of code points
	class  string
}

const format = "\t\t{Lo:0x%04x, Hi:0x%04x, Stride:%d},\n"

var range16Count = 0 // Number of entries in the 16-bit range tables.
var range32Count = 0 // Number of entries in the 32-bit range tables.

const header = `// Generated by maketables.go
// DO NOT EDIT

package linebreak`

const imports = `import (
	"unicode"
)

// Version is the Unicode edition from which the tables are derived.
const Version = %q`

var codePointRe = regexp.MustCompile(`^([0-9A-F]+)(\.\.[0-9A-F]+)?;([A-Z0-9]{2})$`)

// LineBreak.txt has form:
//  4DFF;AL # HEXAGRAM FOR BEFORE COMPLETION
//  4E00..9FCC;ID # <CJK Ideograph, First>..<CJK Ideograph, Last>
func parseCodePoint(line string) (cp codePoint, ok bool) {
	comment := strings.Index(line, "#")
	if comment >= 0 {
		line = line[0:comment]
	}
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		ok = false
		return
	}
	field := strings.Split(line, ";")
	if len(field) != 2 {
		logger.Fatalf("%s: %d fields (expected 2)\n", line, len(field))
	}
	matches := codePointRe.FindStringSubmatch(line)
	if len(matches) != 4 {
		logger.Fatalf("%s: %d matches (expected 3)\n", line, len(matches))
	}
	lo, err := strconv.ParseUint(matches[1], 16, 64)
	if err != nil {
		logger.Fatalf("%.5s...: %s", line, err)
	}
	hi := lo
	if len(matches[2]) > 2 { // ignore leading ..
		hi, err = strconv.ParseUint(matches[2][2:], 16, 64)
		if err != nil {
			logger.Fatalf("%.5s...: %s", line, err)
		}
	}
	name := matches[3]
	return codePoint{uint32(lo), uint32(hi), name}, true
}

func loadCodePoints() map[string][]codePoint {
	if *dataURL == "" {
		flag.Set("data", *url+"LineBreak.txt")
	}
	codePoints := make(map[string][]codePoint)
	input := open(*dataURL)
	for {
		line, err := input.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Fatal(err)
		}
		cp, ok := parseCodePoint(line[0 : len(line)-1])
		if ok {
			codePoints[cp.class] = append(codePoints[cp.class], cp)
		}
	}
	input.close()
	return codePoints
}

func printHeader() {
	fmt.Print(header + "\n\n")
}

func printClasses() {
	fmt.Printf(imports+"\n\n", version())
	fmt.Print(`
	// These are the possible line break classifications.
	// Since new unicode versions may add new types here, applications should be ready
	// to handle unknown values. They may be regarded as G_UNICODE_BREAK_UNKNOWN.
	// See [Unicode Line Breaking Algorithm](http://www.unicode.org/unicode/reports/tr14/).
	`)
	fmt.Printf("type %s int\n\n", ClassTypeName)
	fmt.Print("// Line breaking classes.\n")
	fmt.Print("//\n")
	fmt.Print("// See: http://www.unicode.org/reports/tr14/#Table1\n")
	fmt.Print("const (\n")
	for k, v := range classes {
		if k == 0 {
			fmt.Printf("\t%s%s %s = iota // %s\n", ClassPrefix, v.name, ClassTypeName, v.doc)
		} else {
			fmt.Printf("\t%s%s                   // %s\n", ClassPrefix, v.name, v.doc)
		}
		if k+1 == pairTableSize && k < len(classes)-1 {
			fmt.Printf("\t// Resolved outside of the pair table (> %d).\n", k)
		}
	}
	fmt.Print(")\n\n")

	fmt.Print("// Class returns the line breaking class for the given rune.\n")
	fmt.Printf("func Class(r rune) %s {\n", ClassTypeName)
	fmt.Print("\t// TODO test more common first?\n")
	fmt.Print("\tswitch {\n")

	excludelist := strings.Split(strings.ToUpper(*excludeclasses), ",")
	for _, name := range allClassNames() {
		if excludeClass(name, excludelist) {
			continue
		}
		fmt.Printf("\tcase unicode.Is(%s, r):\n", name)
		fmt.Printf("\t\treturn %s%s\n", ClassPrefix, name)
	}
	fmt.Print("\t}\n")
	fmt.Printf("\treturn %sXX\n", ClassPrefix)
	fmt.Print("}\n\n")
}

func printTables() {
	excludelist := strings.Split(strings.ToUpper(*excludeclasses), ",")

	var list []string
	for _, name := range allClassNames() {
		if !excludeClass(name, excludelist) {
			list = append(list, name)
		}
	}

	codePoints := loadCodePoints()

	decl := make(sort.StringSlice, len(list))
	ndecl := 0
	for _, name := range list {
		cp, ok := codePoints[name]
		if !ok {
			continue
		}
		decl[ndecl] = fmt.Sprintf(
			"\t%s = _%s; // %s is the set of Unicode characters in line breaking class %s.\n",
			name, name, name, name)
		ndecl++
		fmt.Printf("var _%s = &unicode.RangeTable {\n", name)
		ranges := foldAdjacent(cp)
		fmt.Print("\tR16: []unicode.Range16{\n")
		size := 16
		count := &range16Count
		for _, s := range ranges {
			size, count = printRange(s.Lo, s.Hi, s.Stride, size, count)
		}
		fmt.Print("\t},\n")
		if off := findLatinOffset(ranges); off > 0 {
			fmt.Printf("\tLatinOffset: %d,\n", off)
		}
		fmt.Print("}\n\n")
	}
	decl.Sort()
	fmt.Println("// These variables have type *unicode.RangeTable.")
	fmt.Println("var (")
	for _, d := range decl {
		fmt.Print(d)
	}
	fmt.Print(")\n\n")
}

func printSizes() {
	fmt.Printf("// Range entries: %d 16-bit, %d 32-bit, %d total.\n", range16Count, range32Count, range16Count+range32Count)
	range16Bytes := range16Count * 3 * 2
	range32Bytes := range32Count * 3 * 4
	fmt.Printf("// Range bytes: %d 16-bit, %d 32-bit, %d total.\n", range16Bytes, range32Bytes, range16Bytes+range32Bytes)
}

// Extract the version number from the URL
func version() string {
	// Break on slashes and look for the first numeric field
	fields := strings.Split(*url, "/")
	for _, f := range fields {
		if len(f) > 0 && '0' <= f[0] && f[0] <= '9' {
			return f
		}
	}
	logger.Fatal("unknown version")
	return "Unknown"
}

func excludeClass(class string, excludelist []string) bool {
	for _, name := range excludelist {
		if name == class {
			return true
		}
	}
	return false
}

// Tables may have a lot of adjacent elements. Fold them together.
func foldAdjacent(r []codePoint) []unicode.Range32 {
	s := make([]unicode.Range32, 0, len(r))
	j := 0
	for i := 0; i < len(r); i++ {
		if j > 0 && r[i].lo == s[j-1].Hi+1 {
			s[j-1].Hi = r[i].hi
		} else {
			s = s[0 : j+1]
			s[j] = unicode.Range32{
				Lo:     uint32(r[i].lo),
				Hi:     uint32(r[i].hi),
				Stride: 1,
			}
			j++
		}
	}
	return s
}

func printRange(lo, hi, stride uint32, size int, count *int) (int, *int) {
	if size == 16 && hi >= 1<<16 {
		if lo < 1<<16 {
			if lo+stride != hi {
				logger.Fatalf("unexpected straddle: %U %U %d", lo, hi, stride)
			}
			// No range contains U+FFFF as an instance, so split
			// the range into two entries. That way we can maintain
			// the invariant that R32 contains only >= 1<<16.
			fmt.Printf(format, lo, lo, 1)
			lo = hi
			stride = 1
			*count++
		}
		fmt.Print("\t},\n")
		fmt.Print("\tR32: []unicode.Range32{\n")
		size = 32
		count = &range32Count
	}
	fmt.Printf(format, lo, hi, stride)
	*count++
	return size, count
}

func findLatinOffset(ranges []unicode.Range32) int {
	i := 0
	for i < len(ranges) && ranges[i].Hi <= unicode.MaxLatin1 {
		i++
	}
	return i
}

// ----------------------------------------------------------------------------
// Test tables
// go run maketables.go --test > tables_test.go
// ----------------------------------------------------------------------------

type test struct {
	id      int
	comment string
	text    string
	breaks  []int
}

func parseTest(line string) (t test, ok bool) {
	var comment string
	commentIdx := strings.Index(line, "#")
	if commentIdx >= 0 {
		comment = strings.TrimSpace(line[commentIdx+1:])
		line = line[0:commentIdx]
	}
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		ok = false
		return
	}

	var testString string
	var lineBreaks []int
	mark := false
	for _, v := range strings.Split(line, " ") {
		switch v {
		case "×", "÷":
			if mark {
				logger.Fatalf("%q: double break mark\n", line)
			}
			mark = true
			if v == "×" {
				lineBreaks = append(lineBreaks, 0)
			} else {
				lineBreaks = append(lineBreaks, 1)
			}
		default:
			if !mark {
				logger.Fatalf("%q: double code point\n", line)
			}
			mark = false
			u, err := strconv.ParseUint(v, 16, 64)
			if err != nil {
				logger.Fatalf("%s: %s", line, err)
			}
			//cp := fmt.Sprintf("%+q", u)
			//testString += cp[1:len(cp)-1]
			if u >= 1<<16 {
				testString += fmt.Sprintf("\\U%08x", u)
			} else {
				testString += fmt.Sprintf("\\u%04x", u)
			}
		}
	}
	return test{0, comment, testString, lineBreaks}, true
}

func loadTests() []test {
	if *testDataURL == "" {
		flag.Set("testdata", *url+"auxiliary/LineBreakTest.txt")
	}
	var tests []test
	input := open(*testDataURL)
	for {
		line, err := input.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Fatal(err)
		}
		t, ok := parseTest(line[0 : len(line)-1])
		if ok {
			t.id = len(tests)
			tests = append(tests, t)
		}
	}
	input.close()
	return tests
}

func printTests() {
	fmt.Println("type lineBreakTest struct {")
	fmt.Println("\tid     int")
	fmt.Println("\ttext   string")
	fmt.Println("\tbreaks []int")
	fmt.Println("}\n")

	fmt.Println("var lineBreakTests = []lineBreakTest{")
	for _, v := range loadTests() {
		fmt.Printf("\t// %s\n", v.comment)
		fmt.Printf("\t{%d, \"%s\", %#v},\n", v.id, v.text, v.breaks)
	}
	fmt.Println("}")
}