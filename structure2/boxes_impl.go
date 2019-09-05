package structure2

import "github.com/benoitkugler/go-weasyprint/css"

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
	stringSet     []css.NameValue

	elementTag string
	style      css.StyleDict

	firstLetterStyle, firstLineStyle css.StyleDict

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
