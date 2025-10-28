package oidc

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/model"
)

// UserDetailerFromClaims returns a UserDetailer using the claims map.
func UserDetailerFromClaims(ctx Context, claims map[string]any) (detailer UserDetailer, err error) {
	var subject uuid.UUID
	if subject, err = SubjectUUIDFromClaims(claims); err != nil {
		return nil, err
	}

	return UserDetailerFromSubject(ctx, subject)
}

// UserDetailerFromSubjectString returns a UserDetailer using the subject string value.
func UserDetailerFromSubjectString(ctx Context, subjectRaw string) (detailer UserDetailer, err error) {
	var subject uuid.UUID
	if subject, err = SubjectUUIDFromSubjectString(subjectRaw); err != nil {
		return nil, err
	}

	return UserDetailerFromSubject(ctx, subject)
}

// UserDetailerFromSubject returns a UserDetailer using the subject uuid.UUID value.
func UserDetailerFromSubject(ctx Context, subject uuid.UUID) (detailer UserDetailer, err error) {
	var (
		identifier *model.UserOpaqueIdentifier
		details    *authentication.UserDetailsExtended
	)

	if identifier, err = ctx.GetProviderStorage().LoadUserOpaqueIdentifier(ctx, subject); err != nil {
		return nil, err
	}

	if details, err = ctx.GetProviderUser().GetDetailsExtended(identifier.Username); err != nil {
		return nil, err
	}

	return details, nil
}

// SubjectUUIDFromClaims returns the subject uuid.UUID from a claims map.
func SubjectUUIDFromClaims(claims map[string]any) (subject uuid.UUID, err error) {
	var (
		ok    bool
		raw   any
		claim string
	)

	if raw, ok = claims[ClaimSubject]; !ok {
		return uuid.UUID{}, fmt.Errorf("error retrieving claim 'sub' from the original claims")
	}

	if claim, ok = raw.(string); !ok {
		return uuid.UUID{}, fmt.Errorf("error asserting claim 'sub' as a string from the original claims")
	}

	return SubjectUUIDFromSubjectString(claim)
}

// SubjectUUIDFromSubjectString returns the subject uuid.UUID from a raw string value.
func SubjectUUIDFromSubjectString(value string) (subject uuid.UUID, err error) {
	if subject, err = uuid.Parse(value); err != nil {
		return uuid.UUID{}, fmt.Errorf("error parsing claim 'sub' as a UUIDv4 from the original value: %w", err)
	}

	return subject, nil
}
