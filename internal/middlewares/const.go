package middlewares

import (
	"errors"

	"github.com/valyala/fasthttp"
)

var (
	headerXAutheliaURL = []byte("X-Authelia-URL")

	headerAccept        = []byte(fasthttp.HeaderAccept)
	headerContentLength = []byte(fasthttp.HeaderContentLength)
	headerLocation      = []byte(fasthttp.HeaderLocation)

	headerXForwardedProto = []byte(fasthttp.HeaderXForwardedProto)
	headerXForwardedHost  = []byte(fasthttp.HeaderXForwardedHost)
	headerXForwardedFor   = []byte(fasthttp.HeaderXForwardedFor)
	headerXRequestedWith  = []byte(fasthttp.HeaderXRequestedWith)

	headerXForwardedURI    = []byte("X-Forwarded-URI")
	headerXOriginalURL     = []byte("X-Original-URL")
	headerXOriginalMethod  = []byte("X-Original-Method")
	headerXForwardedMethod = []byte("X-Forwarded-Method")

	headerVary   = []byte(fasthttp.HeaderVary)
	headerOrigin = []byte(fasthttp.HeaderOrigin)

	headerAccessControlAllowCredentials = []byte(fasthttp.HeaderAccessControlAllowCredentials)
	headerAccessControlAllowHeaders     = []byte(fasthttp.HeaderAccessControlAllowHeaders)
	headerAccessControlAllowMethods     = []byte(fasthttp.HeaderAccessControlAllowMethods)
	headerAccessControlAllowOrigin      = []byte(fasthttp.HeaderAccessControlAllowOrigin)
	headerAccessControlMaxAge           = []byte(fasthttp.HeaderAccessControlMaxAge)
	headerAccessControlRequestHeaders   = []byte(fasthttp.HeaderAccessControlRequestHeaders)
	headerAccessControlRequestMethod    = []byte(fasthttp.HeaderAccessControlRequestMethod)
	headerRetryAfter                    = []byte(fasthttp.HeaderRetryAfter)

	headerXContentTypeOptions   = []byte(fasthttp.HeaderXContentTypeOptions)
	headerReferrerPolicy        = []byte(fasthttp.HeaderReferrerPolicy)
	headerXFrameOptions         = []byte(fasthttp.HeaderXFrameOptions)
	headerPragma                = []byte(fasthttp.HeaderPragma)
	headerCacheControl          = []byte(fasthttp.HeaderCacheControl)
	headerContentSecurityPolicy = []byte(fasthttp.HeaderContentSecurityPolicy)

	headerPermissionsPolicy         = []byte("Permissions-Policy")
	headerCrossOriginOpenerPolicy   = []byte("Cross-Origin-Opener-Policy")
	headerCrossOriginEmbedderPolicy = []byte("Cross-Origin-Embedder-Policy")
	headerCrossOriginResourcePolicy = []byte("Cross-Origin-Resource-Policy")
	headerXDNSPrefetchControl       = []byte("X-DNS-Prefetch-Control")
)

const (
	HeaderCacheControlNotStore = "no-store"
	HeaderPragmaNoCache        = "no-cache"
)

var (
	headerValueFalse           = []byte("false")
	headerValueTrue            = []byte("true")
	headerValueOff             = []byte("off")
	headerValueMaxAge          = []byte("100")
	headerValueVary            = []byte("Accept-Encoding, Origin")
	headerValueVaryWildcard    = []byte("Accept-Encoding")
	headerValueOriginWildcard  = []byte("*")
	headerValueZero            = []byte("0")
	headerValueCSPNone         = []byte("default-src 'none'")
	headerValueCSPNoneFormPost = []byte("default-src 'none'; script-src 'sha256-skflBqA90WuHvoczvimLdj49ExKdizFjX2Itd6xKZdU='")
	headerValueCSPSelf         = []byte("default-src 'self'")

	headerValueNoSniff                 = []byte("nosniff")
	headerValueStrictOriginCrossOrigin = []byte("strict-origin-when-cross-origin")
	headerValueDENY                    = []byte("DENY")
	headerValueSameOrigin              = []byte("same-origin")
	headerValueCrossOrigin             = []byte("cross-origin")
	headerValueSameSite                = []byte("same-site")
	headerValueUnsafeNone              = []byte("unsafe-none")
	headerValueRequireCORP             = []byte("require-corp")
	headerValueNoCache                 = []byte(HeaderPragmaNoCache)
	headerValueNoStore                 = []byte(HeaderCacheControlNotStore)
	headerValuePermissionsPolicy       = []byte("accelerometer=(), autoplay=(), camera=(), display-capture=(), geolocation=(), gyroscope=(), keyboard-map=(), magnetometer=(), microphone=(), midi=(), payment=(), picture-in-picture=(), screen-wake-lock=(), sync-xhr=(), xr-spatial-tracking=(), interest-cohort=()")
)

const (
	strProtoHTTPS = "https"
	strProtoHTTP  = "http"
	strSlash      = "/"

	queryArgRedirect    = "rd"
	queryArgAutheliaURL = "authelia_url"
	queryArgToken       = "token"
)

const (
	UserValueKeyBaseURL int8 = iota
	UserValueKeyOpenIDConnectResponseModeFormPost
	UserValueKeyRawURI
)

const (
	UserValueRouterKeyExtAuthzPath = "extauthz"
)

const (
	LogMessageStartupCheckError      = "Error occurred running a startup check"
	LogMessageStartupCheckPerforming = "Performing Startup Check"

	ProviderNameNTP              = "ntp"
	ProviderNameStorage          = "storage"
	ProviderNameUser             = "user"
	ProviderNameNotification     = "notification"
	ProviderNameExpressions      = "expressions"
	ProviderNameDuo              = "duo"
	ProviderNameWebAuthnMetaData = "webauthn-metadata"
)

const (
	ContentTypeApplicationJSON = "application/json; charset=utf-8"
	ContentTypeApplicationJWT  = "application/jwt; charset=utf-8"
)

var (
	protoHTTPS = []byte(strProtoHTTPS)
	protoHTTP  = []byte(strProtoHTTP)

	qryArgRedirect    = []byte(queryArgRedirect)
	qryArgAutheliaURL = []byte(queryArgAutheliaURL)

	headerSeparator = []byte(", ")

	contentTypeTextPlain       = []byte("text/plain; charset=utf-8")
	contentTypeTextHTML        = []byte("text/html; charset=utf-8")
	contentTypeApplicationJSON = []byte(ContentTypeApplicationJSON)
	contentTypeApplicationYAML = []byte("application/yaml; charset=utf-8")
)

const (
	headerValueXRequestedWithXHR = "XMLHttpRequest"
)

var okMessageBytes = []byte("{\"status\":\"OK\"}")

const (
	messageOperationFailed                      = "Operation failed"
	messageIdentityVerificationTokenAlreadyUsed = "The identity verification token has already been used"
	messageIdentityVerificationTokenHasExpired  = "The identity verification token has expired"
	messageIdentityVerificationTokenNotValidYet = "The identity verification token is only valid in the future"
	messageIdentityVerificationTokenSig         = "The identity verification token has an invalid signature"
)

var protoHostSeparator = []byte("://")

var errPasswordPolicyNoMet = errors.New("the supplied password does not met the security policy")
