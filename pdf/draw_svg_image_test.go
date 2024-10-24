package pdf

import (
	"fmt"
	"testing"

	tu "github.com/benoitkugler/webrender/utils/testutils"
)

// Test how images are drawn in SVG.

func TestImageSvg(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        ____
        ____
        __B_
        ____
    `, `
      <style>
        @page { size: 4px 4px }
        svg { display: block }
      </style>
      <svg width="4px" height="4px" xmlns="http://www.w3.org/2000/svg">
        <svg x="1" y="1" width="2" height="2" viewBox="0 0 10 10">
          <rect x="5" y="5" width="5" height="5" fill="blue" />
        </svg>
      </svg>
    `)
}

func TestImageSvgViewbox(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        ____
        ____
        __B_
        ____
    `, `
      <style>
        @page { size: 4px 4px }
        svg { display: block }
      </style>
      <svg viewBox="0 0 4 4" xmlns="http://www.w3.org/2000/svg">
        <svg x="1" y="1" width="2" height="2" viewBox="10 10 10 10">
          <rect x="15" y="15" width="5" height="5" fill="blue" />
        </svg>
      </svg>
    `)
}

func TestImageSvgAlignDefault(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        __BRRR__
        __BRRR__
        __RRRG__
        __RRRG__
        ________
        ________
        ________
        ________
    `, `
      <style>
        @page { size: 8px 8px }
        svg { display: block }
      </style>
      <svg width="8px" height="4px" viewBox="0 0 4 4"
           xmlns="http://www.w3.org/2000/svg">
        <rect width="4" height="4" fill="red" />
        <rect width="1" height="2" fill="blue" />
        <rect x="3" y="2" width="1" height="2" fill="lime" />
      </svg>
    `)
}

func TestImageSvgAlignNone(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        BBRRRRRR
        BBRRRRRR
        RRRRRRGG
        RRRRRRGG
        ________
        ________
        ________
        ________
    `, `
      <style>
        @page { size: 8px 8px }
        svg { display: block }
      </style>
      <svg width="8px" height="4px" viewBox="0 0 4 4"
           preserveAspectRatio="none"
           xmlns="http://www.w3.org/2000/svg">
        <rect width="4" height="4" fill="red" />
        <rect width="1" height="2" fill="blue" />
        <rect x="3" y="2" width="1" height="2" fill="lime" />
      </svg>
    `)
}

func TestImageSvgAlignMeetX(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        ____BRRR
        ____BRRR
        ____RRRG
        ____RRRG
        ________
        ________
        ________
        ________
    `, `
      <style>
        @page { size: 8px 8px }
        svg { display: block }
      </style>
      <svg width="8px" height="4px" viewBox="0 0 4 4"
           preserveAspectRatio="xMaxYMax meet"
           xmlns="http://www.w3.org/2000/svg">
        <rect width="4" height="4" fill="red" />
        <rect width="1" height="2" fill="blue" />
        <rect x="3" y="2" width="1" height="2" fill="lime" />
      </svg>
    `)
}

func TestImageSvgAlignMeetY(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        ________
        ________
        ________
        ________
        BRRR____
        BRRR____
        RRRG____
        RRRG____
    `, `
      <style>
        @page { size: 8px 8px }
        svg { display: block }
      </style>
      <svg width="4px" height="8px" viewBox="0 0 4 4"
           preserveAspectRatio="xMaxYMax meet"
           xmlns="http://www.w3.org/2000/svg">
        <rect width="4" height="4" fill="red" />
        <rect width="1" height="2" fill="blue" />
        <rect x="3" y="2" width="1" height="2" fill="lime" />
      </svg>
    `)
}

func TestImageSvgAlignSliceX(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        BBRRRRRR
        BBRRRRRR
        BBRRRRRR
        BBRRRRRR
        ________
        ________
        ________
        ________
    `, `
      <style>
        @page { size: 8px 8px }
        svg { display: block; overflow: hidden }
      </style>
      <svg width="8px" height="4px" viewBox="0 0 4 4"
           preserveAspectRatio="xMinYMin slice"
           xmlns="http://www.w3.org/2000/svg">
        <rect width="4" height="4" fill="red" />
        <rect width="1" height="2" fill="blue" />
        <rect x="3" y="2" width="1" height="2" fill="lime" />
      </svg>
    `)
}

func TestImageSvgAlignSliceY(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        BBRR____
        BBRR____
        BBRR____
        BBRR____
        RRRR____
        RRRR____
        RRRR____
        RRRR____
    `, `
      <style>
        @page { size: 8px 8px }
        svg { display: block; overflow: hidden }
      </style>
      <svg width="4px" height="8px" viewBox="0 0 4 4"
           preserveAspectRatio="xMinYMin slice"
           xmlns="http://www.w3.org/2000/svg">
        <rect width="4" height="4" fill="red" />
        <rect width="1" height="2" fill="blue" />
        <rect x="3" y="2" width="1" height="2" fill="lime" />
      </svg>
    `)
}

func TestImageSvgPercentage(t *testing.T) {
	t.Skip()
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        ____
        ____
        __B_
        ____
    `, `
      <style>
        @page { size: 4px 4px }
        svg { display: block }
      </style>
      <svg width="100%" height="100%" xmlns="http://www.w3.org/2000/svg">
        <svg x="1" y="1" width="50%" height="50%" viewBox="0 0 10 10">
          <rect x="5" y="5" width="5" height="5" fill="blue" />
        </svg>
      </svg>
    `)
}

func TestImageSvgWrong(t *testing.T) {
	assertPixelsEqual(t, `
        ____
        ____
  		____
        ____
    `, `
      <style>
        @page { size: 4px 4px }
        svg { display: block }
      </style>
      <svg width="4px" height="4px" xmlns="http://www.w3.org/2000/svg">
        <That’s bad!
      </svg>
    `)
}

func TestImageImage(t *testing.T) {
	defer tu.CaptureLogs().AssertNoLogs(t)
	assertPixelsEqual(t, `
        rBBB
        BBBB
        BBBB
        BBBB
    `, fmt.Sprintf(`
      <style>
        @page { size: 4px 4px }
        svg { display: block }
      </style>
      <svg width="4px" height="4px" xmlns="http://www.w3.org/2000/svg">
        <image xlink:href="%s" />
      </svg>
    `, "../resources_test/pattern.png"))
}

func TestImageImageWrong(t *testing.T) {
	assertPixelsEqual(t, `
        ____
        ____
        ____
        ____
    `, `
      <style>
        @page { size: 4px 4px }
        svg { display: block }
      </style>
      <svg width="4px" height="4px" xmlns="http://www.w3.org/2000/svg">
        <image xlink:href="it doesn’t exist, mouhahahaha" />
      </svg>
    `)
}
