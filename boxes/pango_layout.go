package boxes

import (
	"strings"

	"github.com/benoitkugler/go-weasyprint/fonts"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/utils"
	"github.com/benoitkugler/textlayout/language"
	"github.com/benoitkugler/textlayout/pango"
)

// PangoLayout wraps a pango.Layout object
type PangoLayout struct {
	Context  TextLayoutContext
	maxWidth pr.MaybeFloat
	layout   *pango.Layout

	metrics *pango.FontMetrics // optional

	Style                pr.Properties
	Text                 string
	firstLineDirection   int
	JustificationSpacing float32
}

func NewPangoLayout(context TextLayoutContext, fontSize pango.Fl, style pr.Properties, justificationSpacing float32, maxWidth pr.MaybeFloat) *PangoLayout {
	var layout PangoLayout
	layout.JustificationSpacing = justificationSpacing
	layout.setup(context, fontSize, style)
	layout.maxWidth = maxWidth
	return &layout
}

func (p *PangoLayout) setup(context TextLayoutContext, fontSize pango.Fl, style pr.Properties) {
	p.Context = context
	p.Style = style
	p.firstLineDirection = 0
	fontmap := context.fontmap
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

	fontDesc.SetAbsoluteSize(utils.PangoUnitsFromDouble(fontSize))

	p.layout = pango.NewLayout(pc)
	p.layout.SetFontDescription(&fontDesc)

	if !style.GetTextDecorationLine().IsNone() {
		metrics := pc.GetMetrics(&fontDesc, lang)
		p.metrics = &metrics
	} else {
		p.metrics = nil
	}

	// FIXME:
}

func (p *PangoLayout) Deactivate() {
	// FIXME: à implémenter
}

type TextLayoutContext struct {
	EnableHinting bool
	FontConfig    *fonts.FontConfiguration
	FontFeatures  map[string]int
	fontmap       pango.FontMap
}
