// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"net/url"
	"os"
	"strings"
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

func TestOIDCConformanceServedByThePortal(t *testing.T) {
	base, err := url.ParseRequestURI(oidcConformanceBaseURL)
	require.NoError(t, err)

	conf, err := os.ReadFile("example/compose/nginx/portal/nginx.conf")
	require.NoError(t, err)

	nginx := string(conf)

	t.Run("ShouldListenOnThePortTheBaseURLUses", func(t *testing.T) {
		assert.Containsf(t, nginx, "listen "+base.Port()+" ssl;",
			"the portal must terminate TLS on %s, which is the port %s resolves to", base.Port(), oidcConformanceBaseURL)
	})

	t.Run("ShouldServeTheConformanceHost", func(t *testing.T) {
		assert.Equal(t, 3, strings.Count(nginx, `server_name ~^conformance\.example([0-9])*\.com$;`),
			"one server block per listener the published nginx image exposed")
		assert.Equal(t, 3, strings.Count(nginx, "set $upstream_endpoint http://conformance-server:8080;"),
			"the portal proxies straight to the server now that the published nginx image is gone")
	})

	t.Run("ShouldExposeTheMutualTLSAndFAPIListeners", func(t *testing.T) {
		// 8444 and 8445 carry no traffic yet, but a profile that reaches for them should find them rather than
		// fail part way through a plan, and they are the reason the published image existed at all.
		assert.Contains(t, nginx, "listen 8444 ssl;")
		assert.Contains(t, nginx, "listen 8445 ssl;")

		assert.Equal(t, 2, strings.Count(nginx, "ssl_verify_client optional_no_ca;"),
			"the mutual TLS listeners ask for a client certificate but leave verifying it to the suite")
		assert.Contains(t, nginx, "ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384;",
			"8445 narrows the ciphers to the FAPI 2.0 Final set")
	})

	t.Run("ShouldReproduceWhatThePublishedImageSet", func(t *testing.T) {
		// The suite depends on each of these: large headers carry its request objects, the long timeouts outlast
		// requests that block up to 60s on a callback, and it rebuilds its own base URL from the forwarded headers.
		for _, directive := range []string{
			"client_header_buffer_size 32k;",
			"large_client_header_buffers 4 32k;",
			"proxy_read_timeout 120s;",
			"proxy_send_timeout 120s;",
			"proxy_set_header  X-Forwarded-Host $host;",
			"proxy_set_header  X-Forwarded-Port $server_port;",
			"proxy_set_header  X-Ssl-Cert $ssl_client_cert;",
		} {
			assert.Equalf(t, 3, strings.Count(nginx, directive),
				"the published nginx image set %s on every listener and the suite relies on it", directive)
		}

		// Only the listener that does not ask for a client certificate tells the suite so.
		assert.Equal(t, 1, strings.Count(nginx, "proxy_set_header  X-Test-Mtls-Called-On-Wrong-Host $mtls_wrong_host;"))
		assert.Equal(t, 1, strings.Count(nginx, "ssl_verify_client off;"))
	})

	t.Run("ShouldResolveTheConformanceHostToThePortal", func(t *testing.T) {
		entries := map[string]string{}

		for _, entry := range HostEntries() {
			entries[entry.Domain] = entry.IP
		}

		require.Contains(t, entries, base.Hostname())
		assert.Equal(t, entries["login.example.com"], entries[base.Hostname()],
			"the conformance host is served by the portal, so it resolves where every other portal domain does")
	})

	t.Run("ShouldRunThePortal", func(t *testing.T) {
		assert.Contains(t, oidcConformanceComposeFiles(""), "internal/suites/example/compose/nginx/portal/compose.yml")
	})
}
