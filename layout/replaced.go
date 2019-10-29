package layout

import pr "github.com/benoitkugler/go-weasyprint/style/properties"

// Layout for images and other replaced elements.
// http://dev.w3.org/csswg/css-images-3/#sizing

// Default sizing algorithm for the concrete object size.
// http://dev.w3.org/csswg/css-images-3/#default-sizing
func defaultImageSizing(intrinsicWidth, intrinsicHeight, intrinsicRatio float32,
	specifiedWidth, specifiedHeight pr.MaybeFloat, defaultWidth, defaultHeight float32) (concreteWidth, concreteHeight pr.Float) {

}
