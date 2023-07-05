package oidc

import (
	"time"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewClientAuthorizationPolicy creates a new ClientAuthorizationPolicy.
func NewClientAuthorizationPolicy(name string, config *schema.OpenIDConnect) (policy ClientAuthorizationPolicy) {
	switch name {
	case authorization.OneFactor.String(), authorization.TwoFactor.String():
		return ClientAuthorizationPolicy{Name: name, DefaultPolicy: authorization.NewLevel(name)}
	default:
		if p, ok := config.Policies[name]; ok {
			policy = ClientAuthorizationPolicy{
				Name:          name,
				DefaultPolicy: authorization.NewLevel(p.DefaultPolicy),
			}

			for _, r := range p.Rules {
				policy.Rules = append(policy.Rules, ClientAuthorizationPolicyRule{
					Policy:   authorization.NewLevel(r.Policy),
					Subjects: authorization.NewSubjects(r.Subjects),
				})
			}

			return policy
		}

		return ClientAuthorizationPolicy{DefaultPolicy: authorization.TwoFactor}
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
	Policy   authorization.Level
}

// MatchesSubjects returns true if the rule matches the subjects.
func (p *ClientAuthorizationPolicyRule) MatchesSubjects(subject authorization.Subject) (match bool) {
	// If there are no subjects in this rule then the subject condition is a match.
	if len(p.Subjects) == 0 {
		return true
	} else if subject.IsAnonymous() {
		return false
	}

	// Iterate over the subjects until we find a match (return true) or until we exit the loop (return false).
	for _, rule := range p.Subjects {
		if rule.IsMatch(subject) {
			return true
		}
	}

	return false
}

// IsMatch returns true if all elements of an AccessControlRule match the object and subject.
func (p *ClientAuthorizationPolicyRule) IsMatch(subject authorization.Subject) (match bool) {
	return p.MatchesSubjects(subject)
}

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
		return explicit
	case ClientConsentModeImplicit:
		return implicit
	case ClientConsentModePreConfigured:
		return preconfigured
	default:
		return ""
	}
}
