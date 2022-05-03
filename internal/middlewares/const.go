package middlewares

import (
	"errors"

	"github.com/valyala/fasthttp"
)

var (
	headerAccept        = []byte(fasthttp.HeaderAccept)
	headerContentLength = []byte(fasthttp.HeaderContentLength)

	headerXForwardedProto = []byte(fasthttp.HeaderXForwardedProto)
	headerXForwardedHost  = []byte(fasthttp.HeaderXForwardedHost)
	headerXForwardedFor   = []byte(fasthttp.HeaderXForwardedFor)
	headerXRequestedWith  = []byte(fasthttp.HeaderXRequestedWith)

	headerXForwardedURI    = []byte("X-Forwarded-URI")
	headerXOriginalURL     = []byte("X-Original-URL")
	headerXForwardedMethod = []byte("X-Forwarded-Method")

	headerVary   = []byte(fasthttp.HeaderVary)
	headerAllow  = []byte(fasthttp.HeaderAllow)
	headerOrigin = []byte(fasthttp.HeaderOrigin)

	headerAccessControlAllowCredentials = []byte(fasthttp.HeaderAccessControlAllowCredentials)
	headerAccessControlAllowHeaders     = []byte(fasthttp.HeaderAccessControlAllowHeaders)
	headerAccessControlAllowMethods     = []byte(fasthttp.HeaderAccessControlAllowMethods)
	headerAccessControlAllowOrigin      = []byte(fasthttp.HeaderAccessControlAllowOrigin)
	headerAccessControlMaxAge           = []byte(fasthttp.HeaderAccessControlMaxAge)
	headerAccessControlRequestHeaders   = []byte(fasthttp.HeaderAccessControlRequestHeaders)
	headerAccessControlRequestMethod    = []byte(fasthttp.HeaderAccessControlRequestMethod)

	headerXContentTypeOptions = []byte(fasthttp.HeaderXContentTypeOptions)
	headerReferrerPolicy      = []byte(fasthttp.HeaderReferrerPolicy)
	headerXFrameOptions       = []byte(fasthttp.HeaderXFrameOptions)
	headerPragma              = []byte(fasthttp.HeaderPragma)
	headerCacheControl        = []byte(fasthttp.HeaderCacheControl)
	headerXXSSProtection      = []byte(fasthttp.HeaderXXSSProtection)

	headerPermissionsPolicy = []byte("Permissions-Policy")
)

var (
	headerValueFalse          = []byte("false")
	headerValueTrue           = []byte("true")
	headerValueMaxAge         = []byte("100")
	headerValueVary           = []byte("Accept-Encoding, Origin")
	headerValueVaryWildcard   = []byte("Accept-Encoding")
	headerValueOriginWildcard = []byte("*")
	headerValueZero           = []byte("0")

	headerValueNoSniff                 = []byte("nosniff")
	headerValueStrictOriginCrossOrigin = []byte("strict-origin-when-cross-origin")
	headerValueSameOrigin              = []byte("SAMEORIGIN")
	headerValueNoCache                 = []byte("no-cache")
	headerValueNoStore                 = []byte("no-store")
	headerValueXSSDisabled             = []byte("0")
	headerValueCohort                  = []byte("interest-cohort=()")
)

var (
	protoHTTPS = []byte("https")
	protoHTTP  = []byte("http")

	// UserValueKeyBaseURL is the User Value key where we store the Base URL.
	UserValueKeyBaseURL = []byte("base_url")

	headerSeparator = []byte(", ")
)

const (
	headerValueXRequestedWithXHR = "XMLHttpRequest"
	contentTypeApplicationJSON   = "application/json"
	contentTypeTextHTML          = "text/html"
)

var okMessageBytes = []byte("{\"status\":\"OK\"}")

const (
	messageOperationFailed                      = "Operation failed"
	messageIdentityVerificationTokenAlreadyUsed = "The identity verification token has already been used"
	messageIdentityVerificationTokenHasExpired  = "The identity verification token has expired"
)

var protoHostSeparator = []byte("://")

var errPasswordPolicyNoMet = errors.New("the supplied password does not met the security policy")
