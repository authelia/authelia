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

// UserOpenIDConnectLinksGET returns the current user's external account links and any pending proposal.
func UserOpenIDConnectLinksGET(ctx *middlewares.AutheliaCtx) {
	userSession, err := ctx.GetSession()
	if err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred loading external OpenID Connect 1.0 links: %s", errStrUserSessionData)

		ctx.SetJSONError(messageOperationFailed)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred loading external OpenID Connect 1.0 links")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	links, err := ctx.Providers.StorageProvider.LoadOpenIDConnectLinksByUsername(ctx, userSession.Username)
	if err != nil && !errors.Is(err, storage.ErrNoOpenIDConnectLink) {
		ctx.Logger.WithError(err).WithField("username", userSession.Username).Error("Error occurred loading external OpenID Connect 1.0 links")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	body := bodyGETOpenIDConnectLinks{Links: make([]bodyOpenIDConnectLink, 0, len(links))}

	for _, link := range links {
		item := bodyOpenIDConnectLink{
			ID:             link.ID,
			CreatedAt:      link.CreatedAt,
			Provider:       link.Provider,
			ProviderName:   openIDConnectProviderName(ctx, link.Provider),
			Issuer:         link.Issuer,
			Subject:        link.Subject,
			RemoteUsername: link.RemoteUsername.String,
		}

		if link.LastUsedAt.Valid {
			item.LastUsedAt = &link.LastUsedAt.Time
		}

		body.Links = append(body.Links, item)
	}

	if pending := openIDConnectPending(ctx, &userSession); pending != nil {
		body.Pending = &bodyOpenIDConnectPending{
			Provider:       pending.Provider,
			ProviderName:   openIDConnectProviderName(ctx, pending.Provider),
			Issuer:         pending.Issuer,
			Subject:        pending.Subject,
			RemoteUsername: pending.RemoteUsername,
			DisplayName:    pending.DisplayName,
			Email:          pending.Email,
		}
	}

	if err = ctx.SetJSONBody(body); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred loading external OpenID Connect 1.0 links: %s", errStrRespBody)
	}
}

// UserOpenIDConnectLinkPUT accepts the pending external account link. This endpoint requires an elevated session as
// it is the security boundary preventing a planted pending proposal (see handler_firstfactor_openid_connect_callback.go)
// from being used to link an attacker controlled external account to the victim's local account.
func UserOpenIDConnectLinkPUT(ctx *middlewares.AutheliaCtx) {
	userSession, err := ctx.GetSession()
	if err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred linking an external OpenID Connect 1.0 account: %s", errStrUserSessionData)

		ctx.SetJSONError(messageOperationFailed)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	pending := openIDConnectPending(ctx, &userSession)

	if pending == nil {
		if err = ctx.SaveSession(userSession); err != nil {
			ctx.Logger.WithError(err).Errorf("Error occurred linking an external OpenID Connect 1.0 account: %s", errStrUserSessionDataSave)
		}

		ctx.SetJSONError(messageOpenIDConnectLinkNonePending)

		return
	}

	link := model.OpenIDConnectLink{
		CreatedAt: ctx.GetClock().Now(),
		Provider:  pending.Provider,
		Issuer:    pending.Issuer,
		Subject:   pending.Subject,
		Username:  userSession.Username,
	}

	if pending.RemoteUsername != "" {
		link.RemoteUsername = sql.NullString{String: pending.RemoteUsername, Valid: true}
	}

	if err = ctx.Providers.StorageProvider.SaveOpenIDConnectLink(ctx, link); err != nil {
		ctx.Logger.WithError(err).WithField("username", userSession.Username).Error("Error occurred linking an external OpenID Connect 1.0 account")

		if isOpenIDConnectLinkConflict(err) {
			ctx.SetJSONError(messageOpenIDConnectLinkConflict)
		} else {
			ctx.SetJSONError(messageOpenIDConnectLinkFailed)
		}

		return
	}

	userSession.OpenIDConnectPending = nil

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred linking an external OpenID Connect 1.0 account: %s", errStrUserSessionDataSave)

		ctx.SetJSONError(messageOpenIDConnectLinkFailed)

		return
	}

	ctx.ReplyOK()
}

// UserOpenIDConnectLinkPendingDELETE declines the pending external account link.
func UserOpenIDConnectLinkPendingDELETE(ctx *middlewares.AutheliaCtx) {
	userSession, err := ctx.GetSession()
	if err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred declining an external OpenID Connect 1.0 account: %s", errStrUserSessionData)

		ctx.SetJSONError(messageOperationFailed)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	userSession.OpenIDConnectPending = nil

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred declining an external OpenID Connect 1.0 account: %s", errStrUserSessionDataSave)

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	ctx.ReplyOK()
}

// UserOpenIDConnectLinkDELETE deletes one of the current user's external account links. This endpoint requires an
// elevated session for the same reason UserOpenIDConnectLinkPUT does; the deletion is scoped to the authenticated
// user's username so one user cannot delete another user's link by guessing its id.
func UserOpenIDConnectLinkDELETE(ctx *middlewares.AutheliaCtx) {
	userSession, err := ctx.GetSession()
	if err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred deleting an external OpenID Connect 1.0 link: %s", errStrUserSessionData)

		ctx.SetJSONError(messageOperationFailed)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred deleting an external OpenID Connect 1.0 link")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	value, ok := ctx.UserValue("linkID").(string)
	if !ok {
		ctx.Logger.Error("Error occurred deleting an external OpenID Connect 1.0 link: the linkID user value wasn't set")

		ctx.SetJSONError(messageOpenIDConnectUnlinkFailed)

		return
	}

	id, err := strconv.Atoi(value)
	if err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred deleting an external OpenID Connect 1.0 link: failed to parse '%s' as an integer", value)

		ctx.SetJSONError(messageOpenIDConnectUnlinkFailed)

		return
	}

	// The delete is scoped to the authenticated user's username at the storage layer so a user cannot delete another
	// user's link by guessing its id.
	if err = ctx.Providers.StorageProvider.DeleteOpenIDConnectLink(ctx, userSession.Username, id); err != nil {
		ctx.Logger.WithError(err).WithField("username", userSession.Username).Error("Error occurred deleting an external OpenID Connect 1.0 link")

		ctx.SetJSONError(messageOpenIDConnectUnlinkFailed)

		return
	}

	ctx.ReplyOK()
}

func openIDConnectProviderName(ctx *middlewares.AutheliaCtx, id string) (name string) {
	if provider, ok := ctx.Providers.OpenIDConnectRelyingParty.Get(id); ok {
		return provider.Name
	}

	return id
}

func openIDConnectPending(ctx *middlewares.AutheliaCtx, userSession *session.UserSession) (pending *session.OpenIDConnectPending) {
	if userSession.OpenIDConnectPending == nil {
		return nil
	}

	if ctx.GetClock().Now().After(userSession.OpenIDConnectPending.Expires) {
		userSession.OpenIDConnectPending = nil

		return nil
	}

	return userSession.OpenIDConnectPending
}

func isOpenIDConnectLinkConflict(err error) bool {
	e := strings.ToLower(err.Error())

	return strings.Contains(e, "unique") || strings.Contains(e, "duplicate")
}
