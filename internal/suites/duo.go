package suites

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// DuoPolicy a type of policy
type DuoPolicy int32

const (
	// Deny deny policy
	Deny DuoPolicy = iota
	// Allow allow policy
	Allow DuoPolicy = iota
)

// ConfigureDuo configure duo api to allow or block auth requests
func ConfigureDuo(t *testing.T, allowDeny DuoPolicy) {
	url := fmt.Sprintf("%s/allow", DuoBaseURL)
	if allowDeny == Deny {
		url = fmt.Sprintf("%s/deny", DuoBaseURL)
	}

	req, err := http.NewRequest("POST", url, nil)
	assert.NoError(t, err)

	client := NewHTTPClient()
	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
}
