package structure

import (
	"errors"
	"fmt"
	"log"
	"strings"

	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/utils"
)






// Create a new equivalent box with given ``newChildren``.
// isStart=true, isEnd=true
func copyWithChildren(box Box, newChildren []Box, isStart, isEnd bool) Box {
	newBox := box.Copy()
	newBox.Box().children = newChildren
	if box.Box().style.GetBoxDecorationBreak() == "slice" {
		newBox.removeDecoration(!isStart, !isEnd)
	}
	return newBox
}

func deepCopy(box Box) Box {
	result := box.Copy()
	l := result.Box().children
	for i, v := range l {
		l[i] = deepCopy(v)
	}
	return result
}

// A flat generator for a box, its children and descendants."""
func descendants(box Box) []Box {
	out := []Box{box}
	for _, child := range box.Box().children {
		out = append(out, descendants(child)...)
	}
	return out
}

// Get the table wrapped by the box.
func (self ParentBox) getWrappedTable() (TableBoxInstance, error) {
	if self.isTableWrapper {
		for _, child := range self.children {
			if asTable, ok := child.(TableBoxInstance); ok {
				return asTable, nil
			}
		}
		return nil, errors.New("Table wrapper without a table")
	}
	return nil, nil
}

func (self ParentBox) pageValues() (pr.Page, pr.Page) {
	start, end := self.Box().pageValues()
	if len(self.children) > 0 {
		startBox, endBox := self.children[0], self.children[len(self.children)-1]
		childStart, _ := startBox.Box().pageValues()
		_, childEnd := endBox.Box().pageValues()
		if !childStart.IsNone() {
			start = childStart
		}
		if !childEnd.IsNone() {
			end = childEnd
		}
	}
	return start, end
}

func LineBoxAnonymousFrom(parent Box, children []Box) *LineBox {
	parentBox := ParentBoxAnonymousFrom(parent, children)
	out := LineBox{ParentBox: *parentBox, textOverflow: "clip"}
	if parentBox.style.GetOverflow() != "visible" {
		out.textOverflow = parentBox.style.GetTextOverflow()
	}
	return &out
}

// func (self BlockBox) pageValues() (int, int) {
// 	return self.BlockContainerBox.pageValues()
// }

func (self InlineLevelBox) removeDecoration(start, end bool) {
	ltr := self.style.GetDirection() == "ltr"
	if start {
		side := "right"
		if ltr {
			side = "left"
		}
		self.resetSpacing(side)
	}
	if end {
		side := "left"
		if ltr {
			side = "right"
		}
		self.resetSpacing(side)
	}
}

func NewInlineBox(elementTag string, style pr.Properties, children []Box) *InlineBox {
	out := InlineBox{}
	parent := NewInlineLevelBox(elementTag, style, children)
	out.InlineLevelBox = *parent
	out.ParentBox = ParentBox{BoxFields: parent.BoxFields}
	return &out
}

// Return the (x, y, w, h) rectangle where the box is clickable.
func (self InlineBox) hitArea() (x float32, y float32, w float32, h float32) {
	return self.InlineLevelBox.borderBoxX(), self.InlineLevelBox.positionY, self.InlineLevelBox.borderWidth(), self.InlineLevelBox.marginHeight()
}

func NewTextBox(elementTag string, style pr.Properties, text string) *TextBox {
	if len(text) == 0 {
		log.Fatalf("empty text")
	}
	box := NewInlineLevelBox(elementTag, style, nil)
	textTransform := style.GetTextTransform()
	if textTransform != "none" {
		switch textTransform {
		case "uppercase":
			text = strings.ToUpper(text)
		case "lowercase":
			text = strings.ToLower(text)
		// Python’s unicode.captitalize is not the same.
		case "capitalize":
			text = strings.ToTitle(text)
		case "full-width":
			text = strings.Map(func(u rune) rune {
				rep, in := asciiToWide[u]
				if !in {
					return -1
				}
				return rep
			}, text)
		}
	}
	if style.GetHyphens() == "none" {
		text = strings.ReplaceAll(text, "\u00AD", "") //  U+00AD SOFT HYPHEN (SHY)
	}
	out := TextBox{InlineLevelBox: *box, text: text}
	return &out
}

// Return a new TextBox identical to this one except for the text.
func (self TextBox) copyWithText(text string) TextBox {
	if len(text) == 0 {
		log.Fatal("empty text")
	}
	newBox := self
	newBox.text = text
	return newBox
}

func (self *InlineBlockBox) Translate(dx, dy float32, ignoreFloats bool) {
	self.AtomicInlineLevelBox.Translate(dx, dy, ignoreFloats)
}

// Definitions for the rules generating anonymous table boxes
// http://www.w3.org/TR/CSS21/tables.html#anonymous-boxes

func (self *TableBox) Translate(dx, dy float32, ignoreFloats bool) {
	if dx == 0 && dy == 0 {
		return
	}
	for index, position := range self.columnPositions {
		self.columnPositions[index] = position + dx
	}
	translate(append(self.children, self.columnGroups...), self.Box(), dx, dy, ignoreFloats)
}

func (self TableBox) pageValues() (pr.Page, pr.Page) {
	return self.ParentBox.pageValues()
}

func (self TableRowGroupBox) IsProperChild(parent Box) bool {
	switch parent.(type) {
	case *TableBox, *InlineTableBox:
		return true
	default:
		return false
	}
}

func (self TableRowBox) IsProperChild(parent Box) bool {
	switch parent.(type) {
	case *TableBox, *InlineTableBox, *TableRowGroupBox:
		return true
	default:
		return false
	}
}

func (self TableColumnGroupBox) IsProperChild(parent Box) bool {
	switch parent.(type) {
	case *TableBox, *InlineTableBox:
		return true
	default:
		return false
	}
}

type withCells interface {
	getCells() []Box
}

// Return cells that originate in the group's columns.
func (self *TableColumnGroupBox) getCells() []Box {
	var out []Box
	for _, column := range self.children {
		for _, cell := range column.(withCells).getCells() {
			out = append(out, cell)
		}
	}
	return out
}

// Return cells that originate in the column.
// May be overriden on instances.
func (self *TableColumnBox) getCells() []Box {
	return []Box{}
}

func (self TableColumnBox) IsProperChild(parent Box) bool {
	switch parent.(type) {
	case *TableBox, *InlineTableBox, *TableColumnGroupBox:
		return true
	default:
		return false
	}
}

func (self TableCaptionBox) IsProperChild(parent Box) bool {
	switch parent.(type) {
	case *TableBox, *InlineTableBox:
		return true
	default:
		return false
	}
}

func TableCaptionBoxIsInstance(box Box) bool {
	_, is := box.(*TableCaptionBox)
	return is
}

func NewPageBox(pageType utils.PageElement, style pr.Properties) *PageBox {
	parent := NewParentBox("", style, nil)
	return &PageBox{ParentBox: *parent, pageType: pageType}
}

func (self PageBox) String() string {
	return fmt.Sprintf("<PageBox %s>", self.pageType)
}

func NewMarginBox(atKeyword string, style pr.Properties) *MarginBox {
	b := NewBlockContainerBox("", style, nil)
	return &MarginBox{BlockContainerBox: *b, atKeyword: atKeyword}
}

func (self MarginBox) String() string {
	return fmt.Sprintf("<MarginBox %s>", self.atKeyword)
}
