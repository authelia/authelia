package authentication

import (
	"crypto/tls"
	"net/mail"
	"time"

	"github.com/go-ldap/ldap/v3"
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

type ldapUserProfile struct {
	DN          string
	Emails      []string
	DisplayName string
	Username    string
	MemberOf    []string
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
