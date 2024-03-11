package authentication

import (
	ber "github.com/go-asn1-ber/asn1-ber"
)

type controlMsftServerPolicyHints struct {
	oid string
}

// GetControlType implements ldap.Control.
func (c *controlMsftServerPolicyHints) GetControlType() string {
	return c.oid
}

// Encode implements ldap.Control.
func (c *controlMsftServerPolicyHints) Encode() (packet *ber.Packet) {
	seq := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "PolicyHintsRequestValue")
	seq.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, 1, "Flags"))

	controlValue := ber.Encode(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, nil, "Control Value (Policy Hints)")
	controlValue.AppendChild(seq)

	packet = ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Control")
	packet.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, c.GetControlType(), "Control Type (LDAP_SERVER_POLICY_HINTS_OID)"))
	packet.AppendChild(ber.NewBoolean(ber.ClassUniversal, ber.TypePrimitive, ber.TagBoolean, true, "Criticality"))

	packet.AppendChild(controlValue)

	return packet
}

// String implements ldap.Control.
func (c *controlMsftServerPolicyHints) String() string {
	return "Enforce the password history length constraint (MS-SAMR section 3.1.1.7.1) during password set: " + c.GetControlType()
}
