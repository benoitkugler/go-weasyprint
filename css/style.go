package css

const (
	Top    Side = "top"
	Bottom Side = "bottom"
	Left   Side = "left"
	Right  Side = "right"
)

type Dimension struct {
	Unit  string
	Value float64
}

type Side string

type StyleDict struct {
	Anonymous bool

	Float    string
	Position string
	Page     int

	Margin      map[Side]Dimension
	Padding     map[Side]Dimension
	BorderWidth map[Side]float64

	Direction string

	TextTransform, Hyphens string
	Display                string
}

// Deep copy
func (s StyleDict) Copy() StyleDict {
	news := s
	news.Margin = make(map[Side]Dimension)
	news.Padding = make(map[Side]Dimension)
	news.BorderWidth = make(map[Side]float64)
	for k, v := range s.Margin {
		news.Margin[k] = v
	}
	for k, v := range s.Padding {
		news.Padding[k] = v
	}
	for k, v := range s.BorderWidth {
		news.BorderWidth[k] = v
	}
	return news
}
