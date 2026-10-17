// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func handleOAuth2AuthorizationBounce(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, r *http.Request) (handled bool) {
	path := ctx.BasePath() + oidc.EndpointPathAuthorization

	// The marker is never honored on a cross-site request as the request the bounce page itself makes never is, so that a
	// separate cross-site request which arrives while a marker is outstanding is still bounced.
	if _, err := r.Cookie(cookieNameOAuth2AuthorizationBounce); err == nil && isOAuth2AuthorizationBounceFollowUp(r) {
		http.SetCookie(rw, newOAuth2AuthorizationBounceCookie(path, -1))

		return false
	}

	if r.Method != http.MethodGet {
		return false
	}

	if !isOAuth2AuthorizationBounceRequired(ctx, r) {
		return false
	}

	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return false
	}

	data := oauth2AuthorizationBounceData{
		Target: (&url.URL{Path: path, RawQuery: query.Encode()}).String(),
	}

	page := &bytes.Buffer{}

	if err = tmplOAuth2AuthorizationBounce.Execute(page, data); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred rendering the authorization request bounce page")

		return false
	}

	http.SetCookie(rw, newOAuth2AuthorizationBounceCookie(path, cookieMaxAgeOAuth2AuthorizationBounce))

	rw.Header().Set(fasthttp.HeaderContentType, valueContentTypeTextHTML)
	rw.Header().Set(fasthttp.HeaderCacheControl, middlewares.HeaderCacheControlNotStore)
	rw.Header().Set(fasthttp.HeaderReferrerPolicy, valueReferrerPolicyNoReferrer)

	rw.WriteHeader(http.StatusOK)

	if _, err = rw.Write(page.Bytes()); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred writing the authorization request bounce page")
	}

	return true
}

func isOAuth2AuthorizationBounceRequired(ctx *middlewares.AutheliaCtx, r *http.Request) bool {
	switch site := r.Header.Get(headerSecFetchSite); {
	case site == "":
		// The user agent does not send the fetch metadata request headers so whether the request is cross-site can't be
		// determined, and the bounce is performed if the session cookie is absent as it's harmless when it's not needed.
	case !strings.EqualFold(site, valueSecFetchSiteCrossSite),
		!strings.EqualFold(r.Header.Get(headerSecFetchMode), valueSecFetchModeNavigate),
		!strings.EqualFold(r.Header.Get(headerSecFetchDest), valueSecFetchDestDocument):
		return false
	}

	provider, err := ctx.GetSessionProvider()
	if err != nil {
		return false
	}

	_, err = r.Cookie(provider.GetConfig().Name)

	return err != nil
}

func isOAuth2AuthorizationBounceFollowUp(r *http.Request) bool {
	return !strings.EqualFold(r.Header.Get(headerSecFetchSite), valueSecFetchSiteCrossSite)
}

func newOAuth2AuthorizationBounceCookie(path string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     cookieNameOAuth2AuthorizationBounce,
		Value:    cookieValueOAuth2AuthorizationBounce,
		Path:     path,
		MaxAge:   maxAge,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}
