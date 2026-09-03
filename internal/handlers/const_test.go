package handlers

import (
	"net/url"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	testRequestMethods = []string{
		fasthttp.MethodOptions, fasthttp.MethodHead, fasthttp.MethodGet,
		fasthttp.MethodDelete, fasthttp.MethodPatch, fasthttp.MethodPost,
		fasthttp.MethodPut, fasthttp.MethodConnect, fasthttp.MethodTrace,
	}

	testXHR = map[string]bool{
		testWithoutAccept: false,
		testWithXHRHeader: true,
	}
)

const (
	testXOriginalMethod = "X-Original-Method"
	testXOriginalUrl    = "X-Original-URL"
	testBypass          = "bypass"
	testWithoutAccept   = "WithoutAccept"
	testWithXHRHeader   = "WithXHRHeader"
)

//nolint:gosec // Test Credentials.
const (
	testBASE32TOTPSecret = "JVHFEUBXJ5CUWN2GGZGDMTKSJNMEQN2YGRJUQM2OKRHECR2QKJGFGRSQJVEVUT2HII2FQSJTKNIVQSCPIJIQ===="
	testJWTSecret        = "abc"
)

const (
	testInactivity           = time.Second * 10
	testRedirectionURLString = "https://www.example.com"
	testUsername             = "john"
	testDisplayName          = "john"
	testEmail                = "john@example.com"
	exampleDotCom            = "example.com"
)

const (
	testValue = "test"
)

var (
	testRedirectionURL = func() *url.URL {
		u, err := url.ParseRequestURI(testRedirectionURLString)
		if err != nil {
			panic(err)
		}

		return u
	}()
)

//nolint:gosec // Test Credentials.
const (
	testOIDCClientSecretDigest = "$plaintext$client-secret"
	testOIDCClientSecretValue  = "client-secret"
)

const (
	testOIDCFormParameterGrantType    = "grant_type"
	testOIDCFormParameterClientSecret = "client_secret"
	testOIDCFormParameterCode         = "code"
	testOIDCFormParameterToken        = "token"
)

const (
	testOIDCScopeBearerAuthz      = "authelia.bearer.authz"
	testOIDCFormParameterAudience = "audience"
)

const testOIDCClientCredentialsID = "client-credentials"

const (
	testOIDCAuthorizationCodeID = "authorization-code"
	//nolint:gosec // This is a redirection URI, not a credential.
	testOIDCRedirectURI = "https://app.example.com/oidc/callback"
)

const testOIDCDeviceCodeID = "device-code"

const testOIDCKeyID = "rsa-default"

const testOIDCClaimsPolicyMerged = "merged-audience"

var testOIDCPreConfiguredDuration = time.Hour * 24
