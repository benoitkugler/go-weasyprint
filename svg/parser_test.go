package svg

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func Test_parsePoints(t *testing.T) {
	tests := []struct {
		dataPoints string
		wantPoints []Fl
		wantErr    bool
	}{
		{"50 160 55 180.2 70 180", []Fl{50, 160, 55, 180.2, 70, 180}, false},
		{"153.423,21.442,12.3e5,", []Fl{153.423, 21.442, 12.3e5}, false},
		{"-11.231-1.388-22.118-3.789-32.621", []Fl{-11.231, -1.388, -22.118, -3.789, -32.621}, false},
		{"7px 8% 10 px 72pt", []Fl{7, 8, 10, 72}, false}, // units are ignored
		{"15,45.7e", nil, true},
		{"50,0 21,90 98,35 2,35 79,90", []Fl{50, 0, 21, 90, 98, 35, 2, 35, 79, 90}, false},
	}
	for _, tt := range tests {
		gotPoints, err := parsePoints(tt.dataPoints, nil)
		if (err != nil) != tt.wantErr {
			t.Errorf("getPoints() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if !reflect.DeepEqual(gotPoints, tt.wantPoints) {
			t.Errorf("getPoints() = %v, want %v", gotPoints, tt.wantPoints)
		}
	}
}

func Test_parseURLFragment(t *testing.T) {
	tests := []struct {
		args string
		want string
	}{
		{"www.google.com#test", "test"},
		{"url(www.google.com#test)", "test"},
		{"url('www.google.com#test')", "test"},
		{`url("www.google.com#test")`, "test"},
		{"www.google.com", ""},
		{"789", ""},
	}
	for _, tt := range tests {
		if got := parseURLFragment(tt.args); got != tt.want {
			t.Errorf("parseURLFragment() = %v, want %v", got, tt.want)
		}
	}
}

func Test_parseFloatList(t *testing.T) {
	tests := []struct {
		args       string
		wantPoints []value
		wantErr    bool
	}{
		{"7px 8% 10px 72pt", []value{{7, Px}, {8, Perc}, {10, Px}, {72, Pt}}, false},
	}
	for _, tt := range tests {
		gotPoints, err := parseValues(tt.args)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseFloatList() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if !reflect.DeepEqual(gotPoints, tt.wantPoints) {
			t.Errorf("parseFloatList() = %v, want %v", gotPoints, tt.wantPoints)
		}
	}
}

func Test_value_resolve(t *testing.T) {
	type args struct {
		fontSize            Fl
		percentageReference Fl
	}
	tests := []struct {
		value value
		args  args
		want  Fl
	}{
		{value: value{u: Px, v: 10}, args: args{}, want: 10},
		{value: value{u: Pt, v: 72}, args: args{}, want: 96},
		{value: value{u: Perc, v: 50}, args: args{percentageReference: 40}, want: 20},
		{value: value{u: Em, v: 10}, args: args{fontSize: 20}, want: 200},
		{value: value{u: Ex, v: 10}, args: args{fontSize: 20}, want: 100},
	}
	for _, tt := range tests {
		if got := tt.value.resolve(tt.args.fontSize, tt.args.percentageReference); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("value.resolve() = %v, want %v", got, tt.want)
		}
	}
}

func Test_parseViewbox(t *testing.T) {
	tests := []struct {
		args    string
		want    Rectangle
		wantErr bool
	}{
		{"0 0 100 100", Rectangle{0, 0, 100, 100}, false},
		{"0 0 100", Rectangle{}, true},
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
	node, err := html.Parse(strings.NewReader(fmt.Sprintf("<html %s></html>", s)))
	if err != nil {
		panic(err)
	}
	return newNodeAttributes(node.FirstChild.Attr)
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
	got, _ = parseValue(attrs["width"])
	assertEqual(t, value{50, Px}, got)
	got, _ = parseValue(attrs["height"])
	assertEqual(t, value{10, Pt}, got)

	attrs = stringToXMLArgs(`visibility="hidden"`)
	assertEqual(t, false, attrs.visible())

	attrs = stringToXMLArgs(`mask="url(#myMask)"`)
	assertEqual(t, "myMask", parseURLFragment(attrs["mask"]))

	attrs = stringToXMLArgs(`marker="url(#m1)" marker-mid="url(#m2)"`)
	assertEqual(t, "m1", parseURLFragment(attrs["marker"]))
	assertEqual(t, "m2", parseURLFragment(attrs["marker-mid"]))
}

func Test_parseTransform(t *testing.T) {
	tests := []struct {
		args    string
		wantOut []transform
		wantErr bool
	}{
		{
			`rotate(-10 50 100)
                translate(-36 45.5%)
                skewX(40pt)
                scale(1em 0.5)
				matrix(1,2,3,4,5,6)
				`,
			[]transform{
				{rotateWithOrigin, [6]value{
					{-10, Px}, {50, Px}, {100, Px},
				}},
				{translate, [6]value{
					{-36, Px}, {45.5, Perc},
				}},
				{skew, [6]value{
					{40, Pt},
				}},
				{scale, [6]value{
					{1, Em}, {0.5, Px},
				}},
				{customMatrix, [6]value{
					{1, Px}, {2, Px}, {3, Px}, {4, Px}, {5, Px}, {6, Px},
				}},
			},
			false,
		},
		{
			`rotate(50 100)`,
			nil,
			true,
		},
		{
			`scale(20 50 100)`,
			nil,
			true,
		},
		{
			`translate(20 50 100)`,
			nil,
			true,
		},
		{
			` `,
			nil,
			false,
		},
	}
	for _, tt := range tests {
		gotOut, err := parseTransform(tt.args)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseTransform() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if !reflect.DeepEqual(gotOut, tt.wantOut) {
			t.Errorf("parseTransform() = %v, want %v", gotOut, tt.wantOut)
		}
	}
}
