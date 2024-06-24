package authorization

import (
	"regexp"
	"strings"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewLevel converts a string policy to int authorization level.
func NewLevel(policy string) Level {
	switch policy {
	case bypass:
		return Bypass
	case oneFactor:
		return OneFactor
	case twoFactor:
		return TwoFactor
	case deny:
		return Denied
	}
	// By default the deny policy applies.
	return Denied
}

// String returns a policy string representation of an authorization.Level.
func (l Level) String() string {
	switch l {
	case Bypass:
		return bypass
	case OneFactor:
		return oneFactor
	case TwoFactor:
		return twoFactor
	case Denied:
		return deny
	default:
		return deny
	}
}

func stringSliceToRegexpSlice(strings []string) (regexps []regexp.Regexp, err error) {
	var pattern *regexp.Regexp

	for _, str := range strings {
		if pattern, err = regexp.Compile(str); err != nil {
			return nil, err
		}

		regexps = append(regexps, *pattern)
	}

	return regexps, nil
}

func schemaSubjectToACLSubject(subjectRule string) (subject SubjectMatcher) {
	if strings.HasPrefix(subjectRule, prefixUser) {
		user := strings.Trim(subjectRule[lenPrefixUser:], " ")

		return AccessControlUser{Name: user}
	}

	if strings.HasPrefix(subjectRule, prefixGroup) {
		group := strings.Trim(subjectRule[lenPrefixGroup:], " ")

		return AccessControlGroup{Name: group}
	}

	if strings.HasPrefix(subjectRule, prefixOAuth2Client) {
		clientID := strings.Trim(subjectRule[lenPrefixOAuth2Client:], " ")

		return AccessControlClient{Provider: "OAuth2", ID: clientID}
	}

	return nil
}

func ruleAddDomain(domainRules []string, rule *AccessControlRule) {
	for _, domainRule := range domainRules {
		subjects, r := NewAccessControlDomain(domainRule)

		rule.Domains = append(rule.Domains, r)

		if !rule.HasSubjects && subjects {
			rule.HasSubjects = true
		}
	}
}

func ruleAddDomainRegex(exps []regexp.Regexp, rule *AccessControlRule) {
	for _, exp := range exps {
		subjects, r := NewAccessControlDomainRegex(exp)

		rule.Domains = append(rule.Domains, r)

		if !rule.HasSubjects && subjects {
			rule.HasSubjects = true
		}
	}
}

func ruleAddResources(exps []regexp.Regexp, rule *AccessControlRule) {
	for _, exp := range exps {
		subjects, r := NewAccessControlResource(exp)

		rule.Resources = append(rule.Resources, r)

		if !rule.HasSubjects && subjects {
			rule.HasSubjects = true
		}
	}
}

func schemaMethodsToACL(methodRules []string) (methods []string) {
	for _, method := range methodRules {
		methods = append(methods, strings.ToUpper(method))
	}

	return methods
}

func schemaSubjectsToACL(subjectRules [][]string) (subjects []AccessControlSubjects) {
	for _, subjectRule := range subjectRules {
		subject := AccessControlSubjects{}

		for _, subjectRuleItem := range subjectRule {
			subject.AddSubject(subjectRuleItem)
		}

		if len(subject.Subjects) != 0 {
			subjects = append(subjects, subject)
		}
	}

	return subjects
}

func domainToPrefixSuffix(domain string) (prefix, suffix string) {
	parts := strings.Split(domain, ".")

	if len(parts) == 1 {
		return "", parts[0]
	}

	return parts[0], strings.Join(parts[1:], ".")
}

func NewSubjects(subjectRules [][]string) (subjects []AccessControlSubjects) {
	return schemaSubjectsToACL(subjectRules)
}

// IsAuthLevelSufficient returns true if the current authenticationLevel is above the authorizationLevel.
func IsAuthLevelSufficient(authenticationLevel authentication.Level, authorizationLevel Level) bool {
	switch authorizationLevel {
	case Denied:
		return false
	case OneFactor:
		return authenticationLevel >= authentication.OneFactor
	case TwoFactor:
		return authenticationLevel >= authentication.TwoFactor
	}

	return true
}

func isOpenIDConnectMFA(config *schema.Configuration) (mfa bool) {
	if config == nil || config.IdentityProviders.OIDC == nil {
		return false
	}

	for _, client := range config.IdentityProviders.OIDC.Clients {
		switch client.AuthorizationPolicy {
		case oneFactor:
			continue
		case twoFactor:
			return true
		default:
			policy, ok := config.IdentityProviders.OIDC.AuthorizationPolicies[client.AuthorizationPolicy]
			if !ok {
				continue
			}

			if policy.DefaultPolicy == twoFactor {
				return true
			}

			for _, rule := range policy.Rules {
				if rule.Policy == twoFactor {
					return true
				}
			}
		}
	}

	return false
}
