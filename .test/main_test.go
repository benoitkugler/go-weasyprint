package test

import (
	"testing"

	"github.com/benoitkugler/gofpdf"
)

func TestClip(t *testing.T) {
	pdf := gofpdf.New(gofpdf.OrientationPortrait, "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.ClipCircle(100, 100, 50, true)
	w, h 
	pdf.Rect(0, 0, 1000, 1000, "F")
	pdf.ClipEnd()
	err := pdf.OutputFileAndClose("test.pdf")
	if err != nil {
		t.Fatal(err)
	}
}
