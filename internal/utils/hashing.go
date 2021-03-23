package utils

import (
	"crypto/sha256"
	"fmt"
)

// HashSHA256FromString takes an input string and calculates the SHA256 checksum returning it as a base16 hash string.
func HashSHA256FromString(input string) (output string) {
	sum := sha256.Sum256([]byte(input))

	return fmt.Sprintf("%x", sum)
}
