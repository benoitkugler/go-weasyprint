// Package svg implements parsing of SVG images.
// It transforms SVG text files into an in-memory structure
// that is easy to draw.
// CSS is supported via the style and cascadia packages.
package svg

import (
	"encoding/xml"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func Test_parseViewbox(t *testing.T) {
	tests := []struct {
		args    string
		want    [4]Fl
		wantErr bool
	}{
		{"0 0 100 100", [4]Fl{0, 0, 100, 100}, false},
		{"0 0 100", [4]Fl{}, true},
	}
	for _, tt := range tests {
		got, err := parseViewbox(tt.args)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseViewbox() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("parseViewbox() = %v, want %v", got, tt.want)
		}
	}
}

func stringToXMLArgs(s string) nodeAttributes {
	out := struct {
		AllAttrs []xml.Attr `xml:",any,attr"`
	}{}
	err := xml.Unmarshal([]byte(fmt.Sprintf("<p %s></p>", s)), &out)
	if err != nil {
		panic(err)
	}
	return newNodeAttributes(out.AllAttrs)
}

func assertEqual(t *testing.T, exp, got interface{}) {
	t.Helper()

	if !reflect.DeepEqual(exp, got) {
		t.Fatalf("expected %v, got %v", exp, got)
	}
}

func Test_parseNodeAttributes(t *testing.T) {
	attrs := stringToXMLArgs(`width="50px" height="10pt" font-size="2em"`)
	got, _ := attrs.fontSize()
	assertEqual(t, value{2, Em}, got)
	got, _ = attrs.width()
	assertEqual(t, value{50, Px}, got)
	got, _ = attrs.height()
	assertEqual(t, value{10, Pt}, got)

	attrs = stringToXMLArgs(`visibility="hidden"`)
	assertEqual(t, true, attrs.noVisible())

	attrs = stringToXMLArgs(`mask="url(#myMask)"`)
	assertEqual(t, "myMask", attrs.mask())

	attrs = stringToXMLArgs(`marker="url(#m1)" marker-mid="url(#m2)"`)
	assertEqual(t, "m1", attrs.marker())
	assertEqual(t, "m2", attrs.markerMid())
}

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
