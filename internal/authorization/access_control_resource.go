package authorization

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/authelia/authelia/v4/internal/utils"
)

// NewAccessControlResource creates a AccessControlResource or AccessControlResourceGroup.
func NewAccessControlResource(pattern regexp.Regexp) SubjectObjectMatcher {
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
		return AccessControlResourceGroup{Pattern: pattern, SubexpNameUser: iuser, SubexpNameGroup: igroup}
	}

	return AccessControlResource{Pattern: pattern}
}

// AccessControlResource represents an ACL resource that matches without named groups.
type AccessControlResource struct {
	Pattern regexp.Regexp
}

// IsMatch returns true if the ACL resource match the object path.
func (acl AccessControlResource) IsMatch(_ Subject, object Object) (match bool) {
	return acl.Pattern.MatchString(object.Path)
}

// String returns a text representation of a AccessControlDomainRegex.
func (acl AccessControlResource) String() string {
	return fmt.Sprintf("resource:%s", acl.Pattern.String())
}

// AccessControlResourceGroup represents an ACL resource  that matches with named groups.
type AccessControlResourceGroup struct {
	Pattern         regexp.Regexp
	SubexpNameUser  int
	SubexpNameGroup int
}

// IsMatch returns true if the ACL resource match the object path.
func (acl AccessControlResourceGroup) IsMatch(subject Subject, object Object) (match bool) {
	matches := acl.Pattern.FindAllStringSubmatch(object.Path, -1)
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
func (acl AccessControlResourceGroup) String() string {
	return fmt.Sprintf("resource:%s", acl.Pattern.String())
}
