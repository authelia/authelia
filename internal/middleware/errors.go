package middleware

import "errors"

var errMissingXForwardedHost = errors.New("Missing header X-Forwarded-Host")
var errMissingXForwardedProto = errors.New("Missing header X-Forwarded-Proto")
