package pdf

import (
	"fmt"
	"image/color"
	"strings"
	"testing"

	"github.com/benoitkugler/webrender/utils/testutils"
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

const tables_source = `
  <style>
    @page { size: 28px; }
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

	assertPixelsEqualFromPixels(t, toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
        _B__ssssss_ssssss_ssssss__B_
        _B__s____s_s____s_s____s__B_
        _B__s____s_s____s_s____s__B_
        _B__s____s_s____s_s____s__B_
        _B__s____s_s____s_s____s__B_
        _B__ssssss_s____s_ssssss__B_
        _B_________s____s_________B_
        _B__sssssssSssssS_ssssss__B_
        _B__s______s____S_s____s__B_
        _B__s______s____S_s____s__B_
        _B__s______s____S_s____s__B_
        _B__s______s____S_s____s__B_
        _B__sssssssSSSSSS_ssssss__B_
        _B________________________B_
        _B__ssssss_ssssss_________B_
        _B__s____s_s____s_________B_
        _B__s____s_s____s_________B_
        _B__s____s_s____s_________B_
        _B__s____s_s____s_________B_
        _B__ssssss_ssssss_________B_
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

	assertPixelsEqualFromPixels(t, toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
        _B__ssssss_ssssss_ssssss__B_
        _B__s____s_s____s_s____s__B_
        _B__s____s_s____s_s____s__B_
        _B__s____s_s____s_s____s__B_
        _B__s____s_s____s_s____s__B_
        _B__ssssss_s____s_ssssss__B_
        _B_________s____s_________B_
        _B__ssssss_SssssSsssssss__B_
        _B__s____s_S____s______s__B_
        _B__s____s_S____s______s__B_
        _B__s____s_S____s______s__B_
        _B__s____s_S____s______s__B_
        _B__ssssss_SSSSSSsssssss__B_
        _B________________________B_
        _B_________ssssss_ssssss__B_
        _B_________s____s_s____s__B_
        _B_________s____s_s____s__B_
        _B_________s____s_s____s__B_
        _B_________s____s_s____s__B_
        _B_________ssssss_ssssss__B_
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

	assertPixelsEqualFromPixels(t, toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBB_________
        _BBBBBBBBBBBBBBBBBB_________
        _BB____s____s____BB_________
        _BB____s____s____BB_________
        _BB____s____s____BB_________
        _BB____s____s____BB_________
        _BBsssss____sssssBB_________
        _BB_________s____BB_________
        _BB_________s____BB_________
        _BB_________s____BB_________
        _BB_________s____BB_________
        _BBssssssssssssssBB_________
        _BB____s____s____BB_________
        _BB____s____s____BB_________
        _BB____s____s____BB_________
        _BB____s____s____BB_________
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

	assertPixelsEqualFromPixels(t, toPix(`
        ____________________________
        _________BBBBBBBBBBBBBBBBBB_
        _________BBBBBBBBBBBBBBBBBB_
        _________BB____s____s____BB_
        _________BB____s____s____BB_
        _________BB____s____s____BB_
        _________BB____s____s____BB_
        _________BBsssss____sssssBB_
        _________BB____s_________BB_
        _________BB____s_________BB_
        _________BB____s_________BB_
        _________BB____s_________BB_
        _________BBssssssssssssssBB_
        _________BB____s____s____BB_
        _________BB____s____s____BB_
        _________BB____s____s____BB_
        _________BB____s____s____BB_
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

	assertPixelsEqualFromPixels(t, toPix(`
		____________________________
		_tttttttttttttttttttttttttt_
		_t________________________t_
		_t_BBBBBBBBBBBBBBBBBB_____t_
		_t_BBBBBBBBBBBBBBBBBB_____t_
		_t_BBBBBBBBBBBBBBBBBB_____t_
		_t_BBBBBBBBBBBBBBBBBB_____t_
		_t_BBBBBBBBBBBBBBBBBB_____t_
		_t_BBBBBBBBBBBBBBBBBB_____t_
		_t_BBBBBBBBBBBBBBBBBB_____t_
		_t_BBBBBBBBBBBBBBBBBB_____t_
		_t_BB____s____s____BB_____t_
		_t_BB____s____s____BB_____t_
		_t_BB____s____s____BB_____t_
		_t_BB____s____s____BB_____t_
		_t_BBsssss____sssssBB_____t_
		_t_BB_________s____BB_____t_
		_t_BB_________s____BB_____t_
		_t_BB_________s____BB_____t_
		_t_BB_________s____BB_____t_
		_t_BBssssssssssssssBB_____t_
		_t________________________t_
		_t________________________t_
		_t________________________t_
		_tttttttttttttttttttttttttt_
		____________________________
		____________________________
		_tttttttttttttttttttttttttt_
		_t_BBssssssssssssssBB_____t_
		_t_BB____s____s____BB_____t_
		_t_BB____s____s____BB_____t_
		_t_BB____s____s____BB_____t_
		_t_BB____s____s____BB_____t_
		_t_BBBBBBBBBBBBBBBBBB_____t_
		_t_BBBBBBBBBBBBBBBBBB_____t_
		_t_BBBBBBBBBBBBBBBBBB_____t_
		_t_BBBBBBBBBBBBBBBBBB_____t_
		_t_BBBBBBBBBBBBBBBBBB_____t_
		_t_BBBBBBBBBBBBBBBBBB_____t_
		_t_BBBBBBBBBBBBBBBBBB_____t_
		_t_BBBBBBBBBBBBBBBBBB_____t_
		_t________________________t_
		_t________________________t_
		_t________________________t_
		_t________________________t_
		_t________________________t_
		_t________________________t_
		_t________________________t_
		_t________________________t_
		_t________________________t_
		_tttttttttttttttttttttttttt_
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

	assertPixelsEqualFromPixels(t, toPix(`
		____________________________
		_tttttttttttttttttttttttttt_
		_t________________________t_
		_t_____BBBBBBBBBBBBBBBBBB_t_
		_t_____BBBBBBBBBBBBBBBBBB_t_
		_t_____BBBBBBBBBBBBBBBBBB_t_
		_t_____BBBBBBBBBBBBBBBBBB_t_
		_t_____BBBBBBBBBBBBBBBBBB_t_
		_t_____BBBBBBBBBBBBBBBBBB_t_
		_t_____BBBBBBBBBBBBBBBBBB_t_
		_t_____BBBBBBBBBBBBBBBBBB_t_
		_t_____BB____s____s____BB_t_
		_t_____BB____s____s____BB_t_
		_t_____BB____s____s____BB_t_
		_t_____BB____s____s____BB_t_
		_t_____BBsssss____sssssBB_t_
		_t_____BB____s_________BB_t_
		_t_____BB____s_________BB_t_
		_t_____BB____s_________BB_t_
		_t_____BB____s_________BB_t_
		_t_____BBssssssssssssssBB_t_
		_t________________________t_
		_t________________________t_
		_t________________________t_
		_tttttttttttttttttttttttttt_
		____________________________
		____________________________
		_tttttttttttttttttttttttttt_
		_t_____BBssssssssssssssBB_t_
		_t_____BB____s____s____BB_t_
		_t_____BB____s____s____BB_t_
		_t_____BB____s____s____BB_t_
		_t_____BB____s____s____BB_t_
		_t_____BBBBBBBBBBBBBBBBBB_t_
		_t_____BBBBBBBBBBBBBBBBBB_t_
		_t_____BBBBBBBBBBBBBBBBBB_t_
		_t_____BBBBBBBBBBBBBBBBBB_t_
		_t_____BBBBBBBBBBBBBBBBBB_t_
		_t_____BBBBBBBBBBBBBBBBBB_t_
		_t_____BBBBBBBBBBBBBBBBBB_t_
		_t_____BBBBBBBBBBBBBBBBBB_t_
		_t________________________t_
		_t________________________t_
		_t________________________t_
		_t________________________t_
		_t________________________t_
		_t________________________t_
		_t________________________t_
		_t________________________t_
		_t________________________t_
		_tttttttttttttttttttttttttt_
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

	assertPixelsEqualFromPixels(t, toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
        _B__ssssss_ssssss_ssssss__B_
        _B__ssssss_ssssss_ssssss__B_
        _B__ssssss_ssssss_ssssss__B_
        _B__ssssss_ssssss_ssssss__B_
        _B__ssssss_ssssss_ssssss__B_
        _B__ssssss_ssssss_ssssss__B_
        _B_________ssssss_________B_
        _B__sssssssSSSSSS_ssssss__B_
        _B__sssssssSSSSSS_ssssss__B_
        _B__sssssssSSSSSS_ssssss__B_
        _B__sssssssSSSSSS_ssssss__B_
        _B__sssssssSSSSSS_ssssss__B_
        _B__sssssssSSSSSS_ssssss__B_
        _B________________________B_
        _B__ssssss_ssssss_________B_
        _B__ssssss_ssssss_________B_
        _B__ssssss_ssssss_________B_
        _B__ssssss_ssssss_________B_
        _B__ssssss_ssssss_________B_
        _B__ssssss_ssssss_________B_
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

	assertPixelsEqualFromPixels(t, toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
        _B__ssssss_ssssss_ssssss__B_
        _B__ssssss_ssssss_ssssss__B_
        _B__ssssss_ssssss_ssssss__B_
        _B__ssssss_ssssss_ssssss__B_
        _B__ssssss_ssssss_ssssss__B_
        _B__ssssss_ssssss_ssssss__B_
        _B_________ssssss_________B_
        _B__ssssss_SSSSSSsssssss__B_
        _B__ssssss_SSSSSSsssssss__B_
        _B__ssssss_SSSSSSsssssss__B_
        _B__ssssss_SSSSSSsssssss__B_
        _B__ssssss_SSSSSSsssssss__B_
        _B__ssssss_SSSSSSsssssss__B_
        _B________________________B_
        _B_________ssssss_ssssss__B_
        _B_________ssssss_ssssss__B_
        _B_________ssssss_ssssss__B_
        _B_________ssssss_ssssss__B_
        _B_________ssssss_ssssss__B_
        _B_________ssssss_ssssss__B_
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

	assertPixelsEqualFromPixels(t, toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
        _B__uuuuuu_uuuuuu_uuuuuu__B_
        _B__uuuuuu_uuuuuu_uuuuuu__B_
        _B__uuuuuu_uuuuuu_uuuuuu__B_
        _B__uuuuuu_uuuuuu_uuuuuu__B_
        _B__uuuuuu_uuuuuu_uuuuuu__B_
        _B__uuuuuu_uuuuuu_uuuuuu__B_
        _B_________uuuuuu_________B_
        _B__uuuuuuupppppp_uuuuuu__B_
        _B__uuuuuuupppppp_uuuuuu__B_
        _B__uuuuuuupppppp_uuuuuu__B_
        _B__uuuuuuupppppp_uuuuuu__B_
        _B__uuuuuuupppppp_uuuuuu__B_
        _B__uuuuuuupppppp_uuuuuu__B_
        _B________________________B_
        _B__ssssss_ssssss_________B_
        _B__ssssss_ssssss_________B_
        _B__ssssss_ssssss_________B_
        _B__ssssss_ssssss_________B_
        _B__ssssss_ssssss_________B_
        _B__ssssss_ssssss_________B_
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

	assertPixelsEqualFromPixels(t, toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
        _B__uuuuuu_uuuuuu_uuuuuu__B_
        _B__uuuuuu_uuuuuu_uuuuuu__B_
        _B__uuuuuu_uuuuuu_uuuuuu__B_
        _B__uuuuuu_uuuuuu_uuuuuu__B_
        _B__uuuuuu_uuuuuu_uuuuuu__B_
        _B__uuuuuu_uuuuuu_uuuuuu__B_
        _B_________uuuuuu_________B_
        _B__uuuuuu_ppppppuuuuuuu__B_
        _B__uuuuuu_ppppppuuuuuuu__B_
        _B__uuuuuu_ppppppuuuuuuu__B_
        _B__uuuuuu_ppppppuuuuuuu__B_
        _B__uuuuuu_ppppppuuuuuuu__B_
        _B__uuuuuu_ppppppuuuuuuu__B_
        _B________________________B_
        _B_________ssssss_ssssss__B_
        _B_________ssssss_ssssss__B_
        _B_________ssssss_ssssss__B_
        _B_________ssssss_ssssss__B_
        _B_________ssssss_ssssss__B_
        _B_________ssssss_ssssss__B_
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

	assertPixelsEqualFromPixels(t, toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
        _B__uuuuuu_uuuuuu_ssssss__B_
        _B__uuuuuu_uuuuuu_ssssss__B_
        _B__uuuuuu_uuuuuu_ssssss__B_
        _B__uuuuuu_uuuuuu_ssssss__B_
        _B__uuuuuu_uuuuuu_ssssss__B_
        _B__uuuuuu_uuuuuu_ssssss__B_
        _B_________uuuuuu_________B_
        _B__uuuuuuupppppp_ssssss__B_
        _B__uuuuuuupppppp_ssssss__B_
        _B__uuuuuuupppppp_ssssss__B_
        _B__uuuuuuupppppp_ssssss__B_
        _B__uuuuuuupppppp_ssssss__B_
        _B__uuuuuuupppppp_ssssss__B_
        _B________________________B_
        _B__uuuuuu_uuuuuu_________B_
        _B__uuuuuu_uuuuuu_________B_
        _B__uuuuuu_uuuuuu_________B_
        _B__uuuuuu_uuuuuu_________B_
        _B__uuuuuu_uuuuuu_________B_
        _B__uuuuuu_uuuuuu_________B_
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

	assertPixelsEqualFromPixels(t, toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
        _B__ssssss_uuuuuu_uuuuuu__B_
        _B__ssssss_uuuuuu_uuuuuu__B_
        _B__ssssss_uuuuuu_uuuuuu__B_
        _B__ssssss_uuuuuu_uuuuuu__B_
        _B__ssssss_uuuuuu_uuuuuu__B_
        _B__ssssss_uuuuuu_uuuuuu__B_
        _B_________uuuuuu_________B_
        _B__ssssss_ppppppuuuuuuu__B_
        _B__ssssss_ppppppuuuuuuu__B_
        _B__ssssss_ppppppuuuuuuu__B_
        _B__ssssss_ppppppuuuuuuu__B_
        _B__ssssss_ppppppuuuuuuu__B_
        _B__ssssss_ppppppuuuuuuu__B_
        _B________________________B_
        _B_________uuuuuu_uuuuuu__B_
        _B_________uuuuuu_uuuuuu__B_
        _B_________uuuuuu_uuuuuu__B_
        _B_________uuuuuu_uuuuuu__B_
        _B_________uuuuuu_uuuuuu__B_
        _B_________uuuuuu_uuuuuu__B_
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

	assertPixelsEqualFromPixels(t, toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
        _B__uuuuuu_uuuuuu_uuuuuu__B_
        _B__uBBBBu_uBBBBu_uBBBBu__B_
        _B__uBBBBu_uBBBBu_uBBBBu__B_
        _B__uBBBBu_uBBBBu_uBBBBu__B_
        _B__uBBBBu_uBBBBu_uBBBBu__B_
        _B__uuuuuu_uBBBBu_uuuuuu__B_
        _B_________uBBBBu_________B_
        _B__ssssssspuuuup_ssssss__B_
        _B__s______uBBBBp_s____s__B_
        _B__s______uBBBBp_s____s__B_
        _B__s______uBBBBp_s____s__B_
        _B__s______uBBBBp_s____s__B_
        _B__ssssssspppppp_ssssss__B_
        _B________________________B_
        _B__ssssss_ssssss_________B_
        _B__s____s_s____s_________B_
        _B__s____s_s____s_________B_
        _B__s____s_s____s_________B_
        _B__s____s_s____s_________B_
        _B__ssssss_ssssss_________B_
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

	assertPixelsEqualFromPixels(t, toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
        _B__uuuuuu_uuuuuu_uuuuuu__B_
        _B__uBBBBu_uBBBBu_uBBBBu__B_
        _B__uBBBBu_uBBBBu_uBBBBu__B_
        _B__uBBBBu_uBBBBu_uBBBBu__B_
        _B__uBBBBu_uBBBBu_uBBBBu__B_
        _B__uuuuuu_uBBBBu_uuuuuu__B_
        _B_________uBBBBu_________B_
        _B__ssssss_puuuupsssssss__B_
        _B__s____s_pBBBBu______s__B_
        _B__s____s_pBBBBu______s__B_
        _B__s____s_pBBBBu______s__B_
        _B__s____s_pBBBBu______s__B_
        _B__ssssss_ppppppsssssss__B_
        _B________________________B_
        _B_________ssssss_ssssss__B_
        _B_________s____s_s____s__B_
        _B_________s____s_s____s__B_
        _B_________s____s_s____s__B_
        _B_________s____s_s____s__B_
        _B_________ssssss_ssssss__B_
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

	assertPixelsEqualFromPixels(t, toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
		_B__uuuuuu_ssssss_ssssss__B_
        _B__uBBBBu_s____s_s____s__B_
        _B__uBBBBu_s____s_s____s__B_
        _B__uBBBBu_s____s_s____s__B_
        _B__uBBBBu_s____s_s____s__B_
        _B__uuuuuu_s____s_ssssss__B_
        _B_________s____s_________B_
        _B__uuuuuuupuuuup_ssssss__B_
        _B__uBBBBBBuBBBBp_s____s__B_
        _B__uBBBBBBuBBBBp_s____s__B_
        _B__uBBBBBBuBBBBp_s____s__B_
        _B__uBBBBBBuBBBBp_s____s__B_
        _B__uuuuuuupppppp_ssssss__B_
        _B________________________B_
        _B__uuuuuu_ssssss_________B_
        _B__uBBBBu_s____s_________B_
        _B__uBBBBu_s____s_________B_
        _B__uBBBBu_s____s_________B_
        _B__uBBBBu_s____s_________B_
        _B__uuuuuu_ssssss_________B_
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

	assertPixelsEqualFromPixels(t, toPix(`
        ____________________________
        _BBBBBBBBBBBBBBBBBBBBBBBBBB_
        _B________________________B_
        _B________________________B_
		_B__ssssss_ssssss_uuuuuu__B_
        _B__s____s_s____s_uBBBBu__B_
        _B__s____s_s____s_uBBBBu__B_
        _B__s____s_s____s_uBBBBu__B_
        _B__s____s_s____s_uBBBBu__B_
        _B__ssssss_s____s_uuuuuu__B_
        _B_________s____s_________B_
        _B__ssssss_puuuupuuuuuuu__B_
        _B__s____s_pBBBBuBBBBBBu__B_
        _B__s____s_pBBBBuBBBBBBu__B_
        _B__s____s_pBBBBuBBBBBBu__B_
        _B__s____s_pBBBBuBBBBBBu__B_
        _B__ssssss_ppppppuuuuuuu__B_
        _B________________________B_
        _B_________ssssss_uuuuuu__B_
        _B_________s____s_uBBBBu__B_
        _B_________s____s_uBBBBu__B_
        _B_________s____s_uBBBBu__B_
        _B_________s____s_uBBBBu__B_
        _B_________ssssss_uuuuuu__B_
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

	assertPixelsEqual(t, `
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
    `, `
      <style>
        @page { size: 22px 18px; margin: 1px; }
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

	assertPixelsEqual(t, `
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
        @page { size: 22px 17px; margin: 1px }
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
	assertPixelsEqual(t, `
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
        @page { size: 20px 10px; margin: 1px }
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

	assertPixelsEqualFromPixels(t, toPix(`
        ____________________________
        _________BBBBBBBBBBBBBBBBBB_
        _________BBBBBBBBBBBBBBBBBB_
        _________BB____s____s____BB_
        _________BB____s____s____BB_
        _________BB____s____s____BB_
        _________BB____s____s____BB_
        _________BBsssss____sssssBB_
        _________BB____s_________BB_
        _________BB____s_________BB_
        _________BB____s_________BB_
        _________BB____s_________BB_
        _________BBssssssssssssssBB_
        _________BB____s____s____BB_
        _________BB____s____s____BB_
        _________BB____s____s____BB_
        _________BB____s____s____BB_
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

	assertPixelsEqualFromPixels(t, toPix(`
        ____________________________
        _tttttttttttttttttttttttttt_
        _t________________________t_
        _t_____BBBBBBBBBBBBBBBBBB_t_
        _t_____BBBBBBBBBBBBBBBBBB_t_
        _t_____BBBBBBBBBBBBBBBBBB_t_
        _t_____BBBBBBBBBBBBBBBBBB_t_
        _t_____BBBBBBBBBBBBBBBBBB_t_
        _t_____BBBBBBBBBBBBBBBBBB_t_
        _t_____BBBBBBBBBBBBBBBBBB_t_
        _t_____BBBBBBBBBBBBBBBBBB_t_
        _t_____BB____s____s____BB_t_
        _t_____BB____s____s____BB_t_
        _t_____BB____s____s____BB_t_
        _t_____BB____s____s____BB_t_
        _t_____BBsssss____sssssBB_t_
        _t_____BB____s_________BB_t_
        _t_____BB____s_________BB_t_
        _t_____BB____s_________BB_t_
        _t_____BB____s_________BB_t_
        _t_____BBssssssssssssssBB_t_
        _t________________________t_
        _t________________________t_
        _t________________________t_
        _tttttttttttttttttttttttttt_
        ____________________________
        ____________________________
        _tttttttttttttttttttttttttt_
        _t_____BBssssssssssssssBB_t_
        _t_____BB____s____s____BB_t_
        _t_____BB____s____s____BB_t_
        _t_____BB____s____s____BB_t_
        _t_____BB____s____s____BB_t_
        _t_____BBBBBBBBBBBBBBBBBB_t_
        _t_____BBBBBBBBBBBBBBBBBB_t_
        _t_____BBBBBBBBBBBBBBBBBB_t_
        _t_____BBBBBBBBBBBBBBBBBB_t_
        _t_____BBBBBBBBBBBBBBBBBB_t_
        _t_____BBBBBBBBBBBBBBBBBB_t_
        _t_____BBBBBBBBBBBBBBBBBB_t_
        _t_____BBBBBBBBBBBBBBBBBB_t_
        _t________________________t_
        _t________________________t_
        _t________________________t_
        _t________________________t_
        _t________________________t_
        _t________________________t_
        _t________________________t_
        _t________________________t_
        _t________________________t_
        _tttttttttttttttttttttttttt_
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

//     assertPixelsEqualFromPixels(t,  toPix(`
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
	assertPixelsEqual(t, `
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
        @page { size: 22px 18px; margin: 1px }
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

	assertPixelsEqual(t, `
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
        @page { size: 20px 10px; margin: 1px }
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

func TestTables_17(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, `
      ________________
      _RRRRRRRRRRRRRR_
      _RRRRRRRRRRRRRR_
      _RR____RR____RR_
      _RR_BB_RR_BB_RR_
      _RR_BB_RR_BB_RR_
      _RR_BB_RR____RR_
      _RR_BB_RR____RR_
      _RR____RR____RR_
      ________________
      ________________
      _RR_BB_RR____RR_
      _RR_BB_RR____RR_
      _RR_BB_RR____RR_
      _RR_BB_RR____RR_
      _RR____RR____RR_
      _RRRRRRRRRRRRRR_
      _RRRRRRRRRRRRRR_
      ________________
      ________________
    `, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page { size: 16px 10px; margin: 1px }
        table { border-collapse: collapse; font-size: 2px; line-height: 1;
                color: blue; font-family: weasyprint }
        td { border: 2px red solid; padding: 1px; line-height: 1 }
      </style>
      <table><tr><td>a a a a</td><td>a</td></tr>`)
}

func TestTables_18(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, `
      ____________
      _RRRRRRRRRR_
      _R________R_
      _R_RRRRRR_R_
      _R_R____R_R_
      _R_R_BB_R_R_
      _R_R_BB_R_R_
      _R_R_BB_R_R_
      _R_R_BB_R_R_
      _R_R____R_R_
      ____________
      ____________
      _R_R_BB_R_R_
      _R_R_BB_R_R_
      _R_R_BB_R_R_
      _R_R_BB_R_R_
      _R_R____R_R_
      _R_RRRRRR_R_
      _R________R_
      _RRRRRRRRRR_
      ____________
      ____________
    `, `
      <style>
        @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
        @page { size: 12px 11px; margin: 1px }
        table { border: 1px red solid; border-spacing: 1px; font-size: 2px;
                line-height: 1; color: blue; font-family: weasyprint }
        td { border: 1px red solid; padding: 1px; line-height: 1; }
      </style>
      <table><tr><td>a a a a</td></tr>`)
}

func TestTables_19(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// Regression test: https://github.com/Kozea/WeasyPrint/issues/1523
	assertPixelsEqual(t, `
      RR
      RR
      RR
      RR
      RR
      RR
      RR
      RR
    `, `
      <style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page { size: 2px 4px }
        table { border-collapse: collapse; color: red }
        body { font-size: 2px; font-family: weasyprint; line-height: 1 }
      </style>
      <table><tr><td>a a a a</td></tr></table>`)
}

func TestTables_20(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, `
      ____________________
      _RRRRRRRRRRRR_______
      _RBBBBBBBBBBR_______
      _RRRRRRRRRRRR_______
      ____________________
    `, `
      <style>
        @page { size: 20px 5px; margin: 1px }
        table { width: 10px; border: 1px red solid }
        td { height: 1px; background: blue }
        col, tr, tbody, tfoot { background: lime }
      </style>
      <table>
      <col></col><col></col>
      <tbody><tr></tr><tr><td></td></tr></tbody>
      <tfoot></tfoot>`)
}

func TestTables_21(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, `
      _________________________
      _rrrrrrrrrrrrrrrrrrrrrrr_
      _rBBBBBBBBBBrBBBBBBBBBBr_
      _rBKKKKKKBBBrBKKKKKKBBBr_
      _rBKKKKKKBBBrBKKKKKKBBBr_
      _rBBBBBBBBBBrBBBBBBBBBBr_
      _rrrrrrrrrrrrrrrrrrrrrrr_
      _________________________
      _________________________
      _________________________
      _________________________
      _________________________
      _rrrrrrrrrrrrrrrrrrrrrrr_
      _rBBBBBBBBBBrBBBBBBBBBBr_
      _rBKKKKKKBBBrBBBBBBBBBBr_
      _rBKKKKKKBBBrBBBBBBBBBBr_
      _rBBBBBBBBBBrBBBBBBBBBBr_
      _rrrrrrrrrrrrrrrrrrrrrrr_
      _________________________
      _________________________
      _________________________
      _________________________
    `, `
      <style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page { size: 25px 11px; margin: 1px }
        table { border-collapse: collapse; font: 2px weasyprint; width: 100% }
        td { background: blue; padding: 1px; border: 1px solid red }
      </style>
      <table>
        <tr><td>abc</td><td>abc</td></tr>
        <tr><td>abc</td><td></td></tr>`)
}

func TestTables_22(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, `
      _________________________
      _rrrrrrrrrrrrrrrrrrrrrrr_
      _rKKKKKKKKKKrKKKKKKKKKKr_
      _rKKKKKKKKKKrKKKKKKKKKKr_
      _rrrrrrrrrrrrrrrrrrrrrrr_
      _rKKKKKKBBBBrBBBBBBBBBBr_
      _rKKKKKKBBBBrBBBBBBBBBBr_
      _rBBBBBBBBBBrBBBBBBBBBBr_
      _________________________
      _________________________
      _rrrrrrrrrrrrrrrrrrrrrrr_
      _rKKKKKKKKKKrKKKKKKKKKKr_
      _rKKKKKKKKKKrKKKKKKKKKKr_
      _rrrrrrrrrrrrrrrrrrrrrrr_
      _rKKKKKKBBBBrBBBBBBBBBBr_
      _rKKKKKKBBBBrBBBBBBBBBBr_
      _rrrrrrrrrrrrrrrrrrrrrrr_
      _________________________
    `, `
      <style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page { size: 25px 9px; margin: 1px }
        table { border-collapse: collapse; font: 2px/1 weasyprint }
        td { background: blue; border: 1px solid red }
      </style>
      <table>
        <thead><tr><td>abcde</td><td>abcde</td></tr></thead>
        <tbody><tr><td>abc abc</td><td></td></tr></tbody>`)
}

func TestTables_23(t *testing.T) {
	t.Skip()
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, `
      _________________________
      _rrrrrrrrrrrrrrrrrrrrrrr_
      _rKKKKKKKKKKrKKKKKKKKKKr_
      _rKKKKKKKKKKrKKKKKKKKKKr_
      _rrrrrrrrrrrrrrrrrrrrrrr_
      _rKKKKKKBBBBrBBBBBBBBBBr_
      _rKKKKKKBBBBrBBBBBBBBBBr_
      _rBBBBBBBBBBrBBBBBBBBBBr_
      _________________________
      _________________________
      _rrrrrrrrrrrrrrrrrrrrrrr_
      _rKKKKKKKKKKrKKKKKKKKKKr_
      _rKKKKKKKKKKrKKKKKKKKKKr_
      _rKKKKKKBBBBrBBBBBBBBBBr_
      _rKKKKKKBBBBrBBBBBBBBBBr_
      _rrrrrrrrrrrrrrrrrrrrrrr_
      _________________________
      _________________________
    `, `
      <style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page { size: 25px 9px; margin: 1px }
        table { border-collapse: collapse; font: 2px/1 weasyprint }
        td { background: blue; border: 1px solid red }
        thead td { border-bottom: none }
      </style>
      <table>
        <thead><tr><td>abcde</td><td>abcde</td></tr></thead>
        <tbody><tr><td>abc abc</td><td></td></tr></tbody>`)
}

func TestRunningElementsTableBorderCollapse(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, strings.Repeat(`
      KK_____________
      KK_____________
      _______________
      _______________
      _______________
      KKKKKKK________
      KRRKRRK________
      KRRKRRK________
      KKKKKKK________
      KRRKRRK________
      KRRKRRK________
      KKKKKKK________
      _______________
      _______________
      _______________
    `, 2), `
      <style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page {
          margin: 0 0 10px 0;
          size: 15px;
          @bottom-left { content: element(table) }
        }
        body { font: 2px/1 weasyprint }
        table {
          border: 1px solid black;
          border-collapse: collapse;
          color: red;
          position: running(table);
        }
        td { border: 1px solid black }
        div { page-break-after: always }
      </style>
      <table>
        <tr> <td>A</td> <td>B</td> </tr>
        <tr> <td>C</td> <td>D</td> </tr>
      </table>
      <div>1</div>
      <div>2</div>
    `)
}

func TestRunningElementsTableBorderCollapseEmpty(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, strings.Repeat(`
      KK________
      KK________
      __________
      __________
      __________
      __________
      __________
      __________
      __________
      __________
    `, 2), `
      <style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page {
          margin: 0 0 5px 0;
          size: 10px;
          @bottom-left { content: element(table) }
        }
        body { font: 2px/1 weasyprint }
        table {
          border: 1px solid black;
          border-collapse: collapse;
          color: red;
          position: running(table);
        }
        td { border: 1px solid black }
        div { page-break-after: always }
      </style>
      <table></table>
      <div>1</div>
      <div>2</div>
    `)
}

func TestRunningElementsTableBorderCollapseBorderStyle(t *testing.T) {
	t.Skip()
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, strings.Repeat(`
      KK_____________
      KK_____________
      _______________
      _______________
      _______________
      KKKZ___________
      KRR_RR_________
      KRR_RR_________
      KKKK__Z________
      KRRKRRK________
      KRRKRRK________
      KKKKKKK________
      _______________
      _______________
      _______________
    `, 2), `
      <style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page {
          margin: 0 0 10px 0;
          size: 15px;
          @bottom-left { content: element(table) }
        }
        body { font: 2px/1 weasyprint }
        table {
          border: 1px solid black;
          border-collapse: collapse;
          color: red;
          position: running(table);
        }
        td { border: 1px solid black }
        div { page-break-after: always }
      </style>
      <table>
        <tr> <td>A</td> <td style="border-style: hidden">B</td> </tr>
        <tr> <td>C</td> <td style="border-style: none">D</td> </tr>
      </table>
      <div>1</div>
      <div>2</div>
    `)
}

func TestRunningElementsTableBorderCollapseSpan(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, strings.Repeat(`
      KK_____________
      KK_____________
      _______________
      _______________
      _______________
      KKKKKKKKKK_____
      KRRKRRKRRK_____
      KRRKRRKRRK_____
      K__KKKKKKK_____
      K__KRR___K_____
      K__KRR___K_____
      KKKKKKKKKK_____
      _______________
      _______________
      _______________
    `, 2), `
      <style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page {
          margin: 0 0 10px 0;
          size: 15px;
          @bottom-left { content: element(table) }
        }
        body { font: 2px/1 weasyprint }
        table {
          border: 1px solid black;
          border-collapse: collapse;
          color: red;
          position: running(table);
        }
        td { border: 1px solid black }
        div { page-break-after: always }
      </style>
      <table>
        <tr> <td rowspan=2>A</td> <td>B</td> <td>C</td> </tr>
        <tr> <td colspan=2>D</td> </tr>
      </table>
      <div>1</div>
      <div>2</div>
    `)
}

func TestRunningElementsTableBorderCollapseMargin(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, strings.Repeat(`
      KK_____________
      KK_____________
      _______________
      _______________
      _______________
      _______________
      ____KKKKKKK____
      ____KRRKRRK____
      ____KRRKRRK____
      ____KKKKKKK____
      ____KRRKRRK____
      ____KRRKRRK____
      ____KKKKKKK____
      _______________
      _______________
    `, 2), `
      <style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page {
          margin: 0 0 10px 0;
          size: 15px;
          @bottom-center { content: element(table); width: 100% }
        }
        body { font: 2px/1 weasyprint }
        table {
          border: 1px solid black;
          border-collapse: collapse;
          color: red;
          margin: 1px auto;
          position: running(table);
        }
        td { border: 1px solid black }
        div { page-break-after: always }
      </style>
      <table>
        <tr> <td>A</td> <td>B</td> </tr>
        <tr> <td>C</td> <td>D</td> </tr>
      </table>
      <div>1</div>
      <div>2</div>
    `)
}
