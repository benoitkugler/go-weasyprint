package css

import "math"

func Min(x, y float32) float32 {
	return float32(math.Min(float64(x), float64(y)))
}

func Max(x, y float32) float32 {
	return float32(math.Max(float64(x), float64(y)))
}
