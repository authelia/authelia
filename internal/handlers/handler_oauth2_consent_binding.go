// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

var errOAuth2ConsentSessionNotBound = errors.New("the consent session is not bound to the user agent")

func handleOAuth2ConsentSessionBind(ctx *middlewares.AutheliaCtx, consent *model.OAuth2ConsentSession) (err error) {
	var provider session.Strategy

	if provider, err = ctx.GetSessionProvider(); err != nil {
		return fmt.Errorf("error occurred obtaining the session provider to bind the consent session: %w", err)
	}

	if consent.Binding, err = provider.AuthorizationBinding(ctx); err != nil {
		return fmt.Errorf("error occurred binding the consent session: %w", err)
	}

	return nil
}

func handleOAuth2ConsentSessionLoadBound(ctx *middlewares.AutheliaCtx, challengeID uuid.UUID) (consent *model.OAuth2ConsentSession, err error) {
	var provider session.Strategy

	if provider, err = ctx.GetSessionProvider(); err != nil {
		return nil, fmt.Errorf("error occurred obtaining the session provider to verify the consent session binding: %w", err)
	}

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, challengeID); err != nil {
		return nil, err
	}

	if !provider.VerifyAuthorizationBinding(ctx, consent.Binding) {
		return nil, errOAuth2ConsentSessionNotBound
	}

	return consent, nil
}
