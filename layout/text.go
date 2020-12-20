package layout

import (
	"strings"

	"github.com/benoitkugler/go-weasyprint/backend"
	bo "github.com/benoitkugler/go-weasyprint/boxes"
	"github.com/benoitkugler/go-weasyprint/layout/text"
	"github.com/benoitkugler/go-weasyprint/style/properties"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
)

type Splitted struct {
	// pango Layout with the first line
	Layout *bo.PangoLayout

	// length in UTF-8 bytes of the first line
	Length int

	// the number of UTF-8 bytes to skip for the next line.
	// May be ``None`` if the whole text fits in one line.
	// This may be greater than ``length`` in case of preserved
	// newline characters.
	ResumeAt *int

	// width in pixels of the first line
	Width pr.Float

	// height in pixels of the first line
	Height pr.Float

	// baseline in pixels of the first line
	Baseline pr.Float
}

// Return an opaque Pango layout with default Pango line-breaks.
// :param style: a style dict of computed values
// :param maxWidth: The maximum available width in the same unit as ``style['font_size']``,
// or ``None`` for unlimited width.
func createLayout(text string, style pr.Properties, context *LayoutContext, maxWidth pr.MaybeFloat, justificationSpacing pr.Float) *bo.PangoLayout {
	// FIXME: à implémenter
	return &bo.PangoLayout{}
}

// Fit as much as possible in the available width for one line of text.
// minimum=False
func SplitFirstLine(text []rune, style pr.Properties, context LayoutContext, maxWidth pr.MaybeFloat, justificationSpacing pr.Float,
	minimum bool) Splitted {
	// FIXME: a implémenter
	return Splitted{}
}

// returns nil or [wordStart, wordEnd]
func getNextWordBoundaries(t []rune, lang string) []int {
	if len(t) < 2 {
		return nil
	}
	out := make([]int, 2)
	hasBroken := false
	for i, attr := range text.GetLogAttrs(t) {
		if attr.IsWordEnd() {
			out[1] = i // word end
			hasBroken = true
			break
		}
		if attr.IsWordBoundary() {
			out[0] = i // word start
		}
	}
	if !hasBroken {
		return nil
	}
	return out
}

func CanBreakText(t []rune, lang string) MaybeBool {
	if len(t) < 2 {
		return nil
	}
	logs := text.GetLogAttrs(t)
	for _, l := range logs[1 : len(logs)-1] {
		if l.IsLineBreak() {
			return Bool(true)
		}
	}
	return Bool(false)
}

// Return a tuple of the used value of ``line-height`` and the baseline.
// The baseline is given from the top edge of line height.
func StrutLayout(style pr.Properties, context *LayoutContext) (pr.Float, pr.Float) {
	// FIXME: à implémenter
	return 0.5, 0.5
}

// Return the ratio 1ex/font_size, according to given style.
func ExRatio(style pr.Properties) pr.Float {
	// FIXME: à implémenter
	return .5
}

// Draw the given ``textbox`` line to the cairo ``context``.
func ShowFirstLine(context backend.Drawer, textbox bo.TextBox, textOverflow string) {
	// FIXME: à implémenter
	// pango.pangoLayoutSetSingleParagraphMode(textbox.PangoLayout.Layout, true)

	// if textOverflow == "ellipsis" {
	// 	maxWidth := context.ClipExtents()[2] - float64(textbox.PositionX)
	// 	pango.pangoLayoutSetWidth(textbox.PangoLayout.Layout, unitsFromDouble(maxWidth))
	// 	pango.pangoLayoutSetEllipsize(textbox.PangoLayout.Layout, pango.PANGOELLIPSIZEEND)
	// }

	// firstLine, _ = textbox.PangoLayout.GetFirstLine()
	// context = ffi.cast("cairoT *", context.Pointer)
	// pangocairo.pangoCairoShowLayoutLine(context, firstLine)
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
func getFontFeatures(fontKerning, fontVariantPosition, fontVariantCaps, fontVariantAlternates string,
	fontVariantEastAsian, fontVariantLigatures, fontVariantNumeric properties.SStrings,
	fontFeatureSettings properties.SIntStrings) map[string]int {

	fontKerning = defaultFontFeature(fontKerning)
	fontVariantPosition = defaultFontFeature(fontVariantPosition)
	fontVariantCaps = defaultFontFeature(fontVariantCaps)
	fontVariantAlternates = defaultFontFeature(fontVariantAlternates)

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

	if fontVariantNumeric.String != "normal" {
		for _, key := range fontVariantNumeric.Strings {
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

	if fontVariantEastAsian.String != "normal" {
		for _, key := range fontVariantEastAsian.Strings {
			features[eastAsianKeys[key]] = 1
		}
	}

	// Step 5: incompatible non-OpenType features, already handled by Pango

	// Step 6: font-feature-settings
	for _, pair := range fontFeatureSettings.Values {
		features[pair.String] = pair.Int
	}

	return features
}
