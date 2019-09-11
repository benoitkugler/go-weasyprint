package utils

import "math"

func Min(x, y float32) float32 {
	return float32(math.Min(float64(x), float64(y)))
}

func Max(x, y float32) float32 {
	return float32(math.Max(float64(x), float64(y)))
}

func MinInt(x, y int) int {
	return int(math.Min(float64(x), float64(y)))
}

func MaxInt(x, y int) int {
	return int(math.Max(float64(x), float64(y)))
}
