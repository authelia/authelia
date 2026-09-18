// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedactShouldRemoveSensitiveValues(t *testing.T) {
	data := &DataSessionElevation{
		Subject: Subject{Username: "john"},
		Notification: &Notification{
			Sent:  true,
			Title: "Confirm your identity",
			Values: &NotificationValues{
				Domain:            "example.com",
				OneTimeCode:       "ABC123",
				RevocationLinkURL: "https://example.com/revoke?token=secret",
			},
		},
	}

	redacted := Redact(data)

	elevation, ok := redacted.(*DataSessionElevation)

	require.True(t, ok)
	assert.Equal(t, "", elevation.Notification.Values.OneTimeCode)
	assert.Equal(t, "", elevation.Notification.Values.RevocationLinkURL)
	assert.Equal(t, "example.com", elevation.Notification.Values.Domain)
	assert.Equal(t, "john", elevation.Username)
}

func TestRedactShouldNotMutateTheInput(t *testing.T) {
	data := &DataSessionElevation{
		Notification: &Notification{Values: &NotificationValues{OneTimeCode: "ABC123"}},
	}

	_ = Redact(data)

	assert.Equal(t, "ABC123", data.Notification.Values.OneTimeCode)
}

func TestRedactShouldRedactSensitiveFieldsInsideSliceElements(t *testing.T) {
	data := &dataWithSensitiveSlice{
		Items: []sensitiveSliceElement{{Token: "secret-1"}, {Token: "secret-2"}},
	}

	redacted := Redact(data)

	typed, ok := redacted.(*dataWithSensitiveSlice)

	require.True(t, ok)
	require.Len(t, typed.Items, 2)
	assert.Equal(t, "", typed.Items[0].Token)
	assert.Equal(t, "", typed.Items[1].Token)

	typed.Items[0].Token = "mutated"
	assert.Equal(t, "secret-1", data.Items[0].Token, "redact must not alias the input's slice backing array")
}

func TestRedactShouldRedactSensitiveFieldsInsideMapValues(t *testing.T) {
	data := &dataWithSensitiveMap{
		Items: map[string]sensitiveMapElement{"a": {Token: "secret"}},
	}

	redacted := Redact(data)

	typed, ok := redacted.(*dataWithSensitiveMap)

	require.True(t, ok)
	assert.Equal(t, "", typed.Items["a"].Token)
	assert.Equal(t, "secret", data.Items["a"].Token)
}

func TestRedactedPayloadsShouldNeverMarshalASensitiveField(t *testing.T) {
	for _, name := range Registered() {
		descriptor, ok := Lookup(name)
		require.True(t, ok, name)

		data := descriptor.New()

		populate(reflect.ValueOf(data).Elem())

		redacted := Redact(data)

		// Without this a registered payload which Redact cannot rebuild would pass every assertion below, because
		// marshaling a nil payload yields the four bytes 'null' in which no sensitive key appears.
		require.NotNil(t, redacted, name)

		encoded, err := json.Marshal(redacted)
		require.NoError(t, err, name)

		for _, key := range sensitiveJSONKeys(reflect.TypeOf(data).Elem()) {
			require.NotEmpty(t, key, "event type %s has a sensitive field with no json tag", name)
			assert.NotContains(t, string(encoded), `"`+key+`"`, "event type %s leaked sensitive field %s", name, key)
		}
	}
}

func TestRedactShouldFailClosed(t *testing.T) {
	testCases := []struct {
		name string
		have Data
	}{
		{"ShouldRejectNil", nil},
		{"ShouldRejectATypedNilPointer", (*DataUserPassword)(nil)},
		{"ShouldRejectAValueWhichIsNotAPointerToAStruct", dataNotAPointerToAStruct{"token": "secret"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Nil(t, Redact(tc.have))
		})
	}
}

func TestMarshalShouldRefuseAPayloadWhichCannotBeRedacted(t *testing.T) {
	event := &Event{
		ID:   "3b9c1f2e-8a41-4d6b-9f7c-2e5a0d81b3c4",
		Type: TypeUserPasswordChanged,
		Data: dataNotAPointerToAStruct{"token": "secret"},
	}

	data, err := Marshal(event, "https://auth.example.com", "4.39.0", true)

	assert.EqualError(t, err, "the user.password.changed payload is nil or could not be redacted")
	assert.Nil(t, data)

	data, err = Marshal(&Event{ID: event.ID, Type: event.Type}, "https://auth.example.com", "4.39.0", false)

	assert.EqualError(t, err, "the user.password.changed payload is nil or could not be redacted")
	assert.Nil(t, data)

	// A typed nil is a non nil interface, so it survives a bare comparison and would otherwise encode as a null data
	// attribute, which no published schema allows. Redaction is disabled here because Redact rejects it first when it
	// is not.
	var typed *DataUserPassword

	data, err = Marshal(&Event{ID: event.ID, Type: event.Type, Data: typed}, "https://auth.example.com", "4.39.0", false)

	assert.EqualError(t, err, "the user.password.changed payload is nil or could not be redacted")
	assert.Nil(t, data)
}

func populate(v reflect.Value) {
	if populateScalar(v) {
		return
	}

	switch v.Kind() { //nolint:exhaustive // Only kinds that can hold data worth populating are handled; the rest are left at their zero value.
	case reflect.Struct:
		populateStruct(v)
	case reflect.Pointer:
		populatePointer(v)
	case reflect.Interface:
		populateInterface(v)
	case reflect.Slice:
		populateSlice(v)
	case reflect.Array:
		populateArray(v)
	case reflect.Map:
		populateMap(v)
	}
}

func populateScalar(v reflect.Value) bool {
	switch v.Kind() { //nolint:exhaustive // Only scalar kinds are handled here; everything else is a container handled by populate.
	case reflect.String:
		v.SetString("populated")
	case reflect.Bool:
		v.SetBool(true)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(1)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		v.SetUint(1)
	case reflect.Float32, reflect.Float64:
		v.SetFloat(1)
	default:
		return false
	}

	return true
}

func populateStruct(v reflect.Value) {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		if !t.Field(i).IsExported() {
			continue
		}

		populate(v.Field(i))
	}
}

func populatePointer(v reflect.Value) {
	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}

	populate(v.Elem())
}

func populateInterface(v reflect.Value) {
	if v.IsNil() {
		v.Set(reflect.ValueOf("populated"))
	}
}

func populateSlice(v reflect.Value) {
	if v.Len() > 0 {
		for i := 0; i < v.Len(); i++ {
			populate(v.Index(i))
		}

		return
	}

	elem := reflect.New(v.Type().Elem()).Elem()

	populate(elem)

	v.Set(reflect.Append(v, elem))
}

func populateArray(v reflect.Value) {
	for i := 0; i < v.Len(); i++ {
		populate(v.Index(i))
	}
}

func populateMap(v reflect.Value) {
	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}

	if v.Len() != 0 || v.Type().Key().Kind() != reflect.String {
		return
	}

	key := reflect.New(v.Type().Key()).Elem()
	key.SetString("populated")

	val := reflect.New(v.Type().Elem()).Elem()

	populate(val)

	v.SetMapIndex(key, val)
}

func sensitiveJSONKeys(t reflect.Type) (keys []string) {
	switch t.Kind() { //nolint:exhaustive // Only kinds that can statically carry a nested struct field need to be walked.
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)

			if !f.IsExported() {
				continue
			}

			if f.Tag.Get(TagSensitive) == "true" {
				keys = append(keys, strings.Split(f.Tag.Get("json"), ",")[0])

				continue
			}

			keys = append(keys, sensitiveJSONKeys(f.Type)...)
		}
	case reflect.Pointer, reflect.Slice, reflect.Array, reflect.Map:
		keys = append(keys, sensitiveJSONKeys(t.Elem())...)
	}

	return keys
}

type dataNotAPointerToAStruct map[string]string

func (d dataNotAPointerToAStruct) EventType() string {
	return TypeUserPasswordChanged
}

type sensitiveSliceElement struct {
	Token string `json:"token,omitempty" sensitive:"true"`
}

type dataWithSensitiveSlice struct {
	Items []sensitiveSliceElement `json:"items,omitempty"`
}

func (*dataWithSensitiveSlice) EventType() string { return "test.sensitive.slice" }

type sensitiveMapElement struct {
	Token string `json:"token,omitempty" sensitive:"true"`
}

type dataWithSensitiveMap struct {
	Items map[string]sensitiveMapElement `json:"items,omitempty"`
}

func (*dataWithSensitiveMap) EventType() string { return "test.sensitive.map" }

// TestRedactShouldOmitIdentityVerificationLinks pins the redaction of the identity verification payload end to end,
// through Marshal rather than through Redact alone. Both link fields carry a token in their query string, so a leak
// here hands a receiver the means to complete a password reset for the user.
func TestRedactShouldOmitIdentityVerificationLinks(t *testing.T) {
	//nolint:gosec // Not a credential: a structurally realistic value so the assertions prove the whole string is gone.
	const (
		token     = "eyJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImpvaG4ifQ.signature"
		linkURL   = "https://auth.example.com/reset-password/step2?token=" + token
		revokeURL = "https://auth.example.com/revoke/one-time-code?id=" + token
	)

	event := NewEvent(&DataIdentityVerification{
		Subject: Subject{Username: "john"},
		Action:  "ResetPassword",
		Notification: NewNotification(nil, false, "Reset your password", nil, &NotificationValues{
			Domain:            "example.com",
			LinkURL:           linkURL,
			LinkText:          "Reset",
			RevocationLinkURL: revokeURL,
		}),
	})

	redacted, err := Marshal(event, "https://auth.example.com", "4.39.0", true)

	require.NoError(t, err)

	assert.NotContains(t, string(redacted), token, "the token must not survive redaction in either link")
	assert.NotContains(t, string(redacted), linkURL)
	assert.NotContains(t, string(redacted), revokeURL)
	assert.NotContains(t, string(redacted), "link_url")
	assert.NotContains(t, string(redacted), "revocation_link_url")

	// The rest of the notification must survive, otherwise the event tells a receiver nothing.
	assert.Contains(t, string(redacted), "example.com")
	assert.Contains(t, string(redacted), "Reset your password")
	assert.Contains(t, string(redacted), "ResetPassword")

	full, err := Marshal(event, "https://auth.example.com", "4.39.0", false)

	require.NoError(t, err)
	assert.Contains(t, string(full), token, "disabling redaction must carry the links through")
}
