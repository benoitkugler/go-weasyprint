package structure

import "github.com/benoitkugler/go-weasyprint/css"

// AllBox unifies all box types
type AllBox interface {
	BaseBox() *Box             // common parts of all boxes
	TableFields() *TableFields // fields for table boxes. Might be nil on onther boxes

	IsParentBox() bool
	IsProperChild(parent AllBox) bool
	IsTableBox() bool

	Copy() AllBox
	copyWithChildren(newChildren []AllBox, isStart, isEnd bool) ParentBox
}

// Box is an abstract base class for all boxes.
type Box struct {
	// Keep track of removed collapsing spaces for wrap opportunities.
	leadingCollapsibleSpace  bool
	trailingCollapsibleSpace bool

	// Default, may be overriden on instances.
	isTableWrapper       bool
	isForRootElement     bool
	isColumn             bool
	transformationMatrix TBD
	bookmarkLabel        TBD
	stringSet            TBD

	elementTag string
	style      css.StyleDict

	positionX, positionY float64

	width, height float64

	marginTop, marginBottom, marginLeft, marginRight float64

	paddingTop, paddingBottom, paddingLeft, paddingRight float64

	borderTopWidth, borderRightWidth, borderBottomWidth, borderLeftWidth float64

	borderTopLeftRadius, borderTopRightRadius, borderBottomRightRadius, borderBottomLeftRadius point

	children []AllBox
}

// ParentBox is a box that has children.
type ParentBox struct {
	Box

	outsideListMarker AllBox
}

// BlockLevelBox is a box that participates in an block formatting context.
//An element with a ``display`` weight of ``block``, ``list-item`` or
//``table`` generates a block-level box.
type BlockLevelBox struct {
	clearance TBD
}

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
}

// InlineLevelBox is a box that participates in an inline formatting context.
//
//An inline-level box that is not an inline box is said to be "atomic". Such
//boxes are inline blocks, replaced elements and inline tables.
//
//An element with a ``display`` weight of ``inline``, ``inline-table``, or
//``inline-block`` generates an inline-level box.
type InlineLevelBox struct {
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
	Box
	InlineLevelBox

	justificationSpacing int
	text                 string
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
	Box

	replacement TBD
}

// BlockReplacedBox is a box that is both replaced and block-level.
// A replaced element with a ``display`` weight of ``block``, ``liste-item`` or
//``table`` generates a block-level replaced box.
type BlockReplacedBox struct {
	ReplacedBox
	BlockLevelBox
}

// InlineReplacedBox is a box that is both replaced and inline-level.
// A replaced element with a ``display`` weight of ``inline``,
//``inline-table``, or ``inline-block`` generates an inline-level replaced
//box.
type InlineReplacedBox struct {
	ReplacedBox
	AtomicInlineLevelBox
}

// TableBox is a box for elements with ``display: table``
type TableBox struct {
	ParentBox
	BlockLevelBox

	tableFields TableFields
}

// InlineTableBox is a box for elements with ``display: inline-table``
type InlineTableBox struct {
	TableBox
}

// TableRowGroupBox is a box for elements with ``display: table-row-group``
type TableRowGroupBox struct {
	ParentBox

	tableFields TableFields
	//properParents = (TableBox, InlineTableBox)
}

// TableRowBox is a box for elements with ``display: table-row``
type TableRowBox struct {
	ParentBox

	tableFields TableFields
	//properParents = (TableBox, InlineTableBox, TableRowGroupBox)
}

// TableColumnGroupBox is a box for elements with ``display: table-column-group``
type TableColumnGroupBox struct {
	ParentBox

	tableFields TableFields

	//properParents = (TableBox, InlineTableBox)
}

// Not really a parent box, but pretending to be removes some corner cases.
// TableColumnBox is a box for elements with ``display: table-column``
type TableColumnBox struct {
	ParentBox

	tableFields TableFields

	//properParents = (TableBox, InlineTableBox, TableColumnGroupBox)
}

// TableCellBox is a box for elements with ``display: table-cell``
type TableCellBox struct {
	BlockContainerBox

	tableFields TableFields
}

// TableCaptionBox is a box for elements with ``display: table-caption``
type TableCaptionBox struct {
	BlockBox

	tableFields TableFields
	//properParents = (TableBox, InlineTableBox)
}

// PageBox is a box for a page
// Initially the whole document will be in the box for the root element.
//	During layout a new page box is created after every page break.
type PageBox struct {
	ParentBox

	pageType TBD
}

// MarginBox is a box in page margins, as defined in CSS3 Paged Media
type MarginBox struct {
	BlockContainerBox

	atKeyword TBD
}

func (b ParentBox) Copy() AllBox           { return &b }
func (b BlockContainerBox) Copy() AllBox   { return &b }
func (b BlockBox) Copy() AllBox            { return &b }
func (b LineBox) Copy() AllBox             { return &b }
func (b InlineBox) Copy() AllBox           { return &b }
func (b TextBox) Copy() AllBox             { return &b }
func (b InlineBlockBox) Copy() AllBox      { return &b }
func (b ReplacedBox) Copy() AllBox         { return &b }
func (b BlockReplacedBox) Copy() AllBox    { return &b }
func (b InlineReplacedBox) Copy() AllBox   { return &b }
func (b TableBox) Copy() AllBox            { return &b }
func (b InlineTableBox) Copy() AllBox      { return &b }
func (b TableRowGroupBox) Copy() AllBox    { return &b }
func (b TableRowBox) Copy() AllBox         { return &b }
func (b TableColumnGroupBox) Copy() AllBox { return &b }
func (b TableColumnBox) Copy() AllBox      { return &b }
func (b TableCellBox) Copy() AllBox        { return &b }
func (b TableCaptionBox) Copy() AllBox     { return &b }
func (b PageBox) Copy() AllBox             { return &b }
func (b MarginBox) Copy() AllBox           { return &b }
