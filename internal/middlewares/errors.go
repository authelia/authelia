package middlewares

import "errors"

var (
	// ErrMissingXForwardedProto is returned on methods which require an X-Forwarded-Proto header.
	ErrMissingXForwardedProto = errors.New("missing required X-Forwarded-Proto header")

	// ErrMissingXForwardedHost is returned on methods which require an X-Forwarded-Host header.
	ErrMissingXForwardedHost = errors.New("missing required X-Forwarded-Host header")

	// ErrMissingXOriginalURL is returned on methods which require an X-Original-URL header.
	ErrMissingXOriginalURL = errors.New("missing required X-Original-URL header")
)
