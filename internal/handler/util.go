package handler

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middleware"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

func isURLUnderProtectedDomain(u *url.URL, domain string) bool {
	hostname := u.Hostname()

	if hostname == domain {
		return true
	}

	parts := strings.SplitN(hostname, ".", 2)

	if len(parts) != 2 {
		return false
	}

	return parts[1] == domain
}

func isSchemeSecure(u *url.URL) bool {
	return u.Scheme == "https" || u.Scheme == "wss"
}

func headerAuthorizationParseBasic(value []byte) (username, password string, err error) {
	if bytes.Equal(value, valueEmpty) {
		return "", "", fmt.Errorf("header is malformed: empty value")
	}

	parts := strings.SplitN(string(value), " ", 2)

	if len(parts) != 2 {
		return "", "", fmt.Errorf("header is malformed: does not appear to have a scheme")
	}

	if parts[0] != headerAuthorizationSchemeBasic {
		return "", "", fmt.Errorf("header is malformed: unexpected scheme '%s': expected scheme '%s'", parts[0], headerAuthorizationSchemeBasic)
	}

	var content []byte

	if content, err = base64.StdEncoding.DecodeString(parts[1]); err != nil {
		return "", "", fmt.Errorf("header is malformed: could not decode credentials: %w", err)
	}

	strContent := string(content)
	s := strings.IndexByte(strContent, ':')

	if s < 1 {
		return "", "", fmt.Errorf("header is malformed: format of header must be <user>:<password> but either doesn't have a colon or username")
	}

	return strContent[:s], strContent[s+1:], nil
}

func isAuthorizationMatching(levelRequired authorization.Level, levelCurrent authentication.Level) authorizationMatching {
	switch {
	case levelRequired == authorization.Bypass:
		return Authorized
	case levelRequired == authorization.Denied && levelCurrent != authentication.NotAuthenticated:
		// If the user is not anonymous, it means that we went through all the rules related to that user identity and
		// can safely conclude their access is actually forbidden. If a user is anonymous however this is not actually
		// possible without some more advanced logic.
		return Forbidden
	case levelRequired == authorization.OneFactor && levelCurrent >= authentication.OneFactor,
		levelRequired == authorization.TwoFactor && levelCurrent >= authentication.TwoFactor:
		return Authorized
	default:
		return NotAuthorized
	}
}

// setForwardedHeaders set the forwarded User, Groups, Name and Email headers.
func setForwardedHeaders(headers *fasthttp.ResponseHeader, authn *Authentication) {
	if authn.Details.Username != "" {
		headers.SetBytesK(headerRemoteUser, authn.Details.Username)
		headers.SetBytesK(headerRemoteGroups, strings.Join(authn.Details.Groups, ","))
		headers.SetBytesK(headerRemoteName, authn.Details.DisplayName)

		if len(authn.Details.Emails) != 0 {
			headers.SetBytesK(headerRemoteEmail, authn.Details.Emails[0])
		} else {
			headers.SetBytesK(headerRemoteEmail, "")
		}
	}
}

func isSessionInactiveTooLong(ctx *middleware.AutheliaCtx, userSession *session.UserSession, isAnonymous bool) (isInactiveTooLong bool) {
	if userSession.KeepMeLoggedIn || isAnonymous || int64(ctx.Providers.SessionProvider.Inactivity.Seconds()) == 0 {
		return false
	}

	isInactiveTooLong = time.Unix(userSession.LastActivity, 0).Add(ctx.Providers.SessionProvider.Inactivity).Before(ctx.Clock.Now())

	ctx.Logger.Tracef("Inactivity report for user '%s'. Current Time: %d, Last Activity: %d, Maximum Inactivity: %d.", userSession.Username, ctx.Clock.Now().Unix(), userSession.LastActivity, int(ctx.Providers.SessionProvider.Inactivity.Seconds()))

	return isInactiveTooLong
}

func handleVerifyGETRedirectionURL(rd, rm string, targetURL *url.URL, forbidden bool) (redirectionURL *url.URL, err error) {
	if rd == "" {
		return nil, nil
	}

	if redirectionURL, err = url.Parse(rd); err != nil {
		return nil, err
	}

	args := url.Values{}

	if forbidden {
		redirectionURL.Path = path.Join(redirectionURL.Path, "error")
		args = url.Values{
			queryArgStrCode:        []string{strconv.Itoa(fasthttp.StatusForbidden)},
			queryArgStrMessage:     []string{fasthttp.StatusMessage(fasthttp.StatusForbidden)},
			queryArgStrTitle:       []string{"Access Denied"},
			queryArgStrDescription: []string{"Your access to this resource is forbidden"},
			queryArgStrURL:         []string{targetURL.String()},
		}

		if rm != "" {
			args.Set(queryArgStrMethod, rm)
		}
	} else {
		args.Set(queryArgStrRD, targetURL.String())

		if rm != "" {
			args.Set(queryArgStrRM, rm)
		}
	}

	redirectionURL.RawQuery = args.Encode()

	return redirectionURL, nil
}

func getProfileRefreshSettings(config schema.AuthenticationBackendConfiguration) (refresh bool, refreshInterval time.Duration) {
	if config.LDAP != nil {
		if config.RefreshInterval == schema.ProfileRefreshDisabled {
			refresh = false
			refreshInterval = 0
		} else {
			refresh = true

			if config.RefreshInterval != schema.ProfileRefreshAlways {
				// Skip Error Check since validator checks it.
				refreshInterval, _ = utils.ParseDurationString(config.RefreshInterval)
			} else {
				refreshInterval = schema.RefreshIntervalAlways
			}
		}
	}

	return refresh, refreshInterval
}
