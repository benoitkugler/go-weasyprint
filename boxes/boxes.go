// Package boxes defines the basic elements of the formatting structure,
// as a tree of boxes.
//
// This tree is build from an HTML document by this package, but the boxes
// are not correctly positionned yet (see the layout package).
package boxes

import (
	"unicode/utf8"

	"github.com/benoitkugler/go-weasyprint/images"

	"github.com/benoitkugler/go-weasyprint/style/parser"
	"github.com/benoitkugler/go-weasyprint/style/tree"

	pr "github.com/benoitkugler/go-weasyprint/style/properties"
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

// http://stackoverflow.com/questions/16317534/
var asciiToWide = map[rune]rune{}

func init() {
	for i := 0x21; i < utf8.RuneSelf; i++ {
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

type BoxITF = Box

// Box is the common interface grouping all possible boxes
type Box interface {
	// IsClassicalBox returns true for all standard boxes defined in this package, but false
	// for the special ones, defined in other packages, like AbsolutePlaceholder or StackingContext.
	IsClassicalBox() bool

	tree.Box

	Type() BoxType

	Box() *BoxFields
	Copy() Box
	AllChildren() []Box
	Translate(box Box, dx, dy pr.Float, ignoreFloats bool)
	RemoveDecoration(box *BoxFields, isStart, isEnd bool)
	PageValues() (pr.Page, pr.Page)
}

type Background struct {
	ImageRendering pr.String
	Layers         []BackgroundLayer
	Color          parser.RGBA
}

type Area struct {
	String string
	Rect   pr.Rectangle
}

type Position struct {
	Point  MaybePoint
	String string
}

type Repeat struct {
	String string
	Reps   [2]string
}

type BackgroundLayer struct {
	Image           images.Image
	Position        Position
	Repeat          Repeat
	ClippedBoxes    []RoundedBox
	PaintingArea    pr.Rectangle
	PositioningArea pr.Rectangle
	Size            [2]pr.Float // width, height
	Unbounded       bool
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
	IsLeader         bool

	properTableChild       bool
	internalTableOrCaption bool
	tabularContainer       bool

	IsAttachment bool
	// isListMarker         bool

	BookmarkLabel string

	StringSet pr.ContentProperties

	ElementTag string
	Style      pr.ElementStyle

	FirstLetterStyle, firstLineStyle pr.ElementStyle

	PositionX, PositionY                                                 pr.Float
	Width, Height, MinWidth, MaxWidth, MinHeight, MaxHeight              pr.MaybeFloat
	Top, Bottom, Left, Right                                             pr.MaybeFloat
	MarginTop, MarginBottom, MarginLeft, MarginRight                     pr.MaybeFloat
	PaddingTop, PaddingBottom, PaddingLeft, PaddingRight                 pr.MaybeFloat
	BorderTopWidth, BorderRightWidth, BorderBottomWidth, BorderLeftWidth pr.MaybeFloat

	BorderTopLeftRadius, BorderTopRightRadius, BorderBottomRightRadius, BorderBottomLeftRadius MaybePoint

	ViewportOverflow string

	Children []Box
	// outsideListMarker Box

	missingLink         tree.Box
	cachedCounterValues tree.CounterValues

	IsHeader, IsFooter bool

	Baseline                      pr.MaybeFloat
	ComputedHeight, ContentHeight pr.MaybeFloat
	VerticalAlign                 string
	Empty                         bool
	span                          int
	Colspan                       int
	Rowspan                       int

	GridX int
	Index int

	FlexBasis                                                      pr.Value
	FlexBaseSize, TargetMainSize, Adjustment, HypotheticalMainSize pr.Float
	FlexFactor, ScaledFlexShrinkFactor                             pr.Float
	Frozen                                                         bool

	GetCells func() []Box // closure which may have default implementation or be set

	ResumeAt *tree.IntList

	Background *Background

	RemoveDecorationSides [4]bool
}

func newBoxFields(elementTag string, style pr.ElementStyle, children []Box) BoxFields {
	return BoxFields{ElementTag: elementTag, Style: style, Children: children}
}

func (box *BoxFields) AllChildren() []Box {
	return box.Children
}

// ContainingBlock implements an interface required for layout.
func (box *BoxFields) ContainingBlock() (width, height pr.MaybeFloat) {
	return box.Width, box.Height
}

func (*BoxFields) isBox() {}

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

// Create a new equivalent box with given ``newChildren``."""
func CopyWithChildren(box Box, newChildren []Box) Box {
	newBox := box.Copy()
	newBox.Box().Children = newChildren
	// Clear and reset removed decorations as we don't want to keep the
	// previous data, for example when a box is split between two pages.
	box.Box().RemoveDecorationSides = [4]bool{}
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

// Descendants returns `b` and its children,
// and their children, etc...
func Descendants(b Box) []Box {
	out := []Box{b}
	for _, child := range b.Box().Children {
		out = append(out, Descendants(child)...)
	}
	return out
}

func (box *BoxFields) GetWrappedTable() TableBoxITF {
	if box.IsTableWrapper {
		for _, child := range box.Children {
			if asTable, ok := child.(TableBoxITF); ok {
				return asTable
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
	for _, child := range box.AllChildren() {
		if !(ignoreFloats && child.Box().IsFloated()) {
			child.Translate(child, dx, dy, ignoreFloats)
		}
	}
}

// ---- Heights and widths -----

// Width of the padding box.
func (b *BoxFields) PaddingWidth() pr.Float {
	return b.Width.V() + b.PaddingLeft.V() + b.PaddingRight.V()
}

// Height of the padding box.
func (b *BoxFields) PaddingHeight() pr.Float {
	return b.Height.V() + b.PaddingTop.V() + b.PaddingBottom.V()
}

// Width of the border box.
func (b *BoxFields) BorderWidth() pr.Float {
	return b.PaddingWidth() + b.BorderLeftWidth.V() + b.BorderRightWidth.V()
}

// Height of the border box.
func (b *BoxFields) BorderHeight() pr.Float {
	return b.PaddingHeight() + b.BorderTopWidth.V() + b.BorderBottomWidth.V()
}

// Width of the margin box (aka. outer box).
func (b *BoxFields) MarginWidth() pr.Float {
	return b.BorderWidth() + b.MarginLeft.V() + b.MarginRight.V()
}

// Height of the margin box (aka. outer box).
func (b *BoxFields) MarginHeight() pr.Float {
	return b.BorderHeight() + b.MarginTop.V() + b.MarginBottom.V()
}

// Corners positions

// Absolute horizontal position of the content box.
func (b *BoxFields) ContentBoxX() pr.Float {
	return b.PositionX + b.MarginLeft.V() + b.PaddingLeft.V() + b.BorderLeftWidth.V()
}

// Absolute vertical position of the content box.
func (b *BoxFields) ContentBoxY() pr.Float {
	return b.PositionY + b.MarginTop.V() + b.PaddingTop.V() + b.BorderTopWidth.V()
}

// Absolute horizontal position of the padding box.
func (b *BoxFields) PaddingBoxX() pr.Float {
	return b.PositionX + b.MarginLeft.V() + b.BorderLeftWidth.V()
}

// Absolute vertical position of the padding box.
func (b *BoxFields) PaddingBoxY() pr.Float {
	return b.PositionY + b.MarginTop.V() + b.BorderTopWidth.V()
}

// Absolute horizontal position of the border box.
func (b *BoxFields) BorderBoxX() pr.Float {
	return b.PositionX + b.MarginLeft.V()
}

// Absolute vertical position of the border box.
func (b *BoxFields) BorderBoxY() pr.Float {
	return b.PositionY + b.MarginTop.V()
}

// HitArea returns the rectangle where the box is clickable.
func HitArea(box Box) pr.Rectangle {
	b := box.Box()
	if _, isInline := box.(InlineBoxITF); isInline {
		return pr.Rectangle{b.BorderBoxX(), b.PositionY, b.BorderWidth(), b.MarginHeight()}
	}
	return pr.Rectangle{b.BorderBoxX(), b.BorderBoxY(), b.BorderWidth(), b.BorderHeight()}
}

type RoundedBox struct {
	X, Y, Width, Height                        pr.Float
	TopLeft, TopRight, BottomRight, BottomLeft Point
}

// Position, size and radii of a box inside the outer border box.
// bt, br, bb, and bl are distances from the outer border box,
// defining a rectangle to be rounded.
func (b *BoxFields) roundedBox(bt, br, bb, bl pr.Float) RoundedBox {
	tlr := b.BorderTopLeftRadius.V()
	trr := b.BorderTopRightRadius.V()
	brr := b.BorderBottomRightRadius.V()
	blr := b.BorderBottomLeftRadius.V()

	tlrx := pr.Max(0, tlr[0]-bl)
	tlry := pr.Max(0, tlr[1]-bt)
	trrx := pr.Max(0, trr[0]-br)
	trry := pr.Max(0, trr[1]-bt)
	brrx := pr.Max(0, brr[0]-br)
	brry := pr.Max(0, brr[1]-bb)
	blrx := pr.Max(0, blr[0]-bl)
	blry := pr.Max(0, blr[1]-bb)

	x := b.BorderBoxX() + bl
	y := b.BorderBoxY() + bt
	width := b.BorderWidth() - bl - br
	height := b.BorderHeight() - bt - bb

	// Fix overlapping curves
	// See http://www.w3.org/TR/css3-background/#corner-overlap
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
	return RoundedBox{
		X: x, Y: y, Width: width, Height: height,
		TopLeft:     Point{tlrx * ratio, tlry * ratio},
		TopRight:    Point{trrx * ratio, trry * ratio},
		BottomRight: Point{brrx * ratio, brry * ratio},
		BottomLeft:  Point{blrx * ratio, blry * ratio},
	}
}

func (b *BoxFields) RoundedBoxRatio(ratio pr.Float) RoundedBox {
	return b.roundedBox(
		b.BorderTopWidth.V()*ratio,
		b.BorderRightWidth.V()*ratio,
		b.BorderBottomWidth.V()*ratio,
		b.BorderLeftWidth.V()*ratio)
}

// Return the position, size and radii of the rounded padding box.
func (b *BoxFields) RoundedPaddingBox() RoundedBox {
	return b.RoundedBoxRatio(1)
}

// Return the position, size and radii of the rounded border box.
func (b *BoxFields) RoundedBorderBox() RoundedBox {
	return b.roundedBox(0, 0, 0, 0)
}

// Return the position, size and radii of the rounded content box.
func (b *BoxFields) RoundedContentBox() RoundedBox {
	return b.roundedBox(
		b.BorderTopWidth.V()+b.PaddingTop.V(),
		b.BorderRightWidth.V()+b.PaddingRight.V(),
		b.BorderBottomWidth.V()+b.PaddingBottom.V(),
		b.BorderLeftWidth.V()+b.PaddingLeft.V())
}

// Positioning schemes

// Return whether this box is floated.
func (b *BoxFields) IsFloated() bool {
	return b.Style.GetFloat() != "none"
}

// Return whether this box is in the absolute positioning scheme.
func (b *BoxFields) IsAbsolutelyPositioned() bool {
	pos := b.Style.GetPosition()
	return !pos.Bool && pos.String == "absolute" || pos.String == "fixed"
}

// Return whether this box is a running element.
func (b *BoxFields) IsRunning() bool {
	pos := b.Style.GetPosition()
	return pos.Bool
}

// Return whether this box is in normal flow.
func (b *BoxFields) IsInNormalFlow() bool {
	return !(b.IsFloated() || b.IsAbsolutelyPositioned() || b.IsRunning())
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

type Side uint8

const (
	SLeft Side = iota
	SRight
	STop
	SBottom
)

func (s Side) String() string {
	switch s {
	case SLeft:
		return "left"
	case SRight:
		return "right"
	case STop:
		return "top"
	case SBottom:
		return "bottom"
	default:
		return ""
	}
}

// Set to 0 the margin, padding and border of ``side``.
func (b *BoxFields) ResetSpacing(side Side) {
	b.RemoveDecorationSides[side] = true

	switch side {
	case STop:
		b.MarginTop = pr.Float(0)
		b.PaddingTop = pr.Float(0)
		b.BorderTopWidth = pr.Float(0)
	case SRight:
		b.MarginRight = pr.Float(0)
		b.PaddingRight = pr.Float(0)
		b.BorderRightWidth = pr.Float(0)
	case SLeft:
		b.MarginLeft = pr.Float(0)
		b.PaddingLeft = pr.Float(0)
		b.BorderLeftWidth = pr.Float(0)
	case SBottom:
		b.MarginBottom = pr.Float(0)
		b.PaddingBottom = pr.Float(0)
		b.BorderBottomWidth = pr.Float(0)
	}
}

func (*BoxFields) RemoveDecoration(box *BoxFields, start, end bool) {
	if box.Style.GetBoxDecorationBreak() == "clone" {
		return
	}
	if start {
		box.ResetSpacing(STop)
	}
	if end {
		box.ResetSpacing(SBottom)
	}
}

// IsInProperParents returns true if `t` is one of the
// the proper parents of `type_`
func (t BoxType) IsInProperParents(type_ BoxType) bool {
	switch type_ {
	case TableRowGroupBoxT, TableColumnGroupBoxT, TableCaptionBoxT:
		return t == TableBoxT || t == InlineTableBoxT
	case TableRowBoxT:
		return t == TableBoxT || t == InlineTableBoxT || t == TableRowGroupBoxT
	case TableColumnBoxT:
		return t == TableBoxT || t == InlineTableBoxT || t == TableColumnGroupBoxT
	default:
		return false
	}
}

// shared utils

type BC struct {
	Text string
	C    []SerBox
}

type SerBox struct {
	Tag     string
	Type    BoxType
	Content BC
}

func (s SerBox) equals(other SerBox) bool {
	if s.Tag != other.Tag || s.Type != other.Type || s.Content.Text != other.Content.Text {
		return false
	}
	return SerializedBoxEquals(s.Content.C, other.Content.C)
}

func SerializedBoxEquals(l1, l2 []SerBox) bool {
	if len(l1) != len(l2) {
		return false
	}
	for j := range l1 {
		if !l1[j].equals(l2[j]) {
			return false
		}
	}
	return true
}

// Transform a box list into a structure easier to compare for testing.
func Serialize(boxList []Box) []SerBox {
	out := make([]SerBox, len(boxList))
	for i, box := range boxList {
		out[i].Tag = box.Box().ElementTag
		out[i].Type = box.Type()
		// all concrete boxes are either text, replaced, column or parent.
		if boxT, ok := box.(*TextBox); ok {
			out[i].Content.Text = boxT.Text
		} else if _, ok := box.(ReplacedBoxITF); ok {
			out[i].Content.Text = "<replaced>"
		} else {
			var cg []Box
			if table, ok := box.(TableBoxITF); ok {
				cg = table.Table().ColumnGroups
			}
			cg = append(cg, box.Box().Children...)
			out[i].Content.C = Serialize(cg)
		}
	}
	return out
}
