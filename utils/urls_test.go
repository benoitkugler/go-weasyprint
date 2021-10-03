package utils

import (
	"reflect"
	"testing"
)

func TestFetchUrlData(t *testing.T) {
	testData := func(input, expected string) {
		tmp, err := parseDataURL([]byte(input))
		if err != nil {
			t.Fatal(err)
		}
		res, err := tmp.toResource("")
		if err != nil {
			t.Fatal(err)
		}
		b, err := decodeToUtf8(res.Content, res.ProtocolEncoding)
		if err != nil {
			t.Fatal(err)
		}
		if got := string(b); got != expected {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	}

	testData("data:image/svg+xml,<svg></svg>", "<svg></svg>")
	testData("data:text/css;charset=ASCII,a%7Bcolor%3AcurrentColor%7D", "a{color:currentColor}")
	testData(`data:text/css;charset=utf-16le;base64,bABpAHsAYwBvAGwAbwByADoAcgBlAGQAfQA=`, "li{color:red}")
	testData("data:,ul {border-width: 1000px !important}", "ul {border-width: 1000px !important}")
}

func TestParseDataURL(t *testing.T) {
	type testCase struct {
		dataURL    string
		parsedData dataURI
	}

	testCases := []testCase{
		{
			"data:image/gif;base64,XXX",
			dataURI{
				mimeType: "image/gif", isBase64: true,
				data:   []byte("XXX"),
				params: make(map[string]string),
			},
		},
		{
			`data:,A%20brief%20note`,
			dataURI{
				mimeType: "text/plain", data: []byte(`A%20brief%20note`),
				params: map[string]string{"charset": "US-ASCII"},
			},
		},
		{
			`data:text/plain;charset=iso-8859-7,%be%fg%be`,
			dataURI{
				mimeType: "text/plain", data: []byte("%be%fg%be"),
				params: map[string]string{"charset": "iso-8859-7"},
			},
		},
		{
			`data:application/vnd-xxx-query,select_vcount,fcol_from_fieldtable/local`,
			dataURI{
				mimeType: "application/vnd-xxx-query",
				data:     []byte("select_vcount,fcol_from_fieldtable/local"),
				params:   make(map[string]string),
			},
		},
		{`data:,`, dataURI{
			mimeType: "text/plain",
			params:   map[string]string{"charset": "US-ASCII"},
			data:     []byte{},
		}},
		{`data:boo/foo;,`, dataURI{
			mimeType: "boo/foo",
			params:   make(map[string]string),
			data:     []byte{},
		}},
	}

	for _, test := range testCases {
		t.Run(test.dataURL, func(t *testing.T) {
			result, _ := parseDataURL([]byte(test.dataURL))
			if !reflect.DeepEqual(result, test.parsedData) {
				t.Errorf("ParseDataURL expected %+v\n, got %+v", test.parsedData, result)
			}
		})
	}
}
