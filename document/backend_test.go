package document

import (
	"image"
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

type outputPage struct {
	pattern
}

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

func (outputPage) SetTrimBox(left, top, right, bottom fl) {
	outputLog.Println("SetTrimBox")
}

func (outputPage) SetBleedBox(left, top, right, bottom fl) {
	outputLog.Println("SetBleedBox")
}

func (outputPage) OnNewStack(f func() error) error {
	outputLog.Println("OnNewStack")
	return f()
}

func (outputPage) AddPattern(width, height, repeatWidth, repeatHeight fl, mt matrix.Transform) backend.Pattern {
	outputLog.Println("AddPattern")
	return pattern{}
}

func (outputPage) SetColorPattern(pattern backend.Pattern, stroke bool) {
	outputLog.Println("SetColorPattern")
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

func (outputPage) Stroke() {
	outputLog.Println("Stroke")
}

func (outputPage) Transform(mt matrix.Transform) {
	outputLog.Println("Transform")
}

func (outputPage) GetTransform() matrix.Transform {
	outputLog.Println("GetTransform")
	return matrix.Identity()
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

func (outputPage) AddFont(face fonts.Face, content []byte) *backend.Font {
	outputLog.Println("AddFont")
	return &backend.Font{Cmap: make(map[pango.Glyph][]rune), Widths: make(map[pango.Glyph]int)}
}

func (outputPage) DrawRasterImage(img image.Image, imageRendering string, width, height fl) {
	outputLog.Println("DrawRasterImage")
}

func (outputPage) DrawGradient(gradient backend.GradientLayout, width, height fl) {
	outputLog.Println("DrawGradient")
}

type pattern struct{}

func (pattern) AddGroup(x, y, width, height fl) backend.OutputGraphic {
	outputLog.Println("AddGroup")
	return outputPage{}
}

func (pattern) DrawGroup(group backend.OutputGraphic) {
	outputLog.Println("DrawGroup")
}
