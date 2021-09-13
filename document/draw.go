package document

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"text/template"

	"github.com/benoitkugler/go-weasyprint/backend"
	"github.com/benoitkugler/go-weasyprint/style/parser"

	"github.com/benoitkugler/go-weasyprint/layout"

	"github.com/benoitkugler/go-weasyprint/images"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/utils"

	bo "github.com/benoitkugler/go-weasyprint/boxes"
)

// Take an "after layout" box tree and draw it onto a cairo context.

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

type Bleed struct {
	Top, Bottom, Left, Right fl
}

// FIXME: check gofpdf support for SVG
type svgArgs struct {
	Width, Height    fl
	Bleed, HalfBleed Bleed
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
func hsv2rgb(hue, saturation, value fl) (r, g, b fl) {
	c := value * saturation
	x := c * fl(1-math.Abs(utils.FloatModulo(float64(hue)/60, 2)-1))
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
func rgb2hsv(red, green, blue fl) (h, s, c fl) {
	cmax := utils.Maxs(red, green, blue)
	cmin := utils.Mins(red, green, blue)
	delta := cmax - cmin
	var hue fl
	if delta == 0 {
		hue = 0
	} else if cmax == red {
		hue = 60 * utils.FloatModulo(float64((green-blue)/delta), 6)
	} else if cmax == green {
		hue = 60 * ((blue-red)/delta + 2)
	} else if cmax == blue {
		hue = 60 * ((red-green)/delta + 4)
	}
	var saturation fl
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
func drawPage(page *bo.PageBox, context Drawer, enableHinting bool) error {
	bleed := Bleed{
		Top:    fl(page.Style.GetBleedTop().Value),
		Bottom: fl(page.Style.GetBleedBottom().Value),
		Left:   fl(page.Style.GetBleedLeft().Value),
		Right:  fl(page.Style.GetBleedRight().Value),
	}
	marks := page.Style.GetMarks()
	stackingContext := NewStackingContextFromPage(page)
	if err := drawBackground(context, stackingContext.box.Box().Background, enableHinting, false, bleed, marks); err != nil {
		return err
	}
	if err := drawBackground(context, page.CanvasBackground, enableHinting, false, Bleed{}, pr.Marks{}); err != nil {
		return err
	}
	drawBorder(context, page, enableHinting)
	return drawStackingContext(context, stackingContext, enableHinting)
}

func drawBoxBackgroundAndBorder(context Drawer, page *bo.PageBox, box Box, enableHinting bool) error {
	if err := drawBackground(context, box.Box().Background, enableHinting, true, Bleed{}, pr.Marks{}); err != nil {
		return err
	}
	if box_, ok := box.(bo.InstanceTableBox); ok {
		box := box_.Table()
		if err := drawTableBackgrounds(context, page, box_, enableHinting); err != nil {
			return err
		}
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
	return nil
}

// Draw a ``stackingContext`` on ``context``.
func drawStackingContext(context Drawer, stackingContext StackingContext, enableHinting bool) error {
	// See http://www.w3.org/TR/CSS2/zindex.html
	return context.OnNewStack(func() error {
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
				fl(box.BorderBoxX()+right.Value),
				fl(box.BorderBoxY()+top.Value),
				fl(left.Value-right.Value),
				fl(bottom.Value-top.Value),
			)
			context.Clip()
		}

		ops := func() error {
			if box.TransformationMatrix != nil {
				if err := box.TransformationMatrix.Copy().Invert(); err != nil {
					return err
				}
				context.Transform(*box.TransformationMatrix)
			}

			// Point 1 is done in drawPage

			// Point 2
			if bo.TypeBlockBox.IsInstance(box_) || bo.IsMarginBox(box_) ||
				bo.TypeInlineBlockBox.IsInstance(box_) || bo.TypeTableCellBox.IsInstance(box_) ||
				bo.IsFlexContainerBox(box_) {
				// The canvas background was removed by setCanvasBackground
				if err := drawBoxBackgroundAndBorder(context, stackingContext.page, box_, enableHinting); err != nil {
					return err
				}
			}
			return context.OnNewStack(func() error {
				if box.Style.GetOverflow() != "visible" {
					// Only clip the content and the children:
					// - the background is already clipped
					// - the border must *not* be clipped
					roundedBoxPath(context, box.RoundedPaddingBox())
					context.Clip()
				}

				// Point 3
				for _, childContext := range stackingContext.negativeZContexts {
					if err := drawStackingContext(context, childContext, enableHinting); err != nil {
						return err
					}
				}

				// Point 4
				for _, block := range stackingContext.blockLevelBoxes {
					if err := drawBoxBackgroundAndBorder(context, stackingContext.page, block, enableHinting); err != nil {
						return err
					}
				}

				// Point 5
				for _, childContext := range stackingContext.floatContexts {
					if err := drawStackingContext(context, childContext, enableHinting); err != nil {
						return err
					}
				}

				// Point 6
				if bo.TypeInlineBox.IsInstance(box_) {
					if err := drawInlineLevel(context, stackingContext.page, box_, enableHinting, 0, "clip"); err != nil {
						return err
					}
				}

				// Point 7
				for _, block := range append([]Box{box_}, stackingContext.blocksAndCells...) {
					if block, ok := block.(bo.InstanceReplacedBox); ok {
						drawReplacedbox(context, block)
					} else {
						for _, child := range block.Box().Children {
							if bo.TypeLineBox.IsInstance(child) {
								if err := drawInlineLevel(context, stackingContext.page, child, enableHinting, 0, "clip"); err != nil {
									return err
								}
							}
						}
					}
				}

				// Point 8
				for _, childContext := range stackingContext.zeroZContexts {
					if err := drawStackingContext(context, childContext, enableHinting); err != nil {
						return err
					}
				}

				// Point 9
				for _, childContext := range stackingContext.positiveZContexts {
					if err := drawStackingContext(context, childContext, enableHinting); err != nil {
						return err
					}
				}
				return nil
			})
			// Point 10
			drawOutlines(context, box_, enableHinting)
			return nil
		}

		opacity := fl(box.Style.GetOpacity())
		if opacity < 1 {
			return context.OnNewStack(func() error {
				context.SetAlpha(opacity)
				return ops()
			})
		} else {
			return ops()
		}
	})
}

// Draw the path of the border radius box.
// ``widths`` is a tuple of the inner widths (top, right, bottom, left) from
// the border box. Radii are adjusted from these values. Default is (0, 0, 0,
// 0).
// Inspired by cairo cookbook
// http://cairographics.org/cookbook/roundedrectangles/
//
func roundedBoxPath(context Drawer, radii bo.RoundedBox) {
	x, y, w, h, tls, trs, brs, bls := float64(radii.X), float64(radii.Y), radii.Width, radii.Height, radii.TopLeft, radii.TopRight, radii.BottomRight, radii.BottomLeft
	// Note: No support for elliptic radius
	tl := float64(pr.Max(tls[0], tls[1]))
	tr := float64(pr.Max(trs[0], trs[1]))
	br := float64(pr.Max(brs[0], brs[1]))
	bl := float64(pr.Max(bls[0], bls[1]))
	context.RoundedRect(x, y, float64(w), float64(h), tl, tr, br, bl)
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
	return drawBackground(context, bg, enableHinting, true, Bleed{}, pr.Marks{})
}

// Draw the background color and image to a ``cairo.Context``
// If ``clipBox`` is set to ``false``, the background is not clipped to the
// border box of the background, but only to the painting area
// clipBox=true bleed=None marks=()
func drawBackground(context Drawer, bg *bo.Background, enableHinting, clipBox bool, bleed Bleed,
	marks pr.Marks) error {
	if bg == nil {
		return nil
	}

	return context.OnNewStack(func() error {
		if enableHinting {
			// Prefer crisp edges on background rectangles.
			context.SetAntialias(backend.AntialiasNone)
		}

		if clipBox {
			for _, box := range bg.Layers[len(bg.Layers)-1].ClippedBoxes {
				roundedBoxPath(context, box)
			}
			context.Clip()
		}

		// Background color
		if bg.Color.A > 0 {
			context.OnNewStack(func() error {
				paintingArea := bg.Layers[len(bg.Layers)-1].PaintingArea.Rect
				if !paintingArea.IsNone() {
					ptx, pty, ptw, pth := paintingArea.Unpack()
					if (bleed != Bleed{}) {
						// Painting area is the PDF BleedBox
						ptx -= bleed.Left
						pty -= bleed.Top
						ptw += bleed.Left + bleed.Right
						pth += bleed.Top + bleed.Bottom
					}
					context.Rectangle(ptx, pty, ptw, pth)
					context.Clip()
				}
				context.SetSourceRgba(bg.Color.Unpack())
				context.Paint()
				return nil
			}) // can't error
		}

		if (bleed != Bleed{}) && !marks.IsNone() {
			x, y, width, height := bg.Layers[len(bg.Layers)-1].PaintingArea.Rect.Unpack()
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
			halfBleed := Bleed{
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
			size := pr.Size{Width: pr.FToV(width), Height: pr.FToV(height)}
			position := bo.Position{Point: bo.MaybePoint{pr.Float(x), pr.Float(y)}}
			repeat := bo.Repeat{Reps: [2]string{"no-repeat", "no-repeat"}}
			unbounded := true
			paintingArea := bo.Area{Rect: pr.Rectangle{pr.Float(x), pr.Float(y), pr.Float(width), pr.Float(height)}}
			positioningArea := bo.Area{Rect: pr.Rectangle{0, 0, pr.Float(width), pr.Float(height)}}
			layer := bo.BackgroundLayer{
				Image: image, Size: size, Position: position, Repeat: repeat, Unbounded: unbounded,
				PaintingArea: paintingArea, PositioningArea: positioningArea,
			}
			bg.Layers = append([]bo.BackgroundLayer{layer}, bg.Layers...)
		}
		// Paint in reversed order: first layer is "closest" to the viewer.
		for _, layer := range reversed(bg.Layers) {
			drawBackgroundImage(context, layer, bg.ImageRendering)
		}
		return nil
	})
}

// Draw the background color && image of the table children.
func drawTableBackgrounds(context Drawer, page *bo.PageBox, table_ bo.InstanceTableBox, enableHinting bool) error {
	table := table_.Table()
	for _, columnGroup := range table.ColumnGroups {
		err := drawBackground2(context, columnGroup.Box().Background, enableHinting)
		if err != nil {
			return err
		}
		for _, column := range columnGroup.Box().Children {
			err = drawBackground2(context, column.Box().Background, enableHinting)
			if err != nil {
				return err
			}
		}
	}
	for _, rowGroup := range table.Children {
		err := drawBackground2(context, rowGroup.Box().Background, enableHinting)
		if err != nil {
			return err
		}
		for _, row := range rowGroup.Box().Children {
			err = drawBackground2(context, row.Box().Background, enableHinting)
			if err != nil {
				return err
			}
			for _, cell := range row.Box().Children {
				cell := cell.Box()
				if table.Style.GetBorderCollapse() == "collapse" ||
					cell.Style.GetEmptyCells() == "show" || !cell.Empty {
					err = drawBackground2(context, cell.Background, enableHinting)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
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
		repeatWidth = math.Max(imageWidth, float64(paintingWidth))
	case "repeat", "round":
		// We repeat the image each imageWidth.
		repeatWidth = imageWidth
	case "space":
		nRepeats := math.Floor(float64(positioningWidth) / imageWidth)
		if nRepeats >= 2 {
			// The repeat width is the whole positioning width with one image
			// removed, divided by (the number of repeated images - 1). This
			// way, we get the width of one image + one space. We ignore
			// background-position for this dimension.
			repeatWidth = (float64(positioningWidth) - imageWidth) / (nRepeats - 1)
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

	pageWidth, pageHeight := context.GetPageSize()
	nRepeatX, nRepeatY := int(math.Ceil(pageWidth/repeatWidth)), int(math.Ceil(pageHeight/repeatHeight))

	context.OnNewStack(func() error {
		if !layer.Unbounded {
			context.Rectangle(float64(paintingX), float64(paintingY),
				paintingWidth, paintingHeight)
			context.Clip()
		} // else: unrestricted, whole page box

		context.Translate(positioningX+float64(positionX.V()),
			positioningY+float64(positionY.V()))
		for i := 0; i < nRepeatX; i += 1 {
			context.OnNewStack(func() error {
				for j := 0; j < nRepeatY; j += 1 {
					layer.Image.Draw(context, imageWidth, imageHeight, imageRendering)
					context.Translate(0, repeatHeight)
				}
				return nil
			})
			context.Translate(repeatWidth, 0)
		}
		return nil
	})
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
				context.OnNewStack(func() error {
					positionX := child.Box().PositionX - (crw.Value+
						box.Style.GetColumnGap().Value)/2
					borderBox := pr.Rectangle{
						positionX, child.Box().PositionY,
						crw.Value, box.Height.V(),
					}
					clipBorderSegment(context, enableHinting,
						box.Style.GetColumnRuleStyle(),
						float64(crw.Value), "left", borderBox,
						&borderWidths, nil)
					drawRectBorder(context, borderBox, borderWidths,
						box.Style.GetColumnRuleStyle(), styledColor(
							box.Style.GetColumnRuleStyle(),
							box.Style.ResolveColor("column_rule_color").RGBA, "left"))
					return nil
				})
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
		stylesSet = utils.NewSet()
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
		drawRoundedBorder(context, *box, styles[0], []Color{colors[0]})
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
		context.OnNewStack(func() error {
			rb := box.RoundedBorderBox()
			roundedBox := pr.Rectangle{rb.X, rb.Y, rb.Width, rb.Height}
			radii := [4]bo.Point{rb.TopLeft, rb.TopRight, rb.BottomRight, rb.BottomLeft}
			clipBorderSegment(context, enableHinting, style, float64(width), side,
				roundedBox, &widths, &radii)
			drawRoundedBorder(context, *box, style, styledColor(style, color, side))
			return nil
		})
	}

	drawColumnBorder()
}

// Clip one segment of box border (border_widths=None, radii=None).
// The strategy is to remove the zones not needed because of the style or the
// side before painting.
func clipBorderSegment(context backend.Drawer, enableHinting bool, style pr.String, width float64, side string,
	borderBox pr.Rectangle, borderWidths *pr.Rectangle, radii *[4]bo.Point) {

	// if enableHinting && style != "dotted" && (
	// // Borders smaller than 1 device unit would disappear
	// // without anti-aliasing.
	// math.Hypot(context.UserToDevice(float64(width), 0)) >= 1 &&
	// 	math.Hypot(context.UserToDevice(0, float64(width))) >= 1) {
	// 	// Avoid an artifact in the corner joining two solid borders
	// 	// of the same color.
	// 	context.SetAntialias(backend.AntialiasNone)
	// }

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

	context.SetFillRule(backend.FillRuleEvenOdd)
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
	context.SetFillRule(backend.FillRuleEvenOdd)
	roundedBoxPath(context, box.RoundedPaddingBox())
	if style == "ridge" || style == "groove" {
		roundedBoxPath(context, box.RoundedBoxRatio(1/2))
		context.SetSourceRgba(color[0].Unpack())
		context.Fill()
		roundedBoxPath(context, box.RoundedBoxRatio(1/2))
		roundedBoxPath(context, box.RoundedBorderBox())
		context.SetSourceRgba(color[1].Unpack())
		context.Fill()
		return
	}
	if style == "double" {
		roundedBoxPath(context, box.RoundedBoxRatio(1/3))
		roundedBoxPath(context, box.RoundedBoxRatio(2/3))
	}
	roundedBoxPath(context, box.RoundedBorderBox())
	context.SetSourceRgba(color[0].Unpack())
	context.Fill()
}

func drawRectBorder(context Drawer, box, widths pr.Rectangle, style pr.String, color []Color) {
	context.SetFillRule(backend.FillRuleEvenOdd)
	bbx, bby, bbw, bbh := box.Unpack()
	bt, br, bb, bl := widths.Unpack()
	context.Rectangle(box.Unpack())
	if style == "ridge" || style == "groove" {
		context.Rectangle(bbx+bl/2, bby+bt/2, bbw-(bl+br)/2, bbh-(bt+bb)/2)
		context.SetSourceRgba(color[0].Unpack())
		context.Fill()
		context.Rectangle(bbx+bl/2, bby+bt/2, bbw-(bl+br)/2, bbh-(bt+bb)/2)
		context.Rectangle(bbx+bl, bby+bt, bbw-bl-br, bbh-bt-bb)
		context.SetSourceRgba(color[1].Unpack())
		context.Fill()
		return
	}
	if style == "double" {
		context.Rectangle(bbx+bl/3, bby+bt/3, bbw-(bl+br)/3, bbh-(bt+bb)/3)
		context.Rectangle(bbx+bl*2/3, bby+bt*2/3, bbw-(bl+br)*2/3, bbh-(bt+bb)*2/3)
	}
	context.Rectangle(bbx+bl, bby+bt, bbw-bl-br, bbh-bt-bb)
	context.SetSourceRgba(color[0].Unpack())
	context.Fill()
}

func drawOutlines(context Drawer, box_ Box, enableHinting bool) {
	box := box_.Box()
	width_ := box.Style.GetOutlineWidth()
	color := box.Style.ResolveColor("outline_color").RGBA
	style := box.Style.GetOutlineStyle()
	if box.Style.GetVisibility() == "visible" && !width_.IsNone() && color.A != 0 {
		width := width_.Value
		outlineBox := pr.Rectangle{
			box.BorderBoxX() - width, box.BorderBoxY() - width,
			box.BorderWidth() + 2*width, box.BorderHeight() + 2*width,
		}
		for _, side := range SIDES {
			context.OnNewStack(func() error {
				clipBorderSegment(context, enableHinting, style, float64(width), side, outlineBox, nil, nil)
				drawRectBorder(context, outlineBox, pr.Rectangle{width, width, width, width},
					style, styledColor(style, color, side))
				return nil
			})
		}
	}

	if bo.IsParentBox(box_) {
		for _, child := range box.Children {
			if bo.IsBox(child) {
				drawOutlines(context, child, enableHinting)
			}
		}
	}
}

type segment struct {
	bo.Border
	side      string
	borderBox pr.Rectangle
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Draw borders of table cells when they collapse.
func drawCollapsedBorders(context Drawer, table *bo.TableBox, enableHinting bool) {
	var rowHeights, rowPositions []pr.Float
	for _, rowGroup := range table.Children {
		for _, row := range rowGroup.Box().Children {
			rowHeights = append(rowHeights, row.Box().Height.V())
			rowPositions = append(rowPositions, row.Box().PositionY)
		}
	}
	columnWidths := table.ColumnWidths
	if len(rowHeights) == 0 || len(columnWidths) == 0 {
		// One of the list is empty: don’t bother with empty tables
		return
	}
	columnPositions := table.ColumnPositions
	gridHeight := len(rowHeights)
	gridWidth := len(columnWidths)
	if gridWidth != len(columnPositions) {
		log.Fatalf("expected same gridWidth and columnPositions length, got %d, %d", gridWidth, len(columnPositions))
	}
	// Add the end of the last column, but make a copy from the table attr.
	columnPositions = append(columnPositions, columnPositions[len(columnPositions)-1]+columnWidths[len(columnWidths)-1])
	// Add the end of the last row. No copy here, we own this list
	rowPositions = append(rowPositions, rowPositions[len(rowPositions)-1]+rowHeights[len(rowHeights)-1])
	verticalBorders, horizontalBorders := table.CollapsedBorderGrid.Vertical, table.CollapsedBorderGrid.Horizontal
	headerRows := 0
	if table.Children[0].Box().IsHeader {
		headerRows = len(table.Children[0].Box().Children)
	}
	footerRows := 0
	if L := len(table.Children); table.Children[L-1].Box().IsFooter {
		footerRows = len(table.Children[L-1].Box().Children)
	}
	skippedRows := table.SkippedRows
	bodyRowsOffset := 0
	if skippedRows != 0 {
		bodyRowsOffset = skippedRows - headerRows
	}
	if headerRows == 0 {
		headerRows = -1
	}
	firstFooterRow := gridHeight + 1
	if footerRows != 0 {
		firstFooterRow = gridHeight - footerRows - 1
	}
	originalGridHeight := len(verticalBorders)
	footerRowsOffset := originalGridHeight - gridHeight

	rowNumber := func(y int, horizontal bool) int {
		if y < (headerRows + boolToInt(horizontal)) {
			return y
		} else if y >= (firstFooterRow + boolToInt(horizontal)) {
			return y + footerRowsOffset
		} else {
			return y + bodyRowsOffset
		}
	}

	var segments []segment

	// vertical=true
	halfMaxWidth := func(borderList [][]bo.Border, yxPairs [][2]int, vertical bool) pr.Float {
		var result pr.Float
		for _, tmp := range yxPairs {
			y, x := tmp[0], tmp[1]
			cond := 0 <= y && y <= gridHeight && 0 <= x && x < gridWidth
			if vertical {
				cond = 0 <= y && y < gridHeight && 0 <= x && x <= gridWidth
			}
			if cond {
				yy := rowNumber(y, !vertical)
				width := pr.Float(borderList[yy][x].Width)
				result = pr.Max(result, width)
			}
		}
		return result / 2
	}

	addVertical := func(x, y int) {
		yy := rowNumber(y, false)
		border := verticalBorders[yy][x]
		if border.Width == 0 || border.Color.RGBA.A == 0 {
			return
		}
		posX := columnPositions[x]
		posY1 := rowPositions[y] - halfMaxWidth(horizontalBorders,
			[][2]int{{y, x - 1}, {y, x}}, false)
		posY2 := rowPositions[y+1] + halfMaxWidth(horizontalBorders,
			[][2]int{{y + 1, x - 1}, {y + 1, x}}, false)
		segments = append(segments, segment{
			Border: border, side: "left",
			borderBox: pr.Rectangle{posX - pr.Float(border.Width)/2, posY1, 0, posY2 - posY1},
		})
	}

	addHorizontal := func(x, y int) {
		yy := rowNumber(y, true)
		border := horizontalBorders[yy][x]
		if border.Width == 0 || border.Color.RGBA.A == 0 {
			return
		}
		posY := rowPositions[y]
		// TODO: change signs for rtl when we support rtl tables?
		posX1 := columnPositions[x] - halfMaxWidth(verticalBorders,
			[][2]int{{y - 1, x}, {y, x}}, true)
		posX2 := columnPositions[x+1] + halfMaxWidth(verticalBorders,
			[][2]int{{y - 1, x + 1}, {y, x + 1}}, true)
		segments = append(segments, segment{
			Border: border, side: "top",
			borderBox: pr.Rectangle{posX1, posY - pr.Float(border.Width)/2, posX2 - posX1, 0},
		})
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
		return segments[i].Border.Score.Lower(segments[j].Border.Score)
	})

	for _, segment := range segments {
		widths := pr.Rectangle{0, 0, 0, pr.Float(segment.Width)}
		if segment.side == "top" {
			widths = pr.Rectangle{pr.Float(segment.Width), 0, 0, 0}
		}
		context.OnNewStack(func() error {
			clipBorderSegment(context, enableHinting, segment.Style, segment.Width, segment.side, segment.borderBox,
				&widths, nil)
			drawRectBorder(context, segment.borderBox, widths, segment.Style,
				styledColor(segment.Style, segment.Color.RGBA, segment.side))
			return nil
		})
	}
}

// Draw the given :class:`boxes.ReplacedBox` to a ``cairo.context``.
func drawReplacedbox(context Drawer, box_ bo.InstanceReplacedBox) {
	box := box_.Replaced()
	if box.Style.GetVisibility() != "visible" || !pr.Is(box.Width) || !pr.Is(box.Height) {
		return
	}

	drawWidth, drawHeight, drawX, drawY := layout.ReplacedboxLayout(box_)

	context.OnNewStack(func() error {
		roundedBoxPath(context, box.RoundedContentBox())
		context.Clip()
		context.Translate(float64(drawX), float64(drawY))
		box.Replacement.Draw(context, float64(drawWidth), float64(drawHeight), box.Style.GetImageRendering())
		return nil
	})
}

// offsetX=0, textOverflow="clip"
func drawInlineLevel(context Drawer, page *bo.PageBox, box_ Box, enableHinting bool, offsetX float64, textOverflow string) error {
	if stackingContext, ok := box_.(StackingContext); ok {
		if !(bo.TypeInlineBlockBox.IsInstance(stackingContext.box) || bo.TypeInlineFlexBox.IsInstance(stackingContext.box)) {
			log.Fatalf("expected InlineBlock or InlineFlex, got %v", stackingContext.box)
		}
		if err := drawStackingContext(context, stackingContext, enableHinting); err != nil {
			return err
		}
	} else {
		box := box_.Box()
		if err := drawBackground2(context, box.Background, enableHinting); err != nil {
			return err
		}
		drawBorder(context, box_, enableHinting)
		textBox, isTextBox := box_.(*bo.TextBox)
		replacedBox, isReplacedBox := box_.(bo.InstanceReplacedBox)
		if layout.IsLine(box_) {
			if lineBox, ok := box_.(*bo.LineBox); ok {
				textOverflow = lineBox.TextOverflow
			}
			for _, child := range box.Children {
				childOffsetX := offsetX
				if _, ok := child.(StackingContext); !ok {
					childOffsetX = offsetX + float64(child.Box().PositionX) - float64(box.PositionX)
				}
				if child, ok := child.(*bo.TextBox); ok {
					drawText(context, *child, enableHinting, childOffsetX, textOverflow)
				} else {
					if err := drawInlineLevel(context, page, child, enableHinting, childOffsetX, textOverflow); err != nil {
						return err
					}
				}
			}
		} else if isReplacedBox {
			drawReplacedbox(context, replacedBox)
		} else if isTextBox {
			// Should only happen for list markers
			drawText(context, *textBox, enableHinting, offsetX, textOverflow)
		} else {
			log.Fatalf("unexpected box %v", box)
		}
	}
	return nil
}

// Draw ``textbox`` to a ``cairo.Context`` from ``PangoCairo.Context``
// 	(offsetX=0,textOverflow="clip")
func drawText(context Drawer, textbox bo.TextBox, enableHinting bool, offsetX float64, textOverflow string) {
	// Pango crashes with font-size: 0
	// FIXME: should we keep this assertion ?
	// assert textbox.Style["fontSize"]

	if textbox.Style.GetVisibility() != "visible" {
		return
	}

	context.MoveTo(float64(textbox.PositionX), float64(textbox.PositionY+textbox.Baseline.V()))
	context.SetSourceRgba(textbox.Style.GetColor().RGBA.Unpack())

	// textbox.PangoLayout.Reactivate(textbox.Style)
	// FIXME:
	// layout.ShowFirstLine(context, textbox, textOverflow)

	values := textbox.Style.GetTextDecorationLine().Decorations

	thickness := float64(textbox.Style.GetFontSize().Value / 18) // Like other browsers do
	if enableHinting && thickness < 1 {
		thickness = 1
	}

	color := textbox.Style.GetTextDecorationColor()
	if color.Type == parser.ColorCurrentColor {
		color = textbox.Style.GetColor()
	}

	var metrics interface{}
	if values.Has("overline") || values.Has("line-through") || values.Has("underline") {
		metrics = textbox.PangoLayout.GetFontMetrics()
	}
	fmt.Println(metrics)
	if values.Has("overline") {
		// FIXME:
		// drawTextDecoration(context, textbox, offsetX, textbox.Baseline-metrics.ascent+thickness/2, thickness, enableHinting, color.RGBA)
	}
	if values.Has("underline") {
		// FIXME:
		// drawTextDecoration(context, textbox, offsetX, textbox.Baseline-metrics.underlinePosition+thickness/2,
		// 	thickness, enableHinting, color.RGBA)
	}
	if values.Has("line-through") {
		// FIXME:
		// drawTextDecoration(context, textbox, offsetX, textbox.Baseline-metrics.strikethroughPosition,
		// 	thickness, enableHinting, color.RGBA)
	}

	// textbox.PangoLayout.deactivate()
}

func drawWave(context Drawer, x, y, width, offsetX, radius float64) {
	context.NewPath()
	diameter := 2 * radius
	waveIndex := offsetX // diameter
	remain := offsetX - waveIndex*diameter

	for width > 0 {
		up := utils.FloatModulo(waveIndex, 2) == 0
		centerX := x - remain + radius
		alpha1 := (1 + remain/diameter) * pi
		alpha2 := (1 + math.Min(1, width/diameter)) * pi

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
func drawTextDecoration(context Drawer, textbox bo.TextBox, offsetX, offsetY, thickness float64,
	enableHinting bool, color Color) {

	style := textbox.Style.GetTextDecorationStyle()
	// with stacked(context) {
	if enableHinting {
		context.SetAntialias(backend.AntialiasNone)
	}
	context.SetSourceRgba(color.Unpack())
	context.SetLineWidth(thickness)

	if style == "dashed" {
		context.SetDash([]float64{5 * thickness}, offsetX)
	} else if style == "dotted" {
		context.SetDash([]float64{thickness}, offsetX)
	}
	posX, posY, width := float64(textbox.PositionX), float64(textbox.PositionY), float64(textbox.Width.V())
	if style == "wavy" {
		drawWave(context, posX, posY+offsetY, width, offsetX, 0.75*thickness)
	} else {
		context.MoveTo(posX, posY+offsetY)
		context.RelLineTo(width, 0)
	}

	if style == "double" {
		delta := 2 * thickness
		context.MoveTo(posX, posY+offsetY+delta)
		context.RelLineTo(width, 0)
	}

	context.Stroke()
}
