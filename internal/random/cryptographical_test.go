package random

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	provider := New()

	require.NotNil(t, provider)

	_, ok := provider.(*Cryptographical)

	assert.True(t, ok)
}

func TestCryptographicalRead(t *testing.T) {
	p := &Cryptographical{}

	testCases := []struct {
		name string
		n    int
	}{
		{"ShouldRead10Bytes", 10},
		{"ShouldRead1Byte", 1},
		{"ShouldRead100Bytes", 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := make([]byte, tc.n)

			n, err := p.Read(data)

			assert.NoError(t, err)
			assert.Equal(t, tc.n, n)
		})
	}
}

func TestCryptographicalBytes(t *testing.T) {
	p := &Cryptographical{}

	t.Run("ShouldReturnDefaultLengthFromBytesErr", func(t *testing.T) {
		data, err := p.BytesErr()

		assert.NoError(t, err)
		assert.Len(t, data, DefaultN)
	})

	t.Run("ShouldReturnDefaultLengthFromBytes", func(t *testing.T) {
		data := p.Bytes()

		assert.Len(t, data, DefaultN)
	})
}

func TestCryptographicalBytesCustomErr(t *testing.T) {
	p := &Cryptographical{}

	testCases := []struct {
		name        string
		n           int
		charset     []byte
		expectedLen int
	}{
		{"ShouldReturnCustomLength", 74, []byte(CharSetAlphabetic), 74},
		{"ShouldReturnExplicitLength", 76, []byte(CharSetAlphabetic), 76},
		{"ShouldReturnDefaultLengthWhenNegative", -5, []byte(CharSetAlphabetic), DefaultN},
		{"ShouldReturnDefaultLengthWhenZero", 0, []byte(CharSetAlphabetic), DefaultN},
		{"ShouldReturnRawBytesWhenCharsetNil", 10, nil, 10},
		{"ShouldReturnRawBytesWhenCharsetEmpty", 10, []byte{}, 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := p.BytesCustomErr(tc.n, tc.charset)

			assert.NoError(t, err)
			assert.Len(t, data, tc.expectedLen)
		})
	}
}

func TestCryptographicalBytesCustom(t *testing.T) {
	p := &Cryptographical{}

	data := p.BytesCustom(20, []byte(CharSetAlphabetic))

	assert.Len(t, data, 20)
}

func TestCryptographicalStringCustom(t *testing.T) {
	p := &Cryptographical{}

	testCases := []struct {
		name        string
		n           int
		charset     string
		expectedLen int
	}{
		{"ShouldReturnCorrectLength", 10, CharSetAlphabetic, 10},
		{"ShouldReturnCorrectLengthAlphaNumeric", 20, CharSetAlphaNumeric, 20},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := p.StringCustom(tc.n, tc.charset)

			assert.Len(t, data, tc.expectedLen)
		})
	}
}

func TestCryptographicalStringCustomErr(t *testing.T) {
	p := &Cryptographical{}

	testCases := []struct {
		name        string
		n           int
		charset     string
		expectedLen int
	}{
		{"ShouldReturnCorrectLength", 11, CharSetAlphabetic, 11},
		{"ShouldReturnDefaultLengthWhenNegative", -1, CharSetAlphabetic, DefaultN},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := p.StringCustomErr(tc.n, tc.charset)

			assert.NoError(t, err)
			assert.Len(t, data, tc.expectedLen)
		})
	}
}

func TestCryptographicalIntn(t *testing.T) {
	p := &Cryptographical{}

	testCases := []struct {
		name string
		n    int
	}{
		{"ShouldReturnValueLessThanN", 999},
		{"ShouldReturnValueLessThan10", 10},
		{"ShouldReturnValueLessThan1", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value := p.Intn(tc.n)

			assert.GreaterOrEqual(t, value, 0)
			assert.Less(t, value, tc.n)
		})
	}
}

func TestCryptographicalIntnErr(t *testing.T) {
	p := &Cryptographical{}

	testCases := []struct {
		name string
		n    int
		err  string
	}{
		{"ShouldReturnValue", 999, ""},
		{"ShouldReturnErrorWhenNegative", -4, "n must be more than 0"},
		{"ShouldReturnErrorWhenZero", 0, "n must be more than 0"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := p.IntnErr(tc.n)

			if len(tc.err) != 0 {
				assert.EqualError(t, err, tc.err)
				assert.Equal(t, 0, value)
			} else {
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, value, 0)
				assert.Less(t, value, tc.n)
			}
		})
	}
}

func TestCryptographicalIntErr(t *testing.T) {
	p := &Cryptographical{}

	testCases := []struct {
		name string
		max  *big.Int
		err  string
	}{
		{"ShouldReturnValue", big.NewInt(999), ""},
		{"ShouldReturnErrorWhenNil", nil, "max is required"},
		{"ShouldReturnErrorWhenNegative", big.NewInt(-1), "max must be 1 or more"},
		{"ShouldReturnErrorWhenZero", big.NewInt(0), "max must be 1 or more"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := p.IntErr(tc.max)

			if len(tc.err) != 0 {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, value)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, value)
				assert.GreaterOrEqual(t, value.Int64(), int64(0))
				assert.Less(t, value.Int64(), tc.max.Int64())
			}
		})
	}
}

func TestCryptographicalInt(t *testing.T) {
	p := &Cryptographical{}

	testCases := []struct {
		name     string
		max      *big.Int
		expected int64
		isError  bool
	}{
		{"ShouldReturnValue", big.NewInt(999), 0, false},
		{"ShouldReturnNegativeOneWhenNil", nil, -1, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value := p.Int(tc.max)

			require.NotNil(t, value)

			if tc.isError {
				assert.Equal(t, tc.expected, value.Int64())
			} else {
				assert.GreaterOrEqual(t, value.Int64(), int64(0))
				assert.Less(t, value.Int64(), tc.max.Int64())
			}
		})
	}
}

func TestCryptographicalPrime(t *testing.T) {
	p := &Cryptographical{}

	testCases := []struct {
		name string
		bits int
	}{
		{"ShouldReturnPrime64", 64},
		{"ShouldReturnPrime128", 128},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prime, err := p.Prime(tc.bits)

			assert.NoError(t, err)
			assert.NotNil(t, prime)
		})
	}
}
