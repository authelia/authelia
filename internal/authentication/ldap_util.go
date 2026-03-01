package authentication

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-ldap/ldap/v3"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

func ldapEntriesContainsEntry(needle *ldap.Entry, haystack []*ldap.Entry) bool {
	if needle == nil || len(haystack) == 0 {
		return false
	}

	for i := 0; i < len(haystack); i++ {
		if haystack[i].DN == needle.DN {
			return true
		}
	}

	return false
}

func ldapGetFeatureSupportFromClient(client LDAPBaseClient) (discovery LDAPDiscovery, err error) {
	var (
		request *ldap.SearchRequest
		result  *ldap.SearchResult
	)

	request = ldapNewSearchRequestRootDSE()

	if result, err = client.Search(request); err != nil {
		return discovery, fmt.Errorf("error occurred during RootDSE search: %w", err)
	}

	if len(result.Entries) != 1 {
		return discovery, fmt.Errorf("error occurred during RootDSE search: %w", ErrLDAPHealthCheckFailedEntryCount)
	}

	return ldapGetDiscoveryFromLDAPEntry(result.Entries[0]), nil
}

func ldapGetDiscoveryFromLDAPEntry(entry *ldap.Entry) (discovery LDAPDiscovery) {
	if entry == nil {
		return
	}

	discovery.Successful = true

	var fallbackVendorName string

	for _, attr := range entry.Attributes {
		switch attr.Name {
		case ldapObjectClassAttribute:
			if utils.IsStringInSliceFold(ldapVendorOpenLDAPObjectClass, attr.Values) {
				fallbackVendorName = ldapVendorNameOpenLDAP
			}
		case ldapSupportedLDAPVersionAttribute:
			discovery.LDAPVersion = ldapGetLDAPVersionDiscoveryFromLDAPEntry(attr)
		case ldapSupportedControlAttribute:
			ldapGetControlsDiscoveryFromLDAPEntry(attr, &discovery.Controls)
		case ldapSupportedExtensionAttribute:
			ldapGetExtensionDiscoveryFromLDAPEntry(attr, &discovery.Extensions)
		case ldapSupportedSASLMechanismsAttribute:
			discovery.SASLMechanisms = attr.Values
		case ldapSupportedFeaturesAttribute:
			discovery.Features.OIDs = attr.Values
		case ldapVendorNameAttribute:
			discovery.Vendor.Name = strings.Join(attr.Values, " ")
		case ldapVendorVersionAttribute:
			discovery.Vendor.Version = strings.Join(attr.Values, " ")
		case ldapDomainFunctionalityAttribute:
			if len(attr.Values) != 1 {
				continue
			}

			fallbackVendorName = ldapVendorNameMicrosoftCorporation
			discovery.Vendor.DomainFunctionalLevel, _ = strconv.Atoi(attr.Values[0])
		case ldapForestFunctionalityAttribute:
			if len(attr.Values) != 1 {
				continue
			}

			fallbackVendorName = ldapVendorNameMicrosoftCorporation
			discovery.Vendor.ForestFunctionalLevel, _ = strconv.Atoi(attr.Values[0])
		}
	}

	if discovery.Vendor.Name == "" {
		discovery.Vendor.Name = fallbackVendorName
	}

	return discovery
}

func ldapGetLDAPVersionDiscoveryFromLDAPEntry(attr *ldap.EntryAttribute) (versions []int) {
	versions = make([]int, 0, len(attr.Values))

	for _, v := range attr.Values {
		version, err := strconv.Atoi(v)
		if err != nil {
			break
		}

		versions = append(versions, version)
	}

	if len(versions) != len(attr.Values) {
		return nil
	}

	return versions
}

func ldapGetControlsDiscoveryFromLDAPEntry(attr *ldap.EntryAttribute, controls *LDAPDiscoveryControls) {
	controls.OIDs = attr.Values

	for _, oid := range attr.Values {
		switch oid {
		case ldapOIDControlMsftServerPolicyHints:
			controls.MsftPwdPolHints = true
		case ldapOIDControlMsftServerPolicyHintsDeprecated:
			controls.MsftPwdPolHintsDeprecated = true
		}
	}
}

func ldapGetExtensionDiscoveryFromLDAPEntry(attr *ldap.EntryAttribute, extensions *LDAPDiscoveryExtensions) {
	extensions.OIDs = attr.Values

	for _, oid := range attr.Values {
		switch oid {
		case ldapOIDExtensionPwdModify:
			extensions.PwdModify = true
		case ldapOIDExtensionTLS:
			extensions.TLS = true
		case ldapOIDExtensionWhoAmI:
			extensions.WhoAmI = true
		}
	}
}

func ldapEscape(inputUsername string) string {
	inputUsername = ldap.EscapeFilter(inputUsername)
	for _, c := range specialLDAPRunes {
		inputUsername = strings.ReplaceAll(inputUsername, string(c), fmt.Sprintf("\\%c", c))
	}

	return inputUsername
}

func getLDAPResultCode(err error) int {
	var e *ldap.Error
	if errors.As(err, &e) {
		return int(e.ResultCode)
	}

	return -1
}

func getValueFromEntry(entry *ldap.Entry, attribute string) string {
	if attribute == "" {
		return ""
	}

	return entry.GetAttributeValue(attribute)
}

func getValuesFromEntry(entry *ldap.Entry, attribute string) []string {
	if attribute == "" {
		return nil
	}

	return entry.GetAttributeValues(attribute)
}

func getExtraValueFromEntry(entry *ldap.Entry, attribute string, properties schema.AuthenticationBackendLDAPAttributesAttribute) (value any, err error) {
	if properties.MultiValued {
		return getExtraValueMultiFromEntry(entry, attribute, properties)
	}

	str := getValueFromEntry(entry, attribute)

	switch properties.ValueType {
	case ValueTypeString:
		value = str
	case ValueTypeInteger:
		if str == "" {
			return nil, nil
		}

		if value, err = strconv.ParseFloat(str, 64); err != nil {
			return nil, fmt.Errorf("cannot parse '%s' with value '%s' as integer: %w", attribute, str, err)
		}
	case ValueTypeBoolean:
		if str == "" {
			return nil, nil
		}

		if value, err = strconv.ParseBool(str); err != nil {
			return nil, fmt.Errorf("cannot parse '%s' with value '%s' as boolean: %w", attribute, str, err)
		}
	}

	return value, nil
}

func getExtraValueMultiFromEntry(entry *ldap.Entry, attribute string, properties schema.AuthenticationBackendLDAPAttributesAttribute) (value any, err error) {
	if entry == nil {
		return nil, fmt.Errorf("failed to get values from nil entry for attribute '%s'", attribute)
	}

	strs := getValuesFromEntry(entry, attribute)

	values := make([]any, len(strs))

	switch properties.GetValueType() {
	case ValueTypeString:
		for i, v := range strs {
			values[i] = v
		}
	case ValueTypeInteger:
		var v float64

		for i, str := range strs {
			if v, err = strconv.ParseFloat(str, 64); err != nil {
				return nil, fmt.Errorf("cannot parse '%s' with value '%s' as integer: %w", attribute, str, err)
			}

			values[i] = v
		}
	case ValueTypeBoolean:
		var v bool

		for i, str := range strs {
			if v, err = strconv.ParseBool(str); err != nil {
				return nil, fmt.Errorf("cannot parse '%s' with value '%s' as boolean: %w", attribute, str, err)
			}

			values[i] = v
		}
	}

	return values, nil
}

func ldapNewSearchRequestRootDSE() *ldap.SearchRequest {
	return ldap.NewSearchRequest("", ldap.ScopeBaseObject, ldap.NeverDerefAliases,
		1, 0, false, ldapBaseObjectFilter, []string{ldapObjectClassAttribute, ldapSupportedLDAPVersionAttribute, ldapSupportedExtensionAttribute, ldapSupportedControlAttribute, ldapSupportedFeaturesAttribute, ldapSupportedSASLMechanismsAttribute, ldapVendorNameAttribute, ldapVendorVersionAttribute, ldapDomainFunctionalityAttribute, ldapForestFunctionalityAttribute}, nil)
}
