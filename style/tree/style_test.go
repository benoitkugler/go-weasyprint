package tree

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"testing"

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
}

func (f fakeHTML) UAStylesheet() CSS {
	return testUAStylesheet
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
		fmt.Println(s.baseUrl)
		got[i] = rsplit(rsplit(s.baseUrl, "/"), ",")
	}
	exp := [2]string{"a%7Bcolor%3AcurrentColor%7D", "doc1.html"}
	if got != exp {
		t.Errorf("expected %v got %v", exp, got)
	}

	// var rules []int
	for _, sheet := range sheets {
		for _, sheetRules := range *sheet.matcher {
			fmt.Printf("%T \n", sheetRules)
			// for _, rule := range sheetRules {
			// 	rules = append(rules, rule)
			// }
		}
		// for rule := range sheet.pageRules {
		// 	rules = append(rules, rule)
		// }
	}
	// if len(rules) != 10 {
	// 	t.Errorf("expected 10 rules, got %d", len(rules))
	// }
	capt.AssertNoLogs(t)
	// TODO: test that the values are correct too
}

//
////@assertNoLogs
//func TestExpandShorthands(t *testing.T) {
//capt := utils.CaptureLogs()
//	sheet = CSS(resourceFilename("sheet2.css"))
//	assert list(sheet.matcher.lowerLocalNameSelectors) == ["li"]
//}
//rules = sheet.matcher.lowerLocalNameSelectors["li"][0][4]
//assert rules[0][0] == "marginBottom"
//assert rules[0][1] == (3, "em")
//assert rules[1][0] == "marginTop"
//assert rules[1][1] == (2, "em")
//assert rules[2][0] == "marginRight"
//assert rules[2][1] == (0, None)
//assert rules[3][0] == "marginBottom"
//assert rules[3][1] == (2, "em")
//assert rules[4][0] == "marginLeft"
//assert rules[4][1] == (0, None)
//assert rules[5][0] == "marginLeft"
//assert rules[5][1] == (4, "em")
//
//// TODO: test that the values are correct too
//
//
////@assertNoLogs
//func TestAnnotateDocument(t *testing.T) {
//capt := utils.CaptureLogs()
//	document = FakeHTML(resourceFilename("doc1.html"))
//	document.UaStylesheets = lambda: [CSS(resourceFilename("miniUa.css"))]
//	styleFor = getAllComputedStyles(
//		document, userStylesheets=[CSS(resourceFilename("user.css"))])
//}
//// Element objects behave as lists of their children
//Head, body = document.etreeElement
//h1, p, ul, div = body
//li0, Li1 = ul
//a, = li0
//span1, = div
//span2, = span1
//
//h1 = styleFor(h1)
//p = styleFor(p)
//ul = styleFor(ul)
//li0 = styleFor(li0)
//div = styleFor(div)
//after = styleFor(a, "after")
//a = styleFor(a)
//span1 = styleFor(span1)
//span2 = styleFor(span2)
//
//assert h1["backgroundImage"] == (
//("url", path2url(resourceFilename("logoSmall.png"))),)
//
//assert h1["fontWeight"] == 700
//assert h1["fontSize"] == 40  // 2em
//
//// x-large * initial = 3/2 * 16 = 24
//assert p["marginTop"] == (24, "px")
//assert p["marginRight"] == (0, "px")
//assert p["marginBottom"] == (24, "px")
//assert p["marginLeft"] == (0, "px")
//assert p["backgroundColor"] == "currentColor"
//
//// 2em * 1.25ex = 2 * 20 * 1.25 * 0.8 = 40
//// 2.5ex * 1.25ex = 2.5 * 0.8 * 20 * 1.25 * 0.8 = 40
//// TODO: ex unit doesn"t work with @font-face fonts, see computedValues.py
//// assert ul["marginTop"] == (40, "px")
//// assert ul["marginRight"] == (40, "px")
//// assert ul["marginBottom"] == (40, "px")
//// assert ul["marginLeft"] == (40, "px")
//
//assert ul["fontWeight"] == 400
//// thick = 5px, 0.25 inches = 96*.25 = 24px
//assert ul["borderTopWidth"] == 0
//assert ul["borderRightWidth"] == 5
//assert ul["borderBottomWidth"] == 0
//assert ul["borderLeftWidth"] == 24
//
//assert li0["fontWeight"] == 700
//assert li0["fontSize"] == 8  // 6pt
//assert li0["marginTop"] == (16, "px")  // 2em
//assert li0["marginRight"] == (0, "px")
//assert li0["marginBottom"] == (16, "px")
//assert li0["marginLeft"] == (32, "px")  // 4em
//
//assert a["textDecorationLine"] == {"underline"}
//assert a["fontWeight"] == 900
//assert a["fontSize"] == 24  // 300% of 8px
//assert a["paddingTop"] == (1, "px")
//assert a["paddingRight"] == (2, "px")
//assert a["paddingBottom"] == (3, "px")
//assert a["paddingLeft"] == (4, "px")
//assert a["borderTopWidth"] == 42
//assert a["borderBottomWidth"] == 42
//
//assert a["color"] == (1, 0, 0, 1)
//assert a["borderTopColor"] == "currentColor"
//
//assert div["fontSize"] == 40  // 2 * 20px
//assert span1["width"] == (160, "px")  // 10 * 16px (root default is 16px)
//assert span1["height"] == (400, "px")  // 10 * (2 * 20px)
//assert span2["fontSize"] == 32
//
//// The href attr should be as := range the source, not made absolute.
//assert after["content"] == (
//("string", " ["), ("string", "home.html"), ("string", "]"))
//assert after["backgroundColor"] == (1, 0, 0, 1)
//assert after["borderTopWidth"] == 42
//assert after["borderBottomWidth"] == 3
//
//// TODO: much more tests here: test that origin && selector precedence
//// && inheritance are correct…
//
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
//Head, body = document.etreeElement
//p, = body
//span, = p
//assert isclose(styleFor(p)["fontSize"], parentSize)
//assert isclose(styleFor(span)["fontSize"], childSize)
//}
