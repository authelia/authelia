// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

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
