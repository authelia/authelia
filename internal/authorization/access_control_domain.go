package authorization

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/authelia/authelia/v4/internal/utils"
)

// NewAccessControlDomain creates a new SubjectObjectMatcher that matches the domain as a basic string.
func NewAccessControlDomain(domain string) (subjects bool, rule AccessControlDomain) {
	m := &AccessControlDomainMatcher{}
	domain = strings.ToLower(domain)

	switch {
	case strings.HasPrefix(domain, "*."):
		m.Wildcard = true
		m.Name = domain[1:]
	case strings.HasPrefix(domain, "{user}."):
		p := regexp.MustCompile(fmt.Sprintf(`(?i)^(?P<User>[a-z0-9-]+)%s$`, strings.ReplaceAll(domain[6:], `.`, `\.`)))

		return NewAccessControlDomainRegex(*p)
	case strings.HasPrefix(domain, "{group}."):
		p := regexp.MustCompile(fmt.Sprintf(`(?i)^(?P<Group>[a-z0-9-]+)%s$`, strings.ReplaceAll(domain[7:], `.`, `\.`)))

		return NewAccessControlDomainRegex(*p)
	default:
		m.Name = domain
	}

	return false, AccessControlDomain{m}
}

// NewAccessControlDomainRegex creates a new SubjectObjectMatcher that matches the domain either in a basic way or
// dynamic User/Group subexpression group way.
func NewAccessControlDomainRegex(p regexp.Regexp) (subjects bool, rule AccessControlDomain) {
	var iuser, igroup = -1, -1

	for i, group := range p.SubexpNames() {
		switch group {
		case subexpNameUser:
			iuser = i
		case subexpNameGroup:
			igroup = i
		}
	}

	if iuser != -1 || igroup != -1 {
		return true, AccessControlDomain{RegexpGroupStringSubjectMatcher{p, iuser, igroup}}
	}

	return false, AccessControlDomain{RegexpStringSubjectMatcher{p}}
}

// AccessControlDomainMatcher is the basic domain matcher.
type AccessControlDomainMatcher struct {
	Name     string
	Wildcard bool
}

// IsMatch returns true if this rule matches.
func (m AccessControlDomainMatcher) IsMatch(domain string, subject Subject) (match bool) {
	switch {
	case m.Wildcard:
		return utils.StringHasSuffixFold(domain, m.Name)
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
