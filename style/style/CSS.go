package style

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/benoitkugler/go-weasyprint/style/parser"

	"github.com/benoitkugler/go-weasyprint/fonts"

	cascadia "github.com/benoitkugler/cascadia2"
	"github.com/benoitkugler/go-weasyprint/style/validation"
	"github.com/benoitkugler/go-weasyprint/utils"
	"golang.org/x/net/html"
)

var HTML5_UA_STYLESHEET, HTML5_PH_STYLESHEET CSS

// LoadStyleSheet should be called once to load stylesheets ressources.
// `path` is the folder containing the 'ressources' directory.
// It will panic on failure.
func LoadStyleSheet(path string) {
	var err error
	HTML5_UA_STYLESHEET, err = newCSS(CssFilename(filepath.Join(path, "ressources", "html5_ua.css")))
	if err != nil {
		panic(err)
	}
	HTML5_PH_STYLESHEET, err = newCSS(CssFilename(filepath.Join(path, "ressources", "html5_ph.css")))
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
	matcher   *matcher
	pageRules *[]pageRule
	baseUrl   string
	fonts     *[]string
}

// checkMimeType = false
func NewCSS(input cssInput, baseUrl string,
	urlFetcher utils.UrlFetcher, checkMimeType bool,
	mediaType string, fontConfig *fonts.FontConfiguration, matcher *matcher,
	pageRules *[]pageRule) (CSS, error) {

	log.Printf("Step 2 - Fetching and parsing CSS - %s", input)

	if urlFetcher == nil {
		urlFetcher = utils.DefaultUrlFetcher
	}
	if mediaType == "" {
		mediaType = "print"
	}
	ressource, err := selectSource(input, baseUrl, urlFetcher, checkMimeType)
	if err != nil {
		return CSS{}, fmt.Errorf("error fetching css input : %s", err)
	}
	defer ressource.content.Close()
	content, err := ioutil.ReadAll(ressource.content)
	if err != nil {
		return CSS{}, fmt.Errorf("cannot read ressource : %s", err)
	}
	stylesheet := parser.ParseStylesheet2(content, false, false)

	if matcher == nil {
		matcher = NewMatcher()
	}
	if pageRules == nil {
		pageRules = &[]pageRule{}
	}
	out := CSS{
		baseUrl:   baseUrl,
		matcher:   matcher,
		pageRules: pageRules,
		fonts:     &[]string{},
	}
	// TODO: fonts are stored here and should be cleaned after rendering

	preprocessStylesheet(mediaType, baseUrl, stylesheet, urlFetcher, matcher,
		out.pageRules, out.fonts, fontConfig, false)
	return out, nil
}

func newCSS(input cssInput) (CSS, error) {
	return NewCSS(input, "", nil, false, "", nil, nil, nil)
}

type cssInput interface {
	isCssInput()
	String() string
}

type CssFilename string
type CssUrl string
type CssString string
type CssReader struct {
	io.ReadCloser
}

func (c CssFilename) isCssInput() {}
func (c CssUrl) isCssInput()      {}
func (c CssString) isCssInput()   {}
func (c CssReader) isCssInput()   {}
func (c CssFilename) String() string {
	return string(c)
}
func (c CssUrl) String() string {
	return string(c)
}
func (c CssString) String() string {
	return string(c)
}
func (c CssReader) String() string {
	return fmt.Sprintf("reader at %p", c.ReadCloser)
}

type source struct {
	content io.ReadCloser
	baseUrl string
}

// Check that only one input is not None, and return it with the
// normalized ``baseUrl``.
// checkCssMimeType=false
// source may have nil content
func selectSource(input cssInput, baseUrl string, urlFetcher utils.UrlFetcher,
	checkCssMimeType bool) (out source, err error) {

	if baseUrl != "" {
		baseUrl, err = utils.EnsureUrl(baseUrl)
		if err != nil {
			return
		}
	}
	switch data := input.(type) {
	case CssFilename:
		if baseUrl == "" {
			baseUrl, err = utils.Path2url(string(data))
			if err != nil {
				return
			}
		}
		f, err := os.Open(string(data))
		if err != nil {
			return source{}, err
		}
		return source{content: f, baseUrl: baseUrl}, nil
	case CssUrl:
		result, err := urlFetcher(string(data))
		if err != nil {
			return source{}, err
		}
		if result.RedirectedUrl == "" {
			result.RedirectedUrl = string(data)
		}
		if checkCssMimeType && result.MimeType != "text/css" {
			log.Printf("Unsupported stylesheet type %s for %s",
				result.MimeType, result.RedirectedUrl)
			return source{baseUrl: baseUrl}, nil
		} else {
			if baseUrl == "" {
				baseUrl = result.RedirectedUrl
			}
			return source{content: result.Content, baseUrl: baseUrl}, nil
		}

	case CssReader:
		return source{content: data.ReadCloser, baseUrl: baseUrl}, nil
	case CssString:
		return source{content: utils.NewBytesCloser(string(data)), baseUrl: baseUrl}, nil
	default:
		return source{}, errors.New("unexpected css input")
	}
}

type match struct {
	selector     cascadia.Selector
	declarations []validation.ValidatedProperty
}

type matcher []match

func NewMatcher() *matcher {
	return &matcher{}
}

type matchResult struct {
	specificity cascadia.Specificity
	pseudoType  string
	payload     []validation.ValidatedProperty
}

func (m matcher) Match(element *html.Node) (out []matchResult) {
	for _, mat := range m {
		for _, det := range mat.selector.MatchDetails(element) {
			out = append(out, matchResult{specificity: det.Specificity, pseudoType: det.PseudoElement, payload: mat.declarations})
		}
	}
	return
}
