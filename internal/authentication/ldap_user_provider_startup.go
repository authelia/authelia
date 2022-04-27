package authentication

import (
	"strings"

	"github.com/go-ldap/ldap/v3"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// StartupCheck implements the startup check provider interface.
func (p *LDAPUserProvider) StartupCheck() (err error) {
	var (
		conn         LDAPConnection
		searchResult *ldap.SearchResult
	)

	if conn, err = p.connect(); err != nil {
		return err
	}

	defer conn.Close()

	searchRequest := ldap.NewSearchRequest("", ldap.ScopeBaseObject, ldap.NeverDerefAliases,
		1, 0, false, "(objectClass=*)", []string{ldapSupportedExtensionAttribute}, nil)

	if searchResult, err = conn.Search(searchRequest); err != nil {
		return err
	}

	if len(searchResult.Entries) != 1 {
		return nil
	}

	// Iterate the attribute values to see what the server supports.
	for _, attr := range searchResult.Entries[0].Attributes {
		if attr.Name == ldapSupportedExtensionAttribute {
			p.log.Tracef("LDAP Supported Extension OIDs: %s", strings.Join(attr.Values, ", "))

			for _, oid := range attr.Values {
				if oid == ldapOIDPasswdModifyExtension {
					p.supportExtensionPasswdModify = true
					break
				}
			}
		}
	}

	if !p.supportExtensionPasswdModify && !p.disableResetPassword &&
		p.config.Implementation != schema.LDAPImplementationActiveDirectory {
		p.log.Warn("Your LDAP server implementation may not support a method for password hashing " +
			"known to Authelia, it's strongly recommended you ensure your directory server hashes the password " +
			"attribute when users reset their password via Authelia.")
	}

	return nil
}

func (p *LDAPUserProvider) parseDynamicUsersConfiguration() {
	p.config.UsersFilter = strings.ReplaceAll(p.config.UsersFilter, "{username_attribute}", p.config.UsernameAttribute)
	p.config.UsersFilter = strings.ReplaceAll(p.config.UsersFilter, "{mail_attribute}", p.config.MailAttribute)
	p.config.UsersFilter = strings.ReplaceAll(p.config.UsersFilter, "{display_name_attribute}", p.config.DisplayNameAttribute)

	p.log.Tracef("Dynamically generated users filter is %s", p.config.UsersFilter)

	p.usersAttributes = []string{
		p.config.DisplayNameAttribute,
		p.config.MailAttribute,
		p.config.UsernameAttribute,
	}

	if p.config.AdditionalUsersDN != "" {
		p.usersBaseDN = p.config.AdditionalUsersDN + "," + p.config.BaseDN
	} else {
		p.usersBaseDN = p.config.BaseDN
	}

	p.log.Tracef("Dynamically generated users BaseDN is %s", p.usersBaseDN)

	if strings.Contains(p.config.UsersFilter, ldapPlaceholderInput) {
		p.usersFilterReplacementInput = true
	}

	p.log.Tracef("Detected user filter replacements that need to be resolved per lookup are: %s=%v",
		ldapPlaceholderInput, p.usersFilterReplacementInput)
}

func (p *LDAPUserProvider) parseDynamicGroupsConfiguration() {
	p.groupsAttributes = []string{
		p.config.GroupNameAttribute,
	}

	if p.config.AdditionalGroupsDN != "" {
		p.groupsBaseDN = ldap.EscapeFilter(p.config.AdditionalGroupsDN + "," + p.config.BaseDN)
	} else {
		p.groupsBaseDN = p.config.BaseDN
	}

	p.log.Tracef("Dynamically generated groups BaseDN is %s", p.groupsBaseDN)

	if strings.Contains(p.config.GroupsFilter, ldapPlaceholderInput) {
		p.groupsFilterReplacementInput = true
	}

	if strings.Contains(p.config.GroupsFilter, ldapPlaceholderUsername) {
		p.groupsFilterReplacementUsername = true
	}

	if strings.Contains(p.config.GroupsFilter, ldapPlaceholderDistinguishedName) {
		p.groupsFilterReplacementDN = true
	}

	p.log.Tracef("Detected group filter replacements that need to be resolved per lookup are: input=%v, username=%v, dn=%v", p.groupsFilterReplacementInput, p.groupsFilterReplacementUsername, p.groupsFilterReplacementDN)
}
