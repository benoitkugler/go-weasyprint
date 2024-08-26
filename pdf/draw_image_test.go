package pdf

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/benoitkugler/webrender/utils/testutils"
	tu "github.com/benoitkugler/webrender/utils/testutils"
)

// Test how images are drawn.

const (
	centeredImage = `
    ________
    ________
    __rBBB__
    __BBBB__
    __BBBB__
    __BBBB__
    ________
    ________
`

	centeredImageOverflow = `
    ________
    ________
    __rBBBBB
    __BBBBBB
    __BBBBBB
    __BBBBBB
    __BBBBBB
    __BBBBBB
`

	resizedImage = `
    ____________
    ____________
    __rrBBBBBB__
    __rrBBBBBB__
    __BBBBBBBB__
    __BBBBBBBB__
    __BBBBBBBB__
    __BBBBBBBB__
    __BBBBBBBB__
    __BBBBBBBB__
    ____________
    ____________
`

	smallResizedImage = `
    ____________
    ____________
    __rBBB______
    __BBBB______
    __BBBB______
    __BBBB______
    ____________
    ____________
    ____________
    ____________
    ____________
    ____________
`

	blueImage = `
    ________
    ________
    __aaaa__
    __aaaa__
    __aaaa__
    __aaaa__
    ________
    ________
`

	noImage = `
    ________
    ________
    ________
    ________
    ________
    ________
    ________
    ________
`

	pageBreak = `
    ________
    ________
    __rBBB__
    __BBBB__
    __BBBB__
    __BBBB__
    ________
    ________

    ________
    ________
    ________
    ________
    ________
    ________
    ________
    ________

    ________
    ________
    __rBBB__
    __BBBB__
    __BBBB__
    __BBBB__
    ________
    ________
`

	table = `
    ________
    ________
    __rBBB__
    __BBBB__
    __BBBB__
    __BBBB__
    ________
    ________

    __rBBB__
    __BBBB__
    __BBBB__
    __BBBB__
    ________
    ________
    ________
    ________
`

	coverImage = `
    ________
    ________
    __BB____
    __BB____
    __BB____
    __BB____
    ________
    ________
`

	borderImage = `
    ________
    _GGGGGG_
    _GrBBBG_
    _GBBBBG_
    _GBBBBG_
    _GBBBBG_
    _GGGGGG_
    ________
`
)

func TestImages(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, data := range [][2]string{
		{"../resources_test/pattern.svg", centeredImage},
		{"../resources_test/pattern.png", centeredImage},
		{"../resources_test/pattern.palette.png", centeredImage},
		{"../resources_test/pattern.gif", centeredImage},
		{"../resources_test/blue.jpg", blueImage},
	} {
		filename, image := data[0], data[1]
		assertPixelsEqual(t, image, fmt.Sprintf(`
      <style>
        @page { size: 8px }
        body { margin: 2px 0 0 2px; background: #fff; font-size: 0 }
		img { overflow: hidden }
      </style>
      <div><img src="%s"></div>`, filename))
	}
}

func TestResizedImages(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	for _, filename := range [...]string{
		"../resources_test/pattern.svg",
		"../resources_test/pattern.png",
		"../resources_test/pattern.palette.png",
		"../resources_test/pattern.gif",
	} {
		assertPixelsEqual(t, resizedImage, fmt.Sprintf(`
      <style>
        @page { size: 12px }
        body { margin: 2px 0 0 2px; background: #fff; font-size: 0 }
        img { display: block; width: 8px; image-rendering: pixelated; overflow: hidden }
      </style>
      <div><img src="%s"></div>`, filename))
	}
}

func TestImageOverflow(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, centeredImageOverflow, `
      <style>
        @page { size: 8px }
        body { margin: 2px 0 0 2px; font-size: 0 }
      </style>
      <div><img src="../resources_test/pattern.svg"></div>`)
}

func TestSvgSizing(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, data := range []struct {
		viewbox       string
		width, height int
	}{
		{"", 0, 0},
		{"", 4, 0},
		{"", 0, 4},
		{"", 4, 4},
		{"0 0 4 4", 4, 0},
		{"0 0 4 4", 0, 4},
		{"0 0 4 4", 4, 4},
		{"0 0 4 4", 4, 4},
	} {
		var viewbox, width, height string
		if data.viewbox != "" {
			viewbox = fmt.Sprintf(`viewbox="%s"`, data.viewbox)
		}
		if data.width != 0 {
			width = fmt.Sprintf(`width="%d"`, +data.width)
		}
		if data.height != 0 {
			height = fmt.Sprintf(`height="%dpx"`, +data.height)
		}
		assertPixelsEqual(t,
			centeredImage, fmt.Sprintf(`
      <style>
        @page { size: 8px }
        body { margin: 2px 0 0 2px; background: #fff; font-size: 0 }
        svg { display: block }
      </style>
      <svg %s %s %s>
        <rect width="4px" height="4px" fill="#00f" />
        <rect width="1px" height="1px" fill="#f00" />
      </svg>`, viewbox, width, height))
	}
}

func TestSvgResizing(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, data := range []struct {
		viewbox       string
		width, height int
		image         string
	}{
		{"", 0, 0, smallResizedImage},
		{"", 8, 0, smallResizedImage},
		{"", 0, 8, smallResizedImage},
		{"", 8, 8, smallResizedImage},
		{"0 0 4 4", 0, 0, resizedImage},
		{"0 0 4 4", 8, 0, resizedImage},
		{"0 0 4 4", 0, 8, resizedImage},
		{"0 0 4 4", 8, 8, resizedImage},
		{"0 0 4 4", 800, 800, resizedImage},
	} {
		var viewbox, width, height string
		if data.viewbox != "" {
			viewbox = fmt.Sprintf(`viewbox="%s"`, data.viewbox)
		}
		if data.width != 0 {
			width = fmt.Sprintf(`width="%d"`, data.width)
		}
		if data.height != 0 {
			height = fmt.Sprintf(`height="%dpx"`, data.height)
		}
		assertPixelsEqual(t, data.image, fmt.Sprintf(`
      <style>
        @page { size: 12px }
        body { margin: 2px 0 0 2px; background: #fff; font-size: 0 }
        svg { display: block; width: 8px }
      </style>
      <svg %s %s %s>
        <rect width="4" height="4" fill="#00f" />
        <rect width="1" height="1" fill="#f00" />
      </svg>`, viewbox, width, height))
	}
}

func TestImagesBlock(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, centeredImage, `
      <style>
        @page { size: 8px }
        body { margin: 0; background: #fff; font-size: 0 }
        img { display: block; margin: 2px auto 0 }
      </style>
      <div><img src="../resources_test/pattern.png"></div>`)
}

func TestImagesNotFound(t *testing.T) {
	capt := testutils.CaptureLogs()
	assertPixelsEqual(t, noImage, `
          <style>
            @page { size: 8px }
            body { margin: 0; background: #fff; font-size: 0 }
            img { display: block; margin: 2px auto 0 }
          </style>
          <div><img src="inexistent1.png" alt=""></div>`)
	logs := capt.Logs()
	testutils.AssertEqual(t, len(logs), 1, "")
	if !(strings.Contains(logs[0], "failed to load image") && strings.Contains(logs[0], "inexistent1.png")) {
		t.Fatalf("unexpected log: %s", logs[0])
	}
}

func TestImagesNoSrc(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, noImage, `
      <style>
        @page { size: 8px }
        body { margin: 0; background: #fff; font-size: 0 }
        img { display: block; margin: 2px auto 0 }
      </style>
      <div><img alt=""></div>`)
}

func TestImagesAlt(t *testing.T) {
	capt := testutils.CaptureLogs()

	base := `<style>
		@page { size: 200px 30px }
		body { margin: 0; background: #fff; font-size: 0 }
	</style>
	<div>%s</div>`

	input0 := fmt.Sprintf(base, `Hello, world!`)
	input1 := fmt.Sprintf(base, `<img src="inexistent2.png" alt="Hello, world!">`)
	input2 := fmt.Sprintf(base, `<img alt="Hello, world!">`)
	input3 := fmt.Sprintf(base, `<img src="data:image/svg+xml,<svg></svg>" alt="Hello, world!">`)

	assertSameRendering(t, input0, input1, 0)
	assertSameRendering(t, input0, input2, 0)
	assertSameRendering(t, input0, input3, 0)

	logs := capt.Logs()
	testutils.AssertEqual(t, len(logs), 1, "")
	if !(strings.Contains(logs[0], "failed to load image") &&
		strings.Contains(logs[0], "inexistent2.png")) {
		t.Fatalf("unexpected log: %s", logs[0])
	}
}

func TestImagesRepeatTransparent(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/1440
	assertPixelsEqual(t, "_\n_\n_", `
      <style>
        @page { size: 1px }
        div { height: 100px; width: 100px; background: url(../resources_test/logo_small.png) }
      </style>
      <div></div><div></div><div></div>`)
}

func TestImagesNoWidth(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, noImage, `
      <style>
        @page { size: 8px }
        body { margin: 2px; background: #fff; font-size: 0 }
      </style>
      <div><img src="../resources_test/pattern.png" alt="not shown"
                style="width: 0; height: 1px"></div>`)
}

func TestImagesNoHeight(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, noImage, `
      <style>
        @page { size: 8px }
        body { margin: 2px; background: #fff; font-size: 0 }
      </style>
      <div><img src="../resources_test/pattern.png" alt="not shown"
                style="width: 1px; height: 0"></div>`)
}

func TestImagesNoWidthHeight(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, noImage, `
      <style>
        @page { size: 8px }
        body { margin: 2px; background: #fff; font-size: 0 }
      </style>
      <div><img src="../resources_test/pattern.png" alt="not shown"
                style="width: 0; height: 0"></div>`)
}

func TestImagesPageBreak(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, pageBreak, `
      <style>
        @page { size: 8px; margin: 2px; background: #fff }
        body { font-size: 0 }
      </style>
      <div><img src="../resources_test/pattern.png"></div>
      <div style="page-break-before: right"><img src="../resources_test/pattern.png"></div>`)
}

func TestImageRepeatInline(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/808
	assertPixelsEqual(t, table, `
      <style>
        @page { size: 8px; margin: 0; background: #fff }
        table { border-collapse: collapse; margin: 2px }
        th, td { border: none; padding: 0 }
        th { height: 4px; line-height: 4px }
        td { height: 2px }
        img { vertical-align: top }
      </style>
      <table>
        <thead>
          <tr><th><img src="../resources_test/pattern.png"></th></tr>
        </thead>
        <tbody>
          <tr><td></td></tr>
          <tr><td></td></tr>
        </tbody>
      </table>`)
}

func TestImageRepeatBlock(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/808
	assertPixelsEqual(t, table, `
      <style>
        @page { size: 8px; margin: 0; background: #fff }
        table { border-collapse: collapse; margin: 2px }
        th, td { border: none; padding: 0 }
        th { height: 4px }
        td { height: 2px }
        img { display: block }
      </style>
      <table>
        <thead>
          <tr><th><img src="../resources_test/pattern.png"></th></tr>
        </thead>
        <tbody>
          <tr><td></td></tr>
          <tr><td></td></tr>
        </tbody>
      </table>`)
}

func TestImagesPadding(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	// Regression test: padding used to be ignored on images
	assertPixelsEqual(t, centeredImage, `
      <style>
        @page { size: 8px; background: #fff }
        body { font-size: 0 }
      </style>
      <div style="line-height: 1px">
        <img src=../resources_test/pattern.png style="padding: 2px 0 0 2px">
      </div>`)
}

func TestImagesInInlineBlock(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	// Regression test: this used to cause an exception
	assertPixelsEqual(t, centeredImage, `
      <style>
        @page { size: 8px }
        body { margin: 2px 0 0 2px; background: #fff; font-size: 0 }
      </style>
      <div style="display: inline-block">
        <p><img src=../resources_test/pattern.png></p>
      </div>`)
}

func TestImagesTransparentText(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	// Test regression: https://github.com/Kozea/WeasyPrint/issues/2131
	assertPixelsEqual(t, centeredImage, `<style>
        @page { size: 8px }
        body { margin: 2px 0 0 2px; font-size: 2px; line-height: 0 }
      </style>
      <div style="color: #0001">123</div>
      <img src="../resources_test/pattern.png">
    `)
}

func TestImagesSharedPattern(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	// The same image is used in a repeating background,
	// then in a non-repating <img>.
	// If Pattern objects are shared carelessly, the image will be repeated.
	assertPixelsEqual(t, `
        ____________
        ____________
        __aaaaaaaa__
        __aaaaaaaa__
        ____________
        __aaaa______
        __aaaa______
        __aaaa______
        __aaaa______
        ____________
        ____________
        ____________
    `, `
      <style>
        @page { size: 12px }
        body { margin: 2px; background: #fff; font-size: 0 }
      </style>
      <div style="background: url(../resources_test/blue.jpg);
                  height: 2px; margin-bottom: 1px"></div>
      <img src=../resources_test/blue.jpg>
    `)
}

func TestImageResolution(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	ref := `
            <style>@page { size: 20px; margin: 2px; background: #fff }</style>
            <div style="font-size: 0">
                <img src="../resources_test/pattern.png" style="width: 8px"></div>
        `
	assertSameRendering(t, ref, `
            <style>@page { size: 20px; margin: 2px; background: #fff }</style>
            <div style="image-resolution: .5dppx; font-size: 0">
                <img src="../resources_test/pattern.png"></div>
        `, 0)
	assertSameRendering(t, ref, `
            <style>@page { size: 20px; margin: 2px; background: #fff }
                   div::before { content: url(../resources_test/pattern.png) }
            </style>
            <div style="image-resolution: .5dppx; font-size: 0"></div>
        `, 0)
	assertSameRendering(t, ref, `
            <style>@page { size: 20px; margin: 2px; background: #fff }
            </style>
            <div style="height: 16px; image-resolution: .5dppx;
                        background: url(../resources_test/pattern.png) no-repeat"></div>
        `, 0)
}

func TestImageCover(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, coverImage, `
      <style>
        @page { size: 8px }
        body { margin: 2px 0 0 2px; background: #fff; font-size: 0 }
        img { object-fit: cover; height: 4px; width: 2px; overflow: hidden }
      </style>
      <img src="../resources_test/pattern.png">`)
}

func TestImageContain(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, centeredImage, `
      <style>
        @page { size: 8px }
        body { margin: 1px 0 0 2px; background: #fff; font-size: 0 }
        img { object-fit: contain; height: 6px; width: 4px; overflow: hidden }
      </style>
      <img src="../resources_test/pattern.png">`)
}

func TestImageNone(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, centeredImage, `
      <style>
        @page { size: 8px }
        body { margin: 1px 0 0 1px; background: #fff; font-size: 0 }
        img { object-fit: none; height: 6px; width: 6px }
      </style>
      <img src="../resources_test/pattern.png">`)
}

func TestImageScaleDown(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, centeredImage, `
      <style>
        @page { size: 8px }
        body { margin: 1px 0 0 1px; background: #fff; font-size: 0 }
        img { object-fit: scale-down; height: 6px; width: 6px }
      </style>
      <img src="../resources_test/pattern.png">`)
}

func TestImagePosition(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, centeredImage, `
      <style>
        @page { size: 8px }
        body { margin: 1px 0 0 1px; background: #fff; font-size: 0 }
        img { object-fit: none; height: 6px; width: 6px;
              object-position: bottom 50% right 50% }
      </style>
      <img src="../resources_test/pattern.png">`)
}

func TestImagesBorder(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, borderImage, `
      <style>
        @page { size: 8px }
        body { margin: 0; font-size: 0 }
        img { margin: 1px; border: 1px solid lime }
      </style>
      <div><img src="pattern.png"></div>`)
}

func TestImagesBorderAbsolute(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertPixelsEqual(t, borderImage, `
      <style>
        @page { size: 8px }
        body { margin: 0; font-size: 0 }
        img { margin: 1px; border: 1px solid lime; position: absolute }
      </style>
      <div><img src="pattern.png"></div>`)
}

func TestImageExif(t *testing.T) {
	// FIXME: for now, we do not support
	// image with exif orientation flag (https://github.com/benoitkugler/webrender/issues/1)
	t.Skip()
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertSameRendering(t,
		`
            <style>@page { size: 10px }</style>
            <img style="display: block" src="../resources_test/not-optimized.jpg">
        `,
		`
            <style>@page { size: 10px }</style>
            <img style="display: block" src="../resources_test/not-optimized-exif.jpg">
        `,
		25,
	)
}

func TestImageExifImageOrientation(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	assertSameRendering(t,
		`
            <style>@page { size: 10px }</style>
            <img style="display: block; image-orientation: 180deg"
                 src="../resources_test/not-optimized-exif.jpg">
        `,
		`
            <style>@page { size: 10px }</style>
            <img style="display: block" src="../resources_test/not-optimized-exif.jpg">
        `,
		25,
	)
}

func TestImageExifImageOrientationKeepFormat(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)
	// Regression test for https://github.com/Kozea/WeasyPrint/issues/1755
	pdf := htmlToBytes(t,
		`
          <style>@page { size: 10px }</style>
          <img style="display: block; image-orientation: 180deg"
               src="../resources_test/not-optimized-exif.jpg">`)
	tu.AssertEqual(t, true, bytes.Contains(pdf, []byte("DCTDecode")), "")
}
