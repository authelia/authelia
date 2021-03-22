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
	return policy == denyPolicy || policy == oneFactorPolicy || policy == twoFactorPolicy || policy == bypassPolicy
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
	for _, r := range configuration.Rules {
		if len(r.Domains) == 0 {
			validator.Push(fmt.Errorf("No access control rules have been defined"))
		}

		if !IsPolicyValid(r.Policy) {
			validator.Push(fmt.Errorf("Policy [%s] for domain: %s is invalid, a policy must either be 'deny', 'two_factor', 'one_factor' or 'bypass'", r.Policy, r.Domains))
		}

		validateNetworks(r, configuration, validator)

		validateResources(r, validator)

		validateSubjects(r, validator)

		validateMethods(r, validator)

		if r.Policy == bypassPolicy && len(r.Subjects) != 0 {
			validator.Push(fmt.Errorf(errAccessControlInvalidPolicyWithSubjects, r.Domains, r.Subjects))
		}
	}
}

func validateNetworks(r schema.ACLRule, configuration schema.AccessControlConfiguration, validator *schema.StructValidator) {
	for _, network := range r.Networks {
		if !IsNetworkValid(network) {
			if !IsNetworkGroupValid(configuration, network) {
				validator.Push(fmt.Errorf("Network %s for domain: %s is not a valid network or network group", r.Networks, r.Domains))
			}
		}
	}
}

func validateResources(r schema.ACLRule, validator *schema.StructValidator) {
	for _, resource := range r.Resources {
		if err := IsResourceValid(resource); err != nil {
			validator.Push(fmt.Errorf("Resource %s for domain: %s is invalid, %s", r.Resources, r.Domains, err))
		}
	}
}

func validateSubjects(r schema.ACLRule, validator *schema.StructValidator) {
	for _, subjectRule := range r.Subjects {
		for _, subject := range subjectRule {
			if !IsSubjectValid(subject) {
				validator.Push(fmt.Errorf("Subject %s for domain: %s is invalid, must start with 'user:' or 'group:'", subjectRule, r.Domains))
			}
		}
	}
}

func validateMethods(r schema.ACLRule, validator *schema.StructValidator) {
	for _, method := range r.Methods {
		if !utils.IsStringInSliceFold(method, validRequestMethods) {
			validator.Push(fmt.Errorf("Method %s for domain: %s is invalid, must be one of the following methods: %s", method, r.Domains, strings.Join(validRequestMethods, ", ")))
		}
	}
}
