package webauthn

import (
	"testing"

	"github.com/go-webauthn/webauthn/metadata"
	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestNewMetadataProviderMemory(t *testing.T) {
	generator := newMetadataProviderMemory(&schema.Configuration{})

	assert.NotNil(t, generator)

	provider, err := generator(&metadata.Metadata{})
	assert.NoError(t, err)
	assert.NotNil(t, provider)
}
