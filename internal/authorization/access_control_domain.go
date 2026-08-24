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
	case strings.HasPrefix(domain, DomainTokenUser), strings.HasPrefix(domain, DomainTokenGroup):
		pattern, _ := DomainTokenPattern(domain)

		return NewAccessControlDomainRegex(*regexp.MustCompile(pattern))
	default:
		m.Name = domain
	}

	return false, AccessControlDomain{m}
}

// DomainTokenPattern returns the regular expression a domain criteria which uses one of the deprecated
// IdentityDomainTokens is translated into, and true if the domain uses one of those tokens.
func DomainTokenPattern(domain string) (pattern string, ok bool) {
	domain = strings.ToLower(domain)

	switch {
	case strings.HasPrefix(domain, DomainTokenUser):
		return fmt.Sprintf(patternDomainUser, regexp.QuoteMeta(domain[len(DomainTokenUser):])), true
	case strings.HasPrefix(domain, DomainTokenGroup):
		return fmt.Sprintf(patternDomainGroup, regexp.QuoteMeta(domain[len(DomainTokenGroup):])), true
	default:
		return "", false
	}
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
