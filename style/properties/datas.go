package properties

import "github.com/benoitkugler/go-weasyprint/style/parser"

const ( // zero field corresponds to null content
	Scalar Unit = iota + 1 // means no unit, but a valid value
	Percentage
	Ex
	Em
	Ch
	Rem
	Px
	Pt
	Pc
	In
	Cm
	Mm
	Q

	Rad
	Turn
	Deg
	Grad
)

var (
	ZeroPixels      = Dimension{Unit: Px}
	zeroPixelsValue = ZeroPixels.ToValue()

	CurrentColor = Color{Type: parser.ColorCurrentColor}
	// How many CSS pixels is one <unit>?
	// http://www.w3.org/TR/CSS21/syndata.html#length-units
	LengthsToPixels = map[Unit]float32{
		Px: 1,
		Pt: 1. / 0.75,
		Pc: 16.,             // LengthsToPixels["pt"] * 12
		In: 96.,             // LengthsToPixels["pt"] * 72
		Cm: 96. / 2.54,      // LengthsToPixels["in"] / 2.54
		Mm: 96. / 25.4,      // LengthsToPixels["in"] / 25.4
		Q:  96. / 25.4 / 4., // LengthsToPixels["mm"] / 4
	}

	// Value in pixels of font-size for <absolute-size> keywords: 12pt (16px) for
	// medium, and scaling factors given in CSS3 for others:
	// http://www.w3.org/TR/css3-fonts/#font-size-prop
	// TODO: this will need to be ordered to implement 'smaller' and 'larger'
	FontSizeKeywords = map[string]float32{ // medium is 16px, others are a ratio of medium
		"xx-small": InitialValues.GetFontSize().Value * 3 / 5,
		"x-small":  InitialValues.GetFontSize().Value * 3 / 4,
		"small":    InitialValues.GetFontSize().Value * 8 / 9,
		"medium":   InitialValues.GetFontSize().Value * 1 / 1,
		"large":    InitialValues.GetFontSize().Value * 6 / 5,
		"x-large":  InitialValues.GetFontSize().Value * 3 / 2,
		"xx-large": InitialValues.GetFontSize().Value * 2 / 1,
	}

	// http://www.w3.org/TR/css3-page/#size
	// name=(width in pixels, height in pixels)
	PageSizes = map[string]Point{
		"a5":     {Dimension{Value: 148, Unit: Mm}, Dimension{Value: 210, Unit: Mm}},
		"a4":     InitialWidthHeight,
		"a3":     {Dimension{Value: 297, Unit: Mm}, Dimension{Value: 420, Unit: Mm}},
		"b5":     {Dimension{Value: 176, Unit: Mm}, Dimension{Value: 250, Unit: Mm}},
		"b4":     {Dimension{Value: 250, Unit: Mm}, Dimension{Value: 353, Unit: Mm}},
		"letter": {Dimension{Value: 8.5, Unit: In}, Dimension{Value: 11, Unit: In}},
		"legal":  {Dimension{Value: 8.5, Unit: In}, Dimension{Value: 14, Unit: In}},
		"ledger": {Dimension{Value: 11, Unit: In}, Dimension{Value: 17, Unit: In}},
	}

	InitialWidthHeight = Point{Dimension{Value: 210, Unit: Mm}, Dimension{Value: 297, Unit: Mm}}
)
