package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

type tmplIssueTemplateData struct {
	Labels   []string
	Versions []string
	Proxies  []string
}

type tmplConfigurationKeysData struct {
	Timestamp time.Time
	Keys      []string
	Package   string
}

type tmplScriptsGEnData struct {
	Package          string
	VersionSwaggerUI string
}

// GitHubTagsJSON represents the JSON struct for the GitHub Tags API.
type GitHubTagsJSON struct {
	Name string `json:"name"`
}

type GitHubReleasesJSON struct {
	ID              int              `json:"id"`
	Name            string           `json:"name"`
	TagName         string           `json:"tag_name"`
	TargetCommitISH string           `json:"target_commitish"`
	NodeID          string           `json:"node_id"`
	Draft           bool             `json:"draft"`
	Prerelease      bool             `json:"prerelease"`
	URL             string           `json:"url"`
	AssetsURL       string           `json:"assets_url"`
	UploadURL       string           `json:"upload_url"`
	HTMLURL         string           `json:"html_url"`
	TarballURL      string           `json:"tarball_url"`
	ZipballURL      string           `json:"zipball_url"`
	Assets          []any            `json:"assets"`
	CreatedAt       time.Time        `json:"created_at"`
	PublishedAt     time.Time        `json:"published_at"`
	Author          GitHubAuthorJSON `json:"author"`
	Body            string           `json:"body"`
}

type GitHubAuthorJSON struct {
	ID                int    `json:"id"`
	Login             string `json:"login"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}

// DocsDataMisc represents the docs misc data schema.
type DocsDataMisc struct {
	CSP               TemplateCSP                   `json:"csp"`
	Latest            string                        `json:"latest"`
	Support           DocsDataMiscSupport           `json:"support"`
	HashingAlgorithms DocsDataMiscHashingAlgorithms `json:"hashing_algorithms"`
}

type DocsDataMiscHashingAlgorithms struct {
	PBKDF2 DocsDataMiscHashingAlgorithmsPBKDF2 `json:"pbkdf2"`
}

type DocsDataMiscHashingAlgorithmsPBKDF2 struct {
	Variants map[string]DocsDataMiscHashingAlgorithmsVariant `json:"variants"`
}

type DocsDataMiscHashingAlgorithmsVariant struct {
	FIPS              string `json:"fips"`
	DefaultIterations string `json:"default_iterations"`
}

type DocsDataMiscSupport struct {
	Traefik []string `json:"traefik"`
}

// TemplateCSP represents the CSP template vars.
type TemplateCSP struct {
	TemplateDefault     string `json:"default"`
	TemplateDevelopment string `json:"development"`
	PlaceholderNONCE    string `json:"nonce"`
}

// ConfigurationKey is the docs json model for the Authelia configuration keys.
type ConfigurationKey struct {
	Path   string `json:"path"`
	Secret bool   `json:"secret"`
	Env    string `json:"env"`
}

type Compose struct {
	Services map[string]ComposeService `json:"services"`
}

type ComposeService struct {
	Image string `json:"image"`
}

const (
	labelAreaPrefixPriority = "priority"
	labelAreaPrefixType     = "type"
	labelAreaPrefixStatus   = "status"
)

type label interface {
	String() string
	LabelDescription() string
}

type labelPriority int

const (
	labelPriorityCritical labelPriority = iota
	labelPriorityHigh
	labelPriorityMedium
	labelPriorityNormal
	labelPriorityLow
	labelPriorityVeryLow
)

var labelPriorityDescriptions = [...]string{
	"Critical",
	"High",
	"Medium",
	"Normal",
	"Low",
	"Very Low",
}

func (l labelPriority) String() string {
	return fmt.Sprintf("%s/%d/%s", labelAreaPrefixPriority, l+1, labelFormatString(labelPriorityDescriptions[l]))
}

func (l labelPriority) LabelDescription() string {
	return labelPriorityDescriptions[l]
}

type labelStatus int

const (
	labelStatusNeedsDesign labelStatus = iota
	labelStatusNeedsTriage
)

var labelStatusDescriptions = [...]string{
	"Needs Design",
	"Needs Triage",
}

func (l labelStatus) String() string {
	return fmt.Sprintf("%s/%s", labelAreaPrefixStatus, labelFormatString(labelStatusDescriptions[l]))
}

func (l labelStatus) LabelDescription() string {
	return labelStatusDescriptions[l]
}

type labelType int

const (
	labelTypeFeature labelType = iota
	labelTypeBugUnconfirmed
	labelTypeBug
)

var labelTypeDescriptions = [...]string{
	"Feature",
	"Bug: Unconfirmed",
	"Bug",
}

func (l labelType) String() string {
	return fmt.Sprintf("%s/%s", labelAreaPrefixType, labelFormatString(labelTypeDescriptions[l]))
}

func (l labelType) LabelDescription() string {
	return labelTypeDescriptions[l]
}

func labelFormatString(in string) string {
	in = strings.ReplaceAll(in, ": ", "/")
	in = strings.ReplaceAll(in, " ", "-")

	return strings.ToLower(in)
}

// CSPValue represents individual CSP values.
type CSPValue struct {
	Name  string
	Value string
}

// PackageJSON represents a NPM package.json file.
type PackageJSON struct {
	Version string `json:"version"`
}

type OpenIDConnectConformanceSuite struct {
	Name    string
	Plan    OpenIDConnectConformanceSuitePlan
	Clients []schema.IdentityProvidersOpenIDConnectClient
}

type OpenIDConnectConformanceSuitePlan struct {
	Name        string `json:"-"`
	Alias       string `json:"alias"`
	Description string `json:"description"`
	Publish     string `json:"publish"`

	Variant          *OpenIDConnectConformanceSuitePlanVariant   `json:"-"`
	Server           OpenIDConnectConformanceSuitePlanServer     `json:"server"`
	Client           *OpenIDConnectConformanceSuitePlanClient    `json:"client,omitempty"`
	ClientAlternate  *OpenIDConnectConformanceSuitePlanClient    `json:"client2,omitempty"`
	ClientSecretPost *OpenIDConnectConformanceSuitePlanClient    `json:"client_secret_post,omitempty"`
	MutualTLS        *OpenIDConnectConformanceSuitePlanMutualTLS `json:"mtls,omitempty"`
	Resource         *OpenIDConnectConformanceSuitePlanResource  `json:"resource,omitempty"`

	EKYCVerifiedClaimsRequest string `json:"ekyc_verified_claims_request,omitempty"`
	EKYCUserinfo              string `json:"ekyc_userinfo,omitempty"`
}

type OpenIDConnectConformanceSuitePlanVariant struct {
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

type OpenIDConnectConformanceSuitePlanServer struct {
	DiscoveryURL          string `json:"discoveryUrl,omitempty"`
	Issuer                string `json:"issuer,omitempty"`
	JWKSURI               string `json:"jwks_uri,omitempty"`
	AuthorizationEndpoint string `json:"authorization_endpoint,omitempty"`
	TokenEndpoint         string `json:"token_endpoint,omitempty"`
	UserinfoEndpoint      string `json:"userinfo_endpoint,omitempty"`
	ACRValues             string `json:"acr_values,omitempty"`
	LoginHint             string `json:"login_hint,omitempty"`
}

type OpenIDConnectConformanceSuitePlanClient struct {
	ID                   string `json:"client_id,omitempty"`
	Secret               string `json:"client_secret,omitempty"` //nolint:gosec
	Name                 string `json:"client_name,omitempty"`
	Scope                string `json:"scope,omitempty"`
	SecretJWTAlgorithm   string `json:"client_secret_jwt_alg,omitempty"`
	DPOPSigningAlgorithm string `json:"dpop_signing_alg,omitempty"`
	InitialAccessToken   string `json:"initial_access_token,omitempty"`
	HintType             string `json:"hint_type,omitempty"`
	HintValue            string `json:"hint_value,omitempty"`
	JWKS                 string `json:"jwks,omitempty"`
}

type OpenIDConnectConformanceSuitePlanMutualTLS struct {
	Certificate          string `json:"certificate,omitempty"`
	Key                  string `json:"key,omitempty"`
	CertificateAuthority string `json:"ca,omitempty"`
}

type OpenIDConnectConformanceSuitePlanResource struct {
	ResourceURL                 string `json:"resourceUrl"`
	ResourceURLAccountRequests  string `json:"resourceUrlAccountRequests,omitempty"`
	ResourceURLAccountsResource string `json:"resourceUrlAccountsResource,omitempty"`
	InstitutionID               string `json:"institution_id,omitempty"`
	RichAuthorizationRequest    string `json:"richAuthorizationRequest,omitempty"`
}
