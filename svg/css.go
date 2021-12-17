package svg

import (
	"encoding/xml"
	"log"
	"strings"

	"github.com/benoitkugler/cascadia"
	"github.com/benoitkugler/go-weasyprint/style/parser"
	"github.com/benoitkugler/go-weasyprint/utils"
)

// Apply CSS to SVG documents.

// http://www.w3.org/TR/SVG/styling.html#StyleElement
func handleStyleElement(d *xml.Decoder, start xml.StartElement) ([]byte, error) {
	if start.Name.Local != "style" {
		return nil, nil
	}
	for _, v := range start.Attr {
		if v.Name.Local == "type" && v.Value != "text/css" {
			return nil, nil
		}
	}

	// extract the css
	var css []byte
	for {
		next, err := d.Token()
		if err != nil {
			return nil, err
		}
		// Token is one of StartElement, EndElement, CharData, Comment, ProcInst, or Directive
		switch next := next.(type) {
		case xml.CharData:
			// handle text and keep going
			css = append(css, next...)
		case xml.EndElement:
			return css, nil
		}
	}
}

func fetchURL(url, baseURL string) ([]byte, string, error) {
	joinedUrl, err := utils.SafeUrljoin(baseURL, url, true)
	if err != nil {
		return nil, "", err
	}
	cssUrl, err := parseURL(joinedUrl)
	if err != nil {
		return nil, "", err
	}
	resolvedURL := cssUrl.String()
	content, err := utils.FetchSource(utils.InputUrl(resolvedURL), baseURL, utils.DefaultUrlFetcher, false)
	if err != nil {
		return nil, "", err
	}
	return content.Content, resolvedURL, nil
}

// Find rules among stylesheet rules and imports.
func findStylesheetsRules(rules []parser.Token, baseUrl string) (out []parser.QualifiedRule) {
	for _, rule := range rules {
		switch rule := rule.(type) {
		case parser.AtRule:
			if rule.AtKeyword.Lower() == "import" && rule.Content == nil {
				urlToken := parser.ParseOneComponentValue(*rule.Prelude)
				var url string
				switch urlToken := urlToken.(type) {
				case parser.StringToken:
					url = urlToken.Value
				case parser.URLToken:
					url = urlToken.Value
				default:
					continue
				}
				cssContent, resolvedURL, err := fetchURL(url, baseUrl)
				if err != nil {
					log.Printf("failed to load stylesheet: %s", err)
					continue
				}

				stylesheet := parser.ParseStylesheetBytes(cssContent, true, true)
				out = append(out, findStylesheetsRules(stylesheet, resolvedURL)...)
			}
			// if rule.AtKeyword.Lower() == "media":
		case parser.QualifiedRule:
			out = append(out, rule)
			// elif rule.type == "error":
		}
	}
	return out
}

type declaration struct {
	property string
	value    []parser.Token
}

// Parse declarations in a given rule content.
func parseDeclarations(input []parser.Token) (normalDeclarations, importantDeclarations []declaration) {
	for _, decl := range parser.ParseDeclarationList(input, false, false) {
		if decl, ok := decl.(parser.Declaration); ok {
			if strings.HasPrefix(string(decl.Name), "-") {
				continue
			}
			if decl.Important {
				importantDeclarations = append(importantDeclarations, declaration{decl.Name.Lower(), decl.Value})
			} else {
				normalDeclarations = append(normalDeclarations, declaration{decl.Name.Lower(), decl.Value})
			}
		}
	}
	return normalDeclarations, importantDeclarations
}

type match struct {
	selector     cascadia.SelectorGroup
	declarations []declaration
}

type matcher []match

// Find stylesheets and return rule matchers.
func parseStylesheets(stylesheets [][]byte, url string) (matcher, matcher) {
	var normalMatcher, importantMatcher matcher
	// Parse rules and fill matchers
	for _, css := range stylesheets {
		stylesheet := parser.ParseStylesheetBytes(css, true, true)
		for _, rule := range findStylesheetsRules(stylesheet, url) {
			normalDeclarations, importantDeclarations := parseDeclarations(*rule.Content)
			prelude := parser.Serialize(*rule.Prelude)
			selector, err := cascadia.ParseGroupWithPseudoElements(prelude)
			if err != nil {
				log.Printf("Invalid or unsupported selector '%s', %s \n", prelude, err)
				continue
			}
			if len(normalDeclarations) != 0 {
				normalMatcher = append(normalMatcher, match{selector: selector, declarations: normalDeclarations})
			}
			if len(importantDeclarations) != 0 {
				importantMatcher = append(importantMatcher, match{selector: selector, declarations: importantDeclarations})
			}
		}
	}
	return normalMatcher, importantMatcher
}
