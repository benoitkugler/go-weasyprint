package pdf

// // Make relative URL references work with our custom URL scheme.
// usesRelative.append("weasyprint-custom")

// @assertNoLogs
// func testUrlFetcher():
//     filename = resourceFilename("pattern.png")
//     with open(filename, "rb") as patternFd:
//         patternPng = patternFd.read()

//     def fetcher(url):
//         if url == "weasyprint-custom:foo/%C3%A9%e9Pattern":
//             return {"string": patternPng, "mimeType": "image/png"}
//         else if url == "weasyprint-custom:foo/bar.css":
//             return {
//                 "string": "body { background: url(é%e9Pattern)",
//                 "mimeType": "text/css"}
//         else if url == "weasyprint-custom:foo/bar.no":
//             return {
//                 "string": "body { background: red }",
//                 "mimeType": "text/no"}
//         else:
//             return defaultUrlFetcher(url)

//     baseUrl = resourceFilename("dummy.html")
//     css = CSS(string="""
//         @page { size: 8px; margin: 2px; background: #fff }
//         body { margin: 0; font-size: 0 }
//     """, baseUrl=baseUrl)

//     def test(html, blank=false) {
//         html = FakeHTML(string=html, urlFetcher=fetcher, baseUrl=baseUrl)
//         checkPngPattern(html.writePng(stylesheets=[css]), blank=blank)
//     }

//     test("<body><img src="pattern.png">")  // Test a "normal" URL
//     test(f"<body><img src="{Path(filename).asUri()}">")
//     test(f"<body><img src="{Path(filename).asUri()}?ignored">")
//     test("<body><img src="weasyprint-custom:foo/é%e9Pattern">")
//     test("<body style="background: url(weasyprint-custom:foo/é%e9Pattern)">")
//     test("<body><li style="list-style: inside "
//          "url(weasyprint-custom:foo/é%e9Pattern)">")
//     test("<link rel=stylesheet href="weasyprint-custom:foo/bar.css"><body>")
//     test("<style>@import "weasyprint-custom:foo/bar.css";</style><body>")
//     test("<style>@import url(weasyprint-custom:foo/bar.css);</style><body>")
//     test("<style>@import url("weasyprint-custom:foo/bar.css");</style><body>")
//     test("<link rel=stylesheet href="weasyprint-custom:foo/bar.css"><body>")

//     with captureLogs() as logs {
//         test("<body><img src="custom:foo/bar">", blank=true)
//     } assert len(logs) == 1
//     assert logs[0].startswith(
//         "ERROR: Failed to load image at "custom:foo/bar"")

//     with captureLogs() as logs {
//         test(
//             "<link rel=stylesheet href="weasyprint-custom:foo/bar.css">"
//             "<link rel=stylesheet href="weasyprint-custom:foo/bar.no"><body>")
//     } assert len(logs) == 1
//     assert logs[0].startswith("ERROR: Unsupported stylesheet type text/no")

//     def fetcher2(url) {
//         assert url == "weasyprint-custom:%C3%A9%e9.css"
//         return {"string": "", "mimeType": "text/css"}
//     } FakeHTML(
//         string="<link rel=stylesheet href="weasyprint-custom:é%e9.css"><body",
//         urlFetcher=fetcher2).render()
