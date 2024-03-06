package oidc

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestClientCredentialsFromBasicAuth(t *testing.T) {
	testCases := []struct {
		name       string
		have       http.Header
		id, secret string
		ok         bool
		err        string
	}{
		{
			"ShouldParseCredentials",
			http.Header{
				fasthttp.HeaderAuthorization: []string{"Basic YWJjOjEyMw=="},
			},
			"abc",
			"123",
			true,
			"",
		},
		{
			"ShouldFailParseCredentialsClientIDNoEscape",
			http.Header{
				fasthttp.HeaderAuthorization: []string{"Basic YSV6YmM6MTIz"},
			},
			"",
			"",
			false,
			"failed to query unescape client id from http authorization header: invalid URL escape \"%zb\"",
		},
		{
			"ShouldFailParseCredentialsClientSecretNoEscape",
			http.Header{
				fasthttp.HeaderAuthorization: []string{"Basic YWJjOjEyMyV6YQ=="},
			},
			"",
			"",
			false,
			"failed to query unescape client secret from http authorization header: invalid URL escape \"%za\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id, secret, ok, err := clientCredentialsFromBasicAuth(tc.have)

			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err)
			}

			assert.Equal(t, tc.id, id)
			assert.Equal(t, tc.secret, secret)
			assert.Equal(t, tc.ok, ok)
		})
	}
}
