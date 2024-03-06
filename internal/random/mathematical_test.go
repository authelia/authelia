package random

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMathematical(t *testing.T) {
	p := NewMathematical()

	data := make([]byte, 10)

	n, err := p.Read(data)
	assert.Equal(t, 10, n)
	assert.NoError(t, err)

	data2, err := p.BytesErr()
	assert.NoError(t, err)
	assert.Len(t, data2, 72)

	data2 = p.Bytes()
	assert.Len(t, data2, 72)

	data2 = p.BytesCustom(74, []byte(CharSetAlphabetic))
	assert.Len(t, data2, 74)

	data2, err = p.BytesCustomErr(76, []byte(CharSetAlphabetic))
	assert.NoError(t, err)
	assert.Len(t, data2, 76)

	data2, err = p.BytesCustomErr(-5, []byte(CharSetAlphabetic))
	assert.NoError(t, err)
	assert.Len(t, data2, 72)

	strdata := p.StringCustom(10, CharSetAlphabetic)
	assert.Len(t, strdata, 10)

	strdata, err = p.StringCustomErr(11, CharSetAlphabetic)
	assert.NoError(t, err)
	assert.Len(t, strdata, 11)

	i := p.Intn(999)
	assert.Greater(t, i, -1)
	assert.Less(t, i, 999)

	i, err = p.IntnErr(999)
	assert.NoError(t, err)
	assert.Greater(t, i, 0)
	assert.Less(t, i, 999)

	i, err = p.IntnErr(-4)
	assert.EqualError(t, err, "n must be more than 0")
	assert.Equal(t, 0, i)

	bi := p.Int(big.NewInt(999))
	assert.Greater(t, bi.Int64(), int64(0))
	assert.Less(t, bi.Int64(), int64(999))

	bi = p.Int(nil)
	assert.Equal(t, int64(-1), bi.Int64())

	bi, err = p.IntErr(nil)
	assert.Nil(t, bi)
	assert.EqualError(t, err, "max is required")

	bi, err = p.IntErr(big.NewInt(-1))
	assert.Nil(t, bi)
	assert.EqualError(t, err, "max must be 1 or more")

	prime, err := p.Prime(64)
	assert.NoError(t, err)
	assert.NotNil(t, prime)
}
