package fontconfig

import (
	"fmt"
	"sort"
)

// An FcPattern holds a set of names with associated value lists; each name refers to a
// property of a font. FcPatterns are used as inputs to the matching code as
// well as holding information about specific fonts. Each property can hold
// one or more values; conventionally all of the same type, although the
// interface doesn't demand that.
//
// We use a very simple implementation: the C code is more refined,
// using a sorted list of (FcObject,FcValueList) pairs.
type FcPattern struct {
	elts map[FcObject]FcValueList
}

// we only support append = true
func (p FcPattern) FcPatternAdd(object FcObject, value interface{}) {
	//  FcPatternObjectAddWithBinding(p, FcObjectFromName(object), value, FcValueBindingStrong, append)
	// object := FcObject(objectS)

	binding := FcValueBindingStrong

	// Make sure the stored type is valid for built-in objects
	// if (!FcObjectValidType (object, value.type)) {
	// fprintf (stderr,
	// 	 "Fontconfig warning: FcPattern object %s does not accept value",
	// 	 FcObjectName (object));
	// FcValuePrintFile (stderr, value);
	// fprintf (stderr, "\n");
	// goto bail1;
	// }

	newV := valueElt{value: value, binding: binding}

	e := p.elts[object]
	e = append(e, newV)
	p.elts[object] = e
}

// Hash returns a value, usable as map key, and
// defining the pattern in terms of equality:
// two patterns with the same hash are considered equal.
func (p FcPattern) Hash() string {
	keys := make([]string, 0, len(p.elts))
	for r := range p.elts {
		keys = append(keys, string(r))
	}
	sort.Strings(keys)

	var hash []byte
	for _, object := range keys {
		v := p.elts[FcObject(object)]
		hash = append(append(hash, object...), v.Hash()...)
	}
	return string(hash)
}

// String returns a human friendly representation,
// mainly used for debugging.
func (p *FcPattern) String() string {
	// TODO: check this

	if p == nil {
		return "Null pattern"
	}
	s := fmt.Sprintf("Pattern has %d elts\n", len(p.elts))

	for obj, vs := range p.elts {
		s += fmt.Sprintf("\t%s:", obj)
		s += fmt.Sprintln(vs)
	}
	return s
}

type PatternElement struct {
	Object FcObject
	Value  interface{}
}

func FcPatternBuild(elements ...PatternElement) *FcPattern {
	p := FcPattern{elts: make(map[FcObject]FcValueList, len(elements))}
	for _, el := range elements {
		p.FcPatternAdd(el.Object, el.Value)
	}
	return &p
}
