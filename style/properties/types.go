package properties

import (
	"github.com/benoitkugler/go-weasyprint/style/parser"
	"github.com/benoitkugler/go-weasyprint/utils"
)

// ------------- Top levels types, implementing CssProperty ------------

type StringSet struct {
	String   string
	Contents SContents
}

type Images []Image

type Centers []Center

type Sizes []Size

type Repeats [][2]string

type Strings []string

type SContent struct {
	String   string
	Contents ContentProperties
}

type NDecorations struct {
	Decorations utils.Set
	None        bool
}

type Transforms []SDimensions

type Values []Value

type SIntStrings struct {
	String string
	Values IntStrings
}

type SStrings struct {
	String  string
	Strings Strings
}

type SDimensions struct {
	String     string
	Dimensions Dimensions
}

type IntStrings []IntString

type Quotes struct {
	Open  Strings
	Close Strings
}

type ContentProperties []ContentProperty

type AttrData struct {
	Fallback   CssProperty
	Name       string
	TypeOrUnit string
}

type Float Fl

type Int int

type Ints3 [3]int

type Page struct {
	String string
	Valid  bool
	Page   int
}

// Dimension or "auto" or "cover" or "contain"
type Size struct {
	String string
	Width  Value
	Height Value
}

type Center struct {
	OriginX string
	OriginY string
	Pos     Point
}

type Color parser.Color

type ContentProperty struct {
	// SStrings for type STRING, attr or string, counter, counters
	// Quote for type QUOTE
	// Url for URI
	Content InnerContent

	Type string
}

type NamedString struct {
	Name   string
	String string
}

type Point [2]Dimension

type Marks struct {
	Crop  bool
	Cross bool
}

type IntString struct {
	String string
	Int    int
}

type String string

type Value struct {
	String string
	Dimension
}

type NamedProperty struct {
	Name     string
	Property ValidatedProperty
}

type NamedProperties []NamedProperty

// ---------------------- helpers types -----------------------------------
type CustomProperty []parser.Token

type SContentProp struct {
	ContentProperty ContentProperty
	String          string
}
type SContentProps []SContentProp

// guard for possible content properties
type InnerContent interface {
	isInnerContent()
}

type Unit uint8

// Dimension without unit is interpreted as float
type Dimension struct {
	Unit  Unit
	Value Float
}

type BoolString struct {
	String string
	Bool   bool
}

type Quote struct {
	Open   bool
	Insert bool
}

// Might be an existing image or a gradient
type Image interface {
	// InnerContent
	CssProperty
	isImage()
}

type (
	NoneImage struct{}
	UrlImage  string
)

type ColorStop struct {
	Color    Color
	Position Dimension
}

type DirectionType struct {
	Corner string
	Angle  Fl
}

type GradientSize struct {
	Keyword  string
	Explicit Point
}

type ColorsStops []ColorStop

type LinearGradient struct {
	ColorStops ColorsStops
	Direction  DirectionType
	Repeating  bool
}

type RadialGradient struct {
	ColorStops ColorsStops
	Shape      string
	Size       GradientSize
	Center     Center
	Repeating  bool
}

type (
	SContents  []SContent
	Dimensions []Dimension
)
