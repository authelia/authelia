package handlers

import (
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"sort"
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

	localPart := parts[0]
	domain := parts[1]

	if len(localPart) <= 2 {
		return strings.Repeat("*", len(localPart)) + "@" + domain
	}

	first := string(localPart[0])
	last := string(localPart[len(localPart)-1])
	middle := strings.Repeat("*", len(localPart)-2)

	return first + middle + last + "@" + domain
}

// MergeUserInfoAndDetails combines the list of attributes in userInfo with the list of users in users.
func MergeUserInfoAndDetails(userInfo []model.UserInfo, users []authentication.UserDetails) []model.UserInfo {
	userDetailsMap := make(map[string]authentication.UserDetails)
	userInfoMap := make(map[string]bool)

	for _, user := range users {
		userDetailsMap[user.Username] = user
	}

	for i, info := range userInfo {
		if details, ok := userDetailsMap[info.Username]; ok {
			userInfo[i].DisplayName = details.DisplayName
			userInfo[i].Emails = details.Emails
			userInfo[i].Groups = details.Groups
			userInfoMap[info.Username] = true
		}
	}

	for _, user := range users {
		if _, exists := userInfoMap[user.Username]; !exists {
			userInfo = append(userInfo, model.UserInfo{
				Username:    user.Username,
				DisplayName: user.DisplayName,
				Emails:      user.Emails,
				Groups:      user.Groups,
			})
		}
	}

	return userInfo
}

const (
	printableUnicodeRegexp = `^[\pL\pM\pN\pP\pS\s]{1,100}$`
	emailRegex             = `^[a-zA-Z0-9+._~!#$%&'*/=?^{|}-]+@[a-zA-Z0-9-.]+\.[a-zA-Z0-9-]+$`
	usernameAndGroupRegex  = `^[a-zA-Z0-9+._\-]{1,100}$`
)

func ValidatePrintableUnicodeString(input string) bool {
	var regex = regexp.MustCompile(printableUnicodeRegexp) //nolint:forbidigo

	return regex.MatchString(input)
}

func ValidateEmailString(input string) bool {
	var regex = regexp.MustCompile(emailRegex)

	return regex.MatchString(input)
}

func ValidateGroups(input []string) (bool, string) {
	for _, group := range input {
		if !ValidateGroup(group) {
			return false, group
		}
	}

	return true, ""
}

func ValidateGroup(input string) bool {
	var regex = regexp.MustCompile(usernameAndGroupRegex)

	return regex.MatchString(input)
}

func ValidateUsername(input string) bool {
	if strings.Contains(input, `@`) {
		return ValidateEmailString(input)
	}

	return ValidatePrintableUnicodeString(input)
}

func UserIsAdmin(ctx *middlewares.AutheliaCtx, userGroups []string) bool {
	return slices.Contains(userGroups, ctx.Configuration.Administration.AdminGroup)
}

func GenerateUserChangeLog(original *authentication.UserDetails, changes *changeUserRequestBody) []string {
	var modifications []string

	if original.DisplayName != changes.DisplayName {
		modifications = append(modifications,
			fmt.Sprintf("display name from '%s' to '%s'", original.DisplayName, changes.DisplayName))
	}

	if original.Emails[0] != changes.Email {
		modifications = append(modifications,
			fmt.Sprintf("email from '%s' to '%s'", original.Emails[0], changes.Email))
	}

	if !reflect.DeepEqual(original.Groups, changes.Groups) {
		modifications = append(modifications,
			fmt.Sprintf("groups from [%v] to [%v]", strings.Join(original.Groups, ", "), strings.Join(changes.Groups, ", ")))
	}

	if changes.Password != "" {
		modifications = append(modifications, "password")
	}

	return modifications
}

func SortedCopy(s []string) []string {
	c := make([]string, len(s))
	c = append(c, s...)
	sort.Strings(c)

	return c
}
