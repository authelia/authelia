package authentication

import (
	"testing"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
)

func TestLDAPGetDiscoveryFromLDAPEntryNil(t *testing.T) {
	discovery := ldapGetDiscoveryFromLDAPEntry(nil)
	assert.Len(t, discovery.Extensions.OIDs, 0)
	assert.Len(t, discovery.Controls.OIDs, 0)
	assert.Equal(t, LDAPDiscovery{}, discovery)
}

func TestLDAPGetFeatureSupportFromEntry(t *testing.T) {
	testCases := []struct {
		description               string
		haveObjectClass           []string
		haveLDAPVersion           []string
		haveControlOIDs           []string
		haveExtensionOIDs         []string
		haveFeatureOIDs           []string
		haveSASLMechanisms        []string
		haveVendorName            []string
		haveVendorVersion         []string
		haveDomainFunctionalLevel []string
		haveForestFunctionalLevel []string
		expected                  LDAPDiscovery
	}{
		{
			description:       "ShouldReturnExtensionPwdModifyExOp",
			haveControlOIDs:   []string{},
			haveExtensionOIDs: []string{ldapOIDExtensionPwdModify},
			expected:          LDAPDiscovery{Successful: true, Extensions: LDAPDiscoveryExtensions{PwdModify: true, OIDs: []string{ldapOIDExtensionPwdModify}}, Controls: LDAPDiscoveryControls{OIDs: []string{}}},
		},
		{
			description:       "ShouldHandleOpenLDAP",
			haveObjectClass:   []string{"top", ldapVendorOpenLDAPObjectClass},
			haveControlOIDs:   []string{},
			haveExtensionOIDs: []string{ldapOIDExtensionPwdModify},
			expected:          LDAPDiscovery{Successful: true, Extensions: LDAPDiscoveryExtensions{PwdModify: true, OIDs: []string{ldapOIDExtensionPwdModify}}, Controls: LDAPDiscoveryControls{OIDs: []string{}}, Vendor: LDAPDiscoveryVendor{Name: "OpenLDAP"}},
		},
		{
			description:       "ShouldReturnExtensionTLS",
			haveControlOIDs:   []string{},
			haveExtensionOIDs: []string{ldapOIDExtensionTLS},
			expected:          LDAPDiscovery{Successful: true, Extensions: LDAPDiscoveryExtensions{TLS: true, OIDs: []string{ldapOIDExtensionTLS}}, Controls: LDAPDiscoveryControls{OIDs: []string{}}},
		},
		{
			description:       "ShouldReturnExtensionAll",
			haveControlOIDs:   []string{},
			haveExtensionOIDs: []string{ldapOIDExtensionTLS, ldapOIDExtensionPwdModify},
			expected:          LDAPDiscovery{Successful: true, Extensions: LDAPDiscoveryExtensions{TLS: true, PwdModify: true, OIDs: []string{ldapOIDExtensionTLS, ldapOIDExtensionPwdModify}}, Controls: LDAPDiscoveryControls{OIDs: []string{}}},
		},
		{
			description:       "ShouldReturnControlMsftPPolHints",
			haveControlOIDs:   []string{ldapOIDControlMsftServerPolicyHints},
			haveExtensionOIDs: []string{},
			expected:          LDAPDiscovery{Successful: true, Extensions: LDAPDiscoveryExtensions{OIDs: []string{}}, Controls: LDAPDiscoveryControls{MsftPwdPolHints: true, OIDs: []string{ldapOIDControlMsftServerPolicyHints}}},
		},
		{
			description:       "ShouldReturnControlMsftPPolHintsDeprecated",
			haveControlOIDs:   []string{ldapOIDControlMsftServerPolicyHintsDeprecated},
			haveExtensionOIDs: []string{},
			expected:          LDAPDiscovery{Successful: true, Extensions: LDAPDiscoveryExtensions{OIDs: []string{}}, Controls: LDAPDiscoveryControls{MsftPwdPolHintsDeprecated: true, OIDs: []string{ldapOIDControlMsftServerPolicyHintsDeprecated}}},
		},
		{
			description:       "ShouldReturnControlAll",
			haveControlOIDs:   []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated},
			haveExtensionOIDs: []string{},
			expected:          LDAPDiscovery{Successful: true, Extensions: LDAPDiscoveryExtensions{OIDs: []string{}}, Controls: LDAPDiscoveryControls{MsftPwdPolHints: true, MsftPwdPolHintsDeprecated: true, OIDs: []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated}}},
		},
		{
			description:       "ShouldReturnExtensionAndControlAll",
			haveControlOIDs:   []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated},
			haveExtensionOIDs: []string{ldapOIDExtensionTLS, ldapOIDExtensionPwdModify},
			expected: LDAPDiscovery{
				Successful: true,
				Controls:   LDAPDiscoveryControls{MsftPwdPolHints: true, MsftPwdPolHintsDeprecated: true, OIDs: []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated}},
				Extensions: LDAPDiscoveryExtensions{TLS: true, PwdModify: true, OIDs: []string{ldapOIDExtensionTLS, ldapOIDExtensionPwdModify}},
			},
		},
		{
			description:        "ShouldReturnAllDiscovery",
			haveLDAPVersion:    []string{"2", "3"},
			haveControlOIDs:    []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated},
			haveExtensionOIDs:  []string{ldapOIDExtensionTLS, ldapOIDExtensionPwdModify},
			haveFeatureOIDs:    []string{"example"},
			haveSASLMechanisms: []string{"SCRAM"},
			haveVendorName:     []string{"Authelia"},
			haveVendorVersion:  []string{"Authelia LDAP v0.1.0"},
			expected: LDAPDiscovery{
				Successful:     true,
				LDAPVersion:    []int{2, 3},
				Controls:       LDAPDiscoveryControls{MsftPwdPolHints: true, MsftPwdPolHintsDeprecated: true, OIDs: []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated}},
				Extensions:     LDAPDiscoveryExtensions{TLS: true, PwdModify: true, OIDs: []string{ldapOIDExtensionTLS, ldapOIDExtensionPwdModify}},
				Features:       LDAPDiscoveryFeatures{OIDs: []string{"example"}},
				SASLMechanisms: []string{"SCRAM"},
				Vendor: LDAPDiscoveryVendor{
					Name:    "Authelia",
					Version: "Authelia LDAP v0.1.0",
				},
			},
		},
		{
			description:               "ShouldReturnMicrosoftAD",
			haveControlOIDs:           []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated},
			haveExtensionOIDs:         []string{ldapOIDExtensionPwdModify},
			haveDomainFunctionalLevel: []string{"4"},
			haveForestFunctionalLevel: []string{"5"},
			expected: LDAPDiscovery{
				Successful: true,
				Extensions: LDAPDiscoveryExtensions{
					OIDs:      []string{ldapOIDExtensionPwdModify},
					PwdModify: true,
				},
				Controls: LDAPDiscoveryControls{
					OIDs:                      []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated},
					MsftPwdPolHints:           true,
					MsftPwdPolHintsDeprecated: true,
				},
				Vendor: LDAPDiscoveryVendor{
					Name:                  "Microsoft Corporation",
					DomainFunctionalLevel: 4,
					ForestFunctionalLevel: 5,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			entry := &ldap.Entry{
				DN:         "",
				Attributes: []*ldap.EntryAttribute{},
			}

			if tc.haveObjectClass != nil {
				entry.Attributes = append(entry.Attributes, &ldap.EntryAttribute{Name: ldapObjectClassAttribute, Values: tc.haveObjectClass})
			}

			if tc.haveLDAPVersion != nil {
				entry.Attributes = append(entry.Attributes, &ldap.EntryAttribute{Name: ldapSupportedLDAPVersionAttribute, Values: tc.haveLDAPVersion})
			}

			if tc.haveExtensionOIDs != nil {
				entry.Attributes = append(entry.Attributes, &ldap.EntryAttribute{Name: ldapSupportedExtensionAttribute, Values: tc.haveExtensionOIDs})
			}

			if tc.haveControlOIDs != nil {
				entry.Attributes = append(entry.Attributes, &ldap.EntryAttribute{Name: ldapSupportedControlAttribute, Values: tc.haveControlOIDs})
			}

			if tc.haveFeatureOIDs != nil {
				entry.Attributes = append(entry.Attributes, &ldap.EntryAttribute{Name: ldapSupportedFeaturesAttribute, Values: tc.haveFeatureOIDs})
			}

			if tc.haveSASLMechanisms != nil {
				entry.Attributes = append(entry.Attributes, &ldap.EntryAttribute{Name: ldapSupportedSASLMechanismsAttribute, Values: tc.haveSASLMechanisms})
			}

			if tc.haveVendorName != nil {
				entry.Attributes = append(entry.Attributes, &ldap.EntryAttribute{Name: ldapVendorNameAttribute, Values: tc.haveVendorName})
			}

			if tc.haveVendorVersion != nil {
				entry.Attributes = append(entry.Attributes, &ldap.EntryAttribute{Name: ldapVendorVersionAttribute, Values: tc.haveVendorVersion})
			}

			if tc.haveDomainFunctionalLevel != nil {
				entry.Attributes = append(entry.Attributes, &ldap.EntryAttribute{Name: ldapDomainFunctionalityAttribute, Values: tc.haveDomainFunctionalLevel})
			}

			if tc.haveForestFunctionalLevel != nil {
				entry.Attributes = append(entry.Attributes, &ldap.EntryAttribute{Name: ldapForestFunctionalityAttribute, Values: tc.haveForestFunctionalLevel})
			}

			actual := ldapGetDiscoveryFromLDAPEntry(entry)

			assert.Equal(t, tc.haveExtensionOIDs, actual.Extensions.OIDs)
			assert.Equal(t, tc.haveControlOIDs, actual.Controls.OIDs)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestLDAPEntriesContainsEntry(t *testing.T) {
	testCases := []struct {
		description string
		have        []*ldap.Entry
		lookingFor  *ldap.Entry
		expected    bool
	}{
		{
			description: "ShouldNotMatchNil",
			have: []*ldap.Entry{
				{DN: "test"},
			},
			lookingFor: nil,
			expected:   false,
		},
		{
			description: "ShouldMatch",
			have: []*ldap.Entry{
				{DN: "test"},
			},
			lookingFor: &ldap.Entry{DN: "test"},
			expected:   true,
		},
		{
			description: "ShouldMatchWhenMultiple",
			have: []*ldap.Entry{
				{DN: "False"},
				{DN: "test"},
			},
			lookingFor: &ldap.Entry{DN: "test"},
			expected:   true,
		},
		{
			description: "ShouldNotMatchDifferent",
			have: []*ldap.Entry{
				{DN: "False"},
				{DN: "test"},
			},
			lookingFor: &ldap.Entry{DN: "not a result"},
			expected:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expected, ldapEntriesContainsEntry(tc.lookingFor, tc.have))
		})
	}
}

var testBERPacketReferral = ber.Packet{
	Children: []*ber.Packet{
		{},
		{
			Identifier: ber.Identifier{
				Tag: ber.TagObjectDescriptor,
			},
			Children: []*ber.Packet{
				{
					Identifier: ber.Identifier{
						Tag: ber.TagBitString,
					},
					Children: []*ber.Packet{
						{
							Value: "ldap://192.168.0.1",
						},
					},
				},
			},
		},
	},
}

func TestLDAPDiscovery_Strings(t *testing.T) {
	testCases := []struct {
		name                   string
		have                   LDAPDiscovery
		expectedExtensions     string
		expectedControls       string
		expectedFeatures       string
		expectedSASLMechanisms string
	}{
		{
			"ShouldHandleAll",
			LDAPDiscovery{
				Successful:     true,
				SASLMechanisms: []string{"SCRAM", "CRAM"},
				Extensions: LDAPDiscoveryExtensions{
					OIDs: []string{ldapOIDExtensionPwdModify, ldapOIDExtensionTLS},
				},
				Controls: LDAPDiscoveryControls{
					OIDs: []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated},
				},
				Features: LDAPDiscoveryFeatures{
					OIDs: []string{"example"},
				},
			},
			"1.3.6.1.4.1.4203.1.11.1, 1.3.6.1.4.1.1466.20037",
			"1.2.840.113556.1.4.2239, 1.2.840.113556.1.4.2066",
			"example",
			"SCRAM, CRAM",
		},
		{
			"ShouldHandleNone",
			LDAPDiscovery{
				Successful: true,
			},
			"none",
			"none",
			"none",
			"none",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			extensions, controls, features, saslMechanisms := tc.have.Strings()

			assert.Equal(t, tc.expectedExtensions, extensions)
			assert.Equal(t, tc.expectedControls, controls)
			assert.Equal(t, tc.expectedFeatures, features)
			assert.Equal(t, tc.expectedSASLMechanisms, saslMechanisms)
		})
	}
}
