package middlewares

const jwtIssuer = "Authelia"

const (
	headerXForwardedProto  = "X-Forwarded-Proto"
	headerXForwardedMethod = "X-Forwarded-Method"
	headerXForwardedHost   = "X-Forwarded-Host"
	headerXForwardedURI    = "X-Forwarded-URI"

	headerXOriginalURL   = "X-Original-URL"
	headerXRequestedWith = "X-Requested-With"

	headerOrigin = "Origin"
	headerVary   = "Vary"

	headerAccessControlRequestHeaders = "Access-Control-Request-Headers"
	headerAccessControlRequestMethod  = "Access-Control-Request-Method"

	headerAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	headerAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	headerAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	headerAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	headerAccessControlMaxAge           = "Access-Control-Max-Age"
)

const (
	headerValueXRequestedWithXHR = "XMLHttpRequest"
)

const (
	contentTypeApplicationJSON = "application/json"
	contentTypeTextHTML        = "text/html"
)

var okMessageBytes = []byte("{\"status\":\"OK\"}")

const (
	messageOperationFailed                      = "Operation failed"
	messageIdentityVerificationTokenAlreadyUsed = "The identity verification token has already been used"
	messageIdentityVerificationTokenHasExpired  = "The identity verification token has expired"
)

var protoHostSeparator = []byte("://")
