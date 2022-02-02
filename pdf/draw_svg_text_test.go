package pdf

import (
	"testing"

	"github.com/benoitkugler/webrender/utils/testutils"
)

// Test how SVG text is drawn.

func TestTextFill(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "text_fill", `
        BBBBBB__BBBBBB______
        BBBBBB__BBBBBB______
    `, `
      <style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page { size: 20px 2px }
        svg { display: block }
      </style>
      <svg width="20px" height="2px" xmlns="http://www.w3.org/2000/svg">
        <text x="0" y="1.5" font-family="weasyprint" font-size="2" fill="blue">
          ABC DEF
        </text>
      </svg>
    `)
}

func TestTextStroke(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "text_stroke", `
        _BBBBBBBBBBBB_______
        _BBBBBBBBBBBB_______
        _BBBBBBBBBBBB_______
        _BBBBBBBBBBBB_______
    `, `
      <style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page { size: 20px 4px }
        svg { display: block }
      </style>
      <svg width="20px" height="4px" xmlns="http://www.w3.org/2000/svg">
        <text x="2" y="2.5" font-family="weasyprint" font-size="2"
              fill="transparent" stroke="blue" stroke-width="2">
          A B C
        </text>
      </svg>
    `)
}

func TestTextX(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "text_x", `
        BB__BB_BBBB_________
        BB__BB_BBBB_________
    `, `
      <style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page { size: 20px 2px }
        svg { display: block }
      </style>
      <svg width="20px" height="2px" xmlns="http://www.w3.org/2000/svg">
        <text x="0 4 7" y="1.5" font-family="weasyprint" font-size="2"
              fill="blue">
          ABCD
        </text>
      </svg>
    `)
}

func TestTextY(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "text_y", `
        __________BBBBB_____BBBBBBBBBB
        __________BBBBB_____BBBBBBBBBB
        __________BBBBB_____BBBBBBBBBB
        __________BBBBB_____BBBBBBBBBB
        __________BBBBB_____BBBBBBBBBB
        BBBBBBBBBB_____BBBBB__________
        BBBBBBBBBB_____BBBBB__________
        BBBBBBBBBB_____BBBBB__________
        BBBBBBBBBB_____BBBBB__________
        BBBBBBBBBB_____BBBBB__________
    `, `
      <style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page { size: 30px 10px }
        svg { display: block }
      </style>
      <svg width="30px" height="10px" xmlns="http://www.w3.org/2000/svg">
        <text x="0" y="9 9 4 9 4" font-family="weasyprint" font-size="5"
              fill="blue">
          ABCDEF
        </text>
      </svg>
    `)
}

func TestTextXy(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "text_xy", `
        __________BBBBB_____BBBBBBBBBB
        __________BBBBB_____BBBBBBBBBB
        __________BBBBB_____BBBBBBBBBB
        __________BBBBB_____BBBBBBBBBB
        __________BBBBB_____BBBBBBBBBB
        BBBBB__________BBBBB__________
        BBBBB__________BBBBB__________
        BBBBB__________BBBBB__________
        BBBBB__________BBBBB__________
        BBBBB__________BBBBB__________
    `, `
      <style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page { size: 30px 10px }
        svg { display: block }
      </style>
      <svg width="30px" height="10px" xmlns="http://www.w3.org/2000/svg">
        <text x="0 10" y="9 4 9 4" font-family="weasyprint" font-size="5"
              fill="blue">
          ABCDE
        </text>
      </svg>
    `)
}

func TestTextDx(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "text_dx", `
        BB__BB_BBBB_________
        BB__BB_BBBB_________
    `, `
      <style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page { size: 20px 2px }
        svg { display: block }
      </style>
      <svg width="20px" height="2px" xmlns="http://www.w3.org/2000/svg">
        <text dx="0 2 1" y="1.5" font-family="weasyprint" font-size="2"
              fill="blue">
          ABCD
        </text>
      </svg>
    `)
}

func TestTextDy(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "text_dy", `
        __________BBBBB_____BBBBBBBBBB
        __________BBBBB_____BBBBBBBBBB
        __________BBBBB_____BBBBBBBBBB
        __________BBBBB_____BBBBBBBBBB
        __________BBBBB_____BBBBBBBBBB
        BBBBBBBBBB_____BBBBB__________
        BBBBBBBBBB_____BBBBB__________
        BBBBBBBBBB_____BBBBB__________
        BBBBBBBBBB_____BBBBB__________
        BBBBBBBBBB_____BBBBB__________
    `, `
      <style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page { size: 30px 10px }
        svg { display: block }
      </style>
      <svg width="30px" height="10px" xmlns="http://www.w3.org/2000/svg">
        <text x="0" dy="9 0 -5 5 -5" font-family="weasyprint" font-size="5"
              fill="blue">
          ABCDEF
        </text>
      </svg>
    `)
}

func TestTextDxDy(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "text_dx_dy", `
        __________BBBBB_____BBBBBBBBBB
        __________BBBBB_____BBBBBBBBBB
        __________BBBBB_____BBBBBBBBBB
        __________BBBBB_____BBBBBBBBBB
        __________BBBBB_____BBBBBBBBBB
        BBBBB__________BBBBB__________
        BBBBB__________BBBBB__________
        BBBBB__________BBBBB__________
        BBBBB__________BBBBB__________
        BBBBB__________BBBBB__________
    `, `
      <style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page { size: 30px 10px }
        svg { display: block }
      </style>
      <svg width="30px" height="10px" xmlns="http://www.w3.org/2000/svg">
        <text dx="0 5" dy="9 -5 5 -5" font-family="weasyprint" font-size="5"
              fill="blue">
          ABCDE
        </text>
      </svg>
    `)
}

func TestTextAnchorStart(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "text_anchor_start", `
        __BBBBBB____________
        __BBBBBB____________
        ____BBBBBB__________
        ____BBBBBB__________
    `, `
      <style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page { size: 20px 4px }
        svg { display: block }
      </style>
      <svg width="20px" height="4px" xmlns="http://www.w3.org/2000/svg">
        <text x="2" y="1.5" font-family="weasyprint" font-size="2"
              fill="blue">
          ABC
        </text>
        <text x="4" y="3.5" font-family="weasyprint" font-size="2"
              fill="blue" text-anchor="start">
          ABC
        </text>
      </svg>
    `)
}

func TestTextAnchorMiddle(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "text_anchor_middle", `
        _______BBBBBB_______
        _______BBBBBB_______
    `, `
      <style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page { size: 20px 2px }
        svg { display: block }
      </style>
      <svg width="20px" height="2px" xmlns="http://www.w3.org/2000/svg">
        <text x="10" y="1.5" font-family="weasyprint" font-size="2"
              fill="blue" text-anchor="middle">
          ABC
        </text>
      </svg>
    `)
}

func TestTextAnchorEnd(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "text_anchor_end", `
        ____________BBBBBB__
        ____________BBBBBB__
    `, `
      <style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page { size: 20px 2px }
        svg { display: block }
      </style>
      <svg width="20px" height="2px" xmlns="http://www.w3.org/2000/svg">
        <text x="18" y="1.5" font-family="weasyprint" font-size="2"
              fill="blue" text-anchor="end">
          ABC
        </text>
      </svg>
    `)
}

func TestTextTspan(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "text_tspan", `
        BBBBBB__BBBBBB______
        BBBBBB__BBBBBB______
    `, `
      <style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page { size: 20px 2px }
        svg { display: block }
      </style>
      <svg width="20px" height="2px" xmlns="http://www.w3.org/2000/svg">
        <text x="10" y="10" font-family="weasyprint" font-size="2" fill="blue">
          <tspan x="0" y="1.5">ABC DEF</tspan>
        </text>
      </svg>
    `)
}
