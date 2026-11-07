// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package authorization

import (
	"sync/atomic"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
)

// authorizerState is an immutable generation of the access control configuration. A new one is built for every
// reload and swapped in atomically; it must never be modified after construction.
type authorizerState struct {
	defaultPolicy Level
	rules         []*AccessControlRule
	mfa           bool
	oidcMFA       bool
	generation    uint64
}

func newAuthorizerState(config schema.AccessControl, oidcMFA bool, generation uint64) (state *authorizerState) {
	state = &authorizerState{
		defaultPolicy: NewLevel(config.DefaultPolicy),
		rules:         NewAccessControlRules(config),
		oidcMFA:       oidcMFA,
		generation:    generation,
	}

	if state.defaultPolicy == TwoFactor {
		state.mfa = true

		return state
	}

	for _, rule := range state.rules {
		if rule.Policy == TwoFactor {
			state.mfa = true

			return state
		}
	}

	state.mfa = oidcMFA

	return state
}

// Authorizer the component in charge of checking whether a user can access a given resource.
//
// The access control rules live in an atomically swapped immutable state so that they can be hot reloaded with Update.
// An Authorizer returned from Snapshot is pinned to the state that was current at the time, which lets a request use a
// single consistent generation for its whole lifetime.
type Authorizer struct {
	state  atomic.Pointer[authorizerState]
	pinned *authorizerState
	log    *logrus.Logger
}

// NewAuthorizer create an instance of authorizer with a given access control config.
func NewAuthorizer(config *schema.Configuration) (authorizer *Authorizer) {
	authorizer = &Authorizer{
		log: logging.Logger(),
	}

	authorizer.state.Store(newAuthorizerState(config.AccessControl, isOpenIDConnectMFA(config), 1))

	return authorizer
}

// current returns the state this Authorizer should evaluate against.
func (p *Authorizer) current() *authorizerState {
	if p.pinned != nil {
		return p.pinned
	}

	return p.state.Load()
}

// Update builds a new generation from the given access control configuration and atomically switches new evaluations
// to it. Evaluations which already started, and Snapshot instances, continue to use the previous generation. It has
// no effect on a pinned Authorizer. Whether any OpenID Connect client requires two-factor is part of the static
// configuration so it is carried over from the previous generation.
//
// Update must not be called concurrently with itself; the dynamic configuration manager serializes calls.
func (p *Authorizer) Update(config schema.AccessControl) (generation uint64) {
	if p.pinned != nil {
		return p.pinned.generation
	}

	previous := p.state.Load()

	generation = previous.generation + 1

	p.state.Store(newAuthorizerState(config, previous.oidcMFA, generation))

	return generation
}

// Snapshot returns an Authorizer pinned to the generation which is current right now. It is cheap and is intended to be
// taken once per request so every authorization decision in that request is consistent.
func (p *Authorizer) Snapshot() *Authorizer {
	return &Authorizer{pinned: p.current(), log: p.log}
}

// Generation returns the generation number of the rules in use, which starts at 1 and increases on every Update.
func (p *Authorizer) Generation() uint64 {
	return p.current().generation
}

// IsSecondFactorEnabled return true if at least one policy is set to second factor.
func (p *Authorizer) IsSecondFactorEnabled() bool {
	return p.current().mfa
}

// GetRequiredLevel retrieve the required level of authorization to access the object.
func (p *Authorizer) GetRequiredLevel(subject Subject, object Object) (hasSubjects bool, level Level) {
	state := p.current()

	p.log.Debugf("Check authorization of subject %s and object %s (method %s).",
		subject.String(), object.String(), object.Method)

	for _, rule := range state.rules {
		if rule.IsMatch(subject, object) {
			p.log.Tracef(traceFmtACLHitMiss, "HIT", rule.Position, subject, object, object.Method, rule.Policy)

			return rule.HasSubjects, rule.Policy
		}

		p.log.Tracef(traceFmtACLHitMiss, "MISS", rule.Position, subject, object, object.Method, rule.Policy)
	}

	p.log.Debugf("No matching rule for subject %s and url %s (method %s) applying default policy", subject, object, object.Method)

	return false, state.defaultPolicy
}

// GetRuleMatchResults iterates through the rules and produces a list of RuleMatchResult provided a subject and object.
func (p *Authorizer) GetRuleMatchResults(subject Subject, object Object) (results []RuleMatchResult) {
	skipped := false

	rules := p.current().rules

	results = make([]RuleMatchResult, len(rules))

	for i, rule := range rules {
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
