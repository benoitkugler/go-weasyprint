package pdf

import (
	"io"

	"github.com/benoitkugler/go-weasyprint/backend"
	"github.com/benoitkugler/go-weasyprint/matrix"
	"github.com/benoitkugler/go-weasyprint/style/parser"
	"github.com/benoitkugler/go-weasyprint/utils"
	cs "github.com/benoitkugler/pdf/contentstream"
	"github.com/benoitkugler/pdf/model"
	"github.com/benoitkugler/textlayout/fonts"
)

type fl = utils.Fl

var (
	_ backend.OutputGraphic = (*group)(nil)
	_ backend.OutputPage    = (*contextPage)(nil)
)

// group implements backend.OutputGraphic and
// is represented by a XObjectForm in PDF
type group struct {
	app           cs.Appearance
	pageRectangle [4]fl // left, top, right, bottom
}

func newGroup(left, top, right, bottom fl) group {
	return group{
		pageRectangle: [4]fl{left, top, right, bottom},
		app:           cs.NewAppearance(right-left, top-bottom),
	}
}

// contextPage implements backend.OutputPage
type contextPage struct {
	page model.PageObject

	embeddedFiles map[string]*model.FileSpec
	group
}

func newContextPage(left, top, right, bottom fl,
	embeddedFiles map[string]*model.FileSpec) *contextPage {
	out := &contextPage{
		embeddedFiles: embeddedFiles,
		group:         newGroup(left, top, right, bottom),
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

// Adjust the media boxes
func (cp *contextPage) SetTrimBox(left fl, top fl, right fl, bottom fl) {
	cp.page.TrimBox = &model.Rectangle{left, top, right, bottom}
}

func (cp *contextPage) SetBleedBox(left fl, top fl, right fl, bottom fl) {
	cp.page.BleedBox = &model.Rectangle{left, top, right, bottom}
}

// graphic operations

// Returns the current page rectangle
func (g *group) GetPageRectangle() (left fl, top fl, right fl, bottom fl) {
	return g.pageRectangle[0], g.pageRectangle[1], g.pageRectangle[2], g.pageRectangle[3]
}

// OnNewStack save the current graphic stack,
// execute the given closure, and restore the stack.
// If an error is encoutered, the stack is still restored
// and the error is returned
func (g *group) OnNewStack(task func() error) error {
	g.app.SaveState()
	err := task()
	_ = g.app.RestoreState() // the calls are balanced
	return err
}

// AddGroup creates a new drawing target with the given
// bounding box.
// If the backend does not support groups, the current target should be returned.
func (group) AddGroup(x fl, y fl, width fl, height fl) backend.OutputGraphic {
	out := newGroup(x, y, x+width, y+height)
	return &out
}

// DrawGroup add the `gr` to the current target. It will panic
// if `gr` was not created with `AddGroup`
func (g *group) DrawGroup(gr backend.OutputGraphic) {
	form := gr.(*group).app.ToXFormObject(true)
	g.app.AddXObject(form)
}

func (g *group) AddPattern(width fl, height fl, repeatWidth fl, repeatHeight fl, mt matrix.Transform) backend.Pattern {
	panic("not implemented") // TODO: Implement
}

func (g *group) SetColorPattern(pattern backend.Pattern, stroke bool) {
	panic("not implemented") // TODO: Implement
}

// Adds a rectangle of the given size to the current path,
// at position ``(x, y)`` in user-space coordinates.
// (X,Y) coordinates are the top left corner of the rectangle.
func (g *group) Rectangle(x fl, y fl, width fl, height fl) {
	g.app.Ops(cs.OpRectangle{X: x, Y: y, W: width, H: height})
}

func (g *group) Clip(evenOdd bool) {
	if evenOdd {
		g.app.Ops(cs.OpEOClip{})
	} else {
		g.app.Ops(cs.OpClip{})
	}
}

func (g *group) SetColorRgba(color parser.RGBA, stroke bool) {
	alpha := color.A
	color.A = 1 // do not take into account the opacity, it is handled by `SetAlpha`
	if stroke {
		g.app.SetColorStroke(color)
	} else {
		g.app.SetColorFill(color)
	}
	g.SetAlpha(alpha, stroke)
}

// Set current alpha
// `stroke` controls whether stroking or filling operations are concerned.
func (g *group) SetAlpha(alpha fl, stroke bool) {
	if stroke {
		g.app.SetStrokeAlpha(alpha)
	} else {
		g.app.SetFillAlpha(alpha)
	}
}

func (g *group) SetLineWidth(width fl) {
	g.app.Ops(cs.OpSetLineWidth{W: width})
}

func (g *group) SetDash(dashes []fl, offset fl) {
	g.app.Ops(cs.OpSetDash{Dash: model.DashPattern{Array: dashes, Phase: offset}})
}

// A drawing operator that fills the current path
// according to the current fill rule,
// (each sub-path is implicitly closed before being filled).
// After `fill`, the current path will is cleared
func (g *group) Fill(evenOdd bool) {
	if evenOdd {
		g.app.Ops(cs.OpEOFill{})
	} else {
		g.app.Ops(cs.OpFill{})
	}
}

// A drawing operator that strokes the current path
// according to the current line width, line join, line cap,
// and dash settings.
// After `Stroke`, the current path will be cleared.
func (g *group) Stroke() {
	g.app.Ops(cs.OpStroke{})
}

// Modifies the current transformation matrix (CTM)
// by applying `mt` as an additional transformation.
// The new transformation of user space takes place
// after any existing transformation.
func (g *group) Transform(mt matrix.Transform) {
	a, b, c, d, e, f := mt.Data()
	g.app.Transform(model.Matrix{a, b, c, d, e, f})
}

// GetTransform returns the current transformation matrix.
func (g *group) GetTransform() matrix.Transform {
	m := g.app.State.Matrix
	return matrix.New(m[0], m[1], m[2], m[3], m[4], m[5])
}

// Begin a new sub-path.
// After this call the current point will be ``(x, y)``.
func (g *group) MoveTo(x fl, y fl) {
	g.app.Ops(cs.OpMoveTo{X: x, Y: y})
}

// Adds a line to the path from the current point
// to position ``(x, y)`` in user-space coordinates.
// After this call the current point will be ``(x, y)``.
// A current point must be defined before using this method.
func (g *group) LineTo(x fl, y fl) {
	g.app.Ops(cs.OpLineTo{X: x, Y: y})
}

// Add cubic Bézier curve to current path.
// The curve shall extend to ``(x3, y3)`` using ``(x1, y1)`` and ``(x2,
// y2)`` as the Bézier control points.
func (g *group) CurveTo(x1 fl, y1 fl, x2 fl, y2 fl, x3 fl, y3 fl) {
	g.app.Ops(cs.OpCubicTo{X1: x1, Y1: y1, X2: x2, Y2: y2, X3: x3, Y3: y3})
}

// DrawText draws the given text using the current fill color.
func (g *group) DrawText(_ backend.TextDrawing) {
	panic("not implemented") // TODO: Implement
}

// AddFont register a new font to be used in the output and return
// an object used to store associated metadata.
// This method will be called several times with the same `face` argument,
// so caching is advised.
func (g *group) AddFont(face fonts.Face, content []byte) *backend.Font {
	panic("not implemented") // TODO: Implement
}

// DrawRasterImage draws the given image at the current point
func (g *group) DrawRasterImage(imgContent io.ReadCloser, imgFormat, imageRendering string, width fl, height fl) {
	mimeType := "image/" + imgFormat
	cs.ParseImage(imgContent)
}

// DrawGradient draws the given gradient at the current point.
// Solid gradient are already handled, meaning that only linear and radial
// must be taken care of.
func (g *group) DrawGradient(layout backend.GradientLayout, width fl, height fl) {
	alphas := make([]fl, len(layout.Colors))
	var needOpacity bool
	for i, c := range layout.Colors {
		alphas[i] = c.A
		if c.A != 1 {
			needOpacity = true
		}
	}

	alphaCouples := make([][2]fl, len(alphas)-1)
	colorCouples := make([][2][3]fl, len(alphas)-1)
	exponents := make([]int, len(alphas)-1)
	for i := range alphaCouples {
		alphaCouples[i] = [2]fl{alphas[i], alphas[i+1]}
		colorCouples[i] = [2][3]fl{
			{layout.Colors[i].R, layout.Colors[i].G, layout.Colors[i].B},
			{layout.Colors[i+1].R, layout.Colors[i+1].G, layout.Colors[i+1].B},
		}
		exponents[i] = 1
	}

	// Premultiply colors
	for i, alpha := range alphas {
		if alpha == 0 {
			if i > 0 {
				colorCouples[i-1][1] = colorCouples[i-1][0]
			}
			if i < len(layout.Colors)-1 {
				colorCouples[i][0] = colorCouples[i][1]
			}
		}
	}
	for i, v := range alphaCouples {
		a0, a1 := v[0], v[1]
		if a0 != 0 && a1 != 0 && v != ([2]fl{1, 1}) {
			exponents[i] = int(a0 / a1)
		}
	}

	var functions, alphaFunctions []model.FunctionDict
	for i, v := range colorCouples {
		c0, c1 := v[0], v[1]
		n := exponents[i]
		fn := model.FunctionDict{
			Domain: []model.Range{{0, 1}},
			FunctionType: model.FunctionExpInterpolation{
				C0: c0[:],
				C1: c1[:],
				N:  n,
			},
		}
		functions = append(functions, fn)

		alphaFn := fn
		a0, a1 := alphaCouples[i][0], alphaCouples[i][1]
		alphaFn.FunctionType = model.FunctionExpInterpolation{
			C0: []model.Fl{a0},
			C1: []model.Fl{a1},
			N:  1,
		}
		alphaFunctions = append(alphaFunctions, alphaFn)
	}
	bg := model.BaseGradient{
		Domain: [2]fl{layout.Positions[0], layout.Positions[len(layout.Positions)-1]},
		Function: []model.FunctionDict{{
			Domain: []model.Range{{layout.Positions[0], layout.Positions[len(layout.Positions)-1]}},
			FunctionType: model.FunctionStitching{
				Functions: functions,
				Bounds:    layout.Positions[1 : len(layout.Positions)-1],
				Encode:    model.FunctionEncodeRepeat(len(layout.Colors) - 1),
			},
		}},
	}
	if !layout.Reapeating {
		bg.Extend = [2]bool{true, true}
	}

	// alpha stream is similar
	alphaBg := bg
	alphaBg.Function = alphaFunctions

	var type_, alphaType model.Shading
	if layout.Kind == "linear" {
		type_ = model.ShadingAxial{
			BaseGradient: bg,
			Coords:       [4]fl{layout.Data[0], layout.Data[1], layout.Data[2], layout.Data[3]},
		}
		alphaType = model.ShadingAxial{
			BaseGradient: alphaBg,
			Coords:       [4]fl{layout.Data[0], layout.Data[1], layout.Data[2], layout.Data[3]},
		}
	} else {
		type_ = model.ShadingRadial{
			BaseGradient: bg,
			Coords:       layout.Data,
		}
		alphaType = model.ShadingRadial{
			BaseGradient: alphaBg,
			Coords:       layout.Data,
		}
	}

	sh := model.ShadingDict{
		ColorSpace:  model.ColorSpaceRGB,
		ShadingType: type_,
	}
	alphaSh := model.ShadingDict{
		ColorSpace:  model.ColorSpaceGray,
		ShadingType: alphaType,
	}

	g.Transform(matrix.New(1, 0, 0, layout.ScaleY, 0, 0))

	if needOpacity {
		alphaStream := cs.NewAppearance(width, height)

		alphaStream.Transform(model.Matrix{1, 0, 0, layout.ScaleY, 0, 0})
		shName := alphaStream.AddShading(&alphaSh)
		alphaStream.Ops(cs.OpShFill{Shading: shName})
		transparency := alphaStream.ToXFormObject(false)

		alphaState := model.GraphicState{
			SMask: model.SoftMaskDict{
				S: model.ObjName("Luminosity"),
				G: &model.XObjectTransparencyGroup{XObjectForm: *transparency},
			},
			Ca:  1,
			AIS: false,
		}

		g.app.SetGraphicState(&alphaState)
	}

	g.app.Shading(&sh)
}
