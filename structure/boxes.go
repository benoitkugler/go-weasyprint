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

var (
	TypeTableRowBox         BoxType = typeTableRowBox{}
	TypeTableRowGroupBox    BoxType = typeTableRowGroupBox{}
	TypeTableColumnBox      BoxType = typeTableColumnBox{}
	TypeTableColumnGroupBox BoxType = typeTableColumnGroupBox{}
	TypeTableBox            BoxType = typeTableBox{}
	TypeTableCellBox        BoxType = typeTableCellBox{}
	TypeInlineTableBox      BoxType = typeInlineTableBox{}
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

type point struct {
	x, y float64
}

type TBD struct{}

func (self *Box) init(elementTag string, style css.StyleDict) {
	self.elementTag = elementTag
	self.style = style
}

func (self Box) String() string {
	return fmt.Sprintf("<Box %s>", self.elementTag)
}

func (self *Box) BaseBox() *Box {
	return self
}

func (self *Box) TableFields() *TableFields {
	return nil
}

func (self Box) IsParentBox() bool {
	return false
}

func (self Box) IsTableBox() bool {
	return false
}

func (self Box) IsBlockContainerBox() bool {
	return false
}

func (self Box) IsInlineLevelBox() bool {
	return false
}

func (self Box) IsProperChild(parent AllBox) bool {
	return false
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
		if !(ignoreFloats && child.BaseBox().isFloated()) {
			child.BaseBox().Translate(dx, dy, ignoreFloats)
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
	return self.style.Strings["float"] != "none"
}

// Return whether this box is in the absolute positioning scheme.
func (self Box) isAbsolutelyPositioned() bool {
	pos := self.style.Strings["position"]
	return pos == "absolute" || pos == "fixed"
}

// Return whether this box is in normal flow.
func (self Box) isInNormalFlow() bool {
	return !(self.isFloated() || self.isAbsolutelyPositioned())
}

// Start and end page values for named pages

// Return start and end page values.
func (self Box) pageValues() (int, int) {
	p := self.style.Page.Page
	return p, p
}

// Set to 0 the margin, padding and border of ``side``.
func (self *Box) resetSpacing(side css.Side) {
	self.style.Values[fmt.Sprintf("margin_%s", side)] = css.ZeroPixels
	self.style.Values[fmt.Sprintf("padding_%s", side)] = css.ZeroPixels
	self.style.Values[fmt.Sprintf("border_%s_width", side)] = css.ZeroPixels

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
	ltr := self.style.Strings["direction"] == "ltr"
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

func (self Box) copyWithChildren(newChildren []AllBox, isStart, isEnd bool) ParentBox {
	return ParentBox{}
}

func (p *ParentBox) init(elementTag string, style css.StyleDict, children []AllBox) {
	p.Box.init(elementTag, style)
	p.children = children
}

func (self ParentBox) IsParentBox() bool {
	return true
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
func CopyWithChildren(box AllBox, newChildren []AllBox, isStart, isEnd bool) AllBox {
	newBox := box.Copy()
	newBox.BaseBox().children = newChildren
	if !isStart {
		newBox.BaseBox().outsideListMarker = nil
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
		childStart, _ := startBox.BaseBox().pageValues()
		_, childEnd := endBox.BaseBox().pageValues()
		if childStart > 0 {
			start = childStart
		}
		if childEnd > 0 {
			end = childEnd
		}
	}
	return start, end
}

func NewBlockBox(elementTag string, style css.StyleDict, children []AllBox) *BlockBox {
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

func BlockBoxAnonymousFrom(parent AllBox, children []AllBox) AllBox {
	return NewBlockBox(parent.BaseBox().elementTag, parent.BaseBox().style.InheritFrom(), children)
}

func (self BlockContainerBox) IsBlockContainerBox() bool {
	return true
}

// func (self BlockBox) pageValues() (int, int) {
// 	return self.BlockContainerBox.pageValues()
// }

func NewLineBox(elementTag string, style css.StyleDict, children []AllBox) *LineBox {
	if !style.Anonymous {
		log.Fatal("style must be anonymous")
	}
	var l LineBox
	l.ParentBox.init(elementTag, style, children)
	return &l
}

func (self InlineLevelBox) IsInlineLevelBox() bool {
	return true
}

func LineBoxAnonymousFrom(parent AllBox, children []AllBox) AllBox {
	return NewLineBox(parent.BaseBox().elementTag, parent.BaseBox().style.InheritFrom(), children)
}

func NewInlineBox(elementTag string, style css.StyleDict, children []AllBox) *InlineBox {
	var out InlineBox
	out.init(elementTag, style, children)
	return &out
}

func InlineBoxIsInstance(box AllBox) bool {
	_, is := box.(*InlineBox)
	return is
}

// Return the (x, y, w, h) rectangle where the box is clickable.
func (self InlineBox) hitArea() (x float64, y float64, w float64, h float64) {
	return self.borderBoxX(), self.positionY, self.borderWidth(), self.marginHeight()
}

func NewTextBox(elementTag string, style css.StyleDict, text string) *TextBox {
	var self TextBox
	if !style.Anonymous {
		panic("style is not anonymous")
	}
	if len(text) == 0 {
		panic("empty text")
	}
	self.Box.init(elementTag, style)
	textTransform := style.Strings["text-transform"]
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

		if style.Strings["hyphens"] == "none" {
			text = strings.ReplaceAll(text, "\u00AD", "") //  U+00AD SOFT HYPHEN (SHY)
		}
	}
	self.text = text
	return &self
}

func (self TextBox) IsInlineLevelBox() bool {
	return true
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

func TextBoxAnonymousFrom(parent AllBox, text string) AllBox {
	return NewTextBox(parent.BaseBox().elementTag, parent.BaseBox().style.InheritFrom(), text)
}

func TextBoxIsInstance(box AllBox) bool {
	_, is := box.(*TextBox)
	return is
}

func NewInlineBlockBox(elementTag string, style css.StyleDict, children []AllBox) *InlineBlockBox {
	var out InlineBlockBox
	out.init(elementTag, style, children)
	return &out
}

func InlineBlockBoxAnonymousFrom(parent AllBox, children []AllBox) AllBox {
	return NewInlineBlockBox(parent.BaseBox().elementTag, parent.BaseBox().style.InheritFrom(), children)
}

func NewReplacedBox(elementTag string, style css.StyleDict, replacement css.ImageType) *ReplacedBox {
	var self ReplacedBox
	self.Box.init(elementTag, style)
	self.replacement = replacement
	return &self
}

func NewBlockReplacedBox(elementTag string, style css.StyleDict, replacement css.ImageType) *BlockReplacedBox {
	var self BlockReplacedBox
	self.ReplacedBox = *NewReplacedBox(elementTag, style, replacement)
	return &self
}

func NewInlineReplacedBox(elementTag string, style css.StyleDict, replacement css.ImageType) *InlineReplacedBox {
	var self InlineReplacedBox
	self.ReplacedBox = *NewReplacedBox(elementTag, style, replacement)
	return &self
}

func (self InlineReplacedBox) IsInlineLevelBox() bool {
	return true
}

func InlineReplacedBoxAnonymousFrom(parent AllBox, replacement css.ImageType) AllBox {
	return NewInlineReplacedBox(parent.BaseBox().elementTag, parent.BaseBox().style.InheritFrom(), replacement)
}

func newTableFields() TableFields {
	return TableFields{
		tabularContainer:       true,
		properTableChild:       true,
		internalTableOrCaption: true,
		span:                   1,
		colspan:                1,
		rowspan:                1,
	}
}

func (t *TableBox) TableFields() *TableFields {
	return &t.tableFields
}

func NewTableBox(elementTag string, style css.StyleDict, children []AllBox) *TableBox {
	var out TableBox
	out.init(elementTag, style, children)
	out.tableFields = newTableFields()
	return &out
}

// Definitions for the rules generating anonymous table boxes
// http://www.w3.org/TR/CSS21/tables.html#anonymous-boxes

func (self TableBox) AllChildren() []AllBox {
	return append(self.children, self.tableFields.columnGroups...)
}

func (self *TableBox) Translate(dx, dy float64, ignoreFloats bool) {
	if dx == 0 && dy == 0 {
		return
	}
	for index, position := range self.tableFields.columnPositions {
		self.tableFields.columnPositions[index] = position + dx
	}
	self.ParentBox.Translate(dx, dy, ignoreFloats)
}

func (self TableBox) pageValues() (int, int) {
	return self.ParentBox.pageValues()
}

func (self TableBox) IsTableBox() bool {
	return true
}

func NewInlineTableBox(elementTag string, style css.StyleDict, children []AllBox) *InlineTableBox {
	return &InlineTableBox{*NewTableBox(elementTag, style, children)}
}

func NewTableRowGroupBox(elementTag string, style css.StyleDict, children []AllBox) *TableRowGroupBox {
	var out TableRowGroupBox
	out.init(elementTag, style, children)
	out.tableFields = newTableFields()
	return &out
}

func (t *TableRowGroupBox) TableFields() *TableFields {
	return &t.tableFields
}

func (self TableRowGroupBox) IsProperChild(parent AllBox) bool {
	switch parent.(type) {
	case *TableBox, *InlineTableBox:
		return true
	default:
		return false
	}
}

func NewTableRowBox(elementTag string, style css.StyleDict, children []AllBox) *TableRowBox {
	var out TableRowBox
	out.init(elementTag, style, children)
	out.tableFields = newTableFields()
	return &out
}

func (t *TableRowBox) TableFields() *TableFields {
	return &t.tableFields
}

func (self TableRowBox) IsProperChild(parent AllBox) bool {
	switch parent.(type) {
	case *TableBox, *InlineTableBox, *TableRowGroupBox:
		return true
	default:
		return false
	}
}

func NewTableColumnGroupBox(elementTag string, style css.StyleDict, children []AllBox) *TableColumnGroupBox {
	var out TableColumnGroupBox
	out.init(elementTag, style, children)
	out.tableFields = newTableFields()
	return &out
}

func (t *TableColumnGroupBox) TableFields() *TableFields {
	return &t.tableFields
}

func (self TableColumnGroupBox) IsProperChild(parent AllBox) bool {
	switch parent.(type) {
	case *TableBox, *InlineTableBox:
		return true
	default:
		return false
	}
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

func NewTableColumnBox(elementTag string, style css.StyleDict, children []AllBox) *TableColumnBox {
	var out TableColumnBox
	out.init(elementTag, style, children)
	out.tableFields = newTableFields()

	return &out
}

func (t *TableColumnBox) TableFields() *TableFields {
	return &t.tableFields
}

// Return cells that originate in the column.
// May be overriden on instances.
func (self TableColumnBox) getCells() []AllBox {
	return nil
}

func (self TableColumnBox) IsProperChild(parent AllBox) bool {
	switch parent.(type) {
	case *TableBox, *InlineTableBox, *TableColumnGroupBox:
		return true
	default:
		return false
	}
}

func NewTableCellBox(elementTag string, style css.StyleDict, children []AllBox) *TableCellBox {
	var out TableCellBox
	out.init(elementTag, style, children)
	out.tableFields = newTableFields()
	return &out
}

func (t *TableCellBox) TableFields() *TableFields {
	return &t.tableFields
}

func NewTableCaptionBox(elementTag string, style css.StyleDict, children []AllBox) *TableCaptionBox {
	var out TableCaptionBox
	out.BlockBox = *NewBlockBox(elementTag, style, children)

	out.tableFields = newTableFields()

	return &out
}

func (t *TableCaptionBox) TableFields() *TableFields {
	return &t.tableFields
}

func (self TableCaptionBox) IsProperChild(parent AllBox) bool {
	switch parent.(type) {
	case *TableBox, *InlineTableBox:
		return true
	default:
		return false
	}
}

func TableCaptionBoxIsInstance(box AllBox) bool {
	_, is := box.(*TableCaptionBox)
	return is
}

func (self *PageBox) init(pageType TBD, style css.StyleDict) {
	self.pageType = pageType
	// Page boxes are not linked to any element.
	self.ParentBox.init("", style, nil)
}

func (self PageBox) String() string {
	return fmt.Sprintf("<PageBox %s>", self.pageType)
}

func (self *MarginBox) init(atKeyword TBD, style css.StyleDict) {
	self.atKeyword = atKeyword
	//  Margin boxes are not linked to any element.
	self.BlockContainerBox.init("", style, nil)
}

func (self MarginBox) String() string {
	return fmt.Sprintf("<MarginBox %s>", self.atKeyword)
}

// -----------------------------------------------------------------
// ----------------- AnonymousFrom constructor ---------------------
// Return an anonymous box that inherits from ``parent``.
// -----------------------------------------------------------------

// Since we dont use reflection, we implements (python) Box types as interfaces
type BoxType interface {
	AnonymousFrom(parent AllBox, children []AllBox) AllBox

	// Returns true if box is of type (or subtype) BoxType
	IsInstance(box AllBox) bool
}

type typeTableRowBox struct{}
type typeTableRowGroupBox struct{}
type typeTableColumnBox struct{}
type typeTableColumnGroupBox struct{}
type typeTableBox struct{}
type typeTableCellBox struct{}
type typeInlineTableBox struct{}

func (t typeTableRowBox) AnonymousFrom(parent AllBox, children []AllBox) AllBox {
	return NewTableRowBox(parent.BaseBox().elementTag, parent.BaseBox().style.InheritFrom(), children)
}

func (t typeTableRowGroupBox) AnonymousFrom(parent AllBox, children []AllBox) AllBox {
	return NewTableRowGroupBox(parent.BaseBox().elementTag, parent.BaseBox().style.InheritFrom(), children)
}

func (t typeTableColumnBox) AnonymousFrom(parent AllBox, children []AllBox) AllBox {
	return NewTableColumnBox(parent.BaseBox().elementTag, parent.BaseBox().style.InheritFrom(), children)
}

func (t typeTableColumnGroupBox) AnonymousFrom(parent AllBox, children []AllBox) AllBox {
	return NewTableColumnGroupBox(parent.BaseBox().elementTag, parent.BaseBox().style.InheritFrom(), children)
}

func (t typeTableBox) AnonymousFrom(parent AllBox, children []AllBox) AllBox {
	return NewTableBox(parent.BaseBox().elementTag, parent.BaseBox().style.InheritFrom(), children)
}

func (t typeTableCellBox) AnonymousFrom(parent AllBox, children []AllBox) AllBox {
	return NewTableCellBox(parent.BaseBox().elementTag, parent.BaseBox().style.InheritFrom(), children)
}

func (t typeInlineTableBox) AnonymousFrom(parent AllBox, children []AllBox) AllBox {
	return NewInlineTableBox(parent.BaseBox().elementTag, parent.BaseBox().style.InheritFrom(), children)
}

func (t typeTableRowBox) IsInstance(child AllBox) bool {
	_, is := child.(*TableRowBox)
	return is
}

func (t typeTableRowGroupBox) IsInstance(child AllBox) bool {
	_, is := child.(*TableRowGroupBox)
	return is
}

func (t typeTableColumnBox) IsInstance(child AllBox) bool {
	_, is := child.(*TableColumnBox)
	return is
}

func (t typeTableColumnGroupBox) IsInstance(child AllBox) bool {
	_, is := child.(*TableColumnGroupBox)
	return is
}

func (t typeTableBox) IsInstance(child AllBox) bool {
	return child.IsTableBox()
}

func (t typeTableCellBox) IsInstance(child AllBox) bool {
	_, is := child.(*TableCellBox)
	return is
}

func (t typeInlineTableBox) IsInstance(child AllBox) bool {
	_, is := child.(*InlineTableBox)
	return is
}
