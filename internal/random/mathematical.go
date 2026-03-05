package random

import (
	crand "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"time"
)

// NewMathematical runs rand.Seed with the current time and returns a random.Provider, specifically *random.Mathematical.
func NewMathematical() *Mathematical {
	return &Mathematical{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())), //nolint:gosec
		lock: &sync.Mutex{},
	}
}

// Mathematical is the random.Provider which uses math/rand and is COMPLETELY UNSAFE FOR PRODUCTION IN MOST SITUATIONS.
// Use random.Cryptographical instead.
type Mathematical struct {
	rand *rand.Rand
	lock *sync.Mutex
}

// Read implements the io.Reader interface.
func (r *Mathematical) Read(p []byte) (n int, err error) {
	r.lock.Lock()

	defer r.lock.Unlock()

	return r.rand.Read(p)
}

// BytesErr returns random data as bytes with the standard random.DefaultN length and can contain any byte values
// (including unreadable byte values). If an error is returned from the random read this function returns it.
func (r *Mathematical) BytesErr() (data []byte, err error) {
	data = make([]byte, DefaultN)

	if _, err = r.Read(data); err != nil {
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

	if _, err = r.Read(data); err != nil {
		return nil, err
	}

	t := len(charset)

	for i := 0; i < n; i++ {
		data[i] = charset[data[i]%byte(t)] //nolint:gosec // This is safe.
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

// Intn returns a random int with a maximum of n.
func (r *Mathematical) Intn(n int) int {
	r.lock.Lock()

	defer r.lock.Unlock()

	return r.rand.Intn(n)
}

// IntnErr returns a random int error combination with a maximum of n.
func (r *Mathematical) IntnErr(n int) (output int, err error) {
	if n <= 0 {
		return 0, fmt.Errorf("n must be more than 0")
	}

	return r.Intn(n), nil
}

// Int returns a random *big.Int with a maximum of max.
func (r *Mathematical) Int(max *big.Int) (value *big.Int) {
	var err error
	if value, err = r.IntErr(max); err != nil {
		return big.NewInt(-1)
	}

	return value
}

// IntErr returns a random *big.Int error combination with a maximum of max.
func (r *Mathematical) IntErr(max *big.Int) (value *big.Int, err error) {
	if max == nil {
		return nil, fmt.Errorf("max is required")
	}

	if max.Int64() <= 0 {
		return nil, fmt.Errorf("max must be 1 or more")
	}

	return big.NewInt(int64(r.Intn(int(max.Int64())))), nil
}

// Prime returns a number of the given bit length that is prime with high probability. Prime will return error for any
// error returned by rand.Read or if bits < 2.
func (r *Mathematical) Prime(bits int) (prime *big.Int, err error) {
	return crand.Prime(r, bits)
}
