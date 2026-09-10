// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

func TestOIDCConformanceComposeSuppliesOIDCLoginPlaceholders(t *testing.T) {
	data, err := os.ReadFile("OIDCConformance/compose.yml")
	require.NoError(t, err)

	compose := struct {
		Services map[string]struct {
			Environment map[string]string `yaml:"environment"`
		} `yaml:"services"`
	}{}

	require.NoError(t, yaml.Unmarshal(data, &compose))

	server, ok := compose.Services["conformance-server"]
	require.True(t, ok, "the compose file must define a conformance-server service")

	for _, name := range []string{"OIDC_GOOGLE_CLIENTID", "OIDC_GOOGLE_SECRET", "OIDC_GITLAB_CLIENTID", "OIDC_GITLAB_SECRET"} {
		value, ok := server.Environment[name]

		assert.Truef(t, ok, "%s must be set: the image's entrypoint interpolates it unconditionally, so leaving it unset overrides the suite's own default with an empty system property", name)
		assert.NotEmptyf(t, value, "%s must not be empty: Spring rejects an empty client id while building OAuth2ClientProperties, which fails the application context before devmode is consulted", name)
	}
}
