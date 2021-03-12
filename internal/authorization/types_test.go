package authorization

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldAppendQueryParamToURL(t *testing.T) {
	targetURL, err := url.Parse("https://domain.example.com/api?type=none")

	require.NoError(t, err)

	object := NewObject(targetURL, "GET")

	assert.Equal(t, "domain.example.com", object.Domain)
	assert.Equal(t, "GET", object.Method)
	assert.Equal(t, "/api?type=none", object.Path)
	assert.Equal(t, "https", object.Scheme)
}
