package validator

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// IsPolicyValid check if policy is valid.
func IsPolicyValid(policy string) (isValid bool) {
	return utils.IsStringInSlice(policy, validACLRulePolicies)
}

// IsSubjectValid check if a subject is valid.
func IsSubjectValid(subject string) (isValid bool) {
	return subject == "" || strings.HasPrefix(subject, "user:") || strings.HasPrefix(subject, "group:") || strings.HasPrefix(subject, "oauth2:client:")
}

// IsNetworkGroupValid check if a network group is valid.
func IsNetworkGroupValid(config schema.AccessControl, network string) bool {
	for _, networks := range config.Networks {
		if network != networks.Name {
			continue
		} else {
			return true
		}
	}

	return false
}

// IsNetworkValid checks if a network is valid.
func IsNetworkValid(network string) (isValid bool) {
	if net.ParseIP(network) == nil {
		_, _, err := net.ParseCIDR(network)
		return err == nil
	}

	return true
}

func ruleDescriptor(position int, rule schema.AccessControlRule) string {
	if len(rule.Domains) == 0 {
		return fmt.Sprintf("#%d", position)
	}

	return fmt.Sprintf("#%d (domain '%s')", position, strings.Join(rule.Domains, ","))
}

// ValidateAccessControl validates access control configuration.
func ValidateAccessControl(config *schema.Configuration, validator *schema.StructValidator) {
	if config.AccessControl.DefaultPolicy == "" {
		config.AccessControl.DefaultPolicy = policyDeny
	}

	if !IsPolicyValid(config.AccessControl.DefaultPolicy) {
		validator.Push(fmt.Errorf(errFmtAccessControlDefaultPolicyValue, strJoinOr(validACLRulePolicies), config.AccessControl.DefaultPolicy))
	}

	for _, n := range config.AccessControl.Networks {
		for _, networks := range n.Networks {
			if !IsNetworkValid(networks) {
				validator.Push(fmt.Errorf(errFmtAccessControlNetworkGroupIPCIDRInvalid, n.Name, networks))
			}
		}
	}
}

// ValidateRules validates an ACL Rule configuration.
func ValidateRules(config *schema.Configuration, validator *schema.StructValidator) {
	if config.AccessControl.Rules == nil || len(config.AccessControl.Rules) == 0 {
		if config.AccessControl.DefaultPolicy != policyOneFactor && config.AccessControl.DefaultPolicy != policyTwoFactor {
			validator.Push(fmt.Errorf(errFmtAccessControlDefaultPolicyWithoutRules, config.AccessControl.DefaultPolicy))

			return
		}

		validator.PushWarning(fmt.Errorf(errFmtAccessControlWarnNoRulesDefaultPolicy, config.AccessControl.DefaultPolicy))

		return
	}

	for i, rule := range config.AccessControl.Rules {
		rulePosition := i + 1

		validateDomains(rulePosition, rule, validator)

		switch rule.Policy {
		case "":
			validator.Push(fmt.Errorf(errFmtAccessControlRuleNoPolicy, ruleDescriptor(rulePosition, rule)))
		default:
			if !IsPolicyValid(rule.Policy) {
				validator.Push(fmt.Errorf(errFmtAccessControlRuleInvalidPolicy, ruleDescriptor(rulePosition, rule), strJoinOr(validACLRulePolicies), rule.Policy))
			}
		}

		validateNetworks(rulePosition, rule, config.AccessControl, validator)

		validateSubjects(rulePosition, rule, validator)

		validateMethods(rulePosition, rule, validator)

		validateQuery(i, rule, config, validator)

		if rule.Policy == policyBypass {
			validateBypass(rulePosition, rule, validator)
		}
	}
}

func validateBypass(rulePosition int, rule schema.AccessControlRule, validator *schema.StructValidator) {
	if len(rule.Subjects) != 0 {
		validator.Push(fmt.Errorf(errAccessControlRuleBypassPolicyInvalidWithSubjects, ruleDescriptor(rulePosition, rule)))
	}

	for _, pattern := range rule.DomainsRegex {
		if utils.IsStringSliceContainsAny(authorization.IdentitySubexpNames, pattern.SubexpNames()) {
			validator.Push(fmt.Errorf(errAccessControlRuleBypassPolicyInvalidWithSubjectsWithGroupDomainRegex, ruleDescriptor(rulePosition, rule)))
			return
		}
	}
}

func validateDomains(rulePosition int, rule schema.AccessControlRule, validator *schema.StructValidator) {
	if len(rule.Domains)+len(rule.DomainsRegex) == 0 {
		validator.Push(fmt.Errorf(errFmtAccessControlRuleNoDomains, ruleDescriptor(rulePosition, rule)))
	}

	for i, domain := range rule.Domains {
		if len(domain) > 1 && domain[0] == '*' && domain[1] != '.' {
			validator.PushWarning(fmt.Errorf("access_control: rule #%d: domain #%d: domain '%s' is ineffective and should probably be '%s' instead", rulePosition, i+1, domain, fmt.Sprintf("*.%s", domain[1:])))
		}
	}
}

func validateNetworks(rulePosition int, rule schema.AccessControlRule, config schema.AccessControl, validator *schema.StructValidator) {
	for _, network := range rule.Networks {
		if !IsNetworkValid(network) {
			if !IsNetworkGroupValid(config, network) {
				validator.Push(fmt.Errorf(errFmtAccessControlRuleNetworksInvalid, ruleDescriptor(rulePosition, rule), network))
			}
		}
	}
}

func validateSubjects(rulePosition int, rule schema.AccessControlRule, validator *schema.StructValidator) {
	for _, subjectRule := range rule.Subjects {
		for _, subject := range subjectRule {
			if !IsSubjectValid(subject) {
				validator.Push(fmt.Errorf(errFmtAccessControlRuleSubjectInvalid, ruleDescriptor(rulePosition, rule), subject))
			}
		}
	}
}

func validateMethods(rulePosition int, rule schema.AccessControlRule, validator *schema.StructValidator) {
	invalid, duplicates := validateList(rule.Methods, validACLHTTPMethodVerbs, true)

	if len(invalid) != 0 {
		validator.Push(fmt.Errorf(errFmtAccessControlRuleInvalidEntries, ruleDescriptor(rulePosition, rule), "methods", strJoinOr(validACLHTTPMethodVerbs), strJoinAnd(invalid)))
	}

	if len(duplicates) != 0 {
		validator.Push(fmt.Errorf(errFmtAccessControlRuleInvalidDuplicates, ruleDescriptor(rulePosition, rule), "methods", strJoinAnd(duplicates)))
	}
}

//nolint:gocyclo
func validateQuery(i int, rule schema.AccessControlRule, config *schema.Configuration, validator *schema.StructValidator) {
	for j := 0; j < len(config.AccessControl.Rules[i].Query); j++ {
		for k := 0; k < len(config.AccessControl.Rules[i].Query[j]); k++ {
			if config.AccessControl.Rules[i].Query[j][k].Operator == "" {
				if config.AccessControl.Rules[i].Query[j][k].Key != "" {
					switch config.AccessControl.Rules[i].Query[j][k].Value {
					case "", nil:
						config.AccessControl.Rules[i].Query[j][k].Operator = operatorPresent
					default:
						config.AccessControl.Rules[i].Query[j][k].Operator = operatorEqual
					}
				}
			} else if !utils.IsStringInSliceFold(config.AccessControl.Rules[i].Query[j][k].Operator, validACLRuleOperators) {
				validator.Push(fmt.Errorf(errFmtAccessControlRuleQueryInvalid, ruleDescriptor(i+1, rule), strJoinOr(validACLRuleOperators), config.AccessControl.Rules[i].Query[j][k].Operator))
			}

			if config.AccessControl.Rules[i].Query[j][k].Key == "" {
				validator.Push(fmt.Errorf(errFmtAccessControlRuleQueryInvalidNoValue, ruleDescriptor(i+1, rule), "key"))
			}

			op := config.AccessControl.Rules[i].Query[j][k].Operator

			if op == "" {
				continue
			}

			switch v := config.AccessControl.Rules[i].Query[j][k].Value.(type) {
			case nil:
				if op != operatorAbsent && op != operatorPresent {
					validator.Push(fmt.Errorf(errFmtAccessControlRuleQueryInvalidNoValueOperator, ruleDescriptor(i+1, rule), "value", op))
				}
			case string:
				switch op {
				case operatorPresent, operatorAbsent:
					if v != "" {
						validator.Push(fmt.Errorf(errFmtAccessControlRuleQueryInvalidValue, ruleDescriptor(i+1, rule), "value", op))
					}
				case operatorPattern, operatorNotPattern:
					var (
						pattern *regexp.Regexp
						err     error
					)

					if pattern, err = regexp.Compile(v); err != nil {
						validator.Push(fmt.Errorf(errFmtAccessControlRuleQueryInvalidValueParse, ruleDescriptor(i+1, rule), "value", err))
					} else {
						config.AccessControl.Rules[i].Query[j][k].Value = pattern
					}
				}
			default:
				validator.Push(fmt.Errorf(errFmtAccessControlRuleQueryInvalidValueType, ruleDescriptor(i+1, rule), v))
			}
		}
	}
}
