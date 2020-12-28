package pango

import (
	"encoding/xml"
	"fmt"
	"testing"
)

type Node struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content []byte     `xml:",innerxml"`
	Nodes   []Node     `xml:",any"`
}

func (n *Node) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	n.Attrs = start.Attr
	type node Node
	fmt.Println(start)
	return d.DecodeElement((*node)(n), &start)
}

func TestBasicParse(t *testing.T) {
	a := "<b>bold <big>big</big> <i>italic</i></b> <s>strikethrough<sub>sub</sub> <small>small</small><sup>sup</sup></s> <tt>tt <u>underline</u></tt>"

	var out Node
	err := xml.Unmarshal([]byte("<toplevel>"+a+"</toplevel>"), &out)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(out.Nodes))
}
