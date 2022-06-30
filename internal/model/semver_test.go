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

		expectedEQ, expectedGT, expectedGE, expectedLT, expectedLE bool
	}{
		{
			desc:       "ShouldCompareVersionLessThanMajor",
			haveFirst:  SemanticVersion{Major: 4, Minor: 30, Patch: 0},
			haveSecond: SemanticVersion{Major: 5, Minor: 3, Patch: 0},
			expectedEQ: false,
			expectedGT: false,
			expectedGE: false,
			expectedLT: true,
			expectedLE: true,
		},
		{
			desc:       "ShouldCompareVersionLessThanMinor",
			haveFirst:  SemanticVersion{Major: 4, Minor: 30, Patch: 0},
			haveSecond: SemanticVersion{Major: 4, Minor: 31, Patch: 0},
			expectedEQ: false,
			expectedGT: false,
			expectedGE: false,
			expectedLT: true,
			expectedLE: true,
		},
		{
			desc:       "ShouldCompareVersionLessThanPatch",
			haveFirst:  SemanticVersion{Major: 4, Minor: 31, Patch: 0},
			haveSecond: SemanticVersion{Major: 4, Minor: 31, Patch: 9},
			expectedEQ: false,
			expectedGT: false,
			expectedGE: false,
			expectedLT: true,
			expectedLE: true,
		},
		{
			desc:       "ShouldCompareVersionEqual",
			haveFirst:  SemanticVersion{Major: 4, Minor: 31, Patch: 0},
			haveSecond: SemanticVersion{Major: 4, Minor: 31, Patch: 0},
			expectedEQ: true,
			expectedGT: false,
			expectedGE: true,
			expectedLT: false,
			expectedLE: true,
		},
		{
			desc:       "ShouldCompareVersionGreaterThanMajor",
			haveFirst:  SemanticVersion{Major: 5, Minor: 0, Patch: 0},
			haveSecond: SemanticVersion{Major: 4, Minor: 30, Patch: 0},
			expectedEQ: false,
			expectedGT: true,
			expectedGE: true,
			expectedLT: false,
			expectedLE: false,
		},
		{
			desc:       "ShouldCompareVersionGreaterThanMinor",
			haveFirst:  SemanticVersion{Major: 4, Minor: 31, Patch: 0},
			haveSecond: SemanticVersion{Major: 4, Minor: 30, Patch: 0},
			expectedEQ: false,
			expectedGT: true,
			expectedGE: true,
			expectedLT: false,
			expectedLE: false,
		},
		{
			desc:       "ShouldCompareVersionGreaterThanPatch",
			haveFirst:  SemanticVersion{Major: 4, Minor: 31, Patch: 5},
			haveSecond: SemanticVersion{Major: 4, Minor: 31, Patch: 0},
			expectedEQ: false,
			expectedGT: true,
			expectedGE: true,
			expectedLT: false,
			expectedLE: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			assert.Equal(t, tc.expectedEQ, tc.haveFirst.Equal(tc.haveSecond))
			assert.Equal(t, tc.expectedGT, tc.haveFirst.GreaterThan(tc.haveSecond))
			assert.Equal(t, tc.expectedGE, tc.haveFirst.GreaterThanOrEqual(tc.haveSecond))
			assert.Equal(t, tc.expectedLT, tc.haveFirst.LessThan(tc.haveSecond))
			assert.Equal(t, tc.expectedLE, tc.haveFirst.LessThanOrEqual(tc.haveSecond))
		})
	}
}
