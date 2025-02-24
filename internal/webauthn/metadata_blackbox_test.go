package webauthn_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	. "github.com/authelia/authelia/v4/internal/webauthn"
)

func TestNewMetaDataProvider(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	store := mocks.NewMockStorage(ctrl)
	provider, err := NewMetaDataProvider(&schema.Configuration{}, store)
	assert.NoError(t, err)
	assert.Nil(t, provider)

	provider, err = NewMetaDataProvider(&schema.Configuration{WebAuthn: schema.WebAuthn{Metadata: schema.WebAuthnMetadata{Enabled: true}}}, store)

	assert.NoError(t, err)
	assert.NotNil(t, provider)
}
