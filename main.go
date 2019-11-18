package goweasyprint

import (
	"github.com/benoitkugler/go-weasyprint/document"
	"github.com/benoitkugler/go-weasyprint/fonts"
	"github.com/benoitkugler/go-weasyprint/pdf"
	"github.com/benoitkugler/go-weasyprint/style/tree"
	"github.com/benoitkugler/go-weasyprint/utils"
	"io"
)

func HtmlToPdf(target io.Writer, htmlContent utils.ContentInput, baseUrl string, urlFetcher utils.UrlFetcher, mediaType string, stylesheets []tree.CSS,
	enableHinting, presentationalHints bool, fontConfig *fonts.FontConfiguration, zoom float64, attachments []utils.Attachment) error {
	parsedHtml, err := tree.NewHTML(htmlContent, baseUrl, urlFetcher, mediaType)
	if err != nil {
		return err
	}
	doc := document.Render(*parsedHtml,stylesheets, enableHinting, presentationalHints, fontConfig)
	return pdf.WritePDF(doc, target, zoom, attachments)
}
