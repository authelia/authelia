package validator

import (
	"golang.org/x/net/publicsuffix"
)

func isCookieDomainAPublicSuffix(domain string) (valid bool) {
	var suffix string

	suffix, _ = publicsuffix.PublicSuffix(domain)

	return len(domain) == len(suffix)
}
