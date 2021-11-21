package pdf

import (
	"fmt"
	"testing"

	"github.com/benoitkugler/go-weasyprint/utils/testutils"
)

// Test opacity.

const opacitySource = `
    <style>
        @page { size: 60px 60px }
        body { margin: 0; background: #fff }
        div { background: #000; width: 20px; height: 20px }
    </style>
    %s`

func TestOpacity_1(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertSameRendering(t, "opacity_0_reference", fmt.Sprintf(opacitySource, `
            <div></div>
        `), fmt.Sprintf(opacitySource, `
            <div></div>
            <div style="opacity: 0"></div>
        `), 0)
}

func TestOpacity_2(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertSameRendering(t, "opacity_color_2", fmt.Sprintf(opacitySource, `
            <div style="background: rgb(102, 102, 102)"></div>
        `), fmt.Sprintf(opacitySource, `
            <div style="opacity: 0.6"></div>
        `), 0)
}

func TestOpacity_3(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertSameRendering(t, "opacity_multiplied_reference", fmt.Sprintf(opacitySource, `
            <div style="background: rgb(102, 102, 102)"></div>
        `), fmt.Sprintf(opacitySource, `
            <div style="opacity: 0.6"></div>
        `), 0)

	assertSameRendering(t, "opacity_multiplied_2", fmt.Sprintf(opacitySource, `
            <div style="opacity: 0.6"></div>
        `), fmt.Sprintf(opacitySource, `
            <div style="background: none; opacity: 0.666666">
                <div style="opacity: 0.9"></div>
            </div>
        `), 0) //  0.9 * 0.666666 == 0.6
}
