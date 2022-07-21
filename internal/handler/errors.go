package handler

import (
	"errors"
)

var (
	errMissingAuthorizationHeaderSchemeBasicForced = errors.New("authorization scheme basic was enforced but the Authorization header is missing")
)
