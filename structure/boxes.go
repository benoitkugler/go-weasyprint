//   Classes for all types of boxes in the CSS formatting structure / box model.
//	 Only define base class and interface, completed in package structure.
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
package boxes

import (
	"errors"
	"fmt"

	"github.com/benoitkugler/go-weasyprint/structure"
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/utils"
)

// http://stackoverflow.com/questions/16317534/
var asciiToWide = map[rune]rune{}

func init() {
	for i := 33; i < 127; i++ {
		asciiToWide[rune(i)] = rune(i + 0xfee0)
	}
	asciiToWide[0x20] = '\u3000'
	asciiToWide[0x2D] = '\u2212'
}

type point [2]float32

// Box is the common interface grouping all possible boxes
type Box interface {
	Box() *BoxFields
	Copy() Box
	String() string
	allChildren() []Box
	translate(box Box, dx float32, dy float32, ignoreFloats bool)
	removeDecoration(box *BoxFields, isStart, isEnd bool)
	pageValues() (pr.Page, pr.Page)
}

// BoxFields is an abstract base class for all boxes.
type BoxFields struct {
	// Keep track of removed collapsing spaces for wrap opportunities.
	leadingCollapsibleSpace  bool
	trailingCollapsibleSpace bool

	// Default, may be overriden on instances.
	isTableWrapper   bool
	isFlexItem       bool
	isForRootElement bool
	isColumn         bool

	isAttachment bool
	// isListMarker         bool
	transformationMatrix interface{}

	bookmarkLabel string

	stringSet pr.SContents

	elementTag string
	style      pr.Properties

	firstLetterStyle, firstLineStyle pr.Properties

	positionX, positionY float32

	width, height float32

	marginTop, marginBottom, marginLeft, marginRight float32

	paddingTop, paddingBottom, paddingLeft, paddingRight float32

	borderTopWidth, borderRightWidth, borderBottomWidth, borderLeftWidth float32

	borderTopLeftRadius, borderTopRightRadius, borderBottomRightRadius, borderBottomLeftRadius point

	viewportOverflow string

	children []Box
	// outsideListMarker Box

	missingLink         Box
	cachedCounterValues CounterValues
}

type TableFields struct {
	properTableChild       bool
	internalTableOrCaption bool
	tabularContainer       bool
	isHeader               bool
	isFooter               bool

	span    int
	colspan int
	rowspan int

	columnGroups    []Box
	columnPositions []float32
}

func NewBoxFields(elementTag string, style pr.Properties, children []Box) BoxFields {
	return BoxFields{elementTag: elementTag, style: style, children: children}
}

func CopyWithChildren(box Box, newChildren []Box, isStart bool, isEnd bool) Box {
	newBox := box.Copy()
	newBox.Box().children = newChildren
	if box.Box().style.GetBoxDecorationBreak() == "slice" {
		newBox.removeDecoration(newBox.Box(), !isStart, !isEnd)
	}
	return newBox
}

func Deepcopy(b Box) Box {
	new := b.Copy()
	newChildren := make([]Box, len(b.Box().children))
	for i, c := range b.Box().children {
		newChildren[i] = Deepcopy(c)
	}
	new.Box().children = newChildren
	return new
}

func Descendants(b Box) []Box {
	out := []Box{b}
	for _, child := range b.Box().children {
		out = append(out, Descendants(child)...)
	}
	return out
}

// BoxType enables passing type as value
type BoxType interface {
	AnonymousFrom(parent Box, children []Box) Box

	// Returns true if box is of type (or subtype) BoxType
	IsInstance(box Box) bool
}

func (box *BoxFields) allChildren() []Box {
	return box.children
}

func (b BoxFields) getWrappedTable() (structure.InstanceTableBox, error) {
	if b.isTableWrapper {
		for _, child := range b.children {
			if asTable, ok := child.(structure.InstanceTableBox); ok {
				return asTable, nil
			}
		}
		return nil, errors.New("Table wrapper without a table")
	}
	return nil, nil
}

// Translate changes the box’s position.
// Also update the children’s positions accordingly.
func (BoxFields) translate(box Box, dx, dy float32, ignoreFloats bool) {
	if dx == 0 && dy == 0 {
		return
	}
	box.Box().positionX += dx
	box.Box().positionY += dy
	for _, child := range box.allChildren() {
		if !(ignoreFloats && child.Box().isFloated()) {
			child.translate(child, dx, dy, ignoreFloats)
		}
	}
}

// ---- Heights and widths -----

// Width of the padding box.
func (self BoxFields) paddingWidth() float32 {
	return self.width + self.paddingLeft + self.paddingRight
}

// Height of the padding box.
func (self BoxFields) paddingHeight() float32 {
	return self.height + self.paddingTop + self.paddingBottom
}

// Width of the border box.
func (self BoxFields) borderWidth() float32 {
	return self.paddingWidth() + self.borderLeftWidth + self.borderRightWidth
}

// Height of the border box.
func (self BoxFields) borderHeight() float32 {
	return self.paddingHeight() + self.borderTopWidth + self.borderBottomWidth
}

// Width of the margin box (aka. outer box).
func (self BoxFields) marginWidth() float32 {
	return self.borderWidth() + self.marginLeft + self.marginRight
}

// Height of the margin box (aka. outer box).
func (self BoxFields) marginHeight() float32 {
	return self.borderHeight() + self.marginTop + self.marginBottom
}

// Corners positions

// Absolute horizontal position of the content box.
func (self BoxFields) contentBoxX() float32 {
	return self.positionX + self.marginLeft + self.paddingLeft + self.borderLeftWidth
}

// Absolute vertical position of the content box.
func (self BoxFields) contentBoxY() float32 {
	return self.positionY + self.marginTop + self.paddingTop + self.borderTopWidth
}

// Absolute horizontal position of the padding box.
func (self BoxFields) paddingBoxX() float32 {
	return self.positionX + self.marginLeft + self.borderLeftWidth
}

// Absolute vertical position of the padding box.
func (self BoxFields) paddingBoxY() float32 {
	return self.positionY + self.marginTop + self.borderTopWidth
}

// Absolute horizontal position of the border box.
func (self BoxFields) borderBoxX() float32 {
	return self.positionX + self.marginLeft
}

// Absolute vertical position of the border box.
func (self BoxFields) borderBoxY() float32 {
	return self.positionY + self.marginTop
}

// Return the rectangle where the box is clickable."""
// "Border area. That's the area that hit-testing is done on."
// http://lists.w3.org/Archives/Public/www-style/2012Jun/0318.html
// TODO: manage the border radii, use outerBorderRadii instead
func (self BoxFields) hitArea() (x float32, y float32, w float32, h float32) {
	return self.borderBoxX(), self.borderBoxY(), self.borderWidth(), self.borderHeight()
}

type roundedBox struct {
	x, y, width, height                        float32
	topLeft, topRight, bottomRight, bottomLeft point
}

// Position, size and radii of a box inside the outer border box.
//bt, br, bb, and bl are distances from the outer border box,
//defining a rectangle to be rounded.
func (self BoxFields) roundedBox(bt, br, bb, bl float32) roundedBox {
	tlr := self.borderTopLeftRadius
	trr := self.borderTopRightRadius
	brr := self.borderBottomRightRadius
	blr := self.borderBottomLeftRadius

	tlrx := utils.Max(0, tlr[0]-bl)
	tlry := utils.Max(0, tlr[1]-bt)
	trrx := utils.Max(0, trr[0]-br)
	trry := utils.Max(0, trr[1]-bt)
	brrx := utils.Max(0, brr[0]-br)
	brry := utils.Max(0, brr[1]-bb)
	blrx := utils.Max(0, blr[0]-bl)
	blry := utils.Max(0, blr[1]-bb)

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
	var ratio float32 = 1.
	for _, point := range points {
		if point[1] > 0 {
			candidat := point[0] / point[1]
			if candidat < ratio {
				ratio = candidat
			}
		}
	}
	return roundedBox{x: x, y: y, width: width, height: height,
		topLeft:     point{tlrx * ratio, tlry * ratio},
		topRight:    point{trrx * ratio, trry * ratio},
		bottomRight: point{brrx * ratio, brry * ratio},
		bottomLeft:  point{blrx * ratio, blry * ratio},
	}
}

func (self BoxFields) roundedBoxRatio(ratio float32) roundedBox {
	return self.roundedBox(
		self.borderTopWidth*ratio,
		self.borderRightWidth*ratio,
		self.borderBottomWidth*ratio,
		self.borderLeftWidth*ratio)
}

// Return the position, size and radii of the rounded padding box.
func (self BoxFields) roundedPaddingBox() roundedBox {
	return self.roundedBox(
		self.borderTopWidth,
		self.borderRightWidth,
		self.borderBottomWidth,
		self.borderLeftWidth)
}

// Return the position, size and radii of the rounded border box.
func (self BoxFields) roundedBorderBox() roundedBox {
	return self.roundedBox(0, 0, 0, 0)
}

// Return the position, size and radii of the rounded content box.
func (self BoxFields) roundedContentBox() roundedBox {
	return self.roundedBox(
		self.borderTopWidth+self.paddingTop,
		self.borderRightWidth+self.paddingRight,
		self.borderBottomWidth+self.paddingBottom,
		self.borderLeftWidth+self.paddingLeft)
}

// Positioning schemes

// Return whether this box is floated.
func (self BoxFields) isFloated() bool {
	return self.style.GetFloat() != "none"
}

// Return whether this box is in the absolute positioning scheme.
func (self BoxFields) isAbsolutelyPositioned() bool {
	pos := self.style.GetPosition()
	return !pos.Bool && pos.String == "absolute" || pos.String == "fixed"
}

// Return whether this box is a running element.
func (self BoxFields) isRunning() bool {
	pos := self.style.GetPosition()
	return pos.Bool && pos.String == "running()"
}

// Return whether this box is in normal flow.
func (self BoxFields) isInNormalFlow() bool {
	return !(self.isFloated() || self.isAbsolutelyPositioned() || self.isRunning())
}

// Start and end page values for named pages

// Return start and end page values.
func (b BoxFields) pageValues() (pr.Page, pr.Page) {
	start := b.style.GetPage()
	end := start
	children := b.children
	if len(children) > 0 {
		startBox, endBox := children[0], children[len(children)-1]
		childStart, _ := startBox.pageValues()
		_, childEnd := endBox.pageValues()
		if !childStart.IsNone() {
			start = childStart
		}
		if !childEnd.IsNone() {
			end = childEnd
		}
	}
	return start, end
}

// Set to 0 the margin, padding and border of ``side``.
func (self *BoxFields) resetSpacing(side string) {
	self.style[fmt.Sprintf("margin_%s", side)] = pr.ZeroPixels.ToValue()
	self.style[fmt.Sprintf("padding_%s", side)] = pr.ZeroPixels.ToValue()
	self.style[fmt.Sprintf("border_%s_width", side)] = pr.FToV(0)

	if side == "top" || side == "bottom" {
		self.style[fmt.Sprintf("border_%s_left_radius", side)] = pr.Point{pr.ZeroPixels, pr.ZeroPixels}
		self.style[fmt.Sprintf("border_%s_right_radius", side)] = pr.Point{pr.ZeroPixels, pr.ZeroPixels}
	} else {
		self.style[fmt.Sprintf("border_bottom_%s_radius", side)] = pr.Point{pr.ZeroPixels, pr.ZeroPixels}
		self.style[fmt.Sprintf("border_top_%s_radius", side)] = pr.Point{pr.ZeroPixels, pr.ZeroPixels}
	}

	switch side {
	case "top":
		self.marginTop = 0
		self.paddingTop = 0
		self.borderTopWidth = 0
	case "right":
		self.marginRight = 0
		self.paddingRight = 0
		self.borderRightWidth = 0
	case "left":
		self.marginLeft = 0
		self.paddingLeft = 0
		self.borderLeftWidth = 0
	case "bottom":
		self.marginBottom = 0
		self.paddingBottom = 0
		self.borderBottomWidth = 0
	}
}

func (BoxFields) removeDecoration(box *BoxFields, start, end bool) {
	if start {
		box.resetSpacing("top")
	}
	if end {
		box.resetSpacing("bottom")
	}
}
