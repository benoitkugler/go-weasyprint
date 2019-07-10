package css

import (
	"fmt"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

type A struct {
	i int
}

type B interface {
	Self() *A
}

func (a *A) Self() *A {
	return a
}

func TestI(t *testing.T) {
	var a A
	b := B(&a)
	p1 := b.Self()
	p2 := &a
	fmt.Printf("%p %p", p1, p2)

}

func TestHtmlParse(t *testing.T) {
	s := "<div><p>mldsdk</p>skdlsldj</div>"
	n, err := html.Parse(strings.NewReader(s))
	if err != nil {
		t.Fatal(err)
	}
	st := n.FirstChild.FirstChild.NextSibling.FirstChild.FirstChild
	fmt.Println(st, st.NextSibling)
}
