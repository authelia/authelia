package authorization

import "strings"

func isDomainMatching(domain string, domainRules []string) bool {
	for _, domainRule := range domainRules {
		if domain == domainRule {
			return true
		} else if strings.HasPrefix(domainRule, "*.") && strings.HasSuffix(domain, domainRule[1:]) {
			return true
		}
	}

	return false
}
