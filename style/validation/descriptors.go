package validation

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/benoitkugler/go-weasyprint/utils"

	"github.com/benoitkugler/go-weasyprint/style/parser"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
)

// Validate descriptors, currently used for @font-face rules.
// See https://www.w3.org/TR/css-fonts-3/#font-resources.

// :copyright: Copyright 2011-2016 Simon Sapin && contributors, see AUTHORS.
// :license: BSD, see LICENSE for details.

var descriptors = map[string]descriptor{
	"font-family":           fontFamilyDescriptor,
	"src":                   src,
	"font-style":            fontStyleDescriptor,
	"font-weight":           fontWeightDescriptor,
	"font-stretch":          fontStretchDescriptor,
	"font-feature-settings": fontFeatureSettingsDescriptor,
	"font-variant":          fontVariant,
}

type RuleDescriptors struct {
	Src                 []pr.NamedString
	FontFamily          pr.String
	FontStyle           pr.String
	FontWeight          pr.IntString
	FontStretch         pr.String
	FontFeatureSettings pr.SIntStrings
	FontVariant         pr.NamedProperties
}

type descriptor = func(tokens []Token, baseUrl string, out *RuleDescriptors) error

// @descriptor()
// ``font-family`` descriptor validation.
// allowSpaces = false
func _fontFamilyDesc(tokens []Token, allowSpaces bool) string {
	allowedTypes := utils.Set{string(parser.TypeIdentToken): utils.Has}
	if allowSpaces {
		allowedTypes.Add(string(parser.TypeWhitespaceToken))
	}
	if len(tokens) == 1 {
		if str, ok := tokens[0].(parser.StringToken); ok {
			return str.Value
		}
	}

	var values []string
	ok := true
	for _, token := range tokens {
		ok = ok && allowedTypes.Has(string(token.Type()))
		if ident, isToken := token.(parser.IdentToken); isToken {
			values = append(values, string(ident.Value))
		}
	}
	if len(tokens) > 0 && ok {
		return strings.Join(values, " ")
	}
	return ""
}

func fontFamilyDescriptor(tokens []Token, _ string, out *RuleDescriptors) error {
	s := _fontFamilyDesc(tokens, false)
	out.FontFamily = pr.String(s)
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

func src(tokens []Token, baseUrl string, out *RuleDescriptors) error {
	for _, part := range SplitOnComma(tokens) {
		result, err := _src(RemoveWhitespace(part), baseUrl)
		if err != nil {
			return err
		}
		if result, ok := result.(pr.NamedString); ok {
			out.Src = append(out.Src, result)
		} else {
			return fmt.Errorf("invalid <src> descriptor: %v", part)
		}
	}
	return nil
}

// @descriptor()
// @singleKeyword
// ``font-style`` descriptor validation.
func fontStyleDescriptor(tokens []Token, _ string, out *RuleDescriptors) error {
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
func fontWeightDescriptor(tokens []Token, _ string, out *RuleDescriptors) error {
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
func fontStretchDescriptor(tokens []Token, _ string, out *RuleDescriptors) error {
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
func fontFeatureSettingsDescriptor(tokens []Token, _ string, out *RuleDescriptors) error {
	s := _fontFeatureSettings(tokens)
	if s.IsNone() {
		return InvalidValue
	}
	out.FontFeatureSettings = s
	return nil
}

// @descriptor()
// ``font-variant`` descriptor validation.
func fontVariant(tokens []Token, _ string, out *RuleDescriptors) error {
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

// Default validator for descriptors.
func validate(baseUrl, name string, tokens []Token, out *RuleDescriptors) error {
	function, ok := descriptors[name]
	if !ok {
		return errors.New("descriptor not supported")
	}

	err := function(tokens, baseUrl, out)
	return err
}

// Filter unsupported names and values for descriptors.
// Log a warning for every ignored descriptor.
// Return a iterable of ``(name, value)`` tuples.
func PreprocessDescriptors(baseUrl string, descriptors []Token) RuleDescriptors {
	var out RuleDescriptors
	for _, descriptor := range descriptors {
		decl, ok := descriptor.(parser.Declaration)
		if !ok || decl.Important {
			continue
		}
		tokens := RemoveWhitespace(decl.Value)
		name := string(decl.Name)
		err := validate(baseUrl, name, tokens, &out)
		if err != nil {
			log.Printf("Ignored `%s:%s` at %d:%d, %s.\n",
				name, parser.Serialize(decl.Value), decl.Position().Line, decl.Position().Column, err)
			continue
		}
	}
	return out
}
