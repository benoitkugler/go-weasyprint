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

	matrixToPdf, matrixToPdfInv matrix.Transform
}

func NewContext() Context {
	out := Context{}
	out.f = gofpdf.New("", "", "", "")
	out.f.SetFont("Helvetica", "", 15)
	out.fileAnnotationsMap = map[string]*gofpdf.Attachment{}
	out.stack = []graphicState{newGraphicState(out.f)} // start with a basic state
	out.matrixToPdf, out.matrixToPdfInv = out.convertionMatrix()
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
	c.Transform(matrix.Translation(tx, ty))
}

func (c *Context) convertionMatrix() (M, Minv matrix.Transform) {
	_, h := c.f.GetPageSize()
	k := c.f.GetConversionRatio()
	conv := matrix.New(k, 0, 0, -k, 0, h*k)
	convInv := conv
	err := convInv.Invert()
	if err != nil {
		c.f.SetError(err)
	}
	return conv, convInv
}

func toTransformMatrix(mt matrix.Transform) gofpdf.TransformMatrix {
	a, b, c, d, e, f := mt.Data()
	return gofpdf.TransformMatrix{A: a, B: b, C: c, D: d, E: e, F: f}
}

func (c *Context) Transform(mt matrix.Transform) {
	mt = matrix.Mul(matrix.Mul(c.matrixToPdf, mt), c.matrixToPdfInv)
	c.f.TransformBegin()
	c.f.Transform(toTransformMatrix(mt))
	c.currentState().transformNest += 1
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

func (c *Context) SetFillRule(r int) {
	c.currentState().fillRule = r
}

// Paths

func (c Context) MoveTo(x, y float64) {
	c.f.MoveTo(x, c.convertY(y))
}
func (c Context) LineTo(x, y float64) {
	c.f.LineTo(x, c.convertY(y))
}
func (c Context) RelLineTo(dx, dy float64) {
	x, y := c.f.GetXY()
	c.LineTo(x+dx, c.convertY(y)+dy)
}
