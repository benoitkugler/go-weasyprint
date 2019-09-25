package structure2

import "github.com/benoitkugler/go-weasyprint/css"

// autogenerated from boxes.go

func NewBoxFields(elementTag string, style css.StyleFor, children []Box) *BoxFields {
	out := BoxFields{
		elementTag: elementTag,
		style:      style,
		children:   children,
	}

	return &out
}
func NewParentBox(elementTag string, style css.StyleFor, children []Box) *ParentBox {
	out := ParentBox{}
	parent := NewBoxFields(elementTag, style, children)
	out.BoxFields = *parent
	return &out
}
func ParentBoxAnonymousFrom(parent Box, children []Box) *ParentBox {
	return NewParentBox(parent.Box().elementTag, parent.Box().style.InheritFrom(), children)
}
func (b ParentBox) Copy() Box { return &b }

func NewBlockLevelBox(elementTag string, style css.StyleFor, children []Box) *BlockLevelBox {
	out := BlockLevelBox{}

	return &out
}
func BlockLevelBoxAnonymousFrom(parent Box, children []Box) *BlockLevelBox {
	return NewBlockLevelBox(parent.Box().elementTag, parent.Box().style.InheritFrom(), children)
}
func (b BlockLevelBox) Copy() Box { return &b }

func NewBlockContainerBox(elementTag string, style css.StyleFor, children []Box) *BlockContainerBox {
	out := BlockContainerBox{}
	parent := NewParentBox(elementTag, style, children)
	out.ParentBox = *parent
	return &out
}
func BlockContainerBoxAnonymousFrom(parent Box, children []Box) *BlockContainerBox {
	return NewBlockContainerBox(parent.Box().elementTag, parent.Box().style.InheritFrom(), children)
}
func (b BlockContainerBox) Copy() Box { return &b }

func NewBlockBox(elementTag string, style css.StyleFor, children []Box) *BlockBox {
	out := BlockBox{}
	parent := NewBlockContainerBox(elementTag, style, children)
	out.BlockContainerBox = *parent
	return &out
}
func BlockBoxAnonymousFrom(parent Box, children []Box) *BlockBox {
	return NewBlockBox(parent.Box().elementTag, parent.Box().style.InheritFrom(), children)
}
func (b BlockBox) Copy() Box { return &b }

func LineBoxAnonymousFrom(parent Box, children []Box) *LineBox {
	return NewLineBox(parent.Box().elementTag, parent.Box().style.InheritFrom(), children)
}
func (b LineBox) Copy() Box { return &b }

func NewInlineLevelBox(elementTag string, style css.StyleFor, children []Box) *InlineLevelBox {
	out := InlineLevelBox{}

	return &out
}
func InlineLevelBoxAnonymousFrom(parent Box, children []Box) *InlineLevelBox {
	return NewInlineLevelBox(parent.Box().elementTag, parent.Box().style.InheritFrom(), children)
}
func (b InlineLevelBox) Copy() Box { return &b }

func NewInlineBox(elementTag string, style css.StyleFor, children []Box) *InlineBox {
	out := InlineBox{}
	parent := NewInlineLevelBox(elementTag, style, children)
	out.InlineLevelBox = *parent
	return &out
}
func InlineBoxAnonymousFrom(parent Box, children []Box) *InlineBox {
	return NewInlineBox(parent.Box().elementTag, parent.Box().style.InheritFrom(), children)
}
func (b InlineBox) Copy() Box { return &b }

func TextBoxAnonymousFrom(parent Box, text string) *TextBox {
	return NewTextBox(parent.Box().elementTag, parent.Box().style.InheritFrom(), text)
}
func (b TextBox) Copy() Box { return &b }

func NewAtomicInlineLevelBox(elementTag string, style css.StyleFor, children []Box) *AtomicInlineLevelBox {
	out := AtomicInlineLevelBox{}
	parent := NewInlineLevelBox(elementTag, style, children)
	out.InlineLevelBox = *parent
	return &out
}
func AtomicInlineLevelBoxAnonymousFrom(parent Box, children []Box) *AtomicInlineLevelBox {
	return NewAtomicInlineLevelBox(parent.Box().elementTag, parent.Box().style.InheritFrom(), children)
}
func (b AtomicInlineLevelBox) Copy() Box { return &b }

func NewInlineBlockBox(elementTag string, style css.StyleFor, children []Box) *InlineBlockBox {
	out := InlineBlockBox{}
	parent := NewAtomicInlineLevelBox(elementTag, style, children)
	out.AtomicInlineLevelBox = *parent
	return &out
}
func InlineBlockBoxAnonymousFrom(parent Box, children []Box) *InlineBlockBox {
	return NewInlineBlockBox(parent.Box().elementTag, parent.Box().style.InheritFrom(), children)
}
func (b InlineBlockBox) Copy() Box { return &b }

func NewReplacedBox(elementTag string, style css.StyleFor, replacement css.ImageType) *ReplacedBox {
	out := ReplacedBox{
		replacement: replacement,
	}
	parent := NewBoxFields(elementTag, style, nil)
	out.BoxFields = *parent
	return &out
}
func ReplacedBoxAnonymousFrom(parent Box, nil css.ImageType) *ReplacedBox {
	return NewReplacedBox(parent.Box().elementTag, parent.Box().style.InheritFrom(), nil)
}
func (b ReplacedBox) Copy() Box { return &b }

func NewBlockReplacedBox(elementTag string, style css.StyleFor, replacement css.ImageType) *BlockReplacedBox {
	out := BlockReplacedBox{}
	parent := NewReplacedBox(elementTag, style, replacement)
	out.ReplacedBox = *parent
	return &out
}
func BlockReplacedBoxAnonymousFrom(parent Box, replacement css.ImageType) *BlockReplacedBox {
	return NewBlockReplacedBox(parent.Box().elementTag, parent.Box().style.InheritFrom(), replacement)
}
func (b BlockReplacedBox) Copy() Box { return &b }

func NewInlineReplacedBox(elementTag string, style css.StyleFor, replacement css.ImageType) *InlineReplacedBox {
	out := InlineReplacedBox{}
	parent := NewReplacedBox(elementTag, style, replacement)
	out.ReplacedBox = *parent
	return &out
}
func InlineReplacedBoxAnonymousFrom(parent Box, replacement css.ImageType) *InlineReplacedBox {
	return NewInlineReplacedBox(parent.Box().elementTag, parent.Box().style.InheritFrom(), replacement)
}
func (b InlineReplacedBox) Copy() Box { return &b }

func NewTableBox(elementTag string, style css.StyleFor, children []Box) *TableBox {
	out := TableBox{
		tabularContainer: true,
	}
	parent := NewParentBox(elementTag, style, children)
	out.ParentBox = *parent
	return &out
}
func TableBoxAnonymousFrom(parent Box, children []Box) *TableBox {
	return NewTableBox(parent.Box().elementTag, parent.Box().style.InheritFrom(), children)
}
func (b TableBox) Copy() Box { return &b }

func NewInlineTableBox(elementTag string, style css.StyleFor, children []Box) *InlineTableBox {
	out := InlineTableBox{}
	parent := NewTableBox(elementTag, style, children)
	out.TableBox = *parent
	return &out
}
func InlineTableBoxAnonymousFrom(parent Box, children []Box) *InlineTableBox {
	return NewInlineTableBox(parent.Box().elementTag, parent.Box().style.InheritFrom(), children)
}
func (b InlineTableBox) Copy() Box { return &b }

func NewTableRowGroupBox(elementTag string, style css.StyleFor, children []Box) *TableRowGroupBox {
	out := TableRowGroupBox{
		properTableChild:       true,
		internalTableOrCaption: true,
		tabularContainer:       true,
	}
	parent := NewParentBox(elementTag, style, children)
	out.ParentBox = *parent
	return &out
}
func TableRowGroupBoxAnonymousFrom(parent Box, children []Box) *TableRowGroupBox {
	return NewTableRowGroupBox(parent.Box().elementTag, parent.Box().style.InheritFrom(), children)
}
func (b TableRowGroupBox) Copy() Box { return &b }

func NewTableRowBox(elementTag string, style css.StyleFor, children []Box) *TableRowBox {
	out := TableRowBox{
		properTableChild:       true,
		internalTableOrCaption: true,
		tabularContainer:       true,
	}
	parent := NewParentBox(elementTag, style, children)
	out.ParentBox = *parent
	return &out
}
func TableRowBoxAnonymousFrom(parent Box, children []Box) *TableRowBox {
	return NewTableRowBox(parent.Box().elementTag, parent.Box().style.InheritFrom(), children)
}
func (b TableRowBox) Copy() Box { return &b }

func NewTableColumnGroupBox(elementTag string, style css.StyleFor, children []Box) *TableColumnGroupBox {
	out := TableColumnGroupBox{
		properTableChild:       true,
		internalTableOrCaption: true,
		span:                   1,
	}
	parent := NewParentBox(elementTag, style, children)
	out.ParentBox = *parent
	return &out
}
func TableColumnGroupBoxAnonymousFrom(parent Box, children []Box) *TableColumnGroupBox {
	return NewTableColumnGroupBox(parent.Box().elementTag, parent.Box().style.InheritFrom(), children)
}
func (b TableColumnGroupBox) Copy() Box { return &b }

func NewTableColumnBox(elementTag string, style css.StyleFor, children []Box) *TableColumnBox {
	out := TableColumnBox{
		properTableChild:       true,
		internalTableOrCaption: true,
		span:                   1,
	}
	parent := NewParentBox(elementTag, style, children)
	out.ParentBox = *parent
	return &out
}
func TableColumnBoxAnonymousFrom(parent Box, children []Box) *TableColumnBox {
	return NewTableColumnBox(parent.Box().elementTag, parent.Box().style.InheritFrom(), children)
}
func (b TableColumnBox) Copy() Box { return &b }

func NewTableCellBox(elementTag string, style css.StyleFor, children []Box) *TableCellBox {
	out := TableCellBox{
		internalTableOrCaption: true,
		colspan:                1,
		rowspan:                1,
	}
	parent := NewBlockContainerBox(elementTag, style, children)
	out.BlockContainerBox = *parent
	return &out
}
func TableCellBoxAnonymousFrom(parent Box, children []Box) *TableCellBox {
	return NewTableCellBox(parent.Box().elementTag, parent.Box().style.InheritFrom(), children)
}
func (b TableCellBox) Copy() Box { return &b }

func NewTableCaptionBox(elementTag string, style css.StyleFor, children []Box) *TableCaptionBox {
	out := TableCaptionBox{
		properTableChild:       true,
		internalTableOrCaption: true,
	}
	parent := NewBlockBox(elementTag, style, children)
	out.BlockBox = *parent
	return &out
}
func TableCaptionBoxAnonymousFrom(parent Box, children []Box) *TableCaptionBox {
	return NewTableCaptionBox(parent.Box().elementTag, parent.Box().style.InheritFrom(), children)
}
func (b TableCaptionBox) Copy() Box { return &b }

func (b PageBox) Copy() Box { return &b }

func (b MarginBox) Copy() Box { return &b }

var (
	TypeTextBox             BoxType = typeTextBox{}
	TypeTableBox            BoxType = typeTableBox{}
	TypeInlineTableBox      BoxType = typeInlineTableBox{}
	TypeTableRowGroupBox    BoxType = typeTableRowGroupBox{}
	TypeTableRowBox         BoxType = typeTableRowBox{}
	TypeTableColumnGroupBox BoxType = typeTableColumnGroupBox{}
	TypeTableColumnBox      BoxType = typeTableColumnBox{}
	TypeTableCellBox        BoxType = typeTableCellBox{}
)

func (t typeTextBox) IsInstance(box Box) bool {
	_, is := box.(*TextBox)
	return is
}

type typeTextBox struct{}

func (t typeTableBox) AnonymousFrom(parent Box, children []Box) Box {
	return TableBoxAnonymousFrom(parent, children)
}

type typeTableBox struct{}

func (t typeInlineTableBox) AnonymousFrom(parent Box, children []Box) Box {
	return InlineTableBoxAnonymousFrom(parent, children)
}

func (t typeInlineTableBox) IsInstance(box Box) bool {
	_, is := box.(*InlineTableBox)
	return is
}

type typeInlineTableBox struct{}

func (t typeTableRowGroupBox) AnonymousFrom(parent Box, children []Box) Box {
	return TableRowGroupBoxAnonymousFrom(parent, children)
}

func (t typeTableRowGroupBox) IsInstance(box Box) bool {
	_, is := box.(*TableRowGroupBox)
	return is
}

type typeTableRowGroupBox struct{}

func (t typeTableRowBox) AnonymousFrom(parent Box, children []Box) Box {
	return TableRowBoxAnonymousFrom(parent, children)
}

func (t typeTableRowBox) IsInstance(box Box) bool {
	_, is := box.(*TableRowBox)
	return is
}

type typeTableRowBox struct{}

func (t typeTableColumnGroupBox) AnonymousFrom(parent Box, children []Box) Box {
	return TableColumnGroupBoxAnonymousFrom(parent, children)
}

func (t typeTableColumnGroupBox) IsInstance(box Box) bool {
	_, is := box.(*TableColumnGroupBox)
	return is
}

type typeTableColumnGroupBox struct{}

func (t typeTableColumnBox) AnonymousFrom(parent Box, children []Box) Box {
	return TableColumnBoxAnonymousFrom(parent, children)
}

func (t typeTableColumnBox) IsInstance(box Box) bool {
	_, is := box.(*TableColumnBox)
	return is
}

type typeTableColumnBox struct{}

func (t typeTableCellBox) AnonymousFrom(parent Box, children []Box) Box {
	return TableCellBoxAnonymousFrom(parent, children)
}

func (t typeTableCellBox) IsInstance(box Box) bool {
	_, is := box.(*TableCellBox)
	return is
}

type typeTableCellBox struct{}
