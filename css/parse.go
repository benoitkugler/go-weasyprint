package css

// Link the css parser with Weasyprint representation of CSS properties

type MaybeInt struct {
	Valid bool
	Int   int
}

type tokenType uint8

type Token struct {
	Type tokenType
	Dimension
	String                string
	LowerValue, LowerName string
	IntValue              MaybeInt
	Arguments             []Token
}
