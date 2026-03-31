package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/templates"

	"github.com/authelia/authelia/v4/internal/utils"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
)

// UserManagementAttributesResponse represents the response structure for user management field metadata.
type UserManagementAttributesResponse struct {
	RequiredFields  []string                                                  `json:"required_attributes"`
	SupportedFields map[string]authentication.UserManagementAttributeMetadata `json:"supported_attributes"`
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

	newUserRequest = &authentication.UserDetailsExtended{}
	if err = ctx.ParseBody(newUserRequest); err != nil {
		ctx.Logger.Error(err, messageUnableToModifyUser)

		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	userDataBuilder := authentication.NewUser(newUserRequest.Username, newUserRequest.Password)

	if newUserRequest.DisplayName != "" {
		userDataBuilder = userDataBuilder.WithDisplayName(newUserRequest.DisplayName)
	}

	if len(newUserRequest.Emails) > 0 {
		if !utils.ValidateEmailString(newUserRequest.Emails[0]) {
			ctx.Logger.Debugf("unable to add user '%s': %s", newUserRequest.GetUsername(), messageInvalidEmail)
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.SetJSONError(fmt.Sprintf("unable to add user '%s': %s", newUserRequest.GetUsername(), messageInvalidEmail))

			return
		}

		userDataBuilder = userDataBuilder.WithEmail(newUserRequest.Emails[0])
	}

	if len(newUserRequest.Groups) > 0 {
		availableGroups, err := ctx.Providers.UserProvider.ListGroups()
		if err != nil {
			ctx.Logger.WithError(err).Error("Error occurred retrieving groups")
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetJSONError(messageOperationFailed)

			return
		}

		for _, group := range newUserRequest.Groups {
			if !slices.Contains(availableGroups, group) {
				ctx.SetStatusCode(fasthttp.StatusBadRequest)
				ctx.SetJSONError(fmt.Sprintf("Group '%s' does not exist", group))

				return
			}
		}

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

	if newUserRequest.MiddleName != "" {
		userDataBuilder = userDataBuilder.WithMiddleName(newUserRequest.MiddleName)
	}

	if newUserRequest.Nickname != "" {
		userDataBuilder = userDataBuilder.WithNickname(newUserRequest.Nickname)
	}

	if newUserRequest.Gender != "" {
		userDataBuilder = userDataBuilder.WithGender(newUserRequest.Gender)
	}

	if newUserRequest.Birthdate != "" {
		userDataBuilder = userDataBuilder.WithBirthdate(newUserRequest.Birthdate)
	}

	if newUserRequest.PhoneNumber != "" {
		userDataBuilder = userDataBuilder.WithPhoneNumber(newUserRequest.PhoneNumber)
	}

	if newUserRequest.PhoneExtension != "" {
		userDataBuilder = userDataBuilder.WithPhoneExtension(newUserRequest.PhoneExtension)
	}

	if newUserRequest.ZoneInfo != "" {
		userDataBuilder = userDataBuilder.WithZoneInfo(newUserRequest.ZoneInfo)
	}

	if newUserRequest.Profile != nil {
		userDataBuilder = userDataBuilder.WithProfile(newUserRequest.Profile.String())
	}

	if newUserRequest.Picture != nil {
		userDataBuilder = userDataBuilder.WithPicture(newUserRequest.Picture.String())
	}

	if newUserRequest.Website != nil {
		userDataBuilder = userDataBuilder.WithWebsite(newUserRequest.Website.String())
	}

	if newUserRequest.Locale != nil {
		userDataBuilder = userDataBuilder.WithLocale(newUserRequest.Locale.String())
	}

	if newUserRequest.Address != nil {
		userDataBuilder = userDataBuilder.WithAddress(
			newUserRequest.Address.StreetAddress,
			newUserRequest.Address.Locality,
			newUserRequest.Address.Region,
			newUserRequest.Address.PostalCode,
			newUserRequest.Address.Country,
		)
	}

	if newUserRequest.Extra != nil {
		for key, value := range newUserRequest.Extra {
			userDataBuilder.WithExtra(key, value)
		}
	}

	userData := userDataBuilder.Build()

	if err = ctx.Providers.UserProvider.ValidateUserData(userData); err != nil {
		if errors.Is(err, authentication.ErrUsernameIsRequired) {
			ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.SetJSONError("Username is required")

			return
		}

		if errors.Is(err, authentication.ErrFamilyNameIsRequired) {
			ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.SetJSONError("Last name is required")

			return
		}

		ctx.Logger.WithError(err).Errorf("Validation failed for new user '%s'", newUserRequest.Username)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.Providers.PasswordPolicy.Check(newUserRequest.Password); err != nil {
		ctx.Error(err, messagePasswordWeak)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messagePasswordWeak)

		return
	}

	if err = ctx.Providers.UserProvider.AddUser(userData); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred creating user '%s'", newUserRequest.Username)

		if ldap.IsErrorAnyOf(err, ldap.LDAPResultEntryAlreadyExists, ldap.LDAPResultConstraintViolation) {
			ctx.Response.SetStatusCode(fasthttp.StatusConflict)
			ctx.SetJSONError("User already exists")

			return
		}

		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if _, err = ctx.Providers.StorageProvider.LoadUserMetadataByUsername(ctx, newUserRequest.Username); err != nil {
		if err = ctx.Providers.StorageProvider.CreateNewUserMetadata(ctx, newUserRequest.Username); err != nil {
			ctx.Logger.WithError(err).Errorf("Error occurred creating metadata for user '%s'", newUserRequest.Username)
			ctx.Response.SetStatusCode(fasthttp.StatusMultiStatus)
			ctx.SetJSONError(messageIncompleteUserCreation)

			return
		}
	} else {
		ctx.Logger.Debugf("User metadata for '%s' already exists, skipping creation", newUserRequest.Username)
	}

	//TODO: Add user email to notify new user of their new account. Configurable.
	ctx.Logger.Debugf("User '%s' was added.", newUserRequest.Username)
	ctx.Response.SetStatusCode(fasthttp.StatusCreated)
	ctx.ReplyOK()
}

type FieldMask struct {
	Paths []string `json:"paths"`
}

// ChangeUserPATCH updates specific fields of a user based on the provided update_mask.
//
//nolint:gocyclo
func ChangeUserPATCH(ctx *middlewares.AutheliaCtx) {
	usernameRaw := ctx.UserValue("username")
	if usernameRaw == nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageUsernameRequired)

		return
	}

	username := usernameRaw.(string)

	var (
		err         error
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

	updateMaskStr := string(ctx.QueryArgs().Peek("update_mask"))
	if updateMaskStr == "" {
		ctx.Logger.Debug("update_mask is required")
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError("update_mask query parameter is required. Specify comma-separated field names to update (e.g., ?update_mask=display_name,emails,address.city)")

		return
	}

	updateMask := strings.Split(updateMaskStr, ",")
	for i := range updateMask {
		updateMask[i] = strings.TrimSpace(updateMask[i])
	}

	var requestBody authentication.UserDetailsExtended
	if err = ctx.ParseBody(&requestBody); err != nil {
		ctx.Logger.WithError(err).Error(messageUnableToModifyUser)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(fmt.Sprintf("Invalid JSON format: %s", err.Error()))

		return
	}

	if username == "" {
		ctx.Logger.Debug("Username is required")
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError("Username is required")

		return
	}

	if slices.Contains(updateMask, "password") || requestBody.Password != "" {
		ctx.Logger.Debug("Password modification not allowed via this endpoint")
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError("Password modification not supported. Use the password change endpoint.")

		return
	}

	if userDetails, err = ctx.Providers.UserProvider.GetUser(username); err != nil {
		ctx.Logger.WithError(err).Errorf("Error retrieving details for user '%s'", username)
		ctx.Response.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetJSONError("User not found")

		return
	}

	if err := validateUpdateMask(updateMask, ctx.Providers.UserProvider.GetSupportedAttributes()); err != nil {
		ctx.Logger.WithError(err).Error("Invalid update_mask")
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(err.Error())

		return
	}

	partialUpdate := &authentication.UserDetailsExtended{
		UserDetails: &authentication.UserDetails{
			Username: username,
		},
	}

	addressFields := filterAddressFields(updateMask)
	if len(addressFields) > 0 {
		partialUpdate.Address = &authentication.UserDetailsAddress{}
	}

	supportedFields := ctx.Providers.UserProvider.GetSupportedAttributes()

	for _, field := range updateMask {
		switch {
		case field == "display_name":
			partialUpdate.DisplayName = requestBody.GetDisplayName()
		case field == "mail":
			partialUpdate.Emails = requestBody.GetEmails()
			if !utils.ValidateEmailString(partialUpdate.Emails[0]) {
				ctx.Logger.Debugf("unable to add user '%s': %s", partialUpdate.GetUsername(), messageInvalidEmail)
				ctx.SetStatusCode(fasthttp.StatusBadRequest)
				ctx.SetJSONError(fmt.Sprintf("unable to add user '%s': %s", partialUpdate.GetUsername(), messageInvalidEmail))

				return
			}
		case field == "groups":
			partialUpdate.Groups = requestBody.GetGroups()
		case field == "given_name":
			partialUpdate.GivenName = requestBody.GivenName
		case field == "family_name":
			partialUpdate.FamilyName = requestBody.FamilyName
		case field == "middle_name":
			partialUpdate.MiddleName = requestBody.MiddleName
		case field == "common_name":
			partialUpdate.CommonName = requestBody.CommonName
		case field == "nickname":
			partialUpdate.Nickname = requestBody.Nickname
		case field == "profile":
			partialUpdate.Profile = requestBody.Profile
		case field == "picture":
			partialUpdate.Picture = requestBody.Picture
		case field == "website":
			partialUpdate.Website = requestBody.Website
		case field == "gender":
			partialUpdate.Gender = requestBody.Gender
		case field == "birthdate":
			partialUpdate.Birthdate = requestBody.Birthdate
		case field == "zoneinfo":
			partialUpdate.ZoneInfo = requestBody.ZoneInfo
		case field == "locale":
			partialUpdate.Locale = requestBody.Locale
		case field == "phone_number":
			partialUpdate.PhoneNumber = requestBody.PhoneNumber
		case field == "phone_extension":
			partialUpdate.PhoneExtension = requestBody.PhoneExtension
		case field == "extra":
			partialUpdate.Extra = requestBody.Extra
		//nolint:goconst
		case field == "address":
			partialUpdate.Address = requestBody.Address
		case strings.HasPrefix(field, "address."):
			if requestBody.Address == nil {
				ctx.Logger.Debugf("Address object not provided for field '%s'", field)
				ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
				ctx.SetJSONError(fmt.Sprintf("Address object must be provided when updating '%s'", field))

				return
			}

			subField := strings.TrimPrefix(field, "address.")
			switch subField {
			case "street_address":
				partialUpdate.Address.StreetAddress = requestBody.Address.StreetAddress
			case "locality":
				partialUpdate.Address.Locality = requestBody.Address.Locality
			case "region":
				partialUpdate.Address.Region = requestBody.Address.Region
			case "postal_code":
				partialUpdate.Address.PostalCode = requestBody.Address.PostalCode
			case "country":
				partialUpdate.Address.Country = requestBody.Address.Country
			default:
				ctx.Logger.Errorf("Unknown address subfield: '%s'", subField)
				ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
				ctx.SetJSONError(fmt.Sprintf("Unknown address field: '%s'", subField))

				return
			}
		case strings.HasPrefix(field, "extra."):
			extraField := strings.TrimPrefix(field, "extra.")

			if _, isExtraField := supportedFields[field]; isExtraField && ctx.Providers.UserProvider.IsExtraAttribute(extraField) {
				if partialUpdate.Extra == nil {
					partialUpdate.Extra = make(map[string]interface{})
				}

				if requestBody.Extra != nil {
					if value, exists := requestBody.Extra[extraField]; exists {
						partialUpdate.Extra[extraField] = value
					} else {
						ctx.Logger.Debugf("Extra field '%s' not provided in extra object", extraField)
						ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
						ctx.SetJSONError(fmt.Sprintf("Field '%s' must be provided in the 'extra' object in the request body", extraField))

						return
					}
				} else {
					ctx.Logger.Debugf("Extra object not provided for extra field '%s'", extraField)
					ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
					ctx.SetJSONError(fmt.Sprintf("'extra' object must be provided in the request body when updating '%s'", field))

					return
				}
			} else {
				ctx.Logger.Errorf("Unknown extra field: '%s'", extraField)
				ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
				ctx.SetJSONError(fmt.Sprintf("Unknown extra field: '%s'", extraField))

				return
			}
		default:
			ctx.Logger.Errorf("Unhandled field in update_mask: '%s'", field)
			ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.SetJSONError(fmt.Sprintf("Unhandled field in update_mask: '%s'. Extra fields must use 'extra.' prefix (e.g., 'extra.%s')", field, field))

			return
		}
	}

	if err = ctx.Providers.UserProvider.ValidatePartialUpdate(partialUpdate, updateMask); err != nil {
		ctx.Logger.WithError(err).Errorf("Validation failed for user '%s'", username)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(fmt.Sprintf("User modification failed: %s", err.Error()))

		return
	}

	if err = ctx.Providers.UserProvider.UpdateUserWithMask(username, partialUpdate, updateMask); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred updating user '%s'", username)

		if errors.Is(err, authentication.ErrGroupNotFound) {
			ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.SetJSONError(fmt.Sprintf("group '%s' not found", username))

			return
		}

		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetJSONError("Failed to update user")

		return
	}

	if changes := GenerateUserChangeLogWithMask(userDetails, partialUpdate, updateMask); len(changes) > 0 {
		ctx.Logger.WithFields(changes).Infof("User '%s' modified by administrator '%s' (fields: %s)",
			username, adminUser.Username, strings.Join(updateMask, ", "))
	}

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

func DeleteUserDELETE(ctx *middlewares.AutheliaCtx) {
	usernameRaw := ctx.UserValue("username")
	if usernameRaw == nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageUsernameRequired)

		return
	}

	username := usernameRaw.(string)

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
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred deleting specified user")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if !UserIsAdmin(ctx, userSession.Groups) {
		ctx.Logger.Warnf("problem retrieving user management fields: user '%s' is not an admin", userSession.Username)
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	//TODO: Delete Opaque User Identifiers, User Preferences, 2FA Devices, Oauth Sessions related to Opaque Ids, Remove User from Backend.

	if err = ctx.Providers.UserProvider.DeleteUser(username); err != nil {
		ctx.Logger.Error(err, messageUnableToDeleteUser)
	}

	if err = ctx.Providers.StorageProvider.DeleteUserByUsername(ctx, username); err != nil {
		ctx.Logger.WithError(err).Error(messageUnableToDeleteUserMetadata)
	}

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

func AdminChangePasswordPOST(ctx *middlewares.AutheliaCtx) {
	usernameRaw := ctx.UserValue("username")
	if usernameRaw == nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageUsernameRequired)

		return
	}

	username := usernameRaw.(string)

	var (
		err         error
		userSession session.UserSession
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving user management fields: %s", errStrUserSessionData)
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred retrieving user management fields")
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if !UserIsAdmin(ctx, userSession.Groups) {
		ctx.Logger.Warnf("problem retrieving user management fields: user '%s' is not an admin", userSession.Username)
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var changePasswordRequestBody struct {
		Password string `json:"password"`
	}

	if err = ctx.ParseBody(&changePasswordRequestBody); err != nil {
		ctx.Logger.WithError(err).
			WithFields(map[string]any{"username": username}).
			Error("Unable to change password for user: unable to parse request body")
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetJSONError(messageUnableToChangePassword)

		return
	}

	if err = ctx.Providers.PasswordPolicy.Check(changePasswordRequestBody.Password); err != nil {
		ctx.Logger.WithError(err).
			WithFields(map[string]any{"username": username}).
			Debug("Unable to change password for user as their new password was weak or empty")
		ctx.SetJSONError(messagePasswordWeak)
		ctx.SetStatusCode(http.StatusBadRequest)

		return
	}

	if err = ctx.Providers.UserProvider.UpdatePassword(username, changePasswordRequestBody.Password); err != nil {
		switch {
		case utils.IsStringInSliceContains(err.Error(), ldapPasswordComplexityCodes),
			utils.IsStringInSliceContains(err.Error(), ldapPasswordComplexityErrors):
			ctx.Logger.WithError(err).
				WithFields(map[string]any{"username": username}).
				Debug("Unable to change password for user as their new password was weak or empty")
			ctx.SetJSONError(messagePasswordWeak)
			ctx.SetStatusCode(http.StatusBadRequest)
		default:
			ctx.Logger.WithError(err).
				WithFields(map[string]any{"username": username}).
				Error("Unable to change password for user for an unknown reason")
			ctx.SetJSONError(messageOperationFailed)
			ctx.SetStatusCode(http.StatusInternalServerError)
		}

		return
	}

	ctx.Logger.WithFields(map[string]any{
		"username":       username,
		"admin_username": userSession.Username,
	}).Debug("User's password was changed by admin")

	userInfo, err := ctx.Providers.UserProvider.GetDetails(username)
	if err != nil {
		ctx.Logger.Error(err)
		ctx.ReplyOK()

		return
	}

	//TODO: should this email be sent? should it have a different action name? remote ip?
	data := templates.EmailEventValues{
		Title:       "Password changed successfully",
		DisplayName: userInfo.DisplayName,
		RemoteIP:    "",
		Details: map[string]any{
			"Action": "Password Change",
		},
		BodyPrefix: eventEmailActionAdminPasswordModifyPrefix,
		BodyEvent:  eventEmailActionPasswordChange,
		BodySuffix: eventEmailActionAdminPasswordModifySuffix,
	}

	addresses := userInfo.Addresses()

	ctx.Logger.WithFields(map[string]any{
		"username": username,
		"email":    addresses[0].String(),
	}).
		Debug("Sending an email to inform user that their password has changed.")

	if err = ctx.Providers.Notifier.Send(ctx, addresses[0], "Password changed successfully", ctx.Providers.Templates.GetEventEmailTemplate(), data); err != nil {
		ctx.Logger.WithError(err).
			WithFields(map[string]any{
				"username": username,
				"email":    addresses[0].String(),
			}).
			Debug("Unable to notify user of password change")
		ctx.ReplyOK()

		return
	}
}

func AdminResetPasswordPOST(ctx *middlewares.AutheliaCtx) {
	usernameRaw := ctx.UserValue("username")
	if usernameRaw == nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageUsernameRequired)

		return
	}

	username := usernameRaw.(string)

	var (
		err         error
		userSession session.UserSession
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving user management fields: %s", errStrUserSessionData)
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred retrieving user management fields")
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if !UserIsAdmin(ctx, userSession.Groups) {
		ctx.Logger.Warnf("problem retrieving user management fields: user '%s' is not an admin", userSession.Username)
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	userInfo, err := ctx.Providers.UserProvider.GetDetails(username)
	if err != nil {
		ctx.Logger.Error(err)
		ctx.ReplyOK()

		return
	}

	var userIdentity = &session.Identity{
		Username:    userInfo.Username,
		Email:       userInfo.Emails[0],
		DisplayName: userInfo.DisplayName,
	}

	var ivArgs = middlewares.IdentityVerificationStartArgs{
		MailTitle:               "Reset your password",
		MailButtonContent:       "Reset",
		MailButtonRevokeContent: "Revoke",
		TargetEndpoint:          "/reset-password/step2",
		RevokeEndpoint:          "/revoke/reset-password",
		ActionClaim:             ActionResetPassword,
		IdentityRetrieverFunc:   nil,
	}

	success, err := middlewares.MintTokenAndSendPasswordResetEmail(ctx, ivArgs, userIdentity)
	if err != nil {
		ctx.Logger.WithError(err).WithFields(map[string]any{
			"username":       userInfo.Username,
			"admin_username": userSession.Username,
		}).Errorf("unable to send reset password email")

		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetJSONError(messageUnableToSendPasswordResetEmail)

		return
	}

	if !success {
		ctx.Logger.WithFields(map[string]any{
			"username":       userInfo.Username,
			"admin_username": userSession.Username,
		}).Errorf("unable to send reset password email")

		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetJSONError(messageUnableToSendPasswordResetEmail)

		return
	}

	ctx.Logger.WithFields(map[string]any{
		"username":       userInfo.Username,
		"admin_username": userSession.Username,
	}).Debugf("Successfully sent reset password email")

	ctx.SetStatusCode(fasthttp.StatusCreated)
}

// UserManagementAttributesGet returns the field metadata for user management operations.
func UserManagementAttributesGet(ctx *middlewares.AutheliaCtx) {
	var (
		err         error
		userSession session.UserSession
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving user management fields: %s", errStrUserSessionData)
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred retrieving user management fields")
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if !UserIsAdmin(ctx, userSession.Groups) {
		ctx.Logger.Warnf("problem retrieving user management fields: user '%s' is not an admin", userSession.Username)
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	response := UserManagementAttributesResponse{
		RequiredFields:  ctx.Providers.UserProvider.GetRequiredAttributes(),
		SupportedFields: ctx.Providers.UserProvider.GetSupportedAttributes(),
	}

	if err = ctx.SetJSONBody(response); err != nil {
		ctx.Logger.WithError(err).Error("Unable to set user management fields response")
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}

func validateUpdateMask(updateMask []string, supportedFields map[string]authentication.UserManagementAttributeMetadata) error {
	for _, field := range updateMask {
		if strings.HasPrefix(field, "address.") {
			subField := strings.TrimPrefix(field, "address.")

			validAddressFields := []string{"street_address", "locality", "region", "postal_code", "country"}
			if !slices.Contains(validAddressFields, subField) {
				return fmt.Errorf("field 'address.%s' is not a valid address field. Valid address fields: %s",
					subField, strings.Join(validAddressFields, ", "))
			}

			continue
		}

		if _, exists := supportedFields[field]; !exists {
			validFields := make([]string, 0, len(supportedFields))
			for fieldName := range supportedFields {
				validFields = append(validFields, fieldName)
			}

			return fmt.Errorf("field '%s' is not a valid or modifiable field. Supported fields: %s",
				field, strings.Join(validFields, ", "))
		}
	}

	return nil
}

// filterAddressFields returns only the address-related fields from the update mask.
func filterAddressFields(updateMask []string) []string {
	var addressFields []string

	for _, field := range updateMask {
		if field == "address" || strings.HasPrefix(field, "address.") {
			addressFields = append(addressFields, field)
		}
	}

	return addressFields
}
