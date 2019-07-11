package utils

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// ------------------------------------ html walk utilities ------------------------------------

// NodeChildren returns the direct children of `element`
func NodeChildren(element html.Node) (children []html.Node) {
	child := element.FirstChild
	for child != nil {
		children = append(children, *child)
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

// GetAttribute returns the attribute `name` or ""
func GetAttribute(element html.Node, name string) string {
	for _, attr := range element.Attr {
		if attr.Key == name {
			return attr.Val
		}
	}
	return ""
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
