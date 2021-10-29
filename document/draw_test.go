package document

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"testing"
	"text/template"

	"github.com/benoitkugler/go-weasyprint/layout/text"
	"github.com/benoitkugler/go-weasyprint/style/tree"
	"github.com/benoitkugler/go-weasyprint/utils"
	"github.com/benoitkugler/textlayout/fontconfig"
	"github.com/benoitkugler/textlayout/pango/fcfonts"
)

const fontmapCache = "../layout/text/test/cache.fc"

var fc *text.FontConfiguration

func init() {
	// this command has to run once
	// fmt.Println("Scanning fonts...")
	// _, err := fc.ScanAndCache(fontmapCache)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	fs, err := fontconfig.LoadFontsetFile(fontmapCache)
	if err != nil {
		log.Fatal(err)
	}
	fc = text.NewFontConfiguration(fcfonts.NewFontMap(fontconfig.Standard.Copy(), fs))
}

// convert a PDF file to an image using Ghostscript, and extract the pixels,
// expecting a one color image
func pdfToColor(filename string) (color.RGBA, error) {
	img, err := pdfToImage(filename)
	if err != nil {
		return color.RGBA{}, err
	}
	rgb, ok := img.(*image.RGBA)
	if !ok {
		return color.RGBA{}, fmt.Errorf("unexpected image %T", img)
	}
	r := img.Bounds()
	col := rgb.RGBAAt(r.Min.X, r.Min.Y)
	for x := r.Min.X; x < r.Max.X; x++ {
		for y := r.Min.Y; y < r.Max.Y; y++ {
			if rgb.At(x, y) != col {
				return color.RGBA{}, fmt.Errorf("unexpected color at %d, %d", x, y)
			}
		}
	}
	return col, nil
}

// convert a PDF file to an image using Ghostscript
func pdfToImage(filename string) (image.Image, error) {
	const (
		resolution   = 96.
		antialiasing = 1
		zoom         = 4. / 30
	)
	cmd := exec.Command("gs", "-q", "-dNOPAUSE", fmt.Sprintf("-dTextAlphaBits=%d", antialiasing),
		fmt.Sprintf("-dGraphicsAlphaBits=%d", antialiasing), "-sDEVICE=png16m",
		fmt.Sprintf("-r%d", int(resolution/zoom)), "-sOutputFile=-", filename)
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

func TestPixels(t *testing.T) {
	col, err := pdfToColor("/home/benoit/Téléchargements/WeasyPrint/tmp_pixels2.pdf")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(col)
}

func TestStacking(t *testing.T) {
	var s StackingContext
	if s.IsClassicalBox() {
		t.Fatal("should not be a classical box")
	}
}

func TestSVG(t *testing.T) {
	tmp := headerSVG + crop + cross
	tp := template.Must(template.New("svg").Parse(tmp))
	if err := tp.Execute(ioutil.Discard, svgArgs{}); err != nil {
		t.Fatal(err)
	}
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

	<p class="initial">Initial</p>
	<p class="empty">Empty</p>
	<p class="space">Space</p>
	`

	doc, err := tree.NewHTML(utils.InputString(htmlContent), "", nil, "")
	if err != nil {
		t.Fatal(err)
	}
	finalDoc := Render(doc, nil, true, fc)
	finalDoc.WriteDocument(output{}, 1, nil)
}

func TestWriteDocument(t *testing.T) {
	doc, err := tree.NewHTML(utils.InputFilename("../resources_test/acid2-test.html"), "", nil, "")
	if err != nil {
		t.Fatal(err)
	}
	finalDoc := Render(doc, nil, true, fc)
	finalDoc.WriteDocument(output{}, 1, nil)
}

func renderUrl(t testing.TB, url string) {
	doc, err := tree.NewHTML(utils.InputUrl(url), "", nil, "")
	if err != nil {
		t.Fatal(err)
	}
	finalDoc := Render(doc, nil, true, fc)
	finalDoc.WriteDocument(output{}, 1, nil)
}

func TestRealPage(t *testing.T) {
	outputLog.SetOutput(io.Discard)
	// renderUrl(t, "http://www.google.com")
	// renderUrl(t, "https://weasyprint.org/")
	// renderUrl(t, "https://en.wikipedia.org/wiki/Go_(programming_language)") // rather big document
	// renderUrl(t, "https://golang.org/doc/go1.17")                           // slow because of text layout
	renderUrl(t, "https://github.com/Kozea/WeasyPrint")
}

func BenchmarkRender(b *testing.B) {
	outputLog.SetOutput(io.Discard)

	for i := 0; i < b.N; i++ {
		renderUrl(b, "https://golang.org/doc/go1.17")
	}
}

// TODO:
// func TestTableVerticalAlign(t *testing.T) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

// 	assertPixels("tableVerticalAlign", pr.Float(28), pr.Float(10), `
//         rrrrrrrrrrrrrrrrrrrrrrrrrrrr
//         rBBBBBBBBBBBBBBBBBBBBBBBBBBr
//         rBrBBBBBBBBBBrrBBrrBBBr
//         rBrBBBBBBrBBrBBrrBBrrBBrBr
//         rBBBrBBBBrBBrBBrrBBrrBBrBr
//         rBBBrBBBBBBBBrrBBrrBBBr
//         rBBBBBrBBBBBB_BB_BBBr
//         rBBBBBrBBBBBB_BB_BBBr
//         rBBBBBBBBBBBBBBBBBBBBBBBBBBr
//         rrrrrrrrrrrrrrrrrrrrrrrrrrrr
//     `, `
//       <style>
//         @font-face { src: url(weasyprint.otf); font-family: weasyprint }
//         @page { size: 28px 10px }
//         html { background: #fff; font-size: 1px; color: red }
//         body { margin: 0; width: 28px; height: 10px }
//         td {
//           width: 1em;
//           padding: 0 !important;
//           border: 1px solid blue;
//           line-height: 1em;
//           font-family: weasyprint;
//         }
//       </style>
//       <table style="border: 1px solid red; border-spacing: 0">
//         <tr>
//           <!-- Test vertical-align: top, auto height -->
//           <td style="vertical-align: top">o o</td>

//           <!-- Test vertical-align: middle, auto height -->
//           <td style="vertical-align: middle">o o</td>

//           <!-- Test vertical-align: bottom, fixed useless height -->
//           <td style="vertical-align: bottom; height: 2em">o o</td>

//           <!-- Test default vertical-align value (baseline),
//                fixed useless height -->
//           <td style="height: 5em">o o</td>

//           <!-- Test vertical-align: baseline with baseline set by next cell,
//                auto height -->
//           <td style="vertical-align: baseline">o o</td>

//           <!-- Set baseline height to 2px, auto height -->
//           <td style="vertical-align: baseline; font-size: 2em">o o</td>

//           <!-- Test padding-bottom, fixed useless height,
//                set the height of the cells to 2 lines * 2em + 2px = 6px -->
//           <td style="vertical-align: baseline; height: 1em;
//                      font-size: 2em; padding-bottom: 2px !important">
//             o o
//           </td>

//           <!-- Test padding-top, auto height -->
//           <td style="vertical-align: top; padding-top: 1em !important">
//             o o
//           </td>
//         </tr>
//       </table>
//     `) // noqa
// }
