package boxes

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/benoitkugler/go-weasyprint/boxes/counters"
	"github.com/benoitkugler/go-weasyprint/images"
	"github.com/benoitkugler/go-weasyprint/style/parser"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/style/tree"
	"github.com/benoitkugler/go-weasyprint/utils"
	"github.com/benoitkugler/go-weasyprint/utils/testutils"
)

func TestInheritance(t *testing.T) {
	// u := NewInlineBox("", nil, nil)
	// u.RemoveDecoration(nil, true, true)
}

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

func parseBase(t *testing.T, content utils.ContentInput, baseUrl string) (*utils.HTMLNode, tree.StyleFor, Gifu, string, *tree.TargetCollector) {
	html, err := tree.NewHTML(content, baseUrl, utils.DefaultUrlFetcher, "")
	if err != nil {
		t.Fatalf("parsing HTML failed: %s", err)
	}
	document := fakeHTML(html)
	cs := make(counters.CounterStyle)
	style := tree.GetAllComputedStyles(document, nil, false, nil, cs, nil, nil)
	imgFetcher := func(url string, forcedMimeType string) images.Image {
		return images.GetImageFromUri(make(map[string]images.Image), document.UrlFetcher, url, forcedMimeType)
	}
	tr := tree.NewTargetCollector()
	return document.Root, *style, imgFetcher, baseUrl, &tr
}

func parse(t *testing.T, htmlContent string) BoxITF {
	a, b, c, d, e := parseBase(t, utils.InputString(htmlContent), baseUrl)
	boxes := elementToBox(a, b, c, d, e, nil)
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
	c    []serializedBox
}

type serializedBox struct {
	tag     string
	type_   BoxType
	content bc
}

func (s serializedBox) equals(other serializedBox) bool {
	if s.tag != other.tag || s.type_ != other.type_ || s.content.text != other.content.text {
		return false
	}
	return serializedBoxEquals(s.content.c, other.content.c)
}

func serializedBoxEquals(l1, l2 []serializedBox) bool {
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
func serialize(boxList []Box) []serializedBox {
	out := make([]serializedBox, len(boxList))
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
func assertTree(t *testing.T, box Box, expected []serializedBox) {
	if tag := box.Box().ElementTag; tag != "html" {
		t.Fatalf("unexpected element: %s", tag)
	}
	if !TypeBlockBox.IsInstance(box) {
		t.Fatal("expected block box")
	}
	if L := len(box.Box().Children); L != 1 {
		t.Fatalf("expected one children, got %d", L)
	}

	box = box.Box().Children[0]
	if !TypeBlockBox.IsInstance(box) {
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
	TypeBlockContainerBox: {TypeBlockLevelBox, TypeLineBox},
	TypeLineBox:           {TypeInlineLevelBox},
	TypeInlineBox:         {TypeInlineLevelBox},
	TypeTableBox: {
		TypeTableCaptionBox,
		TypeTableColumnGroupBox, TypeTableColumnBox,
		TypeTableRowGroupBox, TypeTableRowBox,
	},
	TypeInlineTableBox: {
		TypeTableCaptionBox,
		TypeTableColumnGroupBox, TypeTableColumnBox,
		TypeTableRowGroupBox, TypeTableRowBox,
	},
	TypeTableColumnGroupBox: {TypeTableColumnBox},
	TypeTableRowGroupBox:    {TypeTableRowBox},
	TypeTableRowBox:         {TypeTableCellBox},
}

// Check that the rules regarding boxes are met.
//
// This is not required and only helps debugging.
//
// - A block container can contain either only block-level boxes or
//   only line boxes;
// - Line boxes and inline boxes can only contain inline-level boxes.
func sanityChecks(box Box) error {
	if !TypeParentBox.IsInstance(box) {
		return nil
	}

	acceptableTypesLists, ok := properChildren[box.Type()]
	if !ok {
		return nil // this is less strict than the reference implementation
	}

	for _, child := range box.Box().Children {
		if !child.Box().IsInNormalFlow() {
			continue
		}
		isOk := false
		for _, typeOk := range acceptableTypesLists {
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

func getGrid(t *testing.T, html string) ([][]*Border, [][]*Border) {
	root := parseAndBuild(t, html)
	body := root.Box().Children[0]
	tableWrapper := body.Box().Children[0]
	table := tableWrapper.Box().Children[0].(TableBoxITF)

	buildGrid := func(bg [][]Border) (grid [][]*Border /*maybe nil*/) {
		for _, column := range bg {
			out := make([]*Border, len(column))
			for i, border := range column {
				if border.Width != 0 {
					out[i] = &border
				}
			}
			grid = append(grid, out)
		}
		return grid
	}

	return buildGrid(table.Table().CollapsedBorderGrid.Horizontal),
		buildGrid(table.Table().CollapsedBorderGrid.Vertical)
}

func TestBoxTree(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parse(t, "<p>"), []serializedBox{{"p", TypeBlockBox, bc{}}})
	assertTree(t, parse(t, `
	  <style>
	    span { display: inline-block }
	  </style>
	  <p>Hello <em>World <img src="pattern.png"><span>L</span></em>!</p>`),
		[]serializedBox{
			{
				"p", TypeBlockBox, bc{c: []serializedBox{
					{"p", TypeTextBox, bc{text: "Hello "}},
					{"em", TypeInlineBox, bc{c: []serializedBox{
						{"em", TypeTextBox, bc{text: "World "}},
						{"img", TypeInlineReplacedBox, bc{text: "<replaced>"}},
						{"span", TypeInlineBlockBox, bc{c: []serializedBox{
							{"span", TypeTextBox, bc{text: "L"}},
						}}},
					}}},
					{"p", TypeTextBox, bc{text: "!"}},
				}},
			},
		})
}

func TestHtmlEntities(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	for _, quote := range []string{`"`, "&quot;", "&#x22;", "&#34;"} {
		assertTree(t, parse(t, fmt.Sprintf("<p>%sabc%s", quote, quote)), []serializedBox{
			{"p", TypeBlockBox, bc{c: []serializedBox{
				{"p", TypeTextBox, bc{text: `"abc"`}},
			}}},
		})
	}
}

func TestInlineInBlock1(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	source := "<div>Hello, <em>World</em>!\n<p>Lipsum.</p></div>"
	expected := []serializedBox{
		{"div", TypeBlockBox, bc{
			c: []serializedBox{
				{
					"div", TypeBlockBox,
					bc{c: []serializedBox{
						{"div", TypeLineBox, bc{c: []serializedBox{
							{"div", TypeTextBox, bc{text: "Hello, "}},
							{"em", TypeInlineBox, bc{c: []serializedBox{
								{"em", TypeTextBox, bc{text: "World"}},
							}}},
							{"div", TypeTextBox, bc{text: "!\n"}},
						}}},
					}},
				},
				{"p", TypeBlockBox, bc{c: []serializedBox{
					{"p", TypeLineBox, bc{c: []serializedBox{
						{"p", TypeTextBox, bc{text: "Lipsum."}},
					}}},
				}}},
			},
		}},
	}
	box := parse(t, source)

	assertTree(t, box, []serializedBox{
		{"div", TypeBlockBox, bc{c: []serializedBox{
			{"div", TypeTextBox, bc{text: "Hello, "}},
			{"em", TypeInlineBox, bc{c: []serializedBox{
				{"em", TypeTextBox, bc{text: "World"}},
			}}},
			{"div", TypeTextBox, bc{text: "!\n"}},
			{"p", TypeBlockBox, bc{c: []serializedBox{{"p", TypeTextBox, bc{text: "Lipsum."}}}}},
		}}},
	})

	box = InlineInBlock(box)
	assertTree(t, box, expected)
}

func TestInlineInBlock2(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	source := "<div><p>Lipsum.</p>Hello, <em>World</em>!\n</div>"
	expected := []serializedBox{
		{"div", TypeBlockBox, bc{c: []serializedBox{
			{"p", TypeBlockBox, bc{c: []serializedBox{{"p", TypeLineBox, bc{c: []serializedBox{{"p", TypeTextBox, bc{text: "Lipsum."}}}}}}}},
			{"div", TypeBlockBox, bc{c: []serializedBox{
				{"div", TypeLineBox, bc{c: []serializedBox{
					{"div", TypeTextBox, bc{text: "Hello, "}},
					{"em", TypeInlineBox, bc{c: []serializedBox{{"em", TypeTextBox, bc{text: "World"}}}}},
					{"div", TypeTextBox, bc{text: "!\n"}},
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
	expected := []serializedBox{
		{"p", TypeBlockBox, bc{c: []serializedBox{
			{"p", TypeLineBox, bc{c: []serializedBox{
				{"p", TypeTextBox, bc{text: "Hello "}},
				{"em", TypeBlockBox, bc{c: []serializedBox{{"em", TypeLineBox, bc{c: []serializedBox{{"em", TypeTextBox, bc{text: "World"}}}}}}}},
				{"p", TypeTextBox, bc{text: "!"}},
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

	expected := []serializedBox{
		{"p", TypeBlockBox, bc{c: []serializedBox{
			{"p", TypeLineBox, bc{c: []serializedBox{
				{"p", TypeTextBox, bc{text: "Hello "}},
				{"em", TypeBlockBox, bc{c: []serializedBox{{"em", TypeLineBox, bc{c: []serializedBox{{"em", TypeTextBox, bc{text: "World"}}}}}}}},
				{"p", TypeTextBox, bc{text: "!"}},
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
	assertTree(t, box, []serializedBox{
		{"body", TypeLineBox, bc{c: []serializedBox{
			{"p", TypeInlineBlockBox, bc{c: []serializedBox{
				{"p", TypeLineBox, bc{c: []serializedBox{
					{"p", TypeTextBox, bc{text: "Lorem "}},
					{"em", TypeInlineBox, bc{c: []serializedBox{
						{"em", TypeTextBox, bc{text: "ipsum "}},
						{"strong", TypeInlineBox, bc{c: []serializedBox{
							{"strong", TypeTextBox, bc{text: "dolor "}},
							{"span", TypeBlockBox, bc{c: []serializedBox{{"span", TypeLineBox, bc{c: []serializedBox{{"span", TypeTextBox, bc{text: "sit"}}}}}}}},
							{"strong", TypeTextBox, bc{text: "\n      "}},
							{"span", TypeBlockBox, bc{c: []serializedBox{{"span", TypeLineBox, bc{c: []serializedBox{{"span", TypeTextBox, bc{text: "amet,"}}}}}}}},
						}}},
						{"span", TypeBlockBox, bc{c: []serializedBox{
							{"span", TypeLineBox, bc{c: []serializedBox{
								{"em", TypeInlineBox, bc{c: []serializedBox{
									{"em", TypeTextBox, bc{text: "conse"}},
									{"i", TypeBlockBox, bc{c: []serializedBox{}}},
								}}},
							}}},
						}}},
					}}},
				}}},
			}}},
		}}},
	})

	box = BlockInInline(box)
	assertTree(t, box, []serializedBox{
		{"body", TypeLineBox, bc{c: []serializedBox{
			{"p", TypeInlineBlockBox, bc{c: []serializedBox{
				{"p", TypeBlockBox, bc{c: []serializedBox{
					{"p", TypeLineBox, bc{c: []serializedBox{
						{"p", TypeTextBox, bc{text: "Lorem "}},
						{"em", TypeInlineBox, bc{c: []serializedBox{
							{"em", TypeTextBox, bc{text: "ipsum "}},
							{"strong", TypeInlineBox, bc{c: []serializedBox{{"strong", TypeTextBox, bc{text: "dolor "}}}}},
						}}},
					}}},
				}}},
				{"span", TypeBlockBox, bc{c: []serializedBox{{"span", TypeLineBox, bc{c: []serializedBox{{"span", TypeTextBox, bc{text: "sit"}}}}}}}},
				{"p", TypeBlockBox, bc{c: []serializedBox{
					{"p", TypeLineBox, bc{c: []serializedBox{
						{"em", TypeInlineBox, bc{c: []serializedBox{{"strong", TypeInlineBox, bc{c: []serializedBox{{"strong", TypeTextBox, bc{text: "\n      "}}}}}}}},
					}}},
				}}},
				{"span", TypeBlockBox, bc{c: []serializedBox{{"span", TypeLineBox, bc{c: []serializedBox{{"span", TypeTextBox, bc{text: "amet,"}}}}}}}},
				{"p", TypeBlockBox, bc{c: []serializedBox{
					{"p", TypeLineBox, bc{c: []serializedBox{{"em", TypeInlineBox, bc{c: []serializedBox{{"strong", TypeInlineBox, bc{c: []serializedBox{}}}}}}}}},
				}}},
				{"span", TypeBlockBox, bc{c: []serializedBox{
					{"span", TypeBlockBox, bc{c: []serializedBox{
						{"span", TypeLineBox, bc{c: []serializedBox{{"em", TypeInlineBox, bc{c: []serializedBox{{"em", TypeTextBox, bc{text: "conse"}}}}}}}},
					}}},
					{"i", TypeBlockBox, bc{c: []serializedBox{}}},
					{"span", TypeBlockBox, bc{c: []serializedBox{{"span", TypeLineBox, bc{c: []serializedBox{{"em", TypeInlineBox, bc{c: []serializedBox{}}}}}}}}},
				}}},
				{"p", TypeBlockBox, bc{c: []serializedBox{{"p", TypeLineBox, bc{c: []serializedBox{{"em", TypeInlineBox, bc{c: []serializedBox{}}}}}}}}},
			}}},
		}}},
	})
}

func pxToValue(px pr.Float) pr.Value {
	return pr.Dimension{Unit: pr.Px, Value: px}.ToValue()
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
		if mt := child.Box().Style.GetMarginTop(); mt != pxToValue(0) && mt != pxToValue(42) {
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
		"<pre style=\"white-space: pre-line\">\t  foo\n</pre>"), []serializedBox{
		{"p", TypeBlockBox, bc{c: []serializedBox{
			{"p", TypeLineBox, bc{c: []serializedBox{
				{"p", TypeTextBox, bc{text: "Lorem ipsum "}},
				{"strong", TypeInlineBox, bc{c: []serializedBox{
					{"strong", TypeTextBox, bc{text: "dolor "}},
					{"img", TypeInlineReplacedBox, bc{text: "<replaced>"}},
					{"strong", TypeTextBox, bc{text: " sit "}},
					{"span", TypeBlockBox, bc{c: []serializedBox{}}},
					{"em", TypeInlineBox, bc{c: []serializedBox{{"em", TypeTextBox, bc{text: "amet "}}}}},
					{"strong", TypeTextBox, bc{text: "consectetur"}},
				}}},
				{"p", TypeTextBox, bc{text: "."}},
			}}},
		}}},
		{"pre", TypeBlockBox, bc{c: []serializedBox{{"pre", TypeLineBox, bc{c: []serializedBox{{"pre", TypeTextBox, bc{text: "\t  foo\n"}}}}}}}},
		{"pre", TypeBlockBox, bc{c: []serializedBox{{"pre", TypeLineBox, bc{c: []serializedBox{{"pre", TypeTextBox, bc{text: "\t  foo\n"}}}}}}}},
		{"pre", TypeBlockBox, bc{c: []serializedBox{{"pre", TypeLineBox, bc{c: []serializedBox{{"pre", TypeTextBox, bc{text: " foo\n"}}}}}}}},
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
	styleFor := tree.GetAllComputedStyles(document, nil, false, nil, nil, nil, nil)

	// Force the generation of the style for this page type as it"s generally
	// only done during the rendering.
	styleFor.SetPageTypeComputedStyles(data.type_, document)

	style := styleFor.Get(data.type_, "")
	if m := style.GetMarginTop(); m != pxToValue(data.top) {
		t.Fatalf("expected %f, got %v", data.top, m)
	}
	if m := style.GetMarginRight(); m != pxToValue(data.right) {
		t.Fatalf("expected %f, got %v", data.right, m)
	}
	if m := style.GetMarginBottom(); m != pxToValue(data.bottom) {
		t.Fatalf("expected %f, got %v", data.bottom, m)
	}
	if m := style.GetMarginLeft(); m != pxToValue(data.left) {
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
	assertTree(t, result, []serializedBox{
		{"p", TypeBlockBox, bc{c: []serializedBox{
			{"p", TypeLineBox, bc{c: []serializedBox{
				{"img", TypeInlineReplacedBox, bc{text: "<replaced>"}},
				{"img", TypeInlineBox, bc{c: []serializedBox{{"img", TypeTextBox, bc{text: "No src"}}}}},
				{"img", TypeInlineBox, bc{c: []serializedBox{{"img", TypeTextBox, bc{text: "Inexistent src"}}}}},
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
	assertTree(t, result, []serializedBox{
		{"p", TypeBlockBox, bc{c: []serializedBox{
			{"p", TypeLineBox, bc{c: []serializedBox{{"img", TypeInlineBox, bc{c: []serializedBox{{"img", TypeTextBox, bc{text: "No baseUrl"}}}}}}}},
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
    `), []serializedBox{
		{"x-table", TypeBlockBox, bc{c: []serializedBox{
			{"x-caption", TypeTableCaptionBox, bc{c: []serializedBox{{"x-caption", TypeLineBox, bc{c: []serializedBox{{"x-caption", TypeTextBox, bc{text: "top caption"}}}}}}}},
			{"x-table", TypeTableBox, bc{c: []serializedBox{
				{"x-table", TypeTableColumnGroupBox, bc{c: []serializedBox{{"x-col", TypeTableColumnBox, bc{c: []serializedBox{}}}}}},
				{"x-thead", TypeTableRowGroupBox, bc{c: []serializedBox{{"x-thead", TypeTableRowBox, bc{c: []serializedBox{{"x-th", TypeTableCellBox, bc{c: []serializedBox{}}}}}}}}},
				{"x-table", TypeTableRowGroupBox, bc{c: []serializedBox{
					{"x-tr", TypeTableRowBox, bc{c: []serializedBox{
						{"x-th", TypeTableCellBox, bc{c: []serializedBox{{"x-th", TypeLineBox, bc{c: []serializedBox{{"x-th", TypeTextBox, bc{text: "foo"}}}}}}}},
						{"x-th", TypeTableCellBox, bc{c: []serializedBox{{"x-th", TypeLineBox, bc{c: []serializedBox{{"x-th", TypeTextBox, bc{text: "bar"}}}}}}}},
					}}},
				}}},
				{"x-thead", TypeTableRowGroupBox, bc{c: []serializedBox{}}},
				{"x-table", TypeTableRowGroupBox, bc{c: []serializedBox{
					{"x-tr", TypeTableRowBox, bc{c: []serializedBox{
						{"x-td", TypeTableCellBox, bc{c: []serializedBox{{"x-td", TypeLineBox, bc{c: []serializedBox{{"x-td", TypeTextBox, bc{text: "baz"}}}}}}}},
					}}},
				}}},
				{"x-tfoot", TypeTableRowGroupBox, bc{c: []serializedBox{}}},
			}}},
			{"x-caption", TypeTableCaptionBox, bc{c: []serializedBox{}}},
		}}},
	})
}
