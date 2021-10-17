package layout

import (
	"fmt"
	"net/url"
	"testing"

	bo "github.com/benoitkugler/go-weasyprint/boxes"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	tu "github.com/benoitkugler/go-weasyprint/utils/testutils"
)

//  Tests for image layout.

func getImg(t *testing.T, input string) (Box, Box) {
	page := renderOnePage(t, input)
	html := page.Box().Children[0]
	body := html.Box().Children[0]
	line := body.Box().Children[0]
	img := line.Box().Children[0]
	return body, img
}

func TestImages1(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	for _, html := range []string{
		fmt.Sprintf(`<img src="%s">`, "pattern.png"),
		fmt.Sprintf(`<img src="%s">`, "pattern.gif"),
		fmt.Sprintf(`<img src="%s">`, "blue.jpg"),
		fmt.Sprintf(`<img src="%s">`, "pattern.svg"),
		fmt.Sprintf(`<img src="%s">`, "data:image/svg+xml,"+url.PathEscape(`<svg width='4' height='4'></svg>`)),
		fmt.Sprintf(`<img src="%s">`, "DatA:image/svg+xml,"+url.PathEscape(`<svg width='4px' height='4px'></svg>`)),

		"<embed src=pattern.png>",
		"<embed src=pattern.svg>",
		"<embed src=really-a-png.svg type=image/png>",
		"<embed src=really-a-svg.png type=image/svg+xml>",

		"<object data=pattern.png>",
		"<object data=pattern.svg>",
		"<object data=really-a-png.svg type=image/png>",
		"<object data=really-a-svg.png type=image/svg+xml>",
	} {
		_, img := getImg(t, html)
		tu.AssertEqual(t, img.Box().Width, pr.Float(4), html+": width")
		tu.AssertEqual(t, img.Box().Height, pr.Float(4), html+": height")
	}
}

func TestImages2(t *testing.T) {
	capt := tu.CaptureLogs()
	defer capt.AssertNoLogs(t)

	// With physical units
	data := "data:image/svg+xml," + url.PathEscape(`<svg width="2.54cm" height="0.5in"></svg>`)
	_, img := getImg(t, fmt.Sprintf(`<img src="%s">`, data))
	tu.AssertEqual(t, img.Box().Width, pr.Float(96), "")
	tu.AssertEqual(t, img.Box().Height, pr.Float(48), "")
}

// FIXME:
func TestImages3(t *testing.T) {
	// Invalid images
	for _, urlString := range []string{
		"nonexistent.png",
		"unknownprotocol://weasyprint.org/foo.png",
		"data:image/unknowntype,Not an image",
		// Invalid protocol
		`datå:image/svg+xml,<svg width="4" height="4"></svg>`,
		// zero-byte images
		"data:image/png,",
		"data:image/jpeg,",
		"data:image/svg+xml,",
		// Incorrect format
		"data:image/png,Not a PNG",
		"data:image/jpeg,Not a JPEG",
		"data:image/svg+xml,<svg>invalid xml",
	} {
		capt := tu.CaptureLogs()
		_, img := getImg(t, fmt.Sprintf(`<img src="%s" alt="invalid image">`, urlString))
		tu.AssertEqual(t, len(capt.Logs()), 1, urlString)
		tu.AssertEqual(t, img.Type(), bo.InlineBoxT, urlString) // not a replaced box
		text := img.Box().Children[0]
		tu.AssertEqual(t, text.(*bo.TextBox).Text, "invalid image", "")
	}
}

// @pytest.mark.parametrize("url", (
//     // GIF with JPEG mimetype
//     "data:image/jpeg;base64,"
//     "R0lGODlhAQABAIABAP///wAAACwAAAAAAQABAAACAkQBADs=",
//     // GIF with PNG mimetype
//     "data:image/png;base64,"
//     "R0lGODlhAQABAIABAP///wAAACwAAAAAAQABAAACAkQBADs=",
//     // PNG with JPEG mimetype
//     "data:image/jpeg;base64,"
//     "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC"
//     "0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII=",
//     // SVG with PNG mimetype
//     "data:image/png,<svg width="1" height="1"></svg>",
//     "really-a-svg.png",
//     // PNG with SVG
//     "data:image/svg+xml;base64,"
//     "R0lGODlhAQABAIABAP///wAAACwAAAAAAQABAAACAkQBADs=",
//     "really-a-png.svg",
// ))

// func TestImages4(t*testing.T,url) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     // Sniffing, no logs
//     body, img = getImg("<img src="%s">" % url)
// }

// func TestImages5(t*testing.T,) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     with captureLogs() as logs {
//         renderPages("<img src=nonexistent.png><img src=nonexistent.png>")
//     } // Failures are cached too: only one error
//     tu.AssertEqual(t, len(logs) , 1
//     tu.AssertEqual(t, "ERROR: Failed to load image" := range logs[0]
// }

// func TestImages6(t*testing.T,) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     // Layout rules try to preserve the ratio, so the height should be 40px too {
//     } body, img = getImg("""<body style="font-size: 0">
//         <img src="pattern.png" style="width: 40px">""")
//     tu.AssertEqual(t, body.Box().Height , 40
//     tu.AssertEqual(t, img.positionY , 0
//     tu.AssertEqual(t, img.Box().Width , 40
//     tu.AssertEqual(t, img.Box().Height , 40
// }

// func TestImages7(t*testing.T,) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     body, img = getImg("""<body style="font-size: 0">
//         <img src="pattern.png" style="height: 40px">""")
//     tu.AssertEqual(t, body.Box().Height , 40
//     tu.AssertEqual(t, img.positionY , 0
//     tu.AssertEqual(t, img.Box().Width , 40
//     tu.AssertEqual(t, img.Box().Height , 40
// }

// func TestImages8(t*testing.T,) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     // Same with percentages
//     body, img = getImg("""<body style="font-size: 0"><p style="width: 200px">
//         <img src="pattern.png" style="width: 20%">""")
//     tu.AssertEqual(t, body.Box().Height , 40
//     tu.AssertEqual(t, img.positionY , 0
//     tu.AssertEqual(t, img.Box().Width , 40
//     tu.AssertEqual(t, img.Box().Height , 40
// }

// func TestImages9(t*testing.T,) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     body, img = getImg("""<body style="font-size: 0">
//         <img src="pattern.png" style="min-width: 40px">""")
//     tu.AssertEqual(t, body.Box().Height , 40
//     tu.AssertEqual(t, img.positionY , 0
//     tu.AssertEqual(t, img.Box().Width , 40
//     tu.AssertEqual(t, img.Box().Height , 40
// }

// func TestImages10(t*testing.T,) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     body, img = getImg("<img src="pattern.png" style="max-width: 2px">")
//     tu.AssertEqual(t, img.Box().Width , 2
//     tu.AssertEqual(t, img.Box().Height , 2
// }

// func TestImages11(t*testing.T,) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     // display: table-cell is ignored. XXX Should it?
//     page, = renderPages("""<body style="font-size: 0">
//         <img src="pattern.png" style="width: 40px">
//         <img src="pattern.png" style="width: 60px; display: table-cell">
//     """)
//     html, = page.Box().Children[0]
//     body, = html.Box().Children[0]
//     line, = body.Box().Children[0]
//     img1, img2 = line.Box().Children[0]
//     tu.AssertEqual(t, body.Box().Height , 60
//     tu.AssertEqual(t, img1.Box().Width , 40
//     tu.AssertEqual(t, img1.Box().Height , 40
//     tu.AssertEqual(t, img2.Box().Width , 60
//     tu.AssertEqual(t, img2.Box().Height , 60
//     tu.AssertEqual(t, img1.positionY , 20
//     tu.AssertEqual(t, img2.positionY , 0

// func TestImages12(t*testing.T,):
// capt := tu.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     // Block-level image:
//     page, = renderPages("""
//         <style>
//             @page { size: 100px }
//             img { width: 40px; margin: 10px auto; display: block }
//         </style>
//         <body>
//             <img src="pattern.png">
//     """)
//     html, = page.Box().Children[0]
//     body, = html.Box().Children[0]
//     img, = body.Box().Children[0]
//     tu.AssertEqual(t, img.elementTag , "img"
//     tu.AssertEqual(t, img.positionX , 0
//     tu.AssertEqual(t, img.positionY , 0
//     tu.AssertEqual(t, img.Box().Width , 40
//     tu.AssertEqual(t, img.Box().Height , 40
//     tu.AssertEqual(t, img.contentBoxX() , 30  // (100 - 40) / 2 , 30px for margin-left
//     tu.AssertEqual(t, img.contentBoxY() , 10
// }

// func TestImages13(t*testing.T,) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     page, = renderPages("""
//         <style>
//             @page { size: 100px }
//             img { min-width: 40%; margin: 10px auto; display: block }
//         </style>
//         <body>
//             <img src="pattern.png">
//     """)
//     html, = page.Box().Children[0]
//     body, = html.Box().Children[0]
//     img, = body.Box().Children[0]
//     tu.AssertEqual(t, img.elementTag , "img"
//     tu.AssertEqual(t, img.positionX , 0
//     tu.AssertEqual(t, img.positionY , 0
//     tu.AssertEqual(t, img.Box().Width , 40
//     tu.AssertEqual(t, img.Box().Height , 40
//     tu.AssertEqual(t, img.contentBoxX() , 30  // (100 - 40) / 2 , 30px for margin-left
//     tu.AssertEqual(t, img.contentBoxY() , 10

// func TestImages14(t*testing.T,):
// capt := tu.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     page, = renderPages("""
//         <style>
//             @page { size: 100px }
//             img { min-width: 40px; margin: 10px auto; display: block }
//         </style>
//         <body>
//             <img src="pattern.png">
//     """)
//     html, = page.Box().Children[0]
//     body, = html.Box().Children[0]
//     img, = body.Box().Children[0]
//     tu.AssertEqual(t, img.elementTag , "img"
//     tu.AssertEqual(t, img.positionX , 0
//     tu.AssertEqual(t, img.positionY , 0
//     tu.AssertEqual(t, img.Box().Width , 40
//     tu.AssertEqual(t, img.Box().Height , 40
//     tu.AssertEqual(t, img.contentBoxX() , 30  // (100 - 40) / 2 , 30px for margin-left
//     tu.AssertEqual(t, img.contentBoxY() , 10
// }

// func TestImages15(t*testing.T,) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     page, = renderPages("""
//         <style>
//             @page { size: 100px }
//             img { min-height: 30px; max-width: 2px;
//                   margin: 10px auto; display: block }
//         </style>
//         <body>
//             <img src="pattern.png">
//     """)
//     html, = page.Box().Children[0]
//     body, = html.Box().Children[0]
//     img, = body.Box().Children[0]
//     tu.AssertEqual(t, img.elementTag , "img"
//     tu.AssertEqual(t, img.positionX , 0
//     tu.AssertEqual(t, img.positionY , 0
//     tu.AssertEqual(t, img.Box().Width , 2
//     tu.AssertEqual(t, img.Box().Height , 30
//     tu.AssertEqual(t, img.contentBoxX() , 49  // (100 - 2) / 2 , 49px for margin-left
//     tu.AssertEqual(t, img.contentBoxY() , 10

// func TestImages16(t*testing.T,):
// capt := tu.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     page, = renderPages("""
//         <body style="float: left">
//         <img style="height: 200px; margin: 10px; display: block" src="
//             data:image/svg+xml,
//             <svg width="150" height="100"></svg>
//         ">
//     """)
//     html, = page.Box().Children[0]
//     body, = html.Box().Children[0]
//     img, = body.Box().Children[0]
//     tu.AssertEqual(t, body.Box().Width , 320
//     tu.AssertEqual(t, body.Box().Height , 220
//     tu.AssertEqual(t, img.elementTag , "img"
//     tu.AssertEqual(t, img.Box().Width , 300
//     tu.AssertEqual(t, img.Box().Height , 200
// }

// func TestImages17(t*testing.T,) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     page, = renderPages("""
//         <div style="width: 300px; height: 300px">
//         <img src="
//             data:image/svg+xml,
//             <svg viewBox="0 0 20 10"></svg>
//         ">""")
//     html, = page.Box().Children[0]
//     body, = html.Box().Children[0]
//     div, = body.Box().Children[0]
//     line, = div.Box().Children[0]
//     img, = line.Box().Children[0]
//     tu.AssertEqual(t, div.Box().Width , 300
//     tu.AssertEqual(t, div.Box().Height , 300
//     tu.AssertEqual(t, img.elementTag , "img"
//     tu.AssertEqual(t, img.Box().Width , 300
//     tu.AssertEqual(t, img.Box().Height , 150
// }

// func TestImages18(t*testing.T,) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     // Test regression: https://github.com/Kozea/WeasyPrint/issues/1050
//     page, = renderPages("""
//         <img style="position: absolute" src="
//             data:image/svg+xml,
//             <svg viewBox="0 0 20 10"></svg>
//         ">""")
// }

// func TestLinearGradient(t*testing.T,) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     red = (1, 0, 0, 1)
//     lime = (0, 1, 0, 1)
//     blue = (0, 0, 1, 1)
// }
//     def layout(gradientCss, type="linear", init=(),
//                positions=[0, 1], colors=[blue, lime]) {
//                }
//         page, = renderPages("<style>@page { background: " + gradientCss)
//         layer, = page.background.layers
//         result = layer.image.layout(400, 300)
//         tu.AssertEqual(t, result[0] , 1
//         tu.AssertEqual(t, result[1] , type
//         tu.AssertEqual(t, result[2] , (None if init  , nil  else pytest.approx(init))
//         tu.AssertEqual(t, result[3] , pytest.approx(positions)
//         for color1, color2 := range zip(result[4], colors) {
//             tu.AssertEqual(t, color1 , pytest.approx(color2)
//         }

//     layout("linear-gradient(blue)", "solid", None, [], [blue])
//     layout("repeating-linear-gradient(blue)", "solid", None, [], [blue])
//     layout("linear-gradient(blue, lime)", init=(200, 0, 200, 300))
//     layout("repeating-linear-gradient(blue, lime)", init=(200, 0, 200, 300))
//     layout("repeating-linear-gradient(blue, lime 100px)",
//            positions=[0, 1, 1, 2, 2, 3], init=(200, 0, 200, 300))

//     layout("linear-gradient(to bottom, blue, lime)", init=(200, 0, 200, 300))
//     layout("linear-gradient(to top, blue, lime)", init=(200, 300, 200, 0))
//     layout("linear-gradient(to right, blue, lime)", init=(0, 150, 400, 150))
//     layout("linear-gradient(to left, blue, lime)", init=(400, 150, 0, 150))

//     layout("linear-gradient(to top left, blue, lime)",
//            init=(344, 342, 56, -42))
//     layout("linear-gradient(to top right, blue, lime)",
//            init=(56, 342, 344, -42))
//     layout("linear-gradient(to bottom left, blue, lime)",
//            init=(344, -42, 56, 342))
//     layout("linear-gradient(to bottom right, blue, lime)",
//            init=(56, -42, 344, 342))

//     layout("linear-gradient(270deg, blue, lime)", init=(400, 150, 0, 150))
//     layout("linear-gradient(.75turn, blue, lime)", init=(400, 150, 0, 150))
//     layout("linear-gradient(45deg, blue, lime)", init=(25, 325, 375, -25))
//     layout("linear-gradient(.125turn, blue, lime)", init=(25, 325, 375, -25))
//     layout("linear-gradient(.375turn, blue, lime)", init=(25, -25, 375, 325))
//     layout("linear-gradient(.625turn, blue, lime)", init=(375, -25, 25, 325))
//     layout("linear-gradient(.875turn, blue, lime)", init=(375, 325, 25, -25))

//     layout("linear-gradient(blue 2em, lime 20%)", init=(200, 32, 200, 60))
//     layout("linear-gradient(blue 100px, red, blue, red 160px, lime)",
//            init=(200, 100, 200, 300), colors=[blue, red, blue, red, lime],
//            positions=[0, .1, .2, .3, 1])
//     layout("linear-gradient(blue -100px, blue 0, red -12px, lime 50%)",
//            init=(200, -100, 200, 150), colors=[blue, blue, red, lime],
//            positions=[0, .4, .4, 1])
//     layout("linear-gradient(blue, blue, red, lime -7px)",
//            init=(200, -1, 200, 1), colors=[blue, blue, blue, red, lime, lime],
//            positions=[0, 0.5, 0.5, 0.5, 0.5, 1])
//     layout("repeating-linear-gradient(blue, blue, lime, lime -7px)",
//            "solid", None, [], [(0, .5, .5, 1)])

// func TestRadialGradient(t*testing.T,) {
// 	capt := tu.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

//     red = (1, 0, 0, 1)
//     lime = (0, 1, 0, 1)
//     blue = (0, 0, 1, 1)
// }
//     def layout(gradientCss, type="radial", init=(),
//                positions=[0, 1], colors=[blue, lime], scaleY=1) {
//                }
//         if type_ , "radial" {
//             centerX, centerY, radius0, radius1 = init
//             init = (centerX, centerY / scaleY, radius0,
//                     centerX, centerY / scaleY, radius1)
//         } page, = renderPages("<style>@page { background: " + gradientCss)
//         layer, = page.background.layers
//         result = layer.image.layout(400, 300)
//         tu.AssertEqual(t, result[0] , scaleY
//         tu.AssertEqual(t, result[1] , type
//         tu.AssertEqual(t, result[2] , (None if init  , nil  else pytest.approx(init))
//         tu.AssertEqual(t, result[3] , pytest.approx(positions)
//         for color1, color2 := range zip(result[4], colors) {
//             tu.AssertEqual(t, color1 , pytest.approx(color2)
//         }

//     layout("radial-gradient(blue)", "solid", None, [], [blue])
//     layout("repeating-radial-gradient(blue)", "solid", None, [], [blue])
//     layout("radial-gradient(100px, blue, lime)",
//            init=(200, 150, 0, 100))

//     layout("radial-gradient(100px at right 20px bottom 30px, lime, red)",
//            init=(380, 270, 0, 100), colors=[lime, red])
//     layout("radial-gradient(0 0, blue, lime)",
//            init=(200, 150, 0, 1e-7))
//     layout("radial-gradient(1px 0, blue, lime)",
//            init=(200, 150, 0, 1e7), scaleY=1e-14)
//     layout("radial-gradient(0 1px, blue, lime)",
//            init=(200, 150, 0, 1e-7), scaleY=1e14)
//     layout("repeating-radial-gradient(100px 200px, blue, lime)",
//            positions=[0, 1, 1, 2, 2, 3], init=(200, 150, 0, 300),
//            scaleY=(200 / 100))
//     layout("repeating-radial-gradient(42px, blue -100px, lime 100px)",
//            positions=[-0.5, 0, 0, 1], init=(200, 150, 0, 300),
//            colors=[(0, 0.5, 0.5, 1), lime, blue, lime])
//     layout("radial-gradient(42px, blue -20px, lime -1px)",
//            "solid", None, [], [lime])
//     layout("radial-gradient(42px, blue -20px, lime 0)",
//            "solid", None, [], [lime])
//     layout("radial-gradient(42px, blue -20px, lime 20px)",
//            init=(200, 150, 0, 20), colors=[(0, .5, .5, 1), lime])

//     layout("radial-gradient(100px 120px, blue, lime)",
//            init=(200, 150, 0, 100), scaleY=(120 / 100))
//     layout("radial-gradient(25% 40%, blue, lime)",
//            init=(200, 150, 0, 100), scaleY=(120 / 100))

//     layout("radial-gradient(circle closest-side, blue, lime)",
//            init=(200, 150, 0, 150))
//     layout("radial-gradient(circle closest-side at 150px 50px, blue, lime)",
//            init=(150, 50, 0, 50))
//     layout("radial-gradient(circle closest-side at 45px 50px, blue, lime)",
//            init=(45, 50, 0, 45))
//     layout("radial-gradient(circle closest-side at 420px 50px, blue, lime)",
//            init=(420, 50, 0, 20))
//     layout("radial-gradient(circle closest-side at 420px 281px, blue, lime)",
//            init=(420, 281, 0, 19))

//     layout("radial-gradient(closest-side, blue 20%, lime)",
//            init=(200, 150, 40, 200), scaleY=(150 / 200))
//     layout("radial-gradient(closest-side at 300px 20%, blue, lime)",
//            init=(300, 60, 0, 100), scaleY=(60 / 100))
//     layout("radial-gradient(closest-side at 10% 230px, blue, lime)",
//            init=(40, 230, 0, 40), scaleY=(70 / 40))

//     layout("radial-gradient(circle farthest-side, blue, lime)",
//            init=(200, 150, 0, 200))
//     layout("radial-gradient(circle farthest-side at 150px 50px, blue, lime)",
//            init=(150, 50, 0, 250))
//     layout("radial-gradient(circle farthest-side at 45px 50px, blue, lime)",
//            init=(45, 50, 0, 355))
//     layout("radial-gradient(circle farthest-side at 420px 50px, blue, lime)",
//            init=(420, 50, 0, 420))
//     layout("radial-gradient(circle farthest-side at 220px 310px, blue, lime)",
//            init=(220, 310, 0, 310))

//     layout("radial-gradient(farthest-side, blue, lime)",
//            init=(200, 150, 0, 200), scaleY=(150 / 200))
//     layout("radial-gradient(farthest-side at 300px 20%, blue, lime)",
//            init=(300, 60, 0, 300), scaleY=(240 / 300))
//     layout("radial-gradient(farthest-side at 10% 230px, blue, lime)",
//            init=(40, 230, 0, 360), scaleY=(230 / 360))

//     layout("radial-gradient(circle closest-corner, blue, lime)",
//            init=(200, 150, 0, 250))
//     layout("radial-gradient(circle closest-corner at 340px 80px, blue, lime)",
//            init=(340, 80, 0, 100))
//     layout("radial-gradient(circle closest-corner at 0 342px, blue, lime)",
//            init=(0, 342, 0, 42))

//     layout("radial-gradient(closest-corner, blue, lime)",
//            init=(200, 150, 0, 200 * 2 ** 0.5), scaleY=(150 / 200))
//     layout("radial-gradient(closest-corner at 450px 100px, blue, lime)",
//            init=(450, 100, 0, 50 * 2 ** 0.5), scaleY=(100 / 50))
//     layout("radial-gradient(closest-corner at 40px 210px, blue, lime)",
//            init=(40, 210, 0, 40 * 2 ** 0.5), scaleY=(90 / 40))

//     layout("radial-gradient(circle farthest-corner, blue, lime)",
//            init=(200, 150, 0, 250))
//     layout("radial-gradient(circle farthest-corner"
//            " at 300px -100px, blue, lime)",
//            init=(300, -100, 0, 500))
//     layout("radial-gradient(circle farthest-corner at 400px 0, blue, lime)",
//            init=(400, 0, 0, 500))

//     layout("radial-gradient(farthest-corner, blue, lime)",
//            init=(200, 150, 0, 200 * 2 ** 0.5), scaleY=(150 / 200))
//     layout("radial-gradient(farthest-corner at 450px 100px, blue, lime)",
//            init=(450, 100, 0, 450 * 2 ** 0.5), scaleY=(200 / 450))
//     layout("radial-gradient(farthest-corner at 40px 210px, blue, lime)",
//            init=(40, 210, 0, 360 * 2 ** 0.5), scaleY=(210 / 360))

// @pytest.mark.parametrize("props, divWidth", (
//     ({}, 4),
//     ({"min-width": "10px"}, 10),
//     ({"max-width": "1px"}, 1),
//     ({"width": "10px"}, 10),
//     ({"width": "1px"}, 1),
//     ({"min-height": "10px"}, 10),
//     ({"max-height": "1px"}, 1),
//     ({"height": "10px"}, 10),
//     ({"height": "1px"}, 1),
//     ({"min-width": "10px", "min-height": "1px"}, 10),
//     ({"min-width": "1px", "min-height": "10px"}, 10),
//     ({"max-width": "10px", "max-height": "1px"}, 1),
//     ({"max-width": "1px", "max-height": "10px"}, 1),
// ))
// func TestImageMinMaxWidth(t*testing.T,props, divWidth) {
//     default = {
//         "min-width": "auto", "max-width": "none", "width": "auto",
//         "min-height": "auto", "max-height": "none", "height": "auto"}
//     page, = renderPages("""
//       <style> img { display: block; %s } </style>
//       <div style="display: inline-block">
//         <img src="pattern.png"><img src="pattern.svg">
//       </div>""" % ";".join(
//           f"{key}: {props.get(key, value)}" for key, value := range default.items()))
//     html, = page.Box().Children[0]
//     body, = html.Box().Children[0]
//     line, = body.Box().Children[0]
//     div, = line.Box().Children[0]
//     tu.AssertEqual(t, div.Box().Width , divWidth
