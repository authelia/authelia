package oidc

import (
	"time"

	"github.com/authelia/authelia/v4/internal/authorization"
)

// NewClientConsentPolicy converts the config options into an oidc.ClientConsentPolicy.
func NewClientConsentPolicy(mode string, duration *time.Duration) ClientConsentPolicy {
	switch mode {
	case ClientConsentModeImplicit.String():
		return ClientConsentPolicy{Mode: ClientConsentModeImplicit}
	case ClientConsentModePreConfigured.String():
		return ClientConsentPolicy{Mode: ClientConsentModePreConfigured, Duration: *duration}
	case ClientConsentModeExplicit.String():
		return ClientConsentPolicy{Mode: ClientConsentModeExplicit}
	default:
		return ClientConsentPolicy{Mode: ClientConsentModeExplicit}
	}
}

// NewClientRequestedAudienceMode converts the config option into an oidc.ClientRequestedAudienceMode.
func NewClientRequestedAudienceMode(mode string) ClientRequestedAudienceMode {
	switch mode {
	case ClientRequestedAudienceModeImplicit.String():
		return ClientRequestedAudienceModeImplicit
	case ClientRequestedAudienceModeExplicit.String():
		return ClientRequestedAudienceModeExplicit
	default:
		return ClientRequestedAudienceModeImplicit
	}
}

// ClientAuthorizationPolicy controls and represents a client policy.
type ClientAuthorizationPolicy struct {
	Name          string
	DefaultPolicy authorization.Level
	Rules         []ClientAuthorizationPolicyRule
}

// GetRequiredLevel returns the required authorization.Level given an authorization.Subject.
func (p *ClientAuthorizationPolicy) GetRequiredLevel(subject authorization.Subject) authorization.Level {
	for _, rule := range p.Rules {
		if rule.IsMatch(subject) {
			return rule.Policy
		}
	}

	return p.DefaultPolicy
}

// ClientAuthorizationPolicyRule describes the authorization.Level for particular criteria relevant to OpenID Connect 1.0 Clients.
type ClientAuthorizationPolicyRule struct {
	Subjects []authorization.AccessControlSubjects
	Networks authorization.AccessControlNetworks
	Policy   authorization.Level
}

// MatchesSubjects returns true if the rule matches the subjects.
func (p *ClientAuthorizationPolicyRule) MatchesSubjects(subject authorization.Subject) (match bool) {
	if len(p.Subjects) != 0 && subject.IsAnonymous() {
		return false
	}

	// Iterate over the subjects until we find a match (return true) or until we exit the loop (return false).
	for _, rule := range p.Subjects {
		if !rule.IsMatch(subject) {
			return false
		}
	}

	return p.Networks.IsMatch(subject)
}

// IsMatch returns true if all elements of an AccessControlRule match the object and subject.
func (p *ClientAuthorizationPolicyRule) IsMatch(subject authorization.Subject) (match bool) {
	return p.MatchesSubjects(subject)
}

// ClientConsentPolicy is the consent configuration for a client.
type ClientConsentPolicy struct {
	Mode     ClientConsentMode
	Duration time.Duration
}

// String returns the string representation of the ClientConsentMode.
func (c ClientConsentPolicy) String() string {
	return c.Mode.String()
}

// ClientConsentMode represents the consent mode for a client.
type ClientConsentMode int

const (
	// ClientConsentModeExplicit means the client does not implicitly assume consent, and does not allow pre-configured
	// consent sessions.
	ClientConsentModeExplicit ClientConsentMode = iota

	// ClientConsentModePreConfigured means the client does not implicitly assume consent, but does allow pre-configured
	// consent sessions.
	ClientConsentModePreConfigured

	// ClientConsentModeImplicit means the client does implicitly assume consent, and does not allow pre-configured
	// consent sessions.
	ClientConsentModeImplicit
)

// String returns the string representation of the ClientConsentMode.
func (c ClientConsentMode) String() string {
	switch c {
	case ClientConsentModeExplicit:
		return valueExplicit
	case ClientConsentModeImplicit:
		return valueImplicit
	case ClientConsentModePreConfigured:
		return valuePreconfigured
	default:
		return ""
	}
}

// ClientRequestedAudienceMode represents the requested audience mode for a client.
type ClientRequestedAudienceMode int

const (
	// ClientRequestedAudienceModeExplicit means the client requires that the audience is explicitly requested
	// for it to be considered requested and therefore granted.
	ClientRequestedAudienceModeExplicit ClientRequestedAudienceMode = iota

	// ClientRequestedAudienceModeImplicit means the client implicitly assumes that the requested audience is all of the
	// permitted audiences when the request parameter is absent.
	ClientRequestedAudienceModeImplicit
)

// String returns the string representation of the ClientRequestedAudienceMode.
func (ram ClientRequestedAudienceMode) String() string {
	switch ram {
	case ClientRequestedAudienceModeExplicit:
		return valueExplicit
	case ClientRequestedAudienceModeImplicit:
		return valueImplicit
	default:
		return ""
	}
}
