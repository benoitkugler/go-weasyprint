// The formatting structure is a tree of boxes. It is either "before layout",
// close to the element tree is it built from, or "after layout", with
// line breaks and page breaks.
package boxes

import (
	"fmt"

	"github.com/benoitkugler/go-weasyprint/style/tree"

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

// http://stackoverflow.com/questions/16317534/
var asciiToWide = map[rune]rune{}

func init() {
	for i := 33; i < 127; i++ {
		asciiToWide[rune(i)] = rune(i + 0xfee0)
	}
	asciiToWide[0x20] = '\u3000'
	asciiToWide[0x2D] = '\u2212'
}

type Point [2]float32

// Box is the common interface grouping all possible boxes
type Box interface {
	tree.Box

	Box() *BoxFields
	Copy() Box
	String() string
	IsProperChild(Box) bool
	allChildren() []Box
	Translate(box Box, dx float32, dy float32, ignoreFloats bool)
	removeDecoration(box *BoxFields, isStart, isEnd bool)
	pageValues() (pr.Page, pr.Page)
}

// BoxFields is an abstract base class for all boxes.
type BoxFields struct {
	// Keep track of removed collapsing spaces for wrap opportunities.
	leadingCollapsibleSpace  bool
	trailingCollapsibleSpace bool

	// Default, may be overriden on instances.
	IsTableWrapper   bool
	IsFlexItem       bool
	isForRootElement bool
	// isColumn         bool

	properTableChild       bool
	internalTableOrCaption bool
	tabularContainer       bool

	isAttachment bool
	// isListMarker         bool
	// transformationMatrix interface{}

	bookmarkLabel string

	stringSet pr.ContentProperties

	elementTag string
	Style      pr.Properties

	firstLetterStyle, firstLineStyle pr.Properties

	PositionX, PositionY                                                 float32
	Width, Height, MinWidth, MaxWidth, MinHeight, MaxHeight              pr.MaybeFloat
	Top, Bottom, Left, Right                                             float32
	MarginTop, MarginBottom, MarginLeft, MarginRight                     float32
	PaddingTop, PaddingBottom, PaddingLeft, PaddingRight                 float32
	BorderTopWidth, BorderRightWidth, BorderBottomWidth, BorderLeftWidth pr.MaybeFloat

	BorderTopLeftRadius, BorderTopRightRadius, BorderBottomRightRadius, BorderBottomLeftRadius Point

	viewportOverflow string

	Children []Box
	// outsideListMarker Box

	missingLink         tree.Box
	cachedCounterValues tree.CounterValues

	isHeader bool
	isFooter bool

	span    int
	Colspan int
	Rowspan int

	ColumnGroups        []Box
	columnPositions     []float32
	GridX               int
	collapsedBorderGrid BorderGrids
}

func newBoxFields(elementTag string, style pr.Properties, children []Box) BoxFields {
	return BoxFields{elementTag: elementTag, Style: style, Children: children}
}

// BoxType enables passing type as value
type BoxType interface {
	AnonymousFrom(parent Box, children []Box) Box

	// Returns true if box is of type (or subtype) BoxType
	IsInstance(box Box) bool
}

func (box *BoxFields) allChildren() []Box {
	return box.Children
}

func (box *BoxFields) IsProperChild(b Box) bool {
	return false
}

// ----------------------- needed by target ----------------------

func (box *BoxFields) CachedCounterValues() tree.CounterValues {
	return box.cachedCounterValues
}

func (box *BoxFields) SetCachedCounterValues(cv tree.CounterValues) {
	box.cachedCounterValues = cv
}

func (box *BoxFields) MissingLink() tree.Box {
	return box.missingLink
}

func (box *BoxFields) SetMissingLink(b tree.Box) {
	box.missingLink = b
}

func copyWithChildren(box Box, newChildren []Box, isStart bool, isEnd bool) Box {
	newBox := box.Copy()
	newBox.Box().Children = newChildren
	if box.Box().Style.GetBoxDecorationBreak() == "slice" {
		newBox.removeDecoration(newBox.Box(), !isStart, !isEnd)
	}
	return newBox
}

func deepcopy(b Box) Box {
	new := b.Copy()
	newChildren := make([]Box, len(b.Box().Children))
	for i, c := range b.Box().Children {
		newChildren[i] = deepcopy(c)
	}
	new.Box().Children = newChildren
	return new
}

func descendants(b Box) []Box {
	out := []Box{b}
	for _, child := range b.Box().Children {
		out = append(out, descendants(child)...)
	}
	return out
}

func (b BoxFields) GetWrappedTable() Box {
	if b.IsTableWrapper {
		for _, child := range b.Children {
			if _, ok := child.(InstanceTableBox); ok {
				return child
			}
		}
	}
	return nil
}

// Translate changes the box’s position.
// Also update the children’s positions accordingly.
func (BoxFields) Translate(box Box, dx, dy float32, ignoreFloats bool) {
	if dx == 0 && dy == 0 {
		return
	}
	box.Box().PositionX += dx
	box.Box().PositionY += dy
	for _, child := range box.allChildren() {
		if !(ignoreFloats && child.Box().isFloated()) {
			child.Translate(child, dx, dy, ignoreFloats)
		}
	}
}

// ---- Heights and widths -----

// Width of the padding box.
func (self BoxFields) PaddingWidth() float32 {
	return self.Width.V() + self.PaddingLeft + self.PaddingRight
}

// Height of the padding box.
func (self BoxFields) PaddingHeight() float32 {
	return self.Height.V() + self.PaddingTop + self.PaddingBottom
}

// Width of the border box.
func (self BoxFields) BorderWidth() float32 {
	return self.PaddingWidth() + self.BorderLeftWidth + self.BorderRightWidth
}

// Height of the border box.
func (self BoxFields) BorderHeight() float32 {
	return self.PaddingHeight() + self.BorderTopWidth + self.BorderBottomWidth
}

// Width of the margin box (aka. outer box).
func (self BoxFields) MarginWidth() float32 {
	return self.BorderWidth() + self.MarginLeft + self.MarginRight
}

// Height of the margin box (aka. outer box).
func (self BoxFields) MarginHeight() float32 {
	return self.BorderHeight() + self.MarginTop + self.MarginBottom
}

// Corners positions

// Absolute horizontal position of the content box.
func (self BoxFields) ContentBoxX() float32 {
	return self.PositionX + self.MarginLeft + self.PaddingLeft + self.BorderLeftWidth
}

// Absolute vertical position of the content box.
func (self BoxFields) ContentBoxY() float32 {
	return self.PositionY + self.MarginTop + self.PaddingTop + self.BorderTopWidth
}

// Absolute horizontal position of the padding box.
func (self BoxFields) PaddingBoxX() float32 {
	return self.PositionX + self.MarginLeft + self.BorderLeftWidth
}

// Absolute vertical position of the padding box.
func (self BoxFields) PaddingBoxY() float32 {
	return self.PositionY + self.MarginTop + self.BorderTopWidth
}

// Absolute horizontal position of the border box.
func (self BoxFields) borderBoxX() float32 {
	return self.PositionX + self.MarginLeft
}

// Absolute vertical position of the border box.
func (self BoxFields) borderBoxY() float32 {
	return self.PositionY + self.MarginTop
}

// Return the rectangle where the box is clickable."""
// "Border area. That's the area that hit-testing is done on."
// http://lists.w3.org/Archives/Public/www-style/2012Jun/0318.html
// TODO: manage the border radii, use outerBorderRadii instead
func (self BoxFields) hitArea() (x float32, y float32, w float32, h float32) {
	return self.borderBoxX(), self.borderBoxY(), self.BorderWidth(), self.BorderHeight()
}

type roundedBox struct {
	x, y, width, height                        float32
	topLeft, topRight, bottomRight, bottomLeft Point
}

// Position, size and radii of a box inside the outer border box.
//bt, br, bb, and bl are distances from the outer border box,
//defining a rectangle to be rounded.
func (self BoxFields) roundedBox(bt, br, bb, bl float32) roundedBox {
	tlr := self.BorderTopLeftRadius
	trr := self.BorderTopRightRadius
	brr := self.BorderBottomRightRadius
	blr := self.BorderBottomLeftRadius

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
	width := self.BorderWidth() - bl - br
	height := self.BorderHeight() - bt - bb

	// Fix overlapping curves
	//See http://www.w3.org/TR/css3-background/#corner-overlap
	Points := []Point{
		{width, tlrx + trrx},
		{width, blrx + brrx},
		{height, tlry + blry},
		{height, trry + brry},
	}
	var ratio float32 = 1.
	for _, Point := range Points {
		if Point[1] > 0 {
			candidat := Point[0] / Point[1]
			if candidat < ratio {
				ratio = candidat
			}
		}
	}
	return roundedBox{x: x, y: y, width: width, height: height,
		topLeft:     Point{tlrx * ratio, tlry * ratio},
		topRight:    Point{trrx * ratio, trry * ratio},
		bottomRight: Point{brrx * ratio, brry * ratio},
		bottomLeft:  Point{blrx * ratio, blry * ratio},
	}
}

func (self BoxFields) roundedBoxRatio(ratio float32) roundedBox {
	return self.roundedBox(
		self.BorderTopWidth*ratio,
		self.BorderRightWidth*ratio,
		self.BorderBottomWidth*ratio,
		self.BorderLeftWidth*ratio)
}

// Return the position, size and radii of the rounded padding box.
func (self BoxFields) roundedPaddingBox() roundedBox {
	return self.roundedBox(
		self.BorderTopWidth,
		self.BorderRightWidth,
		self.BorderBottomWidth,
		self.BorderLeftWidth)
}

// Return the position, size and radii of the rounded border box.
func (self BoxFields) roundedBorderBox() roundedBox {
	return self.roundedBox(0, 0, 0, 0)
}

// Return the position, size and radii of the rounded content box.
func (self BoxFields) roundedContentBox() roundedBox {
	return self.roundedBox(
		self.BorderTopWidth+self.PaddingTop,
		self.BorderRightWidth+self.PaddingRight,
		self.BorderBottomWidth+self.PaddingBottom,
		self.BorderLeftWidth+self.PaddingLeft)
}

// Positioning schemes

// Return whether this box is floated.
func (self BoxFields) isFloated() bool {
	return self.Style.GetFloat() != "none"
}

// Return whether this box is in the absolute positioning scheme.
func (self BoxFields) IsAbsolutelyPositioned() bool {
	pos := self.Style.GetPosition()
	return !pos.Bool && pos.String == "absolute" || pos.String == "fixed"
}

// Return whether this box is a running element.
func (self BoxFields) isRunning() bool {
	pos := self.Style.GetPosition()
	return pos.Bool && pos.String == "running()"
}

// Return whether this box is in normal flow.
func (self BoxFields) isInNormalFlow() bool {
	return !(self.isFloated() || self.IsAbsolutelyPositioned() || self.isRunning())
}

// Start and end page values for named pages

// Return start and end page values.
func (b BoxFields) pageValues() (pr.Page, pr.Page) {
	start := b.Style.GetPage()
	end := start
	children := b.Children
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
	self.Style[fmt.Sprintf("margin_%s", side)] = pr.ZeroPixels.ToValue()
	self.Style[fmt.Sprintf("padding_%s", side)] = pr.ZeroPixels.ToValue()
	self.Style[fmt.Sprintf("border_%s_width", side)] = pr.FToV(0)

	if side == "top" || side == "bottom" {
		self.Style[fmt.Sprintf("border_%s_left_radius", side)] = pr.Point{pr.ZeroPixels, pr.ZeroPixels}
		self.Style[fmt.Sprintf("border_%s_right_radius", side)] = pr.Point{pr.ZeroPixels, pr.ZeroPixels}
	} else {
		self.Style[fmt.Sprintf("border_bottom_%s_radius", side)] = pr.Point{pr.ZeroPixels, pr.ZeroPixels}
		self.Style[fmt.Sprintf("border_top_%s_radius", side)] = pr.Point{pr.ZeroPixels, pr.ZeroPixels}
	}

	switch side {
	case "top":
		self.MarginTop = 0
		self.PaddingTop = 0
		self.BorderTopWidth = 0
	case "right":
		self.MarginRight = 0
		self.PaddingRight = 0
		self.BorderRightWidth = 0
	case "left":
		self.MarginLeft = 0
		self.PaddingLeft = 0
		self.BorderLeftWidth = 0
	case "bottom":
		self.MarginBottom = 0
		self.PaddingBottom = 0
		self.BorderBottomWidth = 0
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
