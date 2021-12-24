package document

import (
	"log"
	"os"
	"time"

	"github.com/benoitkugler/go-weasyprint/backend"
	"github.com/benoitkugler/go-weasyprint/matrix"
	"github.com/benoitkugler/go-weasyprint/style/parser"
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/pango"
)

// implements a no-op backend, which can be used to test for crashes

var outputLog = log.New(os.Stdout, "output: ", log.Ltime)

// var outputLog = log.New(io.Discard, "output: ", log.Ltime)

var _ backend.Output = output{}

type output struct{}

func (output) AddPage(left, top, right, bottom fl) backend.OutputPage {
	outputLog.Println("AddPage")
	return outputPage{}
}

func (output) CreateAnchors(anchors [][]backend.Anchor) {
	outputLog.Println("CreateAnchors")
}

func (output) SetAttachments(as []backend.Attachment) {
	outputLog.Println("SetAttachments")
}

func (output) EmbedFile(id string, a backend.Attachment) {
	outputLog.Println("EmbedFile")
}

func (output) SetTitle(title string) {
	outputLog.Println("SetTitle")
}

func (output) SetDescription(description string) {
	outputLog.Println("SetDescription")
}

func (output) SetCreator(creator string) {
	outputLog.Println("SetCreator")
}

func (output) SetAuthors(authors []string) {
	outputLog.Println("SetAuthors")
}

func (output) SetKeywords(keywords []string) {
	outputLog.Println("SetKeywords")
}

func (output) SetProducer(producer string) {
	outputLog.Println("SetProducer")
}

func (output) SetDateCreation(d time.Time) {
	outputLog.Println("SetDateCreation")
}

func (output) SetDateModification(d time.Time) {
	outputLog.Println("SetDateModification")
}

func (output) SetBookmarks([]backend.BookmarkNode) {
	outputLog.Println("AddBookmark")
}

type outputPage struct{}

func (outputPage) AddInternalLink(x, y, w, h fl, anchorName string) {
	outputLog.Println("AddInternalLink")
}

func (outputPage) AddExternalLink(x, y, w, h fl, url string) {
	outputLog.Println("AddExternalLink")
}

func (outputPage) AddFileAnnotation(x, y, w, h fl, id string) {
	outputLog.Println("AddFileAnnotation")
}

func (outputPage) GetPageRectangle() (left, top, right, bottom fl) {
	outputLog.Println("GetPageRectangle")
	return 0, 0, 10, 10
}

func (outputPage) SetMediaBox(left, top, right, bottom fl) {
	outputLog.Println("SetTrimBox")
}

func (outputPage) SetTrimBox(left, top, right, bottom fl) {
	outputLog.Println("SetTrimBox")
}

func (outputPage) SetBleedBox(left, top, right, bottom fl) {
	outputLog.Println("SetBleedBox")
}

func (outputPage) OnNewStack(f func()) {
	outputLog.Println("OnNewStack")
	f()
}

func (outputPage) Rectangle(x fl, y fl, width fl, height fl) {
	outputLog.Println("Rectangle")
}

func (outputPage) Clip(evenOdd bool) {
	outputLog.Println("Clip")
}

func (outputPage) SetColorRgba(color parser.RGBA, stroke bool) {
	outputLog.Println("SetColorRgba")
}

func (outputPage) SetAlpha(alpha fl, stroke bool) {
	outputLog.Println("SetAlpha")
}

func (outputPage) SetLineWidth(width fl) {
	outputLog.Println("SetLineWidth")
}

func (outputPage) SetDash(dashes []fl, offset fl) {
	outputLog.Println("SetDash")
}

func (outputPage) Fill(evenOdd bool) {
	outputLog.Println("Fill")
}

func (outputPage) FillWithImage(backend.Image, backend.BackgroundImageOptions) {
	outputLog.Println("Fill")
}

func (outputPage) Stroke() {
	outputLog.Println("Stroke")
}

func (outputPage) Transform(mt matrix.Transform) {
	outputLog.Println("Transform")
}

func (outputPage) MoveTo(x fl, y fl) {
	outputLog.Println("MoveTo")
}

func (outputPage) LineTo(x fl, y fl) {
	outputLog.Println("LineTo")
}

func (outputPage) CurveTo(x1, y1, x2, y2, x3, y3 fl) {
	outputLog.Println("CurveTo")
}

func (outputPage) DrawText(text backend.TextDrawing) {
	outputLog.Println("DrawText", text)
}

func (outputPage) AddFont(pango.Font, []byte) *backend.Font {
	outputLog.Println("AddFont")
	return &backend.Font{Cmap: make(map[fonts.GID][]rune), Extents: make(map[fonts.GID]backend.GlyphExtents)}
}

func (outputPage) DrawRasterImage(img backend.RasterImage, width, height fl) {
	outputLog.Println("DrawRasterImage")
}

func (outputPage) DrawGradient(gradient backend.GradientLayout, width, height fl) {
	outputLog.Println("DrawGradient")
}

func (outputPage) AddOpacityGroup(x, y, width, height fl) backend.OutputGraphic {
	outputLog.Println("AddGroup")
	return outputPage{}
}

func (outputPage) DrawOpacityGroup(opacity fl, group backend.OutputGraphic) {
	outputLog.Println("DrawGroup")
}
