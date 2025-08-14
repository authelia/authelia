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
	// Create a map for quick lookup of UserInfo by username.
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

	if !reflect.DeepEqual(original.ObjectClasses, changes.ObjectClasses) {
		changeLog["object_classes"] = map[string]interface{}{
			"from": original.ObjectClasses,
			"to":   changes.ObjectClasses,
		}
	}

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

func SortedCopy(s []string) []string {
	c := make([]string, len(s))
	c = append(c, s...)
	sort.Strings(c)

	return c
}
