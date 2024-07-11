package expression

import "github.com/google/cel-go/interpreter"

type UserDetailerActivation struct {
	detailer UserDetailer
}

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
		return a.detailer.GetOpenIDConnectPhoneNumber(), true
	case AttributeUserPhoneExtension:
		return a.detailer.GetPhoneExtension(), true
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
	default:
		extra := a.detailer.GetExtra()

		if extra != nil {
			if object, found = extra[AttributeUserProfile]; found {
				return object, true
			}
		}
	}

	return nil, false
}

func (a *UserDetailerActivation) Parent() interpreter.Activation {
	return nil
}
