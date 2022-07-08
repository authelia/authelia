package authentication

import (
	"crypto/tls"

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
	Close()
	StartTLS(config *tls.Config) (err error)

	Bind(username, password string) (err error)
	UnauthenticatedBind(username string) (err error)

	Modify(modifyRequest *ldap.ModifyRequest) (err error)
	PasswordModify(pwdModifyRequest *ldap.PasswordModifyRequest) (pwdModifyResult *ldap.PasswordModifyResult, err error)

	Search(searchRequest *ldap.SearchRequest) (searchResult *ldap.SearchResult, err error)
}

// UserDetails represent the details retrieved for a given user.
type UserDetails struct {
	Username    string
	DisplayName string
	Emails      []string
	Groups      []string
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
