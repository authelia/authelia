package random

import (
	"io"
	"math/bits"
)

// bytesCharsetErr returns random data as bytes with n length which only contains byte values from the provided charset.
// This is accomplished using rejection sampling instead of the modulo operator.
func bytesCharsetErr(reader io.Reader, n int, charset []byte) (data []byte, err error) {
	data = make([]byte, n)

	t := uint64(len(charset))

	if t == 1 {
		for i := range data {
			data[i] = charset[0]
		}

		return data, nil
	}

	nbits := bits.Len64(t - 1)
	width := (nbits + 7) / 8
	mask := uint64(1)<<nbits - 1

	buf := make([]byte, n*width)

	for i := 0; i < n; {
		chunk := buf[:(n-i)*width]

		if _, err = io.ReadFull(reader, chunk); err != nil {
			return nil, err
		}

		for j := 0; i < n && j+width <= len(chunk); j += width {
			var value uint64

			for k := range width {
				value = value<<8 | uint64(chunk[j+k])
			}

			if value &= mask; value >= t {
				continue
			}

			data[i] = charset[value]

			i++
		}
	}

	return data, nil
}
