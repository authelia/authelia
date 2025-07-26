package authentication

import (
	"github.com/go-ldap/ldap/v3"
)

func (p *LDAPUserProvider) validateADUserData(userData *UserDetailsExtended) error {
	return nil
}

func (p *LDAPUserProvider) createADAddRequest(userData *UserDetailsExtended) (*ldap.AddRequest, error) {
	return nil, nil
}

func (p *LDAPUserProvider) createADDeleteRequest(username string, userDN string) (*ldap.DelRequest, error) {
	return nil, nil
}

func (p *LDAPUserProvider) createADModifyRequest(username string, userData *UserDetailsExtended) (*ldap.ModifyRequest, error) {
	return nil, nil
}

func (p *LDAPUserProvider) getADRequiredFields() []string {
	return nil
}

func (p *LDAPUserProvider) getRADSupportedFields() []string {
	return nil
}

func (p *LDAPUserProvider) getADDefaultObjectClasses() []string {
	return nil
}

func (p *LDAPUserProvider) getADFieldMetadata() map[string]FieldMetadata {
	return nil
}
