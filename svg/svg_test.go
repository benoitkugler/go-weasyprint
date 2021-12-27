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
			filterOffset{dx: value{60, Px}, dy: value{60, Px}, isUnitsBBox: false},
		},
	}) {
		t.Fatal(out.definitions.filters)
	}
}

func TestClipPath(t *testing.T) {
	input := `
	<svg viewBox="0 0 100 100">
	<clipPath id="myClip">
	<!--
		Everything outside the circle will be
		clipped and therefore invisible.
	-->
	<circle cx="40" cy="35" r="35" />
	</clipPath>

	<!-- The original black heart, for reference -->
	<path id="heart" d="M10,30 A20,20,0,0,1,50,30 A20,20,0,0,1,90,30 Q90,60,50,90 Q10,60,10,30 Z" />

	<!--
	Only the portion of the red heart
	inside the clip circle is visible.
	-->
	<use clip-path="url(#myClip)" href="#heart" fill="red" />
	</svg>
	`
	out, err := Parse(strings.NewReader(input), "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.definitions.clipPaths) != 1 {
		t.Fatal()
	}
	cp := out.definitions.clipPaths["myClip"]
	if len(cp.children) != 1 {
		t.Fatal()
	}
	if _, ok := cp.children[0].content.(ellipse); !ok {
		t.Fatal()
	}
}

func TestMask(t *testing.T) {
	input := `
	<svg viewBox="-10 -10 120 120">
	<mask id="myMask" x="5" width="12pt">
		<!-- Everything under a white pixel will be visible -->
		<rect x="0" y="0" width="100" height="100" fill="white" />

		<!-- Everything under a black pixel will be invisible -->
		<path d="M10,35 A20,20,0,0,1,50,35 A20,20,0,0,1,90,35 Q90,65,50,95 Q10,65,10,35 Z" fill="black" />
	</mask>

	<polygon points="-10,110 110,110 110,-10" fill="orange" />

	<!-- with this mask applied, we "punch" a heart shape hole into the circle -->
	<circle cx="50" cy="50" r="50" mask="url(#myMask)" />
	</svg>
	`
	out, err := Parse(strings.NewReader(input), "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.definitions.masks) != 1 {
		t.Fatal()
	}
	ma := out.definitions.masks["myMask"]
	if ma.box != (box{
		x:      value{5, Px},
		y:      value{-10, Perc}, // default
		width:  value{12, Pt},
		height: value{120, Perc}, // default
	}) {
		t.Fatal(ma.box)
	}
	if len(ma.children) != 2 {
		t.Fatal()
	}
	if _, ok := ma.children[0].content.(rect); !ok {
		t.Fatal()
	}
	if _, ok := ma.children[1].content.(path); !ok {
		t.Fatal()
	}
}
