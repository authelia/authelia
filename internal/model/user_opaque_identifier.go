// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// NewUserOpaqueIdentifier either creates a new UserOpaqueIdentifier or returns an error.
func NewUserOpaqueIdentifier(service, sectorID, username string) (id *UserOpaqueIdentifier, err error) {
	var opaqueID uuid.UUID

	if opaqueID, err = uuid.NewRandom(); err != nil {
		return nil, fmt.Errorf("unable to generate uuid: %w", err)
	}

	return &UserOpaqueIdentifier{
		Service:    service,
		SectorID:   sectorID,
		Username:   username,
		Identifier: opaqueID,
	}, nil
}

// UserOpaqueIdentifier represents an opaque identifier for a user. Commonly used with OAuth 2.0 and OpenID Connect.
type UserOpaqueIdentifier struct {
	ID       int    `db:"id" yaml:"-" toml:"-" json:"-"`
	Service  string `db:"service" yaml:"service" toml:"service" json:"service" jsonschema:"title=Service" jsonschema_description:"The service name this UUID is used with."`
	SectorID string `db:"sector_id" yaml:"sector_id" toml:"sector_id" json:"sector_id" jsonschema:"title=Sector Identifier" jsonschema_description:"Sector Identifier this UUID is used with."`
	Username string `db:"username" yaml:"username" toml:"username" json:"username" jsonschema:"title=Username" jsonschema_description:"The username of the user this UUID is for."`

	Identifier uuid.UUID `db:"identifier" yaml:"identifier" toml:"identifier" json:"identifier" jsonschema:"title=Identifier" jsonschema_description:"The random UUID for this opaque identifier."`
}

// UserOpaqueIdentifiersExport represents a UserOpaqueIdentifier export file.
type UserOpaqueIdentifiersExport struct {
	Identifiers []UserOpaqueIdentifier `yaml:"identifiers" toml:"identifiers" json:"identifiers" jsonschema:"title=Identifiers" jsonschema_description:"The list of opaque identifiers."`
}

// OAuth2SessionID is the mapping of an Authelia session's public identifier to the 'sid' claim issued to a sector.
type OAuth2SessionID struct {
	ID        int       `db:"id"`
	Issuer    string    `db:"issuer"`
	SectorID  string    `db:"sector_id"`
	PublicID  string    `db:"public_id"`
	SessionID uuid.UUID `db:"sid"`
	CreatedAt time.Time `db:"created_at"`
}

// OAuth2LogoutSession is a validated OpenID Connect RP-Initiated Logout 1.0 request awaiting the confirmation of the
// End-User. It is identified by the flow id rather than kept in the session, as the request is made cross-site and the
// session cookie isn't available when the session cookie has the SameSite attribute set to Strict, or for a POST.
type OAuth2LogoutSession struct {
	ID          int       `db:"id"`
	FlowID      uuid.UUID `db:"flow_id"`
	ClientID    string    `db:"client_id"`
	Subject     string    `db:"subject"`
	SessionID   string    `db:"sid"`
	RedirectURI string    `db:"redirect_uri"`
	State       string    `db:"state"`
	CreatedAt   time.Time `db:"created_at"`
	ExpiresAt   time.Time `db:"expires_at"`
}
