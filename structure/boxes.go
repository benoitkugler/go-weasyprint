package structure

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

type ConcreteBox interface {
	AllChildren() []*Box
	IsTableBox() bool
}

// Box is an abstract base class for all boxes.
type Box struct {
	ConcreteBox

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
}

func (self *Box) init(elementTag TBD, style css.StyleDict) {
	self.elementTag = elementTag
	self.style = style
}

func (self Box) String() string {
	return fmt.Sprintf("<Box %s>", self.elementTag)
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
	return self.style.Float != "none"
}

// Return whether this box is in the absolute positioning scheme.
func (self Box) isAbsolutelyPositioned() bool {
	return self.style.Position == "absolute" || self.style.Position == "fixed"
}

// Return whether this box is in normal flow.
func (self Box) isInNormalFlow() bool {
	return !(self.isFloated() || self.isAbsolutelyPositioned())
}

// Start and end page values for named pages

// Return start and end page values.
func (self Box) pageValues() (int, int) {
	return self.style.Page, self.style.Page
}

// Set to 0 the margin, padding and border of ``side``.
func (self *Box) resetSpacing(side css.Side) {
	self.style.Margin[side] = css.Dimension{Unit: "px"}
	self.style.Padding[side] = css.Dimension{Unit: "px"}
	self.style.BorderWidth[side] = 0
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

// ParentBox is a box that has children.
type ParentBox struct {
	Box

	children          []*Box
	outsideListMarker *Box
}

func (p *ParentBox) init(elementTag TBD, style css.StyleDict, children []*Box) {
	p.Box.init(elementTag, style)
	p.children = children
	p.ConcreteBox = p
}

func (self ParentBox) AllChildren() []*Box {
	return self.children
}

func (self ParentBox) IsTableBox() bool {
	return false
}

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
func (self ParentBox) copyWithChildren(newChildren []*Box, isStart, isEnd bool) ParentBox {
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
//	if hasattr(child, 'descendants'):
//	for grandChild in child.descendants():
//	yield grandChild
//	else:
//	yield child
//}

// Get the table wrapped by the box.
// Warning, might be nil
func (self ParentBox) getWrappedTable() (*Box, error) {
	if self.isTableWrapper {
		for _, child := range self.children {
			if child.IsTableBox() {
				return child, nil
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
//An element with a ``display`` value of ``block``, ``list-item`` or
//``table`` generates a block-level box.
type BlockLevelBox struct {
	Box

	clearance TBD
}

// BlockContainerBox is a box that contains only block-level boxes or only line boxes.
//
//A box that either contains only block-level boxes or establishes an inline
//formatting context and thus contains only line boxes.
//
//A non-replaced element with a ``display`` value of ``block``,
//``list-item``, ``inline-block`` or 'table-cell' generates a block container
//box.
type BlockContainerBox struct {
	ParentBox
}

// BlockBox is a block-level box that is also a block container.
//
//A non-replaced element with a ``display`` value of ``block``, ``list-item``
//generates a block box.
type BlockBox struct {
	BlockContainerBox
	BlockLevelBox
}

func (self BlockBox) AllChildren() []*Box {
	if self.outsideListMarker != nil {
		return append(self.children, self.outsideListMarker)
	}
	return self.children
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

func (l *LineBox) init(elementTag TBD, style css.StyleDict, children []*Box) {
	if !style.Anonymous {
		log.Fatal("style must be anonymous")
	}
	l.ParentBox.init(elementTag, style, children)
}

// InlineLevelBox is a box that participates in an inline formatting context.
//
//An inline-level box that is not an inline box is said to be "atomic". Such
//boxes are inline blocks, replaced elements and inline tables.
//
//An element with a ``display`` value of ``inline``, ``inline-table``, or
//``inline-block`` generates an inline-level box.
type InlineLevelBox struct {
	Box
}

func (self *InlineLevelBox) removeDecoration(start, end bool) {
	if start || end {
		self.style = self.style.Copy()
	}
	ltr := self.style.Direction == "ltr"
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

// InlineBox is an inline box with inline children.
//
//A box that participates in an inline formatting context and whose content
//also participates in that inline formatting context.
//
//A non-replaced element with a ``display`` value of ``inline`` generates an
//inline box.
type InlineBox struct {
	InlineLevelBox
	ParentBox
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
	InlineLevelBox

	justificationSpacing int
	text                 string
}

func (self *TextBox) init(elementTag TBD, style css.StyleDict, text string) {
	if !style.Anonymous {
		log.Fatal("style is not anonymous")
	}
	if len(text) == 0 {
		log.Fatal("empty text")
	}
	self.InlineLevelBox.init(elementTag, style)
	textTransform := style.TextTransform
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

		if style.Hyphens == "none" {
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
