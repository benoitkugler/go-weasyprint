package properties

import "math"

// During layout, float numbers sometimes need special
// values like "auto" or nil (None in Python).
// This file define a float32-like type handling these cases.

const (
	Auto special = true
)

type MaybeFloat interface {
	V() Float
}

func (f Float) V() Float {
	return f
}

type special bool

func (f special) V() Float {
	return 0
}

// Return true except for 0 or None
func Is(m MaybeFloat) bool {
	if m == nil {
		return true
	}
	if f, ok := m.(Float); ok {
		return f != 0
	}
	return false
}

func MaybeFloatToValue(mf MaybeFloat) Value {
	if mf == nil {
		return Value{}
	}
	if mf == Auto {
		return SToV("auto")
	}
	return mf.V().ToValue()
}

func Min(x, y Float) Float {
	if x < y {
		return x
	}
	return y
}

func Max(x, y Float) Float {
	if x > y {
		return x
	}
	return y
}

func Floor(x Float) Float {
	return Float(math.Floor(float64(x)))
}

func Maxs(values []Float) Float {
	var max Float
	for _, w := range values {
		if w > max {
			max = w
		}
	}
	return max
}
