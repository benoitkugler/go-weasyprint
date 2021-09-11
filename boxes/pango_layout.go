package boxes

import (
	"fmt"
	"strings"

	"github.com/benoitkugler/go-weasyprint/fonts"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/utils"
	"github.com/benoitkugler/textlayout/language"
	"github.com/benoitkugler/textlayout/pango"
)

type PangoLayoutContext interface {
	Fontmap() pango.FontMap
}

// PangoLayout wraps a pango.Layout object
type PangoLayout struct {
	Layout  *pango.Layout
	Style   pr.Properties
	metrics *pango.FontMetrics // optional

	maxWidth pr.MaybeFloat

	Context PangoLayoutContext // will be a *LayoutContext; to avoid circular dependency

	JustificationSpacing pango.Fl
	firstLineDirection   pango.Direction
}

func NewPangoLayout(context PangoLayoutContext, fontSize pango.Fl, style pr.Properties, justificationSpacing pango.Fl, maxWidth pr.MaybeFloat) *PangoLayout {
	var layout PangoLayout

	layout.JustificationSpacing = justificationSpacing
	layout.setup(context, fontSize, style)
	layout.maxWidth = maxWidth

	return &layout
}

func (p *PangoLayout) setup(context PangoLayoutContext, fontSize pango.Fl, style pr.Properties) {
	p.Context = context
	p.Style = style
	p.firstLineDirection = 0
	fontmap := context.Fontmap()
	pc := pango.NewContext(fontmap)
	pc.SetRoundGlyphPositions(false)

	var lang pango.Language
	if flo := style.GetFontLanguageOverride(); flo != "normal" {
		lang = utils.LST_TO_ISO[strings.ToLower(string(flo))]
	} else if lg := style.GetLang().String; lg != "" {
		lang = language.NewLanguage(lg)
	} else {
		lang = pango.DefaultLanguage()
	}
	pc.SetLanguage(lang)

	fontDesc := pango.NewFontDescription()
	fontDesc.SetFamily(strings.Join(style.GetFontFamily(), ","))

	sty, _ := pango.StyleMap.FromString(string(style.GetFontStyle()))
	fontDesc.SetStyle(pango.Style(sty))

	str, _ := pango.StretchMap.FromString(string(style.GetFontStretch()))
	fontDesc.SetStretch(pango.Stretch(str))

	fontDesc.SetWeight(pango.Weight(style.GetFontWeight().Int))

	fontDesc.SetAbsoluteSize(utils.PangoUnitsFromFloat(fontSize))

	if !style.GetTextDecorationLine().IsNone() {
		metrics := pc.GetMetrics(&fontDesc, lang)
		p.metrics = &metrics
	} else {
		p.metrics = nil
	}

	features := getFontFeatures(style)
	var chunks []string
	for k, v := range features {
		chunks = append(chunks, fmt.Sprintf("%s=%d", k, v))
	}
	featuresString := strings.Join(chunks, ",")
	attr := pango.NewAttrFontFeatures(featuresString)

	p.Layout = pango.NewLayout(pc)
	p.Layout.SetFontDescription(&fontDesc)
	p.Layout.SetAttributes(pango.AttrList{attr})
}

func (p *PangoLayout) SetText(text string, justify bool) {
	if index := strings.IndexByte(text, '\n'); index != -1 && len(text) >= index+2 {
		// Keep only the first line plus one character, we don't need more
		text = text[:index+2]
	}

	p.Layout.SetText(text)

	wordSpacing := float32(p.Style.GetWordSpacing().Value)
	if justify {
		// Justification is needed when drawing text but is useless during
		// layout. Ignore it before layout is reactivated before the drawing
		// step.
		wordSpacing += p.JustificationSpacing
	}

	var letterSpacing float32
	if ls := p.Style.GetLetterSpacing(); ls.String != "normal" {
		letterSpacing = float32(ls.Value)
	}

	var (
		attrList                          pango.AttrList
		letterSpacingInt, spaceSpacingInt int32
	)
	if text != "" && (wordSpacing != 0 || letterSpacing != 0) {
		letterSpacingInt = utils.PangoUnitsFromFloat(letterSpacing)
		spaceSpacingInt = utils.PangoUnitsFromFloat(wordSpacing) + letterSpacingInt
		attrList = p.Layout.Attributes
	}

	addAttr := func(start, end int, spacing int32) {
		attr := pango.NewAttrLetterSpacing(spacing)
		attr.StartIndex, attr.EndIndex = start, end
		attrList.Change(attr)
	}

	textRunes := p.Layout.Text
	addAttr(0, len(textRunes), letterSpacingInt)
	for position, c := range textRunes {
		if c == ' ' {
			addAttr(position, position+1, spaceSpacingInt)
		}
	}

	p.Layout.SetAttributes(attrList)

	// Tabs width
	if strings.ContainsRune(text, '\t') {
		p.setTabs()
	}
}

func (p *PangoLayout) setTabs() {
	tabSize := p.Style.GetTabSize()
	width := int(tabSize.Value)
	if tabSize.Unit == 0 { // no unit, means a multiple of the advance width of the space character
		layout := NewPangoLayout(p.Context, float32(p.Style.GetFontSize().Value), p.Style, p.JustificationSpacing, nil)
		layout.SetText(strings.Repeat(" ", int(tabSize.Value)), false)
		line, _ := layout.GetFirstLine()
		widthTmp, _ := LineSize(line, p.Style)
		width = int(widthTmp + 0.5)
	}
	// 0 is not handled correctly by Pango
	if width == 0 {
		width = 1
	}
	tabs := &pango.TabArray{Tabs: []pango.Tab{{Alignment: pango.TAB_LEFT, Location: width}}}
	p.Layout.SetTabs(tabs)
}

func (p *PangoLayout) Deactivate() {
	// FIXME: à implémenter
}

// GetFirstLine returns the first line and the index of the second line, or -1.
func (p *PangoLayout) GetFirstLine() (*pango.LayoutLine, int) {
	firstLine := p.Layout.GetLine(0)
	secondLine := p.Layout.GetLine(1)
	index := -1
	if secondLine != nil {
		index = secondLine.StartIndex
	}

	p.firstLineDirection = firstLine.ResolvedDir

	return firstLine, index
}

type TextLayoutContext struct {
	FontConfig    *fonts.FontConfiguration
	fontmap       pango.FontMap
	EnableHinting bool
	// FontFeatures  map[string]int
}

// LineSize gets the logical width and height of the given `line`.
// `style` is used to add letter spacing (if needed).
func LineSize(line *pango.LayoutLine, style pr.Properties) (float32, float32) {
	var logicalExtents pango.Rectangle
	line.GetExtents(nil, &logicalExtents)
	width := utils.PangoUnitsToFloat(logicalExtents.Width)
	height := utils.PangoUnitsToFloat(logicalExtents.Height)
	if ls := style.GetLetterSpacing(); ls.String != "normal" {
		width += float32(ls.Value)
	}
	return width, height
}

func defaultFontFeature(f string) string {
	if f == "" {
		return "normal"
	}
	return f
}

// Get the font features from the different properties in style.
// See https://www.w3.org/TR/css-fonts-3/#feature-precedence
// default value is "normal"
// pass nil for default ("normal") on fontFeatureSettings
func getFontFeatures(style pr.Properties) map[string]int {
	fontKerning := defaultFontFeature(string(style.GetFontKerning()))
	fontVariantPosition := defaultFontFeature(string(style.GetFontVariantPosition()))
	fontVariantCaps := defaultFontFeature(string(style.GetFontVariantCaps()))
	fontVariantAlternates := defaultFontFeature(string(style.GetFontVariantAlternates()))

	features := map[string]int{}
	ligatureKeys := map[string][]string{
		"common-ligatures":        {"liga", "clig"},
		"historical-ligatures":    {"hlig"},
		"discretionary-ligatures": {"dlig"},
		"contextual":              {"calt"},
	}
	capsKeys := map[string][]string{
		"small-caps":      {"smcp"},
		"all-small-caps":  {"c2sc", "smcp"},
		"petite-caps":     {"pcap"},
		"all-petite-caps": {"c2pc", "pcap"},
		"unicase":         {"unic"},
		"titling-caps":    {"titl"},
	}
	numericKeys := map[string]string{
		"lining-nums":        "lnum",
		"oldstyle-nums":      "onum",
		"proportional-nums":  "pnum",
		"tabular-nums":       "tnum",
		"diagonal-fractions": "frac",
		"stacked-fractions":  "afrc",
		"ordinal":            "ordn",
		"slashed-zero":       "zero",
	}
	eastAsianKeys := map[string]string{
		"jis78":              "jp78",
		"jis83":              "jp83",
		"jis90":              "jp90",
		"jis04":              "jp04",
		"simplified":         "smpl",
		"traditional":        "trad",
		"full-width":         "fwid",
		"proportional-width": "pwid",
		"ruby":               "ruby",
	}

	// Step 1: getting the default, we rely on Pango for this
	// Step 2: @font-face font-variant, done in fonts.addFontFace
	// Step 3: @font-face font-feature-settings, done in fonts.addFontFace

	// Step 4: font-variant && OpenType features

	if fontKerning != "auto" {
		features["kern"] = 0
		if fontKerning == "normal" {
			features["kern"] = 1
		}
	}

	fontVariantLigatures := style.GetFontVariantLigatures()
	if fontVariantLigatures.String == "none" {
		for _, keys := range ligatureKeys {
			for _, key := range keys {
				features[key] = 0
			}
		}
	} else if fontVariantLigatures.String != "normal" {
		for _, ligatureType := range fontVariantLigatures.Strings {
			value := 1
			if strings.HasPrefix(ligatureType, "no-") {
				value = 0
				ligatureType = ligatureType[3:]
			}
			for _, key := range ligatureKeys[ligatureType] {
				features[key] = value
			}
		}
	}

	if fontVariantPosition == "sub" {
		// TODO: the specification asks for additional checks
		// https://www.w3.org/TR/css-fonts-3/#font-variant-position-prop
		features["subs"] = 1
	} else if fontVariantPosition == "super" {
		features["sups"] = 1
	}

	if fontVariantCaps != "normal" {
		// TODO: the specification asks for additional checks
		// https://www.w3.org/TR/css-fonts-3/#font-variant-caps-prop
		for _, key := range capsKeys[fontVariantCaps] {
			features[key] = 1
		}
	}

	if fv := style.GetFontVariantNumeric(); fv.String != "normal" {
		for _, key := range fv.Strings {
			features[numericKeys[key]] = 1
		}
	}

	if fontVariantAlternates != "normal" {
		// TODO: support other values
		// See https://www.w3.org/TR/css-fonts-3/#font-variant-caps-prop
		if fontVariantAlternates == "historical-forms" {
			features["hist"] = 1
		}
	}

	if fv := style.GetFontVariantEastAsian(); fv.String != "normal" {
		for _, key := range fv.Strings {
			features[eastAsianKeys[key]] = 1
		}
	}

	// Step 5: incompatible non-OpenType features, already handled by Pango

	// Step 6: font-feature-settings
	for _, pair := range style.GetFontFeatureSettings().Values {
		features[pair.String] = pair.Int
	}

	return features
}
