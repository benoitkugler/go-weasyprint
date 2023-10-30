package goweasyprint

import (
	"io"

	"github.com/benoitkugler/go-weasyprint/pdf"
	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/html/document"
	"github.com/benoitkugler/webrender/html/tree"
	"github.com/benoitkugler/webrender/text"
	"github.com/benoitkugler/webrender/utils"
)

// HtmlToPdf performs the convertion of an HTML document (`htmlContent`) to a PDF file,
// written in `target`.
// It is a wrapper around the following steps :
//   - `tree.NewHTML` parses the input files (HTML and CSS)
//   - `document.Render` layout the document, creating an intermediate representation ...
//   - ... which is transformed into an in-memory PDF by `document.WriteDocument`, using the `pdf.Ouput` backend.
//   - model.Write eventually serialize the PDF into `target`
//
// See `HtmlToPdfOptions` for more options.
func HtmlToPdf(target io.Writer, htmlContent utils.ContentInput, fontConfig text.FontConfiguration) error {
	return HtmlToPdfOptions(target, htmlContent, "", nil, "", nil, false, fontConfig, 1, nil)
}

// HtmlToPdfOptions is the same as HtmlToPdf, with control overs the following parameters:
//   - `baseUrl` is used as reference for links (stylesheets, images, etc...). If empty, it is
//     deduced from the html content.
//   - `urlFetcher` is a function called when resolving resources. If nil, it defaults to `utils.DefautUrlFetcher`.
//   - `mediaType` is the CSS media type used to query CSS rules. It defaults to "print".
//   - `presentationHints` controls whether or not the additional presentation stylesheet is used.
//   - `zoom` is a zoom factor
//   - `attachements` is an additional list of attachements to include into the PDF file.
func HtmlToPdfOptions(target io.Writer, htmlContent utils.ContentInput, baseUrl string, urlFetcher utils.UrlFetcher,
	mediaType string, stylesheets []tree.CSS, presentationalHints bool, fontConfig text.FontConfiguration, zoom float64, attachments []backend.Attachment,
) error {
	parsedHtml, err := tree.NewHTML(htmlContent, baseUrl, urlFetcher, mediaType)
	if err != nil {
		return err
	}
	doc := document.Render(parsedHtml, stylesheets, presentationalHints, fontConfig)
	output := pdf.NewOutput()
	doc.Write(output, utils.Fl(zoom), attachments)
	pdfDoc := output.Finalize()
	return pdfDoc.Write(target, nil)
}
