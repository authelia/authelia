package utils

import (
	cryptorand "crypto/rand"
	"math/rand"
	"time"
)

func IsStringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func IsStringBase64Valid(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && (r < '+' || r > '/') {
			return false
		}
	}
	return true
}

// Splits a string s into an array with each item being a max of int d
func SplitStringToArrayOfStrings(s string, d int) (slice []string) {
	l := len(s)
	n := l / d
	r := l & d
	for i := 0; i < n; i++ {
		slice = append(slice, s[i*d:i*d+d])
		if i+1 == n && r != 0 {
			slice = append(slice, s[i*d+d:])
		}
	}
	return
}

// RandomString generate a random string of n characters
func RandomString(n int, characters []rune) string {
	prime, err := cryptorand.Prime(cryptorand.Reader, 1024)
	if err != nil {
		rand.Seed(time.Now().UnixNano())
	} else {
		rand.Seed(prime.Int64())
	}
	b := make([]rune, n)
	for i := range b {
		b[i] = characters[rand.Intn(len(characters))]
	}
	return string(b)
}
