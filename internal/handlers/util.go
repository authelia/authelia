package handlers

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"slices"
	"strings"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/templates"
)

const (
	eventLogKeyAction      = "Action"
	eventLogKeyCategory    = "Category"
	eventLogKeyDescription = "Description"

	eventEmailAction2FABody  = "Second Factor Method"
	eventLogAction2FAAdded   = "Second Factor Method Added"
	eventLogAction2FARemoved = "Second Factor Method Removed"

	eventEmailAction2FAPrefix        = "a"
	eventEmailAction2FAAddedSuffix   = "was added to your account."
	eventEmailAction2FARemovedSuffix = "was removed from your account."

	eventEmailActionPasswordModifyPrefix = "your"
	eventEmailActionPasswordReset        = "Password Reset"
	eventEmailActionPasswordChange       = "Password Change"
	eventEmailActionPasswordModifySuffix = "was successful."

	eventLogCategoryOneTimePassword    = "One-Time Password"
	eventLogCategoryWebAuthnCredential = "WebAuthn Credential" //nolint:gosec
)

type emailEventBody struct {
	Prefix string
	Body   string
	Suffix string
}

func ctxLogEvent(ctx *middlewares.AutheliaCtx, username, description string, body emailEventBody, eventDetails map[string]any) {
	var (
		details *authentication.UserDetails
		err     error
	)

	ctx.Logger.Debugf("Getting user details for notification")

	// Send Notification.
	if details, err = ctx.Providers.UserProvider.GetDetails(username); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred looking up user details for user '%s' while attempting to alert them of an important event", username)
		return
	}

	if len(details.Emails) == 0 {
		ctx.Logger.WithError(fmt.Errorf("no email address was found for user")).Errorf("Error occurred looking up user details for user '%s' while attempting to alert them of an important event", username)
		return
	}

	data := templates.EmailEventValues{
		Title:       description,
		DisplayName: details.DisplayName,
		RemoteIP:    ctx.RemoteIP().String(),
		Details:     eventDetails,
		BodyPrefix:  body.Prefix,
		BodyEvent:   body.Body,
		BodySuffix:  body.Suffix,
	}

	ctx.Logger.Debugf("Getting user addresses for notification")

	addresses := details.Addresses()

	ctx.Logger.Debugf("Sending an email to user %s (%s) to inform them of an important event.", username, addresses[0].String())

	if err = ctx.Providers.Notifier.Send(ctx, addresses[0], description, ctx.Providers.Templates.GetEventEmailTemplate(), data); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred sending notification to user '%s' while attempting to alert them of an important event", username)
		return
	}
}

func redactEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}

	localRunes := []rune(parts[0])
	domain := parts[1]

	if len(localRunes) <= 2 {
		return strings.Repeat("*", len(localRunes)) + "@" + domain
	}

	first := string(localRunes[0])
	last := string(localRunes[len(localRunes)-1])
	middle := strings.Repeat("*", len(localRunes)-2)

	return first + middle + last + "@" + domain
}

func isRegulatorSkippedErr(err error) bool {
	var e *authentication.PoolErr

	if errors.As(err, &e) {
		return e.IsDeadlineError()
	}

	return false
}

func MergeUserDetailsWithInfoMany(userDetails []authentication.UserDetailsExtended, userInfoList []model.UserInfo) []authentication.UserDetailsExtended {
	userInfoMap := make(map[string]model.UserInfo)
	for _, info := range userInfoList {
		userInfoMap[info.Username] = info
	}

	result := make([]authentication.UserDetailsExtended, 0, len(userDetails))
	for _, details := range userDetails {
		if details.UserDetails != nil {
			if info, exists := userInfoMap[details.Username]; exists {
				result = append(result, *MergeUserDetailsWithInfo(&details, info))
			} else {
				result = append(result, details)
			}
		}
	}

	return result
}

func MergeUserDetailsWithInfo(userDetails *authentication.UserDetailsExtended, userInfo model.UserInfo) *authentication.UserDetailsExtended {
	merged := userDetails

	merged.LastLoggedIn = userInfo.LastLoggedIn
	merged.LastPasswordChange = userInfo.LastPasswordChange
	merged.UserCreatedAt = userInfo.UserCreatedAt
	merged.Method = userInfo.Method
	merged.HasTOTP = userInfo.HasTOTP
	merged.HasWebAuthn = userInfo.HasWebAuthn
	merged.HasDuo = userInfo.HasDuo

	return merged
}

// MergeUserInfoAndDetails combines the list of attributes in userInfo with the list of users in users.
func MergeUserInfoAndDetails(userInfo []model.UserInfo, users []authentication.UserDetailsExtended) []model.UserInfo {
	// Map of username -> UserDetailsExtended for quick lookup.
	userDetailsMap := make(map[string]authentication.UserDetailsExtended)
	userInfoMap := make(map[string]bool)

	// Build the lookup map.
	for _, user := range users {
		if user.UserDetails != nil {
			userDetailsMap[user.Username] = user
		}
	}

	// Update existing userInfo entries with details from UserDetailsExtended.
	for i, info := range userInfo {
		if details, ok := userDetailsMap[info.Username]; ok && details.UserDetails != nil {
			userInfo[i].DisplayName = details.DisplayName
			userInfo[i].Emails = details.Emails
			userInfo[i].Groups = details.Groups
			userInfoMap[info.Username] = true
		}
	}

	// Add any users from UserDetailsExtended that weren't in the original userInfo.
	for _, user := range users {
		if user.UserDetails != nil {
			if _, exists := userInfoMap[user.Username]; !exists {
				userInfo = append(userInfo, model.UserInfo{
					Username:    user.Username,
					DisplayName: user.DisplayName,
					Emails:      user.Emails,
					Groups:      user.Groups,
				})
			}
		}
	}

	return userInfo
}

func UserIsAdmin(ctx *middlewares.AutheliaCtx, userGroups []string) bool {
	return slices.Contains(userGroups, ctx.Configuration.Administration.AdminGroup)
}

//nolint:gocyclo
//nolint:gocyclo
func GenerateUserChangeLog(original *authentication.UserDetailsExtended, changes *authentication.UserDetailsExtended) map[string]interface{} {
	changeLog := make(map[string]interface{})

	if original.UserDetails != nil && changes.UserDetails != nil {
		if original.DisplayName != changes.DisplayName {
			changeLog["display_name"] = map[string]interface{}{
				"from": original.DisplayName,
				"to":   changes.DisplayName,
			}
		}

		if !reflect.DeepEqual(original.Emails, changes.Emails) {
			changeLog["emails"] = map[string]interface{}{
				"from": original.Emails,
				"to":   changes.Emails,
			}
		}

		if !reflect.DeepEqual(original.Groups, changes.Groups) {
			changeLog["groups"] = map[string]interface{}{
				"from": original.Groups,
				"to":   changes.Groups,
			}
		}
	}

	if original.GivenName != changes.GivenName {
		changeLog["given_name"] = map[string]interface{}{
			"from": original.GivenName,
			"to":   changes.GivenName,
		}
	}

	if original.FamilyName != changes.FamilyName {
		changeLog["family_name"] = map[string]interface{}{
			"from": original.FamilyName,
			"to":   changes.FamilyName,
		}
	}

	if original.MiddleName != changes.MiddleName {
		changeLog["middle_name"] = map[string]interface{}{
			"from": original.MiddleName,
			"to":   changes.MiddleName,
		}
	}

	if original.Nickname != changes.Nickname {
		changeLog["nickname"] = map[string]interface{}{
			"from": original.Nickname,
			"to":   changes.Nickname,
		}
	}

	if original.CommonName != changes.CommonName {
		changeLog["common_name"] = map[string]interface{}{
			"from": original.CommonName,
			"to":   changes.CommonName,
		}
	}

	checkURL := func(fieldName string, oldVal, newVal *url.URL) {
		oldStr := ""
		newStr := ""

		if oldVal != nil {
			oldStr = oldVal.String()
		}

		if newVal != nil {
			newStr = newVal.String()
		}

		if oldStr != newStr {
			changeLog[fieldName] = map[string]interface{}{
				"from": oldStr,
				"to":   newStr,
			}
		}
	}
	//TODO: we probably shouldnt log entire urls -- they could be *really* long.
	checkURL("profile", original.Profile, changes.Profile)
	checkURL("picture", original.Picture, changes.Picture)
	checkURL("website", original.Website, changes.Website)

	if !reflect.DeepEqual(original.Extra, changes.Extra) {
		changeLog["extra"] = map[string]interface{}{
			"from": original.Extra,
			"to":   changes.Extra,
		}
	}

	if changes.Password != "" {
		changeLog["password"] = "changed"
	}

	return changeLog
}

// GenerateUserChangeLogWithMask creates a log of changes between old and new user details, but only for fields specified in the update mask.
//
//nolint:gocyclo
func GenerateUserChangeLogWithMask(oldUser, newUser *authentication.UserDetailsExtended, updateMask []string) map[string]interface{} {
	changes := make(map[string]interface{})

	if oldUser == nil || newUser == nil {
		return changes
	}

	maskSet := make(map[string]bool)
	for _, field := range updateMask {
		maskSet[field] = true
	}

	inMask := func(field string) bool {
		if maskSet[field] {
			return true
		}

		if strings.Contains(field, ".") {
			parent := field[:strings.LastIndex(field, ".")]
			if maskSet[parent] {
				return true
			}
		}

		return false
	}

	if inMask("display_name") && oldUser.GetDisplayName() != newUser.GetDisplayName() {
		changes["display_name"] = map[string]interface{}{
			"from": oldUser.GetDisplayName(),
			"to":   newUser.GetDisplayName(),
		}
	}

	if inMask("first_name") && oldUser.GivenName != newUser.GivenName {
		changes["first_name"] = map[string]interface{}{
			"from": oldUser.GivenName,
			"to":   newUser.GivenName,
		}
	}

	if inMask("last_name") && oldUser.FamilyName != newUser.FamilyName {
		changes["last_name"] = map[string]interface{}{
			"from": oldUser.FamilyName,
			"to":   newUser.FamilyName,
		}
	}

	if inMask("middle_name") && oldUser.MiddleName != newUser.MiddleName {
		changes["middle_name"] = map[string]interface{}{
			"from": oldUser.MiddleName,
			"to":   newUser.MiddleName,
		}
	}

	if inMask("full_name") && oldUser.CommonName != newUser.CommonName {
		changes["full_name"] = map[string]interface{}{
			"from": oldUser.CommonName,
			"to":   newUser.CommonName,
		}
	}

	if inMask("nickname") && oldUser.Nickname != newUser.Nickname {
		changes["nickname"] = map[string]interface{}{
			"from": oldUser.Nickname,
			"to":   newUser.Nickname,
		}
	}

	if inMask("gender") && oldUser.Gender != newUser.Gender {
		changes["gender"] = map[string]interface{}{
			"from": oldUser.Gender,
			"to":   newUser.Gender,
		}
	}

	if inMask("birthdate") && oldUser.Birthdate != newUser.Birthdate {
		changes["birthdate"] = map[string]interface{}{
			"from": oldUser.Birthdate,
			"to":   newUser.Birthdate,
		}
	}

	if inMask("zone_info") && oldUser.ZoneInfo != newUser.ZoneInfo {
		changes["zone_info"] = map[string]interface{}{
			"from": oldUser.ZoneInfo,
			"to":   newUser.ZoneInfo,
		}
	}

	if inMask("phone_number") && oldUser.PhoneNumber != newUser.PhoneNumber {
		changes["phone_number"] = map[string]interface{}{
			"from": oldUser.PhoneNumber,
			"to":   newUser.PhoneNumber,
		}
	}

	if inMask("phone_extension") && oldUser.PhoneExtension != newUser.PhoneExtension {
		changes["phone_extension"] = map[string]interface{}{
			"from": oldUser.PhoneExtension,
			"to":   newUser.PhoneExtension,
		}
	}

	if inMask("locale") && oldUser.GetLocale() != newUser.GetLocale() {
		changes["locale"] = map[string]interface{}{
			"from": oldUser.GetLocale(),
			"to":   newUser.GetLocale(),
		}
	}

	if inMask("profile") && oldUser.GetProfile() != newUser.GetProfile() {
		changes["profile"] = map[string]interface{}{
			"from": oldUser.GetProfile(),
			"to":   newUser.GetProfile(),
		}
	}

	if inMask("picture") && oldUser.GetPicture() != newUser.GetPicture() {
		changes["picture"] = map[string]interface{}{
			"from": oldUser.GetPicture(),
			"to":   newUser.GetPicture(),
		}
	}

	if inMask("website") && oldUser.GetWebsite() != newUser.GetWebsite() {
		changes["website"] = map[string]interface{}{
			"from": oldUser.GetWebsite(),
			"to":   newUser.GetWebsite(),
		}
	}

	if inMask("emails") && !reflect.DeepEqual(oldUser.GetEmails(), newUser.GetEmails()) {
		changes["emails"] = map[string]interface{}{
			"from": oldUser.GetEmails(),
			"to":   newUser.GetEmails(),
		}
	}

	if inMask("groups") && !reflect.DeepEqual(oldUser.GetGroups(), newUser.GetGroups()) {
		changes["groups"] = map[string]interface{}{
			"from": oldUser.GetGroups(),
			"to":   newUser.GetGroups(),
		}
	}

	if inMask("address") || inMask("address.street_address") || inMask("address.locality") ||
		inMask("address.region") || inMask("address.postal_code") || inMask("address.country") {
		addressChanges := make(map[string]map[string]interface{})

		oldAddr := oldUser.Address
		newAddr := newUser.Address

		//nolint:gocritic
		if oldAddr == nil && newAddr != nil {
			changes["address"] = map[string]interface{}{
				"from": nil,
				"to":   newAddr,
			}
		} else if oldAddr != nil && newAddr == nil {
			changes["address"] = map[string]interface{}{
				"from": oldAddr,
				"to":   nil,
			}
		} else if oldAddr != nil && newAddr != nil {
			if (inMask("address") || inMask("address.street_address")) &&
				oldAddr.StreetAddress != newAddr.StreetAddress {
				addressChanges["street_address"] = map[string]interface{}{
					"from": oldAddr.StreetAddress,
					"to":   newAddr.StreetAddress,
				}
			}

			if (inMask("address") || inMask("address.locality")) &&
				oldAddr.Locality != newAddr.Locality {
				addressChanges["locality"] = map[string]interface{}{
					"from": oldAddr.Locality,
					"to":   newAddr.Locality,
				}
			}

			if (inMask("address") || inMask("address.region")) &&
				oldAddr.Region != newAddr.Region {
				addressChanges["region"] = map[string]interface{}{
					"from": oldAddr.Region,
					"to":   newAddr.Region,
				}
			}

			if (inMask("address") || inMask("address.postal_code")) &&
				oldAddr.PostalCode != newAddr.PostalCode {
				addressChanges["postal_code"] = map[string]interface{}{
					"from": oldAddr.PostalCode,
					"to":   newAddr.PostalCode,
				}
			}

			if (inMask("address") || inMask("address.country")) &&
				oldAddr.Country != newAddr.Country {
				addressChanges["country"] = map[string]interface{}{
					"from": oldAddr.Country,
					"to":   newAddr.Country,
				}
			}

			if len(addressChanges) > 0 {
				changes["address"] = addressChanges
			}
		}
	}

	if inMask("extra") {
		extraChanges := make(map[string]interface{})

		for key, newVal := range newUser.Extra {
			if oldVal, exists := oldUser.Extra[key]; !exists || oldVal != newVal {
				extraChanges[key] = map[string]interface{}{
					"from": oldUser.Extra[key],
					"to":   newVal,
				}
			}
		}

		for key := range oldUser.Extra {
			if _, exists := newUser.Extra[key]; !exists {
				extraChanges[key] = map[string]interface{}{
					"from": oldUser.Extra[key],
					"to":   nil,
				}
			}
		}

		if len(extraChanges) > 0 {
			changes["extra"] = extraChanges
		}
	}

	return changes
}
