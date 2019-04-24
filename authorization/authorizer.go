package authorization

import (
	"net"
	"net/url"
	"regexp"
	"strings"

	"github.com/clems4ever/authelia/configuration/schema"
)

const userPrefix = "user:"
const groupPrefix = "group:"

// Authorizer the component in charge of checking whether a user can access a given resource.
type Authorizer struct {
	configuration schema.AccessControlConfiguration
}

// NewAuthorizer create an instance of authorizer with a given access control configuration.
func NewAuthorizer(configuration schema.AccessControlConfiguration) *Authorizer {
	return &Authorizer{
		configuration: configuration,
	}
}

// Subject subject who to check access control for.
type Subject struct {
	Username string
	Groups   []string
	IP       net.IP
}

// Object object to check access control for
type Object struct {
	Domain string
	Path   string
}

func isDomainMatching(domain string, domainRule string) bool {
	if domain == domainRule { // if domain matches exactly
		return true
	} else if strings.HasPrefix(domainRule, "*") && strings.HasSuffix(domain, domainRule[1:]) {
		// If domain pattern starts with *, it's a multi domain pattern.
		return true
	}
	return false
}

func isPathMatching(path string, pathRegexps []string) bool {
	// If there is no regexp patterns, it means that we match any path.
	if len(pathRegexps) == 0 {
		return true
	}

	for _, pathRegexp := range pathRegexps {
		match, err := regexp.MatchString(pathRegexp, path)
		if err != nil {
			// TODO(c.michaud): make sure this is safe in advance to
			// avoid checking this case here.
			continue
		}

		if match {
			return true
		}
	}
	return false
}

func isSubjectMatching(subject Subject, subjectRule string) bool {
	// If no subject is provided in the rule, we match any user.
	if subjectRule == "" {
		return true
	}

	if strings.HasPrefix(subjectRule, userPrefix) {
		user := strings.Trim(subjectRule[len(userPrefix):], " ")
		if user == subject.Username {
			return true
		}
	}

	if strings.HasPrefix(subjectRule, groupPrefix) {
		group := strings.Trim(subjectRule[len(groupPrefix):], " ")
		if isStringInSlice(group, subject.Groups) {
			return true
		}
	}
	return false
}

// isIPMatching check whether user's IP is in one of the network ranges.
func isIPMatching(ip net.IP, networks []string) bool {
	// If no network is provided in the rule, we match any network
	if len(networks) == 0 {
		return true
	}

	for _, network := range networks {
		if !strings.Contains(network, "/") {
			if ip.String() == network {
				return true
			}
			continue
		}
		_, ipNet, err := net.ParseCIDR(network)
		if err != nil {
			// TODO(c.michaud): make sure the rule is valid at startup to
			// to such a case here.
			continue
		}

		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}

func isStringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// selectMatchingSubjectRules take a set of rules and select only the rules matching the subject constraints.
func selectMatchingSubjectRules(rules []schema.ACLRule, subject Subject) []schema.ACLRule {
	selectedRules := []schema.ACLRule{}

	for _, rule := range rules {
		if isSubjectMatching(subject, rule.Subject) &&
			isIPMatching(subject.IP, rule.Networks) {

			selectedRules = append(selectedRules, rule)
		}
	}

	return selectedRules
}

func selectMatchingObjectRules(rules []schema.ACLRule, object Object) []schema.ACLRule {
	selectedRules := []schema.ACLRule{}

	for _, rule := range rules {
		if isDomainMatching(object.Domain, rule.Domain) &&
			isPathMatching(object.Path, rule.Resources) {

			selectedRules = append(selectedRules, rule)
		}
	}
	return selectedRules
}

func selectMatchingRules(rules []schema.ACLRule, subject Subject, object Object) []schema.ACLRule {
	matchingRules := selectMatchingSubjectRules(rules, subject)
	return selectMatchingObjectRules(matchingRules, object)
}

func policyToLevel(policy string) Level {
	switch policy {
	case "bypass":
		return Bypass
	case "one_factor":
		return OneFactor
	case "two_factor":
		return TwoFactor
	case "deny":
		return Denied
	}
	// By default the deny policy applies.
	return Denied
}

// GetRequiredLevel retrieve the required level of authorization to access the object.
func (p *Authorizer) GetRequiredLevel(subject Subject, requestURL url.URL) Level {
	matchingRules := selectMatchingRules(p.configuration.Rules, subject, Object{
		Domain: requestURL.Hostname(),
		Path:   requestURL.Path,
	})

	if len(matchingRules) > 0 {
		return policyToLevel(matchingRules[0].Policy)
	}
	return policyToLevel(p.configuration.DefaultPolicy)
}
