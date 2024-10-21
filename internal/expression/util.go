package expression

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
)

func optExtra(name string, attribute ExtraAttribute) (opt cel.EnvOption) {
	var t *types.Type

	switch attribute.GetValueType() {
	case "string":
		t = cel.StringType
	case "integer":
		t = cel.IntType
	case "boolean":
		t = cel.BoolType
	}

	if attribute.IsMultiValued() {
		return cel.Variable(name, cel.ListType(t))
	} else {
		return cel.Variable(name, t)
	}
}

func newAttributeUserUsername() cel.EnvOption {
	return cel.Variable(AttributeUserUsername, cel.StringType)
}

func newAttributeUserGroups() cel.EnvOption {
	return cel.Variable(AttributeUserGroups, cel.ListType(cel.StringType))
}

func newAttributeUserDisplayName() cel.EnvOption {
	return cel.Variable(AttributeUserDisplayName, cel.StringType)
}

func newAttributeUserEmail() cel.EnvOption {
	return cel.Variable(AttributeUserEmail, cel.StringType)
}

func newAttributeUserEmailVerified() cel.EnvOption {
	return cel.Variable(AttributeUserEmailVerified, cel.BoolType)
}

func newAttributeUserEmails() cel.EnvOption {
	return cel.Variable(AttributeUserEmails, cel.ListType(cel.StringType))
}

func newAttributeUserEmailsExtra() cel.EnvOption {
	return cel.Variable(AttributeUserEmailsExtra, cel.ListType(cel.StringType))
}

func newAttributeUserGivenName() cel.EnvOption {
	return cel.Variable(AttributeUserGivenName, cel.StringType)
}

func newAttributeUserMiddleName() cel.EnvOption {
	return cel.Variable(AttributeUserMiddleName, cel.StringType)
}

func newAttributeUserFamilyName() cel.EnvOption {
	return cel.Variable(AttributeUserFamilyName, cel.StringType)
}

func newAttributeUserNickname() cel.EnvOption {
	return cel.Variable(AttributeUserNickname, cel.StringType)
}

func newAttributeUserProfile() cel.EnvOption {
	return cel.Variable(AttributeUserProfile, cel.StringType)
}

func newAttributeUserPicture() cel.EnvOption {
	return cel.Variable(AttributeUserPicture, cel.StringType)
}

func newAttributeUserWebsite() cel.EnvOption {
	return cel.Variable(AttributeUserWebsite, cel.StringType)
}

func newAttributeUserGender() cel.EnvOption {
	return cel.Variable(AttributeUserGender, cel.StringType)
}

func newAttributeUserBirthdate() cel.EnvOption {
	return cel.Variable(AttributeUserBirthdate, cel.StringType)
}

func newAttributeUserZoneInfo() cel.EnvOption {
	return cel.Variable(AttributeUserZoneInfo, cel.StringType)
}

func newAttributeUserLocale() cel.EnvOption {
	return cel.Variable(AttributeUserLocale, cel.StringType)
}

func newAttributeUserPhoneNumber() cel.EnvOption {
	return cel.Variable(AttributeUserPhoneNumber, cel.StringType)
}

func newAttributeUserPhoneNumberVerified() cel.EnvOption {
	return cel.Variable(AttributeUserPhoneNumberVerified, cel.BoolType)
}

func newAttributeUserPhoneExtension() cel.EnvOption {
	return cel.Variable(AttributeUserPhoneExtension, cel.StringType)
}

func newAttributeUserPhoneNumberRFC3966() cel.EnvOption {
	return cel.Variable(AttributeUserPhoneNumberRFC3966, cel.StringType)
}

func newAttributeUserAddress() cel.EnvOption {
	return cel.Variable(AttributeUserAddress, cel.MapType(cel.StringType, cel.StringType))
}

func newAttributeUserStreetAddress() cel.EnvOption {
	return cel.Variable(AttributeUserStreetAddress, cel.StringType)
}

func newAttributeUserLocality() cel.EnvOption {
	return cel.Variable(AttributeUserLocality, cel.StringType)
}

func newAttributeUserRegion() cel.EnvOption {
	return cel.Variable(AttributeUserRegion, cel.StringType)
}

func newAttributeUserPostalCode() cel.EnvOption {
	return cel.Variable(AttributeUserPostalCode, cel.StringType)
}

func newAttributeUserCountry() cel.EnvOption {
	return cel.Variable(AttributeUserCountry, cel.StringType)
}

func newAttributeUpdatedAt() cel.EnvOption {
	return cel.Variable(AttributeUserUpdatedAt, cel.IntType)
}
