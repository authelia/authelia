package expression

import (
	"fmt"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types/ref"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func NewUserAttributes(config *schema.Configuration) (ua UserAttributeResolver) {
	if config == nil || len(config.Definitions.UserAttributes) == 0 {
		return &UserAttributes{}
	}

	return &UserAttributesExpressions{
		startup:  false,
		config:   config,
		env:      nil,
		programs: map[string]cel.Program{},
	}
}

type UserAttributesExpressions struct {
	startup  bool
	config   *schema.Configuration
	env      *cel.Env
	programs map[string]cel.Program
}

func (e *UserAttributesExpressions) StartupCheck() (err error) {
	if e.startup {
		return nil
	}

	switch {
	case e.config == nil:
		err = fmt.Errorf("error reading config: no authentication backend configured")
	case e.config.AuthenticationBackend.LDAP != nil:
		err = e.ldapStartupCheck()
	case e.config.AuthenticationBackend.File != nil:
		err = e.fileStartupCheck()
	default:
		err = fmt.Errorf("error reading config: no authentication backend configured")
	}

	if err != nil {
		return err
	}

	e.startup = true

	return nil
}

//nolint:gocyclo
func (e *UserAttributesExpressions) ldapStartupCheck() (err error) {
	opts := []cel.EnvOption{
		newAttributeUserUsername(),
		newAttributeUserGroups(),
		newAttributeUserDisplayName(),
		newAttributeUserEmail(),
		newAttributeUserEmailVerified(),
		newAttributeUserEmails(),
		newAttributeUserEmailsExtra(),
		newAttributeUpdatedAt(),
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.GivenName != "" {
		opts = append(opts, newAttributeUserGivenName())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.MiddleName != "" {
		opts = append(opts, newAttributeUserMiddleName())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.FamilyName != "" {
		opts = append(opts, newAttributeUserFamilyName())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.Nickname != "" {
		opts = append(opts, newAttributeUserNickname())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.Profile != "" {
		opts = append(opts, newAttributeUserProfile())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.Picture != "" {
		opts = append(opts, newAttributeUserPicture())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.Website != "" {
		opts = append(opts, newAttributeUserWebsite())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.Gender != "" {
		opts = append(opts, newAttributeUserGender())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.Birthdate != "" {
		opts = append(opts, newAttributeUserBirthdate())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.ZoneInfo != "" {
		opts = append(opts, newAttributeUserZoneInfo())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.Locale != "" {
		opts = append(opts, newAttributeUserLocale())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.PhoneNumber != "" {
		opts = append(opts, newAttributeUserPhoneNumber(), newAttributeUserPhoneNumberVerified())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.PhoneExtension != "" {
		opts = append(opts, newAttributeUserPhoneExtension())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.PhoneNumber != "" {
		opts = append(opts, newAttributeUserPhoneNumberRFC3966())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.StreetAddress != "" ||
		e.config.AuthenticationBackend.LDAP.Attributes.Locality != "" ||
		e.config.AuthenticationBackend.LDAP.Attributes.Region != "" ||
		e.config.AuthenticationBackend.LDAP.Attributes.PostalCode != "" ||
		e.config.AuthenticationBackend.LDAP.Attributes.Country != "" {
		opts = append(opts, newAttributeUserAddress())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.StreetAddress != "" {
		opts = append(opts, newAttributeUserStreetAddress())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.Locality != "" {
		opts = append(opts, newAttributeUserLocality())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.Region != "" {
		opts = append(opts, newAttributeUserRegion())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.PostalCode != "" {
		opts = append(opts, newAttributeUserPostalCode())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.Country != "" {
		opts = append(opts, newAttributeUserCountry())
	}

	for attribute, properties := range e.config.AuthenticationBackend.LDAP.Attributes.Extra {
		opts = append(opts, optExtra(attribute, properties))
	}

	return e.setup(opts...)
}

func (e *UserAttributesExpressions) fileStartupCheck() (err error) {
	opts := getStandardCELEnvOpts()

	for attribute, properties := range e.config.AuthenticationBackend.File.ExtraAttributes {
		opts = append(opts, optExtra(attribute, properties))
	}

	return e.setup(opts...)
}

func (e *UserAttributesExpressions) setup(opts ...cel.EnvOption) (err error) {
	if e.env, err = cel.NewEnv(opts...); err != nil {
		return fmt.Errorf("failed to create common expression language environment: %w", err)
	}

	e.programs = map[string]cel.Program{}

	var program cel.Program

	for name, properties := range e.config.Definitions.UserAttributes {
		ast, issues := e.env.Compile(properties.Expression)
		if issues != nil && issues.Err() != nil {
			return fmt.Errorf("failed to create common expression language environment: failed to parse expression '%s' with value '%s': %w", name, properties.Expression, issues.Err())
		}

		if program, err = e.env.Program(ast); err != nil {
			return fmt.Errorf("failed to create common expression language environment: failed to create expression program for '%s': %w", name, err)
		} else {
			e.programs[name] = program
		}
	}

	return nil
}

func (e *UserAttributesExpressions) Resolve(name string, detailer UserDetailer, updated time.Time) (object any, found bool) {
	activation := &UserDetailerActivation{detailer: &UserAttributeResolverDetailer{UserDetailer: detailer, updated: updated}}

	if program, ok := e.programs[name]; ok {
		var (
			val ref.Val
			err error
		)
		if val, _, err = program.Eval(activation); err != nil {
			return nil, false
		}

		return toNativeValue(val), true
	}

	return activation.ResolveName(name)
}

type UserAttributes struct{}

func (e *UserAttributes) StartupCheck() (err error) {
	return nil
}

func (e *UserAttributes) Resolve(name string, detailer UserDetailer, updated time.Time) (object any, found bool) {
	activation := &UserDetailerActivation{detailer: &UserAttributeResolverDetailer{UserDetailer: detailer, updated: updated}}

	return activation.ResolveName(name)
}
