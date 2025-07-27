package handlers

import (
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
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
func mergeUserData(existing, updates *authentication.UserDetailsExtended) *authentication.UserDetailsExtended {
	merged := &authentication.UserDetailsExtended{
		GivenName:         existing.GivenName,
		FamilyName:        existing.FamilyName,
		MiddleName:        existing.MiddleName,
		Nickname:          existing.Nickname,
		CommonName:        existing.CommonName,
		Profile:           existing.Profile,
		Picture:           existing.Picture,
		Website:           existing.Website,
		Gender:            existing.Gender,
		Birthdate:         existing.Birthdate,
		ZoneInfo:          existing.ZoneInfo,
		Locale:            existing.Locale,
		PhoneNumber:       existing.PhoneNumber,
		PhoneExtension:    existing.PhoneExtension,
		Address:           existing.Address,
		DN:                existing.DN,
		ObjectClass:       make([]string, len(existing.ObjectClass)),
		BackendAttributes: make(map[string]interface{}),
		Disabled:          existing.Disabled,
		Extra:             make(map[string]any),
		UserDetails: &authentication.UserDetails{
			Username:    existing.UserDetails.Username,
			DisplayName: existing.UserDetails.DisplayName,
			Emails:      make([]string, len(existing.UserDetails.Emails)),
			Groups:      make([]string, len(existing.UserDetails.Groups)),
		},
	}

	copy(merged.ObjectClass, existing.ObjectClass)
	copy(merged.UserDetails.Emails, existing.UserDetails.Emails)
	copy(merged.UserDetails.Groups, existing.UserDetails.Groups)

	for k, v := range existing.BackendAttributes {
		merged.BackendAttributes[k] = v
	}

	for k, v := range existing.Extra {
		merged.Extra[k] = v
	}

	if updates.UserDetails != nil {
		if updates.UserDetails.DisplayName != "" {
			merged.UserDetails.DisplayName = updates.UserDetails.DisplayName
		}

		if len(updates.UserDetails.Emails) > 0 {
			merged.UserDetails.Emails = make([]string, len(updates.UserDetails.Emails))
			copy(merged.UserDetails.Emails, updates.UserDetails.Emails)
		}

		if len(updates.UserDetails.Groups) > 0 {
			merged.UserDetails.Groups = make([]string, len(updates.UserDetails.Groups))
			copy(merged.UserDetails.Groups, updates.UserDetails.Groups)
		}
	}

	if updates.Password != "" {
		merged.Password = updates.Password
	}

	if updates.GivenName != "" {
		merged.GivenName = updates.GivenName
	}

	if updates.FamilyName != "" {
		merged.FamilyName = updates.FamilyName
	}

	if updates.CommonName != "" {
		merged.CommonName = updates.CommonName
	}

	if updates.DN != "" {
		merged.DN = updates.DN
	}

	if len(updates.ObjectClass) > 0 {
		merged.ObjectClass = make([]string, len(updates.ObjectClass))
		copy(merged.ObjectClass, updates.ObjectClass)
	}

	if len(updates.BackendAttributes) > 0 {
		for k, v := range updates.BackendAttributes {
			merged.BackendAttributes[k] = v
		}
	}

	merged.Disabled = updates.Disabled

	return merged
}

func UserIsAdmin(ctx *middlewares.AutheliaCtx, userGroups []string) bool {
	return slices.Contains(userGroups, ctx.Configuration.Administration.AdminGroup)
}

func GenerateUserChangeLog(original *authentication.UserDetailsExtended, changes *authentication.UserDetailsExtended) []string {
	var modifications []string

	if original.DisplayName != changes.DisplayName {
		modifications = append(modifications,
			fmt.Sprintf("display name from '%s' to '%s'", original.DisplayName, changes.DisplayName))
	}

	if original.Emails[0] != changes.Emails[0] {
		modifications = append(modifications,
			fmt.Sprintf("email from '%s' to '%s'", original.Emails[0], changes.Emails[0]))
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
