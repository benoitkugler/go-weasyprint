package css

const (
	Top    Side = "top"
	Bottom Side = "bottom"
	Left   Side = "left"
	Right  Side = "right"
)

type Dimension struct {
	Unit  string
	Value float64
}

type Side string

type CounterIncrement struct {
	Name  string
	Value int
}

type CounterIncrements struct {
	Auto bool
	CI   []CounterIncrement
}

// type StyleDict2 struct {
// 	Anonymous bool

// 	Float    string
// 	Position string
// 	Page     int

// 	Margin      map[Side]Dimension
// 	Padding     map[Side]Dimension
// 	BorderWidth map[Side]float64

// 	Direction string

// 	TextTransform, Hyphens string
// 	Display                string

// 	CounterReset []struct {
// 		Name  string
// 		Value int
// 	}
// 	CounterIncrement CounterIncrements
// }

type cascadedValue struct {
	value      interface{}
	precedence int
}

type StyleDict map[string]interface{}

// Deep copy
func (s StyleDict) Copy() StyleDict {
	news := make(StyleDict, len(s))
	for k, v := range s {
		news[k] = v
	}
	return news
}

func (s StyleDict) Anonymous() bool {
	_, in := s["__anonymous"]
	return in
}

func (s StyleDict) SetAnonymous() {
	s["__anonymous"] = true
}

// Get a dict of computed style mixed from parent and cascaded styles.
func computedFromCascaded(element *TBD, cascaded map[string]cascadedValue, parentStyle StyleDict, pseudoType string,
	rootStyle StyleDict, baseUrl string) StyleDict {
	if cascaded == nil && parentStyle != nil {
		// Fast path for anonymous boxes:
		// no cascaded style, only implicitly initial or inherited values.
		computed := InitialValues.Copy()
		for name := range Inherited {
			computed[name] = parentStyle[name]
		}
		// page is not inherited but taken from the ancestor if "auto"
		computed["page"] = parentStyle["page"]
		// border-*-style is none, so border-width computes to zero.
		// Other than that, properties that would need computing are
		// border-*-color, but they do not apply.
		computed["border_top_width"] = 0
		computed["border_bottom_width"] = 0
		computed["border_left_width"] = 0
		computed["border_right_width"] = 0
		computed["outlineWidth"] = 0
		return computed
	}

	// Handle inheritance and initial values
	specified, computed := StyleDict{}, StyleDict{}
	for name, initial := range InitialValues {
		var (
			keyword string
			value   interface{}
		)
		if _, in := cascaded[name]; in {
			vp := cascaded[name]
			keyword = vp.value.(string)
			value = vp.value
		} else {
			if Inherited[name] {
				keyword = "inherit"
			} else {
				keyword = "initial"
			}
		}

		if keyword == "inherit" && parentStyle == nil {
			// On the root element, "inherit" from initial values
			keyword = "initial"
		}

		if keyword == "initial" {
			value = initial
			if !InitialNotComputed[name] {
				// The value is the same as when computed
				computed[name] = value
			}
		} else if keyword == "inherit" {
			value = parentStyle[name]
			// Values in parentStyle are already computed.
			computed[name] = value
		}
		specified[name] = value
	}
	if specified["page"] == "auto" {
		// The page property does not inherit. However, if the page value on
		// an element is auto, then its used value is the value specified on
		// its nearest ancestor with a non-auto value. When specified on the
		// root element, the used value for auto is the empty string.
		val := ""
		if parentStyle != nil {
			val = parentStyle["page"].(string)
		}
		computed["page"] = val
		specified["page"] = val
	}

	return StyleDict(compute(
		element, pseudoType, specified, computed, parentStyle, rootStyle,
		baseUrl))
}
