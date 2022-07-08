package authentication

import (
	"strings"

	"github.com/go-ldap/ldap/v3"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// StartupCheck implements the startup check provider interface.
func (p *LDAPUserProvider) StartupCheck() (err error) {
	var client LDAPClient

	if client, err = p.connect(); err != nil {
		return err
	}

	defer client.Close()

	if p.features, err = p.getServerSupportedFeatures(client); err != nil {
		return err
	}

	if !p.features.Extensions.PwdModifyExOp && !p.disableResetPassword &&
		p.config.Implementation != schema.LDAPImplementationActiveDirectory {
		p.log.Warn("Your LDAP server implementation may not support a method for password hashing " +
			"known to Authelia, it's strongly recommended you ensure your directory server hashes the password " +
			"attribute when users reset their password via Authelia.")
	}

	if p.features.Extensions.TLS && !p.config.StartTLS && !strings.HasPrefix(p.config.URL, "ldaps://") {
		p.log.Error("Your LDAP Server supports TLS but you don't appear to be utilizing it. We strongly " +
			"recommend using the scheme 'ldaps://' or enabling the StartTLS option to secure connections with your " +
			"LDAP Server.")
	}

	if !p.features.Extensions.TLS && p.config.StartTLS {
		p.log.Info("Your LDAP Server does not appear to support TLS but you enabled StartTLS which may result" +
			"in an error.")
	}

	return nil
}

func (p *LDAPUserProvider) getServerSupportedFeatures(client LDAPClient) (features LDAPSupportedFeatures, err error) {
	var (
		searchRequest *ldap.SearchRequest
		searchResult  *ldap.SearchResult
	)

	searchRequest = ldap.NewSearchRequest("", ldap.ScopeBaseObject, ldap.NeverDerefAliases,
		1, 0, false, "(objectClass=*)", []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute}, nil)

	if searchResult, err = client.Search(searchRequest); err != nil {
		return features, err
	}

	if len(searchResult.Entries) != 1 {
		p.log.Errorf("The LDAP Server did not respond appropriately to a RootDSE search. This may result in reduced functionality.")

		return features, nil
	}

	var controlTypeOIDs, extensionOIDs []string

	controlTypeOIDs, extensionOIDs, features = ldapGetFeatureSupportFromEntry(searchResult.Entries[0])

	controlTypes, extensions := none, none

	if len(controlTypeOIDs) != 0 {
		controlTypes = strings.Join(controlTypeOIDs, ", ")
	}

	if len(extensionOIDs) != 0 {
		extensions = strings.Join(extensionOIDs, ", ")
	}

	p.log.Debugf("LDAP Supported OIDs. Control Types: %s. Extensions: %s", controlTypes, extensions)

	return features, nil
}

func (p *LDAPUserProvider) parseDynamicUsersConfiguration() {
	p.config.UsersFilter = strings.ReplaceAll(p.config.UsersFilter, "{username_attribute}", p.config.UsernameAttribute)
	p.config.UsersFilter = strings.ReplaceAll(p.config.UsersFilter, "{mail_attribute}", p.config.MailAttribute)
	p.config.UsersFilter = strings.ReplaceAll(p.config.UsersFilter, "{display_name_attribute}", p.config.DisplayNameAttribute)

	p.log.Tracef("Dynamically generated users filter is %s", p.config.UsersFilter)

	if !utils.IsStringInSlice(p.config.UsernameAttribute, p.usersAttributes) {
		p.usersAttributes = append(p.usersAttributes, p.config.UsernameAttribute)
	}

	if !utils.IsStringInSlice(p.config.MailAttribute, p.usersAttributes) {
		p.usersAttributes = append(p.usersAttributes, p.config.MailAttribute)
	}

	if !utils.IsStringInSlice(p.config.DisplayNameAttribute, p.usersAttributes) {
		p.usersAttributes = append(p.usersAttributes, p.config.DisplayNameAttribute)
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
