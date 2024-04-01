package oidc

import (
	"encoding/json"
	"fmt"
	"net/url"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/utils"
)

func NewClaimRequests(form url.Values) (claims *ClaimRequests, err error) {
	var raw string

	if raw = form.Get(FormParameterClaims); len(raw) == 0 {
		return nil, nil
	}

	claims = &ClaimRequests{}

	if err = json.Unmarshal([]byte(raw), claims); err != nil {
		return nil, oauthelia2.ErrInvalidRequest.WithHint("The OAuth 2.0 client included a malformed 'claims' parameter in the authorization request.").WithWrap(err).WithDebugf("Error occurred attempting to parse the 'claims' parameter: %+v.", err)
	}

	return claims, nil
}

// ClaimRequests is a request for a particular set of claims.
type ClaimRequests struct {
	IDToken  map[string]*ClaimsRequest `json:"id_token,omitempty"`
	UserInfo map[string]*ClaimsRequest `json:"userinfo,omitempty"`
}

// ClaimsRequest is a request for a particular claim.
type ClaimsRequest struct {
	Essential bool  `json:"essential,omitempty"`
	Value     any   `json:"value,omitempty"`
	Values    []any `json:"values,omitempty"`
}

// Matches is a convenience function which tests if a particular value matches this claims request.
//
//nolint:gocyclo
func (r *ClaimsRequest) Matches(value any) (match bool) {
	if r == nil {
		return false
	}

	switch t := value.(type) {
	case int:
		if r.Value != nil {
			if float64(t) != r.Value && t != r.Value {
				return false
			}
		}
	case int64:
		if r.Value != nil {
			if float64(t) != r.Value && t != r.Value {
				return false
			}
		}

		if r.Values != nil {
			found := false

			for _, v := range r.Values {
				if float64(t) == v || t == v {
					found = true

					break
				}
			}

			if !found {
				return false
			}
		}
	case float64:
		if r.Value != nil {
			if t != r.Value {
				return false
			}
		}

		if r.Values != nil {
			found := false

			for _, v := range r.Values {
				if t == v {
					found = true

					break
				}
			}

			if !found {
				return false
			}
		}
	case string:
		if r.Value != nil {
			if t != r.Value {
				return false
			}
		}

		if r.Values != nil {
			found := false

			for _, v := range r.Values {
				if t == v {
					found = true

					break
				}
			}

			if !found {
				return false
			}
		}
	case []string:
		if r.Value != nil {
			if !utils.IsStringInSlice(fmt.Sprintf("%s", value), t) {
				return false
			}
		}

		if r.Values != nil {
			found := false

		outer:
			for _, v := range r.Values {
				for _, w := range t {
					if v == w {
						found = true

						break outer
					}
				}
			}

			if !found {
				return false
			}
		}
	}

	return true
}
