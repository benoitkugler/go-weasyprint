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

func sToV(s string) Value { return Value{String: s} }

func iToV(i int) Value {
	return Value{Dimension: Dimension{Value: float32(i)}}
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
