package main

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiscOIDCConformanceBuildSuites(t *testing.T) {
	testCases := []struct {
		name     string
		have     []string
		expected []string
	}{
		{
			"ShouldHandleDefault",
			nil,
			[]string{"conformance-config", "conformance-basic", "conformance-basic-form-post", "conformance-hybrid", "conformance-hybrid-form-post", "conformance-implicit", "conformance-implicit-form-post"},
		},
		{
			"ShouldHandleSingle",
			[]string{"config"},
			[]string{"conformance-config"},
		},
		{
			"ShouldHandleNone",
			[]string{"none"},
			nil,
		},
	}

	suiteURL := &url.URL{Scheme: "https", Host: "localhost"}
	autheliaURL := &url.URL{Scheme: "https", Host: "auth.example.com"}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := miscOIDCConformanceBuildSuites("authelia", "4.00", "implicit", "one_factor", suiteURL, autheliaURL, tc.have...)

			require.Len(t, actual, len(tc.expected))

			for i, expected := range tc.expected {
				assert.Equal(t, expected, actual[i].Name)
			}
		})
	}
}
