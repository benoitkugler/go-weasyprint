package fontconfig

import (
	"fmt"
)

// the C implementation usee a refined string->int lookup function
type FcObject string

const (
	FC_INVALID         = ""
	FC_FAMILY          = "family"         /* String */
	FC_STYLE           = "style"          /* String */
	FC_SLANT           = "slant"          /* Int */
	FC_WEIGHT          = "weight"         /* Int */
	FC_SIZE            = "size"           /* Range (double) */
	FC_ASPECT          = "aspect"         /* Double */
	FC_PIXEL_SIZE      = "pixelsize"      /* Double */
	FC_SPACING         = "spacing"        /* Int */
	FC_FOUNDRY         = "foundry"        /* String */
	FC_ANTIALIAS       = "antialias"      /* Bool (depends) */
	FC_HINTING         = "hinting"        /* Bool (true) */
	FC_HINT_STYLE      = "hintstyle"      /* Int */
	FC_VERTICAL_LAYOUT = "verticallayout" /* Bool (false) */
	FC_AUTOHINT        = "autohint"       /* Bool (false) */
	/* FC_GLOBAL_ADVANCE is deprecated. this is simply ignored on freetype 2.4.5 or later */
	FC_GLOBAL_ADVANCE  = "globaladvance"  /* Bool (true) */
	FC_WIDTH           = "width"          /* Int */
	FC_FILE            = "file"           /* String */
	FC_INDEX           = "index"          /* Int */
	FC_FT_FACE         = "ftface"         /* FT_Face */
	FC_RASTERIZER      = "rasterizer"     /* String (deprecated) */
	FC_OUTLINE         = "outline"        /* Bool */
	FC_SCALABLE        = "scalable"       /* Bool */
	FC_COLOR           = "color"          /* Bool */
	FC_VARIABLE        = "variable"       /* Bool */
	FC_SCALE           = "scale"          /* double (deprecated) */
	FC_SYMBOL          = "symbol"         /* Bool */
	FC_DPI             = "dpi"            /* double */
	FC_RGBA            = "rgba"           /* Int */
	FC_MINSPACE        = "minspace"       /* Bool use minimum line spacing */
	FC_SOURCE          = "source"         /* String (deprecated) */
	FC_CHARSET         = "charset"        /* CharSet */
	FC_LANG            = "lang"           /* LangSet Set of RFC 3066 langs */
	FC_FONTVERSION     = "fontversion"    /* Int from 'head' table */
	FC_FULLNAME        = "fullname"       /* String */
	FC_FAMILYLANG      = "familylang"     /* String RFC 3066 langs */
	FC_STYLELANG       = "stylelang"      /* String RFC 3066 langs */
	FC_FULLNAMELANG    = "fullnamelang"   /* String RFC 3066 langs */
	FC_CAPABILITY      = "capability"     /* String */
	FC_FONTFORMAT      = "fontformat"     /* String */
	FC_EMBOLDEN        = "embolden"       /* Bool - true if emboldening needed*/
	FC_EMBEDDED_BITMAP = "embeddedbitmap" /* Bool - true to enable embedded bitmaps */
	FC_DECORATIVE      = "decorative"     /* Bool - true if style is a decorative variant */
	FC_LCD_FILTER      = "lcdfilter"      /* Int */
	FC_FONT_FEATURES   = "fontfeatures"   /* String */
	FC_FONT_VARIATIONS = "fontvariations" /* String */
	FC_NAMELANG        = "namelang"       /* String RFC 3866 langs */
	FC_PRGNAME         = "prgname"        /* String */
	FC_HASH            = "hash"           /* String (deprecated) */
	FC_POSTSCRIPT_NAME = "postscriptname" /* String */
	FC_FONT_HAS_HINT   = "fonthashint"    /* Bool - true if font has hinting */
	FC_ORDER           = "order"          /* Integer */
	FC_CHARWIDTH       = "charwidth"      /* Int */
	FC_CHAR_HEIGHT     = "charheight"     /* Int */
	FC_MATRIX          = "matrix"         /* FcMatrix */
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
