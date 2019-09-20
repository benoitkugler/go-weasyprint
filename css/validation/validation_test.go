package validation

import (
	"math"
	"reflect"
	"strings"
	"testing"

	. "github.com/benoitkugler/go-weasyprint/css"
	"github.com/benoitkugler/go-weasyprint/css/parser"
	"github.com/benoitkugler/go-weasyprint/utils"
)

// Helper to test shorthand properties expander functions.
func expandToDict(t *testing.T, css string, expectedError string) Properties {
	declarations := parser.ParseDeclarationList2(css, false, false)

	capt := utils.CaptureLogs()
	baseUrl := "http://weasyprint.org/foo/"
	validated := PreprocessDeclarations(baseUrl, declarations)
	logs := capt.Logs()

	if expectedError != "" {
		if len(logs) != 1 || !strings.Contains(logs[0], expectedError) {
			t.Log(validated)
			t.Fatalf("expected error %s got %v (len : %d)", expectedError, logs, len(logs))
		}
	} else {
		capt.AssertNoLogs(t)
	}
	out := Properties{}
	for _, v := range validated {
		if sv, ok := v.Value.(String); !ok || sv != "initial" {
			out[v.Name] = v.Value
		}
	}
	return out
}

// message="invalid"
func assertInvalid(t *testing.T, css, message string) {
	d := expandToDict(t, css, message)
	if len(d) != 0 {
		t.Fatalf("expected no properties, got %v", d)
	}
}

func assertValidDict(t *testing.T, css string, ref Properties) {
	got := expandToDict(t, css, "")
	if !reflect.DeepEqual(ref, got) {
		t.Fatalf("expected %v got %v", ref, got)
	}
}

func TestNotPrint(t *testing.T) {
	capt := utils.CaptureLogs()
	assertInvalid(t, "volume: 42", "the property does not apply for the print media")
	capt.AssertNoLogs(t)
}

func TestFunction(t *testing.T) {
	capt := utils.CaptureLogs()
	assertValidDict(t, "clip: rect(1px, 3em, auto, auto)", Properties{
		"clip": Values{
			Dimension{Value: 1, Unit: Px}.ToValue(),
			Dimension{Value: 3, Unit: Em}.ToValue(),
			SToV("auto"),
			SToV("auto"),
		},
	})
	assertInvalid(t, "clip: square(1px, 3em, auto, auto)", "invalid")
	assertInvalid(t, "clip: rect(1px, 3em, auto auto)", "invalid")
	assertInvalid(t, "clip: rect(1px, 3em, auto)", "invalid")
	assertInvalid(t, "clip: rect(1px, 3em / auto)", "invalid")
	capt.AssertNoLogs(t)
}

func TestCounters(t *testing.T) {
	capt := utils.CaptureLogs()
	assertValidDict(t, "counter-reset: foo bar 2 baz", Properties{
		"counter_reset": IntStrings{{String: "foo", Int: 0}, {String: "bar", Int: 2}, {String: "baz", Int: 0}},
	})
	assertValidDict(t, "counter-increment: foo bar 2 baz", Properties{
		"counter_increment": SIntStrings{Values: []IntString{{String: "foo", Int: 1}, {String: "bar", Int: 2}, {String: "baz", Int: 1}}},
	})
	assertValidDict(t, "counter-reset: foo", Properties{
		"counter_reset": IntStrings{{String: "foo", Int: 0}},
	})
	assertValidDict(t, "counter-reset: FoO", Properties{
		"counter_reset": IntStrings{{String: "FoO", Int: 0}},
	})
	assertValidDict(t, "counter-increment: foo bAr 2 Bar", Properties{
		"counter_increment": SIntStrings{Values: []IntString{{String: "foo", Int: 1}, {String: "bAr", Int: 2}, {String: "Bar", Int: 1}}},
	})
	assertValidDict(t, "counter-reset: none", Properties{
		"counter_reset": IntStrings{},
	})
	capt.AssertNoLogs(t)
	assertInvalid(t, "counter-reset: foo none", "Invalid counter name")
	assertInvalid(t, "counter-reset: foo initial", "Invalid counter name")
	assertInvalid(t, "counter-reset: foo 3px", "invalid")
	assertInvalid(t, "counter-reset: 3", "invalid")
}

func TestSpacing(t *testing.T) {
	capt := utils.CaptureLogs()
	assertValidDict(t, "letter-spacing: normal", Properties{
		"letter_spacing": SToV("normal"),
	})
	assertValidDict(t, "letter-spacing: 3px", Properties{
		"letter_spacing": Dimension{Value: 3, Unit: Px}.ToValue(),
	})
	assertValidDict(t, "word-spacing: normal", Properties{
		"word_spacing": SToV("normal"),
	})
	assertValidDict(t, "word-spacing: 3px", Properties{
		"word_spacing": Dimension{Value: 3, Unit: Px}.ToValue(),
	})
	capt.AssertNoLogs(t)
	assertInvalid(t, "letter_spacing: normal", "did you mean letter-spacing")
	assertInvalid(t, "letter-spacing: 3", "invalid")
	assertInvalid(t, "word-spacing: 3", "invalid")
}

func TestDecoration(t *testing.T) {
	capt := utils.CaptureLogs()
	assertValidDict(t, "text-decoration: none", Properties{
		"text_decoration": NDecorations{None: true},
	})
	assertValidDict(t, "text-decoration: overline", Properties{
		"text_decoration": NDecorations{Decorations: NewSet("overline")},
	})
	// blink is accepted but ignored
	assertValidDict(t, "text-decoration: overline blink line-through", Properties{
		"text_decoration": NDecorations{Decorations: NewSet("line-through", "overline")},
	})
	capt.AssertNoLogs(t)
}

func TestSize(t *testing.T) {
	capt := utils.CaptureLogs()
	assertValidDict(t, "size: 200px", Properties{
		"size": Point{{Value: 200, Unit: Px}, {Value: 200, Unit: Px}},
	})
	assertValidDict(t, "size: 200px 300pt", Properties{
		"size": Point{{Value: 200, Unit: Px}, {Value: 300, Unit: Pt}},
	})
	assertValidDict(t, "size: auto", Properties{
		"size": Point{{Value: 210, Unit: Mm}, {Value: 297, Unit: Mm}},
	})
	assertValidDict(t, "size: portrait", Properties{
		"size": Point{{Value: 210, Unit: Mm}, {Value: 297, Unit: Mm}},
	})
	assertValidDict(t, "size: landscape", Properties{
		"size": Point{{Value: 297, Unit: Mm}, {Value: 210, Unit: Mm}},
	})
	assertValidDict(t, "size: A3 portrait", Properties{
		"size": Point{{Value: 297, Unit: Mm}, {Value: 420, Unit: Mm}},
	})
	assertValidDict(t, "size: A3 landscape", Properties{
		"size": Point{{Value: 420, Unit: Mm}, {Value: 297, Unit: Mm}},
	})
	assertValidDict(t, "size: portrait A3", Properties{
		"size": Point{{Value: 297, Unit: Mm}, {Value: 420, Unit: Mm}},
	})
	assertValidDict(t, "size: landscape A3", Properties{
		"size": Point{{Value: 420, Unit: Mm}, {Value: 297, Unit: Mm}},
	})
	capt.AssertNoLogs(t)
	assertInvalid(t, "size: A3 landscape A3", "invalid")
	assertInvalid(t, "size: A9", "invalid")
	assertInvalid(t, "size: foo", "invalid")
	assertInvalid(t, "size: foo bar", "invalid")
	assertInvalid(t, "size: 20%", "invalid")
}

func TestTransforms(t *testing.T) {
	capt := utils.CaptureLogs()
	assertValidDict(t, "transform: none", Properties{
		"transform": Transforms{},
	})
	assertValidDict(t, "transform: translate(6px) rotate(90deg)", Properties{
		"transform": Transforms{
			{String: "translate", Dimensions: []Dimension{{Value: 6, Unit: Px}, {Value: 0, Unit: Px}}},
			{String: "rotate", Dimensions: []Dimension{FToD(math.Pi / 2)}},
		},
	})
	assertValidDict(t, "transform: translate(-4px, 0)", Properties{
		"transform": Transforms{{String: "translate", Dimensions: []Dimension{{Value: -4, Unit: Px}, {Value: 0, Unit: Scalar}}}},
	})
	assertValidDict(t, "transform: translate(6px, 20%)", Properties{
		"transform": Transforms{{String: "translate", Dimensions: []Dimension{{Value: 6, Unit: Px}, {Value: 20, Unit: Percentage}}}},
	})
	assertValidDict(t, "transform: scale(2)", Properties{
		"transform": Transforms{{String: "scale", Dimensions: []Dimension{FToD(2), FToD(2)}}},
	})
	capt.AssertNoLogs(t)
	assertInvalid(t, "transform: translate(6px 20%)", "invalid") // missing comma
	assertInvalid(t, "transform: lipsumize(6px)", "invalid")
	assertInvalid(t, "transform: foo", "invalid")
	assertInvalid(t, "transform: scale(2) foo", "invalid")
	assertInvalid(t, "transform: 6px", "invalid")
}

type repeatable interface {
	Repeat(int) CssProperty
}

// Helper checking the background properties.
func assertBackground(t *testing.T, css string, expected Properties) {
	expanded := expandToDict(t, "background: "+css, "")
	col, in := expected["background_color"]
	if !in {
		col = InitialValues["background_color"]
	}
	if expanded["background_color"] != col {
		t.Fatalf("expected %v got %v", col, expanded["background_color"])
	}
	delete(expanded, "background_color")
	delete(expected, "background_color")
	nbLayers := len(expanded.GetBackgroundImage())
	for name, value := range expected {
		if !reflect.DeepEqual(expanded[name], value) {
			t.Fatalf("for %s expected %v got %v", name, value, expanded[name])
		}
		delete(expanded, name)
		delete(expected, name)
	}
	for name, value := range expanded {
		initv := InitialValues[name].(repeatable)
		ref := initv.Repeat(nbLayers)
		if !reflect.DeepEqual(value, ref) {
			t.Fatalf("expected %v got %v", ref, value)
		}
	}
}

// Test the ``background`` property.
func TestExpandBackground(t *testing.T) {
	capt := utils.CaptureLogs()
	assertBackground(t, "red", Properties{
		"background_color": NewColor(1, 0, 0, 1),
	})
	assertBackground(t, "url(lipsum.png)", Properties{
		"background_image": Images{UrlImage("http://weasyprint.org/foo/lipsum.png")},
	})
	assertBackground(t, "no-repeat", Properties{
		"background_repeat": Repeats{{"no-repeat", "no-repeat"}},
	})
	assertBackground(t, "fixed", Properties{
		"background_attachment": Strings{"fixed"},
	})
	assertBackground(t, "repeat no-repeat fixed", Properties{
		"background_repeat":     Repeats{{"repeat", "no-repeat"}},
		"background_attachment": Strings{"fixed"},
	})
	assertBackground(t, "top", Properties{
		"background_position": Centers{{OriginX: "left", OriginY: "top", Pos: Point{Dimension{Value: 50, Unit: Percentage}, Dimension{Value: 0, Unit: Percentage}}}},
	})
	assertBackground(t, "top right", Properties{
		"background_position": Centers{{OriginX: "left", OriginY: "top", Pos: Point{Dimension{Value: 100, Unit: Percentage}, Dimension{Value: 0, Unit: Percentage}}}},
	})
	assertBackground(t, "top right 20px", Properties{
		"background_position": Centers{{OriginX: "right", OriginY: "top", Pos: Point{Dimension{Value: 20, Unit: Px}, Dimension{Value: 0, Unit: Percentage}}}},
	})
	assertBackground(t, "top 1% right 20px", Properties{
		"background_position": Centers{{OriginX: "right", OriginY: "top", Pos: Point{Dimension{Value: 20, Unit: Px}, Dimension{Value: 1, Unit: Percentage}}}},
	})
	assertBackground(t, "top no-repeat", Properties{
		"background_repeat":   Repeats{{"no-repeat", "no-repeat"}},
		"background_position": Centers{{OriginX: "left", OriginY: "top", Pos: Point{Dimension{Value: 50, Unit: Percentage}, Dimension{Value: 0, Unit: Percentage}}}},
	})
	assertBackground(t, "top right no-repeat", Properties{
		"background_repeat":   Repeats{{"no-repeat", "no-repeat"}},
		"background_position": Centers{{OriginX: "left", OriginY: "top", Pos: Point{Dimension{Value: 100, Unit: Percentage}, Dimension{Value: 0, Unit: Percentage}}}},
	})
	assertBackground(t, "top right 20px no-repeat", Properties{
		"background_repeat":   Repeats{{"no-repeat", "no-repeat"}},
		"background_position": Centers{{OriginX: "right", OriginY: "top", Pos: Point{Dimension{Value: 20, Unit: Px}, Dimension{Value: 0, Unit: Percentage}}}},
	})
	assertBackground(t, "top 1% right 20px no-repeat", Properties{
		"background_repeat":   Repeats{{"no-repeat", "no-repeat"}},
		"background_position": Centers{{OriginX: "right", OriginY: "top", Pos: Point{Dimension{Value: 20, Unit: Px}, Dimension{Value: 1, Unit: Percentage}}}},
	})
	assertBackground(t, "url(bar) #f00 repeat-y center left fixed", Properties{
		"background_color":      NewColor(1, 0, 0, 1),
		"background_image":      Images{UrlImage("http://weasyprint.org/foo/bar")},
		"background_repeat":     Repeats{{"no-repeat", "repeat"}},
		"background_attachment": Strings{"fixed"},
		"background_position":   Centers{{OriginX: "left", OriginY: "top", Pos: Point{Dimension{Value: 0, Unit: Percentage}, Dimension{Value: 50, Unit: Percentage}}}},
	})
	assertBackground(t, "#00f 10% 200px", Properties{
		"background_color":    NewColor(0, 0, 1, 1),
		"background_position": Centers{{OriginX: "left", OriginY: "top", Pos: Point{Dimension{Value: 10, Unit: Percentage}, Dimension{Value: 200, Unit: Px}}}},
	})
	assertBackground(t, "right 78px fixed", Properties{
		"background_attachment": Strings{"fixed"},
		"background_position":   Centers{{OriginX: "left", OriginY: "top", Pos: Point{Dimension{Value: 100, Unit: Percentage}, Dimension{Value: 78, Unit: Px}}}},
	})
	assertBackground(t, "center / cover red", Properties{
		"background_size":     Sizes{{String: "cover"}},
		"background_position": Centers{{OriginX: "left", OriginY: "top", Pos: Point{Dimension{Value: 50, Unit: Percentage}, Dimension{Value: 50, Unit: Percentage}}}},
		"background_color":    NewColor(1, 0, 0, 1),
	})
	assertBackground(t, "center / auto red", Properties{
		"background_size":     Sizes{{Width: SToV("auto"), Height: SToV("auto")}},
		"background_position": Centers{{OriginX: "left", OriginY: "top", Pos: Point{Dimension{Value: 50, Unit: Percentage}, Dimension{Value: 50, Unit: Percentage}}}},
		"background_color":    NewColor(1, 0, 0, 1),
	})
	assertBackground(t, "center / 42px", Properties{
		"background_size":     Sizes{{Width: Dimension{Value: 42, Unit: Px}.ToValue(), Height: SToV("auto")}},
		"background_position": Centers{{OriginX: "left", OriginY: "top", Pos: Point{Dimension{Value: 50, Unit: Percentage}, Dimension{Value: 50, Unit: Percentage}}}},
	})
	assertBackground(t, "center / 7% 4em", Properties{
		"background_size":     Sizes{{Width: Dimension{Value: 7, Unit: Percentage}.ToValue(), Height: Dimension{Value: 4, Unit: Em}.ToValue()}},
		"background_position": Centers{{OriginX: "left", OriginY: "top", Pos: Point{Dimension{Value: 50, Unit: Percentage}, Dimension{Value: 50, Unit: Percentage}}}},
	})
	assertBackground(t, "red content-box", Properties{
		"background_color":  NewColor(1, 0, 0, 1),
		"background_origin": Strings{"content-box"},
		"background_clip":   Strings{"content-box"},
	})
	assertBackground(t, "red border-box content-box", Properties{
		"background_color":  NewColor(1, 0, 0, 1),
		"background_origin": Strings{"border-box"},
		"background_clip":   Strings{"content-box"},
	})
	assertBackground(t, "url(bar) center, no-repeat", Properties{
		"background_color": NewColor(0, 0, 0, 0),
		"background_image": Images{UrlImage("http://weasyprint.org/foo/bar"), NoneImage{}},
		"background_position": Centers{
			{OriginX: "left", OriginY: "top", Pos: Point{Dimension{Value: 50, Unit: Percentage}, Dimension{Value: 50, Unit: Percentage}}},
			{OriginX: "left", OriginY: "top", Pos: Point{Dimension{Value: 0, Unit: Percentage}, Dimension{Value: 0, Unit: Percentage}}},
		},
		"background_repeat": Repeats{{"repeat", "repeat"}, {"no-repeat", "no-repeat"}},
	})
	capt.AssertNoLogs(t)
	assertInvalid(t, "background: 10px lipsum", "invalid")
	assertInvalid(t, "background-position: 10px lipsum", "invalid")
	assertInvalid(t, "background: content-box red content-box", "invalid")
	assertInvalid(t, "background-image: inexistent-gradient(blue, green)", "invalid")
	// Color must be in the last layer :
	assertInvalid(t, "background: red, url(foo)", "invalid")
}

func checkPosition(t *testing.T, css string, expected Center) {
	l := expandToDict(t, "background-position:" + css, "")
	for name, v := range l {}
	if name != "background_position" {
		t.Fatalf("expected background_position got %s", name)
	}
	if reflect.DeepEqual(v, []Centers{expected}) {
		t.Fatalf("expected %v got %v", expected, v)
	}
}

// Test the ``background-position`` property.
func TestExpandBackgroundPosition(t *testing.T) {
	capt := utils.CaptureLogs()

	css_xs := [5]string{"left", "center","right"	,"4.5%"	,"12px"    }
	val_xs := [5]Dimension{{Value:0, Unit: Percentage},	 {Value:50, Unit: Percentage},	 {Value:100, Unit: Percentage},	{Value:4.5, Unit: Percentage},{Value:12, Unit: Px}}
	css_ys := [5]string{ "top",		"center",		"bottom",		"7%","1.5px"	}
	val_ys := [5]Dimension{ {Value: 0, Unit: Percentage}, {Value: 50, Unit: Percentage}, {Value: 100, Unit: Percentage},{Value: 7, Unit: Percentage}, {Value: 1.5, Unit: Px}	}
	for i, css_x := range css_xs {
		val_x := val_xs[i]
        for j, css_y := range css_ys {
			val_y := val_ys[i]
            // Two tokens:
			checkPosition(t,fmt.Sprintf("%s %s",css_x, css_y), Center{OriginX:"left", OriginY:"top", Pos:Point{val_x, val_y}})
		}
        // One token:
        checkPosition(t,css_x, Center{OriginX:"left", OriginY:"top", Pos:Point{val_x, {Value:50, Unit:Percentage}}})
	}
    // One token, vertical
    checkPosition(t,"top", Center{OriginX:"left", OriginY:"top", Pos:Point{{Value:50, Unit:Percentage},  {Value:0, Unit:Percentage}}})
    checkPosition(t,"bottom", Center{OriginX:"left", OriginY:"top", Pos:Point{{Value:50, Unit:Percentage},  {Value:100, Unit:Percentage}}})

	// Three tokens:
	 checkPosition(t,"center top 10%", Center{OriginX:"left", OriginY:"top", Pos:Point{{Value:50, Unit:Percentage},  {Value:10, Unit:Percentage}}})
    checkPosition(t,"top 10% center", Center{OriginX:"left", OriginY:"top", Pos:Point{{Value:50, Unit:Percentage},  {Value:10, Unit:Percentage}}})
    checkPosition(t,"center bottom 10%", Center{OriginX:"left", OriginY: "bottom", Pos:Point{Value:50, Unit:Percentage}, {Value:10, Unit:Percentage}})
    checkPosition(t,"bottom 10% center", Center{OriginX:"left", OriginY: "bottom", Pos:Point{Value:50, Unit:Percentage}, {Value:10,Unit: Percentage}})

    checkPosition(t,"right top 10%", Center{OriginX:"right",OriginY: "top", Pos:Point{{Value:0, Unit:Percentage},  {Value:10, Unit:Percentage}}})
    checkPosition(t,"top 10% right", Center{OriginX:"right",OriginY: "top", Pos:Point{{Value:0, Unit:Percentage},  {Value:10, Unit:Percentage}}})
    checkPosition(t,"right bottom 10%", "right", OriginY: "bottom", Pos:Point{Value:0, Unit:Percentage}, {Value:10, Unit:Percentage})
    checkPosition(t,"bottom 10% right", "right", OriginY: "bottom", Pos:Point{Value:0, Unit:Percentage}, {Value:10,Unit: Percentage})

    checkPosition(t,"center left 10%", "left", {Value:10, Unit:Percentage}, "top", {Value:50, Unit:Percentage})
    checkPosition(t,"left 10% center", "left", {Value:10, Unit:Percentage}, "top", {Value:50, Unit:Percentage})
    checkPosition(t,"center right 10%", "right", {Value:10, Unit:Percentage}, "top", {Value:50, Unit:Percentage})
    checkPosition(t,"right 10% center", "right", {Value:10, Unit:Percentage}, "top", (50, Percentage})

    checkPosition(t,"bottom left 10%", "left", {Value:10, Unit:Percentage}, "bottom", {Value:0, Unit:Percentage})
    checkPosition(t,"left 10% bottom", "left", {Value:10, Unit:Percentage}, "bottom", {Value:0, Unit:Percentage})
    checkPosition(t,"bottom right 10%", "right", {Value:10, Unit:Percentage}, "bottom", {Value:0, Unit:Percentage})
    checkPosition(t,"right 10% bottom", "right", {Value:10, Unit:Percentage}, "bottom", (0, Percentage})

	// Four tokens :
	checkPosition(t,"left 10% bottom 3px", "left", {Value:10, Unit:Percentage}, "bottom", (3, Px))
    checkPosition(t,"bottom 3px left 10%", "left", {Value:10, Unit:Percentage}, "bottom", (3, Px))
    checkPosition(t,"right 10% top 3px", "right", {Value:10, Unit:Percentage}, "top", (3, Px))
    checkPosition(t,"top 3px right 10%", "right", {Value:10, Unit:Percentage}, "top", (3, Px))

    assertInvalid("background-position: left center 3px")
    assertInvalid("background-position: 3px left")
    assertInvalid("background-position: bottom 4%")
    assertInvalid("background-position: bottom top")
}
// // Test the ``font`` property.
// capt := utils.CaptureLogs()
// func TestFont(t *testing.T) {
//     assert expandToDict("font: 12px My Fancy Font, serif") == {
//         "fontSize": (12, Px),
//         "fontFamily": ("My Fancy Font", "serif"),
//     }
//     assert expandToDict("font: small/1.2 "Some Font", serif") == {
//         "fontSize": "small",
//         "lineHeight": (1.2, None),
//         "fontFamily": ("Some Font", "serif"),
//     }
//     assert expandToDict("font: small-caps italic 700 large serif") == {
//         "fontStyle": "italic",
//         "fontVariantCaps": "small-caps",
//         "fontWeight": 700,
//         "fontSize": "large",
//         "fontFamily": ("serif",),
//     }
//     assert expandToDict(
//         "font: small-caps condensed normal 700 large serif"
//     ) == {
//         // "fontStyle": "normal",  XXX shouldn’t this be here?
//         "fontStretch": "condensed",
//         "fontVariantCaps": "small-caps",
//         "fontWeight": 700,
//         "fontSize": "large",
//         "fontFamily": ("serif",),
//     }
//     assertInvalid("font-family: "My" Font, serif")
//     assertInvalid("font-family: "My" "Font", serif")
//     assertInvalid("font-family: "My", 12pt, serif")
//     assertInvalid("font: menu", "System fonts are not supported")
//     assertInvalid("font: 12deg My Fancy Font, serif")
//     assertInvalid("font: 12px")
//     assertInvalid("font: 12px/foo serif")
//     assertInvalid("font: 12px "Invalid" family")
// }

// // Test the ``font-variant`` property.
// capt := utils.CaptureLogs()
// func TestFontVariant(t *testing.T) {
//     assert expandToDict("font-variant: normal") == {
//         "fontVariantAlternates": "normal",
//         "fontVariantCaps": "normal",
//         "fontVariantEastAsian": "normal",
//         "fontVariantLigatures": "normal",
//         "fontVariantNumeric": "normal",
//         "fontVariantPosition": "normal",
//     }
//     assert expandToDict("font-variant: none") == {
//         "fontVariantAlternates": "normal",
//         "fontVariantCaps": "normal",
//         "fontVariantEastAsian": "normal",
//         "fontVariantLigatures": "none",
//         "fontVariantNumeric": "normal",
//         "fontVariantPosition": "normal",
//     }
//     assert expandToDict("font-variant: historical-forms petite-caps") == {
//         "fontVariantAlternates": "historical-forms",
//         "fontVariantCaps": "petite-caps",
//     }
//     assert expandToDict(
//         "font-variant: lining-nums contextual small-caps common-ligatures"
//     ) == {
//         "fontVariantLigatures": ("contextual", "common-ligatures"),
//         "fontVariantNumeric": ("lining-nums",),
//         "fontVariantCaps": "small-caps",
//     }
//     assert expandToDict("font-variant: jis78 ruby proportional-width") == {
//         "fontVariantEastAsian": ("jis78", "ruby", "proportional-width"),
//     }
//     // CSS2-style font-variant
//     assert expandToDict("font-variant: small-caps") == {
//         "fontVariantCaps": "small-caps",
//     }
//     assertInvalid("font-variant: normal normal")
//     assertInvalid("font-variant: 2")
//     assertInvalid("font-variant: """)
//     assertInvalid("font-variant: extra")
//     assertInvalid("font-variant: jis78 jis04")
//     assertInvalid("font-variant: full-width lining-nums ordinal normal")
//     assertInvalid("font-variant: diagonal-fractions stacked-fractions")
//     assertInvalid(
//         "font-variant: common-ligatures contextual no-common-ligatures")
//     assertInvalid("font-variant: sub super")
//     assertInvalid("font-variant: slashed-zero slashed-zero")
// }

// // Test the ``line-height`` property.
// capt := utils.CaptureLogs()
// func TestLineHeight(t *testing.T) {
//     assert expandToDict("line-height: 1px") == {"lineHeight": (1, Px)}
//     assert expandToDict("line-height: 1.1%") == {"lineHeight": (1.1, Percentage)}
//     assert expandToDict("line-height: 1em") == {"lineHeight": (1, Em)}
//     assert expandToDict("line-height: 1") == {"lineHeight": (1, None)}
//     assert expandToDict("line-height: 1.3") == {"lineHeight": (1.3, None)}
//     assert expandToDict("line-height: -0") == {"lineHeight": (0, None)}
//     assert expandToDict("line-height: 0px") == {"lineHeight": (0, Px)}
//     assertInvalid("line-height: 1deg")
//     assertInvalid("line-height: -1px")
//     assertInvalid("line-height: -1")
//     assertInvalid("line-height: -0.5%")
//     assertInvalid("line-height: 1px 1px")
// }

// // Test the ``string-set`` property.
// capt := utils.CaptureLogs()
// func TestStringSet(t *testing.T) {
//     assert expandToDict("string-set: test content(text)") == {
//         "stringSet": (("test", (("content", "text"),)),)}
//     assert expandToDict("string-set: test content(before)") == {
//         "stringSet": (("test", (("content", "before"),)),)}
//     assert expandToDict("string-set: test "string"") == {
//         "stringSet": (("test", (("STRING", "string"),)),)}
//     assert expandToDict(
//         "string-set: test1 "string", test2 "string"") == {
//             "stringSet": (
//                 ("test1", (("STRING", "string"),)),
//                 ("test2", (("STRING", "string"),)))}
//     assert expandToDict("string-set: test attr(class)") == {
//         "stringSet": (("test", (("attr", "class"),)),)}
//     assert expandToDict("string-set: test counter(count)") == {
//         "stringSet": (("test", (("counter", ("count", "decimal")),)),)}
//     assert expandToDict(
//         "string-set: test counter(count, upper-roman)") == {
//             "stringSet": (
//                 ("test", (("counter", ("count", "upper-roman")),)),)}
//     assert expandToDict("string-set: test counters(count, ".")") == {
//         "stringSet": (("test", (("counters", ("count", ".", "decimal")),)),)}
//     assert expandToDict(
//         "string-set: test counters(count, ".", upper-roman)") == {
//             "stringSet": (
//                 ("test", (("counters", ("count", ".", "upper-roman")),)),)}
//     assert expandToDict(
//         "string-set: test content(text) "string" "
//         "attr(title) attr(title) counter(count)") == {
//             "stringSet": (("test", (
//                 ("content", "text"), ("STRING", "string"),
//                 ("attr", "title"), ("attr", "title"),
//                 ("counter", ("count", "decimal")),)),)}
// }
//     assertInvalid("string-set: test")
//     assertInvalid("string-set: test test1")
//     assertInvalid("string-set: test content(test)")
//     assertInvalid("string-set: test unknown()")
//     assertInvalid("string-set: test attr(id, class)")

// func TestLinearGradient(t *testing.T) {
// capt := utils.CaptureLogs()
// 	red = (1, 0, 0, 1)
//     lime = (0, 1, 0, 1)
//     blue = (0, 0, 1, 1)
//     pi = math.pi
// }
//     def gradient(css, direction, colors=[blue], stopPositions=[None]) {
//         for repeating, prefix := range ((false, ""), (true, "repeating-")) {
//             expanded = expandToDict(
//                 "background-image: %slinear-gradient(%s)" % (prefix, css))
//             [(_, [(type_, image)])] = expanded.items()
//             assert type_ == "linear-gradient"
//             assert isinstance(image, LinearGradient)
//             assert image.repeating == repeating
//             assert almostEqual((image.directionType, image.direction),
//                                 direction)
//             assert almostEqual(image.colors, colors)
//             assert almostEqual(image.stopPositions, stopPositions)
//         }
//     }

//     def invalid(css) {
//         assertInvalid("background-image: linear-gradient(%s)" % css)
//         assertInvalid("background-image: repeating-linear-gradient(%s)" % css)
//     }

//     invalid(" ")
//     invalid("1% blue")
//     invalid("blue 10deg")
//     invalid("blue 4")
//     invalid("soylent-green 4px")
//     invalid("red 4px 2px")
//     gradient("blue", ("angle", pi))
//     gradient("red", ("angle", pi), [red], [None])
//     gradient("blue 1%, lime,red 2em ", ("angle", pi),
//              [blue, lime, red], [(1, Percentage), None, (2, Em)])
//     invalid("18deg")
//     gradient("18deg, blue", ("angle", pi / 10))
//     gradient("4rad, blue", ("angle", 4))
//     gradient(".25turn, blue", ("angle", pi / 2))
//     gradient("100grad, blue", ("angle", pi / 2))
//     gradient("12rad, blue 1%, lime,red 2em ", ("angle", 12),
//              [blue, lime, red], [(1, Percentage), None, (2, Em)])
//     invalid("10arc-minutes, blue")
//     invalid("10px, blue")
//     invalid("to 90deg, blue")
//     gradient("to top, blue", ("angle", 0))
//     gradient("to right, blue", ("angle", pi / 2))
//     gradient("to bottom, blue", ("angle", pi))
//     gradient("to left, blue", ("angle", pi * 3 / 2))
//     gradient("to right, blue 1%, lime,red 2em ", ("angle", pi / 2),
//              [blue, lime, red], [(1, Percentage), None, (2, Em)])
//     invalid("to the top, blue")
//     invalid("to up, blue")
//     invalid("into top, blue")
//     invalid("top, blue")
//     gradient("to top left, blue", ("corner", "topLeft"))
//     gradient("to left top, blue", ("corner", "topLeft"))
//     gradient("to top right, blue", ("corner", "topRight"))
//     gradient("to right top, blue", ("corner", "topRight"))
//     gradient("to bottom left, blue", ("corner", "bottomLeft"))
//     gradient("to left bottom, blue", ("corner", "bottomLeft"))
//     gradient("to bottom right, blue", ("corner", "bottomRight"))
//     gradient("to right bottom, blue", ("corner", "bottomRight"))
//     invalid("to bottom up, blue")
//     invalid("bottom left, blue")

// func TestOverflowWrap(t *testing.T) {
// capt := utils.CaptureLogs()
// 	assert expandToDict("overflow-wrap: normal") == {
//         "overflowWrap": "normal"}
//     assert expandToDict("overflow-wrap: break-word") == {
//         "overflowWrap": "break-word"}
//     assertInvalid("overflow-wrap: none")
//     assertInvalid("overflow-wrap: normal, break-word")
// }

// func TestExpandWordWrap(t *testing.T) {
// capt := utils.CaptureLogs()
// 	assert expandToDict("word-wrap: normal") == {
//         "overflowWrap": "normal"}
//     assert expandToDict("word-wrap: break-word") == {
//         "overflowWrap": "break-word"}
//     assertInvalid("word-wrap: none")
//     assertInvalid("word-wrap: normal, break-word")
// }

// func TestRadialGradient(t *testing.T) {
// capt := utils.CaptureLogs()
// 	red = (1, 0, 0, 1)
//     lime = (0, 1, 0, 1)
//     blue = (0, 0, 1, 1)
// }
//     def gradient(css, shape="ellipse", size:("keyword", "farthest-corner"),
//                  center=("left", (50, Percentage), "top", (50, Percentage)),
//                  colors=[blue], stopPositions=[None]) {
//                  }
//         for repeating, prefix := range ((false, ""), (true, "repeating-")) {
//             expanded = expandToDict(
//                 "background-image: %sradial-gradient(%s)" % (prefix, css))
//             [(_, [(type_, image)])] = expanded.items()
//             assert type_ == "radial-gradient"
//             assert isinstance(image, RadialGradient)
//             assert image.repeating == repeating
//             assert image.shape == shape
//             assert almostEqual((image.sizeType, image.size), size)
//             assert almostEqual(image.center, center)
//             assert almostEqual(image.colors, colors)
//             assert almostEqual(image.stopPositions, stopPositions)
//         }

//     def invalid(css) {
//         assertInvalid("background-image: radial-gradient(%s)" % css)
//         assertInvalid("background-image: repeating-radial-gradient(%s)" % css)
//     }

//     invalid(" ")
//     invalid("1% blue")
//     invalid("blue 10deg")
//     invalid("blue 4")
//     invalid("soylent-green 4px")
//     invalid("red 4px 2px")
//     gradient("blue")
//     gradient("red", colors=[red])
//     gradient("blue 1%, lime,red 2em ", colors=[blue, lime, red],
//              stopPositions=[(1, Percentage), None, (2, Em)])
//     gradient("circle, blue", "circle")
//     gradient("ellipse, blue", "ellipse")
//     invalid("circle")
//     invalid("square, blue")
//     invalid("closest-triangle, blue")
//     invalid("center, blue")
//     gradient("ellipse closest-corner, blue",
//              "ellipse", ("keyword", "closest-corner"))
//     gradient("circle closest-side, blue",
//              "circle", ("keyword", "closest-side"))
//     gradient("farthest-corner circle, blue",
//              "circle", ("keyword", "farthest-corner"))
//     gradient("farthest-side, blue",
//              "ellipse", ("keyword", "farthest-side"))
//     gradient("5ch, blue",
//              "circle", ("explicit", ((5, "ch"), (5, "ch"))))
//     gradient("5ch circle, blue",
//              "circle", ("explicit", ((5, "ch"), (5, "ch"))))
//     gradient("circle 5ch, blue",
//              "circle", ("explicit", ((5, "ch"), (5, "ch"))))
//     invalid("ellipse 5ch")
//     invalid("5ch ellipse")
//     gradient("10px 50px, blue",
//              "ellipse", ("explicit", ((10, Px), (50, Px))))
//     gradient("10px 50px ellipse, blue",
//              "ellipse", ("explicit", ((10, Px), (50, Px))))
//     gradient("ellipse 10px 50px, blue",
//              "ellipse", ("explicit", ((10, Px), (50, Px))))
//     invalid("circle 10px 50px, blue")
//     invalid("10px 50px circle, blue")
//     invalid("10%, blue")
//     invalid("10% circle, blue")
//     invalid("circle 10%, blue")
//     gradient("10px 50px, blue",
//              "ellipse", ("explicit", ((10, Px), (50, Px))))
//     invalid("at appex, blue")
//     gradient("at top 10% right, blue",
//              center=("right", (0, Percentage), "top", (10, Percentage)))
//     gradient("circle at bottom, blue", shape="circle",
//              center=("left", (50, Percentage), "top", (100, Percentage)))
//     gradient("circle at 10px, blue", shape="circle",
//              center=("left", (10, Px), "top", (50, Percentage)))
//     gradient("closest-side circle at right 5em, blue",
//              shape="circle", size:("keyword", "closest-side"),
//              center=("left", (100, Percentage), "top", (5, Em)))
