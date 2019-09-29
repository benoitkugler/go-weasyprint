package tree

import (
	"log"
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
	if sels[0].String() != "li" {
		t.Errorf("expected 'li' got %s", sels[0].String())
	}

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
	// capt := utils.CaptureLogs()
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
	// capt.AssertNoLogs(t)
}

//
////@assertNoLogs
//func TestPage(t *testing.T) {
//capt := utils.CaptureLogs()
//	document = FakeHTML(resourceFilename("doc1.html"))
//	styleFor = getAllComputedStyles(
//		document, userStylesheets=[CSS(string="""
//	html { color: red }
//	@page { margin: 10px }
//	@page :right {
//	color: blue;
//	margin-bottom: 12pt;
//	font-size: 20px;
//	@top-left { width: 10em }
//	@top-right { font-size: 10px}
//}
//	""")])
//
//	pageType = PageType(
//		side="left", first=true, blank=false, index=0, name="")
//	setPageTypeComputedStyles(pageType, document, styleFor)
//	style = styleFor(pageType)
//	assert style["marginTop"] == (5, "px")
//	assert style["marginLeft"] == (10, "px")
//	assert style["marginBottom"] == (10, "px")
//	assert style["color"] == (1, 0, 0, 1)  // red, inherited from html
//
//	pageType = PageType(
//		side="right", first=true, blank=false, index=0, name="")
//	setPageTypeComputedStyles(pageType, document, styleFor)
//	style = styleFor(pageType)
//	assert style["marginTop"] == (5, "px")
//	assert style["marginLeft"] == (10, "px")
//	assert style["marginBottom"] == (16, "px")
//	assert style["color"] == (0, 0, 1, 1)  // blue
//
//	pageType = PageType(
//		side="left", first=false, blank=false, index=1, name="")
//	setPageTypeComputedStyles(pageType, document, styleFor)
//	style = styleFor(pageType)
//	assert style["marginTop"] == (10, "px")
//	assert style["marginLeft"] == (10, "px")
//	assert style["marginBottom"] == (10, "px")
//	assert style["color"] == (1, 0, 0, 1)  // red, inherited from html
//
//	pageType = PageType(
//		side="right", first=false, blank=false, index=1, name="")
//	setPageTypeComputedStyles(pageType, document, styleFor)
//	style = styleFor(pageType)
//	assert style["marginTop"] == (10, "px")
//	assert style["marginLeft"] == (10, "px")
//	assert style["marginBottom"] == (16, "px")
//	assert style["color"] == (0, 0, 1, 1)  // blue
//
//	pageType = PageType(
//		side="left", first=true, blank=false, index=0, name="")
//	setPageTypeComputedStyles(pageType, document, styleFor)
//	style = styleFor(pageType, "@top-left")
//	assert style is None
//
//	pageType = PageType(
//		side="right", first=true, blank=false, index=0, name="")
//	setPageTypeComputedStyles(pageType, document, styleFor)
//	style = styleFor(pageType, "@top-left")
//	assert style["fontSize"] == 20  // inherited from @page
//	assert style["width"] == (200, "px")
//
//	pageType = PageType(
//		side="right", first=true, blank=false, index=0, name="")
//	setPageTypeComputedStyles(pageType, document, styleFor)
//	style = styleFor(pageType, "@top-right")
//	assert style["fontSize"] == 10
//
//
//	//@assertNoLogs
//	@pytest.mark.parametrize("style, selectors", (
//	capt := utils.CaptureLogs()
//		("@page {}", [{
//		"side": None, "blank": None, "first": None, "name": None,
//		"index": None, "specificity": [0, 0, 0]}]),
//("@page :left {}", [{
//"side": "left", "blank": None, "first": None, "name": None,
//"index": None, "specificity": [0, 0, 1]}]),
//("@page:first:left {}", [{
//"side": "left", "blank": None, "first": true, "name": None,
//"index": None, "specificity": [0, 1, 1]}]),
//("@page pagename {}", [{
//"side": None, "blank": None, "first": None, "name": "pagename",
//"index": None, "specificity": [1, 0, 0]}]),
//("@page pagename:first:right:blank {}", [{
//"side": "right", "blank": true, "first": true, "name": "pagename",
//"index": None, "specificity": [1, 2, 1]}]),
//("@page pagename, :first {}", [
//{"side": None, "blank": None, "first": None, "name": "pagename",
//"index": None, "specificity": [1, 0, 0]},
//{"side": None, "blank": None, "first": true, "name": None,
//"index": None, "specificity": [0, 1, 0]}]),
//("@page :first:first {}", [{
//"side": None, "blank": None, "first": true, "name": None,
//"index": None, "specificity": [0, 2, 0]}]),
//("@page :left:left {}", [{
//"side": "left", "blank": None, "first": None, "name": None,
//"index": None, "specificity": [0, 0, 2]}]),
//("@page :nth(2) {}", [{
//"side": None, "blank": None, "first": None, "name": None,
//"index": (0, 2, None), "specificity": [0, 1, 0]}]),
//("@page :nth(2n + 4) {}", [{
//"side": None, "blank": None, "first": None, "name": None,
//"index": (2, 4, None), "specificity": [0, 1, 0]}]),
//("@page :nth(3n) {}", [{
//"side": None, "blank": None, "first": None, "name": None,
//"index": (3, 0, None), "specificity": [0, 1, 0]}]),
//("@page :nth( n+2 ) {}", [{
//"side": None, "blank": None, "first": None, "name": None,
//"index": (1, 2, None), "specificity": [0, 1, 0]}]),
//("@page :nth(even) {}", [{
//"side": None, "blank": None, "first": None, "name": None,
//"index": (2, 0, None), "specificity": [0, 1, 0]}]),
//("@page pagename:nth(2) {}", [{
//"side": None, "blank": None, "first": None, "name": "pagename",
//"index": (0, 2, None), "specificity": [1, 1, 0]}]),
//("@page page page {}", None),
//("@page :left page {}", None),
//("@page :left, {}", None),
//("@page , {}", None),
//("@page :left, test, {}", None),
//("@page :wrong {}", None),
//("@page :left:wrong {}", None),
//("@page :left:right {}", None),
//))
//func TestPageSelectors(style, selectors):
//atRule, = tinycss2.parseStylesheet(style)
//assert parsePageSelectors(atRule) == selectors
//
//
////@assertNoLogs
//@pytest.mark.parametrize("source, messages", (
//("" +capt := utils.CaptureLogs()
//	":lipsum { margin: 2cm", ["WARNING: Invalid || unsupported selector"]),
//("::lipsum { margin: 2cm", ["WARNING: Invalid || unsupported selector"]),
//("foo { margin-color: red", ["WARNING: Ignored", "unknown property"]),
//("foo { margin-top: red", ["WARNING: Ignored", "invalid value"]),
//("@import "relative-uri.css"",
//["ERROR: Relative URI reference without a base URI"]),
//("@import "invalid-protocol://absolute-URL"",
//["ERROR: Failed to load stylesheet at"]),
//))
//// Check that appropriate warnings are logged.
//func TestWarnings(source, messages):
//with captureLogs() as logs:
//CSS(string=source)
//assert len(logs) == 1, source
//for message := range messages:
//assert message := range logs[0]
//
//
////@assertNoLogs
//func TestWarningsStylesheet():
//capt := utils.CaptureLogs()ht
//ml = "<link rel=stylesheet href=invalid-protocol://absolute>"
//with captureLogs() as logs:
//FakeHTML(string=html).render()
//assert len(logs) == 1
//assert "ERROR: Failed to load stylesheet at" := range logs[0]
//
//
////@assertNoLogs
//@pytest.mark.parametrize("style", (
//"<" +capt := utils.CaptureLogs()
//	"style> html { color red; color: blue; color",
//"<html style="color; color: blue; color red">",
//))
//func TestErrorRecovery(style):
//with captureLogs() as logs:
//document = FakeHTML(string=style)
//page, = document.render().pages
//html, = page.PageBox.children
//assert html.style["color"] == (0, 0, 1, 1)  // blue
//assert len(logs) == 2
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
//	//@assertNoLogs
//	@pytest.mark.parametrize("parentCss, parentSize, childCss, childSize", (
//	capt := utils.CaptureLogs()
//		("10px", 10, "10px", 10),
//		("x-small", 12, "xx-large", 32),
//	("x-large", 24, "2em", 48),
//	("1em", 16, "1em", 16),
//	("1em", 16, "larger", 6 / 5 * 16),
//	("medium", 16, "larger", 6 / 5 * 16),
//	("x-large", 24, "larger", 32),
//	("xx-large", 32, "larger", 1.2 * 32),
//	("1px", 1, "larger", 3 / 5 * 16),
//	("28px", 28, "larger", 32),
//	("100px", 100, "larger", 120),
//	("xx-small", 3 / 5 * 16, "larger", 12),
//	("1em", 16, "smaller", 8 / 9 * 16),
//	("medium", 16, "smaller", 8 / 9 * 16),
//	("x-large", 24, "smaller", 6 / 5 * 16),
//	("xx-large", 32, "smaller", 24),
//	("xx-small", 3 / 5 * 16, "smaller", 0.8 * 3 / 5 * 16),
//	("1px", 1, "smaller", 0.8),
//	("28px", 28, "smaller", 24),
//	("100px", 100, "smaller", 32),
//))
//	func TestFontSize(parentCss, parentSize, childCss, childSize):
//	document = FakeHTML(string="<p>a<span>b")
//	styleFor = getAllComputedStyles(document, userStylesheets=[CSS(
//		string="p{font-size:%s}span{font-size:%s}" % (parentCss, childCss))])
//
//head, body = document.etreeElement
//p, = body
//span, = p
//assert isclose(styleFor(p)["fontSize"], parentSize)
//assert isclose(styleFor(span)["fontSize"], childSize)
//}
