package authentication

import (
	"fmt"
	"net/mail"
	"net/url"

	"golang.org/x/text/language"
)

// UserDetails represent the details retrieved for a given user.
type UserDetails struct {
	Username    string
	DisplayName string
	Emails      []string
	Groups      []string
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
	GivenName      string
	FamilyName     string
	MiddleName     string
	Nickname       string
	Profile        *url.URL
	Picture        *url.URL
	Website        *url.URL
	Gender         string
	Birthdate      string
	ZoneInfo       string
	Locale         *language.Tag
	PhoneNumber    string
	PhoneExtension string
	Address        *UserDetailsAddress

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

func stringURL(uri *url.URL) string {
	if uri == nil {
		return ""
	}

	return uri.String()
}

type UserDetailsAddress struct {
	StreetAddress string
	Locality      string
	Region        string
	PostalCode    string
	Country       string
}

type ldapUserProfile struct {
	DN          string
	Emails      []string
	DisplayName string
	Username    string
	MemberOf    []string
}

type ldapUserProfileExtended struct {
	GivenName      string
	FamilyName     string
	MiddleName     string
	Nickname       string
	Profile        string
	Picture        string
	Website        string
	Gender         string
	Birthdate      string
	ZoneInfo       string
	Locale         string
	PhoneNumber    string
	PhoneExtension string
	Address        *UserDetailsAddress
	Extra          map[string]any

	*ldapUserProfile
}

// LDAPSupportedFeatures represents features which a server may support which are implemented in code.
type LDAPSupportedFeatures struct {
	Extensions   LDAPSupportedExtensions
	ControlTypes LDAPSupportedControlTypes
}

// LDAPSupportedExtensions represents extensions which a server may support which are implemented in code.
type LDAPSupportedExtensions struct {
	TLS           bool
	PwdModifyExOp bool
}

// LDAPSupportedControlTypes represents control types which a server may support which are implemented in code.
type LDAPSupportedControlTypes struct {
	MsftPwdPolHints           bool
	MsftPwdPolHintsDeprecated bool
}

// Level is the type representing a level of authentication.
type Level int

const (
	// NotAuthenticated if the user is not authenticated yet.
	NotAuthenticated Level = iota

	// OneFactor if the user has passed first factor only.
	OneFactor

	// TwoFactor if the user has passed two factors.
	TwoFactor
)

// String returns a string representation of an authentication.Level.
func (l Level) String() string {
	switch l {
	case NotAuthenticated:
		return "not_authenticated"
	case OneFactor:
		return "one_factor"
	case TwoFactor:
		return "two_factor"
	default:
		return "invalid"
	}
}
