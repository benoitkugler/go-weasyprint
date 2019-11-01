package goweasyprint

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"text/template"

	"github.com/benoitkugler/go-weasyprint/backend"
	"github.com/benoitkugler/go-weasyprint/images"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/utils"

	bo "github.com/benoitkugler/go-weasyprint/boxes"
)

// Take an "after layout" box tree and draw it onto a cairo context.

type Drawer = backend.Drawer

var SIDES = [4]string{"top", "right", "bottom", "left"}

const (
	pi = float64(math.Pi)

	headerSVG = `
	<svg height="{{ .Height }}" width="{{ .Width }}"
		 fill="transparent" stroke="black" stroke-width="1"
		 xmlns="http://www.w3.org/2000/svg"
		 xmlns:xlink="http://www.w3.org/1999/xlink">
  	`

	crop = `
  <!-- horizontal top left -->
  <path d="M0,{{ .Bleed.Top }} h{{ .HalfBleed.Left }}" />
  <!-- horizontal top right -->
  <path d="M0,{{ .Bleed.Top }} h{{ .HalfBleed.Right }}"
        transform="translate({{ .Width }},0) scale(-1,1)" />
  <!-- horizontal bottom right -->
  <path d="M0,{{ .Bleed.Bottom }} h{{ .HalfBleed.Right }}"
        transform="translate({{ .Width }},{{ .Height }}) scale(-1,-1)" />
  <!-- horizontal bottom left -->
  <path d="M0,{{ .Bleed.Bottom }} h{{ .HalfBleed.Left }}"
        transform="translate(0,{{ .Height }}) scale(1,-1)" />
  <!-- vertical top left -->
  <path d="M{{ .Bleed.Left }},0 v{{ .HalfBleed.Top }}" />
  <!-- vertical bottom right -->
  <path d="M{{ .Bleed.Right }},0 v{{ .HalfBleed.Bottom }}"
        transform="translate({{ .Width }},{{ .Height }}) scale(-1,-1)" />
  <!-- vertical bottom left -->
  <path d="M{{ .Bleed.Left }},0 v{{ .HalfBleed.Bottom }}"
        transform="translate(0,{{ .Height }}) scale(1,-1)" />
  <!-- vertical top right -->
  <path d="M{{ .Bleed.Right }},0 v{{ .HalfBleed.Top }}"
        transform="translate({{ .Width }},0) scale(-1,1)" />
`
	cross = `
  <!-- top -->
  <circle r="{{ .HalfBleed.Top }}"
          transform="scale(0.5)
                     translate({{ .Width }},{{ .HalfBleed.Top }}) scale(0.5)" />
  <path d="M-{{ .HalfBleed.Top }},{{ .HalfBleed.Top }} h{{ .Bleed.Top }}
           M0,0 v{{ .Bleed.Top }}"
        transform="scale(0.5) translate({{ .Width }},0)" />
  <!-- bottom -->
  <circle r="{{ .HalfBleed.Bottom }}"
          transform="translate(0,{{ .Height }}) scale(0.5)
                     translate({{ .Width }},-{{ .HalfBleed.Bottom }}) scale(0.5)" />
  <path d="M-{{ .HalfBleed.Bottom }},-{{ .HalfBleed.Bottom }} h{{ .Bleed.Bottom }}
           M0,0 v-{{ .Bleed.Bottom }}"
        transform="translate(0,{{ .Height }}) scale(0.5) translate({{ .Width }},0)" />
  <!-- left -->
  <circle r="{{ .HalfBleed.Left }}"
          transform="scale(0.5)
                     translate({{ .HalfBleed.Left }},{{ .Height }}) scale(0.5)" />
  <path d="M{{ .HalfBleed.Left }},-{{ .HalfBleed.Left }} v{{ .Bleed.Left }}
           M0,0 h{{ .Bleed.Left }}"
        transform="scale(0.5) translate(0,{{ .Height }})" />
  <!-- right -->
  <circle r="{{ .HalfBleed.Right }}"
          transform="translate({{ .Width }},0) scale(0.5)
                     translate(-{{ .HalfBleed.Right }},{{ .Height }}) scale(0.5)" />
  <path d="M-{{ .HalfBleed.Right }},-{{ .HalfBleed.Right }} v{{ .Bleed.Right }}
           M0,0 h-{{ .Bleed.Right }}"
        transform="translate({{ .Width }},0)
                   scale(0.5) translate(0,{{ .Height }})" />
`
)

type bleedData struct {
	Top, Bottom, Left, Right pr.Float
}

// FIXME: check gofpdf support for SVG
type svgArgs struct {
	Width, Height    pr.Float
	Bleed, HalfBleed bleedData
}

// The context manager
// 	with stacked(context):
//		<body>
// is equivalent to
//		context.Save()
//		try:
//			<body>
//  	finally:
//			context.restore()
// FIXME: add context manager

// Transform a HSV color to a RGB color.
func hsv2rgb(hue, saturation, value float64) (r, g, b float64) {
	c := value * saturation
	x := c * (1 - math.Abs(utils.FloatModulo(hue/60, 2)-1))
	m := value - c
	switch {
	case 0 <= hue && hue < 60:
		return c + m, x + m, m
	case 60 <= hue && hue < 120:
		return x + m, c + m, m
	case 120 <= hue && hue < 180:
		return m, c + m, x + m
	case 180 <= hue && hue < 240:
		return m, x + m, c + m
	case 240 <= hue && hue < 300:
		return x + m, m, c + m
	case 300 <= hue && hue < 360:
		return c + m, m, x + m
	default:
		log.Fatalf("invalid hue %f", hue)
		return 0, 0, 0
	}
}

// Transform a RGB color to a HSV color.
func rgb2hsv(red, green, blue float64) (h, s, c float64) {
	cmax := utils.Maxs(red, green, blue)
	cmin := utils.Mins(red, green, blue)
	delta := cmax - cmin
	var hue float64
	if delta == 0 {
		hue = 0
	} else if cmax == red {
		hue = 60 * utils.FloatModulo((green-blue)/delta, 6)
	} else if cmax == green {
		hue = 60 * ((blue-red)/delta + 2)
	} else if cmax == blue {
		hue = 60 * ((red-green)/delta + 4)
	}
	var saturation float64
	if delta != 0 {
		saturation = delta / cmax
	}
	return hue, saturation, cmax
}

// Return a darker color.
func darken(color Color) Color {
	hue, saturation, value := rgb2hsv(color.R, color.G, color.B)
	value /= 1.5
	saturation /= 1.25
	r, g, b := hsv2rgb(hue, saturation, value)
	return Color{R: r, G: g, B: b, A: color.A}
}

// Return a lighter color.
func lighten(color Color) Color {
	hue, saturation, value := rgb2hsv(color.R, color.G, color.B)
	value = 1 - (1-value)/1.5
	if saturation != 0 {
		saturation = 1 - (1-saturation)/1.25
	}
	r, g, b := hsv2rgb(hue, saturation, value)
	return Color{R: r, G: g, B: b, A: color.A}
}

// Draw the given PageBox.
func drawPage(page *bo.PageBox, context Drawer, enableHinting bool) {
	bleed := bleedData{
		Top:    page.Style.GetBleedTop().Value,
		Bottom: page.Style.GetBleedBottom().Value,
		Left:   page.Style.GetBleedLeft().Value,
		Right:  page.Style.GetBleedRight().Value,
	}
	marks := page.Style.GetMarks()
	stackingContext := NewStackingContextFromPage(page)
	drawBackground(context, stackingContext.box.Box().Background, enableHinting,
		false, bleed, marks)
	drawBackground(context, page.CanvasBackground, enableHinting, false, bleedData{}, pr.Marks{})
	drawBorder(context, page, enableHinting)
	drawStackingContext(context, stackingContext, enableHinting)
}

func drawBoxBackgroundAndBorder(context Drawer, page *bo.PageBox, box Box, enableHinting bool) {
	drawBackground(context, box.Box().Background, enableHinting, true, bleedData{}, pr.Marks{})
	if box_, ok := box.(bo.InstanceTableBox); ok {
		box := box_.Table()
		drawTableBackgrounds(context, page, box_, enableHinting)
		if box.Style.GetBorderCollapse() == "separate" {
			drawBorder(context, box, enableHinting)
			for _, rowGroup := range box.Children {
				for _, row := range rowGroup.Box().Children {
					for _, cell := range row.Box().Children {
						if cell.Box().Style.GetEmptyCells() == "show" || !cell.Box().Empty {
							drawBorder(context, cell, enableHinting)
						}
					}
				}
			}
		} else {
			drawCollapsedBorders(context, box, enableHinting)
		}
	} else {
		drawBorder(context, box, enableHinting)
	}
}

// Draw a ``stackingContext`` on ``context``.
func drawStackingContext(context Drawer, stackingContext StackingContext, enableHinting bool) {
	// See http://www.w3.org/TR/CSS2/zindex.html
	// with stacked(context) {
	box_ := stackingContext.box
	box := box_.Box()
	if clips := box.Style.GetClip(); box.IsAbsolutelyPositioned() && len(clips) != 0 {
		top, right, bottom, left := clips[0], clips[1], clips[2], clips[3]
		if top.String == "auto" {
			top.Value = 0
		}
		if right.String == "auto" {
			right.Value = 0
		}
		if bottom.String == "auto" {
			bottom.Value = box.BorderHeight()
		}
		if left.String == "auto" {
			left.Value = box.BorderWidth()
		}
		context.Rectangle(
			float64(box.BorderBoxX()+right.Value),
			float64(box.BorderBoxY()+top.Value),
			float64(left.Value-right.Value),
			float64(bottom.Value-top.Value),
		)
		context.Clip()
	}

	if box.Style.GetOpacity() < 1 {
		context.PushGroup()
	}

	if box.TransformationMatrix != nil {
		if err := box.TransformationMatrix.Copy().Invert(); err != nil { // except cairo.CairoError
			return
		}
		context.Transform(box.TransformationMatrix)
	}

	// Point 1 is done in drawPage

	// Point 2
	if bo.TypeBlockBox.IsInstance(box_) || bo.IsMarginBox(box_) ||
		bo.TypeInlineBlockBox.IsInstance(box_) || bo.TypeTableCellBox.IsInstance(box_) ||
		bo.IsFlexContainerBox(box_) {

		// The canvas background was removed by setCanvasBackground
		drawBoxBackgroundAndBorder(context, stackingContext.page, box_, enableHinting)

		// with stacked(context) {
		if box.Style.GetOverflow() != "visible" {
			// Only clip the content && the children:
			// - the background is already clipped
			// - the border must *not* be clipped
			roundedBoxPath(context, box.RoundedPaddingBox())
			context.Clip()
		}

		// Point 3
		for _, childContext := range stackingContext.negativeZContexts {
			drawStackingContext(context, childContext, enableHinting)
		}

		// Point 4
		for _, block := range stackingContext.blockLevelBoxes {
			drawBoxBackgroundAndBorder(context, stackingContext.page, block, enableHinting)
		}

		// Point 5
		for _, childContext := range stackingContext.floatContexts {
			drawStackingContext(context, childContext, enableHinting)
		}

		// Point 6
		if bo.TypeInlineBox.IsInstance(box_) {
			drawInlineLevel(context, stackingContext.page, box_, enableHinting, 0, "clip")
		}

		// Point 7
		for _, block := range append([]Box{box_}, stackingContext.blocksAndCells...) {
			if block, ok := block.(bo.InstanceReplacedBox); ok {
				drawReplacedbox(context, block)
			} else {
				for _, child := range block.Box().Children {
					if bo.TypeLineBox.IsInstance(child) {
						drawInlineLevel(context, stackingContext.page, child, enableHinting, 0, "clip")
					}
				}
			}
		}

		// Point 8
		for _, childContext := range stackingContext.zeroZContexts {
			drawStackingContext(context, childContext, enableHinting)
		}

		// Point 9
		for _, childContext := range stackingContext.positiveZContexts {
			drawStackingContext(context, childContext, enableHinting)
		}

		// Point 10
		drawOutlines(context, box_, enableHinting)

		if op := float64(box.Style.GetOpacity()); op < 1 {
			context.PopGroupToSource()
			context.PaintWithAlpha(op)
		}
	}
}

// Draw the path of the border radius box.
//     ``widths`` is a tuple of the inner widths (top, right, bottom, left) from
//     the border box. Radii are adjusted from these values. Default is (0, 0, 0,
//     0).
//     Inspired by cairo cookbook
//     http://cairographics.org/cookbook/roundedrectangles/
//
func roundedBoxPath(context Drawer, radii bo.RoundedBox) {
	x, y, w, h, tl, tr, br, bl := float64(radii.X), float64(radii.Y), radii.Width, radii.Height, radii.TopLeft, radii.TopRight, radii.BottomRight, radii.BottomLeft

	if tl[0] == 0 || tl[1] == 0 {
		tl = bo.Point{0, 0}
	}
	if tr[0] == 0 || tr[1] == 0 {
		tr = bo.Point{0, 0}
	}
	if br[0] == 0 || br[1] == 0 {
		br = bo.Point{0, 0}
	}
	if bl[0] == 0 || bl[1] == 0 {
		bl = bo.Point{0, 0}
	}

	if (tl == bo.Point{} && tr == bo.Point{} && br == bo.Point{} && bl == bo.Point{}) {
		// No radius, draw a rectangle
		context.Rectangle(float64(x), float64(y), float64(w), float64(h))
		return
	}

	context.MoveTo(float64(x), float64(y))
	context.NewSubPath()
	for i, v := range [4][2]bo.Point{
		{{0, 0}, tl}, {{w, 0}, tr}, {{w, h}, br}, {{0, h}, bl},
	} {
		w, h, rx, ry := float64(v[0][0]), float64(v[0][1]), float64(v[1][0]), float64(v[1][1])
		context.Save()
		context.Translate(x+w, y+h)
		radius := math.Max(rx, ry)
		if radius != 0 {
			context.Scale(math.Min(rx/ry, 1), math.Min(ry/rx, 1))
		}
		aw, ah := 1., 1.
		if w != 0 {
			aw = -1
		}
		if h != 0 {
			ah = -1
		}
		context.Arc(aw*radius, ah*radius, radius,
			(2+float64(i))*pi/2, (3+float64(i))*pi/2)
		context.Restore()
	}
}

func formatSVG(svg string, data svgArgs) (string, error) {
	tmp, err := template.New("svg").Parse(svg)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if err := tmp.Execute(&b, data); err != nil {
		return "", fmt.Errorf("unexpected template error : %s", err)
	}
	return b.String(), nil
}

func reversed(in []bo.BackgroundLayer) []bo.BackgroundLayer {
	N := len(in)
	out := make([]bo.BackgroundLayer, N)
	for i, v := range in {
		out[N-1-i] = v
	}
	return out
}

func drawBackground2(context Drawer, bg *bo.Background, enableHinting bool) error {
	return drawBackground(context, bg, enableHinting, true, bleedData{}, pr.Marks{})
}

// Draw the background color and image to a ``cairo.Context``
// If ``clipBox`` is set to ``false``, the background is not clipped to the
// border box of the background, but only to the painting area
// clipBox=true bleed=None marks=()
func drawBackground(context Drawer, bg *bo.Background, enableHinting, clipBox bool, bleed bleedData,
	marks pr.Marks) error {
	if bg == nil {
		return nil
	}

	// with stacked(context) {
	if enableHinting {
		// Prefer crisp edges on background rectangles.
		context.SetAntialias(cairo.ANTIALIASNONE)
	}

	if clipBox {
		for _, box := range bg.Layers[len(bg.Layers)-1].ClippedBoxes {
			roundedBoxPath(context, box)
		}
		context.Clip()
	}

	// Background color
	if bg.Color.A > 0 {
		// with stacked(context) {
		paintingArea := bg.Layers[len(bg.Layers)-1].PaintingArea.Rect
		if !paintingArea.IsNone() {
			if (bleed != bleedData{}) {
				// Painting area is the PDF BleedBox
				paintingArea[0] -= bleed.Left
				paintingArea[1] -= bleed.Top
				paintingArea[2] += bleed.Left + bleed.Right
				paintingArea[3] += bleed.Top + bleed.Bottom
			}
			context.Rectangle(paintingArea.Unpack())
			context.Clip()
		}
		context.SetSourceRgba(bg.Color.Unpack())
		context.Paint()
		// }
	}

	if (bleed != bleedData{}) && !marks.IsNone() {
		x, y, width, height := bg.Layers[len(bg.Layers)-1].PaintingArea.Rect.Unpack2()
		x -= bleed.Left
		y -= bleed.Top
		width += bleed.Left + bleed.Right
		height += bleed.Top + bleed.Bottom
		svg := headerSVG
		if marks.Crop {
			svg += crop
		}
		if marks.Cross {
			svg += cross
		}
		svg += "</svg>"
		halfBleed := bleedData{
			Top:    bleed.Top * 0.5,
			Bottom: bleed.Bottom * 0.5,
			Left:   bleed.Left * 0.5,
			Right:  bleed.Right * 0.5,
		}
		svg, err := formatSVG(svg, svgArgs{Width: width, Height: height, Bleed: bleed, HalfBleed: halfBleed})
		if err != nil {
			return err
		}
		image, err := images.NewSVGImage(svg, "", nil)
		if err != nil {
			return err
		}

		// Painting area is the PDF media box
		size := pr.Size{Width: width.ToValue(), Height: height.ToValue()}
		position := bo.Position{Point: bo.MaybePoint{x, y}}
		repeat := bo.Repeat{Reps: [2]string{"no-repeat", "no-repeat"}}
		unbounded := true
		paintingArea := bo.Area{Rect: pr.Rectangle{x, y, width, height}}
		positioningArea := bo.Area{Rect: pr.Rectangle{0, 0, width, height}}
		layer := bo.BackgroundLayer{Image: image, Size: size, Position: position, Repeat: repeat, Unbounded: unbounded,
			PaintingArea: paintingArea, PositioningArea: positioningArea}
		bg.Layers = append([]bo.BackgroundLayer{layer}, bg.Layers...)
	}
	// Paint in reversed order: first layer is "closest" to the viewer.
	for _, layer := range reversed(bg.Layers) {
		drawBackgroundImage(context, layer, bg.ImageRendering)
	}
	return nil
}

// Draw the background color && image of the table children.
func drawTableBackgrounds(context Drawer, page *bo.PageBox, table_ bo.InstanceTableBox, enableHinting bool) {
	table := table_.Table()
	for _, columnGroup := range table.ColumnGroups {
		drawBackground2(context, columnGroup.Box().Background, enableHinting)
		for _, column := range columnGroup.Box().Children {
			drawBackground2(context, column.Box().Background, enableHinting)
		}
	}
	for _, rowGroup := range table.Children {
		drawBackground2(context, rowGroup.Box().Background, enableHinting)
		for _, row := range rowGroup.Box().Children {
			drawBackground2(context, row.Box().Background, enableHinting)
			for _, cell := range row.Box().Children {
				cell := cell.Box()
				if table.Style.GetBorderCollapse() == "collapse" ||
					cell.Style.GetEmptyCells() == "show" || !cell.Empty {
					drawBackground2(context, cell.Background, enableHinting)
				}
			}
		}
	}
}

func drawBackgroundImage(context Drawer, layer bo.BackgroundLayer, imageRendering pr.String) {
	if layer.Image == nil {
		return
	}

	paintingX, paintingY, paintingWidth, paintingHeight := layer.PaintingArea.Rect.Unpack()
	positioningX, positioningY, positioningWidth, positioningHeight := layer.PositioningArea.Rect.Unpack()
	positionX, positionY := layer.Position.Point[0], layer.Position.Point[1]
	repeatX, repeatY := layer.Repeat.Reps[0], layer.Repeat.Reps[1]
	imageWidth, imageHeight := float64(layer.Size.Width.Value), float64(layer.Size.Height.Value)
	var repeatWidth, repeatHeight float64
	switch repeatX {
	case "no-repeat":
		// We want at least the whole imageWidth drawn on subSurface, but we
		// want to be sure it will not be repeated on the paintingWidth.
		repeatWidth = math.Max(imageWidth, paintingWidth)
	case "repeat", "round":
		// We repeat the image each imageWidth.
		repeatWidth = imageWidth
	case "space":
		nRepeats := math.Floor(positioningWidth / imageWidth)
		if nRepeats >= 2 {
			// The repeat width is the whole positioning width with one image
			// removed, divided by (the number of repeated images - 1). This
			// way, we get the width of one image + one space. We ignore
			// background-position for this dimension.
			repeatWidth = (positioningWidth - imageWidth) / (nRepeats - 1)
			positionX = pr.Float(0)
		} else {
			// We don't repeat the image.
			repeatWidth = imageWidth
		}
	default:
		log.Fatalf("unexpected repeatX %s", repeatX)
	}

	// Comments above apply here too.
	switch repeatY {
	case "no-repeat":
		repeatHeight = math.Max(imageHeight, paintingHeight)
	case "repeat", "round":
		repeatHeight = imageHeight
	case "space":
		nRepeats := math.Floor(positioningHeight / imageHeight)
		if nRepeats >= 2 {
			repeatHeight = (positioningHeight - imageHeight) / (nRepeats - 1)
			positionY = pr.Float(0)
		} else {
			repeatHeight = imageHeight
		}
	default:
		log.Fatalf("unexpected repeatY %s", repeatY)
	}

	subSurface := cairo.PDFSurface(nil, repeatWidth, repeatHeight)
	var subContext Drawer = cairo.Context(subSurface)
	subContext.Rectangle(0, 0, imageWidth, imageHeight)
	subContext.Clip()
	layer.Image.Draw(subContext, imageWidth, imageHeight, string(imageRendering))
	pattern := cairo.SurfacePattern(subSurface)

	if repeatX == "no-repeat" && repeatY == "no-repeat" {
		pattern.setExtend(cairo.EXTENDNONE)
	} else {
		pattern.setExtend(cairo.EXTENDREPEAT)
	}

	// with stacked(context) {
	if !layer.Unbounded {
		context.Rectangle(float64(paintingX), float64(paintingY),
			paintingWidth, paintingHeight)
		context.Clip()
	} // else: unrestricted, whole page box

	context.Translate(positioningX+float64(positionX.V()),
		positioningY+float64(positionY.V()))
	context.SetSource(pattern)
	context.Paint()
}

// Increment X and Y coordinates by the given offsets.
func xyOffset(x, y, offsetX, offsetY, offset float64) (float64, float64) {
	return x + offsetX*offset, y + offsetY*offset
}

func styledColor(style pr.String, color Color, side string) []Color {
	if style == "inset" || style == "outset" {
		doLighten := (side == "top" || side == "left") != (style == "inset")
		if doLighten {
			return []Color{lighten(color)}
		}
		return []Color{darken(color)}
	} else if style == "ridge" || style == "groove" {
		if (side == "top" || side == "left") != (style == "ridge") {
			return []Color{lighten(color), darken(color)}
		} else {
			return []Color{darken(color), lighten(color)}
		}
	}
	return []Color{color}
}

// Draw the box border to a ``cairo.Context``.
func drawBorder(context Drawer, box_ Box, enableHinting bool) {
	// We need a plan to draw beautiful borders, and that's difficult, no need
	// to lie. Let's try to find the cases that we can handle in a smart way.
	box := box_.Box()

	// Draw column borders.
	drawColumnBorder := func() {
		columns := bo.IsBlockContainerBox(box_) && (box.Style.GetColumnWidth().String != "auto" || box.Style.GetColumnCount().String != "auto")
		if crw := box.Style.GetColumnRuleWidth(); columns && !crw.IsNone() {
			borderWidths := pr.Rectangle{0, 0, 0, crw.Value}
			for _, child := range box.Children[1:] {
				// with stacked(context) {
				positionX := child.Box().PositionX - (crw.Value+
					box.Style.GetColumnGap().Value)/2
				borderBox := pr.Rectangle{positionX, child.Box().PositionY,
					crw.Value, box.Height.V()}
				clipBorderSegment(context, enableHinting,
					box.Style.GetColumnRuleStyle(),
					float64(crw.Value), "left", borderBox,
					&borderWidths, nil)
				drawRectBorder(context, borderBox, borderWidths,
					box.Style.GetColumnRuleStyle(), styledColor(
						box.Style.GetColumnRuleStyle(),
						box.Style.ResolveColor("column_rule_color").RGBA, "left"))
			}
		}
	}

	// The box is hidden, easy.
	if box.Style.GetVisibility() != "visible" {
		drawColumnBorder()
		return
	}

	widths := pr.Rectangle{box.BorderTopWidth.V(), box.BorderRightWidth.V(), box.BorderBottomWidth.V(), box.BorderLeftWidth.V()}

	// No border, return early.
	if widths.IsNone() {
		drawColumnBorder()
		return
	}
	var (
		colors    [4]Color
		colorsSet = map[Color]bool{}
		styles    [4]pr.String
		stylesSet = pr.NewSet()
	)
	for i, side := range SIDES {
		colors[i] = box.Style.ResolveColor(fmt.Sprintf("border_%s_color", side)).RGBA
		colorsSet[colors[i]] = true
		if colors[i].A != 0 {
			styles[i] = box.Style[fmt.Sprintf("border_%s_style", side)].(pr.String)
		}
		stylesSet.Add(string(styles[i]))
	}

	// The 4 sides are solid or double, and they have the same color. Oh yeah!
	// We can draw them so easily!
	if len(stylesSet) == 1 && (stylesSet.Has("solid") || stylesSet.Has("double")) && len(colorsSet) == 1 {
		drawRoundedBorder(context, box_, styles[0], []Color{colors[0]})
		drawColumnBorder()
		return
	}

	// We're not smart enough to find a good way to draw the borders :/. We must
	// draw them side by side.
	for i, side := range SIDES {
		width, color, style := widths[i], colors[i], styles[i]
		if width == 0 || color.IsNone() {
			continue
		}
		// with stacked(context) {
		rb := box.RoundedBorderBox()
		roundedBox := pr.Rectangle{rb.X, rb.Y, rb.Width, rb.Height}
		radii := [4]bo.Point{rb.TopLeft, rb.TopRight, rb.BottomRight, rb.BottomLeft}
		clipBorderSegment(context, enableHinting, style, float64(width), side,
			roundedBox, &widths, &radii)
		drawRoundedBorder(context, box_, style, styledColor(style, color, side))
		// }
	}

	drawColumnBorder()
}

// Clip one segment of box border.
// The strategy is to remove the zones not needed because of the style or the
// side before painting.
// border_widths=None, radii=None
func clipBorderSegment(context Drawer, enableHinting bool, style pr.String, width float64, side string,
	borderBox pr.Rectangle, borderWidths *pr.Rectangle, radii *[4]bo.Point) {

	if enableHinting && style != "dotted" && (
	// Borders smaller than 1 device unit would disappear
	// without anti-aliasing.
	math.Hypot(context.UserToDevice(float64(width), 0)) >= 1 &&
		math.Hypot(context.UserToDevice(0, float64(width))) >= 1) {
		// Avoid an artifact in the corner joining two solid borders
		// of the same color.
		context.SetAntialias(cairo.ANTIALIASNONE)
	}

	bbx, bby, bbw, bbh := borderBox.Unpack()
	var tlh, tlv, trh, trv, brh, brv, blh, blv float64
	if radii != nil {
		tlh, tlv, trh, trv, brh, brv, blh, blv = float64((*radii)[0][0]), float64((*radii)[0][1]), float64((*radii)[1][0]), float64((*radii)[1][1]), float64((*radii)[2][0]), float64((*radii)[2][1]), float64((*radii)[3][0]), float64((*radii)[3][1])
	}
	bt, br, bb, bl := width, width, width, width
	if borderWidths != nil {
		bt, br, bb, bl = borderWidths.Unpack()
	}

	// Get the point use for border transition.
	// The extra boolean returned is ``true`` if the point is in the padding
	// box (ie. the padding box is rounded).
	// This point is not specified. We must be sure to be inside the rounded
	// padding box, and in the zone defined in the "transition zone" allowed
	// by the specification. We chose the corner of the transition zone. It's
	// easy to get and gives quite good results, but it seems to be different
	// from what other browsers do.
	transitionPoint := func(x1, y1, x2, y2 float64) (float64, float64, bool) {
		if math.Abs(x1) > math.Abs(x2) && math.Abs(y1) > math.Abs(y2) {
			return x1, y1, true
		}
		return x2, y2, false
	}

	// Return the length of the half of one ellipsis corner.

	// Inspired by [Ramanujan, S., "Modular Equations and Approximations to
	// pi" Quart. J. Pure. Appl. Math., vol. 45 (1913-1914), pp. 350-372],
	// wonderfully explained by Dr Rob.

	// http://mathforum.org/dr.math/faq/formulas/
	cornerHalfLength := func(a, b float64) float64 {
		x := (a - b) / (a + b)
		return pi / 8 * (a + b) * (1 + 3*x*x/(10+math.Sqrt(4-3*x*x)))
	}
	var (
		px1, px2, py1, py2, way, angle, mainOffset float64
		rounded1, rounded2                         bool
	)
	if side == "top" {
		px1, py1, rounded1 = transitionPoint(tlh, tlv, bl, bt)
		px2, py2, rounded2 = transitionPoint(-trh, trv, -br, bt)
		width = bt
		way = 1
		angle = 1
		mainOffset = bby
	} else if side == "right" {
		px1, py1, rounded1 = transitionPoint(-trh, trv, -br, bt)
		px2, py2, rounded2 = transitionPoint(-brh, -brv, -br, -bb)
		width = br
		way = 1
		angle = 2
		mainOffset = bbx + bbw
	} else if side == "bottom" {
		px1, py1, rounded1 = transitionPoint(blh, -blv, bl, -bb)
		px2, py2, rounded2 = transitionPoint(-brh, -brv, -br, -bb)
		width = bb
		way = -1
		angle = 3
		mainOffset = bby + bbh
	} else if side == "left" {
		px1, py1, rounded1 = transitionPoint(tlh, tlv, bl, bt)
		px2, py2, rounded2 = transitionPoint(blh, -blv, bl, -bb)
		width = bl
		way = -1
		angle = 4
		mainOffset = bbx
	}
	var a1, b1, a2, b2, lineLength, length float64
	if side == "top" || side == "bottom" {
		a1, b1 = px1-bl/2, way*py1-width/2
		a2, b2 = -px2-br/2, way*py2-width/2
		lineLength = bbw - px1 + px2
		length = bbw
		context.MoveTo(bbx+bbw, mainOffset)
		context.RelLineTo(-bbw, 0)
		context.RelLineTo(px1, py1)
		context.RelLineTo(-px1+bbw+px2, -py1+py2)
	} else if side == "left" || side == "right" {
		a1, b1 = -way*px1-width/2, py1-bt/2
		a2, b2 = -way*px2-width/2, -py2-bb/2
		lineLength = bbh - py1 + py2
		length = bbh
		context.MoveTo(mainOffset, bby+bbh)
		context.RelLineTo(0, -bbh)
		context.RelLineTo(px1, py1)
		context.RelLineTo(-px1+px2, -py1+bbh+py2)
	}

	context.SetFillRule(cairo.FILLRULEEVENODD)
	if style == "dotted" || style == "dashed" {
		dash := 3 * width
		if style == "dotted" {
			dash = width
		}
		if rounded1 || rounded2 {
			// At least one of the two corners is rounded
			chl1 := cornerHalfLength(a1, b1)
			chl2 := cornerHalfLength(a2, b2)
			length = lineLength + chl1 + chl2
			dashLength := math.Round(length / dash)
			if rounded1 && rounded2 {
				// 2x dashes
				dash = length / (dashLength + utils.FloatModulo(dashLength, 2))
			} else {
				// 2x - 1/2 dashes
				dash = length / (dashLength + utils.FloatModulo(dashLength, 2) - 0.5)
			}
			dashes1 := int(math.Ceil((chl1 - dash/2) / dash))
			dashes2 := int(math.Ceil((chl2 - dash/2) / dash))
			line := int(math.Floor(lineLength / dash))

			drawDots := func(dashes, line int, way, x, y, px, py, chl float64) (int, float64) {
				if dashes == 0 {
					return line + 1, 0
				}
				var (
					hasBroken              bool
					offset, angle1, angle2 float64
				)
				for i_ := 0; i_ < dashes; i_ += 2 {
					i := float64(i_) + 0.5 // half dash
					angle1 = ((2*angle - way) + i*way*dash/chl) / 4 * pi

					fn := math.Max
					if way > 0 {
						fn = math.Min
					}
					angle2 = fn(
						((2*angle-way)+(i+1)*way*dash/chl)/4*pi,
						angle*pi/2,
					)
					if side == "top" || side == "bottom" {
						context.MoveTo(x+px, mainOffset+py)
						context.LineTo(x+px-way*px*1/math.Tan(angle2), mainOffset)
						context.LineTo(x+px-way*px*1/math.Tan(angle1), mainOffset)
					} else if side == "left" || side == "right" {
						context.MoveTo(mainOffset+px, y+py)
						context.LineTo(mainOffset, y+py+way*py*math.Tan(angle2))
						context.LineTo(mainOffset, y+py+way*py*math.Tan(angle1))
					}
					if angle2 == angle*pi/2 {
						offset = (angle1 - angle2) / ((((2*angle - way) + (i+1)*way*dash/chl) /
							4 * pi) - angle1)
						line += 1
						hasBroken = true
						break
					}
				}
				if !hasBroken {
					offset = 1 - (angle*pi/2-angle2)/(angle2-angle1)
				}
				return line, offset
			}
			var offset float64
			line, offset = drawDots(dashes1, line, way, bbx, bby, px1, py1, chl1)
			line, _ = drawDots(dashes2, line, -way, bbx+bbw, bby+bbh, px2, py2, chl2)

			if lineLength > 1e-6 {
				for i_ := 0; i_ < line; i_ += 2 {
					i := float64(i_) + offset
					var x1, x2, y1, y2 float64
					if side == "top" || side == "bottom" {
						x1 = math.Max(bbx+px1+i*dash, bbx+px1)
						x2 = math.Min(bbx+px1+(i+1)*dash, bbx+bbw+px2)
						y1 = mainOffset
						if way < 0 {
							y1 -= width
						}
						y2 = y1 + width
					} else if side == "left" || side == "right" {
						y1 = math.Max(bby+py1+i*dash, bby+py1)
						y2 = math.Min(bby+py1+(i+1)*dash, bby+bbh+py2)
						x1 = mainOffset
						if way > 0 {
							x1 -= width
						}
						x2 = x1 + width
					}
					context.Rectangle(x1, y1, x2-x1, y2-y1)
				}
			}
		} else {
			// 2x + 1 dashes
			context.Clip()
			denom := math.Round(length/dash) - utils.FloatModulo(math.Round(length/dash)+1, 2)
			dash = length
			if denom != 0 {
				dash /= denom
			}
			maxI := int(math.Round(length / dash))
			for i_ := 0; i_ < maxI; i_ += 2 {
				i := float64(i_)
				switch side {
				case "top":
					context.Rectangle(bbx+i*dash, bby, dash, width)
				case "right":
					context.Rectangle(bbx+bbw-width, bby+i*dash, width, dash)
				case "bottom":
					context.Rectangle(bbx+i*dash, bby+bbh-width, dash, width)
				case "left":
					context.Rectangle(bbx, bby+i*dash, width, dash)
				}
			}
		}
	}
	context.Clip()
}

func drawRoundedBorder(context Drawer, box bo.BoxFields, style pr.String, color []Color) {
	context.SetFillRule(cairo.FILLRULEEVENODD)
	roundedBoxPath(context, box.RoundedPaddingBox())
	if style == "ridge" || style == "groove" {
		roundedBoxPath(context, box.RoundedBoxRatio(1/2))
		context.SetSourceRgba(*color[0])
		context.fill()
		roundedBoxPath(context, box.RoundedBoxRatio(1/2))
		roundedBoxPath(context, box.RoundedBorderBox())
		context.SetSourceRgba(*color[1])
		context.fill()
		return
	}
	if style == "double" {
		roundedBoxPath(context, box.RoundedBoxRatio(1/3))
		roundedBoxPath(context, box.RoundedBoxRatio(2/3))
	}
	roundedBoxPath(context, box.RoundedBorderBox())
	context.SetSourceRgba(*color)
	context.fill()
}

func drawRectBorder(context Drawer, box [4]Point, widths pr.Rectangle, style pr.String, color []Color) {
	context.SetFillRule(cairo.FILLRULEEVENODD)
	bbx, bby, bbw, bbh = box
	bt, br, bb, bl = widths
	context.Rectangle(*box)
	if style == "ridge" || style == "groove" {
		context.Rectangle(bbx+bl/2, bby+bt/2,
			bbw-(bl+br)/2, bbh-(bt+bb)/2)
		context.SetSourceRgba(*color[0])
		context.fill()
		context.Rectangle(bbx+bl/2, bby+bt/2,
			bbw-(bl+br)/2, bbh-(bt+bb)/2)
		context.Rectangle(bbx+bl, bby+bt, bbw-bl-br, bbh-bt-bb)
		context.SetSourceRgba(*color[1])
		context.fill()
		return
	}
	if style == "double" {
		context.Rectangle(
			bbx+bl/3, bby+bt/3,
			bbw-(bl+br)/3, bbh-(bt+bb)/3)
		context.Rectangle(
			bbx+bl*2/3, bby+bt*2/3,
			bbw-(bl+br)*2/3, bbh-(bt+bb)*2/3)
	}
	context.Rectangle(bbx+bl, bby+bt, bbw-bl-br, bbh-bt-bb)
	context.SetSourceRgba(*color)
	context.fill()
}

func drawOutlines(context Drawer, box Box, enableHinting bool) {
	width = box.Style["outlineWidth"]
	color = getColor(box.Style, "outlineColor")
	style = box.Style["outlineStyle"]
	if box.Style["visibility"] == "visible" && width && color.alpha {
		outlineBox = [4]pr.Point{box.BorderBoxX() - width, box.BorderBoxY() - width,
			box.BorderWidth() + 2*width, box.BorderHeight() + 2*width}
		for _, side := range SIDES {
			// with stacked(context) {
			clipBorderSegment(context, enableHinting, style, width, side, outlineBox)
			drawRectBorder(context, outlineBox, [4]float64{width, width, width, width},
				style, styledColor(style, color, side))
			// }
		}
	}

	if bo.IsParentBox(box) {
		for _, child := range box.children {
			if bo.IsBox(child) {
				drawOutlines(context, child, enableHinting)
			}
		}
	}
}

type segment struct{}

// Draw borders of table cells when they collapse.
func drawCollapsedBorders(context Drawer, table *bo.TableBox, enableHinting bool) {
	var rowHeights, rowPositions []float64
	for _, rowGroup := range table.children {
		for _, row := range rowGroup.children {
			rowHeights = append(rowHeights, row.height)
			rowPositions = append(rowPositions, row.positionY)
		}
	}
	columnWidths = table.columnWidths
	if !(rowHeights && columnWidths) {
		// One of the list is empty: don’t bother with empty tables
		return
	}
	columnPositions = list(table.columnPositions)
	gridHeight = len(rowHeights)
	gridWidth = len(columnWidths)
	if gridWidth != len(columnPositions) {
		log.Fatalf("expected same gridWidth and columnPositions length, got %d, %d", gridWidth, len(columnPositions))
	}
	// Add the end of the last column, but make a copy from the table attr.
	columnPositions = append(columnPositions, columnPositions[len(columnPositions)-1]+columnWidths[len(columnWidths)-1])
	// Add the end of the last row. No copy here, we own this list
	rowPositions = append(rowPositions, rowPositions[len(rowPositions)-1]+rowHeights[len(rowHeights)-1])
	verticalBorders, horizontalBorders = table.collapsedBorderGrid
	if table.children[0].isHeader {
		headerRows = len(table.children[0].children)
	} else {
		headerRows = 0
	}
	if table.children[-1].isFooter {
		footerRows = len(table.children[-1].children)
	} else {
		footerRows = 0
	}
	skippedRows = table.skippedRows
	if skippedRows {
		bodyRowsOffset = skippedRows - headerRows
	} else {
		bodyRowsOffset = 0
	}
	if headerRows == 0 {
		headerRows = -1
	}
	if footerRows {
		firstFooterRow = gridHeight - footerRows - 1
	} else {
		firstFooterRow = gridHeight + 1
	}
	originalGridHeight = len(verticalBorders)
	footerRowsOffset = originalGridHeight - gridHeight

	rowNumber := func(y float64, horizontal bool) float64 {
		if y < (headerRows + int(horizontal)) {
			return y
		} else if y >= (firstFooterRow + int(horizontal)) {
			return y + footerRowsOffset
		} else {
			return y + bodyRowsOffset
		}
	}

	var segments []segment

	//vertical=true
	halfMaxWidth := func(borderList, yxPairs, vertical bool) float64 {
		result = 0
		for y, x := range yxPairs {
			cond := 0 <= y <= gridHeight && 0 <= x < gridWidth
			if vertical {
				cond = 0 <= y < gridHeight && 0 <= x <= gridWidth
			}
			if cond {
				yy = rowNumber(y, !vertical)
				tmp = borderList[yy][x]
				// _, (_, width, ) = tmp
				result = max(result, width)
			}
		}
		return result / 2
	}

	addVertical := func(x, y float64) {
		yy = rowNumber(y, false)
		tmp = verticalBorders[yy][x]
		// score, (style, width, color) = tmp
		if width == 0 || color.alpha == 0 {
			return
		}
		posX = columnPositions[x]
		posY1 = rowPositions[y] - halfMaxWidth(horizontalBorders,
			[][2]float64{{y, x - 1}, {y, x}}, false)
		posY2 = rowPositions[y+1] + halfMaxWidth(horizontalBorders,
			[][2]float64{{y + 1, x - 1}, {y + 1, x}}, false)
		segments = append(segments, segment{score, style, width, color, "left",
			{posX - width/2, posY1, 0, posY2 - posY1}})
	}

	addHorizontal := func(x, y float64) {
		yy = rowNumber(y, true)
		tmp = horizontalBorders[yy][x]
		// score, (style, width, color) = tmp
		if width == 0 || color.alpha == 0 {
			return
		}
		posY = rowPositions[y]
		// TODO: change signs for rtl when we support rtl tables?
		posX1 = columnPositions[x] - halfMaxWidth(verticalBorders,
			[][2]float64{{y - 1, x}, {y, x}}, true)
		posX2 = columnPositions[x+1] + halfMaxWidth(verticalBorders,
			[][2]float64{{y - 1, x + 1}, {y, x + 1}}, true)
		segments = append(segments, segment{score, style, width, color, "top",
			{posX1, posY - width/2, posX2 - posX1}, 0})
	}

	for x := 0; x < gridWidth; x += 1 {
		addHorizontal(x, 0)
	}
	for y := 0; y < gridHeight; y += 1 {
		addVertical(0, y)
		for x := 0; x < gridWidth; x += 1 {
			addVertical(x+1, y)
			addHorizontal(x, y+1)
		}
	}

	// Sort bigger scores last (painted later, on top)
	// Since the number of different scores is expected to be small compared
	// to the number of segments, there should be little changes and Timsort
	// should be closer to O(n) than O(n * log(n))
	sort.Slice(segments, func(i, j int) bool {
		return segments[i][0] < segments[j][0] // key=operator.itemgetter(0)
	})

	for _, segment := range segments {
		_, style, width, color, side, borderBox = segment
		if side == "top" {
			widths = [4]float64{width, 0, 0, 0}
		} else {
			widths = [4]float64{0, 0, 0, width}
		}
		// with stacked(context) {
		clipBorderSegment(
			context, enableHinting, style, width, side, borderBox,
			widths)
		drawRectBorder(
			context, borderBox, widths, style,
			styledColor(style, color, side))
		// }
	}
}

// Draw the given :class:`boxes.ReplacedBox` to a ``cairo.context``.
func drawReplacedbox(context Drawer, box bo.InstanceReplacedBox) {
	if box.Style["visibility"] != "visible" || !box.width || !box.height {
		return
	}

	drawWidth, drawHeight, drawX, drawY = replaced.replacedboxLayout(box)

	// with stacked(context) {
	roundedBoxPath(context, box.RoundedContentBox())
	context.Clip()
	context.Translate(drawX, drawY)
	box.Replacement.draw(context, drawWidth, drawHeight, box.Style["imageRendering"])
	// }
}

// offsetX=0, textOverflow="clip"
func drawInlineLevel(context Drawer, page, box Box, enableHinting bool, offsetX float64, textOverflow string) {
	if stackingContext, ok := box.(StackingContext); ok {
		if !(bo.TypeInlineBlockBox.IsInstance(stackingContext.box) || bo.TypeInlineFlexBox.IsInstance(stackingContext.box)) {
			log.Fatalf("expected InlineBlock or InlineFlex, got %v", stackingContext.box)
		}
		drawStackingContext(context, stackingContext, enableHinting)
	} else {
		drawBackground(context, box.Background, enableHinting)
		drawBorder(context, box, enableHinting)
		textBox, isTextBox := box.(*bo.TextBox)
		if layout.IsLine(box) {
			if isinstance(box, bo.LineBox) {
				textOverflow = box.textOverflow
			}
			for _, child := range box.children {
				if _, ok := child.(StackingContext); ok {
					childOffsetX = offsetX
				} else {
					childOffsetX = offsetX + child.positionX - box.positionX
				}
				if child, ok := child.(*bo.TextBox); ok {
					drawText(context, child, enableHinting, childOffsetX, textOverflow)
				} else {
					drawInlineLevel(context, page, child, enableHinting, childOffsetX, textOverflow)
				}
			}
		} else if isinstance(box, bo.InlineReplacedBox) {
			drawReplacedbox(context, box)
		} else if isTextBox {
			// Should only happen for list markers
			drawText(context, box, enableHinting, offsetX, textOverflow)
		} else {
			log.Fatalf("unexpected box %v", box)
		}
	}
}

// Draw ``textbox`` to a ``cairo.Context`` from ``PangoCairo.Context``.
// offsetX=0,textOverflow="clip"
func drawText(context Drawer, textbox *bo.TextBox, enableHinting bool, offsetX float64, textOverflow string) {
	// Pango crashes with font-size: 0
	// FIXME: assert textbox.Style["fontSize"]

	if textbox.Style["visibility"] != "visible" {
		return
	}

	context.MoveTo(textbox.positionX, textbox.positionY+textbox.Baseline)
	context.SetSourceRgba(*textbox.Style["color"])

	textbox.pangoLayout.reactivate(textbox.Style)
	showFirstLine(context, textbox, textOverflow)

	values = textbox.Style["textDecorationLine"]

	thickness = textbox.Style["fontSize"] / 18 // Like other browsers do
	if enableHinting && thickness < 1 {
		thickness = 1
	}

	color = textbox.Style["textDecorationColor"]
	if color == "currentColor" {
		color = textbox.Style["color"]
	}

	if values.Has("overline") || values.Has("line-through") || values.Has("underline") {
		metrics = textbox.pangoLayout.getFontMetrics()
	}
	if values.Has("overline") {
		drawTextDecoration(
			context, textbox, offsetX,
			textbox.Baseline-metrics.ascent+thickness/2,
			thickness, enableHinting, color)
	}
	if values.Has("underline") {
		drawTextDecoration(
			context, textbox, offsetX,
			textbox.Baseline-metrics.underlinePosition+thickness/2,
			thickness, enableHinting, color)
	}
	if values.Has("line-through") {
		drawTextDecoration(context, textbox, offsetX,
			textbox.Baseline-metrics.strikethroughPosition,
			thickness, enableHinting, color)
	}

	textbox.pangoLayout.deactivate()
}

func drawWave(context Drawer, x, y, width, offsetX, radius float64) {
	context.NewPath()
	diameter = 2 * radius
	waveIndex = offsetX // diameter
	remain = offsetX - waveIndex*diameter

	for width > 0 {
		up = waveIndex%2 == 0
		centerX = x - remain + radius
		alpha1 = (1 + remain/diameter) * pi
		alpha2 = (1 + min(1, width/diameter)) * pi

		if up {
			context.Arc(centerX, y, radius, alpha1, alpha2)
		} else {
			context.ArcNegative(centerX, y, radius, -alpha1, -alpha2)
		}

		x += diameter - remain
		width -= diameter - remain
		remain = 0
		waveIndex += 1
	}
}

// Draw text-decoration of ``textbox`` to a ``cairo.Context``.
func drawTextDecoration(context Drawer, textbox *bo.TextBox, offsetX, offsetY, thickness float64,
	enableHinting bool, color Color) {

	style = textbox.Style["textDecorationStyle"]
	// with stacked(context) {
	if enableHinting {
		context.SetAntialias(cairo.ANTIALIASNONE)
	}
	context.SetSourceRgba(*color)
	context.SetLineWidth(thickness)

	if style == "dashed" {
		context.SetDash([]float64{5 * thickness}, offsetX)
	} else if style == "dotted" {
		context.SetDash([]float64{thickness}, offsetX)
	}

	if style == "wavy" {
		drawWave(
			context,
			textbox.positionX, textbox.positionY+offsetY,
			textbox.width, offsetX, 0.75*thickness)
	} else {
		context.MoveTo(textbox.positionX, textbox.positionY+offsetY)
		context.RelLineTo(textbox.width, 0)
	}

	if style == "double" {
		delta = 2 * thickness
		context.MoveTo(
			textbox.positionX, textbox.positionY+offsetY+delta)
		context.RelLineTo(textbox.width, 0)
	}

	context.Stroke()
}
