package random

import (
	"math/rand"
	"time"
)

// Mathematical is the random.Provider which uses math/rand and is PROBABLY UNSAFE FOR PRODUCTION IN MOST SITUATIONS.
// Use random.Cryptographical instead.
type Mathematical struct{}

// Generate returns random data as bytes with the standard random.DefaultN length and can contain any byte values
// (including unreadable byte values).
func (r *Mathematical) Generate() (data []byte) {
	data = make([]byte, DefaultN)

	_, _ = rand.Read(data)

	return data
}

// GenerateCustom returns random data as bytes with n length and can contain only byte values from the provided values.
// If n is less than 1 then DefaultN is used instead.
func (r *Mathematical) GenerateCustom(n int, charset []byte) (data []byte) {
	if n < 1 {
		n = DefaultN
	}

	data = make([]byte, n)

	_, _ = rand.Read(data)

	t := len(charset)

	for i := 0; i < n; i++ {
		data[i] = charset[data[i]%byte(t)]
	}

	return data
}

// GenerateString is an overload of GenerateCustom which takes a characters string and returns a string.
func (r *Mathematical) GenerateString(n int, characters string) (data string) {
	return string(r.GenerateCustom(n, []byte(characters)))
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
