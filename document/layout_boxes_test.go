package document

import (
	"fmt"
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

// same as renderPages, but expects only on laid out page
func renderOnePage(t *testing.T, htmlContent string) *bo.PageBox {
	pages := renderPages(t, htmlContent)
	if len(pages) != 1 {
		t.Fatalf("expected one page, got %v", pages)
	}
	return pages[0]
}

func TestMarginBoxes(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	pages := renderPages(t, `
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
	  `)
	if len(pages) != 2 {
		t.Fatalf("expected 2 pages, got %v", pages)
	}
	page1, page2 := pages[0], pages[1]
	if tag := page1.Children[0].Box().ElementTag; tag != "html" {
		t.Fatalf("expected root = html, got %s", tag)
	}
	if tag := page2.Children[0].Box().ElementTag; tag != "html" {
		t.Fatalf("expected root = html, got %s", tag)
	}

	var marginBoxes1, marginBoxes2 []string
	for _, box := range page1.Children[1:] {
		marginBoxes1 = append(marginBoxes1, box.(*bo.MarginBox).AtKeyword)
	}
	for _, box := range page2.Children[1:] {
		marginBoxes2 = append(marginBoxes2, box.(*bo.MarginBox).AtKeyword)
	}
	if exp := []string{"@top-center", "@bottom-left", "@bottom-left-corner"}; !reflect.DeepEqual(marginBoxes1, exp) {
		t.Fatalf("expected %v, got %v", exp, marginBoxes1)
	}
	if exp := []string{"@top-center"}; !reflect.DeepEqual(marginBoxes2, exp) {
		t.Fatalf("expected %v, got %v", exp, marginBoxes2)
	}

	if len(page2.Children) != 2 {
		t.Fatalf("expected two children, got %v", page2.Children)
	}
	_, topCenter := page2.Children[0], page2.Children[1]
	lineBox := topCenter.Box().Children[0]
	textBox, _ := lineBox.Box().Children[0].(*bo.TextBox)
	if textBox.Text != "Title" {
		t.Fatalf("expected title, got %s", textBox.Text)
	}
}

func TestMarginBoxStringSet1(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Test that both pages get string in the `bottom-center` margin box
	pages := renderPages(t, `
      <style>
        @page {
          @bottom-center { content: string(text_header) }
        }
        p {
          string-set: text_header content();
        }
        .page {
          page-break-before: always;
        }
      </style>
      <p>first assignment</p>
      <div class="page"></div>
    `)
	if len(pages) != 2 {
		t.Fatalf("expected 2 pages, got %v", pages)
	}
	page1, page2 := pages[0], pages[1]

	if len(page2.Children) != 2 {
		t.Fatalf("expected two children, got %v", page2.Children)
	}
	_, bottomCenter := page2.Children[0], page2.Children[1]
	lineBox := bottomCenter.Box().Children[0]
	textBox, _ := lineBox.Box().Children[0].(*bo.TextBox)
	if textBox.Text != "first assignment" {
		t.Fatalf("expected 'first assignment', got %s", textBox.Text)
	}

	if len(page1.Children) != 2 {
		t.Fatalf("expected two children, got %v", page1.Children)
	}
	_, bottomCenter = page1.Children[0], page1.Children[1]

	lineBox = bottomCenter.Box().Children[0]
	textBox, _ = lineBox.Box().Children[0].(*bo.TextBox)
	if textBox.Text != "first assignment" {
		t.Fatalf("expected 'first assignment', got %s", textBox.Text)
	}
}

func TestMarginBoxStringSet2(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	simpleStringSetTest := func(contentVal, extraStyle string) {
		page1 := renderOnePage(t, fmt.Sprintf(`
          <style>
            @page {
              @top-center { content: string(text_header) }
            }
            p {
              string-set: text_header content(%s);
            }
            %s
          </style>
          <p>first assignment</p>
        `, contentVal, extraStyle))

		topCenter := page1.Children[1]
		lineBox := topCenter.Box().Children[0]
		textBox := lineBox.Box().Children[0].(*bo.TextBox)
		if contentVal == "before" || contentVal == "after" {
			if textBox.Text != "pseudo" {
				t.Fatalf("expected 'pseudo', got %s", textBox.Text)
			}
		} else {
			if textBox.Text != "first assignment" {
				t.Fatalf("expected 'first assignment', got %s", textBox.Text)
			}
		}
	}

	// Test each accepted value of `content()` as an arguemnt to `string-set`
	for _, value := range []string{"", "text", "before", "after"} {
		var extraStyle string
		if value == "before" || value == "after" {
			extraStyle = fmt.Sprintf("p:%s{content: 'pseudo'}", value)
		}
		simpleStringSetTest(value, extraStyle)
	}
}

func TestMarginBoxStringSet3(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)
	// Test `first` (default value) ie. use the first assignment on the page
	page1 := renderOnePage(t, `
      <style>
        @page {
          @top-center { content: string(text_header, first) }
        }
        p {
          string-set: text_header content();
        }
      </style>
      <p>first assignment</p>
      <p>Second assignment</p>
    } `)

	topCenter := page1.Children[1]
	lineBox := topCenter.Box().Children[0]
	textBox := lineBox.Box().Children[0].(*bo.TextBox)
	if textBox.Text != "first assignment" {
		t.Fatalf("expected 'first assignment', got %s", textBox.Text)
	}
}

func TestMarginBoxStringSet4(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// test `first-except` ie. exclude from page on which value is assigned
	pages := renderPages(t, `
		<style>
		@page {
		@top-center { content: string(header-nofirst, first-except) }
		}
		p{
		string-set: header-nofirst content();
		}
		.page{
		page-break-before: always;
		}
	</style>
	<p>first_excepted</p>
	<div class="page"></div>
	`)
	if len(pages) != 2 {
		t.Fatalf("expected 2 pages, got %v", pages)
	}
	page1, page2 := pages[0], pages[1]

	topCenter := page1.Box().Children[1]
	if L := len(topCenter.Box().Children); L != 0 {
		t.Fatalf("expected no child, got %d", L)
	}

	topCenter = page2.Box().Children[1]
	lineBox := topCenter.Box().Children[0]
	textBox := lineBox.Box().Children[0].(*bo.TextBox)
	if textBox.Text != "first_excepted" {
		t.Fatalf("expected 'first_excepted', got %s", textBox.Text)
	}
}

func TestMarginBoxStringSet5(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)
	// Test `last` ie. use the most-recent assignment
	page1 := renderOnePage(t, `
      <style>
        @page {
          @top-center { content: string(headerLast, last) }
        }
        p {
          string-set: headerLast content();
        }
      </style>
      <p>String set</p>
      <p>Second assignment</p>
    `)

	topCenter := page1.Children[1]
	lineBox := topCenter.Box().Children[0]
	textBox := lineBox.Box().Children[0].(*bo.TextBox)
	if textBox.Text != "Second assignment" {
		t.Fatalf("expected 'Second assignment', got %s", textBox.Text)
	}
}

func TestMarginBoxStringSet6(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Test multiple complex string-set values
	page1 := renderOnePage(t, `
		<style>
		@page {
		@top-center { content: string(text_header, first) }
		@bottom-center { content: string(text_footer, last) }
		}
		html { counter-reset: a }
		body { counter-increment: a }
		ul { counter-reset: b }
		li {
		counter-increment: b;
		string-set:
			text_header content(before) "-" content() "-" content(after)
						counter(a, upper-roman) '.' counters(b, '|'),
			text_footer content(before) '-' attr(class)
						counters(b, '|') "/" counter(a, upper-roman);
		}
		li:before { content: 'before!' }
		li:after { content: 'after!' }
		li:last-child:before { content: 'before!last' }
		li:last-child:after { content: 'after!last' }
	</style>
	<ul>
		<li class="firstclass">first
		<li>
		<ul>
			<li class="secondclass">second
    `)

	topCenter, bottomCenter := page1.Children[1], page1.Children[2]
	topLineBox := topCenter.Box().Children[0]
	topTextBox := topLineBox.Box().Children[0].(*bo.TextBox)
	if topTextBox.Text != "before!-first-after!I.1" {
		t.Fatalf("expected 'before!-first-after!I.1', got %s", topTextBox.Text)
	}

	bottomLineBox := bottomCenter.Box().Children[0]
	bottomTextBox := bottomLineBox.Box().Children[0].(*bo.TextBox)
	if bottomTextBox.Text != "before!last-secondclass2|1/I" {
		t.Fatalf("expected 'before!last-secondclass2|1/I', got %s", bottomTextBox.Text)
	}
}

func TestMarginBoxStringSet7(t *testing.T) {
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/722
	page1 := renderOnePage(t, `
      <style>
        img { string-set: left attr(alt) }
        img + img { string-set: right attr(alt) }
        @page { @top-left  { content: "[" string(left)  "]" }
                @top-right { content: "{" string(right) "}" } }
      </style>
      <img src=pattern.png alt="Chocolate">
      <img src=noSuchFile.png alt="Cake">
    `)

	topLeft, topRight := page1.Children[1], page1.Children[2]
	leftLineBox := topLeft.Box().Children[0]
	leftTextBox := leftLineBox.Box().Children[0].(*bo.TextBox)
	if leftTextBox.Text != "[Chocolate]" {
		t.Fatalf("expected '[Chocolate]', got %s", leftTextBox.Text)
	}

	rightLineBox := topRight.Box().Children[0]
	rightTextBox := rightLineBox.Box().Children[0].(*bo.TextBox)
	if rightTextBox.Text != "{Cake}" {
		t.Fatalf("expected '{Cake}', got %s", rightTextBox.Text)
	}
}

func TestMarginBoxStringSet8(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Test regression: https://github.com/Kozea/WeasyPrint/issues/726
	pages := renderPages(t, `
      <style>
        @page { @top-left  { content: "[" string(left) "]" } }
        p { page-break-before: always }
        .initial { string-set: left "initial" }
        .empty   { string-set: left ""        }
        .space   { string-set: left " "       }
      </style>

	  <p class="initial">Initial</p>
      <p class="empty">Empty</p>
      <p class="space">Space</p>
    `)
	if len(pages) != 3 {
		t.Fatalf("expected 3 page, got %v", pages)
	}
	page1, page2, page3 := pages[0], pages[1], pages[2]

	topLeft := page1.Box().Children[1]
	leftLineBox := topLeft.Box().Children[0]
	leftTextBox := leftLineBox.Box().Children[0].(*bo.TextBox)
	if leftTextBox.Text != "[initial]" {
		t.Fatalf("expected '[initial]', got %s", leftTextBox.Text)
	}

	topLeft = page2.Box().Children[1]
	leftLineBox = topLeft.Box().Children[0]
	leftTextBox = leftLineBox.Box().Children[0].(*bo.TextBox)
	if leftTextBox.Text != "[]" {
		t.Fatalf("expected '[]', got %s", leftTextBox.Text)
	}

	topLeft = page3.Box().Children[1]
	leftLineBox = topLeft.Box().Children[0]
	leftTextBox = leftLineBox.Box().Children[0].(*bo.TextBox)
	if leftTextBox.Text != "[ ]" {
		t.Fatalf("expected '[ ]', got %s", leftTextBox.Text)
	}
}

func TestMarginBoxStringSet9(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)
	// Test that named strings are case-sensitive
	// See https://github.com/Kozea/WeasyPrint/pull/827
	page1 := renderOnePage(t, `
      <style>
        @page {
          @top-center {
            content: string(text_header, first)
                     " " string(TEXTHeader, first)
          }
        }
        p { string-set: text_header content() }
        div { string-set: TEXTHeader content() }
      </style>
      <p>first assignment</p>
      <div>second assignment</div>
    `)

	topCenter := page1.Children[1]
	lineBox := topCenter.Box().Children[0]
	textBox := lineBox.Box().Children[0].(*bo.TextBox)

	if textBox.Text != "first assignment second assignment" {
		t.Fatalf("expected 'first assignment second assignment', got %s", textBox.Text)
	}
}

// Test page-based counters.
func TestPageCounters(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page {
          /* Make the page content area only 10px high and wide,
             so every word in <p> end up on a page of its own. */
          size: 30px;
          margin: 10px;
          @bottom-center {
            content: "Page " counter(page) " of " counter(pages) ".";
          }
        }
      </style>
      <p>lorem ipsum dolor
    `)
	for pageIndex, page := range pages {
		pageNumber := pageIndex + 1
		bottomCenter := page.Box().Children[1]
		lineBox := bottomCenter.Box().Children[0]
		textBox := lineBox.Box().Children[0].(*bo.TextBox)
		exp := fmt.Sprintf("Page %d of 3.", pageNumber)
		if textBox.Text != exp {
			t.Fatalf("expected '%s', got %s", exp, textBox.Text)
		}
	}
}
