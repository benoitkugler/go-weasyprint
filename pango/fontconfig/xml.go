package fontconfig

// import (
// 	"bytes"
// 	"encoding/xml"
// 	"fmt"
// 	"io"
// 	"strings"
// )

// type FcElement uint8

// const (
// 	FcElementNone FcElement = iota
// 	FcElementFontconfig
// 	FcElementDir
// 	FcElementCacheDir
// 	FcElementCache
// 	FcElementInclude
// 	FcElementConfig
// 	FcElementMatch
// 	FcElementAlias
// 	FcElementDescription
// 	FcElementRemapDir
// 	FcElementResetDirs

// 	FcElementRescan

// 	FcElementPrefer
// 	FcElementAccept
// 	FcElementDefault
// 	FcElementFamily

// 	FcElementSelectfont
// 	FcElementAcceptfont
// 	FcElementRejectfont
// 	FcElementGlob
// 	FcElementPattern
// 	FcElementPatelt

// 	FcElementTest
// 	FcElementEdit
// 	FcElementInt
// 	FcElementDouble
// 	FcElementString
// 	FcElementMatrix
// 	FcElementRange
// 	FcElementBool
// 	FcElementCharSet
// 	FcElementLangSet
// 	FcElementName
// 	FcElementConst
// 	FcElementOr
// 	FcElementAnd
// 	FcElementEq
// 	FcElementNotEq
// 	FcElementLess
// 	FcElementLessEq
// 	FcElementMore
// 	FcElementMoreEq
// 	FcElementContains
// 	FcElementNotContains
// 	FcElementPlus
// 	FcElementMinus
// 	FcElementTimes
// 	FcElementDivide
// 	FcElementNot
// 	FcElementIf
// 	FcElementFloor
// 	FcElementCeil
// 	FcElementRound
// 	FcElementTrunc
// 	FcElementUnknown
// )

// var fcElementMap = [...]struct {
// 	name    string
// 	element FcElement
// }{
// 	{"fontconfig", FcElementFontconfig},
// 	{"dir", FcElementDir},
// 	{"cachedir", FcElementCacheDir},
// 	{"cache", FcElementCache},
// 	{"include", FcElementInclude},
// 	{"config", FcElementConfig},
// 	{"match", FcElementMatch},
// 	{"alias", FcElementAlias},
// 	{"description", FcElementDescription},
// 	{"remap-dir", FcElementRemapDir},
// 	{"reset-dirs", FcElementResetDirs},

// 	{"rescan", FcElementRescan},

// 	{"prefer", FcElementPrefer},
// 	{"accept", FcElementAccept},
// 	{"default", FcElementDefault},
// 	{"family", FcElementFamily},

// 	{"selectfont", FcElementSelectfont},
// 	{"acceptfont", FcElementAcceptfont},
// 	{"rejectfont", FcElementRejectfont},
// 	{"glob", FcElementGlob},
// 	{"pattern", FcElementPattern},
// 	{"patelt", FcElementPatelt},

// 	{"test", FcElementTest},
// 	{"edit", FcElementEdit},
// 	{"int", FcElementInt},
// 	{"double", FcElementDouble},
// 	{"string", FcElementString},
// 	{"matrix", FcElementMatrix},
// 	{"range", FcElementRange},
// 	{"bool", FcElementBool},
// 	{"charset", FcElementCharSet},
// 	{"langset", FcElementLangSet},
// 	{"name", FcElementName},
// 	{"const", FcElementConst},
// 	{"or", FcElementOr},
// 	{"and", FcElementAnd},
// 	{"eq", FcElementEq},
// 	{"not_eq", FcElementNotEq},
// 	{"less", FcElementLess},
// 	{"less_eq", FcElementLessEq},
// 	{"more", FcElementMore},
// 	{"more_eq", FcElementMoreEq},
// 	{"contains", FcElementContains},
// 	{"not_contains", FcElementNotContains},
// 	{"plus", FcElementPlus},
// 	{"minus", FcElementMinus},
// 	{"times", FcElementTimes},
// 	{"divide", FcElementDivide},
// 	{"not", FcElementNot},
// 	{"if", FcElementIf},
// 	{"floor", FcElementFloor},
// 	{"ceil", FcElementCeil},
// 	{"round", FcElementRound},
// 	{"trunc", FcElementTrunc},
// }

// var fcElementIgnoreName = [...]string{
// 	"its:",
// }

// func FcElementMap(name string) FcElement {
// 	for _, elem := range fcElementMap {
// 		if name == elem.name {
// 			return elem.element
// 		}
// 	}
// 	for _, ignoreName := range fcElementIgnoreName {
// 		if strings.HasPrefix(name, ignoreName) {
// 			return FcElementNone
// 		}
// 	}
// 	return FcElementUnknown
// }

// type FcPStack struct {
// 	// struct _FcPStack   *prev;
// 	element FcElement
// 	attr    []xml.Attr
// 	str     *bytes.Buffer
// 	// attr_buf_static [16]byte
// }

// type FcVStackTag uint8

// const (
// 	FcVStackNone FcVStackTag = iota

// 	FcVStackString
// 	FcVStackFamily
// 	FcVStackConstant
// 	FcVStackGlob
// 	FcVStackName
// 	FcVStackPattern

// 	FcVStackPrefer
// 	FcVStackAccept
// 	FcVStackDefault

// 	FcVStackInteger
// 	FcVStackDouble
// 	FcVStackMatrix
// 	FcVStackRange
// 	FcVStackBool
// 	FcVStackCharSet
// 	FcVStackLangSet

// 	FcVStackTest
// 	FcVStackExpr
// 	FcVStackEdit
// )

// type FcVStack struct {
// 	// struct _FcVStack	*prev;
// 	pstack *FcPStack /* related parse element */
// 	tag    FcVStackTag
// 	// union {
// 	// FcChar8		*string;

// 	// int		integer;
// 	// double		_double;
// 	// FcExprMatrix	*matrix;
// 	// FcRange		*range;
// 	// FcBool		bool_;
// 	// FcCharSet	*charset;
// 	// FcLangSet	*langset;
// 	// FcExprName	name;

// 	// FcTest		*test;
// 	// FcQual		qual;
// 	// FcOp		op;
// 	// FcExpr		*expr;
// 	// FcEdit		*edit;

// 	// FcPattern	*pattern;
// 	// }
// 	u interface{}
// }

// const (
// 	FcSevereInfo uint8 = iota
// 	FcSevereWarning
// )

// type FcConfigParse struct {
// 	logger io.Writer

// 	pstack  []FcPStack
// 	vstack  []FcVStack
// 	name    string
// 	config  *FcConfig
// 	ruleset *FcRuleSet
// 	// parser   XML_Parser
// 	scanOnly bool
// }

// func (parse FcConfigParse) message(severe uint8, args ...interface{}) {
// 	s := "unknown"
// 	switch severe {
// 	case FcSevereInfo:
// 		s = "info"
// 	case FcSevereWarning:
// 		s = "warning"
// 	}

// 	if parse.name != "" {
// 		fmt.Fprintf(parse.logger, "fontconfig %s: \"%s\": ", s, parse.name)
// 	} else {
// 		fmt.Fprintf(parse.logger, "fontconfig %s: ", s)
// 	}
// 	fmt.Fprintln(parse.logger, args...)
// }

// func (parse *FcConfigParse) FcStartElement(name string, attr []xml.Attr) {
// 	element := FcElementMap(name)

// 	if element == FcElementUnknown {
// 		parse.message(FcSevereWarning, "unknown element", name)
// 	}

// 	parse.pstackPush(element, attr)
// }

// // push at the end of the slice
// func (parse *FcConfigParse) pstackPush(element FcElement, attr []xml.Attr) {
// 	new := FcPStack{
// 		element: element,
// 		attr:    attr,
// 		str:     &bytes.Buffer{},
// 	}
// 	parse.pstack = append(parse.pstack, new)
// }

// func (parse *FcConfigParse) pstackPop() {
// 	// the encoding/xml package makes sur tag are matching
// 	// so parse.pstack has at least one element

// 	// Don't check the attributes for FcElementNone
// 	if last := parse.pstack[len(parse.pstack)-1]; last.element != FcElementNone {
// 		// Warn about unused attrs.
// 		for _, attr := range last.attr {
// 			parse.message(FcSevereWarning, "invalid attribute", attr.Name.Local)
// 		}
// 	}

// 	parse.vstack = nil
// 	parse.pstack = parse.pstack[:len(parse.pstack)-1]
// }

// func (parse *FcConfigParse) FcEndElement() {
// 	if len(parse.pstack) == 0 {
// 		return
// 	}
// 	last := parse.pstack[len(parse.pstack)-1]
// 	switch last.element {
// 	case FcElementDir:
// 		FcParseDir(parse)
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
// 	}
// 	parse.pstackPop()
// }

// func (parse *FcConfigParse) getAttr(attr string) string {
// 	if len(parse.pstack) == 0 {
// 		return ""
// 	}

// 	attrs := parse.pstack[len(parse.pstack)-1].attr

// 	for i, attrXml := range attrs {
// 		if attr == attrXml.Name.Local {
// 			attrs[i].Name.Local == "" // Mark as used.
// 			return attrXml.Value
// 		}
// 	}
// 	return ""
// }

// func (parse *FcConfigParse) FcParseDir() {
// 	data := parse.pstack[len(parse.pstack)-1].str

// 	if data.Len() == 0 {
// 		parse.message(FcSevereWarning, "empty font directory name ignored")
// 		return
// 	}
// 	attr := parse.getAttr("prefix")
// 	salt := parse.getAttr("salt")
// 	prefix := _get_real_path_from_prefix(parse, data, attr)
// 	if prefix == "" {
// 		// nop
// 	} else if !parse.scanOnly && (!FcStrUsesHome(prefix) || FcConfigHome()) {
// 		if !FcConfigAddFontDir(parse.config, prefix, nil, salt) {
// 			parse.message(FcSevereError, "out of memory; cannot add directory", prefix)
// 		}
// 	}
// 	parse.pstack[len(parse.pstack)-1].str = nil
// }
