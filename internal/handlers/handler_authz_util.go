// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/utils"
)

var authzSessionUserAttributes = []string{
	expression.AttributeUserUsername,
	expression.AttributeUserGroups,
	expression.AttributeUserDisplayName,
	expression.AttributeUserEmail,
	expression.AttributeUserEmails,
	expression.AttributeUserEmailsExtra,
	expression.AttributeUserEmailVerified,
	expression.AttributeUserUpdatedAt,
}

func authzHeadersRequireExtendedUserDetails(headers []AuthzHeader) (extended bool) {
	for _, header := range headers {
		if !utils.IsStringInSlice(header.Attribute, authzSessionUserAttributes) {
			return true
		}
	}

	return false
}

func authzHeaderValue(object any) (value string) {
	switch v := object.(type) {
	case nil:
		return ""
	case string:
		return v
	case []string:
		return strings.Join(v, ",")
	case []any:
		values := make([]string, len(v))

		for i, item := range v {
			values[i] = authzHeaderValue(item)
		}

		return strings.Join(values, ",")
	case bool:
		return strconv.FormatBool(v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func newAnonymousUserDetails() authentication.UserDetailsExtended {
	return authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{}}
}

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

func getSafeAutheliaURL(autheliaURL *url.URL, domain string) (*url.URL, error) {
	switch {
	case utils.HasURIDomainSuffix(autheliaURL, domain):
		return autheliaURL, nil
	default:
		return nil, fmt.Errorf("authelia url '%s' is not valid for detected domain '%s' as the url does not have the domain as a suffix", autheliaURL.String(), domain)
	}
}
