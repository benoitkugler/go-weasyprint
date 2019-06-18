package structure

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

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/labstack/gommon/log"

	"github.com/benoitkugler/go-weasyprint/css"
)

// http://stackoverflow.com/questions/16317534/
var asciiToWide = map[rune]string{}

func init() {
	for i := 33; i < 127; i++ {
		asciiToWide[rune(i)] = string(i + 0xfee0)
	}
	asciiToWide[0x20] = "\u3000"
	asciiToWide[0x2D] = "\u2212"
}

type TBD struct{}

type point struct {
	x, y float64
}

// AllBox unifies all box types
type AllBox interface {
	SetIsForRootElement(b bool)
	Style() css.StyleDict

	Translate(dx, dy float64, ignoreFloats bool)
	paddingWidth() float64
	paddingHeight() float64
	borderWidth() float64
	borderHeight() float64
	marginWidth() float64
	marginHeight() float64
	contentBoxX() float64
	contentBoxY() float64
	paddingBoxX() float64
	paddingBoxY() float64
	borderBoxX() float64
	borderBoxY() float64
	hitArea() (x float64, y float64, w float64, h float64)
	roundedBox(bt, br, bb, bl float64) roundedBox
	roundedBoxRatio(ratio float64) roundedBox
	roundedPaddingBox() roundedBox
	roundedBorderBox() roundedBox
	roundedContentBox() roundedBox
	isFloated() bool
	isAbsolutelyPositioned() bool
	isInNormalFlow() bool
	pageValues() (int, int)
	resetSpacing(side css.Side)
}

// Box is an abstract base class for all boxes.
type Box struct {
	//Definitions for the rules generating anonymous table boxes
	//http://www.w3.org/TR/CSS21/tables.html#anonymous-boxes
	properTableChild       bool
	internalTableOrCaption bool
	tabularContainer       bool

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

	elementTag TBD
	style      css.StyleDict

	positionX, positionY float64

	width, height float64

	marginTop, marginBottom, marginLeft, marginRight float64

	paddingTop, paddingBottom, paddingLeft, paddingRight float64

	borderTopWidth, borderRightWidth, borderBottomWidth, borderLeftWidth float64

	borderTopLeftRadius, borderTopRightRadius, borderBottomRightRadius, borderBottomLeftRadius point

	children []AllBox
}

func (self *Box) init(elementTag TBD, style css.StyleDict) {
	self.elementTag = elementTag
	self.style = style
}

func (self Box) String() string {
	return fmt.Sprintf("<Box %s>", self.elementTag)
}

func (self *Box) SetIsForRootElement(b bool) {
	self.isForRootElement = b
}

func (self Box) Style() css.StyleDict {
	return self.style
}

// Translate changes the box’s position.
// Also update the children’s positions accordingly.
func (self *Box) Translate(dx, dy float64, ignoreFloats bool) {
	// Overridden in ParentBox to also translate children, if any.
	if dx == 0 && dy == 0 {
		return
	}
	self.positionX += dx
	self.positionY += dy
	for _, child := range self.AllChildren() {
		if !(ignoreFloats && child.isFloated()) {
			child.Translate(dx, dy, ignoreFloats)
		}
	}
}

// ---- Heights and widths -----

// Width of the padding box.
func (self Box) paddingWidth() float64 {
	return self.width + self.paddingLeft + self.paddingRight
}

// Height of the padding box.
func (self Box) paddingHeight() float64 {
	return self.height + self.paddingTop + self.paddingBottom
}

// Width of the border box.
func (self Box) borderWidth() float64 {
	return self.paddingWidth() + self.borderLeftWidth + self.borderRightWidth
}

// Height of the border box.
func (self Box) borderHeight() float64 {
	return self.paddingHeight() + self.borderTopWidth + self.borderBottomWidth
}

// Width of the margin box (aka. outer box).
func (self Box) marginWidth() float64 {
	return self.borderWidth() + self.marginLeft + self.marginRight
}

// Height of the margin box (aka. outer box).
func (self Box) marginHeight() float64 {
	return self.borderHeight() + self.marginTop + self.marginBottom
}

// Corners positions

// Absolute horizontal position of the content box.
func (self Box) contentBoxX() float64 {
	return self.positionX + self.marginLeft + self.paddingLeft + self.borderLeftWidth
}

// Absolute vertical position of the content box.
func (self Box) contentBoxY() float64 {
	return self.positionY + self.marginTop + self.paddingTop + self.borderTopWidth
}

// Absolute horizontal position of the padding box.
func (self Box) paddingBoxX() float64 {
	return self.positionX + self.marginLeft + self.borderLeftWidth
}

// Absolute vertical position of the padding box.
func (self Box) paddingBoxY() float64 {
	return self.positionY + self.marginTop + self.borderTopWidth
}

// Absolute horizontal position of the border box.
func (self Box) borderBoxX() float64 {
	return self.positionX + self.marginLeft
}

// Absolute vertical position of the border box.
func (self Box) borderBoxY() float64 {
	return self.positionY + self.marginTop
}

// Return the rectangle where the box is clickable."""
// "Border area. That's the area that hit-testing is done on."
// http://lists.w3.org/Archives/Public/www-style/2012Jun/0318.html
// TODO: manage the border radii, use outerBorderRadii instead
func (self Box) hitArea() (x float64, y float64, w float64, h float64) {
	return self.borderBoxX(), self.borderBoxY(), self.borderWidth(), self.borderHeight()
}

type roundedBox struct {
	x, y, width, height                        float64
	topLeft, topRight, bottomRight, bottomLeft point
}

// Position, size and radii of a box inside the outer border box.
//bt, br, bb, and bl are distances from the outer border box,
//defining a rectangle to be rounded.
func (self Box) roundedBox(bt, br, bb, bl float64) roundedBox {
	tlr := self.borderTopLeftRadius
	trr := self.borderTopRightRadius
	brr := self.borderBottomRightRadius
	blr := self.borderBottomLeftRadius

	tlrx := math.Max(0, tlr.x-bl)
	tlry := math.Max(0, tlr.y-bt)
	trrx := math.Max(0, trr.x-br)
	trry := math.Max(0, trr.y-bt)
	brrx := math.Max(0, brr.x-br)
	brry := math.Max(0, brr.y-bb)
	blrx := math.Max(0, blr.x-bl)
	blry := math.Max(0, blr.y-bb)

	x := self.borderBoxX() + bl
	y := self.borderBoxY() + bt
	width := self.borderWidth() - bl - br
	height := self.borderHeight() - bt - bb

	// Fix overlapping curves
	//See http://www.w3.org/TR/css3-background/#corner-overlap
	points := []point{
		{width, tlrx + trrx},
		{width, blrx + brrx},
		{height, tlry + blry},
		{height, trry + brry},
	}
	ratio := 1.
	for _, point := range points {
		if point.y > 0 {
			candidat := point.x / point.y
			if candidat < ratio {
				ratio = candidat
			}
		}
	}
	return roundedBox{x: x, y: y, width: width, height: height,
		topLeft:     point{x: tlrx * ratio, y: tlry * ratio},
		topRight:    point{x: trrx * ratio, y: trry * ratio},
		bottomRight: point{x: brrx * ratio, y: brry * ratio},
		bottomLeft:  point{x: blrx * ratio, y: blry * ratio},
	}
}

func (self Box) roundedBoxRatio(ratio float64) roundedBox {
	return self.roundedBox(
		self.borderTopWidth*ratio,
		self.borderRightWidth*ratio,
		self.borderBottomWidth*ratio,
		self.borderLeftWidth*ratio)
}

// Return the position, size and radii of the rounded padding box.
func (self Box) roundedPaddingBox() roundedBox {
	return self.roundedBox(
		self.borderTopWidth,
		self.borderRightWidth,
		self.borderBottomWidth,
		self.borderLeftWidth)
}

// Return the position, size and radii of the rounded border box.
func (self Box) roundedBorderBox() roundedBox {
	return self.roundedBox(0, 0, 0, 0)
}

// Return the position, size and radii of the rounded content box.
func (self Box) roundedContentBox() roundedBox {
	return self.roundedBox(
		self.borderTopWidth+self.paddingTop,
		self.borderRightWidth+self.paddingRight,
		self.borderBottomWidth+self.paddingBottom,
		self.borderLeftWidth+self.paddingLeft)

}

// Positioning schemes

// Return whether this box is floated.
func (self Box) isFloated() bool {
	return self.style["float"].(string) != "none"
}

// Return whether this box is in the absolute positioning scheme.
func (self Box) isAbsolutelyPositioned() bool {
	pos := self.style["position"].(string)
	return pos == "absolute" || pos == "fixed"
}

// Return whether this box is in normal flow.
func (self Box) isInNormalFlow() bool {
	return !(self.isFloated() || self.isAbsolutelyPositioned())
}

// Start and end page values for named pages

// Return start and end page values.
func (self Box) pageValues() (int, int) {
	p := self.style["page"].(int)
	return p, p
}

// Set to 0 the margin, padding and border of ``side``.
func (self *Box) resetSpacing(side css.Side) {
	self.style[fmt.Sprintf("margin_%s", side)] = css.Dimension{Unit: "px"}
	self.style[fmt.Sprintf("padding_%s", side)] = css.Dimension{Unit: "px"}
	self.style[fmt.Sprintf("border_%s_width", side)] = 0

	switch side {
	case css.Top:
		self.marginTop = 0
		self.paddingTop = 0
		self.borderTopWidth = 0
	case css.Right:
		self.marginRight = 0
		self.paddingRight = 0
		self.borderRightWidth = 0
	case css.Left:
		self.marginLeft = 0
		self.paddingLeft = 0
		self.borderLeftWidth = 0
	case css.Bottom:
		self.marginBottom = 0
		self.paddingBottom = 0
		self.borderBottomWidth = 0
	}
}

func (self *Box) removeDecoration(start, end bool) {
	if start || end {
		self.style = self.style.Copy()
	}
	ltr := self.style["direction"].(string) == "ltr"
	if start {
		side := css.Right
		if ltr {
			side = css.Left
		}
		self.resetSpacing(side)
	}
	if end {
		side := css.Left
		if ltr {
			side = css.Right
		}
		self.resetSpacing(side)
	}
}

func (self Box) AllChildren() []AllBox {
	return self.children
}

// ParentBox is a box that has children.
type ParentBox struct {
	Box

	outsideListMarker *Box
}

func (p *ParentBox) init(elementTag TBD, style css.StyleDict, children []AllBox) {
	p.Box.init(elementTag, style)
	p.children = children
}

// func (self ParentBox) IsTableBox() bool {
// 	return false
// }

func (self *ParentBox) removeDecoration(start, end bool) {
	if start || end {
		self.style = self.style.Copy()
	}
	if start {
		self.resetSpacing(css.Top)
	}
	if end {
		self.resetSpacing(css.Bottom)
	}
}

// Create a new equivalent box with given ``newChildren``.
func (self ParentBox) copyWithChildren(newChildren []AllBox, isStart, isEnd bool) ParentBox {
	newBox := self
	newBox.children = newChildren
	if !isStart {
		newBox.outsideListMarker = nil
	}
	newBox.removeDecoration(!isStart, !isEnd)
	return newBox
}

//// A flat generator for a box, its chien and descendants."""
//func (self ParentBox) descendants(self) {
//	yield self
//	for child in self.children:
//	if hasattr(child, 'descendants'}
//	for grandChild in child.descendants(}
//	yield grandChild
//	else:
//	yield child
//}

// Get the table wrapped by the box.
// Warning, might be nil
func (self ParentBox) getWrappedTable() (*TableBox, error) {
	if self.isTableWrapper {
		for _, child := range self.children {
			if typedChild, ok := child.(*TableBox); ok {
				return typedChild, nil
			}
		}
		return nil, errors.New("Table wrapper without a table")
	}
	return nil, nil
}

func (self ParentBox) pageValues() (int, int) {
	start, end := self.Box.pageValues()
	if len(self.children) > 0 {
		startBox, endBox := self.children[0], self.children[len(self.children)-1]
		childStart, _ := startBox.pageValues()
		_, childEnd := endBox.pageValues()
		if childStart > 0 {
			start = childStart
		}
		if childEnd > 0 {
			end = childEnd
		}
	}
	return start, end
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

func NewBlockBox(elementTag TBD, style css.StyleDict, children []AllBox) *BlockBox {
	var out BlockBox
	out.init(elementTag, style, children)
	return &out
}

func (self BlockBox) AllChildren() []AllBox {
	if self.outsideListMarker != nil {
		return append(self.children, self.outsideListMarker)
	}
	return self.children
}

// func (self BlockBox) pageValues() (int, int) {
// 	return self.BlockContainerBox.pageValues()
// }

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

func (l *LineBox) init(elementTag TBD, style css.StyleDict, children []AllBox) {
	if !style.Anonymous() {
		log.Fatal("style must be anonymous")
	}
	l.ParentBox.init(elementTag, style, children)
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

func NewInlineBox(elementTag TBD, style css.StyleDict, children []AllBox) *InlineBox {
	var out InlineBox
	out.init(elementTag, style, children)
	return &out
}

// Return the (x, y, w, h) rectangle where the box is clickable.
func (self InlineBox) hitArea() (x float64, y float64, w float64, h float64) {
	return self.borderBoxX(), self.positionY, self.borderWidth(), self.marginHeight()
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

func (self *TextBox) init(elementTag TBD, style css.StyleDict, text string) {
	if !style.Anonymous() {
		log.Fatal("style is not anonymous")
	}
	if len(text) == 0 {
		log.Fatal("empty text")
	}
	self.Box.init(elementTag, style)
	textTransform := style["text-transform"].(string)
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
			var chars []string
			for _, u := range []rune(text) {
				chars = append(chars, asciiToWide[u])
			}
			text = strings.Join(chars, "")
		}

		if style["hyphens"].(string) == "none" {
			text = strings.ReplaceAll(text, "\u00AD", "") //  U+00AD SOFT HYPHEN (SHY)
		}
	}
	self.text = text
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

func NewInlineBlockBox(elementTag TBD, style css.StyleDict, children []AllBox) *InlineBlockBox {
	var out InlineBlockBox
	out.init(elementTag, style, children)
	return &out
}

// ReplacedBox is a box whose content is replaced.
// For example, ``<img>`` are replaced: their content is rendered externally
// and is opaque from CSS’s point of view.
type ReplacedBox struct {
	Box

	replacement TBD
}

func (self *ReplacedBox) init(elementTag TBD, style css.StyleDict, replacement TBD) {
	self.Box.init(elementTag, style)
	self.replacement = replacement
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
	BlockLevelBox
	ParentBox

	tabularContainer bool // default is true
	columnGroups     []AllBox
	columnPositions  []float64
}

func NewTableBox(elementTag TBD, style css.StyleDict, children []AllBox) *TableBox {
	var out TableBox
	out.init(elementTag, style, children)
	out.tabularContainer = true
	return &out
}

// Definitions for the rules generating anonymous table boxes
// http://www.w3.org/TR/CSS21/tables.html#anonymous-boxes

func (self TableBox) AllChildren() []AllBox {
	return append(self.children, self.columnGroups...)
}

func (self *TableBox) Translate(dx, dy float64, ignoreFloats bool) {
	if dx == 0 && dy == 0 {
		return
	}
	for index, position := range self.columnPositions {
		self.columnPositions[index] = position + dx
	}
	self.ParentBox.Box.Translate(dx, dy, ignoreFloats)
}

func (self TableBox) pageValues() (int, int) {
	return self.ParentBox.Box.pageValues()
}

// InlineTableBox is a box for elements with ``display: inline-table``
type InlineTableBox struct {
	TableBox
}

func NewInlineTableBox(elementTag TBD, style css.StyleDict, children []AllBox) *InlineTableBox {
	return &InlineTableBox{*NewTableBox(elementTag, style, children)}
}

// TableRowGroupBox is a box for elements with ``display: table-row-group``
type TableRowGroupBox struct {
	ParentBox
	properTableChild       bool // default weight true
	internalTableOrCaption bool // default weight true
	tabularContainer       bool // default weight true

	//properParents = (TableBox, InlineTableBox)

	// Default values. May be overriden on instances.
	isHeader bool
	isFooter bool
}

func NewTableRowGroupBox(elementTag TBD, style css.StyleDict, children []AllBox) *TableRowGroupBox {
	var out TableRowGroupBox
	out.init(elementTag, style, children)
	out.properTableChild = true
	out.internalTableOrCaption = true
	out.tabularContainer = true
	return &out
}

// TableRowBox is a box for elements with ``display: table-row``
type TableRowBox struct {
	ParentBox
	properTableChild       bool // default weight true
	internalTableOrCaption bool // default weight true
	tabularContainer       bool // default weight true
	//properParents = (TableBox, InlineTableBox, TableRowGroupBox)
}

func NewTableRowBox(elementTag TBD, style css.StyleDict, children []AllBox) *TableRowBox {
	var out TableRowBox
	out.init(elementTag, style, children)
	out.properTableChild = true
	out.internalTableOrCaption = true
	out.tabularContainer = true
	return &out
}

// TableColumnGroupBox is a box for elements with ``display: table-column-group``
type TableColumnGroupBox struct {
	ParentBox

	properTableChild       bool // default weight true
	internalTableOrCaption bool // default weight true
	//properParents = (TableBox, InlineTableBox)

	//Columns groups never have margins or paddings
	marginTop, marginBottom, marginLeft, marginRight     float64
	paddingTop, paddingBottom, paddingLeft, paddingRight float64

	//Default weight. May be overriden on instances.
	span int // default weight 1
}

func NewTableColumnGroupBox(elementTag TBD, style css.StyleDict, children []AllBox) *TableColumnGroupBox {
	var out TableColumnGroupBox
	out.init(elementTag, style, children)
	out.span = 1
	out.properTableChild = true
	out.internalTableOrCaption = true
	return &out
}

// Return cells that originate in the group's columns.
func (self TableColumnGroupBox) getCells() []AllBox {
	var out []AllBox
	for _, column := range self.children {
		switch child := column.(type) {
		case *TableColumnGroupBox:
			for _, cell := range child.getCells() {
				out = append(out, cell)
			}
		case *TableColumnBox:
			for _, cell := range child.getCells() {
				out = append(out, cell)
			}
		default:
			panic("Only Box with getCells() method allowed in children of TableColumnGroupBox")
		}
	}
	return out
}

// Not really a parent box, but pretending to be removes some corner cases.
// TableColumnBox is a box for elements with ``display: table-column``
type TableColumnBox struct {
	ParentBox

	properTableChild       bool // default weight true
	internalTableOrCaption bool // default weight true
	//properParents = (TableBox, InlineTableBox, TableColumnGroupBox)

	//Columns never have margins or paddings
	marginTop, marginBottom, marginLeft, marginRight     float64
	paddingTop, paddingBottom, paddingLeft, paddingRight float64

	//Default weight. May be overriden on instances.
	span int // default weight 1
}

func NewTableColumnBox(elementTag TBD, style css.StyleDict, children []AllBox) *TableColumnBox {
	var out TableColumnBox
	out.init(elementTag, style, children)
	out.span = 1
	out.properTableChild = true
	out.internalTableOrCaption = true

	return &out
}

// Return cells that originate in the column.
// May be overriden on instances.
func (self TableColumnBox) getCells() []AllBox {
	return nil
}

// TableCellBox is a box for elements with ``display: table-cell``
type TableCellBox struct {
	BlockContainerBox

	internalTableOrCaption bool // default weight true

	// Default values. May be overriden on instances.
	colspan int // default weight 1
	rowspan int // default weight 1
}

func NewTableCellBox(elementTag TBD, style css.StyleDict, children []AllBox) *TableCellBox {
	var out TableCellBox
	out.init(elementTag, style, children)
	out.colspan = 1
	out.rowspan = 1
	out.internalTableOrCaption = true
	return &out
}

// TableCaptionBox is a box for elements with ``display: table-caption``
type TableCaptionBox struct {
	BlockBox

	properTableChild       bool // default weight true
	internalTableOrCaption bool // default weight true
	//properParents = (TableBox, InlineTableBox)
}

func NewTableCaptionBox(elementTag TBD, style css.StyleDict, children []AllBox) *TableCaptionBox {
	var out TableCaptionBox
	out.BlockBox = *NewBlockBox(elementTag, style, children)

	out.properTableChild = true
	out.internalTableOrCaption = true

	return &out
}

// PageBox is a box for a page
// Initially the whole document will be in the box for the root element.
//	During layout a new page box is created after every page break.
type PageBox struct {
	ParentBox

	pageType TBD
}

func (self *PageBox) init(pageType TBD, style css.StyleDict) {
	self.pageType = pageType
	// Page boxes are not linked to any element.
	self.ParentBox.init(TBD{}, style, nil)
}

func (self PageBox) String() string {
	return fmt.Sprintf("<PageBox %s>", self.pageType)
}

// MarginBox is a box in page margins, as defined in CSS3 Paged Media
type MarginBox struct {
	BlockContainerBox

	atKeyword TBD
}

func (self *MarginBox) init(atKeyword TBD, style css.StyleDict) {
	self.atKeyword = atKeyword
	//  Margin boxes are not linked to any element.
	self.BlockContainerBox.init(TBD{}, style, nil)
}

func (self MarginBox) String() string {
	return fmt.Sprintf("<MarginBox %s>", self.atKeyword)
}
