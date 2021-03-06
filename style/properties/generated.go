package properties

// autogenerated from initial_values.go

func (BoolString) isCssProperty() {}
func (v BoolString) IsNone() bool {
	return v == BoolString{}
}

func (Center) isCssProperty() {}
func (v Center) IsNone() bool {
	return v == Center{}
}

func (Centers) isCssProperty() {}

func (Color) isCssProperty() {}

func (ContentProperties) isCssProperty() {}

func (Float) isCssProperty() {}

func (Images) isCssProperty() {}

func (Int) isCssProperty() {}

func (IntString) isCssProperty() {}
func (v IntString) IsNone() bool {
	return v == IntString{}
}

func (IntStrings) isCssProperty() {}

func (Ints3) isCssProperty() {}
func (v Ints3) IsNone() bool {
	return v == Ints3{}
}

func (Marks) isCssProperty() {}
func (v Marks) IsNone() bool {
	return v == Marks{}
}

func (NDecorations) isCssProperty() {}
func (v NDecorations) IsNone() bool {
	return v.None == false
}

func (NamedString) isCssProperty() {}
func (v NamedString) IsNone() bool {
	return v == NamedString{}
}

func (Page) isCssProperty() {}
func (v Page) IsNone() bool {
	return v == Page{}
}

func (Point) isCssProperty() {}
func (v Point) IsNone() bool {
	return v == Point{}
}

func (Quotes) isCssProperty() {}
func (v Quotes) IsNone() bool {
	return v.Open == nil && v.Close == nil
}

func (Repeats) isCssProperty() {}

func (SContent) isCssProperty() {}
func (v SContent) IsNone() bool {
	return v.String == "" && v.Contents == nil
}

func (SIntStrings) isCssProperty() {}
func (v SIntStrings) IsNone() bool {
	return v.String == "" && v.Values == nil
}

func (SStrings) isCssProperty() {}
func (v SStrings) IsNone() bool {
	return v.String == "" && v.Strings == nil
}

func (Sizes) isCssProperty() {}

func (String) isCssProperty() {}

func (StringSet) isCssProperty() {}
func (v StringSet) IsNone() bool {
	return v.String == "" && v.Contents == nil
}

func (Strings) isCssProperty() {}

func (Transforms) isCssProperty() {}

func (Value) isCssProperty() {}
func (v Value) IsNone() bool {
	return v == Value{}
}

func (Values) isCssProperty() {}

func (AttrData) isCssProperty() {}
func (v AttrData) IsNone() bool {
	return v.Name == "" && v.TypeOrUnit == ""
}
func (v SContentProp) IsNone() bool {
	return v.String == ""
}
func (v LinearGradient) IsNone() bool {
	return v.ColorStops == nil && v.Direction == DirectionType{} && v.Repeating == false
}
func (v ContentProperty) IsNone() bool {
	return v.Type == ""
}
func (v ColorStop) IsNone() bool {
	return v == ColorStop{}
}
func (v NamedProperty) IsNone() bool {
	return v.Name == ""
}
func (v GradientSize) IsNone() bool {
	return v == GradientSize{}
}
func (v RadialGradient) IsNone() bool {
	return v.ColorStops == nil && v.Shape == "" && v.Size == GradientSize{} && v.Center == Center{} && v.Repeating == false
}
func (v DirectionType) IsNone() bool {
	return v == DirectionType{}
}
func (v SDimensions) IsNone() bool {
	return v.String == "" && v.Dimensions == nil
}
func (v Dimension) IsNone() bool {
	return v == Dimension{}
}
func (v Quote) IsNone() bool {
	return v == Quote{}
}
func (v Size) IsNone() bool {
	return v == Size{}
}
