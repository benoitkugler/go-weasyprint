package weasyprint

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"
)


// Turn an IRI that can contain any Unicode character into an ASCII-only  URI that conforms to RFC 3986.
func iriToUri(url string) string {
    if strings.HasPrefix(url, "data:") {
        // Data URIs can be huge, but don’t need this anyway.
		return url
	}
    // Use UTF-8 as per RFC 3987 (IRI), except for file://
    url = url.encode(FILESYSTEM_ENCODING
                     if url.startswith("file:") else "utf-8")
    // This is a full URI, not just a component. Only %-encode characters
    // that are not allowed at all in URIs. Everthing else is "safe":
    // * Reserved characters: /:?#[]@!$&'()*+,;=
    // * Unreserved characters: ASCII letters, digits and -._~
    //   Of these, only '~' is not in urllib’s "always safe" list.
    // * '%' to avoid double-encoding
	return quote(url, safe=b"/:?#[]@!$&'()*+,;=~%")
	}


// Like urllib.urljoin, but warn if base_url is required but missing.
func url_join(base_url, url, allow_relative, context, context_args):
    if url_is_absolute(url):
        return iriToUri(url)
    elif base_url:
        return iriToUri(urljoin(base_url, url))
    elif allow_relative:
        return iriToUri(url)
    else:
        LOGGER.error('Relative URI reference without a base URI: ' + context,
                     *context_args)
		return None
		
	
// Get the URI corresponding to the ``attrName`` attribute.
// Return "" if:
// * the attribute is empty or missing or,
// * the value is a relative URI but the document has no base URI and
//   ``allowRelative`` is ``False``.
// Otherwise return an URI, absolute if possible.
func getUrlAttribute(element html.Node, attrName, baseUrl string, allowRelative bool) string {
    value = element.get(attrName, "").strip()
    if value != "" {
		return path.Join(baseUrl , value, allowRelative, "<%s %s='%s'>",
			(element.tag, attrName, value))}
			return ""
	}

			
// Return ('external', absolute_uri) or
// ('internal', unquoted_fragment_id) or None.
func getLinkAttribute(element html.Node, attrName string, baseUrl string) []string {
	attrValue := element.get(attrName, "").strip()
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
			parsed, baseParsed := urlsplit(uri), urlsplit(baseUrl)
			// Compare with fragments removed
			if parsed[:len(parsed)-2] == baseParsed[:len(baseParsed)-2] {
				unescaped, err := url.PathUnescape(parsed.fragment)
				if err != nil {
					return nil
				}
				return []string{"internal", unescaped}
			}
		}
		return []string{"external", uri}
	}
	return nil
}
