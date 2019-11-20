// Implements the backend needed to draw a document, using gofpdf.
package pdf

import (
	"github.com/benoitkugler/go-weasyprint/backend"
	"github.com/benoitkugler/go-weasyprint/matrix"
	"github.com/benoitkugler/gofpdf"
)

type graphicState struct {
	clipNest      int // each ClipXXX increment clipNest by 1
	transformNest int
	alpha         float64
	r, g, b       int
	fillRule      int
}

func newGraphicState(f *gofpdf.Fpdf) graphicState {
	out := graphicState{}
	out.alpha, _ = f.GetAlpha()
	out.r, out.g, out.b = f.GetFillColor()
	return out
}

// Context implements Drawer
type Context struct {
	f *gofpdf.Fpdf

	fileAnnotationsMap map[string]*gofpdf.Attachment
	stack              []graphicState // used to implement Save() / Restore() mechanism
}

func NewContext() Context {
	out := Context{}
	out.f = gofpdf.New("", "", "", "")
	out.fileAnnotationsMap = map[string]*gofpdf.Attachment{}
	out.stack = []graphicState{newGraphicState(out.f)} // start with a basic state
	return out
}

// --------- shortcuts  -----------------------
func (c *Context) currentState() *graphicState {
	return &c.stack[len(c.stack)-1]
}
func (c *Context) previousState() *graphicState {
	return &c.stack[len(c.stack)-2]
}

// return y in gofpdf "order" (origin at top left of the page)
func (c Context) convertY(y float64) float64 {
	_, pageHeight := c.f.GetPageSize()
	return pageHeight - y
}

func (c *Context) Save() backend.StackedDrawer {
	newStack := newGraphicState(c.f)
	c.stack = append(c.stack, newStack)
	return c
}

func (c *Context) Restore() {
	s := c.currentState()
	// Restore Clip
	for i := 0; i < s.clipNest; i += 1 {
		c.f.ClipEnd()
	}
	// Restore Transform
	for i := 0; i < s.transformNest; i += 1 {
		c.f.TransformEnd()
	}
	s = c.previousState()
	c.f.SetAlpha(s.alpha, "Normal")
	c.stack = c.stack[:len(c.stack)-1]
}

func (c *Context) Finish() {
	s := c.currentState()
	// Restore Clip
	for i := 0; i < s.clipNest; i += 1 {
		c.f.ClipEnd()
	}
	// Restore Transform
	for i := 0; i < s.transformNest; i += 1 {
		c.f.TransformEnd()
	}
}

func (c *Context) Paint() {
	w, h := c.f.GetPageSize()
	c.f.Rect(0, 0, w, h, "F")
}

func (c *Context) ClipRectangle(x, y, w, h float64) {
	c.currentState().clipNest += 1
	c.f.ClipRect(x, c.convertY(y), w, h, false)
}

func (c *Context) ClipRoundedRect(x, y, w, h, tl, tr, br, bl float64) {
	c.currentState().clipNest += 1
	c.f.ClipRoundedRectExt(x, y, w, h, tl, tr, br, bl, false)
}

func (c *Context) Translate(tx, ty float64) {
	mt := matrix.Identity()
	mt.Translate(tx, ty)
	c.Transform(mt)
}

func (ct *Context) Transform(mt matrix.Transform) {
	ct.f.TransformBegin()
	a, b, c, d, e, f := mt.Data()
	ct.f.Transform(gofpdf.TransformMatrix{A: a, B: b, C: c, D: d, E: e, F: f})
	ct.currentState().transformNest += 1
}

func (c *Context) OpacityGroup(alpha float64) {
	c.Save()
	c.f.SetAlpha(alpha, "Normal")
}

func convert(v float64) int {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return int(v * 255)
}

func (c *Context) SetSourceRgba(r, g, b, a float64) {
	ri, gi, bi := convert(r), convert(g), convert(b)
	c.f.SetAlpha(a, "Normal")
	c.f.SetFillColor(ri, gi, bi)
	c.f.SetDrawColor(ri, gi, bi)
	c.f.SetTextColor(ri, gi, bi)
}
