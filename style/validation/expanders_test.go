package validation

import (
	"testing"

	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/utils"
	"github.com/benoitkugler/go-weasyprint/utils/testutils"
)

// Test the 4-value pr.
func TestExpandFourSides(t *testing.T) {
	capt := testutils.CaptureLogs()
	assertValidDict(t, "margin: inherit", map[string]pr.ValidatedProperty{
		"margin_top":    pr.Inherit.AsCascaded().AsValidated(),
		"margin_right":  pr.Inherit.AsCascaded().AsValidated(),
		"margin_bottom": pr.Inherit.AsCascaded().AsValidated(),
		"margin_left":   pr.Inherit.AsCascaded().AsValidated(),
	})
	assertValidDict(t, "margin: 1em", toValidated(pr.Properties{
		"margin_top":    pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		"margin_right":  pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		"margin_bottom": pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		"margin_left":   pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
	}))
	assertValidDict(t, "margin: -1em auto 20%", toValidated(pr.Properties{
		"margin_top":    pr.Dimension{Value: -1, Unit: pr.Em}.ToValue(),
		"margin_right":  pr.SToV("auto"),
		"margin_bottom": pr.Dimension{Value: 20, Unit: pr.Percentage}.ToValue(),
		"margin_left":   pr.SToV("auto"),
	}))
	assertValidDict(t, "padding: 1em 0", toValidated(pr.Properties{
		"padding_top":    pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		"padding_right":  pr.Dimension{Value: 0, Unit: pr.Scalar}.ToValue(),
		"padding_bottom": pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		"padding_left":   pr.Dimension{Value: 0, Unit: pr.Scalar}.ToValue(),
	}))
	assertValidDict(t, "padding: 1em 0 2%", toValidated(pr.Properties{
		"padding_top":    pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		"padding_right":  pr.Dimension{Value: 0, Unit: pr.Scalar}.ToValue(),
		"padding_bottom": pr.Dimension{Value: 2, Unit: pr.Percentage}.ToValue(),
		"padding_left":   pr.Dimension{Value: 0, Unit: pr.Scalar}.ToValue(),
	}))
	assertValidDict(t, "padding: 1em 0 2em 5px", toValidated(pr.Properties{
		"padding_top":    pr.Dimension{Value: 1, Unit: pr.Em}.ToValue(),
		"padding_right":  pr.Dimension{Value: 0, Unit: pr.Scalar}.ToValue(),
		"padding_bottom": pr.Dimension{Value: 2, Unit: pr.Em}.ToValue(),
		"padding_left":   pr.Dimension{Value: 5, Unit: pr.Px}.ToValue(),
	}))
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
	capt := testutils.CaptureLogs()
	assertValidDict(t, "border-top: 3px dotted red", toValidated(pr.Properties{
		"border_top_width": pr.Dimension{Value: 3, Unit: pr.Px}.ToValue(),
		"border_top_style": pr.String("dotted"),
		"border_top_color": pr.NewColor(1, 0, 0, 1), // red
	}))
	assertValidDict(t, "border-top: 3px dotted", toValidated(pr.Properties{
		"border_top_width": pr.Dimension{Value: 3, Unit: pr.Px}.ToValue(),
		"border_top_style": pr.String("dotted"),
	}))
	assertValidDict(t, "border-top: 3px red", toValidated(pr.Properties{
		"border_top_width": pr.Dimension{Value: 3, Unit: pr.Px}.ToValue(),
		"border_top_color": pr.NewColor(1, 0, 0, 1), // red
	}))
	assertValidDict(t, "border-top: solid", toValidated(pr.Properties{
		"border_top_style": pr.String("solid"),
	}))
	assertValidDict(t, "border: 6px dashed lime", toValidated(pr.Properties{
		"border_top_width": pr.Dimension{Value: 6, Unit: pr.Px}.ToValue(),
		"border_top_style": pr.String("dashed"),
		"border_top_color": pr.NewColor(0, 1, 0, 1), // lime

		"border_left_width": pr.Dimension{Value: 6, Unit: pr.Px}.ToValue(),
		"border_left_style": pr.String("dashed"),
		"border_left_color": pr.NewColor(0, 1, 0, 1), // lime

		"border_bottom_width": pr.Dimension{Value: 6, Unit: pr.Px}.ToValue(),
		"border_bottom_style": pr.String("dashed"),
		"border_bottom_color": pr.NewColor(0, 1, 0, 1), // lime

		"border_right_width": pr.Dimension{Value: 6, Unit: pr.Px}.ToValue(),
		"border_right_style": pr.String("dashed"),
		"border_right_color": pr.NewColor(0, 1, 0, 1), // lime
	}))
	capt.AssertNoLogs(t)
	assertInvalid(t, "border: 6px dashed left", "invalid")
}

// Test the ``list_style`` property.
func TestExpandList_style(t *testing.T) {
	capt := testutils.CaptureLogs()
	assertValidDict(t, "list-style: inherit", map[string]pr.ValidatedProperty{
		"list_style_position": pr.Inherit.AsCascaded().AsValidated(),
		"list_style_image":    pr.Inherit.AsCascaded().AsValidated(),
		"list_style_type":     pr.Inherit.AsCascaded().AsValidated(),
	})
	assertValidDict(t, "list-style: url(../bar/lipsum.png)", toValidated(pr.Properties{
		"list_style_image": pr.UrlImage("http://weasyprint.org/bar/lipsum.png"),
	}))
	assertValidDict(t, "list-style: square", toValidated(pr.Properties{
		"list_style_type": pr.CounterStyleID{Name: "square"},
	}))
	assertValidDict(t, "list-style: circle inside", toValidated(pr.Properties{
		"list_style_position": pr.String("inside"),
		"list_style_type":     pr.CounterStyleID{Name: "circle"},
	}))
	assertValidDict(t, "list-style: none circle inside", toValidated(pr.Properties{
		"list_style_position": pr.String("inside"),
		"list_style_image":    pr.NoneImage{},
		"list_style_type":     pr.CounterStyleID{Name: "circle"},
	}))
	assertValidDict(t, "list-style: none inside none", toValidated(pr.Properties{
		"list_style_position": pr.String("inside"),
		"list_style_image":    pr.NoneImage{},
		"list_style_type":     pr.CounterStyleID{Name: "none"},
	}))
	capt.AssertNoLogs(t)
	assertInvalid(t, "list-style: none inside none none", "invalid")
	assertInvalid(t, "list-style: 1px", "invalid")
	assertInvalid(t, "list-style: circle disc",
		"got multiple type values in a list-style shorthand")
}

// Test the ``font`` property.
func TestFont(t *testing.T) {
	capt := testutils.CaptureLogs()
	assertValidDict(t, "font: 12px My Fancy Font, serif", toValidated(pr.Properties{
		"font_size":   pr.Dimension{Value: 12, Unit: pr.Px}.ToValue(),
		"font_family": pr.Strings{"My Fancy Font", "serif"},
	}))
	assertValidDict(t, `font: small/1.2 "Some Font", serif`, toValidated(pr.Properties{
		"font_size":   pr.SToV("small"),
		"line_height": pr.Dimension{Value: 1.2, Unit: pr.Scalar}.ToValue(),
		"font_family": pr.Strings{"Some Font", "serif"},
	}))
	assertValidDict(t, "font: small-caps italic 700 large serif", toValidated(pr.Properties{
		"font_style":        pr.String("italic"),
		"font_variant_caps": pr.String("small-caps"),
		"font_weight":       pr.IntString{Int: 700},
		"font_size":         pr.SToV("large"),
		"font_family":       pr.Strings{"serif"},
	}))
	assertValidDict(t, "font: small-caps condensed normal 700 large serif", toValidated(pr.Properties{
		// "font_style": String("normal"),  XXX shouldn’t this be here?
		"font_stretch":      pr.String("condensed"),
		"font_variant_caps": pr.String("small-caps"),
		"font_weight":       pr.IntString{Int: 700},
		"font_size":         pr.SToV("large"),
		"font_family":       pr.Strings{"serif"},
	}))
	capt.AssertNoLogs(t)
	assertInvalid(t, `font-family: "My" Font, serif`, "invalid")
	assertInvalid(t, `font-family: "My" "Font", serif`, "invalid")
	assertInvalid(t, `font-family: "My", 12pt, serif`, "invalid")
	assertInvalid(t, `font: menu`, "System fonts are not supported")
	assertInvalid(t, `font: 12deg My Fancy Font, serif`, "invalid")
	assertInvalid(t, `font: 12px`, "invalid")
	assertInvalid(t, `font: 12px/foo serif`, "invalid")
	assertInvalid(t, `font: 12px "Invalid" family`, "invalid")
	assertInvalid(t, "font: normal normal normal normal normal large serif", "invalid")
	assertInvalid(t, "font: normal small-caps italic 700 condensed large serif", "invalid")
	assertInvalid(t, "font: small-caps italic 700 normal condensed large serif", "invalid")
	assertInvalid(t, "font: small-caps italic 700 condensed normal large serif", "invalid")
}

// Test the ``font-variant`` property.
func TestFontVariant(t *testing.T) {
	capt := testutils.CaptureLogs()
	assertValidDict(t, "font-variant: normal", toValidated(pr.Properties{
		"font_variant_alternates": pr.String("normal"),
		"font_variant_caps":       pr.String("normal"),
		"font_variant_east_asian": pr.SStrings{String: "normal"},
		"font_variant_ligatures":  pr.SStrings{String: "normal"},
		"font_variant_numeric":    pr.SStrings{String: "normal"},
		"font_variant_position":   pr.String("normal"),
	}))
	assertValidDict(t, "font-variant: none", toValidated(pr.Properties{
		"font_variant_alternates": pr.String("normal"),
		"font_variant_caps":       pr.String("normal"),
		"font_variant_east_asian": pr.SStrings{String: "normal"},
		"font_variant_ligatures":  pr.SStrings{String: "none"},
		"font_variant_numeric":    pr.SStrings{String: "normal"},
		"font_variant_position":   pr.String("normal"),
	}))
	assertValidDict(t, "font-variant: historical-forms petite-caps", toValidated(pr.Properties{
		"font_variant_alternates": pr.String("historical-forms"),
		"font_variant_caps":       pr.String("petite-caps"),
	}))
	assertValidDict(t, "font-variant: lining-nums contextual small-caps common-ligatures", toValidated(pr.Properties{
		"font_variant_ligatures": pr.SStrings{Strings: []string{"contextual", "common-ligatures"}},
		"font_variant_numeric":   pr.SStrings{Strings: []string{"lining-nums"}},
		"font_variant_caps":      pr.String("small-caps"),
	}))
	assertValidDict(t, "font-variant: jis78 ruby proportional-width", toValidated(pr.Properties{
		"font_variant_east_asian": pr.SStrings{Strings: []string{"jis78", "ruby", "proportional-width"}},
	}))
	// CSS2-style font-variant
	assertValidDict(t, "font-variant: small-caps", toValidated(pr.Properties{
		"font_variant_caps": pr.String("small-caps"),
	}))
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
	capt := testutils.CaptureLogs()
	assertValidDict(t, "word-wrap: normal", toValidated(pr.Properties{
		"overflow_wrap": pr.String("normal"),
	}))
	assertValidDict(t, "word-wrap: break-word", toValidated(pr.Properties{
		"overflow_wrap": pr.String("break-word"),
	}))
	capt.AssertNoLogs(t)
	assertInvalid(t, "word-wrap: none", "invalid")
	assertInvalid(t, "word-wrap: normal, break-word", "invalid")
}

func fillTextDecoration(prop pr.Properties) map[string]pr.ValidatedProperty {
	base := map[string]pr.ValidatedProperty{
		"text_decoration_line":  pr.AsCascaded(pr.NDecorations{None: true}).AsValidated(),
		"text_decoration_color": pr.AsCascaded(pr.CurrentColor).AsValidated(),
		"text_decoration_style": pr.AsCascaded(pr.String("solid")).AsValidated(),
	}
	for k, v := range prop {
		base[k] = pr.AsCascaded(v).AsValidated()
	}
	return base
}

func TestExpandTextDecoration(t *testing.T) {
	capt := testutils.CaptureLogs()

	assertValidDict(t, "text-decoration: none", fillTextDecoration(pr.Properties{
		"text_decoration_line": pr.NDecorations{None: true},
	}))
	assertValidDict(t, "text-decoration: overline", fillTextDecoration(pr.Properties{
		"text_decoration_line": pr.NDecorations{Decorations: utils.NewSet("overline")},
	}))
	assertValidDict(t, "text-decoration: overline blink line-through", fillTextDecoration(pr.Properties{
		"text_decoration_line": pr.NDecorations{Decorations: utils.NewSet("blink", "line-through", "overline")},
	}))
	assertValidDict(t, "text-decoration: red", fillTextDecoration(pr.Properties{
		"text_decoration_color": pr.NewColor(1, 0, 0, 1),
	}))
	capt.AssertNoLogs(t)
}

func TestExpandFlex(t *testing.T) {
	capt := testutils.CaptureLogs()

	assertValidDict(t, "flex: auto", toValidated(pr.Properties{
		"flex_grow":   pr.Float(1),
		"flex_shrink": pr.Float(1),
		"flex_basis":  pr.SToV("auto"),
	}))
	assertValidDict(t, "flex: none", toValidated(pr.Properties{
		"flex_grow":   pr.Float(0),
		"flex_shrink": pr.Float(0),
		"flex_basis":  pr.SToV("auto"),
	}))
	assertValidDict(t, "flex: 10", toValidated(pr.Properties{
		"flex_grow":   pr.Float(10),
		"flex_shrink": pr.Float(1),
		"flex_basis":  pr.ZeroPixels.ToValue(),
	}))
	assertValidDict(t, "flex: 2 2", toValidated(pr.Properties{
		"flex_grow":   pr.Float(2),
		"flex_shrink": pr.Float(2),
		"flex_basis":  pr.ZeroPixels.ToValue(),
	}))
	assertValidDict(t, "flex: 2 2 1px", toValidated(pr.Properties{
		"flex_grow":   pr.Float(2),
		"flex_shrink": pr.Float(2),
		"flex_basis":  pr.Dimension{Value: 1, Unit: pr.Px}.ToValue(),
	}))
	assertValidDict(t, "flex: 2 2 auto", toValidated(pr.Properties{
		"flex_grow":   pr.Float(2),
		"flex_shrink": pr.Float(2),
		"flex_basis":  pr.SToV("auto"),
	}))
	assertValidDict(t, "flex: 2 auto", toValidated(pr.Properties{
		"flex_grow":   pr.Float(2),
		"flex_shrink": pr.Float(1),
		"flex_basis":  pr.SToV("auto"),
	}))

	capt.AssertNoLogs(t)
}

func TestLineClamp(t *testing.T) {
	capt := testutils.CaptureLogs()

	assertValidDict(t, "line-clamp: none", toValidated(pr.Properties{
		"max_lines":      pr.IntString{String: "none"},
		"continue":       pr.String("auto"),
		"block_ellipsis": pr.NamedString{String: "none"},
	}))
	assertValidDict(t, "line-clamp: 2", toValidated(pr.Properties{
		"max_lines":      pr.IntString{Int: 2},
		"continue":       pr.String("discard"),
		"block_ellipsis": pr.NamedString{String: "auto"},
	}))
	assertValidDict(t, `line-clamp: 3 "…"`, toValidated(pr.Properties{
		"max_lines":      pr.IntString{Int: 3},
		"continue":       pr.String("discard"),
		"block_ellipsis": pr.NamedString{Name: "string", String: "…"},
	}))

	capt.AssertNoLogs(t)

	assertInvalid(t, `line-clamp: none none none`, "invalid")
	assertInvalid(t, `line-clamp: 1px`, "invalid")
	assertInvalid(t, `line-clamp: 0 "…"`, "invalid")
	assertInvalid(t, `line-clamp: 1px 2px`, "invalid")
}
