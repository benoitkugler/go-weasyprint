package layout

import (
	"fmt"
	"testing"

	bo "github.com/benoitkugler/go-weasyprint/boxes"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	tu "github.com/benoitkugler/go-weasyprint/utils/testutils"
)

// Tests for multicolumn layout.

func columnsMetrics(columns []Box) (widths, heights, xs, ys []pr.Float) {
	for _, column := range columns {
		widths = append(widths, column.Box().Width.V())
		heights = append(heights, column.Box().Height.V())
		xs = append(xs, column.Box().PositionX.V())
		ys = append(ys, column.Box().PositionY.V())
	}
	return
}

func TestColumns(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, css := range []string{
		"columns: 4",
		"columns: 100px",
		"columns: 4 100px",
		"columns: 100px 4",
		"column-width: 100px",
		"column-count: 4",
	} {
		page := renderOnePage(t, fmt.Sprintf(`
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { %s; column-gap: 0 }
        body { margin: 0; font-family: weasyprint }
        @page { margin: 0; size: 400px 1000px }
      </style>
      <div>
        Ipsum dolor sit amet,
        consectetur adipiscing elit.
        Sed sollicitudin nibh
        et turpis molestie tristique.
      </div>
    `, css))
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		div := body.Box().Children[0]
		columns := div.Box().Children
		tu.AssertEqual(t, len(columns), 4, "len")
		widths, _, xs, ys := columnsMetrics(columns)
		tu.AssertEqual(t, widths, []pr.Float{100, 100, 100, 100}, "widths")
		tu.AssertEqual(t, xs, []pr.Float{0, 100, 200, 300}, "xs")
		tu.AssertEqual(t, ys, []pr.Float{0, 0, 0, 0}, "ys")
	}
}

func TestColumnGap(t *testing.T) {
	for _, data := range []struct {
		value string
		width pr.Float
	}{
		{"normal", 16},  // "normal" is 1em = 16px
		{"unknown", 16}, // default value is normal
		{"15px", 15},
		{"40%", 16},  // percentages are not allowed
		{"-1em", 16}, // negative values are not allowed
	} {
		page := renderOnePage(t, fmt.Sprintf(`
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 3; column-gap: %s }
        body { margin: 0; font-family: weasyprint }
        @page { margin: 0; size: 300px 1000px }
      </style>
      <div>
        Ipsum dolor sit amet,
        consectetur adipiscing elit.
        Sed sollicitudin nibh
        et turpis molestie tristique.
      </div>
    `, data.value))
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		div := body.Box().Children[0]
		columns := div.Box().Children
		tu.AssertEqual(t, len(columns), 3, "len")
		widths, _, xs, ys := columnsMetrics(columns)

		tu.AssertEqual(t, widths, []pr.Float{100 - 2*data.width/3, 100 - 2*data.width/3, 100 - 2*data.width/3}, "widths")
		tu.AssertEqual(t, xs, []pr.Float{0, 100 + data.width/3, 200 + 2*data.width/3}, "xs")
		tu.AssertEqual(t, ys, []pr.Float{0, 0, 0}, "ys")
	}
}

func TestColumnSpan(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        body { margin: 0; font-family: weasyprint; line-height: 1 }
        div { columns: 2; width: 10em; column-gap: 0 }
        section { column-span: all; margin: 1em 0 }
      </style>
 
      <div>
        abc def
        <section>test</section>
        <section>test</section>
        ghi jkl
      </div>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	column1, column2, section1, section2, column3, column4 := unpack6(div)
	tu.AssertEqual(t, [2]pr.Float{column1.Box().PositionX, column1.Box().PositionY}, [2]pr.Float{0, 0}, "column1")
	tu.AssertEqual(t, [2]pr.Float{column2.Box().PositionX, column2.Box().PositionY}, [2]pr.Float{5 * 16, 0}, "column2")
	tu.AssertEqual(t, [2]pr.Float{section1.Box().ContentBoxX(), section1.Box().ContentBoxY()}, [2]pr.Float{0, 32}, "section1")
	tu.AssertEqual(t, [2]pr.Float{section2.Box().ContentBoxX(), section2.Box().ContentBoxY()}, [2]pr.Float{0, 64}, "section2")
	tu.AssertEqual(t, [2]pr.Float{column3.Box().PositionX, column3.Box().PositionY}, [2]pr.Float{0, 96}, "column3")
	tu.AssertEqual(t, [2]pr.Float{column4.Box().PositionX, column4.Box().PositionY}, [2]pr.Float{5 * 16, 96}, "column4")

	tu.AssertEqual(t, column1.Box().Height, pr.Float(16), "column1")
}

func TestColumnsMultipage(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	pages := renderPages(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 2; column-gap: 1px }
        body { margin: 0; font-family: weasyprint;
               font-size: 1px; line-height: 1px }
        @page { margin: 0; size: 3px 2px }
      </style>
      <div>a b c d e f g</div>
    `)
	page1, page2 := pages[0], pages[1]
	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	columns := div.Box().Children
	tu.AssertEqual(t, len(columns), 2, "len")
	tu.AssertEqual(t, len(columns[0].Box().Children), 2, "len")
	tu.AssertEqual(t, len(columns[1].Box().Children), 2, "len")
	tu.AssertEqual(t, columns[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "a", "")
	tu.AssertEqual(t, columns[0].Box().Children[1].Box().Children[0].(*bo.TextBox).Text, "b", "")
	tu.AssertEqual(t, columns[1].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "c", "")
	tu.AssertEqual(t, columns[1].Box().Children[1].Box().Children[0].(*bo.TextBox).Text, "d", "")

	html = page2.Box().Children[0]
	body = html.Box().Children[0]
	div = body.Box().Children[0]
	columns = div.Box().Children
	tu.AssertEqual(t, len(columns), 2, "len")
	tu.AssertEqual(t, len(columns[0].Box().Children), 2, "len")
	tu.AssertEqual(t, len(columns[1].Box().Children), 1, "len")
	tu.AssertEqual(t, columns[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "e", "")
	tu.AssertEqual(t, columns[0].Box().Children[1].Box().Children[0].(*bo.TextBox).Text, "f", "")
	tu.AssertEqual(t, columns[1].Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "g", "")
}

func TestColumnsNotEnoughContent(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 5; column-gap: 0 }
        body { margin: 0; font-family: weasyprint }
        @page { margin: 0; size: 5px; font-size: 1px }
      </style>
      <div>a b c</div>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	tu.AssertEqual(t, div.Box().Width, pr.Float(5), "div")
	columns := div.Box().Children
	tu.AssertEqual(t, len(columns), 3, "len")
	widths, _, xs, ys := columnsMetrics(columns)
	tu.AssertEqual(t, widths, []pr.Float{1, 1, 1}, "widths")
	tu.AssertEqual(t, xs, []pr.Float{0, 1, 2}, "xs")
	tu.AssertEqual(t, ys, []pr.Float{0, 0, 0}, "ys")
}

func TestColumnsEmpty(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 3 }
        body { margin: 0; font-family: weasyprint }
        @page { margin: 0; size: 3px; font-size: 1px }
      </style>
      <div></div>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	tu.AssertEqual(t, div.Box().Width, pr.Float(3), "div")
	tu.AssertEqual(t, div.Box().Height, pr.Float(0), "div")
	columns := div.Box().Children
	tu.AssertEqual(t, len(columns), 0, "len")
}

func TestColumnsFixedHeight(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, prop := range []string{"height", "min-height"} {
		// TODO: we should test when the height is too small
		page := renderOnePage(t, fmt.Sprintf(`
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 4; column-gap: 0; %s: 10px }
        body { margin: 0; font-family: weasyprint; line-height: 1px }
        @page { margin: 0; size: 4px 50px; font-size: 1px }
      </style>
      <div>a b c</div>
    `, prop))
		html := page.Box().Children[0]
		body := html.Box().Children[0]
		div := body.Box().Children[0]
		tu.AssertEqual(t, div.Box().Width, pr.Float(4), "div")
		columns := div.Box().Children
		tu.AssertEqual(t, len(columns), 3, "len")

		widths, heights, xs, ys := columnsMetrics(columns)

		tu.AssertEqual(t, widths, []pr.Float{1, 1, 1}, "widths")
		tu.AssertEqual(t, heights, []pr.Float{10, 10, 10}, "heights")
		tu.AssertEqual(t, xs, []pr.Float{0, 1, 2}, "xs")
		tu.AssertEqual(t, ys, []pr.Float{0, 0, 0}, "ys")
	}
}

func TestColumnsPadding(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        div { columns: 4; column-gap: 0; padding: 1px }
        body { margin: 0; font-family: weasyprint; line-height: 1px }
        @page { margin: 0; size: 6px 50px; font-size: 1px }
      </style>
      <div>a b c</div>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	tu.AssertEqual(t, div.Box().Width, pr.Float(4), "div.Width")
	tu.AssertEqual(t, div.Box().Height, pr.Float(1), "div.Height")
	tu.AssertEqual(t, div.Box().PaddingWidth(), pr.Float(6), "div.PaddingWidth")
	tu.AssertEqual(t, div.Box().PaddingHeight(), pr.Float(3), "div.PaddingHeight")
	columns := div.Box().Children
	tu.AssertEqual(t, len(columns), 3, "len")
	widths, heights, xs, ys := columnsMetrics(columns)
	tu.AssertEqual(t, widths, []pr.Float{1, 1, 1}, "widths")
	tu.AssertEqual(t, heights, []pr.Float{1, 1, 1}, "heights")
	tu.AssertEqual(t, xs, []pr.Float{1, 2, 3}, "xs")
	tu.AssertEqual(t, ys, []pr.Float{1, 1, 1}, "ys")
}

func TestColumnsRelative(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	page := renderOnePage(t, `
      <style>
        @font-face { src: url(weasyprint.otf); font-family: weasyprint }
        article { position: absolute; top: 3px }
        div { columns: 4; column-gap: 0; position: relative;
              top: 1px; left: 2px }
        body { margin: 0; font-family: weasyprint; line-height: 1px }
        @page { margin: 0; size: 4px 50px; font-size: 1px }
      </style>
      <div>a b c d<article>e</article></div>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	tu.AssertEqual(t, div.Box().Width, pr.Float(4), "div")
	columns := div.Box().Children
	widths, _, xs, ys := columnsMetrics(columns)
	tu.AssertEqual(t, widths, []pr.Float{1, 1, 1, 1}, "widths")
	tu.AssertEqual(t, xs, []pr.Float{2, 3, 4, 5}, "xs")
	tu.AssertEqual(t, ys, []pr.Float{1, 1, 1, 1}, "ys")
	column4 := columns[len(columns)-1]
	columnLine := column4.Box().Children[0]
	absoluteArticle := columnLine.Box().Children[1]
	absoluteLine := absoluteArticle.Box().Children[0]
	span := absoluteLine.Box().Children[0]
	tu.AssertEqual(t, span.Box().PositionX, pr.Float(5), "PositionX") // Default position of the 4th column
	tu.AssertEqual(t, span.Box().PositionY, pr.Float(4), "PositionY") // div"s 1px + span"s 3px
}

func TestColumnsRegression1(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test #1 for https://github.com/Kozea/WeasyPrint/issues/659
	pages := renderPages(t, `
      <style>
        @page {margin: 0; width: 100px; height: 100px}
        body {margin: 0; font-size: 1px}
      </style>
      <div style="height:95px">A</div>
      <div style="column-count:2">
        <div style="height:20px">B1</div>
        <div style="height:20px">B2</div>
        <div style="height:20px">B3</div>
      </div>
      <div style="height:95px">C</div>
    `)
	page1, page2, page3 := pages[0], pages[1], pages[2]

	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	tu.AssertEqual(t, div.Box().PositionY, pr.Float(0), "div")
	tu.AssertEqual(t, div.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "A", "div")

	html = page2.Box().Children[0]
	body = html.Box().Children[0]
	div = body.Box().Children[0]
	tu.AssertEqual(t, div.Box().PositionY, pr.Float(0), "div")
	column1, column2 := unpack2(div)
	tu.AssertEqual(t, column1.Box().PositionY, pr.Float(0), "column1")
	tu.AssertEqual(t, column2.Box().PositionY, pr.Float(0), "column2")
	div1, div2 := unpack2(column1)
	div3 := column2.Box().Children[0]
	tu.AssertEqual(t, div1.Box().PositionY, pr.Float(0), "div1")
	tu.AssertEqual(t, div3.Box().PositionY, pr.Float(0), "div1")
	tu.AssertEqual(t, div2.Box().PositionY, pr.Float(20), "div2")
	tu.AssertEqual(t, div1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B1", "div1")
	tu.AssertEqual(t, div2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B2", "div2")
	tu.AssertEqual(t, div3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B3", "div3")

	html = page3.Box().Children[0]
	body = html.Box().Children[0]
	div = body.Box().Children[0]
	tu.AssertEqual(t, div.Box().PositionY, pr.Float(0), "div")
	tu.AssertEqual(t, div.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "C", "div")
}

func TestColumnsRegression2(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test #2 for https://github.com/Kozea/WeasyPrint/issues/659
	pages := renderPages(t, `
      <style>
        @page {margin: 0; width: 100px; height: 100px}
        body {margin: 0; font-size: 1px}
      </style>
      <div style="column-count:2">
        <div style="height:20px">B1</div>
        <div style="height:60px">B2</div>
        <div style="height:60px">B3</div>
        <div style="height:60px">B4</div>
      </div>
    `)
	page1, page2 := pages[0], pages[1]

	html := page1.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	tu.AssertEqual(t, div.Box().PositionY, pr.Float(0), "div")
	column1, column2 := unpack2(div)
	tu.AssertEqual(t, column1.Box().PositionY, pr.Float(0), "column1")
	tu.AssertEqual(t, column2.Box().PositionY, pr.Float(0), "column2")
	div1, div2 := unpack2(column1)
	div3 := column2.Box().Children[0]
	tu.AssertEqual(t, div1.Box().PositionY, pr.Float(0), "div1")
	tu.AssertEqual(t, div3.Box().PositionY, pr.Float(0), "div1")
	tu.AssertEqual(t, div2.Box().PositionY, pr.Float(20), "div2")
	tu.AssertEqual(t, div1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B1", "div1")
	tu.AssertEqual(t, div2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B2", "div2")
	tu.AssertEqual(t, div3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B3", "div3")

	html = page2.Box().Children[0]
	body = html.Box().Children[0]
	div = body.Box().Children[0]
	tu.AssertEqual(t, div.Box().PositionY, pr.Float(0), "div")
	column1 = div.Box().Children[0]
	tu.AssertEqual(t, column1.Box().PositionY, pr.Float(0), "column1")
	div1 = column1.Box().Children[0]
	tu.AssertEqual(t, div1.Box().PositionY, pr.Float(0), "div1")
	tu.AssertEqual(t, div3.Box().PositionY, pr.Float(0), "div1")
	tu.AssertEqual(t, div1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B4", "div1")
}

func TestColumnsRegression3(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test #3 for https://github.com/Kozea/WeasyPrint/issues/659
	page := renderOnePage(t, `
      <style>
        @page {margin: 0; width: 100px; height: 100px}
        body {margin: 0; font-size: 10px}
      </style>
      <div style="column-count:2">
        <div style="height:20px; margin:5px">B1</div>
        <div style="height:60px">B2</div>
        <div style="height:60px">B3</div>
      </div>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	tu.AssertEqual(t, div.Box().PositionY, pr.Float(0), "div")
	column1, column2 := unpack2(div)
	tu.AssertEqual(t, column1.Box().PositionY, pr.Float(0), "column1")
	tu.AssertEqual(t, column2.Box().PositionY, pr.Float(0), "column1")
	div1, div2 := unpack2(column1)
	div3 := column2.Box().Children[0]
	tu.AssertEqual(t, div1.Box().PositionY, pr.Float(0), "div1")
	tu.AssertEqual(t, div3.Box().PositionY, pr.Float(0), "div1")
	tu.AssertEqual(t, div2.Box().PositionY, pr.Float(30), "div2")
	tu.AssertEqual(t, div.Box().Height, pr.Float(5+20+5+60), "div")
	tu.AssertEqual(t, div1.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B1", "div1")
	tu.AssertEqual(t, div2.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B2", "div2")
	tu.AssertEqual(t, div3.Box().Children[0].Box().Children[0].(*bo.TextBox).Text, "B3", "div3")
}

func TestColumnsRegression4(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/897
	page := renderOnePage(t, `
      <div style="position:absolute">
        <div style="column-count:2">
          <div>a</div>
        </div>
      </div>
    `)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	div := body.Box().Children[0]
	tu.AssertEqual(t, div.Box().PositionY, pr.Float(0), "div")
	column1 := div.Box().Children[0]
	tu.AssertEqual(t, column1.Box().PositionY, pr.Float(0), "column1")
	div1 := column1.Box().Children[0]
	tu.AssertEqual(t, div1.Box().PositionY, pr.Float(0), "div1")
}

func TestColumnsRegression5(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1191
	_ = renderPages(t, `
      <style>
        @page {width: 100px; height: 100px}
      </style>
      <div style="height: 1px"></div>
      <div style="columns: 2">
        <div style="break-after: avoid">
          <div style="height: 50px"></div>
        </div>
        <div style="break-after: avoid">
          <div style="height: 50px"></div>
          <p>a</p>
        </div>
      </div>
      <div style="height: 50px"></div>
    `)
}
