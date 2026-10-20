// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

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
	query         bool
	log           *logrus.Logger
}

// NewAuthorizer create an instance of authorizer with a given access control config.
func NewAuthorizer(config *schema.Configuration) (authorizer *Authorizer) {
	authorizer = &Authorizer{
		defaultPolicy: NewLevel(config.AccessControl.DefaultPolicy),
		rules:         NewAccessControlRules(config.AccessControl),
		log:           logging.Logger(),
	}

	for _, rule := range authorizer.rules {
		if len(rule.Query) != 0 {
			authorizer.query = true

			break
		}
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

	authorizer.mfa = isOpenIDConnectMFA(config)

	return authorizer
}

// IsSecondFactorEnabled return true if at least one policy is set to second factor.
func (p *Authorizer) IsSecondFactorEnabled() bool {
	return p.mfa
}

// GetRequiredLevel retrieve the required level of authorization to access the object. If any rule has query criteria
// and the query string of the object is ambiguous, the most restrictive level of every interpretation is returned.
func (p *Authorizer) GetRequiredLevel(subject Subject, object Object) (hasSubjects bool, level Level) {
	hasSubjects, level = p.getRequiredLevel(subject, object)

	if !p.query {
		return hasSubjects, level
	}

	for _, o := range NewObjectQueryInterpretations(object) {
		p.log.Debugf("Check authorization of subject %s and object %s (method %s) using an interpretation of the ambiguous query string.", subject.String(), o.String(), o.Method)

		h, l := p.getRequiredLevel(subject, o)

		switch {
		case l > level:
			hasSubjects, level = h, l
		case l == level:
			hasSubjects = hasSubjects || h
		}
	}

	return hasSubjects, level
}

func (p *Authorizer) getRequiredLevel(subject Subject, object Object) (hasSubjects bool, level Level) {
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
