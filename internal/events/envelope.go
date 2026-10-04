// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
)

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

// Marshal encodes an Event as a CloudEvents structured mode JSON document. When redact is true every credential
// equivalent value is omitted.
//
// An empty source is an error rather than a document, because the CloudEvents source attribute is required to be
// non-empty.
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

// MarshalWithHeaders encodes an Event exactly as Marshal does and also returns the HTTP headers the CloudEvents HTTP
// protocol binding defines for its context attributes, which a sender may add to a structured mode request. The
// header values are derived from the same envelope as the document, so the two always agree, and are percent-encoded
// as the binding requires.
//
// The datacontenttype attribute has no header, because the binding maps it to Content-Type and forbids a
// ce-datacontenttype header.
func MarshalWithHeaders(event *Event, source, version string, redact bool) (data []byte, headers map[string]string, err error) {
	var envelope *Envelope

	if envelope, err = newEnvelope(event, source, version, redact); err != nil {
		return nil, nil, err
	}

	if data, err = json.Marshal(envelope); err != nil {
		return nil, nil, err
	}

	return data, headersOf(envelope), nil
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

	if source == "" {
		return nil, fmt.Errorf("the %s event has no source", event.Type)
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

func headersOf(envelope *Envelope) (headers map[string]string) {
	headers = map[string]string{
		HeaderCloudEventsSpecVersion:     escapeHeaderValue(envelope.SpecVersion),
		HeaderCloudEventsID:              escapeHeaderValue(envelope.ID),
		HeaderCloudEventsType:            escapeHeaderValue(envelope.Type),
		HeaderCloudEventsSource:          escapeHeaderValue(envelope.Source),
		HeaderCloudEventsTime:            escapeHeaderValue(envelope.Time.Format(time.RFC3339Nano)),
		HeaderCloudEventsDataSchema:      escapeHeaderValue(envelope.DataSchema),
		HeaderCloudEventsAutheliaVersion: escapeHeaderValue(envelope.AutheliaVersion),
	}

	if envelope.Subject != "" {
		headers[HeaderCloudEventsSubject] = escapeHeaderValue(envelope.Subject)
	}

	return headers
}

func escapeHeaderValue(value string) (escaped string) {
	var builder strings.Builder

	for i := 0; i < len(value); i++ {
		c := value[i]

		if c > ' ' && c < 0x7F && c != '"' && c != '%' {
			builder.WriteByte(c)

			continue
		}

		builder.WriteByte('%')
		builder.WriteByte(hexUpper[c>>4])
		builder.WriteByte(hexUpper[c&0x0F])
	}

	return builder.String()
}
