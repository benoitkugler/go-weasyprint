package pdf

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/benoitkugler/go-weasyprint/matrix"

	"github.com/benoitkugler/gofpdf"
)

func TestPaint(t *testing.T) {
	c := NewContext()
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetDrawColor(0xff, 0x00, 0x00)
	pdf.SetFillColor(0x99, 0x99, 0x99)
	pdf.SetFont("Helvetica", "", 15)
	pdf.AddPage()
	c.f = pdf
	c.OnNewStack(func() {
		c.RoundedRect(20, 20, 30, 30, 5, 5, 5, 5)
		c.Clip()
		c.Paint()
	})
	finishAndSave(c, t)
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func drawImage(f *gofpdf.Fpdf, w, h float64) {
	n := rand.Intn(256)
	f.SetFillColor(n/2+100, 255-n, n)
	f.Rect(0, 0, w, h, "F")
}

func TestRepeat(t *testing.T) {
	c := NewContext()
	c.f.AddPage()
	c.f.SetXY(20, 20)
	w, h := 15., 12.
	// m := matrix.Identity()
	// m.Rotate(-1)
	// c.Transform(m)
	_, maxH := c.f.GetPageSize()
	nbx := 1
	nby := int(maxH / h)
	for i := 0; i < nbx; i += 1 {
		c.OnNewStack(func() {
			for j := 0; j < nby; j += 1 {
				drawImage(c.f, w, h)
				c.Translate(0, h)
			}
		})
		c.Translate(w, 0)
	}
	finishAndSave(c, t)
}

func finishAndSave(c Context, t *testing.T) {
	c.Finish()
	err := c.f.OutputFileAndClose("test.pdf")
	if err != nil {
		t.Fatal(err)
	}
}

func TestTransform(t *testing.T) {
	c := NewContext()
	c.f.AddPage()
	c.f.TransformBegin()
	c.f.TransformTranslate(10, 20)
	drawImage(c.f, 30, 30)
	c.f.TransformEnd()

	_, h := c.f.GetPageSize()
	k := c.f.GetConversionRatio()
	fmt.Println(h, k)
	tr := matrix.Identity()
	tr.Translate(100, 20)
	conv := matrix.New(k, 0, 0, -k, 0, h*k)
	convInv := conv
	err := convInv.Invert()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(matrix.Mul(conv, convInv))
	tr2 := matrix.Mul(matrix.Mul(conv, tr), convInv)
	fmt.Println(tr, tr2)
	c.f.TransformBegin()
	c.f.Transform(toTransformMatrix(tr2))
	drawImage(c.f, 30, 30)
	c.f.TransformEnd()

	finishAndSave(c, t)
}
