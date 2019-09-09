package validation

import (
	"errors"
	"fmt"
	"strings"

	. "github.com/benoitkugler/go-weasyprint/css"
)

var expanders = map[string]expander{
	"border-color":      expandFourSides,
	"border-style":      expandFourSides,
	"border-width":      expandFourSides,
	"margin":            expandFourSides,
	"padding":           expandFourSides,
	"bleed":             expandFourSides,
	"border-radius":     borderRadius,
	"page-break-after":  expandPageBreakBeforeAfter,
	"page-break-before": expandPageBreakBeforeAfter,
	"page-break-inside": expandPageBreakInside,
	"background":        expandBackground,
	"word-wrap":         expandWordWrap,
	"list-style":        genericExpander("-type", "-position", "-image")(_expandListStyle),
	"border":            expandBorder,
	"border-top":        expandBorderSide,
	"border-right":      expandBorderSide,
	"border-bottom":     expandBorderSide,
	"border-left":       expandBorderSide,
	"column-rule":       expandBorderSide,
	"outline":           expandBorderSide,
	"columns":           genericExpander("column-width", "column-count")(_expandColumns),
	"font-variant": genericExpander("-alternates", "-caps", "-east-asian", "-ligatures",
		"-numeric", "-position")(_fontVariant),
	"font": genericExpander("-style", "-variant-caps", "-weight", "-stretch", "-size",
		"line-height", "-family")(_expandFont),
}

var expandBorderSide = genericExpander("-width", "-color", "-style")(_expandBorderSide)

// Expanders
type namedProperty struct {
	name     string
	property CssProperty
}

type namedTokens struct {
	name   string
	tokens []Token
}

type beforeGeneric = func(baseUrl, name string, tokens []Token) ([]namedTokens, error)

// Decorator helping expanders to handle ``inherit`` && ``initial``.
//     Wrap an expander so that it does not have to handle the "inherit" and
//     "initial" cases, && can just yield name suffixes. Missing suffixes
//     get the initial value.
//
func genericExpander(expandedNames ...string) func(beforeGeneric) expander {
	_expandedNames := Set{}
	for _, name := range expandedNames {
		_expandedNames[name] = Has
	}
	// Decorate the ``wrapped`` expander.
	genericExpanderDecorator := func(wrapped beforeGeneric) expander {

		// Wrap the expander.
		genericExpanderWrapper := func(baseUrl, name string, tokens []Token) (out []namedProperty, err error) {
			keyword := getSingleKeyword(tokens)
			results, toBeValidated := map[string]CssProperty{}, map[string][]Token{}
			var skipValidation bool
			if keyword == "inherit" || keyword == "initial" {
				for _, name := range expandedNames {
					results[name] = String(keyword)
				}
				skipValidation = true
			} else {
				skipValidation = false

				result, err := wrapped(baseUrl, name, tokens)
				if err != nil {
					return nil, err
				}

				for _, nameToken := range result {
					newName, newToken := nameToken.name, nameToken.tokens
					if !_expandedNames.Has(newName) {
						return nil, fmt.Errorf("unknown expanded property %s", newName)
					}
					if _, isIn := toBeValidated[newName]; isIn {
						return nil, fmt.Errorf("got multiple %s values in a %s shorthand",
							strings.Trim(newName, "-"), name)
					}
					toBeValidated[newName] = newToken
				}
			}

			for _, newName := range expandedNames {
				actualNewName := newName
				if strings.HasPrefix(newName, "-") {
					// newName is a suffix
					actualNewName = name + newName
				}
				var (
					value CssProperty
					in    bool
				)
				if skipValidation { // toBeValidated is empty -> ignore it
					value, in = results[newName]
				} else { // results is empty -> ignore it
					tokens, in = toBeValidated[newName]
					if in {
						np, err := validateNonShorthand(baseUrl, actualNewName, tokens, true)
						if err != nil {
							return nil, err
						}
						actualNewName = np.name
						value = np.property
					}
				}
				if !in {
					value = String("initial")
				}

				out = append(out, namedProperty{name: actualNewName, property: value})
			}
			return out, nil
		}
		return genericExpanderWrapper
	}
	return genericExpanderDecorator
}

//@expander("border-color")
//@expander("border-style")
//@expander("border-width")
//@expander("margin")
//@expander("padding")
//@expander("bleed")
// Expand properties setting a token for the four sides of a box.
func expandFourSides(baseUrl, name string, tokens []Token) (out []namedProperty, err error) {
	// Make sure we have 4 tokens
	if len(tokens) == 1 {
		tokens = []Token{tokens[0], tokens[0], tokens[0], tokens[0]}
	} else if len(tokens) == 2 {
		tokens = []Token{tokens[0], tokens[1], tokens[0], tokens[1]} // (bottom, left) defaults to (top, right)
	} else if len(tokens) == 3 {
		tokens = append(tokens, tokens[1]) // left defaults to right
	} else if len(tokens) != 4 {
		return out, fmt.Errorf("Expected 1 to 4 token components got %d", len(tokens))
	}
	var newName string
	for index, suffix := range [4]string{"-top", "-right", "-bottom", "-left"} {
		token := tokens[index]
		i := strings.LastIndex(name, "-")
		if i == -1 {
			newName = name + suffix
		} else {
			// eg. border-color becomes border-*-color, not border-color-*
			newName = name[:i] + suffix + name[i:]
		}
		prop, err := validateNonShorthand(baseUrl, newName, []Token{token}, true)
		if err != nil {
			return out, err
		}
		out = append(out, prop)
	}
	return out, nil
}

//@expander("border-radius")
// Validator for the `border-radius` property.
func borderRadius(baseUrl, name string, tokens []Token) (out []namedProperty, err error) {
	var horizontal, vertical []Token
	current := &horizontal

	for index, token := range tokens {
		if lit, ok := token.(LiteralToken); ok && lit.Value == "/" {
			if current == &horizontal {
				if index == len(tokens)-1 {
					return nil, errors.New("Expected value after '/' separator")
				} else {
					current = &vertical
				}
			} else {
				return nil, errors.New("Expected only one '/' separator")
			}
		} else {
			*current = append(*current, token)
		}
	}

	if len(vertical) == 0 {
		vertical = append([]Token{}, horizontal...)
	}

	for _, values := range [2]*[]Token{&horizontal, &vertical} {
		// Make sure we have 4 tokens
		if len(*values) == 1 {
			*values = []Token{(*values)[0], (*values)[0], (*values)[0], (*values)[0]}
		} else if len(*values) == 2 {
			*values = []Token{(*values)[0], (*values)[1], (*values)[0], (*values)[1]} // (br, bl) defaults to (tl, tr)
		} else if len(*values) == 3 {
			*values = append(*values, (*values)[1]) // bl defaults to tr
		} else if len(*values) != 4 {
			return nil, fmt.Errorf("Expected 1 to 4 token components got %d", len(*values))
		}
	}
	corners := [4]string{"top-left", "top-right", "bottom-right", "bottom-left"}
	for index, corner := range corners {
		newName := fmt.Sprintf("border-%s-radius", corner)
		ts := []Token{horizontal[index], vertical[index]}
		result, err := validateNonShorthand(baseUrl, newName, ts, true)
		if err != nil {
			return nil, err
		}
		out = append(out, result)
	}
	return out, nil
}

//@expander("list-style")
//@genericExpander("-type", "-position", "-image", wantsBaseUrl=true)
// Expand the ``list-style`` shorthand property.
//     See http://www.w3.org/TR/CSS21/generate.html#propdef-list-style
//
func _expandListStyle(baseUrl, name string, tokens []Token) (out []namedTokens, err error) {
	var typeSpecified, imageSpecified bool
	noneCount := 0
	var noneToken Token
	for _, token := range tokens {
		var suffix string
		if getKeyword(token) == "none" {
			// Can be either -style || -image, see at the end which is not
			// otherwise specified.
			noneCount += 1
			noneToken = token
			continue
		}
		if listStyleType([]Token{token}, "") != nil {
			suffix = "-type"
			typeSpecified = true
		} else if listStylePosition([]Token{token}, "") != nil {
			suffix = "-position"
		} else {
			image, err := imageUrl([]Token{token}, baseUrl)
			if err != nil {
				return nil, err
			}
			if image != nil {
				suffix = "-image"
				imageSpecified = true
			} else {
				return nil, InvalidValue
			}
		}
		out = append(out, namedTokens{name: suffix, tokens: []Token{token}})
	}

	if !typeSpecified && noneCount > 0 {
		out = append(out, namedTokens{name: "-type", tokens: []Token{noneToken}})
		noneCount -= 1
	}

	if !imageSpecified && noneCount > 0 {
		out = append(out, namedTokens{name: "-image", tokens: []Token{noneToken}})
		noneCount -= 1
	}

	if noneCount > 0 {
		// Too many none tokens.
		return nil, InvalidValue
	}
	return out, nil
}

//@expander("border")
// Expand the ``border`` shorthand property.
//     See http://www.w3.org/TR/CSS21/box.html#propdef-border
//
func expandBorder(baseUrl, name string, tokens []Token) (out []namedProperty, err error) {
	for _, suffix := range [4]string{"-top", "-right", "-bottom", "-left"} {
		props, err := expandBorderSide(baseUrl, name+suffix, tokens)
		if err != nil {
			return nil, err
		}
		out = append(out, props...)
	}
	return out, nil
}

//@expander("border-top")
//@expander("border-right")
//@expander("border-bottom")
//@expander("border-left")
//@expander("column-rule")
//@expander("outline")
//@genericExpander("-width", "-color", "-style")
// Expand the ``border-*`` shorthand properties.
//     See http://www.w3.org/TR/CSS21/box.html#propdef-border-top
//
func _expandBorderSide(_, name string, tokens []Token) ([]namedTokens, error) {
	out := make([]namedTokens, len(tokens))
	for index, token := range tokens {
		var suffix string
		if !ParseColor(token).IsNone() {
			suffix = "-color"
		} else if borderWidth([]Token{token}, "") != nil {
			suffix = "-width"
		} else if borderStyle([]Token{token}, "") != nil {
			suffix = "-style"
		} else {
			return nil, InvalidValue
		}
		out[index] = namedTokens{name: suffix, tokens: []Token{token}}
	}
	return out, nil
}

type backgroundProps struct {
	color      CssProperty
	image      Image
	repeat     [2]string
	attachment string
	position   Center
	size       Size
	clip       string
	origin     string
	_keys      Set
}

func (b backgroundProps) add(name string) error {
	name = "background_" + name
	if b._keys.Has(name) {
		return InvalidValue
	}
	return nil
}

//@expander("background")
// Expand the ``background`` shorthand property.
//     See http://dev.w3.org/csswg/css3-background/#the-background
//
func expandBackground(baseUrl, name string, tokens []Token) (out []namedProperty, err error) {
	properties := [8]string{
		"background_color", "background_image", "background_repeat",
		"background_attachment", "background_position", "background_size",
		"background_clip", "background_origin"}
	keyword := getSingleKeyword(tokens)
	if keyword == "initial" || keyword == "inherit" {
		for _, name := range properties {
			out = append(out, namedProperty{name: name, property: String(keyword)})
		}
		return
	}

	parseLayer := func(tokens []Token, finalLayer bool) (CssProperty, backgroundProps, error) {
		results := backgroundProps{_keys: Set{}}

		// Make `tokens` a stack
		tokens = reverse(tokens)
		for len(tokens) > 0 {
			repeat := _backgroundRepeat(reverse(tokens[len(tokens)-2:]))
			if repeat != [2]string{} {
				if err = results.add("repeat"); err != nil {
					return Color{}, backgroundProps{}, err
				}
				results.repeat = repeat
				tokens = tokens[:len(tokens)-2]
				continue
			}

			token := tokens[len(tokens)-1:]

			if finalLayer {
				color := otherColors(token, "")
				if color != nil {
					if err = results.add("color"); err != nil {
						return Color{}, backgroundProps{}, err
					}
					results.color = color
					tokens = tokens[:len(tokens)-1]
					continue
				}
			}

			image, err := _backgroundImage(token, baseUrl)
			if err != nil {
				return Color{}, backgroundProps{}, err
			}
			if image != nil {
				if err = results.add("image"); err != nil {
					return Color{}, backgroundProps{}, err
				}
				results.image = image
				tokens = tokens[:len(tokens)-1]
				continue
			}

			repeat = _backgroundRepeat(token)
			if repeat != [2]string{} {
				if err = results.add("repeat"); err != nil {
					return Color{}, backgroundProps{}, err
				}
				results.repeat = repeat
				tokens = tokens[:len(tokens)-1]
				continue
			}

			attachment := _backgroundAttachment(token)
			if attachment != "" {
				if err = results.add("attachment"); err != nil {
					return Color{}, backgroundProps{}, err
				}
				results.attachment = attachment
				tokens = tokens[:len(tokens)-1]
				continue
			}

			index := 4 - len(tokens)
			if index < 0 {
				index = 0
			}
			var position Center
			for _, n := range []int{4, 3, 2, 1}[index:] {
				nTokens := reverse(tokens[len(tokens)-n:])
				position = _backgroundPosition(nTokens)
				if !position.IsNone() {
					if err = results.add("position"); err != nil {
						return Color{}, backgroundProps{}, err
					}
					results.position = position
					tokens = tokens[:len(tokens)-n]
					if len(tokens) > 0 {
						if lit, ok := tokens[len(tokens)-1].(LiteralToken); ok && lit.Value == "/" {
							index := 2 - len(tokens)
							if index < 0 {
								index = 0
							}
							for _, n := range []int{3, 2}[index:] {
								// n includes the "/" delimiter.
								nTokens = reverse(tokens[len(tokens)-n : len(tokens)-1])
								size := _backgroundSize(nTokens)
								if !size.IsNone() {
									if err = results.add("size"); err != nil {
										return Color{}, backgroundProps{}, err
									}
									results.size = size
									tokens = tokens[:len(tokens)-n]
								}
							}
						}
					}
					break
				}
			}
			if !position.IsNone() {
				continue
			}

			origin := _box(tokens)
			if origin != "" {
				if err = results.add("origin"); err != nil {
					return Color{}, backgroundProps{}, err
				}
				results.origin = origin
				tokens = tokens[:len(tokens)-1]
				nextToken := tokens[len(tokens)-1:]

				clip := _box(nextToken)
				if clip != "" {
					if err = results.add("clip"); err != nil {
						return Color{}, backgroundProps{}, err
					}
					results.clip = clip
					tokens = tokens[:len(tokens)-1]
				} else {
					// The same keyword sets both:
					clip := _box(nextToken)
					if clip != "" {
						return Color{}, backgroundProps{}, errors.New("clip shoudn't be empty")
					}
					if err = results.add("clip"); err != nil {
						return Color{}, backgroundProps{}, err
					}
					results.clip = clip
				}
				continue
			}
			return Color{}, backgroundProps{}, InvalidValue
		}

		var color CssProperty = InitialValues.GetBackgroundColor()
		if results._keys.Has("background_color") {
			color = results.color
			delete(results._keys, "background_color")
		}

		if !results._keys.Has("background_image") {
			results.image = InitialValues.GetBackgroundImage()[0]
		}
		if !results._keys.Has("background_repeat") {
			results.repeat = InitialValues.GetBackgroundRepeat()[0]
		}
		if !results._keys.Has("background_attachment") {
			results.attachment = InitialValues.GetBackgroundAttachment()[0]
		}
		if !results._keys.Has("background_position") {
			results.position = InitialValues.GetBackgroundPosition()[0]
		}
		if !results._keys.Has("background_size") {
			results.size = InitialValues.GetBackgroundSize()[0]
		}
		if !results._keys.Has("background_clip") {
			results.clip = InitialValues.GetBackgroundClip()[0]
		}
		if !results._keys.Has("background_origin") {
			results.origin = InitialValues.GetBackgroundOrigin()[0]
		}
		return color, results, nil
	}

	_layers := splitOnComma(tokens)
	n := len(_layers)
	layers := make([][]Token, n)
	for i := range _layers {
		layers[n-1-i] = _layers[i]
	}

	var result_color CssProperty

	var results_images Images
	var results_repeats Repeats
	var results_attachments Strings
	var results_positions Centers
	var results_sizes Sizes
	var results_clips Strings
	var results_origins Strings

	for i, tokens := range layers {
		layerColor, layer, err := parseLayer(tokens, i == 0)
		if i == 0 {
			result_color = layerColor
		}
		if err != nil {
			return nil, err
		}
		results_images = append(results_images, layer.image)
		results_repeats = append(results_repeats, layer.repeat)
		results_attachments = append(results_attachments, layer.attachment)
		results_positions = append(results_positions, layer.position)
		results_sizes = append(results_sizes, layer.size)
		results_clips = append(results_clips, layer.clip)
		results_origins = append(results_origins, layer.origin)
	}

	// un-reverse
	n = len(layers)
	var rev_images = make(Images, n)
	var rev_repeats = make(Repeats, n)
	var rev_attachments = make(Strings, n)
	var rev_positions = make(Centers, n)
	var rev_sizes = make(Sizes, n)
	var rev_clips = make(Strings, n)
	var rev_origins = make(Strings, n)
	for i := range layers {
		rev_images[n-1-i] = results_images[i]
		rev_repeats[n-1-i] = results_repeats[i]
		rev_attachments[n-1-i] = results_attachments[i]
		rev_positions[n-1-i] = results_positions[i]
		rev_sizes[n-1-i] = results_sizes[i]
		rev_clips[n-1-i] = results_clips[i]
		rev_origins[n-1-i] = results_origins[i]
	}

	out = []namedProperty{
		{name: "background_image", property: rev_images},
		{name: "background_repeat", property: rev_repeats},
		{name: "background_attachment", property: rev_attachments},
		{name: "background_position", property: rev_positions},
		{name: "background_size", property: rev_sizes},
		{name: "background_clip", property: rev_clips},
		{name: "background_origin", property: rev_origins},
		{name: "background-color", property: result_color},
	}
	return out, nil
}

//@expander("page-break-after")
//@expander("page-break-before")
// Expand legacy ``page-break-before`` && ``page-break-after`` properties.
//     See https://www.w3.org/TR/css-break-3/#page-break-properties
//
func expandPageBreakBeforeAfter(baseUrl, name string, tokens []Token) (out []namedProperty, err error) {
	keyword := getSingleKeyword(tokens)
	splits := strings.SplitN(name, "-", 1)
	if len(splits) < 2 {
		return nil, fmt.Errorf("bad format for name %s : should contain '-' ", name)
	}
	newName := splits[1]
	if keyword == "auto" || keyword == "left" || keyword == "right" || keyword == "avoid" {
		out = append(out, namedProperty{name: newName, property: String(keyword)})
	} else if keyword == "always" {
		out = append(out, namedProperty{name: newName, property: String("page")})
	}
	return out, nil
}

//@expander("page-break-inside")
// Expand the legacy ``page-break-inside`` property.
//     See https://www.w3.org/TR/css-break-3/#page-break-properties
//
func expandPageBreakInside(baseUrl, name string, tokens []Token) (out []namedProperty, err error) {
	keyword := getSingleKeyword(tokens)
	if keyword == "auto" || keyword == "avoid" {
		out = append(out, namedProperty{name: "break-inside", property: String(keyword)})
	}
	return out, nil
}

//@expander("columns")
//@genericExpander("column-width", "column-count")
// Expand the ``columns`` shorthand property.
func _expandColumns(_, name string, tokens []Token) (out []namedTokens, err error) {
	if len(tokens) == 2 && getKeyword(tokens[0]) == "auto" {
		tokens = reverse(tokens)
	}
	name = ""
	for _, token := range tokens {
		l := []Token{token}
		if columnWidth(l, "") != nil && name != "column-width" {
			name = "column-width"
		} else if columnCount(l, "") != nil {
			name = "column-count"
		} else {
			return nil, InvalidValue
		}
		out = append(out, namedTokens{name: name, tokens: l})
	}
	return out, nil
}

var (
	noneFakeToken   = IdentToken{Value: "none"}
	normalFakeToken = IdentToken{Value: "normal"}
)

//@expander("font-variant")
//@genericExpander("-alternates", "-caps", "-east-asian", "-ligatures",
//   "-numeric", "-position")
// Expand the ``font-variant`` shorthand property.
//     https://www.w3.org/TR/css-fonts-3/#font-variant-prop
//
func _fontVariant(_, name string, tokens []Token) (out []namedTokens, err error) {
	return expandFontVariant(tokens)
}

func expandFontVariant(tokens []Token) (out []namedTokens, err error) {
	keyword := getSingleKeyword(tokens)
	if keyword == "normal" || keyword == "none" {
		out = make([]namedTokens, 6)
		for index, suffix := range [5]string{"-alternates", "-caps", "-east-asian", "-numeric",
			"-position"} {
			out[index] = namedTokens{name: suffix, tokens: []Token{normalFakeToken}}
		}
		token := noneFakeToken
		if keyword == "normal" {
			token = normalFakeToken
		}
		out[6] = namedTokens{name: "-ligatures", tokens: []Token{token}}
	} else {
		features := map[string][]Token{
			"alternates": nil,
			"caps":       nil,
			"east-asian": nil,
			"ligatures":  nil,
			"numeric":    nil,
			"position":   nil,
		}
		for _, token := range tokens {
			keyword := getKeyword(token)
			if keyword == "normal" {
				// We don"t allow "normal", only the specific values
				return nil, errors.New("normal not allowed")
			}
			found := false
			for feature := range features {
				if fontVariantMapper[feature]([]Token{token}, "") != nil {
					features[feature] = append(features[feature], token)
					found = true
					break
				}
			}
			if !found {
				return nil, errors.New("font variant not supported")
			}
		}
		var out []namedTokens
		for feature, tokens := range features {
			if len(tokens) > 0 {
				out = append(out, namedTokens{name: fmt.Sprintf("-%s", feature), tokens: tokens})
			}
		}
	}
	return out, nil
}

var fontVariantMapper = map[string]func(tokens []Token, _ string) CssProperty{
	"alternates": fontVariantAlternates,
	"caps":       fontVariantCaps,
	"east-asian": fontVariantEastAsian,
	"ligatures":  fontVariantLigatures,
	"numeric":    fontVariantNumeric,
	"position":   fontVariantPosition,
}

//@expander("font")
//@genericExpander("-style", "-variant-caps", "-weight", "-stretch", "-size",
//   "line-height", "-family")  // line-height is not a suffix
// Expand the ``font`` shorthand property.
//     https://www.w3.org/TR/css-fonts-3/#font-prop
//
func _expandFont(_, name string, tokens []Token) ([]namedTokens, error) {
	expandFontKeyword := getSingleKeyword(tokens)
	if expandFontKeyword == "caption" || expandFontKeyword == "icon" || expandFontKeyword == "menu" || expandFontKeyword == "message-box" || expandFontKeyword ==
		"small-caption" || expandFontKeyword == "status-bar" {

		return nil, errors.New("System fonts are not supported")
	}
	var (
		out   []namedTokens
		token Token
	)
	// Make `tokens` a stack
	tokens = reverse(tokens)
	// Values for font-style font-variant and font-weight can come in any
	// order and are all optional.
	for len(tokens) > 0 {
		token = tokens[len(tokens)-1]
		tokens = tokens[:len(tokens)-1]
		kw := getKeyword(token)
		if kw == "normal" {
			// Just ignore "normal" keywords. Unspecified properties will get
			// their initial token, which is "normal" for all three here.
			// TODO: fail if there is too many "normal" values?
			continue
		}

		var suffix string
		if fontStyle([]Token{token}, "") != nil {
			suffix = "-style"
		} else if kw == "normal" || kw == "small-caps" {
			suffix = "-variant-caps"
		} else if fontWeight([]Token{token}, "") != nil {
			suffix = "-weight"
		} else if fontStretch([]Token{token}, "") != nil {
			suffix = "-stretch"
		} else {
			// We’re done with these three, continue with font-size
			break
		}
		out = append(out, namedTokens{name: suffix, tokens: []Token{token}})
	}

	// Then font-size is mandatory
	// Latest `token` from the loop.
	fs, err := fontSize([]Token{token}, "")
	if err != nil {
		return nil, err
	}
	if fs == nil {
		return nil, errors.New("font-size is mandatory for short font attribute !")
	}
	out = append(out, namedTokens{name: "-size", tokens: []Token{token}})

	// Then line-height is optional, but font-family is not so the list
	// must not be empty yet
	if len(tokens) == 0 {
		return nil, errors.New("font-familly is mandatory for short font attribute !")
	}

	token = tokens[len(tokens)-1]
	tokens = tokens[:len(tokens)-1]
	if lit, ok := token.(LiteralToken); ok && lit.Value == "/" {
		token = tokens[len(tokens)-1]
		tokens = tokens[:len(tokens)-1]
		if lineHeight([]Token{token}, "") == nil {
			return nil, InvalidValue
		}
		out = append(out, namedTokens{name: "line-height", tokens: []Token{token}})
	} else {
		// We pop()ed a font-family, add it back
		tokens = append(tokens, token)
	}

	// Reverse the stack to get normal list
	tokens = reverse(tokens)
	if fontFamily(tokens, "") == nil {
		return nil, InvalidValue
	}
	out = append(out, namedTokens{name: "-family", tokens: tokens})
	return out, nil
}

//@expander("word-wrap")
// Expand the ``word-wrap`` legacy property.
//     See http://http://www.w3.org/TR/css3-text/#overflow-wrap
//
func expandWordWrap(baseUrl, name string, tokens []Token) ([]namedProperty, error) {
	keyword := overflowWrap(tokens, "")
	if keyword == nil {
		return nil, InvalidValue
	}
	return []namedProperty{{name: "overflow-wrap", property: keyword}}, nil
}

// Default validator for non-shorthand properties.
// required = false
func validateNonShorthand(baseUrl, name string, tokens []Token, required bool) (out namedProperty, err error) {
	if !required && !KnownProperties.Has(name) {
		hyphensName := strings.ReplaceAll(name, "", "-")
		if KnownProperties.Has(hyphensName) {
			return out, fmt.Errorf("did you mean %s?", hyphensName)
		} else {
			return out, errors.New("unknown property")
		}
	}

	if _, isIn := allValidators[name]; !required && !isIn {
		return out, fmt.Errorf("property %s not supported yet", name)
	}
	var value CssProperty
	keyword := getSingleKeyword(tokens)
	if keyword == "initial" || keyword == "inherit" {
		value = String(keyword)
	} else {
		function := validators[name]
		if function != nil {
			value = function(tokens, baseUrl)
		} else {
			functionE := validatorsError[name]
			if functionE != nil {
				var err error
				value, err = functionE(tokens, baseUrl)
				if err != nil {
					return out, err
				}
			}
		}
		if value == nil {
			return out, errors.New("invalid property (nil function return)")
		}
	}
	return namedProperty{name: name, property: value}, nil
}

func defaultValidateShorthand(baseUrl, name string, tokens []Token) ([]namedProperty, error) {
	np, err := validateNonShorthand(baseUrl, name, tokens, false)
	return []namedProperty{np}, err
}
