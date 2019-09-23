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
	Copy2() CascadedProperty
}

type CssProperty interface {
	CascadedProperty
	Copy3() CssProperty
}

type Inherit struct{}
type Initial struct{}

func (v Inherit) Copy() ValidatedProperty { return v }
func (v Initial) Copy() ValidatedProperty { return v }

type VarData struct {
	Name        string // name of a custom property
	Declaration CustomProperty
}

type AttrData struct {
	Name       string
	TypeOrUnit string
	Fallback   CssProperty
}

func (v VarData) Copy2() CascadedProperty {
	out := v
	out.Declaration = v.Declaration.copy()
	return out
}
func (v AttrData) Copy2() CascadedProperty {
	out := v
	out.Fallback = v.Fallback.Copy3()
	return out
}
func (v CustomProperty) Copy2() CascadedProperty {
	return v.copy()
}

func (v VarData) Copy() ValidatedProperty        { return v.Copy2() }
func (v AttrData) Copy() ValidatedProperty       { return v.Copy2() }
func (v CustomProperty) Copy() ValidatedProperty { return v.Copy2() }

// Properties is the general container for validated, cascaded and computed properties.
// In addition to the generic acces, an attempt to provide a "type safe" way is provided through the
// GetXXX and SetXXX methods. It relies on the convention than all the keys should be present,
// and values never be nil.
// "None" values are then encoded by the zero value of the concrete type.
type Properties map[string]CssProperty
