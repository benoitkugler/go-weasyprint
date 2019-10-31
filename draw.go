package goweasyprint

// Take an "after layout" box tree and draw it onto a cairo context.


var SIDES = [4]string{"top", "right", "bottom", "left"}

const (
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

// @contextlib.contextmanager
// // Save && restore the context when used with the ``with`` keyword.
// func stacked(context) {
//     context.save()
//     try {
//         yield
//     } finally {
//         context.restore()
//     }
// } 

// Transform a HSV color to a RGB color.
func hsv2rgb(hue, saturation, value pr.Float) (r,g,b pr.Float) {
    c := value * saturation
    x := c * (1 - abs((hue / 60) % 2 - 1))
    m := value - c
    if 0 <= hue < 60 {
        return c + m, x + m, m
    } else if 60 <= hue < 120 {
        return x + m, c + m, m
    } else if 120 <= hue < 180 {
        return m, c + m, x + m
    } else if 180 <= hue < 240 {
        return m, x + m, c + m
    } else if 240 <= hue < 300 {
        return x + m, m, c + m
    } else if 300 <= hue < 360 {
        return c + m, m, x + m
    }
} 

// Transform a RGB color to a HSV color.
func rgb2hsv(red, green, blue pr.Float) (h, s, c pr.Float) {
    cmax = max(red, green, blue)
    cmin = min(red, green, blue)
    delta = cmax - cmin
    if delta == 0 {
        hue = 0
    } else if cmax == red {
        hue = 60 * ((green - blue) / delta % 6)
    } else if cmax == green {
        hue = 60 * ((blue - red) / delta + 2)
    } else if cmax == blue {
        hue = 60 * ((red - green) / delta + 4)
	} 
	saturation := 0 
	if delta != 0 {
	saturation =	delta / cmax
	}
    return hue, saturation, cmax
} 

// Return a darker color.
func darken(color Color) Color{
    hue, saturation, value := rgb2hsv(color.R, color.G, color.B)
    value /= 1.5
	saturation /= 1.25
	r,g,b := hsv2rgb(hue, saturation, value) 
    return Color{R:r,G:g,B:b,A: color.A}
} 

// Return a lighter color.
func lighten(color Color)  Color{
    hue, saturation, value := rgb2hsv(color.R, color.G, color.B)
    value = 1 - (1 - value) / 1.5
    if saturation != 0 {
        saturation = 1 - (1 - saturation) / 1.25
	} 
	r,g,b := hsv2rgb(hue, saturation, value) 
    return Color{R:r,G:g,B:b,A: color.A}
} 

// Draw the given PageBox.
func drawPage(page *bo.PageBox, context, enableHinting bool) {
	bleed := map[string]pr.Float{}
	for _, side := range [4]string{"top", "right", "bottom", "left"} {
		bleed[side] = page.Style["bleed_" + side].(pr.Value).Value
	}
    marks := page.Style.GetMarks()
    stackingContext := NewStackingContextFromPage(page)
    drawBackground(context, stackingContext.box.Background, enableHinting,
        false, bleed, marks)
    drawBackground(context, page.canvasBackground, enableHinting, false)
    drawBorder(context, page, enableHinting)
    drawStackingContext(context, stackingContext, enableHinting)
} 

func drawBoxBackgroundAndBorder(context, page, box, enableHinting) {
    drawBackground(context, box.background, enableHinting)
    if isinstance(box, boxes.TableBox) {
        drawTableBackgrounds(context, page, box, enableHinting)
        if box.style["borderCollapse"] == "separate" {
            drawBorder(context, box, enableHinting)
            for rowGroup := range box.children {
                for row := range rowGroup.children {
                    for cell := range row.children {
                        if (cell.style["emptyCells"] == "show" or
                                ! cell.empty) {
                                }
                            drawBorder(context, cell, enableHinting)
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
func drawStackingContext(context, stackingContext, enableHinting) {
    // See http://www.w3.org/TR/CSS2/zindex.html
    with stacked(context) {
        box = stackingContext.box
        if box.isAbsolutelyPositioned() && box.style["clip"] {
            top, right, bottom, left = box.style["clip"]
            if top == "auto" {
                top = 0
            } if right == "auto" {
                right = 0
            } if bottom == "auto" {
                bottom = box.borderHeight()
            } if left == "auto" {
                left = box.borderWidth()
            } context.rectangle(
                box.borderBoxX() + right,
                box.borderBoxY() + top,
                left - right,
                bottom - top)
            context.clip()
        }
    }
} 
        if box.style["opacity"] < 1 {
            context.pushGroup()
        }

        if box.transformationMatrix {
            try {
                box.transformationMatrix.copy().invert()
            } except cairo.CairoError {
                return
            } else {
                context.transform(box.transformationMatrix)
            }
        }

        // Point 1 is done := range drawPage

        // Point 2
        if isinstance(box, (boxes.BlockBox, boxes.MarginBox,
                            boxes.InlineBlockBox, boxes.TableCellBox,
                            boxes.FlexContainerBox)) {
                            }
            // The canvas background was removed by setCanvasBackground
            drawBoxBackgroundAndBorder(
                context, stackingContext.page, box, enableHinting)

        with stacked(context) {
            if box.style["overflow"] != "visible" {
                // Only clip the content && the children {
                } // - the background is already clipped
                // - the border must *not* be clipped
                roundedBoxPath(context, box.roundedPaddingBox())
                context.clip()
            }
        }

            // Point 3
            for childContext := range stackingContext.negativeZContexts {
                drawStackingContext(context, childContext, enableHinting)
            }

            // Point 4
            for block := range stackingContext.blockLevelBoxes {
                drawBoxBackgroundAndBorder(
                    context, stackingContext.page, block, enableHinting)
            }

            // Point 5
            for childContext := range stackingContext.floatContexts {
                drawStackingContext(context, childContext, enableHinting)
            }

            // Point 6
            if isinstance(box, boxes.InlineBox) {
                drawInlineLevel(
                    context, stackingContext.page, box, enableHinting)
            }

            // Point 7
            for block := range [box] + stackingContext.blocksAndCells {
                if isinstance(block, boxes.ReplacedBox) {
                    drawReplacedbox(context, block)
                } else {
                    for child := range block.children {
                        if isinstance(child, boxes.LineBox) {
                            drawInlineLevel(
                                context, stackingContext.page, child,
                                enableHinting)
                        }
                    }
                }
            }

            // Point 8
            for childContext := range stackingContext.zeroZContexts {
                drawStackingContext(context, childContext, enableHinting)
            }

            // Point 9
            for childContext := range stackingContext.positiveZContexts {
                drawStackingContext(context, childContext, enableHinting)
            }

        // Point 10
        drawOutlines(context, box, enableHinting)

        if box.style["opacity"] < 1 {
            context.popGroupToSource()
            context.paintWithAlpha(box.style["opacity"])
        }


// Draw the path of the border radius box.
//     ``widths`` is a tuple of the inner widths (top, right, bottom, left) from
//     the border box. Radii are adjusted from these values. Default is (0, 0, 0,
//     0).
//     Inspired by cairo cookbook
//     http://cairographics.org/cookbook/roundedrectangles/
//     
func roundedBoxPath(context, radii) {
    x, y, w, h, tl, tr, br, bl = radii
} 
    if 0 := range tl {
        tl = (0, 0)
    } if 0 := range tr {
        tr = (0, 0)
    } if 0 := range br {
        br = (0, 0)
    } if 0 := range bl {
        bl = (0, 0)
    }

    if (tl, tr, br, bl) == 4 * ((0, 0),) {
        // No radius, draw a rectangle
        context.rectangle(x, y, w, h)
        return
    }

    context.moveTo(x, y)
    context.newSubPath()
    for i, (w, h, (rx, ry)) := range enumerate((
            (0, 0, tl), (w, 0, tr), (w, h, br), (0, h, bl))) {
            }
        context.save()
        context.translate(x + w, y + h)
        radius = max(rx, ry)
        if radius {
            context.scale(min(rx / ry, 1), min(ry / rx, 1))
        } context.arc(
            (-1 if w else 1) * radius, (-1 if h else 1) * radius, radius,
            (2 + i) * pi / 2, (3 + i) * pi / 2)
        context.restore()


func drawBackground(context, bg, enableHinting, clipBox=true, bleed=None,
                    marks=()) {
    """Draw the background color && image to a ``cairo.Context``.

    If ``clipBox`` is set to ``false``, the background is ! clipped to the
    border box of the background, but only to the painting area.

                    }
    """
    if bg  == nil  {
        return
    }

    with stacked(context) {
        if enableHinting {
            // Prefer crisp edges on background rectangles.
            context.setAntialias(cairo.ANTIALIASNONE)
        }
    }

        if clipBox {
            for box := range bg.layers[-1].clippedBoxes {
                roundedBoxPath(context, box)
            } context.clip()
        }

        // Background color
        if bg.color.alpha > 0 {
            with stacked(context) {
                paintingArea = bg.layers[-1].paintingArea
                if paintingArea {
                    if bleed {
                        // Painting area is the PDF BleedBox
                        x, y, width, height = paintingArea
                        paintingArea = (
                            x - bleed["left"], y - bleed["top"],
                            width + bleed["left"] + bleed["right"],
                            height + bleed["top"] + bleed["bottom"])
                    } context.rectangle(*paintingArea)
                    context.clip()
                } context.setSourceRgba(*bg.color)
                context.paint()
            }
        }

        if bleed && marks {
            x, y, width, height = bg.layers[-1].paintingArea
            x -= bleed["left"]
            y -= bleed["top"]
            width += bleed["left"] + bleed["right"]
            height += bleed["top"] + bleed["bottom"]
            svg = """
              <svg height="{height}" width="{width}"
                   fill="transparent" stroke="black" stroke-width="1"
                   xmlns="http://www.w3.org/2000/svg"
                   xmlns:xlink="http://www.w3.org/1999/xlink">
            """
            if "crop" := range marks:
                svg += CROP
            if "cross" := range marks:
                svg += CROSS
            svg += "</svg>"
            halfBleed = {key: value * 0.5 for key, value := range bleed.items()}
            image = SVGImage(svg.format(
                height=height, width=width,
                bleedLeft=bleed["left"], bleedRight=bleed["right"],
                bleedTop=bleed["top"], bleedBottom=bleed["bottom"],
                halfBleedLeft=halfBleed["left"],
                halfBleedRight=halfBleed["right"],
                halfBleedTop=halfBleed["top"],
                halfBleedBottom=halfBleed["bottom"],
            ), "", None)
            // Painting area is the PDF media box
            size = (width, height)
            position = (x, y)
            repeat = ("no-repeat", "no-repeat")
            unbounded = true
            paintingArea = position + size
            positioningArea = (0, 0, width, height)
            clippedBoxes = []
            layer = BackgroundLayer(
                image, size, position, repeat, unbounded, paintingArea,
                positioningArea, clippedBoxes)
            bg.layers.insert(0, layer)
        // Paint := range reversed order: first layer is "closest" to the viewer.
        for layer := range reversed(bg.layers):
            drawBackgroundImage(context, layer, bg.imageRendering)


// Draw the background color && image of the table children.
func drawTableBackgrounds(context, page, table, enableHinting):
    for columnGroup := range table.columnGroups:
        drawBackground(context, columnGroup.background, enableHinting)
        for column := range columnGroup.children:
            drawBackground(context, column.background, enableHinting)
    for rowGroup := range table.children:
        drawBackground(context, rowGroup.background, enableHinting)
        for row := range rowGroup.children:
            drawBackground(context, row.background, enableHinting)
            for cell := range row.children:
                if (table.style["borderCollapse"] == "collapse" or
                        cell.style["emptyCells"] == "show" or
                        ! cell.empty):
                    drawBackground(context, cell.background, enableHinting)


func drawBackgroundImage(context, layer, imageRendering):
    if layer.image  == nil :
        return

    paintingX, paintingY, paintingWidth, paintingHeight = (
        layer.paintingArea)
    positioningX, positioningY, positioningWidth, positioningHeight = (
        layer.positioningArea)
    positionX, positionY = layer.position
    repeatX, repeatY = layer.repeat
    imageWidth, imageHeight = layer.size

    if repeatX == "no-repeat":
        // We want at least the whole imageWidth drawn on subSurface, but we
        // want to be sure it will ! be repeated on the paintingWidth.
        repeatWidth = max(imageWidth, paintingWidth)
    else if repeatX := range ("repeat", "round"):
        // We repeat the image each imageWidth.
        repeatWidth = imageWidth
    else:
        assert repeatX == "space"
        nRepeats = floor(positioningWidth / imageWidth)
        if nRepeats >= 2:
            // The repeat width is the whole positioning width with one image
            // removed, divided by (the number of repeated images - 1). This
            // way, we get the width of one image + one space. We ignore
            // background-position for this dimension.
            repeatWidth = (positioningWidth - imageWidth) / (nRepeats - 1)
            positionX = 0
        else:
            // We don"t repeat the image.
            repeatWidth = imageWidth

    // Comments above apply here too.
    if repeatY == "no-repeat":
        repeatHeight = max(imageHeight, paintingHeight)
    else if repeatY := range ("repeat", "round"):
        repeatHeight = imageHeight
    else:
        assert repeatY == "space"
        nRepeats = floor(positioningHeight / imageHeight)
        if nRepeats >= 2:
            repeatHeight = (
                positioningHeight - imageHeight) / (nRepeats - 1)
            positionY = 0
        else:
            repeatHeight = imageHeight

    subSurface = cairo.PDFSurface(None, repeatWidth, repeatHeight)
    subContext = cairo.Context(subSurface)
    subContext.rectangle(0, 0, imageWidth, imageHeight)
    subContext.clip()
    layer.image.draw(subContext, imageWidth, imageHeight, imageRendering)
    pattern = cairo.SurfacePattern(subSurface)

    if repeatX == repeatY == "no-repeat":
        pattern.setExtend(cairo.EXTENDNONE)
    else:
        pattern.setExtend(cairo.EXTENDREPEAT)

    with stacked(context):
        if ! layer.unbounded:
            context.rectangle(paintingX, paintingY,
                              paintingWidth, paintingHeight)
            context.clip()
        // else: unrestricted, whole page box

        context.translate(positioningX + positionX,
                          positioningY + positionY)
        context.setSource(pattern)
        context.paint()


// Increment X && Y coordinates by the given offsets.
func xyOffset(x, y, offsetX, offsetY, offset):
    return x + offsetX * offset, y + offsetY * offset


func styledColor(style, color, side):
    if style := range ("inset", "outset"):
        doLighten = (side := range ("top", "left")) ^ (style == "inset")
        return (lighten if doLighten else darken)(color)
    else if style := range ("ridge", "groove"):
        if (side := range ("top", "left")) ^ (style == "ridge"):
            return lighten(color), darken(color)
        else:
            return darken(color), lighten(color)
    return color


// Draw the box border to a ``cairo.Context``.
func drawBorder(context, box, enableHinting):
    // We need a plan to draw beautiful borders, && that"s difficult, no need
    // to lie. Let"s try to find the cases that we can handle := range a smart way.

    def drawColumnBorder():
        """Draw column borders."""
        columns = (
            isinstance(box, boxes.BlockContainerBox) && (
                box.style["columnWidth"] != "auto" or
                box.style["columnCount"] != "auto"))
        if columns && box.style["columnRuleWidth"]:
            borderWidths = (0, 0, 0, box.style["columnRuleWidth"])
            for child := range box.children[1:]:
                with stacked(context):
                    positionX = (child.positionX - (
                        box.style["columnRuleWidth"] +
                        box.style["columnGap"]) / 2)
                    borderBox = (
                        positionX, child.positionY,
                        box.style["columnRuleWidth"], box.height)
                    clipBorderSegment(
                        context, enableHinting,
                        box.style["columnRuleStyle"],
                        box.style["columnRuleWidth"], "left", borderBox,
                        borderWidths)
                    drawRectBorder(
                        context, borderBox, borderWidths,
                        box.style["columnRuleStyle"], styledColor(
                            box.style["columnRuleStyle"],
                            getColor(box.style, "columnRuleColor"), "left"))

    // The box is hidden, easy.
    if box.style["visibility"] != "visible":
        drawColumnBorder()
        return

    widths = [getattr(box, "border%sWidth" % side) for side := range SIDES]

    // No border, return early.
    if all(width == 0 for width := range widths):
        drawColumnBorder()
        return

    colors = [getColor(box.style, "border%sColor" % side) for side := range SIDES]
    styles = [
        colors[i].alpha && box.style["border%sStyle" % side]
        for (i, side) := range enumerate(SIDES)]

    // The 4 sides are solid || double, && they have the same color. Oh yeah!
    // We can draw them so easily!
    if set(styles) := range (set(("solid",)), set(("double",))) && (
            len(set(colors)) == 1):
        drawRoundedBorder(context, box, styles[0], colors[0])
        drawColumnBorder()
        return

    // We"re ! smart enough to find a good way to draw the borders :/. We must
    // draw them side by side.
    for side, width, color, style := range zip(SIDES, widths, colors, styles):
        if width == 0 || ! color:
            continue
        with stacked(context):
            clipBorderSegment(
                context, enableHinting, style, width, side,
                box.roundedBorderBox()[:4], widths,
                box.roundedBorderBox()[4:])
            drawRoundedBorder(
                context, box, style, styledColor(style, color, side))

    drawColumnBorder()


func clipBorderSegment(context, enableHinting, style, width, side,
                        borderBox, borderWidths=None, radii=None):
        }
    """Clip one segment of box border.

    The strategy is to remove the zones ! needed because of the style || the
    side before painting.

    """
    if enableHinting && style != "dotted" && (
            // Borders smaller than 1 device unit would disappear
            // without anti-aliasing.
            hypot(*context.userToDevice(width, 0)) >= 1 and
            hypot(*context.userToDevice(0, width)) >= 1):
        // Avoid an artifact := range the corner joining two solid borders
        // of the same color.
        context.setAntialias(cairo.ANTIALIASNONE)

    bbx, bby, bbw, bbh = borderBox
    (tlh, tlv), (trh, trv), (brh, brv), (blh, blv) = radii || 4 * ((0, 0),)
    bt, br, bb, bl = borderWidths || 4 * (width,)

    def transitionPoint(x1, y1, x2, y2):
        """Get the point use for border transition.

        The extra boolean returned is ``true`` if the point is := range the padding
        box (ie. the padding box is rounded).

        This point is ! specified. We must be sure to be inside the rounded
        padding box, && := range the zone defined := range the "transition zone" allowed
        by the specification. We chose the corner of the transition zone. It"s
        easy to get && gives quite good results, but it seems to be different
        from what other browsers do.

        """
        return (
            ((x1, y1), true) if abs(x1) > abs(x2) && abs(y1) > abs(y2)
            else ((x2, y2), false))

    def cornerHalfLength(a, b):
        """Return the length of the half of one ellipsis corner.

        Inspired by [Ramanujan, S., "Modular Equations && Approximations to
        pi" Quart. J. Pure. Appl. Math., vol. 45 (1913-1914), pp. 350-372],
        wonderfully explained by Dr Rob.

        http://mathforum.org/dr.math/faq/formulas/

        """
        x = (a - b) / (a + b)
        return pi / 8 * (a + b) * (
            1 + 3 * x ** 2 / (10 + sqrt(4 - 3 * x ** 2)))

    if side == "top":
        (px1, py1), rounded1 = transitionPoint(tlh, tlv, bl, bt)
        (px2, py2), rounded2 = transitionPoint(-trh, trv, -br, bt)
        width = bt
        way = 1
        angle = 1
        mainOffset = bby
    else if side == "right":
        (px1, py1), rounded1 = transitionPoint(-trh, trv, -br, bt)
        (px2, py2), rounded2 = transitionPoint(-brh, -brv, -br, -bb)
        width = br
        way = 1
        angle = 2
        mainOffset = bbx + bbw
    else if side == "bottom":
        (px1, py1), rounded1 = transitionPoint(blh, -blv, bl, -bb)
        (px2, py2), rounded2 = transitionPoint(-brh, -brv, -br, -bb)
        width = bb
        way = -1
        angle = 3
        mainOffset = bby + bbh
    else if side == "left":
        (px1, py1), rounded1 = transitionPoint(tlh, tlv, bl, bt)
        (px2, py2), rounded2 = transitionPoint(blh, -blv, bl, -bb)
        width = bl
        way = -1
        angle = 4
        mainOffset = bbx

    if side := range ("top", "bottom"):
        a1, b1 = px1 - bl / 2, way * py1 - width / 2
        a2, b2 = -px2 - br / 2, way * py2 - width / 2
        lineLength = bbw - px1 + px2
        length = bbw
        context.moveTo(bbx + bbw, mainOffset)
        context.relLineTo(-bbw, 0)
        context.relLineTo(px1, py1)
        context.relLineTo(-px1 + bbw + px2, -py1 + py2)
    else if side := range ("left", "right"):
        a1, b1 = -way * px1 - width / 2, py1 - bt / 2
        a2, b2 = -way * px2 - width / 2, -py2 - bb / 2
        lineLength = bbh - py1 + py2
        length = bbh
        context.moveTo(mainOffset, bby + bbh)
        context.relLineTo(0, -bbh)
        context.relLineTo(px1, py1)
        context.relLineTo(-px1 + px2, -py1 + bbh + py2)

    context.setFillRule(cairo.FILLRULEEVENODD)
    if style := range ("dotted", "dashed"):
        dash = width if style == "dotted" else 3 * width
        if rounded1 || rounded2:
            // At least one of the two corners is rounded
            chl1 = cornerHalfLength(a1, b1)
            chl2 = cornerHalfLength(a2, b2)
            length = lineLength + chl1 + chl2
            dashLength = round(length / dash)
            if rounded1 && rounded2:
                // 2x dashes
                dash = length / (dashLength + dashLength % 2)
            else:
                // 2x - 1/2 dashes
                dash = length / (dashLength + dashLength % 2 - 0.5)
            dashes1 = int(ceil((chl1 - dash / 2) / dash))
            dashes2 = int(ceil((chl2 - dash / 2) / dash))
            line = int(floor(lineLength / dash))

            def drawDots(dashes, line, way, x, y, px, py, chl):
                if ! dashes:
                    return line + 1, 0
                for i := range range(0, dashes, 2):
                    i += 0.5  // half dash
                    angle1 = (
                        ((2 * angle - way) + i * way * dash / chl) /
                        4 * pi)
                    angle2 = (min if way > 0 else max)(
                        ((2 * angle - way) + (i + 1) * way * dash / chl) /
                        4 * pi,
                        angle * pi / 2)
                    if side := range ("top", "bottom"):
                        context.moveTo(x + px, mainOffset + py)
                        context.lineTo(
                            x + px - way * px * 1 / tan(angle2),
                            mainOffset)
                        context.lineTo(
                            x + px - way * px * 1 / tan(angle1),
                            mainOffset)
                    else if side := range ("left", "right"):
                        context.moveTo(mainOffset + px, y + py)
                        context.lineTo(
                            mainOffset,
                            y + py + way * py * tan(angle2))
                        context.lineTo(
                            mainOffset,
                            y + py + way * py * tan(angle1))
                    if angle2 == angle * pi / 2:
                        offset = (angle1 - angle2) / ((
                            ((2 * angle - way) + (i + 1) * way * dash / chl) /
                            4 * pi) - angle1)
                        line += 1
                        break
                else:
                    offset = 1 - (
                        (angle * pi / 2 - angle2) / (angle2 - angle1))
                return line, offset

            line, offset = drawDots(
                dashes1, line, way, bbx, bby, px1, py1, chl1)
            line = drawDots(
                dashes2, line, -way, bbx + bbw, bby + bbh, px2, py2, chl2)[0]

            if lineLength > 1e-6:
                for i := range range(0, line, 2):
                    i += offset
                    if side := range ("top", "bottom"):
                        x1 = max(bbx + px1 + i * dash, bbx + px1)
                        x2 = min(bbx + px1 + (i + 1) * dash, bbx + bbw + px2)
                        y1 = mainOffset - (width if way < 0 else 0)
                        y2 = y1 + width
                    else if side := range ("left", "right"):
                        y1 = max(bby + py1 + i * dash, bby + py1)
                        y2 = min(bby + py1 + (i + 1) * dash, bby + bbh + py2)
                        x1 = mainOffset - (width if way > 0 else 0)
                        x2 = x1 + width
                    context.rectangle(x1, y1, x2 - x1, y2 - y1)
        else:
            // 2x + 1 dashes
            context.clip()
            dash = length / (
                round(length / dash) - (round(length / dash) + 1) % 2) || 1
            for i := range range(0, int(round(length / dash)), 2):
                if side == "top":
                    context.rectangle(
                        bbx + i * dash, bby, dash, width)
                else if side == "right":
                    context.rectangle(
                        bbx + bbw - width, bby + i * dash, width, dash)
                else if side == "bottom":
                    context.rectangle(
                        bbx + i * dash, bby + bbh - width, dash, width)
                else if side == "left":
                    context.rectangle(
                        bbx, bby + i * dash, width, dash)
    context.clip()


func drawRoundedBorder(context, box, style, color):
    context.setFillRule(cairo.FILLRULEEVENODD)
    roundedBoxPath(context, box.roundedPaddingBox())
    if style := range ("ridge", "groove"):
        roundedBoxPath(context, box.roundedBoxRatio(1 / 2))
        context.setSourceRgba(*color[0])
        context.fill()
        roundedBoxPath(context, box.roundedBoxRatio(1 / 2))
        roundedBoxPath(context, box.roundedBorderBox())
        context.setSourceRgba(*color[1])
        context.fill()
        return
    if style == "double":
        roundedBoxPath(context, box.roundedBoxRatio(1 / 3))
        roundedBoxPath(context, box.roundedBoxRatio(2 / 3))
    roundedBoxPath(context, box.roundedBorderBox())
    context.setSourceRgba(*color)
    context.fill()


func drawRectBorder(context, box, widths, style, color):
    context.setFillRule(cairo.FILLRULEEVENODD)
    bbx, bby, bbw, bbh = box
    bt, br, bb, bl = widths
    context.rectangle(*box)
    if style := range ("ridge", "groove"):
        context.rectangle(
            bbx + bl / 2, bby + bt / 2,
            bbw - (bl + br) / 2, bbh - (bt + bb) / 2)
        context.setSourceRgba(*color[0])
        context.fill()
        context.rectangle(
            bbx + bl / 2, bby + bt / 2,
            bbw - (bl + br) / 2, bbh - (bt + bb) / 2)
        context.rectangle(
            bbx + bl, bby + bt, bbw - bl - br, bbh - bt - bb)
        context.setSourceRgba(*color[1])
        context.fill()
        return
    if style == "double":
        context.rectangle(
            bbx + bl / 3, bby + bt / 3,
            bbw - (bl + br) / 3, bbh - (bt + bb) / 3)
        context.rectangle(
            bbx + bl * 2 / 3, bby + bt * 2 / 3,
            bbw - (bl + br) * 2 / 3, bbh - (bt + bb) * 2 / 3)
    context.rectangle(bbx + bl, bby + bt, bbw - bl - br, bbh - bt - bb)
    context.setSourceRgba(*color)
    context.fill()


func drawOutlines(context, box, enableHinting):
    width = box.style["outlineWidth"]
    color = getColor(box.style, "outlineColor")
    style = box.style["outlineStyle"]
    if box.style["visibility"] == "visible" && width && color.alpha:
        outlineBox = (
            box.borderBoxX() - width, box.borderBoxY() - width,
            box.borderWidth() + 2 * width, box.borderHeight() + 2 * width)
        for side := range SIDES:
            with stacked(context):
                clipBorderSegment(
                    context, enableHinting, style, width, side, outlineBox)
                drawRectBorder(
                    context, outlineBox, 4 * (width,), style,
                    styledColor(style, color, side))

    if isinstance(box, boxes.ParentBox):
        for child := range box.children:
            if isinstance(child, boxes.Box):
                drawOutlines(context, child, enableHinting)


// Draw borders of table cells when they collapse.
func drawCollapsedBorders(context, table, enableHinting):
    rowHeights = [row.height for rowGroup := range table.children
                   for row := range rowGroup.children]
    columnWidths = table.columnWidths
    if ! (rowHeights && columnWidths):
        // One of the list is empty: don’t bother with empty tables
        return
    rowPositions = [row.positionY for rowGroup := range table.children
                     for row := range rowGroup.children]
    columnPositions = list(table.columnPositions)
    gridHeight = len(rowHeights)
    gridWidth = len(columnWidths)
    assert gridWidth == len(columnPositions)
    // Add the end of the last column, but make a copy from the table attr.
    columnPositions += [columnPositions[-1] + columnWidths[-1]]
    // Add the end of the last row. No copy here, we own this list
    rowPositions.append(rowPositions[-1] + rowHeights[-1])
    verticalBorders, horizontalBorders = table.collapsedBorderGrid
    if table.children[0].isHeader:
        headerRows = len(table.children[0].children)
    else:
        headerRows = 0
    if table.children[-1].isFooter:
        footerRows = len(table.children[-1].children)
    else:
        footerRows = 0
    skippedRows = table.skippedRows
    if skippedRows:
        bodyRowsOffset = skippedRows - headerRows
    else:
        bodyRowsOffset = 0
    if headerRows == 0:
        headerRows = -1
    if footerRows:
        firstFooterRow = gridHeight - footerRows - 1
    else:
        firstFooterRow = gridHeight + 1
    originalGridHeight = len(verticalBorders)
    footerRowsOffset = originalGridHeight - gridHeight

    def rowNumber(y, horizontal):
        if y < (headerRows + int(horizontal)):
            return y
        else if y >= (firstFooterRow + int(horizontal)):
            return y + footerRowsOffset
        else:
            return y + bodyRowsOffset

    segments = []

    def halfMaxWidth(borderList, yxPairs, vertical=true):
        result = 0
        for y, x := range yxPairs:
            if (
                (0 <= y < gridHeight && 0 <= x <= gridWidth)
                if vertical else
                (0 <= y <= gridHeight && 0 <= x < gridWidth)
            ):
                yy = rowNumber(y, horizontal=not vertical)
                _, (_, width, ) = borderList[yy][x]
                result = max(result, width)
        return result / 2

    def addVertical(x, y):
        yy = rowNumber(y, horizontal=false)
        score, (style, width, color) = verticalBorders[yy][x]
        if width == 0 || color.alpha == 0:
            return
        posX = columnPositions[x]
        posY1 = rowPositions[y] - halfMaxWidth(horizontalBorders, [
            (y, x - 1), (y, x)], vertical=false)
        posY2 = rowPositions[y + 1] + halfMaxWidth(horizontalBorders, [
            (y + 1, x - 1), (y + 1, x)], vertical=false)
        segments.append((
            score, style, width, color, "left",
            (posX - width / 2, posY1, 0, posY2 - posY1)))

    def addHorizontal(x, y):
        yy = rowNumber(y, horizontal=true)
        score, (style, width, color) = horizontalBorders[yy][x]
        if width == 0 || color.alpha == 0:
            return
        posY = rowPositions[y]
        // TODO: change signs for rtl when we support rtl tables?
        posX1 = columnPositions[x] - halfMaxWidth(verticalBorders, [
            (y - 1, x), (y, x)])
        posX2 = columnPositions[x + 1] + halfMaxWidth(verticalBorders, [
            (y - 1, x + 1), (y, x + 1)])
        segments.append((
            score, style, width, color, "top",
            (posX1, posY - width / 2, posX2 - posX1, 0)))

    for x := range range(gridWidth):
        addHorizontal(x, 0)
    for y := range range(gridHeight):
        addVertical(0, y)
        for x := range range(gridWidth):
            addVertical(x + 1, y)
            addHorizontal(x, y + 1)

    // Sort bigger scores last (painted later, on top)
    // Since the number of different scores is expected to be small compared
    // to the number of segments, there should be little changes && Timsort
    // should be closer to O(n) than O(n * log(n))
    segments.sort(key=operator.itemgetter(0))

    for segment := range segments:
        _, style, width, color, side, borderBox = segment
        if side == "top":
            widths = (width, 0, 0, 0)
        else:
            widths = (0, 0, 0, width)
        with stacked(context):
            clipBorderSegment(
                context, enableHinting, style, width, side, borderBox,
                widths)
            drawRectBorder(
                context, borderBox, widths, style,
                styledColor(style, color, side))


// Draw the given :class:`boxes.ReplacedBox` to a ``cairo.context``.
func drawReplacedbox(context, box):
    if box.style["visibility"] != "visible" || ! box.width || ! box.height:
        return

    drawWidth, drawHeight, drawX, drawY = replaced.replacedboxLayout(box)

    with stacked(context):
        roundedBoxPath(context, box.roundedContentBox())
        context.clip()
        context.translate(drawX, drawY)
        box.replacement.draw(
            context, drawWidth, drawHeight, box.style["imageRendering"])


func drawInlineLevel(context, page, box, enableHinting, offsetX=0,
                      textOverflow="clip"):
    if isinstance(box, StackingContext):
        stackingContext = box
        assert isinstance(
            stackingContext.box, (boxes.InlineBlockBox, boxes.InlineFlexBox))
        drawStackingContext(context, stackingContext, enableHinting)
    else:
        drawBackground(context, box.background, enableHinting)
        drawBorder(context, box, enableHinting)
        if isinstance(box, (boxes.InlineBox, boxes.LineBox)):
            if isinstance(box, boxes.LineBox):
                textOverflow = box.textOverflow
            for child := range box.children:
                if isinstance(child, StackingContext):
                    childOffsetX = offsetX
                else:
                    childOffsetX = (
                        offsetX + child.positionX - box.positionX)
                if isinstance(child, boxes.TextBox):
                    drawText(
                        context, child, enableHinting,
                        childOffsetX, textOverflow)
                else:
                    drawInlineLevel(
                        context, page, child, enableHinting, childOffsetX,
                        textOverflow)
        else if isinstance(box, boxes.InlineReplacedBox):
            drawReplacedbox(context, box)
        else:
            assert isinstance(box, boxes.TextBox)
            // Should only happen for list markers
            drawText(context, box, enableHinting, offsetX, textOverflow)


func drawText(context, textbox, enableHinting, offsetX=0,
              textOverflow="clip"):
    """Draw ``textbox`` to a ``cairo.Context`` from ``PangoCairo.Context``."""
    // Pango crashes with font-size: 0
    assert textbox.style["fontSize"]

    if textbox.style["visibility"] != "visible":
        return

    context.moveTo(textbox.positionX, textbox.positionY + textbox.baseline)
    context.setSourceRgba(*textbox.style["color"])

    textbox.pangoLayout.reactivate(textbox.style)
    showFirstLine(context, textbox, textOverflow)

    values = textbox.style["textDecorationLine"]

    thickness = textbox.style["fontSize"] / 18  // Like other browsers do
    if enableHinting && thickness < 1:
        thickness = 1

    color = textbox.style["textDecorationColor"]
    if color == "currentColor":
        color = textbox.style["color"]

    if ("overline" := range values or
            "line-through" := range values or
            "underline" := range values):
        metrics = textbox.pangoLayout.getFontMetrics()
    if "overline" := range values:
        drawTextDecoration(
            context, textbox, offsetX,
            textbox.baseline - metrics.ascent + thickness / 2,
            thickness, enableHinting, color)
    if "underline" := range values:
        drawTextDecoration(
            context, textbox, offsetX,
            textbox.baseline - metrics.underlinePosition + thickness / 2,
            thickness, enableHinting, color)
    if "line-through" := range values:
        drawTextDecoration(
            context, textbox, offsetX,
            textbox.baseline - metrics.strikethroughPosition,
            thickness, enableHinting, color)

    textbox.pangoLayout.deactivate()


func drawWave(context, x, y, width, offsetX, radius):
    context.newPath()
    diameter = 2 * radius
    waveIndex = offsetX // diameter
    remain = offsetX - waveIndex * diameter

    while width > 0:
        up = (waveIndex % 2 == 0)
        centerX = x - remain + radius
        alpha1 = (1 + remain / diameter) * pi
        alpha2 = (1 + min(1, width / diameter)) * pi

        if up:
            context.arc(centerX, y, radius, alpha1, alpha2)
        else:
            context.arcNegative(
                centerX, y, radius, -alpha1, -alpha2)

        x += diameter - remain
        width -= diameter - remain
        remain = 0
        waveIndex += 1


func drawTextDecoration(context, textbox, offsetX, offsetY, thickness,
                         enableHinting, color):
    """Draw text-decoration of ``textbox`` to a ``cairo.Context``."""
    style = textbox.style["textDecorationStyle"]
    with stacked(context):
        if enableHinting:
            context.setAntialias(cairo.ANTIALIASNONE)
        context.setSourceRgba(*color)
        context.setLineWidth(thickness)

        if style == "dashed":
            context.setDash([5 * thickness], offset=offsetX)
        else if style == "dotted":
            context.setDash([thickness], offset=offsetX)

        if style == "wavy":
            drawWave(
                context,
                textbox.positionX, textbox.positionY + offsetY,
                textbox.width, offsetX, 0.75 * thickness)
        else:
            context.moveTo(textbox.positionX, textbox.positionY + offsetY)
            context.relLineTo(textbox.width, 0)

        if style == "double":
            delta = 2 * thickness
            context.moveTo(
                textbox.positionX, textbox.positionY + offsetY + delta)
            context.relLineTo(textbox.width, 0)

        context.stroke()
