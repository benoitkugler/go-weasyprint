package svg

import (
	"reflect"
	"strings"
	"testing"
)

func TestHandleText(t *testing.T) {
	input := `
		<svg width="100%" height="100%" viewBox="0 0 1000 300"
			xmlns="http://www.w3.org/2000/svg"
			xmlns:xlink="http://www.w3.org/1999/xlink">
		<defs>
			<text id="ReferencedText">
				Referenced character data
			</text>
		</defs>

		<text x="100" y="100" font-size="45" >
			Inline character data
		</text>

		<text>
			<textPath href="#MyPath">
			Quick brown fox jumps over the lazy dog.
			</textPath>
		</text>

		<text x="100" y="200" font-size="45" fill="red" >
			<tref xlink:href="#ReferencedText"/>
		</text>

		<!-- Show outline of canvas using 'rect' element -->
		<rect x="1" y="1" width="998" height="298"
				fill="none" stroke-width="2" />
		</svg>
		`
	img, err := buildSVGTree(strings.NewReader(input), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(img.defs) != 1 {
		t.Fatal("defs")
	}
	if c, has := img.defs["ReferencedText"]; !has || len(c.children) != 0 {
		t.Fatal("defs circle")
	}
}

func TestFilter(t *testing.T) {
	input := `
	<svg width="230" height="120" xmlns="http://www.w3.org/2000/svg">
		<filter id="blurMe">
			<feBlend in="SourceGraphic" in2="floodFill" mode="multiply"/>
			<feOffset in="SourceGraphic" dx="60" dy="60" />
			<feGaussianBlur stdDeviation="5"/>
		</filter>
		
		<circle cx="60" cy="60" r="50" fill="green"/>
		
		<circle cx="170" cy="60" r="50" fill="green" filter="url(#blurMe)"/>
	</svg>
	`
	out, err := Parse(strings.NewReader(input), "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(out.definitions.filters, map[string][]filter{
		"blurMe": {
			filterBlend("multiply"),
			filterOffset{dx: value{v: 60}, dy: value{v: 60}, isUnitsBBox: false},
		},
	}) {
		t.Fatal(out.definitions.filters)
	}
}
