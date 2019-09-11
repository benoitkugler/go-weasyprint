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
	"strconv"
	"strings"

	"golang.org/x/net/html/atom"

	"github.com/andybalholm/cascadia"

	. "github.com/benoitkugler/go-weasyprint/css"
	"github.com/benoitkugler/go-weasyprint/css/validation"
	"github.com/benoitkugler/go-weasyprint/fonts"
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
		is := computedFromCascaded(&html.Node{}, nil, *s, StyleDict{}, "")
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
func computedFromCascaded(element *html.Node, cascaded cascadedStyle, parentStyle, rootStyle StyleDict, baseUrl string) StyleDict {
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

type weight struct {
	precedence  uint8
	specificity [3]int
}

func less(s1, s2 [3]int) bool {
	for i := range s1 {
		if s1[i] < s2[i] {
			return true
		}
		if s1[i] > s2[i] {
			return false
		}
	}
	return true
}

// Less return `true` if w <= other
func (w weight) Less(other weight) bool {
	return w.precedence < other.precedence || (w.precedence == other.precedence && less(w.specificity, other.specificity))
}

type weigthedValue struct {
	value  CssProperty
	weight weight
}

type cascadedStyle = map[string]weigthedValue

type styleKey struct {
	element    *html.Node
	pseudoType string
}

// Set the value for a property on a given element.
// The value is only set if there is no value of greater weight defined yet.
func addDeclaration(cascadedStyles map[styleKey]cascadedStyle, propName string, propValues CssProperty,
	weight weight, element *html.Node, pseudoType string) {
	key := styleKey{element: element, pseudoType: pseudoType}
	style := cascadedStyles[key]
	if style == nil {
		style = cascadedStyle{}
		cascadedStyles[key] = style
	}
	vw := style[propName]
	if vw.weight.Less(weight) {
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
		element, cascaded, parentStyle, rootStyle, baseUrl)
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
	declarations []validation.ValidatedProperty
}

// Do the work that can be done early on stylesheet, before they are
// in a document.
// ignoreImports = false
func preprocessStylesheet(deviceMediaType, baseUrl string, stylesheetRules []Token,
	urlFetcher, matcher []match, pageRules []pageRule, fonts *[]string, fontConfig *fonts.FontConfiguration, ignoreImports bool) {

	for _, rule := range stylesheetRules {
		if atRule, isAtRule := rule.(AtRule); _isContentNone(rule) && (!isAtRule || atRule.AtKeyword.Lower() != "import") {
			continue
		}

		switch typedRule := rule.(type) {
		case QualifiedRule:
			declarations := validation.PreprocessDeclarations(baseUrl, ParseDeclarationList(typedRule.Content, false, false))
			if len(declarations) > 0 {
				selector, err := cascadia.Compile(Serialize(typedRule.Prelude))
				if err != nil {
					log.Printf("Invalid or unsupported selector '%s', %s \n", Serialize(typedRule.Prelude), err)
					continue
				}
				matcher = append(matcher, match{selector: selector, declarations: declarations})
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
				contentRules := ParseRuleList(typedRule.Content, false, false)
				preprocessStylesheet(
					deviceMediaType, baseUrl, contentRules, urlFetcher,
					matcher, pageRules, fonts, fontConfig, true)
			case "page":
				data := parsePageSelectors(typedRule.QualifiedRule)
				if data == nil {
					log.Printf("Unsupported @page selector '%s', the whole @page rule was ignored. \n",
						Serialize(typedRule.Prelude))
					continue
				}
				ignoreImports = true
				for _, pageType := range data {
					specificity := pageType.specificity
					pageType.specificity = [3]int{}

					pageType := pageType // capture for closure inside loop
					match := func(pageNames []string) []page {
						return matchingPageTypes(pageType.page, pageNames)
					}
					content := ParseDeclarationList(typedRule.Content, false, false)
					declarations := validation.PreprocessDeclarations(baseUrl, content)

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
						declarations = validation.PreprocessDeclarations(
							baseUrl,
							ParseDeclarationList(atRule.Content, false, false))
						if len(declarations) > 0 {
							selectorList = []selector{{
								specificity: specificity, rule: "@" + atRule.AtKeyword.Lower(),
								match: match}}
							pageRules = append(pageRules, pageRule{rule: atRule, selectors: selectorList, declarations: declarations})
						}
					}
				}
			case "font-face":
				ignoreImports = true
				content := ParseDeclarationList(typedRule.Content, false, false)
				ruleDescriptors := map[string]validation.Descriptor{}
				for _, desc := range validation.PreprocessDescriptors(baseUrl, content) {
					ruleDescriptors[desc.Name] = desc.Descriptor
				}
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
						fontFilename := fontConfig.AddFontFace(
							ruleDescriptors, urlFetcher)
						if fontFilename != "" {
							*fonts = append(*fonts, fontFilename)
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

type sheet struct {
	sheet       CSS
	origin      string
	specificity []int
}

type sa struct {
	element     *html.Node
	declaration []Token
	baseUrl     string
}

type sas struct {
	sa
	specificity [3]uint8
}

// Yield ``specificity, (element, declaration, baseUrl)`` rules.
//     Rules from "style" attribute are returned with specificity
//     ``(1, 0, 0)``.
//     If ``presentationalHints`` is ``true``, rules from presentational hints
//     are returned with specificity ``(0, 0, 0)``.
// presentationalHints=false
func findStyleAttributes(tree *utils.HTMLNode, presentationalHints bool, baseUrl string) (out []sas) {

	checkStyleAttribute := func(element *utils.HTMLNode, styleAttribute []Token) sa {
		declarations := ParseDeclarationList(styleAttribute, false, false)
		return styleAttribute{element: element, declaration: declarations, baseUrl: baseUrl}
	}

	iter := utils.NewHtmlIterator(tree)
	for iter.HasNext() {
		element := iter.Next()
		specificity := [3]uint8{1, 0, 0}
		styleAttribute := element.Get("style")
		if styleAttribute != "" {
			out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, styleAttribute)})
		}
		if !presentationalHints {
			continue
		}
		specificity = [3]uint8{0, 0, 0}
		switch element.DataAtom {
		case atom.Body:
			// TODO: we should check the container frame element
			for _, pp := range [4][2]string{{"height", "top"}, {"height", "bottom"}, {"width", "left"}, {"width", "right"}} {
				part, position := pp[0], pp[1]
				styleAttribute = ""
				for _, prop := range [2]string{"margin" + part, position + "margin"} {
					s := element.Get(prop)
					if s != "" {
						styleAttribute = fmt.Sprintf("margin-%s:%spx", position, s)
						break
					}
				}
				if styleAttribute != "" {
					out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, styleAttribute)})
				}
			}
			if element.Get("background") != "" {
				styleAttribute = fmt.Sprintf("background-image:url(%s)", element.Get("background"))
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("bgcolor") != "" {
				styleAttribute = fmt.Sprintf("background-color:%s", element.Get("bgcolor"))
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("text") != "" {
				styleAttribute = fmt.Sprintf("color:%s", element.Get("text"))
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, styleAttribute)})
			}
		// TODO: we should support link, vlink, alink
		case atom.Center:
			out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, "text-align:center")})
		case atom.Div:
			align := strings.ToLower(element.Get("align"))
			switch align {
			case "middle":
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, "text-align:center")})
			case "center", "left", "right", "justify":
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, fmt.Sprintf("text-align:%s", align))})
			}
		case atom.Font:
			if element.Get("color") != "" {
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, fmt.Sprintf("color:%s", element.Get("color")))})
			}
			if element.Get("face") != "" {
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, fmt.Sprintf("font-family:%s", element.Get("face")))})
			}
			if element.Get("size") != "" {
				size := strings.TrimSpace(element.Get("size"))
				relativePlus := strings.HasPrefix(size, "+")
				relativeMinus := strings.HasPrefix(size, "-")
				if relativePlus || relativeMinus {
					size = strings.TrimSpace(string([]rune(size)[1:]))
				}
				sizeI, err := strconv.Atoi(size)
				if err != nil {
					log.Printf("Invalid value for size: %s \n", size)
				} else {
					fontSizes := map[int]string{
						1: "x-small",
						2: "small",
						3: "medium",
						4: "large",
						5: "x-large",
						6: "xx-large",
						7: "48px", // 1.5 * xx-large
					}
					if relativePlus {
						sizeI += 3
					} else if relativeMinus {
						sizeI -= 3
					}
					sizeI = utils.MaxInt(1, utils.MintInt(7, sizeI))
					out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, fmt.Sprintf("font-size:%s", fontSizes[size]))})
				}
			}
		case atom.Table:
			// TODO: we should support cellpadding
			if element.Get != ""("cellspacing") {
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, fmt.Sprintf("border-spacing:%spx", element.Get("cellspacing")))})
			}
			if element.Get("cellpadding") != "" {
				cellpadding := element.Get("cellpadding")
				if utils.IsDigit(cellpadding) {
					cellpadding += "px"
				}
				// TODO: don't match subtables cells
				iterElement = utils.NewHtmlIterator(element)
				for iterElement.HasNext() {
					subelement := iterElement.Next()
					if subelement.DataAtom == atom.Td || subelement.DataAtom == atom.Th {
						out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(subelement,
							fmt.Sprintf("padding-left:%s;padding-right:%s;padding-top:%s;padding-bottom:%s;", cellpadding, cellpadding, cellpadding, cellpadding))})
					}
				}
			}
			if element.Get("hspace") != "" {
				hspace := element.Get("hspace")
				if utils.IsDigit(hspace) {
					hspace += "px"
				}
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element,
					fmt.Sprintf("margin-left:%s;margin-right:%s", hspace, hspace))})
			}
			if element.Get("vspace") != "" {
				vspace := element.Get("vspace")
				if utils.IsDigit(vspace) {
					vspace += "px"
				}
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element,
					fmt.Sprintf("margin-top:%s;margin-bottom:%s", vspace, vspace))})
			}
			if element.Get("width") != "" {
				styleAttribute = fmt.Sprintf("width:%s", element.Get("width"))
				if utils.IsDigit(element.Get("width")) {
					styleAttribute += "px"
				}
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("height") {
				styleAttribute = fmt.Sprintf("height:%s", element.Get("height"))
				if utils.IsDigit(element.Get("height")) {
					styleAttribute += "px"
				}
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("background") != "" {
				styleAttribute = fmt.Sprintf("background-image:url(%s)", element.Get("background"))
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("bgcolor") != "" {
				styleAttribute = fmt.Sprintf("background-color:%s", element.Get("bgcolor"))
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("bordercolor") != "" {
				styleAttribute = fmt.Sprintf("border-color:%s", element.Get("bordercolor"))
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("border") != "" {
				styleAttribute = fmt.Sprintf("border-width:%spx", element.Get("border"))
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, styleAttribute)})
			}
		case atom.Tr, atom.Td, atom.Th, atom.Thead, atom.Tbody, atom.Tfoot:
			align := strings.ToLower(element.Get("align"))
			if align == "left" || align == "right" || align == "justify" {
				// TODO: we should align descendants too
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, fmt.Sprintf("text-align:%s", align))})
			}
			if element.Get("background") {
				styleAttribute = fmt.Sprintf("background-image:url(%s)", element.Get("background"))
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("bgcolor") {
				styleAttribute = fmt.Sprintf("background-color:%s", element.Get("bgcolor"))
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, styleAttribute)})
			}
			if element.DataAtom == atom.Tr || element.DataAtom == atom.Td || element.DataAtom == atom.Th {
				if element.Get("height") != "" {
					styleAttribute = fmt.Sprintf("height:%s", element.Get("height"))
					if utils.IsDigit(element.Get("height")) {
						styleAttribute += "px"
					}
					out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, styleAttribute)})
				}
				if element.DataAtom == atom.Td || element.DataAtom == atom.Th {
					if element.Get("width") != "" {
						styleAttribute = fmt.Sprintf("width:%s", element.Get("width"))
						if utils.IsDigit(element.Get("width")) {
							styleAttribute += "px"
						}
						out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, styleAttribute)})
					}
				}
			}
		case atom.Caption:
			align := strings.ToLower(element.Get("align"))
			// TODO: we should align descendants too
			if align == "left" || align == "right" || align == "justify" {
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, fmt.Sprintf("text-align:%s", align))})
			}
		case atom.Col:
			if element.Get("width") != "" {
				styleAttribute = fmt.Sprintf("width:%s", element.Get("width"))
				if utils.IsDigit(element.Get("width")) {
					styleAttribute += "px"
				}
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, styleAttribute)})
			}
		case atom.Hr:
			size = 0
			if element.Get("size") != "" {
				size, err := strconv.Atoi(element.Get("size"))
				if err != nil {
					log.Printf("Invalid value for size: %s \n", element.Get("size"))
				}
			}
			if element.Get("color") != "" || element.Get("noshade") != "" {
				if size >= 1 {
					out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, fmt.Sprintf("border-width:%spx", (size/2)))})
				}
			} else if size == 1 {
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, "border-bottom-width:0")})
			} else if size > 1 {
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, fmt.Sprintf("height:%spx", (size-2)))})
			}

			if element.Get("width") != "" {
				styleAttribute = fmt.Sprintf("width:%s", element.Get("width"))
				if utils.IsDigit(element.Get("width")) {
					styleAttribute += "px"
				}
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("color") != "" {
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, fmt.Sprintf("color:%s"%element.Get("color")))})
			}
		case atom.Iframe, atom.Applet, atom.Embed, atom.Img, atom.Input, atom.Object:
			if element.DataAtom != atom.Input || strings.ToLower(element.Get("type")) == "image" {
				align = strings.ToLower(element.Get("align"))
				if align == "middle" || align == "center" {
					// TODO: middle && center values are wrong
					out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, "vertical-align:middle")})
				}
				if element.Get("hspace") != "" {
					hspace = element.Get("hspace")
					if utils.IsDigit(hspace) {
						hspace += "px"
					}
					out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element,
						fmt.Sprintf("margin-left:%s;margin-right:%s", hspace, hspace))})
				}
				if element.Get("vspace") {
					vspace = element.Get("vspace")
					if utils.IsDigit(vspace) {
						vspace += "px"
					}
					out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element,
						fmt.Sprintf("margin-top:%s;margin-bottom:%s", vspace, vspace))})
				}
				// TODO: img seems to be excluded for width && height, but a
				// lot of W3C tests rely on this attribute being applied to img
				if element.Get("width") != "" {
					styleAttribute = fmt.Sprintf("width:%s", element.Get("width"))
					if utils.IsDigit(element.Get("width")) {
						styleAttribute += "px"
					}
					out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, styleAttribute)})
				}
				if element.Get("height") != "" {
					styleAttribute = fmt.Sprintf("height:%s", element.Get("height"))
					if utils.IsDigit(element.Get("height")) {
						styleAttribute += "px"
					}
					out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element, styleAttribute)})
				}
				if element.DataAtom == atom.Img || element.DataAtom == atom.Object || element.DataAtom == atom.Input {
					if element.Get("border") != "" {
						out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element,
							fmt.Sprintf("border-width:%spx;border-style:solid", element.Get("border")))})
					}
				}
			}
		case atom.Ol:
			// From https://www.w3.org/TR/css-lists-3/
			if element.Get("start") != "" {
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element,
					fmt.Sprintf("counter-reset:list-item %s;counter-increment:list-item -1", element.Get("start")))})
			}
		case atom.Ul:
			// From https://www.w3.org/TR/css-lists-3/
			if element.Get("value") {
				out = append(out, styleAttributeSpecificity{specificity: specificity, styleAttribute: checkStyleAttribute(element,
					fmt.Sprintf("counter-reset:list-item %s;counter-increment:none", element.Get("value")))})
			}
		}
	}
	return out
}

// Compute all the computed styles of all elements in ``html`` document.
// Do everything from finding author stylesheets to parsing and applying them.
// Return a ``styleFor`` function that takes an element and an optional
// pseudo-element type, and return a StyleDict object.
// presentationalHints=false
func getAllComputedStyles(html, userStylesheets []CSS,
	presentationalHints bool, fontConfig *fonts.FontConfiguration,
	pageRules int) {

	// List stylesheets. Order here is not important ("origin" is).
	sheets := []sheet{
		{sheet: HTML5_UA_STYLESHEET, origin: "", specificity: nil},
	}

	if presentationalHints {
		sheets = append(sheets, sheet{sheet: sheet, origin: "author", specificity: []int{0, 0, 0}})
	}
	for _, sheet := range findStylesheets(
		html.wrapperElement, html.mediaType, html.urlFetcher,
		html.baseUrl, fontConfig, pageRules) {
		sheets = append(sheets, sheet{sheet: sheet, origin: "author", specificity: nil})
	}
	for _, sheet := range userStylesheets {
		sheets = append(sheets, sheet{sheet: sheet, origin: "user", specificity: nil})
	}

	// keys: (element, pseudoElementType)
	//    element: an ElementTree Element or the "@page" string for @page styles
	//    pseudoElementType: a string such as "first" (for @page) or "after",
	//        or None for normal elements
	// values: dicts of
	//     keys: property name as a string
	//     values: (values, weight)
	//         values: a PropertyValue-like object
	//         weight: values with a greater weight take precedence, see
	//             http://www.w3.org/TR/CSS21/cascade.html#cascading-order
	cascadedStyles := map[styleKey]cascadedStyle{}

	log.Println("Step 3 - Applying CSS")
	for _, styleAttr := range findStyleAttributes(html.etreeElement, presentationalHints, html.baseUrl) {
		element, declarations, baseUrl = attributes
		for _, vp := range PreprocessDeclarations(styleAttr, declarations) {
			precedence := declarationPrecedence("author", vp.importance)
			we := weight{precedence: precedence, specificity: styleAttr.specificity}
			addDeclaration(cascadedStyles, name, values, we, element, "")
		}
	}
	// keys: (element, pseudoElementType), like cascadedStyles
	// values: StyleDict objects:
	//     keys: property name as a string
	//     values: a PropertyValue-like object
	computedStyles = map[styleKey]Properties{}

	// First, add declarations and set computed styles for "real" elements *in
	// tree order*. Tree order is important so that parents have computed
	// styles before their children, for inheritance.

	// Iterate on all elements, even if there is no cascaded style for them.
	for element := range html.wrapperElement.iterSubtree() {
		for _, sh := range sheets {
			// sheet, origin, sheetSpecificity
			// Add declarations for matched elements
			for _, selector := range sh.sheet.matcher.Match(element) {
				// specificity, order, pseudoType, declarations = selector
				specificity := sheetSpecificity || selector.specificity
				for _, decl := range selector.payload {
					precedence = declarationPrecedence(sh.origin, decl.Important)
					we = weight{precedence: precedence, specificity: specificity}
					addDeclaration(
						cascadedStyles, name, values, weight, element.etreeElement, selector.pseudoType)
				}
			}
		}
		setComputedStyles(cascadedStyles, computedStyles, element.etreeElement,
			element.Parent, html.etreeElement, "", html.baseUrl)
	}

	pageNames := map[page]struct{}{}

	for _, style := range computedStyles {
		pageNames[style.GetPage()] = Has
	}

	for _, sh := range sheets {
		// Add declarations for page elements
		for _, pr := range sh.sheet.pageRules {
			// Rule, selectorList, declarations
			for _, selector := range pr.selectorList {
				// specificity, pseudoType, match = selector
				specificity = sheetSpecificity || selector.specificity
				for _, pageType := range selector.match(pageNames) {
					for _, decl := range declarations {
						// name, values, importance
						precedence = declarationPrecedence(sh.origin, decl.Important)
						we = weight{precedence: precedence, specificity: specificity}
						addDeclaration(
							cascadedStyles, name, values, weight, pageType,
							pseudoType)
					}
				}
			}
		}
	}

	// Then computed styles for pseudo elements, in any order.
	// Pseudo-elements inherit from their associated element so they come
	// last. Do them in a second pass as there is no easy way to iterate
	// on the pseudo-elements for a given element with the current structure
	// of cascadedStyles. (Keys are (element, pseudoType) tuples.)

	// Only iterate on pseudo-elements that have cascaded styles. (Others
	// might as well not exist.)
	for key := range cascadedStyles {
		// element, pseudoType
		if key.pseudoType != "" && !isinstance(element, PageType) {
			setComputedStyles(
				cascadedStyles, computedStyles, element, parent, html.etreeElement,
				key.pseudoType, html.baseUrl)
			// The pseudo-element inherits from the element.
		}
	}

	__get := func(key styleKey) Properties {
		return computedStyles[key]
	}
	// This is mostly useful to make pseudoType optional.
	// Convenience function to get the computed styles for an element.
	styleFor := func(element, pseudoType string, get func(styleKey) Properties) {
		if get == nil {
			get = __get
		}
		style := get(styleKey{element: element, pseudoType: pseudoType})

		if style != nil {
			display := style.GetDisplay()
			if strings.Contains(display, "table") {
				if (display == "table" || display == "inline-table") && style.GetBorderCollapse() == "collapse" {

					// Padding do not apply
					style.SetPaddingTop(ZeroPixels.ToValue())
					style.SetPaddingBottom(ZeroPixels.ToValue())
					style.SetPaddingLeft(ZeroPixels.ToValue())
					style.SetPaddingRight(ZeroPixels.ToValue())
				}
				if strings.HasPrefix(display, "table-") && display != "table-caption" {

					// Margins do not apply
					style.SetMarginTop(ZeroPixels.ToValue())
					style.SetMarginBottom(ZeroPixels.ToValue())
					style.SetMarginLeft(ZeroPixels.ToValue())
					style.SetMarginRight(ZeroPixels.ToValue())
				}
			}

		}

		return style
	}

	return styleFor, cascadedStyles, computedStyles
}
