package expression

import "github.com/google/cel-go/interpreter"

type UserDetailerActivation struct {
	detailer UserDetailer
}

func (a *UserDetailerActivation) ResolveName(name string) (object any, found bool) {
	switch name {
	case attributeKeyUserUsername:
		return a.detailer.GetUsername(), true
	case attributeKeyUserGroups:
		return a.detailer.GetGroups(), true
	case attributeKeyUserDisplayName:
		return a.detailer.GetDisplayName(), true
	case attributeKeyUserEmail:
		if emails := a.detailer.GetEmails(); len(emails) != 0 {
			return emails[0], true
		}

		return "", true
	case attributeKeyUserEmails:
		return a.detailer.GetEmails(), true
	case attributeKeyUserGivenName:
		return a.detailer.GetGivenName(), true
	case attributeKeyUserMiddleName:
		return a.detailer.GetMiddleName(), true
	case attributeKeyUserFamilyName:
		return a.detailer.GetFamilyName(), true
	case attributeKeyUserNickname:
		return a.detailer.GetNickname(), true
	case attributeKeyUserProfile:
		return a.detailer.GetProfile(), true
	case attributeKeyUserPicture:
		return a.detailer.GetPicture(), true
	case attributeKeyUserWebsite:
		return a.detailer.GetWebsite(), true
	case attributeKeyUserGender:
		return a.detailer.GetGender(), true
	case attributeKeyUserBirthdate:
		return a.detailer.GetBirthdate(), true
	case attributeKeyUserZoneInfo:
		return a.detailer.GetZoneInfo(), true
	case attributeKeyUserLocale:
		return a.detailer.GetLocale(), true
	case attributeKeyUserPhoneNumber:
		return a.detailer.GetPhoneNumber(), true
	case attributeKeyUserPhoneExtension:
		return a.detailer.GetPhoneExtension(), true
	case attributeKeyUserStreetAddress:
		return a.detailer.GetStreetAddress(), true
	case attributeKeyUserLocality:
		return a.detailer.GetLocality(), true
	case attributeKeyUserRegion:
		return a.detailer.GetRegion(), true
	case attributeKeyUserPostalCode:
		return a.detailer.GetPostalCode(), true
	case attributeKeyUserCountry:
		return a.detailer.GetCountry(), true
	default:
		extra := a.detailer.GetExtra()

		if extra != nil {
			if object, found = extra[attributeKeyUserProfile]; found {
				return object, true
			}
		}
	}

	return nil, false
}

func (a *UserDetailerActivation) Parent() interpreter.Activation {
	return nil
}

type UserDetailer interface {
	GetUsername() (username string)
	GetGroups() (groups []string)
	GetDisplayName() (name string)
	GetEmails() (emails []string)
	GetGivenName() (given string)
	GetFamilyName() (family string)
	GetMiddleName() (middle string)
	GetNickname() (nickname string)
	GetProfile() (profile string)
	GetPicture() (picture string)
	GetWebsite() (website string)
	GetGender() (gender string)
	GetBirthdate() (birthdate string)
	GetZoneInfo() (info string)
	GetLocale() (locale string)
	GetPhoneNumber() (number string)
	GetPhoneExtension() (extension string)
	GetOpenIDConnectPhoneNumber() (number string)
	GetStreetAddress() (address string)
	GetLocality() (locality string)
	GetRegion() (region string)
	GetPostalCode() (postcode string)
	GetCountry() (country string)
	GetExtra() (extra map[string]any)
}
