package properties

// During layout, float numbers sometimes need special
// values like "auto" or nil (None in Python).
// This file define a float32-like type handling these cases.

const (
	Auto special = true
)

type MaybeFloat interface {
	Auto() bool
	V() float32
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

func (f Float) Auto() bool {
	return false
}

func (f Float) V() float32 {
	return float32(f)
}

type special bool

func (f special) Auto() bool {
	return bool(f)
}

func (f special) V() float32 {
	return -1
}

func MaybeFloatToValue(mf MaybeFloat) Value {
	if mf == nil {
		return Value{}
	}
	if mf.Auto() {
		return SToV("auto")
	}
	return FToV(mf.V())
}
