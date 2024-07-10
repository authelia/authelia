package expression

import (
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

type UserAttributes struct {
	env        *cel.Env
	programs   map[string]cel.Program
	attributes []string
}

func (e *UserAttributes) Resolve(name string, detailer UserDetailer) (object any, found bool) {
	activation := &UserDetailerActivation{detailer: detailer}

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

func NewUserAttributesLDAP(config map[string]string, attrs schema.AuthenticationBackendLDAPAttributes) (ua *UserAttributes, err error) {
	attributes := []string{attributeKeyUserUsername, attributeKeyUserGroups, attributeKeyUserDisplayName, attributeKeyUserEmail, attributeKeyUserEmails}

	opts := []cel.EnvOption{
		newAttributeUserUsername(),
		newAttributeUserGroups(),
		newAttributeUserDisplayName(),
		newAttributeUserEmail(),
		newAttributeUserEmails(),
	}

	if attrs.GivenName != "" {
		attributes = append(attributes, attributeKeyUserGivenName)

		opts = append(opts, newAttributeUserGivenName())
	}

	if attrs.MiddleName != "" {
		attributes = append(attributes, attributeKeyUserMiddleName)

		opts = append(opts, newAttributeUserMiddleName())
	}

	if attrs.FamilyName != "" {
		attributes = append(attributes, attributeKeyUserFamilyName)

		opts = append(opts, newAttributeUserFamilyName())
	}

	if attrs.Nickname != "" {
		attributes = append(attributes, attributeKeyUserNickname)

		opts = append(opts, newAttributeUserNickname())
	}

	if attrs.Profile != "" {
		attributes = append(attributes, attributeKeyUserProfile)

		opts = append(opts, newAttributeUserProfile())
	}

	if attrs.Picture != "" {
		attributes = append(attributes, attributeKeyUserPicture)

		opts = append(opts, newAttributeUserPicture())
	}

	if attrs.Website != "" {
		attributes = append(attributes, attributeKeyUserWebsite)

		opts = append(opts, newAttributeUserWebsite())
	}

	if attrs.Gender != "" {
		attributes = append(attributes, attributeKeyUserGender)

		opts = append(opts, newAttributeUserGender())
	}

	if attrs.Birthdate != "" {
		attributes = append(attributes, attributeKeyUserBirthdate)

		opts = append(opts, newAttributeUserBirthdate())
	}

	if attrs.ZoneInfo != "" {
		attributes = append(attributes, attributeKeyUserZoneInfo)

		opts = append(opts, newAttributeUserZoneInfo())
	}

	if attrs.Locale != "" {
		attributes = append(attributes, attributeKeyUserLocale)

		opts = append(opts, newAttributeUserLocale())
	}

	if attrs.PhoneNumber != "" {
		attributes = append(attributes, attributeKeyUserPhoneNumber)

		opts = append(opts, newAttributeUserPhoneNumber())
	}

	if attrs.PhoneExtension != "" {
		attributes = append(attributes, attributeKeyUserPhoneExtension)

		opts = append(opts, newAttributeUserPhoneExtension())
	}

	if attrs.StreetAddress != "" {
		attributes = append(attributes, attributeKeyUserStreetAddress)

		opts = append(opts, newAttributeUserStreetAddress())
	}

	if attrs.Locality != "" {
		attributes = append(attributes, attributeKeyUserLocality)

		opts = append(opts, newAttributeUserLocality())
	}

	if attrs.Region != "" {
		attributes = append(attributes, attributeKeyUserRegion)

		opts = append(opts, newAttributeUserRegion())
	}

	if attrs.PostalCode != "" {
		attributes = append(attributes, attributeKeyUserPostalCode)

		opts = append(opts, newAttributeUserPostalCode())
	}

	if attrs.Country != "" {
		attributes = append(attributes, attributeKeyUserCountry)

		opts = append(opts, newAttributeUserCountry())
	}

	for attribute, properties := range attrs.Extra {
		var t *types.Type

		switch properties.ValueType {
		case "string":
			t = cel.StringType
		case "integer":
			t = cel.IntType
		case "boolean":
			t = cel.BoolType
		}

		if properties.MultiValued {
			opts = append(opts, cel.Variable(attribute, cel.ListType(t)))
		} else {
			opts = append(opts, cel.Variable(attribute, t))
		}
	}

	var env *cel.Env

	if env, err = cel.NewEnv(opts...); err != nil {
		return nil, fmt.Errorf("failed to create common expression language environment: %w", err)
	}

	ua = &UserAttributes{
		env:      env,
		programs: map[string]cel.Program{},
	}

	for name, expression := range config {
		ast, issues := ua.env.Compile(expression)
		if issues != nil && issues.Err() != nil {
			return nil, fmt.Errorf("failed to create common expression language environment: failed to parse expression '%s' with value '%s': %w", name, expression, err)
		}

		if program, err := ua.env.Program(ast); err != nil {
			return nil, fmt.Errorf("failed to create common expression language environment: failed to create expression program for '%s': %w", name, err)
		} else {
			ua.programs[name] = program
		}
	}

	return ua, nil
}

func NewUserAttributesFile(config map[string]string, attributes map[string]string) (ua *UserAttributes, err error) {
	opts := []cel.EnvOption{
		newAttributeUserUsername(),
		newAttributeUserGroups(),
		newAttributeUserDisplayName(),
		newAttributeUserEmail(),
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
		newAttributeUserPhoneExtension(),
		newAttributeUserStreetAddress(),
		newAttributeUserLocality(),
		newAttributeUserRegion(),
		newAttributeUserPostalCode(),
		newAttributeUserCountry(),
	}
}

func newAttributeUserUsername() cel.EnvOption {
	return cel.Variable(attributeKeyUserUsername, cel.StringType)
}

func newAttributeUserGroups() cel.EnvOption {
	return cel.Variable(attributeKeyUserGroups, cel.ListType(cel.StringType))
}

func newAttributeUserDisplayName() cel.EnvOption {
	return cel.Variable(attributeKeyUserDisplayName, cel.StringType)
}

func newAttributeUserEmail() cel.EnvOption {
	return cel.Variable(attributeKeyUserEmail, cel.StringType)
}

func newAttributeUserEmails() cel.EnvOption {
	return cel.Variable(attributeKeyUserEmails, cel.ListType(cel.StringType))
}

func newAttributeUserGivenName() cel.EnvOption {
	return cel.Variable(attributeKeyUserGivenName, cel.StringType)
}

func newAttributeUserMiddleName() cel.EnvOption {
	return cel.Variable(attributeKeyUserMiddleName, cel.StringType)
}

func newAttributeUserFamilyName() cel.EnvOption {
	return cel.Variable(attributeKeyUserFamilyName, cel.StringType)
}

func newAttributeUserNickname() cel.EnvOption {
	return cel.Variable(attributeKeyUserNickname, cel.StringType)
}

func newAttributeUserProfile() cel.EnvOption {
	return cel.Variable(attributeKeyUserProfile, cel.StringType)
}

func newAttributeUserPicture() cel.EnvOption {
	return cel.Variable(attributeKeyUserPicture, cel.StringType)
}

func newAttributeUserWebsite() cel.EnvOption {
	return cel.Variable(attributeKeyUserWebsite, cel.StringType)
}

func newAttributeUserGender() cel.EnvOption {
	return cel.Variable(attributeKeyUserGender, cel.StringType)
}

func newAttributeUserBirthdate() cel.EnvOption {
	return cel.Variable(attributeKeyUserBirthdate, cel.StringType)
}

func newAttributeUserZoneInfo() cel.EnvOption {
	return cel.Variable(attributeKeyUserZoneInfo, cel.StringType)
}

func newAttributeUserLocale() cel.EnvOption {
	return cel.Variable(attributeKeyUserLocale, cel.StringType)
}

func newAttributeUserPhoneNumber() cel.EnvOption {
	return cel.Variable(attributeKeyUserPhoneNumber, cel.StringType)
}

func newAttributeUserPhoneExtension() cel.EnvOption {
	return cel.Variable(attributeKeyUserPhoneExtension, cel.StringType)
}

func newAttributeUserStreetAddress() cel.EnvOption {
	return cel.Variable(attributeKeyUserStreetAddress, cel.StringType)
}

func newAttributeUserLocality() cel.EnvOption {
	return cel.Variable(attributeKeyUserLocality, cel.StringType)
}

func newAttributeUserRegion() cel.EnvOption {
	return cel.Variable(attributeKeyUserRegion, cel.StringType)
}

func newAttributeUserPostalCode() cel.EnvOption {
	return cel.Variable(attributeKeyUserPostalCode, cel.StringType)
}

func newAttributeUserCountry() cel.EnvOption {
	return cel.Variable(attributeKeyUserCountry, cel.StringType)
}
