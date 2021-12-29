package pdf

import (
	"testing"

	"github.com/benoitkugler/webrender/utils/testutils"
)

// Test how floats are drawn.

func TestFloat(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "float", `
        rBBB__aaaa
        BBBB__aaaa
        BBBB__aaaa
        BBBB__aaaa
        __________
    `, `
      <style>
        @page { size: 10px 5px; background: white }
      </style>
      <div>
        <img style="float: left" src="../resources_test/pattern.png">
        <img style="float: right" src="../resources_test/blue.jpg">
      </div>
    `)
}

func TestFloatRtl(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "float_rtl", `
        rBBB__aaaa
        BBBB__aaaa
        BBBB__aaaa
        BBBB__aaaa
        __________
    `, `
      <style>
        @page { size: 10px 5px; background: white }
      </style>
      <div style="direction: rtl">
        <img style="float: left" src="../resources_test/pattern.png">
        <img style="float: right" src="../resources_test/blue.jpg">
      </div>
    `)
}

func TestFloatInline(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "float_inline", `
        rBBBGG_____aaaa
        BBBBGG_____aaaa
        BBBB_______aaaa
        BBBB_______aaaa
        _______________
    `, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page { size: 15px 5px; background: white }
        body { font-family: weasyprint; font-size: 2px; line-height: 1;
               color: lime }
      </style>
      <div>
        <img style="float: left" src="../resources_test/pattern.png">
        <img style="float: right" src="../resources_test/blue.jpg">
        <span>a</span>
      </div>
    `)
}

func TestFloatInlineRtl(t *testing.T) {
	// capt := testutils.CaptureLogs()
	// defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "float_inline_rtl", `
        rBBB_____GGaaaa
        BBBB_____GGaaaa
        BBBB_______aaaa
        BBBB_______aaaa
        _______________
    `, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page { size: 15px 5px; background: white }
        body { font-family: weasyprint; font-size: 2px; line-height: 1;
               color: lime }
      </style>
      <div style="direction: rtl">
        <img style="float: left" src="../resources_test/pattern.png">
        <img style="float: right" src="../resources_test/blue.jpg">
        <span>a</span>
      </div>
    `)
}

func TestFloatInlineBlock(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "float_inline_block", `
        rBBBGG_____aaaa
        BBBBGG_____aaaa
        BBBB_______aaaa
        BBBB_______aaaa
        _______________
    `, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page { size: 15px 5px; background: white }
        body { font-family: weasyprint; font-size: 2px; line-height: 1;
               color: lime }
      </style>
      <div>
        <img style="float: left" src="../resources_test/pattern.png">
        <img style="float: right" src="../resources_test/blue.jpg">
        <span style="display: inline-block">a</span>
      </div>
    `)
}

func TestFloatInlineBlockRtl(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "float_inline_block_rtl", `
        rBBB_____GGaaaa
        BBBB_____GGaaaa
        BBBB_______aaaa
        BBBB_______aaaa
        _______________
    `, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page { size: 15px 5px; background: white }
        body { font-family: weasyprint; font-size: 2px; line-height: 1;
               color: lime }
      </style>
      <div style="direction: rtl">
        <img style="float: left" src="../resources_test/pattern.png">
        <img style="float: right" src="../resources_test/blue.jpg">
        <span style="display: inline-block">a</span>
      </div>
    `)
}

func TestFloatTable(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "float_table", `
        rBBBGG_____aaaa
        BBBBGG_____aaaa
        BBBB_______aaaa
        BBBB_______aaaa
        _______________
    `, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page { size: 15px 5px; background: white }
        body { font-family: weasyprint; font-size: 2px; line-height: 1;
               color: lime }
      </style>
      <div>
        <img style="float: left" src="../resources_test/pattern.png">
        <img style="float: right" src="../resources_test/blue.jpg">
        <table><tbody><tr><td>a</td></tr></tbody></table>
      </div>
    `)
}

func TestFloatTableRtl(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "float_table_rtl", `
        rBBB_____GGaaaa
        BBBB_____GGaaaa
        BBBB_______aaaa
        BBBB_______aaaa
        _______________
    `, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page { size: 15px 5px; background: white }
        body { font-family: weasyprint; font-size: 2px; line-height: 1;
               color: lime }
      </style>
      <div style="direction: rtl">
        <img style="float: left" src="../resources_test/pattern.png">
        <img style="float: right" src="../resources_test/blue.jpg">
        <table><tbody><tr><td>a</td></tr></tbody></table>
      </div>
    `)
}

func TestFloatInlineTable(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "float_inline_table", `
        rBBBGG_____aaaa
        BBBBGG_____aaaa
        BBBB_______aaaa
        BBBB_______aaaa
        _______________
    `, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page { size: 15px 5px; background: white }
        table { display: inline-table }
        body { font-family: weasyprint; font-size: 2px; line-height: 1;
               color: lime }
      </style>
      <div>
        <img style="float: left" src="../resources_test/pattern.png">
        <img style="float: right" src="../resources_test/blue.jpg">
        <table><tbody><tr><td>a</td></tr></tbody></table>
      </div>
    `)
}

func TestFloatInlineTableRtl(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "float_inline_table_rtl", `
        rBBB_____GGaaaa
        BBBB_____GGaaaa
        BBBB_______aaaa
        BBBB_______aaaa
        _______________
    `, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page { size: 15px 5px; background: white }
        table { display: inline-table }
        body { font-family: weasyprint; font-size: 2px; line-height: 1;
               color: lime }
      </style>
      <div style="direction: rtl">
        <img style="float: left" src="../resources_test/pattern.png">
        <img style="float: right" src="../resources_test/blue.jpg">
        <table><tbody><tr><td>a</td></tr></tbody></table>
      </div>
    `)
}

func TestFloatReplacedBlock(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "float_replaced_block", `
        rBBBaaaa___rBBB
        BBBBaaaa___BBBB
        BBBBaaaa___BBBB
        BBBBaaaa___BBBB
        _______________
    `, `
      <style>
        @page { size: 15px 5px; background: white }
      </style>
      <div>
        <img style="float: left" src="../resources_test/pattern.png">
        <img style="float: right" src="../resources_test/pattern.png">
        <img style="display: block" src="../resources_test/blue.jpg">
      </div>
    `)
}

func TestFloatReplacedBlockRtl(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "float_replaced_block_rtl", `
        rBBB___aaaarBBB
        BBBB___aaaaBBBB
        BBBB___aaaaBBBB
        BBBB___aaaaBBBB
        _______________
    `, `
      <style>
        @page { size: 15px 5px; background: white }
      </style>
      <div style="direction: rtl">
        <img style="float: left" src="../resources_test/pattern.png">
        <img style="float: right" src="../resources_test/pattern.png">
        <img style="display: block" src="../resources_test/blue.jpg">
      </div>
    `)
}

// @pytest.mark.xfail
// func TestFloatReplacedInline(t *testing.T){
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     assertPixelsEqual(t, "float_replaced_inline", 15, 5, `
//         rBBBaaaa___rBBB
//         BBBBaaaa___BBBB
//         BBBBaaaa___BBBB
//         BBBBaaaa___BBBB
//         _______________
//     `, `
//       <style>
//         @page { size: 15px 5px; background: white }
//         body { line-height: 1px }
//       </style>
//       <div>
//         <img style="float: left" src="../resources_test/pattern.png">
//         <img style="float: right" src="../resources_test/pattern.png">
//         <img src="../resources_test/blue.jpg">
//       </div>
//     `)

// @pytest.mark.xfail
// func TestFloatReplacedInlineRtl(t *testing.T){
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     assertPixelsEqual(t, "float_replaced_inline_rtl", 15, 5, `
//         rBBB___aaaarBBB
//         BBBB___aaaaBBBB
//         BBBB___aaaaBBBB
//         BBBB___aaaaBBBB
//         _______________
//     `, `
//       <style>
//         @page { size: 15px 5px; background: white }
//         body { line-height: 1px }
//       </style>
//       <div style="direction: rtl">
//         <img style="float: left" src="../resources_test/pattern.png">
//         <img style="float: right" src="../resources_test/pattern.png">
//         <img src="../resources_test/blue.jpg">
//       </div>
//     `)
