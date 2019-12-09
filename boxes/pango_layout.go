package boxes

import (
	"github.com/benoitkugler/go-weasyprint/fonts"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
)

type PangoLayout struct {
	Text                 string
	JustificationSpacing float64
	Context              interface{} // to prevent import cycle
	Style                pr.Properties
}

type TextLayoutContext struct {
	EnableHinting bool
	FontConfig    *fonts.FontConfiguration
	FontFeatures  map[string]int
}

func (p *PangoLayout) Deactivate() {
	// FIXME: à implémenter
}
