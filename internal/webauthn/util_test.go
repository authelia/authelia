package webauthn_test

import (
	"fmt"
	"testing"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/webauthn"
)

func TestIsCredentialCreationDiscoverable(t *testing.T) {
	testCases := []struct {
		name     string
		have     *protocol.ParsedCredentialCreationData
		expected bool
	}{
		{
			"ShouldHandleNormativeCase",
			&protocol.ParsedCredentialCreationData{
				ParsedPublicKeyCredential: protocol.ParsedPublicKeyCredential{
					ClientExtensionResults: protocol.AuthenticationExtensionsClientOutputs{
						CredProps: &protocol.CredentialPropertiesOutput{RK: ptr(true)},
					},
				},
			},
			true,
		},
		{
			"ShouldReturnFalseResidentKeyFalse",
			&protocol.ParsedCredentialCreationData{
				ParsedPublicKeyCredential: protocol.ParsedPublicKeyCredential{
					ClientExtensionResults: protocol.AuthenticationExtensionsClientOutputs{
						CredProps: &protocol.CredentialPropertiesOutput{RK: ptr(false)},
					},
				},
			},
			false,
		},
		{
			"ShouldReturnFalseResidentKeyNotSet",
			&protocol.ParsedCredentialCreationData{
				ParsedPublicKeyCredential: protocol.ParsedPublicKeyCredential{
					ClientExtensionResults: protocol.AuthenticationExtensionsClientOutputs{
						CredProps: &protocol.CredentialPropertiesOutput{},
					},
				},
			},
			false,
		},
		{
			"ShouldReturnFalsePropsNotSet",
			&protocol.ParsedCredentialCreationData{
				ParsedPublicKeyCredential: protocol.ParsedPublicKeyCredential{
					ClientExtensionResults: protocol.AuthenticationExtensionsClientOutputs{},
				},
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, webauthn.IsCredentialCreationDiscoverable(tc.have))
		})
	}
}

func TestValidateCredentialAllowed(t *testing.T) {
	testCases := []struct {
		name     string
		config   *schema.WebAuthn
		have     *model.WebAuthnCredential
		expected string
	}{
		{
			"ShouldAllowNotConfigured",
			&schema.WebAuthn{},
			&model.WebAuthnCredential{AAGUID: model.NullUUID(uuid.Must(uuid.Parse("7a5d62c8-1164-41a5-807c-af16cccb8af4")))},
			"",
		},
		{
			"ShouldAllowNotConfigured",
			&schema.WebAuthn{},
			&model.WebAuthnCredential{AAGUID: model.NullUUID(uuid.Must(uuid.Parse("7a5d62c8-1164-41a5-807c-af16cccb8af4"))), BackupEligible: true, BackupState: true},
			"",
		},
		{
			"ShouldNotProhibitBackupEligibilityFalse",
			&schema.WebAuthn{Filtering: schema.WebAuthnFiltering{ProhibitBackupEligibility: true}},
			&model.WebAuthnCredential{AAGUID: model.NullUUID(uuid.Must(uuid.Parse("7a5d62c8-1164-41a5-807c-af16cccb8af4")))},
			"",
		},
		{
			"ShouldProhibitBackupEligibilityTrue",
			&schema.WebAuthn{Filtering: schema.WebAuthnFiltering{ProhibitBackupEligibility: true}},
			&model.WebAuthnCredential{AAGUID: model.NullUUID(uuid.Must(uuid.Parse("7a5d62c8-1164-41a5-807c-af16cccb8af4"))), BackupEligible: true},
			"error checking webauthn credential: filters have been configured which prohibit credentials that are backup eligible",
		},
		{
			"ShouldAllowPermittedAAGUIDs",
			&schema.WebAuthn{Filtering: schema.WebAuthnFiltering{PermittedAAGUIDs: []uuid.UUID{uuid.Must(uuid.Parse("7a5d62c8-1164-41a5-807c-af16cccb8af4"))}}},
			&model.WebAuthnCredential{AAGUID: model.NullUUID(uuid.Must(uuid.Parse("7a5d62c8-1164-41a5-807c-af16cccb8af4")))},
			"",
		},
		{
			"ShouldNotAllowUnallowedAAGUID",
			&schema.WebAuthn{Filtering: schema.WebAuthnFiltering{PermittedAAGUIDs: []uuid.UUID{uuid.Must(uuid.Parse("7a5d62c8-1164-41a5-807c-af16cccb8af5"))}}},
			&model.WebAuthnCredential{AAGUID: model.NullUUID(uuid.Must(uuid.Parse("7a5d62c8-1164-41a5-807c-af16cccb8af4")))},
			"error checking webauthn credential: filters have been configured which explicitly require only permitted AAGUID's be used and '7a5d62c8-1164-41a5-807c-af16cccb8af4' is not permitted",
		},
		{
			"ShouldAllowNotProhibitedAAGUID",
			&schema.WebAuthn{Filtering: schema.WebAuthnFiltering{ProhibitedAAGUIDs: []uuid.UUID{uuid.Must(uuid.Parse("7a5d62c8-1164-41a5-807c-af16cccb8af5"))}}},
			&model.WebAuthnCredential{AAGUID: model.NullUUID(uuid.Must(uuid.Parse("7a5d62c8-1164-41a5-807c-af16cccb8af4")))},
			"",
		},
		{
			"Should",
			&schema.WebAuthn{Filtering: schema.WebAuthnFiltering{ProhibitedAAGUIDs: []uuid.UUID{uuid.Must(uuid.Parse("7a5d62c8-1164-41a5-807c-af16cccb8af4"))}}},
			&model.WebAuthnCredential{AAGUID: model.NullUUID(uuid.Must(uuid.Parse("7a5d62c8-1164-41a5-807c-af16cccb8af4")))},
			"error checking webauthn credential: filters have been configured which prohibit the AAGUID '7a5d62c8-1164-41a5-807c-af16cccb8af4' from registration",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := webauthn.ValidateCredentialAllowed(tc.config, tc.have)

			if tc.expected == "" {
				assert.NoError(t, actual)
			} else {
				assert.EqualError(t, actual, tc.expected)
			}
		})
	}
}

func TestFormatError(t *testing.T) {
	testCases := []struct {
		name     string
		have     error
		expected string
	}{
		{
			"ShouldFormatNormalError",
			fmt.Errorf("example"),
			"example",
		},
		{
			"ShouldFormatEnhancedError",
			&protocol.Error{
				Type:    "example_type",
				Details: "example_details",
				DevInfo: "example_dev_info",
			},
			"example_details (example_type): example_dev_info",
		},
		{
			"ShouldFormatEnhancedErrorNoDevInfo",
			&protocol.Error{
				Type:    "example_type",
				Details: "example_details",
			},
			"example_details (example_type)",
		},
		{
			"ShouldFormatEnhancedErrorNoType",
			&protocol.Error{
				Details: "example_details",
				DevInfo: "example_dev_info",
			},
			"example_details: example_dev_info",
		}, {
			"ShouldFormatEnhancedErrorOnlyDetails",
			&protocol.Error{
				Details: "example_details",
			},
			"example_details",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.EqualError(t, webauthn.FormatError(tc.have), tc.expected)
		})
	}
}

func ptr[T any](in T) *T {
	return &in
}
