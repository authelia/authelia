package authorization

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/authelia/authelia/v4/internal/utils"
)

// NewAccessControlDomain creates a new SubjectObjectMatcher that matches the domain as a basic string.
func NewAccessControlDomain(domain string) SubjectObjectMatcher {
	d := AccessControlDomain{}

	domain = strings.ToLower(domain)

	switch {
	case strings.HasPrefix(domain, "*."):
		d.Wildcard = true
		d.Name = domain[1:]
	case strings.HasPrefix(domain, "{user}"):
		d.UserWildcard = true
		d.Name = domain[7:]
	case strings.HasPrefix(domain, "{group}"):
		d.GroupWildcard = true
		d.Name = domain[8:]
	default:
		d.Name = domain
	}

	return d
}

// AccessControlDomain represents an ACL domain.
type AccessControlDomain struct {
	Name          string
	Wildcard      bool
	UserWildcard  bool
	GroupWildcard bool
}

// IsMatch returns true if the ACL domain matches the object domain.
func (acl AccessControlDomain) IsMatch(subject Subject, object Object) (match bool) {
	switch {
	case acl.Wildcard:
		return strings.HasSuffix(object.Domain, acl.Name)
	case acl.UserWildcard:
		return object.Domain == fmt.Sprintf("%s.%s", subject.Username, acl.Name)
	case acl.GroupWildcard:
		prefix, suffix := domainToPrefixSuffix(object.Domain)

		return suffix == acl.Name && utils.IsStringInSliceFold(prefix, subject.Groups)
	default:
		return object.Domain == acl.Name
	}
}

// String returns a string representation of the SubjectObjectMatcher rule.
func (acl AccessControlDomain) String() string {
	return fmt.Sprintf("domain:%s", acl.Name)
}

// NewAccessControlDomainRegex creates a new SubjectObjectMatcher that matches the domain either in a basic way or
// dynamic User/Group subexpression group way.
func NewAccessControlDomainRegex(pattern *regexp.Regexp) SubjectObjectMatcher {
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
		return AccessControlDomainRegex{Pattern: pattern, SubexpNameUser: iuser, SubexpNameGroup: igroup}
	}

	return AccessControlDomainRegexBasic{Pattern: pattern}
}

// AccessControlDomainRegexBasic represents a basic domain regex SubjectObjectMatcher.
type AccessControlDomainRegexBasic struct {
	Pattern *regexp.Regexp
}

// IsMatch returns true if the ACL regex matches the object domain.
func (acl AccessControlDomainRegexBasic) IsMatch(_ Subject, object Object) (match bool) {
	return acl.Pattern.MatchString(object.Domain)
}

// String returns a text representation of a AccessControlDomainRegexBasic.
func (acl AccessControlDomainRegexBasic) String() string {
	return fmt.Sprintf("domain_regex:%s", acl.Pattern.String())
}

// AccessControlDomainRegex represents an ACL domain regex.
type AccessControlDomainRegex struct {
	Pattern         *regexp.Regexp
	SubexpNameUser  int
	SubexpNameGroup int
}

// IsMatch returns true if the ACL regex matches the object domain.
func (acl AccessControlDomainRegex) IsMatch(subject Subject, object Object) (match bool) {
	matches := acl.Pattern.FindAllStringSubmatch(object.Domain, -1)
	if matches == nil {
		return false
	}

	if acl.SubexpNameUser != -1 && !strings.EqualFold(subject.Username, matches[0][acl.SubexpNameUser]) {
		return false
	}

	if acl.SubexpNameGroup != -1 && !utils.IsStringInSliceFold(matches[0][acl.SubexpNameGroup], subject.Groups) {
		return false
	}

	return true
}

// String returns a text representation of a AccessControlDomainRegex.
func (acl AccessControlDomainRegex) String() string {
	return fmt.Sprintf("domain_regex(subexp):%s", acl.Pattern.String())
}
