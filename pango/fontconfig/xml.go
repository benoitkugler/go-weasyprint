package fontconfig

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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

func parseConfig(config *FcConfig, name string, complain, load bool) error {
	// TODO:
	//     FcChar8	    *filename = nil, *realfilename = nil;
	//     int		    fd;
	//     int		    len;
	//     FcStrBuf	    sbuf;
	//     char            buf[BUFSIZ];
	//     FcBool	    ret = FcFalse, complain_again = complain;
	//     FcStrBuf	    reason;

	//     FcStrBufInit (&reason, nil, 0);
	// #ifdef _WIN32
	//     if (!pGetSystemWindowsDirectory)
	//     {
	//         HMODULE hk32 = GetModuleHandleA("kernel32.dll");
	//         if (!(pGetSystemWindowsDirectory = (pfnGetSystemWindowsDirectory) GetProcAddress(hk32, "GetSystemWindowsDirectoryA")))
	//             pGetSystemWindowsDirectory = (pfnGetSystemWindowsDirectory) GetWindowsDirectory;
	//     }
	//     if (!pSHGetFolderPathA)
	//     {
	//         HMODULE hSh = LoadLibraryA("shfolder.dll");
	//         /* the check is done later, because there is no provided fallback */
	//         if (hSh)
	//             pSHGetFolderPathA = (pfnSHGetFolderPathA) GetProcAddress(hSh, "SHGetFolderPathA");
	//     }
	// #endif

	//     filename = GetFilename (config, name);
	//     if (!filename)
	//     {
	// 	FcStrBufString (&reason, (FcChar8 *)"No such file: ");
	// 	FcStrBufString (&reason, name ? name : (FcChar8 *)"(null)");
	// 	goto bail0;
	//     }
	//     realfilename = FcConfigRealFilename (config, name);
	//     if (!realfilename)
	//     {
	// 	FcStrBufString (&reason, (FcChar8 *)"No such realfile: ");
	// 	FcStrBufString (&reason, name ? name : (FcChar8 *)"(null)");
	// 	goto bail0;
	//     }
	//     if (FcStrSetMember (config.availConfigFiles, realfilename))
	//     {
	//         FcStrFree (filename);
	// 	FcStrFree (realfilename);
	//         return true;
	//     }

	//     if (load)
	//     {
	// 	if (!FcStrSetAdd (config.configFiles, filename))
	// 	    goto bail0;
	//     }
	//     if (!FcStrSetAdd (config.availConfigFiles, realfilename))
	// 	goto bail0;

	//     if (isDir (realfilename))
	//     {
	// 	ret = FcConfigParseAndLoadDir (config, name, realfilename, complain, load);
	// 	FcStrFree (filename);
	// 	FcStrFree (realfilename);
	// 	return ret;
	//     }

	//     FcStrBufInit (&sbuf, nil, 0);

	//     fd = FcOpen ((char *) realfilename, O_RDONLY);
	//     if (fd == -1)
	//     {
	// 	FcStrBufString (&reason, (FcChar8 *)"Unable to open ");
	// 	FcStrBufString (&reason, realfilename);
	// 	goto bail1;
	//     }

	//     do {
	// 	len = read (fd, buf, BUFSIZ);
	// 	if (len < 0)
	// 	{
	// 	    int errno_ = errno;
	// 	    char ebuf[BUFSIZ+1];

	// #if HAVE_STRERROR_R
	// 	    strerror_r (errno_, ebuf, BUFSIZ);
	// #elif HAVE_STRERROR
	// 	    char *tmp = strerror (errno_);
	// 	    size_t len = strlen (tmp);
	// 	    memcpy (ebuf, tmp, FC_MIN (BUFSIZ, len));
	// 	    ebuf[FC_MIN (BUFSIZ, len)] = 0;
	// #else
	// 	    ebuf[0] = 0;
	// #endif
	// 	    message (0, FcSevereError, "failed reading config file: %s: %s (errno %d)", realfilename, ebuf, errno_);
	// 	    close (fd);
	// 	    goto bail1;
	// 	}
	// 	FcStrBufData (&sbuf, (const FcChar8 *)buf, len);
	//     } while (len != 0);
	//     close (fd);

	//     ret = FcConfigParseAndLoadFromMemoryInternal (config, filename, FcStrBufDoneStatic (&sbuf), complain, load);
	//     complain_again = FcFalse; /* no need to reclaim here */
	// bail1:
	//     FcStrBufDestroy (&sbuf);
	// bail0:
	//     if (filename)
	// 	FcStrFree (filename);
	//     if (realfilename)
	// 	FcStrFree (realfilename);
	//     if (!complain)
	// 	return true;
	//     if (!ret && complain_again)
	//     {
	// 	if (name)
	// 	    message (0, FcSevereError, "Cannot %s config file \"%s\": %s", load ? "load" : "scan", name, FcStrBufDoneStatic (&reason));
	// 	else
	// 	    message (0, FcSevereError, "Cannot %s default config file: %s", load ? "load" : "scan", FcStrBufDoneStatic (&reason));
	// 	FcStrBufDestroy (&reason);
	// 	return FcFalse;
	//     }
	//     FcStrBufDestroy (&reason);
	//     return ret;
	return nil
}

type FcConfigParse struct {
	logger io.Writer

	pstack  []FcPStack // the top of the stack is at the end of the slice
	vstack  []FcVStack // idem
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

// do not check if the stack is not empty
func (parse *FcConfigParse) p() *FcPStack {
	return &parse.pstack[len(parse.pstack)-1]
}

// return value may be nil
func (parse *FcConfigParse) v() *FcVStack {
	if len(parse.vstack) == 0 {
		return nil
	}
	vstack := &parse.vstack[len(parse.vstack)-1]
	if vstack.pstack == parse.p() {
		return vstack
	}
	return nil
}

func (parser *FcConfigParse) createVAndPush() *FcVStack {
	var v FcVStack
	if len(parser.pstack) >= 2 {
		v.pstack = &parser.pstack[len(parser.pstack)-2]
	}
	parser.vstack = append(parser.vstack, v)
	return &v
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

func (parse *FcConfigParse) vstackPop() {
	if len(parse.vstack) == 0 {
		return
	}
	parse.vstack = parse.vstack[:len(parse.vstack)-1]
}

func (parser *FcConfigParse) FcEndElement() error {
	if len(parser.pstack) == 0 { // nothing to do
		return nil
	}
	var err error
	last := parser.p()
	switch last.element {
	case FcElementDir:
		err = parser.FcParseDir()
	case FcElementCacheDir:
		err = parser.FcParseCacheDir()
	case FcElementCache:
		parser.p().str.Reset() // discard this data; no longer used
	case FcElementInclude:
		err = parser.FcParseInclude()
	case FcElementMatch:
		err = parser.FcParseMatch()
	case FcElementAlias:
		err = parser.FcParseAlias()
		// 	case FcElementDescription:
		// 		FcParseDescription(parser)
		// 	case FcElementRemapDir:
		// 		FcParseRemapDir(parser)
		// 	case FcElementResetDirs:
		// 		FcParseResetDirs(parser)

		// 	case FcElementRescan:
		// 		FcParseRescan(parser)

		// 	case FcElementPrefer:
		// 		FcParseFamilies(parser, FcVStackPrefer)
		// 	case FcElementAccept:
		// 		FcParseFamilies(parser, FcVStackAccept)
		// 	case FcElementDefault:
		// 		FcParseFamilies(parser, FcVStackDefault)
		// 	case FcElementFamily:
		// 		FcParseFamily(parser)

		// 	case FcElementTest:
		// 		FcParseTest(parser)
		// 	case FcElementEdit:
		// 		FcParseEdit(parser)

		// 	case FcElementInt:
		// 		FcParseInt(parser)
		// 	case FcElementDouble:
		// 		FcParseDouble(parser)
		// 	case FcElementString:
		// 		FcParseString(parser, FcVStackString)
		// 	case FcElementMatrix:
		// 		FcParseMatrix(parser)
		// 	case FcElementRange:
		// 		FcParseRange(parser)
		// 	case FcElementBool:
		// 		FcParseBool(parser)
		// 	case FcElementCharSet:
		// 		FcParseCharSet(parser)
		// 	case FcElementLangSet:
		// 		FcParseLangSet(parser)
		// 	case FcElementSelectfont:
		// 	case FcElementAcceptfont, FcElementRejectfont:
		// 		FcParseAcceptRejectFont(parser, parser.pstack.element)
		// 	case FcElementGlob:
		// 		FcParseString(parser, FcVStackGlob)
		// 	case FcElementPattern:
		// 		FcParsePattern(parser)
		// 	case FcElementPatelt:
		// 		FcParsePatelt(parser)
		// 	case FcElementName:
		// 		FcParseName(parser)
		// 	case FcElementConst:
		// 		FcParseString(parser, FcVStackConstant)
	case FcElementOr:
		parser.FcParseBinary(FcOpOr)
	case FcElementAnd:
		parser.FcParseBinary(FcOpAnd)
	case FcElementEq:
		parser.FcParseBinary(FcOpEqual)
	case FcElementNotEq:
		parser.FcParseBinary(FcOpNotEqual)
	case FcElementLess:
		parser.FcParseBinary(FcOpLess)
	case FcElementLessEq:
		parser.FcParseBinary(FcOpLessEqual)
	case FcElementMore:
		parser.FcParseBinary(FcOpMore)
	case FcElementMoreEq:
		parser.FcParseBinary(FcOpMoreEqual)
	case FcElementContains:
		parser.FcParseBinary(FcOpContains)
	case FcElementNotContains:
		parser.FcParseBinary(FcOpNotContains)
	case FcElementPlus:
		parser.FcParseBinary(FcOpPlus)
	case FcElementMinus:
		parser.FcParseBinary(FcOpMinus)
	case FcElementTimes:
		parser.FcParseBinary(FcOpTimes)
	case FcElementDivide:
		parser.FcParseBinary(FcOpDivide)
	case FcElementIf:
		parser.FcParseBinary(FcOpQuest)
	case FcElementNot:
		parser.FcParseUnary(FcOpNot)
	case FcElementFloor:
		parser.FcParseUnary(FcOpFloor)
	case FcElementCeil:
		parser.FcParseUnary(FcOpCeil)
	case FcElementRound:
		parser.FcParseUnary(FcOpRound)
	case FcElementTrunc:
		parser.FcParseUnary(FcOpTrunc)
	}
	parser.pstackPop()
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

// return true if str starts by ~
func usesHome(str string) bool { return strings.HasPrefix(str, "~") }

func xdgDataHome() string {
	if !homeEnabled {
		return ""
	}
	env := os.Getenv("XDG_DATA_HOME")
	if env != "" {
		return env
	}
	home := FcConfigHome()
	return filepath.Join(home, ".local", "share")
}

func xdgCacheHome() string {
	if !homeEnabled {
		return ""
	}
	env := os.Getenv("XDG_CACHE_HOME")
	if env != "" {
		return env
	}
	home := FcConfigHome()
	return filepath.Join(home, ".cache")
}

func xdgConfigHome() string {
	if !homeEnabled {
		return ""
	}
	env := os.Getenv("XDG_CONFIG_HOME")
	if env != "" {
		return env
	}
	home := FcConfigHome()
	return filepath.Join(home, ".config")
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

func (parse *FcConfigParse) FcParseCacheDir() error {
	var prefix string
	attr := parse.getAttr("prefix")
	if attr != "" && attr == "xdg" {
		prefix = xdgCacheHome()
		// home directory might be disabled.: simply ignore this element.
		if prefix == "" {
			return nil
		}
	}
	data := parse.p().str.String()
	if data == "" {
		parse.message(FcSevereWarning, "empty cache directory name ignored")
		return nil
	}
	if prefix != "" {
		data = filepath.Join(prefix, data)
	}
	// TODO: support windows
	// #ifdef _WIN32
	//     else if (data[0] == '/' && fontconfig_instprefix[0] != '\0')   {
	// 	// size_t plen = strlen ((const char *)fontconfig_instprefix);
	// 	// size_t dlen = strlen ((const char *)data);

	// 	prefix = malloc (plen + 1 + dlen + 1);
	// 	// strcpy ((char *) prefix, (char *) fontconfig_instprefix);
	// 	prefix[plen] = FC_DIR_SEPARATOR;
	// 	memcpy (&prefix[plen + 1], data, dlen);
	// 	prefix[plen + 1 + dlen] = 0;
	// 	FcStrFree (data);
	// 	data = prefix;
	//     }  else if data == "WINDOWSTEMPDIR_FONTCONFIG_CACHE" {
	// 	int rc;

	// 	FcStrFree (data);
	// 	data = malloc (1000);
	// 	rc = GetTempPath (800, (LPSTR) data);
	// 	if (rc == 0 || rc > 800) {
	// 	    parse.message ( FcSevereError, "GetTempPath failed");
	// 	    goto bail;
	// 	}
	// 	if (data [strlen ((const char *) data) - 1] != '\\'){
	// 	    strcat ((char *) data, "\\");}
	// 	strcat ((char *) data, "fontconfig\\cache");
	//     }   else if  data ==  "LOCAL_APPDATA_FONTCONFIG_CACHE"  {
	// 	char szFPath[MAX_PATH + 1];
	// 	size_t len;

	// 	if (!(pSHGetFolderPathA && SUCCEEDED(pSHGetFolderPathA(nil, /* CSIDL_LOCAL_APPDATA */ 28, nil, 0, szFPath)))){
	// 	    return errors.New("SHGetFolderPathA failed");
	// 	}
	// 	strncat(szFPath, "\\fontconfig\\cache", MAX_PATH - 1 - strlen(szFPath));
	// 	len = strlen(szFPath) + 1;
	// 	FcStrFree (data);
	// 	data = malloc(len);
	// 	strncpy((char *) data, szFPath, len);
	//     }
	// #endif
	if len(data) == 0 {
		parse.message(FcSevereWarning, "empty cache directory name ignored")
		return nil
	}
	if !parse.scanOnly && (!usesHome(data) || FcConfigHome() != "") {
		err := parse.config.FcConfigAddCacheDir(data)
		if err != nil {
			return fmt.Errorf("fontconfig: cannot add cache directory %s: %s", data, err)
		}
	}
	parse.p().str.Reset()
	return nil
}

func (parser *FcConfigParse) lexBool(bool_ string) FcBool {
	result, err := nameBool(bool_)
	if err != nil {
		parser.message(FcSevereWarning, "\"%s\" is not known boolean", bool_)
	}
	return result
}

func (parser *FcConfigParse) lexBinding(bindingString string) (FcValueBinding, bool) {
	switch bindingString {
	case "", "weak":
		return FcValueBindingWeak, true
	case "strong":
		return FcValueBindingStrong, true
	case "same":
		return FcValueBindingSame, true
	default:
		parser.message(FcSevereWarning, "invalid binding \"%s\"", bindingString)
		return 0, false
	}
}

func isDir(s string) bool {
	f, err := os.Stat(s)
	if err != nil {
		return false
	}
	return f.IsDir()
}

func isFile(s string) bool {
	f, err := os.Stat(s)
	if err != nil {
		return false
	}
	return !f.IsDir()
}

func isLink(s string) bool {
	f, err := os.Stat(s)
	if err != nil {
		return false
	}
	return f.Mode() == os.ModeSymlink
}

// return true on success
func rename(old, new string) bool { return os.Rename(old, new) == nil }

// return true on success
func symlink(old, new string) bool { return os.Symlink(old, new) == nil }

var (
	userdir, userconf string
	userValuesLock    sync.Mutex
)

func getUserdir(s string) string {
	userValuesLock.Lock()
	defer userValuesLock.Unlock()
	if userdir == "" {
		userdir = s
	}
	return userdir
}

func getUserconf(s string) string {
	userValuesLock.Lock()
	defer userValuesLock.Unlock()
	if userconf == "" {
		userconf = s
	}
	return userconf
}

func (parse *FcConfigParse) FcParseInclude() error {
	var (
		ignoreMissing, deprecated bool
		prefix, userdir, userconf string
	)

	s := parse.p().str.String()
	attr := parse.getAttr("ignore_missing")
	if attr != "" && parse.lexBool(attr) == FcTrue {
		ignoreMissing = true
	}
	attr = parse.getAttr("deprecated")
	if attr != "" && parse.lexBool(attr) == FcTrue {
		deprecated = true
	}
	attr = parse.getAttr("prefix")
	if attr == "xdg" {
		prefix = xdgConfigHome()
		// home directory might be disabled: simply ignore this element.
		if prefix == "" {
			return nil
		}
	}
	if prefix != "" {
		s = filepath.Join(prefix, s)
		if isDir(s) {
			userdir = getUserdir(s)
		} else if isFile(s) {
			userconf = getUserconf(s)
		} else {
			/* No config dir nor file on the XDG directory spec compliant place
			 * so need to guess what it is supposed to be.
			 */
			if strings.Index(s, "conf.d") != -1 {
				userdir = getUserdir(s)
			} else {
				userconf = getUserconf(s)
			}
		}
	}

	// flush the ruleset into the queue
	ruleset := parse.ruleset
	parse.ruleset = FcRuleSetCreate(ruleset.name)
	parse.ruleset.enabled = ruleset.enabled
	parse.ruleset.domain, parse.ruleset.description = ruleset.domain, ruleset.description
	for k := range parse.config.subst {
		parse.config.subst[k] = append(parse.config.subst[k], ruleset)
	}

	if err := parseConfig(parse.config, s, !ignoreMissing, !parse.scanOnly); err != nil {
		return err
	}

	if runtime.GOOS != "windows" {
		var warnConf, warnConfd bool
		filename := parse.config.GetFilename(s)
		os.Stat(filename)
		if deprecated == true && filename != "" && userdir != "" && !isLink(filename) {
			if isDir(filename) {
				parent := filepath.Dir(userdir)
				if !isDir(parent) {
					_ = os.Mkdir(parent, 0755)
				}
				if isDir(userdir) || !rename(filename, userdir) || !symlink(userdir, filename) {
					if !warnConfd {
						parse.message(FcSevereWarning, "reading configurations from %s is deprecated. please move it to %s manually", s, userdir)
						warnConfd = true
					}
				}
			} else {
				parent := filepath.Dir(userconf)
				if !isDir(parent) {
					_ = os.Mkdir(parent, 0755)
				}
				if isFile(userconf) || !rename(filename, userconf) || !symlink(userconf, filename) {
					if !warnConf {
						parse.message(FcSevereWarning, "reading configurations from %s is deprecated. please move it to %s manually", s, userconf)
						warnConf = true
					}
				}
			}
		}

	}

	parse.p().str.Reset()
	return nil
}

func (parse *FcConfigParse) FcParseMatch() error {
	var kind FcMatchKind
	kindName := parse.getAttr("target")
	switch kindName {
	case "pattern":
		kind = FcMatchPattern
	case "font":
		kind = FcMatchFont
	case "scan":
		kind = FcMatchScan
	case "":
		kind = FcMatchPattern
	default:
		parse.message(FcSevereWarning, "invalid match target \"%s\"", kindName)
		return nil
	}

	var rules []FcRule // we append, then reverse
	for vstack := parse.v(); vstack != nil; vstack = parse.v() {
		switch vstack.tag {
		case FcVStackTest:
			r := vstack.u
			rules = append(rules, r)
			vstack.tag = FcVStackNone
		case FcVStackEdit:
			edit := vstack.u.(FcEdit)
			if kind == FcMatchScan && edit.object >= fcEnd {
				return fmt.Errorf("<match target=\"scan\"> cannot edit user-defined object \"%s\"", edit.object)
			}
			rules = append(rules, edit)
			vstack.tag = FcVStackNone
		default:
			parse.message(FcSevereWarning, "invalid match element")
		}
		parse.vstackPop()
	}
	revertRules(rules)

	if len(rules) == 0 {
		parse.message(FcSevereWarning, "No <test> nor <edit> elements in <match>")
		return nil
	}
	n := parse.ruleset.add(rules, kind)
	if parse.config.maxObjects < n {
		parse.config.maxObjects = n
	}
	return nil
}

func (parser *FcConfigParse) FcParseAlias() error {
	var (
		family, accept, prefer, def *FcExpr
		rules                       []FcRule // we append, then reverse
	)
	binding, ok := parser.lexBinding(parser.getAttr("binding"))
	if !ok {
		return nil
	}

	for vstack := parser.v(); vstack != nil; vstack = parser.v() {
		switch vstack.tag {
		case FcVStackFamily:
			expr := vstack.u.(*FcExpr)
			if family != nil {
				parser.message(FcSevereWarning, "Having multiple <family> in <alias> isn't supported and may not work as expected")
				family = newExprOp(expr, family, FcOpComma)
			} else {
				family = expr
			}
			vstack.tag = FcVStackNone
		case FcVStackPrefer:
			prefer = vstack.u.(*FcExpr)
			vstack.tag = FcVStackNone
		case FcVStackAccept:
			accept = vstack.u.(*FcExpr)
			vstack.tag = FcVStackNone
		case FcVStackDefault:
			def = vstack.u.(*FcExpr)
			vstack.tag = FcVStackNone
		case FcVStackTest:
			rules = append(rules, vstack.u.(*FcTest))
			vstack.tag = FcVStackNone
		default:
			parser.message(FcSevereWarning, "bad alias")
		}
		parser.vstackPop()
	}
	revertRules(rules)

	if family == nil {
		return fmt.Errorf("missing family in alias")
	}

	if prefer == nil && accept == nil && def == nil {
		return nil
	}

	t := parser.FcTestCreate(FcMatchPattern, FcQualAny, FC_FAMILY,
		opWithFlags(FcOpEqual, FcOpFlagIgnoreBlanks), family)
	rules = append(rules, t)

	if prefer != nil {
		edit := parser.FcEditCreate(FC_FAMILY, FcOpPrepend, prefer, binding)
		rules = append(rules, edit)
	}
	if accept != nil {
		edit := parser.FcEditCreate(FC_FAMILY, FcOpAppend, accept, binding)
		rules = append(rules, edit)
	}
	if def != nil {
		edit := parser.FcEditCreate(FC_FAMILY, FcOpAppendLast, def, binding)
		rules = append(rules, edit)
	}
	n := parser.ruleset.add(rules, FcMatchPattern)
	if parser.config.maxObjects < n {
		parser.config.maxObjects = n
	}
	return nil
}

func (parser *FcConfigParse) FcTestCreate(kind FcMatchKind, qual uint8,
	object FcObject, compare FcOp, expr *FcExpr) *FcTest {
	test := FcTest{kind: kind, qual: qual, op: FcOp(compare), expr: expr}
	o := objects[object.String()]
	test.object = o.object
	if o.parser != nil {
		parser.typecheckExpr(expr, o.parser)
	}
	return &test
}

func (parser *FcConfigParse) FcEditCreate(object FcObject, op FcOp, expr *FcExpr, binding FcValueBinding) FcEdit {
	e := FcEdit{object: object, op: op, expr: expr, binding: binding}
	if o := objects[object.String()]; o.parser != nil {
		parser.typecheckExpr(expr, o.parser)
	}
	return e
}

func (parser *FcConfigParse) popExpr() *FcExpr {
	var expr *FcExpr
	vstack := parser.v()
	if vstack == nil {
		return nil
	}
	switch vstack.tag {
	case FcVStackString, FcVStackFamily:
		expr = &FcExpr{op: FcOpString, u: vstack.u}
	case FcVStackName:
		expr = &FcExpr{op: FcOpField, u: vstack.u}
	case FcVStackConstant:
		expr = &FcExpr{op: FcOpConst, u: vstack.u}
	case FcVStackPrefer, FcVStackAccept, FcVStackDefault:
		expr = vstack.u.(*FcExpr)
		vstack.tag = FcVStackNone
	case FcVStackInteger:
		expr = &FcExpr{op: FcOpInteger, u: vstack.u}
	case FcVStackDouble:
		expr = &FcExpr{op: FcOpDouble, u: vstack.u}
	case FcVStackMatrix:
		expr = &FcExpr{op: FcOpMatrix, u: vstack.u}
	case FcVStackRange:
		expr = &FcExpr{op: FcOpRange, u: vstack.u}
	case FcVStackBool:
		expr = &FcExpr{op: FcOpBool, u: vstack.u}
	case FcVStackCharSet:
		expr = &FcExpr{op: FcOpCharSet, u: vstack.u}
	case FcVStackLangSet:
		expr = &FcExpr{op: FcOpLangSet, u: vstack.u}
	case FcVStackTest, FcVStackExpr:
		expr = vstack.u.(*FcExpr)
		vstack.tag = FcVStackNone
	}
	parser.vstackPop()
	return expr
}

// This builds a tree of binary operations. Note
// that every operator is defined so that if only
// a single operand is contained, the value of the
// whole expression is the value of the operand.
//
// This code reduces in that case to returning that
// operand.
func (parser *FcConfigParse) FcPopBinary(op FcOp) *FcExpr {
	var expr *FcExpr

	for left := parser.popExpr(); left != nil; left = parser.popExpr() {
		if expr != nil {
			expr = newExprOp(left, expr, op)
		} else {
			expr = left
		}
	}
	return expr
}

func (parser *FcConfigParse) FcVStackPushExpr(tag FcVStackTag, expr *FcExpr) {
	vstack := parser.createVAndPush()
	vstack.u = expr
	vstack.tag = tag
}

func (parser *FcConfigParse) FcParseBinary(op FcOp) {
	expr := parser.FcPopBinary(op)
	if expr != nil {
		parser.FcVStackPushExpr(FcVStackExpr, expr)
	}
}

// This builds a a unary operator, it consumes only a single operand
func (parser *FcConfigParse) FcParseUnary(op FcOp) {
	operand := parser.popExpr()
	if operand != nil {
		expr := newExprOp(operand, nil, op)
		parser.FcVStackPushExpr(FcVStackExpr, expr)
	}
}
