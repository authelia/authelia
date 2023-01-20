package random

import (
	"fmt"
	"math/big"
	"math/rand"
	"time"
)

// NewMathematical runs rand.Seed with the current time and returns a random.Provider, specifically *random.Mathematical.
func NewMathematical() *Mathematical {
	rand.Seed(time.Now().UnixNano())

	return &Mathematical{}
}

// Mathematical is the random.Provider which uses math/rand and is COMPLETELY UNSAFE FOR PRODUCTION IN MOST SITUATIONS.
// Use random.Cryptographical instead.
type Mathematical struct{}

// Read implements the io.Reader interface.
func (r *Mathematical) Read(p []byte) (n int, err error) {
	return rand.Read(p) //nolint:gosec
}

// BytesErr returns random data as bytes with the standard random.DefaultN length and can contain any byte values
// (including unreadable byte values). If an error is returned from the random read this function returns it.
func (r *Mathematical) BytesErr() (data []byte, err error) {
	data = make([]byte, DefaultN)

	if _, err = rand.Read(data); err != nil { //nolint:gosec
		return nil, err
	}

	return data, nil
}

// Bytes returns random data as bytes with the standard random.DefaultN length and can contain any byte values
// (including unreadable byte values). If an error is returned from the random read this function ignores it.
func (r *Mathematical) Bytes() (data []byte) {
	data, _ = r.BytesErr()

	return data
}

// BytesCustomErr returns random data as bytes with n length and can contain only byte values from the provided
// values. If n is less than 1 then DefaultN is used instead. If an error is returned from the random read this function
// returns it.
func (r *Mathematical) BytesCustomErr(n int, charset []byte) (data []byte, err error) {
	if n < 1 {
		n = DefaultN
	}

	data = make([]byte, n)

	if _, err = rand.Read(data); err != nil { //nolint:gosec
		return nil, err
	}

	t := len(charset)

	for i := 0; i < n; i++ {
		data[i] = charset[data[i]%byte(t)]
	}

	return data, nil
}

// StringCustomErr is an overload of BytesCustomWithErr which takes a characters string and returns a string.
func (r *Mathematical) StringCustomErr(n int, characters string) (data string, err error) {
	var d []byte

	if d, err = r.BytesCustomErr(n, []byte(characters)); err != nil {
		return "", err
	}

	return string(d), nil
}

// BytesCustom returns random data as bytes with n length and can contain only byte values from the provided values.
// If n is less than 1 then DefaultN is used instead. If an error is returned from the random read this function
// ignores it.
func (r *Mathematical) BytesCustom(n int, charset []byte) (data []byte) {
	data, _ = r.BytesCustomErr(n, charset)

	return data
}

// StringCustom is an overload of BytesCustom which takes a characters string and returns a string.
func (r *Mathematical) StringCustom(n int, characters string) (data string) {
	return string(r.BytesCustom(n, []byte(characters)))
}

// IntErr returns a random *big.Int error combination with a maximum of max.
func (r *Mathematical) IntErr(max *big.Int) (value *big.Int, err error) {
	if max == nil {
		return nil, fmt.Errorf("max is required")
	}

	if max.Sign() <= 0 {
		return nil, fmt.Errorf("max must be 1 or more")
	}

	return big.NewInt(int64(rand.Intn(max.Sign()))), nil //nolint:gosec
}

// Int returns a random *big.Int with a maximum of max.
func (r *Mathematical) Int(max *big.Int) (value *big.Int) {
	var err error

	if value, err = r.IntErr(max); err != nil {
		return big.NewInt(-1)
	}

	return value
}

// IntegerErr returns a random int error combination with a maximum of n.
func (r *Mathematical) IntegerErr(n int) (output int, err error) {
	return r.Integer(n), nil
}

// Integer returns a random int with a maximum of n.
func (r *Mathematical) Integer(n int) int {
	return rand.Intn(n) //nolint:gosec
}
