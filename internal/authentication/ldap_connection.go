package authentication

import (
	"crypto/tls"

	"github.com/go-ldap/ldap/v3"
)

// LDAPConnection interface representing a connection to the ldap.
type LDAPConnection interface {
	Bind(username, password string) error
	Close()

	Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error)
	Modify(modifyRequest *ldap.ModifyRequest) error
	PasswordModify(pwdModifyRequest *ldap.PasswordModifyRequest) (*ldap.PasswordModifyResult, error)
	StartTLS(config *tls.Config) error
}

// LDAPConnectionImpl the production implementation of an ldap connection.
type LDAPConnectionImpl struct {
	conn *ldap.Conn
}

// NewLDAPConnectionImpl create a new ldap connection.
func NewLDAPConnectionImpl(conn *ldap.Conn) *LDAPConnectionImpl {
	return &LDAPConnectionImpl{conn}
}

// Bind binds ldap connection to a username/password.
func (lc *LDAPConnectionImpl) Bind(username, password string) error {
	return lc.conn.Bind(username, password)
}

// Close closes a ldap connection.
func (lc *LDAPConnectionImpl) Close() {
	lc.conn.Close()
}

// Search searches a ldap server.
func (lc *LDAPConnectionImpl) Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error) {
	return lc.conn.Search(searchRequest)
}

// Modify modifies an ldap object.
func (lc *LDAPConnectionImpl) Modify(modifyRequest *ldap.ModifyRequest) error {
	return lc.conn.Modify(modifyRequest)
}

// PasswordModify modifies an ldap objects password.
func (lc *LDAPConnectionImpl) PasswordModify(pwdModifyRequest *ldap.PasswordModifyRequest) error {
	_, err := lc.conn.PasswordModify(pwdModifyRequest)
	return err
}

// StartTLS requests the LDAP server upgrades to TLS encryption.
func (lc *LDAPConnectionImpl) StartTLS(config *tls.Config) error {
	return lc.conn.StartTLS(config)
}
