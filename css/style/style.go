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

	cascadia "github.com/benoitkugler/cascadia2"

	. "github.com/benoitkugler/go-weasyprint/css"
	"github.com/benoitkugler/go-weasyprint/css/parser"
	"github.com/benoitkugler/go-weasyprint/css/validation"
	"github.com/benoitkugler/go-weasyprint/fonts"
	"github.com/benoitkugler/go-weasyprint/utils"
	"golang.org/x/net/html"
)

var (
	// Reject anything not in here
	pseudoElements = Set{"before": Has, "after": Has, "first-line": Has, "first-letter": Has}
)

type Token = parser.Token

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
		is := computedFromCascaded(&utils.HTMLNode{}, nil, *s, StyleDict{}, "")
		is.Anonymous = true
		s.inheritedStyle = &is
	}
	return *s.inheritedStyle
}

func (s StyleDict) ResolveColor(key string) Color {
	value := s.Properties[key].(Color)
	if value.Type == parser.ColorCurrentColor {
		value = s.GetColor()
	}
	return value
}

// Get a dict of computed style mixed from parent and cascaded styles.
func computedFromCascaded(element *utils.HTMLNode, cascaded cascadedStyle, parentStyle, rootStyle StyleDict, baseUrl string) StyleDict {
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

type element interface {
	ToKey(pseudoType string) utils.ElementKey
}

func matchingPageTypes(pageType utils.PageElement, _names map[Page]struct{}) (out []utils.PageElement) {
	sides := []string{"left", "right", ""}
	if pageType.Side != "" {
		sides = []string{pageType.Side}
	}

	blanks := []bool{true}
	if pageType.Blank == false {
		blanks = []bool{true, false}
	}
	firsts := []bool{true}
	if pageType.First == false {
		firsts = []bool{true, false}
	}
	names := []string{pageType.Name}
	if pageType.Name == "" {
		names = []string{""}
		for page := range _names {
			names = append(names, page.String)
		}
	}
	for _, side := range sides {
		for _, blank := range blanks {
			for _, first := range firsts {
				for _, name := range names {
					out = append(out, utils.PageElement{Side: side, Blank: blank, First: first, Name: name})
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
	specificity cascadia.Specificity
}

// Less return `true` if w <= other
func (w weight) Less(other weight) bool {
	return w.precedence < other.precedence || (w.precedence == other.precedence && w.specificity.Less(other.specificity))
}

type weigthedValue struct {
	value  CssProperty
	weight weight
}

type cascadedStyle = map[string]weigthedValue

// Set the value for a property on a given element.
// The value is only set if there is no value of greater weight defined yet.
func addDeclaration(cascadedStyles map[utils.ElementKey]cascadedStyle, propName string, propValues CssProperty,
	weight weight, elt element, pseudoType string) {
	key := elt.ToKey(pseudoType)
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
func setComputedStyles(cascadedStyles map[utils.ElementKey]cascadedStyle, computedStyles map[utils.ElementKey]StyleDict, element, parent,
	root *utils.HTMLNode, pseudoType, baseUrl string) {

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
		parentStyle = computedStyles[utils.ElementKey{Element: parent, PseudoType: ""}]
		rootStyle = computedStyles[utils.ElementKey{Element: root, PseudoType: ""}]
	}
	key := utils.ElementKey{Element: element, PseudoType: pseudoType}
	cascaded := cascadedStyles[key]
	computedStyles[key] = computedFromCascaded(
		element, cascaded, parentStyle, rootStyle, baseUrl)
}

type pageData struct {
	utils.PageElement
	specificity cascadia.Specificity
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
func parsePageSelectors(rule parser.QualifiedRule) (out []pageData) {
	// See https://drafts.csswg.org/css-page-3/#syntax-page-selector

	tokens := validation.RemoveWhitespace(*rule.Prelude)

	// TODO: Specificity is probably wrong, should clean and test that.
	if len(tokens) == 0 {
		out = append(out, pageData{PageElement: utils.PageElement{
			Side: "", Blank: false, First: false, Name: ""}})
		return out
	}

	for len(tokens) > 0 {
		types := pageData{PageElement: utils.PageElement{
			Side: "", Blank: false, First: false, Name: ""}}

		if ident, ok := tokens[0].(parser.IdentToken); ok {
			tokens = tokens[1:]
			types.Name = string(ident.Value)
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
			literal, ok := token.(parser.LiteralToken)
			if !ok {
				return nil
			}

			if literal.Value == ":" {
				if len(tokens) == 0 {
					return nil
				}
				ident, ok := tokens[0].(parser.IdentToken)
				if !ok {
					return nil
				}
				pseudoClass := ident.Value.Lower()
				switch pseudoClass {
				case "left", "right":
					if types.Side != "" {
						return nil
					}
					types.Side = pseudoClass
					types.specificity[2] += 1
				case "blank":
					if types.Blank {
						return nil
					}
					types.Blank = true
					types.specificity[1] += 1

				case "first":
					if types.First {
						return nil
					}
					types.First = true
					types.specificity[1] += 1
				default:
					return nil
				}
			} else if literal.Value == "," {
				if len(tokens) > 0 && (types.specificity != cascadia.Specificity{}) {
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
	case parser.QualifiedRule:
		return token.Content == nil
	case parser.AtRule:
		return token.Content == nil
	default:
		return true
	}
}

type selectorPageRule struct {
	specificity cascadia.Specificity
	pseudoType  string
	match       func(pageNames map[Page]struct{}) []utils.PageElement
}

type pageRule struct {
	rule         parser.AtRule
	selectors    []selectorPageRule
	declarations []validation.ValidatedProperty
}

// Do the work that can be done early on stylesheet, before they are
// in a document.
// ignoreImports = false
func preprocessStylesheet(deviceMediaType, baseUrl string, stylesheetRules []Token,
	urlFetcher utils.UrlFetcher, matcher *matcher, pageRules *[]pageRule, fonts *[]string, fontConfig *fonts.FontConfiguration, ignoreImports bool) {

	for _, rule := range stylesheetRules {
		if atRule, isAtRule := rule.(parser.AtRule); _isContentNone(rule) && (!isAtRule || atRule.AtKeyword.Lower() != "import") {
			continue
		}

		switch typedRule := rule.(type) {
		case parser.QualifiedRule:
			declarations := validation.PreprocessDeclarations(baseUrl, parser.ParseDeclarationList(*typedRule.Content, false, false))
			if len(declarations) > 0 {
				fmt.Println(parser.Serialize(*typedRule.Prelude))
				selector, err := cascadia.Compile(parser.Serialize(*typedRule.Prelude))
				if err != nil {
					log.Printf("Invalid or unsupported selector '%s', %s \n", parser.Serialize(*typedRule.Prelude), err)
					continue
				}
				for _, sel := range selector {
					if _, in := pseudoElements[sel.PseudoElement()]; !in {
						err = fmt.Errorf("Unsupported pseudo-elment : %s", sel.PseudoElement())
						break
					}
				}
				if err != nil {
					log.Println(err)
					continue
				}
				*matcher = append(*matcher, match{selector: selector, declarations: declarations})
				ignoreImports = true
			} else {
				ignoreImports = true
			}
		case parser.AtRule:
			switch typedRule.AtKeyword.Lower() {
			case "import":
				if ignoreImports {
					log.Printf("@import rule '%s' not at the beginning of the whole rule was ignored. \n",
						parser.Serialize(*typedRule.Prelude))
					continue
				}

				tokens := validation.RemoveWhitespace(*typedRule.Prelude)
				var url string
				if len(tokens) > 0 {
					switch str := tokens[0].(type) {
					case parser.URLToken:
						url = str.Value
					case parser.StringToken:
						url = str.Value
					}
				} else {
					continue
				}
				media := parseMediaQuery(tokens[1:])
				if media == nil {
					log.Printf("Invalid media type '%s' the whole @import rule was ignored. \n",
						parser.Serialize(*typedRule.Prelude))
					continue
				}
				if !evaluateMediaQuery(media, deviceMediaType) {
					continue
				}
				url = utils.UrlJoin(baseUrl, url, false, "@import")
				if url != "" {
					_, err := NewCSS(CssUrl(url), "", urlFetcher, false,
						deviceMediaType, fontConfig, matcher, pageRules)
					if err != nil {
						log.Printf("Failed to load stylesheet at %s : %s \n", url, err)
					}
				}
			case "media":
				media := parseMediaQuery(*typedRule.Prelude)
				if media != nil {
					log.Printf("Invalid media type '%s' the whole @media rule was ignored. \n",
						parser.Serialize(*typedRule.Prelude))
					continue
				}
				ignoreImports = true
				if !evaluateMediaQuery(media, deviceMediaType) {
					continue
				}
				contentRules := parser.ParseRuleList(*typedRule.Content, false, false)
				preprocessStylesheet(
					deviceMediaType, baseUrl, contentRules, urlFetcher,
					matcher, pageRules, fonts, fontConfig, true)
			case "page":
				data := parsePageSelectors(typedRule.QualifiedRule)
				if data == nil {
					log.Printf("Unsupported @page selector '%s', the whole @page rule was ignored. \n",
						parser.Serialize(*typedRule.Prelude))
					continue
				}
				ignoreImports = true
				for _, pageType := range data {
					specificity := pageType.specificity
					pageType.specificity = cascadia.Specificity{}

					pageType := pageType // capture for closure inside loop
					match := func(pageNames map[Page]struct{}) []utils.PageElement {
						return matchingPageTypes(pageType.PageElement, pageNames)
					}
					content := parser.ParseDeclarationList(*typedRule.Content, false, false)
					declarations := validation.PreprocessDeclarations(baseUrl, content)

					var selectors []selectorPageRule
					if len(declarations) > 0 {
						selectors = []selectorPageRule{{specificity: specificity, pseudoType: "", match: match}}
						*pageRules = append(*pageRules, pageRule{rule: typedRule, selectors: selectors, declarations: declarations})
					}

					for _, marginRule := range content {
						atRule, ok := marginRule.(parser.AtRule)
						if !ok || atRule.Content == nil {
							continue
						}
						declarations = validation.PreprocessDeclarations(
							baseUrl,
							parser.ParseDeclarationList(*atRule.Content, false, false))
						if len(declarations) > 0 {
							selectors = []selectorPageRule{{
								specificity: specificity, pseudoType: "@" + atRule.AtKeyword.Lower(),
								match: match}}
							*pageRules = append(*pageRules, pageRule{rule: atRule, selectors: selectors, declarations: declarations})
						}
					}
				}
			case "font-face":
				ignoreImports = true
				content := parser.ParseDeclarationList(*typedRule.Content, false, false)
				ruleDescriptors := map[string]validation.Descriptor{}
				for _, desc := range validation.PreprocessDescriptors(baseUrl, content) {
					ruleDescriptors[desc.Name] = desc.Descriptor
				}
				ok := true
				for _, key := range [2]string{"src", "font_family"} {
					if _, in := ruleDescriptors[key]; !in {
						log.Printf(
							`Missing %s descriptor in "@font-face" rule at %d:%d`+"\n",
							strings.ReplaceAll(key, "_", "-"), rule.Position().Line, rule.Position().Column)
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
				if ident, ok := part[0].(parser.IdentToken); ok {
					media = append(media, ident.Value.Lower())
					continue
				}
			}

			log.Printf("Expected a media type, got %s", parser.Serialize(part))
			return nil
		}
		return media
	}
}

type sheet struct {
	sheet       CSS
	origin      string
	specificity []uint8
}

type sa struct {
	element     *utils.HTMLNode
	declaration []Token
	baseUrl     string
}

type sas struct {
	sa
	specificity cascadia.Specificity
}

// Yield ``specificity, (element, declaration, baseUrl)`` rules.
//     Rules from "style" attribute are returned with specificity
//     ``(1, 0, 0)``.
//     If ``presentationalHints`` is ``true``, rules from presentational hints
//     are returned with specificity ``(0, 0, 0)``.
// presentationalHints=false
func findStyleAttributes(tree *utils.HTMLNode, presentationalHints bool, baseUrl string) (out []sas) {

	checkStyleAttribute := func(element *utils.HTMLNode, styleAttribute string) sa {
		declarations := parser.ParseDeclarationList2(styleAttribute, false, false)
		return sa{element: element, declaration: declarations, baseUrl: baseUrl}
	}

	iter := tree.Iter()
	for iter.HasNext() {
		element := iter.Next()
		specificity := cascadia.Specificity{1, 0, 0}
		styleAttribute := element.Get("style")
		if styleAttribute != "" {
			out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, styleAttribute)})
		}
		if !presentationalHints {
			continue
		}
		specificity = cascadia.Specificity{0, 0, 0}
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
					out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, styleAttribute)})
				}
			}
			if element.Get("background") != "" {
				styleAttribute = fmt.Sprintf("background-image:url(%s)", element.Get("background"))
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("bgcolor") != "" {
				styleAttribute = fmt.Sprintf("background-color:%s", element.Get("bgcolor"))
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("text") != "" {
				styleAttribute = fmt.Sprintf("color:%s", element.Get("text"))
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, styleAttribute)})
			}
		// TODO: we should support link, vlink, alink
		case atom.Center:
			out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, "text-align:center")})
		case atom.Div:
			align := strings.ToLower(element.Get("align"))
			switch align {
			case "middle":
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, "text-align:center")})
			case "center", "left", "right", "justify":
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, fmt.Sprintf("text-align:%s", align))})
			}
		case atom.Font:
			if element.Get("color") != "" {
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, fmt.Sprintf("color:%s", element.Get("color")))})
			}
			if element.Get("face") != "" {
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, fmt.Sprintf("font-family:%s", element.Get("face")))})
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
					sizeI = utils.MaxInt(1, utils.MinInt(7, sizeI))
					out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, fmt.Sprintf("font-size:%s", fontSizes[sizeI]))})
				}
			}
		case atom.Table:
			// TODO: we should support cellpadding
			if element.Get("cellspacing") != "" {
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, fmt.Sprintf("border-spacing:%spx", element.Get("cellspacing")))})
			}
			if element.Get("cellpadding") != "" {
				cellpadding := element.Get("cellpadding")
				if utils.IsDigit(cellpadding) {
					cellpadding += "px"
				}
				// TODO: don't match subtables cells
				iterElement := element.Iter()
				for iterElement.HasNext() {
					subelement := iterElement.Next()
					if subelement.DataAtom == atom.Td || subelement.DataAtom == atom.Th {
						out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(subelement,
							fmt.Sprintf("padding-left:%s;padding-right:%s;padding-top:%s;padding-bottom:%s;", cellpadding, cellpadding, cellpadding, cellpadding))})
					}
				}
			}
			if element.Get("hspace") != "" {
				hspace := element.Get("hspace")
				if utils.IsDigit(hspace) {
					hspace += "px"
				}
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element,
					fmt.Sprintf("margin-left:%s;margin-right:%s", hspace, hspace))})
			}
			if element.Get("vspace") != "" {
				vspace := element.Get("vspace")
				if utils.IsDigit(vspace) {
					vspace += "px"
				}
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element,
					fmt.Sprintf("margin-top:%s;margin-bottom:%s", vspace, vspace))})
			}
			if element.Get("width") != "" {
				styleAttribute = fmt.Sprintf("width:%s", element.Get("width"))
				if utils.IsDigit(element.Get("width")) {
					styleAttribute += "px"
				}
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("height") != "" {
				styleAttribute = fmt.Sprintf("height:%s", element.Get("height"))
				if utils.IsDigit(element.Get("height")) {
					styleAttribute += "px"
				}
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("background") != "" {
				styleAttribute = fmt.Sprintf("background-image:url(%s)", element.Get("background"))
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("bgcolor") != "" {
				styleAttribute = fmt.Sprintf("background-color:%s", element.Get("bgcolor"))
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("bordercolor") != "" {
				styleAttribute = fmt.Sprintf("border-color:%s", element.Get("bordercolor"))
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("border") != "" {
				styleAttribute = fmt.Sprintf("border-width:%spx", element.Get("border"))
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, styleAttribute)})
			}
		case atom.Tr, atom.Td, atom.Th, atom.Thead, atom.Tbody, atom.Tfoot:
			align := strings.ToLower(element.Get("align"))
			if align == "left" || align == "right" || align == "justify" {
				// TODO: we should align descendants too
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, fmt.Sprintf("text-align:%s", align))})
			}
			if element.Get("background") != "" {
				styleAttribute = fmt.Sprintf("background-image:url(%s)", element.Get("background"))
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("bgcolor") != "" {
				styleAttribute = fmt.Sprintf("background-color:%s", element.Get("bgcolor"))
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, styleAttribute)})
			}
			if element.DataAtom == atom.Tr || element.DataAtom == atom.Td || element.DataAtom == atom.Th {
				if element.Get("height") != "" {
					styleAttribute = fmt.Sprintf("height:%s", element.Get("height"))
					if utils.IsDigit(element.Get("height")) {
						styleAttribute += "px"
					}
					out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, styleAttribute)})
				}
				if element.DataAtom == atom.Td || element.DataAtom == atom.Th {
					if element.Get("width") != "" {
						styleAttribute = fmt.Sprintf("width:%s", element.Get("width"))
						if utils.IsDigit(element.Get("width")) {
							styleAttribute += "px"
						}
						out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, styleAttribute)})
					}
				}
			}
		case atom.Caption:
			align := strings.ToLower(element.Get("align"))
			// TODO: we should align descendants too
			if align == "left" || align == "right" || align == "justify" {
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, fmt.Sprintf("text-align:%s", align))})
			}
		case atom.Col:
			if element.Get("width") != "" {
				styleAttribute = fmt.Sprintf("width:%s", element.Get("width"))
				if utils.IsDigit(element.Get("width")) {
					styleAttribute += "px"
				}
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, styleAttribute)})
			}
		case atom.Hr:
			size := 0
			if element.Get("size") != "" {
				var err error
				size, err = strconv.Atoi(element.Get("size"))
				if err != nil {
					log.Printf("Invalid value for size: %s \n", element.Get("size"))
				}
			}
			if element.Get("color") != "" || element.Get("noshade") != "" {
				if size >= 1 {
					out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, fmt.Sprintf("border-width:%dpx", size/2))})
				}
			} else if size == 1 {
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, "border-bottom-width:0")})
			} else if size > 1 {
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, fmt.Sprintf("height:%dpx", size-2))})
			}

			if element.Get("width") != "" {
				styleAttribute = fmt.Sprintf("width:%s", element.Get("width"))
				if utils.IsDigit(element.Get("width")) {
					styleAttribute += "px"
				}
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("color") != "" {
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, fmt.Sprintf("color:%s", element.Get("color")))})
			}
		case atom.Iframe, atom.Applet, atom.Embed, atom.Img, atom.Input, atom.Object:
			if element.DataAtom != atom.Input || strings.ToLower(element.Get("type")) == "image" {
				align := strings.ToLower(element.Get("align"))
				if align == "middle" || align == "center" {
					// TODO: middle && center values are wrong
					out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, "vertical-align:middle")})
				}
				if element.Get("hspace") != "" {
					hspace := element.Get("hspace")
					if utils.IsDigit(hspace) {
						hspace += "px"
					}
					out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element,
						fmt.Sprintf("margin-left:%s;margin-right:%s", hspace, hspace))})
				}
				if element.Get("vspace") != "" {
					vspace := element.Get("vspace")
					if utils.IsDigit(vspace) {
						vspace += "px"
					}
					out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element,
						fmt.Sprintf("margin-top:%s;margin-bottom:%s", vspace, vspace))})
				}
				// TODO: img seems to be excluded for width && height, but a
				// lot of W3C tests rely on this attribute being applied to img
				if element.Get("width") != "" {
					styleAttribute = fmt.Sprintf("width:%s", element.Get("width"))
					if utils.IsDigit(element.Get("width")) {
						styleAttribute += "px"
					}
					out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, styleAttribute)})
				}
				if element.Get("height") != "" {
					styleAttribute = fmt.Sprintf("height:%s", element.Get("height"))
					if utils.IsDigit(element.Get("height")) {
						styleAttribute += "px"
					}
					out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element, styleAttribute)})
				}
				if element.DataAtom == atom.Img || element.DataAtom == atom.Object || element.DataAtom == atom.Input {
					if element.Get("border") != "" {
						out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element,
							fmt.Sprintf("border-width:%spx;border-style:solid", element.Get("border")))})
					}
				}
			}
		case atom.Ol:
			// From https://www.w3.org/TR/css-lists-3/
			if element.Get("start") != "" {
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element,
					fmt.Sprintf("counter-reset:list-item %s;counter-increment:list-item -1", element.Get("start")))})
			}
		case atom.Ul:
			// From https://www.w3.org/TR/css-lists-3/
			if element.Get("value") != "" {
				out = append(out, sas{specificity: specificity, sa: checkStyleAttribute(element,
					fmt.Sprintf("counter-reset:list-item %s;counter-increment:none", element.Get("value")))})
			}
		}
	}
	return out
}

// Yield the stylesheets in ``elementTree``.
// The output order is the same as the source order.
func findStylesheets(wrapperElement *utils.HTMLNode, deviceMediaType string, urlFetcher utils.UrlFetcher, baseUrl string,
	fontConfig *fonts.FontConfiguration, pageRules *[]pageRule) (out []CSS, err error) {
	sel := cascadia.MustCompile("style, link")
	for _, _element := range sel.MatchAll((*html.Node)(wrapperElement)) {
		element := (*utils.HTMLNode)(_element)
		mimeType := element.Get("type")
		if mimeType == "" {
			mimeType = "text/css"
		}
		mimeType = strings.TrimSpace(strings.SplitN(mimeType, ";", 1)[0])
		// Only keep "type/subtype" from "type/subtype ; param1; param2".
		if mimeType != "text/css" {
			continue
		}
		mediaAttr := strings.TrimSpace(element.Get("media"))
		if mediaAttr == "" {
			mediaAttr = "all"
		}
		media := strings.Split(mediaAttr, ",")
		for i, s := range media {
			media[i] = strings.TrimSpace(s)
		}
		if !evaluateMediaQuery(media, deviceMediaType) {
			continue
		}
		switch element.DataAtom {
		case atom.Style:
			// Content is text that is directly in the <style> element, not its
			// descendants
			content := element.GetChildText()
			// ElementTree should give us either unicode or  ASCII-only
			// bytestrings, so we don"t need `encoding` here.
			css, err := NewCSS(CssString(content), baseUrl, urlFetcher, false, deviceMediaType,
				fontConfig, nil, pageRules)
			if err != nil {
				return nil, err
			}
			out = append(out, css)
		case atom.Link:
			if element.Get("href") != "" {
				if !element.HasLinkType("stylesheet") || element.HasLinkType("alternate") {
					continue
				}
				href := element.GetUrlAttribute("href", baseUrl, false)
				if href != "" {
					css, err := NewCSS(CssUrl(href), "", urlFetcher, true, deviceMediaType,
						fontConfig, nil, pageRules)
					if err != nil {
						log.Printf("Failed to load stylesheet at %s : %s \n", href, err)
					} else {
						out = append(out, css)
					}
				}
			}
		}
	}
	return out, nil
}

type htmlEntity struct {
	root       *utils.HTMLNode
	mediaType  string
	urlFetcher utils.UrlFetcher
	baseUrl    string
}

type styleGetter = func(element element, pseudoType string, get func(utils.ElementKey) StyleDict) StyleDict

// Compute all the computed styles of all elements in ``html`` document.
// Do everything from finding author stylesheets to parsing and applying them.
// Return a ``styleFor`` function that takes an element and an optional
// pseudo-element type, and return a StyleDict object.
// presentationalHints=false
func GetAllComputedStyles(html htmlEntity, userStylesheets []CSS,
	presentationalHints bool, fontConfig *fonts.FontConfiguration,
	pageRules *[]pageRule) (styleGetter, map[utils.ElementKey]cascadedStyle, map[utils.ElementKey]StyleDict, error) {

	// List stylesheets. Order here is not important ("origin" is).
	sheets := []sheet{
		{sheet: HTML5_UA_STYLESHEET, origin: "", specificity: nil},
	}

	if presentationalHints {
		sheets = append(sheets, sheet{sheet: HTML5_PH_STYLESHEET, origin: "author", specificity: []uint8{0, 0, 0}})
	}
	authorShts, err := findStylesheets(html.root, html.mediaType, html.urlFetcher,
		html.baseUrl, fontConfig, pageRules)
	if err != nil {
		return nil, nil, nil, err
	}
	for _, sht := range authorShts {
		sheets = append(sheets, sheet{sheet: sht, origin: "author", specificity: nil})
	}
	for _, sht := range userStylesheets {
		sheets = append(sheets, sheet{sheet: sht, origin: "user", specificity: nil})
	}

	// keys: (element, pseudoElementType)
	//    Element: an ElementTree Element or the "@page" string for @page styles
	//    pseudoElementType: a string such as "first" (for @page) or "after",
	//        or None for normal elements
	// values: dicts of
	//     keys: property name as a string
	//     values: (values, weight)
	//         values: a PropertyValue-like object
	//         weight: values with a greater weight take precedence, see
	//             http://www.w3.org/TR/CSS21/cascade.html#cascading-order
	cascadedStyles := map[utils.ElementKey]cascadedStyle{}

	log.Println("Step 3 - Applying CSS")
	for _, styleAttr := range findStyleAttributes(html.root, presentationalHints, html.baseUrl) {
		// element, declarations, baseUrl = attributes
		for _, vp := range validation.PreprocessDeclarations(styleAttr.baseUrl, styleAttr.declaration) {
			// name, values, importance = vp
			precedence := declarationPrecedence("author", vp.Important)
			we := weight{precedence: precedence, specificity: styleAttr.specificity}
			addDeclaration(cascadedStyles, vp.Name, vp.Value, we, styleAttr.element, "")
		}
	}
	// keys: (element, pseudoElementType), like cascadedStyles
	// values: StyleDict objects:
	//     keys: property name as a string
	//     values: a PropertyValue-like object
	computedStyles := map[utils.ElementKey]StyleDict{}

	// First, add declarations and set computed styles for "real" elements *in
	// tree order*. Tree order is important so that parents have computed
	// styles before their children, for inheritance.

	// Iterate on all elements, even if there is no cascaded style for them.
	iter := html.root.Iter()
	for iter.HasNext() {
		element := iter.Next()
		for _, sh := range sheets {
			// sheet, origin, sheetSpecificity
			// Add declarations for matched elements
			for _, selector := range sh.sheet.matcher.Match(element.AsHtmlNode()) {
				// specificity, order, pseudoType, declarations = selector
				specificity := selector.specificity
				if len(specificity) == 3 {
					specificity = cascadia.Specificity{sh.specificity[0], sh.specificity[1], sh.specificity[2]}
				}
				for _, decl := range selector.payload {
					// name, values, importance = decl
					precedence := declarationPrecedence(sh.origin, decl.Important)
					we := weight{precedence: precedence, specificity: specificity}
					addDeclaration(cascadedStyles, decl.Name, decl.Value, we, element, selector.pseudoType)
				}
			}
		}
		setComputedStyles(cascadedStyles, computedStyles, element,
			(*utils.HTMLNode)(element.Parent), html.root, "", html.baseUrl)
	}

	pageNames := map[Page]struct{}{}

	for _, style := range computedStyles {
		pageNames[style.GetPage()] = Has
	}

	for _, sh := range sheets {
		// Add declarations for page elements
		for _, pr := range *sh.sheet.pageRules {
			// Rule, selectorList, declarations
			for _, selector := range pr.selectors {
				// specificity, pseudoType, match = selector
				specificity := selector.specificity
				if len(specificity) == 3 {
					specificity = cascadia.Specificity{sh.specificity[0], sh.specificity[1], sh.specificity[2]}
				}
				for _, pageType := range selector.match(pageNames) {
					for _, decl := range pr.declarations {
						// name, values, importance
						precedence := declarationPrecedence(sh.origin, decl.Important)
						we := weight{precedence: precedence, specificity: specificity}
						addDeclaration(
							cascadedStyles, decl.Name, decl.Value, we, pageType,
							selector.pseudoType)
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
		if key.PseudoType != "" && key.Element != nil {
			setComputedStyles(
				cascadedStyles, computedStyles, key.Element, key.Element, html.root,
				key.PseudoType, html.baseUrl)
			// The pseudo-element inherits from the element.
		}
	}

	__get := func(key utils.ElementKey) StyleDict {
		return computedStyles[key]
	}
	// This is mostly useful to make pseudoType optional.
	// Convenience function to get the computed styles for an element.
	styleFor := func(element element, pseudoType string, get func(utils.ElementKey) StyleDict) StyleDict {
		if get == nil {
			get = __get
		}
		style := get(element.ToKey(pseudoType))

		if !style.IsZero() {
			display := style.GetDisplay().String()
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

	return styleFor, cascadedStyles, computedStyles, nil
}
