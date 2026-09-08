package expression

import (
	"time"

	"github.com/authelia/authelia/v4/internal/model"
)

// ExtraAttribute represents the definition of an extra user attribute.
type ExtraAttribute interface {
	IsMultiValued() (multi bool)
	GetValueType() (vtype string)
}

// UserAttributeResolver is an interface which resolves the value of a user attribute.
type UserAttributeResolver interface {
	Resolve(name string, detailer UserDetailer, updated time.Time) (object any, found bool)
	ResolveWithExtra(name string, detailer UserDetailer, updated time.Time, extra map[string]any) (object any, found bool)

	model.StartupCheck
}

// UserAttributeResolverDetailer is a UserDetailer which also records the time the details were updated.
type UserAttributeResolverDetailer struct {
	UserDetailer

	updated time.Time
}

// GetUpdatedAt returns the time the user details were updated.
func (d *UserAttributeResolverDetailer) GetUpdatedAt() time.Time {
	return d.updated
}

// ExtendedUserDetailer is a UserDetailer which also provides the time the details were updated.
type ExtendedUserDetailer interface {
	UserDetailer

	GetUpdatedAt() time.Time
}

// UserDetailer is an interface which provides the details of a user.
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
	GetPhoneNumberRFC3966() (number string)
	GetStreetAddress() (address string)
	GetLocality() (locality string)
	GetRegion() (region string)
	GetPostalCode() (postcode string)
	GetCountry() (country string)
	GetExtra() (extra map[string]any)
}
