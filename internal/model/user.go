package model

import (
	"fmt"
	"net/url"

	"golang.org/x/text/language"
)

// User represents the details of a user as defined in the file based user database.
type User struct {
	Username       string        `koanf:"username" yaml:"username,omitempty" toml:"username,omitempty" json:"username,omitempty"`
	DisplayName    string        `koanf:"displayname" yaml:"displayname,omitempty" toml:"displayname,omitempty" json:"displayname,omitempty"`
	Emails         []string      `koanf:"email" yaml:"email,omitempty" toml:"email,omitempty" json:"email,omitempty"`
	Groups         []string      `koanf:"groups" yaml:"groups,omitempty" toml:"groups,omitempty" json:"groups,omitempty"`
	GivenName      string        `koanf:"given_name" yaml:"given_name,omitempty" toml:"given_name,omitempty" json:"given_name,omitempty"`
	MiddleName     string        `koanf:"middle_name" yaml:"middle_name,omitempty" toml:"middle_name,omitempty" json:"middle_name,omitempty"`
	FamilyName     string        `koanf:"family_name" yaml:"family_name,omitempty" toml:"family_name,omitempty" json:"family_name,omitempty"`
	Nickname       string        `koanf:"nickname" yaml:"nickname,omitempty" toml:"nickname,omitempty" json:"nickname,omitempty"`
	Gender         string        `koanf:"gender" yaml:"gender,omitempty" toml:"gender,omitempty" json:"gender,omitempty"`
	Birthdate      string        `koanf:"birthdate" yaml:"birthdate,omitempty" toml:"birthdate,omitempty" json:"birthdate,omitempty"`
	Website        *url.URL      `koanf:"website" yaml:"website,omitempty" toml:"website,omitempty" json:"website,omitempty"`
	Profile        *url.URL      `koanf:"profile" yaml:"profile,omitempty" toml:"profile,omitempty" json:"profile,omitempty"`
	Picture        *url.URL      `koanf:"picture" yaml:"picture,omitempty" toml:"picture,omitempty" json:"picture,omitempty"`
	ZoneInfo       string        `koanf:"zoneinfo" yaml:"zoneinfo,omitempty" toml:"zoneinfo,omitempty" json:"zoneinfo,omitempty"`
	Locale         *language.Tag `koanf:"locale" yaml:"locale,omitempty" toml:"locale,omitempty" json:"locale,omitempty"`
	PhoneNumber    string        `koanf:"phone_number" yaml:"phone_number,omitempty" toml:"phone_number,omitempty" json:"phone_number,omitempty"`
	PhoneExtension string        `koanf:"phone_extension" yaml:"phone_extension,omitempty" toml:"phone_extension,omitempty" json:"phone_extension,omitempty"`

	Address *UserAddress `koanf:"address" yaml:"address,omitempty" toml:"address,omitempty" json:"address,omitempty"`

	Extra map[string]any `koanf:"extra" yaml:"extra,omitempty" toml:"extra,omitempty" json:"extra,omitempty"`
}

// UserAddress represents the address details of a User.
type UserAddress struct {
	StreetAddress string `koanf:"street_address" yaml:"street_address,omitempty" toml:"street_address,omitempty" json:"street_address,omitempty" jsonschema:"title=Street Address" jsonschema_description:"The street address for the user."`
	Locality      string `koanf:"locality" yaml:"locality,omitempty" toml:"locality,omitempty" json:"locality,omitempty" jsonschema:"title=Locality" jsonschema_description:"The locality for the user."`
	Region        string `koanf:"region" yaml:"region,omitempty" toml:"region,omitempty" json:"region,omitempty" jsonschema:"title=Region" jsonschema_description:"The region for the user."`
	PostalCode    string `koanf:"postal_code" yaml:"postal_code,omitempty" toml:"postal_code,omitempty" json:"postal_code,omitempty" jsonschema:"title=Postal Code" jsonschema_description:"The postal code or postcode for the user."`
	Country       string `koanf:"country" yaml:"country,omitempty" toml:"country,omitempty" json:"country,omitempty" jsonschema:"title=Country" jsonschema_description:"The country for the user."`
}

// GetUsername returns the username.
func (d *User) GetUsername() (username string) {
	return d.Username
}

// GetGroups returns the groups.
func (d *User) GetGroups() (groups []string) {
	return d.Groups
}

// GetDisplayName returns the display name.
func (d *User) GetDisplayName() (name string) {
	return d.DisplayName
}

// GetEmails returns the emails.
func (d *User) GetEmails() (emails []string) {
	return d.Emails
}

// GetGivenName returns the given name.
func (d *User) GetGivenName() (given string) {
	return d.GivenName
}

// GetFamilyName returns the family name.
func (d *User) GetFamilyName() (family string) {
	return d.FamilyName
}

// GetMiddleName returns the middle name.
func (d *User) GetMiddleName() (middle string) {
	return d.MiddleName
}

// GetNickname returns the nickname.
func (d *User) GetNickname() (nickname string) {
	return d.Nickname
}

// GetProfile returns the profile URL as a string.
func (d *User) GetProfile() (profile string) {
	return stringURL(d.Profile)
}

// GetPicture returns the picture URL as a string.
func (d *User) GetPicture() (picture string) {
	return stringURL(d.Picture)
}

// GetWebsite returns the website URL as a string.
func (d *User) GetWebsite() (website string) {
	return stringURL(d.Website)
}

// GetGender returns the gender.
func (d *User) GetGender() (gender string) {
	return d.Gender
}

// GetBirthdate returns the birthdate.
func (d *User) GetBirthdate() (birthdate string) {
	return d.Birthdate
}

// GetZoneInfo returns the zone information.
func (d *User) GetZoneInfo() (info string) {
	return d.ZoneInfo
}

// GetLocale returns the locale as a string.
func (d *User) GetLocale() (locale string) {
	if d.Locale == nil {
		return ""
	}

	return d.Locale.String()
}

// GetPhoneNumber returns the phone number without the extension.
func (d *User) GetPhoneNumber() (number string) {
	return d.PhoneNumber
}

// GetPhoneExtension returns the phone extension.
func (d *User) GetPhoneExtension() (extension string) {
	return d.PhoneExtension
}

// GetPhoneNumberRFC3966 returns the phone number and extension formatted as per RFC3966.
func (d *User) GetPhoneNumberRFC3966() (number string) {
	if d.PhoneNumber == "" {
		return ""
	}

	if d.PhoneExtension == "" {
		return d.PhoneNumber
	}

	return fmt.Sprintf("%s;ext=%s", d.PhoneNumber, d.PhoneExtension)
}

// GetStreetAddress returns the street address.
func (d *User) GetStreetAddress() (address string) {
	if d.Address == nil {
		return ""
	}

	return d.Address.StreetAddress
}

// GetLocality returns the locality.
func (d *User) GetLocality() (locality string) {
	if d.Address == nil {
		return ""
	}

	return d.Address.Locality
}

// GetRegion returns the region.
func (d *User) GetRegion() (region string) {
	if d.Address == nil {
		return ""
	}

	return d.Address.Region
}

// GetPostalCode returns the postal code.
func (d *User) GetPostalCode() (postcode string) {
	if d.Address == nil {
		return ""
	}

	return d.Address.PostalCode
}

// GetCountry returns the country.
func (d *User) GetCountry() (country string) {
	if d.Address == nil {
		return ""
	}

	return d.Address.Country
}

// GetExtra returns the extra attributes.
func (d *User) GetExtra() (extra map[string]any) {
	return d.Extra
}

func stringURL(uri *url.URL) string {
	if uri == nil {
		return ""
	}

	return uri.String()
}
