package boxes

import (
	"fmt"
	"log"
	"strings"

	"github.com/benoitkugler/go-weasyprint/images"
	"github.com/benoitkugler/go-weasyprint/layout/text"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/style/tree"
	"github.com/benoitkugler/go-weasyprint/utils"
)

type BlockLevelBox struct {
	Clearance pr.MaybeFloat
}

type BlockBox struct {
	BoxFields
	BlockLevelBox
}

type LineBox struct {
	BoxFields
	TextIndent    pr.MaybeFloat
	TextOverflow  string
	BlockEllipsis pr.NamedString
}

type InlineLevelBox struct{}

type InlineBox struct {
	InlineLevelBox
	BoxFields
}

type TextBox struct {
	InlineLevelBox
	BoxFields

	PangoLayout          *text.TextLayout
	Text                 string
	OriginalText         string
	JustificationSpacing pr.Float
}

func TextBoxAnonymousFrom(parent Box, text string) *TextBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil, nil)
	out := NewTextBox(parent.Box().ElementTag, style, text)
	return &out
}

type InlineBlockBox struct {
	BoxFields
}

type ReplacedBox struct {
	BoxFields
	Replacement images.Image
}

type BlockReplacedBox struct {
	ReplacedBox
	BlockLevelBox
}

type InlineReplacedBox struct {
	ReplacedBox
}

func InlineReplacedBoxAnonymousFrom(parent Box, replacement images.Image) *InlineReplacedBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil, nil)
	out := NewInlineReplacedBox(parent.Box().ElementTag, style, replacement)
	return &out
}

type TableBox struct {
	BoxFields
	BlockLevelBox

	ColumnWidths, ColumnPositions []pr.Float
	ColumnGroups                  []Box
	CollapsedBorderGrid           BorderGrids
	SkippedRows                   int
}

type InlineTableBox struct {
	TableBox
}

type TableRowGroupBox struct {
	BoxFields
}

type TableRowBox struct {
	BoxFields
}

type TableColumnGroupBox struct {
	BoxFields
}

type TableColumnBox struct {
	BoxFields
}

type TableCellBox struct {
	BoxFields
}

type TableCaptionBox struct {
	BlockBox
}

type PageBox struct {
	BoxFields
	CanvasBackground *Background
	FixedBoxes       []Box
	PageType         utils.PageElement
}

type MarginBox struct {
	BoxFields
	AtKeyword   string
	IsGenerated bool
}

type FlexBox struct {
	BoxFields
	BlockLevelBox
}

type InlineFlexBox struct {
	InlineLevelBox
	BoxFields
}

type classicalBox interface {
	isBox()
}

// IsBox returns true for all standard boxes defined in this package, but false
// for the special ones, defined in other packages, like AbsolutePlaceholder or StackingContext.
func IsBox(b Box) bool {
	_, is := b.(classicalBox)
	return is
}

type methodsBlockLevelBox interface {
	BlockLevel() *BlockLevelBox
}

func (b *BlockLevelBox) BlockLevel() *BlockLevelBox {
	return b
}

func NewBlockBox(elementTag string, style pr.ElementStyle, children []Box) BlockBox {
	out := BlockBox{BoxFields: newBoxFields(elementTag, style, children)}
	return out
}

func LineBoxAnonymousFrom(parent Box, children []Box) Box {
	parentBox := parent.Box()
	style := tree.ComputedFromCascaded(nil, nil, parentBox.Style, nil, "", "", nil, nil)
	out := NewLineBox(parentBox.ElementTag, style, children)
	if parentBox.Style.GetOverflow() != "visible" {
		out.TextOverflow = string(parentBox.Style.GetTextOverflow())
	}
	return &out
}

func NewLineBox(elementTag string, style pr.ElementStyle, children []Box) LineBox {
	out := LineBox{BoxFields: newBoxFields(elementTag, style, children)}
	out.TextOverflow = "clip"
	return out
}

func (InlineLevelBox) removeDecoration(box *BoxFields, start, end bool) {
	ltr := box.Style.GetDirection() == "ltr"
	if start {
		side := "right"
		if ltr {
			side = "left"
		}
		box.ResetSpacing(side)
	}
	if end {
		side := "left"
		if ltr {
			side = "right"
		}
		box.ResetSpacing(side)
	}
}

func NewInlineBox(elementTag string, style pr.ElementStyle, children []Box) InlineBox {
	out := InlineBox{BoxFields: newBoxFields(elementTag, style, children)}
	return out
}

// Return the (x, y, w, h) rectangle where the box is clickable.
func (b *InlineBox) hitArea() (x, y, w, h pr.Float) {
	return b.Box().BorderBoxX(), b.Box().PositionY, b.Box().BorderWidth(), b.Box().MarginHeight()
}

func NewTextBox(elementTag string, style pr.ElementStyle, text string) TextBox {
	if len(text) == 0 {
		log.Fatalf("empty text")
	}
	originalText := text
	textTransform := style.GetTextTransform()
	if textTransform != "none" {
		switch textTransform {
		case "uppercase":
			text = strings.ToUpper(text)
		case "lowercase":
			text = strings.ToLower(text)
		// Python’s unicode.captitalize is not the same.
		case "capitalize":
			text = strings.Title(strings.ToLower(text))
		case "full-width":
			text = strings.Map(func(u rune) rune {
				rep, in := asciiToWide[u]
				if !in {
					return u
				}
				return rep
			}, text)
		}
	}
	if style.GetHyphens() == "none" {
		text = strings.ReplaceAll(text, "\u00AD", "") //  U+00AD SOFT HYPHEN (SHY)
	}
	box := newBoxFields(elementTag, style, nil)
	out := TextBox{BoxFields: box, Text: text, OriginalText: originalText}
	return out
}

// Return a new TextBox identical to this one except for the text.
func (b TextBox) CopyWithText(text string) *TextBox {
	if len(text) == 0 {
		log.Fatal("empty text")
	}
	newBox := b
	newBox.Text = text
	return &newBox
}

func (u TextBox) RemoveDecoration(b *BoxFields, start, end bool) {
	u.InlineLevelBox.removeDecoration(b, start, end)
}

func NewInlineBlockBox(elementTag string, style pr.ElementStyle, children []Box) InlineBlockBox {
	out := InlineBlockBox{BoxFields: newBoxFields(elementTag, style, children)}
	return out
}

func (u InlineBox) RemoveDecoration(b *BoxFields, start, end bool) {
	u.InlineLevelBox.removeDecoration(b, start, end)
}

func NewReplacedBox(elementTag string, style pr.ElementStyle, replacement images.Image) ReplacedBox {
	out := ReplacedBox{BoxFields: newBoxFields(elementTag, style, nil)}
	out.Replacement = replacement
	return out
}

type methodsReplacedBox interface {
	Replaced() *ReplacedBox
}

func (b *ReplacedBox) Replaced() *ReplacedBox {
	return b
}

func NewBlockReplacedBox(elementTag string, style pr.ElementStyle, replacement images.Image) BlockReplacedBox {
	out := BlockReplacedBox{ReplacedBox: NewReplacedBox(elementTag, style, replacement)}
	return out
}

func NewInlineReplacedBox(elementTag string, style pr.ElementStyle, replacement images.Image) InlineReplacedBox {
	out := InlineReplacedBox{ReplacedBox: NewReplacedBox(elementTag, style, replacement)}
	return out
}

func (u InlineReplacedBox) RemoveDecoration(b *BoxFields, start, end bool) {
	u.ReplacedBox.RemoveDecoration(b, start, end)
}

type methodsTableBox interface {
	Table() *TableBox
}

func NewTableBox(elementTag string, style pr.ElementStyle, children []Box) TableBox {
	out := TableBox{BoxFields: newBoxFields(elementTag, style, children)}
	out.tabularContainer = true
	return out
}

// Table implements InstanceTableBox
func (b *TableBox) Table() *TableBox {
	return b
}

func (b *TableBox) AllChildren() []Box {
	return append(b.Box().Children, b.ColumnGroups...)
}

func (b *TableBox) Translate(box Box, dx, dy pr.Float, ignoreFloats bool) {
	if dx == 0 && dy == 0 {
		return
	}
	for index, position := range b.ColumnPositions {
		b.ColumnPositions[index] = position + dx
	}
	b.BoxFields.Translate(box, dx, dy, ignoreFloats)
}

func (b *TableBox) PageValues() (pr.Page, pr.Page) {
	s := b.Box().Style
	return s.GetPage(), s.GetPage()
}

func NewInlineTableBox(elementTag string, style pr.ElementStyle, children []Box) InlineTableBox {
	out := InlineTableBox{TableBox: NewTableBox(elementTag, style, children)}
	return out
}

func NewTableRowGroupBox(elementTag string, style pr.ElementStyle, children []Box) TableRowGroupBox {
	out := TableRowGroupBox{BoxFields: newBoxFields(elementTag, style, children)}
	out.properTableChild = true
	out.internalTableOrCaption = true
	out.tabularContainer = true
	out.IsHeader = false
	out.IsFooter = false
	return out
}

func NewTableRowBox(elementTag string, style pr.ElementStyle, children []Box) TableRowBox {
	out := TableRowBox{BoxFields: newBoxFields(elementTag, style, children)}
	out.properTableChild = true
	out.internalTableOrCaption = true
	out.tabularContainer = true
	return out
}

func NewTableColumnGroupBox(elementTag string, style pr.ElementStyle, children []Box) TableColumnGroupBox {
	out := TableColumnGroupBox{BoxFields: newBoxFields(elementTag, style, children)}
	out.properTableChild = true
	out.internalTableOrCaption = true
	out.span = 1
	out.GetCells = out.defaultGetCells
	return out
}

// Return cells that originate in the group's columns.
func (b *TableColumnGroupBox) defaultGetCells() []Box {
	var out []Box
	for _, column := range b.Box().Children {
		for _, cell := range column.Box().GetCells() {
			out = append(out, cell)
		}
	}
	return out
}

func NewTableColumnBox(elementTag string, style pr.ElementStyle, children []Box) TableColumnBox {
	out := TableColumnBox{BoxFields: newBoxFields(elementTag, style, children)}
	out.properTableChild = true
	out.internalTableOrCaption = true
	out.span = 1
	out.GetCells = out.defaultGetCells
	return out
}

// Return cells that originate in the column.
// May be overriden on instances.
func (b *TableColumnBox) defaultGetCells() []Box {
	return []Box{}
}

func NewTableCellBox(elementTag string, style pr.ElementStyle, children []Box) TableCellBox {
	out := TableCellBox{BoxFields: newBoxFields(elementTag, style, children)}
	out.internalTableOrCaption = true
	out.Colspan = 1
	out.Rowspan = 1
	return out
}

func NewTableCaptionBox(elementTag string, style pr.ElementStyle, children []Box) TableCaptionBox {
	out := TableCaptionBox{BlockBox: NewBlockBox(elementTag, style, children)}
	out.properTableChild = true
	out.internalTableOrCaption = true
	return out
}

func NewPageBox(pageType utils.PageElement, style pr.ElementStyle) *PageBox {
	fields := newBoxFields("", style, nil)
	out := PageBox{BoxFields: fields, PageType: pageType}
	return &out
}

func (b *PageBox) String() string {
	return fmt.Sprintf("<PageBox %v>", b.PageType)
}

func NewMarginBox(atKeyword string, style pr.ElementStyle) *MarginBox {
	fields := newBoxFields("", style, nil)
	out := MarginBox{BoxFields: fields, AtKeyword: atKeyword}
	return &out
}

func (b *MarginBox) String() string {
	return fmt.Sprintf("<MarginBox %s>", b.AtKeyword)
}

func NewFlexBox(elementTag string, style pr.ElementStyle, children []Box) FlexBox {
	out := FlexBox{BoxFields: newBoxFields(elementTag, style, children)}
	return out
}

func NewInlineFlexBox(elementTag string, style pr.ElementStyle, children []Box) InlineFlexBox {
	out := InlineFlexBox{BoxFields: newBoxFields(elementTag, style, children)}
	return out
}

func (u InlineFlexBox) RemoveDecoration(b *BoxFields, start, end bool) {
	u.BoxFields.RemoveDecoration(b, start, end)
}
