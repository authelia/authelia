package authorization

import (
	"strings"
)

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
