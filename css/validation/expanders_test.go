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
		"margin_top":    String("inherit"),
		"margin_right":  String("inherit"),
		"margin_bottom": String("inherit"),
		"margin_left":   String("inherit"),
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

// Test the ``font`` property.
func TestFont(t *testing.T) {
	capt := utils.CaptureLogs()
	assertValidDict(t, "font: 12px My Fancy Font, serif", map[string]CssProperty{
		"font_size":   Dimension{Value: 12, Unit: Px}.ToValue(),
		"font_family": Strings{"My Fancy Font", "serif"},
	})
	assertValidDict(t, `font: small/1.2 "Some Font", serif`, map[string]CssProperty{
		"font_size":   SToV("small"),
		"line_height": Dimension{Value: 1.2, Unit: Scalar}.ToValue(),
		"font_family": Strings{"Some Font", "serif"},
	})
	assertValidDict(t, "font: small-caps italic 700 large serif", map[string]CssProperty{
		"font_style":        String("italic"),
		"font_variant_caps": String("small-caps"),
		"font_weight":       IntString{Int: 700},
		"font_size":         SToV("large"),
		"font_family":       Strings{"serif"},
	})
	assertValidDict(t, "font: small-caps condensed normal 700 large serif", map[string]CssProperty{
		// "font_style": String("normal"),  XXX shouldn’t this be here?
		"font_stretch":      String("condensed"),
		"font_variant_caps": String("small-caps"),
		"font_weight":       IntString{Int: 700},
		"font_size":         SToV("large"),
		"font_family":       Strings{"serif"},
	})
	capt.AssertNoLogs(t)
	assertInvalid(t, `font-family: "My" Font, serif`, "invalid")
	assertInvalid(t, `font-family: "My" "Font", serif`, "invalid")
	assertInvalid(t, `font-family: "My", 12pt, serif`, "invalid")
	assertInvalid(t, `font: menu`, "System fonts are not supported")
	assertInvalid(t, `font: 12deg My Fancy Font, serif`, "invalid")
	assertInvalid(t, `font: 12px`, "invalid")
	assertInvalid(t, `font: 12px/foo serif`, "invalid")
	assertInvalid(t, `font: 12px "Invalid" family`, "invalid")
}

// Test the ``font-variant`` property.
func TestFontVariant(t *testing.T) {
	capt := utils.CaptureLogs()
	assertValidDict(t, "font-variant: normal", Properties{
		"font_variant_alternates": String("normal"),
		"font_variant_caps":       String("normal"),
		"font_variant_east_asian": SStrings{String: "normal"},
		"font_variant_ligatures":  SStrings{String: "normal"},
		"font_variant_numeric":    SStrings{String: "normal"},
		"font_variant_position":   String("normal"),
	})
	assertValidDict(t, "font-variant: none", Properties{
		"font_variant_alternates": String("normal"),
		"font_variant_caps":       String("normal"),
		"font_variant_east_asian": SStrings{String: "normal"},
		"font_variant_ligatures":  SStrings{String: "none"},
		"font_variant_numeric":    SStrings{String: "normal"},
		"font_variant_position":   String("normal"),
	})
	assertValidDict(t, "font-variant: historical-forms petite-caps", Properties{
		"font_variant_alternates": String("historical-forms"),
		"font_variant_caps":       String("petite-caps"),
	})
	assertValidDict(t, "font-variant: lining-nums contextual small-caps common-ligatures", Properties{
		"font_variant_ligatures": SStrings{Strings: []string{"contextual", "common-ligatures"}},
		"font_variant_numeric":   SStrings{Strings: []string{"lining-nums"}},
		"font_variant_caps":      String("small-caps"),
	})
	assertValidDict(t, "font-variant: jis78 ruby proportional-width", Properties{
		"font_variant_east_asian": SStrings{Strings: []string{"jis78", "ruby", "proportional-width"}},
	})
	// CSS2-style font-variant
	assertValidDict(t, "font-variant: small-caps", Properties{
		"font_variant_caps": String("small-caps"),
	})
	capt.AssertNoLogs(t)
	assertInvalid(t, "font-variant: normal normal", "invalid")
	assertInvalid(t, "font-variant: 2", "invalid")
	assertInvalid(t, `font-variant: ""`, "invalid")
	assertInvalid(t, "font-variant: extra", "invalid")
	assertInvalid(t, "font-variant: jis78 jis04", "invalid")
	assertInvalid(t, "font-variant: full-width lining-nums ordinal normal", "invalid")
	assertInvalid(t, "font-variant: diagonal-fractions stacked-fractions", "invalid")
	assertInvalid(t, "font-variant: common-ligatures contextual no-common-ligatures", "invalid")
	assertInvalid(t, "font-variant: sub super", "invalid")
	assertInvalid(t, "font-variant: slashed-zero slashed-zero", "invalid")
}

func TestExpandWordWrap(t *testing.T) {
	capt := utils.CaptureLogs()
	assertValidDict(t, "word-wrap: normal", Properties{
		"overflow_wrap": String("normal"),
	})
	assertValidDict(t, "word-wrap: break-word", Properties{
		"overflow_wrap": String("break-word"),
	})
	capt.AssertNoLogs(t)
	assertInvalid(t, "word-wrap: none", "invalid")
	assertInvalid(t, "word-wrap: normal, break-word", "invalid")
}

func TestExpandFlex(t *testing.T) {
	capt := utils.CaptureLogs()

	assertValidDict(t, "flex: auto", Properties{
		"flex_grow":   Float(1),
		"flex_shrink": Float(1),
		"flex_basis":  SToV("auto"),
	})
	assertValidDict(t, "flex: none", Properties{
		"flex_grow":   Float(0),
		"flex_shrink": Float(0),
		"flex_basis":  SToV("auto"),
	})
	assertValidDict(t, "flex: 10", Properties{
		"flex_grow":   Float(10),
		"flex_shrink": Float(1),
		"flex_basis":  ZeroPixels.ToValue(),
	})
	assertValidDict(t, "flex: 2 2", Properties{
		"flex_grow":   Float(2),
		"flex_shrink": Float(2),
		"flex_basis":  ZeroPixels.ToValue(),
	})
	assertValidDict(t, "flex: 2 2 1px", Properties{
		"flex_grow":   Float(2),
		"flex_shrink": Float(2),
		"flex_basis":  Dimension{Value: 1, Unit: Px}.ToValue(),
	})
	assertValidDict(t, "flex: 2 2 auto", Properties{
		"flex_grow":   Float(2),
		"flex_shrink": Float(2),
		"flex_basis":  SToV("auto"),
	})
	assertValidDict(t, "flex: 2 auto", Properties{
		"flex_grow":   Float(2),
		"flex_shrink": Float(1),
		"flex_basis":  SToV("auto"),
	})

	capt.AssertNoLogs(t)
}
