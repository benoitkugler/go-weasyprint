package structure2

import "github.com/benoitkugler/go-weasyprint/css"

func NewBoxFields(elementTag string, style css.StyleDict, children []Box) *BoxFields {
	out := BoxFields{
		elementTag: elementTag,
		style:      style,
		children:   children,
	}

	return &out
}
func NewParentBox(elementTag string, style css.StyleDict, children []Box) *ParentBox {
	out := ParentBox{}
	parent := NewBoxFields(elementTag, style, children)
	out.BoxFields = *parent
	return &out
}
func NewBlockLevelBox(elementTag string, style css.StyleDict, children []Box) *BlockLevelBox {
	out := BlockLevelBox{}

	return &out
}
func NewBlockContainerBox(elementTag string, style css.StyleDict, children []Box) *BlockContainerBox {
	out := BlockContainerBox{}
	parent := NewParentBox(elementTag, style, children)
	out.ParentBox = *parent
	return &out
}
func NewBlockBox(elementTag string, style css.StyleDict, children []Box) *BlockBox {
	out := BlockBox{}
	parent := NewBlockContainerBox(elementTag, style, children)
	out.BlockContainerBox = *parent
	return &out
}
func NewLineBox(elementTag string, style css.StyleDict, children []Box) *LineBox {
	out := LineBox{}
	parent := NewParentBox(elementTag, style, children)
	out.ParentBox = *parent
	return &out
}
func NewInlineLevelBox(elementTag string, style css.StyleDict, children []Box) *InlineLevelBox {
	out := InlineLevelBox{}

	return &out
}
func NewInlineBox(elementTag string, style css.StyleDict, children []Box) *InlineBox {
	out := InlineBox{}
	parent := NewInlineLevelBox(elementTag, style, children)
	out.InlineLevelBox = *parent
	return &out
}
func NewTextBox(elementTag string, style css.StyleDict, children []Box) *TextBox {
	out := TextBox{}
	parent := NewBoxFields(elementTag, style, children)
	out.BoxFields = *parent
	return &out
}
func NewAtomicInlineLevelBox(elementTag string, style css.StyleDict, children []Box) *AtomicInlineLevelBox {
	out := AtomicInlineLevelBox{}
	parent := NewInlineLevelBox(elementTag, style, children)
	out.InlineLevelBox = *parent
	return &out
}
func NewInlineBlockBox(elementTag string, style css.StyleDict, children []Box) *InlineBlockBox {
	out := InlineBlockBox{}
	parent := NewAtomicInlineLevelBox(elementTag, style, children)
	out.AtomicInlineLevelBox = *parent
	return &out
}
func NewReplacedBox(elementTag string, style css.StyleDict, replacement css.ImageType) *ReplacedBox {
	out := ReplacedBox{
		replacement: replacement,
	}
	parent := NewBoxFields(elementTag, style, nil)
	out.BoxFields = *parent
	return &out
}
func NewBlockReplacedBox(elementTag string, style css.StyleDict, replacement css.ImageType) *BlockReplacedBox {
	out := BlockReplacedBox{}
	parent := NewReplacedBox(elementTag, style, replacement)
	out.ReplacedBox = *parent
	return &out
}
func NewInlineReplacedBox(elementTag string, style css.StyleDict, replacement css.ImageType) *InlineReplacedBox {
	out := InlineReplacedBox{}
	parent := NewReplacedBox(elementTag, style, replacement)
	out.ReplacedBox = *parent
	return &out
}
func NewTableBox(elementTag string, style css.StyleDict, children []Box) *TableBox {
	out := TableBox{
		tabularContainer: true,
	}
	parent := NewParentBox(elementTag, style, children)
	out.ParentBox = *parent
	return &out
}
func NewInlineTableBox(elementTag string, style css.StyleDict, children []Box) *InlineTableBox {
	out := InlineTableBox{}
	parent := NewTableBox(elementTag, style, children)
	out.TableBox = *parent
	return &out
}
func NewTableRowGroupBox(elementTag string, style css.StyleDict, children []Box) *TableRowGroupBox {
	out := TableRowGroupBox{
		properTableChild:       true,
		internalTableOrCaption: true,
		tabularContainer:       true,
	}
	parent := NewParentBox(elementTag, style, children)
	out.ParentBox = *parent
	return &out
}
func NewTableRowBox(elementTag string, style css.StyleDict, children []Box) *TableRowBox {
	out := TableRowBox{
		properTableChild:       true,
		internalTableOrCaption: true,
		tabularContainer:       true,
	}
	parent := NewParentBox(elementTag, style, children)
	out.ParentBox = *parent
	return &out
}
func NewTableColumnGroupBox(elementTag string, style css.StyleDict, children []Box) *TableColumnGroupBox {
	out := TableColumnGroupBox{
		properTableChild:       true,
		internalTableOrCaption: true,
		span:                   1,
	}
	parent := NewParentBox(elementTag, style, children)
	out.ParentBox = *parent
	return &out
}
func NewTableColumnBox(elementTag string, style css.StyleDict, children []Box) *TableColumnBox {
	out := TableColumnBox{
		properTableChild:       true,
		internalTableOrCaption: true,
		span:                   1,
	}
	parent := NewParentBox(elementTag, style, children)
	out.ParentBox = *parent
	return &out
}
func NewTableCellBox(elementTag string, style css.StyleDict, children []Box) *TableCellBox {
	out := TableCellBox{
		internalTableOrCaption: true,
		colspan:                1,
		rowspan:                1,
	}
	parent := NewBlockContainerBox(elementTag, style, children)
	out.BlockContainerBox = *parent
	return &out
}
func NewTableCaptionBox(elementTag string, style css.StyleDict, children []Box) *TableCaptionBox {
	out := TableCaptionBox{
		properTableChild:       true,
		internalTableOrCaption: true,
	}
	parent := NewBlockBox(elementTag, style, children)
	out.BlockBox = *parent
	return &out
}
func NewPageBox(elementTag string, style css.StyleDict, children []Box) *PageBox {
	out := PageBox{}
	parent := NewParentBox(elementTag, style, children)
	out.ParentBox = *parent
	return &out
}
func NewMarginBox(elementTag string, style css.StyleDict, children []Box) *MarginBox {
	out := MarginBox{}
	parent := NewBlockContainerBox(elementTag, style, children)
	out.BlockContainerBox = *parent
	return &out
}
