package document

import (
	"fmt"
	"testing"

	bo "github.com/benoitkugler/go-weasyprint/boxes"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/utils/testutils"
)

func printBoxes(boxes []Box) {
	for _, b := range boxes {
		fmt.Printf("<%s %s> ", b.Type(), b.Box().ElementTag)
	}
}

func TestTextFontSizeZero(t *testing.T) {
	cp := testutils.CaptureLogs()
	cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        p { font-size: 0; }
      </style>
      <p>test font size zero</p>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line := paragraph.Box().Children[0]
	// zero-sized text boxes are removed
	if len(line.Box().Children) != 0 {
		t.Fatalf("expected 0, got %v", line.Box().Children)
	}
	if got := line.Box().Height; got != pr.Float(0) {
		t.Fatalf("expected 0, got %v", got)
	}
	if got := paragraph.Box().Height; got != pr.Float(0) {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestTextSpacedInlines(t *testing.T) {
	cp := testutils.CaptureLogs()
	cp.AssertNoLogs(t)

	page := renderOnePage(t, `
		<p>start <i><b>bi1</b> <b>bi2</b></i> <b>b1</b> end</p>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line := paragraph.Box().Children[0]
	start, i, space, b, end := line.Box().Children[0], line.Box().Children[1], line.Box().Children[2], line.Box().Children[3], line.Box().Children[4]
	if tb := start.(*bo.TextBox); tb.Text != "start " {
		t.Fatalf("expected %s, got %s", "start ", tb.Text)
	}
	if tb := space.(*bo.TextBox); tb.Text != " " {
		t.Fatalf("expected %s, got %s", " ", tb.Text)
	}
	if w := space.Box().Width.V(); w <= 0 {
		t.Fatalf("expected positive width, got %f", w)
	}
	if tb := end.(*bo.TextBox); tb.Text != " end" {
		t.Fatalf("expected %s, got %s", " end", tb.Text)
	}

	bi1, space, bi2 := i.Box().Children[0], i.Box().Children[1], i.Box().Children[2]
	bi1 = bi1.Box().Children[0]
	bi2 = bi2.Box().Children[0]
	if tb := bi1.(*bo.TextBox); tb.Text != "bi1" {
		t.Fatalf("expected %s, got %s", "bi1", tb.Text)
	}
	if tb := space.(*bo.TextBox); tb.Text != " " {
		t.Fatalf("expected %s, got %s", " ", tb.Text)
	}
	if w := space.Box().Width.V(); w <= 0 {
		t.Fatalf("expected positive width, got %f", w)
	}
	if tb := bi2.(*bo.TextBox); tb.Text != "bi2" {
		t.Fatalf("expected %s, got %s", "bi2", tb.Text)
	}

	b1 := b.Box().Children[0]
	if tb := b1.(*bo.TextBox); tb.Text != "b1" {
		t.Fatalf("expected %s, got %s", "b1", tb.Text)
	}
}

func TestTextAlignLeft(t *testing.T) {
	cp := testutils.CaptureLogs()
	cp.AssertNoLogs(t)
	// <-------------------->  page, body
	//     +-----+
	// +---+     |
	// |   |     |
	// +---+-----+

	// ^   ^     ^          ^
	// x=0 x=40  x=100      x=200
	page := renderOnePage(t, `
      <style>
        @page { size: 200px }
      </style>
      <body>
        <img src="pattern.png" style="width: 40px"
        ><img src="pattern.png" style="width: 60px">`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	img1, img2 := line.Box().Children[0], line.Box().Children[1]
	// initial value for text-align: left (in ltr text)
	if got := img1.Box().PositionX; got != 0 {
		t.Fatalf("expected %d, got %f", 0, got)
	}
	if got := img2.Box().PositionX; got != 40 {
		t.Fatalf("expected %d, got %f", 40, got)
	}
}

func TestTextAlignRight(t *testing.T) {
	cp := testutils.CaptureLogs()
	cp.AssertNoLogs(t)
	// <-------------------->  page, body
	//                +-----+
	//            +---+     |
	//            |   |     |
	//            +---+-----+

	// ^          ^   ^     ^
	// x=0        x=100     x=200
	//                x=140
	page := renderOnePage(t, `
      <style>
        @page { size: 200px }
        body { text-align: right }
      </style>
      <body>
        <img src="pattern.png" style="width: 40px"
        ><img src="pattern.png" style="width: 60px">`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	img1, img2 := line.Box().Children[0], line.Box().Children[1]

	if got := img1.Box().PositionX; got != 100 { // 200 - 60 - 40
		t.Fatalf("expected %d, got %f", 100, got)
	}
	if got := img2.Box().PositionX; got != 140 { // 200 - 60
		t.Fatalf("expected %d, got %f", 140, got)
	}
}

func TestTextAlignCenter(t *testing.T) {
	cp := testutils.CaptureLogs()
	cp.AssertNoLogs(t)

	// <-------------------->  page, body
	//           +-----+
	//       +---+     |
	//       |   |     |
	//       +---+-----+

	// ^     ^   ^     ^
	// x=    x=50     x=150
	//           x=90
	page := renderOnePage(t, `
      <style>
        @page { size: 200px }
        body { text-align: center }
      </style>
      <body>
        <img src="pattern.png" style="width: 40px"
        ><img src="pattern.png" style="width: 60px">`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	img1, img2 := line.Box().Children[0], line.Box().Children[1]

	if got := img1.Box().PositionX; got != 50 {
		t.Fatalf("expected %d, got %f", 50, got)
	}
	if got := img2.Box().PositionX; got != 90 {
		t.Fatalf("expected %d, got %f", 90, got)
	}
}

func TestTextAlignJustify(t *testing.T) {
	cp := testutils.CaptureLogs()
	cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @page { size: 300px 1000px }
        body { text-align: justify }
      </style>
      <p><img src="pattern.png" style="width: 40px">
        <strong>
          <img src="pattern.png" style="width: 60px">
          <img src="pattern.png" style="width: 10px">
          <img src="pattern.png" style="width: 100px"
        ></strong><img src="pattern.png" style="width: 290px"
        ><!-- Last image will be on its own line. -->`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line1, line2 := paragraph.Box().Children[0], paragraph.Box().Children[1]
	image1, space1, strong := line1.Box().Children[0], line1.Box().Children[1], line1.Box().Children[2]
	sc := strong.Box().Children
	image2, space2, image3, space3, image4 := sc[0], sc[1], sc[2], sc[3], sc[4]
	image5 := line2.Box().Children[0]
	if text := space1.(*bo.TextBox).Text; text != " " {
		t.Fatalf("expected ' ', got %s", text)
	}
	if text := space2.(*bo.TextBox).Text; text != " " {
		t.Fatalf("expected ' ', got %s", text)
	}
	if text := space3.(*bo.TextBox).Text; text != " " {
		t.Fatalf("expected ' ', got %s", text)
	}

	if got := image1.Box().PositionX; got != 0 {
		t.Fatalf("expected %d, got %f", 0, got)
	}
	if got := space1.Box().PositionX; got != 40 {
		t.Fatalf("expected %d, got %f", 40, got)
	}
	if got := strong.Box().PositionX; got != 70 {
		t.Fatalf("expected %d, got %f", 70, got)
	}
	if got := image2.Box().PositionX; got != 70 {
		t.Fatalf("expected %d, got %f", 70, got)
	}
	if got := space2.Box().PositionX; got != 130 {
		t.Fatalf("expected %d, got %f", 130, got)
	}
	if got := image3.Box().PositionX; got != 160 {
		t.Fatalf("expected %d, got %f", 160, got)
	}
	if got := space3.Box().PositionX; got != 170 {
		t.Fatalf("expected %d, got %f", 170, got)
	}
	if got := image4.Box().PositionX; got != 200 {
		t.Fatalf("expected %d, got %f", 200, got)
	}
	if got := strong.Box().Width.V(); got != 230 {
		t.Fatalf("expected %d, got %f", 230, got)
	}

	if got := image5.Box().PositionX; got != 0 {
		t.Fatalf("expected %d, got %f", 0, got)
	}
}

// func TestTextAlignJustifyAll(t *testing.T) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     page := renderOnePage(t, `
//       <style>
//         @page { size: 300px 1000px }
//         body { text-align: justify-all }
//       </style>
//       <p><img src="pattern.png" style="width: 40px">
//         <strong>
//           <img src="pattern.png" style="width: 60px">
//           <img src="pattern.png" style="width: 10px">
//           <img src="pattern.png" style="width: 100px"
//         ></strong><img src="pattern.png" style="width: 200px">
//         <img src="pattern.png" style="width: 10px">`)
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     paragraph := body.Box().Children[0]
//     line1, line2 = paragraph.Box().Children
//     image1, space1, strong = line1.Box().Children
//     image2, space2, image3, space3, image4 = strong.Box().Children
//     image5, space4, image6 = line2.Box().Children
//     assert space1.text == " "
//     assert space2.text == " "
//     assert space3.text == " "
//     assert space4.text == " "
// }
//     assert image1.positionX == 0
//     assert space1.positionX == 40
//     assert strong.positionX == 70
//     assert image2.positionX == 70
//     assert space2.positionX == 130
//     assert image3.positionX == 160
//     assert space3.positionX == 170
//     assert image4.positionX == 200
//     assert strong.width == 230

//     assert image5.positionX == 0
//     assert space4.positionX == 200
//     assert image6.positionX == 290

// func TestTextAlignAllLast(t *testing.T) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     page := renderOnePage(t, `
//       <style>
//         @page { size: 300px 1000px }
//         body { text-align-all: justify; text-align-last: right }
//       </style>
//       <p><img src="pattern.png" style="width: 40px">
//         <strong>
//           <img src="pattern.png" style="width: 60px">
//           <img src="pattern.png" style="width: 10px">
//           <img src="pattern.png" style="width: 100px"
//         ></strong><img src="pattern.png" style="width: 200px"
//         ><img src="pattern.png" style="width: 10px">`)
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     paragraph := body.Box().Children[0]
//     line1, line2 = paragraph.Box().Children
//     image1, space1, strong = line1.Box().Children
//     image2, space2, image3, space3, image4 = strong.Box().Children
//     image5, image6 = line2.Box().Children
//     assert space1.text == " "
//     assert space2.text == " "
//     assert space3.text == " "
// }
//     assert image1.positionX == 0
//     assert space1.positionX == 40
//     assert strong.positionX == 70
//     assert image2.positionX == 70
//     assert space2.positionX == 130
//     assert image3.positionX == 160
//     assert space3.positionX == 170
//     assert image4.positionX == 200
//     assert strong.width == 230

//     assert image5.positionX == 90
//     assert image6.positionX == 290

// func TestTextAlignNotEnoughSpace(t *testing.T) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     page := renderOnePage(t, `
//       <style>
//         p { text-align: center; width: 0 }
//         span { display: inline-block }
//       </style>
//       <p><span>aaaaaaaaaaaaaaaaaaaaaaaaaa</span></p>`)
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     paragraph := body.Box().Children[0]
//     span, = paragraph.Box().Children
//     assert span.positionX == 0
// }

// func TestTextAlignJustifyNoSpace(t *testing.T) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     // single-word line (zero spaces)
//     page := renderOnePage(t, `
//       <style>
//         body { text-align: justify; width: 50px }
//       </style>
//       <p>Supercalifragilisticexpialidocious bar</p>
//     `)
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     paragraph := body.Box().Children[0]
//     line1, line2 = paragraph.Box().Children
//     text, = line1.Box().Children
//     assert text.positionX == 0

// func TestTextAlignJustifyTextIndent(t *testing.T){
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     // text-indent
//     page := renderOnePage(t, `
//       <style>
//         @page { size: 300px 1000px }
//         body { text-align: justify }
//         p { text-indent: 3px }
//       </style>
//       <p><img src="pattern.png" style="width: 40px">
//         <strong>
//           <img src="pattern.png" style="width: 60px">
//           <img src="pattern.png" style="width: 10px">
//           <img src="pattern.png" style="width: 100px"
//         ></strong><img src="pattern.png" style="width: 290px"
//         ><!-- Last image will be on its own line. -->`)
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     paragraph := body.Box().Children[0]
//     line1, line2 = paragraph.Box().Children
//     image1, space1, strong = line1.Box().Children
//     image2, space2, image3, space3, image4 = strong.Box().Children
//     image5, = line2.Box().Children
//     assert space1.text == " "
//     assert space2.text == " "
//     assert space3.text == " "

//     assert image1.positionX == 3
//     assert space1.positionX == 43
//     assert strong.positionX == 72
//     assert image2.positionX == 72
//     assert space2.positionX == 132
//     assert image3.positionX == 161
//     assert space3.positionX == 171
//     assert image4.positionX == 200
//     assert strong.width == 228

//     assert image5.positionX == 0

// func TestTextAlignJustifyNoBreakBetweenChildren(t *testing.T){
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     // Test justification when line break happens between two inline children
//     // that must stay together.
//     // Test regression: https://github.com/Kozea/WeasyPrint/issues/637
//     page := renderOnePage(t, `
//       <style>
//         @font-face {src: url(weasyprint.otf); font-family: weasyprint}
//         p { text-align: justify; font-family: weasyprint; width: 7em }
//       </style>
//       <p>
//         <span>a</span>
//         <span>b</span>
//         <span>bla</span><span>,</span>
//         <span>b</span>
//       </p>
//     `)
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     paragraph := body.Box().Children[0]
//     line1, line2 = paragraph.Box().Children
// }
//     span1, space1, span2, space2 = line1.Box().Children
//     assert span1.positionX == 0
//     assert span2.positionX == 6 * 16  // 1 character + 5 spaces
//     assert line1.width == 7 * 16  // 7em

//     span1, span2, space1, span3, space2 = line2.Box().Children
//     assert span1.positionX == 0
//     assert span2.positionX == 3 * 16  // 3 characters
//     assert span3.positionX == 5 * 16  // (3 + 1) characters + 1 space

// func TestWordSpacing(t *testing.T) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     // keep the empty <style> as a regression test: element.text  == nil
//     // (Not a string.)
//     page := renderOnePage(t, `
//       <style></style>
//       <body><strong>Lorem ipsum dolor<em>sit amet</em></strong>`)
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     line, = body.Box().Children
//     strong1, = line.Box().Children
// }
//     // TODO: Pango gives only half of word-spacing to a space at the end
//     // of a TextBox. Is this what we want?
//     page := renderOnePage(t, `
//       <style>strong { word-spacing: 11px }</style>
//       <body><strong>Lorem ipsum dolor<em>sit amet</em></strong>`)
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     line, = body.Box().Children
//     strong2, = line.Box().Children
//     assert strong2.width - strong1.width == 33

// func TestLetterSpacing1(t *testing.T) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     page := renderOnePage(t, `
//         <body><strong>Supercalifragilisticexpialidocious</strong>`)
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     line, = body.Box().Children
//     strong1, = line.Box().Children
// }
//     page := renderOnePage(t, `
//         <style>strong { letter-spacing: 11px }</style>
//         <body><strong>Supercalifragilisticexpialidocious</strong>`)
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     line, = body.Box().Children
//     strong2, = line.Box().Children
//     assert strong2.width - strong1.width == 34 * 11

//     // an embedded tag should ! affect the single-line letter spacing
//     page, = renderPages(
//         "<style>strong { letter-spacing: 11px }</style>"
//         "<body><strong>Supercali<span>fragilistic</span>expialidocious"
//         "</strong>")
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     line, = body.Box().Children
//     strong3, = line.Box().Children
//     assert strong3.width == strong2.width

//     // duplicate wrapped lines should also have same overall width
//     // Note work-around for word-wrap bug (issue #163) by marking word
//     // as an inline-block
//     page, = renderPages(
//         "<style>"
//         "  strong {"
//         "    letter-spacing: 11px;"
//         f"    max-width: {strong3.width * 1.5}px"
//         "}"
//         "  span { display: inline-block }"
//         "</style>"
//         "<body><strong>"
//         "  <span>Supercali<i>fragilistic</i>expialidocious</span> "
//         "  <span>Supercali<i>fragilistic</i>expialidocious</span>"
//         "</strong>")
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     line1, line2 = body.Box().Children
//     assert line1.Box().Children[0].width == line2.Box().Children[0].width
//     assert line1.Box().Children[0].width == strong2.width

// @pytest.mark.parametrize("spacing", ("word-spacing", "letter-spacing"))
// func TestSpacingEx(t *testing.Tspacing) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     // Test regression on ex units := range spacing properties
//     renderPages(f"<div style="{spacing}: 2ex">abc def")
// }

// @pytest.mark.parametrize("indent", ("12px", "6%"))
// func TestTextIndent(t *testing.Tindent) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     page := renderOnePage(t, `
//         <style>
//             @page { size: 220px }
//             body { margin: 10px; text-indent: %(indent)s }
//         </style>
//         <p>Some text that is long enough that it take at least three line,
//            but maybe more.
//     ` % {"indent": indent})
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     paragraph := body.Box().Children[0]
//     lines = paragraph.Box().Children
//     text1, = lines[0].Box().Children
//     text2, = lines[1].Box().Children
//     text3, = lines[2].Box().Children
//     assert text1.positionX == 22  // 10px margin-left + 12px indent
//     assert text2.positionX == 10  // No indent
//     assert text3.positionX == 10  // No indent

// func TestTextIndentInline(t *testing.T){
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     // Test regression: https://github.com/Kozea/WeasyPrint/issues/1000
//     page := renderOnePage(t, `
//         <style>
//             @font-face { src: url(weasyprint.otf); font-family: weasyprint }
//             p { display: inline-block; text-indent: 1em;
//                 font-family: weasyprint }
//         </style>
//         <p><span>text
//     `)
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     paragraph := body.Box().Children[0]
//     line := paragraph.Box().Children0
//     assert line.width == (4 + 1) * 16
// }

// @pytest.mark.parametrize("indent", ("12px", "6%"))
// func TestTextIndentMultipage(t *testing.Tindent) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     // Test regression: https://github.com/Kozea/WeasyPrint/issues/706
//     pages = renderPages(`
//         <style>
//             @page { size: 220px 1.5em; margin: 0 }
//             body { margin: 10px; text-indent: %(indent)s }
//         </style>
//         <p>Some text that is long enough that it take at least three line,
//            but maybe more.
//     ` % {"indent": indent})
//     page = pages.pop(0)
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     paragraph := body.Box().Children[0]
//     line := paragraph.Box().Children0
//     text, = line.Box().Children
//     assert text.positionX == 22  // 10px margin-left + 12px indent

//     page = pages.pop(0)
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     paragraph := body.Box().Children[0]
//     line := paragraph.Box().Children0
//     text, = line.Box().Children
//     assert text.positionX == 10  // No indent

// func TestHyphenateCharacter1(t *testing.T){
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     page, = renderPages(
//         "<html style="width: 5em; font-family: weasyprint">"
//         "<style>"
//         "  @font-face {src: url(weasyprint.otf); font-family: weasyprint}"
//         "</style>"
//         "<body style="hyphens: auto;"
//         "hyphenate-character: \"!\"" lang=fr>"
//         "hyphénation")
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     lines = body.Box().Children
//     assert len(lines) > 1
//     assert lines[0].Box().Children[0].text.endswith("!")
//     fullText = "".join(line.Box().Children[0].text for line := range lines)
//     assert fullText.replace("!", "") == "hyphénation"

// func TestHyphenateCharacter2(t *testing.T){
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     page, = renderPages(
//         "<html style="width: 5em; font-family: weasyprint">"
//         "<style>"
//         "  @font-face {src: url(weasyprint.otf); font-family: weasyprint}"
//         "</style>"
//         "<body style="hyphens: auto;"
//         "hyphenate-character: \"à\"" lang=fr>"
//         "hyphénation")
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     lines = body.Box().Children
//     assert len(lines) > 1
//     assert lines[0].Box().Children[0].text.endswith("à")
//     fullText = "".join(line.Box().Children[0].text for line := range lines)
//     assert fullText.replace("à", "") == "hyphénation"

// func TestHyphenateCharacter3(t *testing.T){
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     page, = renderPages(
//         "<html style="width: 5em; font-family: weasyprint">"
//         "<style>"
//         "  @font-face {src: url(weasyprint.otf); font-family: weasyprint}"
//         "</style>"
//         "<body style="hyphens: auto;"
//         "hyphenate-character: \"ù ù\"" lang=fr>"
//         "hyphénation")
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     lines = body.Box().Children
//     assert len(lines) > 1
//     assert lines[0].Box().Children[0].text.endswith("ù ù")
//     fullText = "".join(line.Box().Children[0].text for line := range lines)
//     assert fullText.replace(" ", "").replace("ù", "") == "hyphénation"

// func TestHyphenateCharacter4(t *testing.T){
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     page, = renderPages(
//         "<html style="width: 5em; font-family: weasyprint">"
//         "<style>"
//         "  @font-face {src: url(weasyprint.otf); font-family: weasyprint}"
//         "</style>"
//         "<body style="hyphens: auto;"
//         "hyphenate-character: \"\"" lang=fr>"
//         "hyphénation")
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     lines = body.Box().Children
//     assert len(lines) > 1
//     fullText = "".join(line.Box().Children[0].text for line := range lines)
//     assert fullText == "hyphénation"

// func TestHyphenateCharacter5(t *testing.T){
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     page, = renderPages(
//         "<html style="width: 5em; font-family: weasyprint">"
//         "<style>"
//         "  @font-face {src: url(weasyprint.otf); font-family: weasyprint}"
//         "</style>"
//         "<body style="hyphens: auto;"
//         "hyphenate-character: \"———\"" lang=fr>"
//         "hyphénation")
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     lines = body.Box().Children
//     assert len(lines) > 1
//     assert lines[0].Box().Children[0].text.endswith("———")
//     fullText = "".join(line.Box().Children[0].text for line := range lines)
//     assert fullText.replace("—", "") == "hyphénation"

// func TestHyphenateManual1(t *testing.T){
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     for i := range range(1, len("hyphénation")):
//         for hyphenateCharacter := range ("!", "ù ù"):
//             word = "hyphénation"[:i] + "\u00ad" + "hyphénation"[i:]
//             page, = renderPages(
//                 "<html style="width: 5em; font-family: weasyprint">"
//                 "<style>@font-face {"
//                 "  src: url(weasyprint.otf); font-family: weasyprint}</style>"
//                 "<body style="hyphens: manual;"
//                 f"  hyphenate-character: \"{hyphenateCharacter}\`
//                 f"  lang=fr>{word}")
//             html := page.Box().Children[0]
//             body := html.Box().Children[0]
//             lines = body.Box().Children
//             assert len(lines) == 2
//             assert lines[0].Box().Children[0].text.endswith(hyphenateCharacter)
//             fullText = "".join(
//                 child.text for line := range lines for child := range line.Box().Children)
//             assert fullText.replace(hyphenateCharacter, "") == word

// func TestHyphenateManual2(t *testing.T){
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     for i := range range(1, len("hy phénation")):
//         for hyphenateCharacter := range ("!", "ù ù"):
//             word = "hy phénation"[:i] + "\u00ad" + "hy phénation"[i:]
//             page, = renderPages(
//                 "<html style="width: 5em; font-family: weasyprint">"
//                 "<style>@font-face {"
//                 "  src: url(weasyprint.otf); font-family: weasyprint}</style>"
//                 "<body style="hyphens: manual;"
//                 f"  hyphenate-character: \"{hyphenateCharacter}\`
//                 f"  lang=fr>{word}")
//             html := page.Box().Children[0]
//             body := html.Box().Children[0]
//             lines = body.Box().Children
//             assert len(lines) := range (2, 3)
//             fullText = "".join(
//                 child.text for line := range lines for child := range line.Box().Children)
//             fullText = fullText.replace(hyphenateCharacter, "")
//             if lines[0].Box().Children[0].text.endswith(hyphenateCharacter):
//                 assert fullText == word
//             else:
//                 assert lines[0].Box().Children[0].text.endswith("y")
//                 if len(lines) == 3:
//                     assert lines[1].Box().Children[0].text.endswith(
//                         hyphenateCharacter)

// func TestHyphenateManual3(t *testing.T){
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     // Automatic hyphenation opportunities within a word must be ignored if the
//     // word contains a conditional hyphen, := range favor of the conditional
//     // hyphen(s).
//     page, = renderPages(
//         "<html style="width: 0.1em" lang="en">"
//         "<body style="hyphens: auto">in&shy;lighten&shy;lighten&shy;in")
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     line1, line2, line3, line4 = body.Box().Children
//     assert line1.Box().Children[0].text == "in\xad‐"
//     assert line2.Box().Children[0].text == "lighten\xad‐"
//     assert line3.Box().Children[0].text == "lighten\xad‐"
//     assert line4.Box().Children[0].text == "in"

// func TestHyphenateLimitZone1(t *testing.T){
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     page, = renderPages(
//         "<html style="width: 12em; font-family: weasyprint">"
//         "<style>"
//         "  @font-face {src: url(weasyprint.otf); font-family: weasyprint}"
//         "</style>"
//         "<body style="hyphens: auto;"
//         "hyphenate-limit-zone: 0" lang=fr>"
//         "mmmmm hyphénation")
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     lines = body.Box().Children
//     assert len(lines) == 2
//     assert lines[0].Box().Children[0].text.endswith("‐")
//     fullText = "".join(line.Box().Children[0].text for line := range lines)
//     assert fullText.replace("‐", "") == "mmmmm hyphénation"

// func TestHyphenateLimitZone2(t *testing.T){
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     page, = renderPages(
//         "<html style="width: 12em; font-family: weasyprint">"
//         "<style>"
//         "  @font-face {src: url(weasyprint.otf); font-family: weasyprint}"
//         "</style>"
//         "<body style="hyphens: auto;"
//         "hyphenate-limit-zone: 9em" lang=fr>"
//         "mmmmm hyphénation")
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     lines = body.Box().Children
//     assert len(lines) > 1
//     assert lines[0].Box().Children[0].text.endswith("mm")
//     fullText = "".join(line.Box().Children[0].text for line := range lines)
//     assert fullText == "mmmmmhyphénation"

// func TestHyphenateLimitZone3(t *testing.T){
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     page, = renderPages(
//         "<html style="width: 12em; font-family: weasyprint">"
//         "<style>"
//         "  @font-face {src: url(weasyprint.otf); font-family: weasyprint}"
//         "</style>"
//         "<body style="hyphens: auto;"
//         "hyphenate-limit-zone: 5%" lang=fr>"
//         "mmmmm hyphénation")
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     lines = body.Box().Children
//     assert len(lines) == 2
//     assert lines[0].Box().Children[0].text.endswith("‐")
//     fullText = "".join(line.Box().Children[0].text for line := range lines)
//     assert fullText.replace("‐", "") == "mmmmm hyphénation"

// func TestHyphenateLimitZone4(t *testing.T){
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     page, = renderPages(
//         "<html style="width: 12em; font-family: weasyprint">"
//         "<style>"
//         "  @font-face {src: url(weasyprint.otf); font-family: weasyprint}"
//         "</style>"
//         "<body style="hyphens: auto;"
//         "hyphenate-limit-zone: 95%" lang=fr>"
//         "mmmmm hyphénation")
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     lines = body.Box().Children
//     assert len(lines) > 1
//     assert lines[0].Box().Children[0].text.endswith("mm")
//     fullText = "".join(line.Box().Children[0].text for line := range lines)
//     assert fullText == "mmmmmhyphénation"

// @assertNoLogs
// @pytest.mark.parametrize("css, result", (
//     ("auto", 2),
//     ("auto auto 0", 2),
//     ("0 0 0", 2),
//     ("4 4 auto", 1),
//     ("6 2 4", 2),
//     ("auto 1 auto", 2),
//     ("7 auto auto", 1),
//     ("6 auto auto", 2),
//     ("5 2", 2),
//     ("3", 2),
//     ("2 4 6", 1),
//     ("auto 4", 1),
//     ("auto 2", 2),
// ))
// func TestHyphenateLimitChars(t *testing.Tcss, result):
//     page, = renderPages(
//         "<html style="width: 1em; font-family: weasyprint">"
//         "<style>"
//         "  @font-face {src: url(weasyprint.otf); font-family: weasyprint}"
//         "</style>"
//         "<body style="hyphens: auto;"
//         f"hyphenate-limit-chars: {css}" lang=en>"
//         "hyphen")
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     lines = body.Box().Children
//     assert len(lines) == result

// @assertNoLogs
// @pytest.mark.parametrize("css", (
//     // light·en
//     "3 3 3",  // "en" is shorter than 3
//     "3 6 2",  // "light" is shorter than 6
//     "8",  // "lighten" is shorter than 8
// ))
// func TestHyphenateLimitCharsPunctuation(t *testing.Tcss):
//     // See https://github.com/Kozea/WeasyPrint/issues/109
//     page, = renderPages(
//         "<html style="width: 1em; font-family: weasyprint">"
//         "<style>"
//         "  @font-face {src: url(weasyprint.otf); font-family: weasyprint}"
//         "</style>"
//         "<body style="hyphens: auto;"
//         f"hyphenate-limit-chars: {css}" lang=en>"
//         "..lighten..")
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     lines = body.Box().Children
//     assert len(lines) == 1

// @assertNoLogs
// @pytest.mark.parametrize("wrap, text, test, fullText", (
//     ("break-word", "aaaaaaaa", lambda a: a > 1, "aaaaaaaa"),
//     ("normal", "aaaaaaaa", lambda a: a == 1, "aaaaaaaa"),
//     ("break-word", "hyphenations", lambda a: a > 3,
//      "hy\u2010phen\u2010ations"),
//     ("break-word", "A splitted word.  An hyphenated word.",
//      lambda a: a > 8, "Asplittedword.Anhy\u2010phen\u2010atedword."),
// ))
// func TestOverflowWrap(t *testing.Twrap, text, test, fullText):
//     page := renderOnePage(t, `
//       <style>
//         @font-face {src: url(weasyprint.otf); font-family: weasyprint}
//         body {width: 80px; overflow: hidden; font-family: weasyprint; }
//         span {overflow-wrap: %s; white-space: normal; }
//       </style>
//       <body style="hyphens: auto;" lang="en">
//         <span>%s
//     ` % (wrap, text))
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     lines = []
//     for line := range body.Box().Children {
//         box, = line.Box().Children
//         textBox, = box.Box().Children
//         lines.append(textBox.text)
//     } linesFullText = "".join(line for line := range lines)
//     assert test(len(lines))
//     assert fullText == linesFullText
// }

// func whiteSpaceLines(width, space) {
//     page := renderOnePage(t, `
//       <style>
//         body { font-size: 100px; width: %dpx }
//         span { white-space: %s }
//       </style>
//       <body><span>This +    \n    is text` % (width, space))
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     return body.Box().Children
// }

// func TestWhiteSpace1(t *testing.T) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     line1, line2, line3, line4 = whiteSpaceLines(1, "normal")
//     box1, = line1.Box().Children
//     text1, = box1.Box().Children
//     assert text1.text == "This"
//     box2, = line2.Box().Children
//     text2, = box2.Box().Children
//     assert text2.text == "+"
//     box3, = line3.Box().Children
//     text3, = box3.Box().Children
//     assert text3.text == "is"
//     box4, = line4.Box().Children
//     text4, = box4.Box().Children
//     assert text4.text == "text"
// }

// func TestWhiteSpace2(t *testing.T) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     line1, line2 = whiteSpaceLines(1, "pre")
//     box1, = line1.Box().Children
//     text1, = box1.Box().Children
//     assert text1.text == "This +    "
//     box2, = line2.Box().Children
//     text2, = box2.Box().Children
//     assert text2.text == "    is text"
// }

// func TestWhiteSpace3(t *testing.T) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     line1, = whiteSpaceLines(1, "nowrap")
//     box1, = line1.Box().Children
//     text1, = box1.Box().Children
//     assert text1.text == "This + is text"
// }

// func TestWhiteSpace4(t *testing.T) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     line1, line2, line3, line4, line5 = whiteSpaceLines(1, "pre-wrap")
//     box1, = line1.Box().Children
//     text1, = box1.Box().Children
//     assert text1.text == "This "
//     box2, = line2.Box().Children
//     text2, = box2.Box().Children
//     assert text2.text == "+    "
//     box3, = line3.Box().Children
//     text3, = box3.Box().Children
//     assert text3.text == "    "
//     box4, = line4.Box().Children
//     text4, = box4.Box().Children
//     assert text4.text == "is "
//     box5, = line5.Box().Children
//     text5, = box5.Box().Children
//     assert text5.text == "text"
// }

// func TestWhiteSpace5(t *testing.T) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     line1, line2, line3, line4 = whiteSpaceLines(1, "pre-line")
//     box1, = line1.Box().Children
//     text1, = box1.Box().Children
//     assert text1.text == "This"
//     box2, = line2.Box().Children
//     text2, = box2.Box().Children
//     assert text2.text == "+"
//     box3, = line3.Box().Children
//     text3, = box3.Box().Children
//     assert text3.text == "is"
//     box4, = line4.Box().Children
//     text4, = box4.Box().Children
//     assert text4.text == "text"
// }

// func TestWhiteSpace6(t *testing.T) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     line1, = whiteSpaceLines(1000000, "normal")
//     box1, = line1.Box().Children
//     text1, = box1.Box().Children
//     assert text1.text == "This + is text"
// }

// func TestWhiteSpace7(t *testing.T) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     line1, line2 = whiteSpaceLines(1000000, "pre")
//     box1, = line1.Box().Children
//     text1, = box1.Box().Children
//     assert text1.text == "This +    "
//     box2, = line2.Box().Children
//     text2, = box2.Box().Children
//     assert text2.text == "    is text"
// }

// func TestWhiteSpace8(t *testing.T) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     line1, = whiteSpaceLines(1000000, "nowrap")
//     box1, = line1.Box().Children
//     text1, = box1.Box().Children
//     assert text1.text == "This + is text"
// }

// func TestWhiteSpace9(t *testing.T) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     line1, line2 = whiteSpaceLines(1000000, "pre-wrap")
//     box1, = line1.Box().Children
//     text1, = box1.Box().Children
//     assert text1.text == "This +    "
//     box2, = line2.Box().Children
//     text2, = box2.Box().Children
//     assert text2.text == "    is text"
// }

// func TestWhiteSpace10(t *testing.T) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     line1, line2 = whiteSpaceLines(1000000, "pre-line")
//     box1, = line1.Box().Children
//     text1, = box1.Box().Children
//     assert text1.text == "This +"
//     box2, = line2.Box().Children
//     text2, = box2.Box().Children
//     assert text2.text == "is text"
// }

// func TestWhiteSpace11(t *testing.T) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     // Test regression: https://github.com/Kozea/WeasyPrint/issues/813
//     page := renderOnePage(t, `
//       <style>
//         pre { width: 0 }
//       </style>
//       <body><pre>This<br/>is text`)
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     pre, = body.Box().Children
//     line1, line2 = pre.Box().Children
//     text1, box = line1.Box().Children
//     assert text1.text == "This"
//     assert box.elementTag == "br"
//     text2, = line2.Box().Children
//     assert text2.text == "is text"
// }

// func TestWhiteSpace12(t *testing.T) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     // Test regression: https://github.com/Kozea/WeasyPrint/issues/813
//     page := renderOnePage(t, `
//       <style>
//         pre { width: 0 }
//       </style>
//       <body><pre>This is <span>lol</span> text`)
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     pre, = body.Box().Children
//     line1, = pre.Box().Children
//     text1, span, text2 = line1.Box().Children
//     assert text1.text == "This is "
//     assert span.elementTag == "span"
//     assert text2.text == " text"
// }

// @assertNoLogs
// @pytest.mark.parametrize("value, width", (
//     (8, 144),  // (2 + (8 - 1)) * 16
//     (4, 80),  // (2 + (4 - 1)) * 16
//     ("3em", 64),  // (2 + (3 - 1)) * 16
//     ("25px", 41),  // 2 * 16 + 25 - 1 * 16
//     // (0, 32),  // See Layout.setTabs
// ))
// func TestTabSize(t *testing.Tvalue, width) {
//     page := renderOnePage(t, `
//       <style>
//         @font-face {src: url(weasyprint.otf); font-family: weasyprint}
//         pre { tab-size: %s; font-family: weasyprint }
//       </style>
//       <pre>a&#9;a</pre>
//     ` % value)
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     paragraph := body.Box().Children[0]
//     line := paragraph.Box().Children0
//     assert line.width == width

// func TestTextTransform(t *testing.T){
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     page := renderOnePage(t, `
//       <style>
//         p { text-transform: capitalize }
//         p+p { text-transform: uppercase }
//         p+p+p { text-transform: lowercase }
//         p+p+p+p { text-transform: full-width }
//         p+p+p+p+p { text-transform: none }
//       </style>
// <p>hé lO1</p><p>hé lO1</p><p>hé lO1</p><p>hé lO1</p><p>hé lO1</p>
//     `)
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     p1, p2, p3, p4, p5 = body.Box().Children
//     line1, = p1.Box().Children
//     text1, = line1.Box().Children
//     assert text1.text == "Hé Lo1"
//     line2, = p2.Box().Children
//     text2, = line2.Box().Children
//     assert text2.text == "HÉ LO1"
//     line3, = p3.Box().Children
//     text3, = line3.Box().Children
//     assert text3.text == "hé lo1"
//     line4, = p4.Box().Children
//     text4, = line4.Box().Children
//     assert text4.text == "\uff48é\u3000\uff4c\uff2f\uff11"
//     line5, = p5.Box().Children
//     text5, = line5.Box().Children
//     assert text5.text == "hé lO1"
// }

// func TestTextFloatingPreLine(t *testing.T) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     // Test regression: https://github.com/Kozea/WeasyPrint/issues/610
//     page := renderOnePage(t, `
//       <div style="float: left; white-space: pre-line">This is
//       oh this end </div>
//     `)

// @assertNoLogs
// @pytest.mark.parametrize(
//     "leader, content", (
//         ("dotted", "."),
//         ("solid", ""),
//         ("space", " "),
//         ("" .-"", " .-"),
//     )
// )
// func TestLeaderContent(t *testing.Tleader, content):
//     page := renderOnePage(t, `
//       <style>div::after { content: leader(%s) }</style>
//       <div></div>
//     ` % leader)
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     div, = body.Box().Children
//     line, = div.Box().Children
//     after, = line.Box().Children
//     inline, = after.Box().Children
//     assert inline.Box().Children[0].text == content
// }

// @pytest.mark.xfail
// func TestMaxLines(t *testing.T) {
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     page := renderOnePage(t, `
//       <style>
//         @page {size: 10px 10px;}
//         @font-face {src: url(weasyprint.otf); font-family: weasyprint}
//         p {
//           font-family: weasyprint;
//           font-size: 2px;
//           max-lines: 2;
//         }
//       </style>
//       <p>
//         abcd efgh ijkl
//       </p>
//     `)
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     p1, p2 = body.Box().Children
//     line1, line2 = p1.Box().Children
//     line3, = p2.Box().Children
//     text1, = line1.Box().Children
//     text2, = line2.Box().Children
//     text3, = line3.Box().Children
//     assert text1.text == "abcd"
//     assert text2.text == "efgh"
//     assert text3.text == "ijkl"

// func TestContinue(t *testing.T){
// cp := testutils.CaptureLogs()
// cp.AssertNoLogs(t)
//     page := renderOnePage(t, `
//       <style>
//         @page {size: 10px 4px;}
//         @font-face {src: url(weasyprint.otf); font-family: weasyprint}
//         div {
//           continue: discard;
//           font-family: weasyprint;
//           font-size: 2px;
//         }
//       </style>
//       <div>
//         abcd efgh ijkl
//       </div>
//     `)
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     p, = body.Box().Children
//     line1, line2 = p.Box().Children
//     text1, = line1.Box().Children
//     text2, = line2.Box().Children
//     assert text1.text == "abcd"
//     assert text2.text == "efgh"
