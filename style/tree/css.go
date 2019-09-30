package tree

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"golang.org/x/net/html/charset"

	"github.com/benoitkugler/go-weasyprint/logger"

	"github.com/benoitkugler/go-weasyprint/style/parser"

	"github.com/benoitkugler/go-weasyprint/fonts"

	"github.com/benoitkugler/cascadia"
	"github.com/benoitkugler/go-weasyprint/style/validation"
	"github.com/benoitkugler/go-weasyprint/utils"
	"golang.org/x/net/html"
)

var html5UAStylesheet, html5PHStylesheet CSS

// LoadStyleSheet should be called once to load stylesheets ressources.
// `path` is the folder containing the 'ressources' directory.
// It will panic on failure.
func LoadStyleSheet(path string) {
	var err error
	html5UAStylesheet, err = newCSS(InputFilename(filepath.Join(path, "ressources", "html5_ua.css")))
	if err != nil {
		panic(err)
	}
	html5PHStylesheet, err = newCSS(InputFilename(filepath.Join(path, "ressources", "html5_ph.css")))
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
	matcher   matcher
	pageRules []pageRule
	baseUrl   string
	fonts     []string
}

// checkMimeType = false
func NewCSS(input contentInput, baseUrl string,
	urlFetcher utils.UrlFetcher, checkMimeType bool,
	mediaType string, fontConfig *fonts.FontConfiguration, matcher *matcher,
	pageRules *[]pageRule) (CSS, error) {

	logger.ProgressLogger.Printf("Step 2 - Fetching and parsing CSS - %s", input)

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

	stylesheet := parser.ParseStylesheet2(ressource.content, false, false)

	if matcher == nil {
		matcher = NewMatcher()
	}
	if pageRules == nil {
		pageRules = &[]pageRule{}
	}
	fonts := &[]string{}
	out := CSS{baseUrl: ressource.baseUrl}
	preprocessStylesheet(mediaType, ressource.baseUrl, stylesheet, urlFetcher, matcher,
		pageRules, fonts, fontConfig, false)
	out.matcher = *matcher
	out.pageRules = *pageRules
	out.fonts = *fonts
	return out, nil
}

func (c CSS) IsNone() bool {
	return c.baseUrl == "" && c.fonts == nil && c.matcher == nil && c.pageRules == nil
}

func newCSS(input contentInput) (CSS, error) {
	return NewCSS(input, "", nil, false, "", nil, nil, nil)
}

type contentInput interface {
	isContentInput()
	String() string
}

type InputFilename string
type InputUrl string
type InputString string
type InputReader struct {
	io.ReadCloser
}

func (c InputFilename) isContentInput() {}
func (c InputUrl) isContentInput()      {}
func (c InputString) isContentInput()   {}
func (c InputReader) isContentInput()   {}
func (c InputFilename) String() string {
	return string(c)
}
func (c InputUrl) String() string {
	if strings.HasPrefix(string(c), "data:") {
		return fmt.Sprintf("data url of len. %d", len(c))
	}
	return string(c)
}
func (c InputString) String() string {
	return fmt.Sprintf("string of len. %d", len(c))
}
func (c InputReader) String() string {
	return fmt.Sprintf("reader of type %T", c.ReadCloser)
}

type source struct {
	content []byte // utf8 encoded
	baseUrl string
}

// Check that only one input is not None, and return it with the
// normalized ``baseUrl``.
// checkCssMimeType=false
// source may have nil content
func selectSource(input contentInput, baseUrl string, urlFetcher utils.UrlFetcher,
	checkCssMimeType bool) (out source, err error) {

	if baseUrl != "" {
		baseUrl, err = utils.EnsureUrl(baseUrl)
		if err != nil {
			return
		}
	}
	switch data := input.(type) {
	case InputFilename:
		if baseUrl == "" {
			baseUrl, err = utils.Path2url(string(data))
			if err != nil {
				return
			}
		}
		f, err := ioutil.ReadFile(string(data))
		if err != nil {
			return source{}, err
		}
		return source{content: f, baseUrl: baseUrl}, nil
	case InputUrl:
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
			decoded, err := decodeToUtf8(result.Content, result.ProtocolEncoding)
			if err != nil {
				return source{}, err
			}
			if err = result.Content.Close(); err != nil {
				return source{}, err
			}
			return source{content: decoded, baseUrl: baseUrl}, nil
		}
	case InputReader:
		bt, err := ioutil.ReadAll(data.ReadCloser)
		if err != nil {
			return source{}, err
		}
		if err = data.ReadCloser.Close(); err != nil {
			return source{}, err
		}
		return source{content: bt, baseUrl: baseUrl}, nil
	case InputString:
		return source{content: []byte(data), baseUrl: baseUrl}, nil
	default:
		return source{}, errors.New("unexpected css input")
	}
}

func decodeToUtf8(data io.Reader, encod string) ([]byte, error) {
	if encod == "" { // assume UTF8
		return ioutil.ReadAll(data)
	}
	enc, _ := charset.Lookup(encod)
	if enc == nil {
		return nil, fmt.Errorf("unsupported encoding %s", encod)
	}
	return ioutil.ReadAll(enc.NewDecoder().Reader(data))
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
	specificity cascadia.Specificity
	pseudoType  string
	payload     []validation.ValidatedProperty
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
	A, B  int
	Group []parser.Token //TODO: handle groups
}

func (p pageIndex) IsNone() bool {
	return p.A == 0 && p.B == 0 && p.Group == nil
}

type pageSelector struct {
	Side         string
	Blank, First bool
	Name         string
	Index        pageIndex
	Specificity  cascadia.Specificity
}
