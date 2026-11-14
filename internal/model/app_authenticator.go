// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"database/sql"
	"time"
)

const (
	// AppAuthenticatorRegistrationStateIssued is the state of a registration code that has been issued.
	AppAuthenticatorRegistrationStateIssued = "issued"

	// AppAuthenticatorRegistrationStateBegun is the state of a registration which the device has begun.
	AppAuthenticatorRegistrationStateBegun = "begun"

	// AppAuthenticatorRegistrationStateCompleted is the state of a registration which completed successfully.
	AppAuthenticatorRegistrationStateCompleted = "completed"

	// AppAuthenticatorRegistrationStateVoided is the state of a registration which failed and was voided.
	AppAuthenticatorRegistrationStateVoided = "voided"

	// AppAuthenticatorRegistrationStateExpired is the state of a registration which expired.
	AppAuthenticatorRegistrationStateExpired = "expired"

	// AppAuthenticatorRegistrationStateCancelled is the state of a registration which was cancelled.
	AppAuthenticatorRegistrationStateCancelled = "cancelled"
)

const (
	// AppAuthenticatorChallengeStatePending is the state of a challenge awaiting a decision.
	AppAuthenticatorChallengeStatePending = "pending"

	// AppAuthenticatorChallengeStateApproved is the state of an approved challenge.
	AppAuthenticatorChallengeStateApproved = "approved"

	// AppAuthenticatorChallengeStateDenied is the state of a denied challenge.
	AppAuthenticatorChallengeStateDenied = "denied"

	// AppAuthenticatorChallengeStateVoided is the state of a voided challenge.
	AppAuthenticatorChallengeStateVoided = "voided"

	// AppAuthenticatorChallengeStateExpired is the state of an expired challenge.
	AppAuthenticatorChallengeStateExpired = "expired"
)

// AppAuthenticatorDevice represents a registered Authelia app device.
type AppAuthenticatorDevice struct {
	ID                  int            `db:"id"`
	CreatedAt           time.Time      `db:"created_at"`
	LastUsedAt          sql.NullTime   `db:"last_used_at"`
	DeviceID            string         `db:"device_id"`
	Username            string         `db:"username"`
	BaseURL             string         `db:"base_url"`
	Name                string         `db:"name"`
	Model               string         `db:"model"`
	Platform            string         `db:"platform"`
	ApprovalPublicKey   []byte         `db:"approval_public_key"`
	TransportPublicKey  []byte         `db:"transport_public_key"`
	AttestationFormat   string         `db:"attestation_format"`
	AttestationIdentity sql.NullString `db:"attestation_identity"`
	AttestationDetail   sql.NullString `db:"attestation_detail"`
	AttestationKeyID    []byte         `db:"attestation_key_id"`
	PullIDCurrent       []byte         `db:"pull_id_current"`
	PullIDPrevious      []byte         `db:"pull_id_previous"`
}

// ToData converts the device into the data exposed to the portal.
func (d AppAuthenticatorDevice) ToData() AppAuthenticatorDeviceData {
	data := AppAuthenticatorDeviceData{
		ID:                d.ID,
		CreatedAt:         d.CreatedAt,
		Name:              d.Name,
		Model:             d.Model,
		Platform:          d.Platform,
		AttestationFormat: d.AttestationFormat,
	}

	if d.LastUsedAt.Valid {
		t := d.LastUsedAt.Time

		data.LastUsedAt = &t
	}

	if d.AttestationIdentity.Valid {
		data.AttestationIdentity = d.AttestationIdentity.String
	}

	if d.AttestationDetail.Valid {
		data.AttestationDetail = d.AttestationDetail.String
	}

	return data
}

// AppAuthenticatorDeviceData is the portal representation of an AppAuthenticatorDevice.
type AppAuthenticatorDeviceData struct {
	ID                  int        `json:"id"`
	CreatedAt           time.Time  `json:"created_at"`
	LastUsedAt          *time.Time `json:"last_used_at,omitempty"`
	Name                string     `json:"name"`
	Model               string     `json:"model"`
	Platform            string     `json:"platform"`
	AttestationFormat   string     `json:"attestation_format"`
	AttestationIdentity string     `json:"attestation_identity,omitempty"`
	AttestationDetail   string     `json:"attestation_detail,omitempty"`
}

// AppAuthenticatorRegistration represents an Authelia app registration code and its session.
type AppAuthenticatorRegistration struct {
	ID               int            `db:"id"`
	CreatedAt        time.Time      `db:"created_at"`
	ExpiresAt        time.Time      `db:"expires_at"`
	Code             []byte         `db:"code"`
	Username         string         `db:"username"`
	PortalBinding    []byte         `db:"portal_binding"`
	BaseURL          string         `db:"base_url"`
	Number           string         `db:"number"`
	NonceLength      int            `db:"nonce_length"`
	State            string         `db:"state"`
	BegunAt          sql.NullTime   `db:"begun_at"`
	SessionExpiresAt sql.NullTime   `db:"session_expires_at"`
	ConfirmationID   []byte         `db:"confirmation_id"`
	Challenge        sql.NullString `db:"challenge"`
	CodeChallenge    sql.NullString `db:"code_challenge"`
	DevicePlatform   sql.NullString `db:"device_platform"`
	DeviceModel      sql.NullString `db:"device_model"`
	DeviceName       sql.NullString `db:"device_name"`
	DeviceID         sql.NullString `db:"device_id"`
}

// AppAuthenticatorChallenge represents an Authelia app second factor challenge.
type AppAuthenticatorChallenge struct {
	ID                 int            `db:"id"`
	CreatedAt          time.Time      `db:"created_at"`
	ExpiresAt          time.Time      `db:"expires_at"`
	ConfirmationID     string         `db:"confirmation_id"`
	Username           string         `db:"username"`
	PortalBinding      []byte         `db:"portal_binding"`
	Challenge          string         `db:"challenge"`
	Number             string         `db:"number"`
	NonceLength        int            `db:"nonce_length"`
	State              string         `db:"state"`
	BoundDeviceID      sql.NullString `db:"bound_device_id"`
	BoundCodeChallenge sql.NullString `db:"bound_code_challenge"`
	DecidedAt          sql.NullTime   `db:"decided_at"`
	DecidedDeviceID    sql.NullString `db:"decided_device_id"`
	PortalURL          sql.NullString `db:"portal_url"`
	RequestedURL       sql.NullString `db:"requested_url"`
	RemoteIP           sql.NullString `db:"remote_ip"`
	UserAgent          sql.NullString `db:"user_agent"`
	ConsumedAt         sql.NullTime   `db:"consumed_at"`
}
