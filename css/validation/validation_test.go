package validation

import (
	"fmt"
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
			t.Fatalf("for %s expected error \n%s\n got\n%v (len : %d)", css, expectedError, logs, len(logs))
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
	assertInvalid(t, "counter-reset: foo none", "Invalid counter name: initial.")
	assertInvalid(t, "counter-reset: foo initial", "Invalid counter name: none.")
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
	l := expandToDict(t, "background-position:"+css, "")
	var (
		name string
		v    CssProperty
	)
	for name_, v_ := range l {
		name = name_
		v = v_
	}
	if name != "background_position" {
		t.Fatalf("expected background_position got %s", name)
	}
	exp := Centers{expected}
	if !reflect.DeepEqual(v, exp) {
		t.Fatalf("expected %v got %v", exp, v)
	}
}

// Test the ``background-position`` property.
func TestExpandBackgroundPosition(t *testing.T) {
	capt := utils.CaptureLogs()

	css_xs := [5]string{"left", "center", "right", "4.5%", "12px"}
	val_xs := [5]Dimension{{Value: 0, Unit: Percentage}, {Value: 50, Unit: Percentage}, {Value: 100, Unit: Percentage}, {Value: 4.5, Unit: Percentage}, {Value: 12, Unit: Px}}
	css_ys := [5]string{"top", "center", "bottom", "7%", "1.5px"}
	val_ys := [5]Dimension{{Value: 0, Unit: Percentage}, {Value: 50, Unit: Percentage}, {Value: 100, Unit: Percentage}, {Value: 7, Unit: Percentage}, {Value: 1.5, Unit: Px}}
	for i, css_x := range css_xs {
		val_x := val_xs[i]
		for j, css_y := range css_ys {
			val_y := val_ys[j]
			// Two tokens:
			checkPosition(t, fmt.Sprintf("%s %s", css_x, css_y), Center{OriginX: "left", OriginY: "top", Pos: Point{val_x, val_y}})
		}
		// One token:
		checkPosition(t, css_x, Center{OriginX: "left", OriginY: "top", Pos: Point{val_x, {Value: 50, Unit: Percentage}}})
	}
	// One token, vertical
	checkPosition(t, "top", Center{OriginX: "left", OriginY: "top", Pos: Point{{Value: 50, Unit: Percentage}, {Value: 0, Unit: Percentage}}})
	checkPosition(t, "bottom", Center{OriginX: "left", OriginY: "top", Pos: Point{{Value: 50, Unit: Percentage}, {Value: 100, Unit: Percentage}}})

	// Three tokens:
	checkPosition(t, "center top 10%", Center{OriginX: "left", OriginY: "top", Pos: Point{{Value: 50, Unit: Percentage}, {Value: 10, Unit: Percentage}}})
	checkPosition(t, "top 10% center", Center{OriginX: "left", OriginY: "top", Pos: Point{{Value: 50, Unit: Percentage}, {Value: 10, Unit: Percentage}}})
	checkPosition(t, "center bottom 10%", Center{OriginX: "left", OriginY: "bottom", Pos: Point{{Value: 50, Unit: Percentage}, {Value: 10, Unit: Percentage}}})
	checkPosition(t, "bottom 10% center", Center{OriginX: "left", OriginY: "bottom", Pos: Point{{Value: 50, Unit: Percentage}, {Value: 10, Unit: Percentage}}})

	checkPosition(t, "right top 10%", Center{OriginX: "right", OriginY: "top", Pos: Point{{Value: 0, Unit: Percentage}, {Value: 10, Unit: Percentage}}})
	checkPosition(t, "top 10% right", Center{OriginX: "right", OriginY: "top", Pos: Point{{Value: 0, Unit: Percentage}, {Value: 10, Unit: Percentage}}})
	checkPosition(t, "right bottom 10%", Center{OriginX: "right", OriginY: "bottom", Pos: Point{{Value: 0, Unit: Percentage}, {Value: 10, Unit: Percentage}}})
	checkPosition(t, "bottom 10% right", Center{OriginX: "right", OriginY: "bottom", Pos: Point{{Value: 0, Unit: Percentage}, {Value: 10, Unit: Percentage}}})

	checkPosition(t, "center left 10%", Center{OriginX: "left", OriginY: "top", Pos: Point{{Value: 10, Unit: Percentage}, {Value: 50, Unit: Percentage}}})
	checkPosition(t, "left 10% center", Center{OriginX: "left", OriginY: "top", Pos: Point{{Value: 10, Unit: Percentage}, {Value: 50, Unit: Percentage}}})
	checkPosition(t, "center right 10%", Center{OriginX: "right", OriginY: "top", Pos: Point{{Value: 10, Unit: Percentage}, {Value: 50, Unit: Percentage}}})
	checkPosition(t, "right 10% center", Center{OriginX: "right", OriginY: "top", Pos: Point{{Value: 10, Unit: Percentage}, {Value: 50, Unit: Percentage}}})

	checkPosition(t, "bottom left 10%", Center{OriginX: "left", OriginY: "bottom", Pos: Point{{Value: 10, Unit: Percentage}, {Value: 0, Unit: Percentage}}})
	checkPosition(t, "left 10% bottom", Center{OriginX: "left", OriginY: "bottom", Pos: Point{{Value: 10, Unit: Percentage}, {Value: 0, Unit: Percentage}}})
	checkPosition(t, "bottom right 10%", Center{OriginX: "right", OriginY: "bottom", Pos: Point{{Value: 10, Unit: Percentage}, {Value: 0, Unit: Percentage}}})
	checkPosition(t, "right 10% bottom", Center{OriginX: "right", OriginY: "bottom", Pos: Point{{Value: 10, Unit: Percentage}, {Value: 0, Unit: Percentage}}})

	// Four tokens :
	checkPosition(t, "left 10% bottom 3px", Center{OriginX: "left", OriginY: "bottom", Pos: Point{{Value: 10, Unit: Percentage}, {Value: 3, Unit: Px}}})
	checkPosition(t, "bottom 3px left 10%", Center{OriginX: "left", OriginY: "bottom", Pos: Point{{Value: 10, Unit: Percentage}, {Value: 3, Unit: Px}}})
	checkPosition(t, "right 10% top 3px", Center{OriginX: "right", OriginY: "top", Pos: Point{{Value: 10, Unit: Percentage}, {Value: 3, Unit: Px}}})
	checkPosition(t, "top 3px right 10%", Center{OriginX: "right", OriginY: "top", Pos: Point{{Value: 10, Unit: Percentage}, {Value: 3, Unit: Px}}})

	capt.AssertNoLogs(t)

	assertInvalid(t, "background-position: left center 3px", "invalid")
	assertInvalid(t, "background-position: 3px left", "invalid")
	assertInvalid(t, "background-position: bottom 4%", "invalid")
	assertInvalid(t, "background-position: bottom top", "invalid")
}

// Test the ``line-height`` property.
func TestLineHeight(t *testing.T) {
	capt := utils.CaptureLogs()
	assertValidDict(t, "line-height: 1px", Properties{
		"line_height": Dimension{Value: 1, Unit: Px}.ToValue(),
	})
	assertValidDict(t, "line-height: 1.1%", Properties{
		"line_height": Dimension{Value: 1.1, Unit: Percentage}.ToValue(),
	})
	assertValidDict(t, "line-height: 1em", Properties{
		"line_height": Dimension{Value: 1, Unit: Em}.ToValue(),
	})
	assertValidDict(t, "line-height: 1", Properties{
		"line_height": Dimension{Value: 1, Unit: Scalar}.ToValue(),
	})
	assertValidDict(t, "line-height: 1.3", Properties{
		"line_height": Dimension{Value: 1.3, Unit: Scalar}.ToValue(),
	})
	assertValidDict(t, "line-height: -0", Properties{
		"line_height": Dimension{Value: 0, Unit: Scalar}.ToValue(),
	})
	assertValidDict(t, "line-height: 0px", Properties{
		"line_height": Dimension{Value: 0, Unit: Px}.ToValue(),
	})
	capt.AssertNoLogs(t)
	assertInvalid(t, "line-height: 1deg", "invalid")
	assertInvalid(t, "line-height: -1px", "invalid")
	assertInvalid(t, "line-height: -1", "invalid")
	assertInvalid(t, "line-height: -0.5%", "invalid")
	assertInvalid(t, "line-height: 1px 1px", "invalid")
}

// Test the ``string-set`` property.
func TestStringSet(t *testing.T) {
	capt := utils.CaptureLogs()
	assertValidDict(t, "string-set: test content(text)", Properties{
		"string_set": StringSet{Contents: []SContent{
			{String: "test", Contents: []ContentProperty{{Type: ContentContent, SStrings: SToSS("text")}}},
		}},
	})
	assertValidDict(t, "string-set: test content(before)", Properties{
		"string_set": StringSet{Contents: []SContent{
			{String: "test", Contents: []ContentProperty{{Type: ContentContent, SStrings: SToSS("before")}}},
		}},
	})
	assertValidDict(t, `string-set: test "string"`, Properties{
		"string_set": StringSet{Contents: []SContent{
			{String: "test", Contents: []ContentProperty{{Type: ContentSTRING, SStrings: SToSS("string")}}},
		}},
	})
	assertValidDict(t, `string-set: test1 "string", test2 "string"`, Properties{
		"string_set": StringSet{Contents: []SContent{
			{String: "test1", Contents: []ContentProperty{{Type: ContentSTRING, SStrings: SToSS("string")}}},
			{String: "test2", Contents: []ContentProperty{{Type: ContentSTRING, SStrings: SToSS("string")}}},
		}},
	})
	assertValidDict(t, "string-set: test attr(class)", Properties{
		"string_set": StringSet{Contents: []SContent{
			{String: "test", Contents: []ContentProperty{{Type: ContentAttr, SStrings: SToSS("class")}}},
		}},
	})
	assertValidDict(t, "string-set: test counter(count)", Properties{
		"string_set": StringSet{Contents: []SContent{
			{String: "test", Contents: []ContentProperty{{Type: ContentCounter, SStrings: SStrings{Strings: []string{"count", "decimal"}}}}},
		}},
	})
	assertValidDict(t, "string-set: test counter(count, upper-roman)", Properties{
		"string_set": StringSet{Contents: []SContent{
			{String: "test", Contents: []ContentProperty{{Type: ContentCounter, SStrings: SStrings{Strings: []string{"count", "upper-roman"}}}}},
		}},
	})
	assertValidDict(t, `string-set: test counters(count, ".")`, Properties{
		"string_set": StringSet{Contents: []SContent{
			{String: "test", Contents: []ContentProperty{{Type: ContentCounters, SStrings: SStrings{Strings: []string{"count", ".", "decimal"}}}}},
		}},
	})
	assertValidDict(t, `string-set: test counters(count, ".", upper-roman)`, Properties{
		"string_set": StringSet{Contents: []SContent{
			{String: "test", Contents: []ContentProperty{{Type: ContentCounters, SStrings: SStrings{Strings: []string{"count", ".", "upper-roman"}}}}},
		}},
	})
	assertValidDict(t, `string-set: test content(text) "string" attr(title) attr(title) counter(count)`, Properties{
		"string_set": StringSet{Contents: []SContent{
			{String: "test", Contents: []ContentProperty{
				{Type: ContentContent, SStrings: SToSS("text")},
				{Type: ContentSTRING, SStrings: SToSS("string")},
				{Type: ContentAttr, SStrings: SToSS("title")},
				{Type: ContentAttr, SStrings: SToSS("title")},
				{Type: ContentCounter, SStrings: SStrings{Strings: []string{"count", "decimal"}}},
			}},
		}},
	})

	capt.AssertNoLogs(t)
	assertInvalid(t, "string-set: test", "invalid")
	assertInvalid(t, "string-set: test test1", "invalid")
	assertInvalid(t, "string-set: test content(test)", "invalid")
	assertInvalid(t, "string-set: test unknown()", "invalid")
	assertInvalid(t, "string-set: test attr(id, class)", "invalid")
}

var (
	red          = NewColor(1, 0, 0, 1)
	lime         = NewColor(0, 1, 0, 1)
	blue         = NewColor(0, 0, 1, 1)
	pi   float32 = math.Pi
)

func checkGradientGeneric(t *testing.T, css string, expected Image) {
	repeatings := [2]bool{false, true}
	prefixs := [2]string{"", "repeating-"}
	for i, repeating := range repeatings {
		prefix := prefixs[i]
		var mode string
		switch typed := expected.(type) {
		case LinearGradient:
			typed.Repeating = repeating
			expected = typed
			mode = "linear"
		case RadialGradient:
			typed.Repeating = repeating
			expected = typed
			mode = "radial"
		default:
			t.Fatalf("bad expected gradient !")
		}

		expanded := expandToDict(t, fmt.Sprintf("background-image: %s%s-gradient(%s)", prefix, mode, css), "")
		var image Image
		for _, v := range expanded {
			image = v.(Images)[0]
		}
		if !reflect.DeepEqual(image, expected) {
			t.Fatalf("expected %v got %v", expected, image)
		}
	}
}

func invalidGeneric(mode string, t *testing.T, css string) {
	assertInvalid(t, fmt.Sprintf("background-image: %s-gradient(%s)", mode, css), "invalid")
	assertInvalid(t, fmt.Sprintf("background-image: repeating-%s-gradient(%s)", mode, css), "invalid")
}

func TestLinearGradient(t *testing.T) {
	invalid := func(t *testing.T, css string) {
		invalidGeneric("linear", t, css)
	}

	gradient := func(t *testing.T, css string, direction DirectionType, colors []Color, stopPositions []Dimension) {
		if colors == nil {
			colors = []Color{blue}
		}
		if stopPositions == nil {
			stopPositions = []Dimension{Dimension{}}
		}
		colorStops := make([]ColorStop, len(colors))
		for i, s := range stopPositions {
			colorStops[i] = ColorStop{Color: colors[i], Position: s}
		}
		checkGradientGeneric(t, css, LinearGradient{ColorStops: colorStops, Direction: direction})
	}
	invalid(t, " ")
	invalid(t, "1% blue")
	invalid(t, "blue 10deg")
	invalid(t, "blue 4")
	invalid(t, "soylent-green 4px")
	invalid(t, "red 4px 2px")

	invalid(t, "18deg")

	invalid(t, "10arc-minutes, blue")
	invalid(t, "10px, blue")
	invalid(t, "to 90deg, blue")

	invalid(t, "to the top, blue")
	invalid(t, "to up, blue")
	invalid(t, "into top, blue")
	invalid(t, "top, blue")

	invalid(t, "to bottom up, blue")
	invalid(t, "bottom left, blue")

	capt := utils.CaptureLogs()
	gradient(t, "blue", DirectionType{Angle: pi}, nil, nil)
	gradient(t, "red", DirectionType{Angle: pi}, []Color{red}, []Dimension{Dimension{}})
	gradient(t, "blue 1%, lime,red 2em ", DirectionType{Angle: pi},
		[]Color{blue, lime, red}, []Dimension{{Value: 1, Unit: Percentage}, Dimension{}, {Value: 2, Unit: Em}})

	gradient(t, "18deg, blue", DirectionType{Angle: pi / 10}, nil, nil)
	gradient(t, "4rad, blue", DirectionType{Angle: 4}, nil, nil)
	gradient(t, ".25turn, blue", DirectionType{Angle: pi / 2}, nil, nil)
	gradient(t, "100grad, blue", DirectionType{Angle: pi / 2}, nil, nil)
	gradient(t, "12rad, blue 1%, lime,red 2em ", DirectionType{Angle: 12},
		[]Color{blue, lime, red}, []Dimension{{Value: 1, Unit: Percentage}, Dimension{}, {Value: 2, Unit: Em}})

	gradient(t, "to top, blue", DirectionType{Angle: 0}, nil, nil)
	gradient(t, "to right, blue", DirectionType{Angle: pi / 2}, nil, nil)
	gradient(t, "to bottom, blue", DirectionType{Angle: pi}, nil, nil)
	gradient(t, "to left, blue", DirectionType{Angle: pi * 3 / 2}, nil, nil)
	gradient(t, "to right, blue 1%, lime,red 2em ", DirectionType{Angle: pi / 2},
		[]Color{blue, lime, red}, []Dimension{{Value: 1, Unit: Percentage}, Dimension{}, {Value: 2, Unit: Em}})

	gradient(t, "to top left, blue", DirectionType{Corner: "top_left"}, nil, nil)
	gradient(t, "to left top, blue", DirectionType{Corner: "top_left"}, nil, nil)
	gradient(t, "to top right, blue", DirectionType{Corner: "top_right"}, nil, nil)
	gradient(t, "to right top, blue", DirectionType{Corner: "top_right"}, nil, nil)
	gradient(t, "to bottom left, blue", DirectionType{Corner: "bottom_left"}, nil, nil)
	gradient(t, "to left bottom, blue", DirectionType{Corner: "bottom_left"}, nil, nil)
	gradient(t, "to bottom right, blue", DirectionType{Corner: "bottom_right"}, nil, nil)
	gradient(t, "to right bottom, blue", DirectionType{Corner: "bottom_right"}, nil, nil)
	capt.AssertNoLogs(t)
}

func TestOverflowWrap(t *testing.T) {
	capt := utils.CaptureLogs()
	assertValidDict(t, "overflow-wrap: normal", Properties{
		"overflow_wrap": String("normal"),
	})
	assertValidDict(t, "overflow-wrap: break-word", Properties{
		"overflow_wrap": String("break-word"),
	})
	capt.AssertNoLogs(t)
	assertInvalid(t, "overflow-wrap: none", "invalid")
	assertInvalid(t, "overflow-wrap: normal, break-word", "invalid")
}

func TestRadialGradient(t *testing.T) {
	capt := utils.CaptureLogs()

	gradient := func(t *testing.T, css string, shape string, size GradientSize, center Center, colors []Color, stopPositions []Dimension) {
		if colors == nil {
			colors = []Color{blue}
		}
		if stopPositions == nil {
			stopPositions = []Dimension{Dimension{}}
		}
		colorStops := make([]ColorStop, len(colors))
		for i, s := range stopPositions {
			colorStops[i] = ColorStop{Color: colors[i], Position: s}
		}
		if shape == "" {
			shape = "ellipse"
		}
		if size.IsNone() {
			size = GradientSize{Keyword: "farthest-corner"}
		}
		if center.IsNone() {
			center = Center{OriginX: "left", OriginY: "top", Pos: Point{{Value: 50, Unit: Percentage}, {Value: 50, Unit: Percentage}}}
		}
		checkGradientGeneric(t, css, RadialGradient{ColorStops: colorStops, Shape: shape, Size: size, Center: center})
	}

	invalid := func(t *testing.T, css string) {
		invalidGeneric("radial", t, css)
	}

	invalid(t, " ")
	invalid(t, "1% blue")
	invalid(t, "blue 10deg")
	invalid(t, "blue 4")
	invalid(t, "soylent-green 4px")
	invalid(t, "red 4px 2px")

	invalid(t, "circle")
	invalid(t, "square, blue")
	invalid(t, "closest-triangle, blue")
	invalid(t, "center, blue")

	invalid(t, "ellipse 5ch")
	invalid(t, "5ch ellipse")

	invalid(t, "circle 10px 50px, blue")
	invalid(t, "10px 50px circle, blue")
	invalid(t, "10%, blue")
	invalid(t, "10% circle, blue")
	invalid(t, "circle 10%, blue")

	invalid(t, "at appex, blue")
	capt.AssertNoLogs(t)

	gradient(t, "blue", "", GradientSize{}, Center{}, nil, nil)
	gradient(t, "red", "", GradientSize{}, Center{}, []Color{red}, nil)
	gradient(t, "blue 1%, lime,red 2em ", "", GradientSize{}, Center{},
		[]Color{blue, lime, red},
		[]Dimension{{Value: 1, Unit: Percentage}, Dimension{}, {Value: 2, Unit: Em}})
	gradient(t, "circle, blue", "circle", GradientSize{}, Center{}, nil, nil)
	gradient(t, "ellipse, blue", "ellipse", GradientSize{}, Center{}, nil, nil)

	gradient(t, "ellipse closest-corner, blue",
		"ellipse", GradientSize{Keyword: "closest-corner"}, Center{}, nil, nil)
	gradient(t, "circle closest-side, blue",
		"circle", GradientSize{Keyword: "closest-side"}, Center{}, nil, nil)
	gradient(t, "farthest-corner circle, blue",
		"circle", GradientSize{Keyword: "farthest-corner"}, Center{}, nil, nil)
	gradient(t, "farthest-side, blue",
		"ellipse", GradientSize{Keyword: "farthest-side"}, Center{}, nil, nil)
	gradient(t, "5ch, blue",
		"circle", GradientSize{Explicit: Point{{Value: 5, Unit: Ch}, {Value: 5, Unit: Ch}}}, Center{}, nil, nil)
	gradient(t, "5ch circle, blue",
		"circle", GradientSize{Explicit: Point{{Value: 5, Unit: Ch}, {Value: 5, Unit: Ch}}}, Center{}, nil, nil)
	gradient(t, "circle 5ch, blue",
		"circle", GradientSize{Explicit: Point{{Value: 5, Unit: Ch}, {Value: 5, Unit: Ch}}}, Center{}, nil, nil)

	gradient(t, "10px 50px, blue",
		"ellipse", GradientSize{Explicit: Point{{Value: 10, Unit: Px}, {Value: 50, Unit: Px}}}, Center{}, nil, nil)
	gradient(t, "10px 50px ellipse, blue",
		"ellipse", GradientSize{Explicit: Point{{Value: 10, Unit: Px}, {Value: 50, Unit: Px}}}, Center{}, nil, nil)
	gradient(t, "ellipse 10px 50px, blue",
		"ellipse", GradientSize{Explicit: Point{{Value: 10, Unit: Px}, {Value: 50, Unit: Px}}}, Center{}, nil, nil)

	gradient(t, "10px 50px, blue",
		"ellipse", GradientSize{Explicit: Point{{Value: 10, Unit: Px}, {Value: 50, Unit: Px}}}, Center{}, nil, nil)
	gradient(t, "at top 10% right, blue", "", GradientSize{},
		Center{OriginX: "right", OriginY: "top", Pos: Point{{Value: 0, Unit: Percentage}, {Value: 10, Unit: Percentage}}}, nil, nil)
	gradient(t, "circle at bottom, blue", "circle", GradientSize{},
		Center{OriginX: "left", OriginY: "top", Pos: Point{{Value: 50, Unit: Percentage}, {Value: 100, Unit: Percentage}}}, nil, nil)
	gradient(t, "circle at 10px, blue", "circle", GradientSize{},
		Center{OriginX: "left", OriginY: "top", Pos: Point{{Value: 10, Unit: Px}, {Value: 50, Unit: Percentage}}}, nil, nil)
	gradient(t, "closest-side circle at right 5em, blue",
		"circle", GradientSize{Keyword: "closest-side"},
		Center{OriginX: "left", OriginY: "top", Pos: Point{{Value: 100, Unit: Percentage}, {Value: 5, Unit: Em}}}, nil, nil)
}
