// Implements the backend needed to draw a document, using gofpdf.
package pdf

import (
	"github.com/benoitkugler/go-weasyprint/backend"
	"github.com/benoitkugler/gofpdf"
)

type graphicState struct {
	clipNest int // each ClipXXX increment clipNest by 1
}

// Context implements Drawer
type Context struct {
	f *gofpdf.Fpdf

	fileAnnotationsMap map[string]*gofpdf.Attachment
	stack              []graphicState // used to implement Save() / Restore() mechanism
}

func NewContext() Context {
	out := Context{}
	out.fileAnnotationsMap = map[string]*gofpdf.Attachment{}
	out.stack = []graphicState{{}} // start with a basic state
	return out
}

// --------- shortcuts  -----------------------
func (c *Context) currentState() *graphicState {
	return &c.stack[len(c.stack)-1]
}

// return y in gofpdf "order" (origin at top left of the page)
func (c Context) convertY(y float64) float64 {
	_, pageHeight := c.f.GetPageSize()
	return pageHeight - y
}

func (c *Context) SaveStack() backend.StackedDrawer {
	newStack := graphicState{}
	c.stack = append(c.stack, newStack)
	return c
}

func (c *Context) Restore() {
	s := c.currentState()

	// Restore Clip
	for i := 0; i < s.clipNest; i += 1 {
		c.f.ClipEnd()
	}
	c.stack = c.stack[:len(c.stack)-1]
}

func (c *Context) ClipRectangle(x, y, w, h float64) {
	c.currentState().clipNest += 1
	c.f.ClipRect(x, c.convertY(y), w, h, false)
}
