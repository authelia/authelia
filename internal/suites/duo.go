package suites

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/matryer/is"

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
	is := is.New(t)

	url := fmt.Sprintf("%s/allow", DuoBaseURL)
	if allowDeny == Deny {
		url = fmt.Sprintf("%s/deny", DuoBaseURL)
	}

	req, err := http.NewRequest("POST", url, nil)
	is.NoErr(err)

	client := NewHTTPClient()
	res, err := client.Do(req)
	is.NoErr(err)
	is.Equal(200, res.StatusCode)
}

// ConfigureDuoPreAuth configure duo api to respond with available devices or enrollment Url.
func ConfigureDuoPreAuth(t *testing.T, response duo.PreAuthResponse) {
	is := is.New(t)
	url := fmt.Sprintf("%s/preauth", DuoBaseURL)

	body, err := json.Marshal(response)
	is.NoErr(err)

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	is.NoErr(err)

	client := NewHTTPClient()
	res, err := client.Do(req)
	is.NoErr(err)
	is.Equal(200, res.StatusCode)
}
