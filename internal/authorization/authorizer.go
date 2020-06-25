package authorization

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/logging"
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

func (s Subject) String() string {
	return fmt.Sprintf("username=%s groups=%s ip=%s", s.Username, strings.Join(s.Groups, ","), s.IP.String())
}

// Object object to check access control for.
type Object struct {
	Domain string
	Path   string
}

// selectMatchingSubjectRules take a set of rules and select only the rules matching the subject constraints.
func selectMatchingSubjectRules(rules []schema.ACLRule, subject Subject) []schema.ACLRule {
	selectedRules := []schema.ACLRule{}

	for _, rule := range rules {
		switch {
		case len(rule.Subjects) > 0:
			for _, subjectRule := range rule.Subjects {
				if isSubjectMatching(subject, subjectRule) && isIPMatching(subject.IP, rule.Networks) {
					selectedRules = append(selectedRules, rule)
				}
			}
		default:
			if isIPMatching(subject.IP, rule.Networks) {
				selectedRules = append(selectedRules, rule)
			}
		}
	}

	return selectedRules
}

func selectMatchingObjectRules(rules []schema.ACLRule, object Object) []schema.ACLRule {
	selectedRules := []schema.ACLRule{}

	for _, rule := range rules {
		if isDomainMatching(object.Domain, rule.Domains) && isPathMatching(object.Path, rule.Resources) {
			selectedRules = append(selectedRules, rule)
		}
	}

	return selectedRules
}

func selectMatchingRules(rules []schema.ACLRule, subject Subject, object Object) []schema.ACLRule {
	matchingRules := selectMatchingSubjectRules(rules, subject)
	return selectMatchingObjectRules(matchingRules, object)
}

// PolicyToLevel converts a string policy to int authorization level.
func PolicyToLevel(policy string) Level {
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

// IsSecondFactorEnabled return true if at least one policy is set to second factor.
func (p *Authorizer) IsSecondFactorEnabled() bool {
	if PolicyToLevel(p.configuration.DefaultPolicy) == TwoFactor {
		return true
	}

	for _, r := range p.configuration.Rules {
		if PolicyToLevel(r.Policy) == TwoFactor {
			return true
		}
	}

	return false
}

// GetRequiredLevel retrieve the required level of authorization to access the object.
func (p *Authorizer) GetRequiredLevel(subject Subject, requestURL url.URL) Level {
	logging.Logger().Tracef("Check authorization of subject %s and url %s.",
		subject.String(), requestURL.String())

	matchingRules := selectMatchingRules(p.configuration.Rules, subject, Object{
		Domain: requestURL.Hostname(),
		Path:   requestURL.Path,
	})

	if len(matchingRules) > 0 {
		return PolicyToLevel(matchingRules[0].Policy)
	}

	logging.Logger().Tracef("No matching rule for subject %s and url %s... Applying default policy.",
		subject.String(), requestURL.String())

	return PolicyToLevel(p.configuration.DefaultPolicy)
}

// IsURLMatchingRuleWithGroupSubjects returns true if the request has at least one
// matching ACL with a subject of type group attached to it, otherwise false.
func (p *Authorizer) IsURLMatchingRuleWithGroupSubjects(requestURL url.URL) (hasGroupSubjects bool) {
	for _, rule := range p.configuration.Rules {
		if isDomainMatching(requestURL.Hostname(), rule.Domains) && isPathMatching(requestURL.Path, rule.Resources) {
			for _, subjectRule := range rule.Subjects {
				for _, subject := range subjectRule {
					if strings.HasPrefix(subject, groupPrefix) {
						return true
					}
				}
			}
		}
	}

	return false
}
