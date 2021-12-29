package pdf

import (
	"fmt"
	"strings"
	"testing"

	"github.com/benoitkugler/webrender/utils/testutils"
)

// Test how lists are drawn.

var sansFonts = strings.Join([]string{"DejaVu Sans", "sans"}, " ")

func TestListStyleImage(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	doTest := func(position, pixels string) {
		assertPixelsEqual(t, "list_style_image_"+position, pixels, fmt.Sprintf(`
		<style>
		  @page { size: 12px 10px }
		  body { margin: 0; background: white; font-family: %s }
		  ul { margin: 2px 2px 0 7px; list-style: url(../resources_test/pattern.png) %s;
			   font-size: 2px }
		</style>
		<ul><li></li></ul>`, sansFonts, position))
	}

	doTest("outside",
		// ++++++++++++++      ++++  <li> horizontal margins: 7px 2px
		//               ######      <li> width: 12 - 7 - 2 = 3px
		//             --            list marker margin: 0.5em = 2px
		//     ********              list marker image is 4px wide
		`
        ____________
        ____________
        ___rBBB_____
        ___BBBB_____
        ___BBBB_____
        ___BBBB_____
        ____________
        ____________
        ____________
        ____________
     `)
	doTest("inside",
		//  ++++++++++++++      ++++  <li> horizontal margins: 7px 2px
		//                ######      <li> width: 12 - 7 - 2 = 3px
		//                ********    list marker image is 4px wide: overflow
		`
        ____________
        ____________
        _______rBBB_
        _______BBBB_
        _______BBBB_
        _______BBBB_
        ____________
        ____________
        ____________
        ____________
     `)
}

func TestListStyleImageNone(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "list_style_none", `
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
    `, fmt.Sprintf(`
      <style>
        @page { size: 10px }
        body { margin: 0; background: white; font-family: %s }
        ul { margin: 0 0 0 5px; list-style: none; font-size: 2px; }
      </style>
      <ul><li>`, sansFonts))
}
