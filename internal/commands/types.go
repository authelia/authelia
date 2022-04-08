package commands

import (
	"github.com/authelia/authelia/v4/internal/model"
)

type exportUserOpaqueIdentifiers struct {
	Identifiers []model.UserOpaqueIdentifier `yaml:"identifiers"`
}
