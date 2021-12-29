package pdf

import (
	"testing"

	"github.com/benoitkugler/webrender/utils/testutils"
)

//  Test how leaders are drawn.

func TestLeaderSimple(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	expectedPixels := `
        RR__BBBBBBBB__BB
        RR__BBBBBBBB__BB
        RRRR__BBBB__BBBB
        RRRR__BBBB__BBBB
        RR__BBBB__BBBBBB
        RR__BBBB__BBBBBB
    `
	html := `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page {
          background: white;
          size: 16px 6px;
        }
        body {
          color: red;
          counter-reset: count;
          font-family: weasyprint;
          font-size: 2px;
          line-height: 1;
        }
        div::after {
          color: blue;
          content: ' ' leader(dotted) ' ' counter(count, lower-roman);
          counter-increment: count;
        }
      </style>
      <div>a</div>
      <div>bb</div>
      <div>c</div>
    `
	assertPixelsEqual(t, "leader-simple", expectedPixels, html)
}

func TestLeaderTooLong(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	expectedPixels := `
        RRRRRRRRRR______
        RRRRRRRRRR______
        BBBBBBBBBBBB__BB
        BBBBBBBBBBBB__BB
        RR__RR__RR__RR__
        RR__RR__RR__RR__
        RR__RR__RR______
        RR__RR__RR______
        BBBBBBBBBB__BBBB
        BBBBBBBBBB__BBBB
        RR__RR__RR__RR__
        RR__RR__RR__RR__
        RR__BBBB__BBBBBB
        RR__BBBB__BBBBBB
    `
	html := `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page {
          background: white;
          size: 16px 14px;
        }
        body {
          color: red;
          counter-reset: count;
          font-family: weasyprint;
          font-size: 2px;
          line-height: 1;
        }
        div::after {
          color: blue;
          content: ' ' leader(dotted) ' ' counter(count, lower-roman);
          counter-increment: count;
        }
      </style>
      <div>aaaaa</div>
      <div>a a a a a a a</div>
      <div>a a a a a</div>
    `
	assertPixelsEqual(t, "leader-too-long", expectedPixels, html)
}

func TestLeaderAlone(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	expectedPixels := `
        RRBBBBBBBBBBBBBB
        RRBBBBBBBBBBBBBB
    `
	html := `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page {
          background: white;
          size: 16px 2px;
        }
        body {
          color: red;
          font-family: weasyprint;
          font-size: 2px;
          line-height: 1;
        }
        div::after {
          color: blue;
          content: leader(dotted);
        }
      </style>
      <div>a</div>
    `
	assertPixelsEqual(t, "leader-alone", expectedPixels, html)
}

func TestLeaderContent(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	expectedPixels := `
        RR____BB______BB
        RR____BB______BB
    `
	html := `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page {
          background: white;
          size: 16px 2px;
        }
        body {
          color: red;
          font-family: weasyprint;
          font-size: 2px;
          line-height: 1;
        }
        div::after {
          color: blue;
          content: leader(' . ') 'a';
        }
      </style>
      <div>a</div>
    `
	assertPixelsEqual(t, "leader-content", expectedPixels, html)
}

// @pytest.mark.xfail

// func TestLeaderFloat(t *testing.T) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     expectedPixels := `
//         bbGRR___BB____BB
//         bbGRR___BB____BB
//         GGGRR___BB____BB
//         ___RR___BB____BB
//     `
//     html := `
//       <style>
//         @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
//         @page {
//           background: white;
//           size: 16px 4px;
//         }
//         body {
//           color: red;
//           font-family: weasyprint;
//           font-size: 2px;
//           line-height: 1;
//         }
//         article {
//           background: lime;
//           color: navy;
//           float: left;
//           height: 3px;
//           width: 3px;
//         }
//         div::after {
//           color: blue;
//           content: leader('. ') 'a';
//         }
//       </style>
//       <div>a<article>a</article></div>
//       <div>a</div>
//     `
//     assertPixelsEqual(t, "leader-float" , expectedPixels, html)
// }

func TestLeaderInInline(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	expectedPixels := `
        RR__GGBBBBBB__RR
        RR__GGBBBBBB__RR
    `
	html := `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page {
          background: white;
          size: 16px 2px;
        }
        body {
          color: red;
          font-family: weasyprint;
          font-size: 2px;
          line-height: 1;
        }
        span {
          color: lime;
        }
        span::after {
          color: blue;
          content: leader('-');
        }
      </style>
      <div>a <span>a</span> a</div>
    `
	assertPixelsEqual(t, "leader-in-inline", expectedPixels, html)
}

// @pytest.mark.xfail

// func TestLeaderBadAlignment(t *testing.T) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     expectedPixels := `
//         RRRRRR__________
//         RRRRRR__________
//         ______BB______RR
//         ______BB______RR
//     `
//     html := `
//       <style>
//         @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
//         @page {
//           background: white;
//           size: 16px 4px;
//         }
//         body {
//           color: red;
//           font-family: weasyprint;
//           font-size: 2px;
//           line-height: 1;
//         }
//         div::after {
//           color: blue;
//           content: leader(' - ') 'a';
//         }
//       </style>
//       <div>aaa</div>
//     `
//     assertPixelsEqual(t, 'leader-in-inline', 16, 4, expectedPixels, html)

func TestLeaderSimpleRtl(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	expectedPixels := `
        BB__BBBBBBBB__RR
        BB__BBBBBBBB__RR
        BBBB__BBBB__RRRR
        BBBB__BBBB__RRRR
        BBBBBB__BBBB__RR
        BBBBBB__BBBB__RR
    `
	html := `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page {
          background: white;
          size: 16px 6px;
        }
        body {
          color: red;
          counter-reset: count;
          direction: rtl;
          font-family: weasyprint;
          font-size: 2px;
          line-height: 1;
        }
        div::after {
          color: blue;
          /* RTL Mark used in second space */
          content: ' ' leader(dotted) '‏ ' counter(count, lower-roman);
          counter-increment: count;
        }
      </style>
      <div>a</div>
      <div>bb</div>
      <div>c</div>
    `
	assertPixelsEqual(t, "leader-simple-rtl", expectedPixels, html)
}

func TestLeaderTooLongRtl(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	expectedPixels := `
        ______RRRRRRRRRR
        ______RRRRRRRRRR
        BB__BBBBBBBBBBBB
        BB__BBBBBBBBBBBB
        __RR__RR__RR__RR
        __RR__RR__RR__RR
        ______RR__RR__RR
        ______RR__RR__RR
        BBBB__BBBBBBBBBB
        BBBB__BBBBBBBBBB
        __RR__RR__RR__RR
        __RR__RR__RR__RR
        BBBBBB__BBBB__RR
        BBBBBB__BBBB__RR
    `
	html := `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page {
          background: white;
          size: 16px 14px;
        }
        body {
          color: red;
          counter-reset: count;
          direction: rtl;
          font-family: weasyprint;
          font-size: 2px;
          line-height: 1;
        }
        div::after {
          color: blue;
          /* RTL Mark used in second space */
          content: ' ' leader(dotted) '‏ ' counter(count, lower-roman);
          counter-increment: count;
        }
      </style>
      <div>aaaaa</div>
      <div>a a a a a a a</div>
      <div>a a a a a</div>
    `
	assertPixelsEqual(t, "leader-too-long-rtl", expectedPixels, html)
}

func TestLeaderFloatLeader(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Test regression: https://github.com/Kozea/WeasyPrint/issues/1409
	// Leaders in floats are not displayed at all in many cases with the current
	// implementation, and this case is not really specified. So…
	expectedPixels := `
        RR____________BB
        RR____________BB
        RRRR__________BB
        RRRR__________BB
        RR____________BB
        RR____________BB
    `
	html := `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page {
          background: white;
          size: 16px 6px;
        }
        body {
          color: red;
          font-family: weasyprint;
          font-size: 2px;
          line-height: 1;
        }
        div::after {
          color: blue;
          content: leader(' . ') 'a';
          float: right;
        }
      </style>
      <div>a</div>
      <div>bb</div>
      <div>c</div>
    `
	assertPixelsEqual(t, "leader-float-leader", expectedPixels, html)
}
