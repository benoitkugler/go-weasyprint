package style

import (
	"log"

	"github.com/benoitkugler/go-weasyprint/css"
	"golang.org/x/net/html"
)

type StyleDict struct {
	css.Properties
	Anonymous      bool
	inheritedStyle *StyleDict
}

func NewStyleDict() StyleDict {
	return StyleDict{Properties: css.Properties{}}
}

// IsZero returns `true` if the StyleDict is not initialized.
// Thus, we can use a zero StyleDict as null value.
func (s StyleDict) IsZero() bool {
	return s.Properties == nil
}

// Deep copy.
// inheritedStyle is a shallow copy
func (s StyleDict) Copy() StyleDict {
	out := s
	out.Properties = s.Properties.Copy()
	return out
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

func (s StyleDict) ResolveColor(key string) css.Color {
	value := s.Properties[key].(css.Color)
	if value.Type == css.ColorCurrentColor {
		value = s.GetColor()
	}
	return value
}

// Get a dict of computed style mixed from parent and cascaded styles.
func computedFromCascaded(element html.Node, cascaded map[string]css.IntString, parentStyle StyleDict,
	rootStyle StyleDict, baseUrl string) StyleDict {
	if cascaded == nil && !parentStyle.IsZero() {
		// Fast path for anonymous boxes:
		// no cascaded style, only implicitly initial or inherited values.
		computed := css.InitialValues.Copy()
		for key := range css.Inherited {
			computed[key] = parentStyle.Properties[key]
		}

		// page is not inherited but taken from the ancestor if "auto"
		computed.SetPage(parentStyle.GetPage())
		// border-*-style is none, so border-width computes to zero.
		// Other than that, properties that would need computing are
		// border-*-color, but they do not apply.
		computed.SetBorderTopWidth(css.Value{})
		computed.SetBorderBottomWidth(css.Value{})
		computed.SetBorderLeftWidth(css.Value{})
		computed.SetBorderRightWidth(css.Value{})
		computed.SetOutlineWidth(css.Value{})
		return StyleDict{Properties: computed}
	}

	// Handle inheritance and initial values
	specified, computed := NewStyleDict(), NewStyleDict()
	for name, initial := range css.InitialValues {
		var (
			keyword string
			value   css.CssProperty
		)
		if _, in := cascaded[name]; in {
			vp := cascaded[name]
			keyword = vp.String
			value = css.Value{String: vp.String}
		} else {
			if css.Inherited.Has(name) {
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
			if !css.InitialNotComputed.Has(name) {
				// The value is the same as when computed
				computed.Properties[name] = initial
			}
		} else if keyword == "inherit" {
			value = parentStyle.Properties[name]
			// Values in parentStyle are already computed.
			computed.Properties[name] = value
		}
		specified.Properties[name] = value
	}
	if specified.GetPage().String == "auto" {
		// The page property does not inherit. However, if the page value on
		// an element is auto, then its used value is the value specified on
		// its nearest ancestor with a non-auto value. When specified on the
		// root element, the used value for auto is the empty string.
		val := css.Page{Valid: true, String: ""}
		if !parentStyle.IsZero() {
			val = parentStyle.GetPage()
		}
		computed.SetPage(val)
		specified.SetPage(val)
	}

	return compute(element, specified, computed, parentStyle, rootStyle, baseUrl)
}

type page struct {
	side         string
	blank, first bool
	name         string
}

func matchingPageTypes(pageType page, names []string) (out []page) {
	sides := []string{"left", "right", ""}
	if pageType.side != "" {
		sides = []string{pageType.side}
	}

	blanks := []bool{true}
	if pageType.blank == false {
		blanks = []bool{true, false}
	}
	firsts := []bool{true}
	if pageType.first == false {
		firsts = []bool{true, false}
	}
	names = append(names, "")
	if pageType.name != "" {
		names = []string{pageType.name}
	}
	for _, side := range sides {
		for _, blank := range blanks {
			for _, first := range firsts {
				for _, name := range names {
					out = append(out, page{side: side, blank: blank, first: first, name: name})
				}
			}
		}
	}
	return
}

// Return the boolean evaluation of `queryList` for the given
// `deviceMediaType`.
func evaluateMediaQuery(queryList []string, deviceMediaType string) bool {
	// TODO: actual support for media queries, not just media types
	for _, query := range queryList {
		if "all" == query || deviceMediaType == query {
			return true
		}
	}
	return false
}

// Return the precedence for a declaration.
// Precedence values have no meaning unless compared to each other.
// Acceptable values for ``origin`` are the strings ``"author"``, ``"user"``
// and ``"user agent"``.
//
func declarationPrecedence(origin string, importance bool) uint8 {
	// See http://www.w3.org/TR/CSS21/cascade.html#cascading-order
	if origin == "user agent" {
		return 1
	} else if origin == "user" && !importance {
		return 2
	} else if origin == "author" && !importance {
		return 3
	} else if origin == "author" { // && importance
		return 4
	} else {
		if origin != "user" {
			log.Fatalf("origin should be 'user' got %s", origin)
		}
		return 5
	}
}

type weigthedValue struct {
	value CssProperty
	weight uint8
}

type cascadedStyle = map[string]weigthedValue



// Set the value for a property on a given element.
// The value is only set if there is no value of greater weight defined yet.
func addDeclaration(cascadedStyles, propName string, propValues, weight, element,
	pseudoType=None) {


style = cascadedStyles.setdefault((element, pseudoType), {})
Values, previousWeight = style.get(propName, (None, None))
if previousWeight is None || previousWeight <= weight {
style[propName] = propValues, weight
}
