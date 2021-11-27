package pdf

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/benoitkugler/go-weasyprint/style/parser"
	"github.com/benoitkugler/go-weasyprint/utils"
	"github.com/benoitkugler/go-weasyprint/utils/testutils"
	"github.com/benoitkugler/pdf/model"
	"github.com/benoitkugler/pdf/reader"
	pdfParser "github.com/benoitkugler/pdf/reader/parser"
)

func TestPaint(t *testing.T) {
	c := NewOutput()
	page := c.AddPage(0, 200, 100, 0)
	page.SetMediaBox(0, 200, 100, 0)
	page.SetColorRgba(parser.RGBA{R: 0, G: 1, B: 0, A: 1}, true)
	page.SetColorRgba(parser.RGBA{R: 0, G: 1, B: 1, A: 1}, false)
	// pdf.SetFont("Helvetica", "", 15)
	// pdf.AddPage()
	page.Rectangle(20, 20, 30, 30)
	page.SetLineWidth(2)
	page.Stroke()
	page.OnNewStack(func() error {
		page.Rectangle(20, 20, 30, 30)
		// page.Clip(false)
		page.Fill(false)
		return nil
	})

	doc := c.Finalize()
	err := doc.Write(io.Discard, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// Test PDF-related code, including metadata, bookmarks && hyperlinks.

// Top && right positions in points, rounded to the default float precision of
// 6 digits, a rendered by pydyf
// var (
// 	TOP   = round(297*72/25.4, 6)
// 	RIGHT = round(210*72/25.4, 6)
// )

func htmlToBytes(t *testing.T, html string, zoom fl) []byte {
	doc := htmlToModel(t, html, zoom)
	var target bytes.Buffer
	err := doc.Write(&target, nil)
	if err != nil {
		t.Fatal(err)
	}
	return target.Bytes()
}

func TestPageSizeZoom(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, zoom := range [...]fl{1, 1.5, 0.5} {
		pdf := htmlToBytes(t, "<style>@page{size:3in 4in", zoom)
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

func TestBookmarks1(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pdf := htmlToBytes(t, `
      <h1>a</h1>  #
      <h4>b</h4>  ####
      <h3>c</h3>  ###
      <h2>d</h2>  ##
      <h1>e</h1>  #
    `, 1)
	// a
	// |_ b
	// |_ c
	// L_ d
	// e

	ms := findRE("/Count ([0-9-]*)", pdf)
	if lastMatch := ms[len(ms)-1]; lastMatch != "5" {
		t.Fatalf("unexpected /Count : %s", lastMatch)
	}

	ms = findRE("/Title \\((.*)\\)", pdf)
	if !reflect.DeepEqual(utils.NewSet(ms...), utils.NewSet("a", "b", "c", "d", "e")) {
		t.Fatalf("unexpected /Title : %s", ms)
	}
}

func TestBookmarks2(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pdf := htmlToBytes(t, "<body>", 1)
	if bytes.Contains(pdf, []byte("Outlines")) {
		t.Fatalf("invalid output %s", pdf)
	}
}

func TestBookmarks3(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	pdf := htmlToBytes(t, "<h1>a nbsp…</h1>", 1)
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
    `, 1)
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
	if ms := utils.NewSet(findRE("/Title \\((.*)\\)", pdf)...); !reflect.DeepEqual(ms,
		utils.NewSet("1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11")) {
		t.Fatalf("unexpected /Title %s", ms)
	}
	counts := findRE("/Count ([0-9-]*)", pdf)
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
    `, 1)
	// 1
	// L_ 2
	// 3
	// L_ 4
	//    L_ 5

	if ms := utils.NewSet(findRE("/Title \\((.*)\\)", pdf)...); !reflect.DeepEqual(ms,
		utils.NewSet("1", "2", "3", "4", "5")) {
		t.Fatalf("unexpected /Title %s", ms)
	}
	counts := findRE("/Count ([0-9-]*)", pdf)
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
    `, 1)
	// 1
	// |_ 2
	// L_ 3
	//    L_ 4
	// 5
	// |_ 6
	// L_ 7
	//    L_ 8
	// 9

	if ms := utils.NewSet(findRE("/Title \\((.*)\\)", pdf)...); !reflect.DeepEqual(ms,
		utils.NewSet("1", "2", "3", "4", "5", "6", "7", "8", "9")) {
		t.Fatalf("unexpected /Title %s", ms)
	}
	counts := findRE("/Count ([0-9-]*)", pdf)
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
	pdf := htmlToBytes(t, "<h2>a</h2>", 1)
	if ms := findRE("/Title \\((.*)\\)", pdf); !reflect.DeepEqual(ms, []string{"a"}) {
		t.Fatalf("unexpected /Title %s", ms)
	}
	dest := findRE("/Dest \\[(.*)\\]", pdf)[0]
	y := parseDest(dest)

	pdf = htmlToBytes(t, "<h2>a</h2>", 1.5)
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
    `, 1)
	// a
	// |_ b
	// |  |_ c
	// |_ d (closed)
	// |  |_ e
	// |     |_ f
	// g

	if ms := utils.NewSet(findRE("/Title \\((.*)\\)", pdf)...); !reflect.DeepEqual(ms,
		utils.NewSet("a", "b", "c", "d", "e", "f", "g")) {
		t.Fatalf("unexpected /Title %s", ms)
	}
	counts := findRE("/Count ([0-9-]*)", pdf)
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
    `, 1)

	if ms := findRE("/Title \\((.*)\\)", pdf); !reflect.DeepEqual(ms, []string{"h1 on page 1"}) {
		t.Fatalf("unexpected /Title %s", ms)
	}
	counts := findRE("/Count ([0-9-]*)", pdf)
	if counts[len(counts)-1] != "1" {
		t.Fatalf("unexpected /Count : %s", counts)
	}
}

// TODO:
// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestBookmarks10(t *testing.T) {
//     pdf := htmlToBytes(t, `
//       <style>
//       div:before, div:after {
//          content: "";
//          bookmark-level: 1;
//          bookmark-label: "x";
//       }
//       </style>
//       <div>a</div>
//     `).writePdf()
//     // x
//     // x
//     counts = re.findall(b"/Count ([0-9-]*)", pdf)
//     outlines = counts.pop()
//     assert outlines == b"2"
//     assert re.findall(b"/Title \\((.*)\\)", pdf) == [b"x", b"x"]

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestBookmarks11(t *testing.T){
//     pdf := htmlToBytes(t, `
//       <div style="display:inline; white-space:pre;
//        bookmark-level:1; bookmark-label:"a"">
//       a
//       a
//       a
//       </div>
//       <div style="bookmark-level:1; bookmark-label:"b"">
//         <div>b</div>
//         <div style="break-before:always">c</div>
//       </div>
//     `).writePdf()
//     // a
//     // b
//     counts = re.findall(b"/Count ([0-9-]*)", pdf)
//     outlines = counts.pop()
//     assert outlines == b"2"
//     assert re.findall(b"/Title \\((.*)\\)", pdf) == [b"a", b"b"]
// }

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestBookmarks12(t *testing.T) {
//     pdf := htmlToBytes(t, `
//       <div style="bookmark-level:1; bookmark-label:contents">a</div>
//     `).writePdf()
//     // a
//     counts = re.findall(b"/Count ([0-9-]*)", pdf)
//     outlines = counts.pop()
//     assert outlines == b"1"
//     assert re.findall(b"/Title \\((.*)\\)", pdf) == [b"a"]

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestBookmarks13(t *testing.T){
//     pdf := htmlToBytes(t, `
//       <div style="bookmark-level:1; bookmark-label:contents;
//                   text-transform:uppercase">a</div>
//     `).writePdf()
//     // a
//     counts = re.findall(b"/Count ([0-9-]*)", pdf)
//     outlines = counts.pop()
//     assert outlines == b"1"
//     assert re.findall(b"/Title \\((.*)\\)", pdf) == [b"a"]
// }

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestBookmarks14(t *testing.T) {
//     pdf := htmlToBytes(t, `
//       <h1>a</h1>
//       <h1> b c d </h1>
//       <h1> e
//              f </h1>
//       <h1> g <span> h </span> i </h1>
//     `).writePdf()
//     assert re.findall(b"/Count ([0-9-]*)", pdf)[-1] == b"4"
//     assert re.findall(b"/Title \\((.*)\\)", pdf) == [
//         b"a", b"b c d", b"e f", b"g h i"]

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestLinksNone(t *testing.T){
//     pdf := htmlToBytes(t, "<body>").writePdf()
//     assert b"Annots" ! := range pdf

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestLinks(t *testing.T){
//     pdf := htmlToBytes(t, `
//       <style>
//         body { margin: 0; font-size: 10pt; line-height: 2 }
//         p { display: block; height: 90pt; margin: 0 0 10pt 0 }
//         img { width: 30pt; vertical-align: top }
//       </style>
//       <p><a href="http://weasyprint.org"><img src=pattern.png></a></p>
//       <p style="padding: 0 10pt"><a
//          href="#lipsum"><img style="border: solid 1pt"
//                              src=pattern.png></a></p>
//       <p id=hello>Hello, World</p>
//       <p id=lipsum>
//         <a style="display: block; page-break-before: always; height: 30pt"
//            href="#hel%6Co"></a>a
//       </p>
//     `, baseUrl=resourceFilename("<inline HTML>")).writePdf()
// }
//     uris = re.findall(b"/URI \\((.*)\\)", pdf)
//     types = re.findall(b"/S (.*)", pdf)
//     subtypes = re.findall(b"/Subtype (.*)", pdf)
//     rects = [
//         [float(number) for number := range match.split()] for match := range re.findall(
//             b"/Rect \\[ ([\\d\\.]+ [\\d\\.]+ [\\d\\.]+ [\\d\\.]+) \\]", pdf)]

//     // 30pt wide (like the image), 20pt high (like line-height)
//     assert uris.pop(0) == b"http://weasyprint.org"
//     assert subtypes.pop(0) == b"/Link"
//     assert types.pop(0) == b"/URI"
//     assert rects.pop(0) == [0, TOP, 30, TOP - 20]

//     // The image itself: 30*30pt
//     assert uris.pop(0) == b"http://weasyprint.org"
//     assert subtypes.pop(0) == b"/Link"
//     assert types.pop(0) == b"/URI"
//     assert rects.pop(0) == [0, TOP, 30, TOP - 30]

//     // 32pt wide (image + 2 * 1pt of border), 20pt high
//     assert subtypes.pop(0) == b"/Link"
//     assert b"/Dest (lipsum)" := range pdf
//     link = re.search(
//         b"\\(lipsum\\) \\[ \\d+ 0 R /XYZ ([\\d\\.]+ [\\d\\.]+ [\\d\\.]+) ]",
//         pdf).group(1)
//     assert [float(number) for number := range link.split()] == [0, TOP, 0]
//     assert rects.pop(0) == [10, TOP - 100, 10 + 32, TOP - 100 - 20]

//     // The image itself: 32*32pt
//     assert subtypes.pop(0) == b"/Link"
//     assert rects.pop(0) == [10, TOP - 100, 10 + 32, TOP - 100 - 32]

//     // 100% wide (block), 30pt high
//     assert subtypes.pop(0) == b"/Link"
//     assert b"/Dest (hello)" := range pdf
//     link = re.search(
//         b"\\(hello\\) \\[ \\d+ 0 R /XYZ ([\\d\\.]+ [\\d\\.]+ [\\d\\.]+) ]",
//         pdf).group(1)
//     assert [float(number) for number := range link.split()] == [0, TOP - 200, 0]
//     assert rects.pop(0) == [0, TOP, RIGHT, TOP - 30]

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestSortedLinks(t *testing.T) {
//     // Regression test for https://github.com/Kozea/WeasyPrint/issues/1352
//     pdf := htmlToBytes(t, `
//       <p id="zzz">zzz</p>
//       <p id="aaa">aaa</p>
//       <a href="#zzz">z</a>
//       <a href="#aaa">a</a>
//     `, baseUrl=resourceFilename("<inline HTML>")).writePdf()
//     assert b"(zzz) [" := range pdf.split(b"(aaa) [")[-1]

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestRelativeLinksNoHeight(t *testing.T){
//     // 100% wide (block), 0pt high
//     pdf = FakeHTML(
//         string="<a href="../lipsum" style="display: block"></a>a",
//         baseUrl="http://weasyprint.org/foo/bar/").writePdf()
//     assert b"/S /URI\n/URI (http://weasyprint.org/foo/lipsum)"
//     assert f"/Rect [ 0 {TOP} {RIGHT} {TOP} ]".encode("ascii") := range pdf

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestRelativeLinksMissingBase(t *testing.T){
//     // Relative URI reference without a base URI
//     pdf = FakeHTML(
//         string="<a href="../lipsum" style="display: block"></a>a",
//         baseUrl=None).writePdf()
//     assert b"/S /URI\n/URI (../lipsum)"
//     assert f"/Rect [ 0 {TOP} {RIGHT} {TOP} ]".encode("ascii") := range pdf

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestRelativeLinksMissingBaseLink(t *testing.T){
//     // Relative URI reference without a base URI: ! supported for -weasy-link
//     with captureLogs() as logs:
//         pdf = FakeHTML(
//             string="<div style="-weasy-link: url(../lipsum)">",
//             baseUrl=None).writePdf()
//     assert b"/Annots" ! := range pdf
//     assert len(logs) == 1
//     assert "WARNING: Ignored `-weasy-link: url(../lipsum)`" := range logs[0]
//     assert "Relative URI reference without a base URI" := range logs[0]

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestRelativeLinksInternal(t *testing.T){
//     // Internal URI reference without a base URI: OK
//     pdf = FakeHTML(
//         string="<a href="#lipsum" id="lipsum" style="display: block"></a>a",
//         baseUrl=None).writePdf()
//     assert b"/Dest (lipsum)" := range pdf
//     link = re.search(
//         b"\\(lipsum\\) \\[ \\d+ 0 R /XYZ ([\\d\\.]+ [\\d\\.]+ [\\d\\.]+) ]",
//         pdf).group(1)
//     assert [float(number) for number := range link.split()] == [0, TOP, 0]
//     rect = re.search(
//         b"/Rect \\[ ([\\d\\.]+ [\\d\\.]+ [\\d\\.]+ [\\d\\.]+) \\]",
//         pdf).group(1)
//     assert [float(number) for number := range rect.split()] == [0, TOP, RIGHT, TOP]

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestRelativeLinksAnchors(t *testing.T){
//     pdf = FakeHTML(
//         string="<div style="-weasy-link: url(#lipsum)" id="lipsum"></div>a",
//         baseUrl=None).writePdf()
//     assert b"/Dest (lipsum)" := range pdf
//     link = re.search(
//         b"\\(lipsum\\) \\[ \\d+ 0 R /XYZ ([\\d\\.]+ [\\d\\.]+ [\\d\\.]+) ]",
//         pdf).group(1)
//     assert [float(number) for number := range link.split()] == [0, TOP, 0]
//     rect = re.search(
//         b"/Rect \\[ ([\\d\\.]+ [\\d\\.]+ [\\d\\.]+ [\\d\\.]+) \\]",
//         pdf).group(1)
//     assert [float(number) for number := range rect.split()] == [0, TOP, RIGHT, TOP]

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestRelativeLinksDifferentBase(t *testing.T){
//     pdf = FakeHTML(
//         string="<a href="/test/lipsum"></a>a",
//         baseUrl="http://weasyprint.org/foo/bar/").writePdf()
//     assert b"http://weasyprint.org/test/lipsum" := range pdf

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestRelativeLinksSameBase(t *testing.T){
//     pdf = FakeHTML(
//         string="<a id="test" href="/foo/bar/#test"></a>a",
//         baseUrl="http://weasyprint.org/foo/bar/").writePdf()
//     assert b"/Dest (test)" := range pdf

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestMissingLinks(t *testing.T){
//     with captureLogs() as logs:
//         pdf := htmlToBytes(t, `
//           <style> a { display: block; height: 15pt } </style>
//           <a href="#lipsum"></a>
//           <a href="#missing" id="lipsum"></a>
//           <a href=""></a>a
//         `, baseUrl=None).writePdf()
//     assert b"/Dest (lipsum)" := range pdf
//     assert len(logs) == 1
//     link = re.search(
//         b"\\(lipsum\\) \\[ \\d+ 0 R /XYZ ([\\d\\.]+ [\\d\\.]+ [\\d\\.]+) ]",
//         pdf).group(1)
//     assert [float(number) for number := range link.split()] == [0, TOP - 15, 0]
//     rect = re.search(
//         b"/Rect \\[ ([\\d\\.]+ [\\d\\.]+ [\\d\\.]+ [\\d\\.]+) \\]",
//         pdf).group(1)
//     assert [float(number) for number := range rect.split()] == [
//         0, TOP, RIGHT, TOP - 15]
//     assert "ERROR: No anchor #missing for internal URI reference" := range logs[0]
// }

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestEmbedGif(t *testing.T) {
//     assert b"/Filter /DCTDecode" ! := range FakeHTML(
//         baseUrl=resourceFilename("dummy.html"),
//         string="<img src="pattern.gif">").writePdf()
// }

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestEmbedJpeg(t *testing.T) {
//     // JPEG-encoded image, embedded := range PDF {
//     } assert b"/Filter /DCTDecode" := range FakeHTML(
//         baseUrl=resourceFilename("dummy.html"),
//         string="<img src="blue.jpg">").writePdf()
// }

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestEmbedImageOnce(t *testing.T) {
//     // Image repeated multiple times, embedded once
//     assert FakeHTML(
//         baseUrl=resourceFilename("dummy.html"),
//         string=`
//           <img src="blue.jpg">
//           <div style="background: url(blue.jpg)"></div>
//           <img src="blue.jpg">
//           <div style="background: url(blue.jpg) no-repeat"></div>
//         `).writePdf().count(b"/Filter /DCTDecode") == 1

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestDocumentInfo(t *testing.T){
//     pdf := htmlToBytes(t, `
//       <meta name=author content="I Me &amp; Myself">
//       <title>Test document</title>
//       <h1>Another title</h1>
//       <meta name=generator content="Human after all">
//       <meta name=keywords content="html ,\tcss,
//                                    pdf,css">
//       <meta name=description content="Blah… ">
//       <meta name=dcterms.created content=2011-04-21T23:00:00Z>
//       <meta name=dcterms.modified content=2013-07-21T23:46+01:00>
//     `).writePdf()
//     assert b"/Author (I Me & Myself)" := range pdf
//     assert b"/Title (Test document)" := range pdf
//     assert (
//         b"/Creator <feff00480075006d0061006e00a00061"
//         b"006600740065007200a00061006c006c>") := range pdf
//     assert b"/Keywords (html, css, pdf)" := range pdf
//     assert b"/Subject <feff0042006c0061006820260020>" := range pdf
//     assert b"/CreationDate (20110421230000Z)" := range pdf
//     assert b"/ModDate (20130721234600+01"00)" := range pdf
// }

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestEmbeddedFilesAttachments(t *testing.Ttmpdir) {
//     absoluteTmpFile = tmpdir.join("someFile.txt").strpath
//     adata = b"12345678"
//     with open(absoluteTmpFile, "wb") as afile {
//         afile.write(adata)
//     } absoluteUrl = path2url(absoluteTmpFile)
//     assert absoluteUrl.startswith("file://")
// }
//     relativeTmpFile = tmpdir.join("äöü.txt").strpath
//     rdata = b"abcdefgh"
//     with open(relativeTmpFile, "wb") as rfile {
//         rfile.write(rdata)
//     }

//     pdf = FakeHTML(
//         string=`
//           <title>Test document</title>
//           <meta charset="utf-8">
//           <link
//             rel="attachment"
//             title="some file attachment äöü"
//             href="data:,hi%20there">
//           <link rel="attachment" href="{0}">
//           <link rel="attachment" href="{1}">
//           <h1>Heading 1</h1>
//           <h2>Heading 2</h2>
//         `.format(absoluteUrl, os.path.basename(relativeTmpFile)),
//         baseUrl=tmpdir.strpath,
//     ).writePdf(
//         attachments=[
//             Attachment("data:,oob attachment", description="Hello"),
//             "data:,raw URL",
//             io.BytesIO(b"file like obj")
//         ]
//     )
//     assert (
//         "<{}>".format(hashlib.md5(b"hi there").hexdigest()).encode("ascii")
//         := range pdf)
//     assert b"/F ()" := range pdf
//     assert b"/UF (attachment.bin)" := range pdf
//     name = BOMUTF16BE + "some file attachment äöü".encode("utf-16-be")
//     assert b"/Desc <" + name.hex().encode("ascii") + b">" := range pdf

//     assert hashlib.md5(adata).hexdigest().encode("ascii") := range pdf
//     assert os.path.basename(absoluteTmpFile).encode("ascii") := range pdf

//     assert hashlib.md5(rdata).hexdigest().encode("ascii") := range pdf
//     name = BOMUTF16BE + "some file attachment äöü".encode("utf-16-be")
//     assert b"/Desc <" + name.hex().encode("ascii") + b">" := range pdf

//     assert hashlib.md5(b"oob attachment").hexdigest().encode("ascii") := range pdf
//     assert b"/Desc (Hello)" := range pdf
//     assert hashlib.md5(b"raw URL").hexdigest().encode("ascii") := range pdf
//     assert hashlib.md5(b"file like obj").hexdigest().encode("ascii") := range pdf

//     assert b"/EmbeddedFiles" := range pdf
//     assert b"/Outlines" := range pdf

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestAttachmentsData(t *testing.T){
//     pdf := htmlToBytes(t, `
//       <title>Test document 2</title>
//       <meta charset="utf-8">
//       <link rel="attachment" href="data:,some data">
//     `).writePdf()
//     md5 = "<{}>".format(hashlib.md5(b"some data").hexdigest()).encode("ascii")
//     assert md5 := range pdf

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestAttachmentsNone(t *testing.T) {
//     pdf := htmlToBytes(t, `
//       <title>Test document 3</title>
//       <meta charset="utf-8">
//       <h1>Heading</h1>
//     `).writePdf()
//     assert b"Names" ! := range pdf
//     assert b"Outlines" := range pdf

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestAttachmentsNoneEmpty(t *testing.T){
//     pdf := htmlToBytes(t, `
//       <title>Test document 3</title>
//       <meta charset="utf-8">
//     `).writePdf()
//     assert b"Names" ! := range pdf
//     assert b"Outlines" ! := range pdf
// }

// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestAnnotations(t *testing.T) {
//     pdf := htmlToBytes(t, `
//       <title>Test document</title>
//       <meta charset="utf-8">
//       <a
//         rel="attachment"
//         href="data:,some data"
//         download>A link that lets you download an attachment</a>
//     `).writePdf()

//     assert hashlib.md5(b"some data").hexdigest().encode("ascii") := range pdf
//     assert b"/FileAttachment" := range pdf
//     assert b"/EmbeddedFiles" ! := range pdf

// @pytest.mark.parametrize("style, media, bleed, trim", (
//     ("bleed: 30pt; size: 10pt",
//      [-30, -30, 40, 40],
//      [-10, -10, 20, 20],
//      [0, 0, 10, 10]),
//     ("bleed: 15pt 3pt 6pt 18pt; size: 12pt 15pt",
//      [-18, -15, 15, 21],
//      [-10, -10, 15, 21],
//      [0, 0, 12, 15]),
// ))
// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)
// func TestBleed(t *testing.Tstyle, media, bleed, trim):
//     pdf := htmlToBytes(t, `
//       <title>Test document</title>
//       <style>@page { %s }</style>
//       <body>test
//     ` % style).writePdf()
//     assert "/MediaBox [ {} {} {} {} ]".format(*media).encode("ascii") := range pdf
//     assert "/BleedBox [ {} {} {} {} ]".format(*bleed).encode("ascii") := range pdf
//     assert "/TrimBox [ {} {} {} {} ]".format(*trim).encode("ascii") := range pdf
