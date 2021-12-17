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

func stringToXMLArgs(s string) []xml.Attr {
	out := struct {
		AllAttrs []xml.Attr `xml:",any,attr"`
	}{}
	err := xml.Unmarshal([]byte(fmt.Sprintf("<p %s></p>", s)), &out)
	if err != nil {
		panic(err)
	}
	return out.AllAttrs
}

func Test_parseNodeAttributes(t *testing.T) {
	tests := []struct {
		args    []xml.Attr
		want    nodeAttributes
		wantErr bool
	}{
		{
			args:    stringToXMLArgs(`width="50px" height="10pt" font-size="2em"`),
			want:    nodeAttributes{opacity: 1, fontSize: value{2, Em}, width: value{50, Px}, height: value{10, Pt}},
			wantErr: false,
		},
		{
			args:    stringToXMLArgs(`visibility="hidden"`),
			want:    nodeAttributes{opacity: 1, noVisible: true},
			wantErr: false,
		},
		{
			args:    stringToXMLArgs(`mask="url(#myMask)"`),
			want:    nodeAttributes{opacity: 1, mask: "myMask"},
			wantErr: false,
		},
		{
			args:    stringToXMLArgs(`marker="url(#m1)" marker-mid="url(#m2)"`),
			want:    nodeAttributes{opacity: 1, marker: "m1", markerPosition: [3]string{"", "m2", ""}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		gotNode, err := parseNodeAttributes(tt.args)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseNodeAttributes() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if !reflect.DeepEqual(gotNode, tt.want) {
			t.Errorf("parseNodeAttributes() = %v, want %v", gotNode, tt.want)
		}
	}
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
