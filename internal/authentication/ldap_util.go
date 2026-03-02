package authentication

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/go-ldap/ldap/v3"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
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

func ldapGetFeatureSupportFromClient(client LDAPBaseClient) (features LDAPSupportedFeatures, err error) {
	var (
		request *ldap.SearchRequest
		result  *ldap.SearchResult
	)

	request = ldapNewSearchRequestRootDSE()

	if result, err = client.Search(request); err != nil {
		return features, fmt.Errorf("error occurred during RootDSE search: %w", err)
	}

	if len(result.Entries) != 1 {
		return features, fmt.Errorf("error occurred during RootDSE search: %w", ErrLDAPHealthCheckFailedEntryCount)
	}

	return ldapGetFeatureSupportFromEntry(result.Entries[0]), nil
}

func ldapGetFeatureSupportFromEntry(entry *ldap.Entry) (features LDAPSupportedFeatures) {
	if entry == nil {
		return
	}

	for _, attr := range entry.Attributes {
		switch attr.Name {
		case ldapSupportedControlAttribute:
			features.ControlTypes.OIDs = attr.Values

			for _, oid := range attr.Values {
				switch oid {
				case ldapOIDControlMsftServerPolicyHints:
					features.ControlTypes.MsftPwdPolHints = true
				case ldapOIDControlMsftServerPolicyHintsDeprecated:
					features.ControlTypes.MsftPwdPolHintsDeprecated = true
				}
			}
		case ldapSupportedExtensionAttribute:
			features.Extensions.OIDs = attr.Values

			for _, oid := range attr.Values {
				switch oid {
				case ldapOIDExtensionPwdModify:
					features.Extensions.PwdModify = true
				case ldapOIDExtensionTLS:
					features.Extensions.TLS = true
				case ldapOIDExtensionWhoAmI:
					features.Extensions.WhoAmI = true
				}
			}
		}
	}

	return features
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
		1, 0, false, ldapBaseObjectFilter, []string{ldapSupportedExtensionAttribute, ldapSupportedControlAttribute}, nil)
}
