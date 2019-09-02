package css

// import (
// 	"math"
// 	"strings"
// )

// var (
// 	ConvertersValue = map[string]func(Value) CssProperty{
// 		"top":                  valueToLength,
// 		"right":                valueToLength,
// 		"left":                 valueToLength,
// 		"bottom":               valueToLength,
// 		"margin_top":           valueToLength,
// 		"margin_right":         valueToLength,
// 		"margin_bottom":        valueToLength,
// 		"margin_left":          valueToLength,
// 		"height":               valueToLength,
// 		"width":                valueToLength,
// 		"min_width":            valueToLength,
// 		"min_height":           valueToLength,
// 		"max_width":            valueToLength,
// 		"max_height":           valueToLength,
// 		"padding_top":          valueToLength,
// 		"padding_right":        valueToLength,
// 		"padding_bottom":       valueToLength,
// 		"padding_left":         valueToLength,
// 		"text_indent":          valueToLength,
// 		"hyphenate_limit_zone": valueToLength,

// 		"bleed_left":   valueToBleed,
// 		"bleed_right":  valueToBleed,
// 		"bleed_top":    valueToBleed,
// 		"bleed_bottom": valueToBleed,

// 		"border_top_width":    valueToBorderWidth,
// 		"border_right_width":  valueToBorderWidth,
// 		"border_left_width":   valueToBorderWidth,
// 		"border_bottom_width": valueToBorderWidth,
// 		"column_rule_width":   valueToBorderWidth,
// 		"outline_width":       valueToBorderWidth,

// 		"column_width":   valueToColumnWidth,
// 		"column_gap":     valueToColumnGap,
// 		"font_size":      valueToFontSize,
// 		"font_weight":    valueToFontWeight,
// 		"line_height":    valueToLineHeight,
// 		"tab_size":       valueToTabSize,
// 		"vertical_align": valueToVerticalAlign,
// 		"word_spacing":   valueToWordSpacing,
// 		"letter_spacing": valueToPixelLength,
// 	}
// 	ConvertersString = map[string]func(string) CssProperty{
// 		"break_after":  func(s string) CssProperty { return Break(s) },
// 		"break_before": func(s string) CssProperty { return Break(s) },
// 		"display":      func(s string) CssProperty { return Display(s) },
// 		"float":        func(s string) CssProperty { return Floating(s) },
// 	}
// 	ConvertersLink = map[string]func(Link) CssProperty{
// 		"link":   func(l Link) CssProperty { return l },
// 		"anchor": func(l Link) CssProperty { return Anchor(l) },
// 		"lang":   func(l Link) CssProperty { return Lang(l) },
// 	}
// )

// type StringContent struct {
// 	Name   string
// 	Values []ContentProperty
// }

// func (s StringContent) IsNil() bool {
// 	return s.Name == "" && s.Values == nil
// }

// func (s StringContent) Copy() StringContent {
// 	out := s
// 	out.Values = append([]ContentProperty{}, s.Values...)
// 	return out
// }

// type StringSet struct {
// 	String   string
// 	Contents []StringContent
// }

// func (s StringSet) IsNil() bool {
// 	return s.String == "" && s.Contents == nil
// }

// func (s StringSet) Copy() StringSet {
// 	out := s
// 	out.Contents = make([]StringContent, len(s.Contents))
// 	for index, l := range s.Contents {
// 		out.Contents[index] = l.Copy()
// 	}
// 	return out
// }

// // background-image
// type BackgroundImage []Image

// // background-position
// type BackgroundPosition []Center

// // background-size
// type BackgroundSize []Size

// type BackgroundRepeat [][2]string

// type ContentType int

// const (
// 	ContentQUOTE ContentType = iota + 1 // so that zero field corresponds to null content
// 	ContentSTRING
// 	ContentURI
// 	ContentAttr
// 	ContentCounter
// 	ContentCounters
// 	ContentString
// 	ContentContent
// )

// type ContentProperty struct {
// 	Type ContentType

// 	// Next are values fields
// 	String  string   // for type STRING, URI, attr
// 	Quote   quote    // for type QUOTE
// 	Strings []string // for type string, counter, counters
// }

// func (cp ContentProperty) IsNil() bool {
// 	return cp.Type == 0
// }

// // content
// type Content struct {
// 	String string // 'none' ou 'normal'
// 	List   []ContentProperty
// }

// func (c Content) IsNil() bool {
// 	return c.String == "" && c.List == nil
// }

// // deep copy
// func (c Content) Copy() Content {
// 	out := c
// 	out.List = append([]ContentProperty{}, c.List...)
// 	return out
// }

// type TextDecoration struct {
// 	None        bool
// 	Decorations Set
// }

// // width, height
// type PageSize [2]Dimension

// // transform
// type Transforms []Transform

// // transform-origin
// // border-spacing
// // size
// // clip
// // border-top-left-radius
// // border-top-right-radius
// // border-bottom-left-radius
// // border-bottom-right-radius
// type Lengths []Value

// // marks
// type Marks struct {
// 	Crop, Cross bool
// }

// func (m Marks) IsNone() bool {
// 	return m == Marks{}
// }

// // break_after
// // break_before
// type Break string

// // display
// type Display string

// // float
// type Floating string

// // Standard string property
// type String string

// // top
// // right
// // left
// // bottom
// // margin_top
// // margin_right
// // margin_bottom
// // margin_left
// // height
// // width
// // min_width
// // min_height
// // max_width
// // max_height
// // padding_top
// // padding_right
// // padding_bottom
// // padding_left
// // text_indent
// // hyphenate_limit_zone
// type Length Value

// // bleed_left
// // bleed_right
// // bleed_top
// // bleed_bottom
// type Bleed Value

// // border_top_width
// // border_right_width
// // border_left_width
// // border_bottom_width
// // column_rule_width
// // outline_width
// type BorderWidth Value

// // letter_spacing
// type PixelLength Value

// // column_width
// type ColumnWidth Value

// // column_gap
// type ColumnGap Value

// // font_size
// type FontSize Value

// // font_weight
// type FontWeight Value

// // line_height
// type LineHeight Value

// // tab_size
// type TabSize Value

// // vertical_align
// type VerticalAlign Value

// // word_spacing
// type WordSpacing Value

// // link
// type Link struct {
// 	Type string
// 	Attr string
// }

// // anchor
// type Anchor Link

// // lang
// type Lang Link

// func valueToLength(v Value) CssProperty        { return Length(v) }
// func valueToBleed(v Value) CssProperty         { return Bleed(v) }
// func valueToPixelLength(v Value) CssProperty   { return PixelLength(v) }
// func valueToBorderWidth(v Value) CssProperty   { return BorderWidth(v) }
// func valueToColumnWidth(v Value) CssProperty   { return ColumnWidth(v) }
// func valueToColumnGap(v Value) CssProperty     { return ColumnGap(v) }
// func valueToFontSize(v Value) CssProperty      { return FontSize(v) }
// func valueToFontWeight(v Value) CssProperty    { return FontWeight(v) }
// func valueToLineHeight(v Value) CssProperty    { return LineHeight(v) }
// func valueToTabSize(v Value) CssProperty       { return TabSize(v) }
// func valueToVerticalAlign(v Value) CssProperty { return VerticalAlign(v) }
// func valueToWordSpacing(v Value) CssProperty   { return WordSpacing(v) }

// func (v Value) SetOn(name string, s *StyleDict) {
// 	s.Values[name] = v
// }

// func (v Length) SetOn(name string, s *StyleDict) {
// 	Value(v).SetOn(name, s)
// }
// func (v Bleed) SetOn(name string, s *StyleDict) {
// 	Value(v).SetOn(name, s)
// }
// func (v PixelLength) SetOn(name string, s *StyleDict) {
// 	Value(v).SetOn(name, s)
// }
// func (v BorderWidth) SetOn(name string, s *StyleDict) {
// 	Value(v).SetOn(name, s)
// }
// func (v ColumnWidth) SetOn(name string, s *StyleDict) {
// 	Value(v).SetOn(name, s)
// }
// func (v ColumnGap) SetOn(name string, s *StyleDict) {
// 	Value(v).SetOn(name, s)
// }
// func (v FontSize) SetOn(name string, s *StyleDict) {
// 	Value(v).SetOn(name, s)
// }
// func (v FontWeight) SetOn(name string, s *StyleDict) {
// 	Value(v).SetOn(name, s)
// }
// func (v LineHeight) SetOn(name string, s *StyleDict) {
// 	Value(v).SetOn(name, s)
// }
// func (v TabSize) SetOn(name string, s *StyleDict) {
// 	Value(v).SetOn(name, s)
// }
// func (v VerticalAlign) SetOn(name string, s *StyleDict) {
// 	Value(v).SetOn(name, s)
// }
// func (v WordSpacing) SetOn(name string, s *StyleDict) {
// 	Value(v).SetOn(name, s)
// }

// func (v Break) SetOn(name string, s *StyleDict) {
// 	s.Strings[name] = string(v)
// }
// func (v Display) SetOn(name string, s *StyleDict) {
// 	s.Strings[name] = string(v)
// }
// func (v Floating) SetOn(name string, s *StyleDict) {
// 	s.Strings[name] = string(v)
// }
// func (v String) SetOn(name string, s *StyleDict) {
// 	s.Strings[name] = string(v)
// }

// func (v Link) SetOn(name string, s *StyleDict) {
// 	s.Links[name] = v
// }
// func (v Anchor) SetOn(name string, s *StyleDict) {
// 	s.Links[name] = Link(v)
// }
// func (v Lang) SetOn(name string, s *StyleDict) {
// 	s.Links[name] = Link(v)
// }

// func (v Lengths) SetOn(name string, s *StyleDict) {
// 	s.Lengthss[name] = v
// }

// func (v Color) SetOn(name string, s *StyleDict) {
// 	s.Colors[name] = v
// }

// func (v CounterResets) SetOn(name string, s *StyleDict) {
// 	s.CounterReset = v
// }
// func (v CounterIncrements) SetOn(name string, s *StyleDict) {
// 	s.CounterIncrement = v
// }
// func (v Page) SetOn(name string, s *StyleDict) {
// 	s.Page = v
// }

// func (v BackgroundImage) SetOn(name string, s *StyleDict) {
// 	s.BackgroundImage = v
// }
// func (v BackgroundPosition) SetOn(name string, s *StyleDict) {
// 	s.BackgroundPosition = v
// }
// func (v BackgroundSize) SetOn(name string, s *StyleDict) {
// 	s.BackgroundSize = v
// }
// func (v Content) SetOn(name string, s *StyleDict) {
// 	s.Content = v
// }
// func (v Transforms) SetOn(name string, s *StyleDict) {
// 	s.Transforms = v
// }
// func (v Quotes) SetOn(name string, s *StyleDict) {
// 	s.Quotes = v
// }
// func (v ListStyleImage) SetOn(name string, s *StyleDict) {
// 	s.ListStyleImage = v
// }
// func (v StringSet) SetOn(name string, s *StyleDict) {
// 	s.StringSet = v
// }
// func (v StringContent) SetOn(name string, s *StyleDict) {
// 	s.BookmarkLabel = v
// }
