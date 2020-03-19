package authentication

import (
	"crypto/tls"

	"github.com/go-ldap/ldap/v3"
)

// ********************* CONNECTION *********************

// LDAPConnection interface representing a connection to the ldap.
type LDAPConnection interface {
	Bind(username, password string) error
	Close()

	Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error)
	Modify(modifyRequest *ldap.ModifyRequest) error
}

// LDAPConnectionImpl the production implementation of an ldap connection
type LDAPConnectionImpl struct {
	conn *ldap.Conn
}

// NewLDAPConnectionImpl create a new ldap connection
func NewLDAPConnectionImpl(conn *ldap.Conn) *LDAPConnectionImpl {
	return &LDAPConnectionImpl{conn}
}

func (lc *LDAPConnectionImpl) Bind(username, password string) error {
	return lc.conn.Bind(username, password)
}

func (lc *LDAPConnectionImpl) Close() {
	lc.conn.Close()
}

func (lc *LDAPConnectionImpl) Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error) {
	return lc.conn.Search(searchRequest)
}

func (lc *LDAPConnectionImpl) Modify(modifyRequest *ldap.ModifyRequest) error {
	return lc.conn.Modify(modifyRequest)
}

// ********************* FACTORY ***********************

// LDAPConnectionFactory an interface of factory of ldap connections
type LDAPConnectionFactory interface {
	DialTLS(network, addr string, config *tls.Config) (LDAPConnection, error)
	Dial(network, addr string) (LDAPConnection, error)
}

// LDAPConnectionFactoryImpl the production implementation of an ldap connection factory.
type LDAPConnectionFactoryImpl struct{}

// NewLDAPConnectionFactoryImpl create a concrete ldap connection factory
func NewLDAPConnectionFactoryImpl() *LDAPConnectionFactoryImpl {
	return &LDAPConnectionFactoryImpl{}
}

// DialTLS contact ldap server over TLS.
func (lcf *LDAPConnectionFactoryImpl) DialTLS(network, addr string, config *tls.Config) (LDAPConnection, error) {
	conn, err := ldap.DialTLS(network, addr, config)
	if err != nil {
		return nil, err
	}
	return NewLDAPConnectionImpl(conn), nil
}

// Dial contact ldap server over raw tcp.
func (lcf *LDAPConnectionFactoryImpl) Dial(network, addr string) (LDAPConnection, error) {
	conn, err := ldap.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	return NewLDAPConnectionImpl(conn), nil
}
