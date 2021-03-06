// +build ignore

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/rangetable"
)

const (
	urlIndic     = "https://unicode.org/Public/UCD/latest/ucd/IndicSyllabicCategory.txt"
	urlEmoji     = "https://www.unicode.org/Public/emoji/12.0/emoji-data.txt"
	urlLineBreak = "https://www.unicode.org/Public/12.0.0/ucd/LineBreak.txt"

	urlLineBreakTest = "https://www.unicode.org/Public/UCD/latest/ucd/auxiliary/LineBreakTest.txt"
)

// Supported line breaking classes for Unicode 12.0.0.
// Table loading depends on this: classes not listed here aren't loaded.
var lineBreakClasses = [][2]string{
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
	{"EB", "Emoji Base"},
	{"EM", "Emoji Modifier"},
	{"ZWJ", "Zero width joiner"},
	{"XX", "Unknown"},
}

func main() {
	test := flag.Bool("t", false, "switch to test file")
	flag.Parse()
	if *test {
		printTests()
	} else {
		printTables()
	}
}

func printTables() {
	fmt.Printf(`
	// Generated by gen/main.go
	// DO NOT EDIT

	// Indic data from : %s
	// Emoji data from : %s
	// Breaks data from : %s

	package unicodedata
	
	import "unicode"
	`, urlIndic, urlEmoji, urlLineBreak)

	vars, dict := "", ""
	var classNames []string
	for _, c := range lineBreakClasses {
		classNames = append(classNames, c[0])
		vars += fmt.Sprintf("Break%s = _%s // %s\n", c[0], c[0], c[1])
		dict += fmt.Sprintf("\"%s\" : _%s, \n", c[0], c[0])
	}
	fmt.Printf(` var (
		%s)
	var Breaks = map[string]*unicode.RangeTable{
		%s}
	`, vars, dict)

	loadAndPrint(urlIndic, "Virama", "Vowel_Dependent")
	loadAndPrint(urlEmoji, "Emoji", "Emoji_Presentation", "Emoji_Modifier", "Emoji_Modifier_Base", "Extended_Pictographic")
	loadAndPrint(urlLineBreak, classNames...)
}

func loadAndPrint(url string, types ...string) {
	s := FetchData(url)
	ranges := Parse(s)
	for _, typ := range types {
		PrintTable("_"+typ, ranges[typ])
	}
}

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

func (r Rg) Runes() []rune {
	var out []rune
	for ru := r.Start; ru <= r.End; ru++ {
		out = append(out, rune(ru))
	}
	return out
}

func PrintTable(name string, rt *unicode.RangeTable) {
	fmt.Printf("var %s = &unicode.RangeTable{\n", name)
	fmt.Println("\tR16: []unicode.Range16{")
	for _, r := range rt.R16 {
		fmt.Printf("\t\t{Lo:%#04x, Hi:%#04x, Stride:%d},\n", r.Lo, r.Hi, r.Stride)
	}
	fmt.Println("\t},")
	if len(rt.R32) > 0 {
		fmt.Println("\tR32: []unicode.Range32{")
		for _, r := range rt.R32 {
			fmt.Printf("\t\t{Lo:%#x, Hi:%#x,Stride:%d},\n", r.Lo, r.Hi, r.Stride)
		}
		fmt.Println("\t},")
	}
	if rt.LatinOffset > 0 {
		fmt.Printf("\tLatinOffset: %d,\n", rt.LatinOffset)
	}
	fmt.Printf("}\n\n")
}

func Parse(data string) map[string]*unicode.RangeTable {
	outRanges := map[string][]rune{}
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
		outRanges[typ] = append(outRanges[typ], Rg{Start: start, End: end}.Runes()...)
	}
	out := make(map[string]*unicode.RangeTable, len(outRanges))
	for k, v := range outRanges {
		out[k] = rangetable.New(v...)
	}
	return out
}

// ----------------------------------------------------------------------------
// 	Test tables
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
				log.Fatalf("%q: double break mark\n", line)
			}
			mark = true
			if v == "×" {
				lineBreaks = append(lineBreaks, 0)
			} else {
				lineBreaks = append(lineBreaks, 1)
			}
		default:
			if !mark {
				log.Fatalf("%q: double code point\n", line)
			}
			mark = false
			u, err := strconv.ParseUint(v, 16, 64)
			if err != nil {
				log.Fatalf("%s: %s", line, err)
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
	s := FetchData(urlLineBreakTest)
	var tests []test
	for _, line := range strings.Split(s, "\n") {
		t, ok := parseTest(line)
		if ok {
			t.id = len(tests)
			tests = append(tests, t)
		}
	}
	return tests
}

func printTests() {
	fmt.Printf(`
	// Generated by gen/main.go
	// DO NOT EDIT

	// Breaks test data from : %s

	package text
	
	`, urlLineBreakTest)

	fmt.Println(`type lineBreakTest struct {
		id     int
		text   string
		breaks []int
	}
	`)

	fmt.Println("var lineBreakTests = []lineBreakTest{")
	for _, v := range loadTests() {
		fmt.Printf("\t// %s\n", v.comment)
		fmt.Printf("\t{%d, \"%s\", %#v},\n", v.id, v.text, v.breaks)
	}
	fmt.Println("}")
}
