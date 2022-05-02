package authentication

import (
	"fmt"
	"strings"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/go-ldap/ldap/v3"
)

func ldapEntriesContainsEntry(needle *ldap.Entry, haystack []*ldap.Entry) bool {
	for i := 0; i < len(haystack); i++ {
		if haystack[i].DN == needle.DN {
			return true
		}
	}

	return false
}

func ldapEscape(inputUsername string) string {
	inputUsername = ldap.EscapeFilter(inputUsername)
	for _, c := range specialLDAPRunes {
		inputUsername = strings.ReplaceAll(inputUsername, string(c), fmt.Sprintf("\\%c", c))
	}

	return inputUsername
}

func ldapGetReferral(err error) (referral string, ok bool) {
	if !ldap.IsErrorWithCode(err, ldap.LDAPResultReferral) {
		return "", false
	}

	switch e := err.(type) {
	case *ldap.Error:
		if len(e.Packet.Children) < 2 {
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
