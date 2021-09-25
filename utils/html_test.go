package utils

import (
	"fmt"
	"strings"
	"testing"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var dates = [7]string{
	"1997",
	"1997-07",
	"1997-07-16",
	"1997-07-16T19:20+01:00",
	"1997-07-16T19:20:30+01:00",
	"1997-07-16T19:20:30.45+01:00",
	"1997-07-16T19:20:30.45-01:15",
}

func TestReW3CDate(t *testing.T) {
	for _, date := range dates {
		if !W3CDateRe.MatchString(date) {
			t.Errorf("date %s not matching", date)
		}

		if parseW3cDate("test", date).IsZero() {
			t.Errorf("date %s not parsed", date)
		}
	}
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

func TestHtmlIterator(t *testing.T) {
	n, err := html.Parse(strings.NewReader("<meta></meta><link></link><p><a>sldmskd</a></p>"))
	if err != nil {
		t.Fatal(err)
	}
	iter := NewHtmlIterator(n, atom.Meta, atom.Link)
	for iter.HasNext() {
		elem := iter.Next()
		if elem == nil {
			t.Fatal("inconstistent iterator")
		}
	}
}
