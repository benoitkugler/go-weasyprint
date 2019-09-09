// This module takes care of steps 3 and 4 of “CSS 2.1 processing model”:
// Retrieve stylesheets associated with a document and annotate every element
// with a value for every CSS property.
//
// http://www.w3.org/TR/CSS21/intro.html#processing-model
//
// This module does this in more than two steps. The
// `getAllComputedStyles` function does everything, but it is itsef
// based on other functions in this module.
//
// :copyright: Copyright 2011-2014 Simon Sapin and contributors, see AUTHORS.
// :license: BSD, see LICENSE for details.
package style

import (
	"fmt"
	"log"
	"strings"

	. "github.com/benoitkugler/go-weasyprint/css"
	"github.com/benoitkugler/go-weasyprint/css/validation"
	"github.com/benoitkugler/go-weasyprint/utils"
	"golang.org/x/net/html"
)

var (
	// Reject anything not in here
	pseudoElements = Set{
		"": Has, "before": Has, "after": Has, "first-line": Has, "first-letter": Has,
	}
)

type StyleDict struct {
	Properties
	Anonymous      bool
	inheritedStyle *StyleDict
}

func NewStyleDict() StyleDict {
	return StyleDict{Properties: Properties{}}
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
		is := computedFromCascaded(&html.Node{}, nil, *s, "", StyleDict{}, "")
		is.Anonymous = true
		s.inheritedStyle = &is
	}
	return *s.inheritedStyle
}

func (s StyleDict) ResolveColor(key string) Color {
	value := s.Properties[key].(Color)
	if value.Type == ColorCurrentColor {
		value = s.GetColor()
	}
	return value
}

// Get a dict of computed style mixed from parent and cascaded styles.
func computedFromCascaded(element *html.Node, cascaded cascadedStyle, parentStyle StyleDict, pseudoType string,
	rootStyle StyleDict, baseUrl string) StyleDict {
	if cascaded == nil && !parentStyle.IsZero() {
		// Fast path for anonymous boxes:
		// no cascaded style, only implicitly initial or inherited values.
		computed := InitialValues.Copy()
		for key := range Inherited {
			computed[key] = parentStyle.Properties[key]
		}

		// page is not inherited but taken from the ancestor if "auto"
		computed.SetPage(parentStyle.GetPage())
		// border-*-style is none, so border-width computes to zero.
		// Other than that, properties that would need computing are
		// border-*-color, but they do not apply.
		computed.SetBorderTopWidth(Value{})
		computed.SetBorderBottomWidth(Value{})
		computed.SetBorderLeftWidth(Value{})
		computed.SetBorderRightWidth(Value{})
		computed.SetOutlineWidth(Value{})
		return StyleDict{Properties: computed}
	}

	// Handle inheritance and initial values
	specified, computed := NewStyleDict(), NewStyleDict()
	for name, initial := range InitialValues {
		var (
			keyword String
			value   CssProperty
		)
		if _, in := cascaded[name]; in {
			vp := cascaded[name]
			keyword, _ = vp.value.(String)
			value = vp.value
		} else {
			if Inherited.Has(name) {
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
			if !InitialNotComputed.Has(name) {
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
		val := Page{Valid: true, String: ""}
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
	value  CssProperty
	weight uint8
}

type cascadedStyle = map[string]weigthedValue

type styleKey struct {
	element    *html.Node
	pseudoType string
}

// Set the value for a property on a given element.
// The value is only set if there is no value of greater weight defined yet.
func addDeclaration(cascadedStyles map[styleKey]cascadedStyle, propName string, propValues CssProperty,
	weight uint8, element *html.Node, pseudoType string) {
	key := styleKey{element: element, pseudoType: pseudoType}
	style := cascadedStyles[key]
	if style == nil {
		style = cascadedStyle{}
		cascadedStyles[key] = style
	}
	vw := style[propName]
	if vw.weight <= weight {
		style[propName] = weigthedValue{value: propValues, weight: weight}
	}
}

// Set the computed values of styles to ``element``.
//
// Take the properties left by ``applyStyleRule`` on an element or
// pseudo-element and assign computed values with respect to the cascade,
// declaration priority (ie. ``!important``) and selector specificity.
func setComputedStyles(cascadedStyles map[styleKey]cascadedStyle, computedStyles map[styleKey]StyleDict, element *html.Node, parent,
	root *html.Node, pseudoType, baseUrl string) {

	var parentStyle, rootStyle StyleDict
	if element == root && pseudoType == "" {
		if parent != nil {
			log.Fatal("parent should be nil here")
		}
		parentStyle = StyleDict{}
		rootStyle = StyleDict{Properties: Properties{
			// When specified on the font-size property of the root element, the
			// rem units refer to the property’s initial value.
			"font_size": InitialValues.GetFontSize(),
		}}
	} else {
		if parent == nil {
			log.Fatal("parent shouldn't be nil here")
		}
		parentStyle = computedStyles[styleKey{element: parent, pseudoType: ""}]
		rootStyle = computedStyles[styleKey{element: root, pseudoType: ""}]
	}
	key := styleKey{element: element, pseudoType: pseudoType}
	cascaded := cascadedStyles[key]
	computedStyles[key] = computedFromCascaded(
		element, cascaded, parentStyle, pseudoType, rootStyle, baseUrl)
}

type pageData struct {
	page
	specificity [3]int
}

// Parse a page selector rule.
//     Return a list of page data if the rule is correctly parsed. Page data are a
//     dict containing:
//     - "side" ("left", "right" or ""),
//     - "blank" (true or false),
//     - "first" (true or false),
//     - "name" (page name string or ""), and
//     - "specificity" (list of numbers).
//     Return ``None` if something went wrong while parsing the rule.
func parsePageSelectors(rule QualifiedRule) (out []pageData) {
	// See https://drafts.csswg.org/css-page-3/#syntax-page-selector

	tokens := validation.RemoveWhitespace(rule.Prelude)

	// TODO: Specificity is probably wrong, should clean and test that.
	if len(tokens) == 0 {
		out = append(out, pageData{page: page{
			side: "", blank: false, first: false, name: ""}})
		return out
	}

	for len(tokens) > 0 {
		types := pageData{page: page{
			side: "", blank: false, first: false, name: ""}}

		if ident, ok := tokens[0].(IdentToken); ok {
			tokens = tokens[1:]
			types.name = string(ident.Value)
			types.specificity[0] = 1
		}

		if len(tokens) == 1 {
			return nil
		} else if len(tokens) == 0 {
			out = append(out, types)
			return out
		}

		for len(tokens) > 0 {
			token := tokens[0]
			tokens = tokens[1:]
			literal, ok := token.(LiteralToken)
			if !ok {
				return nil
			}

			if literal.Value == ":" {
				if len(tokens) == 0 {
					return nil
				}
				ident, ok := tokens[0].(IdentToken)
				if !ok {
					return nil
				}
				pseudoClass := ident.Value.Lower()
				switch pseudoClass {
				case "left", "right":
					if types.side != "" {
						return nil
					}
					types.side = pseudoClass
					types.specificity[2] += 1
				case "blank":
					if types.blank {
						return nil
					}
					types.blank = true
					types.specificity[1] += 1

				case "first":
					if types.first {
						return nil
					}
					types.first = true
					types.specificity[1] += 1
				default:
					return nil
				}
			} else if literal.Value == "," {
				if len(tokens) > 0 && types.specificity != [3]int{} {
					break
				} else {
					return nil
				}
			}
		}

		out = append(out, types)
	}

	return out
}

func _isContentNone(rule Token) bool {
	switch token := rule.(type) {
	case QualifiedRule:
		return token.Content == nil
	case AtRule:
		return token.Content == nil
	default:
		return true
	}
}

type selector struct {
	specificity [3]int
	rule        string
	match       func(pageNames []string) []page
}

type pageRule struct {
	rule         AtRule
	selectors    []selector
	declarations []Token
}

// Do the work that can be done early on stylesheet, before they are
// in a document.
// ignoreImports = false
func preprocessStylesheet(deviceMediaType, baseUrl string, stylesheetRules []Token,
	urlFetcher, matcher, pageRules []pageRule, fonts []string, fontConfig int, ignoreImports bool) {

	for _, rule := range stylesheetRules {
		if atRule, isAtRule := rule.(AtRule); _isContentNone(rule) && (!isAtRule || atRule.AtKeyword.Lower() != "import") {
			continue
		}

		switch typedRule := rule.(type) {
		case QualifiedRule:
			declarations := validation.PreprocessDeclarations(baseUrl, ParseDeclarationList(typedRule.Content))
			if len(declarations) > 0 {
				selectors := cssselect2.compileSelectorList(typedRule.Prelude)
				for _, selector := range selectors {
					matcher.addSelector(selector, declarations)
					if !pseudoElements.Has(selector.pseudoElement) {
						err = fmt.Errorf("Unknown pseudo-element: %s", selector.pseudoElement)
						break
					}
				}
				if err != nil {
					log.Printf("Invalid or unsupported selector '%s', %s \n", Serialize(typedRule.Prelude), err)
					continue
				}
				ignoreImports = true
			} else {
				ignoreImports = true
			}
		case AtRule:
			switch typedRule.AtKeyword.Lower() {
			case "import":
				if ignoreImports {
					log.Printf("@import rule '%s' not at the beginning of the whole rule was ignored. \n",
						Serialize(typedRule.Prelude))
					continue
				}

				tokens := validation.RemoveWhitespace(typedRule.Prelude)
				var url string
				if len(tokens) > 0 {
					switch str := tokens[0].(type) {
					case URLToken:
						url = str.Value
					case StringToken:
						url = str.Value
					}
				} else {
					continue
				}
				media := parseMediaQuery(tokens[1:])
				if media == nil {
					log.Printf("Invalid media type '%s' the whole @import rule was ignored. \n",
						Serialize(typedRule.Prelude))
					continue
				}
				if !evaluateMediaQuery(media, deviceMediaType) {
					continue
				}
				url = utils.UrlJoin(baseUrl, url, false, "@import")
				if url != "" {
					_, err := NewCSS()
					// url=url, urlFetcher=urlFetcher,
					// mediaType=deviceMediaType, fontConfig=fontConfig,
					// matcher=matcher, pageRules=pageRules)
					if err != nil {
						log.Printf("Failed to load stylesheet at %s : %s \n", url, err)
					}
				}
			case "media":
				media := parseMediaQuery(typedRule.Prelude)
				if media != nil {
					log.Printf("Invalid media type '%s' the whole @media rule was ignored. \n",
						Serialize(typedRule.Prelude))
					continue
				}
				ignoreImports = true
				if !evaluateMediaQuery(media, deviceMediaType) {
					continue
				}
				contentRules := tinycss2.parseRuleList(rule.content)
				preprocessStylesheet(
					deviceMediaType, baseUrl, contentRules, urlFetcher,
					matcher, pageRules, fonts, fontConfig, true)
			case "page":
				data := parsePageSelectors(rule)
				if data == nil {
					log.Printf("Unsupported @page selector '%s', the whole @page rule was ignored. \n",
						Serialize(rule.prelude))
					continue
				}
				ignoreImports = true
				for _, pageType := range data {
					specificity := pageType.specificity
					pageType.specificity = [3]int{}

					pageType := pageType // capture for closure inside loop
					match := func(pageNames []string) []page {
						return matchingPageTypes(pageType, pageNames)
					}
					content := ParseDeclarationList(rule.content)
					declarations = preprocessDeclarations(baseUrl, content)

					var selectorList []selector
					if len(declarations) > 0 {
						selectorList = []selector{{specificity: specificity, rule: "", match: match}}
						pageRules = append(pageRules, pageRule{rule: typedRule, selectors: selectorList, declarations: declarations})
					}

					for _, marginRule := range content {
						atRule, ok := marginRule.(AtRule)
						if !ok || atRule.Content == nil {
							continue
						}
						declarations = preprocessDeclarations(
							baseUrl,
							ParseDeclarationList(atRule.Content))
						if len(declarations) > 0 {
							selectorList = []selector{{
								specificity: specificity, rule: "@" + atRule.AtKeyword.Lower(),
								match: match}}
							pageRules = append(pageRules, pageRule{rule: marginRule, selectors: selectorList, declarations: declarations})
						}
					}
				}
			case "font-face":
				ignoreImports = true
				content = ParseDeclarationList(rule.content)
				ruleDescriptors = dict(preprocessDescriptors(baseUrl, content))
				ok := true
				for _, key := range [2]string{"src", "fontFamily"} {
					if _, in := ruleDescriptors[key]; !in {
						log.Printf(
							"Missing %s descriptor in '@font-face' rule \n",
							strings.ReplaceAll(key, "_", "-"))
						ok = false
						break
					}
				}
				if ok {
					if fontConfig != nil {
						fontFilename := fontConfig.addFontFace(
							ruleDescriptors, urlFetcher)
						if fontFilename {
							fonts = append(fonts, fontFilename)
						}
					}
				}
			}
		}
	}
}

func parseMediaQuery(tokens []Token) []string {
	tokens = validation.RemoveWhitespace(tokens)
	if len(tokens) == 0 {
		return []string{"all"}
	} else {
		var media []string
		for _, part := range validation.SplitOnComma(tokens) {
			if len(part) == 1 {
				if ident, ok := part[0].(IdentToken); ok {
					media = append(media, ident.Value.Lower())
					continue
				}
			}

			log.Printf("Expected a media type, got %s", Serialize(part))
			return nil
		}
		return media
	}
}
