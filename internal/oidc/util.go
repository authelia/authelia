package oidc

import (
	"strings"

	"github.com/ory/fosite"
)

// IsPushedAuthorizedRequest returns true if the requester has a PushedAuthorizationRequest redirect_uri value.
func IsPushedAuthorizedRequest(r fosite.Requester, prefix string) bool {
	return strings.HasPrefix(r.GetRequestForm().Get(FormParameterRequestURI), prefix)
}
