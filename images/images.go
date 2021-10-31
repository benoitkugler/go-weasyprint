// Fetch and decode images in range various formats.
package images

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"math"
	"strings"

	"github.com/benoitkugler/go-weasyprint/backend"
	"github.com/benoitkugler/go-weasyprint/style/parser"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/utils"
	"github.com/benoitkugler/oksvg/svgicon"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

type Color = parser.RGBA

// Image is the common interface for supported image format,
// such as gradient, SVG, or JPEG, PNG, etc...
type Image interface {
	isImage()
	GetIntrinsicSize(imageResolution, fontSize pr.Value) (width, height, ratio pr.MaybeFloat)
	Draw(context backend.OutputGraphic, concreteWidth, concreteHeight pr.Fl, imageRendering pr.String)
}

var (
	_ Image = RasterImage{}
	_ Image = &SVGImage{}
	_ Image = LinearGradient{}
	_ Image = RadialGradient{}
)

type imageSurface interface {
	getWidth() pr.Float
	getHeight() pr.Float
}

// Map values of the image-rendering property to cairo FILTER values.
// Values are normalized to lower case.
var imagesRenderingToFilter = map[string]int{
	"auto":        0, // cairocffi.FILTERBILINEAR,
	"crisp-edges": 1, // cairocffi.FILTERBEST,
	"pixelated":   2, // cairocffi.FILTERNEAREST,
}

// An error occured when loading an image.
// The image data is probably corrupted or in an invalid format.
func imageLoadingError(err error) error {
	return fmt.Errorf("error loading image : %s", err)
}

type RasterImage struct {
	imageSurface    image.Image
	intrinsicRatio  pr.Float
	intrinsicWidth  pr.Float
	intrinsicHeight pr.Float
	optimizeSize    bool
}

func NewRasterImage(imageSurface image.Image, optimizeSize bool) RasterImage {
	self := RasterImage{}
	self.optimizeSize = optimizeSize
	self.imageSurface = imageSurface
	self.intrinsicWidth = pr.Float(imageSurface.Bounds().Dx())
	self.intrinsicHeight = pr.Float(imageSurface.Bounds().Dy())
	self.intrinsicRatio = pr.Inf
	if self.intrinsicHeight != 0 {
		self.intrinsicRatio = self.intrinsicWidth / self.intrinsicHeight
	}
	return self
}

func (r RasterImage) isImage() {}

func (r RasterImage) GetIntrinsicSize(imageResolution, _ pr.Value) (width, height, ratio pr.MaybeFloat) {
	// Raster images are affected by the "image-resolution" property.
	return r.intrinsicWidth / imageResolution.Value, r.intrinsicHeight / imageResolution.Value, r.intrinsicRatio
}

func (r RasterImage) Draw(context backend.OutputGraphic, concreteWidth, concreteHeight pr.Fl, imageRendering pr.String) {
	hasSize := concreteWidth > 0 && concreteHeight > 0 && r.intrinsicWidth > 0 && r.intrinsicHeight > 0
	if !hasSize {
		return
	}

	context.DrawRasterImage(r.imageSurface, string(imageRendering), concreteWidth, concreteHeight)
}

// class ScaledSVGSurface(cairosvg.surface.SVGSurface) {
//     """
//     Have the cairo Surface object have intrinsic dimension
//     := range pixels instead of points.
//     """
//     @property
//  func deviceUnitsPerUserUnits(self) {
//         scale = super(ScaledSVGSurface, self).deviceUnitsPerUserUnits
//         return scale / 0.75
//     }
// }

// Fake CairoSVG surface used to get SVG attributes.
type FakeSurface struct {
	contextHeight float64
	contextWidth  float64
	fontSize      float64
	dpi           float64
}

func newFakeSurface() FakeSurface {
	return FakeSurface{
		contextHeight: 0,
		contextWidth:  0,
		fontSize:      12,
		dpi:           96,
	}
}

type SVGImage struct {
	icon       *svgicon.SvgIcon
	urlFetcher utils.UrlFetcher
	baseUrl    string
	svgData    []byte
	width      float64
	height     float64
}

func (SVGImage) isImage() {}

func NewSVGImage(svgData io.Reader, baseUrl string, urlFetcher utils.UrlFetcher) (*SVGImage, error) {
	self := new(SVGImage)
	// Don’t pass data URIs to CairoSVG.
	// They are useless for relative URIs anyway.
	if !strings.HasPrefix(strings.ToLower(baseUrl), "data:") {
		self.baseUrl = baseUrl
	}
	content, err := ioutil.ReadAll(svgData)
	if err != nil {
		return nil, imageLoadingError(err)
	}
	self.svgData = content
	self.urlFetcher = urlFetcher

	self.icon, err = svgicon.ReadIconStream(bytes.NewReader(self.svgData), svgicon.StrictErrorMode)
	if err != nil {
		return nil, imageLoadingError(err)
	}
	return self, nil
}

//  func CairosvgUrlFetcher(self, src, mimetype) {
//         data = self.urlFetcher(src)
//         if "string" := range data {
//             return data["string"]
//         } return data["fileObj"].read()
//     }

func (s *SVGImage) GetIntrinsicSize(_, fontSize pr.Value) (pr.MaybeFloat, pr.MaybeFloat, pr.MaybeFloat) {
	width, height := s.icon.Width, s.icon.Height
	if width == "" {
		width = "100%"
	}
	if height == "" {
		height = "100%"
	}

	var intrinsicWidth, intrinsicHeight, ratio pr.MaybeFloat
	if !strings.ContainsRune(width, '%') {
		intrinsicWidth = size(width, fontSize.Value, nil)
	}
	if !strings.ContainsRune(height, '%') {
		intrinsicHeight = size(height, fontSize.Value, nil)
	}

	if intrinsicWidth == nil || intrinsicHeight == nil {
		viewbox := s.icon.ViewBox
		if viewbox.W != 0 && viewbox.H != 0 {
			ratio = pr.Float(viewbox.W / viewbox.H)
			if pr.Is(intrinsicWidth) {
				intrinsicHeight = intrinsicWidth.V() / ratio.V()
			} else if pr.Is(intrinsicHeight) {
				intrinsicWidth = intrinsicHeight.V() * ratio.V()
			}
		}
	} else if pr.Is(intrinsicWidth) && pr.Is(intrinsicHeight) {
		ratio = intrinsicWidth.V() / intrinsicHeight.V()
	}

	return intrinsicWidth, intrinsicHeight, ratio
}

func (SVGImage) Draw(context backend.OutputGraphic, concreteWidth, concreteHeight float64, imageRendering pr.String) {
	log.Println("SVG rendering not implemented yet")
	// FIXME:
	//         try {
	//             svg = ScaledSVGSurface(
	//                 cairosvg.parser.Tree(
	//                     bytestring=self.svgData, url=self.baseUrl,
	//                     urlFetcher=self.CairosvgUrlFetcher),
	//                 output=nil, dpi=96, outputWidth=concreteWidth,
	//                 outputHeight=concreteHeight)
	//             if svg.width && svg.height {
	//                 context.scale(
	//                     concreteWidth / svg.width, concreteHeight / svg.height)
	//                 context.setSourceSurface(svg.cairo)
	//                 context.paint()
	//             }
	//         } except Exception as e {
	//             LOGGER.error(
	//                 "Failed to draw an SVG image at %s : %s", self.baseUrl, e)
	//         }
}

// Cache stores the result of fetching an image, or any error encoutered
type Cache struct {
	images map[string]Image
	errors map[string]error
}

func NewCache() Cache {
	return Cache{
		images: make(map[string]Image),
		errors: make(map[string]error),
	}
}

// Get a cairo Pattern from an image URI.
func GetImageFromUri(cache Cache, fetcher utils.UrlFetcher, optimizeSize bool, url, forcedMimeType string) (Image, error) {
	img, in := cache.images[url]
	if in {
		return img, nil
	}
	if err, in := cache.errors[url]; in {
		return nil, err
	}

	var err error
	defer func() {
		cache.images[url] = img
		cache.errors[url] = err
	}()

	content, err := fetcher(url)
	if err != nil {
		err = fmt.Errorf(`Failed to load image at "%s" (%s)`, url, err)
		return nil, err
	}

	mimeType := forcedMimeType
	if mimeType == "" {
		mimeType = content.MimeType
	}

	var errSvg error
	// Try to rely on given mimetype for SVG
	if mimeType == "image/svg+xml" {
		var svgIm *SVGImage
		svgIm, errSvg = NewSVGImage(content.Content, url, fetcher)
		if svgIm != nil {
			img = svgIm
		}
	}

	// Try pillow for raster images, or for failing SVG
	if img == nil {
		content.Content.Seek(0, io.SeekStart)
		parsedImage, _, errRaster := image.Decode(content.Content)
		if errRaster != nil {
			if errSvg != nil { // Tried SVGImage then raster for a SVG, abort
				err = fmt.Errorf(`Failed to load image at "%s" (%s)`, url, errSvg)
				return nil, err
			}

			// Last chance, try SVG in case mime type in incorrect
			content.Content.Seek(0, io.SeekStart)
			img, errSvg = NewSVGImage(content.Content, url, fetcher)
			if errSvg != nil {
				err = fmt.Errorf(`Failed to load image at "%s" (%s)`, url, errRaster)
				return nil, err
			}
		} else {
			img = NewRasterImage(parsedImage, optimizeSize)
		}
	}

	return img, err
}

// Gradient line size: distance between the starting point and ending point.
// Positions: list of Dimension in px or % (possibliy zero)
// 0 is the starting point, 1 the ending point.
// http://dev.w3.org/csswg/css-images-3/#color-stop-syntax
// Return processed color stops, as a list of floats in px.
func processColorStops(gradientLineSize pr.Float, positions_ []pr.Dimension) []pr.Fl {
	L := len(positions_)
	positions := make([]pr.MaybeFloat, L)
	for i, position := range positions_ {
		positions[i] = pr.ResoudPercentage(position.ToValue(), gradientLineSize)
	}
	// First and last default to 100%
	if positions[0] == nil {
		positions[0] = pr.Float(0)
	}
	if positions[L-1] == nil {
		positions[L-1] = gradientLineSize
	}

	// Make sure positions are increasing.
	previousPos := positions[0].V()
	for i, position := range positions {
		if position != nil {
			if position.V() < previousPos {
				positions[i] = previousPos
			} else {
				previousPos = position.V()
			}
		}
	}

	// Assign missing values
	previousI := L - 1
	for i, position := range positions {
		if position != nil {
			base := positions[previousI]
			increment := (position.V() - base.V()) / pr.Float(i-previousI)
			for j := previousI + 1; j < i; j += 1 {
				positions[j] = base.V() + pr.Float(j)*increment
			}
			previousI = i
		}
	}
	out := make([]pr.Fl, L)
	for i, v := range positions {
		out[i] = pr.Fl(v.V())
	}
	return out
}

// Normalize to [0..1].
// Write on positions.
func normalizeStopPostions(positions []pr.Fl) (pr.Fl, pr.Fl) {
	first := positions[0]
	last := positions[len(positions)-1]
	totalLength := last - first
	if totalLength != 0 {
		for i, pos := range positions {
			positions[i] = (pos - first) / totalLength
		}
	} else {
		for i := range positions {
			positions[i] = 0
		}
	}
	return first, last
}

// http://dev.w3.org/csswg/css-images-3/#find-the-average-color-of-a-gradient
func gradientAverageColor(colors []Color, positions []pr.Fl) Color {
	nbStops := len(positions)
	if nbStops <= 1 || nbStops != len(colors) {
		panic(fmt.Sprintf("expected same length, at least 2, got %d, %d", nbStops, len(colors)))
	}
	totalLength := positions[nbStops-1] - positions[0]
	if totalLength == 0 {
		for i := range positions {
			positions[i] = pr.Fl(i)
		}
		totalLength = pr.Fl(nbStops - 1)
	}
	premulR := make([]utils.Fl, nbStops)
	premulG := make([]utils.Fl, nbStops)
	premulB := make([]utils.Fl, nbStops)
	alpha := make([]utils.Fl, nbStops)
	for i, col := range colors {
		premulR[i] = col.R * col.A
		premulG[i] = col.G * col.A
		premulB[i] = col.B * col.A
		alpha[i] = col.A
	}
	var resultR, resultG, resultB, resultA utils.Fl
	totalWeight := 2 * totalLength
	for i_, position := range positions[1:] {
		i := i_ + 1
		weight := utils.Fl((position - positions[i-1]) / totalWeight)
		j := i - 1
		resultR += premulR[j] * weight
		resultG += premulG[j] * weight
		resultB += premulB[j] * weight
		resultA += alpha[j] * weight
		j = i
		resultR += premulR[j] * weight
		resultG += premulG[j] * weight
		resultB += premulB[j] * weight
		resultA += alpha[j] * weight
	}
	// Un-premultiply:
	if resultA != 0 {
		return Color{
			R: resultR / resultA,
			G: resultG / resultA,
			B: resultB / resultA,
			A: resultA,
		}
	}
	return Color{}
}

var patternsT = map[string]int{
	"linear": 0, // cairocffi.LinearGradient,
	"radial": 1, // cairocffi.RadialGradient,
	"solid":  2, // cairocffi.SolidPattern,
}

type layouter interface {
	// width, height: Gradient box. Top-left is at coordinates (0, 0).
	Layout(width, height pr.Float) backend.GradientLayout
}

type gradient struct {
	layouter

	colors        []Color
	stopPositions []pr.Dimension
	repeating     bool
}

func newGradient(colorStops []pr.ColorStop, repeating bool) gradient {
	self := gradient{}
	if len(colorStops) == 0 {
		log.Fatalf("expected non empty colorStops slice")
	}
	self.colors = make([]Color, len(colorStops))
	self.stopPositions = make([]pr.Dimension, len(colorStops))
	for i, v := range colorStops {
		self.colors[i] = v.Color.RGBA
		self.stopPositions[i] = v.Position
	}
	self.repeating = repeating
	return self
}

func (g gradient) GetIntrinsicSize(_, _ pr.Value) (pr.MaybeFloat, pr.MaybeFloat, pr.MaybeFloat) {
	// Gradients are not affected by image resolution, parent or font size.
	return nil, nil, nil
}

func (g gradient) Draw(dst backend.OutputGraphic, concreteWidth, concreteHeight pr.Fl, imageRendering pr.String) {
	layout := g.layouter.Layout(pr.Float(concreteWidth), pr.Float(concreteHeight))

	if layout.Kind == "solid" {
		dst.Rectangle(0, 0, concreteWidth, concreteHeight)
		dst.SetColorRgba(layout.Colors[0], false)
		dst.Fill(false)
		return
	}

	dst.DrawGradient(layout, concreteWidth, concreteHeight)
}

type LinearGradient struct {
	direction pr.DirectionType
	gradient
}

func (LinearGradient) isImage() {}

func NewLinearGradient(from pr.LinearGradient) LinearGradient {
	self := LinearGradient{gradient: newGradient(from.ColorStops, from.Repeating)}
	self.layouter = self
	// ("corner", keyword) or ("angle", radians)
	self.direction = from.Direction
	return self
}

func toData(c Color) [6]pr.Fl {
	return [6]pr.Fl{pr.Fl(c.R), pr.Fl(c.G), pr.Fl(c.B), pr.Fl(c.A), 0, 0}
}

func reverseColors(a []parser.RGBA) []parser.RGBA {
	n := len(a)
	out := make([]parser.RGBA, n)
	for i := range a {
		out[n-1-i] = a[i]
	}
	return out
}

func reverseFloats(a []pr.Fl) []pr.Fl {
	n := len(a)
	out := make([]pr.Fl, n)
	for i := range a {
		out[n-1-i] = a[i]
	}
	return out
}

func (self LinearGradient) Layout(width, height pr.Float) backend.GradientLayout {
	// Only one color, render the gradient as a solid color
	if len(self.colors) == 1 {
		return backend.GradientLayout{ScaleY: 1, GradientInit: backend.GradientInit{Kind: "solid"}, Colors: []parser.RGBA{self.colors[0]}}
	}
	// (dx, dy) is the unit vector giving the direction of the gradient.
	// Positive dx: right, positive dy: down.
	var dx, dy pr.Fl
	if self.direction.Corner != "" {
		var factorX, factorY pr.Float
		switch self.direction.Corner {
		case "top_left":
			factorX, factorY = -1, -1
		case "top_right":
			factorX, factorY = 1, -1
		case "bottom_left":
			factorX, factorY = -1, 1
		case "bottom_right":
			factorX, factorY = 1, 1
		}
		diagonal := pr.Hypot(width, height)
		// Note the direction swap: dx based on height, dy based on width
		// The gradient line is perpendicular to a diagonal.
		dx = pr.Fl(factorX * height / diagonal)
		dy = pr.Fl(factorY * width / diagonal)
	} else {
		angle := float64(self.direction.Angle) // 0 upwards, then clockwise
		dx = pr.Fl(math.Sin(angle))
		dy = pr.Fl(-math.Cos(angle))
	}

	// Round dx and dy to avoid floating points errors caused by
	// trigonometry and angle units conversions
	dx, dy = utils.Round(dx), utils.Round(dy)

	// Distance between center && ending point,
	// ie. half of between the starting point && ending point :
	colors := self.colors
	vectorLength := pr.Fl(pr.Abs(width*pr.Float(dx)) + pr.Abs(height*pr.Float(dy)))
	positions := processColorStops(pr.Float(vectorLength), self.stopPositions)
	if !self.repeating {
		// Add explicit colors at boundaries if needed, because PDF doesn’t
		// extend color stops that are not displayed
		if positions[0] == positions[1] {
			positions = append([]pr.Fl{positions[0] - 1}, positions...)
			colors = append([]parser.RGBA{colors[0]}, colors...)
		}
		if positions[len(positions)-2] == positions[len(positions)-1] {
			positions = append(positions, positions[len(positions)-1]+1)
			colors = append(colors, colors[len(colors)-1])
		}
	}
	first, last := normalizeStopPostions(positions)
	if self.repeating {
		// Render as a solid color if the first and last positions are equal
		// See https://drafts.csswg.org/css-images-3/#repeating-gradients
		if first == last {
			color := gradientAverageColor(colors, positions)
			return backend.GradientLayout{ScaleY: 1, GradientInit: backend.GradientInit{Kind: "solid"}, Colors: []parser.RGBA{color}}
		}

		// Define defined gradient length and steps between positions
		stopLength := last - first
		// assert stopLength > 0
		positionSteps := make([]pr.Fl, len(positions)-1)
		for i := range positionSteps {
			positionSteps[i] = positions[i+1] - positions[i]
		}

		// Create cycles used to add colors
		nextSteps := append([]pr.Fl{0}, positionSteps...)
		nextColors := colors
		previousSteps := append([]pr.Fl{0}, reverseFloats(positionSteps)...)
		previousColors := reverseColors(colors)

		// Add colors after last step
		for i := 0; last < vectorLength; i++ {
			step := nextSteps[i%len(nextSteps)]
			colors = append(colors, nextColors[i%len(nextColors)])
			positions = append(positions, positions[len(positions)-1]+step)
			last += step * stopLength
		}

		// Add colors before last step
		for i := 0; first > 0; i++ {
			step := previousSteps[i%len(previousSteps)]
			colors = append([]parser.RGBA{previousColors[i%len(previousColors)]}, colors...)
			positions = append([]pr.Fl{positions[0] - step}, positions...)
			first -= step * stopLength
		}
	}

	startX := (pr.Fl(width) - dx*vectorLength) / 2
	startY := (pr.Fl(height) - dy*vectorLength) / 2
	points := [6]pr.Fl{startX + dx*first, startY + dy*first, startX + dx*last, startY + dy*last, 0, 0}
	return backend.GradientLayout{ScaleY: 1, GradientInit: backend.GradientInit{Kind: "linear", Data: points}, Positions: positions, Colors: colors}
}

type RadialGradient struct {
	gradient
	shape  string
	size   pr.GradientSize
	center pr.Center
}

func (RadialGradient) isImage() {}

func NewRadialGradient(from pr.RadialGradient) RadialGradient {
	self := RadialGradient{gradient: newGradient(from.ColorStops, from.Repeating)}
	self.layouter = self
	//  Type of ending shape: "circle" || "ellipse"
	self.shape = from.Shape
	// sizeType: "keyword"
	//   size: "closest-corner", "farthest-corner",
	//         "closest-side", || "farthest-side"
	// sizeType: "explicit"
	//   size: (radiusX, radiusY)
	self.size = from.Size
	// Center of the ending shape. (originX, posX, originY, posY)
	self.center = from.Center
	return self
}

func handleDegenerateRadial(sizeX, sizeY pr.Float) (pr.Float, pr.Float) {
	// http://dev.w3.org/csswg/css-images-3/#degenerate-radials
	if sizeX == 0 && sizeY == 0 {
		sizeX = 1e-7
		sizeY = 1e-7
	} else if sizeX == 0 {
		sizeX = 1e-7
		sizeY = 1e7
	} else if sizeY == 0 {
		sizeX = 1e7
		sizeY = 1e-7
	}
	return sizeX, sizeY
}

func (self RadialGradient) Layout(width, height pr.Float) backend.GradientLayout {
	if len(self.colors) == 1 {
		return backend.GradientLayout{ScaleY: 1, GradientInit: backend.GradientInit{Kind: "solid"}, Colors: []parser.RGBA{self.colors[0]}}
	}
	originX, centerX_, originY, centerY_ := self.center.OriginX, self.center.Pos[0], self.center.OriginY, self.center.Pos[1]
	centerX := pr.ResoudPercentage(centerX_.ToValue(), width).V()
	centerY := pr.ResoudPercentage(centerY_.ToValue(), height).V()
	if originX == "right" {
		centerX = width - centerX
	}
	if originY == "bottom" {
		centerY = height - centerY
	}

	sizeX, sizeY := handleDegenerateRadial(self.resolveSize(width, height, centerX, centerY))
	scaleY := pr.Fl(sizeY / sizeX)

	colors := self.colors
	positions := processColorStops(sizeX, self.stopPositions)
	if !self.repeating {
		// Add explicit colors at boundaries if needed, because PDF doesn’t
		// extend color stops that are not displayed
		if positions[0] > 0 && positions[0] == positions[1] {
			positions = append([]pr.Fl{0}, positions...)
			colors = append([]parser.RGBA{colors[0]}, colors...)
		}
		if positions[len(positions)-2] == positions[len(positions)-1] {
			positions = append(positions, positions[len(positions)-1]+1)
			colors = append(colors, colors[len(colors)-1])
		}
	}

	if positions[0] < 0 {
		// PDF does not like negative radiuses,
		// shift into the positive realm.
		if self.repeating {
			// Add vector lengths to first position until positive
			vectorLength := positions[len(positions)-1] - positions[0]
			offset := vectorLength * pr.Fl(1+math.Floor(float64(-positions[0]/vectorLength)))
			for i, p := range positions {
				positions[i] = p + offset
			}
		} else {
			// only keep colors with position >= 0, interpolate if needed
			if positions[len(positions)-1] <= 0 {
				// All stops are negatives,
				// everything is "padded" with the last color.
				return backend.GradientLayout{ScaleY: 1, GradientInit: backend.GradientInit{Kind: "solid"}, Colors: []parser.RGBA{self.colors[len(self.colors)-1]}}
			}

			for i, position := range positions {
				if position == 0 {
					// Keep colors and positions from this rank
					colors, positions = colors[i:], positions[i:]
					break
				}

				if position > 0 {
					// Interpolate with the previous to get the color at 0.
					color := colors[i]
					negColor := colors[i-1]
					negPosition := positions[i-1]
					if negPosition >= 0 {
						panic(fmt.Sprintf("expected non positive negPosition, got %f", negPosition))
					}
					intermediateColor := gradientAverageColor(
						[]Color{negColor, negColor, color, color},
						[]pr.Fl{negPosition, 0, 0, position})
					colors = append([]Color{intermediateColor}, colors[i:]...)
					positions = append([]pr.Fl{0}, positions[i:]...)
					break
				}
			}

		}
	}

	first, last := normalizeStopPostions(positions)

	// Render as a solid color if the first and last positions are the same
	// See https://drafts.csswg.org/css-images-3/#repeating-gradients
	if first == last && self.repeating {
		color := gradientAverageColor(colors, positions)
		return backend.GradientLayout{ScaleY: 1, GradientInit: backend.GradientInit{Kind: "solid"}, Colors: []parser.RGBA{color}}
	}

	// Define the coordinates of the gradient circles
	circles := [6]pr.Fl{pr.Fl(centerX), pr.Fl(centerY) / scaleY, first, pr.Fl(centerX), pr.Fl(centerY) / scaleY, last}

	if self.repeating {
		circles, positions, colors = self.repeat(width, height, pr.Float(scaleY), circles, positions, colors)
	}

	return backend.GradientLayout{ScaleY: scaleY, GradientInit: backend.GradientInit{Kind: "radial", Data: circles}, Positions: positions, Colors: colors}
}

func (r RadialGradient) repeat(width, height, scaleY pr.Float, points [6]pr.Fl, positions []pr.Fl, colors []parser.RGBA) ([6]pr.Fl, []pr.Fl, []parser.RGBA) {
	// Keep original lists and values, they’re useful
	originalColors := append([]parser.RGBA{}, colors...)
	originalPositions := append([]pr.Fl{}, positions...)
	gradientLength := points[5] - points[2]

	// Get the maximum distance between the center && the corners, to find
	// how many times we have to repeat the colors outside
	maxDistance := pr.Fl(pr.Maxs(
		pr.Hypot(width-pr.Float(points[0]), height/scaleY-pr.Float(points[1])),
		pr.Hypot(width-pr.Float(points[0]), -pr.Float(points[1])*scaleY),
		pr.Hypot(-pr.Float(points[0]), height/scaleY-pr.Float(points[1])),
		pr.Hypot(-pr.Float(points[0]), -pr.Float(points[1])*scaleY),
	))
	repeatAfter := int(math.Ceil(float64((maxDistance - points[5]) / gradientLength)))
	if repeatAfter > 0 {
		// Repeat colors and extrapolate positions
		repeat := 1 + repeatAfter
		colors = make([]parser.RGBA, len(colors)*repeat)
		tmpPositions := make([]pr.Fl, 0, len(positions)*repeat)
		for i := 0; i < repeat; i++ {
			copy(colors[i*len(originalColors):], originalColors)
			for _, position := range positions {
				tmpPositions = append(tmpPositions, pr.Fl(i)+position)
			}
		}
		positions = tmpPositions
		points[5] = points[5] + gradientLength*pr.Fl(repeatAfter)
	}

	if points[2] == 0 {
		// Inner circle has 0 radius, no need to repeat inside, return
		return points, positions, colors
	}

	// Find how many times we have to repeat the colors inside
	repeatBefore := points[2] / gradientLength

	// Set the inner circle size to 0
	points[2] = 0

	// Find how many times the whole gradient can be repeated
	fullRepeat := int(repeatBefore)
	if fullRepeat != 0 {
		// Repeat colors and extrapolate positions
		positionsTmp := positions
		positions = make([]pr.Fl, 0, len(positionsTmp)+len(originalPositions)*fullRepeat)
		for i := 0; i < fullRepeat; i++ {
			colors = append(colors, originalColors...)
			for _, position := range originalPositions {
				positions = append(positions, pr.Fl(i-fullRepeat)+position)
			}
		}
		positions = append(positions, positionsTmp...)
	}

	// Find the ratio of gradient that must be added to reach the center
	partialRepeat := repeatBefore - pr.Fl(fullRepeat)
	if partialRepeat == 0 {
		// No partial repeat, return
		return points, positions, colors
	}

	// Iterate through positions := range reverse order, from the outer
	// circle to the original inner circle, to find positions from
	// the inner circle (including full repeats) to the center
	// assert (originalPositions[0], originalPositions[-1]) == (0, 1)
	// assert 0 < partialRepeat < 1
	reverse := reverseFloats(originalPositions)
	ratio := 1 - partialRepeat
	LC, LP := len(originalColors), len(originalPositions)

	for i_, position := range reverse {
		i := i_ + 1
		if position == ratio {
			// The center is a color of the gradient, truncate original
			// colors and positions and prepend them
			colors = append(originalColors[LC-i:], colors...)
			tmp := originalPositions[LP-i:]
			newPositions := make([]pr.Fl, len(tmp))
			for j, position := range tmp {
				newPositions[j] = position - pr.Fl(fullRepeat) - 1
			}
			positions = append(newPositions, positions...)
			return points, positions, colors
		}
		if position < ratio {
			// The center is between two colors of the gradient,
			// define the center color as the average of these two
			// gradient colors
			color := originalColors[LC-i]
			nextColor := originalColors[LC-(i-1)]
			nextPosition := originalPositions[LP-(i-1)]
			averageColors := []parser.RGBA{color, color, nextColor, nextColor}
			averagePositions := []pr.Fl{position, ratio, ratio, nextPosition}
			zeroColor := gradientAverageColor(averageColors, averagePositions)
			colors = append(append([]parser.RGBA{zeroColor}, originalColors[LC-(i-1):]...), colors...)
			tmp := originalPositions[LP-(i-1):]
			newPositions := make([]pr.Fl, len(tmp))
			for j, position := range tmp {
				newPositions[j] = position - pr.Fl(fullRepeat) - 1
			}
			positions = append(append([]pr.Fl{ratio - 1 - pr.Fl(fullRepeat)}, newPositions...), positions...)
			return points, positions, colors
		}
	}

	return points, positions, colors
}

func (self RadialGradient) resolveSize(width, height, centerX, centerY pr.Float) (pr.Float, pr.Float) {
	if self.size.IsExplicit() {
		sizeX, sizeY := self.size.Explicit[0], self.size.Explicit[1]
		sizeX_ := pr.ResoudPercentage(sizeX.ToValue(), width).V()
		sizeY_ := pr.ResoudPercentage(sizeY.ToValue(), height).V()
		return sizeX_, sizeY_
	}
	left := pr.Abs(centerX)
	right := pr.Abs(width - centerX)
	top := pr.Abs(centerY)
	bottom := pr.Abs(height - centerY)
	pick := pr.Maxs
	if strings.HasPrefix(self.size.Keyword, "closest") {
		pick = pr.Mins
	}
	if strings.HasSuffix(self.size.Keyword, "side") {
		if self.shape == "circle" {
			sizeXy := pick(left, right, top, bottom)
			return sizeXy, sizeXy
		}
		// else: ellipse
		return pick(left, right), pick(top, bottom)
	}
	// else: corner
	if self.shape == "circle" {
		sizeXy := pick(pr.Hypot(left, top), pr.Hypot(left, bottom),
			pr.Hypot(right, top), pr.Hypot(right, bottom))
		return sizeXy, sizeXy
	}
	// else: ellipse
	keys := [4]pr.Float{pr.Hypot(left, top), pr.Hypot(left, bottom), pr.Hypot(right, top), pr.Hypot(right, bottom)}
	m := map[pr.Float][2]pr.Float{
		keys[0]: {left, top},
		keys[1]: {left, bottom},
		keys[2]: {right, top},
		keys[3]: {right, bottom},
	}
	c := m[pick(keys[0], keys[1], keys[2], keys[3])]
	cornerX, cornerY := c[0], c[1]
	return cornerX * pr.Float(math.Sqrt(2)), cornerY * pr.Float(math.Sqrt(2))
}
