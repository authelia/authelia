package suites

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/duo"
)

// DuoPolicy a type of policy.
type DuoPolicy int32

const (
	// Deny deny policy.
	Deny DuoPolicy = iota
	// Allow allow policy.
	Allow DuoPolicy = iota
)

// ConfigureDuo configure duo api to allow or block auth requests.
func ConfigureDuo(t *testing.T, allowDeny DuoPolicy) {
	t.Helper()

	url := fmt.Sprintf("%s/allow", DuoBaseURL)
	if allowDeny == Deny {
		url = fmt.Sprintf("%s/deny", DuoBaseURL)
	}

	req, err := http.NewRequest(fasthttp.MethodPost, url, nil)
	require.NoError(t, err)

	client := NewHTTPClient()
	res, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, fasthttp.StatusOK, res.StatusCode)
}

// ConfigureDuoPreAuth configure duo api to respond with available devices or enrollment Url.
func ConfigureDuoPreAuth(t *testing.T, response duo.PreAuthResponse) {
	t.Helper()

	url := fmt.Sprintf("%s/preauth", DuoBaseURL)

	body, err := json.Marshal(response)
	require.NoError(t, err)

	req, err := http.NewRequest(fasthttp.MethodPost, url, bytes.NewReader(body))
	req.Header.Set(fasthttp.HeaderContentType, "application/json; charset=utf-8")
	require.NoError(t, err)

	client := NewHTTPClient()
	res, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, fasthttp.StatusOK, res.StatusCode)
}
