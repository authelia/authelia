package authentication

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	netmail "net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// GetBaseRequiredAttributesForImplementation returns the base required attributes for a given user management implementation.
func GetBaseRequiredAttributesForImplementation(implementation string) []string {
	switch implementation {
	case schema.LDAPImplementationRFC2307bis:
		return []string{"username", "password", "family_name", "mail"}
	case schema.FileImplementation:
		return []string{"username", "password", "display_name", "mail"}
	default:
		return []string{"username", "password", "family_name", "mail"}
	}
}

// UserDetails represent the details retrieved for a given user.
type UserDetails struct {
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Emails      []string `json:"mail"`
	Groups      []string `json:"groups"`
}

type UserManagementAttributeMetadata struct {
	Type     AttributeType `json:"type"`
	Multiple bool          `json:"multiple,omitempty"`
}

// Addresses returns the Emails []string as []mail.Address formatted with DisplayName as the Name attribute.
func (d *UserDetails) Addresses() (addresses []netmail.Address) {
	if len(d.Emails) == 0 {
		return nil
	}

	addresses = make([]netmail.Address, len(d.Emails))

	for i, email := range d.Emails {
		addresses[i] = netmail.Address{
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
	GivenName      string              `json:"given_name,omitempty"`
	FamilyName     string              `json:"family_name,omitempty"`
	MiddleName     string              `json:"middle_name,omitempty"`
	CommonName     string              `json:"common_name,omitempty"`
	Nickname       string              `json:"nickname,omitempty"`
	Profile        *url.URL            `json:"profile,omitempty"`
	Picture        *url.URL            `json:"picture,omitempty"`
	Website        *url.URL            `json:"website,omitempty"`
	Gender         string              `json:"gender,omitempty"`
	Birthdate      string              `json:"birthdate,omitempty"`
	ZoneInfo       string              `json:"zoneinfo,omitempty"`
	Locale         *language.Tag       `json:"locale,omitempty"`
	PhoneNumber    string              `json:"phone_number,omitempty"`
	PhoneExtension string              `json:"phone_extension,omitempty"`
	Address        *UserDetailsAddress `json:"address,omitempty"`

	Extra map[string]any `json:"extra,omitempty"`

	*UserDetails

	Password string `json:"-"`

	LastLoggedIn       *time.Time `json:"last_logged_in,omitempty"`
	LastPasswordChange *time.Time `json:"last_password_change,omitempty"`
	UserCreatedAt      *time.Time `json:"user_created_at,omitempty"`
	Method             string     `json:"method,omitempty"`
	HasTOTP            bool       `json:"has_totp,omitempty"`
	HasWebAuthn        bool       `json:"has_webauthn,omitempty"`
	HasDuo             bool       `json:"has_duo,omitempty"`
}

// UnmarshalJSON allows the "password" field to be unmarshalled but not included when the struct is marshalled. Effectively making the password ingest-only.
//
//nolint:gocyclo
func (d *UserDetailsExtended) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	var (
		password string
		profile  string
		picture  string
		website  string
		locale   string
		mail     json.RawMessage
	)

	if passwordData, ok := raw["password"]; ok {
		if err := json.Unmarshal(passwordData, &password); err != nil {
			return fmt.Errorf("invalid password: %w", err)
		}

		delete(raw, "password")
	}

	if mailData, ok := raw["mail"]; ok {
		mail = mailData

		delete(raw, "mail")
	}

	if profileData, ok := raw["profile"]; ok {
		if err := json.Unmarshal(profileData, &profile); err != nil {
			return fmt.Errorf("invalid profile: %w", err)
		}

		delete(raw, "profile")
	}

	if pictureData, ok := raw["picture"]; ok {
		if err := json.Unmarshal(pictureData, &picture); err != nil {
			return fmt.Errorf("invalid picture: %w", err)
		}

		delete(raw, "picture")
	}

	if websiteData, ok := raw["website"]; ok {
		if err := json.Unmarshal(websiteData, &website); err != nil {
			return fmt.Errorf("invalid website: %w", err)
		}

		delete(raw, "website")
	}

	if localeData, ok := raw["locale"]; ok {
		if err := json.Unmarshal(localeData, &locale); err != nil {
			return fmt.Errorf("invalid locale: %w", err)
		}

		delete(raw, "locale")
	}

	// Marshal back to JSON without special fields.
	remaining, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	type Alias UserDetailsExtended

	if err := json.Unmarshal(remaining, (*Alias)(d)); err != nil {
		return err
	}

	d.Password = password

	// Handle 'mail' field.
	if len(mail) > 0 {
		var mailStr string
		if err := json.Unmarshal(mail, &mailStr); err == nil {
			if mailStr != "" {
				if d.UserDetails == nil {
					d.UserDetails = &UserDetails{}
				}

				d.Emails = []string{mailStr}
			}
		} else {
			var mailArr []string
			if err := json.Unmarshal(mail, &mailArr); err != nil {
				return fmt.Errorf("mail must be a string or array of strings: %w", err)
			}

			if d.UserDetails == nil {
				d.UserDetails = &UserDetails{}
			}

			d.Emails = mailArr
		}
	}

	if profile != "" {
		parsedURL, err := url.Parse(profile)
		if err != nil {
			return fmt.Errorf("invalid profile URL: %w", err)
		}

		d.Profile = parsedURL
	}

	if picture != "" {
		parsedURL, err := url.Parse(picture)
		if err != nil {
			return fmt.Errorf("invalid picture URL: %w", err)
		}

		d.Picture = parsedURL
	}

	if website != "" {
		parsedURL, err := url.Parse(website)
		if err != nil {
			return fmt.Errorf("invalid website URL: %w", err)
		}

		d.Website = parsedURL
	}

	if locale != "" {
		tag, err := language.Parse(locale)
		if err != nil {
			return fmt.Errorf("invalid locale: %w", err)
		}

		d.Locale = &tag
	}

	return nil
}

// MarshalJSON converts URL and Locale fields to strings for JSON output.
func (d *UserDetailsExtended) MarshalJSON() ([]byte, error) {
	type Alias UserDetailsExtended

	aux := &struct {
		Picture string `json:"picture,omitempty"`
		Profile string `json:"profile,omitempty"`
		Website string `json:"website,omitempty"`
		Locale  string `json:"locale,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}

	if d.Profile != nil {
		aux.Profile = d.Profile.String()
	}

	if d.Picture != nil {
		aux.Picture = d.Picture.String()
	}

	if d.Website != nil {
		aux.Website = d.Website.String()
	}

	if d.Locale != nil {
		aux.Locale = d.Locale.String()
	}

	aux.Alias.Profile = nil
	aux.Alias.Picture = nil
	aux.Alias.Website = nil
	aux.Alias.Locale = nil

	return json.Marshal(aux)
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

// NewUser creates a new user builder with username and password.
func NewUser(username, password string) *UserDetailsExtendedBuilder {
	return &UserDetailsExtendedBuilder{
		data: &UserDetailsExtended{
			Password: password,
			UserDetails: &UserDetails{
				Username: username,
				Emails:   []string{},
				Groups:   []string{},
			},
		},
	}
}

func (b *UserDetailsExtendedBuilder) WithDisplayName(name string) *UserDetailsExtendedBuilder {
	b.data.DisplayName = name
	return b
}

func (b *UserDetailsExtendedBuilder) WithEmail(email string) *UserDetailsExtendedBuilder {
	b.data.Emails = []string{email}
	return b
}

func (b *UserDetailsExtendedBuilder) WithEmails(emails []string) *UserDetailsExtendedBuilder {
	b.data.Emails = emails
	return b
}

func (b *UserDetailsExtendedBuilder) WithGroups(groups []string) *UserDetailsExtendedBuilder {
	b.data.Groups = groups
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

func (b *UserDetailsExtendedBuilder) WithPhoneExtension(extension string) *UserDetailsExtendedBuilder {
	b.data.PhoneExtension = extension
	return b
}

func (b *UserDetailsExtendedBuilder) WithZoneInfo(zoneInfo string) *UserDetailsExtendedBuilder {
	b.data.ZoneInfo = zoneInfo
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

func (b *UserDetailsExtendedBuilder) Build() *UserDetailsExtended {
	return b.data
}

func stringURL(uri *url.URL) string {
	if uri == nil {
		return ""
	}

	return uri.String()
}

// UserDetailsAddress is a structure with a users address information.
type UserDetailsAddress struct {
	StreetAddress string `json:"street_address,omitempty"`
	Locality      string `json:"locality,omitempty"`
	Region        string `json:"region,omitempty"`
	PostalCode    string `json:"postal_code,omitempty"`
	Country       string `json:"country,omitempty"`
}

type AttributeType string

const (
	Text      AttributeType = "text"
	Number    AttributeType = "number"
	Email     AttributeType = "email"
	Password  AttributeType = "password"
	Telephone AttributeType = "tel"
	Url       AttributeType = "url"
	Date      AttributeType = "date"
	Checkbox  AttributeType = "checkbox"
	Groups    AttributeType = "groups"
)

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

// LDAPDiscovery represents various information about a server, such as LDAP Version, Features, Extensions, Controls.
// and SASL Mechanisms.
type LDAPDiscovery struct {
	Successful bool

	LDAPVersion    []int
	SASLMechanisms []string

	Extensions LDAPDiscoveryExtensions
	Controls   LDAPDiscoveryControls
	Features   LDAPDiscoveryFeatures
	Vendor     LDAPDiscoveryVendor
}

func (d LDAPDiscovery) Strings() (extensions, controls, features, saslMechanisms string) {
	if !d.Successful {
		return none, none, none, none
	}

	extensions = d.Extensions.String()
	controls = d.Controls.String()
	features = d.Features.String()

	if len(d.SASLMechanisms) == 0 {
		return extensions, controls, features, none
	}

	saslMechanisms = strings.Join(d.SASLMechanisms, ", ")

	return extensions, controls, features, saslMechanisms
}

// LDAPDiscoveryExtensions represents the extended operations a server supports.
type LDAPDiscoveryExtensions struct {
	OIDs []string

	TLS       bool
	PwdModify bool
	WhoAmI    bool
}

func (s LDAPDiscoveryExtensions) String() string {
	if len(s.OIDs) == 0 {
		return none
	}

	return strings.Join(s.OIDs, ", ")
}

// LDAPDiscoveryControls represents the request and response controls which a server may support.
type LDAPDiscoveryControls struct {
	OIDs []string

	MsftPwdPolHints           bool
	MsftPwdPolHintsDeprecated bool
}

func (s LDAPDiscoveryControls) String() string {
	if len(s.OIDs) == 0 {
		return none
	}

	return strings.Join(s.OIDs, ", ")
}

// LDAPDiscoveryFeatures represents the features a server supports.
type LDAPDiscoveryFeatures struct {
	OIDs []string
}

func (s LDAPDiscoveryFeatures) String() string {
	if len(s.OIDs) == 0 {
		return none
	}

	return strings.Join(s.OIDs, ", ")
}

type LDAPDiscoveryVendor struct {
	Name                  string
	Version               string
	ForestFunctionalLevel int
	DomainFunctionalLevel int
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

// LDAPBaseClient is an extended version of the ldap.Client with some additional functions.
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

	Discovery() (features LDAPDiscovery)
}
