package boxes

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"

	"github.com/benoitkugler/go-weasyprint/boxes/counters"
	"github.com/benoitkugler/go-weasyprint/images"
	"github.com/benoitkugler/go-weasyprint/layout/text"
	"github.com/benoitkugler/go-weasyprint/style/parser"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/style/tree"
	"github.com/benoitkugler/go-weasyprint/utils"
	"github.com/benoitkugler/go-weasyprint/utils/testutils"
)

var (
	_ ReplacedBoxITF = (*ReplacedBox)(nil)
	_ ReplacedBoxITF = (*BlockReplacedBox)(nil)
	_ ReplacedBoxITF = (*InlineReplacedBox)(nil)

	_ BlockLevelBoxITF = (*BlockBox)(nil)
	_ BlockLevelBoxITF = (*BlockReplacedBox)(nil)
	_ BlockLevelBoxITF = (*TableBox)(nil)
	_ BlockLevelBoxITF = (*FlexBox)(nil)

	_ TableBoxITF = (*TableBox)(nil)
	_ TableBoxITF = (*InlineTableBox)(nil)
)

//  Test that the "before layout" box tree is correctly constructed.

func fakeHTML(html *tree.HTML) *tree.HTML {
	html.UAStyleSheet = tree.TestUAStylesheet
	return html
}

func parseBase(t *testing.T, content utils.ContentInput, baseUrl string) (*utils.HTMLNode, *tree.StyleFor, Gifu, string, *tree.TargetCollector, counters.CounterStyle, text.TextLayoutContext) {
	html, err := tree.NewHTML(content, baseUrl, utils.DefaultUrlFetcher, "")
	if err != nil {
		t.Fatalf("parsing HTML failed: %s", err)
	}
	document := fakeHTML(html)
	cs := make(counters.CounterStyle)
	style := tree.GetAllComputedStyles(document, nil, false, nil, cs, nil, nil, nil)
	imgFetcher := func(url string, forcedMimeType string) images.Image {
		out, err := images.GetImageFromUri(make(map[string]images.Image), document.UrlFetcher, false, url, forcedMimeType)
		if err != nil {
			log.Println(err)
		}
		return out
	}
	tr := tree.NewTargetCollector()
	return document.Root, style, imgFetcher, baseUrl, &tr, cs, nil
}

func parse(t *testing.T, htmlContent string) BoxITF {
	a, b, c, d, e, f, _ := parseBase(t, utils.InputString(htmlContent), baseUrl)
	boxes := elementToBox(a, b, c, d, e, f, nil)
	return boxes[0]
}

func parseAndBuild(t *testing.T, htmlContent string) BlockLevelBoxITF {
	return parseAndBuildExt(t, htmlContent, baseUrl)
}

func parseAndBuildExt(t *testing.T, htmlContent, baseUrl string) BlockLevelBoxITF {
	box := BuildFormattingStructure(parseBase(t, utils.InputString(htmlContent), baseUrl))
	if err := sanityChecks(box); err != nil {
		t.Fatalf("sanity check failed: %s", err)
	}
	return box
}

type bc struct {
	text string
	c    []serBox
}

type serBox struct {
	tag     string
	type_   BoxType
	content bc
}

func (s serBox) equals(other serBox) bool {
	if s.tag != other.tag || s.type_ != other.type_ || s.content.text != other.content.text {
		return false
	}
	return serializedBoxEquals(s.content.c, other.content.c)
}

func serializedBoxEquals(l1, l2 []serBox) bool {
	if len(l1) != len(l2) {
		return false
	}
	for j := range l1 {
		if !l1[j].equals(l2[j]) {
			return false
		}
	}
	return true
}

// Transform a box list into a structure easier to compare for testing.
func serialize(boxList []Box) []serBox {
	out := make([]serBox, len(boxList))
	for i, box := range boxList {
		out[i].tag = box.Box().ElementTag
		out[i].type_ = box.Type()
		// all concrete boxes are either text, replaced, column or parent.
		if boxT, ok := box.(*TextBox); ok {
			out[i].content.text = boxT.Text
		} else if _, ok := box.(ReplacedBoxITF); ok {
			out[i].content.text = "<replaced>"
		} else {
			var cg []Box
			if table, ok := box.(TableBoxITF); ok {
				cg = table.Table().ColumnGroups
			}
			cg = append(cg, box.Box().Children...)
			out[i].content.c = serialize(cg)
		}
	}
	return out
}

// Check the box tree equality.
//
// The obtained result is prettified in the message in case of failure.
//
// box: a Box object, starting with <html> and <body> blocks.
// expected: a list of serialized <body> children as returned by to_lists().
func assertTree(t *testing.T, box Box, expected []serBox) {
	if tag := box.Box().ElementTag; tag != "html" {
		t.Fatalf("unexpected element: %s", tag)
	}
	if !BlockBoxT.IsInstance(box) {
		t.Fatal("expected block box")
	}
	if L := len(box.Box().Children); L != 1 {
		t.Fatalf("expected one children, got %d", L)
	}

	box = box.Box().Children[0]
	if !BlockBoxT.IsInstance(box) {
		t.Fatal("expected block box")
	}
	if tag := box.Box().ElementTag; tag != "body" {
		t.Fatalf("unexpected element: %s", tag)
	}

	if got := serialize(box.Box().Children); !serializedBoxEquals(got, expected) {
		t.Fatalf("expected \n%v\n, got\n%v", expected, got)
	}
}

var properChildren = map[BoxType][]BoxType{
	BlockContainerBoxT: {BlockLevelBoxT, LineBoxT},
	LineBoxT:           {InlineLevelBoxT},
	InlineBoxT:         {InlineLevelBoxT},
	TableBoxT: {
		TableCaptionBoxT,
		TableColumnGroupBoxT, TableColumnBoxT,
		TableRowGroupBoxT, TableRowBoxT,
	},
	InlineTableBoxT: {
		TableCaptionBoxT,
		TableColumnGroupBoxT, TableColumnBoxT,
		TableRowGroupBoxT, TableRowBoxT,
	},
	TableColumnGroupBoxT: {TableColumnBoxT},
	TableRowGroupBoxT:    {TableRowBoxT},
	TableRowBoxT:         {TableCellBoxT},
}

// Check that the rules regarding boxes are met.
//
// This is not required and only helps debugging.
//
// - A block container can contain either only block-level boxes or
//   only line boxes;
// - Line boxes and inline boxes can only contain inline-level boxes.
func sanityChecks(box Box) error {
	if !ParentBoxT.IsInstance(box) {
		return nil
	}

	acceptablesListsT, ok := properChildren[box.Type()]
	if !ok {
		return nil // this is less strict than the reference implementation
	}

	for _, child := range box.Box().Children {
		if !child.Box().IsInNormalFlow() {
			continue
		}
		isOk := false
		for _, typeOk := range acceptablesListsT {
			if typeOk.IsInstance(child) {
				isOk = true
				break
			}
		}
		if !isOk {
			return errors.New("invalid children check")
		}
	}

	for _, child := range box.Box().Children {
		if err := sanityChecks(child); err != nil {
			return err
		}
	}

	return nil
}

// func _parse_base(htmlContent, base_url=BASE_URL):
//     document = FakeHTML(string=htmlContent, base_url=base_url)
//     counter_style = CounterStyle()
//     style_for = get_all_computed_styles(document, counter_style=counter_style)
//     get_image_from_uri = functools.partial(
//         images.get_image_from_uri, cache={}, url_fetcher=document.url_fetcher,
//         optimize_size=())
//     target_collector = TargetCollector()
//     return (
//         document.etree_element, style_for, get_image_from_uri, base_url,
//         target_collector, counter_style)

var baseUrl, _ = utils.Path2url("../resources_test/")

func getGrid(t *testing.T, html string) ([][]Border, [][]Border) {
	root := parseAndBuild(t, html)
	body := root.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0].(TableBoxITF)

	buildGrid := func(bg [][]Border) (grid [][]Border /*maybe nil*/) {
		for _, column := range bg {
			out := make([]Border, len(column))
			for i, border := range column {
				if border.Width != 0 {
					border.Score = Score{}
					out[i] = border
				}
			}
			grid = append(grid, out)
		}
		return grid
	}

	return buildGrid(table.Table().CollapsedBorderGrid.Vertical),
		buildGrid(table.Table().CollapsedBorderGrid.Horizontal)
}

func TestBoxTree(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parse(t, "<p>"), []serBox{{"p", BlockBoxT, bc{}}})
	assertTree(t, parse(t, `
	  <style>
	    span { display: inline-block }
	  </style>
	  <p>Hello <em>World <img src="pattern.png"><span>L</span></em>!</p>`),
		[]serBox{
			{
				"p", BlockBoxT, bc{c: []serBox{
					{"p", TextBoxT, bc{text: "Hello "}},
					{"em", InlineBoxT, bc{c: []serBox{
						{"em", TextBoxT, bc{text: "World "}},
						{"img", InlineReplacedBoxT, bc{text: "<replaced>"}},
						{"span", InlineBlockBoxT, bc{c: []serBox{
							{"span", TextBoxT, bc{text: "L"}},
						}}},
					}}},
					{"p", TextBoxT, bc{text: "!"}},
				}},
			},
		})
}

func TestHtmlEntities(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	for _, quote := range []string{`"`, "&quot;", "&#x22;", "&#34;"} {
		assertTree(t, parse(t, fmt.Sprintf("<p>%sabc%s", quote, quote)), []serBox{
			{"p", BlockBoxT, bc{c: []serBox{
				{"p", TextBoxT, bc{text: `"abc"`}},
			}}},
		})
	}
}

func TestInlineInBlock1(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	source := "<div>Hello, <em>World</em>!\n<p>Lipsum.</p></div>"
	expected := []serBox{
		{"div", BlockBoxT, bc{
			c: []serBox{
				{
					"div", BlockBoxT,
					bc{c: []serBox{
						{"div", LineBoxT, bc{c: []serBox{
							{"div", TextBoxT, bc{text: "Hello, "}},
							{"em", InlineBoxT, bc{c: []serBox{
								{"em", TextBoxT, bc{text: "World"}},
							}}},
							{"div", TextBoxT, bc{text: "!\n"}},
						}}},
					}},
				},
				{"p", BlockBoxT, bc{c: []serBox{
					{"p", LineBoxT, bc{c: []serBox{
						{"p", TextBoxT, bc{text: "Lipsum."}},
					}}},
				}}},
			},
		}},
	}
	box := parse(t, source)

	assertTree(t, box, []serBox{
		{"div", BlockBoxT, bc{c: []serBox{
			{"div", TextBoxT, bc{text: "Hello, "}},
			{"em", InlineBoxT, bc{c: []serBox{
				{"em", TextBoxT, bc{text: "World"}},
			}}},
			{"div", TextBoxT, bc{text: "!\n"}},
			{"p", BlockBoxT, bc{c: []serBox{{"p", TextBoxT, bc{text: "Lipsum."}}}}},
		}}},
	})

	box = InlineInBlock(box)
	assertTree(t, box, expected)
}

func TestInlineInBlock2(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	source := "<div><p>Lipsum.</p>Hello, <em>World</em>!\n</div>"
	expected := []serBox{
		{"div", BlockBoxT, bc{c: []serBox{
			{"p", BlockBoxT, bc{c: []serBox{{"p", LineBoxT, bc{c: []serBox{{"p", TextBoxT, bc{text: "Lipsum."}}}}}}}},
			{"div", BlockBoxT, bc{c: []serBox{
				{"div", LineBoxT, bc{c: []serBox{
					{"div", TextBoxT, bc{text: "Hello, "}},
					{"em", InlineBoxT, bc{c: []serBox{{"em", TextBoxT, bc{text: "World"}}}}},
					{"div", TextBoxT, bc{text: "!\n"}},
				}}},
			}}},
		}}},
	}
	box := parse(t, source)
	box = InlineInBlock(box)
	assertTree(t, box, expected)
}

func TestInlineInBlock3(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Absolutes are left := range the lines to get their static position later.
	source := `<p>Hello <em style="position:absolute;
                                    display: block">World</em>!</p>`
	expected := []serBox{
		{"p", BlockBoxT, bc{c: []serBox{
			{"p", LineBoxT, bc{c: []serBox{
				{"p", TextBoxT, bc{text: "Hello "}},
				{"em", BlockBoxT, bc{c: []serBox{{"em", LineBoxT, bc{c: []serBox{{"em", TextBoxT, bc{text: "World"}}}}}}}},
				{"p", TextBoxT, bc{text: "!"}},
			}}},
		}}},
	}
	box := parse(t, source)
	box = InlineInBlock(box)
	assertTree(t, box, expected)
	box = BlockInInline(box)
	assertTree(t, box, expected)
}

func TestInlineInBlock4(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Floats are pull to the top of their containing blocks
	source := `<p>Hello <em style="float: left">World</em>!</p>`

	expected := []serBox{
		{"p", BlockBoxT, bc{c: []serBox{
			{"p", LineBoxT, bc{c: []serBox{
				{"p", TextBoxT, bc{text: "Hello "}},
				{"em", BlockBoxT, bc{c: []serBox{{"em", LineBoxT, bc{c: []serBox{{"em", TextBoxT, bc{text: "World"}}}}}}}},
				{"p", TextBoxT, bc{text: "!"}},
			}}},
		}}},
	}
	box := parse(t, source)
	box = InlineInBlock(box)
	box = BlockInInline(box)
	assertTree(t, box, expected)
}

func TestBlockInInline(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	box := parse(t, `
      <style>
        p { display: inline-block; }
        span, i { display: block; }
      </style>
      <p>Lorem <em>ipsum <strong>dolor <span>sit</span>
      <span>amet,</span></strong><span><em>conse<i>`)
	box = InlineInBlock(box)
	assertTree(t, box, []serBox{
		{"body", LineBoxT, bc{c: []serBox{
			{"p", InlineBlockBoxT, bc{c: []serBox{
				{"p", LineBoxT, bc{c: []serBox{
					{"p", TextBoxT, bc{text: "Lorem "}},
					{"em", InlineBoxT, bc{c: []serBox{
						{"em", TextBoxT, bc{text: "ipsum "}},
						{"strong", InlineBoxT, bc{c: []serBox{
							{"strong", TextBoxT, bc{text: "dolor "}},
							{"span", BlockBoxT, bc{c: []serBox{{"span", LineBoxT, bc{c: []serBox{{"span", TextBoxT, bc{text: "sit"}}}}}}}},
							{"strong", TextBoxT, bc{text: "\n      "}},
							{"span", BlockBoxT, bc{c: []serBox{{"span", LineBoxT, bc{c: []serBox{{"span", TextBoxT, bc{text: "amet,"}}}}}}}},
						}}},
						{"span", BlockBoxT, bc{c: []serBox{
							{"span", LineBoxT, bc{c: []serBox{
								{"em", InlineBoxT, bc{c: []serBox{
									{"em", TextBoxT, bc{text: "conse"}},
									{"i", BlockBoxT, bc{c: []serBox{}}},
								}}},
							}}},
						}}},
					}}},
				}}},
			}}},
		}}},
	})

	box = BlockInInline(box)
	assertTree(t, box, []serBox{
		{"body", LineBoxT, bc{c: []serBox{
			{"p", InlineBlockBoxT, bc{c: []serBox{
				{"p", BlockBoxT, bc{c: []serBox{
					{"p", LineBoxT, bc{c: []serBox{
						{"p", TextBoxT, bc{text: "Lorem "}},
						{"em", InlineBoxT, bc{c: []serBox{
							{"em", TextBoxT, bc{text: "ipsum "}},
							{"strong", InlineBoxT, bc{c: []serBox{{"strong", TextBoxT, bc{text: "dolor "}}}}},
						}}},
					}}},
				}}},
				{"span", BlockBoxT, bc{c: []serBox{{"span", LineBoxT, bc{c: []serBox{{"span", TextBoxT, bc{text: "sit"}}}}}}}},
				{"p", BlockBoxT, bc{c: []serBox{
					{"p", LineBoxT, bc{c: []serBox{
						{"em", InlineBoxT, bc{c: []serBox{{"strong", InlineBoxT, bc{c: []serBox{{"strong", TextBoxT, bc{text: "\n      "}}}}}}}},
					}}},
				}}},
				{"span", BlockBoxT, bc{c: []serBox{{"span", LineBoxT, bc{c: []serBox{{"span", TextBoxT, bc{text: "amet,"}}}}}}}},
				{"p", BlockBoxT, bc{c: []serBox{
					{"p", LineBoxT, bc{c: []serBox{{"em", InlineBoxT, bc{c: []serBox{{"strong", InlineBoxT, bc{c: []serBox{}}}}}}}}},
				}}},
				{"span", BlockBoxT, bc{c: []serBox{
					{"span", BlockBoxT, bc{c: []serBox{
						{"span", LineBoxT, bc{c: []serBox{{"em", InlineBoxT, bc{c: []serBox{{"em", TextBoxT, bc{text: "conse"}}}}}}}},
					}}},
					{"i", BlockBoxT, bc{c: []serBox{}}},
					{"span", BlockBoxT, bc{c: []serBox{{"span", LineBoxT, bc{c: []serBox{{"em", InlineBoxT, bc{c: []serBox{}}}}}}}}},
				}}},
				{"p", BlockBoxT, bc{c: []serBox{{"p", LineBoxT, bc{c: []serBox{{"em", InlineBoxT, bc{c: []serBox{}}}}}}}}},
			}}},
		}}},
	})
}

func TestStyles(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	box := parse(t, `
		  <style>
			span { display: block; }
			* { margin: 42px }
			html { color: blue }
		  </style>
		  <p>Lorem <em>ipsum <strong>dolor <span>sit</span>
			<span>amet,</span></strong><span>consectetur</span></em></p>`)
	box = InlineInBlock(box)
	box = BlockInInline(box)

	descendants := Descendants(box)
	if L := len(descendants); L != 31 {
		t.Fatalf("expected 31 descendants, got %d", L)
	}
	if d := descendants[0]; d != box {
		t.Fatalf("expected box to be the first descendant, got %v", d)
	}

	for _, child := range descendants {
		// All boxes inherit the color
		if c := child.Box().Style.GetColor(); c.RGBA != (parser.RGBA{R: 0, G: 0, B: 1, A: 1}) { // blue
			t.Fatal()
		}
		// Only non-anonymous boxes have margins
		if mt := child.Box().Style.GetMarginTop(); mt != pr.FToPx(0) && mt != pr.FToPx(42) {
			t.Fatal()
		}
	}
}

func TestWhitespaces(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// http://www.w3.org/TR/CSS21/text.html#white-space-model
	assertTree(t, parseAndBuild(t, "<p>Lorem \t\r\n  ipsum\t"+`<strong>  dolor
		<img src=pattern.png> sit
        <span style="position: absolute"></span> <em> amet </em>
        consectetur</strong>.</p>`+
		"<pre>\t  foo\n</pre>"+
		"<pre style=\"white-space: pre-wrap\">\t  foo\n</pre>"+
		"<pre style=\"white-space: pre-line\">\t  foo\n</pre>"), []serBox{
		{"p", BlockBoxT, bc{c: []serBox{
			{"p", LineBoxT, bc{c: []serBox{
				{"p", TextBoxT, bc{text: "Lorem ipsum "}},
				{"strong", InlineBoxT, bc{c: []serBox{
					{"strong", TextBoxT, bc{text: "dolor "}},
					{"img", InlineReplacedBoxT, bc{text: "<replaced>"}},
					{"strong", TextBoxT, bc{text: " sit "}},
					{"span", BlockBoxT, bc{c: []serBox{}}},
					{"em", InlineBoxT, bc{c: []serBox{{"em", TextBoxT, bc{text: "amet "}}}}},
					{"strong", TextBoxT, bc{text: "consectetur"}},
				}}},
				{"p", TextBoxT, bc{text: "."}},
			}}},
		}}},
		{"pre", BlockBoxT, bc{c: []serBox{{"pre", LineBoxT, bc{c: []serBox{{"pre", TextBoxT, bc{text: "\t  foo\n"}}}}}}}},
		{"pre", BlockBoxT, bc{c: []serBox{{"pre", LineBoxT, bc{c: []serBox{{"pre", TextBoxT, bc{text: "\t  foo\n"}}}}}}}},
		{"pre", BlockBoxT, bc{c: []serBox{{"pre", LineBoxT, bc{c: []serBox{{"pre", TextBoxT, bc{text: " foo\n"}}}}}}}},
	})
}

type pageStyleData struct {
	type_                    utils.PageElement
	top, right, bottom, left pr.Float
}

func testPageStyle(t *testing.T, data pageStyleData) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	document, err := tree.NewHTML(utils.InputString(`
      <style>
        @page { margin: 3px }
        @page name { margin-left: 15px; margin-top: 5px }
        @page :nth(3) { margin-bottom: 1px }
        @page :nth(5n+4) { margin-bottom: 2px }
        @page :first { margin-top: 20px }
        @page :right { margin-right: 10px; margin-top: 10px }
        @page :left { margin-left: 10px; margin-top: 10px }
      </style>
    `), "", utils.DefaultUrlFetcher, "")
	if err != nil {
		t.Fatal(err)
	}
	document = fakeHTML(document)
	styleFor := tree.GetAllComputedStyles(document, nil, false, nil, nil, nil, nil, nil)

	// Force the generation of the style for this page type as it"s generally
	// only done during the rendering.
	styleFor.SetPageComputedStylesT(data.type_, document)

	style := styleFor.Get(data.type_, "")
	if m := style.GetMarginTop(); m != pr.FToPx(data.top) {
		t.Fatalf("expected %f, got %v", data.top, m)
	}
	if m := style.GetMarginRight(); m != pr.FToPx(data.right) {
		t.Fatalf("expected %f, got %v", data.right, m)
	}
	if m := style.GetMarginBottom(); m != pr.FToPx(data.bottom) {
		t.Fatalf("expected %f, got %v", data.bottom, m)
	}
	if m := style.GetMarginLeft(); m != pr.FToPx(data.left) {
		t.Fatalf("expected %f, got %v", data.left, m)
	}
}

func TestPageStyle(t *testing.T) {
	for _, data := range []pageStyleData{
		{utils.PageElement{Side: "left", First: true, Index: 0, Blank: false, Name: ""}, 20, 3, 3, 10},
		{utils.PageElement{Side: "right", First: true, Index: 0, Blank: false, Name: ""}, 20, 10, 3, 3},
		{utils.PageElement{Side: "left", First: false, Index: 1, Blank: false, Name: ""}, 10, 3, 3, 10},
		{utils.PageElement{Side: "right", First: false, Index: 1, Blank: false, Name: ""}, 10, 10, 3, 3},
		{utils.PageElement{Side: "right", First: false, Index: 1, Blank: false, Name: "name"}, 5, 10, 3, 15},
		{utils.PageElement{Side: "right", First: false, Index: 2, Blank: false, Name: "name"}, 5, 10, 1, 15},
		{utils.PageElement{Side: "right", First: false, Index: 8, Blank: false, Name: "name"}, 5, 10, 2, 15},
	} {
		testPageStyle(t, data)
	}
}

func TestImages1(t *testing.T) {
	cp := testutils.CaptureLogs()

	result := parseAndBuild(t, `
          <p><img src=pattern.png
            /><img alt="No src"
            /><img src=inexistent.jpg alt="Inexistent src" /></p>`)
	logs := cp.Logs()
	if L := len(logs); L != 1 {
		t.Fatalf("expected one log, got %d", L)
	}
	if !strings.Contains(logs[0], "Failed to load image") {
		t.Fatal(logs[0])
	}
	if !strings.Contains(logs[0], "inexistent.jpg") {
		t.Fatal(logs[0])
	}
	assertTree(t, result, []serBox{
		{"p", BlockBoxT, bc{c: []serBox{
			{"p", LineBoxT, bc{c: []serBox{
				{"img", InlineReplacedBoxT, bc{text: "<replaced>"}},
				{"img", InlineBoxT, bc{c: []serBox{{"img", TextBoxT, bc{text: "No src"}}}}},
				{"img", InlineBoxT, bc{c: []serBox{{"img", TextBoxT, bc{text: "Inexistent src"}}}}},
			}}},
		}}},
	})
}

func TestImages2(t *testing.T) {
	cp := testutils.CaptureLogs()

	result := parseAndBuildExt(t, `<p><img src=pattern.png alt="No baseUrl">`, "")
	logs := cp.Logs()
	if L := len(logs); L != 1 {
		t.Fatalf("expected one log, got %d", L)
	}
	if !strings.Contains(logs[0], "Relative URI reference without a base URI") {
		t.Fatal(logs[0])
	}
	assertTree(t, result, []serBox{
		{"p", BlockBoxT, bc{c: []serBox{
			{"p", LineBoxT, bc{c: []serBox{{"img", InlineBoxT, bc{c: []serBox{{"img", TextBoxT, bc{text: "No baseUrl"}}}}}}}},
		}}},
	})
}

func TestTables1(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Rules in http://www.w3.org/TR/CSS21/tables.html#anonymous-boxes

	// Rule 1.3
	// Also table model: http://www.w3.org/TR/CSS21/tables.html#model
	assertTree(t, parseAndBuild(t, `
      <x-table>
        <x-tr>
          <x-th>foo</x-th>
          <x-th>bar</x-th>
        </x-tr>
        <x-tfoot></x-tfoot>
        <x-thead><x-th></x-th></x-thead>
        <x-caption style="caption-side: bottom"></x-caption>
        <x-thead></x-thead>
        <x-col></x-col>
        <x-caption>top caption</x-caption>
        <x-tr>
          <x-td>baz</x-td>
        </x-tr>
      </x-table>
    `), []serBox{
		{"x-table", BlockBoxT, bc{c: []serBox{
			{"x-caption", TableCaptionBoxT, bc{c: []serBox{{"x-caption", LineBoxT, bc{c: []serBox{{"x-caption", TextBoxT, bc{text: "top caption"}}}}}}}},
			{"x-table", TableBoxT, bc{c: []serBox{
				{"x-table", TableColumnGroupBoxT, bc{c: []serBox{{"x-col", TableColumnBoxT, bc{c: []serBox{}}}}}},
				{"x-thead", TableRowGroupBoxT, bc{c: []serBox{{"x-thead", TableRowBoxT, bc{c: []serBox{{"x-th", TableCellBoxT, bc{c: []serBox{}}}}}}}}},
				{"x-table", TableRowGroupBoxT, bc{c: []serBox{
					{"x-tr", TableRowBoxT, bc{c: []serBox{
						{"x-th", TableCellBoxT, bc{c: []serBox{{"x-th", LineBoxT, bc{c: []serBox{{"x-th", TextBoxT, bc{text: "foo"}}}}}}}},
						{"x-th", TableCellBoxT, bc{c: []serBox{{"x-th", LineBoxT, bc{c: []serBox{{"x-th", TextBoxT, bc{text: "bar"}}}}}}}},
					}}},
				}}},
				{"x-thead", TableRowGroupBoxT, bc{c: []serBox{}}},
				{"x-table", TableRowGroupBoxT, bc{c: []serBox{
					{"x-tr", TableRowBoxT, bc{c: []serBox{
						{"x-td", TableCellBoxT, bc{c: []serBox{{"x-td", LineBoxT, bc{c: []serBox{{"x-td", TextBoxT, bc{text: "baz"}}}}}}}},
					}}},
				}}},
				{"x-tfoot", TableRowGroupBoxT, bc{c: []serBox{}}},
			}}},
			{"x-caption", TableCaptionBoxT, bc{c: []serBox{}}},
		}}},
	})
}

func TestTables2(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Rules 1.4 && 3.1
	assertTree(t, parseAndBuild(t, `
      <span style="display: table-cell">foo</span>
      <span style="display: table-cell">bar</span>
   `), []serBox{
		{"body", BlockBoxT, bc{c: []serBox{
			{"body", TableBoxT, bc{c: []serBox{
				{"body", TableRowGroupBoxT, bc{c: []serBox{
					{"body", TableRowBoxT, bc{c: []serBox{
						{"span", TableCellBoxT, bc{c: []serBox{{"span", LineBoxT, bc{c: []serBox{{"span", TextBoxT, bc{text: "foo"}}}}}}}},
						{"span", TableCellBoxT, bc{c: []serBox{{"span", LineBoxT, bc{c: []serBox{{"span", TextBoxT, bc{text: "bar"}}}}}}}},
					}}},
				}}},
			}}},
		}}},
	})
}

func TestTables3(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// http://www.w3.org/TR/CSS21/tables.html#anonymous-boxes
	// Rules 1.1 && 1.2
	// Rule XXX (not := range the spec): column groups have at least one column child
	assertTree(t, parseAndBuild(t, `
      <span style="display: table-column-group">
        1
        <em style="display: table-column">
          2
          <strong>3</strong>
        </em>
        <strong>4</strong>
      </span>
      <ins style="display: table-column-group"></ins>
    `), []serBox{
		{"body", BlockBoxT, bc{c: []serBox{
			{"body", TableBoxT, bc{c: []serBox{
				{"span", TableColumnGroupBoxT, bc{c: []serBox{{"em", TableColumnBoxT, bc{c: []serBox{}}}}}},
				{"ins", TableColumnGroupBoxT, bc{c: []serBox{{"ins", TableColumnBoxT, bc{c: []serBox{}}}}}},
			}}},
		}}},
	})
}

func TestTables4(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Rules 2.1 then 2.3
	assertTree(t, parseAndBuild(t, "<x-table>foo <div></div></x-table>"), []serBox{
		{"x-table", BlockBoxT, bc{c: []serBox{
			{"x-table", TableBoxT, bc{c: []serBox{
				{"x-table", TableRowGroupBoxT, bc{c: []serBox{
					{"x-table", TableRowBoxT, bc{c: []serBox{
						{"x-table", TableCellBoxT, bc{c: []serBox{
							{"x-table", BlockBoxT, bc{c: []serBox{{"x-table", LineBoxT, bc{c: []serBox{{"x-table", TextBoxT, bc{text: "foo "}}}}}}}},
							{"div", BlockBoxT, bc{c: []serBox{}}},
						}}},
					}}},
				}}},
			}}},
		}}},
	})
}

func TestTables5(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Rule 2.2
	assertTree(t, parseAndBuild(t, `<x-thead style="display: table-header-group"><div></div><x-td></x-td></x-thead>`),
		[]serBox{
			{"body", BlockBoxT, bc{c: []serBox{
				{"body", TableBoxT, bc{c: []serBox{
					{"x-thead", TableRowGroupBoxT, bc{c: []serBox{
						{"x-thead", TableRowBoxT, bc{c: []serBox{
							{"x-thead", TableCellBoxT, bc{c: []serBox{{"div", BlockBoxT, bc{c: []serBox{}}}}}},
							{"x-td", TableCellBoxT, bc{c: []serBox{}}},
						}}},
					}}},
				}}},
			}}},
		})
}

func TestTables6(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Rule 3.2
	assertTree(t, parseAndBuild(t, "<span><x-tr></x-tr></span>"), []serBox{
		{"body", LineBoxT, bc{c: []serBox{
			{"span", InlineBoxT, bc{c: []serBox{
				{"span", InlineBlockBoxT, bc{c: []serBox{
					{"span", InlineTableBoxT, bc{c: []serBox{{"span", TableRowGroupBoxT, bc{c: []serBox{{"x-tr", TableRowBoxT, bc{c: []serBox{}}}}}}}}},
				}}},
			}}},
		}}},
	})
}

func TestTables7(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Rule 3.1
	// Also, rule 1.3 does ! apply: whitespace before && after is preserved
	assertTree(t, parseAndBuild(t, `
		<span>
		  <em style="display: table-cell"></em>
		  <em style="display: table-cell"></em>
		</span>
	  `), []serBox{
		{"body", LineBoxT, bc{c: []serBox{
			{"span", InlineBoxT, bc{c: []serBox{
				{"span", TextBoxT, bc{text: " "}},
				{"span", InlineBlockBoxT, bc{c: []serBox{
					{"span", InlineTableBoxT, bc{c: []serBox{
						{"span", TableRowGroupBoxT, bc{c: []serBox{
							{"span", TableRowBoxT, bc{c: []serBox{
								{"em", TableCellBoxT, bc{c: []serBox{}}},
								{"em", TableCellBoxT, bc{c: []serBox{}}},
							}}},
						}}},
					}}},
				}}},
				{"span", TextBoxT, bc{text: " "}},
			}}},
		}}},
	})
}

func TestTables8(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)
	// Rule 3.2
	assertTree(t, parseAndBuild(t, "<x-tr></x-tr>\t<x-tr></x-tr>"), []serBox{
		{"body", BlockBoxT, bc{c: []serBox{
			{"body", TableBoxT, bc{c: []serBox{
				{"body", TableRowGroupBoxT, bc{c: []serBox{
					{"x-tr", TableRowBoxT, bc{c: []serBox{}}},
					{"x-tr", TableRowBoxT, bc{c: []serBox{}}},
				}}},
			}}},
		}}},
	})
}

func TestTables9(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parseAndBuild(t, "<x-col></x-col>\n<x-colgroup></x-colgroup>"), []serBox{
		{"body", BlockBoxT, bc{c: []serBox{
			{"body", TableBoxT, bc{c: []serBox{
				{"body", TableColumnGroupBoxT, bc{c: []serBox{{"x-col", TableColumnBoxT, bc{c: []serBox{}}}}}},
				{"x-colgroup", TableColumnGroupBoxT, bc{c: []serBox{{"x-colgroup", TableColumnBoxT, bc{c: []serBox{}}}}}},
			}}},
		}}},
	})
}

func TestTableStyle(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	html := parseAndBuild(t, `<table style="margin: 1px; padding: 2px"></table>`)
	body := html.Box().Children[0]
	wrapper := body.Box().Children[0]
	table := wrapper.Box().Children[0]
	if !(BlockBoxT.IsInstance(wrapper)) {
		t.Fatal()
	}
	if !(TableBoxT.IsInstance(table)) {
		t.Fatal()
	}
	if !(wrapper.Box().Style.GetMarginTop() == pr.FToPx(1)) {
		t.Fatal()
	}
	if !(wrapper.Box().Style.GetPaddingTop() == pr.FToPx(0)) {
		t.Fatal()
	}
	if !(table.Box().Style.GetMarginTop() == pr.FToPx(0)) {
		t.Fatal()
	}
	if !(table.Box().Style.GetPaddingTop() == pr.FToPx(2)) {
		t.Fatal()
	}
}

func TestColumnStyle(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	html := parseAndBuild(t, `
      <table>
        <col span=3 style="width: 10px"></col>
        <col span=2></col>
      </table>
    `)
	body := html.Box().Children[0]
	wrapper := body.Box().Children[0]
	table := wrapper.Box().Children[0].(TableBoxITF)
	colgroup := table.Table().ColumnGroups[0]
	var (
		widths []pr.Value
		gridXs []int
	)
	for _, col := range colgroup.Box().Children {
		widths = append(widths, col.Box().Style.GetWidth())
		gridXs = append(gridXs, col.Box().GridX)
	}
	if !reflect.DeepEqual(widths, []pr.Value{
		pr.FToPx(10), pr.FToPx(10), pr.FToPx(10), pr.SToV("auto"), pr.SToV("auto"),
	}) {
		t.Fatal()
	}
	if !reflect.DeepEqual(gridXs, []int{0, 1, 2, 3, 4}) {
		t.Fatal()
	}
	// copies, not the same box object
	if colgroup.Box().Children[0] == colgroup.Box().Children[1] {
		t.Fatal()
	}
}

func TestNestedGridX(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	html := parseAndBuild(t, `
      <table>
        <col span=2></col>
        <colgroup span=2></colgroup>
        <colgroup>
          <col></col>
          <col span=2></col>
        </colgroup>
        <col></col>
      </table>
    `)
	body := html.Box().Children[0]
	wrapper := body.Box().Children[0]
	table := wrapper.Box().Children[0].(TableBoxITF)

	type gridX struct {
		v    int
		cols []int
	}
	var gridXs []gridX
	for _, colgroup := range table.Table().ColumnGroups {
		v := gridX{v: colgroup.Box().GridX}
		for _, col := range colgroup.Box().Children {
			v.cols = append(v.cols, col.Box().GridX)
		}
		gridXs = append(gridXs, v)
	}
	if !reflect.DeepEqual(gridXs, []gridX{
		{0, []int{0, 1}}, {2, []int{2, 3}}, {4, []int{4, 5, 6}}, {7, []int{7}},
	}) {
		t.Fatal()
	}
}

func extractSpans(group Box) (gridXs, colspans, rowspans [][]int) {
	for _, row := range group.Box().Children {
		var gridX, colspan, rowspan []int
		for _, c := range row.Box().Children {
			gridX = append(gridX, c.Box().GridX)
			colspan = append(colspan, c.Box().Colspan)
			rowspan = append(rowspan, c.Box().Rowspan)
		}
		gridXs = append(gridXs, gridX)
		colspans = append(colspans, colspan)
		rowspans = append(rowspans, rowspan)
	}
	return
}

func TestColspanRowspan1(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// +---+---+---+
	// | A | B | C | X
	// +---+---+---+
	// | D |     E | X
	// +---+---+   +---+
	// |  F ...|   |   |   <-- overlap
	// +---+---+---+   +
	// | H | X   X | G |
	// +---+---+   +   +
	// | I | J | X |   |
	// +---+---+   +---+

	// X: empty cells
	html := parseAndBuild(t, `
      <table>
        <tr>
          <td>A <td>B <td>C
        </tr>
        <tr>
          <td>D <td colspan=2 rowspan=2>E
        </tr>
        <tr>
          <td colspan=2>F <td rowspan=0>G
        </tr>
        <tr>
          <td>H
        </tr>
        <tr>
          <td>I <td>J
        </tr>
      </table>
    `)
	body := html.Box().Children[0]
	wrapper := body.Box().Children[0]
	table := wrapper.Box().Children[0].(TableBoxITF)
	group := table.Box().Children[0]

	gridXs, colspans, rowspans := extractSpans(group)

	if !reflect.DeepEqual(gridXs, [][]int{
		{0, 1, 2},
		{0, 1},
		{0, 3},
		{0},
		{0, 1},
	}) {
		t.Fatal()
	}
	if !reflect.DeepEqual(colspans, [][]int{
		{1, 1, 1},
		{1, 2},
		{2, 1},
		{1},
		{1, 1},
	}) {
		t.Fatal()
	}
	if !reflect.DeepEqual(rowspans, [][]int{
		{1, 1, 1},
		{1, 2},
		{1, 3},
		{1},
		{1, 1},
	}) {
		t.Fatal()
	}
}

func TestColspanRowspan2(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// A cell box cannot extend beyond the last row box of a table.
	html := parseAndBuild(t, `
        <table>
            <tr>
                <td rowspan=5></td>
                <td></td>
            </tr>
            <tr>
                <td></td>
            </tr>
        </table>
    `)
	body := html.Box().Children[0]
	wrapper := body.Box().Children[0]
	table := wrapper.Box().Children[0].(TableBoxITF)
	group := table.Box().Children[0]

	gridXs, colspans, rowspans := extractSpans(group)

	if !reflect.DeepEqual(gridXs, [][]int{
		{0, 1},
		{1},
	}) {
		t.Fatal()
	}
	if !reflect.DeepEqual(colspans, [][]int{
		{1, 1},
		{1},
	}) {
		t.Fatal()
	}
	if !reflect.DeepEqual(rowspans, [][]int{
		{2, 1},
		{1},
	}) {
		t.Fatal()
	}
}

func TestBeforeAfter1(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parseAndBuild(t, `
      <style>
        p:before { content: normal }
        div:before { content: none }
        section::before { color: black }
      </style>
      <p></p>
      <div></div>
      <section></section>
    `), []serBox{
		{"p", BlockBoxT, bc{c: []serBox{}}},
		{"div", BlockBoxT, bc{c: []serBox{}}},
		{"section", BlockBoxT, bc{c: []serBox{}}},
	})
}

func TestBeforeAfter2(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parseAndBuild(t, `
      <style>
        p:before { content: "a" "b" }
        p::after { content: "d" "e" }
      </style>
      <p> c </p>
    `), []serBox{
		{"p", BlockBoxT, bc{c: []serBox{
			{"p", LineBoxT, bc{c: []serBox{
				{"p::before", InlineBoxT, bc{c: []serBox{{"p::before", TextBoxT, bc{text: "ab"}}}}},
				{"p", TextBoxT, bc{text: " c "}},
				{"p::after", InlineBoxT, bc{c: []serBox{{"p::after", TextBoxT, bc{text: "de"}}}}},
			}}},
		}}},
	})
}

func TestBeforeAfter3(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)
	assertTree(t, parseAndBuild(t, `
      <style>
        a[href]:before { content: "[" attr(href) "] " }
      </style>
      <p><a href="some url">some text</a></p>
    `), []serBox{
		{"p", BlockBoxT, bc{c: []serBox{
			{"p", LineBoxT, bc{c: []serBox{
				{"a", InlineBoxT, bc{c: []serBox{
					{"a::before", InlineBoxT, bc{c: []serBox{{"a::before", TextBoxT, bc{text: "[some url] "}}}}},
					{"a", TextBoxT, bc{text: "some text"}},
				}}},
			}}},
		}}},
	})
}

func TestBeforeAfter4(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parseAndBuild(t, `
	<style>
		body { quotes: '«' '»' '“' '”' }
		q:before { content: open-quote ' '}
		q:after { content: ' ' close-quote }
	</style>
  	<p><q>Lorem ipsum <q>dolor</q> sit amet</q></p>
    `), []serBox{
		{"p", BlockBoxT, bc{c: []serBox{
			{"p", LineBoxT, bc{c: []serBox{
				{"q", InlineBoxT, bc{c: []serBox{
					{"q::before", InlineBoxT, bc{c: []serBox{{"q::before", TextBoxT, bc{text: "« "}}}}},
					{"q", TextBoxT, bc{text: "Lorem ipsum "}},
					{"q", InlineBoxT, bc{c: []serBox{
						{"q::before", InlineBoxT, bc{c: []serBox{{"q::before", TextBoxT, bc{text: "“ "}}}}},
						{"q", TextBoxT, bc{text: "dolor"}},
						{"q::after", InlineBoxT, bc{c: []serBox{{"q::after", TextBoxT, bc{text: " ”"}}}}},
					}}},
					{"q", TextBoxT, bc{text: " sit amet"}},
					{"q::after", InlineBoxT, bc{c: []serBox{{"q::after", TextBoxT, bc{text: " »"}}}}},
				}}},
			}}},
		}}},
	})
}

func TestBeforeAfter5(t *testing.T) {
	cp := testutils.CaptureLogs()

	assertTree(t, parseAndBuild(t, `
          <style>
            p:before {
              content: "a" url(pattern.png) "b";

              /* Invalid, ignored in favor of the one above.
                 Regression test: this used to crash: */
              content: some-function(nested-function(something));
            }
          </style>
          <p>c</p>
        `), []serBox{
		{"p", BlockBoxT, bc{c: []serBox{
			{"p", LineBoxT, bc{c: []serBox{
				{"p::before", InlineBoxT, bc{c: []serBox{
					{"p::before", TextBoxT, bc{text: "a"}},
					{"p::before", InlineReplacedBoxT, bc{text: "<replaced>"}},
					{"p::before", TextBoxT, bc{text: "b"}},
				}}},
				{"p", TextBoxT, bc{text: "c"}},
			}}},
		}}},
	})

	logs := cp.Logs()
	if L := len(logs); L != 1 {
		t.Fatalf("expected 1 log, got %d", L)
	}
	if !strings.Contains(logs[0], "nested-function(") {
		t.Fatalf("unexpected log: %s", logs[0])
	}
	if !strings.Contains(logs[0], "invalid value") {
		t.Fatalf("unexpected log: %s", logs[0])
	}
}

var (
	black       = pr.NewColor(0, 0, 0, 1)
	red         = pr.NewColor(1, 0, 0, 1)
	green       = pr.NewColor(0, 1, 0, 1) // lime in CSS
	blue        = pr.NewColor(0, 0, 1, 1)
	yellow      = pr.NewColor(1, 1, 0, 1)
	black3      = Border{Style: "solid", Width: 3, Color: black}
	red1        = Border{Style: "solid", Width: 1, Color: red}
	yellow5     = Border{Style: "solid", Width: 5, Color: yellow}
	green5      = Border{Style: "solid", Width: 5, Color: green}
	dashedBlue5 = Border{Style: "dashed", Width: 5, Color: blue}
)

func TestBorderCollapse1(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	html := parseAndBuild(t, "<table></table>")

	body := html.Box().Children[0]
	wrapper := body.Box().Children[0]
	table := wrapper.Box().Children[0].(TableBoxITF)

	if !(table.Table().CollapsedBorderGrid.Horizontal == nil && table.Table().CollapsedBorderGrid.Vertical == nil) {
		t.Fatal()
	}

	gridH, gridV := getGrid(t, `<table style="border-collapse: collapse"></table>`)

	if len(gridH) != 0 || len(gridV) != 0 {
		t.Fatal()
	}
}

func TestBorderCollapse2(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	verticalBorders, horizontalBorders := getGrid(t, `
      <style>td { border: 1px solid red }</style>
      <table style="border-collapse: collapse; border: 3px solid black">
        <tr> <td>A</td> <td>B</td> </tr>
        <tr> <td>C</td> <td>D</td> </tr>
      </table>
    `)
	if !reflect.DeepEqual(verticalBorders, [][]Border{
		{black3, red1, black3},
		{black3, red1, black3},
	}) {
		t.Fatalf("unexepected vertical borders %v", verticalBorders)
	}
	if !reflect.DeepEqual(horizontalBorders, [][]Border{
		{black3, black3},
		{red1, red1},
		{black3, black3},
	}) {
		t.Fatalf("unexepected horizontal borders: %v", horizontalBorders)
	}
}

func TestBorderCollapse3(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// hidden vs. none
	verticalBorders, horizontalBorders := getGrid(t, `
      <style>table, td { border: 3px solid }</style>
      <table style="border-collapse: collapse">
        <tr> <td>A</td> <td style="border-style: hidden">B</td> </tr>
        <tr> <td>C</td> <td style="border-style: none">D</td> </tr>
      </table>
    `)
	if !reflect.DeepEqual(verticalBorders, [][]Border{
		{black3, Border{}, Border{}},
		{black3, black3, black3},
	}) {
		t.Fatalf("unexepected vertical borders %v", verticalBorders)
	}
	if !reflect.DeepEqual(horizontalBorders, [][]Border{
		{black3, Border{}},
		{black3, Border{}},
		{black3, black3},
	}) {
		t.Fatalf("unexepected horizontal borders: %v", horizontalBorders)
	}
}

func TestBorderCollapse4(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	verticalBorders, horizontalBorders := getGrid(t, `
      <style>td { border: 1px solid red }</style>
      <table style="border-collapse: collapse; border: 5px solid yellow">
        <col style="border: 3px solid black" />
        <tr> <td></td> <td></td> <td></td> </tr>
        <tr> <td></td> <td style="border: 5px dashed blue"></td>
          <td style="border: 5px solid lime"></td> </tr>
        <tr> <td></td> <td></td> <td></td> </tr>
        <tr> <td></td> <td></td> <td></td> </tr>
      </table>
    `)

	if !reflect.DeepEqual(verticalBorders, [][]Border{
		{yellow5, black3, red1, yellow5},
		{yellow5, dashedBlue5, green5, green5},
		{yellow5, black3, red1, yellow5},
		{yellow5, black3, red1, yellow5},
	}) {
		t.Fatalf("unexepected vertical borders %v", verticalBorders)
	}
	if !reflect.DeepEqual(horizontalBorders, [][]Border{
		{yellow5, yellow5, yellow5},
		{red1, dashedBlue5, green5},
		{red1, dashedBlue5, green5},
		{red1, red1, red1},
		{yellow5, yellow5, yellow5},
	}) {
		t.Fatalf("unexepected horizontal borders: %v", horizontalBorders)
	}
}

func TestBorderCollapse5(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// rowspan && colspan
	verticalBorders, horizontalBorders := getGrid(t, `
        <style>col, tr { border: 3px solid }</style>
        <table style="border-collapse: collapse">
            <col /><col /><col />
            <tr> <td rowspan=2></td> <td></td> <td></td> </tr>
            <tr>                     <td colspan=2></td> </tr>
        </table>
    `)

	if !reflect.DeepEqual(verticalBorders, [][]Border{
		{black3, black3, black3, black3},
		{black3, black3, Border{}, black3},
	}) {
		t.Fatalf("unexepected vertical borders %v", verticalBorders)
	}
	if !reflect.DeepEqual(horizontalBorders, [][]Border{
		{black3, black3, black3},
		{Border{}, black3, black3},
		{black3, black3, black3},
	}) {
		t.Fatalf("unexepected horizontal borders: %v", horizontalBorders)
	}
}

func testDisplayNoneRoot(t *testing.T, html string) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	box := parseAndBuild(t, html)
	if d := box.Box().Style.GetDisplay(); d != "block" && d != "flow" {
		t.Fatal()
	}
	if len(box.Box().Children) != 0 {
		t.Fatal()
	}
}

func TestDisplayNoneRoot(t *testing.T) {
	for _, html := range []string{
		`<html style="display: none">`,
		`<html style="display: none">abc`,
		`<html style="display: none"><p>abc`,
		`<body style="display: none"><p>abc`,
	} {
		testDisplayNoneRoot(t, html)
	}
}

func TestBuildPages(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parseAndBuild(t, `
	<style>
		@page {
		/* Make the page content area only 10px high and wide,
			so every word in <p> end up on a page of its own. */
		size: 30px;
		margin: 10px;
		@top-center { content: "Title" }
		}
		@page :first {
		@bottom-left { content: "foo" }
		@bottom-left-corner { content: "baz" }
		}
	</style>
	<p>lorem ipsum
	`), []serBox{
		{"p", BlockBoxT, bc{c: []serBox{{"p", LineBoxT, bc{c: []serBox{{"p", TextBoxT, bc{text: "lorem ipsum "}}}}}}}},
	})
}

func TestInlineSpace(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parseAndBuild(t, `
	<p>start <i><b>bi1</b> <b>bi2</b></i> <b>b1</b> end</p>
	`), []serBox{
		{"p", BlockBoxT, bc{c: []serBox{
			{"p", LineBoxT, bc{c: []serBox{
				{"p", TextBoxT, bc{text: "start "}},
				{"i", InlineBoxT, bc{c: []serBox{
					{"b", InlineBoxT, bc{c: []serBox{{"b", TextBoxT, bc{text: "bi1"}}}}},
					{"i", TextBoxT, bc{text: " "}},
					{"b", InlineBoxT, bc{c: []serBox{{"b", TextBoxT, bc{text: "bi2"}}}}},
				}}},
				{"p", TextBoxT, bc{text: " "}},
				{"b", InlineBoxT, bc{c: []serBox{{"b", TextBoxT, bc{text: "b1"}}}}},
				{"p", TextBoxT, bc{text: " end"}},
			}}},
		}}},
	})
}
