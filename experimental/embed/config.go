package embed

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/configuration/validator"
)

// NewConfiguration builds a new configuration given a list of paths and filters. The filters can either be nil or
// generated using NewNamedConfigFileFilters. This function essentially operates the same as Authelia does normally in
// configuration steps.
func NewConfiguration(paths []string, filters []configuration.BytesFilter) (keys []string, config *schema.Configuration, val *schema.StructValidator, err error) {
	sources := configuration.NewDefaultSourcesWithDefaults(
		paths,
		filters,
		configuration.DefaultEnvPrefix,
		configuration.DefaultEnvDelimiter,
		[]configuration.Source{configuration.NewMapSource(configuration.Defaults())})

	val = schema.NewStructValidator()

	var definitions *schema.Definitions

	if definitions, err = configuration.LoadDefinitions(val, sources...); err != nil {
		return nil, nil, nil, err
	}

	config = &schema.Configuration{}

	if keys, err = configuration.LoadAdvanced(
		val,
		"",
		config,
		definitions,
		sources...); err != nil {
		return nil, nil, nil, err
	}

	return keys, config, val, nil
}

// ValidateConfigurationAndKeys performs all configuration validation steps. The provided *schema.StructValidator should
// at minimum be checked for errors before continuing.
func ValidateConfigurationAndKeys(config *schema.Configuration, keys []string, val *schema.StructValidator) {
	ValidateConfigurationKeys(keys, val)
	ValidateConfiguration(config, val)
}

// ValidateConfigurationKeys just the keys validation steps. The provided *schema.StructValidator should
// at minimum be checked for errors before continuing. This should be used prior to using ValidateConfiguration.
func ValidateConfigurationKeys(keys []string, val *schema.StructValidator) {
	validator.ValidateKeys(keys, configuration.GetMultiKeyMappedDeprecationKeys(), configuration.DefaultEnvPrefix, val)
}

// ValidateConfiguration just the configuration validation steps. The provided *schema.StructValidator should
// at minimum be checked for errors before continuing. This should be used after using ValidateConfigurationKeys.
func ValidateConfiguration(config *schema.Configuration, val *schema.StructValidator) {
	validator.ValidateConfiguration(config, val)
}

// NewNamedConfigFileFilters allows configuring a set of file filters. The officially supported filter has the name
// 'template'. The only other one at this stage is 'expand-env' which is deprecated.
func NewNamedConfigFileFilters(names ...string) (filters []configuration.BytesFilter, err error) {
	if filters, err = configuration.NewFileFilters(names); err != nil {
		return nil, fmt.Errorf("error occurred loading filters: %w", err)
	}

	return filters, nil
}
