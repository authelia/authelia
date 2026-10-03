// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"
	"time"
)

// Data is implemented by every event payload.
type Data interface {
	// EventType returns the registered type name of this payload.
	EventType() string
}

// Subject describes the person an occurrence concerns. It is embedded in every payload.
type Subject struct {
	Username    string   `json:"username" jsonschema:"title=Username" jsonschema_description:"The username the occurrence concerns."`
	DisplayName string   `json:"display_name,omitempty" jsonschema:"title=Display Name" jsonschema_description:"The display name recorded in the authentication backend."`
	Emails      []string `json:"emails,omitempty" jsonschema:"title=Emails" jsonschema_description:"The email addresses recorded in the authentication backend."`
	RemoteIP    string   `json:"remote_ip,omitempty" jsonschema:"title=Remote IP" jsonschema_description:"The client address which caused the occurrence."`
}

// Recipient describes one addressee of a notification.
type Recipient struct {
	Email       string `json:"email" jsonschema:"title=Email" jsonschema_description:"The address the notification was addressed to."`
	Username    string `json:"username,omitempty" jsonschema:"title=Username" jsonschema_description:"The username the address belongs to when known."`
	DisplayName string `json:"display_name,omitempty" jsonschema:"title=Display Name" jsonschema_description:"The display name recorded in the authentication backend."`
}

// NotificationValues are the template attribute values of a notification. The rendered HTML and text bodies are never
// included.
type NotificationValues struct {
	BodyPrefix string         `json:"body_prefix,omitempty" jsonschema:"title=Body Prefix"`
	BodyEvent  string         `json:"body_event,omitempty" jsonschema:"title=Body Event"`
	BodySuffix string         `json:"body_suffix,omitempty" jsonschema:"title=Body Suffix"`
	Domain     string         `json:"domain,omitempty" jsonschema:"title=Domain"`
	Details    map[string]any `json:"details,omitempty" jsonschema:"title=Details"`

	LinkURL           string `json:"link_url,omitempty" sensitive:"true" jsonschema:"title=Link URL" jsonschema_description:"Credential equivalent. Omitted unless the destination disables redaction."`
	LinkText          string `json:"link_text,omitempty" jsonschema:"title=Link Text"`
	RevocationLinkURL string `json:"revocation_link_url,omitempty" sensitive:"true" jsonschema:"title=Revocation Link URL" jsonschema_description:"Credential equivalent. Omitted unless the destination disables redaction."`
	OneTimeCode       string `json:"one_time_code,omitempty" sensitive:"true" jsonschema:"title=One Time Code" jsonschema_description:"Credential equivalent. Omitted unless the destination disables redaction."`
}

// Notification describes a notification produced by an occurrence.
type Notification struct {
	Sent       bool                `json:"sent" jsonschema:"title=Sent" jsonschema_description:"Whether the notification was successfully handed to the notifier."`
	Suppressed bool                `json:"suppressed,omitempty" jsonschema:"title=Suppressed" jsonschema_description:"Whether the notification was never handed to a notifier because the notifier is disabled. Sent is false and error is empty in that case."`
	Recipients []Recipient         `json:"recipients,omitempty" jsonschema:"title=Recipients"`
	Title      string              `json:"title,omitempty" jsonschema:"title=Title"`
	Error      string              `json:"error,omitempty" jsonschema:"title=Error" jsonschema_description:"The delivery error when sent is false."`
	Values     *NotificationValues `json:"values,omitempty" jsonschema:"title=Values"`
}

// DataUserPassword is the payload for the com.authelia.user.password.changed and com.authelia.user.password.reset events.
type DataUserPassword struct {
	Subject

	Type         string        `json:"-"`
	Notification *Notification `json:"notification,omitempty" jsonschema:"title=Notification"`
}

// EventType returns the registered type name.
func (d *DataUserPassword) EventType() string {
	return d.Type
}

// DataUserCredential is the payload for the com.authelia.user.credential.* events.
type DataUserCredential struct {
	Subject

	Type         string        `json:"-"`
	Description  string        `json:"description,omitempty" jsonschema:"title=Description" jsonschema_description:"The description of the credential."`
	Notification *Notification `json:"notification,omitempty" jsonschema:"title=Notification"`
}

// EventType returns the registered type name.
func (d *DataUserCredential) EventType() string {
	return d.Type
}

// DataIdentityVerification is the payload for the com.authelia.user.identity_verification.started event.
type DataIdentityVerification struct {
	Subject

	Action       string        `json:"action,omitempty" jsonschema:"title=Action" jsonschema_description:"The action the verification authorizes."`
	Notification *Notification `json:"notification,omitempty" jsonschema:"title=Notification"`
}

// EventType returns the registered type name.
func (d *DataIdentityVerification) EventType() string {
	return TypeUserIdentityVerificationStarted
}

// DataSessionElevation is the payload for the com.authelia.user.session.elevation.requested event.
type DataSessionElevation struct {
	Subject

	Notification *Notification `json:"notification,omitempty" jsonschema:"title=Notification"`
}

// EventType returns the registered type name.
func (d *DataSessionElevation) EventType() string {
	return TypeUserSessionElevationRequested
}

// DataAuthentication is the payload for the com.authelia.security.authentication.* events.
type DataAuthentication struct {
	Subject

	Type   string `json:"-"`
	Stage  string `json:"stage" jsonschema:"enum=first_factor,enum=second_factor,title=Stage"`
	Method string `json:"method" jsonschema:"enum=password,enum=totp,enum=webauthn,enum=duo,title=Method"`
	Reason string `json:"reason,omitempty" jsonschema:"enum=invalid_credentials,enum=user_not_found,enum=banned,enum=internal_error,title=Reason" jsonschema_description:"The failure classification. Present on failures only."`
}

// EventType returns the registered type name.
func (d *DataAuthentication) EventType() string {
	return d.Type
}

// DataStartupCheck is the payload for the com.authelia.system.startup_check event. It is the synthetic probe a destination receives
// when the startup check is enabled, and it never describes a real occurrence: no user did anything, and nothing about
// it should be treated as an authentication, a security signal, or an audit record.
type DataStartupCheck struct {
	Subject

	Probe bool `json:"probe" jsonschema:"const=true,title=Probe" jsonschema_description:"Always true. Marks this as a connectivity probe rather than an occurrence, so a receiver can discard it."`
}

// EventType returns the registered type name.
func (d *DataStartupCheck) EventType() string {
	return TypeSystemStartupCheck
}

// DataBan is the payload for the com.authelia.security.ban.* events.
type DataBan struct {
	Subject

	Type       string `json:"-"`
	Target     string `json:"target" jsonschema:"title=Target" jsonschema_description:"The banned value, a username or an IP address."`
	TargetType string `json:"target_type" jsonschema:"enum=user,enum=ip,title=Target Type"`
	Expires    string `json:"expires,omitempty" jsonschema:"format=date-time,title=Expires" jsonschema_description:"When the ban expires. Present on applied only."`
}

// EventType returns the registered type name.
func (d *DataBan) EventType() string {
	return d.Type
}

// SubjectName returns the banned value, which is what a ban concerns whether it is a username or an address.
func (d *DataBan) SubjectName() string {
	return d.Target
}

// Descriptor describes a registered event type.
type Descriptor struct {
	// Type is the registered event type name.
	Type string

	// DataSchema is the absolute URL of the published JSON Schema for this type's data.
	DataSchema string

	// New returns an empty payload of the type this descriptor describes.
	New func() Data
}

// Emitter accepts an event for asynchronous delivery. Implementations must never block and must never return an
// error, because an authentication or self-service flow must not fail or stall because a webhook receiver is
// unhealthy. Implementations must not retain the context beyond the call, because callers pass the request context and
// the request is recycled once the handler returns.
type Emitter interface {
	Emit(ctx context.Context, event *Event)
}

// Confirmer is implemented by an Emitter whose destinations may grant permission to receive events asynchronously,
// by requesting the callback URL they were given in the CloudEvents abuse protection handshake.
type Confirmer interface {
	// Confirm grants the named destination permission to receive events when the key matches the one in its callback
	// URL. The rate is the value of the WebHook-Allowed-Rate header of the callback request, which may be empty.
	Confirm(name, key, rate string) (confirmed bool)
}

// Event is an occurrence awaiting delivery. The identifier is generated once and is stable across delivery retries so
// that a receiver may use it as an idempotency key.
type Event struct {
	ID      string
	Type    string
	Time    time.Time
	Subject string
	Data    Data
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
	Subject         string    `json:"subject,omitempty" jsonschema:"title=Subject" jsonschema_description:"What the occurrence concerns within the source: the username, or the banned value for a ban. Omitted when there is none."`
	Data            Data      `json:"data" jsonschema:"title=Data"`
}

// EnvelopeBatch is the CloudEvents 1.0 batched content mode representation, which is a JSON array of envelopes. A
// destination which configures batching delivers one of these per request instead of a bare envelope.
type EnvelopeBatch []Envelope
