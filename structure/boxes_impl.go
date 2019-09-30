package structure

import (
	"errors"
	"fmt"
	"log"
	"strings"

	pr "github.com/benoitkugler/go-weasyprint/style/properties"
	"github.com/benoitkugler/go-weasyprint/utils"
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

	bookmarkLabel pr.ContentProperties
	stringSet     pr.StringSet

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
	return nil
}

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

type TBD struct{}

func (self BoxFields) String() string {
	return fmt.Sprintf("<Box %s>", self.elementTag)
}

// Translate changes the box’s position.
// Also update the children’s positions accordingly.
func (self *BoxFields) Translate(dx, dy float32, ignoreFloats bool) {
	translate(self.children, self, dx, dy, ignoreFloats)
}

func translate(children []Box, box *BoxFields, dx, dy float32, ignoreFloats bool) {
	if dx == 0 && dy == 0 {
		return
	}
	box.positionX += dx
	box.positionY += dy
	for _, child := range children {
		if !(ignoreFloats && child.Box().isFloated()) {
			child.Translate(dx, dy, ignoreFloats)
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
func (self BoxFields) pageValues() (pr.Page, pr.Page) {
	p := self.style.GetPage()
	return p, p
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

func (self *BoxFields) removeDecoration(start, end bool) {
	if start {
		self.resetSpacing("top")
	}
	if end {
		self.resetSpacing("bottom")
	}
}

// Create a new equivalent box with given ``newChildren``.
// isStart=true, isEnd=true
func copyWithChildren(box Box, newChildren []Box, isStart, isEnd bool) Box {
	newBox := box.Copy()
	newBox.Box().children = newChildren
	if box.Box().style.GetBoxDecorationBreak() == "slice" {
		newBox.removeDecoration(!isStart, !isEnd)
	}
	return newBox
}

func deepCopy(box Box) Box {
	result := box.Copy()
	l := result.Box().children
	for i, v := range l {
		l[i] = deepCopy(v)
	}
	return result
}

// A flat generator for a box, its children and descendants."""
func descendants(box Box) []Box {
	out := []Box{box}
	for _, child := range box.Box().children {
		out = append(out, descendants(child)...)
	}
	return out
}

// Get the table wrapped by the box.
func (self ParentBox) getWrappedTable() (TableBoxInstance, error) {
	if self.isTableWrapper {
		for _, child := range self.children {
			if asTable, ok := child.(TableBoxInstance); ok {
				return asTable, nil
			}
		}
		return nil, errors.New("Table wrapper without a table")
	}
	return nil, nil
}

func (self ParentBox) pageValues() (pr.Page, pr.Page) {
	start, end := self.Box().pageValues()
	if len(self.children) > 0 {
		startBox, endBox := self.children[0], self.children[len(self.children)-1]
		childStart, _ := startBox.Box().pageValues()
		_, childEnd := endBox.Box().pageValues()
		if !childStart.IsNone() {
			start = childStart
		}
		if !childEnd.IsNone() {
			end = childEnd
		}
	}
	return start, end
}

func LineBoxAnonymousFrom(parent Box, children []Box) *LineBox {
	parentBox := ParentBoxAnonymousFrom(parent, children)
	out := LineBox{ParentBox: *parentBox, textOverflow: "clip"}
	if parentBox.style.GetOverflow() != "visible" {
		out.textOverflow = parentBox.style.GetTextOverflow()
	}
	return &out
}

// func (self BlockBox) pageValues() (int, int) {
// 	return self.BlockContainerBox.pageValues()
// }

func (self InlineLevelBox) removeDecoration(start, end bool) {
	ltr := self.style.GetDirection() == "ltr"
	if start {
		side := "right"
		if ltr {
			side = "left"
		}
		self.resetSpacing(side)
	}
	if end {
		side := "left"
		if ltr {
			side = "right"
		}
		self.resetSpacing(side)
	}
}

func NewInlineBox(elementTag string, style pr.Properties, children []Box) *InlineBox {
	out := InlineBox{}
	parent := NewInlineLevelBox(elementTag, style, children)
	out.InlineLevelBox = *parent
	out.ParentBox = ParentBox{BoxFields: parent.BoxFields}
	return &out
}

// Return the (x, y, w, h) rectangle where the box is clickable.
func (self InlineBox) hitArea() (x float32, y float32, w float32, h float32) {
	return self.InlineLevelBox.borderBoxX(), self.InlineLevelBox.positionY, self.InlineLevelBox.borderWidth(), self.InlineLevelBox.marginHeight()
}

func NewTextBox(elementTag string, style pr.Properties, text string) *TextBox {
	if len(text) == 0 {
		log.Fatalf("empty text")
	}
	box := NewInlineLevelBox(elementTag, style, nil)
	textTransform := style.GetTextTransform()
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
			text = strings.Map(func(u rune) rune {
				rep, in := asciiToWide[u]
				if !in {
					return -1
				}
				return rep
			}, text)
		}
	}
	if style.GetHyphens() == "none" {
		text = strings.ReplaceAll(text, "\u00AD", "") //  U+00AD SOFT HYPHEN (SHY)
	}
	out := TextBox{InlineLevelBox: *box, text: text}
	return &out
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

func (self *InlineBlockBox) Translate(dx, dy float32, ignoreFloats bool) {
	self.AtomicInlineLevelBox.Translate(dx, dy, ignoreFloats)
}

// Definitions for the rules generating anonymous table boxes
// http://www.w3.org/TR/CSS21/tables.html#anonymous-boxes

func (self *TableBox) Translate(dx, dy float32, ignoreFloats bool) {
	if dx == 0 && dy == 0 {
		return
	}
	for index, position := range self.columnPositions {
		self.columnPositions[index] = position + dx
	}
	translate(append(self.children, self.columnGroups...), self.Box(), dx, dy, ignoreFloats)
}

func (self TableBox) pageValues() (pr.Page, pr.Page) {
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

type withCells interface {
	getCells() []Box
}

// Return cells that originate in the group's columns.
func (self *TableColumnGroupBox) getCells() []Box {
	var out []Box
	for _, column := range self.children {
		for _, cell := range column.(withCells).getCells() {
			out = append(out, cell)
		}
	}
	return out
}

// Return cells that originate in the column.
// May be overriden on instances.
func (self *TableColumnBox) getCells() []Box {
	return []Box{}
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

func TableCaptionBoxIsInstance(box Box) bool {
	_, is := box.(*TableCaptionBox)
	return is
}

func NewPageBox(pageType utils.PageElement, style pr.Properties) *PageBox {
	parent := NewParentBox("", style, nil)
	return &PageBox{ParentBox: *parent, pageType: pageType}
}

func (self PageBox) String() string {
	return fmt.Sprintf("<PageBox %s>", self.pageType)
}

func NewMarginBox(atKeyword string, style pr.Properties) *MarginBox {
	b := NewBlockContainerBox("", style, nil)
	return &MarginBox{BlockContainerBox: *b, atKeyword: atKeyword}
}

func (self MarginBox) String() string {
	return fmt.Sprintf("<MarginBox %s>", self.atKeyword)
}
