package pdf

import (
	"image"

	"github.com/benoitkugler/go-weasyprint/backend"
	"github.com/benoitkugler/go-weasyprint/matrix"
	"github.com/benoitkugler/go-weasyprint/style/parser"
	"github.com/benoitkugler/go-weasyprint/utils"
	"github.com/benoitkugler/pdf/model"
	"github.com/benoitkugler/textlayout/fonts"
)

type fl = utils.Fl

// contextPage implements backend.OutputPage
type contextPage struct {
	page model.PageObject

	embeddedFiles map[string]*model.FileSpec
	pageRectangle [4]fl
}

func newContextPage(left, top, right, bottom fl,
	embeddedFiles map[string]*model.FileSpec) *contextPage {
	out := &contextPage{
		pageRectangle: [4]fl{left, top, right, bottom},
		embeddedFiles: embeddedFiles,
	}
	return out
}

func (cp *contextPage) AddInternalLink(xMin, yMin, xMax, yMax fl, anchorName string) {
	an := model.AnnotationDict{
		BaseAnnotation: model.BaseAnnotation{
			Rect: model.Rectangle{xMin, yMin, xMax, yMax},
		},
		Subtype: model.AnnotationLink{
			BS:   &model.BorderStyle{W: model.ObjFloat(0)},
			Dest: model.DestinationString(anchorName),
		},
	}
	cp.page.Annots = append(cp.page.Annots, &an)
}

func (cp *contextPage) AddExternalLink(xMin, yMin, xMax, yMax fl, url string) {
	an := model.AnnotationDict{
		BaseAnnotation: model.BaseAnnotation{
			Rect: model.Rectangle{xMin, yMin, xMax, yMax},
		},
		Subtype: model.AnnotationLink{
			BS: &model.BorderStyle{W: model.ObjFloat(0)},
			A:  model.Action{ActionType: model.ActionURI{URI: url}},
		},
	}
	cp.page.Annots = append(cp.page.Annots, &an)
}

// Add file annotation on the current page
func (cp *contextPage) AddFileAnnotation(xMin, yMin, xMax, yMax fl, fileID string) {
	rect := model.Rectangle{xMin, yMin, xMax, yMax}
	an := model.AnnotationDict{
		BaseAnnotation: model.BaseAnnotation{
			Rect: rect,
			AP: &model.AppearanceDict{
				N: model.AppearanceEntry{"": &model.XObjectForm{
					BBox: rect,
				}},
			},
		},
		Subtype: model.AnnotationFileAttachment{
			FS: cp.embeddedFiles[fileID],
		},
	}
	cp.page.Annots = append(cp.page.Annots, &an)
}

// Returns the current page rectangle
func (cp *contextPage) GetPageRectangle() (left fl, top fl, right fl, bottom fl) {
	return cp.pageRectangle[0], cp.pageRectangle[1], cp.pageRectangle[2], cp.pageRectangle[3]
}

// Adjust the media boxes
func (cp *contextPage) SetTrimBox(left fl, top fl, right fl, bottom fl) {
	cp.page.TrimBox = &model.Rectangle{left, top, right, bottom}
}

func (cp *contextPage) SetBleedBox(left fl, top fl, right fl, bottom fl) {
	cp.page.BleedBox = &model.Rectangle{left, top, right, bottom}
}

// OnNewStack save the current graphic stack,
// execute the given closure, and restore the stack.
// If an error is encoutered, the stack is still restored
// and the error is returned
func (cp *contextPage) OnNewStack(_ func() error) error {
	panic("not implemented") // TODO: Implement
}

// AddGroup creates a new drawing target with the given
// bounding box.
// If the backend does not support groups, the current target should be returned.
func (cp *contextPage) AddGroup(x fl, y fl, width fl, height fl) backend.OutputPage {
	panic("not implemented") // TODO: Implement
}

// DrawGroup draw the given target to the main target.
// If the backend does not support groups,  this should be a no-op.
func (cp *contextPage) DrawGroup(group backend.OutputPage) {
	panic("not implemented") // TODO: Implement
}

func (cp *contextPage) AddPattern(width fl, height fl, repeatWidth fl, repeatHeight fl, mt matrix.Transform) backend.Pattern {
	panic("not implemented") // TODO: Implement
}

func (cp *contextPage) SetColorPattern(pattern backend.Pattern, stroke bool) {
	panic("not implemented") // TODO: Implement
}

// Adds a rectangle
// of the given size to the current path,
// at position ``(x, y)`` in user-space coordinates.
// (X,Y) coordinates are the top left corner of the rectangle.
func (cp *contextPage) Rectangle(x fl, y fl, width fl, height fl) {
	panic("not implemented") // TODO: Implement
}

// Establishes a new clip region
// by intersecting the current clip region
// with the current path as it would be filled by `Fill`
// and according to the fill rule given in `evenOdd`.
//
// After `Clip`, the current path will be cleared.
//
// The current clip region affects all drawing operations
// by effectively masking out any changes to the surface
// that are outside the current clip region.
//
// Calling `Clip` can only make the clip region smaller,
// never larger, but you can call it in the `OnNewStack` closure argument.
func (cp *contextPage) Clip(evenOdd bool) {
	panic("not implemented") // TODO: Implement
}

// Sets the color which will be used for any subsequent drawing operation.
//
// The color and alpha components are
// floating point numbers in the range 0 to 1.
// If the values passed in are outside that range, they will be clamped.
// `stroke` controls whether stroking or filling operations are concerned.
func (cp *contextPage) SetColorRgba(color parser.RGBA, stroke bool) {
	panic("not implemented") // TODO: Implement
}

// Set current alpha
// `stroke` controls whether stroking or filling operations are concerned.
func (cp *contextPage) SetAlpha(alpha fl, stroke bool) {
	panic("not implemented") // TODO: Implement
}

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
// at the time of the call to :meth:`set_line_width`.
// The simplest usage makes both of these spaces identical.
// That is, if there is no change to the CTM
// between a call to :meth:`set_line_width`
// and the stroking operation,
// then one can just pass user-space values to :meth:`set_line_width`
// and ignore this note.
//
// As with the other stroke parameters,
// the current line cap style is examined by
// `Stroke` but does not have any effect during path construction.
func (cp *contextPage) SetLineWidth(width fl) {
	panic("not implemented") // TODO: Implement
}

// Sets the dash pattern to be used by :meth:`stroke`.
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
// in order to distributed dots or squares along a path.
//
// Note: The length values are in user-space units
// as evaluated at the time of stroking.
// This is not necessarily the same as the user space
// at the time of :meth:`set_dash`.
//
// If `dashes` is empty dashing is disabled.
// If it is of length 1 a symmetric pattern is assumed
// with alternating on and off portions of the size specified
// by the single value.
func (cp *contextPage) SetDash(dashes []fl, offset fl) {
	panic("not implemented") // TODO: Implement
}

// A drawing operator that fills the current path
// according to the current fill rule,
// (each sub-path is implicitly closed before being filled).
// After `fill`, the current path will is cleared
func (cp *contextPage) Fill(evenOdd bool) {
	panic("not implemented") // TODO: Implement
}

// A drawing operator that strokes the current path
// according to the current line width, line join, line cap,
// and dash settings.
// After `Stroke`, the current path will be cleared.
func (cp *contextPage) Stroke() {
	panic("not implemented") // TODO: Implement
}

// Modifies the current transformation matrix (CTM)
// by applying `mt` as an additional transformation.
// The new transformation of user space takes place
// after any existing transformation.
// `Restore` or `Finish` should be called.
func (cp *contextPage) Transform(mt matrix.Transform) {
	panic("not implemented") // TODO: Implement
}

// GetTransform returns the current transformation matrix.
func (cp *contextPage) GetTransform() matrix.Transform {
	panic("not implemented") // TODO: Implement
}

// Begin a new sub-path.
// After this call the current point will be ``(x, y)``.
//
// :param x: X position of the new point.
// :param y: Y position of the new point.
func (cp *contextPage) MoveTo(x fl, y fl) {
	panic("not implemented") // TODO: Implement
}

// Adds a line to the path from the current point
// to position ``(x, y)`` in user-space coordinates.
// After this call the current point will be ``(x, y)``.
// A current point must be defined before using this method.
func (cp *contextPage) LineTo(x fl, y fl) {
	panic("not implemented") // TODO: Implement
}

// Add cubic Bézier curve to current path.
// The curve shall extend to ``(x3, y3)`` using ``(x1, y1)`` and ``(x2,
// y2)`` as the Bézier control points.
func (cp *contextPage) CurveTo(x1 fl, y1 fl, x2 fl, y2 fl, x3 fl, y3 fl) {
	panic("not implemented") // TODO: Implement
}

// DrawText draws the given text using the current fill color.
func (cp *contextPage) DrawText(_ backend.TextDrawing) {
	panic("not implemented") // TODO: Implement
}

// AddFont register a new font to be used in the output and return
// an object used to store associated metadata.
// This method will be called several times with the same `face` argument,
// so caching is advised.
func (cp *contextPage) AddFont(face fonts.Face, content []byte) *backend.Font {
	panic("not implemented") // TODO: Implement
}

// DrawRasterImage draws the given image at the current point
func (cp *contextPage) DrawRasterImage(img image.Image, imageRendering string, width fl, height fl) {
	panic("not implemented") // TODO: Implement
}

// DrawGradient draws the given gradient at the current point.
// Solid gradient are already handled, meaning that only linear and radial
// must be taken care of.
func (cp *contextPage) DrawGradient(gradient backend.GradientLayout, width fl, height fl) {
	panic("not implemented") // TODO: Implement
}
