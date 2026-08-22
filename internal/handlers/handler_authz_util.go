package handlers

import (
	"fmt"
	"net/url"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/utils"
)

// newAnonymousUserDetails returns an empty extended user details value. The embedded *UserDetails is initialized as
// consumers dereference the promoted fields without a nil check.
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
