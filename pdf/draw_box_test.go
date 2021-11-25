package pdf

import (
	"bytes"
	"fmt"
	"image/color"
	"os"
	"testing"

	"github.com/benoitkugler/go-weasyprint/document"
	"github.com/benoitkugler/go-weasyprint/style/tree"
	"github.com/benoitkugler/go-weasyprint/utils"
	"github.com/benoitkugler/go-weasyprint/utils/testutils"
)

// Test how boxes, borders, outlines are drawn.

func inputToPixels(t *testing.T, input string) [][]color.RGBA {
	doc := htmlToPDF(t, input, pdfZoom)

	img, err := pdfToImage(doc, pdfZoom)
	if err != nil {
		t.Fatal(err)
	}

	// fmt.Println(doc.Name())
	doc.Close()
	os.Remove(doc.Name())

	return imagePixels(img)
}

func assertPixelsDifferents(t *testing.T, images [][][]color.RGBA) {
	for i, img1 := range images {
		for j, img2 := range images {
			if i == j {
				continue
			}
			if arePixelsAlmostEqual(img1, img2, 2) {
				t.Fatalf("renderings %d and %d should be different", i, j)
			}
		}
	}
}

func testBorders(t *testing.T, cssMargin, prop string) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	source := `
      <style>
        @page { size: 140px 110px }
        html { background: #fff }
        body { width: 100px; height: 70px;
               margin: %s; %s: 10px %s blue }
      </style>
      <body>`

	// Do not test the exact rendering of earch border style but at least
	// check that they do not do the same.
	var pixels [][][]color.RGBA
	for _, borderStyle := range []string{
		"none", "solid", "dashed", "dotted",
		"double",
		"inset", "outset", "groove", "ridge",
	} {
		input := fmt.Sprintf(source, cssMargin, prop, borderStyle)
		pixels = append(pixels, inputToPixels(t, input))
	}
	assertPixelsDifferents(t, pixels)

	width := 140
	height := 110
	margin := 10
	border := 10

	solidPixels := make([][]byte, height)
	for i := range solidPixels {
		solidPixels[i] = bytes.Repeat([]byte{'_'}, width)
	}
	for x := margin; x < width-margin; x++ {
		for y := margin; y < margin+border; y++ {
			solidPixels[y][x] = 'B'
		}
		for y := height - margin - border; y < height-margin; y++ {
			solidPixels[y][x] = 'B'
		}
	}
	for y := margin; y < height-margin; y++ {
		for x := margin; x < margin+border; x++ {
			solidPixels[y][x] = 'B'
		}
		for x := width - margin - border; x < width-margin; x++ {
			solidPixels[y][x] = 'B'
		}
	}

	assertPixelsEqual(t, prop+"_solid", string(bytes.Join(solidPixels, []byte{'\n'})), fmt.Sprintf(source, cssMargin, prop, "solid"))
}

// Test the rendering of borders
func TestBorders(t *testing.T) {
	testBorders(t, "10px", "border")
}

func TestOutlines(t *testing.T) {
	testBorders(t, "20px", "outline")
}

func TestSmallBorders_1(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, borderStyle := range []string{"none", "solid", "dashed", "dotted"} {
		// Regression test for ZeroDivisionError on dashed or dotted borders
		// smaller than a dash/dot.
		// https://github.com/Kozea/WeasyPrint/issues/49
		html := fmt.Sprintf(`
		<style>
			@page { size: 50px 50px }
			html { background: #fff }
			body { margin: 5px; height: 0; border: 10px %s blue }
		</style>
		<body>`, borderStyle)
		f := htmlToPDF(t, html, 1)
		os.Remove(f.Name())
	}
}

func TestSmallBorders_2(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, borderStyle := range []string{"none", "solid", "dashed", "dotted"} {
		// Regression test for ZeroDivisionError on dashed or dotted borders
		// smaller than a dash/dot.
		// https://github.com/Kozea/WeasyPrint/issues/146
		html := fmt.Sprintf(`
		<style>
			@page { size: 50px 50px }
			html { background: #fff }
			body { height: 0; width: 0; border-width: 1px 0; border-style: %s }
		</style>
		<body>`, borderStyle)
		f := htmlToPDF(t, html, 1)
		os.Remove(f.Name())
	}
}

func TestEmBorders(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1378
	html := `<body style="border: 1em solid">`
	f := htmlToPDF(t, html, 1)
	os.Remove(f.Name())
}

func TestMarginBoxes(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "margin_boxes", `
        _______________
        _GGG______BBBB_
        _GGG______BBBB_
        _______________
        _____RRRR______
        _____RRRR______
        _____RRRR______
        _____RRRR______
        _______________
        _bbb______gggg_
        _bbb______gggg_
        _bbb______gggg_
        _bbb______gggg_
        _bbb______gggg_
        _______________
    `, `
      <style>
        html { height: 100% }
        body { background: #f00; height: 100% }
        @page {
          size: 15px;
          margin: 4px 6px 7px 5px;
          background: white;

          @top-left-corner {
            margin: 1px;
            content: " ";
            background: #0f0;
          }
          @top-right-corner {
            margin: 1px;
            content: " ";
            background: #00f;
          }
          @bottom-right-corner {
            margin: 1px;
            content: " ";
            background: #008000;
          }
          @bottom-left-corner {
            margin: 1px;
            content: " ";
            background: #000080;
          }
        }
      </style>
      <body>`)
}

func TestDisplayInlineBlockTwice(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for inline blocks displayed twice.
	// https://github.com/Kozea/WeasyPrint/issues/880
	html := `<div style="background: red; display: inline-block">`

	parsedHtml, err := tree.NewHTML(utils.InputString(html), ".", nil, "")
	if err != nil {
		t.Fatal(err)
	}
	parsedHtml.UAStyleSheet = tree.TestUAStylesheet
	doc := document.Render(parsedHtml, nil, false, fontconfig)

	output := NewOutput()
	doc.WriteDocument(output, 1, nil)
	_ = output.Finalize()

	output = NewOutput()
	doc.WriteDocument(output, 1, nil)
	_ = output.Finalize()
}

func TestRoundedRect(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for inline blocks displayed twice.
	// https://github.com/Kozea/WeasyPrint/issues/880
	html := `
	<style>
	@page {
		size: 40px 25px;
		background: white;
		margin: 2px;
	}
	</style>
	<span style="background: red; border-radius: 5px; border: 2px solid blue;">abc</span>
	`

	f := htmlToPDF(t, html, 1)
	got, err := pdfToImage(f, 1)
	if err != nil {
		t.Fatal(err)
	}

	pngs, err := os.ReadFile("../resources_test/rounded_rect_ref.png")
	if err != nil {
		t.Fatal(err)
	}
	exp, err := pngsToImage(pngs)
	if err != nil {
		t.Fatal(err)
	}

	if !arePixelsAlmostEqual(imagePixels(exp), imagePixels(got), 0) {
		t.Fatal("unexpected pixels")
	}

	f.Close()
	os.Remove(f.Name())
}
