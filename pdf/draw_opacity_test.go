package pdf

import (
	"fmt"
	"testing"

	"github.com/benoitkugler/webrender/utils/testutils"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

// Test opacity.

const opacitySource = `
    <style>
        @page { size: 60px 60px }
        body { margin: 0; background: #fff }
        div { background: #000; width: 20px; height: 20px }
    </style>
    %s`

func TestOpacityZero(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertSameRendering(t,
		fmt.Sprintf(opacitySource, `<div></div>`),
		fmt.Sprintf(opacitySource, `<div></div><div style="opacity: 0"></div>`), 0)
	assertSameRendering(t,
		fmt.Sprintf(opacitySource, `<div></div>`),
		fmt.Sprintf(opacitySource, `<div></div><div style="opacity: 0%"></div>`), 0)
}

func TestOpacityNormalRange(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertSameRendering(t,
		fmt.Sprintf(opacitySource, `<div style="background: rgb(102, 102, 102)"></div>`),
		fmt.Sprintf(opacitySource, `<div style="opacity: 0.6"></div>`), 0)
	assertSameRendering(t,
		fmt.Sprintf(opacitySource, `<div style="background: rgb(102, 102, 102)"></div>`),
		fmt.Sprintf(opacitySource, `<div style="opacity: 60%"></div>`), 0)
	assertSameRendering(t,
		fmt.Sprintf(opacitySource, `<div style="background: rgb(102, 102, 102)"></div>`),
		fmt.Sprintf(opacitySource, `<div style="opacity: 60.0%"></div>`), 0)
}

func TestOpacityNested(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertSameRendering(t, fmt.Sprintf(opacitySource, `
            <div style="background: rgb(102, 102, 102)"></div>
        `), fmt.Sprintf(opacitySource, `
            <div style="opacity: 0.6"></div>
        `), 0)

	assertSameRendering(t, fmt.Sprintf(opacitySource, `
            <div style="opacity: 0.6"></div>
        `), fmt.Sprintf(opacitySource, `
            <div style="background: none; opacity: 0.666666">
                <div style="opacity: 0.9"></div>
            </div>
        `), 0) //  0.9 * 0.666666 == 0.6
}

func TestOpacityPercentClampDown(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertSameRendering(t,
		fmt.Sprintf(opacitySource, `<div></div>`),
		fmt.Sprintf(opacitySource, `<div style="opacity: 1.2"></div>`), 0)
	assertSameRendering(t,
		fmt.Sprintf(opacitySource, `<div></div>`),
		fmt.Sprintf(opacitySource, `<div style="opacity: 120%"></div>`), 0)
}

func TestOpacityPercentClampUp(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertSameRendering(t,
		fmt.Sprintf(opacitySource, `<div></div>`),
		fmt.Sprintf(opacitySource, `<div></div><div style="opacity: -0.2"></div>`), 0)
	assertSameRendering(t,
		fmt.Sprintf(opacitySource, `<div></div>`),
		fmt.Sprintf(opacitySource, `<div></div><div style="opacity: -20%"></div>`), 0)
}
