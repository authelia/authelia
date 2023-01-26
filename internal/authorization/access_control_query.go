package authorization

import (
	"fmt"
	"regexp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewAccessControlQuery creates a new AccessControlQuery rule type.
func NewAccessControlQuery(config [][]schema.AccessControlRuleQuery) (rules []AccessControlQuery) {
	if len(config) == 0 {
		return nil
	}

	for i := 0; i < len(config); i++ {
		var rule []ObjectMatcher

		for j := 0; j < len(config[i]); j++ {
			subRule, err := NewAccessControlQueryObjectMatcher(config[i][j])
			if err != nil {
				continue
			}

			rule = append(rule, subRule)
		}

		rules = append(rules, AccessControlQuery{Rules: rule})
	}

	return rules
}

// AccessControlQuery represents an ACL query args rule.
type AccessControlQuery struct {
	Rules []ObjectMatcher
}

// IsMatch returns true if this rule matches the object.
func (acq AccessControlQuery) IsMatch(object Object) (isMatch bool) {
	for _, rule := range acq.Rules {
		if !rule.IsMatch(object) {
			return false
		}
	}

	return true
}

// NewAccessControlQueryObjectMatcher creates a new ObjectMatcher rule type from a schema.AccessControlRuleQuery.
func NewAccessControlQueryObjectMatcher(rule schema.AccessControlRuleQuery) (matcher ObjectMatcher, err error) {
	switch rule.Operator {
	case operatorPresent, operatorAbsent:
		return &AccessControlQueryMatcherPresent{key: rule.Key, present: rule.Operator == operatorPresent}, nil
	case operatorEqual, operatorNotEqual:
		if value, ok := rule.Value.(string); ok {
			return &AccessControlQueryMatcherEqual{key: rule.Key, value: value, equal: rule.Operator == operatorEqual}, nil
		} else {
			return nil, fmt.Errorf("rule value is not a string and is instead %T", rule.Value)
		}
	case operatorPattern, operatorNotPattern:
		if pattern, ok := rule.Value.(*regexp.Regexp); ok {
			return &AccessControlQueryMatcherPattern{key: rule.Key, pattern: pattern, match: rule.Operator == operatorPattern}, nil
		} else {
			return nil, fmt.Errorf("rule value is not a *regexp.Regexp and is instead %T", rule.Value)
		}
	default:
		return nil, fmt.Errorf("invalid operator: %s", rule.Operator)
	}
}

// AccessControlQueryMatcherEqual is a rule type that checks the equality of a query parameter.
type AccessControlQueryMatcherEqual struct {
	key, value string
	equal      bool
}

// IsMatch returns true if this rule matches the object.
func (acl AccessControlQueryMatcherEqual) IsMatch(object Object) (isMatch bool) {
	switch {
	case acl.equal:
		return object.URL.Query().Get(acl.key) == acl.value
	default:
		return object.URL.Query().Get(acl.key) != acl.value
	}
}

// AccessControlQueryMatcherPresent is a rule type that checks the presence of a query parameter.
type AccessControlQueryMatcherPresent struct {
	key     string
	present bool
}

// IsMatch returns true if this rule matches the object.
func (acl AccessControlQueryMatcherPresent) IsMatch(object Object) (isMatch bool) {
	switch {
	case acl.present:
		return object.URL.Query().Has(acl.key)
	default:
		return !object.URL.Query().Has(acl.key)
	}
}

// AccessControlQueryMatcherPattern is a rule type that checks a query parameter against regex.
type AccessControlQueryMatcherPattern struct {
	key     string
	pattern *regexp.Regexp
	match   bool
}

// IsMatch returns true if this rule matches the object.
func (acl AccessControlQueryMatcherPattern) IsMatch(object Object) (isMatch bool) {
	switch {
	case acl.match:
		return acl.pattern.MatchString(object.URL.Query().Get(acl.key))
	default:
		return !acl.pattern.MatchString(object.URL.Query().Get(acl.key))
	}
}
