package properties

import (
	"log"
	"math"

	"github.com/benoitkugler/go-weasyprint/utils"
)

// During layout, float numbers sometimes need special
// values like "auto" or nil (None in Python).
// This file define a float64-like type handling these cases.

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

// Return true except for 0 or nil
func Is(m MaybeFloat) bool {
	if m == nil {
		return false
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

func Maxs(values ...Float) Float {
	var max Float
	for _, w := range values {
		if w > max {
			max = w
		}
	}
	return max
}

func Mins(values ...Float) Float {
	var min Float
	for _, w := range values {
		if w < min {
			min = w
		}
	}
	return min
}

// FloatModulo implements Python modulo for float numbers, like
//	4.456 % 3
func FloatModulo(x Float, i int) Float {
	x2 := Floor(x)
	diff := x - x2
	return Float(utils.ModLikePython(int(x2), i)) + diff
}

func Hypot(a, b Float) Float {
	return Float(math.Hypot(float64(a), float64(b)))
}

func Abs(x Float) Float { return Float(math.Abs(float64(x))) }

// Return the percentage of the reference value, or the value unchanged.
// ``referTo`` is the length for 100%. If ``referTo`` is not a number, it
// just replaces percentages.
func ResoudPercentage(value Value, referTo Float) MaybeFloat {
	if value.IsNone() {
		return nil
	} else if value.String == "auto" {
		return Auto
	} else if value.Unit == Px {
		return value.Value
	} else {
		if value.Unit != Percentage {
			log.Fatalf("expected percentage, got %d", value.Unit)
		}
		return referTo * value.Value / 100.
	}
}
