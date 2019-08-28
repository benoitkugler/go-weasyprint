package utils

import (
	"fmt"
	"testing"
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
