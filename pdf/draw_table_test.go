package pdf

import (
	"fmt"
	"image/color"
	"testing"

	"github.com/benoitkugler/go-weasyprint/boxes"
	"github.com/benoitkugler/go-weasyprint/utils/testutils"
)

// Test how tables are drawn.

func toPix(pixels_str string) [][]color.RGBA {
	return parsePixelsExt(pixels_str, map[byte]color.RGBA{
		// rgba(255, 0, 0, 0.5) above #fff
		'r': {255, 127, 127, 255},
		// rgba(0, 255, 0, 0.5) above #fff
		'g': {127, 255, 127, 255},
		// r above B above #fff.
		'b': {128, 0, 127, 255},
	})
}

func init() {
	boxes.HTMLHandlers["x-td"] = boxes.HTMLHandlers["td"]
	boxes.HTMLHandlers["x-th"] = boxes.HTMLHandlers["th"]
}

const tables_source = `
  <style>
    @page { size: 28px; background: #fff }
    x-table { margin: 1px; padding: 1px; border-spacing: 1px;
              border: 1px solid transparent }
    x-td { width: 2px; height: 2px; padding: 1px;
           border: 1px solid transparent }
    %s
  </style>
  <x-table>
    <x-colgroup>
      <x-col></x-col>
      <x-col></x-col>
    </x-colgroup>
    <x-col></x-col>
    <x-tbody>
      <x-tr>
        <x-td></x-td>
        <x-td rowspan=2></x-td>
        <x-td></x-td>
      </x-tr>
      <x-tr>
        <x-td colspan=2></x-td>
        <x-td></x-td>
      </x-tr>
    </x-tbody>
    <x-tr>
      <x-td></x-td>
      <x-td></x-td>
    </x-tr>
  </x-table>
`

func TestTables_1(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqualFromPixels(t, "table_borders", toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
        _B__rrrrrr_rrrrrr_rrrrrr__B_
        _B__r____r_r____r_r____r__B_
        _B__r____r_r____r_r____r__B_
        _B__r____r_r____r_r____r__B_
        _B__r____r_r____r_r____r__B_
        _B__rrrrrr_r____r_rrrrrr__B_
        _B_________r____r_________B_
        _B__rrrrrrrSrrrrS_rrrrrr__B_
        _B__r______r____S_r____r__B_
        _B__r______r____S_r____r__B_
        _B__r______r____S_r____r__B_
        _B__r______r____S_r____r__B_
        _B__rrrrrrrSSSSSS_rrrrrr__B_
        _B________________________B_
        _B__rrrrrr_rrrrrr_________B_
        _B__r____r_r____r_________B_
        _B__r____r_r____r_________B_
        _B__r____r_r____r_________B_
        _B__r____r_r____r_________B_
        _B__rrrrrr_rrrrrr_________B_
        _B________________________B_
        _B________________________B_
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        ____________________________
    `), fmt.Sprintf(tables_source, `
      x-table { border-color: #00f; table-layout: fixed }
      x-td { border-color: rgba(255, 0, 0, 0.5) }
    `))
}

func TestTables_1Rtl(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqualFromPixels(t, "table_borders_rtl", toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
        _B__rrrrrr_rrrrrr_rrrrrr__B_
        _B__r____r_r____r_r____r__B_
        _B__r____r_r____r_r____r__B_
        _B__r____r_r____r_r____r__B_
        _B__r____r_r____r_r____r__B_
        _B__rrrrrr_r____r_rrrrrr__B_
        _B_________r____r_________B_
        _B__rrrrrr_SrrrrSrrrrrrr__B_
        _B__r____r_S____r______r__B_
        _B__r____r_S____r______r__B_
        _B__r____r_S____r______r__B_
        _B__r____r_S____r______r__B_
        _B__rrrrrr_SSSSSSrrrrrrr__B_
        _B________________________B_
        _B_________rrrrrr_rrrrrr__B_
        _B_________r____r_r____r__B_
        _B_________r____r_r____r__B_
        _B_________r____r_r____r__B_
        _B_________r____r_r____r__B_
        _B_________rrrrrr_rrrrrr__B_
        _B________________________B_
        _B________________________B_
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        ____________________________
    `), fmt.Sprintf(tables_source, `
      x-table { border-color: #00f; table-layout: fixed;
                direction: rtl; }
      x-td { border-color: rgba(255, 0, 0, 0.5) }
    `))
}

func TestTables_2(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqualFromPixels(t, "table_collapsed_borders", toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBB_________
        _BBBBBBBBBBBBBBBBBB_________
        _BB____r____r____BB_________
        _BB____r____r____BB_________
        _BB____r____r____BB_________
        _BB____r____r____BB_________
        _BBrrrrr____rrrrrBB_________
        _BB_________r____BB_________
        _BB_________r____BB_________
        _BB_________r____BB_________
        _BB_________r____BB_________
        _BBrrrrrrrrrrrrrrBB_________
        _BB____r____r____BB_________
        _BB____r____r____BB_________
        _BB____r____r____BB_________
        _BB____r____r____BB_________
        _BBBBBBBBBBBBBBBBBB_________
        _BBBBBBBBBBBBBBBBBB_________
        ____________________________
        ____________________________
        ____________________________
        ____________________________
        ____________________________
        ____________________________
        ____________________________
        ____________________________
        ____________________________
    `), fmt.Sprintf(tables_source, `
      x-table { border: 2px solid #00f; table-layout: fixed;
                border-collapse: collapse }
      x-td { border-color: #ff7f7f }
    `))
}

func TestTables_2Rtl(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqualFromPixels(t, "table_collapsed_borders_rtl", toPix(`
        ____________________________
        _________BBBBBBBBBBBBBBBBBB_
        _________BBBBBBBBBBBBBBBBBB_
        _________BB____r____r____BB_
        _________BB____r____r____BB_
        _________BB____r____r____BB_
        _________BB____r____r____BB_
        _________BBrrrrr____rrrrrBB_
        _________BB____r_________BB_
        _________BB____r_________BB_
        _________BB____r_________BB_
        _________BB____r_________BB_
        _________BBrrrrrrrrrrrrrrBB_
        _________BB____r____r____BB_
        _________BB____r____r____BB_
        _________BB____r____r____BB_
        _________BB____r____r____BB_
        _________BBBBBBBBBBBBBBBBBB_
        _________BBBBBBBBBBBBBBBBBB_
        ____________________________
        ____________________________
        ____________________________
        ____________________________
        ____________________________
        ____________________________
        ____________________________
        ____________________________
        ____________________________
    `), fmt.Sprintf(tables_source, `
      body { direction: rtl; }
      x-table { border: 2px solid #00f; table-layout: fixed;
                border-collapse: collapse; }
      x-td { border-color: #ff7f7f }
    `))
}

func TestTables_3(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqualFromPixels(t, "table_collapsed_borders_paged", toPix(`
        ____________________________
        _gggggggggggggggggggggggggg_
        _g________________________g_
        _g_BBBBBBBBBBBBBBBBBB_____g_
        _g_BBBBBBBBBBBBBBBBBB_____g_
        _g_BBBBBBBBBBBBBBBBBB_____g_
        _g_BBBBBBBBBBBBBBBBBB_____g_
        _g_BBBBBBBBBBBBBBBBBB_____g_
        _g_BBBBBBBBBBBBBBBBBB_____g_
        _g_BBBBBBBBBBBBBBBBBB_____g_
        _g_BBBBBBBBBBBBBBBBBB_____g_
        _g_BB____r____r____BB_____g_
        _g_BB____r____r____BB_____g_
        _g_BB____r____r____BB_____g_
        _g_BB____r____r____BB_____g_
        _g_BBrrrrr____rrrrrBB_____g_
        _g_BB_________r____BB_____g_
        _g_BB_________r____BB_____g_
        _g_BB_________r____BB_____g_
        _g_BB_________r____BB_____g_
        _g_BBrrrrrrrrrrrrrrBB_____g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _gggggggggggggggggggggggggg_
        ____________________________
        ____________________________
        _gggggggggggggggggggggggggg_
        _g_BBrrrrrrrrrrrrrrBB_____g_
        _g_BB____r____r____BB_____g_
        _g_BB____r____r____BB_____g_
        _g_BB____r____r____BB_____g_
        _g_BB____r____r____BB_____g_
        _g_BBBBBBBBBBBBBBBBBB_____g_
        _g_BBBBBBBBBBBBBBBBBB_____g_
        _g_BBBBBBBBBBBBBBBBBB_____g_
        _g_BBBBBBBBBBBBBBBBBB_____g_
        _g_BBBBBBBBBBBBBBBBBB_____g_
        _g_BBBBBBBBBBBBBBBBBB_____g_
        _g_BBBBBBBBBBBBBBBBBB_____g_
        _g_BBBBBBBBBBBBBBBBBB_____g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _gggggggggggggggggggggggggg_
        ____________________________
    `), fmt.Sprintf(tables_source, `
      x-table { border: solid #00f; border-width: 8px 2px;
                table-layout: fixed; border-collapse: collapse }
      x-td { border-color: #ff7f7f }
      @page { size: 28px 26px; margin: 1px;
              border: 1px solid rgba(0, 255, 0, 0.5); }
    `))
}

func TestTables_3Rtl(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqualFromPixels(t, "table_collapsed_borders_paged_rtl", toPix(`
        ____________________________
        _gggggggggggggggggggggggggg_
        _g________________________g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BB____r____r____BB_g_
        _g_____BB____r____r____BB_g_
        _g_____BB____r____r____BB_g_
        _g_____BB____r____r____BB_g_
        _g_____BBrrrrr____rrrrrBB_g_
        _g_____BB____r_________BB_g_
        _g_____BB____r_________BB_g_
        _g_____BB____r_________BB_g_
        _g_____BB____r_________BB_g_
        _g_____BBrrrrrrrrrrrrrrBB_g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _gggggggggggggggggggggggggg_
        ____________________________
        ____________________________
        _gggggggggggggggggggggggggg_
        _g_____BBrrrrrrrrrrrrrrBB_g_
        _g_____BB____r____r____BB_g_
        _g_____BB____r____r____BB_g_
        _g_____BB____r____r____BB_g_
        _g_____BB____r____r____BB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _gggggggggggggggggggggggggg_
        ____________________________
    `), fmt.Sprintf(tables_source, `
      body { direction: rtl; }
      x-table { border: solid #00f; border-width: 8px 2px;
                table-layout: fixed; border-collapse: collapse; }
      x-td { border-color: #ff7f7f }
      @page { size: 28px 26px; margin: 1px;
              border: 1px solid rgba(0, 255, 0, 0.5); }
    `))
}

func TestTables_4(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqualFromPixels(t, "table_td_backgrounds", toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
        _B__rrrrrr_rrrrrr_rrrrrr__B_
        _B__rrrrrr_rrrrrr_rrrrrr__B_
        _B__rrrrrr_rrrrrr_rrrrrr__B_
        _B__rrrrrr_rrrrrr_rrrrrr__B_
        _B__rrrrrr_rrrrrr_rrrrrr__B_
        _B__rrrrrr_rrrrrr_rrrrrr__B_
        _B_________rrrrrr_________B_
        _B__rrrrrrrSSSSSS_rrrrrr__B_
        _B__rrrrrrrSSSSSS_rrrrrr__B_
        _B__rrrrrrrSSSSSS_rrrrrr__B_
        _B__rrrrrrrSSSSSS_rrrrrr__B_
        _B__rrrrrrrSSSSSS_rrrrrr__B_
        _B__rrrrrrrSSSSSS_rrrrrr__B_
        _B________________________B_
        _B__rrrrrr_rrrrrr_________B_
        _B__rrrrrr_rrrrrr_________B_
        _B__rrrrrr_rrrrrr_________B_
        _B__rrrrrr_rrrrrr_________B_
        _B__rrrrrr_rrrrrr_________B_
        _B__rrrrrr_rrrrrr_________B_
        _B________________________B_
        _B________________________B_
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        ____________________________
    `), fmt.Sprintf(tables_source, `
      x-table { border-color: #00f; table-layout: fixed }
      x-td { background: rgba(255, 0, 0, 0.5) }
    `))
}

func TestTables_4Rtl(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqualFromPixels(t, "table_td_backgrounds_rtl", toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
        _B__rrrrrr_rrrrrr_rrrrrr__B_
        _B__rrrrrr_rrrrrr_rrrrrr__B_
        _B__rrrrrr_rrrrrr_rrrrrr__B_
        _B__rrrrrr_rrrrrr_rrrrrr__B_
        _B__rrrrrr_rrrrrr_rrrrrr__B_
        _B__rrrrrr_rrrrrr_rrrrrr__B_
        _B_________rrrrrr_________B_
        _B__rrrrrr_SSSSSSrrrrrrr__B_
        _B__rrrrrr_SSSSSSrrrrrrr__B_
        _B__rrrrrr_SSSSSSrrrrrrr__B_
        _B__rrrrrr_SSSSSSrrrrrrr__B_
        _B__rrrrrr_SSSSSSrrrrrrr__B_
        _B__rrrrrr_SSSSSSrrrrrrr__B_
        _B________________________B_
        _B_________rrrrrr_rrrrrr__B_
        _B_________rrrrrr_rrrrrr__B_
        _B_________rrrrrr_rrrrrr__B_
        _B_________rrrrrr_rrrrrr__B_
        _B_________rrrrrr_rrrrrr__B_
        _B_________rrrrrr_rrrrrr__B_
        _B________________________B_
        _B________________________B_
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        ____________________________
    `), fmt.Sprintf(tables_source, `
      x-table { border-color: #00f; table-layout: fixed;
                direction: rtl; }
      x-td { background: rgba(255, 0, 0, 0.5) }
    `))
}

func TestTables_5(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqualFromPixels(t, "table_row_backgrounds", toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
        _B__bbbbbb_bbbbbb_bbbbbb__B_
        _B__bbbbbb_bbbbbb_bbbbbb__B_
        _B__bbbbbb_bbbbbb_bbbbbb__B_
        _B__bbbbbb_bbbbbb_bbbbbb__B_
        _B__bbbbbb_bbbbbb_bbbbbb__B_
        _B__bbbbbb_bbbbbb_bbbbbb__B_
        _B_________bbbbbb_________B_
        _B__bbbbbbbpppppp_bbbbbb__B_
        _B__bbbbbbbpppppp_bbbbbb__B_
        _B__bbbbbbbpppppp_bbbbbb__B_
        _B__bbbbbbbpppppp_bbbbbb__B_
        _B__bbbbbbbpppppp_bbbbbb__B_
        _B__bbbbbbbpppppp_bbbbbb__B_
        _B________________________B_
        _B__rrrrrr_rrrrrr_________B_
        _B__rrrrrr_rrrrrr_________B_
        _B__rrrrrr_rrrrrr_________B_
        _B__rrrrrr_rrrrrr_________B_
        _B__rrrrrr_rrrrrr_________B_
        _B__rrrrrr_rrrrrr_________B_
        _B________________________B_
        _B________________________B_
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        ____________________________
    `), fmt.Sprintf(tables_source, `
      x-table { border-color: #00f; table-layout: fixed }
      x-tbody { background: rgba(0, 0, 255, 1) }
      x-tr { background: rgba(255, 0, 0, 0.5) }
    `))
}

func TestTables_5Rtl(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqualFromPixels(t, "table_row_backgrounds_rtl", toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
        _B__bbbbbb_bbbbbb_bbbbbb__B_
        _B__bbbbbb_bbbbbb_bbbbbb__B_
        _B__bbbbbb_bbbbbb_bbbbbb__B_
        _B__bbbbbb_bbbbbb_bbbbbb__B_
        _B__bbbbbb_bbbbbb_bbbbbb__B_
        _B__bbbbbb_bbbbbb_bbbbbb__B_
        _B_________bbbbbb_________B_
        _B__bbbbbb_ppppppbbbbbbb__B_
        _B__bbbbbb_ppppppbbbbbbb__B_
        _B__bbbbbb_ppppppbbbbbbb__B_
        _B__bbbbbb_ppppppbbbbbbb__B_
        _B__bbbbbb_ppppppbbbbbbb__B_
        _B__bbbbbb_ppppppbbbbbbb__B_
        _B________________________B_
        _B_________rrrrrr_rrrrrr__B_
        _B_________rrrrrr_rrrrrr__B_
        _B_________rrrrrr_rrrrrr__B_
        _B_________rrrrrr_rrrrrr__B_
        _B_________rrrrrr_rrrrrr__B_
        _B_________rrrrrr_rrrrrr__B_
        _B________________________B_
        _B________________________B_
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        ____________________________
    `), fmt.Sprintf(tables_source, `
      x-table { border-color: #00f; table-layout: fixed;
                direction: rtl; }
      x-tbody { background: rgba(0, 0, 255, 1) }
      x-tr { background: rgba(255, 0, 0, 0.5) }
    `))
}

func TestTables_6(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqualFromPixels(t, "table_column_backgrounds", toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
        _B__bbbbbb_bbbbbb_rrrrrr__B_
        _B__bbbbbb_bbbbbb_rrrrrr__B_
        _B__bbbbbb_bbbbbb_rrrrrr__B_
        _B__bbbbbb_bbbbbb_rrrrrr__B_
        _B__bbbbbb_bbbbbb_rrrrrr__B_
        _B__bbbbbb_bbbbbb_rrrrrr__B_
        _B_________bbbbbb_________B_
        _B__bbbbbbbpppppp_rrrrrr__B_
        _B__bbbbbbbpppppp_rrrrrr__B_
        _B__bbbbbbbpppppp_rrrrrr__B_
        _B__bbbbbbbpppppp_rrrrrr__B_
        _B__bbbbbbbpppppp_rrrrrr__B_
        _B__bbbbbbbpppppp_rrrrrr__B_
        _B________________________B_
        _B__bbbbbb_bbbbbb_________B_
        _B__bbbbbb_bbbbbb_________B_
        _B__bbbbbb_bbbbbb_________B_
        _B__bbbbbb_bbbbbb_________B_
        _B__bbbbbb_bbbbbb_________B_
        _B__bbbbbb_bbbbbb_________B_
        _B________________________B_
        _B________________________B_
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        ____________________________
    `), fmt.Sprintf(tables_source, `
      x-table { border-color: #00f; table-layout: fixed;}
      x-colgroup { background: rgba(0, 0, 255, 1) }
      x-col { background: rgba(255, 0, 0, 0.5) }
    `))
}

func TestTables_6Rtl(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqualFromPixels(t, "table_column_backgrounds_rtl", toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
        _B__rrrrrr_bbbbbb_bbbbbb__B_
        _B__rrrrrr_bbbbbb_bbbbbb__B_
        _B__rrrrrr_bbbbbb_bbbbbb__B_
        _B__rrrrrr_bbbbbb_bbbbbb__B_
        _B__rrrrrr_bbbbbb_bbbbbb__B_
        _B__rrrrrr_bbbbbb_bbbbbb__B_
        _B_________bbbbbb_________B_
        _B__rrrrrr_ppppppbbbbbbb__B_
        _B__rrrrrr_ppppppbbbbbbb__B_
        _B__rrrrrr_ppppppbbbbbbb__B_
        _B__rrrrrr_ppppppbbbbbbb__B_
        _B__rrrrrr_ppppppbbbbbbb__B_
        _B__rrrrrr_ppppppbbbbbbb__B_
        _B________________________B_
        _B_________bbbbbb_bbbbbb__B_
        _B_________bbbbbb_bbbbbb__B_
        _B_________bbbbbb_bbbbbb__B_
        _B_________bbbbbb_bbbbbb__B_
        _B_________bbbbbb_bbbbbb__B_
        _B_________bbbbbb_bbbbbb__B_
        _B________________________B_
        _B________________________B_
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        ____________________________
    `), fmt.Sprintf(tables_source, `
      x-table { border-color: #00f; table-layout: fixed;
                direction: rtl; }
      x-colgroup { background: rgba(0, 0, 255, 1) }
      x-col { background: rgba(255, 0, 0, 0.5) }
    `))
}

func TestTables_7(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqualFromPixels(t, "table_borders_and_row_backgrounds", toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
        _B__bbbbbb_bbbbbb_bbbbbb__B_
        _B__bBBBBb_bBBBBb_bBBBBb__B_
        _B__bBBBBb_bBBBBb_bBBBBb__B_
        _B__bBBBBb_bBBBBb_bBBBBb__B_
        _B__bBBBBb_bBBBBb_bBBBBb__B_
        _B__bbbbbb_bBBBBb_bbbbbb__B_
        _B_________bBBBBb_________B_
        _B__rrrrrrrpbbbbp_rrrrrr__B_
        _B__r______bBBBBp_r____r__B_
        _B__r______bBBBBp_r____r__B_
        _B__r______bBBBBp_r____r__B_
        _B__r______bBBBBp_r____r__B_
        _B__rrrrrrrpppppp_rrrrrr__B_
        _B________________________B_
        _B__rrrrrr_rrrrrr_________B_
        _B__r____r_r____r_________B_
        _B__r____r_r____r_________B_
        _B__r____r_r____r_________B_
        _B__r____r_r____r_________B_
        _B__rrrrrr_rrrrrr_________B_
        _B________________________B_
        _B________________________B_
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        ____________________________
    `), fmt.Sprintf(tables_source, `
      x-table { border-color: #00f; table-layout: fixed }
      x-tr:first-child { background: blue }
      x-td { border-color: rgba(255, 0, 0, 0.5) }
    `))
}

func TestTables_7Rtl(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqualFromPixels(t, "table_borders_and_row_backgrounds_rtl", toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
        _B__bbbbbb_bbbbbb_bbbbbb__B_
        _B__bBBBBb_bBBBBb_bBBBBb__B_
        _B__bBBBBb_bBBBBb_bBBBBb__B_
        _B__bBBBBb_bBBBBb_bBBBBb__B_
        _B__bBBBBb_bBBBBb_bBBBBb__B_
        _B__bbbbbb_bBBBBb_bbbbbb__B_
        _B_________bBBBBb_________B_
        _B__rrrrrr_pbbbbprrrrrrr__B_
        _B__r____r_pBBBBb______r__B_
        _B__r____r_pBBBBb______r__B_
        _B__r____r_pBBBBb______r__B_
        _B__r____r_pBBBBb______r__B_
        _B__rrrrrr_pppppprrrrrrr__B_
        _B________________________B_
        _B_________rrrrrr_rrrrrr__B_
        _B_________r____r_r____r__B_
        _B_________r____r_r____r__B_
        _B_________r____r_r____r__B_
        _B_________r____r_r____r__B_
        _B_________rrrrrr_rrrrrr__B_
        _B________________________B_
        _B________________________B_
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        ____________________________
    `), fmt.Sprintf(tables_source, `
      x-table { border-color: #00f; table-layout: fixed;
                direction: rtl; }
      x-tr:first-child { background: blue }
      x-td { border-color: rgba(255, 0, 0, 0.5) }
    `))
}

func TestTables_8(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqualFromPixels(t, "table_borders_and_column_backgrounds", toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
        _B__bbbbbb_rrrrrr_rrrrrr__B_
        _B__bBBBBb_r____r_r____r__B_
        _B__bBBBBb_r____r_r____r__B_
        _B__bBBBBb_r____r_r____r__B_
        _B__bBBBBb_r____r_r____r__B_
        _B__bbbbbb_r____r_rrrrrr__B_
        _B_________r____r_________B_
        _B__bbbbbbbpbbbbp_rrrrrr__B_
        _B__bBBBBBBbBBBBp_r____r__B_
        _B__bBBBBBBbBBBBp_r____r__B_
        _B__bBBBBBBbBBBBp_r____r__B_
        _B__bBBBBBBbBBBBp_r____r__B_
        _B__bbbbbbbpppppp_rrrrrr__B_
        _B________________________B_
        _B__bbbbbb_rrrrrr_________B_
        _B__bBBBBb_r____r_________B_
        _B__bBBBBb_r____r_________B_
        _B__bBBBBb_r____r_________B_
        _B__bBBBBb_r____r_________B_
        _B__bbbbbb_rrrrrr_________B_
        _B________________________B_
        _B________________________B_
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        ____________________________
    `), fmt.Sprintf(tables_source, `
      x-table { border-color: #00f; table-layout: fixed }
      x-col:first-child { background: blue }
      x-td { border-color: rgba(255, 0, 0, 0.5) }
    `))
}

func TestTables_8Rtl(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqualFromPixels(t, "table_borders_and_column_backgrounds_rtl", toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
        _B__rrrrrr_rrrrrr_bbbbbb__B_
        _B__r____r_r____r_bBBBBb__B_
        _B__r____r_r____r_bBBBBb__B_
        _B__r____r_r____r_bBBBBb__B_
        _B__r____r_r____r_bBBBBb__B_
        _B__rrrrrr_r____r_bbbbbb__B_
        _B_________r____r_________B_
        _B__rrrrrr_pbbbbpbbbbbbb__B_
        _B__r____r_pBBBBbBBBBBBb__B_
        _B__r____r_pBBBBbBBBBBBb__B_
        _B__r____r_pBBBBbBBBBBBb__B_
        _B__r____r_pBBBBbBBBBBBb__B_
        _B__rrrrrr_ppppppbbbbbbb__B_
        _B________________________B_
        _B_________rrrrrr_bbbbbb__B_
        _B_________r____r_bBBBBb__B_
        _B_________r____r_bBBBBb__B_
        _B_________r____r_bBBBBb__B_
        _B_________r____r_bBBBBb__B_
        _B_________rrrrrr_bbbbbb__B_
        _B________________________B_
        _B________________________B_
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        ____________________________
    `), fmt.Sprintf(tables_source, `
      x-table { border-color: #00f; table-layout: fixed;
                direction: rtl; }
      x-col:first-child { background: blue }
      x-td { border-color: rgba(255, 0, 0, 0.5) }
    `))
}

func TestTables_9(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "collapsed_border_thead", `
        ______________________
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBB____R____R____BBB_
        _BBB____R____R____BBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        __R_____R____R_____R__
        __R_____R____R_____R__
        __RRRRRRRRRRRRRRRRRR__
        __R_____R____R_____R__
        __R_____R____R_____R__
        __RRRRRRRRRRRRRRRRRR__
        ______________________
        ______________________
        ______________________
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBB____R____R____BBB_
        _BBB____R____R____BBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        __R_____R____R_____R__
        __RRRRRRRRRRRRRRRRRR__
        ______________________
        ______________________
        ______________________
        ______________________
        ______________________
        ______________________
        ______________________
        ______________________
    `, `
      <style>
        @page { size: 22px 18px; margin: 1px; background: #fff }
        td { border: 1px red solid; width: 4px; height: 2px; }
      </style>
      <table style="table-layout: fixed; border-collapse: collapse">
        <thead style="border: blue solid; border-width: 3px;
            "><td></td><td></td><td></td></thead>
        <tr><td></td><td></td><td></td></tr>
        <tr><td></td><td></td><td></td></tr>
        <tr><td></td><td></td><td></td></tr>`)
}

func TestTables_10(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "collapsed_border_tfoot", `
        ______________________
        __RRRRRRRRRRRRRRRRRR__
        __R_____R____R_____R__
        __R_____R____R_____R__
        __RRRRRRRRRRRRRRRRRR__
        __R_____R____R_____R__
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBB____R____R____BBB_
        _BBB____R____R____BBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        ______________________
        ______________________
        ______________________
        ______________________
        __RRRRRRRRRRRRRRRRRR__
        __R_____R____R_____R__
        __R_____R____R_____R__
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBB____R____R____BBB_
        _BBB____R____R____BBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        ______________________
        ______________________
        ______________________
        ______________________
        ______________________
    `, `
      <style>
        @page { size: 22px 17px; margin: 1px; background: #fff }
        td { border: 1px red solid; width: 4px; height: 2px; }
      </style>
      <table style="table-layout: fixed; margin-left: 1px;
                    border-collapse: collapse">
        <tr><td></td><td></td><td></td></tr>
        <tr><td></td><td></td><td></td></tr>
        <tr><td></td><td></td><td></td></tr>
        <tfoot style="border: blue solid; border-width: 3px;
            "><td></td><td></td><td></td></tfoot>`)
}

func TestTables_11(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for inline table with collapsed border and alignment
	// rendering borders incorrectly
	// https://github.com/Kozea/WeasyPrint/issues/82
	assertPixelsEqual(t, "inline_text_align", `
      ____________________
      ________RRRRRRRRRRR_
      ________R____R____R_
      ________R____R____R_
      ________R____R____R_
      ________RRRRRRRRRRR_
      ____________________
      ____________________
      ____________________
      ____________________
    `, `
      <style>
        @page { size: 20px 10px; margin: 1px; background: #fff }
        body { text-align: right; font-size: 0 }
        table { display: inline-table; width: 11px }
        td { border: 1px red solid; width: 4px; height: 3px }
      </style>
      <table style="table-layout: fixed; border-collapse: collapse">
        <tr><td></td><td></td></tr>`)
}

func TestTables_12(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqualFromPixels(t, "table_collapsed_borders", toPix(`
        ____________________________
        _________BBBBBBBBBBBBBBBBBB_
        _________BBBBBBBBBBBBBBBBBB_
        _________BB____r____r____BB_
        _________BB____r____r____BB_
        _________BB____r____r____BB_
        _________BB____r____r____BB_
        _________BBrrrrr____rrrrrBB_
        _________BB____r_________BB_
        _________BB____r_________BB_
        _________BB____r_________BB_
        _________BB____r_________BB_
        _________BBrrrrrrrrrrrrrrBB_
        _________BB____r____r____BB_
        _________BB____r____r____BB_
        _________BB____r____r____BB_
        _________BB____r____r____BB_
        _________BBBBBBBBBBBBBBBBBB_
        _________BBBBBBBBBBBBBBBBBB_
        ____________________________
        ____________________________
        ____________________________
        ____________________________
        ____________________________
        ____________________________
        ____________________________
        ____________________________
        ____________________________
    `), fmt.Sprintf(tables_source, `
      body { direction: rtl }
      x-table { border: 2px solid #00f; table-layout: fixed;
                border-collapse: collapse }
      x-td { border-color: #ff7f7f }
    `))
}

func TestTables_13(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqualFromPixels(t, "table_collapsed_borders_paged", toPix(`
        ____________________________
        _gggggggggggggggggggggggggg_
        _g________________________g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BB____r____r____BB_g_
        _g_____BB____r____r____BB_g_
        _g_____BB____r____r____BB_g_
        _g_____BB____r____r____BB_g_
        _g_____BBrrrrr____rrrrrBB_g_
        _g_____BB____r_________BB_g_
        _g_____BB____r_________BB_g_
        _g_____BB____r_________BB_g_
        _g_____BB____r_________BB_g_
        _g_____BBrrrrrrrrrrrrrrBB_g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _gggggggggggggggggggggggggg_
        ____________________________
        ____________________________
        _gggggggggggggggggggggggggg_
        _g_____BBrrrrrrrrrrrrrrBB_g_
        _g_____BB____r____r____BB_g_
        _g_____BB____r____r____BB_g_
        _g_____BB____r____r____BB_g_
        _g_____BB____r____r____BB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g_____BBBBBBBBBBBBBBBBBB_g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _g________________________g_
        _gggggggggggggggggggggggggg_
        ____________________________
    `), fmt.Sprintf(tables_source, `
      body { direction: rtl }
      x-table { border: solid #00f; border-width: 8px 2px;
                table-layout: fixed; border-collapse: collapse }
      x-td { border-color: #ff7f7f }
      @page { size: 28px 26px; margin: 1px;
              border: 1px solid rgba(0, 255, 0, 0.5); }
    `))
}

// @pytest.mark.xfail
// func TestTables_14(t *testing.T) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     assertPixelsEqualFromPixels(t, "table_background_column_paged", toPix(`
//         ____________________________
//         _RRR_RRR_RRR________________
//         _RRR_RRR_RRR________________
//         _RRR_RRR_RRR________________
//         _RRR_RRR_RRR________________
//         _RRR_RRR_RRR________________
//         _RRR_RRR_RRR________________
//         _RRR_RRR_RRR________________
//         _RRR_RRR_RRR________________
//         _RRR_RRR_RRR________________
//         _RRR_RRR_RRR________________
//         _____RRR____________________
//         _RRRRRRR_RRR________________
//         _RRRRRRR_RRR________________
//         _RRRRRRR_RRR________________
//         _RRRRRRR_RRR________________
//         _RRRRRRR_RRR________________
//         _RRRRRRR_RRR________________
//         _RRRRRRR_RRR________________
//         _RRRRRRR_RRR________________
//         _RRRRRRR_RRR________________
//         _RRRRRRR_RRR________________
//         ____________________________
//         ____________________________
//         ____________________________
//         ____________________________
//         ____________________________
//         _RRR_RRR____________________
//         _RRR_RRR____________________
//         _RRR_RRR____________________
//         _RRR_RRR____________________
//         _RRR_RRR____________________
//         _RRR_RRR____________________
//         _RRR_RRR____________________
//         _RRR_RRR____________________
//         _RRR_RRR____________________
//         _RRR_RRR____________________
//         ____________________________
//         ____________________________
//         ____________________________
//         ____________________________
//         ____________________________
//         ____________________________
//         ____________________________
//         ____________________________
//         ____________________________
//         ____________________________
//         ____________________________
//         ____________________________
//         ____________________________
//         ____________________________
//         ____________________________
//     `), fmt.Sprintf( tables_source, `
//       @page { size: 28px 26px }
//       x-table { margin: 0; padding: 0; border: 0 }
//       x-col { background: red }
//       x-td { padding: 0; width: 1px; height: 8px }
//     `))
// }

func TestTables_15(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test for colspan in last body line with footer
	// https://github.com/Kozea/WeasyPrint/issues/1250
	assertPixelsEqual(t, "colspan_last_row", `
        ______________________
        __RRRRRRRRRRRRRRRRRR__
        __R_____R____R_____R__
        __R_____R____R_____R__
        __R_____R____R_____R__
        __RRRRRRRRRRRRRRRRRR__
        __R_____R____R_____R__
        __R_____R____R_____R__
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBB____R____R____BBB_
        _BBB____R____R____BBB_
        _BBB____R____R____BBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        ______________________
        ______________________
        __RRRRRRRRRRRRRRRRRR__
        __R________________R__
        __R________________R__
        __R________________R__
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBB____R____R____BBB_
        _BBB____R____R____BBB_
        _BBB____R____R____BBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        _BBBBBBBBBBBBBBBBBBBB_
        ______________________
        ______________________
        ______________________
        ______________________
    `, `
      <style>
        @page { size: 22px 18px; margin: 1px; background: #fff }
        td { border: 1px red solid; width: 4px; height: 3px; }
      </style>
      <table style="table-layout: fixed; margin-left: 1px;
                    border-collapse: collapse">
        <tr><td></td><td></td><td></td></tr>
        <tr><td></td><td></td><td></td></tr>
        <tr><td colspan="3"></td></tr>
        <tfoot style="border: blue solid; border-width: 3px;
            "><td></td><td></td><td></td></tfoot>`)
}

func TestTables_16(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "table_absolute", `
      ____________________
      _RRRRRRRRRRR________
      _R____R____R________
      _R____R____R________
      _R____R_RRRRRRRRRRR_
      _RRRRRRRRRRR_R____R_
      ________R____R____R_
      ________R____R____R_
      ________RRRRRRRRRRR_
      ____________________
    `, `
      <style>
        @page { size: 20px 10px; margin: 1px; background: #fff }
        body { text-align: right; font-size: 0 }
        table { position: absolute; width: 11px;
                table-layout: fixed; border-collapse: collapse }
        td { border: 1px red solid; width: 4px; height: 3px }
      </style>
      <table style="top: 0; left: 0">
        <tr><td></td><td></td></tr>
      <table style="bottom: 0; right: 0">
        <tr><td></td><td></td></tr>`)
}
