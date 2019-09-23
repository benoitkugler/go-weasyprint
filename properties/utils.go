package properties

var Has = struct{}{}

type Set map[string]struct{}

func (s Set) Add(key string) {
	s[key] = Has
}

func (s Set) Has(key string) bool {
	_, in := s[key]
	return in
}

func NewSet(values ...string) Set {
	s := make(Set, len(values))
	for _, v := range values {
		s.Add(v)
	}
	return s
}
