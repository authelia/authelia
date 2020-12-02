package utils

import (
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldSplitIntoEvenStringsOfFour(t *testing.T) {
	input := testStringInput
	arrayOfStrings := SliceString(input, 4)
	assert.Equal(t, len(arrayOfStrings), 3)
	assert.Equal(t, "abcd", arrayOfStrings[0])
	assert.Equal(t, "efgh", arrayOfStrings[1])
	assert.Equal(t, "ijkl", arrayOfStrings[2])
}

func TestShouldSplitIntoEvenStringsOfOne(t *testing.T) {
	input := testStringInput
	arrayOfStrings := SliceString(input, 1)
	assert.Equal(t, 12, len(arrayOfStrings))
	assert.Equal(t, "a", arrayOfStrings[0])
	assert.Equal(t, "b", arrayOfStrings[1])
	assert.Equal(t, "c", arrayOfStrings[2])
	assert.Equal(t, "d", arrayOfStrings[3])
	assert.Equal(t, "l", arrayOfStrings[11])
}

func TestShouldSplitIntoUnevenStringsOfFour(t *testing.T) {
	input := testStringInput + "m"
	arrayOfStrings := SliceString(input, 4)
	assert.Equal(t, len(arrayOfStrings), 4)
	assert.Equal(t, "abcd", arrayOfStrings[0])
	assert.Equal(t, "efgh", arrayOfStrings[1])
	assert.Equal(t, "ijkl", arrayOfStrings[2])
	assert.Equal(t, "m", arrayOfStrings[3])
}

func TestShouldFindSliceDifferencesDelta(t *testing.T) {
	before := []string{"abc", "onetwothree"}
	after := []string{"abc", "xyz"}
	added, removed := StringSlicesDelta(before, after)
	require.Len(t, added, 1)
	require.Len(t, removed, 1)
	assert.Equal(t, "onetwothree", removed[0])
	assert.Equal(t, "xyz", added[0])
}

func TestShouldNotFindSliceDifferencesDelta(t *testing.T) {
	before := []string{"abc", "onetwothree"}
	after := []string{"abc", "onetwothree"}
	added, removed := StringSlicesDelta(before, after)
	require.Len(t, added, 0)
	require.Len(t, removed, 0)
}

func TestShouldFindSliceDifferences(t *testing.T) {
	a := []string{"abc", "onetwothree"}
	b := []string{"abc", "xyz"}
	diff := IsStringSlicesDifferent(a, b)
	assert.True(t, diff)
}

func TestShouldNotFindSliceDifferences(t *testing.T) {
	a := []string{"abc", "onetwothree"}
	b := []string{"abc", "onetwothree"}
	diff := IsStringSlicesDifferent(a, b)
	assert.False(t, diff)
}

func TestShouldReturnCorrectTLSVersions(t *testing.T) {
	tls13 := "1.3"
	tls13Uint := uint16(tls.VersionTLS13)
	tls12 := "1.2"
	tls12Uint := uint16(tls.VersionTLS12)
	tls11 := "1.1"
	tls11Uint := uint16(tls.VersionTLS11)
	tls10 := "1.0"
	tls10Uint := uint16(tls.VersionTLS10)

	version, err := TLSStringToTLSConfigVersion(tls13)
	assert.Equal(t, tls13Uint, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion("TLS" + tls13)
	assert.Equal(t, tls13Uint, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion(tls12)
	assert.Equal(t, tls12Uint, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion("TLS" + tls12)
	assert.Equal(t, tls12Uint, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion(tls11)
	assert.Equal(t, tls11Uint, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion("TLS" + tls11)
	assert.Equal(t, tls11Uint, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion(tls10)
	assert.Equal(t, tls10Uint, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion("TLS" + tls10)
	assert.Equal(t, tls10Uint, version)
	assert.NoError(t, err)
}

func TestShouldReturnZeroAndErrorOnInvalidTLSVersions(t *testing.T) {
	version, err := TLSStringToTLSConfigVersion("TLS1.4")
	assert.Error(t, err)
	assert.Equal(t, uint16(0), version)
	assert.EqualError(t, err, "supplied TLS version isn't supported")

	version, err = TLSStringToTLSConfigVersion("SSL3.0")
	assert.Error(t, err)
	assert.Equal(t, uint16(0), version)
	assert.EqualError(t, err, "supplied TLS version isn't supported")
}
