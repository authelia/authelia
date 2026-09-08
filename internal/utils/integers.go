package utils

// IsIntegerInSlice returns true if the given integer is in the given slice.
func IsIntegerInSlice(needle int, haystack []int) bool {
	for _, n := range haystack {
		if n == needle {
			return true
		}
	}

	return false
}
