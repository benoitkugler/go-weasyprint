// Package svg implements parsing of SVG images.
// It transforms SVG text files into an in-memory structure
// that is easy to draw.
// CSS is supported via the style and cascadia packages.
package svg

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func parseIcon(t *testing.T, iconPath string) {
	f, err := os.Open(iconPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	_, err = Parse(f, "")
	if err != nil {
		t.Error(err)
	}
}

func TestCorpus(t *testing.T) {
	for _, p := range []string{
		"beach", "cape", "iceberg", "island",
		"mountains", "sea", "trees", "village",
	} {
		parseIcon(t, "testdata/landscapeIcons/"+p+".svg")
	}

	for _, p := range []string{
		"astronaut", "jupiter", "lander", "school-bus", "telescope", "content-cut-light", "defs",
		"24px",
	} {
		parseIcon(t, "testdata/testIcons/"+p+".svg")
	}

	for _, p := range []string{
		"OpacityStrokeDashTest.svg",
		"OpacityStrokeDashTest2.svg",
		"OpacityStrokeDashTest3.svg",
		"TestShapes.svg",
		"TestShapes2.svg",
		"TestShapes3.svg",
		"TestShapes4.svg",
		"TestShapes5.svg",
		"TestShapes6.svg",
	} {
		parseIcon(t, "testdata/"+p)
	}
}

func TestPercentages(t *testing.T) {
	parseIcon(t, "testdata/TestPercentages.svg")
}

func TestInvalidXML(t *testing.T) {
	_, err := Parse(strings.NewReader("dummy"), "")
	if err == nil {
		t.Fatal("expected error on invalid input")
	}
	_, err = Parse(strings.NewReader("<not-svg></not-svg>"), "")
	if err == nil {
		t.Fatal("expected error on invalid input")
	}
}

func TestBuildTree(t *testing.T) {
	input := `
	<svg viewBox="0 0 10 10">
	<style>
		path {
			color: red;
		}
	</style>
	<path style="fontsize: 10px">AA</path>

	</svg>
	`
	tree, err := buildSVGTree(strings.NewReader(input), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.root.children) != 1 {
		t.Fatalf("unexpected children %v", tree.root.children)
	}
	p := tree.root.children[0]
	if !reflect.DeepEqual(p.attrs, nodeAttributes{"fontsize": "10px", "color": "red"}) {
		t.Fatalf("unexpected attributes %v", p.attrs)
	}
}

func TestParseDefs(t *testing.T) {
	input := `
	<svg viewBox="0 0 10 10" xmlns="http://www.w3.org/2000/svg"
	xmlns:xlink="http://www.w3.org/1999/xlink">
	<!-- Some graphical objects to use -->
	<defs>
		<circle id="myCircle" cx="0" cy="0" r="5" />

		<linearGradient id="myGradient" gradientTransform="rotate(90)">
		<stop offset="20%" stop-color="gold" />
		<stop offset="90%" stop-color="red" />
		</linearGradient>
	</defs>

	<!-- using my graphical objects -->
	<use x="5" y="5" href="#myCircle" fill="url('#myGradient')" />
	</svg>
	`
	img, err := Parse(strings.NewReader(input), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(img.defs) != 2 {
		t.Fatal("defs")
	}
	if c, has := img.defs["myCircle"]; !has || len(c.children) != 0 {
		t.Fatal("defs circle")
	}
	if c, has := img.defs["myGradient"]; !has || len(c.children) != 2 {
		t.Fatal("defs gradient")
	}
}
