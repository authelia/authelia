package middlewares

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
var validOverrideAssets = []string{"favicon.ico", "logo.png"}
