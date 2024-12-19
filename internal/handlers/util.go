package handlers

import (
	"errors"
	"fmt"
	"regexp"
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

	eventEmailActionPasswordResetPrefix = "your"
	eventEmailActionPasswordReset       = "Password Reset"
	eventEmailActionPasswordResetSuffix = "was successful."

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
	usernameAndGroupRegex  = `^[a-zA-Z0-9-_,]{1,100}$`
)

func ValidatePrintableUnicodeString(input string) error {
	if strings.Contains(input, `@`) {
		if err := ValidateEmailString(input); err != nil {
			return err
		}
	}

	var regex = regexp.MustCompile(printableUnicodeRegexp) //nolint:forbidigo
	if !regex.MatchString(input) {
		return errors.New(errNotValidPrintableUnicode)
	}

	return nil
}

func ValidateEmailString(input string) error {
	var regex = regexp.MustCompile(emailRegex)
	if !regex.MatchString(input) {
		return errors.New(errNotValidEmail)
	}

	return nil
}

func ValidateGroup(input string) error {
	var regex = regexp.MustCompile(usernameAndGroupRegex)
	if !regex.MatchString(input) {
		return errors.New("groups must only contain letters, numbers, hyphens, commas and underscores")
	}

	return nil
}

func ValidateUsername(input string) error {
	if strings.Contains(input, `@`) {
		if err := ValidateEmailString(input); err != nil {
			return err
		}

		return nil
	}

	var regex = regexp.MustCompile(usernameAndGroupRegex)
	if !regex.MatchString(input) {
		return errors.New("username must only contain letters, numbers, hyphens, commas and underscores or a valid email")
	}

	return nil
}

func UserIsAdmin(ctx *middlewares.AutheliaCtx, userGroups []string) bool {
	return slices.Contains(userGroups, ctx.Configuration.Administration.AdminGroup)
}
