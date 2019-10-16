package properties

// During layout, float numbers sometimes need special
// values like "auto" or nil (None in Python).
// This file define a float32-like type handling these cases.

const (
	Auto special = true
)

type MaybeFloat interface {
	Auto() bool
	V() Float
}

func (f Float) Auto() bool {
	return false
}

func (f Float) V() Float {
	return f
}

type special bool

func (f special) Auto() bool {
	return bool(f)
}

func (f special) V() Float {
	return -1
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
	if mf.Auto() {
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
