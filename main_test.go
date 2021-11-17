package goweasyprint

import (
	"io"
	"log"
	"os"
	"testing"

	"github.com/benoitkugler/go-weasyprint/layout/text"
	"github.com/benoitkugler/go-weasyprint/logger"
	"github.com/benoitkugler/go-weasyprint/utils"
	"github.com/benoitkugler/pdf/reader/file"
	fc "github.com/benoitkugler/textlayout/fontconfig"
	"github.com/benoitkugler/textlayout/pango/fcfonts"
)

const fontmapCache = "layout/text/test/cache.fc"

var fontconfig *text.FontConfiguration

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

func TestRealPage(t *testing.T) {
	f, err := os.Create("/tmp/test_google.pdf")
	if err != nil {
		t.Fatal(err)
	}

	url := "http://www.google.com"
	// url := "https://weasyprint.org/"
	// url := "https://en.wikipedia.org/wiki/Go_(programming_language)" // rather big document
	// url := "https://golang.org/doc/go1.17"                           // slow because of text layout
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
