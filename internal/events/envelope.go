// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
)

// Event is an occurrence awaiting delivery. The identifier is generated once and is stable across delivery retries so
// that a receiver may use it as an idempotency key.
type Event struct {
	ID      string
	Type    string
	Time    time.Time
	Subject string
	Data    Data
}

// NewEvent returns an Event wrapping the payload with a generated identifier and the current time. The identifier is a
// UUIDv7, whose leading bits are a millisecond timestamp, so identifiers sort in occurrence order and a receiver
// storing them as a primary key gets sequential inserts rather than random ones.
func NewEvent(data Data) (event *Event) {
	event = &Event{
		ID:   uuid.Must(uuid.NewV7()).String(),
		Type: data.EventType(),
		Time: time.Now().UTC(),
		Data: data,
	}

	if subject, ok := data.(interface{ SubjectName() string }); ok {
		event.Subject = subject.SubjectName()
	}

	return event
}

// SubjectName returns the username the occurrence concerns, satisfying the interface NewEvent uses to populate the
// CloudEvents subject attribute.
func (s Subject) SubjectName() string {
	return s.Username
}

// Envelope is the CloudEvents 1.0 structured mode representation of an Event.
type Envelope struct {
	SpecVersion     string    `json:"specversion" jsonschema:"const=1.0,title=Spec Version"`
	ID              string    `json:"id" jsonschema:"format=uuid,title=ID" jsonschema_description:"A UUIDv7. Unique per occurrence and stable across delivery retries."`
	Type            string    `json:"type" jsonschema:"title=Type"`
	Source          string    `json:"source" jsonschema:"format=uri,title=Source" jsonschema_description:"The issuer URL of the Authelia instance."`
	Time            time.Time `json:"time" jsonschema:"format=date-time,title=Time"`
	DataContentType string    `json:"datacontenttype" jsonschema:"const=application/json,title=Data Content Type"`
	DataSchema      string    `json:"dataschema" jsonschema:"format=uri,title=Data Schema" jsonschema_description:"Always an https://www.authelia.com/schemas/webhooks/ URI. Mandatory in this profile."`
	AutheliaVersion string    `json:"autheliaversion" jsonschema:"title=Authelia Version"`
	Subject         string    `json:"subject,omitempty" jsonschema:"title=Subject"`
	Data            Data      `json:"data" jsonschema:"title=Data"`
}

// EnvelopeBatch is the CloudEvents 1.0 batched content mode representation, which is a JSON array of envelopes. A
// destination which configures batching delivers one of these per request instead of a bare envelope.
type EnvelopeBatch []Envelope

// Marshal encodes an Event as a CloudEvents structured mode JSON document. When redact is true every credential
// equivalent value is omitted.
//
// A payload which is nil, either because the event carries none or because Redact could not rebuild it and failed
// closed, is an error rather than a document. Encoding one would transmit an envelope whose data attribute is null,
// which no published schema allows, and the alternative of transmitting the payload unredacted is worse still.
func Marshal(event *Event, source, version string, redact bool) (data []byte, err error) {
	var envelope *Envelope

	if envelope, err = newEnvelope(event, source, version, redact); err != nil {
		return nil, err
	}

	return json.Marshal(envelope)
}

// MarshalBatch encodes several Events as a CloudEvents batched content mode JSON document, which is an array of the
// same envelopes Marshal produces. A receiver distinguishes the two by the Content-Type: batched deliveries use
// ContentTypeBatchHeader.
//
// The array preserves the order the events occurred in. An empty batch is an error rather than an empty array,
// because sending one would be a request which tells a receiver nothing.
func MarshalBatch(batch []*Event, source, version string, redact bool) (data []byte, err error) {
	if len(batch) == 0 {
		return nil, fmt.Errorf("the batch is empty")
	}

	envelopes := make(EnvelopeBatch, 0, len(batch))

	for _, event := range batch {
		var envelope *Envelope

		if envelope, err = newEnvelope(event, source, version, redact); err != nil {
			return nil, err
		}

		envelopes = append(envelopes, *envelope)
	}

	return json.Marshal(envelopes)
}

func isNilPayload(payload Data) bool {
	if payload == nil {
		return true
	}

	switch value := reflect.ValueOf(payload); value.Kind() { //nolint:exhaustive // Only the nil-able kinds can hold a typed nil; every other kind is a value and is never nil.
	case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan, reflect.UnsafePointer:
		return value.IsNil()
	default:
		return false
	}
}

func newEnvelope(event *Event, source, version string, redact bool) (envelope *Envelope, err error) {
	payload := event.Data

	if redact {
		payload = Redact(payload)
	}

	if isNilPayload(payload) {
		return nil, fmt.Errorf("the %s payload is nil or could not be redacted", event.Type)
	}

	envelope = &Envelope{
		SpecVersion:     SpecVersion,
		ID:              event.ID,
		Type:            event.Type,
		Source:          source,
		Time:            event.Time,
		DataContentType: DataContentType,
		DataSchema:      schemaBaseURL + event.Type + ".json",
		AutheliaVersion: version,
		Subject:         event.Subject,
		Data:            payload,
	}

	if descriptor, ok := Lookup(event.Type); ok && descriptor.DataSchema != "" {
		envelope.DataSchema = descriptor.DataSchema
	}

	return envelope, nil
}
