package validation

import (
	"testing"

	. "github.com/benoitkugler/go-weasyprint/css"
	"github.com/benoitkugler/go-weasyprint/utils"
)

// Test the 4-value properties.
func TestExpandFourSides(t *testing.T) {
	capt := utils.CaptureLogs()
	assertValidDict(t, "margin: inherit", map[string]CssProperty{
		"margin_top":    String(String("inherit")),
		"margin_right":  String(String("inherit")),
		"margin_bottom": String(String("inherit")),
		"margin_left":   String(String("inherit")),
	})
	assertValidDict(t, "margin: 1em", map[string]CssProperty{
		"margin_top":    Dimension{Value: 1, Unit: Em}.ToValue(),
		"margin_right":  Dimension{Value: 1, Unit: Em}.ToValue(),
		"margin_bottom": Dimension{Value: 1, Unit: Em}.ToValue(),
		"margin_left":   Dimension{Value: 1, Unit: Em}.ToValue(),
	})
	assertValidDict(t, "margin: -1em auto 20%", map[string]CssProperty{
		"margin_top":    Dimension{Value: -1, Unit: Em}.ToValue(),
		"margin_right":  SToV("auto"),
		"margin_bottom": Dimension{Value: 20, Unit: Percentage}.ToValue(),
		"margin_left":   SToV("auto"),
	})
	assertValidDict(t, "padding: 1em 0", map[string]CssProperty{
		"padding_top":    Dimension{Value: 1, Unit: Em}.ToValue(),
		"padding_right":  Dimension{Value: 0, Unit: Scalar}.ToValue(),
		"padding_bottom": Dimension{Value: 1, Unit: Em}.ToValue(),
		"padding_left":   Dimension{Value: 0, Unit: Scalar}.ToValue(),
	})
	assertValidDict(t, "padding: 1em 0 2%", map[string]CssProperty{
		"padding_top":    Dimension{Value: 1, Unit: Em}.ToValue(),
		"padding_right":  Dimension{Value: 0, Unit: Scalar}.ToValue(),
		"padding_bottom": Dimension{Value: 2, Unit: Percentage}.ToValue(),
		"padding_left":   Dimension{Value: 0, Unit: Scalar}.ToValue(),
	})
	assertValidDict(t, "padding: 1em 0 2em 5px", map[string]CssProperty{
		"padding_top":    Dimension{Value: 1, Unit: Em}.ToValue(),
		"padding_right":  Dimension{Value: 0, Unit: Scalar}.ToValue(),
		"padding_bottom": Dimension{Value: 2, Unit: Em}.ToValue(),
		"padding_left":   Dimension{Value: 5, Unit: Px}.ToValue(),
	})
	capt.AssertNoLogs(t)

	assertInvalid(t, "padding: 1 2 3 4 5", "Expected 1 to 4 token components got 5")
	assertInvalid(t, "margin: rgb(0, 0, 0)", "invalid")
	assertInvalid(t, "padding: auto", "invalid")
	assertInvalid(t, "padding: -12px", "invalid")
	assertInvalid(t, "border-width: -3em", "invalid")
	assertInvalid(t, "border-width: 12%", "invalid")
}

// Test the ``border`` property.
func TestExpandBorders(t *testing.T) {
	capt := utils.CaptureLogs()
	assertValidDict(t, "border-top: 3px dotted red", map[string]CssProperty{
		"border_top_width": Dimension{Value: 3, Unit: Px}.ToValue(),
		"border_top_style": String("dotted"),
		"border_top_color": NewColor(1, 0, 0, 1), // red
	})
	assertValidDict(t, "border-top: 3px dotted", map[string]CssProperty{
		"border_top_width": Dimension{Value: 3, Unit: Px}.ToValue(),
		"border_top_style": String("dotted"),
	})
	assertValidDict(t, "border-top: 3px red", map[string]CssProperty{
		"border_top_width": Dimension{Value: 3, Unit: Px}.ToValue(),
		"border_top_color": NewColor(1, 0, 0, 1), // red
	})
	assertValidDict(t, "border-top: solid", map[string]CssProperty{
		"border_top_style": String("solid"),
	})
	assertValidDict(t, "border: 6px dashed lime", map[string]CssProperty{
		"border_top_width": Dimension{Value: 6, Unit: Px}.ToValue(),
		"border_top_style": String("dashed"),
		"border_top_color": NewColor(0, 1, 0, 1), // lime

		"border_left_width": Dimension{Value: 6, Unit: Px}.ToValue(),
		"border_left_style": String("dashed"),
		"border_left_color": NewColor(0, 1, 0, 1), // lime

		"border_bottom_width": Dimension{Value: 6, Unit: Px}.ToValue(),
		"border_bottom_style": String("dashed"),
		"border_bottom_color": NewColor(0, 1, 0, 1), // lime

		"border_right_width": Dimension{Value: 6, Unit: Px}.ToValue(),
		"border_right_style": String("dashed"),
		"border_right_color": NewColor(0, 1, 0, 1), // lime
	})
	capt.AssertNoLogs(t)
	assertInvalid(t, "border: 6px dashed left", "invalid")
}

// Test the ``list_style`` property.
func TestExpandList_style(t *testing.T) {
	capt := utils.CaptureLogs()
	assertValidDict(t, "list-style: inherit", map[string]CssProperty{
		"list_style_position": String("inherit"),
		"list_style_image":    String("inherit"),
		"list_style_type":     String("inherit"),
	})
	assertValidDict(t, "list-style: url(../bar/lipsum.png)", map[string]CssProperty{
		"list_style_image": UrlImage("http://weasyprint.org/bar/lipsum.png"),
	})
	assertValidDict(t, "list-style: square", map[string]CssProperty{
		"list_style_type": String("square"),
	})
	assertValidDict(t, "list-style: circle inside", map[string]CssProperty{
		"list_style_position": String("inside"),
		"list_style_type":     String("circle"),
	})
	assertValidDict(t, "list-style: none circle inside", map[string]CssProperty{
		"list_style_position": String("inside"),
		"list_style_image":    NoneImage{},
		"list_style_type":     String("circle"),
	})
	assertValidDict(t, "list-style: none inside none", map[string]CssProperty{
		"list_style_position": String("inside"),
		"list_style_image":    NoneImage{},
		"list_style_type":     String("none"),
	})
	capt.AssertNoLogs(t)
	assertInvalid(t, "list-style: none inside none none", "invalid")
	assertInvalid(t, "list-style: red", "invalid")
	assertInvalid(t, "list-style: circle disc",
		"got multiple type values in a list-style shorthand")
}
