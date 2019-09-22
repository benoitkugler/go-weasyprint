package validation

import (
	"reflect"
	"testing"

	. "github.com/benoitkugler/go-weasyprint/css"
	"github.com/benoitkugler/go-weasyprint/css/parser"
	"github.com/benoitkugler/go-weasyprint/utils"
)

func processFontFace(css string, t *testing.T) []NamedDescriptor {
	stylesheet := parser.ParseStylesheet2([]byte(css), false, false)
	atRule, ok := stylesheet[0].(parser.AtRule)
	if !ok || atRule.AtKeyword != "font-face" {
		t.Fatalf("expected @font-face got %v", stylesheet[0])
	}
	tokens := parser.ParseDeclarationList(*atRule.Content, false, false)
	return PreprocessDescriptors("http://weasyprint.org/foo/", tokens)
}

func checkNameDescriptor(ref, got NamedDescriptor, t *testing.T) {
	if !reflect.DeepEqual(ref, got) {
		t.Fatalf("expected %v got %v", ref, got)
	}
}

// Test the ``font-face`` rule.
func TestFontFace(t *testing.T) {
	l := processFontFace(`@font-face {
		font-family: Gentium Hard;
		src: url(http://example.com/fonts/Gentium.woff);
	  }`, t)
	fontFamily, src := l[0], l[1]
	checkNameDescriptor(NamedDescriptor{Name: "font_family", Descriptor: String("Gentium Hard")}, fontFamily, t)
	checkNameDescriptor(NamedDescriptor{Name: "src", Descriptor: Contents{NamedString{Name: "external", String: "http://example.com/fonts/Gentium.woff"}}}, src, t)

	l = processFontFace(`@font-face {
          font-family: "Fonty Smiley";
         src: url(Fonty-Smiley.woff);
          font-style: italic;
         font-weight: 200;
         font-stretch: condensed;
        }`, t)
	fontFamily, src, fontStyle, fontWeight, fontStretch := l[0], l[1], l[2], l[3], l[4]
	checkNameDescriptor(NamedDescriptor{Name: "font_family", Descriptor: String("Fonty Smiley")}, fontFamily, t)
	checkNameDescriptor(NamedDescriptor{Name: "src", Descriptor: Contents{NamedString{Name: "external", String: "http://weasyprint.org/foo/Fonty-Smiley.woff"}}}, src, t)
	checkNameDescriptor(NamedDescriptor{Name: "font_style", Descriptor: String("italic")}, fontStyle, t)
	checkNameDescriptor(NamedDescriptor{Name: "font_weight", Descriptor: IntString{Int: 200}}, fontWeight, t)
	checkNameDescriptor(NamedDescriptor{Name: "font_stretch", Descriptor: String("condensed")}, fontStretch, t)

	l = processFontFace(`@font-face {
		font-family: Gentium Hard;
		src: local();
        }`, t)
	fontFamily, src = l[0], l[1]
	checkNameDescriptor(NamedDescriptor{Name: "font_family", Descriptor: String("Gentium Hard")}, fontFamily, t)
	checkNameDescriptor(NamedDescriptor{Name: "src", Descriptor: Contents{NamedString{Name: "local", String: ""}}}, src, t)

	// See bug #487
	l = processFontFace(`@font-face {
		font-family: Gentium Hard;
		src: local(Gentium Hard);
        }`, t)
	fontFamily, src = l[0], l[1]
	checkNameDescriptor(NamedDescriptor{Name: "font_family", Descriptor: String("Gentium Hard")}, fontFamily, t)
	checkNameDescriptor(NamedDescriptor{Name: "src", Descriptor: Contents{NamedString{Name: "local", String: "Gentium Hard"}}}, src, t)
}

// Test bad ``font-face`` rules.
func TestBadFontFace(t *testing.T) {
	logs := utils.CaptureLogs()
	l := processFontFace(`@font-face {`+
		`  font-family: "Bad Font";`+
		`  src: url(BadFont.woff);`+
		`  font-stretch: expanded;`+
		`  font-style: wrong;`+
		`  font-weight: bolder;`+
		`  font-stretch: wrong;`+
		`}`, t)
	fontFamily, src, fontStretch := l[0], l[1], l[2]

	checkNameDescriptor(NamedDescriptor{Name: "font_family", Descriptor: String("Bad Font")}, fontFamily, t)
	checkNameDescriptor(NamedDescriptor{Name: "src", Descriptor: Contents{NamedString{Name: "external", String: "http://weasyprint.org/foo/BadFont.woff"}}}, src, t)
	checkNameDescriptor(NamedDescriptor{Name: "font_stretch", Descriptor: String("expanded")}, fontStretch, t)

	logs.CheckEqual([]string{
		"Ignored `font-style: wrong` at 1:91, invalid or unsupported values for a known CSS property.",
		"Ignored `font-weight: bolder` at 1:111, invalid or unsupported values for a known CSS property.",
		"Ignored `font-stretch: wrong` at 1:133, invalid or unsupported values for a known CSS property.",
	}, t)
}

// see style/style_test.go for other font face tests
