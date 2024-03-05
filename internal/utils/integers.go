package utils

func IsIntegerInSlice(needle int, haystack []int) bool {
	for _, n := range haystack {
		if n == needle {
			return true
		}
	}

	return false
}
