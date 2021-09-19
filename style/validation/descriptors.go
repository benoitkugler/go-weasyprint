package validation

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/benoitkugler/go-weasyprint/boxes/counters"
	"github.com/benoitkugler/go-weasyprint/utils"

	"github.com/benoitkugler/go-weasyprint/style/parser"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
)

// Validate descriptors, currently used for @font-face rules.
// See https://www.w3.org/TR/css-fonts-3/#font-resources.

// :copyright: Copyright 2011-2016 Simon Sapin && contributors, see AUTHORS.
// :license: BSD, see LICENSE for details.

var (
	fontFaceDescriptors = map[string]fontFaceDescriptorParser{
		"font-family":           fontFamilyDescriptor,
		"src":                   src,
		"font-style":            fontStyleDescriptor,
		"font-weight":           fontWeightDescriptor,
		"font-stretch":          fontStretchDescriptor,
		"font-feature-settings": fontFeatureSettingsDescriptor,
		"font-variant":          fontVariant,
	}

	counterStyleDescriptors = map[string]counterStyleDescriptorParser{
		"system":           system,
		"negative":         negative,
		"prefix":           prefix,
		"suffix":           suffix,
		"range":            rangeD,
		"pad":              pad,
		"fallback":         fallback,
		"symbols":          symbols,
		"additive-symbols": additiveSymbols,
	}
)

type FontFaceDescriptors struct {
	Src                 []pr.NamedString
	FontFamily          pr.String
	FontStyle           pr.String
	FontWeight          pr.IntString
	FontStretch         pr.String
	FontFeatureSettings pr.SIntStrings
	FontVariant         pr.NamedProperties
}

type fontFaceDescriptorParser = func(tokens []Token, baseUrl string, out *FontFaceDescriptors) error

// @descriptor()
// ``font-family`` descriptor validation.
// allowSpaces = false
func _fontFamilyDesc(tokens []Token, allowSpaces bool) string {
	allowedsT := utils.Set{string(parser.IdentTokenT): utils.Has}
	if allowSpaces {
		allowedsT.Add(string(parser.WhitespaceTokenT))
	}
	if len(tokens) == 1 {
		if str, ok := tokens[0].(parser.StringToken); ok {
			return str.Value
		}
	}

	var values []string
	ok := true
	for _, token := range tokens {
		ok = ok && allowedsT.Has(string(token.Type()))
		if ident, isToken := token.(parser.IdentToken); isToken {
			values = append(values, string(ident.Value))
		}
	}
	if len(tokens) > 0 && ok {
		return strings.Join(values, " ")
	}
	return ""
}

func fontFamilyDescriptor(tokens []Token, _ string, out *FontFaceDescriptors) error {
	s := _fontFamilyDesc(tokens, false)
	out.FontFamily = pr.String(s)
	if s == "" {
		return InvalidValue
	}
	return nil
}

// @descriptor(wantsBaseUrl=true)
// @commaSeparatedList
// ``src`` descriptor validation.
func _src(tokens []Token, baseUrl string) (pr.InnerContent, error) {
	if len(tokens) > 0 && len(tokens) <= 2 {
		token := tokens[len(tokens)-1]
		tokens = tokens[:len(tokens)-1]
		if fn, ok := token.(parser.FunctionBlock); ok && fn.Name.Lower() == "format" {
			tokens, token = tokens[:len(tokens)-1], tokens[len(tokens)-1]
		}
		if fn, ok := token.(parser.FunctionBlock); ok && fn.Name.Lower() == "local" {
			return pr.NamedString{Name: "local", String: _fontFamilyDesc(*fn.Arguments, true)}, nil
		}
		url, _, err := getUrl(token, baseUrl)
		if err != nil {
			return nil, err
		}
		if !url.IsNone() {
			return url, nil
		}
	}
	return nil, nil
}

func src(tokens []Token, baseUrl string, out *FontFaceDescriptors) error {
	for _, part := range SplitOnComma(tokens) {
		result, err := _src(RemoveWhitespace(part), baseUrl)
		if err != nil {
			return err
		}
		if result, ok := result.(pr.NamedString); ok {
			out.Src = append(out.Src, result)
		} else {
			return InvalidValue
		}
	}
	return nil
}

// @descriptor()
// @singleKeyword
// ``font-style`` descriptor validation.
func fontStyleDescriptor(tokens []Token, _ string, out *FontFaceDescriptors) error {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "normal", "italic", "oblique":
		out.FontStyle = pr.String(keyword)
		return nil
	default:
		return fmt.Errorf("unsupported font-style descriptor: %s", keyword)
	}
}

// @descriptor()
// @singleToken
// ``font-weight`` descriptor validation.
func fontWeightDescriptor(tokens []Token, _ string, out *FontFaceDescriptors) error {
	if len(tokens) != 1 {
		return InvalidValue
	}
	token := tokens[0]
	keyword := getKeyword(token)
	if keyword == "normal" || keyword == "bold" {
		out.FontWeight = pr.IntString{String: keyword}
		return nil
	}
	if number, ok := token.(parser.NumberToken); ok && number.IsInteger {
		v := number.IntValue()
		switch v {
		case 100, 200, 300, 400, 500, 600, 700, 800, 900:
			out.FontWeight = pr.IntString{Int: v}
			return nil
		}
	}
	return InvalidValue
}

// @descriptor()
// @singleKeyword
// Validation for the ``font-stretch`` descriptor.
func fontStretchDescriptor(tokens []Token, _ string, out *FontFaceDescriptors) error {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "ultra-condensed", "extra-condensed", "condensed", "semi-condensed",
		"normal",
		"semi-expanded", "expanded", "extra-expanded", "ultra-expanded":
		out.FontStretch = pr.String(keyword)
		return nil
	default:
		return fmt.Errorf("unsupported font-stretch descriptor: %s", keyword)
	}
}

// @descriptor("font-feature-settings")
// ``font-feature-settings`` descriptor validation.
func fontFeatureSettingsDescriptor(tokens []Token, _ string, out *FontFaceDescriptors) error {
	s := _fontFeatureSettings(tokens)
	if s.IsNone() {
		return InvalidValue
	}
	out.FontFeatureSettings = s
	return nil
}

// @descriptor()
// ``font-variant`` descriptor validation.
func fontVariant(tokens []Token, _ string, out *FontFaceDescriptors) error {
	if len(tokens) == 1 {
		keyword := getKeyword(tokens[0])
		if keyword == "normal" || keyword == "none" || keyword == "inherit" {
			return nil
		}
	}
	var values pr.NamedProperties
	expanded, err := expandFontVariant(tokens)
	if err != nil {
		return err
	}
	for _, subTokens := range expanded {
		prop, err := validateNonShorthand("", "font-variant"+subTokens.name, subTokens.tokens, true)
		if err != nil {
			return InvalidValue
		}
		values = append(values, prop)
	}
	out.FontVariant = values
	return nil
}

func PreprocessFontFaceDescriptors(baseUrl string, descriptors []Token) FontFaceDescriptors {
	var out FontFaceDescriptors
	preprocessDescriptors(baseUrl, descriptors, &out)
	return out
}

// counter-style

type csDescriptors counters.CounterStyleDescriptors

type counterStyleDescriptorParser = func(tokens []Token, baseUrl string, out *csDescriptors) error

// ``system`` descriptor validation.
func system(tokens []Token, _ string, out *csDescriptors) error {
	if len(tokens) == 0 || len(tokens) > 2 {
		return InvalidValue
	}

	switch keyword := getKeyword(tokens[0]); keyword {
	case "extends":
		if len(tokens) == 2 {
			secondKeyword := getKeyword(tokens[1])
			if secondKeyword != "" {
				out.System = counters.CounterStyleSystem{Keyword: keyword, SecondKeyword: secondKeyword, Number: 0}
				return nil
			}
		}
	case "fixed":
		if len(tokens) == 1 {
			out.System = counters.CounterStyleSystem{Keyword: "", SecondKeyword: "fixed", Number: 1}
			return nil
		} else if numb, ok := tokens[1].(parser.NumberToken); ok && numb.IsInteger {
			out.System = counters.CounterStyleSystem{Keyword: "", SecondKeyword: "fixed", Number: numb.IntValue()}
			return nil
		}
	case "cyclic", "numeric", "alphabetic", "symbolic", "additive":
		if len(tokens) == 1 {
			out.System = counters.CounterStyleSystem{Keyword: "", SecondKeyword: keyword, Number: 0}
			return nil
		}
	}

	return InvalidValue
}

// match a StringToken, IdentToken, or a valid url
func stringIdentOrUrl(token Token, baseUrl string) (pr.NamedString, bool) {
	switch token := token.(type) {
	case parser.StringToken:
		return pr.NamedString{Name: "string", String: token.Value}, true
	case parser.IdentToken:
		return pr.NamedString{Name: "string", String: string(token.Value)}, true
	default:
		url, _, _ := getUrl(token, baseUrl)
		if url.Name == "url" {
			return url, true
		}
	}
	return pr.NamedString{}, false
}

// ``negative`` descriptor validation.
func negative(tokens []Token, baseUrl string, out *csDescriptors) error {
	if len(tokens) > 2 {
		return InvalidValue
	}

	var values []pr.NamedString
	for len(tokens) != 0 {
		var token Token
		token, tokens = tokens[len(tokens)-1], tokens[:len(tokens)-1]
		if p, ok := stringIdentOrUrl(token, baseUrl); ok {
			values = append(values, p)
		}
	}

	if len(values) == 1 {
		values = append(values, pr.NamedString{Name: "string", String: ""})
	}

	if len(values) == 2 {
		copy(out.Negative[:], values)
		return nil
	}

	return InvalidValue
}

// @descriptor("counter-style", "prefix", wantsBaseUrl=true)
// @descriptor("counter-style", "suffix", wantsBaseUrl=true)

func prefix(tokens []Token, baseUrl string, out *csDescriptors) (err error) {
	out.Prefix, err = _prefixSuffix(tokens, baseUrl)
	return err
}

func suffix(tokens []Token, baseUrl string, out *csDescriptors) (err error) {
	out.Suffix, err = _prefixSuffix(tokens, baseUrl)
	return err
}

// ``prefix`` && ``suffix`` descriptors validation.
func _prefixSuffix(tokens []Token, baseUrl string) (pr.NamedString, error) {
	if len(tokens) != 1 {
		return pr.NamedString{}, InvalidValue
	}
	token := tokens[0]
	if p, ok := stringIdentOrUrl(token, baseUrl); ok {
		return p, nil
	}
	return pr.NamedString{}, InvalidValue
}

// @descriptor("counter-style")
// @commaSeparatedList
// ``range`` descriptor validation.
func rangeD(tokens []Token, _ string, out *csDescriptors) error {
	if len(tokens) == 1 {
		keyword := getSingleKeyword(tokens)
		if keyword == "auto" {
			out.Range = pr.OptionalRanges{Auto: true}
			return nil
		}
	}

	for _, part := range SplitOnComma(tokens) {
		result, err := range_(RemoveWhitespace(part))
		if err != nil {
			return err
		}
		out.Range.Ranges = append(out.Range.Ranges, result)
	}
	return nil
}

func range_(tokens []Token) ([2]int, error) {
	if len(tokens) == 2 {
		var values [2]int
		for i, token := range tokens {
			switch token := token.(type) {
			case parser.IdentToken:
				if token.Value == "infinite" {
					values[i] = math.MaxInt32
					continue
				}
			case parser.NumberToken:
				if token.IsInteger {
					values[i] = token.IntValue()
					continue
				}
			}
			return [2]int{}, InvalidValue
		}
		if values[0] <= values[1] {
			return values, nil
		}
	}
	return [2]int{}, InvalidValue
}

// @descriptor("counter-style", wantsBaseUrl=true)
// ``pad`` descriptor validation.
func pad(tokens []Token, baseUrl string, out *csDescriptors) error {
	v, err := pad_(tokens, baseUrl)
	if err != nil {
		return err
	}
	out.Pad = v
	return nil
}

func pad_(tokens []Token, baseUrl string) (out pr.IntNamedString, err error) {
	var hasLength, hasSymbol bool

	if len(tokens) != 2 {
		return out, InvalidValue
	}

	for _, token := range tokens {
		switch token := token.(type) {
		case parser.NumberToken:
			if token.IsInteger && token.Value >= 0 && !hasLength {
				out.Int = token.IntValue()
				hasLength = true
			}
		default:
			if p, ok := stringIdentOrUrl(token, baseUrl); ok {
				out.NamedString = p
				hasSymbol = true
			}
		}
	}

	if !(hasLength && hasSymbol) {
		return out, InvalidValue
	}

	return out, nil
}

// @descriptor("counter-style")
// @singleToken
// ``fallback`` descriptor validation.
func fallback(tokens []Token, _ string, out *csDescriptors) error {
	if len(tokens) != 1 {
		return InvalidValue
	}
	token := tokens[0]
	ident := getCustomIdent(token)
	if ident == "none" {
		return InvalidValue
	}
	out.Fallback = ident
	return nil
}

// @descriptor("counter-style", wantsBaseUrl=true)
// ``symbols`` descriptor validation.
func symbols(tokens []Token, baseUrl string, out *csDescriptors) error {
	for _, token := range tokens {
		if p, ok := stringIdentOrUrl(token, baseUrl); ok {
			out.Symbols = append(out.Symbols, p)
		} else {
			return InvalidValue
		}
	}
	return nil
}

// @descriptor("counter-style", wantsBaseUrl=true)
// ``additive-symbols`` descriptor validation.
func additiveSymbols(tokens []Token, baseUrl string, out *csDescriptors) error {
	for _, part := range SplitOnComma(tokens) {
		result, err := pad_(RemoveWhitespace(part), baseUrl)
		if err != nil {
			return err
		}
		if L := len(out.AdditiveSymbols); L != 0 && out.AdditiveSymbols[L-1].Int <= result.Int {
			return InvalidValue
		}
		out.AdditiveSymbols = append(out.AdditiveSymbols, result)
	}
	return nil
}

func PreprocessCounterStyleDescriptors(baseUrl string, descriptors []Token) counters.CounterStyleDescriptors {
	var out counters.CounterStyleDescriptors
	preprocessDescriptors(baseUrl, descriptors, (*csDescriptors)(&out))
	return out
}

type parsedDescriptor interface {
	validateDescriptor(baseUrl, name string, tokens []Token) error
}

// Default validator for descriptors.
func (d *FontFaceDescriptors) validateDescriptor(baseUrl, name string, tokens []Token) error {
	function, ok := fontFaceDescriptors[name]
	if !ok {
		return errors.New("descriptor not supported")
	}

	err := function(tokens, baseUrl, d)
	return err
}

func (d *csDescriptors) validateDescriptor(baseUrl, name string, tokens []Token) error {
	function, ok := counterStyleDescriptors[name]
	if !ok {
		return errors.New("descriptor not supported")
	}

	err := function(tokens, baseUrl, d)
	return err
}

// Filter unsupported names and values for descriptors.
// Log a warning for every ignored descriptor.
func preprocessDescriptors(baseUrl string, descriptors []Token, out parsedDescriptor) {
	for _, descriptor := range descriptors {
		decl, ok := descriptor.(parser.Declaration)
		if !ok || decl.Important {
			continue
		}
		tokens := RemoveWhitespace(decl.Value)
		name := string(decl.Name)
		err := out.validateDescriptor(baseUrl, name, tokens)
		if err != nil {
			log.Printf("Ignored `%s:%s` at %d:%d, %s.\n",
				name, parser.Serialize(decl.Value), decl.Position().Line, decl.Position().Column, err)
			continue
		}
	}
}
