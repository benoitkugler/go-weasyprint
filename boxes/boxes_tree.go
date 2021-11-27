package boxes

import (
	"fmt"
	"log"

	"github.com/benoitkugler/go-weasyprint/images"
	"github.com/benoitkugler/go-weasyprint/layout/text"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/style/tree"
	"github.com/benoitkugler/go-weasyprint/utils"
	"golang.org/x/net/html"
)

type BlockLevelBox struct {
	Clearance pr.MaybeFloat
}

type BlockBox struct {
	BlockLevelBox
	BoxFields
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
	JustificationSpacing pr.Float
}

func TextBoxAnonymousFrom(parent Box, text string) *TextBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil, nil)
	out := NewTextBox(style, parent.Box().Element, parent.Box().PseudoType, text)
	return &out
}

type InlineBlockBox struct {
	BoxFields
}

type ReplacedBox struct {
	Replacement images.Image
	BoxFields
}

type BlockReplacedBox struct {
	BlockLevelBox
	ReplacedBox
}

type InlineReplacedBox struct {
	ReplacedBox
}

func InlineReplacedBoxAnonymousFrom(parent Box, replacement images.Image) *InlineReplacedBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil, nil)
	out := NewInlineReplacedBox(style, parent.Box().Element, parent.Box().PseudoType, replacement)
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
	BlockLevelBox
	BoxFields
}

type InlineFlexBox struct {
	InlineLevelBox
	BoxFields
}

type methodsBlockLevelBox interface {
	BlockLevel() *BlockLevelBox
}

func (b *BlockLevelBox) BlockLevel() *BlockLevelBox {
	return b
}

func NewBlockBox(style pr.ElementStyle, element *html.Node, pseudoType string, children []Box) *BlockBox {
	out := BlockBox{BoxFields: newBoxFields(style, element, pseudoType, children)}
	return &out
}

func LineBoxAnonymousFrom(parent Box, children []Box) Box {
	parentBox := parent.Box()
	style := tree.ComputedFromCascaded(nil, nil, parentBox.Style, nil, "", "", nil, nil)
	out := NewLineBox(style, parentBox.Element, parentBox.PseudoType, children)
	if parentBox.Style.GetOverflow() != "visible" {
		out.TextOverflow = string(parentBox.Style.GetTextOverflow())
	}
	return &out
}

func NewLineBox(style pr.ElementStyle, element *html.Node, pseudoType string, children []Box) LineBox {
	out := LineBox{BoxFields: newBoxFields(style, element, pseudoType, children)}
	out.TextOverflow = "clip"
	out.BlockEllipsis = pr.NamedString{Name: "none"}
	return out
}

func (*InlineLevelBox) RemoveDecoration(box *BoxFields, start, end bool) {
	if box.Style.GetBoxDecorationBreak() == "clone" {
		return
	}
	ltr := box.Style.GetDirection() == "ltr"
	if start {
		side := SRight
		if ltr {
			side = SLeft
		}
		box.ResetSpacing(side)
	}
	if end {
		side := SLeft
		if ltr {
			side = SRight
		}
		box.ResetSpacing(side)
	}
}

func NewInlineBox(style pr.ElementStyle, element *html.Node, pseudoType string, children []Box) *InlineBox {
	out := InlineBox{BoxFields: newBoxFields(style, element, pseudoType, children)}
	return &out
}

func NewTextBox(style pr.ElementStyle, element *html.Node, pseudoType string, text string) TextBox {
	if len(text) == 0 {
		panic("NewTextBox called with empty text")
	}
	box := newBoxFields(style, element, pseudoType, nil)
	out := TextBox{BoxFields: box, Text: text}
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
	u.InlineLevelBox.RemoveDecoration(b, start, end)
}

func NewInlineBlockBox(style pr.ElementStyle, element *html.Node, pseudoType string, children []Box) *InlineBlockBox {
	out := InlineBlockBox{BoxFields: newBoxFields(style, element, pseudoType, children)}
	return &out
}

func (u InlineBox) RemoveDecoration(b *BoxFields, start, end bool) {
	u.InlineLevelBox.RemoveDecoration(b, start, end)
}

func NewReplacedBox(style pr.ElementStyle, element *html.Node, pseudoType string, replacement images.Image) ReplacedBox {
	out := ReplacedBox{BoxFields: newBoxFields(style, element, pseudoType, nil)}
	out.Replacement = replacement
	return out
}

type methodsReplacedBox interface {
	Replaced() *ReplacedBox
}

func (b *ReplacedBox) Replaced() *ReplacedBox {
	return b
}

func NewBlockReplacedBox(style pr.ElementStyle, element *html.Node, pseudoType string, replacement images.Image) BlockReplacedBox {
	out := BlockReplacedBox{ReplacedBox: NewReplacedBox(style, element, pseudoType, replacement)}
	return out
}

func NewInlineReplacedBox(style pr.ElementStyle, element *html.Node, pseudoType string, replacement images.Image) InlineReplacedBox {
	out := InlineReplacedBox{ReplacedBox: NewReplacedBox(style, element, pseudoType, replacement)}
	return out
}

func (u InlineReplacedBox) RemoveDecoration(b *BoxFields, start, end bool) {
	u.ReplacedBox.RemoveDecoration(b, start, end)
}

type methodsTableBox interface {
	Table() *TableBox
}

func NewTableBox(style pr.ElementStyle, element *html.Node, pseudoType string, children []Box) *TableBox {
	out := TableBox{BoxFields: newBoxFields(style, element, pseudoType, children)}
	out.tabularContainer = true
	return &out
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

func NewInlineTableBox(style pr.ElementStyle, element *html.Node, pseudoType string, children []Box) *InlineTableBox {
	out := InlineTableBox{TableBox: *NewTableBox(style, element, pseudoType, children)}
	return &out
}

func NewTableRowGroupBox(style pr.ElementStyle, element *html.Node, pseudoType string, children []Box) *TableRowGroupBox {
	out := TableRowGroupBox{BoxFields: newBoxFields(style, element, pseudoType, children)}
	out.properTableChild = true
	out.internalTableOrCaption = true
	out.tabularContainer = true
	out.IsHeader = false
	out.IsFooter = false
	return &out
}

func NewTableRowBox(style pr.ElementStyle, element *html.Node, pseudoType string, children []Box) *TableRowBox {
	out := TableRowBox{BoxFields: newBoxFields(style, element, pseudoType, children)}
	out.properTableChild = true
	out.internalTableOrCaption = true
	out.tabularContainer = true
	return &out
}

func NewTableColumnGroupBox(style pr.ElementStyle, element *html.Node, pseudoType string, children []Box) *TableColumnGroupBox {
	out := TableColumnGroupBox{BoxFields: newBoxFields(style, element, pseudoType, children)}
	out.properTableChild = true
	out.internalTableOrCaption = true
	out.span = 1
	out.GetCells = out.defaultGetCells
	return &out
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

func NewTableColumnBox(style pr.ElementStyle, element *html.Node, pseudoType string, children []Box) *TableColumnBox {
	out := TableColumnBox{BoxFields: newBoxFields(style, element, pseudoType, children)}
	out.properTableChild = true
	out.internalTableOrCaption = true
	out.span = 1
	// GetCells is setup during table layout
	return &out
}

func NewTableCellBox(style pr.ElementStyle, element *html.Node, pseudoType string, children []Box) *TableCellBox {
	out := TableCellBox{BoxFields: newBoxFields(style, element, pseudoType, children)}
	out.internalTableOrCaption = true
	out.Colspan = 1
	out.Rowspan = 1
	return &out
}

func NewTableCaptionBox(style pr.ElementStyle, element *html.Node, pseudoType string, children []Box) *TableCaptionBox {
	out := TableCaptionBox{BlockBox: *NewBlockBox(style, element, pseudoType, children)}
	out.properTableChild = true
	out.internalTableOrCaption = true
	return &out
}

func NewPageBox(pageType utils.PageElement, style pr.ElementStyle) *PageBox {
	fields := newBoxFields(style, nil, "", nil)
	out := PageBox{BoxFields: fields, PageType: pageType}
	return &out
}

func (b *PageBox) String() string {
	return fmt.Sprintf("<PageBox %v>", b.PageType)
}

func NewMarginBox(atKeyword string, style pr.ElementStyle) *MarginBox {
	fields := newBoxFields(style, nil, "", nil)
	out := MarginBox{BoxFields: fields, AtKeyword: atKeyword}
	return &out
}

func (b *MarginBox) String() string {
	return fmt.Sprintf("<MarginBox %s>", b.AtKeyword)
}

func NewFlexBox(style pr.ElementStyle, element *html.Node, pseudoType string, children []Box) *FlexBox {
	out := FlexBox{BoxFields: newBoxFields(style, element, pseudoType, children)}
	return &out
}

func NewInlineFlexBox(style pr.ElementStyle, element *html.Node, pseudoType string, children []Box) *InlineFlexBox {
	out := InlineFlexBox{BoxFields: newBoxFields(style, element, pseudoType, children)}
	return &out
}

func (u InlineFlexBox) RemoveDecoration(b *BoxFields, start, end bool) {
	u.BoxFields.RemoveDecoration(b, start, end)
}
