package goweasyprint

import (
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	bo "github.com/benoitkugler/go-weasyprint/boxes"
)

func toF(v pr.Value) float64 { return float64(v.Value)}

// Return the matrix for the CSS transforms on this box (possibly nil)
func getMatrix(box_ Box) bo.Matrix {
    // "Transforms apply to block-level and atomic inline-level elements,
    //  but do not apply to elements which may be split into
    //  multiple inline-level boxes."
	// http://www.w3.org/TR/css3-2d-transforms/#introduction
	box := box_.Box()
    if trans := box.Style.GetTransform(); len(trans) != 0 && !bo.TypeInlineBox.IsInstance(box_) {
        borderWidth := box.BorderWidth()
        borderHeight := box.BorderHeight()
		or := box.Style.GetTransformOrigin()
        offsetX := pr.ResoudPercentage(or[0].ToValue(), borderWidth).V()
        offsetY := pr.ResoudPercentage(or[1].ToValue(), borderHeight).V()
        originX := box.BorderBoxX() + offsetX
        originY := box.BorderBoxY() + offsetY
    
        var matrix bo.Matrix = cairo.Matrix()
        matrix.Translate(originX, originY)
        for _, t := range trans {
			name, args := t.String, t.Dimensions
			// The length of args depends on `name`, see package validation for details.
			switch name {
			case"scale" :
				sx, sy := toF(args[0]), toF(args[1])
                matrix.Scale(sx, sy)
            case"rotate" :
				angle := toF(args[0]) 
                matrix.Rotate(angle)
            case"translate" :
                translateX, translateY := args[0], args[1]
                matrix.Translate(
                    float64(pr.ResoudPercentage(translateX.ToValue(), borderWidth).V()),
                    float64(pr.ResoudPercentage(translateY.ToValue(), borderHeight).V()),
                )
			default:
				var leftMat [6]float64
				switch name {
                case "skewx" :
                    leftMat = [6]float64{1, 0, math.Tan(toF(args[0])), 1, 0, 0}
                case "skewy" :
                    leftMat = [6]float64{1, math.Tan(toF(args[0])), 0, 1, 0, 0}
				case "matrix":
					leftMat = [6]float64{toF(args[0]), toF(args[1]), toF(args[2]), 
						toF(args[3]), toF(args[4]), toF(args[5])}
				default:
                    log.Fatalf("unexpected name for CSS transform property : %s", name)
				} 
				matrix = matrix.LeftMat(leftMat) 
            }
		} 
		matrix.Translate(-originX, -originY)
        box.TransformationMatrix = matrix
        return matrix
		}
		return nil 
	}

// Apply a transformation matrix to an axis-aligned rectangle
// and return its axis-aligned bounding box as ``(x, y, width, height)`` 
func rectangleAabb(matrix bo.Matrix, posX, posY, width, height float64) [4]float64 {
    x1, y1 := matrix.TransformPoint(posX, posY)
    x2, y2 := matrix.TransformPoint(posX + width, posY)
    x3, y3 := matrix.TransformPoint(posX, posY + height)
    x4, y4 := matrix.TransformPoint(posX + width, posY + height)
    boxX1 := utils.Mins(x1, x2, x3, x4)
    boxY1 := utils.Mins(y1, y2, y3, y4)
    boxX2 := utils.Maxs(x1, x2, x3, x4)
    boxY2 := utils.Maxs(y1, y2, y3, y4)
    return [4]float64{boxX1, boxY1, boxX2 - boxX1, boxY2 - boxY1}
} 

type linkData struct {
	type_, target string 
	rectangle [4]float64
}

type bookmarkData struct {
	level int 
	label string 
	position [2]float64
	state string
}

func gatherLinksAndBookmarks(box_ bo.Box, bookmarks, links *[]linkData, anchors map[string][2]float64, matrix bo.Matrix) {
    transform := getMatrix(box_)
    if transform != nil {
		if matrix != nil {
			matrix = transform.RightMultiply(matrix.Data())  
		} else {
			matrix = transform
		}
    }
	box := box_.Box()
    bookmarkLabel := box.BookmarkLabel
	bookmarkLevel := 0
    if lvl := box.Style.GetBookmarkLevel(); lvl.String != "none" {
     	bookmarkLevel = lvl.Int
	} 
	state := string(box.Style.GetBookmarkState())
    link := box.Style.GetLink()
    anchorName := box.Style.GetAnchor()
    hasBookmark := bookmarkLabel != "" && bookmarkLevel != 0
    // "link" is inherited but redundant on text boxes
    hasLink := !link.IsNone() && !bo.IsTextBox(box_)
	// In case of duplicate IDs, only the first is an anchor.
	_, inAnchors := anchors[anchorName]
    hasAnchor != anchorName != "" && !inAnchors  
    isAttachment := box.IsAttachment

    if hasBookmark || hasLink || hasAnchor {
        posX, posY, width, height := box.HitArea()
        if hasLink {
            linkType, target := link.Name, link.String
            if linkType == "external" && isAttachment {
                linkType = "attachment"
			} 
			var linkS link
			if matrix != nil {
				linkS = linkData{type_:linkType, target: target,
					rectangle: rectangleAabb(matrix, posX, posY, width, height) }
            } else {
                linkS = linkData{type_:linkType, target: target, rectangle: [4]float64{posX, posY, width, height}}
			} 
			*links = append(*links, linkS)
		} 
		if matrix != nil && (hasBookmark || hasAnchor) {
            posX, posY = matrix.TransformPoint(posX, posY)
		} 
		if hasBookmark {
			*bookmarks = append(*bookmarks, bookmarkData{level:bookmarkLevel,label: bookmarkLabel,
				position: [2]float64{posX, posY}, state:state})
		} 
		if hasAnchor {
            anchors[anchorName] = [2]float64{posX, posY}
        }
    }

    for _, child := range box.AllChildren() {
        gatherLinksAndBookmarks(child, bookmarks, links, anchors, matrix)
    }
}

func toInt(s string, defaut ...int) int {
	if s == "" && len(defaut) > 0 {
		return defaut[0]
	}
	out, err := strconv.Atoi(s) 
	if err != nil {
		log.Fatalf("unexpected string for int : %s", s)
	}
	return out
}

// Tranform W3C date to ISO-8601 format.
func w3cDateToIso(str, attrName string) string {
    if str  == ""  {
        return ""
	} 
	match := utils.W3CDateRe.match(str)
    if len(match)  == 0  {
        log.Printf("Invalid %s date: %s", attrName, str)
        return ""
	} 
	groups = match.groupdict()
    isoDate := fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d",
        toInt(groups["year"]),
        toInt(groups["month"] , 1),
        toInt(groups["day"] , 1),
        toInt(groups["hour"] , 0),
        toInt(groups["minute"] , 0),
        toInt(groups["second"] , 0))
    if groups["hour"] != "" {
        if groups["minute"] == "" {
			log.Fatalf("minute shouldn't be empty when hour is present")
		}
        if groups["tzHour"] != "" {
            if !(strings.HasPrefix(groups["tzHour"], "+") || strings.HasPrefix(groups["tzHour"], "-")) {
				log.Fatalf("tzHour should start by + or -, got %s", groups["tzHour"])
			}
            if groups["tzMinute"] == "" {
				log.Fatalf("tzMinute shouldn't be empty when tzHour is present")
			}
            isoDate += fmt.Sprintf("%+03d:%02d", toInt(groups["tzHour"]), toInt(groups["tzMinute"]))
        } else {
            isoDate += "+00:00"
        }
	} 
	return isoDate
} 

// Represents a single rendered page.
// Should be obtained from `Document.pages` but not
// instantiated directly.
type Page struct {
			// The page width, including margins, in CSS pixels.
			width  float64
 
			// The page height, including margins, in CSS pixels.
			height float64
		
			// The page bleed widths as a `dict` with `"top"`, `"right"`,
			// `"bottom"` and `"left"` as keys, and values in CSS pixels.
			bleed bleedData
		
			// `bookmarkLevel` and `bookmarkLabel` are based on
			// the CSS properties of the same names. `target` is an `(x, y)`
			// point in CSS pixels from the top-left of the page.
			bookmarks []bookmarkData
		
			// The `list` of `(linkType, target, rectangle)` `tuples
			// <tuple>`. A `rectangle` is `(x, y, width, height)`, in CSS
			// pixels from the top-left of the page. `linkType` is one of three
			// strings :
			// * `"external"`: `target` is an absolute URL
			// * `"internal"`: `target` is an anchor name (see
			//   :attr:`Page.anchors`).
			//   The anchor might be defined in another page,
			//   in multiple pages (in which case the first occurence is used),
			//   or not at all.
			// * `"attachment"`: `target` is an absolute URL && points
			//   to a resource to attach to the document.
			links []linkData
		
			// The `dict` mapping each anchor name to its target, an
			// `(x, y)` point in CSS pixels from the top-left of the page.
			anchors map[string][2]float64
}
// enableHinting=false
	func NewPage(pageBox *bo.PageBox, enableHinting bool) {
		self := new(Page)
		self.width = float64(pageBox.MarginWidth())
 		self.height = float64(pageBox.MarginHeight())
	
		self.bleed = bleedData{
			Top :pageBox.Style.GetBleedTop().Value,
			 Right :pageBox.Style.GetBleedRight().Value,
			 Bottom :pageBox.Style.GetBleedBottom().Value,
			 Left :pageBox.Style.GetBleedLeft().Value,
		}
		self.anchors = map[string][2]float64{}
	
		gatherLinksAndBookmarks(
			pageBox, &self.bookmarks, &self.links, self.anchors, nil)
		self.pageBox = pageBox
		self.enableHinting = enableHinting
		}

	// Paint the page in cairo, on any type of surface (leftX=0, topY=0, scale=1, clip=false).
	// leftX is the X coordinate of the left of the page, in cairo user units.
	// topY is the Y coordinate of the top of the page, in cairo user units.
	// scale is the Zoom scale in cairo user units per CSS pixel.
	// clip : whether to clip/cut content outside the page. If false, content can overflow.
	func (self Page) paint(cairoContext Drawer, leftX, topY, scale float64, clip bool) {
		// with stacked(cairoContext) {
			if self.EnableHinting {
				leftX, topY = cairoContext.userToDevice(leftX, topY)
				// Hint := range device space
				leftX = int(leftX)
				topY = int(topY)
				leftX, topY = cairoContext.deviceToUser(leftX, topY)
			} 
			// Make (0, 0) the top-left corner:
			 cairoContext.translate(leftX, topY)
			// Make user units CSS pixels:
			cairoContext.scale(scale, scale)
			if clip {
				width = self.width
				height = self.height
				if self.EnableHinting {
					width, height = (
						cairoContext.userToDeviceDistance(width, height))
					// Hint := range device space
					width = int(math.ceil(width))
					height = int(math.ceil(height))
					width, height = (
						cairoContext.deviceToUserDistance(width, height))
				} 
				cairoContext.rectangle(0, 0, width, height)
				cairoContext.clip()
			} 
			drawPage(self.PageBox, cairoContext, self.EnableHinting)
					}




type BookmarkSubtree struct {
	label string
	destination string
	children string
	state string
}


// A rendered document ready to be painted on a cairo surface.

// Typically obtained from `HTML.render()`, but
// can also be instantiated directly with a list of `pages <Page>`, a
// set of `metadata <DocumentMetadata>`, a `urlFetcher` function, and 
// a `fontConfig <fonts.FontConfiguration>`.
type Document struct {
    // A list of `Page` objects.
    pages []Page
    // A `DocumentMetadata` object.
    // Contains information that does not belong to a specific page
    // but to the whole document.
    metadata utils.DocumentMetadata
    // A function called to fetch external resources such
    // as stylesheets and images. 
    urlFetcher utils.UrlFetcher
    // Keep a reference to fontConfig to avoid its garbage collection until
    // rendering is destroyed. 
    fontConfig 
}

// presentationalHints=false, fontConfig=None
func buildLayoutContext(html tree.HTML, stylesheets []tree.CSS, enableHinting,
                            presentationalHints bool, fontConfig *fonts.FontConfiguration) layout.LayoutContext {
                            
    if fontConfig  == nil  {
        fontConfig = fonts.NewFontConfiguration()
	} 
	targetCollector = tree.NewTargetCollector()
    var pageRules, userStylesheets  []tree.CSS
    for _, css := range stylesheets {
        if ! hasattr(css, "matcher") {
            // css = tree.NewCSS(css, mediaType=html.mediaType, fontConfig=fontConfig)
		} 
		userStylesheets = append(userStylesheets, css)
	} 
	styleFor = tree.GetAllComputedStyles(html, userStylesheets, presentationalHints, fontConfig,
		pageRules, targetCollector)
		cache := make(map[string]int)
    getImageFromUri := func(url, forcedMimeType string) utils.RemoteRessource {
        return images.GetImageFromUri(cache, html.UrlFetcher,url, forforcedMimeType)
	} 
    logger.ProgressLogger.Println("Step 4 - Creating formatting structure")
    context := NewLayoutContext(enableHinting, styleFor, getImageFromUri, fontConfig,targetCollector)
    return context
	}

// presentationalHints=false, fontConfig=None
func render(html tree.HTML, stylesheets []tree.CSS, enableHinting,
	presentationalHints bool, fontConfig *fonts.FontConfiguration) {
        
    if fontConfig  == nil  {
        fontConfig = fonts.NewFontConfiguration()
    }

    context := buildLayoutContext(html, stylesheets, enableHinting, presentationalHints,fontConfig)

    rootBox := bo.BuildFormattingStructure(html.root, context.styleFor, context.getImageFromUri,
        html.baseUrl, context.targetCollector)

    pageBoxes := layout.LayoutDocument(html, rootBox, context)
	pages := make([]Page, len(pageBoxes))
	for i, pageBox := range pageBoxes{
		pages[i] = NewPage(pageBox, enableHinting)
	}
	rendering = NewDocument(pages, html.GetMetadata(), html.urlFetcher, fontConfig)
    return rendering

func NewDocument(pages, metadata, urlFetcher, fontConfig) {
    #: A list of :class:`Page` objects.
    self.pages = pages
    #: A :class:`DocumentMetadata` object.
    #: Contains information that does ! belong to a specific page
    #: but to the whole document.
    self.metadata = metadata
    #: A function || other callable with the same signature as
    #: :func:`defaultUrlFetcher` called to fetch external resources such
    #: as stylesheets && images.  (See :ref:`url-fetchers`.)
    self.urlFetcher = urlFetcher
    // Keep a reference to fontConfig to avoid its garbage collection until
    // rendering is destroyed. This is needed as fontConfig._Del__ removes
    // fonts that may be used when rendering
    self.FontConfig = fontConfig
} 

// Take a subset of the pages.
//     .. versionadded:: 0.15
//     :type pages: :term:`iterable`
//     :param pages:
//         An iterable of :class:`Page` objects from :attr:`pages`.
//     :return:
//         A new :class:`Document` object.
//     Examples:
//     Write two PDF files for odd-numbered && even-numbered pages::
//         // Python lists count from 0 but pages are numbered from 1.
//         // [::2] is a slice of even list indexes but odd-numbered pages.
//         document.copy(document.pages[::2]).writePdf("oddPages.pdf")
//         document.copy(document.pages[1::2]).writePdf("evenPages.pdf")
//     Write each page to a numbred PNG file::
//         for i, page := range enumerate(document.pages):
//             document.copy(page).writePng("page%s.png" % i)
//     Combine multiple documents into one PDF file,
//     using metadata from the first::
//         allPages = [p for doc := range documents for p := range doc.pages]
//         documents[0].copy(allPages).writePdf("combined.pdf")
//     
func copy(self, pages="all") {
    if pages == "all" {
        pages = self.pages
    } else if ! isinstance(pages, list) {
        pages = list(pages)
    } return type(self)(
        pages, self.metadata, self.urlFetcher, self.FontConfig)
} 
// Resolve internal hyperlinks.
//     .. versionadded:: 0.15
//     Links to a missing anchor are removed with a warning.
//     If multiple anchors have the same name, the first one is used.
//     :returns:
//         A generator yielding lists (one per page) like :attr:`Page.links`,
//         except that ``target`` for internal hyperlinks is
//         ``(pageNumber, x, y)`` instead of an anchor name.
//         The page number is a 0-based index into the :attr:`pages` list,
//         && ``x, y`` are := range CSS pixels from the top-left of the page.
//     
func resolveLinks(self) {
    anchors = set()
    pagedAnchors = []
    for i, page := range enumerate(self.pages) {
        pagedAnchors.append([])
        for anchorName, (pointX, pointY) := range page.anchors.items() {
            if anchorName ! := range anchors {
                pagedAnchors[-1].append((anchorName, pointX, pointY))
                anchors.add(anchorName)
            }
        }
    } for page := range self.pages {
        pageLinks = []
        for link := range page.links {
            linkType, anchorName, rectangle = link
            if linkType == "internal" {
                if anchorName ! := range anchors {
                    LOGGER.error(
                        "No anchor #%s for internal URI reference",
                        anchorName)
                } else {
                    pageLinks.append((linkType, anchorName, rectangle))
                }
            } else {
                // External link
                pageLinks.append(link)
            }
        } yield pageLinks, pagedAnchors.pop(0)
    }
} 
// Make a tree of all bookmarks := range the document.
//     .. versionadded:: 0.15
//     :return: A list of bookmark subtrees.
//         A subtree is ``(label, target, children, state)``. ``label`` is
//         a string, ``target`` is ``(pageNumber, x, y)`` like in
//         :meth:`resolveLinks`, && ``children`` is a
//         list of child subtrees.
//     
func makeBookmarkTree(self) {
    root = []
    // At one point := range the document, for each "output" depth, how much
    // to add to get the source level (CSS values of bookmark-level).
    // E.g. with <h1> then <h3>, levelShifts == [0, 1]
    // 1 means that <h3> has depth 3 - 1 = 2 := range the output.
    skippedLevels = []
    lastByDepth = [root]
    previousLevel = 0
    for pageNumber, page := range enumerate(self.pages) {
        for level, label, (pointX, pointY), state := range page.bookmarks {
            if level > previousLevel {
                // Example: if the previous bookmark is a <h2>, the next
                // depth "should" be for <h3>. If now we get a <h6> we’re
                // skipping two levels: append 6 - 3 - 1 = 2
                skippedLevels.append(level - previousLevel - 1)
            } else {
                temp = level
                while temp < previousLevel {
                    temp += 1 + skippedLevels.pop()
                } if temp > previousLevel {
                    // We remove too many "skips", add some back {
                    } skippedLevels.append(temp - previousLevel - 1)
                }
            }
        }
    }
} 
            previousLevel = level
            depth = level - sum(skippedLevels)
            assert depth == len(skippedLevels)
            assert depth >= 1

            children = []
            subtree = BookmarkSubtree(
                label, (pageNumber, pointX, pointY), children, state)
            lastByDepth[depth - 1].append(subtree)
            del lastByDepth[depth:]
            lastByDepth.append(children)
    return root

// Include hyperlinks := range current PDF page.
//     .. versionadded:: 43
//     
func addHyperlinks(self, links, anchors, context, scale) {
    if cairo.cairoVersion() < 11504 {
        return
    }
} 
    // We round floats to avoid locale problems, see
    // https://github.com/Kozea/WeasyPrint/issues/742

    // TODO: Instead of using rects, we could use the drawing rectangles
    // defined by cairo when drawing targets. This would give a feeling
    // similiar to what browsers do with links that span multiple lines.
    for link := range links {
        linkType, linkTarget, rectangle = link
        if linkType == "external" {
            attributes = "rect=[{} {} {} {}] uri="{}"".format(*(
                [int(round(i * scale)) for i := range rectangle] +
                [linkTarget.replace(""", "%27")]))
        } else if linkType == "internal" {
            attributes = "rect=[{} {} {} {}] dest="{}"".format(*(
                [int(round(i * scale)) for i := range rectangle] +
                [linkTarget.replace(""", "%27")]))
        } else if linkType == "attachment" {
            // Attachments are handled := range writePdfMetadata
            continue
        } context.tagBegin(cairo.TAGLINK, attributes)
        context.tagEnd(cairo.TAGLINK)
    }

    for anchor := range anchors {
        anchorName, x, y = anchor
        attributes = "name="{}" x={} y={}".format(
            anchorName.replace(""", "%27"), int(round(x * scale)),
            int(round(y * scale)))
        context.tagBegin(cairo.TAGDEST, attributes)
        context.tagEnd(cairo.TAGDEST)
    }

// Paint the pages := range a PDF file, with meta-data.
//     PDF files written directly by cairo do ! have meta-data such as
//     bookmarks/outlines && hyperlinks.
//     :type target: str, pathlib.Path || file object
//     :param target:
//         A filename where the PDF file is generated, a file object, or
//         :obj:`None`.
//     :type zoom: float
//     :param zoom:
//         The zoom factor := range PDF units per CSS units.  **Warning**:
//         All CSS units are affected, including physical units like
//         ``cm`` && named sizes like ``A4``.  For values other than
//         1, the physical CSS units will thus be "wrong".
//     :type attachments: list
//     :param attachments: A list of additional file attachments for the
//         generated PDF document || :obj:`None`. The list"s elements are
//         :class:`Attachment` objects, filenames, URLs || file-like objects.
//     :returns:
//         The PDF as :obj:`bytes` if ``target`` is ! provided or
//         :obj:`None`, otherwise :obj:`None` (the PDF is written to
//         ``target``).
//     
func writePdf(self, target=None, zoom=1, attachments=None) {
    // 0.75 = 72 PDF point (cairo units) per inch / 96 CSS pixel per inch
    scale = zoom * 0.75
    // Use an in-memory buffer, as we will need to seek for
    // metadata. Directly using the target when possible doesn"t
    // significantly save time && memory use.
    fileObj = io.BytesIO()
    // (1, 1) is overridden by .setSize() below.
    surface = cairo.PDFSurface(fileObj, 1, 1)
    context = cairo.Context(surface)
} 
    PROGRESSLOGGER.info("Step 6 - Drawing")

    pagedLinksAndAnchors = list(self.resolveLinks())
    for page, linksAndAnchors := range zip(
            self.pages, pagedLinksAndAnchors) {
            }
        links, anchors = linksAndAnchors
        surface.setSize(
            math.floor(scale * (
                page.width + page.bleed["left"] + page.bleed["right"])),
            math.floor(scale * (
                page.height + page.bleed["top"] + page.bleed["bottom"])))
        with stacked(context) {
            context.translate(
                page.bleed["left"] * scale, page.bleed["top"] * scale)
            page.paint(context, scale=scale)
            self.addHyperlinks(links, anchors, context, scale)
            surface.showPage()
        }

    PROGRESSLOGGER.info("Step 7 - Adding PDF metadata")

    // TODO: overwrite producer when possible := range cairo
    if cairo.cairoVersion() >= 11504 {
        // Set document information
        for attr, key := range (
                ("title", cairo.PDFMETADATATITLE),
                ("description", cairo.PDFMETADATASUBJECT),
                ("generator", cairo.PDFMETADATACREATOR)) {
                }
            value = getattr(self.metadata, attr)
            if value  != nil  {
                surface.setMetadata(key, value)
            }
        for attr, key := range (
                ("authors", cairo.PDFMETADATAAUTHOR),
                ("keywords", cairo.PDFMETADATAKEYWORDS)) {
                }
            value = getattr(self.metadata, attr)
            if value  != nil  {
                surface.setMetadata(key, ", ".join(value))
            }
        for attr, key := range (
                ("created", cairo.PDFMETADATACREATEDATE),
                ("modified", cairo.PDFMETADATAMODDATE)) {
                }
            value = getattr(self.metadata, attr)
            if value  != nil  {
                surface.setMetadata(key, W3cDateToIso(value, attr))
            }
    }

        // Set bookmarks
        bookmarks = self.makeBookmarkTree()
        levels = [cairo.PDFOUTLINEROOT] * len(bookmarks)
        while bookmarks {
            bookmark = bookmarks.pop(0)
            title = bookmark.label
            destination = bookmark.destination
            children = bookmark.children
            state = bookmark.state
            page, x, y = destination
        }

            // We round floats to avoid locale problems, see
            // https://github.com/Kozea/WeasyPrint/issues/742
            linkAttribs = "page={} pos=[{} {}]".format(
                page + 1, int(round(x * scale)),
                int(round(y * scale)))

            outline = surface.addOutline(
                levels.pop(), title, linkAttribs,
                cairo.PDFOUTLINEFLAGOPEN if state == "open" else 0)
            levels.extend([outline] * len(children))
            bookmarks = children + bookmarks

    surface.finish()

    // Add extra PDF metadata: attachments, embedded files
    attachmentLinks = [
        [link for link := range pageLinks if link[0] == "attachment"]
        for pageLinks, pageAnchors := range pagedLinksAndAnchors]
    // Write extra PDF metadata only when there is a least one from {
    } // - attachments := range metadata
    // - attachments as function parameters
    // - attachments as PDF links
    // - bleed boxes
    condition = (
        self.metadata.attachments or
        attachments or
        any(attachmentLinks) or
        any(any(page.bleed.values()) for page := range self.pages))
    if condition {
        writePdfMetadata(
            fileObj, scale, self.urlFetcher,
            self.metadata.attachments + (attachments || []),
            attachmentLinks, self.pages)
    }

    if target  == nil  {
        return fileObj.getvalue()
    } else {
        fileObj.seek(0)
        if hasattr(target, "write") {
            shutil.copyfileobj(fileObj, target)
        } else {
            with open(target, "wb") as fd {
                shutil.copyfileobj(fileObj, fd)
            }
        }
    }

// Render pages on a cairo image surface.
//     .. versionadded:: 0.17
//     There is no decoration around pages other than those specified := range CSS
//     with ``@page`` rules. The final image is as wide as the widest page.
//     Each page is below the previous one, centered horizontally.
//     :type resolution: float
//     :param resolution:
//         The output resolution := range PNG pixels per CSS inch. At 96 dpi
//         (the default), PNG pixels match the CSS ``px`` unit.
//     :returns:
//         A ``(surface, pngWidth, pngHeight)`` tuple. ``surface`` is a
//         cairo :class:`ImageSurface <cairocffi.ImageSurface>`. ``pngWidth``
//         && ``pngHeight`` are the size of the final image, := range PNG pixels.
//     
func writeImageSurface(self, resolution=96) {
    dppx = resolution / 96
} 
    // This duplicates the hinting logic := range Page.paint. There is a
    // dependency cycle otherwise {
    } //   this → hinting logic → context → surface → this
    // But since we do no transform here, cairoContext.userToDevice and
    // friends are identity functions.
    widths = [int(math.ceil(p.width * dppx)) for p := range self.pages]
    heights = [int(math.ceil(p.height * dppx)) for p := range self.pages]

    maxWidth = max(widths)
    sumHeights = sum(heights)
    surface = cairo.ImageSurface(
        cairo.FORMATARGB32, maxWidth, sumHeights)
    context = cairo.Context(surface)
    posY = 0
    PROGRESSLOGGER.info("Step 6 - Drawing")
    for page, width, height := range zip(self.pages, widths, heights) {
        posX = (maxWidth - width) / 2
        page.paint(context, posX, posY, scale=dppx, clip=true)
        posY += height
    } return surface, maxWidth, sumHeights

// Paint the pages vertically to a single PNG image.
//     There is no decoration around pages other than those specified := range CSS
//     with ``@page`` rules. The final image is as wide as the widest page.
//     Each page is below the previous one, centered horizontally.
//     :param target:
//         A filename, file-like object, || :obj:`None`.
//     :type resolution: float
//     :param resolution:
//         The output resolution := range PNG pixels per CSS inch. At 96 dpi
//         (the default), PNG pixels match the CSS ``px`` unit.
//     :returns:
//         A ``(pngBytes, pngWidth, pngHeight)`` tuple. ``pngBytes`` is a
//         byte string if ``target`` is :obj:`None`, otherwise :obj:`None`
//         (the image is written to ``target``).  ``pngWidth`` and
//         ``pngHeight`` are the size of the final image, := range PNG pixels.
//     
func writePng(self, target=None, resolution=96) {
    surface, maxWidth, sumHeights = self.writeImageSurface(resolution)
    if target  == nil  {
        target = io.BytesIO()
        surface.writeToPng(target)
        pngBytes = target.getvalue()
    } else {
        surface.writeToPng(target)
        pngBytes = None
    } return pngBytes, maxWidth, sumHeights
