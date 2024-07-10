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

// ValidateAccessControl validates access control configuration.
func ValidateAccessControl(config *schema.Configuration, validator *schema.StructValidator) {
	if config.AccessControl.DefaultPolicy == "" {
		config.AccessControl.DefaultPolicy = policyDeny
	}

	if !IsPolicyValid(config.AccessControl.DefaultPolicy) {
		validator.Push(fmt.Errorf(errFmtAccessControlDefaultPolicyValue, utils.StringJoinOr(validACLRulePolicies), config.AccessControl.DefaultPolicy))
	}
}

// ValidateRules validates an ACL Rule configuration.
func ValidateRules(config *schema.Configuration, validator *schema.StructValidator) {
	if len(config.AccessControl.Rules) == 0 {
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
				validator.Push(fmt.Errorf(errFmtAccessControlRuleInvalidPolicy, ruleDescriptor(rulePosition, rule), utils.StringJoinOr(validACLRulePolicies), rule.Policy))
			}
		}

		validateSubjects(rulePosition, rule, config, validator)

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

func validateSubjects(rulePosition int, rule schema.AccessControlRule, config *schema.Configuration, validator *schema.StructValidator) {
	var (
		id      string
		isValid bool
	)

	for _, subjectRule := range rule.Subjects {
		for _, subject := range subjectRule {
			if id, isValid = IsSubjectValid(subject); !isValid {
				validator.Push(fmt.Errorf(errFmtAccessControlRuleSubjectInvalid, ruleDescriptor(rulePosition, rule), subject))

				continue
			}

			if len(id) != 0 && !IsSubjectValidOAuth20(config, id) {
				validator.Push(fmt.Errorf(errFmtAccessControlRuleOAuth2ClientSubjectInvalid, ruleDescriptor(rulePosition, rule), subject, id))
			}
		}
	}
}

func validateMethods(rulePosition int, rule schema.AccessControlRule, validator *schema.StructValidator) {
	invalid, duplicates := validateList(rule.Methods, validACLHTTPMethodVerbs, true)

	if len(invalid) != 0 {
		validator.Push(fmt.Errorf(errFmtAccessControlRuleInvalidEntries, ruleDescriptor(rulePosition, rule), "methods", utils.StringJoinOr(validACLHTTPMethodVerbs), utils.StringJoinAnd(invalid)))
	}

	if len(duplicates) != 0 {
		validator.Push(fmt.Errorf(errFmtAccessControlRuleInvalidDuplicates, ruleDescriptor(rulePosition, rule), "methods", utils.StringJoinAnd(duplicates)))
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
				validator.Push(fmt.Errorf(errFmtAccessControlRuleQueryInvalid, ruleDescriptor(i+1, rule), utils.StringJoinOr(validACLRuleOperators), config.AccessControl.Rules[i].Query[j][k].Operator))
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

// IsPolicyValid check if policy is valid.
func IsPolicyValid(policy string) (isValid bool) {
	return utils.IsStringInSlice(policy, validACLRulePolicies)
}

// IsSubjectValid check if a subject is valid.
func IsSubjectValid(subject string) (id string, isValid bool) {
	if IsSubjectValidBasic(subject) {
		return "", true
	}

	if strings.HasPrefix(subject, "oauth2:client:") {
		return strings.TrimPrefix(subject, "oauth2:client:"), true
	}

	return "", false
}

func IsSubjectValidBasic(subject string) (isValid bool) {
	return strings.HasPrefix(subject, "user:") || strings.HasPrefix(subject, "group:")
}

func IsSubjectValidOAuth20(config *schema.Configuration, id string) (isValid bool) {
	if config.IdentityProviders.OIDC == nil || len(config.IdentityProviders.OIDC.Clients) == 0 {
		return false
	}

	for _, client := range config.IdentityProviders.OIDC.Clients {
		if client.ID == id {
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
