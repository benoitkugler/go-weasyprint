package parser

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

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

func TestRune(t *testing.T) {
	s := strings.Repeat(string(rand.Int()), 1000000)
	ti := time.Now()
	vi := []rune(s)
	fmt.Println(time.Since(ti), len(vi))

	fmt.Println(utf8.DecodeRuneInString("--"))
	fmt.Println("--"[0])
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
