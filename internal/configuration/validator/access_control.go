package validator

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// IsPolicyValid check if policy is valid.
func IsPolicyValid(policy string) (isValid bool) {
	return policy == policyDeny || policy == policyOneFactor || policy == policyTwoFactor || policy == policyBypass
}

// IsResourceValid check if a resource is valid.
func IsResourceValid(resource string) (err error) {
	_, err = regexp.Compile(resource)
	return err
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

func ruleDescriptor(position int, rule schema.ACLRule) string {
	if len(rule.Domains) == 0 {
		return fmt.Sprintf("#%d", position)
	}

	return fmt.Sprintf("#%d (domain '%s')", position, strings.Join(rule.Domains, ","))
}

// ValidateAccessControl validates access control configuration.
func ValidateAccessControl(configuration *schema.AccessControlConfiguration, validator *schema.StructValidator) {
	if configuration.DefaultPolicy == "" {
		configuration.DefaultPolicy = policyDeny
	}

	if !IsPolicyValid(configuration.DefaultPolicy) {
		validator.Push(fmt.Errorf(errFmtAccessControlDefaultPolicyValue, configuration.DefaultPolicy))
	}

	if configuration.Networks != nil {
		for _, n := range configuration.Networks {
			for _, networks := range n.Networks {
				if !IsNetworkValid(networks) {
					validator.Push(fmt.Errorf(errFmtAccessControlNetworkGroupIPCIDRInvalid, n.Name, networks))
				}
			}
		}
	}
}

// ValidateRules validates an ACL Rule configuration.
func ValidateRules(configuration schema.AccessControlConfiguration, validator *schema.StructValidator) {
	if configuration.Rules == nil || len(configuration.Rules) == 0 {
		if configuration.DefaultPolicy != policyOneFactor && configuration.DefaultPolicy != policyTwoFactor {
			validator.Push(fmt.Errorf(errFmtAccessControlDefaultPolicyWithoutRules, configuration.DefaultPolicy))

			return
		}

		validator.PushWarning(fmt.Errorf(errFmtAccessControlWarnNoRulesDefaultPolicy, configuration.DefaultPolicy))

		return
	}

	for i, rule := range configuration.Rules {
		rulePosition := i + 1

		if len(rule.Domains) == 0 {
			validator.Push(fmt.Errorf(errFmtAccessControlRuleNoDomains, ruleDescriptor(rulePosition, rule)))
		}

		if !IsPolicyValid(rule.Policy) {
			validator.Push(fmt.Errorf(errFmtAccessControlRuleInvalidPolicy, ruleDescriptor(rulePosition, rule), rule.Policy))
		}

		validateNetworks(rulePosition, rule, configuration, validator)

		validateResources(rulePosition, rule, validator)

		validateSubjects(rulePosition, rule, validator)

		validateMethods(rulePosition, rule, validator)

		if rule.Policy == policyBypass && len(rule.Subjects) != 0 {
			validator.Push(fmt.Errorf(errAccessControlRuleBypassPolicyInvalidWithSubjects, ruleDescriptor(rulePosition, rule)))
		}
	}
}

func validateNetworks(rulePosition int, rule schema.ACLRule, configuration schema.AccessControlConfiguration, validator *schema.StructValidator) {
	for _, network := range rule.Networks {
		if !IsNetworkValid(network) {
			if !IsNetworkGroupValid(configuration, network) {
				validator.Push(fmt.Errorf(errFmtAccessControlRuleNetworksInvalid, ruleDescriptor(rulePosition, rule), network))
			}
		}
	}
}

func validateResources(rulePosition int, rule schema.ACLRule, validator *schema.StructValidator) {
	for _, resource := range rule.Resources {
		if err := IsResourceValid(resource); err != nil {
			validator.Push(fmt.Errorf(errFmtAccessControlRuleResourceInvalid, ruleDescriptor(rulePosition, rule), resource, err))
		}
	}
}

func validateSubjects(rulePosition int, rule schema.ACLRule, validator *schema.StructValidator) {
	for _, subjectRule := range rule.Subjects {
		for _, subject := range subjectRule {
			if !IsSubjectValid(subject) {
				validator.Push(fmt.Errorf(errFmtAccessControlRuleSubjectInvalid, ruleDescriptor(rulePosition, rule), subject))
			}
		}
	}
}

func validateMethods(rulePosition int, rule schema.ACLRule, validator *schema.StructValidator) {
	for _, method := range rule.Methods {
		if !utils.IsStringInSliceFold(method, validHTTPRequestMethods) {
			validator.Push(fmt.Errorf(errFmtAccessControlRuleMethodInvalid, ruleDescriptor(rulePosition, rule), method, strings.Join(validHTTPRequestMethods, "', '")))
		}
	}
}
