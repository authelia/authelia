package handlers

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func oidcDetailerFromClaims(ctx *middlewares.AutheliaCtx, claims map[string]any) (detailer oidc.UserDetailer, err error) {
	var (
		subject    uuid.UUID
		identifier *model.UserOpaqueIdentifier
		details    *authentication.UserDetails
	)

	if subject, err = oidcSubjectUUIDFromClaims(claims); err != nil {
		return nil, err
	}

	if identifier, err = ctx.Providers.StorageProvider.LoadUserOpaqueIdentifier(ctx, subject); err != nil {
		return nil, err
	}

	if details, err = ctx.Providers.UserProvider.GetDetails(identifier.Username); err != nil {
		return nil, err
	}

	return details, nil
}

func oidcSubjectUUIDFromClaims(claims map[string]any) (subject uuid.UUID, err error) {
	var (
		ok    bool
		raw   any
		claim string
	)

	if raw, ok = claims[oidc.ClaimSubject]; !ok {
		return uuid.UUID{}, fmt.Errorf("error retrieving claim 'sub' from the original claims")
	}

	if claim, ok = raw.(string); !ok {
		return uuid.UUID{}, fmt.Errorf("error asserting claim 'sub' as a string from the original claims")
	}

	if subject, err = uuid.Parse(claim); err != nil {
		return uuid.UUID{}, fmt.Errorf("error parsing claim 'sub' as a UUIDv4 from the original claims: %w", err)
	}

	return subject, nil
}
