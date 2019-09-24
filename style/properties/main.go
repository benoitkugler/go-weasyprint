// This package defines the types needed to handle the various CSS properties.
// There are 3 groups of types for a property, separated by 2 steps : cascading and computation.
// Thus the need of 3 types (see below).
// Schematically, the style computation is :
//		ValidatedProperty (ComputedFromCascaded)-> CascadedPropery (Compute)-> CssProperty
package properties

const (
	Inherit defaultKind = iota + 1
	Initial
)

// CssProperty is final form of a css input :
// "var()", "attr()" and custom properties have been resolved.
type CssProperty interface {
	// Copy implements the deep copy of the property
	Copy() CssProperty
}

// CascadedProperty may contain either a classic CSS property
// or one the 3 special values var(), attr() or custom properties.
// "initial" and "inherited" values have been resolved
type CascadedProperty struct {
	prop            CssProperty
	SpecialProperty specialProperty
}

// AsCss will panic if c.SpecialProperty is not nil.
func (c CascadedProperty) AsCss() CssProperty {
	if c.SpecialProperty != nil {
		panic("attempted to bypass the SpecialProperty of a CascadedProperty")
	}
	return c.prop
}

func (c CascadedProperty) IsNone() bool {
	return c.prop == nil && c.SpecialProperty == nil
}

// ValidatedProperty is valid css input, so it may contain
// a classic property, a special one, or one of the keyword "inherited" or "initial".
type ValidatedProperty struct {
	prop    CascadedProperty
	Default defaultKind
}

// AsCascaded will panic if c.Default is not zero.
func (c ValidatedProperty) AsCascaded() CascadedProperty {
	if c.Default != 0 {
		panic("attempted to bypass the Default of a ValidatedProperty")
	}
	return c.prop
}

type specialProperty interface {
	isSpecialProperty()
}

type defaultKind uint8

func (d defaultKind) ToV() ValidatedProperty {
	return ValidatedProperty{Default: d}
}

type VarData struct {
	Name        string // name of a custom property
	Declaration CustomProperty
}

func (v VarData) IsNone() bool {
	return v.Name == "" && v.Declaration == nil
}

type AttrData struct {
	Name       string
	TypeOrUnit string
	Fallback   CssProperty
}

func (a AttrData) IsNone() bool {
	return a.Name == "" && a.TypeOrUnit == "" && a.Fallback == nil
}

func (v VarData) isSpecialProperty()        {}
func (v AttrData) isSpecialProperty()       {}
func (v CustomProperty) isSpecialProperty() {}

// ---------- Convenience constructor -------------------------------
// Note than a CssProperty can naturally be seen as a CascadedProperty, but not the other way around.

func ToC(prop CssProperty) CascadedProperty {
	return CascadedProperty{prop: prop}
}

func ToC2(spe specialProperty) CascadedProperty {
	return CascadedProperty{SpecialProperty: spe}
}

func (c CascadedProperty) ToV() ValidatedProperty {
	return ValidatedProperty{prop: c}
}

// Properties is the general container for validated, cascaded and computed properties.
// In addition to the generic acces, an attempt to provide a "type safe" way is provided through the
// GetXXX and SetXXX methods. It relies on the convention than all the keys should be present,
// and values never be nil.
// "None" values are then encoded by the zero value of the concrete type.
type Properties map[string]CssProperty
