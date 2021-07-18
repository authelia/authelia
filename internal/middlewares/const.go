package middlewares

const jwtIssuer = "Authelia"

const (
	headerXForwardedProto  = "X-Forwarded-Proto"
	headerXForwardedMethod = "X-Forwarded-Method"
	headerXForwardedHost   = "X-Forwarded-Host"
	headerXForwardedURI    = "X-Forwarded-URI"
	headerXOriginalURL     = "X-Original-URL"
	headerXRequestedWith   = "X-Requested-With"
)

const (
	headerValueXRequestedWithXHR = "XMLHttpRequest"
)

const (
	statusTextMovedPermanently  = "Moved Permanently"
	statusTextFound             = "Found"
	statusTextSeeOther          = "See Other"
	statusTextTemporaryRedirect = "Temporary Redirect"
	statusTextPermanentRedirect = "Permanent Redirect"
	statusTextUnauthorized      = "Unauthorized"
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
