package goweasyprint

import (
	"io"
	"log"
	"os"
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

func TestRealPage(t *testing.T) {
	f, err := os.Create("/tmp/test_google.pdf")
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

	_, err = file.ReadFile("/tmp/test_google.pdf", nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSVG(t *testing.T) {
	f, err := os.Create("/tmp/test_svg.pdf")
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

	_, err = file.ReadFile("/tmp/test_svg.pdf", nil)
	if err != nil {
		t.Fatal(err)
	}
}
