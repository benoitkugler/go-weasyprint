package goweasyprint

import (
	"io"

	"github.com/benoitkugler/go-weasyprint/document"
	"github.com/benoitkugler/go-weasyprint/layout/text"
	"github.com/benoitkugler/go-weasyprint/pdf"
	"github.com/benoitkugler/go-weasyprint/style/tree"
	"github.com/benoitkugler/go-weasyprint/utils"
)

// reference commit : 238e214f9d58da51eed8f790fce34dc5fc1dc67f

func HtmlToPdf(target io.Writer, htmlContent utils.ContentInput, baseUrl string, urlFetcher utils.UrlFetcher, mediaType string, stylesheets []tree.CSS,
	enableHinting, presentationalHints bool, fontConfig *text.FontConfiguration, zoom float64, attachments []utils.Attachment) error {
	parsedHtml, err := tree.NewHTML(htmlContent, baseUrl, urlFetcher, mediaType)
	if err != nil {
		return err
	}
	doc := document.Render(parsedHtml, stylesheets, presentationalHints, fontConfig)
	return pdf.WritePDF(doc, target, zoom, attachments)
}
