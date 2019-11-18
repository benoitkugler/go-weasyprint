package pdf

import (
	"strings"
	"time"

	"github.com/benoitkugler/go-weasyprint/backend"
	"github.com/benoitkugler/gofpdf"
)

func (s Context) SetTitle(title string) {
	s.f.SetTitle(title, true)
}
func (s Context) SetDescription(description string) {
	s.f.SetSubject(description, true)
}
func (s Context) SetCreator(creator string) {
	s.f.SetCreator(creator, true)
}
func (s Context) SetAuthors(authors []string) {
	s.f.SetKeywords(strings.Join(authors, ", "), true)
}
func (s Context) SetKeywords(keywords []string) {
	s.f.SetKeywords(strings.Join(keywords, ", "), true)
}
func (s Context) SetProducer(producer string) {
	s.f.SetProducer(producer, true)
}
func (s Context) SetDateCreation(d time.Time) {
	s.f.SetCreationDate(d)
}
func (s Context) SetDateModification(d time.Time) {
	s.f.SetModificationDate(d)
}

// links

func (c Context) CreateAnchors(anchors [][]backend.Anchor) map[string]int {
	out := map[string]int{}
	for i, l := range anchors {
		page := i + 1 // anchors is 0-based
		for _, anchor := range l {
			y := anchor.Pos[1]
			id := c.f.AddLink()      // create new link
			c.f.SetLink(id, y, page) // place target
			out[anchor.Name] = id    // register association
		}
	}
	return out
}

func (c Context) AddInternalLink(x, y, w, h float64, linkId int) {
	_, pageHeight := c.f.GetPageSize()
	c.f.Link(x, pageHeight-y, w, h, linkId)
}

func (c Context) AddExternalLink(x, y, w, h float64, url string) {
	_, pageHeight := c.f.GetPageSize()
	c.f.LinkString(x, pageHeight-y, w, h, url)
}

// embedded files

// Add global attachments to the file
func (c Context) SetAttachments(as []backend.Attachment) {
	cv := make([]gofpdf.Attachment, len(as))
	for i, a := range as {
		cv[i] = gofpdf.Attachment{Content: a.Content, Filename: a.Title, Description: a.Description}
	}
	c.f.SetAttachments(cv)
}

// Embed a file. Calling this method twice with the same id
// won't embed the content twice.
func (c Context) EmbedFile(id string, a backend.Attachment) {
	ptr := c.fileAnnotationsMap[id] // cache the attachment by id
	if ptr == nil {
		ptr = &gofpdf.Attachment{Content: a.Content, Filename: a.Title, Description: a.Description}
		c.fileAnnotationsMap[id] = ptr
	}
}

// Add file annotation on the current page
func (c Context) AddFileAnnotation(x, y, w, h float64, id string) {
	a := c.fileAnnotationsMap[id]
	_, pageHeight := c.f.GetPageSize()
	c.f.AddAttachmentAnnotation(a, x, pageHeight-y, w, h)
}
