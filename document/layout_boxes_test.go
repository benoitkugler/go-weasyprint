package document

import (
	"log"
	"reflect"
	"testing"

	bo "github.com/benoitkugler/go-weasyprint/boxes"
	"github.com/benoitkugler/go-weasyprint/layout/text"
	"github.com/benoitkugler/go-weasyprint/style/tree"
	"github.com/benoitkugler/go-weasyprint/utils"
	"github.com/benoitkugler/go-weasyprint/utils/testutils"
	fc "github.com/benoitkugler/textlayout/fontconfig"
)

var baseUrl, _ = utils.Path2url("../resources_test/")

const fontmapCache = "../layout/text/test/cache.fc"

var fontconfig *text.FontConfiguration

func init() {
	// this command has to run once
	// fmt.Println("Scanning fonts...")
	// _, err := fc.ScanAndCache(fontmapCache)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	fs, err := fc.LoadFontsetFile(fontmapCache)
	if err != nil {
		log.Fatal(err)
	}
	fontconfig = text.NewFontConfiguration(fc.Standard, fs)
}

func fakeHTML(html *tree.HTML) *tree.HTML {
	html.UAStyleSheet = tree.TestUAStylesheet
	return html
}

// lay out a document and return a list of PageBox objects
func renderPages(t *testing.T, htmlContent string) []*bo.PageBox {
	doc, err := tree.NewHTML(utils.InputString(htmlContent), baseUrl, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	doc = fakeHTML(doc)
	renderedDoc := Render(doc, nil, false, fontconfig, nil)
	out := make([]*bo.PageBox, len(renderedDoc.Pages))
	for i, v := range renderedDoc.Pages {
		out[i] = v.pageBox
	}
	return out
}

func TestMarginBoxes(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	pages := renderPages(t, `
		<style>
		  @page {
			/* Make the page content area only 10px high && wide,
			   so every word := range <p> end up on a page of its own. */
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
	  `)
	if len(pages) != 2 {
		t.Fatal()
	}
	page1, page2 := pages[0], pages[1]
	if page1.Children[0].Box().ElementTag != "html" {
		t.Fatal()
	}
	if page2.Children[0].Box().ElementTag != "html" {
		t.Fatal()
	}

	var marginBoxes1, marginBoxes2 []string
	for _, box := range page1.Children[1:] {
		marginBoxes1 = append(marginBoxes1, box.(*bo.MarginBox).AtKeyword)
	}
	for _, box := range page2.Children[1:] {
		marginBoxes1 = append(marginBoxes1, box.(*bo.MarginBox).AtKeyword)
	}
	if !reflect.DeepEqual(marginBoxes1, []string{"@top-center", "@bottom-left", "@bottom-left-corner"}) {
		t.Fatal()
	}
	if !reflect.DeepEqual(marginBoxes2, []string{"@top-center"}) {
		t.Fatal()
	}

	if len(page1.Children) != 2 {
		t.Fatal()
	}
	_, topCenter := page2.Children[0], page2.Children[1]
	lineBox := topCenter.Box().Children[0]
	textBox, ok := lineBox.Box().Children[0].(*bo.TextBox)
	if !ok || textBox.Text != "Title" {
		t.Fatal()
	}
}

// func TestMarginBoxStringSet1(t *testing.T){
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
//     // Test that both pages get string := range the `bottom-center` margin box
//     page1, page2 = renderPages(`
//       <style>
//         @page {
//           @bottom-center { content: string(textHeader) }
//         }
//         p {
//           string-set: textHeader content();
//         }
//         .page {
//           page-break-before: always;
//         }
//       </style>
//       <p>first assignment</p>
//       <div class="page"></div>
//     `)
// }
//     html, bottomCenter = page2.Children
//     lineBox, = bottomCenter.Children
//     textBox, = lineBox.Children
//     assert textBox.text == "first assignment"

//     html, bottomCenter = page1.Children
//     lineBox, = bottomCenter.Children
//     textBox, = lineBox.Children
//     assert textBox.text == "first assignment"

// func TestMarginBoxStringSet2(t *testing.T) {
//   cp := testutils.CaptureLogs()
//   defer cp.AssertNoLogs(t)

//     def simpleStringSetTest(contentVal, extraStyle="") {
//         page1, = renderPages(`
//           <style>
//             @page {
//               @top-center { content: string(textHeader) }
//             }
//             p {
//               string-set: textHeader content(%(contentVal)s);
//             }
//             %(extraStyle)s
//           </style>
//           <p>first assignment</p>
//         ` % dict(contentVal=contentVal, extraStyle=extraStyle))

//         html, topCenter = page1.Children
//         lineBox, = topCenter.Children
//         textBox, = lineBox.Children
//         if contentVal := range ("before", "after"):
//             assert textBox.text == "pseudo"
//         else:
//             assert textBox.text == "first assignment"

//     // Test each accepted value of `content()` as an arguemnt to `string-set`
//     for value := range ("", "text", "before", "after"):
//         if value := range ("before", "after"):
//             extraStyle = "p:%s{content: "pseudo"}" % value
//             simpleStringSetTest(value, extraStyle)
//         else:
//             simpleStringSetTest(value)

// func TestMarginBoxStringSet3(t *testing.T){
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
//     // Test `first` (default value) ie. use the first assignment on the page
//     page1, = renderPages(`
//       <style>
//         @page {
//           @top-center { content: string(textHeader, first) }
//         }
//         p {
//           string-set: textHeader content();
//         }
//       </style>
//       <p>first assignment</p>
//       <p>Second assignment</p>
//     } `)
// }
//     html, topCenter = page1.Children
//     lineBox, = topCenter.Children
//     textBox, = lineBox.Children
//     assert textBox.text == "first assignment"

// func TestMarginBoxStringSet4(t *testing.T) {
//   cp := testutils.CaptureLogs()
//   defer cp.AssertNoLogs(t)

//     // test `first-except` ie. exclude from page on which value is assigned
//     page1, page2 = renderPages(`
//       <style>
//         @page {
//           @top-center { content: string(headerNofirst, first-except) }
//         }
//         p{
//           string-set: headerNofirst content();
//         }
//         .page{
//           page-break-before: always;
//         }
//       </style>
//       <p>firstExcepted</p>
//       <div class="page"></div>
//     `)
//     html, topCenter = page1.Children
//     assert len(topCenter.Children) == 0

//     html, topCenter = page2.Children
//     lineBox, = topCenter.Children
//     textBox, = lineBox.Children
//     assert textBox.text == "firstExcepted"

// func TestMarginBoxStringSet5(t *testing.T){
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
//     // Test `last` ie. use the most-recent assignment
//     page1, = renderPages(`
//       <style>
//         @page {
//           @top-center { content: string(headerLast, last) }
//         }
//         p {
//           string-set: headerLast content();
//         }
//       </style>
//       <p>String set</p>
//       <p>Second assignment</p>
//     `)
// }
//     html, topCenter = page1.Children[:2]
//     lineBox, = topCenter.Children

//     textBox, = lineBox.Children
//     assert textBox.text == "Second assignment"

// func TestMarginBoxStringSet6(t *testing.T) {
//   cp := testutils.CaptureLogs()
//   defer cp.AssertNoLogs(t)

//     // Test multiple complex string-set values
//     page1, = renderPages(`
//       <style>
//         @page {
//           @top-center { content: string(textHeader, first) }
//           @bottom-center { content: string(textFooter, last) }
//         }
//         html { counter-reset: a }
//         body { counter-increment: a }
//         ul { counter-reset: b }
//         li {
//           counter-increment: b;
//           string-set {
//             textHeader content(before) "-" content() "-" content(after)
//                         counter(a, upper-roman) "." counters(b, "|"),
//             textFooter content(before) "-" attr(class)
//                         counters(b, "|") "/" counter(a, upper-roman);
//           }
//         }
//         li:before { content: "before!" }
//         li:after { content: "after!" }
//         li:last-child:before { content: "before!last" }
//         li:last-child:after { content: "after!last" }
//       </style>
//       <ul>
//         <li class="firstclass">first
//         <li>
//           <ul>
//             <li class="secondclass">second
//     `)

//     html, topCenter, bottomCenter = page1.Children
//     topLineBox, = topCenter.Children
//     topTextBox, = topLineBox.Children
//     assert topTextBox.text == "before!-first-after!I.1"
//     bottomLineBox, = bottomCenter.Children
//     bottomTextBox, = bottomLineBox.Children
//     assert bottomTextBox.text == "before!last-secondclass2|1/I"

// func TestMarginBoxStringSet7(t *testing.T){
//     // Test regression: https://github.com/Kozea/WeasyPrint/issues/722
//     page1, = renderPages(`
//       <style>
//         img { string-set: left attr(alt) }
//         img + img { string-set: right attr(alt) }
//         @page { @top-left  { content: "[" string(left)  "]" }
//                 @top-right { content: "{" string(right) "}" } }
//       </style>
//       <img src=pattern.png alt="Chocolate">
//       <img src=noSuchFile.png alt="Cake">
//     `)
// }
//     html, topLeft, topRight = page1.Children
//     leftLineBox, = topLeft.Children
//     leftTextBox, = leftLineBox.Children
//     assert leftTextBox.text == "[Chocolate]"
//     rightLineBox, = topRight.Children
//     rightTextBox, = rightLineBox.Children
//     assert rightTextBox.text == "{Cake}"

// func TestMarginBoxStringSet8(t *testing.T) {
//   cp := testutils.CaptureLogs()
//   defer cp.AssertNoLogs(t)

//     // Test regression: https://github.com/Kozea/WeasyPrint/issues/726
//     page1, page2, page3 = renderPages(`
//       <style>
//         @page { @top-left  { content: "[" string(left) "]" } }
//         p { page-break-before: always }
//         .initial { string-set: left "initial" }
//         .empty   { string-set: left ""        }
//         .space   { string-set: left " "       }
//       </style>
// }
//       <p class="initial">Initial</p>
//       <p class="empty">Empty</p>
//       <p class="space">Space</p>
//     `)
//     html, topLeft = page1.Children
//     leftLineBox, = topLeft.Children
//     leftTextBox, = leftLineBox.Children
//     assert leftTextBox.text == "[initial]"

//     html, topLeft = page2.Children
//     leftLineBox, = topLeft.Children
//     leftTextBox, = leftLineBox.Children
//     assert leftTextBox.text == "[]"

//     html, topLeft = page3.Children
//     leftLineBox, = topLeft.Children
//     leftTextBox, = leftLineBox.Children
//     assert leftTextBox.text == "[ ]"

// func TestMarginBoxStringSet9(t *testing.T){
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
//     // Test that named strings are case-sensitive
//     // See https://github.com/Kozea/WeasyPrint/pull/827
//     page1, = renderPages(`
//       <style>
//         @page {
//           @top-center {
//             content: string(textHeader, first)
//                      " " string(TEXTHeader, first)
//           }
//         }
//         p { string-set: textHeader content() }
//         div { string-set: TEXTHeader content() }
//       </style>
//       <p>first assignment</p>
//       <div>second assignment</div>
//     `)

//     html, topCenter = page1.Children
//     lineBox, = topCenter.Children
//     textBox, = lineBox.Children
//     assert textBox.text == "first assignment second assignment"

// @assertNoLogs
// // Test page-based counters.
// func TestPageCounters(t *testing.T) {
//     pages = renderPages(`
//       <style>
//         @page {
//           /* Make the page content area only 10px high && wide,
//              so every word := range <p> end up on a page of its own. */
//           size: 30px;
//           margin: 10px;
//           @bottom-center {
//             content: "Page " counter(page) " of " counter(pages) ".";
//           }
//         }
//       </style>
//       <p>lorem ipsum dolor
//     `)
//     for pageNumber, page := range enumerate(pages, 1):
//         html, bottomCenter = page.Children
//         lineBox, = bottomCenter.Children
//         textBox, = lineBox.Children
//         assert textBox.text == "Page {0} of 3.".format(pageNumber)
