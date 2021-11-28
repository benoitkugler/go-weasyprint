package pdf

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/benoitkugler/go-weasyprint/backend"
	"github.com/benoitkugler/go-weasyprint/document"
	"github.com/benoitkugler/go-weasyprint/layout/text"
	"github.com/benoitkugler/go-weasyprint/logger"
	"github.com/benoitkugler/go-weasyprint/style/tree"
	"github.com/benoitkugler/go-weasyprint/utils"
	tu "github.com/benoitkugler/go-weasyprint/utils/testutils"
	"github.com/benoitkugler/pdf/model"
	fc "github.com/benoitkugler/textlayout/fontconfig"
	"github.com/benoitkugler/textlayout/pango/fcfonts"
)

const fontmapCache = "../layout/text/test/cache.fc"

var fontconfig *text.FontConfiguration

var joker = color.RGBA{}

var colorByName = map[byte]color.RGBA{
	'_': {R: 255, G: 255, B: 255, A: 255}, // white
	'R': {R: 255, G: 0, B: 0, A: 255},     // red
	'B': {R: 0, G: 0, B: 255, A: 255},     // blue
	'G': {R: 0, G: 255, B: 0, A: 255},     // lime green
	'V': {R: 191, G: 0, B: 64, A: 255},    // average of 1*B and 3*R.
	'S': {R: 255, G: 63, B: 63, A: 255},   // R above R above #fff
	'r': {R: 255, G: 0, B: 0, A: 255},     // red
	'g': {R: 0, G: 128, B: 0, A: 255},     // half green
	'b': {R: 0, G: 0, B: 128, A: 255},     // half blue
	'v': {R: 128, G: 0, B: 128, A: 255},   // average of B and R.
	'h': {R: 64, G: 0, B: 64, A: 255},     // half average of B and R.
	'a': {R: 0, G: 0, B: 254, A: 255},     // JPG is lossy...
	'p': {R: 192, G: 0, B: 63, A: 255},    // R above R above B above #fff.
	'z': joker,                            // unspecified, accepts anything
}

func init() {
	logger.ProgressLogger.SetOutput(io.Discard)

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
	fontconfig = text.NewFontConfiguration(fcfonts.NewFontMap(fc.Standard.Copy(), fs))
}

// convert a PDF file to an image using Ghostscript, and extract the pixels,
// expecting a one color image
func pdfToColor(img image.Image) (color.RGBA, error) {
	rgb, ok := img.(*image.RGBA)
	if !ok {
		return color.RGBA{}, fmt.Errorf("unexpected image %T", img)
	}
	r := img.Bounds()
	col := rgb.RGBAAt(r.Min.X, r.Min.Y)
	for x := r.Min.X; x < r.Max.X; x++ {
		for y := r.Min.Y; y < r.Max.Y; y++ {
			if rgb.At(x, y) != col {
				return color.RGBA{}, fmt.Errorf("unexpected color at %d, %d: %v != %v", x, y, rgb.At(x, y), col)
			}
		}
	}
	return col, nil
}

// convert a PDF file to an image using Ghostscript
func pdfToImage(f *os.File, zoom utils.Fl) (image.Image, error) {
	const (
		resolution   = 96.
		antialiasing = 1
	)

	cmd := exec.Command("gs", "-q", "-dNOPAUSE", fmt.Sprintf("-dTextAlphaBits=%d", antialiasing),
		fmt.Sprintf("-dGraphicsAlphaBits=%d", antialiasing), "-sDEVICE=png16m",
		fmt.Sprintf("-r%d", int(resolution/zoom)),
		"-dBufferSpace=500000000", // 500 MB
		"-sOutputFile=-", f.Name())

	var output bytes.Buffer
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		return nil, err
	}
	pngs := output.Bytes()
	return pngsToImage(pngs)
}

func pngsToImage(pngs []byte) (image.Image, error) {
	const MAGIC_NUMBER = "\x89\x50\x4e\x47\x0d\x0a\x1a\x0a"

	if !bytes.HasPrefix(pngs, []byte(MAGIC_NUMBER)) {
		return nil, fmt.Errorf("invalid Ghoscript output: %s", string(pngs))
	}

	// for multiples page, we find each of them in the stream
	// and concatenate them
	pages := bytes.Split(pngs[8:], []byte(MAGIC_NUMBER))
	var images []image.Image
	var maxWidth, totalHeight int
	for _, page := range pages {
		img, err := png.Decode(bytes.NewReader(append([]byte(MAGIC_NUMBER), page...)))
		if err != nil {
			return nil, fmt.Errorf("invalid Ghoscript output: %s", err)
		}
		bounds := img.Bounds()
		if bounds.Dx() > maxWidth {
			maxWidth = bounds.Dx()
		}
		totalHeight += bounds.Dy()
		images = append(images, img)
	}

	finalImage := image.NewRGBA(image.Rect(0, 0, maxWidth, totalHeight))
	top := 0
	for _, img := range images {
		sr := img.Bounds()
		// center the image
		dp := image.Point{X: (maxWidth - sr.Dx()) / 2, Y: top}
		r := image.Rectangle{Min: dp, Max: dp.Add(sr.Size())}
		draw.Draw(finalImage, r, img, sr.Min, draw.Src)
		top += sr.Dy()
	}

	return finalImage, nil
}

// use the light UA stylesheet
func htmlToModel(t *testing.T, html string) model.Document {
	return htmlToModelExt(t, html, 1, ".")
}

func htmlToModelExt(t *testing.T, html string, zoom utils.Fl, baseURL string) model.Document {
	return htmlToModelExt2(t, html, zoom, baseURL, nil)
}

func htmlToModelExt2(t *testing.T, html string, zoom utils.Fl, baseURL string, attachments []backend.Attachment) model.Document {
	parsedHtml, err := tree.NewHTML(utils.InputString(html), baseURL, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	parsedHtml.UAStyleSheet = tree.TestUAStylesheet
	doc := document.Render(parsedHtml, nil, false, fontconfig)
	output := NewOutput()
	doc.Write(output, zoom, attachments)
	return output.Finalize()
}

// use the light UA stylesheet
func htmlToPDF(t *testing.T, html string, zoom utils.Fl) *os.File {
	target, err := ioutil.TempFile("", "weasyprint")
	if err != nil {
		t.Fatal(err)
	}

	pdfDoc := htmlToModelExt(t, html, zoom, ".")
	err = pdfDoc.Write(target, nil)
	if err != nil {
		t.Fatal(err)
	}

	return target
}

func TestWriteSimpleDocument(t *testing.T) {
	htmlContent := `      
	<style>
		@page { @top-left  { content: "[" string(left) "]" } }
		p { page-break-before: always }
		.initial { string-set: left "initial" }
		.empty   { string-set: left ""        }
		.space   { string-set: left " "       }
	</style>

	<p class="initial">  Initial</p>
	<p class="empty"> Empty</p>
	<p class="space">Space</p>
	`

	file := htmlToPDF(t, htmlContent, 1)
	defer os.Remove(file.Name())

	ti := time.Now()
	_, err := pdfToImage(file, 1)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(time.Since(ti))
}

func imagePixels(img image.Image) [][]color.RGBA {
	rgb, ok := img.(*image.RGBA)
	if !ok {
		panic(fmt.Errorf("unexpected image %T", img))
	}
	r := img.Bounds()
	var out [][]color.RGBA
	for y := r.Min.Y; y < r.Max.Y; y++ {
		var line []color.RGBA
		for x := r.Min.X; x < r.Max.X; x++ {
			line = append(line, rgb.RGBAAt(x, y))
		}
		out = append(out, line)
	}
	return out
}

const pdfZoom = 4. / 30

func TestZIndex(t *testing.T) {
	cp := tu.CaptureLogs()
	defer cp.AssertNoLogs(t)

	for _, data := range []struct {
		zIndexes []interface{}
		color    byte
	}{
		{[]interface{}{"3", "2", "1"}, 'R'},
		{[]interface{}{"1", "2", "3"}, 'G'},
		{[]interface{}{"1", "2", "-3"}, 'B'},
		{[]interface{}{"1", "2", "auto"}, 'B'},
		{[]interface{}{"-1", "auto", "-2"}, 'B'},
	} {
		source := fmt.Sprintf(`
		<style>
		  @page { size: 30px }
		  body { background: white }
		  div, div * { width: 30px; height: 30px; position: absolute }
		  article { background: red; z-index: %s }
		  section { background: blue; z-index: %s }
		  nav { background: lime; z-index: %s }
		</style>
		<div>
		  <article></article>
		  <section></section>
		  <nav></nav>
		</div>`, data.zIndexes...)
		b := htmlToPDF(t, source, pdfZoom)
		img, err := pdfToImage(b, pdfZoom)
		if err != nil {
			fmt.Println(b.Name())
			t.Fatal(err)
		}
		col, err := pdfToColor(img)
		if err != nil {
			fmt.Println(b.Name())
			t.Fatal(err)
		}

		if exp := colorByName[data.color]; col != exp {
			fmt.Println(b.Name())
			t.Fatalf("expected %v, got %v", exp, col)
		}

		os.Remove(b.Name())
	}
}

// convert from a human-friendly string representation
// to a matrix of colors
func parsePixels(pixels string) [][]color.RGBA {
	return parsePixelsExt(pixels, make(map[byte]color.RGBA))
}

func parsePixelsExt(pixels string, pixelsOveride map[byte]color.RGBA) [][]color.RGBA {
	for k, v := range colorByName {
		if _, has := pixelsOveride[k]; !has {
			pixelsOveride[k] = v
		}
	}

	pixels = strings.TrimSpace(pixels)
	lines := strings.Split(pixels, "\n")
	var out [][]color.RGBA
	for _, line := range lines {
		line = strings.Split(line, "#")[0]
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		row := make([]color.RGBA, len(line)) // line is ASCII only
		for j, c := range line {
			row[j] = pixelsOveride[byte(c)]
		}
		out = append(out, row)
	}
	return out
}

func assertPixelsEqualFromPixels(t *testing.T, context string, expectedPixels [][]color.RGBA, input string) {
	t.Helper()

	got := htmlToPDF(t, input, pdfZoom)

	img, err := pdfToImage(got, pdfZoom)
	if err != nil {
		t.Fatal(context, "file", got.Name(), err)
	}

	gotPixels := imagePixels(img)
	if len(gotPixels) != len(expectedPixels) {
		t.Fatalf("%s (file %s): expected %d pixels rows, got %d", context, got.Name(), len(expectedPixels), len(gotPixels))
	}

	for i, exp := range expectedPixels {
		for j, v := range exp {
			if v == joker {
				gotPixels[i][j] = joker
			}
		}

		if len(gotPixels[i]) != len(exp) {
			t.Fatalf("%s (file %s): unexpected length for row %d : expected %d, got %d", context, got.Name(), i, len(exp), len(gotPixels[i]))
		}

		for j, v := range exp {
			g := gotPixels[i][j]
			if v != g {
				t.Fatalf("%s (file: %s): pixel at (%d, %d): expected %v, got %v", context, got.Name(), i, j, v, g)
			}
		}
	}

	got.Close()
	os.Remove(got.Name())
}

func assertPixelsEqual(t *testing.T, context, expected, input string) {
	t.Helper()
	assertPixelsEqualFromPixels(t, context, parsePixels(expected), input)
}

func arePixelsAlmostEqual(pix1, pix2 [][]color.RGBA, tolerance uint8) bool {
	diff := func(a, b uint8) uint8 {
		if a < b {
			return b - a
		}
		return a - b
	}

	if len(pix1) != len(pix2) {
		return false
	}

	for i, line1 := range pix1 {
		line2 := pix2[i]
		if len(line1) != len(line2) {
			return false
		}

		for j, c1 := range line1 {
			c2 := line2[j]
			if diff(c1.R, c2.R) > tolerance || diff(c1.G, c2.G) > tolerance || diff(c1.B, c2.B) > tolerance || diff(c1.A, c2.A) > tolerance {
				return false
			}
		}
	}

	return true
}

func assertSameRendering(t *testing.T, context, input1, input2 string, tolerance uint8) {
	t.Helper()

	got1 := htmlToPDF(t, input1, pdfZoom)
	got2 := htmlToPDF(t, input2, pdfZoom)

	img1, err := pdfToImage(got1, pdfZoom)
	if err != nil {
		t.Fatal(context, err)
	}
	img2, err := pdfToImage(got2, pdfZoom)
	if err != nil {
		t.Fatal(context, err)
	}

	gotPixels1 := imagePixels(img1)
	gotPixels2 := imagePixels(img2)

	if !arePixelsAlmostEqual(gotPixels1, gotPixels2, tolerance) {
		t.Fatal(context, "got different rendering", got1.Name(), got2.Name())
	}

	got1.Close()
	got2.Close()
	os.Remove(got1.Name())
	os.Remove(got2.Name())
}

func TestTableVerticalAlign(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pixels := `
		rrrrrrrrrrrrrrrrrrrrrrrrrrrr
		rBBBBBBBBBBBBBBBBBBBBBBBBBBr
		rBrBB_BB_BB_BB_BBrrBBrrBB_Br
		rBrBB_BB_BBrBBrBBrrBBrrBBrBr
		rB_BBrBB_BBrBBrBBrrBBrrBBrBr
		rB_BBrBB_BB_BB_BBrrBBrrBB_Br
		rB_BB_BBrBB_BB_BB__BB__BB_Br
		rB_BB_BBrBB_BB_BB__BB__BB_Br
		rBBBBBBBBBBBBBBBBBBBBBBBBBBr
		rrrrrrrrrrrrrrrrrrrrrrrrrrrr
    `

	input := `<style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page { size: 28px 10px }
        html { background: #fff; font-size: 1px; color: red }
        body { margin: 0; width: 28px; height: 10px }
        td {
          width: 1em;
          padding: 0 !important;
          border: 1px solid blue;
          line-height: 1em;
          font-family: weasyprint;
        }
      </style>
      <table style="border: 1px solid red; border-spacing: 0">
        <tr>
          <!-- Test vertical-align: top, auto height -->
          <td style="vertical-align: top">o o</td>

          <!-- Test vertical-align: middle, auto height -->
          <td style="vertical-align: middle">o o</td>

          <!-- Test vertical-align: bottom, fixed useless height -->
          <td style="vertical-align: bottom; height: 2em">o o</td>

          <!-- Test default vertical-align value (baseline),
               fixed useless height -->
          <td style="height: 5em">o o</td>

          <!-- Test vertical-align: baseline with baseline set by next cell,
               auto height -->
          <td style="vertical-align: baseline">o o</td>

          <!-- Set baseline height to 2px, auto height -->
          <td style="vertical-align: baseline; font-size: 2em">o o</td>

          <!-- Test padding-bottom, fixed useless height,
               set the height of the cells to 2 lines * 2em + 2px = 6px -->
          <td style="vertical-align: baseline; height: 1em;
                     font-size: 2em; padding-bottom: 2px !important">
            o o
          </td>

          <!-- Test padding-top, auto height -->
          <td style="vertical-align: top; padding-top: 1em !important">
            o o
          </td>
        </tr>
      </table>
    `
	assertPixelsEqual(t, "", pixels, input)
}
