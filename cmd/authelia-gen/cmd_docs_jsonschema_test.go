// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/model"
)

func TestJSONSchemaVersionTarget(t *testing.T) {
	version := &model.SemanticVersion{Major: 4, Minor: 39, Patch: 25}

	testCases := []struct {
		name     string
		have     []string
		expected string
		err      string
	}{
		{
			"ShouldTargetCurrentByDefault",
			[]string{metaVersionLatest, metaVersionCurrent},
			"v4.39",
			"",
		},
		{
			"ShouldTargetNextMinor",
			[]string{metaVersionLatest, metaVersionMinor},
			"v4.40",
			"",
		},
		{
			"ShouldTargetNextMajor",
			[]string{metaVersionLatest, metaVersionMajor},
			"v5.0",
			"",
		},
		{
			"ShouldTargetCurrentForExplicitVersions",
			[]string{"v4.38"},
			"v4.39",
			"",
		},
		{
			"ShouldErrorOnMajorAndMinor",
			[]string{metaVersionMajor, metaVersionMinor},
			"",
			"failed to generate: meta versions major and minor are mutually exclusive",
		},
		{
			"ShouldErrorOnMajorAndCurrent",
			[]string{metaVersionMajor, metaVersionCurrent},
			"",
			"failed to generate: meta version current is mutually exclusive with major and minor",
		},
		{
			"ShouldErrorOnMinorAndCurrent",
			[]string{metaVersionMinor, metaVersionCurrent},
			"",
			"failed to generate: meta version current is mutually exclusive with major and minor",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := jsonSchemaVersionTarget(version, tc.have)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, jsonSchemaVersionDir(actual))
		})
	}
}
