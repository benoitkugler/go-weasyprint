package pdf

// import (
// 	"bytes"
// 	"io"
// 	"log"
// 	"math"
// 	"net/url"
// 	"path"

// 	"github.com/benoitkugler/go-weasyprint/backend"
// 	"github.com/benoitkugler/go-weasyprint/document"
// 	"github.com/benoitkugler/go-weasyprint/logger"
// 	"github.com/benoitkugler/go-weasyprint/style/tree"
// 	"github.com/benoitkugler/go-weasyprint/utils"
// 	"github.com/benoitkugler/go-weasyprint/version"
// )

// // Paint the pages in a PDF file, with meta-data (zoom=1, attachments=nil).
// //
// // The zoom factor in PDF units per CSS units.  **Warning**:
// // All CSS units are affected, including physical units like
// // ``cm`` and named sizes like ``A4``.  For values other than
// // 1, the physical CSS units will thus be "wrong".
// // `attachments` is a list of additional file attachments for the
// //  generated PDF document.
// func WritePDF(d document.Document, target io.Writer, zoom float64, attachments []utils.Attachment) error {
// 	// 0.75 = 72 PDF point (cairo units) per inch / 96 CSS pixel per inch
// 	scale := zoom * 0.75

// 	context := NewContext()

// 	pagedLinks, pagedAnchors := d.ResolveLinks(scale)

// 	logger.ProgressLogger.Println("Step 6 alpha - Embedding attachments")
// 	// embedded files
// 	var as []backend.Attachment
// 	for _, a := range append(d.Metadata.Attachments, attachments...) {
// 		t :=  d.FetchAttachment(a.Url)
// 		if len(t.Content) != 0 {
// 			as = append(as, t)
// 		}
// 	}
// 	context.SetAttachments(as)

// 	embedFileAnnotations(d, pagedLinks, context)

// 	logger.ProgressLogger.Println("Step 6 - Drawing")

// 	anchorsId := context.CreateAnchors(pagedAnchors)
// 	for i, page := range d.Pages {
// 		context.SetSize(
// 			math.Floor(scale*(page.Width+float64(page.Bleed.Left)+float64(page.Bleed.Right))),
// 			math.Floor(scale*(page.Height+float64(page.Bleed.Top)+float64(page.Bleed.Bottom))),
// 		)
// 		// with stacked(context) {
// 		context.Translate(float64(page.Bleed.Left)*scale, float64(page.Bleed.Top)*scale)
// 		page.Paint(context, 0, 1, scale, false)
// 		d.AddHyperlinks(pagedLinks[i], anchorsId, context, scale)
// 		// - bleed boxes
// 		setMediaBoxes(context, page.Bleed)
// 		context.ShowPage()
// 		// }
// 	}

// 	logger.ProgressLogger.Println("Step 7 - Adding PDF metadata")

// 	// Set document information
// 	context.SetTitle(d.Metadata.Title)
// 	context.SetDescription(d.Metadata.Description)
// 	context.SetCreator(d.Metadata.Generator)
// 	context.SetAuthors(d.Metadata.Authors)
// 	context.SetKeywords(d.Metadata.Keywords)
// 	context.SetProducer(version.VersionString)
// 	context.SetDateCreation(d.Metadata.Created)
// 	context.SetDateModification(d.Metadata.Modified)

// 	// Set bookmarks
// 	bookmarks := d.MakeBookmarkTree()
// 	levels := make([]int, len(bookmarks)) // 0 is the root level
// 	for len(bookmarks) != 0 {
// 		bookmark := bookmarks[0]
// 		bookmarks = bookmarks[1:]
// 		title := bookmark.label
// 		page, y := bookmark.target.pageNumber, bookmark.target.pos[1]
// 		children := bookmark.children
// 		// state := bookmark.state

// 		level := levels[len(levels)-1]
// 		levels = levels[:len(levels)-1]
// 		context.AddBookmark(level, title, page+1, y*scale)
// 		// preparing children bookmarks
// 		childLevel := level + 1
// 		for i := 0; i < len(children); i += 1 {
// 			levels = append(levels, childLevel)
// 		}
// 		bookmarks = append(children, bookmarks...)
// 	}
// 	context.Finish() //FIXME: à garder ?
// }

// func embedFileAnnotations(d document.Document, pagedLinks [][]document.Link, context backend.Drawer) {
// 	// A single link can be split in multiple regions. We don't want to embed
// 	// a file multiple times of course, so keep a reference to every embedded
// 	// URL and reuse the object number.
// 	for _, rl := range pagedLinks {
// 		for _, link := range rl {
// 			if link.Type == "attachment" {
// 				a := d.FetchAttachment(link.Target)
// 				if len(a.Content) != 0 {
// 					context.EmbedFile(link.Target, a)
// 				}
// 			}
// 		}
// 	}
// }

// func setMediaBoxes(context backend.Drawer, bleed document.Bleed) {
// 	// Add bleed box
// 	left, top, right, bottom := context.GetMediaBox()

// 	trimLeft := left + bleed.Left
// 	trimTop := top + bleed.Top
// 	trimRight := right - bleed.Right
// 	trimBottom := bottom - bleed.Bottom

// 	// Arbitrarly set PDF BleedBox between CSS bleed box (PDF MediaBox) and
// 	// CSS page box (PDF TrimBox), at most 12 px from the TrimBox.
// 	bleedLeft := trimLeft - math.Min(12, bleed.Left)
// 	bleedTop := trimTop - math.Min(12, bleed.Top)
// 	bleedRight := trimRight + math.Min(12, bleed.Right)
// 	bleedBottom := trimBottom + math.Min(12, bleed.Bottom)

// 	context.SetTrimBox(trimLeft, trimTop, trimRight, trimBottom)
// 	context.SetBleedBox(bleedLeft, bleedTop, bleedRight, bleedBottom)
// }
