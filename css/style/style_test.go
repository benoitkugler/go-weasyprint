package style

import (
	"testing"

	"github.com/benoitkugler/go-weasyprint/css/parser"
	"github.com/benoitkugler/go-weasyprint/utils"
)

func TestLoadStyleSheets(t *testing.T) {
	LoadStyleSheet("../..")
}

func TestDescriptors(t *testing.T) {
	stylesheet := parser.ParseStylesheet2([]byte("@font-face{}"), false, false)
	logs := utils.CaptureLogs()
	var descriptors []string
	preprocessStylesheet("print", "http://wp.org/foo/", stylesheet, nil, nil, nil,
		&descriptors, nil, false)
	if len(descriptors) > 0 {
		t.Fatalf("expected empty descriptors, got %v", descriptors)
	}
	logs.CheckEqual([]string{
		`Missing src descriptor in "@font-face" rule at 1:1`,
	}, t)

	stylesheet = parser.ParseStylesheet2([]byte("@font-face{src: url(test.woff)}"), false, false)
	logs = utils.CaptureLogs()
	preprocessStylesheet("print", "http://wp.org/foo/", stylesheet, nil, nil, nil,
		&descriptors, nil, false)
	if len(descriptors) > 0 {
		t.Fatalf("expected empty descriptors, got %v", descriptors)
	}
	logs.CheckEqual([]string{
		`Missing font-family descriptor in "@font-face" rule at 1:1`,
	}, t)

	stylesheet = parser.ParseStylesheet2([]byte("@font-face{font-family: test}"), false, false)
	logs = utils.CaptureLogs()
	preprocessStylesheet("print", "http://wp.org/foo/", stylesheet, nil, nil, nil,
		&descriptors, nil, false)
	if len(descriptors) > 0 {
		t.Fatalf("expected empty descriptors, got %v", descriptors)
	}
	logs.CheckEqual([]string{
		`Missing src descriptor in "@font-face" rule at 1:1`,
	}, t)

	stylesheet = parser.ParseStylesheet2([]byte("@font-face { font-family: test; src: wrong }"), false, false)
	logs = utils.CaptureLogs()
	preprocessStylesheet("print", "http://wp.org/foo/", stylesheet, nil, nil, nil,
		&descriptors, nil, false)
	if len(descriptors) > 0 {
		t.Fatalf("expected empty descriptors, got %v", descriptors)
	}
	logs.CheckEqual([]string{
		"Ignored `src: wrong ` at 1:33, invalid or unsupported values for a known CSS property.",
		`Missing src descriptor in "@font-face" rule at 1:1`,
	}, t)

	stylesheet = parser.ParseStylesheet2([]byte("@font-face { font-family: good, bad; src: url(test.woff) }"), false, false)
	logs = utils.CaptureLogs()
	preprocessStylesheet("print", "http://wp.org/foo/", stylesheet, nil, nil, nil,
		&descriptors, nil, false)
	if len(descriptors) > 0 {
		t.Fatalf("expected empty descriptors, got %v", descriptors)
	}
	logs.CheckEqual([]string{
		"Ignored `font-family: good, bad` at 1:14, invalid or unsupported values for a known CSS property.",
		`Missing font-family descriptor in "@font-face" rule at 1:1`,
	}, t)

	stylesheet = parser.ParseStylesheet2([]byte("@font-face { font-family: good, bad; src: really bad }"), false, false)
	logs = utils.CaptureLogs()
	preprocessStylesheet("print", "http://wp.org/foo/", stylesheet, nil, nil, nil,
		&descriptors, nil, false)
	if len(descriptors) > 0 {
		t.Fatalf("expected empty descriptors, got %v", descriptors)
	}
	logs.CheckEqual([]string{
		"Ignored `font-family: good, bad` at 1:14, invalid or unsupported values for a known CSS property.",
		"Ignored `src: really bad ` at 1:38, invalid or unsupported values for a known CSS property.",
		`Missing src descriptor in "@font-face" rule at 1:1`,
	}, t)
}
