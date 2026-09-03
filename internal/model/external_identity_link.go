// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"database/sql"
	"encoding/json"
	"time"
)

// ExternalIdentityLink represents a link between an Authelia user and an external identity provider account.
type ExternalIdentityLink struct {
	ID             int            `db:"id" json:"id"`
	CreatedAt      time.Time      `db:"created_at" json:"created_at"`
	LastUsedAt     sql.NullTime   `db:"last_used_at" json:"-"`
	Type           string         `db:"type" json:"type"`
	Provider       string         `db:"provider" json:"provider"`
	Issuer         string         `db:"issuer" json:"issuer"`
	Subject        string         `db:"subject" json:"subject"`
	Username       string         `db:"username" json:"-"`
	RemoteUsername sql.NullString `db:"remote_username" json:"-"`
	Email          sql.NullString `db:"email" json:"-"`
	Signature      string         `db:"signature" json:"-"`
}

// MarshalJSON returns the ExternalIdentityLink in a JSON friendly manner.
func (l ExternalIdentityLink) MarshalJSON() (data []byte, err error) {
	o := struct {
		ID             int        `json:"id"`
		CreatedAt      time.Time  `json:"created_at"`
		LastUsedAt     *time.Time `json:"last_used_at,omitempty"`
		Type           string     `json:"type"`
		Provider       string     `json:"provider"`
		Issuer         string     `json:"issuer"`
		Subject        string     `json:"subject"`
		RemoteUsername string     `json:"remote_username,omitempty"`
		Email          string     `json:"email,omitempty"`
	}{
		ID:             l.ID,
		CreatedAt:      l.CreatedAt,
		Type:           l.Type,
		Provider:       l.Provider,
		Issuer:         l.Issuer,
		Subject:        l.Subject,
		RemoteUsername: l.RemoteUsername.String,
		Email:          l.Email.String,
	}

	if l.LastUsedAt.Valid {
		o.LastUsedAt = &l.LastUsedAt.Time
	}

	return json.Marshal(o)
}
