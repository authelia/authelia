// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	oauthelia2 "authelia.com/provider/oauth2"

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

// UserDetailerFromSubject returns a [UserDetailer] using the subject [uuid.UUID] value.
func UserDetailerFromSubject(ctx Context, subject uuid.UUID) (detailer UserDetailer, err error) {
	var (
		identifier *model.UserOpaqueIdentifier
		details    *authentication.UserDetailsExtended
	)

	if identifier, err = ctx.GetProviderStorage().LoadUserOpaqueIdentifier(ctx, subject); err != nil {
		return nil, err
	}

	if details, err = getUserDetailsExtended(ctx, ctx.GetUserProvider(), identifier.Username); err != nil {
		return nil, err
	}

	return details, nil
}

type ctxKeyUserDetailsExtended struct{}

type userValueContext interface {
	UserValue(key any) any
	SetUserValue(key, value any)
}

func getUserDetailsExtended(ctx context.Context, users authentication.UserProvider, username string) (details *authentication.UserDetailsExtended, err error) {
	uctx, ok := ctx.(userValueContext)
	if !ok {
		return users.GetDetailsExtended(username)
	}

	memo, _ := uctx.UserValue(ctxKeyUserDetailsExtended{}).(map[string]*authentication.UserDetailsExtended)

	if details, ok = memo[username]; ok {
		return details, nil
	}

	if details, err = users.GetDetailsExtended(username); err != nil {
		return nil, err
	}

	if memo == nil {
		memo = map[string]*authentication.UserDetailsExtended{}

		uctx.SetUserValue(ctxKeyUserDetailsExtended{}, memo)
	}

	memo[username] = details

	if details != nil && details.UserDetails != nil {
		memo[details.Username] = details
	}

	return details, nil
}

// SubjectUUIDFromClaims returns the subject [uuid.UUID] from a claims map.
func SubjectUUIDFromClaims(claims map[string]any) (subject uuid.UUID, err error) {
	var (
		ok    bool
		raw   any
		claim string
	)

	if raw, ok = claims[ClaimSubject]; !ok {
		return uuid.UUID{}, oauthelia2.ErrServerError.WithDebug("Failed to find claim 'sub' in the original claims.")
	}

	if claim, ok = raw.(string); !ok {
		return uuid.UUID{}, oauthelia2.ErrServerError.WithDebug("Failed to parse claim 'sub' as a string.")
	}

	return SubjectUUIDFromSubjectString(claim)
}

// SubjectUUIDFromSubjectString returns the subject [uuid.UUID] from a raw string value.
func SubjectUUIDFromSubjectString(value string) (subject uuid.UUID, err error) {
	if subject, err = uuid.Parse(value); err != nil {
		return uuid.UUID{}, oauthelia2.ErrServerError.WithDebug(fmt.Sprintf("Failed to parse subject '%s' as a UUID.", value))
	}

	return subject, nil
}
