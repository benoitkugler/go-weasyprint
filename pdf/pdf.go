// Implements the backend needed to draw a document, using gofpdf.
package pdf

import (
	"github.com/benoitkugler/gofpdf"
)

// Context implements Drawer
type Context struct {
	f *gofpdf.Fpdf

	fileAnnotationsMap map[string]*gofpdf.Attachment
}

func NewContext() Context {
	out := Context{}
	out.fileAnnotationsMap = map[string]*gofpdf.Attachment{}
	return out
}
