package pdf

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/benoitkugler/pdf/model"
	"github.com/benoitkugler/pdf/reader"
	pdfParser "github.com/benoitkugler/pdf/reader/parser"
	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/css/parser"
	"github.com/benoitkugler/webrender/matrix"
	"github.com/benoitkugler/webrender/utils"
	"github.com/benoitkugler/webrender/utils/testutils"
)

func TestPaint(t *testing.T) {
	c := NewOutput()
	page := c.AddPage(0, 200, 100, 0)
	page.SetMediaBox(0, 200, 100, 0)
	page.State().SetColorRgba(parser.RGBA{R: 0, G: 1, B: 0, A: 1}, true)
	page.State().SetColorRgba(parser.RGBA{R: 0, G: 1, B: 1, A: 1}, false)
	// pdf.SetFont("Helvetica", "", 15)
	// pdf.AddPage()
	page.Rectangle(20, 20, 30, 30)
	page.State().SetLineWidth(2)
	page.Paint(backend.Stroke)
	page.OnNewStack(func() {
		page.Rectangle(20, 20, 30, 30)
		// page.Clip(false)
		page.Paint(backend.FillNonZero)
	})

	doc := c.Finalize()
	err := doc.Write(io.Discard, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGradientOp(t *testing.T) {
	c := NewOutput()
	page := c.AddPage(0, 200, 100, 0)
	page.State().Transform(matrix.New(1, 0, 0, -1, 0, 200)) // PDF uses "mathematical conventions"

	page.Rectangle(10, 10, 50, 50)
	page.State().SetColorRgba(parser.RGBA{R: 0, G: 0, B: 1, A: 1}, false)

	alpha := page.NewGroup(0, 0, 50, 50)
	alpha.State().SetColorRgba(parser.RGBA{R: 1, G: 1, B: 1, A: 1}, false)
	alpha.Rectangle(20, 20, 40, 20)
	alpha.Rectangle(18, 18, 44, 24)
	alpha.Paint(backend.FillEvenOdd)
	page.State().SetAlphaMask(alpha)

	page.Paint(backend.FillEvenOdd)
	// page.DrawGradient(backend.GradientLayout{
	// 	Positions: []fl{0, 0.5, 1},
	// 	Colors:    []parser.RGBA{{0, 0, 1, 1}, {1, 0, 0, 1}, {1, 0, 0, 0}},
	// 	GradientKind: backend.GradientKind{
	// 		Kind:   "linear",
	// 		Coords: [6]float32{0, 0, 50, 50},
	// 	},
	// 	ScaleY: 1,
	// }, 50, 50)

	doc := c.Finalize()
	err := doc.WriteFile("/tmp/op.pdf", nil)
	if err != nil {
		t.Fatal(err)
	}
}

// Test PDF-related code, including metadata, bookmarks && hyperlinks.
func modelToBytes(t *testing.T, doc model.Document) []byte {
	var target bytes.Buffer
	err := doc.Write(&target, nil)
	if err != nil {
		t.Fatal(err)
	}
	return target.Bytes()
}

func htmlToBytes(t *testing.T, html string) []byte {
	doc := htmlToModel(t, html)
	return modelToBytes(t, doc)
}

func TestPageSizeZoom(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, zoom := range [...]fl{1, 1.5, 0.5} {
		pdf := modelToBytes(t, htmlToModelExt(t, "<style>@page{size:3in 4in", zoom, "."))
		expected := fmt.Sprintf("/MediaBox [0 0 %d %d]", int(216*zoom), int(288*zoom))
		if !bytes.Contains(pdf, []byte(expected)) {
			t.Fatalf("invalid pdf output: %s", pdf)
		}
	}
}

// return the first submatch
func findRE(re string, pdf []byte) []string {
	ms := regexp.MustCompile(re).FindAllSubmatch(pdf, -1)
	out := make([]string, len(ms))
	for i, m := range ms {
		out[i] = string(m[1])
	}
	return out
}

func findCountAndTitles(pdf []byte) ([]string, []string) {
	counts := findRE("/Count ([0-9-]*)", pdf)
	titles := findRE("/Title \\((.*)\\)", pdf)
	return counts, titles
}

func TestBookmarks1(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pdf := htmlToBytes(t, `
      <h1>a</h1>  #
      <h4>b</h4>  ####
      <h3>c</h3>  ###
      <h2>d</h2>  ##
      <h1>e</h1>  #
    `)
	// a
	// |_ b
	// |_ c
	// L_ d
	// e

	counts, titles := findCountAndTitles(pdf)
	if lastMatch := counts[len(counts)-1]; lastMatch != "5" {
		t.Fatalf("unexpected /Count : %s", lastMatch)
	}

	if !reflect.DeepEqual(utils.NewSet(titles...), utils.NewSet("a", "b", "c", "d", "e")) {
		t.Fatalf("unexpected /Title : %s", titles)
	}
}

func TestBookmarks2(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pdf := htmlToBytes(t, "<body>")
	if bytes.Contains(pdf, []byte("Outlines")) {
		t.Fatalf("invalid output %s", pdf)
	}
}

func TestBookmarks3(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	pdf := htmlToBytes(t, "<h1>a nbsp…</h1>")
	titles := findRE("/Title \\((.*)\\)", pdf)

	o, err := pdfParser.ParseObject([]byte("(" + titles[0] + ")"))
	if err != nil {
		t.Fatal(err)
	}
	title, ok := model.IsString(o)
	if !ok {
		t.Fatalf("unexpected title %v", o)
	}
	if reader.DecodeTextString(title) != "a nbsp…" {
		t.Fatalf("unexpected title %s", title)
	}
}

func TestBookmarks4(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pdf := htmlToBytes(t, `
      <style>
        * { height: 90pt; margin: 0 0 10pt 0 }
      </style>
      <h1>1</h1>
      <h1>2</h1>
      <h2 style="position: relative; left: 20pt">3</h2>
      <h2>4</h2>
      <h3>5</h3>
      <span style="display: block; page-break-before: always"></span>
      <h2>6</h2>
      <h1>7</h1>
      <h2>8</h2>
      <h3>9</h3>
      <h1>10</h1>
      <h2>11</h2>
    `)
	// 1
	// 2
	// |_ 3
	// |_ 4
	// |  L_ 5
	// L_ 6
	// 7
	// L_ 8
	//    L_ 9
	// 10
	// L_ 11
	counts, titles := findCountAndTitles(pdf)
	if ms := utils.NewSet(titles...); !reflect.DeepEqual(ms,
		utils.NewSet("1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11")) {
		t.Fatalf("unexpected /Title %s", ms)
	}
	counts = counts[1:] // Page count
	if !reflect.DeepEqual(utils.NewSet(counts...), utils.NewSet("0", "4", "0", "1", "0", "0", "2", "1", "0", "1", "0", "11")) {
		t.Fatalf("unexpected /Count : %s", counts)
	}
}

func TestBookmarks5(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pdf := htmlToBytes(t, `
      <h2>1</h2> level 1
      <h4>2</h4> level 2
      <h2>3</h2> level 1
      <h3>4</h3> level 2
      <h4>5</h4> level 3
    `)
	// 1
	// L_ 2
	// 3
	// L_ 4
	//    L_ 5

	counts, titles := findCountAndTitles(pdf)
	if ms := utils.NewSet(titles...); !reflect.DeepEqual(ms,
		utils.NewSet("1", "2", "3", "4", "5")) {
		t.Fatalf("unexpected /Title %s", ms)
	}
	counts = counts[1:] // Page count
	if !reflect.DeepEqual(utils.NewSet(counts...), utils.NewSet("1", "0", "2", "1", "0", "5")) {
		t.Fatalf("unexpected /Count : %s", counts)
	}
}

func TestBookmarks6(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pdf := htmlToBytes(t, `
      <h2>1</h2> h2 level 1
      <h4>2</h4> h4 level 2
      <h3>3</h3> h3 level 2
      <h5>4</h5> h5 level 3
      <h1>5</h1> h1 level 1
      <h2>6</h2> h2 level 2
      <h2>7</h2> h2 level 2
      <h4>8</h4> h4 level 3
      <h1>9</h1> h1 level 1
    `)
	// 1
	// |_ 2
	// L_ 3
	//    L_ 4
	// 5
	// |_ 6
	// L_ 7
	//    L_ 8
	// 9

	counts, titles := findCountAndTitles(pdf)
	if ms := utils.NewSet(titles...); !reflect.DeepEqual(ms,
		utils.NewSet("1", "2", "3", "4", "5", "6", "7", "8", "9")) {
		t.Fatalf("unexpected /Title %s", ms)
	}
	counts = counts[1:] // Page count
	if !reflect.DeepEqual(utils.NewSet(counts...), utils.NewSet("3", "0", "1", "0", "3", "0", "1", "0", "0", "9")) {
		t.Fatalf("unexpected /Count : %s", counts)
	}
}

func TestBookmarks7(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	parseDest := func(de string) int {
		cs := strings.Split(strings.TrimSpace(de), " ")
		f, err := strconv.ParseFloat(cs[len(cs)-2], 64)
		if err != nil {
			t.Fatal(err)
		}
		return int(f)
	}

	// Reference for the next test. zoom=1
	pdf := htmlToBytes(t, "<h2>a</h2>")
	if ms := findRE("/Title \\((.*)\\)", pdf); !reflect.DeepEqual(ms, []string{"a"}) {
		t.Fatalf("unexpected /Title %s", ms)
	}
	dest := findRE("/Dest \\[(.*)\\]", pdf)[0]
	y := parseDest(dest)

	pdf = modelToBytes(t, htmlToModelExt(t, "<h2>a</h2>", 1.5, "."))
	if ms := findRE("/Title \\((.*)\\)", pdf); !reflect.DeepEqual(ms, []string{"a"}) {
		t.Fatalf("unexpected /Title %s", ms)
	}
	dest = findRE("/Dest \\[(.*)\\]", pdf)[0]
	y2 := parseDest(dest)
	if y2 != int(1.5*float64(y)) {
		t.Fatalf("invalid destination with zoom: %s", dest)
	}
}

func TestBookmarks8(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pdf := htmlToBytes(t, `
      <h1>a</h1>
      <h2>b</h2>
      <h3>c</h3>
      <h2 style="bookmark-state: closed">d</h2>
      <h3>e</h3>
      <h4>f</h4>
      <h1>g</h1>
    `)
	// a
	// |_ b
	// |  |_ c
	// |_ d (closed)
	// |  |_ e
	// |     |_ f
	// g

	counts, titles := findCountAndTitles(pdf)
	if ms := utils.NewSet(titles...); !reflect.DeepEqual(ms,
		utils.NewSet("a", "b", "c", "d", "e", "f", "g")) {
		t.Fatalf("unexpected /Title %s", ms)
	}
	counts = counts[1:] // Page count
	if !reflect.DeepEqual(utils.NewSet(counts...), utils.NewSet("3", "1", "0", "-2", "1", "0", "0", "5")) {
		t.Fatalf("unexpected /Count : %s", counts)
	}
}

func TestBookmarks9(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pdf := htmlToBytes(t, `
      <h1 style="bookmark-label: 'h1 on page ' counter(page)">a</h1>
    `)

	counts, titles := findCountAndTitles(pdf)
	if !reflect.DeepEqual(titles, []string{"h1 on page 1"}) {
		t.Fatalf("unexpected /Title %s", titles)
	}
	if counts[len(counts)-1] != "1" {
		t.Fatalf("unexpected /Count : %s", counts)
	}
}

func TestBookmarks10(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pdf := htmlToBytes(t, `
      <style>
      div:before, div:after {
         content: "";
         bookmark-level: 1;
         bookmark-label: "x";
      }
      </style>
      <div>a</div>
    `)
	// x
	// x
	counts, titles := findCountAndTitles(pdf)
	if counts[len(counts)-1] != "2" {
		t.Fatalf("unexpected /Count %v", counts)
	}
	if !reflect.DeepEqual(titles, []string{"x", "x"}) {
		t.Fatalf("unexpected /Title %v", titles)
	}
}

func TestBookmarks11(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pdf := htmlToBytes(t, `
      <div style="display:inline; white-space:pre;
       bookmark-level:1; bookmark-label:'a'">
      a
      a
      a
      </div>
      <div style="bookmark-level:1; bookmark-label:'b'">
        <div>b</div>
        <div style="break-before:always">c</div>
      </div>
    `)
	// a
	// b

	counts, titles := findCountAndTitles(pdf)
	if counts[len(counts)-1] != "2" {
		t.Fatalf("unexpected /Count %v", counts)
	}
	if !reflect.DeepEqual(titles, []string{"b", "a"}) {
		t.Fatalf("unexpected /Title %v", titles)
	}
}

func TestBookmarks12(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pdf := htmlToBytes(t, `
      <div style="bookmark-level:1; bookmark-label:contents">a</div>
    `)
	// a
	counts, titles := findCountAndTitles(pdf)
	if counts[len(counts)-1] != "1" {
		t.Fatalf("unexpected /Count %v", counts)
	}
	if !reflect.DeepEqual(titles, []string{"a"}) {
		t.Fatalf("unexpected /Title %v", titles)
	}
}

func TestBookmarks13(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pdf := htmlToBytes(t, `
      <div style="bookmark-level:1; bookmark-label:contents;
                  text-transform:uppercase">a</div>
    `)
	// a
	counts, titles := findCountAndTitles(pdf)
	if counts[len(counts)-1] != "1" {
		t.Fatalf("unexpected /Count %v", counts)
	}
	if !reflect.DeepEqual(titles, []string{"a"}) {
		t.Fatalf("unexpected /Title %v", titles)
	}
}

func TestBookmarks14(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	pdf := htmlToBytes(t, `
      <h1>a</h1>
      <h1> b c d </h1>
      <h1> e
             f </h1>
      <h1> g <span> h </span> i </h1>
    `)
	counts, titles := findCountAndTitles(pdf)
	if counts[len(counts)-1] != "4" {
		t.Fatalf("unexpected /Count %v", counts)
	}
	if !reflect.DeepEqual(titles, []string{"g h i", "e f", "b c d", "a"}) {
		t.Fatalf("unexpected /Title %v", titles)
	}
}

func TestLinksNone(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	pdf := htmlToBytes(t, "<body>")

	if bytes.Contains(pdf, []byte("Annots")) {
		t.Fatalf("invalid output %s", pdf)
	}
}

// Top and right positions in points, rounded to the default float precision of
// 6 digits, a rendered by pydyf
const (
	TOP   model.Fl = 297 * 72 / 25.4
	RIGHT model.Fl = 210 * 72 / 25.4
)

func round(x fl) fl {
	n := math.Pow10(4)
	return fl(math.Round(float64(x)*n) / n)
}

func roundRect(r model.Rectangle) model.Rectangle {
	return model.Rectangle{Llx: round(r.Llx), Lly: round(r.Lly), Urx: round(r.Urx), Ury: round(r.Ury)}
}

func TestLinks(t *testing.T) {
	// capt := testutils.CaptureLogs()
	// defer capt.AssertNoLogs(t)

	pdf := htmlToModel(t, `
      <style>
        body { margin: 0; font-size: 10pt; line-height: 2 }
        p { display: block; height: 90pt; margin: 0 0 10pt 0 }
        img { width: 30pt; vertical-align: top }
      </style>
      <p><a href="http://weasyprint.org"><img src=../resources_test/pattern.png></a></p>
      <p style="padding: 0 10pt"><a
         href="#lipsum"><img style="border: solid 1pt"
                             src=../resources_test/pattern.png></a></p>
      <p id=hello>Hello, World</p>
      <p id=lipsum>
        <a style="display: block; page-break-before: always; height: 30pt"
           href="#hel%6Co"></a>a
      </p>
    `)

	// uris := findRE("/URI \\((.*)\\)", pdf)
	// subtypes := findRE("/Subtype(.*)", pdf)
	// types := findRE("/S(.*)", pdf)

	// var rects [][4]fl
	// for _, match := range findRE("/Rect \\[([\\d\\.]+ [\\d\\.]+ [\\d\\.]+ [\\d\\.]+)\\]", pdf) {
	// 	fs := strings.Split(match, " ")
	// 	if len(fs) != 4 {
	// 		t.Fatalf("invalid rectangle %s", match)
	// 	}
	// 	var rect [4]fl
	// 	for i, f := range fs {
	// 		rect[i], _ = strconv.ParseFloat(f, 64)
	// 	}
	// 	rects = append(rects, rect)
	// }

	annots := pdf.Catalog.Pages.FlattenInherit()[0].Annots
	if len(annots) != 4 {
		t.Fatalf("expected 4 annotations, got %d", len(annots))
	}

	// 30pt wide (like the image), 20pt high (like line-height)
	if uri := annots[0].Subtype.(model.AnnotationLink).A.ActionType.(model.ActionURI).URI; uri != "http://weasyprint.org" {
		t.Fatalf("unexpected uris: %v", uri)
	}
	exp := roundRect(model.Rectangle{Llx: 0, Lly: TOP, Urx: 30, Ury: TOP - 20})
	if rect := roundRect(annots[0].Rect); rect != exp {
		t.Fatalf("unexpected rects: %#v %#v", rect, exp)
	}

	// The image itself: 30*30pt
	if uri := annots[1].Subtype.(model.AnnotationLink).A.ActionType.(model.ActionURI).URI; uri != "http://weasyprint.org" {
		t.Fatalf("unexpected uris: %v", uri)
	}
	exp = roundRect(model.Rectangle{Llx: 0, Lly: TOP, Urx: 30, Ury: TOP - 30})
	if rect := roundRect(annots[1].Rect); rect != exp {
		t.Fatalf("unexpected rects: %#v %#v", rect, exp)
	}

	// 32pt wide (image + 2 * 1pt of border), 20pt high
	if uri := annots[2].Subtype.(model.AnnotationLink).Dest.(model.DestinationString); uri != "lipsum" {
		t.Fatalf("unexpected uris: %v", uri)
	}
	dest := pdf.Catalog.Names.Dests.LookupTable()["lipsum"]
	location := dest.(model.DestinationExplicitIntern).Location.(model.DestinationLocationXYZ)
	location.Top = model.ObjFloat(round(fl(location.Top.(model.ObjFloat))))
	if location != (model.DestinationLocationXYZ{Left: model.ObjFloat(0), Top: model.ObjFloat(round(TOP)), Zoom: 0}) {
		t.Fatalf("unexpected destination: %v", dest)
	}
	exp = roundRect(model.Rectangle{Llx: 10, Lly: TOP - 100, Urx: 10 + 32, Ury: TOP - 100 - 20})
	if rect := roundRect(annots[2].Rect); rect != exp {
		t.Fatalf("unexpected rects: %#v %#v", rect, exp)
	}

	// The image itself: 32*32pt
	_ = annots[3].Subtype.(model.AnnotationLink)
	exp = roundRect(model.Rectangle{Llx: 10, Lly: TOP - 100, Urx: 10 + 32, Ury: TOP - 100 - 32})
	if rect := roundRect(annots[3].Rect); rect != exp {
		t.Fatalf("unexpected rects: %#v %#v", rect, exp)
	}

	annots = pdf.Catalog.Pages.FlattenInherit()[1].Annots
	if len(annots) != 1 {
		t.Fatalf("expected 1 annotations, got %d", len(annots))
	}

	// 100% wide (block), 30pt high
	if uri := annots[0].Subtype.(model.AnnotationLink).Dest.(model.DestinationString); uri != "hello" {
		t.Fatalf("unexpected uris: %v", uri)
	}
	dest = pdf.Catalog.Names.Dests.LookupTable()["hello"]
	location = dest.(model.DestinationExplicitIntern).Location.(model.DestinationLocationXYZ)
	location.Top = model.ObjFloat(round(fl(location.Top.(model.ObjFloat))))
	if location != (model.DestinationLocationXYZ{Left: model.ObjFloat(0), Top: model.ObjFloat(round(TOP - 200)), Zoom: 0}) {
		t.Fatalf("unexpected destination: %v", dest)
	}
	exp = roundRect(model.Rectangle{Llx: 0, Lly: TOP, Urx: RIGHT, Ury: TOP - 30})
	if rect := roundRect(annots[0].Rect); rect != exp {
		t.Fatalf("unexpected rects: %#v %#v", rect, exp)
	}
}

func TestSortedLinks(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1352
	pdf := htmlToBytes(t, `
      <p id="zzz">zzz</p>
      <p id="aaa">aaa</p>
      <a href="#zzz">z</a>
      <a href="#aaa">a</a>
    `)
	l := strings.Split(string(pdf), "(aaa) [")
	if !strings.Contains(l[len(l)-1], "(zzz) [") {
		t.Fatal()
	}
}

func TestRelativeLinksNoHeight(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// 100% wide (block), 0pt high
	pdf := htmlToModelExt(t, `<a href="../lipsum" style="display: block"></a>a`,
		1, "http://weasyprint.org/foo/bar/")

	annots := pdf.Catalog.Pages.FlattenInherit()[0].Annots
	if len(annots) != 1 {
		t.Fatalf("expected 1 annotations, got %d", len(annots))
	}

	if uri := annots[0].Subtype.(model.AnnotationLink).A.ActionType.(model.ActionURI).URI; uri != "http://weasyprint.org/foo/lipsum" {
		t.Fatalf("unexpected uris: %v", uri)
	}
	exp := roundRect(model.Rectangle{Llx: 0, Lly: TOP, Urx: RIGHT, Ury: TOP})
	if rect := roundRect(annots[0].Rect); rect != exp {
		t.Fatalf("unexpected rects: %#v %#v", rect, exp)
	}
}

func TestRelativeLinksMissingBase(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Relative URI reference without a base URI
	pdf := htmlToModelExt(t, `<a href="../lipsum" style="display: block"></a>a`,
		1, "")

	annots := pdf.Catalog.Pages.FlattenInherit()[0].Annots
	if len(annots) != 1 {
		t.Fatalf("expected 1 annotations, got %d", len(annots))
	}

	if uri := annots[0].Subtype.(model.AnnotationLink).A.ActionType.(model.ActionURI).URI; uri != "../lipsum" {
		t.Fatalf("unexpected uris: %v", uri)
	}
	exp := roundRect(model.Rectangle{Llx: 0, Lly: TOP, Urx: RIGHT, Ury: TOP})
	if rect := roundRect(annots[0].Rect); rect != exp {
		t.Fatalf("unexpected rects: %#v %#v", rect, exp)
	}
}

func TestRelativeLinksMissingBaseLink(t *testing.T) {
	// Relative URI reference without a base URI: ! supported for -weasy-link
	capt := testutils.CaptureLogs()
	pdf := modelToBytes(t, htmlToModelExt(t, `<div style="-weasy-link: url(../lipsum)">`,
		1, ""))
	logs := capt.Logs()

	if bytes.Contains(pdf, []byte("/Annots")) {
		t.Fatal()
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 logs, got %d", len(logs))
	}
	if !strings.Contains(logs[0], "Ignored `-weasy-link: url(../lipsum)`") || !strings.Contains(logs[0], "Relative URI reference without a base URI") {
		t.Fatalf("invalid logs %v", logs)
	}
}

func TestRelativeLinksInternal(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	// Internal URI reference without a base URI: OK
	pdf := htmlToModelExt(t, `<a href="#lipsum" id="lipsum" style="display: block"></a>a`,
		1, "")

	annots := pdf.Catalog.Pages.FlattenInherit()[0].Annots
	if len(annots) != 1 {
		t.Fatalf("expected 1 annotations, got %d", len(annots))
	}

	if uri := annots[0].Subtype.(model.AnnotationLink).Dest.(model.DestinationString); uri != "lipsum" {
		t.Fatalf("unexpected uris: %v", uri)
	}
	dest := pdf.Catalog.Names.Dests.LookupTable()["lipsum"]
	location := dest.(model.DestinationExplicitIntern).Location.(model.DestinationLocationXYZ)
	location.Top = model.ObjFloat(round(fl(location.Top.(model.ObjFloat))))
	if location != (model.DestinationLocationXYZ{Left: model.ObjFloat(0), Top: model.ObjFloat(round(TOP)), Zoom: 0}) {
		t.Fatalf("unexpected destination: %v", location)
	}
	exp := roundRect(model.Rectangle{Llx: 0, Lly: TOP, Urx: RIGHT, Ury: TOP})
	if rect := roundRect(annots[0].Rect); rect != exp {
		t.Fatalf("unexpected rects: %#v %#v", rect, exp)
	}
}

func TestRelativeLinksAnchors(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pdf := htmlToModelExt(t, `<div style="-weasy-link: url(#lipsum)" id="lipsum"></div>a`,
		1, "")

	annots := pdf.Catalog.Pages.FlattenInherit()[0].Annots
	if len(annots) != 1 {
		t.Fatalf("expected 1 annotations, got %d", len(annots))
	}

	if uri := annots[0].Subtype.(model.AnnotationLink).Dest.(model.DestinationString); uri != "lipsum" {
		t.Fatalf("unexpected uris: %v", uri)
	}
	dest := pdf.Catalog.Names.Dests.LookupTable()["lipsum"]
	location := dest.(model.DestinationExplicitIntern).Location.(model.DestinationLocationXYZ)
	location.Top = model.ObjFloat(round(fl(location.Top.(model.ObjFloat))))
	if location != (model.DestinationLocationXYZ{Left: model.ObjFloat(0), Top: model.ObjFloat(round(TOP)), Zoom: 0}) {
		t.Fatalf("unexpected destination: %v", dest)
	}
	exp := roundRect(model.Rectangle{Llx: 0, Lly: TOP, Urx: RIGHT, Ury: TOP})
	if rect := roundRect(annots[0].Rect); rect != exp {
		t.Fatalf("unexpected rects: %#v %#v", rect, exp)
	}
}

func TestRelativeLinksDifferentBase(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pdf := modelToBytes(t, htmlToModelExt(t, `<a href="/test/lipsum"></a>a`,
		1, "http://weasyprint.org/foo/bar/"))
	if !strings.Contains(string(pdf), "http://weasyprint.org/test/lipsum") {
		t.Fatal()
	}
}

func TestRelativeLinksSameBase(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pdf := modelToBytes(t, htmlToModelExt(t, `<a id="test" href="/foo/bar/#test"></a>a`,
		1, "http://weasyprint.org/foo/bar/"))
	if !strings.Contains(string(pdf), "/Dest (test)") {
		t.Fatal()
	}
}

func TestMissingLinks(t *testing.T) {
	capt := testutils.CaptureLogs()
	pdf := htmlToModelExt(t, `
          <style> a { display: block; height: 15pt } </style>
          <a href="#lipsum"></a>
          <a href="#missing" id="lipsum"></a>
          <a href=""></a>a
        `, 1, "")
	logs := capt.Logs()
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	if !strings.Contains(logs[0], "No anchor #missing for internal URI reference") {
		t.Fatalf("unexpected log %s", logs[0])
	}

	annots := pdf.Catalog.Pages.FlattenInherit()[0].Annots
	if len(annots) != 1 {
		t.Fatalf("expected 1 annotations, got %d", len(annots))
	}

	if uri := annots[0].Subtype.(model.AnnotationLink).Dest.(model.DestinationString); uri != "lipsum" {
		t.Fatalf("unexpected uris: %v", uri)
	}
	dest := pdf.Catalog.Names.Dests.LookupTable()["lipsum"]
	location := dest.(model.DestinationExplicitIntern).Location.(model.DestinationLocationXYZ)
	location.Top = model.ObjFloat(round(fl(location.Top.(model.ObjFloat))))
	if location != (model.DestinationLocationXYZ{Left: model.ObjFloat(0), Top: model.ObjFloat(round(TOP - 15)), Zoom: 0}) {
		t.Fatalf("unexpected destination: %v", dest)
	}
	exp := roundRect(model.Rectangle{Llx: 0, Lly: TOP, Urx: RIGHT, Ury: TOP - 15})
	if rect := roundRect(annots[0].Rect); rect != exp {
		t.Fatalf("unexpected rects: %#v %#v", rect, exp)
	}
}

func TestEmbedGif(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pdf := htmlToBytes(t, `<img src="../resources_test/pattern.gif">`)
	if !bytes.Contains(pdf, []byte("/Filter [/FlateDecode]")) {
		t.Fatal()
	}
}

func TestEmbedJpeg(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	// JPEG-encoded image, embedded in PDF
	pdf := htmlToBytes(t, `<img src="../resources_test/blue.jpg">`)
	if !bytes.Contains(pdf, []byte("/Filter [/DCTDecode]")) {
		t.Fatal()
	}
}

func TestEmbedImageOnce(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Image repeated multiple times, embedded once
	pdf := htmlToBytes(t, `
	<img src="../resources_test/blue.jpg">
	<div style="background: url(../resources_test/blue.jpg)"></div>
	<img src="../resources_test/blue.jpg">
	<div style="background: url(../resources_test/blue.jpg) no-repeat"></div>
  `)
	if bytes.Count(pdf, []byte("/Filter [/DCTDecode]")) != 1 {
		t.Fatal()
	}
}

func TestDocumentInfo(t *testing.T) {
	// capt := testutils.CaptureLogs()
	// defer capt.AssertNoLogs(t)

	pdf := htmlToModel(t, `
      <meta name=author content="I Me &amp; Myself">
      <title>Test document</title>
      <h1>Another title</h1>
      <meta name=generator content="Human after all">
      <meta name=keywords content="html ,	css,
                                   pdf,css">
      <meta name=description content="Blah… ">
      <meta name=dcterms.created content=2011-04-21T23:00:00Z>
      <meta name=dcterms.modified content=2013-07-21T23:46+01:00>
    `)
	if s := pdf.Trailer.Info.Title; s != "Test document" {
		t.Fatalf("unexpected Title %v", s)
	}
	if s := pdf.Trailer.Info.Subject; s != "Blah… " {
		t.Fatalf("unexpected Subject %v", s)
	}
	if s := pdf.Trailer.Info.Author; s != "I Me & Myself" {
		t.Fatalf("unexpected Author %v", s)
	}
	if s := pdf.Trailer.Info.Keywords; s != "html, css, pdf" {
		t.Fatalf("unexpected Keywords %v", s)
	}
	if s := pdf.Trailer.Info.Creator; s != "Human after all" {
		t.Fatalf("unexpected Creator %v", s)
	}
	if model.DateTimeString(pdf.Trailer.Info.CreationDate) != "D:20110421230000+00'00'" {
		t.Fatalf("unexpected CreationDate %s", pdf.Trailer.Info.CreationDate)
	}
	if model.DateTimeString(pdf.Trailer.Info.ModDate) != "D:20130721234600+01'00'" {
		t.Fatalf("unexpected ModDate %s", pdf.Trailer.Info.ModDate)
	}
}

func TestEmbeddedFilesAttachments(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	f, err := os.CreateTemp("", "test_pdf_attachements*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	defer os.Remove(f.Name())

	_, err = io.Copy(f, strings.NewReader("12345678"))
	if err != nil {
		t.Fatal(err)
	}

	absoluteUrl, err := utils.PathToURL(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	f2, err := os.CreateTemp("", "test_pdf_äöü*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f2.Close()
	defer os.Remove(f2.Name())

	_, err = io.Copy(f2, strings.NewReader("abcdefgh"))
	if err != nil {
		t.Fatal(err)
	}

	pdf := modelToBytes(t, htmlToModelExt2(t, fmt.Sprintf(`
          <title>Test document</title>
          <meta charset="utf-8">
          <link
            rel="attachment"
            title="some file attachment äöü"
            href="data:,hi%%20there">
          <link rel="attachment" href="%s">
          <link rel="attachment" href="%s">
          <h1>Heading 1</h1>
          <h2>Heading 2</h2>
        `, absoluteUrl, filepath.Base(f2.Name())), 1, os.TempDir(), []backend.Attachment{
		{Content: []byte("oob attachment"), Title: "Hello"},
		{Content: []byte("raw URL")},
		{Content: []byte("file like obj"), Description: "Hello"},
	},
	))

	if !bytes.Contains(pdf, []byte("/UF (attachment.bin)")) {
		t.Fatal("missing /UF (attachment.bin)")
	}
	if !bytes.Contains(pdf, []byte("/Desc (Hello)")) {
		t.Fatal("missing /Desc (Hello)")
	}
	if !bytes.Contains(pdf, []byte("/EmbeddedFiles")) {
		t.Fatal("missing /EmbeddedFiles")
	}
	if !bytes.Contains(pdf, []byte("/Outlines")) {
		t.Fatal("missing /Outlines")
	}

	// name = BOMUTF16BE + "some file attachment äöü".encode("utf-16-be")
	// assert b"/Desc <" + name.hex().encode("ascii") + b">" := range pdf

	// assert hashlib.md5(adata).hexdigest().encode("ascii") := range pdf
	// assert os.path.basename(absoluteTmpFile).encode("ascii") := range pdf

	// assert hashlib.md5(rdata).hexdigest().encode("ascii") := range pdf
	// name = BOMUTF16BE + "some file attachment äöü".encode("utf-16-be")
	// assert b"/Desc <" + name.hex().encode("ascii") + b">" := range pdf

	// assert hashlib.md5(b"oob attachment").hexdigest().encode("ascii") := range pdf
	// assert hashlib.md5(b"raw URL").hexdigest().encode("ascii") := range pdf
	// assert hashlib.md5(b"file like obj").hexdigest().encode("ascii") := range pdf
}

func TestAttachmentsData(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pdf := htmlToBytes(t, `
      <title>Test document 2</title>
      <meta charset="utf-8">
      <link rel="attachment" href="data:,some-data">
    `)
	sum := md5.Sum([]byte("some-data"))
	md5 := model.EspaceHexString(sum[:])
	if !bytes.Contains(pdf, []byte(md5)) {
		t.Fatal("missing or unexpected checksum")
	}
}

func TestAttachmentsNone(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pdf := htmlToBytes(t, `
      <title>Test document 3</title>
      <meta charset="utf-8">
      <h1>Heading</h1>
    `)
	if bytes.Contains(pdf, []byte("Names")) {
		t.Fatal("unexpected Names")
	}
	if !bytes.Contains(pdf, []byte("Outlines")) {
		t.Fatal("missing Outlines")
	}
}

func TestAttachmentsNoneEmpty(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	pdf := htmlToBytes(t, `
      <title>Test document 3</title>
      <meta charset="utf-8">
    `)
	if bytes.Contains(pdf, []byte("Names")) {
		t.Fatal("unexpected Names")
	}
	if bytes.Contains(pdf, []byte("Outlines")) {
		t.Fatal("unexpected Outlines")
	}
}

func TestAnnotations(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pdf := htmlToBytes(t, `
      <title>Test document</title>
      <meta charset="utf-8">
      <a
        rel="attachment"
        href="data:,some-data"
        download>A link that lets you download an attachment</a>
      <a
        rel="attachment"
        href="data:,some-data"
        download>A link that lets you download an attachment</a>
    `)

	sum := md5.Sum([]byte("some-data"))
	md5 := model.EspaceHexString(sum[:])
	if !bytes.Contains(pdf, []byte(md5)) {
		t.Fatal("missing or unexpected checksum")
	}
	if !bytes.Contains(pdf, []byte("FileAttachment")) {
		t.Fatal("missing Names")
	}
	if bytes.Contains(pdf, []byte("EmbeddedFiles")) {
		t.Fatal("unexpected EmbeddedFiles")
	}
}

func TestBleed(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, data := range []struct {
		style              string
		media, bleed, trim model.Rectangle
	}{
		{
			"bleed: 30pt; size: 10pt",
			model.Rectangle{Llx: -30, Lly: -30, Urx: 40, Ury: 40},
			model.Rectangle{Llx: -10, Lly: -10, Urx: 20, Ury: 20},
			model.Rectangle{Llx: 0, Lly: 0, Urx: 10, Ury: 10},
		},
		{
			"bleed: 15pt 3pt 6pt 18pt; size: 12pt 15pt",
			model.Rectangle{Llx: -18, Lly: -15, Urx: 15, Ury: 21},
			model.Rectangle{Llx: -10, Lly: -10, Urx: 15, Ury: 21},
			model.Rectangle{Llx: 0, Lly: 0, Urx: 12, Ury: 15},
		},
	} {
		pdf := htmlToModel(t, fmt.Sprintf(`
      <title>Test document</title>
      <style>@page { %s }</style>
      <body>test
    `, data.style))
		page := pdf.Catalog.Pages.FlattenInherit()[0]
		if *page.MediaBox != data.media {
			t.Fatalf("unexpected /MediaBox %v", page.MediaBox)
		}
		if *page.BleedBox != data.bleed {
			t.Fatalf("unexpected /BleedBox %v", page.BleedBox)
		}
		if *page.TrimBox != data.trim {
			t.Fatalf("unexpected /TrimBox %v", page.TrimBox)
		}
	}
}
