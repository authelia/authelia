package authorization

import (
	"net"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// NewAccessControlRules converts a schema.AccessControlConfiguration into an AccessControlRule slice.
func NewAccessControlRules(config schema.AccessControlConfiguration) (rules []*AccessControlRule) {
	networksMap, networksCacheMap := parseSchemaNetworks(config.Networks)

	for i, schemaRule := range config.Rules {
		rules = append(rules, NewAccessControlRule(i+1, schemaRule, networksMap, networksCacheMap))
	}

	return rules
}

// NewAccessControlRule parses a schema ACL and generates an internal ACL.
func NewAccessControlRule(id int, rule schema.ACLRule, networksMap map[string][]*net.IPNet, networksCacheMap map[string]*net.IPNet) *AccessControlRule {
	return &AccessControlRule{
		ID:        id,
		Domains:   schemaDomainsToACL(rule.Domains),
		Resources: schemaResourcesToACL(rule.Resources),
		Methods:   schemaMethodsToACL(rule.Methods),
		Networks:  schemaNetworksToACL(rule.Networks, networksMap, networksCacheMap),
		Subjects:  schemaSubjectsToACL(rule.Subjects),
		Policy:    PolicyToLevel(rule.Policy),
	}
}

// AccessControlRule controls and represents an ACL internally.
type AccessControlRule struct {
	ID        int
	Domains   []AccessControlDomain
	Resources []AccessControlResource
	Methods   []string
	Networks  []*net.IPNet
	Subjects  []AccessControlSubjects
	Policy    Level
}

// IsMatch returns true if all elements of an AccessControlRule match the object and subject.
func (acr *AccessControlRule) IsMatch(subject Subject, object Object) (match bool) {
	if !isMatchForDomains(subject, object, acr) {
		return false
	}

	if !isMatchForResources(object, acr) {
		return false
	}

	if !isMatchForMethods(object, acr) {
		return false
	}

	if !isMatchForNetworks(subject, acr) {
		return false
	}

	if !isMatchForSubjects(subject, acr) {
		return false
	}

	return true
}

func isMatchForDomains(subject Subject, object Object, acl *AccessControlRule) (match bool) {
	// If there are no domains in this rule then the domain condition is a match.
	if len(acl.Domains) == 0 {
		return true
	}

	// Iterate over the domains until we find a match (return true) or until we exit the loop (return false).
	for _, domain := range acl.Domains {
		if domain.IsMatch(subject, object) {
			return true
		}
	}

	return false
}

func isMatchForResources(object Object, acl *AccessControlRule) (match bool) {
	// If there are no resources in this rule then the resource condition is a match.
	if len(acl.Resources) == 0 {
		return true
	}

	// Iterate over the resources until we find a match (return true) or until we exit the loop (return false).
	for _, resource := range acl.Resources {
		if resource.IsMatch(object) {
			return true
		}
	}

	return false
}

func isMatchForMethods(object Object, acl *AccessControlRule) (match bool) {
	// If there are no methods in this rule then the method condition is a match.
	if len(acl.Methods) == 0 {
		return true
	}

	return utils.IsStringInSlice(object.Method, acl.Methods)
}

func isMatchForNetworks(subject Subject, acl *AccessControlRule) (match bool) {
	// If there are no networks in this rule then the network condition is a match.
	if len(acl.Networks) == 0 {
		return true
	}

	// Iterate over the networks until we find a match (return true) or until we exit the loop (return false).
	for _, network := range acl.Networks {
		if network.Contains(subject.IP) {
			return true
		}
	}

	return false
}

func isMatchForSubjects(subject Subject, acl *AccessControlRule) (match bool) {
	// If there are no subjects in this rule then the subject condition is a match.
	if len(acl.Subjects) == 0 || subject.IsAnonymous() {
		return true
	}

	// Iterate over the subjects until we find a match (return true) or until we exit the loop (return false).
	for _, subjectRule := range acl.Subjects {
		if subjectRule.IsMatch(subject) {
			return true
		}
	}

	return false
}
