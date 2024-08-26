package pdf

import (
	"testing"

	"github.com/benoitkugler/webrender/utils/testutils"
)

// Test transformations.

func Test_2dTransform_1(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, `
        ________
        ________
        __BBBr__
        __BBBB__
        __BBBB__
        __BBBB__
        ________
        ________
    `, `
      <style>
        @page { size: 8px; margin: 2px; background: #fff; }
        div { transform: rotate(90deg); font-size: 0 }
      </style>
      <div><img src="../resources_test/pattern.png"></div>`)
}

func Test_2dTransform_2(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, `
        ____________
        ____________
        _____BBBr___
        _____BBBB___
        _____BBBB___
        _____BBBB___
        ____________
        ____________
        ____________
        ____________
        ____________
        ____________
    `, `
      <style>
        @page { size: 12px; margin: 2px; background: #fff; }
        div { transform: translateX(3px) rotate(90deg);
              font-size: 0; width: 4px }
      </style>
      <div><img src="../resources_test/pattern.png"></div>`)
}

func Test_2dTransform_3(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// A translateX after the rotation is actually a translateY
	assertPixelsEqual(t, `
        ____________
        ____________
        ____________
        ____________
        ____________
        __BBBr______
        __BBBB______
        __BBBB______
        __BBBB______
        ____________
        ____________
        ____________
    `, `
      <style>
        @page { size: 12px; margin: 2px; background: #fff; }
        div { transform: rotate(90deg) translateX(3px);
              font-size: 0; width: 4px }
      </style>
      <div><img src="../resources_test/pattern.png"></div>`)
}

func Test_2dTransform_4(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, `
        ____________
        ____________
        ____________
        ____________
        ____________
        __BBBr______
        __BBBB______
        __BBBB______
        __BBBB______
        ____________
        ____________
        ____________
    `, `
      <style>
        @page { size: 12px; margin: 2px; background: #fff; }
        div { transform: rotate(90deg); font-size: 0; width: 4px }
        img { transform: translateX(3px) }
      </style>
      <div><img src="../resources_test/pattern.png"></div>`)
}

func Test_2dTransform_5(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, `
        ________
        ________
        __BBBr__
        __BBBB__
        __BBBB__
        __BBBB__
        ________
        ________
    `, `
      <style>
        @page { size: 8px; margin: 2px; background: #fff; }
        div { transform: matrix(-1, 0, 0, 1, 0, 0); font-size: 0 }
      </style>
      <div><img src="../resources_test/pattern.png"></div>`)
}

func Test_2dTransform_6(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, `
        ________
        ________
        ________
        ________
        ___rBBB_
        ___BBBB_
        ___BBBB_
        ___BBBB_
    `, `
      <style>
        @page { size: 8px; margin: 2px; background: #fff; }
        div { transform: translate(1px, 2px); font-size: 0 }
      </style>
      <div><img src="../resources_test/pattern.png"></div>`)
}

func Test_2dTransform_7(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, `
        ________
        ________
        ___rBBB_
        ___BBBB_
        ___BBBB_
        ___BBBB_
        ________
        ________
    `, `
      <style>
        @page { size: 8px; margin: 2px; background: #fff; }
        div { transform: translate(25%, 0); font-size: 0 }
      </style>
      <div><img src="../resources_test/pattern.png"></div>`)
}

func Test_2dTransform_8(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, `
        ________
        ________
        _____rBB
        _____BBB
        _____BBB
        _____BBB
        ________
        ________
    `, `
      <style>
        @page { size: 8px; margin: 2px; background: #fff; }
        div { transform: translateX(0.25em); font-size: 12px }
        div div { font-size: 0 }
      </style>
      <div><div><img src="../resources_test/pattern.png"></div></div>`)
}

func Test_2dTransform_9(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, `
        ________
        __rBBB__
        __BBBB__
        __BBBB__
        __BBBB__
        ________
        ________
        ________
    `, `
      <style>
        @page { size: 8px; margin: 2px; background: #fff; }
        div { transform: translateY(-1px); font-size: 0 }
      </style>
      <div><img src="../resources_test/pattern.png"></div>`)
}

func Test_2dTransform_10(t *testing.T) {
	capt := testutils.CaptureLogs()
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
        __________
    `, `
      <style>
        @page { size: 10px; margin: 2px; background: #fff; }
        div { transform: scale(2, 2);
              transform-origin: 1px 1px 1px;
              image-rendering: pixelated;
              font-size: 0 }
      </style>
      <div><img src="../resources_test/pattern.png"></div>`)
}

func Test_2dTransform_11(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, `
        __________
        __rBBB____
        __rBBB____
        __BBBB____
        __BBBB____
        __BBBB____
        __BBBB____
        __BBBB____
        __BBBB____
        __________
    `, `
      <style>
        @page { size: 10px; margin: 2px; background: #fff; }
        div { transform: scale(1, 2);
              transform-origin: 1px 1px;
              image-rendering: pixelated;
              font-size: 0 }
      </style>
      <div><img src="../resources_test/pattern.png"></div>`)
}

func Test_2dTransform_12(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, `
        __________
        __rBBB____
        __rBBB____
        __BBBB____
        __BBBB____
        __BBBB____
        __BBBB____
        __BBBB____
        __BBBB____
        __________
    `, `
      <style>
        @page { size: 10px; margin: 2px; background: #fff; }
        div { transform: scaleY(2);
              transform-origin: 1px 1px 0;
              image-rendering: pixelated;
              font-size: 0 }
      </style>
      <div><img src="../resources_test/pattern.png"></div>`)
}

func Test_2dTransform_13(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, `
        __________
        __________
        _rrBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        _BBBBBBBB_
        __________
        __________
        __________
        __________
    `, `
      <style>
        @page { size: 10px; margin: 2px; background: #fff; }
        div { transform: scaleX(2);
              transform-origin: 1px 1px;
              image-rendering: pixelated;
              font-size: 0 }
      </style>
      <div><img src="../resources_test/pattern.png"></div>`)
}
