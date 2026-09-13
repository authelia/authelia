// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

// UserExternalIdentityLinksGET returns the current user's external account links and any pending proposal.
func UserExternalIdentityLinksGET(ctx *middlewares.AutheliaCtx) {
	userSession, err := ctx.GetSession()
	if err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred loading external identity links: %s", errStrUserSessionData)

		ctx.SetJSONError(messageOperationFailed)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred loading external identity links")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	links, err := ctx.Providers.StorageProvider.LoadExternalIdentityLinksByUsername(ctx, userSession.Username)
	if err != nil && !errors.Is(err, storage.ErrNoExternalIdentityLink) {
		ctx.Logger.WithError(err).WithField("username", userSession.Username).Error("Error occurred loading external identity links")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	body := bodyGETExternalIdentityLinks{Links: make([]bodyExternalIdentityLink, 0, len(links))}

	for _, link := range links {
		item := bodyExternalIdentityLink{
			ID:             link.ID,
			CreatedAt:      link.CreatedAt,
			Provider:       link.Provider,
			ProviderName:   externalIdentityProviderName(ctx, link.Provider),
			Issuer:         link.Issuer,
			Subject:        link.Subject,
			RemoteUsername: link.RemoteUsername.String,
		}

		if link.LastUsedAt.Valid {
			item.LastUsedAt = &link.LastUsedAt.Time
		}

		body.Links = append(body.Links, item)
	}

	if pending := externalIdentityPending(ctx, &userSession); pending != nil {
		body.Pending = &bodyExternalIdentityPending{
			Provider:       pending.Provider,
			ProviderName:   externalIdentityProviderName(ctx, pending.Provider),
			Issuer:         pending.Issuer,
			Subject:        pending.Subject,
			RemoteUsername: pending.RemoteUsername,
			DisplayName:    pending.DisplayName,
			Email:          pending.Email,
		}
	}

	if err = ctx.SetJSONBody(body); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred loading external identity links: %s", errStrRespBody)
	}
}

// UserExternalIdentityLinkPUT accepts the pending external account link. This endpoint requires an elevated session as
// it is the security boundary preventing a planted pending proposal (see handler_firstfactor_external_identity_callback.go)
// from being used to link an attacker controlled external account to the victim's local account.
func UserExternalIdentityLinkPUT(ctx *middlewares.AutheliaCtx) {
	userSession, err := ctx.GetSession()
	if err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred linking an external identity account: %s", errStrUserSessionData)

		ctx.SetJSONError(messageOperationFailed)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	pending := externalIdentityPending(ctx, &userSession)

	if pending == nil {
		if err = ctx.SaveSession(userSession); err != nil {
			ctx.Logger.WithError(err).Errorf("Error occurred linking an external identity account: %s", errStrUserSessionDataSave)
		}

		ctx.SetJSONError(messageExternalIdentityLinkNonePending)

		return
	}

	link := model.ExternalIdentityLink{
		CreatedAt: ctx.GetClock().Now(),
		Type:      pending.Type,
		Provider:  pending.Provider,
		Issuer:    pending.Issuer,
		Subject:   pending.Subject,
		Username:  userSession.Username,
	}

	if pending.RemoteUsername != "" {
		link.RemoteUsername = sql.NullString{String: pending.RemoteUsername, Valid: true}
	}

	if pending.Email != "" {
		link.Email = sql.NullString{String: pending.Email, Valid: true}
	}

	if err = ctx.Providers.StorageProvider.SaveExternalIdentityLink(ctx, link); err != nil {
		ctx.Logger.WithError(err).WithField("username", userSession.Username).Error("Error occurred linking an external identity account")

		if isExternalIdentityLinkConflict(err) {
			ctx.SetJSONError(messageExternalIdentityLinkConflict)
		} else {
			ctx.SetJSONError(messageExternalIdentityLinkFailed)
		}

		return
	}

	userSession.ExternalIdentityPending = nil

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred linking an external identity account: %s", errStrUserSessionDataSave)

		ctx.SetJSONError(messageExternalIdentityLinkFailed)

		return
	}

	ctx.ReplyOK()
}

// UserExternalIdentityLinkPendingDELETE declines the pending external account link.
func UserExternalIdentityLinkPendingDELETE(ctx *middlewares.AutheliaCtx) {
	userSession, err := ctx.GetSession()
	if err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred declining an external identity account: %s", errStrUserSessionData)

		ctx.SetJSONError(messageOperationFailed)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	userSession.ExternalIdentityPending = nil

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred declining an external identity account: %s", errStrUserSessionDataSave)

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	ctx.ReplyOK()
}

// UserExternalIdentityLinkDELETE deletes one of the current user's external account links. This endpoint requires an
// elevated session for the same reason UserExternalIdentityLinkPUT does; the deletion is scoped to the authenticated
// user's username so one user cannot delete another user's link by guessing its id.
func UserExternalIdentityLinkDELETE(ctx *middlewares.AutheliaCtx) {
	userSession, err := ctx.GetSession()
	if err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred deleting an external identity link: %s", errStrUserSessionData)

		ctx.SetJSONError(messageOperationFailed)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred deleting an external identity link")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	value, ok := ctx.UserValue("linkID").(string)
	if !ok {
		ctx.Logger.Error("Error occurred deleting an external identity link: the linkID user value wasn't set")

		ctx.SetJSONError(messageExternalIdentityUnlinkFailed)

		return
	}

	id, err := strconv.Atoi(value)
	if err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred deleting an external identity link: failed to parse '%s' as an integer", value)

		ctx.SetJSONError(messageExternalIdentityUnlinkFailed)

		return
	}

	// The delete is scoped to the authenticated user's username at the storage layer so a user cannot delete another
	// user's link by guessing its id.
	if err = ctx.Providers.StorageProvider.DeleteExternalIdentityLink(ctx, userSession.Username, id); err != nil {
		ctx.Logger.WithError(err).WithField("username", userSession.Username).Error("Error occurred deleting an external identity link")

		ctx.SetJSONError(messageExternalIdentityUnlinkFailed)

		return
	}

	ctx.ReplyOK()
}

func externalIdentityProviderName(ctx *middlewares.AutheliaCtx, id string) (name string) {
	if provider, ok := ctx.Providers.ExternalIdentity.Get(id); ok {
		return provider.Name()
	}

	return id
}

func externalIdentityPending(ctx *middlewares.AutheliaCtx, userSession *session.UserSession) (pending *session.ExternalIdentityPending) {
	if userSession.ExternalIdentityPending == nil {
		return nil
	}

	if ctx.GetClock().Now().After(userSession.ExternalIdentityPending.Expires) {
		userSession.ExternalIdentityPending = nil

		return nil
	}

	return userSession.ExternalIdentityPending
}

func isExternalIdentityLinkConflict(err error) bool {
	e := strings.ToLower(err.Error())

	return strings.Contains(e, "unique") || strings.Contains(e, "duplicate")
}
