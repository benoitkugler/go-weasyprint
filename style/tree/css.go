package tree

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/benoitkugler/go-weasyprint/boxes/counters"
	"github.com/benoitkugler/go-weasyprint/layout/text"
	"github.com/benoitkugler/go-weasyprint/logger"

	"github.com/benoitkugler/go-weasyprint/style/parser"

	"github.com/benoitkugler/cascadia"
	"github.com/benoitkugler/go-weasyprint/style/validation"
	"github.com/benoitkugler/go-weasyprint/utils"
	"golang.org/x/net/html"

	_ "embed"
)

var html5UAStylesheet, html5PHStylesheet CSS

// TestUAStylesheet is a lightweight style sheet
var TestUAStylesheet CSS

//go:embed tests_ua.css
var testUAStylesheet string

func init() {
	var err error
	TestUAStylesheet, err = NewCSSDefault(utils.InputString(testUAStylesheet))
	if err != nil {
		log.Fatal(err)
	}
}

// LoadStyleSheet should be called once to load stylesheets ressources.
// `path` is the folder containing the 'ressources' directory.
// It will panic on failure.
// TODO: use embed
func LoadStyleSheet(path string) {
	var err error
	html5UAStylesheet, err = NewCSSDefault(utils.InputFilename(filepath.Join(path, "ressources", "html5_ua.css")))
	if err != nil {
		panic(err)
	}
	html5PHStylesheet, err = NewCSSDefault(utils.InputFilename(filepath.Join(path, "ressources", "html5_ph.css")))
	if err != nil {
		panic(err)
	}
}

// CSS represents a parsed CSS stylesheet.
// An instance is created in the same way as `HTML`, except that
// the ``tree`` argument is not available. All other arguments are the same.
// An additional argument called ``font_config`` must be provided to handle
// ``@font-config`` rules. The same ``fonts.FontConfiguration`` object must be
// used for different ``CSS`` objects applied to the same document.
// ``CSS`` objects have no public attribute or method. They are only meant to
// be used in the `HTML.WritePdf`, `HTML.WritePng` and
// `HTML.Render` methods of `HTML` objects.
type CSS struct {
	Matcher   matcher
	pageRules []PageRule
	baseUrl   string
	fonts     []string
}

// checkMimeType = false
func NewCSS(input utils.ContentInput, baseUrl string,
	urlFetcher utils.UrlFetcher, checkMimeType bool,
	mediaType string, fontConfig *text.FontConfiguration, matcher *matcher,
	pageRules *[]PageRule, counterStyle counters.CounterStyle) (CSS, error) {

	logger.ProgressLogger.Printf("Step 2 - Fetching and parsing CSS - %s", input)

	if urlFetcher == nil {
		urlFetcher = utils.DefaultUrlFetcher
	}
	if mediaType == "" {
		mediaType = "print"
	}

	ressource, err := utils.SelectSource(input, baseUrl, urlFetcher, checkMimeType)
	if err != nil {
		return CSS{}, fmt.Errorf("error fetching css input : %s", err)
	}

	stylesheet := parser.ParseStylesheet2(ressource.Content, false, false)

	if matcher == nil {
		matcher = NewMatcher()
	}
	if pageRules == nil {
		pageRules = &[]PageRule{}
	}
	fts := &[]string{}
	out := CSS{baseUrl: ressource.BaseUrl}
	preprocessStylesheet(mediaType, ressource.BaseUrl, stylesheet, urlFetcher, matcher,
		pageRules, fts, fontConfig, counterStyle, false)
	out.Matcher = *matcher
	out.pageRules = *pageRules
	out.fonts = *fts
	return out, nil
}

func NewCSSDefault(input utils.ContentInput) (CSS, error) {
	return NewCSS(input, "", nil, false, "", nil, nil, nil, nil)
}

func (c CSS) IsNone() bool {
	return c.baseUrl == "" && c.fonts == nil && c.Matcher == nil && c.pageRules == nil
}

type match struct {
	selector     cascadia.SelectorGroup
	declarations []validation.ValidatedProperty
}

type matcher []match

func NewMatcher() *matcher {
	return &matcher{}
}

type matchResult struct {
	pseudoType  string
	payload     []validation.ValidatedProperty
	specificity cascadia.Specificity
}

func (m matcher) Match(element *html.Node) (out []matchResult) {
	for _, mat := range m {
		for _, sel := range mat.selector {
			if sel.Match(element) {
				out = append(out, matchResult{specificity: sel.Specificity(), pseudoType: sel.PseudoElement(), payload: mat.declarations})
			}
		}
	}
	return
}

type pageIndex struct {
	Group []parser.Token // TODO: handle groups
	A, B  int
}

func (p pageIndex) IsNone() bool {
	return p.A == 0 && p.B == 0 && p.Group == nil
}

type pageSelector struct {
	Side        string
	Name        string
	Index       pageIndex
	Specificity cascadia.Specificity
	Blank       bool
	First       bool
}
