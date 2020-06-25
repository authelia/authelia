package schema

import (
	"fmt"
	"net"
	"strings"
)

// ACLRule represents one ACL rule entry; "weak" coerces a single value into slice.
type ACLRule struct {
	Domains   []string   `mapstructure:"domain,weak"`
	Policy    string     `mapstructure:"policy"`
	Subjects  [][]string `mapstructure:"subject,weak"`
	Networks  []string   `mapstructure:"networks"`
	Resources []string   `mapstructure:"resources"`
}

// IsPolicyValid check if policy is valid.
func IsPolicyValid(policy string) bool {
	return policy == denyPolicy || policy == "one_factor" || policy == "two_factor" || policy == "bypass"
}

// IsSubjectValid check if a subject is valid.
func IsSubjectValid(subject string) bool {
	return subject == "" || strings.HasPrefix(subject, "user:") || strings.HasPrefix(subject, "group:")
}

// IsNetworkValid check if a network is valid.
func IsNetworkValid(network string) bool {
	_, _, err := net.ParseCIDR(network)
	return err == nil
}

// Validate validate an ACL Rule.
func (r *ACLRule) Validate(validator *StructValidator) {
	if len(r.Domains) == 0 {
		validator.Push(fmt.Errorf("Domain must be provided"))
	}

	if !IsPolicyValid(r.Policy) {
		validator.Push(fmt.Errorf("A policy must either be 'deny', 'two_factor', 'one_factor' or 'bypass'"))
	}

	for i, subjectRule := range r.Subjects {
		for j, subject := range subjectRule {
			if !IsSubjectValid(subject) {
				validator.Push(fmt.Errorf("Subject %d-%d must start with 'user:' or 'group:'", i, j))
			}
		}
	}

	for i, network := range r.Networks {
		if !IsNetworkValid(network) {
			validator.Push(fmt.Errorf("Network %d must be a valid CIDR", i))
		}
	}
}

// AccessControlConfiguration represents the configuration related to ACLs.
type AccessControlConfiguration struct {
	DefaultPolicy string    `mapstructure:"default_policy"`
	Rules         []ACLRule `mapstructure:"rules"`
}

// Validate validate the access control configuration.
func (acc *AccessControlConfiguration) Validate(validator *StructValidator) {
	if acc.DefaultPolicy == "" {
		acc.DefaultPolicy = denyPolicy
	}

	if !IsPolicyValid(acc.DefaultPolicy) {
		validator.Push(fmt.Errorf("'default_policy' must either be 'deny', 'two_factor', 'one_factor' or 'bypass'"))
	}
}
