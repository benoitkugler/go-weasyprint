package utils

func Min(x, y float32) float32 {
	if x < y {
		return x
	}
	return y
}

func Max(x, y float32) float32 {
	if x > y {
		return x
	}
	return y
}

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
