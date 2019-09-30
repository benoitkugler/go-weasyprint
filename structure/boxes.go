package structure

import (
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/utils"
)

//    Classes for all types of boxes in the CSS formatting structure / box model.
//
//    See http://www.w3.org/TR/CSS21/visuren.html
//
//    Names are the same as in CSS 2.1 with the exception of ``TextBox``. In
//    WeasyPrint, any text is in a ``TextBox``. What CSS calls anonymous
//    inline boxes are text boxes but not all text boxes are anonymous
//    inline boxes.
//
//    See http://www.w3.org/TR/CSS21/visuren.html#anonymous
//
//    Abstract classes, should not be instantiated:
//
//    * Box
//    * BlockLevelBox
//    * InlineLevelBox
//    * BlockContainerBox
//    * ReplacedBox
//    * ParentBox
//    * AtomicInlineLevelBox
//
//    Concrete classes:
//
//    * PageBox
//    * BlockBox
//    * InlineBox
//    * InlineBlockBox
//    * BlockReplacedBox
//    * InlineReplacedBox
//    * TextBox
//    * LineBox
//    * Various table-related Box subclasses
//
//    All concrete box classes whose name contains "Inline" or "Block" have
//    one of the following "outside" behavior:
//
//    * Block-level (inherits from :class:`BlockLevelBox`)
//    * Inline-level (inherits from :class:`InlineLevelBox`)
//
//    and one of the following "inside" behavior:
//
//    * Block container (inherits from :class:`BlockContainerBox`)
//    * Inline content (InlineBox and :class:`TextBox`)
//    * Replaced content (inherits from :class:`ReplacedBox`)
//
//    ... with various combinations of both.
//
//    See respective docstrings for details.
//
//    :copyright: Copyright 2011-2014 Simon Sapin and contributors, see AUTHORS.
//    :license: BSD, see LICENSE for details.

// Box is the common interface grouping all possible boxes
// For commodity, we abreviate BoxInstance to Box.
type Box interface {
	Box() *BoxFields
	Copy() Box
	removeDecoration(start, end bool)
	Translate(float32, float32, bool)
}

// ParentBoxInstance represents ParentBox and its descendant
type ParentBoxInstance interface {
	Box
	isParentBoxInstance()
}

// BlockLevelBoxInstance represents BlockLevelBox and its descendant
type BlockLevelBoxInstance interface {
	Box
	isBlockLevelBoxInstance()
}

// BlockContainerBoxInstance represents BlockContainerBox and its descendant
type BlockContainerBoxInstance interface {
	ParentBoxInstance
	isBlockContainerBoxInstance()
}

// InlineLevelBoxInstance represents InlineLevelBox and its descendant
type InlineLevelBoxInstance interface {
	Box
	isInlineLevelBoxInstance()
}

// TableBoxInstance represents TableBox and its descendant
type TableBoxInstance interface {
	BlockLevelBoxInstance
	isParentBoxInstance()
	isTableBoxInstance()
}

type FlexContainerBoxInstance interface {
	ParentBoxInstance
	isFlexContainerBoxInstance()
}

// ---------- Concrete types -------------------------------------

// ParentBox is a box that has children.
type ParentBox struct {
	BoxFields
}

func (ParentBox) isParentBoxInstance() {}

func (b *ParentBox) Box() *BoxFields {
	return &b.BoxFields
}

// BlockLevelBox is a box that participates in an block formatting context.
//An element with a ``display`` weight of ``block``, ``list-item`` or
//``table`` generates a block-level box.
type BlockLevelBox struct {
	clearance interface{}
}

func (BlockLevelBox) isBlockLevelBoxInstance() {}

// BlockContainerBox is a box that contains only block-level boxes or only line boxes.
//
//A box that either contains only block-level boxes or establishes an inline
//formatting context and thus contains only line boxes.
//
//A non-replaced element with a ``display`` weight of ``block``,
//``list-item``, ``inline-block`` or 'table-cell' generates a block container
//box.
type BlockContainerBox struct {
	ParentBox
}

func (BlockContainerBox) isBlockContainerBoxInstance() {}

// BlockBox is a block-level box that is also a block container.
//
//A non-replaced element with a ``display`` weight of ``block``, ``list-item``
//generates a block box.
type BlockBox struct {
	BlockContainerBox
	BlockLevelBox
}

// LineBox is a box that represents a line in an inline formatting context.
//
//Can only contain inline-level boxes.
//
//In early stages of building the box tree a single line box contains many
//consecutive inline boxes. Later, during layout phase, each line boxes will
//be split into multiple line boxes, one for each actual line.
type LineBox struct {
	ParentBox

	textOverflow pr.String // init:"clip"
}

// InlineLevelBox is a box that participates in an inline formatting context.
//
//An inline-level box that is not an inline box is said to be "atomic". Such
//boxes are inline blocks, replaced elements and inline tables.
//
//An element with a ``display`` weight of ``inline``, ``inline-table``, or
//``inline-block`` generates an inline-level box.
type InlineLevelBox struct {
	BoxFields
}

func (InlineLevelBox) isInlineLevelBoxInstance() {}

func (b *InlineLevelBox) Box() *BoxFields {
	return &b.BoxFields
}

// InlineBox is an inline box with inline children.
//
//A box that participates in an inline formatting context and whose content
//also participates in that inline formatting context.
//
//A non-replaced element with a ``display`` weight of ``inline`` generates an
//inline box.
type InlineBox struct {
	InlineLevelBox
	ParentBox
}

// TextBox is a box that contains only text and has no box children.
//
//Any text in the document ends up in a text box. What CSS calls "anonymous
//inline boxes" are also text boxes.
type TextBox struct {
	InlineLevelBox

	justificationSpacing int
	text                 string

	// constructor:elementTag string, style pr.Properties, text string
}

func (b *TextBox) Box() *BoxFields {
	return &b.BoxFields
}

// AtomicInlineLevelBox is an atomic box in an inline formatting context.
// This inline-level box cannot be split for line breaks.
type AtomicInlineLevelBox struct {
	InlineLevelBox
}

// InlineBlockBox is a box that is both inline-level and a block container.
// It behaves as inline on the outside and as a block on the inside.
// A non-replaced element with a 'display' weight of 'inline-block' generates
// an inline-block box.
type InlineBlockBox struct {
	AtomicInlineLevelBox
	BlockContainerBox
}

// ReplacedBox is a box whose content is replaced.
// For example, ``<img>`` are replaced: their content is rendered externally
// and is opaque from CSS’s point of view.
type ReplacedBox struct {
	BoxFields

	replacement pr.Image

	// constructor:elementTag string, style pr.Properties, replacement pr.Image
}

func (b *ReplacedBox) Box() *BoxFields {
	return &b.BoxFields
}

// BlockReplacedBox is a box that is both replaced and block-level.
// A replaced element with a ``display`` weight of ``block``, ``liste-item`` or
//``table`` generates a block-level replaced box.
type BlockReplacedBox struct {
	ReplacedBox
	BlockLevelBox

	// constructor:elementTag string, style pr.Properties, replacement pr.Image
}

// InlineReplacedBox is a box that is both replaced and inline-level.
// A replaced element with a ``display`` weight of ``inline``,
//``inline-table``, or ``inline-block`` generates an inline-level replaced
//box.
type InlineReplacedBox struct {
	ReplacedBox
	AtomicInlineLevelBox

	// constructor:elementTag string, style pr.Properties, replacement pr.Image
}

// TableBox is a box for elements with ``display: table``
type TableBox struct {
	ParentBox
	BlockLevelBox

	//Definitions for the rules generating anonymous table boxes
	//http://www.w3.org/TR/CSS21/tables.html#anonymous-boxes
	tabularContainer bool // init:true

	columnGroups    []Box
	columnPositions []float32
}

func (TableBox) isTableBoxInstance() {}

// InlineTableBox is a box for elements with ``display: inline-table``
type InlineTableBox struct {
	TableBox
}

// TableRowGroupBox is a box for elements with ``display: table-row-group``
type TableRowGroupBox struct {
	ParentBox

	properTableChild       bool // init:true
	internalTableOrCaption bool // init:true
	tabularContainer       bool // init:true
	//properParents = (TableBox, InlineTableBox)

	// Default values. May be overriden on instances.
	isHeader bool
	isFooter bool
}

// TableRowBox is a box for elements with ``display: table-row``
type TableRowBox struct {
	ParentBox

	properTableChild       bool // init:true
	internalTableOrCaption bool // init:true
	tabularContainer       bool // init:true
	//properParents = (TableBox, InlineTableBox, TableRowGroupBox)
}

// TableColumnGroupBox is a box for elements with ``display: table-column-group``
type TableColumnGroupBox struct {
	ParentBox

	properTableChild       bool // init:true
	internalTableOrCaption bool // init:true
	//properParents = (TableBox, InlineTableBox)

	//Default weight. May be overriden on instances.
	span int // init:1

	//Columns groups never have margins or paddings
	marginTop, marginBottom, marginLeft, marginRight     float64
	paddingTop, paddingBottom, paddingLeft, paddingRight float64
}

// Not really a parent box, but pretending to be removes some corner cases.
// TableColumnBox is a box for elements with ``display: table-column``
type TableColumnBox struct {
	ParentBox

	properTableChild       bool // init:true
	internalTableOrCaption bool // init:true
	//properParents = (TableBox, InlineTableBox, TableColumnGroupBox)

	//Default weight. May be overriden on instances.
	span int // init:1

	//Columns groups never have margins or paddings
	marginTop, marginBottom, marginLeft, marginRight     float64
	paddingTop, paddingBottom, paddingLeft, paddingRight float64
}

// TableCellBox is a box for elements with ``display: table-cell``
type TableCellBox struct {
	BlockContainerBox

	internalTableOrCaption bool // init:true
	// Default values. May be overriden on instances.
	colspan int // init:1
	rowspan int // init:1
}

// TableCaptionBox is a box for elements with ``display: table-caption``
type TableCaptionBox struct {
	BlockBox

	properTableChild       bool // init:true
	internalTableOrCaption bool // init:true
	//properParents = (TableBox, InlineTableBox)
}

// PageBox is a box for a page
// Initially the whole document will be in the box for the root element.
//	During layout a new page box is created after every page break.
type PageBox struct {
	ParentBox

	pageType utils.PageElement
}

// MarginBox is a box in page margins, as defined in CSS3 Paged Media
type MarginBox struct {
	BlockContainerBox

	atKeyword string
}

// A box that contains only flex-items.
type FlexContainerBox struct {
	ParentBox
}

func (FlexContainerBox) isFlexContainerBoxInstance() {}

// A box that is both block-level and a flex container.
//
// It behaves as block on the outside and as a flex container on the inside.
type FlexBox struct {
	FlexContainerBox
	BlockLevelBox
}

// A box that is both inline-level and a flex container.
//
// It behaves as inline on the outside and as a flex container on the inside.
type InlineFlexBox struct {
	FlexContainerBox
	InlineLevelBox
}
