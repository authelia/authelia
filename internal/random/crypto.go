package random

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// Cryptographical is the production random.Provider which uses crypto/rand.
type Cryptographical struct{}

// BytesErr returns random data as bytes with the standard random.DefaultN length and can contain any byte values
// (including unreadable byte values). If an error is returned from the random read this function returns it.
func (r *Cryptographical) BytesErr() (data []byte, err error) {
	data = make([]byte, DefaultN)

	_, err = rand.Read(data)

	return data, err
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

	for i := 0; i < n; i++ {
		data[i] = charset[data[i]%byte(t)]
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

// IntegerErr returns a random int error combination with a maximum of n.
func (r *Cryptographical) IntegerErr(n int) (value int, err error) {
	if n <= 0 {
		return 0, fmt.Errorf("n must be more than 0")
	}

	max := big.NewInt(int64(n))

	if !max.IsUint64() {
		return 0, fmt.Errorf("generated max is negative")
	}

	if nn, err := rand.Int(rand.Reader, max); err != nil {
		return 0, err
	} else {
		value = int(nn.Int64())
	}

	if value < 0 {
		return 0, fmt.Errorf("generated number is too big for int")
	}

	return value, nil
}

// Integer returns a random int with a maximum of n.
func (r *Cryptographical) Integer(n int) (value int) {
	value, _ = r.IntegerErr(n)

	return value
}
