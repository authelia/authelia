package authorization

import (
	"net"
	"regexp"
	"strings"

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

// AccessControlDomain represents an ACL domain.
type AccessControlDomain struct {
	Name     string
	Wildcard bool
}

// IsMatch returns true if the ACL domain matches the object domain.
func (acd AccessControlDomain) IsMatch(object Object) (match bool) {
	if object.Domain == acd.Name {
		return true
	}

	return acd.Wildcard && strings.HasSuffix(object.Domain, acd.Name)
}

// AccessControlResource represents an ACL resource.
type AccessControlResource struct {
	Pattern *regexp.Regexp
}

// IsMatch returns true if the ACL resource match the object path.
func (acr AccessControlResource) IsMatch(object Object) (match bool) {
	return acr.Pattern.MatchString(object.Path)
}

// AccessControlSubjects represents an ACL subject.
type AccessControlSubjects struct {
	Subjects []AccessControlSubject
}

// AddSubject appends the ACL subject based on a subject rule string.
func (acs *AccessControlSubjects) AddSubject(subjectRule string) {
	subject := schemaSubjectToACLSubject(subjectRule)

	if subject != nil {
		acs.Subjects = append(acs.Subjects, subject)
	}
}

// IsMatch returns true if the ACL subjects match the subject properties.
func (acs AccessControlSubjects) IsMatch(subject Subject) (match bool) {
	for _, rule := range acs.Subjects {
		if !rule.IsMatch(subject) {
			return false
		}
	}

	return true
}

// AccessControlUser represents an ACL subject of type `user:`.
type AccessControlUser struct {
	Name string
}

// IsMatch returns true if the ACL User name matches the subject username.
func (acu AccessControlUser) IsMatch(subject Subject) (match bool) {
	return subject.Username == acu.Name
}

// AccessControlGroup represents an ACL subject of type `group:`.
type AccessControlGroup struct {
	Name string
}

// IsMatch returns true if the ACL Group name matches one of the subjects group names.
func (acg AccessControlGroup) IsMatch(subject Subject) (match bool) {
	return utils.IsStringInSlice(acg.Name, subject.Groups)
}

// AccessControlSubject abstracts an ACL subject of type `group:` or `user:`.
type AccessControlSubject interface {
	IsMatch(subject Subject) (match bool)
}
