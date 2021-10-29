// Package pdf implements the backend needed to draw a document, using github.com/benoitkugler/pdf.
package pdf

import (
	"github.com/benoitkugler/go-weasyprint/backend"
	"github.com/benoitkugler/pdf/model"
)

// type graphicState struct {
// 	clipNest      int // each ClipXXX increment clipNest by 1
// 	transformNest int
// 	alpha         float64
// 	r, g, b       int
// 	fillRule      int
// }

// func newGraphicState(f *gofpdf.Fpdf) graphicState {
// 	out := graphicState{}
// 	out.alpha, _ = f.GetAlpha()
// 	out.r, out.g, out.b = f.GetFillColor()
// 	return out
// }

var (
	_ backend.Output     = (*Context)(nil)
	_ backend.OutputPage = (*contextPage)(nil)
)

// Context implements backend.Output
type Context struct {
	// global map for files embedded in the PDF
	// and used in file annotations
	embeddedFiles map[string]*model.FileSpec

	document model.Document

	// temporary content, will be copied in the document (see Finalize)
	pages []*model.PageObject

	// f *gofpdf.Fpdf

	// fileAnnotationsMap map[string]*gofpdf.Attachment

	// matrixToPdf, matrixToPdfInv matrix.Transform

	// fillRule int
}

func NewContext() *Context {
	out := Context{
		embeddedFiles: make(map[string]*model.FileSpec),
	}
	return &out
}

func (c *Context) AddPage(left, top, right, bottom fl) backend.OutputPage {
	out := newContextPage(left, top, right, bottom, c.embeddedFiles)
	c.pages = append(c.pages, &out.page)
	return out
}
