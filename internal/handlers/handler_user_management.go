package handlers

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
)

type AdminConfigRequestBody struct {
	Enabled                bool   `json:"enabled"`
	AdminGroup             string `json:"admin_group"`
	AllowAdminsToAddAdmins bool   `json:"allow_admins_to_add_admins"`
}

// UserManagementFieldsResponse represents the response structure for user management field metadata.
type UserManagementFieldsResponse struct {
	RequiredFields  []string                                `json:"required_fields"`
	SupportedFields []string                                `json:"supported_fields"`
	FieldMetadata   map[string]authentication.FieldMetadata `json:"field_metadata"`
}

type FieldMask struct {
	Paths []string `json:"paths"`
}

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
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetJSONError(fmt.Sprintf("Failed to create group: %s", err.Error()))

		return
	}

	ctx.Logger.Infof("Group '%s' created by administrator '%s'", requestBody.GroupName, userSession.Username)
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
	ctx.ReplyOK()
}

func DeleteGroupDELETE(ctx *middlewares.AutheliaCtx) {
	groupNameRaw := ctx.UserValue("groupname")
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
		ctx.Logger.WithError(err).Errorf("Error occurred deleting group '%s'", groupName)
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	ctx.Logger.Infof("Group '%s' deleted by administrator '%s'", groupName, userSession.Username)
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
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

	if err := validateUpdateMask(updateMask, ctx.Providers.UserProvider.GetSupportedFields()); err != nil {
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

	for _, field := range updateMask {
		switch {
		case field == "display_name":
			partialUpdate.DisplayName = requestBody.GetDisplayName()
		case field == "emails":
			partialUpdate.Emails = requestBody.GetEmails()
		case field == "groups":
			partialUpdate.Groups = requestBody.GetGroups()
		case field == "first_name":
			partialUpdate.GivenName = requestBody.GivenName
		case field == "last_name":
			partialUpdate.FamilyName = requestBody.FamilyName
		case field == "middle_name":
			partialUpdate.MiddleName = requestBody.MiddleName
		case field == "full_name":
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
		case field == "zone_info":
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
		default:
			ctx.Logger.Errorf("Unhandled field in update_mask: '%s'", field)
			ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetJSONError("Internal error processing update_mask")

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

func validateUpdateMask(updateMask []string, supportedFields []string) error {
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

		if !slices.Contains(supportedFields, field) {
			return fmt.Errorf("field '%s' is not a valid or modifiable field. Supported fields: %s",
				field, strings.Join(supportedFields, ", "))
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

	userData := userDataBuilder.Build()

	if err = ctx.Providers.UserProvider.ValidateUserData(userData); err != nil {
		if errors.Is(err, authentication.ErrUsernameIsRequired) {
			ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.SetJSONError("Username is required")

			return
		}

		if errors.Is(err, authentication.ErrLastNameIsRequired) {
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
	ctx.Response.SetStatusCode(fasthttp.StatusCreated)
	ctx.ReplyOK()
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

	//TODO: Delete Opaque User Identifiers, User Preferences, 2FA Devices, Oauth Sessions related to Opaque Ids, Remove User from Backend.

	if err = ctx.Providers.UserProvider.DeleteUser(username); err != nil {
		ctx.Logger.Error(err, messageUnableToDeleteUser)
	}

	if err = ctx.Providers.StorageProvider.DeleteUserByUsername(ctx, username); err != nil {
		ctx.Logger.WithError(err).Error(messageUnableToDeleteUserMetadata)
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

// UserManagementFieldsGet returns the field metadata for user management operations.
func UserManagementFieldsGet(ctx *middlewares.AutheliaCtx) {
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

	response := UserManagementFieldsResponse{
		RequiredFields:  ctx.Providers.UserProvider.GetRequiredFields(),
		SupportedFields: ctx.Providers.UserProvider.GetSupportedFields(),
		FieldMetadata:   ctx.Providers.UserProvider.GetFieldMetadata(),
	}

	if err = ctx.SetJSONBody(response); err != nil {
		ctx.Logger.WithError(err).Error("Unable to set user management fields response")
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}
