package pdf

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/benoitkugler/go-weasyprint/style/parser"
)

func finishAndSave(c *Output, t *testing.T) {
	doc := c.Finalize()
	err := doc.WriteFile("test.pdf", nil)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Saved in test.pdf")
}

func TestPaint(t *testing.T) {
	c := NewOutput()
	page := c.AddPage(0, 200, 100, 0)
	page.SetColorRgba(parser.RGBA{R: 0, G: 1, B: 0, A: 1}, true)
	page.SetColorRgba(parser.RGBA{R: 0, G: 1, B: 1, A: 1}, false)
	// pdf.SetFont("Helvetica", "", 15)
	// pdf.AddPage()
	page.Rectangle(20, 20, 30, 30)
	page.SetLineWidth(2)
	page.Stroke()
	page.OnNewStack(func() error {
		page.Rectangle(20, 20, 30, 30)
		// page.Clip(false)
		page.Fill(false)
		return nil
	})
	finishAndSave(c, t)
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// func drawImage(f *gofpdf.Fpdf, w, h float64) {
// 	n := rand.Intn(256)
// 	f.SetFillColor(n/2+100, 255-n, n)
// 	f.Rect(0, 0, w, h, "F")
// }

// func TestRepeat(t *testing.T) {
// 	c := NewContext()
// 	c.f.AddPage()
// 	c.f.SetXY(20, 20)
// 	w, h := 15., 12.
// 	// m := matrix.Identity()
// 	// m.Rotate(-1)
// 	// c.Transform(m)
// 	_, maxH := c.f.GetPageSize()
// 	nbx := 1
// 	nby := int(maxH / h)
// 	for i := 0; i < nbx; i += 1 {
// 		c.OnNewStack(func() error {
// 			for j := 0; j < nby; j += 1 {
// 				drawImage(c.f, w, h)
// 				c.Translate(0, h)
// 			}
// 			return nil
// 		})
// 		c.Translate(w, 0)
// 	}
// 	finishAndSave(c, t)
// }

// func TestTransform(t *testing.T) {
// 	c := NewContext()
// 	c.f.AddPage()
// 	c.f.TransformBegin()
// 	c.f.TransformTranslate(10, 20)
// 	drawImage(c.f, 30, 30)
// 	c.f.TransformEnd()

// 	_, h := c.f.GetPageSize()
// 	k := c.f.GetConversionRatio()
// 	fmt.Println(h, k)
// 	tr := matrix.Identity()
// 	tr.Translate(100, 20)
// 	conv := matrix.New(k, 0, 0, -k, 0, h*k)
// 	convInv := conv
// 	err := convInv.Invert()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	fmt.Println(matrix.Mul(conv, convInv))
// 	tr2 := matrix.Mul(matrix.Mul(conv, tr), convInv)
// 	fmt.Println(tr, tr2)
// 	c.f.TransformBegin()
// 	c.f.Transform(toTransformMatrix(tr2))
// 	drawImage(c.f, 30, 30)
// 	c.f.TransformEnd()

// 	finishAndSave(c, t)
// }

// func TestArc(t *testing.T) {
// 	c := NewContext()
// 	c.f.AddPage()
// 	c.f.SetDrawColor(0xff, 0x00, 0x00)
// 	c.f.Arc(20, 20, 10, 10, 0, 10, 370, "D")
// 	finishAndSave(c, t)
// }

// func TestImage(t *testing.T) {
// 	// open ".test/france_belgique.jpeg"
// 	file, err := os.Open("../.test/france_belgique.jpeg")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// decode jpeg into image.Image
// 	img, err := jpeg.Decode(file)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	file.Close()

// 	// resize to width 1000 using Lanczos resampling
// 	// and preserve aspect ratio
// 	m := imaging.Resize(img, 1000, 0, imaging.Lanczos)
// 	m = imaging.Resize(img, 1000, 0, imaging.NearestNeighbor)
// 	out, err := os.Create("../.test/france_belgique_resized.jpeg")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer out.Close()

// 	// write new image to file
// 	jpeg.Encode(out, m, nil)
// }
