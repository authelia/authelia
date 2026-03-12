package webauthn_test

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/webauthn"
)

func TestIsCredentialCreationDiscoverable(t *testing.T) {
	testCases := []struct {
		name     string
		have     *protocol.ParsedCredentialCreationData
		expected bool
		message  string
	}{
		{
			"ShouldHandleNormativeCase",
			&protocol.ParsedCredentialCreationData{
				ParsedPublicKeyCredential: protocol.ParsedPublicKeyCredential{
					ClientExtensionResults: map[string]any{
						webauthn.ExtensionCredProps: map[string]any{
							webauthn.ExtensionCredPropsResidentKey: true,
						},
					},
				},
			},
			true,
			"Determined Credential Discoverability via Client Extension Results",
		},
		{
			"ShouldReturnFalseWrongType",
			&protocol.ParsedCredentialCreationData{
				ParsedPublicKeyCredential: protocol.ParsedPublicKeyCredential{
					ClientExtensionResults: map[string]any{
						webauthn.ExtensionCredProps: map[string]any{
							webauthn.ExtensionCredPropsResidentKey: 1,
						},
					},
				},
			},
			false,
			"Assuming Credential Discoverability is false as the 'rk' field for the 'credProps' extension in the Client Extension Results was not a boolean",
		},
		{
			"ShouldReturnFalseNoKey",
			&protocol.ParsedCredentialCreationData{
				ParsedPublicKeyCredential: protocol.ParsedPublicKeyCredential{
					ClientExtensionResults: map[string]any{
						webauthn.ExtensionCredProps: map[string]any{},
					},
				},
			},
			false,
			"Assuming Credential Discoverability is false as the 'rk' field for the 'credProps' extension was missing from the Client Extension Results",
		},
		{
			"ShouldReturnFalsePropsWrongType",
			&protocol.ParsedCredentialCreationData{
				ParsedPublicKeyCredential: protocol.ParsedPublicKeyCredential{
					ClientExtensionResults: map[string]any{
						webauthn.ExtensionCredProps: []string{},
					},
				},
			},
			false,
			"Assuming Credential Discoverability is false as the 'credProps' extension in the Client Extension Results does not appear to be a dictionary",
		},
		{
			"ShouldReturnFalsePropsNotSet",
			&protocol.ParsedCredentialCreationData{
				ParsedPublicKeyCredential: protocol.ParsedPublicKeyCredential{
					ClientExtensionResults: map[string]any{},
				},
			},
			false,
			"Assuming Credential Discoverability is false as the 'credProps' extension is missing from the Client Extension Results",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := mocks.NewMockAutheliaCtx(t)
			defer ctx.Close()

			ctx.SetLogLevel(logrus.TraceLevel)

			actual := webauthn.IsCredentialCreationDiscoverable(ctx.Ctx.Logger, tc.have)

			assert.Equal(t, tc.expected, actual)

			if tc.message != "" {
				entry := ctx.Hook.LastEntry()

				require.NotNil(t, entry)
				assert.Equal(t, tc.message, entry.Message)
			}
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

func TestGetRelatedOriginConfigByRPID(t *testing.T) {
	testCases := []struct {
		name     string
		config   schema.WebAuthn
		rpid     string
		expected bool
	}{
		{
			"ShouldReturnNilWhenNoRelatedOrigins",
			schema.WebAuthn{},
			"example.com",
			false,
		},
		{
			"ShouldReturnNilWhenNotFound",
			schema.WebAuthn{
				RelatedOrigins: map[string]schema.WebAuthnRelatedOrigin{
					"other.com": {Origins: []*url.URL{MustParseURL("https://other.com")}},
				},
			},
			"example.com",
			false,
		},
		{
			"ShouldReturnMatchWhenFound",
			schema.WebAuthn{
				RelatedOrigins: map[string]schema.WebAuthnRelatedOrigin{
					"example.com": {Origins: []*url.URL{MustParseURL("https://example.com")}},
				},
			},
			"example.com",
			true,
		},
		{
			"ShouldReturnMatchCaseInsensitive",
			schema.WebAuthn{
				RelatedOrigins: map[string]schema.WebAuthnRelatedOrigin{
					"example.com": {Origins: []*url.URL{MustParseURL("https://example.com")}},
				},
			},
			"EXAMPLE.COM",
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := webauthn.GetRelatedOriginConfigByRPID(tc.config, tc.rpid)

			if tc.expected {
				assert.NotNil(t, result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestGetRelatedOriginConfigByOrigin(t *testing.T) {
	testCases := []struct {
		name         string
		config       schema.WebAuthn
		origin       *url.URL
		expectedRPID string
		expectedNil  bool
	}{
		{
			"ShouldReturnNilForNilOrigin",
			schema.WebAuthn{
				RelatedOrigins: map[string]schema.WebAuthnRelatedOrigin{
					"example.com": {Origins: []*url.URL{MustParseURL("https://example.com")}},
				},
			},
			nil,
			"",
			true,
		},
		{
			"ShouldReturnNilWhenNoRelatedOrigins",
			schema.WebAuthn{},
			MustParseURL("https://example.com"),
			"",
			true,
		},
		{
			"ShouldReturnNilWhenOriginNotFound",
			schema.WebAuthn{
				RelatedOrigins: map[string]schema.WebAuthnRelatedOrigin{
					"other.com": {Origins: []*url.URL{MustParseURL("https://other.com")}},
				},
			},
			MustParseURL("https://example.com"),
			"",
			true,
		},
		{
			"ShouldMatchOrigin",
			schema.WebAuthn{
				RelatedOrigins: map[string]schema.WebAuthnRelatedOrigin{
					"example.com": {Origins: []*url.URL{MustParseURL("https://example.com"), MustParseURL("https://auth.example.com")}},
				},
			},
			MustParseURL("https://auth.example.com"),
			"example.com",
			false,
		},
		{
			"ShouldMatchOriginCaseInsensitive",
			schema.WebAuthn{
				RelatedOrigins: map[string]schema.WebAuthnRelatedOrigin{
					"example.com": {Origins: []*url.URL{MustParseURL("https://example.com")}},
				},
			},
			MustParseURL("HTTPS://EXAMPLE.COM"),
			"example.com",
			false,
		},
		{
			"ShouldNotMatchWhenSchemeDiffers",
			schema.WebAuthn{
				RelatedOrigins: map[string]schema.WebAuthnRelatedOrigin{
					"example.com": {Origins: []*url.URL{MustParseURL("https://example.com")}},
				},
			},
			MustParseURL("http://example.com"),
			"",
			true,
		},
		{
			"ShouldNotMatchWhenOriginHasPath",
			schema.WebAuthn{
				RelatedOrigins: map[string]schema.WebAuthnRelatedOrigin{
					"example.com": {Origins: []*url.URL{MustParseURL("https://example.com")}},
				},
			},
			MustParseURL("https://example.com/some/path"),
			"",
			true,
		},
		{
			"ShouldNotMatchWhenConfigOriginHasPath",
			schema.WebAuthn{
				RelatedOrigins: map[string]schema.WebAuthnRelatedOrigin{
					"example.com": {Origins: []*url.URL{MustParseURL("https://example.com/path")}},
				},
			},
			MustParseURL("https://example.com"),
			"",
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rpid, result := webauthn.GetRelatedOriginConfigByOrigin(tc.config, tc.origin)

			if tc.expectedNil {
				assert.Empty(t, rpid)
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tc.expectedRPID, rpid)
				assert.NotNil(t, result)
			}
		})
	}
}

func MustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}

	return u
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
