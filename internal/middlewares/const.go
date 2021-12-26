package middlewares

import (
	"github.com/valyala/fasthttp"
)

var (
	headerXForwardedProto = []byte(fasthttp.HeaderXForwardedProto)
	headerXForwardedHost  = []byte(fasthttp.HeaderXForwardedHost)
	headerXForwardedFor   = []byte(fasthttp.HeaderXForwardedFor)
	headerXRequestedWith  = []byte(fasthttp.HeaderXRequestedWith)
	headerAccept          = []byte(fasthttp.HeaderAccept)

	headerXForwardedURI    = []byte("X-Forwarded-URI")
	headerXOriginalURL     = []byte("X-Original-URL")
	headerXForwardedMethod = []byte("X-Forwarded-Method")

	headerVary                          = []byte(fasthttp.HeaderVary)
	headerOrigin                        = []byte(fasthttp.HeaderOrigin)
	headerAccessControlAllowCredentials = []byte(fasthttp.HeaderAccessControlAllowCredentials)
	headerAccessControlAllowHeaders     = []byte(fasthttp.HeaderAccessControlAllowHeaders)
	headerAccessControlAllowMethods     = []byte(fasthttp.HeaderAccessControlAllowMethods)
	headerAccessControlAllowOrigin      = []byte(fasthttp.HeaderAccessControlAllowOrigin)
	headerAccessControlMaxAge           = []byte(fasthttp.HeaderAccessControlMaxAge)
	headerAccessControlRequestHeaders   = []byte(fasthttp.HeaderAccessControlRequestHeaders)
)

var (
	headerValueFalse     = []byte("false")
	headerValueMaxAge    = []byte("100")
	headerValueVary      = []byte("Accept-Encoding, Origin")
	headerValueMethodGET = []byte("GET")
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
var validOverrideAssets = []string{"favicon.ico", "logo.png"}
