package authorization

import (
	"fmt"
	"regexp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func NewAccessControlQuery(config [][]schema.ACLQueryRule) (rules []AccessControlQuery) {
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

func (acq AccessControlQuery) IsMatch(object Object) (isMatch bool) {
	for _, rule := range acq.Rules {
		if !rule.IsMatch(object) {
			return false
		}
	}

	return true
}

func NewAccessControlQueryObjectMatcher(rule schema.ACLQueryRule) (matcher ObjectMatcher, err error) {
	switch rule.Operator {
	case operatorPresent, operatorAbsent:
		return &AccessControlQueryMatcherPresent{key: rule.Key, present: rule.Operator == operatorPresent}, nil
	case operatorEqual, operatorNotEqual:
		return &AccessControlQueryMatcherEqual{key: rule.Key, value: rule.Value, equal: rule.Operator == operatorEqual}, nil
	case operatorPattern, operatorNotPattern:
		var pattern *regexp.Regexp

		if pattern, err = regexp.Compile(rule.Value); err != nil {
			return nil, fmt.Errorf("could not parse rule regex: %w", err)
		}

		return &AccessControlQueryMatcherPattern{key: rule.Key, pattern: pattern, match: rule.Operator == operatorPattern}, nil
	default:
		return nil, fmt.Errorf("invalid operator: %s", rule.Operator)
	}
}

type AccessControlQueryMatcherEqual struct {
	key, value string
	equal      bool
}

func (acl AccessControlQueryMatcherEqual) IsMatch(object Object) (isMatch bool) {
	switch {
	case acl.equal:
		return object.URL.Query().Get(acl.key) == acl.value
	default:
		return object.URL.Query().Get(acl.key) != acl.value
	}
}

type AccessControlQueryMatcherPresent struct {
	key     string
	present bool
}

func (acl AccessControlQueryMatcherPresent) IsMatch(object Object) (isMatch bool) {
	switch {
	case acl.present:
		return object.URL.Query().Has(acl.key)
	default:
		return !object.URL.Query().Has(acl.key)
	}
}

type AccessControlQueryMatcherPattern struct {
	key     string
	pattern *regexp.Regexp
	match   bool
}

func (acl AccessControlQueryMatcherPattern) IsMatch(object Object) (isMatch bool) {
	switch {
	case acl.match:
		return acl.pattern.MatchString(object.URL.Query().Get(acl.key))
	default:
		return !acl.pattern.MatchString(object.URL.Query().Get(acl.key))
	}
}
