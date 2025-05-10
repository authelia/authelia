package model

import (
	"fmt"
	"net/url"

	"golang.org/x/text/language"
)

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

type UserAddress struct {
	StreetAddress string `koanf:"street_address" yaml:"street_address,omitempty" toml:"street_address,omitempty" json:"street_address,omitempty" jsonschema:"title=Street Address" jsonschema_description:"The street address for the user."`
	Locality      string `koanf:"locality" yaml:"locality,omitempty" toml:"locality,omitempty" json:"locality,omitempty" jsonschema:"title=Locality" jsonschema_description:"The locality for the user."`
	Region        string `koanf:"region" yaml:"region,omitempty" toml:"region,omitempty" json:"region,omitempty" jsonschema:"title=Region" jsonschema_description:"The region for the user."`
	PostalCode    string `koanf:"postal_code" yaml:"postal_code,omitempty" toml:"postal_code,omitempty" json:"postal_code,omitempty" jsonschema:"title=Postal Code" jsonschema_description:"The postal code or postcode for the user."`
	Country       string `koanf:"country" yaml:"country,omitempty" toml:"country,omitempty" json:"country,omitempty" jsonschema:"title=Country" jsonschema_description:"The country for the user."`
}

func (d *User) GetUsername() (username string) {
	return d.Username
}

func (d *User) GetGroups() (groups []string) {
	return d.Groups
}

func (d *User) GetDisplayName() (name string) {
	return d.DisplayName
}

func (d *User) GetEmails() (emails []string) {
	return d.Emails
}

func (d *User) GetGivenName() (given string) {
	return d.GivenName
}

func (d *User) GetFamilyName() (family string) {
	return d.FamilyName
}

func (d *User) GetMiddleName() (middle string) {
	return d.MiddleName
}

func (d *User) GetNickname() (nickname string) {
	return d.Nickname
}

func (d *User) GetProfile() (profile string) {
	return stringURL(d.Profile)
}

func (d *User) GetPicture() (picture string) {
	return stringURL(d.Picture)
}

func (d *User) GetWebsite() (website string) {
	return stringURL(d.Website)
}

func (d *User) GetGender() (gender string) {
	return d.Gender
}

func (d *User) GetBirthdate() (birthdate string) {
	return d.Birthdate
}

func (d *User) GetZoneInfo() (info string) {
	return d.ZoneInfo
}

func (d *User) GetLocale() (locale string) {
	if d.Locale == nil {
		return ""
	}

	return d.Locale.String()
}

func (d *User) GetPhoneNumber() (number string) {
	return d.PhoneNumber
}

func (d *User) GetPhoneExtension() (extension string) {
	return d.PhoneExtension
}

func (d *User) GetPhoneNumberRFC3966() (number string) {
	if d.PhoneNumber == "" {
		return ""
	}

	if d.PhoneExtension == "" {
		return d.PhoneNumber
	}

	return fmt.Sprintf("%s;ext=%s", d.PhoneNumber, d.PhoneExtension)
}

func (d *User) GetStreetAddress() (address string) {
	if d.Address == nil {
		return ""
	}

	return d.Address.StreetAddress
}

func (d *User) GetLocality() (locality string) {
	if d.Address == nil {
		return ""
	}

	return d.Address.Locality
}

func (d *User) GetRegion() (region string) {
	if d.Address == nil {
		return ""
	}

	return d.Address.Region
}

func (d *User) GetPostalCode() (postcode string) {
	if d.Address == nil {
		return ""
	}

	return d.Address.PostalCode
}

func (d *User) GetCountry() (country string) {
	if d.Address == nil {
		return ""
	}

	return d.Address.Country
}

func (d *User) GetExtra() (extra map[string]any) {
	return d.Extra
}

func stringURL(uri *url.URL) string {
	if uri == nil {
		return ""
	}

	return uri.String()
}
