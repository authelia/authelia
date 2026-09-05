package model

import (
	"database/sql"
	"encoding/json"
	"time"
)

// OpenIDConnectLink represents a link between an Authelia user and an external OpenID Connect 1.0 Provider account.
type OpenIDConnectLink struct {
	ID             int            `db:"id" json:"id"`
	CreatedAt      time.Time      `db:"created_at" json:"created_at"`
	LastUsedAt     sql.NullTime   `db:"last_used_at" json:"-"`
	Provider       string         `db:"provider" json:"provider"`
	Issuer         string         `db:"issuer" json:"issuer"`
	Subject        string         `db:"subject" json:"subject"`
	Username       string         `db:"username" json:"-"`
	RemoteUsername sql.NullString `db:"remote_username" json:"-"`
}

// MarshalJSON returns the OpenIDConnectLink in a JSON friendly manner.
func (l OpenIDConnectLink) MarshalJSON() (data []byte, err error) {
	o := struct {
		ID             int        `json:"id"`
		CreatedAt      time.Time  `json:"created_at"`
		LastUsedAt     *time.Time `json:"last_used_at,omitempty"`
		Provider       string     `json:"provider"`
		Issuer         string     `json:"issuer"`
		Subject        string     `json:"subject"`
		RemoteUsername string     `json:"remote_username,omitempty"`
	}{
		ID:             l.ID,
		CreatedAt:      l.CreatedAt,
		Provider:       l.Provider,
		Issuer:         l.Issuer,
		Subject:        l.Subject,
		RemoteUsername: l.RemoteUsername.String,
	}

	if l.LastUsedAt.Valid {
		o.LastUsedAt = &l.LastUsedAt.Time
	}

	return json.Marshal(o)
}
