package random

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBytesCharsetErr(t *testing.T) {
	testCases := []struct {
		name    string
		n       int
		charset []byte
	}{
		{"ShouldHandleSingleValueCharset", 20, []byte("a")},
		{"ShouldHandlePowerOfTwoCharset", 500, []byte(CharSetNumericHex)},
		{"ShouldHandleAlphabeticCharset", 500, []byte(CharSetAlphabetic)},
		{"ShouldHandleAlphaNumericCharset", 500, []byte(CharSetAlphaNumeric)},
		{"ShouldHandleASCIICharset", 500, []byte(CharSetASCII)},
		{"ShouldHandleCharsetLargerThanAByte", 500, bytesRange(300)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := bytesCharsetErr(&Cryptographical{}, tc.n, tc.charset)

			require.NoError(t, err)
			require.Len(t, data, tc.n)

			for i, value := range data {
				assert.Containsf(t, tc.charset, value, "byte at index %d with value %d is not a member of the charset", i, value)
			}
		})
	}
}

func TestBytesCharsetErrShouldNotBeBiasedByTheModulo(t *testing.T) {
	testCases := []struct {
		name    string
		charset []byte
	}{
		{"ShouldNotBiasAlphabeticLower", []byte(CharSetAlphabeticLower)},
		{"ShouldNotBiasNumeric", []byte(CharSetNumeric)},
		{"ShouldNotBiasAlphaNumeric", []byte(CharSetAlphaNumeric)},
		{"ShouldNotBiasUnambiguousUpper", []byte(CharSetUnambiguousUpper)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			n := len(tc.charset) * 64

			data, err := bytesCharsetErr(&sequential{}, n, tc.charset)

			require.NoError(t, err)
			require.Len(t, data, n)

			counts := map[byte]int{}

			for _, value := range data {
				counts[value]++
			}

			require.Len(t, counts, len(tc.charset))

			for _, value := range tc.charset {
				assert.Equalf(t, 64, counts[value], "byte with value %d occurred %d times but should have occurred %d times", value, counts[value], 64)
			}
		})
	}
}

func TestBytesCharsetErrShouldReturnReaderErr(t *testing.T) {
	data, err := bytesCharsetErr(&broken{}, 20, []byte(CharSetAlphabetic))

	assert.EqualError(t, err, "broken reader")
	assert.Nil(t, data)
}

func bytesRange(n int) (data []byte) {
	data = make([]byte, n)

	for i := range data {
		data[i] = byte(i % 256)
	}

	return data
}

type sequential struct {
	value byte
}

func (r *sequential) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = r.value
		r.value++
	}

	return len(p), nil
}

type broken struct{}

func (r *broken) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("broken reader")
}
