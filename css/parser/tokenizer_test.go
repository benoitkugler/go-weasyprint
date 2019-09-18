package parser

import (
	"encoding/json"
	"fmt"
	"log"
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

func TestGen(t *testing.T) {
	s := ` --red  `
	l := ParseComponentValueList(s, false)
	fmt.Println(l)
	s = Serialize(l)
	fmt.Println(s)
}

func TestEscapeRe(t *testing.T) {
	fmt.Println(hexEscapeRe.FindAllString("130 ", -1))
	fmt.Println(hexEscapeRe.FindStringSubmatch("130 "))
}

func TestJSON(t *testing.T) {
	l := []string{"lmkf", "lmldfmlklmf"}
	b, err := json.Marshal(l)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
