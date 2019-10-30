package layout

import (
	"log"

	bo "github.com/benoitkugler/go-weasyprint/boxes"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
)

// Resolve percentages into fixed values.

// Return the percentage of the reference value, or the value unchanged.
// ``referTo`` is the length for 100%. If ``referTo`` is not a number, it
// just replaces percentages.
func percentage(value pr.Value, referTo pr.Float) pr.MaybeFloat {
	if value.IsNone() {
		return nil
	} else if value.String == "auto" {
		return pr.Auto
	} else if value.Unit == pr.Px {
		return value.Value
	} else {
		if value.Unit != pr.Percentage {
			log.Fatalf("expected percentage, got %d", value.Unit)
		}
		return referTo * value.Value / 100.
	}
}

// Compute a used length value from a computed length value.
//
// the return value should be set on the box
func resolveOnePercentage(value pr.Value, propertyName string, referTo pr.Float, mainFlexDirection string) pr.MaybeFloat {
	// box.style has computed values
	// value := box.Style[propertyName]

	// box attributes are used values
	percent := percentage(value, referTo)
	// setattr(box, propertyName, percent)
	if (propertyName == "minWidth" || propertyName == "minHeight") && percent == pr.Auto {
		if mainFlexDirection == "" || propertyName != "min"+mainFlexDirection {
			percent = pr.Float(0)
		}
	}
	return percent
}

func resolvePositionPercentages(box *bo.BoxFields, containingBlock bo.Point) {
	cbWidth, cbHeight := containingBlock[0], containingBlock[1]
	box.Left = resolveOnePercentage(box.Style.GetLeft(), "left", cbWidth, "")
	box.Right = resolveOnePercentage(box.Style.GetRight(), "right", cbWidth, "")
	box.Top = resolveOnePercentage(box.Style.GetTop(), "top", cbHeight, "")
	box.Bottom = resolveOnePercentage(box.Style.GetBottom(), "bottom", cbHeight, "")
}

func resolvePercentages2(box Box, containingBlock bo.BoxFields, mainFlexDirection string) {
	resolvePercentages(box, bo.MaybePoint{containingBlock.Width, containingBlock.Height}, mainFlexDirection)
}

// Set used values as attributes of the box object.
func resolvePercentages(box_ Box, containingBlock bo.MaybePoint, mainFlexDirection string) {
	cbWidth, cbHeight := containingBlock[0], containingBlock[1]
	maybeHeight := cbWidth
	if bo.IsPageBox(box_) {
		maybeHeight = cbHeight
	}
	box := box_.Box()
	box.MarginLeft = resolveOnePercentage(box.Style.GetMarginLeft(), "margin_left", cbWidth.V(), "")
	box.MarginRight = resolveOnePercentage(box.Style.GetMarginRight(), "margin_right", cbWidth.V(), "")
	box.MarginTop = resolveOnePercentage(box.Style.GetMarginTop(), "margin_top", maybeHeight.V(), "")
	box.MarginBottom = resolveOnePercentage(box.Style.GetMarginBottom(), "margin_bottom", maybeHeight.V(), "")
	box.PaddingLeft = resolveOnePercentage(box.Style.GetPaddingLeft(), "padding_left", cbWidth.V(), "")
	box.PaddingRight = resolveOnePercentage(box.Style.GetPaddingRight(), "padding_right", cbWidth.V(), "")
	box.PaddingTop = resolveOnePercentage(box.Style.GetPaddingTop(), "padding_top", maybeHeight.V(), "")
	box.PaddingBottom = resolveOnePercentage(box.Style.GetPaddingBottom(), "padding_bottom", maybeHeight.V(), "")
	box.Width = resolveOnePercentage(box.Style.GetWidth(), "width", cbWidth.V(), "")
	box.MinWidth = resolveOnePercentage(box.Style.GetMinWidth(), "min_width", cbWidth.V(), mainFlexDirection)
	box.MaxWidth = resolveOnePercentage(box.Style.GetMaxWidth(), "max_width", cbWidth.V(), mainFlexDirection)

	// XXX later: top, bottom, left && right on positioned elements

	if cbHeight == pr.Auto {
		// Special handling when the height of the containing block
		// depends on its content.
		height := box.Style.GetHeight()
		if height.String == "auto" || height.Unit == pr.Percentage {
			box.Height = pr.Auto
		} else {
			if height.Unit != pr.Px {
				log.Fatalf("expected percentage, got %d", height.Unit)
			}
			box.Height = height.Value
		}
		box.MinHeight = resolveOnePercentage(box.Style.GetMinHeight(), "min_height", pr.Float(0), mainFlexDirection)
		box.MaxHeight = resolveOnePercentage(box.Style.GetMaxHeight(), "max_height", pr.Inf, mainFlexDirection)
	} else {
		box.Height = resolveOnePercentage(box.Style.GetHeight(), "height", cbHeight.V(), "")
		box.MinHeight = resolveOnePercentage(box.Style.GetMinHeight(), "min_height", cbHeight.V(), mainFlexDirection)
		box.MaxHeight = resolveOnePercentage(box.Style.GetMaxHeight(), "max_height", cbHeight.V(), mainFlexDirection)
	}

	// Used value == computed value
	box.BorderTopWidth = box.Style.GetTop().ToMaybeFloat()
	box.BorderRightWidth = box.Style.GetRight().ToMaybeFloat()
	box.BorderBottomWidth = box.Style.GetBottom().ToMaybeFloat()
	box.BorderLeftWidth = box.Style.GetLeft().ToMaybeFloat()

	// Shrink *content* widths and heights according to box-sizing
	// Thanks heavens and the spec: Our validator rejects negative values
	// for padding and border-width
	var horizontalDelta, verticalDelta pr.Float
	switch box.Style.GetBoxSizing() {
	case "border-box":
		horizontalDelta = box.PaddingLeft.V() + box.PaddingRight.V() + box.BorderLeftWidth.V() + box.BorderRightWidth.V()
		verticalDelta = box.PaddingTop.V() + box.PaddingBottom.V() + box.BorderTopWidth.V() + box.BorderBottomWidth.V()
	case "padding-box":
		horizontalDelta = box.PaddingLeft.V() + box.PaddingRight.V()
		verticalDelta = box.PaddingTop.V() + box.PaddingBottom.V()
	case "content-box":
		horizontalDelta = 0
		verticalDelta = 0
	default:
		log.Fatalf("invalid box sizing %s", box.Style.GetBoxSizing())
	}

	// Keep at least min* >= 0 to prevent funny output in case box.Width or
	// box.Height become negative.
	// Restricting max* seems reasonable, too.
	if horizontalDelta > 0 {
		if box.Width != pr.Auto {
			box.Width = pr.Max(0, box.Width.V()-horizontalDelta)
		}
		box.MaxWidth = pr.Max(0, box.MaxWidth.V()-horizontalDelta)
		if box.MinWidth != pr.Auto {
			box.MinWidth = pr.Max(0, box.MinWidth.V()-horizontalDelta)
		}
	}
	if verticalDelta > 0 {
		if box.Height != pr.Auto {
			box.Height = pr.Max(0, box.Height.V()-verticalDelta)
		}
		box.MaxHeight = pr.Max(0, box.MaxHeight.V()-verticalDelta)
		if box.MinHeight != pr.Auto {
			box.MinHeight = pr.Max(0, box.MinHeight.V()-verticalDelta)
		}
	}
}

func resoudRadius(box *bo.BoxFields, v pr.Point) bo.MaybePoint {
	rx := percentage(v[0].ToValue(), box.BorderWidth())
	ry := percentage(v[1].ToValue(), box.BorderHeight())
	return bo.MaybePoint{rx, ry}
}

func resolveRadiiPercentages(box *bo.BoxFields) {
	box.BorderTopLeftRadius = resoudRadius(box, box.Style.GetBorderTopLeftRadius())
	box.BorderTopRightRadius = resoudRadius(box, box.Style.GetBorderTopRightRadius())
	box.BorderBottomRightRadius = resoudRadius(box, box.Style.GetBorderBottomRightRadius())
	box.BorderBottomLeftRadius = resoudRadius(box, box.Style.GetBorderBottomLeftRadius())
}
