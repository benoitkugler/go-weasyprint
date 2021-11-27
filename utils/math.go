package utils

import (
	"math"
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

type Fl = float64

func Maxs(values ...Fl) Fl {
	max := values[0]
	for _, w := range values {
		if w > max {
			max = w
		}
	}
	return max
}

func Mins(values ...Fl) Fl {
	min := values[0]
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
func FloatModulo(x Fl, i int) Fl {
	x2 := math.Floor(x)
	diff := x - x2
	return float64(ModLikePython(int(x2), i)) + diff
}

// Round rounds f with 12 digits precision
func Round(f Fl) Fl {
	n := math.Pow10(12)
	return Fl(math.Round(float64(f)*n) / n)
}
