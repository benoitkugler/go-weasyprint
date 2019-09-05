package css

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
	NoUnit Unit = iota
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

type CssProperty interface {
	Copy() CssProperty
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

func sToV(s string) Value  { return Value{String: s} }
func fToV(f float32) Value { return Value{Dimension: Dimension{Value: f}} }
func iToV(i int) Value     { return fToV(float32(i)) }

type ContentType int

type ContentProperty struct {
	Type ContentType

	// Next are values fields
	SStrings       // for type STRING, URI, attr or string, counter, counters
	Quote    quote // for type QUOTE
}

func (cp ContentProperty) IsNone() bool {
	return cp.Type == 0
}

func (cp ContentProperty) Copy() ContentProperty {
	out := cp
	out.SStrings = cp.SStrings.Copy()
	return out
}

func (s SContent) IsNone() bool {
	return s.String == "" && s.Contents == nil
}

func (s SContent) Copy() CssProperty {
	out := s
	out.Contents = make([]ContentProperty, len(s.Contents))
	for index, v := range s.Contents {
		out.Contents[index] = v.Copy()
	}
	return out
}

func (t SDimensions) Copy() SDimensions {
	out := t
	out.Dimensions = append([]Dimension{}, t.Dimensions...)
	return out
}

type NoneImage struct{}
type UrlImage string

func (_ NoneImage) isImage() {}
func (_ UrlImage) isImage()  {}

func (_ NoneImage) Copy() Image {
	return NoneImage{}
}
func (s UrlImage) Copy() Image {
	return s
}

func (ss SStrings) Copy() SStrings {
	out := ss
	out.Strings = append([]string{}, ss.Strings...)
	return out
}

func (s StringSet) Copy() CssProperty {
	out := s
	out.Contents = make([]SStrings, len(s.Contents))
	for index, l := range s.Contents {
		out.Contents[index] = l.Copy()
	}
	return out
}

func (x SIntStrings) Copy() CssProperty {
	out := x
	out.Values = append([]IntString{}, x.Values...)
	return out
}
func (q Quotes) Copy() CssProperty {
	return Quotes{Open: append([]string{}, q.Open...), Close: append([]string{}, q.Close...)}
}

func (b Images) Copy() CssProperty {
	out := make(Images, len(b))
	for index, v := range b {
		out[index] = v.Copy()
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
