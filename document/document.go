// This package implements the high level parts of
// document generation, but is still backend independant.
package document

import (
	"log"
	"math"
	"net/url"
	"path"

	"github.com/benoitkugler/go-weasyprint/images"
	mt "github.com/benoitkugler/go-weasyprint/matrix"

	"github.com/benoitkugler/go-weasyprint/fonts"

	"github.com/benoitkugler/go-weasyprint/backend"
	bo "github.com/benoitkugler/go-weasyprint/boxes"
	"github.com/benoitkugler/go-weasyprint/layout"
	"github.com/benoitkugler/go-weasyprint/logger"
	"github.com/benoitkugler/go-weasyprint/style/parser"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/style/tree"
	"github.com/benoitkugler/go-weasyprint/utils"
)

type Drawer = backend.Drawer
type Color = parser.RGBA
type Box = bo.Box

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

type Link struct {
	Type, Target string
	rectangle    [4]float64
}

type bookmarkData struct {
	level    int
	label    string
	position [2]float64
	state    string
}

func gatherLinksAndBookmarks(box_ bo.Box, bookmarks *[]bookmarkData, links *[]Link, anchors map[string][2]float64, matrix *mt.Transform) {
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
			var linkS Link
			if matrix != nil {
				linkS = Link{Type: linkType, Target: target,
					rectangle: rectangleAabb(*matrix, posX, posY, width, height)}
			} else {
				linkS = Link{Type: linkType, Target: target, rectangle: [4]float64{posX, posY, width, height}}
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
	Width float64

	// The page height, including margins, in CSS pixels.
	Height float64

	// The page bleed widths as a `dict` with `"top"`, `"right"`,
	// `"bottom"` and `"left"` as keys, and values in CSS pixels.
	Bleed Bleed

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
	links []Link

	// The `dict` mapping each anchor name to its target, an
	// `(x, y)` point in CSS pixels from the top-left of the page.
	anchors map[string][2]float64

	pageBox       *bo.PageBox
	enableHinting bool
}

// enableHinting=false
func NewPage(pageBox *bo.PageBox, enableHinting bool) Page {
	d := Page{}
	d.Width = float64(pageBox.MarginWidth())
	d.Height = float64(pageBox.MarginHeight())

	d.Bleed = Bleed{
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
func (d Page) Paint(cairoContext Drawer, leftX, topY, scale float64, clip bool) {
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
		width := d.Width
		height := d.Height
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
	Pages []Page
	// A `DocumentMetadata` object.
	// Contains information that does not belong to a specific page
	// but to the whole document.
	Metadata utils.DocumentMetadata
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
func Render(html tree.HTML, stylesheets []tree.CSS, enableHinting,
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
	return Document{Pages: pages, Metadata: html.GetMetadata(), urlFetcher: html.UrlFetcher, fontConfig: fontConfig}
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
		pages = d.Pages
	}
	return Document{Pages: pages, Metadata: d.Metadata, urlFetcher: d.urlFetcher,
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
func (d Document) ResolveLinks(scale float64) ([][]Link, [][]backend.Anchor) {
	anchors := utils.NewSet()
	pagedAnchors := make([][]backend.Anchor, len(d.Pages))
	for i, page := range d.Pages {
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
	pagedLinks := make([][]Link, len(d.Pages))
	for i, page := range d.Pages {
		var pageLinks []Link
		for _, link := range page.links {
			// linkType, anchorName, rectangle = link
			if link.Type == "internal" {
				if !anchors.Has(link.Target) {
					log.Printf("No anchor #%s for internal URI reference\n", link.Target)
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
func (d Document) MakeBookmarkTree() []bookmarkSubtree {
	var root []bookmarkSubtree
	// At one point in the document, for each "output" depth, how much
	// to add to get the source level (CSS values of bookmark-level).
	// E.g. with <h1> then <h3>, levelShifts == [0, 1]
	// 1 means that <h3> has depth 3 - 1 = 2 in the output.
	var skippedLevels []int
	lastByDepth := [][]bookmarkSubtree{root}
	previousLevel := 0
	for pageNumber, page := range d.Pages {
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
func (d Document) AddHyperlinks(links []Link, anchorsId map[string]int, context Drawer, scale float64) {
	// TODO: Instead of using rects, we could use the drawing rectangles
	// defined by cairo when drawing targets. This would give a feeling
	// similiar to what browsers do with links that span multiple lines.
	for _, link := range links {
		linkType, linkTarget, rectangle := link.Type, link.Target, link.rectangle
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

func (d Document) FetchAttachment(attachmentUrl string) backend.Attachment {
	// Attachments from document links like <link> or <a> can only be URLs.
	tmp, err := utils.SelectSource(utils.InputUrl(attachmentUrl), "", d.urlFetcher, false)
	if err != nil {
		log.Printf("Failed to load attachment at url %s: %s\n", attachmentUrl, err)
		return backend.Attachment{}
	}
	source, baseurl := tmp.Content, tmp.BaseUrl
	// TODO: Use the result object from a URL fetch operation to provide more
	// details on the possible filename
	filename := getFilenameFromResult(baseurl)
	return backend.Attachment{Content: source, Title: filename}
}
// Derive a filename from a fetched resource.
// This is either the filename returned by the URL fetcher, the last URL path
// component or a synthetic name if the URL has no path.
func getFilenameFromResult(rawurl string) string {
	var filename string

	// The URL path likely contains a filename, which is a good second guess
	if rawurl != "" {
		u, err := url.Parse(rawurl)
		if err == nil {
			if u.Scheme != "data" {
				filename = path.Base(u.Path)
			}
		}
	}

	if filename == "" {
		// The URL lacks a path altogether. Use a synthetic name.

		// Using guessExtension is a great idea, but sadly the extension is
		// probably random, depending on the alignment of the stars, which car
		// you're driving and which software has been installed on your machine.

		// Unfortuneatly this isn't even imdepodent on one machine, because the
		// extension can depend on PYTHONHASHSEED if mimetypes has multiple
		// extensions to offer
		extension := ".bin"
		filename = "attachment" + extension
	} else {
		filename = utils.Unquote(filename)
	}

	return filename
}