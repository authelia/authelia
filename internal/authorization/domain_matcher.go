package authorization

import "strings"

func isDomainMatching(domain string, domainRule string) bool {
	if domain == domainRule { // if domain matches exactly
		return true
	} else if strings.HasPrefix(domainRule, "*.") && strings.HasSuffix(domain, domainRule[1:]) {
		// If domain pattern starts with *, it's a multi domain pattern.
		return true
	}
	return false
}
