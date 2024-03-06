package authentication

import (
	"fmt"
	"strings"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/go-ldap/ldap/v3"
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

func ldapGetFeatureSupportFromEntry(entry *ldap.Entry) (controlTypeOIDs, extensionOIDs []string, features LDAPSupportedFeatures) {
	if entry == nil {
		return controlTypeOIDs, extensionOIDs, features
	}

	for _, attr := range entry.Attributes {
		switch attr.Name {
		case ldapSupportedControlAttribute:
			controlTypeOIDs = attr.Values

			for _, oid := range attr.Values {
				switch oid {
				case ldapOIDControlMsftServerPolicyHints:
					features.ControlTypes.MsftPwdPolHints = true
				case ldapOIDControlMsftServerPolicyHintsDeprecated:
					features.ControlTypes.MsftPwdPolHintsDeprecated = true
				}
			}
		case ldapSupportedExtensionAttribute:
			extensionOIDs = attr.Values

			for _, oid := range attr.Values {
				switch oid {
				case ldapOIDExtensionPwdModifyExOp:
					features.Extensions.PwdModifyExOp = true
				case ldapOIDExtensionTLS:
					features.Extensions.TLS = true
				}
			}
		}
	}

	return controlTypeOIDs, extensionOIDs, features
}

func ldapEscape(inputUsername string) string {
	inputUsername = ldap.EscapeFilter(inputUsername)
	for _, c := range specialLDAPRunes {
		inputUsername = strings.ReplaceAll(inputUsername, string(c), fmt.Sprintf("\\%c", c))
	}

	return inputUsername
}

func ldapGetReferral(err error) (referral string, ok bool) {
	switch e := err.(type) {
	case *ldap.Error:
		if e.ResultCode != ldap.LDAPResultReferral {
			return "", false
		}

		if e.Packet == nil {
			return "", false
		}

		if len(e.Packet.Children) < 2 {
			return "", false
		}

		if e.Packet.Children[1].Tag != ber.TagObjectDescriptor {
			return "", false
		}

		for i := 0; i < len(e.Packet.Children[1].Children); i++ {
			if e.Packet.Children[1].Children[i].Tag != ber.TagBitString || len(e.Packet.Children[1].Children[i].Children) < 1 {
				continue
			}

			referral, ok = e.Packet.Children[1].Children[i].Children[0].Value.(string)

			if !ok {
				continue
			}

			return referral, true
		}

		return "", false
	default:
		return "", false
	}
}
