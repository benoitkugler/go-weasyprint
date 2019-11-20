package pdf

import (
	"fmt"
	"math/rand"
	"testing"

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
	c.ClipRoundedRect(20, 20, 30, 30, 5, 5, 5, 5)
	c.Paint()
	c.Finish()
	err := pdf.OutputFileAndClose("test.pdf")
	if err != nil {
		t.Fatal(err)
	}
}

func drawImage(f *gofpdf.Fpdf, w, h float64) {
	n := rand.Intn(256)
	fmt.Println(n)
	f.SetFillColor(20, 20, n)
	f.Rect(0, 0, w, h, "F")
}

func TestRepeat(t *testing.T) {
	c := NewContext()
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Helvetica", "", 15)
	pdf.AddPage()
	c.f = pdf
	pdf.SetXY(20, 20)
	w, h := 60., 10.
	// m := matrix.Identity()
	// m.Rotate(-1)
	// c.Transform(m)
	max, _ := pdf.GetPageSize()
	nb := int(max / w)
	for i := 0; i < nb; i += 1 {
		drawImage(pdf, w, h)
		c.Translate(w, 0)
	}
	c.Finish()
	err := pdf.OutputFileAndClose("test.pdf")
	if err != nil {
		t.Fatal(err)
	}
}
