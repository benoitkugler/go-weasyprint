package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const htmlWhitespace = " \t\n\f\r"

var (
	htmlSpaceSeparatedTokensRe = regexp.MustCompile(fmt.Sprintf("[^%s]+", htmlWhitespace))
)

type HTMLNode html.Node

// Get returns the attribute `name` or ""
func (h HTMLNode) Get(name string) string {
	for _, attr := range h.Attr {
		if attr.Key == name {
			return attr.Val
		}
	}
	return ""
}

// ------------------------------------ html walk utilities ------------------------------------

// HtmlIterator simplify the (depth first) walk on an HTML tree.
type HtmlIterator struct {
	toVisit []*html.Node
}

// NewHtmlIterator use `root` as start point.
func NewHtmlIterator(root *html.Node) HtmlIterator {
	return HtmlIterator{toVisit: []*html.Node{root}}
}

func (h HtmlIterator) HasNext() bool {
	return len(h.toVisit) > 0
}

func (h *HtmlIterator) Next() *HTMLNode {
	next := h.toVisit[0]
	h.toVisit = h.toVisit[1:]
	if next.FirstChild != nil {
		h.toVisit = append(h.toVisit, next.FirstChild)
	}
	if next.NextSibling != nil {
		h.toVisit = append(h.toVisit, next.NextSibling)
	}
	return (*HTMLNode)(next)
}

// NodeChildren returns the direct children of `element`
func NodeChildren(element html.Node) (children []*html.Node) {
	child := element.FirstChild
	for child != nil {
		children = append(children, child)
		child = child.NextSibling
	}
	return
}

// Iter recursively `element` (and its children and so on ...) and returns the elements matching one of the given tags
func Iter(element html.Node, tags ...atom.Atom) []html.Node {
	tagsMap := make(map[atom.Atom]bool)
	for _, tag := range tags {
		tagsMap[tag] = true
	}
	var aux func(html.Node) []html.Node
	aux = func(el html.Node) (out []html.Node) {
		if tagsMap[el.DataAtom] {
			out = append(out, el)
		}
		child := el.FirstChild
		for child != nil {
			out = append(out, aux(*child)...)
			child = child.NextSibling
		}
		return
	}
	return aux(element)
}

// GetChildText returns the text directly in the element, not descendants.
func GetChildText(element html.Node) string {
	var content []string
	if element.Type == html.TextNode {
		content = []string{element.Data}
	}

	for _, child := range NodeChildren(element) {
		if child.Type == html.TextNode {
			content = append(content, child.Data)
		}
	}
	return strings.Join(content, "")
}

// Transform (only) ASCII letters to lower case: A-Z is mapped to a-z.
//     This is used for `ASCII case-insensitive
//     <http://whatwg.org/C#ascii-case-insensitive>`_ matching.
//     This is different from the strings.ToLower function
//     which also affect non-ASCII characters,
//     sometimes mapping them into the ASCII range:
//     		keyword = u"Bac\u212Aground"
//     		assert strings.ToLower(keyword) == u"background"
//     		assert asciiLower(keyword) != strings.ToLower(keyword)
//     		assert asciiLower(keyword) == u"bac\u212Aground"
//
func AsciiLower(s string) string {
	rs := []rune(s)
	out := make([]rune, len(rs))
	for index, c := range rs {
		if c < unicode.MaxASCII {
			c = unicode.ToLower(c)
		}
		out[index] = c
	}
	return string(out)
}

// Return whether the given element has a ``rel`` attribute with the
// given link type.
// `linkType` must be a lower-case string.
func ElementHasLinkType(element HTMLNode, linkType string) bool {
	attr := element.Get("rel")
	matchs := htmlSpaceSeparatedTokensRe.FindAllString(attr, -1)
	for _, token := range matchs {
		if AsciiLower(token) == linkType {
			return true
		}
	}
	return false
}
