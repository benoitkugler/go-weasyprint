package utils

import (
	"fmt"
	"log"
	"net/url"
	"path"
	"strings"

	"golang.org/x/net/html"
)

// Turn an IRI that can contain any Unicode character into an ASCII-only  URI that conforms to RFC 3986.
func iriToUri(urlS string) string {
	if strings.HasPrefix(urlS, "data:") {
		// Data URIs can be huge, but don’t need this anyway.
		return urlS
	}
	// This is a full URI, not just a component. Only %-encode characters
	// that are not allowed at all in URIs. Everthing else is "safe":
	// * Reserved characters: /:?#[]@!$&'()*+,;=
	// * Unreserved characters: ASCII letters, digits and -._~
	//   Of these, only '~' is not in urllib’s "always safe" list.
	// * '%' to avoid double-encoding
	return url.PathEscape(urlS)
}

// warn if baseUrl is required but missing.
func UrlJoin(baseUrl, urlS string, allowRelative bool, context ...interface{}) string {
	if path.IsAbs(urlS) {
		return iriToUri(urlS)
	} else if baseUrl != "" {
		return iriToUri(path.Join(baseUrl, urlS))
	} else if allowRelative {
		return iriToUri(urlS)
	} else {
		log.Println("Relative URI reference without a base URI: ", context)
		return ""
	}
}

// Get the URI corresponding to the ``attrName`` attribute.
// Return "" if:
// * the attribute is empty or missing or,
// * the value is a relative URI but the document has no base URI and
//   ``allowRelative`` is ``False``.
// Otherwise return an URI, absolute if possible.
func getUrlAttribute(element html.Node, attrName, baseUrl string, allowRelative bool) string {
	value := strings.TrimSpace(GetAttribute(element, attrName))
	if value != "" {
		return UrlJoin(baseUrl, value, allowRelative,
			fmt.Sprintf("<%s %s='%s'>", element.Data, attrName, value))
	}
	return ""
}

// Return ('external', absolute_uri) or
// ('internal', unquoted_fragment_id) or nil.
func GetLinkAttribute(element html.Node, attrName string, baseUrl string) []string {
	attrValue := strings.TrimSpace(GetAttribute(element, attrName))
	if strings.HasPrefix(attrValue, "#") && len(attrValue) > 1 {
		// Do not require a baseUrl when the value is just a fragment.
		unescaped, err := url.PathUnescape(attrValue[1:])
		if err != nil {
			return nil
		}
		return []string{"internal", unescaped}
	}
	uri := getUrlAttribute(element, attrName, baseUrl, true)
	if uri != "" {
		if baseUrl != "" {
			parsed, err := url.Parse(uri)
			if err != nil {
				log.Println(err)
				return nil
			}
			baseParsed, err := url.Parse(baseUrl)
			if err != nil {
				log.Println(err)
				return nil
			}
			if parsed.Scheme == baseParsed.Scheme && parsed.Host == baseParsed.Host && parsed.Path == baseParsed.Path && parsed.RawQuery == baseParsed.RawQuery {
				// Compare with fragments removed
				unescaped, err := url.PathUnescape(parsed.Fragment)
				if err != nil {
					log.Println(err)
					return nil
				}
				return []string{"internal", unescaped}
			}
		}
		return []string{"external", uri}
	}
	return nil
}
