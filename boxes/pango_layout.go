package boxes

import pr "github.com/benoitkugler/go-weasyprint/style/properties"

type PangoLayout struct {
	Text                 string
	JustificationSpacing float32
	Context              interface{} // to prevent import cycle
	Style                pr.Properties
}

func (p *PangoLayout) Deactivate() {
	// FIXME: à implémenter
}
