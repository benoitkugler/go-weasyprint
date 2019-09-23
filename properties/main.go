// This package defines the types needed to handle the various CSS properties.
// There are 3 groups of types for a property, separated by 2 steps : cascading and computation.
// Thus, we use 3 narrowing interfaces :
//	- ValidatedProperty : all valid css inputs
//	- CascadedProperty : "initial" and "inherited" are resolved
// 	- CssProperty : final form, "var()", "attr()" and custom properties are resolved
//
// Schematically, the style computation is :
//		ValidatedProperty (ComputedFromCascaded)-> CascadedPropery (Compute)-> Property
package properties

type ValidatedProperty interface {
	// Copy implements the deep copy of the property
	Copy() ValidatedProperty
}

type CascadedProperty interface {
	ValidatedProperty
	afterCascaded()
}

type CssProperty interface {
	CascadedProperty
	afterCompute()
}

type Inherit struct{}

type Initial struct{}

type VarData struct {
	Name        string // name of a custom property
	Declaration CustomProperty
}

type AttrData struct {
	Name       string
	TypeOrUnit string
	Fallback   CssProperty
}

type CustomProperty []parser.Token

func (VarData) afterCascaded()        {}
func (AttrData) afterCascaded()       {}
func (CustomProperty) afterCascaded() {}

// Properties is the general container for validated, cascaded and computed properties.
// In addition to the generic acces, an attempt to provide a "type safe" way is provided through the
// GetXXX and SetXXX methods. It relies on the convention than all the keys should be present,
// and values never be nil.
// "None" values are then encoded by the zero value of the concrete type.
type Properties map[string]CssProperty
