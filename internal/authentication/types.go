package authentication

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"net/url"

	"golang.org/x/text/language"

	"github.com/authelia/authelia/v4/internal/clock"
)

// UserDetails represent the details retrieved for a given user.
type UserDetails struct {
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Emails      []string `json:"emails"`
	Groups      []string `json:"groups"`
}

type FieldMetadata struct {
	Required    bool   `json:"required"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Type        string `json:"type"`
	MaxLength   int    `json:"maxLength,omitempty"`
	Pattern     string `json:"pattern,omitempty"`
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
	GivenName      string              `json:"first_name,omitempty"`
	FamilyName     string              `json:"last_name,omitempty"`
	MiddleName     string              `json:"middle_name,omitempty"`
	Nickname       string              `json:"nickname,omitempty"`
	Profile        *url.URL            `json:"profile,omitempty"`
	Picture        *url.URL            `json:"picture,omitempty"`
	Website        *url.URL            `json:"website,omitempty"`
	Gender         string              `json:"gender,omitempty"`
	Birthdate      string              `json:"birthdate,omitempty"`
	ZoneInfo       string              `json:"zone_info,omitempty"`
	Locale         *language.Tag       `json:"locale,omitempty"`
	PhoneNumber    string              `json:"phone_number,omitempty"`
	PhoneExtension string              `json:"phone_extension,omitempty"`
	Address        *UserDetailsAddress `json:"address,omitempty"`

	Extra map[string]any `json:"extra,omitempty"`

	*UserDetails

	Password      string   `json:"-"`
	CommonName    string   `json:"cn,omitempty"`
	ObjectClasses []string `json:"object_classes,omitempty"`
}

// UnmarshalJSON allows the "password" field to be unmarshalled but not included when the struct is marshalled. Effectively making the password ingest-only.
func (d *UserDetailsExtended) UnmarshalJSON(data []byte) error {
	type Alias UserDetailsExtended

	aux := &struct {
		Password string `json:"password"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	d.Password = aux.Password
	return nil
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

type UserDetailsExtendedBuilder struct {
	data *UserDetailsExtended
}

// NewUser creates a new user builder with username and password
func NewUser(username, password string) *UserDetailsExtendedBuilder {
	return &UserDetailsExtendedBuilder{
		data: &UserDetailsExtended{
			Password: password,
			UserDetails: &UserDetails{
				Username: username,
				Emails:   []string{},
				Groups:   []string{},
			},
			ObjectClasses: []string{},
		},
	}
}

func (b *UserDetailsExtendedBuilder) WithDisplayName(name string) *UserDetailsExtendedBuilder {
	b.data.UserDetails.DisplayName = name
	return b
}

func (b *UserDetailsExtendedBuilder) WithEmail(email string) *UserDetailsExtendedBuilder {
	b.data.UserDetails.Emails = []string{email}
	return b
}

func (b *UserDetailsExtendedBuilder) WithEmails(emails []string) *UserDetailsExtendedBuilder {
	b.data.UserDetails.Emails = emails
	return b
}

func (b *UserDetailsExtendedBuilder) WithGroups(groups []string) *UserDetailsExtendedBuilder {
	b.data.UserDetails.Groups = groups
	return b
}

func (b *UserDetailsExtendedBuilder) WithCommonName(cn string) *UserDetailsExtendedBuilder {
	b.data.CommonName = cn
	return b
}

func (b *UserDetailsExtendedBuilder) WithGivenName(given string) *UserDetailsExtendedBuilder {
	b.data.GivenName = given
	return b
}

func (b *UserDetailsExtendedBuilder) WithFamilyName(family string) *UserDetailsExtendedBuilder {
	b.data.FamilyName = family
	return b
}

func (b *UserDetailsExtendedBuilder) WithMiddleName(middle string) *UserDetailsExtendedBuilder {
	b.data.MiddleName = middle
	return b
}

func (b *UserDetailsExtendedBuilder) WithNickname(nickname string) *UserDetailsExtendedBuilder {
	b.data.Nickname = nickname
	return b
}

func (b *UserDetailsExtendedBuilder) WithObjectClasses(classes []string) *UserDetailsExtendedBuilder {
	b.data.ObjectClasses = classes
	return b
}

func (b *UserDetailsExtendedBuilder) WithGender(gender string) *UserDetailsExtendedBuilder {
	b.data.Gender = gender
	return b
}

func (b *UserDetailsExtendedBuilder) WithBirthdate(birthdate string) *UserDetailsExtendedBuilder {
	b.data.Birthdate = birthdate
	return b
}

func (b *UserDetailsExtendedBuilder) WithPhoneNumber(phone string) *UserDetailsExtendedBuilder {
	b.data.PhoneNumber = phone
	return b
}

func (b *UserDetailsExtendedBuilder) WithProfile(profileURL string) *UserDetailsExtendedBuilder {
	if profileURL != "" {
		if uri, err := url.Parse(profileURL); err == nil {
			b.data.Profile = uri
		}
	}
	return b
}

func (b *UserDetailsExtendedBuilder) WithPicture(pictureURL string) *UserDetailsExtendedBuilder {
	if pictureURL != "" {
		if uri, err := url.Parse(pictureURL); err == nil {
			b.data.Picture = uri
		}
	}
	return b
}

func (b *UserDetailsExtendedBuilder) WithWebsite(websiteURL string) *UserDetailsExtendedBuilder {
	if websiteURL != "" {
		if uri, err := url.Parse(websiteURL); err == nil {
			b.data.Website = uri
		}
	}
	return b
}

func (b *UserDetailsExtendedBuilder) WithLocale(locale string) *UserDetailsExtendedBuilder {
	if locale != "" {
		if tag, err := language.Parse(locale); err == nil {
			b.data.Locale = &tag
		}
	}
	return b
}

func (b *UserDetailsExtendedBuilder) WithAddress(street, locality, region, postal, country string) *UserDetailsExtendedBuilder {
	b.data.Address = &UserDetailsAddress{
		StreetAddress: street,
		Locality:      locality,
		Region:        region,
		PostalCode:    postal,
		Country:       country,
	}
	return b
}

func (b *UserDetailsExtendedBuilder) WithExtra(key string, value any) *UserDetailsExtendedBuilder {
	if b.data.Extra == nil {
		b.data.Extra = make(map[string]any)
	}
	b.data.Extra[key] = value
	return b
}

func (b *UserDetailsExtendedBuilder) WithDefaultLDAPObjectClasses() *UserDetailsExtendedBuilder {
	b.data.ObjectClasses = []string{"top", "person", "organizationalPerson", "inetOrgPerson"}
	return b
}

func (b *UserDetailsExtendedBuilder) Build() *UserDetailsExtended {
	return b.data
}

func stringURL(uri *url.URL) string {
	if uri == nil {
		return ""
	}

	return uri.String()
}

type UserDetailsAddress struct {
	StreetAddress string `json:"street_address,omitempty"`
	Locality      string `json:"locality,omitempty"`
	Region        string `json:"region,omitempty"`
	PostalCode    string `json:"postal_code,omitempty"`
	Country       string `json:"country,omitempty"`
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
	OIDs []string

	TLS       bool
	PwdModify bool
	WhoAmI    bool
}

// LDAPSupportedControlTypes represents control types which a server may support which are implemented in code.
type LDAPSupportedControlTypes struct {
	OIDs []string

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

type Context interface {
	context.Context

	GetUserProvider() UserProvider
	GetLogger() *logrus.Entry
	GetClock() clock.Provider
}

func NewPoolCtxErr(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return &PoolErr{
			err:             err,
			isDeadlineError: true,
		}
	}

	return &PoolErr{err: err}
}

type PoolErr struct {
	err             error
	isDeadlineError bool
}

func (e *PoolErr) Error() string {
	return e.err.Error()
}

func (e *PoolErr) Is(target error) bool {
	return errors.Is(e.err, target)
}

func (e *PoolErr) Unwrap() error {
	return e.err
}

func (e *PoolErr) IsDeadlineError() bool {
	return e.isDeadlineError
}

type LDAPBaseClient interface {
	ldap.Client

	GSSAPIBind(client ldap.GSSAPIClient, servicePrincipal, authzid string) (err error)
	GSSAPIBindRequest(client ldap.GSSAPIClient, req *ldap.GSSAPIBindRequest) (err error)
	GSSAPIBindRequestWithAPOptions(client ldap.GSSAPIClient, req *ldap.GSSAPIBindRequest, APOptions []int) (err error)
	MD5Bind(host, username, password string) error
	DigestMD5Bind(digestMD5BindRequest *ldap.DigestMD5BindRequest) (*ldap.DigestMD5BindResult, error)
	NTLMChallengeBind(challenge *ldap.NTLMBindRequest) (result *ldap.NTLMBindResult, err error)
	NTLMBindWithHash(domain, username, hash string) (err error)
	NTLMBind(domain, username, password string) (err error)
	WhoAmI(controls []ldap.Control) (result *ldap.WhoAmIResult, err error)
}

type LDAPExtendedClient interface {
	LDAPBaseClient

	Features() (features LDAPSupportedFeatures)
}
