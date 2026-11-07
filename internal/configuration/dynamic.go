// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package configuration

import (
	"errors"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// LoadDynamic loads the dynamic configuration from the given file or directory paths. A fresh set of sources is
// created on every call, which is required for reloads as sources retain what they previously loaded.
//
// Problems with the individual sources are pushed to the validator, the returned error is only for failures which
// prevent a result from being produced. The keys returned are those which were present in the sources.
func LoadDynamic(val *schema.StructValidator, paths []string, filters []BytesFilter) (dynamic *schema.DynamicConfiguration, keys []string, err error) {
	if len(paths) == 0 {
		return nil, nil, errors.New("no dynamic configuration paths were provided")
	}

	files := NewFilteredFileSources(paths, filters)

	sources := make([]Source, len(files))

	for i, source := range files {
		sources[i] = source
	}

	var definitions *schema.Definitions

	if definitions, err = LoadDefinitions(val, sources...); err != nil {
		return nil, nil, err
	}

	dynamic = &schema.DynamicConfiguration{}

	if keys, err = LoadAdvanced(val, "", dynamic, definitions, sources...); err != nil {
		return nil, keys, err
	}

	// Merge the legacy style access_control.networks into the definitions the same way the static configuration does.
	merged := schema.Definitions{Network: dynamic.Definitions.Network}

	mapDefinitions(val, dynamic.AccessControl.Networks, &merged)

	dynamic.Definitions.Network = merged.Network

	return dynamic, keys, nil
}

// NewDynamicFromStatic creates the dynamic configuration from the sections in a static configuration. It is used when
// no dynamic configuration paths are configured so behavior is unchanged from before it existed.
func NewDynamicFromStatic(config *schema.Configuration) (dynamic *schema.DynamicConfiguration) {
	return &schema.DynamicConfiguration{
		Definitions:   schema.DynamicDefinitions{Network: config.Definitions.Network},
		AccessControl: config.AccessControl,
	}
}
