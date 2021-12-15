package svg

import (
	"reflect"
	"testing"
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
	}
	for _, tt := range tests {
		gotPoints, err := parsePoints(tt.dataPoints)
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
		gotPoints, err := parseFloatList(tt.args)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseFloatList() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if !reflect.DeepEqual(gotPoints, tt.wantPoints) {
			t.Errorf("parseFloatList() = %v, want %v", gotPoints, tt.wantPoints)
		}
	}
}
