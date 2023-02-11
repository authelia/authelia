package authentication

import (
	"crypto/tls"
	"net/mail"
	"time"

	"github.com/go-ldap/ldap/v3"
	"golang.org/x/text/encoding/unicode"
)

// LDAPClientFactory an interface of factory of LDAP clients.
type LDAPClientFactory interface {
	DialURL(addr string, opts ...ldap.DialOpt) (client LDAPClient, err error)
}

// LDAPClient is a cut down version of the ldap.Client interface with just the methods we use.
//
// Methods added to this interface that have a direct correlation with one from ldap.Client should have the same signature.
type LDAPClient interface {
	Start()
	Close()
	IsClosing() bool
	SetTimeout(timeout time.Duration)

	TLSConnectionState() (state tls.ConnectionState, ok bool)
	StartTLS(config *tls.Config) (err error)

	Unbind() (err error)
	Bind(username, password string) (err error)
	SimpleBind(simpleBindRequest *ldap.SimpleBindRequest) (bindResult *ldap.SimpleBindResult, err error)
	MD5Bind(host string, username string, password string) (err error)
	DigestMD5Bind(digestMD5BindRequest *ldap.DigestMD5BindRequest) (digestMD5BindResult *ldap.DigestMD5BindResult, err error)
	UnauthenticatedBind(username string) (err error)
	ExternalBind() (err error)

	Modify(modifyRequest *ldap.ModifyRequest) (err error)
	ModifyWithResult(modifyRequest *ldap.ModifyRequest) (modifyResult *ldap.ModifyResult, err error)
	ModifyDN(m *ldap.ModifyDNRequest) (err error)
	PasswordModify(pwdModifyRequest *ldap.PasswordModifyRequest) (pwdModifyResult *ldap.PasswordModifyResult, err error)

	Add(addRequest *ldap.AddRequest) (err error)
	Del(delRequest *ldap.DelRequest) (err error)

	Search(searchRequest *ldap.SearchRequest) (searchResult *ldap.SearchResult, err error)
	SearchWithPaging(searchRequest *ldap.SearchRequest, pagingSize uint32) (searchResult *ldap.SearchResult, err error)
	Compare(dn string, attribute string, value string) (same bool, err error)

	WhoAmI(controls []ldap.Control) (whoamiResult *ldap.WhoAmIResult, err error)
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

type ldapUserProfile struct {
	DN          string
	Emails      []string
	DisplayName string
	Username    string
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

var utf16LittleEndian = unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
