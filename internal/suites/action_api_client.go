package suites

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// newAPIClientWithProxy creates an http.Client that tunnels its requests through the given HTTP/HTTPS proxy (e.g. one
// of the suite's squid containers), so that Authelia sees the proxy's address as the request's remote IP - the same
// way it would if a browser configured with that proxy had made the request.
func newAPIClientWithProxy(t *testing.T, proxy string) *http.Client {
	t.Helper()

	proxyURL, err := url.Parse(proxy)
	require.NoError(t, err)

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	return &http.Client{
		Timeout: 10 * time.Second,
		Jar:     jar,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Needs to be enabled in suites. Not used in production.
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// apiFirstFactorLogin performs a 1FA login against the API directly (bypassing the frontend) using the given client,
// asserting that it succeeds.
func apiFirstFactorLogin(t *testing.T, client *http.Client, username, password string) {
	t.Helper()

	keepMeLoggedIn := false

	body, err := json.Marshal(map[string]any{
		"username":       username,
		"password":       password,
		"keepMeLoggedIn": keepMeLoggedIn,
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/firstfactor", AutheliaBaseURL), bytes.NewReader(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "first factor login failed for user '%s'", username)
}

// apiLogout logs the given client out via the API.
func apiLogout(t *testing.T, client *http.Client) {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/logout", AutheliaBaseURL), bytes.NewReader([]byte("{}")))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "logout failed")
}
