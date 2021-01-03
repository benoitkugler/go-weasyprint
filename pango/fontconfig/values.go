package fontconfig

type FcObject string // the C implementation stores it as an int, and use a refined string->int lookup function

const (
	FC_FAMILY          FcObject = "family"         /* String */
	FC_STYLE           FcObject = "style"          /* String */
	FC_SLANT           FcObject = "slant"          /* Int */
	FC_WEIGHT          FcObject = "weight"         /* Int */
	FC_SIZE            FcObject = "size"           /* Range (double) */
	FC_ASPECT          FcObject = "aspect"         /* Double */
	FC_PIXEL_SIZE      FcObject = "pixelsize"      /* Double */
	FC_SPACING         FcObject = "spacing"        /* Int */
	FC_FOUNDRY         FcObject = "foundry"        /* String */
	FC_ANTIALIAS       FcObject = "antialias"      /* Bool (depends) */
	FC_HINTING         FcObject = "hinting"        /* Bool (true) */
	FC_HINT_STYLE      FcObject = "hintstyle"      /* Int */
	FC_VERTICAL_LAYOUT FcObject = "verticallayout" /* Bool (false) */
	FC_AUTOHINT        FcObject = "autohint"       /* Bool (false) */
	/* FC_GLOBAL_ADVANCE is deprecated. this is simply ignored on freetype 2.4.5 or later */
	FC_GLOBAL_ADVANCE  FcObject = "globaladvance"  /* Bool (true) */
	FC_WIDTH           FcObject = "width"          /* Int */
	FC_FILE            FcObject = "file"           /* String */
	FC_INDEX           FcObject = "index"          /* Int */
	FC_FT_FACE         FcObject = "ftface"         /* FT_Face */
	FC_RASTERIZER      FcObject = "rasterizer"     /* String (deprecated) */
	FC_OUTLINE         FcObject = "outline"        /* Bool */
	FC_SCALABLE        FcObject = "scalable"       /* Bool */
	FC_COLOR           FcObject = "color"          /* Bool */
	FC_VARIABLE        FcObject = "variable"       /* Bool */
	FC_SCALE           FcObject = "scale"          /* double (deprecated) */
	FC_SYMBOL          FcObject = "symbol"         /* Bool */
	FC_DPI             FcObject = "dpi"            /* double */
	FC_RGBA            FcObject = "rgba"           /* Int */
	FC_MINSPACE        FcObject = "minspace"       /* Bool use minimum line spacing */
	FC_SOURCE          FcObject = "source"         /* String (deprecated) */
	FC_CHARSET         FcObject = "charset"        /* CharSet */
	FC_LANG            FcObject = "lang"           /* LangSet Set of RFC 3066 langs */
	FC_FONTVERSION     FcObject = "fontversion"    /* Int from 'head' table */
	FC_FULLNAME        FcObject = "fullname"       /* String */
	FC_FAMILYLANG      FcObject = "familylang"     /* String RFC 3066 langs */
	FC_STYLELANG       FcObject = "stylelang"      /* String RFC 3066 langs */
	FC_FULLNAMELANG    FcObject = "fullnamelang"   /* String RFC 3066 langs */
	FC_CAPABILITY      FcObject = "capability"     /* String */
	FC_FONTFORMAT      FcObject = "fontformat"     /* String */
	FC_EMBOLDEN        FcObject = "embolden"       /* Bool - true if emboldening needed*/
	FC_EMBEDDED_BITMAP FcObject = "embeddedbitmap" /* Bool - true to enable embedded bitmaps */
	FC_DECORATIVE      FcObject = "decorative"     /* Bool - true if style is a decorative variant */
	FC_LCD_FILTER      FcObject = "lcdfilter"      /* Int */
	FC_FONT_FEATURES   FcObject = "fontfeatures"   /* String */
	FC_FONT_VARIATIONS FcObject = "fontvariations" /* String */
	FC_NAMELANG        FcObject = "namelang"       /* String RFC 3866 langs */
	FC_PRGNAME         FcObject = "prgname"        /* String */
	FC_HASH            FcObject = "hash"           /* String (deprecated) */
	FC_POSTSCRIPT_NAME FcObject = "postscriptname" /* String */
	FC_FONT_HAS_HINT   FcObject = "fonthashint"    /* Bool - true if font has hinting */
	FC_ORDER           FcObject = "order"          /* Integer */
)

type valueElt struct {
	value   interface{}
	binding FcValueBinding
}

type FcValueBinding uint8

const (
	FcValueBindingWeak FcValueBinding = iota
	FcValueBindingStrong
	FcValueBindingSame
)

type FcValueList []valueElt
