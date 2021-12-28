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
	if _, ok := cp.children[0].graphicContent.(ellipse); !ok {
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
	if _, ok := ma.children[0].graphicContent.(rect); !ok {
		t.Fatal()
	}
	if _, ok := ma.children[1].graphicContent.(path); !ok {
		t.Fatal()
	}
}

func TestMarker(t *testing.T) {
	input := `
	<svg viewBox="0 0 120 120" xmlns="http://www.w3.org/2000/svg">
	<defs>
		<marker id="triangle" viewBox="0 0 10 10"
			refX="1" refY="5"
			markerUnits="strokeWidth"
			markerWidth="10" markerHeight="10"
			orient="auto">
			<path d="M 0 0 L 10 5 L 0 10 z" fill="#f00"/>
		</marker>
	</defs>
	<polyline fill="none" stroke="black"
		points="20,100 40,60 70,80 100,20" marker-start="url(#triangle)"/>
	</svg>
	`
	out, err := Parse(strings.NewReader(input), "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.definitions.markers) != 1 {
		t.Fatal()
	}
	ma := out.definitions.markers["triangle"]
	if ma.refX != (value{1, Px}) {
		t.Fatal(ma.refX)
	}
	if ma.refY != (value{5, Px}) {
		t.Fatal(ma.refY)
	}
	if ma.viewbox == nil || *ma.viewbox != (Rectangle{0, 0, 10, 10}) {
		t.Fatal(ma.viewbox)
	}

	if len(ma.children) != 1 {
		t.Fatal()
	}
	if _, ok := ma.children[0].graphicContent.(path); !ok {
		t.Fatal()
	}
}

func TestGradient(t *testing.T) {
	input := `
	<svg width="120" height="240" version="1.1" xmlns="http://www.w3.org/2000/svg">
	<defs>
		<linearGradient id="LinearGradient1">
			<stop class="stop1" offset="0%"/>
			<stop class="stop2" offset="50%"/>
			<stop class="stop3" offset="100%"/>
		</linearGradient>
		<linearGradient id="LinearGradient2" x1="0" x2="0" y1="0" y2="1">
			<stop offset="0%" stop-color="red"/>
			<stop offset="50%" stop-color="black" stop-opacity="0"/>
			<stop offset="100%" stop-color="blue"/>
		</linearGradient>
		<style type="text/css"><![CDATA[
			#rect1 { fill: url(#LinearGradient1); }
			.stop1 { stop-color: red; }
			.stop2 { stop-color: black; stop-opacity: 0; }
			.stop3 { stop-color: blue; }
		]]></style>

		<radialGradient id="RadialGradient1">
			<stop offset="0%" stop-color="red"/>
			<stop offset="100%" stop-color="blue"/>
		</radialGradient>
		<radialGradient id="RadialGradient2" cx="0.25" cy="0.25" r="0.25">
			<stop offset="0%" stop-color="red"/>
			<stop offset="100%" stop-color="blue"/>
		</radialGradient>
	</defs>

	</svg>
	`
	out, err := Parse(strings.NewReader(input), "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.definitions.paintServers) != 4 {
		t.Fatal(out.definitions.paintServers)
	}
	g1, ok := out.definitions.paintServers["LinearGradient1"].(gradient)
	if !ok {
		t.Fatal()
	}
	if _, ok = g1.kind.(gradientLinear); !ok {
		t.Fatal()
	}
	g2, ok := out.definitions.paintServers["LinearGradient2"].(gradient)
	if !ok {
		t.Fatal()
	}
	if _, ok = g2.kind.(gradientLinear); !ok {
		t.Fatal()
	}
	g3, ok := out.definitions.paintServers["RadialGradient1"].(gradient)
	if !ok {
		t.Fatal()
	}
	if _, ok = g3.kind.(gradientRadial); !ok {
		t.Fatal()
	}
	g4, ok := out.definitions.paintServers["RadialGradient2"].(gradient)
	if !ok {
		t.Fatal()
	}
	if _, ok = g4.kind.(gradientRadial); !ok {
		t.Fatal()
	}
}

// TODO: test pattern
