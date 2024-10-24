package authorization

import (
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewAccessControlRules converts a schema.AccessControl into an AccessControlRule slice.
func NewAccessControlRules(config schema.AccessControl) (rules []*AccessControlRule) {
	for i, schemaRule := range config.Rules {
		rules = append(rules, NewAccessControlRule(i+1, schemaRule))
	}

	return rules
}

// NewAccessControlRule parses a schema ACL and generates an internal ACL.
func NewAccessControlRule(pos int, rule schema.AccessControlRule) *AccessControlRule {
	r := &AccessControlRule{
		Position: pos,
		Query:    NewAccessControlQuery(rule.Query),
		Methods:  schemaMethodsToACL(rule.Methods),
		Networks: AccessControlNetworks(rule.Networks),
		Subjects: schemaSubjectsToACL(rule.Subjects),
		Policy:   NewLevel(rule.Policy),
	}

	if len(r.Subjects) != 0 {
		r.HasSubjects = true
	}

	ruleAddDomain(rule.Domains, r)
	ruleAddDomainRegex(rule.DomainsRegex, r)
	ruleAddResources(rule.Resources, r)

	return r
}

// AccessControlRule controls and represents an ACL internally.
type AccessControlRule struct {
	HasSubjects bool

	Position  int
	Domains   []AccessControlDomain
	Resources []AccessControlResource
	Query     []AccessControlQuery
	Methods   []string
	Networks  AccessControlNetworks
	Subjects  []AccessControlSubjects
	Policy    Level
}

// IsMatch returns true if all elements of an AccessControlRule match the object and subject.
func (acr *AccessControlRule) IsMatch(subject Subject, object Object) (match bool) {
	if !acr.MatchesDomains(subject, object) {
		return false
	}

	if !acr.MatchesResources(subject, object) {
		return false
	}

	if !acr.MatchesQuery(object) {
		return false
	}

	if !acr.MatchesMethods(object) {
		return false
	}

	if !acr.MatchesNetworks(subject) {
		return false
	}

	if !acr.MatchesSubjects(subject) {
		return false
	}

	return true
}

// MatchesDomains returns true if the rule matches the domains.
func (acr *AccessControlRule) MatchesDomains(subject Subject, object Object) (matches bool) {
	// If there are no domains in this rule then the domain condition is a match.
	if len(acr.Domains) == 0 {
		return true
	}

	// Iterate over the domains until we find a match (return true) or until we exit the loop (return false).
	for _, domain := range acr.Domains {
		if domain.IsMatch(subject, object) {
			return true
		}
	}

	return false
}

// MatchesResources returns true if the rule matches the resources.
func (acr *AccessControlRule) MatchesResources(subject Subject, object Object) (matches bool) {
	// If there are no resources in this rule then the resource condition is a match.
	if len(acr.Resources) == 0 {
		return true
	}

	// Iterate over the resources until we find a match (return true) or until we exit the loop (return false).
	for _, resource := range acr.Resources {
		if resource.IsMatch(subject, object) {
			return true
		}
	}

	return false
}

// MatchesQuery returns true if the rule matches the query arguments.
func (acr *AccessControlRule) MatchesQuery(object Object) (match bool) {
	// If there are no query rules in this rule then the query condition is a match.
	if len(acr.Query) == 0 {
		return true
	}

	// Iterate over the queries until we find a match (return true) or until we exit the loop (return false).
	for _, query := range acr.Query {
		if query.IsMatch(object) {
			return true
		}
	}

	return false
}

// MatchesMethods returns true if the rule matches the method.
func (acr *AccessControlRule) MatchesMethods(object Object) (match bool) {
	// If there are no methods in this rule then the method condition is a match.
	if len(acr.Methods) == 0 {
		return true
	}

	return utils.IsStringInSlice(object.Method, acr.Methods)
}

// MatchesNetworks returns true if the rule matches the networks.
func (acr *AccessControlRule) MatchesNetworks(subject Subject) (match bool) {
	return acr.Networks.IsMatch(subject)
}

// MatchesSubjects returns true if the rule matches the subjects.
func (acr *AccessControlRule) MatchesSubjects(subject Subject) (match bool) {
	if subject.IsAnonymous() {
		return true
	}

	return acr.MatchesSubjectExact(subject)
}

// MatchesSubjectExact returns true if the rule matches the subjects exactly.
func (acr *AccessControlRule) MatchesSubjectExact(subject Subject) (match bool) {
	// If there are no subjects in this rule then the subject condition is a match.
	if len(acr.Subjects) == 0 {
		return true
	} else if subject.IsAnonymous() {
		return false
	}

	// Iterate over the subjects until we find a match (return true) or until we exit the loop (return false).
	for _, subjectRule := range acr.Subjects {
		if subjectRule.IsMatch(subject) {
			return true
		}
	}

	return false
}
