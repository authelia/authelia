package authentication

import (
	"errors"
	"testing"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
)

func TestLDAPGetReferral(t *testing.T) {
	testCases := []struct {
		description      string
		have             error
		expectedReferral string
		expectedOK       bool
	}{
		{
			description:      "ShouldGetValidPacket",
			have:             &ldap.Error{ResultCode: ldap.LDAPResultReferral, Packet: &testBERPacketReferral},
			expectedReferral: "ldap://192.168.0.1",
			expectedOK:       true,
		},
		{
			description:      "ShouldNotGetInvalidPacketWithNoObjectDescriptor",
			have:             &ldap.Error{ResultCode: ldap.LDAPResultReferral, Packet: &testBERPacketReferralInvalidObjectDescriptor},
			expectedReferral: "",
			expectedOK:       false,
		},
		{
			description:      "ShouldNotGetInvalidPacketWithBadErrorCode",
			have:             &ldap.Error{ResultCode: ldap.LDAPResultBusy, Packet: &testBERPacketReferral},
			expectedReferral: "",
			expectedOK:       false,
		},
		{
			description:      "ShouldNotGetInvalidPacketWithoutBitString",
			have:             &ldap.Error{ResultCode: ldap.LDAPResultReferral, Packet: &testBERPacketReferralWithoutBitString},
			expectedReferral: "",
			expectedOK:       false,
		},
		{
			description:      "ShouldNotGetInvalidPacketWithInvalidBitString",
			have:             &ldap.Error{ResultCode: ldap.LDAPResultReferral, Packet: &testBERPacketReferralWithInvalidBitString},
			expectedReferral: "",
			expectedOK:       false,
		},
		{
			description:      "ShouldNotGetInvalidPacketWithoutEnoughChildren",
			have:             &ldap.Error{ResultCode: ldap.LDAPResultReferral, Packet: &testBERPacketReferralWithoutEnoughChildren},
			expectedReferral: "",
			expectedOK:       false,
		},
		{
			description:      "ShouldNotGetInvalidErrType",
			have:             errors.New("not an err"),
			expectedReferral: "",
			expectedOK:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			referral, ok := ldapGetReferral(tc.have)

			assert.Equal(t, tc.expectedOK, ok)
			assert.Equal(t, tc.expectedReferral, referral)
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

var testBERPacketReferralInvalidObjectDescriptor = ber.Packet{
	Children: []*ber.Packet{
		{},
		{
			Identifier: ber.Identifier{
				Tag: ber.TagEOC,
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

var testBERPacketReferralWithoutBitString = ber.Packet{
	Children: []*ber.Packet{
		{},
		{
			Identifier: ber.Identifier{
				Tag: ber.TagObjectDescriptor,
			},
			Children: []*ber.Packet{
				{
					Identifier: ber.Identifier{
						Tag: ber.TagSequence,
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

var testBERPacketReferralWithInvalidBitString = ber.Packet{
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
							Value: 55,
						},
					},
				},
			},
		},
	},
}

var testBERPacketReferralWithoutEnoughChildren = ber.Packet{
	Children: []*ber.Packet{
		{
			Identifier: ber.Identifier{
				Tag: ber.TagEOC,
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
