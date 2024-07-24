package embed

import (
	"fmt"
	"github.com/authelia/authelia/v4/internal/configuration/validator"

	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func NewConfig(paths []string, filters []configuration.BytesFilter) (keys []string, config *schema.Configuration, val *schema.StructValidator, err error) {
	sources := configuration.NewDefaultSourcesWithDefaults(
		paths,
		filters,
		configuration.DefaultEnvPrefix,
		configuration.DefaultEnvDelimiter,
		nil)

	val = schema.NewStructValidator()

	config = &schema.Configuration{}

	if keys, err = configuration.LoadAdvanced(
		val,
		"",
		config,
		sources...); err != nil {
		return nil, nil, nil, err
	}

	return keys, config, val, nil
}

func ValidateConfigKeys(keys []string, val *schema.StructValidator) {
	validator.ValidateKeys(keys, configuration.GetMultiKeyMappedDeprecationKeys(), configuration.DefaultEnvPrefix, val)
}

func ValidateConfig(config *schema.Configuration, val *schema.StructValidator) {
	validator.ValidateConfiguration(config, val)
}

func NewConfigFileFilters(names ...string) (filters []configuration.BytesFilter, err error) {
	if filters, err = configuration.NewFileFilters(names); err != nil {
		return nil, fmt.Errorf("error occurred loading filters: %w", err)
	}

	return filters, nil
}
