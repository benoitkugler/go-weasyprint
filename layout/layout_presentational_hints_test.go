package layout

import (
	"testing"

	bo "github.com/benoitkugler/go-weasyprint/boxes"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/style/tree"
	"github.com/benoitkugler/go-weasyprint/utils"
	"github.com/benoitkugler/go-weasyprint/utils/testutils"
)

// Test the HTML presentational hints.

var PHTESTINGCSS, _ = tree.NewCSSDefault(utils.InputString(`
@page {margin: 0; size: 1000px 1000px}
body {margin: 0}
`))

func renderWithPH(t *testing.T, input string, withPH bool, baseUrl string) *bo.PageBox {
	doc, err := tree.NewHTML(utils.InputString(input), baseUrl, nil, "")
	if err != nil {
		t.Fatalf("building tree: %s", err)
	}

	return Layout(doc, []tree.CSS{PHTESTINGCSS}, withPH, fontconfig)[0]
}

func TestNoPh(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// Test both CSS and non-CSS rules
	page := renderWithPH(t, `
	<hr size=100 />
	<table align=right width=100><td>0</td></table>
	`, false, "")
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	hr, table := unpack2(body)
	if hr.Box().BorderHeight() == pr.Float(100) {
		t.Fatal("ht")
	}
	assertEqual(t, table.Box().PositionX, pr.Float(0), "table")
}

func TestPhPage(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)
	page := renderWithPH(t, `
      <body marginheight=2 topmargin=3 leftmargin=5
            bgcolor=red text=blue />
    `, true, "")
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	assertEqual(t, body.Box().MarginTop, pr.Float(2), "body")
	assertEqual(t, body.Box().MarginBottom, pr.Float(2), "body")
	assertEqual(t, body.Box().MarginLeft, pr.Float(5), "body")
	assertEqual(t, body.Box().MarginRight, pr.Float(0), "body")
	assertEqual(t, body.Box().Style.GetBackgroundColor(), pr.NewColor(1, 0, 0, 1), "body")
	assertEqual(t, body.Box().Style.GetColor(), pr.NewColor(0, 0, 1, 1), "body")
}

func TestPhFlow(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderWithPH(t, `
      <pre wrap></pre>
      <center></center>
      <div align=center></div>
      <div align=middle></div>
      <div align=left></div>
      <div align=right></div>
      <div align=justify></div>
    `, true, "")
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	pre, center, div1, div2, div3, div4, div5 := unpack7(body)
	assertEqual(t, pre.Box().Style.GetWhiteSpace(), pr.String("pre-wrap"), "pre")
	assertEqual(t, center.Box().Style.GetTextAlignAll(), pr.String("center"), "center")
	assertEqual(t, div1.Box().Style.GetTextAlignAll(), pr.String("center"), "div1")
	assertEqual(t, div2.Box().Style.GetTextAlignAll(), pr.String("center"), "div2")
	assertEqual(t, div3.Box().Style.GetTextAlignAll(), pr.String("left"), "div3")
	assertEqual(t, div4.Box().Style.GetTextAlignAll(), pr.String("right"), "div4")
	assertEqual(t, div5.Box().Style.GetTextAlignAll(), pr.String("justify"), "div5")
}

func TestPhPhrasing(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderWithPH(t, `
      <style>@font-face {
        src: url(weasyprint.otf); font-family: weasyprint
      }</style>
      <br clear=left>
      <br clear=right />
      <br clear=both />
      <br clear=all />
      <font color=red face=weasyprint size=7></font>
      <Font size=4></Font>
      <font size=+5 ></font>
      <font size=-5 ></font>
    `, true, baseUrl)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line1, line2, line3, line4, line5 := unpack5(body)
	br1 := line1.Box().Children[0]
	br2 := line2.Box().Children[0]
	br3 := line3.Box().Children[0]
	br4 := line4.Box().Children[0]
	font1, font2, font3, font4 := unpack4(line5)
	assertEqual(t, br1.Box().Style.GetClear(), pr.String("left"), "br1")
	assertEqual(t, br2.Box().Style.GetClear(), pr.String("right"), "br2")
	assertEqual(t, br3.Box().Style.GetClear(), pr.String("both"), "br3")
	assertEqual(t, br4.Box().Style.GetClear(), pr.String("both"), "br4")
	assertEqual(t, font1.Box().Style.GetColor(), pr.NewColor(1, 0, 0, 1), "font1")
	assertEqual(t, font1.Box().Style.GetFontFamily(), pr.Strings{"weasyprint"}, "font1")
	assertEqual(t, font1.Box().Style.GetFontSize(), pr.FToV(1.5*2*16), "font1")
	assertEqual(t, font2.Box().Style.GetFontSize(), pr.FToV(6./5*16), "font2")
	assertEqual(t, font3.Box().Style.GetFontSize(), pr.FToV(1.5*2*16), "font3")
	assertEqual(t, font4.Box().Style.GetFontSize(), pr.FToV(8./9*16), "font4")
}

func TestPhLists(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderWithPH(t, `
      <ol>
        <li type=A></li>
        <li type=1></li>
        <li type=a></li>
        <li type=i></li>
        <li type=I></li>
      </ol>
      <ul>
        <li type=circle></li>
        <li type=disc></li>
        <li type=square></li>
      </ul>
    `, true, "")
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	ol, ul := unpack2(body)
	oli1, oli2, oli3, oli4, oli5 := unpack5(ol)
	uli1, uli2, uli3 := unpack3(ul)
	assertEqual(t, oli1.Box().Style.GetListStyleType(), pr.CounterStyleID{Name: "upper-alpha"}, "oli1")
	assertEqual(t, oli2.Box().Style.GetListStyleType(), pr.CounterStyleID{Name: "decimal"}, "oli2")
	assertEqual(t, oli3.Box().Style.GetListStyleType(), pr.CounterStyleID{Name: "lower-alpha"}, "oli3")
	assertEqual(t, oli4.Box().Style.GetListStyleType(), pr.CounterStyleID{Name: "lower-roman"}, "oli4")
	assertEqual(t, oli5.Box().Style.GetListStyleType(), pr.CounterStyleID{Name: "upper-roman"}, "oli5")
	assertEqual(t, uli1.Box().Style.GetListStyleType(), pr.CounterStyleID{Name: "circle"}, "uli1")
	assertEqual(t, uli2.Box().Style.GetListStyleType(), pr.CounterStyleID{Name: "disc"}, "uli2")
	assertEqual(t, uli3.Box().Style.GetListStyleType(), pr.CounterStyleID{Name: "square"}, "uli3")
}

func TestPhListsTypes(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)
	page := renderWithPH(t, `
      <ol type=A></ol>
      <ol type=1></ol>
      <ol type=a></ol>
      <ol type=i></ol>
      <ol type=I></ol>
      <ul type=circle></ul>
      <ul type=disc></ul>
      <ul type=square></ul>
    `, true, "")
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	ol1, ol2, ol3, ol4, ol5, ul1, ul2, ul3 := unpack8(body)
	assertEqual(t, ol1.Box().Style.GetListStyleType(), pr.CounterStyleID{Name: "upper-alpha"}, "ol1")
	assertEqual(t, ol2.Box().Style.GetListStyleType(), pr.CounterStyleID{Name: "decimal"}, "ol2")
	assertEqual(t, ol3.Box().Style.GetListStyleType(), pr.CounterStyleID{Name: "lower-alpha"}, "ol3")
	assertEqual(t, ol4.Box().Style.GetListStyleType(), pr.CounterStyleID{Name: "lower-roman"}, "ol4")
	assertEqual(t, ol5.Box().Style.GetListStyleType(), pr.CounterStyleID{Name: "upper-roman"}, "ol5")
	assertEqual(t, ul1.Box().Style.GetListStyleType(), pr.CounterStyleID{Name: "circle"}, "ul1")
	assertEqual(t, ul2.Box().Style.GetListStyleType(), pr.CounterStyleID{Name: "disc"}, "ul2")
	assertEqual(t, ul3.Box().Style.GetListStyleType(), pr.CounterStyleID{Name: "square"}, "ul3")
}

func TestPhTables(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderWithPH(t, `
      <table align=left rules=none></table>
      <table align=right rules=groups></table>
      <table align=center rules=rows></table>
      <table border=10 cellspacing=3 bordercolor=green>
        <thead>
          <tr>
            <th valign=top></th>
          </tr>
        </thead>
        <tr>
          <td nowrap><h1 align=right></h1><p align=center></p></td>
        </tr>
        <tr>
        </tr>
        <tfoot align=justify>
          <tr>
            <td></td>
          </tr>
        </tfoot>
      </table>
    `, true, "")
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	wrapper1, wrapper2, wrapper3, wrapper4 := unpack4(body)
	assertEqual(t, wrapper1.Box().Style.GetFloat(), pr.String("left"), "wrapper1")
	assertEqual(t, wrapper2.Box().Style.GetFloat(), pr.String("right"), "wrapper2")
	assertEqual(t, wrapper3.Box().Style.GetMarginLeft(), pr.SToV("auto"), "wrapper3")
	assertEqual(t, wrapper3.Box().Style.GetMarginRight(), pr.SToV("auto"), "wrapper3")
	assertEqual(t, wrapper1.Box().Children[0].Box().Style.GetBorderLeftStyle(), pr.String("hidden"), "wrapper1")
	assertEqual(t, wrapper1.Box().Style.GetBorderCollapse(), pr.String("collapse"), "wrapper1")
	assertEqual(t, wrapper2.Box().Children[0].Box().Style.GetBorderLeftStyle(), pr.String("hidden"), "wrapper2")
	assertEqual(t, wrapper2.Box().Style.GetBorderCollapse(), pr.String("collapse"), "wrapper2")
	assertEqual(t, wrapper3.Box().Children[0].Box().Style.GetBorderLeftStyle(), pr.String("hidden"), "wrapper3")
	assertEqual(t, wrapper3.Box().Style.GetBorderCollapse(), pr.String("collapse"), "wrapper3")

	table4 := wrapper4.Box().Children[0]
	assertEqual(t, table4.Box().Style.GetBorderTopStyle(), pr.String("outset"), "table4")
	assertEqual(t, table4.Box().Style.GetBorderTopWidth(), pr.FToV(10), "table4")
	assertEqual(t, table4.Box().Style.GetBorderSpacing(), pr.Point{pr.FToPx(3).Dimension, pr.FToPx(3).Dimension}, "table4")
	r, g, b, _ := table4.Box().Style.GetBorderLeftColor().RGBA.Unpack()
	if !(g > r && g > b) {
		t.Fatal("color")
	}
	headGroup, rowsGroup, footGroup := unpack3(table4)
	head := headGroup.Box().Children[0]
	th := head.Box().Children[0]
	assertEqual(t, th.Box().Style.GetVerticalAlign(), pr.SToV("top"), "th")
	line1, _ := unpack2(rowsGroup)
	td := line1.Box().Children[0]
	assertEqual(t, td.Box().Style.GetWhiteSpace(), pr.String("nowrap"), "td")
	assertEqual(t, td.Box().Style.GetBorderTopWidth(), pr.FToV(1), "td")
	assertEqual(t, td.Box().Style.GetBorderTopStyle(), pr.String("inset"), "td")
	h1, p := unpack2(td)
	assertEqual(t, h1.Box().Style.GetTextAlignAll(), pr.String("right"), "h1")
	assertEqual(t, p.Box().Style.GetTextAlignAll(), pr.String("center"), "p")
	foot := footGroup.Box().Children[0]
	tr := foot.Box().Children[0]
	assertEqual(t, tr.Box().Style.GetTextAlignAll(), pr.String("justify"), "tr")
}

func TestPhHr(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderWithPH(t, `
      <hr align=left>
      <hr align=right />
      <hr align=both color=red />
      <hr align=center noshade size=10 />
      <hr align=all size=8 width=100 />
    `, true, "")
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	hr1, hr2, hr3, hr4, hr5 := unpack5(body)
	assertEqual(t, hr1.Box().MarginLeft, pr.Float(0), "hr1")
	assertEqual(t, hr1.Box().Style.GetMarginRight(), pr.SToV("auto"), "hr1")
	assertEqual(t, hr2.Box().Style.GetMarginLeft(), pr.SToV("auto"), "hr2")
	assertEqual(t, hr2.Box().MarginRight, pr.Float(0), "hr2")
	assertEqual(t, hr3.Box().Style.GetMarginLeft(), pr.SToV("auto"), "hr3")
	assertEqual(t, hr3.Box().Style.GetMarginRight(), pr.SToV("auto"), "hr3")
	assertEqual(t, hr3.Box().Style.GetColor(), pr.NewColor(1, 0, 0, 1), "hr3")
	assertEqual(t, hr4.Box().Style.GetMarginLeft(), pr.SToV("auto"), "hr4")
	assertEqual(t, hr4.Box().Style.GetMarginRight(), pr.SToV("auto"), "hr4")
	assertEqual(t, hr4.Box().BorderHeight(), pr.Float(10), "hr4")
	assertEqual(t, hr4.Box().Style.GetBorderTopWidth(), pr.FToV(5), "hr4")
	assertEqual(t, hr5.Box().BorderHeight(), pr.Float(8), "hr5")
	assertEqual(t, hr5.Box().Height, pr.Float(6), "hr5")
	assertEqual(t, hr5.Box().Width, pr.Float(100), "hr5")
	assertEqual(t, hr5.Box().Style.GetBorderTopWidth(), pr.FToV(1), "hr5")
}

func TestPhEmbedded(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	page := renderWithPH(t, `
      <object data="data:image/svg+xml,<svg></svg>"
              align=top hspace=10 vspace=20></object>
      <img src="data:image/svg+xml,<svg></svg>" alt=text
              align=right width=10 height=20 />
      <embed src="data:image/svg+xml,<svg></svg>" align=texttop />
    `, true, "")
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	object, _, img, embed, _ := unpack5(line)
	assertEqual(t, embed.Box().Style.GetVerticalAlign(), pr.SToV("text-top"), "embed")
	assertEqual(t, object.Box().Style.GetVerticalAlign(), pr.SToV("top"), "object")
	assertEqual(t, object.Box().MarginTop, pr.Float(20), "object")
	assertEqual(t, object.Box().MarginLeft, pr.Float(10), "object")
	assertEqual(t, img.Box().Style.GetFloat(), pr.String("right"), "img")
	assertEqual(t, img.Box().Width, pr.Float(10), "img")
	assertEqual(t, img.Box().Height, pr.Float(20), "img")
}
