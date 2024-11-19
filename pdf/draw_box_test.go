package pdf

import (
	"bytes"
	"fmt"
	"image/color"
	"os"
	"testing"

	"github.com/benoitkugler/webrender/html/document"
	"github.com/benoitkugler/webrender/html/tree"
	"github.com/benoitkugler/webrender/utils"
	"github.com/benoitkugler/webrender/utils/testutils"
	tu "github.com/benoitkugler/webrender/utils/testutils"
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

	assertPixelsEqual(t, string(bytes.Join(solidPixels, []byte{'\n'})), fmt.Sprintf(source, cssMargin, prop, "solid"))
}

// Test the rendering of borders
func TestBorders(t *testing.T) {
	testBorders(t, "10px", "border")
}

// Test the rendering of collapsing borders.
func TestBordersTableCollapse(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	source := `
      <style>
        @page { size: 140px 110px }
        table { width: 100px; height: 70px; margin: 10px;  
                border-collapse: collapse; border: 10px %s blue }
      </style>
      <table><td>abc</td>`

	// Do not test the exact rendering of earch border style but at least
	// check that they do not do the same.
	var documents []string
	for _, borderStyle := range [...]string{
		"none", "solid", "dashed", "dotted", "double",
		"outset", /* "groove", */
		"inset",  /* "ridge", */
	} {
		documents = append(documents, fmt.Sprintf(source, borderStyle))
	}
	assertDifferentRenderings(t, documents)

	assertSameRendering(t, fmt.Sprintf(source, "outset"), fmt.Sprintf(source, "groove"), 0)
	assertSameRendering(t, fmt.Sprintf(source, "inset"), fmt.Sprintf(source, "ridge"), 0)
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

func TestBordersBoxSizing(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        ________
        _RRRRRR_
        _R____R_
        _RRRRRR_
        ________
    `, `
      <style>
        @page {
          size: 8px 5px;
        }
        div {
          border: 1px solid red;
          box-sizing: border-box;
          height: 3px;
          margin: 1px;
          min-height: auto;
          min-width: auto;
          width: 6px;
        }
      </style>
      <div></div>
    `)
}

func TestMarginBoxes(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, `
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
	doc.Write(output, 1, nil)
	_ = output.Finalize()

	output = NewOutput()
	doc.Write(output, 1, nil)
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

func TestDrawBorderRadius(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        ___zzzzz
        __zzzzzz
        _zzzzzzz
        zzzzzzzz
        zzzzzzzz
        zzzzzzzR
        zzzzzzRR
        zzzzzRRR
    `, `
      <style>
        @page {
          size: 8px 8px;
        }
        div {
          background: red;
          border-radius: 50% 0 0 0;
          height: 16px;
          width: 16px;
        }
      </style>
      <div></div>
    `)
}

func TestDrawSplitBorderRadius(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        ___zzzzz
        __zzzzzz
        _zzzzzzz
        zzzzzzzz
        zzzzzzzz
        zzzzzzzz
        zzzzzzRR
        zzzzzRRR

        RRRRRRRR
        RRRRRRRR
        RRRRRRRR
        RRRRRRRR
        RRRRRRRR
        RRRRRRRR
        RRRRRRRR
        RRRRRRRR

        zzzzzzRR
        zzzzzzzR
        zzzzzzzz
        zzzzzzzz
        zzzzzzzz
        zzzzzzzz
        _zzzzzzz
        __zzzzzz
    `, `
      <style>
        @page {
          size: 8px 8px;
        }
        div {
          background: red;
          color: transparent;
          border-radius: 8px;
          line-height: 9px;
          width: 16px;
        }
      </style>
      <div>a b c</div>
    `)
}

func TestBorderImageStretch(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        __________
        _RYYYMMMG_
        _M______C_
        _M______C_
        _Y______Y_
        _Y______Y_
        _BYYYCCCK_
        __________
    `, `
      <style>
        @page {
          size: 10px 8px;
        }
        div {
          border: 1px solid black;
          border-image-source: url(../resources_test/border.svg);
          border-image-slice: 25%;
          height: 4px;
          margin: 1px;
          width: 6px;
        }
      </style>
      <div></div>
    `)
}

func TestBorderImageFill(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        __________
        _RYYYMMMG_
        _MbbbgggC_
        _MbbbgggC_
        _YgggbbbY_
        _YgggbbbY_
        _BYYYCCCK_
        __________
    `, `
      <style>
        @page {
          size: 10px 8px;
        }
        div {
          border: 1px solid black;
          border-image-source: url(../resources_test/border.svg);
          border-image-slice: 25% fill;
          height: 4px;
          margin: 1px;
          width: 6px;
        }
      </style>
      <div></div>
    `)
}

func TestBorderImageDefaultSlice(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        _____________
        _RYMG___RYMG_
        _MbgC___MbgC_
        _YgbY___YgbY_
        _BYCK___BYCK_
        _____________
        _____________
        _RYMG___RYMG_
        _MbgC___MbgC_
        _YgbY___YgbY_
        _BYCK___BYCK_
        _____________
    `, `
      <style>
        @page {
          size: 13px 12px;
        }
        div {
          border: 4px solid black;
          border-image-source: url(../resources_test/border.svg);
          height: 2px;
          margin: 1px;
          width: 3px;
        }
      </style>
      <div></div>
    `)
}

func TestBorderImageUnevenWidth(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        ____________
        _RRRYYYMMMG_
        _MMM______C_
        _MMM______C_
        _YYY______Y_
        _YYY______Y_
        _BBBYYYCCCK_
        ____________
    `, `
      <style>
        @page {
          size: 12px 8px;
        }
        div {
          border: 1px solid black;
          border-left-width: 3px;
          border-image-source: url(../resources_test/border.svg);
          border-image-slice: 25%;
          height: 4px;
          margin: 1px;
          width: 6px;
        }
      </style>
      <div></div>
    `)
}

func TestBorderImageNotPercent(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        __________
        _RYYYMMMG_
        _M______C_
        _M______C_
        _Y______Y_
        _Y______Y_
        _BYYYCCCK_
        __________
    `, `
      <style>
        @page {
          size: 10px 8px;
        }
        div {
          border: 1px solid black;
          border-image-source: url(../resources_test/border.svg);
          border-image-slice: 1;
          height: 4px;
          margin: 1px;
          width: 6px;
        }
      </style>
      <div></div>
    `)
}

func TestBorderImageRepeat(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        ___________
        _RYMYMYMYG_
        _M_______C_
        _Y_______Y_
        _M_______C_
        _Y_______Y_
        _BYCYCYCYK_
        ___________
    `, `
      <style>
        @page {
          size: 11px 8px;
        }
        div {
          border: 1px solid black;
          border-image-source: url(../resources_test/border.svg);
          border-image-slice: 25%;
          border-image-repeat: repeat;
          height: 4px;
          margin: 1px;
          width: 7px;
        }
      </style>
      <div></div>
    `)
}

func TestBorderImageSpace(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        _________
        _R_YMC_G_
        _________
        _M_____C_
        _Y_____Y_
        _C_____M_
        _________
        _B_YCM_K_
        _________
    `, `
      <style>
        @page {
          size: 9px 9px;
        }
        div {
          border: 1px solid black;
          border-image-source: url(../resources_test/border2.svg);
          border-image-slice: 20%;
          border-image-repeat: space;
          height: 5px;
          margin: 1px;
          width: 5px;
        }
      </style>
      <div></div>
    `)
}

func TestBorderImageOutset(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        ____________
        _RYYYYMMMMG_
        _M________C_
        _M_bbbbbb_C_
        _M_bbbbbb_C_
        _Y_bbbbbb_Y_
        _Y_bbbbbb_Y_
        _Y________Y_
        _BYYYYCCCCK_
        ____________
    `, `
      <style>
        @page {
          size: 12px 10px;
        }
        div {
          border: 1px solid black;
          border-image-source: url(../resources_test/border.svg);
          border-image-slice: 25%;
          border-image-outset: 2px;
          height: 2px;
          margin: 3px;
          width: 4px;
          background: #000080
        }
      </style>
      <div></div>
    `)
}

func TestBorderImageWidth(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        __________
        _RRYYMMGG_
        _RRYYMMGG_
        _MM____CC_
        _YY____YY_
        _BBYYCCKK_
        _BBYYCCKK_
        __________
    `, `
      <style>
        @page {
          size: 10px 8px;
        }
        div {
          border: 1px solid black;
          border-image-source: url(../resources_test/border.svg);
          border-image-slice: 25%;
          border-image-width: 2;
          height: 4px;
          margin: 1px;
          width: 6px;
        }
      </style>
      <div></div>
    `)
}

func TestBorderImageGradient(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        __________
        _RRRRRRRR_
        _RRRRRRRR_
        _RR____RR_
        _BB____BB_
        _BBBBBBBB_
        _BBBBBBBB_
        __________
    `, `
      <style>
        @page {
          size: 10px 8px;
        }
        div {
          border: 1px solid black;
          border-image-source: linear-gradient(to bottom, red, red 50%, blue 50%, blue);
          border-image-slice: 25%;
          border-image-width: 2;
          height: 4px;
          margin: 1px;
          width: 6px;
        }
      </style>
      <div></div>
    `)
}
