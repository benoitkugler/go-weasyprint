package style

import (
	"github.com/andybalholm/cascadia"
	"github.com/benoitkugler/go-weasyprint/css/validation"
	"golang.org/x/net/html"
)

var HTML5_UA_STYLESHEET, HTML5_PH_STYLESHEET CSS

type CSS struct {
	matcher matcher
}

func NewCSS() (CSS, error) {
	return CSS{}, nil
}

type match struct {
	selector     cascadia.Selector
	declarations []validation.ValidatedProperty
}

type matcher []match

type matchResult struct {
	specificity [3]int
	order       int
	pseudoType  string
	payload     []validation.ValidatedProperty
}

func (m matcher) Match(element *html.Node) []matchResult {
	return nil
}
