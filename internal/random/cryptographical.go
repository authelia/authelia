package random

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
)

// Cryptographical is the production random.Provider which uses crypto/rand.
type Cryptographical struct{}

// Read implements the io.Reader interface.
func (r *Cryptographical) Read(p []byte) (n int, err error) {
	return io.ReadFull(rand.Reader, p)
}

// BytesErr returns random data as bytes with the standard random.DefaultN length and can contain any byte values
// (including unreadable byte values). If an error is returned from the random read this function returns it.
func (r *Cryptographical) BytesErr() (data []byte, err error) {
	return r.BytesCustomErr(0, nil)
}

// Bytes returns random data as bytes with the standard random.DefaultN length and can contain any byte values
// (including unreadable byte values). If an error is returned from the random read this function ignores it.
func (r *Cryptographical) Bytes() (data []byte) {
	data, _ = r.BytesErr()

	return data
}

// BytesCustomErr returns random data as bytes with n length and can contain only byte values from the provided
// values. If n is less than 1 then DefaultN is used instead. If an error is returned from the random read this function
// returns it.
func (r *Cryptographical) BytesCustomErr(n int, charset []byte) (data []byte, err error) {
	if n < 1 {
		n = DefaultN
	}

	data = make([]byte, n)

	if _, err = rand.Read(data); err != nil {
		return nil, err
	}

	t := len(charset)

	if t > 0 {
		for i := 0; i < n; i++ {
			data[i] = charset[data[i]%byte(t)] //nolint:gosec // This is safe.
		}
	}

	return data, nil
}

// StringCustomErr is an overload of BytesCustomWithErr which takes a characters string and returns a string.
func (r *Cryptographical) StringCustomErr(n int, characters string) (data string, err error) {
	var d []byte

	if d, err = r.BytesCustomErr(n, []byte(characters)); err != nil {
		return "", err
	}

	return string(d), nil
}

// BytesCustom returns random data as bytes with n length and can contain only byte values from the provided values.
// If n is less than 1 then DefaultN is used instead. If an error is returned from the random read this function
// ignores it.
func (r *Cryptographical) BytesCustom(n int, charset []byte) (data []byte) {
	data, _ = r.BytesCustomErr(n, charset)

	return data
}

// StringCustom is an overload of BytesCustom which takes a characters string and returns a string.
func (r *Cryptographical) StringCustom(n int, characters string) (data string) {
	return string(r.BytesCustom(n, []byte(characters)))
}

// IntnErr returns a random int error combination with a maximum of n.
func (r *Cryptographical) IntnErr(n int) (value int, err error) {
	if n <= 0 {
		return 0, fmt.Errorf("n must be more than 0")
	}

	max := big.NewInt(int64(n))

	var result *big.Int

	if result, err = r.IntErr(max); err != nil {
		return 0, err
	}

	value = int(result.Int64())

	if value < 0 {
		return 0, fmt.Errorf("generated number is too big for int")
	}

	return value, nil
}

// Intn returns a random int with a maximum of n.
func (r *Cryptographical) Intn(n int) (value int) {
	value, _ = r.IntnErr(n)

	return value
}

// IntErr returns a random *big.Int error combination with a maximum of max.
func (r *Cryptographical) IntErr(max *big.Int) (value *big.Int, err error) {
	if max == nil {
		return nil, fmt.Errorf("max is required")
	}

	if max.Sign() <= 0 {
		return nil, fmt.Errorf("max must be 1 or more")
	}

	return rand.Int(rand.Reader, max)
}

// Int returns a random *big.Int with a maximum of max.
func (r *Cryptographical) Int(max *big.Int) (value *big.Int) {
	var err error
	if value, err = r.IntErr(max); err != nil {
		return big.NewInt(-1)
	}

	return value
}

// Prime returns a number of the given bit length that is prime with high probability. Prime will return error for any
// error returned by rand.Read or if bits < 2.
func (r *Cryptographical) Prime(bits int) (prime *big.Int, err error) {
	return rand.Prime(rand.Reader, bits)
}
