package handlers

import (
	"errors"
	"fmt"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
)

func GetGroupsGET(ctx *middlewares.AutheliaCtx) {
	var (
		err    error
		groups []string
	)
	if groups, err = ctx.Providers.UserProvider.ListGroups(); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred retrieving groups")
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.SetJSONBody(groups); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred encoding groups")
	}
}

func NewGroupPOST(ctx *middlewares.AutheliaCtx) {
	var (
		err         error
		userSession session.UserSession
	)
	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred adding new group: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred adding new group")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if !UserIsAdmin(ctx, userSession.Groups) {
		ctx.Logger.Errorf("Error occurred adding new group: %s", fmt.Sprintf(logFmtErrUserNotAdmin, userSession.Username))

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var requestBody struct {
		GroupName string `json:"name"`
	}

	if err = ctx.ParseBody(&requestBody); err != nil {
		ctx.Logger.WithError(err).Error("Unable to parse request body")

		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError("Invalid request body")

		return
	}

	if requestBody.GroupName == "" {
		ctx.Logger.Error("Group name is required")

		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError("Group name is required")

		return
	}

	if err = ctx.Providers.UserProvider.AddGroup(requestBody.GroupName); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred creating group '%s'", requestBody.GroupName)

		if errors.Is(err, authentication.ErrGroupExists) {
			ctx.SetStatusCode(fasthttp.StatusConflict)
			ctx.SetJSONError("Group already exists")

			return
		}

		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetJSONError("Failed to create group")

		return
	}

	ctx.Logger.Infof("Group '%s' created by administrator '%s'", requestBody.GroupName, userSession.Username)
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
	ctx.ReplyOK()
}

func DeleteGroupDELETE(ctx *middlewares.AutheliaCtx) {
	groupNameRaw := ctx.UserValue("group")
	if groupNameRaw == nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError("Group name is required")

		return
	}

	groupName := groupNameRaw.(string)

	var (
		err         error
		userSession session.UserSession
	)
	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving user session: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred deleting group")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if !UserIsAdmin(ctx, userSession.Groups) {
		ctx.Logger.Errorf("Error occurred deleting group: %s", fmt.Sprintf(logFmtErrUserNotAdmin, userSession.Username))

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.Providers.UserProvider.DeleteGroup(groupName); err != nil {
		if errors.Is(err, authentication.ErrGroupNotFound) {
			ctx.SetStatusCode(fasthttp.StatusOK)
			return
		}

		ctx.Logger.WithError(err).Errorf("Error occurred deleting group '%s'", groupName)
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	ctx.Logger.Infof("Group '%s' deleted by administrator '%s'", groupName, userSession.Username)
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}
