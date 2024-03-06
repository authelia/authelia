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
	ID       int    `db:"id" yaml:"-"`
	Service  string `db:"service" yaml:"service" json:"service" jsonschema:"title=Service" jsonschema_description:"The service name this UUID is used with."`
	SectorID string `db:"sector_id" yaml:"sector_id" json:"sector_id" jsonschema:"title=Sector Identifier" jsonschema_description:"Sector Identifier this UUID is used with."`
	Username string `db:"username" yaml:"username" json:"username" jsonschema:"title=Username" jsonschema_description:"The username of the user this UUID is for."`

	Identifier uuid.UUID `db:"identifier" yaml:"identifier" json:"identifier" jsonschema:"title=Identifier" jsonschema_description:"The random UUID for this opaque identifier."`
}

// UserOpaqueIdentifiersExport represents a UserOpaqueIdentifier export file.
type UserOpaqueIdentifiersExport struct {
	Identifiers []UserOpaqueIdentifier `yaml:"identifiers" json:"identifiers" jsonschema:"title=Identifiers" jsonschema_description:"The list of opaque identifiers."`
}
