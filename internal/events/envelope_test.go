// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEventShouldPopulateIdentifiers(t *testing.T) {
	event := NewEvent(&DataUserPassword{Type: TypeUserPasswordChanged, Subject: Subject{Username: "john"}})

	assert.Equal(t, TypeUserPasswordChanged, event.Type)
	assert.Equal(t, "john", event.Subject)
	assert.NotEmpty(t, event.ID)

	id, err := uuid.Parse(event.ID)

	require.NoError(t, err)
	assert.Equal(t, uuid.Version(7), id.Version(), "the identifier must be a UUIDv7 so identifiers sort in occurrence order")
	assert.False(t, event.Time.IsZero())
}

func TestMarshalShouldProduceACloudEvent(t *testing.T) {
	event := &Event{
		ID:      "3b9c1f2e-8a41-4d6b-9f7c-2e5a0d81b3c4",
		Type:    TypeUserPasswordChanged,
		Time:    time.Date(2026, 9, 12, 4, 5, 6, 0, time.UTC),
		Subject: "john",
		Data:    &DataUserPassword{Type: TypeUserPasswordChanged, Subject: Subject{Username: "john"}},
	}

	data, err := Marshal(event, "https://auth.example.com", "4.39.0", true)
	require.NoError(t, err)

	decoded := map[string]any{}
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "1.0", decoded["specversion"])
	assert.Equal(t, "3b9c1f2e-8a41-4d6b-9f7c-2e5a0d81b3c4", decoded["id"])
	assert.Equal(t, TypeUserPasswordChanged, decoded["type"])
	assert.Equal(t, "https://auth.example.com", decoded["source"])
	assert.Equal(t, "2026-09-12T04:05:06Z", decoded["time"])
	assert.Equal(t, "application/json", decoded["datacontenttype"])
	assert.Equal(t, "https://www.authelia.com/schemas/webhooks/v1/com.authelia.user.password.changed.json", decoded["dataschema"])
	assert.Equal(t, "4.39.0", decoded["autheliaversion"])
	assert.Equal(t, "john", decoded["subject"])
}

func TestMarshalWithHeadersShouldProduceHeadersMatchingTheDocument(t *testing.T) {
	event := &Event{
		ID:      "3b9c1f2e-8a41-4d6b-9f7c-2e5a0d81b3c4",
		Type:    TypeUserPasswordChanged,
		Time:    time.Date(2026, 9, 12, 4, 5, 6, 789000000, time.UTC),
		Subject: "john smith",
		Data:    &DataUserPassword{Type: TypeUserPasswordChanged, Subject: Subject{Username: "john"}},
	}

	data, headers, err := MarshalWithHeaders(event, "https://auth.example.com", "4.39.0 dev", true)
	require.NoError(t, err)

	expected, err := Marshal(event, "https://auth.example.com", "4.39.0 dev", true)
	require.NoError(t, err)

	assert.Equal(t, expected, data)

	decoded := map[string]any{}
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "2026-09-12T04:05:06.789Z", decoded["time"])

	assert.Equal(t, map[string]string{
		"ce-specversion":     "1.0",
		"ce-id":              "3b9c1f2e-8a41-4d6b-9f7c-2e5a0d81b3c4",
		"ce-type":            "com.authelia.user.password.changed",
		"ce-source":          "https://auth.example.com",
		"ce-time":            "2026-09-12T04:05:06.789Z",
		"ce-dataschema":      "https://www.authelia.com/schemas/webhooks/v1/com.authelia.user.password.changed.json",
		"ce-autheliaversion": "4.39.0%20dev",
		"ce-subject":         "john%20smith",
	}, headers)
}

func TestNewEventShouldPopulateTheSubject(t *testing.T) {
	testCases := []struct {
		name     string
		have     Data
		expected string
	}{
		{"ShouldUseTheUsername", &DataUserPassword{Type: TypeUserPasswordChanged, Subject: Subject{Username: "john"}}, "john"},
		{"ShouldUseTheUsernameOfAnAuthentication", &DataAuthentication{Type: TypeSecurityAuthenticationFailed, Username: "john"}, "john"},
		{"ShouldBeEmptyWithoutAUsername", &DataAuthentication{Type: TypeSecurityAuthenticationFailed}, ""},
		{"ShouldUseTheBannedUsername", &DataBan{Type: TypeSecurityBanApplied, Subject: Subject{Username: "john"}, Target: "john", TargetType: TargetTypeUser}, "john"},
		{"ShouldUseTheBannedAddress", &DataBan{Type: TypeSecurityBanApplied, Target: "198.51.100.12", TargetType: TargetTypeIP}, "198.51.100.12"},
		{"ShouldBeEmptyForTheStartupCheck", &DataStartupCheck{Probe: true}, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := NewEvent(tc.have)

			assert.Equal(t, tc.expected, event.Subject)

			data, headers, err := MarshalWithHeaders(event, "https://auth.example.com", "4.39.0", true)
			require.NoError(t, err)

			decoded := map[string]any{}
			require.NoError(t, json.Unmarshal(data, &decoded))

			if tc.expected == "" {
				assert.NotContains(t, decoded, "subject")
				assert.NotContains(t, headers, HeaderCloudEventsSubject)

				return
			}

			assert.Equal(t, tc.expected, decoded["subject"])
			assert.Equal(t, tc.expected, headers[HeaderCloudEventsSubject])
		})
	}
}

func TestMarshalWithHeadersShouldRejectAnEmptySource(t *testing.T) {
	event := NewEvent(&DataUserPassword{Type: TypeUserPasswordChanged, Subject: Subject{Username: "john"}})

	data, headers, err := MarshalWithHeaders(event, "", "4.39.0", true)

	assert.Nil(t, data)
	assert.Nil(t, headers)
	assert.EqualError(t, err, "the com.authelia.user.password.changed event has no source")
}

func TestEscapeHeaderValue(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected string
	}{
		{"ShouldNotEncodeAnEmptyValue", "", ""},
		{"ShouldNotEncodePrintableASCII", "https://auth.example.com/a?b=c&d=e#f~!", "https://auth.example.com/a?b=c&d=e#f~!"},
		{"ShouldEncodeASpace", "4.39.0 dev", "4.39.0%20dev"},
		{"ShouldEncodeADoubleQuote", `a"b`, "a%22b"},
		{"ShouldEncodeAPercentSign", "100%", "100%25"},
		{"ShouldNotDecodeAnExistingEscape", "a%20b", "a%2520b"},
		{"ShouldEncodeNonASCIIAsUpperCaseUTF8", "Euro \u20ac \U0001f600", "Euro%20%E2%82%AC%20%F0%9F%98%80"},
		{"ShouldEncodeLatin1", "calf\u00e9", "calf%C3%A9"},
		{"ShouldEncodeControlCharacters", "a\tb\r\nc\x7f", "a%09b%0D%0Ac%7F"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, escapeHeaderValue(tc.have))
		})
	}
}

func TestMarshalShouldRejectAnEmptySource(t *testing.T) {
	event := NewEvent(&DataUserPassword{Type: TypeUserPasswordChanged, Subject: Subject{Username: "john"}})

	data, err := Marshal(event, "", "4.39.0", true)

	assert.Nil(t, data)
	assert.EqualError(t, err, "the com.authelia.user.password.changed event has no source")

	data, err = MarshalBatch([]*Event{event}, "", "4.39.0", true)

	assert.Nil(t, data)
	assert.EqualError(t, err, "the com.authelia.user.password.changed event has no source")
}

func TestMarshalShouldRedactByDefaultAndNotWhenDisabled(t *testing.T) {
	event := NewEvent(&DataSessionElevation{
		Subject:      Subject{Username: "john"},
		Notification: &Notification{Values: &NotificationValues{OneTimeCode: "ABC123"}},
	})

	redacted, err := Marshal(event, "https://auth.example.com", "4.39.0", true)
	require.NoError(t, err)
	assert.NotContains(t, string(redacted), "ABC123")

	full, err := Marshal(event, "https://auth.example.com", "4.39.0", false)
	require.NoError(t, err)
	assert.Contains(t, string(full), "ABC123")
}

func TestNoOpEmitterShouldDoNothing(t *testing.T) {
	emitter := NewNoOpEmitter()

	assert.NotPanics(t, func() {
		emitter.Emit(context.Background(), NewEvent(&DataUserPassword{Type: TypeUserPasswordChanged}))
		emitter.Emit(context.Background(), nil)
	})
}
