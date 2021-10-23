package layout

import (
	"fmt"
	"strings"
	"testing"

	bo "github.com/benoitkugler/go-weasyprint/boxes"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	tu "github.com/benoitkugler/go-weasyprint/utils/testutils"
)

// Tests for pages layout.

// Test the layout for ``@page`` properties.
func TestPageSizeBasic(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	for _, data := range []struct {
		size          string
		width, height int
	}{
		{"auto", 793, 1122},
		{"2in 10in", 192, 960},
		{"242px", 242, 242},
		{"letter", 816, 1056},
		{"letter portrait", 816, 1056},
		{"letter landscape", 1056, 816},
		{"portrait", 793, 1122},
		{"landscape", 1122, 793},
	} {
		page := renderOnePage(t, fmt.Sprintf("<style>@page { size: %s; }</style>", data.size))
		tu.AssertEqual(t, int(page.Box().MarginWidth()), data.width, "int")
		tu.AssertEqual(t, int(page.Box().MarginHeight()), data.height, "int")
	}
}

func TestPageSizeWithMargin(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `<style>
      @page { size: 200px 300px; margin: 10px 10% 20% 1in }
      body { margin: 8px }
    </style>
    <p style="margin: 0">`)
	tu.AssertEqual(t, page.Box().MarginWidth(), pr.Float(200), "page")
	tu.AssertEqual(t, page.Box().MarginHeight(), pr.Float(300), "page")
	tu.AssertEqual(t, page.Box().PositionX, pr.Float(0), "page")
	tu.AssertEqual(t, page.Box().PositionY, pr.Float(0), "page")
	tu.AssertEqual(t, page.Box().Width, pr.Float(84), "page")   // 200px - 10% - 1 inch
	tu.AssertEqual(t, page.Box().Height, pr.Float(230), "page") // 300px - 10px - 20%

	html := page.Box().Children[0]
	tu.AssertEqual(t, html.Box().ElementTag, "html", "html")
	tu.AssertEqual(t, html.Box().PositionX, pr.Float(96), "html") // 1in
	tu.AssertEqual(t, html.Box().PositionY, pr.Float(10), "html") // root element’s margins do ! collapse
	tu.AssertEqual(t, html.Box().Width, pr.Float(84), "html")

	body := html.Box().Children[0]
	tu.AssertEqual(t, body.Box().ElementTag, "body", "body")
	tu.AssertEqual(t, body.Box().PositionX, pr.Float(96), "body") // 1in
	tu.AssertEqual(t, body.Box().PositionY, pr.Float(10), "body")
	// body has margins in the UA stylesheet
	tu.AssertEqual(t, body.Box().MarginLeft, pr.Float(8), "body")
	tu.AssertEqual(t, body.Box().MarginRight, pr.Float(8), "body")
	tu.AssertEqual(t, body.Box().MarginTop, pr.Float(8), "body")
	tu.AssertEqual(t, body.Box().MarginBottom, pr.Float(8), "body")
	tu.AssertEqual(t, body.Box().Width, pr.Float(68), "body")

	paragraph := body.Box().Children[0]
	tu.AssertEqual(t, paragraph.Box().ElementTag, "p", "paragraph")
	tu.AssertEqual(t, paragraph.Box().PositionX, pr.Float(104), "paragraph") // 1in + 8px
	tu.AssertEqual(t, paragraph.Box().PositionY, pr.Float(18), "paragraph")  // 10px + 8px
	tu.AssertEqual(t, paragraph.Box().Width, pr.Float(68), "paragraph")
}

func TestPageSizeWithMarginBorderPadding(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `<style> @page {
      size: 100px; margin: 1px 2px; padding: 4px 8px;
      border-width: 16px 32px; border-style: solid;
    }</style>`)
	tu.AssertEqual(t, page.Box().Width, pr.Float(16), "page")  // 100 - 2 * 42
	tu.AssertEqual(t, page.Box().Height, pr.Float(58), "page") // 100 - 2 * 21
	html := page.Box().Children[0]
	tu.AssertEqual(t, html.Box().ElementTag, "html", "html")
	tu.AssertEqual(t, html.Box().PositionX, pr.Float(42), "html") // 2 + 8 + 32
	tu.AssertEqual(t, html.Box().PositionY, pr.Float(21), "html") // 1 + 4 + 16
}

func TestPageSizeMargins(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, data := range []struct {
		margin                   string
		top, right, bottom, left pr.Float
	}{
		{"auto", 15, 10, 15, 10},
		{"5px 5px auto auto", 5, 5, 25, 15},
	} {
		page := renderOnePage(t, fmt.Sprintf(`<style>@page {
      size: 106px 206px; width: 80px; height: 170px;
      padding: 1px; border: 2px solid; margin: %s }</style>`, data.margin))
		tu.AssertEqual(t, page.Box().MarginTop, data.top, "page.MarginTop")
		tu.AssertEqual(t, page.Box().MarginRight, data.right, "page.MarginRight")
		tu.AssertEqual(t, page.Box().MarginBottom, data.bottom, "page.MarginBottom")
		tu.AssertEqual(t, page.Box().MarginLeft, data.left, "page.MarginLeft")
	}
}

func TestPageSizeOverConstrained(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, data := range []struct {
		style         string
		width, height pr.Float
	}{
		{
			"size: 4px 10000px; width: 100px; height: 100px;" +
				"padding: 1px; border: 2px solid; margin: 3px",
			112, 112,
		},
		{
			"size: 1000px; margin: 100px; max-width: 500px; min-height: 1500px",
			700, 1700,
		},
		{
			"size: 1000px; margin: 100px; min-width: 1500px; max-height: 500px",
			1700, 700,
		},
	} {
		page := renderOnePage(t, fmt.Sprintf("<style>@page { %s }</style>", data.style))
		tu.AssertEqual(t, page.Box().MarginWidth(), data.width, "page")
		tu.AssertEqual(t, page.Box().MarginHeight(), data.height, "page")
	}
}

func TestPageBreaks(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, html := range []string{
		"<div>1</div>",
		"<div></div>",
		"<img src=pattern.png>",
	} {
		pages := renderPages(t, fmt.Sprintf(`
      <style>
        @page { size: 100px; margin: 10px }
        body { margin: 0 }
        div { height: 30px; font-size: 20px }
        img { height: 30px; display: block }
      </style>
      %s`, strings.Repeat(html, 5)))
		var posY [][]pr.Float
		for _, page := range pages {
			html := page.Box().Children[0]
			body := html.Box().Children[0]
			children := body.Box().Children
			var pos []pr.Float
			for _, child := range children {
				tu.AssertEqual(t, child.Box().ElementTag == "div" || child.Box().ElementTag == "img", true, "tag")
				tu.AssertEqual(t, child.Box().PositionX, pr.Float(10), "positionX")
				pos = append(pos, child.Box().PositionY)
			}
			posY = append(posY, pos)
		}
		tu.AssertEqual(t, posY, [][]pr.Float{{10, 40}, {10, 40}, {10}}, "positionY")
	}
}

func TestPageBreaksComplex1(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { margin: 10px }
        @page :left { margin-left: 50px }
        @page :right { margin-right: 50px }
        html { page-break-before: left }
        div { page-break-after: left }
        ul { page-break-before: always }
      </style>
      <div>1</div>
      <p>2</p>
      <p>3</p>
      <article>
        <section>
          <ul><li>4</li></ul>
        </section>
      </article>
    `)
	page1, page2, page3, page4 := pages[0], pages[1], pages[2], pages[3]

	// The first page is a right page on rtl, but not here because of
	// page-break-before on the root element.
	tu.AssertEqual(t, page1.Box().MarginLeft, pr.Float(50), "page1") // left page
	tu.AssertEqual(t, page1.Box().MarginRight, pr.Float(10), "page1")
	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	line := div.Box().Children[0]
	text := line.Box().Children[0]
	tu.AssertEqual(t, div.Box().ElementTag, "div", "div")
	tu.AssertEqual(t, text.(*bo.TextBox).Text, "1", "text")

	html = page2.Box().Children[0]
	tu.AssertEqual(t, page2.Box().MarginLeft, pr.Float(10), "page2")
	tu.AssertEqual(t, page2.Box().MarginRight, pr.Float(50), "page2") // right page
	tu.AssertEqual(t, len(html.Box().Children), 0, "empty page")      // empty page to get to a left page

	tu.AssertEqual(t, page3.Box().MarginLeft, pr.Float(50), "page3") // left page
	tu.AssertEqual(t, page3.Box().MarginRight, pr.Float(10), "page3")
	html = page3.Box().Children[0]
	body = html.Box().Children[0]
	p1, p2 := unpack2(body)
	tu.AssertEqual(t, p1.Box().ElementTag, "p", "p1")
	tu.AssertEqual(t, p2.Box().ElementTag, "p", "p2")

	tu.AssertEqual(t, page4.Box().MarginLeft, pr.Float(10), "page4")
	tu.AssertEqual(t, page4.Box().MarginRight, pr.Float(50), "page4") // right page
	html = page4.Box().Children[0]
	body = html.Box().Children[0]
	article := body.Box().Children[0]
	section := article.Box().Children[0]
	ulist := section.Box().Children[0]
	tu.AssertEqual(t, ulist.Box().ElementTag, "ul", "ulist")
}

func TestPageBreaksComplex2(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Reference for the following test:
	// Without any "avoid", this breaks after the <div>
	pages := renderPages(t, `
      <style>
        @page { size: 140px; margin: 0 }
        img { height: 25px; vertical-align: top }
        p { orphans: 1; widows: 1 }
      </style>
      <img src=pattern.png>
      <div>
        <p><img src=pattern.png><br/><img src=pattern.png><p>
        <p><img src=pattern.png><br/><img src=pattern.png><p>
      </div><!-- page break here -->
      <img src=pattern.png>
    `)
	page1, page2 := pages[0], pages[1]

	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	img1, div := unpack2(body)
	tu.AssertEqual(t, img1.Box().PositionY, pr.Float(0), "img1")
	tu.AssertEqual(t, img1.Box().Height, pr.Float(25), "img1")
	tu.AssertEqual(t, div.Box().PositionY, pr.Float(25), "div")
	tu.AssertEqual(t, div.Box().Height, pr.Float(100), "div")

	html = page2.Box().Children[0]
	body = html.Box().Children[0]
	img2 := body.Box().Children[0]
	tu.AssertEqual(t, img2.Box().PositionY, pr.Float(0), "img2")
	tu.AssertEqual(t, img2.Box().Height, pr.Float(25), "img2")
}

func TestPageBreaksComplex3(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Adding a few page-break-*: avoid, the only legal break is
	// before the <div>
	pages := renderPages(t, `
      <style>
        @page { size: 140px; margin: 0 }
        img { height: 25px; vertical-align: top }
        p { orphans: 1; widows: 1 }
      </style>
      <img src=pattern.png><!-- page break here -->
      <div>
        <p style="page-break-inside: avoid">
          <img src=pattern.png><br/><img src=pattern.png></p>
        <p style="page-break-before: avoid; page-break-after: avoid; widows: 2"
          ><img src=pattern.png><br/><img src=pattern.png></p>
      </div>
      <img src=pattern.png>
    `)
	page1, page2 := pages[0], pages[1]

	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	img1 := body.Box().Children[0]
	tu.AssertEqual(t, img1.Box().PositionY, pr.Float(0), "img1")
	tu.AssertEqual(t, img1.Box().Height, pr.Float(25), "img1")

	html = page2.Box().Children[0]
	body = html.Box().Children[0]
	div, img2 := unpack2(body)
	tu.AssertEqual(t, div.Box().PositionY, pr.Float(0), "div")
	tu.AssertEqual(t, div.Box().Height, pr.Float(100), "div")
	tu.AssertEqual(t, img2.Box().PositionY, pr.Float(100), "img2")
	tu.AssertEqual(t, img2.Box().Height, pr.Float(25), "img2")
}

func TestPageBreaksComplex4(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @page { size: 140px; margin: 0 }
        img { height: 25px; vertical-align: top }
        p { orphans: 1; widows: 1 }
      </style>
      <img src=pattern.png><!-- page break here -->
      <div>
        <div>
          <p style="page-break-inside: avoid">
            <img src=pattern.png><br/><img src=pattern.png></p>
          <p style="page-break-before:avoid; page-break-after:avoid; widows:2"
            ><img src=pattern.png><br/><img src=pattern.png></p>
        </div>
        <img src=pattern.png>
      </div>
    `)
	page1, page2 := pages[0], pages[1]

	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	img1 := body.Box().Children[0]
	tu.AssertEqual(t, img1.Box().PositionY, pr.Float(0), "img1")
	tu.AssertEqual(t, img1.Box().Height, pr.Float(25), "img1")

	html = page2.Box().Children[0]
	body = html.Box().Children[0]
	outerDiv := body.Box().Children[0]
	innerDiv, img2 := unpack2(outerDiv)
	tu.AssertEqual(t, innerDiv.Box().PositionY, pr.Float(0), "innerDiv")
	tu.AssertEqual(t, innerDiv.Box().Height, pr.Float(100), "innerDiv")
	tu.AssertEqual(t, img2.Box().PositionY, pr.Float(100), "img2")
	tu.AssertEqual(t, img2.Box().Height, pr.Float(25), "img2")
}

// TODO:

// func TestPageBreaksComplex5(t *testing.T) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     // Reference for the next test
//     page1, page2, page3 = renderPages(`
//       <style>
//         @page { size: 100px; margin: 0 }
//         img { height: 30px; display: block; }
//         p { orphans: 1; widows: 1 }
//       </style>
//       <div>
//         <img src=pattern.png style="page-break-after: always">
//         <section>
//           <img src=pattern.png>
//           <img src=pattern.png>
//         </section>
//       </div>
//       <img src=pattern.png><!-- page break here -->
//       <img src=pattern.png>
//     `)
//     html := page1.Box().Children[0]
//     body := html.Box().Children[0]
//     div := body.Box().Children[0]
//     tu.AssertEqual(t, div.Box().Height , 100, "div")
//     html := page2.Box().Children[0]
//     body := html.Box().Children[0]
//     div, img4 = body.Box().Children
//     tu.AssertEqual(t, div.Box().Height , 60, "div")
//     tu.AssertEqual(t, img4.Box().Height , 30, "img4")
//     html := page3.Box().Children[0]
//     body := html.Box().Children[0]
//     img5, = body.Box().Children
//     tu.AssertEqual(t, img5.Box().Height , 30, "img5")

// func TestPageBreaksComplex6(t *testing.T):
// capt := tu.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     page1, page2, page3 = renderPages(`
//       <style>
//         @page { size: 100px; margin: 0 }
//         img { height: 30px; display: block; }
//         p { orphans: 1; widows: 1 }
//       </style>
//       <div>
//         <img src=pattern.png style="page-break-after: always">
//         <section>
//           <img src=pattern.png><!-- page break here -->
//           <img src=pattern.png style="page-break-after: avoid">
//         </section>
//       </div>
//       <img src=pattern.png style="page-break-after: avoid">
//       <img src=pattern.png>
//     `)
//     html := page1.Box().Children[0]
//     body := html.Box().Children[0]
//     div := body.Box().Children[0]
//     tu.AssertEqual(t, div.Box().Height , 100, "div")
//     html := page2.Box().Children[0]
//     body := html.Box().Children[0]
//     div := body.Box().Children[0]
//     section, = div.Box().Children
//     img2, = section.Box().Children
//     tu.AssertEqual(t, img2.Box().Height , 30, "img2")
//     // TODO: currently this is 60: we do ! increase the used height of blocks
//     // to make them fill the blank space at the end of the age when we remove
//     // children from them for some break-*: avoid.
//     // See TODOs := range blocks.blockContainerLayout
//     // tu.AssertEqual(t, div.Box().Height , 100, "div")
//     html := page3.Box().Children[0]
//     body := html.Box().Children[0]
//     div, img4, img5, = body.Box().Children
//     tu.AssertEqual(t, div.Box().Height , 30, "div")
//     tu.AssertEqual(t, img4.Box().Height , 30, "img4")
//     tu.AssertEqual(t, img5.Box().Height , 30, "img5")
// }

// func TestPageBreaksComplex7(t *testing.T) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     page1, page2, page3 = renderPages(`
//       <style>
//         @page { @bottom-center { content: counter(page) } }
//         @page:blank { @bottom-center { content: none } }
//       </style>
//       <p style="page-break-after: right">foo</p>
//       <p>bar</p>
//     `)
//     tu.AssertEqual(t, len(page1.Box().Children) , 2  // content && @bottom-center, "len")
//     tu.AssertEqual(t, len(page2.Box().Children) , 1  // content only, "len")
//     tu.AssertEqual(t, len(page3.Box().Children) , 2  // content && @bottom-center, "len")

// func TestPageBreaksComplex8(t *testing.T):
// capt := tu.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     page1, page2 = renderPages(`
//       <style>
//         @page { size: 75px; margin: 0 }
//         div { height: 20px }
//       </style>
//       <div></div>
//       <section>
//         <div></div>
//         <div style="page-break-after: avoid">
//           <div style="position: absolute"></div>
//           <div style="position: fixed"></div>
//         </div>
//       </section>
//       <div></div>
//     `)
//     html := page1.Box().Children[0]
//     body, Div = html.Box().Children
//     div1, section = body.Box().Children
//     div2, = section.Box().Children
//     tu.AssertEqual(t, div1.Box().PositionY , 0, "div1")
//     tu.AssertEqual(t, div2.Box().PositionY , 20, "div2")
//     tu.AssertEqual(t, div1.Box().Height , 20, "div1")
//     tu.AssertEqual(t, div2.Box().Height , 20, "div2")
//     html := page2.Box().Children[0]
//     body := html.Box().Children[0]
//     section, div4 = body.Box().Children
//     div3, = section.Box().Children
//     absolute, fixed = div3.Box().Children
//     tu.AssertEqual(t, div3.Box().PositionY , 0, "div3")
//     tu.AssertEqual(t, div4.Box().PositionY , 20, "div4")
//     tu.AssertEqual(t, div3.Box().Height , 20, "div3")
//     tu.AssertEqual(t, div4.Box().Height , 20, "div4")
// }

// @pytest.mark.parametrize("breakAfter, marginBreak, marginTop", (
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     ("page", "auto", 5),
//     ("auto", "auto", 0),
//     ("page", "keep", 5),
//     ("auto", "keep", 5),
//     ("page", "discard", 0),
//     ("auto", "discard", 0),
// ))
// func TestMarginBreak(t *testing.TbreakAfter, marginBreak, marginTop) {
//     page1, page2 = renderPages(`
//       <style>
//         @page { size: 70px; margin: 0 }
//         div { height: 63px; margin: 5px 0 8px;
//               break-after: %s; margin-break: %s }
//       </style>
//       <section>
//         <div></div>
//       </section>
//       <section>
//         <div></div>
//       </section>
//     ` % (breakAfter, marginBreak))
//     html := page1.Box().Children[0]
//     body := html.Box().Children[0]
//     section, = body.Box().Children
//     div := section.Box().Children[0]
//     tu.AssertEqual(t, div.Box().MarginTop , 5, "div")

//     html := page2.Box().Children[0]
//     body := html.Box().Children[0]
//     section, = body.Box().Children
//     div := section.Box().Children[0]
//     tu.AssertEqual(t, div.Box().MarginTop , marginTop, "div")

// @pytest.mark.xfail

// func TestMarginBreakClearance(t *testing.T):
// capt := tu.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     page1, page2 = renderPages(`
//       <style>
//         @page { size: 70px; margin: 0 }
//         div { height: 63px; margin: 5px 0 8px; break-after: page }
//       </style>
//       <section>
//         <div></div>
//       </section>
//       <section>
//         <div style="border-top: 1px solid black">
//           <div></div>
//         </div>
//       </section>
//     `)
//     html := page1.Box().Children[0]
//     body := html.Box().Children[0]
//     section, = body.Box().Children
//     div := section.Box().Children[0]
//     tu.AssertEqual(t, div.Box().MarginTop , 5, "div")
// }
//     html := page2.Box().Children[0]
//     body := html.Box().Children[0]
//     section, = body.Box().Children
//     div1, = section.Box().Children
//     tu.AssertEqual(t, div1.Box().MarginTop , 0, "div1")
//     div2, = div1.Box().Children
//     tu.AssertEqual(t, div2.Box().MarginTop , 5, "div2")
//     tu.AssertEqual(t, div2.contentBoxY() , 5, "div2")

// @pytest.mark.parametrize("direction, pageBreak, pagesNumber", (
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     ("ltr", "recto", 3),
//     ("ltr", "verso", 2),
//     ("rtl", "recto", 3),
//     ("rtl", "verso", 2),
//     ("ltr", "right", 3),
//     ("ltr", "left", 2),
//     ("rtl", "right", 2),
//     ("rtl", "left", 3),
// ))
// func TestRectoVersoBreak(t *testing.Tdirection, pageBreak, pagesNumber) {
//     pages = renderPages(`
//       <style>
//         html { direction: %s }
//         p { break-before: %s }
//       </style>
//       abc
//       <p>def</p>
//     ` % (direction, pageBreak))
//     tu.AssertEqual(t, len(pages) , pagesNumber, "len")

// func TestPageNames1(t *testing.T):
// capt := tu.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     pages = renderPages(`
//       <style>
//         @page { size: 100px 100px }
//         section { page: small }
//       </style>
//       <div>
//         <section>large</section>
//       </div>
//     `)
//     page1, = pages
//     tu.AssertEqual(t, (page1.Box().Width, page1.Box().Height) , (100, 100), "(page1")
// }

// func TestPageNames2(t *testing.T) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     pages = renderPages(`
//       <style>
//         @page { size: 100px 100px }
//         @page narrow { margin: 1px }
//         section { page: small }
//       </style>
//       <div>
//         <section>large</section>
//       </div>
//     `)
//     page1, = pages
//     tu.AssertEqual(t, (page1.Box().Width, page1.Box().Height) , (100, 100), "(page1")

// func TestPageNames3(t *testing.T):
// capt := tu.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     pages = renderPages(`
//       <style>
//         @page { margin: 0 }
//         @page narrow { size: 100px 200px }
//         @page large { size: 200px 100px }
//         div { page: narrow }
//         section { page: large }
//       </style>
//       <div>
//         <section>large</section>
//         <section>large</section>
//         <p>narrow</p>
//       </div>
//     `)
//     page1, page2 = pages
// }
//     tu.AssertEqual(t, (page1.Box().Width, page1.Box().Height) , (200, 100), "(page1")
//     html := page1.Box().Children[0]
//     body := html.Box().Children[0]
//     div := body.Box().Children[0]
//     section1, section2 = div.Box().Children
//     tu.AssertEqual(t, section1.Box().ElementTag , section2.Box().ElementTag , "section", "section1")

//     tu.AssertEqual(t, (page2.Box().Width, page2.Box().Height) , (100, 200), "(page2")
//     html := page2.Box().Children[0]
//     body := html.Box().Children[0]
//     div := body.Box().Children[0]
//     p, = div.Box().Children
//     tu.AssertEqual(t, p.Box().ElementTag , "p", "p")

// func TestPageNames4(t *testing.T) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     pages = renderPages(`
//       <style>
//         @page { size: 200px 200px; margin: 0 }
//         @page small { size: 100px 100px }
//         p { page: small }
//       </style>
//       <section>normal</section>
//       <section>normal</section>
//       <p>small</p>
//       <section>small</section>
//     `)
//     page1, page2 = pages

//     tu.AssertEqual(t, (page1.Box().Width, page1.Box().Height) , (200, 200), "(page1")
//     html := page1.Box().Children[0]
//     body := html.Box().Children[0]
//     section1, section2 = body.Box().Children
//     tu.AssertEqual(t, section1.Box().ElementTag , section2.Box().ElementTag , "section", "section1")

//     tu.AssertEqual(t, (page2.Box().Width, page2.Box().Height) , (100, 100), "(page2")
//     html := page2.Box().Children[0]
//     body := html.Box().Children[0]
//     p, section = body.Box().Children
//     tu.AssertEqual(t, p.Box().ElementTag , "p", "p")
//     tu.AssertEqual(t, section.Box().ElementTag , "section", "section")

// func TestPageNames5(t *testing.T):
// capt := tu.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     pages = renderPages(`
//       <style>
//         @page { size: 200px 200px; margin: 0 }
//         @page small { size: 100px 100px }
//         div { page: small }
//       </style>
//       <section><p>a</p>b</section>
//       <section>c<div>d</div></section>
//     `)
//     page1, page2 = pages
// }
//     tu.AssertEqual(t, (page1.Box().Width, page1.Box().Height) , (200, 200), "(page1")
//     html := page1.Box().Children[0]
//     body := html.Box().Children[0]
//     section1, section2 = body.Box().Children
//     tu.AssertEqual(t, section1.Box().ElementTag , section2.Box().ElementTag , "section", "section1")
//     p, line = section1.Box().Children
//     line := section2.Box().Children[0]

//     tu.AssertEqual(t, (page2.Box().Width, page2.Box().Height) , (100, 100), "(page2")
//     html := page2.Box().Children[0]
//     body := html.Box().Children[0]
//     section2, = body.Box().Children
//     div := section2.Box().Children[0]

// func TestPageNames6(t *testing.T) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     pages = renderPages(`
//       <style>
//         @page { margin: 0 }
//         @page large { size: 200px 200px }
//         @page small { size: 100px 100px }
//         section { page: large }
//         div { page: small }
//       </style>
//       <section>a<p>b</p>c</section>
//       <section>d<div>e</div>f</section>
//     `)
//     page1, page2, page3 = pages

//     tu.AssertEqual(t, (page1.Box().Width, page1.Box().Height) , (200, 200), "(page1")
//     html := page1.Box().Children[0]
//     body := html.Box().Children[0]
//     section1, section2 = body.Box().Children
//     tu.AssertEqual(t, section1.Box().ElementTag , section2.Box().ElementTag , "section", "section1")
//     line1, p, line2 = section1.Box().Children
//     line := section2.Box().Children[0]

//     tu.AssertEqual(t, (page2.Box().Width, page2.Box().Height) , (100, 100), "(page2")
//     html := page2.Box().Children[0]
//     body := html.Box().Children[0]
//     section2, = body.Box().Children
//     div := section2.Box().Children[0]

//     tu.AssertEqual(t, (page3.Box().Width, page3.Box().Height) , (200, 200), "(page3")
//     html := page3.Box().Children[0]
//     body := html.Box().Children[0]
//     section2, = body.Box().Children
//     line := section2.Box().Children[0]

// func TestPageNames7(t *testing.T):
// capt := tu.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     pages = renderPages(`
//       <style>
//         @page { size: 200px 200px; margin: 0 }
//         @page small { size: 100px 100px }
//         p { page: small; break-before: right }
//       </style>
//       <section>normal</section>
//       <section>normal</section>
//       <p>small</p>
//       <section>small</section>
//     `)
//     page1, page2, page3 = pages
// }
//     tu.AssertEqual(t, (page1.Box().Width, page1.Box().Height) , (200, 200), "(page1")
//     html := page1.Box().Children[0]
//     body := html.Box().Children[0]
//     section1, section2 = body.Box().Children
//     tu.AssertEqual(t, section1.Box().ElementTag , section2.Box().ElementTag , "section", "section1")

//     tu.AssertEqual(t, (page2.Box().Width, page2.Box().Height) , (200, 200), "(page2")
//     html := page2.Box().Children[0]
//     tu.AssertEqual(t, ! html.Box().Children, "!")

//     tu.AssertEqual(t, (page3.Box().Width, page3.Box().Height) , (100, 100), "(page3")
//     html := page3.Box().Children[0]
//     body := html.Box().Children[0]
//     p, section = body.Box().Children
//     tu.AssertEqual(t, p.Box().ElementTag , "p", "p")
//     tu.AssertEqual(t, section.Box().ElementTag , "section", "section")

// func TestPageNames8(t *testing.T) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     pages = renderPages(`
//       <style>
//         @page small { size: 100px 100px }
//         section { page: small }
//         p { line-height: 80px }
//       </style>
//       <section>
//         <p>small</p>
//         <p>small</p>
//       </section>
//     `)
//     page1, page2 = pages

//     tu.AssertEqual(t, (page1.Box().Width, page1.Box().Height) , (100, 100), "(page1")
//     html := page1.Box().Children[0]
//     body := html.Box().Children[0]
//     section, = body.Box().Children
//     p, = section.Box().Children
//     tu.AssertEqual(t, section.Box().ElementTag , "section", "section")
//     tu.AssertEqual(t, p.Box().ElementTag , "p", "p")

//     tu.AssertEqual(t, (page2.Box().Width, page2.Box().Height) , (100, 100), "(page2")
//     html := page2.Box().Children[0]
//     body := html.Box().Children[0]
//     section, = body.Box().Children
//     p, = section.Box().Children
//     tu.AssertEqual(t, section.Box().ElementTag , "section", "section")
//     tu.AssertEqual(t, p.Box().ElementTag , "p", "p")

// func TestPageNames9(t *testing.T):
// capt := tu.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     pages = renderPages(`
//       <style>
//         @page { size: 200px 200px }
//         @page small { size: 100px 100px }
//         section { break-after: page; page: small }
//         article { page: small }
//       </style>
//       <section>
//         <div>big</div>
//         <div>big</div>
//       </section>
//       <article>
//         <div>small</div>
//         <div>small</div>
//       </article>
//     `)
//     page1, page2, = pages
// }
//     tu.AssertEqual(t, (page1.Box().Width, page1.Box().Height) , (100, 100), "(page1")
//     html := page1.Box().Children[0]
//     body := html.Box().Children[0]
//     section, = body.Box().Children
//     tu.AssertEqual(t, section.Box().ElementTag , "section", "section")

//     tu.AssertEqual(t, (page2.Box().Width, page2.Box().Height) , (100, 100), "(page2")
//     html := page2.Box().Children[0]
//     body := html.Box().Children[0]
//     article, = body.Box().Children
//     tu.AssertEqual(t, article.Box().ElementTag , "article", "article")

// @pytest.mark.parametrize("style, lineCounts", (
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     ("orphans: 2; widows: 2", [4, 3]),
//     ("orphans: 5; widows: 2", [0, 7]),
//     ("orphans: 2; widows: 4", [3, 4]),
//     ("orphans: 4; widows: 4", [0, 7]),
//     ("orphans: 2; widows: 2; page-break-inside: avoid", [0, 7]),
// ))
// func TestOrphansWidowsAvoid(t *testing.Tstyle, lineCounts) {
//     pages = renderPages(`
//       <style>
//         @page { size: 200px }
//         h1 { height: 120px }
//         p { line-height: 20px;
//             width: 1px; /* line break at each word */
//             %s }
//       </style>
//       <h1>Tasty test</h1>
//       <!-- There is room for 4 lines after h1 on the fist page -->
//       <p>one two three four five six seven</p>
//     ` % style)
//     for i, page := range enumerate(pages):
//         html := page.Box().Children[0]
//         body := html.Box().Children[0]
//         bodyChildren = body.Box().Children if i else body.Box().Children[1:]  // skip h1
//         count = len(bodyChildren[0].Box().Children) if bodyChildren else 0
//         tu.AssertEqual(t, lineCounts.pop(0) , count, "lineCounts")
//     tu.AssertEqual(t, ! lineCounts, "!")

// func TestPageAndLineboxBreaking(t *testing.T):
// capt := tu.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     // Empty <span/> tests a corner case := range skipFirstWhitespace()
//     pages = renderPages(`
//       <style>
//         @font-face { src: url(weasyprint.otf); font-family: weasyprint }
//         @page { size: 100px; margin: 2px; border: 1px solid }
//         body { margin: 0 }
//         div { font-family: weasyprint; font-size: 20px }
//       </style>
//       <div><span/>1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15</div>
//     `)
//     texts = []
//     for page := range pages {
//         html := page.Box().Children[0]
//         body := html.Box().Children[0]
//         div := body.Box().Children[0]
//         lines = div.Box().Children
//         for line := range lines {
//             lineTexts = []
//             for child := range line.descendants() {
//                 if isinstance(child, boxes.TextBox) {
//                     lineTexts.append(child.(*bo.TextBox).Text)
//                 }
//             } texts.append("".join(lineTexts))
//         }
//     } tu.AssertEqual(t, len(pages) , 4, "len")
//     tu.AssertEqual(t, "".join(texts) , "1,2,3,4,5,6,7,8,9,10,11,12,13,14,15", """.")
// }

// func TestMarginBoxesFixedDimension1(t *testing.T) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     // Corner boxes
//     page := renderOnePage(t, `
//       <style>
//         @page {
//           @top-left-corner {
//             content: "topLeft";
//             padding: 10px;
//           }
//           @top-right-corner {
//             content: "topRight";
//             padding: 10px;
//           }
//           @bottom-left-corner {
//             content: "bottomLeft";
//             padding: 10px;
//           }
//           @bottom-right-corner {
//             content: "bottomRight";
//             padding: 10px;
//           }
//           size: 1000px;
//           margin-top: 10%;
//           margin-bottom: 40%;
//           margin-left: 20%;
//           margin-right: 30%;
//         }
//       </style>
//     `)
//     html, topLeft, topRight, bottomLeft, bottomRight = page.Box().Children
//     for marginBox, text := range zip(
//             [topLeft, topRight, bottomLeft, bottomRight],
//             ["topLeft", "topRight", "bottomLeft", "bottomRight"]):

//         line := marginBox.Box().Children[0]
//         text := line.Box().Children[0]
//         tu.AssertEqual(t, text , text, "text")

//     // Check positioning && Rule 1 for fixed dimensions
//     tu.AssertEqual(t, topLeft.Box().PositionX , 0, "topLeft")
//     tu.AssertEqual(t, topLeft.Box().PositionY , 0, "topLeft")
//     tu.AssertEqual(t, topLeft.Box().MarginWidth() , 200  // margin-left, "topLeft")
//     tu.AssertEqual(t, topLeft.Box().MarginHeight() , 100  // margin-top, "topLeft")

//     tu.AssertEqual(t, topRight.Box().PositionX , 700  // size-x - margin-right, "topRight")
//     tu.AssertEqual(t, topRight.Box().PositionY , 0, "topRight")
//     tu.AssertEqual(t, topRight.Box().MarginWidth() , 300  // margin-right, "topRight")
//     tu.AssertEqual(t, topRight.Box().MarginHeight() , 100  // margin-top, "topRight")

//     tu.AssertEqual(t, bottomLeft.Box().PositionX , 0, "bottomLeft")
//     tu.AssertEqual(t, bottomLeft.Box().PositionY , 600  // size-y - margin-bottom, "bottomLeft")
//     tu.AssertEqual(t, bottomLeft.Box().MarginWidth() , 200  // margin-left, "bottomLeft")
//     tu.AssertEqual(t, bottomLeft.Box().MarginHeight() , 400  // margin-bottom, "bottomLeft")

//     tu.AssertEqual(t, bottomRight.Box().PositionX , 700  // size-x - margin-right, "bottomRight")
//     tu.AssertEqual(t, bottomRight.Box().PositionY , 600  // size-y - margin-bottom, "bottomRight")
//     tu.AssertEqual(t, bottomRight.Box().MarginWidth() , 300  // margin-right, "bottomRight")
//     tu.AssertEqual(t, bottomRight.Box().MarginHeight() , 400  // margin-bottom, "bottomRight")

// func TestMarginBoxesFixedDimension2(t *testing.T):
// capt := tu.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     // Test rules 2 && 3
//     page := renderOnePage(t, `
//       <style>
//         @page {
//           margin: 100px 200px;
//           @bottom-left-corner { content: ""; margin: 60px }
//         }
//       </style>
//     `)
//     html, marginBox = page.Box().Children
//     tu.AssertEqual(t, marginBox.Box().MarginWidth() , 200, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().MarginLeft , 60, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().MarginRight , 60, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().Width , 80  // 200 - 60 - 60, "marginBox")
// }
//     tu.AssertEqual(t, marginBox.Box().MarginHeight() , 100, "marginBox")
//     // total was too big, the outside margin was ignored {
//     } tu.AssertEqual(t, marginBox.Box().MarginTop , 60, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().MarginBottom , 40  // Not 60, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().Height , 0  // But ! negative, "marginBox")

// func TestMarginBoxesFixedDimension3(t *testing.T) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     // Test rule 3 with a non-auto inner dimension
//     page := renderOnePage(t, `
//       <style>
//         @page {
//           margin: 100px;
//           @left-middle { content: ""; margin: 10px; width: 130px }
//         }
//       </style>
//     `)
//     html, marginBox = page.Box().Children
//     tu.AssertEqual(t, marginBox.Box().MarginWidth() , 100, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().MarginLeft , -40  // Not 10px, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().MarginRight , 10, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().Width , 130  // As specified, "marginBox")

// func TestMarginBoxesFixedDimension4(t *testing.T):
// capt := tu.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     // Test rule 4
//     page := renderOnePage(t, `
//       <style>
//         @page {
//           margin: 100px;
//           @left-bottom {
//             content: "";
//             margin-left: 10px;
//             margin-right: auto;
//             width: 70px;
//           }
//         }
//       </style>
//     `)
//     html, marginBox = page.Box().Children
//     tu.AssertEqual(t, marginBox.Box().MarginWidth() , 100, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().MarginLeft , 10  // 10px this time, no over-constrain, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().MarginRight , 20, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().Width , 70  // As specified, "marginBox")
// }

// func TestMarginBoxesFixedDimension5(t *testing.T) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     // Test rules 2, 3 && 4
//     page := renderOnePage(t, `
//       <style>
//         @page {
//           margin: 100px;
//           @right-top {
//             content: "";
//             margin-right: 10px;
//             margin-left: auto;
//             width: 130px;
//           }
//         }
//       </style>
//     `)
//     html, marginBox = page.Box().Children
//     tu.AssertEqual(t, marginBox.Box().MarginWidth() , 100, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().MarginLeft , 0  // rule 2, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().MarginRight , -30  // rule 3, after rule 2, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().Width , 130  // As specified, "marginBox")

// func TestMarginBoxesFixedDimension6(t *testing.T):
// capt := tu.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     // Test rule 5
//     page := renderOnePage(t, `
//       <style>
//         @page {
//           margin: 100px;
//           @top-left { content: ""; margin-top: 10px; margin-bottom: auto }
//         }
//       </style>
//     `)
//     html, marginBox = page.Box().Children
//     tu.AssertEqual(t, marginBox.Box().MarginHeight() , 100, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().MarginTop , 10, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().MarginBottom , 0, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().Height , 90, "marginBox")
// }

// func TestMarginBoxesFixedDimension7(t *testing.T) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     // Test rule 5
//     page := renderOnePage(t, `
//       <style>
//         @page {
//           margin: 100px;
//           @top-center { content: ""; margin: auto 0 }
//         }
//       </style>
//     `)
//     html, marginBox = page.Box().Children
//     tu.AssertEqual(t, marginBox.Box().MarginHeight() , 100, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().MarginTop , 0, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().MarginBottom , 0, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().Height , 100, "marginBox")

// func TestMarginBoxesFixedDimension8(t *testing.T):
// capt := tu.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     // Test rule 6
//     page := renderOnePage(t, `
//       <style>
//         @page {
//           margin: 100px;
//           @bottom-right { content: ""; margin: auto; height: 70px }
//         }
//       </style>
//     `)
//     html, marginBox = page.Box().Children
//     tu.AssertEqual(t, marginBox.Box().MarginHeight() , 100, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().MarginTop , 15, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().MarginBottom , 15, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().Height , 70, "marginBox")
// }

// func TestMarginBoxesFixedDimension9(t *testing.T) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     // Rule 2 inhibits rule 6
//     page := renderOnePage(t, `
//       <style>
//         @page {
//           margin: 100px;
//           @bottom-center { content: ""; margin: auto 0; height: 150px }
//         }
//       </style>
//     `)
//     html, marginBox = page.Box().Children
//     tu.AssertEqual(t, marginBox.Box().MarginHeight() , 100, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().MarginTop , 0, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().MarginBottom , -50  // outside, "marginBox")
//     tu.AssertEqual(t, marginBox.Box().Height , 150, "marginBox")

// func images(*widths):
//     return " ".join(
//         f"url(\"data:image/svg+xml,<svg width="{width}" height="10"></svg>\")"
//         for width := range widths)

// @pytest.mark.parametrize("css, widths", (
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     (`@top-left { content: %s }
//         @top-center { content: %s }
//         @top-right { content: %s }
//      ` % (images(50, 50), images(50, 50), images(50, 50)),
//      [100, 100, 100]),  // Use preferred widths if they fit
//     (`@top-left { content: %s; margin: auto }
//         @top-center { content: %s }
//         @top-right { content: %s }
//      ` % (images(50, 50), images(50, 50), images(50, 50)),
//      [100, 100, 100]),  // "auto" margins are set to 0
//     (`@top-left { content: %s }
//         @top-center { content: %s }
//         @top-right { content: "foo"; width: 200px }
//      ` % (images(100, 50), images(300, 150)),
//      [150, 300, 200]),  // Use at least minimum widths, even if boxes overlap
//     (`@top-left { content: %s }
//         @top-center { content: %s }
//         @top-right { content: %s }
//      ` % (images(150, 150), images(150, 150), images(150, 150)),
//      [200, 200, 200]),  // Distribute remaining space proportionally
//     (`@top-left { content: %s }
//         @top-center { content: %s }
//         @top-right { content: %s }
//      ` % (images(100, 100, 100), images(100, 100), images(10)),
//      [220, 160, 10]),
//     (`@top-left { content: %s; width: 205px }
//         @top-center { content: %s }
//         @top-right { content: %s }
//      ` % (images(100, 100, 100), images(100, 100), images(10)),
//      [205, 190, 10]),
//     (`@top-left { width: 1000px; margin: 1000px; padding: 1000px;
//                     border: 1000px solid }
//         @top-center { content: %s }
//         @top-right { content: %s }
//      ` % (images(100, 100), images(10)),
//      [200, 10]),  // "width" && other have no effect without "content"
//     (`@top-left { content: ""; width: 200px }
//         @top-center { content: ""; width: 300px }
//         @top-right { content: %s }
//      ` % images(50, 50),  // This leaves 150px for @top-right’s shrink-to-fit
//      [200, 300, 100]),
//     (`@top-left { content: ""; width: 200px }
//         @top-center { content: ""; width: 300px }
//         @top-right { content: %s }
//      ` % images(100, 100, 100),
//      [200, 300, 150]),
//     (`@top-left { content: ""; width: 200px }
//         @top-center { content: ""; width: 300px }
//         @top-right { content: %s }
//      ` % images(170, 175),
//      [200, 300, 175]),
//     (`@top-left { content: ""; width: 200px }
//         @top-center { content: ""; width: 300px }
//         @top-right { content: %s }
//      ` % images(170, 175),
//      [200, 300, 175]),
//     (`@top-left { content: ""; width: 200px }
//         @top-right { content: ""; width: 500px }
//      `,
//      [200, 500]),
//     (`@top-left { content: ""; width: 200px }
//         @top-right { content: %s }
//      ` % images(150, 50, 150),
//      [200, 350]),
//     (`@top-left { content: ""; width: 200px }
//         @top-right { content: %s }
//      ` % images(150, 50, 150, 200),
//      [200, 400]),
//     (`@top-left { content: %s }
//         @top-right { content: ""; width: 200px }
//      ` % images(150, 50, 450),
//      [450, 200]),
//     (`@top-left { content: %s }
//         @top-right { content: %s }
//      ` % (images(150, 100), images(10, 120)),
//      [250, 130]),
//     (`@top-left { content: %s }
//         @top-right { content: %s }
//      ` % (images(550, 100), images(10, 120)),
//      [550, 120]),
//     (`@top-left { content: %s }
//         @top-right { content: %s }
//      ` % (images(250, 60), images(250, 180)),
//      [275, 325]),  // 250 + (100 * 1 / 4), 250 + (100 * 3 / 4)
// ))
// func TestPageStyle(t *testing.Tcss, widths):
//     expectedAtKeywords = [
//         atKeyword for atKeyword := range [
//             "@top-left", "@top-center", "@top-right"]
//         if atKeyword + " { content: " := range css]
//     page := renderOnePage(t, `
//       <style>
//         @page {
//           size: 800px;
//           margin: 100px;
//           padding: 42px;
//           border: 7px solid;
//           %s
//         }
//       </style>
//     ` % css)
//     tu.AssertEqual(t, page.Box().Children[0].Box().ElementTag , "html", "page")
//     marginBoxes = page.Box().Children[1:]
//     tu.AssertEqual(t, [box.atKeyword for box := range marginBoxes] , expectedAtKeywords, "[box")
//     offsets = {"@top-left": 0, "@top-center": 0.5, "@top-right": 1}
//     for box := range marginBoxes {
//         tu.AssertEqual(t, box.Box().PositionX , 100 + offsets[box.atKeyword] * (, "box")
//             600 - box.Box().MarginWidth())
//     } tu.AssertEqual(t, [box.Box().MarginWidth() for box := range marginBoxes] , widths, "[box")
// }

// func TestMarginBoxesVerticalAlign(t *testing.T) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     // 3 px ->    +-----+
//     //            |  1  |
//     //            +-----+
//     #
//     //        43 px ->   +-----+
//     //        53 px ->   |  2  |
//     //                   +-----+
//     #
//     //               83 px ->   +-----+
//     //                          |  3  |
//     //               103px ->   +-----+
//     page := renderOnePage(t, `
//       <style>
//         @page {
//           size: 800px;
//           margin: 106px;  /* margin boxes’ content height is 100px */
// }
//           @top-left {
//             content: "foo"; line-height: 20px; border: 3px solid;
//             vertical-align: top;
//           }
//           @top-center {
//             content: "foo"; line-height: 20px; border: 3px solid;
//             vertical-align: middle;
//           }
//           @top-right {
//             content: "foo"; line-height: 20px; border: 3px solid;
//             vertical-align: bottom;
//           }
//         }
//       </style>
//     `)
//     html, topLeft, topCenter, topRight = page.Box().Children
//     line1, = topLeft.Box().Children
//     line2, = topCenter.Box().Children
//     line3, = topRight.Box().Children
//     tu.AssertEqual(t, line1.Box().PositionY , 3, "line1")
//     tu.AssertEqual(t, line2.Box().PositionY , 43, "line2")
//     tu.AssertEqual(t, line3.Box().PositionY , 83, "line3")

// func TestMarginBoxesElement(t *testing.T):
// capt := tu.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     pages = renderPages(`
//       <style>
//         footer {
//           position: running(footer);
//         }
//         @page {
//           margin: 50px;
//           size: 200px;
//           @bottom-center {
//             content: element(footer);
//           }
//         }
//         h1 {
//           height: 40px;
//         }
//         .pages:before {
//           content: counter(page);
//         }
//         .pages:after {
//           content: counter(pages);
//         }
//       </style>
//       <footer class="pages"> of </footer>
//       <h1>test1</h1>
//       <h1>test2</h1>
//       <h1>test3</h1>
//       <h1>test4</h1>
//       <h1>test5</h1>
//       <h1>test6</h1>
//       <footer>Static</footer>
//     `)
//     footer1Text = "".join(
//         getattr(node, "text", "")
//         for node := range pages[0].Box().Children[1].descendants())
//     tu.AssertEqual(t, footer1Text , "1 of 3", "footer1Text")

//     footer2Text = "".join(
//         getattr(node, "text", "")
//         for node := range pages[1].Box().Children[1].descendants())
//     tu.AssertEqual(t, footer2Text , "2 of 3", "footer2Text")

//     footer3Text = "".join(
//         getattr(node, "text", "")
//         for node := range pages[2].Box().Children[1].descendants())
//     tu.AssertEqual(t, footer3Text , "Static", "footer3Text")

// @pytest.mark.parametrize("argument, texts", (
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     // TODO: start doesn’t work because running elements are removed from the
//     // original tree, && the current implentation in
//     // layout.getRunningElementFor uses the tree to know if it’s at the
//     // beginning of the page

//     // ("start", ("", "2-first", "2-last", "3-last", "5")),

//     ("first", ("", "2-first", "3-first", "3-last", "5")),
//     ("last", ("", "2-last", "3-last", "3-last", "5")),
//     ("first-except", ("", "", "", "3-last", "")),
// ))
// func TestRunningElements(t *testing.Targument, texts) {
//     pages = renderPages(`
//       <style>
//         @page {
//           margin: 50px;
//           size: 200px;
//           @bottom-center { content: element(title %s) }
//         }
//         article { break-after: page }
//         h1 { position: running(title) }
//       </style>
//       <article>
//         <div>1</div>
//       </article>
//       <article>
//         <h1>2-first</h1>
//         <h1>2-last</h1>
//       </article>
//       <article>
//         <p>3</p>
//         <h1>3-first</h1>
//         <h1>3-last</h1>
//       </article>
//       <article>
//       </article>
//       <article>
//         <h1>5</h1>
//       </article>
//     ` % argument)
//     tu.AssertEqual(t, len(pages) , 5, "len")
//     for page, text := range zip(pages, texts):
//         html, margin = page.Box().Children
//         if margin.Box().Children:
//             h1, = margin.Box().Children
//             line := h1.Box().Children[0]
//             textbox, = line.Box().Children
//             tu.AssertEqual(t, textbox.(*bo.TextBox).Text , text, "textbox")
//         else:
//             tu.AssertEqual(t, ! text, "!")

// func TestRunningElementsDisplay(t *testing.T):
// capt := tu.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     page := renderOnePage(t, `
//       <style>
//         @page {
//           margin: 50px;
//           size: 200px;
//           @bottom-left { content: element(inline) }
//           @bottom-center { content: element(block) }
//           @bottom-right { content: element(table) }
//         }
//         table { position: running(table) }
//         div { position: running(block) }
//         span { position: running(inline) }
//       </style>
//       text
//       <table><tr><td>table</td></tr></table>
//       <div>block</div>
//       <span>inline</span>
//     `)
//     html, left, center, right = page.Box().Children
//     tu.AssertEqual(t, "".join(, """.")
//         getattr(node, "text", "") for node := range left.descendants()) , "inline"
//     tu.AssertEqual(t, "".join(, """.")
//         getattr(node, "text", "") for node := range center.descendants()) , "block"
//     tu.AssertEqual(t, "".join(, """.")
//         getattr(node, "text", "") for node := range right.descendants()) , "table"
