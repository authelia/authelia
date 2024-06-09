package webauthn

import (
	"path/filepath"

	"github.com/go-webauthn/webauthn/metadata"
	"github.com/go-webauthn/webauthn/metadata/providers/cached"
	"github.com/go-webauthn/webauthn/metadata/providers/memory"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewMetaDataProvider generates a new metadata.Provider given a *schema.Configuration.
func NewMetaDataProvider(config *schema.Configuration) (provider metadata.Provider, err error) {
	if !config.WebAuthn.Metadata.Enable {
		return nil, nil
	}

	return cached.New(
		cached.WithPath(filepath.Join(config.CacheDirectory, config.WebAuthn.Metadata.Path)),
		cached.WithNew(newMetadataProviderMemory(config)),
	)
}

func newMetadataProviderMemory(config *schema.Configuration) cached.NewFunc {
	return func(mds *metadata.Metadata) (provider metadata.Provider, err error) {
		return memory.New(
			memory.WithMetadata(mds.ToMap()),
			memory.WithValidateEntry(config.WebAuthn.Metadata.ValidateEntry),
			memory.WithValidateEntryPermitZeroAAGUID(config.WebAuthn.Metadata.ValidateEntryPermitZeroAAGUID),
			memory.WithValidateTrustAnchor(config.WebAuthn.Metadata.ValidateTrustAnchor),
			memory.WithValidateStatus(config.WebAuthn.Metadata.ValidateStatus),
			memory.WithStatusUndesired(config.WebAuthn.Metadata.ValidateStatusProhibited),
			memory.WithStatusDesired(config.WebAuthn.Metadata.ValidateStatusPermitted),
		)
	}
}
