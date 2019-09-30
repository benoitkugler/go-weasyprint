package tree

import (
	"fmt"
	"log"
	"math"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	pr "github.com/benoitkugler/go-weasyprint/style/properties"

	"github.com/benoitkugler/cascadia"
	"github.com/benoitkugler/go-weasyprint/style/parser"
	"github.com/benoitkugler/go-weasyprint/utils"
)

// Test the CSS parsing, cascade, inherited && computed values.

var testUAStylesheet CSS

func init() {
	var err error
	testUAStylesheet, err = newCSS(InputFilename("tests_ua.css"))
	if err != nil {
		log.Fatal(err)
	}
}

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

type fakeHTML struct {
	HTML
	customUA CSS
}

func (f fakeHTML) UAStyleSheet() CSS {
	if f.customUA.IsNone() {
		return testUAStylesheet
	}
	return f.customUA
}

func resourceFilename(s string) string {
	return filepath.Join("../../resources_test", s)
}

// equivalent to python s.rsplit(sep, -1)[-1]
func rsplit(s, sep string) string {
	chunks := strings.Split(s, sep)
	return chunks[len(chunks)-1]
}

//@assertNoLogs
func TestFindStylesheets(t *testing.T) {
	capt := utils.CaptureLogs()
	html_, err := newHtml(InputFilename(resourceFilename("doc1.html")))
	if err != nil {
		t.Fatal(err)
	}
	html := fakeHTML{HTML: *html_}
	sheets := findStylesheets(html.root, "print", utils.DefaultUrlFetcher, html.baseUrl, nil, nil)

	if len(sheets) != 2 {
		t.Errorf("expected 2 sheets, got %d", len(sheets))
	}
	// Also test that stylesheets are in tree order
	var got [2]string
	for i, s := range sheets {
		got[i] = rsplit(rsplit(s.baseUrl, "/"), ",")
	}
	exp := [2]string{"a%7Bcolor%3AcurrentColor%7D", "doc1.html"}
	if got != exp {
		t.Errorf("expected %v got %v", exp, got)
	}

	var (
		rules      []cascadia.Sel
		pagesRules []pageRule
	)
	for _, sheet := range sheets {
		for _, sheetRules := range sheet.matcher {
			rules = append(rules, sheetRules.selector...)
		}
		for _, rule := range sheet.pageRules {
			pagesRules = append(pagesRules, rule)
		}
	}
	if len(rules)+len(pagesRules) != 10 {
		t.Errorf("expected 10 rules, got %d", len(rules)+len(pagesRules))
	}
	capt.AssertNoLogs(t)
	// TODO: test that the values are correct too
}

//@assertNoLogs
func TestExpandShorthands(t *testing.T) {
	capt := utils.CaptureLogs()
	sheet, err := newCSS(InputFilename(resourceFilename("sheet2.css")))
	if err != nil {
		t.Fatal(err)
	}
	var sels []cascadia.Sel
	for _, match := range sheet.matcher {
		sels = append(sels, match.selector...)
	}
	if len(sels) != 1 {
		t.Fatalf("expected ['li'] got %v", sels)
	}
	//TODO: update cascadia to support this
	// if sels[0].String() != "li" {
	// 	t.Errorf("expected 'li' got %s", sels[0].String())
	// }

	m := (sheet.matcher)[0].declarations
	if m[0].Name != "margin_bottom" {
		t.Errorf("expected margin_bottom got %s", m[0].Name)
	}
	if (m[0].Value.AsCascaded().AsCss().(pr.Value) != pr.Dimension{Value: 3, Unit: pr.Em}.ToValue()) {
		t.Errorf("expected got %v", m[0].Value)
	}
	if m[1].Name != "margin_top" {
		t.Errorf("expected margin_top got %s", m[1].Name)
	}
	if (m[1].Value.AsCascaded().AsCss().(pr.Value) != pr.Dimension{Value: 2, Unit: pr.Em}.ToValue()) {
		t.Errorf("expected got %v", m[1].Value)
	}
	if m[2].Name != "margin_right" {
		t.Errorf("expected margin_right got %s", m[2].Name)
	}
	if (m[2].Value.AsCascaded().AsCss().(pr.Value) != pr.Dimension{Value: 0, Unit: pr.Scalar}.ToValue()) {
		t.Errorf("expected got %v", m[2].Value)
	}
	if m[3].Name != "margin_bottom" {
		t.Errorf("expected margin_bottom got %s", m[3].Name)
	}
	if (m[3].Value.AsCascaded().AsCss().(pr.Value) != pr.Dimension{Value: 2, Unit: pr.Em}.ToValue()) {
		t.Errorf("expected got %v", m[3].Value)
	}
	if m[4].Name != "margin_left" {
		t.Errorf("expected margin_left got %s", m[4].Name)
	}
	if (m[4].Value.AsCascaded().AsCss().(pr.Value) != pr.Dimension{Value: 0, Unit: pr.Scalar}.ToValue()) {
		t.Errorf("expected got %v", m[4].Value)
	}
	if m[5].Name != "margin_left" {
		t.Errorf("expected margin_left got %s", m[5].Name)
	}
	if (m[5].Value.AsCascaded().AsCss().(pr.Value) != pr.Dimension{Value: 4, Unit: pr.Em}.ToValue()) {
		t.Errorf("expected got %v", m[5].Value)
	}
	capt.AssertNoLogs(t)
	// TODO: test that the values are correct too
}

func assertProp(t *testing.T, got pr.Properties, name string, expected pr.CssProperty) {
	g := got[name]
	if !reflect.DeepEqual(g, expected) {
		t.Fatalf("%s - expected %v got %v", name, expected, g)
	}
}

//@assertNoLogs
func TestAnnotateDocument(t *testing.T) {
	capt := utils.CaptureLogs()
	document_, err := newHtml(InputFilename(resourceFilename("doc1.html")))
	if err != nil {
		t.Fatal(err)
	}
	document := fakeHTML{HTML: *document_}
	document.customUA, err = newCSS(InputFilename(resourceFilename("mini_ua.css")))
	if err != nil {
		t.Fatal(err)
	}

	userStylesheet, err := newCSS(InputFilename(resourceFilename("user.css")))
	if err != nil {
		t.Fatal(err)
	}

	styleFor := GetAllComputedStyles(document, []CSS{userStylesheet}, false, nil, nil, nil)
	// Element objects behave as lists of their children
	body := document.root.NodeChildren(true)[1]
	children := body.NodeChildren(true)
	h1_, p_, ul_, div_ := children[0], children[1], children[2], children[3]
	li0_ := ul_.NodeChildren(true)[0]
	a_ := li0_.NodeChildren(true)[0]
	span1_ := div_.NodeChildren(true)[0]
	span2_ := span1_.NodeChildren(true)[0]

	h1 := styleFor.Get(h1_, "")
	p := styleFor.Get(p_, "")
	ul := styleFor.Get(ul_, "")
	li0 := styleFor.Get(li0_, "")
	div := styleFor.Get(div_, "")
	after := styleFor.Get(a_, "after")
	a := styleFor.Get(a_, "")
	span1 := styleFor.Get(span1_, "")
	span2 := styleFor.Get(span2_, "")

	u, err := utils.Path2url(resourceFilename("logo_small.png"))
	if err != nil {
		t.Fatal(err)
	}
	assertProp(t, h1, "background_image", pr.Images{pr.UrlImage(u)})

	assertProp(t, h1, "font_weight", pr.IntString{Int: 700})
	assertProp(t, h1, "font_size", pr.FToV(40)) // 2em

	// x-large * initial = 3/2 * 16 = 24
	assertProp(t, p, "margin_top", pr.Dimension{Value: 24, Unit: pr.Px}.ToValue())
	assertProp(t, p, "margin_right", pr.Dimension{Value: 0, Unit: pr.Px}.ToValue())
	assertProp(t, p, "margin_bottom", pr.Dimension{Value: 24, Unit: pr.Px}.ToValue())
	assertProp(t, p, "margin_left", pr.Dimension{Value: 0, Unit: pr.Px}.ToValue())
	assertProp(t, p, "background_color", pr.CurrentColor)

	// 2em * 1.25ex = 2 * 20 * 1.25 * 0.8 = 40
	// 2.5ex * 1.25ex = 2.5 * 0.8 * 20 * 1.25 * 0.8 = 40
	// TODO: ex unit doesn"t work with @font-face fonts, see computedValues.py
	// assert ul["marginTop"] , pr.Dimension {Value:40,Unit: pr.Px}
	// .ToValue()assert ul["marginRight"] , pr.Dimension {Value:40,Unit: pr.Px}
	// .ToValue()assert ul["marginBottom"] , pr.Dimension {Value:40,Unit: pr.Px}
	// .ToValue()assert ul["marginLeft"] , pr.Dimension {Value:40,Unit: pr.Px}

	assertProp(t, ul, "font_weight", pr.IntString{Int: 400})
	// thick = 5px, 0.25 inches = 96*.25 = 24px
	assertProp(t, ul, "border_top_width", pr.FToV(0))
	assertProp(t, ul, "border_right_width", pr.FToV(5))
	assertProp(t, ul, "border_bottom_width", pr.FToV(0))
	assertProp(t, ul, "border_left_width", pr.FToV(24))

	assertProp(t, li0, "font_weight", pr.IntString{Int: 700})
	assertProp(t, li0, "font_size", pr.FToV(8))                                      // 6pt)
	assertProp(t, li0, "margin_top", pr.Dimension{Value: 16, Unit: pr.Px}.ToValue()) // 2em)
	assertProp(t, li0, "margin_right", pr.Dimension{Value: 0, Unit: pr.Px}.ToValue())
	assertProp(t, li0, "margin_bottom", pr.Dimension{Value: 16, Unit: pr.Px}.ToValue())
	assertProp(t, li0, "margin_left", pr.Dimension{Value: 32, Unit: pr.Px}.ToValue()) // 4em)

	assertProp(t, a, "text_decoration_line", pr.NDecorations{Decorations: pr.NewSet("underline")})
	assertProp(t, a, "font_weight", pr.IntString{Int: 900})
	assertProp(t, a, "font_size", pr.FToV(24)) // 300% of 8px)
	assertProp(t, a, "padding_top", pr.Dimension{Value: 1, Unit: pr.Px}.ToValue())
	assertProp(t, a, "padding_right", pr.Dimension{Value: 2, Unit: pr.Px}.ToValue())
	assertProp(t, a, "padding_bottom", pr.Dimension{Value: 3, Unit: pr.Px}.ToValue())
	assertProp(t, a, "padding_left", pr.Dimension{Value: 4, Unit: pr.Px}.ToValue())
	assertProp(t, a, "border_top_width", pr.FToV(42))
	assertProp(t, a, "border_bottom_width", pr.FToV(42))

	assertProp(t, a, "color", pr.NewColor(1, 0, 0, 1))
	assertProp(t, a, "border_top_color", pr.CurrentColor)

	assertProp(t, div, "font_size", pr.FToV(40))                                    // 2 * 20px)
	assertProp(t, span1, "width", pr.Dimension{Value: 160, Unit: pr.Px}.ToValue())  // 10 * 16px (root default is 16px))
	assertProp(t, span1, "height", pr.Dimension{Value: 400, Unit: pr.Px}.ToValue()) // 10 * (2 * 20px))
	assertProp(t, span2, "font_size", pr.FToV(32))

	// The href attr should be as in the source, not made absolute.
	assertProp(t, after, "background_color", pr.NewColor(1, 0, 0, 1))
	assertProp(t, after, "border_top_width", pr.FToV(42))
	assertProp(t, after, "border_bottom_width", pr.FToV(3))
	assertProp(t, after, "content", pr.SContent{Contents: pr.ContentProperties{{Type: "string", Content: pr.String(" [")}, {Type: "string", Content: pr.String("home.html")}, {Type: "string", Content: pr.String("]")}}})

	// TODO: much more tests here: test that origin and selector precedence
	// and inheritance are correct…
	capt.AssertNoLogs(t)
}

//@assertNoLogs
func TestPage(t *testing.T) {
	capt := utils.CaptureLogs()
	document_, err := newHtml(InputFilename(resourceFilename("doc1.html")))
	if err != nil {
		t.Fatal(err)
	}
	document := fakeHTML{HTML: *document_}
	css, err := newCSS(InputString(`
		html { color: red }
		@page { margin: 10px }
		@page :right {
		color: blue;
		margin-bottom: 12pt;
		font-size: 20px;
		@top-left { width: 10em }
		@top-right { font-size: 10px}
	}`))
	if err != nil {
		t.Fatal(err)
	}
	styleFor := GetAllComputedStyles(document, []CSS{css}, false, nil, nil, nil)

	pageType := utils.PageElement{Side: "left", First: true, Blank: false, Index: 0, Name: ""}
	styleFor.SetPageTypeComputedStyles(pageType, document)
	style := styleFor.Get(pageType, "")
	assertProp(t, style, "margin_top", pr.Dimension{Value: 5, Unit: pr.Px}.ToValue())
	assertProp(t, style, "margin_left", pr.Dimension{Value: 10, Unit: pr.Px}.ToValue())
	assertProp(t, style, "margin_bottom", pr.Dimension{Value: 10, Unit: pr.Px}.ToValue())
	assertProp(t, style, "color", pr.NewColor(1, 0, 0, 1)) // red, inherited from html

	pageType = utils.PageElement{Side: "right", First: true, Blank: false, Index: 0, Name: ""}
	styleFor.SetPageTypeComputedStyles(pageType, document)
	style = styleFor.Get(pageType, "")
	assertProp(t, style, "margin_top", pr.Dimension{Value: 5, Unit: pr.Px}.ToValue())
	assertProp(t, style, "margin_left", pr.Dimension{Value: 10, Unit: pr.Px}.ToValue())
	assertProp(t, style, "margin_bottom", pr.Dimension{Value: 16, Unit: pr.Px}.ToValue())
	assertProp(t, style, "color", pr.NewColor(0, 0, 1, 1)) // blue

	pageType = utils.PageElement{Side: "left", First: false, Blank: false, Index: 1, Name: ""}
	styleFor.SetPageTypeComputedStyles(pageType, document)
	style = styleFor.Get(pageType, "")
	assertProp(t, style, "margin_top", pr.Dimension{Value: 10, Unit: pr.Px}.ToValue())
	assertProp(t, style, "margin_left", pr.Dimension{Value: 10, Unit: pr.Px}.ToValue())
	assertProp(t, style, "margin_bottom", pr.Dimension{Value: 10, Unit: pr.Px}.ToValue())
	assertProp(t, style, "color", pr.NewColor(1, 0, 0, 1)) // red, inherited from html

	pageType = utils.PageElement{Side: "right", First: false, Blank: false, Index: 1, Name: ""}
	styleFor.SetPageTypeComputedStyles(pageType, document)
	style = styleFor.Get(pageType, "")
	assertProp(t, style, "margin_top", pr.Dimension{Value: 10, Unit: pr.Px}.ToValue())
	assertProp(t, style, "margin_left", pr.Dimension{Value: 10, Unit: pr.Px}.ToValue())
	assertProp(t, style, "margin_bottom", pr.Dimension{Value: 16, Unit: pr.Px}.ToValue())
	assertProp(t, style, "color", pr.NewColor(0, 0, 1, 1)) // blue

	pageType = utils.PageElement{Side: "left", First: true, Blank: false, Index: 0, Name: ""}
	styleFor.SetPageTypeComputedStyles(pageType, document)
	style = styleFor.Get(pageType, "@top-left")
	if style != nil {
		t.Fatal("expected empty (nil) style")
	}

	pageType = utils.PageElement{Side: "right", First: true, Blank: false, Index: 0, Name: ""}
	styleFor.SetPageTypeComputedStyles(pageType, document)
	style = styleFor.Get(pageType, "@top-left")
	assertProp(t, style, "font_size", pr.FToV(20)) // inherited from @page
	assertProp(t, style, "width", pr.Dimension{Value: 200, Unit: pr.Px}.ToValue())

	pageType = utils.PageElement{Side: "right", First: true, Blank: false, Index: 0, Name: ""}
	styleFor.SetPageTypeComputedStyles(pageType, document)
	style = styleFor.Get(pageType, "@top-right")
	assertProp(t, style, "font_size", pr.FToV(10))
	capt.AssertNoLogs(t)
}

type testPageSelector struct {
	sel string
	out []pageSelector
}

var tests = []testPageSelector{
	{sel: "@page {}", out: []pageSelector{
		{Specificity: cascadia.Specificity{0, 0, 0}},
	}},
	{sel: "@page :left {}", out: []pageSelector{
		{Side: "left", Specificity: cascadia.Specificity{0, 0, 1}},
	}},
	{sel: "@page:first:left {}", out: []pageSelector{
		{Side: "left", First: true, Specificity: cascadia.Specificity{0, 1, 1}},
	}},
	{sel: "@page pagename {}", out: []pageSelector{
		{Name: "pagename", Specificity: cascadia.Specificity{1, 0, 0}},
	}},
	{sel: "@page pagename:first:right:blank {}", out: []pageSelector{
		{Side: "right", Blank: true, First: true, Name: "pagename", Specificity: cascadia.Specificity{1, 2, 1}},
	}},
	{sel: "@page pagename, :first {}", out: []pageSelector{
		{Name: "pagename", Specificity: cascadia.Specificity{1, 0, 0}},
		{First: true, Specificity: cascadia.Specificity{0, 1, 0}},
	}},
	{sel: "@page :first:first {}", out: []pageSelector{
		{First: true, Specificity: cascadia.Specificity{0, 2, 0}},
	}},
	{sel: "@page :left:left {}", out: []pageSelector{
		{Side: "left", Specificity: cascadia.Specificity{0, 0, 2}},
	}},
	{sel: "@page :nth(2) {}", out: []pageSelector{
		{Index: pageIndex{A: 0, B: 2}, Specificity: cascadia.Specificity{0, 1, 0}},
	}},
	{sel: "@page :nth(2n + 4) {}", out: []pageSelector{
		{Index: pageIndex{A: 2, B: 4}, Specificity: cascadia.Specificity{0, 1, 0}},
	}},
	{sel: "@page :nth(3n) {}", out: []pageSelector{
		{Index: pageIndex{A: 3, B: 0}, Specificity: cascadia.Specificity{0, 1, 0}},
	}},
	{sel: "@page :nth( n+2 ) {}", out: []pageSelector{
		{Index: pageIndex{A: 1, B: 2}, Specificity: cascadia.Specificity{0, 1, 0}},
	}},
	{sel: "@page :nth(even) {}", out: []pageSelector{
		{Index: pageIndex{A: 2, B: 0}, Specificity: cascadia.Specificity{0, 1, 0}},
	}},
	{sel: "@page pagename:nth(2) {}", out: []pageSelector{
		{Name: "pagename", Index: pageIndex{A: 0, B: 2}, Specificity: cascadia.Specificity{1, 1, 0}},
	}},
	{sel: "@page page page {}"},
	{sel: "@page :left page {}"},
	{sel: "@page :left, {}"},
	{sel: "@page , {}"},
	{sel: "@page :left, test, {}"},
	{sel: "@page :wrong {}"},
	{sel: "@page :left:wrong {}"},
	{sel: "@page :left:right {}"},
}

func TestPageSelectors(t *testing.T) {
	capt := utils.CaptureLogs()
	for _, te := range tests {
		atRule_ := parser.ParseStylesheet2([]byte(te.sel), false, false)[0]
		atRule, ok := atRule_.(parser.QualifiedRule)
		if !ok {
			atRule = atRule_.(parser.AtRule).QualifiedRule
		}
		res := parsePageSelectors(atRule)
		if !reflect.DeepEqual(res, te.out) {
			t.Fatalf("%s : expected %v got %v", te.sel, te.out, res)
		}
	}
	capt.AssertNoLogs(t)
}

type testWarnings struct {
	sel string
	out []string
}

var testsWarnings = [6]testWarnings{
	{sel: ":lipsum { margin: 2cm",
		out: []string{"Invalid or unsupported selector"}},
	{sel: "::lipsum { margin: 2cm",
		out: []string{"Invalid or unsupported selector"}},
	{sel: "foo { margin-color: red",
		out: []string{"Ignored", "unknown property"}},
	{sel: "foo { margin-top: red",
		out: []string{"Ignored", "invalid value"}},
	{sel: `@import "relative-uri.css"`,
		out: []string{"Relative URI reference without a base URI"}},
	{sel: `@import "invalid-protocol://absolute-URL"`,
		out: []string{"Failed to load stylesheet at"}},
}

//@assertNoLogs
// Check that appropriate warnings are logged.
func TestWarnings(t *testing.T) {
	for _, te := range testsWarnings {

		capt := utils.CaptureLogs()
		_, err := newCSS(InputString(te.sel))
		if err != nil {
			t.Fatal(err)
		}
		logs := capt.Logs()
		if len(logs) != 1 {
			t.Fatalf("%s : expected exactly 1 log, got %d", te.sel, len(logs))
		}
		for _, message := range te.out {
			if !strings.Contains(logs[0], message) {
				t.Fatalf("log should contain %s, got %s", message, logs[0])
			}
		}
	}
}

//@assertNoLogs
func TestWarningsStylesheet(t *testing.T) {
	ml := "<link rel=stylesheet href=invalid-protocol://absolute>"
	capt := utils.CaptureLogs()
	html, err := newHtml(InputString(ml))
	if err != nil {
		t.Fatal(err)
	}
	GetAllComputedStyles(html, nil, false, nil, nil, nil)
	logs := capt.Logs()
	if len(logs) != 1 {
		t.Fatalf("expected exactly 1 log, got %d", len(logs))
	}
	if !strings.Contains(logs[0], "Failed to load stylesheet at") {
		t.Fatalf("log should contain 'Failed to load stylesheet at', got %s", logs[0])
	}
}

type testFontSize struct {
	parentCss, childCss   string
	parentSize, childSize float32
}

var testsFs = []testFontSize{
	{parentCss: "10px", parentSize: 10, childCss: "10px", childSize: 10},
	{parentCss: "x-small", parentSize: 12, childCss: "xx-large", childSize: 32},
	{parentCss: "x-large", parentSize: 24, childCss: "2em", childSize: 48},
	{parentCss: "1em", parentSize: 16, childCss: "1em", childSize: 16},
	{parentCss: "1em", parentSize: 16, childCss: "larger", childSize: 6. / 5 * 16},
	{parentCss: "medium", parentSize: 16, childCss: "larger", childSize: 6. / 5 * 16},
	{parentCss: "x-large", parentSize: 24, childCss: "larger", childSize: 32},
	{parentCss: "xx-large", parentSize: 32, childCss: "larger", childSize: 1.2 * 32},
	{parentCss: "1px", parentSize: 1, childCss: "larger", childSize: 3. / 5 * 16},
	{parentCss: "28px", parentSize: 28, childCss: "larger", childSize: 32},
	{parentCss: "100px", parentSize: 100, childCss: "larger", childSize: 120},
	{parentCss: "xx-small", parentSize: 3. / 5 * 16, childCss: "larger", childSize: 12},
	{parentCss: "1em", parentSize: 16, childCss: "smaller", childSize: 8. / 9 * 16},
	{parentCss: "medium", parentSize: 16, childCss: "smaller", childSize: 8. / 9 * 16},
	{parentCss: "x-large", parentSize: 24, childCss: "smaller", childSize: 6. / 5 * 16},
	{parentCss: "xx-large", parentSize: 32, childCss: "smaller", childSize: 24},
	{parentCss: "xx-small", parentSize: 3. / 5 * 16, childCss: "smaller", childSize: 0.8 * 3. / 5 * 16},
	{parentCss: "1px", parentSize: 1, childCss: "smaller", childSize: 0.8},
	{parentCss: "28px", parentSize: 28, childCss: "smaller", childSize: 24},
	{parentCss: "100px", parentSize: 100, childCss: "smaller", childSize: 32},
}

func isClose(a, b float32) bool {
	return math.Abs(math.Round(float64(a-b))) < 1e-5
}

func TestFontSize(t *testing.T) {
	//@assertNoLogs
	capt := utils.CaptureLogs()
	html_, err := newHtml(InputString("<p>a<span>b"))
	if err != nil {
		t.Fatal(err)
	}
	document := fakeHTML{HTML: *html_}
	for _, te := range testsFs {
		css, err := newCSS(InputString(fmt.Sprintf("p{font-size:%s}span{font-size:%s}", te.parentCss, te.childCss)))
		if err != nil {
			t.Fatal(err)
		}
		styleFor := GetAllComputedStyles(document, []CSS{css}, false, nil, nil, nil)
		body := document.root.NodeChildren(true)[1]
		p := body.NodeChildren(true)[0]
		span := p.NodeChildren(true)[1]
		if got := styleFor.Get(p, "").GetFontSize(); !isClose(got.Value, te.parentSize) {
			t.Fatalf("parent: expected %v got %v", te.parentSize, got)
		}
		if got := styleFor.Get(span, "").GetFontSize(); !isClose(got.Value, te.childSize) {
			t.Fatalf("child:expected %v got %v", te.childSize, got)
		}
	}
	capt.AssertNoLogs(t)
}

//TODO: a déplacer dans la partie qui implémente render()

//@assertNoLogs
// @pytest.mark.parametrize("style", (
// "<style> html { color red; color: blue; color",
// "<html style="color; color: blue; color red">",
// ))
// func TestErrorRecovery(t *testing.T) {
// capt := utils.CaptureLogs()
// with captureLogs() as logs:
// document = FakeHTML(string=style)
// page, = document.render().pages
// html, = page.PageBox.children
// assert html.style["color"] == (0, 0, 1, 1)  // blue
// assert len(logs) == 2

//
//
////@assertNoLogs
//func TestLineHeightInheritance():
//capt := utils.CaptureLogs()do
//cument = FakeHTML(string="""
//<style>
//html { font-size: 10px; line-height: 140% }
//section { font-size: 10px; line-height: 1.4 }
//div, p { font-size: 20px; vertical-align: 50% }
//</style>
//<body><div><section><p></p></section></div></body>
//""")
//page, = document.render().pages
//html, = page.PageBox.children
//body, = html.children
//div, = body.children
//section, = div.children
//paragraph, = section.children
//assert html.style["fontSize"] == 10
//assert div.style["fontSize"] == 20
//// 140% of 10px = 14px is inherited from html
//assert strutLayout(div.style)[0] == 14
//assert div.style["verticalAlign"] == 7  // 50 % of 14px
//}
//assert paragraph.style["fontSize"] == 20
//// 1.4 is inherited from p, 1.4 * 20px on em = 28px
//assert strutLayout(paragraph.style)[0] == 28
//assert paragraph.style["verticalAlign"] == 14  // 50% of 28px
//
//
////@assertNoLogs
//func TestImportant(t *testing.T) {
//capt := utils.CaptureLogs()
//	document = FakeHTML(string="""
//	<style>
//		p:nth-child(1) { color: lime }
//	body p:nth-child(2) { color: red }
//}
//p:nth-child(3) { color: lime !important }
//body p:nth-child(3) { color: red }
//
//body p:nth-child(5) { color: lime }
//p:nth-child(5) { color: red }
//
//p:nth-child(6) { color: red }
//p:nth-child(6) { color: lime }
//</style>
//<p></p>
//<p></p>
//<p></p>
//<p></p>
//<p></p>
//<p></p>
//""")
//page, = document.render(stylesheets=[CSS(string="""
//body p:nth-child(1) { color: red }
//p:nth-child(2) { color: lime !important }
//
//p:nth-child(4) { color: lime !important }
//body p:nth-child(4) { color: red }
//""")]).pages
//html, = page.PageBox.children
//body, = html.children
//for paragraph := range body.children {
//assert paragraph.style["color"] == (0, 1, 0, 1)  // lime (light green)
//}
//
//
////@assertNoLogs
//func TestNamedPages(t *testing.T) {
//capt := utils.CaptureLogs()
//	document = FakeHTML(string="""
//	<style>
//	@page NARRow { size: landscape }
//	div { page: AUTO }
//	p { page: NARRow }
//	</style>
//	<div><p><span>a</span></p></div>
//		""")
//	page, = document.render().pages
//	html, = page.PageBox.children
//	body, = html.children
//	div, = body.children
//	p, = div.children
//	span, = p.children
//	assert html.style["page"] == ""
//	assert body.style["page"] == ""
//	assert div.style["page"] == ""
//	assert p.style["page"] == "NARRow"
//	assert span.style["page"] == "NARRow"
//
//
//	//@assertNoLogs
//	@pytest.mark.parametrize("value, width", (
//	capt := utils.CaptureLogs()
//		("96px", 96),
//		("1in", 96),
//	("72pt", 96),
//	("6pc", 96),
//	("2.54cm", 96),
//	("25.4mm", 96),
//	("101.6q", 96),
//	("1.1em", 11),
//	("1.1rem", 17.6),
//	// TODO: ch && ex units don"t work with font-face, see computedValues.py
//	// ("1.1ch", 11),
//	// ("1.5ex", 12),
//))
//	func TestUnits(value, width):
//	document = FakeHTML(baseUrl=BASEURL, string="""
//	<style>@font-face { src: url(AHEM___.TTF); font-family: ahem }</style>
//	<body style="font: 10px ahem"><p style="margin-left: %s"></p>""" % value)
//	page, = document.render().pages
//	html, = page.PageBox.children
//	body, = html.children
//	p, = body.children
//	assert p.marginLeft == width
//
//