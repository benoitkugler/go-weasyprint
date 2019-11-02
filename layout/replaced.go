package layout

import (
	"log"

	bo "github.com/benoitkugler/go-weasyprint/boxes"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
)

// Layout for images and other replaced elements.
// http://dev.w3.org/csswg/css-images-3/#sizing

// Default sizing algorithm for the concrete object size.
// http://dev.w3.org/csswg/css-images-3/#default-sizing
func defaultImageSizing(intrinsicWidth, intrinsicHeight, intrinsicRatio,
	specifiedWidth, specifiedHeight pr.MaybeFloat, defaultWidth, defaultHeight pr.Float) (concreteWidth, concreteHeight pr.Float) {

	if specifiedWidth == pr.Auto {
		specifiedWidth = nil
	}
	if specifiedHeight == pr.Auto {
		specifiedHeight = nil
	}

	if specifiedWidth != nil && specifiedHeight != nil {
		return specifiedWidth.V(), specifiedHeight.V()
	} else if specifiedWidth != nil {
		if intrinsicRatio != nil {
			concreteHeight = specifiedWidth.V() / intrinsicRatio.V()
		} else if intrinsicHeight != nil {
			concreteHeight = intrinsicHeight.V()
		} else {
			concreteHeight = defaultHeight
		}
		return specifiedWidth.V(), concreteHeight
	} else if specifiedHeight != nil {
		if intrinsicRatio != nil {
			concreteWidth = specifiedHeight.V() * intrinsicRatio.V()
		} else if intrinsicWidth != nil {
			concreteWidth = intrinsicWidth.V()
		} else {
			concreteWidth = defaultWidth
		}
		return concreteWidth, specifiedHeight.V()
	} else {
		if intrinsicWidth != nil || intrinsicHeight != nil {
			return defaultImageSizing(intrinsicWidth, intrinsicHeight, intrinsicRatio,
				intrinsicWidth, intrinsicHeight, defaultWidth, defaultHeight)
		} else {
			return containConstraintImageSizing(defaultWidth, defaultHeight, intrinsicRatio)
		}
	}
}

// Cover constraint sizing algorithm for the concrete object size.
// http://dev.w3.org/csswg/css-images-3/#contain-constraint
func containConstraintImageSizing(constraintWidth, constraintHeight pr.Float, intrinsicRatio pr.MaybeFloat) (concreteWidth, concreteHeight pr.Float) {
	return constraintImageSizing(constraintWidth, constraintHeight, intrinsicRatio, false)
}

// Cover constraint sizing algorithm for the concrete object size.
// http://dev.w3.org/csswg/css-images-3/#cover-constraint
func coverConstraintImageSizing(constraintWidth, constraintHeight pr.Float, intrinsicRatio pr.MaybeFloat) (concreteWidth, concreteHeight pr.Float) {
	return constraintImageSizing(constraintWidth, constraintHeight, intrinsicRatio, true)
}

func constraintImageSizing(constraintWidth, constraintHeight pr.Float, intrinsicRatio pr.MaybeFloat, cover bool) (concreteWidth, concreteHeight pr.Float) {
	if intrinsicRatio == nil {
		return constraintWidth, constraintHeight
	} else if cover != (constraintWidth > constraintHeight*intrinsicRatio.V()) {
		return constraintHeight * intrinsicRatio.V(), constraintHeight
	} else {
		return constraintWidth, constraintWidth / intrinsicRatio.V()
	}
}

func ReplacedboxLayout(box_ bo.InstanceReplacedBox) (drawWidth, drawHeight, positionX, positionY pr.Float) {
	box := box_.Replaced()
	// TODO: respect box-sizing ?
	objectFit := box.Style.GetObjectFit()
	position := box.Style.GetObjectPosition()

	image := box.Replacement
	intrinsicWidth, intrinsicHeight := image.GetIntrinsicSize(box.Style.GetImageResolution(), box.Style.GetFontSize())
	if intrinsicWidth == nil || intrinsicHeight == nil {
		intrinsicWidth, intrinsicHeight = containConstraintImageSizing(box.Width.V(), box.Height.V(), image.IntrinsicRatio())
	}

	if objectFit == "fill" {
		drawWidth, drawHeight = box.Width.V(), box.Height.V()
	} else {
		if objectFit == "contain" || objectFit == "scale-down" {
			drawWidth, drawHeight = containConstraintImageSizing(box.Width.V(), box.Height.V(), image.IntrinsicRatio())
		} else if objectFit == "cover" {
			drawWidth, drawHeight = coverConstraintImageSizing(box.Width.V(), box.Height.V(), image.IntrinsicRatio())
		} else if objectFit == "none" {
			drawWidth, drawHeight = intrinsicWidth.V(), intrinsicHeight.V()
		} else {
			log.Fatalf("unexpected objectFit %s", objectFit)
		}

		if objectFit == "scale-down" {
			drawWidth = pr.Min(drawWidth, intrinsicWidth.V())
			drawHeight = pr.Min(drawHeight, intrinsicHeight.V())
		}
	}

	originX, positionX_, originY, positionY_ := position.OriginX, position.Pos[0], position.OriginY, position.Pos[1]
	refX := box.Width.V() - drawWidth
	refY := box.Height.V() - drawHeight

	positionX = pr.ResoudPercentage(positionX_.ToValue(), refX).V()
	positionY = pr.ResoudPercentage(positionY_.ToValue(), refY).V()
	if originX == "right" {
		positionX = refX - positionX
	}
	if originY == "bottom" {
		positionY = refY - positionY
	}

	positionX += box.ContentBoxX()
	positionY += box.ContentBoxY()

	return drawWidth, drawHeight, positionX, positionY
}
