package fontconfig

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
