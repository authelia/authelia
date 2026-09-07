package validator

import (
	"testing"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newDefaultAdministrationConfig() schema.Configuration {
	return schema.Configuration{
		Administration: schema.Administration{},
	}
}

func TestShouldSetDefaultAdministrationValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultAdministrationConfig()

	ValidateNTP(&config, validator)
}
