package oidc

import (
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func NewClientPolicy(name string, config *schema.OpenIDConnect) (policy ClientPolicy) {
	switch name {
	case authorization.OneFactor.String(), authorization.TwoFactor.String():
		return ClientPolicy{DefaultPolicy: authorization.NewLevel(name)}
	default:
		if p, ok := config.Policies[name]; ok {
			policy.DefaultPolicy = authorization.NewLevel(p.DefaultPolicy)

			for _, r := range p.Rules {
				policy.Rules = append(policy.Rules, ClientPolicyRule{
					Policy:   authorization.NewLevel(r.Policy),
					Subjects: authorization.NewSubjects(r.Subjects),
				})
			}

			return policy
		}

		return ClientPolicy{DefaultPolicy: authorization.TwoFactor}
	}
}

// ClientPolicy controls and represents a client policy.
type ClientPolicy struct {
	DefaultPolicy authorization.Level
	Rules         []ClientPolicyRule
}

func (p *ClientPolicy) GetRequiredLevel(subject authorization.Subject) authorization.Level {
	for _, rule := range p.Rules {
		if rule.IsMatch(subject) {
			return rule.Policy
		}
	}

	return p.DefaultPolicy
}

type ClientPolicyRule struct {
	Subjects []authorization.AccessControlSubjects
	Policy   authorization.Level
}

// MatchesSubjects returns true if the rule matches the subjects.
func (p *ClientPolicyRule) MatchesSubjects(subject authorization.Subject) (match bool) {
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
func (p *ClientPolicyRule) IsMatch(subject authorization.Subject) (match bool) {
	return p.MatchesSubjects(subject)
}
