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

func handleOAuth2ConsentSessionBind(ctx *middlewares.AutheliaCtx, consent *model.OAuth2ConsentSession) (binding string, err error) {
	var provider session.Strategy

	if provider, err = ctx.GetSessionProvider(); err != nil {
		return "", fmt.Errorf("error occurred obtaining the session provider to bind the consent session: %w", err)
	}

	if binding, err = provider.UserAgentBinding(ctx); err != nil {
		return "", fmt.Errorf("error occurred binding the consent session: %w", err)
	}

	consent.Binding = session.UserAgentBindingDigest(binding)

	return binding, nil
}

func handleOAuth2ConsentSessionIssueBinding(ctx *middlewares.AutheliaCtx, binding string) (err error) {
	var provider session.Strategy

	if provider, err = ctx.GetSessionProvider(); err != nil {
		return fmt.Errorf("error occurred obtaining the session provider to issue the consent session binding: %w", err)
	}

	provider.IssueUserAgentBinding(ctx, binding)

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

	if !provider.VerifyUserAgentBinding(ctx, consent.Binding) {
		return nil, errOAuth2ConsentSessionNotBound
	}

	return consent, nil
}

func handleOAuth2ConsentSessionLoadBoundOrClaim(ctx *middlewares.AutheliaCtx, challengeID uuid.UUID) (consent *model.OAuth2ConsentSession, err error) {
	var provider session.Strategy

	if provider, err = ctx.GetSessionProvider(); err != nil {
		return nil, fmt.Errorf("error occurred obtaining the session provider to verify the consent session binding: %w", err)
	}

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, challengeID); err != nil {
		return nil, err
	}

	if len(consent.Binding) != 0 {
		if !provider.VerifyUserAgentBinding(ctx, consent.Binding) {
			return nil, errOAuth2ConsentSessionNotBound
		}

		return consent, nil
	}

	var binding string

	if binding, err = handleOAuth2ConsentSessionBind(ctx, consent); err != nil {
		return nil, err
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionBinding(ctx, consent); err != nil {
		return nil, fmt.Errorf("error occurred claiming the consent session: %w", err)
	}

	provider.IssueUserAgentBinding(ctx, binding)

	return consent, nil
}
