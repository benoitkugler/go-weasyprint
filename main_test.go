package goweasyprint

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/benoitkugler/go-weasyprint/pdf/test"
	"github.com/benoitkugler/pdf/reader/file"
	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/text"
	"github.com/benoitkugler/webrender/utils"
	"github.com/go-text/typesetting/opentype/loader"
)

var fontconfig text.FontConfiguration

func init() {
	logger.ProgressLogger.SetOutput(io.Discard)

	var err error
	fontconfig, err = test.LoadTestFontConfig("pdf/test/")
	if err != nil {
		log.Fatalf("creating font configuration: %s", err)
	}
}

func tempFile(s string) string {
	dir := os.TempDir()
	return filepath.Join(dir, s)
}

func TestRealPage(t *testing.T) {
	t.Skip()

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

func TestSVGGradient(t *testing.T) {
	path := tempFile("test_svg_gradient.pdf")

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	input := `
	<svg width="500" height="500" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">

	
	<radialGradient id="rg2" cx="50%" cy="50%"   r="40%" gradientUnits="objectBoundingBox"  
	spreadMethod="repeat" >
	<stop offset="10%" stop-color="goldenrod" />
	<stop offset="30%" stop-color="seagreen" />
	<stop offset="50%" stop-color="cyan" />
	<stop offset="70%" stop-color="black" />
	<stop offset="100%" stop-color="orange" />
	</radialGradient>

	<ellipse cx="300" cy="150" rx="120" ry="100"  style="fill:url(#rg2)" /> 

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

func TestTmp(t *testing.T) {
	f, _ := os.Open("resources_test/weasyprint.otb")
	ft, err := loader.NewLoader(f)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(ft.RawTable(loader.MustNewTag("loca")))
	fmt.Println(ft.HasTable(loader.MustNewTag("glyf")))
	var allTables []loader.Table
	for _, table := range ft.Tables() {
		if table == loader.MustNewTag("loca") || table == loader.MustNewTag("glyf") {
			continue
		}
		fmt.Println(table)
		b, err := ft.RawTable(table)
		if err != nil {
			t.Fatal(err)
		}
		allTables = append(allTables, loader.Table{
			Tag:     table,
			Content: b,
		})
	}
	err = os.WriteFile("weasyprint.otb_fixed", loader.WriteTTF(allTables), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
}
