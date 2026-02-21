package handlers

import (
	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

func friendlyMethod(m string) (fm string) {
	switch m {
	case "":
		return "unknown"
	default:
		return m
	}
}

func friendlyUsername(username string) (fusername string) {
	switch username {
	case "":
		return anonymous
	default:
		return username
	}
}

func isAuthzResult(level authentication.Level, required authorization.Level, ruleHasSubject bool) AuthzResult {
	switch {
	case required == authorization.Bypass:
		return AuthzResultAuthorized
	case required == authorization.Denied && (level != authentication.NotAuthenticated || !ruleHasSubject):
		// If the user is not anonymous, it means that we went through all the rules related to that user identity and
		// can safely conclude their access is actually forbidden. If a user is anonymous however this is not actually
		// possible without some more advanced logic.
		return AuthzResultForbidden
	case required == authorization.OneFactor && level >= authentication.OneFactor,
		required == authorization.TwoFactor && level >= authentication.TwoFactor:
		return AuthzResultAuthorized
	default:
		return AuthzResultUnauthorized
	}
}

// generateVerifySessionHasUpToDateProfileTraceLogs is used to generate trace logs only when trace logging is enabled.
// The information calculated in this function is completely useless other than trace for now.
func generateVerifySessionHasUpToDateProfileTraceLogs(ctx AuthzContext, userSession *session.UserSession,
	details *authentication.UserDetails) {
	groupsAdded, groupsRemoved := utils.StringSlicesDelta(userSession.Groups, details.Groups)
	emailsAdded, emailsRemoved := utils.StringSlicesDelta(userSession.Emails, details.Emails)
	nameDelta := userSession.DisplayName != details.DisplayName

	fields := map[string]any{"username": userSession.Username}
	msg := "User session groups are current"

	if len(groupsAdded) != 0 || len(groupsRemoved) != 0 {
		if len(groupsAdded) != 0 {
			fields["added"] = groupsAdded
		}

		if len(groupsRemoved) != 0 {
			fields["removed"] = groupsRemoved
		}

		msg = "User session groups were updated"
	}

	ctx.GetLogger().WithFields(fields).Trace(msg)

	if len(emailsAdded) != 0 || len(emailsRemoved) != 0 {
		if len(emailsAdded) != 0 {
			fields["added"] = emailsAdded
		} else {
			delete(fields, "added")
		}

		if len(emailsRemoved) != 0 {
			fields["removed"] = emailsRemoved
		} else {
			delete(fields, "removed")
		}

		msg = "User session emails were updated"
	} else {
		msg = "User session emails are current"

		delete(fields, "added")
		delete(fields, "removed")
	}

	ctx.GetLogger().WithFields(fields).Trace(msg)

	if nameDelta {
		ctx.GetLogger().
			WithFields(map[string]any{
				"username": userSession.Username,
				"before":   userSession.DisplayName,
				"after":    details.DisplayName,
			}).
			Trace("User session display name updated")
	} else {
		ctx.GetLogger().Trace("User session display name is current")
	}
}
