package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSemanticVersion(t *testing.T) {
	testCases := []struct {
		desc     string
		have     string
		expected *SemanticVersion
		err      string
	}{
		{
			desc:     "ShouldParseStandardSemVer",
			have:     "4.30.0",
			expected: &SemanticVersion{Major: 4, Minor: 30, Patch: 0},
		},
		{
			desc:     "ShouldParseSemVerWithPre",
			have:     "4.30.0-alpha1",
			expected: &SemanticVersion{Major: 4, Minor: 30, Patch: 0, PreRelease: []string{"alpha1"}},
		},
		{
			desc:     "ShouldParseSemVerWithMeta",
			have:     "4.30.0+build4",
			expected: &SemanticVersion{Major: 4, Minor: 30, Patch: 0, Metadata: []string{"build4"}},
		},
		{
			desc:     "ShouldParseSemVerWithPreAndMeta",
			have:     "4.30.0-alpha1+build4",
			expected: &SemanticVersion{Major: 4, Minor: 30, Patch: 0, PreRelease: []string{"alpha1"}, Metadata: []string{"build4"}},
		},
		{
			desc:     "ShouldParseSemVerWithPreAndMetaMulti",
			have:     "4.30.0-alpha1.test+build4.new",
			expected: &SemanticVersion{Major: 4, Minor: 30, Patch: 0, PreRelease: []string{"alpha1", "test"}, Metadata: []string{"build4", "new"}},
		},
		{
			desc:     "ShouldNotParseInvalidVersion",
			have:     "1.2",
			expected: nil,
			err:      "the input '1.2' failed to match the semantic version pattern",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			version, err := NewSemanticVersion(tc.have)

			if tc.err == "" {
				assert.Nil(t, err)
				require.NotNil(t, version)
				assert.Equal(t, tc.expected, version)
				assert.Equal(t, tc.have, version.String())
			} else {
				assert.Nil(t, version)
				require.NotNil(t, err)
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestSemanticVersionComparisons(t *testing.T) {
	testCases := []struct {
		desc string

		haveFirst, haveSecond SemanticVersion

		expectedEQ, expectedGT, expectedGE, expectedLT, expectedLE, expectedStable, expectedAbsolute bool
	}{
		{
			desc:             "ShouldCompareVersionLessThanMajor",
			haveFirst:        SemanticVersion{Major: 4, Minor: 30, Patch: 0},
			haveSecond:       SemanticVersion{Major: 5, Minor: 3, Patch: 0},
			expectedEQ:       false,
			expectedGT:       false,
			expectedGE:       false,
			expectedLT:       true,
			expectedLE:       true,
			expectedStable:   true,
			expectedAbsolute: true,
		},
		{
			desc:             "ShouldCompareVersionLessThanMinor",
			haveFirst:        SemanticVersion{Major: 4, Minor: 30, Patch: 0},
			haveSecond:       SemanticVersion{Major: 4, Minor: 31, Patch: 0},
			expectedEQ:       false,
			expectedGT:       false,
			expectedGE:       false,
			expectedLT:       true,
			expectedLE:       true,
			expectedStable:   true,
			expectedAbsolute: true,
		},
		{
			desc:             "ShouldCompareVersionLessThanPatch",
			haveFirst:        SemanticVersion{Major: 4, Minor: 31, Patch: 0},
			haveSecond:       SemanticVersion{Major: 4, Minor: 31, Patch: 9},
			expectedEQ:       false,
			expectedGT:       false,
			expectedGE:       false,
			expectedLT:       true,
			expectedLE:       true,
			expectedStable:   true,
			expectedAbsolute: true,
		},
		{
			desc:             "ShouldCompareVersionEqual",
			haveFirst:        SemanticVersion{Major: 4, Minor: 31, Patch: 0},
			haveSecond:       SemanticVersion{Major: 4, Minor: 31, Patch: 0},
			expectedEQ:       true,
			expectedGT:       false,
			expectedGE:       true,
			expectedLT:       false,
			expectedLE:       true,
			expectedStable:   true,
			expectedAbsolute: true,
		},
		{
			desc:             "ShouldCompareVersionEqualBeta",
			haveFirst:        SemanticVersion{Major: 0, Minor: 31, Patch: 0},
			haveSecond:       SemanticVersion{Major: 0, Minor: 31, Patch: 0},
			expectedEQ:       true,
			expectedGT:       false,
			expectedGE:       true,
			expectedLT:       false,
			expectedLE:       true,
			expectedStable:   false,
			expectedAbsolute: true,
		},
		{
			desc:             "ShouldCompareVersionEqualPre",
			haveFirst:        SemanticVersion{Major: 0, Minor: 31, Patch: 0, PreRelease: []string{"beta-1"}},
			haveSecond:       SemanticVersion{Major: 0, Minor: 31, Patch: 0, PreRelease: []string{"beta-1"}},
			expectedEQ:       true,
			expectedGT:       false,
			expectedGE:       true,
			expectedLT:       false,
			expectedLE:       true,
			expectedStable:   false,
			expectedAbsolute: false,
		},
		{
			desc:             "ShouldCompareVersionEqualPre",
			haveFirst:        SemanticVersion{Major: 0, Minor: 31, Patch: 0, Metadata: []string{"beta-1"}},
			haveSecond:       SemanticVersion{Major: 0, Minor: 31, Patch: 0, Metadata: []string{"beta-1"}},
			expectedEQ:       true,
			expectedGT:       false,
			expectedGE:       true,
			expectedLT:       false,
			expectedLE:       true,
			expectedStable:   false,
			expectedAbsolute: false,
		},
		{
			desc:             "ShouldCompareVersionGreaterThanMajor",
			haveFirst:        SemanticVersion{Major: 5, Minor: 0, Patch: 0},
			haveSecond:       SemanticVersion{Major: 4, Minor: 30, Patch: 0},
			expectedEQ:       false,
			expectedGT:       true,
			expectedGE:       true,
			expectedLT:       false,
			expectedLE:       false,
			expectedStable:   true,
			expectedAbsolute: true,
		},
		{
			desc:             "ShouldCompareVersionGreaterThanMinor",
			haveFirst:        SemanticVersion{Major: 4, Minor: 31, Patch: 0},
			haveSecond:       SemanticVersion{Major: 4, Minor: 30, Patch: 0},
			expectedEQ:       false,
			expectedGT:       true,
			expectedGE:       true,
			expectedLT:       false,
			expectedLE:       false,
			expectedStable:   true,
			expectedAbsolute: true,
		},
		{
			desc:             "ShouldCompareVersionGreaterThanPatch",
			haveFirst:        SemanticVersion{Major: 4, Minor: 31, Patch: 5},
			haveSecond:       SemanticVersion{Major: 4, Minor: 31, Patch: 0},
			expectedEQ:       false,
			expectedGT:       true,
			expectedGE:       true,
			expectedLT:       false,
			expectedLE:       false,
			expectedStable:   true,
			expectedAbsolute: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			assert.Equal(t, tc.expectedEQ, tc.haveFirst.Equal(tc.haveSecond))
			assert.Equal(t, tc.expectedGT, tc.haveFirst.GreaterThan(tc.haveSecond))
			assert.Equal(t, tc.expectedGE, tc.haveFirst.GreaterThanOrEqual(tc.haveSecond))
			assert.Equal(t, tc.expectedLT, tc.haveFirst.LessThan(tc.haveSecond))
			assert.Equal(t, tc.expectedLE, tc.haveFirst.LessThanOrEqual(tc.haveSecond))

			assert.Equal(t, tc.expectedStable, tc.haveFirst.IsStable())
			assert.Equal(t, tc.expectedAbsolute, tc.haveFirst.IsAbsolute())

			assert.True(t, tc.haveFirst.Equal(tc.haveFirst.Copy()))
			assert.True(t, tc.haveSecond.Equal(tc.haveSecond.Copy()))
		})
	}
}

func TestSemanticVersion_Next(t *testing.T) {
	v := SemanticVersion{Major: 1, Minor: 2, Patch: 3}

	x := v.NextMajor()

	assert.Equal(t, 2, x.Major)
	assert.Equal(t, 0, x.Minor)
	assert.Equal(t, 0, x.Patch)
	assert.Equal(t, "2.0.0", x.String())

	x = v.NextMinor()

	assert.Equal(t, 1, x.Major)
	assert.Equal(t, 3, x.Minor)
	assert.Equal(t, 0, x.Patch)
	assert.Equal(t, "1.3.0", x.String())

	x = v.NextPatch()

	assert.Equal(t, 1, x.Major)
	assert.Equal(t, 2, x.Minor)
	assert.Equal(t, 4, x.Patch)
	assert.Equal(t, "1.2.4", x.String())
}
