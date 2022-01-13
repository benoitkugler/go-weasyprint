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

// Test how SVG simple patterns are drawn.

func TestPattern(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "pattern", `
        BBrrBBrr
        BBrrBBrr
        rrBBrrBB
        rrBBrrBB
        BBrrBBrr
        BBrrBBrr
        rrBBrrBB
        rrBBrrBB
    `, `
      <style>
        @page { size: 8px }
        svg { display: block }
      </style>
      <svg width="8px" height="8px" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <pattern id="pat" x="0" y="0" width="4" height="4"
            patternUnits="userSpaceOnUse"
            patternContentUnits="userSpaceOnUse">
            <rect x="0" y="0" width="2" height="2" fill="blue" />
            <rect x="0" y="2" width="2" height="2" fill="red" />
            <rect x="2" y="0" width="2" height="2" fill="red" />
            <rect x="2" y="2" width="2" height="2" fill="blue" />
          </pattern>
        </defs>
        <rect x="0" y="0" width="8" height="8" fill="url(#pat)" />
      </svg>
    `)
}

func TestPattern_2(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "pattern_2", `
        BBrrBBrr
        BBrrBBrr
        rrBBrrBB
        rrBBrrBB
        BBrrBBrr
        BBrrBBrr
        rrBBrrBB
        rrBBrrBB
    `, `
      <style>
        @page { size: 8px }
        svg { display: block }
      </style>
      <svg width="8px" height="8px" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <pattern id="pat" x="0" y="0" width="50%" height="50%"
            patternUnits="objectBoundingBox"
            patternContentUnits="userSpaceOnUse">
            <rect x="0" y="0" width="2" height="2" fill="blue" />
            <rect x="0" y="2" width="2" height="2" fill="red" />
            <rect x="2" y="0" width="2" height="2" fill="red" />
            <rect x="2" y="2" width="2" height="2" fill="blue" />
          </pattern>
        </defs>
        <rect x="0" y="0" width="8" height="8" fill="url(#pat)" />
      </svg>
    `)
}

func TestPattern_3(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "pattern_3", `
        BBrrBBrr
        BBrrBBrr
        rrBBrrBB
        rrBBrrBB
        BBrrBBrr
        BBrrBBrr
        rrBBrrBB
        rrBBrrBB
    `, `
      <style>
        @page { size: 8px }
        svg { display: block }
      </style>
      <svg width="8px" height="8px" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <pattern id="pat" x="0" y="0" width="4" height="4"
            patternUnits="userSpaceOnUse"
            patternContentUnits="userSpaceOnUse">
            <rect x="0" y="0" width="2" height="2" fill="blue" />
            <rect x="0" y="2" width="2" height="2" fill="red" />
            <rect x="2" y="0" width="2" height="2" fill="red" />
            <rect x="2" y="2" width="2" height="2" fill="blue" />
          </pattern>
        </defs>
        <rect x="0" y="0" width="8" height="8" fill="url(#pat)" />
      </svg>
    `)
}

func TestPattern_4(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "pattern_4", `
        BBrrBBrr
        BBrrBBrr
        rrBBrrBB
        rrBBrrBB
        BBrrBBrr
        BBrrBBrr
        rrBBrrBB
        rrBBrrBB
    `, `
      <style>
        @page { size: 8px }
        svg { display: block }
      </style>
      <svg width="8px" height="8px" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <pattern id="pat" x="0" y="0" width="4" height="4"
            patternUnits="userSpaceOnUse"
            patternContentUnits="objectBoundingBox">
            <rect x="0" y="0" width="50%" height="50%" fill="blue" />
            <rect x="0" y="50%" width="50%" height="50%" fill="red" />
            <rect x="50%" y="0" width="50%" height="50%" fill="red" />
            <rect x="50%" y="50%" width="50%" height="50%" fill="blue" />
          </pattern>
        </defs>
        <rect x="0" y="0" width="8" height="8" fill="url(#pat)" />
      </svg>
    `)
}
