package handlers

import (
	"fmt"
	"net/url"

	"github.com/valyala/fasthttp"

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

func parseAuthzPortalURL(rawURL []byte) (portalURL *url.URL, err error) {
	if rawURL == nil {
		return nil, nil
	}

	return url.ParseRequestURI(string(rawURL))
}

func getAuthzRedirectStatusCode(ctx AuthzContext, method string) (statusCode int) {
	if ctx.IsXHR() || !ctx.AcceptsMIME("text/html") {
		return fasthttp.StatusUnauthorized
	}

	switch method {
	case fasthttp.MethodGet, fasthttp.MethodOptions, fasthttp.MethodHead, "":
		return fasthttp.StatusFound
	default:
		return fasthttp.StatusSeeOther
	}
}

func doAuthzRedirect(ctx AuthzContext, authn *Authn, redirectionURL *url.URL, statusCode int) {
	ctx.GetLogger().Infof(logFmtAuthzRedirect, authn.Object.String(), authn.Method, authn.Username, statusCode, redirectionURL)

	switch authn.Object.Method {
	case fasthttp.MethodHead:
		ctx.SpecialRedirectNoBody(redirectionURL.String(), statusCode)
	default:
		ctx.SpecialRedirect(redirectionURL.String(), statusCode)
	}
}

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

func getSafeAutheliaURL(autheliaURL *url.URL, domain string) (*url.URL, error) {
	switch {
	case utils.HasURIDomainSuffix(autheliaURL, domain):
		return autheliaURL, nil
	default:
		return nil, fmt.Errorf("authelia url '%s' is not valid for detected domain '%s' as the url does not have the domain as a suffix", autheliaURL.String(), domain)
	}
}
