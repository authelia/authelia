// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateAccessControl validates access control configuration.
func ValidateAccessControl(config *schema.Configuration, validator *schema.StructValidator) {
	validateAccessControlDefaultPolicy(&config.AccessControl, validator)
}

// ValidateRules validates an ACL Rule configuration.
func ValidateRules(config *schema.Configuration, validator *schema.StructValidator) {
	validateAccessControlRules(&config.AccessControl, func(id string) bool { return IsSubjectValidOAuth20(config, id) }, validator)
}

func validateAccessControlDefaultPolicy(config *schema.AccessControl, validator *schema.StructValidator) {
	if config.DefaultPolicy == "" {
		config.DefaultPolicy = policyDeny
	}

	if !IsPolicyValid(config.DefaultPolicy) {
		validator.Push(fmt.Errorf(errFmtAccessControlDefaultPolicyValue, utils.StringJoinOr(validACLRulePolicies), config.DefaultPolicy))
	}
}

// validateAccessControlRules validates the ACL rules. The isOAuth2Client func reports whether an OAuth 2.0 client ID
// referenced by a rule subject exists.
func validateAccessControlRules(config *schema.AccessControl, isOAuth2Client func(id string) bool, validator *schema.StructValidator) {
	if len(config.Rules) == 0 {
		if config.DefaultPolicy != policyOneFactor && config.DefaultPolicy != policyTwoFactor {
			validator.Push(fmt.Errorf(errFmtAccessControlDefaultPolicyWithoutRules, config.DefaultPolicy))

			return
		}

		validator.PushWarning(fmt.Errorf(errFmtAccessControlWarnNoRulesDefaultPolicy, config.DefaultPolicy))

		return
	}

	for i, rule := range config.Rules {
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

		validateSubjects(rulePosition, rule, isOAuth2Client, validator)

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

	for _, domain := range rule.Domains {
		domain = strings.ToLower(domain)

		for _, token := range authorization.IdentityDomainTokens {
			if strings.HasPrefix(domain, token) {
				validator.Push(fmt.Errorf(errAccessControlRuleBypassPolicyInvalidWithSubjectsWithGroupDomain, ruleDescriptor(rulePosition, rule)))

				return
			}
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

		if pattern, ok := authorization.DomainTokenPattern(domain); ok {
			validator.PushWarning(fmt.Errorf(errFmtAccessControlRuleDomainDeprecatedToken, rulePosition, i+1, domain, pattern))
		}
	}
}

func validateSubjects(rulePosition int, rule schema.AccessControlRule, isOAuth2Client func(id string) bool, validator *schema.StructValidator) {
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

			if len(id) != 0 && !isOAuth2Client(id) {
				validator.Push(fmt.Errorf(errFmtAccessControlRuleOAuth2ClientSubjectInvalid, ruleDescriptor(rulePosition, rule), subject, id))
			}
		}
	}
}

func validateMethods(rulePosition int, rule schema.AccessControlRule, validator *schema.StructValidator) {
	invalid, duplicates := validateList(rule.Methods, schema.HTTPMethods, true)

	if len(invalid) != 0 {
		validator.Push(fmt.Errorf(errFmtAccessControlRuleInvalidEntries, ruleDescriptor(rulePosition, rule), "methods", utils.StringJoinOr(schema.HTTPMethods), utils.StringJoinAnd(invalid)))
	}

	if len(duplicates) != 0 {
		validator.Push(fmt.Errorf(errFmtAccessControlRuleInvalidDuplicates, ruleDescriptor(rulePosition, rule), "methods", utils.StringJoinAnd(duplicates)))
	}
}

//nolint:gocyclo
func validateQuery(i int, rule schema.AccessControlRule, config *schema.AccessControl, validator *schema.StructValidator) {
	for j := 0; j < len(config.Rules[i].Query); j++ {
		for k := 0; k < len(config.Rules[i].Query[j]); k++ {
			if config.Rules[i].Query[j][k].Operator == "" {
				if config.Rules[i].Query[j][k].Key != "" {
					switch config.Rules[i].Query[j][k].Value {
					case "", nil:
						config.Rules[i].Query[j][k].Operator = operatorPresent
					default:
						config.Rules[i].Query[j][k].Operator = operatorEqual
					}
				}
			} else if !utils.IsStringInSliceFold(config.Rules[i].Query[j][k].Operator, validACLRuleOperators) {
				validator.Push(fmt.Errorf(errFmtAccessControlRuleQueryInvalid, ruleDescriptor(i+1, rule), utils.StringJoinOr(validACLRuleOperators), config.Rules[i].Query[j][k].Operator))
			}

			if config.Rules[i].Query[j][k].Key == "" {
				validator.Push(fmt.Errorf(errFmtAccessControlRuleQueryInvalidNoValue, ruleDescriptor(i+1, rule), "key"))
			}

			op := config.Rules[i].Query[j][k].Operator

			if op == "" {
				continue
			}

			switch v := config.Rules[i].Query[j][k].Value.(type) {
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
						config.Rules[i].Query[j][k].Value = pattern
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

// IsSubjectValid validates if a subject has a valid prefix and returns the client id if applicable.
func IsSubjectValid(subject string) (id string, isValid bool) {
	switch {
	case IsSubjectValidBasic(subject):
		return "", true
	case strings.HasPrefix(subject, "oauth2:client:"):
		return strings.TrimPrefix(subject, "oauth2:client:"), true
	default:
		return "", false
	}
}

// IsSubjectValidBasic returns true if the given subject is a valid user or group subject.
func IsSubjectValidBasic(subject string) (isValid bool) {
	return strings.HasPrefix(subject, "user:") || strings.HasPrefix(subject, "group:")
}

// IsSubjectValidOAuth20 returns true if the given client id is a registered OAuth 2.0 client.
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

func ruleDescriptor(position int, rule schema.AccessControlRule) string {
	if len(rule.Domains) == 0 {
		return fmt.Sprintf("#%d", position)
	}

	return fmt.Sprintf("#%d (domain '%s')", position, strings.Join(rule.Domains, ","))
}
