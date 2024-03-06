package authorization

import (
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
)

// Authorizer the component in charge of checking whether a user can access a given resource.
type Authorizer struct {
	defaultPolicy Level
	rules         []*AccessControlRule
	mfa           bool
	config        *schema.Configuration
	log           *logrus.Logger
}

// NewAuthorizer create an instance of authorizer with a given access control config.
func NewAuthorizer(config *schema.Configuration) (authorizer *Authorizer) {
	authorizer = &Authorizer{
		defaultPolicy: NewLevel(config.AccessControl.DefaultPolicy),
		rules:         NewAccessControlRules(config.AccessControl),
		config:        config,
		log:           logging.Logger(),
	}

	if authorizer.defaultPolicy == TwoFactor {
		authorizer.mfa = true

		return authorizer
	}

	for _, rule := range authorizer.rules {
		if rule.Policy == TwoFactor {
			authorizer.mfa = true

			return authorizer
		}
	}

	if authorizer.config.IdentityProviders.OIDC != nil {
		for _, client := range authorizer.config.IdentityProviders.OIDC.Clients {
			if client.AuthorizationPolicy == twoFactor {
				authorizer.mfa = true

				return authorizer
			}
		}
	}

	return authorizer
}

// IsSecondFactorEnabled return true if at least one policy is set to second factor.
func (p *Authorizer) IsSecondFactorEnabled() bool {
	return p.mfa
}

// GetRequiredLevel retrieve the required level of authorization to access the object.
func (p *Authorizer) GetRequiredLevel(subject Subject, object Object) (hasSubjects bool, level Level) {
	p.log.Debugf("Check authorization of subject %s and object %s (method %s).",
		subject.String(), object.String(), object.Method)

	for _, rule := range p.rules {
		if rule.IsMatch(subject, object) {
			p.log.Tracef(traceFmtACLHitMiss, "HIT", rule.Position, subject, object, object.Method, rule.Policy)

			return rule.HasSubjects, rule.Policy
		}

		p.log.Tracef(traceFmtACLHitMiss, "MISS", rule.Position, subject, object, object.Method, rule.Policy)
	}

	p.log.Debugf("No matching rule for subject %s and url %s (method %s) applying default policy", subject, object, object.Method)

	return false, p.defaultPolicy
}

// GetRuleMatchResults iterates through the rules and produces a list of RuleMatchResult provided a subject and object.
func (p *Authorizer) GetRuleMatchResults(subject Subject, object Object) (results []RuleMatchResult) {
	skipped := false

	results = make([]RuleMatchResult, len(p.rules))

	for i, rule := range p.rules {
		results[i] = RuleMatchResult{
			Rule:    rule,
			Skipped: skipped,

			MatchDomain:        rule.MatchesDomains(subject, object),
			MatchResources:     rule.MatchesResources(subject, object),
			MatchQuery:         rule.MatchesQuery(object),
			MatchMethods:       rule.MatchesMethods(object),
			MatchNetworks:      rule.MatchesNetworks(subject),
			MatchSubjects:      rule.MatchesSubjects(subject),
			MatchSubjectsExact: rule.MatchesSubjectExact(subject),
		}

		skipped = skipped || results[i].IsMatch()
	}

	return results
}
