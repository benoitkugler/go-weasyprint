package goweasyprint

import (
	"bytes"
	"io"
	"log"
	"math"

	"github.com/benoitkugler/go-weasyprint/images"
	mt "github.com/benoitkugler/go-weasyprint/matrix"
	"github.com/benoitkugler/go-weasyprint/version"

	"github.com/benoitkugler/go-weasyprint/fonts"

	"github.com/benoitkugler/go-weasyprint/backend"
	bo "github.com/benoitkugler/go-weasyprint/boxes"
	"github.com/benoitkugler/go-weasyprint/layout"
	"github.com/benoitkugler/go-weasyprint/logger"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/style/tree"
	"github.com/benoitkugler/go-weasyprint/utils"
)

func toF(v pr.Dimension) float64 { return float64(v.Value) }

// Return the matrix for the CSS transforms on this box (possibly nil)
func getMatrix(box_ Box) *mt.Transform {
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
		originX := float64(box.BorderBoxX() + offsetX)
		originY := float64(box.BorderBoxY() + offsetY)

		var matrix mt.Transform
		matrix.Translate(originX, originY)
		for _, t := range trans {
			name, args := t.String, t.Dimensions
			// The length of args depends on `name`, see package validation for details.
			switch name {
			case "scale":
				sx, sy := toF(args[0]), toF(args[1])
				matrix.Scale(sx, sy)
			case "rotate":
				angle := toF(args[0])
				matrix.Rotate(angle)
			case "translate":
				translateX, translateY := args[0], args[1]
				matrix.Translate(
					float64(pr.ResoudPercentage(translateX.ToValue(), borderWidth).V()),
					float64(pr.ResoudPercentage(translateY.ToValue(), borderHeight).V()),
				)
			default:
				var leftMat mt.Transform
				switch name {
				case "skewx":
					leftMat = mt.New(1, 0, math.Tan(toF(args[0])), 1, 0, 0)
				case "skewy":
					leftMat = mt.New(1, math.Tan(toF(args[0])), 0, 1, 0, 0)
				case "matrix":
					leftMat = mt.New(toF(args[0]), toF(args[1]), toF(args[2]),
						toF(args[3]), toF(args[4]), toF(args[5]))
				default:
					log.Fatalf("unexpected name for CSS transform property : %s", name)
				}
				matrix = mt.Mul(leftMat, matrix)
			}
		}
		matrix.Translate(-originX, -originY)
		box.TransformationMatrix = &matrix
		return &matrix
	}
	return nil
}

// Apply a transformation matrix to an axis-aligned rectangle
// and return its axis-aligned bounding box as ``(x, y, width, height)``
func rectangleAabb(matrix mt.Transform, posX, posY, width, height float64) [4]float64 {
	x1, y1 := matrix.TransformPoint(posX, posY)
	x2, y2 := matrix.TransformPoint(posX+width, posY)
	x3, y3 := matrix.TransformPoint(posX, posY+height)
	x4, y4 := matrix.TransformPoint(posX+width, posY+height)
	boxX1 := utils.Mins(x1, x2, x3, x4)
	boxY1 := utils.Mins(y1, y2, y3, y4)
	boxX2 := utils.Maxs(x1, x2, x3, x4)
	boxY2 := utils.Maxs(y1, y2, y3, y4)
	return [4]float64{boxX1, boxY1, boxX2 - boxX1, boxY2 - boxY1}
}

type linkData struct {
	type_, target string
	rectangle     [4]float64
}

type bookmarkData struct {
	level    int
	label    string
	position [2]float64
	state    string
}

func gatherLinksAndBookmarks(box_ bo.Box, bookmarks *[]bookmarkData, links *[]linkData, anchors map[string][2]float64, matrix *mt.Transform) {
	transform := getMatrix(box_)
	if transform != nil {
		if matrix != nil {
			t := mt.Mul(*matrix, *transform)
			matrix = &t
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
	anchorName := string(box.Style.GetAnchor())
	hasBookmark := bookmarkLabel != "" && bookmarkLevel != 0
	// "link" is inherited but redundant on text boxes
	hasLink := !link.IsNone() && !bo.IsTextBox(box_)
	// In case of duplicate IDs, only the first is an anchor.
	_, inAnchors := anchors[anchorName]
	hasAnchor := anchorName != "" && !inAnchors
	isAttachment := box.IsAttachment

	if hasBookmark || hasLink || hasAnchor {
		posX, posY, width, height := box.HitArea().Unpack()
		if hasLink {
			linkType, target := link.Name, link.String
			if linkType == "external" && isAttachment {
				linkType = "attachment"
			}
			var linkS linkData
			if matrix != nil {
				linkS = linkData{type_: linkType, target: target,
					rectangle: rectangleAabb(*matrix, posX, posY, width, height)}
			} else {
				linkS = linkData{type_: linkType, target: target, rectangle: [4]float64{posX, posY, width, height}}
			}
			*links = append(*links, linkS)
		}
		if matrix != nil && (hasBookmark || hasAnchor) {
			posX, posY = matrix.TransformPoint(posX, posY)
		}
		if hasBookmark {
			*bookmarks = append(*bookmarks, bookmarkData{level: bookmarkLevel, label: bookmarkLabel,
				position: [2]float64{posX, posY}, state: state})
		}
		if hasAnchor {
			anchors[anchorName] = [2]float64{posX, posY}
		}
	}

	for _, child := range box.AllChildren() {
		gatherLinksAndBookmarks(child, bookmarks, links, anchors, matrix)
	}
}

// Represents a single rendered page.
// Should be obtained from `Document.pages` but not
// instantiated directly.
type Page struct {
	// The page width, including margins, in CSS pixels.
	width float64

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

	pageBox       *bo.PageBox
	enableHinting bool
}

// enableHinting=false
func NewPage(pageBox *bo.PageBox, enableHinting bool) Page {
	d := Page{}
	d.width = float64(pageBox.MarginWidth())
	d.height = float64(pageBox.MarginHeight())

	d.bleed = bleedData{
		Top:    float64(pageBox.Style.GetBleedTop().Value),
		Right:  float64(pageBox.Style.GetBleedRight().Value),
		Bottom: float64(pageBox.Style.GetBleedBottom().Value),
		Left:   float64(pageBox.Style.GetBleedLeft().Value),
	}
	d.anchors = map[string][2]float64{}

	gatherLinksAndBookmarks(
		pageBox, &d.bookmarks, &d.links, d.anchors, nil)
	d.pageBox = pageBox
	d.enableHinting = enableHinting
	return d
}

// Paint the page in cairo, on any type of surface (leftX=0, topY=0, scale=1, clip=false).
// leftX is the X coordinate of the left of the page, in cairo user units.
// topY is the Y coordinate of the top of the page, in cairo user units.
// scale is the Zoom scale in cairo user units per CSS pixel.
// clip : whether to clip/cut content outside the page. If false, content can overflow.
func (d Page) paint(cairoContext Drawer, leftX, topY, scale float64, clip bool) {
	// with stacked(cairoContext) {
	if d.enableHinting {
		leftX, topY = cairoContext.UserToDevice(leftX, topY)
		// Hint in device space
		leftX = math.Round(leftX)
		topY = math.Round(topY)
		leftX, topY = cairoContext.DeviceToUser(leftX, topY)
	}
	// Make (0, 0) the top-left corner:
	cairoContext.Translate(leftX, topY)
	// Make user units CSS pixels:
	cairoContext.Scale(scale, scale)
	if clip {
		width := d.width
		height := d.height
		if d.enableHinting {
			width, height = cairoContext.UserToDeviceDistance(width, height)
			// Hint in device space
			width = math.Ceil(width)
			height = math.Ceil(height)
			width, height = cairoContext.DeviceToUserDistance(width, height)
		}
		cairoContext.Rectangle(0, 0, width, height)
		cairoContext.Clip()
	}
	drawPage(d.pageBox, cairoContext, d.enableHinting)
}

// A rendered document ready to be painted on a cairo context.

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
	fontConfig *fonts.FontConfiguration
}

// presentationalHints=false, fontConfig=None
func buildLayoutContext(html tree.HTML, stylesheets []tree.CSS, enableHinting,
	presentationalHints bool, fontConfig *fonts.FontConfiguration) *layout.LayoutContext {

	if fontConfig == nil {
		fontConfig = fonts.NewFontConfiguration()
	}
	targetCollector := tree.NewTargetCollector()
	var (
		pageRules       []tree.PageRule
		userStylesheets []tree.CSS
	)
	for _, css := range stylesheets {
		// if css.Matcher == nil {
		// 	css = tree.NewCSS(css, mediaType=html.mediaType, fontConfig=fontConfig)
		// }
		userStylesheets = append(userStylesheets, css)
	}
	styleFor := tree.GetAllComputedStyles(html, userStylesheets, presentationalHints, fontConfig,
		&pageRules, &targetCollector)
	cache := make(map[string]images.Image)
	getImageFromUri := func(url, forcedMimeType string) images.Image {
		return images.GetImageFromUri(cache, html.UrlFetcher, url, forcedMimeType)
	}
	logger.ProgressLogger.Println("Step 4 - Creating formatting structure")
	context := layout.NewLayoutContext(enableHinting, *styleFor, getImageFromUri, fontConfig, &targetCollector)
	return context
}

// presentationalHints=false, fontConfig=None
func render(html tree.HTML, stylesheets []tree.CSS, enableHinting,
	presentationalHints bool, fontConfig *fonts.FontConfiguration) Document {

	if fontConfig == nil {
		fontConfig = fonts.NewFontConfiguration()
	}

	context := buildLayoutContext(html, stylesheets, enableHinting, presentationalHints, fontConfig)

	rootBox := bo.BuildFormattingStructure(html.Root, context.StyleFor, context.GetImageFromUri,
		html.BaseUrl, context.TargetCollector)

	pageBoxes := layout.LayoutDocument(html, rootBox, context, -1)
	pages := make([]Page, len(pageBoxes))
	for i, pageBox := range pageBoxes {
		pages[i] = NewPage(pageBox, enableHinting)
	}
	return Document{pages: pages, metadata: html.GetMetadata(), urlFetcher: html.UrlFetcher, fontConfig: fontConfig}
}

// Take a subset of the pages.
//
// Examples:
// Write two PDF files for odd-numbered and even-numbered pages:
//     document.Copy(document.pages[::2]).writePdf("oddPages.pdf")
//     document.Copy(document.pages[1::2]).writePdf("evenPages.pdf")
// Combine multiple documents into one PDF file, using metadata from the first:
//		var allPages []Page
// 		for _, doc := range documents {
//		 	for _, p := range doc.pages {
//		 		allPages = append(allPages, p)
//		 	}
//		 }
//		documents[0].Copy(allPages).writePdf("combined.pdf")
func (d Document) Copy(pages []Page, all bool) Document {
	if all {
		pages = d.pages
	}
	return Document{pages: pages, metadata: d.metadata, urlFetcher: d.urlFetcher,
		fontConfig: d.fontConfig}
}

// Resolve internal hyperlinks.
// Links to a missing anchor are removed with a warning.
// If multiple anchors have the same name, the first one is used.
// Returns lists (one per page) like :attr:`Page.links`,
// except that ``target`` for internal hyperlinks is
// ``(pageNumber, x, y)`` instead of an anchor name.
// The page number is a 0-based index into the :attr:`pages` list,
// and ``x, y`` have been scaled (origin is at the top-left of the page).
func (d Document) resolveLinks(scale float64) ([][]linkData, [][]backend.Anchor) {
	anchors := utils.NewSet()
	pagedAnchors := make([][]backend.Anchor, len(d.pages))
	for i, page := range d.pages {
		var current []backend.Anchor
		for anchorName, pos := range page.anchors {
			if !anchors.Has(anchorName) {
				pos[0] *= scale
				pos[1] *= scale
				current = append(current, backend.Anchor{Name: anchorName, Pos: pos})
				anchors.Add(anchorName)
			}
		}
		pagedAnchors[i] = current
	}
	pagedLinks := make([][]linkData, len(d.pages))
	for i, page := range d.pages {
		var pageLinks []linkData
		for _, link := range page.links {
			// linkType, anchorName, rectangle = link
			if link.type_ == "internal" {
				if !anchors.Has(link.target) {
					log.Printf("No anchor #%s for internal URI reference\n", link.target)
				} else {
					pageLinks = append(pageLinks, link)
				}
			} else {
				// External link
				pageLinks = append(pageLinks, link)
			}
		}
		pagedLinks[i] = pageLinks
	}
	return pagedLinks, pagedAnchors
}

type target struct {
	pageNumber int
	pos        [2]float64
}

type bookmarkSubtree struct {
	label    string
	target   target
	children []bookmarkSubtree
	state    string
}

// Make a tree of all bookmarks in the document.
func (d Document) makeBookmarkTree() []bookmarkSubtree {
	var root []bookmarkSubtree
	// At one point in the document, for each "output" depth, how much
	// to add to get the source level (CSS values of bookmark-level).
	// E.g. with <h1> then <h3>, levelShifts == [0, 1]
	// 1 means that <h3> has depth 3 - 1 = 2 in the output.
	var skippedLevels []int
	lastByDepth := [][]bookmarkSubtree{root}
	previousLevel := 0
	for pageNumber, page := range d.pages {
		for _, bk := range page.bookmarks {
			level, label, pos, state := bk.level, bk.label, bk.position, bk.state
			if level > previousLevel {
				// Example: if the previous bookmark is a <h2>, the next
				// depth "should" be for <h3>. If now we get a <h6> we’re
				// skipping two levels: append 6 - 3 - 1 = 2
				skippedLevels = append(skippedLevels, level-previousLevel-1)
			} else {
				temp := level
				for temp < previousLevel {
					pop := skippedLevels[len(skippedLevels)-1]
					skippedLevels = skippedLevels[:len(skippedLevels)-1]
					temp += 1 + pop
				}
				if temp > previousLevel {
					// We remove too many "skips", add some back:
					skippedLevels = append(skippedLevels, temp-previousLevel-1)
				}
			}
			sum := 0
			for _, l := range skippedLevels {
				sum += l
			}
			previousLevel = level
			depth := level - sum
			if depth != len(skippedLevels) || depth < 1 {
				log.Fatalf("expected depth >= 1 and depth == len(skippedLevels) got %d", depth)
			}
			var children []bookmarkSubtree
			subtree := bookmarkSubtree{label: label, target: target{pageNumber: pageNumber, pos: pos}, children: children, state: state}
			lastByDepth[depth-1] = append(lastByDepth[depth-1], subtree)
			lastByDepth = lastByDepth[:depth]
			lastByDepth = append(lastByDepth, children)
		}
	}
	return root
}

// Include hyperlinks in current PDF page.
func (d Document) addHyperlinks(links []linkData, anchorsId map[string]int, context Drawer, scale float64) {
	// TODO: Instead of using rects, we could use the drawing rectangles
	// defined by cairo when drawing targets. This would give a feeling
	// similiar to what browsers do with links that span multiple lines.
	for _, link := range links {
		linkType, linkTarget, rectangle := link.type_, link.target, link.rectangle
		x, y, w, h := rectangle[0]*scale, rectangle[1]*scale, rectangle[2]*scale, rectangle[3]*scale
		if linkType == "external" {
			context.AddExternalLink(x, y, w, h, linkTarget)
		} else if linkType == "internal" {
			context.AddInternalLink(x, y, w, h, anchorsId[linkTarget])
		} else if linkType == "attachment" {
			// actual embedding has be done previously
			context.AddFileAnnotation(x, y, w, h, linkTarget)
		}
	}
}

func fetchAttachment(attachmentUrl string, urlFetcher utils.UrlFetcher) backend.Attachment {
	// Attachments from document links like <link> or <a> can only be URLs.
	tmp, err := tree.SelectSource(tree.InputUrl(attachmentUrl), "", urlFetcher, false)
	if err != nil {
		log.Printf("Failed to load attachment at url %s: %s\n", attachmentUrl, err)
		return backend.Attachment{}
	}
	source, url := tmp.Content, tmp.BaseUrl
	// TODO: Use the result object from a URL fetch operation to provide more
	// details on the possible filename
	filename := getFilenameFromResult(url)
	return backend.Attachment{Content: source, Title: filename}
}

func embedFileAnnotations(pagedLinks [][]linkData, urlFetcher utils.UrlFetcher, context Drawer) {
	// A single link can be split in multiple regions. We don't want to embed
	// a file multiple times of course, so keep a reference to every embedded
	// URL and reuse the object number.
	for _, rl := range pagedLinks {
		for _, link := range rl {
			if link.type_ == "attachment" {
				a := fetchAttachment(link.target, urlFetcher)
				if len(a.Content) != 0 {
					context.EmbedFile(link.target, a)
				}
			}
		}
	}
}

func setMediaBoxes(context backend.Drawer, bleed bleedData) {
	// Add bleed box
	left, top, right, bottom := context.GetMediaBox()

	trimLeft := left + bleed.Left
	trimTop := top + bleed.Top
	trimRight := right - bleed.Right
	trimBottom := bottom - bleed.Bottom

	// Arbitrarly set PDF BleedBox between CSS bleed box (PDF MediaBox) and
	// CSS page box (PDF TrimBox), at most 12 px from the TrimBox.
	bleedLeft := trimLeft - math.Min(12, bleed.Left)
	bleedTop := trimTop - math.Min(12, bleed.Top)
	bleedRight := trimRight + math.Min(12, bleed.Right)
	bleedBottom := trimBottom + math.Min(12, bleed.Bottom)

	context.SetTrimBox(trimLeft, trimTop, trimRight, trimBottom)
	context.SetBleedBox(bleedLeft, bleedTop, bleedRight, bleedBottom)
}

// Paint the pages in a PDF file, with meta-data (zoom=1, attachments=nil).
// PDF files written directly by cairo do not have meta-data such as
// bookmarks/outlines and hyperlinks.
//
// The zoom factor in PDF units per CSS units.  **Warning**:
// All CSS units are affected, including physical units like
// ``cm`` and named sizes like ``A4``.  For values other than
// 1, the physical CSS units will thus be "wrong".
// :type attachments: list
// :param attachments: A list of additional file attachments for the
//     generated PDF document || :obj:`None`. The list"s elements are
//     :class:`Attachment` objects, filenames, URLs || file-like objects.
func (d Document) WritePdf(self, target io.Writer, zoom float64, attachments []utils.Attachment) {
	// 0.75 = 72 PDF point (cairo units) per inch / 96 CSS pixel per inch
	scale := zoom * 0.75
	// Use an in-memory buffer, as we will need to seek for
	// metadata. Directly using the target when possible doesn't
	// significantly save time and memory use.
	var fileObj bytes.Buffer
	// (1, 1) is overridden by .setSize() below.
	// var surface backend.Surface = cairo.PDFSurface(fileObj, 1, 1)
	var context Drawer = cairo.Context(surface)

	pagedLinks, pagedAnchors := d.resolveLinks(scale)

	logger.ProgressLogger.Println("Step 6 alpha - Embedding attachments")
	// embedded files
	var as []backend.Attachment
	for _, a := range append(d.metadata.Attachments, attachments...) {
		t := fetchAttachment(a.Url, d.urlFetcher)
		if len(t.Content) != 0 {
			as = append(as, t)
		}
	}
	context.SetAttachments(as)

	embedFileAnnotations(pagedLinks, d.urlFetcher, context)

	logger.ProgressLogger.Println("Step 6 - Drawing")

	anchorsId := context.CreateAnchors(pagedAnchors)
	for i, page := range d.pages {
		context.SetSize(
			math.Floor(scale*(page.width+float64(page.bleed.Left)+float64(page.bleed.Right))),
			math.Floor(scale*(page.height+float64(page.bleed.Top)+float64(page.bleed.Bottom))),
		)
		// with stacked(context) {
		context.Translate(float64(page.bleed.Left)*scale, float64(page.bleed.Top)*scale)
		page.paint(context, 0, 1, scale, false)
		d.addHyperlinks(pagedLinks[i], anchorsId, context, scale)
		setMediaBoxes(context, page.bleed)
		context.ShowPage()
		// }
	}

	logger.ProgressLogger.Println("Step 7 - Adding PDF metadata")

	// Set document information
	context.SetTitle(d.metadata.Title)
	context.SetDescription(d.metadata.Description)
	context.SetCreator(d.metadata.Generator)
	context.SetAuthors(d.metadata.Authors)
	context.SetKeywords(d.metadata.Keywords)
	context.SetProducer(version.VersionString)
	context.SetDateCreation(d.metadata.Created)
	context.SetDateModification(d.metadata.Modified)

	// Set bookmarks
	bookmarks := d.makeBookmarkTree()
	levels := make([]int, len(bookmarks)) // 0 is the root level
	for len(bookmarks) != 0 {
		bookmark := bookmarks[0]
		bookmarks = bookmarks[1:]
		title := bookmark.label
		page, y := bookmark.target.pageNumber, bookmark.target.pos[1]
		children := bookmark.children
		// state := bookmark.state

		level := levels[len(levels)-1]
		levels = levels[:len(levels)-1]
		context.AddBookmark(level, title, page+1, y*scale)
		// preparing children bookmarks
		childLevel := level + 1
		for i := 0; i < len(children); i += 1 {
			levels = append(levels, childLevel)
		}
		bookmarks = append(children, bookmarks...)
	}
	context.Finish() //FIXME: à garder ?

	// Add extra PDF metadata: attachments, embedded files

	// Write extra PDF metadata only when there is a least one from:
	// - attachments inmetadata
	// - attachments as function parameters
	// - attachments as PDF links
	// - bleed boxes
	anyBleed := false
	for _, page := range d.pages {
		if (page.bleed != bleedData{}) {
			anyBleed = true
			break
		}
	}
	condition := anyBleed
	if condition {
		writePdfMetadata(&fileObj, scale, d.urlFetcher, d.pages)
	}

}
