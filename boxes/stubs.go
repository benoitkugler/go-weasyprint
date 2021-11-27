package boxes

// Code generated by macros/boxes.py DO NOT EDIT

import "github.com/benoitkugler/go-weasyprint/style/tree"

// An atomic box in an inline formatting context.
// This inline-level box cannot be split for line breaks.
type AtomicInlineLevelBoxITF interface {
	InlineLevelBoxITF
	isAtomicInlineLevelBox()
}

// A block-level box that is also a block container.
// A non-replaced element with a ``display`` value of ``block``, ``list-item``
// generates a block box.
type BlockBoxITF interface {
	BlockContainerBoxITF
	BlockLevelBoxITF
	isBlockBox()
}

func (BlockBox) Type() BoxType        { return BlockBoxT }
func (b *BlockBox) Box() *BoxFields   { return &b.BoxFields }
func (b BlockBox) Copy() Box          { return &b }
func (BlockBox) IsClassicalBox() bool { return true }
func (BlockBox) isBlockBox()          {}
func (BlockBox) isParentBox()         {}
func (BlockBox) isBlockContainerBox() {}
func (BlockBox) isBlockLevelBox()     {}

func BlockBoxAnonymousFrom(parent Box, children []Box) *BlockBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil, nil)
	out := NewBlockBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// A box that contains only block-level boxes or only line boxes.
// A box that either contains only block-level boxes or establishes an inline
// formatting context and thus contains only line boxes.
// A non-replaced element with a ``display`` value of ``block``,
// ``list-item``, ``inline-block`` or 'table-cell' generates a block container
// box.
type BlockContainerBoxITF interface {
	ParentBoxITF
	isBlockContainerBox()
}

// A box that participates in an block formatting context.
// An element with a ``display`` value of ``block``, ``list-item`` or
// ``table`` generates a block-level box.
type BlockLevelBoxITF interface {
	BoxITF
	isBlockLevelBox()
	methodsBlockLevelBox
}

// A box that is both replaced and block-level.
// A replaced element with a ``display`` value of ``block``, ``liste-item`` or
// ``table`` generates a block-level replaced box.
type BlockReplacedBoxITF interface {
	ReplacedBoxITF
	BlockLevelBoxITF
	isBlockReplacedBox()
}

func (BlockReplacedBox) Type() BoxType        { return BlockReplacedBoxT }
func (b *BlockReplacedBox) Box() *BoxFields   { return &b.BoxFields }
func (b BlockReplacedBox) Copy() Box          { return &b }
func (BlockReplacedBox) IsClassicalBox() bool { return true }
func (BlockReplacedBox) isBlockReplacedBox()  {}
func (BlockReplacedBox) isReplacedBox()       {}
func (BlockReplacedBox) isBlockLevelBox()     {}

// A box that is both block-level and a flex container.
// It behaves as block on the outside and as a flex container on the inside.
type FlexBoxITF interface {
	BlockLevelBoxITF
	FlexContainerBoxITF
	isFlexBox()
}

func (FlexBox) Type() BoxType        { return FlexBoxT }
func (b *FlexBox) Box() *BoxFields   { return &b.BoxFields }
func (b FlexBox) Copy() Box          { return &b }
func (FlexBox) IsClassicalBox() bool { return true }
func (FlexBox) isFlexBox()           {}
func (FlexBox) isParentBox()         {}
func (FlexBox) isBlockLevelBox()     {}
func (FlexBox) isFlexContainerBox()  {}

func FlexBoxAnonymousFrom(parent Box, children []Box) *FlexBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil, nil)
	out := NewFlexBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// A box that contains only flex-items.
type FlexContainerBoxITF interface {
	ParentBoxITF
	isFlexContainerBox()
}

// A box that is both inline-level and a block container.
// It behaves as inline on the outside and as a block on the inside.
// A non-replaced element with a 'display' value of 'inline-block' generates
// an inline-block box.
type InlineBlockBoxITF interface {
	BlockContainerBoxITF
	AtomicInlineLevelBoxITF
	isInlineBlockBox()
}

func (InlineBlockBox) Type() BoxType           { return InlineBlockBoxT }
func (b *InlineBlockBox) Box() *BoxFields      { return &b.BoxFields }
func (b InlineBlockBox) Copy() Box             { return &b }
func (InlineBlockBox) IsClassicalBox() bool    { return true }
func (InlineBlockBox) isInlineBlockBox()       {}
func (InlineBlockBox) isInlineLevelBox()       {}
func (InlineBlockBox) isParentBox()            {}
func (InlineBlockBox) isBlockContainerBox()    {}
func (InlineBlockBox) isAtomicInlineLevelBox() {}

func InlineBlockBoxAnonymousFrom(parent Box, children []Box) *InlineBlockBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil, nil)
	out := NewInlineBlockBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// An inline box with inline children.
// A box that participates in an inline formatting context and whose content
// also participates in that inline formatting context.
// A non-replaced element with a ``display`` value of ``inline`` generates an
// inline box.
type InlineBoxITF interface {
	ParentBoxITF
	InlineLevelBoxITF
	isInlineBox()
}

func (InlineBox) Type() BoxType        { return InlineBoxT }
func (b *InlineBox) Box() *BoxFields   { return &b.BoxFields }
func (b InlineBox) Copy() Box          { return &b }
func (InlineBox) IsClassicalBox() bool { return true }
func (InlineBox) isInlineBox()         {}
func (InlineBox) isParentBox()         {}
func (InlineBox) isInlineLevelBox()    {}

func InlineBoxAnonymousFrom(parent Box, children []Box) *InlineBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil, nil)
	out := NewInlineBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// A box that is both inline-level and a flex container.
// It behaves as inline on the outside and as a flex container on the inside.
type InlineFlexBoxITF interface {
	InlineLevelBoxITF
	FlexContainerBoxITF
	isInlineFlexBox()
}

func (InlineFlexBox) Type() BoxType        { return InlineFlexBoxT }
func (b *InlineFlexBox) Box() *BoxFields   { return &b.BoxFields }
func (b InlineFlexBox) Copy() Box          { return &b }
func (InlineFlexBox) IsClassicalBox() bool { return true }
func (InlineFlexBox) isInlineFlexBox()     {}
func (InlineFlexBox) isParentBox()         {}
func (InlineFlexBox) isInlineLevelBox()    {}
func (InlineFlexBox) isFlexContainerBox()  {}

func InlineFlexBoxAnonymousFrom(parent Box, children []Box) *InlineFlexBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil, nil)
	out := NewInlineFlexBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// A box that participates in an inline formatting context.
// An inline-level box that is not an inline box is said to be "atomic". Such
// boxes are inline blocks, replaced elements and inline tables.
// An element with a ``display`` value of ``inline``, ``inline-table``, or
// ``inline-block`` generates an inline-level box.
type InlineLevelBoxITF interface {
	BoxITF
	isInlineLevelBox()
}

// A box that is both replaced and inline-level.
// A replaced element with a ``display`` value of ``inline``,
// ``inline-table``, or ``inline-block`` generates an inline-level replaced
// box.
type InlineReplacedBoxITF interface {
	ReplacedBoxITF
	AtomicInlineLevelBoxITF
	isInlineReplacedBox()
}

func (InlineReplacedBox) Type() BoxType           { return InlineReplacedBoxT }
func (b *InlineReplacedBox) Box() *BoxFields      { return &b.BoxFields }
func (b InlineReplacedBox) Copy() Box             { return &b }
func (InlineReplacedBox) IsClassicalBox() bool    { return true }
func (InlineReplacedBox) isInlineReplacedBox()    {}
func (InlineReplacedBox) isInlineLevelBox()       {}
func (InlineReplacedBox) isReplacedBox()          {}
func (InlineReplacedBox) isAtomicInlineLevelBox() {}

// Box for elements with ``display: inline-table``
type InlineTableBoxITF interface {
	TableBoxITF
	isInlineTableBox()
}

func (InlineTableBox) Type() BoxType        { return InlineTableBoxT }
func (b *InlineTableBox) Box() *BoxFields   { return &b.BoxFields }
func (b InlineTableBox) Copy() Box          { return &b }
func (InlineTableBox) IsClassicalBox() bool { return true }
func (InlineTableBox) isInlineTableBox()    {}
func (InlineTableBox) isParentBox()         {}
func (InlineTableBox) isBlockLevelBox()     {}
func (InlineTableBox) isTableBox()          {}

func InlineTableBoxAnonymousFrom(parent Box, children []Box) *InlineTableBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil, nil)
	out := NewInlineTableBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// A box that represents a line in an inline formatting context.
// Can only contain inline-level boxes.
// In early stages of building the box tree a single line box contains many
// consecutive inline boxes. Later, during layout phase, each line boxes will
// be split into multiple line boxes, one for each actual line.
type LineBoxITF interface {
	ParentBoxITF
	isLineBox()
}

func (LineBox) Type() BoxType        { return LineBoxT }
func (b *LineBox) Box() *BoxFields   { return &b.BoxFields }
func (b LineBox) Copy() Box          { return &b }
func (LineBox) IsClassicalBox() bool { return true }
func (LineBox) isLineBox()           {}
func (LineBox) isParentBox()         {}

// Box in page margins, as defined in CSS3 Paged Media
type MarginBoxITF interface {
	BlockContainerBoxITF
	isMarginBox()
}

func (MarginBox) Type() BoxType        { return MarginBoxT }
func (b *MarginBox) Box() *BoxFields   { return &b.BoxFields }
func (b MarginBox) Copy() Box          { return &b }
func (MarginBox) IsClassicalBox() bool { return true }
func (MarginBox) isMarginBox()         {}
func (MarginBox) isParentBox()         {}
func (MarginBox) isBlockContainerBox() {}

// Box for a page.
// Initially the whole document will be in the box for the root element.
// During layout a new page box is created after every page break.
type PageBoxITF interface {
	ParentBoxITF
	isPageBox()
}

func (PageBox) Type() BoxType        { return PageBoxT }
func (b *PageBox) Box() *BoxFields   { return &b.BoxFields }
func (b PageBox) Copy() Box          { return &b }
func (PageBox) IsClassicalBox() bool { return true }
func (PageBox) isPageBox()           {}
func (PageBox) isParentBox()         {}

// A box that has children.
type ParentBoxITF interface {
	BoxITF
	isParentBox()
}

// A box whose content is replaced.
// For example, ``<img>`` are replaced: their content is rendered externally
// and is opaque from CSS’s point of view.
type ReplacedBoxITF interface {
	BoxITF
	isReplacedBox()
	methodsReplacedBox
}

func (ReplacedBox) Type() BoxType        { return ReplacedBoxT }
func (b *ReplacedBox) Box() *BoxFields   { return &b.BoxFields }
func (b ReplacedBox) Copy() Box          { return &b }
func (ReplacedBox) IsClassicalBox() bool { return true }
func (ReplacedBox) isReplacedBox()       {}

// Box for elements with ``display: table``
type TableBoxITF interface {
	ParentBoxITF
	BlockLevelBoxITF
	isTableBox()
	methodsTableBox
}

func (TableBox) Type() BoxType        { return TableBoxT }
func (b *TableBox) Box() *BoxFields   { return &b.BoxFields }
func (b TableBox) Copy() Box          { return &b }
func (TableBox) IsClassicalBox() bool { return true }
func (TableBox) isTableBox()          {}
func (TableBox) isParentBox()         {}
func (TableBox) isBlockLevelBox()     {}

func TableBoxAnonymousFrom(parent Box, children []Box) *TableBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil, nil)
	out := NewTableBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// Box for elements with ``display: table-caption``
type TableCaptionBoxITF interface {
	BlockBoxITF
	isTableCaptionBox()
}

func (TableCaptionBox) Type() BoxType        { return TableCaptionBoxT }
func (b *TableCaptionBox) Box() *BoxFields   { return &b.BoxFields }
func (b TableCaptionBox) Copy() Box          { return &b }
func (TableCaptionBox) IsClassicalBox() bool { return true }
func (TableCaptionBox) isTableCaptionBox()   {}
func (TableCaptionBox) isParentBox()         {}
func (TableCaptionBox) isBlockContainerBox() {}
func (TableCaptionBox) isBlockBox()          {}
func (TableCaptionBox) isBlockLevelBox()     {}

func TableCaptionBoxAnonymousFrom(parent Box, children []Box) *TableCaptionBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil, nil)
	out := NewTableCaptionBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// Box for elements with ``display: table-cell``
type TableCellBoxITF interface {
	BlockContainerBoxITF
	isTableCellBox()
}

func (TableCellBox) Type() BoxType        { return TableCellBoxT }
func (b *TableCellBox) Box() *BoxFields   { return &b.BoxFields }
func (b TableCellBox) Copy() Box          { return &b }
func (TableCellBox) IsClassicalBox() bool { return true }
func (TableCellBox) isTableCellBox()      {}
func (TableCellBox) isParentBox()         {}
func (TableCellBox) isBlockContainerBox() {}

func TableCellBoxAnonymousFrom(parent Box, children []Box) *TableCellBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil, nil)
	out := NewTableCellBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// Box for elements with ``display: table-column``
type TableColumnBoxITF interface {
	ParentBoxITF
	isTableColumnBox()
}

func (TableColumnBox) Type() BoxType        { return TableColumnBoxT }
func (b *TableColumnBox) Box() *BoxFields   { return &b.BoxFields }
func (b TableColumnBox) Copy() Box          { return &b }
func (TableColumnBox) IsClassicalBox() bool { return true }
func (TableColumnBox) isTableColumnBox()    {}
func (TableColumnBox) isParentBox()         {}

func TableColumnBoxAnonymousFrom(parent Box, children []Box) *TableColumnBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil, nil)
	out := NewTableColumnBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// Box for elements with ``display: table-column-group``
type TableColumnGroupBoxITF interface {
	ParentBoxITF
	isTableColumnGroupBox()
}

func (TableColumnGroupBox) Type() BoxType          { return TableColumnGroupBoxT }
func (b *TableColumnGroupBox) Box() *BoxFields     { return &b.BoxFields }
func (b TableColumnGroupBox) Copy() Box            { return &b }
func (TableColumnGroupBox) IsClassicalBox() bool   { return true }
func (TableColumnGroupBox) isTableColumnGroupBox() {}
func (TableColumnGroupBox) isParentBox()           {}

func TableColumnGroupBoxAnonymousFrom(parent Box, children []Box) *TableColumnGroupBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil, nil)
	out := NewTableColumnGroupBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// Box for elements with ``display: table-row``
type TableRowBoxITF interface {
	ParentBoxITF
	isTableRowBox()
}

func (TableRowBox) Type() BoxType        { return TableRowBoxT }
func (b *TableRowBox) Box() *BoxFields   { return &b.BoxFields }
func (b TableRowBox) Copy() Box          { return &b }
func (TableRowBox) IsClassicalBox() bool { return true }
func (TableRowBox) isTableRowBox()       {}
func (TableRowBox) isParentBox()         {}

func TableRowBoxAnonymousFrom(parent Box, children []Box) *TableRowBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil, nil)
	out := NewTableRowBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// Box for elements with ``display: table-row-group``
type TableRowGroupBoxITF interface {
	ParentBoxITF
	isTableRowGroupBox()
}

func (TableRowGroupBox) Type() BoxType        { return TableRowGroupBoxT }
func (b *TableRowGroupBox) Box() *BoxFields   { return &b.BoxFields }
func (b TableRowGroupBox) Copy() Box          { return &b }
func (TableRowGroupBox) IsClassicalBox() bool { return true }
func (TableRowGroupBox) isTableRowGroupBox()  {}
func (TableRowGroupBox) isParentBox()         {}

func TableRowGroupBoxAnonymousFrom(parent Box, children []Box) *TableRowGroupBox {
	style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil, nil)
	out := NewTableRowGroupBox(style, parent.Box().Element, parent.Box().PseudoType, children)
	return out
}

// A box that contains only text and has no box children.
// Any text in the document ends up in a text box. What CSS calls "anonymous
// inline boxes" are also text boxes.
type TextBoxITF interface {
	InlineLevelBoxITF
	isTextBox()
}

func (TextBox) Type() BoxType        { return TextBoxT }
func (b *TextBox) Box() *BoxFields   { return &b.BoxFields }
func (b TextBox) Copy() Box          { return &b }
func (TextBox) IsClassicalBox() bool { return true }
func (TextBox) isTextBox()           {}
func (TextBox) isInlineLevelBox()    {}

// BoxType represents a box type.
type BoxType uint8

const (
	invalidType BoxType = iota
	AtomicInlineLevelBoxT
	BlockBoxT
	BlockContainerBoxT
	BlockLevelBoxT
	BlockReplacedBoxT
	BoxT
	FlexBoxT
	FlexContainerBoxT
	InlineBlockBoxT
	InlineBoxT
	InlineFlexBoxT
	InlineLevelBoxT
	InlineReplacedBoxT
	InlineTableBoxT
	LineBoxT
	MarginBoxT
	PageBoxT
	ParentBoxT
	ReplacedBoxT
	TableBoxT
	TableCaptionBoxT
	TableCellBoxT
	TableColumnBoxT
	TableColumnGroupBoxT
	TableRowBoxT
	TableRowGroupBoxT
	TextBoxT
)

// Returns true is the box is an instance of t.
func (t BoxType) IsInstance(box BoxITF) bool {
	var isInstance bool
	switch t {
	case AtomicInlineLevelBoxT:
		_, isInstance = box.(AtomicInlineLevelBoxITF)
	case BlockBoxT:
		_, isInstance = box.(BlockBoxITF)
	case BlockContainerBoxT:
		_, isInstance = box.(BlockContainerBoxITF)
	case BlockLevelBoxT:
		_, isInstance = box.(BlockLevelBoxITF)
	case BlockReplacedBoxT:
		_, isInstance = box.(BlockReplacedBoxITF)
	case BoxT:
		_, isInstance = box.(BoxITF)
	case FlexBoxT:
		_, isInstance = box.(FlexBoxITF)
	case FlexContainerBoxT:
		_, isInstance = box.(FlexContainerBoxITF)
	case InlineBlockBoxT:
		_, isInstance = box.(InlineBlockBoxITF)
	case InlineBoxT:
		_, isInstance = box.(InlineBoxITF)
	case InlineFlexBoxT:
		_, isInstance = box.(InlineFlexBoxITF)
	case InlineLevelBoxT:
		_, isInstance = box.(InlineLevelBoxITF)
	case InlineReplacedBoxT:
		_, isInstance = box.(InlineReplacedBoxITF)
	case InlineTableBoxT:
		_, isInstance = box.(InlineTableBoxITF)
	case LineBoxT:
		_, isInstance = box.(LineBoxITF)
	case MarginBoxT:
		_, isInstance = box.(MarginBoxITF)
	case PageBoxT:
		_, isInstance = box.(PageBoxITF)
	case ParentBoxT:
		_, isInstance = box.(ParentBoxITF)
	case ReplacedBoxT:
		_, isInstance = box.(ReplacedBoxITF)
	case TableBoxT:
		_, isInstance = box.(TableBoxITF)
	case TableCaptionBoxT:
		_, isInstance = box.(TableCaptionBoxITF)
	case TableCellBoxT:
		_, isInstance = box.(TableCellBoxITF)
	case TableColumnBoxT:
		_, isInstance = box.(TableColumnBoxITF)
	case TableColumnGroupBoxT:
		_, isInstance = box.(TableColumnGroupBoxITF)
	case TableRowBoxT:
		_, isInstance = box.(TableRowBoxITF)
	case TableRowGroupBoxT:
		_, isInstance = box.(TableRowGroupBoxITF)
	case TextBoxT:
		_, isInstance = box.(TextBoxITF)
	}
	return isInstance
}

func (t BoxType) String() string {
	switch t {
	case AtomicInlineLevelBoxT:
		return "AtomicInlineLevelBox"
	case BlockBoxT:
		return "BlockBox"
	case BlockContainerBoxT:
		return "BlockContainerBox"
	case BlockLevelBoxT:
		return "BlockLevelBox"
	case BlockReplacedBoxT:
		return "BlockReplacedBox"
	case BoxT:
		return "Box"
	case FlexBoxT:
		return "FlexBox"
	case FlexContainerBoxT:
		return "FlexContainerBox"
	case InlineBlockBoxT:
		return "InlineBlockBox"
	case InlineBoxT:
		return "InlineBox"
	case InlineFlexBoxT:
		return "InlineFlexBox"
	case InlineLevelBoxT:
		return "InlineLevelBox"
	case InlineReplacedBoxT:
		return "InlineReplacedBox"
	case InlineTableBoxT:
		return "InlineTableBox"
	case LineBoxT:
		return "LineBox"
	case MarginBoxT:
		return "MarginBox"
	case PageBoxT:
		return "PageBox"
	case ParentBoxT:
		return "ParentBox"
	case ReplacedBoxT:
		return "ReplacedBox"
	case TableBoxT:
		return "TableBox"
	case TableCaptionBoxT:
		return "TableCaptionBox"
	case TableCellBoxT:
		return "TableCellBox"
	case TableColumnBoxT:
		return "TableColumnBox"
	case TableColumnGroupBoxT:
		return "TableColumnGroupBox"
	case TableRowBoxT:
		return "TableRowBox"
	case TableRowGroupBoxT:
		return "TableRowGroupBox"
	case TextBoxT:
		return "TextBox"
	}
	return "<invalid box type>"
}

var (
	_ BlockBoxITF            = (*BlockBox)(nil)
	_ BlockReplacedBoxITF    = (*BlockReplacedBox)(nil)
	_ FlexBoxITF             = (*FlexBox)(nil)
	_ InlineBlockBoxITF      = (*InlineBlockBox)(nil)
	_ InlineBoxITF           = (*InlineBox)(nil)
	_ InlineFlexBoxITF       = (*InlineFlexBox)(nil)
	_ InlineReplacedBoxITF   = (*InlineReplacedBox)(nil)
	_ InlineTableBoxITF      = (*InlineTableBox)(nil)
	_ LineBoxITF             = (*LineBox)(nil)
	_ MarginBoxITF           = (*MarginBox)(nil)
	_ PageBoxITF             = (*PageBox)(nil)
	_ ReplacedBoxITF         = (*ReplacedBox)(nil)
	_ TableBoxITF            = (*TableBox)(nil)
	_ TableCaptionBoxITF     = (*TableCaptionBox)(nil)
	_ TableCellBoxITF        = (*TableCellBox)(nil)
	_ TableColumnBoxITF      = (*TableColumnBox)(nil)
	_ TableColumnGroupBoxITF = (*TableColumnGroupBox)(nil)
	_ TableRowBoxITF         = (*TableRowBox)(nil)
	_ TableRowGroupBoxITF    = (*TableRowGroupBox)(nil)
	_ TextBoxITF             = (*TextBox)(nil)
)

func (t BoxType) AnonymousFrom(parent Box, children []Box) Box {
	switch t {
	case BlockBoxT:
		return BlockBoxAnonymousFrom(parent, children)
	case FlexBoxT:
		return FlexBoxAnonymousFrom(parent, children)
	case InlineBlockBoxT:
		return InlineBlockBoxAnonymousFrom(parent, children)
	case InlineBoxT:
		return InlineBoxAnonymousFrom(parent, children)
	case InlineFlexBoxT:
		return InlineFlexBoxAnonymousFrom(parent, children)
	case InlineTableBoxT:
		return InlineTableBoxAnonymousFrom(parent, children)
	case TableBoxT:
		return TableBoxAnonymousFrom(parent, children)
	case TableCaptionBoxT:
		return TableCaptionBoxAnonymousFrom(parent, children)
	case TableCellBoxT:
		return TableCellBoxAnonymousFrom(parent, children)
	case TableColumnBoxT:
		return TableColumnBoxAnonymousFrom(parent, children)
	case TableColumnGroupBoxT:
		return TableColumnGroupBoxAnonymousFrom(parent, children)
	case TableRowBoxT:
		return TableRowBoxAnonymousFrom(parent, children)
	case TableRowGroupBoxT:
		return TableRowGroupBoxAnonymousFrom(parent, children)
	}
	return nil
}
