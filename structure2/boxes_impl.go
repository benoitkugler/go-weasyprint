package structure2

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strings"

	pr "github.com/benoitkugler/go-weasyprint/style/properties"

	"github.com/benoitkugler/go-weasyprint/style/tree"
)

// Complete generated.go for special cases.

// BoxFields is an abstract base class for all boxes.
type BoxFields struct {
	// Keep track of removed collapsing spaces for wrap opportunities.
	leadingCollapsibleSpace  bool
	trailingCollapsibleSpace bool

	// Default, may be overriden on instances.
	isTableWrapper       bool
	isForRootElement     bool
	isColumn             bool
	isAttachment         bool
	isListMarker         bool
	transformationMatrix interface{}

	bookmarkLabel string
	stringSet     pr.StringSet

	elementTag string
	style      tree.StyleFor

	firstLetterStyle, firstLineStyle tree.StyleFor

	positionX, positionY float64

	width, height float64

	marginTop, marginBottom, marginLeft, marginRight float64

	paddingTop, paddingBottom, paddingLeft, paddingRight float64

	borderTopWidth, borderRightWidth, borderBottomWidth, borderLeftWidth float64

	borderTopLeftRadius, borderTopRightRadius, borderBottomRightRadius, borderBottomLeftRadius interface{}

	viewportOverflow string

	children          []Box
	outsideListMarker Box
}

// BoxType enables passing type as value
type BoxType interface {
	AnonymousFrom(parent Box, children []Box) Box

	// Returns true if box is of type (or subtype) BoxType
	IsInstance(box Box) bool
}

func (t typeTableBox) IsInstance(box Box) bool {
	_, is := box.(TableBoxInstance)
	return is
}

func (t typeTextBox) AnonymousFrom(parent Box, children []Box) Box {
	log.Fatal("Can't create anonymous box from text box !")
}

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

func (self BoxFields) String() string {
	return fmt.Sprintf("<Box %s>", self.elementTag)
}

// Translate changes the box’s position.
// Also update the children’s positions accordingly.
func (self *BoxFields) Translate(dx, dy float64, ignoreFloats bool) {
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
func (self BoxFields) paddingWidth() float64 {
	return self.width + self.paddingLeft + self.paddingRight
}

// Height of the padding box.
func (self BoxFields) paddingHeight() float64 {
	return self.height + self.paddingTop + self.paddingBottom
}

// Width of the border box.
func (self BoxFields) borderWidth() float64 {
	return self.paddingWidth() + self.borderLeftWidth + self.borderRightWidth
}

// Height of the border box.
func (self BoxFields) borderHeight() float64 {
	return self.paddingHeight() + self.borderTopWidth + self.borderBottomWidth
}

// Width of the margin box (aka. outer box).
func (self BoxFields) marginWidth() float64 {
	return self.borderWidth() + self.marginLeft + self.marginRight
}

// Height of the margin box (aka. outer box).
func (self BoxFields) marginHeight() float64 {
	return self.borderHeight() + self.marginTop + self.marginBottom
}

// Corners positions

// Absolute horizontal position of the content box.
func (self BoxFields) contentBoxX() float64 {
	return self.positionX + self.marginLeft + self.paddingLeft + self.borderLeftWidth
}

// Absolute vertical position of the content box.
func (self BoxFields) contentBoxY() float64 {
	return self.positionY + self.marginTop + self.paddingTop + self.borderTopWidth
}

// Absolute horizontal position of the padding box.
func (self BoxFields) paddingBoxX() float64 {
	return self.positionX + self.marginLeft + self.borderLeftWidth
}

// Absolute vertical position of the padding box.
func (self BoxFields) paddingBoxY() float64 {
	return self.positionY + self.marginTop + self.borderTopWidth
}

// Absolute horizontal position of the border box.
func (self BoxFields) borderBoxX() float64 {
	return self.positionX + self.marginLeft
}

// Absolute vertical position of the border box.
func (self BoxFields) borderBoxY() float64 {
	return self.positionY + self.marginTop
}

// Return the rectangle where the box is clickable."""
// "Border area. That's the area that hit-testing is done on."
// http://lists.w3.org/Archives/Public/www-style/2012Jun/0318.html
// TODO: manage the border radii, use outerBorderRadii instead
func (self BoxFields) hitArea() (x float64, y float64, w float64, h float64) {
	return self.borderBoxX(), self.borderBoxY(), self.borderWidth(), self.borderHeight()
}

type roundedBox struct {
	x, y, width, height                        float64
	topLeft, topRight, bottomRight, bottomLeft point
}

// Position, size and radii of a box inside the outer border box.
//bt, br, bb, and bl are distances from the outer border box,
//defining a rectangle to be rounded.
func (self BoxFields) roundedBox(bt, br, bb, bl float64) roundedBox {
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

func (self BoxFields) roundedBoxRatio(ratio float64) roundedBox {
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
	return self.style.Strings["float"] != "none"
}

// Return whether this box is in the absolute positioning scheme.
func (self BoxFields) isAbsolutelyPositioned() bool {
	pos := self.style.Strings["position"]
	return pos == "absolute" || pos == "fixed"
}

// Return whether this box is in normal flow.
func (self BoxFields) isInNormalFlow() bool {
	return !(self.isFloated() || self.isAbsolutelyPositioned())
}

// Start and end page values for named pages

// Return start and end page values.
func (self BoxFields) pageValues() (int, int) {
	p := self.style.Page.Page
	return p, p
}

// Set to 0 the margin, padding and border of ``side``.
func (self *BoxFields) resetSpacing(side tree.Side) {
	self.style.Values[fmt.Sprintf("margin_%s", side)] = tree.ZeroPixels
	self.style.Values[fmt.Sprintf("padding_%s", side)] = tree.ZeroPixels
	self.style.Values[fmt.Sprintf("border_%s_width", side)] = tree.ZeroPixels

	switch side {
	case tree.Top:
		self.marginTop = 0
		self.paddingTop = 0
		self.borderTopWidth = 0
	case tree.Right:
		self.marginRight = 0
		self.paddingRight = 0
		self.borderRightWidth = 0
	case tree.Left:
		self.marginLeft = 0
		self.paddingLeft = 0
		self.borderLeftWidth = 0
	case tree.Bottom:
		self.marginBottom = 0
		self.paddingBottom = 0
		self.borderBottomWidth = 0
	}
}

func (self *BoxFields) removeDecoration(start, end bool) {
	if start || end {
		self.style = self.style.Copy()
	}
	ltr := self.style.Strings["direction"] == "ltr"
	if start {
		side := tree.Right
		if ltr {
			side = tree.Left
		}
		self.resetSpacing(side)
	}
	if end {
		side := tree.Left
		if ltr {
			side = tree.Right
		}
		self.resetSpacing(side)
	}
}

func (self BoxFields) AllChildren() []Box {
	return self.children
}

// A flat generator for a box, its children and descendants."""
func (self BoxFields) descendants() []Box {
	return []Box{&self}
}

func (p *ParentBox) init(elementTag string, style tree.StyleFor, children []Box) {
	p.Box.init(elementTag, style)
	p.children = children
}

func (self ParentBox) IsParentBox() bool {
	return true
}

func (self *ParentBox) removeDecoration(start, end bool) {
	if start || end {
		self.style = self.style.Copy()
	}
	if start {
		self.resetSpacing(tree.Top)
	}
	if end {
		self.resetSpacing(tree.Bottom)
	}
}

// A flat generator for a box, its children and descendants."""
func (self ParentBox) descendants() []Box {
	out := []Box{&self}
	for _, child := range self.children {
		out = append(out, child.descendants()...)
	}
	return out
}

// Create a new equivalent box with given ``newChildren``.
func CopyWithChildren(box Box, newChildren []Box, isStart, isEnd bool) Box {
	newBox := box.Copy()
	newBox.BaseBox().children = newChildren
	if !isStart {
		newBox.BaseBox().outsideListMarker = nil
	}
	newBox.removeDecoration(!isStart, !isEnd)
	return newBox
}

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

func (self BlockBox) AllChildren() []Box {
	if self.outsideListMarker != nil {
		return append(self.children, self.outsideListMarker)
	}
	return self.children
}

// func (self BlockBox) pageValues() (int, int) {
// 	return self.BlockContainerBox.pageValues()
// }

func NewLineBox(elementTag string, style tree.StyleFor, children []Box) *LineBox {
	if !style.Anonymous {
		log.Fatal("style must be anonymous")
	}
	var l LineBox
	l.ParentBox.init(elementTag, style, children)
	return &l
}

// Return the (x, y, w, h) rectangle where the box is clickable.
func (self InlineBox) hitArea() (x float64, y float64, w float64, h float64) {
	return self.borderBoxX(), self.positionY, self.borderWidth(), self.marginHeight()
}

func NewTextBox(elementTag string, style tree.StyleFor, text string) *TextBox {
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

// Return a new TextBox identical to this one except for the text.
func (self TextBox) copyWithText(text string) TextBox {
	if len(text) == 0 {
		log.Fatal("empty text")
	}
	newBox := self
	newBox.text = text
	return newBox
}

// Definitions for the rules generating anonymous table boxes
// http://www.w3.org/TR/CSS21/tables.html#anonymous-boxes

func (self TableBox) AllChildren() []Box {
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

func (self TableRowGroupBox) IsProperChild(parent Box) bool {
	switch parent.(type) {
	case *TableBox, *InlineTableBox:
		return true
	default:
		return false
	}
}

func (self TableRowBox) IsProperChild(parent Box) bool {
	switch parent.(type) {
	case *TableBox, *InlineTableBox, *TableRowGroupBox:
		return true
	default:
		return false
	}
}

func (self TableColumnGroupBox) IsProperChild(parent Box) bool {
	switch parent.(type) {
	case *TableBox, *InlineTableBox:
		return true
	default:
		return false
	}
}

// Return cells that originate in the group's columns.
func (self TableColumnGroupBox) getCells() []Box {
	var out []Box
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

// Return cells that originate in the column.
// May be overriden on instances.
func (self TableColumnBox) getCells() []Box {
	return nil
}

func (self TableColumnBox) IsProperChild(parent Box) bool {
	switch parent.(type) {
	case *TableBox, *InlineTableBox, *TableColumnGroupBox:
		return true
	default:
		return false
	}
}

func (self TableCaptionBox) IsProperChild(parent Box) bool {
	switch parent.(type) {
	case *TableBox, *InlineTableBox:
		return true
	default:
		return false
	}
}

func (self PageBox) String() string {
	return fmt.Sprintf("<PageBox %s>", self.pageType)
}

func (self MarginBox) String() string {
	return fmt.Sprintf("<MarginBox %s>", self.atKeyword)
}

func TableCaptionBoxIsInstance(box Box) bool {
	_, is := box.(*TableCaptionBox)
	return is
}
