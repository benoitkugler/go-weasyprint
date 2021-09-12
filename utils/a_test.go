package utils

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"testing"
	"unicode"

	"github.com/vincent-petithory/dataurl"

	"golang.org/x/net/html"

	"github.com/andybalholm/cascadia"
)

func TestQuote(t *testing.T) {
	s := "“”"
	fmt.Println(string(s[0]), string(s[1]))
}

func TestSlice(t *testing.T) {
	s := []int{0, 1, 2, 3, 4, 5, 6, 7, 8}
	fmt.Println(s[9:])
	for i := range s {
		if i == 4 {
			s = s[i+1:]
			break
		}
	}
	fmt.Println(s)

	var out []int
	for i := 1; i < len(s); i += 2 {
		out = append(out, s[i])
	}
	fmt.Println(out)

	a := []int{1, 2, 3, 4, 5, 6, 7, 8}
	p, poped := a[len(a)-1], a[:len(a)-1]
	fmt.Println("poped :", p, poped)
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

func TestInterface(t *testing.T) {
	var i io.Reader
	_, ok := i.(*bytes.Reader)
	fmt.Println(ok)
}

func TestPointer(t *testing.T) {
	var i, j []int

	p := &i

	*p = append(*p, 4, 4, 5, 7, 8, 9, 6, 3, 8, 5, 9, 9, 3)
	p = &j

	*p = append(*p, 4, 4, 5, 7, 8, 9, 6, 3, 8, 5, 9, 9, 3)
	fmt.Println(i, j)
}

type T struct {
	i int
}

func TestLoop(t *testing.T) {
	a := make([]T, 10)
	for _, t := range a {
		t.i = 5
	}
	fmt.Println(a)
}

func TestSelector(t *testing.T) {
	sel, err := cascadia.Compile("style, link")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(sel)
	s := "<html><p>dlfkdfk</p><div><style>sdsd</style><style>sdsd</style><link /></div></html>"
	root, err := html.Parse(strings.NewReader(s))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(sel.MatchAll(root))
}

func TestWalkHtml(t *testing.T) {
	s := "<html><p>dlfkdfk</p><div><span>sdsd/<span><span></span></div></html>"
	root, err := html.Parse(strings.NewReader(s))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(root.FirstChild)
	iter := NewHtmlIterator(root.FirstChild)
	for iter.HasNext() {
		n := iter.Next()
		fmt.Printf("%p %v %s\n", n, n.DataAtom, n.Data)
	}
}

func TestRune(t *testing.T) {
	fmt.Printf("%c", '\'')
	fmt.Printf("%c", '\u2e80')
	var c rune = -1
	fmt.Println(c)
}

func TestUrl(t *testing.T) {
	p := "/ssdsmldk/mldsjkd/erree/"
	fmt.Println(path.Join(path.Dir(p), "mùd.html"))

	_, err := dataurl.DecodeString("data:text/css;charset=utf-16le;base64,                    bABpAHsAYwBvAGwAbwByADoAcgBlAGQAfQA=")
	if err != nil {
		t.Fatal(err)
	}
}

type i struct {
	u int
}

func TestPOINTER(t *testing.T) {
	a := struct {
		i
	}{
		i: i{u: 9},
	}
	ta := a.i
	ta.u += 10
	t.Log(ta, a)
}

func TestReference(t *testing.T) {
	s := []i{{u: 4}, {u: 5}}
	s[0].u = 78
	fmt.Println(s)
	v := s[0]
	fmt.Printf("%p %p", &v, &s[0])
}

func TestSortOrder(t *testing.T) {
	u := []int{1, 2, 8, 9, 4, 5, 7, 6, 1, 2, 8}
	sort.Slice(u, func(i, j int) bool {
		return u[i] < u[j]
	})
	fmt.Println(u)
}

func TestInf(t *testing.T) {
	u := math.Inf(1)
	fmt.Println(u, -u, -u < 0)
}

func TestParseInt(t *testing.T) {
	fmt.Println(strconv.Atoi("+789"))
	fmt.Println(strconv.Atoi("-789"))
}

func TestHex(t *testing.T) {
	md := md5.Sum([]byte("lmkelmezkezmlekm"))
	mdb := make([]byte, len(md))
	for i, v := range md {
		mdb[i] = v
	}
	fmt.Println(mdb)
	s := hex.EncodeToString(mdb)
	fmt.Println(s)
}

func TestScan(t *testing.T) {
	f, _ := os.Open("../pt")
	s := bufio.NewScanner(f)
	s.Scan()
	fmt.Println(len(string(s.Bytes())))
}

func TestFormat(t *testing.T) {
	fmt.Printf("%010d", 45)
}
