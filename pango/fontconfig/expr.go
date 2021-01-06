package fontconfig

import (
	"log"
	"math"
)

type FcOp uint8

const (
	FcOpInteger FcOp = iota
	FcOpDouble
	FcOpString
	FcOpMatrix
	FcOpRange
	FcOpBool
	FcOpCharSet
	FcOpLangSet
	FcOpNil
	FcOpField
	FcOpConst
	FcOpAssign
	FcOpAssignReplace
	FcOpPrependFirst
	FcOpPrepend
	FcOpAppend
	FcOpAppendLast
	FcOpDelete
	FcOpDeleteAll
	FcOpQuest
	FcOpOr
	FcOpAnd
	FcOpEqual
	FcOpNotEqual
	FcOpContains
	FcOpListing
	FcOpNotContains
	FcOpLess
	FcOpLessEqual
	FcOpMore
	FcOpMoreEqual
	FcOpPlus
	FcOpMinus
	FcOpTimes
	FcOpDivide
	FcOpNot
	FcOpComma
	FcOpFloor
	FcOpCeil
	FcOpRound
	FcOpTrunc
	FcOpInvalid
)

func (x FcOp) getFlags() int {
	return (int(x) & 0xffff0000) >> 16
}

const FcOpFlagIgnoreBlanks = 1

type FcExprMatrix struct {
	xx, xy, yx, yy *FcExpr
}

type FcExprName struct {
	object FcObject
	kind   FcMatchKind
}

type exprTree struct {
	left, right *FcExpr
}

type FcExpr struct {
	op FcOp
	u  interface{}
}

// union {
// int		ival;
// double		dval;
// const FcChar8	*sval;
// FcExprMatrix	*mexpr;
// FcBool		bval;
// FcCharSet	*cval;
// FcLangSet	*lval;
// FcRange		*rval;

// FcExprName	name;
// const FcChar8	*constant;
// struct {
//     struct _FcExpr *left, *right;
// } tree;
// } u;

func (e *FcExpr) FcConfigEvaluate(p, p_pat *FcPattern, kind FcMatchKind) FcValue {
	var v FcValue

	switch e.op {
	case FcOpInteger, FcOpDouble, FcOpString, FcOpCharSet, FcOpLangSet, FcOpRange, FcOpBool:
		v = e.u
	case FcOpMatrix:
		mexpr := e.u.(FcExprMatrix)
		v = FcMatrix{} // promotion hint
		xx, xxIsFloat := FcConfigPromote(mexpr.xx.FcConfigEvaluate(p, p_pat, kind), v).(float64)
		xy, xyIsFloat := FcConfigPromote(mexpr.xy.FcConfigEvaluate(p, p_pat, kind), v).(float64)
		yx, yxIsFloat := FcConfigPromote(mexpr.yx.FcConfigEvaluate(p, p_pat, kind), v).(float64)
		yy, yyIsFloat := FcConfigPromote(mexpr.yy.FcConfigEvaluate(p, p_pat, kind), v).(float64)

		if xxIsFloat && xyIsFloat && yxIsFloat && yyIsFloat {
			v = FcMatrix{xx: xx, xy: xy, yx: yx, yy: yy}
		} else {
			v = nil
		}
	case FcOpField:
		name := e.u.(FcExprName)
		var res FcResult
		if kind == FcMatchFont && name.kind == FcMatchPattern {
			v, res = p_pat.FcPatternObjectGet(name.object, 0)
			if res != FcResultMatch {
				v = nil
			}
		} else if kind == FcMatchPattern && name.kind == FcMatchFont {
			log.Println("fFontconfig: <name> tag has target=\"font\" in a <match target=\"pattern\">.")
			v = nil
		} else {
			v, res = p_pat.FcPatternObjectGet(name.object, 0)
			if res != FcResultMatch {
				v = nil
			}
		}
	case FcOpConst:
		if ct, ok := FcNameConstant(e.u.(string)); ok {
			v = ct
		} else {
			v = nil
		}
	case FcOpQuest:
		tree := e.u.(exprTree)
		vl := tree.left.FcConfigEvaluate(p, p_pat, kind)
		if vb, isBool := vl.(FcBool); isBool {
			right := tree.right.u.(exprTree)
			if vb != 0 {
				v = right.left.FcConfigEvaluate(p, p_pat, kind)
			} else {
				v = right.right.FcConfigEvaluate(p, p_pat, kind)
			}
		} else {
			v = nil
		}
	case FcOpEqual, FcOpNotEqual, FcOpLess, FcOpLessEqual, FcOpMore, FcOpMoreEqual, FcOpContains, FcOpNotContains, FcOpListing:
		tree := e.u.(exprTree)
		vl := tree.left.FcConfigEvaluate(p, p_pat, kind)
		vr := tree.right.FcConfigEvaluate(p, p_pat, kind)
		cp := FcConfigCompareValue(vl, e.op, vr)
		v = FcFalse
		if cp {
			v = FcTrue
		}
	case FcOpOr, FcOpAnd, FcOpPlus, FcOpMinus, FcOpTimes, FcOpDivide:
		tree := e.u.(exprTree)
		vl := tree.left.FcConfigEvaluate(p, p_pat, kind)
		vr := tree.right.FcConfigEvaluate(p, p_pat, kind)
		vle := FcConfigPromote(vl, vr)
		vre := FcConfigPromote(vr, vle)
		v = nil
		switch vle := vle.(type) {
		case float64:
			vre, sameType := vre.(float64)
			if !sameType {
				break
			}
			switch e.op {
			case FcOpPlus:
				v = vle + vre
			case FcOpMinus:
				v = vle - vre
			case FcOpTimes:
				v = vle * vre
			case FcOpDivide:
				v = vle / vre
			}
			if vf, ok := v.(float64); ok && vf == float64(int(vf)) {
				v = int(vf)
			}
		case FcBool:
			vre, sameType := vre.(FcBool)
			if !sameType {
				break
			}
			switch e.op {
			case FcOpOr:
				v = vle | vre
			case FcOpAnd:
				v = vle & vre
			}
		case string:
			vre, sameType := vre.(string)
			if !sameType {
				break
			}
			switch e.op {
			case FcOpPlus:
				v = vle + vre
			}
		case FcMatrix:
			vre, sameType := vre.(FcMatrix)
			if !sameType {
				break
			}
			switch e.op {
			case FcOpTimes:
				v = vle.multiply(vre)
			}
		case FcCharSet:
			vre, sameType := vre.(FcCharSet)
			if !sameType {
				break
			}
			switch e.op {
			case FcOpPlus:
				if uc := FcCharSetUnion(vle, vre); uc != nil {
					v = *uc
				}
			case FcOpMinus:
				if sc := FcCharSetSubtract(vle, vre); sc != nil {
					v = sc
				}
			}
		case FcLangSet:
			vre, sameType := vre.(FcLangSet)
			if !sameType {
				break
			}
			switch e.op {
			case FcOpPlus:
				v = FcLangSetUnion(vle, vre)
			case FcOpMinus:
				v = FcLangSetSubtract(vle, vre)
			}
		}
	case FcOpNot:
		tree := e.u.(exprTree)
		vl := tree.left.FcConfigEvaluate(p, p_pat, kind)
		if b, ok := vl.(FcBool); ok {
			v = 1 - b&1
		}
	case FcOpFloor, FcOpCeil, FcOpRound, FcOpTrunc:
		tree := e.u.(exprTree)
		vl := tree.left.FcConfigEvaluate(p, p_pat, kind)
		switch vl := vl.(type) {
		case int:
			v = vl
		case float64:
			switch e.op {
			case FcOpFloor:
				v = int(math.Floor(vl))
			case FcOpCeil:
				v = int(math.Ceil(vl))
			case FcOpRound:
				v = int(math.Round(vl))
			case FcOpTrunc:
				v = int(math.Trunc(vl))
			}
		}
	}
	return v
}

// the C implemention use a pre-allocated buffer to avoid allocations
// we choose to simplify and not use buffer
func FcConfigPromote(v, u FcValue) FcValue {
	switch val := v.(type) {
	case int:
		v = promoteFloat64(float64(val), u)
	case float64:
		v = promoteFloat64(val, u)
	case nil:
		switch u.(type) {
		case FcMatrix:
			v = FcIdentityMatrix
		case FcLangSet:
			v = FcLangSetPromote("")
		case FcCharSet:
			v = FcCharSet{}
		}
	case string:
		if _, ok := u.(FcLangSet); ok {
			v = FcLangSetPromote(val)
		}
	}
	return v
}

func promoteFloat64(val float64, u FcValue) FcValue {
	if _, ok := u.(FcRange); ok {
		return FcRangePromote(val)
	}
	return val
}

func FcConfigCompareValue(left_o FcValue, op FcOp, right_o FcValue) bool {
	flags := op.getFlags()

	retNoMatchingType := false
	if op == FcOpNotEqual || op == FcOpNotContains {
		retNoMatchingType = true
	}
	ret := false

	// to avoid checking for type equality we begin by promoting
	// and we will check later in the type switch
	left_o = FcConfigPromote(left_o, right_o)
	left_o = FcConfigPromote(right_o, left_o)

	switch l := left_o.(type) {
	case int:
		r, sameType := right_o.(int)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case FcOpEqual, FcOpContains, FcOpListing:
			ret = l == r
		case FcOpNotEqual, FcOpNotContains:
			ret = l != r
		case FcOpLess:
			ret = l < r
		case FcOpLessEqual:
			ret = l <= r
		case FcOpMore:
			ret = l > r
		case FcOpMoreEqual:
			ret = l >= r
		}
	case float64:
		r, sameType := right_o.(float64)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case FcOpEqual, FcOpContains, FcOpListing:
			ret = l == r
		case FcOpNotEqual, FcOpNotContains:
			ret = l != r
		case FcOpLess:
			ret = l < r
		case FcOpLessEqual:
			ret = l <= r
		case FcOpMore:
			ret = l > r
		case FcOpMoreEqual:
			ret = l >= r
		}
	case FcBool:
		r, sameType := right_o.(FcBool)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case FcOpEqual:
			ret = l == r
		case FcOpContains, FcOpListing:
			ret = l == r || l >= FcDontCare
		case FcOpNotEqual:
			ret = l != r
		case FcOpNotContains:
			ret = !(l == r || l >= FcDontCare)
		case FcOpLess:
			ret = l != r && r >= FcDontCare
		case FcOpLessEqual:
			ret = l == r || r >= FcDontCare
		case FcOpMore:
			ret = l != r && l >= FcDontCare
		case FcOpMoreEqual:
			ret = l == r || l >= FcDontCare
		}
	case string:
		r, sameType := right_o.(string)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case FcOpEqual, FcOpListing:
			if flags&FcOpFlagIgnoreBlanks != 0 {
				ret = FcStrCmpIgnoreBlanksAndCase(l, r) == 0
			} else {
				ret = FcStrCmpIgnoreCase(l, r) == 0
			}
		case FcOpContains:
			ret = FcStrStrIgnoreCase(l, r) != -1
		case FcOpNotEqual:
			if flags&FcOpFlagIgnoreBlanks != 0 {
				ret = FcStrCmpIgnoreBlanksAndCase(l, r) != 0
			} else {
				ret = FcStrCmpIgnoreCase(l, r) != 0
			}
		case FcOpNotContains:
			ret = FcStrStrIgnoreCase(l, r) == -1
		}
	case FcMatrix:
		r, sameType := right_o.(FcMatrix)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case FcOpEqual, FcOpContains, FcOpListing:
			ret = l == r
		case FcOpNotEqual, FcOpNotContains:
			ret = !(l == r)
		}
	case FcCharSet:
		r, sameType := right_o.(FcCharSet)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case FcOpContains, FcOpListing:
			// left contains right if right is a subset of left
			ret = r.FcCharSetIsSubset(l)
		case FcOpNotContains:
			// left contains right if right is a subset of left
			ret = !r.FcCharSetIsSubset(l)
		case FcOpEqual:
			ret = FcCharSetEqual(l, r)
		case FcOpNotEqual:
			ret = !FcCharSetEqual(l, r)
		}
	case FcLangSet:
		r, sameType := right_o.(FcLangSet)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case FcOpContains, FcOpListing:
			ret = l.FcLangSetContains(r)
		case FcOpNotContains:
			ret = !l.FcLangSetContains(r)
		case FcOpEqual:
			ret = FcLangSetEqual(l, r)
		case FcOpNotEqual:
			ret = !FcLangSetEqual(l, r)
		}
	case nil:
		sameType := right_o == nil
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case FcOpEqual, FcOpContains, FcOpListing:
			ret = true
		}
	case *FtFace:
		r, sameType := right_o.(*FtFace)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case FcOpEqual, FcOpContains, FcOpListing:
			ret = l == r
		case FcOpNotEqual, FcOpNotContains:
			ret = l != r
		}
	case FcRange:
		r, sameType := right_o.(FcRange)
		if !sameType {
			return retNoMatchingType
		}
		ret = FcRangeCompare(op, l, r)
	}
	return ret
}

func (e *FcExpr) FcConfigValues(p, p_pat *FcPattern, kind FcMatchKind, binding FcValueBinding) FcValueList {
	if e == nil {
		return nil
	}

	var l FcValueList
	if e.op == FcOpComma {
		tree := e.u.(exprTree)
		v := tree.left.FcConfigEvaluate(p, p_pat, kind)
		next := tree.right.FcConfigValues(p, p_pat, kind, binding)
		l = append(FcValueList{valueElt{value: v, binding: binding}}, next...)
	} else {
		v := e.FcConfigEvaluate(p, p_pat, kind)
		l = FcValueList{valueElt{value: v, binding: binding}}
	}

	if l[0].value == nil {
		l = l[1:]
	}

	return l
}
