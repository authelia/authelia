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
	assert.Equal(t, "https://www.authelia.com/schemas/webhooks/v1/user.password.changed.json", decoded["dataschema"])
	assert.Equal(t, "4.39.0", decoded["autheliaversion"])
	assert.Equal(t, "john", decoded["subject"])
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
