package authorization

import (
	"regexp"
)

// AccessControlResource represents an ACL resource.
type AccessControlResource struct {
	Pattern *regexp.Regexp
}

// IsMatch returns true if the ACL resource match the object path.
func (acr AccessControlResource) IsMatch(object Object) (match bool) {
	return acr.Pattern.MatchString(object.Path)
}
