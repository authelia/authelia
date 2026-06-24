package schema

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebAuthnRelatedOriginStringOrigins(t *testing.T) {
	testCases := []struct {
		name     string
		have     WebAuthnRelatedOrigin
		expected []string
	}{
		{
			"ShouldReturnEmptySlice",
			WebAuthnRelatedOrigin{
				Origins: []*url.URL{},
			},
			[]string{},
		},
		{
			"ShouldReturnOrigins",
			WebAuthnRelatedOrigin{
				Origins: []*url.URL{
					{Scheme: "https", Host: "example.com"},
					{Scheme: "https", Host: "auth.example.com"},
				},
			},
			[]string{"https://example.com", "https://auth.example.com"},
		},
		{
			"ShouldHandleNilOriginEntry",
			WebAuthnRelatedOrigin{
				Origins: []*url.URL{
					{Scheme: "https", Host: "example.com"},
					nil,
					{Scheme: "https", Host: "auth.example.com"},
				},
			},
			[]string{"https://example.com", "", "https://auth.example.com"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.StringOrigins())
		})
	}
}
