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

type Point [2]pr.Float

type MaybePoint [2]pr.MaybeFloat

func (mp MaybePoint) V() Point {
	return Point{mp[0].V(), mp[1].V()}
}

// Box is the common interface grouping all possible boxes
type Box interface {
	tree.Box

	Box() *BoxFields
	Copy() Box
	String() string
	IsProperChild(Box) bool
	allChildren() []Box
	// ignoreFloats = false
	Translate(box Box, dx, dy pr.Float, ignoreFloats bool)
	RemoveDecoration(box *BoxFields, isStart, isEnd bool)
	PageValues() (pr.Page, pr.Page)
}

// BoxType enables passing type as value
type BoxType interface {
	// Returns true if box is of type (or subtype) BoxType
	IsInstance(box Box) bool

	AnonymousFrom(parent Box, children []Box) Box
}

// BoxFields is an abstract base class for all boxes.
type BoxFields struct {
	// Keep track of removed collapsing spaces for wrap opportunities.
	LeadingCollapsibleSpace  bool
	TrailingCollapsibleSpace bool

	// Default, may be overriden on instances.
	IsTableWrapper   bool
	IsFlexItem       bool
	IsForRootElement bool
	IsColumn         bool

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

	FirstLetterStyle, firstLineStyle pr.Properties

	PositionX, PositionY, Baseline                                       pr.Float
	Width, Height, MinWidth, MaxWidth, MinHeight, MaxHeight              pr.MaybeFloat
	Top, Bottom, Left, Right                                             pr.MaybeFloat
	MarginTop, MarginBottom, MarginLeft, MarginRight                     pr.MaybeFloat
	PaddingTop, PaddingBottom, PaddingLeft, PaddingRight                 pr.MaybeFloat
	BorderTopWidth, BorderRightWidth, BorderBottomWidth, BorderLeftWidth pr.MaybeFloat

	BorderTopLeftRadius, BorderTopRightRadius, BorderBottomRightRadius, BorderBottomLeftRadius MaybePoint

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

	FlexBasis                                                      pr.Value
	FlexBaseSize, TargetMainSize, Adjustment, HypotheticalMainSize pr.Float
	FlexFactor, ScaledFlexShrinkFactor                             pr.Float
	Frozen                                                         bool

	// ResumeAt *SkipStack
	// Index    int
}

func newBoxFields(elementTag string, style pr.Properties, children []Box) BoxFields {
	return BoxFields{elementTag: elementTag, Style: style, Children: children}
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

// isStart = isEnd = true
func CopyWithChildren(box Box, newChildren []Box, isStart bool, isEnd bool) Box {
	newBox := box.Copy()
	newBox.Box().Children = newChildren
	if box.Box().Style.GetBoxDecorationBreak() == "slice" {
		newBox.RemoveDecoration(newBox.Box(), !isStart, !isEnd)
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
			if _, ok := child.(instanceTableBox); ok {
				return child
			}
		}
	}
	return nil
}

// Translate changes the box’s position.
// Also update the children’s positions accordingly.
func (BoxFields) Translate(box Box, dx, dy pr.Float, ignoreFloats bool) {
	if dx == 0 && dy == 0 {
		return
	}
	box.Box().PositionX += dx
	box.Box().PositionY += dy
	for _, child := range box.allChildren() {
		if !(ignoreFloats && child.Box().IsFloated()) {
			child.Translate(child, dx, dy, ignoreFloats)
		}
	}
}

// ---- Heights and widths -----

// Width of the padding box.
func (self BoxFields) PaddingWidth() pr.Float {
	return self.Width.V() + self.PaddingLeft.V() + self.PaddingRight.V()
}

// Height of the padding box.
func (self BoxFields) PaddingHeight() pr.Float {
	return self.Height.V() + self.PaddingTop.V() + self.PaddingBottom.V()
}

// Width of the border box.
func (self BoxFields) BorderWidth() pr.Float {
	return self.PaddingWidth() + self.BorderLeftWidth.V() + self.BorderRightWidth.V()
}

// Height of th.V()e border box.
func (self BoxFields) BorderHeight() pr.Float {
	return self.PaddingHeight() + self.BorderTopWidth.V() + self.BorderBottomWidth.V()
}

// Width of the margin box (aka. outer box).
func (self BoxFields) MarginWidth() pr.Float {
	return self.BorderWidth() + self.MarginLeft.V() + self.MarginRight.V()
}

// Height of the margin box (aka. outer box).
func (self BoxFields) MarginHeight() pr.Float {
	return self.BorderHeight() + self.MarginTop.V() + self.MarginBottom.V()
}

// Corners positions

// Absolute horizontal position of the content box.
func (self BoxFields) ContentBoxX() pr.Float {
	return self.PositionX + self.MarginLeft.V() + self.PaddingLeft.V() + self.BorderLeftWidth.V()
}

// Absolute vertical position of the content box.
func (self BoxFields) ContentBoxY() pr.Float {
	return self.PositionY + self.MarginTop.V() + self.PaddingTop.V() + self.BorderTopWidth.V()
}

// Absolute horizontal position of the padding box.
func (self BoxFields) PaddingBoxX() pr.Float {
	return self.PositionX + self.MarginLeft.V() + self.BorderLeftWidth.V()
}

// Absolute vertical position of the padding box.
func (self BoxFields) PaddingBoxY() pr.Float {
	return self.PositionY + self.MarginTop.V() + self.BorderTopWidth.V()
}

// Absolute horizontal position of the border box.
func (self BoxFields) BorderBoxX() pr.Float {
	return self.PositionX + self.MarginLeft.V()
}

// Absolute vertical position of the border box.
func (self BoxFields) BorderBoxY() pr.Float {
	return self.PositionY + self.MarginTop.V()
}

// Return the rectangle where the box is clickable."""
// "Border area. That's the area that hit-testing is done on."
// http://lists.w3.org/Archives/Public/www-style/2012Jun/0318.html
// TODO: manage the border radii, use outerBorderRadii instead
func (self BoxFields) hitArea() (x, y, w, h pr.Float) {
	return self.BorderBoxX(), self.BorderBoxY(), self.BorderWidth(), self.BorderHeight()
}

type roundedBox struct {
	x, y, width, height                        pr.Float
	topLeft, topRight, bottomRight, bottomLeft Point
}

// Position, size and radii of a box inside the outer border box.
//bt, br, bb, and bl are distances from the outer border box,
//defining a rectangle to be rounded.
func (self BoxFields) roundedBox(bt, br, bb, bl pr.Float) roundedBox {
	tlr := self.BorderTopLeftRadius.V()
	trr := self.BorderTopRightRadius.V()
	brr := self.BorderBottomRightRadius.V()
	blr := self.BorderBottomLeftRadius.V()

	tlrx := pr.Float(utils.Max(0, float32(tlr[0]-bl)))
	tlry := pr.Float(utils.Max(0, float32(tlr[1]-bt)))
	trrx := pr.Float(utils.Max(0, float32(trr[0]-br)))
	trry := pr.Float(utils.Max(0, float32(trr[1]-bt)))
	brrx := pr.Float(utils.Max(0, float32(brr[0]-br)))
	brry := pr.Float(utils.Max(0, float32(brr[1]-bb)))
	blrx := pr.Float(utils.Max(0, float32(blr[0]-bl)))
	blry := pr.Float(utils.Max(0, float32(blr[1]-bb)))

	x := self.BorderBoxX() + bl
	y := self.BorderBoxY() + bt
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
	var ratio pr.Float = 1.
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

func (self BoxFields) roundedBoxRatio(ratio pr.Float) roundedBox {
	return self.roundedBox(
		self.BorderTopWidth.V()*ratio,
		self.BorderRightWidth.V()*ratio,
		self.BorderBottomWidth.V()*ratio,
		self.BorderLeftWidth.V()*ratio)
}

// Return the position, size and radii of the rounded padding box.
func (self BoxFields) roundedPaddingBox() roundedBox {
	return self.roundedBox(
		self.BorderTopWidth.V(),
		self.BorderRightWidth.V(),
		self.BorderBottomWidth.V(),
		self.BorderLeftWidth.V())
}

// Return the position, size and radii of the rounded border box.
func (self BoxFields) roundedBorderBox() roundedBox {
	return self.roundedBox(0, 0, 0, 0)
}

// Return the position, size and radii of the rounded content box.
func (self BoxFields) roundedContentBox() roundedBox {
	return self.roundedBox(
		self.BorderTopWidth.V()+self.PaddingTop.V(),
		self.BorderRightWidth.V()+self.PaddingRight.V(),
		self.BorderBottomWidth.V()+self.PaddingBottom.V(),
		self.BorderLeftWidth.V()+self.PaddingLeft.V())
}

// Positioning schemes

// Return whether this box is floated.
func (self BoxFields) IsFloated() bool {
	return self.Style.GetFloat() != "none"
}

// Return whether this box is in the absolute positioning scheme.
func (self BoxFields) IsAbsolutelyPositioned() bool {
	pos := self.Style.GetPosition()
	return !pos.Bool && pos.String == "absolute" || pos.String == "fixed"
}

// Return whether this box is a running element.
func (self BoxFields) IsRunning() bool {
	pos := self.Style.GetPosition()
	return pos.Bool
}

// Return whether this box is in normal flow.
func (self BoxFields) IsInNormalFlow() bool {
	return !(self.IsFloated() || self.IsAbsolutelyPositioned() || self.IsRunning())
}

// Start and end page values for named pages

// Return start and end page values.
func (b BoxFields) PageValues() (pr.Page, pr.Page) {
	start := b.Style.GetPage()
	end := start
	children := b.Children
	if len(children) > 0 {
		startBox, endBox := children[0], children[len(children)-1]
		childStart, _ := startBox.PageValues()
		_, childEnd := endBox.PageValues()
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
		self.MarginTop = pr.Float(0)
		self.PaddingTop = pr.Float(0)
		self.BorderTopWidth = pr.Float(0)
	case "right":
		self.MarginRight = pr.Float(0)
		self.PaddingRight = pr.Float(0)
		self.BorderRightWidth = pr.Float(0)
	case "left":
		self.MarginLeft = pr.Float(0)
		self.PaddingLeft = pr.Float(0)
		self.BorderLeftWidth = pr.Float(0)
	case "bottom":
		self.MarginBottom = pr.Float(0)
		self.PaddingBottom = pr.Float(0)
		self.BorderBottomWidth = pr.Float(0)
	}
}

func (BoxFields) RemoveDecoration(box *BoxFields, start, end bool) {
	if start {
		box.resetSpacing("top")
	}
	if end {
		box.resetSpacing("bottom")
	}
}
