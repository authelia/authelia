package authorization

import (
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/logging"
)

// Authorizer the component in charge of checking whether a user can access a given resource.
type Authorizer struct {
	defaultPolicy Level
	rules         []*AccessControlRule
}

// NewAuthorizer create an instance of authorizer with a given access control configuration.
func NewAuthorizer(configuration schema.AccessControlConfiguration) *Authorizer {
	if logging.Logger().IsLevelEnabled(logrus.TraceLevel) {
		return &Authorizer{
			defaultPolicy: PolicyToLevel(configuration.DefaultPolicy),
			rules:         NewAccessControlRules(configuration),
		}
	}

	return &Authorizer{
		defaultPolicy: PolicyToLevel(configuration.DefaultPolicy),
		rules:         NewAccessControlRules(configuration),
	}
}

// IsSecondFactorEnabled return true if at least one policy is set to second factor.
func (p *Authorizer) IsSecondFactorEnabled() bool {
	if p.defaultPolicy == TwoFactor {
		return true
	}

	for _, rule := range p.rules {
		if rule.Policy == TwoFactor {
			return true
		}
	}

	return false
}

// GetRequiredLevel retrieve the required level of authorization to access the object.
func (p Authorizer) GetRequiredLevel(subject Subject, object Object) Level {
	logger := logging.Logger()

	logger.Debugf("Check authorization of subject %s and object %s (method %s).",
		subject.String(), object.String(), object.Method)

	for _, rule := range p.rules {
		if rule.IsMatch(subject, object) {
			logger.Tracef(traceFmtACLHitMiss, "HIT", rule.Position, subject.String(), object.String(), object.Method)

			return rule.Policy
		}

		logger.Tracef(traceFmtACLHitMiss, "MISS", rule.Position, subject.String(), object.String(), object.Method)
	}

	logger.Debugf("No matching rule for subject %s and url %s... Applying default policy.",
		subject.String(), object.String())

	return p.defaultPolicy
}
