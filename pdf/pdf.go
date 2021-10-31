// Package pdf implements the backend needed to draw a document, using github.com/benoitkugler/pdf.
package pdf

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/benoitkugler/go-weasyprint/backend"
	"github.com/benoitkugler/pdf/model"
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
	_ backend.Output     = (*Context)(nil)
	_ backend.OutputPage = (*contextPage)(nil)
)

// Context implements backend.Output
type Context struct {
	// global map for files embedded in the PDF
	// and used in file annotations
	embeddedFiles map[string]*model.FileSpec

	document model.Document

	// temporary content, will be copied in the document (see Finalize)
	pages []*model.PageObject

	// f *gofpdf.Fpdf

	// fileAnnotationsMap map[string]*gofpdf.Attachment

	// matrixToPdf, matrixToPdfInv matrix.Transform

	// fillRule int
}

func NewContext() *Context {
	out := Context{
		embeddedFiles: make(map[string]*model.FileSpec),
	}
	return &out
}

func (c *Context) AddPage(left, top, right, bottom fl) backend.OutputPage {
	out := newContextPage(left, top, right, bottom, c.embeddedFiles)
	c.pages = append(c.pages, &out.page)
	return out
}

func (s *Context) SetTitle(title string) {
	s.document.Trailer.Info.Title = title
}

func (s *Context) SetDescription(description string) {
	s.document.Trailer.Info.Subject = description
}

func (s *Context) SetCreator(creator string) {
	s.document.Trailer.Info.Creator = creator
}

func (s *Context) SetAuthors(authors []string) {
	s.document.Trailer.Info.Keywords = strings.Join(authors, ", ")
}

func (s *Context) SetKeywords(keywords []string) {
	s.document.Trailer.Info.Keywords = strings.Join(keywords, ", ")
}

func (s *Context) SetProducer(producer string) {
	s.document.Trailer.Info.Producer = producer
}

func (s *Context) SetDateCreation(d time.Time) {
	s.document.Trailer.Info.CreationDate = d
}

func (s *Context) SetDateModification(d time.Time) {
	s.document.Trailer.Info.ModDate = d
}

func (c *Context) CreateAnchors(anchors [][]backend.Anchor) {
	// pages have been processed, meaning that len(anchors) == len(c.pages)

	var names []model.NameToDest
	for i, l := range anchors {
		page := c.pages[i]
		for _, anchor := range l {
			names = append(names, model.NameToDest{
				Name: model.DestinationString(anchor.Name),
				Destination: model.DestinationExplicitIntern{
					Page: page,
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
	stream, err := model.NewStream(a.Content, model.Filter{Name: model.Flate})
	if err != nil {
		log.Printf("failed to compress attachement %s: %s", a.Title, err)
		// default to non compressed format
		stream = model.Stream{Content: a.Content}
	}
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
func (c *Context) SetAttachments(as []backend.Attachment) {
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
func (c *Context) EmbedFile(fileID string, a backend.Attachment) {
	ptr := c.embeddedFiles[fileID] // cache the attachment by id
	if ptr != nil {
		return
	}

	c.embeddedFiles[fileID] = newFileSpec(a)
}

func bookmarksToOutline(root []backend.BookmarkNode, pages []*model.PageObject) *model.Outline {
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
				Page: pages[node.PageIndex],
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

func (c *Context) SetBookmarks(root []backend.BookmarkNode) {
	c.document.Catalog.Outlines = bookmarksToOutline(root, c.pages)
}
