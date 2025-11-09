package expression

import (
	"time"

	"github.com/google/cel-go/interpreter"
)

func NewUserDetailerActivation(parent interpreter.Activation, detailer UserDetailer, updated time.Time) *UserDetailerActivation {
	return &UserDetailerActivation{
		parent: parent,
		detailer: &UserAttributeResolverDetailer{
			UserDetailer: detailer,
			updated:      updated,
		},
	}
}

type UserDetailerActivation struct {
	parent   interpreter.Activation
	detailer ExtendedUserDetailer
}

//nolint:gocyclo
func (a *UserDetailerActivation) ResolveName(name string) (object any, found bool) {
	switch name {
	case AttributeUserUsername:
		return a.detailer.GetUsername(), true
	case AttributeUserGroups:
		return a.detailer.GetGroups(), true
	case AttributeUserDisplayName:
		return a.detailer.GetDisplayName(), true
	case AttributeUserEmail:
		if emails := a.detailer.GetEmails(); len(emails) != 0 {
			return emails[0], true
		}

		return "", true
	case AttributeUserEmails:
		return a.detailer.GetEmails(), true
	case AttributeUserEmailsExtra:
		emails := a.detailer.GetEmails()
		if len(emails) < 2 {
			return nil, true
		}

		return emails[1:], true
	case AttributeUserEmailVerified:
		return true, true
	case AttributeUserGivenName:
		return a.detailer.GetGivenName(), true
	case AttributeUserMiddleName:
		return a.detailer.GetMiddleName(), true
	case AttributeUserFamilyName:
		return a.detailer.GetFamilyName(), true
	case AttributeUserNickname:
		return a.detailer.GetNickname(), true
	case AttributeUserProfile:
		return a.detailer.GetProfile(), true
	case AttributeUserPicture:
		return a.detailer.GetPicture(), true
	case AttributeUserWebsite:
		return a.detailer.GetWebsite(), true
	case AttributeUserGender:
		return a.detailer.GetGender(), true
	case AttributeUserBirthdate:
		return a.detailer.GetBirthdate(), true
	case AttributeUserZoneInfo:
		return a.detailer.GetZoneInfo(), true
	case AttributeUserLocale:
		return a.detailer.GetLocale(), true
	case AttributeUserPhoneNumber:
		return a.detailer.GetPhoneNumber(), true
	case AttributeUserPhoneNumberRFC3966:
		return a.detailer.GetPhoneNumberRFC3966(), true
	case AttributeUserPhoneExtension:
		return a.detailer.GetPhoneExtension(), true
	case AttributeUserPhoneNumberVerified:
		if a.detailer.GetPhoneNumberRFC3966() == "" {
			return nil, true
		}

		return false, true
	case AttributeUserAddress:
		return a.address(), true
	case AttributeUserStreetAddress:
		return a.detailer.GetStreetAddress(), true
	case AttributeUserLocality:
		return a.detailer.GetLocality(), true
	case AttributeUserRegion:
		return a.detailer.GetRegion(), true
	case AttributeUserPostalCode:
		return a.detailer.GetPostalCode(), true
	case AttributeUserCountry:
		return a.detailer.GetCountry(), true
	case AttributeUserUpdatedAt:
		return a.detailer.GetUpdatedAt().Unix(), true
	default:
		extra := a.detailer.GetExtra()

		if extra != nil {
			if object, found = extra[name]; found {
				return object, true
			}
		}
	}

	if a.parent != nil {
		return a.parent.ResolveName(name)
	}

	return nil, false
}

func (a *UserDetailerActivation) address() (address map[string]any) {
	if a.detailer == nil {
		return nil
	}

	address = map[string]any{}

	var value string

	if value = a.detailer.GetStreetAddress(); value != "" {
		address[AttributeUserStreetAddress] = value
	}

	if value = a.detailer.GetLocality(); value != "" {
		address[AttributeUserLocality] = value
	}

	if value = a.detailer.GetRegion(); value != "" {
		address[AttributeUserRegion] = value
	}

	if value = a.detailer.GetPostalCode(); value != "" {
		address[AttributeUserPostalCode] = value
	}

	if value = a.detailer.GetCountry(); value != "" {
		address[AttributeUserCountry] = value
	}

	if len(address) == 0 {
		return nil
	}

	return address
}

func (a *UserDetailerActivation) Parent() interpreter.Activation {
	return a.parent
}
