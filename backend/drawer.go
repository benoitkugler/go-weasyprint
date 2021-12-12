// Package backend defines a common interface, responsible for graphics primitives.
// This interface is then used to convert a document.Document to the final output.
package backend

import (
	"io"
	"time"

	"github.com/benoitkugler/go-weasyprint/matrix"
	"github.com/benoitkugler/go-weasyprint/style/parser"
	"github.com/benoitkugler/go-weasyprint/utils"
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/pango"
)

type fl = utils.Fl

type Anchor struct {
	Name string
	// Origin at the top-left of the page
	X, Y fl
}

type Attachment struct {
	Title, Description string
	Content            []byte
}

// TextDrawing exposes the positionned text glyphs to draw
// and the associated font, in a backend independent manner
type TextDrawing struct {
	Runs []TextRun

	FontSize utils.Fl
	X, Y     utils.Fl // origin of the text
}

// TextRun is a serie of glyphs with constant font.
type TextRun struct {
	Font   pango.Font
	Glyphs []TextGlyph
}

// TextGlyph stores a glyph and it's position
type TextGlyph struct {
	Glyph    fonts.GID
	Offset   utils.Fl // normalized by FontSize
	Kerning  int      // normalized by FontSize
	XAdvance utils.Fl // how much to move before drawing
}

// GlyphExtents exposes glyph metrics, normalized by the font size.
type GlyphExtents struct {
	Width  int
	Y      int
	Height int
}

// Font stores some metadata used in the output document.
type Font struct {
	Cmap    map[fonts.GID][]rune
	Extents map[fonts.GID]GlyphExtents
	Bbox    [4]int
}

// IsFixedPitch returns true if only one width is used,
// that is if the font is monospaced.
func (f *Font) IsFixedPitch() bool {
	seen := -1
	for _, w := range f.Extents {
		if seen == -1 {
			seen = w.Width
			continue
		}
		if w.Width != seen {
			return false
		}
	}
	return true
}

type GradientInit struct {
	// Kind is either:
	// 	"solid": Colors is then a one element array and Positions and Coords are empty.
	// 	"linear": Coords is (x0, y0, x1, y1)
	// 			  coordinates of the starting and ending points.
	// 	"radial": Coords is (cx0, cy0, radius0, cx1, cy1, radius1)
	// 			  coordinates of the starting end ending circles
	Kind   string
	Coords [6]utils.Fl
}

type GradientLayout struct {
	// list of floats in [0..1].
	// 0 at the starting point, 1 at the ending point.
	Positions []utils.Fl
	Colors    []parser.RGBA

	GradientInit

	// used for ellipses radial gradients. 1 otherwise.
	ScaleY     utils.Fl
	Reapeating bool
}

// BookmarkNode exposes the outline hierarchy of the document
type BookmarkNode struct {
	Label     string
	Children  []BookmarkNode
	Open      bool // state of the outline item
	PageIndex int  // page index (0-based) to link to
	X, Y      fl   // position in the page
}

// Output is the main target to the laid out document,
// in a format agnostic way.
type Output interface {
	// AddPage creates a new page with the given dimensions and returns
	// it to be paint on.
	// The y axis grows downward, meaning bottom > top
	AddPage(left, top, right, bottom fl) OutputPage

	// CreateAnchors register a list of anchors per page, which are named targets of internal links.
	// `anchors` is a 0-based list, meaning anchors in page 1 are at index 0.
	// The origin of internal link has been be added by `OutputPage.AddInternalLink`.
	// `CreateAnchors` is called after all the pages have been created and processed
	CreateAnchors(anchors [][]Anchor)

	// Add global attachments to the file
	SetAttachments(as []Attachment)

	// Embed a file. Calling this method twice with the same id
	// won't embed the content twice.
	// `fileID` will be passed to `OutputPage.AddFileAnnotation`
	EmbedFile(fileID string, a Attachment)

	// Metadatas

	SetTitle(title string)
	SetDescription(description string)
	SetCreator(creator string)
	SetAuthors(authors []string)
	SetKeywords(keywords []string)
	SetProducer(producer string)
	SetDateCreation(d time.Time)
	SetDateModification(d time.Time)

	// SetBookmarks setup the document outline
	SetBookmarks(root []BookmarkNode)
}

// OutputPage is the target of one laid out page
type OutputPage interface {
	// AddInternalLink shows a link on the page, pointing to the
	// named anchor, which will be registered with `Output.CreateAnchors`
	AddInternalLink(xMin, yMin, xMax, yMax fl, anchorName string)

	// AddExternalLink shows a link on the page, pointing to
	// the given url
	AddExternalLink(xMin, yMin, xMax, yMax fl, url string)

	// AddFileAnnotation adds a file annotation on the current page.
	// The file content has been added with `Output.EmbedFile`.
	AddFileAnnotation(xMin, yMin, xMax, yMax fl, fileID string)

	// Adjust the media boxes

	SetMediaBox(left, top, right, bottom fl)
	SetTrimBox(left, top, right, bottom fl)
	SetBleedBox(left, top, right, bottom fl)

	OutputGraphic
}

// RasterImage is an image to be included in the ouput.
type RasterImage struct {
	Content  io.ReadCloser
	MimeType string

	// Rendering is the CSS property for this image.
	Rendering string

	// ID is a unique identifier which permits caching
	// image content when possible.
	ID int
}

// BackgroundImage groups all possible image format for backgrounds,
// like raster image, svg, or gradient
type BackgroundImage interface {
	// Draw shall write the image on the given `context`
	Draw(context GraphicTarget, concreteWidth, concreteHeight fl, imageRendering string)
}

type BackgroundImageOptions struct {
	Rendering                 string // CSS rendering property
	ImageWidth, ImageHeight   fl
	RepeatWidth, RepeatHeight fl
	X, Y                      fl // where to paint the image
}

// OutputGraphic is a surface and the target of graphic operations
type OutputGraphic interface {
	GraphicTarget

	// FillWithImage fills the current path using the given image.
	// Usually, the given image would be painted on an temporary OutputGraphic,
	// which would then be used as fill pattern.
	FillWithImage(img BackgroundImage, options BackgroundImageOptions)
}

type GraphicTarget interface {
	// Returns the current page rectangle
	GetPageRectangle() (left, top, right, bottom fl)

	// OnNewStack save the current graphic stack,
	// execute the given closure, and restore the stack.
	// If an error is encoutered, the stack is still restored
	// and the error is returned
	OnNewStack(func() error) error

	// AddOpacityGroup creates a new drawing target with the given
	// bounding box. The return `OutputGraphic` will be then
	// passed to `DrawOpacityGroup`
	AddOpacityGroup(x, y, width, height fl) OutputGraphic

	// DrawOpacityGroup draw the given target to the main target, applying the given opacity (in [0,1]).
	DrawOpacityGroup(opacity fl, group OutputGraphic)

	// Adds a rectangle
	// of the given size to the current path,
	// at position ``(x, y)`` in user-space coordinates.
	// (X,Y) coordinates are the top left corner of the rectangle.
	Rectangle(x fl, y fl, width fl, height fl)

	// Establishes a new clip region
	// by intersecting the current clip region
	// with the current path as it would be filled by `Fill`
	// and according to the fill rule given in `evenOdd`.
	//
	// After `Clip`, the current path will be cleared (or closed).
	//
	// The current clip region affects all drawing operations
	// by effectively masking out any changes to the surface
	// that are outside the current clip region.
	//
	// Calling `Clip` can only make the clip region smaller,
	// never larger, but you can call it in the `OnNewStack` closure argument,
	// so that the original clip region is restored afterwards.
	Clip(evenOdd bool)

	// Sets the color which will be used for any subsequent drawing operation.
	//
	// The color and alpha components are
	// floating point numbers in the range 0 to 1.
	// If the values passed in are outside that range, they will be clamped.
	// `stroke` controls whether stroking or filling operations are concerned.
	SetColorRgba(color parser.RGBA, stroke bool)

	// Set current alpha
	// `stroke` controls whether stroking or filling operations are concerned.
	SetAlpha(alpha fl, stroke bool)

	// Sets the current line.
	// The line width value specifies the diameter of a pen
	// that is circular in user space,
	// (though device-space pen may be an ellipse in general
	// due to scaling / shear / rotation of the CTM).
	//
	// When the description above refers to user space and CTM
	// it refers to the user space and CTM in effect
	// at the time of the stroking operation,
	// not the user space and CTM in effect
	// at the time of the call to `SetLineWidth`.
	// The simplest usage makes both of these spaces identical.
	// That is, if there is no change to the CTM
	// between a call to `SetLineWidth`
	// and the stroking operation,
	// then one can just pass user-space values to `SetLineWidth`
	// and ignore this note.
	//
	// As with the other stroke parameters,
	// the current line cap style is examined by
	// `Stroke` but does not have any effect during path construction.
	SetLineWidth(width fl)

	// Sets the dash pattern to be used by Stroke.
	// A dash pattern is specified by dashes, a list of positive values.
	// Each value provides the length of alternate "on" and "off"
	// portions of the stroke.
	// `offset` specifies an offset into the pattern
	// at which the stroke begins.
	//
	// Each "on" segment will have caps applied
	// as if the segment were a separate sub-path.
	// In particular, it is valid to use an "on" length of 0
	// with `LINE_CAP_ROUND` or `LINE_CAP_SQUARE`
	// in order to distribute dots or squares along a path.
	//
	// Note: The length values are in user-space units
	// as evaluated at the time of stroking.
	// This is not necessarily the same as the user space
	// at the time of SetDash.
	//
	// If `dashes` is empty dashing is disabled.
	// If it is of length 1 a symmetric pattern is assumed
	// with alternating on and off portions of the size specified
	// by the single value.
	SetDash(dashes []fl, offset fl)

	// A drawing operator that fills the current path
	// according to the current fill rule,
	// (each sub-path is implicitly closed before being filled).
	// After `fill`, the current path will is cleared
	Fill(evenOdd bool)

	// A drawing operator that strokes the current path
	// according to the current line width, line join, line cap,
	// and dash settings.
	// After `Stroke`, the current path will be cleared.
	Stroke()

	// Modifies the current transformation matrix (CTM)
	// by applying `mt` as an additional transformation.
	// The new transformation of user space takes place
	// after any existing transformation.
	Transform(mt matrix.Transform)

	// Begin a new sub-path.
	// After this call the current point will be ``(x, y)``.
	//
	// :param x: X position of the new point.
	// :param y: Y position of the new point.
	MoveTo(x fl, y fl)

	// Adds a line to the path from the current point
	// to position ``(x, y)`` in user-space coordinates.
	// After this call the current point will be ``(x, y)``.
	// A current point must be defined before using this method.
	LineTo(x fl, y fl)

	// Add cubic Bézier curve to current path.
	// The curve shall extend to ``(x3, y3)`` using ``(x1, y1)`` and ``(x2,
	// y2)`` as the Bézier control points.
	CurveTo(x1, y1, x2, y2, x3, y3 fl)

	// DrawText draws the given text using the current fill color.
	// The fonts of the runs have been registred with `AddFont`.
	DrawText(TextDrawing)

	// AddFont register a new font to be used in the output and return
	// an object used to store associated metadata.
	// This method will be called several times with the same `face` argument,
	// so caching is advised.
	AddFont(font pango.Font, content []byte) *Font

	// DrawRasterImage draws the given image at the current point, with the given dimensions.
	DrawRasterImage(img RasterImage, width, height fl)

	// DrawGradient draws the given gradient at the current point.
	// Solid gradient are already handled, meaning that only linear and radial
	// must be taken care of.
	DrawGradient(gradient GradientLayout, width, height fl)
}
