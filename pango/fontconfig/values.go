package fontconfig

import (
	"fmt"
)

// the C implementation use a refined string->int lookup function
type FcObject uint8

// The order is part of the cache signature.
const (
	FC_INVALID         FcObject = iota
	FC_FAMILY                   // String
	FC_FAMILYLANG               // String
	FC_STYLE                    // String
	FC_STYLELANG                // String
	FC_FULLNAME                 // String
	FC_FULLNAMELANG             // String
	FC_SLANT                    // Integer
	FC_WEIGHT                   // Range
	FC_WIDTH                    // Range
	FC_SIZE                     // Range
	FC_ASPECT                   // Double
	FC_PIXEL_SIZE               // Double
	FC_SPACING                  // Integer
	FC_FOUNDRY                  // String
	FC_ANTIALIAS                // Bool
	FC_HINT_STYLE               // Integer
	FC_HINTING                  // Bool
	FC_VERTICAL_LAYOUT          // Bool
	FC_AUTOHINT                 // Bool
	FC_GLOBAL_ADVANCE           // Bool
	FC_FILE                     // String
	FC_INDEX                    // Integer
	FC_RASTERIZER               // String
	FC_OUTLINE                  // Bool
	FC_SCALABLE                 // Bool
	FC_DPI                      // Double
	FC_RGBA                     // Integer
	FC_SCALE                    // Double
	FC_MINSPACE                 // Bool
	FC_CHARWIDTH                // Integer
	FC_CHAR_HEIGHT              // Integer
	FC_MATRIX                   // Matrix
	FC_CHARSET                  // CharSet
	FC_LANG                     // LangSet
	FC_FONTVERSION              // Integer
	FC_CAPABILITY               // String
	FC_FONTFORMAT               // String
	FC_EMBOLDEN                 // Bool
	FC_EMBEDDED_BITMAP          // Bool
	FC_DECORATIVE               // Bool
	FC_LCD_FILTER               // Integer
	FC_NAMELANG                 // String
	FC_FONT_FEATURES            // String
	FC_PRGNAME                  // String
	FC_HASH                     // String
	FC_POSTSCRIPT_NAME          // String
	FC_COLOR                    // Bool
	FC_SYMBOL                   // Bool
	FC_FONT_VARIATIONS          // String
	FC_VARIABLE                 // Bool
	FC_FONT_HAS_HINT            // Bool
	FC_ORDER                    // Integer
)

type FcBool uint8

const (
	FcFalse FcBool = iota
	FcTrue
	FcDontCare
)

type FcRange struct {
	Begin, End float64
}

// Hasher mey be implemented by complex value types,
// for which a custom hash is needed.
// Other type use their string representation.
type Hasher interface {
	Hash() []byte
}

type valueElt struct {
	value   interface{}
	binding FcValueBinding
}

func (v valueElt) hash() []byte {
	if withHash, ok := v.value.(Hasher); ok {
		return withHash.Hash()
	}
	return []byte(fmt.Sprintf("%v", v.value))
}

type FcValueBinding uint8

const (
	FcValueBindingWeak FcValueBinding = iota
	FcValueBindingStrong
	FcValueBindingSame
)

type FcValueList []valueElt

func (vs FcValueList) Hash() []byte {
	var hash []byte
	for _, v := range vs {
		hash = append(hash, v.hash()...)
	}
	return hash
}

func (l FcValueList) prepend(v ...valueElt) FcValueList {
	l = append(l, make(FcValueList, len(v))...)
	copy(l[len(v):], l)
	copy(l, v)
	return l
}

// returns a deep copy
func (l FcValueList) duplicate() FcValueList {
	// TODO: check the pointer types
	return append(FcValueList(nil), l...)
}
