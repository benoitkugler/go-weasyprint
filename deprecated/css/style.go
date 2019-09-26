package css

//// IsZero returns `true` if the StyleFor is not initialized.
//// Thus, we can use a zero StyleFor as null value.
//func (s StyleFor) IsZero() bool {
//	return s.Properties == nil
//}
//
//// Deep copy.
//// inheritedStyle is a shallow copy
//func (s StyleFor) Copy() StyleFor {
//	out := s
//	out.Properties = s.Properties.Copy()
//	return out
//}
//
//// InheritFrom returns a new StyleFor with inherited properties from this one.
//// Non-inherited properties get their initial values.
//// This is the method used for an anonymous box.
//func (s *StyleFor) InheritFrom() StyleFor {
//	if s.inheritedStyle == nil {
//		is := computedFromCascaded(&utils.HTMLNode{}, nil, *s, StyleFor{}, "")
//		is.Anonymous = true
//		s.inheritedStyle = &is
//	}
//	return *s.inheritedStyle
//}
//
//func (s StyleFor) ResolveColor(key string) pr.Color {
//	value := s.Properties[key].(pr.Color)
//	if value.Type == parser.ColorCurrentColor {
//		value = s.GetColor()
//	}
//	return value
//}

//func matchingPageTypes(pageType utils.PageElement, _names map[pr.Page]struct{}) (out []utils.PageElement) {
//	sides := []string{"left", "right", ""}
//	if pageType.Side != "" {
//		sides = []string{pageType.Side}
//	}
//
//	blanks := []bool{true}
//	if pageType.Blank == false {
//		blanks = []bool{true, false}
//	}
//	firsts := []bool{true}
//	if pageType.First == false {
//		firsts = []bool{true, false}
//	}
//	names := []string{pageType.Name}
//	if pageType.Name == "" {
//		names = []string{""}
//		for page := range _names {
//			names = append(names, page.String)
//		}
//	}
//	for _, side := range sides {
//		for _, blank := range blanks {
//			for _, first := range firsts {
//				for _, name := range names {
//					out = append(out, utils.PageElement{Side: side, Blank: blank, First: first, Name: name})
//				}
//			}
//		}
//	}
//	return
//}
