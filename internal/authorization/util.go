package authorization

import (
	"net"
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

func schemaNetworksToACL(networkRules []string, networksMap map[string][]*net.IPNet, networksCacheMap map[string]*net.IPNet) (networks []*net.IPNet) {
	for _, network := range networkRules {
		if _, ok := networksMap[network]; !ok {
			if _, ok := networksCacheMap[network]; ok {
				networks = append(networks, networksCacheMap[network])
			} else {
				cidr, err := parseNetwork(network)
				if err == nil {
					networks = append(networks, cidr)
					networksCacheMap[cidr.String()] = cidr

					if cidr.String() != network {
						networksCacheMap[network] = cidr
					}
				}
			}
		} else {
			networks = append(networks, networksMap[network]...)
		}
	}

	return networks
}

func parseSchemaNetworks(schemaNetworks []schema.AccessControlNetwork) (networksMap map[string][]*net.IPNet, networksCacheMap map[string]*net.IPNet) {
	// These maps store pointers to the net.IPNet values so we can reuse them efficiently.
	// The networksMap contains the named networks as keys, the networksCacheMap contains the CIDR notations as keys.
	networksMap = map[string][]*net.IPNet{}
	networksCacheMap = map[string]*net.IPNet{}

	for _, aclNetwork := range schemaNetworks {
		var networks []*net.IPNet

		for _, networkRule := range aclNetwork.Networks {
			cidr, err := parseNetwork(networkRule)
			if err == nil {
				networks = append(networks, cidr)
				networksCacheMap[cidr.String()] = cidr

				if cidr.String() != networkRule {
					networksCacheMap[networkRule] = cidr
				}
			}
		}

		if _, ok := networksMap[aclNetwork.Name]; len(networks) != 0 && !ok {
			networksMap[aclNetwork.Name] = networks
		}
	}

	return networksMap, networksCacheMap
}

func parseNetwork(networkRule string) (cidr *net.IPNet, err error) {
	if !strings.Contains(networkRule, "/") {
		ip := net.ParseIP(networkRule)
		if ip.To4() != nil {
			_, cidr, err = net.ParseCIDR(networkRule + "/32")
		} else {
			_, cidr, err = net.ParseCIDR(networkRule + "/128")
		}
	} else {
		_, cidr, err = net.ParseCIDR(networkRule)
	}

	return cidr, err
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
