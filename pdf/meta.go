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

// links

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
		var lastChild *model.OutlineItem
		for i, child := range node.Children {
			childItem := nodeToItem(child, out)
			if i == 0 {
				out.First = childItem
			}
			if lastChild != nil {
				lastChild.Next = childItem
			}

			lastChild = childItem
		}
		return out
	}
	var (
		outline   model.Outline
		lastChild *model.OutlineItem
	)
	for i, node := range root {
		item := nodeToItem(node, &outline)
		if i == 0 {
			outline.First = item
		}
		if lastChild != nil {
			lastChild.Next = item
		}

		lastChild = item
	}

	return &outline
}

func (c *Context) SetBookmarks(root []backend.BookmarkNode) {
	c.document.Catalog.Outlines = bookmarksToOutline(root, c.pages)
}
