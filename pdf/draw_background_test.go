package pdf

import (
	"fmt"
	"strings"
	"testing"

	tu "github.com/benoitkugler/webrender/utils/testutils"
)

// Test how backgrounds are drawn.

func TestCanvasBackground(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	for _, data := range [][3]string{
		{"all_blue", strings.Repeat(strings.Repeat("B", 10)+"\n", 10), `
								<style>
								  @page { size: 10px }
								  /* body’s background propagates to the whole canvas */
								  body { margin: 2px; background: #00f; height: 5px }
								</style>
							  <body>`},
		{"blocks", `
					rrrrrrrrrr
					rrrrrrrrrr
					rrBBBBBBrr
					rrBBBBBBrr
					rrBBBBBBrr
					rrBBBBBBrr
					rrBBBBBBrr
					rrrrrrrrrr
					rrrrrrrrrr
					rrrrrrrrrr
				`, `
					<style>
						@page { size: 10px }
						/* html’s background propagates to the whole canvas */
						html { padding: 1px; background: #f00 }
						/* html has a background, so body’s does not propagate */
						body { margin: 1px; background: #00f; height: 5px }
					</style>
					<body>`},
	} {
		assertPixelsEqual(t, data[1], data[2])
	}
}

func TestCanvasBackgroundSize(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	expectedPixels := `
        __________
        __________
        __RRRRRR__
        __RGGGGR__
        __RRRRRR__
        __BBBBBB__
        __BBBBBB__
        __BBBBBB__
        __________
        __________
    `
	html := `
      <style>
         @page { size: 10px; margin: 2px; background: white }
         /* html’s background propagates to the whole canvas */
         html { background: linear-gradient(
           red 0, red 50%, blue 50%, blue 100%); }
         /* html has a background, so body’s does not propagate */
         body { margin: 1px; background: lime; height: 1px }
      </style>
      <body>`
	assertPixelsEqual(t, expectedPixels, html)
}

func TestBackgroundImage(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, data := range [][3]string{
		{"repeat", "url(../resources_test/pattern.png)", `
			______________
			______________
			__rBBBrBBBrB__
			__BBBBBBBBBB__
			__BBBBBBBBBB__
			__BBBBBBBBBB__
			__rBBBrBBBrB__
			__BBBBBBBBBB__
			__BBBBBBBBBB__
			__BBBBBBBBBB__
			__rBBBrBBBrB__
			__BBBBBBBBBB__
			______________
			______________
			______________
			______________
		`},
		{"repeat_x", "url(../resources_test/pattern.png) repeat-x", `
			______________
			______________
			__rBBBrBBBrB__
			__BBBBBBBBBB__
			__BBBBBBBBBB__
			__BBBBBBBBBB__
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
		`},
		{"repeat_y", "url(../resources_test/pattern.png) repeat-y", `
			______________
			______________
			__rBBB________
			__BBBB________
			__BBBB________
			__BBBB________
			__rBBB________
			__BBBB________
			__BBBB________
			__BBBB________
			__rBBB________
			__BBBB________
			______________
			______________
			______________
			______________
		`},

		{"left_top", "url(../resources_test/pattern.png) no-repeat 0 0%", `
			______________
			______________
			__rBBB________
			__BBBB________
			__BBBB________
			__BBBB________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
		`},
		{"center_top", "url(../resources_test/pattern.png) no-repeat 50% 0px", `
			______________
			______________
			_____rBBB_____
			_____BBBB_____
			_____BBBB_____
			_____BBBB_____
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
		`},
		{"right_top", "url(../resources_test/pattern.png) no-repeat 6px top", `
			______________
			______________
			________rBBB__
			________BBBB__
			________BBBB__
			________BBBB__
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
		`},
		{"bottom_6_right_0", "url(../resources_test/pattern.png) no-repeat bottom 6px right 0", `
			______________
			______________
			________rBBB__
			________BBBB__
			________BBBB__
			________BBBB__
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
		`},
		{"left_center", "url(../resources_test/pattern.png) no-repeat left center", `
			______________
			______________
			______________
			______________
			______________
			__rBBB________
			__BBBB________
			__BBBB________
			__BBBB________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
		`},
		{"center_left", "url(../resources_test/pattern.png) no-repeat center left", `
			______________
			______________
			______________
			______________
			______________
			__rBBB________
			__BBBB________
			__BBBB________
			__BBBB________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
		`},
		{"center_center", "url(../resources_test/pattern.png) no-repeat 3px 3px", `
			______________
			______________
			______________
			______________
			______________
			_____rBBB_____
			_____BBBB_____
			_____BBBB_____
			_____BBBB_____
			______________
			______________
			______________
			______________
			______________
			______________
			______________
		`},
		{"right_center", "url(../resources_test/pattern.png) no-repeat 100% 50%", `
			______________
			______________
			______________
			______________
			______________
			________rBBB__
			________BBBB__
			________BBBB__
			________BBBB__
			______________
			______________
			______________
			______________
			______________
			______________
			______________
		`},

		{"left_bottom", "url(../resources_test/pattern.png) no-repeat 0% bottom", `
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			__rBBB________
			__BBBB________
			__BBBB________
			__BBBB________
			______________
			______________
			______________
			______________
		`},
		{"center_bottom", "url(../resources_test/pattern.png) no-repeat center 6px", `
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			_____rBBB_____
			_____BBBB_____
			_____BBBB_____
			_____BBBB_____
			______________
			______________
			______________
			______________
		`},
		{"bottom_center", "url(../resources_test/pattern.png) no-repeat bottom center", `
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			_____rBBB_____
			_____BBBB_____
			_____BBBB_____
			_____BBBB_____
			______________
			______________
			______________
			______________
		`},
		{"right_bottom", "url(../resources_test/pattern.png) no-repeat 6px 100%", `
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			________rBBB__
			________BBBB__
			________BBBB__
			________BBBB__
			______________
			______________
			______________
			______________
		`},

		{"repeat_x_1px_2px", "url(../resources_test/pattern.png) repeat-x 1px 2px", `
			______________
			______________
			______________
			______________
			__BrBBBrBBBr__
			__BBBBBBBBBB__
			__BBBBBBBBBB__
			__BBBBBBBBBB__
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
		`},
		{"repeat_y_local_2px_1px", "url(../resources_test/pattern.png) repeat-y local 2px 1px", `
			______________
			______________
			____BBBB______
			____rBBB______
			____BBBB______
			____BBBB______
			____BBBB______
			____rBBB______
			____BBBB______
			____BBBB______
			____BBBB______
			____rBBB______
			______________
			______________
			______________
			______________
		`},

		{"fixed", "url(../resources_test/pattern.png) no-repeat fixed", `
			# The image is actually here:
			#######
			______________
			______________
			__BB__________
			__BB__________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
		`},
		{"fixed_right", "url(../resources_test/pattern.png) no-repeat fixed right 3px", `
			#                   x x x x
			______________
			______________
			______________
			__________rB__   #
			__________BB__   #
			__________BB__   #
			__________BB__   #
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
		`},
		{"fixed_center_center", "url(../resources_test/pattern.png)no-repeat fixed 50%center", `
			______________
			______________
			______________
			______________
			______________
			______________
			_____rBBB_____
			_____BBBB_____
			_____BBBB_____
			_____BBBB_____
			______________
			______________
			______________
			______________
			______________
			______________
		`},
		{"multi_under", `url(../resources_test/pattern.png) no-repeat,
						   url(../resources_test/pattern.png) no-repeat 2px 1px`, `
			______________
			______________
			__rBBB________
			__BBBBBB______
			__BBBBBB______
			__BBBBBB______
			____BBBB______
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
		`},
		{"multi_over", `url(../resources_test/pattern.png) no-repeat 2px 1px,
						  url(../resources_test/pattern.png) no-repeat`, `
			______________
			______________
			__rBBB________
			__BBrBBB______
			__BBBBBB______
			__BBBBBB______
			____BBBB______
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
			______________
		`},
	} {
		testBackgroundImage(t, data[0], data[1], data[2])
	}
}

func testBackgroundImage(t *testing.T, name, css, pixels string) {
	// pattern.png looks like this:
	//
	//    rBBB
	//    BBBB
	//    BBBB
	//    BBBB

	assertPixelsEqual(t, pixels, fmt.Sprintf(`
      <style>
        @page { size: 14px 16px }
        html { background: #fff }
        body { margin: 2px; height: 10px;
               background: %s }
        p { background: none }
      </style>
      <body>
      <p>&nbsp;`, css))
}

func TestBackgroundImageZeroSizeBackground(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	// Regression test for https://github.com/Kozea/WeasyPrint/issues/217
	assertPixelsEqual(t, `
        __________
        __________
        __________
        __________
        __________
        __________
        __________
        __________
        __________
        __________
    `, `
      <style>
        @page { size: 10px }
        html { background: #fff }
        body { background: url(../resources_test/pattern.png);
               background-size: cover;
               display: inline-block }
      </style>
      <body>`)
}

// Test the background-origin property.
func TestBackgroundOrigin(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	testValue := func(value, pixels, css string) {
		if css == "" {
			css = value
		}
		assertPixelsEqual(t, pixels, fmt.Sprintf(`
            <style>
                @page { size: 12px }
                html { background: #fff }
                body { margin: 1px; padding: 1px; height: 6px;
                       border: 1px solid  transparent;
                       background: url(../resources_test/pattern.png) bottom right no-repeat;
                       background-origin: %s }
            </style>
            <body>`, css))
	}

	testValue("border-box", `
	        ____________
	        ____________
	        ____________
	        ____________
	        ____________
	        ____________
	        ____________
	        _______rBBB_
	        _______BBBB_
	        _______BBBB_
	        _______BBBB_
	        ____________
	    `, "")
	testValue("padding-box", `
	        ____________
	        ____________
	        ____________
	        ____________
	        ____________
	        ____________
	        ______rBBB__
	        ______BBBB__
	        ______BBBB__
	        ______BBBB__
	        ____________
	        ____________
	    `, "")
	testValue("content-box", `
	        ____________
	        ____________
	        ____________
	        ____________
	        ____________
	        _____rBBB___
	        _____BBBB___
	        _____BBBB___
	        _____BBBB___
	        ____________
	        ____________
	        ____________
	    `, "")

	testValue("border-box_clip", `
	        ____________
	        ____________
	        ____________
	        ____________
	        ____________
	        ____________
	        ____________
	        _______rB___
	        _______BB___
	        ____________
	        ____________
	        ____________
	    `, "border-box; background-clip: content-box")
}

func TestBackgroundRepeatSpace_1(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        ____________
        _rBBB__rBBB_
        _BBBB__BBBB_
        _BBBB__BBBB_
        _BBBB__BBBB_
        ____________
        _rBBB__rBBB_
        _BBBB__BBBB_
        _BBBB__BBBB_
        _BBBB__BBBB_
        ____________
        _rBBB__rBBB_
        _BBBB__BBBB_
        _BBBB__BBBB_
        _BBBB__BBBB_
        ____________
    `, `
      <style>
        @page { size: 12px 16px }
        html { background: #fff }
        body { margin: 1px; height: 14px;
               background: url(../resources_test/pattern.png) space; }
      </style>
      <body>`)
}

func TestBackgroundRepeatSpace_2(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        ____________
        _rBBB__rBBB_
        _BBBB__BBBB_
        _BBBB__BBBB_
        _BBBB__BBBB_
        _rBBB__rBBB_
        _BBBB__BBBB_
        _BBBB__BBBB_
        _BBBB__BBBB_
        _rBBB__rBBB_
        _BBBB__BBBB_
        _BBBB__BBBB_
        _BBBB__BBBB_
        ____________
    `, `
      <style>
        @page { size: 12px 14px }
        html { background: #fff }
        body { margin: 1px; height: 12px;
               background: url(../resources_test/pattern.png) space; }
      </style>
      <body>`)
}

func TestBackgroundRepeatSpace_3(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        ____________
        _rBBBrBBBrB_
        _BBBBBBBBBB_
        _BBBBBBBBBB_
        _BBBBBBBBBB_
        ____________
        ____________
        ____________
        _rBBBrBBBrB_
        _BBBBBBBBBB_
        _BBBBBBBBBB_
        _BBBBBBBBBB_
        ____________
    `, `
      <style>
        @page { size: 12px 13px }
        html { background: #fff }
        body { margin: 1px; height: 11px;
               background: url(../resources_test/pattern.png) repeat space; }
      </style>
      <body>`)
}

func TestBackgroundRepeatRound_1(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        __________
        _rrBBBBBB_
        _rrBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _rrBBBBBB_
        _rrBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        __________
    `, `
      <style>
        @page { size: 10px 14px }
        html { background: #fff }
        body { margin: 1px; height: 12px;
               image-rendering: pixelated;
               background: url(../resources_test/pattern.png) top/6px round repeat; }
      </style>
      <body>`)
}

func TestBackgroundRepeatRound_2(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        __________
        _rrBBBBBB_
        _rrBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _rrBBBBBB_
        _rrBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        __________
    `, `
      <style>
        @page { size: 10px 18px }
        html { background: #fff }
        body { margin: 1px; height: 16px;
               image-rendering: pixelated;
               background: url(../resources_test/pattern.png) center/auto 8px repeat round; }
      </style>
      <body>`)
}

func TestBackgroundRepeatRound_3(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        __________
        _rrBBBBBB_
        _rrBBBBBB_
        _rrBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        __________
    `, `
      <style>
        @page { size: 10px 14px }
        html { background: #fff }
        body { margin: 1px; height: 12px;
               image-rendering: pixelated;
               background: url(../resources_test/pattern.png) center/6px 9px round; }
      </style>
      <body>`)
}

func TestBackgroundRepeatRound_4(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        __________
        _rBBBrBBB_
        _rBBBrBBB_
        _rBBBrBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        __________
    `, `
      <style>
        @page { size: 10px 14px }
        html { background: #fff }
        body { margin: 1px; height: 12px;
               image-rendering: pixelated;
               background: url(../resources_test/pattern.png) center/5px 9px round; }
      </style>
      <body>`)
}

func TestBackgroundClip(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, data := range [][2]string{
		{"#00f border-box", `
			________
			_BBBBBB_
			_BBBBBB_
			_BBBBBB_
			_BBBBBB_
			_BBBBBB_
			_BBBBBB_
			________
		`},
		{"#00f padding-box", `
			________
			________
			__BBBB__
			__BBBB__
			__BBBB__
			__BBBB__
			________
			________
		`},
		{"#00f content-box", `
			________
			________
			________
			___BB___
			___BB___
			________
			________
			________
		`},
		{"url(../resources_test/pattern.png) padding-box, #0f0", `
			________
			_GGGGGG_
			_GrBBBG_
			_GBBBBG_
			_GBBBBG_
			_GBBBBG_
			_GGGGGG_
			________
		`},
	} {
		value, pixels := data[0], data[1]
		assertPixelsEqual(t, pixels, fmt.Sprintf(`
		      <style>
		        @page { size: 8px }
		        html { background: #fff }
		        body { margin: 1px; padding: 1px; height: 2px;
		               border: 1px solid  transparent;
		               background: %s }
		      </style>
		      <body>`, value))
	}
}

func TestBackgroundSize(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, data := range [][3]string{
		{"background_size", `
	             ____________
	             ____________
	             ____________
	             ___rrBBBBBB_
	             ___rrBBBBBB_
	             ___BBBBBBBB_
	             ___BBBBBBBB_
	             ___BBBBBBBB_
	             ___BBBBBBBB_
	             ___BBBBBBBB_
	             ___BBBBBBBB_
	             ____________
	         `, `
	           <style>
	             @page { size: 12px }
	             html { background: #fff }
	             body { margin: 1px; height: 10px;
	                    /* Use nearest neighbor algorithm for image resizing: */
	                    image-rendering: pixelated;
	                    background: url(../resources_test/pattern.png) no-repeat
	                                bottom right / 80% 8px; }
	           </style>
	           <body>`},
		{"background_size_auto", `
	             ____________
	             ____________
	             ____________
	             ____________
	             ____________
	             ____________
	             ____________
	             _______rBBB_
	             _______BBBB_
	             _______BBBB_
	             _______BBBB_
	             ____________
	         `, `
	           <style>
	             @page { size: 12px }
	             html { background: #fff }
	             body { margin: 1px; height: 10px;
	                    /* Use nearest neighbor algorithm for image resizing: */
	                    image-rendering: pixelated;
	                    background: url(../resources_test/pattern.png) bottom right/auto no-repeat }
	           </style>
	           <body>`},
		{"background_size_contain", `
	             ______________
	             _rrBBBBBB_____
	             _rrBBBBBB_____
	             _BBBBBBBB_____
	             _BBBBBBBB_____
	             _BBBBBBBB_____
	             _BBBBBBBB_____
	             _BBBBBBBB_____
	             _BBBBBBBB_____
	             ______________
	         `, `
	           <style>
	             @page { size: 14px 10px }
	             html { background: #fff }
	             body { margin: 1px; height: 8px;
	                    /* Use nearest neighbor algorithm for image resizing: */
	                    image-rendering: pixelated;
	                    background: url(../resources_test/pattern.png) no-repeat;
	                    background-size: contain }
	           </style>
	           <body>`},

		{"background_size_mixed", `
	             ______________
	             _rrBBBBBB_____
	             _rrBBBBBB_____
	             _BBBBBBBB_____
	             _BBBBBBBB_____
	             _BBBBBBBB_____
	             _BBBBBBBB_____
	             _BBBBBBBB_____
	             _BBBBBBBB_____
	             ______________
	         `, `
	           <style>
	             @page { size: 14px 10px }
	             html { background: #fff }
	             body { margin: 1px; height: 8px;
	                    /* Use nearest neighbor algorithm for image resizing: */
	                    image-rendering: pixelated;
	                    background: url(../resources_test/pattern.png) no-repeat left / auto 8px;
	                    clip: auto; /* no-op to cover more validation */ }
	           </style>
	           <body>`},
		{"background_size_double", `
	             ______________
	             _rrBBBBBB_____
	             _BBBBBBBB_____
	             _BBBBBBBB_____
	             _BBBBBBBB_____
	             ______________
	             ______________
	             ______________
	             ______________
	             ______________
	         `, `
	           <style>
	             @page { size: 14px 10px }
	             html { background: #fff }
	             body { margin: 1px; height: 8px;
	                    /* Use nearest neighbor algorithm for image resizing: */
	                    image-rendering: pixelated;
	                    background: url(../resources_test/pattern.png) no-repeat 0 0 / 8px 4px;
	                    clip: auto; /* no-op to cover more validation */ }
	           </style>
	           <body>`},
		{"background_size_cover", `
	             ______________
	             _rrrBBBBBBBBB_
	             _rrrBBBBBBBBB_
	             _rrrBBBBBBBBB_
	             _BBBBBBBBBBBB_
	             _BBBBBBBBBBBB_
	             _BBBBBBBBBBBB_
	             _BBBBBBBBBBBB_
	             _BBBBBBBBBBBB_
	             ______________
	         `, `
	           <style>
	             @page { size: 14px 10px }
	             html { background: #fff }
	             body { margin: 1px; height: 8px;
	                    /* Use nearest neighbor algorithm for image resizing: */
	                    image-rendering: pixelated;
	                    background: url(../resources_test/pattern.png) no-repeat right 0/cover }
	           </style>
	           <body>`},
	} {
		assertPixelsEqual(t, data[1], data[2])
	}
}

func TestBleedBackgroundSize(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	expectedPixels := `
        RRRR
        RRRR
        RRRR
        RRRR
    `
	html := `
      <style>
         @page { size: 2px; background: red; bleed: 1px }
      </style>
      <body>`
	assertPixelsEqual(t, expectedPixels, html)
}

func TestBackgroundSizeClip(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        BBBB
        BRBB
        BBBB
        BBBB
    `, `
      <style>
         @page { size: 4px; margin: 1px;
                 background: url(pattern.png) red;
                 background-clip: content-box }
      </style>
      <body>`)
}

func TestPageBackgroundFixed(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1993
	assertPixelsEqual(t, `
        RBBB
        BBBB
        BBBB
        BBBB
    `, `
      <style>
         @page { size: 4px; margin: 1px;
                 background: url(pattern.png) red;
                 background-attachment: fixed; }
      </style>
      <body>`)
}

func TestPageBackgroundFixedBleed(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1993
	assertPixelsEqual(t, `
        RRRRRR
        RRBBBR
        RBBBBR
        RBBBBR
        RBBBBR
        RRRRRR
    `, `
      <style>
         @page { size: 4px; margin: 1px; bleed: 1px;
                 background: url(pattern.png) no-repeat red;
                 background-attachment: fixed; }
      </style>
      <body>`)
}

func TestBleedBackgroundSizeClip(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1943
	assertPixelsEqual(t, `
        BBBBBB
        BBBBBB
        BBRBBB
        BBBBBB
        BBBBBB
        BBBBBB
    `, `
      <style>
         @page { size: 4px; bleed: 1px; margin: 1px;
                 background: url(pattern.png) red;
                 background-clip: content-box }
      </style>
      <body>`)
}

func TestMarksCrop(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        KK__KK
        K____K
        ______
        ______
        K____K
        KK__KK
    `, `
      <style>
         @page { size: 4px; bleed: 1px; margin: 1px; marks: crop }
      </style>
      <body>`)
}

func TestMarksCross(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        __KK__
        ______
        K____K
        K____K
        ______
        __KK__
    `, `
      <style>
         @page { size: 4px; bleed: 1px; margin: 1px; marks: cross }
      </style>
      <body>`)
}

func TestMarksCropCross(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        KKKKKK
        K____K
        K____K
        K____K
        K____K
        KKKKKK
    `, `
      <style>
         @page { size: 4px; bleed: 1px; margin: 1px; marks: crop cross }
      </style>
      <body>`)
}
