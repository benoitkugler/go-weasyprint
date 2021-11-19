package pdf

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/benoitkugler/go-weasyprint/document"
	"github.com/benoitkugler/go-weasyprint/layout/text"
	"github.com/benoitkugler/go-weasyprint/logger"
	"github.com/benoitkugler/go-weasyprint/style/tree"
	"github.com/benoitkugler/go-weasyprint/utils"
	tu "github.com/benoitkugler/go-weasyprint/utils/testutils"
	fc "github.com/benoitkugler/textlayout/fontconfig"
	"github.com/benoitkugler/textlayout/pango/fcfonts"
)

const fontmapCache = "layout/text/test/cache.fc"

var fontconfig *text.FontConfiguration

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
	'z': {},
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
		"-sOutputFile=-", f.Name())
	var output bytes.Buffer
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	img, err := png.Decode(&output)
	if err != nil {
		return nil, fmt.Errorf("invalid Ghoscript output: %s", err)
	}

	return img, nil
}

// use the light UA stylesheet
func htmlToPDF(t *testing.T, html string, zoom utils.Fl) *os.File {
	target, err := ioutil.TempFile("", "weasyprint")
	if err != nil {
		t.Fatal(err)
	}

	parsedHtml, err := tree.NewHTML(utils.InputString(html), ".", nil, "")
	if err != nil {
		t.Fatal(err)
	}
	parsedHtml.UAStyleSheet = tree.TestUAStylesheet
	doc := document.Render(parsedHtml, nil, false, fontconfig)
	output := NewOutput()
	doc.WriteDocument(output, zoom, nil)
	pdfDoc := output.Finalize()
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
	pixels = strings.TrimSpace(pixels)
	lines := strings.Split(pixels, "\n")
	out := make([][]color.RGBA, len(lines))
	for i, line := range lines {
		line = strings.TrimSpace(line)
		out[i] = make([]color.RGBA, len(line)) // line is ASCII only
		for j, c := range line {
			out[i][j] = colorByName[byte(c)]
		}
	}
	return out
}

func assertPixelsEqual(t *testing.T, input string, expected string) {
	got := htmlToPDF(t, input, pdfZoom)

	img, err := pdfToImage(got, pdfZoom)
	if err != nil {
		fmt.Println(got.Name())
		t.Fatal(err)
	}

	gotPixels := imagePixels(img)
	expectedPixels := parsePixels(expected)
	if len(gotPixels) != len(expectedPixels) {
		t.Fatalf("expected pixels length %d, got %d", len(expectedPixels), len(gotPixels))
	}

	for i, exp := range expectedPixels {
		if !reflect.DeepEqual(gotPixels[i], exp) {
			fmt.Println(got.Name())
			t.Fatalf("unexpected pixels at line %d", i)
		}
	}

	got.Close()
	os.Remove(got.Name())
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
        @font-face { src: url(resources_test/weasyprint.otf); font-family: weasyprint }
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
	assertPixelsEqual(t, input, pixels)
}
