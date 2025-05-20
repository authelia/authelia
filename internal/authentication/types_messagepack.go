package authentication

import (
	"fmt"
	"net/mail"
	"net/url"

	"golang.org/x/text/language"
)

//go:generate codecgen -o types_messagepack_gen.go types_messagepack.go

// UserDetails represent the details retrieved for a given user.
type UserDetails struct {
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name,omitempty"`
	Emails      []string `json:"emails,omitempty"`
	Groups      []string `json:"groups,omitempty"`
}

// Addresses returns the Emails []string as []mail.Address formatted with DisplayName as the Name attribute.
func (d *UserDetails) Addresses() (addresses []mail.Address) {
	if len(d.Emails) == 0 {
		return nil
	}

	addresses = make([]mail.Address, len(d.Emails))

	for i, email := range d.Emails {
		addresses[i] = mail.Address{
			Name:    d.DisplayName,
			Address: email,
		}
	}

	return addresses
}

func (d *UserDetails) GetUsername() (username string) {
	return d.Username
}

func (d *UserDetails) GetGroups() (groups []string) {
	return d.Groups
}

func (d *UserDetails) GetDisplayName() (name string) {
	return d.DisplayName
}

func (d *UserDetails) GetEmails() (emails []string) {
	return d.Emails
}

// UserDetailsExtended represents the extended details retrieved for a given user.
type UserDetailsExtended struct {
	GivenName      string              `json:"given,omitempty"`
	FamilyName     string              `json:"family,omitempty"`
	MiddleName     string              `json:"middle,omitempty"`
	Nickname       string              `json:"nick,omitempty"`
	Profile        *url.URL            `json:"profile,omitempty"`
	Picture        *url.URL            `json:"picture,omitempty"`
	Website        *url.URL            `json:"website,omitempty"`
	Gender         string              `json:"gender,omitempty"`
	Birthdate      string              `json:"birthdate,omitempty"`
	ZoneInfo       string              `json:"zone,omitempty"`
	Locale         *language.Tag       `json:"locale,omitempty"`
	PhoneNumber    string              `json:"phone,omitempty"`
	PhoneExtension string              `json:"ext,omitempty"`
	Address        *UserDetailsAddress `json:"address,omitempty"`

	Extra map[string]any

	*UserDetails
}

func (d *UserDetailsExtended) GetGivenName() (given string) {
	return d.GivenName
}

func (d *UserDetailsExtended) GetFamilyName() (family string) {
	return d.FamilyName
}

func (d *UserDetailsExtended) GetMiddleName() (middle string) {
	return d.MiddleName
}

func (d *UserDetailsExtended) GetNickname() (nickname string) {
	return d.Nickname
}

func (d *UserDetailsExtended) GetProfile() (profile string) {
	return stringURL(d.Profile)
}

func (d *UserDetailsExtended) GetPicture() (picture string) {
	return stringURL(d.Picture)
}

func (d *UserDetailsExtended) GetWebsite() (website string) {
	return stringURL(d.Website)
}

func (d *UserDetailsExtended) GetGender() (gender string) {
	return d.Gender
}

func (d *UserDetailsExtended) GetBirthdate() (birthdate string) {
	return d.Birthdate
}

func (d *UserDetailsExtended) GetZoneInfo() (info string) {
	return d.ZoneInfo
}

func (d *UserDetailsExtended) GetLocale() (locale string) {
	if d.Locale == nil {
		return ""
	}

	return d.Locale.String()
}

func (d *UserDetailsExtended) GetPhoneNumber() (number string) {
	return d.PhoneNumber
}

func (d *UserDetailsExtended) GetPhoneExtension() (extension string) {
	return d.PhoneExtension
}

func (d *UserDetailsExtended) GetPhoneNumberRFC3966() (number string) {
	if d.PhoneNumber == "" {
		return ""
	}

	if d.PhoneExtension == "" {
		return d.PhoneNumber
	}

	return fmt.Sprintf("%s;ext=%s", d.PhoneNumber, d.PhoneExtension)
}

func (d *UserDetailsExtended) GetStreetAddress() (address string) {
	if d.Address == nil {
		return ""
	}

	return d.Address.StreetAddress
}

func (d *UserDetailsExtended) GetLocality() (locality string) {
	if d.Address == nil {
		return ""
	}

	return d.Address.Locality
}

func (d *UserDetailsExtended) GetRegion() (region string) {
	if d.Address == nil {
		return ""
	}

	return d.Address.Region
}

func (d *UserDetailsExtended) GetPostalCode() (postcode string) {
	if d.Address == nil {
		return ""
	}

	return d.Address.PostalCode
}

func (d *UserDetailsExtended) GetCountry() (country string) {
	if d.Address == nil {
		return ""
	}

	return d.Address.Country
}

func (d *UserDetailsExtended) GetExtra() (extra map[string]any) {
	return d.Extra
}

type UserDetailsAddress struct {
	StreetAddress string `json:"street,omitempty"`
	Locality      string `json:"locality,omitempty"`
	Region        string `json:"region,omitempty"`
	PostalCode    string `json:"postcode,omitempty"`
	Country       string `json:"country,omitempty"`
}
