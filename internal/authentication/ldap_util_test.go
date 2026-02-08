package authentication

import (
	"testing"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
)

func TestLDAPGetFeatureSupportFromNilEntry(t *testing.T) {
	features := ldapGetFeatureSupportFromEntry(nil)
	assert.Len(t, features.Extensions.OIDs, 0)
	assert.Len(t, features.ControlTypes.OIDs, 0)
	assert.Equal(t, LDAPSupportedFeatures{}, features)
}

func TestLDAPGetFeatureSupportFromEntry(t *testing.T) {
	testCases := []struct {
		description                        string
		haveControlOIDs, haveExtensionOIDs []string
		expected                           LDAPSupportedFeatures
	}{
		{
			description:       "ShouldReturnExtensionPwdModifyExOp",
			haveControlOIDs:   []string{},
			haveExtensionOIDs: []string{ldapOIDExtensionPwdModify},
			expected:          LDAPSupportedFeatures{Extensions: LDAPSupportedExtensions{PwdModify: true, OIDs: []string{ldapOIDExtensionPwdModify}}, ControlTypes: LDAPSupportedControlTypes{OIDs: []string{}}},
		},
		{
			description:       "ShouldReturnExtensionTLS",
			haveControlOIDs:   []string{},
			haveExtensionOIDs: []string{ldapOIDExtensionTLS},
			expected:          LDAPSupportedFeatures{Extensions: LDAPSupportedExtensions{TLS: true, OIDs: []string{ldapOIDExtensionTLS}}, ControlTypes: LDAPSupportedControlTypes{OIDs: []string{}}},
		},
		{
			description:       "ShouldReturnExtensionAll",
			haveControlOIDs:   []string{},
			haveExtensionOIDs: []string{ldapOIDExtensionTLS, ldapOIDExtensionPwdModify},
			expected:          LDAPSupportedFeatures{Extensions: LDAPSupportedExtensions{TLS: true, PwdModify: true, OIDs: []string{ldapOIDExtensionTLS, ldapOIDExtensionPwdModify}}, ControlTypes: LDAPSupportedControlTypes{OIDs: []string{}}},
		},
		{
			description:       "ShouldReturnControlMsftPPolHints",
			haveControlOIDs:   []string{ldapOIDControlMsftServerPolicyHints},
			haveExtensionOIDs: []string{},
			expected:          LDAPSupportedFeatures{Extensions: LDAPSupportedExtensions{OIDs: []string{}}, ControlTypes: LDAPSupportedControlTypes{MsftPwdPolHints: true, OIDs: []string{ldapOIDControlMsftServerPolicyHints}}},
		},
		{
			description:       "ShouldReturnControlMsftPPolHintsDeprecated",
			haveControlOIDs:   []string{ldapOIDControlMsftServerPolicyHintsDeprecated},
			haveExtensionOIDs: []string{},
			expected:          LDAPSupportedFeatures{Extensions: LDAPSupportedExtensions{OIDs: []string{}}, ControlTypes: LDAPSupportedControlTypes{MsftPwdPolHintsDeprecated: true, OIDs: []string{ldapOIDControlMsftServerPolicyHintsDeprecated}}},
		},
		{
			description:       "ShouldReturnControlAll",
			haveControlOIDs:   []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated},
			haveExtensionOIDs: []string{},
			expected:          LDAPSupportedFeatures{Extensions: LDAPSupportedExtensions{OIDs: []string{}}, ControlTypes: LDAPSupportedControlTypes{MsftPwdPolHints: true, MsftPwdPolHintsDeprecated: true, OIDs: []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated}}},
		},
		{
			description:       "ShouldReturnExtensionAndControlAll",
			haveControlOIDs:   []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated},
			haveExtensionOIDs: []string{ldapOIDExtensionTLS, ldapOIDExtensionPwdModify},
			expected: LDAPSupportedFeatures{
				ControlTypes: LDAPSupportedControlTypes{MsftPwdPolHints: true, MsftPwdPolHintsDeprecated: true, OIDs: []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated}},
				Extensions:   LDAPSupportedExtensions{TLS: true, PwdModify: true, OIDs: []string{ldapOIDExtensionTLS, ldapOIDExtensionPwdModify}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			entry := &ldap.Entry{
				DN: "",
				Attributes: []*ldap.EntryAttribute{
					{Name: ldapSupportedExtensionAttribute, Values: tc.haveExtensionOIDs},
					{Name: ldapSupportedControlAttribute, Values: tc.haveControlOIDs},
				},
			}

			actual := ldapGetFeatureSupportFromEntry(entry)

			assert.Equal(t, tc.haveExtensionOIDs, actual.Extensions.OIDs)
			assert.Equal(t, tc.haveControlOIDs, actual.ControlTypes.OIDs)
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
