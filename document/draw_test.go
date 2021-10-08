package document

import (
	"io/ioutil"
	"log"
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

func renderUrl(t *testing.T, url string) {
	doc, err := tree.NewHTML(utils.InputUrl(url), "", nil, "")
	if err != nil {
		t.Fatal(err)
	}
	finalDoc := Render(doc, nil, true, fc)
	finalDoc.WriteDocument(output{}, 1, nil)
}

func TestRealPage(t *testing.T) {
	// renderUrl(t, "http://www.google.com")
	// FIXME:
	renderUrl(t, "https://weasyprint.org/")
}
