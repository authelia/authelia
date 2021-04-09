package validator

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// IsPolicyValid check if policy is valid.
func IsPolicyValid(policy string) (isValid bool) {
	return policy == denyPolicy || policy == "one_factor" || policy == "two_factor" || policy == bypassPolicy
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

// ValidateAccessControl validates access control configuration.
func ValidateAccessControl(configuration schema.AccessControlConfiguration, validator *schema.StructValidator) {
	if !IsPolicyValid(configuration.DefaultPolicy) {
		validator.Push(fmt.Errorf("'default_policy' must either be 'deny', 'two_factor', 'one_factor' or 'bypass'"))
	}

	if configuration.Networks != nil {
		for _, n := range configuration.Networks {
			for _, networks := range n.Networks {
				if !IsNetworkValid(networks) {
					validator.Push(fmt.Errorf("Network %s from network group: %s must be a valid IP or CIDR", n.Networks, n.Name))
				}
			}
		}
	}
}

// ValidateRules validates an ACL Rule configuration.
func ValidateRules(configuration schema.AccessControlConfiguration, validator *schema.StructValidator) {
	if len(configuration.Rules) == 0 {
		defaultPolicy := strings.ToLower(configuration.DefaultPolicy)
		if defaultPolicy == "deny" || defaultPolicy == "bypass" {
			validator.Push(fmt.Errorf("Default policy is [%s] invalid, access control rules must be provided or a policy must either be 'one_factor' or 'two_factor'", defaultPolicy))

			return
		}

		validator.PushWarning(fmt.Errorf("No access control rules have been defined so the default policy %s will be applied to all requests", defaultPolicy))

		return
	}

	for i, rule := range configuration.Rules {
		ruleID := i + 1

		if len(rule.Domains) == 0 {
			validator.Push(fmt.Errorf("Rule #%d is invalid, a policy must have one or more domains", ruleID))
		}

		if !IsPolicyValid(rule.Policy) {
			validator.Push(fmt.Errorf("Policy [%s] for rule #%d domain: %s is invalid, a policy must either be 'deny', 'two_factor', 'one_factor' or 'bypass'", rule.Policy, ruleID, rule.Domains))
		}

		validateNetworks(ruleID, rule, configuration, validator)

		validateResources(ruleID, rule, validator)

		validateSubjects(ruleID, rule, validator)

		validateMethods(ruleID, rule, validator)

		if rule.Policy == bypassPolicy && len(rule.Subjects) != 0 {
			validator.Push(fmt.Errorf(errAccessControlInvalidPolicyWithSubjects, ruleID, rule.Domains, rule.Subjects))
		}
	}
}

func validateNetworks(id int, r schema.ACLRule, configuration schema.AccessControlConfiguration, validator *schema.StructValidator) {
	for _, network := range r.Networks {
		if !IsNetworkValid(network) {
			if !IsNetworkGroupValid(configuration, network) {
				validator.Push(fmt.Errorf("Network %s for rule #%d domain: %s is not a valid network or network group", r.Networks, id, r.Domains))
			}
		}
	}
}

func validateResources(id int, r schema.ACLRule, validator *schema.StructValidator) {
	for _, resource := range r.Resources {
		if err := IsResourceValid(resource); err != nil {
			validator.Push(fmt.Errorf("Resource %s for rule #%d domain: %s is invalid, %s", r.Resources, id, r.Domains, err))
		}
	}
}

func validateSubjects(id int, r schema.ACLRule, validator *schema.StructValidator) {
	for _, subjectRule := range r.Subjects {
		for _, subject := range subjectRule {
			if !IsSubjectValid(subject) {
				validator.Push(fmt.Errorf("Subject %s for rule #%d domain: %s is invalid, must start with 'user:' or 'group:'", subjectRule, id, r.Domains))
			}
		}
	}
}

func validateMethods(id int, r schema.ACLRule, validator *schema.StructValidator) {
	for _, method := range r.Methods {
		if !utils.IsStringInSliceFold(method, validRequestMethods) {
			validator.Push(fmt.Errorf("Method %s for rule #%d domain: %s is invalid, must be one of the following methods: %s", method, id, r.Domains, strings.Join(validRequestMethods, ", ")))
		}
	}
}
