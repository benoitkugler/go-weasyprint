package css

import "github.com/benoitkugler/go-weasyprint/css/parser"

const (
	ContentQUOTE ContentType = iota + 1 // so that zero field corresponds to null content
	ContentSTRING
	ContentURI
	ContentAttr
	ContentCounter
	ContentCounters
	ContentString
	ContentContent
)

const (
	NoUnit Unit = iota // means no value
	Scalar             // means no unit, but a valid value
	Percentage
	Ex
	Em
	Ch
	Rem
	Px
	Pt
	Pc
	In
	Cm
	Mm
	Q
)

var ZeroPixels = Dimension{Unit: Px}

var Has = struct{}{}

type Set map[string]struct{}

func (s Set) Add(key string) {
	s[key] = Has
}

func (s Set) Has(key string) bool {
	_, in := s[key]
	return in
}

func NewSet(values ...string) Set {
	s := make(Set, len(values))
	for _, v := range values {
		s.Add(v)
	}
	return s
}

type CssProperty interface {
	Copy() CssProperty
}

func (p Properties) Copy() Properties {
	out := make(Properties, len(p))
	for name, v := range p {
		out[name] = v.Copy()
	}
	return out
}

func (p Properties) Keys() []string {
	keys := make([]string, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}
	return keys
}

func (s String) String() string {
	return string(s)
}

// ------------ Helpers type --------------------------

type Unit uint8

// Dimension without unit is interpreted as float
type Dimension struct {
	Unit  Unit
	Value float32
}

func (d Dimension) IsNone() bool {
	return d == Dimension{}
}

func (d Dimension) ToValue() Value {
	return Value{Dimension: d}
}

func FToD(f float32) Dimension { return Dimension{Value: f, Unit: Scalar} }
func SToV(s string) Value      { return Value{String: s} }
func FToV(f float32) Value     { return FToD(f).ToValue() }

func NewColor(r, g, b, a float32) Color {
	return Color{RGBA: parser.RGBA{R: r, G: g, B: b, A: a}, Type: parser.ColorRGBA}
}

func (p Point) IsNone() bool {
	return p == Point{}
}

func (p Point) ToSlice() []Dimension {
	return []Dimension{p[0], p[1]}
}

func (s Size) IsNone() bool {
	return s == Size{}
}

func (c Center) IsNone() bool {
	return c == Center{}
}

type Quote struct {
	Open, Insert bool
}

type ContentType int

type ContentProperty struct {
	Type ContentType

	// Next are values fields
	SStrings       // for type STRING, URI, attr or string, counter, counters
	Quote    Quote // for type QUOTE
}

func (cp ContentProperty) IsNone() bool {
	return cp.Type == 0
}

func (cp ContentProperty) Copy() ContentProperty {
	out := cp
	out.SStrings = cp.SStrings.copy()
	return out
}

func (s SContent) IsNone() bool {
	return s.String == "" && s.Contents == nil
}

func (s SContent) copy() SContent {
	out := s
	out.Contents = make([]ContentProperty, len(s.Contents))
	for index, v := range s.Contents {
		out.Contents[index] = v.Copy()
	}
	return out
}
func (s SContent) Copy() CssProperty {
	return s.copy()
}

func (t SDimensions) Copy() SDimensions {
	out := t
	out.Dimensions = append([]Dimension{}, t.Dimensions...)
	return out
}

// Might be an existing image or a gradient
type Image interface {
	copy() Image
	Copy() CssProperty
}

type NoneImage struct{}
type UrlImage string

func (_ NoneImage) copy() Image {
	return NoneImage{}
}
func (s UrlImage) copy() Image {
	return s
}
func (i NoneImage) Copy() CssProperty {
	return i.copy()
}
func (i UrlImage) Copy() CssProperty {
	return i.copy()
}

type ColorStop struct {
	Color    Color
	Position Dimension
}

type DirectionType struct {
	Corner string
	Angle  float32
}

type GradientSize struct {
	Keyword  string
	Explicit Point
}

func (s GradientSize) IsNone() bool {
	return s == GradientSize{}
}

func (s GradientSize) IsExplicit() bool {
	return s.Keyword == ""
}

type LinearGradient struct {
	ColorStops []ColorStop
	Direction  DirectionType
	Repeating  bool
}

type RadialGradient struct {
	ColorStops []ColorStop
	Shape      string
	Size       GradientSize
	Center     Center
	Repeating  bool
}

func (l LinearGradient) copy() Image {
	out := l
	out.ColorStops = append([]ColorStop{}, l.ColorStops...)
	return out
}

func (r RadialGradient) copy() Image {
	out := r
	out.ColorStops = append([]ColorStop{}, r.ColorStops...)
	return out
}

func (l LinearGradient) Copy() CssProperty {
	return l.copy()
}

func (r RadialGradient) Copy() CssProperty {
	return r.copy()
}

func (ss SStrings) copy() SStrings {
	out := ss
	out.Strings = append([]string{}, ss.Strings...)
	return out
}
func (ss SStrings) Copy() CssProperty {
	return ss.copy()
}

func (s StringSet) Copy() CssProperty {
	out := s
	out.Contents = make([]SContent, len(s.Contents))
	for index, l := range s.Contents {
		out.Contents[index] = l.copy()
	}
	return out
}

func (l BookmarkLabel) Copy() CssProperty {
	return append(BookmarkLabel{}, l...)
}

func (l IntStrings) Copy() CssProperty {
	return append(IntStrings{}, l...)
}

func (x SIntStrings) Copy() CssProperty {
	out := x
	out.Values = append([]IntString{}, x.Values...)
	return out
}
func (x NDecorations) Copy() CssProperty {
	out := x
	out.Decorations = Set{}
	for dec := range x.Decorations {
		out.Decorations[dec] = Has
	}
	return out
}

func (q Quotes) Copy() CssProperty {
	return Quotes{Open: append([]string{}, q.Open...), Close: append([]string{}, q.Close...)}
}

func (b Images) Copy() CssProperty {
	out := make(Images, len(b))
	for index, v := range b {
		out[index] = v.copy()
	}
	return out
}
func (b Centers) Copy() CssProperty {
	return append(Centers{}, b...)
}
func (b Sizes) Copy() CssProperty {
	return append(Sizes{}, b...)
}
func (b Repeats) Copy() CssProperty {
	return append(Repeats{}, b...)
}
func (b Values) Copy() CssProperty {
	return append(Values{}, b...)
}
func (b Strings) Copy() CssProperty {
	return append(Strings{}, b...)
}
func (b Transforms) Copy() CssProperty {
	out := make(Transforms, len(b))
	for index, v := range b {
		out[index] = v.Copy()
	}
	return out
}

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

// ------------------ Value type -----------------
func (p Float) Copy() CssProperty       { return p }
func (p Int) Copy() CssProperty         { return p }
func (p Ints3) Copy() CssProperty       { return p }
func (p Page) Copy() CssProperty        { return p }
func (p NamedString) Copy() CssProperty { return p }
func (p Marks) Copy() CssProperty       { return p }
func (p IntString) Copy() CssProperty   { return p }
func (p String) Copy() CssProperty      { return p }
func (p Value) Copy() CssProperty       { return p }
func (p Color) Copy() CssProperty       { return p }
func (p Point) Copy() CssProperty       { return p }
