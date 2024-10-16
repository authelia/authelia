package authentication

import (
	"crypto/tls"
	"fmt"
	"net/mail"
	"net/url"
	"time"

	"github.com/go-ldap/ldap/v3"
	"golang.org/x/text/language"
)

// LDAPClientFactory an interface of factory of LDAP clients.
type LDAPClientFactory interface {
	DialURL(addr string, opts ...ldap.DialOpt) (client LDAPClient, err error)
}

// LDAPClient is a cut down version of the ldap.Client interface with just the methods we use.
//
// Methods added to this interface that have a direct correlation with one from ldap.Client should have the same signature.
type LDAPClient interface {
	Close() (err error)
	IsClosing() bool
	SetTimeout(timeout time.Duration)

	TLSConnectionState() (state tls.ConnectionState, ok bool)
	StartTLS(config *tls.Config) (err error)

	Unbind() (err error)
	Bind(username, password string) (err error)
	SimpleBind(request *ldap.SimpleBindRequest) (result *ldap.SimpleBindResult, err error)
	MD5Bind(host string, username string, password string) (err error)
	DigestMD5Bind(request *ldap.DigestMD5BindRequest) (result *ldap.DigestMD5BindResult, err error)
	UnauthenticatedBind(username string) (err error)
	ExternalBind() (err error)
	NTLMBind(domain string, username string, password string) (err error)
	NTLMUnauthenticatedBind(domain string, username string) (err error)
	NTLMBindWithHash(domain string, username string, hash string) (err error)
	NTLMChallengeBind(request *ldap.NTLMBindRequest) (result *ldap.NTLMBindResult, err error)

	Modify(request *ldap.ModifyRequest) (err error)
	ModifyWithResult(request *ldap.ModifyRequest) (result *ldap.ModifyResult, err error)
	ModifyDN(m *ldap.ModifyDNRequest) (err error)
	PasswordModify(request *ldap.PasswordModifyRequest) (result *ldap.PasswordModifyResult, err error)

	Add(request *ldap.AddRequest) (err error)
	Del(request *ldap.DelRequest) (err error)

	Search(request *ldap.SearchRequest) (result *ldap.SearchResult, err error)
	SearchWithPaging(request *ldap.SearchRequest, pagingSize uint32) (result *ldap.SearchResult, err error)
	Compare(dn string, attribute string, value string) (same bool, err error)

	WhoAmI(controls []ldap.Control) (result *ldap.WhoAmIResult, err error)
}

// UserDetails represent the details retrieved for a given user.
type UserDetails struct {
	Username    string
	DisplayName string
	Emails      []string
	Groups      []string
}

// Addresses returns the Emails []string as []mail.Address formatted with DisplayName as the Name attribute.
func (d UserDetails) Addresses() (addresses []mail.Address) {
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

func (d UserDetails) GetUsername() (username string) {
	return d.Username
}

func (d UserDetails) GetGroups() (groups []string) {
	return d.Groups
}

func (d UserDetails) GetDisplayName() (name string) {
	return d.DisplayName
}

func (d UserDetails) GetEmails() (emails []string) {
	return d.Emails
}

// UserDetailsExtended represent the extended details retrieved for a given user.
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

func (d UserDetailsExtended) GetGivenName() (given string) {
	return d.GivenName
}

func (d UserDetailsExtended) GetFamilyName() (family string) {
	return d.FamilyName
}

func (d UserDetailsExtended) GetMiddleName() (middle string) {
	return d.MiddleName
}

func (d UserDetailsExtended) GetNickname() (nickname string) {
	return d.Nickname
}

func (d UserDetailsExtended) GetProfile() (profile string) {
	if d.Profile == nil {
		return ""
	}

	return d.Profile.String()
}

func (d UserDetailsExtended) GetPicture() (picture string) {
	if d.Picture == nil {
		return ""
	}

	return d.Picture.String()
}

func (d UserDetailsExtended) GetWebsite() (website string) {
	if d.Website == nil {
		return ""
	}

	return d.Website.String()
}

func (d UserDetailsExtended) GetGender() (gender string) {
	return d.Gender
}

func (d UserDetailsExtended) GetBirthdate() (birthdate string) {
	return d.Birthdate
}

func (d UserDetailsExtended) GetZoneInfo() (info string) {
	return d.ZoneInfo
}

func (d UserDetailsExtended) GetLocale() (locale string) {
	if d.Locale == nil {
		return ""
	}

	return d.Locale.String()
}

func (d UserDetailsExtended) GetPhoneNumber() (number string) {
	return d.PhoneNumber
}

func (d UserDetailsExtended) GetPhoneExtension() (extension string) {
	return d.PhoneExtension
}

func (d UserDetailsExtended) GetPhoneNumberRFC3966() (number string) {
	if d.PhoneNumber == "" {
		return ""
	}

	if d.PhoneExtension == "" {
		return d.PhoneNumber
	}

	return fmt.Sprintf("%s;ext=%s", d.PhoneNumber, d.PhoneExtension)
}

func (d UserDetailsExtended) GetStreetAddress() (address string) {
	if d.Address == nil {
		return ""
	}

	return d.Address.StreetAddress
}

func (d UserDetailsExtended) GetLocality() (locality string) {
	if d.Address == nil {
		return ""
	}

	return d.Address.Locality
}

func (d UserDetailsExtended) GetRegion() (region string) {
	if d.Address == nil {
		return ""
	}

	return d.Address.Region
}

func (d UserDetailsExtended) GetPostalCode() (postcode string) {
	if d.Address == nil {
		return ""
	}

	return d.Address.PostalCode
}

func (d UserDetailsExtended) GetCountry() (country string) {
	if d.Address == nil {
		return ""
	}

	return d.Address.Country
}

func (d UserDetailsExtended) GetExtra() (extra map[string]any) {
	return d.Extra
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
