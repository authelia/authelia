package authentication

import (
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// StartupCheck implements the startup check provider interface.
func (p *LDAPUserProvider) StartupCheck(logger *logrus.Logger) (err error) {
	conn, err := p.connect(p.configuration.User, p.configuration.Password)
	if err != nil {
		return err
	}

	defer conn.Close()

	searchRequest := ldap.NewSearchRequest("", ldap.ScopeBaseObject, ldap.NeverDerefAliases,
		1, 0, false, "(objectClass=*)", []string{ldapSupportedExtensionAttribute}, nil)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return err
	}

	if len(sr.Entries) != 1 {
		return nil
	}

	// Iterate the attribute values to see what the server supports.
	for _, attr := range sr.Entries[0].Attributes {
		if attr.Name == ldapSupportedExtensionAttribute {
			logger.Tracef("LDAP Supported Extension OIDs: %s", strings.Join(attr.Values, ", "))

			for _, oid := range attr.Values {
				if oid == ldapOIDPasswdModifyExtension {
					p.supportExtensionPasswdModify = true
					break
				}
			}

			break
		}
	}

	if !p.supportExtensionPasswdModify && !p.disableResetPassword &&
		p.configuration.Implementation != schema.LDAPImplementationActiveDirectory {
		logger.Warn("Your LDAP server implementation may not support a method for password hashing " +
			"known to Authelia, it's strongly recommended you ensure your directory server hashes the password " +
			"attribute when users reset their password via Authelia.")
	}

	return nil
}

func (p *LDAPUserProvider) parseDynamicUsersConfiguration() {
	p.configuration.UsersFilter = strings.ReplaceAll(p.configuration.UsersFilter, "{username_attribute}", p.configuration.UsernameAttribute)
	p.configuration.UsersFilter = strings.ReplaceAll(p.configuration.UsersFilter, "{mail_attribute}", p.configuration.MailAttribute)
	p.configuration.UsersFilter = strings.ReplaceAll(p.configuration.UsersFilter, "{display_name_attribute}", p.configuration.DisplayNameAttribute)

	p.logger.Tracef("Dynamically generated users filter is %s", p.configuration.UsersFilter)

	p.usersAttributes = []string{
		p.configuration.DisplayNameAttribute,
		p.configuration.MailAttribute,
		p.configuration.UsernameAttribute,
	}

	if p.configuration.AdditionalUsersDN != "" {
		p.usersBaseDN = p.configuration.AdditionalUsersDN + "," + p.configuration.BaseDN
	} else {
		p.usersBaseDN = p.configuration.BaseDN
	}

	p.logger.Tracef("Dynamically generated users BaseDN is %s", p.usersBaseDN)

	if strings.Contains(p.configuration.UsersFilter, ldapPlaceholderInput) {
		p.usersFilterReplacementInput = true
	}

	p.logger.Tracef("Detected user filter replacements that need to be resolved per lookup are: %s=%v",
		ldapPlaceholderInput, p.usersFilterReplacementInput)
}

func (p *LDAPUserProvider) parseDynamicGroupsConfiguration() {
	p.groupsAttributes = []string{
		p.configuration.GroupNameAttribute,
	}

	if p.configuration.AdditionalGroupsDN != "" {
		p.groupsBaseDN = ldap.EscapeFilter(p.configuration.AdditionalGroupsDN + "," + p.configuration.BaseDN)
	} else {
		p.groupsBaseDN = p.configuration.BaseDN
	}

	p.logger.Tracef("Dynamically generated groups BaseDN is %s", p.groupsBaseDN)

	if strings.Contains(p.configuration.GroupsFilter, ldapPlaceholderInput) {
		p.groupsFilterReplacementInput = true
	}

	if strings.Contains(p.configuration.GroupsFilter, ldapPlaceholderUsername) {
		p.groupsFilterReplacementUsername = true
	}

	if strings.Contains(p.configuration.GroupsFilter, ldapPlaceholderDistinguishedName) {
		p.groupsFilterReplacementDN = true
	}

	p.logger.Tracef("Detected group filter replacements that need to be resolved per lookup are: input=%v, username=%v, dn=%v", p.groupsFilterReplacementInput, p.groupsFilterReplacementUsername, p.groupsFilterReplacementDN)
}
