package authorization

import (
	"regexp"
)

// NewAccessControlResource creates a AccessControlResource or AccessControlResourceGroup.
func NewAccessControlResource(pattern regexp.Regexp) (subjects bool, rule AccessControlResource) {
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
		return true, AccessControlResource{RegexpGroupStringSubjectMatcher{pattern, iuser, igroup}}
	}

	return false, AccessControlResource{RegexpStringSubjectMatcher{pattern}}
}

// AccessControlResource represents an ACL resource that matches without named groups.
type AccessControlResource struct {
	Matcher StringSubjectMatcher
}

// IsMatch returns true if the ACL resource match the object path.
func (acl AccessControlResource) IsMatch(subject Subject, object Object) (match bool) {
	return acl.Matcher.IsMatch(object.Path, subject)
}
