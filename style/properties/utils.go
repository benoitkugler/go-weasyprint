package properties

import (
	"log"

	"github.com/benoitkugler/go-weasyprint/style/parser"
)

var Has = struct{}{}

type Set map[string]struct{}

func (s Set) Add(key string) {
	s[key] = Has
}

func (s Set) Has(key string) bool {
	_, in := s[key]
	return in
}

func (s Set) copy() Set {
	out := make(Set, len(s))
	for k, v := range s {
		out[k] = v
	}
	return out
}

func (s Set) IsNone() bool { return s == nil }

func NewSet(values ...string) Set {
	s := make(Set, len(values))
	for _, v := range values {
		s.Add(v)
	}
	return s
}

// --------------- Values  -----------------------------------------------

func (d Dimension) ToValue() Value {
	return Value{Dimension: d}
}

func FToD(f float32) Dimension { return Dimension{Value: f, Unit: Scalar} }
func SToV(s string) Value      { return Value{String: s} }
func FToV(f float32) Value     { return FToD(f).ToValue() }

func NewColor(r, g, b, a float32) Color {
	return Color{RGBA: parser.RGBA{R: r, G: g, B: b, A: a}, Type: parser.ColorRGBA}
}

// ----------------- misc  ---------------------------------------

func (p Point) ToSlice() []Dimension {
	return []Dimension{p[0], p[1]}
}

func (s GradientSize) IsExplicit() bool {
	return s.Keyword == ""
}

// -------------- Images ------------------------

func (i NoneImage) isCssProperty()      {}
func (i UrlImage) isCssProperty()       {}
func (i LinearGradient) isCssProperty() {}
func (i RadialGradient) isCssProperty() {}
func (i NoneImage) isImage()            {}
func (i UrlImage) isImage()             {}
func (i LinearGradient) isImage()       {}
func (i RadialGradient) isImage()       {}

// -------------------------- Content Property --------------------------
// func (i NoneImage) copyAsInnerContent() InnerContent      { return i }
// func (i UrlImage) copyAsInnerContent() InnerContent       { return i }
// func (i LinearGradient) copyAsInnerContent() InnerContent { return i.copy() }
// func (i RadialGradient) copyAsInnerContent() InnerContent { return i.copy() }

// contents,
func (s String) isInnerContent()  {}
func (s Strings) isInnerContent() {}

// target
func (s SContentProps) isInnerContent() {}

// url
func (s NamedString) isInnerContent() {}
func (s Dimension) isInnerContent()   {}
func (s Float) isInnerContent()       {}
func (s Int) isInnerContent()         {}
func (s Color) isInnerContent()       {}
func (s Quote) isInnerContent()       {}
func (s AttrData) isInnerContent()    {}

func (c ContentProperty) AsCounter() (counterName, counterStyle string) {
	value, _ := c.Content.(Strings)
	if len(value) < 2 {
		log.Fatalf("invalid counter() content : %v", c.Content)
	}
	return value[0], value[1]
}

func (c ContentProperty) AsCounters() (counterName, separator, counterStyle string) {
	value, _ := c.Content.(Strings)
	if len(value) < 3 {
		log.Fatalf("invalid counters() content : %v", c.Content)
	}
	return value[0], value[1], value[2]
}

func (c ContentProperty) AsStrings() []string {
	value, ok := c.Content.(Strings)
	if !ok {
		log.Fatalf("invalid content (expected []string): %v", c.Content)
	}
	return value
}

func (c ContentProperty) AsTargetCounter() (anchorToken ContentProperty, counterName, counterStyle string) {
	value, _ := c.Content.(SContentProps)
	if len(value) != 3 {
		log.Fatalf("invalid content (expected 3-list of String or ContentProperty): %v", c.Content)
	}
	return value[0].ContentProperty, value[1].String, value[2].String
}

func (c ContentProperty) AsTargetCounters() (anchorToken ContentProperty, counterName string, separator ContentProperty, counterStyle string) {
	value, _ := c.Content.(SContentProps)
	if len(value) != 4 {
		log.Fatalf("invalid content (expected 4-list of String or ContentProperty): %v", c.Content)
	}
	return value[0].ContentProperty, value[1].String, value[2].ContentProperty, value[3].String
}

func (c ContentProperty) AsTargetText() (anchorToken ContentProperty, textStyle string) {
	value, _ := c.Content.(SContentProps)
	if len(value) != 2 {
		log.Fatalf("invalid content (expected 2-list of String or ContentProperty): %v", c.Content)
	}
	return value[0].ContentProperty, value[1].String
}

func (c ContentProperty) AsQuotes() Quote {
	return c.Content.(Quote)
}

// ------------------------- Usefull for test ---------------------------

func (bs Images) Repeat(n int) CssProperty {
	var out Images
	for i := 0; i < n; i++ {
		out = append(out, bs...)
	}
	return out
}

func (bs Centers) Repeat(n int) CssProperty {
	var out Centers
	for i := 0; i < n; i++ {
		out = append(out, bs...)
	}
	return out
}
func (bs Sizes) Repeat(n int) CssProperty {
	var out Sizes
	for i := 0; i < n; i++ {
		out = append(out, bs...)
	}
	return out
}
func (bs Repeats) Repeat(n int) CssProperty {
	var out Repeats
	for i := 0; i < n; i++ {
		out = append(out, bs...)
	}
	return out
}
func (bs Strings) Repeat(n int) CssProperty {
	var out Strings
	for i := 0; i < n; i++ {
		out = append(out, bs...)
	}
	return out
}
