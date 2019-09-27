package tree

import (
	"bytes"
	"fmt"

	"github.com/benoitkugler/go-weasyprint/logger"

	"github.com/benoitkugler/go-weasyprint/utils"
	"golang.org/x/net/html"
)

//Represents an HTML document parsed by net/html.
type HTML struct {
	root       *utils.HTMLNode
	mediaType  string
	urlFetcher utils.UrlFetcher
	baseUrl    string
}

// `baseUrl` is the base used to resolve relative URLs
// (e.g. in ``<img src="../foo.png">``). If not provided, try to use
// the input filename, URL, or ``name`` attribute of :term:`file objects
//        <file object>`.
// `urlFetcher` is a function called to
// fetch external resources such as stylesheets and images, UTF-8 encoded
// `mediaType` is the media type to use for ``@media``, and defaults to ``'print'``.
func NewHTML(htmlContent contentInput, baseUrl string, urlFetcher utils.UrlFetcher, mediaType string) (*HTML, error) {
	logger.ProgressLogger.Println("Step 1 - Fetching and parsing HTML")
	if urlFetcher == nil {
		urlFetcher = utils.DefaultUrlFetcher
	}
	if mediaType == "" {
		mediaType = "print"
	}
	result, err := selectSource(htmlContent, baseUrl, urlFetcher, false)
	if err != nil {
		return nil, fmt.Errorf("can't fetch html input : %s", err)
	}
	root, err := html.Parse(bytes.NewReader(result.content))
	if err != nil {
		return nil, fmt.Errorf("invalid html input : %s", err)
	}
	var out HTML
	out.root = (*utils.HTMLNode)(root)
	out.baseUrl = utils.FindBaseUrl(root, result.baseUrl)
	out.urlFetcher = urlFetcher
	out.mediaType = mediaType
	return &out, nil
}

func newHtml(htmlContent contentInput) (*HTML, error) {
	return NewHTML(htmlContent, "", nil, "")
}

func (h HTML) AsHTML() HTML {
	return h
}

func (h HTML) UAStyleSheet() CSS {
	return html5UAStylesheet
}

func (h HTML) UPHStyleSheet() CSS {
	return html5PHStylesheet
}
