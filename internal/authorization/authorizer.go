package authorization

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/authelia/authelia/internal/authentication"
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
func (p *Authorizer) GetRequiredLevel(subject Subject, object Object) Level {
	logger := logging.Logger()
	logger.Tracef("Check authorization of subject %s and url %s.", subject.String(), object.String())

	for _, rule := range p.rules {
		if rule.IsMatch(subject, object) {
			return rule.Policy
		}
	}

	logger.Tracef("No matching rule for subject %s and url %s... Applying default policy.", subject.String(), object.String())

	return p.defaultPolicy
}

func IsAuthLevelSufficient(authenticationLevel authentication.Level, authorizationLevel Level) bool {
	if authorizationLevel == Denied {
		return false
	} else if authorizationLevel == OneFactor {
		return authenticationLevel >= authentication.OneFactor
	} else if authorizationLevel == TwoFactor {
		return authenticationLevel >= authentication.TwoFactor
	}
	return true
}
