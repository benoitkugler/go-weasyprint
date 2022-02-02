package goweasyprint

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/benoitkugler/pdf/reader/file"
	fc "github.com/benoitkugler/textlayout/fontconfig"
	"github.com/benoitkugler/textlayout/pango/fcfonts"
	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/text"
	"github.com/benoitkugler/webrender/utils"
)

// see pdf/test/draw_test.go
const fontmapCache = "pdf/test/cache.fc"

var fontconfig *text.FontConfiguration

func init() {
	logger.ProgressLogger.SetOutput(io.Discard)

	fs, err := fc.LoadFontsetFile(fontmapCache)
	if err != nil {
		log.Fatal(err)
	}
	fontconfig = text.NewFontConfiguration(fcfonts.NewFontMap(fc.Standard.Copy(), fs))
}

func tempFile(s string) string {
	dir := os.TempDir()
	return filepath.Join(dir, "test_svg_text.pdf")
}

func TestRealPage(t *testing.T) {
	path := tempFile("test_real_page.pdf")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	// url := "http://www.google.com"
	// url := "https://weasyprint.org/"
	// url := "https://en.wikipedia.org/wiki/Go_(programming_language)" // rather big document
	// url := "https://golang.org/doc/go1.17" // slow because of text layout
	url := "https://developer.mozilla.org/en-US/docs/Web/SVG/Attribute/preserveAspectRatio"
	log.SetOutput(io.Discard)
	// url := "https://github.com/Kozea/WeasyPrint"
	err = HtmlToPdf(f, utils.InputUrl(url), fontconfig)
	if err != nil {
		t.Fatal(err)
	}

	f.Close()

	_, err = file.ReadFile(path, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSVGMask(t *testing.T) {
	path := tempFile("test_svg_mask.pdf")

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	input := `
	<style>
		@page { size: 300px }
		svg { display: block }
	</style>
  	<svg viewBox="-10 -10 120 120">
		<mask id="myMask">
			<!-- Everything under a white pixel will be visible -->
			<rect x="0" y="0" width="100" height="100" fill="white" />

			<!-- Everything under a black pixel will be invisible -->
			<path d="M10,35 A20,20,0,0,1,50,35 A20,20,0,0,1,90,35 Q90,65,50,95 Q10,65,10,35 Z" fill="black" />
		</mask>

		<polygon points="-10,110 110,110 110,-10" fill="orange" />

		<!-- with this mask applied, we "punch" a heart shape hole into the circle -->
		<circle cx="50" cy="50" r="50" mask="url(#myMask)" />
	</svg>
	`
	err = HtmlToPdf(f, utils.InputString(input), fontconfig)
	if err != nil {
		t.Fatal(err)
	}

	f.Close()

	_, err = file.ReadFile(path, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSVGText(t *testing.T) {
	path := tempFile("test_svg_text.pdf")

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	input := `
	<style>
		@page { size: 400px }
		svg { display: block }
	</style>
	<svg viewBox="0 0 240 80" xmlns="http://www.w3.org/2000/svg">
		<style>
		.small { font: italic 13px sans-serif; font-style: italic }
		.heavy { font: bold 30px sans-serif; }
	
		/* Note that the color of the text is set with the    *
		* fill property, the color property is for HTML only */
		.Rrrrr { font: italic 40px serif; fill: red; }
		</style>
	
		<text x="20" y="35" class="small">My</text>
		<text x="40" y="35" class="heavy">cat</text>
		<text x="55" y="55" class="small">is</text>
		<text x="65" y="55" class="Rrrrr">Grumpy!</text>
	</svg>
	`
	err = HtmlToPdf(f, utils.InputString(input), fontconfig)
	if err != nil {
		t.Fatal(err)
	}

	f.Close()

	_, err = file.ReadFile(path, nil)
	if err != nil {
		t.Fatal(err)
	}
}
