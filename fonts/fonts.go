package fonts

//FIXME: à implémenter

type FontConfiguration struct{}

func NewFontConfiguration() *FontConfiguration {
	return &FontConfiguration{}
}

func (f *FontConfiguration) AddFontFace(ruleDescriptors, urlFetcher interface{}) string { return "" }
