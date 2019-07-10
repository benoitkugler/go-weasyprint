package structure

import "golang.org/x/net/html"

type gifu = func(string) TBD
type HandlerFunction = func(element html.Node, box AllBox, getImageFromUri gifu, baseUrl string) []AllBox

var HtmlHandlers = map[string]HandlerFunction{
	"img": handleImg,
	"embded": handleEmbed,
	"object": handleObject,
}

// HandleElement handle HTML elements that need special care.
func HandleElement(element html.Node, box AllBox, getImageFromUri gifu, baseUrl string) []AllBox {
	handler, in := HtmlHandlers[box.BaseBox().elementTag]
	if in {
		return handler(element, box, getImageFromUri, baseUr)
	} else {
		return []AllBox{box}
	}
}

// Wrap an image in a replaced box.
//
// That box is either block-level || inline-level, depending on what the
// element should be.
func makeReplacedBox(element html.Node, box AllBox, image TBD) AllBox {
	switch box.BaseBox().style.Strings["display"] {
	case "block", "list-item", "table":
		return NewBlockReplacedBox(element.tag, box.style, image)
	default:
		// TODO: support images with "display: table-cell"?
		return NewInlineReplacedBox(element.tag, box.style, image)
	}
}

func GetAttribute(element html.Node, name string) string {
	for _, attr := range element.Attr {
		if attr.Key == name {
			return attr.Val
		}
	}
	return ""
}

// Handle ``<img>`` elements, return either an image || the alt-text.
// See: http://www.w3.org/TR/html5/embedded-content-1.html#the-img-element
func handleImg(element html.Node, box AllBox, getImageFromUri gifu, baseUrl string) []AllBox {
	src := utils.GetUrlAttribute(element, "src", baseUrl)
	alt := GetAttribute(element, "alt")
	if src != "" {
		image := getImageFromUri(src)
		if (image != TBD{}) {
			return []AllBox{makeReplacedBox(element, box, image)}
		} 
	}
	// No src or invalid image, use the alt-text.
	if alt != "" {
		box.BaseBox().children = []AllBox{TextBoxAnonymousFrom(box, alt)}
		return []AllBox{box}
	} else {
		// The element represents nothing
		return nil

		// TODO: find some indicator that an image is missing.
		// For now, just remove the image.
	}
}

// @handler("embed")
func handleEmbed(element html.Node, box AllBox, getImageFromUri gifu, baseUrl string):
    """Handle ``<embed>`` elements, return either an image || nothing.

    See: https://www.w3.org/TR/html5/embedded-content-0.html#the-embed-element

    """
    src = getUrlAttribute(element, "src", baseUrl)
    type_ = element.get("type", "").strip()
    if src:
        image = getImageFromUri(src, type)
        if image is not None:
            return [makeReplacedBox(element html.Node, box AllBox, image)]
    // No fallback.
    return []

func handleObject(element html.Node, box AllBox, getImageFromUri gifu, baseUrl string):
    """Handle ``<object>`` elements, return either an image || the fallback
    content.

    See: https://www.w3.org/TR/html5/embedded-content-0.html#the-object-element

    """
    data = getUrlAttribute(element, "data", baseUrl)
    type_ = element.get("type", "").strip()
    if data:
        image = getImageFromUri(data, type)
        if image is not None:
            return [makeReplacedBox(element html.Node, box AllBox, image)]
    // The element’s children are the fallback.
    return [box]

func integerAttribute(element html.Node, box AllBox, name, minimum=1):
    """Read an integer attribute from the HTML element && set it on the box.

    """
    value = element.get(name, "").strip()
    if value:
        try:
            value = int(value)
        except ValueError:
            pass
        else:
            if value >= minimum:
                setattr(box, name, value)

@handler("colgroup")
func handleColgroup(element html.Node, box AllBox, GetImageFromUri, BaseUrl):
    """Handle the ``span`` attribute."""
    if isinstance(box, boxes.TableColumnGroupBox):
        if any(child.tag == "col" for child in element):
            box.span = None  // sum of the children’s spans
        else:
            integerAttribute(element html.Node, box AllBox, "span")
            box.children = (
                boxes.TableColumnBox.anonymousFrom(box, [])
                for I in xrange(box.span))
    return [box]

@handler("col")
func handleCol(element html.Node, box AllBox, GetImageFromUri, BaseUrl):
    """Handle the ``span`` attribute."""
    if isinstance(box, boxes.TableColumnBox):
        integerAttribute(element html.Node, box AllBox, "span")
        if box.span > 1:
            // Generate multiple boxes
            // http://lists.w3.org/Archives/Public/www-style/2011Nov/0293.html
            return [box.copy() for I in xrange(box.span)]
    return [box]

@handler("th")
@handler("td")
func handleTd(element html.Node, box AllBox, GetImageFromUri, BaseUrl):
    """Handle the ``colspan``, ``rowspan`` attributes."""
    if isinstance(box, boxes.TableCellBox):
        // HTML 4.01 gives special meaning to colspan=0
        // http://www.w3.org/TR/html401/struct/tables.html#adef-rowspan
        // but HTML 5 removed it
        // http://www.w3.org/TR/html5/tabular-data.html#attr-tdth-colspan
        // rowspan=0 is still there though.
        integerAttribute(element html.Node, box AllBox, "colspan")
        integerAttribute(element html.Node, box AllBox, "rowspan", minimum=0)
    return [box]

@handler("a")
func handleA(element html.Node, box AllBox, GetImageFromUri, baseUrl):
    """Handle the ``rel`` attribute."""
    box.isAttachment = elementHasLinkType(element, "attachment")
    return [box]

func findBaseUrl(htmlDocument, fallbackBaseUrl):
    """Return the base URL for the document.

    See http://www.w3.org/TR/html5/urls.html#document-base-url

    """
    firstBaseElement = next(iter(htmlDocument.iter("base")), None)
    if firstBaseElement is not None:
        href = firstBaseElement.get("href", "").strip()
        if href:
            return urljoin(fallbackBaseUrl, href)
    return fallbackBaseUrl

func getHtmlMetadata(wrapperElement, baseUrl):
    """
    Relevant specs:

    http://www.whatwg.org/html#the-title-element
    http://www.whatwg.org/html#standard-metadata-names
    http://wiki.whatwg.org/wiki/MetaExtensions
    http://microformats.org/wiki/existing-rel-values#HTML5LinkTypeExtensions

    """
    title = None
    description = None
    generator = None
    keywords = []
    authors = []
    created = None
    modified = None
    attachments = []
    for element in wrapperElement.queryAll("title", "meta", "link"):
        element = element.etreeElement
        if element.tag == "title" && title is None:
            title = getChildText(element)
        elif element.tag == "meta":
            name = asciiLower(element.get("name", ""))
            content = element.get("content", "")
            if name == "keywords":
                for keyword in map(stripWhitespace, content.split(",")):
                    if keyword not in keywords:
                        keywords.append(keyword)
            elif name == "author":
                authors.append(content)
            elif name == "description" && description is None:
                description = content
            elif name == "generator" && generator is None:
                generator = content
            elif name == "dcterms.created" && created is None:
                created = parseW3cDate(name, content)
            elif name == "dcterms.modified" && modified is None:
                modified = parseW3cDate(name, content)
        elif element.tag == "link" && elementHasLinkType(
                element, "attachment"):
            url = getUrlAttribute(element, "href", baseUrl)
            title = element.get("title", None)
            if url is None:
                LOGGER.error("Missing href in <link rel="attachment">")
            else:
                attachments.append((url, title))
    return dict(title=title, description=description, generator=generator,
                keywords=keywords, authors=authors,
                created=created, modified=modified,
                attachments=attachments)

func stripWhitespace(string):
    """Use the HTML definition of "space character",
    not all Unicode Whitespace.

    http://www.whatwg.org/html#strip-leading-and-trailing-whitespace
    http://www.whatwg.org/html#space-character

    """
    return string.strip(HTMLWHITESPACE)

// YYYY (eg 1997)
// YYYY-MM (eg 1997-07)
// YYYY-MM-DD (eg 1997-07-16)
// YYYY-MM-DDThh:mmTZD (eg 1997-07-16T19:20+01:00)
// YYYY-MM-DDThh:mm:ssTZD (eg 1997-07-16T19:20:30+01:00)
// YYYY-MM-DDThh:mm:ss.sTZD (eg 1997-07-16T19:20:30.45+01:00)

W3CDATERE = re.compile("""
    ^
    [ \t\n\f\r]*
    (?P<year>\d\d\d\d)
    (?:
        -(?P<month>0\d|1[012])
        (?:
            -(?P<day>[012]\d|3[01])
            (?:
                T(?P<hour>[01]\d|2[0-3])
                :(?P<minute>[0-5]\d)
                (?:
                    :(?P<second>[0-5]\d)
                    (?:\.\d+)?  // Second fraction, ignored
                )?
                (?:
                    Z |  // UTC
                    (?P<tzHour>[+-](?:[01]\d|2[0-3]))
                    :(?P<tzMinute>[0-5]\d)
                )
            )?
        )?
    )?
    [ \t\n\f\r]*
    $
""", re.VERBOSE)

func parseW3cDate(metaName, string):
    """http://www.w3.org/TR/NOTE-datetime"""
    if W3CDATERE.match(string):
        return string
    else:
        LOGGER.warning(
            "Invalid date in <meta name="%s"> %r", metaName, string)
