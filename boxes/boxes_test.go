package boxes

import (
	"errors"
	"reflect"
	"testing"

	"github.com/benoitkugler/go-weasyprint/boxes/counters"
	"github.com/benoitkugler/go-weasyprint/images"
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

// Like parse() but also run all corrections on boxes.
func parse(t *testing.T, html_content string) BlockLevelBoxITF {
	box := BuildFormattingStructure(parseBase(t, utils.InputString(html_content), baseUrl))
	if err := sanityChecks(box); err != nil {
		t.Fatalf("sanity check failed: %s", err)
	}
	return box
}

type serializedBox struct {
	tag     string
	content struct {
		text      string
		childrens []serializedBox
	}
	type_ BoxType
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
			out[i].content.childrens = serialize(cg)
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

	if !reflect.DeepEqual(serialize(box.Box().Children), expected) {
		t.Fatalf("expected %v\n, got\n%v", expected, box.Box().Children)
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

// func _parse_base(html_content, base_url=BASE_URL):
//     document = FakeHTML(string=html_content, base_url=base_url)
//     counter_style = CounterStyle()
//     style_for = get_all_computed_styles(document, counter_style=counter_style)
//     get_image_from_uri = functools.partial(
//         images.get_image_from_uri, cache={}, url_fetcher=document.url_fetcher,
//         optimize_size=())
//     target_collector = TargetCollector()
//     return (
//         document.etree_element, style_for, get_image_from_uri, base_url,
//         target_collector, counter_style)

const baseUrl = "../resources_test/"

func getGrid(t *testing.T, html string) ([][]*Border, [][]*Border) {
	root := parse(t, html)
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

var none = struct {
	text      string
	childrens []serializedBox
}{}

func TestBoxTree(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parse(t, "<p>"), []serializedBox{{"p", none, TypeBlockBox}})

	// assertTree(t, parse("""
	//   <style>
	//     span { display: inline-block }
	//   </style>
	//   <p>Hello <em>World <img src="pattern.png"><span>L</span></em>!</p>"""), [
	//       ("p", "Block", [
	//         ("p", "Text", "Hello "),
	//         ("em", "Inline", [
	//             ("em", "Text", "World "),
	//             ("img", "InlineReplaced", "<replaced>"),
	//             ("span", "InlineBlock", [
	//                 ("span", "Text", "L")])]),
	//         ("p", "Text", "!")])])
}
