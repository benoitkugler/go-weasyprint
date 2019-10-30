package layout

import (
	"log"
	"math"
	"strings"

	bo "github.com/benoitkugler/go-weasyprint/boxes"
	"github.com/benoitkugler/go-weasyprint/images"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/utils"
)

func boxRectangle(box bo.BoxFields, whichRectangle string) [4]pr.Float {
	switch whichRectangle {
	case "border-box":
		return [4]pr.Float{
			box.BorderBoxX(),
			box.BorderBoxY(),
			box.BorderWidth(),
			box.BorderHeight(),
		}
	case "padding-box":
		return [4]pr.Float{
			box.PaddingBoxX(),
			box.PaddingBoxY(),
			box.PaddingWidth(),
			box.PaddingHeight(),
		}
	case "content-box":
		return [4]pr.Float{
			box.ContentBoxX(),
			box.ContentBoxY(),
			box.Width.V(),
			box.Height.V(),
		}
	default:
		log.Fatalf("unexpected whichRectangle : %s", whichRectangle)
		return [4]pr.Float{}
	}
}

// emulate Python itertools.cycle
// i is the current iteration index, N the length of the target slice.
func cycle(i, N int) int {
	return i % N
}

// Fetch and position background images.
func layoutBoxBackgrounds(page *bo.PageBox, box_ Box, getImageFromUri bo.Gifu) {
	// Resolve percentages in border-radius properties
	box := box_.Box()
	resolveRadiiPercentages(box)

	for _, child := range box.AllChildren() {
		layoutBoxBackgrounds(page, child, getImageFromUri)
	}

	style := box.Style
	if style.GetVisibility() == "hidden" {
		box.Background = nil
		if page != box_ { // Pages need a background for bleed box
			return
		}
	}

	bs := style.GetBackgroundImage()
	images := make([]images.Image, len(bs))
	anyImages := false
	for i, v := range bs {
		//FIXME: quel lien entre pr.Image et images.Image ??
		images[i] = v
		if url, ok := v.(pr.UrlImage); ok {
			images[i] = getImageFromUri(string(url), "")
		}
		if images[i] != nil {
			anyImages = true
		}
	}
	color := style.ResolveColor("background_color")
	if color.RGBA.A == 0 && !anyImages {
		box.Background = nil
		if page != box_ { // Pages need a background for bleed box
			return
		}
	}

	sizes := style.GetBackgroundSize()
	sizesN := len(sizes)
	clips := style.GetBackgroundClip()
	clipsN := len(clips)
	repeats := style.GetBackgroundRepeat()
	repeatsN := len(repeats)
	origins := style.GetBackgroundOrigin()
	originsN := len(origins)
	positions := style.GetBackgroundPosition()
	positionsN := len(positions)
	attachments := style.GetBackgroundAttachment()
	attachmentsN := len(attachments)

	ir := style.GetImageResolution()
	layers := make([]bo.BackgroundLayer, len(images))
	for i, img := range images {
		layers[i] = layoutBackgroundLayer(box_, page, ir, img,
			sizes[cycle(i, sizesN)],
			clips[cycle(i, clipsN)],
			repeats[cycle(i, repeatsN)],
			origins[cycle(i, originsN)],
			positions[cycle(i, positionsN)],
			attachments[cycle(i, attachmentsN)],
		)
	}

	box.Background = &bo.Background{Color: color, ImageRendering: style.GetImageRendering(), Layers: layers}
}

func layoutBackgroundLayer(box_ Box, page *bo.PageBox, resolution pr.Value, image images.Image, size pr.Size, clip string, repeat [2]string,
	origin string, position pr.Center, attachment string) bo.BackgroundLayer {

	// TODO: respect box-sizing for table cells?
	var (
		clippedBoxes []bo.RoundedBox
		paintingArea [4]pr.Float
	)
	box := box_.Box()
	if box_ == page {
		paintingArea = [4]pr.Float{0, 0, page.MarginWidth(), page.MarginHeight()}
		// XXX: how does border-radius work on pages?
		clippedBoxes = []bo.RoundedBox{box.RoundedBorderBox()}
	} else if bo.TypeTableRowGroupBox.IsInstance(box_) {
		clippedBoxes = nil
		var totalHeight pr.Float
		for _, row_ := range box.Children {
			row := row_.Box()
			if len(row.Children) != 0 {
				var max pr.Float
				for _, cell := range row.Children {
					clippedBoxes = append(clippedBoxes, cell.Box().RoundedBorderBox())
					if v := cell.Box().BorderBoxY() + cell.Box().BorderHeight(); v > max {
						max = v
					}
				}
				totalHeight = pr.Max(totalHeight, max)
			}
		}
		paintingArea = [4]pr.Float{box.BorderBoxX(), box.BorderBoxY(),
			box.BorderBoxX() + box.BorderWidth(), totalHeight}
	} else if bo.TypeTableRowBox.IsInstance(box_) {
		if len(box.Children) != 0 {
			clippedBoxes = nil
			var max pr.Float
			for _, cell := range box.Children {
				clippedBoxes = append(clippedBoxes, cell.Box().RoundedBorderBox())
				if v := cell.Box().BorderHeight(); v > max {
					max = v
				}
			}
			height := max
			paintingArea = [4]pr.Float{box.BorderBoxX(), box.BorderBoxY(),
				box.BorderBoxX() + box.BorderWidth(), box.BorderBoxY() + height}
		}
	} else if bo.TypeTableColumnGroupBox.IsInstance(box_) || bo.TypeTableColumnBox.IsInstance(box_) {
		cells := box.GetCells()
		if len(cells) != 0 {
			clippedBoxes = nil
			var max pr.Float
			for _, cell := range cells {
				clippedBoxes = append(clippedBoxes, cell.Box().RoundedBorderBox())
				if v := cell.Box().BorderBoxY() + cell.Box().BorderWidth(); v > max {
					max = v
				}
			}
			maxX := max
			paintingArea = [4]pr.Float{box.BorderBoxX(), box.BorderBoxY(),
				maxX - box.BorderBoxX(), box.BorderBoxY() + box.BorderHeight()}
		}
	} else {
		paintingArea = boxRectangle(*box, clip)
		switch clip {
		case "border-box":
			clippedBoxes = []bo.RoundedBox{box.RoundedBorderBox()}
		case "padding-box":
			clippedBoxes = []bo.RoundedBox{box.RoundedPaddingBox()}
		case "content-box":
			clippedBoxes = []bo.RoundedBox{box.RoundedContentBox()}
		default:
			log.Fatalf("unexpected clip : %s", clip)
		}
	}

	var hasZero bool
	if image != nil {
		v1, v2 := image.GetIntrinsicSize(pr.FToV(1), pr.FToV(1))
		hasZero = v1 == pr.Float(0) || v2 == pr.Float(0)
	}
	if image == nil || hasZero {
		return bo.BackgroundLayer{
			Image: nil, Unbounded: box_ == page, PaintingArea: bo.Area{Rect: paintingArea},
			Size: pr.Size{String: "unused"}, Position: bo.Position{String: "unused"}, Repeat: bo.Repeat{String: "unused"},
			PositioningArea: bo.Area{String: "unused"}, ClippedBoxes: clippedBoxes}
	}

	var positioningArea [4]pr.Float
	if attachment == "fixed" {
		// Initial containing block
		positioningArea = boxRectangle(page.BoxFields, "content-box")
	} else {
		positioningArea = boxRectangle(*box, origin)
	}

	_, _, positioningWidth, positioningHeight := positioningArea[0], positioningArea[1], positioningArea[2], positioningArea[3]
	// paintingX, paintingY, paintingWidth, paintingHeight := paintingArea[0], paintingArea[1], paintingArea[2], paintingArea[3]
	var imageWidth, imageHeight pr.Float
	if size.String == "cover" {
		imageWidth, imageHeight = coverConstraintImageSizing(positioningWidth, positioningHeight, image.IntrinsicRatio())
	} else if size.String == "contain" {
		imageWidth, imageHeight = containConstraintImageSizing(positioningWidth, positioningHeight, image.IntrinsicRatio())
	} else {
		sizeWidth, sizeHeight := size.Width, size.Height
		iwidth, iheight := image.GetIntrinsicSize(resolution, box.Style.GetFontSize())
		imageWidth, imageHeight = defaultImageSizing(iwidth, iheight, image.IntrinsicRatio(),
			percentage(sizeWidth, positioningWidth), percentage(sizeHeight, positioningHeight), positioningWidth, positioningHeight)
	}

	originX, positionX_, originY, positionY_ := position.OriginX, position.Pos[0], position.OriginY, position.Pos[1]
	refX := positioningWidth - imageWidth
	refY := positioningHeight - imageHeight
	positionX := percentage(positionX_.ToValue(), refX)
	positionY := percentage(positionY_.ToValue(), refY)
	if originX == "right" {
		positionX = refX - positionX.V()
	}
	if originY == "bottom" {
		positionY = refY - positionY.V()
	}

	repeatX, repeatY := repeat[0], repeat[1]

	if repeatX == "round" {
		nRepeats := utils.MaxInt(1, int(math.Round(float64(positioningWidth/imageWidth))))
		newWidth := positioningWidth / pr.Float(nRepeats)
		positionX = pr.Float(0) // Ignore background-position for this dimension
		if repeatY != "round" && size.Height.String == "auto" {
			imageHeight *= newWidth / imageWidth
		}
		imageWidth = newWidth
	}
	if repeatY == "round" {
		nRepeats := utils.MaxInt(1, int(math.Round(float64(positioningHeight/imageHeight))))
		newHeight := positioningHeight / pr.Float(nRepeats)
		positionY = pr.Float(0) // Ignore background-position for this dimension
		if repeatX != "round" && size.Width.String == "auto" {
			imageWidth *= newHeight / imageHeight
		}
		imageHeight = newHeight
	}

	return bo.BackgroundLayer{
		Image:           image,
		Size:            pr.Size{Width: imageWidth.ToValue(), Height: imageHeight.ToValue()},
		Position:        bo.Position{Point: bo.MaybePoint{positionX, positionY}},
		Repeat:          bo.Repeat{Reps: repeat},
		Unbounded:       box_ == page,
		PaintingArea:    bo.Area{Rect: paintingArea},
		PositioningArea: bo.Area{Rect: positioningArea},
		ClippedBoxes:    clippedBoxes,
	}
}

// Set a ``canvasBackground`` attribute on the PageBox,
//  with style for the canvas background, taken from the root elememt
//  or a <body> child of the root element.
//  See http://www.w3.org/TR/CSS21/colors.html#background
func setCanvasBackground(page *bo.PageBox) {
	rootBox_ := page.Children[0]
	rootBox := rootBox_.Box()
	if bo.IsMarginBox(rootBox_) {
		log.Fatalf("unexpected margin box as first child of page")
	}
	chosenBox_ := rootBox_
	if strings.ToLower(rootBox.ElementTag) == "html" && rootBox.Background == nil {
		for _, child := range rootBox.Children {
			if strings.ToLower(child.Box().ElementTag) == "body" {
				chosenBox_ = child
				break
			}
		}
	}
	chosenBox := chosenBox_.Box()
	if chosenBox.Background != nil {
		paintingArea := boxRectangle(page.BoxFields, "padding-box")
		for i, l := range chosenBox.Background.Layers {
			l.PaintingArea = bo.Area{Rect: paintingArea}
			chosenBox.Background.Layers[i] = l
		}
		// TODO: shouldn’t background-clip be considered here?
		*page.CanvasBackground = *chosenBox.Background
		chosenBox.Background = nil
	} else {
		page.CanvasBackground = nil
	}
}

func LayoutBackgrounds(page *bo.PageBox, getImageFromUri bo.Gifu) {
	layoutBoxBackgrounds(page, page, getImageFromUri)
	setCanvasBackground(page)
}
