package boxes

import (
	"reflect"
	"strings"
	"testing"

	"github.com/benoitkugler/go-weasyprint/style/tree"
	"github.com/benoitkugler/go-weasyprint/utils/testutils"
)

func TestCounters1(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	exp := func(counter string) serBox {
		return serBox{"p", BlockBoxT, bc{c: []serBox{
			{"p", LineBoxT, bc{c: []serBox{{"p::before", InlineBoxT, bc{c: []serBox{{"p::before", TextBoxT, bc{text: counter}}}}}}}},
		}}}
	}
	var expected []serBox
	for _, counter := range strings.Fields("0 1 3  2 4 6  -11 -9 -7  44 46 48") {
		expected = append(expected, exp(counter))
	}
	assertTree(t, parseAndBuild(t, `
      <style>
        p { counter-increment: p 2 }
        p:before { content: counter(p); }
        p:nth-child(1) { counter-increment: none; }
        p:nth-child(2) { counter-increment: p; }
      </style>
      <p></p>
      <p></p>
      <p></p>
      <p style="counter-reset: p 117 p"></p>
      <p></p>
      <p></p>
      <p style="counter-reset: p -13"></p>
      <p></p>
      <p></p>
      <p style="counter-reset: p 42"></p>
      <p></p>
      <p></p>`), expected)
}

func TestCounters2(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parseAndBuild(t, `
      <ol style="list-style-position: inside">
        <li></li>
        <li></li>
        <li></li>
        <li><ol>
          <li></li>
          <li style="counter-increment: none"></li>
          <li></li>
        </ol></li>
        <li></li>
      </ol>`), []serBox{
		{"ol", BlockBoxT, bc{c: []serBox{
			{"li", BlockBoxT, bc{c: []serBox{
				{"li", LineBoxT, bc{c: []serBox{{"li::marker", InlineBoxT, bc{c: []serBox{{"li::marker", TextBoxT, bc{text: "1. "}}}}}}}},
			}}},
			{"li", BlockBoxT, bc{c: []serBox{
				{"li", LineBoxT, bc{c: []serBox{{"li::marker", InlineBoxT, bc{c: []serBox{{"li::marker", TextBoxT, bc{text: "2. "}}}}}}}},
			}}},
			{"li", BlockBoxT, bc{c: []serBox{
				{"li", LineBoxT, bc{c: []serBox{{"li::marker", InlineBoxT, bc{c: []serBox{{"li::marker", TextBoxT, bc{text: "3. "}}}}}}}},
			}}},
			{"li", BlockBoxT, bc{c: []serBox{
				{"li", BlockBoxT, bc{c: []serBox{
					{"li", LineBoxT, bc{c: []serBox{{"li::marker", InlineBoxT, bc{c: []serBox{{"li::marker", TextBoxT, bc{text: "4. "}}}}}}}},
				}}},
				{"ol", BlockBoxT, bc{c: []serBox{
					{"li", BlockBoxT, bc{c: []serBox{
						{"li", LineBoxT, bc{c: []serBox{{"li::marker", InlineBoxT, bc{c: []serBox{{"li::marker", TextBoxT, bc{text: "1. "}}}}}}}},
					}}},
					{"li", BlockBoxT, bc{c: []serBox{
						{"li", LineBoxT, bc{c: []serBox{{"li::marker", InlineBoxT, bc{c: []serBox{{"li::marker", TextBoxT, bc{text: "1. "}}}}}}}},
					}}},
					{"li", BlockBoxT, bc{c: []serBox{
						{"li", LineBoxT, bc{c: []serBox{{"li::marker", InlineBoxT, bc{c: []serBox{{"li::marker", TextBoxT, bc{text: "2. "}}}}}}}},
					}}},
				}}},
			}}},
			{"li", BlockBoxT, bc{c: []serBox{
				{"li", LineBoxT, bc{c: []serBox{{"li::marker", InlineBoxT, bc{c: []serBox{{"li::marker", TextBoxT, bc{text: "5. "}}}}}}}},
			}}},
		}}},
	})
}

func TestCounters3(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parseAndBuild(t, `
      <style>
        p { display: list-item; list-style: inside decimal }
      </style>
      <div>
        <p></p>
        <p></p>
        <p style="counter-reset: list-item 7 list-item -56"></p>
      </div>
      <p></p>`), []serBox{
		{"div", BlockBoxT, bc{c: []serBox{
			{"p", BlockBoxT, bc{c: []serBox{
				{"p", LineBoxT, bc{c: []serBox{{"p::marker", InlineBoxT, bc{c: []serBox{{"p::marker", TextBoxT, bc{text: "1. "}}}}}}}},
			}}},
			{"p", BlockBoxT, bc{c: []serBox{
				{"p", LineBoxT, bc{c: []serBox{{"p::marker", InlineBoxT, bc{c: []serBox{{"p::marker", TextBoxT, bc{text: "2. "}}}}}}}},
			}}},
			{"p", BlockBoxT, bc{c: []serBox{
				{"p", LineBoxT, bc{c: []serBox{{"p::marker", InlineBoxT, bc{c: []serBox{{"p::marker", TextBoxT, bc{text: "-55. "}}}}}}}},
			}}},
		}}},
		{"p", BlockBoxT, bc{c: []serBox{
			{"p", LineBoxT, bc{c: []serBox{{"p::marker", InlineBoxT, bc{c: []serBox{{"p::marker", TextBoxT, bc{text: "1. "}}}}}}}},
		}}},
	})
}

func TestCounters4(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parseAndBuild(t, `
      <style>
        section:before { counter-reset: h; content: "" }
        h1:before { counter-increment: h; content: counters(h, ".") }
      </style>
      <body>
        <section><h1></h1>
          <h1></h1>
          <section><h1></h1>
            <h1></h1>
          </section>
          <h1></h1>
        </section>
      </body>`), []serBox{
		{"section", BlockBoxT, bc{c: []serBox{
			{"section", BlockBoxT, bc{c: []serBox{{"section", LineBoxT, bc{c: []serBox{{"section::before", InlineBoxT, bc{c: []serBox{}}}}}}}}},
			{"h1", BlockBoxT, bc{c: []serBox{
				{"h1", LineBoxT, bc{c: []serBox{{"h1::before", InlineBoxT, bc{c: []serBox{{"h1::before", TextBoxT, bc{text: "1"}}}}}}}},
			}}},
			{"h1", BlockBoxT, bc{c: []serBox{
				{"h1", LineBoxT, bc{c: []serBox{{"h1::before", InlineBoxT, bc{c: []serBox{{"h1::before", TextBoxT, bc{text: "2"}}}}}}}},
			}}},
			{"section", BlockBoxT, bc{c: []serBox{
				{"section", BlockBoxT, bc{c: []serBox{{"section", LineBoxT, bc{c: []serBox{{"section::before", InlineBoxT, bc{c: []serBox{}}}}}}}}},
				{"h1", BlockBoxT, bc{c: []serBox{
					{"h1", LineBoxT, bc{c: []serBox{{"h1::before", InlineBoxT, bc{c: []serBox{{"h1::before", TextBoxT, bc{text: "2.1"}}}}}}}},
				}}},
				{"h1", BlockBoxT, bc{c: []serBox{
					{"h1", LineBoxT, bc{c: []serBox{{"h1::before", InlineBoxT, bc{c: []serBox{{"h1::before", TextBoxT, bc{text: "2.2"}}}}}}}},
				}}},
			}}},
			{"h1", BlockBoxT, bc{c: []serBox{
				{"h1", LineBoxT, bc{c: []serBox{{"h1::before", InlineBoxT, bc{c: []serBox{{"h1::before", TextBoxT, bc{text: "3"}}}}}}}},
			}}},
		}}},
	})
}

func TestCounters5(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parseAndBuild(t, `
      <style>
        p:before { content: counter(c) }
      </style>
      <div>
        <span style="counter-reset: c">
          Scope created now, deleted after the div
        </span>
      </div>
      <p></p>`), []serBox{
		{"div", BlockBoxT, bc{c: []serBox{
			{"div", LineBoxT, bc{c: []serBox{{"span", InlineBoxT, bc{c: []serBox{{"span", TextBoxT, bc{text: "Scope created now, deleted after the div "}}}}}}}},
		}}},
		{"p", BlockBoxT, bc{c: []serBox{
			{"p", LineBoxT, bc{c: []serBox{{"p::before", InlineBoxT, bc{c: []serBox{{"p::before", TextBoxT, bc{text: "0"}}}}}}}},
		}}},
	})
}

func TestCounters6(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	// counter-increment may interfere with display: list-item
	assertTree(t, parseAndBuild(t, `
      <p style="counter-increment: c;
                display: list-item; list-style: inside decimal">`), []serBox{
		{"p", BlockBoxT, bc{c: []serBox{
			{"p", LineBoxT, bc{c: []serBox{{"p::marker", InlineBoxT, bc{c: []serBox{{"p::marker", TextBoxT, bc{text: "0. "}}}}}}}},
		}}},
	})
}

func TestCounters7(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	exp := func(counter string) serBox {
		return serBox{"p", BlockBoxT, bc{c: []serBox{
			{"p", LineBoxT, bc{c: []serBox{{"p::before", InlineBoxT, bc{c: []serBox{{"p::before", TextBoxT, bc{text: counter}}}}}}}},
		}}}
	}
	var expected []serBox
	for _, counter := range strings.Fields("2.0 2.3 4.3") {
		expected = append(expected, exp(counter))
	}
	// Test that counters are case-sensitive
	// See https://github.com/Kozea/WeasyPrint/pull/827
	assertTree(t, parseAndBuild(t, `
      <style>
        p { counter-increment: p 2 }
        p:before { content: counter(p) "." counter(P); }
      </style>
      <p></p>
      <p style="counter-increment: P 3"></p>
      <p></p>`), expected)
}

func TestCounters8(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	assertTree(t, parseAndBuild(t, `
      <style>
        p:before { content: 'a'; display: list-item }
      </style>
      <p></p>
      <p></p>`), []serBox{
		{"p", BlockBoxT, bc{c: []serBox{
			{"p::before", BlockBoxT, bc{c: []serBox{
				{"p::marker", BlockBoxT, bc{c: []serBox{{"p::marker", LineBoxT, bc{c: []serBox{{"p::marker", TextBoxT, bc{text: "• "}}}}}}}},
				{"p::before", BlockBoxT, bc{c: []serBox{{"p::before", LineBoxT, bc{c: []serBox{{"p::before", TextBoxT, bc{text: "a"}}}}}}}},
			}}},
		}}},
		{"p", BlockBoxT, bc{c: []serBox{
			{"p::before", BlockBoxT, bc{c: []serBox{
				{"p::marker", BlockBoxT, bc{c: []serBox{{"p::marker", LineBoxT, bc{c: []serBox{{"p::marker", TextBoxT, bc{text: "• "}}}}}}}},
				{"p::before", BlockBoxT, bc{c: []serBox{{"p::before", LineBoxT, bc{c: []serBox{{"p::before", TextBoxT, bc{text: "a"}}}}}}}},
			}}},
		}}},
	})
}

func TestCounterStyles1(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	exp := func(counter string) serBox {
		return serBox{"p", BlockBoxT, bc{c: []serBox{
			{"p", LineBoxT, bc{c: []serBox{{"p::before", InlineBoxT, bc{c: []serBox{{"p::before", TextBoxT, bc{text: counter}}}}}}}},
		}}}
	}
	var expected []serBox
	for _, counter := range strings.Fields("--  •  ◦  ▪  -7 Counter:-6 -5:Counter") {
		expected = append(expected, exp(counter))
	}
	assertTree(t, parseAndBuild(t, `
      <style>
        body { --var: 'Counter'; counter-reset: p -12 }
        p { counter-increment: p }
        p:nth-child(1):before { content: '-' counter(p, none) '-'; }
        p:nth-child(2):before { content: counter(p, disc); }
        p:nth-child(3):before { content: counter(p, circle); }
        p:nth-child(4):before { content: counter(p, square); }
        p:nth-child(5):before { content: counter(p); }
        p:nth-child(6):before { content: var(--var) ':' counter(p); }
        p:nth-child(7):before { content: counter(p) ':' var(--var); }
      </style>
      <p></p>
      <p></p>
      <p></p>
      <p></p>
      <p></p>
      <p></p>
      <p></p>
    `), expected)
}

func TestCounterStyles2(t *testing.T) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	exp := func(counter string) serBox {
		return serBox{"p", BlockBoxT, bc{c: []serBox{
			{"p", LineBoxT, bc{c: []serBox{{"p::before", InlineBoxT, bc{c: []serBox{{"p::before", TextBoxT, bc{text: counter}}}}}}}},
		}}}
	}
	var expected []serBox
	for _, counter := range strings.Fields("-1986 -1985  -11 -10 -9 -8  -1 00 01 02  09 10 11 99 100 101  4135 4136") {
		expected = append(expected, exp(counter))
	}

	assertTree(t, parseAndBuild(t, `
      <style>
        p { counter-increment: p }
        p::before { content: counter(p, decimal-leading-zero); }
      </style>
      <p style="counter-reset: p -1987"></p>
      <p></p>
      <p style="counter-reset: p -12"></p>
      <p></p>
      <p></p>
      <p></p>
      <p style="counter-reset: p -2"></p>
      <p></p>
      <p></p>
      <p></p>
      <p style="counter-reset: p 8"></p>
      <p></p>
      <p></p>
      <p style="counter-reset: p 98"></p>
      <p></p>
      <p></p>
      <p style="counter-reset: p 4134"></p>
      <p></p>
    `), expected)
}

func testCounterStyle(t *testing.T, style string, inputs []int, expected string) {
	cp := testutils.CaptureLogs()
	defer cp.AssertNoLogs(t)

	render := tree.UACounterStyle.RenderValue
	var results []string
	for _, value := range inputs {
		results = append(results, render(value, style))
	}
	if !reflect.DeepEqual(results, strings.Fields(expected)) {
		t.Fatalf("unexpected counters for style %s: %v", style, results)
	}
}

func TestCounterStyles(t *testing.T) {
	testCounterStyle(t, "decimal-leading-zero", []int{
		-1986, -1985,
		-11, -10, -9, -8,
		-1, 0, 1, 2,
		9, 10, 11,
		99, 100, 101,
		4135, 4136,
	}, `
        -1986 -1985  -11 -10 -9 -8  -1 00 01 02  09 10 11
        99 100 101  4135 4136
    `)

	testCounterStyle(t, "lower-roman", []int{
		-1986, -1985,
		-1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		49, 50,
		389, 390,
		3489, 3490, 3491,
		4999, 5000, 5001,
	}, `
		-1986 -1985  -1 0 i ii iii iv v vi vii viii ix x xi xii
		xlix l  ccclxxxix cccxc  mmmcdlxxxix mmmcdxc mmmcdxci
		4999 5000 5001
    `)
	testCounterStyle(t, "upper-roman", []int{
		-1986, -1985,
		-1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		49, 50,
		389, 390,
		3489, 3490, 3491,
		4999, 5000, 5001,
	}, `
	        -1986 -1985  -1 0 I II III IV V VI VII VIII IX X XI XII
	        XLIX L  CCCLXXXIX CCCXC  MMMCDLXXXIX MMMCDXC MMMCDXCI
	        4999 5000 5001
    `)

	testCounterStyle(t, "lower-alpha", []int{
		-1986, -1985,
		-1, 0, 1, 2, 3, 4,
		25, 26, 27, 28, 29,
		2002, 2003,
	}, `
		-1986 -1985  -1 0 a b c d  y z aa ab ac bxz bya
    `)

	testCounterStyle(t, "upper-alpha", []int{
		-1986, -1985,
		-1, 0, 1, 2, 3, 4,
		25, 26, 27, 28, 29,
		2002, 2003,
	}, `
		-1986 -1985  -1 0 A B C D  Y Z AA AB AC BXZ BYA
    `)

	testCounterStyle(t, "lower-latin", []int{
		-1986, -1985,
		-1, 0, 1, 2, 3, 4,
		25, 26, 27, 28, 29,
		2002, 2003,
	}, `
		-1986 -1985  -1 0 a b c d  y z aa ab ac bxz bya
    `)

	testCounterStyle(t, "lower-latin", []int{
		-1986, -1985,
		-1, 0, 1, 2, 3, 4,
		25, 26, 27, 28, 29,
		2002, 2003,
	}, `
		-1986 -1985  -1 0 a b c d  y z aa ab ac bxz bya
    `)

	testCounterStyle(t, "upper-latin", []int{
		-1986, -1985,
		-1, 0, 1, 2, 3, 4,
		25, 26, 27, 28, 29,
		2002, 2003,
	}, `
        -1986 -1985  -1 0 A B C D  Y Z AA AB AC BXZ BYA
    `)

	testCounterStyle(t, "georgian", []int{
		-1986, -1985,
		-1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		20, 30, 40, 50, 60, 70, 80, 90, 100,
		200, 300, 400, 500, 600, 700, 800, 900, 1000,
		2000, 3000, 4000, 5000, 6000, 7000, 8000, 9000, 10000,
		19999, 20000, 20001,
	}, `
        -1986 -1985  -1 0 ა
        ბ გ დ ე ვ ზ ჱ თ ი ია იბ
        კ ლ მ ნ ჲ ო პ ჟ რ
        ს ტ ჳ ფ ქ ღ ყ შ ჩ
        ც ძ წ ჭ ხ ჴ ჯ ჰ ჵ
        ჵჰშჟთ 20000 20001
    `)

	testCounterStyle(t, "armenian", []int{
		-1986, -1985,
		-1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		20, 30, 40, 50, 60, 70, 80, 90, 100,
		200, 300, 400, 500, 600, 700, 800, 900, 1000,
		2000, 3000, 4000, 5000, 6000, 7000, 8000, 9000,
		9999, 10000, 10001,
	}, `
        -1986 -1985  -1 0 Ա
        Բ Գ Դ Ե Զ Է Ը Թ Ժ ԺԱ ԺԲ
        Ի Լ Խ Ծ Կ Հ Ձ Ղ Ճ
        Մ Յ Ն Շ Ո Չ Պ Ջ Ռ
        Ս Վ Տ Ր Ց Ւ Փ Ք
        ՔՋՂԹ 10000 10001
    `)
}

// TODO: move these tests in layout package

// @assertNoLogs
// @pytest.mark.parametrize("arguments, values", (
//     ("cyclic "a" "b" "c"", ("a ", "b ", "c ", "a ")),
//     ("symbolic "a" "b"", ("a ", "b ", "aa ", "bb ")),
//     (""a" "b"", ("a ", "b ", "aa ", "bb ")),
//     ("alphabetic "a" "b"", ("a ", "b ", "aa ", "ab ")),
//     ("fixed "a" "b"", ("a ", "b ", "3 ", "4 ")),
//     ("numeric "0" "1" "2"", ("1 ", "2 ", "10 ", "11 ")),
// ))
// func TestCounterSymbols(t *testing.Targuments, values):
//     page, = renderPages(`
//       <style>
//         ol { list-style-type: symbols(%s) }
//       </style>
//       <ol>
//         <li>abc</li>
//         <li>abc</li>
//         <li>abc</li>
//         <li>abc</li>
//       </ol>
//     ` % arguments)
//     html, = page.children
//     body, = html.children
//     ol, = body.children
//     li1, li2, li3, li4 = ol.children
//     assert li1.children[0].children[0].children[0].text == values[0]
//     assert li2.children[0].children[0].children[0].text == values[1]
//     assert li3.children[0].children[0].children[0].text == values[2]
//     assert li4.children[0].children[0].children[0].text == values[3]
// }

// @assertNoLogs
// @pytest.mark.parametrize("styleType, values", (
//     ("decimal", ("1. ", "2. ", "3. ", "4. ")),
//     (""/"", ("/", "/", "/", "/")),
// ))
// func TestListStyleTypes(t *testing.TstyleType, values) {
//     page, = renderPages(`
//       <style>
//         ol { list-style-type: %s }
//       </style>
//       <ol>
//         <li>abc</li>
//         <li>abc</li>
//         <li>abc</li>
//         <li>abc</li>
//       </ol>
//     ` % styleType)
//     html, = page.children
//     body, = html.children
//     ol, = body.children
//     li1, li2, li3, li4 = ol.children
//     assert li1.children[0].children[0].children[0].text == values[0]
//     assert li2.children[0].children[0].children[0].text == values[1]
//     assert li3.children[0].children[0].children[0].text == values[2]
//     assert li4.children[0].children[0].children[0].text == values[3]

// func TestCounterSet(t *testing.T):
//     page, = renderPages(`
//       <style>
//         body { counter-reset: h2 0 h3 4; font-size: 1px }
//         article { counter-reset: h2 2 }
//         h1 { counter-increment: h1 }
//         h1::before { content: counter(h1) }
//         h2 { counter-increment: h2; counter-set: h3 3 }
//         h2::before { content: counter(h2) }
//         h3 { counter-increment: h3 }
//         h3::before { content: counter(h3) }
//       </style>
//       <article>
//         <h1></h1>
//       </article>
//       <article>
//         <h2></h2>
//         <h3></h3>
//       </article>
//       <article>
//         <h3></h3>
//       </article>
//       <article>
//         <h2></h2>
//       </article>
//       <article>
//         <h3></h3>
//         <h3></h3>
//       </article>
//       <article>
//         <h1></h1>
//         <h2></h2>
//         <h3></h3>
//       </article>
//     `)
//     html, = page.children
//     body, = html.children
//     art1, art2, art3, art4, art5, art6 = body.children
// }
//     h1, = art1.children
//     assert h1.children[0].children[0].children[0].text == "1"

//     h2, h3, = art2.children
//     assert h2.children[0].children[0].children[0].text == "3"
//     assert h3.children[0].children[0].children[0].text == "4"

//     h3, = art3.children
//     assert h3.children[0].children[0].children[0].text == "5"

//     h2, = art4.children
//     assert h2.children[0].children[0].children[0].text == "3"

//     h31, h32 = art5.children
//     assert h31.children[0].children[0].children[0].text == "4"
//     assert h32.children[0].children[0].children[0].text == "5"

//     h1, h2, h3 = art6.children
//     assert h1.children[0].children[0].children[0].text == "1"
//     assert h2.children[0].children[0].children[0].text == "3"
//     assert h3.children[0].children[0].children[0].text == "4"

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
