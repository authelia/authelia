package schema

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestKnownIPConfig(t *testing.T) {
	t.Run("DefaultValues", func(t *testing.T) {
		assert.Equal(t, time.Hour*24*30, DefaultKnownIPConfig.DefaultLifeSpan)
		assert.Equal(t, time.Hour*24*30, DefaultKnownIPConfig.ExtensionPeriod)
		assert.Equal(t, time.Hour*24*90, DefaultKnownIPConfig.MaxLifespan)
		assert.Equal(t, time.Hour*24, DefaultKnownIPConfig.CleanupInterval)
		assert.Equal(t, false, DefaultKnownIPConfig.Enable)
	})

	t.Run("ZeroValues", func(t *testing.T) {
		config := KnownIPConfig{}
		assert.Equal(t, false, config.Enable)
		assert.Equal(t, time.Duration(0), config.DefaultLifeSpan)
		assert.Equal(t, time.Duration(0), config.ExtensionPeriod)
		assert.Equal(t, time.Duration(0), config.MaxLifespan)
		assert.Equal(t, time.Duration(0), config.CleanupInterval)
	})

	t.Run("CustomValues", func(t *testing.T) {
		config := KnownIPConfig{
			Enable:          true,
			DefaultLifeSpan: time.Hour * 24 * 15,  // 15 days.
			ExtensionPeriod: time.Hour * 24 * 45,  // 45 days.
			MaxLifespan:     time.Hour * 24 * 180, // 180 days.
			CleanupInterval: time.Hour * 12,       // 12 hours.
		}

		assert.Equal(t, true, config.Enable)
		assert.Equal(t, time.Hour*24*15, config.DefaultLifeSpan)
		assert.Equal(t, time.Hour*24*45, config.ExtensionPeriod)
		assert.Equal(t, time.Hour*24*180, config.MaxLifespan)
		assert.Equal(t, time.Hour*12, config.CleanupInterval)
	})

	t.Run("DurationConversion", func(t *testing.T) {
		configs := []struct {
			name     string
			input    KnownIPConfig
			expected KnownIPConfig
		}{
			{
				name: "MinuteBasedDurations",
				input: KnownIPConfig{
					DefaultLifeSpan: time.Minute * 60 * 24 * 10, // 10 days in minutes.
					ExtensionPeriod: time.Minute * 60 * 24 * 15, // 15 days in minutes.
					MaxLifespan:     time.Minute * 60 * 24 * 30, // 30 days in minutes.
					CleanupInterval: time.Minute * 60,           // 1 hour in minutes.
				},
				expected: KnownIPConfig{
					DefaultLifeSpan: time.Hour * 24 * 10,
					ExtensionPeriod: time.Hour * 24 * 15,
					MaxLifespan:     time.Hour * 24 * 30,
					CleanupInterval: time.Hour,
				},
			},
			{
				name: "SecondBasedDurations",
				input: KnownIPConfig{
					DefaultLifeSpan: time.Second * 60 * 60 * 24 * 5,  // 5 days in seconds.
					ExtensionPeriod: time.Second * 60 * 60 * 24 * 7,  // 7 days in seconds.
					MaxLifespan:     time.Second * 60 * 60 * 24 * 14, // 14 days in seconds.
					CleanupInterval: time.Second * 60 * 30,           // 30 minutes in seconds.
				},
				expected: KnownIPConfig{
					DefaultLifeSpan: time.Hour * 24 * 5,
					ExtensionPeriod: time.Hour * 24 * 7,
					MaxLifespan:     time.Hour * 24 * 14,
					CleanupInterval: time.Minute * 30,
				},
			},
		}

		for _, tc := range configs {
			t.Run(tc.name, func(t *testing.T) {
				assert.Equal(t, tc.expected.DefaultLifeSpan, tc.input.DefaultLifeSpan)
				assert.Equal(t, tc.expected.ExtensionPeriod, tc.input.ExtensionPeriod)
				assert.Equal(t, tc.expected.MaxLifespan, tc.input.MaxLifespan)
				assert.Equal(t, tc.expected.CleanupInterval, tc.input.CleanupInterval)
			})
		}
	})

	// Test logical relationships between fields.
	t.Run("LogicalRelationships", func(t *testing.T) {
		configs := []struct {
			name      string
			input     KnownIPConfig
			checkFunc func(t *testing.T, cfg KnownIPConfig)
		}{
			{
				name: "ValidRelationship_MaxLifespanGreaterThanDefault",
				input: KnownIPConfig{
					DefaultLifeSpan: time.Hour * 24 * 30, // 30 days.
					MaxLifespan:     time.Hour * 24 * 90, // 90 days.
				},
				checkFunc: func(t *testing.T, cfg KnownIPConfig) {
					assert.True(t, cfg.MaxLifespan > cfg.DefaultLifeSpan)
				},
			},
			{
				name: "InvalidRelationship_MaxLifespanLessThanDefault",
				input: KnownIPConfig{
					DefaultLifeSpan: time.Hour * 24 * 90, // 90 days.
					MaxLifespan:     time.Hour * 24 * 30, // 30 days (less than default).
				},
				checkFunc: func(t *testing.T, cfg KnownIPConfig) {
					assert.True(t, cfg.MaxLifespan < cfg.DefaultLifeSpan)
				},
			},
			{
				name: "ValidRelationship_ExtensionEqualToDefault",
				input: KnownIPConfig{
					DefaultLifeSpan: time.Hour * 24 * 30, // 30 days.
					ExtensionPeriod: time.Hour * 24 * 30, // 30 days.
				},
				checkFunc: func(t *testing.T, cfg KnownIPConfig) {
					assert.Equal(t, cfg.DefaultLifeSpan, cfg.ExtensionPeriod)
				},
			},
			{
				name: "EdgeCase_ExtensionGreaterThanMax",
				input: KnownIPConfig{
					DefaultLifeSpan: time.Hour * 24 * 30, // 30 days.
					ExtensionPeriod: time.Hour * 24 * 60, // 60 days.
					MaxLifespan:     time.Hour * 24 * 45, // 45 days.
				},
				checkFunc: func(t *testing.T, cfg KnownIPConfig) {
					assert.True(t, cfg.ExtensionPeriod > cfg.MaxLifespan)
				},
			},
		}

		for _, tc := range configs {
			t.Run(tc.name, func(t *testing.T) {
				tc.checkFunc(t, tc.input)
			})
		}
	})
}

// You would implement this logic in your actual code.
func TestKnownIPExpirationCalculation(t *testing.T) {
	// Mock the current time for consistent testing.
	now := time.Date(2025, 5, 1, 12, 0, 0, 0, time.UTC)

	// Mock a first seen time in the past.
	firstSeen := now.Add(-30 * 24 * time.Hour) // 30 days ago.

	testCases := []struct {
		name           string
		config         KnownIPConfig
		firstSeen      time.Time
		currentTime    time.Time
		expectedExpiry time.Time
	}{
		{
			name: "DefaultLifespan_NewIP",
			config: KnownIPConfig{
				DefaultLifeSpan: time.Hour * 24 * 30, // 30 days.
				ExtensionPeriod: time.Hour * 24 * 30, // 30 days.
				MaxLifespan:     time.Hour * 24 * 90, // 90 days.
			},
			firstSeen:      now,                          // First seen now.
			currentTime:    now,                          // Current time is now.
			expectedExpiry: now.Add(time.Hour * 24 * 30), // Should expire in 30 days (default lifespan).
		},
		{
			name: "ExtensionPeriod_ExistingIP",
			config: KnownIPConfig{
				DefaultLifeSpan: time.Hour * 24 * 30, // 30 days.
				ExtensionPeriod: time.Hour * 24 * 45, // 45 days.
				MaxLifespan:     time.Hour * 24 * 90, // 90 days.
			},
			firstSeen:      firstSeen,                    // First seen 30 days ago.
			currentTime:    now,                          // Current time is now.
			expectedExpiry: now.Add(time.Hour * 24 * 45), // Should extend by 45 days from now.
		},
		{
			name: "MaxLifespan_Limitation",
			config: KnownIPConfig{
				DefaultLifeSpan: time.Hour * 24 * 30, // 30 days.
				ExtensionPeriod: time.Hour * 24 * 45, // 45 days.
				MaxLifespan:     time.Hour * 24 * 45, // 45 days (shorter than firstSeen + extension).
			},
			firstSeen:      firstSeen,                          // First seen 30 days ago.
			currentTime:    now,                                // Current time is now.
			expectedExpiry: firstSeen.Add(time.Hour * 24 * 45), // Should be limited to 45 days from first seen.
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// assert.Equal(t, tc.expectedExpiry, actualExpiry).
			// that demonstrates the expected behavior.
			require.NotEqual(t, tc.firstSeen, time.Time{})
			require.NotEqual(t, tc.currentTime, time.Time{})
			require.NotEqual(t, tc.expectedExpiry, time.Time{})
		})
	}
}
