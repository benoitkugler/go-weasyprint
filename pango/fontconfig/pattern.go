package fontconfig

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"unicode"
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

// NewFcPattern returns an empty, initalized pattern
func NewFcPattern() *FcPattern {
	return &FcPattern{elts: make(map[FcObject]FcValueList)}
}

// Duplicate returns a new pattern that matches
// `p`. Each pattern may be modified without affecting the other.
func (p *FcPattern) Duplicate() *FcPattern {
	if p == nil {
		return nil
	}
	out := FcPattern{elts: make(map[FcObject]FcValueList, len(p.elts))}
	for o, l := range p.elts {
		out.elts[o] = l.duplicate()
	}
	return &out
}

// Add adds the given value for the given object, with a strong binding.
// `appendMode` controls the location of insertion in the current list.
func (p FcPattern) Add(object FcObject, value interface{}, appendMode bool) {
	p.addWithBinding(object, value, FcValueBindingStrong, appendMode)
}

func (p FcPattern) addWithBinding(object FcObject, value interface{}, binding FcValueBinding, appendMode bool) {
	newV := valueElt{value: value, binding: binding}
	p.AddList(object, FcValueList{newV}, appendMode)
}

func (p FcPattern) FcPatternObjectAddBool(object FcObject, value bool)   { p.Add(object, value, true) }
func (p FcPattern) FcPatternObjectAddInteger(object FcObject, value int) { p.Add(object, value, true) }
func (p FcPattern) FcPatternObjectAddDouble(object FcObject, value float64) {
	p.Add(object, value, true)
}
func (p FcPattern) FcPatternObjectAddString(object FcObject, value string) {
	p.Add(object, value, true)
}

// Add adds the given list of values for the given object.
// `appendMode` controls the location of insertion in the current list.
func (p FcPattern) AddList(object FcObject, list FcValueList, appendMode bool) {
	//  FcPatternObjectAddWithBinding(p, FcObjectFromName(object), value, FcValueBindingStrong, append)
	// object := FcObject(objectS)

	// Make sure the stored type is valid for built-in objects
	for _, value := range list {
		if !object.hasValidType(value.value) {
			log.Printf("fontconfig: pattern object %s does not accept value %v", object, value.value)
			return
		}
	}

	e := p.elts[object]
	if appendMode {
		e = append(e, list...)
	} else {
		e = e.prepend(list...)
	}
	p.elts[object] = e
}

func (p FcPattern) del(obj FcObject) { delete(p.elts, obj) }

func (p FcPattern) sortedKeys() []FcObject {
	keys := make([]FcObject, 0, len(p.elts))
	for r := range p.elts {
		keys = append(keys, r)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

// Hash returns a value, usable as map key, and
// defining the pattern in terms of equality:
// two patterns with the same hash are considered equal.
func (p FcPattern) Hash() string {
	var hash []byte
	for _, object := range p.sortedKeys() {
		v := p.elts[object]
		hash = append(append(hash, byte(object), ':'), v.Hash()...)
	}
	return string(hash)
}

// String returns a human friendly representation,
// mainly used for debugging.
func (p *FcPattern) String() string {
	if p == nil {
		return "Nil pattern"
	}
	s := fmt.Sprintf("%d elements pattern:\n", len(p.elts))

	for obj, vs := range p.elts {
		s += fmt.Sprintf("\t%s: %v\n", obj, vs)
	}
	return s
}

func (p FcPattern) FcPatternObjectGet(object FcObject, id int) (FcValue, FcResult) {
	e := p.elts[object]
	if e == nil {
		return nil, FcResultNoMatch
	}
	if id >= len(e) {
		return nil, FcResultNoId
	}
	return e[id].value, FcResultMatch
}

func (p FcPattern) FcPatternObjectGetBool(object FcObject, id int) (FcBool, FcResult) {
	v, r := p.FcPatternObjectGet(object, id)
	if r != FcResultMatch {
		return 0, r
	}
	out, ok := v.(FcBool)
	if !ok {
		return 0, FcResultTypeMismatch
	}
	return out, FcResultMatch
}

func (p FcPattern) FcPatternObjectGetString(object FcObject, id int) (string, FcResult) {
	v, r := p.FcPatternObjectGet(object, id)
	if r != FcResultMatch {
		return "", r
	}
	out, ok := v.(string)
	if !ok {
		return "", FcResultTypeMismatch
	}
	return out, FcResultMatch
}

func (p FcPattern) FcPatternObjectGetCharSet(object FcObject, id int) (FcCharset, FcResult) {
	v, r := p.FcPatternObjectGet(object, id)
	if r != FcResultMatch {
		return FcCharset{}, r
	}
	out, ok := v.(FcCharset)
	if !ok {
		return FcCharset{}, FcResultTypeMismatch
	}
	return out, FcResultMatch
}

// Add all of the elements in 's' to 'p'
func (p *FcPattern) append(s *FcPattern) {
	for object, list := range s.elts {
		for _, v := range list {
			p.addWithBinding(object, v.value, v.binding, true)
		}
	}
}

func (pat *FcPattern) addFullname() bool {
	b, res := pat.FcPatternObjectGetBool(FC_VARIABLE, 0)
	if res == FcResultMatch && b != FcFalse {
		return true
	}

	var (
		lang, style string
		n           int
	)
	lang, res = pat.FcPatternObjectGetString(FC_FAMILYLANG, n)
	for ; res == FcResultMatch; lang, res = pat.FcPatternObjectGetString(FC_FAMILYLANG, n) {
		if lang == "en" {
			break
		}
		n++
		lang = ""
	}
	if lang == "" {
		n = 0
	}
	family, res := pat.FcPatternObjectGetString(FC_FAMILY, n)
	if res != FcResultMatch {
		return false
	}
	family = strings.TrimRightFunc(family, unicode.IsSpace)
	lang = ""
	lang, res = pat.FcPatternObjectGetString(FC_STYLELANG, n)
	for ; res == FcResultMatch; lang, res = pat.FcPatternObjectGetString(FC_STYLELANG, n) {
		if lang == "en" {
			break
		}
		n++
		lang = ""
	}
	if lang == "" {
		n = 0
	}
	style, res = pat.FcPatternObjectGetString(FC_STYLE, n)
	if res != FcResultMatch {
		return false
	}

	style = strings.TrimLeftFunc(style, unicode.IsSpace)
	sbuf := []byte(family)
	if FcStrCmpIgnoreBlanksAndCase(style, "Regular") != 0 {
		sbuf = append(sbuf, ' ')
		sbuf = append(sbuf, style...)
	}
	pat.del(FC_FULLNAME)
	pat.Add(FC_FULLNAME, string(sbuf), true)
	pat.del(FC_FULLNAMELANG)
	pat.Add(FC_FULLNAMELANG, "en", true)

	return true
}

type PatternElement struct {
	Object FcObject
	Value  FcValue
}

// TODO: check the pointer types in values
func FcPatternBuild(elements ...PatternElement) *FcPattern {
	p := FcPattern{elts: make(map[FcObject]FcValueList, len(elements))}
	for _, el := range elements {
		p.Add(el.Object, el.Value, true)
	}
	return &p
}

func (p *FcPattern) FcConfigPatternAdd(object FcObject, list FcValueList, append bool, table *FamilyTable) {
	e := p.elts[object]
	e.insert(-1, append, list, object, table)
	p.elts[object] = e
}

// Delete all values associated with a field
func (p *FcPattern) FcConfigPatternDel(object FcObject, table *FamilyTable) {
	e := p.elts[object]

	if object == FC_FAMILY && table != nil {
		for _, v := range e {
			table.del(v.value.(string))
		}
	}

	delete(p.elts, object)
}

// remove the empty lists
func (p *FcPattern) canon(object FcObject) {
	e := p.elts[object]
	if len(e) == 0 {
		delete(p.elts, object)
	}
}
