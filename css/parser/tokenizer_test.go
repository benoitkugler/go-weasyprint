package parser

import (
	"fmt"
	"testing"

	"github.com/gorilla/css/scanner"
)

func showTokens(s string) {
	sc := scanner.New(s)
	for {
		token := sc.Next()
		if token.Type == scanner.TokenEOF {
			break
		}
		fmt.Println(token.Type, token.Value)
	}
}
func TestUnicodeRange(t *testing.T) {
	s := `html {
		color: blue;
		background: url(http://www.google.fr);
	}
	p.test {
		font-size: 15px;
	}
	`
	l := ParseComponentValueList(s, false)
	fmt.Println(l)
	s = Serialize(l)
	fmt.Println(s)
}
