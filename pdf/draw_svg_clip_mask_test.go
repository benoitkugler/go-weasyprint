package pdf

import (
	"testing"

	"github.com/benoitkugler/webrender/utils/testutils"
)

// Test clip-path attribute.

func TestClipPath(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, `
        _________
        _________
        __RRRRR__
        __RBBBR__
        __RBBBR__
        __RBBBR__
        __RRRRR__
        _________
        _________
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <clipPath id="clip">
            <rect x="2" y="2" width="5" height="5" />
          </clipPath>
        </defs>
        <rect x="2" y="2" width="5" height="5" stroke-width="2"
              stroke="red" fill="blue" clip-path="url(#clip)" />
      </svg>
    `)
}

func TestClipPathOnGroup(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, `
        _________
        _________
        __BBBB___
        __BRRRR__
        __BRRRR__
        __BRRRR__
        ___RRRR__
        _________
        _________
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <clipPath id="clip">
            <rect x="2" y="2" width="5" height="5" />
          </clipPath>
        </defs>
        <g clip-path="url(#clip)">
          <rect x="1" y="1" width="5" height="5" fill="blue" />
          <rect x="3" y="3" width="5" height="5" fill="red" />
        </g>
      </svg>
    `)
}

func TestClipPathGroupOnGroup(t *testing.T) {
	// TODO: pytest.mark.xfail

	t.Skip()

	// capt := testutils.CaptureLogs()
	// defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, `
        _________
        _________
        __BB_____
        __BR_____
        _________
        _____RR__
        _____RR__
        _________
        _________
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <clipPath id="clip">
            <rect x="2" y="2" width="2" height="2" />
            <rect x="3" y="3" width="2" height="2" />
          </clipPath>
        </defs>
        <g clip-path="url(#clip)">
          <rect x="1" y="1" width="5" height="5" fill="blue" />
          <rect x="3" y="3" width="5" height="5" fill="red" />
        </g>
      </svg>
    `)
}
