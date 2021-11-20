package pdf

// Test how images are drawn.
// TODO: complete when SVG is supported

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
)

// func TestImages(t *testing.T) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

// 	for _, data := range [][2]string{
// 		{"pattern.svg", centeredImage},
// 		{"pattern.png", centeredImage},
// 		{"pattern.palette.png", centeredImage},
// 		{"pattern.gif", centeredImage},
// 		{"blue.jpg", blueImage},
// 	} {
// 		filename, image := data[0], data[1]
// 		assertPixelsEqual(t, "inline_image_"+filename, image, fmt.Sprintf(`
//       <style>
//         @page { size: 8px }
//         body { margin: 2px 0 0 2px; background: #fff; font-size: 0 }
//       </style>
//       <div><img src="../resources_test/%s"></div>`, filename))
// 	}
// }

// @pytest.mark.parametrize("filename", (
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     "pattern.svg",
//     "pattern.png",
//     "pattern.palette.png",
//     "pattern.gif",
// ))
// func TestResizedImages(t *testing.T, filename) {
//     assertPixelsEqual(t, f"resized_image_{filename}", 12, 12, resizedImage, `
//       <style>
//         @page { size: 12px }
//         body { margin: 2px 0 0 2px; background: #fff; font-size: 0 }
//         img { display: block; width: 8px; image-rendering: pixelated }
//       </style>
//       <div><img src="%s"></div>` % filename)

// @pytest.mark.parametrize("viewbox, width, height", (
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     (None, None, None),
//     (None, 4, None),
//     (None, None, 4),
//     (None, 4, 4),
//     ("0 0 4 4", 4, None),
//     ("0 0 4 4", None, 4),
//     ("0 0 4 4", 4, 4),
//     ("0 0 4 4", 4, 4),
// ))
// func TestSvgSizing(t *testing.T, viewbox, width, height) {
//     assertPixelsEqual(t,
//         f"svg_sizing_{viewbox}_{width}_{height}", 8, 8,
//         centeredImage, `
//       <style>
//         @page { size: 8px }
//         body { margin: 2px 0 0 2px; background: #fff; font-size: 0 }
//         svg { display: block }
//       </style>
//       <svg %s %s %s>
//         <rect width="4px" height="4px" fill="#00f" />
//         <rect width="1px" height="1px" fill="#f00" />
//       </svg>` % (
//           f"width="{width}"" if width else ",
//           f"height="{height}px"" if height else ",
//           f"viewbox="{viewbox}"" if viewbox else "))

// @pytest.mark.parametrize("viewbox, width, height, image", (
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     (None, None, None, smallResizedImage),
//     (None, 8, None, smallResizedImage),
//     (None, None, 8, smallResizedImage),
//     (None, 8, 8, smallResizedImage),
//     ("0 0 4 4", None, None, resizedImage),
//     ("0 0 4 4", 8, None, resizedImage),
//     ("0 0 4 4", None, 8, resizedImage),
//     ("0 0 4 4", 8, 8, resizedImage),
//     ("0 0 4 4", 800, 800, resizedImage),
// ))
// func TestSvgResizing(t *testing.T, viewbox, width, height, image) {
//     assertPixelsEqual(t,
//         f"svg_resizing_{viewbox}_{width}_{height}", 12, 12, image, `
//       <style>
//         @page { size: 12px }
//         body { margin: 2px 0 0 2px; background: #fff; font-size: 0 }
//         svg { display: block; width: 8px }
//       </style>
//       <svg %s %s %s>
//         <rect width="4" height="4" fill="#00f" />
//         <rect width="1" height="1" fill="#f00" />
//       </svg>` % (
//           f"width="{width}"" if width else ",
//           f"height="{height}px"" if height else ",
//           f"viewbox="{viewbox}"" if viewbox else "))

// func TestImagesBlock(t *testing.T, ) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     assertPixelsEqual(t, "block_image", 8, 8, centeredImage, `
//       <style>
//         @page { size: 8px }
//         body { margin: 0; background: #fff; font-size: 0 }
//         img { display: block; margin: 2px auto 0 }
//       </style>
//       <div><img src="pattern.png"></div>`)

// func TestImagesNotFound(t *testing.T, ) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     with capture_logs() as logs:
//         assertPixelsEqual(t, "image_not_found", 8, 8, noImage, `
//           <style>
//             @page { size: 8px }
//             body { margin: 0; background: #fff; font-size: 0 }
//             img { display: block; margin: 2px auto 0 }
//           </style>
//           <div><img src="inexistent1.png" alt=""></div>`)
//     assert len(logs) == 1
//     assert "ERROR: Failed to load image" in logs[0]
//     assert "inexistent1.png" in logs[0]

// func TestImagesNoSrc(t *testing.T, ) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     assertPixelsEqual(t, "image_no_src", 8, 8, noImage, `
//       <style>
//         @page { size: 8px }
//         body { margin: 0; background: #fff; font-size: 0 }
//         img { display: block; margin: 2px auto 0 }
//       </style>
//       <div><img alt=""></div>`)

// func TestImagesAlt(t *testing.T, ) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     with capture_logs() as logs:
//         assert_same_rendering(200, 30, [
//             (name, `
//               <style>
//                 @page { size: 200px 30px }
//                 body { margin: 0; background: #fff; font-size: 0 }
//               </style>
//               <div>%s</div>` % html)
//             for name, html in [
//                 ("image_alt_text_reference", "Hello, world!"),
//                 ("image_alt_text_not_found",
//                     "<img src="inexistent2.png" alt="Hello, world!">"),
//                 ("image_alt_text_no_src",
//                     "<img alt="Hello, world!">"),
//                 ("image_svg_no_intrinsic_size",
//                     `<img src="data:image/svg+xml,<svg></svg>"
//                             alt="Hello, world!">`),
//             ]
//         ])
//     assert len(logs) == 1
//     assert "ERROR: Failed to load image" in logs[0]
//     assert "inexistent2.png" in logs[0]

// func TestImagesRepeatTransparent(t *testing.T, ) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     # Test regression: https://github.com/Kozea/WeasyPrint/issues/1440
//     assertPixelsEqual(t, "image_repeat_transparent", 1, 3, "_\n_\n_", `
//       <style>
//         @page { size: 1px }
//         div { height: 100px; width: 100px; background: url(logo_small.png) }
//       </style>
//       <div></div><div></div><div></div>`)

// func TestImagesNoWidth(t *testing.T, ) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     assertPixelsEqual(t, "image_0x1", 8, 8, noImage, `
//       <style>
//         @page { size: 8px }
//         body { margin: 2px; background: #fff; font-size: 0 }
//       </style>
//       <div><img src="pattern.png" alt="not shown"
//                 style="width: 0; height: 1px"></div>`)

// func TestImagesNoHeight(t *testing.T, ) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     assertPixelsEqual(t, "image_1x0", 8, 8, noImage, `
//       <style>
//         @page { size: 8px }
//         body { margin: 2px; background: #fff; font-size: 0 }
//       </style>
//       <div><img src="pattern.png" alt="not shown"
//                 style="width: 1px; height: 0"></div>`)

// func TestImagesNoWidthHeight(t *testing.T, ) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     assertPixelsEqual(t, "image_0x0", 8, 8, noImage, `
//       <style>
//         @page { size: 8px }
//         body { margin: 2px; background: #fff; font-size: 0 }
//       </style>
//       <div><img src="pattern.png" alt="not shown"
//                 style="width: 0; height: 0"></div>`)

// func TestImagesPageBreak(t *testing.T, ) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     assertPixelsEqual(t, "image_page_break", 8, 3 * 8, pageBreak, `
//       <style>
//         @page { size: 8px; margin: 2px; background: #fff }
//         body { font-size: 0 }
//       </style>
//       <div><img src="pattern.png"></div>
//       <div style="page-break-before: right"><img src="pattern.png"></div>`)

// func TestImageRepeatInline(t *testing.T, ) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     # Test regression: https://github.com/Kozea/WeasyPrint/issues/808
//     assertPixelsEqual(t, "image_page_repeat_inline", 8, 2 * 8, table, `
//       <style>
//         @page { size: 8px; margin: 0; background: #fff }
//         table { border-collapse: collapse; margin: 2px }
//         th, td { border: none; padding: 0 }
//         th { height: 4px; line-height: 4px }
//         td { height: 2px }
//         img { vertical-align: top }
//       </style>
//       <table>
//         <thead>
//           <tr><th><img src="pattern.png"></th></tr>
//         </thead>
//         <tbody>
//           <tr><td></td></tr>
//           <tr><td></td></tr>
//         </tbody>
//       </table>`)

// func TestImageRepeatBlock(t *testing.T, ) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     # Test regression: https://github.com/Kozea/WeasyPrint/issues/808
//     assertPixelsEqual(t, "image_page_repeat_block", 8, 2 * 8, table, `
//       <style>
//         @page { size: 8px; margin: 0; background: #fff }
//         table { border-collapse: collapse; margin: 2px }
//         th, td { border: none; padding: 0 }
//         th { height: 4px }
//         td { height: 2px }
//         img { display: block }
//       </style>
//       <table>
//         <thead>
//           <tr><th><img src="pattern.png"></th></tr>
//         </thead>
//         <tbody>
//           <tr><td></td></tr>
//           <tr><td></td></tr>
//         </tbody>
//       </table>`)

// func TestImagesPadding(t *testing.T, ) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     # Regression test: padding used to be ignored on images
//     assertPixelsEqual(t, "image_with_padding", 8, 8, centeredImage, `
//       <style>
//         @page { size: 8px; background: #fff }
//         body { font-size: 0 }
//       </style>
//       <div style="line-height: 1px">
//         <img src=pattern.png style="padding: 2px 0 0 2px">
//       </div>`)

// func TestImagesInInlineBlock(t *testing.T, ) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     # Regression test: this used to cause an exception
//     assertPixelsEqual(t, "image_in_inline_block", 8, 8, centeredImage, `
//       <style>
//         @page { size: 8px }
//         body { margin: 2px 0 0 2px; background: #fff; font-size: 0 }
//       </style>
//       <div style="display: inline-block">
//         <p><img src=pattern.png></p>
//       </div>`)

// func TestImagesSharedPattern(t *testing.T, ) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     # The same image is used in a repeating background,
//     # then in a non-repating <img>.
//     # If Pattern objects are shared carelessly, the image will be repeated.
//     assertPixelsEqual(t, "image_shared_pattern", 12, 12, `
//         ____________
//         ____________
//         __aaaaaaaa__
//         __aaaaaaaa__
//         ____________
//         __aaaa______
//         __aaaa______
//         __aaaa______
//         __aaaa______
//         ____________
//         ____________
//         ____________
//     `, `
//       <style>
//         @page { size: 12px }
//         body { margin: 2px; background: #fff; font-size: 0 }
//       </style>
//       <div style="background: url(blue.jpg);
//                   height: 2px; margin-bottom: 1px"></div>
//       <img src=blue.jpg>
//     `)

// func TestImageResolution(t *testing.T, ) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     assert_same_rendering(20, 20, [
//         ("image_resolution_ref", `
//             <style>@page { size: 20px; margin: 2px; background: #fff }</style>
//             <div style="font-size: 0">
//                 <img src="pattern.png" style="width: 8px"></div>
//         `),
//         ("image_resolution_img", `
//             <style>@page { size: 20px; margin: 2px; background: #fff }</style>
//             <div style="image-resolution: .5dppx; font-size: 0">
//                 <img src="pattern.png"></div>
//         `),
//         ("image_resolution_content", `
//             <style>@page { size: 20px; margin: 2px; background: #fff }
//                    div::before { content: url(pattern.png) }
//             </style>
//             <div style="image-resolution: .5dppx; font-size: 0"></div>
//         `),
//         ("image_resolution_background", `
//             <style>@page { size: 20px; margin: 2px; background: #fff }
//             </style>
//             <div style="height: 16px; image-resolution: .5dppx;
//                         background: url(pattern.png) no-repeat"></div>
//         `),
//     ])

// func TestImageCover(t *testing.T, ) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     assertPixelsEqual(t, "image_cover", 8, 8, coverImage, `
//       <style>
//         @page { size: 8px }
//         body { margin: 2px 0 0 2px; background: #fff; font-size: 0 }
//         img { object-fit: cover; height: 4px; width: 2px }
//       </style>
//       <img src="pattern.png">`)

// func TestImageContain(t *testing.T, ) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     assertPixelsEqual(t, "image_contain", 8, 8, centeredImage, `
//       <style>
//         @page { size: 8px }
//         body { margin: 1px 0 0 2px; background: #fff; font-size: 0 }
//         img { object-fit: contain; height: 6px; width: 4px }
//       </style>
//       <img src="pattern.png">`)

// func TestImageNone(t *testing.T, ) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     assertPixelsEqual(t, "image_none", 8, 8, centeredImage, `
//       <style>
//         @page { size: 8px }
//         body { margin: 1px 0 0 1px; background: #fff; font-size: 0 }
//         img { object-fit: none; height: 6px; width: 6px }
//       </style>
//       <img src="pattern.png">`)

// func TestImageScaleDown(t *testing.T, ) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     assertPixelsEqual(t, "image_scale_down", 8, 8, centeredImage, `
//       <style>
//         @page { size: 8px }
//         body { margin: 1px 0 0 1px; background: #fff; font-size: 0 }
//         img { object-fit: scale-down; height: 6px; width: 6px }
//       </style>
//       <img src="pattern.png">`)

// func TestImagePosition(t *testing.T, ) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)
//     assertPixelsEqual(t, "image_position", 8, 8, centeredImage, `
//       <style>
//         @page { size: 8px }
//         body { margin: 1px 0 0 1px; background: #fff; font-size: 0 }
//         img { object-fit: none; height: 6px; width: 6px;
//               object-position: bottom 50% right 50% }
//       </style>
//       <img src="pattern.png">`)
