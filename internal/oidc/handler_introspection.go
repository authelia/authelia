package oidc

import (
	"context"
	"net/http"
	"strings"

	"github.com/ory/fosite"
	"github.com/ory/x/errorsx"
	"github.com/valyala/fasthttp"
	"golang.org/x/text/language"
)

// NewIntrospectionRequest shadows the fosite version of this function.
func (p *OpenIDConnectProvider) NewIntrospectionRequest(ctx context.Context, r *http.Request, session fosite.Session) (_ fosite.IntrospectionResponder, err error) {
	ctx = context.WithValue(ctx, fosite.RequestContextKey, r)

	if r.Method != fasthttp.MethodPost {
		return &IntrospectionResponse{Active: false}, errorsx.WithStack(fosite.ErrInvalidRequest.WithHintf("HTTP method is '%s' but expected 'POST'.", r.Method))
	} else if err := r.ParseMultipartForm(1 << 20); err != nil && err != http.ErrNotMultipart {
		return &IntrospectionResponse{Active: false}, errorsx.WithStack(fosite.ErrInvalidRequest.WithHint("Unable to parse HTTP body, make sure to send a properly formatted form request body.").WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
	} else if len(r.PostForm) == 0 {
		return &IntrospectionResponse{Active: false}, errorsx.WithStack(fosite.ErrInvalidRequest.WithHint("The POST body can not be empty."))
	}

	token := r.PostForm.Get(FormParameterToken)
	tokenTypeHint := r.PostForm.Get(FormParameterTokenTypeHint)

	var client fosite.Client

	if client, err = p.handleNewIntrospectionRequestClientAuthentication(ctx, r, session, token); err != nil {
		return &IntrospectionResponse{Active: false}, err
	}

	var (
		ar  fosite.AccessRequester
		use fosite.TokenUse
	)

	if use, ar, err = p.IntrospectToken(ctx, token, fosite.TokenUse(tokenTypeHint), session, fosite.RemoveEmpty(strings.Split(r.PostForm.Get(FormParameterScope), " "))...); err != nil {
		return &IntrospectionResponse{Active: false}, errorsx.WithStack(fosite.ErrInactiveToken.WithHint("An introspection strategy indicated that the token is inactive.").WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
	}

	accessTokenType := ""

	if use == fosite.AccessToken {
		accessTokenType = fosite.BearerAccessToken
	}

	return &IntrospectionResponse{
		Client:          client,
		Active:          true,
		AccessRequester: ar,
		TokenUse:        use,
		AccessTokenType: accessTokenType,
	}, nil
}

func (p *OpenIDConnectProvider) handleNewIntrospectionRequestClientAuthentication(ctx context.Context, r *http.Request, session fosite.Session, token string) (client fosite.Client, err error) {
	if clientToken := fosite.AccessTokenFromRequest(r); clientToken != "" {
		if token == clientToken {
			return nil, errorsx.WithStack(fosite.ErrRequestUnauthorized.WithHint("Bearer and introspection token are identical."))
		}

		var (
			ar  fosite.AccessRequester
			use fosite.TokenUse
		)

		if use, ar, err = p.IntrospectToken(ctx, clientToken, fosite.AccessToken, session.Clone()); err != nil {
			return nil, errorsx.WithStack(fosite.ErrRequestUnauthorized.WithHint("HTTP Authorization header missing, malformed, or credentials used are invalid.").WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
		} else if use != "" && use != fosite.AccessToken {
			return nil, errorsx.WithStack(fosite.ErrRequestUnauthorized.WithHintf("HTTP Authorization header did not provide a token of type 'access_token', got type '%s'.", use))
		}

		client = ar.GetClient()
	} else {
		var (
			clientID, clientSecret string
			ok                     bool
		)

		switch clientID, clientSecret, ok, err = clientCredentialsFromBasicAuth(r.Header); {
		case err != nil:
			return nil, errorsx.WithStack(fosite.ErrRequestUnauthorized.WithWrap(err).WithHint("HTTP Authorization header malformed."))
		case !ok:
			return nil, errorsx.WithStack(fosite.ErrRequestUnauthorized.WithHint("HTTP Authorization header missing."))
		}

		if client, err = p.Store.GetClient(ctx, clientID); err != nil {
			return nil, errorsx.WithStack(fosite.ErrRequestUnauthorized.WithHint("Unable to find OAuth 2.0 Client from HTTP basic authorization header.").WithWrap(err).WithDebug(ErrorToDebugRFC6749Error(err).Error()))
		}

		// Enforce client authentication.
		if err = p.checkClientSecret(ctx, client, []byte(clientSecret)); err != nil {
			return nil, errorsx.WithStack(fosite.ErrRequestUnauthorized.WithHint("OAuth 2.0 Client credentials are invalid."))
		}
	}

	return client, nil
}

// IntrospectionResponse is a copy of the fosite.IntrospectionResponse which also includes a fosite.Client to satisfy
// the ClientRequesterResponder interface so we can perform the JWT Response for OAuth 2.0 Token Introspection with the
// correct audience.
type IntrospectionResponse struct {
	Client          fosite.Client          `json:"-"`
	Active          bool                   `json:"active"`
	AccessRequester fosite.AccessRequester `json:"extra"`
	TokenUse        fosite.TokenUse        `json:"token_use,omitempty"`
	AccessTokenType string                 `json:"token_type,omitempty"`
	Lang            language.Tag           `json:"-"`
}

func (r *IntrospectionResponse) IsActive() bool {
	return r.Active
}

func (r *IntrospectionResponse) GetClient() fosite.Client {
	return r.Client
}

func (r *IntrospectionResponse) GetAccessRequester() fosite.AccessRequester {
	return r.AccessRequester
}

func (r *IntrospectionResponse) GetTokenUse() fosite.TokenUse {
	return r.TokenUse
}

func (r *IntrospectionResponse) GetAccessTokenType() string {
	return r.AccessTokenType
}

var (
	_ fosite.IntrospectionResponder = (*IntrospectionResponse)(nil)
)
