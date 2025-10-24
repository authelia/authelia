package authentication

import (
	"testing"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
)

func TestLDAPGetFeatureSupportFromNilEntry(t *testing.T) {
	control, extension, feature := ldapGetFeatureSupportFromEntry(nil)
	assert.Len(t, control, 0)
	assert.Len(t, extension, 0)
	assert.Equal(t, LDAPSupportedFeatures{}, feature)
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
			haveExtensionOIDs: []string{ldapOIDExtensionPwdModifyExOp},
			expected:          LDAPSupportedFeatures{Extensions: LDAPSupportedExtensions{PwdModifyExOp: true}},
		},
		{
			description:       "ShouldReturnExtensionTLS",
			haveControlOIDs:   []string{},
			haveExtensionOIDs: []string{ldapOIDExtensionTLS},
			expected:          LDAPSupportedFeatures{Extensions: LDAPSupportedExtensions{TLS: true}},
		},
		{
			description:       "ShouldReturnExtensionAll",
			haveControlOIDs:   []string{},
			haveExtensionOIDs: []string{ldapOIDExtensionTLS, ldapOIDExtensionPwdModifyExOp},
			expected:          LDAPSupportedFeatures{Extensions: LDAPSupportedExtensions{TLS: true, PwdModifyExOp: true}},
		},
		{
			description:       "ShouldReturnControlMsftPPolHints",
			haveControlOIDs:   []string{ldapOIDControlMsftServerPolicyHints},
			haveExtensionOIDs: []string{},
			expected:          LDAPSupportedFeatures{ControlTypes: LDAPSupportedControlTypes{MsftPwdPolHints: true}},
		},
		{
			description:       "ShouldReturnControlMsftPPolHintsDeprecated",
			haveControlOIDs:   []string{ldapOIDControlMsftServerPolicyHintsDeprecated},
			haveExtensionOIDs: []string{},
			expected:          LDAPSupportedFeatures{ControlTypes: LDAPSupportedControlTypes{MsftPwdPolHintsDeprecated: true}},
		},
		{
			description:       "ShouldReturnControlAll",
			haveControlOIDs:   []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated},
			haveExtensionOIDs: []string{},
			expected:          LDAPSupportedFeatures{ControlTypes: LDAPSupportedControlTypes{MsftPwdPolHints: true, MsftPwdPolHintsDeprecated: true}},
		},
		{
			description:       "ShouldReturnExtensionAndControlAll",
			haveControlOIDs:   []string{ldapOIDControlMsftServerPolicyHints, ldapOIDControlMsftServerPolicyHintsDeprecated},
			haveExtensionOIDs: []string{ldapOIDExtensionTLS, ldapOIDExtensionPwdModifyExOp},
			expected: LDAPSupportedFeatures{
				ControlTypes: LDAPSupportedControlTypes{MsftPwdPolHints: true, MsftPwdPolHintsDeprecated: true},
				Extensions:   LDAPSupportedExtensions{TLS: true, PwdModifyExOp: true},
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

			actualControlOIDs, actualExtensionOIDs, actual := ldapGetFeatureSupportFromEntry(entry)

			assert.Equal(t, tc.haveExtensionOIDs, actualExtensionOIDs)
			assert.Equal(t, tc.haveControlOIDs, actualControlOIDs)
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
