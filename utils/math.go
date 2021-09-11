package utils

import (
	"math"

	"github.com/benoitkugler/textlayout/pango"
)

func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func MaxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

type Fl = pango.Fl

func Maxs(values ...Fl) Fl {
	var max Fl
	for _, w := range values {
		if w > max {
			max = w
		}
	}
	return max
}

func Mins(values ...Fl) Fl {
	var min Fl
	for _, w := range values {
		if w < min {
			min = w
		}
	}
	return min
}

func ModLikePython(d, m int) int {
	var res int = d % m
	if (res < 0 && m > 0) || (res > 0 && m < 0) {
		return res + m
	}
	return res
}

// FloatModulo implements Python modulo for float numbers, like
//	4.456 % 3
func FloatModulo(x float64, i int) float64 {
	x2 := math.Floor(x)
	diff := x - x2
	return float64(ModLikePython(int(x2), i)) + diff
}
