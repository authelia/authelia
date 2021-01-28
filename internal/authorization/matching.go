package authorization

import (
	"github.com/authelia/authelia/internal/utils"
)

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
