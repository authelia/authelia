package authorization

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/authelia/authelia/v4/internal/utils"
)

// NewAccessControlDomain creates a new SubjectObjectMatcher that matches the domain as a basic string.
func NewAccessControlDomain(domain string) AccessControlDomain {
	m := &AccessControlDomainMatcher{}
	domain = strings.ToLower(domain)

	switch {
	case strings.HasPrefix(domain, "*."):
		m.Wildcard = true
		m.Name = domain[1:]
	case strings.HasPrefix(domain, "{user}"):
		m.UserWildcard = true
		m.Name = domain[7:]
	case strings.HasPrefix(domain, "{group}"):
		m.GroupWildcard = true
		m.Name = domain[8:]
	default:
		m.Name = domain
	}

	return AccessControlDomain{m}
}

// NewAccessControlDomainRegex creates a new SubjectObjectMatcher that matches the domain either in a basic way or
// dynamic User/Group subexpression group way.
func NewAccessControlDomainRegex(pattern regexp.Regexp) AccessControlDomain {
	var iuser, igroup = -1, -1

	for i, group := range pattern.SubexpNames() {
		switch group {
		case subexpNameUser:
			iuser = i
		case subexpNameGroup:
			igroup = i
		}
	}

	if iuser != -1 || igroup != -1 {
		return AccessControlDomain{RegexpGroupStringSubjectMatcher{pattern, iuser, igroup}}
	}

	return AccessControlDomain{RegexpStringSubjectMatcher{pattern}}
}

// AccessControlDomainMatcher is the basic domain matcher.
type AccessControlDomainMatcher struct {
	Name          string
	Wildcard      bool
	UserWildcard  bool
	GroupWildcard bool
}

// IsMatch returns true if this rule matches.
func (m AccessControlDomainMatcher) IsMatch(domain string, subject Subject) (match bool) {
	switch {
	case m.Wildcard:
		return strings.HasSuffix(domain, m.Name)
	case m.UserWildcard:
		return domain == fmt.Sprintf("%s.%s", subject.Username, m.Name)
	case m.GroupWildcard:
		prefix, suffix := domainToPrefixSuffix(domain)

		return suffix == m.Name && utils.IsStringInSliceFold(prefix, subject.Groups)
	default:
		return strings.EqualFold(domain, m.Name)
	}
}

// AccessControlDomain represents an ACL domain.
type AccessControlDomain struct {
	Matcher StringSubjectMatcher
}

// IsMatch returns true if the ACL domain matches the object domain.
func (acl AccessControlDomain) IsMatch(subject Subject, object Object) (match bool) {
	return acl.Matcher.IsMatch(object.Domain, subject)
}
