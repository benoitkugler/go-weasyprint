// Implements the backend needed to draw a document, using gofpdf.
package pdf

import (
	"math"

	"github.com/benoitkugler/go-weasyprint/backend"
	"github.com/benoitkugler/go-weasyprint/matrix"
	"github.com/benoitkugler/gofpdf"
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

// Context implements Drawer
type Context struct {
	f *gofpdf.Fpdf

	fileAnnotationsMap map[string]*gofpdf.Attachment

	matrixToPdf, matrixToPdfInv matrix.Transform

	fillRule int
}

func NewContext() Context {
	out := Context{}
	out.f = gofpdf.New("", "", "", "")
	out.f.SetFont("Helvetica", "", 15)
	out.fileAnnotationsMap = map[string]*gofpdf.Attachment{}
	out.matrixToPdf, out.matrixToPdfInv = out.convertionMatrix()
	return out
}

// return y in gofpdf "order" (origin at top left of the page)
func (c Context) convertY(y float64) float64 {
	_, pageHeight := c.f.GetPageSize()
	return pageHeight - y
}

func (c Context) GetPageSize() (width, height float64) {
	return c.f.GetPageSize()
}

func (c *Context) OnNewStack(f func() error) error {
	c.f.Save()
	err := f()
	c.f.Restore()
	return err
}

func (c *Context) Finish() {

}

func (c Context) Fill() {
	s := "f"
	if c.fillRule == backend.FillRuleEvenOdd {
		s = "f*"
	}
	c.f.DrawPath(s)
}

func (c Context) Stroke() {
	c.f.DrawPath("S")
}

func (c *Context) Paint() {
	w, h := c.f.GetPageSize()
	c.f.Rect(0, 0, w, h, "F")
}

func (c Context) Clip() {
	c.f.ClipPath(c.fillRule == backend.FillRuleEvenOdd)
}

func (c *Context) Translate(tx, ty float64) {
	c.Transform(matrix.Translation(tx, ty))
}

func (c *Context) Scale(sx, sy float64) {
	c.Transform(matrix.Scaling(sx, sy))
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
	c.f.RawTransform(toTransformMatrix(mt))
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

func (c Context) SetSourceRgba(r, g, b, a float64) {
	ri, gi, bi := convert(r), convert(g), convert(b)
	c.f.SetAlpha(a, "Normal")
	c.f.SetFillColor(ri, gi, bi)
	c.f.SetDrawColor(ri, gi, bi)
	c.f.SetTextColor(ri, gi, bi)
}
func (c *Context) SetAlpha(a float64) {
	c.f.SetAlpha(a, "Normal")
}

func (c *Context) SetFillRule(r int) {
	c.fillRule = r
}

func (c Context) SetLineWidth(w float64) {
	c.f.SetLineWidth(w)
}

func (c Context) SetDash(dashes []float64, offset float64) {
	c.f.SetDashPattern(dashes, offset)
}

// Paths

// NewPath only clears the current path.
func (c Context) NewPath() {
	c.f.RawWriteStr(" n")
}

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

func (c Context) Rectangle(x, y, w, h float64) {
	c.f.RectPath(x, c.convertY(y), w, h)
}

func (c Context) RoundedRect(x, y, w, h, tl, tr, br, bl float64) {
	c.f.RoundedRectPath(x, y, w, h, tl, tr, br, bl)
}

func (c Context) Arc(xc, yc, radius, angle1, angle2 float64) {
	// in degrees
	angle1, angle2 = angle1*180/math.Pi, angle2*180/math.Pi
	// to draw a "positive"  arc, we need to ensure that angle 2  >= angle 1
	for angle2 < angle1 {
		angle2 += 360
	}
	c.f.ArcTo(xc, c.convertY(yc), radius, radius, 0, angle1, angle2)
}
func (c Context) ArcNegative(xc, yc, radius, angle1, angle2 float64) {
	// in degrees
	angle1, angle2 = angle1*180/math.Pi, angle2*180/math.Pi
	// to draw a "negative"  arc, we need to ensure that angle1  >= angle 2
	for angle1 < angle2 {
		angle2 -= 360
	}
	c.f.ArcTo(xc, c.convertY(yc), radius, radius, 0, angle1, angle2)
}
