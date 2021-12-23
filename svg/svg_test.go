package svg

import (
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
