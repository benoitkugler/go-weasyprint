package style

import (
	"github.com/andybalholm/cascadia"
	"github.com/benoitkugler/go-weasyprint/css/validation"
	"golang.org/x/net/html"
)

var HTML5_UA_STYLESHEET, HTML5_PH_STYLESHEET CSS

type CSS struct {
	matcher    matcher
	pageRules  []pageRule
	urlFetcher func(string) interface{}
}

func NewCSS(content string) (CSS, error) {
	return CSS{}, nil
}

type match struct {
	selector     cascadia.Selector
	declarations []validation.ValidatedProperty
}

type matcher []match

type matchResult struct {
	specificity [3]uint8
	order       int
	pseudoType  string
	payload     []validation.ValidatedProperty
}

func (m matcher) Match(element *html.Node) []matchResult {
	return nil
}
