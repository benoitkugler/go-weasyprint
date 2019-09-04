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

func sToV(s string) Value  { return Value{String: s} }
func fToV(f float32) Value { return Value{Dimension: Dimension{Value: f}} }
func iToV(i int) Value     { return fToV(float32(i)) }

type CssProperty interface {
	ComputeValue(computer *computer, name string) CssProperty
	Copy() CssProperty
}

type ImageType interface{}

// Dimension or string
type Value struct {
	Dimension
	String string
}

type ContentType int

type ContentProperty struct {
	Type ContentType

	// Next are values fields
	String  string   // for type STRING, URI, attr
	Quote   quote    // for type QUOTE
	Strings []string // for type string, counter, counters
}

func (cp ContentProperty) IsNone() bool {
	return cp.Type == 0
}

type StringContent struct {
	Name   string
	Values []ContentProperty
}

func (s StringContent) IsNone() bool {
	return s.Name == "" && s.Values == nil
}

func (s StringContent) Copy() StringContent {
	out := s
	out.Values = append([]ContentProperty{}, s.Values...)
	return out
}

type Transform struct {
	Function string
	Args     []Dimension
}

func (t Transform) Copy() Transform {
	out := t
	out.Args = append([]Dimension{}, t.Args...)
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

// ----------- Copy implementations --------------------

// deep copy
func (c Content) Copy() CssProperty {
	out := c
	out.List = append([]ContentProperty{}, c.List...)
	return out
}

func (s StringSet) Copy() CssProperty {
	out := s
	out.Contents = make([]StringContent, len(s.Contents))
	for index, l := range s.Contents {
		out.Contents[index] = l.Copy()
	}
	return out
}

func (x CounterIncrements) Copy() CssProperty {
	out := x
	out.CI = append([]IntString{}, x.CI...)
	return out
}

func (x CounterResets) Copy() CssProperty {
	return append(CounterResets{}, x...)
}

func (q Quotes) Copy() CssProperty {
	return Quotes{Open: append([]string{}, q.Open...), Close: append([]string{}, q.Close...)}
}

func (b BackgroundImage) Copy() CssProperty {
	out := make(BackgroundImage, len(b))
	for index, v := range b {
		out[index] = v.Copy()
	}
	return out
}
func (b BackgroundPosition) Copy() CssProperty {
	return append(BackgroundPosition{}, b...)
}
func (b BackgroundSize) Copy() CssProperty {
	return append(BackgroundSize{}, b...)
}
func (b BackgroundRepeat) Copy() CssProperty {
	return append(BackgroundRepeat{}, b...)
}
func (b Lengths) Copy() CssProperty {
	return append(Lengths{}, b...)
}
func (b Transforms) Copy() CssProperty {
	out := make(Transforms, len(b))
	for index, v := range b {
		out[index] = v.Copy()
	}
	return out
}

// value types

func (v Value) Copy() CssProperty          { return v }
func (v Marks) Copy() CssProperty          { return v }
func (v ListStyleImage) Copy() CssProperty { return v }
func (v Break) Copy() CssProperty          { return v }
func (v Display) Copy() CssProperty        { return v }
func (v Floating) Copy() CssProperty       { return v }
func (v String) Copy() CssProperty         { return v }
func (v Length) Copy() CssProperty         { return v }
func (v Bleed) Copy() CssProperty          { return v }
func (v BorderWidth) Copy() CssProperty    { return v }
func (v PixelLength) Copy() CssProperty    { return v }
func (v ColumnWidth) Copy() CssProperty    { return v }
func (v ColumnGap) Copy() CssProperty      { return v }
func (v FontSize) Copy() CssProperty       { return v }
func (v FontWeight) Copy() CssProperty     { return v }
func (v LineHeight) Copy() CssProperty     { return v }
func (v TabSize) Copy() CssProperty        { return v }
func (v VerticalAlign) Copy() CssProperty  { return v }
func (v WordSpacing) Copy() CssProperty    { return v }
func (v Link) Copy() CssProperty           { return v }
func (v Anchor) Copy() CssProperty         { return v }
func (v Lang) Copy() CssProperty           { return v }
func (v WidthHeight) Copy() CssProperty    { return v }
func (v Page) Copy() CssProperty           { return v }
func (v Float) Copy() CssProperty          { return v }
func (v Int) Copy() CssProperty            { return v }
func (v IntString) Copy() CssProperty      { return v }
func (v Dimension) Copy() CssProperty      { return v }
