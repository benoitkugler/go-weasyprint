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

// Helper checking the background properties.
func assertBackground(t *testing.T, css string, expected Properties) {
	expanded := expandToDict(t, "background: " + css, "")
	col, in := expected["background_color"]
	if !in {
		col = InitialValues["background_color"]
	}
	if expanded["background_color"] != col {
		t.Fail("expected %v got %v", col, expanded["background_color"])
	}
	delete(expanded, "background_color")
	delete(expected, "background_color")
    nbLayers := len(expanded.GetBackgroundImage())
    for name, value := range expected {
		if expanded[name] != value {
			t.Fatalf("expected %v got %v", value, expanded[name])
		}
	} 
	for name, value := range expanded {
        if isinstance(value, list) {
            // TinyCSS returns lists where StyleDict stores tuples for hashing
            // purpose
            value = tuple(value)
        } assert value == INITIALVALUES[name] * nbLayers
    }
}

// // Test the ``background`` property.
// capt := utils.CaptureLogs()
// func TestExpandBackground(t *testing.T) {
//     assertBackground("red", background_color=(1, 0, 0, 1))
//     assertBackground(
//         "url(lipsum.png)",
//         backgroundImage=[("url", "http://weasyprint.org/foo/lipsum.png")])
//     assertBackground(
//         "no-repeat",
//         backgroundRepeat=[("no-repeat", "no-repeat")])
//     assertBackground("fixed", backgroundAttachment=["fixed"])
//     assertBackground(
//         "repeat no-repeat fixed",
//         backgroundRepeat=[("repeat", "no-repeat")],
//         backgroundAttachment=["fixed"])
//     assertBackground(
//         "top",
//         backgroundPosition=[("left", (50, "%"), "top", (0, "%"))])
//     assertBackground(
//         "top right",
//         backgroundPosition=[("left", (100, "%"), "top", (0, "%"))])
//     assertBackground(
//         "top right 20px",
//         backgroundPosition=[("right", (20, Px), "top", (0, "%"))])
//     assertBackground(
//         "top 1% right 20px",
//         backgroundPosition=[("right", (20, Px), "top", (1, "%"))])
//     assertBackground(
//         "top no-repeat",
//         backgroundRepeat=[("no-repeat", "no-repeat")],
//         backgroundPosition=[("left", (50, "%"), "top", (0, "%"))])
//     assertBackground(
//         "top right no-repeat",
//         backgroundRepeat=[("no-repeat", "no-repeat")],
//         backgroundPosition=[("left", (100, "%"), "top", (0, "%"))])
//     assertBackground(
//         "top right 20px no-repeat",
//         backgroundRepeat=[("no-repeat", "no-repeat")],
//         backgroundPosition=[("right", (20, Px), "top", (0, "%"))])
//     assertBackground(
//         "top 1% right 20px no-repeat",
//         backgroundRepeat=[("no-repeat", "no-repeat")],
//         backgroundPosition=[("right", (20, Px), "top", (1, "%"))])
//     assertBackground(
//         "url(bar) #f00 repeat-y center left fixed",
//         background_color=(1, 0, 0, 1),
//         backgroundImage=[("url", "http://weasyprint.org/foo/bar")],
//         backgroundRepeat=[("no-repeat", "repeat")],
//         backgroundAttachment=["fixed"],
//         backgroundPosition=[("left", (0, "%"), "top", (50, "%"))])
//     assertBackground(
//         "#00f 10% 200px",
//         background_color=(0, 0, 1, 1),
//         backgroundPosition=[("left", (10, "%"), "top", (200, Px))])
//     assertBackground(
//         "right 78px fixed",
//         backgroundAttachment=["fixed"],
//         backgroundPosition=[("left", (100, "%"), "top", (78, Px))])
//     assertBackground(
//         "center / cover red",
//         backgroundSize=["cover"],
//         backgroundPosition=[("left", (50, "%"), "top", (50, "%"))],
//         background_color=(1, 0, 0, 1))
//     assertBackground(
//         "center / auto red",
//         backgroundSize=[("auto", "auto")],
//         backgroundPosition=[("left", (50, "%"), "top", (50, "%"))],
//         background_color=(1, 0, 0, 1))
//     assertBackground(
//         "center / 42px",
//         backgroundSize=[((42, Px), "auto")],
//         backgroundPosition=[("left", (50, "%"), "top", (50, "%"))])
//     assertBackground(
//         "center / 7% 4em",
//         backgroundSize=[((7, "%"), (4, "em"))],
//         backgroundPosition=[("left", (50, "%"), "top", (50, "%"))])
//     assertBackground(
//         "red content-box",
//         background_color=(1, 0, 0, 1),
//         backgroundOrigin=["content-box"],
//         backgroundClip=["content-box"])
//     assertBackground(
//         "red border-box content-box",
//         background_color=(1, 0, 0, 1),
//         backgroundOrigin=["border-box"],
//         backgroundClip=["content-box"])
//     assertBackground(
//         "url(bar) center, no-repeat",
//         background_color=(0, 0, 0, 0),
//         backgroundImage=[("url", "http://weasyprint.org/foo/bar"),
//                           ("none", None)],
//         backgroundPosition=[("left", (50, "%"), "top", (50, "%")),
//                              ("left", (0, "%"), "top", (0, "%"))],
//         backgroundRepeat=[("repeat", "repeat"), ("no-repeat", "no-repeat")])
//     assertInvalid("background: 10px lipsum")
//     assertInvalid("background-position: 10px lipsum")
//     assertInvalid("background: content-box red content-box")
//     assertInvalid("background-image: inexistent-gradient(blue, green)")
//     // Color must be := range the last layer {
//     } assertInvalid("background: red, url(foo)")
// }

// // Test the ``background-position`` property.
// capt := utils.CaptureLogs()
// func TestExpandBackgroundPosition(t *testing.T) {
//     def position(css, *expected) {
//         [(name, [value])] = expandToDict(
//             "background-position:" + css).items()
//         assert name == "backgroundPosition"
//         assert value == expected
//     } for cssX, valX := range [
//         ("left", (0, "%")), ("center", (50, "%")), ("right", (100, "%")),
//         ("4.5%", (4.5, "%")), ("12px", (12, Px))
//     ] {
//         for cssY, valY := range [
//             ("top", (0, "%")), ("center", (50, "%")), ("bottom", (100, "%")),
//             ("7%", (7, "%")), ("1.5px", (1.5, Px))
//         ] {
//             // Two tokens {
//             } position("%s %s" % (cssX, cssY), "left", valX, "top", valY)
//         } // One token {
//         } position(cssX, "left", valX, "top", (50, "%"))
//     } // One token, vertical
//     position("top", "left", (50, "%"), "top", (0, "%"))
//     position("bottom", "left", (50, "%"), "top", (100, "%"))
// }
//     // Three tokens {
//     } position("center top 10%", "left", (50, "%"), "top", (10, "%"))
//     position("top 10% center", "left", (50, "%"), "top", (10, "%"))
//     position("center bottom 10%", "left", (50, "%"), "bottom", (10, "%"))
//     position("bottom 10% center", "left", (50, "%"), "bottom", (10, "%"))

//     position("right top 10%", "right", (0, "%"), "top", (10, "%"))
//     position("top 10% right", "right", (0, "%"), "top", (10, "%"))
//     position("right bottom 10%", "right", (0, "%"), "bottom", (10, "%"))
//     position("bottom 10% right", "right", (0, "%"), "bottom", (10, "%"))

//     position("center left 10%", "left", (10, "%"), "top", (50, "%"))
//     position("left 10% center", "left", (10, "%"), "top", (50, "%"))
//     position("center right 10%", "right", (10, "%"), "top", (50, "%"))
//     position("right 10% center", "right", (10, "%"), "top", (50, "%"))

//     position("bottom left 10%", "left", (10, "%"), "bottom", (0, "%"))
//     position("left 10% bottom", "left", (10, "%"), "bottom", (0, "%"))
//     position("bottom right 10%", "right", (10, "%"), "bottom", (0, "%"))
//     position("right 10% bottom", "right", (10, "%"), "bottom", (0, "%"))

//     // Four tokens {
//     } position("left 10% bottom 3px", "left", (10, "%"), "bottom", (3, Px))
//     position("bottom 3px left 10%", "left", (10, "%"), "bottom", (3, Px))
//     position("right 10% top 3px", "right", (10, "%"), "top", (3, Px))
//     position("top 3px right 10%", "right", (10, "%"), "top", (3, Px))

//     assertInvalid("background-position: left center 3px")
//     assertInvalid("background-position: 3px left")
//     assertInvalid("background-position: bottom 4%")
//     assertInvalid("background-position: bottom top")

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
//     assert expandToDict("line-height: 1.1%") == {"lineHeight": (1.1, "%")}
//     assert expandToDict("line-height: 1em") == {"lineHeight": (1, "em")}
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
//              [blue, lime, red], [(1, "%"), None, (2, "em")])
//     invalid("18deg")
//     gradient("18deg, blue", ("angle", pi / 10))
//     gradient("4rad, blue", ("angle", 4))
//     gradient(".25turn, blue", ("angle", pi / 2))
//     gradient("100grad, blue", ("angle", pi / 2))
//     gradient("12rad, blue 1%, lime,red 2em ", ("angle", 12),
//              [blue, lime, red], [(1, "%"), None, (2, "em")])
//     invalid("10arc-minutes, blue")
//     invalid("10px, blue")
//     invalid("to 90deg, blue")
//     gradient("to top, blue", ("angle", 0))
//     gradient("to right, blue", ("angle", pi / 2))
//     gradient("to bottom, blue", ("angle", pi))
//     gradient("to left, blue", ("angle", pi * 3 / 2))
//     gradient("to right, blue 1%, lime,red 2em ", ("angle", pi / 2),
//              [blue, lime, red], [(1, "%"), None, (2, "em")])
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
//     def gradient(css, shape="ellipse", size=("keyword", "farthest-corner"),
//                  center=("left", (50, "%"), "top", (50, "%")),
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
//              stopPositions=[(1, "%"), None, (2, "em")])
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
//              center=("right", (0, "%"), "top", (10, "%")))
//     gradient("circle at bottom, blue", shape="circle",
//              center=("left", (50, "%"), "top", (100, "%")))
//     gradient("circle at 10px, blue", shape="circle",
//              center=("left", (10, Px), "top", (50, "%")))
//     gradient("closest-side circle at right 5em, blue",
//              shape="circle", size=("keyword", "closest-side"),
//              center=("left", (100, "%"), "top", (5, "em")))
