package pdf

import (
	"testing"

	"github.com/benoitkugler/webrender/utils/testutils"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

// Test the currentColor value.

const green2x2 = `
GG
GG
`

func TestCurrentColor1(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, green2x2, `
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

	assertPixelsEqual(t, green2x2, `
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

	assertPixelsEqual(t, green2x2, `
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

	assertPixelsEqual(t, green2x2, `
      <style>
        @page { size: 2px }
        html { color: red; border-color: currentColor; }
        body { margin: 0 }
        table { border-collapse: collapse;
                color: lime; border: 1px solid; border-color: inherit }
      </style>
      <table><td>`)
}

func TestCurrentColorSvg_1(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, "KK\nKK", `
      <style>
        @page { size: 2px }
        svg { display: block }
      </style>
      <svg xmlns="http://www.w3.org/2000/svg"
           width="2" height="2" fill="currentColor">
        <rect width="2" height="2"></rect>
      </svg>`)
}

func TestCurrentColorSvg_2(t *testing.T) {
	t.Skip()
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, "GG\nGG", `
      <style>
        @page { size: 2px }
        svg { display: block }
        body { color: lime }
      </style>
      <svg xmlns="http://www.w3.org/2000/svg"
           width="2" height="2">
        <rect width="2" height="2" fill="currentColor"></rect>
      </svg>`)
}

func TestCurrentColorVariable(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	// Regression test for https://github.com/Kozea/WeasyPrint/issues/2010
	assertPixelsEqual(t, "GG\nGG", `
      <style>
        @font-face { src: url(../resources_test/weasyprint.otf); font-family: weasyprint }
        @page { size: 2px }
        html { color: lime; font-family: weasyprint; --var: currentColor }
        div { color: var(--var); font-size: 2px; line-height: 1 }
      </style>
      <div>aa`)
}

func TestCurrentColorVariableBorder(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	// Regression test for https://github.com/Kozea/WeasyPrint/issues/2010
	assertPixelsEqual(t, "GG\nGG", `
      <style>
        @page { size: 2px }
        html { color: lime; --var: currentColor }
        div { color: var(--var); width: 0; height: 0; border: 1px solid }
      </style>
      <div>`)
}
