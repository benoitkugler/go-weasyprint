// This package defines the types needed to handle the various CSS properties.
// It is an attempt to provide a (type) safe way of manipulating them.
// To do so, we have to respect the following convention : a property will always have
// one of the following concrete types :
//	- nil : means no value provided
//	- Inherited or Initial : special type matching CSS "inherited" and "initial" keywords
//	- the corresponding type, defined in InitialValues
package properties

const (
	None CssKind = iota
	Inherited
	Initial
	Normal
)

type CssKind uint8

// Copy implements CssProperty
func (c CssKind) Copy() CssProperty { return c }

type CssProperty interface {
	// Copy implements the deep copy of the property
	Copy() CssProperty
}

// Properties is the general container for validated and typed properties.
// In addition to the generic acces, a type safe way is provided through the
// GetXXX and SetXXX methods.
type Properties map[string]CssProperty

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
