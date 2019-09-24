package properties

import "github.com/benoitkugler/go-weasyprint/style/parser"

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

func (s AttrData) copy() AttrData {
	out := s
	s.Fallback = s.Fallback.Copy()
	return out
}

// -------------- Images ------------------------

func (i NoneImage) copyAsImage() Image      { return i }
func (i UrlImage) copyAsImage() Image       { return i }
func (i LinearGradient) copyAsImage() Image { return i.copy() }
func (i RadialGradient) copyAsImage() Image { return i.copy() }
func (i NoneImage) Copy() CssProperty       { return i.copyAsImage() }
func (i UrlImage) Copy() CssProperty        { return i.copyAsImage() }
func (i LinearGradient) Copy() CssProperty  { return i.copyAsImage() }
func (i RadialGradient) Copy() CssProperty  { return i.copyAsImage() }

// -------------------------- Content Property --------------------------
// func (i NoneImage) copyAsInnerContent() InnerContent      { return i }
// func (i UrlImage) copyAsInnerContent() InnerContent       { return i }
// func (i LinearGradient) copyAsInnerContent() InnerContent { return i.copy() }
// func (i RadialGradient) copyAsInnerContent() InnerContent { return i.copy() }

// contents,
func (s String) copyAsInnerContent() InnerContent  { return s }
func (s Strings) copyAsInnerContent() InnerContent { return s.copy() }

// target
func (s SContentProps) copyAsInnerContent() InnerContent { return s.copy() }

// url
func (s NamedString) copyAsInnerContent() InnerContent { return s }
func (s Quote) copyAsInnerContent() InnerContent       { return s }
func (s AttrData) copyAsInnerContent() InnerContent    { return s.copy() }

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
