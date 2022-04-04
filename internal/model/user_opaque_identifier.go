package model

import (
	"fmt"

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
	ID       int    `db:"id"`
	Service  string `db:"service"`
	SectorID string `db:"sector_id"`
	Username string `db:"username"`

	Identifier uuid.UUID `db:"identifier"`
}
