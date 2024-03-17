package oidc

import (
	"context"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// ClientSecretDigest decorates the *schema.PasswordDigest with the relevant functions to implement oauth2.ClientSecret.
type ClientSecretDigest struct {
	*schema.PasswordDigest
}

// Compare decorates the *schema.PasswordDigest's implementation to satisfy oauth2.ClientSecret's Compare function.
func (d *ClientSecretDigest) Compare(ctx context.Context, rawSecret []byte) (err error) {
	if d.PasswordDigest == nil || d.PasswordDigest.Digest == nil {
		return oauthelia2.ErrClientSecretNotRegistered
	}

	if d.MatchBytes(rawSecret) {
		return nil
	}

	return errClientSecretMismatch
}
