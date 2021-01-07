package fontconfig

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// import (
// 	"bytes"
// 	"encoding/xml"
// 	"fmt"
// 	"io"
// 	"strings"
// )

type FcElement uint8

const (
	FcElementNone FcElement = iota
	FcElementFontconfig
	FcElementDir
	FcElementCacheDir
	FcElementCache
	FcElementInclude
	FcElementConfig
	FcElementMatch
	FcElementAlias
	FcElementDescription
	FcElementRemapDir
	FcElementResetDirs

	FcElementRescan

	FcElementPrefer
	FcElementAccept
	FcElementDefault
	FcElementFamily

	FcElementSelectfont
	FcElementAcceptfont
	FcElementRejectfont
	FcElementGlob
	FcElementPattern
	FcElementPatelt

	FcElementTest
	FcElementEdit
	FcElementInt
	FcElementDouble
	FcElementString
	FcElementMatrix
	FcElementRange
	FcElementBool
	FcElementCharSet
	FcElementLangSet
	FcElementName
	FcElementConst
	FcElementOr
	FcElementAnd
	FcElementEq
	FcElementNotEq
	FcElementLess
	FcElementLessEq
	FcElementMore
	FcElementMoreEq
	FcElementContains
	FcElementNotContains
	FcElementPlus
	FcElementMinus
	FcElementTimes
	FcElementDivide
	FcElementNot
	FcElementIf
	FcElementFloor
	FcElementCeil
	FcElementRound
	FcElementTrunc
	FcElementUnknown
)

var fcElementMap = [...]string{
	FcElementFontconfig:  "fontconfig",
	FcElementDir:         "dir",
	FcElementCacheDir:    "cachedir",
	FcElementCache:       "cache",
	FcElementInclude:     "include",
	FcElementConfig:      "config",
	FcElementMatch:       "match",
	FcElementAlias:       "alias",
	FcElementDescription: "description",
	FcElementRemapDir:    "remap-dir",
	FcElementResetDirs:   "reset-dirs",

	FcElementRescan: "rescan",

	FcElementPrefer:  "prefer",
	FcElementAccept:  "accept",
	FcElementDefault: "default",
	FcElementFamily:  "family",

	FcElementSelectfont: "selectfont",
	FcElementAcceptfont: "acceptfont",
	FcElementRejectfont: "rejectfont",
	FcElementGlob:       "glob",
	FcElementPattern:    "pattern",
	FcElementPatelt:     "patelt",

	FcElementTest:        "test",
	FcElementEdit:        "edit",
	FcElementInt:         "int",
	FcElementDouble:      "double",
	FcElementString:      "string",
	FcElementMatrix:      "matrix",
	FcElementRange:       "range",
	FcElementBool:        "bool",
	FcElementCharSet:     "charset",
	FcElementLangSet:     "langset",
	FcElementName:        "name",
	FcElementConst:       "const",
	FcElementOr:          "or",
	FcElementAnd:         "and",
	FcElementEq:          "eq",
	FcElementNotEq:       "not_eq",
	FcElementLess:        "less",
	FcElementLessEq:      "less_eq",
	FcElementMore:        "more",
	FcElementMoreEq:      "more_eq",
	FcElementContains:    "contains",
	FcElementNotContains: "not_contains",
	FcElementPlus:        "plus",
	FcElementMinus:       "minus",
	FcElementTimes:       "times",
	FcElementDivide:      "divide",
	FcElementNot:         "not",
	FcElementIf:          "if",
	FcElementFloor:       "floor",
	FcElementCeil:        "ceil",
	FcElementRound:       "round",
	FcElementTrunc:       "trunc",
}

var fcElementIgnoreName = [...]string{
	"its:",
}

func FcElementMap(name string) FcElement {
	for i, elem := range fcElementMap {
		if name == elem {
			return FcElement(i)
		}
	}
	for _, ignoreName := range fcElementIgnoreName {
		if strings.HasPrefix(name, ignoreName) {
			return FcElementNone
		}
	}
	return FcElementUnknown
}

func (e FcElement) String() string {
	if int(e) >= len(fcElementMap) {
		return fmt.Sprintf("invalid element %d", e)
	}
	return fcElementMap[e]
}

type FcPStack struct {
	// struct _FcPStack   *prev;
	element FcElement
	attr    []xml.Attr
	str     *bytes.Buffer
	// attr_buf_static [16]byte
}

type FcVStackTag uint8

const (
	FcVStackNone FcVStackTag = iota

	FcVStackString
	FcVStackFamily
	FcVStackConstant
	FcVStackGlob
	FcVStackName
	FcVStackPattern

	FcVStackPrefer
	FcVStackAccept
	FcVStackDefault

	FcVStackInteger
	FcVStackDouble
	FcVStackMatrix
	FcVStackRange
	FcVStackBool
	FcVStackCharSet
	FcVStackLangSet

	FcVStackTest
	FcVStackExpr
	FcVStackEdit
)

type FcVStack struct {
	// struct _FcVStack	*prev;
	pstack *FcPStack /* related parse element */
	tag    FcVStackTag
	// union {
	// FcChar8		*string;

	// int		integer;
	// double		_double;
	// FcExprMatrix	*matrix;
	// FcRange		*range;
	// FcBool		bool_;
	// FcCharSet	*charset;
	// FcLangSet	*langset;
	// FcExprName	name;

	// FcTest		*test;
	// FcQual		qual;
	// FcOp		op;
	// FcExpr		*expr;
	// FcEdit		*edit;

	// FcPattern	*pattern;
	// }
	u interface{}
}

const (
	FcSevereInfo uint8 = iota
	FcSevereWarning
)

type FcConfigParse struct {
	logger io.Writer

	pstack  []FcPStack
	vstack  []FcVStack
	name    string
	config  *FcConfig
	ruleset *FcRuleSet
	// parser   XML_Parser
	scanOnly bool
}

func (parse FcConfigParse) message(severe uint8, format string, args ...interface{}) {
	s := "unknown"
	switch severe {
	case FcSevereInfo:
		s = "info"
	case FcSevereWarning:
		s = "warning"
	}

	if parse.name != "" {
		fmt.Fprintf(parse.logger, "fontconfig %s: \"%s\": ", s, parse.name)
	} else {
		fmt.Fprintf(parse.logger, "fontconfig %s: ", s)
	}
	fmt.Fprintf(parse.logger, format, args...)
}

// do not check the stack is not empty
func (parse *FcConfigParse) p() *FcPStack {
	return &parse.pstack[len(parse.pstack)-1]
}

func (parse *FcConfigParse) FcStartElement(name string, attr []xml.Attr) {
	element := FcElementMap(name)

	if element == FcElementUnknown {
		parse.message(FcSevereWarning, "unknown element %s", name)
	}

	parse.pstackPush(element, attr)
}

// push at the end of the slice
func (parse *FcConfigParse) pstackPush(element FcElement, attr []xml.Attr) {
	new := FcPStack{
		element: element,
		attr:    attr,
		str:     &bytes.Buffer{},
	}
	parse.pstack = append(parse.pstack, new)
}

func (parse *FcConfigParse) pstackPop() {
	// the encoding/xml package makes sur tag are matching
	// so parse.pstack has at least one element

	// Don't check the attributes for FcElementNone
	if last := parse.p(); last.element != FcElementNone {
		// Warn about unused attrs.
		for _, attr := range last.attr {
			parse.message(FcSevereWarning, "invalid attribute %s", attr.Name.Local)
		}
	}

	parse.vstack = nil
	parse.pstack = parse.pstack[:len(parse.pstack)-1]
}

func (parse *FcConfigParse) FcEndElement() error {
	if len(parse.pstack) == 0 { // nothing to do
		return nil
	}
	var err error
	last := parse.p()
	switch last.element {
	case FcElementDir:
		err = parse.FcParseDir()
		// 	case FcElementCacheDir:
		// 		FcParseCacheDir(parse)
		// 	case FcElementCache:
		// 		data = FcStrBufDoneStatic(&parse.pstack.str)
		// 		// discard this data; no longer used
		// 		FcStrBufDestroy(&parse.pstack.str)
		// 	case FcElementInclude:
		// 		FcParseInclude(parse)
		// 	case FcElementMatch:
		// 		FcParseMatch(parse)
		// 	case FcElementAlias:
		// 		FcParseAlias(parse)
		// 	case FcElementDescription:
		// 		FcParseDescription(parse)
		// 	case FcElementRemapDir:
		// 		FcParseRemapDir(parse)
		// 	case FcElementResetDirs:
		// 		FcParseResetDirs(parse)

		// 	case FcElementRescan:
		// 		FcParseRescan(parse)

		// 	case FcElementPrefer:
		// 		FcParseFamilies(parse, FcVStackPrefer)
		// 	case FcElementAccept:
		// 		FcParseFamilies(parse, FcVStackAccept)
		// 	case FcElementDefault:
		// 		FcParseFamilies(parse, FcVStackDefault)
		// 	case FcElementFamily:
		// 		FcParseFamily(parse)

		// 	case FcElementTest:
		// 		FcParseTest(parse)
		// 	case FcElementEdit:
		// 		FcParseEdit(parse)

		// 	case FcElementInt:
		// 		FcParseInt(parse)
		// 	case FcElementDouble:
		// 		FcParseDouble(parse)
		// 	case FcElementString:
		// 		FcParseString(parse, FcVStackString)
		// 	case FcElementMatrix:
		// 		FcParseMatrix(parse)
		// 	case FcElementRange:
		// 		FcParseRange(parse)
		// 	case FcElementBool:
		// 		FcParseBool(parse)
		// 	case FcElementCharSet:
		// 		FcParseCharSet(parse)
		// 	case FcElementLangSet:
		// 		FcParseLangSet(parse)
		// 	case FcElementSelectfont:
		// 	case FcElementAcceptfont, FcElementRejectfont:
		// 		FcParseAcceptRejectFont(parse, parse.pstack.element)
		// 	case FcElementGlob:
		// 		FcParseString(parse, FcVStackGlob)
		// 	case FcElementPattern:
		// 		FcParsePattern(parse)
		// 	case FcElementPatelt:
		// 		FcParsePatelt(parse)
		// 	case FcElementName:
		// 		FcParseName(parse)
		// 	case FcElementConst:
		// 		FcParseString(parse, FcVStackConstant)
		// 	case FcElementOr:
		// 		FcParseBinary(parse, FcOpOr)
		// 	case FcElementAnd:
		// 		FcParseBinary(parse, FcOpAnd)
		// 	case FcElementEq:
		// 		FcParseBinary(parse, FcOpEqual)
		// 	case FcElementNotEq:
		// 		FcParseBinary(parse, FcOpNotEqual)
		// 	case FcElementLess:
		// 		FcParseBinary(parse, FcOpLess)
		// 	case FcElementLessEq:
		// 		FcParseBinary(parse, FcOpLessEqual)
		// 	case FcElementMore:
		// 		FcParseBinary(parse, FcOpMore)
		// 	case FcElementMoreEq:
		// 		FcParseBinary(parse, FcOpMoreEqual)
		// 	case FcElementContains:
		// 		FcParseBinary(parse, FcOpContains)
		// 	case FcElementNotContains:
		// 		FcParseBinary(parse, FcOpNotContains)
		// 	case FcElementPlus:
		// 		FcParseBinary(parse, FcOpPlus)
		// 	case FcElementMinus:
		// 		FcParseBinary(parse, FcOpMinus)
		// 	case FcElementTimes:
		// 		FcParseBinary(parse, FcOpTimes)
		// 	case FcElementDivide:
		// 		FcParseBinary(parse, FcOpDivide)
		// 	case FcElementNot:
		// 		FcParseUnary(parse, FcOpNot)
		// 	case FcElementIf:
		// 		FcParseBinary(parse, FcOpQuest)
		// 	case FcElementFloor:
		// 		FcParseUnary(parse, FcOpFloor)
		// 	case FcElementCeil:
		// 		FcParseUnary(parse, FcOpCeil)
		// 	case FcElementRound:
		// 		FcParseUnary(parse, FcOpRound)
		// 	case FcElementTrunc:
		// 		FcParseUnary(parse, FcOpTrunc)
	}
	parse.pstackPop()
	return err
}

func (parse *FcConfigParse) getAttr(attr string) string {
	if len(parse.pstack) == 0 {
		return ""
	}

	attrs := parse.p().attr

	for i, attrXml := range attrs {
		if attr == attrXml.Name.Local {
			attrs[i].Name.Local = "" // Mark as used.
			return attrXml.Value
		}
	}
	return ""
}

func (parse *FcConfigParse) FcParseDir() error {
	data := parse.p().str
	if data.Len() == 0 {
		parse.message(FcSevereWarning, "empty font directory name ignored")
		return nil
	}
	attr := parse.getAttr("prefix")
	salt := parse.getAttr("salt")
	prefix, err := parse.getRealPathFromPrefix(data.String(), attr)
	if err != nil {
		return err
	}
	if prefix == "" {
		// nop
	} else if !parse.scanOnly && (!usesHome(prefix) || FcConfigHome() != "") {
		if err := parse.config.addFontDir(prefix, "", salt); err != nil {
			return fmt.Errorf("fontconfig: cannot add directory %s: %s", prefix, err)
		}
	}
	parse.p().str.Reset()
	return nil
}

func usesHome(str string) bool { return strings.HasPrefix(str, "~") }

func xdgDataHome() string {
	env := os.Getenv("XDG_DATA_HOME")
	if !homeEnabled {
		return ""
	}
	if env != "" {
		return env
	}
	home := FcConfigHome()
	return filepath.Join(home, ".local", "share")
}

func (parse *FcConfigParse) getRealPathFromPrefix(path, prefix string) (string, error) {
	var parent string
	switch prefix {
	case "xdg":
		parent := xdgDataHome()
		if parent == "" { // Home directory might be disabled
			return "", nil
		}
	case "default", "cwd":
		// Nothing to do
	case "relative":
		parent = filepath.Dir(parse.name)
		if parent == "." {
			return "", nil
		}

	// #ifndef _WIN32
	// /* For Win32, check this later for dealing with special cases */
	default:
		if !filepath.IsAbs(path) && path[0] != '~' {
			parse.message(FcSevereWarning,
				"Use of ambiguous path in <%s> element. please add prefix=\"cwd\" if current behavior is desired.",
				parse.p().element)
		}
		// #else
	}

	// TODO: support windows
	//     if  path ==  "CUSTOMFONTDIR"  {
	// 	// FcChar8 *p;
	// 	// path = buffer;
	// 	if (!GetModuleFileName (nil, (LPCH) buffer, sizeof (buffer) - 20)) 	{
	// 	    parse.message ( FcSevereError, "GetModuleFileName failed");
	// 	    return ""
	// 	}
	// 	/*
	// 	 * Must use the multi-byte aware function to search
	// 	 * for backslash because East Asian double-byte code
	// 	 * pages have characters with backslash as the second
	// 	 * byte.
	// 	 */
	// 	p = _mbsrchr (path, '\\');
	// 	if (p) *p = '\0';
	// 	strcat ((char *) path, "\\fonts");
	//     }
	//     else if (strcmp ((const char *) path, "APPSHAREFONTDIR") == 0)
	//     {
	// 	FcChar8 *p;
	// 	path = buffer;
	// 	if (!GetModuleFileName (nil, (LPCH) buffer, sizeof (buffer) - 20))
	// 	{
	// 	    parse.message ( FcSevereError, "GetModuleFileName failed");
	// 	    return nil;
	// 	}
	// 	p = _mbsrchr (path, '\\');
	// 	if (p) *p = '\0';
	// 	strcat ((char *) path, "\\..\\share\\fonts");
	//     }
	//     else if (strcmp ((const char *) path, "WINDOWSFONTDIR") == 0)
	//     {
	// 	int rc;
	// 	path = buffer;
	// 	rc = pGetSystemWindowsDirectory ((LPSTR) buffer, sizeof (buffer) - 20);
	// 	if (rc == 0 || rc > sizeof (buffer) - 20)
	// 	{
	// 	    parse.message ( FcSevereError, "GetSystemWindowsDirectory failed");
	// 	    return nil;
	// 	}
	// 	if (path [strlen ((const char *) path) - 1] != '\\')
	// 	    strcat ((char *) path, "\\");
	// 	strcat ((char *) path, "fonts");
	//     }
	//     else
	//     {
	// 	if (!prefix)
	// 	{
	// 	    if (!FcStrIsAbsoluteFilename (path) && path[0] != '~')
	// 		parse.message ( FcSevereWarning, "Use of ambiguous path in <%s> element. please add prefix=\"cwd\" if current behavior is desired.", FcElementReverseMap (parse.pstack.element));
	// 	}
	//     }
	// #endif

	if parent != "" {
		return filepath.Join(parent, path), nil
	}
	return path, nil
}
