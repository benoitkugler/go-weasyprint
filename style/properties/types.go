package properties

import "github.com/benoitkugler/go-weasyprint/style/parser"

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
	None        bool
	Decorations Set
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
	Name       string
	TypeOrUnit string
	Fallback   CssProperty
}

type Float float32

type Int int

type Ints3 [3]int

type Page struct {
	Valid  bool
	String string
	Page   int
}

// Dimension or "auto" or "cover" or "contain"
type Size struct {
	Width  Value
	Height Value
	String string
}

type Center struct {
	OriginX string
	OriginY string
	Pos     Point
}

type Color parser.Color

type ContentProperty struct {
	Type string

	// SStrings for type STRING, attr or string, counter, counters
	// Quote for type QUOTE
	// Url for URI
	Content InnerContent
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
	Int    int
	String string
}

type String string

type Value struct {
	Dimension
	String string
}

// ---------------------- Descriptors --------------------------

type Descriptor interface {
	isDescriptor()
}

type Contents []InnerContent

type NamedProperty struct {
	Name     string
	Property ValidatedProperty
}

type NamedProperties []NamedProperty

func (d String) isDescriptor()          {}
func (d IntString) isDescriptor()       {}
func (d Contents) isDescriptor()        {}
func (d SIntStrings) isDescriptor()     {}
func (d NamedProperties) isDescriptor() {}

// ---------------------- helpers types -----------------------------------
type CustomProperty []parser.Token

type SContentProp struct {
	String          string
	ContentProperty ContentProperty
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
	Value float32
}

type BoolString struct {
	Bool   bool
	String string
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

type NoneImage struct{}
type UrlImage string

type ColorStop struct {
	Color    Color
	Position Dimension
}

type DirectionType struct {
	Corner string
	Angle  float32
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

type SContents []SContent
type Dimensions []Dimension
