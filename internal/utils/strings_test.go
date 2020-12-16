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

func TestShouldFindStringInSliceContains(t *testing.T) {
	a := "abc"
	b := []string{"abc", "onetwothree"}
	s := IsStringInSliceContains(a, b)
	assert.True(t, s)
}

func TestShouldNotFindStringInSliceContains(t *testing.T) {
	a := "xyz"
	b := []string{"abc", "onetwothree"}
	s := IsStringInSliceContains(a, b)
	assert.False(t, s)
}

func TestShouldReturnCorrectTLSVersions(t *testing.T) {
	tls13 := uint16(tls.VersionTLS13)
	tls12 := uint16(tls.VersionTLS12)
	tls11 := uint16(tls.VersionTLS11)
	tls10 := uint16(tls.VersionTLS10)

	version, err := TLSStringToTLSConfigVersion(TLS13)
	assert.Equal(t, tls13, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion("TLS" + TLS13)
	assert.Equal(t, tls13, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion(TLS12)
	assert.Equal(t, tls12, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion("TLS" + TLS12)
	assert.Equal(t, tls12, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion(TLS11)
	assert.Equal(t, tls11, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion("TLS" + TLS11)
	assert.Equal(t, tls11, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion(TLS10)
	assert.Equal(t, tls10, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion("TLS" + TLS10)
	assert.Equal(t, tls10, version)
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
