package authentication

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestControlMsftServerPolicyHints(t *testing.T) {
	ct := &controlMsftServerPolicyHints{
		oid: ldapOIDControlMsftServerPolicyHints,
	}

	assert.Equal(t, ldapOIDControlMsftServerPolicyHints, ct.GetControlType())
	assert.Equal(t, "Enforce the password history length constraint (MS-SAMR section 3.1.1.7.1) during password set: 1.2.840.113556.1.4.2239", ct.String())
	assert.NotNil(t, ct.Encode())
}
