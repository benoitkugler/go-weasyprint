package validation

import (
	"reflect"
	"strings"
	"testing"

	. "github.com/benoitkugler/go-weasyprint/css"
	"github.com/benoitkugler/go-weasyprint/css/parser"
	"github.com/benoitkugler/go-weasyprint/utils"
)

// Helper to test shorthand properties expander functions.
func expandToDict(t *testing.T, css string, expectedError string) map[string]CssProperty {
	declarations := parser.ParseDeclarationList2(css, false, false)

	capt := utils.CaptureLogs()
	baseUrl := "http://weasyprint.org/foo/"
	validated := PreprocessDeclarations(baseUrl, declarations)
	logs := capt.Logs()

	if expectedError != "" {
		if len(logs) != 1 || !strings.Contains(logs[0], expectedError) {
			t.Fatalf("expected %s got %v (len : %d)", expectedError, logs, len(logs))
		}
	} else {
		capt.AssertNoLogs(t)
	}
	out := map[string]CssProperty{}
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

func TestNotPrint(t *testing.T) {
	capt := utils.CaptureLogs()
	assertInvalid(t, "volume: 42", "the property does not apply for the print media")
	capt.AssertNoLogs(t)
}

func TestFunction(t *testing.T) {
	capt := utils.CaptureLogs()
	d := expandToDict(t, "clip: rect(1px, 3em, auto, auto)", "")
	ref := map[string]CssProperty{
		"clip": Values{
			Dimension{Value: 1, Unit: Px}.ToValue(),
			Dimension{Value: 3, Unit: Em}.ToValue(),
			SToV("auto"),
			SToV("auto"),
		},
	}
	if !reflect.DeepEqual(ref, d) {
		t.Fatalf("expected %v got %v", ref, d)
	}
	assertInvalid(t, "clip: square(1px, 3em, auto, auto)", "invalid")
	assertInvalid(t, "clip: rect(1px, 3em, auto auto)", "invalid")
	assertInvalid(t, "clip: rect(1px, 3em, auto)", "invalid")
	assertInvalid(t, "clip: rect(1px, 3em / auto)", "invalid")
	capt.AssertNoLogs(t)
}

// func TestCounters(t *testing.T) {
// capt := utils.CaptureLogs()
// 	assert expandToDict("counter-reset: foo bar 2 baz") == {
//         "counterReset": (("foo", 0), ("bar", 2), ("baz", 0))}
//     assert expandToDict("counter-increment: foo bar 2 baz") == {
//         "counterIncrement": (("foo", 1), ("bar", 2), ("baz", 1))}
//     assert expandToDict("counter-reset: foo") == {
//         "counterReset": (("foo", 0),)}
//     assert expandToDict("counter-reset: FoO") == {
//         "counterReset": (("FoO", 0),)}
//     assert expandToDict("counter-increment: foo bAr 2 Bar") == {
//         "counterIncrement": (("foo", 1), ("bAr", 2), ("Bar", 1))}
//     assert expandToDict("counter-reset: none") == {
//         "counterReset": ()}
//     assert expandToDict(
//         "counter-reset: foo none", "Invalid counter name") == {}
//     assert expandToDict(
//         "counter-reset: foo initial", "Invalid counter name") == {}
//     assertInvalid("counter-reset: foo 3px")
//     assertInvalid("counter-reset: 3")
// }

// func TestSpacing(t *testing.T) {
// capt := utils.CaptureLogs()
// 	assert expandToDict("letter-spacing: normal") == {
//         "letterSpacing": "normal"}
//     assert expandToDict("letter-spacing: 3px") == {
//         "letterSpacing": (3, "px")}
//     assertInvalid("letter-spacing: 3")
//     assert expandToDict(
//         "letterSpacing: normal", "did you mean letter-spacing") == {}
// }
//     assert expandToDict("word-spacing: normal") == {
//         "wordSpacing": "normal"}
//     assert expandToDict("word-spacing: 3px") == {
//         "wordSpacing": (3, "px")}
//     assertInvalid("word-spacing: 3")

// func TestDecoration(t *testing.T) {
// capt := utils.CaptureLogs()
// 	assert expandToDict("text-decoration: none") == {
//         "textDecoration": "none"}
//     assert expandToDict("text-decoration: overline") == {
//         "textDecoration": frozenset(["overline"])}
//     // blink is accepted but ignored
//     assert expandToDict("text-decoration: overline blink line-through") == {
//         "textDecoration": frozenset(["line-through", "overline"])}
// }

// func TestSize(t *testing.T) {
// capt := utils.CaptureLogs()
// 	assert expandToDict("size: 200px") == {
//         "size": ((200, "px"), (200, "px"))}
//     assert expandToDict("size: 200px 300pt") == {
//         "size": ((200, "px"), (300, "pt"))}
//     assert expandToDict("size: auto") == {
//         "size": ((210, "mm"), (297, "mm"))}
//     assert expandToDict("size: portrait") == {
//         "size": ((210, "mm"), (297, "mm"))}
//     assert expandToDict("size: landscape") == {
//         "size": ((297, "mm"), (210, "mm"))}
//     assert expandToDict("size: A3 portrait") == {
//         "size": ((297, "mm"), (420, "mm"))}
//     assert expandToDict("size: A3 landscape") == {
//         "size": ((420, "mm"), (297, "mm"))}
//     assert expandToDict("size: portrait A3") == {
//         "size": ((297, "mm"), (420, "mm"))}
//     assert expandToDict("size: landscape A3") == {
//         "size": ((420, "mm"), (297, "mm"))}
//     assertInvalid("size: A3 landscape A3")
//     assertInvalid("size: A9")
//     assertInvalid("size: foo")
//     assertInvalid("size: foo bar")
//     assertInvalid("size: 20%")
// }

// func TestTransforms(t *testing.T) {
// capt := utils.CaptureLogs()
// 	assert expandToDict("transform: none") == {
//         "transform": ()}
//     assert expandToDict(
//         "transform: translate(6px) rotate(90deg)"
//     ) == {"transform": (("translate", ((6, "px"), (0, "px"))),
//                         ("rotate", math.pi / 2))}
//     assert expandToDict(
//         "transform: translate(-4px, 0)"
//     ) == {"transform": (("translate", ((-4, "px"), (0, None))),)}
//     assert expandToDict(
//         "transform: translate(6px, 20%)"
//     ) == {"transform": (("translate", ((6, "px"), (20, "%"))),)}
//     assert expandToDict(
//         "transform: scale(2)"
//     ) == {"transform": (("scale", (2, 2)),)}
//     assertInvalid("transform: translate(6px 20%)")  // missing comma
//     assertInvalid("transform: lipsumize(6px)")
//     assertInvalid("transform: foo")
//     assertInvalid("transform: scale(2) foo")
//     assertInvalid("transform: 6px")
// }

// // Test the 4-value properties.
// capt := utils.CaptureLogs()
// func TestExpandFourSides(t *testing.T) {
//     assert expandToDict("margin: inherit") == {
//         "marginTop": "inherit",
//         "marginRight": "inherit",
//         "marginBottom": "inherit",
//         "marginLeft": "inherit",
//     }
//     assert expandToDict("margin: 1em") == {
//         "marginTop": (1, "em"),
//         "marginRight": (1, "em"),
//         "marginBottom": (1, "em"),
//         "marginLeft": (1, "em"),
//     }
//     assert expandToDict("margin: -1em auto 20%") == {
//         "marginTop": (-1, "em"),
//         "marginRight": "auto",
//         "marginBottom": (20, "%"),
//         "marginLeft": "auto",
//     }
//     assert expandToDict("padding: 1em 0") == {
//         "paddingTop": (1, "em"),
//         "paddingRight": (0, None),
//         "paddingBottom": (1, "em"),
//         "paddingLeft": (0, None),
//     }
//     assert expandToDict("padding: 1em 0 2%") == {
//         "paddingTop": (1, "em"),
//         "paddingRight": (0, None),
//         "paddingBottom": (2, "%"),
//         "paddingLeft": (0, None),
//     }
//     assert expandToDict("padding: 1em 0 2em 5px") == {
//         "paddingTop": (1, "em"),
//         "paddingRight": (0, None),
//         "paddingBottom": (2, "em"),
//         "paddingLeft": (5, "px"),
//     }
//     assert expandToDict(
//         "padding: 1 2 3 4 5",
//         "Expected 1 to 4 token components got 5") == {}
//     assertInvalid("margin: rgb(0, 0, 0)")
//     assertInvalid("padding: auto")
//     assertInvalid("padding: -12px")
//     assertInvalid("border-width: -3em")
//     assertInvalid("border-width: 12%")
// }

// // Test the ``border`` property.
// capt := utils.CaptureLogs()
// func TestExpandBorders(t *testing.T) {
//     assert expandToDict("border-top: 3px dotted red") == {
//         "borderTopWidth": (3, "px"),
//         "borderTopStyle": "dotted",
//         "borderTopColor": (1, 0, 0, 1),  // red
//     }
//     assert expandToDict("border-top: 3px dotted") == {
//         "borderTopWidth": (3, "px"),
//         "borderTopStyle": "dotted",
//     }
//     assert expandToDict("border-top: 3px red") == {
//         "borderTopWidth": (3, "px"),
//         "borderTopColor": (1, 0, 0, 1),  // red
//     }
//     assert expandToDict("border-top: solid") == {
//         "borderTopStyle": "solid",
//     }
//     assert expandToDict("border: 6px dashed lime") == {
//         "borderTopWidth": (6, "px"),
//         "borderTopStyle": "dashed",
//         "borderTopColor": (0, 1, 0, 1),  // lime
// }
//         "borderLeftWidth": (6, "px"),
//         "borderLeftStyle": "dashed",
//         "borderLeftColor": (0, 1, 0, 1),  // lime

//         "borderBottomWidth": (6, "px"),
//         "borderBottomStyle": "dashed",
//         "borderBottomColor": (0, 1, 0, 1),  // lime

//         "borderRightWidth": (6, "px"),
//         "borderRightStyle": "dashed",
//         "borderRightColor": (0, 1, 0, 1),  // lime
//     }
//     assertInvalid("border: 6px dashed left")

// // Test the ``listStyle`` property.
// capt := utils.CaptureLogs()
// func TestExpandListStyle(t *testing.T) {
//     assert expandToDict("list-style: inherit") == {
//         "listStylePosition": "inherit",
//         "listStyleImage": "inherit",
//         "listStyleType": "inherit",
//     }
//     assert expandToDict("list-style: url(../bar/lipsum.png)") == {
//         "listStyleImage": ("url", "http://weasyprint.org/bar/lipsum.png"),
//     }
//     assert expandToDict("list-style: square") == {
//         "listStyleType": "square",
//     }
//     assert expandToDict("list-style: circle inside") == {
//         "listStylePosition": "inside",
//         "listStyleType": "circle",
//     }
//     assert expandToDict("list-style: none circle inside") == {
//         "listStylePosition": "inside",
//         "listStyleImage": ("none", None),
//         "listStyleType": "circle",
//     }
//     assert expandToDict("list-style: none inside none") == {
//         "listStylePosition": "inside",
//         "listStyleImage": ("none", None),
//         "listStyleType": "none",
//     }
//     assertInvalid("list-style: none inside none none")
//     assertInvalid("list-style: red")
//     assertInvalid("list-style: circle disc",
//                    "got multiple type values := range a list-style shorthand")
// }

// // Helper checking the background properties.
// func assertBackground(css, **expected) {
//     expanded = expandToDict("background: " + css)
//     assert expanded.pop("backgroundColor") == expected.pop(
//         "backgroundColor", INITIALVALUES["backgroundColor"])
//     nbLayers = len(expanded["backgroundImage"])
//     for name, value := range expected.items() {
//         assert expanded.pop(name) == value
//     } for name, value := range expanded.items() {
//         if isinstance(value, list) {
//             // TinyCSS returns lists where StyleDict stores tuples for hashing
//             // purpose
//             value = tuple(value)
//         } assert value == INITIALVALUES[name] * nbLayers
//     }
// }

// // Test the ``background`` property.
// capt := utils.CaptureLogs()
// func TestExpandBackground(t *testing.T) {
//     assertBackground("red", backgroundColor=(1, 0, 0, 1))
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
//         backgroundPosition=[("right", (20, "px"), "top", (0, "%"))])
//     assertBackground(
//         "top 1% right 20px",
//         backgroundPosition=[("right", (20, "px"), "top", (1, "%"))])
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
//         backgroundPosition=[("right", (20, "px"), "top", (0, "%"))])
//     assertBackground(
//         "top 1% right 20px no-repeat",
//         backgroundRepeat=[("no-repeat", "no-repeat")],
//         backgroundPosition=[("right", (20, "px"), "top", (1, "%"))])
//     assertBackground(
//         "url(bar) #f00 repeat-y center left fixed",
//         backgroundColor=(1, 0, 0, 1),
//         backgroundImage=[("url", "http://weasyprint.org/foo/bar")],
//         backgroundRepeat=[("no-repeat", "repeat")],
//         backgroundAttachment=["fixed"],
//         backgroundPosition=[("left", (0, "%"), "top", (50, "%"))])
//     assertBackground(
//         "#00f 10% 200px",
//         backgroundColor=(0, 0, 1, 1),
//         backgroundPosition=[("left", (10, "%"), "top", (200, "px"))])
//     assertBackground(
//         "right 78px fixed",
//         backgroundAttachment=["fixed"],
//         backgroundPosition=[("left", (100, "%"), "top", (78, "px"))])
//     assertBackground(
//         "center / cover red",
//         backgroundSize=["cover"],
//         backgroundPosition=[("left", (50, "%"), "top", (50, "%"))],
//         backgroundColor=(1, 0, 0, 1))
//     assertBackground(
//         "center / auto red",
//         backgroundSize=[("auto", "auto")],
//         backgroundPosition=[("left", (50, "%"), "top", (50, "%"))],
//         backgroundColor=(1, 0, 0, 1))
//     assertBackground(
//         "center / 42px",
//         backgroundSize=[((42, "px"), "auto")],
//         backgroundPosition=[("left", (50, "%"), "top", (50, "%"))])
//     assertBackground(
//         "center / 7% 4em",
//         backgroundSize=[((7, "%"), (4, "em"))],
//         backgroundPosition=[("left", (50, "%"), "top", (50, "%"))])
//     assertBackground(
//         "red content-box",
//         backgroundColor=(1, 0, 0, 1),
//         backgroundOrigin=["content-box"],
//         backgroundClip=["content-box"])
//     assertBackground(
//         "red border-box content-box",
//         backgroundColor=(1, 0, 0, 1),
//         backgroundOrigin=["border-box"],
//         backgroundClip=["content-box"])
//     assertBackground(
//         "url(bar) center, no-repeat",
//         backgroundColor=(0, 0, 0, 0),
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
//         ("4.5%", (4.5, "%")), ("12px", (12, "px"))
//     ] {
//         for cssY, valY := range [
//             ("top", (0, "%")), ("center", (50, "%")), ("bottom", (100, "%")),
//             ("7%", (7, "%")), ("1.5px", (1.5, "px"))
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
//     } position("left 10% bottom 3px", "left", (10, "%"), "bottom", (3, "px"))
//     position("bottom 3px left 10%", "left", (10, "%"), "bottom", (3, "px"))
//     position("right 10% top 3px", "right", (10, "%"), "top", (3, "px"))
//     position("top 3px right 10%", "right", (10, "%"), "top", (3, "px"))

//     assertInvalid("background-position: left center 3px")
//     assertInvalid("background-position: 3px left")
//     assertInvalid("background-position: bottom 4%")
//     assertInvalid("background-position: bottom top")

// // Test the ``font`` property.
// capt := utils.CaptureLogs()
// func TestFont(t *testing.T) {
//     assert expandToDict("font: 12px My Fancy Font, serif") == {
//         "fontSize": (12, "px"),
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
//     assert expandToDict("line-height: 1px") == {"lineHeight": (1, "px")}
//     assert expandToDict("line-height: 1.1%") == {"lineHeight": (1.1, "%")}
//     assert expandToDict("line-height: 1em") == {"lineHeight": (1, "em")}
//     assert expandToDict("line-height: 1") == {"lineHeight": (1, None)}
//     assert expandToDict("line-height: 1.3") == {"lineHeight": (1.3, None)}
//     assert expandToDict("line-height: -0") == {"lineHeight": (0, None)}
//     assert expandToDict("line-height: 0px") == {"lineHeight": (0, "px")}
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
//              "ellipse", ("explicit", ((10, "px"), (50, "px"))))
//     gradient("10px 50px ellipse, blue",
//              "ellipse", ("explicit", ((10, "px"), (50, "px"))))
//     gradient("ellipse 10px 50px, blue",
//              "ellipse", ("explicit", ((10, "px"), (50, "px"))))
//     invalid("circle 10px 50px, blue")
//     invalid("10px 50px circle, blue")
//     invalid("10%, blue")
//     invalid("10% circle, blue")
//     invalid("circle 10%, blue")
//     gradient("10px 50px, blue",
//              "ellipse", ("explicit", ((10, "px"), (50, "px"))))
//     invalid("at appex, blue")
//     gradient("at top 10% right, blue",
//              center=("right", (0, "%"), "top", (10, "%")))
//     gradient("circle at bottom, blue", shape="circle",
//              center=("left", (50, "%"), "top", (100, "%")))
//     gradient("circle at 10px, blue", shape="circle",
//              center=("left", (10, "px"), "top", (50, "%")))
//     gradient("closest-side circle at right 5em, blue",
//              shape="circle", size=("keyword", "closest-side"),
//              center=("left", (100, "%"), "top", (5, "em")))
