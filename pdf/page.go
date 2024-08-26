package pdf

import (
	"log"
	"strings"

	cs "github.com/benoitkugler/pdf/contentstream"
	"github.com/benoitkugler/pdf/model"
	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/css/parser"
	"github.com/benoitkugler/webrender/matrix"
	"github.com/benoitkugler/webrender/utils"
)

type fl = utils.Fl

var (
	_ backend.Canvas       = (*group)(nil)
	_ backend.Page         = (*outputPage)(nil)
	_ backend.GraphicState = (*group)(nil)
)

func (g *group) SetAlphaMask(mask backend.Canvas) {
	alphaStream := mask.(*group).app
	g.drawMask(&alphaStream)
}

func (g *group) SetColorPattern(p backend.Canvas, contentWidth, contentHeight fl, mt matrix.Transform, stroke bool) {
	mat := model.Matrix{mt.A, mt.B, mt.C, mt.D, mt.E, mt.F}
	mat = mat.Multiply(g.app.State.Matrix)

	// initiate the pattern ...
	_, _, gridWidth, gridHeight := p.GetRectangle()
	pattern := &model.PatternTiling{
		XStep: gridWidth, YStep: gridHeight,
		Matrix:     mat,
		PaintType:  1,
		TilingType: 1,
	}

	contentXObject := p.(*group).app.ToXFormObject(compressStreams)
	// wrap the content into a Do command
	patternApp := cs.NewGraphicStream(model.Rectangle{Llx: 0, Lly: 0, Urx: contentWidth, Ury: contentHeight})
	patternApp.AddXObject(contentXObject)
	patternApp.ApplyToTilling(pattern)

	// select the color space
	patternName := g.app.AddPattern(pattern)
	if stroke {
		g.app.Ops(cs.OpSetStrokeColorSpace{ColorSpace: model.ColorSpacePattern})
		g.app.Ops(cs.OpSetStrokeColorN{Pattern: patternName})
	} else {
		g.app.Ops(cs.OpSetFillColorSpace{ColorSpace: model.ColorSpacePattern})
		g.app.Ops(cs.OpSetFillColorN{Pattern: patternName})
	}
}

func (g *group) SetBlendingMode(mode string) {
	// PDF blend modes have TitleCase
	chunks := strings.Split(mode, "-")
	for i, s := range chunks {
		chunks[i] = strings.Title(s)
	}
	bm := strings.Join(chunks, "")
	g.app.SetGraphicState(&model.GraphicState{BM: []model.Name{model.ObjName(bm)}})
}

func (g *group) Clip(evenOdd bool) {
	if evenOdd {
		g.app.Ops(cs.OpEOClip{}, cs.OpEndPath{})
	} else {
		g.app.Ops(cs.OpClip{}, cs.OpEndPath{})
	}
}

func (g *group) SetColorRgba(color parser.RGBA, stroke bool) {
	alpha := color.A
	color.A = 1 // do not take into account the opacity, it is handled by `setXXXAlpha`
	if stroke {
		g.app.SetColorStroke(color)
		g.app.SetStrokeAlpha(alpha)
	} else {
		g.app.SetColorFill(color)
		g.app.SetFillAlpha(alpha)
	}
}

func (g *group) SetLineWidth(width fl) {
	g.app.Ops(cs.OpSetLineWidth{W: width})
}

func (g *group) SetDash(dashes []fl, offset fl) {
	g.app.Ops(cs.OpSetDash{Dash: model.DashPattern{Array: dashes, Phase: offset}})
}

func (g *group) SetStrokeOptions(opts backend.StrokeOptions) {
	g.app.Ops(
		cs.OpSetLineCap{Style: uint8(opts.LineCap)},
		cs.OpSetLineJoin{Style: uint8(opts.LineJoin)},
		cs.OpSetMiterLimit{Limit: opts.MiterLimit},
	)
}

func (g *group) GetTransform() matrix.Transform {
	m := g.app.State.Matrix
	return matrix.New(m[0], m[1], m[2], m[3], m[4], m[5])
}

// Modifies the current transformation matrix (CTM)
// by applying `mt` as an additional transformation.
// The new transformation of user space takes place
// after any existing transformation.
func (g *group) Transform(mt matrix.Transform) {
	g.app.Transform(model.Matrix{mt.A, mt.B, mt.C, mt.D, mt.E, mt.F})
}

// group implements backend.Canvas and
// is represented by a XObjectForm in PDF
type group struct {
	cache

	app cs.GraphicStream
}

func newGroup(cache cache,
	left, top, right, bottom fl,
) group {
	return group{
		cache: cache,
		app:   cs.NewGraphicStream(model.Rectangle{Llx: left, Lly: top, Urx: right, Ury: bottom}), // y grows downward
	}
}

func (g *group) State() backend.GraphicState { return g }

// outputPage implements backend.Page
type outputPage struct {
	page model.PageObject // the content stream is written in `group`

	customMediaBox *model.Rectangle // overing bbox

	embeddedFiles map[string]*model.FileSpec
	group
}

func newContextPage(left, top, right, bottom fl,
	embeddedFiles map[string]*model.FileSpec,
	cache cache,
) *outputPage {
	out := &outputPage{
		embeddedFiles: embeddedFiles,
		group:         newGroup(cache, left, top, right, bottom),
	}
	return out
}

// update the underlying PageObject with the content stream
func (cp *outputPage) finalize() {
	// the MediaBox is the unsclaled BBox. TODO: why ?
	cp.app.ApplyToPageObject(&cp.page, compressStreams)
	if cp.customMediaBox != nil {
		cp.page.MediaBox = cp.customMediaBox
	}
}

func (cp *outputPage) AddInternalLink(xMin, yMin, xMax, yMax fl, anchorName string) {
	an := model.AnnotationDict{
		BaseAnnotation: model.BaseAnnotation{
			Rect: model.Rectangle{Llx: xMin, Lly: yMin, Urx: xMax, Ury: yMax},
		},
		Subtype: model.AnnotationLink{
			BS:   &model.BorderStyle{W: model.ObjFloat(0)},
			Dest: model.DestinationString(anchorName),
		},
	}
	cp.page.Annots = append(cp.page.Annots, &an)
}

func (cp *outputPage) AddExternalLink(xMin, yMin, xMax, yMax fl, url string) {
	an := model.AnnotationDict{
		BaseAnnotation: model.BaseAnnotation{
			Rect: model.Rectangle{Llx: xMin, Lly: yMin, Urx: xMax, Ury: yMax},
		},
		Subtype: model.AnnotationLink{
			BS: &model.BorderStyle{W: model.ObjFloat(0)},
			A:  model.Action{ActionType: model.ActionURI{URI: url}},
		},
	}
	cp.page.Annots = append(cp.page.Annots, &an)
}

// Add file annotation on the current page
func (cp *outputPage) AddFileAnnotation(xMin, yMin, xMax, yMax fl, fileID string) {
	rect := model.Rectangle{Llx: xMin, Lly: yMin, Urx: xMax, Ury: yMax}
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
func (cp *outputPage) SetMediaBox(left fl, top fl, right fl, bottom fl) {
	cp.customMediaBox = &model.Rectangle{Llx: left, Lly: top, Urx: right, Ury: bottom}
}

func (cp *outputPage) SetTrimBox(left fl, top fl, right fl, bottom fl) {
	cp.page.TrimBox = &model.Rectangle{Llx: left, Lly: top, Urx: right, Ury: bottom}
}

func (cp *outputPage) SetBleedBox(left fl, top fl, right fl, bottom fl) {
	cp.page.BleedBox = &model.Rectangle{Llx: left, Lly: top, Urx: right, Ury: bottom}
}

// graphic operations

// Returns the current page rectangle
func (g *group) GetRectangle() (left fl, top fl, right fl, bottom fl) {
	bbox := g.app.BoundingBox
	return bbox.Llx, bbox.Lly, bbox.Urx, bbox.Ury
}

// OnNewStack save the current graphic stack,
// execute the given closure, and restore the stack.
// If an error is encoutered, the stack is still restored
// and the error is returned
func (g *group) OnNewStack(task func()) {
	g.app.SaveState()
	task()
	_ = g.app.RestoreState() // the calls are balanced
}

// NewGroup creates a new drawing target with the given
// bounding box.
func (g *group) NewGroup(x fl, y fl, width fl, height fl) backend.Canvas {
	out := newGroup(g.cache, x, y, x+width, y+height)
	return &out
}

// DrawGroup add the `gr` content to the current target. It will panic
// if `gr` was not created with `AddGroup`
func (g *group) DrawWithOpacity(opacity fl, gr backend.Canvas) {
	content := gr.(*group).app.ToXFormObject(compressStreams)
	form := &model.XObjectTransparencyGroup{
		XObjectForm: *content,
		CS:          model.ColorSpaceRGB,
		I:           true,
	}
	g.app.SetFillAlpha(opacity)
	g.app.SetStrokeAlpha(opacity)
	g.app.AddXObject(form)
}

func (g *group) drawMask(app *cs.GraphicStream) {
	transparency := app.ToXFormObject(compressStreams)
	g.app.SetAlphaMask(transparency)
}

// Adds a rectangle of the given size to the current path,
// at position “(x, y)“ in user-space coordinates.
// (X,Y) coordinates are the top left corner of the rectangle.
func (g *group) Rectangle(x fl, y fl, width fl, height fl) {
	g.app.Ops(cs.OpRectangle{X: x, Y: y, W: width, H: height})
}

// A drawing operator that fills the current path
// according to the current fill rule,
// (each sub-path is implicitly closed before being filled).
// After `fill`, the current path will is cleared
func (g *group) Paint(op backend.PaintOp) {
	fill := op&(backend.FillEvenOdd|backend.FillNonZero) != 0
	stroke := op&backend.Stroke != 0
	evenOdd := op&backend.FillEvenOdd != 0
	if fill && stroke {
		if evenOdd {
			g.app.Ops(cs.OpEOFillStroke{})
		} else {
			g.app.Ops(cs.OpFillStroke{})
		}
	} else if fill {
		if evenOdd {
			g.app.Ops(cs.OpEOFill{})
		} else {
			g.app.Ops(cs.OpFill{})
		}
	} else if stroke {
		g.app.Ops(cs.OpStroke{})
	} else {
		g.app.Ops(cs.OpEndPath{})
	}
}

// Begin a new sub-path.
// After this call the current point will be “(x, y)“.
func (g *group) MoveTo(x fl, y fl) {
	g.app.Ops(cs.OpMoveTo{X: x, Y: y})
}

// Adds a line to the path from the current point
// to position “(x, y)“ in user-space coordinates.
// After this call the current point will be “(x, y)“.
// A current point must be defined before using this method.
func (g *group) LineTo(x fl, y fl) {
	g.app.Ops(cs.OpLineTo{X: x, Y: y})
}

// Add cubic Bézier curve to current path.
// The curve shall extend to “(x3, y3)“ using “(x1, y1)“ and “(x2,
// y2)“ as the Bézier control points.
func (g *group) CubicTo(x1, y1, x2, y2, x3, y3 fl) {
	g.app.Ops(cs.OpCubicTo{X1: x1, Y1: y1, X2: x2, Y2: y2, X3: x3, Y3: y3})
}

// ClosePath close the current path, which will apply line join style.
func (g *group) ClosePath() { g.app.Ops(cs.OpClosePath{}) }

// DrawRasterImage draws the given image at the current point
func (g *group) DrawRasterImage(img backend.RasterImage, width fl, height fl) {
	// check the global cache
	obj, has := g.images[img.ID]
	if !has {
		var err error
		obj, _, err = cs.ParseImage(img.Content, img.MimeType)
		if err != nil {
			log.Printf("failed to process image: %s", err)
			return
		}
		obj.Interpolate = img.Rendering == "auto"
		g.images[img.ID] = obj
	}

	g.app.AddXObjectDims(obj, 0, height, width, -height)
}

// DrawGradient draws the given gradient at the current point.
// Solid gradient are already handled, meaning that only linear and radial
// must be taken care of.
func (g *group) DrawGradient(layout backend.GradientLayout, width fl, height fl) {
	grad := cs.GradientComplex{
		Offsets:    layout.Positions,
		Colors:     make([][4]fl, len(layout.Colors)),
		Reapeating: layout.Reapeating,
	}
	for i, c := range layout.Colors {
		grad.Colors[i] = [4]fl{c.R, c.G, c.B, c.A}
	}

	if layout.Kind == "linear" {
		grad.Direction = cs.GradientLinear{layout.Coords[0], layout.Coords[1], layout.Coords[2], layout.Coords[3]}
	} else {
		grad.Direction = cs.GradientRadial(layout.Coords)
	}

	sh, alphaSh := grad.BuildShadings()

	g.Transform(matrix.New(1, 0, 0, layout.ScaleY, 0, 0))

	if alphaSh != nil {
		alphaStream := cs.NewGraphicStream(model.Rectangle{Llx: 0, Lly: 0, Urx: width, Ury: height})

		alphaStream.Transform(model.Matrix{1, 0, 0, layout.ScaleY, 0, 0})
		shName := alphaStream.AddShading(alphaSh)
		alphaStream.Ops(cs.OpShFill{Shading: shName})

		g.drawMask(&alphaStream)
	}

	g.app.Shading(sh)
}
