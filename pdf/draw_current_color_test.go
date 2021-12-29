package pdf

import (
	"testing"

	"github.com/benoitkugler/webrender/utils/testutils"
)

// Test the currentColor value.

const green2x2 = `
GG
GG
`

func TestCurrentColor1(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "backgroundCurrentColor", green2x2, `
      <style>
        @page { size: 2px }
        html, body { height: 100%; margin: 0 }
        html { color: red; background: currentColor }
        body { color: lime; background: inherit }
      </style>
      <body>`)
}

func TestCurrentColor2(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "borderCurrentColor", green2x2, `
      <style>
        @page { size: 2px }
        html { color: red; border-color: currentColor }
        body { color: lime; border: 1px solid; border-color: inherit;
               margin: 0 }
      </style>
      <body>`)
}

func TestCurrentColor3(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "outlineCurrentColor", green2x2, `
      <style>
        @page { size: 2px }
        html { color: red; outline-color: currentColor }
        body { color: lime; outline: 1px solid; outline-color: inherit;
               margin: 1px }
      </style>
      <body>`)
}

func TestCurrentColor4(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "borderCollapseCurrentColor", green2x2, `
      <style>
        @page { size: 2px }
        html { color: red; border-color: currentColor; }
        body { margin: 0 }
        table { border-collapse: collapse;
                color: lime; border: 1px solid; border-color: inherit }
      </style>
      <table><td>`)
}
