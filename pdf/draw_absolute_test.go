package pdf

import (
	"testing"

	tu "github.com/benoitkugler/webrender/utils/testutils"
)

// Test how absolutes are drawn

func TestAbsoluteSplit_1(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        BBBBRRRRRRRR____
        BBBBRRRRRRRR____
        BBBBRR__________
        BBBBRR__________
    `, `
        <style>
            @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
            @page {
                size: 16px 2px;
            }
            body {
                color: red;
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            div.split {
                color: blue;
                left: 0;
                position: absolute;
                top: 0;
                width: 4px;
            }
        </style>
        <div class="split">aa aa</div>
        <div>bbbbbb bbb</div>
    `)
}

func TestAbsoluteSplit_2(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        RRRRRRRRRRRRBBBB
        RRRRRRRRRRRRBBBB
        RRRR________BBBB
        RRRR________BBBB
    `, `
        <style>
            @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
            @page {
                size: 16px 2px;
            }
            body {
                color: red;
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            div.split {
                color: blue;
                position: absolute;
                top: 0;
                right: 0;
                width: 4px;
            }
        </style>
        <div class="split">aa aa</div>
        <div>bbbbbb bb</div>
    `)
}

func TestAbsoluteSplit_3(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        BBBBRRRRRRRR____
        BBBBRRRRRRRR____
        RRRRRRRRRR______
        RRRRRRRRRR______
    `, `
        <style>
            @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
            @page {
                size: 16px 2px;
            }
            body {
                color: red;
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            div.split {
                color: blue;
                position: absolute;
                top: 0;
                left: 0;
                width: 4px;
            }
        </style>
        <div class="split">aa</div>
        <div>bbbbbb bbbbb</div>
    `)
}

func TestAbsoluteSplit_4(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        RRRRRRRRRRRRBBBB
        RRRRRRRRRRRRBBBB
        RRRRRRRRRR______
        RRRRRRRRRR______
    `, `
        <style>
            @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
            @page {
                size: 16px 2px;
            }
            body {
                color: red;
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            div.split {
                color: blue;
                position: absolute;
                top: 0;
                right: 0;
                width: 4px;
            }
        </style>
        <div class="split">aa</div>
        <div>bbbbbb bbbbb</div>
    `)
}

func TestAbsoluteSplit_5(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        BBBBRRRR____gggg
        BBBBRRRR____gggg
        BBBBRRRRRR__gggg
        BBBBRRRRRR__gggg
    `, `
        <style>
            @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
            @page {
                size: 16px 2px;
            }
            body {
                color: red;
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            div.split {
                color: blue;
                position: absolute;
                top: 0;
                left: 0;
                width: 4px;
            }
            div.split2 {
                color: green;
                position: absolute;
                top: 0;
                right: 0;
                width: 4px;
        </style>
        <div class="split">aa aa</div>
        <div class="split2">cc cc</div>
        <div>bbbb bbbbb</div>
    `)
}

func TestAbsoluteSplit_6(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        BBBBRRRR____gggg
        BBBBRRRR____gggg
        BBBBRRRRRR______
        BBBBRRRRRR______
    `, `
        <style>
            @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
            @page {
                size: 16px 2px;
            }
            body {
                color: red;
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            div.split {
                color: blue;
                position: absolute;
                width: 4px;
            }
            div.split2 {
                color: green;
                position: absolute;
                top: 0;
                right: 0;
                width: 4px;
        </style>
        <div class="split">aa aa</div>
        <div class="split2">cc</div>
        <div>bbbb bbbbb</div>
    `)
}

func TestAbsoluteSplit_7(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        BBBBRRRRRRRRgggg
        BBBBRRRRRRRRgggg
        ____RRRR____gggg
        ____RRRR____gggg
    `, `
        <style>
            @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
            @page {
                size: 16px 2px;
            }
            body {
                color: red;
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            div.split {
                color: blue;
                position: absolute;
                width: 4px;
            }
            div.split2 {
                color: green;
                position: absolute;
                top: 0;
                right: 0;
                width: 4px;
            }
            div.push {
                margin-left: 4px;
            }
        </style>
        <div class="split">aa</div>
        <div class="split2">cc cc</div>
        <div class="push">bbbb bb</div>
    `)
}

func TestAbsoluteSplit_8(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        ______
        ______
        ______
        ______
        __RR__
        __RR__
        ______
        ______
    `, `
        <style>
            @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
            @page {
                margin: 2px 0;
                size: 6px 8px;
            }
            body {
                color: red;
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            div {
                position: absolute;
                left: 2px;
                top: 2px;
                width: 2px;
            }
        </style>
        <div>a a a a</div>
    `)
}

func TestAbsoluteSplit_9(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        ______
        ______
        BBRRBB
        BBRRBB
        BBRR__
        BBRR__
        ______
        ______
    `, `
        <style>
            @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
            @page {
                margin: 2px 0;
                size: 6px 8px;
            }
            body {
                color: blue;
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            div {
                color: red;
                position: absolute;
                left: 2px;
                top: 0;
                width: 2px;
            }
        </style>
        aaa a<div>a a a a</div>
    `)
}

func TestAbsoluteSplit_10(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        BB____
        BB____
        __RR__
        __RR__
        __RR__
        __RR__

        BBRR__
        BBRR__
        __RR__
        __RR__
        ______
        ______
    `, `
        <style>
            @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
            @page {
                size: 6px;
            }
            body {
                color: blue;
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            div {
                color: red;
                position: absolute;
                left: 2px;
                top: 2px;
                width: 2px;
            }
            div + article {
                break-before: page;
            }
        </style>
        <article>a</article>
        <div>a a a a</div>
        <article>a</article>
    `)
}

func TestAbsoluteSplit_11(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        BBBBBB
        BBBBBB
        BBRRBB
        BBRRBB
        __RR__
        __RR__
    `, `
        <style>
            @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
            @page {
                size: 6px;
            }
            body {
                color: blue;
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            div {
                bottom: 0;
                color: red;
                position: absolute;
                left: 2px;
                width: 2px;
            }
        </style>
        aaa aaa<div>a a</div>
    `)
}

func TestAbsoluteNextPage(t *testing.T) {
	t.Skip()
	// TODO: currently, the layout of absolute boxes forces to render a box,
	// even when it doesn’t fit in the page. This workaround avoids placeholders
	// with no box. Instead, we should remove these placeholders, or avoid
	// crashes when they’re rendered.
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        RRRRRRRRRR______
        RRRRRRRRRR______
        RRRRRRRRRR______
        RRRRRRRRRR______
        BBBBBBRRRR______
        BBBBBBRRRR______
        BBBBBB__________
        ________________
    `, `
        <style>
            @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
            @page {
                size: 16px 4px;
            }
            body {
                color: red;
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            div.split {
                color: blue;
                position: absolute;
                font-size: 3px;
            }
        </style>
        aaaaa aaaaa
        <div class="split">bb</div>
        aaaaa
    `)
}

func TestAbsoluteRtl_1(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        __________RRRRRR
        __________RRRRRR
        ________________
    `, `
        <style>
            @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
            @page {
                size: 16px 3px;
            }
            body {
                direction: rtl;
            }
            div {
                color: red;
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
                position: absolute;
            }
        </style>
        <div>bbb</div>
    `)
}

func TestAbsoluteRtl_2(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        ________________
        _________RRRRRR_
        _________RRRRRR_
    `, `
        <style>
            @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
            @page {
                size: 16px 3px;
            }
            body {
                direction: rtl;
            }
            div {
                color: red;
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
                padding: 1px;
                position: absolute;
            }
        </style>
        <div>bbb</div>
    `)
}

func TestAbsoluteRtl_3(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        ________________
        RRRRRR__________
        RRRRRR__________
    `, `
        <style>
            @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
            @page {
                size: 16px 3px;
            }
            body {
                direction: rtl;
            }
            div {
                bottom: 0;
                color: red;
                font-family: weasyprint;
                font-size: 2px;
                left: 0;
                line-height: 1;
                position: absolute;
            }
        </style>
        <div>bbb</div>
    `)
}

func TestAbsoluteRtl_4(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        ________________
        _________RRRRRR_
        _________RRRRRR_
    `, `
        <style>
            @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
            @page {
                size: 16px 3px;
            }
            body {
                direction: rtl;
            }
            div {
                color: red;
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
                position: absolute;
                right: 1px;
                top: 1px;
            }
        </style>
        <div>bbb</div>
    `)
}

func TestAbsoluteRtl_5(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        RRRRRR__________
        RRRRRR__________
        ________________
    `, `
        <style>
            @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
            @page {
                size: 16px 3px;
            }
            div {
                color: red;
                direction: rtl;
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
                position: absolute;
            }
        </style>
        <div>bbb</div>
    `)
}

func TestAbsolutePagesCounter(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        ______
        _RR___
        _RR___
        _RR___
        _RR___
        _____B
        ______
        _RR___
        _RR___
        _BB___
        _BB___
        _____B
    `, `
        <style>
            @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
            @page {
                font-family: weasyprint;
                margin: 1px;
                size: 6px 6px;
                @bottom-right-corner {
                    color: blue;
                    content: counter(pages);
                    font-size: 1px;
                }
            }
            body {
                color: red;
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
            }
            div {
                color: blue;
                position: absolute;
            }
        </style>
        a a a <div>a a</div>
    `)
}

func TestAbsolutePagesCounterOrphans(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        ______
        _RR___
        _RR___
        _RR___
        _RR___
        ______
        ______
        ______
        _____B
        ______
        _RR___
        _RR___
        _BB___
        _BB___
        _GG___
        _GG___
        ______
        _____B
    `, `
        <style>
            @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
            @page {
                font-family: weasyprint;
                margin: 1px;
                size: 6px 9px;
                @bottom-right-corner {
                    color: blue;
                    content: counter(pages);
                    font-size: 1px;
                }
            }
            body {
                color: red;
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
                orphans: 2;
                widows: 2;
            }
            div {
                color: blue;
                position: absolute;
            }
            div ~ div {
                color: lime;
            }
        </style>
        a a a <div>a a a</div> a <div>a a a</div>
    `)
}

func TestAbsoluteInInline(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        ______
        _GG___
        _GG___
        _GG___
        _GG___
        ______
        ______
        ______
        ______

        ______
        _RR___
        _RR___
        _RR___
        _RR___
        _BB___
        _BB___
        ______
        ______
    `, `
        <style>
            @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
            @page {
                margin: 1px;
                size: 6px 9px;
            }
            body {
                color: red;
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
                orphans: 2;
                widows: 2;
            }
            p {
                color: lime;
            }
            div {
                color: blue;
                position: absolute;
            }
        </style>
        <p>a a</p> a a <div>a</div>
    `)
}

func TestFixedInInline(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        ______
        _GG___
        _GG___
        _GG___
        _GG___
        _BB___
        _BB___
        ______
        ______

        ______
        _RR___
        _RR___
        _RR___
        _RR___
        _BB___
        _BB___
        ______
        ______
    `, `
        <style>
            @font-face {src: url(../resources_test/weasyprint.otf); font-family: weasyprint}
            @page {
                margin: 1px;
                size: 6px 9px;
            }
            body {
                color: red;
                font-family: weasyprint;
                font-size: 2px;
                line-height: 1;
                orphans: 2;
                widows: 2;
            }
            p {
                color: lime;
            }
            div {
                color: blue;
                position: fixed;
            }
        </style>
        <p>a a</p> a a <div>a</div>
    `)
}

func TestAbsoluteImageBackground(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, `
        ____
        _RBB
        _BBB
        _BBB
    `, `
        <style>
          @page {
            size: 4px;
          }
          img {
            background: blue;
            position: absolute;
            top: 1px;
            left: 1px;
          }
        </style>
        <img src="../resources_test/pattern-transparent.svg" />
    `)
}
