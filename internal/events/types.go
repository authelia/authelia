// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package events

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

// DataUserPassword is the payload for the user.password.changed and user.password.reset events.
type DataUserPassword struct {
	Subject

	Type         string        `json:"-"`
	Notification *Notification `json:"notification,omitempty" jsonschema:"title=Notification"`
}

// EventType returns the registered type name.
func (d *DataUserPassword) EventType() string {
	return d.Type
}

// DataUserCredential is the payload for the user.credential.* events.
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

// DataIdentityVerification is the payload for the user.identity_verification.started event.
type DataIdentityVerification struct {
	Subject

	Action       string        `json:"action,omitempty" jsonschema:"title=Action" jsonschema_description:"The action the verification authorizes."`
	Notification *Notification `json:"notification,omitempty" jsonschema:"title=Notification"`
}

// EventType returns the registered type name.
func (d *DataIdentityVerification) EventType() string {
	return TypeUserIdentityVerificationStarted
}

// DataSessionElevation is the payload for the user.session.elevation.requested event.
type DataSessionElevation struct {
	Subject

	Notification *Notification `json:"notification,omitempty" jsonschema:"title=Notification"`
}

// EventType returns the registered type name.
func (d *DataSessionElevation) EventType() string {
	return TypeUserSessionElevationRequested
}

// DataAuthentication is the payload for the security.authentication.* events.
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

// DataStartupCheck is the payload for the system.startup_check event. It is the synthetic probe a destination receives
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

// DataBan is the payload for the security.ban.* events.
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
