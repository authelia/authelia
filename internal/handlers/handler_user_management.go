package handlers

import (
	"fmt"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
)

type deleteUserRequestBody struct {
	Username string `json:"username"`
}

type AdminConfigRequestBody struct {
	Enabled                bool   `json:"enabled"`
	AdminGroup             string `json:"admin_group"`
	AllowAdminsToAddAdmins bool   `json:"allow_admins_to_add_admins"`
}

// ChangeUserPUT takes a changeUserRequestBody object and saves any changes.
//
//nolint:gocyclo
func ChangeUserPUT(ctx *middlewares.AutheliaCtx) {
	var (
		err         error
		requestBody *authentication.UserDetailsExtended
		userDetails *authentication.UserDetailsExtended
		adminUser   session.UserSession
	)

	if adminUser, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred modifying user: %s", errStrUserSessionData)
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)
		return
	}

	if adminUser.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred modifying user")
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)
		return
	}

	if !UserIsAdmin(ctx, adminUser.Groups) {
		ctx.Logger.Errorf("Error occurred modifying user: %s", fmt.Sprintf(logFmtErrUserNotAdmin, adminUser.Username))
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)
		return
	}

	requestBody = &authentication.UserDetailsExtended{}
	if err = ctx.ParseBody(requestBody); err != nil {
		ctx.Logger.Error(err, messageUnableToModifyUser)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError("Invalid JSON format")
		return
	}

	if requestBody == nil || requestBody.UserDetails == nil {
		ctx.Logger.Debug("Invalid request body structure")
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError("Invalid request structure")
		return
	}

	if requestBody.UserDetails.Username == "" {
		ctx.Logger.Debug("Username is required")
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError("Username is required")
		return
	}

	if requestBody.Password != "" {
		ctx.Logger.Debug("Password modification not allowed via this endpoint")
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError("Password modification not supported. Use the password change endpoint.")
		return
	}

	if userDetails, err = ctx.Providers.UserProvider.GetUser(requestBody.Username); err != nil {
		ctx.Logger.WithError(err).Errorf("Error retrieving details for user '%s'", requestBody.Username)
		ctx.Response.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetJSONError("User not found")
		return
	}

	requestBody.Password = ""

	if err = ctx.Providers.UserProvider.ValidateUserData(requestBody); err != nil {
		ctx.Logger.WithError(err).Errorf("Validation failed for user '%s'", requestBody.Username)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(fmt.Sprintf("User modification failed: %s", err.Error()))
		return
	}

	if err = ctx.Providers.UserProvider.UpdateUser(requestBody.Username, requestBody); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred updating user '%s'", requestBody.Username)
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetJSONError("Failed to update user")
		return
	}

	if changes := GenerateUserChangeLog(userDetails, requestBody); len(changes) > 0 {
		ctx.Logger.WithFields(changes).Infof("User '%s' modified by administrator '%s'",
			requestBody.UserDetails.Username, adminUser.Username)
	}

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

//nolint:gocyclo
func NewUserPOST(ctx *middlewares.AutheliaCtx) {
	var (
		err            error
		userSession    session.UserSession
		newUserRequest *authentication.UserDetailsExtended
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred adding new user: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred adding new user")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if !UserIsAdmin(ctx, userSession.Groups) {
		ctx.Logger.Errorf("Error occurred adding new user: %s", fmt.Sprintf(logFmtErrUserNotAdmin, userSession.Username))

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.ParseBody(&newUserRequest); err != nil {
		ctx.Logger.Error(err, messageUnableToAddUser)

		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	userDataBuilder := authentication.NewUser(newUserRequest.Username, newUserRequest.Password)

	if newUserRequest.DisplayName != "" {
		userDataBuilder = userDataBuilder.WithDisplayName(newUserRequest.DisplayName)
	}

	if len(newUserRequest.Emails) > 0 {
		userDataBuilder = userDataBuilder.WithEmail(newUserRequest.Emails[0])
	}

	if len(newUserRequest.Groups) > 0 {
		userDataBuilder = userDataBuilder.WithGroups(newUserRequest.Groups)
	}

	if newUserRequest.CommonName != "" {
		userDataBuilder = userDataBuilder.WithCommonName(newUserRequest.CommonName)
	}

	if newUserRequest.GivenName != "" {
		userDataBuilder = userDataBuilder.WithGivenName(newUserRequest.GivenName)
	}

	if newUserRequest.FamilyName != "" {
		userDataBuilder = userDataBuilder.WithFamilyName(newUserRequest.FamilyName)
	}

	if len(newUserRequest.ObjectClasses) > 0 {
		userDataBuilder = userDataBuilder.WithObjectClasses(newUserRequest.ObjectClasses)
	}

	userData := userDataBuilder.Build()

	if err = ctx.Providers.UserProvider.ValidateUserData(userData); err != nil {
		ctx.Logger.WithError(err).Errorf("Validation failed for new user '%s'", newUserRequest.Username)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageOperationFailed)
	}

	if err = ctx.Providers.PasswordPolicy.Check(newUserRequest.Password); err != nil {
		ctx.Error(err, messagePasswordWeak)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messagePasswordWeak)

		return
	}

	if err = ctx.Providers.UserProvider.AddUser(userData); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred creating user '%s'", newUserRequest.Username)
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.Providers.StorageProvider.CreateNewUserMetadata(ctx, newUserRequest.Username); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred creating metadata for user '%s'", newUserRequest.Username)
		ctx.Response.SetStatusCode(fasthttp.StatusMultiStatus)
		ctx.SetJSONError(messageIncompleteUserCreation)

		return
	}

	//TODO: Add user email to notify new user of their new account. Configurable.
	ctx.Logger.Debugf("User '%s' was added.", newUserRequest.Username)
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

func DeleteUserDELETE(ctx *middlewares.AutheliaCtx) {
	var (
		err         error
		userSession session.UserSession
		requestBody deleteUserRequestBody
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred deleting specified user: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred deleting specified user")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.ParseBody(&requestBody); err != nil {
		ctx.Logger.WithError(err).Error(messageUnableToDeleteUser)
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)

		return
	}

	// Delete Opaque User Identifiers, User Preferences, 2FA Devices, Oauth Sessions related to Opaque Ids, Remove User from Backend.

	if err = ctx.Providers.UserProvider.DeleteUser(requestBody.Username); err != nil {
		ctx.Logger.Error(err, messageUnableToDeleteUser)
	}

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

func AdminConfigGET(ctx *middlewares.AutheliaCtx) {
	var (
		err         error
		userSession session.UserSession
		adminConfig AdminConfigRequestBody
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving admin config: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred retrieving admin config")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	adminConfig = AdminConfigRequestBody{
		Enabled:                ctx.Configuration.Administration.Enabled,
		AdminGroup:             ctx.Configuration.Administration.AdminGroup,
		AllowAdminsToAddAdmins: ctx.Configuration.Administration.AllowAdminsToAddAdmins,
	}

	err = ctx.SetJSONBody(adminConfig)
	if err != nil {
		ctx.Logger.Errorf("Unable to set admin config response in body: %+v", err)
	}
}
