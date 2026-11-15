// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/jsonschema"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

func TestJSONSchemaAddSchemaProperty(t *testing.T) {
	testCases := []struct {
		name string
		have any
	}{
		{"ShouldAddToConfiguration", &schema.Configuration{}},
		{"ShouldAddToUserDatabase", &authentication.FileUserDatabase{}},
		{"ShouldAddToTOTPExport", &model.TOTPConfigurationDataExport{}},
		{"ShouldAddToWebAuthnExport", &model.WebAuthnCredentialDataExport{}},
		{"ShouldAddToIdentifiersExport", &model.UserOpaqueIdentifiersExport{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := (&jsonschema.Reflector{RequiredFromJSONSchemaTags: true}).Reflect(tc.have)

			jsonSchemaAddSchemaProperty(s)

			require.NotEmpty(t, s.Ref)

			root := s.Definitions[s.Ref[len("#/$defs/"):]]

			require.NotNil(t, root)
			require.NotNil(t, root.Properties)

			value, ok := root.Properties.Get("$schema")
			require.True(t, ok)

			property, ok := value.(*jsonschema.Schema)
			require.True(t, ok)

			assert.Equal(t, "string", property.Type)
			assert.Equal(t, "uri", property.Format)
		})
	}

	t.Run("ShouldIgnoreSchemaWithoutProperties", func(t *testing.T) {
		s := &jsonschema.Schema{Type: "string"}

		jsonSchemaAddSchemaProperty(s)

		assert.Nil(t, s.Properties)
	})
}

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
