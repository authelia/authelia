package schema

import (
	"fmt"
	"net"
	"strings"
)

// ACLRule represent one ACL rule
type ACLRule struct {
	Domain    string   `mapstructure:"domain"`
	Policy    string   `mapstructure:"policy"`
	Subject   string   `mapstructure:"subject"`
	Networks  []string `mapstructure:"networks"`
	Resources []string `mapstructure:"resources"`
}

// IsPolicyValid check if policy is valid
func IsPolicyValid(policy string) bool {
	return policy == "deny" || policy == "one_factor" || policy == "two_factor" || policy == "bypass"
}

// IsSubjectValid check if a subject is valid
func IsSubjectValid(subject string) bool {
	return subject == "" || strings.HasPrefix(subject, "user:") || strings.HasPrefix(subject, "group:")
}

// IsNetworkValid check if a network is valid
func IsNetworkValid(network string) bool {
	_, _, err := net.ParseCIDR(network)
	return err == nil
}

// Validate validate an ACL Rule
func (r *ACLRule) Validate(validator *StructValidator) {
	if r.Domain == "" {
		validator.Push(fmt.Errorf("Domain must be provided"))
	}

	if !IsPolicyValid(r.Policy) {
		validator.Push(fmt.Errorf("A policy must either be 'deny', 'two_factor', 'one_factor' or 'bypass'"))
	}

	if !IsSubjectValid(r.Subject) {
		validator.Push(fmt.Errorf("A subject must start with 'user:' or 'group:'"))
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

// Validate validate the access control configuration
func (acc *AccessControlConfiguration) Validate(validator *StructValidator) {
	if acc.DefaultPolicy == "" {
		acc.DefaultPolicy = "deny"
	}

	if !IsPolicyValid(acc.DefaultPolicy) {
		validator.Push(fmt.Errorf("'default_policy' must either be 'deny', 'two_factor', 'one_factor' or 'bypass'"))
	}
}
