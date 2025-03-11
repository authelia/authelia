package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticationBackendExtraAttribute(t *testing.T) {
	testCases := []struct {
		name  string
		have  AuthenticationBackendExtraAttribute
		vtype string
		mv    bool
	}{
		{
			"ShouldReturnDefaultsWhenEmpty",
			AuthenticationBackendExtraAttribute{},
			"",
			false,
		},
		{
			"ShouldHandleStringTypeWithMultiValue",
			AuthenticationBackendExtraAttribute{
				ValueType:   "string",
				MultiValued: true,
			},
			"string",
			true,
		},
		{
			"ShouldHandleIntegerType",
			AuthenticationBackendExtraAttribute{
				ValueType:   "integer",
				MultiValued: false,
			},
			"integer",
			false,
		},
		{
			"ShouldHandleBooleanTypeWithMultiValue",
			AuthenticationBackendExtraAttribute{
				ValueType:   "boolean",
				MultiValued: true,
			},
			"boolean",
			true,
		},
		{
			"ShouldHandleEmptyTypeWithMultiValue",
			AuthenticationBackendExtraAttribute{
				ValueType:   "",
				MultiValued: true,
			},
			"",
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.vtype, tc.have.GetValueType())
			assert.Equal(t, tc.mv, tc.have.IsMultiValued())
		})
	}
}

func TestAuthenticationBackendLDAPAttributesAttribute(t *testing.T) {
	testCases := []struct {
		name  string
		have  AuthenticationBackendLDAPAttributesAttribute
		vtype string
		mv    bool
	}{
		{
			"ShouldReturnDefaultsWhenEmpty",
			AuthenticationBackendLDAPAttributesAttribute{},
			"",
			false,
		},
		{
			"ShouldHandleMultiValuedIntegerType",
			AuthenticationBackendLDAPAttributesAttribute{
				AuthenticationBackendExtraAttribute: AuthenticationBackendExtraAttribute{
					MultiValued: true,
					ValueType:   "integer",
				},
			},
			"integer",
			true,
		},
		{
			"ShouldHandleCommonLDAPAttribute",
			AuthenticationBackendLDAPAttributesAttribute{
				AuthenticationBackendExtraAttribute: AuthenticationBackendExtraAttribute{
					MultiValued: true,
					ValueType:   "string",
				},
				Name: "memberOf",
			},
			"string",
			true,
		},
		{
			"ShouldHandleBinaryAttribute",
			AuthenticationBackendLDAPAttributesAttribute{
				AuthenticationBackendExtraAttribute: AuthenticationBackendExtraAttribute{
					MultiValued: false,
					ValueType:   "binary",
				},
				Name: "userCertificate",
			},
			"binary",
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.vtype, tc.have.GetValueType())
			assert.Equal(t, tc.mv, tc.have.IsMultiValued())
		})
	}
}
