package validator

import (
	"fmt"
	"net"
	"strings"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// IsPolicyValid check if policy is valid.
func IsPolicyValid(policy string) (isValid bool) {
	return policy == policyDeny || policy == policyOneFactor || policy == policyTwoFactor || policy == policyBypass
}

// IsSubjectValid check if a subject is valid.
func IsSubjectValid(subject string) (isValid bool) {
	return subject == "" || strings.HasPrefix(subject, "user:") || strings.HasPrefix(subject, "group:")
}

// IsNetworkGroupValid check if a network group is valid.
func IsNetworkGroupValid(configuration schema.AccessControlConfiguration, network string) bool {
	for _, networks := range configuration.Networks {
		if network != networks.Name {
			continue
		} else {
			return true
		}
	}

	return false
}

// IsNetworkValid check if a network is valid.
func IsNetworkValid(network string) (isValid bool) {
	if net.ParseIP(network) == nil {
		_, _, err := net.ParseCIDR(network)
		return err == nil
	}

	return true
}

// ValidateAccessControl validates access control configuration.
func ValidateAccessControl(configuration *schema.AccessControlConfiguration, validator *schema.StructValidator) {
	if configuration.DefaultPolicy == "" {
		configuration.DefaultPolicy = policyDeny
	}

	if !IsPolicyValid(configuration.DefaultPolicy) {
		validator.Push(fmt.Errorf(errAccessControlInvalidDefaultPolicy))
	}

	if configuration.Networks != nil {
		for _, n := range configuration.Networks {
			for _, networks := range n.Networks {
				if !IsNetworkValid(networks) {
					validator.Push(fmt.Errorf(errAccessControlInvalidNetwork, n.Networks, n.Name))
				}
			}
		}
	}
}

// ValidateRules validates an ACL Rule configuration.
func ValidateRules(configuration schema.AccessControlConfiguration, validator *schema.StructValidator) {
	if configuration.Rules == nil || len(configuration.Rules) == 0 {
		if configuration.DefaultPolicy != policyOneFactor && configuration.DefaultPolicy != policyTwoFactor {
			validator.Push(fmt.Errorf(errAccessControlInvalidDefaultPolicyNoRules, configuration.DefaultPolicy))

			return
		}

		validator.PushWarning(fmt.Errorf(warnAccessControlNoRules, configuration.DefaultPolicy))

		return
	}

	for i, rule := range configuration.Rules {
		rulePosition := i + 1

		if len(rule.Domains)+len(rule.DomainsRegex) == 0 {
			validator.Push(fmt.Errorf(errAccessControlRuleNoDomains, rulePosition))
		}

		if !IsPolicyValid(rule.Policy) {
			validator.Push(fmt.Errorf(errAccessControlRuleInvalidPolicy, rulePosition, rule.Policy))
		}

		validateNetworks(rulePosition, rule, configuration, validator)

		validateSubjects(rulePosition, rule, validator)

		validateMethods(rulePosition, rule, validator)

		if rule.Policy == policyBypass {
			validateBypass(rulePosition, rule, validator)
		}
	}
}

func validateBypass(rulePosition int, rule schema.ACLRule, validator *schema.StructValidator) {
	if len(rule.Subjects) != 0 {
		validator.Push(fmt.Errorf(errAccessControlRuleInvalidPolicyWithSubjects, rulePosition))
	}

	for _, pattern := range rule.DomainsRegex {
		if utils.IsStringSliceContainsAny(authorization.IdentitySubexpNames, pattern.SubexpNames()) {
			validator.Push(fmt.Errorf(errAccessControlRuleInvalidPolicyWithSpecialDomainRegexp, rulePosition))
			return
		}
	}
}

func validateNetworks(rulePosition int, rule schema.ACLRule, configuration schema.AccessControlConfiguration, validator *schema.StructValidator) {
	for _, network := range rule.Networks {
		if !IsNetworkValid(network) {
			if !IsNetworkGroupValid(configuration, network) {
				validator.Push(fmt.Errorf(errAccessControlRuleNetworkInvalid, rulePosition, rule.Networks))
			}
		}
	}
}

func validateSubjects(rulePosition int, rule schema.ACLRule, validator *schema.StructValidator) {
	for _, subjectRule := range rule.Subjects {
		for _, subject := range subjectRule {
			if !IsSubjectValid(subject) {
				validator.Push(fmt.Errorf(errAccessControlRuleInvalidSubjectPrefix, rulePosition, subjectRule))
			}
		}
	}
}

func validateMethods(rulePosition int, rule schema.ACLRule, validator *schema.StructValidator) {
	for _, method := range rule.Methods {
		if !utils.IsStringInSliceFold(method, validHTTPRequestMethods) {
			validator.Push(fmt.Errorf(errAccessControlRuleInvalidMethod, rulePosition, method, strings.Join(validHTTPRequestMethods, ", ")))
		}
	}
}
