// This module takes care of steps 3 and 4 of “CSS 2.1 processing model”:
// Retrieve stylesheets associated with a document and annotate every Element
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
package tree

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/benoitkugler/go-weasyprint/boxes/counters"
	"github.com/benoitkugler/go-weasyprint/layout/text"
	"github.com/benoitkugler/go-weasyprint/logger"

	"golang.org/x/net/html/atom"

	"github.com/benoitkugler/cascadia"

	"github.com/benoitkugler/go-weasyprint/style/parser"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/style/validation"
	"github.com/benoitkugler/go-weasyprint/utils"
	"golang.org/x/net/html"
)

// Reject anything not in here
var pseudoElements = utils.NewSet("", "before", "after", "marker", "first-line", "first-letter")

type Token = parser.Token

// StyleFor provides a convenience function `Get` to get the computed styles for an Element.
type StyleFor struct {
	CascadedStyles map[utils.ElementKey]cascadedStyle
	computedStyles map[utils.ElementKey]pr.ElementStyle
	textContext    text.TextLayoutContext
	sheets         []sheet
}

func newStyleFor(html *HTML, sheets []sheet, presentationalHints bool,
	targetColllector *TargetCollector, textContext text.TextLayoutContext) *StyleFor {
	out := StyleFor{
		CascadedStyles: map[utils.ElementKey]cascadedStyle{},
		computedStyles: map[utils.ElementKey]pr.ElementStyle{},
		sheets:         sheets,
		textContext:    textContext,
	}

	logger.ProgressLogger.Printf("Step 3 - Applying CSS - %d sheet(s)\n", len(sheets))

	for _, styleAttr := range findStyleAttributes(html.Root, presentationalHints, html.BaseUrl) {
		// Element, declarations, BaseUrl = attributes
		style, ok := out.CascadedStyles[styleAttr.element.ToKey("")]
		if !ok {
			style = cascadedStyle{}
			out.CascadedStyles[styleAttr.element.ToKey("")] = style
		}
		for _, decl := range validation.PreprocessDeclarations(styleAttr.baseUrl, styleAttr.declaration) {
			// name, values, importance = decl
			precedence := declarationPrecedence("author", decl.Important)
			we := weight{precedence: precedence, specificity: styleAttr.specificity}
			oldWeight := style[decl.Name].weight
			if oldWeight.isNone() || oldWeight.Less(we) {
				style[decl.Name] = weigthedValue{weight: we, value: decl.Value}
			}
		}
	}

	// First, add declarations and set computed styles for "real" elements *in
	// tree order*. Tree order is important so that parents have computed
	// styles before their children, for inheritance.

	// Iterate on all elements, even if there is no cascaded style for them.
	iter := html.Root.Iter()
	for iter.HasNext() {
		element := iter.Next()
		for _, sh := range sheets {
			// sheet, origin, sheetSpecificity
			// Add declarations for matched elements
			matchedSelectors := sh.sheet.Matcher.Match(element.AsHtmlNode())
			for _, selector := range matchedSelectors {
				// specificity, order, pseudoType, declarations = selector
				specificity := selector.specificity
				if len(sh.specificity) == 3 {
					specificity = cascadia.Specificity{sh.specificity[0], sh.specificity[1], sh.specificity[2]}
				}
				key := element.ToKey(selector.pseudoType)
				style, in := out.CascadedStyles[key]
				if !in {
					style = cascadedStyle{}
					out.CascadedStyles[key] = style
				}

				for _, decl := range selector.payload {
					// name, values, importance = decl
					precedence := declarationPrecedence(sh.origin, decl.Important)
					we := weight{precedence: precedence, specificity: specificity}
					oldWeight := style[decl.Name].weight
					if oldWeight.isNone() || oldWeight.Less(we) {
						style[decl.Name] = weigthedValue{weight: we, value: decl.Value}
					}
				}
			}
		}
		out.SetComputedStyles(element, (*utils.HTMLNode)(element.Parent), html.Root, "", html.BaseUrl,
			targetColllector)
	}

	// Then computed styles for pseudo elements, in any order.
	// Pseudo-elements inherit from their associated Element so they come
	// last. Do them in a second pass as there is no easy way to iterate
	// on the pseudo-elements for a given Element with the current structure
	// of CascadedStyles. (Keys are (Element, pseudoType) tuples.)

	// Only iterate on pseudo-elements that have cascaded styles. (Others
	// might as well not exist.)
	for key := range out.CascadedStyles {
		// Element, pseudoType
		if key.PseudoType != "" && !key.IsPageType() {
			out.SetComputedStyles(key.Element, key.Element, html.Root,
				key.PseudoType, html.BaseUrl, targetColllector)
			// The pseudo-Element inherits from the Element.
		}
	}
	// Clear the cascaded styles, we don't need them anymore. Keep the
	// dictionary, it is used later for page margins.
	for k := range out.CascadedStyles {
		delete(out.CascadedStyles, k)
	}
	return &out
}

// Set the computed values of styles to ``Element``.
//
// Take the properties left by ``applyStyleRule`` on an Element or
// pseudo-Element and assign computed values with respect to the cascade,
// declaration priority (ie. ``!important``) and selector specificity.
func (self *StyleFor) SetComputedStyles(element, parent Element,
	root *utils.HTMLNode, pseudoType, baseUrl string,
	targetCollector *TargetCollector) {

	var (
		parentStyle pr.ElementStyle
		rootStyle   pr.Properties
	)
	if element == root && pseudoType == "" {
		if node, ok := parent.(*utils.HTMLNode); parent != nil && (ok && node != nil) {
			log.Fatal("parent should be nil here")
		}
		rootStyle = pr.Properties{
			// When specified on the font-size property of the Root Element, the
			// rem units refer to the property’s initial value.
			"font_size": pr.InitialValues.GetFontSize(),
		}
	} else {
		if parent == nil {
			log.Fatal("parent shouldn't be nil here")
		}
		parentStyle = self.computedStyles[parent.ToKey("")]
		rootStyle = self.computedStyles[utils.ElementKey{Element: root, PseudoType: ""}].(*ComputedStyle).dict
	}
	key := element.ToKey(pseudoType)
	cascaded, in := self.CascadedStyles[key]
	if !in {
		cascaded = cascadedStyle{}
	}
	self.computedStyles[key] = ComputedFromCascaded(element, cascaded, parentStyle,
		rootStyle, pseudoType, baseUrl, targetCollector, self.textContext)
}

func (s StyleFor) Get(element Element, pseudoType string) pr.ElementStyle {
	style := s.computedStyles[element.ToKey(pseudoType)]
	if style != nil {
		display := style.GetDisplay()
		if display.Has("table") && style.GetBorderCollapse() == "collapse" {
			// Padding do not apply
			style.SetPaddingTop(pr.ZeroPixels.ToValue())
			style.SetPaddingBottom(pr.ZeroPixels.ToValue())
			style.SetPaddingLeft(pr.ZeroPixels.ToValue())
			style.SetPaddingRight(pr.ZeroPixels.ToValue())
		}

		if d := display[0]; display[1] == "" && display[2] == "" && strings.HasPrefix(d, "table-") && d != "table-caption" {
			// Margins do not apply
			style.SetMarginTop(pr.ZeroPixels.ToValue())
			style.SetMarginBottom(pr.ZeroPixels.ToValue())
			style.SetMarginLeft(pr.ZeroPixels.ToValue())
			style.SetMarginRight(pr.ZeroPixels.ToValue())
		}
	}
	return style
}

func (s StyleFor) addPageDeclarations(page_T utils.PageElement) {
	for _, sh := range s.sheets {
		// Add declarations for page elements
		for _, pageR := range sh.sheet.pageRules {
			// Rule, selectorList, declarations
			for _, selector := range pageR.selectors {
				// specificity, pseudoType, selector_page_type = selector
				if pageMatchT(selector.pageType, page_T) {
					specificity := selector.specificity
					if len(sh.specificity) == 3 {
						specificity = cascadia.Specificity{sh.specificity[0], sh.specificity[1], sh.specificity[2]}
					}
					style, in := s.CascadedStyles[page_T.ToKey(selector.pseudoType)]
					if !in {
						style = cascadedStyle{}
						s.CascadedStyles[page_T.ToKey(selector.pseudoType)] = style
					}

					for _, decl := range pageR.declarations {
						// name, values, importance
						precedence := declarationPrecedence(sh.origin, decl.Important)
						we := weight{precedence: precedence, specificity: specificity}
						oldWeight := style[decl.Name].weight
						if oldWeight.isNone() || oldWeight.Less(we) {
							style[decl.Name] = weigthedValue{weight: we, value: decl.Value}
						}
					}
				}
			}
		}
	}
}

// pr.ElementStyle provides on demand access of computed properties
// for a box.

var (
	_ pr.ElementStyle = (*ComputedStyle)(nil)
	_ pr.ElementStyle = (*AnonymousStyle)(nil)
)

type computedRequirements struct {
	float    pr.String
	display  pr.Display
	position pr.BoolString
}

// ComputedStyle provides on demand access of computed properties
type ComputedStyle struct {
	dict      pr.Properties
	variables map[string]pr.ValidatedProperty

	textContext text.TextLayoutContext

	rootStyle   pr.Properties
	parentStyle pr.ElementStyle
	cascaded    cascadedStyle
	element     Element
	pseudoType  string
	baseUrl     string
	specified   computedRequirements
}

func newComputedStyle(parentStyle pr.ElementStyle, cascaded cascadedStyle,
	element Element, pseudoType string, rootStyle pr.Properties, baseUrl string, textContext text.TextLayoutContext) *ComputedStyle {
	out := &ComputedStyle{
		dict:        make(pr.Properties),
		variables:   make(map[string]pr.ValidatedProperty),
		textContext: textContext,
		parentStyle: parentStyle,
		cascaded:    cascaded,
		element:     element,
		pseudoType:  pseudoType,
		rootStyle:   rootStyle,
		baseUrl:     baseUrl,
	}

	// inherit the variables
	if parentStyle != nil {
		for k, v := range parentStyle.GetVariables() {
			out.variables[k] = v
		}
	}
	for k, v := range cascaded {
		if strings.HasPrefix(k, "__") {
			out.variables[k] = v.value
		}
	}

	// Set specified value needed for computed value
	_, position := out.cascadeValue("position")
	_, display := out.cascadeValue("display")
	_, float := out.cascadeValue("float")

	if position.SpecialProperty == nil && position.ToCascaded().Default == 0 {
		out.specified.position = position.ToCascaded().ToCSS().(pr.BoolString)
	}
	if display.SpecialProperty == nil && display.ToCascaded().Default == 0 {
		out.specified.display = display.ToCascaded().ToCSS().(pr.Display)
	}
	if float.SpecialProperty == nil && float.ToCascaded().Default == 0 {
		out.specified.float = float.ToCascaded().ToCSS().(pr.String)
	}

	return out
}

func (c *ComputedStyle) isRootElement() bool { return c.parentStyle == nil }

func (c *ComputedStyle) Set(key string, value pr.CssProperty) { c.dict[key] = value }

func (c *ComputedStyle) Copy() pr.ElementStyle {
	out := newComputedStyle(c.parentStyle, c.cascaded, c.element, c.pseudoType, c.rootStyle, c.baseUrl, c.textContext)
	out.dict.UpdateWith(c.dict)
	return out
}

func (c *ComputedStyle) GetParentStyle() pr.ElementStyle { return c.parentStyle }

func (c *ComputedStyle) GetVariables() map[string]pr.ValidatedProperty { return c.variables }

func (c *ComputedStyle) cascadeValue(key string) (pr.DefaultKind, pr.ValidatedProperty) {
	var (
		value   pr.ValidatedProperty
		keyword pr.DefaultKind
	)
	if casc, in := c.cascaded[key]; in {
		value = casc.value
		if value.SpecialProperty == nil {
			keyword = value.ToCascaded().Default
		}
	} else {
		if pr.Inherited.Has(key) {
			keyword = pr.Inherit
		} else {
			keyword = pr.Initial
		}
	}

	if keyword == pr.Inherit && c.isRootElement() {
		// On the Root Element, "inherit" from initial values
		keyword = pr.Initial
	}

	if keyword == pr.Initial {
		value_ := pr.InitialValues[key]
		value = pr.AsCascaded(value_).AsValidated()
	} else if keyword == pr.Inherit {
		value_ := c.parentStyle.Get(key)
		value = pr.AsCascaded(value_).AsValidated()
	}

	return keyword, value
}

// provide on demand computation
func (c *ComputedStyle) Get(key string) pr.CssProperty {
	// check the cache
	if v, has := c.dict[key]; has {
		return v
	}

	keyword, value := c.cascadeValue(key)

	if keyword == pr.Initial && !pr.InitialNotComputed.Has(key) {
		// The value is the same as when computed
		c.dict[key] = value.ToCascaded().ToCSS()
	} else if keyword == pr.Inherit {
		// Values in parentStyle are already computed.
		c.dict[key] = value.ToCascaded().ToCSS()
	}

	if key == "page" && value.SpecialProperty == nil && value.ToCascaded().Default == 0 && value.ToCascaded().ToCSS().(pr.Page).String == "auto" {
		// The page property does not inherit. However, if the page value on
		// an Element is auto, then its used value is the value specified on
		// its nearest ancestor with a non-auto value. When specified on the
		// Root Element, the used value for auto is the empty string.
		value_ := pr.Page{Valid: true, String: ""}
		if c.parentStyle != nil {
			value_ = c.parentStyle.GetPage()
		}
		c.dict.SetPage(value_)
		value = pr.AsCascaded(value_).AsValidated()
	}

	// check the cache again
	if v, has := c.dict[key]; has {
		return v
	}

	var (
		newValue             pr.CascadedProperty
		alreadyComputedValue bool
	)

	// special case for content property. See validation.multiValProperties
	if key == "content" && value.SpecialProperty == nil && value.ToCascaded().Default == 0 {
		contentValue := value.ToCascaded().ToCSS().(pr.SContent)
		resolvedContents := make(pr.ContentProperties, 0, len(contentValue.Contents))
		for _, v := range contentValue.Contents {
			if varData, ok := v.Content.(pr.VarData); ok {
				newValue, alreadyComputedValue = computeVariable(varData, key, c.variables, c.baseUrl, c.parentStyle)
				if newValue.Default != 0 {
					log.Printf("invalid default in content: %v", newValue)
					newValue.Default = 0
				}
				newValue, _ := newValue.ToCSS().(pr.SContent)
				if len(newValue.Contents) == 1 {
					resolvedContents = append(resolvedContents, newValue.Contents[0])
				}
			} else {
				resolvedContents = append(resolvedContents, v)
			}
		}
		contentValue.Contents = resolvedContents
		value = pr.AsCascaded(contentValue).AsValidated()
	}

	if varData, ok := value.SpecialProperty.(pr.VarData); ok {
		newValue, alreadyComputedValue = computeVariable(varData, key, c.variables, c.baseUrl, c.parentStyle)
	} else {
		newValue = value.ToCascaded()
	}
	if newValue.Default != 0 {
		log.Printf("invalid default after variable resolution: %v", newValue)
		newValue.Default = 0
	}

	outValue := newValue.ToCSS()

	fn := computerFunctions[key]
	if fn != nil && !alreadyComputedValue {
		outValue = fn(c, key, outValue)
	}
	c.dict[key] = outValue
	return outValue
}

// AnonymousStyle provides on demand access of computed properties,
// optimized for anonymous boxes
type AnonymousStyle struct {
	dict        pr.Properties
	parentStyle pr.ElementStyle
	variables   map[string]pr.ValidatedProperty
}

func newAnonymousStyle(parentStyle pr.ElementStyle) *AnonymousStyle {
	out := &AnonymousStyle{
		parentStyle: parentStyle,
		variables:   make(map[string]pr.ValidatedProperty),
		dict:        make(pr.Properties),
	}
	// inherit the variables
	if parentStyle != nil {
		for k, v := range parentStyle.GetVariables() {
			out.variables[k] = v
		}
	}
	// border-*-style is none, so border-width computes to zero.
	// Other than that, properties that would need computing are
	// border-*-color, but they do not apply.
	out.dict.SetBorderTopWidth(pr.Value{})
	out.dict.SetBorderBottomWidth(pr.Value{})
	out.dict.SetBorderLeftWidth(pr.Value{})
	out.dict.SetBorderRightWidth(pr.Value{})
	out.dict.SetOutlineWidth(pr.Value{})

	return out
}

func (c *AnonymousStyle) Set(key string, value pr.CssProperty) { c.dict[key] = value }

func (c *AnonymousStyle) Copy() pr.ElementStyle {
	out := newAnonymousStyle(c.parentStyle)
	out.dict.UpdateWith(c.dict)
	return out
}

func (c *AnonymousStyle) GetParentStyle() pr.ElementStyle { return c.parentStyle }

func (c *AnonymousStyle) GetVariables() map[string]pr.ValidatedProperty { return c.variables }

func (a *AnonymousStyle) Get(key string) pr.CssProperty {
	if v, has := a.dict[key]; has {
		return v
	}

	if pr.Inherited.Has(key) || strings.HasPrefix(key, "__") {
		a.dict[key] = a.parentStyle.Get(key)
	} else if key == "page" {
		// page is not inherited but taken from the ancestor if 'auto'
		a.dict[key] = a.parentStyle.Get(key)
	} else {
		a.dict[key] = pr.InitialValues[key]
	}

	return a.dict[key]
}

// ResolveColor return the color for `key`, replacing
// `currentColor` with p["color"]
// It panics if the key has not concrete type `Color`
// replace Python getColor function
func ResolveColor(style pr.ElementStyle, key string) pr.Color {
	value := style.Get(key).(pr.Color)
	if value.Type == parser.ColorCurrentColor {
		return style.GetColor()
	}
	return value
}

func pageMatchT(selectorPageType pageSelector, pageType utils.PageElement) bool {
	if selectorPageType.Side != "" && selectorPageType.Side != pageType.Side {
		return false
	}
	if selectorPageType.Blank && selectorPageType.Blank != pageType.Blank {
		return false
	}
	if selectorPageType.First && selectorPageType.First != pageType.First {
		return false
	}
	if selectorPageType.Name != "" && selectorPageType.Name != pageType.Name {
		return false
	}
	if !selectorPageType.Index.IsNone() {
		a, b := selectorPageType.Index.A, selectorPageType.Index.B
		// TODO: handle group
		if a != 0 {
			if (pageType.Index+1-b)%a != 0 {
				return false
			}
		} else {
			if pageType.Index+1 != b {
				return false
			}
		}
	}
	return true
}

// Yield the stylesheets in ``elementTree``.
// The output order is the same as the source order.
func findStylesheets(wrapperElement *utils.HTMLNode, deviceMediaType string, urlFetcher utils.UrlFetcher, baseUrl string,
	fontConfig *text.FontConfiguration, counterStyle counters.CounterStyle, pageRules *[]PageRule) (out []CSS) {
	sel := cascadia.MustCompile("style, link")
	for _, _element := range sel.MatchAll((*html.Node)(wrapperElement)) {
		element := (*utils.HTMLNode)(_element)
		mimeType := element.Get("type")
		if mimeType == "" {
			mimeType = "text/css"
		}
		mimeType = strings.TrimSpace(strings.SplitN(mimeType, ";", 2)[0])
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
			// Content is text that is directly in the <style> Element, not its
			// descendants
			content := element.GetChildText()
			// ElementTree should give us either unicode or  ASCII-only
			// bytestrings, so we don"t need `encoding` here.
			css, err := NewCSS(utils.InputString(content), baseUrl, urlFetcher, false, deviceMediaType,
				fontConfig, nil, pageRules, counterStyle)
			if err != nil {
				log.Printf("Invalid style %s : %s \n", content, err)
			} else {
				out = append(out, css)
			}
		case atom.Link:
			if element.Get("href") != "" {
				if !element.HasLinkType("stylesheet") || element.HasLinkType("alternate") {
					continue
				}
				href := element.GetUrlAttribute("href", baseUrl, false)
				if href != "" {
					css, err := NewCSS(utils.InputUrl(href), "", urlFetcher, true, deviceMediaType,
						fontConfig, nil, pageRules, counterStyle)
					if err != nil {
						log.Printf("Failed to load stylesheet at %s : %s \n", href, err)
					} else {
						out = append(out, css)
					}
				}
			}
		}
	}
	return out
}

// Yield ``specificity, (Element, declaration, BaseUrl)`` rules.
// Rules from "style" attribute are returned with specificity
// ``(1, 0, 0)``.
// If ``presentationalHints`` is ``true``, rules from presentational hints
// are returned with specificity ``(0, 0, 0)``.
// presentationalHints=false
func findStyleAttributes(tree *utils.HTMLNode, presentationalHints bool, baseUrl string) (out []styleAttrSpec) {
	checkStyleAttribute := func(element *utils.HTMLNode, styleAttribute string) styleAttr {
		declarations := parser.ParseDeclarationList2(styleAttribute, false, false)
		return styleAttr{element: element, declaration: declarations, baseUrl: baseUrl}
	}

	iter := tree.Iter()
	for iter.HasNext() {
		element := iter.Next()
		specificity := cascadia.Specificity{1, 0, 0}
		styleAttribute := element.Get("style")
		if styleAttribute != "" {
			out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, styleAttribute)})
		}
		if !presentationalHints {
			continue
		}
		specificity = cascadia.Specificity{0, 0, 0}
		switch element.DataAtom {
		case atom.Body:
			// TODO: we should check the container frame Element
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
					out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, styleAttribute)})
				}
			}
			if element.Get("background") != "" {
				styleAttribute = fmt.Sprintf("background-image:url(%s)", element.Get("background"))
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("bgcolor") != "" {
				styleAttribute = fmt.Sprintf("background-color:%s", element.Get("bgcolor"))
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("text") != "" {
				styleAttribute = fmt.Sprintf("color:%s", element.Get("text"))
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, styleAttribute)})
			}
		// TODO: we should support link, vlink, alink
		case atom.Center:
			out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, "text-align:center")})
		case atom.Div:
			align := strings.ToLower(element.Get("align"))
			switch align {
			case "middle":
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, "text-align:center")})
			case "center", "left", "right", "justify":
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, fmt.Sprintf("text-align:%s", align))})
			}
		case atom.Font:
			if element.Get("color") != "" {
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, fmt.Sprintf("color:%s", element.Get("color")))})
			}
			if element.Get("face") != "" {
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, fmt.Sprintf("font-family:%s", element.Get("face")))})
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
					out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, fmt.Sprintf("font-size:%s", fontSizes[sizeI]))})
				}
			}
		case atom.Table:
			// TODO: we should support cellpadding
			if element.Get("cellspacing") != "" {
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, fmt.Sprintf("border-spacing:%spx", element.Get("cellspacing")))})
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
						out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(subelement,
							fmt.Sprintf("padding-left:%s;padding-right:%s;padding-top:%s;padding-bottom:%s;", cellpadding, cellpadding, cellpadding, cellpadding))})
					}
				}
			}
			if element.Get("hspace") != "" {
				hspace := element.Get("hspace")
				if utils.IsDigit(hspace) {
					hspace += "px"
				}
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element,
					fmt.Sprintf("margin-left:%s;margin-right:%s", hspace, hspace))})
			}
			if element.Get("vspace") != "" {
				vspace := element.Get("vspace")
				if utils.IsDigit(vspace) {
					vspace += "px"
				}
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element,
					fmt.Sprintf("margin-top:%s;margin-bottom:%s", vspace, vspace))})
			}
			if element.Get("width") != "" {
				styleAttribute = fmt.Sprintf("width:%s", element.Get("width"))
				if utils.IsDigit(element.Get("width")) {
					styleAttribute += "px"
				}
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("height") != "" {
				styleAttribute = fmt.Sprintf("height:%s", element.Get("height"))
				if utils.IsDigit(element.Get("height")) {
					styleAttribute += "px"
				}
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("background") != "" {
				styleAttribute = fmt.Sprintf("background-image:url(%s)", element.Get("background"))
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("bgcolor") != "" {
				styleAttribute = fmt.Sprintf("background-color:%s", element.Get("bgcolor"))
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("bordercolor") != "" {
				styleAttribute = fmt.Sprintf("border-color:%s", element.Get("bordercolor"))
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("border") != "" {
				styleAttribute = fmt.Sprintf("border-width:%spx", element.Get("border"))
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, styleAttribute)})
			}
		case atom.Tr, atom.Td, atom.Th, atom.Thead, atom.Tbody, atom.Tfoot:
			align := strings.ToLower(element.Get("align"))
			if align == "left" || align == "right" || align == "justify" {
				// TODO: we should align descendants too
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, fmt.Sprintf("text-align:%s", align))})
			}
			if element.Get("background") != "" {
				styleAttribute = fmt.Sprintf("background-image:url(%s)", element.Get("background"))
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("bgcolor") != "" {
				styleAttribute = fmt.Sprintf("background-color:%s", element.Get("bgcolor"))
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, styleAttribute)})
			}
			if element.DataAtom == atom.Tr || element.DataAtom == atom.Td || element.DataAtom == atom.Th {
				if element.Get("height") != "" {
					styleAttribute = fmt.Sprintf("height:%s", element.Get("height"))
					if utils.IsDigit(element.Get("height")) {
						styleAttribute += "px"
					}
					out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, styleAttribute)})
				}
				if element.DataAtom == atom.Td || element.DataAtom == atom.Th {
					if element.Get("width") != "" {
						styleAttribute = fmt.Sprintf("width:%s", element.Get("width"))
						if utils.IsDigit(element.Get("width")) {
							styleAttribute += "px"
						}
						out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, styleAttribute)})
					}
				}
			}
		case atom.Caption:
			align := strings.ToLower(element.Get("align"))
			// TODO: we should align descendants too
			if align == "left" || align == "right" || align == "justify" {
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, fmt.Sprintf("text-align:%s", align))})
			}
		case atom.Col:
			if element.Get("width") != "" {
				styleAttribute = fmt.Sprintf("width:%s", element.Get("width"))
				if utils.IsDigit(element.Get("width")) {
					styleAttribute += "px"
				}
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, styleAttribute)})
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
			if element.HasAttr("color") || element.HasAttr("noshade") {
				if size >= 1 {
					out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, fmt.Sprintf("border-width:%dpx", size/2))})
				}
			} else if size == 1 {
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, "border-bottom-width:0")})
			} else if size > 1 {
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, fmt.Sprintf("height:%dpx", size-2))})
			}

			if element.Get("width") != "" {
				styleAttribute = fmt.Sprintf("width:%s", element.Get("width"))
				if utils.IsDigit(element.Get("width")) {
					styleAttribute += "px"
				}
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, styleAttribute)})
			}
			if element.Get("color") != "" {
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, fmt.Sprintf("color:%s", element.Get("color")))})
			}
		case atom.Iframe, atom.Applet, atom.Embed, atom.Img, atom.Input, atom.Object:
			if element.DataAtom != atom.Input || strings.ToLower(element.Get("type")) == "image" {
				align := strings.ToLower(element.Get("align"))
				if align == "middle" || align == "center" {
					// TODO: middle && center values are wrong
					out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, "vertical-align:middle")})
				}
				if element.Get("hspace") != "" {
					hspace := element.Get("hspace")
					if utils.IsDigit(hspace) {
						hspace += "px"
					}
					out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element,
						fmt.Sprintf("margin-left:%s;margin-right:%s", hspace, hspace))})
				}
				if element.Get("vspace") != "" {
					vspace := element.Get("vspace")
					if utils.IsDigit(vspace) {
						vspace += "px"
					}
					out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element,
						fmt.Sprintf("margin-top:%s;margin-bottom:%s", vspace, vspace))})
				}
				// TODO: img seems to be excluded for width && height, but a
				// lot of W3C tests rely on this attribute being applied to img
				if element.Get("width") != "" {
					styleAttribute = fmt.Sprintf("width:%s", element.Get("width"))
					if utils.IsDigit(element.Get("width")) {
						styleAttribute += "px"
					}
					out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, styleAttribute)})
				}
				if element.Get("height") != "" {
					styleAttribute = fmt.Sprintf("height:%s", element.Get("height"))
					if utils.IsDigit(element.Get("height")) {
						styleAttribute += "px"
					}
					out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element, styleAttribute)})
				}
				if element.DataAtom == atom.Img || element.DataAtom == atom.Object || element.DataAtom == atom.Input {
					if element.Get("border") != "" {
						out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element,
							fmt.Sprintf("border-width:%spx;border-style:solid", element.Get("border")))})
					}
				}
			}
		case atom.Ol:
			// From https://www.w3.org/TR/css-lists-3/
			if element.Get("start") != "" {
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element,
					fmt.Sprintf("counter-reset:list-item %s;counter-increment:list-item -1", element.Get("start")))})
			}
		case atom.Ul:
			// From https://www.w3.org/TR/css-lists-3/
			if element.Get("value") != "" {
				out = append(out, styleAttrSpec{specificity: specificity, styleAttr: checkStyleAttribute(element,
					fmt.Sprintf("counter-reset:list-item %s;counter-increment:none", element.Get("value")))})
			}
		}
	}
	return out
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

// Get a dict of computed style mixed from parent and cascaded styles.
func ComputedFromCascaded(element Element, cascaded cascadedStyle, parentStyle pr.ElementStyle, rootStyle pr.Properties, pseudoType, baseUrl string,
	targetCollector *TargetCollector, textContext text.TextLayoutContext) pr.ElementStyle {
	if cascaded == nil && parentStyle != nil {
		return newAnonymousStyle(parentStyle)
	}

	style := newComputedStyle(parentStyle, cascaded, element, pseudoType, rootStyle, baseUrl, textContext)
	if targetCollector != nil {
		targetCollector.collectAnchor(string(style.GetAnchor()))
	}
	return style
}

// either a html node or a page type
type Element interface {
	ToKey(pseudoType string) utils.ElementKey
}

type weight struct {
	precedence  uint8
	specificity cascadia.Specificity
}

func (w weight) isNone() bool {
	return w == weight{}
}

// Less return `true` if w <= other
func (w weight) Less(other weight) bool {
	return w.precedence < other.precedence || (w.precedence == other.precedence && (w.specificity.Less(other.specificity) || w.specificity == other.specificity))
}

type weigthedValue struct {
	value  pr.ValidatedProperty
	weight weight
}

type cascadedStyle = map[string]weigthedValue

// Parse a page selector rule.
//     Return a list of page data if the rule is correctly parsed. Page data are a
//     dict containing:
//     - "side" ("left", "right" or ""),
//     - "blank" (true or false),
//     - "first" (true or false),
//     - "name" (page name string or ""), and
//     - "specificity" (list of numbers).
//     Return ``None` if something went wrong while parsing the rule.
func parsePageSelectors(rule parser.QualifiedRule) (out []pageSelector) {
	// See https://drafts.csswg.org/css-page-3/#syntax-page-selector

	tokens := validation.RemoveWhitespace(*rule.Prelude)

	// TODO: Specificity is probably wrong, should clean and test that.
	if len(tokens) == 0 {
		out = append(out, pageSelector{})
		return out
	}

	for len(tokens) > 0 {
		var types_ pageSelector

		if ident, ok := tokens[0].(parser.IdentToken); ok {
			tokens = tokens[1:]
			types_.Name = string(ident.Value)
			types_.Specificity[0] = 1
		}

		if len(tokens) == 1 {
			return nil
		} else if len(tokens) == 0 {
			out = append(out, types_)
			return out
		}

		for len(tokens) > 0 {
			token_ := tokens[0]
			tokens = tokens[1:]
			literal, ok := token_.(parser.LiteralToken)
			if !ok {
				return nil
			}

			if literal.Value == ":" {
				if len(tokens) == 0 {
					return nil
				}
				switch firstToken := tokens[0].(type) {
				case parser.IdentToken:
					tokens = tokens[1:]
					pseudoClass := firstToken.Value.Lower()
					switch pseudoClass {
					case "left", "right":
						if types_.Side != "" && types_.Side != pseudoClass {
							return nil
						}
						types_.Side = pseudoClass
						types_.Specificity[2] += 1
						continue
					case "blank":
						types_.Blank = true
						types_.Specificity[1] += 1
						continue
					case "first":
						types_.First = true
						types_.Specificity[1] += 1
						continue
					}
				case parser.FunctionBlock:
					tokens = tokens[1:]
					if firstToken.Name != "nth" {
						return nil
					}
					var group []parser.Token
					nth := *firstToken.Arguments
					for i, argument := range *firstToken.Arguments {
						if ident, ok := argument.(parser.IdentToken); ok && ident.Value == "of" {
							nth = (*firstToken.Arguments)[:(i - 1)]
							group = (*firstToken.Arguments)[i:]
						}
					}
					nthValues := parser.ParseNth(nth)
					if nthValues == nil {
						return nil
					}
					if group != nil {
						var group_ []parser.Token
						for _, token := range group {
							if ty := token.Type(); ty != parser.CommentT && ty != parser.WhitespaceTokenT {
								group_ = append(group_, token)
							}
						}
						if len(group_) != 1 {
							return nil
						}
						if _, ok := group_[0].(parser.IdentToken); ok {
							// TODO: handle page groups
							return nil
						}
						return nil
					}
					types_.Index = pageIndex{
						A:     nthValues[0],
						B:     nthValues[1],
						Group: group,
					}
					// TODO: specificity is not specified yet
					// https://github.com/w3c/csswg-drafts/issues/3524
					types_.Specificity[1] += 1
					continue
				}
				return nil

			} else if literal.Value == "," {
				if len(tokens) > 0 && (types_.Specificity != cascadia.Specificity{}) {
					break
				} else {
					return nil
				}
			}
		}
		out = append(out, types_)
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
	pseudoType  string
	pageType    pageSelector
	specificity cascadia.Specificity
}

type PageRule struct {
	rule         parser.AtRule
	selectors    []selectorPageRule
	declarations []validation.ValidatedProperty
}

// Do the work that can be done early on stylesheet, before they are
// in a document.
// ignoreImports = false
func preprocessStylesheet(deviceMediaType, baseUrl string, stylesheetRules []Token,
	urlFetcher utils.UrlFetcher, matcher *matcher, pageRules *[]PageRule,
	fonts *[]string, fontConfig *text.FontConfiguration, counterStyle counters.CounterStyle, ignoreImports bool) {

	for _, rule := range stylesheetRules {
		if atRule, isAtRule := rule.(parser.AtRule); _isContentNone(rule) && (!isAtRule || atRule.AtKeyword.Lower() != "import") {
			continue
		}

		switch rule := rule.(type) {
		case parser.QualifiedRule:
			declarations := validation.PreprocessDeclarations(baseUrl, parser.ParseDeclarationList(*rule.Content, false, false))
			if len(declarations) > 0 {
				prelude := parser.Serialize(*rule.Prelude)
				selector, err := cascadia.ParseGroupWithPseudoElements(prelude)
				if err != nil {
					log.Printf("Invalid or unsupported selector '%s', %s \n", prelude, err)
					continue
				}
				for _, sel := range selector {
					if _, in := pseudoElements[sel.PseudoElement()]; !in {
						err = fmt.Errorf("Unsupported pseudo-element : %s", sel.PseudoElement())
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
			switch rule.AtKeyword.Lower() {
			case "import":
				if ignoreImports {
					log.Printf("@import rule '%s' not at the beginning of the whole rule was ignored. \n",
						parser.Serialize(*rule.Prelude))
					continue
				}

				tokens := validation.RemoveWhitespace(*rule.Prelude)
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
						parser.Serialize(*rule.Prelude))
					continue
				}
				if !evaluateMediaQuery(media, deviceMediaType) {
					continue
				}
				url = utils.UrlJoin(baseUrl, url, false, "@import")
				if url != "" {
					_, err := NewCSS(utils.InputUrl(url), "", urlFetcher, false,
						deviceMediaType, fontConfig, matcher, pageRules, counterStyle)
					if err != nil {
						log.Printf("Failed to load stylesheet at %s : %s \n", url, err)
					}
				}
			case "media":
				media := parseMediaQuery(*rule.Prelude)
				if media == nil {
					log.Printf("Invalid media type '%s' the whole @media rule was ignored. \n",
						parser.Serialize(*rule.Prelude))
					continue
				}
				ignoreImports = true
				if !evaluateMediaQuery(media, deviceMediaType) {
					continue
				}
				contentRules := parser.ParseRuleList(*rule.Content, false, false)
				preprocessStylesheet(
					deviceMediaType, baseUrl, contentRules, urlFetcher,
					matcher, pageRules, fonts, fontConfig, counterStyle, true)
			case "page":
				data := parsePageSelectors(rule.QualifiedRule)
				if data == nil {
					log.Printf("Unsupported @page selector '%s', the whole @page rule was ignored. \n",
						parser.Serialize(*rule.Prelude))
					continue
				}
				ignoreImports = true
				for _, pageType := range data {
					specificity := pageType.Specificity
					pageType.Specificity = cascadia.Specificity{}
					content := parser.ParseDeclarationList(*rule.Content, false, false)
					declarations := validation.PreprocessDeclarations(baseUrl, content)

					var selectors []selectorPageRule
					if len(declarations) > 0 {
						selectors = []selectorPageRule{{specificity: specificity, pseudoType: "", pageType: pageType}}
						*pageRules = append(*pageRules, PageRule{rule: rule, selectors: selectors, declarations: declarations})
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
								pageType: pageType,
							}}
							*pageRules = append(*pageRules, PageRule{rule: atRule, selectors: selectors, declarations: declarations})
						}
					}
				}
			case "font-face":
				ignoreImports = true
				content := parser.ParseDeclarationList(*rule.Content, false, false)
				ruleDescriptors := validation.PreprocessFontFaceDescriptors(baseUrl, content)
				if ruleDescriptors.Src == nil {
					log.Printf(`Missing src descriptor in "@font-face" rule at %d:%d`+"\n",
						rule.Position().Line, rule.Position().Column)
				}
				if ruleDescriptors.FontFamily == "" {
					log.Printf(`Missing font-family descriptor in "@font-face" rule at %d:%d`+"\n",
						rule.Position().Line, rule.Position().Column)
				}
				if ruleDescriptors.Src != nil && ruleDescriptors.FontFamily != "" && fontConfig != nil {
					fontFilename := fontConfig.AddFontFace(ruleDescriptors, urlFetcher)
					if fontFilename != "" {
						*fonts = append(*fonts, fontFilename)
					}
				}

			case "counter-style":
				name := validation.ParseCounterStyleName(*rule.Prelude, counterStyle)
				if name == "" {
					log.Printf(`Invalid counter style name %s, the whole @counter-style rule was ignored at %d:%d.`,
						parser.Serialize(*rule.Prelude), rule.Position().Line, rule.Position().Column)
					continue
				}

				ignoreImports = true
				content := parser.ParseDeclarationList(*rule.Content, false, false)
				desc := validation.PreprocessCounterStyleDescriptors(baseUrl, content)

				if err := desc.Validate(); err != nil {
					log.Printf("In counter style %s at %d:%d, %s", name, rule.Line, rule.Column, err)
					continue
				}

				counterStyle[name] = desc
			}
		}
	}
}

type sheet struct {
	sheet       CSS
	origin      string
	specificity []int
}

type styleAttr struct {
	element     *utils.HTMLNode
	baseUrl     string
	declaration []Token
}

type styleAttrSpec struct {
	styleAttr
	specificity cascadia.Specificity
}

// Compute all the computed styles of all elements in `html` document.
// Do everything from finding author stylesheets to parsing and applying them.
//
// Return a `StyleFor` function like object that takes an Element and an optional
// pseudo-Element type, and return a style dict object.
// presentationalHints=false
func GetAllComputedStyles(html *HTML, userStylesheets []CSS,
	presentationalHints bool, fontConfig *text.FontConfiguration,
	counterStyle counters.CounterStyle, pageRules *[]PageRule,
	targetCollector *TargetCollector, textContext text.TextLayoutContext) *StyleFor {

	if counterStyle == nil {
		counterStyle = make(counters.CounterStyle)
	}
	// add the UA counters
	for k, v := range UACounterStyle {
		counterStyle[k] = v
	}

	// List stylesheets. Order here is not important ("origin" is).
	sheets := []sheet{
		{sheet: html.UAStyleSheet, origin: "user agent", specificity: nil},
	}

	if presentationalHints {
		sheets = append(sheets, sheet{sheet: html.PHStyleSheet, origin: "author", specificity: []int{0, 0, 0}})
	}
	authorShts := findStylesheets(html.Root, html.mediaType, html.UrlFetcher,
		html.BaseUrl, fontConfig, counterStyle, pageRules)
	for _, sht := range authorShts {
		sheets = append(sheets, sheet{sheet: sht, origin: "author", specificity: nil})
	}
	for _, sht := range userStylesheets {
		sheets = append(sheets, sheet{sheet: sht, origin: "user", specificity: nil})
	}
	return newStyleFor(html, sheets, presentationalHints, targetCollector, textContext)
}

// Set style for page types and pseudo-types matching ``pageType``.
func (styleFor StyleFor) SetPageComputedStylesT(pageType utils.PageElement, html *HTML) {
	styleFor.addPageDeclarations(pageType)

	// Apply style for page
	// @page inherits from the Root Element :
	// http://lists.w3.org/Archives/Public/www-style/2012Jan/1164.html
	styleFor.SetComputedStyles(pageType, html.Root, html.Root, "", html.BaseUrl, nil)

	// Apply style for page pseudo-elements (margin boxes)
	for key := range styleFor.CascadedStyles {
		// Element, pseudoType = key
		if key.PseudoType != "" && key.PageType == pageType {
			// The pseudo-Element inherits from the Element.
			styleFor.SetComputedStyles(key.PageType, key.PageType, html.Root, key.PseudoType, html.BaseUrl, nil)
		}
	}
}
