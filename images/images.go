// Fetch and decode images in range various formats.
package images

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/benoitkugler/go-weasyprint/backend"
	"github.com/benoitkugler/go-weasyprint/style/parser"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/utils"
	"golang.org/x/net/html"
)

type Color = parser.RGBA

// Image is the common interface for supported image format,
// such as gradient, SVG, or JPEG, PNG, etc...
type Image interface {
	isImage()
	GetIntrinsicSize(imageResolution, fontSize pr.Value) (width, height pr.MaybeFloat)
	IntrinsicRatio() pr.MaybeFloat
	Draw(context backend.Drawer, concreteWidth, concreteHeight float64, imageRendering pr.String)
}

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
	imageSurface    imageSurface
	intrinsicRatio  pr.Float
	intrinsicWidth  pr.Float
	intrinsicHeight pr.Float
}

func NewRasterImage(imageSurface imageSurface) RasterImage {
	self := RasterImage{}
	self.imageSurface = imageSurface
	self.intrinsicWidth = imageSurface.getWidth()
	self.intrinsicHeight = imageSurface.getHeight()
	self.intrinsicRatio = pr.Inf
	if self.intrinsicHeight != 0 {
		self.intrinsicRatio = self.intrinsicWidth / self.intrinsicHeight
	}
	return self
}

func (r RasterImage) isImage() {}

func (r RasterImage) IntrinsicRatio() pr.MaybeFloat { return r.intrinsicRatio }

func (r RasterImage) GetIntrinsicSize(imageResolution, _ pr.Value) (width, height pr.MaybeFloat) {
	// Raster images are affected by the "image-resolution" property.
	return r.intrinsicWidth / imageResolution.Value, r.intrinsicHeight / imageResolution.Value
}

func (r RasterImage) Draw(context backend.Drawer, concreteWidth, concreteHeight pr.Float, imageRendering string) {
	hasSize := concreteWidth > 0 && concreteHeight > 0 && r.intrinsicWidth > 0 && r.intrinsicHeight > 0
	if !hasSize {
		return
	}

	// Use the real intrinsic size here,
	// not affected by "image-resolution".
	context.Scale(float64(concreteWidth/r.intrinsicWidth), float64(concreteHeight/r.intrinsicHeight))
	// FIXME:
	// context.setSourceSurface(r.imageSurface)
	// context.getSource().setFilter(imagesRenderingToFilter[imageRendering])
	// context.paint()
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
	intrinsicWidth  pr.MaybeFloat
	intrinsicHeight pr.MaybeFloat
	intrinsicRatio  pr.MaybeFloat
	tree            *utils.HTMLNode
	urlFetcher      utils.UrlFetcher
	baseUrl         string
	svgData         string
	width           float64
	height          float64
}

func (SVGImage) isImage() {}

func NewSVGImage(svgData string, baseUrl string, urlFetcher utils.UrlFetcher) (*SVGImage, error) {
	self := new(SVGImage)
	// Don’t pass data URIs to CairoSVG.
	// They are useless for relative URIs anyway.
	if !strings.HasPrefix(strings.ToLower(baseUrl), "data:") {
		self.baseUrl = baseUrl
	}
	self.svgData = svgData
	self.urlFetcher = urlFetcher
	tree, err := html.Parse(strings.NewReader(self.svgData))
	if err != nil {
		return nil, imageLoadingError(err)
	}
	self.tree = (*utils.HTMLNode)(tree)
	return self, nil
}

func (s SVGImage) IntrinsicRatio() pr.MaybeFloat {
	return s.intrinsicRatio
}

//  func CairosvgUrlFetcher(self, src, mimetype) {
//         data = self.urlFetcher(src)
//         if "string" := range data {
//             return data["string"]
//         } return data["fileObj"].read()
//     }

func (s *SVGImage) GetIntrinsicSize(_, fontSize pr.Value) (width, height pr.MaybeFloat) {
	// Vector images may be affected by the font size.
	fakeSurface := newFakeSurface()
	fakeSurface.fontSize = float64(fontSize.Value)
	// Percentages don't provide an intrinsic size, we transform percentages
	// into 0 using a (0, 0) context size :
	// http://www.w3.org/TR/SVG/coords.html#IntrinsicSizing
	s.width = size(fakeSurface, s.tree.Get("width"), floatOrString{s: "xy"})
	s.height = size(fakeSurface, s.tree.Get("height"), floatOrString{s: "xy"})
	_, _, viewbox := nodeFormat(fakeSurface, s.tree, true)
	s.intrinsicWidth = nil
	if s.width != 0 {
		s.intrinsicWidth = pr.Float(s.width)
	}
	s.intrinsicHeight = nil
	if s.height != 0 {
		s.intrinsicHeight = pr.Float(s.height)
	}
	s.intrinsicRatio = nil
	if len(viewbox) != 0 {
		if s.width != 0 && s.height != 0 {
			s.intrinsicRatio = pr.Float(s.width / s.height)
		} else {
			if viewbox[2] != 0 && viewbox[3] != 0 {
				s.intrinsicRatio = pr.Float(viewbox[2] / viewbox[3])
				if s.width != 0 {
					s.intrinsicHeight = pr.Float(s.width) / s.intrinsicRatio.V()
				} else if s.height != 0 {
					s.intrinsicWidth = pr.Float(s.height) * s.intrinsicRatio.V()
				}
			}
		}
	} else if s.width != 0 && s.height != 0 {
		s.intrinsicRatio = pr.Float(s.width / s.height)
	}
	return s.intrinsicWidth, s.intrinsicHeight
}

func (SVGImage) Draw(context backend.Drawer, concreteWidth, concreteHeight float64, imageRendering pr.String) {
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

// Get a cairo Pattern from an image URI.
func GetImageFromUri(cache map[string]Image, fetcher utils.UrlFetcher, url, _ string) Image {
	image, in := cache[url]
	if in {
		return image
	}
	_, err := fetcher(url)
	if err != nil {
		log.Printf(`Failed to load image at "%s" (%s)`, url, err)
		return nil
	}
	// FIXME: à implémenter
	image = RadialGradient{}
	//     try {
	//         with fetch(urlFetcher, url) as result {
	//             if "string" := range result {
	//                 string = result["string"]
	//             } else {
	//                 string = result["fileObj"].read()
	//             } mimeType = forcedMimeType || result["mimeType"]
	//             if mimeType == "image/svg+xml" {
	//                 // No fallback for XML-based mimetypes as defined by MIME
	//                 // Sniffing Standard, see https://mimesniff.spec.whatwg.org/
	//                 image = SVGImage(string, url, urlFetcher)
	//             } else {
	//                 // Try to rely on given mimetype
	//                 try {
	//                     if mimeType == "image/png" {
	//                         try {
	//                             surface = cairocffi.ImageSurface.createFromPng(
	//                                 BytesIO(string))
	//                         } except Exception as exception {
	//                             raise ImageLoadingError.fromException(exception)
	//                         } else {
	//                             image = RasterImage(surface)
	//                         }
	//                     } else {
	//                         image = nil
	//                     }
	//                 } except ImageLoadingError {
	//                     image = nil
	//                 }
	//             }
	//         }
	//     }

	//                 // Relying on mimetype didn"t work, give the image to GDK-Pixbuf
	//                 if ! image {
	//                     if pixbuf  == nil  {
	//                         raise ImageLoadingError(
	//                             "Could ! load GDK-Pixbuf. PNG && SVG are "
	//                             "the only image formats available.")
	//                     } try {
	//                         image = SVGImage(string, url, urlFetcher)
	//                     } except BaseException {
	//                         try {
	//                             surface, formatName = (
	//                                 pixbuf.decodeToImageSurface(string))
	//                         } except pixbuf.ImageLoadingError as exception {
	//                             raise ImageLoadingError(str(exception))
	//                         } if formatName == "jpeg" {
	//                             surface.setMimeData("image/jpeg", string)
	//                         } image = RasterImage(surface)
	//                     }
	//                 }
	//     except (URLFetchingError, ImageLoadingError) as exc {
	//         LOGGER.error("Failed to load image at "%s" (%s)", url, exc)
	//         image = nil
	//     }
	cache[url] = image
	return image
}

// Gradient line size: distance between the starting point and ending point.
// Positions: list of Dimension in px or % (possibliy zero)
// 0 is the starting point, 1 the ending point.
// http://dev.w3.org/csswg/css-images-3/#color-stop-syntax
// Return processed color stops, as a list of floats in px.
func processColorStops(gradientLineSize pr.Float, positions_ []pr.Dimension) []pr.Float {
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
	out := make([]pr.Float, L)
	for i, v := range positions {
		out[i] = v.V()
	}
	return out
}

// Normalize to [0..1].
// Write on positions.
func normalizeStopPostions(positions []pr.Float) (pr.Float, pr.Float) {
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
func gradientAverageColor(colors []Color, positions []pr.Float) Color {
	nbStops := len(positions)
	if nbStops <= 1 || nbStops != len(colors) {
		log.Fatalf("expected same length, at least 2, got %d, %d", nbStops, len(colors))
	}
	totalLength := positions[nbStops-1] - positions[0]
	if totalLength == 0 {
		for i := range positions {
			positions[i] = pr.Float(i)
		}
		totalLength = pr.Float(nbStops - 1)
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
	for i, position := range positions[1:] {
		i = i + 1
		weight := utils.Fl((position - positions[i-1]) / totalWeight)
		for j := i - 1; j < i; j += 1 {
			resultR += premulR[j] * weight
			resultG += premulG[j] * weight
			resultB += premulB[j] * weight
			resultA += alpha[j] * weight
		}
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

var patternTypes = map[string]int{
	"linear": 0, // cairocffi.LinearGradient,
	"radial": 1, // cairocffi.RadialGradient,
	"solid":  2, // cairocffi.SolidPattern,
}

type gradientInit struct {
	// type_ is either:
	// 	"solid": init is (r, g, b, a). positions && colors are empty.
	// 	"linear": init is (x0, y0, x1, y1)
	// 			  coordinates of the starting && ending points.
	// 	"radial": init is (cx0, cy0, radius0, cx1, cy1, radius1)
	// 			  coordinates of the starting end ending circles
	type_ string
	data  [6]pr.Float
}

type gradientLayout struct {
	//  list of floats in [0..1].
	// 0 at the starting point, 1 at the ending point.
	positions []pr.Float
	colors    []Color
	init      gradientInit

	// used for ellipses radial gradients. 1 otherwise.
	scaleY pr.Float
}

type layouter interface {
	// width, height: Gradient box. Top-left is at coordinates (0, 0).
	// userToDeviceDistance: a (dx, dy) -> (ddx, ddy) function
	layout(width, height pr.Float, userToDeviceDistance func(dx, dy pr.Float) (ddx, ddy pr.Float)) gradientLayout
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

func (g gradient) GetIntrinsicSize(_, _ pr.Value) (pr.MaybeFloat, pr.MaybeFloat) {
	// Gradients are not affected by image resolution, parent or font size.
	return nil, nil
}

func (g gradient) IntrinsicRatio() pr.MaybeFloat {
	return nil
}

func (g gradient) Draw(context backend.Drawer, concreteWidth, concreteHeight float64, imageRendering pr.String) {
	// FIXME:
	log.Println("drawing gradient...")
	//  func draw(self, context, concreteWidth, concreteHeight, ImageRendering) {
	//         scaleY, type_, init, stopPositions, stopColors = self.layout(
	//             concreteWidth, concreteHeight, context.userToDeviceDistance)
	//         context.scale(1, scaleY{
	//         pattern = patternTypes[tymap[string]int*init)
	// :       for position, color := range zip(stopPositions, stopColors) {:			pattern.addColorStop:c
	// }Rgba(position, *color)
	//         } pattern.setExtend(cairocffi.EXTENDREPEAT if self.repeating
	//                            else cairocffi.EXTENDPAD)
	//         context.setSource(pattern)
	//         context.paint()
	//     }
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

func toData(c Color) [6]pr.Float {
	return [6]pr.Float{pr.Float(c.R), pr.Float(c.G), pr.Float(c.B), pr.Float(c.A), 0, 0}
}

func (self LinearGradient) layout(width, height pr.Float, userToDeviceDistance func(dx, dy pr.Float) (ddx, ddy pr.Float)) gradientLayout {
	if len(self.colors) == 1 {
		return gradientLayout{scaleY: 1, init: gradientInit{type_: "solid", data: toData(self.colors[0])}}
	}
	// (dx, dy) is the unit vector giving the direction of the gradient.
	// Positive dx: right, positive dy: down.
	var dx, dy pr.Float
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
		dx = factorX * height / diagonal
		dy = factorY * width / diagonal
	} else {
		angle := float64(self.direction.Angle) // 0 upwards, then clockwise
		dx = pr.Float(math.Sin(angle))
		dy = pr.Float(-math.Cos(angle))
	}
	// Distance between center && ending point,
	// ie. half of between the starting point && ending point :
	distance := pr.Abs(width*dx) + pr.Abs(height*dy)
	positions := processColorStops(distance, self.stopPositions)
	first, last := normalizeStopPostions(positions)

	devicePerUserUnits := pr.Hypot(userToDeviceDistance(dx, dy))
	if (last-first)*devicePerUserUnits < pr.Float(len(positions)) {
		if self.repeating {
			color := gradientAverageColor(self.colors, positions)
			return gradientLayout{scaleY: 1, init: gradientInit{type_: "solid", data: toData(color)}}
		} else {
			// 100 is an Arbitrary non-zero number of device units.
			offset := 100 / devicePerUserUnits
			if first != last {
				factor := (offset + last - first) / (last - first)
				for i, pos := range positions {
					positions[i] = pos / factor
				}
			}
			last += offset
		}
	}
	startX := (width - dx*distance) / 2
	startY := (height - dy*distance) / 2
	points := [6]pr.Float{startX + dx*first, startY + dy*first, startX + dx*last, startY + dy*last, 0, 0}
	return gradientLayout{scaleY: 1, init: gradientInit{type_: "linear", data: points}, positions: positions, colors: self.colors}
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

func (self RadialGradient) layout(width, height pr.Float, userToDeviceDistance func(dx, dy pr.Float) (ddx, ddy pr.Float)) gradientLayout {
	if len(self.colors) == 1 {
		return gradientLayout{scaleY: 1, init: gradientInit{type_: "solid", data: toData(self.colors[0])}}
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

	sizeX, sizeY := self.resolveSize(width, height, centerX, centerY)
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
	scaleY := sizeY / sizeX

	colors := self.colors
	positions := processColorStops(sizeX, self.stopPositions)
	gradientLineSize := positions[len(positions)-1] - positions[0]

	unit1 := pr.Hypot(userToDeviceDistance(1, 0))
	unit2 := pr.Hypot(userToDeviceDistance(0, scaleY))
	if self.repeating && (gradientLineSize*unit1 < pr.Float(len(positions)) || gradientLineSize*unit2 < pr.Float(len(positions))) {
		color := gradientAverageColor(colors, positions)
		return gradientLayout{scaleY: 1, init: gradientInit{type_: "solid", data: toData(color)}}
	}
	if positions[0] < 0 {
		// Cairo does not like negative radiuses,
		// shift into the positive realm.
		if self.repeating {
			offset := gradientLineSize * pr.Float(math.Ceil(float64(-positions[0]/gradientLineSize)))
			for i, p := range positions {
				positions[i] = p + offset
			}
		} else {
			var hasBroken bool
			for i, position := range positions {
				if position > 0 {
					// `i` is the first positive stop.
					// Interpolate with the previous to get the color at 0.
					if i <= 0 {
						log.Fatalf("expected non zero i")
					}
					color := colors[i]
					negColor := colors[i-1]
					negPosition := positions[i-1]
					if negPosition >= 0 {
						log.Fatalf("expected non positive negPosition, got %f", negPosition)
					}
					intermediateColor := gradientAverageColor(
						[]Color{negColor, negColor, color, color},
						[]pr.Float{negPosition, 0, 0, position})
					colors = append([]Color{intermediateColor}, colors[i:]...)
					positions = append([]pr.Float{0}, positions[i:]...)
					hasBroken = true
					break
				}
			}
			if !hasBroken {
				// All stops are negatives,
				// everything is "padded" with the last color.
				return gradientLayout{scaleY: 1, init: gradientInit{type_: "solid", data: toData(self.colors[len(self.colors)-1])}}
			}
		}
	}

	first, last := normalizeStopPostions(positions)
	if last == first {
		last += 100 // Arbitrary non-zero
	}

	circles := [6]pr.Float{centerX, centerY / scaleY, first, centerX, centerY / scaleY, last}
	return gradientLayout{scaleY: scaleY, init: gradientInit{type_: "radial", data: circles}, positions: positions, colors: colors}
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
