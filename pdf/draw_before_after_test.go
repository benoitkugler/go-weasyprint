package pdf

import (
	"testing"

	"github.com/benoitkugler/webrender/utils/testutils"
)

// Test how before and after pseudo elements are drawn.

const tolerance = 2

func TestBeforeAfter1(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertSameRendering(t,
		`
            <style>
                @page { size: 300px 30px }
                body { margin: 0; background: #fff }
                a[href]:before { content: "[" attr(href) "] " }
            </style>
            <p><a href="some url">some content</a></p>
        `, `
            <style>
                @page { size: 300px 30px }
                body { margin: 0; background: #fff }
            </style>
            <p><a href="another url"><span>[some url] </span>some content</p>
        `, tolerance)
}

func TestBeforeAfter2(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertSameRendering(t, `
            <style>
                @page { size: 500px 30px }
                body { margin: 0; background: #fff; quotes: "«" "»" "“" "”" }
                q:before { content: open-quote " "}
                q:after { content: " " close-quote }
            </style>
            <p><q>Lorem ipsum <q>dolor</q> sit amet</q></p>
        `, `
            <style>
                @page { size: 500px 30px }
                body { margin: 0; background: #fff }
                q:before, q:after { content: none }
            </style>
            <p><span><span>« </span>Lorem ipsum
                <span><span>“ </span>dolor<span> ”</span></span>
                sit amet<span> »</span></span></p>
        `, tolerance)
}

func TestBeforeAfter3(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertSameRendering(t, `
            <style>
                @page { size: 100px 30px }
                body { margin: 0; background: #fff; }
                p:before { content: "a" url(../resources_test/pattern.png) "b"}
            </style>
            <p>c</p>
        `, `
            <style>
                @page { size: 100px 30px }
                body { margin: 0; background: #fff }
            </style>
            <p><span>a<img src="../resources_test/pattern.png" alt="Missing image">b</span>c</p>
        `, tolerance)
}
