package layout

import (
	"log"
	"math"

	bo "github.com/benoitkugler/go-weasyprint/boxes"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/utils"
)

// Resolve percentages into fixed values.

// avoid to use interface or struct
const (
	Auto float32 = -17849812.1238415
	None float32 = -13465864.2187267
)

// Return the percentage of the reference value, or the value unchanged.
// ``referTo`` is the length for 100%. If ``referTo`` is not a number, it
// just replaces percentages.
func percentage(value pr.Value, referTo float32) float32 {
	if value.IsNone() {
		return None
	} else if value.String == "auto" {
		return Auto
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
func resolveOnePercentage(value pr.Value, propertyName string, referTo float32, mainFlexDirection string) float32 {
	// box.style has computed values
	// value := box.Style[propertyName]

	// box attributes are used values
	percent := percentage(value, referTo)
	// setattr(box, propertyName, percent)
	if (propertyName == "minWidth" || propertyName == "minHeight") && percent == Auto {
		if mainFlexDirection == "" || propertyName != "min"+mainFlexDirection {
			percent = 0
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

func resoudValue(v pr.Value) float32 {
	if v.String == "auto" {
		return Auto
	}
	return v.Value
}

// Set used values as attributes of the box object.
func resolvePercentages(box_ bo.Box, containingBlock bo.Point, mainFlexDirection string) {
	// if isinstance(containingBlock, boxes.Box) {
	//     // cb is short for containing block
	//     cbWidth = containingBlock.width
	//     cbHeight = containingBlock.height
	// }
	cbWidth, cbHeight := containingBlock[0], containingBlock[1]
	maybeHeight := cbWidth
	if _, is := box_.(bo.InstancePageBox); is {
		maybeHeight = cbHeight
	}
	box := box_.Box()
	box.MarginLeft = resolveOnePercentage(box.Style.GetMarginLeft(), "margin_left", cbWidth, "")
	box.MarginRight = resolveOnePercentage(box.Style.GetMarginRight(), "margin_right", cbWidth, "")
	box.MarginTop = resolveOnePercentage(box.Style.GetMarginTop(), "margin_top", maybeHeight, "")
	box.MarginBottom = resolveOnePercentage(box.Style.GetMarginBottom(), "margin_bottom", maybeHeight, "")
	box.PaddingLeft = resolveOnePercentage(box.Style.GetPaddingLeft(), "padding_left", cbWidth, "")
	box.PaddingRight = resolveOnePercentage(box.Style.GetPaddingRight(), "padding_right", cbWidth, "")
	box.PaddingTop = resolveOnePercentage(box.Style.GetPaddingTop(), "padding_top", maybeHeight, "")
	box.PaddingBottom = resolveOnePercentage(box.Style.GetPaddingBottom(), "padding_bottom", maybeHeight, "")
	box.Width = resolveOnePercentage(box.Style.GetWidth(), "width", cbWidth, "")
	box.MinWidth = resolveOnePercentage(box.Style.GetMinWidth(), "min_width", cbWidth, mainFlexDirection)
	box.MaxWidth = resolveOnePercentage(box.Style.GetMaxWidth(), "max_width", cbWidth, mainFlexDirection)

	// XXX later: top, bottom, left && right on positioned elements

	if cbHeight == Auto {
		// Special handling when the height of the containing block
		// depends on its content.
		height := box.Style.GetHeight()
		if height.String == "auto" || height.Unit == pr.Percentage {
			box.Height = Auto
		} else {
			if height.Unit != pr.Px {
				log.Fatalf("expected percentage, got %d", height.Unit)
			}
			box.Height = height.Value
		}
		box.MinHeight = resolveOnePercentage(box.Style.GetMinHeight(), "min_height", 0, mainFlexDirection)
		box.MaxHeight = resolveOnePercentage(box.Style.GetMaxHeight(), "max_height", pr.Inf, mainFlexDirection)
	} else {
		box.Height = resolveOnePercentage(box.Style.GetHeight(), "height", cbHeight, "")
		box.MinHeight = resolveOnePercentage(box.Style.GetMinHeight(), "min_height", cbHeight, mainFlexDirection)
		box.MaxHeight = resolveOnePercentage(box.Style.GetMaxHeight(), "max_height", cbHeight, mainFlexDirection)
	}

	// Used value == computed value
	box.BorderTopWidth = resoudValue(box.Style.GetTop())
	box.BorderRightWidth = resoudValue(box.Style.GetRight())
	box.BorderBottomWidth = resoudValue(box.Style.GetBottom())
	box.BorderLeftWidth = resoudValue(box.Style.GetLeft())

	// Shrink *content* widths and heights according to box-sizing
	// Thanks heavens and the spec: Our validator rejects negative values
	// for padding and border-width
	var horizontalDelta, verticalDelta float32
	switch box.Style.GetBoxSizing() {
	case "border-box":
		horizontalDelta = box.PaddingLeft + box.PaddingRight + box.BorderLeftWidth + box.BorderRightWidth
		verticalDelta = box.PaddingTop + box.PaddingBottom + box.BorderTopWidth + box.BorderBottomWidth
	case "padding-box":
		horizontalDelta = box.PaddingLeft + box.PaddingRight
		verticalDelta = box.PaddingTop + box.PaddingBottom
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
		if box.Width != Auto {
			box.Width = utils.Max(0, box.Width-horizontalDelta)
		}
		box.MaxWidth = utils.Max(0, box.MaxWidth-horizontalDelta)
		if box.MinWidth != Auto {
			box.MinWidth = utils.Max(0, box.MinWidth-horizontalDelta)
		}
	}
	if verticalDelta > 0 {
		if box.Height != Auto {
			box.Height = utils.Max(0, box.Height-verticalDelta)
		}
		box.MaxHeight = utils.Max(0, box.MaxHeight-verticalDelta)
		if box.MinHeight != Auto {
			box.MinHeight = utils.Max(0, box.MinHeight-verticalDelta)
		}
	}
}

func resoudRadius(box *bo.BoxFields, v pr.Point) bo.Point {
	rx := percentage(v[0].ToValue(), box.BorderWidth())
	ry := percentage(v[1].ToValue(), box.BorderHeight())
	return bo.Point{rx, ry}
}

func resolveRadiiPercentages(box *bo.BoxFields) {
	box.BorderTopLeftRadius = resoudRadius(box, box.Style.GetBorderTopLeftRadius())
	box.BorderTopRightRadius = resoudRadius(box, box.Style.GetBorderTopRightRadius())
	box.BorderBottomRightRadius = resoudRadius(box, box.Style.GetBorderBottomRightRadius())
	box.BorderBottomLeftRadius = resoudRadius(box, box.Style.GetBorderBottomLeftRadius())
}
