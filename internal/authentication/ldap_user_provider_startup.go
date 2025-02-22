package authentication

import (
	"fmt"
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

	if p.features.Extensions.TLS && !p.config.StartTLS && !p.config.Address.IsExplicitlySecure() {
		p.log.Error("Your LDAP Server supports TLS but you don't appear to be utilizing it. We strongly " +
			"recommend using the scheme 'ldaps://' or enabling the StartTLS option to secure connections with your " +
			"LDAP Server.")
	}

	return nil
}

func (p *LDAPUserProvider) getServerSupportedFeatures(client LDAPClient) (features LDAPSupportedFeatures, err error) {
	var (
		request *ldap.SearchRequest
		result  *ldap.SearchResult
	)

	request = ldap.NewSearchRequest("", ldap.ScopeBaseObject, ldap.NeverDerefAliases,
		1, 0, false, ldapBaseObjectFilter, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute}, nil)

	if result, err = client.Search(request); err != nil {
		if p.config.PermitFeatureDetectionFailure {
			p.log.WithError(err).Warnf("Error occurred during RootDSE search. This may result in reduced functionality.")

			return features, nil
		}

		return features, fmt.Errorf("error occurred during RootDSE search: %w", err)
	}

	if len(result.Entries) != 1 {
		p.log.Errorf("The LDAP Server did not respond appropriately to a RootDSE search. This may result in reduced functionality.")

		return features, nil
	}

	var controlTypeOIDs, extensionOIDs []string

	controlTypeOIDs, extensionOIDs, features = ldapGetFeatureSupportFromEntry(result.Entries[0])

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

//nolint:gocyclo
func (p *LDAPUserProvider) parseDynamicUsersConfiguration() {
	p.config.UsersFilter = strings.ReplaceAll(p.config.UsersFilter, ldapPlaceholderDistinguishedNameAttribute, p.config.Attributes.DistinguishedName)
	p.config.UsersFilter = strings.ReplaceAll(p.config.UsersFilter, ldapPlaceholderUsernameAttribute, p.config.Attributes.Username)
	p.config.UsersFilter = strings.ReplaceAll(p.config.UsersFilter, ldapPlaceholderDisplayNameAttribute, p.config.Attributes.DisplayName)
	p.config.UsersFilter = strings.ReplaceAll(p.config.UsersFilter, ldapPlaceholderMailAttribute, p.config.Attributes.Mail)
	p.config.UsersFilter = strings.ReplaceAll(p.config.UsersFilter, ldapPlaceholderMemberOfAttribute, p.config.Attributes.MemberOf)

	p.log.Tracef("Dynamically generated users filter is %s", p.config.UsersFilter)

	if len(p.config.Attributes.Username) != 0 && !utils.IsStringInSlice(p.config.Attributes.Username, p.usersAttributes) {
		p.usersAttributes = append(p.usersAttributes, p.config.Attributes.Username)
		p.usersAttributesExtended = append(p.usersAttributesExtended, p.config.Attributes.Username)
	}

	if len(p.config.Attributes.Mail) != 0 && !utils.IsStringInSlice(p.config.Attributes.Mail, p.usersAttributes) {
		p.usersAttributes = append(p.usersAttributes, p.config.Attributes.Mail)
		p.usersAttributesExtended = append(p.usersAttributesExtended, p.config.Attributes.Mail)
	}

	if len(p.config.Attributes.DisplayName) != 0 && !utils.IsStringInSlice(p.config.Attributes.DisplayName, p.usersAttributes) {
		p.usersAttributes = append(p.usersAttributes, p.config.Attributes.DisplayName)
		p.usersAttributesExtended = append(p.usersAttributesExtended, p.config.Attributes.DisplayName)
	}

	if len(p.config.Attributes.MemberOf) != 0 {
		if !utils.IsStringInSlice(p.config.Attributes.MemberOf, p.usersAttributes) {
			p.usersAttributes = append(p.usersAttributes, p.config.Attributes.MemberOf)
			p.usersAttributesExtended = append(p.usersAttributesExtended, p.config.Attributes.MemberOf)
		}
	}

	attributesExtended := []string{
		p.config.Attributes.GivenName,
		p.config.Attributes.MiddleName,
		p.config.Attributes.FamilyName,
		p.config.Attributes.Nickname,
		p.config.Attributes.Gender,
		p.config.Attributes.Birthdate,
		p.config.Attributes.Website,
		p.config.Attributes.Profile,
		p.config.Attributes.Picture,
		p.config.Attributes.ZoneInfo,
		p.config.Attributes.Locale,
		p.config.Attributes.PhoneNumber,
		p.config.Attributes.PhoneExtension,
		p.config.Attributes.StreetAddress,
		p.config.Attributes.Locality,
		p.config.Attributes.Region,
		p.config.Attributes.PostalCode,
		p.config.Attributes.Country,
	}

	for _, attribute := range attributesExtended {
		if len(attribute) != 0 && !utils.IsStringInSlice(attribute, p.usersAttributesExtended) {
			p.usersAttributesExtended = append(p.usersAttributesExtended, attribute)
		}
	}

	for attribute := range p.config.Attributes.Extra {
		if !utils.IsStringInSlice(attribute, p.usersAttributesExtended) {
			p.usersAttributesExtended = append(p.usersAttributesExtended, attribute)
		}
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

	if strings.Contains(p.config.UsersFilter, ldapPlaceholderDateTimeGeneralized) {
		p.usersFilterReplacementDateTimeGeneralized = true
	}

	if strings.Contains(p.config.UsersFilter, ldapPlaceholderDateTimeUnixEpoch) {
		p.usersFilterReplacementDateTimeUnixEpoch = true
	}

	if strings.Contains(p.config.UsersFilter, ldapPlaceholderDateTimeMicrosoftNTTimeEpoch) {
		p.usersFilterReplacementDateTimeMicrosoftNTTimeEpoch = true
	}

	p.log.Tracef("Detected user filter replacements that need to be resolved per lookup are: %s=%v",
		ldapPlaceholderInput, p.usersFilterReplacementInput)
}

func (p *LDAPUserProvider) parseDynamicGroupsConfiguration() {
	p.config.GroupsFilter = strings.ReplaceAll(p.config.GroupsFilter, ldapPlaceholderDistinguishedNameAttribute, p.config.Attributes.DistinguishedName)
	p.config.GroupsFilter = strings.ReplaceAll(p.config.GroupsFilter, ldapPlaceholderUsernameAttribute, p.config.Attributes.Username)
	p.config.GroupsFilter = strings.ReplaceAll(p.config.GroupsFilter, ldapPlaceholderDisplayNameAttribute, p.config.Attributes.DisplayName)
	p.config.GroupsFilter = strings.ReplaceAll(p.config.GroupsFilter, ldapPlaceholderMailAttribute, p.config.Attributes.Mail)
	p.config.GroupsFilter = strings.ReplaceAll(p.config.GroupsFilter, ldapPlaceholderMemberOfAttribute, p.config.Attributes.MemberOf)

	if len(p.config.Attributes.GroupName) != 0 && !utils.IsStringInSlice(p.config.Attributes.GroupName, p.groupsAttributes) {
		p.groupsAttributes = append(p.groupsAttributes, p.config.Attributes.GroupName)
	}

	if p.config.AdditionalGroupsDN != "" {
		p.groupsBaseDN = p.config.AdditionalGroupsDN + "," + p.config.BaseDN
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

	if strings.Contains(p.config.GroupsFilter, ldapPlaceholderMemberOfDistinguishedName) {
		p.groupsFilterReplacementsMemberOfDN = true
	}

	if strings.Contains(p.config.GroupsFilter, ldapPlaceholderMemberOfRelativeDistinguishedName) {
		p.groupsFilterReplacementsMemberOfRDN = true
	}

	p.log.Tracef("Detected group filter replacements that need to be resolved per lookup are: input=%v, username=%v, dn=%v", p.groupsFilterReplacementInput, p.groupsFilterReplacementUsername, p.groupsFilterReplacementDN)
}
