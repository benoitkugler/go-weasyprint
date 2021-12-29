// Package pdf implements the backend needed to draw a document, using github.com/benoitkugler/pdf.
package pdf

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/benoitkugler/pdf/model"
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/pango"
	"github.com/benoitkugler/webrender/backend"
)

// type graphicState struct {
// 	clipNest      int // each ClipXXX increment clipNest by 1
// 	transformNest int
// 	alpha         float64
// 	r, g, b       int
// 	fillRule      int
// }

// func newGraphicState(f *gofpdf.Fpdf) graphicState {
// 	out := graphicState{}
// 	out.alpha, _ = f.GetAlpha()
// 	out.r, out.g, out.b = f.GetFillColor()
// 	return out
// }

var (
	_ backend.Document = (*Output)(nil)
	_ backend.Page     = (*outputPage)(nil)
)

type cache struct {
	// global shared cache for image content
	images map[int]*model.XObjectImage

	// global shared cache for fonts
	fonts map[pango.Font]pdfFont

	// global shared cache for font files
	// the same face may be used at different size
	// and we don't want to duplicate the font file
	fontFiles map[fonts.Face]*model.FontFile
}

// Output implements backend.Output
type Output struct {
	// global map for files embedded in the PDF
	// and used in file annotations
	embeddedFiles map[string]*model.FileSpec

	cache cache

	document model.Document

	// temporary content, will be copied in the document (see `finalize`)
	pages []*outputPage
}

func NewOutput() *Output {
	out := Output{
		embeddedFiles: make(map[string]*model.FileSpec),
		cache: cache{
			images:    make(map[int]*model.XObjectImage),
			fonts:     make(map[pango.Font]pdfFont),
			fontFiles: make(map[fonts.Face]*model.FontFile),
		},
	}
	return &out
}

func (c *Output) AddPage(left, top, right, bottom fl) backend.Page {
	out := newContextPage(left, top, right, bottom, c.embeddedFiles, c.cache)
	c.pages = append(c.pages, out)
	return out
}

func (s *Output) SetTitle(title string) {
	s.document.Trailer.Info.Title = title
}

func (s *Output) SetDescription(description string) {
	s.document.Trailer.Info.Subject = description
}

func (s *Output) SetCreator(creator string) {
	s.document.Trailer.Info.Creator = creator
}

func (s *Output) SetAuthors(authors []string) {
	s.document.Trailer.Info.Author = strings.Join(authors, ", ")
}

func (s *Output) SetKeywords(keywords []string) {
	s.document.Trailer.Info.Keywords = strings.Join(keywords, ", ")
}

func (s *Output) SetProducer(producer string) {
	s.document.Trailer.Info.Producer = producer
}

func (s *Output) SetDateCreation(d time.Time) {
	s.document.Trailer.Info.CreationDate = d
}

func (s *Output) SetDateModification(d time.Time) {
	s.document.Trailer.Info.ModDate = d
}

func (c *Output) CreateAnchors(anchors [][]backend.Anchor) {
	// pages have been processed, meaning that len(anchors) == len(c.pages)

	var names []model.NameToDest
	for i, l := range anchors {
		page := c.pages[i]
		for _, anchor := range l {
			names = append(names, model.NameToDest{
				Name: model.DestinationString(anchor.Name),
				Destination: model.DestinationExplicitIntern{
					Page: &page.page,
					Location: model.DestinationLocationXYZ{
						Left: model.ObjFloat(anchor.X),
						Top:  model.ObjFloat(anchor.Y),
					},
				},
			})
		}
	}

	sort.Slice(names, func(i, j int) bool { return names[i].Name < names[j].Name })

	c.document.Catalog.Names.Dests.Names = names
}

// embedded files

func newFileSpec(a backend.Attachment) *model.FileSpec {
	stream := model.NewCompressedStream(a.Content)
	fs := &model.FileSpec{
		UF:   a.Title,
		Desc: a.Description,
		EF: &model.EmbeddedFileStream{
			Stream: stream,
		},
	}
	fs.EF.Params.SetChecksumAndSize(a.Content)
	return fs
}

// Add global attachments to the file, which are compressed using FlateDecode filter
func (c *Output) SetAttachments(as []backend.Attachment) {
	var files model.EmbeddedFileTree
	for i, a := range as {
		fs := newFileSpec(a)
		files = append(files, model.NameToFile{
			Name:     fmt.Sprintf("attachement_%d", i),
			FileSpec: fs,
		})
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })

	c.document.Catalog.Names.EmbeddedFiles = files
}

// Embed a file. Calling this method twice with the same id
// won't embed the content twice.
func (c *Output) EmbedFile(fileID string, a backend.Attachment) {
	ptr := c.embeddedFiles[fileID] // cache the attachment by id
	if ptr != nil {
		return
	}

	c.embeddedFiles[fileID] = newFileSpec(a)
}

func bookmarksToOutline(root []backend.BookmarkNode, pages []*outputPage) *model.Outline {
	var nodeToItem func(node backend.BookmarkNode, parent model.OutlineNode) *model.OutlineItem

	nodesToItem := func(nodes []backend.BookmarkNode, parent model.OutlineNode) *model.OutlineItem {
		var first, lastChild *model.OutlineItem
		for i, node := range nodes {
			item := nodeToItem(node, parent)
			if i == 0 {
				first = item
			}
			if lastChild != nil {
				lastChild.Next = item
			}
			lastChild = item
		}
		return first
	}

	nodeToItem = func(node backend.BookmarkNode, parent model.OutlineNode) *model.OutlineItem {
		out := &model.OutlineItem{
			Parent: parent,
			Title:  node.Label,
			Open:   node.Open,
			Dest: model.DestinationExplicitIntern{
				Page: &pages[node.PageIndex].page,
				Location: model.DestinationLocationXYZ{
					Left: model.ObjFloat(node.X),
					Top:  model.ObjFloat(node.Y),
				},
			},
		}
		out.First = nodesToItem(node.Children, out)
		return out
	}
	var outline model.Outline
	outline.First = nodesToItem(root, &outline)
	return &outline
}

func (c *Output) SetBookmarks(root []backend.BookmarkNode) {
	c.document.Catalog.Outlines = bookmarksToOutline(root, c.pages)
}

// Finalize setup and returns the final document
func (c *Output) Finalize() model.Document {
	pages := make([]model.PageNode, len(c.pages))
	for i, p := range c.pages {
		p.finalize()
		pages[i] = &p.page
	}
	c.document.Catalog.Pages = model.PageTree{
		Kids: pages,
	}

	// fonts
	c.writeFonts()

	return c.document
}
