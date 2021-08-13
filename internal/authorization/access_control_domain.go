package authorization

import (
	"fmt"
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
