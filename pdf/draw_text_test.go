package pdf

import (
	"os"
	"testing"

	"github.com/benoitkugler/go-weasyprint/utils/testutils"
)

// Test how text is drawn.

func TestTextOverflowClip(t *testing.T) {
	assertPixelsEqual(t, "text_overflow", `
        _________
        _RRRRRRR_
        _RRRRRRR_
        _________
        _RR__RRR_
        _RR__RRR_
        _________
    `, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page {
          size: 9px 7px;
          background: white;
        }
        body {
          color: red;
          font-family: weasyprint;
          font-size: 2px;
        }
        div {
          line-height: 1;
          margin: 1px;
          overflow: hidden;
          width: 3.5em;
        }
      </style>
      <div>abcde</div>
      <div style="white-space: nowrap">a bcde</div>`)
}

func TestTextOverflowEllipsis(t *testing.T) {
	assertPixelsEqual(t, "text_overflow", `
        _________
        _RRRRRR__
        _RRRRRR__
        _________
        _RR__RR__
        _RR__RR__
        _________
        _RRRRRR__
        _RRRRRR__
        _________
        _RRRRRRR_
        _RRRRRRR_
        _________
        _RRRRRRR_
        _RRRRRRR_
        _________
    `, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page {
          background: white;
          size: 9px 16px;
        }
        body {
          color: red;
          font-family: weasyprint;
          font-size: 2px;
        }
        div {
          line-height: 1;
          margin: 1px;
          overflow: hidden;
          text-overflow: ellipsis;
          width: 3.5em;
        }
        div div {
          margin: 0;
        }
      </style>
      <div>abcde</div>
      <div style="white-space: nowrap">a bcde</div>
      <div><span>a<span>b</span>cd</span>e</div>
      <div><div style="text-overflow: clip">abcde</div></div>
      <div><div style="overflow: visible">abcde</div></div>
`)
}

func TestTextAlignRtlTrailingWhitespace(t *testing.T) {
	// Test text alignment for rtl text with trailing space.
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/1111
	assertPixelsEqual(t, "text_overflow", `
        _________
        _rrrrBBB_
        _________
        _rrrrBBB_
        _________
        _BBBrrrr_
        _________
        _BBBrrrr_
        _________
    `, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page { background: white; size: 9px }
        body { font-family: weasyprint; color: blue; font-size: 1px }
        p { background: red; line-height: 1; width: 7em; margin: 1em }
      </style>
      <!-- &#8207 forces Unicode RTL direction for the following chars -->
      <p style="direction: rtl"> abc </p>
      <p style="direction: rtl"> &#8207;abc </p>
      <p style="direction: ltr"> abc </p>
      <p style="direction: ltr"> &#8207;abc </p>
    `)
}

func TestMaxLinesEllipsis(t *testing.T) {
	assertPixelsEqual(t, "max_lines_ellipsis", `
        BBBBBBBB__
        BBBBBBBB__
        BBBBBBBBBB
        BBBBBBBBBB
        __________
        __________
        __________
        __________
        __________
        __________
    `, `
      <style>
        @page {size: 10px 10px;}
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        p {
          block-ellipsis: auto;
          color: blue;
          font-family: weasyprint;
          font-size: 2px;
          max-lines: 2;
        }
      </style>
      <p>
        abcd efgh ijkl
      </p>
    `)
}

// @pytest.mark.xfail
// func TestMaxLinesNested(t *testing.T) {
//     assertPixelsEqual(t, "max_lines_nested", 10, 12, `
//         BBBBBBBBBB
//         BBBBBBBBBB
//         BBBBBBBBBB
//         BBBBBBBBBB
//         rrrrrrrrrr
//         rrrrrrrrrr
//         rrrrrrrrrr
//         rrrrrrrrrr
//         BBBBBBBBBB
//         BBBBBBBBBB
//         __________
//         __________
//     `, `
//       <style>
//         @page {size: 10px 12px;}
//         @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
//         div {
//           continue: discard;
//           font-family: weasyprint;
//           font-size: 2px;
//         }
//         #a {
//           color: blue;
//           max-lines: 5;
//         }
//         #b {
//           color: red
//           max-lines: 2;
//         }
//       </style>
//       <div id=a>
//         aaaaa
//         aaaaa
//         <div id=b>
//           bbbbb
//           bbbbb
//           bbbbb
//           bbbbb
//         </div>
//         aaaaa
//         aaaaa
//       </div>
//     `)

func TestLineClamp(t *testing.T) {
	assertPixelsEqual(t, "line_clamp", `
        BBBB__BB__
        BBBB__BB__
        BBBB__BB__
        BBBB__BB__
        BBBBBBBBBB
        BBBBBBBBBB
        __________
        __________
        __________
        __________
    `, `
      <style>
        @page {size: 10px 10px;}
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        p {
          color: blue;
          font-family: weasyprint;
          font-size: 2px;
          line-clamp: 3 "(…)";
        }
      </style>

      <p>
        aa a
        bb b
        cc c
        dddd
        eeee
        ffff
        gggg
        hhhh
      </p>
    `)
}

func TestLineClampNone(t *testing.T) {
	assertPixelsEqual(t, "line_clamp_none", `
        BBBB__BB__
        BBBB__BB__
        BBBB__BB__
        BBBB__BB__
        BBBB__BB__
        BBBB__BB__
        __________
        __________
        __________
        __________
    `, `
      <style>
        @page {size: 10px 10px;}
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        p {
          color: blue;
          font-family: weasyprint;
          font-size: 2px;
          max-lines: 1;
          continue: discard;
          block-ellipsis: "…";
          line-clamp: none;
        }
      </style>

      <p>
        aa a
        bb b
        cc c
      </p>
    `)
}

func TestLineClampNumber(t *testing.T) {
	assertPixelsEqual(t, "line_clamp_number", `
        BBBB__BB__
        BBBB__BB__
        BBBB__BB__
        BBBB__BB__
        BBBB__BBBB
        BBBB__BBBB
        __________
        __________
        __________
        __________
    `, `
      <style>
        @page {size: 10px 10px;}
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        p {
          color: blue;
          font-family: weasyprint;
          font-size: 2px;
          line-clamp: 3;
        }
      </style>

      <p>
        aa a
        bb b
        cc c
        dddd
        eeee
      </p>
    `)
}

// @pytest.mark.xfail
// func TestEllipsisNested(t *testing.T) {
//     assertPixelsEqual(t, "ellipsis_nested", 10, 10, `
//         BBBBBB____
//         BBBBBB____
//         BBBBBB____
//         BBBBBB____
//         BBBBBB____
//         BBBBBB____
//         BBBBBB____
//         BBBBBB____
//         BBBBBBBB__
//         BBBBBBBB__
//     `, `
//       <style>
//         @page {size: 10px 10px;}
//         @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
//         div {
//           block-ellipsis: auto;
//           color: blue;
//           continue: discard;
//           font-family: weasyprint;
//           font-size: 2px;
//         }
//       </style>
//       <div>
//         <p>aaa</p>
//         <p>aaa</p>
//         <p>aaa</p>
//         <p>aaa</p>
//         <p>aaa</p>
//         <p>aaa</p>
//       </div>
//     `)

func TestTextAlignRight(t *testing.T) {
	assertPixelsEqual(t, "text_align_right", `
        _________
        __RR__RR_
        __RR__RR_
        ______RR_
        ______RR_
        _________
    `, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page {
          size: 9px 6px;
          background: white;
        }
        body {
          color: red;
          font-family: weasyprint;
          font-size: 2px;
        }
        div {
          line-height: 1;
          margin: 1px;
          text-align: right;
        }
      </style>
      <div>a c e</div>`)
}

func TestTextAlignJustify(t *testing.T) {
	assertPixelsEqual(t, "text_align_justify", `
        _________
        _RR___RR_
        _RR___RR_
        _RR______
        _RR______
        _________
    `, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page {
          size: 9px 6px;
          background: white;
        }
        body {
          color: red;
          font-family: weasyprint;
          font-size: 2px;
        }
        div {
          line-height: 1;
          margin: 1px;
          text-align: justify;
        }
      </style>
      <div>a c e</div>`)
}

func TestTextWordSpacing(t *testing.T) {
	assertPixelsEqual(t, "text_word_spacing", `
        ___________________
        _RR____RR____RR____
        _RR____RR____RR____
        ___________________
    `, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page {
          size: 19px 4px;
          background: white;
        }
        body {
          color: red;
          font-family: weasyprint;
          font-size: 2px;
        }
        div {
          line-height: 1;
          margin: 1px;
          word-spacing: 1em;
        }
      </style>
      <div>a c e</div>`)
}

func TestTextLetterSpacing(t *testing.T) {
	assertPixelsEqual(t, "text_letter_spacing", `
        ___________________
        _RR____RR____RR____
        _RR____RR____RR____
        ___________________
    `, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page {
          size: 19px 4px;
          background: white;
        }
        body {
          color: red;
          font-family: weasyprint;
          font-size: 2px;
        }
        div {
          line-height: 1;
          margin: 1px;
          letter-spacing: 2em;
        }
      </style>
      <div>ace</div>`)
}

func TestTextUnderline(t *testing.T) {
	assertPixelsEqual(t, "text_underline", `
        _____________
        _zzzzzzzzzzz_
        _zRRRRRRRRRz_
        _zRRRRRRRRRz_
        _zBBBBBBBBBz_
        _zzzzzzzzzzz_
        _____________
    `, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page {
          size: 13px 7px;
          background: white;
          margin: 2px;
        }
        body {
          color: red;
          font-family: weasyprint;
          font-size: 3px;
          text-decoration: underline blue;
        }
      </style>
      <div>abc</div>`)
}

func TestTextOverline(t *testing.T) {
	// Ascent value seems to be a bit random, don’t try to get the exact
	// position of the line
	assertPixelsEqual(t, "text_overline", `
        _____________
        _zzzzzzzzzzz_
        _zzzzzzzzzzz_
        _zRRRRRRRRRz_
        _zRRRRRRRRRz_
        _zzzzzzzzzzz_
        _____________
    `, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page {
          size: 13px 7px;
          background: white;
          margin: 2px;
        }
        body {
          color: red;
          font-family: weasyprint;
          font-size: 3px;
          text-decoration: overline blue;
        }
      </style>
      <div>abc</div>`)
}

func TestTextLineThrough(t *testing.T) {
	assertPixelsEqual(t, "text_line_through", `
        _____________
        _zzzzzzzzzzz_
        _zRRRRRRRRRz_
        _zBBBBBBBBBz_
        _zRRRRRRRRRz_
        _zzzzzzzzzzz_
        _____________
    `, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page {
          size: 13px 7px;
          background: white;
          margin: 2px;
        }
        body {
          color: red;
          font-family: weasyprint;
          font-size: 3px;
          text-decoration: line-through blue;
        }
      </style>
      <div>abc</div>`)
}

func TestZeroWidthCharacter(t *testing.T) {
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/1508
	assertPixelsEqual(t, "zero_width_character", `
        ______
        _RRRR_
        _RRRR_
        ______
    `, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page {
          size: 6px 4px;
          background: white;
          margin: 1px;
        }
        body {
          color: red;
          font-family: weasyprint;
          font-size: 2px;
          line-height: 1;
        }
      </style>
      <div>a&zwnj;b</div>`)
}

func TestTextUnderlineDashed(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	f := htmlToPDF(t, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page {
          size: 13px 7px;
          background: white;
          margin: 2px;
        }
        body {
          color: red;
          font-family: weasyprint;
          font-size: 3px;
          text-decoration: underline dashed blue;
        }
      </style
      <div>abc</div>`, 1)
	f.Close()
	os.Remove(f.Name())
}

func TestTextUnderlineDotted(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	f := htmlToPDF(t, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page {
          size: 13px 7px;
          background: white;
          margin: 2px;
        }
        body {
          color: red;
          font-family: weasyprint;
          font-size: 3px;
          text-decoration: underline dotted blue;
        }
      </style
      <div>abc</div>`, 1)

	f.Close()
	os.Remove(f.Name())
}
