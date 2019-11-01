package goweasyprint

import (
	"log"
	"math"
	"sort"

	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/utils"

	bo "github.com/benoitkugler/go-weasyprint/boxes"
)

// Take an "after layout" box tree and draw it onto a cairo context.

var SIDES = [4]string{"top", "right", "bottom", "left"}

const (
	pi   = float64(math.Pi)
	CROP = `
  <!-- horizontal top left -->
  <path d="M0,{bleedTop} h{halfBleedLeft}" />
  <!-- horizontal top right -->
  <path d="M0,{bleedTop} h{halfBleedRight}"
        transform="translate({width},0) scale(-1,1)" />
  <!-- horizontal bottom right -->
  <path d="M0,{bleedBottom} h{halfBleedRight}"
        transform="translate({width},{height}) scale(-1,-1)" />
  <!-- horizontal bottom left -->
  <path d="M0,{bleedBottom} h{halfBleedLeft}"
        transform="translate(0,{height}) scale(1,-1)" />
  <!-- vertical top left -->
  <path d="M{bleedLeft},0 v{halfBleedTop}" />
  <!-- vertical bottom right -->
  <path d="M{bleedRight},0 v{halfBleedBottom}"
        transform="translate({width},{height}) scale(-1,-1)" />
  <!-- vertical bottom left -->
  <path d="M{bleedLeft},0 v{halfBleedBottom}"
        transform="translate(0,{height}) scale(1,-1)" />
  <!-- vertical top right -->
  <path d="M{bleedRight},0 v{halfBleedTop}"
        transform="translate({width},0) scale(-1,1)" />
`
	CROSS = `
  <!-- top -->
  <circle r="{halfBleedTop}"
          transform="scale(0.5)
                     translate({width},{halfBleedTop}) scale(0.5)" />
  <path d="M-{halfBleedTop},{halfBleedTop} h{bleedTop}
           M0,0 v{bleedTop}"
        transform="scale(0.5) translate({width},0)" />
  <!-- bottom -->
  <circle r="{halfBleedBottom}"
          transform="translate(0,{height}) scale(0.5)
                     translate({width},-{halfBleedBottom}) scale(0.5)" />
  <path d="M-{halfBleedBottom},-{halfBleedBottom} h{bleedBottom}
           M0,0 v-{bleedBottom}"
        transform="translate(0,{height}) scale(0.5) translate({width},0)" />
  <!-- left -->
  <circle r="{halfBleedLeft}"
          transform="scale(0.5)
                     translate({halfBleedLeft},{height}) scale(0.5)" />
  <path d="M{halfBleedLeft},-{halfBleedLeft} v{bleedLeft}
           M0,0 h{bleedLeft}"
        transform="scale(0.5) translate(0,{height})" />
  <!-- right -->
  <circle r="{halfBleedRight}"
          transform="translate({width},0) scale(0.5)
                     translate(-{halfBleedRight},{height}) scale(0.5)" />
  <path d="M-{halfBleedRight},-{halfBleedRight} v{bleedRight}
           M0,0 h-{bleedRight}"
        transform="translate({width},0)
                   scale(0.5) translate(0,{height})" />
`
)

// The context manager
// 	with stacked(context):
//		<body>
// is equivalent to
//		context.save()
//		try:
//			<body>
//  	finally:
//			context.restore()

// Transform a HSV color to a RGB color.
func hsv2rgb(hue, saturation, value float64) (r, g, b float64) {
	c := value * saturation
	x := c * (1 - utils.Abs(utils.FloatModulo(hue/60, 2)-1))
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
	cmax := math.Maxs(red, green, blue)
	cmin := math.Mins(red, green, blue)
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
	bleed := map[string]float64{}
	for _, side := range [4]string{"top", "right", "bottom", "left"} {
		bleed[side] = float64(page.Style["bleed_"+side].(pr.Value).Value)
	}
	marks := page.Style.GetMarks()
	stackingContext := NewStackingContextFromPage(page)
	drawBackground(context, stackingContext.box.Box().Background, enableHinting,
		false, bleed, marks)
	drawBackground(context, page.CanvasBackground, enableHinting, false, nil, pr.Marks{})
	drawBorder(context, page, enableHinting)
	drawStackingContext(context, stackingContext, enableHinting)
}

func drawBoxBackgroundAndBorder(context Drawer, page *bo.PageBox, box Box, enableHinting bool) {
	drawBackground(context, box.Box().Background, enableHinting, true, nil, pr.Marks{})
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
			box.BorderBoxX()+right.Value,
			box.BorderBoxY()+top.Value,
			left.Value-right.Value,
			bottom.Value-top.Value)
		context.Clip()
	}

	if box.Style.GetOpacity() < 1 {
		context.PushGroup()
	}

	if box.TransformationMatrix != nil {
		if err = box.TransformationMatrix.copy().invert(); err != nil { // } except cairo.CairoError {
			return
		}
		context.Transform(box.transformationMatrix)
	}

	// Point 1 is done in drawPage

	// Point 2
	if bo.TypeBlockBox.IsInstance(box) || bo.TypeMarginBox.IsInstance(box) ||
		bo.TypeInlineBlockBox.IsInstance(box) || bo.TypeTableCellBox.IsInstance(box) ||
		bo.TypeFlexContainerBox.IsInstance(box) {

		// The canvas background was removed by setCanvasBackground
		drawBoxBackgroundAndBorder(context, stackingContext.page, box, enableHinting)

		// with stacked(context) {
		if box.Style["overflow"] != "visible" {
			// Only clip the content && the children:
			// - the background is already clipped
			// - the border must *not* be clipped
			roundedBoxPath(context, box.roundedPaddingBox())
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
		if bo.TypeInlineBox.IsInstance(box) {
			drawInlineLevel(context, stackingContext.page, box, enableHinting)
		}

		// Point 7
		for _, block := range append([]Box{box}, stackingContext.blocksAndCells...) {
			if block, ok := block.(bo.InstanceReplacedBox); ok {
				drawReplacedbox(context, block)
			} else {
				for _, child := range block.children {
					if bo.TypeLineBox.IsInstance(child) {
						drawInlineLevel(context, stackingContext.page, child, enableHinting)
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
		drawOutlines(context, box, enableHinting)

		if op := box.Style.GetOpacity(); op < 1 {
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
	x, y, w, h, tl, tr, br, bl = radii

	if tl[0] == 0 || tl[1] == 0 {
		tl = pr.Point{0, 0}
	}
	if tr[0] == 0 || tr[1] == 0 {
		tr = pr.Point{0, 0}
	}
	if br[0] == 0 || br[1] == 0 {
		br = pr.Point{0, 0}
	}
	if bl[0] == 0 || bl[1] == 0 {
		bl = pr.Point{0, 0}
	}

	if (tl == pr.Point{} && tr == pr.Point{} && br == pr.Point{} && bl == pr.Point{}) {
		// No radius, draw a rectangle
		context.Rectangle(x, y, w, h)
		return
	}

	context.moveTo(x, y)
	context.newSubPath()
	for i, v := range [4][2]pr.Point{
		{{0, 0}, tl}, {{w, 0}, tr}, {{w, h}, br}, {{0, h}, bl},
	} {
		w, h, rx, ry := v[0][0], v[0][1], v[1][0], v[1][1]
		context.save()
		context.Translate(x+w, y+h)
		radius := pr.Max(rx, ry)
		if radius != 0 {
			context.scale(pr.Min(rx/ry, 1), pr.Min(ry/rx, 1))
		}
		aw, ah := 1, 1
		if w != 0 {
			aw = -1
		}
		if h != 0 {
			ah = -1
		}
		context.arc(aw*radius, ah*radius, radius,
			(2+i)*pi/2, (3+i)*pi/2)
		context.Restore()
	}
}

// Draw the background color and image to a ``cairo.Context``
// If ``clipBox`` is set to ``false``, the background is not clipped to the
// border box of the background, but only to the painting area
// clipBox=true bleed=None marks=()
func drawBackground(context Drawer, bg *bo.Background, enableHinting, clipBox bool, bleed map[string]float64,
	marks pr.Marks) {
	if bg == nil {
		return
	}

	// with stacked(context) {
	if enableHinting {
		// Prefer crisp edges on background rectangles.
		context.setAntialias(cairo.ANTIALIASNONE)
	}

	if clipBox {
		for _, box := range bg.layers[len(bg.layers)-1].clippedBoxes {
			roundedBoxPath(context, box)
		}
		context.Clip()
	}

	// Background color
	if bg.color.alpha > 0 {
		// with stacked(context) {
		paintingArea = bg.layers[-1].paintingArea
		if paintingArea {
			if bleed {
				// Painting area is the PDF BleedBox
				x, y, width, height = paintingArea
				paintingArea = [4]float64{
					x - bleed["left"],
					y - bleed["top"],
					width + bleed["left"] + bleed["right"],
					height + bleed["top"] + bleed["bottom"],
				}
			}
			context.Rectangle(paintingArea)
			context.Clip()
		}
		context.setSourceRgba(bg.color)
		context.Paint()
		// }
	}

	if bleed && marks {
		x, y, width, height := bg.layers[-1].paintingArea
		x -= bleed["left"]
		y -= bleed["top"]
		width += bleed["left"] + bleed["right"]
		height += bleed["top"] + bleed["bottom"]
		svg := `
              <svg height="{height}" width="{width}"
                   fill="transparent" stroke="black" stroke-width="1"
                   xmlns="http://www.w3.org/2000/svg"
                   xmlns:xlink="http://www.w3.org/1999/xlink">
            `
		if marks.Crop {
			svg += CROP
		}
		if marks.Cross {
			svg += CROSS
		}
		svg += "</svg>"
		// halfBleed = {key: value * 0.5 for key, value := range bleed.items()}
		// image = SVGImage(svg.format(
		//     height=height, width=width,
		//     bleedLeft=bleed["left"], bleedRight=bleed["right"],
		//     bleedTop=bleed["top"], bleedBottom=bleed["bottom"],
		//     halfBleedLeft=halfBleed["left"],
		//     halfBleedRight=halfBleed["right"],
		//     halfBleedTop=halfBleed["top"],
		//     halfBleedBottom=halfBleed["bottom"],
		// ), "", None)
		// Painting area is the PDF media box
		size := pr.Size{width, height}
		position := bo.Position{x, y}
		repeat := bo.Repeat{Reps: {"no-repeat", "no-repeat"}}
		unbounded := true
		paintingArea := position + size
		positioningArea := bo.Area{Rect: {0, 0, width, height}}
		layer = bo.BackgroundLayer{iamge: image, size, position, repeat, unbounded, paintingArea,
			positioningArea, nil}
		bg.layers.insert(0, layer)
	}
	// Paint in reversed order: first layer is "closest" to the viewer.
	for _, layer := range reversed(bg.layers) {
		drawBackgroundImage(context, layer, bg.imageRendering)
	}
}

// Draw the background color && image of the table children.
func drawTableBackgrounds(context Drawer, page *bo.PageBox, table bo.InstanceTableBox, enableHinting bool) {
	for _, columnGroup := range table.columnGroups {
		drawBackground(context, columnGroup.background, enableHinting)
		for _, column := range columnGroup.children {
			drawBackground(context, column.background, enableHinting)
		}
	}
	for _, rowGroup := range table.children {
		drawBackground(context, rowGroup.background, enableHinting)
		for _, row := range rowGroup.children {
			drawBackground(context, row.background, enableHinting)
			for _, cell := range row.children {
				if table.style["borderCollapse"] == "collapse" ||
					cell.style["emptyCells"] == "show" || !cell.empty {

					drawBackground(context, cell.background, enableHinting)
				}
			}
		}
	}
}

func drawBackgroundImage(context Drawer, layer bo.BackgroundLayer, imageRendering bool) {
	if layer.image == nil {
		return
	}

	paintingX, paintingY, paintingWidth, paintingHeight = layer.paintingArea
	positioningX, positioningY, positioningWidth, positioningHeight = layer.positioningArea
	positionX, positionY = layer.position
	repeatX, repeatY = layer.repeat
	imageWidth, imageHeight = layer.size

	switch repeatX {
	case "no-repeat":
		// We want at least the whole imageWidth drawn on subSurface, but we
		// want to be sure it will ! be repeated on the paintingWidth.
		repeatWidth = max(imageWidth, paintingWidth)
	case "repeat", "round":
		// We repeat the image each imageWidth.
		repeatWidth = imageWidth
	case "space":
		nRepeats = floor(positioningWidth / imageWidth)
		if nRepeats >= 2 {
			// The repeat width is the whole positioning width with one image
			// removed, divided by (the number of repeated images - 1). This
			// way, we get the width of one image + one space. We ignore
			// background-position for this dimension.
			repeatWidth = (positioningWidth - imageWidth) / (nRepeats - 1)
			positionX = 0
		} else {
			// We don"t repeat the image.
			repeatWidth = imageWidth
		}
	default:
		log.Fatalf("unexpected repeatX %s", repeatX)
	}

	// Comments above apply here too.
	switch repeatY {
	case "no-repeat":
		repeatHeight = max(imageHeight, paintingHeight)
	case "repeat", "round":
		repeatHeight = imageHeight
	case "space":
		nRepeats = floor(positioningHeight / imageHeight)
		if nRepeats >= 2 {
			repeatHeight = (positioningHeight - imageHeight) / (nRepeats - 1)
			positionY = 0
		} else {
			repeatHeight = imageHeight
		}
	default:
		log.Fatalf("unexpected repeatY %s", repeatY)
	}

	subSurface = cairo.PDFSurface(None, repeatWidth, repeatHeight)
	subContext = cairo.Context(subSurface)
	subContext.rectangle(0, 0, imageWidth, imageHeight)
	subContext.clip()
	layer.image.draw(subContext, imageWidth, imageHeight, imageRendering)
	pattern = cairo.SurfacePattern(subSurface)

	if repeatX == repeatY == "no-repeat" {
		pattern.setExtend(cairo.EXTENDNONE)
	} else {
		pattern.setExtend(cairo.EXTENDREPEAT)
	}

	// with stacked(context) {
	if !layer.unbounded {
		context.Rectangle(paintingX, paintingY,
			paintingWidth, paintingHeight)
		context.Clip()
	} // else: unrestricted, whole page box

	context.Translate(positioningX+positionX,
		positioningY+positionY)
	context.setSource(pattern)
	context.Paint()
}

// Increment X and Y coordinates by the given offsets.
func xyOffset(x, y, offsetX, offsetY, offset float64) (float64, float64) {
	return x + offsetX*offset, y + offsetY*offset
}

func styledColor(style string, color Color, side string) Color {
	if style == "inset" || style == "outset" {
		doLighten := (side == "top" || side == "left") != (style == "inset")
		if doLighten {
			return lighten(color)
		}
		return darken(color)
	} else if style == "ridge" || style == "groove" {
		if (side == "top" || side == "left") != (style == "ridge") {
			return lighten(color), darken(color)
		} else {
			return darken(color), lighten(color)
		}
	}
	return color
}

// Draw the box border to a ``cairo.Context``.
func drawBorder(context Drawer, box Box, enableHinting bool) {
	// We need a plan to draw beautiful borders, and that's difficult, no need
	// to lie. Let's try to find the cases that we can handle in a smart way.

	// Draw column borders.
	drawColumnBorder := func() {
		columns := bo.TypeBlockContainerBox.IsInstance(box) && (box.Style["columnWidth"] != "auto" || box.Style["columnCount"] != "auto")
		if columns && box.Style["columnRuleWidth"] {
			borderWidths := [4]float64{0, 0, 0, box.Style["columnRuleWidth"]}
			for _, child := range box.children[1:] {
				// with stacked(context) {
				positionX = child.positionX - (box.Style["columnRuleWidth"]+
									box.Style["columnGap"])/2
				borderBox = [4]float64{positionX, child.positionY,
					box.Style["columnRuleWidth"], box.height}
				clipBorderSegment(context, enableHinting,
					box.Style["columnRuleStyle"],
					box.Style["columnRuleWidth"], "left", borderBox,
					borderWidths)
				drawRectBorder(context, borderBox, borderWidths,
					box.Style["columnRuleStyle"], styledColor(
						box.Style["columnRuleStyle"],
						getColor(box.Style, "columnRuleColor"), "left"))

			}
		}
	}

	// The box is hidden, easy.
	if box.Style["visibility"] != "visible" {
		drawColumnBorder()
		return
	}

	widths = [4]float64{box, "border%sWidth"}

	// No border, return early.
	if widths == [4]float64{} {
		drawColumnBorder()
		return
	}
	var (
		colors    [4]Color
		colorsSet = map[Color]bool{}
		stylesSet = pr.NewSet()
	)
	for i, side := range SIDES {
		colors[i] = box.Style.ResolveColor("border_%s_color" % side)
		colorsSet[colors[i]] = true
		styles[i] = colors[i].alpha && box.Style["border%sStyle"%side]
		stylesSet.Add(styles[i])
	}

	// The 4 sides are solid or double, and they have the same color. Oh yeah!
	// We can draw them so easily!
	if len(stylesSet) == 1 && (stylesSet.Has("solid") || stylesSet.Has("double")) && len(colorsSet) == 1 {
		drawRoundedBorder(context, box, styles[0], colors[0])
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
		clipBorderSegment(context, enableHinting, style, width, side,
			box.roundedBorderBox()[:4], widths, box.roundedBorderBox()[4:])
		drawRoundedBorder(context, box, style, styledColor(style, color, side))
		// }
	}

	drawColumnBorder()
}

// Clip one segment of box border.
// The strategy is to remove the zones not needed because of the style or the
// side before painting.
func clipBorderSegment(context Drawer, enableHinting bool, style string, width, side,
	borderBox, borderWidths *[4]float64, radii *bo.RoundedBox) {

	if enableHinting && style != "dotted" && (
	// Borders smaller than 1 device unit would disappear
	// without anti-aliasing.
	hypot(context.userToDevice(width, 0)) >= 1 &&
		hypot(context.userToDevice(0, width)) >= 1) {
		// Avoid an artifact in the corner joining two solid borders
		// of the same color.
		context.setAntialias(cairo.ANTIALIASNONE)
	}

	bbx, bby, bbw, bbh = borderBox
	var tlh, tlv, trh, trv, brh, brv, blh, blv float64
	if radii != nil {
		tlh, tlv, trh, trv, brh, brv, blh, blv = radii
	}
	var bt, br, bb, bl = width
	if borderWidths != nil {
		bt, br, bb, bl = borderWidths[0], borderWidths[1], borderWidths[2], borderWidths[3]
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
		if utils.Abs(x1) > utils.Abs(x2) && utils.Abs(y1) > utils.Abs(y2) {
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
		return pi / 8 * (a + b) * (1 + 3*x**2/(10+sqrt(4-3*x**2)))
	}

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

	if side == "top" || side == "bottom" {
		a1, b1 = px1-bl/2, way*py1-width/2
		a2, b2 = -px2-br/2, way*py2-width/2
		lineLength = bbw - px1 + px2
		length = bbw
		context.moveTo(bbx+bbw, mainOffset)
		context.RelLineTo(-bbw, 0)
		context.RelLineTo(px1, py1)
		context.RelLineTo(-px1+bbw+px2, -py1+py2)
	} else if side == "left" || side == "right" {
		a1, b1 = -way*px1-width/2, py1-bt/2
		a2, b2 = -way*px2-width/2, -py2-bb/2
		lineLength = bbh - py1 + py2
		length = bbh
		context.moveTo(mainOffset, bby+bbh)
		context.RelLineTo(0, -bbh)
		context.RelLineTo(px1, py1)
		context.RelLineTo(-px1+px2, -py1+bbh+py2)
	}

	context.setFillRule(cairo.FILLRULEEVENODD)
	if style == "dotted" || style == "dashed" {
		dash := 3 * width
		if style == "dotted" {
			dash = width
		}
		if rounded1 || rounded2 {
			// At least one of the two corners is rounded
			chl1 = cornerHalfLength(a1, b1)
			chl2 = cornerHalfLength(a2, b2)
			length = lineLength + chl1 + chl2
			dashLength = round(length / dash)
			if rounded1 && rounded2 {
				// 2x dashes
				dash = length / (dashLength + dashLength%2)
			} else {
				// 2x - 1/2 dashes
				dash = length / (dashLength + dashLength%2 - 0.5)
			}
			dashes1 = int(ceil((chl1 - dash/2) / dash))
			dashes2 = int(ceil((chl2 - dash/2) / dash))
			line = int(floor(lineLength / dash))

			drawDots := func(dashes, line, way, x, y, px, py, chl float64) (float64, float64) {
				if !dashes {
					return line + 1, 0
				}
				var hasBroken bool
				for i := 0; i < dashes; i += 2 {
					i += 0.5 // half dash
					angle1 = ((2*angle - way) + i*way*dash/chl) / 4 * pi

					fn := pr.Max
					if way > 0 {
						fn = pr.Min
					}
					angle2 = fn(
						((2*angle-way)+(i+1)*way*dash/chl)/4*pi,
						angle*pi/2,
					)
					if side == "top" || side == "bottom" {
						context.moveTo(x+px, mainOffset+py)
						context.lineTo(x+px-way*px*1/tan(angle2), mainOffset)
						context.lineTo(x+px-way*px*1/tan(angle1), mainOffset)
					} else if side == "left" || side == "right" {
						context.moveTo(mainOffset+px, y+py)
						context.lineTo(mainOffset, y+py+way*py*tan(angle2))
						context.lineTo(mainOffset, y+py+way*py*tan(angle1))
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

			line, offset = drawDots(dashes1, line, way, bbx, bby, px1, py1, chl1)
			line, _ = drawDots(dashes2, line, -way, bbx+bbw, bby+bbh, px2, py2, chl2)

			if lineLength > 1e-6 {
				for i := 0; i < line; i += 2 {
					i += offset
					if side == "top" || side == "bottom" {
						x1 = max(bbx+px1+i*dash, bbx+px1)
						x2 = min(bbx+px1+(i+1)*dash, bbx+bbw+px2)
						y1 = mainOffset
						if way < 0 {
							y1 -= width
						}
						y2 = y1 + width
					} else if side == "left" || side == "right" {
						y1 = max(bby+py1+i*dash, bby+py1)
						y2 = min(bby+py1+(i+1)*dash, bby+bbh+py2)
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
			denom := round(length/dash) - (round(length/dash)+1)%2
			dash = length
			if denom != 0 {
				dash /= denom
			}
			maxI := int(round(length / dash))
			for i := 0; i < maxI; i += 2 {
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

func drawRoundedBorder(context Drawer, box Box, style, color Color) {
	context.setFillRule(cairo.FILLRULEEVENODD)
	roundedBoxPath(context, box.roundedPaddingBox())
	if style == "ridge" || style == "groove" {
		roundedBoxPath(context, box.roundedBoxRatio(1/2))
		context.setSourceRgba(*color[0])
		context.fill()
		roundedBoxPath(context, box.roundedBoxRatio(1/2))
		roundedBoxPath(context, box.roundedBorderBox())
		context.setSourceRgba(*color[1])
		context.fill()
		return
	}
	if style == "double" {
		roundedBoxPath(context, box.roundedBoxRatio(1/3))
		roundedBoxPath(context, box.roundedBoxRatio(2/3))
	}
	roundedBoxPath(context, box.roundedBorderBox())
	context.setSourceRgba(*color)
	context.fill()
}

func drawRectBorder(context Drawer, box [4]Point, widths, style, color Color) {
	context.setFillRule(cairo.FILLRULEEVENODD)
	bbx, bby, bbw, bbh = box
	bt, br, bb, bl = widths
	context.Rectangle(*box)
	if style == "ridge" || style == "groove" {
		context.Rectangle(bbx+bl/2, bby+bt/2,
			bbw-(bl+br)/2, bbh-(bt+bb)/2)
		context.setSourceRgba(*color[0])
		context.fill()
		context.Rectangle(bbx+bl/2, bby+bt/2,
			bbw-(bl+br)/2, bbh-(bt+bb)/2)
		context.Rectangle(bbx+bl, bby+bt, bbw-bl-br, bbh-bt-bb)
		context.setSourceRgba(*color[1])
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
	context.setSourceRgba(*color)
	context.fill()
}

func drawOutlines(context Drawer, box, enableHinting bool) {
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
	roundedBoxPath(context, box.roundedContentBox())
	context.Clip()
	context.Translate(drawX, drawY)
	box.replacement.draw(context, drawWidth, drawHeight, box.Style["imageRendering"])
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

	context.moveTo(textbox.positionX, textbox.positionY+textbox.Baseline)
	context.setSourceRgba(*textbox.Style["color"])

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
	context.newPath()
	diameter = 2 * radius
	waveIndex = offsetX // diameter
	remain = offsetX - waveIndex*diameter

	for width > 0 {
		up = waveIndex%2 == 0
		centerX = x - remain + radius
		alpha1 = (1 + remain/diameter) * pi
		alpha2 = (1 + min(1, width/diameter)) * pi

		if up {
			context.arc(centerX, y, radius, alpha1, alpha2)
		} else {
			context.arcNegative(centerX, y, radius, -alpha1, -alpha2)
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
		context.setAntialias(cairo.ANTIALIASNONE)
	}
	context.setSourceRgba(*color)
	context.setLineWidth(thickness)

	if style == "dashed" {
		context.setDash([]float64{5 * thickness}, offsetX)
	} else if style == "dotted" {
		context.setDash([]float64{thickness}, offsetX)
	}

	if style == "wavy" {
		drawWave(
			context,
			textbox.positionX, textbox.positionY+offsetY,
			textbox.width, offsetX, 0.75*thickness)
	} else {
		context.moveTo(textbox.positionX, textbox.positionY+offsetY)
		context.RelLineTo(textbox.width, 0)
	}

	if style == "double" {
		delta = 2 * thickness
		context.moveTo(
			textbox.positionX, textbox.positionY+offsetY+delta)
		context.RelLineTo(textbox.width, 0)
	}

	context.stroke()
}
