package pdf

import (
	"fmt"
	"testing"

	"github.com/benoitkugler/webrender/utils/testutils"
)

// Test visibility.

const visibilitySource = `
  <style>
    @page { size: 12px 7px }
    body { background: #fff; font: 1px/1 serif }
    img { margin: 1px 0 0 1px; }
    %s
  </style>
  <div>
    <img src="../resources_test/pattern.png">
    <span><img src="../resources_test/pattern.png"></span>
  </div>`

func TestVisibility_1(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "visibility_reference", `
        ____________
        _rBBB_rBBB__
        _BBBB_BBBB__
        _BBBB_BBBB__
        _BBBB_BBBB__
        ____________
        ____________
    `, fmt.Sprintf(visibilitySource, ""))
}

func TestVisibility_2(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "visibility_hidden", `
        ____________
        ____________
        ____________
        ____________
        ____________
        ____________
        ____________
    `, fmt.Sprintf(visibilitySource, "div { visibility: hidden }"))
}

func TestVisibility_3(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "visibility_mixed", `
        ____________
        ______rBBB__
        ______BBBB__
        ______BBBB__
        ______BBBB__
        ____________
        ____________
    `, fmt.Sprintf(visibilitySource, `div { visibility: hidden }
                                 span { visibility: visible } `))
}
