package document

import (
	"fmt"
	"reflect"
	"strings"
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

func assertEqual(t *testing.T, got, exp interface{}, context string) {
	if !reflect.DeepEqual(exp, got) {
		t.Fatalf("%s: expected %v, got  %v", context, exp, got)
	}
}

// unpack 3 chilren
func unpack3(box Box) (c1, c2, c3 Box) {
	return box.Box().Children[0], box.Box().Children[1], box.Box().Children[2]
}

// unpack 4 chilren
func unpack4(box Box) (c1, c2, c3, c4 Box) {
	return box.Box().Children[0], box.Box().Children[1], box.Box().Children[2], box.Box().Children[3]
}

// unpack 5 chilren
func unpack5(box Box) (c1, c2, c3, c4, c5 Box) {
	return box.Box().Children[0], box.Box().Children[1], box.Box().Children[2], box.Box().Children[3], box.Box().Children[4]
}

func TestTextFontSizeZero(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
	assertEqual(t, len(line.Box().Children), 0, "children")
	assertEqual(t, line.Box().Height, pr.Float(0), "line")
	assertEqual(t, paragraph.Box().Height, pr.Float(0), "paragraph")
}

func TestTextSpacedInlines(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
		<p>start <i><b>bi1</b> <b>bi2</b></i> <b>b1</b> end</p>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line := paragraph.Box().Children[0]
	start, i, space, b, end := unpack5(line)

	assertEqual(t, start.(*bo.TextBox).Text, "start ", "start")
	assertEqual(t, space.(*bo.TextBox).Text, " ", "space")
	assertEqual(t, end.(*bo.TextBox).Text, " end", "end")
	if w := space.Box().Width.V(); w <= 0 {
		t.Fatalf("expected positive width, got %f", w)
	}

	bi1, space, bi2 := unpack3(i)
	bi1 = bi1.Box().Children[0]
	bi2 = bi2.Box().Children[0]
	assertEqual(t, bi1.(*bo.TextBox).Text, "bi1", "bi1")
	assertEqual(t, space.(*bo.TextBox).Text, " ", "space")
	assertEqual(t, bi2.(*bo.TextBox).Text, "bi2", "bi2")
	if w := space.Box().Width.V(); w <= 0 {
		t.Fatalf("expected positive width, got %f", w)
	}

	b1 := b.Box().Children[0]
	assertEqual(t, b1.(*bo.TextBox).Text, "b1", "b1")
}

func TestTextAlignLeft(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)
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
	assertEqual(t, img1.Box().PositionX, pr.Float(0), "img1")
	assertEqual(t, img2.Box().PositionX, pr.Float(40), "img2")
}

func TestTextAlignRight(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)
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

	assertEqual(t, img1.Box().PositionX, pr.Float(100), "img1") // 200 - 60 - 40
	assertEqual(t, img2.Box().PositionX, pr.Float(140), "img2") // 200 - 60
}

func TestTextAlignCenter(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

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

	assertEqual(t, img1.Box().PositionX, pr.Float(50), "img1")
	assertEqual(t, img2.Box().PositionX, pr.Float(90), "img2")
}

func TestTextAlignJustify(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

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
	image1, space1, strong := unpack3(line1)
	image2, space2, image3, space3, image4 := unpack5(strong)
	image5 := line2.Box().Children[0]
	assertEqual(t, space1.(*bo.TextBox).Text, " ", "space1")
	assertEqual(t, space2.(*bo.TextBox).Text, " ", "space2")
	assertEqual(t, space3.(*bo.TextBox).Text, " ", "space3")

	assertEqual(t, image1.Box().PositionX, pr.Float(0), "image1")
	assertEqual(t, space1.Box().PositionX, pr.Float(40), "space1")
	assertEqual(t, strong.Box().PositionX, pr.Float(70), "strong")
	assertEqual(t, image2.Box().PositionX, pr.Float(70), "image2")
	assertEqual(t, space2.Box().PositionX, pr.Float(130), "space2")
	assertEqual(t, image3.Box().PositionX, pr.Float(160), "image3")
	assertEqual(t, space3.Box().PositionX, pr.Float(170), "space3")
	assertEqual(t, image4.Box().PositionX, pr.Float(200), "image4")
	assertEqual(t, strong.Box().Width.V(), pr.Float(230), "strong")
	assertEqual(t, image5.Box().PositionX, pr.Float(0), "image5")
}

func TestTextAlignJustifyAll(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @page { size: 300px 1000px }
        body { text-align: justify-all }
      </style>
      <p><img src="pattern.png" style="width: 40px">
        <strong>
          <img src="pattern.png" style="width: 60px">
          <img src="pattern.png" style="width: 10px">
          <img src="pattern.png" style="width: 100px"
        ></strong><img src="pattern.png" style="width: 200px">
        <img src="pattern.png" style="width: 10px">`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line1, line2 := paragraph.Box().Children[0], paragraph.Box().Children[1]
	image1, space1, strong := unpack3(line1)
	image2, space2, image3, space3, image4 := unpack5(strong)
	image5, space4, image6 := unpack3(line2)
	assertEqual(t, space1.(*bo.TextBox).Text, " ", "space1")
	assertEqual(t, space2.(*bo.TextBox).Text, " ", "space2")
	assertEqual(t, space3.(*bo.TextBox).Text, " ", "space3")
	assertEqual(t, space4.(*bo.TextBox).Text, " ", "space4")

	assertEqual(t, image1.Box().PositionX, pr.Float(0), "image1")
	assertEqual(t, space1.Box().PositionX, pr.Float(40), "space1")
	assertEqual(t, strong.Box().PositionX, pr.Float(70), "strong")
	assertEqual(t, image2.Box().PositionX, pr.Float(70), "image2")
	assertEqual(t, space2.Box().PositionX, pr.Float(130), "space2")
	assertEqual(t, image3.Box().PositionX, pr.Float(160), "image3")
	assertEqual(t, space3.Box().PositionX, pr.Float(170), "space3")
	assertEqual(t, image4.Box().PositionX, pr.Float(200), "image4")
	assertEqual(t, strong.Box().Width, pr.Float(230), "strong")

	assertEqual(t, image5.Box().PositionX, pr.Float(0), "image5")
	assertEqual(t, space4.Box().PositionX, pr.Float(200), "space4")
	assertEqual(t, image6.Box().PositionX, pr.Float(290), "image6")
}

func TestTextAlignAllLast(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @page { size: 300px 1000px }
        body { text-align-all: justify; text-align-last: right }
      </style>
      <p><img src="pattern.png" style="width: 40px">
        <strong>
          <img src="pattern.png" style="width: 60px">
          <img src="pattern.png" style="width: 10px">
          <img src="pattern.png" style="width: 100px"
        ></strong><img src="pattern.png" style="width: 200px"
        ><img src="pattern.png" style="width: 10px">`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line1, line2 := paragraph.Box().Children[0], paragraph.Box().Children[1]
	image1, space1, strong := unpack3(line1)
	image2, space2, image3, space3, image4 := unpack5(strong)
	image5, image6 := line2.Box().Children[0], line2.Box().Children[1]

	assertEqual(t, space1.(*bo.TextBox).Text, " ", "space1")
	assertEqual(t, space2.(*bo.TextBox).Text, " ", "space2")
	assertEqual(t, space3.(*bo.TextBox).Text, " ", "space3")

	assertEqual(t, image1.Box().PositionX, pr.Float(0), "image1")
	assertEqual(t, space1.Box().PositionX, pr.Float(40), "space1")
	assertEqual(t, strong.Box().PositionX, pr.Float(70), "strong")
	assertEqual(t, image2.Box().PositionX, pr.Float(70), "image2")
	assertEqual(t, space2.Box().PositionX, pr.Float(130), "space2")
	assertEqual(t, image3.Box().PositionX, pr.Float(160), "image3")
	assertEqual(t, space3.Box().PositionX, pr.Float(170), "space3")
	assertEqual(t, image4.Box().PositionX, pr.Float(200), "image4")
	assertEqual(t, strong.Box().Width, pr.Float(230), "strong")

	assertEqual(t, image5.Box().PositionX, pr.Float(90), "image5")
	assertEqual(t, image6.Box().PositionX, pr.Float(290), "image6")
}

func TestTextAlignNotEnoughSpace(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        p { text-align: center; width: 0 }
        span { display: inline-block }
      </style>
      <p><span>aaaaaaaaaaaaaaaaaaaaaaaaaa</span></p>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	span := paragraph.Box().Children[0]
	assertEqual(t, span.Box().PositionX, pr.Float(0), "span")
}

func TestTextAlignJustifyNoSpace(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// single-word line (zero spaces)
	page := renderOnePage(t, `
      <style>
        body { text-align: justify; width: 50px }
      </style>
      <p>Supercalifragilisticexpialidocious bar</p>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line1, _ := paragraph.Box().Children[0], paragraph.Box().Children[1]
	text := line1.Box().Children[0]
	assertEqual(t, text.Box().PositionX, pr.Float(0), "text")
}

func TestTextAlignJustifyTextIndent(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)
	// text-indent
	page := renderOnePage(t, `
      <style>
        @page { size: 300px 1000px }
        body { text-align: justify }
        p { text-indent: 3px }
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
	image1, space1, strong := unpack3(line1)
	image2, space2, image3, space3, image4 := unpack5(strong)
	image5 := line2.Box().Children[0]

	assertEqual(t, space1.(*bo.TextBox).Text, " ", "space1")
	assertEqual(t, space2.(*bo.TextBox).Text, " ", "space2")
	assertEqual(t, space3.(*bo.TextBox).Text, " ", "space3")

	assertEqual(t, image1.Box().PositionX, pr.Float(3), "image1")
	assertEqual(t, space1.Box().PositionX, pr.Float(43), "space1")
	assertEqual(t, strong.Box().PositionX, pr.Float(72), "strong")
	assertEqual(t, image2.Box().PositionX, pr.Float(72), "image2")
	assertEqual(t, space2.Box().PositionX, pr.Float(132), "space2")
	assertEqual(t, image3.Box().PositionX, pr.Float(161), "image3")
	assertEqual(t, space3.Box().PositionX, pr.Float(171), "space3")
	assertEqual(t, image4.Box().PositionX, pr.Float(200), "image4")
	assertEqual(t, strong.Box().Width, pr.Float(228), "strong")

	assertEqual(t, image5.Box().PositionX, pr.Float(0), "image5")
}

func TestTextAlignJustifyNoBreakBetweenChildren(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Test justification when line break happens between two inline children
	// that must stay together.
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/637
	page := renderOnePage(t, `
      <style>
        @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        p { text-align: justify; font-family: weasyprint; width: 7em }
      </style>
      <p>
        <span>a</span>
        <span>b</span>
        <span>bla</span><span>,</span>
        <span>b</span>
      </p>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line1, line2 := paragraph.Box().Children[0], paragraph.Box().Children[1]
	span1, _, span2, _ := unpack4(line1)
	assertEqual(t, span1.Box().PositionX, pr.Float(0), "span1")
	assertEqual(t, span2.Box().PositionX, pr.Float(6*16), "span2") // 1 character + 5 spaces
	assertEqual(t, line1.Box().Width, pr.Float(7*16), "line1")     // 7em

	span1, span2, _, span3, _ := unpack5(line2)
	assertEqual(t, span1.Box().PositionX, pr.Float(0), "span1")
	assertEqual(t, span2.Box().PositionX, pr.Float(3*16), "span2") // 3 characters
	assertEqual(t, span3.Box().PositionX, pr.Float(5*16), "span3") // (3 + 1) characters + 1 space
}

func TestWordSpacing(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)
	// keep the empty <style> as a regression test: element.text is nil
	// (Not a string.)
	page := renderOnePage(t, `
      <style></style>
      <body><strong>Lorem ipsum dolor<em>sit amet</em></strong>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	strong1 := line.Box().Children[0]

	page = renderOnePage(t, `
      <style>strong { word-spacing: 11px }</style>
      <body><strong>Lorem ipsum dolor<em>sit amet</em></strong>`)
	html = page.Box().Children[0]
	body = html.Box().Children[0]
	line = body.Box().Children[0]
	strong2 := line.Box().Children[0]
	assertEqual(t, strong2.Box().Width.V()-strong1.Box().Width.V(), pr.Float(33), "strong distance")
}

func TestLetterSpacing1(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
        <body><strong>Supercalifragilisticexpialidocious</strong>`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	strong1 := line.Box().Children[0]

	page = renderOnePage(t, `
        <style>strong { letter-spacing: 11px }</style>
        <body><strong>Supercalifragilisticexpialidocious</strong>`)
	html = page.Box().Children[0]
	body = html.Box().Children[0]
	line = body.Box().Children[0]
	strong2 := line.Box().Children[0]
	assertEqual(t, strong2.Box().Width.V()-strong1.Box().Width.V(), pr.Float(34*11), "strong distance")

	// an embedded tag should ! affect the single-line letter spacing
	page = renderOnePage(t,
		"<style>strong { letter-spacing: 11px }</style>"+
			"<body><strong>Supercali<span>fragilistic</span>expialidocious"+
			"</strong>")
	html = page.Box().Children[0]
	body = html.Box().Children[0]
	line = body.Box().Children[0]
	strong3 := line.Box().Children[0]
	assertEqual(t, strong3.Box().Width, strong2.Box().Width, "strong")

	// duplicate wrapped lines should also have same overall width
	// Note work-around for word-wrap bug (issue #163) by marking word
	// as an inline-block
	page = renderOnePage(t, fmt.Sprintf(`<style>
          strong {
            letter-spacing: 11px;
            max-width: %fpx
        }
          span { display: inline-block }
        </style>
        <body><strong>
          <span>Supercali<i>fragilistic</i>expialidocious</span> 
          <span>Supercali<i>fragilistic</i>expialidocious</span>
        </strong>`, strong3.Box().Width.V()*1.5))
	html = page.Box().Children[0]
	body = html.Box().Children[0]
	line1, line2 := body.Box().Children[0], body.Box().Children[1]
	assertEqual(t, line1.Box().Children[0].Box().Width, line2.Box().Children[0].Box().Width, "")
	assertEqual(t, line1.Box().Children[0].Box().Width, strong2.Box().Width, "")
}

func TestSpacingEx(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Test regression on ex units in spacing properties
	for _, spacing := range []string{"word-spacing", "letter-spacing"} {
		renderPages(t, fmt.Sprintf(`<div style="%s: 2ex">abc def`, spacing))
	}
}

func TestTextIndent(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	for _, indent := range []string{"12px", "6%"} {
		page := renderOnePage(t, fmt.Sprintf(`
        <style>
            @page { size: 220px }
            body { margin: 10px; text-indent: %s }
        </style>
        <p>Some text that is long enough that it take at least three line,
           but maybe more.
    `, indent))
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		paragraph := body.Box().Children[0]
		lines := paragraph.Box().Children
		text1 := lines[0].Box().Children[0]
		text2 := lines[1].Box().Children[0]
		text3 := lines[2].Box().Children[0]
		assertEqual(t, text1.Box().PositionX, pr.Float(22), "text1") // 10px margin-left + 12px indent
		assertEqual(t, text2.Box().PositionX, pr.Float(10), "text2") // No indent
		assertEqual(t, text3.Box().PositionX, pr.Float(10), "text3") // No indent
	}
}

func TestTextIndentInline(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/1000
	page := renderOnePage(t, `
        <style>
            @font-face { src: url(weasyprint.otf); font-family: weasyprint }
            p { display: inline-block; text-indent: 1em;
                font-family: weasyprint }
        </style>
        <p><span>text
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	paragraph := body.Box().Children[0]
	line := paragraph.Box().Children[0]
	assertEqual(t, line.Box().Width, pr.Float((4+1)*16), "")
}

func TestTextIndentMultipage(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/706

	for _, indent := range []string{"12px", "6%"} {
		pages := renderPages(t, fmt.Sprintf(`
        <style>
            @page { size: 220px 1.5em; margin: 0 }
            body { margin: 10px; text-indent: %s }
        </style>
        <p>Some text that is long enough that it take at least three line,
           but maybe more.
    `, indent))
		page := pages[0]
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		paragraph := body.Box().Children[0]
		line := paragraph.Box().Children[0]
		text := line.Box().Children[0]
		assertEqual(t, text.Box().PositionX, pr.Float(22), "") // 10px margin-left + 12px indent

		page = pages[1]
		html = page.Box().Children[0]
		body = html.Box().Children[0]
		paragraph = body.Box().Children[0]
		line = paragraph.Box().Children[0]
		text = line.Box().Children[0]
		assertEqual(t, text.Box().PositionX, pr.Float(10), "") // No indent
	}
}

func TestHyphenateCharacter1(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderOnePage(t, `
        <html style="width: 5em; font-family: weasyprint">
        <style>
          @font-face {src: url(weasyprint.otf); font-family: weasyprint}
        </style>
        <body style="hyphens: auto;  hyphenate-character: '!'" lang=fr>
        hyphénation`)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	lines := body.Box().Children
	if !(len(lines) > 1) {
		t.Fatalf("expected > 1, got %v", lines)
	}
	if text := lines[0].Box().Children[0].(*bo.TextBox).Text; !strings.HasSuffix(text, "!") {
		t.Fatalf("unexpected %s", text)
	}
	fullText := ""
	for _, line := range lines {
		fullText += line.Box().Children[0].(*bo.TextBox).Text
	}
	assertEqual(t, strings.ReplaceAll(fullText, "!", ""), "hyphénation", "")
}

// func TestHyphenateCharacter2(t *testing.T){
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
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
//     assertEqual(t, len(lines) > 1
//     assertEqual(t, lines[0].Box().Children[0].text.endswith("à")
//     fullText = "".join(line.Box().Children[0].text for line := range lines)
//     assertEqual(t, fullText.replace("à", "") == "hyphénation"

// func TestHyphenateCharacter3(t *testing.T){
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
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
//     assertEqual(t, len(lines) > 1
//     assertEqual(t, lines[0].Box().Children[0].text.endswith("ù ù")
//     fullText = "".join(line.Box().Children[0].text for line := range lines)
//     assertEqual(t, fullText.replace(" ", "").replace("ù", "") == "hyphénation"

// func TestHyphenateCharacter4(t *testing.T){
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
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
//     assertEqual(t, len(lines) > 1
//     fullText = "".join(line.Box().Children[0].text for line := range lines)
//     assertEqual(t, fullText == "hyphénation"

// func TestHyphenateCharacter5(t *testing.T){
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
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
//     assertEqual(t, len(lines) > 1
//     assertEqual(t, lines[0].Box().Children[0].text.endswith("———")
//     fullText = "".join(line.Box().Children[0].text for line := range lines)
//     assertEqual(t, fullText.replace("—", "") == "hyphénation"

// func TestHyphenateManual1(t *testing.T){
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
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
//             assertEqual(t, len(lines) == 2
//             assertEqual(t, lines[0].Box().Children[0].text.endswith(hyphenateCharacter)
//             fullText = "".join(
//                 child.text for line := range lines for child := range line.Box().Children)
//             assertEqual(t, fullText.replace(hyphenateCharacter, "") == word

// func TestHyphenateManual2(t *testing.T){
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
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
//             assertEqual(t, len(lines) := range (2, 3)
//             fullText = "".join(
//                 child.text for line := range lines for child := range line.Box().Children)
//             fullText = fullText.replace(hyphenateCharacter, "")
//             if lines[0].Box().Children[0].text.endswith(hyphenateCharacter):
//                 assertEqual(t, fullText == word
//             else:
//                 assertEqual(t, lines[0].Box().Children[0].text.endswith("y")
//                 if len(lines) == 3:
//                     assertEqual(t, lines[1].Box().Children[0].text.endswith(
//                         hyphenateCharacter)

// func TestHyphenateManual3(t *testing.T){
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
//     // Automatic hyphenation opportunities within a word must be ignored if the
//     // word contains a conditional hyphen, := range favor of the conditional
//     // hyphen(s).
//     page, = renderPages(
//         "<html style="width: 0.1em" lang="en">"
//         "<body style="hyphens: auto">in&shy;lighten&shy;lighten&shy;in")
//     html := page.Box().Children[0]
//     body := html.Box().Children[0]
//     line1, line2, line3, line4 = body.Box().Children
//     assertEqual(t, line1.Box().Children[0].text == "in\xad‐"
//     assertEqual(t, line2.Box().Children[0].text == "lighten\xad‐"
//     assertEqual(t, line3.Box().Children[0].text == "lighten\xad‐"
//     assertEqual(t, line4.Box().Children[0].text == "in"

// func TestHyphenateLimitZone1(t *testing.T){
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
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
//     assertEqual(t, len(lines) == 2
//     assertEqual(t, lines[0].Box().Children[0].text.endswith("‐")
//     fullText = "".join(line.Box().Children[0].text for line := range lines)
//     assertEqual(t, fullText.replace("‐", "") == "mmmmm hyphénation"

// func TestHyphenateLimitZone2(t *testing.T){
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
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
//     assertEqual(t, len(lines) > 1
//     assertEqual(t, lines[0].Box().Children[0].text.endswith("mm")
//     fullText = "".join(line.Box().Children[0].text for line := range lines)
//     assertEqual(t, fullText == "mmmmmhyphénation"

// func TestHyphenateLimitZone3(t *testing.T){
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
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
//     assertEqual(t, len(lines) == 2
//     assertEqual(t, lines[0].Box().Children[0].text.endswith("‐")
//     fullText = "".join(line.Box().Children[0].text for line := range lines)
//     assertEqual(t, fullText.replace("‐", "") == "mmmmm hyphénation"

// func TestHyphenateLimitZone4(t *testing.T){
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
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
//     assertEqual(t, len(lines) > 1
//     assertEqual(t, lines[0].Box().Children[0].text.endswith("mm")
//     fullText = "".join(line.Box().Children[0].text for line := range lines)
//     assertEqual(t, fullText == "mmmmmhyphénation"

// @assertEqual(t,NoLogs
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
//     assertEqual(t, len(lines) == result

// @assertEqual(t,NoLogs
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
//     assertEqual(t, len(lines) == 1

// @assertEqual(t,NoLogs
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
//     assertEqual(t, test(len(lines))
//     assertEqual(t, fullText == linesFullText
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
// defer cp.AssertNoLogs(t)
//     line1, line2, line3, line4 = whiteSpaceLines(1, "normal")
//     box1, = line1.Box().Children
//     text1, = box1.Box().Children
//     assertEqual(t, text1.text == "This"
//     box2, = line2.Box().Children
//     text2, = box2.Box().Children
//     assertEqual(t, text2.text == "+"
//     box3, = line3.Box().Children
//     text3, = box3.Box().Children
//     assertEqual(t, text3.text == "is"
//     box4, = line4.Box().Children
//     text4, = box4.Box().Children
//     assertEqual(t, text4.text == "text"
// }

// func TestWhiteSpace2(t *testing.T) {
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
//     line1, line2 = whiteSpaceLines(1, "pre")
//     box1, = line1.Box().Children
//     text1, = box1.Box().Children
//     assertEqual(t, text1.text == "This +    "
//     box2, = line2.Box().Children
//     text2, = box2.Box().Children
//     assertEqual(t, text2.text == "    is text"
// }

// func TestWhiteSpace3(t *testing.T) {
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
//     line1, = whiteSpaceLines(1, "nowrap")
//     box1, = line1.Box().Children
//     text1, = box1.Box().Children
//     assertEqual(t, text1.text == "This + is text"
// }

// func TestWhiteSpace4(t *testing.T) {
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
//     line1, line2, line3, line4, line5 = whiteSpaceLines(1, "pre-wrap")
//     box1, = line1.Box().Children
//     text1, = box1.Box().Children
//     assertEqual(t, text1.text == "This "
//     box2, = line2.Box().Children
//     text2, = box2.Box().Children
//     assertEqual(t, text2.text == "+    "
//     box3, = line3.Box().Children
//     text3, = box3.Box().Children
//     assertEqual(t, text3.text == "    "
//     box4, = line4.Box().Children
//     text4, = box4.Box().Children
//     assertEqual(t, text4.text == "is "
//     box5, = line5.Box().Children
//     text5, = box5.Box().Children
//     assertEqual(t, text5.text == "text"
// }

// func TestWhiteSpace5(t *testing.T) {
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
//     line1, line2, line3, line4 = whiteSpaceLines(1, "pre-line")
//     box1, = line1.Box().Children
//     text1, = box1.Box().Children
//     assertEqual(t, text1.text == "This"
//     box2, = line2.Box().Children
//     text2, = box2.Box().Children
//     assertEqual(t, text2.text == "+"
//     box3, = line3.Box().Children
//     text3, = box3.Box().Children
//     assertEqual(t, text3.text == "is"
//     box4, = line4.Box().Children
//     text4, = box4.Box().Children
//     assertEqual(t, text4.text == "text"
// }

// func TestWhiteSpace6(t *testing.T) {
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
//     line1, = whiteSpaceLines(1000000, "normal")
//     box1, = line1.Box().Children
//     text1, = box1.Box().Children
//     assertEqual(t, text1.text == "This + is text"
// }

// func TestWhiteSpace7(t *testing.T) {
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
//     line1, line2 = whiteSpaceLines(1000000, "pre")
//     box1, = line1.Box().Children
//     text1, = box1.Box().Children
//     assertEqual(t, text1.text == "This +    "
//     box2, = line2.Box().Children
//     text2, = box2.Box().Children
//     assertEqual(t, text2.text == "    is text"
// }

// func TestWhiteSpace8(t *testing.T) {
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
//     line1, = whiteSpaceLines(1000000, "nowrap")
//     box1, = line1.Box().Children
//     text1, = box1.Box().Children
//     assertEqual(t, text1.text == "This + is text"
// }

// func TestWhiteSpace9(t *testing.T) {
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
//     line1, line2 = whiteSpaceLines(1000000, "pre-wrap")
//     box1, = line1.Box().Children
//     text1, = box1.Box().Children
//     assertEqual(t, text1.text == "This +    "
//     box2, = line2.Box().Children
//     text2, = box2.Box().Children
//     assertEqual(t, text2.text == "    is text"
// }

// func TestWhiteSpace10(t *testing.T) {
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
//     line1, line2 = whiteSpaceLines(1000000, "pre-line")
//     box1, = line1.Box().Children
//     text1, = box1.Box().Children
//     assertEqual(t, text1.text == "This +"
//     box2, = line2.Box().Children
//     text2, = box2.Box().Children
//     assertEqual(t, text2.text == "is text"
// }

// func TestWhiteSpace11(t *testing.T) {
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
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
//     assertEqual(t, text1.text == "This"
//     assertEqual(t, box.elementTag == "br"
//     text2, = line2.Box().Children
//     assertEqual(t, text2.text == "is text"
// }

// func TestWhiteSpace12(t *testing.T) {
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
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
//     assertEqual(t, text1.text == "This is "
//     assertEqual(t, span.elementTag == "span"
//     assertEqual(t, text2.text == " text"
// }

// @assertEqual(t,NoLogs
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
//     line := paragraph.Box().Children[0]
//     assertEqual(t, line.Box().Width == width

// func TestTextTransform(t *testing.T){
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
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
//     assertEqual(t, text1.text == "Hé Lo1"
//     line2, = p2.Box().Children
//     text2, = line2.Box().Children
//     assertEqual(t, text2.text == "HÉ LO1"
//     line3, = p3.Box().Children
//     text3, = line3.Box().Children
//     assertEqual(t, text3.text == "hé lo1"
//     line4, = p4.Box().Children
//     text4, = line4.Box().Children
//     assertEqual(t, text4.text == "\uff48é\u3000\uff4c\uff2f\uff11"
//     line5, = p5.Box().Children
//     text5, = line5.Box().Children
//     assertEqual(t, text5.text == "hé lO1"
// }

// func TestTextFloatingPreLine(t *testing.T) {
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
//     // Test regression: https://github.com/Kozea/WeasyPrint/issues/610
//     page := renderOnePage(t, `
//       <div style="float: left; white-space: pre-line">This is
//       oh this end </div>
//     `)

// @assertEqual(t,NoLogs
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
//     assertEqual(t, inline.Box().Children[0].text == content
// }

// @pytest.mark.xfail
// func TestMaxLines(t *testing.T) {
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
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
//     assertEqual(t, text1.text == "abcd"
//     assertEqual(t, text2.text == "efgh"
//     assertEqual(t, text3.text == "ijkl"

// func TestContinue(t *testing.T){
// cp := testutils.CaptureLogs()
// defer cp.AssertNoLogs(t)
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
//     assertEqual(t, text1.text == "abcd"
//     assertEqual(t, text2.text == "efgh"
