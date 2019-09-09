package validation

import (
	"errors"
	"log"
	"strings"

	. "github.com/benoitkugler/go-weasyprint/css"
	"github.com/benoitkugler/go-weasyprint/utils"
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

type Descriptor interface{}

type NamedDescriptor struct {
	Name       string
	Descriptor Descriptor
}

type descriptor = func(tokens []Token, baseUrl string) (Descriptor, error)

// @descriptor()
// ``font-family`` descriptor validation.
// allowSpaces = false
func _fontFamilyDesc(tokens []Token, allowSpaces bool) string {
	allowedTypes := Set{string(TypeIdentToken): Has}
	if allowSpaces {
		allowedTypes.Add(string(TypeWhitespaceToken))
	}
	if len(tokens) == 1 {
		if str, ok := tokens[0].(StringToken); ok {
			return str.Value
		}
	}

	var values []string
	ok := true
	for _, token := range tokens {
		ok = ok && allowedTypes.Has(string(token.Type()))
		if ident, ok := token.(IdentToken); ok {
			values = append(values, string(ident.Value))
		}
	}
	if len(tokens) > 0 && ok {
		return strings.Join(values, " ")
	}
	return ""
}

func fontFamilyDescriptor(tokens []Token, _ string) (Descriptor, error) {
	s := _fontFamilyDesc(tokens, false)
	if s == "" {
		return nil, nil
	}
	return s, nil
}

// @descriptor(wantsBaseUrl=true)
// @commaSeparatedList
// ``src`` descriptor validation.
func _src(tokens []Token, baseUrl string) (NamedString, error) {
	if len(tokens) > 0 && len(tokens) <= 2 {
		token := tokens[len(tokens)-1]
		tokens = tokens[:len(tokens)-1]
		if fn, ok := token.(FunctionBlock); ok && fn.Name.Lower() == "format" {
			tokens, token = tokens[:len(tokens)-1], tokens[len(tokens)-1]
		}
		if fn, ok := token.(FunctionBlock); ok && fn.Name.Lower() == "local" {
			return NamedString{Name: "local", String: _fontFamilyDesc(fn.Arguments, true)}, nil
		}
		if url, ok := token.(URLToken); ok {
			if strings.HasPrefix(url.Value, "#") {
				trimed := strings.TrimPrefix(url.Value, "#")
				return NamedString{Name: "internal", String: utils.Unquote(trimed)}, nil
			} else {
				s, err := safeUrljoin(baseUrl, url.Value)
				if err != nil {
					return NamedString{}, err
				}
				return NamedString{Name: "external", String: s}, nil
			}
		}
	}
	return NamedString{}, nil
}

func src(tokens []Token, baseUrl string) (Descriptor, error) {
	var out []NamedString
	for _, part := range SplitOnComma(tokens) {
		result, err := _src(RemoveWhitespace(part), baseUrl)
		if err != nil {
			return nil, err
		}
		if (result == NamedString{}) {
			return nil, nil
		}
		out = append(out, result)
	}
	return out, nil
}

// @descriptor()
// @singleKeyword
// ``font-style`` descriptor validation.
func fontStyleDescriptor(tokens []Token, _ string) (Descriptor, error) {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "normal", "italic", "oblique":
		return keyword, nil
	default:
		return nil, nil
	}
}

// @descriptor()
// @singleToken
// ``font-weight`` descriptor validation.
func fontWeightDescriptor(tokens []Token, _ string) (Descriptor, error) {
	if len(tokens) != 1 {
		return nil, nil
	}
	token := tokens[0]
	keyword := getKeyword(token)
	if keyword == "normal" || keyword == "bold" {
		return keyword, nil
	}
	if number, ok := token.(NumberToken); ok && number.IsInteger {
		v := number.IntValue()
		switch v {
		case 100, 200, 300, 400, 500, 600, 700, 800, 900:
			return v, nil
		}
	}
	return nil, nil
}

// @descriptor()
// @singleKeyword
// Validation for the ``font-stretch`` descriptor.
func fontStretchDescriptor(tokens []Token, _ string) (Descriptor, error) {
	keyword := getSingleKeyword(tokens)
	switch keyword {
	case "ultra-condensed", "extra-condensed", "condensed", "semi-condensed",
		"normal",
		"semi-expanded", "expanded", "extra-expanded", "ultra-expanded":
		return keyword, nil
	default:
		return nil, nil
	}
}

// @descriptor("font-feature-settings")
// ``font-feature-settings`` descriptor validation.
func fontFeatureSettingsDescriptor(tokens []Token, _ string) (Descriptor, error) {
	return fontFeatureSettings(tokens, ""), nil
}

// @descriptor()
// ``font-variant`` descriptor validation.
func fontVariant(tokens []Token, _ string) (Descriptor, error) {
	if len(tokens) == 1 {
		keyword := getKeyword(tokens[0])
		if keyword == "normal" || keyword == "none" || keyword == "inherit" {
			return []namedProperty{}, nil
		}
	}
	var values []namedProperty
	expanded, err := expandFontVariant(tokens)
	if err != nil {
		return nil, err
	}
	for _, subTokens := range expanded {
		prop, err := validateNonShorthand("", "font-variant"+subTokens.name, subTokens.tokens, true)
		if err != nil {
			return nil, nil
		}
		values = append(values, prop)

	}
	return values, nil
}

// Default validator for descriptors.
func validate(baseUrl, name string, tokens []Token) (Descriptor, error) {
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
		decl, ok := descriptor.(Declaration)
		if !ok || decl.Important {
			continue
		}
		tokens := RemoveWhitespace(decl.Value)
		name := string(decl.Name)
		result, err := validate(baseUrl, name, tokens)

		if err != nil {
			log.Printf("Ignored `%s:%s`, %s. \n", name, Serialize(decl.Value), err)
			continue
		}

		out = append(out, NamedDescriptor{
			Name:       strings.ReplaceAll(name, "-", "_"),
			Descriptor: result,
		})

	}
	return out
}
