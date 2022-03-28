package model

import (
	"github.com/google/uuid"
)

// NewOpaqueUserID either creates a new OpaqueUserID or returns an error.
func NewOpaqueUserID(sectorID, username string) (id *OpaqueUserID, err error) {
	var opaqueID uuid.UUID

	if opaqueID, err = uuid.NewRandom(); err != nil {
		return nil, err
	}

	return &OpaqueUserID{
		SectorID: sectorID,
		Username: username,
		OpaqueID: opaqueID,
	}, nil
}

// OpaqueUserID represents an opaque identifier for a user. Commonly used with OAuth 2.0 and OpenID Connect.
type OpaqueUserID struct {
	ID       int    `db:"id"`
	SectorID string `db:"sector_id"`
	Username string `db:"username"`

	OpaqueID uuid.UUID `db:"opaque_id"`
}
