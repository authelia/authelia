package regulation

import (
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/utils"
)

// Regulator an authentication regulator preventing attackers to brute force the service.
type Regulator struct {
	// Is the regulation enabled.
	enabled bool

	config schema.RegulationConfiguration

	storageProvider storage.RegulatorProvider

	clock utils.Clock
}
