package authorization

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/authelia/authelia/v4/internal/utils"
)

// AccessControlDomain represents an ACL domain.
type AccessControlDomain struct {
	Name          string
	Wildcard      bool
	UserWildcard  bool
	GroupWildcard bool
}

// IsMatch returns true if the ACL domain matches the object domain.
func (acd AccessControlDomain) IsMatch(subject Subject, object Object) (match bool) {
	switch {
	case acd.Wildcard:
		return strings.HasSuffix(object.Domain, acd.Name)
	case acd.UserWildcard:
		return object.Domain == fmt.Sprintf("%s.%s", subject.Username, acd.Name)
	case acd.GroupWildcard:
		prefix, suffix := domainToPrefixSuffix(object.Domain)

		return suffix == acd.Name && utils.IsStringInSliceFold(prefix, subject.Groups)
	default:
		return object.Domain == acd.Name
	}
}

// AccessControlDomainRegex represents an ACL domain regex.
type AccessControlDomainRegex struct {
	Pattern *regexp.Regexp
}

// IsMatch returns true if the ACL domain matches the object domain.
func (acdr AccessControlDomainRegex) IsMatch(subject Subject, object Object) (match bool) {
	matches := acdr.Pattern.FindAllStringSubmatch(object.Domain, -1)
	if matches == nil {
		return false
	}

	subexpNames := acdr.Pattern.SubexpNames()

	if !utils.IsStringSliceContainsAny(IdentitySubexpNames, subexpNames) {
		return true
	}

	var user, group string

	for i, regexGroup := range subexpNames {
		switch regexGroup {
		case subexpNameUser:
			user = matches[0][i]
		case subexpNameGroup:
			group = matches[0][i]
		}
	}

	if user != "" && !strings.EqualFold(subject.Username, user) {
		return false
	}

	if group != "" && !utils.IsStringInSliceFold(group, subject.Groups) {
		return false
	}

	return true
}
