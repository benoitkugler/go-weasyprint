package document

import (
	"fmt"
	"testing"

	bo "github.com/benoitkugler/go-weasyprint/boxes"
	"github.com/benoitkugler/go-weasyprint/utils/testutils"
)

func TestTestCounterSymbols(t *testing.T) {
	for _, arg := range []struct {
		argument string
		values   [4]string
	}{
		{argument: `symbols(cyclic "a" "b" "c")`, values: [4]string{"a ", "b ", "c ", "a "}},
		{argument: `symbols(symbolic "a" "b")`, values: [4]string{"a ", "b ", "aa ", "bb "}},
		{argument: `symbols("a" "b")`, values: [4]string{"a ", "b ", "aa ", "bb "}},
		{argument: `symbols(alphabetic "a" "b")`, values: [4]string{"a ", "b ", "aa ", "ab "}},
		{argument: `symbols(fixed "a" "b")`, values: [4]string{"a ", "b ", "3 ", "4 "}},
		{argument: `symbols(numeric "0" "1" "2")`, values: [4]string{"1 ", "2 ", "10 ", "11 "}},
		{argument: `decimal`, values: [4]string{"1. ", "2. ", "3. ", "4. "}},
		{argument: `"/"`, values: [4]string{"/", "/", "/", "/"}},
	} {
		testCounterSymbols(t, arg.argument, arg.values)
	}
}

func testCounterSymbols(t *testing.T, arguments string, values [4]string) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	pages := renderPages(t, fmt.Sprintf(`
      <style>
        ol { list-style-type: %s }
      </style>
      <ol>
        <li>abc</li>
        <li>abc</li>
        <li>abc</li>
        <li>abc</li>
      </ol>
    `, arguments))

	if len(pages) != 1 {
		t.Fatalf("expected 1 page, got %v", pages)
	}
	page := pages[0]

	html := page.Box().Children[0]
	body := html.Box().Children[0]
	ol := body.Box().Children[0].Box()
	li1, li2, li3, li4 := ol.Children[0], ol.Children[1], ol.Children[2], ol.Children[3]
	if tb := li1.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox); tb.Text != values[0] {
		t.Fatalf("for symbols %s and item %d, expected %s got %s", arguments, 0, values[0], tb.Text)
	}
	if tb := li2.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox); tb.Text != values[1] {
		t.Fatalf("for symbols %s and item %d, expected %s got %s", arguments, 1, values[1], tb.Text)
	}
	if tb := li3.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox); tb.Text != values[2] {
		t.Fatalf("for symbols %s and item %d, expected %s got %s", arguments, 2, values[2], tb.Text)
	}
	if tb := li4.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox); tb.Text != values[3] {
		t.Fatalf("for symbols %s and item %d, expected %s got %s", arguments, 3, values[3], tb.Text)
	}
}

func TestCounterSet(t *testing.T) {
	pages := renderPages(t, `
      <style>
        body { counter-reset: h2 0 h3 4; font-size: 1px }
        article { counter-reset: h2 2 }
        h1 { counter-increment: h1 }
        h1::before { content: counter(h1) }
        h2 { counter-increment: h2; counter-set: h3 3 }
        h2::before { content: counter(h2) }
        h3 { counter-increment: h3 }
        h3::before { content: counter(h3) }
      </style>
      <article>
        <h1></h1>
      </article>
      <article>
        <h2></h2>
        <h3></h3>
      </article>
      <article>
        <h3></h3>
      </article>
      <article>
        <h2></h2>
      </article>
      <article>
        <h3></h3>
        <h3></h3>
      </article>
      <article>
        <h1></h1>
        <h2></h2>
        <h3></h3>
      </article>
    `)
	if len(pages) != 1 {
		t.Fatalf("expected 1 page, got %v", pages)
	}
	page := pages[0]

	html := page.Box().Children[0]
	body := html.Box().Children[0]
	chs := body.Box().Children
	art1, art2, art3, art4, art5, art6 := chs[0], chs[1], chs[2], chs[3], chs[4], chs[5]

	h1 := art1.Box().Children[0]
	if text := h1.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text; text != "1" {
		t.Fatalf("expected %s, got %s", "1", text)
	}

	h2, h3 := art2.Box().Children[0], art2.Box().Children[1]
	if text := h2.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text; text != "3" {
		t.Fatalf("expected %s, got %s", "3", text)
	}
	if text := h3.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text; text != "4" {
		t.Fatalf("expected %s, got %s", "4", text)
	}

	h3 = art3.Box().Children[0]
	if text := h3.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text; text != "5" {
		t.Fatalf("expected %s, got %s", "5", text)
	}

	h2 = art4.Box().Children[0]
	if text := h2.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text; text != "3" {
		t.Fatalf("expected %s, got %s", "3", text)
	}

	h31, h32 := art5.Box().Children[0], art5.Box().Children[1]
	if text := h31.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text; text != "4" {
		t.Fatalf("expected %s, got %s", "4", text)
	}
	if text := h32.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text; text != "5" {
		t.Fatalf("expected %s, got %s", "5", text)
	}

	h1, h2, h3 = art6.Box().Children[0], art6.Box().Children[1], art6.Box().Children[2]
	if text := h1.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text; text != "1" {
		t.Fatalf("expected %s, got %s", "1", text)
	}
	if text := h2.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text; text != "3" {
		t.Fatalf("expected %s, got %s", "3", text)
	}
	if text := h3.Box().Children[0].Box().Children[0].Box().Children[0].(*bo.TextBox).Text; text != "4" {
		t.Fatalf("expected %s, got %s", "4", text)
	}
}

// func TestCounterMultipleExtends(t *testing.T) {
//     // Inspired by W3C failing test system-extends-invalid
//     page, = renderPages(`
//       <style>
//         @counter-style a {
//           system: extends b;
//           prefix: a;
//         }
//         @counter-style b {
//           system: extends c;
//           suffix: b;
//         }
//         @counter-style c {
//           system: extends b;
//           pad: 2 c;
//         }
//         @counter-style d {
//           system: extends d;
//           prefix: d;
//         }
//         @counter-style e {
//           system: extends unknown;
//           prefix: e;
//         }
//         @counter-style f {
//           system: extends decimal;
//           symbols: a;
//         }
//         @counter-style g {
//           system: extends decimal;
//           additive-symbols: 1 a;
//         }
//       </style>
//       <ol>
//         <li style="list-style-type: a"></li>
//         <li style="list-style-type: b"></li>
//         <li style="list-style-type: c"></li>
//         <li style="list-style-type: d"></li>
//         <li style="list-style-type: e"></li>
//         <li style="list-style-type: f"></li>
//         <li style="list-style-type: g"></li>
//         <li style="list-style-type: h"></li>
//       </ol>
//     `)
//     html, = page.children
//     body, = html.children
//     ol, = body.children
//     li1, li2, li3, li4, li5, li6, li7, li8 = ol.children
//     assert li1.children[0].children[0].children[0].text == "a1b"
//     assert li2.children[0].children[0].children[0].text == "2b"
//     assert li3.children[0].children[0].children[0].text == "c3. "
//     assert li4.children[0].children[0].children[0].text == "d4. "
//     assert li5.children[0].children[0].children[0].text == "e5. "
//     assert li6.children[0].children[0].children[0].text == "6. "
//     assert li7.children[0].children[0].children[0].text == "7. "
//     assert li8.children[0].children[0].children[0].text == "8. "
