package css

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/benoitkugler/go-weasyprint/utils"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
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

func TestIterTree(t *testing.T) {
	s := "<html><head><base></base></head></html>"
	n, err := html.Parse(strings.NewReader(s))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(utils.Iter(*n, atom.Head))
}

func TestRe(t *testing.T) {
	s := regexp.MustCompile(
		`^` +
			"[ \t\n\f\r]*" +
			`(?P<year>\d\d\d\d)` +
			`(?:` +
			`-(?P<month>0\d|1[012])` +
			`(?:` +
			`-(?P<day>[012]\d|3[01])` +
			`(?:` +
			`T(?P<hour>[01]\d|2[0-3])` +
			`:(?P<minute>[0-5]\d)` +
			`(?:` +
			`:(?P<second>[0-5]\d)` +
			`(?:\.\d+)?` + // Second fraction, ignored
			`)?` +
			`(?:` +
			`Z |` + //# UTC
			`(?P<tzHour>[+-](?:[01]\d|2[0-3]))` +
			`:(?P<tzMinute>[0-5]\d)` +
			`)` +
			`)?` +
			`)?` +
			`)?` +
			"[ \t\n\f\r]*" +
			`$`)
	fmt.Println(s)

	fmt.Println(s.MatchString("1997"))
	fmt.Println(s.MatchString("1997-07"))
	fmt.Println(s.MatchString("1997-07-16"))
	fmt.Println(s.MatchString("1997-07-16T19:20+01:00"))
	fmt.Println(s.MatchString("1997-07-16T19:20:30+01:00"))
	fmt.Println(s.MatchString("1997-07-16T19:20:30.45+01:00"))
}
