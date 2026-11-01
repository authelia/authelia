// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package conformance

import (
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Suite represents a single OpenID Connect conformance suite.
type Suite struct {
	Name    string
	Plan    Plan
	Clients []schema.IdentityProvidersOpenIDConnectClient
}

// Plan represents the test plan for a Suite.
type Plan struct {
	Name        string `json:"-"`
	Alias       string `json:"alias"`
	Description string `json:"description"`
	Publish     string `json:"publish"`

	Variant          *PlanVariant   `json:"-"`
	Server           PlanServer     `json:"server"`
	Client           *PlanClient    `json:"client,omitempty"`
	ClientAlternate  *PlanClient    `json:"client2,omitempty"`
	ClientSecretPost *PlanClient    `json:"client_secret_post,omitempty"`
	MutualTLS        *PlanMutualTLS `json:"mtls,omitempty"`
	Resource         *PlanResource  `json:"resource,omitempty"`

	EKYCVerifiedClaimsRequest string `json:"ekyc_verified_claims_request,omitempty"`
	EKYCUserinfo              string `json:"ekyc_userinfo,omitempty"`
}

// PlanVariant represents the variant of a Plan.
type PlanVariant struct {
	ServerMetadata           string `json:"server_metadata,omitempty"`
	ClientRegistration       string `json:"client_registration,omitempty"`
	ClientAuthType           string `json:"client_auth_type,omitempty"`
	ResponseType             string `json:"response_type,omitempty"`
	ResponseMode             string `json:"response_mode,omitempty"`
	CIBAMode                 string `json:"ciba_mode,omitempty"`
	FAPIProfile              string `json:"fapi_profile,omitempty"`
	FAPIResponseMode         string `json:"fapi_response_mode,omitempty"`
	FAPIAuthRequestMethod    string `json:"fapi_auth_request_method,omitempty"`
	SenderConstrain          string `json:"sender_constrain,omitempty"`
	AuthorizationRequestType string `json:"authorization_request_type,omitempty"`
	OpenID                   string `json:"openid,omitempty"`
}

// PlanServer represents the server details of a Plan.
type PlanServer struct {
	DiscoveryURL          string `json:"discoveryUrl,omitempty"`
	Issuer                string `json:"issuer,omitempty"`
	JWKSURI               string `json:"jwks_uri,omitempty"`
	AuthorizationEndpoint string `json:"authorization_endpoint,omitempty"`
	TokenEndpoint         string `json:"token_endpoint,omitempty"`
	UserinfoEndpoint      string `json:"userinfo_endpoint,omitempty"`
	ACRValues             string `json:"acr_values,omitempty"`
	LoginHint             string `json:"login_hint,omitempty"`
}

// PlanClient represents a client registered with a Plan.
type PlanClient struct {
	ID                   string `json:"client_id,omitempty"`
	Secret               string `json:"client_secret,omitempty"`
	Name                 string `json:"client_name,omitempty"`
	Scope                string `json:"scope,omitempty"`
	SecretJWTAlgorithm   string `json:"client_secret_jwt_alg,omitempty"`
	DPOPSigningAlgorithm string `json:"dpop_signing_alg,omitempty"`
	InitialAccessToken   string `json:"initial_access_token,omitempty"`
	HintType             string `json:"hint_type,omitempty"`
	HintValue            string `json:"hint_value,omitempty"`
	JWKS                 string `json:"jwks,omitempty"`
}

// PlanMutualTLS represents the Mutual TLS details of a Plan.
type PlanMutualTLS struct {
	Certificate          string `json:"certificate,omitempty"`
	Key                  string `json:"key,omitempty"`
	CertificateAuthority string `json:"ca,omitempty"`
}

// PlanResource represents the protected resource details of a Plan.
type PlanResource struct {
	ResourceURL                 string `json:"resourceUrl"`
	ResourceURLAccountRequests  string `json:"resourceUrlAccountRequests,omitempty"`
	ResourceURLAccountsResource string `json:"resourceUrlAccountsResource,omitempty"`
	InstitutionID               string `json:"institution_id,omitempty"`
	RichAuthorizationRequest    string `json:"richAuthorizationRequest,omitempty"`
}
