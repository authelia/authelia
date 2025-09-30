package utils

import "cmp"

// Clamp returns value clamped between min and max (inclusive).
func Clamp[T cmp.Ordered](value, min, max T) T {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
