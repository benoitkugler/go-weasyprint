package utils

import (
	"fmt"
	"strings"
	"testing"
	"unicode"
)

func TestQuote(t *testing.T) {
	s := "“”"
	fmt.Println(string(s[0]), string(s[1]))
}

func TestSlice(t *testing.T) {
	s := []int{0, 1, 2, 3, 4, 5, 6, 7, 8}
	var out []int
	for i := 1; i < len(s); i += 2 {
		out = append(out, s[i])
	}
	fmt.Println(out)
}

func TestUnicode(t *testing.T) {
	for _, c := range "abc€" {
		fmt.Println(c)
	}

	for _, letter := range "amcp" {
		fmt.Println(0x20 <= letter && letter <= 0x7f)
	}
	// fmt.Println([]rune("€"))
}

func TestLower(t *testing.T) {
	keyword := "Bac\u212Aground"
	rs := []rune(keyword)
	out := make([]rune, len(rs))
	for index, c := range rs {
		fmt.Println(index, c)
		if c <= unicode.MaxASCII {
			c = unicode.ToLower(c)
		}
		out[index] = c
	}

	fmt.Println(keyword == "BacKground")

	fmt.Println(strings.ToLower(keyword) == "background")
	// fmt.Println(asciiLower(keyword) != strings.ToLower(keyword))
	// fmt.Println(asciiLower(keyword) == "bac\u212Aground")
	fmt.Println(unicode.MaxASCII)

	fmt.Println(out, string(out))
}

func TestPointer(t *testing.T) {
	var i, j []int

	p := &i

	*p = append(*p, 4, 4, 5, 7, 8, 9, 6, 3, 8, 5, 9, 9, 3)
	p = &j

	*p = append(*p, 4, 4, 5, 7, 8, 9, 6, 3, 8, 5, 9, 9, 3)
	fmt.Println(i, j)
}
