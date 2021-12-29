package pdf

import (
	"fmt"
	"testing"

	"github.com/benoitkugler/webrender/utils/testutils"
)

// Test overflow and clipping.

func TestOverflow_1(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// See test_images
	assertPixelsEqual(t, "inline_image_overflow", `
        ________
        ________
        __rBBB__
        __BBBB__
        ________
        ________
        ________
        ________
    `, `
      <style>
        @page { size: 8px }
        body { margin: 2px 0 0 2px; background: #fff; font-size:0 }
        div { height: 2px; overflow: hidden }
      </style>
      <div><img src="../resources_test/pattern.png"></div>`)
}

func TestOverflow_2(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// <body> is only 1px high, but its overflow is propageted to the viewport
	// ie. the padding edge of the page box.
	assertPixelsEqual(t, "inline_image_viewport_overflow", `
        ________
        ________
        __rBBB__
        __BBBB__
        __BBBB__
        ________
        ________
        ________
    `, `
      <style>
        @page { size: 8px; background: #fff; margin: 2px 2px 3px 2px }
        body { height: 1px; overflow: hidden; font-size: 0 }
      </style>
      <div><img src="../resources_test/pattern.png"></div>`)
}

func TestOverflow_3(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Assert that the border is not clipped by overflow: hidden
	assertPixelsEqual(t, "border_box_overflow", `
        ________
        ________
        __BBBB__
        __B__B__
        __B__B__
        __BBBB__
        ________
        ________
    `, `
      <style>
        @page { size: 8px; background: #fff; margin: 2px; }
        div { width: 2px; height: 2px; overflow: hidden;
              border: 1px solid blue; }
      </style>
      <div></div>`)
}

func TestOverflow_4(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Assert that the page margins aren't clipped by body's overflow
	assertPixelsEqual(t, "border_box_overflow", `
        rr______
        rr______
        __BBBB__
        __BBBB__
        __BBBB__
        __BBBB__
        ________
        ________
    `, `
      <style>
        @page {
          size: 8px;
          margin: 2px;
          background:#fff;
          @top-left-corner { content: ''; background:#f00; } }
        body { overflow: auto; background:#00f; }
      </style>
    `)
}

func TestClip(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, data := range [][2]string{

		{"5px, 5px, 9px, auto", `
				______________
				______________
				______________
				______________
				______________
				______________
				______rBBBrBg_
				______BBBBBBg_
				______BBBBBBg_
				______BBBBBBg_
				______________
				______________
				______________
				______________
				______________
				______________
			`},
		{"5px, 5px, auto, 10px", `
				______________
				______________
				______________
				______________
				______________
				______________
				______rBBBr___
				______BBBBB___
				______BBBBB___
				______BBBBB___
				______rBBBr___
				______BBBBB___
				______ggggg___
				______________
				______________
				______________
			`},
		{"5px, auto, 9px, 10px", `
				______________
				______________
				______________
				______________
				______________
				______________
				_grBBBrBBBr___
				_gBBBBBBBBB___
				_gBBBBBBBBB___
				_gBBBBBBBBB___
				______________
				______________
				______________
				______________
				______________
				______________
			`},
		{"auto, 5px, 9px, 10px", `
				______________
				______ggggg___
				______rBBBr___
				______BBBBB___
				______BBBBB___
				______BBBBB___
				______rBBBr___
				______BBBBB___
				______BBBBB___
				______BBBBB___
				______________
				______________
				______________
				______________
				______________
				______________
			`},
	} {

		css, pixels := data[0], data[1]
		assertPixelsEqual(t, "background_repeat_clipped", pixels, fmt.Sprintf(`
      <style>
        @page { size: 14px 16px; background: #fff }
        div { margin: 1px; border: 1px green solid;
              background: url(../resources_test/pattern.png);
              position: absolute; /* clip only applies on abspos */
              top: 0; bottom: 2px; left: 0; right: 0;
              clip: rect(%s); }
      </style>
      <div>`, css))
	}
}
