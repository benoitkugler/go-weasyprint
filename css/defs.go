package css

import (
	"github.com/benoitkugler/go-weasyprint/css/parser"
	"github.com/benoitkugler/go-weasyprint/utils"
)

const ( // zero field corresponds to null content
	Scalar Unit = iota + 1 // means no unit, but a valid value
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
	// Copy implements the deep copy of the property
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

type Url utils.Url

type AttrFallback struct {
	Name  string
	Value CssProperty
}

func (a AttrFallback) copy() AttrFallback {
	a.Value = a.Value.Copy() // deep copy
	return a
}

type Attr struct {
	Name, TypeOrUnit string
	Fallback         AttrFallback
}

func (a Attr) IsNone() bool {
	return a == Attr{}
}

type NamedTokens struct {
	Name   string
	Tokens []parser.Token
}

type SContentProp struct {
	String          string
	ContentProperty ContentProperty
}
type SContentProps []SContentProp

// guard for possible content properties
type InnerContents interface {
	copyAsContent() InnerContents
}

func (cp ContentProperty) IsNone() bool {
	return cp.Type == "" && cp.Content == nil
}

func (cp ContentProperty) copy() ContentProperty {
	out := cp
	out.Content = cp.Content.copyAsContent()
	return out
}

func (cp ContentProperty) Copy() CssProperty {
	return cp.copy()
}

func (s String) copyAsContent() InnerContents      { return s }
func (s NamedString) copyAsContent() InnerContents { return s }
func (s Strings) copyAsContent() InnerContents     { return s.copy() }
func (s Quote) copyAsContent() InnerContents       { return s }
func (s Url) copyAsContent() InnerContents         { return s }
func (s Attr) copyAsContent() InnerContents {
	s.Fallback = s.Fallback.copy()
	return s
}
func (s NamedTokens) copyAsContent() InnerContents {
	s.Tokens = append([]parser.Token{}, s.Tokens...)
	return s
}
func (s SContentProps) copyAsContent() InnerContents {
	out := make(SContentProps, len(s))
	for i, v := range s {
		out[i] = v
		out[i].ContentProperty = v.ContentProperty.copy()
	}
	return out
}

func (s SContent) IsNone() bool {
	return s.String == "" && s.Contents == nil
}

func (s SContent) copy() SContent {
	out := s
	out.Contents = make([]ContentProperty, len(s.Contents))
	for index, v := range s.Contents {
		out.Contents[index] = v.copy()
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
	InnerContents
	copy() Image
	Copy() CssProperty
}

type NoneImage struct{}
type UrlImage string

func (_ NoneImage) copy() Image {
	return NoneImage{}
}
func (_ NoneImage) copyAsContent() InnerContents {
	return NoneImage{}
}
func (s UrlImage) copy() Image {
	return s
}
func (s UrlImage) copyAsContent() InnerContents {
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
func (l LinearGradient) copyAsContent() InnerContents {
	return l.copy()
}

func (r RadialGradient) copy() Image {
	out := r
	out.ColorStops = append([]ColorStop{}, r.ColorStops...)
	return out
}
func (l RadialGradient) copyAsContent() InnerContents {
	return l.copy()
}

func (l LinearGradient) Copy() CssProperty {
	return l.copy()
}

func (r RadialGradient) Copy() CssProperty {
	return r.copy()
}

func SToSS(s string) SStrings {
	return SStrings{String: s}
}

func (ss SStrings) copy() SStrings {
	out := ss
	out.Strings = append([]string{}, ss.Strings...)
	return out
}
func (ss SStrings) IsNone() bool {
	return ss.String == "" && ss.Strings == nil
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
func (b Strings) copy() Strings {
	return append(Strings{}, b...)
}
func (b Strings) Copy() CssProperty {
	return b.copy()
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
func (p Center) Copy() CssProperty      { return p }
