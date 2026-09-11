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

func TestOIDCConformanceComposeMongoDB(t *testing.T) {
	type service struct {
		Environment map[string]string `yaml:"environment"`
		DependsOn   []string          `yaml:"depends_on"`
	}

	read := func(t *testing.T, path string) map[string]service {
		t.Helper()

		data, err := os.ReadFile(path)
		require.NoError(t, err)

		compose := struct {
			Services map[string]service `yaml:"services"`
		}{}

		require.NoError(t, yaml.Unmarshal(data, &compose))

		return compose.Services
	}

	t.Run("ShouldLeaveTheDatabaseToItsOwnComposeFile", func(t *testing.T) {
		services := read(t, "OIDCConformance/compose.yml")

		assert.NotContains(t, services, "conformance-mongodb")
		assert.NotContains(t, services["conformance-server"].DependsOn, "conformance-mongodb")
		assert.Equal(t, "${"+oidcConformanceMongoDBHostEnv+":-conformance-mongodb}", services["conformance-server"].Environment["MONGODB_HOST"],
			"the server uses the database the suite was pointed at, and its own otherwise")
	})

	t.Run("ShouldRunTheDatabaseFromItsOwnComposeFile", func(t *testing.T) {
		services := read(t, oidcConformanceMongoDBComposeFile)

		assert.Contains(t, services, "conformance-mongodb")
		assert.Contains(t, services["conformance-server"].DependsOn, "conformance-mongodb")
	})
}

func TestOIDCConformanceComposeFiles(t *testing.T) {
	assert.Contains(t, oidcConformanceComposeFiles(""), "internal/suites/"+oidcConformanceMongoDBComposeFile)
	assert.Contains(t, oidcConformanceLogServices(""), "conformance-mongodb")

	assert.NotContains(t, oidcConformanceComposeFiles("mongodb.example.com"), "internal/suites/"+oidcConformanceMongoDBComposeFile,
		"a database the suite is pointed at replaces its own")
	assert.NotContains(t, oidcConformanceLogServices("mongodb.example.com"), "conformance-mongodb", "there is no container to print the logs of")

	for _, host := range []string{"", "mongodb.example.com"} {
		assert.Contains(t, oidcConformanceComposeFiles(host), "internal/suites/OIDCConformance/compose.yml")
		assert.Contains(t, oidcConformanceLogServices(host), "conformance-server")
	}
}
