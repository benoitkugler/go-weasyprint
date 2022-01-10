package pdf

import (
	"testing"

	"github.com/benoitkugler/webrender/utils/testutils"
)

func TestUse(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "use", `
        RRRRR_____
        RRRRR_____
        __________
        ___RRRRR__
        ___RRRRR__
        __________
        _____RRRRR
        _____RRRRR
        __________
        __________
    `, `
      <style>
        @page { size: 10px }
        svg { display: block }
      </style>
      <svg width="10px" height="10px" xmlns="http://www.w3.org/2000/svg"
           xlink="http://www.w3.org/1999/xlink">
        <defs>
          <rect id="rectangle" width="5" height="2" fill="red" />
        </defs>
        <use xlink:href="#rectangle" />
        <use xlink:href="#rectangle" x="3" y="3" />
        <use xlink:href="#rectangle" x="5" y="6" />
      </svg>
    `)
}
