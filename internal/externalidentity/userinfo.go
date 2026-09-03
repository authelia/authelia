// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package externalidentity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/hashicorp/go-retryablehttp"

	"authelia.com/provider/oauth2/token/jose"
)

const (
	headerAuthorization = "Authorization"

	mimeApplicationJWT = "application/jwt"

	userinfoResponseLimit = 1024 * 512
)

// UserInfo requests the UserInfo Endpoint with the access token from the token response. It is only requested once
// the ID Token has been validated, and the 'sub' claim of the UserInfo Response must exactly match the validated
// 'sub' claim of the ID Token as the response may otherwise be substituted (OpenID Connect Core 1.0 Section 5.3.2).
// The display claims of the UserInfo Response are only adopted where the ID Token did not provide them. When the
// provider has no UserInfo Endpoint this does nothing.
func (p *OpenIDConnectProvider) UserInfo(ctx context.Context, accessToken string, claims *IdentityClaims) (err error) {
	if err = p.Resolve(ctx); err != nil {
		return err
	}

	if p.endpointUserInfo == "" {
		return nil
	}

	if accessToken == "" {
		return fmt.Errorf("error requesting the userinfo: %w", ErrUserInfoAccessTokenMissing)
	}

	var req *retryablehttp.Request

	if req, err = retryablehttp.NewRequestWithContext(ctx, http.MethodGet, p.endpointUserInfo, nil); err != nil {
		return fmt.Errorf("error requesting the userinfo: %w", err)
	}

	// The UserInfo Response is signed only for a client registered with a 'userinfo_signed_response_alg', so the
	// response is expected in exactly the form the configuration says the client is registered for.
	expected := mimeApplicationJSON

	if p.algUserInfo != "" {
		expected = mimeApplicationJWT
	}

	req.Header.Set(headerAccept, expected)
	req.Header.Set(headerAuthorization, "Bearer "+accessToken)

	var resp *http.Response

	if resp, err = p.client.Do(req); err != nil {
		return fmt.Errorf("error requesting the userinfo: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error requesting the userinfo: %w: the userinfo endpoint returned status code %d", ErrUserInfoResponseInvalid, resp.StatusCode)
	}

	// A response in any other form is rejected rather than accepted as a fallback, as accepting an unsigned response
	// from a provider which is expected to sign it would make the signature optional to anyone able to alter it.
	if mediaType, _, _ := mime.ParseMediaType(resp.Header.Get(headerContentType)); mediaType != expected {
		return fmt.Errorf("error requesting the userinfo: %w: the content type '%s' is not '%s'", ErrUserInfoResponseInvalid, resp.Header.Get(headerContentType), expected)
	}

	var data []byte

	if data, err = io.ReadAll(io.LimitReader(resp.Body, userinfoResponseLimit)); err != nil {
		return fmt.Errorf("error requesting the userinfo: %w", err)
	}

	var mapped map[string]any

	if p.algUserInfo != "" {
		if mapped, err = p.verifySignedUserInfo(ctx, data); err != nil {
			return fmt.Errorf("error requesting the userinfo: %w", err)
		}
	} else if err = json.Unmarshal(data, &mapped); err != nil {
		return fmt.Errorf("error requesting the userinfo: %w: %w", ErrUserInfoResponseInvalid, err)
	}

	subject, _ := mapped[claimSubject].(string)

	switch {
	case subject == "":
		return fmt.Errorf("error requesting the userinfo: %w: the 'sub' claim is required but it is absent", ErrUserInfoResponseInvalid)
	case subject != claims.Subject:
		return fmt.Errorf("error requesting the userinfo: %w", ErrUserInfoSubjectMismatch)
	}

	adoptStringClaim(&claims.PreferredUsername, mapped, claimPreferredUsername)
	adoptStringClaim(&claims.Name, mapped, claimName)
	adoptStringClaim(&claims.Email, mapped, claimEmail)

	return nil
}

// verifySignedUserInfo verifies a signed UserInfo Response and returns its claims (OpenID Connect Core 1.0 Section
// 5.3.2). The signature is verified against the same key set as the ID Token, with the same key selection and
// refetching. The 'iss' and 'aud' claims are optional in a signed UserInfo Response, however when they are present
// they must identify the configured issuer and this client, as a response signed by the provider for another client
// is otherwise indistinguishable from one signed for this client.
func (p *OpenIDConnectProvider) verifySignedUserInfo(ctx context.Context, data []byte) (mapped map[string]any, err error) {
	var jws *jose.JSONWebSignature

	if jws, err = jose.ParseSignedCompact(strings.TrimSpace(string(data)), []jose.SignatureAlgorithm{jose.SignatureAlgorithm(p.algUserInfo)}); err != nil || len(jws.Signatures) != 1 {
		return nil, fmt.Errorf("%w: the signature could not be verified", ErrUserInfoResponseInvalid)
	}

	var payload []byte

	if payload, err = verifyJWSSignature(ctx, p.keys, jws, ValidateOptions{Alg: p.algUserInfo, JWKSURI: p.jwksURI}); err != nil {
		if errors.Is(err, ErrTokenNoKey) {
			return nil, fmt.Errorf("%w: no key in the json web key set matched the signature", ErrUserInfoResponseInvalid)
		}

		return nil, fmt.Errorf("%w: the signature could not be verified", ErrUserInfoResponseInvalid)
	}

	if err = json.Unmarshal(payload, &mapped); err != nil || mapped == nil {
		return nil, fmt.Errorf("%w: the payload is not a JSON object", ErrUserInfoResponseInvalid)
	}

	if value, has := mapped[claimIssuer]; has {
		if issuer, ok := value.(string); !ok || issuer != p.issuer {
			return nil, fmt.Errorf("%w: the 'iss' claim does not match the configured issuer", ErrUserInfoResponseInvalid)
		}
	}

	if value, has := mapped[claimAudience]; has {
		if audience, ok := audienceClaim(value); !ok || !containsString(audience, p.clientID) {
			return nil, fmt.Errorf("%w: the 'aud' claim does not contain the client id '%s'", ErrUserInfoResponseInvalid, p.clientID)
		}
	}

	return mapped, nil
}

func adoptStringClaim(value *string, mapped map[string]any, name string) {
	if claim, ok := mapped[name].(string); ok && claim != "" {
		*value = claim
	}
}
