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

// @assertNoLogs
// func testHtmlMeta() {
//     def assertMeta(html, **meta) {
//         meta.setdefault("title", None)
//         meta.setdefault("authors", [])
//         meta.setdefault("keywords", [])
//         meta.setdefault("generator", None)
//         meta.setdefault("description", None)
//         meta.setdefault("created", None)
//         meta.setdefault("modified", None)
//         meta.setdefault("attachments", [])
//         assert vars(FakeHTML(string=html).render().metadata) == meta
//     }
// }
//     assertMeta("<body>")
//     assertMeta(
//         """
//             <meta name=author content="I Me &amp; Myself">
//             <meta name=author content="Smith, John">
//             <title>Test document</title>
//             <h1>Another title</h1>
//             <meta name=generator content="Human after all">
//             <meta name=dummy content=ignored>
//             <meta name=dummy>
//             <meta content=ignored>
//             <meta>
//             <meta name=keywords content="html ,\tcss,
//                                          pdf,css">
//             <meta name=dcterms.created content=2011-04>
//             <meta name=dcterms.created content=2011-05>
//             <meta name=dcterms.modified content=2013>
//             <meta name=keywords content="Python; pydyf">
//             <meta name=description content="Blah… ">
//         """,
//         authors=["I Me & Myself", "Smith, John"],
//         title="Test document",
//         generator="Human after all",
//         keywords=["html", "css", "pdf", "Python; pydyf"],
//         description="Blah… ",
//         created="2011-04",
//         modified="2013")
//     assertMeta(
//         """
//             <title>One</title>
//             <meta name=Author>
//             <title>Two</title>
//             <title>Three</title>
//             <meta name=author content=Me>
//         """,
//         title="One",
//         authors=["", "Me"])

// @assertNoLogs
// func testHttp() {
//     def gzipCompress(data) {
//         fileObj = io.BytesIO()
//         gzipFile = gzip.GzipFile(fileobj=fileObj, mode="wb")
//         gzipFile.write(data)
//         gzipFile.close()
//         return fileObj.getvalue()
//     }
// }
//     with httpServer({
//         "/gzip": lambda env: (
//             (gzipCompress(b"<html test=ok>"), [("Content-Encoding", "gzip")])
//             if "gzip" := range env.get("HTTPACCEPTENCODING", "") else
//             (b"<html test=accept-encoding-header-fail>", [])
//         ),
//         "/deflate": lambda env: (
//             (zlib.compress(b"<html test=ok>"),
//              [("Content-Encoding", "deflate")])
//             if "deflate" := range env.get("HTTPACCEPTENCODING", "") else
//             (b"<html test=accept-encoding-header-fail>", [])
//         ),
//         "/raw-deflate": lambda env: (
//             // Remove zlib header && checksum
//             (zlib.compress(b"<html test=ok>")[2:-4],
//              [("Content-Encoding", "deflate")])
//             if "deflate" := range env.get("HTTPACCEPTENCODING", "") else
//             (b"<html test=accept-encoding-header-fail>", [])
//         ),
//     }) as rootUrl {
//         assert HTML(rootUrl + "/gzip").etreeElement.get("test") == "ok"
//         assert HTML(rootUrl + "/deflate").etreeElement.get("test") == "ok"
//         assert HTML(
//             rootUrl + "/raw-deflate").etreeElement.get("test") == "ok"
