package utils

import (
	"fmt"
	"testing"
)

func TestQuote(t *testing.T) {
	s := "“”"
	fmt.Println(string(s[0]), string(s[1]))
}
