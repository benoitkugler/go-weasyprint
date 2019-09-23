package validation

import (
	"errors"
	"log"
	"strings"

	"github.com/benoitkugler/go-weasyprint/style/css"
	"github.com/benoitkugler/go-weasyprint/style/parser"
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

type NamedDescriptor struct {
	Name       string
	Descriptor css.Descriptor
}

type descriptor = func(tokens []Token, baseUrl string) (css.Descriptor, error)

// @descriptor()
// ``font-family`` descriptor validation.
// allowSpaces = false
func _fontFamilyDesc(tokens []Token, allowSpaces bool) string {
	allowedTypes := css.Set{string(parser.TypeIdentToken): css.Has}
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
		if ident, ok := token.(parser.IdentToken); ok {
			values = append(values, string(ident.Value))
		}
	}
	if len(tokens) > 0 && ok {
		return strings.Join(values, " ")
	}
	return ""
}

func fontFamilyDescriptor(tokens []Token, _ string) (css.Descriptor, error) {
	s := _fontFamilyDesc(tokens, false)
	if s == "" {
		return nil, nil
	}
	return css.String(s), nil
}

// @descriptor(wantsBaseUrl=true)
// @commaSeparatedList
// ``src`` descriptor validation.
func _src(tokens []Token, baseUrl string) (css.InnerContents, error) {
	if len(tokens) > 0 && len(tokens) <= 2 {
		token := tokens[len(tokens)-1]
		tokens = tokens[:len(tokens)-1]
		if fn, ok := token.(parser.FunctionBlock); ok && fn.Name.Lower() == "format" {
			tokens, token = tokens[:len(tokens)-1], tokens[len(tokens)-1]
		}
		if fn, ok := token.(parser.FunctionBlock); ok && fn.Name.Lower() == "local" {
			return css.NamedString{Name: "local", String: _fontFamilyDesc(*fn.Arguments, true)}, nil
		}
		url, err := getUrl(token, baseUrl)
		if err != nil {
			return nil, err
		}
		if !url.IsNone() && url.Type == "url" {
			return url.Content, nil
		}
	}
	return nil, nil
}

func src(tokens []Token, baseUrl string) (css.Descriptor, error) {
	var out css.Contents
	for _, part := range SplitOnComma(tokens) {
		result, err := _src(RemoveWhitespace(part), baseUrl)
		if err != nil {
			return nil, err
		}
		if result == nil {
			return nil, nil
		}
		out = append(out, result)
	}
	return out, nil
}

// @descriptor()
// @singleKeyword
// ``font-style`` descriptor validation.
func fontStyleDescriptor(tokens []Token, _ string) (css.Descriptor, error) {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "normal", "italic", "oblique":
		return css.String(keyword), nil
	default:
		return nil, nil
	}
}

// @descriptor()
// @singleToken
// ``font-weight`` descriptor validation.
func fontWeightDescriptor(tokens []Token, _ string) (css.Descriptor, error) {
	if len(tokens) != 1 {
		return nil, nil
	}
	token := tokens[0]
	keyword := getKeyword(token)
	if keyword == "normal" || keyword == "bold" {
		return css.IntString{String: keyword}, nil
	}
	if number, ok := token.(parser.NumberToken); ok && number.IsInteger {
		v := number.IntValue()
		switch v {
		case 100, 200, 300, 400, 500, 600, 700, 800, 900:
			return css.IntString{Int: v}, nil
		}
	}
	return nil, nil
}

// @descriptor()
// @singleKeyword
// Validation for the ``font-stretch`` descriptor.
func fontStretchDescriptor(tokens []Token, _ string) (css.Descriptor, error) {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "ultra-condensed", "extra-condensed", "condensed", "semi-condensed",
		"normal",
		"semi-expanded", "expanded", "extra-expanded", "ultra-expanded":
		return css.String(keyword), nil
	default:
		return nil, nil
	}
}

// @descriptor("font-feature-settings")
// ``font-feature-settings`` descriptor validation.
func fontFeatureSettingsDescriptor(tokens []Token, _ string) (css.Descriptor, error) {
	s := _fontFeatureSettings(tokens)
	if s.IsNone() {
		return nil, nil
	}
	return s, nil
}

// @descriptor()
// ``font-variant`` descriptor validation.
func fontVariant(tokens []Token, _ string) (css.Descriptor, error) {
	if len(tokens) == 1 {
		keyword := getKeyword(tokens[0])
		if keyword == "normal" || keyword == "none" || keyword == "inherit" {
			return css.NamedProperties{}, nil
		}
	}
	var values css.NamedProperties
	expanded, err := expandFontVariant(tokens)
	if err != nil {
		return nil, err
	}
	for _, subTokens := range expanded {
		prop, err := validateNonShorthand("", "font-variant"+subTokens.Name, subTokens.Tokens, true)
		if err != nil {
			return nil, nil
		}
		values = append(values, prop)

	}
	return values, nil
}

// Default validator for descriptors.
func validate(baseUrl, name string, tokens []Token) (css.Descriptor, error) {
	function, ok := descriptors[name]
	if !ok {
		return nil, errors.New("descriptor not supported")
	}

	value, err := function(tokens, baseUrl)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, InvalidValue
	}
	return value, nil
}

// Filter unsupported names and values for descriptors.
//     Log a warning for every ignored descriptor.
//     Return a iterable of ``(name, value)`` tuples.
//
func PreprocessDescriptors(baseUrl string, descriptors []Token) []NamedDescriptor {
	var out []NamedDescriptor
	for _, descriptor := range descriptors {
		decl, ok := descriptor.(parser.Declaration)
		if !ok || decl.Important {
			continue
		}
		tokens := RemoveWhitespace(decl.Value)
		name := string(decl.Name)
		result, err := validate(baseUrl, name, tokens)

		if err != nil {
			log.Printf("Ignored `%s:%s` at %d:%d, %s.\n", name, parser.Serialize(decl.Value), decl.Position().Line, decl.Position().Column, err)
			continue
		}

		out = append(out, NamedDescriptor{
			Name:       strings.ReplaceAll(name, "-", "_"),
			Descriptor: result,
		})

	}
	return out
}
