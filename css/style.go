package css

import (
	"golang.org/x/net/html"
)

const (
	Top    Side = "top"
	Bottom Side = "bottom"
	Left   Side = "left"
	Right  Side = "right"
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

type Unit uint8

// Dimension without unit is interpreted as float
type Dimension struct {
	Unit  Unit
	Value float32
}

func (d Dimension) IsNone() bool {
	return d == Dimension{}
}

func toDim(v float32) Dimension {
	return Dimension{Value: v}
}

type Point struct {
	X, Y Dimension
}

func (t Point) IsNone() bool {
	return t == Point{}
}

type TransformOrigin Point

type Side string

type NameValue struct {
	Name, Value string
}

type FontVariant struct {
	String string
	Values []string
}

func (f FontVariant) IsNone() bool {
	return f.String == "" && f.Values == nil
}

type IntString struct {
	Name  string
	Value int
}

type cascadedValue struct {
	value      string
	precedence int
}

type StyleDict struct {

	Anonymous bool
	Strings   map[string]string
	Values    map[string]Value
	Links     map[string]Link
	Lengthss  map[string]Lengths
	Colors    map[string]Color

	inheritedStyle *StyleDict
}

func NewStyleDict() StyleDict {
	var out StyleDict
	out.Strings = make(map[string]string)
	out.Values = make(map[string]Value)
	out.Links = make(map[string]Link)
	out.Lengthss = make(map[string]Lengths)
	out.Colors = make(map[string]Color)
	return out
}

// IsZero returns `true` if the StyleDict is not initialized.
// Thus, we can use a zero StyleDict as null value.
func (s StyleDict) IsZero() bool {
	return s.Strings == nil
}

// Deep copy
func (s StyleDict) Copy() StyleDict {
	out := s
	out.MiscProperties = s.MiscProperties.Copy()
	out.Strings = make(map[string]string, len(s.Strings))
	out.Values = make(map[string]Value, len(s.Values))
	out.Links = make(map[string]Link, len(s.Links))
	out.Lengthss = make(map[string]Lengths, len(s.Lengthss))
	out.Colors = make(map[string]Color, len(s.Colors))
	for k, v := range s.Strings {
		out.Strings[k] = v
	}
	for k, v := range s.Values {
		out.Values[k] = v
	}
	for k, v := range s.Links {
		out.Links[k] = v
	}
	for k, v := range s.Lengthss {
		out.Lengthss[k] = v
	}
	for k, v := range s.Colors {
		out.Colors[k] = v
	}
	return out
}

func (s StyleDict) Items() map[string]CssProperty {
	out := make(map[string]CssProperty)
	for k, v := range s.Strings {
		fn := ConvertersString[k]
		if fn != nil {
			out[k] = fn(v)
		} else {
			out[k] = String(v)
		}
	}
	for k, v := range s.Values {
		fn := ConvertersValue[k]
		if fn != nil {
			out[k] = fn(v)
		} else {
			out[k] = v
		}
	}
	for k, v := range s.Lengthss {
		out[k] = v
	}
	for k, v := range s.Colors {
		out[k] = v
	}
	for k, v := range s.Links {
		fn := ConvertersLink[k]
		if fn != nil {
			out[k] = fn(v)
		} else {
			out[k] = v
		}
	}
	for k, v := range s.MiscProperties.Items() {
		out[k] = v
	}
	return out
}

func (s StyleDict) Keys() []string {
	items := s.Items()
	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, k)
	}
	return keys
}

// InheritFrom returns a new StyleDict with inherited properties from this one.
// Non-inherited properties get their initial values.
// This is the method used for an anonymous box.
func (s *StyleDict) InheritFrom() StyleDict {
	if s.inheritedStyle == nil {
		is := computedFromCascaded(html.Node{}, nil, *s, StyleDict{}, "")
		is.Anonymous = true
		s.inheritedStyle = &is
	}
	return *s.inheritedStyle
}

func (s StyleDict) GetColor(key string) Color {
	value := s.Colors[key]
	if value.String == "currentColor" {
		value = s.Colors["color"]
	}
	return value
}

// Get a dict of computed style mixed from parent and cascaded styles.
func computedFromCascaded(element html.Node, cascaded map[string]cascadedValue, parentStyle StyleDict,
	rootStyle StyleDict, baseUrl string) StyleDict {
	if cascaded == nil && !parentStyle.IsZero() {
		// Fast path for anonymous boxes:
		// no cascaded style, only implicitly initial or inherited values.
		computed := InitialValues.Copy()
		parentStyleItems := parentStyle.Items()
		for key := range Inherited {
			parentStyleItems[key].SetOn(key, &computed)
		}

		// page is not inherited but taken from the ancestor if "auto"
		computed.Page = parentStyle.Page
		// border-*-style is none, so border-width computes to zero.
		// Other than that, properties that would need computing are
		// border-*-color, but they do not apply.
		computed.Values["border_top_width"] = Value{}
		computed.Values["border_bottom_width"] = Value{}
		computed.Values["border_left_width"] = Value{}
		computed.Values["border_right_width"] = Value{}
		computed.Values["outline_width"] = Value{}
		return computed
	}

	// Handle inheritance and initial values
	specified, computed := NewStyleDict(), NewStyleDict()
	parentItems := parentStyle.Items()
	for name, initial := range InitialValues.Items() {
		var (
			keyword string
			value   CssProperty
		)
		if _, in := cascaded[name]; in {
			vp := cascaded[name]
			keyword = vp.value
			value = Value{String: vp.value}
		} else {
			if Inherited[name] {
				keyword = "inherit"
			} else {
				keyword = "initial"
			}
		}

		if keyword == "inherit" && parentStyle.IsZero() {
			// On the root element, "inherit" from initial values
			keyword = "initial"
		}

		if keyword == "initial" {
			value = initial
			if !InitialNotComputed[name] {
				// The value is the same as when computed
				initial.SetOn(name, &computed)
			}
		} else if keyword == "inherit" {
			value = parentItems[name]
			// Values in parentStyle are already computed.
			value.SetOn(name, &computed)
		}
		value.SetOn(name, &specified)
	}
	if specified.Page.String == "auto" {
		// The page property does not inherit. However, if the page value on
		// an element is auto, then its used value is the value specified on
		// its nearest ancestor with a non-auto value. When specified on the
		// root element, the used value for auto is the empty string.
		val := Page{Valid: true, String: ""}
		if !parentStyle.IsZero() {
			val = parentStyle.Page
		}
		computed.Page = val
		specified.Page = val
	}

	return compute(element, specified, computed, parentStyle, rootStyle, baseUrl)
}
