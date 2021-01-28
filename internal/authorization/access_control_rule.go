package authorization

import (
	"net"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// NewAccessControlRules converts a schema.AccessControlConfiguration into an AccessControlRule slice.
func NewAccessControlRules(config schema.AccessControlConfiguration) (rules []*AccessControlRule) {
	networksMap, networksCacheMap := parseSchemaNetworks(config.Networks)

	for _, schemaRule := range config.Rules {
		rules = append(rules, NewAccessControlRule(schemaRule, networksMap, networksCacheMap))
	}

	return rules
}

// NewAccessControlRule parses a schema ACL and generates an internal ACL.
func NewAccessControlRule(rule schema.ACLRule, networksMap map[string][]*net.IPNet, networksCacheMap map[string]*net.IPNet) *AccessControlRule {
	return &AccessControlRule{
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
	Domains   []AccessControlDomain
	Resources []AccessControlResource
	Methods   []string
	Networks  []*net.IPNet
	Subjects  []AccessControlSubjects
	Policy    Level
}

// IsMatch returns true if all elements of an AccessControlRule match the object and subject.
func (acr *AccessControlRule) IsMatch(subject Subject, object Object) (match bool) {
	if !isMatchForDomains(object, acr) {
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

func isMatchForDomains(object Object, acl *AccessControlRule) (match bool) {
	if len(acl.Domains) == 0 {
		return true
	}

	for _, domain := range acl.Domains {
		if domain.IsMatch(object) {
			return true
		}
	}

	return false
}

func isMatchForResources(object Object, acl *AccessControlRule) (match bool) {
	if len(acl.Resources) == 0 {
		return true
	}

	for _, resource := range acl.Resources {
		if resource.IsMatch(object) {
			return true
		}
	}

	return false
}

func isMatchForMethods(object Object, acl *AccessControlRule) (match bool) {
	if len(acl.Methods) == 0 {
		return true
	}

	return utils.IsStringInSlice(object.Method, acl.Methods)
}

func isMatchForNetworks(subject Subject, acl *AccessControlRule) (match bool) {
	if len(acl.Networks) == 0 {
		return true
	}

	for _, network := range acl.Networks {
		if network.Contains(subject.IP) {
			return true
		}
	}

	return false
}

func isMatchForSubjects(subject Subject, acl *AccessControlRule) (match bool) {
	if len(acl.Subjects) == 0 {
		return true
	}

	for _, subjectRule := range acl.Subjects {
		if subjectRule.IsMatch(subject) {
			return true
		}
	}

	return false
}
