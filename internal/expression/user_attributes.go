package expression

import (
	"fmt"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types/ref"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

func NewUserAttributes(config *schema.Configuration) (ua UserAttributeResolver) {
	if len(config.Definitions.UserAttributes) == 0 {
		return &UserAttributes{}
	} else {
		return &UserAttributesExpressions{
			startup:    false,
			config:     config,
			env:        nil,
			programs:   map[string]cel.Program{},
			attributes: nil,
		}
	}
}

type UserAttributesExpressions struct {
	startup    bool
	config     *schema.Configuration
	env        *cel.Env
	programs   map[string]cel.Program
	attributes []string
}

func (e *UserAttributesExpressions) StartupCheck() (err error) {
	if e.startup {
		return nil
	}

	switch {
	case e.config == nil:
		return fmt.Errorf("error reading config: no authentication backend configured")
	case e.config.AuthenticationBackend.LDAP != nil:
		return e.ldapStartupCheck()
	case e.config.AuthenticationBackend.File != nil:
		return e.fileStartupCheck()
	default:
		return fmt.Errorf("error reading config: no authentication backend configured")
	}
}

func (e *UserAttributesExpressions) ldapStartupCheck() (err error) {
	e.attributes = []string{AttributeUserUsername, AttributeUserGroups, AttributeUserDisplayName, AttributeUserEmail, AttributeUserEmails}

	opts := []cel.EnvOption{
		newAttributeUserUsername(),
		newAttributeUserGroups(),
		newAttributeUserDisplayName(),
		newAttributeUserEmail(),
		newAttributeUserEmailVerified(),
		newAttributeUserEmails(),
		newAttributeUpdatedAt(),
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.GivenName != "" {
		e.attributes = append(e.attributes, AttributeUserGivenName)

		opts = append(opts, newAttributeUserGivenName())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.MiddleName != "" {
		e.attributes = append(e.attributes, AttributeUserMiddleName)

		opts = append(opts, newAttributeUserMiddleName())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.FamilyName != "" {
		e.attributes = append(e.attributes, AttributeUserFamilyName)

		opts = append(opts, newAttributeUserFamilyName())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.Nickname != "" {
		e.attributes = append(e.attributes, AttributeUserNickname)

		opts = append(opts, newAttributeUserNickname())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.Profile != "" {
		e.attributes = append(e.attributes, AttributeUserProfile)

		opts = append(opts, newAttributeUserProfile())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.Picture != "" {
		e.attributes = append(e.attributes, AttributeUserPicture)

		opts = append(opts, newAttributeUserPicture())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.Website != "" {
		e.attributes = append(e.attributes, AttributeUserWebsite)

		opts = append(opts, newAttributeUserWebsite())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.Gender != "" {
		e.attributes = append(e.attributes, AttributeUserGender)

		opts = append(opts, newAttributeUserGender())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.Birthdate != "" {
		e.attributes = append(e.attributes, AttributeUserBirthdate)

		opts = append(opts, newAttributeUserBirthdate())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.ZoneInfo != "" {
		e.attributes = append(e.attributes, AttributeUserZoneInfo)

		opts = append(opts, newAttributeUserZoneInfo())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.Locale != "" {
		e.attributes = append(e.attributes, AttributeUserLocale)

		opts = append(opts, newAttributeUserLocale())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.PhoneNumber != "" {
		e.attributes = append(e.attributes, AttributeUserPhoneNumber, AttributeUserPhoneNumberVerified)

		opts = append(opts, newAttributeUserPhoneNumber(), newAttributeUserPhoneNumberVerified())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.PhoneExtension != "" {
		e.attributes = append(e.attributes, AttributeUserPhoneExtension)

		opts = append(opts, newAttributeUserPhoneExtension())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.PhoneNumber != "" || e.config.AuthenticationBackend.LDAP.Attributes.PhoneExtension != "" {
		e.attributes = append(e.attributes, AttributeUserPhoneNumberRFC3966)

		opts = append(opts, newAttributeUserPhoneNumberRFC3966())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.StreetAddress != "" ||
		e.config.AuthenticationBackend.LDAP.Attributes.Locality != "" ||
		e.config.AuthenticationBackend.LDAP.Attributes.Region != "" ||
		e.config.AuthenticationBackend.LDAP.Attributes.PostalCode != "" ||
		e.config.AuthenticationBackend.LDAP.Attributes.Country != "" {
		e.attributes = append(e.attributes, AttributeUserAddress)

		opts = append(opts, newAttributeUserAddress())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.StreetAddress != "" {
		e.attributes = append(e.attributes, AttributeUserStreetAddress)

		opts = append(opts, newAttributeUserStreetAddress())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.Locality != "" {
		e.attributes = append(e.attributes, AttributeUserLocality)

		opts = append(opts, newAttributeUserLocality())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.Region != "" {
		e.attributes = append(e.attributes, AttributeUserRegion)

		opts = append(opts, newAttributeUserRegion())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.PostalCode != "" {
		e.attributes = append(e.attributes, AttributeUserPostalCode)

		opts = append(opts, newAttributeUserPostalCode())
	}

	if e.config.AuthenticationBackend.LDAP.Attributes.Country != "" {
		e.attributes = append(e.attributes, AttributeUserCountry)

		opts = append(opts, newAttributeUserCountry())
	}

	for attribute, properties := range e.config.AuthenticationBackend.LDAP.Attributes.Extra {
		optsAddExtra(opts, attribute, properties)
	}

	return e.setup(opts...)
}

func (e *UserAttributesExpressions) fileStartupCheck() (err error) {
	opts := []cel.EnvOption{
		newAttributeUserUsername(),
		newAttributeUserGroups(),
		newAttributeUserDisplayName(),
		newAttributeUserEmail(),
		newAttributeUserEmailVerified(),
		newAttributeUserEmails(),
		newAttributeUserGivenName(),
		newAttributeUserMiddleName(),
		newAttributeUserFamilyName(),
		newAttributeUserNickname(),
		newAttributeUserProfile(),
		newAttributeUserPicture(),
		newAttributeUserWebsite(),
		newAttributeUserGender(),
		newAttributeUserBirthdate(),
		newAttributeUserZoneInfo(),
		newAttributeUserLocale(),
		newAttributeUserPhoneNumber(),
		newAttributeUserPhoneNumberVerified(),
		newAttributeUserPhoneExtension(),
		newAttributeUserPhoneNumberRFC3966(),
		newAttributeUserAddress(),
		newAttributeUserStreetAddress(),
		newAttributeUserLocality(),
		newAttributeUserRegion(),
		newAttributeUserPostalCode(),
		newAttributeUserCountry(),
		newAttributeUpdatedAt(),
	}

	for attribute, properties := range e.config.AuthenticationBackend.File.ExtraAttributes {
		optsAddExtra(opts, attribute, properties)
	}

	return e.setup(opts...)
}

func (e *UserAttributesExpressions) setup(opts ...cel.EnvOption) (err error) {
	if e.env, err = cel.NewEnv(opts...); err != nil {
		return fmt.Errorf("failed to create common expression language environment: %w", err)
	}

	e.programs = map[string]cel.Program{}

	for name, properties := range e.config.Definitions.UserAttributes {
		ast, issues := e.env.Compile(properties.Expression)
		if issues != nil && issues.Err() != nil {
			return fmt.Errorf("failed to create common expression language environment: failed to parse expression '%s' with value '%s': %w", name, properties.Expression, err)
		}

		if program, err := e.env.Program(ast); err != nil {
			return fmt.Errorf("failed to create common expression language environment: failed to create expression program for '%s': %w", name, err)
		} else {
			e.programs[name] = program
		}
	}

	return nil
}

func (e *UserAttributesExpressions) Resolve(name string, detailer UserDetailer, updated time.Time) (object any, found bool) {
	activation := &UserDetailerActivation{detailer: &UserAttributeResolverDetailer{UserDetailer: detailer, updated: updated}}

	if utils.IsStringInSlice(name, e.attributes) {
		return activation.ResolveName(name)
	}

	program, ok := e.programs[name]

	if !ok {
		return nil, false
	}

	var (
		val ref.Val
		err error
	)

	val, _, err = program.Eval(activation)

	if err != nil {
		return nil, false
	}

	return val.Value(), true
}

type UserAttributes struct{}

func (e *UserAttributes) StartupCheck() (err error) {
	return nil
}

func (e *UserAttributes) Resolve(name string, detailer UserDetailer, updated time.Time) (object any, found bool) {
	activation := &UserDetailerActivation{detailer: &UserAttributeResolverDetailer{UserDetailer: detailer, updated: updated}}

	return activation.ResolveName(name)
}
